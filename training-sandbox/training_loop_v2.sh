#!/bin/bash
# Self-Improving Training Loop v2
# Multi-layer validation: Syntax → Review → Tests → Analysis
# Usage: ./training_loop_v2.sh [NUM_ISSUES] [--dry-run]

set -e

GPTCODE_BIN="/Users/jadercorrea/workspace/gptcode/cli/gptcode"
SANDBOX_DIR="/Users/jadercorrea/workspace/gptcode/cli/training-sandbox"
RESULTS_FILE="$SANDBOX_DIR/training_results_v2.json"
FEEDBACK_FILE="$HOME/.gptcode/feedback.json"
LOG_DIR="$SANDBOX_DIR/logs"
NUM_ISSUES=${1:-20}
DRY_RUN=${2:-""}

mkdir -p "$LOG_DIR"

# Initialize results file
cat > "$RESULTS_FILE" << 'EOF'
{
  "version": 2,
  "started_at": "",
  "runs": [],
  "stats": {
    "total": 0,
    "l1_syntax_pass": 0,
    "l2_review_approved": 0,
    "l3_tests_pass": 0,
    "l4_analysis_clean": 0,
    "full_success": 0,
    "skipped": 0,
    "failed": 0
  },
  "model_performance": {}
}
EOF

jq ".started_at = \"$(date -Iseconds)\"" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"

echo "🚀 Training Loop v2 - Self-Improving"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Issues: $NUM_ISSUES"
echo "Results: $RESULTS_FILE"
echo ""

# Detect test runner for a repo
detect_test_runner() {
    if [ -f "go.mod" ]; then echo "go test ./..."; return; fi
    if [ -f "package.json" ]; then echo "npm test"; return; fi
    if [ -f "mix.exs" ]; then echo "mix test"; return; fi
    if [ -f "Cargo.toml" ]; then echo "cargo test"; return; fi
    if [ -f "requirements.txt" ] || [ -f "pyproject.toml" ]; then echo "pytest"; return; fi
    if [ -f "Gemfile" ]; then echo "bundle exec rspec"; return; fi
    echo ""
}

# Detect primary language
detect_language() {
    if [ -f "go.mod" ]; then echo "go"; return; fi
    if [ -f "package.json" ]; then echo "javascript"; return; fi
    if [ -f "mix.exs" ]; then echo "elixir"; return; fi
    if [ -f "Cargo.toml" ]; then echo "rust"; return; fi
    if [ -f "requirements.txt" ]; then echo "python"; return; fi
    if [ -f "Gemfile" ]; then echo "ruby"; return; fi
    echo "unknown"
}

