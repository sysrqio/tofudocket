# tofudocket

Local OpenTofu and Terraform plan evidence for DORA change audit. The CLI reads a
`plan.json` from `tofu show -json`, binds the change to a Git commit, optionally
ingests Conftest policy output, and writes a signed digest and readiness score to
`evidence.json`. All processing is local; `audit` makes no network calls.

## Requirements

- Go 1.22+
- Git (for commit binding during `audit`)

## Install

```bash
go install github.com/sysrqio/tofudocket/cmd/tofudocket@latest
```

Or build from source:

```bash
make build
./bin/tofudocket version
```

## Usage

Generate a plan (OpenTofu example):

```bash
tofu plan -out=plan.tfplan
tofu show -json plan.tfplan > plan.json
```

Run audit:

```bash
tofudocket audit --plan plan.json --repo . --output evidence.json
tofudocket audit --plan plan.json --conftest policy.sarif.json --format table
tofudocket validate --file evidence.json
tofudocket version
```

### Flags (`audit`)

| Flag | Default | Description |
|------|---------|-------------|
| `--plan` | (required) | Path to JSON plan |
| `--lockfile` | `.terraform.lock.hcl` | Provider lock file path (recorded in evidence) |
| `--repo` | `.` | Git repository root |
| `--conftest` | | Conftest SARIF 2.1 or JSON findings |
| `--drift-exit-code` | `0` | Exit code from a prior drift-only plan step |
| `--output` | `evidence.json` | Output path |
| `--format` | `json` | Stdout: `json` (silent) or `table` |

### Exit codes (`audit`)

| Code | Meaning |
|------|---------|
| 0 | Evidence written; DORA readiness score ≥ 50 |
| 1 | I/O or parse error |
| 2 | Score &lt; 50 (policy/drift/git penalties) |

### DORA readiness score

```
100 - (10 × critical policies) - (5 × high policies)
    - (30 if drift exit code ≠ 0) - (10 if git working tree dirty)
```

Score is clamped to 0–100.

### Evidence schema

See `fixtures/` and the output of `tofudocket validate`. Core fields:

- `schema_version`, `generated_at`
- `git.commit`, `git.dirty`
- `plan_digest` (SHA-256 over plan file bytes)
- `resource_actions[]`
- `policy_findings` (critical/high counts)
- `drift.detected`, `drift.exit_code`
- `dora_readiness_score`

## Development

```bash
make test
make build
```

## License

Apache-2.0 — see [LICENSE](LICENSE).
