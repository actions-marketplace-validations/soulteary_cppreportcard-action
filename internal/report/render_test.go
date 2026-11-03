package report

import (
	"encoding/json"
	"github.com/soulteary/cppreportcard-action/internal/model"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReportFilesAndEscaping(t *testing.T) {
	score := 85.0
	r := model.Result{Mode: "basic", Complete: true, Grade: "B", Status: "findings", Score: &score, Files: 2, Checks: []model.Check{
		{Tool: "lint", Version: "1|2", Status: "ok", Total: 2, Analyzed: 2, Error: "<failure>\nnext", Diagnostics: []model.Diagnostic{{File: "a|b", Line: 3, Severity: "warning", Rule: "x`y", Message: "<&>\r\n|`", New: true}}},
		{Tool: "empty", Status: "skipped"}}}

	text := Markdown(r)
	for _, want := range []string{"Score: 85.0", "&lt;failure&gt; next", "a&#124;b", "x&#96;y", "&lt;&amp;&gt;  &#124;&#96;", "No quality grade"} {
		if want == "No quality grade" {
			if strings.Contains(text, want) {
				t.Fatal(text)
			}
			continue
		}
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %s", want, text)
		}
	}
	r.Complete = false
	r.Score = nil
	if !strings.Contains(Markdown(r), "Analysis incomplete") {
		t.Fatal("missing incomplete warning")
	}
	dir := t.TempDir()
	if err := Write(dir, r); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"report.json", "report.md", "badge.svg"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || len(b) == 0 {
			t.Fatalf("%s: %v", name, err)
		}
	}
	data, _ := os.ReadFile(filepath.Join(dir, "report.json"))
	var got model.Result
	if err := json.Unmarshal(data, &got); err != nil || got.Files != 2 {
		t.Fatalf("invalid report: %s %v", data, err)
	}
	blocked := filepath.Join(dir, "file")
	if err := os.WriteFile(blocked, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if Write(filepath.Join(blocked, "out"), r) == nil {
		t.Fatal("expected mkdir error")
	}
	r.Score = new(float64)
	*r.Score = math.NaN()
	if Write(dir, r) == nil {
		t.Fatal("expected JSON error")
	}
	r.Score = nil
	for _, name := range []string{"report.json", "report.md", "badge.svg"} {
		out := t.TempDir()
		if err := os.Mkdir(filepath.Join(out, name), 0700); err != nil {
			t.Fatal(err)
		}
		if Write(out, r) == nil {
			t.Fatalf("expected write error at %s", name)
		}
	}
}
func TestBadgePalette(t *testing.T) {
	for grade, want := range map[string]string{"A+": "44cc11", "A": "97ca00", "B": "a4a61d", "C": "dfb317", "D": "fe7d37", "E": "e05d44", "F": "e05d44", "unknown": "9f9f9f"} {
		if got := gradeColor(grade); got != want {
			t.Errorf("%s=%s want %s", grade, got, want)
		}
	}
}
