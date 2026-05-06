#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
RUNBOOK_PATH = REPO_ROOT / "training/experiments/radishmind-core-task-scoped-builder-broader-review-runbook-v0.json"
ENTRY_PATH = REPO_ROOT / "training/experiments/radishmind-core-task-scoped-builder-broader-review-entry-v0.json"

EXPECTED_SEGMENTS = {
    "full-holdout-9": {
        "sample_count": 9,
        "candidate_manifest": "scripts/checks/fixtures/radishmind-core-full-holdout-candidate-manifest.json",
        "candidate_eval_manifest": "scripts/checks/fixtures/radishmind-core-full-holdout-candidate-eval-manifest.json",
        "output_dir": "tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-full-holdout-timeout300",
        "builder_output_count": 9,
    },
    "holdout6-v2-non-overlap": {
        "sample_count": 6,
        "candidate_manifest": "scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-manifest.json",
        "candidate_eval_manifest": "scripts/checks/fixtures/radishmind-core-holdout-probe-v2-candidate-eval-manifest.json",
        "output_dir": "tmp/radishmind-core-broader-review-qwen15b-task-scoped-builder-v2-timeout300",
        "builder_output_count": 6,
    },
}
EXPECTED_TASK_COUNTS = {
    "radishflow/suggest_flowsheet_edits": 5,
    "radishflow/suggest_ghost_completion": 5,
    "radish/answer_docs_question": 5,
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


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse {path.relative_to(REPO_ROOT)}: {exc}") from exc


def require_existing_file(path_value: str, *, field_name: str) -> None:
    require(path_value.strip() == path_value and path_value.strip(), f"{field_name} must be a clean non-empty path")
    require((REPO_ROOT / path_value).is_file(), f"{field_name} points to a missing file: {path_value}")


def require_tmp_path(path_value: str, *, field_name: str) -> None:
    require(path_value.startswith("tmp/"), f"{field_name} must stay under tmp/: {path_value}")


def require_dict(document: dict[str, Any], key: str) -> dict[str, Any]:
    value = document.get(key)
    require(isinstance(value, dict), f"{key} must be an object")
    return value


def require_args(args: list[Any], *, field_name: str, required_flags: tuple[str, ...]) -> None:
    require(all(isinstance(item, str) and item for item in args), f"{field_name}.command_args must contain strings")
    for flag in required_flags:
        require(flag in args, f"{field_name}.command_args missing {flag}")
    for blocked_flag in ("--repair-hard-fields", "--inject-hard-fields", "--build-suggest-edits-response"):
        require(blocked_flag not in args, f"{field_name}.command_args must not use {blocked_flag}")


def arg_value(args: list[str], flag: str) -> str:
    require(flag in args, f"missing flag: {flag}")
    index = args.index(flag)
    require(index + 1 < len(args), f"{flag} must have a value")
    return args[index + 1]


def check_segment(segment: dict[str, Any]) -> None:
    segment_id = str(segment.get("segment_id") or "")
    require(segment_id in EXPECTED_SEGMENTS, f"unexpected segment_id: {segment_id}")
    expected = EXPECTED_SEGMENTS[segment_id]
    require(segment.get("sample_count") == expected["sample_count"], f"{segment_id} sample_count mismatch")
    require(segment.get("candidate_manifest") == expected["candidate_manifest"], f"{segment_id} candidate manifest mismatch")
    require(segment.get("candidate_eval_manifest") == expected["candidate_eval_manifest"], f"{segment_id} eval manifest mismatch")
    require_existing_file(str(segment.get("candidate_manifest") or ""), field_name=f"{segment_id}.candidate_manifest")
    require_existing_file(str(segment.get("candidate_eval_manifest") or ""), field_name=f"{segment_id}.candidate_eval_manifest")

    artifact_paths = require_dict(segment, "artifact_paths")
    output_dir = str(artifact_paths.get("candidate_output_dir") or "")
    require(output_dir == expected["output_dir"], f"{segment_id} candidate_output_dir mismatch")
    for key in ("candidate_output_dir", "candidate_summary", "offline_eval_run", "natural_language_audit"):
        require_tmp_path(str(artifact_paths.get(key) or ""), field_name=f"{segment_id}.artifact_paths.{key}")
    require(
        str(artifact_paths.get("candidate_summary") or "") == f"{output_dir}/summary.json",
        f"{segment_id} candidate_summary path mismatch",
    )

    batch_run = require_dict(segment, "official_batch_run")
    batch_args = batch_run.get("command_args")
    require(isinstance(batch_args, list), f"{segment_id} official_batch_run.command_args must be a list")
    require_args(
        batch_args,
        field_name=f"{segment_id}.official_batch_run",
        required_flags=(
            "--manifest",
            "--provider",
            "--model-dir",
            "--output-dir",
            "--summary-output",
            "--allow-invalid-output",
            "--validate-task",
            "--build-task-scoped-response",
            "--sample-timeout-seconds",
        ),
    )
    require(arg_value(batch_args, "--manifest") == expected["candidate_manifest"], f"{segment_id} batch manifest mismatch")
    require(arg_value(batch_args, "--provider") == "local_transformers", f"{segment_id} provider mismatch")
    require(arg_value(batch_args, "--output-dir") == output_dir, f"{segment_id} output-dir mismatch")
    require(arg_value(batch_args, "--summary-output") == f"{output_dir}/summary.json", f"{segment_id} summary-output mismatch")
    require(arg_value(batch_args, "--sample-timeout-seconds") == "300", f"{segment_id} timeout mismatch")

    offline_eval = require_dict(segment, "offline_eval")
    offline_args = offline_eval.get("command_args")
    require(isinstance(offline_args, list), f"{segment_id} offline_eval.command_args must be a list")
    require_args(
        offline_args,
        field_name=f"{segment_id}.offline_eval",
        required_flags=("--manifest", "--candidate-summary", "--candidate-output-dir", "--output", "--check-output"),
    )
    require(arg_value(offline_args, "--manifest") == expected["candidate_eval_manifest"], f"{segment_id} offline eval manifest mismatch")
    require(arg_value(offline_args, "--candidate-summary") == f"{output_dir}/summary.json", f"{segment_id} offline eval summary mismatch")
    require(arg_value(offline_args, "--candidate-output-dir") == output_dir, f"{segment_id} offline eval output dir mismatch")
    require_tmp_path(arg_value(offline_args, "--output"), field_name=f"{segment_id}.offline_eval.output")

    audit = require_dict(segment, "natural_language_audit")
    audit_args = audit.get("command_args")
    require(isinstance(audit_args, list), f"{segment_id} natural_language_audit.command_args must be a list")
    require_args(
        audit_args,
        field_name=f"{segment_id}.natural_language_audit",
        required_flags=("--manifest", "--candidate-summary", "--candidate-output-dir", "--output"),
    )
    require(arg_value(audit_args, "--manifest") == expected["candidate_eval_manifest"], f"{segment_id} audit manifest mismatch")
    require(arg_value(audit_args, "--candidate-summary") == f"{output_dir}/summary.json", f"{segment_id} audit summary mismatch")
    require(arg_value(audit_args, "--candidate-output-dir") == output_dir, f"{segment_id} audit output dir mismatch")
    require_tmp_path(arg_value(audit_args, "--output"), field_name=f"{segment_id}.audit.output")

    expectations = require_dict(segment, "acceptance_expectations")
    candidate_summary = require_dict(expectations, "candidate_summary")
    require(candidate_summary.get("sample_count") == expected["sample_count"], f"{segment_id} expected sample_count mismatch")
    require(candidate_summary.get("schema_valid_rate") == 1.0, f"{segment_id} schema_valid expectation mismatch")
    require(candidate_summary.get("task_valid_rate") == 1.0, f"{segment_id} task_valid expectation mismatch")
    require(candidate_summary.get("builder_output_count") == expected["builder_output_count"], f"{segment_id} builder count mismatch")
    require(candidate_summary.get("timeout_count") == 0, f"{segment_id} timeout expectation mismatch")
    require(expectations.get("offline_eval_promotion_status") == "no_promotion_planned", f"{segment_id} promotion expectation mismatch")
    require(expectations.get("natural_language_audit_status") == "pass", f"{segment_id} audit status expectation mismatch")
    require(expectations.get("natural_language_violation_count") == 0, f"{segment_id} audit violation expectation mismatch")


def main() -> int:
    document = load_json(RUNBOOK_PATH)
    require(isinstance(document, dict), "runbook must be a JSON object")
    require(document.get("schema_version") == 1, "runbook schema_version must be 1")
    require(
        document.get("kind") == "radishmind_core_task_scoped_builder_broader_review_runbook",
        "runbook kind mismatch",
    )
    require(
        document.get("runbook_id") == "radishmind-core-task-scoped-builder-broader-review-runbook-v0",
        "runbook_id mismatch",
    )
    require(document.get("status") == "planned", "runbook status must remain planned")
    require(document.get("phase") == "M4-preparation", "runbook phase mismatch")
    require(document.get("does_not_run_models") is True, "runbook must not run models")
    require(document.get("does_not_generate_jsonl") is True, "runbook must not generate JSONL")
    require(document.get("does_not_mark_review_pass") is True, "runbook must not mark review pass")
    require(document.get("candidate_track") == "--build-task-scoped-response", "candidate_track mismatch")
    require(document.get("review_entry") == "training/experiments/radishmind-core-task-scoped-builder-broader-review-entry-v0.json", "review entry mismatch")
    require_existing_file(str(document.get("review_entry") or ""), field_name="review_entry")
    require_existing_file(str(document.get("review_plan") or ""), field_name="review_plan")
    require(document.get("review_batch_id") == "task-scoped-builder-broader-15-sample-review-v0", "review batch mismatch")

    entry = load_json(ENTRY_PATH)
    entry_surface = require_dict(entry, "sample_surface")
    sample_surface = require_dict(document, "sample_surface")
    require(sample_surface.get("surface_id") == entry_surface.get("surface_id"), "sample surface id mismatch")
    require(sample_surface.get("sample_count") == 15, "sample surface count mismatch")
    require(sample_surface.get("sample_set_ids") == ["full-holdout-9", "holdout6-v2-non-overlap"], "sample set ids mismatch")
    require(sample_surface.get("task_counts") == EXPECTED_TASK_COUNTS, "task_counts mismatch")

    model = require_dict(document, "model")
    require(model.get("provider") == "local_transformers", "provider must be local_transformers")
    require(str(model.get("model_dir") or "").startswith("/"), "model_dir must be an explicit local path")
    require(model.get("sample_timeout_seconds") == 300, "sample timeout must be 300")
    require(model.get("does_not_download_model_artifacts") is True, "model must not download artifacts")

    segments = document.get("execution_segments")
    require(isinstance(segments, list) and len(segments) == 2, "runbook must include two execution segments")
    seen_segments: set[str] = set()
    for segment in segments:
        require(isinstance(segment, dict), "execution segment must be an object")
        seen_segments.add(str(segment.get("segment_id") or ""))
        check_segment(segment)
    require(seen_segments == set(EXPECTED_SEGMENTS), "execution segment ids mismatch")

    review_expectation = require_dict(document, "post_run_review_record_expectation")
    require(
        review_expectation.get("future_record_path")
        == "training/datasets/radishmind-core-task-scoped-builder-broader-review-records-v0.json",
        "future review record path mismatch",
    )
    require(
        review_expectation.get("kind") == "radishmind_core_task_scoped_builder_broader_review_records",
        "future review record kind mismatch",
    )
    require(review_expectation.get("review_batch_id") == document.get("review_batch_id"), "future review batch id mismatch")
    require(review_expectation.get("status_before_review") == "pending_review", "future review status mismatch")
    require(review_expectation.get("minimum_record_count") == 15, "future review minimum_record_count mismatch")
    require(set(review_expectation.get("must_include_segments") or []) == set(EXPECTED_SEGMENTS), "future review segment list mismatch")
    require(set(review_expectation.get("required_dimensions") or []) == REQUIRED_DIMENSIONS, "future review dimensions mismatch")
    require("reviewed_pass" in str(review_expectation.get("completion_rule") or ""), "completion rule must mention reviewed_pass")

    artifact_policy = require_dict(document, "artifact_policy")
    require(artifact_policy.get("generated_outputs_default_location") == "tmp/", "generated outputs must default to tmp/")
    committed_allowed = artifact_policy.get("committed_allowed")
    require(isinstance(committed_allowed, list) and "runbook" in committed_allowed, "artifact policy must allow runbook")
    committed_disallowed = artifact_policy.get("committed_disallowed")
    require(isinstance(committed_disallowed, list), "committed_disallowed must be a list")
    for blocked in ("candidate responses", "provider raw dump", "large jsonl", "model weights", "checkpoint"):
        require(blocked in committed_disallowed, f"artifact policy must disallow {blocked}")

    operator_notes = document.get("operator_notes")
    require(isinstance(operator_notes, list) and len(operator_notes) >= 3, "operator_notes must include execution notes")
    joined_notes = "\n".join(str(note) for note in operator_notes)
    require("two execution segments" in joined_notes, "operator notes must keep two-segment execution boundary")
    require("tooling-route evidence only" in joined_notes, "operator notes must reject raw promotion")

    print("radishmind core task-scoped builder broader review runbook check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
