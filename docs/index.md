---
layout: default
title: GPTCode
description: AI coding assistant with specialized agents and validation
---

<div class="hero">
  <h1>The Open Source<br/>AI Coding Assistant</h1>
  <p>Transparent, Autonomous, and Local. <strong>Bring Your Own Keys</strong>. Run with Groq, OpenAI, or locally with Ollama. No subscriptions, full control.</p>
  <div class="hero-cta">
    <a href="#quick-start" class="btn btn-primary">Get Started</a>
    <a href="https://github.com/jadercorrea/gptcode" class="btn btn-secondary">View on GitHub</a>
  </div>
</div>

{% include explainer.html %}

<div class="features">
  <a href="/features#agent-based-architecture" class="feature-card">
    <h3>Autonomous Execution</h3>
    <p><code>chu do "task"</code> orchestrates 4 specialized agents: Analyzer → Planner → Editor → Validator. Auto-retry with model switching when validation fails.</p>
  </a>
  
  <a href="/features#file-validation" class="feature-card">
    <h3>Validation & Safety</h3>
    <p>File validation prevents creating unintended files. Success criteria auto-verified. No surprise scripts or configs. Supervised vs autonomous modes.</p>
  </a>
  
  <a href="/features#smart-context" class="feature-card">
    <h3>Intelligent Context</h3>
    <p>Dependency graph + PageRank identifies relevant files. 5x token reduction (100k → 20k). ML intent classification (1ms routing).</p>
  </a>
  
  <a href="/features#cost-optimization" class="feature-card">
    <h3>Radically Affordable</h3>
    <p>$0-5/month vs $20-30/month subscriptions. Use Groq for speed, Ollama for free. Auto-selects best models per agent.</p>
  </a>
  
  <a href="/commands#interactive-modes" class="feature-card">
    <h3>Interactive Modes</h3>
    <p><code>chu chat</code> for conversations. <code>chu run</code> for tasks with follow-up. Context-aware from CLI or Neovim plugin.</p>
  </a>
  
  <a href="/commands#workflow-commands-research--plan--implement" class="feature-card">
    <h3>Manual Workflow</h3>
    <p>Break down complex tasks: <code>chu research</code> → <code>chu plan</code> → <code>chu implement</code>. Full control when you need it.</p>
  </a>
</div>


<div class="section">
  <h2 class="section-title">Agent Orchestration: Analyzer → Planner → Editor → Validator</h2>
  <p class="section-subtitle">Fast routing, focused context, safe edits, and verified results — end to end</p>
  
  <div class="workflow-steps">
    <div class="workflow-step">
      <h3>Analyzer</h3>
      <p>Scans the codebase, builds dependency graph and selects only the relevant files</p>
      <pre><code>chu do "add authentication"</code></pre>
    </div>
    
    <div class="workflow-step">
      <h3>Planner</h3>
      <p>Creates a concrete plan with success criteria and allowed file list</p>
      <pre><code>chu do "add authentication"</code></pre>
    </div>
    
    <div class="workflow-step">
      <h3>Editor</h3>
      <p>Applies changes incrementally with file validation and auto-recovery</p>
      <pre><code>chu do "add authentication"</code></pre>
    </div>
    
    <div class="workflow-step">
      <h3>Validator</h3>
      <p>Runs tests and checks success criteria before finishing</p>
      <pre><code>chu do "add authentication"</code></pre>
    </div>
  </div>
  
  <p style="text-align: center; margin-top: 2rem;">
    <a href="/features#agent-based-architecture" class="btn btn-primary">Learn the Architecture</a>
    <a href="/features#file-validation" class="btn btn-secondary">See Validation & Safety</a>
  </p>
</div>

<div class="section" id="quick-start">
  <h2 class="section-title">Quick Start</h2>
  
  <div class="quick-start">
    <h3>1. Install CLI</h3>
    <pre><code>go install github.com/jadercorrea/gptcode/cmd/chu@latest
