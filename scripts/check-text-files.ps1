[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$pythonRunner = Join-Path $repoRoot "scripts/run-python.ps1"
$scriptPath = Join-Path $repoRoot "scripts/check-text-files.py"

if (-not (Test-Path -LiteralPath $scriptPath -PathType Leaf)) {
    throw "missing python checker: scripts/check-text-files.py"
}

& $pythonRunner $scriptPath
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
