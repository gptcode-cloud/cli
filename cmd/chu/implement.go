package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"chuchu/internal/config"
	"chuchu/internal/llm"
	"chuchu/internal/maestro"

	"github.com/spf13/cobra"
)

var implementCmd = &cobra.Command{
	Use:   "implement <plan_file>",
	Short: "Execute an implementation plan",
	Long: `Execute an implementation plan step-by-step.

By default, prompts for confirmation before each step.
Use --auto for autonomous execution with automatic verification and retry.

Examples:
  chu implement plan.md
  chu implement plan.md --auto
  chu implement plan.md --auto --lint
  chu implement plan.md --auto --max-retries 5`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planPath := args[0]
		autoMode, _ := cmd.Flags().GetBool("auto")
		
		if autoMode {
			return runAutonomousImplement(cmd, planPath)
		}
		
		return runInteractiveImplement(planPath)
	},
}

func init() {
	implementCmd.Flags().Bool("auto", false, "Autonomous execution with verification and retry")
	implementCmd.Flags().Int("max-retries", 3, "Maximum retry attempts per step (only with --auto)")
	implementCmd.Flags().Bool("lint", false, "Enable lint verification (only with --auto)")
	implementCmd.Flags().Bool("resume", false, "Resume from last checkpoint (only with --auto)")
}

func runAutonomousImplement(cmd *cobra.Command, planPath string) error {
	maxRetries, _ := cmd.Flags().GetInt("max-retries")
	lint, _ := cmd.Flags().GetBool("lint")
	resume, _ := cmd.Flags().GetBool("resume")

	planContent, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("failed to read plan: %w", err)
	}

	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load setup: %w", err)
	}

	backendName := setup.Defaults.Backend
	backendCfg := setup.Backend[backendName]
	cwd, _ := os.Getwd()

	var provider llm.Provider
	if backendCfg.Type == "ollama" {
		provider = llm.NewOllama(backendCfg.BaseURL)
	} else {
		provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	}

	model := backendCfg.GetModelForAgent("editor")

	m := maestro.NewMaestro(provider, cwd, model)
	if maxRetries > 0 {
		m.MaxRetries = maxRetries
	}

	if lint {
		m.Verifiers = append(m.Verifiers, maestro.NewLintVerifier(cwd))
	}

	if resume {
		fmt.Fprintln(os.Stderr, "⚙  Attempting to resume from checkpoint...")
		if err := m.ResumeExecution(context.Background(), string(planContent)); err != nil {
			fmt.Fprintf(os.Stderr, "⚠  Resume failed, starting fresh: %v\n", err)
			resume = false
		}
	}

	if !resume {
		fmt.Fprintf(os.Stderr, "🚀 Starting autonomous execution of %s...\n", planPath)
		if err := m.ExecutePlan(context.Background(), string(planContent)); err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}
	}

	fmt.Fprintln(os.Stderr, "✓ Execution completed successfully!")
	return nil
}

func runInteractiveImplement(planPath string) error {
	planContent, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("failed to read plan: %w", err)
	}

	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load setup: %w", err)
	}

	backendName := setup.Defaults.Backend
	backendCfg := setup.Backend[backendName]
	cwd, _ := os.Getwd()

	var provider llm.Provider
	if backendCfg.Type == "ollama" {
		provider = llm.NewOllama(backendCfg.BaseURL)
	} else {
		provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	}

	model := backendCfg.GetModelForAgent("editor")

	m := maestro.NewMaestro(provider, cwd, model)
	steps := m.ParsePlan(string(planContent))

	fmt.Fprintf(os.Stderr, "📋 Plan loaded: %d steps\n\n", len(steps))

	reader := bufio.NewReader(os.Stdin)

	for i, step := range steps {
		fmt.Fprintf(os.Stderr, "\033[34m─── Step %d/%d: %s ───\033[0m\n", i+1, len(steps), step.Title)
		fmt.Fprintf(os.Stderr, "\n%s\n\n", strings.TrimSpace(step.Content))
		
		fmt.Fprint(os.Stderr, "Execute this step? [Y/n/q]: ")
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response == "q" || response == "quit" {
			fmt.Fprintln(os.Stderr, "\n⚠  Implementation cancelled")
			return nil
		}

		if response == "n" || response == "no" {
			fmt.Fprintln(os.Stderr, "⊘ Skipped")
			continue
		}

		if err := m.ExecuteStep(context.Background(), step); err != nil {
			fmt.Fprintf(os.Stderr, "\n\033[31m✗ Step failed: %v\033[0m\n", err)
			fmt.Fprint(os.Stderr, "Continue anyway? [y/N]: ")
			continueResp, _ := reader.ReadString('\n')
			continueResp = strings.ToLower(strings.TrimSpace(continueResp))
			if continueResp != "y" && continueResp != "yes" {
				return fmt.Errorf("implementation stopped at step %d", i+1)
			}
		} else {
			fmt.Fprintln(os.Stderr, "\033[32m✓ Step completed\033[0m")
		}
	}

	fmt.Fprintln(os.Stderr, "\n🎉 Implementation completed!")
	fmt.Fprintln(os.Stderr, "\nNext steps:")
	fmt.Fprintln(os.Stderr, "  • Review changes: git diff")
	fmt.Fprintln(os.Stderr, "  • Run tests")
	fmt.Fprintln(os.Stderr, "  • Commit changes")

	return nil
}
