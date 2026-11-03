# Dot-source this file to configure only this PowerShell process.
param([string]$GoBin, [string]$DshBin, [ValidateSet('mxc', 'elevated', 'unelevated')][string]$CodexWindowsSandbox)
$ErrorActionPreference = 'Stop'
if ($CodexWindowsSandbox) { $env:AO_CODEX_WINDOWS_SANDBOX = $CodexWindowsSandbox }
if ($GoBin) {
    if (-not (Test-Path -LiteralPath (Join-Path $GoBin 'go.exe'))) { throw 'GoBin must contain go.exe' }
    $env:PATH = $GoBin + ';' + $env:PATH
}
if ($DshBin) {
    if (-not (Test-Path -LiteralPath (Join-Path $DshBin 'dsh.cmd'))) { throw 'DshBin must contain dsh.cmd' }
    $env:PATH = $DshBin + ';' + $env:PATH
}
# Never attach development or validation to the installed AO application.
$labRoot = Join-Path $env:USERPROFILE '.cache\ai-dev-studio-ultimate\lab'
New-Item -ItemType Directory -Force -Path $labRoot | Out-Null
$env:AO_DATA_DIR = Join-Path $labRoot 'data'
$env:AO_RUN_FILE = Join-Path $labRoot 'running.json'
$env:AO_DEV_ELECTRON_DIR = Join-Path $labRoot 'electron'
$env:AO_PORT = '31337'
$env:AO_TELEMETRY_REMOTE = 'off'
$env:AO_TELEMETRY_EVENTS = 'off'
$env:AO_TELEMETRY_METRICS = 'off'
Write-Output "Isolated development state: $labRoot"
