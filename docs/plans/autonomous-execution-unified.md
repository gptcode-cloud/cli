# Autonomous Multi-Step Task Execution (Unified Plan)

**Status:** IN PLANNING  
**Last Updated:** 2025-12-01  
**Depends On:** Maestro Phase 1 (Complete), Agent Tool Use Examples (Complete)

## Overview

This plan unifies `auto-task-execution.md` and `multi-step-execution.md` into a single autonomous execution system, incorporating insights from Anthropic's advanced tool use patterns.

### Key Clarifications

**Agent Routing vs. Tool Discovery:**
- **Agent Routing** (Chuchu) - ML classifier selects specialized agents (Analyzer, Planner, Editor, Validator)
- **Tool Discovery** (Anthropic) - On-demand loading of external API tools (GitHub, Slack, Jira)

**What Chuchu Has:**
- ✅ Agent routing (1ms ML classifier)
- ✅ Agent-specific tool sets (read_file, write_file, apply_patch, etc.)
- ✅ Tool Use Examples in prompts (2025-12-01)
- ✅ Validation gates and auto-recovery

**What We Can Adopt from Anthropic:**
- ⚠️ **Programmatic Tool Calling** - Let LLM write code to orchestrate tools (reduce context)
- ⏳ **Tool Search** - Only if we add 20+ external MCP servers (not immediate need)

---

## Problem Statement

Current limitations:
1. **Vague tasks fail**: `chu do "read docs/_posts/X.md and create features page"`
2. **Complex multi-step tasks hit max iterations**: Editor reaches limit without completing
3. **No decomposition**: Can't break "reorganize docs" into phases
4. **Limited context management**: Loads all files upfront, hits token limits

**Current Capabilities (2025-12-01):**
- ✅ `chu implement plan.md --auto` - executes plans with retry
- ✅ ML recommender - selects optimal models per agent
- ✅ Verification system - build + test + lint
- ✅ Error recovery with model switching
- ✅ Tool Use Examples in all agent prompts

**Missing:**
- ❌ Auto-decomposition of vague tasks into movements
- ❌ Programmatic tool orchestration (reduce context bloat)
- ❌ Multi-file context management with progressive disclosure
- ❌ Self-validation for content quality (not just tests)

---

## Solution Architecture

### Layer 1: Task Intelligence

**File:** `internal/autonomous/analyzer.go`

```go
type TaskAnalysis struct {
    Intent           string          // "create", "unify", "refactor", "reorganize"
    RequiredFiles    []string        // Files mentioned or discovered
    OutputFiles      []string        // Files to create/modify
    Movements        []Movement      // De composed phases
    Complexity       int             // 1-10
    RequiresPlan     bool
    CanAutoExecute   bool
    ProgrammaticMode bool            // Use code orchestration vs. LLM calls
}

type Movement struct {
    ID              string
    Name            string
    Goal            string
    Dependencies    []string         // Movement IDs that must complete first
    SuccessCriteria []string
    EstimatedTokens int
}

func (a *TaskAnalyzer) Analyze(ctx context.Context, task string) (*TaskAnalysis, error) {
    // 1. Extract intent (create/read/update/delete/refactor/unify/reorganize)
    // 2. Discover files (explicit + implicit patterns like "all feature files")
    // 3. Estimate complexity
    // 4. Decompose into movements if complexity > 7
    // 5. Decide execution mode (direct, programmatic, or decomposed)
}
```

**Capabilities:**
- Detect verbs and extract file patterns
- **Decomposition**: Complex tasks → movements (phases)
- **Mode selection**: 
  - Simple task (complexity \u003c 5) → direct execution
  - Data-heavy task → programmatic mode
  - Very complex (complexity \u003e 7) → movement-based

---

### Layer 2: Movement-Based Execution (Symphony Pattern)

**File:** `internal/autonomous/symphony.go`

```go
type Symphony struct {
    ID              string
    Name            string
    Goal            string
    Movements       []Movement
    CurrentMovement int
    StartTime       time.Time
    CompletedAt     *time.Time
}

type SymphonyExecutor struct {
    analyzer  *TaskAnalyzer
    editor    *EditorAgent
    validator *SelfValidator
}

func (s *SymphonyExecutor) Execute(ctx context.Context, symphony *Symphony) error {
    for i, movement := range symphony.Movements {
        // 1. Check dependencies completed
        if !s.dependenciesMet(movement) {
            return fmt.Errorf("dependencies not met for movement %s", movement.ID)
        }
        
        // 2. Execute movement
        symphony.CurrentMovement = i
        err := s.executeMovement(ctx, &movement)
        if err != nil {
            // Try recovery
            if recovered := s.recoverMovement(ctx, &movement, err); !recovered {
                return err
            }
        }
        
        // 3. Validate before next movement
        if !s.validator.ValidateMovement(ctx, &movement) {
            return fmt.Errorf("movement validation failed: %s", movement.Name)
        }
        
        // 4. Save progress (enable resume)
        s.saveCheckpoint(symphony)
    }
    
    return nil
}
```

