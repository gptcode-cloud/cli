# E2E Test Coverage Expansion Plan

**Document:** Strategic plan for comprehensive test coverage of GPTCode  
**Status:** In Development  
**Last Updated:** 2025-11-25

## Executive Summary

This document outlines a systematic approach to expand E2E test coverage from basic command execution to comprehensive real-world scenarios, with emphasis on local model validation (Ollama) and quality assurance.

## Current State

### ✅ Phase 1 Complete: Run Command Basics
- DevOps command execution with history
- Working directory and environment management
- Single-shot automation (CI/CD compatible)
- **Coverage:** 3 scenarios
- **Status:** All passing

### 🔄 Phase 2 In Progress: Core CLI Commands
- **Chat mode:** Multi-turn conversational testing
- **TDD:** Test generation and implementation workflows
- **Research/Plan:** Knowledge extraction and planning
- **Expected:** 5-7 scenarios

### ⏳ Phase 3 Planned: Local Model Validation
Focus on Ollama for cost-free, privacy-preserving testing
- Model selection and switching
- Hybrid setup (local + cloud)
- Performance baseline establishment

### ⏳ Phase 4 Planned: Neovim Integration
Headless testing of editor plugin
- Chat interface via RPC
- Model switching
- Profile management

### ⏳ Phase 5 Planned: Real Project Integration
Complete workflows on realistic codebases
- Go project workflows
- Elixir project workflows
- CI/CD integration

## Technical Architecture

### Test Framework Structure

```
tests/e2e/
├── e2e.sh                 # Main runner
├── lib/
│   └── helpers.sh        # Shared utilities
├── scenarios/
│   ├── <name>.sh        # Individual scenario tests
│   └── fixtures/        # Test data and sample projects
└── E2E_ROADMAP.md      # Long-term vision
```

### Helper Functions Available

```bash
# Setup/Teardown
setup_test_dir "test name"
cleanup_test_dir

# Command Execution
run_chu_command "cmd" "arg1" "arg2"
run_chu_with_input "cmd" "input" "flags"

# Assertions
assert_exit_code 0
assert_contains "$OUTPUT" "expected"
assert_not_contains "$OUTPUT" "unexpected"
assert_file_exists "path"
assert_dir_exists "path"

# Test Project Creation
create_test_file "name" "content"
create_go_project "name"
```

## Phase 2 Implementation: Core CLI Commands

### 2.1 Chat Mode Testing

**Problem Being Solved:**
- Chat mode needs to preserve context across multiple turns
- File context must be automatically loaded
- LLM responses must reference the correct code

**Scenarios:**

#### conversational_code_exploration.sh
```bash
# What it tests:
1. Create Go project with sample code
2. Ask about struct definition
3. Ask about function logic
4. Ask for improvement suggestions
5. Verify context is being used

# Why it matters:
- Validates that conversations stay grounded in actual code
- Ensures file discovery works automatically
- Tests context window management
```

#### code_explanation_with_errors.sh
```bash
# What it tests:
1. Create project with buggy code
2. Ask to explain error behavior
3. Request fix suggestions
4. Validate understanding

# Why it matters:
- Real-world debugging scenario
- Tests LLM's ability to identify issues
- Validates explanatory quality
```

### 2.2 TDD Workflow Testing

**Problem Being Solved:**
- TDD mode needs to generate tests before implementation
- Generated tests must be executable
- Implementation suggestions must pass tests

**Scenarios:**

#### tdd_new_feature_development.sh
```bash
# What it tests:
1. Create empty Go project
2. Request TDD for calculator
3. Validate test generation
4. Check implementation structure

# Why it matters:
- Tests core TDD workflow
- Validates test quality
- Ensures testability of suggestions
```

#### tdd_bug_fix_with_regression_test.sh
```bash
# What it tests:
1. Create project with bug
2. Generate failing test for bug
3. Implement fix
4. Verify all tests pass

# Why it matters:
- Real debugging scenario
- Tests regression prevention
- Validates fix quality
```

### 2.3 Research & Planning

**Problem Being Solved:**
- Need to understand existing codebases before implementing
- Plans must be actionable and verified
- Changes must follow planned approach

**Scenarios:**

#### research_and_planning_workflow.sh
```bash
# What it tests:
1. Create project with auth system
2. Research how auth works
3. Plan JWT token addition
4. Validate plan structure

# Why it matters:
- Tests knowledge extraction
- Validates plan quality
- Ensures structured approach
```

#### research_multistep_analysis.sh
```bash
# What it tests:
1. Create complex project
2. Research interconnected systems
3. Identify dependencies
4. Propose refactoring

# Why it matters:
- Tests deep codebase understanding
- Validates cross-module analysis
- Ensures recommendation quality
```

## Phase 3: Local Model Validation

### Purpose
Establish that GPTCode works well with free, local models (Ollama) without depending on cloud APIs.

### Key Scenarios

#### automatic_model_selection.sh
```bash
# Prerequisites:
- Ollama running with qwen2.5-coder:7b
- Multiple models configured

# Tests:
1. Run chu do without specifying model
2. Verify automatic selection
3. Check selection reasoning
4. Validate choice is appropriate
```

#### model_retry_on_validation_failure.sh
```bash
# Prerequisites:
- Multiple Ollama models available
- Validator that can detect failures

# Tests:
1. Trigger scenario that fails first attempt
2. Verify automatic model switching
3. Check success on retry
4. Validate cost/performance tradeoff
```

#### ollama_local_only_execution.sh
```bash
# Prerequisites:
- Ollama running locally
- No cloud backend configured

# Tests:
1. Run all chu commands
2. Monitor network traffic (verify no external calls)
3. Validate privacy preservation
4. Check performance metrics
```

