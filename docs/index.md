---
layout: default
title: Chuchu
description: AI coding assistant with specialized agents and validation
---

<div class="hero">
  <h1>The Open Source<br/>AI Coding Assistant</h1>
  <p>Transparent, Autonomous, and Local. <strong>Bring Your Own Keys</strong>. Run with Groq, OpenAI, or locally with Ollama. No subscriptions, full control.</p>
  <div class="hero-cta">
    <a href="#quick-start" class="btn btn-primary">Get Started</a>
    <a href="https://github.com/jadercorrea/chuchu" class="btn btn-secondary">View on GitHub</a>
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
  <h2 class="section-title">How It Works</h2>
  <p class="section-subtitle">Orchestrated agents with validation and auto-retry</p>
  
  <div class="mermaid">
graph TB
    User["chu do 'add feature'"] --> Orchestrator{Orchestrator}
    
    Orchestrator --> Analyzer["Analyzer<br/>Understands codebase<br/>Reads relevant files"]
    Analyzer --> Planner["Planner<br/>Creates minimal plan<br/>Lists files to modify"]
    Planner --> Validation["File Validation<br/>Extracts allowed files<br/>Blocks extras"]
    Validation --> Editor["Editor<br/>Executes changes<br/>ONLY planned files"]
    Editor --> Validator["Validator<br/>Checks success criteria<br/>Validates results"]
    
    Validator -->|Success| Done["Task Complete"]
    Validator -->|Fail| Retry["Auto-retry<br/>with feedback"]
    Retry --> Editor
    
    style Analyzer fill:#3b82f6,color:#fff
    style Planner fill:#8b5cf6,color:#fff  
    style Editor fill:#10b981,color:#fff
    style Validator fill:#f59e0b,color:#fff
    style Validation fill:#ef4444,color:#fff
  </div>
</div>

<div class="section">
  <h2 class="section-title">Structured Workflow: Research → Plan → Implement</h2>
  <p class="section-subtitle">Go from feature idea to tested code with AI assistance at each step</p>
  
  <div class="workflow-steps">
    <div class="workflow-step">
      <h3>Research</h3>
      <p>Understand your codebase before making changes</p>
      <pre><code>chu research "How does authentication work?"</code></pre>
      <p>Chuchu searches semantically, reads relevant files, and documents findings in <code>~/.chuchu/research/</code></p>
    </div>
    
    <div class="workflow-step">
      <h3>Plan</h3>
      <p>Create detailed implementation plan with phases</p>
      <pre><code>chu plan "Add password reset feature"</code></pre>
      <p>Generates step-by-step plan with clear goals, file changes, and test requirements</p>
    </div>
    
    <div class="workflow-step">
      <h3>Implement</h3>
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
  
  <p>Most AI coding assistants lock you into expensive subscriptions ($20-30/month) with black-box model selection and no validation. Chuchu gives you:</p>
  
  <ul>
    <li><strong>Specialized agents</strong>: 4 agents working together with validation</li>
    <li><strong>Safety first</strong>: File validation + success criteria prevent mistakes</li>
    <li><strong>Full control</strong>: Supervised vs autonomous modes, any OpenAI-compatible provider</li>
    <li><strong>Radically affordable</strong>: $0-5/month vs $20-30/month subscriptions</li>
    <li><strong>Local option</strong>: Run completely offline with Ollama for $0</li>
    <li><strong>TDD support</strong>: Write tests first when you want (<code>chu tdd</code>)</li>
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
