---
layout: post
title: "Why Chuchu? Building an Affordable AI Coding Assistant"
date: 2024-11-13
author: Jader Corrêa
---

# Why Chuchu? Building an Affordable AI Coding Assistant

## The Problem

AI coding assistants have revolutionized how we write code. Tools like GitHub Copilot, Cursor, and others offer incredible productivity boosts. But there's a catch: they're expensive.

- **GitHub Copilot**: $10-20/month
- **Cursor**: $20/month
- **Other AI IDEs**: $15-30/month

For many developers—especially students, open-source contributors, or those in developing countries—these costs add up quickly. A $20/month subscription might not seem like much, but it's a significant barrier when:

- You're learning to code
- You contribute to OSS in your free time
- Your local currency makes it 3-5x more expensive
- You're working on personal projects with no income

## The Solution: Chuchu

Chuchu is a **TDD-first AI coding assistant** designed to be:

### 1. **Radically Affordable**

Use the LLM providers you want:

- **Groq**: ~$0.05-0.79 per 1M tokens (blazing fast, ultra-cheap)
- **Ollama**: $0.00 (100% local, completely free)
- **OpenRouter**: Access to 200+ models, pay only for what you use
- **OpenAI/Anthropic**: Direct API access (often 10x cheaper than subscriptions)

**Real cost comparison**:
- Copilot: $20/month = $240/year
- Chuchu with Groq: **$2-5/month** typical usage = $24-60/year
- Chuchu with Ollama: **$0/year** (hardware you already own)

### 2. **TDD-First by Design**

Unlike chat-based assistants, Chuchu follows Test-Driven Development:

```
You: "Add user authentication"
Chuchu: 
  1. Writes failing tests
  2. Implements the feature
  3. Ensures tests pass
  4. Refactors code
```

This approach:
- Catches bugs early
- Creates self-documenting code
- Builds confidence in AI-generated code
- Teaches best practices

### 3. **Neovim Native**

Built for Neovim from day one:
- Semantic file navigation
- Tree-sitter integration
- LSP-powered context awareness
- Stays out of your way until needed

No Electron. No web UI bloat. Just your terminal and editor.

### 4. **Model Flexibility**

Switch between models based on task:
- **Router agent** (fast, cheap): llama-3.1-8b-instant
- **Query agent** (smart): llama-3.3-70b-versatile
- **Editor agent** (large context): kimi-k2-0905-1t-256k
- **Research agent** (free): groq/compound

Mix and match. Change anytime. No vendor lock-in.

## Who Is This For?

Chuchu is perfect if you:

- ✅ Can't afford $20+/month for AI coding tools
- ✅ Want control over your AI spending
- ✅ Prefer terminal/Neovim workflows
- ✅ Care about test coverage
- ✅ Want to learn TDD practices
- ✅ Need offline/local AI options
- ✅ Value transparency and open source

## The Vision

We believe AI coding assistance should be:

1. **Accessible**: No one should be priced out of productivity tools
2. **Transparent**: Know what you're paying, what models you're using
3. **Educational**: Learn TDD while getting work done
4. **Privacy-Conscious**: Local models = your code stays yours

Chuchu isn't trying to replace paid tools for everyone. If Cursor works for you and the cost is fine, that's great! But for the millions of developers who can't afford subscriptions, or who want more control, Chuchu offers a real alternative.

## Get Started

```bash
# Install
go install github.com/jadercorrea/chuchu/cmd/chu@latest

# Configure (use free Groq or local Ollama)
chu setup

# Start coding
nvim
```

**Total cost to try it**: $0

Join us in making AI coding assistance accessible to everyone.

---

*Have questions? Join our [GitHub Discussions](https://github.com/jadercorrea/chuchu/discussions)*
