# Plan: Autonomous Multi-Step Task Execution

## Problem Statement

Chuchu can select optimal models, execute plans with retry logic, and collect feedback. However, it cannot autonomously handle complex implicit tasks like:

```bash
chu "read docs/_posts/2025-11-25-ensemble-optimization.md and create a features page"
chu "unify all feature files in /guides"
```

These tasks require:
- Implicit requirement discovery (what files? what structure?)
- Multi-file context management
- Autonomous decision-making (structure, naming, organization)
- Self-validation without explicit tests
- End-to-end execution without approval steps

**Current State:**
- ✅ `chu auto plan.md` - executes existing plans with retry
- ✅ Guided mode - creates plans but requires approval
- ✅ Implement mode - edits files but needs explicit instructions
- ✅ ML recommender - selects optimal models
- ❌ Autonomous task decomposition from vague input
- ❌ Implicit requirement discovery
- ❌ Self-directed multi-file operations
- ❌ Content-aware restructuring

## Proposed Solution

### Phase 1: Task Intelligence Layer

Add cognitive capabilities to transform vague tasks into actionable plans.

#### 1.1 Task Analyzer

**File:** `internal/autonomous/analyzer.go`

```go
type TaskAnalysis struct {
    Intent           string   // "create", "unify", "refactor", etc.
    RequiredFiles    []string // Files that must be read
    OutputFiles      []string // Files to create/modify
    ImplicitSteps    []Step   // Discovered sub-tasks
    Complexity       int      // 1-10 scale
    RequiresPlan     bool
    CanAutoExecute   bool
}

type TaskAnalyzer struct {
    llm *llm.Client
}

func (a *TaskAnalyzer) Analyze(ctx context.Context, task string) (*TaskAnalysis, error) {
    prompt := fmt.Sprintf(`
Analyze this task and extract:
1. Primary intent (create/read/update/delete/refactor/unify)
2. Files mentioned or implied
3. Hidden requirements (e.g., "unify" implies: list files, read all, decide structure, create unified file)
4. Complexity (1-10)
5. Can this be executed without user approval?

Task: %s

Return JSON with: intent, required_files, output_files, implicit_steps, complexity, can_auto_execute
`, task)
    
    response := a.llm.Analyze(ctx, prompt)
    return parseTaskAnalysis(response)
}
```

**Capabilities:**
- Detect verbs (read, create, unify, refactor)
- Extract file patterns from natural language
- Discover implicit steps (list → read → analyze → create)
- Assess if task is safe for auto-execution

**Examples:**
```
Input: "read post X and create features page"
Output:
  Intent: create
  RequiredFiles: [docs/_posts/2025-11-25-ensemble-optimization.md]
  OutputFiles: [inferred from content]
  ImplicitSteps:
    1. Read source file
    2. Extract key features
    3. Design page structure
    4. Create markdown file
  Complexity: 6
  CanAutoExecute: true

Input: "unify all feature files in /guides"
Output:
  Intent: unify
  RequiredFiles: [guides/**/*.md]
  OutputFiles: [guides/unified-features.md]
  ImplicitSteps:
    1. List all files in /guides
    2. Read each file
    3. Analyze common structure
    4. Merge content
    5. Create unified file
  Complexity: 8
  CanAutoExecute: true
```

#### 1.2 Requirement Discovery

**File:** `internal/autonomous/discovery.go`

```go
type RequirementDiscovery struct {
    fs  *FileSystem
    llm *llm.Client
}

func (d *RequirementDiscovery) DiscoverFiles(ctx context.Context, pattern string) ([]string, error) {
    // "all feature files in /guides" → glob pattern + semantic filter
    
    // 1. Convert natural language to glob
    glob := d.naturalLanguageToGlob(pattern)
    
    // 2. List candidate files
    candidates := d.fs.Glob(glob)
    
    // 3. Semantic filter (is this really a "feature file"?)
    return d.semanticFilter(ctx, candidates, pattern)
}

func (d *RequirementDiscovery) InferStructure(ctx context.Context, files []string) (*Structure, error) {
    // Read all files, analyze common patterns, propose unified structure
    contents := d.fs.ReadAll(files)
    
    prompt := fmt.Sprintf(`
