package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	OpenRouterAPI = "https://openrouter.ai/api/v1/models"
	OllamaAPI     = "http://localhost:11434/api/tags"
	OpenAIAPI     = "https://api.openai.com/v1/models"
	GroqAPI       = "https://api.groq.com/openai/v1/models"
	AnthropicAPI  = "https://api.anthropic.com/v1/models"
	CohereAPI     = "https://api.cohere.ai/v1/models"
)

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

type TopProviderInfo struct {
	ContextLength      int  `json:"context_length"`
	MaxCompletionTokens int  `json:"max_completion_tokens"`
	IsModerated        bool `json:"is_moderated"`
}

type ModelAPI struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	ContextWindow  int             `json:"context_window"`
	Pricing        Pricing         `json:"pricing"`
	TopProvider    TopProviderInfo `json:"top_provider"`
	Description    string          `json:"description"`
	Architecture   json.RawMessage `json:"architecture"`
	PerplexityRate *float64        `json:"perplexity_rate"`
	SupportsTools  bool            `json:"supports_tools"`
}

type APIResponse struct {
	Data []ModelAPI `json:"data"`
}

type OllamaModel struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
	Details   struct {
		Family        string   `json:"family"`
		ParameterSize string   `json:"parameter_size"`
		Format        string   `json:"format"`
	} `json:"details"`
}

type OllamaResponse struct {
	Models []OllamaModel `json:"models"`
}

