package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"chuchu/internal/github"
	"chuchu/internal/validation"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "GitHub issue management and automation",
	Long: `Manage GitHub issues and automate issue resolution.

Examples:
  chu issue fix 123              Fix issue #123 autonomously
  chu issue fix 123 --repo owner/repo  Fix issue from specific repo
  chu issue show 123             Show issue details`,
}

var issueFixCmd = &cobra.Command{
	Use:   "fix <issue-number>",
	Short: "Autonomously fix a GitHub issue",
	Long: `Fetch a GitHub issue, create a branch, implement the fix, run tests, 
and create a pull request.

This command will:
1. Fetch issue details from GitHub
2. Extract requirements
3. Create a branch (issue-N-description)
4. Analyze codebase and implement changes
5. Run tests and linters
6. Commit changes with issue reference
7. Push branch
8. Create pull request

Examples:
  chu issue fix 123                    Fix issue #123
  chu issue fix 123 --repo owner/repo Fix from specific repo
  chu issue fix 123 --draft           Create draft PR`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueNum, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[0])
		}

		repo, _ := cmd.Flags().GetString("repo")

		if repo == "" {
			repo = detectGitHubRepo()
			if repo == "" {
				return fmt.Errorf("could not detect GitHub repository. Use --repo flag")
			}
		}

		fmt.Printf("🔍 Fetching issue #%d from %s...\n\n", issueNum, repo)

		client := github.NewClient(repo)
		workDir, _ := os.Getwd()
		client.SetWorkDir(workDir)

		issue, err := client.FetchIssue(issueNum)
		if err != nil {
			return fmt.Errorf("failed to fetch issue: %w", err)
		}

		fmt.Printf("📋 Issue #%d: %s\n", issue.Number, issue.Title)
		fmt.Printf("   State: %s\n", issue.State)
		fmt.Printf("   Author: %s\n", issue.Author)
		if len(issue.Labels) > 0 {
			fmt.Printf("   Labels: %s\n", strings.Join(issue.Labels, ", "))
		}
		fmt.Println()

		reqs := issue.ExtractRequirements()
		if len(reqs) > 0 {
			fmt.Println("📝 Requirements:")
			for i, req := range reqs {
				fmt.Printf("   %d. %s\n", i+1, req)
			}
			fmt.Println()
		}

		branchName := issue.CreateBranchName()
		fmt.Printf("🌿 Creating branch: %s\n", branchName)
		
		if err := client.CreateBranch(branchName, ""); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}

		fmt.Println("\n⚠️  Manual implementation required (Symphony integration coming soon)")
		fmt.Println("   Please make your changes, then run:")
		fmt.Printf("   chu issue commit %d\n", issueNum)

		return nil
	},
}

var issueShowCmd = &cobra.Command{
	Use:   "show <issue-number>",
	Short: "Show GitHub issue details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueNum, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[0])
		}

		repo, _ := cmd.Flags().GetString("repo")
		if repo == "" {
			repo = detectGitHubRepo()
			if repo == "" {
				return fmt.Errorf("could not detect GitHub repository. Use --repo flag")
			}
		}

		client := github.NewClient(repo)
		issue, err := client.FetchIssue(issueNum)
		if err != nil {
			return fmt.Errorf("failed to fetch issue: %w", err)
		}

		fmt.Printf("Issue #%d: %s\n", issue.Number, issue.Title)
		fmt.Printf("State: %s\n", issue.State)
		fmt.Printf("Author: %s\n", issue.Author)
		fmt.Printf("URL: %s\n", issue.URL)
		fmt.Printf("Created: %s\n", issue.CreatedAt)
		fmt.Printf("Updated: %s\n", issue.UpdatedAt)
		
		if len(issue.Labels) > 0 {
			fmt.Printf("Labels: %s\n", strings.Join(issue.Labels, ", "))
		}
		
		if len(issue.Assignees) > 0 {
			fmt.Printf("Assignees: %s\n", strings.Join(issue.Assignees, ", "))
		}

		if issue.Body != "" {
			fmt.Printf("\nDescription:\n%s\n", issue.Body)
		}

		reqs := issue.ExtractRequirements()
		if len(reqs) > 0 {
			fmt.Println("\nExtracted Requirements:")
			for i, req := range reqs {
				fmt.Printf("%d. %s\n", i+1, req)
			}
		}

		return nil
	},
}

