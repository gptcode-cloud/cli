//go:build e2e

package planning

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestResearchHelp tests that research command exists and shows help
func TestResearchHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	cmd := exec.Command("gptcode", "research", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chu research --help failed: %v", err)
	}

	outputStr := string(output)

	// Should mention research
	if !strings.Contains(outputStr, "research") && !strings.Contains(outputStr, "Research") {
		t.Errorf("Expected help to mention research, got: %s", outputStr)
	}
}

// TestPlanHelp tests that plan command exists and shows help
func TestPlanHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	cmd := exec.Command("gptcode", "plan", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chu plan --help failed: %v", err)
	}

	outputStr := string(output)

	// Should mention plan
	if !strings.Contains(outputStr, "plan") && !strings.Contains(outputStr, "Plan") {
		t.Errorf("Expected help to mention plan, got: %s", outputStr)
	}
}

// TestTDDHelp tests that TDD command exists and shows help
func TestTDDHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	cmd := exec.Command("gptcode", "tdd", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chu tdd --help failed: %v", err)
	}

	outputStr := string(output)

	// Should mention TDD
	if !strings.Contains(outputStr, "tdd") && !strings.Contains(outputStr, "TDD") {
		t.Errorf("Expected help to mention TDD, got: %s", outputStr)
	}
}

// TestDoHelp tests that do command exists and shows help
func TestDoHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	cmd := exec.Command("gptcode", "do", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chu do --help failed: %v", err)
	}

	outputStr := string(output)

	// Should mention autonomous or task execution
	if !strings.Contains(outputStr, "do") && !strings.Contains(outputStr, "Do") {
		t.Errorf("Expected help to mention do command, got: %s", outputStr)
	}
}

// TestResearchBasic tests basic research command (if implemented)
func TestResearchBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Skip("Research command not yet fully implemented - placeholder test")

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Create a simple codebase
	err := os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Try to research the codebase
	cmd := exec.Command("gptcode", "research", "How does this project work?")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Research command failed (expected if not implemented): %v", err)
		t.Logf("Output: %s", output)
	}
}

// TestPlanGeneration tests basic plan generation (if implemented)
func TestPlanGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Skip("Plan command not yet fully implemented - placeholder test")

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Create a simple project structure
	err := os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Try to create a plan
	cmd := exec.Command("gptcode", "plan", "Add error handling")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Plan command failed (expected if not implemented): %v", err)
		t.Logf("Output: %s", output)
	}
}

// TestTDDWorkflow tests TDD workflow (if implemented)
func TestTDDWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Skip("TDD workflow not yet fully implemented - placeholder test")

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Initialize a Go module
	cmd := exec.Command("go", "mod", "init", "test")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Try TDD mode
	cmd = exec.Command("gptcode", "tdd", "create a sum function")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("TDD command failed (expected if not implemented): %v", err)
		t.Logf("Output: %s", output)
	}
}

// TestCommandsExist validates all planning-related commands are registered
func TestCommandsExist(t *testing.T) {
	commands := []string{"research", "plan", "tdd", "do"}

	for _, cmdName := range commands {
		t.Run(cmdName, func(t *testing.T) {
			cmd := exec.Command("gptcode", cmdName, "--help")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Command '%s' failed: %v\nOutput: %s", cmdName, err, output)
			}
		})
	}
}
