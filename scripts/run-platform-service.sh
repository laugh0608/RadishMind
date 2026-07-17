#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
platform_dir="${repo_root}/services/platform"

profile="local-product"
command=""
remaining_args=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile)
      if [[ $# -lt 2 ]]; then
        echo "missing value for --profile" >&2
        exit 2
      fi
      profile="$2"
      shift 2
      ;;
    --profile=*)
      profile="${1#*=}"
      shift
      ;;
    *)
      if [[ -z "${command}" ]]; then
        command="$1"
      else
        remaining_args+=("$1")
      fi
      shift
      ;;
  esac
done

command="${command:-serve}"

set_default_env() {
  local name="$1"
  local value="$2"
  if [[ -z "${!name:-}" ]]; then
    export "${name}=${value}"
  fi
}

case "${profile}" in
  local-product)
    if [[ -n "${RADISHMIND_LOCAL_PERSISTENCE_MODE:-}" && "${RADISHMIND_LOCAL_PERSISTENCE_MODE}" != "sqlite_dev" ]]; then
      echo "local-product profile requires RADISHMIND_LOCAL_PERSISTENCE_MODE=sqlite_dev; use --profile configured for explicit component configuration" >&2
      exit 2
    fi
    export RADISHMIND_LOCAL_PERSISTENCE_MODE="sqlite_dev"
    set_default_env RADISHMIND_SQLITE_DEV_DATABASE_PATH "${repo_root}/var/sqlite-dev/radishmind.db"
    set_default_env RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH "1"
    set_default_env RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP "1"
    set_default_env RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE "1"
    set_default_env RADISHMIND_APPLICATION_DRAFT_DEV_HTTP "1"
    set_default_env RADISHMIND_APPLICATION_DRAFT_DEV_WRITE "1"
    set_default_env RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP "1"
    set_default_env RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE "1"
    set_default_env RADISHMIND_APPLICATION_CATALOG_DEV_HTTP "1"
    set_default_env RADISHMIND_APPLICATION_CATALOG_DEV_WRITE "1"
    set_default_env RADISHMIND_API_KEY_LIFECYCLE_DEV_HTTP "1"
    set_default_env RADISHMIND_API_KEY_LIFECYCLE_DEV_WRITE "1"
    set_default_env RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV "1"
    set_default_env RADISHMIND_WORKFLOW_EXECUTOR_DEV "1"
    set_default_env RADISHMIND_WORKFLOW_TOOL_ACTION_DEV "1"
    set_default_env RADISHMIND_WORKFLOW_HTTP_TOOL_EXECUTION_DEV "1"
    set_default_env RADISHMIND_WORKFLOW_RAG_SNAPSHOT_DEV "1"
    ;;
  configured)
    ;;
  *)
    echo "unsupported platform startup profile: ${profile}" >&2
    echo "supported profiles: local-product, configured" >&2
    exit 2
    ;;
esac

if ! command -v go >/dev/null 2>&1; then
  echo "go is required for scripts/run-platform-service.sh" >&2
  exit 1
fi

export GOCACHE="${GOCACHE:-/tmp/radishmind-go-build-cache}"
if [[ -z "${RADISHMIND_PLATFORM_PYTHON_BIN:-}" ]]; then
  if [[ -x "${repo_root}/.venv/bin/python" ]]; then
    export RADISHMIND_PLATFORM_PYTHON_BIN="${repo_root}/.venv/bin/python"
  else
    echo "missing repository virtual environment: ${repo_root}/.venv/bin/python" >&2
    echo "run ./scripts/bootstrap-dev.sh before running scripts/run-platform-service.sh" >&2
    exit 1
  fi
fi

if [[ -z "${RADISHMIND_PLATFORM_CONFIG:-}" ]]; then
  default_config="${repo_root}/tmp/radishmind-platform.local.json"
  if [[ -f "${default_config}" ]]; then
    export RADISHMIND_PLATFORM_CONFIG="${default_config}"
  fi
fi

cd "${platform_dir}"

case "${command}" in
  serve)
    if [[ "${#remaining_args[@]}" -gt 0 ]]; then
      exec go run ./cmd/radishmind-platform "${remaining_args[@]}"
    fi
    exec go run ./cmd/radishmind-platform
    ;;
  config-summary | config-check | diagnostics)
    if [[ "${#remaining_args[@]}" -gt 0 ]]; then
      exec go run ./cmd/radishmind-platform "${command}" "${remaining_args[@]}"
    fi
    exec go run ./cmd/radishmind-platform "${command}"
    ;;
  *)
    echo "unsupported platform service command: ${command}" >&2
    echo "supported commands: serve, config-summary, config-check, diagnostics" >&2
    exit 2
    ;;
esac
