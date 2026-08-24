#!/usr/bin/env bash
# One-click friendly install for MCP over MOQT transport.
set -euo pipefail

REPO="${MCP_MOQT_REPO:-github.com/mcp-moqt/mcp-moqt-transport}"
BIN_DIR="${MCP_MOQT_BIN:-${HOME}/.local/bin}"
VERSION="${MCP_MOQT_VERSION:-latest}"

# If run from a local checkout, prefer building that tree.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

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
if [[ -f "${REPO_ROOT}/go.mod" ]] && grep -q "module github.com/mcp-moqt/mcp-moqt-transport" "${REPO_ROOT}/go.mod" 2>/dev/null; then
  echo "    mode   : local checkout (${REPO_ROOT})"
  (cd "${REPO_ROOT}" && GOBIN="${BIN_DIR}" go install ./cmd/mcp-moqt)
else
  echo "    mode   : remote go install"
  GOBIN="${BIN_DIR}" go install "${REPO}/cmd/mcp-moqt@${VERSION}"
fi

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

Quick start (stdio for Cursor / MCP hosts):
  mcp-moqt server -stdio

Quick start (QUIC/MOQT):
  mcp-moqt server -addr 127.0.0.1:8080
  mcp-moqt client -addr 127.0.0.1:8080

Config samples:
  configs/mcp/cursor.mcp.json
  configs/mcp/claude_desktop.json

Docker:
  docker compose up --build

EOF
