package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePlan(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "plan.json")
	plan, err := ParsePlan(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Resources) != 2 {
		t.Fatalf("resources: got %d want 2", len(plan.Resources))
	}
	if plan.Resources[0].Address != "aws_s3_bucket.audit_logs" {
		t.Fatalf("unexpected address %q", plan.Resources[0].Address)
	}
	dig := SHA256Digest(plan.RawBytes)
	if dig == "" || dig[:7] != "sha256:" {
		t.Fatalf("bad digest %q", dig)
	}
}

func TestParsePlanMissingFile(t *testing.T) {
	_, err := ParsePlan(filepath.Join(t.TempDir(), "nope.json"))
	if err == nil {
		t.Fatal("expected error for missing plan file")
	}
}

func TestParsePlanInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plan.json")
	if err := os.WriteFile(path, []byte(`{broken`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParsePlan(path)
	if err == nil {
		t.Fatal("expected error for invalid plan json")
	}
}
