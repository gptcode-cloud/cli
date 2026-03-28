package cicd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Client struct {
	repo  string
	owner string
	name  string
	token string
}

func NewClient(repo string) *Client {
	parts := strings.Split(repo, "/")
	if len(parts) >= 2 {
		return &Client{
			repo:  repo,
			owner: parts[len(parts)-2],
			name:  parts[len(parts)-1],
			token: os.Getenv("GITHUB_TOKEN"),
		}
	}
	return &Client{repo: repo}
}

type StatusCheck struct {
	Context   string `json:"context"`
	State     string `json:"state"`
	TargetURL string `json:"target_url"`
}

func (c *Client) GetStatusChecks(ctx context.Context, ref string) ([]StatusCheck, error) {
	output, err := exec.Command("gh", "api", fmt.Sprintf("/repos/%s/%s/commits/%s/statuses", c.owner, c.name, ref)).Output()
	if err != nil {
		return nil, err
	}

	var statuses []StatusCheck
	json.Unmarshal(output, &statuses)
	return statuses, nil
}

func (c *Client) GetPRStatus(ctx context.Context) (*PRStatus, error) {
	output, err := exec.Command("gh", "pr", "status", "--json", "number,title,state,headRefName,url").Output()
	if err != nil {
		return nil, err
	}

	var prs []PRInfo
	json.Unmarshal(output, &prs)

	return &PRStatus{PRs: prs}, nil
}

type PRInfo struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HeadRef string `json:"headRefName"`
	URL     string `json:"url"`
}

type PRStatus struct {
	PRs []PRInfo
}

type AutoMerge struct {
	client *Client
	ctx    context.Context
}

func NewAutoMerge(repo string) *AutoMerge {
	return &AutoMerge{
		client: NewClient(repo),
		ctx:    context.Background(),
	}
}

func (am *AutoMerge) CanAutoMerge() (bool, error) {
	output, err := exec.Command("gh", "api", "graphql", "-f", fmt.Sprintf(`query { repository(owner: "%s", name: "%s") { mergeQueue { method } } }`, am.client.owner, am.client.name)).Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), " SQUASH"), nil
}

func (am *AutoMerge) EnableAutoMerge(prNumber int) error {
	cmd := exec.Command("gh", "pr", "merge", "--admin", "--squash", "--delete-branch", fmt.Sprintf("%d", prNumber))
	_, err := cmd.CombinedOutput()
	return err
}

func (am *AutoMerge) WaitForCI(prNumber int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		statuses, err := am.client.GetStatusChecks(am.ctx, fmt.Sprintf("PR-%d", prNumber))
		if err != nil {
			return err
		}

		allPassed := true
		for _, s := range statuses {
			if s.State != "success" {
				allPassed = false
				break
			}
		}

		if allPassed && len(statuses) > 0 {
			return nil
		}

		time.Sleep(30 * time.Second)
	}

	return fmt.Errorf("timeout waiting for CI")
}

type CIDeploy struct {
	client *Client
}

func NewCIDeploy(repo string) *CIDeploy {
	return &CIDeploy{client: NewClient(repo)}
}

type DeployResult struct {
	URL         string
	ID          string
	Environment string
	Status      string
}

func (cd *CIDeploy) Deploy(ctx context.Context, ref string, environment string) (*DeployResult, error) {
	cmd := exec.Command("gh", "api", "repos/"+cd.client.repo+"/actions/runs",
		"--method", "POST",
		"-f", fmt.Sprintf(`{"ref":"%s","environment":"%s"}`, ref, environment))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var run struct {
		ID          int    `json:"id"`
		HTMLURL     string `json:"html_url"`
		Environment string `json:"environment"`
		Status      string `json:"status"`
	}
	json.Unmarshal(output, &run)

	return &DeployResult{
		URL:         run.HTMLURL,
		ID:          fmt.Sprintf("%d", run.ID),
		Environment: run.Environment,
		Status:      run.Status,
	}, nil
}

func (cd *CIDeploy) GetDeployments(ctx context.Context) ([]Deployment, error) {
	cmd := exec.Command("gh", "api", "repos/"+cd.client.repo+"/deployments")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var deployments []Deployment
	json.Unmarshal(output, &deployments)
	return deployments, nil
}

type Deployment struct {
	ID          int    `json:"id"`
	SHA         string `json:"sha"`
	Ref         string `json:"ref"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
	Creator     struct {
		Login string `json:"login"`
	} `json:"creator"`
}

func (cd *CIDeploy) GetDeploymentStatus(ctx context.Context, deploymentID int) (*DeploymentStatus, error) {
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/deployments/%d/statuses", cd.client.repo, deploymentID))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var statuses []DeploymentStatus
	json.Unmarshal(output, &statuses)
	if len(statuses) > 0 {
		return &statuses[0], nil
	}
	return nil, fmt.Errorf("no status found")
}

type DeploymentStatus struct {
	ID          int    `json:"id"`
	State       string `json:"state"`
	Description string `json:"description"`
	Environment string `json:"environment"`
	Creator     struct {
		Login string `json:"login"`
	} `json:"creator"`
}

type CIWorkflow struct {
	client *Client
}

func NewCIWorkflow(repo string) *CIWorkflow {
	return &CIWorkflow{client: NewClient(repo)}
}

func (cw *CIWorkflow) ListWorkflows() ([]Workflow, error) {
	cmd := exec.Command("gh", "api", "repos/"+cw.client.repo+"/actions/workflows")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result struct {
		Workflows []Workflow `json:"workflows"`
	}
	json.Unmarshal(output, &result.Workflows)
	return result.Workflows, nil
}

type Workflow struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	State     string `json:"state"`
	UpdatedAt string `json:"updated_at"`
}

func (cw *CIWorkflow) TriggerWorkflow(name string, inputs map[string]string) error {
	inputStr := ""
	for k, v := range inputs {
		inputStr += fmt.Sprintf(" -f %s=%s", k, v)
	}
	cmd := exec.Command("sh", "-c", fmt.Sprintf("gh workflow run %s%s", name, inputStr))
	_, err := cmd.CombinedOutput()
	return err
}

func (cw *CIWorkflow) GetWorkflowRuns(name string) ([]WorkflowRun, error) {
	cmd := exec.Command("gh", "run", "list", "--workflow", name, "--json", "number,title,status,conclusion,databaseId")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var runs []WorkflowRun
	json.Unmarshal(output, &runs)
	return runs, nil
}

type WorkflowRun struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	ID         int    `json:"databaseId"`
}

func (cw *CIWorkflow) WatchRun(runID int, timeout time.Duration) error {
	cmd := exec.Command("gh", "run", "watch", fmt.Sprintf("%d", runID), "--exit-status")
	cmd.Dir = os.Getenv("PWD")
	return cmd.Run()
}

type PRComment struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	Author    string `json:"author"`
	CreatedAt string `json:"created_at"`
}

func (c *Client) AddPRComment(prNumber int, body string) error {
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/%s/issues/%d/comments", c.owner, c.name, prNumber),
		"--method", "POST",
		"-f", fmt.Sprintf(`{"body":"%s"}`, body))
	_, err := cmd.CombinedOutput()
	return err
}

func (c *Client) UpdatePRDescription(prNumber int, body string) error {
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/%s/issues/%d", c.owner, c.name, prNumber),
		"--method", "PATCH",
		"-f", fmt.Sprintf(`{"body":"%s"}`, body))
	_, err := cmd.CombinedOutput()
	return err
}
