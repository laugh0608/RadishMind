#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
RECORDS_PATH = REPO_ROOT / "training/datasets/radishmind-core-task-scoped-builder-broader-review-records-v0.json"
ENTRY_PATH = REPO_ROOT / "training/experiments/radishmind-core-task-scoped-builder-broader-review-entry-v0.json"
RUNBOOK_PATH = REPO_ROOT / "training/experiments/radishmind-core-task-scoped-builder-broader-review-runbook-v0.json"
TRAINING_MANIFESTS = [
    REPO_ROOT / "scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json",
    REPO_ROOT / "scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-manifest.json",
]

REQUIRED_DIMENSIONS = {
    "machine_schema_and_task_validity",
    "natural_language_placeholder_absence",
    "ghost_candidate_semantics",
    "citation_explanation_quality",
    "factual_sufficiency",
    "fallback_acceptability",
    "risk_confirmation_and_advisory_boundary",
    "training_holdout_leakage",
}
EXPECTED_TASK_COUNTS = {
    "radishflow/suggest_flowsheet_edits": 5,
    "radishflow/suggest_ghost_completion": 5,
    "radish/answer_docs_question": 5,
}
SEGMENT_OUTPUT_DIRS = {
    "full-holdout-9": "tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-full-holdout-timeout300",
    "holdout6-v2-non-overlap": "tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-v2-timeout300",
}
EXPECTED_V2_TRAINING_OVERLAP = {
    "radishflow-suggest-ghost-completion-flash-vapor-outlet-001",
    "radishflow-suggest-ghost-completion-valve-ambiguous-no-tab-001",
    "radish-answer-docs-question-docs-faq-forum-conflict-001",
    "radish-answer-docs-question-evidence-gap-001",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse {path.relative_to(REPO_ROOT)}: {exc}") from exc


def collect_manifest_paths(document: Any) -> set[str]:
    paths: set[str] = set()
    if isinstance(document, dict):
        for key, value in document.items():
            if key == "path" and isinstance(value, str):
                paths.add(value)
            paths.update(collect_manifest_paths(value))
    elif isinstance(document, list):
        for item in document:
            paths.update(collect_manifest_paths(item))
    return paths


def require_tmp_path(path_value: str, *, field_name: str) -> None:
    require(path_value.startswith("tmp/"), f"{field_name} must stay under tmp/: {path_value}")


def expected_sample_sets(entry: dict[str, Any]) -> dict[str, str]:
    sample_surface = entry.get("sample_surface")
    require(isinstance(sample_surface, dict), "entry sample_surface must be an object")
    by_set = sample_surface.get("sample_ids_by_sample_set")
    require(isinstance(by_set, dict), "entry sample_ids_by_sample_set must be an object")
    expected: dict[str, str] = {}
    for sample_set_id, sample_ids in by_set.items():
        require(sample_set_id in SEGMENT_OUTPUT_DIRS, f"unexpected sample set in entry: {sample_set_id}")
        require(isinstance(sample_ids, list), f"sample set must be a list: {sample_set_id}")
        for sample_id in sample_ids:
            require(isinstance(sample_id, str) and sample_id, f"invalid sample id in {sample_set_id}")
            require(sample_id not in expected, f"sample appears in multiple sets: {sample_id}")
            expected[sample_id] = sample_set_id
    require(len(expected) == 15, "broader review must cover 15 unique samples")
    require(sample_surface.get("task_counts") == EXPECTED_TASK_COUNTS, "entry task counts mismatch")
    require(sample_surface.get("overlap_detected") is False, "entry must not report overlap")
    return expected


def main() -> int:
    document = load_json(RECORDS_PATH)
    entry = load_json(ENTRY_PATH)
    runbook = load_json(RUNBOOK_PATH)
    require(isinstance(document, dict), "broader review records must be a JSON object")
    require(document.get("schema_version") == 1, "schema_version must be 1")
    require(
        document.get("kind") == "radishmind_core_task_scoped_builder_broader_review_records",
        "review records kind mismatch",
    )
    require(document.get("phase") == "M4-preparation", "phase mismatch")
    require(document.get("review_batch_id") == "task-scoped-builder-broader-15-sample-review-v0", "batch id mismatch")
    require(document.get("status") == "pending_review", "broader records must remain pending_review")
    require(document.get("reviewer") is None, "pending broader batch must not have a reviewer")
    require(document.get("reviewed_at") is None, "pending broader batch must not have reviewed_at")
    require(document.get("candidate_track") == "--build-task-scoped-response", "candidate_track mismatch")
    require(document.get("does_not_run_models") is True, "records must not run models")
    require(document.get("does_not_generate_jsonl") is True, "records must not generate JSONL")
    require(document.get("does_not_mark_review_pass") is True, "records must not mark review pass")
    require(document.get("does_not_mark_raw_promotion") is True, "records must not promote raw")
    require(document.get("does_not_mark_training_acceptance") is True, "records must not accept training")

    require(document.get("review_entry") == str(ENTRY_PATH.relative_to(REPO_ROOT)), "review_entry mismatch")
    require(document.get("review_runbook") == str(RUNBOOK_PATH.relative_to(REPO_ROOT)), "review_runbook mismatch")
    require(
        runbook.get("post_run_review_record_expectation", {}).get("future_record_path")
        == str(RECORDS_PATH.relative_to(REPO_ROOT)),
        "runbook future record path mismatch",
    )

    expected_sets = expected_sample_sets(entry)
    source_artifacts = document.get("source_artifacts")
    require(isinstance(source_artifacts, dict), "source_artifacts must be an object")
    require(set(source_artifacts) == set(SEGMENT_OUTPUT_DIRS), "source_artifacts segment set mismatch")
    for segment_id, output_dir in SEGMENT_OUTPUT_DIRS.items():
        artifacts = source_artifacts.get(segment_id)
        require(isinstance(artifacts, dict), f"source_artifacts.{segment_id} must be an object")
        require(artifacts.get("candidate_summary") == f"{output_dir}/summary.json", f"{segment_id} summary path mismatch")
        for key, value in artifacts.items():
            require_tmp_path(str(value or ""), field_name=f"source_artifacts.{segment_id}.{key}")

    pending_gate = document.get("pending_gate_summary")
    require(isinstance(pending_gate, dict), "pending_gate_summary must be an object")
    for key in ("machine_gate_status", "natural_language_audit_status"):
        require(str(pending_gate.get(key) or "").startswith("pending_"), f"{key} must remain pending")
    overlap_policy = str(pending_gate.get("training_manifest_overlap_policy") or "")
    require("full-holdout-9 must remain excluded" in overlap_policy, "overlap policy must protect full-holdout")
    require("holdout6-v2-non-overlap is retained as a regression surface" in overlap_policy, "overlap policy must scope v2")
    require("must not mark training acceptance" in overlap_policy, "overlap policy must reject training acceptance")
    completion_rule = str(pending_gate.get("completion_rule") or "")
    require("reviewed_pass" in completion_rule and "both execution segments" in completion_rule, "completion rule mismatch")

    records = document.get("records")
    require(isinstance(records, list) and len(records) == 15, "records must contain 15 pending records")
    by_sample = {
        str(record.get("sample_id") or ""): record
        for record in records
        if isinstance(record, dict)
    }
    require(set(by_sample) == set(expected_sets), "broader review sample ids mismatch")

    task_counts = {task: 0 for task in EXPECTED_TASK_COUNTS}
    pending_count = 0
    source_eval_paths_by_set: dict[str, set[str]] = {segment_id: set() for segment_id in SEGMENT_OUTPUT_DIRS}
    sample_ids_by_source_path: dict[str, str] = {}
    for sample_id, record in by_sample.items():
        sample_set_id = str(record.get("sample_set_id") or "")
        require(sample_set_id == expected_sets[sample_id], f"sample_set_id mismatch: {sample_id}")
        task = f"{record.get('project')}/{record.get('task')}"
        require(task in task_counts, f"unexpected task for {sample_id}: {task}")
        task_counts[task] += 1
        require(record.get("reviewer") is None, f"pending record must not have reviewer: {sample_id}")
        require(record.get("reviewed_at") is None, f"pending record must not have reviewed_at: {sample_id}")
        require(record.get("review_status") == "pending_review", f"record must remain pending_review: {sample_id}")
        pending_count += 1

        source_eval_sample = str(record.get("source_eval_sample") or "")
        require((REPO_ROOT / source_eval_sample).is_file(), f"missing source eval sample: {sample_id}")
        source_eval_paths_by_set[sample_set_id].add(source_eval_sample)
        sample_ids_by_source_path[source_eval_sample] = sample_id
        output_dir = SEGMENT_OUTPUT_DIRS[sample_set_id]
        require(
            record.get("candidate_response_file") == f"{output_dir}/responses/{sample_id}.candidate-response.json",
            f"candidate response path mismatch: {sample_id}",
        )
        require_tmp_path(str(record.get("candidate_response_file") or ""), field_name=f"{sample_id}.candidate_response_file")
        require_tmp_path(str(record.get("offline_eval_run") or ""), field_name=f"{sample_id}.offline_eval_run")
        require(record.get("natural_language_audit_case_id") == sample_id, f"audit case mismatch: {sample_id}")

        dimensions = record.get("dimension_results")
        require(isinstance(dimensions, dict), f"dimension_results must be an object: {sample_id}")
        require(set(dimensions) == REQUIRED_DIMENSIONS, f"dimension set mismatch: {sample_id}")
        require("reviewed_pass" not in set(dimensions.values()), f"pending record must not include reviewed_pass: {sample_id}")
        require(dimensions.get("machine_schema_and_task_validity") == "pending", f"machine gate must be pending: {sample_id}")
        require(dimensions.get("training_holdout_leakage") == "pending", f"holdout leakage must be pending: {sample_id}")
        if task == "radishflow/suggest_ghost_completion":
            require(dimensions.get("ghost_candidate_semantics") == "pending", f"ghost semantics must be pending: {sample_id}")
            require(dimensions.get("citation_explanation_quality") == "not_applicable", f"ghost citation review must be n/a: {sample_id}")
        else:
            require(dimensions.get("ghost_candidate_semantics") == "not_applicable", f"non-ghost semantics must be n/a: {sample_id}")
            require(dimensions.get("citation_explanation_quality") == "pending", f"citation review must be pending: {sample_id}")

    require(task_counts == EXPECTED_TASK_COUNTS, "broader review task counts mismatch")
    training_paths: set[str] = set()
    for manifest_path in TRAINING_MANIFESTS:
        training_paths.update(collect_manifest_paths(load_json(manifest_path)))
    full_holdout_overlap = source_eval_paths_by_set["full-holdout-9"].intersection(training_paths)
    require(not full_holdout_overlap, "full-holdout broader review samples overlap training conversion manifests")
    v2_overlap_samples = {
        sample_ids_by_source_path[path]
        for path in source_eval_paths_by_set["holdout6-v2-non-overlap"].intersection(training_paths)
    }
    require(v2_overlap_samples == EXPECTED_V2_TRAINING_OVERLAP, "v2 regression overlap sample set mismatch")

    batch = document.get("batch_summary")
    require(isinstance(batch, dict), "batch_summary must be an object")
    require(batch.get("record_count") == 15, "batch record count mismatch")
    require(batch.get("pending_review_count") == pending_count, "batch pending count mismatch")
    require(batch.get("reviewed_pass_count") == 0, "batch pass count must remain zero")
    require(batch.get("reviewed_changes_required_count") == 0, "batch changes-required count must remain zero")
    decision = str(batch.get("decision") or "")
    require("pending review surface only" in decision, "batch decision must state pending-only scope")
    require("raw promotion" in decision and "training acceptance" in decision, "batch decision must reject promotion/acceptance")

    print("radishmind core task-scoped builder broader review records check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
