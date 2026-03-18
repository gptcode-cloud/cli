package autonomous

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type GitManager struct {
	cwd string
}

func NewGitManager(cwd string) *GitManager {
	return &GitManager{cwd: cwd}
}

type CommitResult struct {
	Success   bool
	Message   string
	Files     []string
	CommitSHA string
}

// HasChanges checks if there are uncommitted changes
func (g *GitManager) HasChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.cwd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// GetChangedFiles returns list of changed files
func (g *GitManager) GetChangedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = g.cwd
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, f := range files {
		if f != "" {
			result = append(result, f)
		}
	}
	return result, nil
}

// GetStagedFiles returns list of staged files
func (g *GitManager) GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = g.cwd
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --cached failed: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, f := range files {
		if f != "" {
			result = append(result, f)
		}
	}
	return result, nil
}

// AutoCommit creates an automatic commit with a descriptive message
func (g *GitManager) AutoCommit(context context.Context, taskDescription string) (*CommitResult, error) {
	result := &CommitResult{}

	// Check for changes
	hasChanges, err := g.HasChanges()
	if err != nil {
		return nil, err
	}
	if !hasChanges {
		result.Success = true
		result.Message = "No changes to commit"
		return result, nil
	}

	// Stage all changes
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = g.cwd
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git add failed: %w", err)
	}

	// Get staged files
	stagedFiles, err := g.GetStagedFiles()
	if err != nil {
		return nil, err
	}
	result.Files = stagedFiles

	// Generate commit message
	commitMsg := g.GenerateCommitMessage(taskDescription, stagedFiles)

	// Commit
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = g.cwd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git commit failed: %w: %s", err, string(output))
	}

	result.Success = true
	result.Message = commitMsg

	// Get commit SHA
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = g.cwd
	sha, err := cmd.Output()
	if err == nil {
		result.CommitSHA = strings.TrimSpace(string(sha))
	}

	return result, nil
}

// GenerateCommitMessage creates a commit message based on changes
func (g *GitManager) GenerateCommitMessage(task string, files []string) string {
	// Analyze file types
	var fileTypes []string
	var hasTests, hasDocs, hasFix bool

	for _, f := range files {
		lower := strings.ToLower(f)
		if strings.HasSuffix(f, "_test.go") || strings.HasSuffix(f, ".test.ts") || strings.HasSuffix(f, ".spec.ts") {
			hasTests = true
		} else if strings.HasSuffix(f, ".md") || strings.HasSuffix(f, ".txt") {
			hasDocs = true
		} else if strings.Contains(lower, "fix") || strings.Contains(lower, "bug") {
			hasFix = true
		}

		ext := getExtension(f)
		if ext != "" && !contains(fileTypes, ext) {
			fileTypes = append(fileTypes, ext)
		}
	}

	// Generate message
	var prefix string
	if hasFix {
		prefix = "fix"
	} else if hasTests && !hasDocs {
		prefix = "test"
	} else if hasDocs && !hasTests {
		prefix = "docs"
	} else {
		prefix = "chore"
	}

	// Clean task description
	task = cleanTaskDescription(task)

	// Build message
	msg := fmt.Sprintf("%s: %s", prefix, task)

	if len(fileTypes) > 0 {
		msg += fmt.Sprintf(" [%s]", strings.Join(fileTypes, ", "))
	}

	return msg
}

// HasRemote checks if there's a remote to push to
func (g *GitManager) HasRemote() bool {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = g.cwd
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// Push pushes commits to remote
func (g *GitManager) Push() error {
	cmd := exec.Command("git", "push")
	cmd.Dir = g.cwd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w: %s", err, string(output))
	}
	return nil
}

// GetStatus returns current git status
func (g *GitManager) GetStatus() (string, error) {
	cmd := exec.Command("git", "status")
	cmd.Dir = g.cwd
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}
	return string(output), nil
}

// GetCurrentBranch returns the current branch name
func (g *GitManager) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.cwd
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CreateBranch creates a new branch
func (g *GitManager) CreateBranch(name string) error {
	cmd := exec.Command("git", "checkout", "-b", name)
	cmd.Dir = g.cwd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout -b failed: %w: %s", err, string(output))
	}
	return nil
}

func getExtension(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func cleanTaskDescription(task string) string {
	// Remove common prefixes
	task = strings.TrimPrefix(task, "fix: ")
	task = strings.TrimPrefix(task, "fix ")
	task = strings.TrimPrefix(task, "Fix ")

	// Truncate if too long
	if len(task) > 72 {
		task = task[:69] + "..."
	}

	return task
}

// TaskCommitter helps with automatic git operations for tasks
type TaskCommitter struct {
	git *GitManager
}

func NewTaskCommitter(cwd string) *TaskCommitter {
	return &TaskCommitter{
		git: NewGitManager(cwd),
	}
}

// CommitTask automatically commits with appropriate message
func (tc *TaskCommitter) CommitTask(ctx context.Context, task string) (*CommitResult, error) {
	return tc.git.AutoCommit(ctx, task)
}

// ShouldCommit determines if a commit is worthwhile
func (tc *TaskCommitter) ShouldCommit() (bool, error) {
	hasChanges, err := tc.git.HasChanges()
	if err != nil || !hasChanges {
		return false, err
	}

	// Don't commit if only lock files or generated files changed
	files, err := tc.git.GetChangedFiles()
	if err != nil {
		return false, err
	}

	skipFiles := map[string]bool{
		"package-lock.json": true,
		"yarn.lock":         true,
		"go.sum":            true,
		"Gemfile.lock":      true,
		"pnpm-lock.yaml":    true,
	}

	for _, f := range files {
		// Skip if any meaningful file changed
		if !skipFiles[f] && !strings.HasSuffix(f, ".lock") {
			return true, nil
		}
	}

	return false, nil
}

// AutoGitWorkflow runs a complete git workflow: commit and optionally push
func (tc *TaskCommitter) AutoGitWorkflow(ctx context.Context, task string, push bool) (*CommitResult, error) {
	// Check if we should commit
	shouldCommit, err := tc.ShouldCommit()
	if err != nil {
		return nil, err
	}
	if !shouldCommit {
		return &CommitResult{
			Success: true,
			Message: "No meaningful changes to commit",
		}, nil
	}

	// Commit
	result, err := tc.git.AutoCommit(ctx, task)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return result, nil
	}

	// Push if requested and available
	if push && tc.git.HasRemote() {
		if err := tc.git.Push(); err != nil {
			result.Message += " (push failed)"
		} else {
			result.Message += " (pushed)"
		}
	}

	return result, nil
}

// GetTimestamp returns current timestamp for commit messages
func GetTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
