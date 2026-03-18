package autonomous

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Telemetry struct {
	events     []TelemetryEvent
	metrics    *Metrics
	mu         sync.Mutex
	sessionID  string
	startTime  time.Time
	outputFile string
}

type TelemetryEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type Metrics struct {
	TasksCompleted  int
	TasksFailed     int
	TotalTokensIn   int64
	TotalTokensOut  int64
	TotalCost       float64
	AvgResponseTime time.Duration
	mu              sync.Mutex
}

func NewTelemetry(outputFile string) *Telemetry {
	return &Telemetry{
		events:     []TelemetryEvent{},
		metrics:    &Metrics{},
		sessionID:  generateSessionID(),
		startTime:  time.Now(),
		outputFile: outputFile,
	}
}

func (t *Telemetry) Track(eventType string, data map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	event := TelemetryEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	t.events = append(t.events, event)

	// Update metrics based on event type
	switch eventType {
	case "task_complete":
		t.metrics.TasksCompleted++
	case "task_failed":
		t.metrics.TasksFailed++
	case "llm_request":
		if tokens, ok := data["tokens_in"].(int); ok {
			t.metrics.TotalTokensIn += int64(tokens)
		}
		if tokens, ok := data["tokens_out"].(int); ok {
			t.metrics.TotalTokensOut += int64(tokens)
		}
		if cost, ok := data["cost"].(float64); ok {
			t.metrics.TotalCost += cost
		}
	}
}

func (t *Telemetry) GetMetrics() *Metrics {
	t.metrics.mu.Lock()
	defer t.metrics.mu.Unlock()
	return &Metrics{
		TasksCompleted: t.metrics.TasksCompleted,
		TasksFailed:    t.metrics.TasksFailed,
		TotalTokensIn:  t.metrics.TotalTokensIn,
		TotalTokensOut: t.metrics.TotalTokensOut,
		TotalCost:      t.metrics.TotalCost,
	}
}

func (t *Telemetry) Save() error {
	if t.outputFile == "" {
		return nil
	}

	dir := filepath.Dir(t.outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data := struct {
		SessionID string           `json:"session_id"`
		StartTime time.Time        `json:"start_time"`
		EndTime   time.Time        `json:"end_time"`
		Duration  time.Duration    `json:"duration"`
		Metrics   *Metrics         `json:"metrics"`
		Events    []TelemetryEvent `json:"events"`
	}{
		SessionID: t.sessionID,
		StartTime: t.startTime,
		EndTime:   time.Now(),
		Duration:  time.Since(t.startTime),
		Metrics:   t.GetMetrics(),
		Events:    t.events,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(t.outputFile, jsonData, 0644)
}

func (t *Telemetry) Summary() string {
	m := t.GetMetrics()
	return fmt.Sprintf(`Telemetry Summary:
  Session: %s
  Duration: %v
  Tasks: %d completed, %d failed
  Tokens: %d in, %d out
  Cost: $%.4f`,
		t.sessionID[:8],
		time.Since(t.startTime),
		m.TasksCompleted,
		m.TasksFailed,
		m.TotalTokensIn,
		m.TotalTokensOut,
		m.TotalCost,
	)
}

func (t *Telemetry) SuccessRate() float64 {
	total := t.metrics.TasksCompleted + t.metrics.TasksFailed
	if total == 0 {
		return 0
	}
	return float64(t.metrics.TasksCompleted) / float64(total) * 100
}

func generateSessionID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
		time.Sleep(1 * time.Nanosecond)
	}
	return fmt.Sprintf("%x", b)
}

type MetricsCollector struct {
	counters map[string]int
	gauges   map[string]float64
	histo    map[string][]time.Duration
	mu       sync.Mutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters: make(map[string]int),
		gauges:   make(map[string]float64),
		histo:    make(map[string][]time.Duration),
	}
}

func (mc *MetricsCollector) IncrementCounter(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.counters[name]++
}

func (mc *MetricsCollector) SetGauge(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.gauges[name] = value
}

func (mc *MetricsCollector) RecordDuration(name string, d time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.histo[name] = append(mc.histo[name], d)
}

func (mc *MetricsCollector) GetReport() string {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	var report string
	report += "Metrics Report\n"
	report += "==============\n\n"

	report += "Counters:\n"
	for k, v := range mc.counters {
		report += fmt.Sprintf("  %s: %d\n", k, v)
	}

	report += "\nGauges:\n"
	for k, v := range mc.gauges {
		report += fmt.Sprintf("  %s: %.2f\n", k, v)
	}

	report += "\nHistograms:\n"
	for k, v := range mc.histo {
		if len(v) > 0 {
			var sum time.Duration
			for _, d := range v {
				sum += d
			}
			avg := sum / time.Duration(len(v))
			report += fmt.Sprintf("  %s: count=%d avg=%v\n", k, len(v), avg)
		}
	}

	return report
}

type PerformanceTracker struct {
	startTime time.Time
	tasks     map[string]*TaskMetrics
	mu        sync.Mutex
}

type TaskMetrics struct {
	Name      string
	Duration  time.Duration
	TokensIn  int
	TokensOut int
	Cost      float64
	Retries   int
	Success   bool
}

func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		startTime: time.Now(),
		tasks:     make(map[string]*TaskMetrics),
	}
}

func (pt *PerformanceTracker) StartTask(name string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.tasks[name] = &TaskMetrics{Name: name}
}

func (pt *PerformanceTracker) EndTask(name string, success bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if t, ok := pt.tasks[name]; ok {
		t.Duration = time.Since(pt.startTime)
		t.Success = success
	}
}

func (pt *PerformanceTracker) SetTaskMetrics(name string, tokensIn, tokensOut int, cost float64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if t, ok := pt.tasks[name]; ok {
		t.TokensIn = tokensIn
		t.TokensOut = tokensOut
		t.Cost = cost
	}
}

func (pt *PerformanceTracker) GetReport() string {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	var totalCost float64
	var totalTokensIn, totalTokensOut int
	var successCount, failCount int

	for _, t := range pt.tasks {
		totalCost += t.Cost
		totalTokensIn += t.TokensIn
		totalTokensOut += t.TokensOut
		if t.Success {
			successCount++
		} else {
			failCount++
		}
	}

	return fmt.Sprintf(`Performance Report
==================
Tasks: %d completed, %d failed
Tokens: %d in, %d out
Total Cost: $%.4f
Uptime: %v`,
		successCount,
		failCount,
		totalTokensIn,
		totalTokensOut,
		totalCost,
		time.Since(pt.startTime),
	)
}
