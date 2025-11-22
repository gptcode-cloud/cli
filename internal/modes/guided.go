package modes

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"chuchu/internal/agents"
	"chuchu/internal/config"
	"chuchu/internal/events"
	"chuchu/internal/llm"
	"chuchu/internal/ml"
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
	_ = g.events.Status("Analyzing task...")

	draftPlan, err := g.createDraftPlan(ctx, userMessage)
	if err != nil {
		return fmt.Errorf("failed to create draft: %w", err)
	}

	draftPath, err := g.saveDraft(draftPlan)
	if err != nil {
		return fmt.Errorf("failed to save draft: %w", err)
	}

	_ = g.events.OpenPlan(draftPath)
	_ = g.events.Message("Draft plan created.")

	_ = g.events.Status("Creating detailed plan...")

	fullPlan, err := g.createDetailedPlan(ctx, userMessage, draftPlan)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	planPath, err := g.savePlan(fullPlan)
	if err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	_ = g.events.OpenPlan(planPath)
	_ = g.events.Message("Detailed plan ready. Send 'implement' to start implementation.")

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
			_ = g.events.Status(status)
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

func (g *GuidedMode) Implement(ctx context.Context, plan string) error {
	editorAgent := agents.NewEditor(g.provider, g.cwd, g.model)

	statusCallback := func(status string) {
		_ = g.events.Status(status)
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
	_ = os.MkdirAll(dir, 0755)

	path := fmt.Sprintf("%s/draft.md", dir)
	return path, os.WriteFile(path, []byte(content), 0644)
}

func (g *GuidedMode) savePlan(content string) (string, error) {
	home, _ := os.UserHomeDir()
	dir := fmt.Sprintf("%s/.chuchu/plans", home)
	_ = os.MkdirAll(dir, 0755)

	timestamp := time.Now().Format("2006-01-02-150405")
	path := fmt.Sprintf("%s/%s-plan.md", dir, timestamp)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	currentPath := fmt.Sprintf("%s/.chuchu/current_plan.txt", home)
	_ = os.WriteFile(currentPath, []byte(content), 0644)

	return path, nil
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
		_ = g.events.Notify("Timeout waiting for confirmation", "warn")
		return false
	}
}

func IsComplexTask(message string) bool {
	// Try to load ML model
	p, err := ml.LoadEmbedded("complexity_detection")
	if err != nil {
		return false
	}
	label, probs := p.Predict(message)
	if label == "multistep" {
		return true
	}
	if label == "complex" {
		threshold := 0.55
		if setup, err2 := config.LoadSetup(); err2 == nil {
			if setup.Defaults.MLComplexThreshold > 0 {
				threshold = setup.Defaults.MLComplexThreshold
			}
		}
		if v, ok := probs["complex"]; ok && v >= threshold {
			return true
		}
	}
	return false
}
