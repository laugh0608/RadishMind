[CmdletBinding()]
param(
    [string]$SampleDir,
    [string]$SamplePattern,
    [ValidateSet("mock", "openai-compatible")]
    [string]$Provider = "openai-compatible",
    [string]$Model,
    [string]$BaseUrl,
    [string]$ApiKey,
    [double]$Temperature = 0.0,
    [Parameter(Mandatory = $true)]
    [string]$OutputRoot,
    [Parameter(Mandatory = $true)]
    [string]$CollectionBatch,
    [string]$CaptureOrigin,
    [string[]]$CaptureTag,
    [string]$CaptureNotes,
    [int]$MaxAttempts = 3,
    [double]$RetryBaseDelaySeconds = 5.0,
    [string]$ManifestOutput,
    [string]$ManifestDescription,
    [switch]$Resume,
    [switch]$ContinueOnError,
    [switch]$SkipAudit,
    [string]$AuditReportOutput,
    [string]$ReplayIndexOutput,
    [switch]$BuildNegativeReplay,
    [string]$NegativeOutputDir,
    [switch]$FailOnAuditViolation
)

$ErrorActionPreference = "Stop"

function Get-PythonLauncher {
    foreach ($candidate in @("python", "python3", "py")) {
        $command = Get-Command $candidate -ErrorAction SilentlyContinue
        if ($null -ne $command) {
            return $candidate
        }
    }

    throw "python is required for scripts/run-radish-docs-qa-real-batch.ps1"
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$scriptPath = Join-Path $repoRoot "scripts/run-radish-docs-qa-real-batch.py"

if (-not (Test-Path -LiteralPath $scriptPath -PathType Leaf)) {
    throw "missing python runner: scripts/run-radish-docs-qa-real-batch.py"
}

$pythonLauncher = Get-PythonLauncher
$arguments = @()
if ($pythonLauncher -eq "py") {
    $arguments += "-3"
}
$arguments += $scriptPath
$arguments += "--provider"
$arguments += $Provider
$arguments += "--temperature"
$arguments += $Temperature.ToString([System.Globalization.CultureInfo]::InvariantCulture)
$arguments += "--output-root"
$arguments += $OutputRoot
$arguments += "--collection-batch"
$arguments += $CollectionBatch

if (-not [string]::IsNullOrWhiteSpace($SampleDir)) {
    $arguments += "--sample-dir"
    $arguments += $SampleDir
}

if (-not [string]::IsNullOrWhiteSpace($SamplePattern)) {
    $arguments += "--sample-pattern"
    $arguments += $SamplePattern
}

if (-not [string]::IsNullOrWhiteSpace($Model)) {
    $arguments += "--model"
    $arguments += $Model
}

if (-not [string]::IsNullOrWhiteSpace($BaseUrl)) {
    $arguments += "--base-url"
    $arguments += $BaseUrl
}

if (-not [string]::IsNullOrWhiteSpace($ApiKey)) {
    $arguments += "--api-key"
    $arguments += $ApiKey
}

if (-not [string]::IsNullOrWhiteSpace($CaptureOrigin)) {
    $arguments += "--capture-origin"
    $arguments += $CaptureOrigin
}

if ($CaptureTag.Count -gt 0) {
    foreach ($item in $CaptureTag) {
        $arguments += "--capture-tag"
        $arguments += $item
    }
}

if (-not [string]::IsNullOrWhiteSpace($CaptureNotes)) {
    $arguments += "--capture-notes"
    $arguments += $CaptureNotes
}

if ($MaxAttempts -ne 3) {
    $arguments += "--max-attempts"
    $arguments += $MaxAttempts
}

if ($RetryBaseDelaySeconds -ne 5.0) {
    $arguments += "--retry-base-delay-seconds"
    $arguments += $RetryBaseDelaySeconds.ToString([System.Globalization.CultureInfo]::InvariantCulture)
}

if (-not [string]::IsNullOrWhiteSpace($ManifestOutput)) {
    $arguments += "--manifest-output"
    $arguments += $ManifestOutput
}

if (-not [string]::IsNullOrWhiteSpace($ManifestDescription)) {
    $arguments += "--manifest-description"
    $arguments += $ManifestDescription
}

if ($Resume) {
    $arguments += "--resume"
}

if ($ContinueOnError) {
    $arguments += "--continue-on-error"
}

if ($SkipAudit) {
    $arguments += "--skip-audit"
}

if (-not [string]::IsNullOrWhiteSpace($AuditReportOutput)) {
    $arguments += "--audit-report-output"
    $arguments += $AuditReportOutput
}

if (-not [string]::IsNullOrWhiteSpace($ReplayIndexOutput)) {
    $arguments += "--replay-index-output"
    $arguments += $ReplayIndexOutput
}

if ($BuildNegativeReplay) {
    $arguments += "--build-negative-replay"
}

if (-not [string]::IsNullOrWhiteSpace($NegativeOutputDir)) {
    $arguments += "--negative-output-dir"
    $arguments += $NegativeOutputDir
}

if ($FailOnAuditViolation) {
    $arguments += "--fail-on-audit-violation"
}

& $pythonLauncher @arguments
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
