#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
RECORDS_PATH = REPO_ROOT / "training/datasets/radishmind-core-task-scoped-builder-full-holdout-review-records-v0.json"
HOLDOUT_PATH = REPO_ROOT / "training/datasets/copilot-training-holdout-split-v0.json"
TRAINING_MANIFESTS = [
    REPO_ROOT / "scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json",
    REPO_ROOT / "scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-manifest.json",
]

EXPECTED_SAMPLES = {
    "radishflow-suggest-flowsheet-edits-action-citation-ordering-diagnostic-artifact-snapshot-001",
    "radishflow-suggest-flowsheet-edits-compressor-parameter-update-ordering-001",
    "radishflow-suggest-flowsheet-edits-valve-local-fix-vs-global-balance-001",
    "radishflow-suggest-ghost-completion-flash-inlet-001",
    "radishflow-suggest-ghost-completion-heater-stream-name-001",
    "radishflow-suggest-ghost-completion-mixer-standard-outlet-001",
    "radish-answer-docs-question-attachment-mixed-001",
    "radish-answer-docs-question-docs-attachments-faq-001",
    "radish-answer-docs-question-navigation-001",
}

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


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse {path.relative_to(REPO_ROOT)}: {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def collect_holdout_samples(document: dict[str, Any]) -> set[str]:
    sample_ids: set[str] = set()
    for group in document.get("holdout_selection") or []:
        if not isinstance(group, dict):
            continue
        for sample in group.get("selected_samples") or []:
            if isinstance(sample, dict):
                sample_ids.add(str(sample.get("sample_id") or ""))
    return {sample_id for sample_id in sample_ids if sample_id}


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


def main() -> int:
    document = load_json(RECORDS_PATH)
    require(isinstance(document, dict), "human review records must be a JSON object")
    require(document.get("schema_version") == 1, "schema_version must be 1")
    require(
        document.get("kind") == "radishmind_core_task_scoped_builder_human_review_records",
        "review records kind mismatch",
    )
    require(document.get("phase") == "M4-preparation", "phase mismatch")
    require(document.get("review_batch_id") == "task-scoped-builder-full-holdout-9-full-review-v0", "batch id mismatch")
    require(document.get("sample_set_id") == "full-holdout-9", "sample_set_id mismatch")
    require(document.get("candidate_track") == "--build-task-scoped-response", "candidate_track mismatch")
    require(document.get("status") == "reviewed_pass", "batch status must be reviewed_pass")
    require(document.get("does_not_run_models") is True, "review records must not run models")
    require(document.get("does_not_generate_jsonl") is True, "review records must not generate JSONL")
    require(document.get("does_not_mark_raw_promotion") is True, "review records must not promote raw")
    require(document.get("does_not_mark_training_acceptance") is True, "review records must not accept training")

    for key, value in (document.get("source_artifacts") or {}).items():
        path_value = str(value or "")
        require(path_value.startswith("tmp/"), f"source artifact must stay under tmp/: {key}={path_value}")

    machine = document.get("machine_gate_summary")
    require(isinstance(machine, dict), "machine_gate_summary must be an object")
    require(machine.get("schema_valid_rate") == 1.0, "schema_valid_rate mismatch")
    require(machine.get("task_valid_rate") == 1.0, "task_valid_rate mismatch")
    require(machine.get("builder_output_count") == 9, "builder_output_count mismatch")
    require(machine.get("timeout_count") == 0, "timeout_count mismatch")
    require(machine.get("offline_eval_promotion_status") == "no_promotion_planned", "promotion status mismatch")

    audit = document.get("natural_language_audit_summary")
    require(isinstance(audit, dict), "natural_language_audit_summary must be an object")
    require(audit.get("violation_count") == 0, "natural-language violations must be zero")
    require(audit.get("warning_count") == 3, "warning_count must preserve docs QA short-title warnings")
    require(audit.get("fallback_natural_field_count") == 6, "fallback count mismatch")
    require(audit.get("natural_field_count") == 42, "natural field count mismatch")

    holdout = load_json(HOLDOUT_PATH)
    holdout_samples = collect_holdout_samples(holdout)
    require(holdout_samples == EXPECTED_SAMPLES, "review records expected sample set differs from holdout manifest")

    records = document.get("records")
    require(isinstance(records, list) and len(records) == 9, "records must contain all 9 full-holdout samples")
    by_sample = {
        str(record.get("sample_id") or ""): record
        for record in records
        if isinstance(record, dict)
    }
    require(set(by_sample) == EXPECTED_SAMPLES, "review record sample ids mismatch")

    pass_count = 0
    changes_required = []
    warning_samples = []
    fallback_samples = []
    for sample_id, record in by_sample.items():
        status = str(record.get("review_status") or "")
        if status == "reviewed_pass":
            pass_count += 1
        elif status == "reviewed_changes_required":
            changes_required.append(sample_id)
        else:
            raise SystemExit(f"unexpected review_status for {sample_id}: {status}")

        candidate_path = str(record.get("candidate_response_file") or "")
        require(candidate_path.startswith("tmp/"), f"candidate response must stay under tmp/: {sample_id}")
        require(str(record.get("natural_language_audit_case_id") or "") == sample_id, f"audit case mismatch: {sample_id}")

        dimensions = record.get("dimension_results")
        require(isinstance(dimensions, dict), f"dimension_results must be an object: {sample_id}")
        missing = REQUIRED_DIMENSIONS - set(dimensions)
        if missing:
            raise SystemExit(f"{sample_id} missing review dimension: {sorted(missing)[0]}")
        require(dimensions.get("machine_schema_and_task_validity") == "reviewed_pass", f"machine gate not passed: {sample_id}")
        require(dimensions.get("training_holdout_leakage") == "reviewed_pass", f"holdout leakage not passed: {sample_id}")
        require(dimensions.get("risk_confirmation_and_advisory_boundary") == "reviewed_pass", f"risk boundary not passed: {sample_id}")

        warnings = record.get("warnings") if isinstance(record.get("warnings"), list) else []
        if any("very short" in str(warning) for warning in warnings):
            warning_samples.append(sample_id)
        findings = "\n".join(str(item) for item in (record.get("review_findings") or []))
        if "Fallback" in findings or "fallback" in findings:
            fallback_samples.append(sample_id)

    require(pass_count == 9, "exactly 9 records should be reviewed_pass")
    require(not changes_required, "no records should remain reviewed_changes_required")
    compressor = by_sample["radishflow-suggest-flowsheet-edits-compressor-parameter-update-ordering-001"]
    compressor_dimensions = compressor.get("dimension_results") or {}
    require(
        compressor_dimensions.get("citation_explanation_quality") == "reviewed_pass",
        "compressor citation quality must pass",
    )
    require(
        compressor_dimensions.get("fallback_acceptability") == "reviewed_pass",
        "compressor fallback acceptability must pass",
    )
    compressor_findings = "\n".join(str(item) for item in compressor.get("review_findings") or [])
    require("minimum_delta_kpa=90" in compressor_findings, "compressor numeric detail closure missing")
    require("suggested_minimum=8" in compressor_findings, "compressor suggested_minimum closure missing")
    require("indexed to diagnostics" in compressor_findings, "compressor tightened citation fixture note missing")
    require(
        "broad `artifact:flowsheet_document` blocker is closed" in compressor_findings,
        "compressor broad citation closure missing",
    )
    require(
        compressor.get("exit_reason") is None,
        "compressor exit reason must be cleared after citation-tightened rerun passes",
    )

    require(
        set(warning_samples)
        == {
            "radish-answer-docs-question-attachment-mixed-001",
            "radish-answer-docs-question-docs-attachments-faq-001",
            "radish-answer-docs-question-navigation-001",
        },
        "docs QA short-title warning sample set mismatch",
    )
    require(
        "radishflow-suggest-flowsheet-edits-action-citation-ordering-diagnostic-artifact-snapshot-001"
        in set(fallback_samples),
        "fallback acceptability record missing for action citation sample",
    )

    leakage = document.get("holdout_leakage_review")
    require(isinstance(leakage, dict), "holdout_leakage_review must be an object")
    require(leakage.get("status") == "reviewed_pass", "holdout leakage review must pass")
    require(leakage.get("overlap_detected") is False, "holdout leakage must be false")
    training_paths: set[str] = set()
    for manifest_path in TRAINING_MANIFESTS:
        training_paths.update(collect_manifest_paths(load_json(manifest_path)))
    holdout_paths = {
        str(record.get("source_eval_sample") or "")
        for record in records
        if isinstance(record, dict)
    }
    require(not holdout_paths.intersection(training_paths), "holdout sample path overlaps training conversion manifests")

    batch = document.get("batch_summary")
    require(isinstance(batch, dict), "batch_summary must be an object")
    require(batch.get("reviewed_pass_count") == 9, "batch pass count mismatch")
    require(batch.get("reviewed_changes_required_count") == 0, "batch changes-required count mismatch")
    batch_decision = str(batch.get("decision") or "")
    require("broader task-scoped builder review" in batch_decision, "batch decision must allow broader review")
    require("raw promotion" in batch_decision, "batch decision must reject raw promotion")

    print("radishmind core task-scoped builder human review records check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
