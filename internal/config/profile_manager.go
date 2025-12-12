package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

func getSetupPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gptcode", "setup.yaml"), nil
}

type BackendProfile struct {
	Name        string
	AgentModels map[string]string
}

func loadSetupForProfiles() (*Setup, error) {
	setupPath, err := getSetupPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(setupPath)
	if err != nil {
		return nil, err
	}

	var setup Setup
	if err := yaml.Unmarshal(data, &setup); err != nil {
		return nil, err
	}

	return &setup, nil
}

func saveSetupForProfiles(setup *Setup) error {
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

func ListBackendProfiles(backendName string) ([]string, error) {
	setup, err := loadSetupForProfiles()
	if err != nil {
		return nil, err
	}

	backend, ok := setup.Backend[backendName]
	if !ok {
		return nil, fmt.Errorf("backend %s not found", backendName)
	}

	profiles := []string{}

	if len(backend.AgentModels.Router) > 0 || len(backend.AgentModels.Query) > 0 ||
		len(backend.AgentModels.Editor) > 0 || len(backend.AgentModels.Research) > 0 {
		profiles = append(profiles, "default")
	}

	for name := range backend.Profiles {
		profiles = append(profiles, name)
	}

	sort.Strings(profiles)

	if len(profiles) == 0 {
		profiles = append(profiles, "default")
	}

	return profiles, nil
}

func GetBackendProfile(backendName, profileName string) (*BackendProfile, error) {
	setup, err := loadSetupForProfiles()
	if err != nil {
		return nil, err
	}

	backend, ok := setup.Backend[backendName]
	if !ok {
		return nil, fmt.Errorf("backend %s not found", backendName)
	}

	profile := &BackendProfile{
		Name:        profileName,
		AgentModels: make(map[string]string),
	}

	var agentModels AgentModels

	if profileName == "default" {
		agentModels = backend.AgentModels
	} else if profileCfg, ok := backend.Profiles[profileName]; ok {
		agentModels = profileCfg.AgentModels
	} else {
		return nil, fmt.Errorf("profile %s not found in backend %s", profileName, backendName)
	}

	if agentModels.Router != "" {
		profile.AgentModels["router"] = agentModels.Router
	}
	if agentModels.Query != "" {
		profile.AgentModels["query"] = agentModels.Query
	}
	if agentModels.Editor != "" {
		profile.AgentModels["editor"] = agentModels.Editor
	}
	if agentModels.Research != "" {
		profile.AgentModels["research"] = agentModels.Research
	}

	return profile, nil
}

func CreateBackendProfile(backendName, profileName string) error {
	setup, err := loadSetupForProfiles()
	if err != nil {
		return err
	}

	backend, ok := setup.Backend[backendName]
	if !ok {
		return fmt.Errorf("backend %s not found", backendName)
	}

	if backend.Profiles == nil {
		backend.Profiles = make(map[string]ProfileConfig)
	}

	if _, exists := backend.Profiles[profileName]; exists {
		return fmt.Errorf("profile %s already exists", profileName)
	}

	backend.Profiles[profileName] = ProfileConfig{
		AgentModels: AgentModels{},
	}

	setup.Backend[backendName] = backend

	return saveSetupForProfiles(setup)
}

func DeleteBackendProfile(backendName, profileName string) error {
	if profileName == "default" {
		return fmt.Errorf("cannot delete default profile")
	}

	setup, err := loadSetupForProfiles()
	if err != nil {
		return err
	}

	backend, ok := setup.Backend[backendName]
	if !ok {
		return fmt.Errorf("backend %s not found", backendName)
	}

	if backend.Profiles == nil {
		return fmt.Errorf("profile %s not found", profileName)
	}

	if _, exists := backend.Profiles[profileName]; !exists {
		return fmt.Errorf("profile %s not found", profileName)
	}

	if setup.Defaults.Backend == backendName && setup.Defaults.Profile == profileName {
		return fmt.Errorf("cannot delete profile %s: it is currently active", profileName)
	}

	delete(backend.Profiles, profileName)
	setup.Backend[backendName] = backend

	return saveSetupForProfiles(setup)
}

func SetProfileAgentModel(backendName, profileName, agent, model string) error {
	setup, err := loadSetupForProfiles()
	if err != nil {
		return err
	}

	backend, ok := setup.Backend[backendName]
	if !ok {
		return fmt.Errorf("backend %s not found", backendName)
	}

	if profileName == "default" {
		switch agent {
		case "router":
			backend.AgentModels.Router = model
		case "query":
			backend.AgentModels.Query = model
		case "editor":
			backend.AgentModels.Editor = model
		case "research":
			backend.AgentModels.Research = model
		default:
			return fmt.Errorf("invalid agent type: %s", agent)
		}
	} else {
		if backend.Profiles == nil {
			return fmt.Errorf("profile %s not found", profileName)
		}

		profileCfg, ok := backend.Profiles[profileName]
		if !ok {
			return fmt.Errorf("profile %s not found", profileName)
		}

		switch agent {
		case "router":
			profileCfg.AgentModels.Router = model
		case "query":
			profileCfg.AgentModels.Query = model
		case "editor":
			profileCfg.AgentModels.Editor = model
		case "research":
			profileCfg.AgentModels.Research = model
		default:
			return fmt.Errorf("invalid agent type: %s", agent)
		}

		backend.Profiles[profileName] = profileCfg
	}

	setup.Backend[backendName] = backend

	return saveSetupForProfiles(setup)
}
