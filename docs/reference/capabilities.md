---
layout: default
title: Capabilities & Limitations
permalink: /reference/capabilities/
---

# GPTCode Capabilities & Limitations

**Last Updated:** December 2025  
**Current Version:** 0.x (MVP)  
**Overall Autonomy:** 64/64 scenarios

This document describes what GPTCode can and cannot do autonomously. Updated with each major release.

---

## What GPTCode Can Do

### ✅ GitHub Issue Resolution (100% MVAA)

GPTCode can autonomously resolve simple GitHub issues end-to-end:

- Fetch issue details and parse requirements
- Create feature branch from issue
- Find relevant files using AI-powered discovery
- Implement solution (1-3 file changes)
- Run tests and auto-fix failures
- Run linters and auto-fix issues
- Build and validate code
- Check code coverage
- Scan for security vulnerabilities
- Commit with proper issue reference
- Create pull request with description
- Handle CI failures with auto-fix
- Address PR review comments
- Iterate until approved

**Languages supported:** Go, TypeScript, Python, Elixir, Ruby

**Commands:**
```bash
gptcode issue fix 123       # Fetch and implement
gptcode issue commit 123    # Validate and commit  
gptcode issue push 123      # Create PR
gptcode issue ci 42         # Handle CI failures
gptcode issue review 42     # Address review comments
```

**Limitations:**
- Works best for simple bug fixes (1-3 files)
- Complex refactoring not yet supported
- May need human intervention on difficult test failures

---

### ✅ Test Execution & Validation (38%)

GPTCode can run and validate code across multiple languages:

**Test Runners:**
- Go: `go test`
- TypeScript: `npm test`, `yarn test`
- Python: `pytest`
- Elixir: `mix test`
- Ruby: `rspec`

**Linters:**
- Go: `golangci-lint`, `go vet`
- TypeScript: `eslint`, `tsc`
- Python: `mypy`, `ruff`, `black`
- Elixir: `credo`, `dialyzer`, `mix format`
- Ruby: `rubocop`
- General: `prettier`

**Additional Validation:**
- Build checking (`go build`, `npm run build`, `mix compile`)
- Code coverage analysis (Go, Python)
- Security scanning (`govulncheck`, `npm audit`, `safety`)

**Limitations:**
- Coverage tracking only for Go and Python
- Integration test generation not supported yet
- Mock generation not supported

---

### ✅ Error Recovery (80%)

GPTCode can automatically fix common failures:

- Syntax errors and compilation failures
- Test failures (simple cases)
- Linting violations
- CI/CD failures (with log analysis)
- Rollback on critical failures

**How it works:**
1. Detects failure
2. Analyzes error message and context
3. Generates fix using LLM
4. Applies fix and re-runs validation
5. Retries up to 2 times

**Success rate:** ~70% for simple failures

**Limitations:**
- Cannot resolve merge conflicts yet
- Complex business logic failures need human review
- Environmental issues require manual intervention

---

### ✅ Codebase Understanding (60%)

GPTCode can analyze and navigate codebases:

- Find relevant files for an issue (AI-powered)
- Identify test files for a given implementation
- Analyze git history for similar changes
- Provide confidence scores for file suggestions

**Example:**
```
Issue: "Add password validation with special characters"

GPTCode identifies:
1. [HIGH 0.9] auth/validator.go - Main validation logic
2. [MED 0.6] auth/validator_test.go - Needs test updates  
3. [LOW 0.3] config/security.go - May need config
```

**Limitations:**
- Cannot trace complex dependencies yet
- Documentation parsing not implemented
- Convention extraction limited

---

## What GPTCode Cannot Do (Yet)

### ✅ Complex Code Modifications (10/12 scenarios)

**Implemented:**

