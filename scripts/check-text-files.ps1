[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

function Get-PythonLauncher {
    foreach ($candidate in @("python", "python3", "py")) {
        $command = Get-Command $candidate -ErrorAction SilentlyContinue
        if ($null -ne $command) {
            return $candidate
        }
    }

    throw "python is required for scripts/check-text-files.ps1"
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$scriptPath = Join-Path $repoRoot "scripts/check-text-files.py"

if (-not (Test-Path -LiteralPath $scriptPath -PathType Leaf)) {
    throw "missing python checker: scripts/check-text-files.py"
}

$pythonLauncher = Get-PythonLauncher
$arguments = @()
if ($pythonLauncher -eq "py") {
    $arguments += "-3"
}
$arguments += $scriptPath

& $pythonLauncher @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
