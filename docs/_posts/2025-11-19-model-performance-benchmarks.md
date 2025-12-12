---
layout: post
title: "Model Performance Benchmarks: Real-World Coding Comparisons"
date: 2025-11-19
author: Jader Correa
description: "Real-world coding benchmarks comparing models on HumanEval, AIME, and speed metrics. Find the best models for your budget and use case."
tags: [benchmarks, performance, models, comparison]
---

# Model Performance Benchmarks: Real-World Coding Comparisons

*Updated January 2025*

**Important**: AI models evolve rapidly. Benchmark your models using established coding benchmarks like HumanEval[^1], SWE-Bench[^2], and LiveCodeBench[^3].

1. Testing models with your specific workload
2. Checking [Groq configurations]({% post_url 2025-11-15-groq-optimal-configs %}) for current recommendations
3. Exploring [OpenRouter guide]({% post_url 2025-11-16-openrouter-multi-provider %}) for latest models
4. Using `gptcode models search` to discover available models

## Speed vs Quality Trade-offs

### Speed Champions (Groq)

| Model | Speed (TPS) | Use Case |
|-------|-------------|----------|
| **Llama 3.1 8B** | 840+ | Router, fast classification |
| **Qwen3 32B** | 650 | Fast coding with good quality |
| **GPT-OSS 120B** | 500 | Query/research with reasoning |
| **DeepSeek-R1-Qwen-32B** | 600 | Code generation (83.3% AIME) |

*Groq's LPU technology delivers unmatched inference speed.*

### Quality Leaders (OpenRouter)

Based on 2025 benchmarks and real-world testing:

| Model | Strength | Context | Cost |
|-------|----------|---------|------|
| **Claude 4.5 Sonnet** | Code review, debugging | 200k | Premium |
| **Grok 4.1 Fast** | Agentic tasks, 2M context | 2M | Free tier |
| **Qwen 2.5 Coder 32B** | Code generation (88.4% HumanEval) | 131k | Budget |
| **GPT-OSS 120B** | Reasoning, comprehension | 128k | Budget |

## Current Recommendations (2025)

### For Speed + Budget
**Groq Backend**:
```bash
gptcode profiles create groq speed
gptcode profiles set-agent groq speed router llama-3.1-8b-instant
gptcode profiles set-agent groq speed editor llama-3.3-70b-versatile
gptcode profiles set-agent groq speed query llama-3.3-70b-versatile
gptcode profiles set-agent groq speed research groq/compound
```
- **Router**: Llama 3.1 8B Instant (840 TPS, ultra-cheap)
- **Editor**: Llama 3.3 70B Versatile (strong all-around)
- **Query**: Llama 3.3 70B Versatile (good reasoning)
- **Research**: Groq Compound (web search + tools)

### For Maximum Quality
**OpenRouter Backend**:
```bash
gptcode profiles create openrouter quality
gptcode profiles set-agent openrouter quality router google/gemini-2.0-flash-exp:free
gptcode profiles set-agent openrouter quality editor anthropic/claude-4.5-sonnet
gptcode profiles set-agent openrouter quality query anthropic/claude-4.5-sonnet
gptcode profiles set-agent openrouter quality research x-ai/grok-4.1-fast:free
```
- **Router**: Gemini 2.0 Flash (free, fast)
- **Editor**: Claude 4.5 Sonnet (premium quality)
- **Query**: Claude 4.5 Sonnet (best code understanding)
- **Research**: Grok 4.1 Fast (2M context, free tier)

### For Zero Cost
**OpenRouter Free Models**:
```bash
gptcode profiles create openrouter free
gptcode profiles set-agent openrouter free router google/gemini-2.0-flash-exp:free
gptcode profiles set-agent openrouter free editor moonshotai/kimi-k2:free
gptcode profiles set-agent openrouter free query google/gemini-2.0-flash-exp:free
gptcode profiles set-agent openrouter free research x-ai/grok-4.1-fast:free
```
- **Router**: Gemini 2.0 Flash (fastest TTFT)
- **Editor**: Kimi K2 (good coding, free)
- **Query**: Gemini 2.0 Flash or Grok 4.1 Fast (2M context)
- **Research**: Grok 4.1 Fast (2M context, agentic design)

**Note**: Free models have rate limits. For consistent availability, consider adding your own API keys or using paid tiers.

## Discovering Models

Use GPTCode's model search to find available models:
```bash
# Search by provider
gptcode models search groq llama

# Search by features
gptcode models search free coding

# Filter by agent type
gptcode models search --agent editor openrouter
```

See our [detailed configuration guides]({% post_url 2025-11-15-groq-optimal-configs %}) for setup instructions and cost breakdowns.

## References

[^1]: Chen, M., Tworek, J., Jun, H., Yuan, Q., et al. (2021). Evaluating large language models trained on code. *arXiv preprint arXiv:2107.03374*. https://arxiv.org/abs/2107.03374

[^2]: Jimenez, C. E., Yang, J., Wettig, A., et al. (2024). SWE-bench: Can Language Models Resolve Real-World GitHub Issues? *ICLR 2024*. https://arxiv.org/abs/2310.06770

[^3]: Jain, N., Han, K., Gu, A., et al. (2024). LiveCodeBench: Holistic and Contamination Free Evaluation of Large Language Models for Code. *arXiv preprint arXiv:2403.07974*. https://arxiv.org/abs/2403.07974
