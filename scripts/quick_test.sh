#!/bin/bash

# Quick Capability Test - Focus on known working features
# Tests are lenient - just check if chu completes without fatal errors

set +e  # Don't exit on error

CHU="/Users/jadercorrea/workspace/opensource/gptcode/chu"
TEST_DIR="/Users/jadercorrea/workspace/opensource/chu-test-workspace"

echo " Quick Capability Test"
echo "========================"
echo ""

PASSED=0
FAILED=0

test_task() {
    local name="$1"
    local task="$2"
    local dir="${3:-$TEST_DIR}"
    
    echo -n "Testing: $name... "
    
    cd "$dir"
    if timeout 60 "$CHU" do "$task" >/dev/null 2>&1; then
        echo " PASSED"
        ((PASSED++))
    else
        echo " FAILED"
        ((FAILED++))
    fi
}

# Setup
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# Category 1: Simple Shell Commands (should work)
echo "## 1. Shell Commands"
test_task "Echo test" "echo hello world"
test_task "List files" "list files in current directory" "$TEST_DIR"
test_task "Check date" "what's the current date"

# Category 2: File Operations (should work)
echo ""
echo "## 2. File Operations"
test_task "Create file" "create a file named test.txt with content 'hello'"
test_task "Read file" "read the content of test.txt"
test_task "Create JSON" "create a package.json with name 'test-app' and version '1.0.0'"

# Category 3: Code Generation (should work)
echo ""
echo "## 3. Code Generation"
test_task "Create Python script" "create a simple Python hello world script in hello.py"
test_task "Create JS function" "create a JavaScript function that adds two numbers in add.js"

# Category 4: Documentation (should work)
echo ""
echo "## 4. Documentation"  
test_task "Create README" "create a README.md for a test project"
test_task "Create markdown list" "create a TODO list in markdown format in TODO.md"

# Category 5: Git Operations (might have success criteria issues)
echo ""
echo "## 5. Git Operations (lenient)"
cd "$TEST_DIR"
git init test-repo >/dev/null 2>&1
cd test-repo
git config user.email "test@test.com"
git config user.name "Test"
echo "test" > file.txt
git add . && git commit -m "init" >/dev/null 2>&1

# Just check if these don't crash completely
timeout 60 "$CHU" do "show git status" >/dev/null 2>&1 && echo " Git status" || echo " Git status"

echo ""
echo "========================"
echo "Results:  $PASSED passed,  $FAILED failed"
echo ""

exit 0
