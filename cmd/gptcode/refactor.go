package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"gptcode/internal/compat"
	"gptcode/internal/config"
	"gptcode/internal/llm"
	"gptcode/internal/refactor"
	"github.com/spf13/cobra"
)

var refactorCmd = &cobra.Command{
	Use:   "refactor",
	Short: "Refactoring and code coordination tools",
	Long:  `Coordinate complex refactorings across multiple files.`,
}

var refactorAPICmd = &cobra.Command{
	Use:   "api",
	Short: "Coordinate API changes across routes, handlers and tests",
	Long: `Detect API routes and ensure handlers and tests are updated.

Examples:
  chu refactor api           # Analyze and coordinate all API changes`,
	RunE: runRefactorAPI,
}

var refactorSignatureCmd = &cobra.Command{
	Use:   "signature <function> <new-signature>",
	Short: "Change function signature and update all call sites",
	Long: `Refactor function signature across all files that use it.

Examples:
  chu refactor signature ProcessData "(ctx context.Context, data []byte) error"
  chu refactor signature handleRequest "(w http.ResponseWriter, r *http.Request, logger *log.Logger)"`,
	Args: cobra.ExactArgs(2),
	RunE: runRefactorSignature,
}

var refactorBreakingCmd = &cobra.Command{
	Use:   "breaking",
	Short: "Detect breaking changes and update all consumers",
	Long: `Analyze git diff for breaking API changes and coordinate updates.

Detects:
- Function signature changes
- Removed functions/types
- Type definition changes

Generates migration plan and updates all consuming code.

Examples:
  chu refactor breaking    # Analyze and update breaking changes`,
	RunE: runRefactorBreaking,
}

