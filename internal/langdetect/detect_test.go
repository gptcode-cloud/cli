package langdetect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		wantLang  Language
	}{
		{
			name: "elixir project",
			files: map[string]string{
				"mix.exs": "defmodule MyApp.MixProject do\nend",
			},
			wantLang: Elixir,
		},
		{
			name: "ruby project",
			files: map[string]string{
				"Gemfile": "source 'https://rubygems.org'",
			},
			wantLang: Ruby,
		},
		{
			name: "go project",
			files: map[string]string{
				"go.mod": "module example.com/myapp",
			},
			wantLang: Go,
		},
		{
			name: "typescript project",
			files: map[string]string{
				"tsconfig.json": "{}",
			},
			wantLang: TypeScript,
		},
		{
			name: "python project",
			files: map[string]string{
				"requirements.txt": "flask==2.0.0",
			},
			wantLang: Python,
		},
		{
			name: "go by file extensions",
			files: map[string]string{
				"main.go":    "package main",
				"handler.go": "package main",
			},
			wantLang: Go,
		},
		{
			name: "elixir by file extensions",
			files: map[string]string{
				"lib/app.ex":  "defmodule App do\nend",
				"test/app_test.exs": "defmodule AppTest do\nend",
			},
			wantLang: Elixir,
		},
		{
			name:     "unknown project",
			files:    map[string]string{},
			wantLang: Unknown,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create dir %s: %v", dir, err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", fullPath, err)
				}
			}
			
			got := DetectLanguage(tmpDir)
			if got != tt.wantLang {
				t.Errorf("DetectLanguage() = %v, want %v", got, tt.wantLang)
			}
		})
	}
}

func TestDetectFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     Language
	}{
		{"main.go", Go},
		{"app.ex", Elixir},
		{"test.exs", Elixir},
		{"controller.rb", Ruby},
		{"component.tsx", TypeScript},
		{"script.js", TypeScript},
		{"app.py", Python},
		{"readme.md", Unknown},
		{"Makefile", Unknown},
	}
	
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetectFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("DetectFromFilename(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
