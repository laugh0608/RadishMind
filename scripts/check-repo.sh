#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

skip_text_files=0

for arg in "$@"; do
  case "$arg" in
    --skip-text-files)
      skip_text_files=1
      ;;
    *)
      echo "unsupported argument: $arg" >&2
      exit 2
      ;;
  esac
done

if [[ "$skip_text_files" -eq 0 ]]; then
  bash "${repo_root}/scripts/check-text-files.sh"
fi

required_files=(
  "AGENTS.md"
  "LICENSE"
  ".editorconfig"
  ".gitattributes"
  ".github/PULL_REQUEST_TEMPLATE.md"
  ".github/rulesets/README.md"
  ".github/rulesets/master-protection.json"
  ".github/workflows/pr-check.yml"
  ".github/workflows/release-check.yml"
  "README.md"
  "docs/README.md"
  "docs/task-cards/README.md"
  "docs/task-cards/radishflow-explain-diagnostics.md"
  "docs/task-cards/radishflow-suggest-flowsheet-edits.md"
  "docs/task-cards/radishflow-explain-control-plane-state.md"
  "docs/task-cards/radish-answer-docs-question.md"
  "docs/radishmind-product-scope.md"
  "docs/radishmind-architecture.md"
  "docs/radishmind-roadmap.md"
  "docs/radishmind-integration-contracts.md"
  "docs/adr/0001-branch-and-pr-governance.md"
  "docs/devlogs/README.md"
  "docs/devlogs/2026-W14.md"
  "contracts/README.md"
  "contracts/copilot-request.schema.json"
  "contracts/copilot-response.schema.json"
  "datasets/README.md"
  "datasets/eval/README.md"
  "datasets/eval/radishflow-task-sample.schema.json"
  "datasets/eval/radish-task-sample.schema.json"
  "scripts/check-text-files.ps1"
  "scripts/check-text-files.sh"
  "scripts/check-repo.ps1"
  "scripts/check-repo.sh"
)

for relative_path in "${required_files[@]}"; do
  if [[ ! -f "${repo_root}/${relative_path}" ]]; then
    echo "missing required file: ${relative_path}" >&2
    exit 1
  fi
done

grep -Fq '当前常态开发分支为 `dev`' "${repo_root}/AGENTS.md" || {
  echo "AGENTS.md does not mention dev as the default development branch" >&2
  exit 1
}

grep -Fq '默认目标分支为 `dev`' "${repo_root}/.github/PULL_REQUEST_TEMPLATE.md" || {
  echo "PULL_REQUEST_TEMPLATE.md does not mention dev as the default target branch" >&2
  exit 1
}

for pattern in '"refs/heads/master"' '"context": "Repo Hygiene"' '"context": "Planning Baseline"'; do
  grep -Fq "$pattern" "${repo_root}/.github/rulesets/master-protection.json" || {
    echo ".github/rulesets/master-protection.json is missing expected content: ${pattern}" >&2
    exit 1
  }
done

for pattern in 'name: PR Checks' '- master' 'name: Repo Hygiene' 'name: Planning Baseline'; do
  grep -Fq -- "$pattern" "${repo_root}/.github/workflows/pr-check.yml" || {
    echo ".github/workflows/pr-check.yml is missing expected content: ${pattern}" >&2
    exit 1
  }
done

for pattern in 'name: Release Checks' 'v*-dev' 'v*-test' 'v*-release' 'name: Release Repo Hygiene' 'name: Release Planning Baseline'; do
  grep -Fq -- "$pattern" "${repo_root}/.github/workflows/release-check.yml" || {
    echo ".github/workflows/release-check.yml is missing expected content: ${pattern}" >&2
    exit 1
  }
done

echo "repository baseline checks passed."

