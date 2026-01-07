package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

var trainingCmd = &cobra.Command{
	Use:   "training",
	Short: "Training loop management and analysis",
}

var trainingAnalyzeCmd = &cobra.Command{
	Use:   "analyze [results-file]",
	Short: "Analyze training results for patterns and improvement opportunities",
	Long: `Analyze training loop results to identify:
- Common failure patterns
- Model performance by language
- Success rate trends
- Recommendations for improvement`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resultsFile := filepath.Join(os.Getenv("HOME"), "workspace/gptcode/cli/training-sandbox/training_results_v2.json")
		if len(args) > 0 {
			resultsFile = args[0]
		}

		data, err := os.ReadFile(resultsFile)
		if err != nil {
			return fmt.Errorf("failed to read results file: %w", err)
		}

		var results TrainingResults
		if err := json.Unmarshal(data, &results); err != nil {
			return fmt.Errorf("failed to parse results: %w", err)
		}

		analyze(results)
		return nil
	},
}

type TrainingResults struct {
	Version   int           `json:"version"`
	StartedAt string        `json:"started_at"`
	Runs      []TrainingRun `json:"runs"`
	Stats     TrainingStats `json:"stats"`
}

type TrainingRun struct {
	Repo         string `json:"repo"`
	Issue        int    `json:"issue"`
	Language     string `json:"language"`
	Result       string `json:"result"`
	L1Syntax     bool   `json:"l1_syntax"`
	L2Review     string `json:"l2_review"`
	L3Tests      bool   `json:"l3_tests"`
	L4Analysis   bool   `json:"l4_analysis"`
	Duration     int    `json:"duration"`
	FilesChanged int    `json:"files_changed"`
}

type TrainingStats struct {
	Total            int `json:"total"`
	L1SyntaxPass     int `json:"l1_syntax_pass"`
	L2ReviewApproved int `json:"l2_review_approved"`
	L3TestsPass      int `json:"l3_tests_pass"`
	L4AnalysisClean  int `json:"l4_analysis_clean"`
	FullSuccess      int `json:"full_success"`
	Skipped          int `json:"skipped"`
	Failed           int `json:"failed"`
}

func analyze(results TrainingResults) {
	fmt.Println("📊 Training Analysis Report")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Started: %s\n", results.StartedAt)
	fmt.Printf("Total runs: %d\n\n", len(results.Runs))

	// Overall stats
	fmt.Println("## Overall Statistics")
	fmt.Printf("  Full Success: %d/%d (%.1f%%)\n",
		results.Stats.FullSuccess, results.Stats.Total,
		percent(results.Stats.FullSuccess, results.Stats.Total))
	fmt.Printf("  L1 Syntax Pass: %d\n", results.Stats.L1SyntaxPass)
	fmt.Printf("  L2 Review Approved: %d\n", results.Stats.L2ReviewApproved)
	fmt.Printf("  L3 Tests Pass: %d\n", results.Stats.L3TestsPass)
	fmt.Printf("  L4 Analysis Clean: %d\n", results.Stats.L4AnalysisClean)
	fmt.Println()

	// By language
	fmt.Println("## Performance by Language")
	langStats := make(map[string]struct{ total, success int })
	for _, run := range results.Runs {
		stats := langStats[run.Language]
		stats.total++
		if run.Result == "success" {
			stats.success++
		}
		langStats[run.Language] = stats
	}

	for lang, stats := range langStats {
		fmt.Printf("  %s: %d/%d (%.1f%%)\n", lang, stats.success, stats.total, percent(stats.success, stats.total))
	}
	fmt.Println()

	// Review verdict distribution
	fmt.Println("## Review Verdict Distribution")
	verdicts := make(map[string]int)
	for _, run := range results.Runs {
		verdicts[run.L2Review]++
	}
	for verdict, count := range verdicts {
		fmt.Printf("  %s: %d\n", verdict, count)
	}
	fmt.Println()

	// Identify patterns in failures
	fmt.Println("## Failure Patterns")
	patterns := identifyPatterns(results.Runs)
	for _, p := range patterns {
		fmt.Printf("  ⚠️  %s: %d occurrences\n", p.name, p.count)
	}
	fmt.Println()

	// Recommendations
	fmt.Println("## Recommendations")
	recommendations := generateRecommendations(results)
	for i, rec := range recommendations {
		fmt.Printf("  %d. %s\n", i+1, rec)
	}
}

type pattern struct {
	name  string
	count int
}

func identifyPatterns(runs []TrainingRun) []pattern {
	patterns := make(map[string]int)

	for _, run := range runs {
		// L2 failure patterns
		if !run.L1Syntax && run.L2Review != "approved" {
			patterns["L2: Review rejected changes"]++
		}

		// L3 failure patterns
		if run.L1Syntax && !run.L3Tests {
			patterns["L3: Tests failed after changes"]++
		}

		// L4 failure patterns
		if run.L1Syntax && !run.L4Analysis {
			patterns["L4: Error patterns in execution logs"]++
		}

		// Duration outliers
		if run.Duration > 120 {
			patterns["Slow execution (>2min)"]++
		}
	}

	var result []pattern
	for name, count := range patterns {
		result = append(result, pattern{name, count})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].count > result[j].count
	})

	return result
}

func generateRecommendations(results TrainingResults) []string {
	var recs []string

	// Check review approval rate
	if results.Stats.L2ReviewApproved < results.Stats.L1SyntaxPass/2 {
		recs = append(recs, "Low review approval rate. Consider improving code-review skill or prompts.")
	}

	// Check test pass rate
	if results.Stats.L3TestsPass < results.Stats.L1SyntaxPass/2 {
		recs = append(recs, "Many test failures. Consider adding test awareness to editor prompts.")
	}

	// Check for language gaps
	langSuccess := make(map[string]float64)
	langTotal := make(map[string]int)
	for _, run := range results.Runs {
		langTotal[run.Language]++
		if run.Result == "success" {
			langSuccess[run.Language]++
		}
	}

	for lang, total := range langTotal {
		rate := langSuccess[lang] / float64(total)
		if rate < 0.3 && total >= 5 {
			recs = append(recs, fmt.Sprintf("Poor performance on %s (%.0f%%). Consider improving %s skill.", lang, rate*100, lang))
		}
	}

	if len(recs) == 0 {
		recs = append(recs, "No major issues identified. Continue training for more data.")
	}

	return recs
}

func percent(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) * 100 / float64(b)
}

var trainingStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show training statistics summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		resultsFile := filepath.Join(os.Getenv("HOME"), "workspace/gptcode/cli/training-sandbox/training_results_v2.json")

		data, err := os.ReadFile(resultsFile)
		if err != nil {
			return fmt.Errorf("no training results found")
		}

		var results TrainingResults
		if err := json.Unmarshal(data, &results); err != nil {
			return fmt.Errorf("failed to parse results: %w", err)
		}

		fmt.Println("📊 Training Stats")
		fmt.Printf("  Total: %d | Success: %d (%.1f%%)\n",
			results.Stats.Total,
			results.Stats.FullSuccess,
			percent(results.Stats.FullSuccess, results.Stats.Total))
		fmt.Printf("  L1 Syntax: %d | L2 Review: %d | L3 Tests: %d | L4 Clean: %d\n",
			results.Stats.L1SyntaxPass,
			results.Stats.L2ReviewApproved,
			results.Stats.L3TestsPass,
			results.Stats.L4AnalysisClean)

		return nil
	},
}

func init() {
	trainingCmd.AddCommand(trainingAnalyzeCmd)
	trainingCmd.AddCommand(trainingStatsCmd)
}
