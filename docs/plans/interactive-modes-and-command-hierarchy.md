# Interactive Modes and Command Hierarchy - Implementation Plan

**Status:** In Progress  
**Created:** 2025-11-25  
**Priority:** High (Fase 1), Medium-Low (Fases 2-3)

## Overview

Restructure Chuchu's command hierarchy to highlight `chu do` as the flagship feature, improve documentation to emphasize agent-based architecture and validation, and prepare foundation for interactive modes (chat REPL and run follow-up).

## Motivation

### Problems
1. ❌ `chu chat` is single-shot (no conversation context)
2. ❌ `chu run` executes and exits (no iteration/follow-up)
3. ❌ Documentation doesn't emphasize `chu do` as main feature
4. ❌ CLI help has duplicates (models/model, feedback 2x, graph 2x)
5. ❌ TDD-first messaging overshadows agent architecture + validation

### Goals
- 🎯 Position `chu do` as the autonomous copilot (flagship)
- 🎯 Organize commands into logical categories
- 🎯 Align CLI help with documentation hierarchy
- 🎯 Prepare infrastructure for interactive features
- 🎯 Emphasize: **Agents + Validation** not just TDD

---

## Proposed Command Hierarchy

### **Tier 1: Core Copilot** ⭐
**`chu do` - FLAGSHIP**
- Autonomous execution: Analyzer → Planner → Editor → Validator
- Auto-retry with model switching
- File validation (no extra files)
- Success criteria checking

Flags:
- `--supervised` - Manual approval before implementation
- `--interactive` - Prompt when model selection is ambiguous
- `--dry-run` - Show plan only
- `-v` - Verbose (show model selection)
- `--max-attempts N` - Max retry attempts (default 3)

### **Tier 2: Interactive Modes** 💬
**`chu chat` - Conversational**
- Currently: Single-shot Q&A
- **Planned (Fase 2):** REPL with conversation context

**`chu run` - Execute with follow-up**
- Currently: Execute and exit
- **Planned (Fase 2):** Iterate after command execution

### **Tier 3: Manual Workflow** 🔧
- `chu research` - Analyze codebase
- `chu plan` - Create plan without executing
- `chu implement` - Execute existing plan

### **Tier 4: Specialized Tools** 🎯
- `chu tdd` - Test-Driven Development
- `chu feature` - Feature generation
- `chu review` - Code review

### **Tier 5: Management** ⚙️
- `chu model` - Model management
- `chu config` - Configuration
- `chu backend` - Backend setup
- `chu feedback` - User feedback tracking
- `chu ml` - Machine learning
- `chu graph` - Dependency graph

---

## Implementation Phases

### ✅ **Fase 1: CLI Help & Documentation** (IMMEDIATE)

**Target:** This PR/commit

#### 1.1 Update CLI Help (cmd/chu/main.go)
- [x] Change Short: `"Chuchu – AI Coding Assistant with Specialized Agents"`
- [x] Rewrite Long help with categories:
  ```
  ## COPILOT (Autonomous)
  ## INTERACTIVE (Conversational)
  ## WORKFLOW (Manual Control)
  ## SPECIALIZED TOOLS
  ## MODEL MANAGEMENT
  ## CONFIGURATION
  ## ADVANCED
  ```
- [x] Remove duplicate: `modelsCmd` (line 94)
- [x] Verify no duplicates: feedbackCmd, graphCmd

#### 1.2 Homepage (docs/index.md)
- [x] Hero section: Emphasize agents + validation
- [x] Update tagline: "Autonomous execution with validation"
- [x] Reorder features:
  1. 🚀 Autonomous Execution (`chu do`)
  2. 💬 Interactive Modes
  3. 🔧 Manual Workflow
  4. 🤖 Model Selection
  5. ✅ Validation
  6. 💰 Cost
- [x] Add agent flow diagram

#### 1.3 Features Page (docs/features.md)
- [x] Restructure sections:
  1. Agent-Based Architecture
  2. Validation & Safety
  3. Intelligence Features
  4. Cost Optimization
  5. Developer Experience

#### 1.4 Commands Page (docs/commands.md)
- [x] Add highlighted section for `chu do` at top
- [x] Categorize remaining commands
- [x] Examples for each category

**Success Criteria:**
- ✅ CLI help mentions `chu do` first
- ✅ Commands organized in clear categories
- ✅ No duplicates
- ✅ Documentation emphasizes agents + validation
- ✅ Consistent messaging across CLI and docs

---

### ⏳ **Fase 2: Interactive Chat Mode** (FUTURE PR)

**Priority:** Medium  
**Estimated Effort:** 3-5 days

#### Goals
Transform `chu chat` from single-shot to REPL with conversation context.

#### Current Behavior
```bash
chu chat "explain this function"
# returns answer and exits
```

#### Desired Behavior
```bash
chu chat
> explain this function
[AI response]
> now optimize it
[AI continues with context]
> what about edge cases?
[AI remembers prior conversation]
```

#### Implementation Tasks

**2.1 Create REPL Infrastructure**
- [ ] Add readline library dependency
- [ ] Create `internal/repl/chat_repl.go`
  - Input loop
  - History management
  - Command parsing (`/exit`, `/clear`, `/save`, `/help`)
- [ ] Message history struct with context window management

**2.2 Context Management**
- [ ] Store conversation history (sliding window)
- [ ] File context integration (read current directory files)
- [ ] Context size tracking (warn when approaching limits)
- [ ] `/clear` command to reset context

