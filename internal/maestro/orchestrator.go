package maestro

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gptcode/internal/agents"
	"gptcode/internal/events"
	"gptcode/internal/llm"
)

// Maestro orchestrates autonomous execution with verification and recovery
type Maestro struct {
	Provider       llm.Provider
	CWD            string
	Model          string
	Events         *events.Emitter
	Verifiers      []Verifier
	Recovery       *RecoveryStrategy
	Checkpoints    *CheckpointSystem
	MaxRetries     int
	ModifiedFiles  []string
	CurrentStepIdx int
}

// NewMaestro creates a new Maestro orchestrator
func NewMaestro(provider llm.Provider, cwd, model string) *Maestro {
	checkpoints := NewCheckpointSystem(cwd)
	return &Maestro{
		Provider:    provider,
		CWD:         cwd,
		Model:       model,
		Events:      events.NewEmitter(os.Stderr),
		Verifiers:   []Verifier{NewBuildVerifier(cwd), NewTestVerifier(cwd)},
		Recovery:    NewRecoveryStrategy(3, checkpoints),
		Checkpoints: checkpoints,
		MaxRetries:  3,
	}
}

// ExecutePlan runs the autonomous execution loop
func (m *Maestro) ExecutePlan(ctx context.Context, planContent string) error {
	m.CurrentStepIdx = 0
	m.ModifiedFiles = nil
	_ = m.Events.Status("\u001b[36mStarting autonomous execution...\u001b[0m")

	// Parse plan into steps (simple version: split by phases)
	steps := m.parsePlan(planContent)

	for stepIdx, step := range steps {
		_ = m.Events.Status(fmt.Sprintf("\u001b[34mStep %d/%d\u001b[0m: %s", stepIdx+1, len(steps), step.Title))

		var lastCheckpoint *Checkpoint
		var err error

		// Try execution with retries
		for attempt := 0; attempt < m.MaxRetries; attempt++ {
			if attempt > 0 {
				_ = m.Events.Status(fmt.Sprintf("Retry %d/%d", attempt, m.MaxRetries))
			}

			// Execute the step
			m.CurrentStepIdx = stepIdx
			before := m.gitChangedFiles()
			err = m.executeStep(ctx, step)
			after := m.gitChangedFiles()
			m.ModifiedFiles = diffStringSlices(before, after)
			if err != nil {
				_ = m.Events.Notify(fmt.Sprintf("\u001b[31mExecution failed\u001b[0m: %v", err), "error")
				continue
			}

			// Verify the changes
			verifyResult, verifyErr := m.verify(ctx)
			if verifyErr != nil {
				_ = m.Events.Notify(fmt.Sprintf("\u001b[31mVerification error\u001b[0m: %v", verifyErr), "error")
				err = verifyErr
				continue
			}

			if !verifyResult.Success {
				_ = m.Events.Notify(fmt.Sprintf("\u001b[33mVerification failed\u001b[0m: %s", verifyResult.Output), "warn")

				// Classify error and decide recovery strategy
				errorType := ClassifyError(verifyResult.Output)
				_ = m.Events.Status(fmt.Sprintf("Error type: %s, attempting recovery...", errorType))

				// For build errors, rollback if we have a checkpoint
				if errorType == ErrorBuild && lastCheckpoint != nil {
					_ = m.Events.Status("\u001b[35mRolling back to last checkpoint...\u001b[0m")
					if rollbackErr := m.Recovery.Rollback(lastCheckpoint.ID); rollbackErr != nil {
						_ = m.Events.Notify(fmt.Sprintf("Rollback failed: %v", rollbackErr), "error")
					}
				}

				err = fmt.Errorf("verification failed: %s", verifyResult.Error)
				continue
			}

			// Success! Save checkpoint
			_ = m.Events.Status("\u001b[32mVerification passed\u001b[0m, saving checkpoint...")
			_, err = m.Checkpoints.Save(stepIdx, m.ModifiedFiles)
			if err != nil {
				_ = m.Events.Notify(fmt.Sprintf("Checkpoint save failed: %v", err), "warn")
			}

			_ = m.Events.Complete()
			break
		}

		if err != nil {
			return fmt.Errorf("step %d failed after %d retries: %w", stepIdx, m.MaxRetries, err)
		}
	}

	_ = m.Events.Message("\u001b[32mAutonomous execution completed successfully!\u001b[0m")
	return nil
}

