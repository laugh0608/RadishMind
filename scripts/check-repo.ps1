[CmdletBinding()]
param(
    [switch]$SkipTextFiles,
    [switch]$Fast
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

function Get-PythonLauncher {
    $venvPython = Join-Path $repoRoot ".venv/Scripts/python.exe"
    if (Test-Path -LiteralPath $venvPython -PathType Leaf) {
        return $venvPython
    }

    foreach ($candidate in @("python", "python3", "py")) {
        $command = Get-Command $candidate -ErrorAction SilentlyContinue
        if ($null -ne $command) {
            return $candidate
        }
    }

    throw "python is required for scripts/check-repo.ps1"
}

$scriptPath = Join-Path $repoRoot "scripts/check-repo.py"

if (-not (Test-Path -LiteralPath $scriptPath -PathType Leaf)) {
    throw "missing python checker: scripts/check-repo.py"
}

$pythonLauncher = Get-PythonLauncher
$arguments = @()
if ($pythonLauncher -eq "py") {
    $arguments += "-3"
}
$arguments += $scriptPath

if ($SkipTextFiles) {
    $arguments += "--skip-text-files"
}

if ($Fast) {
    $arguments += "--fast"
}

& $pythonLauncher @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
