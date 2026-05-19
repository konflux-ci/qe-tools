---
name: adding-a-subcommand
description: Use when adding a new CLI subcommand to qe-tools, or when understanding the pattern existing commands follow
---

# Adding a Subcommand

## Overview

Every qe-tools subcommand follows the Cobra + Viper pattern: flags in `init()`, env var fallback via viper, validation in `PreRunE`, execution in `RunE`. Business logic goes in `pkg/`, command wiring in `cmd/`.

## When to Use

- Adding a new subcommand to the CLI
- Understanding how existing commands are structured
- Refactoring a command to follow repo conventions
- **Not for**: Adding a flag to an existing command (just edit the existing `init()`)

## Step-by-Step

### 1. Create the command package

```
cmd/<name>/
  <name>.go           # Command definition + cobra wiring
  <name>_test.go      # Tests for command registration + integration
```

### 2. Create the business logic package (if non-trivial)

```
pkg/<name>/
  <name>.go           # Core logic, interfaces, types
  <name>_test.go      # Unit tests for logic
```

### 3. Wire flags and viper in `init()`

```go
func init() {
    MyCmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner")
    MyCmd.Flags().IntVar(&prNumber, "pr-number", 0, "Pull request number")

    _ = viper.BindEnv("GITHUB_TOKEN", "GITHUB_TOKEN")
    _ = viper.BindPFlag("owner", MyCmd.Flags().Lookup("owner"))
}
```

**Key**: Viper binds lowercase flag names but reads UPPERCASE env vars.

### 4. Validate in `PreRunE`

```go
PreRunE: func(cmd *cobra.Command, args []string) error {
    if viper.GetString("GITHUB_TOKEN") == "" {
        return fmt.Errorf("environment variable GITHUB_TOKEN must be set")
    }
    if owner == "" {
        return fmt.Errorf("--owner is required")
    }
    return nil
},
```

### 5. Register in root.go

```go
import "github.com/konflux-ci/qe-tools/cmd/mycommand"

func init() {
    rootCmd.AddCommand(mycommand.MyCmd)
}
```

### 6. Add config file (if needed)

Place at `config/<name>/config.yaml`. Commands read config relative to CWD -- the Dockerfile copies `config/` to `/qe-tools/config`.

### 7. Add GitHub Actions workflow (if scheduled/automated)

See `.github/workflows/estimate-review.yml` as a reference for PR-triggered commands.

## Exemplar

Reference implementation: `cmd/estimate/reviewTime.go` -- has flags, viper bindings, PreRunE validation, GitHub API integration with interface-based mocking, config file, CI workflow, and comprehensive tests.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Forgetting to register in `root.go` | Binary builds but command doesn't appear in `--help` |
| Using `http.DefaultClient` | Wrap with `&http.Client{Timeout: 30 * time.Second}` |
| Hardcoding config path | Use relative path; respect CWD differences between local dev and container |
| Skipping `PreRunE` validation | Users get cryptic nil-pointer errors instead of clear messages |
| Not adding tests | Every command needs at least: flag registration tests, PreRunE validation tests, and business logic tests |
