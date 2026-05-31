[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$BatchArtifactSummary,
    [int]$Top = 1,
    [string]$SampleDir,
    [ValidateSet("same_sample", "cross_sample")]
    [string]$ReplayMode,
    [switch]$FailOnViolation,
    [string]$SummaryOutput,
    [switch]$Check
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$pythonRunner = Join-Path $repoRoot "scripts/run-python.ps1"
$scriptPath = Join-Path $repoRoot "scripts/run-radish-docs-qa-negative-recommended.py"

if (-not (Test-Path -LiteralPath $scriptPath -PathType Leaf)) {
    throw "missing python runner: scripts/run-radish-docs-qa-negative-recommended.py"
}

$arguments = @()
$arguments += $scriptPath
$arguments += "--batch-artifact-summary"
$arguments += $BatchArtifactSummary
$arguments += "--top"
$arguments += $Top

if (-not [string]::IsNullOrWhiteSpace($SampleDir)) {
    $arguments += "--sample-dir"
    $arguments += $SampleDir
}

if (-not [string]::IsNullOrWhiteSpace($ReplayMode)) {
    $arguments += "--replay-mode"
    $arguments += $ReplayMode
}

if ($FailOnViolation) {
    $arguments += "--fail-on-violation"
}

if (-not [string]::IsNullOrWhiteSpace($SummaryOutput)) {
    $arguments += "--summary-output"
    $arguments += $SummaryOutput
}

if ($Check) {
    $arguments += "--check"
}

& $pythonRunner @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
