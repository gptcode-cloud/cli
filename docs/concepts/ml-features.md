---
title: ML-Powered Intelligence
description: Embedded machine learning for faster routing and smarter task classification
---

# ML-Powered Intelligence

GPTCode embeds lightweight ML models for instant decision-making with zero external dependencies.

---

## Overview

Two ML models power GPTCode's intelligence:

1. **Complexity Classifier** – Automatically triggers Guided Mode for complex tasks
2. **Intent Classifier** – Routes user requests 500x faster than LLM calls

Both models:
- Run in pure Go (no Python runtime)
- Make predictions in ~1ms
- Cost zero API calls
- Fall back to LLM when uncertain

---

## 1. Complexity Classifier

### Purpose

Analyzes task descriptions and classifies them as:
- **simple** – Single-file edits, typo fixes
- **complex** – Multi-file features, integrations
- **multistep** – Sequential tasks with "then", "after", "first"

### Auto-Activation

When you run `chu chat`, the complexity classifier decides if Guided Mode should activate:

```bash
chu chat "fix typo in readme"
# → Simple task, stays in chat mode

chu chat "implement oauth2 with jwt"
# → Complex task, automatically switches to Guided Mode
```

### Configuration

```bash
# View current threshold (default: 0.55)
chu config get defaults.ml_complex_threshold

# Increase threshold (fewer Guided Mode triggers)
chu config set defaults.ml_complex_threshold 0.7

# Decrease threshold (more Guided Mode triggers)
chu config set defaults.ml_complex_threshold 0.4
```

**Threshold Guide:**
- `0.3-0.4` – Aggressive (lots of Guided Mode)
- `0.5-0.6` – Balanced (default)
- `0.7-0.8` – Conservative (mostly chat)

### Training Data

Located at `ml/complexity_detection/data/`:
- `training_data.csv` – 132 labeled examples
- `eval.csv` – 50 evaluation examples

**Example entries:**
```csv
message,label,notes
"fix typo in readme",0,Simple single file edit
"implement oauth2 authentication",1,Complex multi-file feature
"first add tests then refactor api",2,Explicit sequence
```

Labels: `0=simple`, `1=complex`, `2=multistep`

---

## 2. Intent Classifier

### Purpose

Classifies user intent for routing to specialized agents:
- **query** – Read/understand code ("explain", "show", "list")
- **editor** – Modify code ("add", "fix", "refactor")
- **research** – External info ("how to", "best practices")
- **review** – Code review ("check for bugs", "audit")

### Performance vs LLM

| Metric | ML Classifier | LLM Router |
|--------|---------------|------------|
| Latency | ~1ms | ~500ms |
| Cost | $0 | $0.0001-0.001 |
| Accuracy | 85-90% | 95%+ |
| Fallback | Yes (LLM) | N/A |

### Smart Fallback

If confidence < threshold, falls back to LLM:

```go
// Internal logic
confidence := mlPredict(userMessage)
if confidence >= threshold {
    return mlResult  // Fast path: 1ms
} else {
    return llmCall()  // Slow path: 500ms, more accurate
}
```

### Configuration

```bash
# View current threshold (default: 0.7)
chu config get defaults.ml_intent_threshold

# Higher = more LLM fallbacks (slower but safer)
chu config set defaults.ml_intent_threshold 0.85

# Lower = more ML predictions (faster but riskier)
chu config set defaults.ml_intent_threshold 0.6
```

**Threshold Guide:**
- `0.5-0.6` – Aggressive (95% ML, 5% LLM)
- `0.7-0.8` – Balanced (80% ML, 20% LLM)
- `0.85-0.9` – Conservative (60% ML, 40% LLM)

### Training Data

Located at `ml/intent/data/`:
- `training_data.csv` – 185 labeled examples
- `eval.csv` – 46 evaluation examples

**Example entries:**
```csv
message,label
"explain this code",query
"add error handling",editor
"how to implement oauth",research
"check for bugs",review
```

---

## CLI Commands

### List Models

```bash
chu ml list
```

Shows available models and their status.

### Train Models

```bash
# Train complexity classifier
chu ml train complexity

# Train intent classifier
chu ml train intent
```

Automatically:
1. Creates Python venv
2. Installs dependencies
3. Trains model
4. Exports to JSON

### Test Models

```bash
# Interactive testing
chu ml test complexity
chu ml test intent

# Single prediction
chu ml test complexity "implement oauth"
chu ml test intent "explain this code"
```

### Evaluate Models

