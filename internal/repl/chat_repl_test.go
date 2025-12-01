package repl

import (
	"testing"
)

func TestContextManagerAddMessage(t *testing.T) {
	cm := NewContextManager(8000, 50)

	cm.AddMessage("user", "Hello", 5)
	cm.AddMessage("assistant", "Hi there!", 10)
	cm.AddMessage("user", "How are you?", 12)

	if len(cm.messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(cm.messages))
	}

	if cm.messages[0].Role != "user" {
		t.Errorf("Expected first message role to be 'user', got '%s'", cm.messages[0].Role)
	}

	if cm.messages[1].Role != "assistant" {
		t.Errorf("Expected second message role to be 'assistant', got '%s'", cm.messages[1].Role)
	}

	totalTokens := cm.getTotalTokens()
	expectedTokens := 5 + 10 + 12
	if totalTokens != expectedTokens {
		t.Errorf("Expected %d total tokens, got %d", expectedTokens, totalTokens)
	}
}

func TestContextManagerGetContext(t *testing.T) {
	cm := NewContextManager(8000, 50)

	cm.AddMessage("user", "What is Go?", 10)
	cm.AddMessage("assistant", "Go is a programming language", 20)
	cm.AddMessage("user", "Who created it?", 12)

	context := cm.GetContext()

	if context == "" {
		t.Error("Expected non-empty context")
	}

	if !contains(context, "What is Go?") {
		t.Error("Context should contain user message")
	}

	if !contains(context, "Go is a programming language") {
		t.Error("Context should contain assistant response")
	}

	if !contains(context, "Who created it?") {
		t.Error("Context should contain follow-up question")
	}
}

func TestContextManagerClear(t *testing.T) {
	cm := NewContextManager(8000, 50)

	cm.AddMessage("user", "Test message 1", 10)
	cm.AddMessage("assistant", "Response 1", 10)

	if len(cm.messages) != 2 {
		t.Errorf("Expected 2 messages before clear, got %d", len(cm.messages))
	}

	cm.Clear()

	if len(cm.messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(cm.messages))
	}
}

func TestContextManagerTokenLimit(t *testing.T) {
	cm := NewContextManager(100, 50) // Small token limit

	// Add messages that exceed token limit
	cm.AddMessage("user", "Message 1", 50)
	cm.AddMessage("assistant", "Response 1", 50)
	cm.AddMessage("user", "Message 2", 50) // This should trigger removal of oldest

	// Should have removed oldest message to stay under limit
	if len(cm.messages) < 2 {
		t.Logf("Context enforced token limit, %d messages remaining", len(cm.messages))
	}

	totalTokens := cm.getTotalTokens()
	if totalTokens > 100 {
		t.Errorf("Token limit exceeded: %d > 100", totalTokens)
	}
}

func TestContextManagerMessageLimit(t *testing.T) {
	cm := NewContextManager(8000, 5) // Max 5 messages

	// Add more than 5 messages
	for i := 0; i < 10; i++ {
		cm.AddMessage("user", "Message", 10)
	}

	if len(cm.messages) > 5 {
		t.Errorf("Message limit exceeded: %d > 5", len(cm.messages))
	}
}

func TestContextManagerGetRecentMessages(t *testing.T) {
	cm := NewContextManager(8000, 50)

	cm.AddMessage("user", "Msg 1", 10)
	cm.AddMessage("assistant", "Resp 1", 10)
	cm.AddMessage("user", "Msg 2", 10)
	cm.AddMessage("assistant", "Resp 2", 10)
	cm.AddMessage("user", "Msg 3", 10)

	recent := cm.GetRecentMessages(2)
	if len(recent) != 2 {
		t.Errorf("Expected 2 recent messages, got %d", len(recent))
	}

	// Should be last 2 messages
	if recent[0].Content != "Resp 2" {
		t.Errorf("Expected 'Resp 2', got '%s'", recent[0].Content)
	}

	if recent[1].Content != "Msg 3" {
		t.Errorf("Expected 'Msg 3', got '%s'", recent[1].Content)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
