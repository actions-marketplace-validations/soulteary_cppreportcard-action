package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Command struct {
	Directory string   `json:"directory"`
	File      string   `json:"file"`
	Arguments []string `json:"arguments,omitempty"`
	Command   string   `json:"command,omitempty"`
}

func Excluded(rel string, exclude []string) bool {
	rel = filepath.ToSlash(rel)
	for _, x := range exclude {
		x = strings.Trim(filepath.ToSlash(x), "/")
		if x == "" {
			continue
		}
		if strings.Contains(x, "/") {
			if rel == x || strings.HasPrefix(rel, x+"/") {
				return true
			}
		} else {
			for _, p := range strings.Split(rel, "/") {
				if p == x {
					return true
				}
			}
		}
	}
	return false
}
func Relative(root, path string) (string, error) {
	rel, e := filepath.Rel(root, path)
	if e != nil {
		return "", e
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path is outside project: %s", path)
	}
	return filepath.ToSlash(rel), nil
}
func Discover(root string, exclude []string) ([]string, error) {
	var files []string
	e := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := Relative(root, p)
		if e != nil {
			return e
		}
		if rel != "." && Excluded(rel, exclude) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(d.Name(), "._") {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(p)) {
		case ".c", ".cc", ".cpp", ".cxx", ".h", ".hh", ".hpp", ".hxx":
			files = append(files, rel)
		}
		return nil
	})
	sort.Strings(files)
	return files, e
}
func LoadCommands(root, path string, exclude []string) ([]Command, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var cmds []Command
	if e = json.Unmarshal(data, &cmds); e != nil {
		return nil, e
	}
	if len(cmds) == 0 {
		return nil, fmt.Errorf("empty compilation database")
	}
	var out []Command
	seen := map[string]bool{}
	for _, c := range cmds {
		if !filepath.IsAbs(c.Directory) {
			return nil, fmt.Errorf("compilation directory must be absolute: %s", c.Directory)
		}
		if info, e := os.Stat(c.Directory); e != nil || !info.IsDir() {
			return nil, fmt.Errorf("unavailable compilation directory: %s", c.Directory)
		}
		p := c.File
		if !filepath.IsAbs(p) {
			p = filepath.Join(c.Directory, p)
		}
		p = filepath.Clean(p)
		rel, e := Relative(root, p)
		if e != nil {
			continue
		}
		if Excluded(rel, exclude) {
			continue
		}
		if c.Command == "" && len(c.Arguments) == 0 {
			return nil, fmt.Errorf("missing compile command for %s", rel)
		}
		if _, e = os.Stat(p); e != nil {
			return nil, e
		}
		c.File = p
		// Each distinct configuration remains an analysis unit.
		key := p + "\x00" + c.Directory + "\x00" + c.Command + strings.Join(c.Arguments, "\x00")
		if !seen[key] {
			seen[key] = true
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no project compilation units after filtering")
	}
	return out, nil
}
