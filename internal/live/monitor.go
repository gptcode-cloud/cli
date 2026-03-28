package live

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// AgentUsage holds real usage data parsed from agent log files
type AgentUsage struct {
	Agent     string  // e.g. "Antigravity", "Cursor", "Windsurf"
	Model     string  // e.g. "claude-opus-4-6-thinking", "gemini-2.5-pro"
	Provider  string  // e.g. "anthropic", "google"
	APICalls  int     // total API calls today
	FirstCall string  // timestamp of first call
	LastCall  string  // timestamp of last call
	Exhausted bool    // hit rate limit today?
	ResetIn   string  // reset time if exhausted
	QuotaUsed float64 // 0.0-1.0 estimated
}

// MonitorResult aggregates usage across all agents
type MonitorResult struct {
	Agents        []AgentUsage
	ScanTime      time.Time
	TotalAPICalls int
}

// Monitor scans installed AI agents and returns real usage data
func Monitor() (*MonitorResult, error) {
	result := &MonitorResult{
		ScanTime: time.Now(),
	}

	// Scan each known agent
	scanners := []func() []AgentUsage{
		scanAntigravity,
		scanCursor,
		scanWindsurf,
	}

	for _, scanner := range scanners {
		usages := scanner()
		for _, u := range usages {
			result.Agents = append(result.Agents, u)
			result.TotalAPICalls += u.APICalls
		}
	}

	// Sort by API calls descending
	sort.Slice(result.Agents, func(i, j int) bool {
		return result.Agents[i].APICalls > result.Agents[j].APICalls
	})

	return result, nil
}

// --- Antigravity Scanner ---

func antigravityLogDir() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "Antigravity", "logs")
	}
	// Linux
	return filepath.Join(home, ".config", "Antigravity", "logs")
}

func scanAntigravity() []AgentUsage {
	logDir := antigravityLogDir()
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return nil
	}

	// Find the most recent Antigravity.log files
	var logFiles []string
	filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "Antigravity.log" && strings.Contains(path, "google.antigravity") {
			// Only consider logs modified today
			if info.ModTime().Format("2006-01-02") == time.Now().Format("2006-01-02") {
				logFiles = append(logFiles, path)
			}
		}
		return nil
	})

	if len(logFiles) == 0 {
		return nil
	}

	// Parse all today's log files and aggregate by model
	modelUsage := make(map[string]*AgentUsage)
	today := time.Now().Format("2006-01-02")

	for _, logFile := range logFiles {
		parseAntigravityLog(logFile, today, modelUsage)
	}

	var result []AgentUsage
	for _, u := range modelUsage {
		// Estimate quota based on known RPD limits
		u.QuotaUsed = estimateQuota(u.Model, u.APICalls)
		result = append(result, *u)
	}

	return result
}

