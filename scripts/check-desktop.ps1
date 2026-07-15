param([switch]$SkipDesktopBuild)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$appDir = Join-Path $repoRoot "app"
$frontendDir = Join-Path $appDir "frontend"
$frontendSource = Join-Path $appDir "frontend\src"
. (Join-Path $PSScriptRoot "toolchain.ps1")

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

$goFiles = Get-ChildItem -LiteralPath $appDir -Recurse -Filter "*.go" -File |
    Where-Object { $_.FullName -notlike "$frontendDir\*" }
$unformatted = & gofmt -l $goFiles.FullName
if ($unformatted) {
    $unformatted | ForEach-Object { Write-Error "Not formatted: $_" }
    throw "Run gofmt on the files listed above."
}

$previousGoToolchain = $env:GOTOOLCHAIN
$previousGoProxy = $env:GOPROXY
$previousNpmOffline = $env:npm_config_offline

try {
    $env:GOTOOLCHAIN = "go1.26.5"
    $env:GOPROXY = "off"
    $env:npm_config_offline = "true"

    Assert-DevelopmentToolchain

    Push-Location $frontendDir
    try {
        npm run check
        if ($LASTEXITCODE -ne 0) { throw "Frontend checks failed." }
    } finally {
        Pop-Location
    }

    Push-Location $appDir
    try {
        go tool sqlc compile
        if ($LASTEXITCODE -ne 0) { throw "sqlc query validation failed." }

        go tool sqlc diff
        if ($LASTEXITCODE -ne 0) { throw "Generated sqlc code is stale. Run 'go tool sqlc generate'." }

        go mod tidy -diff
        if ($LASTEXITCODE -ne 0) { throw "go.mod or go.sum is not tidy." }

        $goPackages = Get-DesktopGoPackages

        go vet @goPackages
        if ($LASTEXITCODE -ne 0) { throw "Go vet failed." }

        go tool staticcheck @goPackages
        if ($LASTEXITCODE -ne 0) { throw "Staticcheck failed." }

        $workflowFiles = Get-ChildItem -LiteralPath (Join-Path $repoRoot ".github\workflows") -Filter "*.yml" -File
        go tool actionlint $workflowFiles.FullName
        if ($LASTEXITCODE -ne 0) { throw "GitHub Actions validation failed." }

        $raceFlag = @()
        if ((& go env CGO_ENABLED).Trim() -eq "1") {
            $raceFlag = @("-race")
        } else {
            Write-Warning "CGO is disabled; running Go tests without the race detector. CI still enforces the race build on Linux."
        }

        go test @raceFlag -shuffle=on -count=1 @goPackages
        if ($LASTEXITCODE -ne 0) { throw "Go tests failed." }

        if (-not $SkipDesktopBuild) {
            go tool wails build -m -nosyncgomod -nopackage -webview2 error -o app-check.exe
            if ($LASTEXITCODE -ne 0) {
                throw "Wails build failed with exit code $LASTEXITCODE."
            }
        }
    } finally {
        Pop-Location
    }
} finally {
    $env:GOTOOLCHAIN = $previousGoToolchain
    $env:GOPROXY = $previousGoProxy
    $env:npm_config_offline = $previousNpmOffline
}

Write-Host "Desktop checks passed without dependency downloads or remote runtime services."
