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

func TestSingleShotAutomation(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	t.Run("Execute command and exit immediately", func(t *testing.T) {
		output := runChuRun(t, tmpDir, "echo 'Build started'", "--raw", 1*time.Minute)

		if !strings.Contains(output, "Build started") {
			t.Errorf("Expected 'Build started' in output, got: %s", output)
		}

		t.Logf("✓ Single-shot execution works")
	})

	t.Run("List files in single-shot mode", func(t *testing.T) {
		// Create test files
		readmePath := filepath.Join(tmpDir, "README.md")
		if err := os.WriteFile(readmePath, []byte("# Project\n"), 0644); err != nil {
			t.Fatalf("Failed to create README.md: %v", err)
		}

		srcDir := filepath.Join(tmpDir, "src")
		if err := os.MkdirAll(srcDir, 0755); err != nil {
			t.Fatalf("Failed to create src dir: %v", err)
		}

		mainPath := filepath.Join(srcDir, "main.go")
		if err := os.WriteFile(mainPath, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("Failed to create main.go: %v", err)
		}

		output := runChuRun(t, tmpDir, "ls -R", "--raw", 1*time.Minute)

		if !strings.Contains(output, "README.md") {
			t.Errorf("Expected 'README.md' in output, got: %s", output)
		}
		if !strings.Contains(output, "main.go") {
			t.Errorf("Expected 'main.go' in output, got: %s", output)
		}

		t.Logf("✓ Listed files recursively")
	})

	t.Run("Command with arguments", func(t *testing.T) {
		output := runChuRun(t, tmpDir, "cat README.md", "--raw", 1*time.Minute)

		if !strings.Contains(output, "# Project") {
			t.Errorf("Expected '# Project' in output, got: %s", output)
		}

		t.Logf("✓ Command with arguments works")
	})

	t.Run("Verify no REPL banner in single-shot", func(t *testing.T) {
		output := runChuRun(t, tmpDir, "pwd", "--raw", 1*time.Minute)

		if strings.Contains(output, "GPTCode Run REPL") {
			t.Errorf("Expected no REPL banner in single-shot mode, got: %s", output)
		}

		t.Logf("✓ No REPL banner shown")
	})

	t.Logf("✓ All single-shot automation tests passed")
}

func runChuRun(t *testing.T, dir string, command string, flag string, timeout time.Duration) string {
	t.Helper()

	cmd := exec.Command("gptcode", "run", command, flag)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GPTCODE_TELEMETRY=false")

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
