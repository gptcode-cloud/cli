package modes

import (
	"context"

	"chuchu/internal/agents"
	"chuchu/internal/autonomous"
	"chuchu/internal/events"
	"chuchu/internal/llm"
)

// AutonomousExecutor wraps autonomous execution for use across modes
type AutonomousExecutor struct {
	events       *events.Emitter
	provider     llm.Provider
	baseProvider llm.Provider
	cwd          string
	model        string
	editorModel  string
	executor     *autonomous.Executor
}

// NewAutonomousExecutor creates a new autonomous executor
func NewAutonomousExecutor(provider llm.Provider, baseProvider llm.Provider, cwd string, model string, editorModel string) *AutonomousExecutor {
	// Create autonomous components
	classifier := agents.NewClassifier(provider, model)
	analyzer := autonomous.NewTaskAnalyzer(classifier, provider, cwd, model)
	planner := agents.NewPlanner(provider, model)
	editor := agents.NewEditor(baseProvider, cwd, editorModel)
	validator := agents.NewValidator(baseProvider, cwd, model)

	executor := autonomous.NewExecutor(analyzer, planner, editor, validator, cwd)

	return &AutonomousExecutor{
		events:       events.NewEmitter(nil),
		provider:     provider,
		baseProvider: baseProvider,
		cwd:          cwd,
		model:        model,
		editorModel:  editorModel,
		executor:     executor,
	}
}

// Execute runs autonomous execution with Symphony pattern if task is complex
func (a *AutonomousExecutor) Execute(ctx context.Context, task string) error {
	// Delegate to autonomous executor
	return a.executor.Execute(ctx, task)
}

// ShouldUseAutonomous determines if a task should use autonomous mode
func ShouldUseAutonomous(ctx context.Context, task string) bool {
	// Check if task mentions "reorganize", "refactor", "unify", etc.
	// These are strong indicators of complex multi-step tasks
	complexVerbs := []string{"reorganize", "refactor", "unify", "merge", "split", "restructure"}

	for _, verb := range complexVerbs {
		if contains(task, verb) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
