set -e
setopt nonomatch

sugg='fly ssh console --exec "iex -S mix"'
corr='fly ssh console --pty -C "/app/bin/platform remote"'

echo "Suggested: $sugg"
echo "[Pressing Ctrl+g]"
mkdir -p ~/.gptcode
print -r -- "$sugg" > ~/.gptcode/last_suggestion_cmd
echo

echo "Running corrected command"
echo "$corr"

gptcode feedback submit \
  --sentiment=bad \
  --kind=command \
  --source=shell \
  --agent=editor \
  --task="Open Elixir console on Fly.io" \
  --wrong="$sugg" \
--correct="$corr" \
  --capture-diff

echo

echo "Stats after hook:"
gptcode feedback stats | sed -n '1,40p'

echo

echo "Last feedback event:"
lf=$(ls -t ~/.gptcode/feedback/*.json 2>/dev/null | head -n1)
if [ -n "$lf" ]; then
  tail -n 40 "$lf"
else
  echo "(no feedback file found)"
fi
