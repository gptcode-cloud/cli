package intelligence

// ModelCatalog centralizes model capabilities and metadata
type ModelCatalog struct {
	Models map[string]ModelInfo
}

type ModelInfo struct {
	Backend           string
	Name              string
	SupportsFunctions bool
	CostPer1M         float64  // Cost per 1M tokens (input+output avg)
	SpeedTPS          int      // avg tokens per second
	Agents            []string // which agents this model supports (editor, query, etc.)
}

// NewModelCatalog creates a catalog with default models
func NewModelCatalog() *ModelCatalog {
	return &ModelCatalog{
		Models: map[string]ModelInfo{
			// OpenRouter models
			"openrouter/moonshotai/kimi-k2:free": {
				Backend:           "openrouter",
				Name:              "moonshotai/kimi-k2:free",
				SupportsFunctions: true,
				CostPer1M:         0.0,
				SpeedTPS:          300,
				Agents:            []string{"editor"},
			},
			"openrouter/google/gemini-2.0-flash-exp:free": {
				Backend:           "openrouter",
				Name:              "google/gemini-2.0-flash-exp:free",
				SupportsFunctions: true,
				CostPer1M:         0.0,
				SpeedTPS:          300,
				Agents:            []string{"editor"},
			},
			"openrouter/anthropic/claude-3.5-sonnet": {
				Backend:           "openrouter",
				Name:              "anthropic/claude-3.5-sonnet",
				SupportsFunctions: true,
				CostPer1M:         3.0,
				SpeedTPS:          300,
				Agents:            []string{"editor"},
			},
			// Groq models
			"groq/moonshotai/kimi-k2-instruct-0905": {
				Backend:           "groq",
				Name:              "moonshotai/kimi-k2-instruct-0905",
				SupportsFunctions: true,
				CostPer1M:         2.0,
				SpeedTPS:          500,
				Agents:            []string{"editor"},
			},
			// OpenAI models
			"openai/gpt-4-turbo": {
				Backend:           "openai",
				Name:              "gpt-4-turbo",
				SupportsFunctions: true,
				CostPer1M:         10.0,
				SpeedTPS:          400,
				Agents:            []string{"editor"},
			},
			"openai/gpt-4": {
				Backend:           "openai",
				Name:              "gpt-4",
				SupportsFunctions: true,
				CostPer1M:         30.0,
				SpeedTPS:          400,
				Agents:            []string{"editor"},
			},
			// Ollama models
			"ollama/qwen3-coder": {
				Backend:           "ollama",
				Name:              "qwen3-coder",
				SupportsFunctions: true,
				CostPer1M:         0.0,
				SpeedTPS:          200,
				Agents:            []string{"editor"},
			},
		},
	}
}

// GetModelsForAgent returns models that support the given agent type
func (c *ModelCatalog) GetModelsForAgent(agent string) []ModelInfo {
	var models []ModelInfo
	for _, info := range c.Models {
		for _, a := range info.Agents {
			if a == agent {
				models = append(models, info)
				break
			}
		}
	}
	return models
}

// GetModelInfo returns info for a specific model, with fallback defaults
func (c *ModelCatalog) GetModelInfo(backend, model string) ModelInfo {
	key := backend + "/" + model
	if info, ok := c.Models[key]; ok {
		return info
	}
	// Fallback for unknown models
	return ModelInfo{
		Backend:           backend,
		Name:              model,
		SupportsFunctions: false,
		CostPer1M:         1.0,
		SpeedTPS:          300,
		Agents:            []string{},
	}
}
