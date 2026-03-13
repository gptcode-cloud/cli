package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Issue represents a GitHub issue
type Issue struct {
	Number     int      `json:"number"`
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	State      string   `json:"state"`
	Labels     []string `json:"labels"`
	Author     string   `json:"author"`
	URL        string   `json:"url"`
	Assignees  []string `json:"assignees"`
	Milestone  string   `json:"milestone"`
	CreatedAt  string   `json:"createdAt"`
	UpdatedAt  string   `json:"updatedAt"`
	Comments   int      `json:"comments"`
	Repository string   `json:"repository"`
}

// Client represents a GitHub client using gh CLI
type Client struct {
	repo    string // owner/repo format
	workDir string
}

func (c *Client) SetWorkDir(dir string) {
	c.workDir = dir
}

// HasExistingPR checks if there's already a PR for this issue
func (c *Client) HasExistingPR(issueNum int) (bool, error) {
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/issues/%d", c.repo, issueNum), "--jq", ".pull_request")
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		// No PR exists if we get an error
		return false, nil
	}
	return strings.Contains(string(output), "pull_request"), nil
}

// NewClient creates a new GitHub client
func NewClient(repo string) *Client {
	return &Client{repo: repo}
}

// FetchIssue fetches a GitHub issue by number
func (c *Client) FetchIssue(issueNumber int) (*Issue, error) {
	// Use gh CLI to fetch issue details in JSON format
	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(issueNumber),
		"--json", "number,title,body,state,labels,author,url,assignees,milestone,createdAt,updatedAt",
		"--repo", c.repo)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue #%d: %w\nOutput: %s", issueNumber, err, string(output))
	}

	var rawIssue struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		State  string `json:"state"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		URL       string `json:"url"`
		Assignees []struct {
			Login string `json:"login"`
		} `json:"assignees"`
		Milestone struct {
			Title string `json:"title"`
		} `json:"milestone"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}

	if err := json.Unmarshal(output, &rawIssue); err != nil {
		return nil, fmt.Errorf("failed to parse issue JSON: %w", err)
	}

	// Convert to our Issue struct
	issue := &Issue{
		Number:     rawIssue.Number,
		Title:      rawIssue.Title,
		Body:       rawIssue.Body,
		State:      rawIssue.State,
		Author:     rawIssue.Author.Login,
		URL:        rawIssue.URL,
		CreatedAt:  rawIssue.CreatedAt,
		UpdatedAt:  rawIssue.UpdatedAt,
		Comments:   0,
		Repository: c.repo,
	}

	// Extract label names
	for _, label := range rawIssue.Labels {
		issue.Labels = append(issue.Labels, label.Name)
	}

	// Extract assignee logins
	for _, assignee := range rawIssue.Assignees {
		issue.Assignees = append(issue.Assignees, assignee.Login)
	}

	// Set milestone
	issue.Milestone = rawIssue.Milestone.Title

	return issue, nil
}

// ExtractRequirements extracts actionable requirements from issue
func (i *Issue) ExtractRequirements() []string {
	requirements := []string{}

	// Simple extraction: look for task lists, numbered lists, bullet points
	lines := strings.Split(i.Body, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Task list items: - [ ] or - [x]
		if strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "- [x]") {
			req := strings.TrimPrefix(trimmed, "- [ ]")
			req = strings.TrimPrefix(req, "- [x]")
			req = strings.TrimSpace(req)
			if req != "" {
				requirements = append(requirements, req)
			}
			continue
		}

		// Bullet points
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			req := strings.TrimPrefix(trimmed, "- ")
			req = strings.TrimPrefix(req, "* ")
			req = strings.TrimSpace(req)
			if req != "" && len(req) > 10 { // Skip very short bullets
				requirements = append(requirements, req)
			}
			continue
		}

		// Numbered lists: 1. 2. etc
		if len(trimmed) > 3 && trimmed[0] >= '0' && trimmed[0] <= '9' && trimmed[1] == '.' {
			req := trimmed[3:]
			req = strings.TrimSpace(req)
			if req != "" {
				requirements = append(requirements, req)
			}
		}
	}

	// If no structured requirements found, use title as requirement
	if len(requirements) == 0 {
		requirements = append(requirements, i.Title)
	}

	return requirements
}

// ParseReferences extracts issue/PR references from text
func ParseReferences(text string) []string {
	references := []string{}

	// Match #123 pattern
	words := strings.Fields(text)
	for _, word := range words {
		if strings.HasPrefix(word, "#") && len(word) > 1 {
			// Remove trailing punctuation
			ref := strings.TrimRight(word, ".,;:!?")
			// Verify it's a number after #
			numPart := ref[1:]
			if _, err := strconv.Atoi(numPart); err == nil {
				references = append(references, ref)
			}
		}
	}

	return references
}

// CreateBranchName generates a branch name from issue
func (i *Issue) CreateBranchName() string {
	// Format: issue-123-short-description
	title := strings.ToLower(i.Title)

	// Replace spaces and special chars with hyphens
	title = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, title)

	// Remove consecutive hyphens
	for strings.Contains(title, "--") {
		title = strings.ReplaceAll(title, "--", "-")
	}

	// Trim hyphens from ends
	title = strings.Trim(title, "-")

	// Limit length
	if len(title) > 50 {
		title = title[:50]
		title = strings.Trim(title, "-")
	}

	return fmt.Sprintf("issue-%d-%s", i.Number, title)
}
