---
layout: post
title: "ML-Powered Intelligence: 500x Faster, 80% Cheaper"
date: 2025-11-22
author: Jader Corrêa
tags: [machine-learning, performance, cost-optimization]
---

# ML-Powered Intelligence: 500x Faster, 80% Cheaper

*November 22, 2025*

Today we're announcing embedded machine learning in Chuchu. Two lightweight ML models now power instant decision-making with zero external dependencies and zero API costs.

## The Problem with LLM-Only Routing

Every time you interact with an AI coding assistant, there's a hidden cost:

**Intent classification** - deciding whether you want to read code, edit it, or search documentation - requires an LLM call:
- **Latency**: ~500ms per request
- **Cost**: $0.0005 per classification
- **Scales poorly**: 1000 requests/day = $15/month just for routing

For simple decisions like "is this a query or an edit?", calling a 70B parameter model is overkill.

## The Solution: Hybrid ML + LLM

Chuchu now embeds two ML models that run locally in pure Go:

### 1. Intent Classifier

Routes user requests to the right agent (query, editor, research, review):

| Metric | ML Classifier | LLM Router |
|--------|---------------|------------|
| **Latency** | ~1ms | ~500ms |
| **Cost** | $0 | $0.0005 |
| **Accuracy** | 85-90% | 95%+ |

**Smart fallback**: When the ML model is uncertain (confidence < threshold), it falls back to the LLM. You get speed when possible, accuracy when needed.

### 2. Complexity Classifier

Auto-detects if a task is simple or complex:

```bash
chu chat "fix typo in readme"
# → Simple task, stays in chat mode

chu chat "implement oauth2 with jwt"
# → Complex task, automatically activates Guided Mode
```

No more manual mode switching. The system adapts to task complexity automatically.

## Real-World Impact

For **1000 requests per day** (typical heavy usage):

### Without ML (LLM only)
- Latency: 500ms × 1000 = **8.3 minutes of waiting**
- Cost: $0.0005 × 1000 = **$15/month** just for routing

### With ML (80% ML, 20% LLM fallback)
- Latency: (1ms × 800) + (500ms × 200) = **1.7 minutes**
- Cost: ($0 × 800) + ($0.0005 × 200) = **$3/month**

**Result:**
- **83% faster** (8.3min → 1.7min)
- **80% cheaper** ($15 → $3)
- **Same accuracy** (smart fallback maintains quality)

## How It Works

Both models use the same architecture:

```
User Input
    ↓
TF-IDF Vectorization (1-3 grams)
    ↓
Logistic Regression
    ↓
Confidence Score
    ↓
High confidence → ML result (1ms)
Low confidence → LLM fallback (500ms)
```

**Model specs:**
- **Size**: 19-66KB (tiny!)
- **Runtime**: Pure Go, zero Python dependencies
- **Training**: scikit-learn, exports to JSON
- **Inference**: <1ms per prediction

## Configuration

Both models have configurable confidence thresholds:

```bash
# Intent classification (default: 0.7)
chu config get defaults.ml_intent_threshold
chu config set defaults.ml_intent_threshold 0.8  # More conservative

# Complexity detection (default: 0.55)
chu config get defaults.ml_complex_threshold
chu config set defaults.ml_complex_threshold 0.6  # Less Guided Mode
```

**Lower threshold** = more ML, faster but riskier
**Higher threshold** = more LLM fallback, slower but safer

## CLI Commands

### Test the models interactively

```bash
# Test intent classification
chu ml test intent
> explain this code
Prediction: query (confidence: 0.89)

# Test complexity detection
chu ml test complexity
> implement oauth2 authentication
Prediction: complex (confidence: 0.91) → Would trigger Guided Mode
```

### Evaluate accuracy

```bash
chu ml eval intent
Overall Accuracy: 89.1%

chu ml eval complexity
Overall Accuracy: 92.3%
```

### Train your own models

```bash
# Retrain with your own examples
chu ml train intent
chu ml train complexity
```

Training data is in `ml/{model}/data/training_data.csv` - add your own examples and retrain to customize the models for your workflow.