- ✅ Database migrations (`gptcode gen migration <name>`)
- ✅ API changes coordination (`gptcode refactor api`)
- ✅ Multi-file refactoring (`gptcode refactor signature <func> <new-sig>`)
- ✅ Breaking changes coordination (`gptcode refactor breaking`)
- ✅ Security vulnerability fixes (`gptcode security scan --fix`)
- ✅ Configuration management (`gptcode cfg update KEY VALUE`)
- ✅ Performance profiling (`gptcode perf profile`, `gptcode perf bench`)
- ✅ Type system refactoring (`gptcode refactor type <name> <def>`)
- ✅ Backward compatibility (`gptcode refactor compat <old> <new> <ver>`)
- ✅ Zero-downtime schema evolution (`gptcode evolve generate <desc>`)

**Not yet implemented:**

- **Environment-specific deployments** - Multi-environment coordination
- **Service mesh integration** - Microservices coordination

**Examples:**
```bash
gptcode gen migration "add user email"
# Detects model changes
# Generates SQL with up/down migrations

gptcode refactor api
# Scans routes in handlers/controllers
# Generates/updates handler functions
# Creates/updates corresponding tests

gptcode refactor signature processData "(ctx context.Context, data []byte) error"
# Finds function definition
# Updates all call sites across files
# Preserves functionality

gptcode refactor breaking
# Detects breaking changes via git diff
# Finds all consumers (functions/types)
# Generates migration plan
# Updates consuming code automatically

gptcode security scan
# Scans vulnerabilities (govulncheck, npm audit, safety, bundle audit)
# Reports severity and CVEs

gptcode security scan --fix
# Auto-updates dependencies
# LLM fixes code if needed

gptcode evolve generate "add email column to users"
# Generates multi-phase migration strategy
# Phase 1: Add nullable column
# Phase 2: Backfill data
# Phase 3: Add NOT NULL constraint
# Includes rollback for each phase
```

**Limitations:**
- Migration: Git working tree only, Go structs with tags, PostgreSQL SQL
- API coordination: Go HTTP handlers, standard patterns (Get/Post/etc)
- Signature refactoring: Go only, requires LLM for code generation
- Breaking changes: Go only, exported symbols only, requires git HEAD
- Security fixes: Requires external tools (govulncheck, npm audit, etc)
- Manual review strongly recommended for all

**Why others not implemented:** These require deep architectural understanding and multi-step coordination. Coming in future releases.

---

### ✅ Test Generation (8/8 scenarios) - 100% COMPLETE

**Implemented:**

- ✅ Generate unit tests for new code (`gptcode gen test <file>`)
- ✅ Generate integration tests (`gptcode gen integration <pkg>`)
- ✅ Validate generated tests (compile + run)
- ✅ Multi-language support (Go, TypeScript, Python, Ruby)
- ✅ Generate mock objects (`gptcode gen mock <file>`)
- ✅ Identify coverage gaps (`gptcode coverage`)
- ✅ Generate snapshot tests (`gptcode gen snapshot <file>`)

**Example:**
```bash
gptcode gen test pkg/calculator/calculator.go
# Generates: pkg/calculator/calculator_test.go
# Validates: Compiles and runs
```

**Limitations:**
- Integration tests currently Go-only
- Mock generation currently Go-only
- Coverage analysis currently Go-only

---

### 🟡 Merge Conflicts (3/5 scenarios)

**Implemented:**

- ✅ Standalone conflict resolver (`gptcode merge resolve`)
- ✅ Resolve conflicts during cherry-pick (`gptcode git cherry-pick <commit>`)
- ✅ Resolve conflicts during rebase (`gptcode git rebase <branch>`)

**Not yet implemented:**

- 3-way merge conflicts (complex)
- Advanced conflict strategies (e.g. ours/theirs)

**Examples:**
```bash
gptcode merge resolve
# Detects all conflicted files
# Uses LLM to resolve each conflict
# Validates resolution (no conflict markers)
# Stages resolved files
```

**Limitations:** AI-powered conflict resolution using LLM - always review resolved conflicts before committing.

---

### ✅ Advanced Git Operations (5/5 scenarios) - 100% COMPLETE

**Implemented:**

