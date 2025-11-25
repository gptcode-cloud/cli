# Plano: Maestro - Autonomous End-to-End Execution

## Visão Geral

**Objetivo:** Chuchu resolve tarefas completamente de forma autônoma, do research até código funcionando e testado.

**Fases:**
- **Fase 1:** Autonomous local execution (CLI + Neovim)
- **Fase 2:** GitHub Actions integration + public offering

---

## Fase 1: Autonomous Local Execution

**Meta (MVP):** `chu auto plan.md` executa implement → verify com retries. Integração research → plan virá ainda na Fase 1.

### 1.1 Orchestrator (Maestro Core)

**Arquivo:** `internal/maestro/orchestrator.go`

**Responsabilidade:** Loop principal de execução com retry e error recovery.

```go
type Maestro struct {
    research  *agents.ResearchAgent
    planner   *agents.PlanAgent
    editor    *agents.EditorAgent
    verifier  *Verifier
    maxRetries int
}

func (m *Maestro) AutoResolve(ctx context.Context, task string) error {
    // 1. Research
    research := m.research.Execute(ctx, task)
    
    // 2. Plan
    plan := m.planner.CreatePlan(ctx, research)
    
    // 3. Implement (with retry)
    for attempt := 1; attempt <= m.maxRetries; attempt++ {
        err := m.implementPlan(ctx, plan)
        if err == nil {
            break
        }
        
        // Error recovery
        plan = m.adjustPlan(ctx, plan, err)
    }
    
    // 4. Verify
    return m.verifier.VerifyChanges(ctx)
}
```

**Tarefas:**
- [ ] Criar estrutura base do Maestro
- [ ] Implementar loop de retry com backoff
- [ ] Adicionar logging/status updates
- [ ] Integrar com agents existentes

**Tempo estimado:** 3 dias

---

### 1.2 Verification System

**Arquivo:** `internal/maestro/verifier.go`

**Responsabilidade:** Validar que mudanças compilam, passam em testes e seguem padrões.

```go
type Verifier struct {
    projectRoot string
    language    string
}

type VerificationResult struct {
    BuildSuccess  bool
    TestsPass     bool
    LintClean     bool
    Errors        []string
}

func (v *Verifier) VerifyChanges(ctx context.Context) error {
    result := VerificationResult{}
    
    // 1. Detect test command
    testCmd := v.detectTestCommand()
    
    // 2. Run tests
    result.TestsPass, result.Errors = v.runTests(testCmd)
    
    // 3. Run build/compile check
    result.BuildSuccess = v.verifyBuild()
    
    // 4. Run lint (optional)
    result.LintClean = v.runLint()
    
    if !result.TestsPass || !result.BuildSuccess {
        return fmt.Errorf("verification failed: %v", result.Errors)
    }
    
    return nil
}
```

**Detecção automática de comandos:**
- Go: `go test ./...`, `go build`
- Node: `npm test`, `npm run build`
- Python: `pytest`, `python -m py_compile`
- Elixir: `mix test`, `mix compile`
- Ruby: `bundle exec rspec`, `ruby -c`

**Tarefas:**
- [ ] Implementar auto-detecção de test framework
- [ ] Parser de output de testes (JUnit XML, TAP, etc)
- [ ] Implementar verificação de build
- [ ] Adicionar lint opcional
- [ ] Estruturar erros para feedback ao LLM

**Tempo estimado:** 4 dias

---

### 1.3 Error Recovery Strategy

**Arquivo:** `internal/maestro/recovery.go`

**Responsabilidade:** Analisar erros e ajustar plano de execução.

```go
type RecoveryStrategy struct {
    editor *agents.EditorAgent
}

func (r *RecoveryStrategy) AdjustPlan(
    ctx context.Context,
    plan string,
    verificationErr error,
) (string, error) {
    // 1. Parse error details
    errorAnalysis := r.analyzeError(verificationErr)
    
    // 2. Generate fix prompt
    fixPrompt := fmt.Sprintf(`
Previous implementation failed with error:
%s

Files affected: %v
Error type: %s

Please adjust the implementation to fix this error.
`, errorAnalysis.Message, errorAnalysis.Files, errorAnalysis.Type)
    
    // 3. Ask editor agent to fix
    return r.editor.Fix(ctx, plan, fixPrompt)
}
```

**Tipos de erro a detectar:**
- Syntax error → linha/arquivo específico
- Type error → tipo esperado vs recebido
- Test failure → teste específico, assertion
- Import error → dependência faltando
- Runtime error → stacktrace

**Tarefas:**
- [ ] Parser de erros por linguagem
- [ ] Classificador de tipo de erro
- [ ] Template de prompts de correção
- [ ] Limite de retries por tipo de erro

**Tempo estimado:** 3 dias

---

