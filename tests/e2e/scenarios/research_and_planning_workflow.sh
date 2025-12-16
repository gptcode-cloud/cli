#!/usr/bin/env bash
#
# Scenario: Research and Planning Workflow
#
# A team member needs to understand a codebase before implementing a feature.
# This validates the research → plan workflow for structured development.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

TEST_NAME="Research and Planning Workflow"

echo " Scenario: $TEST_NAME"
echo "========================================="
echo ""
echo "Simulating: Developer researching codebase and planning new feature"
echo ""

setup_test_dir "$TEST_NAME"

echo "Step 1: Create sample Go project with auth"
echo "-------------------------------------------"
create_go_project "ecommerce"

cat > user.go <<'EOF'
package main

type User struct {
	ID       int
	Email    string
	Password string
	Role     string
}

func (u *User) HasPermission(action string) bool {
	if u.Role == "admin" {
		return true
	}
	return action == "read"
}
EOF

cat > auth.go <<'EOF'
package main

type AuthService struct {
	users map[int]*User
}

func (a *AuthService) Authenticate(email, password string) (*User, error) {
	for _, user := range a.users {
		if user.Email == email && user.Password == password {
			return user, nil
		}
	}
	return nil, ErrInvalidCredentials
}
EOF

echo ""
echo "Step 2: Research authentication system"
echo "--------------------------------------"
run_gptcode_command_with_timeout "research" "How does the authentication system work?"
assert_exit_code 0
assert_contains "$OUTPUT" "auth"

echo ""
echo "Step 3: Research user roles and permissions"
echo "-------------------------------------------"
run_gptcode_command_with_timeout "research" "Explain the user roles and permission system"
assert_exit_code 0

echo ""
echo "Step 4: Create plan for adding JWT tokens"
echo "------------------------------------------"
run_gptcode_command_with_timeout "plan" "Add JWT token-based authentication while keeping existing auth"
assert_exit_code 0

echo ""
echo "Step 5: Plan for adding OAuth integration"
echo "------------------------------------------"
run_gptcode_command_with_timeout "plan" "Integrate OAuth2 for third-party login"
assert_exit_code 0

cleanup_test_dir

echo ""
echo "========================================="
echo " Scenario passed: $TEST_NAME"
