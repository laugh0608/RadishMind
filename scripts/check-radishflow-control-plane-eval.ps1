[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$runnerPath = Join-Path $repoRoot "scripts/run-radishflow-control-plane-regression.ps1"

if (-not (Test-Path -LiteralPath $runnerPath -PathType Leaf)) {
    throw "missing regression runner: scripts/run-radishflow-control-plane-regression.ps1"
}

& $runnerPath -FailOnViolation
if ($LASTEXITCODE -ne 0) {
    throw "command failed: ./scripts/run-radishflow-control-plane-regression.ps1"
}
