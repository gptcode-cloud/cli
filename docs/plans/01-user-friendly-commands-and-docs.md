# Objetivo
Atualizar CLI para comandos user-friendly e sincronizar toda documentação com o estado real do projeto, seguindo as diretrizes dos notebooks compartilhados.

# Contexto Atual
- ✅ Já implementamos `gptcode profile` (singular) com subcomandos friendly
- ⚠️ Ainda temos comandos verbosos como `gptcode config get defaults.backend`
- ⚠️ Documentação menciona comandos inexistentes ou desatualizados
- ⚠️ Foco em TDD quando deveria destacar agents + validation

# Mudanças no CLI

## 1. Backend Commands (Singular)
Adicionar `gptcode backend` (singular) com subcomandos friendly:

### Comandos Novos
```bash
gptcode backend                    # show current backend
gptcode backend list               # list all backends
gptcode backend show [name]        # show backend config
gptcode backend use <name>         # switch to backend
gptcode backend create <name> <type> <url>  # já existe
gptcode backend delete <name>      # já existe
```

### Implementação
- Criar `backendShowCmd` para mostrar backend atual
- Criar `backendUseCmd` para trocar backend (atualiza defaults.backend)
- Modificar `backendCmd` para ter RunE que mostra backend atual
- Arquivos: `cmd/gptcode/main.go`

## 2. Model Commands (já existe, verificar completude)
Verificar se `gptcode model` já tem todos os comandos friendly:
```bash
gptcode model list
gptcode model recommend
gptcode model install
gptcode model update
```

## 3. Atualizar Help Principal
Atualizar `rootCmd.Long` para:
- Remover menção a `gptcode config get/set` do help principal
- Adicionar `gptcode profile` e `gptcode backend` nos exemplos
- Manter categorização já feita

# Atualização de Documentação

## 4. README.md
Atualizar seções:

### Hero Section
- Manter mensagem atual sobre agents
- Adicionar diagrama de orchestração (se não existir)

### Backend Management (linha ~262)
Substituir:
```bash
gptcode config set defaults.backend mygroq
```
Por:
```bash
gptcode backend use mygroq
```

### Profile Management (linha ~280)
Adicionar exemplos com `gptcode profile` (singular):
```bash
gptcode profile                  # show current
gptcode profile list            # list all
gptcode profile use groq.speed  # switch
```

## 5. docs/commands.md
Atualizar seção de configuração:

### Backend Management
```bash
gptcode backend                 # Show current backend
gptcode backend list           # List all backends  
gptcode backend use groq       # Switch backend
gptcode backend create <name> <type> <url>
gptcode backend delete <name>
```

### Profile Management
```bash
gptcode profile                      # Show current profile
gptcode profile list [backend]       # List all profiles
gptcode profile show [backend.profile]
gptcode profile use backend.profile  # Switch profile
```

### Configuration (Advanced)
Mover `gptcode config get/set` para seção "Advanced" e marcar como "para uso avançado".

## 6. docs/index.md (Homepage)
Baseado no notebook "Repensando o Site":

### Hero Section
Verificar se já tem:
- Mensagem: "AI Coding Assistant with Specialized Agents"
- Subtítulo sobre Analyzer → Planner → Editor → Validator
- Menção a $0-5/month

### Features Cards
Ordem proposta:
1. 🚀 Agent Orchestration
2. ✅ File Validation & Success Criteria
3. 🧠 Intelligent Context (dependency graph)
4. 💰 Radically Affordable
5. 🔧 Supervised vs Autonomous
6. 🎯 Deep Neovim Integration

### Diagrama
Adicionar diagrama mermaid de agent orchestration (conforme notebook).

## 7. docs/features.md
Reestruturar seções:

### Nova ordem
1. Agent-Based Architecture
   - Analyzer, Planner, Editor, Validator
2. Validation & Safety
   - File validation, success criteria, over-engineering protection
3. Intelligence Features
   - ML routing, dependency graph, context optimization
4. Cost Optimization
   - Mix models per agent, profile management
5. Developer Experience
   - Neovim integration, workflow
6. TDD Features (mover para final)

## 8. Posts do Blog
Baseado no notebook "Revisão completa dos posts":

### Alta Prioridade (correções críticas)
- `2025-11-17-ollama-local-setup.md`
  - Remover seção de "hybrid setup" com múltiplos backends
  - Clarificar que só um backend por vez
  
- `2025-11-21-profile-management.md`
  - Atualizar exemplos para usar `gptcode profile` (singular)
  - Adicionar `gptcode profile use`

### Média Prioridade
- `2025-11-15-groq-optimal-configs.md`
  - Atualizar comandos de config
  
- `2025-11-16-openrouter-multi-provider.md`
  - Atualizar comandos de config

### Padrão de Busca e Substituição
Em todos os posts:
- `gptcode config set defaults.backend X` → `gptcode backend use X`
- `gptcode config set defaults.profile Y` → `gptcode profile use X.Y`
- `gptcode profiles list X` → `gptcode profile list X`
- `gptcode profiles show X Y` → `gptcode profile show X.Y`

## 9. Guias
- `docs/guides/getting-started.md`
  - Atualizar comandos de configuração
  - Usar `gptcode backend` e `gptcode profile` nos exemplos

# Ordem de Implementação

## Fase 1: CLI Core (Alta Prioridade)
1. ✅ Implementar `gptcode backend` (singular) commands
2. ✅ Atualizar help principal (remover config get/set)
3. ✅ Testar todos os comandos novos
4. ✅ Build e install

## Fase 2: Docs Core (Alta Prioridade)
5. ✅ Atualizar README.md (comandos)
6. ✅ Atualizar docs/commands.md
7. ✅ Atualizar docs/index.md (hero + features)

## Fase 3: Docs Features (Média Prioridade)
8. ✅ Reestruturar docs/features.md
9. ✅ Atualizar docs/guides/getting-started.md

## Fase 4: Blog Posts (Média Prioridade)
10. ✅ Atualizar profile-management.md
11. ✅ Atualizar ollama-local-setup.md
12. ✅ Buscar e substituir em outros posts

## Fase 5: Validação (Baixa Prioridade)
13. ⏳ Grep por `gptcode config set defaults` em todos os arquivos
14. ⏳ Verificar se algum doc menciona comandos inexistentes
15. ⏳ Criar checklist de comandos vs docs

# Success Criteria

## CLI
- ✅ `gptcode backend` mostra backend atual
- ✅ `gptcode backend use <name>` troca backend
- ✅ `gptcode profile use <backend>.<profile>` troca ambos
- ✅ Help principal não menciona config get/set
- ✅ Todos os comandos friendly funcionam

## Documentação
- ✅ README usa comandos friendly
- ✅ docs/commands.md atualizado
- ✅ docs/index.md destaca agents
- ✅ docs/features.md reorganizado
- ✅ Posts do blog atualizados
- ✅ Nenhuma menção a comandos inexistentes
- ✅ Mensagem consistente: "Agents + Validation"

## Consistência
- ✅ CLI help alinhado com docs
- ✅ Todos os exemplos usam comandos friendly
- ✅ TDD mencionado mas não dominante
- ✅ Foco em agents, orchestration, validation
