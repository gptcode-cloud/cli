package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/term"

	"chuchu/internal/agents"
	"chuchu/internal/config"
	"chuchu/internal/llm"
	"chuchu/internal/output"
	"chuchu/internal/prompt"
)

type ChatHistory struct {
	Messages []llm.ChatMessage `json:"messages"`
}

func Chat(input string, args []string) {
	os.Stdout.Sync()
	
	fmt.Fprintf(os.Stderr, "[CHAT] Starting Chat function, input len=%d\n", len(input))
	
	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[CHAT] Input: %s\n", input[:min(100, len(input))])
	}

	setup, _ := config.LoadSetup()

	var history ChatHistory
	if input != "" {
		// Try to unmarshal as JSON history
		if err := json.Unmarshal([]byte(input), &history); err != nil {
			// If input looks like JSON but failed to parse, log error and return
			// This prevents sending raw JSON as a user message which blows up context
			if len(input) > 0 && input[0] == '{' {
				fmt.Fprintf(os.Stderr, "Error parsing chat history: %v\n", err)
				return
			}
			// Otherwise treat as a single new message (CLI usage)
			history.Messages = []llm.ChatMessage{{Role: "user", Content: input}}
		}
	}

	// Truncate history to avoid context limits
	history.Messages = truncateHistory(history.Messages, 20)

	backendName := setup.Defaults.Backend

	if len(args) >= 2 && args[1] != "" {
		backendName = args[1]
	}

	backendCfg := setup.Backend[backendName]

	cwd, _ := os.Getwd()

	var provider llm.Provider
	if backendCfg.Type == "ollama" {
		provider = llm.NewOllama(backendCfg.BaseURL)
	} else {
		provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	}

	researchModel := backendCfg.GetModelForAgent("research")
	orchestrator := llm.NewOrchestrator(backendCfg.BaseURL, backendName, provider, researchModel)

	if len(history.Messages) == 0 || history.Messages[len(history.Messages)-1].Role != "user" {
		fmt.Fprintln(os.Stderr, "\nERROR: Invalid message history")
		fmt.Println("Erro: Invalid message history - must have at least one user message")
		return
	}

	lastUserMessage := history.Messages[len(history.Messages)-1].Content

	if IsComplexTask(lastUserMessage) {
		fmt.Fprintln(os.Stderr, "[CHAT] Complexity detected, using guided mode")
		fmt.Fprintf(os.Stderr, "[CHAT] Message: %s\n", lastUserMessage)
		queryModel := backendCfg.GetModelForAgent("query")
		guided := NewGuidedMode(orchestrator, cwd, queryModel)
		if err := guided.Execute(context.Background(), lastUserMessage); err != nil {
			fmt.Fprintln(os.Stderr, "Guided mode error:", err)
			fmt.Println("Erro:", err)
		}
		os.Stdout.Sync()
		
		time.Sleep(200 * time.Millisecond)
		
		io.Copy(io.Discard, os.Stdin)
		return
	}

	var stopSpinner chan bool
	if os.Getenv("CHUCHU_DEBUG") != "1" {
		stopSpinner = make(chan bool, 1)
		go showSpinner(stopSpinner)
	}

	routerModel := backendCfg.GetModelForAgent("router")
	editorModel := backendCfg.GetModelForAgent("editor")
	queryModel := backendCfg.GetModelForAgent("query")

	coordinator := agents.NewCoordinator(provider, orchestrator, cwd, routerModel, editorModel, queryModel, researchModel)

	statusCallback := func(status string) {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[STATUS] %s\n", status)
		} else {
			fmt.Fprintf(os.Stderr, "\r[STATUS] %s", status)
		}
	}

	result, err := coordinator.Execute(context.Background(), history.Messages, statusCallback)

	if os.Getenv("CHUCHU_DEBUG") != "1" {
		stopSpinner <- true
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(os.Stderr, "\r\033[K")
	}

	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	isTerminal := isInteractiveTerminal()

	if isTerminal {
		parsed := output.ParseMarkdown(result)

		rendered, err := output.RenderMarkdown(parsed.RenderedText)
		if err != nil {
			rendered = result
		}

		fmt.Println(output.Separator())
		fmt.Print(rendered)
		fmt.Println(output.Separator())

		if len(parsed.CodeBlocks) > 0 {
			for _, block := range parsed.CodeBlocks {
				action := output.PromptCodeBlock(block, len(parsed.CodeBlocks))
				output.HandleCodeBlock(action, block.Code)
			}
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, output.Success("All commands processed."))
			fmt.Fprintln(os.Stderr, "")
			fmt.Println(output.Separator())
		}
	} else {
		fmt.Println(result)
	}
}

func isInteractiveTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func showSpinner(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			return
		default:
			fmt.Fprintf(os.Stderr, "\r%s Thinking...", spinner[i%len(spinner)])
			os.Stderr.Sync()
			i++
			time.Sleep(80 * time.Millisecond)
		}
	}
}

func RunChat(builder *prompt.Builder, provider llm.Provider, model string, cliArgs []string) error {
	input, _ := io.ReadAll(os.Stdin)
	Chat(string(input), cliArgs)
	return nil
}

func truncateHistory(messages []llm.ChatMessage, maxMessages int) []llm.ChatMessage {
	if len(messages) <= maxMessages {
		return messages
	}

	// Always keep the system prompt if it exists (though usually it's added later)
	// For now, just keep the last N messages
	return messages[len(messages)-maxMessages:]
}
