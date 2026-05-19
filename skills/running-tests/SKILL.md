---
name: running-tests
description: Use when running, writing, or troubleshooting unit tests for qe-tools commands or packages
---

# Running Tests

## Overview

All tests are Go unit tests using stdlib `testing`. Tests live next to source (`foo.go` -> `foo_test.go`). External APIs (GitHub, Slack, Quay) are always mocked -- no live services needed.

## When to Use

- Running tests after making changes
- Writing tests for a new or modified command
- Adding regression tests for a bug fix
- **Not for**: Lint or formatting (use `make lint` / `make fmt`)

## Quick Reference

| Action | Command |
|--------|---------|
| Run all tests | `make test` |
| Run single package | `go test ./cmd/prowjob/... -v` |
| Run single test | `go test ./cmd/prowjob/... -v -run TestConstructMessage` |
| Run with race detector | `make cover` |
| Check coverage | `go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out` |

## Writing Tests

### File placement

Test file goes next to the source it tests:

```
cmd/mycommand/
  mycommand.go          # Command implementation
  mycommand_test.go     # Tests for this command
pkg/mypackage/
  mypackage.go          # Library code
  mypackage_test.go     # Tests for this package
```

### Table-driven pattern

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "valid input", input: "hello", want: "HELLO", wantErr: false},
        {name: "empty input", input: "", want: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

### Mocking external APIs

See `cmd/estimate/reviewTime_test.go` for the exemplar: uses `go-github` types directly in table-driven tests, mocks PR file lists with `github.CommitFile` structs, and tests both the estimation formula and label classification without calling GitHub.

For commands calling Slack or OCI registries, define an interface around the external call, implement a mock struct, and inject it in tests.

### What to test

- **Bug fixes**: Regression test with the exact input that caused the bug
- **Commands**: Flag registration, PreRunE validation, business logic
- **Regex parsing**: Both matching and non-matching inputs -- validate capture group length before indexing
- **Error paths**: API failures, missing env vars, malformed input
- **Boundary cases**: Zero values, nil slices, empty strings

## Exemplar

Reference test file: `cmd/estimate/reviewTime_test.go` -- table-driven, covers file extension mapping, estimation formula, label classification, and command registration. Shows the pattern for testing a command with GitHub API types without calling real services.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Testing only the happy path | Always include negative/edge cases (nil, empty, error) |
| Calling real APIs in tests | Mock everything external via interfaces |
| Skipping tests for bug fixes | Every fix needs a regression test proving it works |
| Not testing regex match safety | Test with input that doesn't match the pattern (see `constructMessage` in `periodicSlackReport.go`) |
| Forgetting to run `make test` | Always verify all tests pass before finishing |
