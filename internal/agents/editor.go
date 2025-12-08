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

const editorPrompt = `You are a code editor and executor. Your job is to modify files AND execute shell commands.

WORKFLOW:
1. For file reading: Call read_file to get current content
2. For shell commands: Call run_command (e.g., "gh pr list", "go test", "npm run lint")
3. For file modification: Call apply_patch for small changes, or write_file for new files/large rewrites
4. **WHEN DONE**: Stop immediately. Do NOT call tools again. Return success message.

CRITICAL RULES:
- Use run_command for ANY shell operation (git, gh, tests, linters, etc)
- Use apply_patch whenever possible to save tokens and reduce risk
- For apply_patch, the "search" block must MATCH EXACTLY (including whitespace)
- For write_file, provide the COMPLETE file content
- NEVER use placeholders like "[previous content]" or "[rest of file]"
- NEVER create fake/placeholder files instead of using run_command
- **IDEMPOTENCY**: Before modifying, check if change already exists. Don't apply same patch twice
- **ONE CHANGE PER FILE**: After modifying a file, do NOT modify it again in same turn
- **GO PACKAGE NAMES**: When editing Go files, NEVER change the package declaration unless explicitly asked. If main.go has "package main", ALL files in the same directory MUST use "package main". Do NOT infer package names from filenames (e.g., utils.go should NOT have "package utils" if it's in a package main directory)

EXAMPLE 1 - Using run_command (for shell operations):
Task: "Get list of open pull requests"

run_command(command="gh pr list --state open --json number,title")

Returns:
  [{"number": 42, "title": "Add new feature"}]

EXAMPLE 2 - Using apply_patch (preferred for small changes):
Task: "Add JWT verification to auth handler"

Step 1: read_file(path="auth/handler.go")
Returns:
  func VerifyToken(token string) bool {
      // TODO: implement
      return false
  }

Step 2: apply_patch(path="auth/handler.go",
  search="func VerifyToken(token string) bool {\n    // TODO: implement\n    return false\n}",
  replace="func VerifyToken(token string) (*Claims, error) {\n    claims := &Claims{}\n    parsed, err := jwt.ParseWithClaims(token, claims, keyFunc)\n    if err != nil || !parsed.Valid {\n        return nil, err\n    }\n    return claims, nil\n}")

EXAMPLE 3 - Using write_file (for new files):
Task: "Create new config file"

write_file(path="config/app.yaml",
  content="database:\n  host: localhost\n  port: 5432\n  name: myapp\n\nserver:\n  port: 8080\n  debug: false")

EXAMPLE 4 - Exact whitespace matching (CRITICAL):
BAD:
  search="    return false"  # 4 spaces
  (file has 2 spaces → WILL FAIL)

GOOD:
  1. Read file first
  2. Copy EXACT whitespace from file content
  3. search="  return false"  # 2 spaces (matches file)

EXAMPLE 5 - Appending to file (ONE TIME ONLY):
Task: "Add 'Goodbye' to hello.txt"

Step 1: read_file(path="hello.txt")
Returns: "Hello World"

Step 2: apply_patch(path="hello.txt",
  search="Hello World",
  replace="Hello World\nGoodbye")

Step 3: STOP. Return "Line added successfully". DO NOT read file again.

EXAMPLE 6 - Completion (CRITICAL):
After executing ALL required changes:
- Return a brief success message
- DO NOT call any more tools
- DO NOT verify by reading files again unless validation failed

Be direct. No explanations unless there's an error.`

