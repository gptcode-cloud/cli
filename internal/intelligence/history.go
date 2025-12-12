package intelligence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type TaskExecution struct {
	Timestamp time.Time              `json:"timestamp"`
	Task      string                 `json:"task"`
	Backend   string                 `json:"backend"`
	Model     string                 `json:"model"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Features  map[string]interface{} `json:"features,omitempty"`
	LatencyMs int64                  `json:"latency_ms,omitempty"`
}

func RecordExecution(exec TaskExecution) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	historyPath := filepath.Join(home, ".gptcode", "task_execution_history.jsonl")

	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	exec.Timestamp = time.Now()
	data, err := json.Marshal(exec)
	if err != nil {
		return err
	}

	_, err = f.Write(append(data, '\n'))
	return err
}

type ModelSuccess struct {
	Backend     string
	Model       string
	SuccessRate float64
	TotalTasks  int
	AvgLatency  int64
}

func GetRecentModelPerformance(taskType string, limit int) ([]ModelSuccess, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	historyPath := filepath.Join(home, ".gptcode", "task_execution_history.jsonl")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ModelSuccess{}, nil
		}
		return nil, err
	}

	lines := make([]TaskExecution, 0)
	for _, line := range splitLines(data) {
		var exec TaskExecution
		if err := json.Unmarshal(line, &exec); err != nil {
			continue
		}
		lines = append(lines, exec)
	}

	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}

	stats := make(map[string]*struct {
		successes    int
		total        int
		backend      string
		model        string
		totalLatency int64
		latencyCount int
	})

	for _, exec := range lines {
		key := exec.Backend + "/" + exec.Model
		if stats[key] == nil {
			stats[key] = &struct {
				successes    int
				total        int
				backend      string
				model        string
				totalLatency int64
				latencyCount int
			}{backend: exec.Backend, model: exec.Model}
		}
		stats[key].total++
		if exec.Success {
			stats[key].successes++
		}
		if exec.LatencyMs > 0 {
			stats[key].totalLatency += exec.LatencyMs
			stats[key].latencyCount++
		}
	}

	results := make([]ModelSuccess, 0, len(stats))
	for _, s := range stats {
		avgLatency := int64(0)
		if s.latencyCount > 0 {
			avgLatency = s.totalLatency / int64(s.latencyCount)
		}
		results = append(results, ModelSuccess{
			Backend:     s.backend,
			Model:       s.model,
			SuccessRate: float64(s.successes) / float64(s.total),
			TotalTasks:  s.total,
			AvgLatency:  avgLatency,
		})
	}

	sortBySuccessRate(results)
	return results, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

func sortBySuccessRate(results []ModelSuccess) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].SuccessRate > results[i].SuccessRate {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
