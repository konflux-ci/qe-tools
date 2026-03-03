package analyzetestresults

import (
	"testing"

	"github.com/konflux-ci/qe-tools/pkg/types"
)

func TestAnalyzeTestResultsCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "analyze-test-results command has correct use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if AnalyzeTestResultsCmd.Use != "analyze-test-results" {
				t.Errorf("expected Use to be 'analyze-test-results', got '%s'", AnalyzeTestResultsCmd.Use)
			}
			if AnalyzeTestResultsCmd.Short == "" {
				t.Error("expected Short description to be set")
			}
			if AnalyzeTestResultsCmd.PreRunE == nil {
				t.Error("expected PreRunE to be set")
			}
			if AnalyzeTestResultsCmd.RunE == nil {
				t.Error("expected RunE to be set")
			}
		})
	}
}

func TestAnalyzeTestResultsFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{
			name:     "oci-artifact-ref flag exists",
			flagName: types.OciArtifactRefParamName,
		},
		{
			name:     "junit-filename flag exists",
			flagName: types.JUnitFilenameParamName,
		},
		{
			name:     "cluster-provision-log-filename flag exists",
			flagName: types.ClusterProvisionLogFileParamName,
		},
		{
			name:     "e2e-test-run-log-filename flag exists",
			flagName: types.E2ETestRunLogFileParamName,
		},
		{
			name:     "output-filename flag exists",
			flagName: types.OutputFilenameParamName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := AnalyzeTestResultsCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag '%s' not found", tt.flagName)
			}
		})
	}
}

func TestAnalyzeTestResultsPreRunE(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
	}{
		{
			name:    "missing oci-artifact-ref",
			env:     map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for the test
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			err := AnalyzeTestResultsCmd.PreRunE(AnalyzeTestResultsCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("PreRunE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
