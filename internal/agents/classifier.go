package agents

import (
	"context"
	"fmt"
	"strings"

	"chuchu/internal/config"
	"chuchu/internal/llm"
	"chuchu/internal/ml"
)

type Intent string

const (
	IntentQuery    Intent = "query"
	IntentEdit     Intent = "edit"
	IntentResearch Intent = "research"
	IntentTest     Intent = "test"
	IntentReview   Intent = "review"
)

type Classifier struct {
	provider llm.Provider
	model    string
}

func NewClassifier(provider llm.Provider, model string) *Classifier {
	return &Classifier{
		provider: provider,
		model:    model,
	}
}

const routerPrompt = `You are a request classifier. Analyze the user's request and classify it into ONE category.

Categories:
- "query": User wants to READ/UNDERSTAND code (list files, read file, search, explain)
- "edit": User wants to MODIFY code (add, remove, change, refactor files)
- "research": User wants EXTERNAL information (web search, documentation lookup)
- "test": User wants to RUN tests or commands
- "review": User wants CODE REVIEW or CRITIQUE (analyze, check for bugs, improve quality)

Respond with ONLY the category name, nothing else.

Examples:
"list go files" → query
"remove TODO comment from main.go" → edit
"what is the capital of France?" → research
"run tests" → test
"review this file" → review
"check for bugs in main.go" → review
"explain how authentication works" → query
"add error handling to user.go" → edit`

func (c *Classifier) ClassifyIntent(ctx context.Context, userMessage string) (Intent, error) {
	p, err := ml.LoadEmbedded("intent")
	if err == nil {
		setup, _ := config.LoadSetup()
		threshold := setup.Defaults.MLIntentThreshold
		if threshold == 0 {
			threshold = 0.7
		}
		
		label, probs := p.Predict(userMessage)
		confidence := probs[label]
		
		if confidence >= threshold {
			intent := mapMLLabelToIntent(label)
			if intent != "" {
				return intent, nil
			}
		}
	}
	
	resp, err := c.provider.Chat(ctx, llm.ChatRequest{
		SystemPrompt: routerPrompt,
		UserPrompt:   userMessage,
		Model:        c.model,
	})
	if err != nil {
		return "", err
	}

	intent := strings.TrimSpace(strings.ToLower(resp.Text))

	switch intent {
	case "query":
		return IntentQuery, nil
	case "edit", "editor":
		return IntentEdit, nil
	case "research":
		return IntentResearch, nil
	case "test":
		return IntentTest, nil
	case "review":
		return IntentReview, nil
	default:
		return IntentQuery, fmt.Errorf("unknown intent: %s", intent)
	}
}

func mapMLLabelToIntent(label string) Intent {
	switch label {
	 case "router":
		return IntentQuery
	case "query":
		return IntentQuery
	case "editor":
		return IntentEdit
	case "research":
		return IntentResearch
	case "review":
		return IntentReview
	default:
		return ""
	}
}
