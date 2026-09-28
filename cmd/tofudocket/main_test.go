package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAuditCLI(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	initGit(t, dir)
	root, _ := os.Getwd()
	planFixture := filepath.Join(root, "..", "..", "fixtures", "plan.json")
	planData, err := os.ReadFile(planFixture)
	if err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(dir, "plan.json")
	if err := os.WriteFile(planPath, planData, 0o644); err != nil {
		t.Fatal(err)
	}
	c := exec.Command("git", "-C", dir, "add", "plan.json")
	_, _ = c.CombinedOutput()
	c = exec.Command("git", "-C", dir, "commit", "-m", "plan")
	if o, e := c.CombinedOutput(); e != nil {
		t.Fatalf("commit plan: %v %s", e, o)
	}
	outPath := filepath.Join(dir, "evidence.json")
	cmd := exec.Command("go", "run", ".", "audit",
		"--plan", planPath,
		"--repo", dir,
		"--output", outPath,
	)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("audit: %v\n%s", err, out)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatal(err)
	}
}

func initGit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
	} {
		c := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if o, e := c.CombinedOutput(); e != nil {
			t.Fatalf("git: %v %s", e, o)
		}
	}
	_ = os.WriteFile(filepath.Join(dir, "f"), []byte("1"), 0o644)
	c := exec.Command("git", "-C", dir, "add", "f")
	_, _ = c.CombinedOutput()
	c = exec.Command("git", "-C", dir, "commit", "-m", "m")
	if o, e := c.CombinedOutput(); e != nil {
		t.Fatalf("commit: %v %s", e, o)
	}
}
