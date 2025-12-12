# Complete Workflow Guide: From Idea to Implementation

This guide shows you how to use GPTCode's complete workflow: **research → plan → implement**.

## Overview

GPTCode helps you implement features through a structured, AI-assisted process:

1. **Research** - Understand your codebase and gather context
2. **Plan** - Create a detailed implementation plan
3. **Implement** - Execute the plan (interactive or autonomous)

## Step 1: Research

Use research mode to explore your codebase and understand the context before making changes.

### When to use Research

- Learning how a feature works
- Understanding architecture before making changes
- Exploring dependencies and relationships
- Documenting existing code

### Example

```bash
chu research "How does user authentication work in this codebase?"
```

**What happens:**
- Semantic search finds relevant files
- Reads and analyzes code structure
- Can search web for best practices
- Creates a research document in `~/.gptcode/research/`

**Output:**
```
Research saved to: ~/.gptcode/research/2025-01-23-authentication-research.md

Key findings:
- Authentication uses JWT tokens
- Middleware in auth/middleware.go
- User model in models/user.go
- Tests in auth/auth_test.go
```

## Step 2: Plan

Create a detailed implementation plan based on your research.

### When to use Plan

- Before implementing new features
- For complex refactoring
- When you need review/approval
- To break down large tasks

### Example

```bash
chu plan "Add OAuth2 support to authentication"
```

**What happens:**
- Reviews research (if available)
- Analyzes current codebase
- Creates step-by-step implementation plan
- Breaks down into phases with clear goals
- Saves to `~/.gptcode/plans/`

**Output:**
```
Plan saved to: ~/.gptcode/plans/2025-01-23-oauth2-support.md

Plan overview:
1. Add OAuth2 library dependency
2. Create OAuth2 configuration
3. Implement OAuth2 handlers
4. Update middleware
5. Add tests
6. Update documentation
```

## Step 3: Implement

Execute your plan - choose between interactive or autonomous mode.

### Mode A: Interactive (Default)

**Best for:**
- Learning the codebase
- Complex/sensitive changes
- When you want control over each step
- Reviewing changes before proceeding

**Usage:**
```bash
chu implement ~/.gptcode/plans/2025-01-23-oauth2-support.md
```

**What happens:**
```
📋 Plan loaded: 5 steps

─── Step 1/5: Add OAuth2 library dependency ───

Add go-oauth2/oauth2 library to go.mod...

Execute this step? [Y/n/q]: Y
✓ Step completed

─── Step 2/5: Create OAuth2 configuration ───

Create config/oauth2.go with provider configuration...

Execute this step? [Y/n/q]: n
⊘ Skipped

─── Step 3/5: Implement OAuth2 handlers ───
...
```

**Options:**
- `Y` or Enter - Execute the step
- `n` - Skip this step
- `q` - Quit/cancel implementation

**On errors:**
```
✗ Step failed: build error

Continue anyway? [y/N]: n
```

### Mode B: Autonomous (--auto)

**Best for:**
- Well-defined plans
- Trusted AI agents
- Faster iteration
- Batch processing

**Usage:**
```bash
chu implement ~/.gptcode/plans/2025-01-23-oauth2-support.md --auto
```

**What happens:**
- Executes all steps automatically
- Verifies each step with build + tests
- Retries on errors (up to 3 times)
- Creates checkpoints after success
- Rolls back on failure

**With options:**
```bash
# Enable lint verification
chu implement plan.md --auto --lint

# Custom retry limit
chu implement plan.md --auto --max-retries 5

# Resume from checkpoint
chu implement plan.md --auto --resume
```

**Output:**
```
🚀 Starting autonomous execution...

Step 1/5: Add OAuth2 library dependency
✓ Verification passed, saving checkpoint...

Step 2/5: Create OAuth2 configuration
✓ Verification passed, saving checkpoint...

Step 3/5: Implement OAuth2 handlers
⚠  Verification failed: build error
Error type: build, attempting recovery...
Retry 1/3
✓ Verification passed, saving checkpoint...

✓ Execution completed successfully!
```

## Complete Example: Adding a New Feature

Let's walk through adding a "password reset" feature:

