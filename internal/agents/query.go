package agents

import (
	"context"
	"fmt"
	"os"

	"chuchu/internal/llm"
	"chuchu/internal/tools"
)

type QueryAgent struct {
	provider llm.Provider
	cwd      string
	model    string
}

func NewQuery(provider llm.Provider, cwd string, model string) *QueryAgent {
	return &QueryAgent{
		provider: provider,
		cwd:      cwd,
		model:    model,
	}
}

const queryPrompt = `You are a code reader and explainer. Your job is to READ and UNDERSTAND code.

You can:
- List files in directories
- Read file contents
- Search for patterns
- Explain code structure

You CANNOT modify files. Be concise and direct in your explanations.`

func (q *QueryAgent) Execute(ctx context.Context, userMessage string) (string, error) {
	toolDefs := []interface{}{
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "read_file",
				"description": "Read file contents",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "File path",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "list_files",
				"description": "List files in directory",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Directory path",
						},
						"pattern": map[string]interface{}{
							"type":        "string",
							"description": "Glob pattern (e.g., *.go)",
						},
					},
				},
			},
		},
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "search_code",
				"description": "Search for pattern in code",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"pattern": map[string]interface{}{
							"type":        "string",
							"description": "Search pattern",
						},
						"file_pattern": map[string]interface{}{
							"type":        "string",
							"description": "File pattern filter",
						},
					},
					"required": []string{"pattern"},
				},
			},
		},
	}

	messages := []llm.ChatMessage{
		{Role: "user", Content: userMessage},
	}

	maxIterations := 3
	for i := 0; i < maxIterations; i++ {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[QUERY] Iteration %d/%d\n", i+1, maxIterations)
		}

		resp, err := q.provider.Chat(ctx, llm.ChatRequest{
			SystemPrompt: queryPrompt,
			Messages:     messages,
			Tools:        toolDefs,
			Model:        q.model,
		})
		if err != nil {
			return "", err
		}

		if len(resp.ToolCalls) == 0 {
			return resp.Text, nil
		}

		messages = append(messages, llm.ChatMessage{
			Role:      "assistant",
			Content:   resp.Text,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			llmCall := tools.LLMToolCall{
				ID:        tc.ID,
				Name:      tc.Name,
				Arguments: tc.Arguments,
			}
			result := tools.ExecuteToolFromLLM(llmCall, q.cwd)
			
			content := result.Result
			if result.Error != "" {
				content = "Error: " + result.Error
			}
			if content == "" {
				content = "Success"
			}
			
			messages = append(messages, llm.ChatMessage{
				Role:       "tool",
				Content:    content,
				Name:       tc.Name,
				ToolCallID: tc.ID,
			})
		}

		if i >= 1 {
			finalResp, err := q.provider.Chat(ctx, llm.ChatRequest{
				SystemPrompt: queryPrompt + "\n\nProvide your final answer based on the tool results above. Do NOT call more tools.",
				Messages:     messages,
				Model:        q.model,
			})
			if err != nil {
				return "", err
			}
			return finalResp.Text, nil
		}
	}

	return "Query reached max iterations", nil
}
