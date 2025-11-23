package main

import (
	"os"

	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chuchu_lang_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(tmpDir)

	tests := []struct {
		filename string
		expected string
	}{
		{"go.mod", "go"},
		{"package.json", "typescript"},
		{"mix.exs", "elixir"},
		{"Gemfile", "ruby"},
		{"requirements.txt", "python"},
		{"Cargo.toml", "rust"},
		{"unknown.txt", "unknown"},
	}

	for _, tt := range tests {
		if tt.filename != "unknown.txt" {
			os.WriteFile(tt.filename, []byte(""), 0644)
		}
		
		got := detectLanguage()
		if got != tt.expected {
			t.Errorf("detectLanguage() with %s = %s, want %s", tt.filename, got, tt.expected)
		}
		
		if tt.filename != "unknown.txt" {
			os.Remove(tt.filename)
		}
	}
}
