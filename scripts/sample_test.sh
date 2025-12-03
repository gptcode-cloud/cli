#!/bin/bash

# Chu Sample Capabilities Test
# Tests representative capabilities from each category

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEST_DIR="$PROJECT_ROOT/../chu-test-workspace"
CHU_BIN="$PROJECT_ROOT/chu"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

echo "🧪 Chu Sample Capabilities Test"
echo "================================"
echo ""

test_cap() {
    local name="$1"
    local command="$2"
    local validation="$3"
    
    echo -n "Testing: $name... "
    
    if eval "$command" > /tmp/chu_test.log 2>&1; then
        if [ -z "$validation" ] || eval "$validation" > /dev/null 2>&1; then
            echo -e "${GREEN}✅${NC}"
            ((PASSED++))
            return 0
        fi
    fi
    
    echo -e "${RED}❌${NC}"
    echo "  Last output: $(tail -1 /tmp/chu_test.log)"
    ((FAILED++))
    return 1
}

mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

echo "## 1. Shell Commands & Data Processing"
test_cap "Check disk usage" \
    "cd $TEST_DIR && $CHU_BIN do 'show me how much disk space is being used'" \
    ":"

test_cap "Find large files" \
    "cd $TEST_DIR && $CHU_BIN do 'find the 5 largest files in this directory'" \
    ":"

test_cap "Search TODO comments" \
    "cd $TEST_DIR && echo 'TODO: fix this' > test.txt && $CHU_BIN do 'show me all TODO comments in my code'" \
    ":"

test_cap "Count lines of code" \
    "cd $TEST_DIR && echo 'line1' > a.txt && echo 'line2' > b.txt && $CHU_BIN do 'how many lines of code do I have?'" \
    ":"

echo ""
echo "## 2. File Operations"
test_cap "Create config file" \
    "cd $TEST_DIR && rm -f config.yaml && $CHU_BIN do 'create a config.yaml with database settings'" \
    "[ -f config.yaml ]"

test_cap "Show file contents" \
    "cd $TEST_DIR && echo 'important data' > data.txt && $CHU_BIN do 'what is in data.txt?'" \
    ":"

test_cap "Create JSON data" \
    "cd $TEST_DIR && rm -f users.json && $CHU_BIN do 'create users.json with a list of 3 sample users'" \
    "[ -f users.json ]"

test_cap "Add to file" \
    "cd $TEST_DIR && echo '# Notes' > notes.md && $CHU_BIN do 'add a new note about testing to notes.md'" \
    "grep -iq 'test' notes.md"

echo ""
echo "## 3. Code Generation"
test_cap "Python script" \
    "cd $TEST_DIR && rm -f fetch_data.py && $CHU_BIN do 'write a Python script to fetch data from an API'" \
    "[ -f fetch_data.py ]"

test_cap "JavaScript utility" \
    "cd $TEST_DIR && rm -f helpers.js && $CHU_BIN do 'create a JavaScript module with array utility functions'" \
    "[ -f helpers.js ]"

test_cap "Go HTTP handler" \
    "cd $TEST_DIR && rm -f server.go && $CHU_BIN do 'create a basic HTTP server in Go'" \
    "[ -f server.go ]"

echo ""
echo "## 4. Git Operations"
if [ -d "$TEST_DIR/git-test" ]; then
    rm -rf "$TEST_DIR/git-test"
fi
mkdir -p "$TEST_DIR/git-test"
cd "$TEST_DIR/git-test"
git init > /dev/null 2>&1
git config user.email "test@chu.dev"
git config user.name "Test"
echo "# Test" > README.md
git add . > /dev/null 2>&1
git commit -m "init" > /dev/null 2>&1

test_cap "Check repo status" \
    "cd $TEST_DIR/git-test && $CHU_BIN do 'what is the current status of my repo?'" \
    ":"

test_cap "Show recent commits" \
    "cd $TEST_DIR/git-test && $CHU_BIN do 'show me the recent commits'" \
    ":"

test_cap "Show changes" \
    "cd $TEST_DIR/git-test && echo 'new feature' >> README.md && $CHU_BIN do 'what did I change?'" \
    ":"

echo ""
echo "## 5. Documentation"
cd "$TEST_DIR"
test_cap "Document project" \
    "cd $TEST_DIR && rm -f README.md && $CHU_BIN do 'write a README for my project'" \
    "[ -f README.md ]"

test_cap "Generate changelog" \
    "cd $TEST_DIR && rm -f CHANGELOG.md && $CHU_BIN do 'create a changelog documenting version 1.0.0'" \
    "[ -f CHANGELOG.md ]"

echo ""
echo "## 6. Package Management"
cd "$TEST_DIR"
test_cap "Initialize Node project" \
    "cd $TEST_DIR && rm -f package.json && $CHU_BIN do 'initialize a new Node.js project'" \
    "[ -f package.json ]"

echo ""
echo "## 7. Multi-tool Tasks"
test_cap "Find and replace text" \
    "cd $TEST_DIR && echo 'oldvalue' > config.txt && $CHU_BIN do 'replace oldvalue with newvalue in config.txt'" \
    "grep -q 'newvalue' config.txt"

test_cap "Search codebase" \
    "cd $TEST_DIR && echo 'function test()' > code.js && $CHU_BIN do 'find all function definitions'" \
    ":"

echo ""
echo "================================"
echo -e "Results: ${GREEN}✅ $PASSED passed${NC}, ${RED}❌ $FAILED failed${NC}"
echo "Total: $((PASSED + FAILED)) tests"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${YELLOW}⚠️  Some tests failed${NC}"
    exit 1
fi
