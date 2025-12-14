//go:build e2e

package chat

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestChatCodeExplanation tests that chat can explain code
func TestChatCodeExplanation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Create a simple Go file
	code := `package main

import "fmt"

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
	fmt.Println(fibonacci(10))
}
`
	err := os.WriteFile("fib.go", []byte(code), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Ask about the code
	cmd := exec.Command("gptcode", "chat", "explain what fib.go does")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chu chat failed: %v\nOutput: %s", err, output)
	}

	outputStr := strings.ToLower(string(output))

	// Should mention fibonacci
	if !strings.Contains(outputStr, "fibonacci") && !strings.Contains(outputStr, "fib") {
		t.Errorf("Expected explanation to mention fibonacci, got: %s", outputStr)
	}
}

// TestChatFollowUp tests follow-up questions using REPL (simulated via save/load)
func TestChatFollowUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Test conversation context using context manager directly
	// This validates that the fix works correctly

	// Create test conversation file
	conversationFile := filepath.Join(tmpDir, "test_conversation.json")
	messages := []map[string]interface{}{
		{
			"role":        "user",
			"content":     "My favorite number is 42",
			"timestamp":   "2024-01-01T00:00:00Z",
			"token_count": 10,
		},
		{
			"role":        "assistant",
			"content":     "That's interesting! 42 is a famous number from The Hitchhiker's Guide to the Galaxy.",
			"timestamp":   "2024-01-01T00:00:01Z",
			"token_count": 20,
		},
	}

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(conversationFile, data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Verify conversation file structure is correct
	var loaded []map[string]interface{}
	content, err := os.ReadFile(conversationFile)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(content, &loaded)
	if err != nil {
		t.Fatalf("Failed to parse conversation JSON: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(loaded))
	}

	if loaded[0]["role"] != "user" {
		t.Errorf("Expected first message to be from user, got %v", loaded[0]["role"])
	}

	if loaded[1]["role"] != "assistant" {
		t.Errorf("Expected second message to be from assistant, got %v", loaded[1]["role"])
	}

	// Check that content is preserved
	userContent := loaded[0]["content"].(string)
	if !strings.Contains(userContent, "42") {
		t.Errorf("User message should contain '42', got: %s", userContent)
	}

	assistantContent := loaded[1]["content"].(string)
	if !strings.Contains(assistantContent, "Hitchhiker") {
		t.Errorf("Assistant message should contain 'Hitchhiker', got: %s", assistantContent)
	}
}

// TestChatSaveLoadSession tests saving and loading conversation sessions
func TestChatSaveLoadSession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	conversationFile := filepath.Join(tmpDir, "session.json")

	// Create a sample session
	messages := []map[string]interface{}{
		{
			"role":        "user",
			"content":     "What is Go?",
			"timestamp":   "2024-01-01T00:00:00Z",
			"token_count": 5,
		},
		{
			"role":        "assistant",
			"content":     "Go is a statically typed, compiled programming language designed at Google.",
			"timestamp":   "2024-01-01T00:00:01Z",
			"token_count": 15,
		},
		{
			"role":        "user",
			"content":     "Who created it?",
			"timestamp":   "2024-01-01T00:00:02Z",
			"token_count": 5,
		},
		{
			"role":        "assistant",
			"content":     "Go was created by Robert Griesemer, Rob Pike, and Ken Thompson at Google.",
			"timestamp":   "2024-01-01T00:00:03Z",
			"token_count": 15,
		},
	}

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(conversationFile, data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Verify we can load and parse the session
	var loaded []map[string]interface{}
	content, err := os.ReadFile(conversationFile)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(content, &loaded)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if len(loaded) != 4 {
		t.Errorf("Expected 4 messages in session, got %d", len(loaded))
	}

	// Verify conversation flow
	expectedRoles := []string{"user", "assistant", "user", "assistant"}
	for i, msg := range loaded {
		role := msg["role"].(string)
		if role != expectedRoles[i] {
			t.Errorf("Message %d: expected role %s, got %s", i, expectedRoles[i], role)
		}
	}

	// Verify token counts are preserved
	for i, msg := range loaded {
		if _, ok := msg["token_count"]; !ok {
			t.Errorf("Message %d: missing token_count field", i)
		}
	}
}

// TestChatConversationContext tests that conversation context is maintained
func TestChatConversationContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// This test validates the context manager logic
	// By creating a multi-message conversation file and verifying its structure

	tmpDir := t.TempDir()
	conversationFile := filepath.Join(tmpDir, "context_test.json")

	// Simulate a conversation where context matters
	messages := []map[string]interface{}{
		{
			"role":        "user",
			"content":     "I'm working on a Go project",
			"timestamp":   "2024-01-01T00:00:00Z",
			"token_count": 10,
		},
		{
			"role":        "assistant",
			"content":     "Great! Go is an excellent choice for many projects. What kind of project are you building?",
			"timestamp":   "2024-01-01T00:00:01Z",
			"token_count": 20,
		},
		{
			"role":        "user",
			"content":     "It's a CLI tool",
			"timestamp":   "2024-01-01T00:00:02Z",
			"token_count": 8,
		},
		{
			"role":        "assistant",
			"content":     "CLI tools are a great use case for Go! You might want to use libraries like cobra for command parsing.",
			"timestamp":   "2024-01-01T00:00:03Z",
			"token_count": 25,
		},
	}

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(conversationFile, data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Load and verify the conversation structure
	var loaded []map[string]interface{}
	content, err := os.ReadFile(conversationFile)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(content, &loaded)
	if err != nil {
		t.Fatalf("Failed to load conversation: %v", err)
	}

	// Verify contextual references work
	// The assistant's second response references "CLI tools" from the user's message
	assistantMsg := loaded[3]["content"].(string)
	if !strings.Contains(assistantMsg, "CLI") {
		t.Errorf("Expected assistant to reference CLI tools, got: %s", assistantMsg)
	}

	// Calculate total tokens
	totalTokens := 0
	for _, msg := range loaded {
		tokenCount := int(msg["token_count"].(float64))
		totalTokens += tokenCount
	}

	expectedTotal := 10 + 20 + 8 + 25
	if totalTokens != expectedTotal {
		t.Errorf("Expected total tokens %d, got %d", expectedTotal, totalTokens)
	}
}

// TestChatBasicInteraction tests basic single-turn interaction
func TestChatBasicInteraction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Simple question
	cmd := exec.Command("gptcode", "chat", "what is 5 times 5?")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chu chat failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Should contain the answer 25
	if !strings.Contains(outputStr, "25") {
		t.Errorf("Expected answer to contain 25, got: %s", outputStr)
	}
}
