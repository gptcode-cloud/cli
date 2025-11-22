package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"chuchu/internal/catalog"
	"chuchu/internal/config"
	"chuchu/internal/elixir"
	"chuchu/internal/feedback"
	"chuchu/internal/langdetect"
	"chuchu/internal/llm"
	"chuchu/internal/memory"
	"chuchu/internal/ml"
	"chuchu/internal/modes"
	"chuchu/internal/ollama"
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

Code quality:
  chu review [target]      - Review code for bugs, security, and improvements

Interactive modes:
  chu chat                 - Code-focused conversation (use from CLI or Neovim)
  chu tdd                  - TDD mode (tests → implementation → refine)

Feature generation:
  chu feature "desc"       - Generate tests + implementation (auto-detects language)

Setup:
  chu setup                - Initialize ~/.chuchu configuration
  chu key [backend]        - Add/update API key for backend
  chu backend list         - List configured backends
  chu backend create <name> <type> <base-url> - Create new backend
  chu backend delete <name> - Delete backend
  chu config get <key>     - Get configuration value
  chu config set <key> <value> - Set configuration value
  chu detect-language [path] - Detect project language
  chu models update        - Update model catalog from OpenRouter API
  chu models install <model> - Install Ollama model if not present
  chu profiles list <backend>              - List profiles for backend
  chu profiles show <backend> <profile>    - Show profile configuration
  chu profiles create <backend> <profile>  - Create new profile
  chu profiles set-agent <backend> <profile> <agent> <model>  - Set agent model
  chu feedback good|bad [--backend] [--model] [--agent] [--context] - Record feedback
  chu feedback stats               - View feedback statistics

Machine Learning:
  chu ml list                      - List available ML models
  chu ml train <model>             - Train ML model
  chu ml test <model> [query]      - Test ML model
  chu ml eval <model> [-f file]    - Evaluate model on dataset
  chu ml predict <text>            - Predict using embedded Go model`,
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(keyCmd)
	rootCmd.AddCommand(backendCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(detectLanguageCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(feedbackCmd)
	rootCmd.AddCommand(mlCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(tddCmd)
	rootCmd.AddCommand(researchCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(implementCmd)
	rootCmd.AddCommand(featureCmd)
	rootCmd.AddCommand(reviewCmd)
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

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Manage backends",
}

var backendListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured backends",
	RunE: func(cmd *cobra.Command, args []string) error {
		backends, err := config.ListBackends()
		if err != nil {
			return fmt.Errorf("failed to list backends: %w", err)
		}
		
		setup, _ := config.LoadSetup()
		defaultBackend := setup.Defaults.Backend
		
		for _, name := range backends {
			if name == defaultBackend {
				fmt.Printf("%s (default)\n", name)
			} else {
				fmt.Println(name)
			}
		}
		return nil
	},
}

var backendCreateCmd = &cobra.Command{
	Use:   "create <name> <type> <base-url>",
	Short: "Create new backend",
	Long: `Create a new backend configuration.

Type must be: openai, ollama

Examples:
  chu backend create mygroq openai https://api.groq.com/openai/v1
  chu backend create local ollama http://localhost:11434`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		backendType := args[1]
		baseURL := args[2]
		
		if backendType != "openai" && backendType != "ollama" {
			return fmt.Errorf("type must be 'openai' or 'ollama'")
		}
		
		if err := config.CreateBackend(name, backendType, baseURL); err != nil {
			return err
		}
		
		fmt.Printf("✓ Created backend: %s\n", name)
		fmt.Println("\nNext steps:")
		if backendType == "openai" {
			fmt.Printf("  chu key %s                    # Set API key\n", name)
		}
		fmt.Printf("  chu config set backend.%s.default_model <model>\n", name)
		fmt.Printf("  chu config set defaults.backend %s\n", name)
		return nil
	},
}

var backendDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete backend",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		
		if err := config.DeleteBackend(name); err != nil {
			return err
		}
		
		fmt.Printf("✓ Deleted backend: %s\n", name)
		return nil
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get configuration value",
	Long: `Get a configuration value from ~/.chuchu/setup.yaml

Examples:
  chu config get defaults.backend
  chu config get defaults.profile
  chu config get backend.groq.default_model`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value, err := config.GetConfig(key)
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Long: `Set a configuration value in ~/.chuchu/setup.yaml

Examples:
  chu config set defaults.backend groq
  chu config set defaults.profile speed
  chu config set backend.groq.default_model llama-3.3-70b-versatile`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]
		if err := config.SetConfig(key, value); err != nil {
			return err
		}
		fmt.Printf("✓ Set %s = %s\n", key, value)
		return nil
	},
}

var detectLanguageCmd = &cobra.Command{
	Use:   "detect-language [path]",
	Short: "Detect project language",
	Long: `Detect the primary programming language of a project.

Checks for language-specific files:
- Elixir: mix.exs
- Ruby: Gemfile, config/application.rb
- Go: go.mod
- TypeScript/JavaScript: tsconfig.json, package.json
- Python: requirements.txt, setup.py, pyproject.toml

If no marker files found, analyzes file extensions in directory.

Examples:
  chu detect-language
  chu detect-language /path/to/project
  chu detect-language .`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		lang := langdetect.DetectLanguage(path)
		fmt.Println(lang)
		return nil
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

var modelsSearchCmd = &cobra.Command{
	Use:   "search [term1] [term2] ...",
	Short: "Search models in catalog with filtering and sorting",
	Long: `Search models from a backend with multi-term filtering.
Models are automatically sorted by price (lowest first), then by context window (largest first).

Single term: filters across all backends
  chu models search gemini

Multiple terms: ANDs all terms (must match all)
  chu models search groq llama     # groq models with "llama" in name
  chu models search free coding     # free models tagged as coding

Flags override positional backend:
  chu models search gemini --backend openrouter`,
	RunE: func(cmd *cobra.Command, args []string) error {
		backendFlag, _ := cmd.Flags().GetString("backend")
		agentFlag, _ := cmd.Flags().GetString("agent")
		
		var queryTerms []string
		if len(args) > 0 {
			queryTerms = args
		}
		
		models, err := catalog.SearchModelsMulti(backendFlag, queryTerms, agentFlag)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}
		
		type ModelJSON struct {
			ID            string   `json:"id"`
			Name          string   `json:"name"`
			Tags          []string `json:"tags"`
			Recommended   bool     `json:"recommended"`
			ContextWindow int      `json:"context_window"`
			PricePrompt   float64  `json:"pricing_prompt_per_m_tokens"`
			PriceComp     float64  `json:"pricing_completion_per_m_tokens"`
			Installed     bool     `json:"installed"`
			FeedbackScore float64  `json:"feedback_score"`
		}
		
		result := make([]ModelJSON, 0, len(models))
		for _, m := range models {
			recommended := false
			if agentFlag != "" {
				for _, rec := range m.RecommendedFor {
					if rec == agentFlag {
						recommended = true
						break
					}
				}
			}
			
			result = append(result, ModelJSON{
				ID:            m.ID,
				Name:          m.Name,
				Tags:          m.Tags,
				Recommended:   recommended,
				ContextWindow: m.ContextWindow,
				PricePrompt:   m.PricingPrompt,
				PriceComp:     m.PricingComp,
				Installed:     m.Installed,
				FeedbackScore: m.FeedbackScore,
			})
		}
		
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
		
		return nil
	},
}

var modelsInstallCmd = &cobra.Command{
	Use:   "install <model>",
	Short: "Install Ollama model if not present",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]
		
		installed, err := ollama.IsInstalled(modelName)
		if err != nil {
			return fmt.Errorf("failed to check model status: %w", err)
		}
		
		if installed {
			fmt.Printf("✓ Model %s already installed\n", modelName)
			return nil
		}
		
		fmt.Printf("Installing model %s...\n", modelName)
		progressCallback := func(status string) {
			fmt.Println(status)
		}
		
		if err := ollama.Install(modelName, progressCallback); err != nil {
			return fmt.Errorf("failed to install model: %w", err)
		}
		
		fmt.Printf("✓ Model %s installed successfully\n", modelName)
		return nil
	},
}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage backend profiles",
}

