package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCompilationDatabaseValidation(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "a.cpp")
	if err := os.WriteFile(file, []byte("int main(){}"), 0600); err != nil {
		t.Fatal(err)
	}
	valid := Command{Directory: root, File: "a.cpp", Arguments: []string{"c++", "-c", "a.cpp"}}
	for _, tc := range []struct {
		name     string
		commands []Command
		raw      string
		count    int
	}{
		{"valid", []Command{valid, valid}, "", 1},
		{"variants", []Command{valid, {Directory: root, File: file, Command: "c++ -DALT -c a.cpp"}}, "", 2},
		{"bad JSON", nil, "{", 0}, {"empty", nil, "[]", 0},
		{"relative directory", []Command{{Directory: ".", File: file, Command: "cc"}}, "", 0},
		{"missing directory", []Command{{Directory: filepath.Join(root, "missing"), File: file, Command: "cc"}}, "", 0},
		{"non-directory", []Command{{Directory: file, File: file, Command: "cc"}}, "", 0},
		{"missing command", []Command{{Directory: root, File: file}}, "", 0},
		{"missing source", []Command{{Directory: root, File: "missing.cpp", Command: "cc"}}, "", 0},
		{"outside", []Command{{Directory: root, File: filepath.Join(filepath.Dir(root), "outside.cpp"), Command: "cc"}}, "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := []byte(tc.raw)
			if tc.raw == "" {
				b, _ = json.Marshal(tc.commands)
			}
			p := filepath.Join(root, "compile_commands.json")
			if err := os.WriteFile(p, b, 0600); err != nil {
				t.Fatal(err)
			}
			got, err := LoadCommands(root, p, nil)
			if tc.count == 0 {
				if err == nil {
					t.Fatalf("accepted %+v", tc)
				}
			} else if err != nil || len(got) != tc.count || !filepath.IsAbs(got[0].File) {
				t.Fatalf("%+v %v", got, err)
			}
		})
	}
	b, _ := json.Marshal([]Command{valid})
	p := filepath.Join(root, "compile_commands.json")
	os.WriteFile(p, b, 0600)
	if _, err := LoadCommands(root, "compile_commands.json", []string{"a.cpp"}); err == nil {
		t.Fatal("excluded source kept")
	}
	if _, err := LoadCommands(root, "missing.json", nil); err == nil {
		t.Fatal("missing database accepted")
	}
}
