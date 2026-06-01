[CmdletBinding()]
param(
    [string]$SampleDir,
    [string[]]$SamplePaths,
    [switch]$FailOnViolation
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$pythonRunner = Join-Path $repoRoot "scripts/run-python.ps1"
$scriptPath = Join-Path $repoRoot "scripts/run-eval-regression.py"

if (-not (Test-Path -LiteralPath $scriptPath -PathType Leaf)) {
    throw "missing python runner: scripts/run-eval-regression.py"
}

$arguments = @()
$arguments += $scriptPath
$arguments += "radishflow-suggest-edits"

if (-not [string]::IsNullOrWhiteSpace($SampleDir)) {
    $arguments += "--sample-dir"
    $arguments += $SampleDir
}

if ($SamplePaths.Count -gt 0) {
    $arguments += "--sample-paths"
    $arguments += $SamplePaths
}

if ($FailOnViolation) {
    $arguments += "--fail-on-violation"
}

& $pythonRunner @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