chu setup</code></pre>
    
    <h3>2. Add Neovim Plugin</h3>
    <pre><code>-- lazy.nvim
{
  dir = "~/workspace/gptcode/neovim",
  config = function()
    require("gptcode").setup()
  end,
  keys = {
    { "&lt;C-d&gt;", "&lt;cmd&gt;GPTCodeChat&lt;cr&gt;", desc = "Toggle Chat" },
    { "&lt;C-m&gt;", "&lt;cmd&gt;GPTCodeModels&lt;cr&gt;", desc = "Profiles" },
  }
}</code></pre>
    
    <h3>3. Start Coding</h3>
    <pre><code>chu chat "add user authentication with JWT"
chu research "best practices for error handling"
chu plan "implement rate limiting"</code></pre>
  </div>
</div>

<div class="section">
  <h2 class="section-title">Core Capabilities</h2>
  
  <h3>Three Ways to Work</h3>
  <ul>
    <li><strong>Autonomous Copilot</strong>: <code>chu do "task"</code> handles everything—analysis, planning, execution, validation</li>
    <li><strong>Interactive Chat</strong>: <code>chu chat</code> for conversations with context awareness and follow-ups</li>
    <li><strong>Structured Workflow</strong>: <code>chu research</code> → <code>chu plan</code> → <code>chu implement</code> for full control</li>
  </ul>
  
  <h3>Special Modes</h3>
  <ul>
    <li><strong>TDD Mode</strong>: <code>chu tdd</code> for test-driven development workflow</li>
    <li><strong>Code Review</strong>: <code>chu review</code> for automated bug detection and security analysis</li>
    <li><strong>Task Execution</strong>: <code>chu run</code> for tasks with follow-up conversations</li>
    <li><strong>Web Research</strong>: Built-in search and documentation lookup</li>
  </ul>
  
  <h3>Intelligence & Optimization</h3>
  <ul>
    <li><strong>Multi-Agent Architecture</strong>: Router, Query, Editor, Research agents working together</li>
    <li><strong>ML-Powered</strong>: Intent classification (1ms) and complexity detection with zero API calls</li>
    <li><strong>Dependency Graph</strong>: Smart context selection with 5x token reduction (Go, Python, JS/TS, Ruby, Rust)</li>
    <li><strong>Cost Optimized</strong>: Mix cheap/free models per agent ($0-5/month vs $20-30/month)</li>
  </ul>
  
  <h3>Developer Experience</h3>
  <ul>
    <li><strong>Profile Management</strong>: Switch between cost/speed/quality configurations instantly</li>
    <li><strong>Model Flexibility</strong>: Groq, Ollama, OpenRouter, OpenAI, Anthropic, DeepSeek</li>
    <li><strong>Neovim Integration</strong>: Floating chat, model search (300+ models), LSP/Tree-sitter aware</li>
    <li><strong>Validation & Safety</strong>: File validation, success criteria, supervised mode</li>
  </ul>
</div>

<div class="section">
  <h2 class="section-title">Why GPTCode?</h2>
  
  <p>GPTCode isn't trying to beat Cursor or Copilot. It's trying to be different—and yours.</p>
  
  <ul>
    <li><strong>Transparent</strong>: When it breaks, you can read and fix the code</li>
    <li><strong>Hackable</strong>: Don't like something? Change it—it's just Go</li>
    <li><strong>Model agnostic</strong>: Switch LLMs in 2 minutes (Groq, Ollama, OpenAI, etc.)</li>
    <li><strong>Honest</strong>: E2E tests at 55%—no "95% accuracy" marketing</li>
    <li><strong>Affordable</strong>: $2–5/month (Groq) or $0/month (Ollama)</li>
  </ul>
  
  <p>
    <a href="/blog/2025-12-06-why-gptcode-isnt-trying-to-beat-anyone">Read the full positioning →</a>
    · <a href="/blog/2025-11-13-why-gptcode">Original vision →</a>
    · <a href="/blog/2025-12-01-agent-routing-vs-tool-search">Agent routing vs tool search →</a>
    · <a href="/blog/2025-12-02-intelligent-model-selection">Intelligent model selection →</a>
    · <a href="/blog/2025-12-03-dependency-graph-context-optimization">Dependency graph →</a>
    · <a href="/blog/2025-12-04-chat-repl-conversational-coding">Chat REPL →</a>
  </p>
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
