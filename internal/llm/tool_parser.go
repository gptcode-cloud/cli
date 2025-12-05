package llm

import (
	"encoding/json"
	"regexp"
	"strings"
)

func ParseToolCallsFromText(text string) []ChatToolCall {
	var calls []ChatToolCall

	calls = append(calls, parsePythonStyle(text)...)
	calls = append(calls, parseXMLStyle(text)...)
	calls = append(calls, parseGroqStyle(text)...)
	calls = append(calls, parseSimpleStyle(text)...)

	return calls
}

func parsePythonStyle(text string) []ChatToolCall {
	re := regexp.MustCompile(`\[(\w+)\((.*?)\)\]`)
	matches := re.FindAllStringSubmatch(text, -1)

	var calls []ChatToolCall
	for i, match := range matches {
		if len(match) < 3 {
			continue
		}

		funcName := match[1]
		argsStr := match[2]

		argsMap := make(map[string]interface{})
		if argsStr != "" {
			argPairs := strings.Split(argsStr, ",")
			for _, pair := range argPairs {
				parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
					argsMap[key] = val
				}
			}
		}

		argsJSON, _ := json.Marshal(argsMap)
		calls = append(calls, ChatToolCall{
			ID:        generateID("call", i),
			Name:      funcName,
			Arguments: string(argsJSON),
		})
	}

	return calls
}

func parseXMLStyle(text string) []ChatToolCall {
	re := regexp.MustCompile(`<function=(\w+)=?(.*?)>`)
	matches := re.FindAllStringSubmatch(text, -1)

	var calls []ChatToolCall
	for i, match := range matches {
		if len(match) < 2 {
			continue
		}

		funcName := match[1]
		var argsJSON string

		if len(match) > 2 && match[2] != "" {
			argsStr := strings.TrimSpace(match[2])

			if strings.HasPrefix(argsStr, "{") && strings.HasSuffix(argsStr, "}") {
				argsJSON = argsStr
			} else {
				argsMap := make(map[string]interface{})
				argPairs := strings.Split(argsStr, ",")
				for _, pair := range argPairs {
					parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						val := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
						argsMap[key] = val
					}
				}
				jsonBytes, _ := json.Marshal(argsMap)
				argsJSON = string(jsonBytes)
			}
		} else {
			argsJSON = "{}"
		}

		calls = append(calls, ChatToolCall{
			ID:        generateID("call", i),
			Name:      funcName,
			Arguments: argsJSON,
		})
	}

	return calls
}

func parseGroqStyle(text string) []ChatToolCall {
	re := regexp.MustCompile(`(\w+)\((.*?)\)</function>`)
	matches := re.FindAllStringSubmatch(text, -1)

	var calls []ChatToolCall
	for i, match := range matches {
		if len(match) < 3 {
			continue
		}

		funcName := match[1]
		argsStr := match[2]

		argsMap := make(map[string]interface{})
		if argsStr != "" {
			argPairs := strings.Split(argsStr, ",")
			for _, pair := range argPairs {
				parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
					argsMap[key] = val
				}
			}
		}

		argsJSON, _ := json.Marshal(argsMap)
		calls = append(calls, ChatToolCall{
			ID:        generateID("call", i),
			Name:      funcName,
			Arguments: string(argsJSON),
		})
	}

	return calls
}

// parseSimpleStyle handles bare function calls like: run_command(command="git log")
func parseSimpleStyle(text string) []ChatToolCall {
	// Match: function_name(arg="value") or function_name(arg='value')
	re := regexp.MustCompile(`(?m)^\s*(\w+)\((.*)\)\s*$`)
	matches := re.FindAllStringSubmatch(text, -1)

	var calls []ChatToolCall
	for i, match := range matches {
		if len(match) < 3 {
			continue
		}

		funcName := match[1]
		argsStr := match[2]

		argsMap := make(map[string]interface{})
		if argsStr != "" {
			// Split by comma, but respect quotes
			argPairs := splitArgs(argsStr)
			for _, pair := range argPairs {
				parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
					argsMap[key] = val
				}
			}
		}

		argsJSON, _ := json.Marshal(argsMap)
		calls = append(calls, ChatToolCall{
			ID:        generateID("call", i),
			Name:      funcName,
			Arguments: string(argsJSON),
		})
	}

	return calls
}

// splitArgs splits function arguments by comma, respecting quotes
func splitArgs(argsStr string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range argsStr {
		if ch == '"' || ch == '\'' {
			if !inQuote {
				inQuote = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuote = false
			}
			current.WriteRune(ch)
		} else if ch == ',' && !inQuote {
			args = append(args, current.String())
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func generateID(prefix string, index int) string {
	return prefix + "_" + string(rune('a'+index))
}
