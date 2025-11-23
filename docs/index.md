---
layout: default
title: Chuchu
description: An affordable, TDD-first AI coding assistant for Neovim
---

<div class="hero">
  <h1>AI Coding Assistant<br/>You Can Actually Afford</h1>
  <p>Test-driven development with multi-agent AI. Deep Neovim integration. Mix cheap and free models. $2-5/month with Groq or $0 with Ollama.</p>
  <div class="hero-cta">
    <a href="#quick-start" class="btn btn-primary">Get Started</a>
    <a href="https://github.com/jadercorrea/chuchu" class="btn btn-secondary">View on GitHub</a>
  </div>
</div>

<div class="features">
  <a href="/features#tdd-first-workflow" class="feature-card">
    <h3>TDD-First Workflow</h3>
    <p>Writes tests before implementation. Focuses on small, testable functions. Keeps you honest with clear requirements.</p>
  </a>
  
  <a href="/features#multi-agent-architecture" class="feature-card">
    <h3>Multi-Agent Architecture</h3>
    <p>Router, Query, Editor, and Research agents. Each specialized for its task. Mix models based on cost and quality needs.</p>
  </a>
  
  <a href="/features#cost-optimization" class="feature-card">
    <h3>Radically Affordable</h3>
    <p>Use Groq for $2-5/month or Ollama locally for free. Configure per-agent models to optimize cost vs performance.</p>
  </a>
  
  <a href="/features#ml-powered-intelligence" class="feature-card">
    <h3>ML-Powered Routing</h3>
    <p>Intent classification in 1ms (500x faster than LLM). Complexity detection auto-triggers planning mode. Zero external dependencies.</p>
  </a>
  
  <a href="/features#smart-context-selection" class="feature-card">
    <h3>Smart Context Selection</h3>
    <p>Dependency graph + PageRank identifies relevant files. 5x token reduction (100k → 20k). Better responses, lower costs.</p>
  </a>
  
  <a href="/features#neovim-integration" class="feature-card">
    <h3>Deep Neovim Integration</h3>
    <p>Native chat interface. Profile management UI. Model search and auto-install. LSP and Tree-sitter aware.</p>
  </a>
</div>

<div class="section">
  <h2 class="section-title">How It Works</h2>
  <p class="section-subtitle">Multi-agent system with intelligent routing and context optimization</p>
  
  <div class="mermaid">
graph TB
    User[User Query] --> ML[ML Intent Classifier<br/>1ms]
    ML -->|query| QA[Query Agent<br/>Code Reading]
    ML -->|edit| EA[Editor Agent<br/>Code Writing]
    ML -->|research| RA[Research Agent<br/>Web Search]
    ML -->|uncertain| LLM[LLM Fallback<br/>500ms]
    LLM --> QA
    LLM --> EA
    LLM --> RA
    
    Context[Dependency Graph<br/>Context Optimizer] --> QA
    Context --> EA
    
    QA --> Response[Response]
    EA --> Response
    RA --> Response
    
    style ML fill:#16a34a
    style Context fill:#2563eb
  </div>
</div>

<div class="section">
  <h2 class="section-title">Structured Workflow: Research → Plan → Implement</h2>
  <p class="section-subtitle">Go from feature idea to tested code with AI assistance at each step</p>
  
  <div class="workflow-steps">
    <div class="workflow-step">
      <h3>🔍 Research</h3>
      <p>Understand your codebase before making changes</p>
      <pre><code>chu research "How does authentication work?"</code></pre>
      <p>Chuchu searches semantically, reads relevant files, and documents findings in <code>~/.chuchu/research/</code></p>
    </div>
    
    <div class="workflow-step">
      <h3>📋 Plan</h3>
      <p>Create detailed implementation plan with phases</p>
      <pre><code>chu plan "Add password reset feature"</code></pre>
      <p>Generates step-by-step plan with clear goals, file changes, and test requirements</p>
    </div>
    
    <div class="workflow-step">
      <h3>⚡ Implement</h3>
      <p>Execute plan interactively or autonomously</p>
      <pre><code>chu implement plan.md
chu implement plan.md --auto</code></pre>
      <p>Interactive mode for control, autonomous mode for speed with automatic verification and retry</p>
    </div>
  </div>
  
  <p style="text-align: center; margin-top: 2rem;">
    <a href="/workflow-guide" class="btn btn-primary">Complete Workflow Guide</a>
    <a href="/blog/2025-11-24-complete-workflow-guide" class="btn btn-secondary">Read Tutorial</a>
  </p>