## Phase 4: Neovim Integration Testing

### Architecture

Neovim will be run in headless mode using RPC calls:

```bash
# Start Neovim headless with plugin
nvim --headless +"call rpcrequest(1, 'nvim_exec_lua', 'require(\"gptcode\").setup()', {})" ...

# Send RPC calls to test functionality
nvim_call "GPTCodeChat"
nvim_send_keys "test message"
nvim_get_buffer_content
```

### Key Scenarios

#### nvim_chat_interface_headless.sh
```bash
# Tests:
1. Start Neovim headless
2. Trigger :GPTCodeChat
3. Send message via RPC
4. Verify response buffer
```

#### nvim_model_switching.sh
```bash
# Tests:
1. Open model selector (<C-m>)
2. Switch model via RPC
3. Verify in chat session
4. Validate model change reflected
```

## Phase 5: Real Project Integration

### Go Project Workflow

```bash
# Steps:
1. Clone/create sample Go project (real structure)
2. Run chu research "authentication"
3. Generate plan for new feature
4. Execute with chu do
5. Verify: tests pass, code compiles, no regressions
```

### Elixir Project Workflow

```bash
# Steps:
1. Create Elixir project
2. Use chu tdd for new feature
3. Verify ExUnit tests generated
4. Check implementation compiles
5. Run mix test to validate
```

## Success Metrics

### Quantitative
- **Coverage:** >70% of CLI commands have E2E tests
- **Scenarios:** 30+ realistic test scenarios
- **Pass Rate:** 100% with Ollama, 95%+ with cloud models
- **Performance:** <5 minutes total runtime
- **Flakiness:** <2% retry rate

### Qualitative
- Tests accurately represent user workflows
- Failure messages are clear and actionable
- Documentation is comprehensive
- New developers can add scenarios easily

## Implementation Timeline

### Week 1-2: Phase 2 (Core Commands)
- [ ] Implement chat mode scenarios (3)
- [ ] Implement TDD scenarios (3)
- [ ] Implement research/planning scenarios (3)
- [ ] Fix any blocking issues
- [ ] All 9 scenarios passing

### Week 3-4: Phase 3 (Local Models)
- [ ] Setup Ollama in CI/CD
- [ ] Implement model selection scenarios (3)
- [ ] Validate local-only execution
- [ ] Performance baseline established
- [ ] Cost analysis completed

### Week 5-6: Phase 4 (Neovim)
- [ ] Setup headless Neovim RPC
- [ ] Implement plugin test scenarios (4)
- [ ] Validate in CI/CD headless mode
- [ ] All plugin functions tested

### Week 7-8: Phase 5 (Real Projects)
- [ ] Create sample Go project
- [ ] Create sample Elixir project
- [ ] Implement real workflow scenarios (5)
- [ ] End-to-end validation complete

## Testing Environment Requirements

### Minimum Setup
- Go 1.22+
- Bash 4+
- Ollama with at least 1 model installed

### Recommended Setup
- Ollama: qwen2.5-coder:7b (7B parameters, fast)
- Optional: Groq API key for cloud comparison
- Optional: Neovim 0.10+ for plugin testing
- GitHub Actions for CI/CD

### CI/CD Considerations
- Run tests in fresh Docker container
- Install Ollama model before running tests
- Cache Ollama models to save bandwidth
- Run in headless mode for all tests
- Timeout: 10 minutes per test, 30 minutes total

## Failure Handling

### Expected Failures
1. **Model not installed** → Clear error message, skip scenario
2. **API rate limit** → Retry with exponential backoff
3. **Network timeout** → Mark as flaky, investigate
4. **Validation failure** → Log plan/output, flag for review

### Investigation Procedure
```bash
# 1. Run single scenario with debug
GPTCODE_DEBUG=1 ./tests/e2e/scenarios/scenario_name.sh

# 2. Check generated files
ls -la $TEST_DIR

# 3. Review model performance
chu ml predict "test query"

# 4. Validate configuration
chu config get defaults.backend
```

## Future Enhancements

### Testing Improvements
- [ ] Screenshot capture in Neovim tests
- [ ] Performance profiling
- [ ] Cost tracking per scenario
- [ ] Comparison matrix (local vs cloud)
- [ ] Regression detection

### Coverage Expansion
- [ ] Multi-language projects (Python, Ruby, Elixir)
- [ ] Monorepo scenarios
- [ ] Large codebase handling
- [ ] Team collaboration workflows
- [ ] Integration with popular IDEs

### Infrastructure
- [ ] Distributed test execution
- [ ] Test result dashboard
- [ ] Performance trends tracking
- [ ] Model performance comparison

## Documentation Requirements

### For Each Scenario
```bash
#!/usr/bin/env bash
#
# Scenario: <Name>
#
# Problem: What real-world issue does this solve?
# Workflow: What user journey does this represent?
# Validation: What proves success?
```

### README for New Developers
- How to add new scenarios
- Common assertions and utilities
- Debugging tips
- CI/CD integration

### Architecture Documentation
- Test flow diagrams
- Dependency graph
- Model selection logic
- Context management

## Maintenance Plan

### Weekly
- Monitor test pass/fail rates
- Review new test additions
- Update model recommendations

### Monthly
- Performance analysis
- Cost reports
- User feedback integration
- Model benchmark updates

### Quarterly
- Plan next phase implementation
- Evaluate new tools/frameworks
- Architecture review

## Related Documents
- `tests/E2E_ROADMAP.md` - High-level phases and timeline
- `.github/workflows/e2e.yml` - CI/CD configuration
- `docs/testing-guide.md` - User-facing testing documentation
