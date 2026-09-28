package evidence

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sysrqio/tofudocket/internal/score"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, dir, "init")
}

func gitCommitAll(t *testing.T, dir, msg string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "add", "-A")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v %s", err, out)
	}
	cmd = exec.Command("git", "-C", dir, "commit", "-m", msg)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", err, out)
	}
}

func TestBuildAndValidate(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	planSrc := filepath.Join("..", "..", "fixtures", "plan.json")
	planData, err := os.ReadFile(planSrc)
	if err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(dir, "plan.json")
	if err := os.WriteFile(planPath, planData, 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, dir, "add plan")
	doc, err := Build(BuildOptions{
		PlanPath: planPath,
		RepoPath: dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if doc.DoraReadinessScore < score.PassThreshold {
		t.Fatalf("score %d below threshold", doc.DoraReadinessScore)
	}
	out := filepath.Join(dir, "evidence.json")
	if err := WriteJSON(out, doc); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFile(out); err != nil {
		t.Fatal(err)
	}
}

func TestBuildWithDriftLowScore(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	planPath := filepath.Join(dir, "plan.json")
	data, _ := os.ReadFile(filepath.Join("..", "..", "fixtures", "plan.json"))
	_ = os.WriteFile(planPath, data, 0o644)
	sarif := filepath.Join(dir, "policy.sarif.json")
	_ = os.WriteFile(sarif, []byte(`{"version":"2.1.0","runs":[{"results":[
		{"level":"error","ruleId":"R1","message":{"text":"a"}},
		{"level":"error","ruleId":"R2","message":{"text":"b"}}
	]}]}`), 0o644)
	gitCommitAll(t, dir, "audit inputs")
	doc, err := Build(BuildOptions{
		PlanPath:      planPath,
		RepoPath:      dir,
		DriftExitCode: 2,
		ConftestPath:  sarif,
	})
	if err != nil {
		t.Fatal(err)
	}
	if doc.DoraReadinessScore != 50 {
		t.Fatalf("expected score 50, got %d", doc.DoraReadinessScore)
	}
	_ = os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("x"), 0o644)
	doc2, err := Build(BuildOptions{
		PlanPath:      planPath,
		RepoPath:      dir,
		DriftExitCode: 2,
		ConftestPath:  sarif,
	})
	if err != nil {
		t.Fatal(err)
	}
	if doc2.DoraReadinessScore >= score.PassThreshold {
		t.Fatalf("expected score < %d with dirty repo, got %d", score.PassThreshold, doc2.DoraReadinessScore)
	}
}