</div>

<div class="section" id="quick-start">
  <h2 class="section-title">Quick Start</h2>
  
  <div class="quick-start">
    <h3>1. Install CLI</h3>
    <pre><code>go install github.com/jadercorrea/chuchu/cmd/chu@latest
chu setup</code></pre>
    
    <h3>2. Add Neovim Plugin</h3>
    <pre><code>-- lazy.nvim
{
  dir = "~/workspace/chuchu/neovim",
  config = function()
    require("chuchu").setup()
  end,
  keys = {
    { "&lt;C-d&gt;", "&lt;cmd&gt;ChuchuChat&lt;cr&gt;", desc = "Toggle Chat" },
    { "&lt;C-m&gt;", "&lt;cmd&gt;ChuchuModels&lt;cr&gt;", desc = "Profiles" },
  }
}</code></pre>
    
    <h3>3. Start Coding</h3>
    <pre><code>chu chat "add user authentication with JWT"
chu research "best practices for error handling"
chu plan "implement rate limiting"</code></pre>
  </div>
</div>

<div class="section">
  <h2 class="section-title">Features</h2>
  
  <h3>Profile Management</h3>
  <p>Switch between model configurations instantly. Budget profile with Groq Llama 3.1 8B. Quality profile with GPT-4. Local profile with Ollama.</p>
  
  <h3>Cost Optimization</h3>
  <ul>
    <li><strong>Router agent</strong>: Fast, cheap model (Llama 3.1 8B @ $0.05/M tokens)</li>
    <li><strong>Query agent</strong>: Balanced model (GPT-OSS 120B @ $0.15/M)</li>
    <li><strong>Editor agent</strong>: Quality model (DeepSeek R1 @ $0.14/M)</li>
    <li><strong>Research agent</strong>: Context-heavy model (Grok 4.1 @ free tier)</li>
  </ul>
  
  <h3>ML Intelligence</h3>
  <ul>
    <li><strong>Intent classifier</strong>: 1ms routing, 89% accuracy, smart LLM fallback</li>
    <li><strong>Complexity detector</strong>: Auto-triggers planning for complex tasks</li>
    <li><strong>Pure Go inference</strong>: No Python runtime required</li>
  </ul>
  
  <h3>Context Optimization</h3>
  <ul>
    <li>Dependency graph analysis (Go, Python, JS/TS, Ruby, Rust)</li>
    <li>PageRank file importance ranking</li>
    <li>1-hop neighbor expansion</li>
    <li>5x token reduction with better accuracy</li>
  </ul>
  
  <h3>Neovim Features</h3>
  <ul>
    <li>Floating chat window with syntax highlighting</li>
    <li>Model search UI (193+ Ollama models)</li>
    <li>Profile management interface</li>
    <li>LSP-aware code context</li>
    <li>Tree-sitter integration</li>
  </ul>
</div>

<div class="section">
  <h2 class="section-title">Why Chuchu?</h2>
  
  <p>Most AI coding assistants lock you into expensive subscriptions ($20-30/month) with black-box model selection. Chuchu gives you:</p>
  
  <ul>
    <li><strong>Full control</strong>: Choose any OpenAI-compatible provider</li>
    <li><strong>Cost transparency</strong>: See exactly what you're paying per token</li>
    <li><strong>Flexibility</strong>: Mix models based on task complexity</li>
    <li><strong>Local option</strong>: Run completely offline with Ollama</li>
    <li><strong>TDD focus</strong>: Tests first, implementation second</li>
  </ul>
  
  <p>Read the full story: <a href="/blog/2025-11-13-why-chuchu">Why Chuchu?</a></p>
</div>

<div class="section">
  <h2 class="section-title">Documentation</h2>
  
  <ul>
    <li><a href="/commands">Commands Reference</a> – Complete CLI command guide</li>
    <li><a href="/research">Research Mode</a> – Web search and documentation lookup</li>
    <li><a href="/plan">Plan Mode</a> – Planning and implementation workflow</li>
    <li><a href="/ml-features">ML Features</a> – Intent classification and complexity detection</li>
    <li><a href="/compare">Compare Models</a> – Interactive model comparison tool</li>
    <li><a href="/blog">Blog</a> – Configuration guides and best practices</li>
  </ul>
</div>

<script type="module">
  import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
  mermaid.initialize({ startOnLoad: true, theme: 'base' });
</script>
