package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFailuresAndTimeout(t *testing.T) {
	dir := t.TempDir()
	for _, path := range []string{filepath.Join(dir, "missing"), dir} {
		if _, err := Load(path); err == nil {
			t.Errorf("accepted %s", path)
		}
	}
	for _, text := range []string{`{`, `{"unknown":1}`, `{} {}`, `{"mode":"bad"}`, `{"timeout_seconds":0}`, `{"jobs":0}`, `{"jobs":65}`, `{"tools":null}`, `{"tools":{"clang-format":""}}`} {
		p := filepath.Join(dir, "config.json")
		if err := os.WriteFile(p, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(p); err == nil {
			t.Errorf("accepted %s", text)
		}
	}
	c := Default()
	if c.Timeout() != time.Duration(c.TimeoutSeconds)*time.Second {
		t.Fatal("incorrect timeout")
	}
}
