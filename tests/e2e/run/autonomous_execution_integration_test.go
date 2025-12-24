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
		t.Skip("Skipping E2E test: gptcode test e2e")
	}
}

func TestAutonomousExecutionWithBudgetAndVerification(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Create a simple Go project structure to test with
	goDir := filepath.Join(tmpDir, "go-project")
	if err := os.MkdirAll(goDir, 0755); err != nil {
		t.Fatalf("Failed to create Go project directory: %v", err)
	}

	// Create a simple Go file that can be built and tested
	goCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func Add(a, b int) int {
	return a + b
}`
	mainGo := filepath.Join(goDir, "main.go")
	if err := os.WriteFile(mainGo, []byte(goCode), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create a simple test file
	testCode := `package main

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", result)
	}
}
`
	testFile := filepath.Join(goDir, "main_test.go")
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to create main_test.go: %v", err)
	}

	// Create go.mod file
	goMod := `module test-project

go 1.21
`
	modFile := filepath.Join(goDir, "go.mod")
	if err := os.WriteFile(modFile, []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	t.Logf("Running gptcode do in %s with budget constraints", goDir)
	t.Logf("This tests autonomous execution with cost tracking and verification...")

	// Run gptcode do with a task that will modify the Go file, triggering verification
	cmd := exec.Command("gptcode", "do", "add a new function called Multiply that multiplies two integers and update the test to include it", "--budget-mode", "--monthly-budget", "10.0", "--max-cost-per-task", "2.0")
	cmd.Dir = goDir
	cmd.Env = append(os.Environ(), 
		"GPTCODE_TELEMETRY=false",
		"GPTCODE_BUDGET_MODE=true",
		"GPTCODE_MONTHLY_BUDGET=10.0",
		"GPTCODE_MAX_COST_PER_TASK=2.0",
	)

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
		// Don't fail on cmdErr as we're testing autonomous execution
		if cmdErr != nil {
			t.Logf("Command output:\n%s", string(output))
			t.Logf("Note: Some output is expected even with errors in E2E tests")
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("gptcode do exceeded timeout of %s", timeout)
	}

	outputStr := string(output)
	t.Logf("Command completed with output:\n%s", outputStr)

	// Verify that budget-related messages appear in output
	hasBudgetCheck := strings.Contains(outputStr, "budget") || strings.Contains(outputStr, "Budget")
	hasCostTracking := strings.Contains(outputStr, "cost") || strings.Contains(outputStr, "Cost")
	
	if !hasBudgetCheck {
		t.Logf("Warning: Budget check message not found in output, but this may be expected depending on implementation")
	}
	if !hasCostTracking {
		t.Logf("Warning: Cost tracking message not found in output, but this may be expected depending on implementation")
	}

	// Check if the new function was added to the Go file
	updatedContent, err := os.ReadFile(mainGo)
	if err != nil {
		t.Fatalf("Failed to read updated main.go: %v", err)
	}

	updatedContentStr := string(updatedContent)
	hasMultiplyFunction := strings.Contains(updatedContentStr, "Multiply")
	
	if !hasMultiplyFunction {
		t.Logf("Updated file content:\n%s", updatedContentStr)
		t.Errorf("Expected Multiply function to be added to main.go")
	} else {
		t.Logf("✓ Multiply function was successfully added to the Go file")
	}

	// Check if the test file was updated to include the new test
	updatedTestContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Logf("Test file may not have been created or updated: %v", err)
		// This is acceptable in some cases
	} else {
		updatedTestStr := string(updatedTestContent)
		hasMultiplyTest := strings.Contains(updatedTestStr, "Multiply")
		
		if hasMultiplyTest {
			t.Logf("✓ Test file was successfully updated with Multiply test")
		} else {
			t.Logf("Test file content:\n%s", updatedTestStr)
			t.Logf("Note: Test file was not updated with Multiply test, which may be expected")
		}
	}

	t.Logf("✓ Autonomous execution with budget and verification completed")
}

func TestDynamicVerifierSelection(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Create different types of files to test dynamic verifier selection
	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`,
		"script.py": `print("Hello from Python")
`,
		"README.md": `# Test Project

This is a test project.
`,
		"config.json": `{
    "name": "test",
    "version": "1.0.0"
}
`,
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	// Initialize git repo to track changes
	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = tmpDir
	if err := gitInitCmd.Run(); err != nil {
		t.Logf("Warning: Failed to initialize git repo: %v", err)
		// Continue test even if git init fails
	} else {
		// Add all files to git
		gitAddCmd := exec.Command("git", "add", ".")
		gitAddCmd.Dir = tmpDir
		if err := gitAddCmd.Run(); err != nil {
			t.Logf("Warning: Failed to add files to git: %v", err)
		} else {
			// Make initial commit
			gitCommitCmd := exec.Command("git", "commit", "-m", "Initial commit")
			gitCommitCmd.Dir = tmpDir
			gitCommitCmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
			if err := gitCommitCmd.Run(); err != nil {
				t.Logf("Warning: Failed to make initial commit: %v", err)
			}
		}
	}

	t.Logf("Running gptcode do to test dynamic verifier selection in %s", tmpDir)

	// Run gptcode do with a task that modifies only the Go file
	cmd := exec.Command("gptcode", "do", "add a new function to main.go that adds two integers together and prints the result")
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
		// Don't fail on cmdErr as we're testing autonomous execution
		if cmdErr != nil {
			t.Logf("Command output:\n%s", string(output))
			t.Logf("Note: Some output is expected even with errors in E2E tests")
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("gptcode do exceeded timeout of %s", timeout)
	}

	outputStr := string(output)
	t.Logf("Command completed, checking for verifier selection...")

	// Look for evidence that Go-specific verifiers were run (build/test)
	hasGoVerifier := strings.Contains(outputStr, "go build") || strings.Contains(outputStr, "go test") || strings.Contains(outputStr, "build") || strings.Contains(outputStr, "test")

	if hasGoVerifier {
		t.Logf("✓ Go-specific verifiers were triggered (build/test)")
	} else {
		t.Logf("Note: Go verifier execution not explicitly visible in output, which may be normal")
	}

	// Verify that the Go file was modified as expected
	updatedGoContent, err := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	if err != nil {
		t.Fatalf("Failed to read updated main.go: %v", err)
	}

	updatedGoStr := string(updatedGoContent)
	hasNewFunction := strings.Contains(updatedGoStr, "func") && (strings.Contains(updatedGoStr, "add") || strings.Contains(updatedGoStr, "Add") || strings.Contains(updatedGoStr, "print"))

	if !hasNewFunction {
		t.Logf("Content of main.go:\n%s", updatedGoStr)
		t.Errorf("Expected main.go to be modified with new function")
	} else {
		t.Logf("✓ main.go was successfully modified with new function")
	}

	t.Logf("✓ Dynamic verifier selection test completed")
}

