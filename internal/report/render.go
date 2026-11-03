package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/soulteary/cppreportcard-action/internal/model"
)

func escape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "|", "&#124;", "\n", " ", "\r", " ", "`", "&#96;")
	return r.Replace(s)
}
func Markdown(r model.Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# C++ Report Card\n\n**Grade: %s** · Mode: %s · Status: %s\n\n", r.Grade, r.Mode, r.Status)
	if r.Score != nil {
		fmt.Fprintf(&b, "Score: %.1f / 100\n\n", *r.Score)
	}
	fmt.Fprintf(&b, "Files: %d · New errors: %d · Scoring version: %d\n\n", r.Files, r.NewErrors, r.ScoringVersion)
	if r.Mode == "basic" {
		b.WriteString("Basic analysis does not use a complete build context and is not equivalent to full analysis.\n\n")
	}
	if !r.Complete {
		b.WriteString("**Analysis incomplete. No quality grade is available.**\n\n")
	}
	b.WriteString("| Tool | Version | Status | Analyzed units |\n| --- | --- | --- | --- |\n")
	for _, c := range r.Checks {
		fmt.Fprintf(&b, "| %s | %s | %s | %d / %d |\n", escape(c.Tool), escape(c.Version), c.Status, c.Analyzed, c.Total)
	}
	for _, c := range r.Checks {
		fmt.Fprintf(&b, "\n## %s\n\n", c.Tool)
		if c.Error != "" {
			fmt.Fprintf(&b, "%s\n\n", escape(c.Error))
		}
		if len(c.Diagnostics) == 0 {
			continue
		}
		b.WriteString("| File | Line | Severity | Rule | New | Message |\n| --- | --- | --- | --- | --- | --- |\n")
		for _, d := range c.Diagnostics {
			fmt.Fprintf(&b, "| %s | %d | %s | %s | %t | %s |\n", escape(d.File), d.Line, d.Severity, escape(d.Rule), d.New, escape(d.Message))
		}
	}
	return b.String()
}
func Write(dir string, r model.Result) error {
	if e := os.MkdirAll(dir, 0755); e != nil {
		return e
	}
	data, e := json.MarshalIndent(r, "", "  ")
	if e != nil {
		return e
	}
	for _, f := range []struct {
		name string
		data []byte
	}{{"report.json", append(data, '\n')}, {"report.md", []byte(Markdown(r))}, {"badge.svg", []byte(SVG(r))}} {
		if e = os.WriteFile(filepath.Join(dir, f.name), f.data, 0644); e != nil {
			return e
		}
	}
	return nil
}
