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

	"chuchu/internal/maestro"
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
	analyzer *TaskAnalyzer
	maestro  *maestro.Conductor
	cwd      string
}

// NewExecutor creates a new symphony executor
func NewExecutor(
	analyzer *TaskAnalyzer,
	maestro *maestro.Conductor,
	cwd string,
) *Executor {
	return &Executor{
		analyzer: analyzer,
		maestro:  maestro,
		cwd:      cwd,
	}
}

// Execute executes a task autonomously
func (e *Executor) Execute(ctx context.Context, task string) error {
	// 1. Analyze task
	fmt.Println("Analyzing task...")
	analysis, err := e.analyzer.Analyze(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to analyze task: %w", err)
	}

	fmt.Printf("   Intent: %s\n", analysis.Intent)
	fmt.Printf("   Complexity: %d/10\n", analysis.Complexity)

	// 2. If simple (complexity <= 5 from ML analysis), execute directly
	if analysis.Complexity <= 5 {
		fmt.Println("\nExecuting directly (simple task)...")
		return e.executeDirect(ctx, task, analysis)
	}

	// 3. Complex task (ML scored >= 7): decompose into Symphony movements
	if len(analysis.Movements) == 0 {
		fmt.Println("\n[WARNING] Model failed to decompose task (returned empty plan).")
		fmt.Println("Falling back to direct execution...")
		return e.executeDirect(ctx, task, analysis)
	}

	fmt.Printf("\nComplex task detected! Creating symphony with %d movements...\n\n", len(analysis.Movements))

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

		fmt.Printf("Movement %d/%d: %s\n", i+1, len(symphony.Movements), movement.Name)
		fmt.Printf("   Goal: %s\n", movement.Goal)

		err := e.executeMovement(ctx, &symphony.Movements[i])
		if err != nil {
			symphony.Status = "failed"
			return fmt.Errorf("movement %d failed: %w", i+1, err)
		}

		fmt.Printf("   [OK] Movement %d complete\n\n", i+1)

		// Save checkpoint (enable resume)
		if err := e.saveCheckpoint(symphony); err != nil {
			fmt.Printf("   [WARNING] Failed to save checkpoint: %v\n", err)
		}
	}

	now := time.Now()
	symphony.Status = "completed"
	symphony.CompletedAt = &now

	fmt.Println("[OK] Symphony complete!")
	return nil
}

// executeDirect executes a simple task without decomposition
func (e *Executor) executeDirect(ctx context.Context, task string, analysis *TaskAnalysis) error {
	// Delegate to Maestro with complexity
	complexityStr := "simple"
	if analysis.Complexity >= 7 {
		complexityStr = "complex"
	} else if analysis.Complexity >= 5 {
		complexityStr = "medium"
	}
	return e.maestro.ExecuteTask(ctx, task, complexityStr)
}

// executeMovement executes a single movement with retry on review failure
func (e *Executor) executeMovement(ctx context.Context, movement *Movement) error {
	movement.Status = "executing"

	// Delegate to Maestro (movements are complex by definition)
	err := e.maestro.ExecuteTask(ctx, movement.Goal, "complex")
	if err != nil {
		movement.Status = "failed"
		return err
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
