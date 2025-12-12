#!/bin/bash
set -e

echo "=== Installing GPTCode Neovim Plugin ==="
echo ""

if ! command -v nvim &> /dev/null; then
    echo "Error: Neovim is not installed"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

NVIM_CONFIG="${XDG_CONFIG_HOME:-$HOME/.config}/nvim"
if [ ! -d "$NVIM_CONFIG" ]; then
    echo "Error: Neovim config directory not found at $NVIM_CONFIG"
    exit 1
fi

# Create plugin directory
PLUGIN_DIR="$NVIM_CONFIG/pack/plugins/start/gptcode"
echo "1. Creating plugin directory: $PLUGIN_DIR"
mkdir -p "$PLUGIN_DIR/lua"

# Copy lua files
echo "2. Copying Lua files..."
cp -v ./neovim/lua/gptcode/init.lua "$PLUGIN_DIR/lua/gptcode.lua"

# Ensure binary is in place
if [ ! -f "$HOME/.local/bin/chu" ]; then
    echo ""
    echo "Warning: chu binary not found at ~/.local/bin/chu"
    echo "Building it now..."
    go build -o "$HOME/.local/bin/chu" ./cmd/chu
    echo "✓ Binary built and installed"
fi

echo ""
echo "✓ Plugin installed successfully!"
echo ""
echo "Next steps:"
echo "1. Restart Neovim or run: :luafile $PLUGIN_DIR/lua/gptcode.lua"
echo "2. Test with: :lua require('gptcode').start_code_conversation()"
echo ""
echo "The plugin is now using: ~/.local/bin/chu"
