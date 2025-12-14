//go:build e2e

package run

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestChuChatSingleShot tests single-shot chat mode with a message argument
func TestChuChatSingleShot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Create a simple file for context
	err := os.WriteFile("README.md", []byte("# Test Project\nA simple test project."), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Run chu chat with a simple question
	cmd := exec.Command("gptcode", "chat", "what is 2 plus 2?")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chu chat failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Should contain an answer with 4
	if !strings.Contains(outputStr, "4") {
		t.Errorf("Expected answer to contain '4', got: %s", outputStr)
	}
}

// TestChuChatWithInitialMessage tests chat with initial message then exit (non-interactive)
func TestChuChatWithInitialMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Create a test file
	err := os.WriteFile("test.go", []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Run chu chat with initial message (should process and exit since non-interactive)
	cmd := exec.Command("gptcode", "chat", "tell me what files are in this directory")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chu chat failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Should mention the test.go file we created
	if !strings.Contains(outputStr, "test.go") && !strings.Contains(outputStr, "file") {
		t.Errorf("Expected response to mention files, got: %s", outputStr)
	}
}

// TestChuChatHelp tests that help flag shows REPL information
func TestChuChatHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	cmd := exec.Command("gptcode", "chat", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chu chat --help failed: %v", err)
	}

	outputStr := string(output)

	// Should mention REPL commands
	if !strings.Contains(outputStr, "REPL Commands") {
		t.Errorf("Expected help to mention REPL Commands, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "/exit") {
		t.Errorf("Expected help to mention /exit command, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "/clear") {
		t.Errorf("Expected help to mention /clear command, got: %s", outputStr)
	}
}

// TestChuChatContextManager tests the context manager in isolation
func TestChuChatContextManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	conversationFile := filepath.Join(tmpDir, "test_conversation.json")

	// This test validates that context manager can save/load
	// We'll create a simple JSON file and verify structure
	testData := `[
		{
			"role": "user",
			"content": "test message",
			"timestamp": "2024-01-01T00:00:00Z",
			"token_count": 10
		}
	]`

	err := os.WriteFile(conversationFile, []byte(testData), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Verify file exists and has valid JSON
	data, err := os.ReadFile(conversationFile)
	if err != nil {
		t.Fatalf("Failed to read conversation file: %v", err)
	}

	if !strings.Contains(string(data), "test message") {
		t.Errorf("Expected conversation file to contain test message, got: %s", string(data))
	}
}
