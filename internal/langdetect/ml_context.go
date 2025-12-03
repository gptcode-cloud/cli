package langdetect

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed ml_context_model.json
var contextModelJSON []byte

// RandomForestModel represents a trained Random Forest classifier
type RandomForestModel struct {
	Type         string     `json:"type"`
	NEstimators  int        `json:"n_estimators"`
	Classes      []string   `json:"classes"`
	FeatureNames []string   `json:"feature_names"`
	Trees        []TreeNode `json:"trees"`
}

// TreeNode represents a decision tree in the forest
type TreeNode struct {
	Feature       []int         `json:"feature"`
	Threshold     []float64     `json:"threshold"`
	Value         [][][]float64 `json:"value"`
	ChildrenLeft  []int         `json:"children_left"`
	ChildrenRight []int         `json:"children_right"`
}

// ContextPredictor uses ML to predict project context
type ContextPredictor struct {
	model *RandomForestModel
}

// NewContextPredictor loads the embedded ML model
func NewContextPredictor() (*ContextPredictor, error) {
	var model RandomForestModel
	if err := json.Unmarshal(contextModelJSON, &model); err != nil {
		return nil, fmt.Errorf("failed to load context model: %w", err)
	}
	
	return &ContextPredictor{model: &model}, nil
}

// Predict classifies project context based on language breakdown
func (p *ContextPredictor) Predict(breakdown *LanguageBreakdown) string {
	features := p.extractFeatures(breakdown)
	
	// Get predictions from all trees
	votes := make(map[string]int)
	for _, tree := range p.model.Trees {
		class := p.predictTree(&tree, features)
		votes[class]++
	}
	
	// Return class with most votes
	var bestClass string
	var maxVotes int
	for class, count := range votes {
		if count > maxVotes {
			maxVotes = count
			bestClass = class
		}
	}
	
	return bestClass
}

func (p *ContextPredictor) extractFeatures(breakdown *LanguageBreakdown) []float64 {
	// Count languages with >1%
	langCount := 0
	for _, pct := range breakdown.Languages {
		if pct > 0.01 {
			langCount++
		}
	}
	
	// Get primary and secondary ratios
	type langPct struct {
		lang string
		pct  float64
	}
	sorted := []langPct{}
	for lang, pct := range breakdown.Languages {
		sorted = append(sorted, langPct{lang, pct})
	}
	
	// Sort by percentage descending
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].pct > sorted[i].pct {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	primaryRatio := 0.0
	secondaryRatio := 0.0
	if len(sorted) > 0 {
		primaryRatio = sorted[0].pct
	}
	if len(sorted) > 1 {
		secondaryRatio = sorted[1].pct
	}
	
	// Check for specific language types
	hasDocs := p.hasLanguageType(breakdown, []string{"markdown", "rst", "asciidoc"})
	hasTests := p.hasLanguageType(breakdown, []string{"test"})
	hasScripts := p.hasLanguageType(breakdown, []string{"shell", "bash", "makefile", "powershell"})
	hasInfra := p.hasLanguageType(breakdown, []string{"dockerfile", "terraform", "hcl"})
	hasData := p.hasLanguageType(breakdown, []string{"csv", "json", "yaml", "toml", "xml"})
	
	return []float64{
		float64(langCount),
		primaryRatio,
		secondaryRatio,
		boolToFloat(hasDocs),
		boolToFloat(hasTests),
		boolToFloat(hasScripts),
		boolToFloat(hasInfra),
		boolToFloat(hasData),
	}
}

func (p *ContextPredictor) hasLanguageType(breakdown *LanguageBreakdown, keywords []string) bool {
	for lang := range breakdown.Languages {
		lowerLang := strings.ToLower(lang)
		for _, keyword := range keywords {
			if strings.Contains(lowerLang, keyword) {
				return true
			}
		}
	}
	return false
}

func (p *ContextPredictor) predictTree(tree *TreeNode, features []float64) string {
	nodeIdx := 0
	
	for {
		// Check if leaf node
		if tree.ChildrenLeft[nodeIdx] == -1 {
			// Get class with highest value
			values := tree.Value[nodeIdx][0]
			maxIdx := 0
			maxVal := values[0]
			for i, val := range values {
				if val > maxVal {
					maxVal = val
					maxIdx = i
				}
			}
			return p.model.Classes[maxIdx]
		}
		
		// Internal node - follow left or right
		featureIdx := tree.Feature[nodeIdx]
		threshold := tree.Threshold[nodeIdx]
		
		if features[featureIdx] <= threshold {
			nodeIdx = tree.ChildrenLeft[nodeIdx]
		} else {
			nodeIdx = tree.ChildrenRight[nodeIdx]
		}
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
