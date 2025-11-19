---
title: Commands Reference
description: Complete reference for all Chuchu CLI commands
---

# Commands Reference

Complete guide to all `chu` commands and their usage.

---

## Setup Commands

### `chu setup`

Initialize Chuchu configuration at `~/.chuchu`.

```bash
chu setup
```

Creates:
- `~/.chuchu/profile.yaml` – backend and model configuration
- `~/.chuchu/system_prompt.md` – base system prompt
- `~/.chuchu/memories.jsonl` – memory store for examples

### `chu key [backend]`

Add or update API key for a backend provider.

```bash
chu key openrouter
chu key groq
```

### `chu models update`

Update model catalog from available providers (OpenRouter, Groq, OpenAI, etc.).

```bash
chu models update
```

---

## Interactive Modes

### `chu chat`

Code-focused conversation mode. Routes queries to appropriate agents based on intent.

```bash
chu chat
chu chat "explain how authentication works"
echo "list go files" | chu chat
```

**Agent routing:**
- `query` – read/understand code
- `edit` – modify code
- `research` – external information
- `test` – run tests or commands
- `review` – code review and critique

### `chu tdd`

Incremental TDD mode. Generates tests first, then implementation.

```bash
chu tdd
chu tdd "slugify function with unicode support"
```

Workflow:
1. Clarify requirements
2. Generate tests
3. Generate implementation
4. Iterate and refine

---

## Workflow Commands (Research → Plan → Implement)

### `chu research [question]`

Document codebase and understand architecture.

```bash
chu research "How does authentication work?"
chu research "Explain the payment flow"
```

Creates a research document with findings and analysis.

### `chu plan [task]`

Create detailed implementation plan with phases.

```bash
chu plan "Add user authentication"
chu plan "Implement webhook system"
```

Generates:
- Problem statement
- Current state analysis
- Proposed changes with phases
- Saved to `~/.chuchu/plans/`

### `chu implement <plan_file>`

Execute an approved plan phase-by-phase with verification.

```bash
chu implement ~/.chuchu/plans/2025-01-15-add-auth.md
```

Each phase:
1. Implemented
2. Verified (tests run)
3. User confirms before next phase

---

## Code Quality

### `chu review [target]`

**NEW**: Review code for bugs, security issues, and improvements against coding standards.

```bash
chu review main.go
chu review ./src
chu review .
chu review internal/agents/ --focus security
```

**Options:**
- `--focus` / `-f` – Focus area (security, performance, error handling)

**Reviews against standards:**
- Naming conventions (Clean Code, Code Complete)
- Language-specific best practices
- TDD principles
- Error handling and edge cases

**Output structure:**
1. **Summary**: Overall assessment
2. **Critical Issues**: Must-fix bugs or security risks
3. **Suggestions**: Quality/performance improvements
4. **Nitpicks**: Style, naming preferences

**Examples:**
```bash
chu review main.go --focus "error handling"
chu review . --focus performance
chu review src/auth --focus security
```

---

## Feature Generation

### `chu feature [description]`

Generate tests + implementation with auto-detected language.

```bash
chu feature "slugify with unicode support and max length"
```

**Supported languages:**
- Elixir (mix.exs)
- Ruby (Gemfile)
- Go (go.mod)
- TypeScript (package.json)
- Python (requirements.txt)
- Rust (Cargo.toml)

---

## Execution Mode

### `chu run [task]`

Execute general tasks: HTTP requests, CLI commands, DevOps actions.

```bash
chu run "make a GET request to https://api.github.com/users/octocat"
chu run "deploy to staging using fly deploy"
chu run "check if postgres is running"
```

Perfect for operational tasks without TDD ceremony.

---

## Command Comparison

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `chat` | Interactive conversation | Quick questions, exploratory work |
| `review` | Code review | Before commit, quality check, security audit |
| `tdd` | TDD workflow | New features requiring tests |
| `research` | Understand codebase | Architecture analysis, onboarding |
| `plan` | Create implementation plan | Large features, complex changes |
| `implement` | Execute plan | Structured feature implementation |
| `feature` | Quick feature generation | Small, focused features |
| `run` | Execute tasks | DevOps, HTTP requests, CLI commands |

---

## Environment Variables

### `CHUCHU_DEBUG`

Enable debug output to stderr.

```bash
CHUCHU_DEBUG=1 chu chat
```

Shows:
- Agent routing decisions
- Iteration counts
- Tool execution details

---

## Configuration

All configuration lives in `~/.chuchu/`:

```
~/.chuchu/
├── profile.yaml          # Backend and model settings
├── system_prompt.md      # Base system prompt
├── memories.jsonl        # Example memory store
└── plans/               # Saved implementation plans
    └── 2025-01-15-add-auth.md
```

### Example profile.yaml

```yaml
defaults:
  backend: groq
  model: fast

backends:
  groq:
    type: chat_completion
    base_url: https://api.groq.com/openai/v1
    default_model: llama-3.3-70b-versatile
    models:
      fast: llama-3.3-70b-versatile
      smart: llama-3.3-70b-specdec
```

---

## Next Steps

- See [Research Mode](./research.html) for workflow details
- See [Plan Mode](./plan.html) for plan structure
