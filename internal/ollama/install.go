package ollama

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type InstalledModel struct {
	Name string `json:"name"`
}

type InstalledModelsResponse struct {
	Models []InstalledModel `json:"models"`
}

func IsInstalled(modelName string) (bool, error) {
	if modelName == "" {
		return false, fmt.Errorf("model name cannot be empty")
	}

	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false, fmt.Errorf("ollama not running or not accessible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var installed InstalledModelsResponse
	if err := json.Unmarshal(body, &installed); err != nil {
		return false, err
	}

	for _, model := range installed.Models {
		if model.Name == modelName || strings.HasPrefix(model.Name, modelName+":") {
			return true, nil
		}
	}

	return false, nil
}

func Install(modelName string, progressCallback func(string)) error {
	cmd := exec.Command("ollama", "pull", modelName)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ollama pull: %w", err)
	}

	if progressCallback != nil {
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stdout.Read(buf)
				if n > 0 {
					progressCallback(string(buf[:n]))
				}
				if err != nil {
					break
				}
			}
		}()
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ollama pull failed: %w", err)
	}

	return nil
}

func CheckAndInstall(modelName string, autoInstall bool, progressCallback func(string)) error {
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	installed, err := IsInstalled(modelName)
	if err != nil {
		return err
	}

	if installed {
		return nil
	}

	if !autoInstall {
		return fmt.Errorf("model %s not installed. Run: ollama pull %s", modelName, modelName)
	}

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Installing %s...\n", modelName))
	}

	return Install(modelName, progressCallback)
}

func IsRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
