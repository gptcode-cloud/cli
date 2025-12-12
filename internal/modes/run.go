package modes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gptcode/internal/llm"
	"gptcode/internal/prompt"
	"gptcode/internal/tools"
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

	sys += "\n\n## EXECUTION MODE - Autonomous Problem Solver\n\n" +
		"Your job: SOLVE the user's problem completely and autonomously.\n\n" +
		"### For TROUBLESHOOTING/DIAGNOSIS:\n" +
		"- Execute non-sudo diagnostic commands automatically\n" +
		"- NEVER use sudo commands (they block on password)\n" +
		"- For sudo-required info, present commands for manual execution\n" +
		"- Analyze results and drill down WITHOUT asking\n" +
		"- Identify root cause with available data\n" +
		"- Present final diagnosis with evidence\n" +
		"- Provide ready-to-execute commands (including sudo ones for manual run)\n\n" +
		"Example for 'high disk usage on macOS':\n" +
		"1. Run: df -h /\n" +
		"2. Run: du -h -d 3 ~/Library | sort -h | tail -n 20\n" +
		"3. Run: tmutil listlocalsnapshots /\n" +
		"4. Run: du -sh ~/Library/Developer/Xcode/* ~/Library/Caches/* 2>/dev/null\n" +
		"5. Check: docker system df (if docker available)\n" +
		"6. Present findings: 'Found: 450GB in Xcode DerivedData, 200GB in snapshots'\n" +
		"7. Provide commands:\n" +
		"   - Safe (no sudo): rm -rf ~/Library/Developer/Xcode/DerivedData/*\n" +
		"   - Requires sudo: sudo tmutil deletelocalsnapshots 2024-10-15-093045\n" +
		"   (Present sudo commands but explain they need manual execution)\n\n" +
		"### For SIMPLE EXECUTION:\n" +
		"- Execute immediately (non-sudo only)\n" +
		"- Show results\n" +
		"- Examples: HTTP requests, deployments, file operations\n\n" +
		"**CRITICAL**:\n" +
		"- NEVER execute sudo commands (they block)\n" +
		"- Present sudo commands as suggestions with explanation\n" +
		"- Be AUTONOMOUS with available tools\n"

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
			_ = json.Unmarshal([]byte(tc.Arguments), &args)

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
