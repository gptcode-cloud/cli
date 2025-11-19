package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"chuchu/internal/catalog"
	"chuchu/internal/config"
	"chuchu/internal/elixir"
	"chuchu/internal/llm"
	"chuchu/internal/memory"
	"chuchu/internal/modes"
	"chuchu/internal/prompt"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "chu",
	Short: "Chuchu – strict TDD-first coding companion",
	Long: `Chuchu – strict TDD-first coding companion

No bullshit, no giant blobs of code, no skipping tests.

General execution:
  chu run "task"           - Execute tasks: HTTP requests, CLI commands, devops

Workflow modes (research → plan → implement):
  chu research "question"  - Document codebase and understand architecture
  chu plan "task"          - Create detailed implementation plan
  chu implement plan.md    - Execute plan phase-by-phase with verification

Interactive modes:
  chu chat                 - Code-focused conversation (use from CLI or Neovim)
  chu tdd                  - TDD mode (tests → implementation → refine)

Language-specific TDD:
  chu elixir               - Generate ExUnit tests + implementation for Elixir
  chu ts                   - Generate tests + implementation for TypeScript

Setup:
  chu setup                - Initialize ~/.chuchu configuration
  chu key [backend]        - Add/update API key for backend
  chu models update        - Update model catalog from OpenRouter API`,
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(keyCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(tddCmd)
	rootCmd.AddCommand(researchCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(implementCmd)
	rootCmd.AddCommand(elixirCmd)
	rootCmd.AddCommand(tsCmd)
}

func newBuilderAndLLM(lang, mode, hint string) (*prompt.Builder, llm.Provider, string, error) {
	setup, err := config.LoadSetup()
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to load setup: %w", err)
	}

	store, _ := memory.LoadStore()
	builder := prompt.NewDefaultBuilder(store)

	backendName := setup.Defaults.Backend
	modelAlias := setup.Defaults.Model

	backendCfg := setup.Backend[backendName]
	model := backendCfg.DefaultModel
	if alias, ok := backendCfg.Models[modelAlias]; ok {
		model = alias
	} else if modelAlias != "" {
		model = modelAlias
	}

	var provider llm.Provider
	if strings.Contains(model, "compound") {
		var customExec llm.Provider
		customModel := backendCfg.DefaultModel
		
		if backendCfg.Type == "ollama" {
			customExec = llm.NewOllama(backendCfg.BaseURL)
		} else {
			customExec = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
		}
		
		provider = llm.NewOrchestrator(backendCfg.BaseURL, backendName, customExec, customModel)
	} else {
		if backendCfg.Type == "ollama" {
			provider = llm.NewOllama(backendCfg.BaseURL)
		} else {
			provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
		}
	}

	return builder, provider, model, nil
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initialize ~/.chuchu with default profile and system prompt",
	Run: func(cmd *cobra.Command, args []string) {
		config.RunSetup()
	},
}

var keyCmd = &cobra.Command{
	Use:   "key [backend]",
	Short: "Add or update API key for a backend (e.g., chu key openrouter)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backendName := args[0]
		return config.UpdateAPIKey(backendName)
	},
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Manage model catalog",
}

var modelsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update model catalog from multiple sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Fetching models from available sources...")
		
		apiKeys := map[string]string{
			"groq":      os.Getenv("GROQ_API_KEY"),
			"openai":    os.Getenv("OPENAI_API_KEY"),
			"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
			"cohere":    os.Getenv("COHERE_API_KEY"),
		}
		
		catalogPath := catalog.GetCatalogPath()
		if err := catalog.FetchAndSave(catalogPath, apiKeys); err != nil {
			return fmt.Errorf("failed to update catalog: %w", err)
		}
		fmt.Printf("✓ Model catalog updated: %s\n", catalogPath)
		return nil
	},
}

func init() {
	modelsCmd.AddCommand(modelsUpdateCmd)
}

var chatCmd = &cobra.Command{
	Use:   "chat [message] [lang] [backend] [model]",
	Short: "Chat mode (code-focused conversation)",
	Run: func(cmd *cobra.Command, args []string) {
		var input string
		if len(args) > 0 && args[0] != "" {
			input = args[0]
			args = args[1:]
		} else {
			stdinBytes, _ := io.ReadAll(os.Stdin)
			input = string(stdinBytes)
		}
		modes.Chat(input, args)
	},
}

var tddCmd = &cobra.Command{
	Use:   "tdd",
	Short: "TDD mode (tests → implementation → refine)",
	RunE: func(cmd *cobra.Command, args []string) error {
		builder, provider, model, err := newBuilderAndLLM("general", "tdd", "")
		if err != nil {
			return err
		}
		return modes.RunTDD(builder, provider, model)
	},
}

var researchCmd = &cobra.Command{
	Use:   "research [question]",
	Short: "Research mode - document codebase and understand architecture",
	Long: `Research mode uses subagents to explore the codebase and document findings.
Provide a research question or area to investigate.

Example: chu research "How does authentication work?"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return modes.RunResearch(args)
	},
}

var planCmd = &cobra.Command{
	Use:   "plan [task]",
	Short: "Plan mode - create detailed implementation plan with phases",
	Long: `Plan mode creates a detailed implementation plan through interactive research.
Provide a task description or path to a ticket/spec file.

Example: chu plan "Add user authentication"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return modes.RunPlan(args)
	},
}

var implementCmd = &cobra.Command{
	Use:   "implement <plan_file>",
	Short: "Implement mode - execute plan with verification at each phase",
	Long: `Implement mode executes an approved plan from ~/.chuchu/plans/ directory.
Each phase is implemented and verified before proceeding.

Example: chu implement ~/.chuchu/plans/2025-01-15-add-auth.md`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return modes.RunImplement(args[0])
	},
}

var runCmd = &cobra.Command{
	Use:   "run [task]",
	Short: "Execute general tasks: HTTP requests, CLI commands, devops actions",
	Long: `Execute mode for general operational tasks without TDD ceremony.

Perfect for:
- HTTP requests (curl, API calls)
- CLI tools (fly, docker, kubectl, gh)
- DevOps tasks (deployments, infrastructure checks)
- Any command execution or task automation

Examples:
  chu run "make a GET request to https://api.github.com/users/octocat"
  chu run "deploy to staging using fly deploy"
  chu run "check if postgres is running"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		builder, provider, model, err := newBuilderAndLLM("general", "run", "")
		if err != nil {
			return err
		}
		return modes.RunExecute(builder, provider, model, args)
	},
}

var elixirCmd = &cobra.Command{
	Use:   "elixir",
	Short: "Generate ExUnit tests + implementation for an Elixir feature",
	RunE: func(cmd *cobra.Command, args []string) error {
		builder, provider, model, err := newBuilderAndLLM("elixir", "tdd", "")
		if err != nil {
			return err
		}
		return elixir.RunFeatureElixir(builder, provider, model)
	},
}

var tsCmd = &cobra.Command{
	Use:   "ts [feature]",
	Short: "Generate tests + implementation for a TypeScript feature",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		builder, provider, model, err := newBuilderAndLLM("typescript", "tdd", args[0])
		if err != nil {
			return err
		}
		return modes.RunTDD(builder, provider, model)
	},
}