### 1.4 Checkpoint System

**Arquivo:** `internal/maestro/checkpoint.go`

**Responsabilidade:** Salvar estado entre fases para permitir resume.

```go
type Checkpoint struct {
    TaskID      string
    Phase       string // "research", "plan", "implement", "verify"
    Data        map[string]interface{}
    Timestamp   time.Time
    FilesChanged []string
}

func (c *Checkpoint) Save() error {
    path := filepath.Join(os.UserHomeDir(), ".chuchu", "checkpoints", c.TaskID+".json")
    return writeJSON(path, c)
}

func LoadCheckpoint(taskID string) (*Checkpoint, error) {
    path := filepath.Join(os.UserHomeDir(), ".chuchu", "checkpoints", taskID+".json")
    return readJSON(path)
}
```

**Tarefas:**
- [ ] Estrutura de checkpoint
- [ ] Save/load de estado
- [ ] Comando `chu resume <task-id>`
- [ ] Limpeza de checkpoints antigos

**Tempo estimado:** 2 dias

---

### 1.5 CLI Command: `chu auto`

**Arquivo:** `cmd/chu/auto.go`

**Uso (MVP Fase 1):**
```bash
chu auto docs/plans/your-plan.md
chu auto docs/plans/your-plan.md --max-retries 5
chu auto docs/plans/your-plan.md --resume
```

Notas:
- MVP executa implement → verify com retries a partir de um arquivo de plano.
- Integração research → plan será adicionada ao `auto` em uma iteração posterior da Fase 1.

**Flags:**
- `--resume <task-id>`: Continue de checkpoint
- `--max-retries <n>`: Máximo de tentativas (default: 3)
- `--verify-only`: Só roda verificação
- `--dry-run`: Mostra o que faria sem executar

**Tarefas:**
- [ ] Criar comando CLI
- [ ] Integrar com Maestro
- [ ] Status updates via stderr
- [ ] Progress bar opcional

**Tempo estimado:** 2 dias

---

### 1.6 Neovim Integration

**Arquivo:** `neovim/lua/chuchu/init.lua`

**Comando:** `:ChuchuAuto`

Status: comando básico implementado para executar `chu auto <plan_file>` via jobstart, com janela de status. Integrações de progress detalhado virão em iterações.

**Comportamento:**
- Abre prompt para tarefa
- Executa em background (jobstart)
- Mostra progress em janela lateral
- Notifica quando completa ou falha
- Permite abort (Ctrl+C)

**Status display:**
```
🐺 Chuchu Auto
─────────────────
Phase: Implement
Retry: 1/3
⚙ Running tests...
✓ Research complete
✓ Plan created
⚙ Implementing...
```

**Tarefas:**
- [ ] Comando `:ChuchuAuto`
- [ ] Progress window
- [ ] Job control (start/stop)
- [ ] Integrar com tool events

**Tempo estimado:** 3 dias

---

### 1.7 Rollback System

**Arquivo:** `internal/maestro/checkpoint.go` (restore) + `internal/maestro/orchestrator.go` (git diff tracking)

**Responsabilidade:** Desfazer mudanças se verificação falhar após max retries.

```go
func (m *Maestro) Rollback(checkpoint *Checkpoint) error {
    for _, file := range checkpoint.FilesChanged {
        // Restore from git or backup
        exec.Command("git", "checkout", "HEAD", file).Run()
    }
    return nil
}
```

**Estratégias:**
- Git stash de mudanças antes de cada fase
- Backup de arquivos modificados
- Opção de commit incremental (commit por fase bem-sucedida)

**Tarefas:**
- [ ] Git stash integration
- [ ] Backup manual (se não for repo git)
- [ ] Comando `chu rollback <task-id>`

**Tempo estimado:** 2 dias

---

### Resumo Fase 1

**Duração total:** 3-4 semanas

**Deliverables:**
- ✅ `chu auto plan.md` executa implement → verify com retry
- ✅ Verificação automática (build + test) com auto-detecção de linguagem
- ✅ Checkpoint/resume (básico)
- ✅ Neovim `:ChuchuAuto` (básico)
- ✅ Rollback via checkpoints + git diff tracking

**Validação (próximos passos):**
- [ ] 10 tasks reais executadas com sucesso
- [ ] Taxa de sucesso > 70% no primeiro try
- [ ] Taxa de sucesso > 90% com retries
- [ ] Tempo médio < 5 minutos por task

