---
layout: default
title: Features
description: Complete feature set of GPTCode AI coding assistant
---

# Features

## Agent-Based Architecture

### Autonomous Execution with `gptcode do`

The flagship command that orchestrates 4 specialized agents working in sequence:

1. **Analyzer**: Understands codebase, reads relevant files using dependency graph
2. **Planner**: Creates minimal implementation plan, lists files to modify
3. **Editor**: Executes changes ONLY on planned files (file validation)
4. **Validator**: Verifies success criteria, triggers auto-retry if needed

**Usage:**
```bash
gptcode do "add JWT authentication"
gptcode do "fix bug in payment processing" --supervised
gptcode do "refactor error handling" --interactive
```

**Flags:**
- `--supervised` - Manual approval before implementation
- `--interactive` - Prompt when model selection is ambiguous
- `--dry-run` - Show plan only
- `-v` - Verbose (show model selection)
- `--max-attempts N` - Max retry attempts (default 3)

**Benefits:**
- Automatic model selection per agent (queries performance history)
- Auto-retry with better models when validation fails
- File validation prevents unintended changes
- Success criteria checked before completion

[See full agent flow diagram on homepage →](/)

---

## Validation & Safety

### File Validation
The Editor agent can **only** modify files explicitly mentioned in the Planner's output. This prevents:
- Creating unexpected configuration files
- Modifying unrelated code
- Adding surprise scripts

### Success Criteria Validation
The Validator agent automatically:
1. Checks if task completion criteria are met
2. Runs tests if applicable
3. Verifies file changes match plan
4. Triggers retry with feedback if validation fails

### Supervised vs Autonomous Modes
- **Autonomous** (default): Fast execution with automatic validation
- **Supervised** (`--supervised`): Manual approval before implementation

Choose based on task criticality.

---

## ML-Powered Intelligence

### Intent Classification
Routes requests in 1ms instead of 500ms LLM calls. Classifies user intent (query, edit, research, review) with 89% accuracy and smart LLM fallback when uncertain.

**Benefits:**
- 500x faster routing (1ms vs 500ms)
- 80% cost reduction for routing operations
- Zero API calls for confident predictions
- Smart fallback maintains quality

**Configuration:**
```bash
gptcode config get defaults.ml_intent_threshold  # default: 0.7
gptcode config set defaults.ml_intent_threshold 0.8
```

### Complexity Detection
Automatically triggers Guided Mode (research → plan → implement) for complex multi-step tasks.

**Configuration:**
```bash
gptcode config get defaults.ml_complex_threshold  # default: 0.55
gptcode config set defaults.ml_complex_threshold 0.6
```

**CLI Commands:**
```bash
gptcode ml list                    # List available models
gptcode ml test intent "query"     # Test intent classification
gptcode ml eval intent             # Evaluate accuracy
gptcode ml train intent            # Retrain model (requires Python)
```

[Read more about ML features →](/ml-features)

---

## Smart Context Selection

### Dependency Graph Analysis
Automatically builds a graph of your codebase's file dependencies and uses PageRank to identify important files.

**How it works:**
1. Analyzes imports/requires to build dependency graph
2. Ranks files by importance using PageRank
3. Matches query terms to relevant files
4. Expands to 1-hop neighbors (dependencies + dependents)
5. Provides top 5 most relevant files as context

**Benefits:**
- 5x token reduction (100k → 20k tokens)
- Better responses with focused context
- Automatic, transparent operation
- Cached for performance

**Supported languages:**
Go, Python, JavaScript/TypeScript, Ruby, Rust

**Debug mode:**
```bash
GPTCODE_DEBUG=1 gptcode chat "your query"
# [GRAPH] Built graph: 142 nodes, 287 edges
# [GRAPH] Selected 5 files:
# [GRAPH]   1. internal/agents/router.go (score: 0.842)
```

[Read more about graph features →](/graph-features)

---

## Multi-Agent Architecture

### Specialized Agents

**Router Agent** (fast, cheap)
- Intent classification and routing
- Recommended: Llama 3.1 8B Instant (840 TPS, $0.05/M)

**Query Agent** (comprehension)
- Code reading and analysis
- Recommended: GPT-OSS 120B ($0.15/M) or Qwen 2.5 Coder

**Editor Agent** (code generation)
- Code writing and modification
- Recommended: DeepSeek R1 Distill (83.3% AIME) or Qwen 2.5 Coder

**Research Agent** (web search)
- Web search and documentation lookup
- Recommended: Grok 4.1 Fast (2M context, free tier)

### Agent Configuration

```yaml
backend:
  groq:
    agent_models:
      router: llama-3.1-8b-instant
      query: gpt-oss-120b-128k
      editor: deepseek-r1-distill-qwen-32b
      research: groq/compound
```

