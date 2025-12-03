package langdetect

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// LanguageBreakdown represents language distribution in a project
type LanguageBreakdown struct {
	Languages map[string]float64 // language -> percentage
	Primary   string             // dominant language
	Context   string             // "pure_code", "polyglot_balanced", etc
}

// Detector uses go-enry (GitHub Linguist) for language detection
type Detector struct {
	path      string
	mlContext *ContextPredictor
}

// NewDetector creates a new language detector for a project path
func NewDetector(path string) *Detector {
	// Load ML model (fallback to rules if load fails)
	mlContext, _ := NewContextPredictor()
	return &Detector{
		path:      path,
		mlContext: mlContext,
	}
}

// Detect analyzes the project and returns language breakdown
func (d *Detector) Detect() (*LanguageBreakdown, error) {
	langBytes := make(map[string]int64)
	totalBytes := int64(0)

	// Convert to absolute path
	absPath, err := filepath.Abs(d.path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}

		if info.IsDir() {
			// Skip common directories to ignore
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" ||
				name == ".venv" || name == "venv" || name == "dist" || name == "build" ||
				name == "bin" {
				return filepath.SkipDir
			}
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}

		// Determine language
		lang := enry.GetLanguage(path, content)
		if lang == "" {
			return nil
		}

		// Skip vendored, generated, and documentation files
		// But allow markdown/docs for documentation projects
		if enry.IsVendor(path) || enry.IsGenerated(path, content) {
			return nil
		}

		// Accumulate bytes per language
		langBytes[lang] += int64(len(content))
		totalBytes += int64(len(content))

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	if totalBytes == 0 {
		return &LanguageBreakdown{
			Languages: map[string]float64{},
			Primary:   "unknown",
			Context:   "unknown",
		}, nil
	}

	// Convert to percentages
	languages := make(map[string]float64)
	for lang, bytes := range langBytes {
		languages[lang] = float64(bytes) / float64(totalBytes)
	}

	// Find primary language
	primary := d.findPrimary(languages)

	// Create temporary breakdown for ML
	tempBreakdown := &LanguageBreakdown{
		Languages: languages,
		Primary:   primary,
	}

	// Determine context using ML or fallback to rules
	var context string
	if d.mlContext != nil {
		context = d.mlContext.Predict(tempBreakdown)
	} else {
		context = d.determineContext(languages)
	}

	return &LanguageBreakdown{
		Languages: languages,
		Primary:   primary,
		Context:   context,
	}, nil
}

func (d *Detector) findPrimary(languages map[string]float64) string {
	var primary string
	var maxPct float64

	for lang, pct := range languages {
		if pct > maxPct {
			maxPct = pct
			primary = lang
		}
	}

	if primary == "" {
		return "unknown"
	}
	return primary
}

func (d *Detector) determineContext(languages map[string]float64) string {
	// Count significant languages (>= 10%)
	significant := []string{}
	for lang, pct := range languages {
		if pct >= 0.10 {
			significant = append(significant, lang)
		}
	}

	// Get primary language ratio
	primary := d.findPrimary(languages)
	primaryRatio := languages[primary]

	// Determine context
	if len(significant) == 1 {
		if primaryRatio > 0.80 {
			return "pure_code"
		}
		return "scripted" // has small supporting files
	}

	if len(significant) >= 3 {
		return "polyglot_balanced"
	}

	if len(significant) == 2 {
		// Check if secondary is scripting language
		langs := []string{}
		for _, lang := range significant {
			langs = append(langs, strings.ToLower(lang))
		}
		for _, lang := range langs {
			if lang == "shell" || lang == "bash" || lang == "makefile" || lang == "dockerfile" {
				return "polyglot_scripted"
			}
		}
		return "polyglot_balanced"
	}

	return "mixed"
}

// FormatBreakdown returns a formatted string representation
func FormatBreakdown(breakdown *LanguageBreakdown) string {
	if len(breakdown.Languages) == 0 {
		return "No languages detected"
	}

	// Sort by percentage descending
	type langPct struct {
		lang string
		pct  float64
	}
	sorted := []langPct{}
	for lang, pct := range breakdown.Languages {
		sorted = append(sorted, langPct{lang, pct})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].pct > sorted[j].pct
	})

	var b strings.Builder
	b.WriteString("Language Breakdown:\n")
	for _, item := range sorted {
		pct := item.pct * 100
		bar := strings.Repeat("█", int(pct/5))
		b.WriteString(fmt.Sprintf("  %-15s %5.1f%%  %s\n", item.lang, pct, bar))
	}
	b.WriteString(fmt.Sprintf("\nProject Context: %s\n", breakdown.Context))

	return b.String()
}