**2.3 Commands**
- `/exit` or `/quit` - Exit chat
- `/clear` - Clear conversation history
- `/save <filename>` - Save conversation to file
- `/help` - Show available commands
- `/context` - Show current context size
- `/files` - List files in context

**2.4 Update cmd/chu/chat.go**
- [ ] Detect if args provided (single-shot mode for compatibility)
- [ ] Default to REPL mode when no args
- [ ] Flag `--once` to force single-shot behavior

**Files to Create/Modify:**
- `internal/repl/chat_repl.go` (new)
- `internal/repl/context_manager.go` (new)
- `cmd/chu/chat.go` (modify)
- `go.mod` (add readline dependency)

**Success Criteria:**
- User can start `chu chat` and have multi-turn conversation
- Context is maintained across turns
- `/exit` cleanly exits
- `/clear` resets conversation
- `chu chat "quick question"` still works (backwards compat)

---

### ⏳ **Fase 3: Run Follow-up Mode** (FUTURE PR)

**Priority:** Medium-Low  
**Estimated Effort:** 2-4 days

#### Goals
Allow iteration/follow-up after `chu run` executes commands.

#### Current Behavior
```bash
chu run "check postgres status"
# shows output and exits
```

#### Desired Behavior
```bash
chu run "check postgres status"
[command output]
> now restart it
[executes restart]
> verify it's running
[checks status again]
```

#### Implementation Tasks

**3.1 Command Execution Context**
- [ ] Store command history in session
- [ ] Capture stdout/stderr of executed commands
- [ ] Track exit codes
- [ ] Environment state tracking (cwd, env vars)

**3.2 REPL for Run Mode**
- [ ] Reuse REPL infrastructure from Fase 2
- [ ] Add command context to prompts
- [ ] Allow referencing previous outputs (`$last`, `$1`, `$2`)
- [ ] Shell variable tracking

**3.3 Commands**
- `/history` - Show command history
- `/output N` - Show output of command N
- `/cd <dir>` - Change directory for next commands
- `/env <var>=<value>` - Set environment variable
- `/exit` - Exit run session

**3.4 Update cmd/chu/run.go**
- [ ] Add `--once` flag for single-shot (current behavior)
- [ ] Default to interactive mode
- [ ] Session persistence option

**Files to Create/Modify:**
- `internal/repl/run_repl.go` (new)
- `internal/repl/command_history.go` (new)
- `cmd/chu/run.go` (modify)

**Success Criteria:**
- User can execute command and provide follow-ups
- Previous command outputs are accessible
- Working directory changes persist
- `chu run "one-off command" --once` still works

---

## Documentation Updates

### index.md (Homepage)
```markdown
# Chuchu
## AI Coding Assistant with Specialized Agents

Autonomous execution with validation.  
**Analyzer → Planner → Editor → Validator**

$0-5/month vs $20-30/month subscriptions.

[Get Started](#installation) [GitHub](https://github.com/jadercorrea/chuchu)
```

### commands.md Structure
```markdown
# Commands Reference

## chu do - Autonomous Execution ⭐

The flagship copilot command...

### Examples
### How it Works
### Flags
### Benefits

---

## Interactive Modes

### chu chat
### chu run

---

## Manual Workflow
...
```

---

## Testing Strategy

### Fase 1 (Manual)
- [ ] `chu --help` shows new categories
- [ ] No duplicate commands
- [ ] Documentation reads coherently
- [ ] Agent flow diagram renders correctly

### Fase 2 (Chat REPL)
- [ ] Multi-turn conversation maintains context
- [ ] `/clear` resets properly
- [ ] `/save` exports conversation
- [ ] Backwards compatible with `chu chat "question"`

### Fase 3 (Run Follow-up)
- [ ] Command outputs are captured
- [ ] Follow-up commands execute in same context
- [ ] `/history` shows all commands
- [ ] `--once` flag works for scripts

---

## Success Metrics

### Quantitative
- CLI help mentions `chu do` in first 3 lines
- Documentation has agent diagram
- Zero duplicate commands in help
- `chu chat` session >5 turns without breaking

### Qualitative
- Users understand `chu do` is the main feature
- Clear distinction between autonomous/interactive/manual modes
- Documentation emphasizes validation + agents over TDD
- Consistent messaging across all touchpoints

---

## Dependencies

### Fase 1
- None (documentation + CLI text changes)

### Fase 2
- Go readline library (e.g. `github.com/chzyer/readline`)
- Context window tracking
- Message history storage

### Fase 3
- Command output buffering
- Session state management
- REPL infrastructure from Fase 2

---

## Open Questions

1. **Chat persistence:** Should conversations be auto-saved?
2. **Run sessions:** Should sessions be resumable across terminal restarts?
3. **Model selection in interactive mode:** Use same model throughout or switch per turn?
4. **Context limits:** Hard limit or soft warning when approaching max tokens?

---

## Related Work

- Existing REPL implementation in other tools (ipython, node REPL)
- Cursor's chat interface (inspiration for UX)
- GitHub Copilot CLI (command follow-up patterns)

---

## Rollout Plan

1. **Fase 1:** Merge to main immediately (no breaking changes)
2. **Fase 2:** Feature branch → beta testing → main (v0.x.0 bump)
3. **Fase 3:** Same as Fase 2 (can be combined if ready simultaneously)

---

## Notes

- Keep backwards compatibility for scripting use cases
- Interactive modes should be opt-in initially
- Document behavior changes in CHANGELOG
- Update blog posts after each phase
