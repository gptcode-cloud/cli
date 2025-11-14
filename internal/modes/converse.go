package modes

import (
	"bufio"
	"chuchu/internal/config"
	"chuchu/internal/llm"
	"chuchu/internal/memory"
	"chuchu/internal/prompt"
	"context"
	"fmt"
	"os"
	"strings"
)

func Converse(input string, args []string) {
	setup, _ := config.LoadSetup()
	store, _ := memory.LoadStore()

	lang := setup.Defaults.Lang
	if len(args) >= 1 && args[0] != "" {
		lang = args[0]
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

	pb := prompt.NewDefaultBuilder(store)
	hint := input
	if len(hint) > 200 {
		hint = hint[:200]
	}
	sys := pb.BuildSystemPrompt(prompt.BuildOptions{
		Lang: lang,
		Mode: "converse",
		Hint: hint,
	})

	sys += `

## Conversational Flow

When the user asks you to create code, implement features, or solve technical problems:

1. **Clarify first** - Ask specific questions about:
   - Environment and deployment preferences
   - Data formats and persistence
   - Frequency/scheduling if applicable
   - Performance/cost constraints
   - Preferred languages/tools

2. **After clarification**, propose a plan:
   - List files/modules to create or modify
   - Describe the approach
   - Ask for confirmation

3. **Generate code ONLY after confirmation**:
   - Use ` + "```tests" + ` and ` + "```impl" + ` blocks for each module
   - Each test/impl pair becomes a separate tab in the editor

**IMPORTANT**: Do NOT generate code blocks immediately. Always clarify and plan first.
`

	var provider llm.Provider
	switch backendCfg.Type {
	case "openai":
		provider = llm.NewOpenAI(backendCfg.BaseURL, backendName)
	default:
		provider = llm.NewOllama(backendCfg.BaseURL)
	}

	conversation := []string{input}
	reader := bufio.NewReader(os.Stdin)

	for {
		userPrompt := strings.Join(conversation, "\n\n---\n\n")

		resp, err := provider.Chat(context.Background(), llm.ChatRequest{
			SystemPrompt: sys,
			UserPrompt:   userPrompt,
			Model:        model,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			return
		}

		output := strings.TrimSpace(resp.Text)
		fmt.Println(output)

		if strings.Contains(output, "```tests") && strings.Contains(output, "```impl") {
			break
		}

		fmt.Fprint(os.Stderr, "\n> ")
		nextInput, _ := reader.ReadString('\n')
		nextInput = strings.TrimSpace(nextInput)

		if nextInput == "" || nextInput == "exit" || nextInput == "quit" {
			break
		}

		conversation = append(conversation, "Assistant: "+output)
		conversation = append(conversation, "User: "+nextInput)
	}
}
