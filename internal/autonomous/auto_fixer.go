package autonomous

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type AutoFixer struct {
	cwd string
}

func NewAutoFixer(cwd string) *AutoFixer {
	return &AutoFixer{cwd: cwd}
}

type FixResult struct {
	Success bool
	Action  string
	Output  string
	Error   string
}

func (af *AutoFixer) RunAll() *FixResult {
	// Run all auto-fix operations in order
	results := []FixResult{}

	// 1. Format code
	if r := af.FormatCode(); r.Success {
		results = append(results, *r)
	}

	// 2. Fix dependencies
	if r := af.FixDependencies(); r.Success {
		results = append(results, *r)
	}

	// 3. Lint fix
	if r := af.LintFix(); r.Success {
		results = append(results, *r)
	}

	// 4. Type check
	if r := af.TypeCheck(); r.Success {
		results = append(results, *r)
	}

	// Return combined result
	success := true
	var output strings.Builder
	for _, r := range results {
		if !r.Success {
			success = false
		}
		if r.Output != "" {
			output.WriteString(r.Output)
			output.WriteString("\n")
		}
	}

	return &FixResult{
		Success: success,
		Action:  "auto_fix_all",
		Output:  output.String(),
	}
}

func (af *AutoFixer) FormatCode() *FixResult {
	result := &FixResult{Action: "format"}

	// Detect project type and format accordingly
	if af.hasFile("package.json") {
		// Node.js/TypeScript
		if af.hasCommand("npx prettier") {
			out, err := af.run("npx prettier --write .")
			result.Success = err == nil
			result.Output = out
			return result
		}
	}

	if af.hasFile("go.mod") {
		// Go
		out, err := af.run("gofmt -w .")
		result.Success = err == nil
		result.Output = out
		return result
	}

	if af.hasFile("pyproject.toml") || af.hasFile("setup.py") {
		// Python
		if af.hasCommand("black") {
			out, err := af.run("black .")
			result.Success = err == nil
			result.Output = out
			return result
		}
	}

	result.Success = true
	result.Output = "No formatter detected"
	return result
}

func (af *AutoFixer) FixDependencies() *FixResult {
	result := &FixResult{Action: "fix_deps"}

	if af.hasFile("package.json") {
		out, err := af.run("npm install")
		result.Success = err == nil
		result.Output = out
		if err != nil {
			result.Error = err.Error()
		}
		return result
	}

	if af.hasFile("go.mod") {
		out, err := af.run("go mod tidy")
		result.Success = err == nil
		result.Output = out
		if err != nil {
			result.Error = err.Error()
		}
		return result
	}

	if af.hasFile("requirements.txt") {
		out, err := af.run("pip install -r requirements.txt")
		result.Success = err == nil
		result.Output = out
		if err != nil {
			result.Error = err.Error()
		}
		return result
	}

	if af.hasFile("Cargo.toml") {
		out, err := af.run("cargo update")
		result.Success = err == nil
		result.Output = out
		if err != nil {
			result.Error = err.Error()
		}
		return result
	}

	result.Success = true
	result.Output = "No dependency manager detected"
	return result
}

func (af *AutoFixer) LintFix() *FixResult {
	result := &FixResult{Action: "lint_fix"}

	if af.hasFile("package.json") {
		// ESLint
		if af.hasCommand("npx eslint") {
			out, err := af.run("npx eslint --fix .")
			result.Success = err == nil
			result.Output = out
			return result
		}
	}

	if af.hasFile("go.mod") {
		// Go vet
		out, err := af.run("go vet ./...")
		result.Success = err == nil
		result.Output = out
		return result
	}

	if af.hasFile(".eslintrc") || af.hasFile("eslint.config.js") {
		out, err := af.run("npx eslint --fix .")
		result.Success = err == nil
		result.Output = out
		return result
	}

	result.Success = true
	result.Output = "No linter detected"
	return result
}

