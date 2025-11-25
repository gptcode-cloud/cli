# Multi-Step Task Execution - Maestro Pattern

## Problema Atual

O `chu do` tenta executar tarefas complexas de uma vez, resultando em:
- Editor atingindo max iterations sem completar
- Mudanças inconsistentes ou incorretas
- Criação de arquivos não solicitados
- Falta de validação entre etapas

## Solução: Maestro Pattern (Symphony Execution)

Inspirado em uma sinfonia com múltiplos movimentos, o chu deve:

1. **Análise e Decomposição** (Maestro)
   - Entender a tarefa completa
   - Quebrar em "movements" (etapas independentes)
   - Definir ordem de execução
   - Estabelecer critérios de sucesso para cada movimento

2. **Execução de Movements** (Orchestra)
   - Executar um movement por vez
   - Validar sucesso antes de próximo
   - Permitir feedback/ajuste entre movements
   - Registrar progresso e aprendizados

3. **Validação e Conclusão** (Finale)
   - Revisar todas as mudanças
   - Confirmar objetivos atingidos
   - Gerar relatório de execução

## Arquitetura Proposta

### 1. Movement Definition

```go
type Movement struct {
    ID          string
    Name        string
    Description string
    Goal        string
    Dependencies []string  // IDs of movements that must complete first
    
    // Execution
    Plan        string
    Status      MovementStatus // pending, executing, completed, failed
    
    // Results
    FilesModified []string
    FilesCreated  []string
    Error         string
    
    // Validation
    SuccessCriteria []string
    ValidationSteps []string
}

type MovementStatus string

const (
    MovementPending   MovementStatus = "pending"
    MovementExecuting MovementStatus = "executing"
    MovementCompleted MovementStatus = "completed"
    MovementFailed    MovementStatus = "failed"
)
```

### 2. Maestro Agent

```go
type MaestroAgent struct {
    orchestrator llm.Provider
    cwd          string
}

func (m *MaestroAgent) Decompose(ctx context.Context, task string) (*Symphony, error) {
    // 1. Analyze task
    // 2. Identify independent subtasks
    // 3. Create movements with dependencies
    // 4. Return Symphony with execution plan
}

type Symphony struct {
    Name      string
    Goal      string
    Movements []Movement
    
    // Execution tracking
    CurrentMovement int
    StartTime       time.Time
    CompletedAt     *time.Time
}
```

### 3. Movement Executor

```go
func (m *MaestroAgent) ExecuteMovement(ctx context.Context, movement *Movement) error {
    // 1. Generate detailed plan for this movement ONLY
    // 2. Execute with editor agent (base provider, not orchestrator)
    // 3. Validate success criteria
    // 4. Update movement status
    // 5. Return error if validation fails
}
```

### 4. CLI Integration

```bash
# Auto mode - full symphony execution
chu do "reorganize docs" --auto

# Interactive mode - execute movement by movement
chu do "reorganize docs"
# Output:
# 🎼 Symphony: Documentation Reorganization
# 
# Movements:
# 1. [pending] Analyze current structure and create index
# 2. [pending] Split features.md into individual pages  
# 3. [pending] Split commands.md into categorized pages
# 4. [pending] Create navigation structure
# 5. [pending] Update Jekyll configuration
#
# Execute movement 1? [Y/n/skip/quit]

# Resume interrupted symphony
chu do --resume <symphony-id>
```

## Exemplo: Reorganização de Docs

### Movement 1: Analyze and Index
**Goal**: Understand current doc structure and create content index

**Plan**:
- Read all docs/*.md files
- Categorize by type (feature, command, guide, blog)
- Create inventory of what exists
- Identify gaps and overlaps
- Save to ~/.chuchu/symphonies/docs-reorg/movement-1.json

**Success Criteria**:
- Inventory file exists
- All docs files counted
- Categories identified

**Files Created**: `~/.chuchu/symphonies/docs-reorg/inventory.json`

### Movement 2: Features Pages
**Goal**: Break features.md into individual feature pages

**Dependencies**: Movement 1

**Plan**:
- Read current features.md
- Identify distinct features (TDD, Multi-Agent, Cost Optimization, etc.)
- Create docs/features/ directory
- Create one .md per feature
- Update index.md links

**Success Criteria**:
- docs/features/ directory exists
- 6+ feature files created
- Each file has proper front matter
- index.md links work

**Files Created**: 
- docs/features/tdd-workflow.md
- docs/features/multi-agent.md
- docs/features/cost-optimization.md
- (etc)

**Files Modified**:
- docs/index.md

### Movement 3: Commands Pages
**Goal**: Organize commands into categorized pages

**Dependencies**: Movement 1

**Plan**:
- Read commands.md
- Group by category (chat, planning, research, ml, backend)
- Create docs/commands/ directory with categorized files
- Create navigation index

**Success Criteria**:
- docs/commands/ directory exists
- Commands grouped logically
- Navigation works

**Files Created**:
- docs/commands/chat.md
- docs/commands/planning.md
- docs/commands/research.md
- docs/commands/ml.md

### Movement 4: Guides
**Goal**: Create learning path guides

**Dependencies**: Movements 1, 2, 3

**Plan**:
- Create docs/guides/ if not exists
- Add getting-started.md
- Add advanced-usage.md
- Link from index.md

**Success Criteria**:
- Guides directory exists
- At least 2 guides created
- Guides reference features and commands properly

### Movement 5: Jekyll Navigation
**Goal**: Update Jekyll config for new structure

**Dependencies**: All previous movements

**Plan**:
- Read current _config.yml or similar
- Add navigation structure for new directories
- Update sidebar/menu
- Test locally if possible

**Success Criteria**:
- Jekyll config updated
- Navigation structure reflects new organization

## Implementation Plan

### Phase 1: Core Maestro (Week 1)
- [ ] Create Movement struct and types
- [ ] Implement Maestro decomposition
- [ ] Basic movement executor
- [ ] Save/load symphony state

### Phase 2: CLI Integration (Week 1-2)
- [ ] chu do --decompose flag (show movements without executing)
- [ ] Interactive movement execution
- [ ] Resume capability
- [ ] Progress tracking

### Phase 3: Validation (Week 2)
- [ ] Success criteria checking
- [ ] File-based validation
- [ ] Git status integration
- [ ] Rollback on failure

### Phase 4: Learning (Week 3+)
- [ ] Record symphony success/failure
- [ ] Learn optimal movement decomposition
- [ ] Improve validation criteria
- [ ] Auto-suggest improvements

## Testing Strategy

### Unit Tests
- Movement decomposition for various task types
- Movement dependency resolution
- Success criteria validation

### Integration Tests  
- Complete symphony execution (simple task)
- Resume from middle of symphony
- Handle failed movements
- Validate rollback

### Real-World Tests
1. Docs reorganization (this task)
2. Add new feature with tests
3. Refactor large module
4. Multi-file code review

## Success Metrics

- 90%+ task completion rate for complex tasks
- <5% unwanted file modifications
- User satisfaction with movement breakdown
- Successful resume rate >80%

## Rollout Plan

1. Implement core + simple test
2. Test with docs reorganization task
3. Gather feedback, iterate
4. Add advanced features (learning, auto-suggestions)
5. Document patterns for common task types
