package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"chuchu/internal/config"
	"chuchu/internal/intelligence"
	"chuchu/internal/llm"
	"chuchu/internal/modes"
)

var doCmd = &cobra.Command{
	Use:   "do [task]",
	Short: "Autonomously execute vague or multi-step tasks with auto-recovery",
	Long: `Execute tasks autonomously with intelligent model selection and auto-recovery.

The 'do' command learns from past executions and automatically retries with better models when failures occur.

Examples:
  chu do "create a hello.txt file with Hello World"
  chu do "read docs/README.md and create a getting-started guide"
  chu do "unify all feature files in /guides"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task := strings.Join(args, " ")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")
		maxAttempts, _ := cmd.Flags().GetInt("max-attempts")

		if verbose {
			fmt.Fprintf(os.Stderr, "Task: %s\n", task)
			fmt.Fprintf(os.Stderr, "Dry-run: %v\n", dryRun)
			fmt.Fprintf(os.Stderr, "Max attempts: %d\n\n", maxAttempts)
		}

		if dryRun {
			return runDoAnalysis(task, verbose)
		}

		return runDoExecutionWithRetry(task, verbose, maxAttempts)
	},
}

func init() {
	rootCmd.AddCommand(doCmd)

	doCmd.Flags().Bool("dry-run", false, "Show analysis and plan without executing")
	doCmd.Flags().BoolP("verbose", "v", false, "Show detailed progress")
	doCmd.Flags().Int("max-attempts", 3, "Maximum retry attempts with different models")
}

func runDoAnalysis(task string, verbose bool) error {
	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load setup: %w", err)
	}

	backendName := setup.Defaults.Backend
	backendCfg := setup.Backend[backendName]

	var provider llm.Provider
	if backendCfg.Type == "ollama" {
		provider = llm.NewOllama(backendCfg.BaseURL)
	} else {
		provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	}

	queryModel := backendCfg.GetModelForAgent("query")

	fmt.Println("=== Task Analysis ===")
	fmt.Printf("Task: %s\n\n", task)

	analysisPrompt := fmt.Sprintf(`Analyze this task and determine:
1. Primary intent (create, read, update, refactor, unify, etc.)
2. Files that need to be read (if any)
3. Files that will be created or modified
4. Key steps required
5. Estimated complexity (1-10)

Task: %s

Provide a brief analysis.`, task)

	resp, err := provider.Chat(context.Background(), llm.ChatRequest{
		SystemPrompt: "You analyze tasks to determine requirements and complexity.",
		UserPrompt:   analysisPrompt,
		Model:        queryModel,
	})
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	fmt.Println(resp.Text)
	fmt.Println("\n=== Execution Plan ===")
	fmt.Println("This would create a detailed plan and execute using guided mode.")
	fmt.Println("\nTo execute: run without --dry-run flag")

	return nil
}

func runDoExecutionWithRetry(task string, verbose bool, maxAttempts int) error {
	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load setup: %w", err)
	}

	currentBackend := setup.Defaults.Backend
	currentBackendCfg := setup.Backend[currentBackend]
	currentEditorModel := currentBackendCfg.GetModelForAgent("editor")

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 && verbose {
			fmt.Fprintf(os.Stderr, "\n=== Attempt %d/%d ===\n", attempt, maxAttempts)
		}

		startTime := time.Now()
		err := runDoExecution(task, verbose, setup, currentBackend, currentEditorModel)
		elapsed := time.Since(startTime).Milliseconds()

		if err == nil {
			_ = intelligence.RecordExecution(intelligence.TaskExecution{
				Task:      task,
				Backend:   currentBackend,
				Model:     currentEditorModel,
				Success:   true,
				LatencyMs: elapsed,
			})

			if verbose {
				fmt.Fprintf(os.Stderr, "\n✓ Task completed successfully\n")
			}
			return nil
		}

		_ = intelligence.RecordExecution(intelligence.TaskExecution{
			Task:    task,
			Backend: currentBackend,
			Model:   currentEditorModel,
			Success: false,
			Error:   err.Error(),
		})

		errMsg := err.Error()
		looksLikeToolError := strings.Contains(errMsg, "tool") || strings.Contains(errMsg, "function") ||
			strings.Contains(errMsg, "not available") || strings.Contains(errMsg, "not supported")

		if !looksLikeToolError {
			return fmt.Errorf("task failed: %w", err)
		}

		if attempt >= maxAttempts {
			return fmt.Errorf("task failed after %d attempts: %w", maxAttempts, err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "\n❌ Attempt %d failed: %v\n", attempt, err)
			fmt.Fprintf(os.Stderr, "🤔 Asking intelligence system for alternative model...\n")
		}

		recommendations, err := intelligence.RecommendModelForRetry(setup, "editor", currentBackend, currentEditorModel, task)
		if err != nil || len(recommendations) == 0 {
			return fmt.Errorf("no alternative models available: %w", err)
		}

		rec := recommendations[0]

		if verbose {
			fmt.Fprintf(os.Stderr, "\n💡 Intelligence recommends: %s/%s\n", rec.Backend, rec.Model)
			fmt.Fprintf(os.Stderr, "   Confidence: %.0f%%\n", rec.Confidence*100)
			fmt.Fprintf(os.Stderr, "   Reason: %s\n", rec.Reason)
			fmt.Fprintf(os.Stderr, "\n🔄 Retrying with recommended model...\n")
		}

		if rec.Backend != currentBackend {
			if _, exists := setup.Backend[rec.Backend]; exists {
				currentBackend = rec.Backend
				currentBackendCfg = setup.Backend[currentBackend]
			}
		}

		currentEditorModel = rec.Model
	}

	return fmt.Errorf("task failed after %d attempts", maxAttempts)
}

func runDoExecution(task string, verbose bool, setup *config.Setup, backendName string, editorModel string) error {
	backendCfg := setup.Backend[backendName]

	cwd, _ := os.Getwd()

	var provider llm.Provider
	if backendCfg.Type == "ollama" {
		provider = llm.NewOllama(backendCfg.BaseURL)
	} else {
		provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	}

	researchModel := backendCfg.GetModelForAgent("research")
	orchestrator := llm.NewOrchestrator(backendCfg.BaseURL, backendName, provider, researchModel)

	queryModel := backendCfg.GetModelForAgent("query")

	if verbose {
		fmt.Fprintf(os.Stderr, "Backend: %s\n", backendName)
		fmt.Fprintf(os.Stderr, "Editor Model: %s\n", editorModel)
		fmt.Fprintf(os.Stderr, "Query Model: %s\n\n", queryModel)
	}

	guided := modes.NewGuidedMode(orchestrator, cwd, queryModel)

	if verbose {
		fmt.Fprintf(os.Stderr, "Creating plan...\n")
	}

	if err := guided.Execute(context.Background(), task); err != nil {
		return fmt.Errorf("plan creation failed: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Plan created. Starting implementation...\n")
	}

	guidedWithCustomEditor := modes.NewGuidedModeWithCustomModel(orchestrator, cwd, queryModel, editorModel)

	if err := guidedWithCustomEditor.Implement(context.Background(), task); err != nil {
		return fmt.Errorf("implementation failed: %w", err)
	}

	return nil
}
