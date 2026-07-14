$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$appDir = Join-Path $repoRoot "app"
$frontendSource = Join-Path $appDir "frontend\src"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is required. Run .\scripts\setup-dev.ps1 first."
}
if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
    throw "Node.js and npm are required. Run .\scripts\setup-dev.ps1 first."
}
$goBin = Join-Path (go env GOPATH) "bin"
$wails = Join-Path $goBin "wails.exe"
if (-not (Test-Path $wails)) {
    throw "Wails is required. Run .\scripts\setup-dev.ps1 first."
}

$remotePatterns = @(
    'from\s+["'']axios["'']',
    'new\s+EventSource\s*\(',
    'https://api\.vezono\.com'
)
$remoteReferences = Get-ChildItem -LiteralPath $frontendSource -Recurse -File |
    Where-Object { $_.Extension -in ".js", ".jsx", ".ts", ".tsx" } |
    Select-String -Pattern $remotePatterns
if ($remoteReferences) {
    $remoteReferences | ForEach-Object { Write-Error $_.ToString() }
    throw "The desktop frontend contains a forbidden remote runtime dependency."
}

$goFiles = Get-ChildItem -LiteralPath $appDir -Recurse -Filter "*.go" -File
$unformatted = & gofmt -l $goFiles.FullName
if ($unformatted) {
    $unformatted | ForEach-Object { Write-Error "Not formatted: $_" }
    throw "Run gofmt on the files listed above."
}

$previousGoToolchain = $env:GOTOOLCHAIN
$previousGoProxy = $env:GOPROXY
$previousGoSumDB = $env:GOSUMDB
$previousNpmOffline = $env:npm_config_offline

try {
    $env:GOTOOLCHAIN = "local"
    $env:GOPROXY = "off"
    $env:GOSUMDB = "off"
    $env:npm_config_offline = "true"

    Push-Location $appDir
    try {
        go vet ./...
        go test -count=1 ./...
        & $wails build -m -nosyncgomod -nopackage -webview2 error -o app-check.exe
        if ($LASTEXITCODE -ne 0) {
            throw "Wails build failed with exit code $LASTEXITCODE."
        }
    } finally {
        Pop-Location
    }
} finally {
    $env:GOTOOLCHAIN = $previousGoToolchain
    $env:GOPROXY = $previousGoProxy
    $env:GOSUMDB = $previousGoSumDB
    $env:npm_config_offline = $previousNpmOffline
}

Write-Host "Desktop checks passed without dependency downloads or remote runtime services."
