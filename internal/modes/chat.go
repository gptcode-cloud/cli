package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"chuchu/internal/config"
	"chuchu/internal/llm"
	"chuchu/internal/memory"
	"chuchu/internal/prompt"
	"chuchu/internal/tools"
)

type ChatHistory struct {
	Messages []llm.ChatMessage `json:"messages"`
}

func Chat(input string, args []string) {
	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[CHAT] Starting Chat function\n")
	}
	setup, _ := config.LoadSetup()
	store, _ := memory.LoadStore()

	var history ChatHistory
	if input != "" {
		err := json.Unmarshal([]byte(input), &history)
		if err != nil {
			if os.Getenv("CHUCHU_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "JSON parse error: %v\n", err)
				fmt.Fprintf(os.Stderr, "Input: %s\n", input[:min(100, len(input))])
			}
			history.Messages = []llm.ChatMessage{{Role: "user", Content: input}}
		} else {
			if os.Getenv("CHUCHU_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "JSON parsed OK: %d messages\n", len(history.Messages))
			}
		}
	}

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
	hint := ""
	if len(history.Messages) > 0 {
		hint = history.Messages[len(history.Messages)-1].Content
	}
	if len(hint) > 200 {
		hint = hint[:200]
	}
	sys := pb.BuildSystemPrompt(prompt.BuildOptions{
		Lang: lang,
		Mode: "chat",
		Hint: hint,
	})

	sys += "\n\n## CRITICAL: Tool Calling Protocol\n\n" +
		"IMPORTANT: Use JSON tool calling format, NOT XML tags.\n" +
		"NEVER write <function=...> or <parameter=...> tags in your response.\n" +
		"Use the tool_calls field in the API response instead.\n\n" +
		"## Tool Calling Rules\n\n" +
		"1. **Only call a tool if the task CANNOT be answered without it**\n" +
		"2. **Call each tool ONLY ONCE per task** - do not repeat tool calls\n" +
		"3. **After getting tool results, provide a final answer WITHOUT calling more tools**\n" +
		"4. **If you have enough information, respond directly** - no tools needed\n\n" +
		"## You are an EXECUTOR, not a describer\n\n" +
		"When the user asks you to DO something, you MUST:\n" +
		"1. Call the tool ONCE to get the information\n" +
		"2. Use the tool result to formulate your answer\n" +
		"3. Give the final answer WITHOUT calling the tool again\n\n" +
		"DO NOT:\n" +
		"- Call the same tool multiple times with similar arguments\n" +
		"- Call tools after you already have the information\n" +
		"- Say what you're going to do before doing it\n" +
		"- Write XML tags like <function> or <parameter>\n\n" +
		"STOPPING RULE:\n" +
		"- If you just received tool results, formulate final answer and STOP\n" +
		"- Do NOT call more tools unless user asks a NEW question\n\n" +
		"## Available Tools\n\n" +
		"- **run_command**: Execute ANY shell command\n" +
		"- **list_files**: List files (use pattern: '*.go', '*.ex')\n" +
		"- **read_file**: Read file contents\n" +
		"- **write_file**: Write/update file contents\n" +
		"- **search_code**: Search for patterns\n" +
		"- **read_guideline**: Access guidelines (tdd, naming, languages)\n\n" +
		"## Examples of CORRECT behavior\n\n" +
		"❌ WRONG: 'I will run the tests for you using go test'\n" +
		"✅ RIGHT: [silently calls run_command] 'Tests passed. No test files found in any package.'\n\n" +
		"❌ WRONG: 'Let me list the Go files in internal/tools'\n" +
		"✅ RIGHT: [calls list_files] 'Found: tools.go'\n\n" +
		"❌ WRONG: '<function=list_files><parameter=path></parameter></function>'\n" +
		"✅ RIGHT: Use JSON tool_calls in API response, not XML\n\n" +
		"## When Making Code Changes\n\n" +
		"If the user asks you to CHANGE code (edit, remove, refactor, etc.):\n" +
		"1. Use read_file to read the current content\n" +
		"2. Modify the content in memory (remove lines, change code, etc.)\n" +
		"3. Use write_file to save the ENTIRE MODIFIED file\n" +
		"4. Repeat for each file that needs changes\n" +
		"5. Be concise - files will open in editor for review\n\n" +
		"CRITICAL RULES:\n" +
		"- write_file 'content' parameter MUST contain the FULL file content as text\n" +
		"- NEVER use placeholders like [read_file path=...] in content\n" +
		"- NEVER use markers like [content here] or [modified content]\n" +
		"- The content parameter receives ACTUAL file text, not references\n" +
		"- If you don't have the full content, call read_file FIRST\n\n" +
		"Example (removing comments):\n" +
		"User: 'Remove comment from line 5 in main.go'\n" +
		"YOU:\n" +
		"  1. Call read_file with path='main.go'\n" +
		"  2. Receive full file content (100 lines)\n" +
		"  3. Remove comment from line 5 in memory\n" +
		"  4. Call write_file with path='main.go' and content='<all 100 lines with modification>'\n" +
		"  5. Say: 'Done'\n\n" +
		"❌ WRONG: write_file({'path': 'main.go', 'content': '[read_file path=main.go]'})\n" +
		"❌ WRONG: write_file({'path': 'main.go', 'content': '[modified content]'})\n" +
		"❌ WRONG: Just read files and describe what to change\n" +
		"❌ WRONG: Show code in ```impl blocks\n" +
		"✅ RIGHT: read_file → get actual text → modify text → write_file with actual text\n\n" +
		"**Remember**: write_file needs REAL file content, not symbolic references.\n"

	var provider llm.Provider
	if strings.Contains(model, "compound") {
		var customExec llm.Provider
		customModel := backendCfg.DefaultModel
		
		if backendCfg.Type == "ollama" {
			customExec = llm.NewOllama(backendCfg.BaseURL)
		} else {
			customExec = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
		}
		
		provider = llm.NewOrchestrator(backendCfg.BaseURL, backendName, customExec, customModel)
	} else {
		if backendCfg.Type == "ollama" {
			provider = llm.NewOllama(backendCfg.BaseURL)
		} else {
			provider = llm.NewChatCompletion(backendCfg.BaseURL, backendName)
		}
	}

	cwd, _ := os.Getwd()
	toolsRaw := tools.GetAvailableTools()
	var availableTools []interface{}
	for _, t := range toolsRaw {
		availableTools = append(availableTools, t)
	}
	
	if os.Getenv("CHUCHU_DEBUG") == "1" {
		toolsJSON, _ := json.MarshalIndent(availableTools, "", "  ")
		fmt.Fprintf(os.Stderr, "\n=== TOOLS DEFINITION ===\n%s\n\n", string(toolsJSON))
	}
	
	messages := history.Messages
	
	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "\n### MESSAGES RECEIVED: %d\n", len(messages))
		for i, msg := range messages {
			fmt.Fprintf(os.Stderr, "  [%d] role=%s content=%s...\n", i, msg.Role, msg.Content[:min(50, len(msg.Content))])
		}
	}
	
	if len(messages) == 0 || messages[len(messages)-1].Role != "user" {
		fmt.Fprintln(os.Stderr, "\nERROR: Invalid message history")
		if len(messages) > 0 {
			fmt.Fprintf(os.Stderr, "  Last message role: %s\n", messages[len(messages)-1].Role)
		}
		fmt.Println("Erro: Invalid message history - must have at least one user message")
		return
	}
	
	maxIterations := 10
	
	for i := 0; i < maxIterations; i++ {
		req := llm.ChatRequest{
			SystemPrompt: sys,
			Model:        model,
			Tools:        availableTools,
			Messages:     messages,
		}
		
		if os.Getenv("CHUCHU_DEBUG") == "1" && i == 0 {
			fmt.Fprintf(os.Stderr, "\n### TOOLS SENT: %d tools\n", len(availableTools))
			fmt.Fprintf(os.Stderr, "### MESSAGES: %d messages\n\n", len(messages))
		}
		
		resp, err := provider.Chat(context.Background(), req)
		if err != nil {
			fmt.Println("Erro:", err)
			return
		}
		
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "### RESPONSE: text_len=%d, tool_calls=%d\n", len(resp.Text), len(resp.ToolCalls))
		}
		
		if resp.Text != "" {
			fmt.Println(strings.TrimSpace(resp.Text))
		}
		
	if len(resp.ToolCalls) == 0 {
			if os.Getenv("CHUCHU_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "### NO TOOL CALLS - Breaking loop\n")
			}
			break
		}
			
		messages = append(messages, llm.ChatMessage{
			Role:      "assistant",
			Content:   resp.Text,
			ToolCalls: resp.ToolCalls,
		})
		
		for _, tc := range resp.ToolCalls {
			event := map[string]interface{}{
				"type": "tool_start",
				"tool": tc.Name,
				"args": tc.Arguments,
			}
			eventJSON, _ := json.Marshal(event)
			fmt.Printf("__EVENT__%s__EVENT__\n", string(eventJSON))
			
			var args map[string]interface{}
			json.Unmarshal([]byte(tc.Arguments), &args)
			
			toolCall := tools.ToolCall{
				Name:      tc.Name,
				Arguments: args,
			}
			
			result := tools.ExecuteTool(toolCall, cwd)
			
			if result.Error != "" {
				eventEnd := map[string]interface{}{
					"type":   "tool_end",
					"tool":   tc.Name,
					"error":  result.Error,
				}
				eventJSON, _ := json.Marshal(eventEnd)
				fmt.Printf("__EVENT__%s__EVENT__\n", string(eventJSON))
				
				messages = append(messages, llm.ChatMessage{
					Role:    "tool",
					Content: fmt.Sprintf("Error: %s", result.Error),
					Name:    tc.Name,
					ToolCallID: tc.ID,
				})
			} else {
				eventEnd := map[string]interface{}{
					"type":   "tool_end",
					"tool":   tc.Name,
					"result": result.Result,
				}
				
				if tc.Name == "write_file" {
					if path, ok := args["path"].(string); ok {
						eventEnd["path"] = path
					}
				}
				
				eventJSON, _ := json.Marshal(eventEnd)
				fmt.Printf("__EVENT__%s__EVENT__\n", string(eventJSON))
				
				content := result.Result
				if content == "" {
					content = "Success"
				}
				
				messages = append(messages, llm.ChatMessage{
					Role:    "tool",
					Content: content,
					Name:    tc.Name,
					ToolCallID: tc.ID,
				})
			}
		}
	}
}

func RunChat(builder *prompt.Builder, provider llm.Provider, model string, cliArgs []string) error {
	input, _ := io.ReadAll(os.Stdin)
	Chat(string(input), cliArgs)
	return nil
}
