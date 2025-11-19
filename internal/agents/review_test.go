package agents

import (
	"context"
	"testing"

	"chuchu/internal/llm"
)

type MockProvider struct {
	Response    string
	ToolCalls   []llm.ChatToolCall
	CallCount   int
	Responses   []string
	ToolCallsAt [][]llm.ChatToolCall
}

func (m *MockProvider) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	if len(m.Responses) > 0 && m.CallCount < len(m.Responses) {
		resp := &llm.ChatResponse{
			Text: m.Responses[m.CallCount],
		}
		if len(m.ToolCallsAt) > m.CallCount {
			resp.ToolCalls = m.ToolCallsAt[m.CallCount]
		}
		m.CallCount++
		return resp, nil
	}

	return &llm.ChatResponse{
		Text:      m.Response,
		ToolCalls: m.ToolCalls,
	}, nil
}

func TestReviewAgent(t *testing.T) {
	t.Run("simple review without tools", func(t *testing.T) {
		mock := &MockProvider{
			Response: "Code looks good.",
		}

		agent := NewReview(mock, ".", "test-model")
		result, err := agent.Execute(context.Background(), []llm.ChatMessage{{Role: "user", Content: "review main.go"}}, nil)

		if err != nil {
			t.Fatalf("ReviewAgent failed: %v", err)
		}
		if result != "Code looks good." {
			t.Errorf("Expected 'Code looks good.', got '%s'", result)
		}
	})

	t.Run("review with tool calls", func(t *testing.T) {
		mock := &MockProvider{
			Responses: []string{
				"Need to read file",
				"File analyzed. Found issues.",
			},
			ToolCallsAt: [][]llm.ChatToolCall{
				{
					{ID: "1", Name: "read_file", Arguments: `{"path": "test.go"}`},
				},
				{},
			},
		}

		agent := NewReview(mock, ".", "test-model")
		result, err := agent.Execute(context.Background(), []llm.ChatMessage{{Role: "user", Content: "review test.go"}}, nil)

		if err != nil {
			t.Fatalf("ReviewAgent with tools failed: %v", err)
		}
		if result == "" {
			t.Error("Expected non-empty result")
		}
	})

	t.Run("with status callback", func(t *testing.T) {
		mock := &MockProvider{
			Response: "Analysis complete.",
		}

		var statusUpdates []string
		callback := func(status string) {
			statusUpdates = append(statusUpdates, status)
		}

		agent := NewReview(mock, ".", "test-model")
		_, err := agent.Execute(context.Background(), []llm.ChatMessage{{Role: "user", Content: "review"}}, callback)

		if err != nil {
			t.Fatalf("ReviewAgent with callback failed: %v", err)
		}
		if len(statusUpdates) == 0 {
			t.Error("Expected status updates but got none")
		}
	})
}
