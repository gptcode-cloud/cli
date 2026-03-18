package autonomous

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type LearningEngine struct {
	cwd        string
	memoryFile string
	strategies map[string]*Strategy
	feedback   []Feedback
}

type Strategy struct {
	ID          string    `json:"id"`
	TaskPattern string    `json:"task_pattern"`
	Actions     []string  `json:"actions"`
	SuccessRate float64   `json:"success_rate"`
	Uses        int       `json:"uses"`
	LastUsed    time.Time `json:"last_used"`
	CreatedAt   time.Time `json:"created_at"`
}

type Feedback struct {
	Task      string    `json:"task"`
	Success   bool      `json:"success"`
	Actions   []string  `json:"actions"`
	Timestamp time.Time `json:"timestamp"`
	Notes     string    `json:"notes,omitempty"`
}

func NewLearningEngine(cwd string) *LearningEngine {
	le := &LearningEngine{
		cwd:        cwd,
		memoryFile: filepath.Join(cwd, ".gptcode", "learned_strategies.json"),
		strategies: make(map[string]*Strategy),
		feedback:   []Feedback{},
	}
	le.load()
	return le
}

func (le *LearningEngine) RecordAttempt(task string, actions []string, success bool, notes string) {
	feedback := Feedback{
		Task:      task,
		Success:   success,
		Actions:   actions,
		Timestamp: time.Now(),
		Notes:     notes,
	}
	le.feedback = append(le.feedback, feedback)

	// Update strategy success rate
	pattern := le.extractPattern(task)
	if strategy, ok := le.strategies[pattern]; ok {
		// Update existing strategy
		total := float64(strategy.Uses)
		strategy.SuccessRate = (strategy.SuccessRate*total + boolToFloat(success)) / (total + 1)
		strategy.Uses++
		strategy.LastUsed = time.Now()
	} else {
		// Create new strategy
		le.strategies[pattern] = &Strategy{
			ID:          pattern,
			TaskPattern: pattern,
			Actions:     actions,
			SuccessRate: boolToFloat(success),
			Uses:        1,
			LastUsed:    time.Now(),
			CreatedAt:   time.Now(),
		}
	}

	le.save()
}

func (le *LearningEngine) GetBestStrategy(task string) []string {
	// Find matching strategies
	var matches []*Strategy
	for _, s := range le.strategies {
		if le.matchesPattern(task, s.TaskPattern) {
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		return nil
	}

	// Sort by success rate and recency
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].SuccessRate != matches[j].SuccessRate {
			return matches[i].SuccessRate > matches[j].SuccessRate
		}
		return matches[i].LastUsed.After(matches[j].LastUsed)
	})

	return matches[0].Actions
}

func (le *LearningEngine) GetSuggestions(task string) []string {
	// Get top 3 strategies for similar tasks
	var strategies []*Strategy
	for _, s := range le.strategies {
		if le.matchesPattern(task, s.TaskPattern) && s.SuccessRate > 0.5 {
			strategies = append(strategies, s)
		}
	}

	sort.Slice(strategies, func(i, j int) bool {
		return strategies[i].SuccessRate > strategies[j].SuccessRate
	})

	var suggestions []string
	for i := 0; i < 3 && i < len(strategies); i++ {
		suggestions = append(suggestions, fmt.Sprintf("[%s%% success] %s",
			fmt.Sprintf("%.0f", strategies[i].SuccessRate*100),
			strings.Join(strategies[i].Actions, " → ")))
	}

	return suggestions
}

func (le *LearningEngine) extractPattern(task string) string {
	// Extract key pattern from task
	task = strings.ToLower(task)
	task = strings.TrimSpace(task)

	// Remove common words
	words := strings.Fields(task)
	var keywords []string
	for _, w := range words {
		if len(w) > 3 && !isCommonWord(w) {
			keywords = append(keywords, w)
		}
	}

	return strings.Join(keywords, " ")
}

