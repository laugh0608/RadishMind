#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

python_bin="${repo_root}/.venv/bin/python"
if [[ ! -x "${python_bin}" ]]; then
  python_bin="python3"
fi

if ! command -v "${python_bin}" >/dev/null 2>&1; then
  echo "python3 or .venv/bin/python is required for scripts/check-repo.sh" >&2
  exit 1
fi

"${python_bin}" "${repo_root}/scripts/check-repo.py" "$@"
