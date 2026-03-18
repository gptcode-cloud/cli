package autonomous

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type SelfHealer struct {
	cwd string
}

func NewSelfHealer(cwd string) *SelfHealer {
	return &SelfHealer{cwd: cwd}
}

type HealResult struct {
	Success   bool
	Fixed     bool
	Action    string
	Message   string
	NewOutput string
	Output    string
}

func (sh *SelfHealer) AnalyzeAndHeal(output string) *HealResult {
	result := &HealResult{Output: output}

	// Check for common issues and try to fix

	// 1. Missing dependencies
	if strings.Contains(output, "cannot find package") ||
		strings.Contains(output, "module not found") {
		return sh.healMissingDeps(output)
	}

	// 2. Missing imports
	if strings.Contains(output, "undefined") ||
		strings.Contains(output, "import") && strings.Contains(output, "not used") {
		return sh.healMissingImports(output)
	}

	// 3. Syntax errors
	if strings.Contains(output, "syntax error") ||
		strings.Contains(output, "unexpected") {
		return sh.healSyntaxError(output)
	}

	// 4. Package errors (Go)
	if strings.Contains(output, "found packages") {
		return sh.healPackageError(output)
	}

	// 5. Format errors
	if strings.Contains(output, "format") && strings.Contains(output, "error") {
		return sh.healFormatError()
	}

	// 6. Type errors
	if strings.Contains(output, "cannot infer") ||
		strings.Contains(output, "type mismatch") {
		return &HealResult{
			Success: false,
			Fixed:   false,
			Action:  "needs_manual_fix",
			Message: "Type error requires manual attention",
		}
	}

	return result
}

func (sh *SelfHealer) healMissingDeps(output string) *HealResult {
	// Detect package manager
	if sh.fileExists("package.json") {
		cmd := exec.Command("npm", "install")
		cmd.Dir = sh.cwd
		out, err := cmd.CombinedOutput()
		if err != nil {
			return &HealResult{
				Success: false,
				Action:  "npm_install_failed",
				Message: string(out),
			}
		}
		return &HealResult{
			Success:   true,
			Fixed:     true,
			Action:    "npm_install",
			Message:   "Installed dependencies",
			NewOutput: string(out),
		}
	}

	if sh.fileExists("go.mod") {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = sh.cwd
		out, err := cmd.CombinedOutput()
		if err != nil {
			return &HealResult{
				Success: false,
				Action:  "go_mod_tidy_failed",
				Message: string(out),
			}
		}
		return &HealResult{
			Success:   true,
			Fixed:     true,
			Action:    "go_mod_tidy",
			Message:   "Tidied Go modules",
			NewOutput: string(out),
		}
	}

	if sh.fileExists("requirements.txt") {
		cmd := exec.Command("pip", "install", "-r", "requirements.txt")
		cmd.Dir = sh.cwd
		out, err := cmd.CombinedOutput()
		if err != nil {
			return &HealResult{
				Success: false,
				Action:  "pip_install_failed",
				Message: string(out),
			}
		}
		return &HealResult{
			Success:   true,
			Fixed:     true,
			Action:    "pip_install",
			Message:   "Installed Python dependencies",
			NewOutput: string(out),
		}
	}

	return &HealResult{
		Success: false,
		Action:  "deps_unknown",
		Message: "Could not detect package manager",
	}
}

func (sh *SelfHealer) healMissingImports(output string) *HealResult {
	result := &HealResult{Action: "analyze_imports"}

	// Try gofmt for Go files
	if sh.fileExists("go.mod") {
		cmd := exec.Command("gofmt", "-w", ".")
		cmd.Dir = sh.cwd
		out, err := cmd.CombinedOutput()
		if err != nil {
			return &HealResult{
				Success: false,
				Action:  "gofmt_failed",
				Message: string(out),
			}
		}
		return &HealResult{
			Success:   true,
			Fixed:     true,
			Action:    "gofmt",
			Message:   "Formatted Go files",
			NewOutput: string(out),
		}
	}

	return result
}

func (sh *SelfHealer) healSyntaxError(output string) *HealResult {
	// Extract line number if possible
	re := regexp.MustCompile(`line\s+(\d+)`)
	matches := re.FindStringSubmatch(output)

	if len(matches) > 1 {
		return &HealResult{
			Success: false,
			Fixed:   false,
			Action:  "syntax_error",
			Message: "Syntax error at line " + matches[1] + " - manual fix needed",
		}
	}

	return &HealResult{
		Success: false,
		Fixed:   false,
		Action:  "syntax_error",
		Message: "Syntax error - manual fix needed",
	}
}

func (sh *SelfHealer) healPackageError(output string) *HealResult {
	result := &HealResult{Action: "package_error"}

	// Find files with wrong package names
	files, err := os.ReadDir(sh.cwd)
	if err != nil {
		return &HealResult{
			Success: false,
			Action:  "package_scan_failed",
			Message: err.Error(),
		}
	}

	var wrongPackage []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".go") {
			content, err := os.ReadFile(sh.cwd + "/" + f.Name())
			if err == nil {
				if strings.Contains(string(content), "package main") && strings.Contains(string(content), "package ") {
					wrongPackage = append(wrongPackage, f.Name())
				}
			}
		}
	}

	if len(wrongPackage) > 0 {
		return &HealResult{
			Success: false,
			Fixed:   false,
			Action:  "package_mismatch",
			Message: "Found files with wrong package: " + strings.Join(wrongPackage, ", "),
		}
	}

	return result
}

func (sh *SelfHealer) healFormatError() *HealResult {
	result := &HealResult{Action: "format"}

	// Try prettier for JS/TS
	if sh.fileExists("package.json") {
		cmd := exec.Command("npx", "prettier", "--write", ".")
		cmd.Dir = sh.cwd
		out, err := cmd.CombinedOutput()
		if err == nil {
			return &HealResult{
				Success:   true,
				Fixed:     true,
				Action:    "prettier",
				Message:   "Formatted with Prettier",
				NewOutput: string(out),
			}
		}
	}

	// Try gofmt
	if sh.fileExists("go.mod") {
		cmd := exec.Command("gofmt", "-w", ".")
		cmd.Dir = sh.cwd
		out, err := cmd.CombinedOutput()
		if err == nil {
			return &HealResult{
				Success:   true,
				Fixed:     true,
				Action:    "gofmt",
				Message:   "Formatted with gofmt",
				NewOutput: string(out),
			}
		}
	}

	return result
}

func (sh *SelfHealer) fileExists(name string) bool {
	_, err := os.Stat(sh.cwd + "/" + name)
	return err == nil
}

func (sh *SelfHealer) GetAutoFixCommands(errMsg string) []string {
	cmds := []string{}

	lower := strings.ToLower(errMsg)

	// Go-specific
	if sh.fileExists("go.mod") {
		if strings.Contains(lower, "cannot find package") {
			cmds = append(cmds, "go mod tidy")
		}
		if strings.Contains(lower, "format") {
			cmds = append(cmds, "gofmt -w .")
		}
	}

	// Node-specific
	if sh.fileExists("package.json") {
		if strings.Contains(lower, "cannot find module") {
			cmds = append(cmds, "npm install")
		}
		if strings.Contains(lower, "eslint") {
			cmds = append(cmds, "npx eslint --fix .")
		}
		if strings.Contains(lower, "prettier") {
			cmds = append(cmds, "npx prettier --write .")
		}
	}

	// Python-specific
	if sh.fileExists("requirements.txt") {
		if strings.Contains(lower, "module not found") {
			cmds = append(cmds, "pip install -r requirements.txt")
		}
	}

	return cmds
}
