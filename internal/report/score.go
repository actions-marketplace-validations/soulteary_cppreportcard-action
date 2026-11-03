package report

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/soulteary/cppreportcard-action/internal/model"
)

// Finalize deduplicates per tool and scores affected files, not warning counts.
// Baselines never change the absolute grade; they affect only new-issue gates.
func Finalize(r *model.Result, baseline map[string]bool) {
	r.Complete = true
	r.NewErrors = 0
	for i := range r.Checks {
		c := &r.Checks[i]
		if c.Status == "error" {
			r.Complete = false
		}
		seen := map[string]bool{}
		out := []model.Diagnostic{}
		for _, d := range c.Diagnostics {
			// Line numbers intentionally excluded: adding a comment must not make an old finding new.
			sum := sha256.Sum256([]byte(strings.Join([]string{d.Tool, d.Rule, d.File, d.Message}, "\x00")))
			d.Fingerprint = fmt.Sprintf("%x", sum[:])
			d.New = !baseline[d.Fingerprint]
			key := fmt.Sprintf("%s:%d:%d", d.Fingerprint, d.Line, d.Column)
			if seen[key] {
				continue
			}
			seen[key] = true
			if d.New && d.Severity == "error" {
				r.NewErrors++
			}
			out = append(out, d)
		}
		sort.Slice(out, func(i, j int) bool {
			a, b := out[i], out[j]
			if a.File != b.File {
				return a.File < b.File
			}
			if a.Line != b.Line {
				return a.Line < b.Line
			}
			if a.Rule != b.Rule {
				return a.Rule < b.Rule
			}
			return a.Message < b.Message
		})
		c.Diagnostics = out
	}
	r.Score = nil
	r.Grade = "incomplete"
	r.Status = "incomplete"
	if !r.Complete {
		return
	}
	// v1: worst affected-file proportion among completed checks; any error caps at 59.
	score := 100.0
	hasError := false
	for _, c := range r.Checks {
		if c.Status == "skipped" {
			continue
		}
		affected := map[string]bool{}
		for _, d := range c.Diagnostics {
			affected[d.File] = true
			if d.Severity == "error" {
				hasError = true
			}
		}
		if r.Files > 0 {
			s := 100 * (1 - float64(len(affected))/float64(r.Files))
			if s < score {
				score = s
			}
		}
	}
	if score < 0 {
		score = 0
	}
	if hasError && score > 59 {
		score = 59
	}
	r.Score = &score
	r.Status = "passed"
	for _, c := range r.Checks {
		if len(c.Diagnostics) > 0 {
			r.Status = "findings"
		}
	}
	switch {
	case score >= 95:
		r.Grade = "A+"
	case score >= 90:
		r.Grade = "A"
	case score >= 80:
		r.Grade = "B"
	case score >= 70:
		r.Grade = "C"
	case score >= 60:
		r.Grade = "D"
	default:
		r.Grade = "F"
	}
}
