[CmdletBinding()]
param(
    [switch]$SkipTextFiles
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location $repoRoot

try {
    if (-not $SkipTextFiles) {
        & (Join-Path $repoRoot "scripts/check-text-files.ps1")
        if ($LASTEXITCODE -ne 0) {
            throw "command failed: ./scripts/check-text-files.ps1"
        }
    }

    & (Join-Path $repoRoot "scripts/check-radish-docs-qa-eval.ps1")
    if ($LASTEXITCODE -ne 0) {
        throw "command failed: ./scripts/check-radish-docs-qa-eval.ps1"
    }

    $requiredFiles = @(
        "AGENTS.md",
        "CLAUDE.md",
        "LICENSE",
        ".editorconfig",
        ".gitattributes",
        ".github/PULL_REQUEST_TEMPLATE.md",
        ".github/rulesets/README.md",
        ".github/rulesets/master-protection.json",
        ".github/workflows/pr-check.yml",
        ".github/workflows/release-check.yml",
        "README.md",
        "docs/README.md",
        "docs/task-cards/README.md",
        "docs/task-cards/radishflow-explain-diagnostics.md",
        "docs/task-cards/radishflow-suggest-flowsheet-edits.md",
        "docs/task-cards/radishflow-explain-control-plane-state.md",
        "docs/task-cards/radish-answer-docs-question.md",
        "docs/radishmind-product-scope.md",
        "docs/radishmind-architecture.md",
        "docs/radishmind-roadmap.md",
        "docs/radishmind-integration-contracts.md",
        "docs/adr/0001-branch-and-pr-governance.md",
        "docs/devlogs/README.md",
        "docs/devlogs/2026-W14.md",
        "contracts/README.md",
        "contracts/copilot-request.schema.json",
        "contracts/copilot-response.schema.json",
        "datasets/README.md",
        "datasets/eval/README.md",
        "datasets/eval/radishflow-task-sample.schema.json",
        "datasets/eval/radish-task-sample.schema.json",
        "datasets/eval/radish/answer-docs-question-direct-answer-001.json",
        "datasets/eval/radish/answer-docs-question-evidence-gap-001.json",
        "datasets/eval/radish/answer-docs-question-navigation-001.json",
        "scripts/check-radish-docs-qa-eval.ps1",
        "scripts/check-radish-docs-qa-eval.sh",
        "scripts/check-text-files.ps1",
        "scripts/check-text-files.sh",
        "scripts/check-repo.ps1",
        "scripts/check-repo.sh"
    )

    foreach ($relativePath in $requiredFiles) {
        if (-not (Test-Path -LiteralPath (Join-Path $repoRoot $relativePath) -PathType Leaf)) {
            throw "missing required file: $relativePath"
        }
    }

    $agentsContent = Get-Content -Path "AGENTS.md" -Raw
    if (-not $agentsContent.Contains('当前常态开发分支为 `dev`')) {
        throw "AGENTS.md does not mention dev as the default development branch"
    }

    $prTemplateContent = Get-Content -Path ".github/PULL_REQUEST_TEMPLATE.md" -Raw
    if (-not $prTemplateContent.Contains('默认目标分支为 `dev`')) {
        throw "PULL_REQUEST_TEMPLATE.md does not mention dev as the default target branch"
    }

    $ruleset = Get-Content -Path ".github/rulesets/master-protection.json" -Raw | ConvertFrom-Json
    if ($ruleset.conditions.ref_name.include -notcontains "refs/heads/master") {
        throw "master-protection.json does not target refs/heads/master"
    }

    $requiredCheckRule = $ruleset.rules | Where-Object { $_.type -eq "required_status_checks" } | Select-Object -First 1
    if ($null -eq $requiredCheckRule) {
        throw "master-protection.json is missing required_status_checks"
    }

    $contexts = @($requiredCheckRule.parameters.required_status_checks | ForEach-Object { $_.context })
    if ($contexts -notcontains "Repo Hygiene") {
        throw "master-protection.json is missing Repo Hygiene required check"
    }

    if ($contexts -notcontains "Planning Baseline") {
        throw "master-protection.json is missing Planning Baseline required check"
    }

    $prWorkflow = Get-Content -Path ".github/workflows/pr-check.yml" -Raw
    foreach ($pattern in @("name: PR Checks", "- master", "name: Repo Hygiene", "name: Planning Baseline")) {
        if (-not $prWorkflow.Contains($pattern)) {
            throw ".github/workflows/pr-check.yml is missing expected content: $pattern"
        }
    }

    $releaseWorkflow = Get-Content -Path ".github/workflows/release-check.yml" -Raw
    foreach ($pattern in @("name: Release Checks", "v*-dev", "v*-test", "v*-release", "name: Release Repo Hygiene", "name: Release Planning Baseline")) {
        if (-not $releaseWorkflow.Contains($pattern)) {
            throw ".github/workflows/release-check.yml is missing expected content: $pattern"
        }
    }

    Write-Host "repository baseline checks passed."
}
finally {
    Pop-Location
}
