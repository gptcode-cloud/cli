#!/usr/bin/env bash

set -euo pipefail

TEST_DIR=""
OUTPUT=""
EXIT_CODE=0

show_current_profile() {
    echo "E2E Test Configuration"
    echo "======================"
    echo ""
    echo "Using current chu profile:"
    chu profile 2>&1 | head -6
    echo ""
    echo "Note: Configure profile with 'chu profile use <backend>.<profile>'"
    echo "      Example: chu profile use groq.budget"
    echo ""
}

setup_test_dir() {
    local test_name="$1"
    local safe_name=$(echo "$test_name" | tr ' ' '-' | tr '[:upper:]' '[:lower:]')
    TEST_DIR=$(mktemp -d "/tmp/gptcode-e2e-${safe_name}-XXXXXX")
    echo "📁 Test directory: $TEST_DIR"
    cd "$TEST_DIR"
}

cleanup_test_dir() {
    if [ -n "$TEST_DIR" ] && [ -d "$TEST_DIR" ]; then
        rm -rf "$TEST_DIR"
        echo "🧹 Cleaned up test directory"
    fi
}

run_chu_command() {
    local cmd="$1"
    shift
    
    set +e
    OUTPUT=$(chu "$cmd" "$@" 2>&1)
    EXIT_CODE=$?
    set -e
    
    echo "📤 Command output:"
    echo "$OUTPUT"
    echo "Exit code: $EXIT_CODE"
}

run_chu_command_with_timeout() {
    local timeout_seconds="${GPTCODE_E2E_TIMEOUT:-180}"
    local cmd="$1"
    shift
    
    set +e
    OUTPUT=$(timeout "$timeout_seconds" chu "$cmd" "$@" 2>&1)
    EXIT_CODE=$?
    set -e
    
    if [ "$EXIT_CODE" -eq 124 ]; then
        echo " Command timed out after ${timeout_seconds}s"
        echo "This usually means:"
        echo "  - Backend not properly configured"
        echo "  - Model taking too long to respond"
        echo "  - Network issues"
        exit 1
    fi
    
    echo "📤 Command output:"
    echo "$OUTPUT"
    echo "Exit code: $EXIT_CODE"
}

run_chu_with_input() {
    local cmd="$1"
    local input="$2"
    shift 2
    
    set +e
    OUTPUT=$(echo "$input" | chu "$cmd" "$@" 2>&1)
    EXIT_CODE=$?
    set -e
    
    echo "📤 Command output:"
    echo "$OUTPUT"
    echo "Exit code: $EXIT_CODE"
}

assert_contains() {
    local text="$1"
    local expected="$2"
    
    if echo "$text" | grep -q "$expected"; then
        echo "✓ Text contains '$expected'"
    else
        echo "✗ FAILED: Text does not contain '$expected'"
        echo "Text was:"
        echo "$text"
        exit 1
    fi
}

assert_not_contains() {
    local text="$1"
    local unexpected="$2"
    
    if echo "$text" | grep -q "$unexpected"; then
        echo "✗ FAILED: Text unexpectedly contains '$unexpected'"
        echo "Text was:"
        echo "$text"
        exit 1
    else
        echo "✓ Text does not contain '$unexpected'"
    fi
}

assert_exit_code() {
    local expected="$1"
    
    if [ "$EXIT_CODE" -eq "$expected" ]; then
        echo "✓ Exit code is $expected"
    else
        echo "✗ FAILED: Exit code is $EXIT_CODE, expected $expected"
        exit 1
    fi
}

assert_file_exists() {
    local filepath="$1"
    
    if [ -f "$filepath" ]; then
        echo "✓ File exists: $filepath"
    else
        echo "✗ FAILED: File does not exist: $filepath"
        exit 1
    fi
}

assert_file_not_exists() {
    local filepath="$1"
    
    if [ ! -f "$filepath" ]; then
        echo "✓ File does not exist: $filepath"
    else
        echo "✗ FAILED: File unexpectedly exists: $filepath"
        exit 1
    fi
}

assert_dir_exists() {
    local dirpath="$1"
    
    if [ -d "$dirpath" ]; then
        echo "✓ Directory exists: $dirpath"
    else
        echo "✗ FAILED: Directory does not exist: $dirpath"
        exit 1
    fi
}

assert_file_contains() {
    local filepath="$1"
    local expected="$2"
    
    if [ ! -f "$filepath" ]; then
        echo "✗ FAILED: File does not exist: $filepath"
        exit 1
    fi
    
    if grep -q "$expected" "$filepath"; then
        echo "✓ File contains '$expected': $filepath"
    else
        echo "✗ FAILED: File does not contain '$expected': $filepath"
        echo "File contents:"
        cat "$filepath"
        exit 1
    fi
}

create_test_file() {
    local filename="$1"
    local content="${2:-}"
    
    if [ -n "$content" ]; then
        echo "$content" > "$filename"
    else
        touch "$filename"
    fi
    echo " Created test file: $filename"
}

create_go_project() {
    local project_name="$1"
    
    mkdir -p "$project_name"
    cd "$project_name"
    
    cat > go.mod <<EOF
module $project_name

go 1.22
EOF
    
    cat > main.go <<EOF
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
EOF
    
    echo "🔧 Created Go project: $project_name"
}
