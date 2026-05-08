#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
RUNBOOK_PATH = REPO_ROOT / "training/experiments/radishmind-core-constrained-guided-decoding-runbook-v0.json"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse {path.relative_to(REPO_ROOT)}: {exc}") from exc


def require_existing_file(path_value: str, *, field_name: str) -> None:
    require(path_value.strip() == path_value and path_value.strip(), f"{field_name} must be a non-empty clean path")
    require((REPO_ROOT / path_value).is_file(), f"{field_name} points to a missing file: {path_value}")


def require_tmp_path(path_value: str, *, field_name: str) -> None:
    require(path_value.startswith("tmp/"), f"{field_name} must stay under tmp/: {path_value}")


def require_args(args: list[Any], *, field_name: str, required_flags: tuple[str, ...]) -> None:
    require(all(isinstance(item, str) and item for item in args), f"{field_name} command_args must contain non-empty strings")
    for flag in required_flags:
        require(flag in args, f"{field_name} command_args missing {flag}")
    for blocked_flag in (
        "--repair-hard-fields",
        "--inject-hard-fields",
        "--build-suggest-edits-response",
        "--build-task-scoped-response",
    ):
        require(blocked_flag not in args, f"{field_name} command_args must not use {blocked_flag}")


def arg_value(args: list[str], flag: str) -> str:
    require(flag in args, f"missing flag: {flag}")
    index = args.index(flag)
    require(index + 1 < len(args), f"{flag} must have a value")
    return args[index + 1]


def require_dict(document: dict[str, Any], key: str) -> dict[str, Any]:
    value = document.get(key)
    require(isinstance(value, dict), f"{key} must be an object")
    return value


