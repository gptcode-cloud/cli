package repl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Message represents a single chat message
type Message struct {
	Role       string    `json:"role"` // "user" or "assistant"
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	TokenCount int       `json:"token_count"`
}

// ContextManager manages conversation history with sliding window
type ContextManager struct {
	messages    []Message
	maxTokens   int               // Max tokens to keep in context
	maxMessages int               // Max number of messages to keep
	currentDir  string            // Current working directory
	fileContext map[string]string // File path -> content
}

// NewContextManager creates a new context manager
func NewContextManager(maxTokens int, maxMessages int) *ContextManager {
	return &ContextManager{
		messages:    make([]Message, 0),
		maxTokens:   maxTokens,
		maxMessages: maxMessages,
		fileContext: make(map[string]string),
	}
}

// AddMessage adds a message to the conversation history
func (cm *ContextManager) AddMessage(role, content string, tokenCount int) {
	msg := Message{
		Role:       role,
		Content:    content,
		Timestamp:  time.Now(),
		TokenCount: tokenCount,
	}

	cm.messages = append(cm.messages, msg)

	// Ensure window size limits
	cm.enforceLimits()
}

// enforceLimits keeps the conversation within configured limits
func (cm *ContextManager) enforceLimits() {
	// Keep tokens under limit
	totalTokens := cm.getTotalTokens()

	for totalTokens > cm.maxTokens && len(cm.messages) > 2 {
		// Remove the oldest message (but keep at least 2 messages)
		cm.messages = cm.messages[1:]
		totalTokens = cm.getTotalTokens()
	}

	// Keep message count under limit
	for len(cm.messages) > cm.maxMessages && len(cm.messages) > 2 {
		cm.messages = cm.messages[1:]
	}
}

// getTotalTokens calculates total tokens in all messages
func (cm *ContextManager) getTotalTokens() int {
	total := 0
	for _, msg := range cm.messages {
		total += msg.TokenCount
	}
	return total
}

// GetContext returns the formatted conversation history for the AI
func (cm *ContextManager) GetContext() string {
	var parts []string

	for _, msg := range cm.messages {
		role := msg.Role
		switch role {
		case "user":
			role = "User"
		case "assistant":
			role = "Assistant"
		default:
			// Capitalize first letter for other roles
			if len(role) > 0 {
				role = strings.ToUpper(role[:1]) + role[1:]
			}
		}
		part := fmt.Sprintf("%s: %s", role, msg.Content)
		parts = append(parts, part)
	}

	return strings.Join(parts, "\n")
}

// GetLastMessage returns the most recent message
func (cm *ContextManager) GetLastMessage() *Message {
	if len(cm.messages) == 0 {
		return nil
	}
	return &cm.messages[len(cm.messages)-1]
}

// GetUserInput returns the last user message
func (cm *ContextManager) GetUserInput() string {
	for i := len(cm.messages) - 1; i >= 0; i-- {
		if cm.messages[i].Role == "user" {
			return cm.messages[i].Content
		}
	}
	return ""
}

// Clear resets the conversation history
func (cm *ContextManager) Clear() {
	cm.messages = make([]Message, 0)
}

// GetRecentMessages returns the last N messages
func (cm *ContextManager) GetRecentMessages(n int) []Message {
	if n >= len(cm.messages) {
		return cm.messages
	}
	start := len(cm.messages) - n
	return cm.messages[start:]
}

// UpdateFileContext updates the file context for the current directory
func (cm *ContextManager) UpdateFileContext() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	cm.currentDir = cwd

	// Clear old context
	cm.fileContext = make(map[string]string)

	// Read relevant files from current directory
	entries, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Add README files first
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		base := strings.ToLower(filepath.Base(name))

		// Check for README or important files
		if strings.HasPrefix(base, "readme") || strings.HasPrefix(base, "license") {
			if content, err := os.ReadFile(name); err == nil && len(content) < 10000 {
				cm.fileContext[name] = string(content)
			}
		}

		// Add source code files (up to 1000 chars each)
		if isSourceFile(ext) {
			if content, err := os.ReadFile(name); err == nil && len(content) < 1000 {
				cm.fileContext[name] = string(content)
			}
		}
	}

	return nil
}

// isSourceFile checks if a file is a source code file
func isSourceFile(ext string) bool {
	sourceExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true,
		".rb": true, ".java": true, ".php": true, ".rs": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".sh": true, ".bash": true, ".html": true, ".css": true,
		".md": true,
	}
	return sourceExts[ext]
}

// GetFileContext returns the current file context as formatted string
func (cm *ContextManager) GetFileContext() string {
	if len(cm.fileContext) == 0 {
		return ""
	}

	var parts []string
	for filename, content := range cm.fileContext {
		// Limit content size
		if len(content) > 500 {
			content = content[:500] + "\n[...truncated...]"
		}
		parts = append(parts, fmt.Sprintf("### %s\n%s\n", filename, content))
	}

	return strings.Join(parts, "\n")
}

// GetStatus returns current context statistics
func (cm *ContextManager) GetStatus() string {
	totalTokens := cm.getTotalTokens()
	return fmt.Sprintf("Messages: %d/%d, Tokens: %d/%d, Files in context: %d",
		len(cm.messages), cm.maxMessages, totalTokens, cm.maxTokens, len(cm.fileContext))
}

// SaveConversation saves the conversation to a file
func (cm *ContextManager) SaveConversation(filename string) error {
	data, err := json.MarshalIndent(cm.messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadConversation loads a conversation from a file
func (cm *ContextManager) LoadConversation(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var messages []Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to unmarshal messages: %w", err)
	}

	cm.messages = messages
	cm.enforceLimits()
	return nil
}
