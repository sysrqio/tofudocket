package evidence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFileMissing(t *testing.T) {
	err := ValidateFile(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("expected error for missing evidence file")
	}
}

func TestValidateFileInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.json")
	if err := os.WriteFile(path, []byte(`{`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFile(path); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestValidateFileSchemaNegatives(t *testing.T) {
	base := Document{
		SchemaVersion:      SchemaVersion,
		GeneratedAt:        "2026-01-01T00:00:00Z",
		PlanDigest:         "sha256:abc",
		DoraReadinessScore: 80,
		Git:                  GitInfo{Commit: "deadbeef"},
	}
	tests := []struct {
		name string
		mut  func(*Document)
	}{
		{"missing schema_version", func(d *Document) { d.SchemaVersion = "" }},
		{"wrong schema_version", func(d *Document) { d.SchemaVersion = "9.9" }},
		{"missing generated_at", func(d *Document) { d.GeneratedAt = "" }},
		{"bad plan_digest", func(d *Document) { d.PlanDigest = "md5:nope" }},
		{"missing git commit", func(d *Document) { d.Git.Commit = "" }},
		{"score out of range high", func(d *Document) { d.DoraReadinessScore = 101 }},
		{"score out of range low", func(d *Document) { d.DoraReadinessScore = -1 }},
		{"negative critical count", func(d *Document) { d.PolicyFindings.Critical = -1 }},
		{"negative high count", func(d *Document) { d.PolicyFindings.High = -1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "evidence.json")
			doc := base
			tt.mut(&doc)
			if err := WriteJSON(path, &doc); err != nil {
				t.Fatal(err)
			}
			if err := ValidateFile(path); err == nil {
				t.Fatalf("expected validation error for %s", tt.name)
			}
		})
	}
}
