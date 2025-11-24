# chu do - Autonomous Task Execution with Self-Healing

## Overview

`chu do` executes tasks autonomously with intelligent auto-recovery. When a model fails, the system automatically learns from the error, recommends an alternative model, and retries—all without user intervention.

## Basic Usage

```bash
chu do "create a hello.txt file with Hello World"
chu do "read docs/README.md and create a getting-started guide"
chu do "unify all feature files in /guides"
```

## Intelligence System

Unlike traditional fallback systems with hardcoded alternatives, `chu do` uses **machine learning** to recommend models based on actual execution history.

### How It Learns

1. **Records every execution** to `~/.chuchu/task_execution_history.jsonl`
   - Task description
   - Model/backend used
   - Success or failure
   - Error details
   - Latency

2. **Calculates success rates** per model/backend combination
   - Initial: 50% confidence (based on known capabilities)
   - After ≥3 tasks: uses actual historical success rate

3. **Recommends intelligently**
   - Prioritizes models with high success rates
   - Considers backend availability
   - Avoids recently failed models
   - Can switch backends automatically

### Example Learning Progression

**First execution:**
```
💡 Intelligence recommends: openrouter/moonshotai/kimi-k2:free
   Confidence: 50%
   Reason: Known to support function calling
```

**After 3 successful tasks:**
```
💡 Intelligence recommends: openrouter/moonshotai/kimi-k2:free
   Confidence: 100%
   Reason: Historical success rate: 100% (3 tasks)
```

## Auto-Recovery Flow

```
┌─────────────────────────────────────────────┐
│ 1. Attempt with current model              │
│    chu do "create file"                     │
└──────────────┬──────────────────────────────┘
               │
               ↓ Fails (tool not available)
               │
┌──────────────┴──────────────────────────────┐
│ 2. Query Intelligence System                │
│    - Check execution history                │
│    - Calculate success rates                │
│    - Recommend best alternative             │
└──────────────┬──────────────────────────────┘
               │
               ↓ openrouter/kimi:free (100%)
               │
┌──────────────┴──────────────────────────────┐
│ 3. Automatic Retry                          │
│    - Switch to recommended model            │
│    - Can change backend if needed           │
│    - No user intervention required          │
└──────────────┬──────────────────────────────┘
               │
               ↓ Success!
               │
┌──────────────┴──────────────────────────────┐
│ 4. Record Result                            │
│    - Update success rate                    │
│    - Improve future recommendations         │
└─────────────────────────────────────────────┘
```

## Command Flags

| Flag | Shorthand | Description | Default |
|------|-----------|-------------|---------|
| `--dry-run` | | Show analysis without executing | false |
| `--verbose` | `-v` | Show intelligence decisions and retries | false |
| `--max-attempts` | | Maximum retry attempts | 3 |

## Examples

### Basic Execution
```bash
chu do "create a config.yaml file"
```

### With Verbose Output
```bash
chu do "extract todos from code" --verbose
```

Shows:
- Current backend/model
- Failure details
- Intelligence recommendation with confidence
- Retry attempts
- Success confirmation

### Dry Run Analysis
```bash
chu do "refactor util functions" --dry-run
```

Analyzes without executing:
- Task intent
- Files affected
- Complexity estimate
- Required steps

## Model Capabilities

The intelligence system tracks models that support **function calling** (tools for file editing):

### Known Compatible
- **OpenRouter**: `moonshotai/kimi-k2:free`, `google/gemini-2.0-flash-exp:free`
- **Groq**: `moonshotai/kimi-k2-instruct-0905` (requires API support)
- **OpenAI**: `gpt-4-turbo`, `gpt-4`
- **Ollama**: `qwen3-coder`

### Backend Switching

The system can automatically switch backends during retry:

```
Attempt 1: groq/model-x → Failed
Attempt 2: openrouter/model-y → Success ✓
```

This requires both backends to be configured in `~/.chuchu/setup.yaml`.

## Execution History

View your execution history:

```bash
cat ~/.chuchu/task_execution_history.jsonl | jq
```

Example output:
```json
{
  "timestamp": "2025-11-24T14:30:38Z",
  "task": "create a hello.txt file",
  "backend": "groq",
  "model": "moonshotai/kimi-k2-instruct-0905",
  "success": false,
  "error": "tool 'read_file' not available"
}
{
  "timestamp": "2025-11-24T14:31:04Z",
  "task": "create a hello.txt file",
  "backend": "openrouter",
  "model": "moonshotai/kimi-k2:free",
  "success": true,
  "latency_ms": 25554
}
```

## Troubleshooting

### No alternative models available

**Problem**: Intelligence system can't find compatible models.

**Solution**: Configure additional backends with function-calling models:

```bash
chu setup  # Add OpenRouter or other providers
```

### Still failing after retries

**Problem**: All attempted models failed.

**Possible causes**:
- Task requires capabilities beyond function calling
- All configured backends have issues
- API keys missing or invalid

**Solution**:
```bash
chu key openrouter  # Add missing API keys
chu do "task" --verbose  # See detailed error messages
```

### Recommendations not improving

**Problem**: Confidence stays at 50% after multiple tasks.

**Reason**: System needs ≥3 executions per model to use historical data.

**Solution**: Keep using the system—it will improve automatically.

## Best Practices

### Let It Learn
- Run tasks naturally
- Don't manually intervene in retries
- System improves with usage

### Use Verbose Mode Initially
```bash
chu do "task" --verbose
```

Helps you understand:
- Which models work for your setup
- Why certain models are recommended
- How confidence builds over time

### Configure Multiple Backends

More backends = more alternatives = higher success rate:

```yaml
backend:
  groq:
    # ...
  openrouter:
    # ...
  ollama:
    # ...
```

### Check History Periodically

```bash
cat ~/.chuchu/task_execution_history.jsonl | \
  jq -s 'group_by(.backend + "/" + .model) | 
         map({
           model: (.[0].backend + "/" + .[0].model),
           success_rate: (map(select(.success)) | length) / length,
           total: length
         })'
```

## vs chu guided

| Feature | chu do | chu guided |
|---------|--------|------------|
| **Approval** | None (autonomous) | Required |
| **Auto-recovery** | ✓ Intelligent retry | ✗ Manual config change |
| **Learning** | ✓ Improves over time | ✗ Static |
| **Speed** | Fast (auto-retry) | Slower (user approval) |
| **Safety** | Medium | High |

**Use `chu do` when:**
- Task is low-risk
- You want autonomous execution
- System has learned your setup

**Use `chu guided` when:**
- Task affects >10 files
- Deleting/moving critical files
- You want to review the plan first

## Implementation

Current version:
- History-based learning (no external ML training needed)
- Automatic retry with model switching
- Cross-backend recommendations
- Real-time confidence calculation

Future enhancements (planned):
- Task feature extraction (complexity, file count)
- Cost optimization in recommendations
- Latency-aware model selection
- Advanced ML models (XGBoost, KAN ensemble)

## Related Commands

- `chu guided` - Interactive mode with plan approval
- `chu plan` - Create plan without execution
- `chu setup` - Configure backends and models
