---
layout: post
title: "The Transcript Is Not the State: Toward a Semantic IR for Coding Agents"
date: 2026-08-25
author: Jader Correa
series: Evidence-Based AI Engineering
format: Engineering essay
tags: [ai-agents, coding-agents, intermediate-representation, verification, agent-memory, programming-languages]
description: "Evidence-carrying handoffs, linguistic realization, and AI-native code point toward the same missing layer: a semantic intermediate representation for agent work."
---

<aside class="paper-abstract" aria-label="Executive summary">
  <strong>In brief: a transcript records words, not state.</strong>
  Coding agents currently use free-form text to carry observations, memory,
  decisions, instructions, and claims of success. They use source files in a
  similarly overloaded way: as implementation, interface, explanation, and
  editing surface. Evidence-carrying handoffs and AI-native program
  representations suggest a different architecture—stable semantic objects
  with explicit evidence, effects, and constraints, projected into language or
  target code only when needed. This is a design hypothesis, not an efficacy
  claim.
</aside>

A coding agent fails halfway through a task. It has inspected the repository,
rejected two plausible hypotheses, found a failing test, produced an incomplete
patch, and exhausted its budget.

What should the next agent inherit?

The usual answer is some amount of text: the complete transcript, a generated
summary, or a new prompt explaining what happened. But a transcript is not the
state of the work. It is one linguistic realization of a trajectory. It mixes
observations with guesses, evidence with explanation, and failed actions with
conclusions.

The same problem appears when an agent writes software. A source file is a
serialization of a program, but coding agents treat it as if it were the
program's only meaningful representation. They repeatedly read and rewrite
text in order to recover types, effects, dependencies, intent, and constraints
that compilers often already know in structured form.

This suggests a broader working thesis:

> **Reliable coding agents need semantic intermediate representations for
> their work, not only transcripts and source files.**

