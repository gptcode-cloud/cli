package autonomous

import (
	"context"
	"testing"

	"chuchu/internal/agents"
	"chuchu/internal/llm"
)

// Integration test for full Symphony execution
func TestSymphonyExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Mock provider
	provider := &mockSymphonyProvider{
		complexity: 8, // Trigger symphony decomposition
		movements: []Movement{
			{
				ID:              "m1",
				Name:            "Step 1",
				Goal:            "Create test file",
				Dependencies:    []string{},
				SuccessCriteria: []string{"file exists"},
				Status:          "pending",
			},
		},
	}

	classifier := agents.NewClassifier(provider, "test-model")
	analyzer := NewTaskAnalyzer(classifier, provider, "/tmp", "test-model")
	planner := agents.NewPlanner(provider, "test-model")
	editor := agents.NewEditorWithFileValidation(provider, "/tmp", "test-model", []string{})
	validator := agents.NewValidator(provider, "/tmp", "test-model")

	executor := NewExecutor(analyzer, planner, editor, validator, "/tmp")

	// NOTE: This would actually try to execute - keeping as placeholder
	// In real test, we'd need more sophisticated mocking
	_ = executor

	t.Skip("Full integration test requires complete mock setup")
}

// Mock provider for symphony tests
type mockSymphonyProvider struct {
	complexity int
	movements  []Movement
}

func (m *mockSymphonyProvider) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	// Return complexity score or movement decomposition based on prompt
	if req.UserPrompt != "" && len(req.UserPrompt) > 100 {
		// This is a complexity scoring request
		return &llm.ChatResponse{
			Text: "8",
		}, nil
	}

	// This is a movement decomposition request
	return &llm.ChatResponse{
		Text: `[
			{
				"id": "m1",
				"name": "Test Movement",
				"description": "Test description",
				"goal": "Test goal",
				"dependencies": [],
				"required_files": [],
				"output_files": ["test.txt"],
				"success_criteria": ["file exists"]
			}
		]`,
	}, nil
}