### 1. Research the codebase

```bash
chu research "How is user authentication currently implemented?"
```

Review the generated research document to understand:
- Current auth flow
- Where to add the reset feature
- Existing patterns to follow

### 2. Create a plan

```bash
chu plan "Add password reset feature with email verification"
```

Review the generated plan at `~/.gptcode/plans/2025-01-23-password-reset.md`:

```markdown
# Password Reset Feature Plan

## Phase 1: Database Changes
- Add reset_token and token_expiry to users table
- Create migration

## Phase 2: Email Service
- Setup email configuration
- Create password reset email template
- Implement send_reset_email function

## Phase 3: API Endpoints
- POST /auth/forgot-password (request reset)
- POST /auth/reset-password (submit new password)
- Add validation and rate limiting

## Phase 4: Tests
- Unit tests for token generation
- Integration tests for email flow
- End-to-end tests for reset process
```

### 3. Implement interactively (first time)

```bash
chu implement ~/.gptcode/plans/2025-01-23-password-reset.md
```

Walk through each phase, reviewing and confirming:
- See what will be changed
- Skip phases if needed
- Learn as you go

### 4. Implement autonomously (iteration/refinement)

After reviewing and adjusting the plan:

```bash
chu implement ~/.gptcode/plans/2025-01-23-password-reset.md --auto --lint
```

Let GPTCode:
- Execute all phases
- Verify each step
- Handle errors automatically
- Complete the implementation

## Tips for Best Results

### Research Phase
- Be specific in your questions
- Reference specific files/modules
- Ask multiple related questions
- Review research before planning

### Planning Phase  
- Include clear acceptance criteria
- Break down into small, testable steps
- Specify files to be created/modified
- Review and edit plan before implementing

### Implementation Phase
- **Interactive mode**: Use when learning or for sensitive changes
- **Autonomous mode**: Use for well-defined, tested plans
- Always review changes with `git diff` after
- Run full test suite before committing

## Neovim Integration

All modes work from Neovim:

```vim
" Research in chat
<C-d>
> research: How does authentication work?

" Plan from chat
> plan: Add password reset feature

" Implement autonomously
:GPTCodeAuto
" Or: <leader>ca
" Prompts for plan file, runs: chu implement <file> --auto
```

## FAQ

### When should I use interactive vs autonomous mode?

**Interactive:**
- First time implementing a feature
- Learning unfamiliar codebase
- High-risk/production changes
- Need to understand each step

**Autonomous:**
- Trusted, well-defined plans
- Repetitive implementations
- Rapid prototyping
- Batch processing multiple features

### What if implementation fails?

**Interactive mode:**
- Prompted to continue or stop
- Review error and adjust plan
- Re-run specific step

**Autonomous mode:**
- Automatic retry (3x by default)
- Error recovery with targeted fixes
- Rollback to last checkpoint
- Resume with `--resume`

### Can I edit the plan mid-implementation?

Yes! 
- Interactive: quit, edit plan, restart
- Autonomous: cancel (Ctrl+C), edit, resume with `--resume`

### How do I know if my plan is good?

Good plans have:
- ✅ Clear, single-responsibility steps
- ✅ Specific file paths and changes
- ✅ Test requirements for each phase
- ✅ Incremental, verifiable progress
- ✅ Rollback points (checkpoints)

Bad plans have:
- ❌ Vague "implement feature X"
- ❌ Too many changes in one step
- ❌ No verification criteria
- ❌ Circular dependencies

### What languages are supported?

Verification works for:
- Go
- TypeScript/JavaScript
- Python
- Elixir
- Ruby

Implementation works for any language (uses LLM), but verification is language-specific.

## Next Steps

1. Try the workflow with a small feature
2. Start with research to understand your codebase
3. Create a detailed plan
4. Use interactive mode first, then autonomous
5. Review results and iterate

**See also:**
- [Model Configuration Guide](https://jadercorrea.github.io/gptcode/blog/2025-11-18-groq-optimal-configs)
- [Cost Optimization Tips](https://jadercorrea.github.io/gptcode/blog/2024-11-22-cost-tracking-optimization)
- [Examples: docs/examples/](../examples/)
