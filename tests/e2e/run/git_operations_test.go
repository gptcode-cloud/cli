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

func TestGitOperations(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Initialize git repo
	setupGitRepo(t, tmpDir)

	t.Run("Show repository status", func(t *testing.T) {
		output := runChuDo(t, tmpDir, "show me the git status", 2*time.Minute)
		if !strings.Contains(output, "On branch master") && !strings.Contains(output, "On branch main") {
			t.Errorf("Expected git status output, got: %s", output)
		}
	})

	t.Run("View commit history", func(t *testing.T) {
		output := runChuDo(t, tmpDir, "show me the recent commits", 2*time.Minute)
		if !strings.Contains(output, "Initial commit") {
			t.Errorf("Expected 'Initial commit' in output, got: %s", output)
		}
	})

	t.Run("Show changes (diff)", func(t *testing.T) {
		// Make a change
		readmePath := filepath.Join(tmpDir, "README.md")
		err := os.WriteFile(readmePath, []byte("# Test Project\nNew content\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to modify README.md: %v", err)
		}

		output := runChuDo(t, tmpDir, "what changes did I make?", 2*time.Minute)
		// Just verify it executes successfully
		if output == "" {
			t.Error("Expected some output for git diff")
		}
	})

	t.Run("List branches", func(t *testing.T) {
		output := runChuDo(t, tmpDir, "show me all git branches", 2*time.Minute)
		if !strings.Contains(output, "master") && !strings.Contains(output, "main") {
			t.Errorf("Expected branch name in output, got: %s", output)
		}
	})

	t.Run("Show untracked files", func(t *testing.T) {
		// Create untracked file
		testFile := filepath.Join(tmpDir, "test.js")
		err := os.WriteFile(testFile, []byte("console.log('test')"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test.js: %v", err)
		}

		output := runChuDo(t, tmpDir, "show me what files are not tracked by git", 2*time.Minute)
		if !strings.Contains(output, "test.js") {
			t.Errorf("Expected 'test.js' in output, got: %s", output)
		}
	})

	t.Logf("✓ All git operations tests passed")
}

func setupGitRepo(t *testing.T, dir string) {
	t.Helper()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@chu.test"},
		{"git", "config", "user.name", "Chu Test"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to run %v: %v", args, err)
		}
	}

	// Create and commit README
	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Project\n"), 0644); err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}

	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}
}

func runChuDo(t *testing.T, dir string, task string, timeout time.Duration) string {
	t.Helper()

	cmd := exec.Command("gptcode", "do", task)
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
			t.Fatalf("chu do failed: %v", cmdErr)
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("chu do exceeded timeout of %s", timeout)
	}

	return string(output)
}
