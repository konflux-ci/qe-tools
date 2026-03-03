package prow

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

func TestDetermineJobTargetFromYAML(t *testing.T) {
	tests := []struct {
		name           string
		prowJob        *v1.ProwJob
		expectedTarget string
		expectedError  string
	}{
		{
			name: "Valid YAML with target argument",
			prowJob: &v1.ProwJob{
				Spec: v1.ProwJobSpec{
					PodSpec: &corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Args: []string{
									"--scenario=execute",
									"--target=appstudio-e2e-tests",
								},
							},
						},
					},
				},
			},
			expectedTarget: "appstudio-e2e-tests",
		},
		{
			name: "Valid YAML with different target",
			prowJob: &v1.ProwJob{
				Spec: v1.ProwJobSpec{
					PodSpec: &corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Args: []string{
									"--target=integration-service-e2e",
									"--other-arg=value",
								},
							},
						},
					},
				},
			},
			expectedTarget: "integration-service-e2e",
		},
		{
			name: "Missing target argument",
			prowJob: &v1.ProwJob{
				Spec: v1.ProwJobSpec{
					PodSpec: &corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Args: []string{
									"--scenario=execute",
									"--other-arg=value",
								},
							},
						},
					},
				},
			},
			expectedError: "expected",
		},
		{
			name: "Invalid target format (no equals sign)",
			prowJob: &v1.ProwJob{
				Spec: v1.ProwJobSpec{
					PodSpec: &corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Args: []string{
									"--target",
									"appstudio-e2e-tests",
								},
							},
						},
					},
				},
			},
			expectedError: "expected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := determineJobTargetFromYAML(tt.prowJob)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if target != tt.expectedTarget {
				t.Errorf("expected target %q, got %q", tt.expectedTarget, target)
			}
		})
	}
}

func TestParseJobSpec(t *testing.T) {
	tests := []struct {
		name          string
		jobSpecData   string
		expectedType  string
		expectedJob   string
		expectedError bool
	}{
		{
			name: "Valid JSON",
			jobSpecData: `{
				"type": "presubmit",
				"job": "pull-ci-konflux-ci-e2e-tests-main-e2e-tests",
				"refs": {
					"org": "konflux-ci",
					"repo": "e2e-tests",
					"repo_link": "https://github.com/konflux-ci/e2e-tests",
					"pulls": [
						{
							"number": 123,
							"author": "test-author",
							"sha": "abc123",
							"link": "https://github.com/konflux-ci/e2e-tests/pull/123",
							"author_link": "https://github.com/test-author"
						}
					]
				}
			}`,
			expectedType: "presubmit",
			expectedJob:  "pull-ci-konflux-ci-e2e-tests-main-e2e-tests",
		},
		{
			name:          "Invalid JSON",
			jobSpecData:   `{"type": "presubmit", invalid json`,
			expectedError: true,
		},
		{
			name:          "Empty string",
			jobSpecData:   "",
			expectedError: true,
		},
		{
			name: "Minimal valid JSON",
			jobSpecData: `{
				"type": "periodic",
				"job": "periodic-job-name",
				"refs": {
					"org": "org-name",
					"repo": "repo-name",
					"repo_link": "https://github.com/org-name/repo-name",
					"pulls": []
				}
			}`,
			expectedType: "periodic",
			expectedJob:  "periodic-job-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobSpec, err := ParseJobSpec(tt.jobSpecData)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if jobSpec.Type != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, jobSpec.Type)
			}

			if jobSpec.Job != tt.expectedJob {
				t.Errorf("expected job %q, got %q", tt.expectedJob, jobSpec.Job)
			}
		})
	}
}

func TestDetermineJobTargetFromProwJobURL(t *testing.T) {
	tests := []struct {
		name           string
		prowJobURL     string
		expectedTarget string
		expectedError  bool
	}{
		{
			name:           "infra-deployments URL",
			prowJobURL:     "https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/redhat-appstudio_infra-deployments/123/pull-ci-redhat-appstudio-infra-deployments-main-appstudio-e2e-tests/123",
			expectedTarget: "appstudio-e2e-tests",
		},
		{
			name:           "e2e-tests URL",
			prowJobURL:     "https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/konflux-ci_e2e-tests/456/pull-ci-konflux-ci-e2e-tests-main-redhat-appstudio-e2e/456",
			expectedTarget: "redhat-appstudio-e2e",
		},
		{
			name:           "integration-service URL",
			prowJobURL:     "https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/konflux-ci_integration-service/789/pull-ci-konflux-ci-integration-service-main-integration-service-e2e/789",
			expectedTarget: "integration-service-e2e",
		},
		{
			name:          "Unknown URL pattern",
			prowJobURL:    "https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/unknown-org_unknown-repo/123/pull-ci-unknown-job/123",
			expectedError: true,
		},
		{
			name:          "Empty URL",
			prowJobURL:    "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := determineJobTargetFromProwJobURL(tt.prowJobURL)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if target != tt.expectedTarget {
				t.Errorf("expected target %q, got %q", tt.expectedTarget, target)
			}
		})
	}
}

