package memory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Store interface {
	LastRelevant(lang string) string
}

type JSONLMemStore struct {
	Path         string
	MaxEntries   int
	GlobalMaxLen int
	FallbackLang string
}

func NewJSONLMemStore() *JSONLMemStore {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".gptcode", "memories.jsonl")
	return &JSONLMemStore{
		Path:         path,
		MaxEntries:   5,
		GlobalMaxLen: 4000,
		FallbackLang: "",
	}
}

type entry struct {
	Timestamp string `json:"timestamp"`
	Kind      string `json:"kind"`
	Language  string `json:"language"`
	File      string `json:"file"`
	Snippet   string `json:"snippet"`
}

func LoadStore() (Store, error) {
	return NewJSONLMemStore(), nil
}

func (s *JSONLMemStore) LastRelevant(lang string) string {
	f, err := os.Open(s.Path)
	if err != nil {
		return ""
	}
	defer f.Close()

	var entries []entry
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 512*1024)
	sc.Buffer(buf, 2*1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e entry
		if json.Unmarshal([]byte(line), &e) == nil {
			entries = append(entries, e)
		}
	}
	if len(entries) == 0 {
		return ""
	}

	filtered := filter(entries, lang)
	if len(filtered) == 0 {
		filtered = entries
	}

	if len(filtered) > s.MaxEntries {
		filtered = filtered[len(filtered)-s.MaxEntries:]
	}

	var b strings.Builder
	for _, e := range filtered {
		if b.Len() >= s.GlobalMaxLen {
			break
		}

		ts := e.Timestamp
		if parsed, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
			ts = parsed.UTC().Format(time.RFC3339)
		}

		b.WriteString(fmt.Sprintf("• [%s] (%s) %s\n", ts, e.Language, e.File))
		sn := e.Snippet
		if len(sn) > 800 {
			sn = sn[:800] + "..."
		}
		b.WriteString(sn)
		b.WriteString("\n\n")
	}

	return b.String()
}

func filter(list []entry, lang string) []entry {
	var out []entry
	for _, e := range list {
		if strings.EqualFold(e.Language, lang) {
			out = append(out, e)
		}
	}
	return out
}
