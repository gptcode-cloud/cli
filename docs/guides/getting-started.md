# Getting Started

This guide helps you install GPTCode, configure providers, and start using the core workflows. It also includes a 10‑second quick start for universal feedback capture.

## Install

- Build from source:
```bash
# from repository root
go build -o ./bin/gptcode ./cmd/gptcode
```
- Or use your preferred package manager (coming soon).

## Initial setup
```bash
gt setup             # creates ~/.gptcode/setup.yaml
gptcode key openrouter    # add API key(s) as needed
gt backend           # check current backend
gt backend list      # list all backends
gt profile           # check current profile
```

## Quick start: two‑keystroke feedback (Ctrl+g)
Capture corrections from any CLI as training signals.
```bash
# zsh
gt feedback hook install --with-diff --and-source

# bash
gt feedback hook install --shell=bash --with-diff --and-source

# fish
gt feedback hook install --shell=fish --with-diff
```
Usage:
1) Type/paste the suggested command
2) Press Ctrl+g to mark the suggestion
3) Edit (if needed) and press Enter

GPTCode records good/bad outcomes and saves changed files and optional git patch.

Check stats:
```bash
gt feedback stats
```

## Core commands

- Chat (code‑focused Q&A):
```bash
gt chat "how does auth middleware work?"
```

- Orchestrated execution (Analyzer → Planner → Editor → Validator):
```bash
gt do "add feature"
gt do --supervised "refactor module"
```

- Model management:
```bash
gt model list
gt model recommend editor
```

## Troubleshooting
- Missing API keys: `gptcode key <backend>`
- Hook not active: `source ~/.zshrc` (or your shell rc) and try Ctrl+g again
- Files not captured: ensure you are inside a git repo
