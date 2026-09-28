package evidence

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sysrqio/tofudocket/internal/parser"
	"github.com/sysrqio/tofudocket/internal/score"
)

const SchemaVersion = "1.0"

// Document is the audit evidence payload written to evidence.json.
type Document struct {
	SchemaVersion      string                 `json:"schema_version"`
	GeneratedAt        string                 `json:"generated_at"`
	Git                GitInfo                `json:"git"`
	PlanDigest         string                 `json:"plan_digest"`
	LockfilePath       string                 `json:"lockfile_path,omitempty"`
	ResourceActions    []parser.ResourceAction `json:"resource_actions"`
	PolicyFindings     PolicyBlock            `json:"policy_findings"`
	Drift              DriftInfo              `json:"drift"`
	DoraReadinessScore int                    `json:"dora_readiness_score"`
	PlanMeta           PlanMeta               `json:"plan_meta,omitempty"`
}

type GitInfo struct {
	Commit string `json:"commit"`
	Dirty  bool   `json:"dirty"`
}

type PolicyBlock struct {
	Critical int                    `json:"critical"`
	High     int                    `json:"high"`
	Details  []parser.PolicyDetail  `json:"details"`
}

type DriftInfo struct {
	Detected bool `json:"detected"`
	ExitCode int  `json:"exit_code"`
}

type PlanMeta struct {
	FormatVersion    string `json:"format_version,omitempty"`
	TerraformVersion string `json:"terraform_version,omitempty"`
}

// BuildOptions configures evidence assembly.
type BuildOptions struct {
	PlanPath       string
	LockfilePath   string
	RepoPath       string
	ConftestPath   string
	DriftExitCode  int
}

// Build assembles evidence from parsed inputs.
func Build(opts BuildOptions) (*Document, error) {
	plan, err := parser.ParsePlan(opts.PlanPath)
	if err != nil {
		return nil, err
	}
	git, err := parser.GitRevParse(opts.RepoPath)
	if err != nil {
		return nil, err
	}
	policies, err := parser.ParseConftest(opts.ConftestPath)
	if err != nil {
		return nil, err
	}
	sc := score.Compute(score.Input{
		CriticalPolicies: policies.Critical,
		HighPolicies:     policies.High,
		DriftExitCode:    opts.DriftExitCode,
		GitDirty:         git.Dirty,
	})
	doc := &Document{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Git: GitInfo{
			Commit: git.Commit,
			Dirty:  git.Dirty,
		},
		PlanDigest:      parser.SHA256Digest(plan.RawBytes),
		LockfilePath:    opts.LockfilePath,
		ResourceActions: plan.Resources,
		PolicyFindings: PolicyBlock{
			Critical: policies.Critical,
			High:     policies.High,
			Details:  policies.Details,
		},
		Drift: DriftInfo{
			Detected: opts.DriftExitCode != 0,
			ExitCode: opts.DriftExitCode,
		},
		DoraReadinessScore: sc,
		PlanMeta: PlanMeta{
			FormatVersion:    plan.FormatVersion,
			TerraformVersion: plan.TerraformVersion,
		},
	}
	return doc, nil
}

// WriteJSON marshals doc to path with mode 0644.
func WriteJSON(path string, doc *Document) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal evidence: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write evidence: %w", err)
	}
	return nil
}
