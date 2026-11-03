package report

import (
	"github.com/soulteary/cppreportcard-action/internal/model"
	"testing"
)

func TestIncompleteNeverScores(t *testing.T) {
	r := model.Result{Files: 1, Checks: []model.Check{{Status: "error"}}}
	Finalize(&r, nil)
	if r.Score != nil || r.Complete || r.Grade != "incomplete" {
		t.Fatal(r)
	}
}
func TestDedupBaseline(t *testing.T) {
	d := model.Diagnostic{Tool: "cppcheck", Rule: "nullPointer", File: "a.cpp", Line: 2, Severity: "error", Message: "null"}
	r := model.Result{Files: 10, Checks: []model.Check{{Status: "findings", Diagnostics: []model.Diagnostic{d, d}}}}
	Finalize(&r, nil)
	if r.NewErrors != 1 || len(r.Checks[0].Diagnostics) != 1 || *r.Score > 59 {
		t.Fatal(r)
	}
	fp := r.Checks[0].Diagnostics[0].Fingerprint
	r.Checks[0].Diagnostics[0].Line = 30
	Finalize(&r, map[string]bool{fp: true})
	if r.NewErrors != 0 {
		t.Fatal("line drift invalidated baseline")
	}
}
