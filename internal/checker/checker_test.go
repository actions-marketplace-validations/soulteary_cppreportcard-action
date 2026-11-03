package checker

import (
	"path/filepath"
	"testing"
)

func TestParsers(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "x.cpp") + ":3:4: warning: bad thing [bugprone-test]\n"
	ds, b := ParseTidy(root, out, nil)
	if b || len(ds) != 1 || ds[0].Rule != "bugprone-test" {
		t.Fatalf("%+v %v", ds, b)
	}
	_, b = ParseTidy(root, filepath.Join(root, "x.cpp")+":3:4: fatal error: header missing", nil)
	if !b {
		t.Fatal("missed compile error")
	}
	ds, b, e := ParseCppcheck(root, `<?xml version="1.0"?><results><errors><error id="nullPointer" severity="error" msg="null"><location file="x.cpp" line="4" column="2"/></error></errors></results>`, nil)
	if e != nil || b || len(ds) != 1 || ds[0].Severity != "error" {
		t.Fatalf("%+v %v %v", ds, b, e)
	}
	if _, _, e = ParseCppcheck(root, "broken", nil); e == nil {
		t.Fatal("accepted missing XML")
	}
}

func TestAnalyzerSeverityAndMalformedXML(t *testing.T) {
	root := t.TempDir()
	ds, broken := ParseTidy(root, filepath.Join(root, "x.cpp")+":1:1: warning: null pointer [clang-analyzer-core.NullDereference]", nil)
	if broken || len(ds) != 1 || ds[0].Severity != "error" {
		t.Fatalf("%+v %v", ds, broken)
	}
	if _, _, e := ParseCppcheck(root, `<?xml version="1.0"?><wrong/>`, nil); e == nil {
		t.Fatal("wrong root accepted")
	}
}
