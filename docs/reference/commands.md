---
title: Commands Reference
description: Complete reference for all GPTCode CLI commands
---

# Commands Reference

Complete guide to all `chu` commands and their usage.

## Quick Navigation

<div class="command-nav">
  <a href="#setup-commands">Setup</a>
  <a href="#interactive-modes">Interactive</a>
  <a href="#workflow-commands-research--plan--implement">Workflow</a>
  <a href="#code-quality">Review</a>
  <a href="#feature-generation">Features</a>
  <a href="#execution-mode">Run</a>
  <a href="#machine-learning-commands">ML</a>
  <a href="#dependency-graph-commands">Graph</a>
  <a href="#configuration">Config</a>
</div>

<style>
.command-nav {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  padding: 1rem;
  background: #16161e;
  border-radius: 8px;
  margin-bottom: 2rem;
  border: 1px solid #3b4261;
}
.command-nav a {
  padding: 0.5rem 1rem;
  background: #1a1b26;
  border: 1px solid #3b4261;
  border-radius: 4px;
  text-decoration: none;
  color: #c0caf5;
  font-weight: 500;
  transition: all 0.2s;
}
.command-nav a:hover {
  background: #7aa2f7;
  color: #1a1b26;
  border-color: #7aa2f7;
}
.copy-btn {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  padding: 0.25rem 0.5rem;
  background: #4a90e2;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.75rem;
  opacity: 0;
  transition: opacity 0.2s;
}
.highlight:hover .copy-btn {
  opacity: 1;
}
.copy-btn:hover {
  background: #357abd;
}
.highlight {
  position: relative;
}
</style>

<script>
document.addEventListener('DOMContentLoaded', function() {
  document.querySelectorAll('pre code').forEach(function(codeBlock) {
    const button = document.createElement('button');
    button.className = 'copy-btn';
    button.textContent = 'Copy';
    button.addEventListener('click', function() {
      navigator.clipboard.writeText(codeBlock.textContent).then(function() {
        button.textContent = 'Copied!';
        setTimeout(function() {
          button.textContent = 'Copy';
        }, 2000);
      });
    });
    codeBlock.parentElement.appendChild(button);
  });
});
</script>

---

## `gptcode do` - Autonomous Execution

**The flagship copilot command.** Orchestrates 4 specialized agents to autonomously complete tasks with validation and auto-retry.

### How It Works

```bash
gptcode do "add JWT authentication"
```

**Agent Flow:**
1. **Analyzer** - Understands codebase using dependency graph, reads relevant files
2. **Planner** - Creates minimal implementation plan, lists files to modify
3. **File Validation** - Extracts allowed files, blocks extras
4. **Editor** - Executes changes ONLY on planned files
5. **Validator** - Checks success criteria, triggers auto-retry if validation fails

### Examples

```bash
gptcode do "fix authentication bug in login handler"
gptcode do "refactor error handling to use custom types"
gptcode do "add rate limiting to API endpoints" --supervised
gptcode do "optimize database queries" --interactive
```

### Flags

- `--supervised` - Require manual approval before implementation (critical tasks)
- `--interactive` - Prompt when model selection is ambiguous
- `--dry-run` - Show plan only, don't execute
- `-v` / `--verbose` - Show model selection and agent decisions
- `--max-attempts N` - Maximum retry attempts (default: 3)

### Benefits

- Automatic model selection: queries performance history and picks the best model per agent  
- Auto-retry with feedback: switches to better models when validation fails  
- File validation: prevents creating unintended files or modifying wrong code  
- Success criteria: verifies task completion before finishing  
- Cost optimized: uses cheaper models for routing, better models for editing  

