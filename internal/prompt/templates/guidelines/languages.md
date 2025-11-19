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

## Python
- Use `pytest` for testing.
- Use type hints (PEP 484) and check with `mypy` if possible.
- Follow PEP 8 style guidelines.
- Prefer list comprehensions and generators over loops where appropriate.
- Use `pathlib` instead of `os.path`.

## Rust
- Use `cargo test` and built-in testing framework.
- Follow idiomatic Rust: use `Result`/`Option` instead of panics.
- Avoid `unwrap()` or `expect()` in production code; handle errors properly.
- Use `clippy` for linting.
- Prefer iterators over loops.

## C / C++
- Use modern C++ (C++17/20) features: smart pointers, `std::optional`, `auto`.
- Avoid raw pointers; use RAII for resource management.
- Use `CMake` or `Make` for build systems.
- Use Google Test or Catch2 for testing.
- Header files should have `#pragma once` or include guards.

## Error Handling and Edge Cases

You must always think about edge cases:

- Empty inputs (empty lists, nil/undefined/null, missing keys).
- Invalid inputs (wrong types, out-of-range values).
- External failures (network, IO, DB).

For each critical behavior:

- Mention at least the main edge cases and how they are handled.
- In tests, include both happy path and at least one failure/edge scenario.
