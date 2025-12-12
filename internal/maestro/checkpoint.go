package maestro

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Checkpoint represents a saved state of the execution
type Checkpoint struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Step      int               `json:"step"`
	Files     map[string]string `json:"files"` // path -> content hash
}

// CheckpointSystem manages saving and loading checkpoints
type CheckpointSystem struct {
	RootDir string
}

// NewCheckpointSystem creates a new checkpoint system
func NewCheckpointSystem(rootDir string) *CheckpointSystem {
	return &CheckpointSystem{
		RootDir: filepath.Join(rootDir, ".gptcode", "checkpoints"),
	}
}

// Save creates a checkpoint for the current state
// For MVP, we'll just save the plan step.
// For file rollback, we might need a more complex solution (like git or full copy).
// Let's assume we rely on git for file versioning for now, and this just tracks progress?
// The plan said "Snapshot of file hashes".
// If we want true rollback without git, we need to copy files.
// Let's implement a simple file backup for modified files.
func (cs *CheckpointSystem) Save(step int, modifiedFiles []string) (*Checkpoint, error) {
	if err := os.MkdirAll(cs.RootDir, 0755); err != nil {
		return nil, err
	}

	id := fmt.Sprintf("ckpt_%d_%d", step, time.Now().Unix())
	ckptDir := filepath.Join(cs.RootDir, id)
	if err := os.MkdirAll(ckptDir, 0755); err != nil {
		return nil, err
	}

	files := make(map[string]string)

	// Backup modified files
	for _, file := range modifiedFiles {
		// Calculate hash
		hash, err := hashFile(file)
		if err != nil {
			continue // Skip if file doesn't exist or error
		}
		files[file] = hash

		// Copy file content to checkpoint dir
		// We use the hash as filename to avoid directory structure issues
		dst := filepath.Join(ckptDir, hash)
		if err := copyFile(file, dst); err != nil {
			return nil, err
		}
	}

	ckpt := &Checkpoint{
		ID:        id,
		Timestamp: time.Now(),
		Step:      step,
		Files:     files,
	}

	// Save metadata
	metaPath := filepath.Join(ckptDir, "metadata.json")
	data, err := json.MarshalIndent(ckpt, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return nil, err
	}

	return ckpt, nil
}

// Restore restores files from a checkpoint
func (cs *CheckpointSystem) Restore(id string) error {
	ckptDir := filepath.Join(cs.RootDir, id)
	metaPath := filepath.Join(ckptDir, "metadata.json")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		return err
	}

	var ckpt Checkpoint
	if err := json.Unmarshal(data, &ckpt); err != nil {
		return err
	}

	// Restore files
	for path, hash := range ckpt.Files {
		src := filepath.Join(ckptDir, hash)
		if err := copyFile(src, path); err != nil {
			return fmt.Errorf("failed to restore %s: %w", path, err)
		}
	}

	return nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
