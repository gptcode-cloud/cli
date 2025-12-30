---
layout: post
title: "What We Actually Built: CLI Workflows + Live Dashboard"
date: 2025-12-30
author: Jader Correa
tags: [cli, dashboard, workflow, release]
---

# What We Actually Built: CLI Workflows + Live Dashboard

This is a summary of the features that shipped in December. No marketing speak, just what's there and how to use it.

## The Workflow Commands

We added three commands that actually work:

### `gptcode plan`

Creates a markdown plan before you start coding. Why? Because jumping straight into implementation usually means you'll miss something.

```bash
gptcode plan "Add user authentication with JWT"
```

This gives you:
- Current state analysis (what exists)
- What we're building (and what we're NOT doing)
- Phases with success criteria
- Verification steps

The plan lives in `~/.gptcode/plans/`. You can edit it before implementing.

### `gptcode implement`

Takes a plan file and executes it phase by phase:

```bash
gptcode implement ~/.gptcode/plans/auth-plan.md
```

What makes this different:
- **Checkpoints**: If it fails at phase 3, you can resume from phase 3
- **Verification**: Runs tests/build after each phase
- **Interactive mode**: Asks before making changes (default)

### `gptcode do`

End-to-end autonomous execution. Give it a task, it figures out the rest:

```bash
gptcode do "Fix the failing test in auth_test.go"
```

This uses the Symphony pattern internally:
1. Analyzes task complexity
2. Decomposes into movements (steps)
3. Executes with checkpointing
4. Verifies results

**Honest assessment**: Works well for focused tasks (fix this bug, add this function). Complex refactors still need human oversight.

## The Maestro System

Under the hood, there's an orchestrator called Maestro that coordinates everything:

```
Task → Analyzer → Planner → Executor → Verifier
                     ↓
              Checkpoint/Recovery
```

Key features:
- **Dynamic verifier selection**: Only runs Go tests if you modified `.go` files
- **Model selection**: Routes to cheaper models for simple tasks
- **Budget tracking**: Stops if you're about to exceed your limit

## Live Dashboard

The web dashboard (`/live`) now shows real-time telemetry from CLI sessions.

### What you can see:
- Active sessions with token/cost tracking
- File tree explorer for the current project
- Context sync between CLI and browser

### Context Sync

This is the useful part. You can edit `.gptcode/context/` in the CLI and see it in the dashboard (and vice versa):

```bash
# CLI updates context
echo "## New Architecture\nMigrated to microservices" >> .gptcode/context/shared.md

# Dashboard shows it immediately
# Edit in browser → CLI picks up changes
```

### E2E Encryption

If you're working on sensitive code, enable Private Mode:

```bash
gptcode context live --private
```

The server becomes a blind relay. It sees encrypted blobs, nothing readable. Uses X25519 + ChaCha20-Poly1305 (same as Signal).

## How to Try It

```bash
# Update CLI
go install github.com/gptcode-cloud/cli/cmd/gptcode@latest

# Try the workflow
gptcode plan "Add rate limiting to API endpoints"
gptcode implement ~/.gptcode/plans/rate-limiting-plan.md

# Or go autonomous
gptcode do "Add a health check endpoint"

# Run Live Dashboard (requires Elixir)
cd live && mix deps.get && mix phx.server
```

## What's Still Rough

Being honest:
- Symphony movement quality varies. Simple tasks: good. Complex refactors: hit or miss.
- Live Dashboard UX needs work. It's functional, not pretty.
- Context sync occasionally has race conditions.

Check the [E2E test results](https://github.com/gptcode-cloud/cli/actions) for current pass rates.

## What's Next

- Improve Symphony decomposition (targeting higher E2E pass rate)
- Better cost tracking granularity
- Dashboard UX improvements
- More verifiers (lint, type-check)

---

Questions or issues? [Open an issue](https://github.com/gptcode-cloud/cli/issues) or [join the discussion](https://github.com/gptcode-cloud/cli/discussions).
