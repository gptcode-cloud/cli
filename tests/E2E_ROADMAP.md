# E2E Testing Roadmap

This document outlines the plan for comprehensive end-to-end testing of GPTCode.

## Current Status (Phase 1) ✅

**Implemented:** Basic `chu run` command scenarios
- DevOps command execution with history
- Working directory and environment management  
- Single-shot command for automation

**Framework:** Test harness with helpers, assertions, and scenario runner

## Phase 2: Core CLI Commands 🔄

### 2.1 Chat Mode Testing
**Scenarios:**
- `conversational_code_exploration.sh`
  - Developer asks about existing code
  - Follow-up questions maintain context
  - Tests multi-turn conversation
  
- `code_explanation_workflow.sh`
  - Request explanation of complex function
  - Ask for simplification suggestions
  - Validate context preservation

### 2.2 TDD Workflow Testing
**Scenarios:**
- `tdd_new_feature_development.sh`
  - Create new module with tests first
  - Validate test generation
  - Check implementation against tests
  - Verify test execution

- `tdd_bug_fix_with_regression_test.sh`
  - Reproduce bug scenario
  - Generate failing test
  - Fix implementation
  - Verify test passes

### 2.3 Research → Plan → Implement Workflow
**Scenarios:**
- `complete_feature_workflow.sh`
  - Research existing auth system
  - Plan password reset feature
  - Implement with step verification
  - Validate end-to-end flow

- `refactoring_with_planning.sh`
  - Research current structure
  - Plan refactoring approach
  - Implement incremental changes
  - Verify no breaking changes

## Phase 3: Local Model Validation 🎯

**Goal:** Ensure quality with Ollama models (free, local, privacy-focused)

### 3.1 Model Selection Testing
**Scenarios:**
- `automatic_model_selection.sh`
  - Test `chu do` auto-selects best models per agent
  - Verify different models for router/query/editor
  - Validate fallback to defaults when needed

- `model_switching_on_failure.sh`
  - Trigger validation failure
  - Verify automatic retry with different model
  - Check success after model switch

### 3.2 Ollama-Specific Testing
**Scenarios:**
- `ollama_local_execution.sh`
  - Configure Ollama as backend
  - Test all agents with local models
  - Verify no external API calls
  - Validate privacy preservation

- `hybrid_ollama_cloud_setup.sh`
  - Router: local (llama 8B)
  - Editor: cloud (GPT-4)
  - Verify switching works correctly
  - Test cost optimization

## Phase 4: Neovim Headless Testing 🚀

**Goal:** Validate Neovim plugin without GUI

### 4.1 Basic Plugin Functions
**Scenarios:**
- `nvim_chat_interface.sh`
  - Open Neovim headless
  - Trigger `:GPTCodeChat`
  - Send message via RPC
  - Verify response in buffer

- `nvim_model_switching.sh`
  - Open model selector (`<C-m>`)
  - Switch profile via RPC
  - Verify new models loaded
  - Test in chat session

### 4.2 Profile Management
**Scenarios:**
- `nvim_profile_creation.sh`
  - Create new profile via Neovim
  - Configure agent models
  - Save and load profile
  - Verify persistence

- `nvim_model_search_install.sh`
  - Search for models (`<leader>ms`)
  - Install Ollama model
  - Set as active model
  - Verify in chat

## Phase 5: Integration & Real Projects 🏗️

### 5.1 Real Codebase Testing
**Scenarios:**
- `go_project_full_workflow.sh`
  - Clone sample Go project
  - Run `chu research "authentication"`
  - Generate plan for new feature
  - Implement with `chu do`
  - Verify tests pass

- `elixir_project_tdd_workflow.sh`
  - Create new Elixir module
  - Use `chu tdd` for development
  - Verify ExUnit tests generated
  - Check implementation quality

### 5.2 CI/CD Integration
**Scenarios:**
- `github_actions_workflow.sh`
  - Simulate GitHub Actions environment
  - Run automated code review
  - Generate test coverage report
  - Validate no regressions

## Implementation Strategy

### Test Environment Setup
```bash
# Each scenario should:
1. Setup isolated test directory
2. Configure Ollama if needed
3. Set up mock project (Go, Elixir, etc.)
4. Run chu commands
5. Validate outputs
6. Clean up
```

### Model Requirements
- **Minimum:** qwen2.5-coder:7b (Ollama, free)
- **Recommended:** llama3.3:70b (Groq, cheap) + qwen2.5-coder:7b (local)
- **Test Matrix:** Both local-only and hybrid setups

### Success Criteria
- ✅ All scenarios pass with Ollama models
- ✅ All scenarios pass with Groq models
- ✅ Neovim headless tests work without GUI
- ✅ Real project workflows complete successfully
- ✅ No false positives or flaky tests
- ✅ Tests run in < 5 minutes total

## Running Tests

```bash
# Run all tests
./tests/e2e.sh

# Run specific phase
./tests/e2e.sh --phase 2

# Run with specific backend
E2E_BACKEND=ollama ./tests/e2e.sh

# Run single scenario
./tests/e2e/scenarios/conversational_code_exploration.sh
```

## Priority Order

1. **Phase 2.1 & 2.2** - Core CLI commands (chat, tdd)
2. **Phase 3.1** - Model selection validation
3. **Phase 3.2** - Ollama local testing
4. **Phase 4.1** - Neovim basic functions
5. **Phase 5.1** - Real project integration
6. **Phase 4.2** - Neovim advanced features
7. **Phase 5.2** - CI/CD integration

## Notes

- Tests should be **deterministic** - no random failures
- Use **realistic scenarios** from actual user workflows
- **Document failures** - when a test fails, it should be clear why
- **Keep scenarios focused** - one primary concern per test
- **Use mocks sparingly** - prefer real commands over mocks
