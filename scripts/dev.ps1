$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$goBin = Join-Path (go env GOPATH) "bin"
$wails = Join-Path $goBin "wails.exe"
if (-not (Test-Path $wails)) {
    throw "Wails is not installed. Run .\scripts\setup-dev.ps1 first."
}

if (-not $env:SAAS_DATA_DIR) {
    $env:SAAS_DATA_DIR = Join-Path $env:APPDATA "saas-dev"
}

Push-Location (Join-Path $repoRoot "app")
try {
    & $wails dev
} finally {
    Pop-Location
}
