#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
CAST1="assets/feedback-demo.cast"
GIF1="assets/feedback-demo.gif"
CAST2="assets/feedback-hook-demo.cast"
GIF2="assets/feedback-hook-demo.gif"
CAST3="assets/feedback-story.cast"
GIF3="assets/feedback-story.gif"
TRIES=${TRIES:-3}

record_and_check() {
  local cmd="$1" cast="$2" gif="$3"
  local i=0
  while (( i < TRIES )); do
    ((i++))
    asciinema rec --overwrite --command "$cmd" "$cast" --quiet || true
if grep -q "wrong_response" "$cast" && grep -q "correct_response" "$cast"; then
      docker run --rm -v "$PWD":/data ghcr.io/asciinema/agg:latest --theme dracula --font-size 16 "/data/$cast" "/data/$gif" >/dev/null 2>&1 || true
      if [[ -s "$gif" ]]; then
        return 0
      fi
    fi
    sleep 0.3
  done
  return 1
}

chmod +x assets/record-feedback-demo.zsh assets/record-feedback-hook-demo.zsh || true

record_and_check "zsh assets/record-feedback-demo.zsh" "$CAST1" "$GIF1"
record_and_check "zsh assets/record-feedback-hook-demo.zsh" "$CAST2" "$GIF2"
record_and_check "zsh assets/record-feedback-story.zsh" "$CAST3" "$GIF3"

echo "OK: demos built"
