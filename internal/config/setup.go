package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func RunSetup() {
	home, _ := os.UserHomeDir()
	target := filepath.Join(home, ".chuchu")

	if err := os.MkdirAll(target, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "Chuchu: failed to create ~/.chuchu:", err)
		return
	}

	templateDir := detectTemplateDir()

	copyIfMissing(templateDir, target, "profile.yaml")
	copyIfMissing(templateDir, target, "system_prompt.md")

	setupPath := filepath.Join(target, "setup.yaml")
	if _, err := os.Stat(setupPath); err == nil {
		fmt.Fprintln(os.Stderr, "\nsetup.yaml already exists.")
		fmt.Fprint(os.Stderr, "Reconfigure? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y") {
			fmt.Fprintln(os.Stderr, "Chuchu: setup complete → ~/.chuchu")
			return
		}
	}

	setup := interactiveSetup()
	if err := saveSetup(setupPath, setup); err != nil {
		fmt.Fprintln(os.Stderr, "Chuchu: failed to save setup.yaml:", err)
		return
	}

	fmt.Fprintln(os.Stderr, "\nChuchu: setup complete → ~/.chuchu")
}

func detectTemplateDir() string {
	if env := os.Getenv("CHUCHU_TEMPLATES_DIR"); env != "" {
		return env
	}
	if _, err := os.Stat("internal/prompt/templates"); err == nil {
		return "internal/prompt/templates"
	}
	return "templates"
}


func LoadSetup() (*Setup, error) {
	path := filepath.Join(configDir(), "setup.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return &Setup{}, err
	}
	var s Setup
	if err := yaml.Unmarshal(b, &s); err != nil {
		return &Setup{}, err
	}
	return &s, nil
}

