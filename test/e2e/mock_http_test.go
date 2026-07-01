//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPeriodicReportWithMockServer(t *testing.T) {
	buildLog := `
Some preamble output
Reporting job state 'succeeded'
Ran 50 of 100 Specs in 120.5 seconds
PASS! -- 50 Passed | 0 Failed | 0 Pending | 50 Skipped
Ran for 2h0m30s
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/build-log.txt") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(buildLog))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	r := runCLIWithEnv(t, []string{"PROW_URL=" + server.URL}, "prowjob", "periodic-report")
	if r.exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", r.exitCode, r.stderr)
	}
	if !strings.Contains(r.stdout, "Job Succeeded") {
		t.Errorf("expected 'Job Succeeded' in output, got: %s", r.stdout)
	}
}

func TestPeriodicReportFailedJob(t *testing.T) {
	buildLog := `
Some preamble
Reporting job state 'failed'
Ran 10 of 20 Specs in 45.5 seconds
FAIL! -- 8 Passed | 2 Failed | 0 Pending | 10 Skipped
Ran for 0h45m30s
Summarizing 2 Failures:
[FAIL] [It] should create a component successfully
[FAIL] [It] should deploy the application
Test Suite Failed
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(buildLog))
	}))
	defer server.Close()

	r := runCLIWithEnv(t, []string{"PROW_URL=" + server.URL}, "prowjob", "periodic-report")
	if r.exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", r.exitCode, r.stderr)
	}
	if !strings.Contains(r.stdout, "Test Results:") {
		t.Errorf("expected 'Test Results:' in output, got: %s", r.stdout)
	}
	if !strings.Contains(r.stdout, "2 Failed") {
		t.Errorf("expected '2 Failed' in output, got: %s", r.stdout)
	}
}

func TestPeriodicReportMissingEnvVar(t *testing.T) {
	r := runCLI(t, "prowjob", "periodic-report")
	if r.exitCode == 0 {
		t.Error("expected non-zero exit code without PROW_URL")
	}
}

type statusPageResponse struct {
	Components []statusComponent `json:"components"`
	Incidents  []interface{}     `json:"incidents"`
	Status     statusInfo        `json:"status"`
}

type statusComponent struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type statusInfo struct {
	Indicator   string `json:"indicator"`
	Description string `json:"description"`
}

func TestHealthCheckWithMockServer(t *testing.T) {
	response := statusPageResponse{
		Components: []statusComponent{
			{Name: "API", Status: "operational"},
			{Name: "Registry", Status: "operational"},
			{Name: "Build System", Status: "operational"},
		},
		Incidents: []interface{}{},
		Status: statusInfo{
			Indicator:   "none",
			Description: "All Systems Operational",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	configContent := fmt.Sprintf(`externalServices:
  - name: mock-service
    criticalComponents:
      - API
      - Registry
    statusPageURL: %s/api/v2/summary.json
`, server.URL)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	artifactDir := filepath.Join(tmpDir, "artifacts")
	if err := os.MkdirAll(artifactDir, 0o750); err != nil {
		t.Fatalf("failed to create artifact dir: %v", err)
	}

	r := runCLIWithEnv(t,
		[]string{"ARTIFACT_DIR=" + artifactDir},
		"prowjob", "health-check", "--config="+configPath,
	)
	if r.exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s\nstdout: %s", r.exitCode, r.stderr, r.stdout)
	}
}

func TestHealthCheckDegradedService(t *testing.T) {
	response := statusPageResponse{
		Components: []statusComponent{
			{Name: "API", Status: "major_outage"},
			{Name: "Registry", Status: "operational"},
		},
		Incidents: []interface{}{},
		Status: statusInfo{
			Indicator:   "major",
			Description: "Major System Outage",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	configContent := fmt.Sprintf(`externalServices:
  - name: mock-service
    criticalComponents:
      - API
      - Registry
    statusPageURL: %s/api/v2/summary.json
`, server.URL)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(configPath, []byte(configContent), 0o644)

	artifactDir := filepath.Join(tmpDir, "artifacts")
	_ = os.MkdirAll(artifactDir, 0o750)

	r := runCLIWithEnv(t,
		[]string{"ARTIFACT_DIR=" + artifactDir},
		"prowjob", "health-check", "--config="+configPath, "--fail-if-unhealthy",
	)
	if r.exitCode == 0 {
		t.Error("expected non-zero exit code when critical component has major_outage")
	}
}