These files need to be unified:
%s

Analyze and propose:
1. Common sections across files
2. Unified structure
3. How to merge without losing information

Return JSON with: sections, merge_strategy
`, formatContents(contents))
    
    return d.llm.Infer(ctx, prompt)
}
```

**Capabilities:**
- Convert "all feature files" → `guides/**/*feature*.md`
- Semantic filtering (is this a feature file?)
- Structure inference from multiple files
- Merge strategy proposal

#### 1.3 Autonomous Executor

**File:** `internal/autonomous/executor.go`

```go
type AutonomousExecutor struct {
    analyzer  *TaskAnalyzer
    discovery *RequirementDiscovery
    maestro   *maestro.Maestro
    validator *SelfValidator
}

func (e *AutonomousExecutor) Execute(ctx context.Context, task string) error {
    // 1. Analyze task
    analysis := e.analyzer.Analyze(ctx, task)
    
    if !analysis.CanAutoExecute {
        return fmt.Errorf("task requires user approval - use 'chu guided' instead")
    }
    
    // 2. Discover requirements
    if hasPattern(analysis.RequiredFiles) {
        files := e.discovery.DiscoverFiles(ctx, analysis.RequiredFiles[0])
        analysis.RequiredFiles = files
    }
    
    // 3. Create execution plan
    plan := e.createPlan(ctx, analysis)
    
    // 4. Execute with self-validation loop
    for attempt := 1; attempt <= 3; attempt++ {
        err := e.executeSteps(ctx, plan.Steps)
        if err != nil {
            plan = e.adjustPlan(ctx, plan, err)
            continue
        }
        
        // 5. Self-validate
        if e.validator.Validate(ctx, plan, analysis.Intent) {
            return nil
        }
        
        // 6. Self-reflect and adjust
        reflection := e.reflect(ctx, plan, analysis)
        plan = e.applyReflection(plan, reflection)
    }
    
    return fmt.Errorf("failed after 3 attempts")
}
```

**Flow:**
1. Task analysis (intent extraction)
2. Requirement discovery (implicit file patterns)
3. Plan creation (automatic, no approval)
4. Execution loop with validation
5. Self-reflection if validation fails

#### 1.4 Self-Validator

**File:** `internal/autonomous/validator.go`

```go
type SelfValidator struct {
    llm *llm.Client
    fs  *FileSystem
}

type ValidationResult struct {
    Success bool
    Issues  []string
    Score   float64 // 0-1
}

func (v *SelfValidator) Validate(ctx context.Context, plan *Plan, intent string) bool {
    result := ValidationResult{Success: true, Score: 1.0}
    
    // Check 1: Files exist
    for _, file := range plan.OutputFiles {
        if !v.fs.Exists(file) {
            result.Issues = append(result.Issues, "missing file: "+file)
            result.Success = false
        }
    }
    
    // Check 2: Content quality (LLM-based)
    for _, file := range plan.OutputFiles {
        content := v.fs.Read(file)
        quality := v.assessQuality(ctx, content, intent)
        result.Score *= quality
        
        if quality < 0.7 {
            result.Issues = append(result.Issues, fmt.Sprintf("low quality: %s (%.2f)", file, quality))
            result.Success = false
        }
    }
    
    // Check 3: Intent fulfilled
    if !v.checkIntent(ctx, plan, intent) {
        result.Issues = append(result.Issues, "intent not fulfilled")
        result.Success = false
    }
    
    return result.Success
}

func (v *SelfValidator) assessQuality(ctx context.Context, content string, intent string) float64 {
    prompt := fmt.Sprintf(`
Rate content quality (0-1) for intent: %s

Content:
%s

