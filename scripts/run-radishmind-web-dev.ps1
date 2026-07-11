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
  -VerifyOnly           Probe existing backend/frontend processes only.
  -ExitAfterProbe       Start missing local processes, probe, then stop spawned processes.
"@
    exit 0
}

if ($SavedDraftDev -and $Mode -ne "dev-live") {
    throw "-SavedDraftDev requires -Mode dev-live"
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$webDir = Join-Path $repoRoot "apps/radishmind-web"
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
    $routes = @(
        "/v1/control-plane/tenants/$encodedTenant/summary",
        "/v1/user-workspace/applications",
        "/v1/user-workspace/api-keys",
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
    [Console]::Error.WriteLine("- Saved Draft dev mode: pass -SavedDraftDev so backend and web opt in together; reused processes must already use the same mode.")
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
            if ($SavedDraftDev) {
                $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP = "1"
                $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE = "1"
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
                    -Arguments @("-NoProfile", "-ExecutionPolicy", "Bypass", "-File", $platformWrapper, "-Command", "serve") `
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
                if ($SavedDraftDev) {
                    $env:VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE = "dev-saved-draft-http"
                }
            }
            else {
                Remove-Item Env:VITE_RADISHMIND_READ_SOURCE -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL -ErrorAction SilentlyContinue
                Remove-Item Env:VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE -ErrorAction SilentlyContinue
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
        if ($SavedDraftDev) {
            Wait-Until -Name "Saved Draft dev route" -Probe {
                Invoke-SavedWorkflowDraftProbe `
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
        if ($SavedDraftDev) {
            Write-Step "Saved Draft memory-dev read/write mode passed for $savedDraftWorkspaceId/$savedDraftApplicationId."
        }
    }
    Write-Step "This is a dev-only launcher, not a production supervisor. It does not implement real auth, database, repository adapter, executor, confirmation, writeback, or replay."

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
