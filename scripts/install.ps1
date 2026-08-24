# One-click friendly install for MCP over MOQT transport (Windows PowerShell).
param(
    [string]$Repo = $(if ($env:MCP_MOQT_REPO) { $env:MCP_MOQT_REPO } else { "github.com/mcp-moqt/mcp-moqt-transport" }),
    [string]$BinDir = $(if ($env:MCP_MOQT_BIN) { $env:MCP_MOQT_BIN } else { Join-Path $env:USERPROFILE ".local\bin" }),
    [string]$Version = $(if ($env:MCP_MOQT_VERSION) { $env:MCP_MOQT_VERSION } else { "latest" })
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Resolve-Path (Join-Path $ScriptDir "..")

Write-Host "==> MCP over MOQT friendly installer"
Write-Host "    module : $Repo"
Write-Host "    version: $Version"
Write-Host "    bin dir: $BinDir"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go is required. Install from https://go.dev/dl/"
}

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
$env:GOBIN = $BinDir

$goMod = Join-Path $RepoRoot "go.mod"
if ((Test-Path $goMod) -and (Select-String -Path $goMod -Pattern "module github.com/mcp-moqt/mcp-moqt-transport" -Quiet)) {
    Write-Host "==> Installing CLI from local checkout ($RepoRoot)"
    Push-Location $RepoRoot
    try {
        go install ./cmd/mcp-moqt
    } finally {
        Pop-Location
    }
} else {
    Write-Host "==> Installing CLI via go install"
    go install "$Repo/cmd/mcp-moqt@$Version"
}

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

Quick start (stdio):
  mcp-moqt server -stdio

Quick start (QUIC/MOQT):
  mcp-moqt server -addr 127.0.0.1:8080
  mcp-moqt client -addr 127.0.0.1:8080

Config samples:
  configs/mcp/cursor.mcp.json
  configs/mcp/claude_desktop.json

Docker:
  docker compose up --build

"@
