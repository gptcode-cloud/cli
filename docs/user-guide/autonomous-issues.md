# Autonomous GitHub Issue Resolution

GPTCode can now autonomously resolve GitHub issues from start to finish, including iterating on CI failures and review comments until the PR is approved.

**Status: MVP Complete (100% MVAA Coverage)**  
**Supported Languages:** Go, TypeScript, Python, Elixir, Ruby

## Quick Start

```bash
# 1. Fix an issue
gptcode issue fix 123

# 2. Validate and commit
gptcode issue commit 123 --auto-fix --check-coverage --security-scan

# 3. Create pull request
gptcode issue push 123

# 4. Handle CI failures (if any)
gptcode issue ci 42

# 5. Address review comments
gptcode issue review 42

# 6. Repeat steps 4-5 until approved!
```

## Complete Workflow

### Step 1: Fix Issue

```bash
gptcode issue fix 123 [--repo owner/repo] [--autonomous] [--find-files]
```

**What it does:**
1. Fetches issue #123 from GitHub
2. Extracts requirements from issue body
3. Creates branch `issue-123-description`
4. Finds relevant files (AI-powered with confidence scoring)
5. Implements solution using Symphony autonomous executor
6. Shows next steps (commit, push)

**Options:**
- `--repo owner/repo` - Specify repository (auto-detected from git remote)
- `--autonomous` - Use autonomous executor (default: true)
- `--find-files` - Find relevant files before implementation (default: true)

**Example output:**
```
🔍 Fetching issue #123 from owner/repo...

📋 Issue #123: Add password validation
   State: open
   Author: reviewer
   Labels: enhancement

📝 Requirements:
   1. Add minimum 8 characters validation
   2. Require at least one special character

🌿 Creating branch: issue-123-add-password-validation

🔍 Finding relevant files...

Relevant files identified:
1. [HIGH] auth/validator.go - Contains validation logic
2. [MED] auth/validator_test.go - Test file
3. [LOW] config/security.go - Security config

[Implementation via Symphony...]

[OK] Implementation complete

Next steps:
   gptcode issue commit 123
   gptcode issue push 123
```

### Step 2: Commit with Validation

```bash
gptcode issue commit 123 [options]
```

**What it does:**
1. Runs comprehensive validation pipeline
2. Auto-fixes failures if enabled
3. Commits changes with proper issue reference ("Closes #123")

**Options:**
- `--message` - Custom commit message
- `--skip-tests` - Skip test execution
- `--skip-lint` - Skip linting
- `--skip-build` - Skip build check
- `--auto-fix` - Auto-fix test/lint failures (default: true)
- `--check-coverage` - Check code coverage
- `--min-coverage N` - Minimum coverage threshold (0-100)
- `--security-scan` - Run security vulnerability scan

**Validation Pipeline:**
1. **Build Check** - Compiles code (Go, TypeScript, Elixir)
2. **Tests** - Runs language-specific test suite
   - Go: `go test ./...`
   - TypeScript: `npm test`
   - Python: `pytest`
   - Elixir: `mix test`
   - Ruby: `rspec`
3. **Linting** - Runs multiple linters
   - Go: golangci-lint
   - TypeScript: eslint, tsc
   - Python: pytest, mypy, ruff, black
   - Elixir: credo, mix format, dialyzer
   - Ruby: rubocop
4. **Coverage** - Analyzes test coverage (if enabled)
5. **Security** - Scans for vulnerabilities (if enabled)
   - Go: govulncheck
   - Node: npm audit
   - Python: safety

**Auto-fix:**
When tests or linters fail, GPTCode uses LLM to:
1. Analyze the failure
2. Generate a fix
3. Apply the fix
4. Re-run validation
5. Retry up to 2 times

**Example output:**
```
💾 Committing changes for issue #123...

🔨 Running build...
✅ Build successful

🧪 Running tests...
❌ Tests failed (2 passed, 1 failed)

🔧 Attempting auto-fix...
✅ Tests fixed automatically

🔍 Running linters...
✅ All linters passed (golangci-lint: 0 issues)

📊 Checking code coverage...
✅ Coverage: 85.2% (threshold: 80%)

🔒 Running security scan...
✅ No vulnerabilities found

✅ Changes committed

✨ All validation passed!
Next steps:
  gptcode issue push 123
```