func parseAntigravityLog(logFile, today string, usage map[string]*AgentUsage) {
	f, err := os.Open(logFile)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 256*1024)
	scanner.Buffer(buf, 1024*1024)

	currentModel := ""
	// Track exhaustion events with timestamps to compute rolling window
	type exhaustionEvent struct {
		model     string
		timestamp string // when the 429 happened
		resetStr  string // e.g. "2h49m19s"
	}
	var exhaustions []exhaustionEvent

	// First pass: collect all lines, detect models, and track exhaustion events
	type apiCall struct {
		model     string
		timestamp string
	}
	var calls []apiCall

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, today) {
			continue
		}

		// Detect model from log lines (error messages, planner requests, etc.)
		if strings.Contains(line, "claude-opus") {
			currentModel = "claude-opus-4-6-thinking"
		} else if strings.Contains(line, "claude-sonnet") {
			currentModel = "claude-sonnet-4"
		} else if strings.Contains(line, "gemini-2.5-pro") || strings.Contains(line, "gemini_2.5_pro") {
			currentModel = "gemini-2.5-pro"
		} else if strings.Contains(line, "gemini-2.5-flash") || strings.Contains(line, "gemini_2.5_flash") {
			currentModel = "gemini-2.5-flash"
		} else if strings.Contains(line, "gemini-2.0-flash") || strings.Contains(line, "gemini_2.0_flash") {
			currentModel = "gemini-2.0-flash"
		}

		// Count API calls
		if strings.Contains(line, "streamGenerateContent") {
			ts := ""
			if len(line) >= 23 {
				ts = line[:23]
			}
			model := currentModel
			if model == "" {
				model = "unknown"
			}
			calls = append(calls, apiCall{model: model, timestamp: ts})
		}

		// Track exhaustion events
		if strings.Contains(line, "RESOURCE_EXHAUSTED") || strings.Contains(line, "code 429") {
			ts := ""
			if len(line) >= 23 {
				ts = line[:23]
			}
			resetStr := ""
			if idx := strings.Index(line, "reset after "); idx > 0 {
				rest := line[idx+len("reset after "):]
				// Extract duration like "2h49m19s" or "0s"
				endIdx := strings.Index(rest, ":")
				if endIdx > 0 {
					resetStr = strings.TrimRight(rest[:endIdx], ".")
				}
			}
			exhaustions = append(exhaustions, exhaustionEvent{
				model:     currentModel,
				timestamp: ts,
				resetStr:  resetStr,
			})
		}
	}

	// Find the last reset time per model
	lastResetTime := make(map[string]string) // model -> timestamp when quota reset
	for _, evt := range exhaustions {
		model := evt.model
		if model == "" {
			model = "unknown"
		}
		// Parse the reset duration and calculate when quota reset
		if evt.resetStr != "" && evt.resetStr != "0s" {
			dur, err := time.ParseDuration(evt.resetStr)
			if err == nil {
				// Parse the exhaustion timestamp
				t, err := time.Parse("2006-01-02 15:04:05", evt.timestamp[:19])
				if err == nil {
					resetAt := t.Add(dur)
					resetTs := resetAt.Format("2006-01-02 15:04:05.000")
					// Keep the latest reset time
					if existing, ok := lastResetTime[model]; !ok || resetTs > existing {
						lastResetTime[model] = resetTs
					}
				}
			}
		}
	}

	// If unknown model has calls but we detected a model from exhaustion, merge them
	// The Antigravity log doesn't repeat the model name on normal API calls
	detectedModel := currentModel
	if detectedModel == "" && len(exhaustions) > 0 {
		detectedModel = exhaustions[len(exhaustions)-1].model
	}

	// Count calls per model, only SINCE the last reset
	for _, call := range calls {
		model := call.model
		// Merge "unknown" into the detected model if we have one
		if model == "unknown" && detectedModel != "" {
			model = detectedModel
		}

		// Check if this call is after the last reset for this model
		resetTs, hasReset := lastResetTime[model]
		if hasReset && call.timestamp <= resetTs {
			// This call was before the quota reset — skip it for current window
			// But still track total daily calls
			if _, ok := usage[model]; !ok {
				usage[model] = &AgentUsage{
					Agent:    "Antigravity",
					Model:    model,
					Provider: detectProvider(model),
				}
			}
			// Don't count these in APICalls (which represents current window)
			continue
		}

		if _, ok := usage[model]; !ok {
			usage[model] = &AgentUsage{
				Agent:     "Antigravity",
				Model:     model,
				Provider:  detectProvider(model),
				FirstCall: call.timestamp,
			}
		}
		usage[model].APICalls++
		usage[model].LastCall = call.timestamp
		if usage[model].FirstCall == "" {
			usage[model].FirstCall = call.timestamp
		}
	}

	// Mark models that had exhaustion today (for informational display)
	for _, evt := range exhaustions {
		model := evt.model
		if model == "" {
			model = detectedModel
		}
		if model == "" {
			continue
		}
		if u, ok := usage[model]; ok {
			// Only mark as currently exhausted if it hasn't reset yet
			if resetTs, hasReset := lastResetTime[model]; hasReset {
				now := time.Now().Format("2006-01-02 15:04:05.000")
				if now < resetTs {
					u.Exhausted = true
					u.ResetIn = evt.resetStr
				}
				// If reset is in the past, quota recovered — don't mark exhausted
			} else {
				u.Exhausted = true
			}
		}
	}
}

func detectProvider(model string) string {
	if strings.Contains(model, "claude") {
		return "anthropic"
	}
	if strings.Contains(model, "gemini") {
		return "google"
	}
	if strings.Contains(model, "gpt") || strings.Contains(model, "o3") || strings.Contains(model, "o4") {
		return "openai"
	}
	return "unknown"
}

// Known daily request limits per model
var knownRPD = map[string]int{
	"claude-opus-4-6-thinking": 800,   // Google One AI Premium estimate
	"claude-sonnet-4":          2000,  // Google One AI Premium estimate
	"gemini-2.5-pro":           25,    // Free tier
	"gemini-2.5-flash":         500,   // Free tier
	"gemini-2.0-flash":         1500,  // Free tier
	"gpt-4o":                   10000, // Paid tier
}

func estimateQuota(model string, apiCalls int) float64 {
	rpd, ok := knownRPD[model]
	if !ok {
		// Try prefix match
		for key, val := range knownRPD {
			if strings.HasPrefix(model, key) {
				rpd = val
				ok = true
				break
			}
		}
	}
	if !ok || rpd == 0 {
		return 0
	}
	quota := float64(apiCalls) / float64(rpd)
	if quota > 1.0 {
		quota = 1.0
	}
	return quota
}

