package coffeebreak

import (
	"testing"
)

func TestCoffeeBreakCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "coffee-break command has correct use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if CoffeeBreakCmd.Use != "coffee-break" {
				t.Errorf("expected Use to be 'coffee-break', got '%s'", CoffeeBreakCmd.Use)
			}
			if CoffeeBreakCmd.Short == "" {
				t.Error("expected Short description to be set")
			}
			if CoffeeBreakCmd.PreRunE == nil {
				t.Error("expected PreRunE to be set")
			}
			if CoffeeBreakCmd.Run == nil {
				t.Error("expected Run to be set")
			}
		})
	}
}

func TestCoffeeBreakPreRunE(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
	}{
		{
			name:    "missing slack_token",
			env:     map[string]string{"hacbs_channel_id": "test"},
			wantErr: true,
		},
		{
			name:    "missing hacbs_channel_id",
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

			err := CoffeeBreakCmd.PreRunE(CoffeeBreakCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("PreRunE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
