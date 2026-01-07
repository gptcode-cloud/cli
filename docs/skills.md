---
layout: default
title: Skills
description: Language-specific expertise that gt injects into AI prompts for generating production-quality, idiomatic code.
permalink: /skills/
---

# GPTCode Skills

**Skills are the secret sauce** behind `gt`'s ability to generate production-quality code. When you run any `gt` command, it automatically detects your project's language and injects the relevant skill into the AI's system prompt.

> 💡 **The result**: Instead of generic code that "works", you get idiomatic code that follows community best practices, proper error handling, and language-specific patterns.

## Why Skills Matter

Without skills, AI models produce **generic** code:
- No language idioms
- Inconsistent error handling
- Poor naming conventions
- Missing documentation patterns

With skills, `gt` produces **production-ready** code:
- Idiomatic patterns (e.g., Go's explicit error handling, Elixir's pattern matching)
- Consistent style following community guidelines
- Proper documentation and testing patterns
- Framework-specific best practices (Rails, Phoenix, React)

## How Skills Work

```
┌──────────────────────────────────────────────────────────┐
│  You run: gt do "add user authentication"                │
│                        ↓                                 │
│  gt detects: Ruby on Rails project (Gemfile, config/)   │
│                        ↓                                 │
│  gt injects: Rails skill + Ruby skill into prompt       │
│                        ↓                                 │
│  AI generates: Service objects, proper migrations,      │
│                RSpec tests, Devise patterns             │
└──────────────────────────────────────────────────────────┘
```

## Available Skills

### Language-Specific

| Skill | Language | Description |
|-------|----------|-------------|
| [Go](/skills/go) | Go | Error handling, naming, interfaces, concurrency |
| [Elixir](/skills/elixir) | Elixir | Pattern matching, OTP, Phoenix, Ecto |
| [Ruby](/skills/ruby) | Ruby | Method design, error handling, testing |
| [Rails](/skills/rails) | Ruby | Active Record, controllers, services, RSpec |
| [Python](/skills/python) | Python | PEP 8, type hints, pytest, comprehensions |
| [TypeScript](/skills/typescript) | TypeScript | Types, generics, async patterns, React |
| [JavaScript](/skills/javascript) | JavaScript | ES6+, async/await, modules, array methods |
| [Rust](/skills/rust) | Rust | Ownership, error handling, iterators, traits |

### General

| Skill | Description |
|-------|-------------|
| [TDD Bug Fix](/skills/tdd-bug-fix) | Write failing tests before fixing bugs |
| [Code Review](/skills/code-review) | Structured code review with priorities |
| [Git Commit](/skills/git-commit) | Conventional commit messages |

## Installing Skills

```bash
# List available skills
gt skills list

# Install a specific skill
gt skills install ruby

# Install all built-in skills
gt skills install-all

# View skill content
gt skills show ruby
```

## Creating Custom Skills

You can create custom skills for your team or Stack:

1. Create a markdown file in `~/.gptcode/skills/`
2. Add frontmatter with `name`, `language`, and `description`
3. The skill will be automatically loaded when working with that language

### Example Custom Skill

```markdown
---
name: my-company-style
language: typescript
description: Our company's TypeScript conventions
---

# Company TypeScript Style

## Always use strict mode
...
```

## Contributing Skills

Want to add a skill for your favorite language? [Open a PR](https://github.com/gptcode/cli/tree/main/docs/_skills) with your skill markdown file.

---

## Using Skills in Other Tools

The skills that power `gt` can also be used in other AI coding tools. Here's how to apply them in your favorite environment:

### Cursor

Cursor has the **strongest support** for project rules via `.cursorrules`:

```bash
# Export skills directly to Cursor format
gt skills show go > .cursorrules

# Or combine multiple skills
gt skills show go >> .cursorrules
gt skills show tdd-bug-fix >> .cursorrules
```

Cursor reads `.cursorrules` automatically and applies them to **every** interaction—no need to ask.

### VS Code (GitHub Copilot)

GitHub Copilot supports custom instructions via files in your project:

**Project-wide instructions:**
```bash
# Export a gt skill to Copilot format
gt skills show ruby > .github/copilot-instructions.md
```

The file `.github/copilot-instructions.md` is automatically read by Copilot before every interaction.

**Reusable prompts:**
```bash
# Create a prompt for TDD workflow
gt skills show tdd-bug-fix > .github/prompts/tdd.prompt.md
```

Use in Copilot Chat by referencing the prompt.

### Antigravity (Gemini in IDE)

If you're using Google's Antigravity (Gemini IDE integration):

1. Create `.gemini/settings.json` in your project:
```json
{
  "customInstructions": "See .gemini/skills/ for coding guidelines"
}
```

2. Export skills to `.gemini/skills/`:
```bash
mkdir -p .gemini/skills
gt skills show python > .gemini/skills/python.md
gt skills show tdd-bug-fix > .gemini/skills/tdd.md
```

### Google Gemini (AI Studio / API)

For Gemini, use skills as system instructions:

```bash
# Show skill content to copy
gt skills show typescript
```

Paste the content into:
- **AI Studio**: System Instructions field
- **API**: `system_instruction` parameter in your request

```python
# Example with Gemini API
import google.generativeai as genai

model = genai.GenerativeModel(
    'gemini-pro',
    system_instruction=open('typescript-skill.md').read()
)
```

### Claude (Anthropic Console / API)

Claude supports system prompts where you can inject skills:

```bash
# Export for Claude
gt skills show elixir > claude-system.md
```

Use in:
- **claude.ai**: Start conversation with "Follow these guidelines:" + paste skill
- **API**: Set as `system` parameter

```python
# Example with Claude API
import anthropic

client = anthropic.Anthropic()
message = client.messages.create(
    model="claude-3-opus-20240229",
    system=open('elixir-skill.md').read(),
    messages=[{"role": "user", "content": "Create a GenServer"}]
)
```

### Replit (Agent / Ghostwriter)

Replit doesn't auto-load rule files yet, but you can use a convention:

```bash
# Export to a RULES file
gt skills show python > RULES.md
```

**Tips for Replit:**
1. Keep `RULES.md` **open in an editor tab**—Ghostwriter prioritizes open files
2. When using Replit Agent, start with: *"Read RULES.md and follow strictly"*

---

## Quick Reference

| Tool | Config Location | Auto-loaded? |
|------|-----------------|--------------|
| **gt** | Built-in | ✅ Yes |
| **Cursor** | `.cursorrules` | ✅ Yes (strong) |
| **VS Code Copilot** | `.github/copilot-instructions.md` | ✅ Yes |
| **Antigravity** | `.gemini/skills/` | ⚡ Partial |
| **Gemini** | System instruction | ❌ Manual |
| **Claude** | System prompt | ❌ Manual |
| **Replit** | `RULES.md` (convention) | ❌ Manual |

> 💡 **Pro tip**: Commit your skills files to git so your whole team benefits from consistent AI-generated code!
