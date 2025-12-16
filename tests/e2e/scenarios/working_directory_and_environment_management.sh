#!/usr/bin/env bash
#
# Scenario: Working Directory and Environment Management
#
# A developer working across multiple directories needs to manage
# environment variables and navigate between project folders.
# This validates /cd and /env commands in chu run REPL.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

TEST_NAME="Working Directory and Environment Management"

echo " Scenario: $TEST_NAME"
echo "========================================="
echo ""
echo "Simulating: Developer navigating directories and setting"
echo "environment variables for different contexts"
echo ""

setup_test_dir "$TEST_NAME"

echo "Step 1: Create project structure"
echo "---------------------------------"
mkdir -p frontend backend config
create_test_file "frontend/index.html" "<html><body>Frontend</body></html>"
create_test_file "backend/server.go" "package main"
create_test_file "config/dev.env" "API_URL=http://localhost:3000"

echo ""
echo "Step 2: Navigate to frontend directory"
echo "---------------------------------------"
run_gptcode_with_input "run" "/cd frontend
pwd
/exit" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "Changed to"
assert_contains "$OUTPUT" "frontend"

echo ""
echo "Step 3: Set environment variable"
echo "--------------------------------"
run_gptcode_with_input "run" "/env NODE_ENV=development
/env NODE_ENV
/exit" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "Set NODE_ENV=development"
assert_contains "$OUTPUT" "NODE_ENV=development"

echo ""
echo "Step 4: List all environment variables"
echo "---------------------------------------"
run_gptcode_with_input "run" "/env API_KEY=secret123
/env DB_URL=postgres://localhost
/env
/exit" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "API_KEY=secret123"
assert_contains "$OUTPUT" "DB_URL=postgres://localhost"

echo ""
echo "Step 5: Change directory and verify pwd"
echo "----------------------------------------"
run_gptcode_with_input "run" "/cd backend
pwd
ls
/exit" "--raw"
assert_exit_code 0
assert_contains "$OUTPUT" "backend"
assert_contains "$OUTPUT" "server.go"

cleanup_test_dir

echo ""
echo "========================================="
echo " Scenario passed: $TEST_NAME"
