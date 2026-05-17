#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

RECORDS_PATH = REPO_ROOT / "training/datasets/radishmind-core-guided-capacity-review-records-v0.json"
EXPERIMENT_PATH = REPO_ROOT / "training/experiments/radishmind-core-structured-output-decision-experiment-v0.json"
SUMMARY_TOKEN = "','answers':["
EXPECTED_BASELINE_SUMMARY_LEAK_COUNT = 0
EXPECTED_CHALLENGER_SUMMARY_LEAK_COUNT = 6


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


def require_tmp_path(path_value: str, *, field_name: str) -> None:
    require(path_value.startswith("tmp/"), f"{field_name} must stay under tmp/: {path_value}")


def response_text(path_value: str) -> str:
    path = REPO_ROOT / path_value
    require(path.is_file(), f"missing response file: {repo_rel(path)}")
    return path.read_text(encoding="utf-8")


def validate_record(record: dict[str, Any], *, require_local_artifacts: bool) -> None:
    require(record.get("schema_version") == 1, "record schema_version must be 1")
    require(record.get("kind") == "radishmind_core_guided_capacity_review_records", "record kind mismatch")
    require(record.get("phase") == "M4-preparation", "record phase mismatch")
    require(record.get("review_batch_id") == "qwen3-4b-guided-vs-qwen25-3b-holdout6-review-v0", "review_batch_id mismatch")
    require(record.get("status") == "reviewed_changes_required", "record status mismatch")
    require(record.get("reviewer") == "repository-reviewer", "reviewer mismatch")
    require(record.get("reviewed_at") == "2026-05-10T00:00:00Z", "reviewed_at mismatch")
    require(record.get("does_not_run_models") is True, "record must not run models")
    require(record.get("does_not_generate_jsonl") is True, "record must not generate JSONL")
    require(record.get("does_not_mark_raw_promotion") is True, "record must not mark raw promotion")
    require(record.get("does_not_mark_training_acceptance") is True, "record must not mark training acceptance")

    scope = record.get("comparison_scope")
    require(isinstance(scope, dict), "comparison_scope must be an object")
    require(scope.get("sample_set_id") == "holdout6-v2-non-overlap", "sample_set_id mismatch")
    require(scope.get("baseline_model") == "Qwen2.5-3B-Instruct", "baseline_model mismatch")
    require(scope.get("challenger_model") == "Qwen3-4B-Instruct-2507", "challenger_model mismatch")
    require(scope.get("candidate_track") == "raw_guided_json_schema", "candidate_track mismatch")
    require(scope.get("sample_count") == 6, "comparison sample_count mismatch")

    source_artifacts = record.get("source_artifacts")
    require(isinstance(source_artifacts, dict), "source_artifacts must be an object")
    for side in ("baseline", "challenger"):
        side_artifacts = source_artifacts.get(side)
        require(isinstance(side_artifacts, dict), f"source_artifacts.{side} must be an object")
        for key, value in side_artifacts.items():
            require_tmp_path(str(value or ""), field_name=f"source_artifacts.{side}.{key}")

    comparison_dimensions = record.get("comparison_dimensions")
    require(isinstance(comparison_dimensions, dict), "comparison_dimensions must be an object")
    require(
        set(comparison_dimensions) == {"summary_leakage", "title_rationale_generalization", "cross_object_semantics", "ghost_docs_noise"},
        "comparison dimension set mismatch",
    )
    for dimension_name, dimension_record in comparison_dimensions.items():
        require(isinstance(dimension_record, dict), f"{dimension_name} must be an object")
        require(dimension_record.get("status") == "reviewed_changes_required", f"{dimension_name} status mismatch")
        findings = dimension_record.get("findings")
        require(isinstance(findings, list) and findings, f"{dimension_name} findings must be populated")

    sample_reviews = record.get("sample_reviews")
    require(isinstance(sample_reviews, list) and len(sample_reviews) == 6, "sample_reviews must contain 6 entries")
    seen_samples: set[str] = set()
    baseline_leak_count = 0
    challenger_leak_count = 0
    for sample_review in sample_reviews:
        require(isinstance(sample_review, dict), "sample_review must be an object")
        sample_id = str(sample_review.get("sample_id") or "")
        require(sample_id, "sample_id must be non-empty")
        require(sample_id not in seen_samples, f"duplicate sample_id: {sample_id}")
        seen_samples.add(sample_id)
        require(sample_review.get("comparison_result") == "reviewed_changes_required", f"comparison result mismatch for {sample_id}")
        baseline_path = str(sample_review.get("baseline_response_file") or "")
        challenger_path = str(sample_review.get("challenger_response_file") or "")
        require_tmp_path(baseline_path, field_name=f"{sample_id}.baseline_response_file")
        require_tmp_path(challenger_path, field_name=f"{sample_id}.challenger_response_file")
        if require_local_artifacts:
            baseline_text = response_text(baseline_path)
            challenger_text = response_text(challenger_path)
            if SUMMARY_TOKEN in baseline_text:
                baseline_leak_count += 1
            if SUMMARY_TOKEN in challenger_text:
                challenger_leak_count += 1

    require(seen_samples == {
        "radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001",
        "radishflow-suggest-flowsheet-edits-efficiency-range-ordering-001",
        "radishflow-suggest-ghost-completion-flash-vapor-outlet-001",
        "radish-answer-docs-question-docs-faq-forum-conflict-001",
        "radish-answer-docs-question-evidence-gap-001",
        "radishflow-suggest-ghost-completion-valve-ambiguous-no-tab-001",
    }, "sample review sample set mismatch")
    if require_local_artifacts:
        require(baseline_leak_count == EXPECTED_BASELINE_SUMMARY_LEAK_COUNT, "baseline summary leak count mismatch")
        require(challenger_leak_count == EXPECTED_CHALLENGER_SUMMARY_LEAK_COUNT, "challenger summary leak count mismatch")

    batch_summary = record.get("batch_summary")
    require(isinstance(batch_summary, dict), "batch_summary must be an object")
    require(batch_summary.get("sample_count") == 6, "batch sample_count mismatch")
    require(batch_summary.get("comparison_pass_sample_count") == 0, "batch pass sample count mismatch")
    require(batch_summary.get("comparison_changes_required_sample_count") == 6, "batch changes-required sample count mismatch")
    require(batch_summary.get("comparison_dimension_pass_count") == 0, "batch pass dimension count mismatch")
    require(batch_summary.get("comparison_dimension_changes_required_count") == 4, "batch changes-required dimension count mismatch")
    decision = str(batch_summary.get("decision") or "")
    require("partial cleanup" in decision, "batch decision must preserve partial cleanup conclusion")
    require("summary leakage" in decision, "batch decision must mention summary leakage")


