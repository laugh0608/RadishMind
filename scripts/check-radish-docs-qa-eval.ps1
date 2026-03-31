[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$sampleDir = Join-Path $repoRoot "datasets/eval/radish"
$schemaPath = Join-Path $repoRoot "datasets/eval/radish-task-sample.schema.json"
$expectedSamples = @(
    "answer-docs-question-direct-answer-001.json",
    "answer-docs-question-evidence-gap-001.json",
    "answer-docs-question-navigation-001.json"
)
$allowedSearchScopes = @("docs", "wiki", "faq", "forum", "console", "attachments")
$allowedPrimaryKinds = @("markdown", "text")
$violations = New-Object System.Collections.Generic.List[string]

function Add-Violation {
    param([string]$Message)

    $violations.Add($Message)
}

function Get-Array {
    param($Value)

    if ($null -eq $Value) {
        return @()
    }

    if ($Value -is [System.Array]) {
        return $Value
    }

    return @($Value)
}

if (-not (Test-Path -LiteralPath $schemaPath -PathType Leaf)) {
    throw "missing schema file: datasets/eval/radish-task-sample.schema.json"
}

$null = Get-Content -Path $schemaPath -Raw | ConvertFrom-Json

if (-not (Test-Path -LiteralPath $sampleDir -PathType Container)) {
    throw "missing sample directory: datasets/eval/radish"
}

$sampleFiles = Get-ChildItem -Path $sampleDir -Filter *.json | Sort-Object Name
if ($sampleFiles.Count -ne $expectedSamples.Count) {
    Add-Violation("datasets/eval/radish: expected {0} sample files, found {1}" -f $expectedSamples.Count, $sampleFiles.Count)
}

foreach ($expected in $expectedSamples) {
    $fullPath = Join-Path $sampleDir $expected
    if (-not (Test-Path -LiteralPath $fullPath -PathType Leaf)) {
        Add-Violation("missing expected sample: datasets/eval/radish/{0}" -f $expected)
        continue
    }

    $sample = Get-Content -Path $fullPath -Raw | ConvertFrom-Json

    if ($sample.project -ne "radish") {
        Add-Violation("{0}: project must be 'radish'" -f $expected)
    }

    if ($sample.task -ne "answer_docs_question") {
        Add-Violation("{0}: task must be 'answer_docs_question'" -f $expected)
    }

    if ($sample.golden_response.project -ne "radish") {
        Add-Violation("{0}: golden_response.project must be 'radish'" -f $expected)
    }

    if ($sample.golden_response.task -ne "answer_docs_question") {
        Add-Violation("{0}: golden_response.task must be 'answer_docs_question'" -f $expected)
    }

    if ($sample.golden_response.risk_level -ne $sample.evaluation.expected_risk_level) {
        Add-Violation("{0}: golden_response.risk_level does not match evaluation.expected_risk_level" -f $expected)
    }

    if ($sample.golden_response.status -ne $sample.expected_response_shape.status) {
        Add-Violation("{0}: golden_response.status does not match expected_response_shape.status" -f $expected)
    }

    $artifacts = Get-Array $sample.input_request.artifacts
    $primaryArtifacts = @($artifacts | Where-Object { $_.role -eq "primary" })
    $supportingArtifacts = @($artifacts | Where-Object { $_.role -eq "supporting" })
    $referenceArtifacts = @($artifacts | Where-Object { $_.role -eq "reference" })

    if ($primaryArtifacts.Count -ne 1) {
        Add-Violation("{0}: must contain exactly 1 primary artifact" -f $expected)
    }

    if ($supportingArtifacts.Count -gt 3) {
        Add-Violation("{0}: must contain at most 3 supporting artifacts" -f $expected)
    }

    if ($referenceArtifacts.Count -gt 2) {
        Add-Violation("{0}: must contain at most 2 reference artifacts" -f $expected)
    }

    foreach ($artifact in $primaryArtifacts) {
        if ($allowedPrimaryKinds -notcontains $artifact.kind) {
            Add-Violation("{0}: primary artifact kind must be markdown or text" -f $expected)
        }
    }

    $context = $sample.input_request.context
    if ($null -eq $context.current_app -or $null -eq $context.route -or $null -eq $context.resource) {
        Add-Violation("{0}: input_request.context must include current_app, route and resource" -f $expected)
    }

    foreach ($scope in (Get-Array $context.search_scope)) {
        if ($allowedSearchScopes -notcontains [string]$scope) {
            Add-Violation("{0}: unsupported search_scope '{1}'" -f $expected, $scope)
        }
    }

    $answers = Get-Array $sample.golden_response.answers
    $issues = Get-Array $sample.golden_response.issues
    $proposedActions = Get-Array $sample.golden_response.proposed_actions
    $citations = Get-Array $sample.golden_response.citations

    if ($sample.expected_response_shape.requires_answers -and $answers.Count -lt 1) {
        Add-Violation("{0}: golden_response must contain at least 1 answer" -f $expected)
    }

    if ($sample.expected_response_shape.requires_issues -and $issues.Count -lt 1) {
        Add-Violation("{0}: golden_response must contain at least 1 issue" -f $expected)
    }

    if (-not $sample.expected_response_shape.requires_issues -and $issues.Count -gt 0) {
        Add-Violation("{0}: golden_response should not contain issues" -f $expected)
    }

    if ($sample.expected_response_shape.requires_citations -and $citations.Count -lt 1) {
        Add-Violation("{0}: golden_response must contain at least 1 citation" -f $expected)
    }

    if (-not $sample.expected_response_shape.allow_proposed_actions -and $proposedActions.Count -gt 0) {
        Add-Violation("{0}: golden_response should not contain proposed_actions" -f $expected)
    }

    if ($sample.expected_response_shape.required_action_kinds) {
        $requiredKinds = Get-Array $sample.expected_response_shape.required_action_kinds
        $actualKinds = @(Get-Array $sample.golden_response.proposed_actions | ForEach-Object { [string]$_.kind })
        foreach ($requiredKind in $requiredKinds) {
            if ($actualKinds -notcontains [string]$requiredKind) {
                Add-Violation("{0}: golden_response is missing required action kind '{1}'" -f $expected, $requiredKind)
            }
        }
    }

    $citationIds = @(Get-Array $sample.golden_response.citations | ForEach-Object { [string]$_.id })
    $referencedCitationIds = @()
    foreach ($answer in $answers) {
        $referencedCitationIds += @(Get-Array $answer.citation_ids | ForEach-Object { [string]$_ })
    }
    foreach ($issue in $issues) {
        $referencedCitationIds += @(Get-Array $issue.citation_ids | ForEach-Object { [string]$_ })
    }
    foreach ($action in $proposedActions) {
        $referencedCitationIds += @(Get-Array $action.citation_ids | ForEach-Object { [string]$_ })
    }

    foreach ($citationId in ($referencedCitationIds | Sort-Object -Unique)) {
        if ($citationIds -notcontains $citationId) {
            Add-Violation("{0}: referenced citation id '{1}' is missing from golden_response.citations" -f $expected, $citationId)
        }
    }
}

if ($violations.Count -gt 0) {
    $violations | ForEach-Object { Write-Error $_ }
    exit 1
}

Write-Host "radish docs qa eval checks passed."
