package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type EnhancedWorkflow struct {
	AutonomousWorkflow
	learning      *LearningEngine
	telemetry     *Telemetry
	budgetManager *BudgetManager
	selfHealer    *SelfHealer
	loopDetector  *LoopDetector
	refactorCoord *RefactorCoordinator
	cwd           string
}

type BudgetManager struct {
	mu           sync.Mutex
	dailyLimit   float64
	sessionLimit float64
	spentToday   float64
	spentSession float64
	alerts       []BudgetAlert
	budgetFile   string
}

type BudgetAlert struct {
	Level     string    `json:"level"`
	Threshold float64   `json:"threshold"`
	Current   float64   `json:"current"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type LoopDetector struct {
	mu               sync.Mutex
	actions          []string
	maxHistory       int
	windowSize       int
	similarityThresh float64
}

type RefactorCoordinator struct {
	mu      sync.Mutex
	changes []FileChange
	backups map[string][]byte
	phase   string
}

type FileChange struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Hash    string `json:"hash"`
}

func NewEnhancedWorkflow(cwd string) *EnhancedWorkflow {
	learning := NewLearningEngine(cwd)
	telemetry := NewTelemetry("")
	budget := NewBudgetManager(cwd)
	selfHealer := NewSelfHealer(cwd)
	loopDetector := NewLoopDetector(50, 5)
	refactorCoord := NewRefactorCoordinator()

	ew := &EnhancedWorkflow{
		cwd:           cwd,
		learning:      learning,
		telemetry:     telemetry,
		budgetManager: budget,
		selfHealer:    selfHealer,
		loopDetector:  loopDetector,
		refactorCoord: refactorCoord,
	}

	ew.AutonomousWorkflow = *NewAutonomousWorkflow(cwd)
	return ew
}

func NewBudgetManager(cwd string) *BudgetManager {
	bm := &BudgetManager{
		dailyLimit:   10.0,
		sessionLimit: 2.0,
		budgetFile:   filepath.Join(cwd, ".gptcode", "budget.json"),
	}
	bm.load()
	return bm
}

func (bm *BudgetManager) RecordCost(cost float64, task string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.spentSession += cost
	bm.spentToday += cost

	if cost > bm.sessionLimit*0.5 && len(bm.alerts) == 0 {
		bm.alerts = append(bm.alerts, BudgetAlert{
			Level:     "warning",
			Threshold: bm.sessionLimit * 0.5,
			Current:   bm.spentSession,
			Message:   fmt.Sprintf("Half of session budget used ($%.2f/$%.2f)", bm.spentSession, bm.sessionLimit),
			Timestamp: time.Now(),
		})
	}

	if bm.spentSession >= bm.sessionLimit {
		bm.alerts = append(bm.alerts, BudgetAlert{
			Level:     "critical",
			Threshold: bm.sessionLimit,
			Current:   bm.spentSession,
			Message:   fmt.Sprintf("Session budget exceeded! ($%.2f/$%.2f)", bm.spentSession, bm.sessionLimit),
			Timestamp: time.Now(),
		})
	}

	bm.save()
}

func (bm *BudgetManager) CanProceed() (bool, string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.spentSession >= bm.sessionLimit {
		return false, fmt.Sprintf("Session budget exceeded (%.2f/%.2f)", bm.spentSession, bm.sessionLimit)
	}

	if bm.spentToday >= bm.dailyLimit {
		return false, fmt.Sprintf("Daily budget exceeded (%.2f/%.2f)", bm.spentToday, bm.dailyLimit)
	}

	return true, ""
}

func (bm *BudgetManager) SetLimits(daily, session float64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.dailyLimit = daily
	bm.sessionLimit = session
}

func (bm *BudgetManager) GetAlerts() []BudgetAlert {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return bm.alerts
}

func (bm *BudgetManager) ClearAlerts() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.alerts = []BudgetAlert{}
}

func (bm *BudgetManager) save() {
	data := struct {
		SpentToday   float64 `json:"spent_today"`
		SpentSession float64 `json:"spent_session"`
		Date         string  `json:"date"`
	}{
		SpentToday:   bm.spentToday,
		SpentSession: bm.spentSession,
		Date:         time.Now().Format("2006-01-02"),
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(bm.budgetFile, jsonData, 0644)
}

func (bm *BudgetManager) load() {
	data, err := os.ReadFile(bm.budgetFile)
	if err != nil {
		return
	}

	var loaded struct {
		SpentToday   float64 `json:"spent_today"`
		SpentSession float64 `json:"spent_session"`
		Date         string  `json:"date"`
	}

	if err := json.Unmarshal(data, &loaded); err != nil {
		return
	}

	today := time.Now().Format("2006-01-02")
	if loaded.Date == today {
		bm.spentToday = loaded.SpentToday
	}
	bm.spentSession = loaded.SpentSession
}

func NewLoopDetector(maxHistory, windowSize int) *LoopDetector {
	return &LoopDetector{
		maxHistory:       maxHistory,
		windowSize:       windowSize,
		similarityThresh: 0.8,
	}
}

func (ld *LoopDetector) Record(action string) bool {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	ld.actions = append(ld.actions, action)

	if len(ld.actions) > ld.maxHistory {
		ld.actions = ld.actions[len(ld.actions)-ld.maxHistory:]
	}

	return ld.DetectLoop()
}

func (ld *LoopDetector) DetectLoop() bool {
	if len(ld.actions) < ld.windowSize*2 {
		return false
	}

	recent := ld.actions[len(ld.actions)-ld.windowSize:]
	prev := ld.actions[len(ld.actions)-ld.windowSize*2 : len(ld.actions)-ld.windowSize]

	matchCount := 0
	for i := range recent {
		if recent[i] == prev[i] {
			matchCount++
		}
	}

	similarity := float64(matchCount) / float64(ld.windowSize)
	return similarity >= ld.similarityThresh
}

func (ld *LoopDetector) Reset() {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	ld.actions = []string{}
}

func (ld *LoopDetector) GetHistory() []string {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	return append([]string{}, ld.actions...)
}

func NewRefactorCoordinator() *RefactorCoordinator {
	return &RefactorCoordinator{
		backups: make(map[string][]byte),
		phase:   "planning",
	}
}

func (rc *RefactorCoordinator) BeginRefactor(files []string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.phase = "backup"
	rc.backups = make(map[string][]byte)
	rc.changes = []FileChange{}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err == nil {
			rc.backups[f] = data
		}
	}

	rc.phase = "executing"
	return nil
}

func (rc *RefactorCoordinator) RecordChange(path, content string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	hash := fmt.Sprintf("%x", hashStrings(content))
	rc.changes = append(rc.changes, FileChange{
		Path:    path,
		Content: content,
		Hash:    hash,
	})
}

func (rc *RefactorCoordinator) Commit() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.phase = "committing"

	for _, change := range rc.changes {
		if err := os.WriteFile(change.Path, []byte(change.Content), 0644); err != nil {
			rc.rollbackLocked()
			return fmt.Errorf("failed to write %s: %w", change.Path, err)
		}
	}

	rc.backups = make(map[string][]byte)
	rc.changes = []FileChange{}
	rc.phase = "completed"
	return nil
}

func (rc *RefactorCoordinator) Rollback() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.rollbackLocked()
}

func (rc *RefactorCoordinator) rollbackLocked() {
	for path, content := range rc.backups {
		os.WriteFile(path, content, 0644)
	}
	rc.backups = make(map[string][]byte)
	rc.changes = []FileChange{}
	rc.phase = "rolled_back"
}

func (rc *RefactorCoordinator) Validate() ([]string, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	var errors []string

	for _, change := range rc.changes {
		data, err := os.ReadFile(change.Path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("cannot read %s: %v", change.Path, err))
			continue
		}

		hash := fmt.Sprintf("%x", hashStrings(string(data)))
		if hash != change.Hash {
			errors = append(errors, fmt.Sprintf("content mismatch for %s", change.Path))
		}
	}

	return errors, nil
}

func hashStrings(s string) string {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return fmt.Sprintf("%x", h)
}

func (ew *EnhancedWorkflow) ExecuteWithLearning(ctx context.Context, task string) error {
	if suggestions := ew.learning.GetSuggestions(task); len(suggestions) > 0 {
		fmt.Println("\n[LEARNING] Suggested approaches:")
		for _, s := range suggestions {
			fmt.Printf("  → %s\n", s)
		}
	}

	canProceed, msg := ew.budgetManager.CanProceed()
	if !canProceed {
		return fmt.Errorf("budget limit reached: %s", msg)
	}

	if len(ew.budgetManager.GetAlerts()) > 0 {
		fmt.Println("\n[BUDGET ALERTS]")
		for _, a := range ew.budgetManager.GetAlerts() {
			fmt.Printf("  [%s] %s\n", a.Level, a.Message)
		}
	}

	startTime := time.Now()
	success := false
	var actions []string

	defer func() {
		duration := time.Since(startTime)
		ew.learning.RecordAttempt(task, actions, success, fmt.Sprintf("duration=%v", duration))
		eventType := "task_failed"
		if success {
			eventType = "task_complete"
		}
		ew.telemetry.Track(eventType, map[string]interface{}{
			"task":     task,
			"duration": duration,
			"success":  success,
		})
	}()

	if err := ew.ExecuteTask(ctx, task); err != nil {
		success = false
		return err
	}

	success = true
	return nil
}

func (ew *EnhancedWorkflow) ExecuteWithHealing(ctx context.Context, task string) error {
	attempts := 0
	maxAttempts := 3
	var lastErr error

	for attempts < maxAttempts {
		attempts++

		if attempts > 1 {
			fmt.Printf("\n[RETRY] Attempt %d/%d\n", attempts, maxAttempts)
			ew.loopDetector.Reset()
		}

		if loopDetected := ew.loopDetector.Record(task); loopDetected {
			fmt.Println("[WARNING] Loop detected! Bailing out.")
			return fmt.Errorf("loop detected after %d attempts", attempts-1)
		}

		if err := ew.ExecuteWithLearning(ctx, task); err != nil {
			lastErr = err
			healResult := ew.selfHealer.AnalyzeAndHeal(err.Error())
			if healResult.Fixed {
				fmt.Printf("[HEALED] %s: %s\n", healResult.Action, healResult.Message)
				continue
			}
			return err
		}
		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

func (ew *EnhancedWorkflow) RefactorFiles(ctx context.Context, files []string, transform func(string) string) error {
	if err := ew.refactorCoord.BeginRefactor(files); err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			ew.refactorCoord.Rollback()
		}
	}()

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			ew.refactorCoord.Rollback()
			return fmt.Errorf("cannot read %s: %w", f, err)
		}

		transformed := transform(string(content))
		ew.refactorCoord.RecordChange(f, transformed)
	}

	if err := ew.refactorCoord.Commit(); err != nil {
		ew.refactorCoord.Rollback()
		return err
	}

	if errors, _ := ew.refactorCoord.Validate(); len(errors) > 0 {
		fmt.Println("[VALIDATION ERRORS]")
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	return nil
}

func (ew *EnhancedWorkflow) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"learning":     ew.learning.GetStats(),
		"budget":       fmt.Sprintf("$%.2f/$%.2f session, $%.2f/$%.2f daily", ew.budgetManager.spentSession, ew.budgetManager.sessionLimit, ew.budgetManager.spentToday, ew.budgetManager.dailyLimit),
		"telemetry":    ew.telemetry.Summary(),
		"loop_history": len(ew.loopDetector.GetHistory()),
	}
}
