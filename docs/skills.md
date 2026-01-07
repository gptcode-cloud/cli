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