// DisplayName returns the dashboard-friendly model name
func (u *AgentUsage) DisplayName() string {
	switch {
	case strings.Contains(u.Model, "claude-opus"):
		return "Claude Opus 4.6 (Thinking)"
	case strings.Contains(u.Model, "claude-sonnet"):
		return "Claude Sonnet 4"
	case strings.Contains(u.Model, "gemini-2.5-pro"):
		return "Gemini 2.5 Pro"
	case strings.Contains(u.Model, "gemini-2.5-flash"):
		return "Gemini 2.5 Flash"
	case strings.Contains(u.Model, "gemini-2.0-flash"):
		return "Gemini 2.0 Flash"
	default:
		return u.Model
	}
}

// --- Cursor Scanner ---

func cursorLogDir() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "Cursor", "logs")
	}
	return filepath.Join(home, ".config", "Cursor", "logs")
}

func scanCursor() []AgentUsage {
	logDir := cursorLogDir()
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return nil
	}

	// Cursor logs: look for API call patterns
	var logFiles []string
	filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, ".log") && info.ModTime().Format("2006-01-02") == time.Now().Format("2006-01-02") {
			logFiles = append(logFiles, path)
		}
		return nil
	})

	if len(logFiles) == 0 {
		return nil
	}

	apiCalls := 0
	today := time.Now().Format("2006-01-02")
	for _, logFile := range logFiles {
		apiCalls += countPatternInLog(logFile, today, []string{
			"api.openai.com",
			"api.anthropic.com",
			"generativelanguage.googleapis.com",
			"chat/completions",
		})
	}

	if apiCalls == 0 {
		return nil
	}

	return []AgentUsage{{
		Agent:     "Cursor",
		Model:     "unknown",
		Provider:  "mixed",
		APICalls:  apiCalls,
		QuotaUsed: 0, // Cursor has its own quota system
	}}
}

// --- Windsurf Scanner ---

func windsurfLogDir() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "Windsurf", "logs")
	}
	return filepath.Join(home, ".config", "Windsurf", "logs")
}

func scanWindsurf() []AgentUsage {
	logDir := windsurfLogDir()
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return nil
	}

	var logFiles []string
	filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, ".log") && info.ModTime().Format("2006-01-02") == time.Now().Format("2006-01-02") {
			logFiles = append(logFiles, path)
		}
		return nil
	})

	if len(logFiles) == 0 {
		return nil
	}

	apiCalls := 0
	today := time.Now().Format("2006-01-02")
	for _, logFile := range logFiles {
		apiCalls += countPatternInLog(logFile, today, []string{
			"api.openai.com",
			"api.anthropic.com",
			"completions",
		})
	}

	if apiCalls == 0 {
		return nil
	}

	return []AgentUsage{{
		Agent:     "Windsurf",
		Model:     "unknown",
		Provider:  "mixed",
		APICalls:  apiCalls,
		QuotaUsed: 0,
	}}
}

// --- Shared Utilities ---

func countPatternInLog(logFile, today string, patterns []string) int {
	f, err := os.Open(logFile)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, today) {
			continue
		}
		for _, pattern := range patterns {
			if strings.Contains(line, pattern) {
				count++
				break
			}
		}
	}
	return count
}

// ReportToLive sends all discovered agent usage to the Live Dashboard
func (r *MonitorResult) ReportToLive(reportConfig *ReportConfig) error {
	for _, usage := range r.Agents {
		agentID := fmt.Sprintf("monitor-%s-%s", strings.ToLower(usage.Agent), strings.ReplaceAll(usage.Model, " ", "-"))

		// Connect
		payload := map[string]interface{}{
			"agent_id":   agentID,
			"engine":     "monitor",
			"type":       usage.Agent,
			"model":      usage.DisplayName(),
			"provider":   usage.Provider,
			"context":    "local",
			"task":       fmt.Sprintf("%d API calls today", usage.APICalls),
			"hostname":   reportConfig.Hostname,
			"quota_used": usage.QuotaUsed,
		}

		if err := reportConfig.post("/api/report/connect", payload); err != nil {
			log.Printf("⚠️  Failed to report %s: %v", usage.Agent, err)
			continue
		}

		// Send a step with the summary
		desc := fmt.Sprintf("%d API calls (%s → %s)", usage.APICalls, usage.FirstCall, usage.LastCall)
		if usage.Exhausted {
			desc += " ⚠️ HIT RATE LIMIT"
		}

		stepPayload := map[string]interface{}{
			"agent_id":    agentID,
			"description": desc,
			"type":        "step",
			"quota_used":  usage.QuotaUsed,
		}
		reportConfig.post("/api/report/step", stepPayload)

		log.Printf("📊 %s (%s): %d calls, %.0f%% used", usage.Agent, usage.DisplayName(), usage.APICalls, usage.QuotaUsed*100)
	}

	return nil
}
