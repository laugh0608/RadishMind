[CmdletBinding()]
param(
    [string]$SampleDir,
    [string[]]$SamplePaths,
    [switch]$FailOnViolation
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
if ([string]::IsNullOrWhiteSpace($SampleDir)) {
    $SampleDir = Join-Path $repoRoot "datasets/eval/radish"
}

$sampleSchemaPath = Join-Path $repoRoot "datasets/eval/radish-task-sample.schema.json"
$requestSchemaPath = Join-Path $repoRoot "contracts/copilot-request.schema.json"
$responseSchemaPath = Join-Path $repoRoot "contracts/copilot-response.schema.json"
$taskCardPath = Join-Path $repoRoot "docs/task-cards/radish-answer-docs-question.md"
$allowedPrimaryKinds = @("markdown", "text")
$unofficialSourceTypes = @("faq", "forum")

function Add-Violation {
    param(
        [System.Collections.Generic.List[string]]$Violations,
        [string]$Message
    )

    $Violations.Add($Message)
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

function Try-GetPropertyValue {
    param(
        $Object,
        [string]$Name,
        [ref]$Value
    )

    if ($null -eq $Object) {
        return $false
    }

    if ($Object -is [System.Collections.IDictionary]) {
        if ($Object.Contains($Name)) {
            $Value.Value = $Object[$Name]
            return $true
        }

        return $false
    }

    $property = $Object.PSObject.Properties[$Name]
    if ($null -eq $property) {
        return $false
    }

    $Value.Value = $property.Value
    return $true
}

function Convert-ExpectationLiteral {
    param([string]$Literal)

    $trimmed = $Literal.Trim()

    if ($trimmed -match '^(?i:true|false)$') {
        return [System.Convert]::ToBoolean($trimmed)
    }

    if ($trimmed -match '^(?i:null)$') {
        return $null
    }

    if ($trimmed -match '^-?\d+$') {
        return [long]$trimmed
    }

    if ($trimmed -match '^-?\d+\.\d+$') {
        return [double]$trimmed
    }

    if ($trimmed.Length -ge 2 -and $trimmed.StartsWith('"') -and $trimmed.EndsWith('"')) {
        return $trimmed.Substring(1, $trimmed.Length - 2)
    }

    return $trimmed
}

function Resolve-JsonPath {
    param(
        $Document,
        [string]$Path
    )

    if ([string]::IsNullOrWhiteSpace($Path) -or -not $Path.StartsWith('$')) {
        throw "unsupported json path: $Path"
    }

    $cursor = $Document
    $remaining = $Path.Substring(1)

    while ($remaining.Length -gt 0) {
        if ($remaining -match '^\.([A-Za-z0-9_]+)(.*)$') {
            $propertyName = $Matches[1]
            $remaining = $Matches[2]
            $value = $null
            if (-not (Try-GetPropertyValue -Object $cursor -Name $propertyName -Value ([ref]$value))) {
                return [pscustomobject]@{
                    Exists = $false
                    Value = $null
                }
            }

            $cursor = $value
            continue
        }

        if ($remaining -match '^\[(\d+)\](.*)$') {
            $index = [int]$Matches[1]
            $remaining = $Matches[2]
            $items = @(Get-Array $cursor)
            if ($index -ge $items.Count) {
                return [pscustomobject]@{
                    Exists = $false
                    Value = $null
                }
            }

            $cursor = $items[$index]
            continue
        }

        throw "unsupported json path: $Path"
    }

    return [pscustomobject]@{
        Exists = $true
        Value = $cursor
    }
}

function Test-PathExpectations {
    param(
        $Document,
        [string[]]$Expressions,
        [bool]$ShouldExist,
        [string]$SampleName,
        [System.Collections.Generic.List[string]]$Violations
    )

    foreach ($expression in (Get-Array $Expressions)) {
        $text = [string]$expression
        if ([string]::IsNullOrWhiteSpace($text)) {
            continue
        }

        $parts = $text -split '=', 2
        $path = $parts[0]
        $hasExpectedValue = $parts.Count -eq 2
        $expectedValue = $null
        if ($hasExpectedValue) {
            $expectedValue = Convert-ExpectationLiteral -Literal $parts[1]
        }

        $resolved = Resolve-JsonPath -Document $Document -Path $path
        if ($ShouldExist) {
            if (-not $resolved.Exists) {
                Add-Violation -Violations $Violations -Message ("{0}: missing required json path '{1}'" -f $SampleName, $text)
                continue
            }

            if ($hasExpectedValue -and $resolved.Value -ne $expectedValue) {
                Add-Violation -Violations $Violations -Message ("{0}: json path '{1}' expected '{2}' but found '{3}'" -f $SampleName, $path, $expectedValue, $resolved.Value)
            }

            continue
        }

        if (-not $hasExpectedValue) {
            if ($resolved.Exists) {
                Add-Violation -Violations $Violations -Message ("{0}: json path '{1}' should not exist" -f $SampleName, $path)
            }

            continue
        }

        if ($resolved.Exists -and $resolved.Value -eq $expectedValue) {
            Add-Violation -Violations $Violations -Message ("{0}: json path '{1}' should not equal '{2}'" -f $SampleName, $path, $expectedValue)
        }
    }
}

function Test-DocumentAgainstSchema {
    param(
        $Document,
        [string]$SchemaPath,
        [string]$DocumentName,
        [System.Collections.Generic.List[string]]$Violations
    )

    try {
        $json = $Document | ConvertTo-Json -Depth 100
    }
    catch {
        Add-Violation -Violations $Violations -Message ("{0}: failed to serialize for schema validation: {1}" -f $DocumentName, $_.Exception.Message)
        return
    }

    try {
        $isValid = $json | Test-Json -SchemaFile $SchemaPath -WarningAction SilentlyContinue -ErrorAction Stop
        if (-not $isValid) {
            Add-Violation -Violations $Violations -Message ("{0}: does not match schema '{1}'" -f $DocumentName, (Split-Path -Leaf $SchemaPath))
        }
    }
    catch {
        Add-Violation -Violations $Violations -Message ("{0}: schema validation failed against '{1}': {2}" -f $DocumentName, (Split-Path -Leaf $SchemaPath), $_.Exception.Message)
    }
}

function Test-RetrievalContract {
    param(
        $Sample,
        [string]$SampleName,
        [System.Collections.Generic.List[string]]$Violations
    )

    $request = $Sample.input_request
    $context = $request.context
    $retrieval = $Sample.retrieval_expectations
    $artifacts = @(Get-Array $request.artifacts)
    $primaryArtifacts = @($artifacts | Where-Object { $_.role -eq "primary" })
    $supportingArtifacts = @($artifacts | Where-Object { $_.role -eq "supporting" })
    $referenceArtifacts = @($artifacts | Where-Object { $_.role -eq "reference" })
    $searchScopes = @(Get-Array $context.search_scope | ForEach-Object { [string]$_ })

    if ($retrieval.require_primary_artifact -and $primaryArtifacts.Count -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: retrieval_expectations require a primary artifact" -f $SampleName)
    }

    if ($primaryArtifacts.Count -gt [int]$retrieval.max_primary_artifacts) {
        Add-Violation -Violations $Violations -Message ("{0}: primary artifact count exceeds retrieval_expectations.max_primary_artifacts" -f $SampleName)
    }

    if ($supportingArtifacts.Count -gt [int]$retrieval.max_supporting_artifacts) {
        Add-Violation -Violations $Violations -Message ("{0}: supporting artifact count exceeds retrieval_expectations.max_supporting_artifacts" -f $SampleName)
    }

    if ($referenceArtifacts.Count -gt [int]$retrieval.max_reference_artifacts) {
        Add-Violation -Violations $Violations -Message ("{0}: reference artifact count exceeds retrieval_expectations.max_reference_artifacts" -f $SampleName)
    }

    foreach ($scope in $searchScopes) {
        if ((@(Get-Array $retrieval.allowed_search_scopes) | ForEach-Object { [string]$_ }) -notcontains $scope) {
            Add-Violation -Violations $Violations -Message ("{0}: search_scope '{1}' is not allowed by retrieval_expectations" -f $SampleName, $scope)
        }
    }

    foreach ($requiredScope in (Get-Array $retrieval.required_search_scopes | ForEach-Object { [string]$_ })) {
        if ($searchScopes -notcontains $requiredScope) {
            Add-Violation -Violations $Violations -Message ("{0}: search_scope is missing required scope '{1}'" -f $SampleName, $requiredScope)
        }
    }

    foreach ($disallowedScope in (Get-Array $retrieval.disallowed_search_scopes | ForEach-Object { [string]$_ })) {
        if ($searchScopes -contains $disallowedScope) {
            Add-Violation -Violations $Violations -Message ("{0}: search_scope contains disallowed scope '{1}'" -f $SampleName, $disallowedScope)
        }
    }

    $resourceSlug = [string]$context.resource.slug

    foreach ($artifact in $artifacts) {
        $role = [string]$artifact.role
        $metadata = $artifact.metadata
        $sourceType = [string]$metadata.source_type
        $pageSlug = [string]$metadata.page_slug
        $fragmentId = [string]$metadata.fragment_id
        $retrievalRank = $metadata.retrieval_rank
        $isOfficial = $metadata.is_official

        if ($role -eq "primary" -and $allowedPrimaryKinds -notcontains [string]$artifact.kind) {
            Add-Violation -Violations $Violations -Message ("{0}: primary artifact kind must be markdown or text" -f $SampleName)
        }

        if ($retrieval.require_artifact_source_metadata) {
            if ([string]::IsNullOrWhiteSpace($sourceType)) {
                Add-Violation -Violations $Violations -Message ("{0}: artifact '{1}' is missing metadata.source_type" -f $SampleName, $artifact.name)
            }

            if ([string]::IsNullOrWhiteSpace($pageSlug)) {
                Add-Violation -Violations $Violations -Message ("{0}: artifact '{1}' is missing metadata.page_slug" -f $SampleName, $artifact.name)
            }

            if ([string]::IsNullOrWhiteSpace($fragmentId)) {
                Add-Violation -Violations $Violations -Message ("{0}: artifact '{1}' is missing metadata.fragment_id" -f $SampleName, $artifact.name)
            }

            if ($null -eq $retrievalRank -or [int]$retrievalRank -lt 1) {
                Add-Violation -Violations $Violations -Message ("{0}: artifact '{1}' must carry metadata.retrieval_rank >= 1" -f $SampleName, $artifact.name)
            }

            if ($null -eq $isOfficial) {
                Add-Violation -Violations $Violations -Message ("{0}: artifact '{1}' is missing metadata.is_official" -f $SampleName, $artifact.name)
            }
        }

        $allowedSourceTypes = switch ($role) {
            "primary" { @(Get-Array $retrieval.allowed_primary_source_types | ForEach-Object { [string]$_ }) }
            "supporting" { @(Get-Array $retrieval.allowed_supporting_source_types | ForEach-Object { [string]$_ }) }
            "reference" { @(Get-Array $retrieval.allowed_reference_source_types | ForEach-Object { [string]$_ }) }
            default { @() }
        }

        if ($allowedSourceTypes.Count -gt 0 -and $allowedSourceTypes -notcontains $sourceType) {
            Add-Violation -Violations $Violations -Message ("{0}: artifact '{1}' source_type '{2}' is not allowed for role '{3}'" -f $SampleName, $artifact.name, $sourceType, $role)
        }

        if ($retrieval.require_primary_resource_match -and $role -eq "primary" -and -not [string]::IsNullOrWhiteSpace($resourceSlug) -and $pageSlug -ne $resourceSlug) {
            Add-Violation -Violations $Violations -Message ("{0}: primary artifact '{1}' must match context.resource.slug" -f $SampleName, $artifact.name)
        }

        if ($unofficialSourceTypes -contains $sourceType) {
            if (-not $retrieval.allow_unofficial_sources) {
                Add-Violation -Violations $Violations -Message ("{0}: unofficial source_type '{1}' is not allowed by retrieval_expectations" -f $SampleName, $sourceType)
            }

            if ($role -eq "primary") {
                Add-Violation -Violations $Violations -Message ("{0}: unofficial source_type '{1}' cannot be used as the primary artifact" -f $SampleName, $sourceType)
            }

            if ($isOfficial -ne $false) {
                Add-Violation -Violations $Violations -Message ("{0}: unofficial source_type '{1}' must set metadata.is_official to false" -f $SampleName, $sourceType)
            }

            continue
        }

        if ($retrieval.require_artifact_source_metadata -and $null -ne $isOfficial -and $isOfficial -ne $true) {
            Add-Violation -Violations $Violations -Message ("{0}: official source artifact '{1}' must set metadata.is_official to true" -f $SampleName, $artifact.name)
        }
    }
}

function Test-ResponseContract {
    param(
        $Sample,
        [string]$SampleName,
        [System.Collections.Generic.List[string]]$Violations
    )

    $response = $Sample.golden_response
    $shape = $Sample.expected_response_shape
    $evaluation = $Sample.evaluation
    $answers = @(Get-Array $response.answers)
    $issues = @(Get-Array $response.issues)
    $actions = @(Get-Array $response.proposed_actions)
    $citations = @(Get-Array $response.citations)

    if ($response.project -ne "radish") {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response.project must be 'radish'" -f $SampleName)
    }

    if ($response.task -ne "answer_docs_question") {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response.task must be 'answer_docs_question'" -f $SampleName)
    }

    if ([string]$response.status -ne [string]$shape.status) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response.status does not match expected_response_shape.status" -f $SampleName)
    }

    if ([string]$response.risk_level -ne [string]$evaluation.expected_risk_level) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response.risk_level does not match evaluation.expected_risk_level" -f $SampleName)
    }

    if ($shape.requires_summary -and [string]::IsNullOrWhiteSpace([string]$response.summary)) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response.summary is required" -f $SampleName)
    }

    if ($shape.requires_answers -and $answers.Count -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response must contain at least 1 answer" -f $SampleName)
    }

    if ($shape.requires_issues -and $issues.Count -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response must contain at least 1 issue" -f $SampleName)
    }

    if (-not $shape.requires_issues -and $issues.Count -gt 0) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response should not contain issues" -f $SampleName)
    }

    if ($shape.requires_citations -and $citations.Count -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response must contain at least 1 citation" -f $SampleName)
    }

    if (-not $shape.allow_proposed_actions -and $actions.Count -gt 0) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response should not contain proposed_actions" -f $SampleName)
    }

    foreach ($requiredKind in (Get-Array $shape.required_action_kinds | ForEach-Object { [string]$_ })) {
        $actualKinds = @(Get-Array $actions | ForEach-Object { [string]$_.kind })
        if ($actualKinds -notcontains $requiredKind) {
            Add-Violation -Violations $Violations -Message ("{0}: golden_response is missing required action kind '{1}'" -f $SampleName, $requiredKind)
        }
    }

    foreach ($action in $actions) {
        if ([string]$action.risk_level -ne "low") {
            Add-Violation -Violations $Violations -Message ("{0}: proposed action '{1}' must remain low risk for answer_docs_question" -f $SampleName, $action.title)
        }

        if ($action.requires_confirmation -ne $false) {
            Add-Violation -Violations $Violations -Message ("{0}: proposed action '{1}' must not require confirmation" -f $SampleName, $action.title)
        }
    }

    $citationIds = @(Get-Array $citations | ForEach-Object { [string]$_.id })
    $referencedCitationIds = @()
    foreach ($answer in $answers) {
        $referencedCitationIds += @(Get-Array $answer.citation_ids | ForEach-Object { [string]$_ })
    }
    foreach ($issue in $issues) {
        $referencedCitationIds += @(Get-Array $issue.citation_ids | ForEach-Object { [string]$_ })
    }
    foreach ($action in $actions) {
        $referencedCitationIds += @(Get-Array $action.citation_ids | ForEach-Object { [string]$_ })
    }

    foreach ($citationId in ($referencedCitationIds | Sort-Object -Unique)) {
        if ($citationIds -notcontains $citationId) {
            Add-Violation -Violations $Violations -Message ("{0}: referenced citation id '{1}' is missing from golden_response.citations" -f $SampleName, $citationId)
        }
    }

    Test-PathExpectations -Document $response -Expressions (Get-Array $evaluation.must_have_json_paths) -ShouldExist $true -SampleName $SampleName -Violations $Violations
    Test-PathExpectations -Document $response -Expressions (Get-Array $evaluation.must_not_have_json_paths) -ShouldExist $false -SampleName $SampleName -Violations $Violations
}

