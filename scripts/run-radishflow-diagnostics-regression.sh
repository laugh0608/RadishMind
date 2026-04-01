#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

if ! command -v pwsh >/dev/null 2>&1; then
  echo "pwsh is required for scripts/run-radishflow-diagnostics-regression.sh" >&2
  exit 1
fi

pwsh "${repo_root}/scripts/run-radishflow-diagnostics-regression.ps1" "$@"
