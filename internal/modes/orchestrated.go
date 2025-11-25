package modes

import (
	"context"
	"fmt"
	"os"

	"chuchu/internal/agents"
	"chuchu/internal/events"
	"chuchu/internal/llm"
)

type OrchestratedMode struct {
	events       *events.Emitter
	provider     llm.Provider
	baseProvider llm.Provider
	cwd          string
	model        string
	editorModel  string
}

func NewOrchestratedMode(provider llm.Provider, baseProvider llm.Provider, cwd string, model string, editorModel string) *OrchestratedMode {
	return &OrchestratedMode{
		events:       events.NewEmitter(os.Stderr),
		provider:     provider,
		baseProvider: baseProvider,
		cwd:          cwd,
		model:        model,
		editorModel:  editorModel,
	}
}

func (o *OrchestratedMode) Execute(ctx context.Context, userMessage string) error {
	statusCallback := func(status string) {
		_ = o.events.Status(status)
	}

	_ = o.events.Status("Starting orchestrated execution...")

	analyzerAgent := agents.NewAnalyzer(o.baseProvider, o.cwd, o.model)
	analysis, err := analyzerAgent.Analyze(ctx, userMessage, statusCallback)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[ORCHESTRATED] Analysis: %s\n", analysis[:min(200, len(analysis))])
	}

	plannerAgent := agents.NewPlanner(o.provider, o.model)
	plan, err := plannerAgent.CreatePlan(ctx, userMessage, analysis, statusCallback)
	if err != nil {
		return fmt.Errorf("planning failed: %w", err)
	}

	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[ORCHESTRATED] Plan created\n")
	}

	_ = o.events.Message("Plan created. Executing implementation...")

	allowedFiles := extractFilesFromPlan(plan)

	var editorAgent *agents.EditorAgent
	if len(allowedFiles) > 0 {
		editorAgent = agents.NewEditorWithFileValidation(o.baseProvider, o.cwd, o.editorModel, allowedFiles)
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATED] Allowed files: %v\n", allowedFiles)
		}
	} else {
		editorAgent = agents.NewEditor(o.baseProvider, o.cwd, o.editorModel)
	}

	validatorAgent := agents.NewValidator(o.baseProvider, o.cwd, o.model)

	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATED] Implementation attempt %d/%d\n", attempt+1, maxRetries+1)
		}

		implementPrompt := fmt.Sprintf(`Implement EXACTLY what this plan says:

---
%s
---

ONLY modify files listed in the plan. ONLY make changes described. NO extras.`, plan)

		history := []llm.ChatMessage{
			{Role: "user", Content: implementPrompt},
		}

		result, modifiedFiles, err := editorAgent.Execute(ctx, history, statusCallback)
		if err != nil {
			return fmt.Errorf("implementation failed: %w", err)
		}

		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATED] Editor result: %s\n", result)
		}

		// Use actually modified files for validation if available
		filesToValidate := allowedFiles
		if len(modifiedFiles) > 0 {
			filesToValidate = modifiedFiles
		}

		validationResult, err := validatorAgent.Validate(ctx, plan, filesToValidate, statusCallback)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		if validationResult.Success {
			_ = o.events.Message("✓ Implementation validated successfully")
			return nil
		}

		if attempt < maxRetries {
			_ = o.events.Message(fmt.Sprintf("Validation failed. Retrying... (%d/%d)", attempt+2, maxRetries+1))
			for _, issue := range validationResult.Issues {
				_ = o.events.Message(fmt.Sprintf("  Issue: %s", issue))
			}

			editorAgent = agents.NewEditorWithFileValidation(o.baseProvider, o.cwd, o.editorModel, allowedFiles)
		} else {
			_ = o.events.Message("Implementation completed but validation failed.")
			for _, issue := range validationResult.Issues {
				_ = o.events.Message(fmt.Sprintf("  - %s", issue))
			}
			return fmt.Errorf("validation failed after max retries")
		}
	}

	return nil
}
