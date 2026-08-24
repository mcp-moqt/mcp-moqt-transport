#!/usr/bin/env bash
# One-click friendly install for MCP over MOQT transport.
set -euo pipefail

REPO="${MCP_MOQT_REPO:-github.com/mcp-moqt/mcp-moqt-transport}"
BIN_DIR="${MCP_MOQT_BIN:-${HOME}/.local/bin}"
VERSION="${MCP_MOQT_VERSION:-latest}"

echo "==> MCP over MOQT friendly installer"
echo "    module : ${REPO}"
echo "    version: ${VERSION}"
echo "    bin dir: ${BIN_DIR}"

if ! command -v go >/dev/null 2>&1; then
  echo "error: Go is required. Install from https://go.dev/dl/" >&2
  exit 1
fi

mkdir -p "${BIN_DIR}"

echo "==> Installing CLI (mcp-moqt)"
GOBIN="${BIN_DIR}" go install "${REPO}/cmd/mcp-moqt@${VERSION}"

# Ensure PATH hint
case ":${PATH}:" in
  *":${BIN_DIR}:"*) ;;
  *)
    echo "==> Tip: add ${BIN_DIR} to PATH"
    echo "    export PATH=\"${BIN_DIR}:\$PATH\""
    ;;
esac

echo "==> Verifying"
"${BIN_DIR}/mcp-moqt" version || true
"${BIN_DIR}/mcp-moqt" doctor || true

cat <<EOF

Install complete.

Quick start:
  mcp-moqt server -addr 127.0.0.1:8080
  mcp-moqt client -addr 127.0.0.1:8080

Docker alternative:
  docker compose up --build

EOF
