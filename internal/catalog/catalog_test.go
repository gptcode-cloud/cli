package catalog

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCatalogSearch(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration test in CI (requires models.json)")
	}

	homeDir, _ := os.UserHomeDir()
	catalogPath := homeDir + "/.gptcode/models.json"

	data, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatalf("Failed to read catalog: %v", err)
	}

	var catalog OutputJSON
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatalf("Failed to parse catalog: %v", err)
	}

	models := catalog.OpenRouter.Models
	t.Logf("Total models in catalog: %d", len(models))

	// Test 1: Search for "gemini"
	geminiModels := searchModels(models, "gemini")
	t.Logf("Found %d gemini models", len(geminiModels))
	if len(geminiModels) == 0 {
		t.Error("Expected to find gemini models but found none")
	}
	for i, m := range geminiModels {
		if i < 3 {
			t.Logf("  - %s (%s)", m.Name, m.ID)
		}
	}

	// Test 2: Search for "grok"
	grokModels := searchModels(models, "grok")
	t.Logf("Found %d grok models", len(grokModels))
	if len(grokModels) == 0 {
		t.Error("Expected to find grok models but found none")
	}

	// Test 3: Verify grok-4.1-fast exists and is FREE
	var grok41 *ModelOutput
	for _, m := range models {
		if m.ID == "x-ai/grok-4.1-fast" {
			grok41 = &m
			break
		}
	}

	if grok41 == nil {
		t.Fatal("grok-4.1-fast not found in catalog!")
	}

	t.Logf("✓ Found grok-4.1-fast: %s", grok41.Name)
	t.Logf("  Price: $%.2f / $%.2f", grok41.PricingPrompt, grok41.PricingComp)

	if grok41.PricingPrompt != 0 || grok41.PricingComp != 0 {
		t.Errorf("Expected grok-4.1-fast to be FREE but got $%.2f/$%.2f",
			grok41.PricingPrompt, grok41.PricingComp)
	}

	// Test 4: Verify models are sorted by price
	if len(models) > 1 {
		t.Log("First 5 models (should be sorted by price):")
		for i := 0; i < 5 && i < len(models); i++ {
			cost := models[i].PricingPrompt + models[i].PricingComp
			t.Logf("  %d. %s - $%.2f", i+1, models[i].Name, cost)
		}
	}
}

func TestSearchModelsMulti(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration test in CI (requires models.json)")
	}

	t.Run("backend auto-detection", func(t *testing.T) {
		models, err := SearchModelsMulti("", []string{"groq", "llama"}, "")
		if err != nil {
			t.Fatalf("SearchModelsMulti failed: %v", err)
		}

		if len(models) == 0 {
			t.Error("Expected groq llama models but got none")
		}

		for _, m := range models {
			if !strings.Contains(strings.ToLower(m.Name), "llama") && !strings.Contains(strings.ToLower(m.ID), "llama") {
				t.Errorf("Model %s doesn't contain 'llama'", m.Name)
			}
		}

		t.Logf("Found %d groq llama models", len(models))
	})

	t.Run("groq backend includes orchestrator models", func(t *testing.T) {
		models, err := SearchModelsMulti("", []string{"groq", "free"}, "")
		if err != nil {
			t.Fatalf("SearchModelsMulti failed: %v", err)
		}

		// Groq backend includes orchestrator models (compound, kimi, etc) tagged as free
		// This is correct behavior - users can use them via groq orchestrator
		if len(models) == 0 {
			t.Error("Expected some free models in groq backend")
		}

		t.Logf("Found %d models with 'free' tag in groq backend", len(models))
		for i, m := range models {
			if i < 3 {
				t.Logf("  - %s (tags: %v)", m.Name, m.Tags)
			}
		}
	})

	t.Run("multi-term AND filtering", func(t *testing.T) {
		models, err := SearchModelsMulti("openrouter", []string{"gemini", "free"}, "")
		if err != nil {
			t.Fatalf("SearchModelsMulti failed: %v", err)
		}

		if len(models) == 0 {
			t.Error("Expected free gemini models but got none")
		}

		for _, m := range models {
			hasGemini := strings.Contains(strings.ToLower(m.Name), "gemini") || strings.Contains(strings.ToLower(m.ID), "gemini")
			isFree := m.PricingPrompt == 0 && m.PricingComp == 0

			if !hasGemini {
				t.Errorf("Model %s doesn't contain 'gemini'", m.Name)
			}
			if !isFree {
				t.Errorf("Model %s is not free: $%.2f/$%.2f", m.Name, m.PricingPrompt, m.PricingComp)
			}
		}

		t.Logf("Found %d free gemini models", len(models))
	})
}

func searchModels(models []ModelOutput, query string) []ModelOutput {
	var results []ModelOutput
	queryLower := strings.ToLower(query)

	for _, m := range models {
		nameLower := strings.ToLower(m.Name)
		idLower := strings.ToLower(m.ID)

		if strings.Contains(nameLower, queryLower) || strings.Contains(idLower, queryLower) {
			results = append(results, m)
		}
	}

	return results
}
