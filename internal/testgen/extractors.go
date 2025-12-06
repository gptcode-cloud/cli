package testgen

import (
	"strings"
)

func extractCode(response string) string {
	lines := strings.Split(response, "\n")
	var code []string
	inCodeBlock := false
	foundCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				foundCodeBlock = true
				break
			}
			inCodeBlock = true
			continue
		}
		
		if inCodeBlock {
			code = append(code, line)
		}
	}

	if foundCodeBlock && len(code) > 0 {
		return strings.Join(code, "\n")
	}

	return response
}
