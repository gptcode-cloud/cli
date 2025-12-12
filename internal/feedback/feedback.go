package feedback

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Sentiment string

const (
	SentimentGood Sentiment = "good"
	SentimentBad  Sentiment = "bad"
)

type EventKind string

type Event struct {
	Timestamp       time.Time         `json:"timestamp"`
	Sentiment       Sentiment         `json:"sentiment"`
	Backend         string            `json:"backend"`
	Model           string            `json:"model"`
	Agent           string            `json:"agent"`
	Task            string            `json:"task,omitempty"`
	Context         string            `json:"context,omitempty"`
	WrongResponse   string            `json:"wrong_response,omitempty"`
	CorrectResponse string            `json:"correct_response,omitempty"`
	Source          string            `json:"source,omitempty"`
	Kind            EventKind         `json:"kind,omitempty"`
	Files           []string          `json:"files,omitempty"`
	DiffPath        string            `json:"diff_path,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

func GetFeedbackDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gptcode", "feedback")
}

func ensureFeedbackDir() error {
	dir := GetFeedbackDir()
	return os.MkdirAll(dir, 0755)
}

func Record(event Event) error {
	if err := ensureFeedbackDir(); err != nil {
		return fmt.Errorf("failed to create feedback dir: %w", err)
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	filename := fmt.Sprintf("%s.json", event.Timestamp.Format("2006-01-02"))
	path := filepath.Join(GetFeedbackDir(), filename)

	var events []Event
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &events); err != nil {
			return fmt.Errorf("failed to parse existing feedback: %w", err)
		}
	}

	events = append(events, event)

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal feedback: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write feedback: %w", err)
	}

	return nil
}

func LoadAll() ([]Event, error) {
	dir := GetFeedbackDir()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, fmt.Errorf("failed to read feedback dir: %w", err)
	}

	var allEvents []Event
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var events []Event
		if err := json.Unmarshal(data, &events); err != nil {
			continue
		}

		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

type Stats struct {
	TotalEvents  int                     `json:"total_events"`
	GoodCount    int                     `json:"good_count"`
	BadCount     int                     `json:"bad_count"`
	ByBackend    map[string]BackendStats `json:"by_backend"`
	ByModel      map[string]ModelStats   `json:"by_model"`
	ByAgent      map[string]AgentStats   `json:"by_agent"`
	RecentEvents []Event                 `json:"recent_events,omitempty"`
}

type BackendStats struct {
	Good  int     `json:"good"`
	Bad   int     `json:"bad"`
	Total int     `json:"total"`
	Ratio float64 `json:"ratio"`
}

type ModelStats struct {
	Good    int     `json:"good"`
	Bad     int     `json:"bad"`
	Total   int     `json:"total"`
	Ratio   float64 `json:"ratio"`
	Backend string  `json:"backend"`
}

type AgentStats struct {
	Good  int     `json:"good"`
	Bad   int     `json:"bad"`
	Total int     `json:"total"`
	Ratio float64 `json:"ratio"`
}

// ConvertToModelFeedback converts feedback events to model selector format
func ConvertToModelFeedback(events []Event, language string) []map[string]interface{} {
	var result []map[string]interface{}

	for _, e := range events {
		// Map agent to action
		var action string
		switch strings.ToLower(e.Agent) {
		case "editor":
			action = "edit"
		case "reviewer", "validator":
			action = "review"
		case "planner":
			action = "plan"
		case "research":
			action = "research"
		default:
			// Skip events without clear action mapping
			continue
		}

		// Determine complexity from context/task
		complexity := "simple"
		if strings.Contains(strings.ToLower(e.Context), "complex") ||
			strings.Contains(strings.ToLower(e.Task), "refactor") ||
			strings.Contains(strings.ToLower(e.Task), "reorganize") {
			complexity = "complex"
		}

		// Detect language from task or use provided
		lang := language
		if strings.Contains(strings.ToLower(e.Task), ".go") {
			lang = "go"
		} else if strings.Contains(strings.ToLower(e.Task), ".py") {
			lang = "python"
		} else if strings.Contains(strings.ToLower(e.Task), ".ts") || strings.Contains(strings.ToLower(e.Task), ".js") {
			lang = "typescript"
		}

		fb := map[string]interface{}{
			"model_id":   e.Model,
			"action":     action,
			"language":   lang,
			"success":    e.Sentiment == SentimentGood,
			"complexity": complexity,
			"backend":    e.Backend,
			"timestamp":  e.Timestamp,
		}

		result = append(result, fb)
	}

	return result
}

func Analyze(events []Event) Stats {
	stats := Stats{
		TotalEvents: len(events),
		ByBackend:   make(map[string]BackendStats),
		ByModel:     make(map[string]ModelStats),
		ByAgent:     make(map[string]AgentStats),
	}

	for _, e := range events {
		switch e.Sentiment {
		case SentimentGood:
			stats.GoodCount++
		case SentimentBad:
			stats.BadCount++
		}

		if e.Backend != "" {
			bs := stats.ByBackend[e.Backend]
			bs.Total++
			switch e.Sentiment {
			case SentimentGood:
				bs.Good++
			case SentimentBad:
				bs.Bad++
			}
			if bs.Total > 0 {
				bs.Ratio = float64(bs.Good) / float64(bs.Total)
			}
			stats.ByBackend[e.Backend] = bs
		}

		if e.Model != "" {
			ms := stats.ByModel[e.Model]
			ms.Total++
			ms.Backend = e.Backend
			switch e.Sentiment {
			case SentimentGood:
				ms.Good++
			case SentimentBad:
				ms.Bad++
			}
			if ms.Total > 0 {
				ms.Ratio = float64(ms.Good) / float64(ms.Total)
			}
			stats.ByModel[e.Model] = ms
		}

		if e.Agent != "" {
			as := stats.ByAgent[e.Agent]
			as.Total++
			switch e.Sentiment {
			case SentimentGood:
				as.Good++
			case SentimentBad:
				as.Bad++
			}
			if as.Total > 0 {
				as.Ratio = float64(as.Good) / float64(as.Total)
			}
			stats.ByAgent[e.Agent] = as
		}
	}

	if len(events) > 10 {
		stats.RecentEvents = events[len(events)-10:]
	} else {
		stats.RecentEvents = events
	}

	return stats
}

// PromptForFeedback prompts user for feedback after task completion
// Returns: sentiment, correctResponse (if provided), shouldRecord
func PromptForFeedback() (Sentiment, string, bool) {
	fmt.Fprintf(os.Stderr, "\n\nWas this response helpful? [Y/n/e] ")
	fmt.Fprintf(os.Stderr, "\n  Y - Yes, good response\n")
	fmt.Fprintf(os.Stderr, "  n - No, bad response\n")
	fmt.Fprintf(os.Stderr, "  e - Edit correct response\n")
	fmt.Fprintf(os.Stderr, "  <enter> - Skip\n")
	fmt.Fprintf(os.Stderr, "> ")

	var input string
	fmt.Scanln(&input)

	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "y", "yes":
		return SentimentGood, "", true
	case "n", "no":
		return SentimentBad, "", true
	case "e", "edit":
		fmt.Fprintf(os.Stderr, "\nProvide the correct response:\n> ")
		var correctResponse string
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			correctResponse = scanner.Text()
		}
		if correctResponse != "" {
			fmt.Fprintf(os.Stderr, "\nThank you! This will improve future responses.\n")
			return SentimentBad, correctResponse, true
		}
		return SentimentBad, "", false
	default:
		return "", "", false
	}
}

func GetBestModels(agent string, minSamples int) []string {
	events, err := LoadAll()
	if err != nil {
		return nil
	}

	stats := Analyze(events)

	type modelRating struct {
		model string
		ratio float64
		total int
	}

	var candidates []modelRating
	for model, ms := range stats.ByModel {
		if ms.Total >= minSamples {
			candidates = append(candidates, modelRating{
				model: model,
				ratio: ms.Ratio,
				total: ms.Total,
			})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].ratio > candidates[i].ratio ||
				(candidates[j].ratio == candidates[i].ratio && candidates[j].total > candidates[i].total) {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	var best []string
	for _, c := range candidates {
		best = append(best, c.model)
	}

	return best
}
