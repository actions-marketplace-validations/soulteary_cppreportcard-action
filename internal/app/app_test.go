package app

import (
	"bytes"
	"encoding/json"
	"github.com/soulteary/cppreportcard-action/internal/model"
	"os"
	"path/filepath"
	"testing"
)

func TestMissingToolsProduceReports(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "a.cpp"), []byte("int main() {}\n"), 0644)
	cfg := `{"tools":{"clang-format":"/nonexistent/format","clang-tidy":"/nonexistent/tidy","cppcheck":"/nonexistent/cppcheck"}}`
	os.WriteFile(filepath.Join(d, "config.json"), []byte(cfg), 0644)
	var out, err bytes.Buffer
	code := Run([]string{"--directory", d, "--config", "config.json", "--json"}, &out, &err, "test")
	if code != 2 {
		t.Fatalf("code %d: %s", code, err.String())
	}
	var r model.Result
	if e := json.Unmarshal(out.Bytes(), &r); e != nil {
		t.Fatal(e)
	}
	if r.Complete || r.Score != nil {
		t.Fatal(r)
	}
	for _, f := range []string{"report.json", "report.md", "badge.svg"} {
		if _, e := os.Stat(filepath.Join(d, ".cppreportcard", f)); e != nil {
			t.Fatal(e)
		}
	}
}
func TestInvalidArgs(t *testing.T) {
	for _, args := range [][]string{{"--mode", "weird"}, {"--threshold", "101"}, {"--fail-on", "yes"}} {
		var b bytes.Buffer
		if Run(args, &b, &b, "test") != 2 {
			t.Fatal(args)
		}
	}
}
