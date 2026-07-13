$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$appDir = Join-Path $repoRoot "app"
$frontendDir = Join-Path $appDir "frontend"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is required. Install the version declared in app/go.mod or newer."
}
if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
    throw "Node.js and npm are required."
}

Push-Location $appDir
try {
    go mod download
} finally {
    Pop-Location
}

Push-Location $frontendDir
try {
    npm ci
} finally {
    Pop-Location
}

$goBin = Join-Path (go env GOPATH) "bin"
$wails = Join-Path $goBin "wails.exe"
if (-not (Test-Path $wails)) {
    go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
}

Write-Host "Development dependencies are ready."
Write-Host "Run: .\scripts\dev.ps1"
