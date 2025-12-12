package modes

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gptcode/internal/agents"
	"gptcode/internal/config"
	"gptcode/internal/llm"
	"gptcode/internal/output"

	"golang.org/x/term"
)

func RunResearch(args []string) error {
	question := ""
	if len(args) > 0 {
		question = strings.Join(args, " ")
	}

	if question == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Fprintln(os.Stderr, "Research mode - Analyze codebase and external docs")
		fmt.Fprintln(os.Stderr, "\nWhat would you like to research?")
		fmt.Fprint(os.Stderr, "> ")
		if scanner.Scan() {
			question = scanner.Text()
		}
	}

	if question == "" {
		return fmt.Errorf("no research question provided")
	}

	setup, _ := config.LoadSetup()
	backendName := setup.Defaults.Backend
	backendCfg := setup.Backend[backendName]
	cwd, _ := os.Getwd()

	urls := extractURLs(question)
	var externalDocs string

	if len(urls) > 0 {
		fmt.Fprintf(os.Stderr, "⠋ Fetching external documentation...\n")

		var orchestrator *llm.OrchestratorProvider
		if backendCfg.Type == "ollama" {
			customExec := llm.NewOllama(backendCfg.BaseURL)
			orchestrator = llm.NewOrchestrator(backendCfg.BaseURL, backendName, customExec, backendCfg.DefaultModel)
		} else {
			customExec := llm.NewChatCompletion(backendCfg.BaseURL, backendName)
			orchestrator = llm.NewOrchestrator(backendCfg.BaseURL, backendName, customExec, backendCfg.DefaultModel)
		}

		researchAgent := agents.NewResearch(orchestrator)
		for _, url := range urls {
			docPrompt := fmt.Sprintf("Visit %s and summarize the key implementation details for: %s", url, question)
			docResult, err := researchAgent.Execute(context.Background(), []llm.ChatMessage{{Role: "user", Content: docPrompt}}, nil)
			if err == nil {
				externalDocs += fmt.Sprintf("\n\n## Documentation from %s\n\n%s", url, docResult)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "⠋ Analyzing codebase...\n")

	var customExec llm.Provider
	if backendCfg.Type == "ollama" {
		customExec = llm.NewOllama(backendCfg.BaseURL)
	} else {
		customExec = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	}

	queryModel := backendCfg.GetModelForAgent("query")
	queryAgent := agents.NewQuery(customExec, cwd, queryModel)

	codebasePrompt := fmt.Sprintf(`Brief codebase overview:

1. List root directory files (use list_files on ".")
2. Identify main language/framework
3. Suggest 2-3 key directories for: %s

Keep response under 150 words.`, question)

	codebaseAnalysis, err := queryAgent.Execute(context.Background(), []llm.ChatMessage{{Role: "user", Content: codebasePrompt}}, nil)
	if err != nil {
		return fmt.Errorf("codebase analysis failed: %w", err)
	}

	home, _ := os.UserHomeDir()
	researchDir := filepath.Join(home, ".gptcode", "research")
	_ = os.MkdirAll(researchDir, 0755)

	fullResearch := fmt.Sprintf(`# Research: %s

## Summary

%s

%s

## Generated
%s`, question, codebaseAnalysis, externalDocs, time.Now().Format("2006-01-02 15:04:05"))

	if term.IsTerminal(int(os.Stdout.Fd())) {
		rendered, err := output.RenderMarkdown(fullResearch)
		if err != nil {
			rendered = fullResearch
		}
		fmt.Println(output.Separator())
		fmt.Print(rendered)
		fmt.Println(output.Separator())
	} else {
		fmt.Println(fullResearch)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	sanitizedQuestion := sanitizeFilename(question)
	filename := fmt.Sprintf("%s_%s.md", timestamp, sanitizedQuestion)
	researchPath := filepath.Join(researchDir, filename)

	err = os.WriteFile(researchPath, []byte(fullResearch), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Could not save research to %s: %v\n", researchPath, err)
	} else {
		fmt.Fprintf(os.Stderr, "\n✓ Research saved to: %s\n", researchPath)
	}

	return nil
}

func extractURLs(text string) []string {
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	return urlRegex.FindAllString(text, -1)
}

func sanitizeFilename(question string) string {
	sanitized := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		if r == ' ' {
			return '-'
		}
		return -1
	}, question)
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}
	return sanitized
}
