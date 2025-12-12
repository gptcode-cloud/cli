package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ActionType string

const (
	ActionEdit     ActionType = "edit"
	ActionReview   ActionType = "review"
	ActionPlan     ActionType = "plan"
	ActionResearch ActionType = "research"
	ActionRoute    ActionType = "route"
)

type ModelCapabilities struct {
	SupportsTools          bool   `json:"supports_tools"`
	SupportsFileOperations bool   `json:"supports_file_operations"`
	SupportsCodeExecution  bool   `json:"supports_code_execution"`
	Notes                  string `json:"notes"`
}

type ModelInfo struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	CostPer1M      float64           `json:"cost_per_1m"`
	RateLimitDaily int               `json:"rate_limit_daily"`
	ContextWindow  int               `json:"context_window"`
	TokensPerSec   int               `json:"tokens_per_sec"`
	Capabilities   ModelCapabilities `json:"capabilities"`
	Backend        string
}

type ModelFeedback struct {
	ModelID    string     `json:"model_id"`
	Action     ActionType `json:"action"`
	Language   string     `json:"language"`
	Success    bool       `json:"success"`
	Complexity string     `json:"complexity"` // simple, complex, multistep
}

type ModelUsage struct {
	Requests     int    `json:"requests"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	CachedTokens int    `json:"cached_tokens"`
	LastError    string `json:"last_error,omitempty"`
}

type ModelSelector struct {
	catalog  map[string][]ModelInfo
	feedback []ModelFeedback
	usage    map[string]map[string]ModelUsage
	setup    *Setup
}

func NewModelSelector(setup *Setup) (*ModelSelector, error) {
	selector := &ModelSelector{
		catalog:  make(map[string][]ModelInfo),
		feedback: []ModelFeedback{},
		usage:    make(map[string]map[string]ModelUsage),
		setup:    setup,
	}

	if err := selector.loadCatalog(); err != nil {
		return nil, fmt.Errorf("failed to load catalog: %w", err)
	}

	if err := selector.loadFeedback(); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] Could not load feedback: %v\n", err)
	}

	if err := selector.loadUsage(); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] Could not load usage: %v\n", err)
	}

	return selector, nil
}

func (ms *ModelSelector) loadCatalog() error {
	catalogPath := filepath.Join(configDir(), "models_catalog.json")
	data, err := os.ReadFile(catalogPath)
	if err != nil {
		return err
	}

	var rawCatalog map[string]interface{}
	if err := json.Unmarshal(data, &rawCatalog); err != nil {
		return err
	}

	for backend, backendData := range rawCatalog {
		backendMap, ok := backendData.(map[string]interface{})
		if !ok {
			continue
		}

		modelsData, ok := backendMap["models"].([]interface{})
		if !ok {
			continue
		}

		for _, modelData := range modelsData {
			modelMap, ok := modelData.(map[string]interface{})
			if !ok {
				continue
			}

			model := ModelInfo{
				Backend: backend,
			}

			if id, ok := modelMap["id"].(string); ok {
				model.ID = id
			}
			if name, ok := modelMap["name"].(string); ok {
				model.Name = name
			}
			if cost, ok := modelMap["cost_per_1m"].(float64); ok {
				model.CostPer1M = cost
			}
			if limit, ok := modelMap["rate_limit_daily"].(float64); ok {
				model.RateLimitDaily = int(limit)
			}
			if ctx, ok := modelMap["context_window"].(float64); ok {
				model.ContextWindow = int(ctx)
			}
			if tps, ok := modelMap["tokens_per_sec"].(float64); ok {
				model.TokensPerSec = int(tps)
			}

			// Default capabilities to true for backward compatibility
			model.Capabilities.SupportsTools = true
			model.Capabilities.SupportsFileOperations = true
			model.Capabilities.SupportsCodeExecution = false

			if caps, ok := modelMap["capabilities"].(map[string]interface{}); ok {
				if val, ok := caps["supports_tools"].(bool); ok {
					model.Capabilities.SupportsTools = val
				}
				if val, ok := caps["supports_file_operations"].(bool); ok {
					model.Capabilities.SupportsFileOperations = val
				}
				if val, ok := caps["supports_code_execution"].(bool); ok {
					model.Capabilities.SupportsCodeExecution = val
				}
				if val, ok := caps["notes"].(string); ok {
					model.Capabilities.Notes = val
				}
			}

			ms.catalog[backend] = append(ms.catalog[backend], model)
		}
	}

	return nil
}

func (ms *ModelSelector) loadFeedback() error {
	// Try to load from feedback system first (new location)
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Import feedback package to use existing system
	// For now, we'll load directly from the feedback directory
	feedbackDir := filepath.Join(home, ".gptcode", "feedback")
	entries, err := os.ReadDir(feedbackDir)
	if err != nil {
		// Feedback dir doesn't exist yet, that's OK
		return nil
	}

	// Load all feedback events
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(feedbackDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var rawEvents []map[string]interface{}
		if err := json.Unmarshal(data, &rawEvents); err != nil {
			continue
		}

		// Convert feedback events to model feedback format
		for _, event := range rawEvents {
			fb := ms.convertFeedbackEvent(event)
			if fb.ModelID != "" && fb.Action != "" {
				ms.feedback = append(ms.feedback, fb)
			}
		}
	}

	return nil
}

// convertFeedbackEvent converts a feedback event to ModelFeedback
func (ms *ModelSelector) convertFeedbackEvent(event map[string]interface{}) ModelFeedback {
	fb := ModelFeedback{}

	// Extract model
	if val, ok := event["model"].(string); ok {
		fb.ModelID = val
	}

	// Map agent to action
	if agent, ok := event["agent"].(string); ok {
		switch strings.ToLower(agent) {
		case "editor":
			fb.Action = ActionEdit
		case "reviewer", "validator":
			fb.Action = ActionReview
		case "planner":
			fb.Action = ActionPlan
		case "research":
			fb.Action = ActionResearch
		}
	}

	// Determine success from sentiment
	if sentiment, ok := event["sentiment"].(string); ok {
		fb.Success = sentiment == "good"
	}

	// Try to extract language from task
	fb.Language = "unknown"
	if task, ok := event["task"].(string); ok {
		taskLower := strings.ToLower(task)
		if strings.Contains(taskLower, ".go") {
			fb.Language = "go"
		} else if strings.Contains(taskLower, ".py") {
			fb.Language = "python"
		} else if strings.Contains(taskLower, ".ts") || strings.Contains(taskLower, ".js") {
			fb.Language = "typescript"
		} else if strings.Contains(taskLower, ".ex") || strings.Contains(taskLower, ".exs") {
			fb.Language = "elixir"
		}

		// Determine complexity
		fb.Complexity = "simple"
		if strings.Contains(taskLower, "refactor") ||
			strings.Contains(taskLower, "reorganize") ||
			strings.Contains(taskLower, "complex") {
			fb.Complexity = "complex"
		}
	}

	return fb
}

func (ms *ModelSelector) loadUsage() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	usagePath := filepath.Join(home, ".gptcode", "usage.json")
	data, err := os.ReadFile(usagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &ms.usage); err != nil {
		return err
	}

	return nil
}

func (ms *ModelSelector) saveUsage() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	usagePath := filepath.Join(home, ".gptcode", "usage.json")
	data, err := json.MarshalIndent(ms.usage, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(usagePath, data, 0644)
}

func (ms *ModelSelector) RecordUsage(backend, model string, success bool, errorMsg string) {
	ms.RecordUsageWithTokens(backend, model, success, errorMsg, 0, 0, 0)
}

func (ms *ModelSelector) RecordUsageWithTokens(backend, model string, success bool, errorMsg string, inputTokens, outputTokens, cachedTokens int) {
	today := time.Now().Format("2006-01-02")
	if ms.usage[today] == nil {
		ms.usage[today] = make(map[string]ModelUsage)
	}

	key := backend + "/" + model
	usage := ms.usage[today][key]
	usage.Requests++
	usage.InputTokens += inputTokens
	usage.OutputTokens += outputTokens
	usage.CachedTokens += cachedTokens
	if !success {
		usage.LastError = errorMsg
	}
	ms.usage[today][key] = usage

	if err := ms.saveUsage(); err != nil && os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[WARN] Failed to save usage: %v\n", err)
	}
}

func (ms *ModelSelector) getTodayUsage(backend, model string) ModelUsage {
	today := time.Now().Format("2006-01-02")
	if ms.usage[today] == nil {
		return ModelUsage{}
	}

	key := backend + "/" + model
	return ms.usage[today][key]
}

func (ms *ModelSelector) SelectModel(action ActionType, language string, complexity string) (backend string, model string, err error) {
	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[MODEL_SELECTOR] SelectModel called: action=%s lang=%s complexity=%s\n",
			action, language, complexity)
	}

	mode := ms.setup.Defaults.Mode
	defaultBackend := ms.setup.Defaults.Backend

	type scoredModel struct {
		backend string
		model   string
		score   float64
	}
	var scored []scoredModel

	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[MODEL_SELECTOR] SelectModel action=%s lang=%s mode=%s defaultBackend=%s\n",
			action, language, mode, defaultBackend)
		fmt.Fprintf(os.Stderr, "[MODEL_SELECTOR] Catalog has %d backends\n", len(ms.catalog))
	}

	for backend, models := range ms.catalog {
		if mode == "local" && backend != "ollama" {
			continue
		}
		if mode == "cloud" && backend == "ollama" {
			continue
		}

		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[MODEL_SELECTOR] Checking backend=%s with %d models\n", backend, len(models))
		}

		for _, modelInfo := range models {
			score := ms.scoreModel(modelInfo, action, language, complexity)
			if score > 0 {
				// Boost score for default backend to prioritize it
				if backend == defaultBackend {
					score += 100
				}
				scored = append(scored, scoredModel{
					backend: backend,
					model:   modelInfo.ID,
					score:   score,
				})
			}
		}
	}

	if len(scored) == 0 {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[MODEL_SELECTOR] No models scored > 0 for action=%s lang=%s\n", action, language)
			fmt.Fprintf(os.Stderr, "[MODEL_SELECTOR] Checked %d models in catalog\n", len(ms.catalog[defaultBackend]))
		}
		return "", "", fmt.Errorf("no suitable model found for action=%s lang=%s", action, language)
	}

	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	best := scored[0]

	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[MODEL_SELECTOR] Action=%s Lang=%s -> %s/%s (score=%.2f)\n",
			action, language, best.backend, best.model, best.score)
	}

	return best.backend, best.model, nil
}

func (ms *ModelSelector) scoreModel(model ModelInfo, action ActionType, language string, complexity string) float64 {
	if action == ActionEdit || action == ActionReview {
		if !model.Capabilities.SupportsFileOperations {
			if os.Getenv("CHUCHU_DEBUG") == "1" && action == ActionEdit {
				fmt.Fprintf(os.Stderr, "[MODEL_SELECTOR] Model %s/%s rejected: no file operations support\n", model.Backend, model.ID)
			}
			return 0
		}
	}

	score := 100.0

	usage := ms.getTodayUsage(model.Backend, model.ID)
	if model.RateLimitDaily > 0 {
		utilization := float64(usage.Requests) / float64(model.RateLimitDaily)
		score -= utilization * 50
		if utilization >= 0.9 {
			score -= 50
		}
	}

	if usage.LastError != "" {
		score -= 30
	}

	if model.CostPer1M > 0 {
		score -= (model.CostPer1M / 10.0) * 30
	}

	if model.ContextWindow > 0 {
		score += (float64(model.ContextWindow) / 100000.0) * 10
	}

	if model.TokensPerSec > 0 {
		score += (float64(model.TokensPerSec) / 100.0) * 5
	}

	for _, fb := range ms.feedback {
		if fb.ModelID != model.ID {
			continue
		}
		if fb.Action == action && strings.EqualFold(fb.Language, language) {
			if fb.Success {
				score += 20
			} else {
				score -= 40
			}
		}
	}

	if complexity == "simple" {
		if strings.Contains(strings.ToLower(model.ID), "instant") ||
			strings.Contains(strings.ToLower(model.ID), "8b") ||
			strings.Contains(strings.ToLower(model.ID), "3b") {
			score += 15
		}
	}

	if complexity == "complex" || complexity == "multistep" {
		if strings.Contains(strings.ToLower(model.ID), "70b") ||
			strings.Contains(strings.ToLower(model.ID), "large") {
			score += 20
		}
	}

	if action == ActionEdit || action == ActionReview {
		if strings.Contains(strings.ToLower(model.ID), "coder") ||
			strings.Contains(strings.ToLower(model.ID), "code") {
			score += 25
		}
	}

	return score
}
