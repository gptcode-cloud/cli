package autonomous

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chuchu/internal/agents"
	"chuchu/internal/llm"
)

// Symphony represents a multi-movement task execution
type Symphony struct {
	ID              string     `json:"id"`
	Task            string     `json:"task"`
	Movements       []Movement `json:"movements"`
	CurrentMovement int        `json:"current_movement"`
	Status          string     `json:"status"` // "pending", "executing", "completed", "failed"
	StartTime       time.Time  `json:"start_time"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

// Executor executes tasks using the Symphony pattern
type Executor struct {
	analyzer  *TaskAnalyzer
	planner   *agents.PlannerAgent
	editor    *agents.EditorAgent
	validator *agents.ValidatorAgent
	cwd       string
}

// NewExecutor creates a new symphony executor
func NewExecutor(
	analyzer *TaskAnalyzer,
	planner *agents.PlannerAgent,
	editor *agents.EditorAgent,
	validator *agents.ValidatorAgent,
	cwd string,
) *Executor {
	return &Executor{
		analyzer:  analyzer,
		planner:   planner,
		editor:    editor,
		validator: validator,
		cwd:       cwd,
	}
}

// Execute executes a task autonomously
func (e *Executor) Execute(ctx context.Context, task string) error {
	// 1. Analyze task
	fmt.Println("🔍 Analyzing task...")
	analysis, err := e.analyzer.Analyze(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to analyze task: %w", err)
	}

	fmt.Printf("   Intent: %s\n", analysis.Intent)
	fmt.Printf("   Complexity: %d/10\n", analysis.Complexity)

	// 2. If simple (complexity < 7), execute directly
	if analysis.Complexity < 7 {
		fmt.Println("\n✨ Executing directly (simple task)...")
		return e.executeDirect(ctx, task, analysis)
	}

	// 3. Create symphony
	fmt.Printf("\n🎼 Complex task detected! Creating symphony with %d movements...\n\n", len(analysis.Movements))

	symphony := &Symphony{
		ID:              generateID(),
		Task:            task,
		Movements:       analysis.Movements,
		CurrentMovement: 0,
		Status:          "executing",
		StartTime:       time.Now(),
	}

	// 4. Execute each movement
	for i, movement := range symphony.Movements {
		symphony.CurrentMovement = i

		fmt.Printf("🎵 Movement %d/%d: %s\n", i+1, len(symphony.Movements), movement.Name)
		fmt.Printf("   Goal: %s\n", movement.Goal)

		err := e.executeMovement(ctx, &symphony.Movements[i])
		if err != nil {
			symphony.Status = "failed"
			return fmt.Errorf("movement %d failed: %w", i+1, err)
		}

		fmt.Printf("   ✅ Movement %d complete\n\n", i+1)

		// Save checkpoint (enable resume)
		if err := e.saveCheckpoint(symphony); err != nil {
			fmt.Printf("   ⚠️  Warning: failed to save checkpoint: %v\n", err)
		}
	}

	now := time.Now()
	symphony.Status = "completed"
	symphony.CompletedAt = &now

	fmt.Println("✅ Symphony complete!")
	return nil
}

// executeDirect executes a simple task without decomposition
func (e *Executor) executeDirect(ctx context.Context, task string, analysis *TaskAnalysis) error {
	// Use existing orchestrated mode workflow
	// For now, just create a plan and execute

	// 1. Create plan
	plan, err := e.planner.CreatePlan(ctx, task, "", nil)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// 2. Execute with editor
	history := []llm.ChatMessage{
		{Role: "user", Content: plan},
	}

	result, modifiedFiles, err := e.editor.Execute(ctx, history, nil)
	if err != nil {
		return fmt.Errorf("failed to execute: %w", err)
	}

	fmt.Printf("\n✅ Task complete!\n")
	fmt.Printf("   Modified: %d files\n", len(modifiedFiles))
	if result != "" {
		fmt.Printf("   %s\n", result)
	}

	return nil
}

// executeMovement executes a single movement
func (e *Executor) executeMovement(ctx context.Context, movement *Movement) error {
	movement.Status = "executing"

	// 1. Create plan for this movement only
	fmt.Println("   📋 Creating plan...")
	plan, err := e.planner.CreatePlan(ctx, movement.Goal, "", nil)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// 2. Execute with editor
	fmt.Println("   ✏️  Executing changes...")
	history := []llm.ChatMessage{
		{Role: "user", Content: plan},
	}

	_, modifiedFiles, err := e.editor.Execute(ctx, history, nil)
	if err != nil {
		return fmt.Errorf("failed to execute movement: %w", err)
	}

	// 3. Validate success criteria
	if len(movement.SuccessCriteria) > 0 {
		fmt.Println("   🔎 Validating...")
		validation, err := e.validator.Validate(ctx, plan, modifiedFiles, nil)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		if !validation.Success {
			movement.Status = "failed"
			return fmt.Errorf("validation failed: %v", validation.Issues)
		}
	}

	movement.Status = "completed"
	return nil
}

// saveCheckpoint saves symphony state for resume capability
func (e *Executor) saveCheckpoint(symphony *Symphony) error {
	checkpointsDir := filepath.Join(os.Getenv("HOME"), ".chuchu", "symphonies")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return err
	}

	checkpointPath := filepath.Join(checkpointsDir, symphony.ID+".json")

	data, err := json.MarshalIndent(symphony, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(checkpointPath, data, 0644)
}

// generateID generates a random symphony ID
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
