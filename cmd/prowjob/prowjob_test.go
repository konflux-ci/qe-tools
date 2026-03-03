package prowjob

import (
	"testing"

	"github.com/konflux-ci/qe-tools/pkg/types"
	"github.com/spf13/cobra"
)

func TestProwjobCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "prowjob command has correct use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ProwjobCmd.Use != "prowjob" {
				t.Errorf("expected Use to be 'prowjob', got '%s'", ProwjobCmd.Use)
			}
			if ProwjobCmd.Short == "" {
				t.Error("expected Short description to be set")
			}
		})
	}
}

func TestProwjobSubcommands(t *testing.T) {
	tests := []struct {
		name           string
		subcommandName string
	}{
		{
			name:           "periodic-report subcommand exists",
			subcommandName: "periodic-report",
		},
		{
			name:           "create-report subcommand exists",
			subcommandName: "create-report",
		},
		{
			name:           "health-check subcommand exists",
			subcommandName: "health-check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			for _, cmd := range ProwjobCmd.Commands() {
				if cmd.Use == tt.subcommandName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("subcommand '%s' not found", tt.subcommandName)
			}
		})
	}
}

func TestProwjobCommandFlags(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		flagName    string
		shouldExist bool
	}{
		{
			name:        "create-report has artifact-dir flag",
			cmd:         "create-report",
			flagName:    types.ArtifactDirParamName,
			shouldExist: true,
		},
		{
			name:        "health-check has artifact-dir flag",
			cmd:         "health-check",
			flagName:    types.ArtifactDirParamName,
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetCmd *cobra.Command
			for _, cmd := range ProwjobCmd.Commands() {
				if cmd.Use == tt.cmd {
					targetCmd = cmd
					break
				}
			}

			if targetCmd == nil {
				t.Fatalf("command '%s' not found", tt.cmd)
			}

			flag := targetCmd.Flags().Lookup(tt.flagName)
			if tt.shouldExist && flag == nil {
				t.Errorf("flag '%s' not found in command '%s'", tt.flagName, tt.cmd)
			} else if !tt.shouldExist && flag != nil {
				t.Errorf("flag '%s' should not exist in command '%s'", tt.flagName, tt.cmd)
			}
		})
	}
}
