---
layout: post
title: "Switching Between Local and Cloud: Flexibility When You Need It"
date: 2025-11-18
author: Jader Correa
description: "Switch between local Ollama models for privacy and cloud providers for power. Get the best of both worlds with GPTCode's flexible backend configuration."
tags: [configuration, hybrid, cloud, local, strategy]
---

# Switching Between Local and Cloud: Flexibility When You Need It

The debate between "Local AI" (privacy, free) and "Cloud AI" (intelligence, speed) is a false dichotomy.

## How It Works

GPTCode uses **one backend at a time**, but you can easily switch between local and cloud configurations depending on your current context:

- **Working on sensitive code?** Switch to Ollama (local, private)
- **Need maximum intelligence?** Switch to OpenRouter/Groq (cloud, powerful)
- **Offline or traveling?** Use Ollama without internet
- **Complex task requiring best models?** Use cloud providers

## Configuration Setup

Maintain both configurations in your `~/.gptcode/setup.yaml`:

```yaml
defaults:
  backend: ollama  # Default to local

backend:
  ollama:
    type: ollama
    base_url: http://localhost:11434
    default_model: qwen2.5-coder:7b
    agent_models:
      router: llama3.1:8b
      query: qwen2.5-coder:7b
      editor: qwen2.5-coder:7b
      research: llama3.1:8b
  
  groq:
    type: openai
    base_url: https://api.groq.com/openai/v1
    default_model: gpt-oss-120b-128k
    agent_models:
      router: llama-3.1-8b-instant
      query: gpt-oss-120b-128k
      editor: deepseek-r1-distill-qwen-32b
      research: gpt-oss-120b-128k
```

```

## Switching Between Backends

### In Neovim

Use the backend selector with `Ctrl+X` in the chat buffer to switch on the fly.

### Via Config File

Edit `~/.gptcode/setup.yaml` and change the `defaults.backend` value:

```yaml
defaults:
  backend: groq  # Switch to cloud
```

Then restart your chat session.

## Benefits of This Approach

1.  **Privacy Control**: Keep sensitive code local by using Ollama. Switch to cloud only when needed.
2.  **Cost Optimization**: Use free local models for routine work, pay for cloud only when you need superior intelligence.
3.  **Offline Capability**: Continue working even without internet by switching to Ollama.
4.  **Flexibility**: Choose the right tool for the right job - local for speed and privacy, cloud for power.

## Hardware Requirements

For effective local model usage:
-   **Mac**: M1/M2/M3 with at least 16GB RAM.
-   **Linux/Windows**: NVIDIA GPU with 8GB+ VRAM.

If you don't have powerful hardware, you can still switch between different cloud providers (Groq for speed, OpenRouter for quality) without running anything locally.

## Real-World Usage Patterns

**Morning routine work** (repetitive, familiar code):
- Use Ollama locally
- Fast, free, private
- Perfect for refactoring, small fixes

**Complex feature implementation** (new architecture, tricky logic):
- Switch to OpenRouter with Claude 4.5 Sonnet
- Maximum intelligence for critical decisions
- Worth the cost for important work

**Quick experiments** (learning, trying things out):
- Use Groq for blazing fast feedback
- Cheap enough for experimentation
- Great for iterating quickly

This flexible approach gives you the **best of all worlds**: privacy when you need it, power when you want it, and cost control always.