func (af *AutoFixer) TypeCheck() *FixResult {
	result := &FixResult{Action: "type_check"}

	if af.hasFile("tsconfig.json") {
		out, err := af.run("npx tsc --noEmit")
		result.Success = err == nil
		result.Output = out
		if err != nil {
			result.Error = err.Error()
		}
		return result
	}

	if af.hasFile("go.mod") {
		out, err := af.run("go build ./...")
		result.Success = err == nil
		result.Output = out
		if err != nil {
			result.Error = err.Error()
		}
		return result
	}

	result.Success = true
	result.Output = "No type checker detected"
	return result
}

func (af *AutoFixer) UpdateDependencies() *FixResult {
	result := &FixResult{Action: "update_deps"}

	if af.hasFile("package.json") {
		out, err := af.run("npm update")
		result.Success = err == nil
		result.Output = out
		if err != nil {
			result.Error = err.Error()
		}
		return result
	}

	if af.hasFile("go.mod") {
		out, err := af.run("go get -u ./...")
		result.Success = err == nil
		result.Output = out
		if err != nil {
			result.Error = err.Error()
		}
		return result
	}

	result.Success = true
	result.Output = "No dependency updater detected"
	return result
}

func (af *AutoFixer) hasFile(name string) bool {
	path := af.cwd + "/" + name
	_, err := os.Stat(path)
	return err == nil
}

func (af *AutoFixer) hasCommand(name string) bool {
	parts := strings.Fields(name)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = af.cwd
	return cmd.Run() == nil
}

func (af *AutoFixer) run(cmd string) (string, error) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Dir = af.cwd
	output, err := execCmd.CombinedOutput()
	return string(output), err
}

type MergeResolver struct {
	cwd string
}

func NewMergeResolver(cwd string) *MergeResolver {
	return &MergeResolver{cwd: cwd}
}

type MergeResult struct {
	Resolved  bool
	Conflicts int
	Output    string
	Error     string
}

func (mr *MergeResolver) HasConflicts() bool {
	// Check if there are merge conflicts in git status
	out, err := mr.run("git status --porcelain")
	if err != nil {
		return false
	}

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "UU") || strings.Contains(line, "AA") || strings.Contains(line, "DD") {
			return true
		}
	}
	return false
}

func (mr *MergeResolver) GetConflicts() []string {
	out, err := mr.run("git diff --name-only --diff-filter=U")
	if err != nil {
		return nil
	}

	var conflicts []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			conflicts = append(conflicts, line)
		}
	}
	return conflicts
}

func (mr *MergeResolver) AutoResolve() *MergeResult {
	result := &MergeResult{}

	if !mr.HasConflicts() {
		result.Resolved = true
		result.Output = "No conflicts to resolve"
		return result
	}

	conflicts := mr.GetConflicts()
	result.Conflicts = len(conflicts)

	if result.Conflicts == 0 {
		result.Resolved = true
		return result
	}

	// Try git's built-in merge conflict resolution
	// Accept 'ours' or 'theirs' for simple cases
	_, err := mr.run("git status")
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Simple auto-resolve: accept ours for all
	for _, file := range conflicts {
		// Check if conflict is simple (no complex merges needed)
		content, err := os.ReadFile(mr.cwd + "/" + file)
		if err != nil {
			continue
		}

		// Count conflict markers
		conflictCount := strings.Count(string(content), "<<<<<<<")

		// If there's only one conflict marker, it's simple
		if conflictCount == 1 {
			// Try to resolve by taking our changes
			mr.run("git checkout --ours " + file)
			mr.run("git add " + file)
		}
	}

	// Check remaining conflicts
	if mr.HasConflicts() {
		result.Conflicts = len(mr.GetConflicts())
		result.Error = fmt.Sprintf("%d conflicts remaining", result.Conflicts)
		return result
	}

	result.Resolved = true
	result.Output = "All conflicts resolved"
	return result
}

func (mr *MergeResolver) run(cmd string) (string, error) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	execCmd := exec.Command(parts[0], parts[1:]...)
	execCmd.Dir = mr.cwd
	output, err := execCmd.CombinedOutput()
	return string(output), err
}
