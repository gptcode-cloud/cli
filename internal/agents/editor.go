package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"chuchu/internal/llm"
	"chuchu/internal/tools"
)

type EditorAgent struct {
	provider     llm.Provider
	cwd          string
	model        string
	allowedFiles []string
}

func NewEditor(provider llm.Provider, cwd string, model string) *EditorAgent {
	return &EditorAgent{
		provider:     provider,
		cwd:          cwd,
		model:        model,
		allowedFiles: nil,
	}
}

func NewEditorWithFileValidation(provider llm.Provider, cwd string, model string, allowedFiles []string) *EditorAgent {
	return &EditorAgent{
		provider:     provider,
		cwd:          cwd,
		model:        model,
		allowedFiles: allowedFiles,
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

		if os.Getenv("CHUCHU_DEBUG") == "1" && i == 0 {
			fmt.Fprintf(os.Stderr, "[EDITOR] Messages count: %d\n", len(messages))
			if len(messages) > 0 {
				fmt.Fprintf(os.Stderr, "[EDITOR] First message: %s...\n", messages[0].Content[:min(200, len(messages[0].Content))])
			}
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

		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[EDITOR] Response text length: %d\n", len(resp.Text))
			fmt.Fprintf(os.Stderr, "[EDITOR] Tool calls: %d\n", len(resp.ToolCalls))
			if len(resp.Text) > 0 {
				fmt.Fprintf(os.Stderr, "[EDITOR] Response preview: %s...\n", resp.Text[:min(200, len(resp.Text))])
			}
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

			if tc.Name == "write_file" {
				var argsMap map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Arguments), &argsMap); err == nil {
					if err := e.validateFileWrite(argsMap); err != nil {
						messages = append(messages, llm.ChatMessage{
							Role:       "tool",
							Content:    fmt.Sprintf("Error: %s. Only modify files mentioned in the plan.", err.Error()),
							Name:       tc.Name,
							ToolCallID: tc.ID,
						})
						continue
					}
				}
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

func (e *EditorAgent) validateFileWrite(args map[string]interface{}) error {
	if len(e.allowedFiles) == 0 {
		return nil
	}

	path, ok := args["path"].(string)
	if !ok {
		return nil
	}

	for _, allowed := range e.allowedFiles {
		if path == allowed || strings.HasSuffix(allowed, path) || strings.Contains(allowed, path) {
			return nil
		}
	}

	return &FileValidationError{
		Path:    path,
		Message: fmt.Sprintf("File '%s' is not in the allowed list. Plan mentions: %v", path, e.allowedFiles),
	}
}
