package agents

import (
	"context"
	"fmt"
	"os"

	"chuchu/internal/llm"
	"chuchu/internal/tools"
)

type EditorAgent struct {
	provider llm.Provider
	cwd      string
	model    string
}

func NewEditor(provider llm.Provider, cwd string, model string) *EditorAgent {
	return &EditorAgent{
		provider: provider,
		cwd:      cwd,
		model:    model,
	}
}

const editorPrompt = `You are a code editor. Your ONLY job is to modify files.

WORKFLOW:
1. Call read_file to get current content
2. Modify the content in your response
3. Call apply_patch for small changes, or write_file for new files/large rewrites.

CRITICAL RULES:
- Use apply_patch whenever possible to save tokens and reduce risk.
- For apply_patch, the "search" block must MATCH EXACTLY (including whitespace).
- For write_file, provide the COMPLETE file content.
- NEVER use placeholders like "[previous content]" or "[rest of file]".

Example (Patch):
User: "Remove line 3 from test.go"
You:
1. read_file(path="test.go")
2. apply_patch(path="test.go", search="line3\n", replace="")

Be direct. No explanations unless there's an error.`

func (e *EditorAgent) Execute(ctx context.Context, history []llm.ChatMessage, statusCallback StatusCallback) (string, error) {
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
				"name":        "write_file",
				"description": "Write COMPLETE file content (all lines)",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "File path",
						},
						"content": map[string]interface{}{
							"type":        "string",
							"description": "FULL file content with ALL lines",
						},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "run_command",
				"description": "Run shell command (tests, linter, etc)",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "Command to execute",
						},
					},
					"required": []string{"command"},
				},
			},
		},
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "project_map",
				"description": "Get project structure",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"max_depth": map[string]interface{}{
							"type":        "integer",
							"description": "Max depth",
						},
					},
				},
			},
		},
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "apply_patch",
				"description": "Replace text block",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "File path",
						},
						"search": map[string]interface{}{
							"type":        "string",
							"description": "Exact text to replace",
						},
						"replace": map[string]interface{}{
							"type":        "string",
							"description": "New text",
						},
					},
					"required": []string{"path", "search", "replace"},
				},
			},
		},
	}

	// Copy history to avoid mutating the original slice in the loop
	messages := make([]llm.ChatMessage, len(history))
	copy(messages, history)

	maxIterations := 5
	for i := 0; i < maxIterations; i++ {
		if statusCallback != nil {
			statusCallback(fmt.Sprintf("Editor: Thinking (Iteration %d/%d)...", i+1, maxIterations))
		}
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[EDITOR] Iteration %d/%d\n", i+1, maxIterations)
		}

		resp, err := e.provider.Chat(ctx, llm.ChatRequest{
			SystemPrompt: editorPrompt,
			Messages:     messages,
			Tools:        toolDefs,
			Model:        e.model,
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
			if statusCallback != nil {
				statusCallback(fmt.Sprintf("Editor: Executing %s...", tc.Name))
			}
			result := tools.ExecuteToolFromLLM(llmCall, e.cwd)

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

			if os.Getenv("CHUCHU_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[EDITOR] Executed %s: %s\n", tc.Name, result.Result[:min(50, len(result.Result))])
			}
		}
	}

	return "Editor reached max iterations", nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
