# 🐺 Chuchu

[![CI](https://github.com/jadercorrea/chuchu/actions/workflows/ci.yml/badge.svg)](https://github.com/jadercorrea/chuchu/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jadercorrea/chuchu)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub Issues](https://img.shields.io/github/issues/jadercorrea/chuchu)](https://github.com/jadercorrea/chuchu/issues)

Chuchu (pronounced "shoo-shoo", Brazilian slang for something small and cute) is a command-line AI coding assistant that helps you write better code through Test-Driven Development—without breaking the bank.

## Why Chuchu?

**Radically affordable**: Use Groq for $2-5/month or Ollama for **$0/month**. Compare that to $20-30/month subscriptions.

[Read the full story →](https://jadercorrea.github.io/chuchu/blog/2025-11-13-why-chuchu)

## Features

- **TDD-First**: Writes tests before implementation
- **Multi-Agent Architecture**: Router, Query, Editor, and Research agents
- **Cost Control**: Mix and match cheap/free models per agent
- **Profile Management**: Switch between multiple model configurations
- **Model Flexibility**: Groq, Ollama, OpenRouter, OpenAI, Anthropic
- **Neovim Native**: Deep integration with LSP, Tree-sitter, file navigation
- **Web Search**: Research agent can search and summarize web content
- **Auto-Install Models**: Discover and install 193+ Ollama models from Neovim
- **Feedback & Learning**: Track model performance and improve recommendations

## The Chuchu Way: Research → Plan → Implement

Unlike traditional AI coding assistants that generate code immediately, Chuchu uses a structured workflow:

```bash
# 1. Research: Understand your codebase
chu research "How does authentication work?"

# 2. Plan: Create detailed implementation steps
chu plan "Add password reset feature"

# 3. Implement: Execute with verification
chu implement plan.md              # Interactive (step-by-step)
chu implement plan.md --auto       # Autonomous (fully automated)
```

**Why this matters:**
- ✅ Context-aware changes that fit your codebase
- ✅ Incremental verification at each step
- ✅ Choose your control level (interactive or autonomous)
- ✅ Lower costs through better planning

**[Read the Complete Workflow Guide](docs/workflow-guide.md)**

## Quick Start

### Installation

```bash
# Install via go
go install github.com/jadercorrea/chuchu/cmd/chu@latest

# Or build from source
git clone https://github.com/jadercorrea/chuchu
cd chuchu
go install ./cmd/chu
```

### Setup

```bash
# Interactive setup wizard
chu setup

# For ultra-cheap setup, use Groq (get free key at console.groq.com)
# For free local setup, use Ollama (no API key needed)
```

### Neovim Integration

Add to your Neovim config:

```lua
-- lazy.nvim
{
  dir = "~/workspace/chuchu/neovim",  -- adjust path to your clone
  config = function()
    require("chuchu").setup()
  end,
  keys = {
    { "<C-d>", "<cmd>ChuchuChat<cr>", desc = "Toggle Chuchu Chat" },
    { "<C-m>", "<cmd>ChuchuModels<cr>", desc = "Switch Model/Profile" },
    { "<leader>ms", "<cmd>ChuchuModelSearch<cr>", desc = "Search & Install Models" },
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
- `<leader>ca` - Autonomous Execution (:ChuchuAuto)
  - Execute implementation plans with verification
  - Shows progress in real-time notifications

### ML-Powered Intelligence (Built-in)

Chuchu embeds two lightweight ML models for instant decision-making with zero external dependencies:

#### 1. Complexity Detection
Automatically triggers Guided Mode for complex/multistep tasks in `chu chat`.

**Configuration:**
```bash
# View/set complexity threshold (default: 0.55)
chu config get defaults.ml_complex_threshold
chu config set defaults.ml_complex_threshold 0.6
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
chu config get defaults.ml_intent_threshold
chu config set defaults.ml_intent_threshold 0.8
```

**ML CLI Commands:**
```bash
# List available models
chu ml list

# Train models (uses Python)
chu ml train complexity
chu ml train intent

# Test models
chu ml test intent "explain this code"
chu ml eval intent -f ml/intent/data/eval.csv

# Pure-Go inference (no Python runtime)
chu ml predict "your task"                    # complexity (default)
chu ml predict complexity "implement oauth"   # explicit
chu ml predict intent "explain this code"     # intent classification
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
CHUCHU_DEBUG=1 chu chat "your query"
# [GRAPH] Built graph: 142 nodes, 287 edges
# [GRAPH] Selected 5 files:
# [GRAPH]   1. internal/agents/router.go (score: 0.842)
# [GRAPH]   2. internal/llm/provider.go (score: 0.731)
```

**Example:**
```bash
chu chat "fix bug in authentication"
# Without graph: Sends entire codebase (100k tokens)
# With graph: Sends auth.go + user.go + middleware.go + session.go + config.go (18k tokens)
```

## Usage

**[Complete Workflow Guide](docs/workflow-guide.md)** - Learn the full research → plan → implement workflow

### Chat Mode

```bash
chu chat "explain this function"
chu chat "add error handling to the database connection"
```

### Research Mode

```bash
chu research "how does goroutine scheduling work"
chu research "best practices for error handling in Go"
```

### Planning & Implementation

```bash
chu plan "add user authentication with JWT"
chu implement plan.md
```

### Autonomous Execution (Maestro)

**Fully autonomous execution with verification and error recovery:**

```bash
# Execute a plan with automatic verification
chu implement docs/plans/my-implementation.md --auto

# With custom retry limit
chu implement docs/plans/my-implementation.md --auto --max-retries 5

# Resume from last checkpoint
chu implement docs/plans/my-implementation.md --auto --resume

# Enable lint verification (optional)
chu implement docs/plans/my-implementation.md --auto --lint
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
:ChuchuAuto        " prompts for plan file and runs: chu implement <file> --auto
" Or keymap: <leader>ca
```

### Backend Management

```bash
# List configured backends
chu backend list

# Create new backend
chu backend create mygroq openai https://api.groq.com/openai/v1
chu key mygroq  # Set API key
chu config set backend.mygroq.default_model llama-3.3-70b-versatile

# Switch default backend
chu config set defaults.backend mygroq

# Delete backend
chu backend delete mygroq
```

### Profile Management

```bash
# List profiles for a backend
chu profiles list groq

# Show profile configuration
chu profiles show groq default

# Create new profile
chu profiles create groq speed

# Configure agents
chu profiles set-agent groq speed router llama-3.1-8b-instant
chu profiles set-agent groq speed query llama-3.1-8b-instant
```

### Model Discovery & Installation

```bash
# Search for ollama models
chu models search ollama llama3

# Search with multiple filters (ANDed)
chu models search ollama coding fast

# Install ollama model
chu models install llama3.1:8b
```

### Feedback & Learning

```bash
# Record positive feedback
chu feedback good --backend groq --model llama-3.3-70b-versatile --agent query

# Record negative feedback
chu feedback bad --backend groq --model llama-3.1-8b-instant --agent router

# View statistics
chu feedback stats
```

Chuchu learns from your feedback to recommend better models over time.

## Configuration Examples

### Budget Setup ($2-5/month)

```yaml
defaults:
  backend: groq
  
backend:
  groq:
    agent_models:
      router: llama-3.1-8b-instant      # $0.05/$0.08 per 1M tokens
      query: llama-3.3-70b-versatile    # $0.59/$0.79 per 1M tokens
      editor: llama-3.3-70b-versatile
      research: groq/compound           # Free!
```

### Free Local Setup ($0/month)

```yaml
defaults:
  backend: ollama
  
backend:
  ollama:
    agent_models:
      router: llama3.1:8b
      query: qwen3-coder:latest
      editor: qwen3-coder:latest
      research: qwen3-coder:latest
```

### Multiple Profiles per Backend

```yaml
defaults:
  backend: groq
  profile: speed  # or: default, quality, free
  
backend:
  groq:
    profiles:
      speed:
        agent_models:
          router: llama-3.1-8b-instant
          query: llama-3.1-8b-instant
          editor: llama-3.1-8b-instant
          research: llama-3.1-8b-instant
      quality:
        agent_models:
          router: llama-3.1-8b-instant
          query: llama-3.3-70b-versatile
          editor: llama-3.3-70b-versatile
          research: groq/compound
```

[See more configurations →](https://jadercorrea.github.io/chuchu/blog/2025-11-15-groq-optimal-configs)

## Architecture

Chuchu uses specialized agents for different tasks:

- **Router**: Fast intent classification (8B model)
- **Query**: Smart code analysis (70B model)
- **Editor**: Code generation with large context (256K context)
- **Research**: Web search and documentation (free Compound model)

Each agent can use a different model, optimizing for cost vs capability.

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Releases & Versioning

Chuchu follows [Semantic Versioning](https://semver.org/) (MAJOR.MINOR.PATCH).

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

Chuchu includes a comprehensive E2E testing framework using **Go tests** with real chu commands and **local Ollama models** (zero API costs, privacy-preserving).

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
chu setup  # or manually configure ~/.chuchu/setup.yaml
```

### Running Tests

**Interactive profile selection (first time):**
```bash
chu test e2e --interactive
# Lists available profiles, saves selection as default
```

**With default profile:**
```bash
chu test e2e              # Run all tests
chu test e2e run          # Run only 'run' category tests
```

**With specific profile:**
```bash
chu test e2e --profile local
chu test e2e run --profile local
```

**With notifications (macOS):**
```bash
chu test e2e --notify
# Shows desktop notification when tests complete
```

**Custom timeout:**
```bash
chu test e2e --timeout 600  # 10 minutes per test
```

**Features:**
- ⏱️ Real-time progress bar with countdown
- 📊 Live test status (passed/failed counts)
- 🔔 macOS desktop notifications on completion
- ⚡ Automatically uses configured profile models
- 📁 Tests run in isolated temp directories

### Test Configuration

**Config file (`~/.chuchu/setup.yaml`):**
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
- `CHUCHU_NO_NOTIFY` - Set to disable notifications

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
	
	cmd := exec.Command("chu", "your", "command")
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
chu test e2e <category>
```

3. Categories:
- `run` - Single-shot commands
- `chat` - Interactive chat mode  
- `tdd` - Test-driven development
- `integration` - Multi-step workflows

### Test Utilities

**Built-in Go testing:**
- `t.TempDir()` - Creates isolated test directory
- `exec.Command()` - Executes chu commands
- `os.ReadFile()` - Validates file contents
- `os.ReadDir()` - Validates directory structure

**Environment variables:**
- Tests automatically skip when `E2E_BACKEND` not set
- Use `skipIfNoE2E(t)` helper in your tests
- chu test e2e sets E2E_BACKEND, E2E_PROFILE, E2E_TIMEOUT

**Timeout handling:**
- Use Go channels for async execution with timeout
- Example: 5-minute timeout for slow Ollama operations

### Test Output Example

```bash
$ chu test e2e run

🧪 Chuchu E2E Tests
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
    chu_do_test.go:28: Running chu do in /tmp/TestChuDoCreateFile123
    chu_do_test.go:29: This may take 2-5 minutes with local Ollama...
    chu_do_test.go:60: ✓ chu do successfully created hello.txt
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

- **Website**: [jadercorrea.github.io/chuchu](https://jadercorrea.github.io/chuchu)
- **Blog**: [jadercorrea.github.io/chuchu/blog](https://jadercorrea.github.io/chuchu/blog)
- **Issues**: [GitHub Issues](https://github.com/jadercorrea/chuchu/issues)

## Community

- Ask questions in [Issues](https://github.com/jadercorrea/chuchu/issues)
- Report bugs in [Issues](https://github.com/jadercorrea/chuchu/issues)

---

Made with ❤️ for developers who can't afford $20/month subscriptions
