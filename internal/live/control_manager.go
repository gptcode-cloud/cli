package live

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
)

// ControlManager provides global pause/resume/kill support
// for ALL CLI subcommands, not just the Maestro orchestrator.
// It is initialized once in PersistentPreRunE and accessible globally.
type ControlManager struct {
	isPaused  int32
	pauseCond *sync.Cond
	mu        sync.Mutex

	// OnPause/OnResume/OnKill are optional hooks for subcommand-specific behavior
	OnPause  func()
	OnResume func()
	OnKill   func()
}

var (
	globalControl     *ControlManager
	globalControlOnce sync.Once
)

// GetControlManager returns the singleton ControlManager
func GetControlManager() *ControlManager {
	globalControlOnce.Do(func() {
		cm := &ControlManager{}
		cm.pauseCond = sync.NewCond(&cm.mu)
		globalControl = cm
	})
	return globalControl
}

// RegisterWithClient wires the ControlManager to the live WebSocket client.
// Called once during PersistentPreRunE after the client connects.
func (cm *ControlManager) RegisterWithClient(client *Client) {
	if client == nil {
		return
	}

	client.OnCommand(func(command string, payload map[string]interface{}) {
		switch command {
		case "pause":
			cm.Pause()
		case "resume":
			cm.Resume()
		case "kill":
			cm.Kill()
		case "update_prompt":
			cm.handlePrompt(payload)
		default:
			log.Printf("Live: unknown control command: %s", command)
		}
	})
}

// Pause halts execution at the next checkpoint
func (cm *ControlManager) Pause() {
	atomic.StoreInt32(&cm.isPaused, 1)
	log.Printf("Live: Agent PAUSED by dashboard")
	if cm.OnPause != nil {
		cm.OnPause()
	}
}

// Resume continues execution
func (cm *ControlManager) Resume() {
	atomic.StoreInt32(&cm.isPaused, 0)
	cm.pauseCond.Broadcast()
	log.Printf("Live: Agent RESUMED by dashboard")
	if cm.OnResume != nil {
		cm.OnResume()
	}
}

// Kill terminates the process
func (cm *ControlManager) Kill() {
	log.Printf("Live: Agent KILLED by dashboard")
	if cm.OnKill != nil {
		cm.OnKill()
	}
	os.Exit(1)
}

// CheckPause blocks the calling goroutine if the agent is paused.
// Call this at strategic checkpoints in any subcommand's execution loop.
func (cm *ControlManager) CheckPause() {
	if atomic.LoadInt32(&cm.isPaused) == 1 {
		cm.mu.Lock()
		for atomic.LoadInt32(&cm.isPaused) == 1 {
			cm.pauseCond.Wait()
		}
		cm.mu.Unlock()
	}
}

// IsPaused returns whether the agent is currently paused
func (cm *ControlManager) IsPaused() bool {
	return atomic.LoadInt32(&cm.isPaused) == 1
}

// handlePrompt receives a prompt sent from the Dashboard and writes it
// to .gptcode/context/prompt.md so any agent (including external ones)
// can pick it up.
func (cm *ControlManager) handlePrompt(payload map[string]interface{}) {
	prompt, _ := payload["prompt"].(string)
	if prompt == "" {
		return
	}

	log.Printf("Live: Received remote prompt (%d chars)", len(prompt))

	// Write to context file so any watcher (CLI, n8n, external agent) can pick it up
	if err := WriteContextFile("prompt", prompt); err != nil {
		// If .gptcode dir doesn't exist, write to a temp location
		tmpPath := fmt.Sprintf("/tmp/gptcode-prompt-%s.md", GetAgentID())
		if writeErr := os.WriteFile(tmpPath, []byte(prompt), 0644); writeErr != nil {
			log.Printf("Live: failed to write prompt: %v", writeErr)
		} else {
			log.Printf("Live: Prompt written to %s", tmpPath)
		}
	}
}
