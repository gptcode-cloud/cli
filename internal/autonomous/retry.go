package autonomous

import (
	"context"
	"strings"
	"time"
)

type SmartRetry struct {
	maxAttempts int
	backoff     time.Duration
	strategies  []RetryStrategy
}

type RetryStrategy func(err error, attempt int) (RetryDecision, string)

type RetryDecision int

const (
	RetryContinue RetryDecision = iota
	RetryAbort
	RetrySkip
	RetryAlternative
)

type RetryResult struct {
	Success     bool
	Attempts    int
	FinalError  error
	Strategy    string
	RecoveryTip string
}

func NewSmartRetry(maxAttempts int) *SmartRetry {
	return &SmartRetry{
		maxAttempts: maxAttempts,
		backoff:     1 * time.Second,
		strategies:  []RetryStrategy{},
	}
}

func (sr *SmartRetry) AddStrategy(strategy RetryStrategy) {
	sr.strategies = append(sr.strategies, strategy)
}

func (sr *SmartRetry) Execute(ctx context.Context, task string, fn func() error) (*RetryResult, error) {
	result := &RetryResult{}
	var lastErr error

	for attempt := 1; attempt <= sr.maxAttempts; attempt++ {
		result.Attempts = attempt

		// Execute task
		err := fn()
		if err == nil {
			result.Success = true
			return result, nil
		}

		lastErr = err
		errMsg := err.Error()

		// Analyze error and decide next step
		decision, strategy := sr.analyzeError(err, attempt)
		result.Strategy = strategy

		switch decision {
		case RetryAbort:
			result.FinalError = err
			result.RecoveryTip = sr.getRecoveryTip(errMsg)
			return result, nil

		case RetrySkip:
			return result, nil

		case RetryAlternative:
			// Try alternative approach
			result.RecoveryTip = "Trying alternative approach"
			continue
		}

		// Retry with backoff
		if attempt < sr.maxAttempts {
			backoff := sr.backoff * time.Duration(attempt)
			time.Sleep(backoff)
		}
	}

	result.FinalError = lastErr
	result.RecoveryTip = sr.getRecoveryTip(lastErr.Error())
	return result, nil
}

func (sr *SmartRetry) analyzeError(err error, attempt int) (RetryDecision, string) {
	errMsg := strings.ToLower(err.Error())

	// Immediate abort errors
	abortPatterns := []string{
		"permission denied",
		"connection refused",
		"authentication failed",
		"unauthorized",
		"not found",
		"invalid credentials",
	}

	for _, pattern := range abortPatterns {
		if strings.Contains(errMsg, pattern) {
			return RetryAbort, "abort:" + pattern
		}
	}

	// Retryable errors
	retryPatterns := []string{
		"timeout",
		"rate limit",
		"temporary failure",
		"network",
		"connection reset",
		"too many requests",
	}

	for _, pattern := range retryPatterns {
		if strings.Contains(errMsg, pattern) {
			return RetryContinue, "retry:" + pattern
		}
	}

	// Test failures - might need code fix
	if strings.Contains(errMsg, "test") && strings.Contains(errMsg, "fail") {
		if attempt >= 2 {
			return RetryAbort, "needs_code_fix"
		}
		return RetryContinue, "retry:test_failure"
	}

	// Compilation errors - might need fix
	if strings.Contains(errMsg, "syntax") || strings.Contains(errMsg, "compile") {
		if attempt >= 2 {
			return RetryAbort, "needs_compile_fix"
		}
		return RetryContinue, "retry:compile_error"
	}

	// Custom strategies
	for _, strategy := range sr.strategies {
		decision, strategy := strategy(err, attempt)
		if decision != RetryContinue {
			return decision, strategy
		}
	}

	// Default: retry once more
	if attempt >= sr.maxAttempts {
		return RetryAbort, "max_attempts"
	}
	return RetryContinue, "default_retry"
}

func (sr *SmartRetry) getRecoveryTip(errMsg string) string {
	tips := map[string]string{
		"timeout":            "Try increasing timeout or check network connectivity",
		"rate limit":         "Wait a few seconds and retry, or use a slower rate",
		"permission denied":  "Check file/directory permissions",
		"connection refused": "Ensure the service is running",
		"test fail":          "Fix the failing tests before proceeding",
		"syntax":             "Fix syntax errors in the code",
		"compile":            "Fix compilation errors",
	}

	for pattern, tip := range tips {
		if strings.Contains(strings.ToLower(errMsg), pattern) {
			return tip
		}
	}

	return "Manual intervention required"
}

type ErrorClassifier struct{}

func NewErrorClassifier() *ErrorClassifier {
	return &ErrorClassifier{}
}

type ErrorCategory struct {
	Category string
	Severity string
	CanRetry bool
	Tip      string
}

func (ec *ErrorClassifier) Classify(err error) *ErrorCategory {
	errMsg := strings.ToLower(err.Error())

	// Syntax errors
	if strings.Contains(errMsg, "syntax") || strings.Contains(errMsg, "parse error") {
		return &ErrorCategory{
			Category: "syntax",
			Severity: "high",
			CanRetry: false,
			Tip:      "Fix syntax errors before retrying",
		}
	}

	// Test failures
	if strings.Contains(errMsg, "test") && strings.Contains(errMsg, "fail") {
		return &ErrorCategory{
			Category: "test",
			Severity: "medium",
			CanRetry: true,
			Tip:      "Fix failing tests or update snapshots",
		}
	}

	// Network errors
	networkPatterns := []string{"timeout", "network", "connection", "dns"}
	for _, pattern := range networkPatterns {
		if strings.Contains(errMsg, pattern) {
			return &ErrorCategory{
				Category: "network",
				Severity: "low",
				CanRetry: true,
				Tip:      "Check network connection and retry",
			}
		}
	}

	// Permission errors
	if strings.Contains(errMsg, "permission") || strings.Contains(errMsg, "denied") {
		return &ErrorCategory{
			Category: "permission",
			Severity: "high",
			CanRetry: false,
			Tip:      "Fix file/directory permissions",
		}
	}

	// Rate limit
	if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "too many") {
		return &ErrorCategory{
			Category: "rate_limit",
			Severity: "low",
			CanRetry: true,
			Tip:      "Wait and retry with exponential backoff",
		}
	}

	// Default
	return &ErrorCategory{
		Category: "unknown",
		Severity: "unknown",
		CanRetry: true,
		Tip:      "Investigate the error",
	}
}

// RetryPolicy defines when to retry
type RetryPolicy struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	RetryOnError []string
	AbortOnError []string
}

func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		RetryOnError: []string{
			"timeout",
			"network",
			"connection",
			"rate limit",
			"temporary",
		},
		AbortOnError: []string{
			"permission denied",
			"not found",
			"unauthorized",
			"authentication",
		},
	}
}

func (p *RetryPolicy) ShouldRetry(err error) bool {
	errMsg := strings.ToLower(err.Error())

	// Check abort conditions first
	for _, pattern := range p.AbortOnError {
		if strings.Contains(errMsg, pattern) {
			return false
		}
	}

	// Check retry conditions
	for _, pattern := range p.RetryOnError {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	// Default: don't retry
	return false
}

func (p *RetryPolicy) NextDelay(attempt int) time.Duration {
	delay := float64(p.InitialDelay) * pow(p.Multiplier, float64(attempt-1))
	if delay > float64(p.MaxDelay) {
		delay = float64(p.MaxDelay)
	}
	return time.Duration(delay)
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}
