package llm

import "context"

type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

type ChatRequest struct {
	SystemPrompt string
	UserPrompt   string
	Model        string
}

type ChatResponse struct {
	Text string
}
