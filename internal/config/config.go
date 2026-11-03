// Package config uses JSON so configuration has no runtime dependencies.
package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type Config struct {
	Mode            string            `json:"mode"`
	CompileCommands string            `json:"compile_commands"`
	Exclude         []string          `json:"exclude"`
	Checks          string            `json:"tidy_checks"`
	FormatStyle     string            `json:"format_style"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
	Jobs            int               `json:"jobs"`
	Tools           map[string]string `json:"tools"`
}

func Default() Config {
	return Config{
		Mode: "basic", CompileCommands: "build/compile_commands.json",
		Exclude: []string{".git", "build", "dist", "vendor", "third_party", "external", "generated"},
		Checks:  "-*,clang-analyzer-*,bugprone-*", FormatStyle: "LLVM", TimeoutSeconds: 120, Jobs: 2,
		Tools: map[string]string{"clang-format": "clang-format", "clang-tidy": "clang-tidy", "cppcheck": "cppcheck"},
	}
}
func Load(path string) (Config, error) {
	c := Default()
	if path != "" {
		f, e := os.Open(path)
		if e != nil {
			return c, e
		}
		defer f.Close()
		d := json.NewDecoder(f)
		d.DisallowUnknownFields()
		if e = d.Decode(&c); e != nil {
			return c, e
		}
		if e = d.Decode(new(any)); e != io.EOF {
			return c, fmt.Errorf("config must contain one JSON object")
		}
	}
	if c.Mode != "basic" && c.Mode != "full" {
		return c, fmt.Errorf("mode must be basic or full")
	}
	if c.TimeoutSeconds < 1 || c.Jobs < 1 || c.Jobs > 64 {
		return c, fmt.Errorf("timeout must be positive and jobs must be 1..64")
	}
	if c.Tools == nil {
		return c, fmt.Errorf("tools must not be null")
	}
	for _, t := range []string{"clang-format", "clang-tidy", "cppcheck"} {
		if c.Tools[t] == "" {
			return c, fmt.Errorf("missing tool: %s", t)
		}
	}
	return c, nil
}
func (c Config) Timeout() time.Duration { return time.Duration(c.TimeoutSeconds) * time.Second }