### Step 3: Push and Create PR

```bash
gptcode issue push 123 [--repo owner/repo] [--draft]
```

**What it does:**
1. Pushes branch to origin
2. Creates pull request via `gh` CLI
3. Links PR to issue with "Closes #123"
4. Copies labels from issue to PR
5. Shows PR URL

**Options:**
- `--repo owner/repo` - Specify repository
- `--draft` - Create draft pull request

**Example output:**
```
🚀 Pushing issue-123-add-password-validation...
✅ Branch pushed

📝 Creating pull request...
✅ Pull request created: https://github.com/owner/repo/pull/42
   PR #42: Fix: Add password validation
```

### Step 4: Handle CI Failures

```bash
gptcode issue ci 42 [--repo owner/repo]
```

**What it does:**
1. Monitors CI checks via `gh pr checks`
2. Fetches logs from failed checks
3. Parses and extracts error information
4. Analyzes failure using LLM
5. Generates and applies fix
6. Commits and pushes automatically
7. CI reruns automatically

**Common CI issues handled:**
- Test failures
- Build errors
- Linting issues
- Dependency problems
- Environment issues

**Example output:**
```
🔍 Checking CI status for PR #42...

❌ Found 2 failed check(s):

1. Tests - fail
2. Lint - fail

📜 Fetching CI logs...
🔎 Analyzing failures...

Detected error: TestPasswordValidation failed

🔧 Generating fix...
✅ Fix generated

Recommended changes:
## Analysis
Test expects special character validation but it's not implemented

## Solution
Add special character check in validator

## Changes
- auth/validator.go: Add regexp check for special chars

📦 Committing fix...
✅ Changes committed
🚀 Pushing...

✅ CI fix pushed
   View PR: https://github.com/owner/repo/pull/42

⏳ CI checks will run again automatically
```

### Step 5: Address Review Comments

```bash
gptcode issue review 42 [--repo owner/repo]
```

**What it does:**
1. Fetches unresolved review comments via GitHub API
2. Processes each comment autonomously
3. Implements requested changes
4. Commits with reference to PR
5. Pushes updates

**Example output:**
```
🔍 Fetching review comments for PR #42...

📝 Found 3 unresolved comment(s):

1. [@reviewer] auth/validator.go:45
   Please add error handling here

2. [@reviewer] auth/validator_test.go:23
   Add test for empty password case

3. [@reviewer] README.md:12
   Update documentation with new validation rules

🔧 Addressing review comments...

[1/3] Processing comment from @reviewer on auth/validator.go
[Implementation via Symphony...]
✅ Comment addressed

[2/3] Processing comment from @reviewer on auth/validator_test.go
[Implementation via Symphony...]
✅ Comment addressed

[3/3] Processing comment from @reviewer on README.md
[Implementation via Symphony...]
✅ Comment addressed

📦 Committing changes...
✅ Changes committed
🚀 Pushing...

✨ Successfully addressed 3 review comment(s)
   View PR: https://github.com/owner/repo/pull/42
```

### Step 6: Iterate Until Approved

Repeat steps 4-5 as needed:

```bash
# Check if new CI failures appeared
gptcode issue ci 42

# Check if new review comments appeared
gptcode issue review 42

# Repeat until PR is approved and merged!
```

## Advanced Usage

### Custom Commit Messages

```bash
gptcode issue commit 123 --message "feat: Add password validation with special char requirement"
```

### Strict Coverage Requirements

```bash
gptcode issue commit 123 --check-coverage --min-coverage 90
```

### Security-First Validation

```bash
gptcode issue commit 123 --security-scan --skip-tests
```

### Manual Steps Only

```bash
gptcode issue fix 123 --autonomous=false    # Show what needs to be done
gptcode issue commit 123 --skip-tests --skip-lint  # Only commit
```

## Architecture

### Modules

**GitHub Integration** (`internal/github/`)
- Issue fetching and parsing
- PR creation and management
- Review comment handling
- Commit operations

**Codebase Understanding** (`internal/codebase/`)
- AI-powered file discovery
- Confidence scoring (HIGH/MED/LOW)
- Test file identification
- Git history analysis

