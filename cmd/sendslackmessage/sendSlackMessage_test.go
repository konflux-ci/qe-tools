package sendslackmessage

import (
	"testing"
)

func TestSendSlackMessageCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "send-slack-message command has correct use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if SendSlackMessageCmd.Use != "send-slack-message" {
				t.Errorf("expected Use to be 'send-slack-message', got '%s'", SendSlackMessageCmd.Use)
			}
			if SendSlackMessageCmd.Short == "" {
				t.Error("expected Short description to be set")
			}
			if SendSlackMessageCmd.PreRunE == nil {
				t.Error("expected PreRunE to be set")
			}
			if SendSlackMessageCmd.Run == nil {
				t.Error("expected Run to be set")
			}
		})
	}
}

func TestSendSlackMessagePreRunE(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
	}{
		{
			name:    "missing slack_token",
			env:     map[string]string{"channel_id": "test"},
			wantErr: true,
		},
		{
			name:    "missing channel_id",
			env:     map[string]string{"slack_token": "test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for the test
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			err := SendSlackMessageCmd.PreRunE(SendSlackMessageCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("PreRunE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendSlackMessageFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		required bool
	}{
		{
			name:     "message flag exists",
			flagName: "message",
			required: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := SendSlackMessageCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag '%s' not found", tt.flagName)
			}
		})
	}
}
