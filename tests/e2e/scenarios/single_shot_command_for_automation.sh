#!/usr/bin/env bash
#
# Scenario: Single-Shot Command Execution for Automation
#
# Validates that chu run can be used in scripts and CI/CD pipelines
# with single-shot execution (no REPL), maintaining backwards compatibility.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

TEST_NAME="Single-Shot Command for Automation"

echo " Scenario: $TEST_NAME"
echo "========================================="
echo ""
echo "Simulating: CI/CD pipeline running automated commands"
echo "with chu run in non-interactive mode"
echo ""

setup_test_dir "$TEST_NAME"

echo "Step 1: Execute command and exit immediately"
echo "---------------------------------------------"
run_gptcode_command "run" "echo 'Build started'" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "Build started"

echo ""
echo "Step 2: List files in single-shot mode"
echo "---------------------------------------"
create_test_file "README.md" "# Project"
mkdir -p src
create_test_file "src/main.go" "package main"
run_gptcode_command "run" "ls -R" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "README.md"
assert_contains "$OUTPUT" "main.go"

echo ""
echo "Step 3: Piped input for automated scripts"
echo "------------------------------------------"
run_gptcode_with_input "run" "test automation" "--once"
assert_exit_code 0

echo ""
echo "Step 4: Command with arguments"
echo "-------------------------------"
run_gptcode_command "run" "cat README.md" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "# Project"

echo ""
echo "Step 5: Verify no REPL banner in single-shot"
echo "---------------------------------------------"
run_gptcode_command "run" "pwd" "--raw"
assert_exit_code 0
assert_not_contains "$OUTPUT" "GPTCode Run REPL"

cleanup_test_dir

echo ""
echo "========================================="
echo " Scenario passed: $TEST_NAME"
