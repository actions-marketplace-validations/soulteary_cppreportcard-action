package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/soulteary/cppreportcard-action/internal/checker"
	"github.com/soulteary/cppreportcard-action/internal/config"
	"github.com/soulteary/cppreportcard-action/internal/model"
	"github.com/soulteary/cppreportcard-action/internal/project"
	"github.com/soulteary/cppreportcard-action/internal/report"
)

// Run returns 0 for success, 1 for a policy violation, 2 for incomplete analysis or invalid input.
func Run(args []string, stdout, stderr io.Writer, version string) int {
	fs := flag.NewFlagSet("cppreportcard-cli", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("directory", ".", "project root")
	cfg := fs.String("config", "", "JSON configuration path relative to project root")
	mode := fs.String("mode", "", "basic or full (overrides config)")
	db := fs.String("compile-commands", "", "compilation database path relative to project root")
	output := fs.String("output-dir", ".cppreportcard", "report directory relative to project root")
	baseline := fs.String("baseline", "", "previous report.json (same mode and scoring version)")
	policy := fs.String("fail-on", "new-errors", "none, errors, new-errors, or findings")
	threshold := fs.Float64("threshold", 0, "minimum score 0..100; 0 disables")
	asJSON := fs.Bool("json", false, "emit result JSON to stdout; exit policy still applies")
	ver := fs.Bool("version", false, "print version")
	if e := fs.Parse(args); e != nil {
		if e == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if *ver {
		fmt.Fprintln(stdout, version)
		return 0
	}
	fail := func(e error) int { fmt.Fprintln(stderr, e); return 2 }
	if fs.NArg() != 0 {
		return fail(fmt.Errorf("unexpected positional arguments"))
	}
	if *threshold < 0 || *threshold > 100 {
		return fail(fmt.Errorf("threshold must be 0..100"))
	}
	switch *policy {
	case "none", "errors", "new-errors", "findings":
	default:
		return fail(fmt.Errorf("invalid fail-on policy"))
	}
	abs, e := filepath.Abs(*root)
	if e != nil {
		return fail(e)
	}
	abs, e = filepath.EvalSymlinks(abs)
	if e != nil {
		return fail(e)
	}
	resolve := func(p string) string {
		if p == "" || filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(abs, p)
	}
	c, e := config.Load(resolve(*cfg))
	if e != nil {
		return fail(e)
	}
	if *mode != "" {
		c.Mode = *mode
	}
	if c.Mode != "basic" && c.Mode != "full" {
		return fail(fmt.Errorf("invalid mode"))
	}
	if *db != "" {
		c.CompileCommands = *db
	}
	files, e := project.Discover(abs, c.Exclude)
	if e != nil {
		return fail(e)
	}
	if len(files) == 0 {
		return fail(fmt.Errorf("no C/C++ files found"))
	}
	r := model.Result{SchemaVersion: 1, ScoringVersion: 1, Mode: c.Mode, Files: len(files), Checks: []model.Check{}}
	var units []project.Command
	if c.Mode == "full" {
		units, e = project.LoadCommands(abs, c.CompileCommands, c.Exclude)
		if e != nil {
			r.Checks = append(r.Checks, model.Check{Tool: "compilation-database", Status: "error", Error: e.Error(), Diagnostics: []model.Diagnostic{}})
		}
	}
	prior := map[string]bool{}
	if *baseline != "" {
		data, e := os.ReadFile(resolve(*baseline))
		if e != nil {
			return fail(e)
		}
		var old model.Result
		if e = json.Unmarshal(data, &old); e != nil {
			return fail(e)
		}
		if old.SchemaVersion != 1 || old.ScoringVersion != 1 || old.Mode != c.Mode || !old.Complete {
			return fail(fmt.Errorf("baseline must be complete and match schema, scoring version and mode"))
		}
		for _, check := range old.Checks {
			for _, d := range check.Diagnostics {
				prior[d.Fingerprint] = true
			}
		}
	}
	runner := checker.Runner{Root: abs, Config: c}
	r.Checks = append(r.Checks, runner.Format(files))
	if c.Mode == "full" && len(units) == 0 {
		r.Checks = append(r.Checks, model.Check{Tool: "clang-tidy", Status: "skipped", Error: "invalid compilation database", Diagnostics: []model.Diagnostic{}}, model.Check{Tool: "cppcheck", Status: "skipped", Error: "invalid compilation database", Diagnostics: []model.Diagnostic{}})
	} else {
		r.Checks = append(r.Checks, runner.Cppcheck(files, units))
		if c.Mode == "full" {
			r.Checks = append(r.Checks, runner.Tidy(units))
		} else {
			r.Checks = append(r.Checks, model.Check{Tool: "clang-tidy", Status: "skipped", Error: "requires full mode", Diagnostics: []model.Diagnostic{}})
		}
	}
	report.Finalize(&r, prior)
	if e = report.Write(resolve(*output), r); e != nil {
		return fail(e)
	}
	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if e = enc.Encode(r); e != nil {
			return fail(e)
		}
	} else {
		fmt.Fprintf(stdout, "C++ Report Card: %s (%s, %s)\nReports: %s\n", r.Grade, r.Mode, r.Status, resolve(*output))
	}
	if !r.Complete {
		return 2
	}
	violated := *threshold > 0 && r.Score != nil && *r.Score < *threshold
	for _, check := range r.Checks {
		for _, d := range check.Diagnostics {
			if *policy == "findings" || (*policy == "errors" && d.Severity == "error") || (*policy == "new-errors" && d.Severity == "error" && d.New) {
				violated = true
			}
		}
	}
	if violated {
		return 1
	}
	return 0
}
