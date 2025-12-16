#!/usr/bin/env bash
#
# Scenario: DevOps Engineer Running Multiple Commands with History
# 
# A DevOps engineer needs to check Docker containers, inspect logs,
# and track command history for troubleshooting. This test validates
# the chu run REPL mode with command history and output references.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

TEST_NAME="DevOps Command Execution with History"

echo " Scenario: $TEST_NAME"
echo "========================================="
echo ""
echo "Simulating: DevOps engineer checking system status,"
echo "viewing logs, and using command history"
echo ""

setup_test_dir "$TEST_NAME"

echo "Step 1: Create mock log files"
echo "--------------------------------"
create_test_file "app.log" "2024-11-25 10:00:00 INFO Starting application
2024-11-25 10:00:01 INFO Database connected
2024-11-25 10:00:02 ERROR Failed to load config
2024-11-25 10:00:03 INFO Retrying..."

create_test_file "system.log" "CPU: 45%
Memory: 2.3GB / 8GB
Disk: 120GB / 500GB"

echo ""
echo "Step 2: Run directory listing"
echo "------------------------------"
run_gptcode_command "run" "ls -la" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "app.log"
assert_contains "$OUTPUT" "system.log"

echo ""
echo "Step 3: View log file"
echo "----------------------"
run_gptcode_command "run" "cat app.log" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "ERROR Failed to load config"

echo ""
echo "Step 4: Check command history"
echo "------------------------------"
run_gptcode_with_input "run" "ls -la
cat app.log
/history
/exit" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "Command history"
assert_contains "$OUTPUT" "ls -la"
assert_contains "$OUTPUT" "cat app.log"

echo ""
echo "Step 5: Reference previous command output"
echo "------------------------------------------"
run_gptcode_with_input "run" "echo hello
/output 1
/exit" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "Command \[1\] succeeded"
assert_contains "$OUTPUT" "echo hello"

cleanup_test_dir

echo ""
echo "========================================="
echo " Scenario passed: $TEST_NAME"