type ModelSource struct {
	Models   []ModelAPI
	Provider string
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

func FetchAndSave(outputPath string, apiKeys map[string]string) error {
	sources := []ModelSource{}

	ollamaModels, err := fetchOllamaModels()
	if err == nil {
		sources = append(sources, ModelSource{Models: ollamaModels, Provider: "ollama"})
	}

	openRouterModels, err := fetchOpenRouterModels()
	if err == nil {
		sources = append(sources, ModelSource{Models: openRouterModels, Provider: "openrouter"})
	}

	if apiKey, ok := apiKeys["groq"]; ok && apiKey != "" {
		groqModels, err := fetchGroqModels(apiKey)
		if err == nil {
			sources = append(sources, ModelSource{Models: groqModels, Provider: "groq"})
		}
	}

	if apiKey, ok := apiKeys["openai"]; ok && apiKey != "" {
		openAIModels, err := fetchOpenAIModels(apiKey)
		if err == nil {
			sources = append(sources, ModelSource{Models: openAIModels, Provider: "openai"})
		}
	}

	if apiKey, ok := apiKeys["anthropic"]; ok && apiKey != "" {
		anthropicModels, err := fetchAnthropicModels(apiKey)
		if err == nil {
			sources = append(sources, ModelSource{Models: anthropicModels, Provider: "anthropic"})
		}
	}

	if apiKey, ok := apiKeys["cohere"]; ok && apiKey != "" {
		cohereModels, err := fetchCohereModels(apiKey)
		if err == nil {
			sources = append(sources, ModelSource{Models: cohereModels, Provider: "cohere"})
		}
	}

	categorized := categorizeAndTagModels(sources)

	if err := saveJSON(categorized, outputPath); err != nil {
		return fmt.Errorf("failed to save JSON: %w", err)
	}

	return nil
}

func fetchOllamaModels() ([]ModelAPI, error) {
	resp, err := http.Get(OllamaAPI)
	if err != nil {
		return nil, fmt.Errorf("ollama not available: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ollama response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse ollama JSON: %w", err)
	}

	models := make([]ModelAPI, 0, len(ollamaResp.Models))
	for _, m := range ollamaResp.Models {
		modelAPI := ModelAPI{
			ID:            m.Name,
			Name:          m.Model,
			ContextWindow: 8192,
			Pricing:       Pricing{Prompt: 0, Completion: 0},
			Description:   fmt.Sprintf("Local Ollama model - %s", m.Details.ParameterSize),
		}
		models = append(models, modelAPI)
	}

	return models, nil
}

func fetchOpenRouterModels() ([]ModelAPI, error) {
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

func fetchGroqModels(apiKey string) ([]ModelAPI, error) {
	req, _ := http.NewRequest("GET", GroqAPI, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq API status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	pricingMap, err := scrapeGroqPricing()
	if err != nil {
		return apiResp.Data, nil
	}

	for i := range apiResp.Data {
		if apiResp.Data[i].Name == "" {
			apiResp.Data[i].Name = apiResp.Data[i].ID
		}
		
		if pricing, ok := pricingMap[apiResp.Data[i].ID]; ok {
			apiResp.Data[i].Pricing = pricing
		}
	}

	return apiResp.Data, nil
}

func scrapeGroqPricing() (map[string]Pricing, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	var modelsData []map[string]interface{}
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://groq.com/pricing/"),
		chromedp.Sleep(5*time.Second),
		chromedp.Evaluate(`
			(() => {
				const tryNowLinks = Array.from(document.querySelectorAll('a'))
					.filter(a => a.innerText.trim() === 'Try Now' && a.href.includes('?model='));
				
				return tryNowLinks.map(link => {
					const modelID = new URL(link.href).searchParams.get('model');
					
					let container = link;
					for (let i = 0; i < 4; i++) {
						if (container.parentElement) {
							container = container.parentElement;
						}
					}
					
					const allText = container.innerText || '';
					const allLines = allText.split('\\n');
					
					const tryNowIdx = allLines.findIndex(line => line.trim() === 'Try Now' && allText.includes(modelID));
					if (tryNowIdx < 0) {
						return { id: modelID, text: allText.substring(0, 500) };
					}
					
					const startIdx = Math.max(0, tryNowIdx - 10);
					const relevantLines = allLines.slice(startIdx, tryNowIdx + 1);
					
					return {
						id: modelID,
						text: relevantLines.join('\\n')
					};
				}).filter(m => m.id && m.text);
			})()
		`, &modelsData),
	)
	if err != nil {
		return nil, err
	}

	pricingMap := make(map[string]Pricing)
	inputPriceRegex := regexp.MustCompile(`Input Token Price[^$]*\$(\d+(?:\.\d+)?)`)
	outputPriceRegex := regexp.MustCompile(`Output Token Price[^$]*\$(\d+(?:\.\d+)?)`)

	for _, data := range modelsData {
		modelID, ok := data["id"].(string)
		if !ok {
			continue
		}
		
		text, ok := data["text"].(string)
		if !ok {
			continue
		}

		inputMatch := inputPriceRegex.FindStringSubmatch(text)
		outputMatch := outputPriceRegex.FindStringSubmatch(text)

		if len(inputMatch) > 1 && len(outputMatch) > 1 {
			promptPrice, _ := strconv.ParseFloat(inputMatch[1], 64)
			completionPrice, _ := strconv.ParseFloat(outputMatch[1], 64)

			pricingMap[modelID] = Pricing{
				Prompt:     promptPrice,
				Completion: completionPrice,
			}
		}
	}

	return pricingMap, nil
}

func fetchOpenAIModels(apiKey string) ([]ModelAPI, error) {
	req, _ := http.NewRequest("GET", OpenAIAPI, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func fetchAnthropicModels(apiKey string) ([]ModelAPI, error) {
	req, _ := http.NewRequest("GET", AnthropicAPI, nil)
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func fetchCohereModels(apiKey string) ([]ModelAPI, error) {
	req, _ := http.NewRequest("GET", CohereAPI, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cohere status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func categorizeAndTagModels(sources []ModelSource) OutputJSON {
	output := OutputJSON{
		Groq:       ProviderOutput{Models: []ModelOutput{}},
		OpenRouter: ProviderOutput{Models: []ModelOutput{}},
		Ollama:     ProviderOutput{Models: []ModelOutput{}},
		OpenAI:     ProviderOutput{Models: []ModelOutput{}},
		DeepSeek:   ProviderOutput{Models: []ModelOutput{}},
	}

	openRouterMap := make(map[string]ModelAPI)
	for _, source := range sources {
		if source.Provider == "openrouter" {
			for _, m := range source.Models {
				openRouterMap[m.ID] = m
			}
			break
		}
	}

	for _, source := range sources {
		for _, m := range source.Models {
			modelOutput := ModelOutput{
				ID:             m.ID,
				Name:           m.Name,
				ContextWindow:  m.ContextWindow,
				PricingPrompt:  m.Pricing.Prompt,
				PricingComp:    m.Pricing.Completion,
				Tags:           inferTags(m),
				RecommendedFor: inferRecommendedFor(m),
			}

			switch source.Provider {
			case "ollama":
				output.Ollama.Models = append(output.Ollama.Models, modelOutput)
			case "groq":
				output.Groq.Models = append(output.Groq.Models, modelOutput)
			case "openrouter":
				output.OpenRouter.Models = append(output.OpenRouter.Models, modelOutput)
				
				parts := strings.Split(m.ID, "/")
				if len(parts) > 1 {
					providerPrefix := strings.ToLower(parts[0])
					switch providerPrefix {
					case "openai":
						output.OpenAI.Models = append(output.OpenAI.Models, modelOutput)
					case "deepseek":
						output.DeepSeek.Models = append(output.DeepSeek.Models, modelOutput)
					}
				}
			case "openai":
				output.OpenAI.Models = append(output.OpenAI.Models, modelOutput)
			case "anthropic":
				output.OpenRouter.Models = append(output.OpenRouter.Models, modelOutput)
			case "cohere":
				output.OpenRouter.Models = append(output.OpenRouter.Models, modelOutput)
			case "deepseek":
				output.DeepSeek.Models = append(output.DeepSeek.Models, modelOutput)
			}
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

	if strings.Contains(m.Name, "Haiku") || strings.Contains(m.Name, "Mini") || strings.Contains(m.ID, "8b") {
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
