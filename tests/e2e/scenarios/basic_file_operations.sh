#!/usr/bin/env bash
#
# Scenario: Basic File Operations
#
# Validates core file manipulation capabilities that users rely on daily:
# - Creating files with content
# - Reading and displaying file contents
# - Modifying existing files
# - Creating structured data (JSON, YAML)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/.."
source "$SCRIPT_DIR/lib/helpers.sh"

TEST_NAME="Basic File Operations"

echo " Scenario: $TEST_NAME"
echo "========================================="
echo ""
echo "Testing: File creation, reading, and modification"
echo ""

setup_test_dir "$TEST_NAME"

echo "Test 1: Create text file with content"
echo "--------------------------------------"
run_gptcode_command "do" "create a file named hello.txt with the text 'Hello World'"
assert_exit_code 0
assert_file_exists "hello.txt"
assert_file_contains "hello.txt" "Hello"

echo ""
echo "Test 2: Read and display file contents"
echo "---------------------------------------"
run_gptcode_command "do" "show me what's in hello.txt"
assert_exit_code 0
assert_contains "$OUTPUT" "Hello"

echo ""
echo "Test 3: Append to existing file"
echo "--------------------------------"
run_gptcode_command "do" "add a new line saying 'Goodbye' to hello.txt"
assert_exit_code 0
assert_file_contains "hello.txt" "Goodbye"

echo ""
echo "Test 4: Create JSON file with structure"
echo "----------------------------------------"
run_gptcode_command "do" "create config.json with name=myapp and version=1.0"
assert_exit_code 0
assert_file_exists "config.json"
assert_file_contains "config.json" "myapp"
assert_file_contains "config.json" "1.0"

echo ""
echo "Test 5: Create YAML configuration"
echo "----------------------------------"
run_gptcode_command "do" "create settings.yaml with database host localhost and port 5432"
assert_exit_code 0
assert_file_exists "settings.yaml"
assert_file_contains "settings.yaml" "localhost"
assert_file_contains "settings.yaml" "5432"

echo ""
echo "Test 6: List all files in directory"
echo "------------------------------------"
run_gptcode_command "do" "list all files in the current directory"
assert_exit_code 0
assert_contains "$OUTPUT" "hello.txt"
assert_contains "$OUTPUT" "config.json"
assert_contains "$OUTPUT" "settings.yaml"

cleanup_test_dir

echo ""
echo "========================================="
echo " Scenario passed: $TEST_NAME"