func (le *LearningEngine) matchesPattern(task, pattern string) bool {
	task = strings.ToLower(task)
	pattern = strings.ToLower(pattern)

	taskWords := strings.Fields(task)
	patternWords := strings.Fields(pattern)

	matchCount := 0
	for _, pw := range patternWords {
		for _, tw := range taskWords {
			if strings.Contains(tw, pw) || strings.Contains(pw, tw) {
				matchCount++
				break
			}
		}
	}

	threshold := float64(len(patternWords)) * 0.5
	return float64(matchCount) >= threshold
}

func (le *LearningEngine) load() {
	data, err := os.ReadFile(le.memoryFile)
	if err != nil {
		return
	}

	var loaded struct {
		Strategies map[string]*Strategy `json:"strategies"`
		Feedback   []Feedback           `json:"feedback"`
	}

	if err := json.Unmarshal(data, &loaded); err != nil {
		return
	}

	le.strategies = loaded.Strategies
	le.feedback = loaded.Feedback
}

func (le *LearningEngine) save() error {
	dir := filepath.Dir(le.memoryFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data := struct {
		Strategies map[string]*Strategy `json:"strategies"`
		Feedback   []Feedback           `json:"feedback"`
	}{
		Strategies: le.strategies,
		Feedback:   le.feedback,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(le.memoryFile, jsonData, 0644)
}

func (le *LearningEngine) GetStats() string {
	total := len(le.strategies)
	avgRate := 0.0
	mostUsed := 0

	for _, s := range le.strategies {
		avgRate += s.SuccessRate
		if s.Uses > mostUsed {
			mostUsed = s.Uses
		}
	}

	if total > 0 {
		avgRate /= float64(total)
	}

	return fmt.Sprintf("Strategies: %d, Avg Success Rate: %.0f%%, Most Used: %d times",
		total, avgRate*100, mostUsed)
}

func (le *LearningEngine) Clear() {
	le.strategies = make(map[string]*Strategy)
	le.feedback = []Feedback{}
	le.save()
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func isCommonWord(word string) bool {
	common := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "shall": true, "can": true,
		"need": true, "dare": true, "ought": true, "used": true,
		"to": true, "of": true, "in": true, "for": true, "on": true,
		"with": true, "at": true, "by": true, "from": true, "as": true,
		"into": true, "through": true, "during": true, "before": true, "after": true,
		"above": true, "below": true, "between": true, "under": true, "again": true,
		"further": true, "then": true, "once": true, "here": true, "there": true,
		"when": true, "where": true, "why": true, "how": true, "all": true,
		"each": true, "few": true, "more": true, "most": true, "other": true,
		"some": true, "such": true, "no": true, "nor": true, "not": true,
		"only": true, "own": true, "same": true, "so": true, "than": true,
		"too": true, "very": true, "just": true, "and": true, "but": true,
		"if": true, "or": true, "because": true, "until": true, "while": true,
		"that": true, "this": true, "these": true, "those": true, "what": true,
		"which": true, "who": true, "whom": true, "fix": true, "add": true,
		"create": true, "update": true, "delete": true, "remove": true,
	}

	return common[strings.ToLower(word)]
}

type ExperienceCache struct {
	maxAge      time.Duration
	experiences map[string]*Experience
}

type Experience struct {
	Task       string
	Solution   string
	Success    bool
	Complexity int
	CreatedAt  time.Time
}

func NewExperienceCache(maxAge time.Duration) *ExperienceCache {
	return &ExperienceCache{
		maxAge:      maxAge,
		experiences: make(map[string]*Experience),
	}
}

func (ec *ExperienceCache) Get(task string) *Experience {
	exp, ok := ec.experiences[task]
	if !ok {
		return nil
	}

	if time.Since(exp.CreatedAt) > ec.maxAge {
		delete(ec.experiences, task)
		return nil
	}

	return exp
}

func (ec *ExperienceCache) Store(task string, exp *Experience) {
	exp.CreatedAt = time.Now()
	ec.experiences[task] = exp
}

func (ec *ExperienceCache) Clear() {
	ec.experiences = make(map[string]*Experience)
}
