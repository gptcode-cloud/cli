package catalog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

const OpenRouterAPI = "https://openrouter.ai/api/v1/models"

type Pricing struct {
	Prompt     float64 `json:"-"`
	Completion float64 `json:"-"`
}

func (p *Pricing) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if promptStr, ok := raw["prompt"].(string); ok {
		if val, err := strconv.ParseFloat(promptStr, 64); err == nil {
			p.Prompt = val * 1_000_000
		}
	} else if promptFloat, ok := raw["prompt"].(float64); ok {
		p.Prompt = promptFloat * 1_000_000
	}

	if compStr, ok := raw["completion"].(string); ok {
		if val, err := strconv.ParseFloat(compStr, 64); err == nil {
			p.Completion = val * 1_000_000
		}
	} else if compFloat, ok := raw["completion"].(float64); ok {
		p.Completion = compFloat * 1_000_000
	}

	return nil
}

type Provider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ModelAPI struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	ContextWindow  int             `json:"context_window"`
	Pricing        Pricing         `json:"pricing"`
	TopProvider    Provider        `json:"top_provider"`
	Description    string          `json:"description"`
	Architecture   json.RawMessage `json:"architecture"`
	PerplexityRate *float64        `json:"perplexity_rate"`
	SupportsTools  bool            `json:"supports_tools"`
}

type APIResponse struct {
	Data []ModelAPI `json:"data"`
}

type ModelOutput struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Tags           []string `json:"tags"`
	RecommendedFor []string `json:"recommended_for"`
	ContextWindow  int      `json:"context_window"`
	PricingPrompt  float64  `json:"pricing_prompt_per_m_tokens"`
	PricingComp    float64  `json:"pricing_completion_per_m_tokens"`
}

type ProviderOutput struct {
	Models []ModelOutput `json:"models"`
}

type OutputJSON struct {
	Groq       ProviderOutput `json:"groq"`
	OpenRouter ProviderOutput `json:"openrouter"`
	Ollama     ProviderOutput `json:"ollama"`
	OpenAI     ProviderOutput `json:"openai"`
	DeepSeek   ProviderOutput `json:"deepseek"`
}

func FetchAndSave(outputPath string) error {
	models, err := fetchModels()
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	categorized := categorizeAndTagModels(models)

	if err := saveJSON(categorized, outputPath); err != nil {
		return fmt.Errorf("failed to save JSON: %w", err)
	}

	return nil
}

func fetchModels() ([]ModelAPI, error) {
	resp, err := http.Get(OpenRouterAPI)
	if err != nil {
		return nil, fmt.Errorf("falha na requisição HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code inesperado: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler o corpo da resposta: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON: %w", err)
	}

	return apiResp.Data, nil
}

func categorizeAndTagModels(models []ModelAPI) OutputJSON {
	output := OutputJSON{
		Groq:       ProviderOutput{Models: []ModelOutput{}},
		OpenRouter: ProviderOutput{Models: []ModelOutput{}},
		Ollama:     ProviderOutput{Models: []ModelOutput{}},
		OpenAI:     ProviderOutput{Models: []ModelOutput{}},
		DeepSeek:   ProviderOutput{Models: []ModelOutput{}},
	}

	for _, m := range models {
		modelOutput := ModelOutput{
			ID:            m.ID,
			Name:          m.Name,
			ContextWindow: m.ContextWindow,
			PricingPrompt: m.Pricing.Prompt,
			PricingComp:   m.Pricing.Completion,
			Tags:          inferTags(m),
			RecommendedFor: inferRecommendedFor(m),
		}

		output.OpenRouter.Models = append(output.OpenRouter.Models, modelOutput)

		parts := strings.Split(m.ID, "/")
		providerPrefix := ""
		if len(parts) > 1 {
			providerPrefix = strings.ToLower(parts[0])
		}

		switch providerPrefix {
		case "openai":
			output.OpenAI.Models = append(output.OpenAI.Models, modelOutput)
		case "deepseek":
			output.DeepSeek.Models = append(output.DeepSeek.Models, modelOutput)
		case "meta-llama", "mistralai", "qwen", "google":
			if m.Pricing.Prompt == 0 && m.Pricing.Completion == 0 {
				output.Ollama.Models = append(output.Ollama.Models, modelOutput)
			}
		}

		if m.Pricing.Prompt <= 0.5 && m.Pricing.Completion <= 1.5 {
			output.Groq.Models = append(output.Groq.Models, modelOutput)
		}
	}

	sortOutputModels(&output)

	return output
}

