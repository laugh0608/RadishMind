[CmdletBinding()]
param(
    [string]$Command = "",
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$RemainingArgs
)

$ErrorActionPreference = "Stop"

$repoRoot = $PSScriptRoot
$webDevScript = Join-Path $repoRoot "scripts/run-radishmind-web-dev.ps1"
$consoleDevScript = Join-Path $repoRoot "scripts/run-radishmind-console-dev.ps1"
$platformScript = Join-Path $repoRoot "scripts/run-platform-service.ps1"

if (-not (Test-Path -LiteralPath $webDevScript -PathType Leaf)) {
    throw "Cannot find $webDevScript. Run this script from the RadishMind repository root."
}

function Show-Usage {
@"
Usage: pwsh ./start.ps1 [-Command command] [args]

Commands:
  web-live       Start product web UI + dev-only fake read backend. Default for frontend/backend integration.
  web-offline    Start product web UI with offline fixtures only.
  console        Start local ops console + platform backend.
  platform       Start platform backend with dev-only read auth enabled.
  web-build      Run apps/radishmind-web build.
  diagnostics    Run platform diagnostics.
  check-fast     Run ./scripts/check-repo.ps1 -Fast.
  help           Show this help.

Without -Command, this script opens an interactive menu.
Extra args are forwarded to the selected dev launcher.
"@
}

function Show-Banner {
@"
========================================
   ____           _ _     _     __  __ _           _
  |  _ \ __ _  __| (_)___| |__ |  \/  (_)_ __   __| |
  | |_) / _` |/ _` | / __| '_ \| |\/| | | '_ \ / _` |
  |  _ < (_| | (_| | \__ \ | | | |  | | | | | | (_| |
  |_| \_\__,_|\__,_|_|___/_| |_|_|  |_|_|_| |_|\__,_|
        RadishMind local dev launcher
========================================
"@
}

function Show-Menu {
    Write-Host
    Write-Host "RadishMind local start menu"
    Write-Host "---------------------------"
    Write-Host "  0. Exit"
    Write-Host
    Write-Host "[Product UI]"
    Write-Host "  1. Product Web dev-live  (platform @ http://127.0.0.1:7000 + web @ http://127.0.0.1:4100)"
    Write-Host "  2. Product Web offline   (web @ http://127.0.0.1:4100)"
    Write-Host
    Write-Host "[Ops / Backend]"
    Write-Host "  3. Local Ops Console     (platform @ http://127.0.0.1:7000 + console @ http://127.0.0.1:4000)"
    Write-Host "  4. Platform backend      (dev-only read auth enabled @ http://127.0.0.1:7000)"
    Write-Host "  5. Platform diagnostics"
    Write-Host
    Write-Host "[Verification]"
    Write-Host "  6. Web build             (apps/radishmind-web)"
    Write-Host "  7. Fast repo check       (pwsh ./scripts/check-repo.ps1 -Fast)"
    Write-Host
    Write-Host "Tip: Product Web dev-live is the default choice for frontend/backend integration."
    Write-Host "Tip: If macOS Control Center / AirPlay occupies backend 7000, run pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -BackendUrl http://127.0.0.1:7100."
    Write-Host
}

function Invoke-WebLive {
    & $webDevScript -Mode "dev-live" @RemainingArgs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-WebOffline {
    & $webDevScript -Mode "offline" @RemainingArgs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-Console {
    & $consoleDevScript @RemainingArgs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-Platform {
    if (-not $env:RADISHMIND_PLATFORM_LISTEN_ADDR) {
        $env:RADISHMIND_PLATFORM_LISTEN_ADDR = "127.0.0.1:7000"
    }
    if (-not $env:RADISHMIND_PLATFORM_PROVIDER) {
        $env:RADISHMIND_PLATFORM_PROVIDER = "mock"
    }
    if (-not $env:RADISHMIND_PLATFORM_MODEL) {
        $env:RADISHMIND_PLATFORM_MODEL = "radishmind-local-dev"
    }
    $env:RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH = "1"
    & $platformScript -Command serve @RemainingArgs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-WebBuild {
    Push-Location (Join-Path $repoRoot "apps/radishmind-web")
    try {
        npm run build
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    finally {
        Pop-Location
    }
}

function Invoke-Diagnostics {
    & $platformScript -Command diagnostics @RemainingArgs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-CheckFast {
    & (Join-Path $repoRoot "scripts/check-repo.ps1") -Fast @RemainingArgs
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Invoke-CommandName {
    param([string]$Name)
    switch ($Name) {
        { $_ -in @("web-live", "web", "dev-live") } { Invoke-WebLive; return }
        { $_ -in @("web-offline", "offline") } { Invoke-WebOffline; return }
        "console" { Invoke-Console; return }
        { $_ -in @("platform", "backend") } { Invoke-Platform; return }
        { $_ -in @("web-build", "build") } { Invoke-WebBuild; return }
        { $_ -in @("diagnostics", "diag") } { Invoke-Diagnostics; return }
        { $_ -in @("check-fast", "check") } { Invoke-CheckFast; return }
        { $_ -in @("help", "-h", "--help") } { Show-Usage; return }
        default {
            [Console]::Error.WriteLine("Unknown command: $Name")
            Show-Usage
            exit 2
        }
    }
}

if (-not [string]::IsNullOrWhiteSpace($Command)) {
    Invoke-CommandName -Name $Command
    exit 0
}

Show-Banner
Show-Menu
$choice = Read-Host "Enter choice number"

switch ($choice) {
    "0" { Write-Host "Bye."; exit 0 }
    "1" { Invoke-WebLive }
    "2" { Invoke-WebOffline }
    "3" { Invoke-Console }
    "4" { Invoke-Platform }
    "5" { Invoke-Diagnostics }
    "6" { Invoke-WebBuild }
    "7" { Invoke-CheckFast }
    default {
        [Console]::Error.WriteLine("Unknown choice: $choice")
        exit 2
    }
}
