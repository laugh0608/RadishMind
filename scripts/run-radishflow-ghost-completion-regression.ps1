[CmdletBinding()]
param(
    [string]$SampleDir,
    [string[]]$SamplePaths,
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

    throw "python is required for scripts/run-radishflow-ghost-completion-regression.ps1"
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
$arguments += "radishflow-ghost-completion"

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

& $pythonLauncher @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
