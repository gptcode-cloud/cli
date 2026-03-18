package maestro

import (
	"fmt"
	"strings"
)

// RecoveryStrategy handles error recovery during execution
type RecoveryStrategy struct {
	MaxRetries  int
	Checkpoints *CheckpointSystem
	Verbose     bool // Enable detailed logging
}

// RecoveryContext provides context for error recovery
type RecoveryContext struct {
	ErrorType     ErrorType
	ErrorOutput   string
	ModifiedFiles []string
	StepIndex     int
	Attempts      int
	MaxAttempts   int
}

// NewRecoveryStrategy creates a new recovery strategy
func NewRecoveryStrategy(maxRetries int, checkpoints *CheckpointSystem) *RecoveryStrategy {
	return &RecoveryStrategy{
		MaxRetries:  maxRetries,
		Checkpoints: checkpoints,
	}
}

// Rollback reverts changes to a specific checkpoint
func (rs *RecoveryStrategy) Rollback(checkpointID string) error {
	if rs.Checkpoints == nil {
		return fmt.Errorf("checkpoint system not initialized")
	}
	return rs.Checkpoints.Restore(checkpointID)
}

// ErrorType represents the category of an error
type ErrorType string

const (
	ErrorSyntax   ErrorType = "syntax"
	ErrorLogic    ErrorType = "logic"
	ErrorBuild    ErrorType = "build"
	ErrorTest     ErrorType = "test"
	ErrorSnapshot ErrorType = "snapshot"
	ErrorLint     ErrorType = "lint"
	ErrorUnknown  ErrorType = "unknown"
)

// ClassifyError attempts to categorize an error based on output
func ClassifyError(output string) ErrorType {
	output = strings.ToLower(output)

	// Syntax errors
	if strings.Contains(output, "syntax error") ||
		strings.Contains(output, "expected") ||
		strings.Contains(output, "undefined") ||
		strings.Contains(output, "unexpected") ||
		strings.Contains(output, "missing ") ||
		strings.Contains(output, "extra ") ||
		strings.Contains(output, "parse error") {
		return ErrorSyntax
	}

	// Build errors
	if strings.Contains(output, "build failed") ||
		strings.Contains(output, "cannot load") ||
		strings.Contains(output, "compile") ||
		strings.Contains(output, "redeclared") ||
		strings.Contains(output, "undefined:") ||
		strings.Contains(output, "import cycle") {
		return ErrorBuild
	}

	// Test errors
	if strings.Contains(output, "test failed") ||
		strings.Contains(output, "fail") ||
		strings.Contains(output, "assert") ||
		strings.Contains(output, "expected") && strings.Contains(output, "but got") {
		// Check for snapshot-specific errors first
		if strings.Contains(output, "snapshot") && strings.Contains(output, "failed") {
			return ErrorSnapshot
		}
		return ErrorTest
	}

	// Snapshot errors (may not have "test failed" keyword)
	if strings.Contains(output, "snapshots failed") ||
		strings.Contains(output, "snapshot") && strings.Contains(output, "updated") ||
		strings.Contains(output, "snapshots updated") {
		return ErrorSnapshot
	}

	// Lint errors
	if strings.Contains(output, "lint") ||
		strings.Contains(output, "golangci-lint") ||
		strings.Contains(output, "gofmt") ||
		strings.Contains(output, "ineffassign") ||
		strings.Contains(output, "vet") {
		return ErrorLint
	}

	return ErrorUnknown
}

func (rs *RecoveryStrategy) GenerateFixPrompt(errorType ErrorType, errorOutput string) string {
	ctx := &RecoveryContext{
		ErrorType:   errorType,
		ErrorOutput: errorOutput,
	}
	return rs.GenerateFixPromptWithContext(ctx)
}