var profilesListCmd = &cobra.Command{
	Use:   "list <backend>",
	Short: "List profiles for a backend",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend := args[0]
		profiles, err := config.ListBackendProfiles(backend)
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}
		
		encoder := json.NewEncoder(os.Stdout)
		if err := encoder.Encode(profiles); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
		return nil
	},
}

var profilesShowCmd = &cobra.Command{
	Use:   "show <backend> <profile>",
	Short: "Show profile configuration",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend := args[0]
		profileName := args[1]
		
		profile, err := config.GetBackendProfile(backend, profileName)
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}
		
		fmt.Printf("Profile: %s/%s\n", backend, profile.Name)
		if len(profile.AgentModels) == 0 {
			fmt.Println("  (no agent models configured)")
		} else {
			for agent, model := range profile.AgentModels {
				fmt.Printf("  %s: %s\n", agent, model)
			}
		}
		return nil
	},
}

var profilesCreateCmd = &cobra.Command{
	Use:   "create <backend> <profile-name>",
	Short: "Create new profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend := args[0]
		name := args[1]
		
		if err := config.CreateBackendProfile(backend, name); err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}
		
		fmt.Printf("✓ Created profile: %s/%s\n", backend, name)
		fmt.Println("\nConfigure agent models using:")
		fmt.Printf("  chu profiles set-agent %s %s router <model>\n", backend, name)
		fmt.Printf("  chu profiles set-agent %s %s query <model>\n", backend, name)
		fmt.Printf("  chu profiles set-agent %s %s editor <model>\n", backend, name)
		fmt.Printf("  chu profiles set-agent %s %s research <model>\n", backend, name)
		return nil
	},
}

var profilesSetAgentCmd = &cobra.Command{
	Use:   "set-agent <backend> <profile> <agent> <model>",
	Short: "Set agent model in profile",
	Long: `Set the model for a specific agent in a profile.

Agent types: router, query, editor, research

Example:
  chu profiles set-agent openrouter free router google/gemini-2.0-flash-exp:free`,
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend := args[0]
		profile := args[1]
		agent := args[2]
		model := args[3]
		
		if err := config.SetProfileAgentModel(backend, profile, agent, model); err != nil {
			return fmt.Errorf("failed to set agent model: %w", err)
		}
		
		fmt.Printf("✓ Set %s/%s %s = %s\n", backend, profile, agent, model)
		return nil
	},
}

var profilesDeleteCmd = &cobra.Command{
	Use:   "delete <backend> <profile>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend := args[0]
		profile := args[1]
		
		if err := config.DeleteBackendProfile(backend, profile); err != nil {
			return fmt.Errorf("failed to delete profile: %w", err)
		}
		
		fmt.Printf("✓ Deleted profile: %s/%s\n", backend, profile)
		return nil
	},
}

var feedbackCmd = &cobra.Command{
	Use:   "feedback",
	Short: "Record and analyze feedback for model performance",
}

var feedbackGoodCmd = &cobra.Command{
	Use:   "good",
	Short: "Record positive feedback",
	RunE: func(cmd *cobra.Command, args []string) error {
		backend, _ := cmd.Flags().GetString("backend")
		model, _ := cmd.Flags().GetString("model")
		agent, _ := cmd.Flags().GetString("agent")
		context, _ := cmd.Flags().GetString("context")
		
		event := feedback.Event{
			Sentiment: feedback.SentimentGood,
			Backend:   backend,
			Model:     model,
			Agent:     agent,
			Context:   context,
		}
		
		if err := feedback.Record(event); err != nil {
			return fmt.Errorf("failed to record feedback: %w", err)
		}
		
		fmt.Println("✓ Positive feedback recorded")
		return nil
	},
}

