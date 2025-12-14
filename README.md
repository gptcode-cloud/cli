# 🐺 GPTCode

[![CI](https://github.com/jadercorrea/gptcode/actions/workflows/ci.yml/badge.svg)](https://github.com/jadercorrea/gptcode/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jadercorrea/gptcode)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub Issues](https://img.shields.io/github/issues/jadercorrea/gptcode)](https://github.com/jadercorrea/gptcode/issues)

GPTCode (pronounced "shoo-shoo", Brazilian slang for something small and cute) is a command-line AI coding assistant that helps you write better code through Test-Driven Development—without breaking the bank.

## Why GPTCode?

**GPTCode isn't trying to beat Cursor or Copilot. It's trying to be different.**

- **Transparent**: When it breaks, you can read and fix the code
- **Hackable**: Don't like something? Change it—it's just Go
- **Model Agnostic**: Switch LLMs in 2 minutes (Groq, Ollama, OpenAI, etc)
- **Honest**: E2E tests at 55% (not "95% accuracy" marketing)
- **Affordable**: $2-5/month (Groq) or **$0/month** (Ollama)

**Not better. Different. Yours.**

[Read the full positioning →](https://jadercorrea.github.io/gptcode/blog/2025-12-06-why-gptcode-isnt-trying-to-beat-anyone) | [Original vision →](https://jadercorrea.github.io/gptcode/blog/2025-11-13-why-gptcode)

## Features

### Core Capabilities
- 🤖 **Autonomous Copilot** (`chu do`) - Full task execution with agent orchestration
- 💬 **Interactive Chat** - Code-focused conversations with context awareness
- 🔄 **Structured Workflow** - Research → Plan → Implement with verification
- 🧪 **TDD Mode** - Test-driven development with auto-generated tests
- 🔍 **Code Review** - Automated bug detection, security analysis, and improvements

### Intelligence & Optimization
- 🧠 **Multi-Agent Architecture** - Specialized agents (Router, Query, Editor, Research)
- 📊 **ML-Powered** - Embedded complexity detection and intent classification (1ms, zero API calls)
- 🗺️ **Dependency Graph** - Smart context selection (5x token reduction)
- 💰 **Cost Optimized** - Mix cheap/free models per agent ($0-5/month vs $20-30/month)
- 📈 **Feedback Learning** - Improves recommendations from user feedback

### Developer Experience
- ⚡ **Profile Management** - Switch between cost/speed/quality configurations
- 🎯 **Model Flexibility** - Groq, Ollama, OpenRouter, OpenAI, Anthropic, DeepSeek
- 🔌 **Neovim Integration** - Deep LSP, Tree-sitter, file navigation
- 🔎 **Model Discovery** - Search, install, and configure 300+ models
- 📚 **Web Research** - Built-in web search and documentation lookup

## Three Ways to Work with GPTCode

### 1. 🤖 Autonomous GitHub Issue Resolution (NEW - MVP Complete!) 🎆

GPTCode can now autonomously resolve GitHub issues end-to-end with 100% MVAA coverage:

```bash
# Complete autonomous workflow
gptcode issue fix 123                    # Fetch issue, find files, implement
gptcode issue commit 123 --auto-fix      # Test, lint, build, auto-fix failures
gptcode issue push 123                   # Create PR, link to issue
gptcode issue ci 42                      # Handle CI failures
gptcode issue review 42                  # Address review comments
# Iterate until approved!
```

**What GPTCode Does Autonomously:**
- ✅ Fetches issue and extracts requirements
- ✅ Finds relevant files (AI-powered)
- ✅ Implements changes (Symphony orchestration)
- ✅ Runs tests and auto-fixes failures
- ✅ Runs linters and auto-fixes issues
- ✅ Checks build, coverage, and security
- ✅ Creates PR and links to issue
- ✅ Handles CI failures with auto-fix
- ✅ Addresses review comments autonomously

**Supported Languages:** Go, TypeScript, Python, Elixir, Ruby  
**[Complete Autonomous Issue Guide →](docs/autonomous-issues.md)**

### 2. 🤖 Autonomous Copilot (Fastest)

Let gptcode handle everything - analysis, planning, execution, and validation:

```bash
gptcode do "add user authentication"
gptcode do "fix bug in payment processing"
gptcode do "add password reset feature" --supervised
```

**Features:**
- Automatic agent orchestration (Query → Plan → Edit → Validate)
- Built-in error recovery and retries
- Language-specific testing (Go, TypeScript, Python, Ruby, Elixir)
- Optional supervision mode (confirm before execution)

### 2. 💬 Interactive Chat (Most Flexible)

Conversational interface for exploration and quick tasks:

```bash
gptcode chat "explain this function"
gptcode chat "add error handling to the database connection"
gptcode chat  # Enter interactive REPL
```

**Features:**
- ML-powered complexity detection (auto-triggers guided mode)
- Smart context selection (dependency graph analysis)
- Follow-up questions and refinement
- Seamless Neovim integration

### 3. 🔄 Structured Workflow (Most Control)

Manual control over each phase for complex changes:

```bash
# 1. Research: Understand your codebase
gptcode research "How does authentication work?"

# 2. Plan: Create detailed implementation steps
gptcode plan "Add password reset feature"

# 3. Implement: Execute with verification
gptcode implement plan.md              # Interactive (step-by-step)
gptcode implement plan.md --auto       # Autonomous (fully automated)
```

**Why use structured workflow:**
- ✅ Review and adjust plans before execution
- ✅ Incremental verification at each step
- ✅ Better for large, complex changes
- ✅ Lower costs through explicit planning

### Special Modes

```bash
gptcode tdd                  # Test-driven development
gptcode feature "user auth"  # Auto-generate tests + implementation
gptcode review               # Code review for current changes
gptcode review path/to/file  # Review specific file
gptcode run "deploy to prod" # Task execution with follow-up
```

**[Complete Workflow Guide](docs/workflow-guide.md)** | **[Autonomous Mode Deep Dive](docs/autonomous-mode.md)**

## Quick Start

### Installation

```bash
# Install via go
go install github.com/jadercorrea/gptcode/cmd/gptcode@latest

# Or build from source
git clone https://github.com/jadercorrea/gptcode
cd gptcode
go install ./cmd/gptcode
```

### Setup

```bash
# Interactive setup wizard
gptcode setup

# For ultra-cheap setup, use Groq (get free key at console.groq.com)
# For free local setup, use Ollama (no API key needed)
```

### Neovim Integration

Add to your Neovim config:

```lua
-- lazy.nvim
{
  dir = "~/workspace/gptcode/neovim",  -- adjust path to your clone
  config = function()
    require("gptcode").setup()
  end,
  keys = {
    { "<C-d>", "<cmd>GPTCodeChat<cr>", desc = "Toggle GPTCode Chat" },
    { "<C-m>", "<cmd>GPTCodeModels<cr>", desc = "Switch Model/Profile" },
    { "<leader>ms", "<cmd>GPTCodeModelSearch<cr>", desc = "Search & Install Models" },
  }
}
```

**Key Features:**
- `<C-d>` - Open/close chat interface
- `<C-m>` - Profile Management
  - Create new profiles
  - Load existing profiles
  - Configure agent models (router, query, editor, research)
  - Show profile details
  - Delete profiles
- `<leader>ms` - Search & Install Models
  - Multi-term search (e.g., "ollama llama3", "groq coding fast")
  - Shows pricing, context window, tags, and installation status (✓)
  - Auto-install Ollama models
  - Set as default or use for current session
- `<leader>ca` - Autonomous Execution (:GPTCodeAuto)
  - Execute implementation plans with verification
  - Shows progress in real-time notifications

### ML-Powered Intelligence (Built-in)

GPTCode embeds two lightweight ML models for instant decision-making with zero external dependencies:

#### 1. Complexity Detection
Automatically triggers Guided Mode for complex/multistep tasks in `gptcode chat`.

**Configuration:**
```bash
# View/set complexity threshold (default: 0.55)
gptcode config get defaults.ml_complex_threshold
gptcode config set defaults.ml_complex_threshold 0.6
```

#### 2. Intent Classifier
Replaces LLM call with 1ms local inference to route requests (query/editor/research/review).

**Benefits:**
- **500x faster**: 1ms vs 500ms LLM latency
- **Cost savings**: Zero API calls for routing
- **Fallback**: Uses LLM if confidence < threshold

**Configuration:**
```bash
# View/set intent threshold (default: 0.7)
gptcode config get defaults.ml_intent_threshold
gptcode config set defaults.ml_intent_threshold 0.8
```

**ML CLI Commands:**
```bash
# List available models
gptcode ml list

# Train models (uses Python)
gptcode ml train complexity
gptcode ml train intent

# Test models
gptcode ml test intent "explain this code"
gptcode ml eval intent -f ml/intent/data/eval.csv

# Pure-Go inference (no Python runtime)
gptcode ml predict "your task"                    # complexity (default)
gptcode ml predict complexity "implement oauth"   # explicit
gptcode ml predict intent "explain this code"     # intent classification
```

#### 3. Dependency Graph + Context Optimization

Automatically analyzes your codebase structure to provide only relevant context to the LLM.

**How it works:**
1. Builds a graph of file dependencies (imports/requires)
2. Runs PageRank to identify central/important files
3. Matches query terms to relevant files
4. Expands to 1-hop neighbors (dependencies + dependents)
5. Provides top 5 most relevant files as context

**Benefits:**
- **5x token reduction**: 100k → 20k tokens (only relevant files)
- **Better responses**: LLM sees focused context, not noise
- **Automatic**: Works transparently in `chu chat`
- **Cached**: Graph rebuilt only when files change

**Supported Languages:**
- Go, Python, JavaScript/TypeScript
- Ruby, Rust

**Control:**
```bash
# Debug mode shows graph stats
GPTCODE_DEBUG=1 gptcode chat "your query"
# [GRAPH] Built graph: 142 nodes, 287 edges
# [GRAPH] Selected 5 files:
# [GRAPH]   1. internal/agents/router.go (score: 0.842)
# [GRAPH]   2. internal/llm/provider.go (score: 0.731)
```

**Example:**
```bash
gptcode chat "fix bug in authentication"
# Without graph: Sends entire codebase (100k tokens)
# With graph: Sends auth.go + user.go + middleware.go + session.go + config.go (18k tokens)
```

## Usage Examples

### Quick Tasks (Autonomous)

```bash
# Let gptcode handle everything
gptcode do "add logging to the API handlers"
gptcode do "create a dockerfile for this project"
gptcode do "fix the failing tests in user_test.go"

# With supervision (confirm before changes)
gptcode do "refactor the authentication module" --supervised
```

### Exploration & Learning (Chat)

```bash
# Ask questions about your codebase
gptcode chat "how does the auth system work?"
gptcode chat "where is user validation happening?"

# Quick fixes and modifications
gptcode chat "add error handling to database connections"
gptcode chat "optimize the query in getUsers"
```

### Complex Changes (Structured Workflow)

```bash
# Step 1: Research the codebase
gptcode research "current authentication implementation"

# Step 2: Create detailed plan
gptcode plan "add OAuth2 support"

# Step 3: Execute plan
gptcode implement docs/plans/oauth2-implementation.md
gptcode implement docs/plans/oauth2-implementation.md --auto  # Fully autonomous
```

### Specialized Workflows

```bash
# Test-driven development
gptcode tdd
# 1. Describe feature
# 2. Generate tests
# 3. Implement
# 4. Refine

# Feature generation (tests + implementation)
gptcode feature "user registration with email verification"

# Code review
gptcode review                # Review staged changes
gptcode review src/auth.go    # Review specific file
gptcode review --full         # Full codebase review

# Task execution with context
gptcode run "deploy to staging"
gptcode run "migrate database"
```

### Advanced Git Operations (NEW! 🎯)

GPTCode provides AI-powered Git operations for complex workflows:

```bash
# Git Bisect - Find which commit introduced a bug
gptcode git bisect v1.0.0 HEAD
# Automatically runs tests on each commit
# Uses LLM to analyze the breaking commit

# Cherry-pick with conflict resolution
gptcode git cherry-pick abc123 def456
# Applies commits with AI-powered conflict resolution

# Smart Rebase
gptcode git rebase main
# Rebases with automatic conflict resolution

# Squash commits with AI-generated message
gptcode git squash HEAD~3
# Squashes last 3 commits
# Generates professional commit message via LLM

# Improve commit messages
gptcode git reword HEAD
# Suggests improved commit message following best practices

# Resolve merge conflicts
gptcode merge resolve
# Detects and resolves all conflicted files
# Uses LLM to merge changes intelligently
```

**Features:**
- ✅ AI-powered conflict resolution
- ✅ Automatic commit message generation
- ✅ Test-based bisect automation
- ✅ Context-aware merge decisions

**[Complete Git Guide →](docs/guides/git-operations.md)**

### Autonomous Execution (Maestro)

**Fully autonomous execution with verification and error recovery:**

```bash
# Execute a plan with automatic verification
gptcode implement docs/plans/my-implementation.md --auto

# With custom retry limit
gptcode implement docs/plans/my-implementation.md --auto --max-retries 5

# Resume from last checkpoint
gptcode implement docs/plans/my-implementation.md --auto --resume

# Enable lint verification (optional)
gptcode implement docs/plans/my-implementation.md --auto --lint
```

**Interactive Mode (default):**
- Prompts for confirmation before each step
- Shows step details and context
- Options: execute, skip, or quit
- On error: choose to continue or stop

**Autonomous Mode (`--auto`):**
- Executes plan steps automatically
- Verifies changes with build + tests (auto-detects language)
- Automatic error recovery with intelligent retry
- Checkpoints after each successful step
- Rollback on failure
- Language support: Go, TypeScript/JavaScript, Python, Elixir, Ruby
- Optional lint verification (golangci-lint, eslint, ruff, rubocop, mix format)

**Neovim Integration:**
```vim
:GPTCodeAuto        " prompts for plan file and runs: gptcode implement <file> --auto
" Or keymap: <leader>ca
```

### Backend Management

```bash
# Show current backend
gptcode backend

# List all backends
gptcode backend list

# Show backend details
gptcode backend show groq

# Switch backend
gptcode backend use groq

# Create new backend
gptcode backend create mygroq openai https://api.groq.com/openai/v1
gptcode key mygroq  # Set API key
gptcode config set backend.mygroq.default_model llama-3.3-70b-versatile
gptcode backend use mygroq

# Delete backend
gptcode backend delete mygroq
```

### Profile Management

```bash
# Show current profile
gptcode profile

# List all profiles
gptcode profile list

# Show profile details
gptcode profile show groq.speed

# Switch profile (backend + profile)
gptcode profile use groq.speed

# Create new profile
gptcode profiles create groq speed

# Configure agents
gptcode profiles set-agent groq speed router llama-3.1-8b-instant
gptcode profiles set-agent groq speed query llama-3.1-8b-instant
```

### Model Discovery & Installation

```bash
# Search for ollama models
gptcode models search ollama llama3

# Search with multiple filters (ANDed)
gptcode models search ollama coding fast

# Install ollama model
gptcode models install llama3.1:8b
```

### Feedback & Learning

```bash
# Record positive feedback
gptcode feedback good --backend groq --model llama-3.3-70b-versatile --agent query

# Record negative feedback
gptcode feedback bad --backend groq --model llama-3.1-8b-instant --agent router

# View statistics
gptcode feedback stats
```

GPTCode learns from your feedback to recommend better models over time.

## Configuration & Profiles

GPTCode supports multiple backends and profiles optimized for different use cases.

### Quick Profile Switching

```bash
# List available profiles
gptcode profile list

# Switch profiles
gptcode profile use ollama.default      # $0/month (local)
gptcode profile use openrouter.free    # $0/month (cloud, rate-limited)
gptcode profile use groq.budget        # ~$0.85/month (3M tokens)
gptcode profile use groq.performance   # ~$2.41/month (3M tokens)

# Show current profile
gptcode profile
```

### Pre-configured Profiles

#### Free Local (`ollama.default`)
**Cost**: $0/month | **Setup**: Requires Ollama installed

```yaml
defaults:
  backend: ollama
  profile: default

backend:
  ollama:
    profiles:
      default:
        agent_models:
          router: llama3.1:8b
          query: gpt-oss:latest
          editor: qwen3-coder:latest
          research: gpt-oss:latest
```

#### Free Cloud (`openrouter.free`)
**Cost**: $0/month | **Setup**: Get free API key at openrouter.ai

```yaml
defaults:
  backend: openrouter
  profile: free

backend:
  openrouter:
    profiles:
      free:
        agent_models:
          router: google/gemini-2.0-flash-exp:free
          query: deepseek/deepseek-chat-v3.1:free
          editor: moonshotai/kimi-k2:free
          research: google/gemini-2.0-flash-exp:free
```

#### Budget ($0.85/month for 3M tokens)
**Cost**: ~$0.28/1M tokens | **Setup**: Get API key at console.groq.com

```yaml
backend:
  groq:
    profiles:
      budget:
        agent_models:
          router: llama-3.1-8b-instant      # $0.05/$0.08
          query: openai/gpt-oss-120b        # $0.15/$0.60
          editor: qwen/qwen3-32b            # $0.29/$0.59 (coding-focused)
          research: groq/compound           # $0.15/$0.60 base
```

**Why these models?**
- Router: 8B for speed (called most frequently)
- Query: 120B for reasoning
- Editor: 32B coding-specialized (40% cheaper than generic 70B)
- Research: Compound with built-in tools

#### Performance ($2.41/month for 3M tokens)
**Cost**: ~$0.80/1M tokens | **Premium quality**

```yaml
backend:
  groq:
    profiles:
      performance:
        agent_models:
          router: llama-3.1-8b-instant           # Speed still matters
          query: openai/gpt-oss-120b             # Same (already optimal)
          editor: moonshotai/kimi-k2-instruct    # $1.00/$3.00 (262K context)
          research: groq/compound                # Same (best with tools)
```

**Cost breakdown** (3M tokens/month):
- Router (40%): $0.06
- Query (30%): $0.34  
- Editor (25%): $1.95 (81% of total cost!)
- Research (5%): $0.06

### Creating Custom Profiles

```bash
# Create new profile
gptcode profiles create groq myprofile

# Configure agents
gptcode profiles set-agent groq myprofile router llama-3.1-8b-instant
gptcode profiles set-agent groq myprofile query openai/gpt-oss-120b
gptcode profiles set-agent groq myprofile editor llama-3.3-70b-versatile
gptcode profiles set-agent groq myprofile research groq/compound

# Use it
gptcode profile use groq.myprofile
```

### Cost Comparison

| Profile | Monthly Cost | Per 1M Tokens | Use Case |
|---------|--------------|---------------|----------|
| **ollama.default** | $0 | $0 | Local, privacy, no internet |
| **openrouter.free** | $0 | $0 | Cloud free tier, rate limits |
| **groq.budget** | $0.85 (3M) | $0.28 | Cost-optimized cloud |
| **groq.performance** | $2.41 (3M) | $0.80 | Quality-first cloud |
| **Claude Pro** | $200 | $41.67 | Traditional subscription |

**Groq is 99% cheaper** than Claude Pro for equivalent usage!

**[Complete Profile Guide](https://jadercorrea.github.io/gptcode/blog/2025-11-15-groq-optimal-configs)**

## Architecture

GPTCode uses specialized agents for different tasks:

- **Router**: Fast intent classification (8B model)
- **Query**: Smart code analysis (70B model)
- **Editor**: Code generation with large context (256K context)
- **Research**: Web search and documentation (free Compound model)

Each agent can use a different model, optimizing for cost vs capability.

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Releases & Versioning

GPTCode follows [Semantic Versioning](https://semver.org/) (MAJOR.MINOR.PATCH).

### Automatic Releases

When code is merged to `main` and CI passes, a new **patch version** is automatically created:
- `v0.0.1` → `v0.0.2` → `v0.0.3`

Weekly (Mondays 9AM UTC), the model catalog is updated. If models change, a new patch release is created automatically.

### Manual Releases (Major/Minor)

For breaking changes or new features, create a tag manually:

```bash
# Minor version (new features, backwards compatible)
git tag -a v0.2.0 -m "Release v0.2.0: Add profile management"
git push origin v0.2.0

# Major version (breaking changes)
git tag -a v1.0.0 -m "Release v1.0.0: Stable API"
git push origin v1.0.0

# Specific patch (bug fixes)
git tag -a v1.0.1 -m "Release v1.0.1: Fix authentication bug"
git push origin v1.0.1
```

The CD pipeline will automatically build binaries and create a GitHub release.

## Documentation & Blog

The website and blog are built with Jekyll and hosted on GitHub Pages.

## End-to-End Testing

GPTCode includes a comprehensive E2E testing framework using **Go tests** with real chu commands and **local Ollama models** (zero API costs, privacy-preserving).

### Requirements

**Software:**
- Ollama installed and running (`brew install ollama` on macOS)
- At least one profile configured with Ollama models
- Recommended 'local' profile models:
  - `llama3.1:8b` (4.7GB) - router agent
  - `qwen3-coder:latest` (18GB) - editor agent  
  - `gpt-oss:latest` (13GB) - query/research agents

**Installation:**
```bash
# 1. Install Ollama
brew install ollama  # macOS
# or visit https://ollama.ai for other platforms

# 2. Pull required models
ollama pull llama3.1:8b
ollama pull qwen3-coder:latest
ollama pull gpt-oss:latest

# 3. Create E2E profile (if not exists)
gptcode setup  # or manually configure ~/.gptcode/setup.yaml
```

### Running Tests

**Interactive profile selection (first time):**
```bash
gptcode test e2e --interactive
# Lists available profiles, saves selection as default
```

**With default profile:**
```bash
gptcode test e2e              # Run all tests
gptcode test e2e run          # Run only 'run' category tests
```

**With specific profile:**
```bash
gptcode test e2e --profile local
gptcode test e2e run --profile local
```

**With notifications (macOS):**
```bash
gptcode test e2e --notify
# Shows desktop notification when tests complete
```

**Custom timeout:**
```bash
gptcode test e2e --timeout 600  # 10 minutes per test
```

**Features:**
- ⏱️ Real-time progress bar with countdown
- 📊 Live test status (passed/failed counts)
- 🔔 macOS desktop notifications on completion
- ⚡ Automatically uses configured profile models
- 📁 Tests run in isolated temp directories

### Test Configuration

**Config file (`~/.gptcode/setup.yaml`):**
```yaml
e2e:
  default_profile: local      # Profile to use for tests
  timeout: 600                # Timeout per test (seconds)
  notify: true                # Desktop notifications
  parallel: 1                 # Parallel test execution (future)

backend:
  ollama:
    profiles:
      local:
        agent_models:
          router: llama3.1:8b
          query: gpt-oss:latest
          editor: qwen3-coder:latest
          research: gpt-oss:latest
```

**Environment variables (for tests):**
- `E2E_BACKEND` - Backend being used (set by chu test e2e)
- `E2E_PROFILE` - Profile being used (set by chu test e2e)
- `E2E_TIMEOUT` - Timeout in seconds (set by chu test e2e)
- `GPTCODE_NO_NOTIFY` - Set to disable notifications

### Current Test Coverage

**Run Command Tests (tests/e2e/run/):**
- ✅ `TestE2EConfiguration` - Validates E2E environment setup
- ✅ `TestChuCommand` - Verifies chu binary availability
- ✅ `TestChuDoCreateFile` - Tests file creation with specific content
- ✅ `TestChuDoModifyFile` - Tests file modification
- ✅ `TestChuDoNoUnintendedFiles` - Tests file validation (no extras)
- ⏭️ `TestChuDoTimeout` - Validates execution timeout (skipped, too slow)

**Known Limitations:**
- Ollama models are slow (2-5 minutes per test with local models)
- Tests use 10-minute timeout by default (600s)
- Progress bar updates every second during test execution
- Recommended for overnight runs or CI with longer timeouts

### Recommended Profile Configuration

The 'local' profile uses different models per agent, optimizing for their specific tasks:

| Agent | Model | Size | Purpose |
|-------|-------|------|----------|
| Router | `llama3.1:8b` | 4.7GB | Fast intent classification |
| Query | `gpt-oss:latest` | 13GB | Code analysis, reasoning |
| Editor | `qwen3-coder:latest` | 18GB | Code generation |
| Research | `gpt-oss:latest` | 13GB | Codebase analysis |

**Why this matters:**
- Router needs speed (8B) for quick routing
- Editor needs coding capability (Qwen3-coder specializes in code)
- Query/Research need reasoning (GPT-OSS balances capability and speed)

### Adding New Test Scenarios

1. Create test file in `tests/e2e/<category>/`:
```go
package category_test

import (
	"os"
	"os/exec"
	"testing"
)

func skipIfNoE2E(t *testing.T) {
	if os.Getenv("E2E_BACKEND") == "" {
		t.Skip("Skipping E2E test: run via 'chu test e2e'")
	}
}

func TestYourFeature(t *testing.T) {
	skipIfNoE2E(t)

	tmpDir := t.TempDir()
	
	cmd := exec.Command("gptcode", "your", "command")
	cmd.Dir = tmpDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Test failed: %v\nOutput: %s", err, output)
	}
	
	// Your assertions here
}
```

2. Run tests:
```bash
gptcode test e2e <category>
```

3. Categories:
- `run` - Single-shot commands
- `chat` - Interactive chat mode  
- `tdd` - Test-driven development
- `integration` - Multi-step workflows

### Test Utilities

**Built-in Go testing:**
- `t.TempDir()` - Creates isolated test directory
- `exec.Command()` - Executes gptcode commands
- `os.ReadFile()` - Validates file contents
- `os.ReadDir()` - Validates directory structure

**Environment variables:**
- Tests automatically skip when `E2E_BACKEND` not set
- Use `skipIfNoE2E(t)` helper in your tests
- gptcode test e2e sets E2E_BACKEND, E2E_PROFILE, E2E_TIMEOUT

**Timeout handling:**
- Use Go channels for async execution with timeout
- Example: 5-minute timeout for slow Ollama operations

### Test Output Example

```bash
$ gptcode test e2e run

🧪 GPTCode E2E Tests
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Backend:  ollama
Profile:  local
Category: run
Timeout:  600s per test
Notify:   enabled

Agent Models:
  Router:   llama3.1:8b
  Query:    gpt-oss:latest
  Editor:   qwen3-coder:latest
  Research: gpt-oss:latest

Running Run tests from tests/e2e/run...

=== RUN   TestChuDoCreateFile
    chu_do_test.go:28: Running gptcode do in /tmp/TestChuDoCreateFile123
    chu_do_test.go:29: This may take 2-5 minutes with local Ollama...
    chu_do_test.go:60: ✓ gptcode do successfully created hello.txt
--- PASS: TestChuDoCreateFile (143.21s)

⏱️  2m23s | ✅ 1 passed | ❌ 0 failed | ⏳ 7m37s remaining | 🔄 TestChuDoModifyFile

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ All tests passed! (4/4)
⏱️  Total time: 8m45s

[macOS notification: "✅ All tests passed (4/4)"]
```
### Running Locally

```bash
cd docs
bundle install
bundle exec jekyll serve --port 4040
```

Site will be available at `http://localhost:4040/`

### Writing Blog Posts

1. Create a new post in `docs/_posts/`:
   - Filename format: `YYYY-MM-DD-title-slug.md`
   - Example: `2025-11-22-ml-powered-intelligence.md`

2. Add front matter:
   ```yaml
   ---
   layout: post
   title: "Your Post Title"
   date: 2025-11-22
   author: Jader Correa
   tags: [tag1, tag2]
   ---
   ```

3. Write content in Markdown

4. Test locally:
   ```bash
   cd docs
   bundle exec jekyll serve --port 4040
   ```

5. Submit via Pull Request

### Deployment

The site auto-deploys via GitHub Actions when changes are merged to `main`.

**Pull Request Process:**
1. Fork the repository
2. Create a branch: `git checkout -b blog/your-post-title`
3. Add your post in `docs/_posts/`
4. Commit: `git commit -m "Add blog post: Your Title"`
5. Push: `git push origin blog/your-post-title`
6. Open Pull Request on GitHub
7. After review and merge, post goes live automatically

**Note:** Posts with future dates won't appear until that date.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- **Website**: [jadercorrea.github.io/gptcode](https://jadercorrea.github.io/gptcode)
- **Blog**: [jadercorrea.github.io/gptcode/blog](https://jadercorrea.github.io/gptcode/blog)
- **Issues**: [GitHub Issues](https://github.com/jadercorrea/gptcode/issues)

## Community

- Ask questions in [Issues](https://github.com/jadercorrea/gptcode/issues)
- Report bugs in [Issues](https://github.com/jadercorrea/gptcode/issues)

---

Made with ❤️ for developers who can't afford $20/month subscriptions
