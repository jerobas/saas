[CmdletBinding()]
param(
    [ValidateSet("Seed", "Start", "Clean")]
    [string]$Action = "Seed",

    [string]$DataDirectory,

    [ValidateRange(1, 200)]
    [int]$Scale = 30,

    [switch]$Force
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$markerName = ".sweeters-demo-data"
$markerContents = "sweeters-demo-data-v1"

if (-not $DataDirectory) {
    $DataDirectory = Join-Path $repoRoot "tmp\demo-data"
}

$dataPath = [System.IO.Path]::GetFullPath($DataDirectory)
$databasePath = Join-Path $dataPath "app.db"
$markerPath = Join-Path $dataPath $markerName
$developmentDataPath = [System.IO.Path]::GetFullPath((Join-Path $env:APPDATA "saas-dev"))
$packagedDataPath = [System.IO.Path]::GetFullPath((Join-Path $env:APPDATA "app"))

function Assert-SafeDemoPath {
    if ($dataPath -eq [System.IO.Path]::GetPathRoot($dataPath)) {
        throw "Refusing to use a filesystem root as the demo data directory."
    }
    if ($dataPath -eq [System.IO.Path]::GetFullPath($repoRoot)) {
        throw "Refusing to use the repository root as the demo data directory."
    }
    if ($dataPath -eq $developmentDataPath -or $dataPath -eq $packagedDataPath) {
        throw "Refusing to use a normal Sweeters data directory: $dataPath"
    }
}

function Assert-DemoMarker {
    if (-not (Test-Path -LiteralPath $markerPath -PathType Leaf)) {
        throw "Demo marker not found at $markerPath. Refusing to clean or start this directory."
    }
    $actualMarker = (Get-Content -LiteralPath $markerPath -Raw).Trim()
    if ($actualMarker -ne $markerContents) {
        throw "Demo marker is not recognized. Refusing to clean or start this directory."
    }
}

Assert-SafeDemoPath

switch ($Action) {
    "Seed" {
        if (Test-Path -LiteralPath $dataPath) {
            if (-not (Test-Path -LiteralPath $markerPath -PathType Leaf)) {
                throw "Directory already exists without the demo marker: $dataPath"
            }
            Assert-DemoMarker
            if (Test-Path -LiteralPath $databasePath) {
                throw "Demo database already exists. Clean the marker-protected directory first: $dataPath"
            }
        } else {
            New-Item -ItemType Directory -Path $dataPath | Out-Null
        }

        Set-Content -LiteralPath $markerPath -Value $markerContents -NoNewline

        . (Join-Path $PSScriptRoot "toolchain.ps1")
        $previousGoToolchain = $env:GOTOOLCHAIN
        $env:GOTOOLCHAIN = "go1.26.5"
        try {
            Assert-DevelopmentToolchain
            Push-Location (Join-Path $repoRoot "app")
            try {
                go run ./cmd/demo-data -database $databasePath -scale $Scale
                if ($LASTEXITCODE -ne 0) { throw "Demo data generation failed." }
            } finally {
                Pop-Location
            }
        } finally {
            $env:GOTOOLCHAIN = $previousGoToolchain
        }

        Write-Host ""
        Write-Host "Launch Sweeters with the demo database:"
        Write-Host "  .\scripts\demo-data.ps1 Start -DataDirectory `"$dataPath`""
        Write-Host ""
        Write-Host "Remove every generated record afterward:"
        Write-Host "  .\scripts\demo-data.ps1 Clean -DataDirectory `"$dataPath`" -Force"
    }

    "Start" {
        Assert-DemoMarker
        if (-not (Test-Path -LiteralPath $databasePath -PathType Leaf)) {
            throw "Demo database not found at $databasePath. Seed this directory first."
        }

        $previousDataDirectory = $env:SAAS_DATA_DIR
        $env:SAAS_DATA_DIR = $dataPath
        try {
            & (Join-Path $PSScriptRoot "dev.ps1")
        } finally {
            $env:SAAS_DATA_DIR = $previousDataDirectory
        }
    }

    "Clean" {
        Assert-DemoMarker
        if (-not $Force) {
            throw "Clean removes the entire marker-protected demo directory. Re-run with -Force."
        }

        $resolvedDemoPath = (Resolve-Path -LiteralPath $dataPath).Path
        if ($resolvedDemoPath -ne $dataPath) {
            throw "Resolved demo path changed unexpectedly. Refusing to clean: $resolvedDemoPath"
        }
        Remove-Item -LiteralPath $resolvedDemoPath -Recurse -Force
        Write-Host "Removed demo database and all generated records from $resolvedDemoPath"
    }
}