[Compare models →](/compare)

---

## Profile Management

Switch between model configurations instantly:

- **Budget profile**: Groq with Llama 3.1 8B ($2-5/month)
- **Quality profile**: GPT-4 or Claude for critical work
- **Local profile**: Ollama for complete privacy ($0/month)
- **Hybrid profile**: Mix cloud and local models

**Neovim UI:**
- `<C-m>` - Profile management interface
- Create, load, edit, delete profiles
- Configure per-agent models
- View profile details and costs

**CLI:**
```bash
gptcode backend list           # List configured backends
gptcode backend use groq       # Switch to Groq backend
```

---

## TDD-First Workflow

### Test-Driven Development
- Writes tests before implementation
- Focuses on small, testable functions
- Enforces clear requirements
- Keeps functions focused and maintainable

### Commands
```bash
gptcode tdd                    # Interactive TDD mode
gptcode feature "description"  # Generate tests + implementation
```

### Workflow
1. Describe feature requirements
2. AI generates tests first
3. Tests guide implementation
4. Verify with test suite
5. Iterate until green

---

## Neovim Integration

### Chat Interface
- Floating window with syntax highlighting
- Context-aware suggestions
- LSP and Tree-sitter integration
- Persistent chat history

### Model Management
- Search 193+ Ollama models
- Auto-install models directly from Neovim
- View pricing and context windows
- Set default or session-specific models

### Key Bindings (configurable)
```lua
<C-d>      -- Toggle chat interface
<C-m>      -- Profile management
<leader>ms -- Model search and install
```

### Features
- Code context from LSP
- Tree-sitter aware
- Multiple file support
- Diff preview
- Interactive code review

---

## Cost Optimization

### Per-Agent Pricing
Configure different model tiers based on task importance:

| Agent | Model | Input | Output | Use Case |
|-------|-------|-------|--------|----------|
| Router | Llama 3.1 8B | $0.05 | $0.08 | Fast intent classification |
| Query | GPT-OSS 120B | $0.15 | $0.60 | Code comprehension |
| Editor | DeepSeek R1 | $0.14 | $0.42 | Code generation |
| Research | Grok 4.1 Free | $0.00 | $0.00 | Web search |

### Monthly Cost Examples
- **Budget**: $2-5/month (Groq with small models)
- **Balanced**: $10-20/month (mix of models)
- **Quality**: $30-50/month (premium models for editor)
- **Local**: $0/month (Ollama only)

[See optimal configurations →](/blog/2025-11-15-groq-optimal-configs)

---

## Local Deployment

### Ollama Support
Run completely offline with Ollama:

**Recommended models:**
- Qwen 2.5 Coder 32B (88.4% HumanEval, requires 32GB RAM)
- DeepSeek Coder 33B (81.1% HumanEval, requires 32GB RAM)
- Llama 3.1 8B (budget option, 8GB RAM)

**Configuration:**
```yaml
backend:
  ollama:
    base_url: http://localhost:11434
    default_model: qwen2.5-coder:32b
```

**Benefits:**
- Zero API costs
- Complete privacy
- No internet required
- No rate limits

[Setup guide →](/blog/2025-11-17-ollama-local-setup)

---

## OpenRouter Integration

Access 100+ models through single API:

- Free tier models (Grok 4.1 Fast, GPT-OSS)
- Premium models (Claude, GPT-4)
- Fallback routing
- Automatic retries

**Configuration:**
```yaml
backend:
  openrouter:
    base_url: https://openrouter.ai/api/v1
    default_model: anthropic/claude-4.5-sonnet
```

[OpenRouter setup →](/blog/2025-11-16-openrouter-multi-provider)

---

## Research & Planning

### Research Mode
Comprehensive codebase research with parallel sub-agents:

```bash
gptcode research "how does authentication work"
```

- Spawns specialized research agents
- Analyzes dependencies and patterns
- Generates detailed documentation
- Creates research artifacts

[Research workflow →](/prompts#research)

### Planning Mode
Interactive plan creation with iteration:

```bash
gptcode plan "add JWT authentication"
```

- Guided question/answer flow
- Validates against codebase
- Phases implementation
- Generates detailed specs

[Planning workflow →](/prompts#plan)

### Implementation
Execute plans with verification:

```bash
gptcode implement plan.md
```

- Step-by-step execution
- Automated testing
- Progress tracking
- Rollback support

---

## Model Comparison

Interactive tool to compare LLMs for coding:

- Side-by-side comparison (up to 4 models)
- Coding-specific benchmarks (HumanEval, SWE-Bench)
- Cost calculator for workflows
- Filter by provider, cost, speed, role

[Compare models →](/compare)
