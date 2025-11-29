package repl

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"chuchu/internal/config"
	"chuchu/internal/llm"
	"chuchu/internal/modes"
	"chuchu/internal/prompt"
	"github.com/chzyer/readline"
)

// ChatREPL implements a Read-Eval-Print Loop for chat conversations
type ChatREPL struct {
	rl      *readline.Instance
	ctxMgr  *ContextManager
	builder *prompt.Builder
	model   string
}

// NewChatREPL creates a new chat REPL instance
func NewChatREPL(maxTokens, maxMessages int) (*ChatREPL, error) {
	// Initialize readline
	rl, err := readline.New("> ")
	if err != nil {
		return nil, fmt.Errorf("failed to create readline instance: %w", err)
	}

	// Set up history file
	histFile, err := os.UserHomeDir()
	if err == nil {
		histFile = histFile + "/.chuchu_history"
		rl, err = readline.NewEx(&readline.Config{
			Prompt:              "> ",
			HistoryFile:         histFile,
			AutoComplete:        nil,
			InterruptPrompt:     "^C",
			HistorySearchFold:   true,
			FuncFilterInputRune: filterInput,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create readline instance with history: %w", err)
		}
	}

	// Load Chuchu configuration
	setup, err := config.LoadSetup()
	var model string
	if err != nil {
		// Use default values if we can't load configuration
		model = "gpt-4"
	} else {
		// Get backend and model info
		backendName := setup.Defaults.Backend
		modelAlias := setup.Defaults.Model
		backendCfg := setup.Backend[backendName]

		model = backendCfg.DefaultModel
		if alias, ok := backendCfg.Models[modelAlias]; ok {
			model = alias
		} else if modelAlias != "" {
			model = modelAlias
		}
	}

	// Initialize builder
	builder := prompt.NewDefaultBuilder(nil)

	ctxMgr := NewContextManager(maxTokens, maxMessages)

	return &ChatREPL{
		rl:      rl,
		ctxMgr:  ctxMgr,
		builder: builder,
		model:   model,
	}, nil
}

// filterInput prevents REPL from treating certain runes as special
func filterInput(r rune) (rune, bool) {
	switch r {
	case readline.CharCtrlZ, readline.CharCtrlL:
		return r, false
	}
	return r, true
}

// RunWithInitialMessage starts REPL with an optional initial message to process first
func (r *ChatREPL) RunWithInitialMessage(initialMessage string) error {
	if initialMessage != "" {
		// Process initial message first
		if err := r.processMessage(initialMessage); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Println("")

		// If non-interactive (piped input), exit after processing
		if !isInteractiveTTY() {
			return nil
		}
	}
	return r.Run()
}

// Run starts the REPL loop
func (r *ChatREPL) Run() error {
	if !isInteractiveTTY() {
		// Non-interactive, don't show prompts
		return nil
	}
	fmt.Println("Chuchu Chat REPL - Type /help for commands")
	fmt.Println("")

	// Update file context initially
	if err := r.ctxMgr.UpdateFileContext(); err != nil {
		fmt.Printf("Warning: Failed to load file context: %v\n", err)
	}

	for {
		// Read input
		line, err := r.rl.Readline()
		if err != nil { // readline errors, including io.EOF
			if err == readline.ErrInterrupt {
				fmt.Println("\nUse /exit or Ctrl+D to quit")
				continue
			} else if err == io.EOF {
				fmt.Println("\nGoodbye!")
				break
			}
			fmt.Fprintf(os.Stderr, "REPL error: %v\n", err)
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Trim whitespace
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle special commands starting with '/'
		if strings.HasPrefix(line, "/") {
			shouldContinue, shouldExit := r.handleCommand(line)
			if shouldExit {
				break
			}
			if shouldContinue {
				continue
			}
		}

		// Process normal message
		if err := r.processMessage(line); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
}

// handleCommand processes REPL commands
// Returns (shouldContinue, shouldExit)
func (r *ChatREPL) handleCommand(cmd string) (bool, bool) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false, false
	}

	switch parts[0] {
	case "/exit", "/quit":
		fmt.Println("Goodbye!")
		return false, true

	case "/clear":
		fmt.Println("Conversation history cleared.")
		r.ctxMgr.Clear()
		return true, false

	case "/help":
		r.showHelp()
		return true, false

	case "/save":
		if len(parts) < 2 {
			fmt.Println("Usage: /save <filename>")
			return true, false
		}
		if err := r.ctxMgr.SaveConversation(parts[1]); err != nil {
			fmt.Printf("Failed to save: %v\n", err)
		} else {
			fmt.Printf("Conversation saved to %s\n", parts[1])
		}
		return true, false

	case "/load":
		if len(parts) < 2 {
			fmt.Println("Usage: /load <filename>")
			return true, false
		}
		if err := r.ctxMgr.LoadConversation(parts[1]); err != nil {
			fmt.Printf("Failed to load: %v\n", err)
		} else {
			fmt.Printf("Conversation loaded from %s\n", parts[1])
		}
		return true, false

	case "/context":
		fmt.Println(r.ctxMgr.GetStatus())
		return true, false

	case "/files":
		fmt.Println("Files in context:")
		files := r.ctxMgr.GetFileContext()
		if files == "" {
			fmt.Println("  No files in context")
		} else {
			fmt.Println(files)
		}
		return true, false

	case "/history":
		r.showHistory()
		return true, false

	default:
		fmt.Printf("Unknown command: %s (type /help for available commands)\n", parts[0])
		return true, false
	}
}

// processMessage handles regular chat messages
func (r *ChatREPL) processMessage(input string) error {
	// Add user message to context
	inputTokens := estimateTokens(input)
	r.ctxMgr.AddMessage("user", input, inputTokens)

	// Combine conversation history with file context
	conversationContext := r.ctxMgr.GetContext()
	fileContext := r.ctxMgr.GetFileContext()

	// Prepare prompt with context
	fullPrompt := ""
	if fileContext != "" {
		fullPrompt += "Relevant files:\n" + fileContext + "\n"
	}
	if conversationContext != "" {
		fullPrompt += "Previous conversation:\n" + conversationContext + "\n"
	}

	// Process the message using existing chat mode functionality
	// Note: We need to extract the model and builder config from the setup
	fmt.Printf("Processing: %s\n", input)

	// For now, use the existing modes.Chat with the combined prompt
	modes.Chat(fullPrompt+"\nUser: "+input, []string{})

	return nil
}

// showHelp displays available REPL commands
func (r *ChatREPL) showHelp() {
	fmt.Println("REPL Commands:")
	fmt.Println("  /exit, /quit   - Exit the chat")
	fmt.Println("  /clear         - Clear conversation history")
	fmt.Println("  /save <file>   - Save conversation to file")
	fmt.Println("  /load <file>   - Load conversation from file")
	fmt.Println("  /context       - Show context statistics")
	fmt.Println("  /files         - List files in context")
	fmt.Println("  /history       - Show conversation history")
	fmt.Println("  /help          - Show this help")
	fmt.Println("")
	fmt.Println("All other input will be processed as a chat message.")
}

// showHistory displays the conversation history
func (r *ChatREPL) showHistory() {
	messages := r.ctxMgr.GetRecentMessages(5)
	fmt.Printf("Last %d messages:\n", len(messages))
	for _, msg := range messages {
		role := msg.Role
		switch role {
		case "user":
			role = "User"
		case "assistant":
			role = "Assistant"
		default:
			if len(role) > 0 {
				role = strings.ToUpper(role[:1]) + role[1:]
			}
		}
		fmt.Printf("[%s] %s: %s\n",
			msg.Timestamp.Format("15:04"),
			role,
			msg.Content)
	}
}

func estimateTokens(text string) int {
	return len(text) / 4
}

func isInteractiveTTY() bool {
	cmd := exec.Command("tty", "-s")
	return cmd.Run() == nil
}

func isOpsQuery(s string) bool {
	q := strings.ToLower(s)
	keys := []string{
		"system data",
		"disk usage",
		"storage",
		"disk space",
		"out of space",
		"investigate",
		"troubleshoot",
		"diagnose",
		"macos",
		"apfs",
		"snapshot",
		"time machine",
		"cache",
		"xcode",
		"docker",
		"high usage",
		"dados do sistema",
		"armazenamento",
		"disco",
	}
	for _, k := range keys {
		if strings.Contains(q, k) {
			return true
		}
	}
	return false
}

func RunSingleShot(input string, args []string) error {
	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[RunSingleShot] Input: %s\n", input)
		fmt.Fprintf(os.Stderr, "[RunSingleShot] isOpsQuery: %v\n", isOpsQuery(input))
	}
	if isOpsQuery(input) {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[RunSingleShot] Routing to run mode\n")
		}
		setup, _ := config.LoadSetup()
		backendName := setup.Defaults.Backend
		backendCfg := setup.Backend[backendName]

		// Use query agent model from profile
		queryModel := backendCfg.GetModelForAgent("query")

		var provider llm.Provider
		if backendCfg.Type == "ollama" {
			provider = llm.NewOllama(backendCfg.BaseURL)
		} else {
			provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
		}
		builder := prompt.NewDefaultBuilder(nil)
		return modes.RunExecute(builder, provider, queryModel, []string{input})
	}
	modes.Chat(input, args)
	return nil
}
