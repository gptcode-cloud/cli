package maestro

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gptcode/internal/langdetect"
)

type VerificationResult struct {
	Success bool
	Output  string
	Error   error
}

type Verifier interface {
	Verify(ctx context.Context) (*VerificationResult, error)
}

type TestVerifier struct {
	Dir      string
	Language string
}

func NewTestVerifier(dir string) *TestVerifier {
	lang := detectLanguage(dir)
	return &TestVerifier{Dir: dir, Language: lang}
}

func (v *TestVerifier) Verify(ctx context.Context) (*VerificationResult, error) {
	var cmd *exec.Cmd

	switch v.Language {
	case "go":
		cmd = exec.CommandContext(ctx, "go", "test", "./...")
	case "javascript", "typescript":
		if fileExists(filepath.Join(v.Dir, "package.json")) {
			cmd = exec.CommandContext(ctx, "npm", "test")
		} else {
			return &VerificationResult{Success: true}, nil
		}
	case "python":
		if fileExists(filepath.Join(v.Dir, "pytest.ini")) || fileExists(filepath.Join(v.Dir, "setup.py")) {
			cmd = exec.CommandContext(ctx, "pytest")
		} else {
			return &VerificationResult{Success: true}, nil
		}
	case "elixir":
		cmd = exec.CommandContext(ctx, "mix", "test")
	case "ruby":
		if fileExists(filepath.Join(v.Dir, "Gemfile")) {
			cmd = exec.CommandContext(ctx, "bundle", "exec", "rspec")
		} else {
			cmd = exec.CommandContext(ctx, "rspec")
		}
	default:
		return &VerificationResult{Success: true}, nil
	}

	cmd.Dir = v.Dir
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		return &VerificationResult{
			Success: false,
			Output:  outStr,
			Error:   fmt.Errorf("tests failed: %w", err),
		}, nil
	}

	return &VerificationResult{
		Success: true,
		Output:  outStr,
		Error:   nil,
	}, nil
}

type BuildVerifier struct {
	Dir      string
	Language string
}

func NewBuildVerifier(dir string) *BuildVerifier {
	lang := detectLanguage(dir)
	return &BuildVerifier{Dir: dir, Language: lang}
}

func (v *BuildVerifier) Verify(ctx context.Context) (*VerificationResult, error) {
	// Get modified files from git
	gitCmd := exec.CommandContext(ctx, "git", "--no-pager", "diff", "--name-only")
	gitCmd.Dir = v.Dir
	gitOut, err := gitCmd.CombinedOutput()
	if err != nil {
		return &VerificationResult{Success: true, Output: "Could not get modified files, skipping build"}, nil
	}

	modifiedFiles := strings.Split(strings.TrimSpace(string(gitOut)), "\n")

	// If no files modified at all, skip build (likely a read-only task)
	if len(modifiedFiles) == 0 || (len(modifiedFiles) == 1 && modifiedFiles[0] == "") {
		return &VerificationResult{Success: true, Output: "No files modified, skipping build"}, nil
	}

	// Check if any modified file is a code file
	hasCodeFiles := false
	codeExtensions := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true,
		".jsx": true, ".tsx": true, ".java": true, ".c": true,
		".cpp": true, ".rs": true, ".rb": true, ".ex": true,
		".exs": true,
	}
	// Explicitly ignore documentation and data files
	nonCodeExtensions := map[string]bool{
		".md": true, ".txt": true, ".json": true, ".yaml": true,
		".yml": true, ".xml": true, ".html": true, ".css": true,
	}

	for _, file := range modifiedFiles {
		if file == "" {
			continue
		}
		// Skip if it's a non-code file
		isNonCode := false
		for ext := range nonCodeExtensions {
			if strings.HasSuffix(file, ext) {
				isNonCode = true
				break
			}
		}
		if isNonCode {
			continue
		}
		// Check if it's a code file
		for ext := range codeExtensions {
			if strings.HasSuffix(file, ext) {
				hasCodeFiles = true
				break
			}
		}
		if hasCodeFiles {
			break
		}
	}

	if !hasCodeFiles {
		return &VerificationResult{Success: true, Output: "No code files modified, skipping build"}, nil
	}

	var cmd *exec.Cmd

	switch v.Language {
	case "go":
		cmd = exec.CommandContext(ctx, "go", "build", "./...")
	case "javascript", "typescript":
		if fileExists(filepath.Join(v.Dir, "package.json")) {
			cmd = exec.CommandContext(ctx, "npm", "run", "build")
		} else {
			return &VerificationResult{Success: true}, nil
		}
	case "python":
		cmd = exec.CommandContext(ctx, "python", "-m", "py_compile")
	case "elixir":
		cmd = exec.CommandContext(ctx, "mix", "compile")
	case "ruby":
		return &VerificationResult{Success: true}, nil
	default:
		return &VerificationResult{Success: true}, nil
	}

	cmd.Dir = v.Dir
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		return &VerificationResult{
			Success: false,
			Output:  outStr,
			Error:   fmt.Errorf("build failed: %w", err),
		}, nil
	}

	return &VerificationResult{
		Success: true,
		Output:  outStr,
		Error:   nil,
	}, nil
}

func detectLanguage(dir string) string {
	lang := langdetect.DetectLanguage(dir)
	switch lang {
	case langdetect.Go:
		return "go"
	case langdetect.TypeScript:
		return "typescript"
	case langdetect.Python:
		return "python"
	case langdetect.Elixir:
		return "elixir"
	case langdetect.Ruby:
		return "ruby"
	default:
		return "unknown"
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
