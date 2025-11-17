package llm

import "context"

type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

type ChatRequest struct {
	SystemPrompt string
	UserPrompt   string
	Model        string
	Messages     []ChatMessage
	Tools        []interface{}
}

type ChatMessage struct {
	Role       string          `json:"role"`
	Content    string          `json:"content"`
	Name       string          `json:"name,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	ToolCalls  []ChatToolCall  `json:"tool_calls,omitempty"`
}

type ChatResponse struct {
	Text      string
	ToolCalls []ChatToolCall
}

type ChatToolCall struct {
	ID        string
	Name      string
	Arguments string
}
