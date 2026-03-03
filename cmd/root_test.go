package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "root command has correct use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rootCmd.Use != "qe-tools" {
				t.Errorf("expected Use to be 'qe-tools', got '%s'", rootCmd.Use)
			}
			if rootCmd.Short == "" {
				t.Error("expected Short description to be set")
			}
		})
	}
}

func TestInitConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "initConfig does not panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("initConfig panicked: %v", r)
				}
			}()
			initConfig()
		})
	}
}

func TestRootCommandFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{
			name:     "config flag exists",
			flagName: "config",
		},
		{
			name:     "toggle flag exists",
			flagName: "toggle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := rootCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				persistentFlag := rootCmd.PersistentFlags().Lookup(tt.flagName)
				if persistentFlag == nil {
					t.Errorf("flag '%s' not found", tt.flagName)
				}
			}
		})
	}
}
