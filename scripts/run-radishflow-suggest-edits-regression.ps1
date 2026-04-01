[CmdletBinding()]
param(
    [string]$SampleDir,
    [string[]]$SamplePaths,
    [switch]$FailOnViolation
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
if ([string]::IsNullOrWhiteSpace($SampleDir)) {
    $SampleDir = Join-Path $repoRoot "datasets/eval/radishflow"
}

$sampleSchemaPath = Join-Path $repoRoot "datasets/eval/radishflow-task-sample.schema.json"
$requestSchemaPath = Join-Path $repoRoot "contracts/copilot-request.schema.json"
$responseSchemaPath = Join-Path $repoRoot "contracts/copilot-response.schema.json"
$taskCardPath = Join-Path $repoRoot "docs/task-cards/radishflow-suggest-flowsheet-edits.md"

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

function Get-RiskRank {
    param([string]$RiskLevel)

    switch ($RiskLevel) {
        "low" { return 1 }
        "medium" { return 2 }
        "high" { return 3 }
        default { return 0 }
    }
}

function Get-PropertyCount {
    param($Object)

    if ($null -eq $Object) {
        return 0
    }

    if ($Object -is [System.Collections.IDictionary]) {
        return $Object.Count
    }

    return @($Object.PSObject.Properties).Count
}

function Test-RequestContract {
    param(
        $Sample,
        [string]$SampleName,
        [System.Collections.Generic.List[string]]$Violations
    )

    $request = $Sample.input_request
    $context = $request.context
    $artifacts = @(Get-Array $request.artifacts)
    $primaryArtifacts = @($artifacts | Where-Object { $_.role -eq "primary" })
    $diagnostics = @(Get-Array $context.diagnostics)
    $selectedUnitIds = @(Get-Array $context.selected_unit_ids)
    $selectedStreamIds = @(Get-Array $context.selected_stream_ids)

    if ($request.project -ne "radishflow") {
        Add-Violation -Violations $Violations -Message ("{0}: input_request.project must be 'radishflow'" -f $SampleName)
    }

    if ($request.task -ne "suggest_flowsheet_edits") {
        Add-Violation -Violations $Violations -Message ("{0}: input_request.task must be 'suggest_flowsheet_edits'" -f $SampleName)
    }

    if ($primaryArtifacts.Count -ne 1) {
        Add-Violation -Violations $Violations -Message ("{0}: input_request must contain exactly one primary artifact" -f $SampleName)
    }
    else {
        $primaryArtifact = $primaryArtifacts[0]
        if ([string]$primaryArtifact.name -ne "flowsheet_document") {
            Add-Violation -Violations $Violations -Message ("{0}: primary artifact name must be 'flowsheet_document'" -f $SampleName)
        }

        if ([string]$primaryArtifact.kind -ne "json") {
            Add-Violation -Violations $Violations -Message ("{0}: primary artifact kind must be 'json'" -f $SampleName)
        }

        if ([string]$primaryArtifact.mime_type -ne "application/json") {
            Add-Violation -Violations $Violations -Message ("{0}: primary artifact mime_type must be 'application/json'" -f $SampleName)
        }
    }

    if ($null -eq $context.document_revision) {
        Add-Violation -Violations $Violations -Message ("{0}: context.document_revision is required" -f $SampleName)
    }

    if ($diagnostics.Count -eq 0) {
        Add-Violation -Violations $Violations -Message ("{0}: suggest_flowsheet_edits samples must include diagnostics" -f $SampleName)
    }

    if ($selectedUnitIds.Count -eq 0 -and $selectedStreamIds.Count -eq 0) {
        Add-Violation -Violations $Violations -Message ("{0}: request must include selected_unit_ids or selected_stream_ids" -f $SampleName)
    }
}

function Test-ResponseContract {
    param(
        $Sample,
        $Response,
        [string]$ResponseLabel,
        [string]$SampleName,
        [System.Collections.Generic.List[string]]$Violations
    )

    $shape = $Sample.expected_response_shape
    $evaluation = $Sample.evaluation
    $answers = @(Get-Array $Response.answers)
    $issues = @(Get-Array $Response.issues)
    $actions = @(Get-Array $Response.proposed_actions)
    $citations = @(Get-Array $Response.citations)

    if ($Response.project -ne "radishflow") {
        Add-Violation -Violations $Violations -Message ("{0}: {1}.project must be 'radishflow'" -f $SampleName, $ResponseLabel)
    }

    if ($Response.task -ne "suggest_flowsheet_edits") {
        Add-Violation -Violations $Violations -Message ("{0}: {1}.task must be 'suggest_flowsheet_edits'" -f $SampleName, $ResponseLabel)
    }

    if ([string]$Response.status -ne [string]$shape.status) {
        Add-Violation -Violations $Violations -Message ("{0}: {1}.status does not match expected_response_shape.status" -f $SampleName, $ResponseLabel)
    }

    if ([string]$Response.risk_level -ne [string]$evaluation.expected_risk_level) {
        Add-Violation -Violations $Violations -Message ("{0}: {1}.risk_level does not match evaluation.expected_risk_level" -f $SampleName, $ResponseLabel)
    }

    if ($shape.requires_summary -and [string]::IsNullOrWhiteSpace([string]$Response.summary)) {
        Add-Violation -Violations $Violations -Message ("{0}: {1}.summary is required" -f $SampleName, $ResponseLabel)
    }

    if ($shape.requires_answers -and $answers.Count -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: {1} must contain at least 1 answer" -f $SampleName, $ResponseLabel)
    }

    if ($shape.requires_issues -and $issues.Count -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: {1} must contain at least 1 issue" -f $SampleName, $ResponseLabel)
    }

    if ($shape.requires_citations -and $citations.Count -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: {1} must contain at least 1 citation" -f $SampleName, $ResponseLabel)
    }

    if ($actions.Count -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: {1} must contain at least 1 proposed_action" -f $SampleName, $ResponseLabel)
    }

    $actualActionKinds = @(Get-Array $actions | ForEach-Object { [string]$_.kind })
    foreach ($requiredKind in (Get-Array $shape.required_action_kinds | ForEach-Object { [string]$_ })) {
        if ($actualActionKinds -notcontains $requiredKind) {
            Add-Violation -Violations $Violations -Message ("{0}: {1} is missing required action kind '{2}'" -f $SampleName, $ResponseLabel, $requiredKind)
        }
    }

    $highestActionRisk = 0
    $candidateEditCount = 0
    foreach ($action in $actions) {
        $highestActionRisk = [Math]::Max($highestActionRisk, (Get-RiskRank -RiskLevel ([string]$action.risk_level)))

        if ([string]$action.kind -ne "candidate_edit") {
            Add-Violation -Violations $Violations -Message ("{0}: {1} actions must remain candidate_edit for this task" -f $SampleName, $ResponseLabel)
            continue
        }

        $candidateEditCount += 1

        if ($null -eq $action.target) {
            Add-Violation -Violations $Violations -Message ("{0}: {1} candidate_edit must include target" -f $SampleName, $ResponseLabel)
        }
        elseif ([string]::IsNullOrWhiteSpace([string]$action.target.type) -or [string]::IsNullOrWhiteSpace([string]$action.target.id)) {
            Add-Violation -Violations $Violations -Message ("{0}: {1} candidate_edit target must include non-empty type and id" -f $SampleName, $ResponseLabel)
        }

        if ($null -eq $action.patch) {
            Add-Violation -Violations $Violations -Message ("{0}: {1} candidate_edit must include patch" -f $SampleName, $ResponseLabel)
        }
        elseif ((Get-PropertyCount -Object $action.patch) -lt 1) {
            Add-Violation -Violations $Violations -Message ("{0}: {1} candidate_edit patch must not be empty" -f $SampleName, $ResponseLabel)
        }

        if ($action.requires_confirmation -ne $true) {
            Add-Violation -Violations $Violations -Message ("{0}: {1} candidate_edit must set requires_confirmation=true" -f $SampleName, $ResponseLabel)
        }
    }

    if ($candidateEditCount -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: {1} must contain at least 1 candidate_edit" -f $SampleName, $ResponseLabel)
    }

    if ($Response.requires_confirmation -ne $true) {
        Add-Violation -Violations $Violations -Message ("{0}: {1}.requires_confirmation must be true when proposed_actions exist" -f $SampleName, $ResponseLabel)
    }

    if ((Get-RiskRank -RiskLevel ([string]$Response.risk_level)) -ne $highestActionRisk) {
        Add-Violation -Violations $Violations -Message ("{0}: {1}.risk_level must equal the highest proposed_action risk" -f $SampleName, $ResponseLabel)
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
            Add-Violation -Violations $Violations -Message ("{0}: referenced citation id '{1}' is missing from {2}.citations" -f $SampleName, $citationId, $ResponseLabel)
        }
    }

    Test-PathExpectations -Document $Response -Expressions (Get-Array $evaluation.must_have_json_paths) -ShouldExist $true -SampleName ("{0}:{1}" -f $SampleName, $ResponseLabel) -Violations $Violations
    Test-PathExpectations -Document $Response -Expressions (Get-Array $evaluation.must_not_have_json_paths) -ShouldExist $false -SampleName ("{0}:{1}" -f $SampleName, $ResponseLabel) -Violations $Violations
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
    throw "no sample files found for RadishFlow suggest edits regression"
}