Criteria:
- Completeness
- Structure
- Clarity
- Correctness

Return only a number 0-1.
`, intent, content)
    
    return v.llm.Score(ctx, prompt)
}
```

**Validation Levels:**
1. **Structural:** Files exist, valid format
2. **Content:** Quality assessment via LLM
3. **Intent:** Did it achieve the goal?

#### 1.5 Self-Reflection Loop

**File:** `internal/autonomous/reflection.go`

```go
type Reflection struct {
    WhatWentWrong string
    WhyItFailed   string
    HowToFix      string
    Adjustments   []Adjustment
}

type Adjustment struct {
    Step   int
    Change string
}

func (e *AutonomousExecutor) reflect(ctx context.Context, plan *Plan, analysis *TaskAnalysis) *Reflection {
    prompt := fmt.Sprintf(`
Task: %s
Plan executed: %v
Intent: %s

What went wrong?
Why did validation fail?
How should I adjust the plan?

Return JSON with: what_went_wrong, why_it_failed, how_to_fix, adjustments: [{step: N, change: "..."}]
`, analysis.Intent, plan.Steps, analysis.Intent)
    
    return e.llm.Reflect(ctx, prompt)
}

func (e *AutonomousExecutor) applyReflection(plan *Plan, reflection *Reflection) *Plan {
    newPlan := plan.Clone()
    
    for _, adj := range reflection.Adjustments {
        if adj.Step < len(newPlan.Steps) {
            newPlan.Steps[adj.Step].Instruction += "\n\nAdjustment: " + adj.Change
        }
    }
    
    return newPlan
}
```

**Capabilities:**
- Analyze failure reasons
- Propose specific adjustments
- Modify plan without starting over
- Learn from mistakes within session

### Phase 2: CLI Integration

#### 2.1 New Command: `chu do`

**File:** `cmd/chu/do.go`

```bash
# Auto-execute vague tasks
chu do "read post X and create features page"
chu do "unify all feature files in /guides"

# With dry-run
chu do "refactor all markdown files" --dry-run

# With verbosity
chu do "create changelog from commits" --verbose
```

**Flags:**
- `--dry-run`: Show plan without executing
- `--verbose`: Show all steps and validations
- `--max-attempts <n>`: Override default retry limit (default: 3)
- `--no-validate`: Skip self-validation (risky)

**Implementation:**
```go
func runDo(cmd *cobra.Command, args []string) error {
    task := strings.Join(args, " ")
    
    executor := autonomous.NewExecutor(config)
    
    if dryRun {
        analysis := executor.Analyze(task)
        plan := executor.CreatePlan(analysis)
        fmt.Printf("Plan:\n%s\n", plan.Format())
        return nil
    }
    
    return executor.Execute(context.Background(), task)
}
```

#### 2.2 Safety Guards

**File:** `internal/autonomous/safety.go`

```go
type SafetyCheck struct {
    IsDestructive bool
    AffectsCount  int
    RiskLevel     string // "low", "medium", "high"
}

func (s *SafetyChecker) Check(analysis *TaskAnalysis) (*SafetyCheck, error) {
    check := &SafetyCheck{RiskLevel: "low"}
    
    // Destructive verbs
    if contains(analysis.Intent, []string{"delete", "remove", "destroy"}) {
        check.IsDestructive = true
        check.RiskLevel = "high"
    }
    
    // Many files affected
    if len(analysis.RequiredFiles) > 10 {
        check.AffectsCount = len(analysis.RequiredFiles)
        check.RiskLevel = "medium"
    }
    
    // Block high-risk tasks
    if check.RiskLevel == "high" {
        return nil, fmt.Errorf("high-risk task - use 'chu guided' for review")
    }
    
    return check, nil
}
```

**Blocked patterns:**
- `delete all`
- `remove everything`
- `refactor > 10 files` (requires approval)

### Phase 3: Context Management

#### 3.1 Multi-File Context Builder

**File:** `internal/autonomous/context.go`