// GenerateFixPromptWithContext generates a fix prompt with full context information
func (rs *RecoveryStrategy) GenerateFixPromptWithContext(ctx *RecoveryContext) string {
	var prompt string

	switch ctx.ErrorType {
	case ErrorSyntax:
		prompt = fmt.Sprintf(`The previous implementation has SYNTAX ERRORS.

Error output:
%s

Please fix the syntax errors. Focus on:
- Missing/extra parentheses, brackets, or braces
- Incorrect variable names or typos
- Missing imports or declarations
- Type mismatches
- Parse errors

Provide corrected code that compiles without syntax errors.`, ctx.ErrorOutput)

	case ErrorBuild:
		prompt = fmt.Sprintf(`The code failed to BUILD/COMPILE.

Error output:
%s

Please fix the build errors. Common issues:
- Missing dependencies (check imports)
- Type errors or incompatible types
- Undefined functions or variables
- Circular dependencies
- Variable redeclarations
- Import cycles

Provide corrected code that builds successfully.`, ctx.ErrorOutput)

	case ErrorTest:
		prompt = fmt.Sprintf(`The code builds but TESTS ARE FAILING.

Test output:
%s

Please fix the implementation to make tests pass. Common issues:
- Logic errors in implementation
- Edge cases not handled
- Incorrect return values
- Missing null/error checks
- Incorrect test expectations

Analyze the failing tests and correct the implementation.`, ctx.ErrorOutput)

	case ErrorSnapshot:
		snapshotPrompt := `SNAPSHOT TESTS FAILED - OUTPUT CHANGED

Test output:
%s

WHAT THIS MEANS:
The code implementation is correct, but the OUTPUT FORMAT has changed.
This is often EXPECTED when making formatting or behavior improvements.

ANALYSIS REQUIRED:
1. Look at the diff in the output:
   - "- Snapshot" = OLD expected value
   - "+ Received" = NEW actual value

2. Decide if the new output is:
   - CORRECT (better formatting, improved style) → Update snapshots
   - WRONG (bug introduced) → Fix the implementation

SPECIFIC GUIDANCE:
- If improving formatting/style: Run test update command (e.g., 'npm test -- -u')
- If introducing a bug: Fix the implementation to match expected output

EXAMPLES OF CORRECT SNAPSHOT UPDATES:
- Better line breaking in code formatters (Prettier, black)
- Improved error messages
- More consistent formatting
- Removal of unnecessary whitespace

EXAMPLES OF WRONG CHANGES (should NOT update):
- Breaking existing functionality
- Incorrect calculations
- Missing data in output
- Changed business logic`

		prompt = fmt.Sprintf(snapshotPrompt, ctx.ErrorOutput)

	case ErrorLint:
		prompt = fmt.Sprintf(`The code has LINT ERRORS.

Error output:
%s

Please fix the lint issues. Common issues:
- Code style violations
- Unused variables or imports
- Inefficient assignments
- Code smells
- Formatting issues

Provide corrected code that passes all lint checks.`, ctx.ErrorOutput)

	case ErrorLogic:
		prompt = fmt.Sprintf(`The code has LOGIC ERRORS.

Error output:
%s

Please review and fix the logic. Consider:
- Algorithm correctness
- Boundary conditions
- State management
- Data flow
- Edge cases

Provide corrected implementation with proper logic.`, ctx.ErrorOutput)

	default:
		prompt = fmt.Sprintf(`The previous implementation FAILED with the following error:

%s

Please analyze the error and provide a corrected implementation.`, ctx.ErrorOutput)
	}

	// Add context information to help with recovery
	if len(ctx.ModifiedFiles) > 0 {
		prompt += fmt.Sprintf(`

Context: The following files were modified in this step:
%s

When fixing the error, consider the changes made to these files.`,
			strings.Join(ctx.ModifiedFiles, "\n"))
	}

	if ctx.StepIndex >= 0 {
		prompt += fmt.Sprintf(`\n\nStep: This error occurred during step %d of the execution plan.`, ctx.StepIndex+1)
	}

	if ctx.Attempts > 0 {
		prompt += fmt.Sprintf(`\n\nRetry: This is attempt %d of %d. Previous attempts failed with the same or similar errors.`,
			ctx.Attempts, ctx.MaxAttempts)
	}

	return prompt
}

