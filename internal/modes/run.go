package modes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"chuchu/internal/llm"
	"chuchu/internal/prompt"
	"chuchu/internal/tools"
)

func RunExecute(builder *prompt.Builder, provider llm.Provider, model string, args []string) error {
	task := ""
	if len(args) > 0 {
		task = strings.Join(args, " ")
	}

	if task == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Fprintln(os.Stderr, "Run mode - Execute any task")
		fmt.Fprintln(os.Stderr, "\nWhat task do you want to execute?")
		fmt.Fprint(os.Stderr, "> ")
		if scanner.Scan() {
			task = scanner.Text()
		}
	}

	if task == "" {
		return fmt.Errorf("no task provided")
	}

	sys := builder.BuildSystemPrompt(prompt.BuildOptions{
		Lang: "general",
		Mode: "run",
		Hint: task,
	})

	sys += "\n\n## EXECUTION MODE - You are a task executor\n\n" +
		"Your job: Execute the user's task IMMEDIATELY using tools. No ceremony, no TDD, just get it done.\n\n" +
		"Common tasks:\n" +
		"- HTTP requests → Use run_command with curl\n" +
		"- CLI operations → Use run_command (fly, docker, kubectl, gh, etc.)\n" +
		"- File operations → Use read_file, list_files\n" +
		"- DevOps tasks → Execute relevant commands\n" +
		"- Multi-step operations → Execute each step, use results in next\n\n" +
		"Examples:\n" +
		"- User: 'GET https://api.github.com/users/octocat'\n" +
		"  YOU: run_command({'command': 'curl -s https://api.github.com/users/octocat'})\n\n" +
		"- User: 'deploy to staging'\n" +
		"  YOU: run_command({'command': 'fly deploy --config fly.staging.toml'})\n\n" +
		"- User: 'check postgres status'\n" +
		"  YOU: run_command({'command': 'pg_isready -h localhost'})\n\n" +
		"**CRITICAL**: ACT immediately. Show results. Be concise.\n"

	cwd, _ := os.Getwd()
	toolsRaw := tools.GetAvailableTools()
	var availableTools []interface{}
	for _, t := range toolsRaw {
		availableTools = append(availableTools, t)
	}

	messages := []llm.ChatMessage{
		{Role: "user", Content: task},
	}

	maxIterations := 15

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
			fmt.Println(strings.TrimSpace(resp.Text))
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

	return nil
}