I arrived at this thesis while designing a public protocol preview for
[evidence-carrying recovery after coding-agent failure](https://github.com/jadercorrea/ai-experiments/blob/bf56671f11de099b0ac1be61237601eedd7225ec/experiments/coding-agents/evidence-carrying-handoffs/2026-08-21/PROTOCOL_PREVIEW.md).
The protocol is not a completed experiment. It records zero study-model calls,
authorizes no calibration or confirmatory execution, and supports no claim that
structured handoffs improve recovery. What it does provide is a concrete
representation of the problem.

That representation now looks less like a special-purpose handoff format and
more like the beginning of a missing systems layer.

## A transcript is doing too many jobs

Agent transcripts are convenient because language models naturally consume and
produce text. Convenience has gradually turned the transcript into a universal
container for:

- repository observations;
- hypotheses and rejected hypotheses;
- plans and decisions;
- commands and their outputs;
- patches and verification results;
- memory transferred between attempts;
- explanations presented to humans;
- claims that a task is complete.

These objects do not have the same semantics or authority.

An observed command exit code is different from an agent's interpretation of
that command. A hypothesis is different from a fact. A patch is different from
evidence that the patch resolves the issue. A concise summary is different from
the failed trajectory from which it was derived.

When all of them are flattened into text, the receiver must reconstruct those
distinctions probabilistically. It must infer which statements were observed,
which were proposed, which remain uncertain, and which refer to repository
state that may no longer exist.

This is one reason raw trajectories are not obviously the best recovery
context. They preserve detail, but they also preserve noise, abandoned paths,
incorrect assumptions, repeated commands, and persuasive language generated
before the failure was understood. A summary reduces volume, but it can remove
the evidence needed to audit its claims.

The problem is not simply context length. It is missing structure.

## What the handoff protocol forced into the open

The proposed recovery study compares three context policies after the same
qualifying sender failure:

1. **Clean restart:** the original task and immutable repository state.
2. **Raw trajectory:** the clean inputs plus the allowed failed transcript.
3. **Structured evidence:** the clean inputs plus an evidence-linked handoff.

Every receiver begins from the same source commit. Sender patches, untracked
files, caches, and process state are excluded unless the assigned treatment
explicitly represents them. This isolates inherited information from inherited
filesystem state.

Defining the structured treatment required more than asking a model to “write a
good summary.” The current schema separates:

- normalized trace events;
- inspected files and executed commands;
- observations;
- rejected hypotheses;
- unresolved failures;
- the failed patch and its public-evaluator status;
- semantic propositions;
- linguistic realizations of those propositions;
- references from factual claims to observed events;
- content hashes for events and bundles.

Every factual observation, rejected hypothesis, and unresolved failure must
reference evidence. The validator rejects private-reasoning fields and prevents
hidden-evaluator events from entering participant-visible context.

This does not prove that a proposition is true. It establishes a narrower
property: the claim is attributable to specific, untampered observed material.
An independent executable evaluator remains the authority for task resolution.

I will call this layer an **epistemic IR**: an intermediate representation of
what the system observed, what it currently believes, what it rejected, and
what remains unresolved.

The term is useful as an architectural description, not as a claim that the
schema captures knowledge completely.

## Meaning and wording are not interchangeable

The handoff design also records something that agent-memory formats can easily
treat as incidental: the language in which a proposition is realized.

Work on motion-event lexicalization shows why that matters. Languages can
package path, manner, cause, and other components of an event differently.
Talmy's work on [lexicalization patterns](https://dingo.sbs.arizona.edu/~hharley/courses/PDF/TalmyLexicalizationPatterns.pdf)
and Slobin's account of
[thinking for speaking](https://benjamins.com/catalog/prag.1.1.01slo)
motivated decades of research into how language-specific expression can guide
attention during formulation. Experimental results do not justify a simple
claim that language determines thought, but they make one engineering
assumption unsafe: that translating or summarizing a trajectory preserves every
operationally relevant distinction.

Multilingual LLM evaluations provide a related warning. Prompt language and
mixed-language reasoning can change measured behavior. That does not prove a
language effect in coding-agent handoffs, so the current protocol records and
controls language rather than treating it as a causal factor.

The design therefore separates a proposition from its realizations. A stable
subject-predicate-object structure can retain evidence references while English
and Portuguese realizations express it differently.

This separation is deliberately modest. The structured proposition is not a
universal, language-independent truth. It is another model, designed to be more
stable and auditable than free text. It may omit distinctions that a particular
language expresses naturally. Its adequacy must be tested rather than assumed.

That boundary matters for any proposed AI-native representation. Replacing text
with a graph does not remove interpretation. It changes where interpretation
happens and makes some of its consequences inspectable.

## Source code has the same representational problem

Coding agents currently operate primarily on source text:

```text
read file → infer structure → edit characters → parse → type-check → test
```

This workflow repeatedly converts between a structured program and a textual
projection. It is appropriate for human-authored repositories because source
code is the shared artifact humans review and maintain. It is not obvious that
text should remain the primary authoring target for autonomous agents.

A coding agent changing a function usually cares about semantic operations:

- locate a symbol;
- inspect its callers and effects;
- add a parameter;
- preserve a public contract;
- replace an implementation;
- introduce a database read;
- prove or test a postcondition;
- identify which requirements the change satisfies.

A character diff only indirectly represents those operations.

Compiler infrastructure already demonstrates that one program can have several
representations. [MLIR](https://mlir.llvm.org/docs/LangRef/) supports
human-readable text, in-memory structures, compact serialization, multiple
dialects, and progressive lowering toward target-specific code. The important
idea is not that coding agents should emit MLIR directly. It is that no single
surface representation needs to serve authors, analyzers, optimizers, storage,
and machines equally well.

Recent experimental languages are beginning to explore this space. AIRL uses a
[typed graph and semantic patches](https://github.com/osalabs/airl) as the
authoring target for coding agents. Edict uses a
[JSON AST with types, effects, contracts, and Wasm generation](https://github.com/Sowiedu/Edict).
[NURL](https://nurl-lang.org/) designs a compact grammar around short dependency
windows and predictable prefix structure. Research on the
[SimPy grammar](https://arxiv.org/abs/2404.16333) found a more modest token
reduction—roughly 10–13.5 percent depending on the tokenizer—than the dramatic
savings sometimes assumed for AI-oriented syntax.

There is no established standard here, and the existence of prototypes is not
evidence that the architecture will outperform mature languages in real
repository work. It does show that the design space is now active enough that a
new project needs a sharper contribution than “an AST in JSON that compiles to
Wasm.”

## From epistemic IR to action IR

The handoff problem and the program-representation problem are not identical.
The first transfers knowledge between attempts. The second represents software
and changes to software. They share a deeper boundary:

> Separate stable semantic objects from the language or target used to realize
> them, and bind consequential claims to evidence or constraints.

One possible stack looks like this:

```text
observed trajectory
        ↓
evidence IR
  events, commands, files, hashes
        ↓
epistemic IR
  observations, hypotheses, uncertainty
        ↓
intent and change IR
  goals, typed transformations, effects, contracts
        ↓
projections
  handoff text · source code · diff · visualization
        ↓
execution and verification
        ↓
new observed events
```

In this architecture, a successor does not inherit an authoritative narrative.
It receives a checked semantic bundle and one or more appropriate projections.
A coding agent does not necessarily rewrite a file. It proposes a typed,
atomic, reversible semantic change.

Conceptually, such a change might say:

```text
change AddUserLookup
  select symbol UserService.find

  require
    input.id is validated
    result.id == input.id when found

  allow effects
    db.read<User>

  transform
    add parameter id: UserId
    replace implementation with DbGet(User, id)

  preserve
    public API compatibility
    authorization policy

  discharge
    typecheck
    tests [user_lookup, authorization]
```

The example is intentionally not a proposed final syntax. The canonical object
could be a graph or typed tree manipulated through structured tool calls. A
compact agent syntax and a readable human syntax could both be projections.

The essential properties are more important than punctuation:

- stable symbol and node identities;
- explicit types, capabilities, and effects;
- atomic semantic patches rather than full-file rewrites;
- preconditions, postconditions, and preserved invariants;
- requirement-to-change traceability;
- executable verification obligations;
- deterministic lowering to existing ecosystems;
- reversible operations and auditable provenance.

## Language or ISA is a false choice

An AI-native programming language and a semantic instruction set solve
different layers of the same problem.

The language provides composition: functions, modules, abstraction, types,
contracts, and reusable domain concepts. A small semantic core provides the
trusted operations and operational meaning. Backends lower those operations to
TypeScript, Rust, SQL, Wasm, cloud APIs, or other targets.

```text
agent-oriented language
  abstractions and composition
            ↓
semantic core IR
  minimal checked operations
            ↓
target adapters
  ecosystems and platforms
```

Starting with the core is attractive because it forces the project to define
semantics before syntax. It also works with current models through structured
tool calls, without requiring a custom tokenizer or model retraining. Surface
languages can then evolve from evidence about which representations models and
humans use effectively.

Compilation does not erase the existing software ecosystem. Machine backends
can handle instruction sets and operating-system boundaries, but databases,
frameworks, cloud services, organization-specific schemas, and business rules
still need versioned semantic adapters. The ecosystem cost moves into checked
bindings; it does not disappear.

## Structure removes some errors, not the hard problem

A semantic IR can make entire categories of failure impossible or easier to
detect:

- invalid syntax;
- references to unavailable symbols;
- undeclared effects;
- malformed transformations;
- unauthorized capabilities;
- stale evidence references;
- changes that violate locally expressible type or contract rules.

It cannot automatically establish that the user's intent was understood.

A perfectly typed program can implement the wrong cancellation policy. A proof
can establish that an implementation satisfies a specification while the
specification misrepresents the product requirement. A content hash can show
that evidence was not modified without showing that the associated conclusion
is complete.

This distinction mirrors the evidence-carrying handoff protocol:

```text
integrity ≠ truth
well-typed ≠ intended
verified against a contract ≠ correct contract
```

The purpose of structure is not to claim perfect agents. It is to reduce the
surface on which probabilistic reconstruction is unnecessarily repeated and to
make remaining uncertainty explicit.

## A falsifiable experiment

The useful next step is not to build a general-purpose language and wait for an
ecosystem. It is to test a narrow representation against ordinary file editing.

I would begin with repository maintenance in one target language and a small
semantic core:

- pure functions and algebraic data types;
- `Option` and `Result` instead of implicit null and exceptions;
- total pattern matching;
- declared database, HTTP, and filesystem effects;
- a closed symbol catalog;
- stable node identifiers;
- atomic semantic patches;
- a TypeScript projection and an IR interpreter;
- no general theorem prover in the first version.

The same model would attempt matched tasks under two policies:

```text
A: read and edit source files
B: query semantic state and propose checked IR patches
```

The comparison should measure total model tokens, hidden-test Pass@1, syntax and
type failures, repair cycles, time to resolution, required context, unsupported
task rate, and human comprehension of the projected change.

The unsupported-task rate is especially important. A small language can look
reliable by refusing to represent difficult work. Expressiveness and escape
hatches must be reported alongside correctness.

A credible initial result would not need an 80 percent token reduction or near-
perfect Pass@1. Eliminating syntax and nonexistent-symbol failures while
reducing total tokens or repair cycles without lowering hidden-test resolution
would already justify a larger experiment.

Negative evidence would be useful too. If models struggle more with structured
IR than familiar source code, if projections lose critical intent, or if adapter
maintenance dominates the savings, those results should constrain the design
before it becomes a platform.

## Text remains essential

This thesis is not an argument for hiding software from humans.

Humans still need readable projections, semantic diffs, counterexamples,
source maps, explanations, and reproducible builds. Incident response,
regulation, security review, and ordinary maintenance all require inspectable
artifacts.

The claim is narrower:

> Human-readable text should remain a first-class projection without
> necessarily remaining the only canonical state of agent work.

That applies to handoffs as much as programs. A human may prefer a concise
English explanation. A receiving agent may benefit from a typed evidence graph.
An auditor may need the referenced command events. These views can coexist if
they derive from the same versioned semantic object.

## From workflow truth to semantic state

I previously described GPTCode's working principle as
[“The workflow is the source of truth.”]({{ '/blog/2026-07-25-the-workflow-is-the-source-of-truth' | relative_url }})

The evidence-carrying handoff design extends that principle.

A workflow can define stages, permissions, and verification boundaries, but its
state should not remain trapped in fluent transcripts. If research findings,
rejected hypotheses, planned changes, implementation effects, and verification
evidence matter to later decisions, they need identities and contracts of their
own.

Models generate possibilities. Repositories define constraints. Verification
establishes executable evidence. A semantic IR can preserve the relationships
between them.

The transcript is still valuable. It may contain nuance that the structured
representation missed. It can help a human understand how a failure unfolded.
It can serve as evidence for improving the schema.

But it is not the state.
