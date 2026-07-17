[CmdletBinding()]
param(
    [ValidateSet("offline", "dev-live")]
    [string]$Mode = "dev-live",
    [string]$BackendUrl = "http://127.0.0.1:7000",
    [string]$FrontendUrl = "http://127.0.0.1:4100",
    [string]$TenantRef = "tenant_demo",
    [string]$SubjectRef = "subject_demo_user",
    [int]$TimeoutSeconds = 60,
    [switch]$SavedDraftDev,
    [switch]$SavedDraftPostgresDevTest,
    [switch]$WorkflowDiagnosticsDev,
    [switch]$GatewayRequestPostgresDevTest,
    [switch]$ApplicationDraftDev,
    [switch]$ApplicationPublishDev,
    [switch]$ApplicationPublishPostgresDevTest,
    [switch]$ApplicationCatalogPostgresDevTest,
    [switch]$APIKeyLocalProduct,
    [switch]$WorkflowHTTPToolLocalProduct,
    [switch]$VerifyOnly,
    [switch]$ExitAfterProbe,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

if ($Help) {
@"
Usage: pwsh ./scripts/run-radishmind-web-dev.ps1 [-Mode offline|dev-live] [options]

Options:
  -Mode MODE            offline or dev-live. Default: dev-live
  -BackendUrl URL       Platform base URL. Default: http://127.0.0.1:7000
  -FrontendUrl URL      Web UI URL. Default: http://127.0.0.1:4100
  -TimeoutSeconds N     Probe timeout. Default: 60
  -SavedDraftDev        Enable the explicit memory-dev Saved Draft read/write path.
  -SavedDraftPostgresDevTest
                        Enable the explicit PostgreSQL dev/test Saved Draft path.
  -WorkflowDiagnosticsDev
                        Enable fixed mock Workflow failure scenarios; requires a Saved Draft dev mode.
  -GatewayRequestPostgresDevTest
                        Enable durable dev/test Gateway Request History and the scoped Gateway Playground.
  -ApplicationDraftDev  Enable the explicit memory-dev Application Configuration Draft path.
  -ApplicationPublishDev
                        Enable memory-dev Application Draft and Publish Candidate review.
  -ApplicationPublishPostgresDevTest
                        Enable PostgreSQL dev/test Application Draft and Publish Candidate review.
  -ApplicationCatalogPostgresDevTest
                        Enable the PostgreSQL dev/test Application Catalog lifecycle UI.
  -APIKeyLocalProduct   Enable the SQLite local-product Application/API key/Playground chain.
  -WorkflowHTTPToolLocalProduct
                        Enable the SQLite local-product Workflow HTTP Tool chain.
  -VerifyOnly           Probe existing backend/frontend processes only.
  -ExitAfterProbe       Start missing local processes, probe, then stop spawned processes.
"@
    exit 0
}

if ($SavedDraftDev -and $Mode -ne "dev-live") {
    throw "-SavedDraftDev requires -Mode dev-live"
}
if ($SavedDraftPostgresDevTest -and $Mode -ne "dev-live") {
    throw "-SavedDraftPostgresDevTest requires -Mode dev-live"
}
if ($SavedDraftDev -and $SavedDraftPostgresDevTest) {
    throw "Choose either -SavedDraftDev or -SavedDraftPostgresDevTest"
}
if ($WorkflowDiagnosticsDev -and -not ($SavedDraftDev -or $SavedDraftPostgresDevTest)) {
    throw "-WorkflowDiagnosticsDev requires -SavedDraftDev or -SavedDraftPostgresDevTest"
}
if ($GatewayRequestPostgresDevTest -and $Mode -ne "dev-live") {
    throw "-GatewayRequestPostgresDevTest requires -Mode dev-live"
}
if ($ApplicationDraftDev -and $Mode -ne "dev-live") {
    throw "-ApplicationDraftDev requires -Mode dev-live"
}
if ($ApplicationPublishDev -and $Mode -ne "dev-live") {
    throw "-ApplicationPublishDev requires -Mode dev-live"
}
if ($ApplicationPublishPostgresDevTest -and $Mode -ne "dev-live") {
    throw "-ApplicationPublishPostgresDevTest requires -Mode dev-live"
}
if ($ApplicationCatalogPostgresDevTest -and $Mode -ne "dev-live") {
    throw "-ApplicationCatalogPostgresDevTest requires -Mode dev-live"
}
if ($APIKeyLocalProduct -and $Mode -ne "dev-live") {
    throw "-APIKeyLocalProduct requires -Mode dev-live"
}
if ($WorkflowHTTPToolLocalProduct -and $Mode -ne "dev-live") {
    throw "-WorkflowHTTPToolLocalProduct requires -Mode dev-live"
}
if ($ApplicationPublishDev -and $ApplicationPublishPostgresDevTest) {
    throw "Choose either -ApplicationPublishDev or -ApplicationPublishPostgresDevTest"
}
if ($ApplicationDraftDev -and $ApplicationPublishPostgresDevTest) {
    throw "-ApplicationDraftDev cannot be combined with -ApplicationPublishPostgresDevTest"
}

$explicitComponentMode = $SavedDraftDev -or $SavedDraftPostgresDevTest -or $GatewayRequestPostgresDevTest -or `
    $ApplicationDraftDev -or $ApplicationPublishDev -or $ApplicationPublishPostgresDevTest -or $ApplicationCatalogPostgresDevTest
if ($APIKeyLocalProduct -and $explicitComponentMode) {
    throw "-APIKeyLocalProduct cannot be combined with explicit memory/PostgreSQL component modes"
}
if ($WorkflowHTTPToolLocalProduct -and $explicitComponentMode) {
    throw "-WorkflowHTTPToolLocalProduct cannot be combined with explicit memory/PostgreSQL component modes"
}
$platformProfile = if ($explicitComponentMode) { "configured" } else { "local-product" }

$savedDraftEnabled = $SavedDraftDev -or $SavedDraftPostgresDevTest -or $WorkflowHTTPToolLocalProduct

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$webDir = Join-Path $repoRoot "apps/radishmind-web"
$platformDir = Join-Path $repoRoot "services/platform"
$platformWrapper = Join-Path $repoRoot "scripts/run-platform-service.ps1"
$logDir = Join-Path $repoRoot "tmp/radishmind-web-dev"
$savedDraftWorkspaceId = "workspace_demo"
$savedDraftApplicationId = "app_flow_copilot"
$spawnedProcesses = @()

function Write-Step {
    param([string]$Message)
    Write-Host "[radishmind-web-dev] $Message"
}

function ConvertTo-ProcessArgument {
    param([string]$Value)
    if ($Value -match '[\s"]') {
        return '"' + ($Value -replace '"', '\"') + '"'
    }
    return $Value
}

function Get-RequiredCommand {
    param([string[]]$Names)
    foreach ($name in $Names) {
        $command = Get-Command $name -ErrorAction SilentlyContinue
        if ($null -ne $command) {
            return $command.Source
        }
    }
    throw "Missing required command: $($Names -join ' or ')"
}

function Get-SavedDraftDatabaseUrl {
    if ($env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL) {
        return $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL
    }
    $pgHost = if ($env:PGHOST) { $env:PGHOST } else { "127.0.0.1" }
    $pgPort = if ($env:PGPORT) { $env:PGPORT } else { "55432" }
    $pgUser = if ($env:PGUSER) { $env:PGUSER } else { "radishmind_runtime" }
    $pgDatabase = if ($env:PGDATABASE) { $env:PGDATABASE } else { "radishmind_saved_draft_test" }
    $pgSslMode = if ($env:PGSSLMODE) { $env:PGSSLMODE } else { "disable" }
    $encodedUser = [Uri]::EscapeDataString($pgUser)
    $userInfo = $encodedUser
    if ($env:PGPASSWORD) {
        $encodedPassword = [Uri]::EscapeDataString($env:PGPASSWORD)
        $userInfo = "${encodedUser}:${encodedPassword}"
    }
    if ($pgHost.Contains(":") -and -not $pgHost.StartsWith("[")) {
        $pgHost = "[$pgHost]"
    }
    $encodedDatabase = [Uri]::EscapeDataString($pgDatabase)
    $encodedSslMode = [Uri]::EscapeDataString($pgSslMode)
    return "postgresql://${userInfo}@${pgHost}:${pgPort}/${encodedDatabase}?sslmode=${encodedSslMode}"
}

function Invoke-SavedDraftPostgresMigrationStatus {
    $goPath = Get-RequiredCommand -Names @("go")
    $env:RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH = "1"
    $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP = "1"
    $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE = "1"
    $env:RADISHMIND_WORKFLOW_EXECUTOR_DEV = "1"
    $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE = "postgres_dev_test"
    $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
    $env:RADISHMIND_APPLICATION_DRAFT_DEV_HTTP = "1"
    $env:RADISHMIND_APPLICATION_DRAFT_DEV_WRITE = "1"
    $env:RADISHMIND_APPLICATION_DRAFT_STORE = "postgres_dev_test"
    $env:RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
    $env:RADISHMIND_WORKFLOW_RUN_STORE = "postgres_dev_test"
    $env:RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
    $env:RADISHMIND_GATEWAY_REQUEST_STORE = "postgres_dev_test"
    $env:RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
    $env:RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV = "1"
    $env:RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP = "1"
    $env:RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE = "1"
    $env:RADISHMIND_APPLICATION_PUBLISH_STORE = "postgres_dev_test"
    $env:RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
    $env:RADISHMIND_APPLICATION_CATALOG_DEV_HTTP = "1"
    $env:RADISHMIND_APPLICATION_CATALOG_DEV_WRITE = "1"
    $env:RADISHMIND_APPLICATION_CATALOG_STORE = "postgres_dev_test"
    $env:RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
    Push-Location $platformDir
    try {
        & $goPath run ./cmd/radishmind-workflow-draft-migrate status | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "Saved Draft PostgreSQL migration preflight failed"
        }
        & $goPath run ./cmd/radishmind-application-draft-migrate status | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "Application Draft PostgreSQL migration preflight failed"
        }
        & $goPath run ./cmd/radishmind-workflow-run-migrate status | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "Workflow Run PostgreSQL migration preflight failed"
        }
        & $goPath run ./cmd/radishmind-gateway-request-migrate status | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "Gateway Request PostgreSQL migration preflight failed"
        }
        if ($ApplicationPublishPostgresDevTest) {
            & $goPath run ./cmd/radishmind-application-publish-migrate status | Out-Null
            if ($LASTEXITCODE -ne 0) {
                throw "Application Publish PostgreSQL migration preflight failed"
            }
        }
        if ($ApplicationCatalogPostgresDevTest) {
            & $goPath run ./cmd/radishmind-application-catalog-migrate status | Out-Null
            if ($LASTEXITCODE -ne 0) {
                throw "Application Catalog PostgreSQL migration preflight failed"
            }
        }
    }
    finally {
        Pop-Location
    }
}

function Get-HttpUri {
    param([string]$Value)
    try {
        $uri = [Uri]$Value
    }
    catch {
        throw "Invalid URL: $Value"
    }
    if ($uri.Scheme -notin @("http", "https")) {
        throw "URL must use http or https: $Value"
    }
    if ($uri.Port -lt 0) {
        throw "URL must include an explicit local development port: $Value"
    }
    return $uri
}

function Test-BrowserUnsafePort {
    param([int]$Port)
    $blockedPorts = @(
        1, 7, 9, 11, 13, 15, 17, 19, 20, 21, 22, 23, 25, 37, 42, 43, 53, 69,
        77, 79, 87, 95, 101, 102, 103, 104, 109, 110, 111, 113, 115, 117, 119,
        123, 135, 137, 139, 143, 161, 179, 389, 427, 465, 512, 513, 514, 515,
        526, 530, 531, 532, 540, 548, 554, 556, 563, 587, 601, 636, 989, 990,
        993, 995, 1719, 1720, 1723, 2049, 3659, 4045, 4190, 5060, 5061, 6000,
        6566, 6697, 10080
    )
    return ($blockedPorts -contains $Port) -or ($Port -ge 6665 -and $Port -le 6669)
}

function Assert-BrowserSafeLocalUrl {
    param([Uri]$Uri)
    if (Test-BrowserUnsafePort -Port $Uri.Port) {
        throw @"
Browser unsafe port: $($Uri.Port)
Browsers can fail with ERR_UNSAFE_PORT before reaching $Uri.
Use the local defaults instead: backend http://127.0.0.1:7000 and web http://127.0.0.1:4100.
"@
    }
}

function Test-TcpPort {
    param(
        [string]$HostName,
        [int]$Port
    )
    $client = [System.Net.Sockets.TcpClient]::new()
    try {
        $connect = $client.ConnectAsync($HostName, $Port)
        if (-not $connect.Wait(500)) {
            return $false
        }
        return $client.Connected
    }
    catch {
        return $false
    }
    finally {
        $client.Dispose()
    }
}

function Invoke-JsonProbe {
    param(
        [string]$Url,
        [string]$ExpectedKind
    )
    $response = Invoke-WebRequest -Uri $Url -Method Get -Headers @{ Accept = "application/json" } -TimeoutSec 5
    if ($response.StatusCode -lt 200 -or $response.StatusCode -ge 300) {
        throw "Unexpected HTTP status $($response.StatusCode) from $Url"
    }
    $json = $response.Content | ConvertFrom-Json
    if ($ExpectedKind -and $json.kind -ne $ExpectedKind) {
        throw "Unexpected kind from $Url. expected=$ExpectedKind actual=$($json.kind)"
    }
    return $json
}

function Invoke-ControlPlaneReadRoutesProbe {
    param(
        [string]$BaseUrl,
        [string]$Tenant,
        [string]$Subject
    )
    $encodedTenant = [Uri]::EscapeDataString($Tenant)
    $applicationRoute = "/v1/user-workspace/applications"
    if ($ApplicationCatalogPostgresDevTest -or $platformProfile -eq "local-product") {
        $applicationRoute += "?workspace_id=$([Uri]::EscapeDataString($savedDraftWorkspaceId))&lifecycle_state=active&limit=1"
    }
    $apiKeyRoute = "/v1/user-workspace/api-keys"
    if ($platformProfile -eq "local-product") {
        $apiKeyRoute += "?workspace_id=$([Uri]::EscapeDataString($savedDraftWorkspaceId))&limit=1"
    }
    $routes = @(
        "/v1/control-plane/tenants/$encodedTenant/summary",
        $applicationRoute,
        $apiKeyRoute,
        "/v1/user-workspace/usage/quota-summary",
        "/v1/user-workspace/workflow-definitions",
        "/v1/user-workspace/runs",
        "/v1/control-plane/audit"
    )
    $headers = @{
        Accept = "application/json"
        "X-Request-Id" = "dev-live-script-probe"
        "X-RadishMind-Dev-Read-Identity" = "dev-live-read-consumer"
        "X-RadishMind-Dev-Read-Tenant" = $Tenant
        "X-RadishMind-Dev-Read-Subject" = $Subject
        "X-RadishMind-Dev-Read-Scopes" = "tenant:read,applications:read,api_keys:read,usage:read,runs:read,audit:read"
        "X-RadishMind-Dev-Read-Audit" = "audit_dev_live_script_probe"
    }
    foreach ($route in $routes) {
        $url = $BaseUrl.TrimEnd("/") + $route
        $response = Invoke-WebRequest -Uri $url -Method Get -Headers $headers -TimeoutSec 5
        if ($response.StatusCode -lt 200 -or $response.StatusCode -ge 300) {
            throw "Unexpected HTTP status $($response.StatusCode) from $url"
        }
        $json = $response.Content | ConvertFrom-Json
        if ($null -ne $json.failure_code) {
            throw "Route returned failure_code=$($json.failure_code) from $url"
        }
        if ($null -eq $json.items) {
            throw "Route did not return items[] from $url"
        }
    }
}

function Invoke-GatewayAPIKeyModeProbe {
    param([string]$BaseUrl)
    $url = $BaseUrl.TrimEnd("/") + "/v1/models"
    try {
        Invoke-WebRequest -Uri $url -Method Get `
            -Headers @{ Accept = "application/json"; "X-Request-Id" = "api-key-mode-probe" } `
            -TimeoutSec 5 | Out-Null
        throw "Gateway accepted an unauthenticated model request; api_key_dev_test mode is not active"
    }
    catch {
        $response = $_.Exception.Response
        if ($null -eq $response -or [int]$response.StatusCode -ne 401) {
            throw
        }
        if ($response.Content -is [System.Net.Http.HttpContent]) {
            $body = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
        }
        elseif ($response.PSObject.Methods.Name -contains "GetResponseStream") {
            $stream = $response.GetResponseStream()
            $reader = [System.IO.StreamReader]::new($stream)
            try { $body = $reader.ReadToEnd() } finally { $reader.Dispose(); $stream.Dispose() }
        }
        else {
            throw "Gateway API key mode probe could not read the HTTP failure body"
        }
        $document = $body | ConvertFrom-Json
        if ($document.error.code -ne "api_key_missing") {
            throw "Gateway API key mode probe returned an unexpected failure: $($document.error.code)"
        }
    }
}

function Invoke-SavedWorkflowDraftProbe {
    param(
        [string]$BaseUrl,
        [string]$Tenant,
        [string]$Subject,
        [string]$WorkspaceId,
        [string]$ApplicationId
    )
    $query = "workspace_id=$([Uri]::EscapeDataString($WorkspaceId))&application_id=$([Uri]::EscapeDataString($ApplicationId))"
    $url = $BaseUrl.TrimEnd("/") + "/v1/user-workspace/workflow-drafts?$query"
    $headers = @{
        Accept = "application/json"
        "X-Request-Id" = "dev-live-saved-draft-probe"
        "X-RadishMind-Dev-Read-Identity" = "dev-live-saved-draft-probe"
        "X-RadishMind-Dev-Read-Tenant" = $Tenant
        "X-RadishMind-Dev-Read-Subject" = $Subject
        "X-RadishMind-Dev-Read-Scopes" = "workflow_drafts:read,workflow_drafts:write"
        "X-RadishMind-Dev-Read-Audit" = "audit_dev_live_saved_draft_probe"
        "X-RadishMind-Dev-Workflow-Workspace" = $WorkspaceId
        "X-RadishMind-Dev-Workflow-Application" = $ApplicationId
    }
    $response = Invoke-WebRequest -Uri $url -Method Get -Headers $headers -TimeoutSec 5
    if ($response.StatusCode -lt 200 -or $response.StatusCode -ge 300) {
        throw "Unexpected HTTP status $($response.StatusCode) from $url"
    }
    $json = $response.Content | ConvertFrom-Json
    if ($null -ne $json.failure_code) {
        throw "Saved Draft route returned failure_code=$($json.failure_code) from $url"
    }
    if ($json.PSObject.Properties.Name -notcontains "draft_summaries") {
        throw "Saved Draft route did not return draft_summaries[] from $url"
    }
}

function Invoke-WorkflowExecutorProbe {
    param(
        [string]$BaseUrl,
        [string]$Tenant,
        [string]$Subject,
        [string]$WorkspaceId,
        [string]$ApplicationId
    )
    $url = $BaseUrl.TrimEnd("/") + "/v1/user-workspace/workflow-drafts/draft_executor_probe_missing/runs"
    $headers = @{
        Accept = "application/json"
        "Content-Type" = "application/json"
        "X-Request-Id" = "dev-live-workflow-executor-probe"
        "X-RadishMind-Dev-Read-Identity" = "dev-live-workflow-executor-probe"
        "X-RadishMind-Dev-Read-Tenant" = $Tenant
        "X-RadishMind-Dev-Read-Subject" = $Subject
        "X-RadishMind-Dev-Read-Scopes" = "workflow_drafts:read,workflow_runs:execute,workflow_runs:read"
        "X-RadishMind-Dev-Read-Audit" = "audit_dev_live_workflow_executor_probe"
        "X-RadishMind-Dev-Workflow-Workspace" = $WorkspaceId
        "X-RadishMind-Dev-Workflow-Application" = $ApplicationId
    }
    $body = @{
        workspace_id = $WorkspaceId
        application_id = $ApplicationId
        input_text = "executor dev gate probe"
        condition_values = @{}
    } | ConvertTo-Json -Depth 4
    $response = Invoke-WebRequest -Uri $url -Method Post -Headers $headers -Body $body -TimeoutSec 5
    if ($response.StatusCode -lt 200 -or $response.StatusCode -ge 300) {
        throw "Unexpected HTTP status $($response.StatusCode) from $url"
    }
    $json = $response.Content | ConvertFrom-Json
    if ($json.failure_code -ne "workflow_run_draft_not_found") {
        throw "Workflow Executor dev gate probe returned unexpected failure_code=$($json.failure_code)"
    }
    if ($null -ne $json.run) {
        throw "Workflow Executor dev gate probe must not create a run for a missing draft"
    }
    $listUrl = $BaseUrl.TrimEnd("/") + "/v1/user-workspace/workflow-runs?workspace_id=$WorkspaceId&application_id=$ApplicationId&limit=1"
    $listResponse = Invoke-WebRequest -Uri $listUrl -Method Get -Headers $headers -TimeoutSec 5
    $listJson = $listResponse.Content | ConvertFrom-Json
    if ($null -ne $listJson.failure_code -or $listJson.PSObject.Properties.Name -notcontains "runs") {
        throw "Workflow run history list probe did not return a successful runs[] envelope"
    }
}

function Invoke-CorsProbe {
    param(
        [string]$ReadUrl,
        [string]$Origin
    )
    $headers = @{
        Origin = $Origin
        "Access-Control-Request-Method" = "GET"
        "Access-Control-Request-Headers" = "X-RadishMind-Dev-Read-Identity"
    }
    $response = Invoke-WebRequest -Uri $ReadUrl -Method Options -Headers $headers -TimeoutSec 5
    $allowOrigin = $response.Headers["Access-Control-Allow-Origin"]
    $allowHeaders = $response.Headers["Access-Control-Allow-Headers"]
    if ($allowOrigin -ne $Origin) {
        throw "CORS preflight did not allow $Origin. actual Access-Control-Allow-Origin=$allowOrigin"
    }
    if ($allowHeaders -notmatch "X-RadishMind-Dev-Read-Identity") {
        throw "CORS preflight did not allow dev read identity header"
    }
}

function Invoke-PageProbe {
    param([string]$Url)
    $response = Invoke-WebRequest -Uri $Url -Method Get -TimeoutSec 5
    if ($response.StatusCode -lt 200 -or $response.StatusCode -ge 300) {
        throw "Unexpected HTTP status $($response.StatusCode) from $Url"
    }
    if ($response.Content -notmatch '<div id="root">') {
        throw "Frontend responded, but it does not look like the RadishMind web shell."
    }
}

function Wait-Until {
    param(
        [string]$Name,
        [scriptblock]$Probe
    )
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    $lastError = $null
    while ((Get-Date) -lt $deadline) {
        try {
            & $Probe
            Write-Step "$Name probe passed."
            return
        }
        catch {
            $lastError = $_.Exception.Message
            Start-Sleep -Milliseconds 750
        }
    }
    throw "$Name probe failed after ${TimeoutSeconds}s. Last error: $lastError"
}

function Assert-ExistingBackend {
    param(
        [string]$HealthzUrl,
        [Uri]$BackendUri
    )
    try {
        Invoke-JsonProbe -Url $HealthzUrl -ExpectedKind "" | Out-Null
        Write-Step "Existing backend healthz probe passed."
    }
    catch {
        [Console]::Error.WriteLine("Backend port $($BackendUri.Port) is open, but $HealthzUrl did not answer like RadishMind platform.")
        [Console]::Error.WriteLine("Last error: $($_.Exception.Message)")
        throw "backend port $($BackendUri.Port) is occupied by a non-RadishMind service. Stop that process or pass -BackendUrl with a free local port."
    }
}

function Start-LoggedProcess {
    param(
        [string]$Name,
        [string]$FilePath,
        [string[]]$Arguments,
        [string]$WorkingDirectory
    )
    New-Item -ItemType Directory -Force -Path $logDir | Out-Null
    $stdoutPath = Join-Path $logDir "$Name.out.log"
    $stderrPath = Join-Path $logDir "$Name.err.log"
    $argumentLine = ($Arguments | ForEach-Object { ConvertTo-ProcessArgument $_ }) -join " "
    Write-Step "Starting $Name. Logs: $stdoutPath ; $stderrPath"
    $process = Start-Process `
        -FilePath $FilePath `
        -ArgumentList $argumentLine `
        -WorkingDirectory $WorkingDirectory `
        -RedirectStandardOutput $stdoutPath `
        -RedirectStandardError $stderrPath `
        -WindowStyle Hidden `
        -PassThru
    $script:spawnedProcesses += $process
    return $process
}

