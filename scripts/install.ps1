# One-click friendly install for MCP over MOQT transport (Windows PowerShell).
param(
    [string]$Repo = $(if ($env:MCP_MOQT_REPO) { $env:MCP_MOQT_REPO } else { "github.com/mcp-moqt/mcp-moqt-transport" }),
    [string]$BinDir = $(if ($env:MCP_MOQT_BIN) { $env:MCP_MOQT_BIN } else { Join-Path $env:USERPROFILE ".local\bin" }),
    [string]$Version = $(if ($env:MCP_MOQT_VERSION) { $env:MCP_MOQT_VERSION } else { "latest" })
)

$ErrorActionPreference = "Stop"

Write-Host "==> MCP over MOQT friendly installer"
Write-Host "    module : $Repo"
Write-Host "    version: $Version"
Write-Host "    bin dir: $BinDir"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go is required. Install from https://go.dev/dl/"
}

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

Write-Host "==> Installing CLI (mcp-moqt)"
$env:GOBIN = $BinDir
go install "$Repo/cmd/mcp-moqt@$Version"

$pathParts = $env:Path -split ";"
if ($pathParts -notcontains $BinDir) {
    Write-Host "==> Tip: add $BinDir to PATH"
    Write-Host "    `$env:Path = `"$BinDir;`$env:Path`""
}

Write-Host "==> Verifying"
& "$BinDir\mcp-moqt.exe" version
& "$BinDir\mcp-moqt.exe" doctor

Write-Host @"

Install complete.

Quick start:
  mcp-moqt server -addr 127.0.0.1:8080
  mcp-moqt client -addr 127.0.0.1:8080

Docker alternative:
  docker compose up --build

"@
