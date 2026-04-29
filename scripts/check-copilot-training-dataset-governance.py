#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
MANIFEST_PATH = REPO_ROOT / "training/datasets/copilot-training-dataset-governance-v0.json"
EXPECTED_TASKS = {
    "suggest_flowsheet_edits",
    "suggest_ghost_completion",
    "answer_docs_question",
}
EXPECTED_SOURCES = {"golden_response", "teacher_capture"}


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse {path.relative_to(REPO_ROOT)}: {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_existing_file(path_value: str, *, field_name: str) -> None:
    require(path_value.strip() == path_value and path_value.strip(), f"{field_name} must be a non-empty clean path")
    path = REPO_ROOT / path_value
    require(path.is_file(), f"{field_name} points to a missing file: {path_value}")


def check_source_conversion_manifests(document: dict[str, Any]) -> None:
    entries = document.get("source_conversion_manifests")
    require(isinstance(entries, list) and len(entries) == 2, "source_conversion_manifests must include two entries")
    seen_sources: set[str] = set()
    for entry in entries:
        require(isinstance(entry, dict), "source conversion entry must be an object")
        source = str(entry.get("source") or "")
        seen_sources.add(source)
        require(source in EXPECTED_SOURCES, f"unexpected source conversion entry: {source}")
        require_existing_file(str(entry.get("manifest") or ""), field_name=f"{source}.manifest")
        require_existing_file(str(entry.get("summary") or ""), field_name=f"{source}.summary")
        summary = load_json(REPO_ROOT / str(entry["summary"]))
        require(summary.get("source") == source, f"{source}.summary source mismatch")
        require(summary.get("does_not_run_models") is True, f"{source}.summary must not run models")
        require(summary.get("sample_count") == entry.get("current_sample_count"), f"{source}.sample_count mismatch")
    require(seen_sources == EXPECTED_SOURCES, "source conversion manifests must cover golden_response and teacher_capture")


def check_candidate_record_eligibility(document: dict[str, Any]) -> None:
    eligibility = document.get("candidate_record_eligibility")
    require(isinstance(eligibility, dict), "candidate_record_eligibility must be an object")
    allow = eligibility.get("may_enter_training_set_when_all_true")
    deny = eligibility.get("must_not_enter_training_set_when_any_true")
    require(isinstance(allow, list) and len(allow) >= 8, "candidate record allow list is too small")
    require(isinstance(deny, list) and len(deny) >= 8, "candidate record deny list is too small")
    for required_item in (
        "selected_sample_status_is_pass",
        "record_project_task_sample_id_match_committed_eval_sample",
        "risk_and_confirmation_boundaries_are_preserved",
    ):
        require(required_item in allow, f"candidate record allow list is missing {required_item}")
    for rejected_item in (
        "audit_status_failed_or_missing",
        "schema_invalid_record_or_response",
        "requires_confirmation_weakened_or_removed",
        "provider_raw_dump_without_committed_candidate_record",
    ):
        require(rejected_item in deny, f"candidate record deny list is missing {rejected_item}")


def check_sampling_policy(document: dict[str, Any]) -> None:
    sampling = document.get("sampling_policy")
    require(isinstance(sampling, dict), "sampling_policy must be an object")
    seed_review = sampling.get("current_seed_set_review")
    larger_review = sampling.get("larger_candidate_record_pool_review")
    require(isinstance(seed_review, dict), "current_seed_set_review must be an object")
    require(seed_review.get("manual_review_ratio") == 1.0, "current seed set must require full review")
    require(isinstance(larger_review, dict), "larger_candidate_record_pool_review must be an object")
    ratio = larger_review.get("default_manual_review_ratio")
    require(isinstance(ratio, (int, float)) and 0 < ratio <= 1, "default manual review ratio must be in (0, 1]")
    require(ratio >= 0.2, "default manual review ratio must not drop below 20 percent")
    require(int(larger_review.get("minimum_reviewed_records_per_stratum") or 0) >= 5, "minimum reviewed records per stratum is too low")
    full_review_required_for = larger_review.get("full_review_required_for")
    require(isinstance(full_review_required_for, list), "full_review_required_for must be a list")
    for required_item in (
        "new_project_or_task",
        "high_risk_actions",
        "requires_confirmation_true",
        "schema_or_contract_version_change",
    ):
        require(required_item in full_review_required_for, f"full review list is missing {required_item}")
    holdout = larger_review.get("holdout_policy")
    require(isinstance(holdout, dict), "holdout_policy must be an object")
    holdout_ratio = holdout.get("offline_eval_holdout_ratio")
    require(isinstance(holdout_ratio, (int, float)) and holdout_ratio >= 0.1, "offline eval holdout ratio is too low")
    require(int(holdout.get("minimum_holdout_records_per_task") or 0) >= 3, "minimum holdout records per task is too low")


def check_artifact_policy(document: dict[str, Any]) -> None:
    artifact_policy = document.get("artifact_policy")
    require(isinstance(artifact_policy, dict), "artifact_policy must be an object")
    allowed = artifact_policy.get("committed_allowed")
    disallowed = artifact_policy.get("committed_disallowed")
    require(isinstance(allowed, list) and "manifest" in allowed and "review_record" in allowed, "committed allowed list is incomplete")
    require(isinstance(disallowed, list), "committed disallowed list must be a list")
    for blocked in ("large_jsonl", "model_weights", "checkpoint", "provider_raw_dump"):
        require(blocked in disallowed, f"artifact policy must disallow {blocked}")
    require(artifact_policy.get("default_generated_jsonl_location") == "tmp/", "generated JSONL must default to tmp/")


def main() -> int:
    document = load_json(MANIFEST_PATH)
    require(isinstance(document, dict), "training dataset governance manifest must be an object")
    require(document.get("schema_version") == 1, "schema_version must be 1")
    require(document.get("kind") == "copilot_training_dataset_governance_manifest", "manifest kind mismatch")
    require(document.get("status") == "draft", "first governance manifest should remain draft")

    scope = document.get("scope")
    require(isinstance(scope, dict), "scope must be an object")
    require(set(scope.get("tasks") or []) == EXPECTED_TASKS, "scope tasks drifted from M4 seed task set")
    require_existing_file(str(scope.get("training_sample_contract") or ""), field_name="training_sample_contract")
    require_existing_file(str(scope.get("conversion_entrypoint") or ""), field_name="conversion_entrypoint")

    check_source_conversion_manifests(document)
    check_candidate_record_eligibility(document)
    check_sampling_policy(document)
    check_artifact_policy(document)

    offline_eval = document.get("offline_eval_connection")
    require(isinstance(offline_eval, dict), "offline_eval_connection must be an object")
    require_existing_file(str(offline_eval.get("eval_run_contract") or ""), field_name="eval_run_contract")
    require_existing_file(str(offline_eval.get("planned_eval_fixture") or ""), field_name="planned_eval_fixture")

    exit_conditions = document.get("exit_conditions")
    require(isinstance(exit_conditions, list) and len(exit_conditions) >= 6, "exit_conditions must cover key retirement cases")

    print("copilot training dataset governance checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
