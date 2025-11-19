---
layout: default
title: "Cost Tracking and Optimization: Monitor Your AI Spending"
---

# Cost Tracking and Optimization: Monitor Your AI Spending

*November 22, 2024*

As you integrate AI into your daily workflow, API costs can accumulate. Chuchu is designed to be cost-effective, but visibility is key. This guide shows you how to track your spending and optimize your configuration for maximum value.

## Built-in Cost Tracking

Chuchu automatically tracks token usage for every request. You can view your usage statistics directly from the CLI.

```bash
chu stats
```

This command outputs a summary of your usage:

```text
Total Requests: 1,245
Total Input Tokens: 45.2M
Total Output Tokens: 2.1M
Estimated Cost (Groq): $4.52
Estimated Cost (OpenRouter): $1.20
Total Cost: $5.72
```

## Optimization Strategies

### 1. Use Smaller Models for Routing
The `Router` agent runs on *every* interaction to determine intent. Using a massive model like GPT-4o for this is overkill and expensive.
- **Recommendation**: Use `llama-3.1-8b-instant` on Groq. It costs pennies per million tokens and is extremely fast.

### 2. Cache Context (Coming Soon)
We are working on context caching support. This will allow you to "pre-load" your codebase into the model's context once and only pay for the diffs in subsequent requests.

### 3. Local Models for Trivial Tasks
If you have a decent GPU (or even a modern M-series Mac), you can offload simple tasks to Ollama.
- Configure `router` and `editor` to use a local `qwen2.5-coder:7b`.
- Use paid APIs only for `review` and complex `query` tasks.

## Setting Budget Alerts

You can set a monthly budget warning in your config:

```yaml
settings:
  budget_warning_usd: 10.00
```

Chuchu will warn you when your estimated usage for the month exceeds this threshold, helping you avoid surprise bills.
