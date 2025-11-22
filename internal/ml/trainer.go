package ml

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AvailableModels lists all trainable ML models
var AvailableModels = map[string]ModelInfo{
	"complexity": {
		Name:        "complexity",
		Description: "Task complexity classifier (simple/complex/multistep)",
		Path:        "ml/complexity_detection",
		SetupScript: "setup_and_train.sh",
	},
	"complexity_detection": {
		Name:        "complexity_detection",
		Description: "Task complexity classifier (simple/complex/multistep) [deprecated: use 'complexity']",
		Path:        "ml/complexity_detection",
		SetupScript: "setup_and_train.sh",
	},
	"intent": {
		Name:        "intent",
		Description: "Intent classifier (query/editor/research/review)",
		Path:        "ml/intent",
		SetupScript: "setup_and_train.sh",
	},
}

// ModelInfo contains metadata about a trainable model
type ModelInfo struct {
	Name        string
	Description string
	Path        string
	SetupScript string
}

// Trainer handles ML model training
type Trainer struct {
	workDir string
}

// NewTrainer creates a new ML trainer
func NewTrainer(workDir string) *Trainer {
	return &Trainer{workDir: workDir}
}

// ListModels lists all available models
func (t *Trainer) ListModels() {
	fmt.Println("\n📚 Available ML models:")
	fmt.Println(strings.Repeat("=", 60))

	for _, model := range AvailableModels {
		fmt.Printf("\n  %s\n", model.Name)
		fmt.Printf("    %s\n", model.Description)
		fmt.Printf("    Location: %s\n", model.Path)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("\nUsage:")
	fmt.Println("  chu ml list")
	fmt.Println("  chu ml train <model_name>")
	fmt.Println("  chu ml test <model_name> [query]")
	fmt.Println("\nExample:")
	fmt.Println("  chu ml train complexity_detection")
	fmt.Println("  chu ml test complexity_detection \"fix typo\"")
	fmt.Println()
}

// Train trains a specific model
func (t *Trainer) Train(modelName string) error {
	// Validate model exists
	model, exists := AvailableModels[modelName]
	if !exists {
		return fmt.Errorf("model '%s' not found. Run 'chu train' to list available models", modelName)
	}

	// Build full path to model directory
	modelPath := filepath.Join(t.workDir, model.Path)

	// Check if model directory exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model directory not found: %s", modelPath)
	}

	// Build path to setup script
	scriptPath := filepath.Join(modelPath, model.SetupScript)

	// Check if setup script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("setup script not found: %s\nExpected: %s", model.SetupScript, scriptPath)
	}

	// Display info
	fmt.Printf("\n🤖 Training model: %s\n", model.Name)
	fmt.Printf("📝 %s\n", model.Description)
	fmt.Printf("📂 Location: %s\n", modelPath)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Run setup script
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = modelPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("training failed: %w", err)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("✅ Model '%s' trained successfully!\n", model.Name)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	return nil
}

// Test tests a trained model
func (t *Trainer) Test(modelName, query string) error {
	model, exists := AvailableModels[modelName]
	if !exists {
		return fmt.Errorf("model '%s' not found. Run 'chu ml list' to see available models", modelName)
	}

	modelPath := filepath.Join(t.workDir, model.Path)

	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model directory not found: %s", modelPath)
	}

	scriptPath := filepath.Join(modelPath, "scripts", "predict.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("prediction script not found: %s", scriptPath)
	}

	venvPath := filepath.Join(modelPath, "venv", "bin", "python")
	if _, err := os.Stat(venvPath); os.IsNotExist(err) {
		return fmt.Errorf("model not trained yet. Run 'chu ml train %s' first", modelName)
	}

	fmt.Printf("\n🧪 Testing model: %s\n", model.Name)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	var cmd *exec.Cmd
	if query != "" {
		cmd = exec.Command(venvPath, scriptPath, query)
	} else {
		cmd = exec.Command(venvPath, scriptPath)
	}

	cmd.Dir = modelPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	return nil
}

func (t *Trainer) Eval(modelName, evalFile string) error {
	model, exists := AvailableModels[modelName]
	if !exists {
		return fmt.Errorf("model '%s' not found. Run 'chu ml list' to see available models", modelName)
	}
	modelPath := filepath.Join(t.workDir, model.Path)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model directory not found: %s", modelPath)
	}
	scriptPath := filepath.Join(modelPath, "scripts", "eval.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("evaluation script not found: %s", scriptPath)
	}
	venvPath := filepath.Join(modelPath, "venv", "bin", "python")
	if _, err := os.Stat(venvPath); os.IsNotExist(err) {
		return fmt.Errorf("model not trained yet. Run 'chu ml train %s' first", modelName)
	}
	fmt.Printf("\n🧪 Evaluating model: %s\n", model.Name)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	var cmd *exec.Cmd
	if evalFile != "" {
		cmd = exec.Command(venvPath, scriptPath, evalFile)
	} else {
		cmd = exec.Command(venvPath, scriptPath)
	}
	cmd.Dir = modelPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("eval failed: %w", err)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	return nil
}
