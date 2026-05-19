---
name: pr-definition-of-done
description: Use when preparing, reviewing, or finalizing a pull request to the qe-tools repository
---

# PR Definition of Done

## Overview

Checklist for pull requests to qe-tools. Every item must pass before merge.

## Checklist

### Code Quality

- [ ] `make test` passes (all packages, parallel=1)
- [ ] `make lint` passes (golangci-lint with `.golang-ci.yml`)
- [ ] `make fmt` produces no changes (gofumpt + gci)
- [ ] Error handling wraps context: `fmt.Errorf("doing X: %w", err)`
- [ ] No hardcoded secrets or tokens -- use env vars via viper
- [ ] HTTP clients use explicit timeouts (not `http.DefaultClient`)
- [ ] Regex match results checked for length before indexing capture groups
- [ ] Buffered writers (`bufio.NewWriter`) flushed/closed before file handle closes

### Tests

- [ ] New features include unit tests covering positive and negative cases
- [ ] Bug fixes include a regression test proving the fix
- [ ] Previously untested code touched by this PR now has tests
- [ ] Tests use table-driven patterns with descriptive names
- [ ] External APIs are mocked -- no real service calls in tests
- [ ] All code paths in conditional branches are covered (match/no-match, success/failure)

### Cobra + Viper Conventions

- [ ] Flags defined in `init()`, bound to viper for env var fallback
- [ ] Required inputs validated in `PreRunE`
- [ ] Command registered in `cmd/root.go` via `rootCmd.AddCommand()`
- [ ] Shared types/constants in `pkg/types/types.go`

### CI/Automation

- [ ] If adding a new scheduled command, add corresponding GitHub Actions workflow
- [ ] Commit messages follow conventional commits format
- [ ] `AGENTS.md` stays under 60 lines if modified

### Documentation

- [ ] New commands include usage examples in help text (`Long` field or `Example` field)
- [ ] Config file changes reflected in `config/` directory
- [ ] CLAUDE.md updated if new subcommands, data flows, or pitfalls added
