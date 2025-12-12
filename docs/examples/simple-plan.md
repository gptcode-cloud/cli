# Simple Plan Example

This is a minimal plan to demonstrate `gptcode implement` with interactive and autonomous modes.

## Add utility function

Create a new file `utils/math.go` with a simple Add function.

**Requirements:**
- Package name: `utils`
- Function: `Add(a, b int) int`
- Returns: sum of a and b

**Files:**
- Create: `utils/math.go`

## Add tests

Create tests for the Add function.

**Requirements:**
- Package name: `utils`
- Test file: `utils/math_test.go`
- Test cases: positive numbers, zero, negative numbers

**Files:**
- Create: `utils/math_test.go`

## Usage

**Interactive mode (default):**
```bash
gptcode implement docs/examples/simple-plan.md
```
- Prompts before each step
- Shows step content
- Options: Y (execute), n (skip), q (quit)
- On error: prompts to continue or stop

**Autonomous mode:**
```bash
gptcode implement docs/examples/simple-plan.md --auto
```
- Executes all steps automatically
- Verifies with build + tests after each step
- Retries on errors (up to 3 times by default)
- Creates checkpoints
- Rollback on failure

**With lint:**
```bash
gptcode implement docs/examples/simple-plan.md --auto --lint
```

**Resume from checkpoint:**
```bash
gptcode implement docs/examples/simple-plan.md --auto --resume
```
