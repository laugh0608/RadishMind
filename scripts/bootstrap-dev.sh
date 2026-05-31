#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
venv_python="${repo_root}/.venv/bin/python"

if [[ ! -x "${venv_python}" ]]; then
  if ! command -v python3 >/dev/null 2>&1; then
    echo "python3 is required to create ${repo_root}/.venv" >&2
    exit 1
  fi
  python3 -m venv "${repo_root}/.venv"
fi

"${venv_python}" -m pip install --upgrade pip
"${venv_python}" -m pip install -r "${repo_root}/requirements-dev.txt"

echo "RadishMind development virtual environment is ready: ${venv_python}"
