package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
	Tool          string   `json:"tool"`
	Result        string   `json:"result"`
	Error         string   `json:"error,omitempty"`
	ModifiedFiles []string `json:"modified_files,omitempty"`
}

func GetAvailableTools() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "read_file",
				"description": "Read the contents of a file in the current repository",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Relative path to the file from repository root",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "list_files",
				"description": "List files in a directory of the current repository",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Relative path to directory from repository root (empty for root)",
						},
						"pattern": map[string]interface{}{
							"type":        "string",
							"description": "Optional glob pattern to filter files (e.g., '*.go', 'test_*.ex')",
						},
					},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "run_command",
				"description": "Execute a shell command in the repository directory (use for tests, linting, etc.)",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "Shell command to execute",
						},
					},
					"required": []string{"command"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "search_code",
				"description": "Search for a pattern in code files using grep",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"pattern": map[string]interface{}{
							"type":        "string",
							"description": "Search pattern (regex)",
						},
						"file_pattern": map[string]interface{}{
							"type":        "string",
							"description": "Optional file pattern to limit search (e.g., '*.go')",
						},
					},
					"required": []string{"pattern"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "read_guideline",
				"description": "Read detailed coding guidelines from ~/.gptcode/guidelines/ directory. Use when you need language-specific guidance, naming conventions, or TDD workflow details.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"guideline": map[string]interface{}{
							"type":        "string",
							"description": "Guideline name: 'tdd', 'naming', or 'languages'",
							"enum":        []string{"tdd", "naming", "languages"},
						},
					},
					"required": []string{"guideline"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "write_file",
				"description": "Write content to a file (creates or overwrites). Use this to save edited files.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Relative path to the file from repository root",
						},
						"content": map[string]interface{}{
							"type":        "string",
							"description": "Complete file content to write",
						},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "project_map",
				"description": "Get a tree-like view of the project structure",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"max_depth": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum depth to traverse (default 3)",
						},
					},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "apply_patch",
				"description": "Replace a block of text in a file",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "File path",
						},
						"search": map[string]interface{}{
							"type":        "string",
							"description": "Exact text block to replace",
						},
						"replace": map[string]interface{}{
							"type":        "string",
							"description": "New text block",
						},
					},
					"required": []string{"path", "search", "replace"},
				},
			},
		},
	}
}

func ExecuteTool(call ToolCall, workdir string) ToolResult {
	switch call.Name {
	case "read_file":
		return readFile(call, workdir)
	case "list_files":
		return listFiles(call, workdir)
	case "run_command":
		return runCommand(call, workdir)
	case "search_code":
		return searchCode(call, workdir)
	case "read_guideline":
		return readGuideline(call)
	case "write_file":
		return writeFile(call, workdir)
	case "project_map":
		return ProjectMap(call, workdir)
	case "apply_patch":
		return ApplyPatch(call, workdir)
	default:
		return ToolResult{
			Tool:  call.Name,
			Error: "Unknown tool",
		}
	}
}

type LLMToolCall struct {
	ID        string
	Name      string
	Arguments string
}

func ExecuteToolFromLLM(call LLMToolCall, workdir string) ToolResult {
	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(call.Arguments), &argsMap); err != nil {
		return ToolResult{
			Tool:  call.Name,
			Error: fmt.Sprintf("Failed to parse arguments: %v", err),
		}
	}

	toolCall := ToolCall{
		Name:      call.Name,
		Arguments: argsMap,
	}

	return ExecuteTool(toolCall, workdir)
}

func readFile(call ToolCall, workdir string) ToolResult {
	path, ok := call.Arguments["path"].(string)
	if !ok {
		return ToolResult{Tool: "read_file", Error: "path parameter required"}
	}

	fullPath := filepath.Join(workdir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return ToolResult{Tool: "read_file", Error: err.Error()}
	}

	result := string(content)
	lines := strings.Split(result, "\n")
	if len(lines) > 200 {
		truncated := strings.Join(lines[:200], "\n")
		result = truncated + fmt.Sprintf("\n... (truncated, %d total lines)", len(lines))
	}

	return ToolResult{
		Tool:   "read_file",
		Result: result,
	}
}

