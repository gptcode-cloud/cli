# gptcode do - Intelligent Task Execution with Efficiency Optimization

## Overview

`gptcode do` executes tasks autonomously with **intelligent efficiency optimization**. The system evaluates all available models across cost, speed, reliability, and availability to find the optimal path. When obstacles appear (failures, rate limits, high costs), it automatically adapts and finds a better route.

## Basic Usage

```bash
gptcode do "create a hello.txt file with Hello World"
gptcode do "read docs/README.md and create a getting-started guide"
gptcode do "unify all feature files in /guides"
```

## Intelligence System

Unlike traditional fallback systems with hardcoded alternatives, `gptcode do` uses **multi-criteria optimization** to find the most efficient model:

**Evaluated Criteria:**
- Success Rate (50%) - Reliability from historical data
- Speed (20%) - Tokens per second
- Cost (20%) - Prefer free models when viable
- Availability (10%) - Rate limits, quotas
- Latency - Real execution time tracking

### How It Learns

1. **Records every execution** to `~/.gptcode/task_execution_history.jsonl`
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
   Overall Score: 0.76
   Success Rate: 50% | Speed: 300 TPS | Cost: $0.000/1M | Latency: 0ms
   Reason: Known capable, Speed: 300 TPS, Cost: $0.00/1M
```

**After 4 successful tasks:**
```
💡 Intelligence recommends: openrouter/moonshotai/kimi-k2:free
   Overall Score: 0.88
   Success Rate: 100% | Speed: 300 TPS | Cost: $0.000/1M | Latency: 20191ms
   Reason: Success: 100% (4 tasks), Speed: 300 TPS, Cost: $0.00/1M
```

## Auto-Recovery Flow

```
┌─────────────────────────────────────────────┐
│ 1. Attempt with current model              │
│    gptcode do "create file"                     │
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

## Execution Modes

### Autonomous Mode (Default)
Uses **orchestrated agent decomposition** for better results:

1. **Analyzer Agent** - Understands codebase context
2. **Planner Agent** - Creates minimal, focused plan
3. **Editor Agent** - Executes ONLY planned changes
4. **Validator Agent** - Verifies success criteria

```bash
gptcode do "add debug flag to config.ini"
```

**Benefits:**
- File validation prevents unintended changes
- Success criteria validation with auto-retry (max 2 attempts)
- Over-engineering protection (no helper scripts unless requested)

### Supervised Mode
Requires manual plan approval before execution:

```bash
gptcode do --supervised "refactor authentication module"
```

**Use when:**
- Task affects critical code
- You want to review the plan first
- Changes involve >5 files

## Command Flags

| Flag | Shorthand | Description | Default |
|------|-----------|-------------|---------|
| `--supervised` | | Require manual plan approval | false |
| `--dry-run` | | Show analysis without executing | false |
| `--verbose` | `-v` | Show intelligence decisions and retries | false |
| `--max-attempts` | | Maximum retry attempts | 3 |

## Examples

### Basic Execution
```bash
gptcode do "create a config.yaml file"
```

### With Verbose Output
```bash
gptcode do "extract todos from code" --verbose
```

Shows:
- Current backend/model
- Failure details
- Intelligence recommendation with confidence
- Retry attempts
- Success confirmation

### Dry Run Analysis
```bash
gptcode do "refactor util functions" --dry-run
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

This requires both backends to be configured in `~/.gptcode/setup.yaml`.

## Execution History

View your execution history:

```bash
cat ~/.gptcode/task_execution_history.jsonl | jq
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
gptcode setup  # Add OpenRouter or other providers
```

### Still failing after retries

**Problem**: All attempted models failed.

**Possible causes**:
- Task requires capabilities beyond function calling
- All configured backends have issues
- API keys missing or invalid

**Solution**:
```bash
gptcode key openrouter  # Add missing API keys
gptcode do "task" --verbose  # See detailed error messages
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
gptcode do "task" --verbose
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
cat ~/.gptcode/task_execution_history.jsonl | \
  jq -s 'group_by(.backend + "/" + .model) | 
         map({
           model: (.[0].backend + "/" + .[0].model),
           success_rate: (map(select(.success)) | length) / length,
           total: length
         })'
```

## vs chu guided

| Feature | gptcode do | chu guided |
|---------|--------|------------|
| **Approval** | None (autonomous) | Required |
| **Auto-recovery** | ✓ Intelligent retry | ✗ Manual config change |
| **Learning** | ✓ Improves over time | ✗ Static |
| **Speed** | Fast (auto-retry) | Slower (user approval) |
| **Safety** | Medium | High |

**Use `gptcode do` when:**
- Task is low-risk
- You want autonomous execution
- System has learned your setup

**Use `gptcode guided` when:**
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

- `gptcode guided` - Interactive mode with plan approval
- `gptcode plan` - Create plan without execution
- `gptcode setup` - Configure backends and models
