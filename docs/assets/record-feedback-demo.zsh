set -e
setopt nonomatch

echo "Suggested: fly ssh console --exec \"iex -S mix\""
echo

echo "Submitting corrected command as feedback"
gptcode feedback submit \
  --sentiment=bad \
  --kind=command \
  --source=shell \
  --agent=editor \
  --task="Open Elixir console on Fly.io" \
  --wrong='fly ssh console --exec "iex -S mix"' \
  --correct='fly ssh console --pty -C "/app/bin/platform remote"' \
  --capture-diff

echo
echo "Last feedback event:"
lf=$(ls -t ~/.gptcode/feedback/*.json 2>/dev/null | head -n1)
if [ -n "$lf" ]; then
  tail -n 40 "$lf"
else
  echo "(no feedback file found)"
fi

echo
echo "Training:"
echo "python3 ml/intent/scripts/process_feedback.py"
echo "gptcode ml train intent"
