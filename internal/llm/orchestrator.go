package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"gptcode/internal/tools"
)

type OrchestratorProvider struct {
	compound       *ChatCompletionProvider
	customExecutor Provider
	customModel    string
}

func NewOrchestrator(baseURL, backendName string, customExecutor Provider, customModel string) *OrchestratorProvider {
	return &OrchestratorProvider{
		compound:       NewChatCompletion(baseURL, backendName),
		customExecutor: customExecutor,
		customModel:    customModel,
	}
}

var compoundBuiltInTools = map[string]bool{
	"web_search":         true,
	"code_interpreter":   true,
	"visit_website":      true,
	"browser_automation": true,
	"wolfram_alpha":      true,
}

func (o *OrchestratorProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if os.Getenv("GPTCODE_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "\n[ORCHESTRATOR] Starting autonomous execution loop\n")
	}

	maxIterations := 10
	toolCallHistory := make(map[string]int)
	conversation := append([]ChatMessage{}, req.Messages...)

	if req.UserPrompt != "" {
		conversation = append(conversation, ChatMessage{
			Role:    "user",
			Content: req.UserPrompt,
		})
	}

	for iteration := 0; iteration < maxIterations; iteration++ {
		if os.Getenv("GPTCODE_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Iteration %d/%d\n", iteration+1, maxIterations)
		}

		executorReq := ChatRequest{
			SystemPrompt: req.SystemPrompt,
			Model:        o.customModel,
			Messages:     conversation,
			Tools:        req.Tools,
		}

		resp, err := o.customExecutor.Chat(ctx, executorReq)
		if err != nil {
			return nil, err
		}

		if len(resp.ToolCalls) == 0 && resp.Text != "" {
			parsedCalls := ParseToolCallsFromText(resp.Text)
			if len(parsedCalls) > 0 {
				if os.Getenv("GPTCODE_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Parsed %d tool calls from text\n", len(parsedCalls))
				}
				resp.ToolCalls = parsedCalls
			}
		}

		if len(resp.ToolCalls) == 0 {
			if os.Getenv("GPTCODE_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] No more tools to call, returning final response\n")
			}
			return &ChatResponse{
				Text: resp.Text,
			}, nil
		}

		conversation = append(conversation, ChatMessage{
			Role:      "assistant",
			Content:   resp.Text,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			toolKey := fmt.Sprintf("%s:%s", tc.Name, tc.Arguments)
			toolCallHistory[toolKey]++

			if toolCallHistory[toolKey] > 1 {
				if os.Getenv("GPTCODE_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Tool %s called %d times with same args - forcing stop\n", tc.Name, toolCallHistory[toolKey])
				}

				for i := len(conversation) - 1; i >= 0; i-- {
					if conversation[i].Role == "tool" && conversation[i].Name == tc.Name {
						return &ChatResponse{
							Text: conversation[i].Content,
						}, nil
					}
				}

				return &ChatResponse{
					Text: "Task completed.",
				}, nil
			}

			var toolResult string

			if compoundBuiltInTools[tc.Name] {
				if os.Getenv("GPTCODE_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Executing Compound tool: %s\n", tc.Name)
				}

				var argsMap map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Arguments), &argsMap); err == nil {
					prompt := tc.Arguments
					if query, ok := argsMap["query"].(string); ok {
						prompt = query
					} else if q, ok := argsMap["q"].(string); ok {
						prompt = q
					}

					compoundResp, compoundErr := o.compound.Chat(ctx, ChatRequest{
						Model:        "groq/compound",
						SystemPrompt: req.SystemPrompt,
						UserPrompt:   prompt,
					})

					if compoundErr != nil {
						toolResult = fmt.Sprintf("Error: %v", compoundErr)
					} else {
						toolResult = compoundResp.Text
					}
				} else {
					toolResult = "Error: invalid arguments"
				}
			} else {
				if os.Getenv("GPTCODE_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Executing custom tool: %s\n", tc.Name)
				}

				var argsMap map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Arguments), &argsMap); err != nil {
					toolResult = fmt.Sprintf("Error parsing arguments: %v", err)
				} else {
					cwd, _ := os.Getwd()
					toolCall := tools.ToolCall{
						Name:      tc.Name,
						Arguments: argsMap,
					}
					result := tools.ExecuteTool(toolCall, cwd)
					if result.Error != "" {
						toolResult = fmt.Sprintf("Error: %s", result.Error)
					} else {
						toolResult = result.Result
						if toolResult == "" {
							toolResult = "Success"
						}
					}
				}
			}

			conversation = append(conversation, ChatMessage{
				Role:       "tool",
				Content:    toolResult,
				Name:       tc.Name,
				ToolCallID: tc.ID,
			})
		}

		if iteration >= 2 {
			if os.Getenv("GPTCODE_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[ORCHESTRATOR] Reached iteration limit - forcing final answer\n")
			}

			finalReq := ChatRequest{
				SystemPrompt: req.SystemPrompt + "\n\nCRITICAL: You have already called tools. Now provide your FINAL ANSWER based on the tool results. DO NOT call any more tools. Just answer the question directly.",
				Model:        o.customModel,
				Messages:     conversation,
				Tools:        nil,
			}

			finalResp, err := o.customExecutor.Chat(ctx, finalReq)
			if err != nil {
				return nil, err
			}

			return &ChatResponse{
				Text: finalResp.Text,
			}, nil
		}
	}

	return &ChatResponse{
		Text: "Maximum iterations reached. Task may be incomplete.",
	}, nil
}

func (o *OrchestratorProvider) ChatStream(ctx context.Context, req ChatRequest, callback func(chunk string)) error {
	resp, err := o.Chat(ctx, req)
	if err != nil {
		return err
	}
	callback(resp.Text)
	return nil
}
