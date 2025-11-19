---
layout: default
title: "OpenRouter Multi-Provider Setup: The Best of All Worlds"
---

# OpenRouter Multi-Provider Setup: The Best of All Worlds

*November 20, 2024*

One of Chuchu's most powerful features is its ability to mix and match models from different providers. While Groq offers incredible speed for agents like the `Router` and `Editor`, sometimes you need the reasoning capabilities of models hosted elsewhere. Enter **OpenRouter**.

## Why Multi-Provider?

Different agents have different needs:
- **Router**: Needs speed above all else. Groq's Llama 3.1 8B is perfect.
- **Editor**: Needs speed and code generation capability. Groq's Llama 3.1 70B or 8B works well.
- **Query/Research**: Needs deep reasoning and large context windows. Anthropic's Claude 3.5 Sonnet or OpenAI's GPT-4o are often superior here.

## Configuration Guide

To set up a multi-provider environment, you'll need an OpenRouter API key.

1.  **Get your Key**: Sign up at [openrouter.ai](https://openrouter.ai) and create a key.
2.  **Add to Chuchu**:
    ```bash
    chu key openrouter sk-or-v1-...
    ```

3.  **Configure `~/.chuchu/config.yaml`**:

    Here is an optimal configuration that uses Groq for speed and OpenRouter for complex tasks:

    ```yaml
    backends:
      groq:
        type: openai
        base_url: https://api.groq.com/openai/v1
        default_model: llama-3.1-70b-versatile
      
      openrouter:
        type: openai
        base_url: https://openrouter.ai/api/v1
        default_model: anthropic/claude-3.5-sonnet

    agents:
      router:
        backend: groq
        model: llama-3.1-8b-instant  # Super fast routing
      
      editor:
        backend: groq
        model: llama-3.1-70b-versatile # Fast code generation
      
      query:
        backend: openrouter
        model: anthropic/claude-3.5-sonnet # Deep reasoning for complex questions
      
      research:
        backend: openrouter
        model: openai/gpt-4o # Excellent web browsing and synthesis
      
      review:
        backend: openrouter
        model: anthropic/claude-3.5-sonnet # Top-tier code review
    ```

## The Result

With this setup, your `chu chat` experience feels instantaneous for simple interactions (routing, small edits) but switches to the world's smartest models when you ask deep questions or need a thorough code review. It's the perfect balance of latency, cost, and intelligence.
