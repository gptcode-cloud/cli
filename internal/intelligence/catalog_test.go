package intelligence

import (
	"testing"
)

func TestNewModelCatalog(t *testing.T) {
	catalog := NewModelCatalog()

	if catalog == nil {
		t.Fatal("NewModelCatalog returned nil")
	}

	if len(catalog.Models) == 0 {
		t.Error("Expected catalog to have models, got empty map")
	}

	// Check some known models exist
	expectedModels := []string{
		"openrouter/moonshotai/kimi-k2:free",
		"groq/moonshotai/kimi-k2-instruct-0905",
		"openai/gpt-4-turbo",
		"ollama/qwen3-coder",
	}

	for _, key := range expectedModels {
		if _, ok := catalog.Models[key]; !ok {
			t.Errorf("Expected model %s not found in catalog", key)
		}
	}
}

func TestGetModelsForAgent(t *testing.T) {
	catalog := NewModelCatalog()

	tests := []struct {
		name          string
		agent         string
		expectAtLeast int
	}{
		{"editor agent", "editor", 7},
		{"nonexistent agent", "nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models := catalog.GetModelsForAgent(tt.agent)
			if len(models) < tt.expectAtLeast {
				t.Errorf("GetModelsForAgent(%s) returned %d models, expected at least %d",
					tt.agent, len(models), tt.expectAtLeast)
			}
		})
	}
}

func TestGetModelInfo(t *testing.T) {
	catalog := NewModelCatalog()

	tests := []struct {
		name              string
		backend           string
		model             string
		expectExists      bool
		expectFunctions   bool
		expectFreeOrCheap bool
	}{
		{
			name:              "known free model",
			backend:           "openrouter",
			model:             "moonshotai/kimi-k2:free",
			expectExists:      true,
			expectFunctions:   true,
			expectFreeOrCheap: true,
		},
		{
			name:              "known paid model",
			backend:           "openai",
			model:             "gpt-4-turbo",
			expectExists:      true,
			expectFunctions:   true,
			expectFreeOrCheap: false,
		},
		{
			name:              "unknown model with fallback",
			backend:           "unknown",
			model:             "unknown-model",
			expectExists:      false,
			expectFunctions:   false,
			expectFreeOrCheap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := catalog.GetModelInfo(tt.backend, tt.model)

			if info.Backend != tt.backend {
				t.Errorf("Expected backend %s, got %s", tt.backend, info.Backend)
			}

			if info.Name != tt.model {
				t.Errorf("Expected model name %s, got %s", tt.model, info.Name)
			}

			if tt.expectExists {
				if info.SupportsFunctions != tt.expectFunctions {
					t.Errorf("Expected SupportsFunctions=%v, got %v", tt.expectFunctions, info.SupportsFunctions)
				}

				if tt.expectFreeOrCheap && info.CostPer1M > 1.0 {
					t.Errorf("Expected free or cheap model, got cost %f", info.CostPer1M)
				}

				if !tt.expectFreeOrCheap && info.CostPer1M == 0.0 {
					t.Errorf("Expected paid model, got cost 0")
				}
			} else {
				// Fallback should return reasonable defaults
				if info.SpeedTPS != 300 {
					t.Errorf("Expected fallback speed 300, got %d", info.SpeedTPS)
				}
			}
		})
	}
}

func TestModelCatalogEditorModels(t *testing.T) {
	catalog := NewModelCatalog()
	editorModels := catalog.GetModelsForAgent("editor")

	// All editor models should support functions
	for _, model := range editorModels {
		if !model.SupportsFunctions {
			t.Errorf("Editor model %s/%s should support functions", model.Backend, model.Name)
		}

		if model.CostPer1M < 0 {
			t.Errorf("Model %s/%s has negative cost", model.Backend, model.Name)
		}

		if model.SpeedTPS <= 0 {
			t.Errorf("Model %s/%s has invalid speed: %d", model.Backend, model.Name, model.SpeedTPS)
		}
	}
}
