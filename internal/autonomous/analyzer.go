package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"chuchu/internal/agents"
	"chuchu/internal/llm"
)

// TaskAnalysis represents the result of analyzing a task
type TaskAnalysis struct {
	Intent        string     `json:"intent"`
	Verb          string     `json:"verb"`
	Complexity    int        `json:"complexity"`
	RequiredFiles []string   `json:"required_files"`
	OutputFiles   []string   `json:"output_files"`
	Movements     []Movement `json:"movements,omitempty"`
}

// Movement represents a single phase in a complex task
type Movement struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Goal            string   `json:"goal"`
	Dependencies    []string `json:"dependencies"`
	RequiredFiles   []string `json:"required_files"`
	OutputFiles     []string `json:"output_files"`
	SuccessCriteria []string `json:"success_criteria"`
	Status          string   `json:"status"` // "pending", "executing", "completed", "failed"
}

// TaskAnalyzer analyzes tasks and decomposes them into movements if complex
type TaskAnalyzer struct {
	classifier *agents.Classifier
	llm        llm.Provider
	cwd        string
	model      string
}

// NewTaskAnalyzer creates a new task analyzer
func NewTaskAnalyzer(classifier *agents.Classifier, llmProvider llm.Provider, cwd string, model string) *TaskAnalyzer {
	return &TaskAnalyzer{
		classifier: classifier,
		llm:        llmProvider,
		cwd:        cwd,
		model:      model,
	}
}

// Analyze analyzes a task and determines if it needs decomposition
func (a *TaskAnalyzer) Analyze(ctx context.Context, task string) (*TaskAnalysis, error) {
	analysis := &TaskAnalysis{}

	// 1. Use existing classifier for intent
	intent, err := a.classifier.ClassifyIntent(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to classify intent: %w", err)
	}
	analysis.Intent = string(intent)

	// 2. Extract verb and files
	analysis.Verb = extractVerb(task)
	analysis.RequiredFiles = extractFileMentions(task)

	// 3. Estimate complexity (1-10)
	complexity, err := a.estimateComplexity(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate complexity: %w", err)
	}
	analysis.Complexity = complexity

	// 4. If complex (>= 7), decompose into movements
	if complexity >= 7 {
		movements, err := a.decomposeIntoMovements(ctx, task, analysis)
		if err != nil {
			return nil, fmt.Errorf("failed to decompose into movements: %w", err)
		}
		analysis.Movements = movements
	}

	return analysis, nil
}

// extractVerb extracts the primary verb from a task
func extractVerb(task string) string {
	task = strings.ToLower(task)

	verbs := []string{
		"create", "add", "remove", "delete", "update", "modify",
		"refactor", "reorganize", "unify", "split", "merge",
		"read", "list", "show", "explain", "analyze",
	}

	for _, verb := range verbs {
		if strings.Contains(task, verb) {
			return verb
		}
	}

	return "unknown"
}

// extractFileMentions extracts explicit file paths from the task
func extractFileMentions(task string) []string {
	// Match patterns like:
	// - docs/_posts/file.md
	// - src/main.go
	// - /absolute/path.txt
	// Order matters: longer extensions first to avoid .js matching before .json
	filePattern := regexp.MustCompile(`[a-zA-Z0-9_\-./]+\.(json|yaml|yml|md|go|js|ts|py|txt|html|css)`)
	matches := filePattern.FindAllString(task, -1)

	// Deduplicate
	seen := make(map[string]bool)
	var files []string
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			files = append(files, match)
		}
	}

	return files
}

// estimateComplexity uses LLM to score task complexity 1-10
func (a *TaskAnalyzer) estimateComplexity(ctx context.Context, task string) (int, error) {
	prompt := fmt.Sprintf(`Rate the complexity of this task on a scale of 1-10.

Task: %s

Complexity scale:
- 1-3: Simple (single file, clear action)
  Examples: "create hello.md with greeting", "read main.go"
  
- 4-6: Medium (2-5 files, straightforward)
  Examples: "add error handling to auth.go", "update docs/readme.md with new commands"
  
- 7-8: Complex (multiple files, requires planning)
  Examples: "reorganize docs files into categories", "refactor authentication system"
  
- 9-10: Very complex (many files, multiple phases)
  Examples: "migrate entire codebase from X to Y", "redesign application architecture"

Consider:
- Number of files involved
- Number of distinct steps required
- Ambiguity in requirements
- Potential for errors

Respond with ONLY a number 1-10, nothing else.`, task)

	resp, err := a.llm.Chat(ctx, llm.ChatRequest{
		UserPrompt: prompt,
		Model:      a.model,
	})
	if err != nil {
		return 0, err
	}

	// Parse response
	scoreStr := strings.TrimSpace(resp.Text)
	var score int
	_, err = fmt.Sscanf(scoreStr, "%d", &score)
	if err != nil {
		// Fallback: try to find first number
		re := regexp.MustCompile(`\d+`)
		match := re.FindString(scoreStr)
		if match != "" {
			fmt.Sscanf(match, "%d", &score)
		}
	}

	// Clamp to 1-10
	if score < 1 {
		score = 1
	}
	if score > 10 {
		score = 10
	}

	return score, nil
}

// decomposeIntoMovements breaks a complex task into movements
func (a *TaskAnalyzer) decomposeIntoMovements(ctx context.Context, task string, analysis *TaskAnalysis) ([]Movement, error) {
	prompt := fmt.Sprintf(`Decompose this complex task into 2-5 independent movements (phases).

Task: %s
Intent: %s
Verb: %s

Rules:
- Each movement should be independently executable
- Define clear dependencies (Movement B depends on Movement A)
- Each movement should have 1-3 success criteria
- Movements should be sequential (not parallel)
- Be specific about files to read/create

Example:
Task: "reorganize all docs files"
Response:
[
  {
    "id": "movement-1",
    "name": "Analyze Structure",
    "description": "Create inventory of all documentation files",
    "goal": "Understand current docs organization",
    "dependencies": [],
    "required_files": ["docs/**/*.md"],
    "output_files": ["~/.chuchu/inventory.json"],
    "success_criteria": [
      "inventory.json exists",
      "all docs files are cataloged",
      "files are categorized by type"
    ]
  },
  {
    "id": "movement-2",
    "name": "Split Features",
    "description": "Break features.md into individual feature pages",
    "goal": "Create separate page for each feature",
    "dependencies": ["movement-1"],
    "required_files": ["docs/features.md"],
    "output_files": ["docs/features/*.md"],
    "success_criteria": [
      "docs/features/ directory exists",
      "6+ feature files created",
      "each file has proper front matter"
    ]
  }
]

Return ONLY valid JSON array of movements, no explanation.`, task, analysis.Intent, analysis.Verb)

	resp, err := a.llm.Chat(ctx, llm.ChatRequest{
		UserPrompt: prompt,
		Model:      a.model,
	})
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var movements []Movement
	responseText := strings.TrimSpace(resp.Text)

	// Remove markdown code blocks if present
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	err = json.Unmarshal([]byte(responseText), &movements)
	if err != nil {
		return nil, fmt.Errorf("failed to parse movements JSON: %w\nResponse: %s", err, responseText)
	}

	// Initialize status
	for i := range movements {
		movements[i].Status = "pending"
	}

	return movements, nil
}
