package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type PullRequest struct {
	Number     int      `json:"number"`
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	State      string   `json:"state"`
	HeadBranch string   `json:"headRefName"`
	BaseBranch string   `json:"baseRefName"`
	URL        string   `json:"url"`
	Author     string   `json:"author"`
	Labels     []string `json:"labels"`
	Assignees  []string `json:"assignees"`
	Reviewers  []string `json:"reviewers"`
	IsDraft    bool     `json:"isDraft"`
	Repository string   `json:"repository"`
}

type ReviewComment struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
}

type Review struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	State     string `json:"state"`
	Body      string `json:"body"`
	Comments  []ReviewComment
	CreatedAt string `json:"createdAt"`
}

type CommitOptions struct {
	Message     string
	IssueNumber int
	FilePaths   []string
	AllFiles    bool
}

func (c *Client) CreateBranch(branchName string, fromBranch string) error {
	if fromBranch == "" {
		fromBranch = c.DetectDefaultBranch()
	}

	checkoutCmd := exec.Command("git", "checkout", "-b", branchName, fromBranch)
	if c.workDir != "" {
		checkoutCmd.Dir = c.workDir
	}
	output, err := checkoutCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "already exists") {
			checkoutCmd = exec.Command("git", "checkout", branchName)
			if c.workDir != "" {
				checkoutCmd.Dir = c.workDir
			}
			output, err = checkoutCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to checkout existing branch %s: %w\nOutput: %s", branchName, err, string(output))
			}
			return nil
		}
		return fmt.Errorf("failed to create branch %s: %w\nOutput: %s", branchName, err, string(output))
	}

	return nil
}

// DetectDefaultBranch detects the default branch (main, master, trunk, etc.)
func (c *Client) DetectDefaultBranch() string {
	// Use git rev-parse to get current branch name
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}
	output, err := cmd.CombinedOutput()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		if branch != "HEAD" {
			return branch
		}
	}

	// Fallback: try common default branch names
	for _, branch := range []string{"main", "master", "trunk", "develop", "dev"} {
		cmd := exec.Command("git", "rev-parse", "--verify", branch)
		if c.workDir != "" {
			cmd.Dir = c.workDir
		}
		if _, err := cmd.CombinedOutput(); err == nil {
			return branch
		}
	}

	// Final fallback
	return "main"
}

// GetForkRemote determines if we're in a fork and returns the correct remote to push to
func (c *Client) GetForkRemote() (remoteName string, isFork bool) {
	// Get the current user from gh
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "origin", false
	}
	currentUser := strings.TrimSpace(string(output))

	// Parse the repo to get owner
	parts := strings.Split(c.repo, "/")
	if len(parts) != 2 {
		return "origin", false
	}
	repoOwner := parts[0]

	// If the repo owner matches the current user, we're not in a fork scenario
	// or we're working with our own repo
	if repoOwner == currentUser {
		return "origin", false
	}

	// Check if there's a remote for the current user (fork)
	cmd = exec.Command("git", "remote", "-v")
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "origin", false
	}

	// Look for a remote that belongs to the current user
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "(push)") && strings.Contains(line, currentUser) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[0], true
			}
		}
	}

	// If no fork remote found, try to add one
	if repoOwner != currentUser {
		// The original repo is the "upstream" or "origin"
		// We need to add the user's fork as a remote
		forkURL := fmt.Sprintf("git@github.com:%s/%s.git", currentUser, parts[1])
		addCmd := exec.Command("git", "remote", "add", "fork", forkURL)
		if c.workDir != "" {
			addCmd.Dir = c.workDir
		}
		if _, err := addCmd.CombinedOutput(); err == nil {
			return "fork", true
		}
	}

	return "origin", false
}

func (c *Client) CommitChanges(opts CommitOptions) error {
	if opts.AllFiles {
		addCmd := exec.Command("git", "add", "-A")
		if c.workDir != "" {
			addCmd.Dir = c.workDir
		}
		if output, err := addCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stage all files: %w\nOutput: %s", err, string(output))
		}
	} else if len(opts.FilePaths) > 0 {
		args := append([]string{"add"}, opts.FilePaths...)
		addCmd := exec.Command("git", args...)
		if c.workDir != "" {
			addCmd.Dir = c.workDir
		}
		if output, err := addCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stage files: %w\nOutput: %s", err, string(output))
		}
	}

	commitMsg := opts.Message
	if opts.IssueNumber > 0 {
		commitMsg = fmt.Sprintf("%s\n\nCloses #%d", commitMsg, opts.IssueNumber)
	}

	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	if c.workDir != "" {
		commitCmd.Dir = c.workDir
	}
	output, err := commitCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "nothing to commit") {
			return nil
		}
		return fmt.Errorf("failed to commit: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *Client) PushBranch(branchName string) error {
	remote, _ := c.GetForkRemote()
	pushCmd := exec.Command("git", "push", "-u", remote, branchName)
	if c.workDir != "" {
		pushCmd.Dir = c.workDir
	}
	output, err := pushCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push branch %s: %w\nOutput: %s", branchName, err, string(output))
	}

	return nil
}

