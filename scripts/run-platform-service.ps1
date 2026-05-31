[CmdletBinding()]
param(
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
    $env:GOCACHE = Join-Path ([System.IO.Path]::GetTempPath()) "radishmind-go-build-cache"
}

if (-not $env:RADISHMIND_PLATFORM_PYTHON_BIN) {
    $venvPythonWindows = Join-Path $repoRoot ".venv/Scripts/python.exe"
    $venvPythonUnix = Join-Path $repoRoot ".venv/bin/python"
    if (Test-Path -LiteralPath $venvPythonWindows -PathType Leaf) {
        $env:RADISHMIND_PLATFORM_PYTHON_BIN = $venvPythonWindows
    }
    elseif (Test-Path -LiteralPath $venvPythonUnix -PathType Leaf) {
        $env:RADISHMIND_PLATFORM_PYTHON_BIN = $venvPythonUnix
    }
    else {
        throw "missing repository virtual environment. Run pwsh ./scripts/bootstrap-dev.ps1 before running scripts/run-platform-service.ps1"
    }
}

if (-not $env:RADISHMIND_PLATFORM_CONFIG) {
    $defaultConfig = Join-Path $repoRoot "tmp/radishmind-platform.local.json"
    if (Test-Path -LiteralPath $defaultConfig -PathType Leaf) {
        $env:RADISHMIND_PLATFORM_CONFIG = $defaultConfig
    }
}

Push-Location $platformDir
try {
    switch ($Command) {
        "serve" {
            & go run ./cmd/radishmind-platform @RemainingArgs
        }
        { $_ -in @("config-summary", "config-check", "diagnostics") } {
            & go run ./cmd/radishmind-platform $Command @RemainingArgs
        }
        default {
            [Console]::Error.WriteLine("unsupported platform service command: $Command")
            [Console]::Error.WriteLine("supported commands: serve, config-summary, config-check, diagnostics")
            exit 2
        }
    }
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
finally {
    Pop-Location
}
