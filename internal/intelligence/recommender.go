package intelligence

import (
	"chuchu/internal/config"
	"fmt"
)

type ModelRecommendation struct {
	Backend    string
	Model      string
	Reason     string
	Confidence float64
	Score      float64
	Metrics    RecommendationMetrics
}

type RecommendationMetrics struct {
	SuccessRate  float64
	AvgLatencyMs int64
	CostPer1M    float64
	Availability float64
	SpeedTPS     int
}

var DefaultCatalog = NewModelCatalog()

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

	// Use catalog instead of hardcoded map
	candidateModels := DefaultCatalog.GetModelsForAgent(agentType)

	for _, modelInfo := range candidateModels {
		backend := modelInfo.Backend
		model := modelInfo.Name

		if backend == failedBackend && model == failedModel {
			continue
		}

		backendCfg, backendExists := setup.Backend[backend]
		if !backendExists {
			continue
		}

		key := backend + "/" + model
		h, hasHistory := historyMap[key]

		metrics := RecommendationMetrics{
			SuccessRate:  0.5,
			Availability: 1.0,
			CostPer1M:    modelInfo.CostPer1M,
			SpeedTPS:     modelInfo.SpeedTPS,
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
