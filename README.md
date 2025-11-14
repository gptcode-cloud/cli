# Chuchu 🐶

Strict, impatient, TDD-first coding companion for your **terminal** and **Neovim**.

Chuchu does not coddle you.
Chuchu demands clarity, tests, and small focused functions.
Chuchu keeps you sharp.

---

# Badges

![Go Version](https://img.shields.io/badge/go-1.22+-blue)
![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)
![Mode: TDD](https://img.shields.io/badge/mode-TDD_only-red)
![Backends](https://img.shields.io/badge/backends-groq%20%7C%20ollama%20%7C%20deepinfra-lightgrey)

---

# High-Level Architecture

```
                   ┌──────────────────────────────┐
                   │          Neovim UI           │
                   │   - :ChuchuFeature           │
                   │   - Feedback shortcuts       │
                   └───────────────┬─────────────┘
                                   │
                                   ▼
                       ┌───────────────────────┐
                       │      chuchu CLI       │
                       │  chu feature-elixir   │
                       │  chu chat / chu tdd   │
                       └───────────────┬───────┘
                                       │
                                       ▼
                    ┌──────────────────────────────────┐
                    │        Prompt Builder             │
                    │  - profile.yaml                   │
                    │  - system_prompt.md               │
                    │  - JSONL memory context           │
                    └───────────────────┬──────────────┘
                                        │
                                        ▼
                        ┌──────────────────────────────────┐
                        │        LLM Providers             │
                        │  - Groq (remote, fast)           │
                        │  - Ollama (local, private)       │
                        │  - DeepInfra / OpenAI-compatible │
                        └──────────────────────────────────┘
```

---

# Feature Flow Diagram

```
User: "I want a feature X" 
            │
            ▼
      Neovim Plugin
- Opens floating input
- Sends query → CLI
            │
            ▼
         CLI (chu)
- Detects project
- Loads profile/system prompt
- Loads memory examples
- Builds LLM request
            │
            ▼
         LLM Provider
- Returns:
  • test file
  • implementation file
  • reasoning (optional)
            │
            ▼
      Neovim Plugin
- Opens tests on left
- Opens implementation on right
- Chat buffer at bottom
```

---

# Installation

## Requirements

- Go 1.22+
- Neovim 0.9+
- At least one backend:
  - Ollama
  - OpenAI-compatible providers

---

## Cobra Installation

Chuchu uses Cobra.  
Install it if missing:

```
go get github.com/spf13/cobra@latest
```

---

# CLI Installation

## Using Makefile

```
make install
chu setup
```

## Or script

```
chmod +x scripts/install.sh
./scripts/install.sh
```

---

# Neovim Installation

Using **lazy.nvim**:

```lua
{
  dir = "~/workspace/chuchu/neovim",
  config = function()
    require("chuchu").setup()
  end,
}
```

Available commands:

```
:ChuchuCode         (,cd) - Generate TDD code
:ChuchuVerified     (,vf) - Mark code as verified/good
:ChuchuFailed       (,fr) - Mark code as failed/bad
:ChuchuShell        (,xs) - Get help with shell commands
:ChuchuToggleChat   (,cc) - Toggle conversation panel
```

---

# Memory System (`~/.chuchu/memories.jsonl`)

Each feedback call saves a JSONL line:

```json
{
  "timestamp": "2025-11-13T23:59:00Z",
  "kind": "good",
  "language": "elixir",
  "file": "/path/to/module.ex",
  "snippet": "defmodule Example do ..."
}
```

Chuchu uses these as **few-shot examples** to adapt to your coding style.

---

# Example Workflows

## 1. Generate Code (Elixir)

```
:ChuchuCode (or ,cd)

“Add invoice total calculator with
- rounding
- discounts
- validation for empty list”
```

Results:

- `test/my_app/invoice_total_test.exs`
- `lib/my_app/invoice_total.ex`
- Chat buffer explaining reasoning

---

## Sample Test Output (Elixir)

```elixir
defmodule MyApp.InvoiceTotalTest do
  use ExUnit.Case, async: true

  describe "calculate/1" do
    test "sums prices and applies discount" do
      items = [%{price: 50}, %{price: 70}]
      assert InvoiceTotal.calculate(items, discount: 0.1) == 108.0
    end

    test "rejects empty list" do
      assert {:error, :empty_items} = InvoiceTotal.calculate([])
    end
  end
end
```

---

## Sample Implementation (Elixir)

```elixir
defmodule MyApp.InvoiceTotal do
  def calculate([], _opts \ []), do: {:error, :empty_items}

  def calculate(items, opts \ []) do
    total =
      items
      |> Enum.map(& &1.price)
      |> Enum.sum()

    discount = Keyword.get(opts, :discount, 0.0)
    total * (1 - discount)
  end
end
```

---

# 2. Feature (Typescript)

```
echo "slugify utility with:
- unicode support
- optional max length
- collapse duplicates" | chu feature-ts
```

Generated:

- `src/utils/slugify.ts`
- `tests/slugify.test.ts`

---

# Chat Mode Example (TS)

```
> chu chat
Chuchu: State your problem clearly.
User: I'm debugging a race condition in express middleware.
Chuchu: Show the code. No summaries.
```

---

# TDD Mode Example

```
> chu tdd
Chuchu: Describe the unit you are adding.
User: Token expiration validator.

Chuchu: Tests first. Give me inputs + expected outputs.
```

---

# Philosophy

- Think before coding
- Ask clarifying questions
- Write tests first
- Keep functions small
- Avoid magic
- Prefer explicit data transformations
- Naming is everything

Chuchu gives structure, not fluff.

---

# License

MIT

---