func TestGetArtifactsDirectoryPrefix(t *testing.T) {
	tests := []struct {
		name           string
		prowJobURL     string
		jobTarget      string
		expectedPrefix string
		expectedError  bool
	}{
		{
			name:           "Valid Prow URL",
			prowJobURL:     "https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/redhat-appstudio_infra-deployments/123/pull-ci-redhat-appstudio-infra-deployments-main-appstudio-e2e-tests/456",
			jobTarget:      "appstudio-e2e-tests",
			expectedPrefix: "pr-logs/pull/redhat-appstudio_infra-deployments/123/pull-ci-redhat-appstudio-infra-deployments-main-appstudio-e2e-tests/456/artifacts/appstudio-e2e-tests/",
		},
		{
			name:           "Another valid URL",
			prowJobURL:     "https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/konflux-ci_e2e-tests/789/pull-ci-konflux-ci-e2e-tests-main-redhat-appstudio-e2e/101112",
			jobTarget:      "redhat-appstudio-e2e",
			expectedPrefix: "pr-logs/pull/konflux-ci_e2e-tests/789/pull-ci-konflux-ci-e2e-tests-main-redhat-appstudio-e2e/101112/artifacts/redhat-appstudio-e2e/",
		},
		{
			name:          "Invalid URL (no bucket name)",
			prowJobURL:    "https://prow.ci.openshift.org/view/gs/pr-logs/pull/repo/123",
			jobTarget:     "target",
			expectedError: true,
		},
		{
			name:          "Empty URL",
			prowJobURL:    "",
			jobTarget:     "target",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := &ArtifactScanner{}
			prefix, err := getArtifactsDirectoryPrefix(as, tt.prowJobURL, tt.jobTarget)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if prefix != tt.expectedPrefix {
				t.Errorf("expected prefix %q, got %q", tt.expectedPrefix, prefix)
			}

			if as.ArtifactDirectoryPrefix != tt.expectedPrefix {
				t.Errorf("expected ArtifactDirectoryPrefix to be set to %q, got %q", tt.expectedPrefix, as.ArtifactDirectoryPrefix)
			}
		})
	}
}

func TestGetParentStepName(t *testing.T) {
	tests := []struct {
		name                    string
		fullArtifactName        string
		artifactDirectoryPrefix string
		expectedStepName        string
		expectedError           bool
	}{
		{
			name:                    "Valid path with single step",
			fullArtifactName:        "pr-logs/pull/repo/123/job/456/artifacts/target/redhat-appstudio-e2e/artifacts/e2e-report.xml",
			artifactDirectoryPrefix: "pr-logs/pull/repo/123/job/456/artifacts/target/",
			expectedStepName:        "redhat-appstudio-e2e",
		},
		{
			name:                    "Valid path with different step name",
			fullArtifactName:        "pr-logs/pull/repo/456/job/789/artifacts/target/gather-extra/artifacts/build-log.txt",
			artifactDirectoryPrefix: "pr-logs/pull/repo/456/job/789/artifacts/target/",
			expectedStepName:        "gather-extra",
		},
		{
			name:                    "No prefix match",
			fullArtifactName:        "some/other/path/file.xml",
			artifactDirectoryPrefix: "pr-logs/pull/repo/123/artifacts/target/",
			expectedError:           true,
		},
		{
			name:                    "Empty artifact name",
			fullArtifactName:        "",
			artifactDirectoryPrefix: "pr-logs/pull/repo/123/artifacts/target/",
			expectedError:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stepName, err := getParentStepName(tt.fullArtifactName, tt.artifactDirectoryPrefix)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if stepName != tt.expectedStepName {
				t.Errorf("expected step name %q, got %q", tt.expectedStepName, stepName)
			}
		})
	}
}

func TestGetFileName(t *testing.T) {
	tests := []struct {
		name                    string
		fullArtifactName        string
		artifactDirectoryPrefix string
		expectedFileName        string
		expectedError           bool
	}{
		{
			name:                    "Simple filename",
			fullArtifactName:        "pr-logs/pull/repo/123/job/456/artifacts/target/step/artifacts/e2e-report.xml",
			artifactDirectoryPrefix: "pr-logs/pull/repo/123/job/456/artifacts/target/",
			expectedFileName:        "e2e-report.xml",
		},
		{
			name:                    "Nested path with filename",
			fullArtifactName:        "pr-logs/pull/repo/123/job/456/artifacts/target/step/nested/dir/build-log.txt",
			artifactDirectoryPrefix: "pr-logs/pull/repo/123/job/456/artifacts/target/",
			expectedFileName:        "build-log.txt",
		},
		{
			name:                    "Single level after prefix",
			fullArtifactName:        "pr-logs/pull/repo/123/job/456/artifacts/target/file.json",
			artifactDirectoryPrefix: "pr-logs/pull/repo/123/job/456/artifacts/target/",
			expectedFileName:        "file.json",
		},
		{
			name:                    "No prefix match",
			fullArtifactName:        "some/other/path/file.xml",
			artifactDirectoryPrefix: "pr-logs/pull/repo/123/artifacts/target/",
			expectedError:           true,
		},
		{
			name:                    "Empty artifact name",
			fullArtifactName:        "",
			artifactDirectoryPrefix: "pr-logs/pull/repo/123/artifacts/target/",
			expectedError:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileName, err := getFileName(tt.fullArtifactName, tt.artifactDirectoryPrefix)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if fileName != tt.expectedFileName {
				t.Errorf("expected filename %q, got %q", tt.expectedFileName, fileName)
			}
		})
	}
}
