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
        go mod download
        if ($LASTEXITCODE -ne 0) { throw "Could not download Go dependencies." }

        go tool wails version
        if ($LASTEXITCODE -ne 0) { throw "Could not prepare the pinned Wails CLI." }

        go tool staticcheck -version
        if ($LASTEXITCODE -ne 0) { throw "Could not prepare the pinned Staticcheck CLI." }

        go tool actionlint -version
        if ($LASTEXITCODE -ne 0) { throw "Could not prepare the pinned actionlint CLI." }

        go tool govulncheck -version
        if ($LASTEXITCODE -ne 0) { throw "Could not prepare the pinned govulncheck CLI." }
    } finally {
        Pop-Location
    }

    Push-Location $frontendDir
    try {
        npm ci
        if ($LASTEXITCODE -ne 0) { throw "Could not install frontend dependencies." }

        npx --no-install playwright install chromium
        if ($LASTEXITCODE -ne 0) { throw "Could not install the Playwright Chromium browser." }
    } finally {
        Pop-Location
    }
} finally {
    $env:GOTOOLCHAIN = $previousGoToolchain
}

Write-Host "Development dependencies are ready."
Write-Host "Run: .\scripts\dev.ps1"
