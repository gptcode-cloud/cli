package intelligence

import (
	"fmt"
	"strings"
	"chuchu/internal/config"
)

type ModelRecommendation struct {
	Backend      string
	Model        string
	Reason       string
	Confidence   float64
	Score        float64
	Metrics      RecommendationMetrics
}

type RecommendationMetrics struct {
	SuccessRate   float64
	AvgLatencyMs  int64
	CostPer1M     float64
	Availability  float64
	SpeedTPS      int
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
	latencyMap := make(map[string]int64)
	for _, h := range history {
		key := h.Backend + "/" + h.Model
		historyMap[key] = h
		if h.AvgLatency > 0 {
			latencyMap[key] = h.AvgLatency
		}
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

			metrics := RecommendationMetrics{
				SuccessRate:  0.5,
				Availability: 1.0,
				CostPer1M:    getModelCost(backend, model),
				SpeedTPS:     getModelSpeed(backend, model),
			}

			if hasHistory && h.TotalTasks >= 3 {
				metrics.SuccessRate = h.SuccessRate
			}

			if latency, ok := latencyMap[key]; ok {
				metrics.AvgLatencyMs = latency
			}

			score := calculateScore(metrics, backend == failedBackend)
			confidence := metrics.SuccessRate

			reason := buildReason(metrics, hasHistory, h.TotalTasks)

			modelCfg, modelExists := backendCfg.Models[model]
			if !modelExists {
				modelCfg = model
			}

			recommendations = append(recommendations, ModelRecommendation{
				Backend:    backend,
				Model:      modelCfg,
				Reason:     reason,
				Confidence: confidence,
				Score:      score,
				Metrics:    metrics,
			})
		}
	}

	sortByScore(recommendations)

	return recommendations, nil
}

func calculateScore(m RecommendationMetrics, sameBackend bool) float64 {
	successWeight := 0.5
	speedWeight := 0.2
	costWeight := 0.2
	availWeight := 0.1

	successScore := m.SuccessRate

	speedScore := 0.5
	if m.SpeedTPS > 0 {
		speedScore = min(float64(m.SpeedTPS)/1000.0, 1.0)
	}
	if m.AvgLatencyMs > 0 {
		latencyScore := 1.0 - min(float64(m.AvgLatencyMs)/10000.0, 0.5)
		speedScore = (speedScore + latencyScore) / 2
	}

	costScore := 1.0
	if m.CostPer1M > 0 {
		costScore = 1.0 - min(m.CostPer1M/2.0, 0.8)
	}

	availScore := m.Availability

	score := successWeight*successScore +
		speedWeight*speedScore +
		costWeight*costScore +
		availWeight*availScore

	if sameBackend {
		score *= 0.95
	}

	return score
}

func buildReason(m RecommendationMetrics, hasHistory bool, totalTasks int) string {
	if hasHistory && totalTasks >= 3 {
		return fmt.Sprintf("Success: %.0f%% (%d tasks), Speed: %d TPS, Cost: $%.2f/1M",
			m.SuccessRate*100, totalTasks, m.SpeedTPS, m.CostPer1M)
	}
	return fmt.Sprintf("Known capable, Speed: %d TPS, Cost: $%.2f/1M",
		m.SpeedTPS, m.CostPer1M)
}

func getModelCost(backend, model string) float64 {
	if strings.Contains(model, ":free") || strings.Contains(backend, "ollama") {
		return 0.0
	}
	
	costs := map[string]float64{
		"gpt-4-turbo":                    0.01,
		"gpt-4":                          0.03,
		"claude-3.5-sonnet":              0.003,
		"moonshotai/kimi-k2-instruct":    0.002,
		"llama-3.3-70b-versatile":        0.00,
		"qwen3-coder":                    0.00,
	}
	
	for prefix, cost := range costs {
		if strings.Contains(model, prefix) {
			return cost
		}
	}
	
	return 0.001
}

func getModelSpeed(backend, model string) int {
	speeds := map[string]int{
		"groq":       500,
		"ollama":     200,
		"openrouter": 300,
		"openai":     400,
	}
	
	if speed, ok := speeds[backend]; ok {
		return speed
	}
	
	return 300
}

func sortByScore(recs []ModelRecommendation) {
	for i := 0; i < len(recs)-1; i++ {
		for j := i + 1; j < len(recs); j++ {
			if recs[j].Score > recs[i].Score {
				recs[i], recs[j] = recs[j], recs[i]
			}
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
