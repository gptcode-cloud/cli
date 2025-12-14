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

func skipIfNoE2E(t *testing.T) {
	if os.Getenv("E2E_BACKEND") == "" {
		t.Skip("Skipping E2E test: run via 'chu test e2e'")
	}
}

func TestChuDoCreateFile(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	t.Logf("Running chu do in %s", tmpDir)
	t.Logf("This may take 2-5 minutes with local Ollama...")

	cmd := exec.Command("gptcode", "do", "create a file called hello.txt with content 'Hello from GPTCode E2E test'")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GPTCODE_TELEMETRY=false")

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

	helloFile := filepath.Join(tmpDir, "hello.txt")
	content, err := os.ReadFile(helloFile)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expectedContent := "Hello from GPTCode E2E test"
	if !strings.Contains(string(content), expectedContent) {
		t.Errorf("File content mismatch.\nExpected to contain: %s\nGot: %s", expectedContent, string(content))
	}

	t.Logf("✓ chu do successfully created hello.txt")
}

func TestChuDoModifyFile(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("original content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Logf("Running chu do in %s", tmpDir)
	t.Logf("This may take 2-5 minutes with local Ollama...")

	cmd := exec.Command("gptcode", "do", "modify test.txt to say 'modified by E2E test'")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GPTCODE_TELEMETRY=false")

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

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expectedContent := "modified by E2E test"
	if !strings.Contains(string(content), expectedContent) {
		t.Errorf("File content mismatch.\nExpected to contain: %s\nGot: %s", expectedContent, string(content))
	}

	t.Logf("✓ chu do successfully modified test.txt")
}

func TestChuDoTimeout(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	t.Skip("Skipping timeout test - takes too long with local Ollama (5+ minutes expected)")

	tmpDir := t.TempDir()

	cmd := exec.Command("gptcode", "do", "create hello.txt")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GPTCODE_TELEMETRY=false")

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	timeout := 5 * time.Minute
	select {
	case err := <-done:
		if err != nil {
			t.Logf("Command failed (expected for some backends): %v", err)
		} else {
			t.Logf("✓ chu do completed within %s", timeout)
		}
	case <-time.After(timeout):
		if err := cmd.Process.Kill(); err != nil {
			t.Logf("Failed to kill process: %v", err)
		}
		t.Fatalf("chu do exceeded timeout of %s", timeout)
	}
}

func TestChuDoNoUnintendedFiles(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	t.Logf("Running chu do in %s", tmpDir)
	t.Logf("This may take 2-5 minutes with local Ollama...")

	cmd := exec.Command("gptcode", "do", "create result.txt with just the word 'success'")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GPTCODE_TELEMETRY=false")

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

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(entries) != 1 {
		var files []string
		for _, entry := range entries {
			files = append(files, entry.Name())
		}
		t.Errorf("Expected exactly 1 file, got %d: %v", len(entries), files)
	}

	if len(entries) > 0 && entries[0].Name() != "result.txt" {
		t.Errorf("Expected result.txt, got %s", entries[0].Name())
	}

	t.Logf("✓ chu do created only the intended file")
}
