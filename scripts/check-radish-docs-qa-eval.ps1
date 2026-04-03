[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$runnerPath = Join-Path $repoRoot "scripts/run-radish-docs-qa-regression.ps1"

if (-not (Test-Path -LiteralPath $runnerPath -PathType Leaf)) {
    throw "missing regression runner: scripts/run-radish-docs-qa-regression.ps1"
}

& $runnerPath -FailOnViolation
if ($LASTEXITCODE -ne 0) {
    throw "command failed: ./scripts/run-radish-docs-qa-regression.ps1"
}
