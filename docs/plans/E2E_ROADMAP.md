# E2E Roadmap

This document lists the end‑to‑end test scenarios for **Chuchu**. It is located in `docs/plans/` as requested.

## Phase 2 – Core CLI Commands
- **conversational_code_exploration.sh** – Tests multi‑turn chat context with a sample Go project.
- **tdd_new_feature_development.sh** – Validates the TDD workflow by generating tests before implementation.
- **research_and_planning_workflow.sh** – Exercises the research & planning mode on a sample authentication project.

## Phase 3 – Local Model Validation (Ollama)
- **automatic_model_selection.sh** – Verifies automatic model selection when multiple Ollama models are available.
- **model_retry_on_validation_failure.sh** – Ensures the system retries with a fallback model on validation failures.
- **ollama_local_only_execution.sh** – Confirms that all commands run without any external network traffic.

## Phase 4 – Neovim Integration
- **nvim_chat_interface_headless.sh** – Runs the Neovim chat interface in head‑less mode via RPC.
- **nvim_model_switching.sh** – Tests model switching through Neovim keybindings.

## Phase 5 – Real‑Project Integration
- **real_go_project_workflow.sh** – Executes a full workflow on a realistic Go codebase (clone, research, plan, implement).
- **real_elixir_project_workflow.sh** – Executes a full workflow on an Elixir codebase, including test generation and execution.

Each script is executable (`chmod +x`) and exits with status 0 on success. The CI pipeline runs all scripts with:
```bash
chmod +x tests/e2e/scenarios/*.sh && ./tests/e2e/scenarios/*.sh
```