func listFiles(call ToolCall, workdir string) ToolResult {
	pathArg, _ := call.Arguments["path"].(string)
	pattern, _ := call.Arguments["pattern"].(string)

	targetPath := filepath.Join(workdir, pathArg)

	var files []string
	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(workdir, path)

		if pattern != "" {
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if !matched {
				return nil
			}
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return ToolResult{Tool: "list_files", Error: err.Error()}
	}

	result := strings.Join(files, "\n")
	if len(files) > 30 {
		truncated := strings.Join(files[:30], "\n")
		result = truncated + fmt.Sprintf("\n... (%d more files, %d total)", len(files)-30, len(files))
	}

	return ToolResult{
		Tool:   "list_files",
		Result: result,
	}
}

func runCommand(call ToolCall, workdir string) ToolResult {
	command, ok := call.Arguments["command"].(string)
	if !ok {
		return ToolResult{Tool: "run_command", Error: "command parameter required"}
	}

	// Block sudo commands to prevent password prompts
	if strings.Contains(command, "sudo ") || strings.HasPrefix(command, "sudo") {
		return ToolResult{
			Tool:  "run_command",
			Error: "sudo commands not allowed in autonomous mode. Present sudo commands to user for manual execution.",
		}
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workdir
	output, err := cmd.CombinedOutput()

	result := ToolResult{
		Tool:   "run_command",
		Result: string(output),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

func searchCode(call ToolCall, workdir string) ToolResult {
	pattern, ok := call.Arguments["pattern"].(string)
	if !ok {
		return ToolResult{Tool: "search_code", Error: "pattern parameter required"}
	}

	filePattern, _ := call.Arguments["file_pattern"].(string)

	args := []string{"-r", "-n", pattern}
	if filePattern != "" {
		args = append(args, "--include="+filePattern)
	}
	args = append(args, workdir)

	cmd := exec.Command("grep", args...)
	output, err := cmd.CombinedOutput()

	result := ToolResult{
		Tool:   "search_code",
		Result: string(output),
	}

	if err != nil && len(output) == 0 {
		result.Error = "No matches found"
	}

	return result
}

func readGuideline(call ToolCall) ToolResult {
	guideline, ok := call.Arguments["guideline"].(string)
	if !ok {
		return ToolResult{Tool: "read_guideline", Error: "guideline parameter required"}
	}

	validGuidelines := map[string]bool{
		"tdd":       true,
		"naming":    true,
		"languages": true,
	}

	if !validGuidelines[guideline] {
		return ToolResult{
			Tool:  "read_guideline",
			Error: fmt.Sprintf("invalid guideline '%s'. Must be one of: tdd, naming, languages", guideline),
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ToolResult{Tool: "read_guideline", Error: fmt.Sprintf("could not get home dir: %v", err)}
	}

	guidelinePath := filepath.Join(home, ".gptcode", "guidelines", guideline+".md")
	content, err := os.ReadFile(guidelinePath)
	if err != nil {
		return ToolResult{Tool: "read_guideline", Error: fmt.Sprintf("could not read guideline: %v", err)}
	}

	return ToolResult{
		Tool:   "read_guideline",
		Result: string(content),
	}
}

func writeFile(call ToolCall, workdir string) ToolResult {
	path, ok := call.Arguments["path"].(string)
	if !ok {
		return ToolResult{Tool: "write_file", Error: "path parameter required"}
	}

	content, ok := call.Arguments["content"].(string)
	if !ok {
		return ToolResult{Tool: "write_file", Error: "content parameter required"}
	}

	fullPath := filepath.Join(workdir, path)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ToolResult{Tool: "write_file", Error: fmt.Sprintf("could not create directory: %v", err)}
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return ToolResult{Tool: "write_file", Error: err.Error()}
	}

	return ToolResult{
		Tool:          "write_file",
		Result:        fmt.Sprintf("File written successfully: %s (%d bytes)", path, len(content)),
		ModifiedFiles: []string{path},
	}
}