var feedbackBadCmd = &cobra.Command{
	Use:   "bad",
	Short: "Record negative feedback",
	RunE: func(cmd *cobra.Command, args []string) error {
		backend, _ := cmd.Flags().GetString("backend")
		model, _ := cmd.Flags().GetString("model")
		agent, _ := cmd.Flags().GetString("agent")
		context, _ := cmd.Flags().GetString("context")
		
		event := feedback.Event{
			Sentiment: feedback.SentimentBad,
			Backend:   backend,
			Model:     model,
			Agent:     agent,
			Context:   context,
		}
		
		if err := feedback.Record(event); err != nil {
			return fmt.Errorf("failed to record feedback: %w", err)
		}
		
		fmt.Println("✓ Negative feedback recorded")
		return nil
	},
}

var feedbackStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View feedback statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		events, err := feedback.LoadAll()
		if err != nil {
			return fmt.Errorf("failed to load feedback: %w", err)
		}
		
		if len(events) == 0 {
			fmt.Println("No feedback recorded yet")
			return nil
		}
		
		stats := feedback.Analyze(events)
		
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(stats); err != nil {
			return fmt.Errorf("failed to encode stats: %w", err)
		}
		
		return nil
	},
}

func init() {
	backendCmd.AddCommand(backendListCmd)
	backendCmd.AddCommand(backendCreateCmd)
	backendCmd.AddCommand(backendDeleteCmd)
	
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	
	rootCmd.AddCommand(profilesCmd)
	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesShowCmd)
	profilesCmd.AddCommand(profilesCreateCmd)
	profilesCmd.AddCommand(profilesSetAgentCmd)
	profilesCmd.AddCommand(profilesDeleteCmd)
	
	rootCmd.AddCommand(feedbackCmd)
	feedbackCmd.AddCommand(feedbackGoodCmd)
	feedbackCmd.AddCommand(feedbackBadCmd)
	feedbackCmd.AddCommand(feedbackStatsCmd)
	
	feedbackGoodCmd.Flags().String("backend", "", "Backend used")
	feedbackGoodCmd.Flags().String("model", "", "Model used")
	feedbackGoodCmd.Flags().String("agent", "", "Agent type (router, query, editor, research)")
	feedbackGoodCmd.Flags().String("context", "", "Additional context")
	
	feedbackBadCmd.Flags().String("backend", "", "Backend used")
	feedbackBadCmd.Flags().String("model", "", "Model used")
	feedbackBadCmd.Flags().String("agent", "", "Agent type (router, query, editor, research)")
	feedbackBadCmd.Flags().String("context", "", "Additional context")
	
	modelsCmd.AddCommand(modelsUpdateCmd)
	modelsCmd.AddCommand(modelsSearchCmd)
	modelsCmd.AddCommand(modelsInstallCmd)
	
	modelsSearchCmd.Flags().StringP("backend", "b", "openrouter", "Backend to search (openrouter, groq, ollama, etc)")
	modelsSearchCmd.Flags().StringP("agent", "a", "", "Agent type (router, query, editor, research)")
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
		desc := ""
		if len(args) > 0 {
			desc = args[0]
		}
		return modes.RunTDD(builder, provider, model, desc)
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

var featureCmd = &cobra.Command{
	Use:   "feature [description]",
	Short: "Generate tests + implementation for a feature (auto-detects language)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lang := detectLanguage()
		fmt.Printf("Detected language: %s\n", lang)

		builder, provider, model, err := newBuilderAndLLM(lang, "tdd", args[0])
		if err != nil {
			return err
		}

		if lang == "elixir" {
			return elixir.RunFeatureElixir(builder, provider, model)
		}
		
		// Default to generic TDD for other languages
		return modes.RunTDD(builder, provider, model, args[0])
	},
}

var mlCmd = &cobra.Command{
	Use:   "ml",
	Short: "Machine learning model management",
	Long: `Manage machine learning models for Chuchu.

Available commands:
  list  - List available models
  train - Train a model
  test  - Test a trained model

Examples:
  chu ml list
  chu ml train complexity_detection
  chu ml test complexity_detection "implement oauth2"`,
}

var mlListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available ML models",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		trainer := ml.NewTrainer(cwd)
		trainer.ListModels()
		return nil
	},
}

var mlTrainCmd = &cobra.Command{
	Use:   "train <model-name>",
	Short: "Train an ML model",
	Long: `Train a machine learning model.

Examples:
  chu ml train complexity_detection`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		
		trainer := ml.NewTrainer(cwd)
		modelName := args[0]
		
		if err := trainer.Train(modelName); err != nil {
			return fmt.Errorf("training failed: %w", err)
		}
		
		return nil
	},
}

