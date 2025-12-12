# E2E Testing Roadmap

**Last Updated:** 2025-11-26  
**Status:** Phase 1 Complete ✅ | Phase 2-5 Planned

This document tracks E2E testing implementation progress for **GPTCode**.

## Implementation Approach

**Current (✅):** Go tests with `gptcode test e2e` command  
**Previous:** Bash scripts (deprecated, replaced by Go tests)

### Why Go Tests?
- Type-safe, maintainable test code
- Better error messages and debugging
- Integrated with standard Go tooling
- Progress tracking with real-time countdown
- macOS desktop notifications
- Profile-based test execution

---

## ✅ Phase 1 – E2E Infrastructure (COMPLETE)

### Command Implementation
- ✅ `gptcode test e2e` - Profile-based test execution
- ✅ `gptcode test e2e --interactive` - Interactive profile selection
- ✅ `gptcode test e2e run` - Category-based execution
- ✅ `gptcode test e2e --notify` - macOS desktop notifications
- ✅ `gptcode test e2e --timeout N` - Custom timeout configuration

### Test Runner Features
- ✅ Real-time progress bar with countdown
- ✅ Live test status (passed/failed/skipped)
- ✅ Profile configuration from setup.yaml
- ✅ Environment variable injection (E2E_BACKEND, E2E_PROFILE, E2E_TIMEOUT)
- ✅ Automatic test discovery in categories

### Current Tests (tests/e2e/run/)
- ✅ `TestE2EConfiguration` - Validates E2E environment
- ✅ `TestGptcodeCommand` - Verifies gptcode binary availability
- ✅ `TestChuDoCreateFile` - File creation with content validation
- ✅ `TestChuDoModifyFile` - File modification validation
- ✅ `TestChuDoNoUnintendedFiles` - Extra file detection
- ⏭️ `TestChuDoTimeout` - Timeout validation (skipped, too slow with local Ollama)

### Configuration
```yaml
e2e:
  default_profile: local
  timeout: 600  # 10 minutes for local Ollama
  notify: true
  parallel: 1
```

---

## ✅ Phase 2 – Chat & Interactive Commands (COMPLETE)

### Goals
- ✅ Test `gptcode chat` single-shot and REPL mode
- ✅ Test conversation context management
- ✅ Validate response capture and history

### Implemented Tests (tests/e2e/chat/)
- ✅ `TestChatBasicInteraction` - Single Q&A
- ✅ `TestChatCodeExplanation` - Code understanding
- ✅ `TestChatFollowUp` - Conversation context validation
- ✅ `TestChatSaveLoadSession` - Session persistence
- ✅ `TestChatConversationContext` - Multi-turn context

### Unit Tests (internal/repl/)
- ✅ `TestContextManagerAddMessage` - Message addition
- ✅ `TestContextManagerGetContext` - Context retrieval
- ✅ `TestContextManagerClear` - History clearing
- ✅ `TestContextManagerTokenLimit` - Token limits
- ✅ `TestContextManagerMessageLimit` - Message limits
- ✅ `TestContextManagerGetRecentMessages` - Recent messages

---

## 🚧 Phase 3 – Research & Planning (PARTIAL)

### Goals
- ✅ Validate commands exist and show help
- ⏭️ Test `gptcode research` functionality (placeholder)
- ⏭️ Test `gptcode plan` generation (placeholder)
- ⏭️ Validate research → plan workflow (placeholder)

### Implemented Tests (tests/e2e/planning/)
- ✅ `TestResearchHelp` - Command exists
- ✅ `TestPlanHelp` - Command exists
- ✅ `TestTDDHelp` - Command exists  
- ✅ `TestDoHelp` - Command exists
- ✅ `TestCommandsExist` - All commands registered
- ⏭️ `TestResearchBasic` - Research output quality (skipped)
- ⏭️ `TestPlanGeneration` - Plan generation (skipped)
- ⏭️ `TestTDDWorkflow` - TDD workflow (skipped)

---

## 🚧 Phase 4 – Autonomous Execution (PLANNED)

### Goals
- Test `gptcode implement plan.md`
- Test `gptcode implement --auto` with verification
- Validate retry logic and error recovery

### Planned Tests (tests/e2e/integration/)
- [ ] `TestImplementInteractive` - Step-by-step execution
- [ ] `TestImplementAuto` - Autonomous with verification
- [ ] `TestImplementRetry` - Error recovery
- [ ] `TestImplementResume` - Checkpoint resume

---

## 🚧 Phase 5 – Real Project Workflows (FUTURE)

### Goals
- Test on realistic codebases (Go, Elixir, TypeScript)
- Validate full workflow: research → plan → implement → verify
- Performance benchmarking

### Planned Tests (tests/e2e/integration/)
- [ ] `TestGoProjectWorkflow` - Full Go project
- [ ] `TestElixirProjectWorkflow` - Full Elixir project
- [ ] `TestTypeScriptProjectWorkflow` - Full TS project

---

## Running Tests

```bash
# Run all tests
gptcode test e2e

# Run specific category
gptcode test e2e run
gptcode test e2e chat
gptcode test e2e integration

# With notifications
gptcode test e2e --notify

# Custom timeout (for slow local models)
gptcode test e2e --timeout 900  # 15 minutes
```

---

## Success Criteria

### Phase 1 (✅ Complete)
- ✅ Test infrastructure working
- ✅ Profile-based execution
- ✅ Progress tracking and notifications
- ✅ Real gptcode command execution
- ✅ File validation

### Phase 2-5 (Pending)
- Test coverage for all major commands
- 90%+ pass rate with local Ollama
- < 15 min execution time for full suite
- Automated CI integration

---

## Migration Notes

### Bash Scripts → Go Tests
**Deprecated:**
- `tests/e2e/scenarios/*.sh` - Old bash-based tests
- `tests/e2e.sh` - Old runner script

**Replaced by:**
- `tests/e2e/<category>/*_test.go` - Go test files
- `gptcode test e2e` - New test runner

**Advantages:**
- Type-safe test code
- Better IDE support
- Standard Go test tooling
- Real-time progress tracking
- Better error messages
