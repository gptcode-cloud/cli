package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var perfCmd = &cobra.Command{
	Use:   "perf",
	Short: "Performance profiling and optimization",
	Long:  `Profile code performance and identify bottlenecks.`,
}

var perfProfileCmd = &cobra.Command{
	Use:   "profile [target]",
	Short: "Run performance profiling",
	Long: `Profile application performance (CPU, memory, etc).

Examples:
  gptcode perf profile          # Profile with go test
  gptcode perf profile ./cmd    # Profile specific package`,
	RunE: runPerfProfile,
}

var perfBenchCmd = &cobra.Command{
	Use:   "bench [pattern]",
	Short: "Run benchmarks and analyze results",
	Long: `Run benchmarks and provide optimization suggestions.

Examples:
  gptcode perf bench          # Run all benchmarks
  gptcode perf bench BenchmarkFoo  # Run specific benchmark`,
	RunE: runPerfBench,
}

func init() {
	rootCmd.AddCommand(perfCmd)
	perfCmd.AddCommand(perfProfileCmd)
	perfCmd.AddCommand(perfBenchCmd)
}

func runPerfProfile(cmd *cobra.Command, args []string) error {
	target := "./..."
	if len(args) > 0 {
		target = args[0]
	}

	fmt.Println("🔥 Starting performance profiling...")

	cpuCmd := exec.Command("go", "test", "-cpuprofile=cpu.prof", "-memprofile=mem.prof", target)
	cpuCmd.Stdout = os.Stdout
	cpuCmd.Stderr = os.Stderr

	if err := cpuCmd.Run(); err != nil {
		return fmt.Errorf("profiling failed: %w", err)
	}

	fmt.Println("\n✅ Profiling complete!")
	fmt.Println("\n📊 Generated profiles:")
	fmt.Println("  - cpu.prof (CPU profile)")
	fmt.Println("  - mem.prof (Memory profile)")
	fmt.Println("\n💡 Analyze with:")
	fmt.Println("  go tool pprof cpu.prof")
	fmt.Println("  go tool pprof mem.prof")

	return nil
}

func runPerfBench(cmd *cobra.Command, args []string) error {
	pattern := ""
	if len(args) > 0 {
		pattern = args[0]
	}

	fmt.Println("⚡ Running benchmarks...")

	benchArgs := []string{"test", "-bench", "."}
	if pattern != "" {
		benchArgs = []string{"test", "-bench", pattern}
	}
	benchArgs = append(benchArgs, "-benchmem", "./...")

	benchCmd := exec.Command("go", benchArgs...)
	output, err := benchCmd.CombinedOutput()

	fmt.Print(string(output))

	if err != nil && !strings.Contains(string(output), "PASS") {
		return fmt.Errorf("benchmarks failed: %w", err)
	}

	fmt.Println("\n💡 Optimization tips:")
	if strings.Contains(string(output), "allocs/op") {
		fmt.Println("  - Review allocations for hot paths")
		fmt.Println("  - Consider object pooling for frequent allocations")
	}
	fmt.Println("  - Profile with: gptcode perf profile")
	fmt.Println("  - Compare with: benchstat old.txt new.txt")

	return nil
}
