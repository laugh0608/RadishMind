[CmdletBinding()]
param(
    [switch]$SkipTextFiles,
    [switch]$Fast
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

function Get-RepoVenvPython {
    $venvPythonWindows = Join-Path $repoRoot ".venv/Scripts/python.exe"
    if (Test-Path -LiteralPath $venvPythonWindows -PathType Leaf) {
        return $venvPythonWindows
    }

    $venvPythonUnix = Join-Path $repoRoot ".venv/bin/python"
    if (Test-Path -LiteralPath $venvPythonUnix -PathType Leaf) {
        return $venvPythonUnix
    }

    throw "missing repository virtual environment. Run pwsh ./scripts/bootstrap-dev.ps1 before running scripts/check-repo.ps1"
}

$scriptPath = Join-Path $repoRoot "scripts/check-repo.py"

if (-not (Test-Path -LiteralPath $scriptPath -PathType Leaf)) {
    throw "missing python checker: scripts/check-repo.py"
}

$pythonLauncher = Get-RepoVenvPython
$arguments = @()
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
