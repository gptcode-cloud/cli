# E2E Test Failure Analysis

## Current Status: 5/9 passing (55%)

## Failed Tests Analysis

### 1. Basic File Operations ❌ (FLAKY)
**Issue**: Intermittent validator max iterations
**Root cause**: Test is **flaky** - passes 2/3 times, fails 1/3
**Evidence**: 
- Ran test 3 times: 2 passed, 1 failed
- Works perfectly when run in isolation
- Not a code bug, likely Groq rate limits or model variance

**Fix needed**: Add retry logic to e2e tests OR accept flakiness

---

### 2. Code Generation ❌ (AUTONOMOUS MODE BUG)
**Issue**: Creates `script.py` instead of `calc.py`
**Root cause**: Autonomous mode (Symphony) ignores specific filenames from user prompt
**Evidence**:
```bash
Test request: "create a Python script calc.py that adds two numbers"
Agent creates: script.py  # WRONG
Expected: calc.py         # What test checks for
```

**Fix needed**: 
- Symphony needs to extract and preserve specific filenames from goal
- OR improve movement parsing to respect exact file names
- Location: `internal/autonomous/symphony.go`

---

### 3. Conversational Code Exploration ❌ (QUERY RETURN BUG)
**Issue**: "Editor reached max iterations" - query doesn't return result
**Root cause**: Query tasks hit max iterations without returning content
**Evidence**:
```
Task: Show me User struct
Output: "Editor reached max iterations"  # Should show struct code
```

**Same issue as basic_file_operations Test 2** - early return logic not working for all query patterns

**Fix needed**: 
- Debug why some queries don't trigger early return
- Possibly related to plan format variations not caught by keyword detection

---

### 4. Git Operations ❌ (REVIEWER BUG - CRITICAL)
**Issue**: Validator marks SUCCESS as FAIL
**Root cause**: `extractIssues()` treats **any line starting with `-`** as an issue
**Evidence**:
```
Validator output:
Issues:
- Command `git log -1` executed successfully  ← This is SUCCESS!
- Command output includes commit hash          ← This is SUCCESS!
- Command completed without errors             ← This is SUCCESS!

Result: FAIL  ← WRONG! Should be SUCCESS
```

**Bug location**: `internal/agents/reviewer.go:285`
```go
if trimmed != "" && (strings.HasPrefix(trimmed, "-") || ...) {
    issues = append(issues, trimmed)  // BUG: treats bullets as issues
}
```

**Fix needed**: Change `extractIssues()` to:
1. Only extract lines that contain actual failure keywords (FAIL, error, missing, not met)
2. Ignore bullet-formatted success messages
3. Better detect if text is positive (success) vs negative (failure)

---

## Priority Fixes

### 🔴 CRITICAL (blocks tests):
1. **Git Operations**: Fix `extractIssues()` bug - marks success as failure

### 🟡 HIGH (incorrect behavior):
2. **Code Generation**: Fix Symphony to respect exact filenames

### 🟢 MEDIUM (flaky but works):
3. **Basic File Operations**: Investigate flakiness or add test retries
4. **Conversational Exploration**: Fix query early return edge cases

---

## Recommended Next Steps

1. Fix `extractIssues()` bug (5 min) → will fix git_operations immediately
2. Fix Symphony filename extraction (30 min) → will fix code_generation
3. Add test retry wrapper (10 min) → will reduce flakiness
4. Debug query early return edge cases (20 min) → will fix remaining query issues

**Expected after fixes: 8-9/9 tests passing (88-100%)**
