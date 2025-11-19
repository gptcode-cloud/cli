# 🐺 Chuchu

**An affordable, TDD-first AI coding assistant for Neovim**

[![Go Version](https://img.shields.io/github/go-mod/go-version/jadercorrea/chuchu)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub Discussions](https://img.shields.io/github/discussions/jadercorrea/chuchu)](https://github.com/jadercorrea/chuchu/discussions)

Chuchu (pronounced "shoo-shoo", Brazilian slang for something small and cute) is a command-line AI coding assistant that helps you write better code through Test-Driven Development—without breaking the bank.

## 💰 Why Chuchu?

**Radically affordable**: Use Groq for $2-5/month or Ollama for **$0/month**. Compare that to $20-30/month subscriptions.

[Read the full story →](https://jadercorrea.github.io/chuchu/blog/2025-01-19-why-chuchu)

## ✨ Features

- **🧪 TDD-First**: Writes tests before implementation
- **🎯 Multi-Agent Architecture**: Router, Query, Editor, and Research agents
- **💸 Cost Control**: Mix and match cheap/free models per agent
- **🔌 Model Flexibility**: Groq, Ollama, OpenRouter, OpenAI, Anthropic
- **📦 Neovim Native**: Deep integration with LSP, Tree-sitter, file navigation
- **🌐 Web Search**: Research agent can search and summarize web content
- **🚀 Auto-Install Models**: Discover and install 193+ Ollama models from Neovim

## 🚀 Quick Start

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
    { "<leader>cc", "<cmd>lua require('chuchu').toggle()<cr>", desc = "Toggle Chuchu" },
    { "<leader>cs", "<cmd>lua require('chuchu').send_message()<cr>", desc = "Send message" },
  }
}
```

## 📖 Usage

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

## 💡 Configuration Examples

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

[See more configurations →](https://jadercorrea.github.io/chuchu/blog/2025-01-18-groq-optimal-configs)

## 🏗️ Architecture

Chuchu uses specialized agents for different tasks:

- **Router**: Fast intent classification (8B model)
- **Query**: Smart code analysis (70B model)
- **Editor**: Code generation with large context (256K context)
- **Research**: Web search and documentation (free Compound model)

Each agent can use a different model, optimizing for cost vs capability.

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🔗 Links

- **Website**: [jadercorrea.github.io/chuchu](https://jadercorrea.github.io/chuchu)
- **Blog**: [jadercorrea.github.io/chuchu/blog](https://jadercorrea.github.io/chuchu/blog)
- **Discussions**: [GitHub Discussions](https://github.com/jadercorrea/chuchu/discussions)
- **Issues**: [GitHub Issues](https://github.com/jadercorrea/chuchu/issues)

## 💬 Community

- Ask questions in [Discussions](https://github.com/jadercorrea/chuchu/discussions)
- Share your configs in [Show and Tell](https://github.com/jadercorrea/chuchu/discussions/categories/show-and-tell)
- Report bugs in [Issues](https://github.com/jadercorrea/chuchu/issues)

---

Made with ❤️ for developers who can't afford $20/month subscriptions
