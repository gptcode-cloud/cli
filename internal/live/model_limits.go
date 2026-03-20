package live

import "fmt"

// ModelLimits holds known limits for a model
type ModelLimits struct {
	ContextWindow int    `json:"context_window"` // max tokens
	MaxOutput     int    `json:"max_output"`     // max output tokens
	RPM           int    `json:"rpm"`            // requests per minute (free tier)
	TPM           int    `json:"tpm"`            // tokens per minute (free tier)
	RPD           int    `json:"rpd"`            // requests per day (free tier, 0 = unlimited)
	Provider      string `json:"provider"`
	Tier          string `json:"tier"` // "free", "paid", "enterprise"
}

// knownModels maps model name patterns to their limits
// Sources: official API docs as of March 2026
var knownModels = map[string]ModelLimits{
	// Google Gemini
	"gemini-2.5-pro": {
		ContextWindow: 1048576, MaxOutput: 65536,
		RPM: 5, TPM: 250000, RPD: 25,
		Provider: "google", Tier: "free",
	},
	"gemini-2.5-pro-preview": {
		ContextWindow: 1048576, MaxOutput: 65536,
		RPM: 5, TPM: 250000, RPD: 25,
		Provider: "google", Tier: "free",
	},
	"gemini-2.5-flash": {
		ContextWindow: 1048576, MaxOutput: 65536,
		RPM: 15, TPM: 1000000, RPD: 500,
		Provider: "google", Tier: "free",
	},
	"gemini-2.5-flash-preview": {
		ContextWindow: 1048576, MaxOutput: 65536,
		RPM: 15, TPM: 1000000, RPD: 500,
		Provider: "google", Tier: "free",
	},
	"gemini-2.0-flash": {
		ContextWindow: 1048576, MaxOutput: 8192,
		RPM: 15, TPM: 1000000, RPD: 1500,
		Provider: "google", Tier: "free",
	},

	// Anthropic Claude
	"claude-sonnet-4-20250514": {
		ContextWindow: 200000, MaxOutput: 64000,
		RPM: 50, TPM: 40000, RPD: 0,
		Provider: "anthropic", Tier: "paid",
	},
	"claude-3.5-sonnet": {
		ContextWindow: 200000, MaxOutput: 8192,
		RPM: 50, TPM: 40000, RPD: 0,
		Provider: "anthropic", Tier: "paid",
	},
	"claude-3-haiku": {
		ContextWindow: 200000, MaxOutput: 4096,
		RPM: 50, TPM: 50000, RPD: 0,
		Provider: "anthropic", Tier: "paid",
	},
	"claude-3-opus": {
		ContextWindow: 200000, MaxOutput: 4096,
		RPM: 50, TPM: 20000, RPD: 0,
		Provider: "anthropic", Tier: "paid",
	},

	// OpenAI
	"gpt-4o": {
		ContextWindow: 128000, MaxOutput: 16384,
		RPM: 500, TPM: 30000, RPD: 0,
		Provider: "openai", Tier: "paid",
	},
	"gpt-4o-mini": {
		ContextWindow: 128000, MaxOutput: 16384,
		RPM: 500, TPM: 200000, RPD: 0,
		Provider: "openai", Tier: "paid",
	},
	"gpt-4-turbo": {
		ContextWindow: 128000, MaxOutput: 4096,
		RPM: 500, TPM: 30000, RPD: 0,
		Provider: "openai", Tier: "paid",
	},
	"o3-mini": {
		ContextWindow: 200000, MaxOutput: 100000,
		RPM: 500, TPM: 200000, RPD: 0,
		Provider: "openai", Tier: "paid",
	},

	// DeepSeek
	"deepseek-chat": {
		ContextWindow: 64000, MaxOutput: 8192,
		RPM: 60, TPM: 1000000, RPD: 0,
		Provider: "deepseek", Tier: "paid",
	},
	"deepseek-reasoner": {
		ContextWindow: 64000, MaxOutput: 8192,
		RPM: 60, TPM: 1000000, RPD: 0,
		Provider: "deepseek", Tier: "paid",
	},

	// Groq
	"llama-3.3-70b-versatile": {
		ContextWindow: 128000, MaxOutput: 32768,
		RPM: 30, TPM: 6000, RPD: 14400,
		Provider: "groq", Tier: "free",
	},

	// OpenRouter (pass-through, varies)
	"openrouter/auto": {
		ContextWindow: 128000, MaxOutput: 16384,
		RPM: 200, TPM: 200000, RPD: 0,
		Provider: "openrouter", Tier: "paid",
	},
}

// GetModelLimits returns known limits for a model name
// It tries exact match first, then prefix matching
func GetModelLimits(model string) *ModelLimits {
	if model == "" {
		return nil
	}

	// Exact match
	if limits, ok := knownModels[model]; ok {
		return &limits
	}

	// Try prefix matching (e.g. "gemini-2.5-pro-exp-0827" matches "gemini-2.5-pro")
	bestMatch := ""
	for key := range knownModels {
		if len(key) > len(bestMatch) && len(model) >= len(key) && model[:len(key)] == key {
			bestMatch = key
		}
	}
	if bestMatch != "" {
		limits := knownModels[bestMatch]
		return &limits
	}

	return nil
}

// FormatContextWindow returns a human-readable context window string
func FormatContextWindow(tokens int) string {
	if tokens >= 1000000 {
		return formatFloat(float64(tokens)/1000000) + "M"
	}
	if tokens >= 1000 {
		return formatFloat(float64(tokens)/1000) + "K"
	}
	return formatInt(tokens)
}

func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return formatInt(int(f))
	}
	// One decimal place
	return fmt.Sprintf("%.1f", f)
}

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}
