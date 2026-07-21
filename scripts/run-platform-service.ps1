[CmdletBinding()]
param(
    [string]$Command = "serve",
    [string]$Profile = "local-product",
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$RemainingArgs
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$platformDir = Join-Path $repoRoot "services/platform"

function Set-DefaultEnvironmentValue {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true)]
        [string]$Value
    )
    $currentValue = [Environment]::GetEnvironmentVariable($Name, "Process")
    if ([string]::IsNullOrWhiteSpace($currentValue)) {
        [Environment]::SetEnvironmentVariable($Name, $Value, "Process")
    }
}

switch ($Profile) {
    "local-product" {
        if ($env:RADISHMIND_LOCAL_PERSISTENCE_MODE -and $env:RADISHMIND_LOCAL_PERSISTENCE_MODE -ne "sqlite_dev") {
            [Console]::Error.WriteLine("local-product profile requires RADISHMIND_LOCAL_PERSISTENCE_MODE=sqlite_dev; use -Profile configured for explicit component configuration")
            exit 2
        }
        $env:RADISHMIND_LOCAL_PERSISTENCE_MODE = "sqlite_dev"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_SQLITE_DEV_DATABASE_PATH" -Value (Join-Path $repoRoot "var/sqlite-dev/radishmind.db")
        Set-DefaultEnvironmentValue -Name "RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_APPLICATION_DRAFT_DEV_HTTP" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_APPLICATION_DRAFT_DEV_WRITE" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_APPLICATION_CATALOG_DEV_HTTP" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_APPLICATION_CATALOG_DEV_WRITE" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_PROMPT_APPLICATION_TEMPLATE_DEV_HTTP" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_PROMPT_APPLICATION_TEMPLATE_DEV_WRITE" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_API_KEY_LIFECYCLE_DEV_HTTP" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_API_KEY_LIFECYCLE_DEV_WRITE" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_WORKFLOW_EXECUTOR_DEV" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_WORKFLOW_TOOL_ACTION_DEV" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_WORKFLOW_HTTP_TOOL_EXECUTION_DEV" -Value "1"
        Set-DefaultEnvironmentValue -Name "RADISHMIND_WORKFLOW_RAG_SNAPSHOT_DEV" -Value "1"
    }
    "configured" {}
    default {
        [Console]::Error.WriteLine("unsupported platform startup profile: $Profile")
        [Console]::Error.WriteLine("supported profiles: local-product, configured")
        exit 2
    }
}

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
