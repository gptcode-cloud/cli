package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigGetSet(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}
	gptcodeDir := filepath.Join(tmpDir, ".gptcode")
	if err := os.MkdirAll(gptcodeDir, 0755); err != nil {
		t.Fatalf("Failed to create .gptcode dir: %v", err)
	}

	setupYAML := `defaults:
    backend: groq
    profile: default
    model: llama-3.3-70b-versatile
    lang: go
backend:
    groq:
        type: openai
        base_url: https://api.groq.com/openai/v1
        default_model: llama-3.3-70b-versatile
        models:
            llama70b: llama-3.3-70b-versatile
`
	setupPath := filepath.Join(gptcodeDir, "setup.yaml")
	if err := os.WriteFile(setupPath, []byte(setupYAML), 0644); err != nil {
		t.Fatalf("Failed to write test setup.yaml: %v", err)
	}

	tests := []struct {
		name      string
		key       string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "get defaults.backend",
			key:       "defaults.backend",
			wantValue: "groq",
			wantErr:   false,
		},
		{
			name:      "get defaults.profile",
			key:       "defaults.profile",
			wantValue: "default",
			wantErr:   false,
		},
		{
			name:      "get backend.groq.type",
			key:       "backend.groq.type",
			wantValue: "openai",
			wantErr:   false,
		},
		{
			name:      "get invalid key",
			key:       "invalid.key",
			wantValue: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfig(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantValue {
				t.Errorf("GetConfig() = %v, want %v", got, tt.wantValue)
			}
		})
	}

	t.Run("set and get", func(t *testing.T) {
		if err := SetConfig("defaults.profile", "speed"); err != nil {
			t.Fatalf("SetConfig() error = %v", err)
		}

		got, err := GetConfig("defaults.profile")
		if err != nil {
			t.Fatalf("GetConfig() error = %v", err)
		}

		if got != "speed" {
			t.Errorf("After SetConfig, got %v, want speed", got)
		}
	})
}
