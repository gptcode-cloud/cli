#!/bin/sh
# codegpt installer
# Usage: curl -fsSL https://opsrig.io/install-codegpt.sh | sh

set -e

REPO="opsrig/codegpt"
BINARY_NAME="codegpt"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Download URL (placeholder - needs GitHub releases)
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}_${OS}_${ARCH}"

echo "Installing ${BINARY_NAME}..."
mkdir -p "$INSTALL_DIR"

if command -v curl > /dev/null; then
    curl -fsSL "$DOWNLOAD_URL" -o "${INSTALL_DIR}/${BINARY_NAME}"
elif command -v wget > /dev/null; then
    wget -q "$DOWNLOAD_URL" -O "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo "Error: curl or wget required"
    exit 1
fi

chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Add to PATH hint
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo ""
    echo "Add to your shell profile:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo ""
echo "✓ ${BINARY_NAME} installed to ${INSTALL_DIR}/${BINARY_NAME}"
echo ""
echo "Get started:"
echo "  ${BINARY_NAME} do \"create a hello world server\""
