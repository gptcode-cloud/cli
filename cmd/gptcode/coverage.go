package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gptcode/internal/config"
	"gptcode/internal/coverage"
	"gptcode/internal/llm"
)

var coverageCmd = &cobra.Command{
	Use:   "coverage [package]",
	Short: "Analyze test coverage and identify gaps",
	Long: `Analyze test coverage for a package and identify functions that need tests.

Examples:
  gptcode coverage ./...           # Analyze all packages
  gptcode coverage ./internal/...  # Analyze internal packages
  gptcode coverage .               # Analyze current package`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCoverage,
}

var coverageModel string

func init() {
	rootCmd.AddCommand(coverageCmd)
	coverageCmd.Flags().StringVar(&coverageModel, "model", "", "LLM model to use (default: from config)")
}

func runCoverage(cmd *cobra.Command, args []string) error {
	pkgPath := "./..."
	if len(args) > 0 {
		pkgPath = args[0]
	}

	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, model, err := getCoverageProvider(setup)
	if err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	analyzer := coverage.NewCoverageAnalyzer(provider, model, workDir)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Printf("📊 Analyzing coverage for: %s\n\n", pkgPath)

	result, err := analyzer.Analyze(ctx, pkgPath)
	if err != nil {
		return fmt.Errorf("coverage analysis failed: %w", err)
	}

	fmt.Println(result.Report)
	fmt.Printf("\n📈 Total Coverage: %.1f%%\n", result.TotalCoverage)

	if len(result.Gaps) > 0 {
		fmt.Printf("⚠️  %d function(s) need attention\n", len(result.Gaps))
	}

	return nil
}

func getCoverageProvider(setup *config.Setup) (llm.Provider, string, error) {
	model := coverageModel
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
