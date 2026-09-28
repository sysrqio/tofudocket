package parser

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GitBinding captures commit SHA and working tree cleanliness.
type GitBinding struct {
	Commit string
	Dirty  bool
}

// GitRevParse runs git in repoPath and returns HEAD commit and dirty flag.
func GitRevParse(repoPath string) (GitBinding, error) {
	var g GitBinding
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return g, fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	g.Commit = strings.TrimSpace(string(out))

	statusCmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	statusOut, err := statusCmd.Output()
	if err != nil {
		return g, fmt.Errorf("git status: %w", err)
	}
	g.Dirty = len(bytes.TrimSpace(statusOut)) > 0
	return g, nil
}
