package download

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		since   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "parse hours",
			since:   "4h",
			want:    4 * time.Hour,
			wantErr: false,
		},
		{
			name:    "parse minutes",
			since:   "30m",
			want:    30 * time.Minute,
			wantErr: false,
		},
		{
			name:    "parse seconds",
			since:   "60s",
			want:    60 * time.Second,
			wantErr: false,
		},
		{
			name:    "parse days",
			since:   "2d",
			want:    48 * time.Hour,
			wantErr: false,
		},
		{
			name:    "parse single day",
			since:   "1d",
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "invalid format",
			since:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "negative value",
			since:   "-1h",
			want:    -1 * time.Hour,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.since)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloadCommand(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "download command has correct use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if downloadCmd.Use != "download" {
				t.Errorf("expected Use to be 'download', got '%s'", downloadCmd.Use)
			}
			if downloadCmd.Short == "" {
				t.Error("expected Short description to be set")
			}
			if downloadCmd.RunE == nil {
				t.Error("expected RunE to be set")
			}
		})
	}
}

func TestDownloadCommandFlags(t *testing.T) {
	// Initialize the command
	cmd := Init()

	tests := []struct {
		name     string
		flagName string
	}{
		{
			name:     "repo flag exists",
			flagName: "repo",
		},
		{
			name:     "repos flag exists",
			flagName: "repos",
		},
		{
			name:     "since flag exists",
			flagName: "since",
		},
		{
			name:     "oci-cache flag exists",
			flagName: "oci-cache",
		},
		{
			name:     "artifacts-output flag exists",
			flagName: "artifacts-output",
		},
		{
			name:     "no-cache flag exists",
			flagName: "no-cache",
		},
		{
			name:     "uncompress-gz-files flag exists",
			flagName: "uncompress-gz-files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag '%s' not found", tt.flagName)
			}
		})
	}
}