foreach ($requiredPath in @($sampleSchemaPath, $requestSchemaPath, $responseSchemaPath, $taskCardPath)) {
    if (-not (Test-Path -LiteralPath $requiredPath -PathType Leaf)) {
        throw "missing required file: $requiredPath"
    }
}

if ($SamplePaths.Count -gt 0) {
    $sampleFiles = foreach ($path in $SamplePaths) {
        if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
            throw "missing sample file: $path"
        }

        Get-Item -LiteralPath $path
    }
}
else {
    if (-not (Test-Path -LiteralPath $SampleDir -PathType Container)) {
        throw "missing sample directory: $SampleDir"
    }

    $sampleFiles = Get-ChildItem -LiteralPath $SampleDir -Filter *.json | Sort-Object Name
}

if ($sampleFiles.Count -eq 0) {
    throw "no sample files found for Radish docs QA regression"
}

$allViolations = New-Object System.Collections.Generic.List[string]

foreach ($sampleFile in $sampleFiles) {
    $sampleName = $sampleFile.Name
    $violations = New-Object System.Collections.Generic.List[string]

    try {
        $sample = Get-Content -LiteralPath $sampleFile.FullName -Raw | ConvertFrom-Json
    }
    catch {
        Add-Violation -Violations $violations -Message ("{0}: failed to parse json: {1}" -f $sampleName, $_.Exception.Message)
        $sample = $null
    }

    if ($null -ne $sample) {
        Test-DocumentAgainstSchema -Document $sample -SchemaPath $sampleSchemaPath -DocumentName ("{0} sample" -f $sampleName) -Violations $violations
        Test-DocumentAgainstSchema -Document $sample.input_request -SchemaPath $requestSchemaPath -DocumentName ("{0} input_request" -f $sampleName) -Violations $violations
        Test-DocumentAgainstSchema -Document $sample.golden_response -SchemaPath $responseSchemaPath -DocumentName ("{0} golden_response" -f $sampleName) -Violations $violations
        Test-RetrievalContract -Sample $sample -SampleName $sampleName -Violations $violations
        Test-ResponseContract -Sample $sample -SampleName $sampleName -Violations $violations
    }

    if ($violations.Count -gt 0) {
        Write-Host ("FAIL {0}" -f $sampleName)
        foreach ($violation in $violations) {
            Write-Host ("  - {0}" -f $violation)
            $allViolations.Add($violation)
        }
        continue
    }

    Write-Host ("PASS {0}" -f $sampleName)
}

if ($allViolations.Count -gt 0) {
    if ($FailOnViolation) {
        exit 1
    }

    Write-Warning ("radish docs qa regression found {0} violation(s)." -f $allViolations.Count)
    return
}

Write-Host "radish docs qa regression passed."
