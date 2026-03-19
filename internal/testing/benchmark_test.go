package testing

import (
	"testing"
	"time"
)

func BenchmarkGTVersion(b *testing.B) {
	cli := NewCLI("gt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cli.Run("version")
	}
}

func BenchmarkGTHelp(b *testing.B) {
	cli := NewCLI("gt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cli.Run("--help")
	}
}

func BenchmarkGTStatus(b *testing.B) {
	cli := NewCLI("gt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cli.Run("status")
	}
}

func BenchmarkGTModelList(b *testing.B) {
	cli := NewCLI("gt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cli.Run("model", "list")
	}
}

type PerformanceMetrics struct {
	Iterations int
	TotalTime  time.Duration
	AvgTime    time.Duration
	MinTime    time.Duration
	MaxTime    time.Duration
	P50Time    time.Duration
	P95Time    time.Duration
	P99Time    time.Duration
}

func RunPerformanceTest(t *testing.T, name string, fn func()) *PerformanceMetrics {
	iterations := 10
	times := make([]time.Duration, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		fn()
		times[i] = time.Since(start)
	}

	var total time.Duration
	var min = time.Hour
	var max time.Duration

	for _, d := range times {
		total += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	avg := total / time.Duration(iterations)

	// Simple percentile calculation
	p50 := times[iterations/2]
	p95 := times[int(float64(iterations)*0.95)]
	p99Idx := int(float64(iterations) * 0.99)
	if p99Idx >= iterations {
		p99Idx = iterations - 1
	}
	p99 := times[p99Idx]

	return &PerformanceMetrics{
		Iterations: iterations,
		TotalTime:  total,
		AvgTime:    avg,
		MinTime:    min,
		MaxTime:    max,
		P50Time:    p50,
		P95Time:    p95,
		P99Time:    p99,
	}
}

func (m *PerformanceMetrics) Report(t *testing.T) {
	t.Logf("Performance: %d iterations", m.Iterations)
	t.Logf("Total Time: %v", m.TotalTime)
	t.Logf("Avg Time: %v", m.AvgTime)
	t.Logf("Min Time: %v", m.MinTime)
	t.Logf("Max Time: %v", m.MaxTime)
	t.Logf("P50: %v", m.P50Time)
	t.Logf("P95: %v", m.P95Time)
	t.Logf("P99: %v", m.P99Time)
}

func TestPerformanceGTVersion(t *testing.T) {
	cli := NewCLI("gt")
	metrics := RunPerformanceTest(t, "GT Version", func() {
		cli.Run("version")
	})
	metrics.Report(t)
}

func TestPerformanceGTStatus(t *testing.T) {
	cli := NewCLI("gt")
	cli.workDir = t.TempDir()
	metrics := RunPerformanceTest(t, "GT Status", func() {
		cli.Run("status")
	})
	metrics.Report(t)
}

func TestPerformanceGTHelp(t *testing.T) {
	cli := NewCLI("gt")
	metrics := RunPerformanceTest(t, "GT Help", func() {
		cli.Run("--help")
	})
	metrics.Report(t)
}
