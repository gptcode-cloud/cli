---
title: Chuchu
description: Strict, TDD-first coding companion for your terminal and Neovim.
---

# Chuchu 🐶

Strict, impatient, TDD-first coding companion for your **terminal** and **Neovim**.

No bloated UI. No fake “AI buddy”.  
Chuchu is here to keep you honest: clear requirements, tests first, small focused functions.

---

## What is Chuchu?

Chuchu is a thin layer around modern LLMs (local or remote) that enforces:

- **TDD-first workflow** – tests drive design, implementation comes after.
- **Small increments** – no 400-line god functions.
- **Clear naming & structure** – readable by a tired senior engineer at 2am.
- **Editor integration** – tight Neovim + terminal workflow.
- **Memory** – feedback from your editor is stored and reused as examples.

You keep your stack, your editor, your tools.  
Chuchu just makes them a lot more disciplined.

---

## How it Works

```mermaid
flowchart LR
  NV[Neovim
:ChuchuFeature] --> CLI[chuchu CLI
chu feature-*]
  CLI --> PB[Prompt Builder
profile.yaml + system_prompt.md + memory]
  PB --> LLM[LLM Provider
(Groq / Ollama / OpenAI-compatible)]
  LLM --> CLI
  CLI --> NV
```

1. You call `:ChuchuFeature` inside Neovim.
2. The plugin calls `chu feature-<lang>` (Elixir, TS, etc).
3. The CLI:
   - loads your profile + system prompt,
   - pulls recent examples from `~/.chuchu/memories.jsonl`,
   - builds a strict TDD-oriented system prompt.
4. The LLM returns:
   - a **tests** file,
   - an **implementation** file,
   - optionally some structured reasoning.
5. The plugin opens:
   - tests on the left,
   - implementation on the right,
   - conversation log in a side split.

You stay in the editor, moving between tests and implementation, as you should.

---

## Key Components

- **CLI (`chu`)**
  - `chu setup` – bootstrap `~/.chuchu` (profile + system prompt)
  - `chu chat` – code-focused chat
  - `chu review` – code review for bugs, security, and quality
  - `chu tdd` – incremental TDD mode
  - `chu research` – document codebase and understand architecture
  - `chu plan` – create detailed implementation plan
  - `chu implement` – execute plan with verification
  - `chu feature` – generate tests + implementation (auto-detects language)

- **Neovim plugin**
  - `:ChuchuFeature` – feature workflow with split layout
  - `:ChuchuFeedbackGood` / `:ChuchuFeedbackBad`
    - store current buffer as examples in `~/.chuchu/memories.jsonl`

- **Memory**
  - JSON Lines file at `~/.chuchu/memories.jsonl`
  - Loaded by the CLI via a `JSONLMemStore`
  - Used as few-shot examples per language

---

## Installation

### 1. Requirements

- Go 1.22+
- Neovim 0.9+ (for plugin)
- At least one LLM backend:
  - [Ollama](https://ollama.com/)
  - Groq
  - OpenAI-compatible provider

### 2. Cobra (CLI framework)

Chuchu CLI uses [Cobra](https://github.com/spf13/cobra). Install it with:

```bash
go get github.com/spf13/cobra@latest
```

### 3. Install CLI

Using the Makefile:

```bash
make install
chu setup
```

Or using the helper script:

```bash
chmod +x scripts/install.sh
./scripts/install.sh
```

### 4. Neovim Plugin

Example using `lazy.nvim`:

```lua
{
  dir = "~/workspace/chuchu/neovim",
  config = function()
    require("chuchu").setup()
  end,
}
```

Then inside Neovim:

```vim
:ChuchuFeature
:ChuchuFeedbackGood
:ChuchuFeedbackBad
```

---

## Example: Elixir Feature

```elixir
:ChuchuFeature
```

You type:

> “Add an `InvoiceTotal` module that:
>  - sums line items,
>  - applies an optional discount,
>  - fails fast on empty lists.”

Chuchu generates:

- `test/my_app/invoice_total_test.exs`
- `lib/my_app/invoice_total.ex`

You iterate on tests and implementation from there.

---

## Example: TypeScript Feature

```bash
echo "slugify with unicode support and max length" | chu feature-ts
```

Chuchu generates:

- `tests/slugify.test.ts`
- `src/slugify.ts`

You run tests with Vitest or Jest, tweak behavior, and repeat.

---

## Philosophy

- Think before coding.
- Ask clarifying questions.
- Write tests first.
- Keep functions small.
- Name things clearly.
- Avoid magic.

Chuchu is not your friend.  
Chuchu is the strict reviewer sitting next to you, full-time.

---

## Documentation

- [Commands Reference](./commands.html) – Complete guide to all CLI commands
- [Research Mode](./research.html) – Research workflow details
- [Plan Mode](./plan.html) – Planning workflow

---

## License

MIT
