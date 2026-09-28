package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var cliBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "tofudocket-cli-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	cliBin = filepath.Join(dir, "tofudocket")
	build := exec.Command("go", "build", "-o", cliBin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		panic(string(out) + err.Error())
	}
	os.Exit(m.Run())
}

func cmdDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}

func runCLI(t *testing.T, args ...string) (exitCode int, stdout, stderr string) {
	t.Helper()
	cmd := exec.Command(cliBin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("run cli: %v", err)
		}
	}
	return exitCode, outBuf.String(), errBuf.String()
}

func TestAuditCLIMissingPlan(t *testing.T) {
	code, _, stderr := runCLI(t, "audit", "--repo", ".")
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "plan") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestAuditCLIInvalidPlan(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	initGit(t, dir)
	badPlan := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badPlan, []byte(`{not valid`), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runCLI(t, "audit", "--plan", badPlan, "--repo", dir)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "audit error") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestAuditCLIMissingPlanFile(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	initGit(t, dir)
	missing := filepath.Join(dir, "missing-plan.json")
	code, _, _ := runCLI(t, "audit", "--plan", missing, "--repo", dir)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestAuditCLIExit2DriftAndPolicy(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	initGit(t, dir)
	planPath := writePlanFixture(t, dir)
	sarif := filepath.Join(dir, "policy.sarif.json")
	criticals := strings.Repeat(`{"level":"error","ruleId":"R","message":{"text":"x"}},`, 3)
	payload := `{"version":"2.1.0","runs":[{"results":[` + strings.TrimSuffix(criticals, ",") + `]}]}`
	if err := os.WriteFile(sarif, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAddCommit(t, dir, "inputs")
	outPath := filepath.Join(dir, "evidence.json")
	code, _, _ := runCLI(t, "audit",
		"--plan", planPath,
		"--repo", dir,
		"--conftest", sarif,
		"--drift-exit-code", "2",
		"--output", outPath,
	)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (drift + policy)", code)
	}
}

func TestAuditCLIExit2DirtyRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	initGit(t, dir)
	planPath := writePlanFixture(t, dir)
	gitAddCommit(t, dir, "plan")
	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}
	sarif := filepath.Join(dir, "policy.sarif.json")
	payload := `{"version":"2.1.0","runs":[{"results":[
		{"level":"error","ruleId":"R1","message":{"text":"a"}},
		{"level":"error","ruleId":"R2","message":{"text":"b"}}
	]}]}`
	if err := os.WriteFile(sarif, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "evidence.json")
	code, _, _ := runCLI(t, "audit",
		"--plan", planPath,
		"--repo", dir,
		"--conftest", sarif,
		"--drift-exit-code", "2",
		"--output", outPath,
	)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (dirty + drift + policy)", code)
	}
}

func TestAuditCLIExit2PolicyOnly(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	initGit(t, dir)
	planPath := writePlanFixture(t, dir)
	sarif := filepath.Join(dir, "policy.sarif.json")
	criticals := strings.Repeat(`{"level":"error","ruleId":"R","message":{"text":"x"}},`, 6)
	payload := `{"version":"2.1.0","runs":[{"results":[` + strings.TrimSuffix(criticals, ",") + `]}]}`
	if err := os.WriteFile(sarif, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAddCommit(t, dir, "plan")
	outPath := filepath.Join(dir, "evidence.json")
	code, _, _ := runCLI(t, "audit",
		"--plan", planPath,
		"--repo", dir,
		"--conftest", sarif,
		"--output", outPath,
	)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (policy-only low score)", code)
	}
}

func TestVersionCLI(t *testing.T) {
	code, stdout, _ := runCLI(t, "version")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "tofudocket") || !strings.Contains(stdout, "0.1.0") {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestValidateCLISuccess(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	initGit(t, dir)
	planPath := writePlanFixture(t, dir)
	gitAddCommit(t, dir, "plan")
	evidencePath := filepath.Join(dir, "evidence.json")
	code, _, _ := runCLI(t, "audit",
		"--plan", planPath,
		"--repo", dir,
		"--output", evidencePath,
	)
	if code != 0 {
		t.Fatalf("audit setup exit = %d", code)
	}
	code, stdout, _ := runCLI(t, "validate", "-f", evidencePath)
	if code != 0 {
		t.Fatalf("validate exit = %d", code)
	}
	if !strings.Contains(stdout, "valid") {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestValidateCLIMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope.json")
	code, _, stderr := runCLI(t, "validate", "-f", path)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "validate error") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestValidateCLIInvalidSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"9.9"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runCLI(t, "validate", "-f", path)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "validate error") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestAuditCLI(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	initGit(t, dir)
	planPath := writePlanFixture(t, dir)
	gitAddCommit(t, dir, "plan")
	outPath := filepath.Join(dir, "evidence.json")
	code, _, out := runCLI(t, "audit",
		"--plan", planPath,
		"--repo", dir,
		"--output", outPath,
	)
	if code != 0 {
		t.Fatalf("audit: exit %d\n%s", code, out)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatal(err)
	}
}

func writePlanFixture(t *testing.T, dir string) string {
	t.Helper()
	root := cmdDir(t)
	planFixture := filepath.Join(root, "..", "..", "fixtures", "plan.json")
	planData, err := os.ReadFile(planFixture)
	if err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(dir, "plan.json")
	if err := os.WriteFile(planPath, planData, 0o644); err != nil {
		t.Fatal(err)
	}
	return planPath
}

func gitAddCommit(t *testing.T, dir, msg string) {
	t.Helper()
	c := exec.Command("git", "-C", dir, "add", "-A")
	if o, e := c.CombinedOutput(); e != nil {
		t.Fatalf("git add: %v %s", e, o)
	}
	c = exec.Command("git", "-C", dir, "commit", "-m", msg)
	if o, e := c.CombinedOutput(); e != nil {
		t.Fatalf("git commit: %v %s", e, o)
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
