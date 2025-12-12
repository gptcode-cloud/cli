---
layout: post
title: "OpenRouter: Access to Premium Models in One Place"
date: 2025-11-16
author: Jader Correa
description: "OpenRouter provides unified access to the best AI models from multiple providers through a single API, including Claude 4.5, Grok 4.1, and free tier models."
tags: [configuration, openrouter, multi-provider, setup]
---

# OpenRouter: Access to Premium Models in One Place

OpenRouter provides unified access to the best AI models from multiple providers through a single API.

## Why OpenRouter?

OpenRouter aggregates models from:
- **Anthropic**: Claude 4.5 Sonnet - exceptional reasoning and code quality
- **xAI**: Grok 4.1 Fast - best agentic tool calling with 2M context
- **OpenAI**: GPT-OSS 120B, o1/o3 series - reasoning and synthesis
- **Meta**: Llama models - fast and cost-effective
- **Alibaba**: Qwen Coder - specialized code generation (88% HumanEval)
- **Google**: Gemini - multimodal capabilities

You use **one backend at a time**, but can configure different agents to use different models from OpenRouter's catalog.

## Configuration Guide

### Step 1: Get Your OpenRouter Key

1. Sign up at [openrouter.ai](https://openrouter.ai)
2. Create an API key
3. Add it to GPTCode:
   ```bash
   chu key openrouter sk-or-v1-...
   ```

### Step 2: Configure Agent Models

Edit `~/.gptcode/setup.yaml` with this killer configuration:

```yaml
backend:
  openrouter:
    type: openai
    base_url: https://openrouter.ai/api/v1
    default_model: anthropic/claude-4.5-sonnet
    agent_models:
      router: meta-llama/llama-3.1-8b-instruct
      query: anthropic/claude-4.5-sonnet
      editor: anthropic/claude-4.5-sonnet
      research: x-ai/grok-4.1-fast
      review: anthropic/claude-4.5-sonnet
```

## Why This Configuration Works

### Router: `meta-llama/llama-3.1-8b-instruct`
- **Purpose**: Fast intent classification
- **Why**: Cheapest and fastest for simple routing decisions
- **Cost**: ~$0.05/1M tokens input

### Query: `anthropic/claude-4.5-sonnet`
- **Purpose**: Reading and analyzing code
- **Why**: Best comprehension and reasoning capabilities
- **Context**: 200k tokens
- **Cost**: ~$3/1M tokens input, $15/1M output

### Editor: `anthropic/claude-4.5-sonnet`
- **Purpose**: Writing and modifying code
- **Why**: Superior code generation quality and reliability
- **Strength**: Excellent at following instructions and maintaining code style

### Research: `x-ai/grok-4.1-fast`
- **Purpose**: Web search and tool use
- **Why**: Designed specifically for agentic workflows with tool calling
- **Context**: 2M tokens (massive context window!)
- **Special**: Can enable/disable reasoning with `reasoning_enabled` parameter
- **Cost**: Competitive pricing for agentic use cases

### Review: `anthropic/claude-4.5-sonnet`
- **Purpose**: Code review and analysis
- **Why**: Catches subtle bugs and provides thoughtful feedback

## Alternative Configurations

### All-Free (Zero Cost!) 🎉
```yaml
agent_models:
  router: google/gemini-2.0-flash-exp:free
  query: x-ai/grok-4.1-fast
  editor: kwaipilot/kat-coder-pro-v1:free
  research: x-ai/grok-4.1-fast
  review: qwen/qwen-3-coder-480b-a35b:free
```

**All FREE models! Perfect for unlimited usage with zero cost:**
- **Gemini 2.0 Flash**: Fastest time-to-first-token for instant routing
- **Grok 4.1 Fast**: 2M context window for massive codebases
- **KAT-Coder-Pro V1**: 73.4% on SWE-Bench, specialized for agentic coding
- **Qwen3 Coder 480B**: MoE architecture with deep code understanding

### Budget-Conscious (Lower Cost)
```yaml
agent_models:
  router: meta-llama/llama-3.1-8b-instruct
  query: openai/gpt-oss-120b
  editor: alibaba/qwen-2.5-coder-32b-instruct
  research: x-ai/grok-4.1-fast
  review: anthropic/claude-4.5-sonnet
```

Use GPT-OSS 120B for query (better reasoning than Llama 3.3 70B at lower cost), Qwen 2.5 Coder for editor (88.4% HumanEval), and premium models for research and review.

### All-In Performance (Maximum Quality)
```yaml
agent_models:
  router: meta-llama/llama-3.1-8b-instruct
  query: anthropic/claude-4.5-sonnet
  editor: anthropic/claude-4.5-sonnet
  research: x-ai/grok-4.1-fast
  review: openai/o1
```

Add OpenAI's o1 for code review when you need the absolute best reasoning.

### Grok-Heavy (Agentic Focused)
```yaml
agent_models:
  router: meta-llama/llama-3.1-8b-instruct
  query: x-ai/grok-4.1-fast
  editor: anthropic/claude-4.5-sonnet
  research: x-ai/grok-4.1-fast
  review: x-ai/grok-4.1-fast
```

Maximize Grok's 2M context and agentic capabilities for complex multi-step tasks.

## Free Models Deep Dive

### Grok 4.1 Fast (x-ai/grok-4.1-fast) - FREE!

Grok 4.1 Fast is particularly interesting for AI coding agents:

- **Agentic Design**: Built from the ground up for tool calling and multi-step workflows
- **Real-World Use Cases**: Excels at customer support, deep research, and complex debugging
- **2M Context Window**: Can see your entire codebase at once
- **Reasoning Control**: Enable reasoning for complex tasks, disable for speed:
  ```python
  "reasoning_enabled": true  # for complex multi-step debugging
  "reasoning_enabled": false # for quick file searches
  ```
- **Cost**: $0/$0 (currently free on OpenRouter)

### Gemini 2.0 Flash Experimental (google/gemini-2.0-flash-exp:free) - FREE!

- **Fastest TTFT**: Significantly faster time-to-first-token than Gemini 1.5
- **Quality**: On par with larger models like Gemini Pro 1.5
- **Context**: 1.05M tokens - huge for a free model
- **Strengths**: Multimodal understanding, coding, complex instructions, function calling
- **Perfect For**: Router agent - instant responses for intent classification
- **Cost**: $0/$0 (experimental free tier)

### KAT-Coder-Pro V1 (kwaipilot/kat-coder-pro-v1:free) - FREE!

- **SWE-Bench**: 73.4% solve rate on SWE-Bench Verified benchmark
- **Agentic Coding**: Designed specifically for software engineering tasks
- **Multi-Stage Training**: Mid-training, SFT, RFT, and scalable agentic RL
- **Context**: 256K tokens
- **Strengths**: Tool use, multi-turn interaction, instruction following
- **Perfect For**: Editor agent - generates high-quality production code
- **Cost**: $0/$0 (free tier)

### Qwen3 Coder 480B A35B (qwen/qwen-3-coder-480b-a35b:free) - FREE!

- **MoE Architecture**: 480B total parameters, 35B active per forward pass
- **Experts**: 8 out of 160 experts active per token
- **Context**: 262K tokens
- **Strengths**: Function calling, tool use, long-context reasoning over repositories
- **Perfect For**: Code review - deep analysis with MoE reasoning
- **Cost**: $0/$0 for requests under 128k tokens
- **Note**: Pricing increases for >128k input tokens (still very cheap)

## Cost Comparison (Approximate)

| Model | Input ($/1M) | Output ($/1M) | Context | Best For |
|-------|--------------|---------------|---------|----------|
| **FREE MODELS** |||||
| grok-4.1-fast | **$0.00** | **$0.00** | 2M | Agentic workflows |
| gemini-2.0-flash-exp | **$0.00** | **$0.00** | 1.05M | Fast routing |
| kat-coder-pro-v1 | **$0.00** | **$0.00** | 256K | Code generation |
| qwen3-coder-480b | **$0.00** | **$0.00** | 262K | Code review |
| **PAID MODELS** |||||
| llama-3.1-8b | $0.05 | $0.05 | 128K | Budget router |
| gpt-oss-120b | $0.15 | $0.60 | 128K | Budget query/research |
| qwen-2.5-coder-32b | $0.14 | $0.14 | 131K | Budget editor (88% HumanEval) |
| claude-4.5-sonnet | $3.00 | $15.00 | 200K | Premium quality |
| o1/o3 | $15.00+ | $60.00+ | 200K | Deep reasoning |

*Prices are approximate and subject to change. Check [openrouter.ai/models](https://openrouter.ai/models) for current pricing.*

## Setup

1. Update your model catalog:
   ```bash
   chu models update
   ```

2. Verify your configuration:
   ```bash
   chu profile
   ```

3. Test with a chat:
   ```bash
   chu chat
   ```

## The Result

With OpenRouter, you get access to the best models from every major provider without managing multiple API keys and configurations. Your agents automatically use the right model for each task.

### All-Free Configuration: $0/month
With the all-free setup, you get:
- **Gemini 2.0 Flash**: Instant routing responses
- **Grok 4.1 Fast**: 2M context for massive codebases (query & research)
- **KAT-Coder-Pro V1**: 73.4% SWE-Bench performance for code generation
- **Qwen3 Coder 480B**: MoE reasoning for thorough code reviews

**Zero cost. Unlimited usage. Professional quality.**

### Premium Configuration: ~$5-10/month
For maximum quality:
- **Claude 4.5** for high-quality code generation and analysis
- **Grok 4.1 Fast** for agentic research with massive context
- **Llama 3.1** for cost-effective routing

It's the most flexible and powerful way to run GPTCode.

## Tips

- **Start FREE**: Try the all-free configuration first - it's surprisingly powerful
- **Grok 4.1 Fast is free**: Use it liberally for its 2M context window
- **Monitor usage**: Check [openrouter.ai/dashboard](https://openrouter.ai/dashboard) even for free tier
- **Free tier limitations**: Some free models may have rate limits or change pricing
- **Upgrade selectively**: If you need more quality, upgrade just the editor to Claude 4.5
- **Reasoning control**: Enable Grok's reasoning for complex debugging, disable for speed

---

*Share your optimized OpenRouter configurations in [GitHub Discussions](https://github.com/yourusername/gptcode/discussions)!*
