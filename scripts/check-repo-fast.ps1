[CmdletBinding()]
param(
    [switch]$SkipTextFiles
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$wrapperPath = Join-Path $repoRoot "scripts/check-repo.ps1"

if (-not (Test-Path -LiteralPath $wrapperPath -PathType Leaf)) {
    throw "missing wrapper script: scripts/check-repo.ps1"
}

$arguments = @()
if ($SkipTextFiles) {
    $arguments += "-SkipTextFiles"
}
$arguments += "-Fast"

& $wrapperPath @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
