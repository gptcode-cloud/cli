//go:build e2e

package run_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBasicFileOperations(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	t.Run("Create text file with content", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "create a file named hello.txt with the text 'Hello World'", 2*time.Minute)

		helloFile := filepath.Join(tmpDir, "hello.txt")
		content, err := os.ReadFile(helloFile)
		if err != nil {
			t.Fatalf("Failed to read hello.txt: %v", err)
		}

		if !strings.Contains(string(content), "Hello") {
			t.Errorf("Expected 'Hello' in hello.txt, got: %s", string(content))
		}

		t.Logf("✓ Created hello.txt with content")
	})

	t.Run("Read and display file contents", func(t *testing.T) {
		output := runChuDo(t, tmpDir, "show me what's in hello.txt", 2*time.Minute)

		if !strings.Contains(output, "Hello") {
			t.Errorf("Expected 'Hello' in output, got: %s", output)
		}

		t.Logf("✓ Read file contents")
	})

	t.Run("Append to existing file", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "add a new line saying 'Goodbye' to hello.txt", 2*time.Minute)

		helloFile := filepath.Join(tmpDir, "hello.txt")
		content, err := os.ReadFile(helloFile)
		if err != nil {
			t.Fatalf("Failed to read hello.txt: %v", err)
		}

		if !strings.Contains(string(content), "Goodbye") {
			t.Errorf("Expected 'Goodbye' in hello.txt, got: %s", string(content))
		}

		t.Logf("✓ Appended to file")
	})

	t.Run("Create JSON file with structure", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "create config.json with name=myapp and version=1.0", 2*time.Minute)

		configFile := filepath.Join(tmpDir, "config.json")
		content, err := os.ReadFile(configFile)
		if err != nil {
			t.Fatalf("Failed to read config.json: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "myapp") {
			t.Errorf("Expected 'myapp' in config.json, got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "1.0") {
			t.Errorf("Expected '1.0' in config.json, got: %s", contentStr)
		}

		t.Logf("✓ Created JSON file")
	})

	t.Run("Create YAML configuration", func(t *testing.T) {
		_ = runChuDo(t, tmpDir, "create settings.yaml with database host localhost and port 5432", 2*time.Minute)

		yamlFile := filepath.Join(tmpDir, "settings.yaml")
		content, err := os.ReadFile(yamlFile)
		if err != nil {
			t.Fatalf("Failed to read settings.yaml: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "localhost") {
			t.Errorf("Expected 'localhost' in settings.yaml, got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "5432") {
			t.Errorf("Expected '5432' in settings.yaml, got: %s", contentStr)
		}

		t.Logf("✓ Created YAML file")
	})

	t.Run("List all files in directory", func(t *testing.T) {
		output := runChuDo(t, tmpDir, "list all files in the current directory", 2*time.Minute)

		// Check that all created files are mentioned
		expectedFiles := []string{"hello.txt", "config.json", "settings.yaml"}
		for _, file := range expectedFiles {
			if !strings.Contains(output, file) {
				t.Errorf("Expected '%s' in output, got: %s", file, output)
			}
		}

		t.Logf("✓ Listed all files")
	})

	t.Logf("✓ All basic file operations tests passed")
}
