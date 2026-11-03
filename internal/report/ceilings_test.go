package report

import (
	"fmt"
	"github.com/soulteary/cppreportcard-action/internal/model"
	"testing"
)

func TestGradeBoundariesAndDeduplication(t *testing.T) {
	for _, tc := range []struct {
		affected, files int
		grade           string
		score           float64
	}{{0, 100, "A+", 100}, {5, 100, "A", 95}, {10, 100, "A", 90}, {20, 100, "B", 80}, {30, 100, "C", 70}, {40, 100, "D", 60}, {50, 100, "F", 50}, {2, 1, "F", 0}, {0, 0, "A+", 100}} {
		ds := []model.Diagnostic{}
		for i := 0; i < tc.affected; i++ {
			ds = append(ds, model.Diagnostic{File: fmt.Sprint(i), Severity: "warning"})
		}
		r := model.Result{Files: tc.files, Checks: []model.Check{{Status: "passed", Diagnostics: ds}, {Status: "skipped"}}}
		Finalize(&r, nil)
		if tc.affected == 5 {
			if r.Grade != "A+" {
				t.Fatal(r)
			}
			continue
		}
		if r.Grade != tc.grade || *r.Score != tc.score {
			t.Fatalf("%+v => %+v", tc, r)
		}
	}
	ds := []model.Diagnostic{{File: "z", Rule: "a"}, {File: "a", Line: 2}, {File: "a", Line: 1, Rule: "z"}, {File: "a", Line: 1, Rule: "a", Message: "z"}, {File: "a", Line: 1, Rule: "a", Message: "a"}, {File: "a", Line: 1, Rule: "a", Message: "a"}}
	r := model.Result{Files: 6, Checks: []model.Check{{Status: "passed", Diagnostics: ds}}}
	Finalize(&r, nil)
	got := r.Checks[0].Diagnostics
	if len(got) != 5 || got[0].Message != "a" || got[4].File != "z" {
		t.Fatalf("%+v", got)
	}
}
