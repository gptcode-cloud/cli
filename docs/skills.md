---
layout: default
title: Skills
description: Language-specific guidelines and best practices that gptcode injects into AI prompts for better code quality.
permalink: /skills/
---

# GPTCode Skills

Skills are language-specific guidelines that gptcode automatically injects into AI prompts when working with code. They help the AI write idiomatic, maintainable code following community best practices.

## How Skills Work

1. **Detection**: gptcode detects the programming language of your project
2. **Injection**: The relevant skill is automatically injected into the AI's system prompt
3. **Better Output**: The AI follows language-specific patterns and idioms

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
gptcode skills list

# Install a specific skill
gptcode skills install ruby

# Install all built-in skills
gptcode skills install-all

# View skill content
gptcode skills show ruby
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
