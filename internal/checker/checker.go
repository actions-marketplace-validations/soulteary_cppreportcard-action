// Package checker adapts external analyzers without executing shell commands.
package checker

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/soulteary/cppreportcard-action/internal/config"
	"github.com/soulteary/cppreportcard-action/internal/model"
	"github.com/soulteary/cppreportcard-action/internal/project"
)

type Runner struct {
	Root   string
	Config config.Config
}

func (r Runner) command(tool string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Config.Timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, r.Config.Tools[tool], args...)
	cmd.Dir = r.Root
	out, e := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return string(out), fmt.Errorf("%s: %w", tool, ctx.Err())
	}
	return string(out), e
}
func (r Runner) base(tool string) model.Check {
	c := model.Check{Tool: tool, Status: "passed", Diagnostics: []model.Diagnostic{}}
	out, e := r.command(tool, "--version")
	if e != nil {
		c.Status = "error"
		c.Error = fmt.Sprintf("tool unavailable: %v", e)
	} else {
		c.Version = strings.Split(strings.TrimSpace(out), "\n")[0]
	}
	return c
}
func finish(c *model.Check) {
	if c.Status != "error" && len(c.Diagnostics) > 0 {
		c.Status = "findings"
	}
}
func (r Runner) Format(files []string) model.Check {
	c := r.base("clang-format")
	c.Total = len(files)
	if c.Status == "error" {
		return c
	}
	for _, f := range files {
		original, e := os.ReadFile(filepath.Join(r.Root, f))
		if e != nil {
			c.Status = "error"
			c.Error = e.Error()
			continue
		}
		// Output the formatted source rather than modifying the checkout.
		out, e := r.command(c.Tool, "--style=file", "--fallback-style="+r.Config.FormatStyle, "--", f)
		if e != nil {
			c.Status = "error"
			c.Error = fmt.Sprintf("%s: %v: %s", f, e, out)
			continue
		}
		c.Analyzed++
		if !bytes.Equal(original, []byte(out)) {
			c.Diagnostics = append(c.Diagnostics, model.Diagnostic{Tool: c.Tool, Rule: "format", Severity: "warning", File: f, Line: 1, Column: 1, Message: "File differs from configured clang-format style"})
		}
	}
	finish(&c)
	return c
}

var tidyLine = regexp.MustCompile(`^(.+?):(\d+):(\d+): (warning|error|fatal error): (.*?)(?: \[([^\]]+)\])?$`)

func ParseTidy(root, out string, exclude []string) ([]model.Diagnostic, bool) {
	ds := []model.Diagnostic{}
	broken := false
	for _, line := range strings.Split(out, "\n") {
		m := tidyLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		sev := m[4]
		rule := m[6]
		if sev == "error" || sev == "fatal error" || strings.HasPrefix(rule, "clang-diagnostic-error") {
			broken = true
			sev = "error"
		}
		p := m[1]
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		rel, e := project.Relative(root, p)
		if e != nil || project.Excluded(rel, exclude) {
			continue
		}
		n, _ := strconv.Atoi(m[2])
		col, _ := strconv.Atoi(m[3])
		if strings.HasPrefix(rule, "clang-analyzer-") {
			sev = "error"
		}
		if rule == "" {
			rule = "clang-diagnostic"
		}
		ds = append(ds, model.Diagnostic{Tool: "clang-tidy", Rule: rule, Severity: sev, File: rel, Line: n, Column: col, Message: m[5]})
	}
	return ds, broken
}
func (r Runner) Tidy(cmds []project.Command) model.Check {
	c := r.base("clang-tidy")
	c.Total = len(cmds)
	if c.Status == "error" {
		return c
	}
	// One-entry databases preserve multiple configurations of the same file.
	jobs := make(chan project.Command)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < r.Config.Jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for unit := range jobs {
				ds, broken, e := r.tidyUnit(unit)
				mu.Lock()
				c.Diagnostics = append(c.Diagnostics, ds...)
				if e != nil || broken {
					c.Status = "error"
					if e != nil {
						c.Error = e.Error()
					} else {
						c.Error = "compiler diagnostics prevented complete analysis"
					}
				} else {
					c.Analyzed++
				}
				mu.Unlock()
			}
		}()
	}
	for _, unit := range cmds {
		jobs <- unit
	}
	close(jobs)
	wg.Wait()
	finish(&c)
	return c
}
func (r Runner) tidyUnit(unit project.Command) ([]model.Diagnostic, bool, error) {
	dir, e := os.MkdirTemp("", "cppreportcard-compdb-")
	if e != nil {
		return nil, false, e
	}
	defer os.RemoveAll(dir)
	data, e := jsonCommands([]project.Command{unit})
	if e != nil {
		return nil, false, e
	}
	if e = os.WriteFile(filepath.Join(dir, "compile_commands.json"), data, 0600); e != nil {
		return nil, false, e
	}
	out, e := r.command("clang-tidy", "-p", dir, "--checks="+r.Config.Checks, unit.File)
	ds, b := ParseTidy(r.Root, out, r.Config.Exclude)
	if e != nil {
		return ds, b, fmt.Errorf("clang-tidy: %v: %s", e, out)
	}
	return ds, b, nil
}

