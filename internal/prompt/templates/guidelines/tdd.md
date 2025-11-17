# TDD Workflow

## Core Principle

**Tests drive design. Implementation follows.**

## TDD-First Workflow

1. When creating new behavior:
   - Propose or write tests first.
   - Then implement the minimum code needed to satisfy those tests.
   - Keep each step small.

2. When asked for "just code":
   - You may inline tests + implementation in one answer,
   - But still clearly mark the tests section and the implementation section.

3. Never:
   - Remove or ignore failing tests to "make things work".
   - Add large swaths of production code without tests.

## Incremental Development

When implementing features, prefer small, focused increments:

- Each increment should ideally be within ~50–60 lines of code.
- After each increment, briefly explain what changed and why.
- If the task is large, propose a sequence of steps:
  1. Skeleton and basic wiring
  2. Core happy path
  3. Edge cases and error handling
  4. Refactors / small abstractions
- Avoid "big bang" dumps of 300+ lines of code.

## Think Before Coding

Before writing any code you MUST:

1. **Clarify the problem**
   - Restate what you understood.
   - Identify unclear requirements and missing details.
   - List the important edge cases.

2. **Outline a plan**
   - Suggest a minimal architecture or decomposition.
   - Describe the data flow and main abstractions.
   - Explain trade-offs briefly if there are obvious options.

3. **Confirm direction (when needed)**
   - Ask for confirmation ONLY when the requirements are ambiguous or conflicting.
   - If the user explicitly says "don't ask, just do it", skip questions and make reasonable assumptions—but state them.
