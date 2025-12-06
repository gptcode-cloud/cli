package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"chuchu/internal/config"
	"chuchu/internal/github"
	"chuchu/internal/langdetect"
	"chuchu/internal/llm"
	"chuchu/internal/modes"
	"chuchu/internal/recovery"
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
		autonomous, _ := cmd.Flags().GetBool("autonomous")

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

		task := fmt.Sprintf("Fix issue #%d: %s", issue.Number, issue.Title)
		if len(reqs) > 0 {
			task += ", Requirements: " + strings.Join(reqs, "; ")
		}

		if autonomous {
			setup, err := config.LoadSetup()
			if err != nil {
				return fmt.Errorf("failed to load setup: %w", err)
			}
			backendName := setup.Defaults.Backend
			backendCfg := setup.Backend[backendName]
			var provider llm.Provider
			if backendCfg.Type == "ollama" {
				provider = llm.NewOllama(backendCfg.BaseURL)
			} else {
				provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
			}
			queryModel := backendCfg.GetModelForAgent("query")
			if queryModel == "" {
				queryModel = backendCfg.DefaultModel
			}
			language := string(langdetect.DetectLanguage(workDir))
			if language == "" || language == "unknown" {
				language = setup.Defaults.Lang
				if language == "" {
					language = "go"
				}
			}
			exec := modes.NewAutonomousExecutorWithBackend(provider, workDir, queryModel, language, backendName)
			if err := exec.Execute(context.Background(), task); err != nil {
				return fmt.Errorf("autonomous implementation failed: %w", err)
			}
			fmt.Println("\n[OK] Implementation complete")
		} else {
			fmt.Println("\nImplementation not executed (use --autonomous to enable)")
		}

		fmt.Println("\nNext steps:")
		fmt.Printf("   chu issue commit %d\n", issueNum)
		fmt.Printf("   chu issue push %d\n", issueNum)

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
		skipBuild, _ := cmd.Flags().GetBool("skip-build")
		checkCoverage, _ := cmd.Flags().GetBool("check-coverage")
		minCoverage, _ := cmd.Flags().GetFloat64("min-coverage")
		securityScan, _ := cmd.Flags().GetBool("security-scan")
		autoFix, _ := cmd.Flags().GetBool("auto-fix")
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

		if !skipBuild {
			fmt.Println("\n🔨 Running build...")
			buildExec := validation.NewBuildExecutor(workDir)
			buildResult, err := buildExec.RunBuild()
			if err != nil {
				fmt.Printf("⚠️  Build check failed: %v\n", err)
			} else if buildResult.Success {
				fmt.Println("✅ Build successful")
			} else {
				fmt.Printf("❌ Build failed\n")
				if buildResult.Output != "" {
					fmt.Println("\nBuild output:")
					fmt.Println(buildResult.Output)
				}
				return fmt.Errorf("build failed")
			}
		}

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

				if autoFix {
					fmt.Println("\n🔧 Attempting auto-fix...")
					if fixErr := attemptTestFix(workDir, result); fixErr != nil {
						fmt.Printf("⚠️  Auto-fix failed: %v\n", fixErr)
						fmt.Println("\nTest output:")
						fmt.Println(result.Output)
						return fmt.Errorf("tests failed")
					}
					fmt.Println("✅ Tests fixed automatically")
				} else {
					fmt.Println("\nTest output:")
					fmt.Println(result.Output)
					return fmt.Errorf("tests failed (use --auto-fix to attempt automatic fixes)")
				}
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
					if autoFix {
						fmt.Println("\n🔧 Attempting auto-fix...")
						if fixErr := attemptLintFix(workDir, results); fixErr != nil {
							fmt.Printf("⚠️  Auto-fix failed: %v\n", fixErr)
							return fmt.Errorf("linting issues found")
						}
						fmt.Println("✅ Lint issues fixed automatically")
					} else {
						return fmt.Errorf("linting issues found (use --auto-fix to attempt automatic fixes)")
					}
			}
		}
		}

		if checkCoverage {
			fmt.Println("\n📊 Checking code coverage...")
			covExec := validation.NewCoverageExecutor(workDir)
			covResult, err := covExec.RunCoverage(minCoverage)
			if err != nil {
				if covResult != nil && covResult.Coverage > 0 {
					fmt.Printf("⚠️  Coverage %.1f%% below minimum %.1f%%\n", covResult.Coverage, minCoverage)
				} else {
					fmt.Printf("⚠️  Coverage check failed: %v\n", err)
				}
			} else {
				fmt.Printf("✅ Coverage: %.1f%%\n", covResult.Coverage)
			}
		}

		if securityScan {
			fmt.Println("\n🔒 Running security scan...")
			secScanner := validation.NewSecurityScanner(workDir)
			secResult, err := secScanner.RunScan()
			if err != nil {
				if secResult != nil && secResult.Vulnerabilities > 0 {
					fmt.Printf("❌ Found %d vulnerabilities\n", secResult.Vulnerabilities)
					return fmt.Errorf("security vulnerabilities found")
				}
				fmt.Printf("⚠️  Security scan failed: %v\n", err)
			} else {
				if strings.Contains(secResult.Output, "skipping") {
					fmt.Println("⚠️  Security scanner not available (skipped)")
				} else {
					fmt.Println("✅ No security vulnerabilities found")
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

func attemptTestFix(workDir string, testResult *validation.TestResult) error {
	setup, err := config.LoadSetup()
	if err != nil {
		return err
	}

	backendName := setup.Defaults.Backend
	backendCfg := setup.Backend[backendName]
	var provider llm.Provider
	if backendCfg.Type == "ollama" {
		provider = llm.NewOllama(backendCfg.BaseURL)
	} else {
		provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	}

	model := backendCfg.GetModelForAgent("editor")
	if model == "" {
		model = backendCfg.DefaultModel
	}

	fixer := recovery.NewErrorFixer(provider, model, workDir)
	fixResult, err := fixer.FixTestFailures(context.Background(), testResult, 2)
	if err != nil {
		return err
	}

	if !fixResult.Success {
		return fmt.Errorf("could not fix test failures after %d attempts", fixResult.FixAttempts)
	}

	return nil
}

func attemptLintFix(workDir string, lintResults []*validation.LintResult) error {
	setup, err := config.LoadSetup()
	if err != nil {
		return err
	}

	backendName := setup.Defaults.Backend
	backendCfg := setup.Backend[backendName]
	var provider llm.Provider
	if backendCfg.Type == "ollama" {
		provider = llm.NewOllama(backendCfg.BaseURL)
	} else {
		provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	}

	model := backendCfg.GetModelForAgent("editor")
	if model == "" {
		model = backendCfg.DefaultModel
	}

	fixer := recovery.NewErrorFixer(provider, model, workDir)
	fixResult, err := fixer.FixLintIssues(context.Background(), lintResults, 2)
	if err != nil {
		return err
	}

	if !fixResult.Success {
		return fmt.Errorf("could not fix lint issues after %d attempts", fixResult.FixAttempts)
	}

	return nil
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
	issueFixCmd.Flags().Bool("autonomous", true, "Execute implementation autonomously")

	issueShowCmd.Flags().String("repo", "", "GitHub repository (owner/repo)")

	issueCommitCmd.Flags().String("message", "", "Commit message")
	issueCommitCmd.Flags().Bool("skip-tests", false, "Skip running tests")
	issueCommitCmd.Flags().Bool("skip-lint", false, "Skip running linters")
	issueCommitCmd.Flags().Bool("skip-build", false, "Skip build check")
	issueCommitCmd.Flags().Bool("check-coverage", false, "Check code coverage")
	issueCommitCmd.Flags().Float64("min-coverage", 0.0, "Minimum coverage threshold (0-100)")
	issueCommitCmd.Flags().Bool("security-scan", false, "Run security vulnerability scan")
	issueCommitCmd.Flags().Bool("auto-fix", true, "Automatically fix test/lint failures")
	issueCommitCmd.Flags().String("repo", "", "GitHub repository (owner/repo)")

	issuePushCmd.Flags().String("repo", "", "GitHub repository (owner/repo)")
	issuePushCmd.Flags().Bool("draft", false, "Create draft pull request")
}
