//go:build e2e

package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	binPath string
	covDir  string
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "qe-tools-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	binPath = filepath.Join(tmpDir, "qe-tools-cov")
	covDir = filepath.Join(tmpDir, "covdata")

	if err := os.MkdirAll(covDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create covdata dir: %v\n", err)
		os.Exit(1)
	}

	rootDir := findRepoRoot()
	cmd := exec.Command("go", "build", "-cover", "-covermode=atomic", "-o", binPath, ".")
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build instrumented binary: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	// Copy covdata to project tmp/covdata for Makefile to pick up
	projectCovDir := filepath.Join(rootDir, "tmp", "covdata")
	if err := os.MkdirAll(projectCovDir, 0o750); err == nil {
		entries, _ := os.ReadDir(covDir)
		for _, e := range entries {
			data, _ := os.ReadFile(filepath.Join(covDir, e.Name()))
			_ = os.WriteFile(filepath.Join(projectCovDir, e.Name()), data, 0o644)
		}
	}

	os.RemoveAll(tmpDir)
	os.Exit(code)
}

func findRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	panic("could not find repository root (no go.mod found)")
}

type cliResult struct {
	exitCode int
	stdout   string
	stderr   string
}

func runCLI(t *testing.T, args ...string) cliResult {
	t.Helper()
	return runCLIWithEnv(t, nil, args...)
}

func runCLIWithEnv(t *testing.T, extraEnv []string, args ...string) cliResult {
	t.Helper()
	cmd := exec.Command(binPath, args...)

	env := []string{"GOCOVERDIR=" + covDir}
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GOCOVERDIR=") {
			env = append(env, e)
		}
	}
	env = append(env, extraEnv...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("failed to run CLI: %v", err)
	}

	return cliResult{
		exitCode: exitCode,
		stdout:   stdout.String(),
		stderr:   stderr.String(),
	}
}

func TestRootHelp(t *testing.T) {
	r := runCLI(t, "--help")
	if r.exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", r.exitCode)
	}
	if !bytes.Contains([]byte(r.stdout), []byte("Available Commands")) {
		t.Error("--help output should contain 'Available Commands'")
	}
}

func TestProwjobHelp(t *testing.T) {
	r := runCLI(t, "prowjob", "--help")
	if r.exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", r.exitCode)
	}
	if !bytes.Contains([]byte(r.stdout), []byte("create-report")) {
		t.Error("prowjob --help should mention create-report")
	}
}

func TestProwjobCreateReportMissingParams(t *testing.T) {
	r := runCLI(t, "prowjob", "create-report")
	if r.exitCode == 0 {
		t.Error("expected non-zero exit code when prow-job-id is missing")
	}
}

func TestEstimateReviewMissingToken(t *testing.T) {
	r := runCLI(t, "estimate-review", "--pr-number=1", "--repository=foo/bar")
	if r.exitCode == 0 {
		t.Error("expected non-zero exit code without GITHUB_TOKEN")
	}
}

func TestDownloadInvalidRef(t *testing.T) {
	r := runCLI(t, "download", "--oci-ref=invalid-not-quay")
	if r.exitCode == 0 {
		t.Error("expected non-zero exit code for invalid OCI ref")
	}
}

func TestAnalyzeTestResultsMissingOCIRef(t *testing.T) {
	r := runCLI(t, "analyze-test-results")
	if r.exitCode == 0 {
		t.Error("expected non-zero exit code without OCI ref")
	}
}

func TestSendSlackMessageMissingToken(t *testing.T) {
	r := runCLI(t, "send-slack-message", "--channel=test", "--message=hi")
	if r.exitCode == 0 {
		t.Error("expected non-zero exit code without SLACK_TOKEN")
	}
}

func TestCoffeeBreakMissingToken(t *testing.T) {
	r := runCLI(t, "coffee-break")
	if r.exitCode == 0 {
		t.Error("expected non-zero exit code without SLACK_TOKEN")
	}
}

func TestWebhookReportPortalMissingParams(t *testing.T) {
	r := runCLI(t, "webhook", "report-portal")
	if r.exitCode == 0 {
		t.Error("expected non-zero exit code without required params")
	}
}
