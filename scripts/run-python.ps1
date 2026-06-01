[CmdletBinding()]
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$PythonArgs
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$venvPythonWindows = Join-Path $repoRoot ".venv/Scripts/python.exe"
$venvPythonUnix = Join-Path $repoRoot ".venv/bin/python"

if (Test-Path -LiteralPath $venvPythonWindows -PathType Leaf) {
    $venvPython = $venvPythonWindows
} elseif (Test-Path -LiteralPath $venvPythonUnix -PathType Leaf) {
    $venvPython = $venvPythonUnix
} else {
    throw "missing repository virtual environment. Run pwsh ./scripts/bootstrap-dev.ps1 before running Python-based scripts"
}

& $venvPython @PythonArgs
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
