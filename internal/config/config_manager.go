package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func GetConfig(key string) (string, error) {
	setup, err := LoadSetup()
	if err != nil {
		return "", err
	}

	value, err := getNestedValue(setup, key)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", value), nil
}

func SetConfig(key, value string) error {
	setup, err := LoadSetup()
	if err != nil {
		return err
	}

	if err := setNestedValue(setup, key, value); err != nil {
		return err
	}

	return saveSetupConfig(setup)
}

func getNestedValue(setup *Setup, key string) (interface{}, error) {
	parts := strings.Split(key, ".")
	
	switch parts[0] {
	case "defaults":
		if len(parts) < 2 {
			return nil, fmt.Errorf("defaults key requires subfield (e.g., defaults.backend)")
		}
		switch parts[1] {
		case "backend":
			return setup.Defaults.Backend, nil
		case "profile":
			return setup.Defaults.Profile, nil
		case "model":
			return setup.Defaults.Model, nil
		case "lang":
			return setup.Defaults.Lang, nil
		case "system_prompt_file":
			return setup.Defaults.SystemPromptFile, nil
		case "ml_complex_threshold":
			return setup.Defaults.MLComplexThreshold, nil
		case "ml_intent_threshold":
			return setup.Defaults.MLIntentThreshold, nil
		default:
			return nil, fmt.Errorf("unknown defaults field: %s", parts[1])
		}
	
	case "backend":
		if len(parts) == 1 {
			var backends []string
			for name := range setup.Backend {
				backends = append(backends, name)
			}
			return backends, nil
		}
		
		if len(parts) < 3 {
			return nil, fmt.Errorf("backend key requires: backend.<name>.<field>")
		}
		backendName := parts[1]
		backend, ok := setup.Backend[backendName]
		if !ok {
			return nil, fmt.Errorf("backend %s not found", backendName)
		}
		
		switch parts[2] {
		case "type":
			return backend.Type, nil
		case "base_url":
			return backend.BaseURL, nil
		case "default_model":
			return backend.DefaultModel, nil
		default:
			return nil, fmt.Errorf("unknown backend field: %s", parts[2])
		}
	
	default:
		return nil, fmt.Errorf("unknown config section: %s", parts[0])
	}
}

func setNestedValue(setup *Setup, key, value string) error {
	parts := strings.Split(key, ".")
	
	switch parts[0] {
	case "defaults":
		if len(parts) < 2 {
			return fmt.Errorf("defaults key requires subfield (e.g., defaults.backend)")
		}
		switch parts[1] {
		case "backend":
			if _, ok := setup.Backend[value]; !ok {
				return fmt.Errorf("backend %s does not exist", value)
			}
			setup.Defaults.Backend = value
		case "profile":
			setup.Defaults.Profile = value
		case "model":
			setup.Defaults.Model = value
		case "lang":
			setup.Defaults.Lang = value
		case "system_prompt_file":
			setup.Defaults.SystemPromptFile = value
		case "ml_complex_threshold":
			var f float64
			if _, err := fmt.Sscan(value, &f); err != nil {
				return fmt.Errorf("invalid float value for ml_complex_threshold: %s", value)
			}
			setup.Defaults.MLComplexThreshold = f
		case "ml_intent_threshold":
			var f float64
			if _, err := fmt.Sscan(value, &f); err != nil {
				return fmt.Errorf("invalid float value for ml_intent_threshold: %s", value)
			}
			setup.Defaults.MLIntentThreshold = f
		default:
			return fmt.Errorf("unknown defaults field: %s", parts[1])
		}
	
	case "backend":
		if len(parts) < 3 {
			return fmt.Errorf("backend key requires: backend.<name>.<field>")
		}
		backendName := parts[1]
		backend, ok := setup.Backend[backendName]
		if !ok {
			return fmt.Errorf("backend %s not found. Use 'chu backend create %s' first", backendName, backendName)
		}
		
		switch parts[2] {
		case "type":
			backend.Type = value
		case "base_url":
			backend.BaseURL = value
		case "default_model":
			backend.DefaultModel = value
		default:
			return fmt.Errorf("unknown backend field: %s", parts[2])
		}
		
		setup.Backend[backendName] = backend
	
	default:
		return fmt.Errorf("unknown config section: %s", parts[0])
	}
	
	return nil
}

func CreateBackend(name, backendType, baseURL string) error {
	setup, err := LoadSetup()
	if err != nil {
		return err
	}
	
	if _, exists := setup.Backend[name]; exists {
		return fmt.Errorf("backend %s already exists", name)
	}
	
	setup.Backend[name] = BackendConfig{
		Type:    backendType,
		BaseURL: baseURL,
		Models:  make(map[string]string),
	}
	
	return saveSetupConfig(setup)
}

func DeleteBackend(name string) error {
	setup, err := LoadSetup()
	if err != nil {
		return err
	}
	
	if _, exists := setup.Backend[name]; !exists {
		return fmt.Errorf("backend %s does not exist", name)
	}
	
	if setup.Defaults.Backend == name {
		return fmt.Errorf("cannot delete backend %s: it is set as default", name)
	}
	
	delete(setup.Backend, name)
	
	return saveSetupConfig(setup)
}

func ListBackends() ([]string, error) {
	setup, err := LoadSetup()
	if err != nil {
		return nil, err
	}
	
	var backends []string
	for name := range setup.Backend {
		backends = append(backends, name)
	}
	
	return backends, nil
}

func saveSetupConfig(setup *Setup) error {
	setupPath, err := getSetupPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(setup)
	if err != nil {
		return err
	}

	return os.WriteFile(setupPath, data, 0644)
}