function Stop-SpawnedProcesses {
    foreach ($process in $script:spawnedProcesses) {
        try {
            if ($null -ne $process -and -not $process.HasExited) {
                Write-Step "Stopping process $($process.Id)."
                Stop-Process -Id $process.Id -ErrorAction SilentlyContinue
            }
        }
        catch {
            Write-Warning $_.Exception.Message
        }
    }
}

function Show-FailureHelp {
    param([string]$Message)
    [Console]::Error.WriteLine("")
    [Console]::Error.WriteLine("RadishMind web dev entry failed:")
    [Console]::Error.WriteLine($Message)
    [Console]::Error.WriteLine("")
    [Console]::Error.WriteLine("Common local failures:")
    [Console]::Error.WriteLine("- Port conflict: backend should answer on http://127.0.0.1:7000 and web on http://127.0.0.1:4100.")
    [Console]::Error.WriteLine("- macOS port 7000 conflict: AirPlay / Control Center may occupy it; retry with -BackendUrl http://127.0.0.1:7100.")
    [Console]::Error.WriteLine("- Dev-live auth: backend must be started with RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1 for fake-store-backed read routes.")
    [Console]::Error.WriteLine("- Workflow mode: choose -SavedDraftDev, -SavedDraftPostgresDevTest, or -WorkflowHTTPToolLocalProduct so backend and web opt in together.")
    [Console]::Error.WriteLine("- API key local product: use -APIKeyLocalProduct by itself so SQLite lifecycle and api_key_dev_test auth stay aligned.")
    [Console]::Error.WriteLine("- PostgreSQL dev/test mode: start and migrate it with the saved draft PostgreSQL dev/test runner.")
    [Console]::Error.WriteLine("- CORS/preflight: platform should allow http://127.0.0.1:4100 and dev read headers in local development.")
    [Console]::Error.WriteLine("- Missing dependencies: run npm install in apps/radishmind-web if npm cannot start Vite.")
    [Console]::Error.WriteLine("- Backend diagnostics: run pwsh ./scripts/run-platform-service.ps1 -Command diagnostics from the repo root.")
    [Console]::Error.WriteLine("- Logs: $logDir")
}

