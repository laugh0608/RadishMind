[CmdletBinding()]
param(
    [ValidateSet("up", "status", "migrate", "test", "check", "down", "help")]
    [string]$Action = "check"
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$platformDir = Join-Path $repoRoot "services/platform"
$composeFile = Join-Path $repoRoot "deploy/docker-compose.saved-draft-dev-test.yaml"

$pgHost = if ($env:PGHOST) { $env:PGHOST } else { "127.0.0.1" }
$pgPort = if ($env:PGPORT) { $env:PGPORT } else { "55432" }
$pgDatabase = if ($env:PGDATABASE) { $env:PGDATABASE } else { "radishmind_saved_draft_test" }
$pgSslMode = if ($env:PGSSLMODE) { $env:PGSSLMODE } else { "disable" }
$runtimeUser = if ($env:PGUSER) { $env:PGUSER } else { "radishmind_runtime" }
$runtimePassword = if ($env:PGPASSWORD) { $env:PGPASSWORD } else { "" }
$migrationUser = if ($env:RADISHMIND_SAVED_DRAFT_MIGRATION_USER) { $env:RADISHMIND_SAVED_DRAFT_MIGRATION_USER } else { "radishmind_migrator" }
$migrationPassword = if ($env:RADISHMIND_SAVED_DRAFT_MIGRATION_PASSWORD) { $env:RADISHMIND_SAVED_DRAFT_MIGRATION_PASSWORD } else { "" }

function Write-Usage {
@"
Usage: pwsh ./scripts/run-workflow-saved-draft-postgres-dev-test.ps1 [-Action ACTION]

Actions:
  up       Start the local PostgreSQL 17 dev/test container and wait for health.
  status   Inspect the saved draft migration marker without changing the database.
  migrate  Apply the reviewed saved draft migration through the one-time runner.
  test     Run the destructive integration suite against a database whose name contains "test".
  check    Start PostgreSQL, run integration tests, then leave the schema migrated and ready.
  down     Stop the local container while preserving its named volume.

The platform runtime connection comes from PGHOST, PGPORT, PGUSER, PGDATABASE,
PGPASSWORD and PGSSLMODE. The one-time migration identity comes from
RADISHMIND_SAVED_DRAFT_MIGRATION_USER / _PASSWORD. Defaults target the repository's
loopback-only dev/test Compose service and keep runtime DML separate from migration
DDL. The script assembles both database URLs in memory and never prints them.
"@
}

function Write-Step {
    param([string]$Message)
    Write-Host "[saved-draft-postgres-dev-test] $Message"
}

function Get-RequiredCommand {
    param([string]$Name)
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if ($null -eq $command) {
        throw "Missing required command: $Name"
    }
    return $command.Source
}

function Get-DatabaseUrl {
    param(
        [string]$DatabaseUser,
        [string]$DatabasePassword
    )
    $encodedUser = [Uri]::EscapeDataString($DatabaseUser)
    $userInfo = $encodedUser
    if ($DatabasePassword) {
        $encodedPassword = [Uri]::EscapeDataString($DatabasePassword)
        $userInfo = "${encodedUser}:${encodedPassword}"
    }
    $hostText = $pgHost
    if ($hostText.Contains(":") -and -not $hostText.StartsWith("[")) {
        $hostText = "[$hostText]"
    }
    $encodedDatabase = [Uri]::EscapeDataString($pgDatabase)
    $encodedSslMode = [Uri]::EscapeDataString($pgSslMode)
    return "postgresql://${userInfo}@${hostText}:${pgPort}/${encodedDatabase}?sslmode=${encodedSslMode}"
}

function Invoke-Compose {
    param([string[]]$Arguments)
    $docker = Get-RequiredCommand -Name "docker"
    & $docker compose -f $composeFile @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose failed with exit code $LASTEXITCODE"
    }
}

function Set-DatabaseComponentEnvironment {
    param(
        [string]$DatabaseUser,
        [string]$DatabasePassword
    )
    $env:PGHOST = $pgHost
    $env:PGPORT = $pgPort
    $env:PGUSER = $DatabaseUser
    $env:PGPASSWORD = $DatabasePassword
    $env:PGDATABASE = $pgDatabase
    $env:PGSSLMODE = $pgSslMode
}

