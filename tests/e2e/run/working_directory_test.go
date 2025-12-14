//go:build e2e

package run_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorkingDirectoryManagement(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Create project structure
	setupProjectDirs(t, tmpDir)

	t.Run("Navigate to frontend directory", func(t *testing.T) {
		input := "/cd frontend\npwd\n/exit\n"
		output := runChuRunWithInput(t, tmpDir, input, 1*time.Minute)

		if !strings.Contains(output, "frontend") {
			t.Errorf("Expected 'frontend' in output, got: %s", output)
		}

		t.Logf("✓ Changed directory to frontend")
	})

	t.Run("Set environment variable", func(t *testing.T) {
		input := "/env NODE_ENV=development\n/env NODE_ENV\n/exit\n"
		output := runChuRunWithInput(t, tmpDir, input, 1*time.Minute)

		if !strings.Contains(output, "NODE_ENV=development") {
			t.Errorf("Expected 'NODE_ENV=development' in output, got: %s", output)
		}

		t.Logf("✓ Set environment variable")
	})

	t.Run("List all environment variables", func(t *testing.T) {
		input := "/env API_KEY=secret123\n/env DB_URL=postgres://localhost\n/env\n/exit\n"
		output := runChuRunWithInput(t, tmpDir, input, 1*time.Minute)

		if !strings.Contains(output, "API_KEY=secret123") {
			t.Errorf("Expected 'API_KEY=secret123' in output, got: %s", output)
		}
		if !strings.Contains(output, "DB_URL=postgres://localhost") {
			t.Errorf("Expected 'DB_URL=postgres://localhost' in output, got: %s", output)
		}

		t.Logf("✓ Listed environment variables")
	})

	t.Run("Change directory and verify pwd", func(t *testing.T) {
		input := "/cd backend\npwd\nls\n/exit\n"
		output := runChuRunWithInput(t, tmpDir, input, 1*time.Minute)

		if !strings.Contains(output, "backend") {
			t.Errorf("Expected 'backend' in output, got: %s", output)
		}
		if !strings.Contains(output, "server.go") {
			t.Errorf("Expected 'server.go' in output, got: %s", output)
		}

		t.Logf("✓ Changed to backend and verified files")
	})

	t.Logf("✓ All working directory tests passed")
}

func setupProjectDirs(t *testing.T, tmpDir string) {
	t.Helper()

	// Create directories
	dirs := []string{"frontend", "backend", "config"}
	for _, dir := range dirs {
		path := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("Failed to create %s: %v", dir, err)
		}
	}

	// Create files
	files := map[string]string{
		"frontend/index.html": "<html><body>Frontend</body></html>",
		"backend/server.go":   "package main",
		"config/dev.env":      "API_URL=http://localhost:3000",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", path, err)
		}
	}
}

func runChuRunWithInput(t *testing.T, dir string, input string, timeout time.Duration) string {
	t.Helper()

	cmd := exec.Command("gptcode", "run", "--raw")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GPTCODE_TELEMETRY=false")
	cmd.Stdin = strings.NewReader(input)

	done := make(chan struct{})
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = cmd.CombinedOutput()
		close(done)
	}()

	select {
	case <-done:
		if cmdErr != nil {
			t.Logf("Command output:\n%s", string(output))
			t.Fatalf("chu run failed: %v", cmdErr)
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("chu run exceeded timeout of %s", timeout)
	}

	return string(output)
}