```bash
# Use default eval.csv
chu ml eval complexity
chu ml eval intent

# Use custom dataset
chu ml eval intent -f my_eval.csv
```

Shows:
- Overall accuracy
- Per-class precision/recall/F1
- Confusion matrix
- Low-confidence predictions

### Predict (Go Runtime)

```bash
# Default: complexity
chu ml predict "implement oauth"

# Explicit model
chu ml predict complexity "fix typo"
chu ml predict intent "explain this code"
```

Uses embedded Go model – no Python required.

---

## Model Architecture

Both models use the same architecture:

```
Input Text
    ↓
Tokenize + Clean
    ↓
TF-IDF Vectorization (1-3 grams)
    ↓
Logistic Regression
    ↓
Softmax → Probabilities
```

### TF-IDF Features
- **Vocabulary**: Top 1000 most informative terms
- **Ngrams**: Unigrams, bigrams, trigrams
- **Min DF**: 1 (include rare terms)
- **Max DF**: 0.9 (exclude too common terms)

### Classifier
- **Model**: Multinomial Logistic Regression
- **Solver**: L-BFGS
- **Regularization**: C=1.0
- **Class weights**: Balanced

### Export Format

Models export to JSON with:
- TF-IDF vocabulary (term → index mapping)
- IDF weights (term importance)
- Classifier coefficients (one per class)
- Classifier intercepts
- Class labels

Example structure:
```json
{
  "metadata": {
    "version": "1.0.0",
    "model_type": "intent",
    "trained_at": "2025-11-22T12:00:00Z",
    "accuracy": 0.891
  },
  "tfidf": {
    "vocabulary": {"implement": 42, "oauth": 215, ...},
    "idf_weights": [2.3, 1.8, ...]
  },
  "classifier": {
    "coefficients": [[0.5, -0.2, ...], ...],
    "intercepts": [0.1, -0.3, ...],
    "classes": ["query", "editor", "research", "review"]
  }
}
```

---

## Customizing Models

### Add Training Examples

Edit `ml/{model}/data/training_data.csv`:

```csv
message,label
"your new example",appropriate_label
```

Then retrain:
```bash
chu ml train {model}
```

### Adjust Hyperparameters

Edit `ml/{model}/scripts/train.py`:

```python
vectorizer = TfidfVectorizer(
    max_features=1000,     # Vocabulary size
    ngram_range=(1, 3),    # Unigrams to trigrams
    min_df=1,              # Minimum document frequency
    max_df=0.9             # Maximum document frequency
)

clf = LogisticRegression(
    C=1.0,                 # Regularization strength
    max_iter=1000,         # Training iterations
    class_weight='balanced' # Handle class imbalance
)
```

### Export to Go

After training, copy model to assets:

```bash
cp ml/{model}/models/{model}_model.json internal/ml/assets/
```

Then rebuild:
```bash
go build -o bin/chu cmd/chu/main.go
```

---

## Performance Metrics

### Complexity Classifier
- **Accuracy**: 89-92%
- **Training time**: ~1s
- **Model size**: 19KB
- **Inference time**: <1ms

### Intent Classifier
- **Accuracy**: 85-90%
- **Training time**: ~2s
- **Model size**: 66KB
- **Inference time**: <1ms

### ROI Analysis

For 1000 requests/day:

**Without ML (LLM only):**
- Latency: 500ms × 1000 = 8.3 minutes
- Cost: $0.0005 × 1000 = $0.50/day = $15/month

**With ML (80% ML, 20% LLM):**
- Latency: (1ms × 800) + (500ms × 200) = 100.8s = 1.7 minutes
- Cost: ($0 × 800) + ($0.0005 × 200) = $0.10/day = $3/month

**Savings:**
- 83% faster (8.3min → 1.7min)
- 80% cheaper ($15 → $3)

---

## Troubleshooting

### Model not found

```bash
# Train the model first
chu ml train complexity
chu ml train intent
```

### Python dependencies

Models require Python 3.8+ with:
- pandas
- scikit-learn
- numpy

Install manually:
```bash
cd ml/{model}
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Low accuracy

1. Add more training examples
2. Balance classes (equal examples per label)
3. Increase `max_features` in TF-IDF
4. Adjust regularization `C` parameter

### Embedded model outdated

After retraining, copy new model:
```bash
cp ml/{model}/models/{model}_model.json internal/ml/assets/
go build ./cmd/chu
```

---

## Next Steps

- Explore [Commands Reference](./commands.html) for full CLI
- See [Configuration](./index.html#configuration) for setup
- Read training scripts in `ml/{model}/scripts/`
