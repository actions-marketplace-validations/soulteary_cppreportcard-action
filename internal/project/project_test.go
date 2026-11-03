package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscovery(t *testing.T) {
	d := t.TempDir()
	for _, p := range []string{"src/a.cpp", "third_party/b.cpp", "include/a.hpp", "build/c.cpp"} {
		if e := os.MkdirAll(filepath.Dir(filepath.Join(d, p)), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(filepath.Join(d, p), nil, 0644); e != nil {
			t.Fatal(e)
		}
	}
	f, e := Discover(d, []string{"third_party", "build"})
	if e != nil || len(f) != 2 {
		t.Fatalf("%v %v", f, e)
	}
	if _, e = Relative(d, filepath.Dir(d)); e == nil {
		t.Fatal("accepted outside path")
	}
}
func TestEmptyDatabase(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "compile_commands.json")
	os.WriteFile(p, []byte("[]"), 0644)
	if _, e := LoadCommands(d, p, nil); e == nil {
		t.Fatal("accepted empty database")
	}
}
