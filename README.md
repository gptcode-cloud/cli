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
chu implement
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
