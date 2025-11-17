package modes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chuchu/internal/llm"
	"chuchu/internal/prompt"
	"chuchu/internal/tools"
)

func RunResearch(builder *prompt.Builder, provider llm.Provider, model string, args []string) error {
	question := ""
	if len(args) > 0 {
		question = strings.Join(args, " ")
	}

	if question == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Fprintln(os.Stderr, "Research mode - Document codebase as-is")
		fmt.Fprintln(os.Stderr, "\nWhat would you like to research?")
		fmt.Fprint(os.Stderr, "> ")
		if scanner.Scan() {
			question = scanner.Text()
		}
	}

	if question == "" {
		return fmt.Errorf("no research question provided")
	}

	home, _ := os.UserHomeDir()
	researchDir := filepath.Join(home, ".chuchu", "research")
	os.MkdirAll(researchDir, 0755)

	sys := builder.BuildSystemPrompt(prompt.BuildOptions{
		Lang: "general",
		Mode: "research",
		Hint: question,
	})

	researchPrompt := fmt.Sprintf(`You are conducting research on a codebase to answer this question:

%s

CRITICAL: Your job is to DOCUMENT the codebase as it exists today, NOT to suggest improvements.

Process:
1. Use available tools to explore the codebase proactively
2. Read relevant files completely (use read_file tool)
3. Search for patterns (use search_code tool)
4. List directory structures (use list_files tool)
5. Synthesize findings into a clear research document

Your research document should include:

## Research Question
[The original question]

## Summary
[High-level findings that answer the question]

## Detailed Findings
[Specific components, files, and how they work]
- Component/File: path/to/file.ext
  - What it does
  - How it connects to other parts
  - Key implementation details

## Code References
[Specific file:line references with descriptions]

## Architecture
[Current patterns, conventions, and design found in the codebase]

Remember: Document what IS, not what SHOULD BE. No recommendations or improvements.

Begin your research now.`, question)

	cwd, _ := os.Getwd()
	toolsRaw := tools.GetAvailableTools()
	var availableTools []interface{}
	for _, t := range toolsRaw {
		availableTools = append(availableTools, t)
	}

	messages := []llm.ChatMessage{
		{Role: "user", Content: researchPrompt},
	}

	maxIterations := 15
	var fullResearch strings.Builder

	for i := 0; i < maxIterations; i++ {
		req := llm.ChatRequest{
			SystemPrompt: sys,
			Model:        model,
			Tools:        availableTools,
			Messages:     messages,
		}

		resp, err := provider.Chat(context.Background(), req)
		if err != nil {
			return fmt.Errorf("LLM error: %w", err)
		}

		if resp.Text != "" {
			fmt.Println(resp.Text)
			fullResearch.WriteString(resp.Text)
			fullResearch.WriteString("\n\n")
		}

		if len(resp.ToolCalls) == 0 {
			break
		}

		messages = append(messages, llm.ChatMessage{
			Role:      "assistant",
			Content:   resp.Text,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			var args map[string]interface{}
			json.Unmarshal([]byte(tc.Arguments), &args)

			toolCall := tools.ToolCall{
				Name:      tc.Name,
				Arguments: args,
			}

			result := tools.ExecuteTool(toolCall, cwd)

			if result.Error != "" {
				messages = append(messages, llm.ChatMessage{
					Role:       "tool",
					Content:    fmt.Sprintf("Error: %s", result.Error),
					Name:       tc.Name,
					ToolCallID: tc.ID,
				})
			} else {
				messages = append(messages, llm.ChatMessage{
					Role:       "tool",
					Content:    result.Result,
					Name:       tc.Name,
					ToolCallID: tc.ID,
				})
			}
		}
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	sanitizedQuestion := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		if r == ' ' {
			return '-'
		}
		return -1
	}, question)
	if len(sanitizedQuestion) > 50 {
		sanitizedQuestion = sanitizedQuestion[:50]
	}

	filename := fmt.Sprintf("%s_%s.md", timestamp, sanitizedQuestion)
	researchPath := filepath.Join(researchDir, filename)

	err := os.WriteFile(researchPath, []byte(fullResearch.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Could not save research to %s: %v\n", researchPath, err)
	} else {
		fmt.Fprintf(os.Stderr, "\n\n✓ Research saved to: %s\n", researchPath)
	}

	return nil
}