[See full agent architecture →](/features#agent-based-architecture)

---

## Setup Commands

### `gptcode setup`

Initialize GPTCode configuration at `~/.gptcode`.

```bash
gptcode setup
```

Creates:
- `~/.gptcode/profile.yaml` – backend and model configuration
- `~/.gptcode/system_prompt.md` – base system prompt
- `~/.gptcode/memories.jsonl` – memory store for examples

### `gptcode key [backend]`

Add or update API key for a backend provider.

```bash
gptcode key openrouter
gptcode key groq
```

### `gptcode models update`

Update model catalog from available providers (OpenRouter, Groq, OpenAI, etc.).

```bash
gptcode models update
```

---

## Interactive Modes

### `gptcode chat`

Code-focused conversation mode. Routes queries to appropriate agents based on intent.

```bash
gptcode chat
gptcode chat "explain how authentication works"
echo "list go files" | gptcode chat
```

**Agent routing:**
- `query` – read/understand code
- `edit` – modify code
- `research` – external information
- `test` – run tests or commands
- `review` – code review and critique

### `gptcode tdd`

Incremental TDD mode. Generates tests first, then implementation.

```bash
gptcode tdd
gptcode tdd "slugify function with unicode support"
```

Workflow:
1. Clarify requirements
2. Generate tests
3. Generate implementation
4. Iterate and refine

---

## Workflow Commands (Research → Plan → Implement)

### `gptcode research [question]`

Document codebase and understand architecture.

```bash
gptcode research "How does authentication work?"
gptcode research "Explain the payment flow"
```

Creates a research document with findings and analysis.

### `gptcode plan [task]`

Create detailed implementation plan with phases.

```bash
gptcode plan "Add user authentication"
gptcode plan "Implement webhook system"
```

Generates:
- Problem statement
- Current state analysis
- Proposed changes with phases
- Saved to `~/.gptcode/plans/`

### `gptcode implement <plan_file>`

Execute an approved plan phase-by-phase with verification.

```bash
gptcode implement ~/.gptcode/plans/2025-01-15-add-auth.md
```

Each phase:
1. Implemented
2. Verified (tests run)
3. User confirms before next phase

---

## Code Quality

### `gptcode review [target]`

**NEW**: Review code for bugs, security issues, and improvements against coding standards.

```bash
gptcode review main.go
gptcode review ./src
gptcode review .
gptcode review internal/agents/ --focus security
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
gptcode review main.go --focus "error handling"
gptcode review . --focus performance
gptcode review src/auth --focus security
```

---

## Feature Generation

### `gptcode feature [description]`

Generate tests + implementation with auto-detected language.

```bash
gptcode feature "slugify with unicode support and max length"
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

### `gptcode run [task]`

Execute tasks with follow-up support. Two modes available:

**1. AI-assisted mode (default when no args provided):**
```bash
gptcode run                                    # Start interactive session
gptcode run "deploy to staging" --once         # Single AI execution
```

**2. Direct REPL mode with command history:**
```bash
gptcode run --raw                              # Interactive command execution
gptcode run "docker ps" --raw                  # Execute command and exit
```

#### AI-Assisted Mode

Provides intelligent command suggestions and execution:
- Command history tracking
- Output reference ($1, $2, $last)
- Directory and environment management
- Context preservation across commands

```bash
gptcode run
> deploy to staging
[AI executes fly deploy command]
> check if it's running
[AI executes status check]
> roll back if there are errors
[AI conditionally executes rollback]
```

#### Direct REPL Mode

Direct shell command execution with enhanced features:
```bash
gptcode run --raw
> ls -la
> cat $1                    # Reference previous command
> /history                  # Show command history
> /output 1                 # Show output of command 1
> /cd /tmp                  # Change directory
> /env MY_VAR=value         # Set environment variable
> /exit                     # Exit REPL
```

**REPL Commands:**
- `/exit`, `/quit` - Exit run session
- `/help` - Show available commands
- `/history` - Show command history
- `/output <id>` - Show output of previous command
- `/cd <dir>` - Change working directory
- `/env [key[=value]]` - Show/set environment variables

**Command References:**
- `$last` - Reference the last command
- `$1`, `$2`, ... - Reference command by ID

#### Examples

```bash
# AI-assisted operational tasks
gptcode run "check postgres status"
gptcode run "make GET request to api.github.com/users/octocat"

# Direct command REPL for DevOps
gptcode run --raw
> docker ps
> docker logs $1            # Reference container from previous output
> /cd /var/log
> tail -f app.log

# Single-shot with piped input
echo "deploy to production" | gptcode run --once
```

#### Flags

- `--raw` - Use direct command REPL mode (no AI)
- `--once` - Force single-shot mode (backwards compatible)

Perfect for operational tasks, DevOps workflows, and command execution with history.

---

## Machine Learning Commands

### `gptcode ml list`

List available ML models.

```bash
gptcode ml list
```

Shows:
- Model name
- Description
- Location
- Status (trained/not trained)

### `gptcode ml train <model>`

Train an ML model using Python.

```bash
gptcode ml train complexity
gptcode ml train intent
```

**Available models:**
- `complexity` – Task complexity classifier (simple/complex/multistep)
- `intent` – Intent classifier (query/editor/research/review)

**Requirements:**
- Python 3.8+
- Will create venv and install dependencies automatically

### `gptcode ml test <model> [query]`

Test a trained model with a query.

```bash
gptcode ml test complexity "implement oauth"
gptcode ml test intent "explain this code"
```

Shows prediction and probabilities for all classes.

### `gptcode ml eval <model> [-f file]`

Evaluate model performance on test dataset.

```bash
gptcode ml eval complexity
gptcode ml eval intent -f ml/intent/data/eval.csv
```

Shows:
- Accuracy
- Precision/Recall/F1 per class
- Confusion matrix
- Low-confidence predictions

### `gptcode ml predict [model] <text>`

Make prediction using embedded Go model (no Python runtime).

```bash
gptcode ml predict "implement auth"               # uses complexity (default)
gptcode ml predict complexity "fix typo"          # explicit model
gptcode ml predict intent "explain this code"     # intent classification
```

**Fast path:**
- 1ms inference (vs 500ms LLM)
- Zero API cost
- Pure Go, no Python runtime

---

## ML Configuration

### Complexity Threshold

Controls when Guided Mode is automatically activated.

```bash
# View current threshold (default: 0.55)
gptcode config get defaults.ml_complex_threshold

# Set threshold (0.0-1.0)
gptcode config set defaults.ml_complex_threshold 0.6
```

Higher threshold = less sensitive (fewer Guided Mode triggers)

### Intent Threshold

Controls when ML router is used instead of LLM.

```bash
# View current threshold (default: 0.7)
gptcode config get defaults.ml_intent_threshold

# Set threshold (0.0-1.0)
gptcode config set defaults.ml_intent_threshold 0.8
```

Higher threshold = more LLM fallbacks (more accurate but slower/expensive)

---

## Dependency Graph Commands

### `gptcode graph build`

Force rebuild dependency graph, ignoring cache.

```bash
gptcode graph build
```

Shows:
- Number of nodes (files)
- Number of edges (dependencies)
- Build time

**When to use:**
- After major refactoring
- After adding/removing many files
- If cache seems stale

### `gptcode graph query <terms>`

Find relevant files for a query term using PageRank.

```bash
gptcode graph query "authentication"
gptcode graph query "database connection"
gptcode graph query "api routes"
```

Shows:
- Matching files ranked by importance
- PageRank scores
- Why each file was selected

**How it works:**
1. Keyword matching in file paths
2. Neighbor expansion (imports/imported-by)
3. PageRank weighting
4. Top N selection

---

## Graph Configuration

### Max Files

Control how many files are added to context in chat mode.

```bash
# View current setting (default: 5)
gptcode config get defaults.graph_max_files

# Set max files (1-20)
gptcode config set defaults.graph_max_files 8
```

**Recommendations:**
- Small projects (<50 files): 3-5
- Medium projects (50-500 files): 5-8
- Large projects (500+ files): 8-12

### Debug Graph

```bash
export CHUCHU_DEBUG=1
gptcode chat "your query"  # Shows graph stats
```

Shows:
- Nodes/edges count
- Selected files
- PageRank scores
- Build time (cache hit/miss)

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

## Backend Management

### `gptcode backend`

Show current backend.

```bash
gptcode backend
```

Shows:
- Backend name
- Type (openai/ollama)
- Base URL
- Default model

### `gptcode backend list`

List all configured backends.

```bash
gptcode backend list
```

### `gptcode backend show [name]`

Show backend configuration. Shows current if no name provided.

```bash
gptcode backend show groq
```

Shows:
- Type and URL
- Default model
- All configured models

### `gptcode backend use <name>`

Switch to a backend.

```bash
gptcode backend use groq
gptcode backend use openrouter
gptcode backend use ollama
```

### `gptcode backend create`

Create a new backend.

```bash
gptcode backend create mygroq openai https://api.groq.com/openai/v1
gptcode key mygroq  # Set API key
gptcode config set backend.mygroq.default_model llama-3.3-70b-versatile
gptcode backend use mygroq
```

### `gptcode backend delete`

Delete a backend.

```bash
gptcode backend delete mygroq
```

---

## Profile Management

### `gptcode profile`

Show current profile.

```bash
gptcode profile
```

Shows:
- Backend and profile name
- Agent models (router, query, editor, research)

### `gptcode profile list [backend]`

List all profiles. Optionally filter by backend.

```bash
gptcode profile list              # All profiles
gptcode profile list groq        # Only groq profiles
```

Shows profiles in `backend.profile` format.

### `gptcode profile show [backend.profile]`

Show profile configuration. Shows current if not specified.

```bash
gptcode profile show groq.speed
gptcode profile show              # Current profile
```

### `gptcode profile use <backend>.<profile>`

Switch to a backend and profile in one command.

```bash
gptcode profile use groq.speed
gptcode profile use openrouter.free
gptcode profile use ollama.local
```

**Benefits:**
- Faster than switching backend and profile separately
- Atomic operation (both or neither)
- Easier to remember

### Advanced Profile Commands

For creating and configuring profiles, use `gptcode profiles` (plural):

```bash
# Create new profile
gptcode profiles create groq speed

# Configure agents
gptcode profiles set-agent groq speed router llama-3.1-8b-instant
gptcode profiles set-agent groq speed query llama-3.1-8b-instant
gptcode profiles set-agent groq speed editor llama-3.1-8b-instant
gptcode profiles set-agent groq speed research llama-3.1-8b-instant
```

---

## Environment Variables

### `CHUCHU_DEBUG`

Enable debug output to stderr.

```bash
CHUCHU_DEBUG=1 gptcode chat
```

Shows:
- Agent routing decisions
- Iteration counts
- Tool execution details

---

## Configuration

All configuration lives in `~/.gptcode/`:

```
~/.gptcode/
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

## Advanced Configuration

### Direct Config Manipulation

For advanced users who need direct access to configuration values:

```bash
# Get configuration value
gptcode config get defaults.backend
gptcode config get defaults.profile
gptcode config get backend.groq.default_model

# Set configuration value
gptcode config set defaults.backend groq
gptcode config set defaults.profile speed
gptcode config set backend.groq.default_model llama-3.3-70b-versatile
```

**Note:** For most use cases, prefer the user-friendly commands:
- `gptcode backend use` instead of `gptcode config set defaults.backend`
- `gptcode profile use` instead of `gptcode config set defaults.profile`

---

## Next Steps

- See [Research Mode](./research.html) for workflow details
- See [Plan Mode](./plan.html) for plan structure