// ResumeExecution continues from the last successful checkpoint
func (m *Maestro) ResumeExecution(ctx context.Context, planContent string) error {
	steps := m.parsePlan(planContent)

	// Load latest checkpoint directory and infer step index
	// Simple approach: scan .gptcode/checkpoints and pick the latest, then parse step index from id
	ckptDir := m.Checkpoints.RootDir
	dirs, err := os.ReadDir(ckptDir)
	if err != nil || len(dirs) == 0 {
		return fmt.Errorf("no checkpoints to resume")
	}

	latest := dirs[len(dirs)-1].Name()
	var step int
	_, err = fmt.Sscanf(latest, "ckpt_%d_", &step)
	if err != nil {
		return fmt.Errorf("failed to parse checkpoint step: %w", err)
	}

	m.CurrentStepIdx = step + 1
	if m.CurrentStepIdx >= len(steps) {
		return fmt.Errorf("nothing to resume: all steps completed")
	}

	return m.ExecutePlan(ctx, planContent)
}

// executeStep runs a single step of the plan
func (m *Maestro) ExecuteStep(ctx context.Context, step PlanStep) error {
	return m.executeStep(ctx, step)
}

func (m *Maestro) executeStep(ctx context.Context, step PlanStep) error {
	editorAgent := agents.NewEditor(m.Provider, m.CWD, m.Model)

	statusCallback := func(status string) {
		_ = m.Events.Status(status)
	}

	history := []llm.ChatMessage{
		{Role: "user", Content: fmt.Sprintf("Implement this step:\n\n# %s\n\n%s", step.Title, step.Content)},
	}

	_, modifiedFiles, err := editorAgent.Execute(ctx, history, statusCallback)
	// Track modified files in the maestro state if needed, but for now we rely on git diff in the outer loop
	// or we could return them here.
	// The outer loop uses gitChangedFiles(), but we should probably use these modifiedFiles too.
	// However, executeStep returns error only.
	// Let's just update the signature call for now.
	_ = modifiedFiles
	return err
}

// verify runs all verifiers
func (m *Maestro) verify(ctx context.Context) (*VerificationResult, error) {
	for _, verifier := range m.Verifiers {
		result, err := verifier.Verify(ctx)
		if err != nil {
			return nil, err
		}
		if !result.Success {
			return result, nil
		}
	}

	return &VerificationResult{Success: true}, nil
}

func (m *Maestro) gitChangedFiles() []string {
	cmd := exec.Command("git", "--no-pager", "diff", "--name-only")
	cmd.Dir = m.CWD
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			files = append(files, l)
		}
	}
	return files
}

func diffStringSlices(a, b []string) []string {
	m := make(map[string]bool)
	for _, x := range a {
		m[x] = true
	}
	var d []string
	for _, y := range b {
		if !m[y] {
			d = append(d, y)
		}
	}
	return d
}

type PlanStep struct {
	Title    string
	Content  string
	SubSteps []PlanStep
}

func (m *Maestro) ParsePlan(plan string) []PlanStep {
	return m.parsePlan(plan)
}

func (m *Maestro) parsePlan(plan string) []PlanStep {
	var steps []PlanStep
	lines := strings.Split(plan, "\n")

	var currentStep *PlanStep
	var currentSubStep *PlanStep

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			if currentStep != nil {
				if currentSubStep != nil {
					currentStep.SubSteps = append(currentStep.SubSteps, *currentSubStep)
					currentSubStep = nil
				}
				steps = append(steps, *currentStep)
			}
			currentStep = &PlanStep{
				Title:    strings.TrimPrefix(line, "## "),
				Content:  "",
				SubSteps: []PlanStep{},
			}
			currentSubStep = nil
		} else if strings.HasPrefix(line, "### ") {
			if currentSubStep != nil && currentStep != nil {
				currentStep.SubSteps = append(currentStep.SubSteps, *currentSubStep)
			}
			currentSubStep = &PlanStep{
				Title:    strings.TrimPrefix(line, "### "),
				Content:  "",
				SubSteps: []PlanStep{},
			}
		} else if currentSubStep != nil {
			currentSubStep.Content += line + "\n"
		} else if currentStep != nil {
			currentStep.Content += line + "\n"
		}
	}

	if currentStep != nil {
		if currentSubStep != nil {
			currentStep.SubSteps = append(currentStep.SubSteps, *currentSubStep)
		}
		steps = append(steps, *currentStep)
	}

	return flattenSteps(steps)
}

func flattenSteps(steps []PlanStep) []PlanStep {
	var flat []PlanStep
	for _, step := range steps {
		if len(step.SubSteps) > 0 {
			for _, sub := range step.SubSteps {
				flat = append(flat, PlanStep{
					Title:   step.Title + " / " + sub.Title,
					Content: step.Content + "\n" + sub.Content,
				})
			}
		} else {
			flat = append(flat, step)
		}
	}
	return flat
}
