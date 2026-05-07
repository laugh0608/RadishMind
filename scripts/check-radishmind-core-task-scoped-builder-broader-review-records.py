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
EXPECTED_SEGMENT_SUMMARIES = {
    "full-holdout-9": {
        "sample_count": 9,
        "builder_output_count": 9,
        "schema_valid_rate": 1.0,
        "task_valid_rate": 1.0,
        "timeout_count": 0,
        "audit_warning_count": 3,
        "audit_natural_field_count": 42,
        "audit_merged_natural_field_count": 36,
        "audit_fallback_natural_field_count": 6,
        "audit_fallback_natural_field_rate": 0.142857,
    },
    "holdout6-v2-non-overlap": {
        "sample_count": 6,
        "builder_output_count": 6,
        "schema_valid_rate": 1.0,
        "task_valid_rate": 1.0,
        "timeout_count": 0,
        "audit_warning_count": 1,
        "audit_natural_field_count": 32,
        "audit_merged_natural_field_count": 30,
        "audit_fallback_natural_field_count": 2,
        "audit_fallback_natural_field_rate": 0.0625,
    },
}
REVIEWED_CHANGES_REQUIRED = {
    "radishflow-suggest-flowsheet-edits-action-citation-ordering-diagnostic-artifact-snapshot-001",
    "radishflow-suggest-flowsheet-edits-compressor-parameter-update-ordering-001",
    "radish-answer-docs-question-attachment-mixed-001",
    "radish-answer-docs-question-docs-attachments-faq-001",
    "radish-answer-docs-question-navigation-001",
    "radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001",
    "radishflow-suggest-flowsheet-edits-efficiency-range-ordering-001",
    "radishflow-suggest-ghost-completion-valve-ambiguous-no-tab-001",
    "radish-answer-docs-question-docs-faq-forum-conflict-001",
    "radish-answer-docs-question-evidence-gap-001",
}
REVIEWED_PASS = {
    "radishflow-suggest-flowsheet-edits-valve-local-fix-vs-global-balance-001",
    "radishflow-suggest-ghost-completion-flash-inlet-001",
    "radishflow-suggest-ghost-completion-heater-stream-name-001",
    "radishflow-suggest-ghost-completion-mixer-standard-outlet-001",
    "radishflow-suggest-ghost-completion-flash-vapor-outlet-001",
}
PENDING_REVIEW_EXPECTED = 0
REVIEWED_PASS_EXPECTED = 5
REVIEWED_CHANGES_REQUIRED_EXPECTED = 10


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


def require_all_result_records_pass(run_document: dict[str, Any], *, segment_id: str) -> None:
    result_records = run_document.get("result_records")
    require(isinstance(result_records, list) and result_records, f"{segment_id} offline eval result_records must be present")
    for record in result_records:
        require(isinstance(record, dict), f"{segment_id} offline eval result record must be an object")
        require(record.get("status") == "completed", f"{segment_id} offline eval record must be completed")
        metric_results = record.get("metric_results")
        require(isinstance(metric_results, list) and metric_results, f"{segment_id} offline eval metric_results must be present")
        for metric in metric_results:
            require(isinstance(metric, dict), f"{segment_id} metric result must be an object")
            require(metric.get("passed") is True, f"{segment_id} offline eval metric must pass")


