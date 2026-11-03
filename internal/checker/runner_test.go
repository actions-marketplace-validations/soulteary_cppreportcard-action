package checker

import (
	"github.com/soulteary/cppreportcard-action/internal/config"
	"github.com/soulteary/cppreportcard-action/internal/project"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func toolStub(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(p, []byte("#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'tool 1.0'; exit 0; fi\n"+body+"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	return p
}
func runnerFixture(t *testing.T) Runner {
	t.Helper()
	r := Runner{Root: t.TempDir(), Config: config.Default()}
	if err := os.WriteFile(filepath.Join(r.Root, "a.cpp"), []byte("original\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return r
}
func TestFormattingAndExecution(t *testing.T) {
	r := runnerFixture(t)
	for _, tc := range []struct {
		body, status string
		analyzed     int
	}{{"printf 'original\\n'", "passed", 1}, {"printf 'formatted\\n'", "findings", 1}, {"echo failed; exit 2", "error", 0}} {
		r.Config.Tools["clang-format"] = toolStub(t, tc.body)
		c := r.Format([]string{"a.cpp"})
		if c.Status != tc.status || c.Analyzed != tc.analyzed || c.Version != "tool 1.0" {
			t.Fatalf("%+v", c)
		}
	}
	r.Config.Tools["clang-format"] = toolStub(t, "exit 0")
	if c := r.Format([]string{"missing.cpp"}); c.Status != "error" {
		t.Fatal(c)
	}
	r.Config.Tools["clang-format"] = "/missing/tool"
	if c := r.Format(nil); c.Status != "error" {
		t.Fatal(c)
	}
	r.Config.TimeoutSeconds = -1
	if _, err := r.command("clang-format"); err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatal(err)
	}
}
func TestTidyWorkersAndDatabase(t *testing.T) {
	r := runnerFixture(t)
	cmds := []project.Command{{Directory: r.Root, File: filepath.Join(r.Root, "a.cpp"), Command: "c++ -c a.cpp"}}
	for _, tc := range []struct {
		body, status string
		analyzed     int
	}{{`test -f "$2/compile_commands.json" || exit 3; echo 'a.cpp:2:3: warning: issue [bugprone-x]'`, "findings", 1}, {`echo 'a.cpp:2:3: error: syntax'`, "error", 0}, {`echo failure; exit 2`, "error", 0}, {`exit 0`, "passed", 1}} {
		r.Config.Tools["clang-tidy"] = toolStub(t, tc.body)
		c := r.Tidy(cmds)
		if c.Status != tc.status || c.Analyzed != tc.analyzed {
			t.Fatalf("%+v", c)
		}
	}
	r.Config.Tools["clang-tidy"] = "/missing/tool"
	if r.Tidy(cmds).Status != "error" {
		t.Fatal("missing tool accepted")
	}
	r.Config.Tools["clang-tidy"] = toolStub(t, "exit 0")
	t.Setenv("TMPDIR", filepath.Join(r.Root, "missing"))
	if r.Tidy(cmds).Status != "error" {
		t.Fatal("unavailable temp directory ignored")
	}
}
func TestCppcheckExecution(t *testing.T) {
	r := runnerFixture(t)
	cmds := []project.Command{{Directory: r.Root, File: filepath.Join(r.Root, "a.cpp"), Command: "cc"}}
	for _, tc := range []struct{ body, status string }{
		{`echo '<results><errors/></results>'`, "passed"},
		{`echo '<results><errors><error id="style" severity="style" msg="issue"><location file="a.cpp" line="1"/></error></errors></results>'`, "findings"},
		{`echo '<results><errors><error id="syntaxError" severity="error" msg="bad"/></errors></results>'`, "error"},
		{`echo bad`, "error"}, {`echo '<results/>'; exit 2`, "error"},
	} {
		for _, units := range [][]project.Command{nil, cmds} {
			r.Config.Tools["cppcheck"] = toolStub(t, tc.body)
			c := r.Cppcheck([]string{"a.cpp"}, units)
			if c.Status != tc.status || c.Total != 1 {
				t.Fatalf("%+v", c)
			}
		}
	}
	r.Config.Tools["cppcheck"] = "/missing/tool"
	if r.Cppcheck(nil, nil).Status != "error" {
		t.Fatal("missing tool accepted")
	}
	r.Config.Tools["cppcheck"] = toolStub(t, "exit 0")
	t.Setenv("TMPDIR", filepath.Join(r.Root, "missing"))
	if r.Cppcheck(nil, nil).Status != "error" {
		t.Fatal("temp failure ignored")
	}
}
func TestParserFiltering(t *testing.T) {
	root := t.TempDir()
	out := "ignore\n../outside.cpp:1:1: warning: bad\nskip/x.cpp:1:1: warning: bad\na.cpp:1:1: warning: plain"
	ds, b := ParseTidy(root, out, []string{"skip"})
	if b || len(ds) != 1 || ds[0].Rule != "clang-diagnostic" {
		t.Fatalf("%+v %v", ds, b)
	}
	xml := `<results><errors><error id="x" severity="warning"/><error id="x" severity="error"/><error id="x" severity="style"><location file="../outside.cpp"/></error><error id="x" severity="style"><location file="skip/a.cpp"/></error></errors></results>`
	ds, b, err := ParseCppcheck(root, xml, []string{"skip"})
	if err != nil || !b || len(ds) != 0 {
		t.Fatalf("%+v %v %v", ds, b, err)
	}
	if _, _, err = ParseCppcheck(root, "<results>", nil); err == nil {
		t.Fatal("malformed XML accepted")
	}
}
