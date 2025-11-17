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

func RunPlan(builder *prompt.Builder, provider llm.Provider, model string, args []string) error {
	task := ""
	if len(args) > 0 {
		task = strings.Join(args, " ")
	}

	if task == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Fprintln(os.Stderr, "Plan mode - Create detailed implementation plan")
		fmt.Fprintln(os.Stderr, "\nWhat task would you like to plan?")
		fmt.Fprint(os.Stderr, "> ")
		if scanner.Scan() {
			task = scanner.Text()
		}
	}

	if task == "" {
		return fmt.Errorf("no task provided")
	}

	home, _ := os.UserHomeDir()
	plansDir := filepath.Join(home, ".chuchu", "plans")
	os.MkdirAll(plansDir, 0755)

	sys := builder.BuildSystemPrompt(prompt.BuildOptions{
		Lang: "general",
		Mode: "plan",
		Hint: task,
	})

	planPrompt := fmt.Sprintf(`You are creating a detailed implementation plan for this task:

%s

Process:
1. First, use tools to understand the current codebase:
   - Read relevant files (use read_file tool)
   - Search for similar implementations (use search_code tool)
   - Explore directory structures (use list_files tool)

2. After gathering context, create a structured plan with:

# [Task Name] Implementation Plan

## Overview
[Brief description of what we're implementing and why]

## Current State Analysis
[What exists now, what's missing, key constraints discovered]

## Desired End State
[Specification of the desired end state after this plan is complete, and how to verify it]

## Key Discoveries
- [Important finding with file:line reference]
- [Pattern to follow]
- [Constraint to work within]

## What We're NOT Doing
[Explicitly list out-of-scope items to prevent scope creep]

## Implementation Approach
[High-level strategy and reasoning]

## Phase 1: [Descriptive Name]

### Overview
[What this phase accomplishes]

### Changes Required

#### 1. [Component/File Group]
**File**: path/to/file.ext
**Changes**: [Summary of changes]

### Success Criteria

#### Automated Verification:
- [ ] Tests pass: make test
- [ ] Linting passes: make lint
- [ ] Build succeeds: make build

#### Manual Verification:
- [ ] Feature works as expected when tested
- [ ] No regressions in related features

**Implementation Note**: After completing this phase and all automated verification passes, pause for manual confirmation before proceeding to next phase.

---

[Repeat Phase structure for each major step]

## Testing Strategy
[What to test and how]

## References
[Links to research, similar code, etc.]

Begin by exploring the codebase, then present your findings and ask for feedback before writing the detailed plan.`, task)

	cwd, _ := os.Getwd()
	toolsRaw := tools.GetAvailableTools()
	var availableTools []interface{}
	for _, t := range toolsRaw {
		availableTools = append(availableTools, t)
	}

	messages := []llm.ChatMessage{
		{Role: "user", Content: planPrompt},
	}

	maxIterations := 20
	var fullPlan strings.Builder
	planFinalized := false

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
			
			if strings.Contains(strings.ToLower(resp.Text), "# ") && 
			   strings.Contains(strings.ToLower(resp.Text), "implementation plan") {
				fullPlan.WriteString(resp.Text)
				fullPlan.WriteString("\n\n")
			}
		}

		if len(resp.ToolCalls) == 0 {
			if !planFinalized {
				fmt.Fprintln(os.Stderr, "\n---")
				fmt.Fprintln(os.Stderr, "Continue refining the plan? (y/n or provide feedback)")
				fmt.Fprint(os.Stderr, "> ")
				
				if scanner.Scan() {
					feedback := strings.TrimSpace(scanner.Text())
					if feedback == "n" || feedback == "no" {
						planFinalized = true
						break
					}
					if feedback == "y" || feedback == "yes" || feedback == "" {
						messages = append(messages, llm.ChatMessage{
							Role: "assistant",
							Content: resp.Text,
						})
						messages = append(messages, llm.ChatMessage{
							Role: "user",
							Content: "Please continue refining the plan.",
						})
						continue
					}
					
					messages = append(messages, llm.ChatMessage{
						Role: "assistant",
						Content: resp.Text,
					})
					messages = append(messages, llm.ChatMessage{
						Role: "user",
						Content: feedback,
					})
					continue
				}
			}
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

	if fullPlan.Len() == 0 {
		fmt.Fprintln(os.Stderr, "\nNo plan was generated. Try providing more specific requirements.")
		return nil
	}

	timestamp := time.Now().Format("2006-01-02")
	sanitizedTask := strings.Map(func(r rune) rune {
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
	}, task)
	if len(sanitizedTask) > 50 {
		sanitizedTask = sanitizedTask[:50]
	}

	filename := fmt.Sprintf("%s_%s.md", timestamp, sanitizedTask)
	planPath := filepath.Join(plansDir, filename)

	err := os.WriteFile(planPath, []byte(fullPlan.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Could not save plan to %s: %v\n", planPath, err)
	} else {
		fmt.Fprintf(os.Stderr, "\n\n✓ Plan saved to: %s\n", planPath)
		fmt.Fprintf(os.Stderr, "\nTo implement this plan, run:\n  chu implement %s\n", planPath)
	}

	return nil
}
