#!/bin/bash
set -e

echo "🐺 GPTCode Installation Script"
echo ""

INSTALL_DIR="${HOME}/.config/nvim/pack/plugins/start/gptcode"

echo "1. Installing Go binary..."
cd "$(dirname "$0")"
go install ./cmd/chu
echo "   ✓ Binary installed to $(which chu)"

echo ""
echo "2. Installing Neovim plugin..."
rm -rf "${INSTALL_DIR}"
mkdir -p "$(dirname "${INSTALL_DIR}")"
cp -r neovim "${INSTALL_DIR}"
echo "   ✓ Plugin installed to ${INSTALL_DIR}"

echo ""
echo "3. Verifying installation..."
CHU_PATH=$(which chu)
if [ -z "$CHU_PATH" ]; then
  echo "   ✗ chu not found in PATH"
  exit 1
fi

echo "   Testing chu..."
echo '{"messages":[{"role":"user","content":"test"}]}' | timeout 2 chu chat >/dev/null 2>&1 && echo "   ✓ chu works" || echo "   ⚠ chu may need API keys (run 'chu setup')"

echo ""
echo "4. Checking Neovim plugin..."
if [ -f "${INSTALL_DIR}/lua/gptcode/init.lua" ]; then
  USES_PATH=$(grep -c '"chu"' "${INSTALL_DIR}/lua/gptcode/init.lua" || true)
  if [ "$USES_PATH" -ge 2 ]; then
    echo "   ✓ Plugin uses chu from PATH"
  else
    echo "   ✗ Plugin has hardcoded path"
    exit 1
  fi
else
  echo "   ✗ Plugin file not found"
  exit 1
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "Next steps:"
echo "  1. Restart Neovim"
echo "  2. Run :luafile ${HOME}/workspace/opensource/gptcode/debug_live.lua"
echo "  3. Open chat with <C-d>"
echo "  4. Send 'Analyse codebase and add pix payment'"
echo "  5. Monitor: tail -f /tmp/gptcode_debug.log"
echo ""
echo "If it shows 'thinking...':"
echo "  - Check /tmp/gptcode_debug.log for actual binary path"
echo "  - Verify stdout events are received"
echo "  - Test CLI directly: echo '{\"messages\":[...]}' | chu chat"
