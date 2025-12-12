package modes

import (
	"context"
	"fmt"

	"gptcode/internal/agents"
	"gptcode/internal/autonomous"
	"gptcode/internal/config"
	"gptcode/internal/events"
	"gptcode/internal/llm"
	"gptcode/internal/maestro"
)

// AutonomousExecutor wraps autonomous execution for use across modes
type AutonomousExecutor struct {
	events   *events.Emitter
	provider llm.Provider
	cwd      string
	model    string
	executor *autonomous.Executor
}

// NewAutonomousExecutor creates a new autonomous executor
func NewAutonomousExecutor(provider llm.Provider, cwd string, model string, language string) *AutonomousExecutor {
	return NewAutonomousExecutorWithBackend(provider, cwd, model, language, "")
}

// NewAutonomousExecutorWithBackend creates executor with specific backend override
func NewAutonomousExecutorWithBackend(provider llm.Provider, cwd string, model string, language string, backendName string) *AutonomousExecutor {
	// Load setup
	setup, err := config.LoadSetup()
	if err != nil {
		fmt.Printf("[WARN] Failed to load setup: %v, using defaults\n", err)
		// Create minimal setup
		setup = &config.Setup{
			Backend: make(map[string]config.BackendConfig),
		}
		setup.Defaults.Backend = "groq"
	}

	// Override backend if specified (for retry system)
	if backendName != "" {
		setup.Defaults.Backend = backendName
	}

	// Create model selector
	selector, err := config.NewModelSelector(setup)
	if err != nil {
		fmt.Printf("[WARN] Failed to create model selector: %v\n", err)
	}

	// Create Maestro
	conductor := maestro.NewConductor(selector, setup, cwd, language)

	// Create autonomous components
	classifier := agents.NewClassifier(provider, model)
	analyzer := autonomous.NewTaskAnalyzer(classifier, provider, cwd, model)

	executor := autonomous.NewExecutor(analyzer, conductor, cwd)

	return &AutonomousExecutor{
		events:   events.NewEmitter(nil),
		provider: provider,
		cwd:      cwd,
		model:    model,
		executor: executor,
	}
}

// Execute runs autonomous execution with Symphony pattern if task is complex
func (a *AutonomousExecutor) Execute(ctx context.Context, task string) error {
	// Delegate to autonomous executor
	return a.executor.Execute(ctx, task)
}

// ShouldUseAutonomous determines if a task should use autonomous mode
// This is a lightweight heuristic check before full analysis.
// The real complexity scoring happens in TaskAnalyzer.estimateComplexity()
func ShouldUseAutonomous(ctx context.Context, task string) bool {
	// Always return false - let the ML-based complexity analysis decide
	// TaskAnalyzer.Analyze() will trigger Symphony if complexity >= 7
	return false
}
