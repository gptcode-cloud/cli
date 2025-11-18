package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Philosophy []string        `yaml:"philosophy"`
	Languages  []string        `yaml:"languages"`
	Style      map[string]any  `yaml:"style"`
	Defaults   ProfileDefaults `yaml:"defaults"`
}

type Setup struct {
	Defaults struct {
		Backend string `yaml:"backend"`
		Model   string `yaml:"model"`
		Lang    string `yaml:"lang"`
	} `yaml:"defaults"`
	Backend map[string]BackendConfig `yaml:"backend"`
}

type BackendConfig struct {
	Type         string            `yaml:"type"`
	BaseURL      string            `yaml:"base_url"`
	DefaultModel string            `yaml:"default_model"`
	Models       map[string]string `yaml:"models"`
	AgentModels  AgentModels       `yaml:"agent_models,omitempty"`
}

type AgentModels struct {
	Router   string `yaml:"router,omitempty"`
	Query    string `yaml:"query,omitempty"`
	Editor   string `yaml:"editor,omitempty"`
	Research string `yaml:"research,omitempty"`
}

type ProfileDefaults struct {
	Backend string `yaml:"backend"`
	Model   string `yaml:"model"`
}

func configDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".chuchu")
}

func LoadProfile() (*Profile, error) {
	path := filepath.Join(configDir(), "profile.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return &Profile{}, err
	}
	var p Profile
	if err := yaml.Unmarshal(b, &p); err != nil {
		return &Profile{}, err
	}
	return &p, nil
}

func (bc *BackendConfig) GetModelForAgent(agentType string) string {
	switch agentType {
	case "router":
		if bc.AgentModels.Router != "" {
			return bc.AgentModels.Router
		}
	case "query":
		if bc.AgentModels.Query != "" {
			return bc.AgentModels.Query
		}
	case "editor":
		if bc.AgentModels.Editor != "" {
			return bc.AgentModels.Editor
		}
	case "research":
		if bc.AgentModels.Research != "" {
			return bc.AgentModels.Research
		}
	}
	return bc.DefaultModel
}