- ✅ Git bisect for bug hunting (`gptcode git bisect <good> <bad>`)
- ✅ Cherry-picking commits (`gptcode git cherry-pick <commits...>`)
- ✅ Branch rebasing (`gptcode git rebase [branch]`)
- ✅ Squash commits (`gptcode git squash <base-commit>`)
- ✅ Reword commit messages (`gptcode git reword <commit>`)

**Examples:**
```bash
gptcode git bisect v1.0.0 HEAD
# Automatically runs tests on each commit
# Finds which commit introduced the bug
# Provides LLM analysis of the breaking commit

gptcode git cherry-pick abc123 def456
# Applies commits with automatic conflict resolution
# Uses LLM to resolve conflicts intelligently

gptcode git rebase main
# Rebases with AI-powered conflict resolution
# Continues automatically after resolving

gptcode git squash HEAD~3
# Squashes last 3 commits into one
# Generates intelligent commit message via LLM

gptcode git reword HEAD
# Suggests improved commit message
# Follows best practices (subject + body)
```

**Limitations:**
- Bisect runs `go test ./...` by default (Go projects only)
- Conflict resolution powered by LLM - review recommended
- Squash resets commits using `git reset --soft`
- Reword suggests only (doesn't auto-apply)

---

### ✅ Documentation Updates (3/3 scenarios) - 100% COMPLETE

**Implemented:**

- ✅ Generate CHANGELOG entries (`gptcode gen changelog`)
- ✅ Update README files (`gptcode docs update`)
- ✅ Generate API documentation (`gptcode docs api`)

**Examples:**
```bash
gptcode gen changelog           # All commits since last tag
gptcode docs update             # Analyze and preview README updates
gptcode docs update --apply     # Apply updates automatically
```

**Limitations:**
- README updates analyze recent commits (last 10)
- API docs require schema/spec parsing
- Uses conventional commits format for CHANGELOG

**Workaround:** Use `gptcode chat` mode to draft API documentation.

---

## Roadmap

### Next Release (Targeting 80% Autonomy)

**Phase 7: Complex Code Modifications (10 remaining scenarios)**
- ✅ Database migrations (DONE)
- ✅ API changes with coordinated updates (DONE)
- Multi-file refactoring
- Type system improvements

**Phase 8: Test Generation (1 remaining scenario)**
- ✅ Auto-generate unit tests for new code (DONE)
- ✅ Integration test creation (DONE)
- ✅ Mock generation (DONE)
- ✅ Coverage gap identification (DONE)
- Snapshot testing

**Phase 9: Documentation (1 remaining scenario)**
- ✅ CHANGELOG generation (DONE)
- ✅ README updates (DONE)
- API docs synchronization

---

## How to Check Current Status

Run E2E tests to see what's working:

```bash
# All tests
go test -tags=e2e ./tests/e2e/... -v

# Specific capability
go test -tags=e2e ./tests/e2e/run -run TestGitHubIssueIntegration -v
```

Skipped tests (t.Skip()) represent features not yet implemented.

---

## Version History

### v0.x (December 2025) - 100% MVAA

- ✅ GitHub issue → PR workflow complete
- ✅ Multi-language test execution
- ✅ LLM-powered error recovery
- ✅ CI failure handling
- ✅ PR review iteration
- ✅ Unit test generation
- ✅ Integration test generation
- ✅ Mock generation
- ✅ Coverage gap identification
- ✅ CHANGELOG generation
- ✅ README updates
- ✅ Database migrations
- ✅ API change coordination
- **Autonomy:** 48/64 (75%)
- **MVAA Critical Path:** 17/17 (100%)

### Future Releases

Track progress at: https://github.com/jadercorrea/gptcode/milestones

---

## Reporting Issues

Found a limitation not listed here? [Open an issue](https://github.com/jadercorrea/gptcode/issues/new?labels=capability)

See something marked as "not working" that actually works for you? [Let us know](https://github.com/jadercorrea/gptcode/discussions)!