def require_review_status(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def validate_segment_artifacts(
    segment_id: str,
    artifacts: dict[str, Any],
    execution_gate: dict[str, Any],
) -> None:
    expected = EXPECTED_SEGMENT_SUMMARIES[segment_id]
    summary_document = load_json(REPO_ROOT / str(artifacts.get("candidate_summary")))
    require(isinstance(summary_document, dict), f"{segment_id} candidate summary must be an object")
    require(summary_document.get("sample_count") == expected["sample_count"], f"{segment_id} sample_count mismatch")
    observation = summary_document.get("observation_summary")
    require(isinstance(observation, dict), f"{segment_id} observation_summary must be an object")
    require(observation.get("schema_valid_rate") == expected["schema_valid_rate"], f"{segment_id} schema_valid_rate mismatch")
    require(observation.get("task_valid_rate") == expected["task_valid_rate"], f"{segment_id} task_valid_rate mismatch")
    generation = observation.get("generation_summary")
    require(isinstance(generation, dict), f"{segment_id} generation_summary must be an object")
    require(generation.get("timeout_count") == expected["timeout_count"], f"{segment_id} timeout_count mismatch")
    postprocess = summary_document.get("postprocess_policy")
    require(isinstance(postprocess, dict), f"{segment_id} postprocess_policy must be an object")
    require(postprocess.get("build_task_scoped_response") is True, f"{segment_id} build_task_scoped_response must be true")
    require(postprocess.get("builder_output_count") == expected["builder_output_count"], f"{segment_id} builder_output_count mismatch")

    run_document = load_json(REPO_ROOT / str(artifacts.get("offline_eval_run")))
    require(isinstance(run_document, dict), f"{segment_id} offline eval run must be an object")
    execution = run_document.get("execution")
    require(isinstance(execution, dict), f"{segment_id} execution must be an object")
    require(execution.get("run_status") == "completed", f"{segment_id} offline eval run_status mismatch")
    decision = run_document.get("decision")
    require(isinstance(decision, dict), f"{segment_id} decision must be an object")
    require(decision.get("promotion_status") == "no_promotion_planned", f"{segment_id} promotion_status mismatch")
    require_all_result_records_pass(run_document, segment_id=segment_id)

    audit_document = load_json(REPO_ROOT / str(artifacts.get("natural_language_audit")))
    require(isinstance(audit_document, dict), f"{segment_id} natural-language audit must be an object")
    audit_summary = audit_document.get("summary")
    require(isinstance(audit_summary, dict), f"{segment_id} audit summary must be an object")
    require(audit_summary.get("status") == "pass", f"{segment_id} audit status mismatch")
    require(audit_summary.get("violation_count") == 0, f"{segment_id} audit violation_count mismatch")
    require(audit_summary.get("warning_count") == expected["audit_warning_count"], f"{segment_id} audit warning_count mismatch")
    require(
        audit_summary.get("natural_field_count") == expected["audit_natural_field_count"],
        f"{segment_id} audit natural_field_count mismatch",
    )
    require(
        audit_summary.get("merged_natural_field_count") == expected["audit_merged_natural_field_count"],
        f"{segment_id} audit merged_natural_field_count mismatch",
    )
    require(
        audit_summary.get("fallback_natural_field_count") == expected["audit_fallback_natural_field_count"],
        f"{segment_id} audit fallback_natural_field_count mismatch",
    )
    require(
        audit_summary.get("fallback_natural_field_rate") == expected["audit_fallback_natural_field_rate"],
        f"{segment_id} audit fallback_natural_field_rate mismatch",
    )

    segments = execution_gate.get("segments")
    require(isinstance(segments, dict), "execution_gate_summary.segments must be an object")
    segment_summary = segments.get(segment_id)
    require(isinstance(segment_summary, dict), f"execution_gate_summary.segments.{segment_id} must be an object")
    machine_gate_summary = segment_summary.get("machine_gate_summary")
    require(isinstance(machine_gate_summary, dict), f"{segment_id} machine_gate_summary must be an object")
    require(machine_gate_summary.get("schema_valid_rate") == expected["schema_valid_rate"], f"{segment_id} machine gate schema_valid_rate mismatch")
    require(machine_gate_summary.get("task_valid_rate") == expected["task_valid_rate"], f"{segment_id} machine gate task_valid_rate mismatch")
    require(machine_gate_summary.get("builder_output_count") == expected["builder_output_count"], f"{segment_id} machine gate builder_output_count mismatch")
    require(machine_gate_summary.get("timeout_count") == expected["timeout_count"], f"{segment_id} machine gate timeout_count mismatch")

    offline_eval_summary = segment_summary.get("offline_eval_summary")
    require(isinstance(offline_eval_summary, dict), f"{segment_id} offline_eval_summary must be an object")
    require(offline_eval_summary.get("run_status") == "completed", f"{segment_id} offline eval summary run_status mismatch")
    require(
        offline_eval_summary.get("promotion_status") == "no_promotion_planned",
        f"{segment_id} offline eval summary promotion_status mismatch",
    )
    require(
        offline_eval_summary.get("all_result_records_passed") is True,
        f"{segment_id} offline eval summary must record all_result_records_passed",
    )

    audit_summary_record = segment_summary.get("natural_language_audit_summary")
    require(isinstance(audit_summary_record, dict), f"{segment_id} natural_language_audit_summary must be an object")
    require(audit_summary_record == audit_summary, f"{segment_id} audit summary must match recorded summary")


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
    require(document.get("status") == "reviewed_changes_required", "broader records must be reviewed_changes_required")
    require(document.get("reviewer") == "repository-reviewer", "broader batch reviewer mismatch")
    require(document.get("reviewed_at") == "2026-05-07T00:00:00Z", "broader batch reviewed_at mismatch")
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

    execution_gate = document.get("execution_gate_summary")
    require(isinstance(execution_gate, dict), "execution_gate_summary must be an object")
    require(
        execution_gate.get("machine_gate_status") == "passed",
        "machine_gate_status must record passed",
    )
    require(
        execution_gate.get("natural_language_audit_status") == "passed",
        "natural_language_audit_status must record passed",
    )
    require(execution_gate.get("human_review_status") == "reviewed_changes_required", "human_review_status must record reviewed_changes_required")
    require(execution_gate.get("holdout_leakage_status") == "reviewed_pass", "holdout_leakage_status must record reviewed_pass")
    overlap_policy = str(execution_gate.get("training_manifest_overlap_policy") or "")
    require("full-holdout-9 must remain excluded" in overlap_policy, "overlap policy must protect full-holdout")
    require("holdout6-v2-non-overlap is retained as a regression surface" in overlap_policy, "overlap policy must scope v2")
    require("must not mark training acceptance" in overlap_policy, "overlap policy must reject training acceptance")
    completion_rule = str(execution_gate.get("completion_rule") or "")
    require("reviewed_pass" in completion_rule and "both execution segments" in completion_rule, "completion rule mismatch")
    for segment_id in SEGMENT_OUTPUT_DIRS:
        validate_segment_artifacts(segment_id, source_artifacts[segment_id], execution_gate)

    records = document.get("records")
    require(isinstance(records, list) and len(records) == 15, "records must contain 15 review records")
    by_sample = {
        str(record.get("sample_id") or ""): record
        for record in records
        if isinstance(record, dict)
    }
    require(set(by_sample) == set(expected_sets), "broader review sample ids mismatch")

    task_counts = {task: 0 for task in EXPECTED_TASK_COUNTS}
    pending_count = 0
    reviewed_pass_count = 0
    reviewed_changes_required_count = 0
    blocking_change_required_samples: list[str] = []
    source_eval_paths_by_set: dict[str, set[str]] = {segment_id: set() for segment_id in SEGMENT_OUTPUT_DIRS}
    sample_ids_by_source_path: dict[str, str] = {}
    for sample_id, record in by_sample.items():
        sample_set_id = str(record.get("sample_set_id") or "")
        require(sample_set_id == expected_sets[sample_id], f"sample_set_id mismatch: {sample_id}")
        task = f"{record.get('project')}/{record.get('task')}"
        require(task in task_counts, f"unexpected task for {sample_id}: {task}")
        task_counts[task] += 1
        review_status = str(record.get("review_status") or "")
        if sample_id in REVIEWED_CHANGES_REQUIRED:
            require_review_status(review_status == "reviewed_changes_required", f"record must be reviewed_changes_required: {sample_id}")
            require(record.get("reviewer") == "repository-reviewer", f"reviewed record reviewer mismatch: {sample_id}")
            require(record.get("reviewed_at") == "2026-05-07T00:00:00Z", f"reviewed_at mismatch: {sample_id}")
            reviewed_changes_required_count += 1
            blocking_change_required_samples.append(sample_id)
        elif sample_id in REVIEWED_PASS:
            require_review_status(review_status == "reviewed_pass", f"record must be reviewed_pass: {sample_id}")
            require(record.get("reviewer") == "repository-reviewer", f"reviewed record reviewer mismatch: {sample_id}")
            require(record.get("reviewed_at") == "2026-05-07T00:00:00Z", f"reviewed_at mismatch: {sample_id}")
            reviewed_pass_count += 1
        else:
            raise SystemExit(f"sample missing expected review bucket: {sample_id}")

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
        if sample_id in REVIEWED_CHANGES_REQUIRED:
            require(dimensions.get("machine_schema_and_task_validity") == "reviewed_pass", f"machine gate must pass after review: {sample_id}")
            require(dimensions.get("training_holdout_leakage") == "reviewed_pass", f"holdout leakage must pass after review: {sample_id}")
            require(
                dimensions.get("risk_confirmation_and_advisory_boundary") == "reviewed_pass",
                f"risk boundary must pass after review: {sample_id}",
            )
            review_notes = record.get("review_notes")
            require(isinstance(review_notes, list) and review_notes, f"review_notes must be populated for reviewed record: {sample_id}")
            require(record.get("exit_reason"), f"reviewed_changes_required record must include exit_reason: {sample_id}")
            if sample_id == "radishflow-suggest-flowsheet-edits-efficiency-range-ordering-001":
                require(
                    dimensions.get("citation_explanation_quality") == "reviewed_changes_required",
                    "efficiency range citation quality must require changes",
                )
                require(
                    dimensions.get("fallback_acceptability") == "reviewed_changes_required",
                    "efficiency range fallback acceptability must require changes",
                )
            if task == "radishflow/suggest_ghost_completion":
                require(
                    dimensions.get("ghost_candidate_semantics") == "reviewed_changes_required",
                    f"ghost blocker must mark ghost_candidate_semantics as reviewed_changes_required: {sample_id}",
                )
                require(
                    dimensions.get("natural_language_placeholder_absence") == "reviewed_changes_required",
                    f"ghost blocker must mark natural-language quality as reviewed_changes_required: {sample_id}",
                )
            else:
                require(dimensions.get("ghost_candidate_semantics") == "not_applicable", f"non-ghost semantics must be n/a: {sample_id}")
        else:
            require(record.get("review_notes"), f"reviewed_pass record must include review notes: {sample_id}")
            require(record.get("exit_reason") is None, f"reviewed_pass record must not include exit_reason: {sample_id}")
            require("pending" not in set(dimensions.values()), f"reviewed_pass record must not include pending dimensions: {sample_id}")
            require("reviewed_changes_required" not in set(dimensions.values()), f"reviewed_pass record must not include reviewed_changes_required dimensions: {sample_id}")
            require(dimensions.get("machine_schema_and_task_validity") == "reviewed_pass", f"machine gate must pass after review: {sample_id}")
            require(dimensions.get("training_holdout_leakage") == "reviewed_pass", f"holdout leakage must pass after review: {sample_id}")
            require(
                dimensions.get("risk_confirmation_and_advisory_boundary") == "reviewed_pass",
                f"risk boundary must pass after review: {sample_id}",
            )
            if task == "radishflow/suggest_ghost_completion":
                require(dimensions.get("ghost_candidate_semantics") == "reviewed_pass", f"ghost semantics must be reviewed_pass: {sample_id}")
                require(dimensions.get("citation_explanation_quality") == "not_applicable", f"ghost citation review must be n/a: {sample_id}")
                require(dimensions.get("fallback_acceptability") == "not_applicable", f"ghost fallback must be n/a: {sample_id}")
            else:
                require(dimensions.get("ghost_candidate_semantics") == "not_applicable", f"non-ghost semantics must be n/a: {sample_id}")
                require(dimensions.get("citation_explanation_quality") == "reviewed_pass", f"citation review must be reviewed_pass: {sample_id}")
                require(dimensions.get("factual_sufficiency") == "reviewed_pass", f"factual sufficiency must be reviewed_pass: {sample_id}")

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
    require(batch.get("reviewed_pass_count") == reviewed_pass_count, "batch reviewed_pass_count mismatch")
    require(pending_count == PENDING_REVIEW_EXPECTED, "unexpected pending review count")
    require(reviewed_pass_count == REVIEWED_PASS_EXPECTED, "unexpected reviewed_pass count")
    require(reviewed_changes_required_count == REVIEWED_CHANGES_REQUIRED_EXPECTED, "unexpected reviewed_changes_required count")
    require(batch.get("reviewed_changes_required_count") == reviewed_changes_required_count, "batch changes-required count mismatch")
    require(batch.get("blocking_change_required_samples") == blocking_change_required_samples, "blocking_change_required_samples mismatch")
    decision = str(batch.get("decision") or "")
    require("completed machine gate" in decision, "batch decision must mention completed machine gate evidence")
    require("full human review evidence" in decision, "batch decision must mention completed human review evidence")
    require("reviewed_changes_required" in decision, "batch decision must mention reviewed_changes_required blockers")
    require("reviewed_pass" in decision, "batch decision must mention reviewed_pass subset")
    require("raw promotion" in decision and "training acceptance" in decision, "batch decision must reject promotion/acceptance")

    print("radishmind core task-scoped builder broader review records check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