def build_summary(record: dict[str, Any], *, require_local_artifacts: bool) -> dict[str, Any]:
    sample_reviews = record.get("sample_reviews") or []
    baseline_leak_count = EXPECTED_BASELINE_SUMMARY_LEAK_COUNT
    challenger_leak_count = EXPECTED_CHALLENGER_SUMMARY_LEAK_COUNT
    if require_local_artifacts:
        baseline_leak_count = 0
        challenger_leak_count = 0
        for sample_review in sample_reviews:
            if not isinstance(sample_review, dict):
                continue
            if SUMMARY_TOKEN in response_text(str(sample_review.get("baseline_response_file") or "")):
                baseline_leak_count += 1
            if SUMMARY_TOKEN in response_text(str(sample_review.get("challenger_response_file") or "")):
                challenger_leak_count += 1
    batch_summary = record.get("batch_summary") if isinstance(record.get("batch_summary"), dict) else {}
    comparison_scope = record.get("comparison_scope") if isinstance(record.get("comparison_scope"), dict) else {}
    return {
        "schema_version": 1,
        "kind": "radishmind_core_guided_capacity_review_records_summary",
        "phase": "M4-preparation",
        "record_path": repo_rel(RECORDS_PATH),
        "review_batch_id": record.get("review_batch_id"),
        "status": record.get("status"),
        "sample_set_id": comparison_scope.get("sample_set_id"),
        "baseline_model": comparison_scope.get("baseline_model"),
        "challenger_model": comparison_scope.get("challenger_model"),
        "sample_count": len(sample_reviews),
        "comparison_result": record.get("status"),
        "baseline_summary_leak_count": baseline_leak_count,
        "challenger_summary_leak_count": challenger_leak_count,
        "comparison_dimension_changes_required_count": 4,
        "sample_reviewed_changes_required_count": int(batch_summary.get("comparison_changes_required_sample_count") or 0),
        "sample_reviewed_pass_count": int(batch_summary.get("comparison_pass_sample_count") or 0),
        "decision": batch_summary.get("decision"),
        "not_raw_promotion": True,
        "not_training_acceptance": True,
    }


def parse_args() -> Any:
    import argparse

    parser = argparse.ArgumentParser(description="Check the committed 3B/4B guided capacity review records.")
    parser.add_argument("--summary-output", default="", help="Optional path to write the normalized summary.")
    parser.add_argument("--check-summary", default="", help="Optional expected summary path.")
    parser.add_argument(
        "--require-local-artifacts",
        action="store_true",
        help="Also open the referenced tmp/ candidate responses. This is for local post-run audits, not default fast checks.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    record = load_json(RECORDS_PATH)
    require(isinstance(record, dict), "guided capacity review records must be a JSON object")
    validate_record(record, require_local_artifacts=args.require_local_artifacts)

    experiment = load_json(EXPERIMENT_PATH)
    require(isinstance(experiment, dict), "experiment must be a JSON object")
    conclusion = experiment.get("current_conclusion")
    require(isinstance(conclusion, dict), "experiment current_conclusion must be an object")
    guided_capacity_audit = conclusion.get("guided_capacity_audit")
    require(isinstance(guided_capacity_audit, dict), "guided_capacity_audit must be an object")
    require(guided_capacity_audit.get("review_record_path") == repo_rel(RECORDS_PATH), "guided_capacity_audit review_record_path mismatch")
    require(guided_capacity_audit.get("review_status") == "reviewed_changes_required", "guided_capacity_audit review_status mismatch")

    normalized_summary = build_summary(record, require_local_artifacts=args.require_local_artifacts)
    if normalized_summary["decision"] != record.get("batch_summary", {}).get("decision"):
        raise SystemExit("batch summary decision mismatch")

    if args.summary_output:
        path = Path(args.summary_output)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(normalized_summary, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    if args.check_summary:
        expected = load_json(REPO_ROOT / args.check_summary)
        if expected != normalized_summary:
            raise SystemExit("guided capacity review records summary mismatch")

    print("radishmind core guided capacity review records check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
