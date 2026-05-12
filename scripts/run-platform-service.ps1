[CmdletBinding()]
param(
    [ValidateSet("serve", "config-summary", "config-check", "diagnostics")]
    [string]$Command = "serve",
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$RemainingArgs
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$platformDir = Join-Path $repoRoot "services/platform"

$goCommand = Get-Command go -ErrorAction SilentlyContinue
if ($null -eq $goCommand) {
    throw "go is required for scripts/run-platform-service.ps1"
}

if (-not $env:GOCACHE) {
    $env:GOCACHE = "/tmp/radishmind-go-build-cache"
}

if (-not $env:RADISHMIND_PLATFORM_CONFIG) {
    $defaultConfig = Join-Path $repoRoot "tmp/radishmind-platform.local.json"
    if (Test-Path -LiteralPath $defaultConfig -PathType Leaf) {
        $env:RADISHMIND_PLATFORM_CONFIG = $defaultConfig
    }
}

Push-Location $platformDir
try {
    if ($Command -eq "serve") {
        & go run ./cmd/radishmind-platform @RemainingArgs
    } else {
        & go run ./cmd/radishmind-platform $Command @RemainingArgs
    }
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
finally {
    Pop-Location
}
