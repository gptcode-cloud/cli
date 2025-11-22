package ollama

import (
	"os"
	"testing"
)

func TestIsInstalled(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration test in CI (requires Ollama service)")
	}
	
	tests := []struct {
		name         string
		modelName    string
		wantErr      bool
		wantInstalled bool
	}{
		{
			name:         "check existing model",
			modelName:    "llama3.1:8b",
			wantErr:      false,
			wantInstalled: true,
		},
		{
			name:         "check nonexistent model",
			modelName:    "nonexistent-model-xyz",
			wantErr:      false,
			wantInstalled: false,
		},
		{
			name:      "empty model name",
			modelName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installed, err := IsInstalled(tt.modelName)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsInstalled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && installed != tt.wantInstalled {
				t.Errorf("IsInstalled() = %v, want %v", installed, tt.wantInstalled)
			}
		})
	}
}

func TestCheckAndInstall(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration test in CI (requires Ollama service)")
	}
	
	tests := []struct {
		name        string
		modelName   string
		autoInstall bool
		wantErr     bool
	}{
		{
			name:        "check installed model without auto-install",
			modelName:   "llama3.1:8b",
			autoInstall: false,
			wantErr:     false,
		},
		{
			name:        "check nonexistent model without auto-install",
			modelName:   "nonexistent-xyz",
			autoInstall: false,
			wantErr:     true,
		},
		{
			name:        "empty model name",
			modelName:   "",
			autoInstall: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAndInstall(tt.modelName, tt.autoInstall, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAndInstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
