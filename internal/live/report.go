package live

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReportConfig holds the HTTP report configuration
type ReportConfig struct {
	BaseURL   string
	Client    *http.Client
	AgentID   string
	Hostname  string
	Workspace string
	Task      string
	Model     string
}

// DefaultReportConfig creates a new report config
func DefaultReportConfig() *ReportConfig {
	hostname, _ := os.Hostname()
	cwd, _ := os.Getwd()
	workspace := filepath.Base(cwd)

	return &ReportConfig{
		BaseURL:   "http://localhost:4004",
		Client:    &http.Client{Timeout: 5 * time.Second},
		Hostname:  hostname,
		Workspace: workspace,
	}
}

// SetBaseURL sets the base URL for HTTP API
func (r *ReportConfig) SetBaseURL(url string) {
	// Convert ws:// to http:// for HTTP API
	if strings.HasPrefix(url, "ws://") {
		url = "http://" + strings.TrimPrefix(url, "ws://")
	} else if strings.HasPrefix(url, "wss://") {
		url = "https://" + strings.TrimPrefix(url, "wss://")
	}
	r.BaseURL = url
}

// Connect reports agent connection to Live Dashboard
func (r *ReportConfig) Connect(agentID, agentType, task string) error {
	r.AgentID = agentID
	r.Task = task

	// Override workspace from cwd if provided
	if cwd, err := os.Getwd(); err == nil {
		r.Workspace = filepath.Base(cwd)
	}

	payload := map[string]interface{}{
		"agent_id": agentID,
		"engine":   "gptcode",
		"type":     agentType,
		"context":  r.Workspace,
		"task":     task,
		"hostname": r.Hostname,
		"model":    r.Model,
	}

	return r.post("/api/report/connect", payload)
}

// Step reports a step to Live Dashboard
func (r *ReportConfig) Step(description string, stepType string) error {
	if r.AgentID == "" {
		return nil // Not connected
	}

	payload := map[string]interface{}{
		"agent_id":    r.AgentID,
		"description": description,
	}

	if stepType != "" {
		payload["type"] = stepType
	}

	return r.post("/api/report/step", payload)
}

// Disconnect reports agent disconnect to Live Dashboard
func (r *ReportConfig) Disconnect() error {
	if r.AgentID == "" {
		return nil // Not connected
	}

	payload := map[string]interface{}{
		"agent_id": r.AgentID,
	}

	return r.post("/api/report/disconnect", payload)
}

func (r *ReportConfig) post(endpoint string, payload map[string]interface{}) error {
	url := r.BaseURL + endpoint

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := r.Client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	log.Printf("[Live Report] %s -> %s", endpoint, payload)
	return nil
}
