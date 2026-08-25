# The Transcript Is Not the State

Canonical essay:
https://gptcode.dev/blog/2026-08-25-the-transcript-is-not-the-state

Evidence-carrying handoff protocol preview:
https://github.com/jadercorrea/ai-experiments/blob/bf56671f11de099b0ac1be61237601eedd7225ec/experiments/coding-agents/evidence-carrying-handoffs/2026-08-21/PROTOCOL_PREVIEW.md

## LinkedIn — English

A coding agent fails halfway through a repository task.

It inspected relevant files, rejected two plausible hypotheses, found a failing
test, produced an incomplete patch, and exhausted its budget.

What should the next agent inherit?

The usual answer is text: the full transcript, a generated summary, or a new
prompt explaining what happened.

But the transcript is not the state.

It mixes observations with guesses, evidence with explanation, failed actions
with conclusions, and source state with persuasive language generated before
the failure was understood.

I recently published a protocol preview for an experiment comparing three
recovery policies after the same coding-agent failure:

1. clean restart;
2. raw failed trajectory;
3. structured evidence-carrying handoff.

Designing the third condition required normalized events, semantic propositions,
rejected hypotheses, unresolved failures, evidence references, linguistic
realizations, and content hashes.

That started to look like more than a handoff format. It looked like an
intermediate representation of the agent's epistemic state.

The same representational problem exists in code generation. Coding agents read
and rewrite source text while repeatedly reconstructing symbols, types, effects,
dependencies, intent, and constraints. A character diff only indirectly
represents the semantic change the agent is trying to make.

This led me to a broader hypothesis:

Reliable coding agents need semantic intermediate representations for their
work—not only transcripts and source files.

Such a representation would not make agents perfectly correct. Integrity is
not truth. Well-typed is not intended. Verification against a contract does not
prove that the contract captures the product requirement.

The goal is narrower: stop forcing probabilistic models to reconstruct
structure that the system could preserve explicitly.

I wrote about how evidence-carrying handoffs, motion-event lexicalization,
semantic patches, and AI-native programming languages converge on this missing
layer.

The Transcript Is Not the State: Toward a Semantic IR for Coding Agents

https://gptcode.dev/blog/2026-08-25-the-transcript-is-not-the-state

#AIAgents #CodingAgents #SoftwareEngineering #ProgrammingLanguages #AIResearch

## LinkedIn — Portuguese

Um coding agent falha no meio de uma tarefa.

Ele inspecionou os arquivos relevantes, rejeitou duas hipóteses, encontrou um
teste falhando, produziu um patch incompleto e esgotou o orçamento.

O que o próximo agente deveria herdar?

A resposta habitual é texto: a trajetória completa, um resumo gerado ou um novo
prompt explicando o que aconteceu.

Mas o transcript não é o estado.

Ele mistura observações com hipóteses, evidências com explicações, ações que
falharam com conclusões e linguagem persuasiva produzida antes que a falha fosse
compreendida.

Recentemente publiquei o preview de um protocolo experimental que compara três
políticas de recuperação após a mesma falha de um coding agent:

1. reinício limpo;
2. trajetória bruta da tentativa anterior;
3. handoff estruturado carregando evidências.

Definir a terceira condição exigiu eventos normalizados, proposições semânticas,
hipóteses rejeitadas, falhas ainda não resolvidas, referências a evidências,
realizações linguísticas e hashes de conteúdo.

Isso começou a parecer mais que um formato de handoff. Parecia uma representação
intermediária do estado epistêmico do agente.

O mesmo problema aparece na geração de código. Agentes leem e reescrevem texto
enquanto reconstroem repetidamente símbolos, tipos, efeitos, dependências,
intenção e restrições. Um diff de caracteres representa apenas indiretamente a
mudança semântica pretendida.

Cheguei então a uma hipótese mais ampla:

Coding agents confiáveis precisam de representações intermediárias semânticas
para seu trabalho — não apenas transcripts e arquivos de código.

Isso não torna agentes perfeitamente corretos. Integridade não é verdade. Um
programa tipado não é necessariamente o programa pretendido. Verificar uma
implementação contra um contrato não prova que o contrato representa o requisito
do produto.

O objetivo é mais específico: deixar de obrigar modelos probabilísticos a
reconstruir estruturas que o sistema poderia preservar explicitamente.

Escrevi sobre como evidence-carrying handoffs, lexicalização de movimento,
patches semânticos e linguagens AI-native convergem para essa camada ausente.

The Transcript Is Not the State: Toward a Semantic IR for Coding Agents

https://gptcode.dev/blog/2026-08-25-the-transcript-is-not-the-state

#AIAgents #CodingAgents #EngenhariaDeSoftware #LinguagensDeProgramacao

## X thread

1/10

A coding agent fails halfway through a task after inspecting files, rejecting
hypotheses, finding a failing test, and producing an incomplete patch.

What should its successor inherit?

The transcript?

The transcript is not the state.

2/10

A transcript mixes objects with different authority:

- observations;
- guesses;
- commands;
- explanations;
- failed actions;
- verification results;
- claims of success.

The next model must reconstruct those distinctions probabilistically.

3/10

I published a protocol preview comparing three recovery policies after the same
coding-agent failure:

1. clean restart;
2. raw trajectory;
3. structured evidence-carrying handoff.

There are no study results yet. This is a design under review.

4/10

The structured handoff separates:

- normalized events;
- observations and rejected hypotheses;
- unresolved failures;
- semantic propositions;
- linguistic realizations;
- evidence references and hashes.

It behaves like an epistemic IR.

5/10

Why separate propositions from wording?

