package modes

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"chuchu/internal/agents"
	"chuchu/internal/events"
	"chuchu/internal/llm"
)

type GuidedMode struct {
	events   *events.Emitter
	provider llm.Provider
	cwd      string
	model    string
}

func NewGuidedMode(provider llm.Provider, cwd string, model string) *GuidedMode {
	return &GuidedMode{
		events:   events.NewEmitter(os.Stderr),
		provider: provider,
		cwd:      cwd,
		model:    model,
	}
}

func (g *GuidedMode) Execute(ctx context.Context, userMessage string) error {
	g.events.Message("🚀 Starting guided workflow...")
	g.events.Status("🔄 Analyzing task...")

	draftPlan, err := g.createDraftPlan(ctx, userMessage)
	if err != nil {
		return fmt.Errorf("failed to create draft: %w", err)
	}

	draftPath, err := g.saveDraft(draftPlan)
	if err != nil {
		return fmt.Errorf("failed to save draft: %w", err)
	}

	g.events.OpenPlan(draftPath)
	g.events.Message("📋 Draft plan created. Review in opened tab.")
	g.events.Message("✅ Guided workflow complete. To continue: chu plan → chu implement")
	
	return nil

	g.events.Status("📝 Creating detailed plan...")
	
	fullPlan, err := g.createDetailedPlan(ctx, userMessage, draftPlan)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	planPath, err := g.savePlan(fullPlan)
	if err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	g.events.OpenPlan(planPath)
	g.events.Message("✅ Plan ready. Review and confirm to implement.")

	if !g.waitForConfirmation("Proceed with implementation?", "plan_implement") {
		g.events.Message("❌ Implementation cancelled")
		return nil
	}

	g.events.Status("🔨 Implementing...")

	if err := g.implement(ctx, fullPlan); err != nil {
		return fmt.Errorf("implementation failed: %w", err)
	}

	g.events.Complete()
	g.events.Message("✅ Implementation complete. Review files and run tests.")

	return nil
}

func (g *GuidedMode) createDraftPlan(ctx context.Context, task string) (string, error) {
	prompt := fmt.Sprintf(`You are creating a DRAFT implementation plan.

Task: %s

Create a brief outline covering:
1. Overview (2-3 sentences)
2. Key steps (3-5 bullet points)
3. Files likely affected

Keep it concise - this is just a draft for user approval.`, task)

	resp, err := g.provider.Chat(ctx, llm.ChatRequest{
		SystemPrompt: "You create concise technical plans.",
		UserPrompt:   prompt,
		Model:        g.model,
	})
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}

func (g *GuidedMode) createDetailedPlan(ctx context.Context, task string, draft string) (string, error) {
	research := "No research available"
	
	if orchestrator, ok := g.provider.(*llm.OrchestratorProvider); ok {
		researchAgent := agents.NewResearch(orchestrator)
		
		statusCallback := func(status string) {
			g.events.Status(status)
		}

		history := []llm.ChatMessage{
			{Role: "user", Content: fmt.Sprintf("Research this codebase for: %s", task)},
		}

		if result, err := researchAgent.Execute(ctx, history, statusCallback); err == nil {
			research = result
		}
	}

	prompt := fmt.Sprintf(`Create a detailed implementation plan.

Task: %s

Draft outline:
%s

Codebase research:
%s

Create a structured plan with:
# Implementation Plan

## Problem Statement
[What we're solving]

## Current State
[Relevant architecture/files]

## Proposed Changes
### Phase 1: [Name]
- Step 1
- Step 2

### Phase 2: [Name]
- Step 1
- Step 2

## Files to Create/Modify
- path/to/file.ext: [purpose]

## Testing Strategy
[How to verify]`, task, draft, research)

	resp, err := g.provider.Chat(ctx, llm.ChatRequest{
		SystemPrompt: "You create detailed, actionable technical plans.",
		UserPrompt:   prompt,
		Model:        g.model,
	})
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}

func (g *GuidedMode) implement(ctx context.Context, plan string) error {
	editorAgent := agents.NewEditor(g.provider, g.cwd, g.model)

	statusCallback := func(status string) {
		g.events.Status(status)
	}

	history := []llm.ChatMessage{
		{Role: "user", Content: fmt.Sprintf("Implement this plan:\n\n%s", plan)},
	}

	_, err := editorAgent.Execute(ctx, history, statusCallback)
	return err
}

func (g *GuidedMode) saveDraft(content string) (string, error) {
	home, _ := os.UserHomeDir()
	dir := fmt.Sprintf("%s/.chuchu/plans", home)
	os.MkdirAll(dir, 0755)

	path := fmt.Sprintf("%s/draft.md", dir)
	return path, os.WriteFile(path, []byte(content), 0644)
}

func (g *GuidedMode) savePlan(content string) (string, error) {
	home, _ := os.UserHomeDir()
	dir := fmt.Sprintf("%s/.chuchu/plans", home)
	os.MkdirAll(dir, 0755)

	timestamp := time.Now().Format("2006-01-02-150405")
	path := fmt.Sprintf("%s/%s-plan.md", dir, timestamp)
	return path, os.WriteFile(path, []byte(content), 0644)
}

func (g *GuidedMode) waitForConfirmation(prompt string, id string) bool {
	if err := g.events.Confirm(prompt, id); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to emit confirm event: %v\n", err)
	}
	
	os.Stdout.Sync()
	time.Sleep(100 * time.Millisecond)
	
	responseChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	
	go func() {
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- response
	}()
	
	timeout := time.After(120 * time.Second)
	select {
	case response := <-responseChan:
		response = strings.TrimSpace(strings.ToLower(response))
		return response == "y" || response == "yes"
	case <-errorChan:
		return false
	case <-timeout:
		g.events.Notify("Timeout waiting for confirmation", "warn")
		return false
	}
}

func IsComplexTask(message string) bool {
	lower := strings.ToLower(message)

	complexKeywords := []string{
		"analyz",
		"analys",
		"integrat",
		"implement",
		"add",
		"creat",
		"build",
		"setup",
		"configur",
	}

	multiStepIndicators := []string{
		"first",
		"then",
		"after",
		"and then",
	}

	keywordCount := 0
	for _, kw := range complexKeywords {
		if strings.Contains(lower, kw) {
			keywordCount++
		}
	}

	for _, indicator := range multiStepIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}

	return keywordCount >= 2 || (keywordCount >= 1 && len(message) > 30)
}
