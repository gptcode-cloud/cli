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
	for _, stat := range stats {
		totalRequests += stat.Requests
		totalTokens += stat.Tokens
	}

	if totalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", totalRequests)
	}

	if totalTokens != 350 {
		t.Errorf("Expected 350 total tokens, got %d", totalTokens)
	}
}
