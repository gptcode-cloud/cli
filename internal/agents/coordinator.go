package agents

import (
	"context"
	"fmt"
	"os"

	"chuchu/internal/llm"
)

type Coordinator struct {
	router   *RouterAgent
	editor   *EditorAgent
	query    *QueryAgent
	research *ResearchAgent
	review   *ReviewAgent
}

func NewCoordinator(
	provider llm.Provider,
	orchestrator *llm.OrchestratorProvider,
	cwd string,
	routerModel string,
	editorModel string,
	queryModel string,
	researchModel string,
) *Coordinator {
	return &Coordinator{
		router:   NewRouter(provider, routerModel),
		editor:   NewEditor(provider, cwd, editorModel),
		query:    NewQuery(provider, cwd, queryModel),
		research: NewResearch(orchestrator),
		review:   NewReview(provider, cwd, queryModel), // Use query model for review as it needs good reasoning
	}
}

func (c *Coordinator) Execute(ctx context.Context, history []llm.ChatMessage, statusCallback StatusCallback) (string, error) {
	// Use the last user message for intent classification
	lastMessage := ""
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "user" {
			lastMessage = history[i].Content
			break
		}
	}

	if statusCallback != nil {
		statusCallback("Router: Classifying intent...")
	}

	intent, err := c.router.ClassifyIntent(ctx, lastMessage)
	if err != nil {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[COORDINATOR] Router error: %v, defaulting to query\n", err)
		}
		intent = IntentQuery
	}

	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[COORDINATOR] Intent classified as: %s\n", intent)
	}

	if statusCallback != nil {
		statusCallback(fmt.Sprintf("Coordinator: Routing to %s agent...", intent))
	}

	switch intent {
	case IntentEdit:
		return c.editor.Execute(ctx, history, statusCallback)
	case IntentQuery:
		return c.query.Execute(ctx, history, statusCallback)
	case IntentResearch:
		return c.research.Execute(ctx, history, statusCallback)
	case IntentTest:
		return c.editor.Execute(ctx, history, statusCallback)
	case IntentReview:
		return c.review.Execute(ctx, history, statusCallback)
	default:
		return c.query.Execute(ctx, history, statusCallback)
	}
}