func (c *Client) CreatePR(opts PRCreateOptions) (*PullRequest, error) {
	args := []string{"pr", "create"}

	if opts.Title != "" {
		args = append(args, "--title", opts.Title)
	}

	if opts.Body != "" {
		args = append(args, "--body", opts.Body)
	}

	if opts.BaseBranch != "" {
		args = append(args, "--base", opts.BaseBranch)
	}

	if opts.IsDraft {
		args = append(args, "--draft")
	}

	for _, label := range opts.Labels {
		args = append(args, "--label", label)
	}

	for _, assignee := range opts.Assignees {
		args = append(args, "--assignee", assignee)
	}

	for _, reviewer := range opts.Reviewers {
		args = append(args, "--reviewer", reviewer)
	}

	args = append(args, "--repo", c.repo)

	cmd := exec.Command("gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w\nOutput: %s", err, string(output))
	}

	prURL := strings.TrimSpace(string(output))

	parts := strings.Split(prURL, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid PR URL format: %s", prURL)
	}

	prNumber := 0
	if numStr := parts[len(parts)-1]; numStr != "" {
		prNumber, _ = strconv.Atoi(numStr)
	}

	return &PullRequest{
		Number:     prNumber,
		Title:      opts.Title,
		Body:       opts.Body,
		URL:        prURL,
		HeadBranch: opts.HeadBranch,
		BaseBranch: opts.BaseBranch,
		IsDraft:    opts.IsDraft,
		Labels:     opts.Labels,
		Assignees:  opts.Assignees,
		Reviewers:  opts.Reviewers,
		Repository: c.repo,
	}, nil
}

type PRCreateOptions struct {
	Title      string
	Body       string
	HeadBranch string
	BaseBranch string
	IsDraft    bool
	Labels     []string
	Assignees  []string
	Reviewers  []string
}

func (c *Client) AddLabelsToPR(prNumber int, labels []string) error {
	for _, label := range labels {
		cmd := exec.Command("gh", "pr", "edit", strconv.Itoa(prNumber), "--add-label", label, "--repo", c.repo)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add label %s to PR #%d: %w\nOutput: %s", label, prNumber, err, string(output))
		}
	}
	return nil
}

func (c *Client) AddReviewersToPR(prNumber int, reviewers []string) error {
	for _, reviewer := range reviewers {
		cmd := exec.Command("gh", "pr", "edit", strconv.Itoa(prNumber), "--add-reviewer", reviewer, "--repo", c.repo)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add reviewer %s to PR #%d: %w\nOutput: %s", reviewer, prNumber, err, string(output))
		}
	}
	return nil
}

func (c *Client) FetchPRReviews(prNumber int) ([]Review, error) {
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(prNumber),
		"--json", "reviews",
		"--repo", c.repo)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR reviews: %w\nOutput: %s", err, string(output))
	}

	var result struct {
		Reviews []struct {
			ID     string `json:"id"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			State     string `json:"state"`
			Body      string `json:"body"`
			CreatedAt string `json:"createdAt"`
		} `json:"reviews"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse reviews: %w", err)
	}

	var reviews []Review
	for _, r := range result.Reviews {
		reviews = append(reviews, Review{
			ID:        r.ID,
			Author:    r.Author.Login,
			State:     r.State,
			Body:      r.Body,
			CreatedAt: r.CreatedAt,
		})
	}

	return reviews, nil
}

func (c *Client) FetchPRComments(prNumber int) ([]ReviewComment, error) {
	cmd := exec.Command("gh", "api",
		fmt.Sprintf("/repos/%s/pulls/%d/comments", c.repo, prNumber))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR comments: %w\nOutput: %s", err, string(output))
	}

	var apiComments []struct {
		ID   int64 `json:"id"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
		Body      string `json:"body"`
		Path      string `json:"path"`
		Line      int    `json:"line"`
		State     string `json:"state"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(output, &apiComments); err != nil {
		return nil, fmt.Errorf("failed to parse comments: %w", err)
	}

	var comments []ReviewComment
	for _, c := range apiComments {
		comments = append(comments, ReviewComment{
			ID:        strconv.FormatInt(c.ID, 10),
			Author:    c.User.Login,
			Body:      c.Body,
			Path:      c.Path,
			Line:      c.Line,
			State:     c.State,
			CreatedAt: c.CreatedAt,
		})
	}

	return comments, nil
}

func (c *Client) GetUnresolvedComments(prNumber int) ([]ReviewComment, error) {
	comments, err := c.FetchPRComments(prNumber)
	if err != nil {
		return nil, err
	}

	var unresolved []ReviewComment
	for _, comment := range comments {
		if comment.State != "RESOLVED" && comment.Body != "" {
			unresolved = append(unresolved, comment)
		}
	}

	return unresolved, nil
}

func GeneratePRBody(issue *Issue, changes []string) string {
	body := fmt.Sprintf("Closes #%d\n\n", issue.Number)
	body += "## Changes\n\n"

	for _, change := range changes {
		body += fmt.Sprintf("- %s\n", change)
	}

	if len(issue.ExtractRequirements()) > 0 {
		body += "\n## Requirements Addressed\n\n"
		for _, req := range issue.ExtractRequirements() {
			body += fmt.Sprintf("- %s\n", req)
		}
	}

	return body
}
