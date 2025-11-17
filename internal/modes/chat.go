package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"chuchu/internal/agents"
	"chuchu/internal/config"
	"chuchu/internal/llm"
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
	modelAlias := setup.Defaults.Model

	if len(args) >= 2 && args[1] != "" {
		backendName = args[1]
	}
	if len(args) >= 3 && args[2] != "" {
		modelAlias = args[2]
	}

	backendCfg := setup.Backend[backendName]
	model := backendCfg.DefaultModel
	if alias, ok := backendCfg.Models[modelAlias]; ok {
		model = alias
	} else if modelAlias != "" {
		model = modelAlias
	}

	cwd, _ := os.Getwd()

	var provider llm.Provider
	var orchestrator *llm.OrchestratorProvider

	if strings.Contains(model, "compound") {
		var customExec llm.Provider
		customModel := backendCfg.DefaultModel

		if backendCfg.Type == "ollama" {
			customExec = llm.NewOllama(backendCfg.BaseURL)
		} else {
			customExec = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
		}

		orchestrator = llm.NewOrchestrator(backendCfg.BaseURL, backendName, customExec, customModel)
		provider = customExec
	} else {
		if backendCfg.Type == "ollama" {
			provider = llm.NewOllama(backendCfg.BaseURL)
		} else {
			provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
		}
	}

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

	coordinator := agents.NewCoordinator(provider, orchestrator, cwd)
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

	fmt.Println(result)
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
