[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$venvPythonWindows = Join-Path $repoRoot ".venv/Scripts/python.exe"
$venvPythonUnix = Join-Path $repoRoot ".venv/bin/python"

function Get-RepoVenvPython {
    if (Test-Path -LiteralPath $venvPythonWindows -PathType Leaf) {
        return $venvPythonWindows
    }
    if (Test-Path -LiteralPath $venvPythonUnix -PathType Leaf) {
        return $venvPythonUnix
    }
    return $null
}

function Get-SystemPython {
    foreach ($candidate in @("python3", "python", "py")) {
        $command = Get-Command $candidate -ErrorAction SilentlyContinue
        if ($null -ne $command) {
            return $candidate
        }
    }
    throw "python3, python or py is required to create .venv"
}

$venvPython = Get-RepoVenvPython
if ($null -eq $venvPython) {
    $pythonLauncher = Get-SystemPython
    if ($pythonLauncher -eq "py") {
        & $pythonLauncher -3 -m venv (Join-Path $repoRoot ".venv")
    } else {
        & $pythonLauncher -m venv (Join-Path $repoRoot ".venv")
    }
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
    $venvPython = Get-RepoVenvPython
}

if ($null -eq $venvPython) {
    throw "failed to create repository virtual environment at .venv"
}

& $venvPython -m pip install --upgrade pip
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

& $venvPython -m pip install -r (Join-Path $repoRoot "requirements-dev.txt")
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "RadishMind development virtual environment is ready: $venvPython"
