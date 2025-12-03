#!/usr/bin/env bash
#
# Chuchu E2E Test Suite
#
# Runs realistic scenarios testing Chuchu commands in real-world use cases.
# Each scenario represents actual user workflows (DevOps, CI/CD, development).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$SCRIPT_DIR/e2e"

BACKEND=""
PROFILE=""

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --backend)
                BACKEND="$2"
                shift 2
                ;;
            --profile)
                PROFILE="$2"
                shift 2
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --backend BACKEND    Backend to use (e.g., groq, ollama, openai)"
                echo "  --profile PROFILE    Profile to use (e.g., local, default)"
                echo "  -h, --help           Show this help message"
                echo ""
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                echo "Run with --help for usage information"
                exit 1
                ;;
        esac
    done
}

parse_args "$@"

if [ -n "$BACKEND" ]; then
    export CHUCHU_E2E_BACKEND="$BACKEND"
fi

if [ -n "$PROFILE" ]; then
    export CHUCHU_E2E_PROFILE="$PROFILE"
fi

source "$E2E_DIR/lib/helpers.sh"

echo " Chuchu E2E Test Suite"
echo "============================"
echo ""

check_chu_installed() {
    if ! command -v chu &> /dev/null; then
        echo " Error: chu command not found"
        echo ""
        echo "Please install chu first:"
        echo "  cd $(dirname "$SCRIPT_DIR")"
        echo "  go install ./cmd/chu"
        exit 1
    fi
    
    echo "✓ chu command found: $(which chu)"
    echo ""
}

run_scenario() {
    local scenario_file="$1"
    local scenario_name=$(basename "$scenario_file" .sh | tr '_' ' ' | sed 's/.*/\u&/')
    
    echo ""
    echo " Running scenario: $scenario_name"
    echo "---"
    
    if bash "$scenario_file"; then
        echo " PASSED: $scenario_name"
        return 0
    else
        echo " FAILED: $scenario_name"
        return 1
    fi
}

main() {
    check_chu_installed
    setup_e2e_backend "$BACKEND" "$PROFILE"
    
    local failed=0
    local passed=0
    local total=0
    
    echo "Discovering scenarios..."
    echo ""
    
    for scenario_file in "$E2E_DIR"/scenarios/*.sh; do
        if [ -f "$scenario_file" ]; then
            total=$((total + 1))
            if run_scenario "$scenario_file"; then
                passed=$((passed + 1))
            else
                failed=$((failed + 1))
            fi
        fi
    done
    
    echo ""
    echo "============================"
    echo " Test Results"
    echo "============================"
    echo "Total:  $total"
    echo "Passed: $passed"
    echo "Failed: $failed"
    echo ""
    
    if [ $failed -eq 0 ]; then
        echo " All scenarios passed!"
        exit 0
    else
        echo " Some scenarios failed"
        exit 1
    fi
}

main