**Validation** (`internal/validation/`)
- Multi-language test execution
- Comprehensive linting (12 tools)
- Build checking
- Coverage analysis
- Security scanning

**Error Recovery** (`internal/recovery/`)
- LLM-powered error analysis
- Auto-fix generation
- Retry strategies
- Rollback on critical failure

**CI Handling** (`internal/ci/`)
- Status monitoring
- Log fetching and parsing
- Failure analysis
- Auto-fix generation

### Supported Tools

**Test Runners:**
- Go: `go test`
- TypeScript: `npm test`, `jest`
- Python: `pytest`
- Elixir: `mix test`
- Ruby: `rspec`

**Linters:**
- Go: `golangci-lint`, `go vet`
- TypeScript: `eslint`, `tsc`, `prettier`
- Python: `pytest`, `mypy`, `ruff`, `black`
- Elixir: `credo`, `mix format`, `dialyzer`
- Ruby: `rubocop`

**Build Tools:**
- Go: `go build`
- TypeScript: `npm run build`, `tsc`
- Elixir: `mix compile`

**Security Scanners:**
- Go: `govulncheck`
- Node: `npm audit`
- Python: `safety`

## Configuration

### GitHub CLI Setup

```bash
# Install gh CLI
brew install gh  # macOS
# or apt install gh  # Linux

# Authenticate
gh auth login

# Verify
gh auth status
```

### Repository Detection

GPTCode auto-detects the repository from your git remote:

```bash
git remote get-url origin
# → https://github.com/owner/repo.git
```

Or specify explicitly:

```bash
gptcode issue fix 123 --repo owner/repo
```

## Limitations

### Current Constraints
- Single-file or localized changes work best
- Complex multi-file refactorings may need manual review
- Some CI platforms not supported (GitHub Actions works best)
- Merge conflicts must be resolved manually

### What's Not Autonomous (Yet)
- Merging the PR (can be done manually or via GitHub settings)
- Resolving merge conflicts with main branch
- Handling complex architectural changes
- Understanding implicit business logic requirements

## Success Metrics

**MVAA (Minimum Viable Autonomous Agent) Coverage: 17/17 (100%)**

The autonomous issue workflow can handle:
- ✅ Simple bug fixes (1-3 files)
- ✅ Small feature additions
- ✅ Test coverage improvements
- ✅ Linting/formatting fixes
- ✅ Documentation updates
- ✅ Dependency updates
- ✅ Security patches

**Overall Autonomy Coverage: 38/64 scenarios (59%)**

## Troubleshooting

### "Could not detect GitHub repository"
```bash
# Check git remote
git remote -v

# Add if missing
git remote add origin https://github.com/owner/repo.git

# Or specify explicitly
gptcode issue fix 123 --repo owner/repo
```

### "gh: command not found"
```bash
# Install GitHub CLI
brew install gh      # macOS
apt install gh       # Linux
choco install gh     # Windows

# Authenticate
gh auth login
```

### "Failed to fetch issue"
```bash
# Check authentication
gh auth status

# Re-authenticate if needed
gh auth login

# Verify access to repository
gh repo view owner/repo
```

### CI Failures Not Detected
```bash
# Check if CI is running
gh pr checks 42

# Wait for CI to complete
# Then run:
gptcode issue ci 42
```

### Review Comments Not Found
```bash
# Verify there are unresolved comments
gh pr view 42

# Check API access
gh api /repos/owner/repo/pulls/42/comments
```

## Next Steps

After achieving 100% MVAA:

1. **Test it!** Try on real issues in your repositories
2. **Report bugs** via GitHub issues
3. **Contribute** improvements (see CONTRIBUTING.md)
4. **Share feedback** on what works and what doesn't

## Related Documentation

- [Workflow Guide](workflow-guide.md) - General usage patterns
- [Autonomous Mode Deep Dive](commands.md#autonomous-mode) - Technical details
- [Gap Analysis](AUTONOMY_GAP_ANALYSIS.md) - Implementation progress
- [Contributing](../CONTRIBUTING.md) - How to contribute

## Changelog

**2025-12-06** - MVP Complete (100% MVAA)
- Added `gptcode issue ci` for CI failure handling
- Added `gptcode issue review` for review comment handling
- Achieved 100% MVAA Critical Path coverage
- Overall autonomy: 59% (38/64 scenarios)