## Cost per 100M Tokens

Routing costs scale with usage. For 100M tokens processed:

**Without ML (LLM routing):**
- Tokens: 100M tokens ÷ 75 tokens/request = 1.33M routing calls
- Cost: 1.33M × $0.0005 = **$665**

**With ML (80% ML, 20% LLM):**
- ML calls: 1.06M requests = $0
- LLM fallback: 266k requests × $0.0005 = **$133**
- **Savings: $532** (80% reduction)

## Impact on Free Tier Profiles

Many OpenRouter models have rate limits (free tier). ML routing helps you stay under limits:

**Example: Grok 4.1 Fast (free tier)**
- Limit: ~100 requests/minute
- Without ML: Router uses 100 req/min = **hitting limit**
- With ML: Router uses 20 req/min = **80% capacity freed**

This means you can handle 5x more chat interactions before hitting rate limits on free models.

## Why This Matters

**Speed**: Fast routing means the assistant feels instant. No more waiting 500ms just to figure out what you want to do.

**Cost**: Routing happens on every interaction. At scale, ML routing saves real money without sacrificing quality.

**Privacy**: ML models run locally. Your task descriptions never leave your machine for simple routing decisions.

**Offline**: No internet? No problem. ML models work completely offline.

**Customizable**: Don't like the default behavior? Adjust thresholds or retrain with your own data.

## Real Usage Example

```bash
$ chu chat

> show me the auth code
# ML: "show" → query agent (1ms)
# Reads relevant files, explains auth flow

> add rate limiting to login endpoint  
# ML: "add" → editor agent (1ms)
# Opens editor, implements rate limiting

> how do other projects handle rate limiting?
# ML: uncertain (confidence: 0.65 < 0.7)
# Fallback: LLM router → research agent (500ms)
# Searches web, summarizes best practices

> fix the test that's failing
# ML: "fix" + "test" → editor agent (1ms)
# Runs tests, identifies issue, fixes it
```

**Result**: 3 out of 4 requests used ML (1ms), only 1 used LLM (500ms). Total routing time: **503ms** instead of **2000ms**.

## The Hybrid Advantage

Pure ML would be faster but less accurate.
Pure LLM would be more accurate but slower and expensive.

**Hybrid ML + LLM** gives you the best of both worlds:
- Fast path for confident decisions (80-90% of requests)
- Smart fallback for edge cases
- Configurable balance between speed and accuracy

## Technical Details

For implementation details, training data format, and hyperparameter tuning, see our [ML Features documentation](../ml-features).

For performance benchmarks and ROI analysis, check the full metrics in the docs.

## Try It Now

The ML models are embedded in Chuchu by default - no setup required. They "just work."

Want to see them in action?

```bash
# Interactive testing
chu ml test intent
chu ml test complexity

# Check your current config
chu config get defaults.ml_intent_threshold
chu config get defaults.ml_complex_threshold

# View training data
cat ml/intent/data/training_data.csv
cat ml/complexity_detection/data/training_data.csv
```

## What's Next

We're exploring:
- **Confidence calibration**: Even better fallback decisions
- **Active learning**: Learn from your corrections
- **Code-specific features**: Leverage AST patterns, not just text
- **Personalization**: Models that adapt to your coding style

But the foundation is here today: fast, cheap, accurate routing powered by embedded ML.

---

## Summary

**500x faster**: 1ms vs 500ms routing latency

**80% cheaper**: $3 vs $15/month for typical usage

**Zero deps runtime**: Inference in pure Go

**Smart fallback**: LLM when uncertain

**Savings at scale**: $532 per 100M tokens

**Rate-limit friendly**: 5x more requests on free tier profiles

---

*Have questions about the ML system? Check out the [full documentation](../ml-features) or ask in [GitHub Discussions](https://github.com/jadercorrea/chuchu/discussions)!*

## See Also

- [Full ML Features Documentation](../ml-features) - Technical deep dive
- [Groq Optimal Configurations](2024-11-18-groq-optimal-configs) - Combine ML routing with cheap Groq models
- [Cost Optimization](../commands#configuration) - Additional cost-saving strategies
