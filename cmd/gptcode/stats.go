package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"gptcode/internal/config"
)

var modeCmd = &cobra.Command{
	Use:   "mode [cloud|local]",
	Short: "Show or switch execution mode",
	Long: `Show current mode or switch between cloud and local execution.

Modes:
  cloud - Use cloud providers (OpenRouter, Groq, etc)
  local - Use local Ollama models only

Examples:
  gptcode mode            # Show current mode
  gptcode mode cloud      # Switch to cloud mode
  gptcode mode local      # Switch to local mode`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		setup, err := config.LoadSetup()
		if err != nil {
			return fmt.Errorf("failed to load setup: %w", err)
		}

		if len(args) == 0 {
			mode := setup.Defaults.Mode
			if mode == "" {
				mode = "cloud"
			}
			fmt.Printf("Current mode: %s\n", mode)
			return nil
		}

		newMode := args[0]
		if newMode != "cloud" && newMode != "local" {
			return fmt.Errorf("mode must be 'cloud' or 'local'")
		}

		if err := config.SetConfig("defaults.mode", newMode); err != nil {
			return fmt.Errorf("failed to set mode: %w", err)
		}

		fmt.Printf("✓ Switched to %s mode\n", newMode)
		return nil
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats [--today|--week|--all]",
	Short: "Display usage statistics with elegant dashboard",
	Long: `Display usage statistics in a beautiful dashboard format.

Flags:
  --today  Show today's stats only
  --week   Show last 7 days
  --all    Show all time stats (default)

Examples:
  gptcode stats
  gptcode stats --today
  gptcode stats --week`,
	RunE: func(cmd *cobra.Command, args []string) error {
		today, _ := cmd.Flags().GetBool("today")
		week, _ := cmd.Flags().GetBool("week")

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		usagePath := filepath.Join(home, ".gptcode", "usage.json")
		data, err := os.ReadFile(usagePath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No usage data yet. Start using gptcode to see stats!")
				return nil
			}
			return err
		}

		var usage map[string]map[string]struct {
			Requests     int    `json:"requests"`
			InputTokens  int    `json:"input_tokens"`
			OutputTokens int    `json:"output_tokens"`
			CachedTokens int    `json:"cached_tokens"`
			LastError    string `json:"last_error,omitempty"`
		}

		if err := json.Unmarshal(data, &usage); err != nil {
			return err
		}

		return displayStatsBox(usage, today, week)
	},
}

