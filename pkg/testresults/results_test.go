package testresults

import (
	"encoding/xml"
	"testing"

	"github.com/bsm/ginkgo/v2/reporters"
	"github.com/konflux-ci/qe-tools/pkg/oci"
)

func TestCollectTestFilesData(t *testing.T) {
	junitFilename := "junit.xml"
	e2eLogFilename := "e2e-test.log"
	clusterLogFilename := "cluster-provision.log"

	tests := []struct {
		name                string
		filesPathMap        oci.FilesPathMap
		expectedFailureType FailureType
		expectedJUnit       bool
		expectedE2ELog      bool
		expectedClusterLog  bool
	}{
		{
			name: "JUnit file found - TestCaseFailure",
			filesPathMap: oci.FilesPathMap{
				"path/to/junit.xml": {
					Filename: junitFilename,
					Content:  createSampleJUnitXML(1, 0),
				},
			},
			expectedFailureType: TestCaseFailure,
			expectedJUnit:       true,
		},
		{
			name: "No JUnit, E2E log found - TestRunFailure",
			filesPathMap: oci.FilesPathMap{
				"path/to/e2e.log": {
					Filename: e2eLogFilename,
					Content:  "E2E test log content",
				},
			},
			expectedFailureType: TestRunFailure,
			expectedE2ELog:      true,
		},
		{
			name: "No JUnit/E2E, cluster log found - ClusterCreationFailure",
			filesPathMap: oci.FilesPathMap{
				"path/to/cluster.log": {
					Filename: clusterLogFilename,
					Content:  "Cluster provision failed",
				},
			},
			expectedFailureType: ClusterCreationFailure,
			expectedClusterLog:  true,
		},
		{
			name:                "No files found - OtherFailure",
			filesPathMap:        oci.FilesPathMap{},
			expectedFailureType: OtherFailure,
		},
		{
			name: "Both E2E and cluster logs present - TestRunFailure takes priority",
			filesPathMap: oci.FilesPathMap{
				"path/to/e2e.log": {
					Filename: e2eLogFilename,
					Content:  "E2E test log",
				},
				"path/to/cluster.log": {
					Filename: clusterLogFilename,
					Content:  "Cluster log",
				},
			},
			expectedFailureType: TestRunFailure,
			expectedE2ELog:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &FailedTestCasesReport{}
			report.CollectTestFilesData(tt.filesPathMap, junitFilename, e2eLogFilename, clusterLogFilename)

			if report.FailureType != tt.expectedFailureType {
				t.Errorf("expected FailureType %q, got %q", tt.expectedFailureType, report.FailureType)
			}

			if tt.expectedJUnit && report.JUnitTestSuites == nil {
				t.Error("expected JUnitTestSuites to be populated")
			}

			if tt.expectedE2ELog && report.E2ETestLog == "" {
				t.Error("expected E2ETestLog to be populated")
			}

			if tt.expectedClusterLog && report.ClusterProvisionLog == "" {
				t.Error("expected ClusterProvisionLog to be populated")
			}
		})
	}
}

func TestGetFailedTestCases(t *testing.T) {
	tests := []struct {
		name          string
		junitXML      string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "Multiple failures in test suite",
			junitXML:      createSampleJUnitXML(2, 1),
			expectedCount: 2,
			expectedNames: []string{"Test Case 1", "Test Case 2"},
		},
		{
			name:          "Mixed success and failure",
			junitXML:      createSampleJUnitXML(1, 0),
			expectedCount: 1,
			expectedNames: []string{"Test Case 1"},
		},
		{
			name:          "All passing tests",
			junitXML:      createSampleJUnitXML(0, 1),
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name:          "Test with error instead of failure",
			junitXML:      createJUnitWithError(),
			expectedCount: 1,
			expectedNames: []string{"Test Case With Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var junitSuites reporters.JUnitTestSuites
			if err := xml.Unmarshal([]byte(tt.junitXML), &junitSuites); err != nil {
				t.Fatalf("failed to unmarshal test JUnit XML: %v", err)
			}

			report := &FailedTestCasesReport{
				JUnitTestSuites: &junitSuites,
			}

			failedCases := report.GetFailedTestCases()

			if len(failedCases) != tt.expectedCount {
				t.Errorf("expected %d failed test cases, got %d", tt.expectedCount, len(failedCases))
			}

			if tt.expectedCount > 0 {
				for i, expectedName := range tt.expectedNames {
					if i >= len(failedCases) {
						break
					}
					if failedCases[i].Name != expectedName {
						t.Errorf("expected test case name %q, got %q", expectedName, failedCases[i].Name)
					}
				}
			}
		})
	}
}

// createSampleJUnitXML creates a sample JUnit XML with specified number of failures and successes per suite
func createSampleJUnitXML(failuresPerSuite, successesPerSuite int) string {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>`

	for suite := 0; suite < 1; suite++ {
		xml += `
    <testsuite name="Test Suite" tests="` + string(rune(failuresPerSuite+successesPerSuite+'0')) + `" failures="` + string(rune(failuresPerSuite+'0')) + `" errors="0">`

		for i := 1; i <= failuresPerSuite; i++ {
			xml += `
        <testcase name="Test Case ` + string(rune(i+'0')) + `" classname="TestClass" time="1.23">
            <failure message="Test failed" type="AssertionError">Expected true but got false</failure>
        </testcase>`
		}

		for i := 1; i <= successesPerSuite; i++ {
			xml += `
        <testcase name="Passing Test ` + string(rune(i+'0')) + `" classname="TestClass" time="0.5">
        </testcase>`
		}

		xml += `
    </testsuite>`
	}

	xml += `
</testsuites>`
	return xml
}

// createJUnitWithError creates a JUnit XML with an error instead of a failure
func createJUnitWithError() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
    <testsuite name="Error Suite" tests="1" failures="0" errors="1">
        <testcase name="Test Case With Error" classname="TestClass" time="1.0">
            <error message="Runtime error" type="RuntimeError">Unexpected error occurred</error>
        </testcase>
    </testsuite>
</testsuites>`
}
