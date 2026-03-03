package webhook

import (
	"testing"
)

func TestWebhookCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "webhook command has correct use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if WebhookCmd.Use != "webhook" {
				t.Errorf("expected Use to be 'webhook', got '%s'", WebhookCmd.Use)
			}
			if WebhookCmd.Short == "" {
				t.Error("expected Short description to be set")
			}
		})
	}
}

func TestWebhookSubcommands(t *testing.T) {
	tests := []struct {
		name           string
		subcommandName string
	}{
		{
			name:           "report-portal subcommand exists",
			subcommandName: "report-portal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			for _, cmd := range WebhookCmd.Commands() {
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
