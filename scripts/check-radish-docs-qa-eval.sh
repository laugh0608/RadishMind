#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
sample_dir="${repo_root}/datasets/eval/radish"
schema_file="${repo_root}/datasets/eval/radish-task-sample.schema.json"

expected_samples=(
  "answer-docs-question-direct-answer-001.json"
  "answer-docs-question-evidence-gap-001.json"
  "answer-docs-question-navigation-001.json"
)

[[ -f "$schema_file" ]] || {
  echo "missing schema file: datasets/eval/radish-task-sample.schema.json" >&2
  exit 1
}

[[ -d "$sample_dir" ]] || {
  echo "missing sample directory: datasets/eval/radish" >&2
  exit 1
}

for sample in "${expected_samples[@]}"; do
  sample_path="${sample_dir}/${sample}"
  [[ -f "$sample_path" ]] || {
    echo "missing expected sample: datasets/eval/radish/${sample}" >&2
    exit 1
  }

  for pattern in '"project": "radish"' '"task": "answer_docs_question"' '"golden_response"' '"summary"' '"citations"' '"risk_level"'; do
    grep -Fq "$pattern" "$sample_path" || {
      echo "${sample}: missing expected content: ${pattern}" >&2
      exit 1
    }
  done
done

grep -Fq '"kind": "read_only_check"' "${sample_dir}/answer-docs-question-navigation-001.json" || {
  echo "answer-docs-question-navigation-001.json: missing read_only_check action" >&2
  exit 1
}

echo "radish docs qa eval checks passed."
