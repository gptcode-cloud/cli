---
layout: post
title: "Optimal Groq Configurations for Different Budgets"
date: 2024-11-18
author: Jader Corrêa
tags: [configuration, groq, optimization, cost]
---

# Optimal Groq Configurations for Different Budgets

Groq offers blazing-fast inference speeds with their LPU technology. Here are optimized agent model configurations for different use cases and budgets.

## Understanding Agent Roles

Chuchu uses specialized agents for different tasks:

- **Router**: Fast intent classification (needs speed, not depth)
- **Query**: Reading and analyzing code (needs comprehension)
- **Editor**: Writing and modifying code (needs code generation quality)
- **Research**: Web search and documentation lookup (benefits from tool use)

## Budget-Conscious Configuration ($0.05 - $0.79 per 1M tokens)

Best balance of cost and performance for most developers:

```yaml
backend:
  groq:
    base_url: https://api.groq.com/openai/v1
    default_model: llama-3.1-8b-instant
    agent_models:
      router: llama-3.1-8b-instant      # $0.05/$0.08
      query: gpt-oss-20b-128k                # $0.075/$0.30
      editor: llama-3.3-70b-versatile    # $0.59/$0.79
      research: groq/compound-mini           # $0.11/$0.34 (with tools!)
```

**Monthly estimate** (100M tokens): ~$48

### Why this works:
- Router uses cheapest/fastest model (840 TPS!)
- Query uses efficient 20B model with good comprehension
- Editor uses Llama 3.3 70B for quality code generation
- Research uses Compound Mini with web search capabilities

## Performance-Focused Configuration ($0.11 - $3.00 per 1M tokens)

For projects where code quality is critical and budget is flexible:

```yaml
backend:
  groq:
    base_url: https://api.groq.com/openai/v1
    default_model: llama-3.3-70b-versatile
    agent_models:
      router: llama-3.1-8b-instant      # $0.05/$0.08
      query: llama-3.3-70b-versatile    # $0.59/$0.79
      editor: moonshotai/kimi-k2-instruct-0905          # $1.00/$3.00 (256k context!)
      research: groq/compound                # $0.15/$0.60 (full tools)
```

**Monthly estimate** (100M tokens): ~$120

### Why this works:
- Kimi K2 has 1 trillion parameters and 256k context window
- Llama 3.3 70B handles query tasks with deep understanding
- Full Compound system with GPT-OSS-120B for research
- Still uses fast router for cost efficiency

## Research-Heavy Configuration

Optimized for projects with extensive documentation and web research needs:

```yaml
backend:
  groq:
    base_url: https://api.groq.com/openai/v1
    default_model: gpt-oss-20b-128k
    agent_models:
      router: llama-3.1-8b-instant      # $0.05/$0.08
      query: gpt-oss-20b-128k                # $0.075/$0.30
      editor: gpt-oss-120b-128k              # $0.15/$0.60
      research: groq/compound                # $0.15/$0.60 (tools)
```

**Monthly estimate** (100M tokens): ~$37

### Why this works:
- GPT-OSS models excel at information synthesis
- Full Compound system with web search and browser automation
- Good balance of comprehension and generation

## Speed-Optimized Configuration

When latency matters more than token cost:

```yaml
backend:
  groq:
    base_url: https://api.groq.com/openai/v1
    default_model: llama-3.1-8b-instant
    agent_models:
      router: llama-3.1-8b-instant      # 840 TPS
      query: llama-3.1-8b-instant       # 840 TPS
      editor: llama-3.3-70b-versatile   # 394 TPS
      research: llama-4-scout-17bx16e-128k  # 594 TPS
```

**Monthly estimate** (100M tokens): ~$27

### Why this works:
- Prioritizes Groq's fastest models (TPS = tokens per second)
- Llama 3.1 8B at 840 TPS for instant responses
- Still uses 70B for editor where quality matters
- Llama 4 Scout provides good speed for research

## Model Specifications Reference

| Model | Input | Output | Context | Speed (TPS) | Best For |
|-------|-------|--------|---------|-------------|----------|
| llama-31-8b-instant | $0.05 | $0.08 | 128k | 840 | Router, fast tasks |
| gpt-oss-20b | $0.075 | $0.30 | 128k | 1000 | Query, analysis |
| llama-33-70b-versatile | $0.59 | $0.79 | 128k | 394 | Editor, quality code |
| kimi-k2-0905-1t | $1.00 | $3.00 | 256k | 200 | Large context, complex edits |
| gpt-oss-120b | $0.15 | $0.60 | 128k | 500 | Research, synthesis |
| groq/compound | $0.15 | $0.60 | 131k | 450 | Research with tools |
| groq/compound-mini | $0.11 | $0.34 | 131k | 500 | Budget research with tools |

*Prices per 1M tokens. TPS = tokens per second throughput.*

## Groq Compound Systems

Compound models are special - they combine multiple models with tool capabilities:

### groq/compound
- **Models**: GPT-OSS-120B + Llama 4 Scout
- **Tools**: Web search, code execution, browser automation, Wolfram Alpha
- **Pricing**: Base model pricing + tool costs
  - Basic web search: $5/1000 requests
  - Advanced web search: $8/1000 requests
  - Visit website: $1/1000 requests
  - Code execution: $0.18/hour
  - Browser automation: $0.08/hour

### groq/compound-mini
- **Models**: Llama 4 Scout only
- **Tools**: Same as compound
- **Pricing**: Lower base model cost + tool costs

## Setting Up

1. Update your model catalog:
```bash
chu models update
```

2. Switch to Groq backend and configure agent models in Neovim:
```
Ctrl+X (in chat buffer)
```

3. Or edit `~/.chuchu/setup.yaml` directly with your chosen configuration

## Tips

- Start with budget-conscious config and upgrade specific agents as needed
- Use `groq/compound-mini` for research if you don't need GPT-OSS-120B
- Router agent is called most frequently - keep it fast and cheap
- Editor agent output quality matters most - invest there first
- Monitor your usage at [console.groq.com](https://console.groq.com)

## Switching Between Configs

You can maintain multiple backend configurations:

```yaml
backend:
  groq-fast:
    type: openai
    base_url: https://api.groq.com/openai/v1
    default_model: llama-3.1-8b-instant
    # speed-optimized config
  
  groq-quality:
    type: openai
    base_url: https://api.groq.com/openai/v1
    default_model: llama-3.3-70b-versatile
    # performance-focused config
```

Switch in Neovim with `Ctrl+X` or via CLI:
```bash
chu chat --backend groq-fast
chu chat --backend groq-quality
```

---

*Have your own optimized configuration? Share it on [GitHub Discussions](https://github.com/yourusername/chuchu/discussions)!*