func (e *EditorAgent) Execute(ctx context.Context, history []llm.ChatMessage, statusCallback StatusCallback) (string, []string, error) {
	var modifiedFiles []string
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

	// Higher iteration limit for complex editing tasks
	// The editor may need multiple tool calls to:
	// 1. Read existing files
	// 2. Apply patches or write files
	// 3. Run verification commands
	// 4. Fix errors discovered during execution
	maxIterations := 10
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
			return "", nil, err
		}

		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[EDITOR] Response text length: %d\n", len(resp.Text))
			fmt.Fprintf(os.Stderr, "[EDITOR] Tool calls: %d\n", len(resp.ToolCalls))
			if len(resp.Text) > 0 {
				fmt.Fprintf(os.Stderr, "[EDITOR] Response preview: %s...\n", resp.Text[:min(200, len(resp.Text))])
			}
		}

		if len(resp.ToolCalls) == 0 {
			parsedCalls := llm.ParseToolCallsFromText(resp.Text)
			if len(parsedCalls) > 0 {
				if os.Getenv("CHUCHU_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[EDITOR] Parsed %d tool calls from text\n", len(parsedCalls))
				}
				resp.ToolCalls = parsedCalls
				messages = append(messages, llm.ChatMessage{
					Role:      "assistant",
					Content:   "",
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

					if tc.Name == "write_file" || tc.Name == "apply_patch" {
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
					if len(result.ModifiedFiles) > 0 {
						modifiedFiles = append(modifiedFiles, result.ModifiedFiles...)
					}

					content := result.Result
					if result.Error != "" {
						content = "Error: " + result.Error
					}
					if content == "" {
						content = "Success"
					}

					// For read-only operations on pure query tasks, return immediately
					if len(modifiedFiles) == 0 && (tc.Name == "read_file" || (tc.Name == "run_command" && result.Error == "")) {
						if result.Result != "" && result.Error == "" {
							isQueryTask := len(messages) > 0 && !containsEditKeywords(messages[0].Content)
							if isQueryTask {
								if os.Getenv("CHUCHU_DEBUG") == "1" {
									fmt.Fprintf(os.Stderr, "[EDITOR] Early return for query task, result length=%d\n", len(result.Result))
								}
								return result.Result, modifiedFiles, nil
							}
						}
					}

					// Truncate very long content to prevent API errors
					// Max ~10k chars (~2500 tokens) per tool result
					maxContentLength := 10000
					if len(content) > maxContentLength {
						lines := strings.Split(content, "\n")
						if len(lines) > 200 {
							// Show first 100 and last 100 lines for large files
							firstLines := strings.Join(lines[:100], "\n")
							lastLines := strings.Join(lines[len(lines)-100:], "\n")
							content = fmt.Sprintf("%s\n\n... [%d lines omitted] ...\n\n%s",
								firstLines, len(lines)-200, lastLines)
						} else {
							// Just truncate by character count
							content = content[:maxContentLength] + "\n\n... [truncated]"
						}
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
				continue
			}
			return resp.Text, modifiedFiles, nil
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

			if tc.Name == "write_file" || tc.Name == "apply_patch" {
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
			if len(result.ModifiedFiles) > 0 {
				modifiedFiles = append(modifiedFiles, result.ModifiedFiles...)
			}

			content := result.Result
			if result.Error != "" {
				content = "Error: " + result.Error
			}
			if content == "" {
				content = "Success"
			}

			// For read-only operations on pure query tasks, return immediately
			if len(modifiedFiles) == 0 && (tc.Name == "read_file" || (tc.Name == "run_command" && result.Error == "")) {
				if result.Result != "" && result.Error == "" {
					isQueryTask := len(messages) > 0 && !containsEditKeywords(messages[0].Content)
					if isQueryTask {
						if os.Getenv("CHUCHU_DEBUG") == "1" {
							fmt.Fprintf(os.Stderr, "[EDITOR] Early return for query task (path 2), result length=%d\n", len(result.Result))
						}
						return result.Result, modifiedFiles, nil
					}
				}
			}

			// Truncate very long content to prevent API errors
			// Max ~10k chars (~2500 tokens) per tool result
			maxContentLength := 10000
			if len(content) > maxContentLength {
				lines := strings.Split(content, "\n")
				if len(lines) > 200 {
					// Show first 100 and last 100 lines for large files
					firstLines := strings.Join(lines[:100], "\n")
					lastLines := strings.Join(lines[len(lines)-100:], "\n")
					content = fmt.Sprintf("%s\n\n... [%d lines omitted] ...\n\n%s",
						firstLines, len(lines)-200, lastLines)
				} else {
					// Just truncate by character count
					content = content[:maxContentLength] + "\n\n... [truncated]"
				}
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

	return "Editor reached max iterations", modifiedFiles, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func containsEditKeywords(text string) bool {
	lower := strings.ToLower(text)

	// Strong indicators: plan explicitly mentions files to modify/create
	// But if it says "None" right after, that's NOT an edit
	// Check both "Files to modify:" and "## Files to modify" formats
	if strings.Contains(lower, "files to modify") {
		// Make sure it's not "Files to Modify: None" or "Files to modify\nNone"
		if !strings.Contains(lower, "modify\nnone") &&
			!strings.Contains(lower, "modify: none") &&
			!strings.Contains(lower, "modify:\nnone") {
			return true
		}
	}
	if strings.Contains(lower, "files to create") {
		if !strings.Contains(lower, "create\nnone") &&
			!strings.Contains(lower, "create: none") &&
			!strings.Contains(lower, "create:\nnone") {
			return true
		}
	}

	// Check for file operations (but exclude "change" which appears in plan headings)
	editKeywords := []string{
		"write_file", "apply_patch", // Tool calls
		"modify file", "create file", "update file", "patch file",
		"add to", "append to", "insert into",
		"delete from", "remove from",
		"rename", "move file", "rewrite",
	}
	for _, keyword := range editKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}

	return false
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