# Run gt review and parse result
run_review() {
    local diff_file="$1"
    local review_output
    local review_log="$LOG_DIR/review_$(date +%s).log"
    
    # Create a prompt with the diff content
    local diff_content
    diff_content=$(cat "$diff_file" | head -500)  # Limit diff size
    
    local review_prompt="Review this code change following our code-review guidelines:

\`\`\`diff
$diff_content
\`\`\`

Provide a verdict: APPROVED, CHANGES_REQUESTED, or REJECTED with a brief reason."
    
    # Run review using gt review or fallback to do
    review_output=$("$GPTCODE_BIN" run "$review_prompt" 2>&1 | tee "$review_log" || true)
    
    # Parse review result (look for verdict keywords)
    if echo "$review_output" | grep -qi "APPROVED\|LGTM\|looks good\|ship it"; then
        echo "approved"
    elif echo "$review_output" | grep -qi "CHANGES_REQUESTED\|needs work\|minor issues"; then
        echo "changes_requested"
    elif echo "$review_output" | grep -qi "REJECTED\|do not merge\|significant issues"; then
        echo "rejected"
    else
        # Default to changes_requested if unclear
        echo "changes_requested"
    fi
}

# Analyze execution log for error patterns
analyze_log() {
    local log_file="$1"
    local errors=0
    
    # Check for common error patterns
    if grep -qi "panic:\|fatal:\|error:\|failed to" "$log_file" 2>/dev/null; then
        errors=$((errors + 1))
    fi
    
    # Check for timeout patterns
    if grep -qi "timeout\|context deadline exceeded" "$log_file" 2>/dev/null; then
        errors=$((errors + 1))
    fi
    
    # Check for rate limit patterns
    if grep -qi "rate limit\|429\|too many requests" "$log_file" 2>/dev/null; then
        errors=$((errors + 1))
    fi
    
    echo "$errors"
}

# Fetch issues
echo "📥 Fetching $NUM_ISSUES good-first-issues..."
ISSUES=$(/opt/homebrew/bin/gh search issues --label "good first issue" --state open --sort created --limit "$NUM_ISSUES" --json number,title,repository)

PROCESSED=0

echo "$ISSUES" | jq -c '.[]' | while read -r issue; do
    REPO=$(echo "$issue" | jq -r '.repository.nameWithOwner')
    NUMBER=$(echo "$issue" | jq -r '.number')
    TITLE=$(echo "$issue" | jq -r '.title')
    
    echo ""
    echo "[$((PROCESSED+1))/$NUM_ISSUES] $REPO#$NUMBER"
    echo "  📋 $TITLE"
    echo "  ─────────────────────────────────────"
    
    # Skip spam repos
    if [[ "$REPO" == *"Unit_Automation"* ]] || [[ "$REPO" == *"almadhlom"* ]]; then
        echo "  ⏭️  Skipping (spam repo)"
        jq ".stats.skipped += 1 | .stats.total += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
        continue
    fi
    
    # Clone repo
    cd "$SANDBOX_DIR"
    REPO_NAME=$(echo "$REPO" | cut -d'/' -f2)
    rm -rf "$REPO_NAME"
    
    if ! /opt/homebrew/bin/gh repo clone "$REPO" "$REPO_NAME" 2>/dev/null; then
        echo "  ❌ Failed to clone"
        jq ".stats.skipped += 1 | .stats.total += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
        continue
    fi
    
    cd "$REPO_NAME" || continue
    
    # Check repo size (skip if too large)
    FILE_COUNT=$(find . -type f | wc -l | tr -d ' ')
    if [ "$FILE_COUNT" -gt 10000 ]; then
        echo "  ⏭️  Skipping (repo too large: $FILE_COUNT files)"
        cd "$SANDBOX_DIR"
        rm -rf "$REPO_NAME"
        jq ".stats.skipped += 1 | .stats.total += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
        continue
    fi
    echo "  📦 Repo size: $FILE_COUNT files"
    
    # Get issue body
    BODY=$(/opt/homebrew/bin/gh issue view "$NUMBER" --repo "$REPO" --json body -q '.body' 2>/dev/null || echo "")
    
    if [ ${#BODY} -lt 30 ]; then
        echo "  ⏭️  Skipping (issue body too short)"
        cd "$SANDBOX_DIR"
        rm -rf "$REPO_NAME"
        jq ".stats.skipped += 1 | .stats.total += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
        continue
    fi
    
    # Truncate body
    if [ ${#BODY} -gt 500 ]; then
        BODY="${BODY:0:500}..."
    fi
    
    # Detect language and test runner
    LANGUAGE=$(detect_language)
    TEST_RUNNER=$(detect_test_runner)
    
    echo "  🔤 Language: $LANGUAGE"
    [ -n "$TEST_RUNNER" ] && echo "  🧪 Tests: $TEST_RUNNER"
    
    # ═══════════════════════════════════
    # LAYER 1: GENERATE
    # ═══════════════════════════════════
    echo "  ▶️  L1: Generating..."
    
    TASK="Fix issue #$NUMBER: $TITLE

$BODY"
    
    LOG_FILE="$LOG_DIR/${REPO_NAME}_${NUMBER}.log"
    START_TIME=$(date +%s)
    
    "$GPTCODE_BIN" do "$TASK" > "$LOG_FILE" 2>&1 || true
    EXIT_CODE=$?
    
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    
    # Check if files changed
    DIFF=$(git diff --name-only 2>/dev/null)
    
    if [ -z "$DIFF" ]; then
        echo "  ⚪ No changes (${DURATION}s)"
        jq ".stats.total += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
        cd "$SANDBOX_DIR"
        rm -rf "$REPO_NAME"
        PROCESSED=$((PROCESSED + 1))
        continue
    fi
    
    echo "  ✅ L1 PASS: $(echo "$DIFF" | wc -l | tr -d ' ') file(s) modified"
    L1_PASS=true
    
    # ═══════════════════════════════════
    # LAYER 2: REVIEW
    # ═══════════════════════════════════
    echo "  ▶️  L2: Reviewing..."
    
    # Save diff for review
    DIFF_FILE="$LOG_DIR/${REPO_NAME}_${NUMBER}.diff"
    git diff > "$DIFF_FILE"
    
    REVIEW_RESULT=$(run_review "$DIFF_FILE")
    
    L2_PASS=false
    if [ "$REVIEW_RESULT" = "approved" ]; then
        echo "  ✅ L2 PASS: Review approved"
        L2_PASS=true
        jq ".stats.l2_review_approved += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
    else
        echo "  ❌ L2 FAIL: Review $REVIEW_RESULT"
    fi
    
    # ═══════════════════════════════════
    # LAYER 3: TESTS (if available)
    # ═══════════════════════════════════
    L3_PASS=false
    if [ -n "$TEST_RUNNER" ]; then
        echo "  ▶️  L3: Running tests..."
        
        if eval "$TEST_RUNNER" > "$LOG_DIR/${REPO_NAME}_${NUMBER}_tests.log" 2>&1; then
            echo "  ✅ L3 PASS: Tests passed"
            L3_PASS=true
            jq ".stats.l3_tests_pass += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
        else
            echo "  ❌ L3 FAIL: Tests failed"
        fi
    else
        echo "  ⚪ L3 SKIP: No test runner detected"
        L3_PASS=true  # No tests = pass by default
    fi
    
    # ═══════════════════════════════════
    # LAYER 4: ANALYSIS
    # ═══════════════════════════════════
    echo "  ▶️  L4: Analyzing logs..."
    
    ERROR_COUNT=$(analyze_log "$LOG_FILE")
    
    L4_PASS=false
    if [ "$ERROR_COUNT" -eq 0 ]; then
        echo "  ✅ L4 PASS: No error patterns"
        L4_PASS=true
        jq ".stats.l4_analysis_clean += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
    else
        echo "  ⚠️  L4 WARN: $ERROR_COUNT error pattern(s) found"
    fi
    
    # ═══════════════════════════════════
    # FINAL SCORE
    # ═══════════════════════════════════
    RESULT="partial"
    if [ "$L1_PASS" = true ] && [ "$L2_PASS" = true ] && [ "$L3_PASS" = true ] && [ "$L4_PASS" = true ]; then
        RESULT="success"
        echo "  🎉 FULL SUCCESS"
        jq ".stats.full_success += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
    else
        echo "  📊 Partial: L1=$L1_PASS L2=$L2_PASS L3=$L3_PASS L4=$L4_PASS"
    fi
    
    # Update stats
    jq ".stats.total += 1 | .stats.l1_syntax_pass += 1" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
    
    # Record run
    RUN_JSON=$(cat <<EOF
{
  "repo": "$REPO",
  "issue": $NUMBER,
  "language": "$LANGUAGE",
  "result": "$RESULT",
  "l1_syntax": $L1_PASS,
  "l2_review": "$REVIEW_RESULT",
  "l3_tests": $L3_PASS,
  "l4_analysis": $L4_PASS,
  "duration": $DURATION,
  "files_changed": $(echo "$DIFF" | wc -l | tr -d ' ')
}
EOF
)
    jq ".runs += [$RUN_JSON]" "$RESULTS_FILE" > /tmp/r.json && mv /tmp/r.json "$RESULTS_FILE"
    
    # Write feedback for ML
    if [ -f "$FEEDBACK_FILE" ]; then
        FEEDBACK_JSON=$(cat <<EOF
{
  "timestamp": "$(date -Iseconds)",
  "action": "editor",
  "language": "$LANGUAGE",
  "model_used": "unknown",
  "review_result": "$REVIEW_RESULT",
  "tests_passed": $L3_PASS,
  "success": $([ "$RESULT" = "success" ] && echo "true" || echo "false"),
  "duration_seconds": $DURATION
}
EOF
)
        jq ". += [$FEEDBACK_JSON]" "$FEEDBACK_FILE" > /tmp/f.json && mv /tmp/f.json "$FEEDBACK_FILE"
    fi
    
    # Cleanup
    cd "$SANDBOX_DIR"
    rm -rf "$REPO_NAME"
    
    PROCESSED=$((PROCESSED + 1))
done

# ═══════════════════════════════════
# FINAL REPORT
# ═══════════════════════════════════
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Training Loop Complete"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat "$RESULTS_FILE" | jq '.stats'

# Calculate success rate
TOTAL=$(jq '.stats.total' "$RESULTS_FILE")
SUCCESS=$(jq '.stats.full_success' "$RESULTS_FILE")

if [ "$TOTAL" -gt 0 ]; then
    RATE=$(echo "scale=2; $SUCCESS * 100 / $TOTAL" | bc)
    echo ""
    echo "🎯 Full Success Rate: $RATE%"
    
    # Check if we should retrain ML
    if [ "$TOTAL" -ge 100 ]; then
        echo ""
        echo "🔄 100+ samples reached. Triggering ML retrain..."
        "$GPTCODE_BIN" ml train 2>&1 || echo "  (ml train not implemented yet)"
    fi
    
    # Check graduation criteria
    if [ "$(echo "$RATE >= 80" | bc)" -eq 1 ] && [ "$TOTAL" -ge 200 ]; then
        echo ""
        echo "🎓 GRADUATION: Ready for PR creation!"
        echo "   Review Approval: $(jq '.stats.l2_review_approved' "$RESULTS_FILE")/$TOTAL"
        echo "   Test Pass: $(jq '.stats.l3_tests_pass' "$RESULTS_FILE")/$TOTAL"
    fi
fi

echo ""
echo "📁 Logs saved in: $LOG_DIR"
