package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"chuchu/internal/config"
	"chuchu/internal/llm"
	"chuchu/internal/testgen"
	"github.com/spf13/cobra"
)

var testgenCmd = &cobra.Command{
	Use:   "testgen",
	Short: "Generate tests for code",
	Long:  `Generate unit tests for source files using LLM.`,
}

var testgenUnitCmd = &cobra.Command{
	Use:   "unit <file>",
	Short: "Generate unit tests for a source file",
	Args:  cobra.ExactArgs(1),
	RunE:  runTestgenUnit,
}

var testgenModel string

func init() {
	rootCmd.AddCommand(testgenCmd)
	testgenCmd.AddCommand(testgenUnitCmd)
	
	testgenCmd.PersistentFlags().StringVar(&testgenModel, "model", "", "LLM model to use (default: from config)")
}

func runTestgenUnit(cmd *cobra.Command, args []string) error {
	sourceFile := args[0]

	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, model, err := getTestgenProvider(setup)
	if err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	generator, err := testgen.NewTestGenerator(provider, model, workDir)
	if err != nil {
		return fmt.Errorf("failed to create test generator: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	fmt.Printf("🧪 Generating unit tests for: %s\n", sourceFile)
	
	result, err := generator.GenerateUnitTests(ctx, sourceFile)
	if err != nil && result == nil {
		return fmt.Errorf("failed to generate tests: %w", err)
	}

	if result.Valid {
		fmt.Printf("✅ Generated %s (valid)\n", result.TestFile)
	} else {
		fmt.Printf("⚠️  Generated %s (may have compilation issues)\n", result.TestFile)
		if result.Error != nil {
			fmt.Printf("   Error: %v\n", result.Error)
		}
	}

	return nil
}

func getTestgenProvider(setup *config.Setup) (llm.Provider, string, error) {
	model := testgenModel
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
		model = backendCfg.GetModelForAgent("query")
		if model == "" {
			model = backendCfg.DefaultModel
		}
	}
	
	if model == "" {
		return nil, "", fmt.Errorf("no model configured")
	}
	
	return provider, model, nil
}
