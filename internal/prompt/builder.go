package prompt

import (
	"fmt"
	"os"
	"path/filepath"
)

type BuildOptions struct {
	Lang string
	Mode string
	Hint string
}

type Builder struct {
	ProfilePath string
	SystemPath  string
	Store       MemoryStore
}

func NewDefaultBuilder(store MemoryStore) *Builder {
	home, _ := os.UserHomeDir()
	profile := filepath.Join(home, ".gptcode", "profile.yaml")
	system := filepath.Join(home, ".gptcode", "system_prompt.md")
	return &Builder{
		ProfilePath: profile,
		SystemPath:  system,
		Store:       store,
	}
}

func (b *Builder) BuildSystemPrompt(opts BuildOptions) string {
	base := mustReadFile(b.SystemPath)
	profile := mustReadFile(b.ProfilePath)
	mem := ""
	if b.Store != nil {
		mem = b.Store.LastRelevant(opts.Lang)
	}

	return fmt.Sprintf(`%s

---

# GPTCode Profile (YAML)

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
		return ""
	}
	return string(data)
}

type MemoryStore interface {
	LastRelevant(lang string) string
}
