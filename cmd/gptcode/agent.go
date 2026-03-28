package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"gptcode/internal/live"
	"gptcode/internal/llm"
	"gptcode/internal/modes"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run gptcode as a centralized autofix agent",
	Long: `This command is intended to be run exclusively by the gptcode-live centralized agent system.
It expects specific environment variables to authenticate and clone a shallow copy of the repository,
fetch the CI logs, run the fix, and report back via WebSocket.`,
	RunE: runAgent,
}

func init() {
	rootCmd.AddCommand(agentCmd)
}

func runAgent(cmd *cobra.Command, args []string) error {
	repo := os.Getenv("REPO")
	branch := os.Getenv("BRANCH")
	sha := os.Getenv("SHA")
	runID := os.Getenv("RUN_ID")
	token := os.Getenv("INSTALL_TOKEN")
	liveURL := os.Getenv("GPTCODE_LIVE_URL")
	agentID := os.Getenv("AGENT_ID")

	if repo == "" || branch == "" || sha == "" || runID == "" || token == "" {
		return fmt.Errorf("missing required agent environment variables (REPO, BRANCH, SHA, RUN_ID, INSTALL_TOKEN)")
	}

	fmt.Printf("🚀 Starting AutoFix Agent for %s@%s (Run %s)\n", repo, branch, runID)
	if liveURL != "" {
		fmt.Printf("📡 Reporting progress to %s with agent_id %s\n", liveURL, agentID)
	}

	// 1. Shallow clone the target repo using the installation token
	cloneDir := "/work"
	// Ensure directory doesn't exist from a previous run
	os.RemoveAll(cloneDir)

	repoURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", token, repo)

	fmt.Printf("📦 Cloning %s (branch: %s) into %s...\n", repo, branch, cloneDir)
	cmdClone := exec.Command("git", "clone", "--depth", "1", "--branch", branch, repoURL, cloneDir)
	cmdClone.Stdout = os.Stdout
	cmdClone.Stderr = os.Stderr
	err := cmdClone.Run()
	if err != nil {
		// Fallback: try 'master' if the specified branch doesn't exist
		fmt.Printf("⚠️  Branch '%s' not found, trying 'master'...\n", branch)
		os.RemoveAll(cloneDir)
		cmdClone2 := exec.Command("git", "clone", "--depth", "1", "--branch", "master", repoURL, cloneDir)
		cmdClone2.Stdout = os.Stdout
		cmdClone2.Stderr = os.Stderr
		err = cmdClone2.Run()
		if err != nil {
			// Final fallback: clone default branch
			fmt.Printf("⚠️  Branch 'master' not found either, cloning default branch...\n")
			os.RemoveAll(cloneDir)
			cmdClone3 := exec.Command("git", "clone", "--depth", "1", repoURL, cloneDir)
			cmdClone3.Stdout = os.Stdout
			cmdClone3.Stderr = os.Stderr
			err = cmdClone3.Run()
			if err != nil {
				fmt.Printf("❌ Failed to clone repository: %v\n", err)
				return err
			}
		}
	}
	defer os.RemoveAll(cloneDir) // Clean up if running locally, though ephemeral VMs will just die

	// Set GH_TOKEN so the gh CLI can create PRs
	os.Setenv("GH_TOKEN", token)
	// Configure git identity for commits
	exec.Command("git", "config", "--global", "user.email", "gptcode-agent@zapfy.ai").Run()
	exec.Command("git", "config", "--global", "user.name", "GPTCode Agent").Run()

	// Change to the working directory
	err = os.Chdir(cloneDir)
	if err != nil {
		return fmt.Errorf("failed to change to working directory %s: %w", cloneDir, err)
	}

	// 2. Extract contextual instruction
	agentType := os.Getenv("AGENT_TYPE")
	if agentType == "" {
		agentType = "ci"
	}

	// Skip validation for Sentry agents — CI/CD will handle compilation
	if agentType == "sentry" {
		os.Setenv("SKIP_VALIDATION", "1")
	}

	var prompt string
	if agentType == "sentry" {
		sentryTitle := os.Getenv("SENTRY_TITLE")
		sentryStack := os.Getenv("SENTRY_STACKTRACE")

		// Detect default branch from the cloned repo
		defaultBranch := "main"
		if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
			if b := strings.TrimSpace(string(out)); b != "" {
				defaultBranch = b
			}
		}

		branchName := fmt.Sprintf("fix/sentry-%s", runID)

		prompt = fmt.Sprintf(`You are a production bug-fix agent. A runtime error was caught by Sentry.

ERROR: %s

STACKTRACE:
%s

INSTRUCTIONS — follow EXACTLY in order:

1. Use search_code to find the file mentioned in the stacktrace.
2. Use read_file to read that file.
3. Identify the bug. Apply a MINIMAL fix using apply_patch or write_file.
4. Use run_command to create a branch, commit, push, and open a PR:

   git checkout -b %s && \
   git add -A && \
   git commit -m "fix: %s" && \
   git push origin %s && \
   gh pr create --title "fix: %s" --body "Automated fix for Sentry error: %s" --base %s

RULES:
- Fix ONE file only. Do NOT refactor, do NOT write tests.
- Do NOT run mix, npm, go build, or any build/test commands.
- Do NOT explore the project. Go directly to the file from the stacktrace.
- If you cannot identify the file from the stacktrace, use search_code with a key term from the error.
- You MUST create the PR. The task is NOT complete until the PR is created.`,
			sentryTitle, sentryStack,
			branchName, sentryTitle, branchName,
			sentryTitle, sentryTitle, defaultBranch)
	} else {
		// Default to CI pipeline logic
		prompt = fmt.Sprintf(`The CI pipeline just failed for this repository on branch '%s'.
The failing commit is %s.

Please discover what failed by examining the code, tests, and standard CI configurations
(such as npm test, mix test, go test, etc.) and write a fix. Once fixed, create a PR.`, branch, sha)
	}

	ctx := context.Background()

	backendName := "openrouter"
	baseURL := "https://openrouter.ai/api/v1"
	provider := llm.NewChatCompletion(baseURL, backendName)

	language := "elixir" // Most Zapfy repos are Elixir
	queryModel := os.Getenv("MODEL")
	if queryModel == "" {
		queryModel = "anthropic/claude-sonnet-4"
	}

	liveClient := live.GetClient()
	var reportConfig *live.ReportConfig
	if liveURL != "" {
		reportConfig = live.DefaultReportConfig()
		reportConfig.SetBaseURL(liveURL)
		reportConfig.AgentID = agentID

		// Register on the Live Dashboard BEFORE starting execution
		taskLabel := "Sentry Exception Repair"
		if agentType != "sentry" {
			taskLabel = "CI Breakdown Repair"
		}
		reportConfig.Model = queryModel
		if err := reportConfig.Connect(agentID, agentType, taskLabel); err != nil {
			fmt.Printf("⚠️  Live Dashboard connect failed (non-fatal): %v\n", err)
		} else {
			fmt.Printf("📡 Connected to Live Dashboard as %s\n", agentID)
		}
	} else if liveClient != nil {
		reportConfig = live.DefaultReportConfig()
		reportConfig.AgentID = live.GetAgentID()
	}

	executor := modes.NewAutonomousExecutorWithLive(provider, ".", queryModel, language, liveClient, reportConfig, backendName)

	fmt.Println("🤖 Starting AutoFix process...")
	err = executor.Execute(ctx, prompt)
	if err != nil {
		fmt.Printf("❌ AutoFix failed: %v\n", err)
		return err
	}

	fmt.Println("✅ AutoFix completed successfully.")
	return nil
}
