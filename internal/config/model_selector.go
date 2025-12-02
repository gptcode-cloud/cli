package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Capabilities ModelCapabilities `json:"capabilities"`
	Backend      string            // inferred from catalog structure
}

type ModelFeedback struct {
	ModelID    string     `json:"model_id"`
	Action     ActionType `json:"action"`
	Language   string     `json:"language"`
	Success    bool       `json:"success"`
	Complexity string     `json:"complexity"` // simple, complex, multistep
}

type ModelSelector struct {
	catalog  map[string][]ModelInfo // backend -> models
	feedback []ModelFeedback
	setup    *Setup
}

func NewModelSelector(setup *Setup) (*ModelSelector, error) {
	selector := &ModelSelector{
		catalog:  make(map[string][]ModelInfo),
		feedback: []ModelFeedback{},
		setup:    setup,
	}

	if err := selector.loadCatalog(); err != nil {
		return nil, fmt.Errorf("failed to load catalog: %w", err)
	}

	if err := selector.loadFeedback(); err != nil {
		// Non-critical, just log
		fmt.Fprintf(os.Stderr, "[WARN] Could not load feedback: %v\n", err)
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

			// Parse capabilities
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
	feedbackDir := filepath.Join(home, ".chuchu", "feedback")
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

// SelectModel escolhe o melhor modelo para a ação
func (ms *ModelSelector) SelectModel(action ActionType, language string, complexity string) (backend string, model string, err error) {
	// Determine preferred backend order based on mode
	preferredBackend := ms.setup.Defaults.Backend
	
	// Use mode to determine backend preference
	mode := ms.setup.Defaults.Mode
	if mode == "local" {
		preferredBackend = "ollama"
	} else if mode == "cloud" || mode == "" {
		// Cloud mode: prefer groq or openrouter
		if preferredBackend == "" || preferredBackend == "ollama" {
			preferredBackend = "groq"
		}
	}
	
	backendOrder := []string{preferredBackend}
	
	// Add fallbacks based on mode
	if mode == "local" || mode == "" {
		for _, backend := range []string{"ollama", "groq", "openrouter"} {
			if backend != preferredBackend {
				backendOrder = append(backendOrder, backend)
			}
		}
	} else {
		for _, backend := range []string{"groq", "openrouter", "ollama"} {
			if backend != preferredBackend {
				backendOrder = append(backendOrder, backend)
			}
		}
	}

	// Score each available model
	type scoredModel struct {
		backend string
		model   string
		score   float64
	}
	var scored []scoredModel

	for _, backend := range backendOrder {
		models, ok := ms.catalog[backend]
		if !ok {
			continue
		}

		for _, modelInfo := range models {
			score := ms.scoreModel(modelInfo, action, language, complexity)
			if score > 0 {
				scored = append(scored, scoredModel{
					backend: backend,
					model:   modelInfo.ID,
					score:   score,
				})
			}
		}
	}

	if len(scored) == 0 {
		return "", "", fmt.Errorf("no suitable model found for action=%s lang=%s", action, language)
	}

	// Sort by score (highest first)
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
	score := 0.0

	// Base requirements
	if action == ActionEdit || action == ActionReview {
		if !model.Capabilities.SupportsFileOperations {
			return 0 // Hard requirement
		}
		score += 50
	}

	// Feedback-based scoring
	for _, fb := range ms.feedback {
		if fb.ModelID != model.ID {
			continue
		}
		if fb.Action == action && strings.EqualFold(fb.Language, language) {
			if fb.Success {
				score += 30
			} else {
				score -= 50 // Strong penalty for known failures
			}
		}
	}

	// Prefer faster models for simple tasks
	if complexity == "simple" {
		if strings.Contains(strings.ToLower(model.ID), "instant") ||
			strings.Contains(strings.ToLower(model.ID), "8b") {
			score += 10
		}
	}

	// Prefer larger models for complex tasks
	if complexity == "complex" || complexity == "multistep" {
		if strings.Contains(strings.ToLower(model.ID), "70b") ||
			strings.Contains(strings.ToLower(model.ID), "large") {
			score += 15
		}
	}

	// Code-specialized models bonus for edit/review
	if action == ActionEdit || action == ActionReview {
		if strings.Contains(strings.ToLower(model.ID), "coder") ||
			strings.Contains(strings.ToLower(model.ID), "code") {
			score += 20
		}
	}

	// Backend preference (prefer configured backend)
	if model.Backend == ms.setup.Defaults.Backend {
		score += 5
	}

	return score
}
