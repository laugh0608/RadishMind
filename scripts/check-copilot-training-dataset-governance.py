#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
MANIFEST_PATH = REPO_ROOT / "training/datasets/copilot-training-dataset-governance-v0.json"
REVIEW_RECORD_PATH = REPO_ROOT / "training/datasets/copilot-training-review-record-v0.json"
HOLDOUT_SPLIT_PATH = REPO_ROOT / "training/datasets/copilot-training-holdout-split-v0.json"
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


def repo_rel(path: Path) -> str:
    try:
        return path.relative_to(REPO_ROOT).as_posix()
    except ValueError:
        return path.as_posix()


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


def collect_training_seed_sample_ids(document: dict[str, Any]) -> set[str]:
    seed_sample_ids: set[str] = set()
    for entry in document.get("source_conversion_manifests") or []:
        if not isinstance(entry, dict):
            continue
        manifest_path = REPO_ROOT / str(entry.get("manifest") or "")
        manifest = load_json(manifest_path)
        groups = manifest.get("selection") or manifest.get("record_selection") or []
        require(isinstance(groups, list), f"{manifest_path.relative_to(REPO_ROOT)} selection must be a list")
        for group in groups:
            require(isinstance(group, dict), "training conversion manifest group must be an object")
            for sample in group.get("samples") or []:
                require(isinstance(sample, dict), "training conversion manifest sample must be an object")
                sample_path_value = str(sample.get("path") or "")
                require_existing_file(sample_path_value, field_name="training seed sample")
                sample_document = load_json(REPO_ROOT / sample_path_value)
                require(isinstance(sample_document, dict), f"{sample_path_value} must be an object")
                sample_id = str(sample_document.get("sample_id") or "")
                require(sample_id, f"{sample_path_value} is missing sample_id")
                seed_sample_ids.add(sample_id)
    return seed_sample_ids


def iter_holdout_samples(document: dict[str, Any]) -> list[dict[str, Any]]:
    selection = document.get("holdout_selection")
    require(isinstance(selection, list), "holdout_selection must be a list")
    samples: list[dict[str, Any]] = []
    for group in selection:
        require(isinstance(group, dict), "holdout selection group must be an object")
        project = str(group.get("project") or "")
        task = str(group.get("task") or "")
        require(project in {"radishflow", "radish"}, f"unexpected holdout project: {project}")
        require(task in EXPECTED_TASKS, f"unexpected holdout task: {task}")
        minimum_sample_count = group.get("minimum_sample_count")
        require(isinstance(minimum_sample_count, int) and minimum_sample_count >= 3, f"{project}/{task} holdout minimum too low")
        selected = group.get("selected_samples")
        require(isinstance(selected, list), f"{project}/{task} selected_samples must be a list")
        require(len(selected) >= minimum_sample_count, f"{project}/{task} selected holdout count below minimum")
        for sample in selected:
            require(isinstance(sample, dict), "holdout sample must be an object")
            samples.append({"project": project, "task": task, **sample})
    return samples


def check_holdout_split(governance: dict[str, Any], seed_sample_ids: set[str]) -> None:
    document = load_json(HOLDOUT_SPLIT_PATH)
    require(isinstance(document, dict), "holdout split manifest must be an object")
    require(document.get("schema_version") == 1, "holdout split schema_version must be 1")
    require(document.get("kind") == "copilot_training_holdout_split_manifest", "holdout split kind mismatch")
    require(document.get("status") == "planned", "holdout split must remain planned before model eval")
    require(document.get("does_not_generate_jsonl") is True, "holdout split must not generate JSONL")
    require(document.get("does_not_run_models") is True, "holdout split must not run models")
    require(document.get("governance_manifest") == repo_rel(MANIFEST_PATH), "holdout governance manifest mismatch")

    split_policy = document.get("split_policy")
    require(isinstance(split_policy, dict), "holdout split_policy must be an object")
    governance_holdout = (((governance.get("sampling_policy") or {}).get("larger_candidate_record_pool_review") or {}).get("holdout_policy") or {})
    require(
        split_policy.get("offline_eval_holdout_ratio") == governance_holdout.get("offline_eval_holdout_ratio"),
        "holdout ratio must match governance manifest",
    )
    require(
        split_policy.get("minimum_holdout_records_per_task") == governance_holdout.get("minimum_holdout_records_per_task"),
        "minimum holdout records per task must match governance manifest",
    )

    excluded_manifests = document.get("excluded_training_conversion_manifests")
    require(isinstance(excluded_manifests, list) and len(excluded_manifests) == 2, "holdout must exclude both training conversion manifests")
    for path_value in excluded_manifests:
        require_existing_file(str(path_value), field_name="excluded training conversion manifest")

    holdout_samples = iter_holdout_samples(document)
    require(len(holdout_samples) == 9, "first holdout split must include exactly 9 samples")
    task_counts: dict[str, int] = {}
    seen_sample_ids: set[str] = set()
    for sample in holdout_samples:
        sample_id = str(sample.get("sample_id") or "")
        path_value = str(sample.get("path") or "")
        require(sample_id and path_value, "holdout sample must include sample_id and path")
        require(sample_id not in seen_sample_ids, f"duplicate holdout sample_id: {sample_id}")
        seen_sample_ids.add(sample_id)
        require(sample_id not in seed_sample_ids, f"holdout sample overlaps current training seed: {sample_id}")
        require(sample.get("holdout_role") == "offline_eval_holdout", f"{sample_id} holdout_role mismatch")
        coverage_tags = sample.get("coverage_tags")
        require(isinstance(coverage_tags, list) and coverage_tags, f"{sample_id} must include coverage_tags")
        sample_path = REPO_ROOT / path_value
        require(sample_path.is_file(), f"holdout sample path missing: {path_value}")
        sample_document = load_json(sample_path)
        require(isinstance(sample_document, dict), f"{path_value} must be an object")
        require(sample_document.get("sample_id") == sample_id, f"{path_value} sample_id mismatch")
        require(sample_document.get("project") == sample["project"], f"{path_value} project mismatch")
        require(sample_document.get("task") == sample["task"], f"{path_value} task mismatch")
        task_key = f"{sample['project']}/{sample['task']}"
        task_counts[task_key] = task_counts.get(task_key, 0) + 1

    expected_task_counts = {
        "radish/answer_docs_question": 3,
        "radishflow/suggest_flowsheet_edits": 3,
        "radishflow/suggest_ghost_completion": 3,
    }
    require(task_counts == expected_task_counts, "holdout task_counts mismatch")
    summary = document.get("summary")
    require(isinstance(summary, dict), "holdout summary must be an object")
    require(summary.get("holdout_sample_count") == 9, "holdout summary sample count mismatch")
    require(summary.get("task_counts") == expected_task_counts, "holdout summary task counts mismatch")
    require(summary.get("overlaps_current_training_seed") is False, "holdout summary must report no seed overlap")
    require(summary.get("review_status") == "pending_review", "holdout review status must remain pending")