$allViolations = New-Object System.Collections.Generic.List[string]
$matchedSampleCount = 0

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

    if ($null -ne $sample -and [string]$sample.task -eq "suggest_flowsheet_edits") {
        $matchedSampleCount += 1
        Test-DocumentAgainstSchema -Document $sample -SchemaPath $sampleSchemaPath -DocumentName ("{0} sample" -f $sampleName) -Violations $violations
        Test-DocumentAgainstSchema -Document $sample.input_request -SchemaPath $requestSchemaPath -DocumentName ("{0} input_request" -f $sampleName) -Violations $violations
        Test-DocumentAgainstSchema -Document $sample.golden_response -SchemaPath $responseSchemaPath -DocumentName ("{0} golden_response" -f $sampleName) -Violations $violations
        Test-RequestContract -Sample $sample -SampleName $sampleName -Violations $violations
        Test-ResponseContract -Sample $sample -Response $sample.golden_response -ResponseLabel "golden_response" -SampleName $sampleName -Violations $violations

        if ($null -ne $sample.candidate_response) {
            Test-DocumentAgainstSchema -Document $sample.candidate_response -SchemaPath $responseSchemaPath -DocumentName ("{0} candidate_response" -f $sampleName) -Violations $violations
            Test-ResponseContract -Sample $sample -Response $sample.candidate_response -ResponseLabel "candidate_response" -SampleName $sampleName -Violations $violations
        }
    }
    elseif ($null -eq $sample) {
        # parse failure already recorded
    }
    else {
        continue
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

if ($matchedSampleCount -eq 0) {
    throw "no suggest_flowsheet_edits sample files found for RadishFlow suggest edits regression"
}

if ($allViolations.Count -gt 0) {
    if ($FailOnViolation) {
        exit 1
    }

    Write-Warning ("radishflow suggest edits regression found {0} violation(s)." -f $allViolations.Count)
    return
}

Write-Host "radishflow suggest edits regression passed."