var refactorTypeCmd = &cobra.Command{
	Use:   "type <typename> <new-definition>",
	Short: "Refactor type definition and propagate changes",
	Long: `Change type definition and update all usages.

Examples:
  chu refactor type User "struct{ID int; Name string; Email string}"
  chu refactor type Config "map[string]interface{}"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRefactorType,
}

var refactorCompatCmd = &cobra.Command{
	Use:   "compat <old-api> <new-api> <version>",
	Short: "Add backward compatibility wrapper",
	Long: `Generate wrapper code for deprecated API.

Examples:
  chu refactor compat ProcessData ProcessDataV2 v2.0.0
  chu refactor compat GetUser GetUserByID v1.5.0`,
	Args: cobra.ExactArgs(3),
	RunE: runRefactorCompat,
}

var refactorModel string

func init() {
	rootCmd.AddCommand(refactorCmd)
	refactorCmd.AddCommand(refactorAPICmd)
	refactorCmd.AddCommand(refactorSignatureCmd)
	refactorCmd.AddCommand(refactorBreakingCmd)
	refactorCmd.AddCommand(refactorTypeCmd)
	refactorCmd.AddCommand(refactorCompatCmd)

	refactorCmd.PersistentFlags().StringVar(&refactorModel, "model", "", "LLM model to use (default: from config)")
}

func runRefactorAPI(cmd *cobra.Command, args []string) error {
	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, model, err := getRefactorProvider(setup)
	if err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	coordinator := refactor.NewAPICoordinator(provider, model, workDir)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("🔄 Analyzing API changes...")

	result, err := coordinator.CoordinateChanges(ctx)
	if err != nil {
		return fmt.Errorf("coordination failed: %w", err)
	}

	if len(result.Changes) == 0 {
		fmt.Println("✅ No API changes detected")
		return nil
	}

	fmt.Printf("\n📍 Found %d API endpoint(s):\n", len(result.Changes))
	for _, change := range result.Changes {
		fmt.Printf("  %s %s -> %s\n", change.Method, change.Path, change.Handler)
	}

	if len(result.UpdatedFiles) > 0 {
		fmt.Printf("\n📝 Updated %d file(s):\n", len(result.UpdatedFiles))
		for _, file := range result.UpdatedFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  %d error(s) occurred:\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	if result.Valid {
		fmt.Println("\n✅ API coordination complete")
	} else {
		fmt.Println("\n⚠️  API coordination completed with errors")
	}

	return nil
}

func getRefactorProvider(setup *config.Setup) (llm.Provider, string, error) {
	model := refactorModel
	backendName := setup.Defaults.Backend
	if backendName == "" {
		backendName = "anthropic"
	}

	backendCfg, ok := setup.Backend[backendName]
	if !ok {
		return nil, "", fmt.Errorf("backend %s not configured", backendName)
	}

	var provider llm.Provider
	if backendCfg.Type == "ollama" {
		provider = llm.NewOllama(backendCfg.BaseURL)
	} else {
		provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	}

	if model == "" {
		model = backendCfg.GetModelForAgent("editor")
		if model == "" {
			model = backendCfg.DefaultModel
		}
	}

	if model == "" {
		return nil, "", fmt.Errorf("no model configured")
	}

	return provider, model, nil
}

func runRefactorSignature(cmd *cobra.Command, args []string) error {
	funcName := args[0]
	newSig := args[1]

	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, model, err := getRefactorProvider(setup)
	if err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	refactor := refactor.NewSignatureRefactor(provider, model, workDir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fmt.Printf("🔄 Refactoring function %s...\n", funcName)

	result, err := refactor.RefactorSignature(ctx, funcName, newSig)
	if err != nil {
		return fmt.Errorf("refactoring failed: %w", err)
	}

	fmt.Printf("\n📍 Function signature changed:\n")
	fmt.Printf("  Old: %s\n", result.OldSignature)
	fmt.Printf("  New: func %s%s\n", funcName, newSig)

	if len(result.UpdatedFiles) > 0 {
		fmt.Printf("\n📝 Updated %d file(s):\n", len(result.UpdatedFiles))
		for _, file := range result.UpdatedFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  %d error(s) occurred:\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("refactoring completed with errors")
	}

	fmt.Println("\n✅ Signature refactoring complete")
	return nil
}

func runRefactorBreaking(cmd *cobra.Command, args []string) error {
	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, model, err := getRefactorProvider(setup)
	if err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	coordinator := refactor.NewBreakingCoordinator(provider, model, workDir)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	fmt.Println("🔍 Detecting breaking changes...")

	result, err := coordinator.DetectAndCoordinate(ctx)
	if err != nil {
		return fmt.Errorf("coordination failed: %w", err)
	}

	if len(result.Changes) == 0 {
		fmt.Println("✅ No breaking changes detected")
		return nil
	}

	fmt.Printf("\n⚠️  Found %d breaking change(s):\n", len(result.Changes))
	for i, change := range result.Changes {
		fmt.Printf("\n%d. %s\n", i+1, change.Description)
		fmt.Printf("   Type: %s\n", change.Type)
		fmt.Printf("   Symbol: %s.%s\n", change.Package, change.Symbol)
		fmt.Printf("   Old: %s\n", change.OldAPI)
		if change.NewAPI != "" {
			fmt.Printf("   New: %s\n", change.NewAPI)
		}

		key := fmt.Sprintf("%s.%s", change.Package, change.Symbol)
		if consumers, ok := result.Consumers[key]; ok && len(consumers) > 0 {
			fmt.Printf("   Affected: %d file(s)\n", len(consumers))
		}
	}

	if result.MigrationPlan != "" {
		fmt.Println("\n📝 Migration Plan:")
		fmt.Println(result.MigrationPlan)
	}

	if len(result.UpdatedFiles) > 0 {
		fmt.Printf("\n📝 Updated %d consumer file(s):\n", len(result.UpdatedFiles))
		for _, file := range result.UpdatedFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  %d error(s) occurred:\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("breaking change coordination completed with errors")
	}

	fmt.Println("\n✅ Breaking change coordination complete")
	fmt.Println("\n⚠️  Manual review recommended before committing")
	return nil
}

func runRefactorType(cmd *cobra.Command, args []string) error {
	typeName := args[0]
	newDef := ""
	if len(args) > 1 {
		newDef = args[1]
	}

	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, model, err := getRefactorProvider(setup)
	if err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	refactorer := refactor.NewTypeRefactor(provider, model, workDir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fmt.Printf("🔍 Analyzing type %s...\n", typeName)

	propagate := newDef != ""
	result, err := refactorer.RefactorType(ctx, typeName, newDef, propagate)
	if err != nil {
		return fmt.Errorf("type refactoring failed: %w", err)
	}

	fmt.Println("\n📊 Impact Analysis:")
	fmt.Println(result.ImpactReport)

	if propagate && len(result.UpdatedFiles) > 0 {
		fmt.Printf("\n📝 Updated %d file(s)\n", len(result.UpdatedFiles))
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  %d error(s):\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	return nil
}

func runRefactorCompat(cmd *cobra.Command, args []string) error {
	oldAPI := args[0]
	newAPI := args[1]
	version := args[2]

	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, model, err := getRefactorProvider(setup)
	if err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	manager := compat.NewCompatManager(provider, model, workDir)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Printf("🔄 Creating backward compatibility for %s...\n", oldAPI)

	reason := "API improvement"
	report, err := manager.AddDeprecation(ctx, oldAPI, newAPI, version, reason)
	if err != nil {
		return fmt.Errorf("failed to create compatibility layer: %w", err)
	}

	fmt.Println("\n✅ Generated:")
	fmt.Println("  - Wrapper code (compat/deprecated.go)")
	fmt.Println("  - Migration guide (MIGRATION.md)")
	fmt.Printf("  - Breaking change scheduled: %s\n", report.BreakingIn)

	if err := manager.SaveCompatibilityFiles(report); err != nil {
		return fmt.Errorf("failed to save files: %w", err)
	}

	fmt.Println("\n📝 Next steps:")
	fmt.Println("  1. Review generated files")
	fmt.Println("  2. Add deprecation notices to docs")
	fmt.Println("  3. Update CHANGELOG")

	return nil
}
