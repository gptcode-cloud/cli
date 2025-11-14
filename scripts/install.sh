#!/usr/bin/env bash
set -euo pipefail

echo "[chuchu] Building and installing CLI..."

# Descobre go bin
if [[ -z "${GOBIN:-}" ]]; then
  GOBIN="$(go env GOPATH)/bin"
fi

echo "[chuchu] GOBIN=${GOBIN}"

# Instala binário
go install ./cmd/chu

# Garante que esteja no PATH
if ! command -v chu >/dev/null 2>&1; then
  echo "[chuchu] Warning: 'chu' is not on your PATH."
  echo "         Make sure ${GOBIN} is in your PATH."
else
  echo "[chuchu] Running initial setup..."
  chu setup || true
fi

echo "[chuchu] Done."
