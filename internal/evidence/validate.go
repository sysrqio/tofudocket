package evidence

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ValidateFile checks evidence.json against the local schema rules.
func ValidateFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read evidence: %w", err)
	}
	var doc Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse evidence json: %w", err)
	}
	if doc.SchemaVersion == "" {
		return fmt.Errorf("missing schema_version")
	}
	if doc.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q (want %s)", doc.SchemaVersion, SchemaVersion)
	}
	if doc.GeneratedAt == "" {
		return fmt.Errorf("missing generated_at")
	}
	if doc.PlanDigest == "" || !strings.HasPrefix(doc.PlanDigest, "sha256:") {
		return fmt.Errorf("invalid or missing plan_digest")
	}
	if doc.Git.Commit == "" {
		return fmt.Errorf("missing git.commit")
	}
	if doc.DoraReadinessScore < 0 || doc.DoraReadinessScore > 100 {
		return fmt.Errorf("dora_readiness_score out of range")
	}
	if doc.PolicyFindings.Critical < 0 || doc.PolicyFindings.High < 0 {
		return fmt.Errorf("negative policy counts")
	}
	return nil
}
