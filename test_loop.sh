#!/bin/bash

cd /Users/jadercorrea/workspace/opensource/gptcode

echo "=== Starting test loop ==="
iteration=1

while true; do
  echo ""
  echo "==================== ITERATION $iteration ===================="
  
  echo "[1/4] Building..."
  go install ./cmd/gptcode 2>&1 | grep -v "^$"
  
  echo "[2/4] Testing CLI..."
  timeout 5 bash -c 'echo "{\"messages\":[{\"role\":\"user\",\"content\":\"Analyse codebase and add pix\"}]}" | gptcode chat 2>/dev/null' | head -20
  
  echo "[3/4] Testing Neovim event parsing..."
  nvim --headless -u test_events_simple.lua 2>&1 | grep -E "\[TEST\]"
  
  echo "[4/4] Testing Neovim full flow..."
  nvim --headless -u test_full_flow.lua 2>&1 | grep -E "\[TEST\]" | tail -10
  
  echo ""
  echo "Press Enter to run next iteration, or Ctrl+C to exit"
  read -t 30 || echo "Auto-continuing..."
  
  iteration=$((iteration + 1))
done
