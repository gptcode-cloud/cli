---
layout: default
title: Prompts
description: System prompts, workflows, and best practices for GPTCode agents
---

# Prompts & Workflows

GPTCode uses carefully crafted system prompts for each agent and workflow. Understanding these helps you get better results.

## Quick Navigation

<div class="prompt-nav">
  <a href="#research-mode">Research</a>
  <a href="#plan-mode">Plan</a>
  <a href="#agent-prompts">Agents</a>
  <a href="#best-practices">Best Practices</a>
  <a href="#common-patterns">Patterns</a>
  <a href="#customizing-prompts">Customize</a>
  <a href="#troubleshooting">Troubleshooting</a>
</div>

<style>
.prompt-nav {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  padding: 1rem;
  background: #16161e;
  border-radius: 8px;
  margin-bottom: 2rem;
  border: 1px solid #3b4261;
}
.prompt-nav a {
  padding: 0.5rem 1rem;
  background: #1a1b26;
  border: 1px solid #3b4261;
  border-radius: 4px;
  text-decoration: none;
  color: #c0caf5;
  font-weight: 500;
  transition: all 0.2s;
}
.prompt-nav a:hover {
  background: #bb9af7;
  color: #1a1b26;
  border-color: #bb9af7;
}
.copy-btn {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  padding: 0.25rem 0.5rem;
  background: #8b5cf6;
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
  background: #7c3aed;
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

## Research Mode

**Purpose**: Document codebase as-is without suggesting improvements.

### When to use
- Understanding how existing code works
- Finding where features are implemented
- Tracing data flow and dependencies
- Creating technical documentation

### Workflow
```bash
chu research "how does authentication work"
```

1. **Context gathering**: Reads mentioned files fully
2. **Parallel research**: Spawns specialized sub-agents
   - `codebase-locator`: Finds relevant files
   - `codebase-analyzer`: Explains how code works
   - `thoughts-locator`: Searches historical context
3. **Synthesis**: Combines findings with file:line references
4. **Documentation**: Generates research artifact

### Key principles
- **Document, don't critique**: Describe what IS, not what SHOULD BE
- **No recommendations**: Only explain current implementation
- **Concrete references**: Always include file paths and line numbers
- **Historical context**: Use `thoughts/` directory for past decisions

### Example output
```
## Findings

Authentication is handled in internal/auth/:

1. internal/auth/middleware.go:45-67
   - ValidateToken() checks JWT signature
   - Extracts user_id from claims
   
2. internal/auth/provider.go:123
   - GenerateToken() creates JWT with 24h expiry
   - Uses RS256 algorithm
```

[Full research prompt →](/research)

---

## Plan Mode

**Purpose**: Create detailed implementation plans through interactive iteration.

### When to use
- Starting new features
- Complex refactoring
- Architectural changes
- Multi-step implementations

### Workflow
```bash
chu plan "add JWT authentication"
```

1. **Initial research**: Gathers context about current state
2. **Interactive questioning**: Clarifies requirements and constraints
3. **Design options**: Presents approaches with pros/cons
4. **Phased planning**: Breaks work into implementable phases
5. **Validation**: Verifies plan against codebase reality

### Key principles
- **Be skeptical**: Question assumptions, verify with code
- **Research first**: Don't assume, read the actual implementation
- **Interactive**: Work with user to refine understanding
- **Concrete**: Include specific files, functions, test paths
- **Phased**: Break into testable increments

### Example output
```markdown
# JWT Authentication Implementation Plan

## Phase 1: Token Generation
Files to modify:
- internal/auth/provider.go - Add GenerateJWT()
- internal/models/user.go - Add RefreshToken field
Tests:
- internal/auth/provider_test.go - Token generation tests

## Phase 2: Middleware
Files to create:
- internal/auth/middleware.go - ValidateToken middleware
Tests:
- internal/auth/middleware_test.go - Auth middleware tests
```

[Full plan prompt →](/plan)

---

## Agent Prompts

### Router Agent

**Role**: Fast intent classification

**System prompt highlights:**
- Classify user intent: query, edit, research, review
- Return JSON with intent and confidence
- No explanations, just routing decision
- Fallback to LLM if ML model uncertain

**Model requirements:**
- Speed > quality (need <100ms response)
- Small context window ok (routing is simple)
- Recommended: Llama 3.1 8B Instant (840 TPS)

---

### Query Agent

**Role**: Code reading and analysis

**System prompt highlights:**
- Read and explain existing code
- Don't suggest changes unless asked
- Use concrete file:line references
- Trace data flow and dependencies
- Answer "how does X work"

**Model requirements:**
- Strong comprehension
- Larger context window (32k-128k)
- Recommended: GPT-OSS 120B, Qwen 2.5 Coder

**Example prompts:**
```
explain how authentication works
show me where user sessions are stored
trace the flow from HTTP request to database
```

---

### Editor Agent

**Role**: Code writing and modification

**System prompt highlights:**
- Write tests before implementation
- Keep functions small and focused
- Follow existing code patterns
- Include error handling
- Document assumptions

**Model requirements:**
- Best code generation quality
- Strong reasoning (for TDD)
- Recommended: DeepSeek R1 Distill, Qwen 2.5 Coder

**Example prompts:**
```
add rate limiting to login endpoint
implement JWT token refresh
refactor user validation to use helpers
```

---

### Research Agent

**Role**: Web search and documentation

**System prompt highlights:**
- Search web for best practices
- Find documentation and examples
- Summarize findings with links
- Compare different approaches

**Model requirements:**
- Large context window (for long docs)
- Good summarization
- Recommended: Grok 4.1 Fast (2M context, free)

**Example prompts:**
```
best practices for JWT implementation in Go
how do other projects handle rate limiting
compare bcrypt vs argon2 for passwords
```

---

## Best Practices

### Writing Good Prompts

**Be specific:**
```
❌ "fix the auth bug"
✅ "fix bug where JWT tokens expire too early in auth/provider.go"
```

**Provide context:**
```
❌ "add error handling"
✅ "add error handling to the database connection in main.go, similar to how redis.go handles it"
```

**State constraints:**
```
❌ "improve performance"
✅ "improve login performance without changing the database schema"
```

### Leveraging Agent Specialization

**Route to the right agent:**
- Questions about existing code → Query agent
- Writing new code → Editor agent
- External research needed → Research agent

**Example:**
```bash
# Understanding (Query)
chu chat "how does the current auth work"

# Research external approaches (Research)
chu research "JWT best practices"

# Planning (Plan)
chu plan "add JWT authentication"

# Implementation (Editor)
chu implement plan.md
```

### Iterative Refinement

Start broad, then narrow:
```bash
# 1. High-level understanding
chu research "authentication system"

# 2. Specific area
chu chat "show me the token validation logic"

# 3. Targeted change
chu chat "add refresh token support to the existing JWT validation"
```

---

## Common Patterns

### Pattern: Research → Plan → Implement

For new features:
```bash
# 1. Understand current state
chu research "current user management system"

# 2. Create detailed plan
chu plan "add role-based permissions"

# 3. Execute plan
chu implement permissions-plan.md
```

### Pattern: Query → Edit → Verify

For bug fixes:
```bash
# 1. Find the bug
chu chat "trace the login flow to find where session expires"

# 2. Fix it
chu chat "fix session expiry in auth/session.go"

# 3. Verify
chu chat "write tests for session expiry edge cases"
```

### Pattern: Research → Compare → Decide

For architectural decisions:
```bash
# 1. Research options
chu research "authentication libraries in Go"

# 2. Compare approaches
chu research "JWT vs session-based auth pros/cons"

# 3. Make informed decision
chu plan "implement JWT authentication using github.com/golang-jwt"
```

---

## Customizing Prompts

### System Prompts Location
- Research: `~/.gptcode/prompts/research.md`
- Plan: `~/.gptcode/prompts/plan.md`
- Agent prompts: Configured per-profile in `~/.gptcode/profiles/*.yaml`

### Override Example
```yaml
# ~/.gptcode/profiles/my-profile.yaml
backend:
  groq:
    agent_models:
      editor: qwen2.5-coder:32b
    agent_prompts:
      editor: |
        You are a TDD-focused code generator.
        Always write tests first.
        Keep functions under 50 lines.
        Use descriptive variable names.
```

---

## Troubleshooting

### Agent not understanding context
- Use dependency graph: Automatically includes relevant files
- Be explicit: Mention specific files/functions
- Check context window: Large codebases may hit limits

### Responses too generic
- Provide concrete examples from your codebase
- Reference existing patterns to follow
- Use Research mode first to gather context

### Wrong agent selected
- ML classifier has 89% accuracy, uses LLM fallback
- Adjust threshold: `chu config set defaults.ml_intent_threshold 0.8`
- Or specify explicitly: `chu chat --agent query "your question"`

---

## Related Resources

- [Full Research Prompt](/research) - Complete system prompt
- [Full Plan Prompt](/plan) - Complete system prompt
- [Commands Reference](/commands) - All CLI commands
- [Features](/features) - Complete feature list