```go
type ContextBuilder struct {
    maxTokens int
}

func (c *ContextBuilder) BuildContext(files []string) (*Context, error) {
    total := 0
    context := &Context{}
    
    for _, file := range files {
        content := readFile(file)
        tokens := countTokens(content)
        
        if total+tokens > c.maxTokens {
            // Summarize instead of including full content
            summary := c.summarize(content)
            context.Add(file, summary)
        } else {
            context.Add(file, content)
            total += tokens
        }
    }
    
    return context, nil
}

func (c *ContextBuilder) summarize(content string) string {
    // Use LLM to create concise summary
    // Or extract headings + first paragraph
    return extractKeyPoints(content)
}
```

**Strategy:**
- Full content for small files (< 500 tokens)
- Summaries for large files
- Semantic compression for multiple files

#### 3.2 Progressive Disclosure

**File:** `internal/autonomous/disclosure.go`

```go
func (e *AutonomousExecutor) executeWithProgressive(ctx context.Context, steps []Step) error {
    context := NewContext()
    
    for i, step := range steps {
        // Only load context needed for this step
        stepContext := e.buildStepContext(step, context)
        
        err := e.executeStep(ctx, step, stepContext)
        if err != nil {
            return err
        }
        
        // Update context with results
        context.AddStepResult(i, step.Output)
    }
    
    return nil
}
```

**Benefit:** Don't load all files upfront - load as needed per step

### Phase 4: Integration with Existing Systems

#### 4.1 Connect to ML Recommender

```go
func (e *AutonomousExecutor) selectAgent(analysis *TaskAnalysis) (string, error) {
    // Use ML recommender to select optimal agent
    backend := recommender.GetRecommendedForAgentWithTask(
        "autonomous",
        analysis.Intent,
    )
    return backend, nil
}
```

#### 4.2 Auto-Feedback Recording

```go
func (e *AutonomousExecutor) Execute(ctx context.Context, task string) error {
    start := time.Now()
    analysis := e.analyzer.Analyze(ctx, task)
    
    backend := e.selectAgent(analysis)
    err := e.executeWithBackend(ctx, analysis, backend)
    
    // Record feedback for ML training
    feedback.RecordTaskExecution(feedback.TaskExecution{
        Task:      task,
        Agent:     "autonomous",
        Backend:   backend,
        Duration:  time.Since(start),
        Success:   err == nil,
        Timestamp: start,
    })
    
    return err
}
```

#### 4.3 Knowledge Graph Query

```go
func (e *AutonomousExecutor) enrichAnalysis(ctx context.Context, analysis *TaskAnalysis) error {
    // Query knowledge graph for related concepts
    for _, file := range analysis.RequiredFiles {
        related := knowledge.Query(file, 5)
        analysis.RelatedConcepts = append(analysis.RelatedConcepts, related...)
    }
    return nil
}
```

### Phase 5: Neovim Integration

#### 5.1 New Command: `:ChuchuDo`

**File:** `neovim/lua/chuchu/autonomous.lua`

```lua
function M.do_task()
  local task = vim.fn.input("Task: ")
  if task == "" then return end
  
  local cmd = string.format("chu do '%s'", task)
  
  -- Execute in background with status window
  local bufnr = vim.api.nvim_create_buf(false, true)
  local win = vim.api.nvim_open_win(bufnr, true, {
    relative = "editor",
    width = 60,
    height = 15,
    row = 5,
    col = 10,
    border = "rounded",
    title = " Chuchu Autonomous ",
  })
  
  vim.fn.jobstart(cmd, {
    on_stdout = function(_, data)
      vim.api.nvim_buf_set_lines(bufnr, -1, -1, false, data)
    end,
    on_exit = function(_, code)
      if code == 0 then
        vim.api.nvim_buf_set_lines(bufnr, -1, -1, false, {"✅ Task completed"})
      else
        vim.api.nvim_buf_set_lines(bufnr, -1, -1, false, {"❌ Task failed"})
      end
    end,
  })
end

vim.api.nvim_create_user_command("ChuchuDo", M.do_task, {})
```

