package maestro

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gptcode/internal/agents"
	"gptcode/internal/config"
	"gptcode/internal/feedback"
	"gptcode/internal/llm"
)

// Conductor is the central coordinator (Maestro) that orchestrates all agents
type Conductor struct {
	selector *config.ModelSelector
	setup    *config.Setup
	cwd      string
	language string
}

// NewConductor creates a new Maestro conductor
func NewConductor(
	selector *config.ModelSelector,
	setup *config.Setup,
	cwd string,
	language string,
) *Conductor {
	return &Conductor{
		selector: selector,
		setup:    setup,
		cwd:      cwd,
		language: language,
	}
}

// ExecuteTask orchestrates the execution of a task
func (c *Conductor) ExecuteTask(ctx context.Context, task string, complexity string) error {
	if os.Getenv("GPTCODE_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[MAESTRO] ExecuteTask called: task=%s complexity=%s lang=%s\n", task, complexity, c.language)
	}

	// Select model for planning
	planBackend, planModel, err := c.selector.SelectModel(config.ActionPlan, c.language, complexity)
	if err != nil {
		return fmt.Errorf("failed to select planner model: %w", err)
	}

	if os.Getenv("GPTCODE_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[MAESTRO] Planner: %s/%s\n", planBackend, planModel)
	}

	// Create planner with selected model
	planProvider := c.createProvider(planBackend)
	planner := agents.NewPlanner(planProvider, planModel)

	fmt.Println("Creating plan...")
	plan, err := planner.CreatePlan(ctx, task, "", nil)
	c.selector.RecordUsage(planBackend, planModel, err == nil, errorMsg(err))
	if err != nil {
		return fmt.Errorf("planning failed: %w", err)
	}

	// Build conversation history
	history := []llm.ChatMessage{
		{Role: "user", Content: plan},
	}

	// Higher retry limit for autonomous error fixing loops
	// For syntax errors, the agent needs multiple attempts to:
	// 1. Fix initial error
	// 2. Run build to discover new errors
	// 3. Fix cascading errors
	// 4. Verify final solution
	maxAttempts := 10
	if os.Getenv("GPTCODE_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[MAESTRO] maxAttempts = %d\n", maxAttempts)
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			fmt.Printf("Retrying (attempt %d/%d)...\n", attempt, maxAttempts)
		}

		// Select model for editing
		if os.Getenv("GPTCODE_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[MAESTRO] About to select editor model for lang=%s complexity=%s\n", c.language, complexity)
		}
		editBackend, editModel, err := c.selector.SelectModel(config.ActionEdit, c.language, complexity)
		if err != nil {
			if os.Getenv("GPTCODE_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[MAESTRO] SelectModel failed: %v\n", err)
			}
			return fmt.Errorf("failed to select editor model: %w", err)
		}

		if os.Getenv("GPTCODE_DEBUG") == "1" && attempt == 1 {
			fmt.Fprintf(os.Stderr, "[MAESTRO] Editor: %s/%s\n", editBackend, editModel)
		}

		// Create editor with selected model
		editProvider := c.createProvider(editBackend)
		editor := agents.NewEditor(editProvider, c.cwd, editModel)

		// Execute with editor
		fmt.Println("Executing changes...")
		result, modifiedFiles, err := editor.Execute(ctx, history, nil)
		c.selector.RecordUsage(editBackend, editModel, err == nil, errorMsg(err))
		if err != nil {
			if attempt < maxAttempts {
				fmt.Printf("[WARNING] Execution error: %v\n", err)
				feedback := c.formatExecutionError(err)
				history = append(history, llm.ChatMessage{
					Role:    "user",
					Content: feedback,
				})
				continue
			}
			return fmt.Errorf("execution failed: %w", err)
		}

		// Check if this is a query-only task (no validation needed)
		if c.isQueryTask(plan, modifiedFiles) {
			c.recordFeedback(editBackend, editModel, "editor", task, true)

			fmt.Printf("\n[OK] Task complete!\n")
			fmt.Printf("   Modified: %d files\n", len(modifiedFiles))
			if result != "" {
				fmt.Printf("   %s\n", result)
			}
			return nil
		}

		// Select model for review
		reviewBackend, reviewModel, err := c.selector.SelectModel(config.ActionReview, c.language, complexity)
		if err != nil {
			return fmt.Errorf("failed to select reviewer model: %w", err)
		}

		if os.Getenv("GPTCODE_DEBUG") == "1" && attempt == 1 {
			fmt.Fprintf(os.Stderr, "[MAESTRO] Reviewer: %s/%s\n", reviewBackend, reviewModel)
		}

		// Create reviewer with selected model
		reviewProvider := c.createProvider(reviewBackend)
		reviewer := agents.NewReviewer(reviewProvider, c.cwd, reviewModel)

		// Validate
		fmt.Println("Validating...")
		review, err := reviewer.Review(ctx, plan, modifiedFiles, nil)
		c.selector.RecordUsage(reviewBackend, reviewModel, err == nil, errorMsg(err))
		if err != nil {
			if attempt < maxAttempts {
				fmt.Printf("[WARNING] Validation error: %v\n", err)
				feedback := c.formatValidationError(err)
				history = append(history, llm.ChatMessage{
					Role:    "user",
					Content: feedback,
				})
				continue
			}
			return fmt.Errorf("review failed: %w", err)
		}

		if !review.Success {
			if attempt < maxAttempts {
				issuesStr := strings.Join(review.Issues, "\n")
				fmt.Printf("[WARNING] Validation failed:\n%s\n", issuesStr)
				feedback := c.formatValidationIssues(review.Issues)
				history = append(history, llm.ChatMessage{
					Role:    "user",
					Content: feedback,
				})
				continue
			}
			return fmt.Errorf("review failed after %d attempts: %v", maxAttempts, review.Issues)
		}

		// Success! Record positive feedback
		c.recordFeedback(editBackend, editModel, "editor", task, true)
		c.recordFeedback(reviewBackend, reviewModel, "reviewer", task, true)

		fmt.Printf("\n[OK] Task complete!\n")
		fmt.Printf("   Modified: %d files\n", len(modifiedFiles))
		if result != "" {
			fmt.Printf("   %s\n", result)
		}
		return nil
	}

	// Task failed after all attempts - record negative feedback
	editBackend, editModel, _ := c.selector.SelectModel(config.ActionEdit, c.language, complexity)
	reviewBackend, reviewModel, _ := c.selector.SelectModel(config.ActionReview, c.language, complexity)
	c.recordFeedback(editBackend, editModel, "editor", task, false)
	c.recordFeedback(reviewBackend, reviewModel, "reviewer", task, false)

	return fmt.Errorf("task failed after %d attempts", maxAttempts)
}

func errorMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (c *Conductor) recordFeedback(backend, model, agent, task string, success bool) {
	sentiment := feedback.SentimentBad
	if success {
		sentiment = feedback.SentimentGood
	}

	event := feedback.Event{
		Sentiment: sentiment,
		Backend:   backend,
		Model:     model,
		Agent:     agent,
		Task:      task,
		Context:   fmt.Sprintf("language=%s", c.language),
	}

	if err := feedback.Record(event); err != nil {
		if os.Getenv("GPTCODE_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[WARN] Failed to record feedback: %v\n", err)
		}
	}
}

// createProvider creates an LLM provider for the given backend
func (c *Conductor) createProvider(backendName string) llm.Provider {
	backendCfg, ok := c.setup.Backend[backendName]
	if !ok {
		// Fallback to default
		backendName = c.setup.Defaults.Backend
		backendCfg = c.setup.Backend[backendName]
	}

	if backendCfg.Type == "ollama" {
		return llm.NewOllama(backendCfg.BaseURL)
	}
	return llm.NewChatCompletion(backendCfg.BaseURL, backendName)
}

// formatExecutionError creates clear feedback for execution errors
func (c *Conductor) formatExecutionError(err error) string {
	return fmt.Sprintf(`EXECUTION FAILED

Error: %v

INSTRUCTIONS:
1. Read the error message carefully
2. Identify what went wrong
3. Fix the specific issue mentioned
4. Try again

Be precise and fix only what's broken.`, err)
}

// formatValidationError creates clear feedback for review errors
func (c *Conductor) formatValidationError(err error) string {
	return fmt.Sprintf(`VALIDATION SYSTEM ERROR

Error: %v

INSTRUCTIONS:
The review process itself failed. This might mean:
- A required tool is not available
- Syntax errors prevent running build/test commands
- File read errors

Please ensure your code is syntactically correct and try again.`, err)
}

// formatValidationIssues creates clear feedback for review failures
func (c *Conductor) formatValidationIssues(issues []string) string {
	var sb strings.Builder
	sb.WriteString("VALIDATION FAILED\n\n")
	sb.WriteString("The following issues were found:\n")
	for i, issue := range issues {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
	}
	sb.WriteString("\nINSTRUCTIONS:\n")
	sb.WriteString("1. Fix each issue listed above\n")
	sb.WriteString("2. Pay special attention to:\n")

	// Check if there's a Go package mismatch error
	hasPackageError := false
	for _, issue := range issues {
		if strings.Contains(issue, "found packages") {
			hasPackageError = true
			break
		}
	}

	if hasPackageError {
		sb.WriteString("   - **CRITICAL**: Package name mismatch! Read ALL .go files in the directory to see the correct package name, then fix ONLY the wrong file(s)\n")
	}
	sb.WriteString("   - Correct package names (all .go files in same directory must have same package)\n")
	sb.WriteString("   - Missing imports\n")
	sb.WriteString("   - Compilation errors\n")
	sb.WriteString("3. Do NOT change what's already correct\n")
	sb.WriteString("4. Only fix the specific problems mentioned\n")
	return sb.String()
}

// isQueryTask checks if task is query-only (no validation needed)
func (c *Conductor) isQueryTask(plan string, modifiedFiles []string) bool {
	if len(modifiedFiles) > 0 {
		return false
	}

	lower := strings.ToLower(plan)
	queryIndicators := []string{
		"run command",
		"execute command",
		"git status",
		"git log",
		"gh pr list",
		"read file",
		"show",
		"display",
		"explain",
		"what is",
		"what does",
		"what means",
		"tell me about",
		"describe",
		"files to modify\nnone",
		"files to modify: none",
		"files to create\nnone",
		"files to create: none",
		"no files to modify",
		"no files to create",
	}

	for _, indicator := range queryIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}

	return false
}
