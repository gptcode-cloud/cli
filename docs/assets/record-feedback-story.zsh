set -e
setopt nonomatch

sugg='fly ssh console --exec "iex -S mix"'
corr='fly ssh console --pty -C "/app/bin/platform remote"'

# 1) Show suggested command and try it (expect failure)
echo "Suggested: $sugg"
echo "Trying suggested command (expected to fail)"
eval "$sugg" || true

# 2) User presses Ctrl+g to mark suggestion (simulate)
echo "[Pressing Ctrl+g]"
mkdir -p ~/.gptcode
print -r -- "$sugg" > ~/.gptcode/last_suggestion_cmd

# 3) Edit & Run corrected command (hook will record feedback)
echo "Editing to corrected command"
echo "$corr"
eval "$corr" || true
# 4) Show stats and last event (hook submits automatically)
sleep 0.3
echo
echo "Stats after correction:"
gptcode feedback stats | sed -n '1,80p'

echo

echo "Last feedback event:"
lf=$(ls -t ~/.gptcode/feedback/*.json 2>/dev/null | head -n1)
if [ -n "$lf" ]; then
  tail -n 200 "$lf" | sed -n '1,200p'
else
  echo "(no feedback file found)"
fi
