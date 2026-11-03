package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRejectInvalidConfig(t *testing.T) {
	for _, data := range []string{`{"unknown":true}`, `{} {}`, `{"mode":"oops"}`, `{"jobs":0}`, `{"tools":null}`} {
		p := filepath.Join(t.TempDir(), "config.json")
		os.WriteFile(p, []byte(data), 0644)
		if _, e := Load(p); e == nil {
			t.Errorf("accepted %s", data)
		}
	}
}
