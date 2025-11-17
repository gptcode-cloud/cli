package agents

import (
	"context"
	"fmt"
	"strings"

	"chuchu/internal/llm"
)

type Intent string

const (
	IntentQuery    Intent = "query"
	IntentEdit     Intent = "edit"
	IntentResearch Intent = "research"
	IntentTest     Intent = "test"
)

type RouterAgent struct {
	provider llm.Provider
}

func NewRouter(provider llm.Provider) *RouterAgent {
	return &RouterAgent{provider: provider}
}

const routerPrompt = `You are a request classifier. Analyze the user's request and classify it into ONE category.

Categories:
- "query": User wants to READ/UNDERSTAND code (list files, read file, search, explain)
- "edit": User wants to MODIFY code (add, remove, change, refactor files)
- "research": User wants EXTERNAL information (web search, documentation lookup)
- "test": User wants to RUN tests or commands

Respond with ONLY the category name, nothing else.

Examples:
"list go files" → query
"remove TODO comment from main.go" → edit
"what is the capital of France?" → research
"run tests" → test
"explain how authentication works" → query
"add error handling to user.go" → edit`

func (r *RouterAgent) ClassifyIntent(ctx context.Context, userMessage string) (Intent, error) {
	resp, err := r.provider.Chat(ctx, llm.ChatRequest{
		SystemPrompt: routerPrompt,
		UserPrompt:   userMessage,
		Model:        "llama-3.3-70b-versatile",
	})
	if err != nil {
		return "", err
	}

	intent := strings.TrimSpace(strings.ToLower(resp.Text))
	
	switch intent {
	case "query":
		return IntentQuery, nil
	case "edit":
		return IntentEdit, nil
	case "research":
		return IntentResearch, nil
	case "test":
		return IntentTest, nil
	default:
		return IntentQuery, fmt.Errorf("unknown intent: %s", intent)
	}
}
