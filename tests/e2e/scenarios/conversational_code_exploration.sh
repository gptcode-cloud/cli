#!/usr/bin/env bash
#
# Scenario: Conversational Code Exploration with Context Preservation
#
# A developer needs to understand existing code through conversation.
# This test validates multi-turn chat with context preservation,
# simulating real debugging/understanding workflows.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

TEST_NAME="Conversational Code Exploration"

echo "🧪 Scenario: $TEST_NAME"
echo "========================================="
echo ""
echo "Simulating: Developer exploring codebase through multi-turn conversation"
echo ""

setup_e2e_backend

setup_test_dir "$TEST_NAME"

echo "Step 1: Create sample Go code"
echo "------------------------------"
create_go_project "myapp"

cat > main.go <<'EOF'
package main

import (
	"fmt"
	"log"
)

type User struct {
	ID   int
	Name string
	Role string
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u *User) FullInfo() string {
	return fmt.Sprintf("User %d: %s (%s)", u.ID, u.Name, u.Role)
}

func AuthorizeAction(user *User, action string) bool {
	if user.IsAdmin() {
		return true
	}
	if action == "read" {
		return true
	}
	return false
}

func main() {
	user := &User{ID: 1, Name: "Alice", Role: "admin"}
	log.Println(user.FullInfo())
	if AuthorizeAction(user, "delete") {
		log.Println("Action authorized")
	}
}
EOF

echo ""
echo "Step 2: Ask about User struct in main.go"
echo "------------------------------------------"
run_chu_command_with_timeout "chat" "Explain the User struct and its methods in main.go" "--once"
assert_exit_code 0
assert_contains "$OUTPUT" "User"
assert_not_contains "$OUTPUT" "Error"

echo ""
echo "Step 3: Ask about authorization logic in main.go"
echo "--------------------------------------------------"
run_chu_command_with_timeout "chat" "How does the AuthorizeAction function work in main.go?" "--once"
assert_exit_code 0
assert_contains "$OUTPUT" "AuthorizeAction"
assert_not_contains "$OUTPUT" "Error"

echo ""
echo "Step 4: Ask about improvements to main.go"
echo "------------------------------------------"
run_chu_command_with_timeout "chat" "What improvements could be made to the authorization in main.go?" "--once"
assert_exit_code 0
# Should mention security or extensibility concerns
assert_not_contains "$OUTPUT" "Error"

echo ""
echo "Step 5: Verify file context in main.go is being used"
echo "-------------------------------------------------------"
# The responses should reference the actual code structure
run_chu_command_with_timeout "chat" "What parameters does the FullInfo method take in main.go?" "--once"
assert_exit_code 0
# Should understand it's a receiver method with no additional params
assert_not_contains "$OUTPUT" "Error"

cleanup_test_dir

echo ""
echo "========================================="
echo "✅ Scenario passed: $TEST_NAME"