**Benefits:**
- **Isolation**: Each movement is independent, validated separately
- **Resumable**: Can resume from checkpoint if interrupted
- **Clear progress**: User sees "Movement 2/5 complete"

---

### Layer 3: Programmatic Tool Calling (Anthropic Pattern)

**File:** `internal/autonomous/programmatic.go`

Adopt Anthropic's pattern: Let LLM write **code** to orchestrate tools, keep intermediates out of context.

**Use Cases:**
1. **Analyzer reading 50+ files**: Write code to load files, extract dependencies, return only graph
2. **Validator running 100 tests**: Write code to parse output, return only failures
3. **Multi-file operations**: Write code to process all files, return only summary

```go
type ProgrammaticExecutor struct {
    codeRunner *CodeRunner // Sandboxed Python/JS execution
    tools      map[string]Tool
}

func (p *ProgrammaticExecutor) ExecuteWithCode(ctx context.Context, task string) (string, error) {
    // 1. LLM writes orchestration code
    prompt := fmt.Sprintf(`
Task: %s

Available tools:
- read_file(path) → string
- list_files(pattern) → []string
- run_tests(pattern) → {passed, failed, output}

Write Python code to:
1. Orchestrate these tools
2. Process results
3. Return only final summary (not all intermediate data)

Example:
    files = list_files("docs/**/*.md")
    contents = [read_file(f) for f in files]
    # Process...
    print(json.dumps(summary))  # Only summary enters context
`, task)

    code := p.llm.GenerateCode(ctx, prompt)
    
    // 2. Run code in sandbox
    result := p.codeRunner.Run(code, p.tools)
    
    // 3. Only result enters context
    return result.Output, result.Error
}
```

**Impact:**
- 50+ file reads: 100K tokens → 5K tokens (summary only)
- 100 test runs: 200K output → 2K tokens (failures only)
- Matches Anthropic's 37% token reduction

---

## Implementation Timeline

### Phase 1: Task Intelligence (2 weeks)
- [ ] Task analyzer with intent extraction
- [ ] File pattern discovery ("all feature files" → glob + semantic filter)
- [ ] Movement decomposition (complexity \u003e 7)
- [ ] Safety checks (block destructive tasks)

### Phase 2: Symphony Execution (2 weeks)
- [ ] Symphony struct and persistence
- [ ] Movement executor with dependency resolution
- [ ] Self-validator (structural + content)
- [ ] Checkpoint/resume capability

### Phase 3: Programmatic Tool Calling (1 week)
- [ ] Code runner (sandboxed Python execution)
- [ ] Tool bridge (read_file, list_files, run_tests)
- [ ] LLM code generation prompt

### Phase 4: Context Management (1 week)
- [ ] Context builder with token budgets
- [ ] Summarization for large files
- [ ] Movement-scoped context loading

### Phase 5: CLI + Integration (1 week)
- [ ] `chu do` command
- [ ] `--dry-run`, `--programmatic`, `--resume` flags
- [ ] Integration with ML recommender

**Total: 7-8 weeks**

---

## Comparison with Anthropic

| Feature | Anthropic | Chuchu (Proposed) |
|---------|-----------|-------------------|
| **Discovery** | Tool Search (50+ API tools) | Agent Routing + Movement Decomposition |
| **Orchestration** | Programmatic Tool Calling (Python) | Symphony Pattern + Programmatic Mode |
| **Examples** | Tool Use Examples | ✅ Implemented (2025-12-01) |
| **Context Reduction** | 85% (tool defer + code) | 80% (agent routing + programmatic) |
| **Use Case** | Many external APIs | Complex multi-file code tasks |

**Key Differences:**
- Anthropic: Single mega-agent + 50+ external tools
- Chuchu: Specialized agents + movement decomposition

**Shared Patterns:**
- ✅ Tool Use Examples → Already adopted
- ✅ Programmatic orchestration → Implementing (Phase 3)
- ⏳ Tool Search → Only if we add MCP servers (future)

---

## Next Steps

1. Validate approach with user
2. Create GitHub milestone "Autonomous Symphony"
3. Start with Phase 1 (Task Intelligence)
4. Test with real use case (docs reorganization)
5. Measure token savings from programmatic mode
