#!/usr/bin/env bash
set -euo pipefail

action="${1:-check}"
if [[ "$#" -gt 1 ]]; then
  echo "Usage: $0 [up|status|migrate|test|check|down]" >&2
  exit 2
fi

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
platform_dir="${repo_root}/services/platform"
compose_file="${repo_root}/deploy/docker-compose.saved-draft-dev-test.yaml"

pg_host="${PGHOST:-127.0.0.1}"
pg_port="${PGPORT:-55432}"
pg_database="${PGDATABASE:-radishmind_saved_draft_test}"
pg_sslmode="${PGSSLMODE:-disable}"
runtime_user="${PGUSER:-radishmind_runtime}"
runtime_password="${PGPASSWORD:-}"
migration_user="${RADISHMIND_SAVED_DRAFT_MIGRATION_USER:-radishmind_migrator}"
migration_password="${RADISHMIND_SAVED_DRAFT_MIGRATION_PASSWORD:-}"

usage() {
  cat <<'EOF'
Usage: scripts/run-workflow-saved-draft-postgres-dev-test.sh [action]

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
EOF
}

step() {
  echo "[saved-draft-postgres-dev-test] $*"
}

require_command() {
  local command_name="$1"
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "Missing required command: ${command_name}" >&2
    exit 1
  fi
}

build_database_url() {
  local database_user="$1"
  local database_password="$2"
  local python_bin="${repo_root}/.venv/bin/python"
  if [[ ! -x "${python_bin}" ]]; then
    echo "Missing repository virtual environment: ${python_bin}" >&2
    echo "Run ./scripts/bootstrap-dev.sh first." >&2
    exit 1
  fi
  PGHOST="${pg_host}" \
  PGPORT="${pg_port}" \
  PGUSER="${database_user}" \
  PGPASSWORD="${database_password}" \
  PGDATABASE="${pg_database}" \
  PGSSLMODE="${pg_sslmode}" \
  "${python_bin}" - <<'PY'
import os
from urllib.parse import quote, urlencode

host = os.environ["PGHOST"]
port = os.environ["PGPORT"]
user = quote(os.environ["PGUSER"], safe="")
password = os.environ.get("PGPASSWORD", "")
userinfo = user if not password else f"{user}:{quote(password, safe='')}"
database = quote(os.environ["PGDATABASE"], safe="")
query = urlencode({"sslmode": os.environ["PGSSLMODE"]})
if ":" in host and not host.startswith("["):
    host = f"[{host}]"
print(f"postgresql://{userinfo}@{host}:{port}/{database}?{query}")
PY
}

compose() {
  require_command docker
  docker compose -f "${compose_file}" "$@"
}

run_migration() {
  local migration_action="$1"
  local runtime_database_url
  local migration_database_url
  require_command go
  runtime_database_url="$(build_database_url "${runtime_user}" "${runtime_password}")"
  migration_database_url="$(build_database_url "${migration_user}" "${migration_password}")"
  (
    export RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH="1"
    export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP="1"
    export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE="1"
		export RADISHMIND_WORKFLOW_EXECUTOR_DEV="1"
    export RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE="postgres_dev_test"
    export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL="${runtime_database_url}"
    export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL="${migration_database_url}"
		export RADISHMIND_APPLICATION_DRAFT_DEV_HTTP="1"
		export RADISHMIND_APPLICATION_DRAFT_DEV_WRITE="1"
		export RADISHMIND_APPLICATION_DRAFT_STORE="postgres_dev_test"
		export RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL="${runtime_database_url}"
		export RADISHMIND_APPLICATION_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL="${migration_database_url}"
		export RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP="1"
		export RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE="1"
		export RADISHMIND_APPLICATION_PUBLISH_STORE="postgres_dev_test"
		export RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL="${runtime_database_url}"
		export RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_MIGRATION_DATABASE_URL="${migration_database_url}"
    export RADISHMIND_WORKFLOW_RUN_STORE="postgres_dev_test"
		export RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL="${runtime_database_url}"
		export RADISHMIND_WORKFLOW_RUN_DEV_TEST_MIGRATION_DATABASE_URL="${migration_database_url}"
		export RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV="1"
		export RADISHMIND_GATEWAY_REQUEST_STORE="postgres_dev_test"
		export RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL="${runtime_database_url}"
		export RADISHMIND_GATEWAY_REQUEST_DEV_TEST_MIGRATION_DATABASE_URL="${migration_database_url}"
    cd "${platform_dir}"
    go run ./cmd/radishmind-workflow-draft-migrate "${migration_action}"
		go run ./cmd/radishmind-application-draft-migrate "${migration_action}"
		go run ./cmd/radishmind-application-publish-migrate "${migration_action}"
		go run ./cmd/radishmind-workflow-run-migrate "${migration_action}"
		go run ./cmd/radishmind-gateway-request-migrate "${migration_action}"
		export RADISHMIND_CONTROL_PLANE_READ_STORE="postgres_dev_test"
		export RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_DATABASE_URL="${runtime_database_url}"
		export RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_MIGRATION_DATABASE_URL="${migration_database_url}"
		go run ./cmd/radishmind-control-plane-read-migrate "${migration_action}"
  )
}

run_integration_test() {
  require_command go
  (
    export RADISHMIND_RUN_POSTGRES_INTEGRATION="1"
    export PGHOST="${pg_host}"
    export PGPORT="${pg_port}"
    export PGUSER="${migration_user}"
    export PGPASSWORD="${migration_password}"
    export PGDATABASE="${pg_database}"
    export PGSSLMODE="${pg_sslmode}"
    export RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER="${runtime_user}"
    export RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD="${runtime_password}"
    cd "${platform_dir}"
    go test -count=1 -tags=postgres_integration ./internal/httpapi
  )
}

case "${action}" in
  up)
    step "Starting loopback-only PostgreSQL dev/test service."
    compose up -d --wait
    ;;
  status)
    step "Inspecting the saved draft schema marker."
    run_migration status
    ;;
  migrate)
    step "Applying the reviewed saved draft migration."
    run_migration up
    ;;
  test)
    step "Running the PostgreSQL repository integration suite."
    run_integration_test
    ;;
  check)
    step "Starting loopback-only PostgreSQL dev/test service."
    compose up -d --wait
    step "Running the PostgreSQL repository integration suite."
    run_integration_test
    step "Restoring the reviewed workflow draft, application draft, application publish, workflow run, Gateway request, and Control Plane Admin read schemas for interactive development."
    run_migration up
    ;;
  down)
    step "Stopping PostgreSQL while preserving its named volume."
    compose down
    ;;
  -h | --help | help)
    usage
    ;;
  *)
    echo "Unsupported action: ${action}" >&2
    usage >&2
    exit 2
    ;;
esac