// AdvancedRecovery attempts to apply context-aware recovery strategies
func (rs *RecoveryStrategy) AdvancedRecovery(ctx *RecoveryContext) (string, bool) {
	// For specific error patterns, return targeted fix prompts
	output := strings.ToLower(ctx.ErrorOutput)

	// Check for snapshot failures - these often mean the new behavior is CORRECT
	// This check is done first because it should override other error types
	hasSnapshotKeywords := strings.Contains(output, "snapshot") &&
		(strings.Contains(output, "failed") || strings.Contains(output, "updated"))
	hasReceivedKeywords := strings.Contains(output, "+ received") || strings.Contains(output, "- snapshot")

	if hasSnapshotKeywords || hasReceivedKeywords {
		snapshotPrompt := `SNAPSHOT TEST FAILURE DETECTED

The test output shows that snapshot tests failed. This usually means the OUTPUT FORMAT changed, NOT that the code is wrong.

EXAMINING THE DIFF:
%s

CRITICAL DECISION:
Look at the diff above carefully:
- "- Snapshot" (or "- Expected") = OLD expected value
- "+ Received" = NEW actual value from your changes

DECIDE WHICH IS CORRECT:

IF THE NEW OUTPUT IS CORRECT (better formatting, improved style):
→ Update the snapshots by running the test update command
  Examples: npm test -- -u | npm test -- --update | npx jest -u

IF THE NEW OUTPUT IS WRONG (bug introduced):
→ Fix your implementation to produce the correct output

COMMON CASES FOR SNAPSHOT UPDATES (correct):
- Prettier/black formatting improvements
- Better error message formatting
- Improved line breaking
- More consistent style

COMMON CASES TO FIX (NOT snapshot updates):
- Changed business logic
- Incorrect calculations
- Missing data
- Breaking functionality

ANALYZE THE DIFF AND MAKE THE RIGHT CHOICE.`

		// Extract just the relevant diff portion for clarity
		diffOutput := extractSnapshotDiff(ctx.ErrorOutput)
		return fmt.Sprintf(snapshotPrompt, diffOutput), true
	}

	// Check for common Go build errors
	if ctx.ErrorType == ErrorBuild {
		if strings.Contains(output, "cannot find package") {
			return fmt.Sprintf(`The build failed because a package cannot be found.

Error output:
%s

You need to add the missing dependency to go.mod using 'go mod tidy' or add the import to your code.`, ctx.ErrorOutput), true
		}
		if strings.Contains(output, "undefined: ") {
			return fmt.Sprintf(`The build failed because a symbol is undefined.

Error output:
%s

You need to import the package that contains the undefined symbol or define the symbol in the current package.`, ctx.ErrorOutput), true
		}
		if strings.Contains(output, "redeclared in this block") {
			return fmt.Sprintf(`The build failed because a symbol is redeclared.

Error output:
%s

You need to remove or rename one of the redeclared symbols.`, ctx.ErrorOutput), true
		}
	}

	// Check for common test errors
	if ctx.ErrorType == ErrorTest {
		if strings.Contains(output, "expected") && strings.Contains(output, "got") {
			return fmt.Sprintf(`The test failed because the actual output did not match expected output.

Error output:
%s

Check the implementation against the test expectations and fix the return values or logic.`, ctx.ErrorOutput), true
		}
	}

	// Check for snapshot test failures - these often mean the new behavior is CORRECT
	if strings.Contains(output, "snapshot") && (strings.Contains(output, "failed") || strings.Contains(output, "updated")) {
		snapshotPrompt := `SNAPSHOT TEST FAILURE DETECTED

Error output:
%s

IMPORTANT ANALYSIS:
Snapshot test failures often mean the OUTPUT CHANGED, not that the code is wrong.
The new output may actually be CORRECT (better formatting, improved behavior).

DECISION REQUIRED:
1. If the new output is CORRECT (better formatting, improved behavior):
   → Update snapshots by running: npm test -- -u (Jest), npm test -- --update (Vitest), or similar
2. If the new output is WRONG (bug, regression):
   → Fix the implementation to match the expected output

Look at the diff in the error output:
- "- Snapshot" = old expected value
- "+ Received" = new actual value

Which is correct? Make the appropriate fix.`

		return fmt.Sprintf(snapshotPrompt, ctx.ErrorOutput), true
	}

	// Check for common lint errors
	if ctx.ErrorType == ErrorLint {
		if strings.Contains(output, "ineffassign") {
			return fmt.Sprintf(`The lint failed because of an ineffectual assignment.

Error output:
%s

Remove the assignment that has no effect.`, ctx.ErrorOutput), true
		}
		if strings.Contains(output, "unused") {
			return fmt.Sprintf(`The lint failed because of unused variables or imports.

Error output:
%s

Remove the unused variables or imports.`, ctx.ErrorOutput), true
		}
	}

	return "", false // No specific recovery strategy found
}

// extractSnapshotDiff extracts the relevant diff portion from snapshot test output
func extractSnapshotDiff(output string) string {
	lines := strings.Split(output, "\n")
	var diffLines []string
	inDiff := false
	maxLines := 50 // Limit to prevent too much output

	for i, line := range lines {
		// Start capturing when we see diff indicators
		if strings.Contains(line, "- Snapshot") ||
			strings.Contains(line, "+ Received") ||
			strings.Contains(line, "- Expected") ||
			strings.Contains(line, "+ Received") ||
			strings.Contains(line, "=======") {
			inDiff = true
		}

		if inDiff && len(diffLines) < maxLines {
			diffLines = append(diffLines, line)
		}

		// Stop after the output section
		if inDiff && strings.Contains(line, "Tests:") {
			break
		}

		// Safety limit
		if i > 200 {
			break
		}
	}

	if len(diffLines) == 0 {
		// Fallback: return first 500 chars of output
		if len(output) > 500 {
			return output[:500] + "\n... (output truncated)"
		}
		return output
	}

	return strings.Join(diffLines, "\n")
}
