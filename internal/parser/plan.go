package parser

import (
	"encoding/json"
	"fmt"
	"os"
)

// PlanSummary holds parsed resource change actions from a tofu/terraform show -json plan.
type PlanSummary struct {
	FormatVersion   string
	TerraformVersion string
	Resources       []ResourceAction
	RawBytes        []byte
}

type ResourceAction struct {
	Address string   `json:"address"`
	Actions []string `json:"actions"`
}

type planJSON struct {
	FormatVersion    string `json:"format_version"`
	TerraformVersion string `json:"terraform_version"`
	ResourceChanges  []struct {
		Address string `json:"address"`
		Change  struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

// ParsePlan reads and parses an OpenTofu/Terraform JSON plan file.
func ParsePlan(path string) (*PlanSummary, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read plan: %w", err)
	}
	var p planJSON
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("parse plan json: %w", err)
	}
	out := &PlanSummary{
		FormatVersion:    p.FormatVersion,
		TerraformVersion: p.TerraformVersion,
		RawBytes:         raw,
	}
	for _, rc := range p.ResourceChanges {
		out.Resources = append(out.Resources, ResourceAction{
			Address: rc.Address,
			Actions: append([]string(nil), rc.Change.Actions...),
		})
	}
	return out, nil
}
