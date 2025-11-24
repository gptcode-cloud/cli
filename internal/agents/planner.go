package agents

import (
	"context"
	"fmt"

	"chuchu/internal/llm"
)

type PlannerAgent struct {
	provider llm.Provider
	model    string
}

func NewPlanner(provider llm.Provider, model string) *PlannerAgent {
	return &PlannerAgent{
		provider: provider,
		model:    model,
	}
}

const plannerPrompt = `You are a minimal planner. Your ONLY job is to create focused, minimal plans.

WORKFLOW:
1. Read the task and analysis
2. List ONLY files that need changes
3. Describe ONLY necessary changes

CRITICAL RULES:
- Be EXTREMELY minimal - solve the task in the SIMPLEST way possible
- NO helper scripts (Python, bash, etc) unless explicitly requested
- NO automation scripts
- NO tests unless explicitly requested
- NO documentation unless explicitly requested
- ONLY modify existing files OR files explicitly requested in the task
- If the task asks to "create a file with content X", create THAT FILE DIRECTLY, do NOT create a script to generate it
- Focus on the EXACT task, nothing more

BAD EXAMPLE:
Task: "Create summary.md with file list"
Bad Plan: Create generate_summary.py script that generates summary.md
Good Plan: Create summary.md directly with the file list

Create minimal, direct plans.`

func (p *PlannerAgent) CreatePlan(ctx context.Context, task string, analysis string, statusCallback StatusCallback) (string, error) {
	if statusCallback != nil {
		statusCallback("Planner: Creating minimal plan...")
	}

	planPrompt := fmt.Sprintf(`Create a MINIMAL implementation plan.

Task: %s

Codebase Analysis:
---
%s
---

Create a brief plan:
# Plan

## Files to modify
[List ONLY files that exist OR are explicitly requested in the task]

## Changes
[Describe ONLY the minimal, direct changes needed]
[If task asks to create file with content, create THAT file, NOT a script]

## Success Criteria
[How to verify it worked]

REMEMBER:
- NO scripts unless explicitly requested
- NO automation unless explicitly requested
- Solve the task DIRECTLY in the simplest way
- Keep it MINIMAL. NO extra features.`, task, analysis)

	resp, err := p.provider.Chat(ctx, llm.ChatRequest{
		SystemPrompt: plannerPrompt,
		UserPrompt:   planPrompt,
		Model:        p.model,
	})
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}
