//go:build e2e

package run_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestTDDFeatureDevelopment(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Create Go project structure
	setupGoProject(t, tmpDir)

	t.Run("Generate tests for calculator", func(t *testing.T) {
		output := runChuTDD(t, tmpDir, "Create a Calculator struct with Add, Subtract, Multiply, Divide methods for integers", 5*time.Minute)

		// Just verify it runs successfully
		if len(output) == 0 {
			t.Error("Expected some output for calculator TDD")
		}

		t.Logf("✓ Generated calculator tests")
	})

	t.Run("Generate tests for string utility", func(t *testing.T) {
		output := runChuTDD(t, tmpDir, "Write tests for a string utility with Reverse and ToUpperCase functions", 5*time.Minute)

		if len(output) == 0 {
			t.Error("Expected some output for string utility TDD")
		}

		t.Logf("✓ Generated string utility tests")
	})

	t.Run("Generate tests for validator", func(t *testing.T) {
		output := runChuTDD(t, tmpDir, "Create a Validator with email and phone validation", 5*time.Minute)

		if len(output) == 0 {
			t.Error("Expected some output for validator TDD")
		}

		t.Logf("✓ Generated validator tests")
	})

	t.Run("Generate tests for cache", func(t *testing.T) {
		output := runChuTDD(t, tmpDir, "Build a Cache with Get, Set, and Delete operations", 5*time.Minute)

		if len(output) == 0 {
			t.Error("Expected some output for cache TDD")
		}

		t.Logf("✓ Generated cache tests")
	})

	t.Logf("✓ All TDD development tests passed")
}

func setupGoProject(t *testing.T, tmpDir string) {
	t.Helper()

	goModContent := `module myapp

go 1.22

require github.com/stretchr/testify v1.8.4
`

	modPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create math_utils directory
	mathUtilsDir := filepath.Join(tmpDir, "math_utils")
	if err := os.MkdirAll(mathUtilsDir, 0755); err != nil {
		t.Fatalf("Failed to create math_utils dir: %v", err)
	}
}

func runChuTDD(t *testing.T, dir string, task string, timeout time.Duration) string {
	t.Helper()

	cmd := exec.Command("gptcode", "tdd", task)
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
			t.Fatalf("chu tdd failed: %v", cmdErr)
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("chu tdd exceeded timeout of %s", timeout)
	}

	return string(output)
}
