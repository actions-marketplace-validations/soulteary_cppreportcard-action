package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/soulteary/cppreportcard-action/internal/config"
	"github.com/soulteary/cppreportcard-action/internal/model"
	"github.com/soulteary/cppreportcard-action/internal/project"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func put(t *testing.T, p, s string) {
	t.Helper()
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, []byte(s), 0600); e != nil {
		t.Fatal(e)
	}
}
func fixture(t *testing.T) (string, []string) {
	t.Helper()
	root := t.TempDir()
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "a.cpp"), "int main() {}\n")
	c := config.Default()
	for _, name := range []string{"clang-format", "clang-tidy", "cppcheck"} {
		p := filepath.Join(root, name)
		body := "exit 0"
		if name == "clang-format" {
			body = "cat a.cpp"
		}
		if name == "cppcheck" {
			body = `if [ "$CPP_FINDING" = 1 ]; then echo '<results><errors><error id="nullPointer" severity="error" msg="issue"><location file="a.cpp" line="1"/></error></errors></results>'; else echo '<results><errors/></results>'; fi`
		}
		put(t, p, "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'tool 1';exit 0;fi\n"+body+"\n")
		if err := os.Chmod(p, 0700); err != nil {
			t.Fatal(err)
		}
		c.Tools[name] = p
	}
	b, _ := json.Marshal(c)
	put(t, filepath.Join(root, "config.json"), string(b))
	return root, []string{"--directory", root, "--config", "config.json"}
}
func invoke(t *testing.T, args []string, want int) model.Result {
	t.Helper()
	var out, errout bytes.Buffer
	code := Run(append(args, "--json"), &out, &errout, "test")
	if code != want {
		t.Fatalf("exit %d want %d %s%s", code, want, &out, &errout)
	}
	var r model.Result
	if e := json.Unmarshal(out.Bytes(), &r); e != nil {
		t.Fatal(e, out.String())
	}
	return r
}
func TestCLIFlagsAndErrors(t *testing.T) {
	for _, tc := range []struct {
		args     []string
		code     int
		contains string
	}{{[]string{"--help"}, 0, "Usage"}, {[]string{"--version"}, 0, "test"}, {[]string{"--bad"}, 2, "flag"}, {[]string{"stray"}, 2, "positional"}, {[]string{"--threshold", "NaN"}, 2, "threshold"}, {[]string{"--threshold", "Inf"}, 2, "threshold"}, {[]string{"--threshold", "101"}, 2, "threshold"}, {[]string{"--fail-on", "bad"}, 2, "fail-on"}, {[]string{"--directory", filepath.Join(t.TempDir(), "missing")}, 2, "no such"}} {
		var out, errout bytes.Buffer
		if code := Run(tc.args, &out, &errout, "test"); code != tc.code || !strings.Contains(out.String()+errout.String(), tc.contains) {
			t.Fatalf("%v %d %s%s", tc.args, code, &out, &errout)
		}
	}
	root, args := fixture(t)
	put(t, filepath.Join(root, "bad.json"), "{")
	for _, extra := range [][]string{{"--mode", "bad"}, {"--config", "missing"}, {"--config", "bad.json"}, {"--baseline", "missing"}, {"--baseline", "bad.json"}} {
		if Run(append(args, extra...), io.Discard, io.Discard, "test") != 2 {
			t.Fatal(extra)
		}
	}
	if Run([]string{"--directory", t.TempDir()}, io.Discard, io.Discard, "test") != 2 {
		t.Fatal("empty accepted")
	}
}

type failingOutput struct{}

func (failingOutput) Write([]byte) (int, error) { return 0, errors.New("write failure") }
func TestCLIAnalysisPoliciesAndArtifacts(t *testing.T) {
	root, args := fixture(t)
	r := invoke(t, args, 0)
	if !r.Complete || r.Grade != "A+" {
		t.Fatal(r)
	}
	t.Setenv("CPP_FINDING", "1")
	for _, policy := range []string{"findings", "errors", "new-errors"} {
		r = invoke(t, append(args, "--fail-on", policy), 1)
		if r.NewErrors != 1 {
			t.Fatal(r)
		}
	}
	r = invoke(t, append(args, "--fail-on", "none"), 0)
	b, _ := json.Marshal(r)
	put(t, filepath.Join(root, "baseline.json"), string(b))
	r = invoke(t, append(args, "--baseline", "baseline.json"), 0)
	if r.NewErrors != 0 {
		t.Fatal(r)
	}
	invoke(t, append(args, "--fail-on", "none", "--threshold", "100"), 1)
	put(t, filepath.Join(root, "baseline.json"), "{}")
	if Run(append(args, "--baseline", "baseline.json"), io.Discard, io.Discard, "test") != 2 {
		t.Fatal("invalid baseline accepted")
	}
	t.Setenv("CPP_FINDING", "")
	var out bytes.Buffer
	if Run(args, &out, io.Discard, "test") != 0 || !strings.Contains(out.String(), "C++ Report Card") {
		t.Fatal(out.String())
	}
	if Run(append(args, "--json"), failingOutput{}, io.Discard, "test") != 2 {
		t.Fatal("writer failure ignored")
	}
	put(t, filepath.Join(root, "blocked"), "x")
	if Run(append(args, "--output-dir", "blocked/out"), io.Discard, io.Discard, "test") != 2 {
		t.Fatal("file error ignored")
	}
	invoke(t, append(args, "--mode", "full"), 2)
	cmds := []project.Command{{Directory: root, File: "a.cpp", Command: "c++ -c a.cpp"}}
	b, _ = json.Marshal(cmds)
	put(t, filepath.Join(root, "compile_commands.json"), string(b))
	r = invoke(t, append(args, "--mode", "full", "--compile-commands", "compile_commands.json"), 0)
	if !r.Complete {
		t.Fatal(r)
	}
	put(t, filepath.Join(root, "uncovered.cpp"), "int x;\n")
	put(t, filepath.Join(root, "a.h"), "#pragma once\n")
	r = invoke(t, append(args, "--mode", "full", "--compile-commands", "compile_commands.json"), 2)
	if r.Complete {
		t.Fatal("uncovered source accepted")
	}
}