$backendUri = Get-HttpUri -Value $BackendUrl
$frontendUri = Get-HttpUri -Value $FrontendUrl
Assert-BrowserSafeLocalUrl -Uri $backendUri
Assert-BrowserSafeLocalUrl -Uri $frontendUri

$healthzUrl = "$($BackendUrl.TrimEnd('/'))/healthz"
$tenantSummaryUrl = "$($BackendUrl.TrimEnd('/'))/v1/control-plane/tenants/$TenantRef/summary"
$frontendOrigin = $FrontendUrl.TrimEnd("/")

try {
    if (-not (Test-Path -LiteralPath (Join-Path $webDir "package.json") -PathType Leaf)) {
        throw "Missing RadishMind web package.json: $webDir"
    }
    if ($Mode -eq "dev-live" -and -not (Test-Path -LiteralPath $platformWrapper -PathType Leaf)) {
        throw "Missing platform wrapper: $platformWrapper"
    }
    if ($SavedDraftPostgresDevTest -or $GatewayRequestPostgresDevTest -or $ApplicationPublishPostgresDevTest -or $ApplicationCatalogPostgresDevTest) {
        Invoke-SavedDraftPostgresMigrationStatus
        Write-Step "PostgreSQL dev/test migration preflight passed."
    }

    if (-not $VerifyOnly) {
        $npmPath = Get-RequiredCommand -Names @("npm.cmd", "npm")
        if ($Mode -eq "dev-live") {
            $pwshPath = Get-RequiredCommand -Names @("pwsh", "powershell")

            if (-not $env:RADISHMIND_PLATFORM_LISTEN_ADDR) {
                $env:RADISHMIND_PLATFORM_LISTEN_ADDR = "$($backendUri.Host):$($backendUri.Port)"
            }
            if (-not $env:RADISHMIND_PLATFORM_PROVIDER) {
                $env:RADISHMIND_PLATFORM_PROVIDER = "mock"
            }
            if (-not $env:RADISHMIND_PLATFORM_MODEL) {
                $env:RADISHMIND_PLATFORM_MODEL = "radishmind-local-dev"
            }
            $env:RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH = "1"
            if ($APIKeyLocalProduct) {
                $env:RADISHMIND_GATEWAY_AUTH_MODE = "api_key_dev_test"
            }
            if ($ApplicationPublishPostgresDevTest) {
                $databaseUrl = Get-SavedDraftDatabaseUrl
                $env:RADISHMIND_APPLICATION_DRAFT_DEV_HTTP = "1"
                $env:RADISHMIND_APPLICATION_DRAFT_DEV_WRITE = "1"
                $env:RADISHMIND_APPLICATION_DRAFT_STORE = "postgres_dev_test"
                $env:RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL = $databaseUrl
                $env:RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP = "1"
                $env:RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE = "1"
                $env:RADISHMIND_APPLICATION_PUBLISH_STORE = "postgres_dev_test"
                $env:RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL = $databaseUrl
            }
            elseif ($ApplicationPublishDev) {
                $env:RADISHMIND_APPLICATION_DRAFT_DEV_HTTP = "1"
                $env:RADISHMIND_APPLICATION_DRAFT_DEV_WRITE = "1"
                $env:RADISHMIND_APPLICATION_DRAFT_STORE = "memory_dev"
                $env:RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP = "1"
                $env:RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE = "1"
                $env:RADISHMIND_APPLICATION_PUBLISH_STORE = "memory_dev"
            }
            elseif ($ApplicationDraftDev) {
                $env:RADISHMIND_APPLICATION_DRAFT_DEV_HTTP = "1"
                $env:RADISHMIND_APPLICATION_DRAFT_DEV_WRITE = "1"
                $env:RADISHMIND_APPLICATION_DRAFT_STORE = "memory_dev"
            }
            if ($SavedDraftDev) {
                $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP = "1"
                $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE = "1"
                $env:RADISHMIND_WORKFLOW_EXECUTOR_DEV = "1"
                $env:RADISHMIND_WORKFLOW_TOOL_ACTION_DEV = "1"
                $env:RADISHMIND_WORKFLOW_HTTP_TOOL_EXECUTION_DEV = "1"
                $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE = "memory_dev"
            }
            elseif ($SavedDraftPostgresDevTest) {
                $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP = "1"
                $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE = "1"
                $env:RADISHMIND_WORKFLOW_EXECUTOR_DEV = "1"
                $env:RADISHMIND_WORKFLOW_TOOL_ACTION_DEV = "1"
                $env:RADISHMIND_WORKFLOW_HTTP_TOOL_EXECUTION_DEV = "1"
                $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE = "postgres_dev_test"
                $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
				$env:RADISHMIND_WORKFLOW_RUN_STORE = "postgres_dev_test"
				$env:RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
            }
            if ($WorkflowDiagnosticsDev) {
                $env:RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV = "1"
            }
            if ($GatewayRequestPostgresDevTest) {
                $env:RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV = "1"
                $env:RADISHMIND_GATEWAY_REQUEST_STORE = "postgres_dev_test"
                $env:RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
            }
            if ($ApplicationCatalogPostgresDevTest) {
                $env:RADISHMIND_APPLICATION_CATALOG_DEV_HTTP = "1"
                $env:RADISHMIND_APPLICATION_CATALOG_DEV_WRITE = "1"
                $env:RADISHMIND_APPLICATION_CATALOG_STORE = "postgres_dev_test"
                $env:RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL = Get-SavedDraftDatabaseUrl
            }

            $backendPortOpen = Test-TcpPort -HostName $backendUri.Host -Port $backendUri.Port
            if ($backendPortOpen) {
                Write-Step "Backend port $($backendUri.Port) is already open; reusing it if probes pass."
                Assert-ExistingBackend -HealthzUrl $healthzUrl -BackendUri $backendUri
            }
            else {
                Start-LoggedProcess `
                    -Name "platform" `
                    -FilePath $pwshPath `
                    -Arguments @("-NoProfile", "-ExecutionPolicy", "Bypass", "-File", $platformWrapper, "-Profile", $platformProfile, "-Command", "serve") `
                    -WorkingDirectory $repoRoot | Out-Null
            }
        }

        $frontendPortOpen = Test-TcpPort -HostName $frontendUri.Host -Port $frontendUri.Port
        if ($frontendPortOpen) {
            Write-Step "Web port $($frontendUri.Port) is already open; reusing it if probes pass."
        }
        else {
            if ($Mode -eq "dev-live") {
                $env:VITE_RADISHMIND_READ_SOURCE = "dev-live-http"
                $env:VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL = $BackendUrl.TrimEnd("/")
                $env:VITE_RADISHMIND_DEV_READ_TENANT_REF = $TenantRef
                $env:VITE_RADISHMIND_DEV_READ_SUBJECT_REF = $SubjectRef
                if ($ApplicationDraftDev -or $ApplicationPublishDev -or $ApplicationPublishPostgresDevTest -or $APIKeyLocalProduct) {
                    $env:VITE_RADISHMIND_APPLICATION_DRAFT_SOURCE = "dev-application-draft-http"
                    $env:VITE_RADISHMIND_APPLICATION_DRAFT_BASE_URL = $BackendUrl.TrimEnd("/")
                    $env:VITE_RADISHMIND_APPLICATION_DRAFT_WORKSPACE_ID = $savedDraftWorkspaceId
                }
                if ($ApplicationPublishDev -or $ApplicationPublishPostgresDevTest -or $APIKeyLocalProduct) {
                    $env:VITE_RADISHMIND_APPLICATION_PUBLISH_SOURCE = "dev-application-publish-http"
                    $env:VITE_RADISHMIND_APPLICATION_PUBLISH_BASE_URL = $BackendUrl.TrimEnd("/")
                    $env:VITE_RADISHMIND_APPLICATION_PUBLISH_WORKSPACE_ID = $savedDraftWorkspaceId
                }
                if ($ApplicationCatalogPostgresDevTest -or $APIKeyLocalProduct -or $WorkflowHTTPToolLocalProduct) {
                    $env:VITE_RADISHMIND_APPLICATION_CATALOG_SOURCE = "dev-application-catalog-http"
                    $env:VITE_RADISHMIND_APPLICATION_CATALOG_BASE_URL = $BackendUrl.TrimEnd("/")
                    $env:VITE_RADISHMIND_APPLICATION_CATALOG_WORKSPACE_ID = $savedDraftWorkspaceId
                }
                if ($APIKeyLocalProduct -or $WorkflowHTTPToolLocalProduct) {
                    $env:VITE_RADISHMIND_API_KEY_LIFECYCLE_SOURCE = "dev-api-key-lifecycle-http"
                    $env:VITE_RADISHMIND_API_KEY_LIFECYCLE_BASE_URL = $BackendUrl.TrimEnd("/")
                    $env:VITE_RADISHMIND_API_KEY_LIFECYCLE_WORKSPACE_ID = $savedDraftWorkspaceId
                }
                if ($APIKeyLocalProduct) {
                    $env:VITE_RADISHMIND_GATEWAY_AUTH_MODE = "api_key_dev_test"
                }
                if ($savedDraftEnabled) {
                    $env:VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE = "dev-saved-draft-http"
                    $env:VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE = "dev-workflow-executor-http"
                    $env:VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SOURCE = "dev-workflow-http-tool-http"
                    $env:VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SCOPE_GRANTS = "workflow_drafts:read,workflow_tool_actions:plan,workflow_tool_actions:read,workflow_tool_actions:confirm,workflow_tool_actions:execute,workflow_runs:execute"
                }
                if ($WorkflowDiagnosticsDev) {
                    $env:VITE_RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV = "true"
                }
                if ($GatewayRequestPostgresDevTest -or $APIKeyLocalProduct) {
                    $env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SOURCE = "dev-gateway-request-history-http"
                    $env:VITE_RADISHMIND_GATEWAY_PLAYGROUND_SOURCE = "dev-gateway-playground-http"
                    $env:VITE_RADISHMIND_GATEWAY_PLAYGROUND_BASE_URL = $BackendUrl.TrimEnd("/")
                    $playgroundModel = "radishmind-local-dev"
                    if ($env:RADISHMIND_PLATFORM_MODEL) {
                        $playgroundModel = $env:RADISHMIND_PLATFORM_MODEL
                    }
                    $env:VITE_RADISHMIND_GATEWAY_PLAYGROUND_MODEL = $playgroundModel
                    $env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_BASE_URL = $BackendUrl.TrimEnd("/")
                    $env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_TENANT_REF = $TenantRef
                    $env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_WORKSPACE_ID = $savedDraftWorkspaceId
                    $env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_CONSUMER_REF = "consumer_web_dev"
                    $env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_APPLICATION_ID = $savedDraftApplicationId
                    $env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SUBJECT_REF = $SubjectRef
                }
            }
            else {
                Remove-Item Env:VITE_RADISHMIND_READ_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SCOPE_GRANTS -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_APPLICATION_DRAFT_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_APPLICATION_PUBLISH_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_APPLICATION_PUBLISH_BASE_URL -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_APPLICATION_PUBLISH_WORKSPACE_ID -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_APPLICATION_CATALOG_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_APPLICATION_CATALOG_BASE_URL -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_APPLICATION_CATALOG_WORKSPACE_ID -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_API_KEY_LIFECYCLE_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_API_KEY_LIFECYCLE_BASE_URL -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_API_KEY_LIFECYCLE_WORKSPACE_ID -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_AUTH_MODE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_PLAYGROUND_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_PLAYGROUND_BASE_URL -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_PLAYGROUND_MODEL -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_BASE_URL -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_TENANT_REF -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_WORKSPACE_ID -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_CONSUMER_REF -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_APPLICATION_ID -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SUBJECT_REF -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_APPLICATION_DRAFT_BASE_URL -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_APPLICATION_DRAFT_WORKSPACE_ID -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV -ErrorAction SilentlyContinue
            }

            Start-LoggedProcess `
                -Name "web" `
                -FilePath $npmPath `
                -Arguments @("run", "dev") `
                -WorkingDirectory $webDir | Out-Null
        }
    }

    if ($Mode -eq "dev-live") {
        Wait-Until -Name "backend healthz" -Probe { Invoke-JsonProbe -Url $healthzUrl -ExpectedKind "" | Out-Null }
        Wait-Until -Name "dev-live read route CORS" -Probe { Invoke-CorsProbe -ReadUrl $tenantSummaryUrl -Origin $frontendOrigin }
        Wait-Until -Name "dev-live read routes" -Probe {
            Invoke-ControlPlaneReadRoutesProbe -BaseUrl $BackendUrl -Tenant $TenantRef -Subject $SubjectRef
        }
        if ($APIKeyLocalProduct) {
            Wait-Until -Name "Gateway API key auth mode" -Probe { Invoke-GatewayAPIKeyModeProbe -BaseUrl $BackendUrl }
        }
        if ($savedDraftEnabled) {
            Wait-Until -Name "Saved Draft dev route" -Probe {
                Invoke-SavedWorkflowDraftProbe `
                    -BaseUrl $BackendUrl `
                    -Tenant $TenantRef `
                    -Subject $SubjectRef `
                    -WorkspaceId $savedDraftWorkspaceId `
                    -ApplicationId $savedDraftApplicationId
            }
            Wait-Until -Name "Workflow Executor dev route" -Probe {
                Invoke-WorkflowExecutorProbe `
                    -BaseUrl $BackendUrl `
                    -Tenant $TenantRef `
                    -Subject $SubjectRef `
                    -WorkspaceId $savedDraftWorkspaceId `
                    -ApplicationId $savedDraftApplicationId
            }
        }
    }

    Wait-Until -Name "frontend web" -Probe { Invoke-PageProbe -Url $FrontendUrl }

    Write-Step "RadishMind web is ready: $FrontendUrl"
    if ($Mode -eq "dev-live") {
        Write-Step "Dev-live read backend passed: $healthzUrl ; $tenantSummaryUrl"
        if ($SavedDraftPostgresDevTest) {
            Write-Step "Saved Draft PostgreSQL dev/test read/write mode passed for $savedDraftWorkspaceId/$savedDraftApplicationId."
        }
        if ($GatewayRequestPostgresDevTest) {
            Write-Step "Gateway Playground and Request History PostgreSQL dev/test mode enabled for $savedDraftWorkspaceId/consumer_web_dev/$savedDraftApplicationId."
        }
        elseif ($SavedDraftDev) {
            Write-Step "Saved Draft memory-dev read/write mode passed for $savedDraftWorkspaceId/$savedDraftApplicationId."
        }
        if ($APIKeyLocalProduct) {
            Write-Step "API key SQLite local-product Web chain enabled for $savedDraftWorkspaceId; raw credentials remain browser-memory only."
        }
        if ($WorkflowHTTPToolLocalProduct) {
            Write-Step "Workflow HTTP Tool SQLite local-product Web chain enabled for $savedDraftWorkspaceId/$savedDraftApplicationId; approve and execute remain separate actions."
        }
    }
    Write-Step "This is a dev-only launcher, not a production supervisor. Controlled execution is dev-only; production auth, secret resolution, unrestricted tools, automatic confirmation, writeback and replay remain disabled."

    if (-not $VerifyOnly -and -not $ExitAfterProbe -and $spawnedProcesses.Count -gt 0) {
        Write-Step "Press Ctrl+C to stop spawned local dev processes."
        while ($true) {
            foreach ($process in $spawnedProcesses) {
                if ($process.HasExited) {
                    throw "Spawned process exited early: pid=$($process.Id) exit_code=$($process.ExitCode). Check logs in $logDir"
                }
            }
            Start-Sleep -Seconds 2
        }
    }
}
catch {
    Show-FailureHelp -Message $_.Exception.Message
    Stop-SpawnedProcesses
    exit 1
}
finally {
    if (-not $VerifyOnly) {
        Stop-SpawnedProcesses
    }
}
