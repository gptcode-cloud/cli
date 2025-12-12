# Universal Feedback Capture (Two Keystrokes)

## Overview
Capture feedback from any CLI with two keystrokes:
- Press Ctrl+g to mark the current line as the suggested command
- Press Enter to run; the hook automatically records whether it was **good** (same) or **bad** (different) and saves:
  - wrong_response / correct_response
  - changed files (git diff --name-only)
  - diff_path with the full patch (optional)

## Installation

### Requirements
- zsh, bash, or fish
- `chu` installed (mise or local binary)

### Install the hook
```bash
# zsh (with diff)
gptcode feedback hook install --with-diff

# bash
gptcode feedback hook install --shell=bash --with-diff

# fish
gptcode feedback hook install --shell=fish --with-diff
```

This creates and references a hook at `~/.gptcode/feedback_hook.<shell>` and updates your shell rc.

## Usage

1) Type or paste the suggested command on the terminal line
2) Press **Ctrl+g** to mark the suggestion
3) Edit if needed and press **Enter**

The hook compares what you ran with the suggestion:
- Same → `good` with `correct = command`
- Different → `bad` with `wrong = suggestion`, `correct = command`

If the directory is a git repo:
- `files` contains `git diff --name-only`
- With `--with-diff`, the full patch is saved to `~/.gptcode/diffs/<timestamp>.patch` and the path is stored in `diff_path`.

## Generate the demos via CLI
```bash
gptcode demo feedback create           # also available as: `gptcode demo feedback:create` or `gptcode demo feedback.create`
```

## Check events
```bash
gptcode feedback stats
```

## Manual/programmatic submission (optional)
```bash
gptcode feedback submit \
  --sentiment=bad --kind=command --source=shell --agent=editor \
  --task='open Elixir console on Fly.io' \
  --wrong='fly ssh console --exec "iex -S mix"' \
  --correct='fly ssh console --pty -C "/app/bin/platform remote"' \
  --files fly.toml --files deploy.sh --capture-diff
```

## Integrating your own UIs/CLIs
If your UI suggests commands, the user can simply press **Ctrl+g** before running. No app changes needed.

## Troubleshooting
- "Ctrl+g does nothing": reload your shell rc (`source ~/.zshrc` or equivalent). In zsh, confirm the binding with `bindkey | grep chu_mark_suggestion_widget`.
- "Option+S types ß": use **Ctrl+g**. Alt/Option depends on keyboard layout.
- "Files not captured": make sure you're inside a git repo (`git rev-parse --is-inside-work-tree`).