def check_review_record() -> None:
    document = load_json(REVIEW_RECORD_PATH)
    require(isinstance(document, dict), "review record must be an object")
    require(document.get("schema_version") == 1, "review record schema_version must be 1")
    require(document.get("kind") == "copilot_training_review_record", "review record kind mismatch")
    require(document.get("status") == "planned", "review record must remain planned")
    require(document.get("does_not_generate_jsonl") is True, "review record must not generate JSONL")
    require(document.get("does_not_run_models") is True, "review record must not run models")
    require(document.get("governance_manifest") == repo_rel(MANIFEST_PATH), "review governance manifest mismatch")

    record_template = document.get("record_template")
    require(isinstance(record_template, dict), "review record_template must be an object")
    dimensions = record_template.get("review_dimensions")
    require(isinstance(dimensions, dict), "review dimensions template must be an object")
    for required_dimension in (
        "schema_validity",
        "citation_alignment",
        "requires_confirmation_preservation",
        "offline_eval_holdout_leakage",
    ):
        require(dimensions.get(required_dimension) == "pending", f"review dimension {required_dimension} must be pending")

    allowed_statuses = document.get("allowed_review_statuses")
    require(isinstance(allowed_statuses, list), "allowed_review_statuses must be a list")
    for status in ("pending_review", "reviewed_pass", "reviewed_changes_required", "rejected", "deprecated"):
        require(status in allowed_statuses, f"allowed review statuses missing {status}")

    planned_batches = document.get("planned_review_batches")
    require(isinstance(planned_batches, list) and len(planned_batches) == 3, "review record must include three planned batches")
    batch_by_id = {
        str(batch.get("review_batch_id") or ""): batch
        for batch in planned_batches
        if isinstance(batch, dict)
    }
    for batch_id, expected_count in (
        ("seed-golden-response-full-review-v0", 9),
        ("seed-teacher-capture-full-review-v0", 9),
        ("offline-eval-holdout-leakage-review-v0", 9),
    ):
        batch = batch_by_id.get(batch_id)
        require(isinstance(batch, dict), f"review record missing batch {batch_id}")
        require(batch.get("review_status") == "pending_review", f"{batch_id} must remain pending")
        require(batch.get("manual_review_ratio") == 1.0, f"{batch_id} must require full review")
        require(batch.get("minimum_reviewed_records") == expected_count, f"{batch_id} minimum reviewed records mismatch")
        if "conversion_manifest" in batch:
            require_existing_file(str(batch.get("conversion_manifest") or ""), field_name=f"{batch_id}.conversion_manifest")
            require_existing_file(str(batch.get("conversion_summary") or ""), field_name=f"{batch_id}.conversion_summary")
        if "holdout_manifest" in batch:
            require_existing_file(str(batch.get("holdout_manifest") or ""), field_name=f"{batch_id}.holdout_manifest")

    artifact_policy = document.get("artifact_policy")
    require(isinstance(artifact_policy, dict), "review artifact policy must be an object")
    disallowed = artifact_policy.get("committed_disallowed")
    require(isinstance(disallowed, list), "review committed_disallowed must be a list")
    for blocked in ("large_jsonl", "model_weights", "checkpoint", "provider_raw_dump"):
        require(blocked in disallowed, f"review record must disallow {blocked}")


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

    seed_sample_ids = collect_training_seed_sample_ids(document)
    check_holdout_split(document, seed_sample_ids)
    check_review_record()

    print("copilot training dataset governance checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
