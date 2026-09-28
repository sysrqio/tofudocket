package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// PolicyFindings aggregates conftest / policy scanner output.
type PolicyFindings struct {
	Critical int
	High     int
	Details  []PolicyDetail
}

type PolicyDetail struct {
	RuleID  string `json:"rule_id,omitempty"`
	Message string `json:"message,omitempty"`
	Level   string `json:"level,omitempty"`
}

// ParseConftest reads optional Conftest JSON or SARIF v2.1 output.
func ParseConftest(path string) (*PolicyFindings, error) {
	if path == "" {
		return &PolicyFindings{}, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read conftest output: %w", err)
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		return parseConftestJSONArray(raw)
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("parse conftest json: %w", err)
	}
	if _, ok := probe["$schema"]; ok {
		if s, ok2 := probe["version"]; ok2 {
			var ver string
			_ = json.Unmarshal(s, &ver)
			if strings.Contains(ver, "2.1") || strings.HasPrefix(ver, "2.1") {
				return parseSARIF(raw)
			}
		}
		return parseSARIF(raw)
	}
	if _, ok := probe["runs"]; ok {
		return parseSARIF(raw)
	}
	return parseConftestJSONArray(raw)
}

type sarifDoc struct {
	Runs []struct {
		Results []struct {
			Level   string `json:"level"`
			RuleID  string `json:"ruleId"`
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
		} `json:"results"`
	} `json:"runs"`
}

func parseSARIF(raw []byte) (*PolicyFindings, error) {
	var doc sarifDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse sarif: %w", err)
	}
	findings := &PolicyFindings{}
	for _, run := range doc.Runs {
		for _, r := range run.Results {
			lvl := strings.ToLower(r.Level)
			detail := PolicyDetail{
				RuleID:  r.RuleID,
				Message: r.Message.Text,
				Level:   lvl,
			}
			switch lvl {
			case "error", "critical":
				findings.Critical++
			case "warning", "high":
				findings.High++
			default:
				if lvl == "" {
					findings.High++
					detail.Level = "warning"
				}
			}
			findings.Details = append(findings.Details, detail)
		}
	}
	return findings, nil
}

func parseConftestJSONArray(raw []byte) (*PolicyFindings, error) {
	var items []struct {
		Msg     string `json:"msg"`
		Metadata struct {
			Severity string `json:"severity"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse conftest array: %w", err)
	}
	findings := &PolicyFindings{}
	for _, it := range items {
		sev := strings.ToLower(it.Metadata.Severity)
		detail := PolicyDetail{Message: it.Msg, Level: sev}
		switch sev {
		case "critical", "error":
			findings.Critical++
		case "high", "warning":
			findings.High++
		default:
			findings.High++
			detail.Level = "high"
		}
		findings.Details = append(findings.Details, detail)
	}
	return findings, nil
}
