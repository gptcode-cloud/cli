package observability

import (
	"time"
)

// Decision represents a routing or algorithmic choice made by the system
type Decision struct {
	Type         string             `json:"type"` // "route", "model_selection", "retry", etc.
	Chosen       string             `json:"chosen"`
	Alternatives []string           `json:"alternatives,omitempty"`
	Attribution  map[string]float64 `json:"attribution,omitempty"` // feature -> weight
	Reasoning    string             `json:"reasoning,omitempty"`
}

// Metrics captures performance and cost data
type Metrics struct {
	DurationMs   int64   `json:"duration_ms"`
	TokensIn     int     `json:"tokens_in,omitempty"`
	TokensOut    int     `json:"tokens_out,omitempty"`
	Cost         float64 `json:"cost,omitempty"`
	CacheHit     bool    `json:"cache_hit,omitempty"`
	RetryCount   int     `json:"retry_count,omitempty"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// StepTrace represents a single step in the execution flow
type StepTrace struct {
	Node      string                 `json:"node"`
	Timestamp time.Time              `json:"timestamp"`
	Metrics   Metrics                `json:"metrics"`
	Inputs    map[string]interface{} `json:"inputs,omitempty"`
	Outputs   map[string]interface{} `json:"outputs,omitempty"`
	Decision  *Decision              `json:"decision,omitempty"`
}

// SessionTrace represents a complete command execution
type SessionTrace struct {
	SessionID   string      `json:"session_id"`
	Command     string      `json:"command"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     *time.Time  `json:"end_time,omitempty"`
	Steps       []StepTrace `json:"steps"`
	Path        []string    `json:"path"` // Ordered list of nodes visited
	TotalCost   float64     `json:"total_cost"`
	TotalTimeMs int64       `json:"total_time_ms"`
	Success     bool        `json:"success"`
}

// Tracer is the interface for recording execution traces
type Tracer interface {
	// Begin starts a new session trace
	Begin(sessionID, command string) error

	// RecordStep logs a step execution
	RecordStep(step StepTrace) error

	// RecordDecision logs a routing or choice decision
	RecordDecision(node string, decision Decision) error

	// RecordMetrics logs performance metrics
	RecordMetrics(node string, metrics Metrics) error

	// End finalizes the session trace
	End(success bool) error

	// Flush writes accumulated traces to storage
	Flush() error
}

// Example usage pattern:
//
// func (o *Orchestrator) Execute(cmd string) error {
//     tracer := observability.NewTracer()
//     sessionID := uuid.New().String()
//
//     tracer.Begin(sessionID, cmd)
//     defer tracer.End(true)
//
//     // At decision points
//     decision := o.intelligenceSystem.Route(context)
//     tracer.RecordDecision("IntelligenceSystem", Decision{
//         Type: "route",
//         Chosen: decision.Agent.Name(),
//         Alternatives: []string{"DoCommand", "OrchestratedMode"},
//         Attribution: map[string]float64{
//             "complexity_score": 0.6,
//             "multi_file": 0.3,
//             "has_context": 0.1,
//         },
//         Reasoning: "High complexity task requiring planning",
//     })
//
//     // At execution points
//     start := time.Now()
//     result := agent.Execute(context)
//     tracer.RecordMetrics("PlannerAgent", Metrics{
//         DurationMs: time.Since(start).Milliseconds(),
//         TokensIn: result.TokensUsed.Input,
//         TokensOut: result.TokensUsed.Output,
//         Cost: result.Cost,
//     })
//
//     return nil
// }
