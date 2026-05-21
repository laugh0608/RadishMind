[CmdletBinding()]
param(
    [string]$BackendUrl = "http://127.0.0.1:7000",
    [string]$FrontendUrl = "http://127.0.0.1:4000",
    [int]$TimeoutSeconds = 60,
    [switch]$VerifyOnly,
    [switch]$ExitAfterProbe
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$consoleDir = Join-Path $repoRoot "apps/radishmind-console"
$platformWrapper = Join-Path $repoRoot "scripts/run-platform-service.ps1"
$logDir = Join-Path $repoRoot "tmp/radishmind-console-dev"

$spawnedProcesses = @()

function Write-Step {
    param([string]$Message)
    Write-Host "[radishmind-console-dev] $Message"
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
Browsers can fail with ERR_UNSAFE_PORT before reaching RadishMind.
Use the local defaults instead: backend http://127.0.0.1:7000 and frontend http://127.0.0.1:4000.
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

function Invoke-PageProbe {
    param([string]$Url)
    $response = Invoke-WebRequest -Uri $Url -Method Get -TimeoutSec 5
    if ($response.StatusCode -lt 200 -or $response.StatusCode -ge 300) {
        throw "Unexpected HTTP status $($response.StatusCode) from $Url"
    }
    if ($response.Content -notmatch '<div id="root">') {
        throw "Frontend responded, but it does not look like the RadishMind console shell."
    }
}

function Invoke-CorsProbe {
    param(
        [string]$OverviewUrl,
        [string]$Origin
    )
    $headers = @{
        Origin = $Origin
        "Access-Control-Request-Method" = "GET"
    }
    $response = Invoke-WebRequest -Uri $OverviewUrl -Method Options -Headers $headers -TimeoutSec 5
    $allowOrigin = $response.Headers["Access-Control-Allow-Origin"]
    if ($allowOrigin -ne $Origin) {
        throw "CORS preflight did not allow $Origin. actual Access-Control-Allow-Origin=$allowOrigin"
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
    [Console]::Error.WriteLine("RadishMind console dev entry failed:")
    [Console]::Error.WriteLine($Message)
    [Console]::Error.WriteLine("")
    [Console]::Error.WriteLine("Common local failures:")
    [Console]::Error.WriteLine("- Port conflict: backend must answer on http://127.0.0.1:7000 and frontend on http://127.0.0.1:4000. Stop the other process or change the existing service back to the defaults.")
    [Console]::Error.WriteLine("- CORS/preflight: the platform service only allows http://127.0.0.1:4000 and http://localhost:4000 as local console origins.")
    [Console]::Error.WriteLine("- Browser unsafe port: ERR_UNSAFE_PORT means the browser blocked the port before reaching the app. Keep backend 7000 and frontend 4000.")
    [Console]::Error.WriteLine("- Missing dependencies: run npm install in apps/radishmind-console if npm cannot start Vite.")
    [Console]::Error.WriteLine("- Backend diagnostics: run pwsh ./scripts/run-platform-service.ps1 -Command diagnostics from the repo root.")
    [Console]::Error.WriteLine("- Logs: $logDir")
}

$backendUri = Get-HttpUri -Value $BackendUrl
$frontendUri = Get-HttpUri -Value $FrontendUrl
Assert-BrowserSafeLocalUrl -Uri $backendUri
Assert-BrowserSafeLocalUrl -Uri $frontendUri

$healthzUrl = "$BackendUrl/healthz"
$overviewUrl = "$BackendUrl/v1/platform/overview"
$localSmokeUrl = "$BackendUrl/v1/platform/local-smoke"
$frontendOrigin = $FrontendUrl.TrimEnd("/")

try {
    if (-not (Test-Path -LiteralPath $platformWrapper -PathType Leaf)) {
        throw "Missing platform wrapper: $platformWrapper"
    }
    if (-not (Test-Path -LiteralPath (Join-Path $consoleDir "package.json") -PathType Leaf)) {
        throw "Missing RadishMind console package.json: $consoleDir"
    }

    if (-not $VerifyOnly) {
        $pwshPath = Get-RequiredCommand -Names @("pwsh", "powershell")
        $npmPath = Get-RequiredCommand -Names @("npm.cmd", "npm")

        if (-not $env:RADISHMIND_PLATFORM_LISTEN_ADDR) {
            $env:RADISHMIND_PLATFORM_LISTEN_ADDR = "$($backendUri.Host):$($backendUri.Port)"
        }
        if (-not $env:RADISHMIND_PLATFORM_PROVIDER) {
            $env:RADISHMIND_PLATFORM_PROVIDER = "mock"
        }
        if (-not $env:RADISHMIND_PLATFORM_MODEL) {
            $env:RADISHMIND_PLATFORM_MODEL = "radishmind-local-dev"
        }
        if (-not $env:VITE_RADISHMIND_PLATFORM_BASE_URL) {
            $env:VITE_RADISHMIND_PLATFORM_BASE_URL = $BackendUrl
        }

        $backendPortOpen = Test-TcpPort -HostName $backendUri.Host -Port $backendUri.Port
        if ($backendPortOpen) {
            Write-Step "Backend port $($backendUri.Port) is already open; reusing it if probes pass."
        }
        else {
            Start-LoggedProcess `
                -Name "platform" `
                -FilePath $pwshPath `
                -Arguments @("-NoProfile", "-ExecutionPolicy", "Bypass", "-File", $platformWrapper, "-Command", "serve") `
                -WorkingDirectory $repoRoot | Out-Null
        }

        $frontendPortOpen = Test-TcpPort -HostName $frontendUri.Host -Port $frontendUri.Port
        if ($frontendPortOpen) {
            Write-Step "Frontend port $($frontendUri.Port) is already open; reusing it if probes pass."
        }
        else {
            Start-LoggedProcess `
                -Name "console" `
                -FilePath $npmPath `
                -Arguments @("run", "dev") `
                -WorkingDirectory $consoleDir | Out-Null
        }
    }

    Wait-Until -Name "backend healthz" -Probe { Invoke-JsonProbe -Url $healthzUrl -ExpectedKind "" | Out-Null }
    Wait-Until -Name "platform overview" -Probe { Invoke-JsonProbe -Url $overviewUrl -ExpectedKind "platform_overview" | Out-Null }
    Wait-Until -Name "platform local-smoke" -Probe { Invoke-JsonProbe -Url $localSmokeUrl -ExpectedKind "platform_local_smoke" | Out-Null }
    Wait-Until -Name "local console CORS" -Probe { Invoke-CorsProbe -OverviewUrl $overviewUrl -Origin $frontendOrigin }
    Wait-Until -Name "frontend console" -Probe { Invoke-PageProbe -Url $FrontendUrl }

    Write-Step "Local console is ready: $FrontendUrl"
    Write-Step "Backend probes passed: $healthzUrl ; $overviewUrl ; $localSmokeUrl"
    Write-Step "This is a dev-only launcher, not a production supervisor. It does not implement executor, durable store, confirmation, writeback, or replay."

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