**Status Atual (MVP):**
- ✅ Comando `chu auto` funcional
- ✅ Verificação automática (build + test) com auto-detecção de linguagem
- ✅ Retry com error recovery e prompts específicos
- ✅ Checkpoint/resume básico
- ✅ Rollback via git diff tracking
- ✅ Lint verification opcional (--lint)
- ✅ Parser de planos com suporte a sub-steps (##, ###)
- ✅ Testes unitários e E2E
- ✅ Neovim :ChuchuAuto
- ✅ Documentação atualizada

---

## Fase 2: GitHub Actions Integration

**Meta:** `chu issue-to-pr #123` cria PR automaticamente a partir de issue.

### 2.1 GitHub Client

**Arquivo:** `internal/github/client.go`

**Responsabilidade:** Integração com GitHub API.

```go
type GitHubClient struct {
    token string
    repo  string
}

func (c *GitHubClient) GetIssue(number int) (*Issue, error)
func (c *GitHubClient) CreateBranch(name string) error
func (c *GitHubClient) CreatePR(opts PROptions) (*PullRequest, error)
func (c *GitHubClient) CommentOnPR(number int, comment string) error
func (c *GitHubClient) LinkPRToIssue(pr, issue int) error
```

**Tarefas:**
- [ ] Wrapper da GitHub API
- [ ] Autenticação (token)
- [ ] Issue parser
- [ ] PR creator
- [ ] Comment handler

**Tempo estimado:** 3 dias

---

### 2.2 Git Operations

**Arquivo:** `internal/git/operations.go`

**Responsabilidade:** Operações git locais.

```go
func CreateBranchFromIssue(issueNumber int) (string, error)
func CommitChanges(message string, files []string) error
func PushBranch(branch string) error
func GetChangedFiles() ([]string, error)
```

**Tarefas:**
- [ ] Git wrapper
- [ ] Branch naming convention
- [ ] Commit message template
- [ ] Push com retry

**Tempo estimado:** 2 dias

---

### 2.3 PR Description Generator

**Arquivo:** `internal/github/pr_description.go`

**Responsabilidade:** Gerar descrição de PR baseada em mudanças.

```go
func GeneratePRDescription(opts PRDescriptionOptions) (string, error) {
    return fmt.Sprintf(`
## Description
Resolves #%d

%s

## Changes
%s

## Testing
%s

## Checklist
- [x] Tests added/updated
- [x] Documentation updated
- [x] Code follows style guide
`, opts.IssueNumber, opts.Summary, opts.ChangesSummary, opts.TestEvidence)
}
```

**Tarefas:**
- [ ] Template de PR description
- [ ] Summarizer de mudanças
- [ ] Link para issue
- [ ] Checklist automático

**Tempo estimado:** 2 dias

---

### 2.4 CLI Command: `chu issue-to-pr`

**Arquivo:** `cmd/issue_to_pr.go`

**Uso:**
```bash
chu issue-to-pr #123
chu issue-to-pr #123 --draft
chu issue-to-pr #123 --no-push  # local only
```

**Fluxo:**
1. Fetch issue from GitHub
2. Extract requirements
3. Run `auto` mode
4. Create branch
5. Commit changes
6. Push branch
7. Create PR (draft ou normal)
8. Link PR to issue

**Tarefas:**
- [ ] Comando CLI
- [ ] Issue parser
- [ ] Branch/commit/push flow
- [ ] PR creation
- [ ] Error handling

**Tempo estimado:** 3 dias

---

### 2.5 GitHub Action Workflow

**Arquivo:** `.github/workflows/chuchu-auto.yml`

```yaml
name: Chuchu Auto-Resolve

on:
  issues:
    types: [labeled]

jobs:
  auto-resolve:
    if: contains(github.event.issue.labels.*.name, 'chuchu-auto')
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Chuchu
        run: |
          curl -L https://github.com/jadercorrea/chuchu/releases/latest/download/chu-linux-amd64 -o /usr/local/bin/chu
          chmod +x /usr/local/bin/chu
          chu setup --non-interactive
      
      - name: Configure API Keys
        env:
          GROQ_API_KEY: ${{ secrets.GROQ_API_KEY }}
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          chu key groq --key $GROQ_API_KEY
      
      - name: Auto-resolve issue
        run: |
          chu issue-to-pr ${{ github.event.issue.number }} --draft
```

**Tarefas:**
- [ ] Criar workflow template
- [ ] Documentação de setup
- [ ] Exemplo de secrets necessários

**Tempo estimado:** 1 dia

---

### 2.6 Docker Image (Public)

**Arquivo:** `Dockerfile`

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o chu cmd/main.go

FROM alpine:latest
RUN apk add --no-cache git
COPY --from=builder /app/chu /usr/local/bin/chu
ENTRYPOINT ["chu"]
```

**Publish:**
- Docker Hub: `jadercorrea/chuchu:latest`
- GitHub Container Registry: `ghcr.io/jadercorrea/chuchu:latest`

**Tarefas:**
- [ ] Dockerfile otimizado
- [ ] Multi-arch (amd64, arm64)
- [ ] CI para publish automático
- [ ] Documentação de uso

**Tempo estimado:** 2 dias

---

### 2.7 CircleCI Orb (opcional)

**Arquivo:** `orb/chuchu.yml`

```yaml
version: 2.1
orbs:
  chuchu: jadercorrea/chuchu@1.0.0

workflows:
  auto-resolve-issues:
    jobs:
      - chuchu/auto-resolve:
          issue-number: $CIRCLE_PR_NUMBER
          api-key: $GROQ_API_KEY
```

**Tarefas:**
- [ ] Criar orb structure
- [ ] Publish no CircleCI registry
- [ ] Documentação

**Tempo estimado:** 2 dias (se decidir fazer)

---

### 2.8 Documentation Site Update

**Posts a criar:**
- "Autonomous Execution with Chuchu" (como usar `chu auto`)
- "GitHub Actions Integration" (setup workflow)
- "Self-Service AI: From Issue to PR" (fluxo completo)

**Atualizar:**
- README com badges de CI
- Docs de comandos (adicionar `auto`, `issue-to-pr`)
- Troubleshooting guide

**Tarefas:**
- [ ] 3 novos posts no blog
- [ ] Atualizar README
- [ ] Video demo (opcional)

**Tempo estimado:** 3 dias

---

### Resumo Fase 2

**Duração total:** 2-3 semanas

**Deliverables:**
- ✅ `chu issue-to-pr` cria PRs automaticamente
- ✅ GitHub Action workflow funcional
- ✅ Docker image pública
- ✅ Documentação completa
- ✅ (Opcional) CircleCI orb

**Validação:**
- [ ] GitHub Action roda com sucesso em 3+ repos
- [ ] PR criado automaticamente a partir de issue
- [ ] Docker image pull < 100MB
- [ ] Documentação clara para setup

---

## Cronograma Total

**Fase 1:** 3-4 semanas
**Fase 2:** 2-3 semanas

**Total:** 5-7 semanas (1.5-2 meses)

---

## Riscos e Mitigações

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| Loop infinito de retries | Alto | Max retries hard limit (3) |
| Breaking changes em prod | Crítico | Sempre `--draft` por padrão, require approval |
| Custo de API explode | Alto | Budget limit no config, abort se > threshold |
| Verificação falha silenciosamente | Médio | Logging robusto, notify on failure |
| GitHub rate limits | Médio | Backoff exponencial, cache de dados |

---

## Métricas de Sucesso

**Fase 1:**
- Taxa de sucesso > 70% (first try)
- Taxa de sucesso > 90% (com retries)
- Tempo médio < 5 min por task

**Fase 2:**
- 10+ repos usando GitHub Action
- 50+ PRs gerados com sucesso
- < 10% de PRs rejeitados por qualidade

---

## Notas de Implementação

### Dependências Existentes

**Já implementado:**
- ✅ Agentes especializados (router, query, editor, research)
- ✅ Perfis de configuração por backend
- ✅ Research mode com semantic search + web search
- ✅ Plan mode
- ✅ Implement mode (básico)
- ✅ Tool events system (neovim integration)

**Faltando (crítico para Maestro):**
- ❌ Orchestration loop
- ❌ Error recovery
- ❌ Verification system
- ❌ Checkpoint/resume
- ❌ GitHub integration

### Considerações Técnicas

**Context Management:**
- Maestro deve gerenciar context budget
- Summarize fases anteriores se context > 80%
- Prune mensagens antigas mantendo só essencial

**Error Classification:**
- Syntax errors: retry com correção específica
- Logic errors: retry com análise de testes
- Dependency errors: sugerir install antes de retry
- Timeout errors: não contar como tentativa

**Testing Strategy:**
- Unit tests: Verifier, RecoveryStrategy, Checkpoint
- Integration tests: Maestro end-to-end em repo mock
- Manual validation: 10+ issues reais antes de release

**Performance:**
- Research phase: 30-60s
- Plan phase: 20-40s
- Implement phase: 1-2 min por iteração
- Verify phase: 10-30s (depende de test suite)
- **Total esperado:** 3-5 min para task simples

### Priorização Interna

**Must Have (Fase 1 MVP):**
1. Maestro core loop
2. Verification system (build + test)
3. Error recovery básico
4. CLI command `chu auto`

**Should Have (Fase 1 complete):**
5. Checkpoint/resume
6. Rollback system
7. Neovim integration

**Could Have (Fase 2):**
8. GitHub integration
9. Docker image
10. CircleCI orb

---

## Próximos Passos

1. **Validar arquitetura** com equipe/comunidade
2. **Criar issues** no GitHub para cada tarefa
3. **Setup milestone** "Maestro Phase 1" e "Phase 2"
4. **Começar implementação** por Maestro core (1.1)
5. **Iterar** com feedback de early adopters