var issueCommitCmd = &cobra.Command{
	Use:   "commit <issue-number>",
	Short: "Commit changes with issue reference",
	Long: `Commit staged changes with proper issue reference and run validation.

This will:
1. Commit changes with "Closes #N" reference
2. Run tests (unless --skip-tests)
3. Run linters (unless --skip-lint)
4. Report validation results`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueNum, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[0])
		}

		message, _ := cmd.Flags().GetString("message")
		skipTests, _ := cmd.Flags().GetBool("skip-tests")
		skipLint, _ := cmd.Flags().GetBool("skip-lint")
		repo, _ := cmd.Flags().GetString("repo")

		if repo == "" {
			repo = detectGitHubRepo()
		}

		if message == "" {
			message = fmt.Sprintf("Fix issue #%d", issueNum)
		}

		workDir, _ := os.Getwd()
		client := github.NewClient(repo)
		client.SetWorkDir(workDir)

		fmt.Printf("💾 Committing changes for issue #%d...\n", issueNum)
		
		err = client.CommitChanges(github.CommitOptions{
			Message:     message,
			IssueNumber: issueNum,
			AllFiles:    true,
		})
		if err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}

		fmt.Println("✅ Changes committed")

		if !skipTests {
			fmt.Println("\n🧪 Running tests...")
			testExec := validation.NewTestExecutor(workDir)
			result, err := testExec.RunTests()
			
			if err != nil {
				fmt.Printf("⚠️  Tests encountered error: %v\n", err)
			} else if result.Success {
				fmt.Printf("✅ All tests passed (%d passed)\n", result.Passed)
			} else {
				fmt.Printf("❌ Tests failed (%d passed, %d failed)\n", result.Passed, result.Failed)
				fmt.Println("\nTest output:")
				fmt.Println(result.Output)
				return fmt.Errorf("tests failed")
			}
		}

		if !skipLint {
			fmt.Println("\n🔍 Running linters...")
			lintExec := validation.NewLinterExecutor(workDir)
			results, err := lintExec.RunLinters()
			
			if err != nil {
				fmt.Printf("⚠️  Linters encountered error: %v\n", err)
			} else {
				allPassed := true
				for _, result := range results {
					if result.Success && result.Issues == 0 {
						fmt.Printf("✅ %s: no issues\n", result.Tool)
					} else {
						allPassed = false
						fmt.Printf("❌ %s: %d issues (%d errors, %d warnings)\n", 
							result.Tool, result.Issues, result.Errors, result.Warnings)
					}
				}
				
				if !allPassed {
					return fmt.Errorf("linting issues found")
				}
			}
		}

		fmt.Println("\n✨ All validation passed!")
		fmt.Println("Next steps:")
		fmt.Printf("  chu issue push %d\n", issueNum)

		return nil
	},
}

var issuePushCmd = &cobra.Command{
	Use:   "push <issue-number>",
	Short: "Push branch and create pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueNum, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[0])
		}

		repo, _ := cmd.Flags().GetString("repo")
		draft, _ := cmd.Flags().GetBool("draft")

		if repo == "" {
			repo = detectGitHubRepo()
			if repo == "" {
				return fmt.Errorf("could not detect GitHub repository")
			}
		}

		workDir, _ := os.Getwd()
		client := github.NewClient(repo)
		client.SetWorkDir(workDir)

		issue, err := client.FetchIssue(issueNum)
		if err != nil {
			return fmt.Errorf("failed to fetch issue: %w", err)
		}

		branchName := issue.CreateBranchName()
		
		fmt.Printf("🚀 Pushing branch %s...\n", branchName)
		if err := client.PushBranch(branchName); err != nil {
			return fmt.Errorf("failed to push branch: %w", err)
		}

		fmt.Println("✅ Branch pushed")

		fmt.Println("\n📝 Creating pull request...")
		
		changes := []string{"Implemented fix for issue"}
		prBody := github.GeneratePRBody(issue, changes)

		pr, err := client.CreatePR(github.PRCreateOptions{
			Title:      fmt.Sprintf("Fix: %s", issue.Title),
			Body:       prBody,
			HeadBranch: branchName,
			BaseBranch: "main",
			IsDraft:    draft,
			Labels:     issue.Labels,
		})

		if err != nil {
			return fmt.Errorf("failed to create PR: %w", err)
		}

		fmt.Printf("✅ Pull request created: %s\n", pr.URL)
		fmt.Printf("   PR #%d: %s\n", pr.Number, pr.Title)
		
		return nil
	},
}

func detectGitHubRepo() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	url := strings.TrimSpace(string(output))
	
	if strings.Contains(url, "github.com") {
		parts := strings.Split(url, "github.com")
		if len(parts) < 2 {
			return ""
		}
		
		repo := strings.Trim(parts[1], ":/")
		repo = strings.TrimSuffix(repo, ".git")
		
		return repo
	}

	return ""
}

func init() {
	issueCmd.AddCommand(issueFixCmd)
	issueCmd.AddCommand(issueShowCmd)
	issueCmd.AddCommand(issueCommitCmd)
	issueCmd.AddCommand(issuePushCmd)

	issueFixCmd.Flags().String("repo", "", "GitHub repository (owner/repo)")
	issueFixCmd.Flags().Bool("draft", false, "Create draft pull request")
	issueFixCmd.Flags().Bool("skip-tests", false, "Skip running tests")
	issueFixCmd.Flags().Bool("skip-lint", false, "Skip running linters")

	issueShowCmd.Flags().String("repo", "", "GitHub repository (owner/repo)")

	issueCommitCmd.Flags().String("message", "", "Commit message")
	issueCommitCmd.Flags().Bool("skip-tests", false, "Skip running tests")
	issueCommitCmd.Flags().Bool("skip-lint", false, "Skip running linters")
	issueCommitCmd.Flags().String("repo", "", "GitHub repository (owner/repo)")

	issuePushCmd.Flags().String("repo", "", "GitHub repository (owner/repo)")
	issuePushCmd.Flags().Bool("draft", false, "Create draft pull request")
}