func inferTags(m ModelAPI) []string {
	tags := []string{}
	cost := m.Pricing.Prompt + m.Pricing.Completion

	if cost == 0 {
		tags = append(tags, "free")
	} else if cost < 1.0 { // Ex: < $1 per 1M tokens
		tags = append(tags, "cheap")
	} else if cost > 10.0 { // Ex: > $10 per 1M tokens
		tags = append(tags, "premium")
	}

	if m.TopProvider.ID == "groq" {
		tags = append(tags, "low-latency")
		if strings.Contains(m.ID, "8b") {
			tags = append(tags, "very-fast")
		} else {
			tags = append(tags, "fast")
		}
	} else if strings.Contains(m.Name, "Haiku") || strings.Contains(m.Name, "Mini") || strings.Contains(m.ID, "8b") {
		tags = append(tags, "fast")
	}

	if m.ContextWindow > 100000 {
		tags = append(tags, "128k-context")
	} else if m.ContextWindow > 30000 {
		tags = append(tags, "large-context")
	}

	if strings.Contains(m.Description, "web search") || strings.Contains(m.Name, "Sonar") || m.SupportsTools {
		tags = append(tags, "web-search")
	}
	if strings.Contains(m.Description, "coding") || strings.Contains(m.Name, "Coder") || strings.Contains(m.ID, "codellama") {
		tags = append(tags, "coding")
	}
	if strings.Contains(m.ID, "llama-guard") {
		tags = append(tags, "moderation")
	}

  // heuristics example for "hig quality"
	if m.PerplexityRate != nil && *m.PerplexityRate < 2.0 {
		tags = append(tags, "best-quality")
	} else if !strings.Contains(m.Name, "Mini") && !strings.Contains(m.Name, "Haiku") {
		tags = append(tags, "versatile")
	}
    
    // Garantir unicidade das tags
    uniqueTags := make(map[string]bool)
    finalTags := []string{}
    for _, t := range tags {
        if !uniqueTags[t] {
            uniqueTags[t] = true
            finalTags = append(finalTags, t)
        }
    }
	return finalTags
}

func inferRecommendedFor(m ModelAPI) []string {
	recs := []string{}

	if strings.Contains(m.Name, "Instant") || strings.Contains(m.Name, "Mini") || strings.Contains(m.Name, "Haiku") || strings.Contains(m.ID, "8b") {
		recs = append(recs, "router")
	} else {
		recs = append(recs, "query", "editor")
	}

	if strings.Contains(m.Description, "coding") || strings.Contains(m.Name, "Coder") {
		recs = append(recs, "editor")
	}
	if strings.Contains(m.Description, "web search") || strings.Contains(m.Name, "Sonar") {
		recs = append(recs, "research")
	}
	return recs
}

func sortOutputModels(output *OutputJSON) {
	sort.Slice(output.Groq.Models, func(i, j int) bool {
		return output.Groq.Models[i].Name < output.Groq.Models[j].Name
	})
	sort.Slice(output.OpenRouter.Models, func(i, j int) bool {
		return output.OpenRouter.Models[i].Name < output.OpenRouter.Models[j].Name
	})
	sort.Slice(output.Ollama.Models, func(i, j int) bool {
		return output.Ollama.Models[i].Name < output.Ollama.Models[j].Name
	})
	sort.Slice(output.OpenAI.Models, func(i, j int) bool {
		return output.OpenAI.Models[i].Name < output.OpenAI.Models[j].Name
	})
	sort.Slice(output.DeepSeek.Models, func(i, j int) bool {
		return output.DeepSeek.Models[i].Name < output.DeepSeek.Models[j].Name
	})
}

func saveJSON(data OutputJSON, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating json file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("error writing JSON content: %w", err)
	}
	return nil
}