type cppXML struct {
	XMLName xml.Name `xml:"results"`
	Errors  []struct {
		ID        string `xml:"id,attr"`
		Severity  string `xml:"severity,attr"`
		Message   string `xml:"msg,attr"`
		Locations []struct {
			File   string `xml:"file,attr"`
			Line   int    `xml:"line,attr"`
			Column int    `xml:"column,attr"`
		} `xml:"location"`
	} `xml:"errors>error"`
}

func ParseCppcheck(root, out string, exclude []string) ([]model.Diagnostic, bool, error) {
	start := strings.Index(out, "<?xml")
	if start < 0 {
		start = strings.Index(out, "<results")
	}
	if start < 0 {
		return nil, false, fmt.Errorf("cppcheck did not emit XML")
	}
	var v cppXML
	if e := xml.Unmarshal([]byte(out[start:]), &v); e != nil {
		return nil, false, e
	}
	ds := []model.Diagnostic{}
	broken := false
	for _, x := range v.Errors {
		switch x.ID {
		case "syntaxError", "internalError", "preprocessorErrorDirective", "noValidConfiguration", "missingInclude", "missingIncludeSystem":
			broken = true
		}
		if len(x.Locations) == 0 {
			if x.Severity == "error" {
				broken = true
			}
			continue
		}
		p := x.Locations[0].File
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		rel, e := project.Relative(root, p)
		if e != nil || project.Excluded(rel, exclude) {
			continue
		}
		sev := x.Severity
		if sev != "error" {
			sev = "warning"
		}
		ds = append(ds, model.Diagnostic{Tool: "cppcheck", Rule: x.ID, Severity: sev, File: rel, Line: x.Locations[0].Line, Column: x.Locations[0].Column, Message: x.Message})
	}
	return ds, broken, nil
}
func (r Runner) Cppcheck(files []string, cmds []project.Command) model.Check {
	c := r.base("cppcheck")
	if c.Status == "error" {
		return c
	}
	dir, e := os.MkdirTemp("", "cppreportcard-cppcheck-")
	if e != nil {
		c.Status = "error"
		c.Error = e.Error()
		return c
	}
	defer os.RemoveAll(dir)
	args := []string{"--quiet", "--xml", "--xml-version=2", "--enable=warning,performance,portability,information", "--inline-suppr", "--error-exitcode=0", "-j", strconv.Itoa(r.Config.Jobs)}
	if len(cmds) > 0 {
		data, err := jsonCommands(cmds)
		e = err
		if e == nil {
			e = os.WriteFile(filepath.Join(dir, "compile_commands.json"), data, 0600)
		}
		if e != nil {
			c.Status = "error"
			c.Error = e.Error()
			return c
		}
		args = append(args, "--project="+filepath.Join(dir, "compile_commands.json"))
		c.Total = len(cmds)
	} else {
		var b strings.Builder
		for _, f := range files {
			b.WriteString(filepath.Join(r.Root, f))
			b.WriteByte('\n')
		}
		e = os.WriteFile(filepath.Join(dir, "files.txt"), []byte(b.String()), 0600)
		if e != nil {
			c.Status = "error"
			c.Error = e.Error()
			return c
		}
		args = append(args, "--file-list="+filepath.Join(dir, "files.txt"))
		c.Total = len(files)
	}
	out, e := r.command(c.Tool, args...)
	ds, broken, parseErr := ParseCppcheck(r.Root, out, r.Config.Exclude)
	c.Diagnostics = ds
	if e != nil || parseErr != nil || broken {
		c.Status = "error"
		c.Error = fmt.Sprintf("incomplete cppcheck analysis (execution=%v, parse=%v, configuration=%t)", e, parseErr, broken)
	} else {
		c.Analyzed = c.Total
	}
	finish(&c)
	return c
}
