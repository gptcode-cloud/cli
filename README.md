# GPTCode CLI

> Autonomous AI Coding Assistant — **$0-5/month** vs $20-30/month subscriptions

## Quick Start

```bash
# Install (creates both gptcode and gt commands)
curl -sSL https://gptcode.dev/install.sh | bash

# Or using go install
go install github.com/gptcode/cli/cmd/gptcode@latest
```

## Usage

Use `gptcode` or the short alias `gt`:

```bash
# Autonomous mode - AI completes the task
gt do "add user authentication"

# Interactive chat
gt chat

# Code review
gt review

# Research mode
gt research "how does the payment system work?"
```

## Commands

| Command | Description |
|---------|-------------|
| `gt do "task"` | Autonomous task completion |
| `gt chat` | Interactive code conversation |
| `gt run "task"` | Execute with follow-up |
| `gt research "question"` | Document codebase/architecture |
| `gt plan "task"` | Create implementation plan |
| `gt implement plan.md` | Execute plan step-by-step |
| `gt review` | Code review for bugs/security |
| `gt tdd` | Test-driven development mode |
| `gt feature "desc"` | Generate tests + implementation |

## Skills

Language-specific guidelines injected into prompts:

```bash
gt skills list              # List available skills
gt skills install ruby      # Install Ruby skill
gt skills install-all       # Install all skills
gt skills show go           # View skill content
```

Available: Go, Elixir, Ruby, Rails, Python, TypeScript, JavaScript, Rust

## Configuration

```bash
gt setup                    # Initialize ~/.gptcode
gt key groq                 # Set API key
gt backend use groq         # Switch backend
gt profile use groq.speed   # Switch profile
```

## Why GPTCode?

- **Cost**: $0-5/month using Groq/OpenRouter free tiers
- **Model Selection**: Intelligent routing to best model per task
- **Skills**: Language-specific guidelines for idiomatic code
- **E2E Encryption**: Your code never stored on our servers

## Documentation

- [Installation Guide](https://gptcode.dev/guides/installation)
- [Configuration](https://gptcode.dev/guides/configuration)
- [Skills Index](https://gptcode.dev/skills)
- [API Reference](https://gptcode.dev/reference)

## License

MIT