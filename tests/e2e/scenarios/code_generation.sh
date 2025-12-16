#!/usr/bin/env bash
#
# Scenario: Code Generation
#
# Validates code generation capabilities across multiple languages:
# - Python scripts
# - JavaScript/TypeScript modules
# - Go programs
# - Shell scripts
# - Configuration files

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/.."
source "$SCRIPT_DIR/lib/helpers.sh"

TEST_NAME="Code Generation"

echo " Scenario: $TEST_NAME"
echo "========================================="
echo ""
echo "Testing: Multi-language code generation"
echo ""

setup_test_dir "$TEST_NAME"

echo "Test 1: Generate Python script"
echo "-------------------------------"
run_gptcode_command "do" "create a Python script calc.py that adds two numbers"
assert_exit_code 0
assert_file_exists "calc.py"
assert_file_contains "calc.py" "def"

echo ""
echo "Test 2: Generate JavaScript module"
echo "-----------------------------------"
run_gptcode_command "do" "create utils.js with array utility functions"
assert_exit_code 0
assert_file_exists "utils.js"
assert_file_contains "utils.js" "function"

echo ""
echo "Test 3: Generate Go program"
echo "---------------------------"
run_gptcode_command "do" "create main.go with a hello world HTTP server"
assert_exit_code 0
assert_file_exists "main.go"
assert_file_contains "main.go" "package main"
assert_file_contains "main.go" "http"

echo ""
echo "Test 4: Generate shell script"
echo "------------------------------"
run_gptcode_command "do" "create backup.sh that backs up files to a directory"
assert_exit_code 0
assert_file_exists "backup.sh"
assert_file_contains "backup.sh" "#!/"

echo ""
echo "Test 5: Generate package.json"
echo "------------------------------"
run_gptcode_command "do" "initialize a Node.js project with package.json"
assert_exit_code 0
assert_file_exists "package.json"
assert_file_contains "package.json" "name"
assert_file_contains "package.json" "version"

echo ""
echo "Test 6: Generate Makefile"
echo "-------------------------"
run_gptcode_command "do" "create a Makefile with build and test targets"
assert_exit_code 0
assert_file_exists "Makefile"
assert_file_contains "Makefile" "build"
assert_file_contains "Makefile" "test"

cleanup_test_dir

echo ""
echo "========================================="
echo " Scenario passed: $TEST_NAME"
