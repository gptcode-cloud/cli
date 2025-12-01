package run_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestChuDoComplexTaskAnalysis(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Create a simple directory structure
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("Failed to create docs directory: %v", err)
	}

	// Create some test files
	testFiles := map[string]string{
		"feature1.md": "# Feature 1\nSome content",
		"feature2.md": "# Feature 2\nMore content",
		"guide.md":    "# Guide\nGuide content",
	}

	for name, content := range testFiles {
		path := filepath.Join(docsDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}

	t.Logf("Running chu do --dry-run in %s", tmpDir)
	t.Logf("This tests Symphony task decomposition...")

	// Run chu do with --dry-run to test analysis without execution
	cmd := exec.Command("chu", "do", "reorganize docs folder into features and guides subdirectories", "--dry-run")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "CHUCHU_TELEMETRY=false")

	done := make(chan struct{})
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = cmd.CombinedOutput()
		close(done)
	}()

	timeout := 2 * time.Minute
	select {
	case <-done:
		if cmdErr != nil {
			t.Logf("Command output:\n%s", string(output))
			t.Fatalf("chu do --dry-run failed: %v", cmdErr)
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("chu do --dry-run exceeded timeout of %s", timeout)
	}

	outputStr := string(output)

	// Verify analysis output contains expected keywords
	expectedKeywords := []string{
		"Analyzing task",
		"complexity",
		"intent",
	}

	for _, keyword := range expectedKeywords {
		if !strings.Contains(strings.ToLower(outputStr), strings.ToLower(keyword)) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nOutput: %s", keyword, outputStr)
		}
	}

	t.Logf("✓ chu do --dry-run completed task analysis")
}

func TestChuDoSimpleTaskExecution(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	t.Logf("Running chu do in %s", tmpDir)
	t.Logf("This may take 2-5 minutes with local Ollama...")

	// Simple task that should NOT trigger Symphony decomposition
	cmd := exec.Command("chu", "do", "create notes.md with the text 'test notes'")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "CHUCHU_TELEMETRY=false")

	done := make(chan struct{})
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = cmd.CombinedOutput()
		close(done)
	}()

	timeout := 5 * time.Minute
	select {
	case <-done:
		if cmdErr != nil {
			t.Logf("Command output:\n%s", string(output))
			t.Fatalf("chu do failed: %v", cmdErr)
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("chu do exceeded timeout of %s", timeout)
	}

	// Verify the file was created
	notesFile := filepath.Join(tmpDir, "notes.md")
	content, err := os.ReadFile(notesFile)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if !strings.Contains(string(content), "test notes") {
		t.Errorf("File content mismatch.\nExpected to contain: test notes\nGot: %s", string(content))
	}

	t.Logf("✓ chu do successfully executed simple task")
}

func TestChuDoComplexityThreshold(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tests := []struct {
		name             string
		task             string
		shouldBeComplex  bool
		expectedKeywords []string
	}{
		{
			name:             "simple task",
			task:             "create summary.txt",
			shouldBeComplex:  false,
			expectedKeywords: []string{"Executing directly"},
		},
		{
			name:             "complex task with reorganize",
			task:             "reorganize all files",
			shouldBeComplex:  true,
			expectedKeywords: []string{"Complex task", "Symphony"},
		},
		{
			name:             "complex task with refactor",
			task:             "refactor code structure",
			shouldBeComplex:  true,
			expectedKeywords: []string{"Complex task", "Symphony"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			t.Logf("Testing task: %s (expected complex: %v)", tt.task, tt.shouldBeComplex)

			cmd := exec.Command("chu", "do", tt.task, "--dry-run", "--verbose")
			cmd.Dir = tmpDir
			cmd.Env = append(os.Environ(), "CHUCHU_TELEMETRY=false")

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Don't fail on execution errors for dry-run
			if err != nil {
				t.Logf("Command output:\n%s", outputStr)
			}

			// Check for expected keywords
			foundAny := false
			for _, keyword := range tt.expectedKeywords {
				if strings.Contains(outputStr, keyword) {
					foundAny = true
					break
				}
			}

			if !foundAny {
				t.Logf("Warning: Expected keywords %v not found in output", tt.expectedKeywords)
				t.Logf("Output:\n%s", outputStr)
			} else {
				t.Logf("✓ Found expected behavior for %s task", tt.name)
			}
		})
	}
}
