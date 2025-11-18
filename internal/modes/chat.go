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
	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[CHAT] Starting Chat function\n")
	}

	setup, _ := config.LoadSetup()

	var history ChatHistory
	if input != "" {
		err := json.Unmarshal([]byte(input), &history)
		if err != nil {
			history.Messages = []llm.ChatMessage{{Role: "user", Content: input}}
		}
	}

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

	userMessage := history.Messages[len(history.Messages)-1].Content

	var stopSpinner chan bool
	if os.Getenv("CHUCHU_DEBUG") != "1" {
		stopSpinner = make(chan bool, 1)
		go showSpinner(stopSpinner)
	}

	routerModel := backendCfg.GetModelForAgent("router")
	editorModel := backendCfg.GetModelForAgent("editor")
	queryModel := backendCfg.GetModelForAgent("query")
	
	coordinator := agents.NewCoordinator(provider, orchestrator, cwd, routerModel, editorModel, queryModel, researchModel)
	result, err := coordinator.Execute(context.Background(), userMessage)

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