def main() -> int:
    document = load_json(RUNBOOK_PATH)
    require(isinstance(document, dict), "runbook must be a JSON object")
    require(document.get("schema_version") == 1, "runbook schema_version must be 1")
    require(
        document.get("kind") == "radishmind_core_constrained_guided_decoding_runbook",
        "runbook kind mismatch",
    )
    require(
        document.get("runbook_id") == "radishmind-core-constrained-guided-decoding-runbook-v0",
        "runbook_id mismatch",
    )
    require(document.get("status") == "planned", "runbook status must remain planned")
    require(document.get("phase") == "M4-preparation", "runbook phase mismatch")
    require(document.get("does_not_run_models") is True, "runbook must not run models")
    require(document.get("does_not_generate_jsonl") is True, "runbook must not generate JSONL")
    require(document.get("does_not_mark_raw_promotion") is True, "runbook must not mark raw promotion")
    require(document.get("does_not_mark_training_acceptance") is True, "runbook must not mark training acceptance")
    require(
        document.get("implementation_status") == "wrapper_and_contract_landed_runtime_support_pending_verification",
        "implementation_status mismatch",
    )

    require_existing_file(str(document.get("decision_experiment") or ""), field_name="decision_experiment")
    require_existing_file(
        str(document.get("structured_output_run_set_summary") or ""),
        field_name="structured_output_run_set_summary",
    )
    require_existing_file(str(document.get("broader_review_records") or ""), field_name="broader_review_records")

    reasoning = document.get("priority_reasoning")
    require(isinstance(reasoning, list) and len(reasoning) == 3, "priority_reasoning must include three items")
    joined_reasoning = "\n".join(str(item) for item in reasoning)
    require("raw_v2" in joined_reasoning, "priority_reasoning must mention raw_v2")
    require("15/15 `reviewed_pass`" in joined_reasoning, "priority_reasoning must mention broader reviewed_pass evidence")

    sample_surface = require_dict(document, "sample_surface")
    require(sample_surface.get("sample_set_id") == "holdout6-v2-non-overlap", "sample_set_id mismatch")
    require(sample_surface.get("sample_count") == 6, "sample_count mismatch")
    require(sample_surface.get("task_counts") == {
        "radishflow/suggest_flowsheet_edits": 2,
        "radishflow/suggest_ghost_completion": 2,
        "radish/answer_docs_question": 2,
    }, "task_counts mismatch")
    require_existing_file(str(sample_surface.get("candidate_manifest") or ""), field_name="candidate_manifest")
    require_existing_file(str(sample_surface.get("candidate_eval_manifest") or ""), field_name="candidate_eval_manifest")

    execution_policy = require_dict(document, "execution_policy")
    require(execution_policy.get("provider") == "local_transformers", "provider must be local_transformers")
    require(str(execution_policy.get("model_dir") or "").startswith("/"), "model_dir must be an explicit local path")
    require(execution_policy.get("sample_timeout_seconds") == 300, "sample_timeout_seconds must be 300")
    require(execution_policy.get("allow_invalid_output") is True, "allow_invalid_output must stay true")
    require(execution_policy.get("validate_task") is True, "validate_task must stay true")
    require(execution_policy.get("generated_outputs_default_location") == "tmp/", "generated outputs must stay under tmp/")
    require(
        execution_policy.get("must_share_timeout_with_raw_baseline") is True,
        "guided decoding must keep the same timeout as raw baseline",
    )

    implementation_gap = require_dict(document, "implementation_gap")
    require(
        implementation_gap.get("status") == "wrapper_and_contract_ready_waiting_for_local_runtime_support_verification",
        "implementation_gap status mismatch",
    )
    blocked_components = implementation_gap.get("blocked_components")
    require(isinstance(blocked_components, list) and len(blocked_components) == 2, "blocked_components must include two blockers")
    joined_blockers = "\n".join(str(item) for item in blocked_components)
    require("scripts/run-radishmind-core-candidate.py" in joined_blockers, "blocked_components must mention candidate wrapper")
    require("GenerationConfig.guided_decoding" in joined_blockers, "blocked_components must mention runtime guided-decoding hook")
    require("services/runtime/inference_provider.py" in joined_blockers, "blocked_components must mention provider contract")
    prereqs = implementation_gap.get("must_land_before_user_run")
    require(isinstance(prereqs, list) and len(prereqs) == 2, "must_land_before_user_run must include two prerequisites")
    joined_prereqs = "\n".join(str(item) for item in prereqs)
    require("deterministic repository check" in joined_prereqs, "must_land_before_user_run must require a repository check")
    require("runtime-support failure boundary" in joined_prereqs, "must_land_before_user_run must mention runtime-support boundary")

    variant = require_dict(document, "planned_variant")
    require(variant.get("variant_id") == "constrained_or_guided_decoding", "variant_id mismatch")
    require(variant.get("proposed_cli_flag") == "--guided-decoding", "proposed_cli_flag mismatch")
    require(variant.get("proposed_cli_value") == "json_schema", "proposed_cli_value mismatch")
    require(variant.get("candidate_track") == "raw_guided_json_schema", "candidate_track mismatch")
    require(
        variant.get("mutual_exclusion_with") == [
            "--repair-hard-fields",
            "--inject-hard-fields",
            "--build-suggest-edits-response",
            "--build-task-scoped-response",
        ],
        "mutual_exclusion_with mismatch",
    )

    artifact_paths = require_dict(document, "planned_artifact_paths")
    for key in ("candidate_output_dir", "candidate_summary", "offline_eval_run"):
        require_tmp_path(str(artifact_paths.get(key) or ""), field_name=f"planned_artifact_paths.{key}")
    output_dir = str(artifact_paths.get("candidate_output_dir") or "")
    require(
        str(artifact_paths.get("candidate_summary") or "") == f"{output_dir}/summary.json",
        "candidate_summary path mismatch",
    )

    planned_run = require_dict(document, "planned_local_run")
    run_args = planned_run.get("command_args")
    require(isinstance(run_args, list), "planned_local_run.command_args must be a list")
    require_args(
        run_args,
        field_name="planned_local_run",
        required_flags=(
            "--manifest",
            "--provider",
            "--model-dir",
            "--output-dir",
            "--summary-output",
            "--allow-invalid-output",
            "--validate-task",
            "--guided-decoding",
            "--sample-timeout-seconds",
        ),
    )
    require(arg_value(run_args, "--manifest") == sample_surface["candidate_manifest"], "planned_local_run manifest mismatch")
    require(arg_value(run_args, "--provider") == "local_transformers", "planned_local_run provider mismatch")
    require(arg_value(run_args, "--output-dir") == output_dir, "planned_local_run output_dir mismatch")
    require(arg_value(run_args, "--summary-output") == f"{output_dir}/summary.json", "planned_local_run summary_output mismatch")
    require(arg_value(run_args, "--guided-decoding") == "json_schema", "planned_local_run guided-decoding value mismatch")
    require(arg_value(run_args, "--sample-timeout-seconds") == "300", "planned_local_run timeout mismatch")

    planned_eval = require_dict(document, "planned_offline_eval")
    eval_args = planned_eval.get("command_args")
    require(isinstance(eval_args, list), "planned_offline_eval.command_args must be a list")
    require_args(
        eval_args,
        field_name="planned_offline_eval",
        required_flags=("--manifest", "--candidate-summary", "--candidate-output-dir", "--output", "--check-output"),
    )
    require(arg_value(eval_args, "--manifest") == sample_surface["candidate_eval_manifest"], "planned_offline_eval manifest mismatch")
    require(arg_value(eval_args, "--candidate-summary") == f"{output_dir}/summary.json", "planned_offline_eval summary mismatch")
    require(arg_value(eval_args, "--candidate-output-dir") == output_dir, "planned_offline_eval output dir mismatch")
    require_tmp_path(arg_value(eval_args, "--output"), field_name="planned_offline_eval.output")

    branches = document.get("decision_branches")
    require(isinstance(branches, list) and len(branches) == 2, "decision_branches must include two branches")
    branch_ids = {str(branch.get("branch_id") or "") for branch in branches if isinstance(branch, dict)}
    require(branch_ids == {"guided_decoding_helpful", "guided_decoding_insufficient"}, "decision_branches id mismatch")
    helpful = next(branch for branch in branches if isinstance(branch, dict) and branch.get("branch_id") == "guided_decoding_helpful")
    insufficient = next(branch for branch in branches if isinstance(branch, dict) and branch.get("branch_id") == "guided_decoding_insufficient")
    helpful_when = "\n".join(str(item) for item in helpful.get("enter_when") or [])
    insufficient_when = "\n".join(str(item) for item in insufficient.get("enter_when") or [])
    require("0.8333333333333334" in helpful_when, "helpful branch must compare against raw schema baseline")
    require("0.4" in helpful_when, "helpful branch must compare against raw task baseline")
    require("0.8333333333333334" in insufficient_when, "insufficient branch must keep raw schema baseline")
    require("0.4" in insufficient_when, "insufficient branch must keep raw task baseline")
    require("`3B/4B`" in str(insufficient.get("next_step") or ""), "insufficient branch must point to 3B/4B comparison")

    operator_notes = document.get("operator_notes")
    require(isinstance(operator_notes, list) and len(operator_notes) == 3, "operator_notes must include three notes")
    joined_notes = "\n".join(str(item) for item in operator_notes)
    require("Do not ask the user to run" in joined_notes, "operator_notes must preserve no-user-run-before-runtime-support boundary")
    require("guided-decoding hook" in joined_notes, "operator_notes must mention runtime guided-decoding hook")
    require("300s" in joined_notes, "operator_notes must preserve the shared timeout boundary")
    require("not raw promotion" in joined_notes, "operator_notes must reject raw promotion")

    print("radishmind core constrained/guided decoding runbook check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