var mlTestCmd = &cobra.Command{
	Use:   "test <model-name> [query]",
	Short: "Test a trained ML model",
	Long: `Test a trained model with example queries.

Without a query, runs pre-defined test examples.
With a query, tests that specific input.

Examples:
  chu ml test complexity_detection
  chu ml test complexity_detection "fix typo in readme"
  chu ml test complexity_detection "implement oauth2 with google"`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		
		trainer := ml.NewTrainer(cwd)
		modelName := args[0]
		
		var query string
		if len(args) > 1 {
			query = args[1]
		}
		
		if err := trainer.Test(modelName, query); err != nil {
			return fmt.Errorf("test failed: %w", err)
		}
		
		return nil
	},
}

var mlEvalCmd = &cobra.Command{
	Use:   "eval <model-name>",
	Short: "Evaluate a trained ML model on a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		trainer := ml.NewTrainer(cwd)
		file, _ := cmd.Flags().GetString("file")
		if err := trainer.Eval(args[0], file); err != nil {
			return fmt.Errorf("eval failed: %w", err)
		}
		return nil
	},
}

var mlPredictCmd = &cobra.Command{
	Use:   "predict [model-name] <text>",
	Short: "Predict using embedded Go model (no Python)",
	Long: `Predict using embedded Go model (no Python).

If model-name is omitted, defaults to 'complexity_detection'.

Examples:
  chu ml predict "fix typo in readme"
  chu ml predict complexity_detection "implement oauth"
  chu ml predict router_agent "explain this code"`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var modelName, text string
		
		if len(args) == 1 {
			modelName = "complexity_detection"
			text = args[0]
		} else {
			modelName = args[0]
			text = strings.Join(args[1:], " ")
		}
		
		p, err := ml.LoadEmbedded(modelName)
		if err != nil { return err }
		
		label, probs := p.Predict(text)
		fmt.Printf("Model: %s\n", modelName)
		fmt.Printf("Prediction: %s\n", label)
		fmt.Println("Probabilities:")
		
		for _, pair := range ml.SortedProbs(probs) {
			fmt.Printf("  %-12s %s\n", pair[0]+":", pair[1])
		}
		return nil
	},
}

func init() {
	mlCmd.AddCommand(mlListCmd)
	mlCmd.AddCommand(mlTrainCmd)
	mlCmd.AddCommand(mlTestCmd)
	mlCmd.AddCommand(mlEvalCmd)
	mlCmd.AddCommand(mlPredictCmd)
	mlEvalCmd.Flags().StringP("file", "f", "", "Path to eval CSV with columns: message,label")
}

var reviewCmd = &cobra.Command{
	Use:   "review [file or directory]",
	Short: "Review code for bugs, security issues, and improvements",
	Long: `Review mode performs detailed code analysis.

Review a specific file:
  chu review main.go
  chu review src/auth.go

Review a directory:
  chu review .
  chu review ./src

Focus on specific aspects:
  chu review main.go --focus security
  chu review . --focus performance
  chu review src/ --focus "error handling"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}

		focus, _ := cmd.Flags().GetString("focus")

		return modes.RunReview(modes.ReviewOptions{
			Target: target,
			Focus:  focus,
		})
	},
}

func init() {
	reviewCmd.Flags().StringP("focus", "f", "", "Focus area for review (e.g., security, performance, error handling)")
}

func detectLanguage() string {
	if _, err := os.Stat("mix.exs"); err == nil {
		return "elixir"
	}
	if _, err := os.Stat("Gemfile"); err == nil {
		return "ruby"
	}
	if _, err := os.Stat("go.mod"); err == nil {
		return "go"
	}
	if _, err := os.Stat("package.json"); err == nil {
		return "typescript"
	}
	if _, err := os.Stat("requirements.txt"); err == nil {
		return "python"
	}
	if _, err := os.Stat("Cargo.toml"); err == nil {
		return "rust"
	}
	return "unknown"
}
