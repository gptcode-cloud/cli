package modes

import (
	"chuchu/internal/config"
	"chuchu/internal/llm"
	"chuchu/internal/memory"
	"chuchu/internal/prompt"
	"context"
	"fmt"
	"strings"
)

func TDD(input string, args []string) {
	setup, _ := config.LoadSetup()
	store, _ := memory.LoadStore()

	lang := setup.Defaults.Lang

	backendName := setup.Defaults.Backend
	modelAlias := setup.Defaults.Model

	if len(args) >= 1 && args[0] != "" {
		backendName = args[0]
	}
	if len(args) >= 2 && args[1] != "" {
		modelAlias = args[1]
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
		Mode: "tdd",
		Hint: hint,
	})

	user := fmt.Sprintf(`
Language: %s

Task:
Generate tests first, then the minimum implementation to pass them.

FORMAT STRICT:
` + "```" + `tests
# file + test code
` + "```" + `

` + "```" + `impl
# file + implementation
` + "```" + `

Details:
%s
`, lang, input)

	var provider llm.Provider
	switch backendCfg.Type {
	case "openai":
		provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
	default:
		provider = llm.NewOllama(backendCfg.BaseURL)
	}

	resp, err := provider.Chat(context.Background(), llm.ChatRequest{
		SystemPrompt: sys,
		UserPrompt:   user,
		Model:        model,
	})
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Println(strings.TrimSpace(resp.Text))
}

func RunTDD(builder *prompt.Builder, provider llm.Provider, model string) error {
	input := ""
	args := []string{}
	TDD(input, args)
	return nil
}

func RunFeatureTS(builder *prompt.Builder, provider llm.Provider, model string) error {
	input := ""
	args := []string{}
	TDD(input, args)
	return nil
}
