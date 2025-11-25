---
title: Commands Reference
description: Complete reference for all Chuchu CLI commands
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
  background: #f5f5f5;
  border-radius: 8px;
  margin-bottom: 2rem;
}
.command-nav a {
  padding: 0.5rem 1rem;
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  text-decoration: none;
  color: #333;
  font-weight: 500;
  transition: all 0.2s;
}
.command-nav a:hover {
  background: #4a90e2;
  color: white;
  border-color: #4a90e2;
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

## `chu do` - Autonomous Execution

**The flagship copilot command.** Orchestrates 4 specialized agents to autonomously complete tasks with validation and auto-retry.

### How It Works

```bash
chu do "add JWT authentication"
```

**Agent Flow:**
1. **Analyzer** - Understands codebase using dependency graph, reads relevant files
2. **Planner** - Creates minimal implementation plan, lists files to modify
3. **File Validation** - Extracts allowed files, blocks extras
4. **Editor** - Executes changes ONLY on planned files
5. **Validator** - Checks success criteria, triggers auto-retry if validation fails

### Examples

```bash
chu do "fix authentication bug in login handler"
chu do "refactor error handling to use custom types"
chu do "add rate limiting to API endpoints" --supervised
chu do "optimize database queries" --interactive
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

## Machine Learning Commands

### `chu ml list`

List available ML models.

```bash
chu ml list
```

Shows:
- Model name
- Description
- Location
- Status (trained/not trained)

### `chu ml train <model>`

Train an ML model using Python.

```bash
chu ml train complexity
chu ml train intent
```

**Available models:**
- `complexity` – Task complexity classifier (simple/complex/multistep)
- `intent` – Intent classifier (query/editor/research/review)

**Requirements:**
- Python 3.8+
- Will create venv and install dependencies automatically

### `chu ml test <model> [query]`

Test a trained model with a query.

```bash
chu ml test complexity "implement oauth"
chu ml test intent "explain this code"
```

Shows prediction and probabilities for all classes.

### `chu ml eval <model> [-f file]`

Evaluate model performance on test dataset.

```bash
chu ml eval complexity
chu ml eval intent -f ml/intent/data/eval.csv
```

Shows:
- Accuracy
- Precision/Recall/F1 per class
- Confusion matrix
- Low-confidence predictions

### `chu ml predict [model] <text>`

Make prediction using embedded Go model (no Python runtime).

```bash
chu ml predict "implement auth"               # uses complexity (default)
chu ml predict complexity "fix typo"          # explicit model
chu ml predict intent "explain this code"     # intent classification
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
chu config get defaults.ml_complex_threshold

# Set threshold (0.0-1.0)
chu config set defaults.ml_complex_threshold 0.6
```

Higher threshold = less sensitive (fewer Guided Mode triggers)

### Intent Threshold

Controls when ML router is used instead of LLM.

```bash
# View current threshold (default: 0.7)
chu config get defaults.ml_intent_threshold

# Set threshold (0.0-1.0)
chu config set defaults.ml_intent_threshold 0.8
```

Higher threshold = more LLM fallbacks (more accurate but slower/expensive)

---

## Dependency Graph Commands

### `chu graph build`

Force rebuild dependency graph, ignoring cache.

```bash
chu graph build
```

Shows:
- Number of nodes (files)
- Number of edges (dependencies)
- Build time

**When to use:**
- After major refactoring
- After adding/removing many files
- If cache seems stale

### `chu graph query <terms>`

Find relevant files for a query term using PageRank.

```bash
chu graph query "authentication"
chu graph query "database connection"
chu graph query "api routes"
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
chu config get defaults.graph_max_files

# Set max files (1-20)
chu config set defaults.graph_max_files 8
```

**Recommendations:**
- Small projects (<50 files): 3-5
- Medium projects (50-500 files): 5-8
- Large projects (500+ files): 8-12

### Debug Graph

```bash
export CHUCHU_DEBUG=1
chu chat "your query"  # Shows graph stats
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
