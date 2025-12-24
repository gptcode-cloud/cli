package telemetry

import (
	"context"
	"testing"
	"time"
)

func TestNewTelemetry(t *testing.T) {
	tel := NewTelemetry()
	if tel == nil {
		t.Fatal("NewTelemetry returned nil")
	}

	if tel.tracer == nil {
		t.Error("Expected tracer to be initialized")
	}
}

func TestRecordStep(t *testing.T) {
	tel := NewTelemetry()
	ctx := context.Background()

	event := StepEvent{
		StepIndex:    0,
		StepName:     "Test Step",
		FilesTouched: []string{"file1.go", "file2.go"},
		Success:      true,
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(100 * time.Millisecond),
		DurationMs:   100,
	}

	// Should not panic
	tel.RecordStep(ctx, event)
}

func TestRecordStepWithError(t *testing.T) {
	tel := NewTelemetry()
	ctx := context.Background()

	event := StepEvent{
		StepIndex:    1,
		StepName:     "Failed Step",
		FilesTouched: []string{},
		Success:      false,
		ErrorMessage: "test error",
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(50 * time.Millisecond),
		DurationMs:   50,
	}

	// Should not panic
	tel.RecordStep(ctx, event)
}

func TestNewUsageTracker(t *testing.T) {
	tracker := NewUsageTracker()

	if tracker == nil {
		t.Fatal("NewUsageTracker returned nil")
	}

	if tracker.requests == nil {
		t.Error("Expected requests map to be initialized")
	}

	if tracker.tokens == nil {
		t.Error("Expected tokens map to be initialized")
	}
}

func TestRecordRequest(t *testing.T) {
	tracker := NewUsageTracker()

	tests := []struct {
		backend string
		model   string
		tokens  int
	}{
		{"openai", "gpt-4", 1000},
		{"openai", "gpt-4", 500},
		{"groq", "llama-3.3", 2000},
	}

	for _, tt := range tests {
		tracker.RecordRequest(tt.backend, tt.model, tt.tokens)
	}

	stats := tracker.GetStats()

	// Check openai/gpt-4
	key := "openai/gpt-4"
	if stat, ok := stats[key]; ok {
		if stat.Requests != 2 {
			t.Errorf("Expected 2 requests for %s, got %d", key, stat.Requests)
		}
		if stat.Tokens != 1500 {
			t.Errorf("Expected 1500 tokens for %s, got %d", key, stat.Tokens)
		}
		// Note: Cost will be calculated based on model catalog, so we check if it's positive
		if stat.Cost <= 0 {
			t.Errorf("Expected positive cost for %s, got %f", key, stat.Cost)
		}
	} else {
		t.Errorf("Expected stats for %s", key)
	}

	// Check groq/llama-3.3
	key = "groq/llama-3.3"
	if stat, ok := stats[key]; ok {
		if stat.Requests != 1 {
			t.Errorf("Expected 1 request for %s, got %d", key, stat.Requests)
		}
		if stat.Tokens != 2000 {
			t.Errorf("Expected 2000 tokens for %s, got %d", key, stat.Tokens)
		}
		if stat.Cost <= 0 {
			t.Errorf("Expected positive cost for %s, got %f", key, stat.Cost)
		}
	} else {
		t.Errorf("Expected stats for %s", key)
	}
}

func TestGetStatsEmpty(t *testing.T) {
	tracker := NewUsageTracker()
	stats := tracker.GetStats()

	if len(stats) != 0 {
		t.Errorf("Expected empty stats, got %d entries", len(stats))
	}
}

func TestUsageTrackerMultipleModels(t *testing.T) {
	tracker := NewUsageTracker()

	// Record usage for multiple models
	tracker.RecordRequest("openai", "gpt-4", 100)
	tracker.RecordRequest("openai", "gpt-3.5", 50)
	tracker.RecordRequest("groq", "llama-3.3", 200)

	stats := tracker.GetStats()

	if len(stats) != 3 {
		t.Errorf("Expected 3 models in stats, got %d", len(stats))
	}

	// Verify total requests
	totalRequests := 0
	totalTokens := 0
	totalCost := 0.0
	for _, stat := range stats {
		totalRequests += stat.Requests
		totalTokens += stat.Tokens
		totalCost += stat.Cost
	}

	if totalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", totalRequests)
	}

	if totalTokens != 350 {
		t.Errorf("Expected 350 total tokens, got %d", totalTokens)
	}

	if totalCost <= 0 {
		t.Errorf("Expected positive total cost, got %f", totalCost)
	}
}

func TestBudgetTracking(t *testing.T) {
	tracker := NewUsageTracker()

	// Record some requests with cheaper models
	tracker.RecordRequest("groq", "llama-3.1-8b-instant", 100000)  // 0.1M tokens at $0.05/M = $0.005
	tracker.RecordRequest("groq", "llama-3.1-8b-instant", 50000)   // 0.05M tokens at $0.05/M = $0.0025
	tracker.RecordRequest("groq", "llama-3.1-8b-instant", 200000) // 0.2M tokens at $0.05/M = $0.01

	// Check total cost
	totalCost := tracker.GetTotalCost()
	if totalCost <= 0 {
		t.Errorf("Expected positive total cost, got %f", totalCost)
	}

	// Test budget checking - with budget that should be exceeded
	exceeded, remaining := tracker.CheckBudget(0.01) // $0.01 budget (should be exceeded since we spent ~$0.0175)
	if !exceeded {
		t.Errorf("Expected budget to be exceeded, but it wasn't. Total cost: %f", totalCost)
	}
	if remaining >= 0 {
		t.Errorf("Expected negative remaining, got %f", remaining)
	}

	// Test budget checking - with budget that should not be exceeded
	exceeded, remaining = tracker.CheckBudget(1.0) // $1 budget
	if exceeded {
		t.Errorf("Expected budget to not be exceeded, but it was. Total cost: %f, budget: 1.0", totalCost)
	}
	if remaining <= 0 {
		t.Errorf("Expected positive remaining, got %f", remaining)
	}

	// Test budget checking - with no budget set
	exceeded, remaining = tracker.CheckBudget(0) // No budget
	if exceeded {
		t.Errorf("Expected no budget exceeded when budget is 0, got exceeded: %t", exceeded)
	}
	if remaining != -1 {
		t.Errorf("Expected -1 remaining when no budget set, got %f", remaining)
	}
}
