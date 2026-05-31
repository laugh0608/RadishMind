[CmdletBinding()]
param(
    [string]$SampleDir,
    [string[]]$SamplePaths,
    [string]$NegativeReplayIndex,
    [string]$BatchArtifactSummary,
    [string[]]$GroupId,
    [string[]]$RecordId,
    [int]$RecommendedGroupsTop,
    [ValidateSet("same_sample", "cross_sample")]
    [string]$ReplayMode,
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
$arguments += "radish-docs-qa-negative"

if (-not [string]::IsNullOrWhiteSpace($SampleDir)) {
    $arguments += "--sample-dir"
    $arguments += $SampleDir
}

if ($SamplePaths.Count -gt 0) {
    $arguments += "--sample-paths"
    $arguments += $SamplePaths
}

if (-not [string]::IsNullOrWhiteSpace($NegativeReplayIndex)) {
    $arguments += "--negative-replay-index"
    $arguments += $NegativeReplayIndex
}

if (-not [string]::IsNullOrWhiteSpace($BatchArtifactSummary)) {
    $arguments += "--batch-artifact-summary"
    $arguments += $BatchArtifactSummary
}

if ($GroupId.Count -gt 0) {
    foreach ($item in $GroupId) {
        $arguments += "--group-id"
        $arguments += $item
    }
}

if ($RecordId.Count -gt 0) {
    foreach ($item in $RecordId) {
        $arguments += "--record-id"
        $arguments += $item
    }
}

if ($RecommendedGroupsTop -gt 0) {
    $arguments += "--recommended-groups-top"
    $arguments += $RecommendedGroupsTop
}

if (-not [string]::IsNullOrWhiteSpace($ReplayMode)) {
    $arguments += "--replay-mode"
    $arguments += $ReplayMode
}

if ($FailOnViolation) {
    $arguments += "--fail-on-violation"
}

& $pythonRunner @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
