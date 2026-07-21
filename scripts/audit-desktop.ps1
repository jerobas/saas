$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$appDir = Join-Path $repoRoot "app"
$frontendDir = Join-Path $appDir "frontend"
. (Join-Path $PSScriptRoot "toolchain.ps1")

$previousGoToolchain = $env:GOTOOLCHAIN
$env:GOTOOLCHAIN = "go1.26.5"

try {
    Assert-DevelopmentToolchain

    Push-Location $appDir
    try {
        $goPackages = Get-DesktopGoPackages
        go tool govulncheck @goPackages
        if ($LASTEXITCODE -ne 0) { throw "Go vulnerability audit failed." }
    } finally {
        Pop-Location
    }

    Push-Location $frontendDir
    try {
        npm audit --audit-level=high
        if ($LASTEXITCODE -ne 0) { throw "Frontend vulnerability audit failed." }
    } finally {
        Pop-Location
    }
} finally {
    $env:GOTOOLCHAIN = $previousGoToolchain
}

Write-Host "Desktop dependency audits passed."
