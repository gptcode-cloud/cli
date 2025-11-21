package langdetect

import (
	"os"
	"path/filepath"
	"strings"
)

type Language string

const (
	Elixir     Language = "elixir"
	Ruby       Language = "ruby"
	Go         Language = "go"
	TypeScript Language = "typescript"
	Python     Language = "python"
	Unknown    Language = "unknown"
)

func DetectLanguage(path string) Language {
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return Unknown
	}

	if fileExists(filepath.Join(absPath, "mix.exs")) {
		return Elixir
	}

	if fileExists(filepath.Join(absPath, "Gemfile")) ||
		fileExists(filepath.Join(absPath, "config", "application.rb")) {
		return Ruby
	}

	if fileExists(filepath.Join(absPath, "go.mod")) {
		return Go
	}

	if fileExists(filepath.Join(absPath, "tsconfig.json")) ||
		fileExists(filepath.Join(absPath, "package.json")) {
		return TypeScript
	}

	if fileExists(filepath.Join(absPath, "requirements.txt")) ||
		fileExists(filepath.Join(absPath, "setup.py")) ||
		fileExists(filepath.Join(absPath, "pyproject.toml")) {
		return Python
	}

	langCounts := make(map[Language]int)
	
	walkErr := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if info.IsDir() {
			return nil
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".ex", ".exs":
			langCounts[Elixir]++
		case ".rb", ".rake":
			langCounts[Ruby]++
		case ".go":
			langCounts[Go]++
		case ".ts", ".tsx", ".js", ".jsx":
			langCounts[TypeScript]++
		case ".py":
			langCounts[Python]++
		}
		
		return nil
	})
	
	if walkErr != nil {
		return Unknown
	}

	maxCount := 0
	detected := Unknown
	for lang, count := range langCounts {
		if count > maxCount {
			maxCount = count
			detected = lang
		}
	}

	return detected
}

func DetectFromFilename(filename string) Language {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".ex", ".exs":
		return Elixir
	case ".rb", ".rake":
		return Ruby
	case ".go":
		return Go
	case ".ts", ".tsx", ".js", ".jsx":
		return TypeScript
	case ".py":
		return Python
	default:
		return Unknown
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