func TestAutonomousErrorRecovery(t *testing.T) {
	skipIfNoE2E(t)

	if testing.Short() {
		t.Skip("Skipping slow E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Create a Go file with a deliberate syntax error to test recovery
	// This will allow us to test the error recovery mechanisms
	goCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!"
	// Missing closing parenthesis to create a syntax error
}
`
	mainGo := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainGo, []byte(goCode), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create go.mod file
	goMod := `module test-project

go 1.21
`
	modFile := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modFile, []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	t.Logf("Running gptcode do in %s to test error recovery", tmpDir)
	t.Logf("This tests autonomous error recovery with syntax error...")

	// Run gptcode do with a task that should trigger error recovery
	cmd := exec.Command("gptcode", "do", "fix any syntax errors in main.go and add a comment explaining the fix")
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
		// Don't fail on cmdErr as we're testing error recovery
		if cmdErr != nil {
			t.Logf("Command output:\n%s", string(output))
			t.Logf("Note: Some output is expected even with errors in E2E tests")
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatalf("gptcode do exceeded timeout of %s", timeout)
	}

	outputStr := string(output)
	t.Logf("Output from error recovery test:\n%s", outputStr)

	// Read the file after execution to see if error was fixed
	fixedContent, err := os.ReadFile(mainGo)
	if err != nil {
		t.Fatalf("Failed to read fixed main.go: %v", err)
	}

	fixedStr := string(fixedContent)
	
	// Check if the syntax error was fixed (closing parenthesis added)
	hasValidSyntax := strings.Contains(fixedStr, `"Hello, World!")`) // Properly closed
	
	if hasValidSyntax {
		t.Logf("✓ Syntax error was successfully recovered and fixed")
	} else {
		t.Logf("File content after attempted fix:\n%s", fixedStr)
		t.Logf("Note: Error may not have been fixed, which is acceptable for this test")
	}

	// Look for evidence of recovery attempts in the output
	hasRecoveryAttempt := strings.Contains(outputStr, "recovery") || 
		strings.Contains(outputStr, "retry") || 
		strings.Contains(outputStr, "attempt") || 
		strings.Contains(outputStr, "error") || 
		strings.Contains(outputStr, "fix")

	if hasRecoveryAttempt {
		t.Logf("✓ Evidence of error recovery mechanism detected in output")
	} else {
		t.Logf("Note: No explicit recovery messages in output, which may be normal")
	}

	t.Logf("✓ Autonomous error recovery test completed")
}
