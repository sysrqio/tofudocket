package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConftestSARIF(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "conftest.sarif.json")
	f, err := ParseConftest(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.High != 1 || f.Critical != 0 {
		t.Fatalf("got critical=%d high=%d", f.Critical, f.High)
	}
}

func TestParseConftestEmpty(t *testing.T) {
	f, err := ParseConftest("")
	if err != nil || f.Critical != 0 || f.High != 0 {
		t.Fatalf("empty: %+v err=%v", f, err)
	}
}

func TestParseConftestJSONArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conftest.json")
	raw := `[
		{"msg": "deny public bucket", "metadata": {"severity": "critical"}},
		{"msg": "missing tags", "metadata": {"severity": "high"}},
		{"msg": "info only", "metadata": {"severity": "low"}}
	]`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := ParseConftest(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Critical != 1 || f.High != 2 {
		t.Fatalf("got critical=%d high=%d", f.Critical, f.High)
	}
	if len(f.Details) != 3 {
		t.Fatalf("details: got %d want 3", len(f.Details))
	}
}

func TestParseConftestInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`not json`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseConftest(path)
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestParseConftestMissingFile(t *testing.T) {
	_, err := ParseConftest(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
