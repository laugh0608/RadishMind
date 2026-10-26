[CmdletBinding()]
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$action = if ($Arguments.Count -gt 0) { $Arguments[0] } else { "help" }
$forwarded = if ($Arguments.Count -gt 1) { $Arguments[1..($Arguments.Count - 1)] } else { @() }
$binaryName = if ($IsWindows) { "radishmind-platform.exe" } else { "radishmind-platform" }
$binary = Join-Path $repoRoot "tmp/workbench-bin/$binaryName"
Push-Location $repoRoot
try {
    if ($action -in @("help", "--help")) {
        Write-Output "Usage: ./start.ps1 workbench <build|check|init|serve|bootstrap-admin|backup|upgrade|restore> [options]"
        Write-Output "Use --config <file>; backup/upgrade require --output <new-directory>; restore requires --source and --output."
        Write-Output "Bootstrap requires --user-id and --audit-ref. Stop serve before data operations. No dependencies are installed automatically."
        exit 0
    }
    if ($action -eq "build") {
        if ($forwarded.Count -ne 0) { throw "build does not accept arguments" }
        if (-not (Test-Path "apps/radishmind-web/node_modules" -PathType Container)) { throw "Install the documented Web dependencies explicitly first." }
        Push-Location "apps/radishmind-web"
        try { npm run build -- --mode workbench; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE } }
        finally { Pop-Location }
        New-Item -ItemType Directory -Force -Path (Split-Path -Parent $binary) | Out-Null
        $buildVersion = git rev-parse --short=12 HEAD
        if ($LASTEXITCODE -ne 0) { throw "Cannot determine source revision" }
        $sourceChanges = git status --porcelain
        if ($LASTEXITCODE -ne 0) { throw "Cannot inspect source state" }
        if ($sourceChanges) { $buildVersion = "$buildVersion-dirty" }
        $previousProxy = $env:GOPROXY
        $previousToolchain = $env:GOTOOLCHAIN
        Push-Location "services/platform"
        try { $env:GOPROXY = "off"; $env:GOTOOLCHAIN = "local"; go build -ldflags "-X main.buildVersion=$buildVersion" -o $binary ./cmd/radishmind-platform; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE } }
        finally { $env:GOPROXY = $previousProxy; $env:GOTOOLCHAIN = $previousToolchain; Pop-Location }
        exit 0
    }
    if (-not (Test-Path $binary -PathType Leaf)) { throw "Build the workbench first: ./start.ps1 workbench build" }
    & $binary workbench $action @forwarded
    exit $LASTEXITCODE
}
finally { Pop-Location }
