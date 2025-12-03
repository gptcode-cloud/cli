package agents

import (
	"context"
	"fmt"
	"os"
	"strings"

	"chuchu/internal/llm"
	"chuchu/internal/tools"
)

type ReviewerAgent struct {
	provider llm.Provider
	cwd      string
	model    string
}

type ReviewResult struct {
	Success     bool
	Issues      []string
	Suggestions string
}

type ReviewerConfig struct {
	OnValidationFail func(issues []string) (shouldRetry bool, newBackend string, newModel string)
}

func NewReviewer(provider llm.Provider, cwd string, model string) *ReviewerAgent {
	return &ReviewerAgent{
		provider: provider,
		cwd:      cwd,
		model:    model,
	}
}

const reviewerPrompt = `You are a code reviewer. Your job is to verify if changes meet the success criteria.

WORKFLOW:
1. Read the files that were modified
2. Run commands to verify (build, test, lint, etc)
3. Compare against the success criteria
4. Report pass/fail with specific issues

CRITICAL RULES:
- ONLY run build/compile commands if code files were modified
- Skip build if no files were modified (read-only tasks like git status, gh pr list)
- Skip build if only documentation files (.md, .txt, .json) were modified
- For Go code changes: run 'go build' to check compilation
- For TypeScript/Node code changes: run 'npm run build' or 'tsc'
- For Python code changes: check syntax with 'python -m py_compile file.py'
- Be specific about what's wrong
- If something is missing, say exactly what
- If criteria is met, say "SUCCESS"
- Focus on the actual requirements, not style

EXAMPLE 1 - Validation SUCCESS:
Success Criteria:
  - Tests pass: go test ./auth/...
  - File auth/handler.go contains Login function
  - Lints clean: golangci-lint run

Validation:
  1. Read auth/handler.go → contains func Login()
  2. Would run: go test ./auth/... → (assume passes)
  3. Would run: golangci-lint run → (assume clean)

Result: SUCCESS

EXAMPLE 2 - Validation FAIL with specific issues:
Success Criteria:
  - Tests pass: make test
  - File middleware/jwt.go exports VerifyToken function
  - No hardcoded secrets

Validation:
  1. Read middleware/jwt.go
  
Issues found:
  - FAIL: VerifyToken is not exported (lowercase verifyToken)
  - FAIL: JWT secret is hardcoded on line 15: "hardcoded-secret-123"
  - Tests: Would need to run make test to verify

Result:
FAIL

Issues:
- VerifyToken function must be exported (capitalize: VerifyToken)
- Remove hardcoded secret on line 15, use environment variable
- Run make test to verify tests pass

EXAMPLE 3 - Clear issue reporting:
BAD:
  "Something is wrong with the file"
  
GOOD:
  "middleware/auth.go line 42: Missing error handling for jwt.Parse"
  "tests/auth_test.go: No test for Login with invalid password case"

Be direct and precise.`

func (v *ReviewerAgent) Review(ctx context.Context, plan string, modifiedFiles []string, statusCallback StatusCallback) (*ReviewResult, error) {
	if statusCallback != nil {
		statusCallback("Reviewer: Analyzing changes...")
	}

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
				"name":        "run_command",
				"description": "Run shell command to verify build, tests, linter, etc",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "Command to execute (e.g. 'go build', 'npm test')",
						},
					},
					"required": []string{"command"},
				},
			},
		},
	}

	filesStr := ""
	if len(modifiedFiles) > 0 {
		filesStr = fmt.Sprintf("\nFiles that were modified: %v", modifiedFiles)
	}

	reviewPrompt := fmt.Sprintf(`Validate if the implementation meets the requirements.

Plan and Success Criteria:
---
%s
---
%s

TASK:
1. Read the modified files to see what was changed (if any)
2. **ONLY** run build commands if code files (.go, .py, .js, .ts, etc) were modified
3. Skip build if:
   - No files were modified (read-only task like 'git status')
   - Only docs (.md, .txt, .json, .yaml) were modified
4. Check if changes meet the success criteria from the plan
5. Report:
   - Say "SUCCESS" if all criteria are met
   - If build was needed and failed, report "FAIL" with specific errors
   - If requirements not met, list specific issues

Be smart: Don't run 'go build' if this is not a Go project or no Go files changed.
Be precise and specific.`, plan, filesStr)

	history := []llm.ChatMessage{
		{Role: "user", Content: reviewPrompt},
	}

	maxIterations := 3
	for i := 0; i < maxIterations; i++ {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[VALIDATOR] Iteration %d/%d\n", i+1, maxIterations)
		}

		resp, err := v.provider.Chat(ctx, llm.ChatRequest{
			SystemPrompt: reviewerPrompt,
			Messages:     history,
			Tools:        toolDefs,
			Model:        v.model,
		})
		if err != nil {
			return nil, err
		}

		if len(resp.ToolCalls) == 0 {
			result := &ReviewResult{
				Success:     false,
				Issues:      []string{},
				Suggestions: resp.Text,
			}

			if containsSuccess(resp.Text) {
				result.Success = true
			} else {
				result.Issues = extractIssues(resp.Text)
			}

			if os.Getenv("CHUCHU_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[VALIDATOR] Result: success=%v, issues=%v\n", result.Success, result.Issues)
			}

			return result, nil
		}

		history = append(history, llm.ChatMessage{
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
			result := tools.ExecuteToolFromLLM(llmCall, v.cwd)

			content := result.Result
			if result.Error != "" {
				content = "Error: " + result.Error
			}
			if content == "" {
				content = "Success"
			}

			history = append(history, llm.ChatMessage{
				Role:       "tool",
				Content:    content,
				Name:       tc.Name,
				ToolCallID: tc.ID,
			})
		}
	}

	return &ReviewResult{
		Success:     false,
		Issues:      []string{"Validator reached max iterations"},
		Suggestions: "Unable to complete review",
	}, nil
}

func containsSuccess(text string) bool {
	lowerText := strings.ToLower(text)

	// Check for explicit success statements
	if strings.Contains(lowerText, "task is complete") ||
		strings.Contains(lowerText, "task was completed") ||
		strings.Contains(lowerText, "success criteria are met") ||
		strings.Contains(lowerText, "all criteria are met") ||
		strings.Contains(lowerText, "requirements are met") ||
		strings.Contains(lowerText, "successfully completed") {
		return true
	}

	// Fallback: explicit SUCCESS keyword without failure indicators
	if strings.Contains(lowerText, " success") || strings.HasPrefix(lowerText, "success") {
		hasFail := strings.Contains(lowerText, " fail") ||
			strings.Contains(lowerText, "failed") ||
			strings.Contains(lowerText, " error") ||
			strings.Contains(lowerText, "not met")
		return !hasFail
	}

	return false
}

func extractIssues(text string) []string {
	issues := []string{}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if trimmed != "" && (strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "•") || strings.Contains(lower, "issue") || strings.Contains(lower, "missing") || strings.Contains(lower, "error")) {
			issues = append(issues, trimmed)
		}
	}

	if len(issues) == 0 {
		issues = append(issues, text)
	}

	return issues
}
