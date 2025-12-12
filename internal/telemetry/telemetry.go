package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Telemetry provides OpenTelemetry-based telemetry tracking
type Telemetry struct {
	tracer trace.Tracer
}

// NewTelemetry creates a new telemetry instance
func NewTelemetry() *Telemetry {
	return &Telemetry{
		tracer: otel.Tracer("gptcode"),
	}
}

// StepEvent represents a single step in execution
type StepEvent struct {
	StepIndex    int
	StepName     string
	FilesTouched []string
	Success      bool
	ErrorMessage string
	StartTime    time.Time
	EndTime      time.Time
	DurationMs   int64
}

// RecordStep records a complete step execution
func (t *Telemetry) RecordStep(ctx context.Context, event StepEvent) {
	_, span := t.tracer.Start(ctx, event.StepName)
	defer span.End()

	span.SetAttributes(
		attribute.Int("step.index", event.StepIndex),
		attribute.StringSlice("step.files_touched", event.FilesTouched),
		attribute.Bool("step.success", event.Success),
		attribute.Int64("step.duration_ms", event.DurationMs),
	)

	if event.ErrorMessage != "" {
		span.SetAttributes(attribute.String("step.error", event.ErrorMessage))
	}
}

// UsageTracker tracks API request and token usage
type UsageTracker struct {
	requests map[string]int // backend/model -> request count
	tokens   map[string]int // backend/model -> token count
}

// NewUsageTracker creates a new usage tracker
func NewUsageTracker() *UsageTracker {
	return &UsageTracker{
		requests: make(map[string]int),
		tokens:   make(map[string]int),
	}
}

// RecordRequest records an API request
func (u *UsageTracker) RecordRequest(backend, model string, tokens int) {
	key := backend + "/" + model
	u.requests[key]++
	u.tokens[key] += tokens
}

// GetStats returns usage statistics
func (u *UsageTracker) GetStats() map[string]UsageStats {
	stats := make(map[string]UsageStats)
	for key, requests := range u.requests {
		stats[key] = UsageStats{
			Requests: requests,
			Tokens:   u.tokens[key],
		}
	}
	return stats
}

// UsageStats represents usage statistics for a model
type UsageStats struct {
	Requests int
	Tokens   int
}
