package prowjob

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	prowUtils "k8s.io/test-infra/prow/pod-utils/downwardapi"

	"github.com/konflux-ci/qe-tools/pkg/status"
	"github.com/konflux-ci/qe-tools/pkg/types"
)

func TestIsCriticalComponent(t *testing.T) {
	tests := []struct {
		name      string
		service   Service
		component status.Component
		want      bool
	}{
		{
			name:      "component listed as critical",
			service:   Service{Name: "GitHub", CriticalComponents: []string{"API Requests", "Actions"}},
			component: status.Component{Name: "Actions"},
			want:      true,
		},
		{
			name:      "component not listed as critical",
			service:   Service{Name: "GitHub", CriticalComponents: []string{"API Requests"}},
			component: status.Component{Name: "Codespaces"},
			want:      false,
		},
		{
			name:      "service declares no critical components",
			service:   Service{Name: "Quay"},
			component: status.Component{Name: "Registry"},
			want:      false,
		},
		{
			name:      "match is case sensitive",
			service:   Service{Name: "GitHub", CriticalComponents: []string{"Actions"}},
			component: status.Component{Name: "actions"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCriticalComponent(tt.service, tt.component); got != tt.want {
				t.Errorf("isCriticalComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildPRMessage(t *testing.T) {
	services := []Service{
		{Name: "GitHub", StatusPageURL: "https://www.githubstatus.com/api/v2/summary.json"},
		{Name: "Quay", StatusPageURL: "https://status.quay.io/api/v2/summary.json"},
		{Name: "Broken", StatusPageURL: "://not-a-url"},
	}

	tests := []struct {
		name            string
		unhealthy       map[string][]string
		failIfUnhealthy bool
		wantContains    []string
		wantMissing     []string
	}{
		{
			name:            "single service, fail-if-unhealthy set",
			unhealthy:       map[string][]string{"GitHub": {"Actions"}},
			failIfUnhealthy: true,
			wantContains: []string{
				"- GitHub: Actions",
				"E2E tests won't run on your PR",
				"https://www.githubstatus.com",
				"/retest-required",
			},
			wantMissing: []string{"https://status.quay.io", "E2E tests will probably fail"},
		},
		{
			name:            "several components of one service are joined",
			unhealthy:       map[string][]string{"Quay": {"Registry", "Builds"}},
			failIfUnhealthy: false,
			wantContains: []string{
				"- Quay: Registry, Builds",
				"E2E tests will probably fail",
				"https://status.quay.io",
			},
			wantMissing: []string{"https://www.githubstatus.com"},
		},
		{
			name:            "unparseable status page URL is skipped, not fatal",
			unhealthy:       map[string][]string{"Broken": {"Everything"}},
			failIfUnhealthy: false,
			wantContains:    []string{"- Broken: Everything"},
			wantMissing:     []string{"://not-a-url"},
		},
		{
			name:            "no unhealthy components still yields a well formed message",
			unhealthy:       map[string][]string{},
			failIfUnhealthy: false,
			wantContains:    []string{"Detected an outage", "/retest-required"},
			wantMissing:     []string{"https://www.githubstatus.com", "https://status.quay.io"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous := healthCheckConfig
			healthCheckConfig = HealthCheckConfig{ExternalServices: services}
			defer func() { healthCheckConfig = previous }()

			got := buildPRMessage(&HealthCheckStatus{UnhealthyCriticalComponents: tt.unhealthy}, tt.failIfUnhealthy)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildPRMessage() is missing %q, got:\n%s", want, got)
				}
			}
			for _, unwanted := range tt.wantMissing {
				if strings.Contains(got, unwanted) {
					t.Errorf("buildPRMessage() unexpectedly contains %q, got:\n%s", unwanted, got)
				}
			}
		})
	}
}

// writeHealthCheckConfig writes a minimal valid health-check config and returns its path.
func writeHealthCheckConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	body := "externalServices:" +
		"\n  - name: GitHub" +
		"\n    statusPageURL: https://www.githubstatus.com/api/v2/summary.json" +
		"\n    criticalComponents:" +
		"\n      - Actions" +
		"\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return path
}

func TestHealthCheckPreRunE(t *testing.T) {
	notifyEnv := map[string]string{
		types.GithubTokenEnv:    "token",
		prowUtils.RepoOwnerEnv:  "konflux-ci",
		prowUtils.RepoNameEnv:   "qe-tools",
		prowUtils.PullNumberEnv: "1",
	}

	tests := []struct {
		name        string
		notifyOnPR  bool
		omitEnvVar  string
		wantErr     bool
		wantErrPart string
	}{
		{
			name:       "notify-on-pr unset needs no env vars",
			notifyOnPR: false,
		},
		{
			name:       "notify-on-pr set with every env var present",
			notifyOnPR: true,
		},
		{
			name:        "notify-on-pr set without a github token",
			notifyOnPR:  true,
			omitEnvVar:  types.GithubTokenEnv,
			wantErr:     true,
			wantErrPart: strings.ToUpper(types.GithubTokenEnv),
		},
		{
			name:        "notify-on-pr set without a pull number",
			notifyOnPR:  true,
			omitEnvVar:  prowUtils.PullNumberEnv,
			wantErr:     true,
			wantErrPart: strings.ToUpper(prowUtils.PullNumberEnv),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)

			viper.SetConfigFile(writeHealthCheckConfig(t))
			viper.Set(notifyOnPRParamName, tt.notifyOnPR)
			for name, value := range notifyEnv {
				if name == tt.omitEnvVar {
					continue
				}
				viper.Set(name, value)
			}

			err := healthCheckCmd.PreRunE(healthCheckCmd, nil)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("PreRunE() succeeded, want an error about %s", tt.wantErrPart)
				}
				if !strings.Contains(err.Error(), tt.wantErrPart) {
					t.Errorf("PreRunE() error = %q, want it to mention %s", err, tt.wantErrPart)
				}
				return
			}
			if err != nil {
				t.Fatalf("PreRunE() failed: %v", err)
			}
			if len(healthCheckConfig.ExternalServices) != 1 {
				t.Fatalf("PreRunE() parsed %d services, want 1", len(healthCheckConfig.ExternalServices))
			}
			if got := healthCheckConfig.ExternalServices[0].Name; got != "GitHub" {
				t.Errorf("parsed service name = %q, want %q", got, "GitHub")
			}
		})
	}
}
