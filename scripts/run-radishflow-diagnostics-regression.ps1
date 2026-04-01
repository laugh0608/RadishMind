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
$taskCardPath = Join-Path $repoRoot "docs/task-cards/radishflow-explain-diagnostics.md"

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

    if ($request.task -ne "explain_diagnostics") {
        Add-Violation -Violations $Violations -Message ("{0}: input_request.task must be 'explain_diagnostics'" -f $SampleName)
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

    if ($null -eq $context.diagnostic_summary -and $diagnostics.Count -eq 0) {
        Add-Violation -Violations $Violations -Message ("{0}: context must include diagnostic_summary or diagnostics" -f $SampleName)
    }

    if ($diagnostics.Count -eq 0) {
        Add-Violation -Violations $Violations -Message ("{0}: explain_diagnostics samples must include at least one diagnostics entry" -f $SampleName)
    }

    foreach ($diagnostic in $diagnostics) {
        if ([string]::IsNullOrWhiteSpace([string]$diagnostic.message)) {
            Add-Violation -Violations $Violations -Message ("{0}: each diagnostic must include message" -f $SampleName)
        }

        if ([string]::IsNullOrWhiteSpace([string]$diagnostic.severity)) {
            Add-Violation -Violations $Violations -Message ("{0}: each diagnostic must include severity" -f $SampleName)
        }
    }

    if ($selectedUnitIds.Count -eq 0 -and $selectedStreamIds.Count -eq 0 -and [string]$context.diagnostic_scope -ne "global") {
        Add-Violation -Violations $Violations -Message ("{0}: request must include selected_unit_ids, selected_stream_ids, or diagnostic_scope='global'" -f $SampleName)
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

    if ($response.project -ne "radishflow") {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response.project must be 'radishflow'" -f $SampleName)
    }

    if ($response.task -ne "explain_diagnostics") {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response.task must be 'explain_diagnostics'" -f $SampleName)
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

    if ($shape.requires_citations -and $citations.Count -lt 1) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response must contain at least 1 citation" -f $SampleName)
    }

    if (-not $shape.allow_proposed_actions -and $actions.Count -gt 0) {
        Add-Violation -Violations $Violations -Message ("{0}: golden_response should not contain proposed_actions" -f $SampleName)
    }

    $actualAnswerKinds = @(Get-Array $answers | ForEach-Object { [string]$_.kind })
    foreach ($requiredAnswerKind in (Get-Array $shape.required_answer_kinds | ForEach-Object { [string]$_ })) {
        if ($actualAnswerKinds -notcontains $requiredAnswerKind) {
            Add-Violation -Violations $Violations -Message ("{0}: golden_response is missing required answer kind '{1}'" -f $SampleName, $requiredAnswerKind)
        }
    }

    $actualActionKinds = @(Get-Array $actions | ForEach-Object { [string]$_.kind })
    foreach ($requiredKind in (Get-Array $shape.required_action_kinds | ForEach-Object { [string]$_ })) {
        if ($actualActionKinds -notcontains $requiredKind) {
            Add-Violation -Violations $Violations -Message ("{0}: golden_response is missing required action kind '{1}'" -f $SampleName, $requiredKind)
        }
    }

    if ($actions.Count -gt 0 -and $response.requires_confirmation -ne $true) {
        Add-Violation -Violations $Violations -Message ("{0}: responses with proposed_actions must set requires_confirmation=true" -f $SampleName)
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
    throw "no sample files found for RadishFlow diagnostics regression"
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
        Test-RequestContract -Sample $sample -SampleName $sampleName -Violations $violations
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

    Write-Warning ("radishflow diagnostics regression found {0} violation(s)." -f $allViolations.Count)
    return
}

Write-Host "radishflow diagnostics regression passed."
