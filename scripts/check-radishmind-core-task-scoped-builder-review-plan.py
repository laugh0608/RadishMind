#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

REVIEW_PLAN_PATH = REPO_ROOT / "training/datasets/radishmind-core-task-scoped-builder-review-plan-v0.json"


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
    require((REPO_ROOT / path_value).is_file(), f"{field_name} points to a missing file: {path_value}")


def main() -> int:
    document = load_json(REVIEW_PLAN_PATH)
    require(isinstance(document, dict), "review plan must be a JSON object")
    require(document.get("schema_version") == 1, "review plan schema_version must be 1")
    require(document.get("kind") == "radishmind_core_task_scoped_builder_review_plan", "review plan kind mismatch")
    require(document.get("review_plan_id") == "radishmind-core-task-scoped-builder-review-plan-v0", "review_plan_id mismatch")
    require(document.get("status") == "planned", "review plan status must remain planned")
    require(document.get("phase") == "M4-preparation", "review plan phase mismatch")
    require(document.get("does_not_run_models") is True, "review plan must not run models")
    require(document.get("does_not_generate_jsonl") is True, "review plan must not generate JSONL")
    require(document.get("does_not_mark_review_pass") is True, "review plan must not mark review pass")
    require(
        str(document.get("created_for") or "").strip() == "task-scoped-response-builder-sample-expansion",
        "created_for mismatch",
    )
    require_existing_file(str(document.get("source_run_set_summary") or ""), field_name="source_run_set_summary")
    require_existing_file(str(document.get("source_audit_summary") or ""), field_name="source_audit_summary")
    require_existing_file(str(document.get("source_experiment") or ""), field_name="source_experiment")

    review_scope = document.get("review_scope")
    require(isinstance(review_scope, dict), "review_scope must be an object")
    require(review_scope.get("candidate_track") == "--build-task-scoped-response", "candidate_track mismatch")
    require(review_scope.get("current_observed_sample_set") == "holdout6-v2-non-overlap", "current sample set mismatch")
    next_expansion = review_scope.get("next_expansion_sample_sets")
    require(isinstance(next_expansion, list) and len(next_expansion) == 2, "next_expansion_sample_sets must have two entries")
    for entry in next_expansion:
        require(isinstance(entry, dict), "next_expansion entry must be an object")
        require_existing_file(str(entry.get("candidate_manifest") or ""), field_name="candidate_manifest")
        require_existing_file(str(entry.get("candidate_eval_manifest") or ""), field_name="candidate_eval_manifest")
        require_existing_file(str(entry.get("expected_summary") or ""), field_name="expected_summary")
        require(int(entry.get("minimum_sample_count") or 0) >= 2, "minimum_sample_count is too low")
        require(str(entry.get("review_mode") or "").strip(), "review_mode must be non-empty")
        require(str(entry.get("reason") or "").strip(), "reason must be non-empty")
    require(
        set(str(entry.get("sample_set_id") or "") for entry in next_expansion)
        == {"full-holdout-9", "holdout6-v2-non-overlap"},
        "unexpected next_expansion sample_set ids",
    )
    must_remain_separate_from = review_scope.get("must_remain_separate_from")
    require(isinstance(must_remain_separate_from, list), "must_remain_separate_from must be a list")
    for required_item in (
        "raw_model_promotion",
        "training_sample_acceptance",
        "production_contract_acceptance",
        "model_size_upgrade_decision",
    ):
        require(required_item in must_remain_separate_from, f"review_scope missing separation rule: {required_item}")

    review_dimensions = document.get("review_dimensions")
    require(isinstance(review_dimensions, list) and len(review_dimensions) >= 6, "review_dimensions must be populated")
    review_dimension_ids = {str(entry.get("id") or "") for entry in review_dimensions if isinstance(entry, dict)}
    for required_id in (
        "machine_schema_and_task_validity",
        "natural_language_placeholder_absence",
        "citation_explanation_quality",
        "factual_sufficiency",
        "risk_confirmation_and_advisory_boundary",
        "training_holdout_leakage",
    ):
        require(required_id in review_dimension_ids, f"missing review dimension: {required_id}")

    planned_review_batches = document.get("planned_review_batches")
    require(isinstance(planned_review_batches, list) and len(planned_review_batches) == 2, "planned_review_batches must contain two entries")
    for entry in planned_review_batches:
        require(isinstance(entry, dict), "planned review batch must be an object")
        require(str(entry.get("review_status") or "") == "pending_review", "planned review batch must stay pending_review")
        require(isinstance(entry.get("requires_offline_eval_run"), bool) and entry["requires_offline_eval_run"], "planned review batch must require offline eval")
        require(isinstance(entry.get("requires_natural_language_audit"), bool) and entry["requires_natural_language_audit"], "planned review batch must require natural language audit")
        require(isinstance(entry.get("requires_human_reviewer"), bool) and entry["requires_human_reviewer"], "planned review batch must require human reviewer")
        require(str(entry.get("review_reason") or "").strip(), "planned review batch must include a reason")
    require(
        {str(entry.get("review_batch_id") or "") for entry in planned_review_batches}
        == {
            "task-scoped-builder-full-holdout-9-full-review-v0",
            "task-scoped-builder-v2-targeted-regression-review-v0",
        },
        "unexpected planned review batch ids",
    )

    acceptance_policy = document.get("acceptance_policy")
    require(isinstance(acceptance_policy, dict), "acceptance_policy must be an object")
    may_expand = acceptance_policy.get("may_expand_task_scoped_builder_when_all_true")
    must_not_expand = acceptance_policy.get("must_not_expand_when_any_true")
    require(isinstance(may_expand, list) and len(may_expand) >= 8, "acceptance policy expansion gate is incomplete")
    require(isinstance(must_not_expand, list) and len(must_not_expand) >= 6, "acceptance policy blocker list is incomplete")
    for required_item in (
        "run_set_summary_keeps_raw_track_blocked",
        "candidate_outputs_remain_under_tmp",
        "offline_eval_run_has_no_blocking_metric_failures",
        "natural_language_audit_violation_count_is_zero",
        "required_human_review_records_are_reviewed_pass",
        "citation_explanation_quality_reviewed_pass",
        "factual_sufficiency_reviewed_pass",
        "risk_confirmation_and_advisory_boundary_reviewed_pass",
        "training_holdout_leakage_reviewed_pass",
    ):
        require(required_item in may_expand, f"acceptance policy missing expansion gate: {required_item}")
    for required_item in (
        "raw_track_is_marked_as_promoted",
        "builder_or_repaired_track_is_used_as_raw_promotion_evidence",
        "candidate_outputs_are_committed_outside_tmp",
        "natural_language_audit_has_violations",
        "review_batch_marked_pass_without_reviewer_and_timestamp",
        "holdout_leakage_detected",
    ):
        require(required_item in must_not_expand, f"acceptance policy missing blocker: {required_item}")

    artifact_policy = document.get("artifact_policy")
    require(isinstance(artifact_policy, dict), "artifact_policy must be an object")
    require(
        isinstance(artifact_policy.get("committed_allowed"), list)
        and {"review_plan", "review_summary", "review_record_template", "deterministic_audit_summary"}.issubset(
            set(artifact_policy["committed_allowed"])
        ),
        "artifact policy committed_allowed list is incomplete",
    )
    require(isinstance(artifact_policy.get("committed_disallowed"), list), "artifact policy committed_disallowed must be a list")
    for blocked in ("local_model_candidate_responses", "provider_raw_dump", "large_jsonl", "model_weights", "checkpoint", "adapter_binary"):
        require(blocked in artifact_policy["committed_disallowed"], f"artifact policy must disallow {blocked}")
    require(artifact_policy.get("generated_outputs_default_location") == "tmp/", "generated outputs must default to tmp/")

    print("radishmind core task-scoped builder review plan check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
