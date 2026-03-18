package autonomous

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type StateManager struct {
	cwd       string
	stateFile string
	state     *State
	mu        sync.RWMutex
}

type State struct {
	SessionID   string       `json:"session_id"`
	Tasks       []TaskState  `json:"tasks"`
	LastUpdate  time.Time    `json:"last_update"`
	Checkpoints []Checkpoint `json:"checkpoints"`
}

type TaskState struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"` // pending, running, completed, failed
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      string     `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
	Progress    float64    `json:"progress"` // 0-100
}

type Checkpoint struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

func NewStateManager(cwd string) (*StateManager, error) {
	sm := &StateManager{
		cwd:       cwd,
		stateFile: filepath.Join(cwd, ".gptcode", "state.json"),
		state:     &State{},
	}

	if err := sm.load(); err != nil {
		// State file doesn't exist yet - that's OK
		sm.state.SessionID = generateID()
	}

	return sm, nil
}

func (sm *StateManager) load() error {
	data, err := os.ReadFile(sm.stateFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, sm.state)
}

func (sm *StateManager) save() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.state.LastUpdate = time.Now()

	dir := filepath.Dir(sm.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(sm.state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.stateFile, data, 0644)
}

func (sm *StateManager) StartTask(name string) string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	task := TaskState{
		ID:        generateID(),
		Name:      name,
		Status:    "running",
		StartedAt: time.Now(),
		Progress:  0,
	}

	sm.state.Tasks = append(sm.state.Tasks, task)
	sm.save()

	return task.ID
}

func (sm *StateManager) UpdateTask(id string, progress float64, result string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i := range sm.state.Tasks {
		if sm.state.Tasks[i].ID == id {
			sm.state.Tasks[i].Progress = progress
			if result != "" {
				sm.state.Tasks[i].Result = result
			}
			return sm.save()
		}
	}

	return fmt.Errorf("task not found: %s", id)
}

func (sm *StateManager) CompleteTask(id string, result string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for i := range sm.state.Tasks {
		if sm.state.Tasks[i].ID == id {
			sm.state.Tasks[i].Status = "completed"
			sm.state.Tasks[i].CompletedAt = &now
			sm.state.Tasks[i].Progress = 100
			sm.state.Tasks[i].Result = result
			return sm.save()
		}
	}

	return fmt.Errorf("task not found: %s", id)
}

func (sm *StateManager) FailTask(id string, errMsg string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for i := range sm.state.Tasks {
		if sm.state.Tasks[i].ID == id {
			sm.state.Tasks[i].Status = "failed"
			sm.state.Tasks[i].CompletedAt = &now
			sm.state.Tasks[i].Error = errMsg
			return sm.save()
		}
	}

	return fmt.Errorf("task not found: %s", id)
}

func (sm *StateManager) SaveCheckpoint(taskID string, data string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	checkpoint := Checkpoint{
		ID:        generateID(),
		TaskID:    taskID,
		Data:      data,
		CreatedAt: time.Now(),
	}

	sm.state.Checkpoints = append(sm.state.Checkpoints, checkpoint)
	return sm.save()
}

func (sm *StateManager) GetCheckpoint(taskID string) (*Checkpoint, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for i := len(sm.state.Checkpoints) - 1; i >= 0; i-- {
		if sm.state.Checkpoints[i].TaskID == taskID {
			return &sm.state.Checkpoints[i], nil
		}
	}

	return nil, fmt.Errorf("no checkpoint found for task: %s", taskID)
}

func (sm *StateManager) GetTask(id string) (*TaskState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, task := range sm.state.Tasks {
		if task.ID == id {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task not found: %s", id)
}

func (sm *StateManager) GetRunningTasks() []TaskState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var running []TaskState
	for _, task := range sm.state.Tasks {
		if task.Status == "running" {
			running = append(running, task)
		}
	}

	return running
}

func (sm *StateManager) ResumeTask(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i := range sm.state.Tasks {
		if sm.state.Tasks[i].ID == id {
			sm.state.Tasks[i].Status = "running"
			return sm.save()
		}
	}

	return fmt.Errorf("task not found: %s", id)
}

func (sm *StateManager) ClearCompleted() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var pending []TaskState
	for _, task := range sm.state.Tasks {
		if task.Status != "completed" && task.Status != "failed" {
			pending = append(pending, task)
		}
	}

	sm.state.Tasks = pending
	sm.state.Checkpoints = nil // Clear checkpoints too

	return sm.save()
}

func (sm *StateManager) GetSummary() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	total := len(sm.state.Tasks)
	completed := 0
	failed := 0
	running := 0

	for _, task := range sm.state.Tasks {
		switch task.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		case "running":
			running++
		}
	}

	return fmt.Sprintf("Tasks: %d total, %d completed, %d failed, %d running",
		total, completed, failed, running)
}

func (sm *StateManager) Export() ([]byte, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return json.MarshalIndent(sm.state, "", "  ")
}

func (sm *StateManager) Import(data []byte) error {
	var imported State
	if err := json.Unmarshal(data, &imported); err != nil {
		return err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state = &imported
	return sm.save()
}
