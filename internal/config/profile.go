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
