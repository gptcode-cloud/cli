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
- NO intermediate/temporary files for data storage - use shell commands to get data
- ONLY modify existing files OR files explicitly requested in the task
- If the task asks to "create a file with content X", create THAT FILE DIRECTLY, do NOT create a script to generate it
- If task is about "retrieving" or "getting" data, use shell commands, NOT file creation
- Focus on the EXACT task, nothing more

EXAMPLE 1 - Adding authentication:
Task: "add user authentication"

Plan:
# Implementation Plan

## Files to Modify
- auth/handler.go (add Login, Logout functions)
- server.go (register auth routes)
- middleware/auth.go (create JWT middleware)

## Changes
1. auth/handler.go:
   - Add Login(w http.ResponseWriter, r *http.Request)
   - Add Logout(w http.ResponseWriter, r *http.Request)
   - Use bcrypt for password hashing

2. server.go:
   - Register POST /login and /logout routes
   - Add auth middleware to protected routes

3. middleware/auth.go:
   - Create JWT verification middleware
   - Parse and validate tokens from Authorization header

## Success Criteria
- Tests pass: go test ./auth/...
- Can login with valid credentials
- Protected routes return 401 without valid token
- Lints clean: golangci-lint run

EXAMPLE 2 - Direct file creation (NO scripts):
Task: "Create summary.md with project file list"

BAD Plan:
  Create generate_summary.py that scans files and writes summary.md

GOOD Plan:
# Plan

## Files to Create
- summary.md

## Changes
Create summary.md with:
- List of all Go files
- Brief description of each
- Project structure overview

## Success Criteria
- File summary.md exists
- Contains complete file list
- Markdown formatting is valid

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
[2-5 specific, testable criteria to verify the changes work]
[Examples: "Tests pass: make test", "File X contains Y", "Command Z produces output W"]

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