function Invoke-Migration {
    param([ValidateSet("status", "up")][string]$MigrationAction)
    $go = Get-RequiredCommand -Name "go"
    Set-DatabaseComponentEnvironment -DatabaseUser $runtimeUser -DatabasePassword $runtimePassword
    $env:RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH = "1"
    $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP = "1"
    $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE = "1"
    $env:RADISHMIND_WORKFLOW_EXECUTOR_DEV = "1"
    $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE = "postgres_dev_test"
    $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $runtimeUser -DatabasePassword $runtimePassword
    $env:RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $migrationUser -DatabasePassword $migrationPassword
    $env:RADISHMIND_APPLICATION_DRAFT_DEV_HTTP = "1"
    $env:RADISHMIND_APPLICATION_DRAFT_DEV_WRITE = "1"
    $env:RADISHMIND_APPLICATION_DRAFT_STORE = "postgres_dev_test"
    $env:RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $runtimeUser -DatabasePassword $runtimePassword
    $env:RADISHMIND_APPLICATION_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $migrationUser -DatabasePassword $migrationPassword
    $env:RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP = "1"
    $env:RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE = "1"
    $env:RADISHMIND_APPLICATION_PUBLISH_STORE = "postgres_dev_test"
    $env:RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $runtimeUser -DatabasePassword $runtimePassword
    $env:RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_MIGRATION_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $migrationUser -DatabasePassword $migrationPassword
    $env:RADISHMIND_WORKFLOW_RUN_STORE = "postgres_dev_test"
    $env:RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $runtimeUser -DatabasePassword $runtimePassword
    $env:RADISHMIND_WORKFLOW_RUN_DEV_TEST_MIGRATION_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $migrationUser -DatabasePassword $migrationPassword
    $env:RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV = "1"
    $env:RADISHMIND_GATEWAY_REQUEST_STORE = "postgres_dev_test"
    $env:RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $runtimeUser -DatabasePassword $runtimePassword
    $env:RADISHMIND_GATEWAY_REQUEST_DEV_TEST_MIGRATION_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $migrationUser -DatabasePassword $migrationPassword
    Push-Location $platformDir
    try {
        & $go run ./cmd/radishmind-workflow-draft-migrate $MigrationAction
        if ($LASTEXITCODE -ne 0) {
            throw "saved draft migration runner failed with exit code $LASTEXITCODE"
        }
        & $go run ./cmd/radishmind-application-draft-migrate $MigrationAction
        if ($LASTEXITCODE -ne 0) {
            throw "application draft migration runner failed with exit code $LASTEXITCODE"
        }
        & $go run ./cmd/radishmind-application-publish-migrate $MigrationAction
        if ($LASTEXITCODE -ne 0) {
            throw "application publish migration runner failed with exit code $LASTEXITCODE"
        }
        & $go run ./cmd/radishmind-workflow-run-migrate $MigrationAction
        if ($LASTEXITCODE -ne 0) {
            throw "workflow run migration runner failed with exit code $LASTEXITCODE"
        }
        & $go run ./cmd/radishmind-gateway-request-migrate $MigrationAction
        if ($LASTEXITCODE -ne 0) {
            throw "Gateway request migration runner failed with exit code $LASTEXITCODE"
        }
        $env:RADISHMIND_CONTROL_PLANE_READ_STORE = "postgres_dev_test"
        $env:RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $runtimeUser -DatabasePassword $runtimePassword
        $env:RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_MIGRATION_DATABASE_URL = Get-DatabaseUrl -DatabaseUser $migrationUser -DatabasePassword $migrationPassword
        & $go run ./cmd/radishmind-control-plane-read-migrate $MigrationAction
        if ($LASTEXITCODE -ne 0) {
            throw "Control Plane Admin read migration runner failed with exit code $LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }
}

function Invoke-IntegrationTest {
    $go = Get-RequiredCommand -Name "go"
    Set-DatabaseComponentEnvironment -DatabaseUser $migrationUser -DatabasePassword $migrationPassword
    $env:RADISHMIND_RUN_POSTGRES_INTEGRATION = "1"
    $env:RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER = $runtimeUser
    $env:RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD = $runtimePassword
    Push-Location $platformDir
    try {
        & $go test -count=1 -tags=postgres_integration ./internal/httpapi
        if ($LASTEXITCODE -ne 0) {
            throw "saved draft PostgreSQL integration test failed with exit code $LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }
}

switch ($Action) {
    "up" {
        Write-Step "Starting loopback-only PostgreSQL dev/test service."
        Invoke-Compose -Arguments @("up", "-d", "--wait")
    }
    "status" {
        Write-Step "Inspecting the saved draft schema marker."
        Invoke-Migration -MigrationAction "status"
    }
    "migrate" {
        Write-Step "Applying the reviewed saved draft migration."
        Invoke-Migration -MigrationAction "up"
    }
    "test" {
        Write-Step "Running the PostgreSQL repository integration suite."
        Invoke-IntegrationTest
    }
    "check" {
        Write-Step "Starting loopback-only PostgreSQL dev/test service."
        Invoke-Compose -Arguments @("up", "-d", "--wait")
        Write-Step "Running the PostgreSQL repository integration suite."
        Invoke-IntegrationTest
        Write-Step "Restoring the reviewed saved draft, application draft, application publish, workflow run, Gateway request, and Control Plane Admin read schemas for interactive development."
        Invoke-Migration -MigrationAction "up"
    }
    "down" {
        Write-Step "Stopping PostgreSQL while preserving its named volume."
        Invoke-Compose -Arguments @("down")
    }
    "help" {
        Write-Usage
    }
}