func displayStatsBox(usage map[string]map[string]struct {
	Requests     int    `json:"requests"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	CachedTokens int    `json:"cached_tokens"`
	LastError    string `json:"last_error,omitempty"`
}, todayOnly, weekOnly bool) error {
	now := time.Now()
	todayStr := now.Format("2006-01-02")

	var dates []string
	for date := range usage {
		if todayOnly && date != todayStr {
			continue
		}
		if weekOnly {
			d, _ := time.Parse("2006-01-02", date)
			if now.Sub(d) > 7*24*time.Hour {
				continue
			}
		}
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	totalRequests := 0
	totalErrors := 0
	totalInputTokens := 0
	totalOutputTokens := 0
	totalCachedTokens := 0
	modelStats := make(map[string]int)

	for _, date := range dates {
		models := usage[date]
		for model, stats := range models {
			totalRequests += stats.Requests
			totalInputTokens += stats.InputTokens
			totalOutputTokens += stats.OutputTokens
			totalCachedTokens += stats.CachedTokens
			if stats.LastError != "" {
				totalErrors++
			}
			modelStats[model] += stats.Requests
		}
	}

	successRate := 100.0
	if totalRequests > 0 {
		successRate = float64(totalRequests-totalErrors) / float64(totalRequests) * 100
	}

	cacheHitRate := 0.0
	if totalInputTokens > 0 {
		cacheHitRate = float64(totalCachedTokens) / float64(totalInputTokens) * 100
	}

	width := 88
	fmt.Println(strings.Repeat("─", width))
	fmt.Println()
	fmt.Println("  Usage Statistics")
	fmt.Println()

	period := "All Time"
	if todayOnly {
		period = "Today"
	} else if weekOnly {
		period = "Last 7 Days"
	}
	fmt.Printf("  Period:              %s\n", period)
	fmt.Printf("  Total Requests:      %d\n", totalRequests)
	fmt.Printf("  Success Rate:        %.1f%%\n", successRate)
	fmt.Println()

	if totalInputTokens > 0 || totalOutputTokens > 0 {
		fmt.Println("  Token Usage")
		fmt.Printf("  Input Tokens:        %s\n", formatNumber(totalInputTokens))
		fmt.Printf("  Output Tokens:       %s\n", formatNumber(totalOutputTokens))
		if totalCachedTokens > 0 {
			fmt.Printf("  Cached Tokens:       %s (%.1f%% cache hit)\n", formatNumber(totalCachedTokens), cacheHitRate)
			fmt.Println()
			fmt.Printf("   Cache savings: %s tokens, reducing costs\n", formatNumber(totalCachedTokens))
		}
		fmt.Println()
	}

	fmt.Println("  Model Usage          Requests  Status")
	fmt.Println("  " + strings.Repeat("─", width-4))

	type modelStat struct {
		name     string
		requests int
		hasError bool
	}
	var stats []modelStat
	for model, reqs := range modelStats {
		hasErr := false
		for _, date := range dates {
			if models, ok := usage[date]; ok {
				if m, ok := models[model]; ok && m.LastError != "" {
					hasErr = true
					break
				}
			}
		}
		stats = append(stats, modelStat{model, reqs, hasErr})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].requests > stats[j].requests
	})

	for _, s := range stats {
		status := "✓"
		if s.hasError {
			status = "⚠"
		}
		parts := strings.Split(s.name, "/")
		modelName := s.name
		if len(parts) > 1 {
			modelName = parts[len(parts)-1]
		}
		if len(modelName) > 30 {
			modelName = modelName[:27] + "..."
		}
		fmt.Printf("  %-32s %8d  %s\n", modelName, s.requests, status)
	}

	fmt.Println()
	fmt.Println("  » Tip: Use 'gptcode stats --today' for today's activity")
	fmt.Println()
	fmt.Println(strings.Repeat("─", width))

	return nil
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}

var budgetCmd = &cobra.Command{
	Use:   "budget [set|show] [amount]",
	Short: "Manage budget settings for cost-conscious usage",
	Long: `Manage budget settings to control costs and optimize for low-cost alternatives.

Commands:
  set [amount] - Set monthly budget limit in USD
  show         - Show current budget settings and usage
  enable       - Enable budget mode to prioritize free/low-cost models
  disable      - Disable budget mode

Examples:
  gptcode budget show          # Show current budget settings
  gptcode budget set 10        # Set monthly budget to $10
  gptcode budget enable        # Enable budget-conscious mode
  gptcode budget disable       # Disable budget mode`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := args[0]
		setup, err := config.LoadSetup()
		if err != nil {
			return fmt.Errorf("failed to load setup: %w", err)
		}

		switch command {
		case "show":
			fmt.Printf("Budget Mode: %t\n", setup.Defaults.BudgetMode)
			if setup.Defaults.MonthlyBudget > 0 {
				fmt.Printf("Monthly Budget: $%.2f\n", setup.Defaults.MonthlyBudget)
			} else {
				fmt.Printf("Monthly Budget: Not set\n")
			}
			if setup.Defaults.MaxCostPerTask > 0 {
				fmt.Printf("Max Cost Per Task: $%.4f\n", setup.Defaults.MaxCostPerTask)
			} else {
				fmt.Printf("Max Cost Per Task: Not set\n")
			}
			fmt.Printf("Current Mode: %s\n", setup.Defaults.Mode)
			fmt.Println("\nRecommendation: For low-cost usage, consider 'gptcode mode local' to use free local models")

		case "set":
			if len(args) < 2 {
				return fmt.Errorf("please provide a budget amount")
			}
			var amount float64
			fmt.Sscanf(args[1], "%f", &amount)
			if amount < 0 {
				return fmt.Errorf("budget amount must be non-negative")
			}
			setup.Defaults.MonthlyBudget = amount
			setup.Defaults.BudgetMode = true // Enable budget mode when setting a budget
			if err := config.SaveSetup(setup); err != nil {
				return fmt.Errorf("failed to save setup: %w", err)
			}
			fmt.Printf("✓ Monthly budget set to $%.2f and budget mode enabled\n", amount)

		case "enable":
			setup.Defaults.BudgetMode = true
			if setup.Defaults.MonthlyBudget == 0 {
				fmt.Println("⚠ Monthly budget not set. Consider setting a budget with 'gptcode budget set [amount]'")
			}
			if err := config.SaveSetup(setup); err != nil {
				return fmt.Errorf("failed to save setup: %w", err)
			}
			fmt.Println("✓ Budget mode enabled. System will prioritize free and low-cost models.")

		case "disable":
			setup.Defaults.BudgetMode = false
			if err := config.SaveSetup(setup); err != nil {
				return fmt.Errorf("failed to save setup: %w", err)
			}
			fmt.Println("✓ Budget mode disabled.")

		case "max-task":
			if len(args) < 2 {
				return fmt.Errorf("please provide a max cost per task")
			}
			var amount float64
			fmt.Sscanf(args[1], "%f", &amount)
			if amount < 0 {
				return fmt.Errorf("max cost per task must be non-negative")
			}
			setup.Defaults.MaxCostPerTask = amount
			setup.Defaults.BudgetMode = true // Enable budget mode when setting max cost
			if err := config.SaveSetup(setup); err != nil {
				return fmt.Errorf("failed to save setup: %w", err)
			}
			fmt.Printf("✓ Max cost per task set to $%.4f and budget mode enabled\n", amount)

		default:
			return fmt.Errorf("unknown command: %s. Use 'show', 'set', 'enable', 'disable', or 'max-task'", command)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(modeCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(budgetCmd)

	statsCmd.Flags().Bool("today", false, "Show today's stats only")
	statsCmd.Flags().Bool("week", false, "Show last 7 days")
	statsCmd.Flags().Bool("all", false, "Show all time stats")
}
