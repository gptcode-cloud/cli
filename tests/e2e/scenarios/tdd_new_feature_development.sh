#!/usr/bin/env bash
#
# Scenario: TDD New Feature Development
#
# A developer starts from scratch and wants tests before implementation.
# This validates the TDD workflow: requirements → test generation → implementation.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

TEST_NAME="TDD New Feature Development"

echo " Scenario: $TEST_NAME"
echo "========================================="
echo ""
echo "Simulating: Developer using TDD to build a new feature"
echo ""

setup_test_dir "$TEST_NAME"

echo "Step 1: Create Go project structure"
echo "------------------------------------"
create_go_project "myapp"

cat > go.mod <<'EOF'
module myapp

go 1.22

require github.com/stretchr/testify v1.8.4
EOF

mkdir -p math_utils

echo ""
echo "Step 2: Use chu tdd to generate tests for a calculator"
echo "-------------------------------------------------------"
run_gptcode_command_with_timeout "tdd" "Create a Calculator struct with Add, Subtract, Multiply, Divide methods for integers"
assert_exit_code 0

echo ""
echo "Step 3: Verify test output mentions calculator operations"
echo "-----------------------------------------------------------"
run_gptcode_command_with_timeout "tdd" "Write tests for a string utility with Reverse and ToUpperCase functions"
assert_exit_code 0

echo ""
echo "Step 4: Check that implementation suggestions are provided"
echo "-----------------------------------------------------------"
run_gptcode_command_with_timeout "tdd" "Create a Validator with email and phone validation"
assert_exit_code 0

echo ""
echo "Step 5: Verify TDD output structure"
echo "------------------------------------"
run_gptcode_command_with_timeout "tdd" "Build a Cache with Get, Set, and Delete operations"
assert_exit_code 0

cleanup_test_dir

echo ""
echo "========================================="
echo " Scenario passed: $TEST_NAME"
