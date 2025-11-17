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

func RunImplement(builder *prompt.Builder, provider llm.Provider, model string, planPath string) error {
	planContent, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("could not read plan file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Implementing plan from: %s\n\n", planPath)

	sys := builder.BuildSystemPrompt(prompt.BuildOptions{
		Lang: "general",
		Mode: "implement",
		Hint: "Implement plan phase by phase",
	})

	implementPrompt := fmt.Sprintf(`You are implementing an approved technical plan. Here is the complete plan:

---
%s
---

Your job:
1. Read the plan carefully and understand all phases
2. Check for any existing checkmarks (- [x]) to see what's already done
3. Start implementing from the first unchecked phase
4. For each phase:
   a. Read all relevant files mentioned in the phase
   b. Implement the changes described
   c. Run automated verification steps (tests, linting, etc.)
   d. Report completion and wait for manual verification confirmation
5. Update checkboxes in the plan as you complete items (use read_file and write to update the plan)

IMPORTANT:
- Follow the plan's intent while adapting to what you find
- Implement each phase fully before moving to next
- After completing automated verification for a phase, PAUSE and inform the user
- Do NOT proceed to next phase until user confirms manual verification is complete
- If something doesn't match the plan, explain the mismatch and ask for guidance

Begin by reading any files mentioned in Phase 1, then start implementing.`, string(planContent))

	cwd, _ := os.Getwd()
	toolsRaw := tools.GetAvailableTools()
	var availableTools []interface{}
	for _, t := range toolsRaw {
		availableTools = append(availableTools, t)
	}

	messages := []llm.ChatMessage{
		{Role: "user", Content: implementPrompt},
	}

	maxIterations := 30
	scanner := bufio.NewScanner(os.Stdin)

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

			if strings.Contains(strings.ToLower(resp.Text), "ready for manual verification") ||
				strings.Contains(strings.ToLower(resp.Text), "manual verification") ||
				strings.Contains(strings.ToLower(resp.Text), "please perform") {
				
				fmt.Fprintln(os.Stderr, "\n---")
				fmt.Fprintln(os.Stderr, "Manual verification required. Have you completed the manual testing?")
				fmt.Fprintln(os.Stderr, "(yes/no or provide feedback)")
				fmt.Fprint(os.Stderr, "> ")

				if scanner.Scan() {
					feedback := strings.TrimSpace(scanner.Text())
					if feedback == "yes" || feedback == "y" {
						messages = append(messages, llm.ChatMessage{
							Role:    "assistant",
							Content: resp.Text,
						})
						messages = append(messages, llm.ChatMessage{
							Role:    "user",
							Content: "Manual verification complete. Please proceed to the next phase.",
						})
						continue
					} else if feedback == "no" || feedback == "n" {
						messages = append(messages, llm.ChatMessage{
							Role:    "assistant",
							Content: resp.Text,
						})
						messages = append(messages, llm.ChatMessage{
							Role:    "user",
							Content: "Manual verification failed. Please investigate and fix the issues.",
						})
						continue
					} else {
						messages = append(messages, llm.ChatMessage{
							Role:    "assistant",
							Content: resp.Text,
						})
						messages = append(messages, llm.ChatMessage{
							Role:    "user",
							Content: feedback,
						})
						continue
					}
				}
			}
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

	fmt.Fprintln(os.Stderr, "\n✓ Implementation session complete")
	return nil
}
