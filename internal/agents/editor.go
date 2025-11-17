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
}

func NewEditor(provider llm.Provider, cwd string) *EditorAgent {
	return &EditorAgent{
		provider: provider,
		cwd:      cwd,
	}
}

const editorPrompt = `You are a code editor. Your ONLY job is to modify files.

WORKFLOW:
1. Call read_file to get current content
2. Modify the content in your response
3. Call write_file with the COMPLETE modified file

CRITICAL RULES FOR write_file:
- The "content" parameter MUST contain the ENTIRE file text
- NEVER use placeholders like "[previous content]" or "[rest of file]"
- NEVER use references like "[read_file path=...]"
- You must provide EVERY LINE of the file, even unchanged lines

Example:
User: "Remove line 3 from test.go"
You:
1. read_file(path="test.go") → Returns 10 lines
2. Remove line 3 from the text
3. write_file(path="test.go", content="line1\nline2\nline4\nline5\n...line10")

Be direct. No explanations unless there's an error.`

func (e *EditorAgent) Execute(ctx context.Context, userMessage string) (string, error) {
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
	}

	messages := []llm.ChatMessage{
		{Role: "user", Content: userMessage},
	}

	maxIterations := 5
	for i := 0; i < maxIterations; i++ {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[EDITOR] Iteration %d/%d\n", i+1, maxIterations)
		}

		resp, err := e.provider.Chat(ctx, llm.ChatRequest{
			SystemPrompt: editorPrompt,
			Messages:     messages,
			Tools:        toolDefs,
			Model:        "llama-3.3-70b-versatile",
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
