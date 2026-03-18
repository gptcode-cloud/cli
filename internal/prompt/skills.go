package prompt

import (
	"strings"
)

// SkillsLoader loads language-specific and product skills
type SkillsLoader struct {
	// We use embedded skills by default
}

// NewSkillsLoader creates a loader that uses embedded skills
func NewSkillsLoader() *SkillsLoader {
	return &SkillsLoader{}
}

// languageSkillMap maps language names to skill files
var languageSkillMap = map[string]string{
	"go":         "go.md",
	"golang":     "go.md",
	"elixir":     "elixir.md",
	"ex":         "elixir.md",
	"ruby":       "ruby.md",
	"rb":         "ruby.md",
	"rails":      "rails.md",
	"python":     "python.md",
	"py":         "python.md",
	"typescript": "typescript.md",
	"ts":         "typescript.md",
	"javascript": "javascript.md",
	"js":         "javascript.md",
	"rust":       "rust.md",
	"rs":         "rust.md",
}

// productSkillKeywords maps task keywords to product skill files
var productSkillKeywords = map[string]string{
	// Design System
	"component": "design-system.md",
	"ui":        "design-system.md",
	"frontend":  "design-system.md",
	"button":    "design-system.md",
	"form":      "design-system.md",
	"modal":     "design-system.md",
	"storybook": "design-system.md",
	"design":    "design-system.md",
	"css":       "design-system.md",
	"styled":    "design-system.md",

	// Product Metrics
	"tracking":   "product-metrics.md",
	"analytics":  "product-metrics.md",
	"metrics":    "product-metrics.md",
	"pixel":      "product-metrics.md",
	"gtag":       "product-metrics.md",
	"ga4":        "product-metrics.md",
	"utm":        "product-metrics.md",
	"conversion": "product-metrics.md",
	"funnel":     "product-metrics.md",

	// Production Ready
	"production":    "production-ready.md",
	"deploy":        "production-ready.md",
	"error":         "production-ready.md",
	"health":        "production-ready.md",
	"feature flag":  "production-ready.md",
	"circuit":       "production-ready.md",
	"retry":         "production-ready.md",
	"timeout":       "production-ready.md",
	"logging":       "production-ready.md",
	"observability": "production-ready.md",

	// QA Automation
	"test":          "qa-automation.md",
	"e2e":           "qa-automation.md",
	"playwright":    "qa-automation.md",
	"cypress":       "qa-automation.md",
	"selenium":      "qa-automation.md",
	"a11y":          "qa-automation.md",
	"accessibility": "qa-automation.md",
	"visual":        "qa-automation.md",
	"screenshot":    "qa-automation.md",

	// Security
	"security":       "security.md",
	"auth":           "security.md",
	"authentication": "security.md",
	"authorization":  "security.md",
	"owasp":          "security.md",
	"xss":            "security.md",
	"csrf":           "security.md",
	"injection":      "security.md",
	"sanitize":       "security.md",
	"validate":       "security.md",
	"password":       "security.md",
	"token":          "security.md",
	"jwt":            "security.md",

	// DevOps
	"devops":         "devops.md",
	"docker":         "devops.md",
	"dockerfile":     "devops.md",
	"kubernetes":     "devops.md",
	"k8s":            "devops.md",
	"helm":           "devops.md",
	"terraform":      "devops.md",
	"ci/cd":          "devops.md",
	"pipeline":       "devops.md",
	"github actions": "devops.md",

	// SysOps
	"sysops":   "sysops.md",
	"bash":     "sysops.md",
	"shell":    "sysops.md",
	"linux":    "sysops.md",
	"systemd":  "sysops.md",
	"cron":     "sysops.md",
	"ssh":      "sysops.md",
	"firewall": "sysops.md",
	"nginx":    "sysops.md",

	// SecOps
	"secops":        "secops.md",
	"vulnerability": "secops.md",
	"incident":      "secops.md",
	"waf":           "secops.md",
	"siem":          "secops.md",
	"audit":         "secops.md",
	"compliance":    "secops.md",

	// MLOps
	"mlops":         "mlops.md",
	"ml":            "mlops.md",
	"model":         "mlops.md",
	"training":      "mlops.md",
	"inference":     "mlops.md",
	"mlflow":        "mlops.md",
	"experiment":    "mlops.md",
	"feature store": "mlops.md",
}

// workflowSkillMap maps workflow types to skill files
var workflowSkillMap = map[string]string{
	"tdd":    "tdd-bug-fix.md",
	"bug":    "tdd-bug-fix.md",
	"fix":    "tdd-bug-fix.md",
	"review": "code-review.md",
	"commit": "git-commit.md",
	"git":    "git-commit.md",
}

// loadEmbeddedSkill loads a skill from embedded files
func loadEmbeddedSkill(fileName string) string {
	content, err := embeddedSkills.ReadFile("skills/" + fileName)
	if err != nil {
		return ""
	}
	return string(content)
}

// LoadForLanguage returns the skill content for a given language
func (sl *SkillsLoader) LoadForLanguage(lang string) string {
	langLower := strings.ToLower(lang)

	fileName, ok := languageSkillMap[langLower]
	if !ok {
		return ""
	}

	return loadEmbeddedSkill(fileName)
}

// LoadByName returns a specific skill by name (e.g., "tdd-bug-fix", "code-review")
func (sl *SkillsLoader) LoadByName(name string) string {
	// Try with .md extension
	content := loadEmbeddedSkill(name + ".md")
	if content != "" {
		return content
	}

	// Try as-is (shouldn't happen but just in case)
	return loadEmbeddedSkill(name)
}

// LoadProductSkillsForTask analyzes task description and returns relevant product skills
func (sl *SkillsLoader) LoadProductSkillsForTask(task string) []string {
	if task == "" {
		return nil
	}

	taskLower := strings.ToLower(task)

	// Track which skills we've already added to avoid duplicates
	addedSkills := make(map[string]bool)
	var skills []string

	// Check for product skill keywords
	for keyword, fileName := range productSkillKeywords {
		if strings.Contains(taskLower, keyword) && !addedSkills[fileName] {
			content := loadEmbeddedSkill(fileName)
			if content != "" {
				skills = append(skills, content)
				addedSkills[fileName] = true
			}
		}
	}

	// Check for workflow skill keywords
	for keyword, fileName := range workflowSkillMap {
		if strings.Contains(taskLower, keyword) && !addedSkills[fileName] {
			content := loadEmbeddedSkill(fileName)
			if content != "" {
				skills = append(skills, content)
				addedSkills[fileName] = true
			}
		}
	}

	return skills
}

// ListAvailable returns the names of all available skills
func (sl *SkillsLoader) ListAvailable() []string {
	entries, err := embeddedSkills.ReadDir("skills")
	if err != nil {
		return nil
	}

	var skills []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			// Remove .md extension for display
			name := strings.TrimSuffix(entry.Name(), ".md")
			skills = append(skills, name)
		}
	}

	return skills
}

// GetSkillCategories returns skills organized by category
func (sl *SkillsLoader) GetSkillCategories() map[string][]string {
	return map[string][]string{
		"language": {
			"go", "elixir", "ruby", "rails", "python",
			"typescript", "javascript", "rust",
		},
		"workflow": {
			"tdd-bug-fix", "code-review", "git-commit",
		},
		"product": {
			"design-system", "product-metrics",
			"production-ready", "qa-automation",
		},
		"ops": {
			"security", "devops", "sysops", "secops", "mlops",
		},
	}
}
