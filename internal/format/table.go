package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/sysrqio/tofudocket/internal/evidence"
)

// PrintTable writes a human-readable summary to w.
func PrintTable(w io.Writer, doc *evidence.Document) {
	var b strings.Builder
	b.WriteString("tofudocket audit summary\n")
	b.WriteString(strings.Repeat("-", 40) + "\n")
	fmt.Fprintf(&b, "DORA readiness score: %d\n", doc.DoraReadinessScore)
	fmt.Fprintf(&b, "Plan digest:          %s\n", doc.PlanDigest)
	fmt.Fprintf(&b, "Git commit:           %s (dirty=%v)\n", doc.Git.Commit, doc.Git.Dirty)
	fmt.Fprintf(&b, "Drift detected:       %v (exit=%d)\n", doc.Drift.Detected, doc.Drift.ExitCode)
	fmt.Fprintf(&b, "Policy critical/high: %d / %d\n", doc.PolicyFindings.Critical, doc.PolicyFindings.High)
	b.WriteString("\nResource actions (top 5):\n")
	limit := 5
	if len(doc.ResourceActions) < limit {
		limit = len(doc.ResourceActions)
	}
	for i := 0; i < limit; i++ {
		ra := doc.ResourceActions[i]
		fmt.Fprintf(&b, "  %s  [%s]\n", ra.Address, strings.Join(ra.Actions, ","))
	}
	if len(doc.ResourceActions) > 5 {
		fmt.Fprintf(&b, "  ... and %d more\n", len(doc.ResourceActions)-5)
	}
	_, _ = io.WriteString(w, b.String())
}
