# Language-Specific Guidelines

## General Principles (All Languages)

- Prefer small functions and modules/components.
- One responsibility per function; one concept per module.
- Use intention-revealing names (no `doWork` or `runStuff`).
- Avoid global state and hidden side-effects.
- Handle errors explicitly and predictably.
- Be consistent within the chosen stack.

## Elixir

- Use pure functions and pattern matching where it simplifies logic.
- Use ExUnit with `describe` blocks and clearly named tests.
- Keep context modules focused; avoid monolithic "everything" contexts.
- Do not overuse macros; prefer plain functions and modules.

## Ruby / Rails

- Follow Sandi Metz principles: tiny objects, tiny methods.
- Keep controllers thin; push domain logic into POROs / service objects.
- Keep ActiveRecord models focused on persistence and simple invariants.
- Use clear service object naming: `Orders::ProcessPayment`, `Users::SignupService`, etc.

## TypeScript / JavaScript

- Use modern TypeScript: strict typing, `async/await`, ES modules.
- Prefer pure functions and data transformation pipelines.
- Avoid unnecessary classes; use them only when modeling stateful domain objects.
- For tests, prefer Vitest or Jest with small, focused test suites.

## Go

- Use small packages with clear boundaries.
- Prefer simple, explicit interfaces defined at the consumer side.
- Use table-driven tests with meaningful case names.
- Handle errors explicitly; avoid panics for normal control flow.

## Error Handling and Edge Cases

You must always think about edge cases:

- Empty inputs (empty lists, nil/undefined/null, missing keys).
- Invalid inputs (wrong types, out-of-range values).
- External failures (network, IO, DB).

For each critical behavior:

- Mention at least the main edge cases and how they are handled.
- In tests, include both happy path and at least one failure/edge scenario.
