#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for scripts/check-text-files.sh" >&2
  exit 1
fi

python3 "${repo_root}/scripts/check-text-files.py"
