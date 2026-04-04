[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$BatchArtifactSummary,
    [int]$Top = 1,
    [string]$SampleDir,
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

    throw "python is required for scripts/run-radish-docs-qa-negative-recommended.ps1"
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$scriptPath = Join-Path $repoRoot "scripts/run-radish-docs-qa-negative-recommended.py"

if (-not (Test-Path -LiteralPath $scriptPath -PathType Leaf)) {
    throw "missing python runner: scripts/run-radish-docs-qa-negative-recommended.py"
}

$pythonLauncher = Get-PythonLauncher
$arguments = @()
if ($pythonLauncher -eq "py") {
    $arguments += "-3"
}
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

& $pythonLauncher @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
