package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sysrqio/tofudocket/internal/evidence"
	"github.com/sysrqio/tofudocket/internal/format"
	"github.com/sysrqio/tofudocket/internal/score"
)

var version = "0.1.0"

func main() {
	if err := newRoot().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "tofudocket",
		Short: "Local OpenTofu plan evidence for DORA change audit",
	}
	root.AddCommand(newAuditCmd(), newVersionCmd(), newValidateCmd())
	return root
}

func newAuditCmd() *cobra.Command {
	var (
		planPath      string
		lockfilePath  string
		repoPath      string
		conftestPath  string
		driftExitCode int
		outputPath    string
		formatOpt     string
	)
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Generate DORA change evidence from a tofu plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			if planPath == "" {
				return fmt.Errorf("--plan is required")
			}
			doc, err := evidence.Build(evidence.BuildOptions{
				PlanPath:      planPath,
				LockfilePath:  lockfilePath,
				RepoPath:      repoPath,
				ConftestPath:  conftestPath,
				DriftExitCode: driftExitCode,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "audit error: %v\n", err)
				os.Exit(1)
			}
			if err := evidence.WriteJSON(outputPath, doc); err != nil {
				fmt.Fprintf(os.Stderr, "audit error: %v\n", err)
				os.Exit(1)
			}
			if formatOpt == "table" {
				format.PrintTable(os.Stdout, doc)
			} else if formatOpt != "json" {
				return fmt.Errorf("unsupported --format %q (use json or table)", formatOpt)
			}
			if doc.DoraReadinessScore < score.PassThreshold {
				os.Exit(2)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&planPath, "plan", "", "Path to plan.json (tofu show -json) (required)")
	cmd.Flags().StringVar(&lockfilePath, "lockfile", ".terraform.lock.hcl", "Provider lock file path")
	cmd.Flags().StringVar(&repoPath, "repo", ".", "Git repository path")
	cmd.Flags().StringVar(&conftestPath, "conftest", "", "Optional Conftest SARIF/JSON output")
	cmd.Flags().IntVar(&driftExitCode, "drift-exit-code", 0, "Exit code from last plan-only drift check")
	cmd.Flags().StringVar(&outputPath, "output", "evidence.json", "Evidence output path")
	cmd.Flags().StringVar(&formatOpt, "format", "json", "Stdout format: json or table")
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("tofudocket %s (OpenTofu/Terraform plan.json compatible)\n", version)
		},
	}
}

func newValidateCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate evidence.json schema locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				path = "evidence.json"
			}
			if err := evidence.ValidateFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "validate error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("evidence.json is valid")
			return nil
		},
	}
	cmd.Flags().StringVarP(&path, "file", "f", "evidence.json", "Path to evidence.json")
	return cmd
}
