package intelligence

import (
	"fmt"
	"chuchu/internal/config"
)

type ModelRecommendation struct {
	Backend      string
	Model        string
	Reason       string
	Confidence   float64
}

var modelsWithFunctionCalling = map[string]map[string][]string{
	"openrouter": {
		"editor": []string{
			"moonshotai/kimi-k2:free",
			"google/gemini-2.0-flash-exp:free",
			"anthropic/claude-3.5-sonnet",
		},
	},
	"groq": {
		"editor": []string{
			"moonshotai/kimi-k2-instruct-0905",
		},
	},
	"openai": {
		"editor": []string{
			"gpt-4-turbo",
			"gpt-4",
		},
	},
	"ollama": {
		"editor": []string{
			"qwen3-coder",
		},
	},
}

func RecommendModelForRetry(setup *config.Setup, agentType string, failedBackend string, failedModel string, task string) ([]ModelRecommendation, error) {
	recommendations := make([]ModelRecommendation, 0)

	history, err := GetRecentModelPerformance("", 100)
	if err != nil {
		history = []ModelSuccess{}
	}

	historyMap := make(map[string]ModelSuccess)
	for _, h := range history {
		key := h.Backend + "/" + h.Model
		historyMap[key] = h
	}

	for backend, agents := range modelsWithFunctionCalling {
		models, exists := agents[agentType]
		if !exists {
			continue
		}

		backendCfg, backendExists := setup.Backend[backend]
		if !backendExists {
			continue
		}

		for _, model := range models {
			if backend == failedBackend && model == failedModel {
				continue
			}

			key := backend + "/" + model
			h, hasHistory := historyMap[key]

			confidence := 0.5
			reason := fmt.Sprintf("Known to support function calling")

			if hasHistory {
				if h.TotalTasks >= 3 {
					confidence = h.SuccessRate
					reason = fmt.Sprintf("Historical success rate: %.0f%% (%d tasks)", h.SuccessRate*100, h.TotalTasks)
				}
			}

			if backend == failedBackend {
				confidence *= 0.9
				reason += " (same backend)"
			}

			modelCfg, modelExists := backendCfg.Models[model]
			if !modelExists {
				modelCfg = model
			}

			recommendations = append(recommendations, ModelRecommendation{
				Backend:    backend,
				Model:      modelCfg,
				Reason:     reason,
				Confidence: confidence,
			})
		}
	}

	sortByConfidence(recommendations)

	return recommendations, nil
}

func sortByConfidence(recs []ModelRecommendation) {
	for i := 0; i < len(recs)-1; i++ {
		for j := i + 1; j < len(recs); j++ {
			if recs[j].Confidence > recs[i].Confidence {
				recs[i], recs[j] = recs[j], recs[i]
			}
		}
	}
}