**Usage in Neovim:**
```
:ChuchuDo
Task: unify all feature files in /guides
[Shows real-time progress in floating window]
```

## Implementation Timeline

### Week 1-2: Task Intelligence
- Task analyzer with intent extraction
- Requirement discovery (file pattern resolution)
- Safety checks
- Unit tests

### Week 3-4: Execution Loop
- Autonomous executor with retry logic
- Self-validator (structural + content)
- Self-reflection system
- Integration tests

### Week 5: Context Management
- Multi-file context builder
- Progressive disclosure
- Token budget management

### Week 6: CLI + Neovim
- `chu do` command
- Neovim `:ChuchuDo`
- Integration with ML recommender
- Auto-feedback recording

### Week 7: Testing + Documentation
- End-to-end tests with real tasks
- Blog post: "Autonomous Task Execution"
- Update README
- Video demo

**Total: 7 weeks**

## Success Criteria

**Functional:**
- ✅ `chu do "read X and create Y"` works end-to-end
- ✅ File pattern discovery (e.g., "all feature files")
- ✅ Self-validation catches errors
- ✅ Self-reflection improves on retry
- ✅ Safety guards block destructive tasks

**Quality:**
- Success rate > 75% on first attempt
- Success rate > 90% with retries
- Average execution time < 3 minutes
- No false positives on safety checks

**Integration:**
- Works with existing `chu auto` (plan execution)
- Uses ML recommender for backend selection
- Records feedback for training
- Queries knowledge graph for context

## Comparison: Guided vs Auto vs Do

| Feature | Guided Mode | Auto Mode | Do Mode (New) |
|---------|-------------|-----------|---------------|
| Input | User query | Plan file | Natural language task |
| Planning | Creates plan, waits approval | Executes existing plan | Auto-creates plan |
| Execution | After approval | Immediate with retry | Immediate with validation |
| File discovery | Manual | From plan | Automatic |
| Self-validation | No | Build + test only | Content + intent |
| Use case | Complex new features | Execute reviewed plans | Vague multi-file tasks |

**Example Tasks:**

```bash
# Guided: User needs to review plan first
chu guided "implement authentication system"

# Auto: Execute pre-approved plan
chu auto docs/plans/add-auth.md

# Do: Autonomous execution of clear task
chu do "read blog post and create features page"
chu do "unify all feature files"
```

## Next Steps

1. Create GitHub issues for each component
2. Setup milestone "Autonomous Task Execution"
3. Begin with Task Analyzer (1.1)
4. Iterate with real use cases:
   - "Read X and create Y"
   - "Unify files in directory"
   - "Extract concepts from blog posts"
5. Collect metrics on success rate
6. Adjust self-validation thresholds based on data

## Dependencies

**Existing (Ready):**
- ✅ Maestro core loop (retry logic)
- ✅ Agents (editor, research)
- ✅ ML recommender
- ✅ Feedback system
- ✅ Knowledge graph

**New (To Build):**
- ❌ Task analyzer
- ❌ Requirement discovery
- ❌ Self-validator
- ❌ Self-reflection
- ❌ Context builder
- ❌ CLI `chu do`

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Misinterprets vague tasks | High | Clear error messages, suggest `chu guided` |
| Overwrites files accidentally | Critical | Git stash before execution, rollback on failure |
| Infinite reflection loop | Medium | Hard limit on attempts (3) |
| Context overflow | Medium | Progressive disclosure, summarization |
| LLM hallucination in analysis | High | Validation of discovered files (must exist) |

## Related Documents

- `maestro-autonomous-execution-plan.md` - Plan execution with retry
- `internal/intelligence/INTEGRATION_COMPLETE.md` - ML recommender integration
- `docs/_posts/2025-11-25-ensemble-optimization.md` - Intelligence layers overview
