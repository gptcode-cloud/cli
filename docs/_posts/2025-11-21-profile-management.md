---
layout: post
title: "Profile Management: Multiple Configurations per Backend"
date: 2025-11-21
author: Jader Correa
description: "Switch between multiple named configurations per backend. Create speed, quality, and free profiles to optimize your workflow without editing config files."
tags: [features, configuration, profiles, productivity]
---

GPTCode now supports **profiles** - multiple named configurations per backend. This lets you quickly switch between different model setups without editing configuration files.

## Why Profiles?

Different tasks need different model configurations:

- **Speed profile**: Fast models for all agents (lower latency, lower cost)
- **Quality profile**: Best models for complex tasks (higher accuracy)
- **Free profile**: Only free-tier models for experimentation
- **Production profile**: Reliable, battle-tested models

Instead of manually editing `~/.gptcode/setup.yaml`, create and switch between profiles.

## Quick Start

### Show Current Profile

```bash
chu profile
# Current: groq/default
#   router: llama-3.1-8b-instant
#   query: llama-3.3-70b-versatile
#   editor: moonshotai/kimi-k2-instruct-0905
#   research: groq/compound
```

### List Available Profiles

```bash
chu profile list
# groq.default (current)
# groq.speed
# groq.quality
# openrouter.default
# openrouter.free

chu profile list groq
# groq.default (current)
# groq.speed
# groq.quality
```

### Show Profile Configuration

```bash
chu profile show groq.speed
# groq/speed
#   router: llama-3.1-8b-instant
#   query: llama-3.1-8b-instant
#   editor: llama-3.1-8b-instant
#   research: llama-3.1-8b-instant
```

### Switch Profile

```bash
chu profile use groq.speed
# ✓ Switched to groq/speed
```

### Create New Profile

```bash
chu profiles create groq speed
# ✓ Created profile: groq/speed
# 
# Configure agent models using:
#   chu profiles set-agent groq speed router <model>
#   chu profiles set-agent groq speed query <model>
#   chu profiles set-agent groq speed editor <model>
#   chu profiles set-agent groq speed research <model>
```

### Configure Agent Models

```bash
chu profiles set-agent groq speed router llama-3.1-8b-instant
chu profiles set-agent groq speed query llama-3.1-8b-instant
chu profiles set-agent groq speed editor llama-3.1-8b-instant
chu profiles set-agent groq speed research llama-3.1-8b-instant
# ✓ Set groq/speed router = llama-3.1-8b-instant
# ...
```

## Profile Structure in setup.yaml

Profiles are stored in your `~/.gptcode/setup.yaml`:

```yaml
defaults:
    backend: groq
    profile: default  # currently active profile

backend:
    groq:
        # ... backend config ...
        agent_models:     # "default" profile (backwards compatible)
            router: llama-3.1-8b-instant
            query: llama-3.3-70b-versatile
            editor: moonshotai/kimi-k2-instruct-0905
            research: groq/compound
        profiles:         # named profiles
            speed:
                agent_models:
                    router: llama-3.1-8b-instant
                    query: llama-3.1-8b-instant
                    editor: llama-3.1-8b-instant
                    research: llama-3.1-8b-instant
            quality:
                agent_models:
                    router: llama-3.3-70b-versatile
                    query: llama-3.3-70b-versatile
                    editor: llama-3.3-70b-versatile
                    research: groq/compound
```

## Switching Profiles in Neovim

In Neovim, press `Ctrl+M` to open the model selector:

1. Select backend (e.g., "groq")
2. Select profile (e.g., "speed", "quality")
3. The chat header updates to show active profile

```
🐺 GPTCode
Backend: Groq / speed
  router: llama-3.1-8b-instant
  query: llama-3.1-8b-instant
  editor: llama-3.1-8b-instant
  research: llama-3.1-8b-instant
```

## Example: OpenRouter Free Models

Create a profile using only free-tier OpenRouter models:

```bash
chu profiles create openrouter free

chu profiles set-agent openrouter free router \
  google/gemini-2.0-flash-exp:free

chu profiles set-agent openrouter free query \
  google/gemini-2.0-flash-exp:free

chu profiles set-agent openrouter free editor \
  moonshotai/kimi-k2:free

chu profiles set-agent openrouter free research \
  google/gemini-2.0-flash-exp:free
```

Now you can experiment with free models without API costs.

## Agent Types

Each profile configures four agent types:

- **router**: Fast model for intent classification (determines which agent handles request)
- **query**: Model for reading and analyzing code
- **editor**: Model for writing and modifying code
- **research**: Model for web search and documentation lookup

## Best Practices

### Speed Profile

Use fast, cheap models for rapid iteration:

```bash
chu profiles create groq speed
chu profiles set-agent groq speed router llama-3.1-8b-instant
chu profiles set-agent groq speed query llama-3.1-8b-instant
chu profiles set-agent groq speed editor llama-3.1-8b-instant
chu profiles set-agent groq speed research llama-3.1-8b-instant
```

### Quality Profile

Use best available models for complex tasks:

```bash
chu profiles create groq quality
chu profiles set-agent groq quality router llama-3.1-8b-instant  # routing is simple
chu profiles set-agent groq quality query llama-3.3-70b-versatile
chu profiles set-agent groq quality editor llama-3.3-70b-versatile
chu profiles set-agent groq quality research groq/compound
```

### Specialized Profiles

Create profiles for specific use cases:

```bash
# Code-heavy tasks
chu profiles create groq coding
chu profiles set-agent groq coding editor deepseek-v3

# Research-heavy tasks
chu profiles create groq docs
chu profiles set-agent groq docs research groq/compound
```

## Migration from Old Config

If you have `agent_models` at the backend root level (old format), it automatically becomes the "default" profile. No manual migration needed.

Old format:
```yaml
backend:
    groq:
        agent_models:
            router: ...
```

Works as:
```bash
chu profile show groq.default
# Shows the models from agent_models
```

## Troubleshooting

### Profile Not Found

```bash
chu profile list groq
# Check if profile exists

chu profiles create groq myprofile
# Create if missing (use plural 'profiles' for creation)
```

### Wrong Models Showing

Verify profile configuration:

```bash
chu profile show groq.myprofile
# Check each agent's model

chu profiles set-agent groq myprofile router correct-model
# Fix individual agents (use plural 'profiles' for configuration)
```

### Profile Changes Not Reflecting in Neovim

Restart Neovim or reload the buffer. The plugin reads configuration on startup.

## Implementation Details

Profiles use proper YAML marshaling (not manual parsing), ensuring:

- Clean, maintainable code
- Proper error handling
- Type safety
- Easy extensibility

The profile system replaces fragile space-counting logic with Go's `yaml.v3` library.

## Future Enhancements

Planned features:

- `chu profiles copy <src> <dst>` - Clone existing profile
- `chu profiles delete <backend> <profile>` - Remove profile
- `chu profiles export/import` - Share profiles between machines
- Profile templates for common use cases

## Related Posts

- [Groq Optimal Configurations]({% post_url 2025-11-15-groq-optimal-configs %})
- [OpenRouter Multi-Provider Setup]({% post_url 2025-11-16-openrouter-multi-provider %})
- [Ollama Local Setup]({% post_url 2025-11-17-ollama-local-setup %})

---

Profiles make it easy to switch between different model configurations without manual editing. Create profiles for different use cases and switch seamlessly in Neovim or via CLI.
