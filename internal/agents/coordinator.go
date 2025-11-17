package agents

import (
	"context"
	"fmt"
	"os"

	"chuchu/internal/llm"
)

type Coordinator struct {
	router       *RouterAgent
	editor       *EditorAgent
	query        *QueryAgent
	research     *ResearchAgent
}

func NewCoordinator(
	provider llm.Provider,
	orchestrator *llm.OrchestratorProvider,
	cwd string,
) *Coordinator {
	return &Coordinator{
		router:   NewRouter(provider),
		editor:   NewEditor(provider, cwd),
		query:    NewQuery(provider, cwd),
		research: NewResearch(orchestrator),
	}
}

func (c *Coordinator) Execute(ctx context.Context, userMessage string) (string, error) {
	intent, err := c.router.ClassifyIntent(ctx, userMessage)
	if err != nil {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[COORDINATOR] Router error: %v, defaulting to query\n", err)
		}
		intent = IntentQuery
	}

	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[COORDINATOR] Intent classified as: %s\n", intent)
	}

	switch intent {
	case IntentEdit:
		return c.editor.Execute(ctx, userMessage)
	case IntentQuery:
		return c.query.Execute(ctx, userMessage)
	case IntentResearch:
		return c.research.Execute(ctx, userMessage)
	case IntentTest:
		return c.editor.Execute(ctx, userMessage)
	default:
		return c.query.Execute(ctx, userMessage)
	}
}
