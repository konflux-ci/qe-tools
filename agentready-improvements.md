# AgentReady Improvement Plan

**Baseline score**: 64.4/100 Silver (26 assessed, 7 skipped/N/A)
**Branch**: KFLUXDP-982 (commit 076b1ef3)
**Date**: 2026-05-19

**Passes** (15): claude_md_file (100), test_execution (100), lock_files (100), one_command_setup (100), separation_of_concerns (100), structured_logging (100), file_size_limits (100), inline_documentation (100), type_annotations (95), standard_layout (95), deterministic_enforcement (60), container_setup (50), dependency_security (35)

**Fails** (13): single_file_verification (0), conventional_commits (0), repomix_config (0), architecture_decisions (0), openapi_specs (0), code_smells (0), issue_pr_templates (0), gitignore_completeness (8), design_intent (30), concise_documentation (64), readme_structure (67), ci_quality_gates (80)

Actions are ordered by effort-to-impact ratio. Items marked **N/A for this repo** are false positives from a tool not fully tuned to Go CLI projects.

---

## Tier 1 -- Do This Week

### Single-File Verification (0/100, +5 pts potential)

**Action**: Add single-file lint and vet commands to CLAUDE.md Build section. The commands work today but aren't documented in the format the tool expects.

```bash
# Lint a single file
golangci-lint run --config .golang-ci.yml cmd/prowjob/periodicSlackReport.go

# Vet a single package
go vet ./cmd/prowjob/...

# Run tests for a single package
go test ./cmd/prowjob/... -v
```

### README Structure (67/100, +5 pts potential)

**Action**: README is 21 lines with only Installation and Development sections. Add a Usage section with examples of running subcommands. The content exists in CLAUDE.md but the tool checks README specifically.

    ## Usage

    ```bash
    # Estimate PR review time
    qe-tools estimate-review --owner konflux-ci --repo qe-tools --pr-number 42

    # Generate Prow job report
    ARTIFACT_DIR=./output qe-tools prowjob create-report

    # Download OCI artifacts
    qe-tools download --repository quay.io/org/repo --tag latest
    ```

### CI Quality Gates (80/100, +5 pts potential)

**Action**: The repo has lint (`lint.yml`) and test (`test.yml`) gates but no type-check gate. Go doesn't have a separate type-check step (the compiler does it), but the tool expects one. Options:

1. **Quick fix**: Add `go vet ./...` as a named step in the lint workflow (go vet does static analysis including type checking)
2. **Alternative**: Add `staticcheck` as a separate CI job, which the tool is more likely to detect

---

## Tier 2 -- Do This Month

### Conventional Commits (0/100, +3 pts potential)

**Action**: The repo already has `commitlint.yml` in CI and `.pre-commit-config.yaml`, but the tool didn't detect the commitlint configuration. The workflow uses `wagoid/commitlint-github-action` which reads from a separate config. Options:

1. Add `.commitlintrc.json` to the repo root (the tool checks for this file):
```json
{
  "extends": ["@commitlint/config-conventional"]
}
```
2. Or add `conventional-pre-commit` to `.pre-commit-config.yaml`

### .gitignore Completeness (8/100, +2 pts potential)

**Action**: Only 1/12 expected patterns present. Add Go-specific and editor patterns:

```
# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
cover.out

# IDE
.idea/
.vscode/
*.swp
*.swo
```

### Concise Documentation (64/100)

**Action**: README has 3 headings in 21 lines (14.3 headings per 100 lines, target 3-5). The issue is that the README is very short, making the heading ratio look inflated. Adding the Usage section (above) will naturally fix this ratio by adding content.

---

## Tier 3 -- Do This Quarter

### Architecture Decision Records (0/100)

**Action**: Create `docs/adr/` with initial decisions. Good candidates:

- ADR-0001: Cobra + Viper pattern for all subcommands
- ADR-0002: OCI artifact handling via ORAS library
- ADR-0003: Prow job artifact scanning from GCS buckets

### Design Intent (30/100)

**Action**: CLAUDE.md Design Choices section partially covers this but lacks preconditions/invariants format. Add `docs/design/` with:

- `oci-architecture.md` -- invariants for the OCI controller (Quay-only, semaphore concurrency, tag pagination)
- `review-estimation.md` -- preconditions for the estimation formula (config file location, GitHub token scope)

### Repomix Config (0/100)

**Action**: Low priority. Repomix generates AI-friendly context from code. If desired: `agentready repomix-generate --init`. The CLAUDE.md + AGENTS.md files already serve this purpose for this repo.

### OpenAPI Specs (0/100)

**N/A for this repo.** This is a CLI tool, not a web API. No REST endpoints to document. Can be excluded via `.agentready-config.yaml`:

```yaml
excluded_attributes:
  - openapi_specs
```

---

## Tier 4 -- Low Priority

### Code Smells / Linters (0/100)

**Action**: The repo uses `golangci-lint` with `.golang-ci.yml` but the tool didn't detect it (expects standard config file names). Options:
1. Rename `.golang-ci.yml` to `.golangci.yml` (standard name)
2. Add `actionlint` for GitHub Actions and `markdownlint` for docs

### Issue & PR Templates (0/100)

**Action**: Add `.github/PULL_REQUEST_TEMPLATE.md` and `.github/ISSUE_TEMPLATE/` with bug report and feature request templates. Standard GitHub hygiene.

### Container Setup (50/100)

**Action**: Dockerfile exists and scores 50. To reach 100, add a `docker-compose.yml` or document container usage in README. Low impact for agent readiness.

---

## Estimated Score After Improvements

| Change | Points | Effort |
|--------|--------|--------|
| Single-file verification (document commands) | +5 | 10 min |
| README Usage section | +5 | 15 min |
| CI type-check gate (go vet step) | +2-5 | 15 min |
| Conventional commits config file | +3 | 5 min |
| .gitignore completeness | +1-2 | 5 min |
| ADRs (initial 1-3) | +1-2 | 1 hr |

**Estimated new score: ~83-86/100 (Gold)** with Tier 1+2 actions.
OpenAPI exclusion would remove a false-positive failure and potentially push higher.

---

## Not Applicable / False Positives

| Attribute | Reason |
|-----------|--------|
| OpenAPI Specs | CLI tool, no REST API |
| dbt (4 attributes) | Not a dbt project |
| Progressive Disclosure | Not applicable to Go |
| Branch Protection | Requires GitHub API (can't assess locally) |
| Cyclomatic Complexity | Skipped (missing `gocyclo` tool) |
