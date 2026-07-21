$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
. (Join-Path $PSScriptRoot "toolchain.ps1")

if (-not $env:SAAS_DATA_DIR) {
    $env:SAAS_DATA_DIR = Join-Path $env:APPDATA "saas-dev"
}

$previousGoToolchain = $env:GOTOOLCHAIN
$env:GOTOOLCHAIN = "go1.26.5"

try {
    Assert-DevelopmentToolchain

    Push-Location (Join-Path $repoRoot "app")
    try {
        go tool wails dev
        if ($LASTEXITCODE -ne 0) { throw "Wails development mode failed." }
    } finally {
        Pop-Location
    }
} finally {
    $env:GOTOOLCHAIN = $previousGoToolchain
}
