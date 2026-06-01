#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

python_bin="${repo_root}/.venv/bin/python"
if [[ ! -x "${python_bin}" ]]; then
  echo "missing repository virtual environment: ${python_bin}" >&2
  echo "run ./scripts/bootstrap-dev.sh before running scripts/check-repo.sh" >&2
  exit 1
fi

"${python_bin}" "${repo_root}/scripts/check-repo.py" "$@"
