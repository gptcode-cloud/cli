---
layout: post
title: "When Should a Coding Agent Go to the Cloud?"
date: 2026-08-20
author: Jader Correa
tags: [ai-agents, local-first, model-routing, evaluation, research, open-source]
description: "What a six-task calibration taught me about local-first coding-agent routing, false acceptance, hosted cost, and knowing when not to scale an experiment."
---

Local models make an appealing promise for coding agents: keep work private,
avoid hosted inference cost, and use the cloud only when a task truly needs it.

The difficult part is not switching providers. It is deciding **when a local
attempt is trustworthy enough to accept and when the system should start over
in the cloud**.

I recently completed a calibration pilot built around that decision. The
result was useful, but not because the original hypothesis won.

The tested local-first policy reduced observed hosted inference cost by 45.56%
relative to a paired cloud-only condition. It also resolved fewer tasks: 2/6
instead of 3/6. One patch that looked acceptable to the router failed the
held-out evaluator.

That was enough evidence to stop the planned confirmatory experiment.

This post explains the hypothesis, what the pilot measured, why the no-go was
the correct outcome, and how the failure changed the next version of the
experiment. The complete protocol, evidence, analysis, and checksums are
available in the public
[2026.08.20.1 research release](https://github.com/jadercorrea/ai-experiments/releases/tag/2026.08.20.1).

## From backend choice to routing policy

An earlier GPTCode article described
[switching between local and cloud backends]({{ '/blog/2025-11-18-hybrid-cloud-local' | relative_url }})
as a configuration choice. That remains useful, but it leaves the most
interesting systems question unanswered:

> Can an agent try a task locally, observe only legitimate execution signals,
> and escalate to a stronger hosted model without materially reducing verified
> resolution?

The phrase “legitimate execution signals” matters. A router can observe whether
the local agent timed out, whether inference failed, whether a material patch
exists, and whether public repository checks pass. It cannot inspect the hidden
tests later used to determine whether the issue was actually resolved.

This makes local-first routing a policy-evaluation problem, not a model beauty
contest. The unit being evaluated is the complete system:

```text
model + agent scaffold + tools + time budget + routing rule + evaluator
```

A local model can be capable while the routing policy is unsafe. A cloud model
can be stronger while an unnecessary local attempt still makes the complete
workflow slower. The comparison has to preserve those system effects.

## The three conditions

The pilot used six real issue-derived Go tasks from six repositories, with two
tasks in each of three provisional complexity strata. Each task was evaluated
under three policies:

1. **Local-only:** Qwen3-Coder 30B A3B Q4_K_M through Ollama.
2. **Cloud-only:** Claude Sonnet 4.6 through Amazon Bedrock.
3. **Local-first clean fallback:** attempt the local policy, then either accept
   the local patch or reset to the immutable base and invoke the cloud policy.

The local-first router escalated on four observable conditions:

- hard timeout;
- backend inference failure;
- no material patch;
- failure of repository-provided public verification.

It accepted a local result only after a clean agent exit, a material patch, no
recorded inference failure, and successful public verification. Hidden
evaluation was isolated from the routing decision.

The models, prompts, schedules, timeouts, cost cap, fallback semantics, and
routing implementation were frozen before the local-first outcomes were
generated. Invalid runs were retained with explicit invalidation records
instead of being silently discarded.

## What happened

| Policy | Resolved | Hosted cost | Total wall time |
| --- | ---: | ---: | ---: |
| Local-only | 3/6 | USD 0 | 9,479.207 s |
| Cloud-only Sonnet 4.6 | 3/6 | USD 7.417368 | 2,365.354 s |
| Local-first clean fallback | 2/6 | USD 4.037955 | 10,627.930 s |

The cost result initially looks encouraging. Local-first spent 45.56% less on
hosted inference than cloud-only, exceeding the pre-specified 30% cost-reduction
target.

But cost was only one side of the hypothesis.

Local-first resolved 2/6 tasks while cloud-only resolved 3/6. The descriptive
resolution-rate difference was -16.67 percentage points, outside the frozen
-10 percentage-point non-inferiority margin.

With six calibration tasks, that difference is not a population estimate. It
does not prove that local-first routing is generally inferior. It does show
that this policy and this design were not ready for confirmatory claims.

The router accepted two tasks locally and escalated four. The escalations came
from two hard timeouts, one backend failure, and one no-material-patch outcome.
Of the two local acceptances, one resolved the issue. The other passed every
observable routing check but failed held-out evaluation.

That false acceptance is the most important result in the pilot.

## Public checks are not the same as resolution

A router operates before the hidden evaluator. It therefore needs a proxy for
“safe to accept.” In v1, that proxy combined process health, patch materiality,
and public tests.

The Fiber task exposed the boundary of that rule. The local patch looked valid
from the router's perspective:

- the agent exited cleanly;
- inference did not fail;
- a material patch existed;
- public verification passed.

The held-out evaluator still rejected it.

This is not a harness invalidation. It is evidence that the acceptance rule had
insufficient precision. If a local-first policy saves money partly by accepting
plausible but incorrect patches, the saving is not operationally meaningful.

The system needs stronger post-local signals without leaking hidden outcomes
into the router. That may include more informative repository-owned checks,
patch-risk features, or a policy that sends ambiguous tasks directly to the
cloud before paying the latency of a local attempt.

## Local inference was not free

The hosted-cost column records provider spend. It does not monetize local
hardware, electricity, or opportunity cost.

Wall time makes that omission visible. Cloud-only completed its six calibration
trajectories in about 39 minutes of recorded total wall time. Local-only took
about 2 hours and 38 minutes. Local-first took about 2 hours and 57 minutes.

Two local-first tasks consumed the full local timeout before escalating. One of
those later consumed the cloud timeout as well. Another completed its local
stage but required fallback after a backend failure.

“Try local first” sounds inexpensive when only API charges are counted. At the
system level, it can add substantial latency and monopolize a developer machine
before the cloud attempt even begins.

That is why the next policy cannot simply run every task locally and wait for a
failure. It needs a pre-treatment admission decision.

## Calibration prevented an expensive weak study

The original confirmatory plan used 90 tasks, 5 trajectories per task-policy
pair, and 1,350 runs across the three policies. Under its planning assumptions,
simulation estimated 91.28% joint power for the resolution and cost criteria.

Completed calibration changed the assumptions:

| Simulation scenario | Joint power |
| --- | ---: |
| Frozen planning assumptions | 91.28% |
| Observed escalation rate substituted | 25.92% |
| Observed escalation and resolution substituted | 1.18% |

An exploratory search recovered more than 90% simulated joint power only at a
tested point with 1,080 tasks, 270 repositories, and 48,600 policy runs. Under
calibration-based projections, that point implied roughly USD 34,137 of hosted
inference and 16,854 serial hours.

Those figures are feasibility projections, not money spent or a recommended
design. Their purpose is to show why proceeding with the original treatment
would have been difficult to justify.

The calibration did its job: it discovered that a planned large experiment
would no longer answer the question with the precision originally expected.

## A no-go is progress when it is recorded

It is tempting to treat a pilot as preliminary work that disappears when the
main study changes. That would remove exactly the evidence needed to understand
why the later design exists.

I published this pilot because it records:

- the original hypothesis and frozen decision rule;
- valid, failed, and invalidated trajectories;
- the cost-resolution tradeoff actually observed;
- a concrete false-acceptance failure;
- the power analysis that stopped the confirmatory campaign;
- the methodological boundary between v1 and its successor.

The publication is deliberately bounded. It supports claims about the method,
infrastructure, routing behavior, and feasibility of this calibration. It does
not support population non-inferiority claims.

The canonical artifacts are:

- [calibration pilot report](https://github.com/jadercorrea/ai-experiments/blob/2026.08.20.1/experiments/coding-agents/local-first-routing/2026-07-30/CALIBRATION_PILOT_REPORT.md);
- [pilot manuscript](https://github.com/jadercorrea/ai-experiments/blob/2026.08.20.1/experiments/coding-agents/local-first-routing/2026-07-30/publication/PILOT_MANUSCRIPT.md);
- [machine-readable local-first summary](https://github.com/jadercorrea/ai-experiments/blob/2026.08.20.1/experiments/coding-agents/local-first-routing/2026-07-30/calibration/local-first-summary.json);
- [release archive and checksum](https://github.com/jadercorrea/ai-experiments/releases/tag/2026.08.20.1).

## What changes in v2

The successor is a separately named experiment, not an updated label placed on
the same treatment.

Its intended policy makes two decisions:

```text
pre-treatment task signals
          |
   +------+------+
   |             |
cloud direct   local attempt
                 |
          post-local signals
              |       |
            accept  fallback
```

The first decision avoids local attempts for tasks whose observable features
make them poor candidates. The second strengthens the evidence required to
accept a local result.

The six v1 repositories are development-only for the successor. New
construction, calibration, and confirmatory pools must be repository-disjoint.
The v2 policy, models, prompts, budgets, timeouts, gates, and hashes must be
frozen before its calibration runs begin.

Most importantly, v1 and v2 outcomes will not be pooled as if they were
repetitions of one unchanged treatment.

## The question remains open

The pilot did not answer when a coding agent should go to the cloud in general.
It answered a narrower and necessary question:

> Was this transparent local-first policy ready to support a larger
> confirmatory evaluation?

The answer was no.

That answer saved compute, cloud budget, and time. It also produced a better
research question for the next series: can a selective router recognize, using
only legitimate pre-treatment and post-local signals, when local execution is
worth attempting?

That is a more difficult question than choosing a backend in a configuration
file. It is also the question a reliable local-first agent system eventually
has to answer.

---

*Jader Correa is a principal engineer and founder working on AI agents,
developer tools, and distributed systems.*
