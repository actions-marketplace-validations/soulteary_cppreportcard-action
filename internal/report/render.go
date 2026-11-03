package report

import (
	"encoding/json"
	"fmt"
	"html"
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
func SVG(r model.Result) string {
	label := "C++ " + r.Mode
	value := r.Grade
	color := "#777"
	if r.Complete {
		color = "#d99a00"
		if r.Grade == "A+" || r.Grade == "A" {
			color = "#22863a"
		}
		if r.Grade == "F" {
			color = "#c62828"
		}
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="220" height="20" role="img" aria-label="%s: %s"><title>%s: %s</title><rect width="110" height="20" fill="#555"/><rect x="110" width="110" height="20" fill="%s"/><g fill="#fff" font-family="Verdana,sans-serif" font-size="11" text-anchor="middle"><text x="55" y="14">%s</text><text x="165" y="14">%s</text></g></svg>`, html.EscapeString(label), html.EscapeString(value), html.EscapeString(label), html.EscapeString(value), color, html.EscapeString(label), html.EscapeString(value)) + "\n"
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
