package prompt

import (
	"fmt"
	"os"
	"path/filepath"
)

// BuildOptions holds context for building the system prompt.
type BuildOptions struct {
	Lang string
	Mode string
	Hint string
}

// Builder combines profile + base system prompt + memory.
type Builder struct {
	ProfilePath string
	SystemPath  string
	Store       MemoryStore // your existing memory abstraction
}

// NewDefaultBuilder creates a builder pointing to the default prompt files.
func NewDefaultBuilder(store MemoryStore) *Builder {
	home, _ := os.UserHomeDir()
	// These can be overridden via config if you prefer.
	profile := filepath.Join(home, ".chuchu", "profile.yaml")
	system := filepath.Join(home, ".chuchu", "system_prompt.md")
	return &Builder{
		ProfilePath: profile,
		SystemPath:  system,
		Store:       store,
	}
}

// BuildSystemPrompt loads the base system prompt, profile, and relevant memory
// and composes a single system prompt string to send to the LLM.
func (b *Builder) BuildSystemPrompt(opts BuildOptions) string {
	base := mustReadFile(b.SystemPath)
	profile := mustReadFile(b.ProfilePath)
	mem := ""
	if b.Store != nil {
		mem = b.Store.LastRelevant(opts.Lang)
	}

	return fmt.Sprintf(`%s

---

# Chuchu Profile (YAML)

%s

---

# Relevant Memory

%s

---

# Current Session Context

Language: %s
Mode: %s
Hint: %s
`, base, profile, mem, opts.Lang, opts.Mode, opts.Hint)
}

func mustReadFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		// If file missing, degrade gracefully to empty string.
		return ""
	}
	return string(data)
}

// MemoryStore is an interface you can implement on top of ~/.chuchu/memories.jsonl.
type MemoryStore interface {
	LastRelevant(lang string) string
}
