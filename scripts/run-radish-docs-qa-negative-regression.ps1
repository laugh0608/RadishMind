[CmdletBinding()]
param(
    [string]$SampleDir,
    [string[]]$SamplePaths,
    [string]$NegativeReplayIndex,
    [string]$BatchArtifactSummary,
    [string[]]$GroupId,
    [string[]]$RecordId,
    [ValidateSet("same_sample", "cross_sample")]
    [string]$ReplayMode,
    [switch]$FailOnViolation
)

$ErrorActionPreference = "Stop"

function Get-PythonLauncher {
    foreach ($candidate in @("python", "python3", "py")) {
        $command = Get-Command $candidate -ErrorAction SilentlyContinue
        if ($null -ne $command) {
            return $candidate
        }
    }

    throw "python is required for scripts/run-radish-docs-qa-negative-regression.ps1"
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$scriptPath = Join-Path $repoRoot "scripts/run-eval-regression.py"

if (-not (Test-Path -LiteralPath $scriptPath -PathType Leaf)) {
    throw "missing python runner: scripts/run-eval-regression.py"
}

$pythonLauncher = Get-PythonLauncher
$arguments = @()
if ($pythonLauncher -eq "py") {
    $arguments += "-3"
}
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

if (-not [string]::IsNullOrWhiteSpace($ReplayMode)) {
    $arguments += "--replay-mode"
    $arguments += $ReplayMode
}

if ($FailOnViolation) {
    $arguments += "--fail-on-violation"
}

& $pythonLauncher @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
