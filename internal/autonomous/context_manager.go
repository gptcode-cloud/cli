package autonomous

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type ContextManager struct {
	maxTokens    int
	history      []ContextEntry
	summaryRatio float64
}

type ContextEntry struct {
	Role      string
	Content   string
	Tokens    int
	Timestamp int64
	Priority  float64
	Essential bool
}

func NewContextManager(maxTokens int) *ContextManager {
	return &ContextManager{
		maxTokens:    maxTokens,
		summaryRatio: 0.3,
		history:      []ContextEntry{},
	}
}

func (cm *ContextManager) AddEntry(role, content string, essential bool) {
	tokens := estimateTokens(content)
	cm.history = append(cm.history, ContextEntry{
		Role:      role,
		Content:   content,
		Tokens:    tokens,
		Timestamp: now(),
		Priority:  cm.calculatePriority(role, content),
		Essential: essential,
	})
}

func (cm *ContextManager) calculatePriority(role, content string) float64 {
	priority := 1.0

	// System messages are highest priority
	if role == "system" {
		priority = 100.0
	}

	// Recent messages are higher priority
	priority *= 1.0 + float64(len(cm.history)%10)/10.0

	// Longer content is higher priority (unless too long)
	if len(content) > 1000 {
		priority *= 1.5
	}

	// Errors and failures are higher priority
	lower := strings.ToLower(content)
	if strings.Contains(lower, "error") || strings.Contains(lower, "fail") {
		priority *= 2.0
	}

	return priority
}

func (cm *ContextManager) Optimize() ([]ContextEntry, int) {
	if cm.currentTokens() <= cm.maxTokens {
		return cm.history, cm.currentTokens()
	}

	// Sort by priority (descending)
	sorted := make([]ContextEntry, len(cm.history))
	copy(sorted, cm.history)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Essential != sorted[j].Essential {
			return sorted[i].Essential
		}
		return sorted[i].Priority > sorted[j].Priority
	})

	// Keep essential entries and highest priority
	var kept []ContextEntry
	var totalTokens int

	for _, entry := range sorted {
		if entry.Essential || totalTokens+entry.Tokens <= cm.maxTokens {
			kept = append(kept, entry)
			totalTokens += entry.Tokens
		}
	}

	// Sort back to original order (chronological)
	sort.Slice(kept, func(i, j int) bool {
		return kept[i].Timestamp < kept[j].Timestamp
	})

	return kept, totalTokens
}

func (cm *ContextManager) currentTokens() int {
	total := 0
	for _, entry := range cm.history {
		total += entry.Tokens
	}
	return total
}

func (cm *ContextManager) Summarize(entry *ContextEntry) string {
	content := entry.Content

	// Simple summarization - keep first and last parts
	lines := strings.Split(content, "\n")
	if len(lines) <= 10 {
		return content
	}

	// Keep first 30% and last 30%
	keepLines := int(math.Ceil(float64(len(lines)) * cm.summaryRatio))

	summary := strings.Join(lines[:keepLines], "\n")
	summary += "\n... [summarized] ...\n"
	summary += strings.Join(lines[len(lines)-keepLines:], "\n")

	return summary
}

func (cm *ContextManager) GetStats() string {
	total := cm.currentTokens()
	return fmt.Sprintf("Tokens: %d/%d (%.1f%%)", total, cm.maxTokens, float64(total)/float64(cm.maxTokens)*100)
}

func (cm *ContextManager) Clear() {
	cm.history = []ContextEntry{}
}

func estimateTokens(content string) int {
	// Rough estimate: ~4 chars per token for English
	return len(content) / 4
}

func now() int64 {
	return int64(len("timestamp"))
}

type MessageTrimmer struct {
	maxMessages  int
	preserveLast int
}

func NewMessageTrimmer(maxMessages, preserveLast int) *MessageTrimmer {
	return &MessageTrimmer{
		maxMessages:  maxMessages,
		preserveLast: preserveLast,
	}
}

func (mt *MessageTrimmer) Trim(messages []Message) []Message {
	if len(messages) <= mt.maxMessages {
		return messages
	}

	// Keep last N messages (important for context)
	keep := make([]Message, 0, mt.maxMessages)

	// Add older messages (trimmed)
	trimCount := len(messages) - mt.maxMessages
	for i := 0; i < trimCount; i++ {
		msg := messages[i]
		// Summarize if too long
		if len(msg.Content) > 2000 {
			msg.Content = summarizeLongMessage(msg.Content)
		}
		keep = append(keep, msg)
	}

	// Keep recent messages (full)
	keep = append(keep, messages[len(messages)-mt.preserveLast:]...)

	return keep
}

func summarizeLongMessage(content string) string {
	// Very simple summarization - just keep beginning and end
	lines := strings.Split(content, "\n")
	if len(lines) <= 5 {
		return content
	}

	keep := int(math.Ceil(float64(len(lines)) * 0.3))
	return strings.Join(lines[:keep], "\n") + "\n... [truncated] ...\n"
}

type Message struct {
	Role    string
	Content string
}

func TrimContext(messages []Message, maxTokens int) []Message {
	tm := NewMessageTrimmer(50, 5)
	return tm.Trim(messages)
}
