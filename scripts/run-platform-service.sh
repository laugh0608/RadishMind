#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
platform_dir="${repo_root}/services/platform"

if ! command -v go >/dev/null 2>&1; then
  echo "go is required for scripts/run-platform-service.sh" >&2
  exit 1
fi

export GOCACHE="${GOCACHE:-/tmp/radishmind-go-build-cache}"
if [[ -z "${RADISHMIND_PLATFORM_CONFIG:-}" ]]; then
  default_config="${repo_root}/tmp/radishmind-platform.local.json"
  if [[ -f "${default_config}" ]]; then
    export RADISHMIND_PLATFORM_CONFIG="${default_config}"
  fi
fi

command="${1:-serve}"
shift || true

cd "${platform_dir}"

case "${command}" in
  serve)
    exec go run ./cmd/radishmind-platform "$@"
    ;;
  config-summary | config-check)
    exec go run ./cmd/radishmind-platform "${command}" "$@"
    ;;
  *)
    echo "unsupported platform service command: ${command}" >&2
    echo "supported commands: serve, config-summary, config-check" >&2
    exit 2
    ;;
esac