func interactiveSetup() *Setup {
	reader := bufio.NewReader(os.Stdin)
	setup := &Setup{
		Backend: make(map[string]BackendConfig),
	}

	fmt.Fprintln(os.Stderr, "\n=== Chuchu Setup ===")
	fmt.Fprintln(os.Stderr, "\nWhich LLM backends do you want to configure?")
	fmt.Fprintln(os.Stderr, "1) Local (Ollama)")
	fmt.Fprintln(os.Stderr, "2) OpenAI-compatible API (OpenAI, OpenRouter, etc)")
	fmt.Fprintln(os.Stderr, "3) Both")
	fmt.Fprint(os.Stderr, "\nChoice (1-3): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	useLocal := choice == "1" || choice == "3"
	useAPI := choice == "2" || choice == "3"

	if useLocal {
		fmt.Fprintln(os.Stderr, "\n--- Ollama (Local) ---")
		fmt.Fprint(os.Stderr, "Base URL [http://localhost:11434]: ")
		baseURL, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}

		for {
			fmt.Fprintln(os.Stderr, "\nModels (one or more, comma-separated):")
			fmt.Fprintln(os.Stderr, "  Examples: qwen3-coder,gpt-oss")
			fmt.Fprint(os.Stderr, "Models: ")
			modelsInput, _ := reader.ReadString('\n')
			modelsInput = strings.TrimSpace(modelsInput)
			if modelsInput == "" {
				fmt.Fprintln(os.Stderr, "At least one model is required")
				continue
			}

			modelsList := strings.Split(modelsInput, ",")
			modelsMap := make(map[string]string)
			for _, m := range modelsList {
				m = strings.TrimSpace(m)
				if m != "" {
					modelsMap[m] = m
				}
			}

			defaultModel := ""
			if len(modelsList) > 0 {
				defaultModel = strings.TrimSpace(modelsList[0])
			}

			setup.Backend["ollama"] = BackendConfig{
				Type:         "ollama",
				BaseURL:      baseURL,
				DefaultModel: defaultModel,
				Models:       modelsMap,
			}
			break
		}
	}

	if useAPI {
		for {
			fmt.Fprintln(os.Stderr, "\n--- OpenAI-compatible API Service ---")
			fmt.Fprintln(os.Stderr, "Examples: groq, openrouter, openai, deepseek, deepinfra")
			fmt.Fprint(os.Stderr, "\nService name (empty to finish): ")
			backendName, _ := reader.ReadString('\n')
			backendName = strings.TrimSpace(backendName)
			if backendName == "" {
				break
			}

			knownURLs := map[string]string{
				"groq":       "https://api.groq.com/openai/v1",
				"openrouter": "https://openrouter.ai/api/v1",
				"openai":     "https://api.openai.com/v1",
				"deepseek":   "https://api.deepseek.com/v1",
				"deepinfra":  "https://api.deepinfra.com/v1/openai",
			}

			defaultURL := knownURLs[backendName]
			if defaultURL != "" {
				fmt.Fprintf(os.Stderr, "Base URL [%s]: ", defaultURL)
			} else {
				fmt.Fprint(os.Stderr, "Base URL: ")
			}
			baseURL, _ := reader.ReadString('\n')
			baseURL = strings.TrimSpace(baseURL)
			if baseURL == "" {
				if defaultURL != "" {
					baseURL = defaultURL
				} else {
					fmt.Fprintln(os.Stderr, "Base URL is required, skipping...")
					continue
				}
			}

			fmt.Fprint(os.Stderr, "API Key: ")
			apiKey, _ := reader.ReadString('\n')
			apiKey = strings.TrimSpace(apiKey)

			fmt.Fprintln(os.Stderr, "\nModels (one or more, comma-separated):")
			fmt.Fprintln(os.Stderr, "  Example for Groq: llama-3.3-70b-versatile,llama-3.1-8b-instant")
			fmt.Fprintln(os.Stderr, "  Example for OpenRouter: kwaipilot/kat-coder-pro")
			fmt.Fprint(os.Stderr, "Models: ")
			modelsInput, _ := reader.ReadString('\n')
			modelsInput = strings.TrimSpace(modelsInput)
			if modelsInput == "" {
				fmt.Fprintln(os.Stderr, "At least one model is required, skipping...")
				continue
			}

			modelsList := strings.Split(modelsInput, ",")
			modelsMap := make(map[string]string)
			for _, m := range modelsList {
				m = strings.TrimSpace(m)
				if m != "" {
					modelsMap[m] = m
				}
			}

			defaultModel := ""
			if len(modelsList) > 0 {
				defaultModel = strings.TrimSpace(modelsList[0])
			}

			setup.Backend[backendName] = BackendConfig{
				Type:         "openai",
				BaseURL:      baseURL,
				DefaultModel: defaultModel,
				Models:       modelsMap,
			}

			if apiKey != "" {
				envVar := strings.ToUpper(backendName) + "_API_KEY"
				os.Setenv(envVar, apiKey)
				fmt.Fprintf(os.Stderr, "\nNote: Add 'export %s=%s' to your shell profile\n", envVar, apiKey)
			}
		}
	}

	fmt.Fprintln(os.Stderr, "\n--- Defaults ---")
	availableBackends := []string{}
	for name := range setup.Backend {
		availableBackends = append(availableBackends, name)
	}
	defaultBackend := ""
	if len(availableBackends) > 0 {
		defaultBackend = availableBackends[0]
	}
	fmt.Fprintf(os.Stderr, "Available backends: %s\n", strings.Join(availableBackends, ", "))
	fmt.Fprintf(os.Stderr, "Default backend [%s]: ", defaultBackend)
	backend, _ := reader.ReadString('\n')
	backend = strings.TrimSpace(backend)
	if backend == "" {
		backend = defaultBackend
	}
	setup.Defaults.Backend = backend

	defaultModel := ""
	if cfg, ok := setup.Backend[backend]; ok {
		defaultModel = cfg.DefaultModel
	}
	fmt.Fprintf(os.Stderr, "Default model [%s]: ", defaultModel)
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModel
	}
	setup.Defaults.Model = model

	fmt.Fprint(os.Stderr, "Default language [go]: ")
	lang, _ := reader.ReadString('\n')
	lang = strings.TrimSpace(lang)
	if lang == "" {
		lang = "go"
	}
	setup.Defaults.Lang = lang

	return setup
}

func saveSetup(path string, setup *Setup) error {
	data, err := yaml.Marshal(setup)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func copyIfMissing(srcDir, dstDir, file string) {
	src := filepath.Join(srcDir, file)
	dst := filepath.Join(dstDir, file)

	if _, err := os.Stat(dst); err == nil {
		fmt.Fprintln(os.Stderr, "keeping existing", dst)
		return
	}

	data, err := os.ReadFile(src)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not read template", src, ":", err)
		return
	}

	if err := os.WriteFile(dst, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "could not write", dst, ":", err)
		return
	}

	fmt.Fprintln(os.Stderr, "wrote", dst)
}
