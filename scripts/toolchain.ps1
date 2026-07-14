Set-StrictMode -Version Latest

function Assert-CommandAvailable {
    param([Parameter(Mandatory)][string]$Name)

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "$Name is required. See docs/development/toolchain.md."
    }
}

function ConvertTo-ToolVersion {
    param(
        [Parameter(Mandatory)][string]$Value,
        [Parameter(Mandatory)][string]$Tool
    )

    try {
        return [Version]::Parse($Value.Trim().TrimStart("v"))
    } catch {
        throw "Could not parse the $Tool version '$Value'."
    }
}

function Assert-DevelopmentToolchain {
    Assert-CommandAvailable "go"
    Assert-CommandAvailable "node"
    Assert-CommandAvailable "npm"

    $goVersionOutput = & go env GOVERSION
    $goVersion = "$goVersionOutput".Trim()
    if ($LASTEXITCODE -ne 0 -or $goVersion -ne "go1.26.5") {
        throw "Go toolchain 1.26.5 is required; found '$goVersion'. Run scripts/setup-dev.ps1 with network access."
    }

    $nodeVersion = ConvertTo-ToolVersion (& node --version) "Node.js"
    if ($nodeVersion -lt [Version]"24.17.0" -or $nodeVersion -ge [Version]"25.0.0") {
        throw "Node.js 24.17.0 or a newer Node 24 patch is required; found $nodeVersion."
    }

    $npmVersion = ConvertTo-ToolVersion (& npm --version) "npm"
    if ($npmVersion -lt [Version]"11.13.0" -or $npmVersion -ge [Version]"12.0.0") {
        throw "npm 11.13.0 or a newer npm 11 patch is required; found $npmVersion."
    }
}

function Get-DesktopGoPackages {
    $packages = @(& go list ./...)
    if ($LASTEXITCODE -ne 0) {
        throw "Could not enumerate desktop Go packages."
    }

    return @($packages | Where-Object { $_ -notmatch "/frontend/" })
}
