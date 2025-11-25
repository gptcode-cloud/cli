#!/usr/bin/env bash

set -euo pipefail

TEST_DIR=""
OUTPUT=""
EXIT_CODE=0

# E2E Test Configuration
# Uses 'local' profile with optimized model allocation:
#   router: llama3.1:8b (fast)
#   query: gpt-oss:latest (reasoning)
#   editor: qwen3-coder:latest (coding)
#   research: gpt-oss:latest (reasoning)
CHUCHU_E2E_BACKEND="${CHUCHU_E2E_BACKEND:-ollama}"
CHUCHU_E2E_PROFILE="${CHUCHU_E2E_PROFILE:-local}"

setup_e2e_backend() {
    echo "🔧 Configuring E2E test backend..."
    
    # Check if Ollama is available
    if ! command -v ollama &> /dev/null; then
        echo ""
        echo "❌ ERROR: Ollama is required for E2E tests but not found"
        echo ""
        echo "To run E2E tests, you need:"
        echo "  1. Install Ollama: https://ollama.ai"
        echo "  2. Pull a recommended model:"
        echo "     ollama pull qwen2.5-coder:7b  (recommended, ~4GB)"
        echo "     ollama pull llama3.1:8b       (alternative, ~4.7GB)"
        echo "     ollama pull codellama:7b      (alternative, ~3.8GB)"
        echo ""
        exit 1
    fi
    
    echo "✓ Ollama found"
    
    # Check if required models are available
    local required_models=("llama3.1:8b" "qwen3-coder:latest" "gpt-oss:latest")
    local missing_models=()
    
    set +e +o pipefail
    for model in "${required_models[@]}"; do
        ollama list | grep -q "$model"
        if [ $? -ne 0 ]; then
            missing_models+=("$model")
        fi
    done
    set -e -o pipefail
    
    if [ ${#missing_models[@]} -gt 0 ]; then
        echo ""
        echo "❌ ERROR: Missing required models for 'local' profile:"
        for model in "${missing_models[@]}"; do
            echo "  - $model"
        done
        echo ""
        echo "Please pull the missing models:"
        for model in "${missing_models[@]}"; do
            echo "  ollama pull $model"
        done
        echo ""
        exit 1
    fi
    
    echo "✓ All required models available (llama3.1:8b, qwen3-coder:latest, gpt-oss:latest)"
    
    # Configure Chuchu to use Ollama with local profile
    export CHUCHU_BACKEND="$CHUCHU_E2E_BACKEND"
    chu config set defaults.backend ollama 2>&1 > /dev/null || true
    chu config set defaults.profile "$CHUCHU_E2E_PROFILE" 2>&1 > /dev/null || true
    
    echo "✓ Backend configured: $CHUCHU_E2E_BACKEND with profile $CHUCHU_E2E_PROFILE"
    echo "  Router: llama3.1:8b | Query: gpt-oss:latest | Editor: qwen3-coder:latest"
    echo ""
}

setup_test_dir() {
    local test_name="$1"
    local safe_name=$(echo "$test_name" | tr ' ' '-' | tr '[:upper:]' '[:lower:]')
    TEST_DIR=$(mktemp -d "/tmp/chuchu-e2e-${safe_name}-XXXXXX")
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
    local timeout_seconds="${CHUCHU_E2E_TIMEOUT:-180}"
    local cmd="$1"
    shift
    
    set +e
    OUTPUT=$(timeout "$timeout_seconds" chu "$cmd" "$@" 2>&1)
    EXIT_CODE=$?
    set -e
    
    if [ "$EXIT_CODE" -eq 124 ]; then
        echo "❌ Command timed out after ${timeout_seconds}s"
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

create_test_file() {
    local filename="$1"
    local content="${2:-}"
    
    if [ -n "$content" ]; then
        echo "$content" > "$filename"
    else
        touch "$filename"
    fi
    echo "📝 Created test file: $filename"
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
