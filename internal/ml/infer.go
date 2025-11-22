package ml

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

type modelData struct {
	TFIDF struct {
		Vocabulary map[string]int `json:"vocabulary"`
		IDF        []float64      `json:"idf_weights"`
	} `json:"tfidf"`
	Classifier struct {
		Coefficients [][]float64 `json:"coefficients"`
		Intercepts   []float64   `json:"intercepts"`
		Classes      []string    `json:"classes"`
	} `json:"classifier"`
}

type Predictor struct {
	modelName  string
	vocab      map[string]int
	idf        []float64
	coefs      [][]float64
	intercepts []float64
	classes    []string
}

func LoadEmbedded(modelName string) (*Predictor, error) {
	data, ok := embeddedModels[modelName]
	if !ok {
		return nil, fmt.Errorf("model '%s' not found in embedded assets", modelName)
	}

	var m modelData
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model '%s': %w", modelName, err)
	}
	return &Predictor{
		modelName:  modelName,
		vocab:      m.TFIDF.Vocabulary,
		idf:        m.TFIDF.IDF,
		coefs:      m.Classifier.Coefficients,
		intercepts: m.Classifier.Intercepts,
		classes:    m.Classifier.Classes,
	}, nil
}

var ws = regexp.MustCompile(`\s+`)
var nonWord = regexp.MustCompile(`[^a-z0-9]+`)

func tokenize(s string) []string {
	s = strings.ToLower(s)
	s = nonWord.ReplaceAllString(s, " ")
	parts := ws.Split(strings.TrimSpace(s), -1)
	if len(parts) == 1 && parts[0] == "" {
		return []string{}
	}
	return parts
}

func ngrams(tokens []string, n int) []string {
	if n <= 1 {
		return append([]string{}, tokens...)
	}
	out := make([]string, 0)
	for i := 0; i <= len(tokens)-n; i++ {
		out = append(out, strings.Join(tokens[i:i+n], " "))
	}
	return out
}

func (p *Predictor) vectorize(text string) []float64 {
	toks := tokenize(text)
	uni := ngrams(toks, 1)
	bi := ngrams(toks, 2)
	tri := ngrams(toks, 3)
	all := append(append(uni, bi...), tri...)
	counts := map[int]float64{}
	for _, t := range all {
		if idx, ok := p.vocab[t]; ok {
			counts[idx] += 1
		}
	}
	v := make([]float64, len(p.idf))
	var norm float64
	for idx, tf := range counts {
		val := tf * p.idf[idx]
		v[idx] = val
		norm += val * val
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range v {
			v[i] /= norm
		}
	}
	return v
}

func softmax(x []float64) []float64 {
	max := x[0]
	for i := 1; i < len(x); i++ {
		if x[i] > max {
			max = x[i]
		}
	}
	exp := make([]float64, len(x))
	var sum float64
	for i := range x {
		e := math.Exp(x[i] - max)
		exp[i] = e
		sum += e
	}
	for i := range exp {
		exp[i] /= sum
	}
	return exp
}

func (p *Predictor) Predict(text string) (string, map[string]float64) {
	x := p.vectorize(text)
	logits := make([]float64, len(p.coefs))
	for c := range p.coefs {
		sum := p.intercepts[c]
		row := p.coefs[c]
		for j := 0; j < len(x) && j < len(row); j++ {
			sum += row[j] * x[j]
		}
		logits[c] = sum
	}

	// Apply complexity-specific heuristics
	if p.modelName == "complexity_detection" {
		lower := " " + strings.ToLower(text) + " "
		idxComplex := indexOf(p.classes, "complex")
		idxMulti := indexOf(p.classes, "multistep")
		multiCues := []string{" then ", " and then ", ", then ", "; then ", " after ", " followed by ", " first ", " second ", " third "}
		complexCues := []string{"oauth", "oauth2", "oidc", "migrate", "migration", "deploy", "docker", "kubectl", "k8s", "s3", "pipeline", "airflow", "terraform", "ansible", "kafka", "stripe", "payment", "upload", "script", "bash", "nginx", "logs"}
		if anyContains(lower, multiCues) && idxMulti >= 0 {
			logits[idxMulti] += 1.0
		}
		if anyContains(lower, complexCues) && idxComplex >= 0 {
			logits[idxComplex] += 1.5
		}
	}

	probs := softmax(logits)
	best := 0
	for i := 1; i < len(probs); i++ {
		if probs[i] > probs[best] {
			best = i
		}
	}
	out := map[string]float64{}
	for i, c := range p.classes {
		out[c] = probs[i]
	}
	return p.classes[best], out
}

func anyContains(s string, cues []string) bool {
	for _, c := range cues {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}

func indexOf(arr []string, val string) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}

func SortedProbs(m map[string]float64) [][2]string {
	t := make([][2]string, 0, len(m))
	for k, v := range m {
		t = append(t, [2]string{k, formatFloat(v)})
	}
	sort.Slice(t, func(i, j int) bool { return t[i][1] > t[j][1] })
	return t
}

func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(sprintf("%.4f", f), "0"), ".")
}

func sprintf(format string, a ...any) string { return fmtSprintf(format, a...) }

var fmtSprintf = func(format string, a ...any) string { return fmt.Sprintf(format, a...) }
