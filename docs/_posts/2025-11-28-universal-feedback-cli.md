---
layout: post
title: "Learn from Any CLI: Two‑Keystroke Feedback (Ctrl+g)"
date: 2025-11-28
author: Jader Correa
description: "Shell hooks and a universal sink capture corrections from any CLI with two keystrokes. Your machine learns from you."
tags: [dx, feedback, learning]
---

# Learn from Any CLI: Two‑Keystroke Feedback (Ctrl+g)

With two keystrokes you turn everyday corrections into continuous learning for GPTCode.

## How it works
- Press **Ctrl+g** to mark the suggestion on the terminal line
- Edit and press **Enter**
- The hook automatically records:
  - Sentiment: `good` (same) or `bad` (different)
  - `wrong_response` e `correct_response`
  - `files` changed (git diff --name-only)
  - `diff_path` with the full patch (optional)

## Installation
```bash
# zsh (with diff)
chu feedback hook install --with-diff --and-source
# bash
chu feedback hook install --shell=bash --with-diff --and-source
# fish
chu feedback hook install --shell=fish --with-diff
```

## See it in action

### Keyboard flow (Ctrl+g → run)
![Two‑keystroke feedback (hook path)](../assets/feedback-hook-demo.gif?v={{ site.github.build_revision | default: site.time | date: '%s' }})

### Programmatic flow (CLI submit)
![Two‑keystroke feedback demo](../assets/feedback-demo.gif?v={{ site.github.build_revision | default: site.time | date: '%s' }})

### How it’s used for training
```bash
# Convert feedback JSON to training examples
python3 ml/intent/scripts/process_feedback.py

# Retrain the intent model (feedback weighted)
chu ml train intent
```

## For UIs/CLIs
No API integration required. Let users press **Ctrl+g** before running.

## Advanced
- Manual/programmatic submissions: `chu feedback submit --json -`
- Attach files: `--files path --files another`
- Capture patch: `--capture-diff` (saved under `~/.gptcode/diffs/*.patch`)

## References
- Guia: [Universal Feedback Capture](../guides/feedback)
- Post: [ML-Powered Intelligence](2025-11-22-ml-powered-intelligence)