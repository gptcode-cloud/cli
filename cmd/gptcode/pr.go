package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr [subcommand]",
	Short: "GitHub Pull Request operations",
}

var prCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a pull request from current branch",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPRCreate()
	},
}

func init() {
	prCmd.AddCommand(prCreateCmd)
}

func runPRCreate() error {
	// Check if git repo
	if !isGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	// Get current branch
	branch, err := getCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	// Check for uncommitted changes
	if hasUncommittedChanges() {
		fmt.Println("Warning: You have uncommitted changes")
		fmt.Println("Commit or stash them before creating a PR")
		return nil
	}

	// Check if branch has upstream
	hasUpstream, err := hasUpstreamBranch(branch)
	if err != nil {
		return fmt.Errorf("failed to check upstream: %w", err)
	}

	if !hasUpstream {
		// Push and set upstream
		fmt.Printf("Pushing branch '%s' and setting upstream...\n", branch)
		if err := pushWithUpstream(branch); err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}
	}

	// Get remote
	remote := getRemote()
	if remote == "" {
		return fmt.Errorf("no remote configured")
	}

	// Get repo info
	repo := getRepoInfo()
	if repo == "" {
		return fmt.Errorf("failed to determine repository")
	}

	// Generate PR title from branch name
	title := generatePRTitle(branch)

	// Open PR in browser or use gh cli
	fmt.Printf("Creating PR for '%s' to '%s'\n", branch, repo)

	// Try using gh CLI
	if isGhInstalled() {
		return createPRWithGh(repo, branch, title)
	}

	// Fallback to opening in browser
	url := fmt.Sprintf("https://github.com/%s/compare/main...%s?expand=1", repo, branch)
	fmt.Printf("Open this URL to create PR:\n%s\n", url)

	// Try to open automatically
	openCmd := exec.Command("open", url)
	if err := openCmd.Run(); err != nil {
		fmt.Println("Could not open browser automatically")
	}

	return nil
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir, _ = os.Getwd()
	return cmd.Run() == nil
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir, _ = os.Getwd()
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func hasUncommittedChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir, _ = os.Getwd()
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

func hasUpstreamBranch(branch string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	cmd.Dir, _ = os.Getwd()
	err := cmd.Run()
	return err == nil, nil
}

func pushWithUpstream(branch string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir, _ = os.Getwd()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("push failed: %s", string(output))
	}
	return nil
}

func getRemote() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir, _ = os.Getwd()
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse git@github.com:user/repo.git or https://github.com/user/repo
	url := strings.TrimSpace(string(output))
	url = strings.TrimPrefix(url, "git@github.com:")
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimSuffix(url, ".git")

	return url
}

func getRepoInfo() string {
	remote := getRemote()
	if remote == "" {
		// Try to get from git remote
		cmd := exec.Command("git", "remote", "-v")
		cmd.Dir, _ = os.Getwd()
		output, err := cmd.Output()
		if err != nil {
			return ""
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "origin") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					remote = parts[1]
					remote = strings.TrimPrefix(remote, "git@github.com:")
					remote = strings.TrimPrefix(remote, "https://github.com/")
					remote = strings.TrimSuffix(remote, ".git")
				}
			}
		}
	}
	return remote
}

func generatePRTitle(branch string) string {
	// Convert branch name to title
	title := strings.ReplaceAll(branch, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")

	// Capitalize words
	words := strings.Split(title, " ")
	var result []string
	for _, word := range words {
		if len(word) > 0 {
			result = append(result, strings.ToUpper(string(word[0]))+word[1:])
		}
	}
	return strings.Join(result, " ")
}

func isGhInstalled() bool {
	cmd := exec.Command("which", "gh")
	return cmd.Run() == nil
}

func createPRWithGh(repo, branch, title string) error {
	cmd := exec.Command("gh", "pr", "create",
		"--repo", repo,
		"--head", branch,
		"--title", title,
		"--fill")
	cmd.Dir, _ = os.Getwd()
	output, err := cmd.CombinedOutput()

	if err != nil {
		// PR might already exist
		if strings.Contains(string(output), "already exists") {
			fmt.Println("PR already exists for this branch")
			cmd = exec.Command("gh", "pr", "view", "--repo", repo, "--web")
			cmd.Run()
			return nil
		}
		return fmt.Errorf("failed to create PR: %s", string(output))
	}

	fmt.Printf("PR created successfully: %s\n", strings.TrimSpace(string(output)))
	return nil
}
