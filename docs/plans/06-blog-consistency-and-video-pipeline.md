# Revisão dos Posts e Pipeline Blog→Vídeo (Zapfy-AI)

Status: Fase 1 concluída em 2025-12-05  
Escopo: Padronizar posts recentes do blog e criar um pipeline para gerar vídeos em PT-BR automaticamente a partir desses posts, publicando no canal Zapfy-AI.  
Dependências: Nenhuma crítica (pipeline pode começar com posts já padronizados).

## Contexto
Os posts anteriores a 2025-12-01 já seguiam uma estrutura consistente (frontmatter com `description` e `tags`, tom narrativo pessoal, "References", "See Also" e CTA para GitHub Discussions). Os posts a partir de 2025-12-01 até 2025-12-05 divergiam desse padrão.

## Objetivos
- Garantir consistência editorial e técnica em todos os posts.
- Transformar cada post em um vídeo curto em português com narrativa clara, demonstração e CTA.
- Viabilizar publicação contínua no canal Zapfy-AI.
- Preparar terreno para posicionamento de GPTCode dentro da org Zapfy.

## Fases
### Fase 1 — Padronização dos posts (Concluída)
Posts ajustados: 2025-12-01 → 2025-12-05.
- Frontmatter: `author: Jader Correa`, `tags` (não `categories`), `description` adicionada.
- Estrutura: seções "References", "See Also" e CTA para Discussions.
- Tom: introduções com narrativa conectando posts anteriores.

### Fase 2 — Roteiro PT-BR por post
Para cada post, gerar um roteiro seguindo este esqueleto:
1. Abertura (hook em 8–12s) — dor do público SMB/Dev.
2. Contexto rápido — por que isso importa agora.
3. Demonstração/explicação com 1–2 exemplos concretos.
4. Resultado/impacto (métricas/tokens/tempo quando fizer sentido).
5. CTA: "Assine o Zapfy-AI" + link para repo/discussion.
Saída: `docs/scripts/YYYY-MM-DD-<slug>.pt-BR.md`.

### Fase 3 — Narração/TTS e assets
- TTS: modelo PT-BR (voz neutra) com fallback local.
- Legendas: SRT gerado do texto (timestamps aproximados por parágrafo).
- Assets: capturas/diagramas minimalistas a partir de trechos do post (ex.: PageRank, pipeline de agentes).
Saídas:
- `assets/video/YYYY-MM-DD-<slug>/voiceover.wav`
- `assets/video/YYYY-MM-DD-<slug>/captions.srt`
- `assets/video/YYYY-MM-DD-<slug>/stills/*.png`

### Fase 4 — Render e publicação
- Render: template de 1080p com faixas (narrador, música leve, gráficos).
- Automação: script para juntar voz + imagens + legendas.
- Publicação: título/descrição/hashtags em PT-BR derivados do post.
- Checklist de QA: duração 3–5 min, áudio limpo, sem textos pequenos.

### Fase 5 — Migração para org Zapfy
- Transferência do repositório para `zapfy/gptcode`.
- Atualização de `go.mod`, README, badges e links do site.
- Post/Discussion explicando a mudança (OS continua MIT, mantido pela Zapfy).

## Prioridade
- Este plano entra como prioridade 06 (após 01–05 já existentes). Não bloqueia roadmap técnico e habilita aquisição/marketing.

## Critérios de Sucesso
- 100% dos posts recentes padronizados (12-01 a 12-05) — OK.
- 3 primeiros vídeos publicados (Why GPTCode, Context Engineering, Why GPTCode Isn’t Trying...).
- Tempo total por vídeo ≤ 2h (roteiro→render→publicação) após setup.
- +10% de engajamento médio nas páginas de docs relacionadas (baseline a ser medido).

## Riscos e Mitigações
- Risco: Vídeos muito técnicos para SMB → Mitigar com roteiros mais curtos, foco em valor prático.
- Risco: TTS artificial → Avaliar 2 vozes PT-BR e ajustar entonação/velocidade.
- Risco: Migração de repo quebra links → Redirecionos e varredura de links após migração.

## Próximas Ações
1. Gerar roteiros PT-BR para 3 vídeos iniciais.
2. Testar pipeline TTS + render com um post (piloto).
3. Preparar comunicado de migração do repositório.
