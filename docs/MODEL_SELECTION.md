# Intelligent Model Selection

## Overview
Chuchu uses a multi-dimensional scoring system to automatically select the best LLM for each task based on:

1. **Availability** - Rate limits & health status
2. **Cost** - $/1M tokens (prioritizes free models)
3. **Context Window** - Larger is better for complex tasks
4. **Speed** - tokens/sec for responsiveness

## Scoring Formula

```go
score = 100
  - (requests_today / rate_limit) * 50
  - (usage.LastError != "") * 30
  - (cost_per_1m / 10.0) * 30
  + (context_window / 100000.0) * 10
  + (tokens_per_sec / 100.0) * 5
  + feedback_bonus/penalty
  + complexity_bonus
  + code_specialization_bonus
```

## Default Prioritization

**Mode: cloud (default)**
- OpenRouter free models → Groq → Ollama
- Gemini 2.0 Flash free (1M context, 150 tok/s)
- Llama 3.2 3B free (128k context, 200 tok/s)
- Qwen 2.5 Coder 32B ($0.3/1M, specialized)

**Mode: local**
- Ollama only
- qwen2.5-coder:32b (free, local)
- llama3.2:3b (lightweight)

## Usage Tracking

`~/.chuchu/usage.json`:
```json
{
  "2025-12-02": {
    "openrouter/gemini-2.0-flash-exp:free": {
      "requests": 847,
      "last_error": null
    }
  }
}
```

When a model hits 90% of daily rate limit, score drops by 50 points.

## Feedback Learning

Historical success/failure tracked in `~/.chuchu/feedback/`:
- Success: +20 score
- Failure: -40 score
- ML training after 20-30 tasks

## Model Catalog

`~/.chuchu/models_catalog.json` contains:
```json
{
  "openrouter": {
    "models": [{
      "id": "google/gemini-2.0-flash-exp:free",
      "cost_per_1m": 0,
      "rate_limit_daily": 1000,
      "context_window": 1000000,
      "tokens_per_sec": 150,
      "capabilities": {
        "supports_tools": true,
        "supports_file_operations": true
      }
    }]
  }
}
```

## Adding New Models

Run enrichment script:
```bash
python3 ml/scripts/enrich_catalog.py
```

Or edit `models_catalog.json` directly with cost/limits/context/speed.

## Monitoring

**View usage statistics**:
```bash
python3 ml/scripts/show_usage.py
```

Output:
```
📊 Model Usage Statistics

🔥 2025-12-02
  ✓ openrouter/gemini-2.0-flash-exp:free: 47 requests
  ✓ groq/llama-3.3-70b-versatile: 12 requests
  ❌ groq/compound: 3 requests
     └─ Last error: Provider error...
```

**Debug mode**:
```bash
CHUCHU_DEBUG=1 chu do "add function"
```

Shows: `[MODEL_SELECTOR] Action=edit Lang=go -> openrouter/gemini-2.0-flash-exp:free (score=145.30)`