Motion-event lexicalization and multilingual LLM evaluations warn that language
is not always a neutral transport. Different realizations can package or
foreground different parts of an event.

6/10

The same representational problem exists in code.

Agents edit characters while repeatedly reconstructing symbols, types, effects,
dependencies, and intent.

A textual diff is only an indirect representation of a semantic change.

7/10

Imagine an agent proposing a typed operation instead:

“Change symbol X, allow db.read, preserve authorization, satisfy postcondition
Y, verify with tests Z.”

Then project that change into TypeScript, Rust, a human diff, or a visual view.

8/10

This would eliminate some failures, not correctness itself.

Integrity ≠ truth.

Well-typed ≠ intended.

Verified against a contract ≠ correct contract.

The goal is to preserve structure and expose the remaining uncertainty.

9/10

The falsifiable experiment is straightforward:

A: agents read and edit source files.

B: agents query semantic state and submit checked semantic patches.

Measure total tokens, hidden-test Pass@1, repair cycles, latency, and unsupported
tasks.

10/10

Human-readable text remains essential. It becomes a projection, not necessarily
the only canonical state.

The Transcript Is Not the State: Toward a Semantic IR for Coding Agents

https://gptcode.dev/blog/2026-08-25-the-transcript-is-not-the-state

## Bluesky / Mastodon

A transcript records words, not state. Coding agents use text to carry memory,
decisions, evidence, and claims of success—and source files as the primary
representation of program intent. I argue for a semantic IR for agent work:

https://gptcode.dev/blog/2026-08-25-the-transcript-is-not-the-state

## Hacker News

Suggested title:

The Transcript Is Not the State: Toward a Semantic IR for Coding Agents

Submission URL:

https://gptcode.dev/blog/2026-08-25-the-transcript-is-not-the-state

First comment:

Author here. This essay grew out of a protocol preview for recovery after a
coding-agent failure. The proposed experiment compares a clean restart, the raw
failed trajectory, and a structured handoff whose claims reference normalized
repository and execution evidence.

There are no outcomes yet—the preview records zero study-model calls and makes
no efficacy claim. The interesting design consequence was that the handoff
started to resemble an intermediate representation of the agent's epistemic
state: events, propositions, rejected hypotheses, unresolved failures,
linguistic realizations, evidence references, and hashes.

The essay asks whether the same architecture should extend to code changes.
Instead of repeatedly rewriting source text, a coding agent could propose typed,
atomic semantic changes with explicit effects, preserved invariants, and
verification obligations, then project them into target code and human-readable
diffs.

I am especially interested in counterexamples and adjacent systems. AIRL,
Edict, NURL, MLIR, AI-oriented grammar research, program synthesis, and proof-
carrying code cover parts of the space. The open question is whether a semantic
editing interface improves real repository maintenance when compared under the
same model and executable evaluation.

Protocol preview:
https://github.com/jadercorrea/ai-experiments/blob/bf56671f11de099b0ac1be61237601eedd7225ec/experiments/coding-agents/evidence-carrying-handoffs/2026-08-21/PROTOCOL_PREVIEW.md

## Reddit — r/ProgrammingLanguages or r/LocalLLaMA

Suggested title:

Toward a semantic intermediate representation for coding agents

Post:

I have been working on a protocol for recovery after coding-agent failure. One
of its treatments replaces the raw failed trajectory with a structured handoff:
normalized events, propositions, rejected hypotheses, unresolved failures,
linguistic realizations, evidence references, and content hashes.

While designing it, I started seeing the handoff as an epistemic IR rather than
just a better summary. That led to a broader question: should coding agents also
operate on an action/program IR instead of treating source text as their primary
authoring surface?

The proposed architecture is:

```text
evidence IR → epistemic IR → intent/change IR → target projections
```

The claim is not that structure solves semantic correctness. Integrity is not
truth, types do not establish intent, and proofs remain relative to their
specifications. The narrower hypothesis is that agents should not repeatedly
reconstruct symbols, effects, dependencies, and evidence when the system could
preserve them explicitly.

The essay includes a bounded experiment: compare ordinary file editing with
semantic queries and checked patches under the same model, measuring total
tokens, hidden-test Pass@1, repair cycles, latency, and unsupported-task rate.

I would value criticism from people working on compilers, program synthesis,
formal methods, agent memory, or code-model evaluation:

https://gptcode.dev/blog/2026-08-25-the-transcript-is-not-the-state

## GitHub Discussion

Title:

Should coding agents operate on semantic state instead of transcripts and files?

Body:

I published an engineering essay connecting three problems:

1. recovery after coding-agent failure;
2. language-dependent realization of agent memory;
3. semantic program and change representations for coding agents.

The working thesis is that reliable agents need a versioned semantic substrate
for observations, beliefs, changes, effects, and verification evidence. Text and
source code remain essential human projections, but need not be the only
canonical state.

The essay proposes a narrow A/B experiment rather than a general-purpose
language implementation. I would particularly value feedback on:

- the smallest useful semantic core;
- existing systems or literature the essay misses;
- how to measure unsupported-task rate without rewarding a deliberately narrow
  language;
- whether bidirectional source projection is necessary for a credible first
  experiment;
- which repository tasks would expose the strongest counterexamples.

Essay:
https://gptcode.dev/blog/2026-08-25-the-transcript-is-not-the-state

Protocol preview:
https://github.com/jadercorrea/ai-experiments/blob/bf56671f11de099b0ac1be61237601eedd7225ec/experiments/coding-agents/evidence-carrying-handoffs/2026-08-21/PROTOCOL_PREVIEW.md
