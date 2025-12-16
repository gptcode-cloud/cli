#!/usr/bin/env bash
#
# Scenario: Git Operations
#
# Validates core git workflow capabilities:
# - Status and diff
# - Log and history
# - Branch operations
# - Commit and staging

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/.."
source "$SCRIPT_DIR/lib/helpers.sh"

TEST_NAME="Git Operations"

echo " Scenario: $TEST_NAME"
echo "========================================="
echo ""
echo "Testing: Essential git workflow commands"
echo ""

setup_test_dir "$TEST_NAME"

# Initialize git repo
git init > /dev/null 2>&1
git config user.email "test@chu.test"
git config user.name "Chu Test"
echo "# Test Project" > README.md
git add README.md
git commit -m "Initial commit" > /dev/null 2>&1

echo "Test 1: Show repository status"
echo "-------------------------------"
run_gptcode_command "do" "show me the git status"
assert_exit_code 0

echo ""
echo "Test 2: View commit history"
echo "---------------------------"
run_gptcode_command "do" "show me the recent commits"
assert_exit_code 0
assert_contains "$OUTPUT" "Initial commit"

echo ""
echo "Test 3: Show changes (diff)"
echo "---------------------------"
echo "New content" >> README.md
run_gptcode_command "do" "what changes did I make?"
assert_exit_code 0

echo ""
echo "Test 4: List branches"
echo "---------------------"
run_gptcode_command "do" "show me all git branches"
assert_exit_code 0
assert_contains "$OUTPUT" "master\|main"

echo ""
echo "Test 5: Create new file and check status"
echo "-----------------------------------------"
echo "console.log('test')" > test.js
run_gptcode_command "do" "show me what files are not tracked by git"
assert_exit_code 0
assert_contains "$OUTPUT" "test.js"

cleanup_test_dir

echo ""
echo "========================================="
echo " Scenario passed: $TEST_NAME"
