package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContextInit(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	if err := runContextInit(nil, nil); err != nil {
		t.Fatalf("contextInit failed: %v", err)
	}

	gptcodeDir := filepath.Join(tmpDir, ".gptcode")
	if _, err := os.Stat(gptcodeDir); os.IsNotExist(err) {
		t.Error(".gptcode directory was not created")
	}

	contextDir := filepath.Join(gptcodeDir, "context")
	if _, err := os.Stat(contextDir); os.IsNotExist(err) {
		t.Error(".gptcode/context directory was not created")
	}

	expectedFiles := []string{
		filepath.Join(contextDir, "shared.md"),
		filepath.Join(contextDir, "next.md"),
		filepath.Join(contextDir, "roadmap.md"),
		filepath.Join(gptcodeDir, "config.yml"),
		filepath.Join(gptcodeDir, ".gitignore"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}

	content, err := os.ReadFile(filepath.Join(contextDir, "shared.md"))
	if err != nil {
		t.Fatalf("Failed to read shared.md: %v", err)
	}
	if len(content) == 0 {
		t.Error("shared.md is empty")
	}
}

func TestContextInitAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(".gptcode", 0755); err != nil {
		t.Fatal(err)
	}

	err := runContextInit(nil, nil)
	if err == nil {
		t.Error("Expected error when .gptcode already exists")
	}
}

func TestGetGPTCodeDir(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	gptcodeDir := filepath.Join(tmpDir, ".gptcode")
	if err := os.Mkdir(gptcodeDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	found, err := getGPTCodeDir()
	if err != nil {
		t.Fatalf("getGPTCodeDir failed: %v", err)
	}

	// Resolve symlinks for macOS /var -> /private/var
	expected, _ := filepath.EvalSymlinks(gptcodeDir)
	foundResolved, _ := filepath.EvalSymlinks(found)
	if foundResolved != expected {
		t.Errorf("Expected %s, got %s", expected, foundResolved)
	}
}

func TestGetGPTCodeDirInSubdir(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	gptcodeDir := filepath.Join(tmpDir, ".gptcode")
	if err := os.Mkdir(gptcodeDir, 0755); err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(tmpDir, "sub", "nested")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(subdir); err != nil {
		t.Fatal(err)
	}

	found, err := getGPTCodeDir()
	if err != nil {
		t.Fatalf("getGPTCodeDir failed: %v", err)
	}

	// Resolve symlinks for macOS /var -> /private/var
	expected, _ := filepath.EvalSymlinks(gptcodeDir)
	foundResolved, _ := filepath.EvalSymlinks(found)
	if foundResolved != expected {
		t.Errorf("Expected %s, got %s", expected, foundResolved)
	}
}

func TestGetGPTCodeDirNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	_, err := getGPTCodeDir()
	if err == nil {
		t.Error("Expected error when .gptcode not found")
	}
}

func TestBuildContextContent(t *testing.T) {
	tmpDir := t.TempDir()
	gptcodeDir := filepath.Join(tmpDir, ".gptcode")
	contextDir := filepath.Join(gptcodeDir, "context")

	if err := os.MkdirAll(contextDir, 0755); err != nil {
		t.Fatal(err)
	}

	sharedContent := "# Shared Context\nTest content"
	nextContent := "# Next Tasks\nTODO items"

	if err := os.WriteFile(filepath.Join(contextDir, "shared.md"), []byte(sharedContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contextDir, "next.md"), []byte(nextContent), 0644); err != nil {
		t.Fatal(err)
	}

	content, err := buildContextContent(gptcodeDir, []string{"shared", "next"})
	if err != nil {
		t.Fatalf("buildContextContent failed: %v", err)
	}

	if content == "" {
		t.Error("buildContextContent returned empty string")
	}

	if !contains(content, "# Shared Context") {
		t.Error("Content does not contain shared.md")
	}
	if !contains(content, "# Next Tasks") {
		t.Error("Content does not contain next.md")
	}
}

func TestContextAdd(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	if err := runContextInit(nil, nil); err != nil {
		t.Fatal(err)
	}

	args := []string{"shared", "## New Section\nNew content"}
	if err := runContextAdd(nil, args); err != nil {
		t.Fatalf("contextAdd failed: %v", err)
	}

	gptcodeDir := filepath.Join(tmpDir, ".gptcode")
	sharedPath := filepath.Join(gptcodeDir, "context", "shared.md")
	content, err := os.ReadFile(sharedPath)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(string(content), "New Section") {
		t.Error("Added content not found in shared.md")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
