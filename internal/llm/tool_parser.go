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

func generateID(prefix string, index int) string {
	return prefix + "_" + string(rune('a'+index))
}
