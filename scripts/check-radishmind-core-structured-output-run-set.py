#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
EXPERIMENT_PATH = REPO_ROOT / "training/experiments/radishmind-core-structured-output-decision-experiment-v0.json"

EXPECTED_SAMPLE_SET = "holdout6-v2-non-overlap"
EXPECTED_FULL_HOLDOUT_SAMPLE_SET = "full-holdout-9"
EXPECTED_MODEL = "Qwen2.5-1.5B-Instruct"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Check the committed M4 structured-output run-set decision summary.",
    )
    parser.add_argument(
        "--summary-output",
        default="",
        help="Optional path to write the normalized structured-output run-set summary.",
    )
    parser.add_argument(
        "--check-summary",
        default="",
        help="Optional expected summary path. Fails if regenerated summary differs.",
    )
    return parser.parse_args()


def repo_path(path_value: str) -> Path:
    path = Path(path_value)
    return path if path.is_absolute() else REPO_ROOT / path


def repo_rel(path: Path) -> str:
    try:
        return path.relative_to(REPO_ROOT).as_posix()
    except ValueError:
        return path.as_posix()


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{repo_rel(path)}': {exc}") from exc


def write_json(path: Path, document: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_dict(document: dict[str, Any], key: str) -> dict[str, Any]:
    value = document.get(key)
    require(isinstance(value, dict), f"{key} must be an object")
    return value


def require_list(document: dict[str, Any], key: str) -> list[Any]:
    value = document.get(key)
    require(isinstance(value, list), f"{key} must be a list")
    return value


def require_tmp_artifacts(track_id: str, artifacts: dict[str, Any]) -> None:
    require(artifacts, f"{track_id} must record local artifacts")
    for artifact_key, artifact_path in artifacts.items():
        path_value = str(artifact_path or "").strip()
        require(path_value.startswith("tmp/"), f"{track_id}.{artifact_key} must stay under tmp/: {path_value}")


def observed_metrics(
    candidate_summary: dict[str, Any],
    *,
    include_task_counts: bool = False,
) -> dict[str, Any]:
    metrics: dict[str, Any] = {
        "schema_valid_rate": candidate_summary.get("schema_valid_rate"),
        "task_valid_rate": candidate_summary.get("task_valid_rate"),
        "task_validation_attempted": candidate_summary.get("task_validation_attempted"),
        "timeout_count": candidate_summary.get("timeout_count"),
        "hit_max_new_tokens_count": candidate_summary.get("hit_max_new_tokens_count"),
    }
    if "builder_output_count" in candidate_summary:
        metrics["builder_output_count"] = candidate_summary.get("builder_output_count")
    if "repaired_output_count" in candidate_summary:
        metrics["repaired_output_count"] = candidate_summary.get("repaired_output_count")
    if "injected_output_count" in candidate_summary:
        metrics["injected_output_count"] = candidate_summary.get("injected_output_count")
    generation_summary = candidate_summary.get("generation_summary")
    if isinstance(generation_summary, dict):
        metrics["generation"] = {
            "samples_measured": generation_summary.get("samples_measured"),
            "total_generation_seconds": generation_summary.get("total_generation_seconds"),
            "avg_generation_seconds": generation_summary.get("avg_generation_seconds"),
            "total_input_tokens": generation_summary.get("total_input_tokens"),
            "total_output_tokens": generation_summary.get("total_output_tokens"),
            "avg_output_tokens": generation_summary.get("avg_output_tokens"),
            "json_extracted_count": generation_summary.get("json_extracted_count"),
            "hit_max_new_tokens_count": generation_summary.get("hit_max_new_tokens_count"),
            "timeout_count": generation_summary.get("timeout_count"),
        }
    else:
        for key in (
            "total_generation_seconds",
            "avg_generation_seconds",
            "total_input_tokens",
            "total_output_tokens",
            "avg_output_tokens",
            "json_extracted_count",
        ):
            if key in candidate_summary:
                metrics[key] = candidate_summary.get(key)
    if include_task_counts and "task_valid_counts_by_task" in candidate_summary:
        metrics["task_valid_counts_by_task"] = candidate_summary.get("task_valid_counts_by_task")
    return {key: value for key, value in metrics.items() if value is not None}


def offline_eval_status(source: dict[str, Any]) -> str:
    offline_summary = source.get("offline_eval_summary") if isinstance(source.get("offline_eval_summary"), dict) else {}
    return str(offline_summary.get("promotion_status") or source.get("promotion_status") or "").strip()


def check_experiment_header(experiment: dict[str, Any]) -> None:
    require(experiment.get("schema_version") == 1, "experiment schema_version must be 1")
    require(experiment.get("kind") == "radishmind_core_decision_experiment", "experiment kind mismatch")
    require(experiment.get("phase") == "M4-preparation", "experiment phase mismatch")
    require(experiment.get("status") == "planned", "experiment status mismatch")
    execution_policy = require_dict(experiment, "execution_policy")
    require(
        execution_policy.get("generated_outputs_must_remain_in_tmp") is True,
        "candidate outputs must stay under tmp",
    )
    require(
        execution_policy.get("requires_user_run_for_local_models") is True,
        "local model runs must remain user-run",
    )
    require(execution_policy.get("allow_invalid_output") is True, "experiment must preserve invalid outputs")
    require(execution_policy.get("validate_task") is True, "experiment must validate task outputs")


def build_raw_repaired_tracks(experiment: dict[str, Any]) -> list[dict[str, Any]]:
    observation = require_dict(experiment, "first_run_observation")
    require(observation.get("status") == "completed", "first run observation must be completed")
    require(observation.get("model") == EXPECTED_MODEL, "first run model mismatch")
    require(observation.get("sample_set") == EXPECTED_SAMPLE_SET, "first run sample set mismatch")
    execution_policy = require_dict(observation, "execution_policy")
    require(execution_policy.get("committed_outputs") is False, "first run candidate outputs must not be committed")
    require(execution_policy.get("sample_timeout_seconds") == 300, "first run timeout must stay 300 seconds")
    require(execution_policy.get("raw_and_repaired_share_timeout") is True, "raw/repaired tracks must share timeout")
    require_tmp_artifacts("first_run_observation", require_dict(observation, "local_artifacts"))

    raw = require_dict(observation, "raw")
    repaired = require_dict(observation, "repaired")
    raw_generation = require_dict(raw, "generation_summary")
    repaired_generation = require_dict(repaired, "generation_summary")
    require(raw.get("promotion_status") == "blocked", "raw track must remain blocked")
    require(float(raw.get("schema_valid_rate")) < 1.0, "raw schema_valid_rate must show failure")
    require(float(raw.get("task_valid_rate")) < 1.0, "raw task_valid_rate must show failure")
    require(raw_generation.get("timeout_count") == 0, "raw initial run should not be timeout-driven")
    require(repaired.get("promotion_status") == "no_promotion_planned", "repaired track must not promote")
    require(repaired.get("schema_valid_rate") == 1.0, "repaired schema_valid_rate mismatch")
    require(repaired.get("task_valid_rate") == 1.0, "repaired task_valid_rate mismatch")
    require(repaired_generation.get("timeout_count") == 0, "repaired initial run should not time out")

    signal = require_dict(observation, "route_decision_signal")
    require(
        signal.get("status") == "favor_structured_output_tooling_before_model_size_comparison",
        "first run route decision signal mismatch",
    )
    return [
        {
            "track_id": "raw_v2",
            "track_kind": "raw",
            "model": observation.get("model"),
            "sample_set": observation.get("sample_set"),
            "promotion_status": raw.get("promotion_status"),
            "machine_metrics": observed_metrics(raw),
            "blocking_task_groups": [
                "radishflow/suggest_flowsheet_edits",
                "radishflow/suggest_ghost_completion",
                "radish/answer_docs_question",
            ],
            "route_signal": signal.get("status"),
        },
        {
            "track_id": "repaired_v2",
            "track_kind": "repair_hard_fields",
            "model": observation.get("model"),
            "sample_set": observation.get("sample_set"),
            "promotion_status": repaired.get("promotion_status"),
            "machine_metrics": observed_metrics(repaired),
            "repaired_path_count": len(require_dict(repaired, "repaired_path_counts")),
            "not_raw_capability_evidence": True,
            "route_signal": signal.get("status"),
        },
    ]


def build_injection_track(experiment: dict[str, Any]) -> dict[str, Any]:
    observation = require_dict(experiment, "hard_field_injection_rerun_observation_2026_05_04")
    require(observation.get("status") == "completed_blocked", "injection rerun must remain completed_blocked")
    require(observation.get("model") == EXPECTED_MODEL, "injection rerun model mismatch")
    require(observation.get("sample_set") == EXPECTED_SAMPLE_SET, "injection rerun sample set mismatch")
    execution_policy = require_dict(observation, "execution_policy")
    require(execution_policy.get("committed_outputs") is False, "injection outputs must not be committed")
    require(execution_policy.get("sample_timeout_seconds") == 300, "injection timeout must stay 300 seconds")
    require_tmp_artifacts("hard_field_injection_rerun", require_dict(observation, "local_artifacts"))
    summary = require_dict(observation, "candidate_summary")
    offline_summary = require_dict(observation, "offline_eval_summary")
    require(summary.get("schema_valid_rate") == 0.8333333333333334, "injection schema_valid_rate mismatch")
    require(summary.get("task_valid_rate") == 0.8, "injection task_valid_rate mismatch")
    require(summary.get("timeout_count") == 1, "injection must preserve observed timeout")
    require(summary.get("blocked_task") == "radishflow/suggest_flowsheet_edits", "injection blocked task mismatch")
    require(offline_summary.get("promotion_status") == "blocked", "injection offline eval must remain blocked")
    return {
        "track_id": "hard_field_injection_v2_rerun",
        "track_kind": "inject_hard_fields",
        "model": observation.get("model"),
        "sample_set": observation.get("sample_set"),
        "promotion_status": offline_summary.get("promotion_status"),
        "machine_metrics": observed_metrics(summary),
        "passed_task_groups": offline_summary.get("passed_task_groups") or [],
        "blocked_task_groups": offline_summary.get("blocked_task_groups") or [],
        "route_signal": require_dict(observation, "route_decision_signal").get("status"),
    }


def build_suggest_builder_track(experiment: dict[str, Any]) -> dict[str, Any]:
    observation = require_dict(experiment, "suggest_edits_response_builder_local_v2_2026_05_04")
    require(
        observation.get("status") == "suggest_edits_passed_but_overall_blocked",
        "suggest-edits builder status mismatch",
    )
    require(observation.get("sample_set") == EXPECTED_SAMPLE_SET, "suggest-edits builder sample set mismatch")
    require_tmp_artifacts("suggest_edits_response_builder", require_dict(observation, "local_artifacts"))
    summary = require_dict(observation, "candidate_summary")
    offline_summary = require_dict(observation, "offline_eval_summary")
    require(summary.get("schema_valid_rate") == 1.0, "suggest builder schema_valid_rate mismatch")
    require(summary.get("task_valid_rate") == 0.6666666666666666, "suggest builder task_valid_rate mismatch")
    require(summary.get("builder_output_count") == 2, "suggest builder output count mismatch")
    require(summary.get("timeout_count") == 0, "suggest builder should not time out")
    require(offline_summary.get("promotion_status") == "blocked", "suggest builder overall must remain blocked")
    require(
        "radishflow/suggest_flowsheet_edits" in (offline_summary.get("passed_task_groups") or []),
        "suggest builder must pass suggest_flowsheet_edits",
    )
    return {
        "track_id": "suggest_edits_response_builder_v2",
        "track_kind": "suggest_edits_response_builder",
        "model": EXPECTED_MODEL,
        "sample_set": observation.get("sample_set"),
        "promotion_status": offline_summary.get("promotion_status"),
        "machine_metrics": observed_metrics(summary, include_task_counts=True),
        "passed_task_groups": offline_summary.get("passed_task_groups") or [],
        "blocked_task_groups": offline_summary.get("blocked_task_groups") or [],
        "route_signal": require_dict(observation, "route_decision_signal").get("status"),
    }


def build_task_scoped_tracks(experiment: dict[str, Any]) -> list[dict[str, Any]]:
    initial = require_dict(experiment, "task_scoped_response_builder_local_v2_2026_05_04")
    fixed = require_dict(experiment, "task_scoped_natural_language_guardrail_local_v2_fixed_2026_05_04")
    require(
        initial.get("status") == "machine_metrics_passed_with_human_review_required",
        "task-scoped initial status mismatch",
    )
    require(fixed.get("status") == "machine_and_target_human_review_passed", "fixed guardrail status mismatch")
    for track_id, observation in (
        ("task_scoped_response_builder", initial),
        ("task_scoped_natural_language_guardrail_fixed", fixed),
    ):
        require(observation.get("sample_set") == EXPECTED_SAMPLE_SET, f"{track_id} sample set mismatch")
        require_tmp_artifacts(track_id, require_dict(observation, "local_artifacts"))
        summary = require_dict(observation, "candidate_summary")
        offline_summary = require_dict(observation, "offline_eval_summary")
        require(summary.get("schema_valid_rate") == 1.0, f"{track_id} schema_valid_rate mismatch")
        require(summary.get("task_valid_rate") == 1.0, f"{track_id} task_valid_rate mismatch")
        require(summary.get("builder_output_count") == 6, f"{track_id} builder_output_count mismatch")
        require(summary.get("timeout_count") == 0, f"{track_id} should not time out")
        require(offline_summary.get("promotion_status") == "no_promotion_planned", f"{track_id} promotion status mismatch")

    human_observations = require_list(initial, "human_review_observations")
    require(human_observations, "initial task-scoped track must preserve human review observations")
    target_review = require_dict(fixed, "target_response_review")
    require(target_review.get("docs_source_conflict_placeholder_removed") is True, "fixed guardrail docs review flag mismatch")
    require(
        target_review.get("ghost_vapor_legal_regulation_mistranslation_removed") is True,
        "fixed guardrail ghost review flag mismatch",
    )
    return [
        {
            "track_id": "task_scoped_response_builder_v2",
            "track_kind": "task_scoped_response_builder",
            "model": EXPECTED_MODEL,
            "sample_set": initial.get("sample_set"),
            "promotion_status": require_dict(initial, "offline_eval_summary").get("promotion_status"),
            "machine_metrics": observed_metrics(require_dict(initial, "candidate_summary"), include_task_counts=True),
            "human_review_status": "target_risks_found",
            "human_review_observation_count": len(human_observations),
            "route_signal": require_dict(initial, "route_decision_signal").get("status"),
        },
        {
            "track_id": "task_scoped_builder_guardrail_fixed_v2",
            "track_kind": "task_scoped_response_builder_with_natural_language_guardrail",
            "model": EXPECTED_MODEL,
            "sample_set": fixed.get("sample_set"),
            "promotion_status": require_dict(fixed, "offline_eval_summary").get("promotion_status"),
            "machine_metrics": observed_metrics(require_dict(fixed, "candidate_summary")),
            "human_review_status": "target_risks_passed",
            "target_response_review": {
                "docs_source_conflict_placeholder_removed": True,
                "ghost_vapor_legal_regulation_mistranslation_removed": True,
            },
            "route_signal": require_dict(fixed, "route_decision_signal").get("status"),
        },
    ]


def build_audit_gate(experiment: dict[str, Any]) -> dict[str, Any]:
    audit = require_dict(experiment, "task_scoped_natural_language_audit_gate_2026_05_04")
    require(audit.get("status") == "repository_gate_added", "natural-language audit gate status mismatch")
    require(audit.get("sample_set") == EXPECTED_SAMPLE_SET, "natural-language audit gate sample set mismatch")
    implementation = require_dict(audit, "implementation")
    require(
        implementation.get("script") == "scripts/audit-radishmind-core-task-scoped-natural-language.py",
        "audit gate script mismatch",
    )
    require(
        implementation.get("required_candidate_track") == "--build-task-scoped-response",
        "audit gate candidate track mismatch",
    )
    fixture_summary = require_dict(audit, "fixture_audit_summary")
    require(fixture_summary.get("status") == "pass", "audit fixture status mismatch")
    require(fixture_summary.get("violation_count") == 0, "audit fixture must have no violations")
    require(fixture_summary.get("fallback_natural_field_rate") == 0.0625, "audit fallback rate mismatch")
    return {
        "gate_id": "task_scoped_natural_language_audit_gate",
        "status": audit.get("status"),
        "script": implementation.get("script"),
        "committed_summary": implementation.get("committed_summary"),
        "required_candidate_track": implementation.get("required_candidate_track"),
        "audit_summary": {
            "status": fixture_summary.get("status"),
            "sample_count": fixture_summary.get("sample_count"),
            "violation_count": fixture_summary.get("violation_count"),
            "fallback_natural_field_rate": fixture_summary.get("fallback_natural_field_rate"),
        },
        "route_signal": require_dict(audit, "route_decision_signal").get("status"),
    }


def build_full_holdout_track(experiment: dict[str, Any]) -> dict[str, Any]:
    observation = require_dict(experiment, "task_scoped_builder_full_holdout_9_post_run_review_2026_05_05")
    require(observation.get("status") == "tightened_rerun_reviewed_pass", "full-holdout post-run status mismatch")
    require(observation.get("sample_set") == EXPECTED_FULL_HOLDOUT_SAMPLE_SET, "full-holdout sample set mismatch")
    require(observation.get("candidate_track") == "--build-task-scoped-response", "full-holdout candidate track mismatch")
    require_tmp_artifacts("task_scoped_builder_full_holdout_9", require_dict(observation, "local_artifacts"))

    summary = require_dict(observation, "candidate_summary")
    require(summary.get("status") == "current_tightened_rerun", "full-holdout summary must be current rerun")
    require(summary.get("sample_count") == 9, "full-holdout sample_count mismatch")
    require(summary.get("schema_valid_rate") == 1.0, "full-holdout schema_valid_rate mismatch")
    require(summary.get("task_valid_rate") == 1.0, "full-holdout task_valid_rate mismatch")
    require(summary.get("builder_output_count") == 9, "full-holdout builder_output_count mismatch")
    require(summary.get("timeout_count") == 0, "full-holdout timeout_count mismatch")
    require(summary.get("hit_max_new_tokens_count") == 0, "full-holdout hit_max_new_tokens_count mismatch")
    require(summary.get("total_generation_seconds") == 890.533, "full-holdout generation seconds mismatch")

    offline_summary = require_dict(observation, "offline_eval_summary")
    require(
        offline_summary.get("promotion_status") == "no_promotion_planned",
        "full-holdout promotion status mismatch",
    )
    require(offline_summary.get("requires_human_review") is True, "full-holdout must require human review")
    blocking_metrics = require_dict(offline_summary, "blocking_metrics")
    require(
        all(value == 1.0 for value in blocking_metrics.values()),
        "full-holdout blocking metrics must all pass after tightened rerun",
    )

    audit_summary = require_dict(observation, "natural_language_audit_summary")
    require(audit_summary.get("status") == "pass", "full-holdout audit status mismatch")
    require(audit_summary.get("violation_count") == 0, "full-holdout audit must have zero violations")
    require(audit_summary.get("warning_count") == 3, "full-holdout audit warning_count mismatch")
    require(
        audit_summary.get("fallback_natural_field_rate") == 0.142857,
        "full-holdout fallback_natural_field_rate mismatch",
    )

    manual_spot_check = require_dict(observation, "manual_spot_check")
    require(
        manual_spot_check.get("status") == "reviewed_pass",
        "full-holdout manual spot-check status mismatch",
    )
    checked_samples = require_list(manual_spot_check, "checked_samples")
    require(len(checked_samples) == 9, "full-holdout manual spot-check must record all reviewed responses")
    followups = require_list(manual_spot_check, "review_blockers_or_followups")
    require(followups, "full-holdout manual spot-check must record review followups")
    joined_followups = "\n".join(str(item) for item in followups)
    require("numeric detail blocker is closed" in joined_followups, "full-holdout review followups must close numeric blocker")
    require("indexed diagnostics" in joined_followups, "full-holdout review followups must preserve indexed citation closure")
    require("non-blocking" in joined_followups, "full-holdout review followups must accept short title warning")
    require("task-specific and accepted" in joined_followups, "full-holdout review followups must preserve fallback acceptance")
    require("Holdout leakage review passed" in joined_followups, "full-holdout review followups must preserve leakage decision")

    human_review_records = require_dict(observation, "human_review_records")
    require(
        human_review_records.get("status") == "reviewed_pass",
        "full-holdout human review records must pass",
    )
    require(
        human_review_records.get("record_path")
        == "training/datasets/radishmind-core-task-scoped-builder-full-holdout-review-records-v0.json",
        "full-holdout human review record path mismatch",
    )
    require(human_review_records.get("record_count") == 9, "full-holdout human review record count mismatch")
    require(human_review_records.get("reviewed_pass_count") == 9, "full-holdout reviewed_pass count mismatch")
    require(human_review_records.get("reviewed_changes_required_count") == 0, "full-holdout reviewed_changes_required count mismatch")
    require(
        human_review_records.get("expansion_acceptance")
        == "ready_for_broader_review",
        "full-holdout expansion acceptance decision mismatch",
    )

    tightened_followup = require_dict(observation, "tightened_fixture_followup")
    require(tightened_followup.get("status") == "implemented_and_rerun_passed", "tightened followup status mismatch")
    new_assertions = require_list(tightened_followup, "new_assertions")
    require(len(new_assertions) == 3, "tightened followup must record three new numeric assertions")
    stale_check = require_dict(tightened_followup, "stale_artifact_check")
    require(stale_check.get("promotion_status") == "blocked", "stale artifact check must remain blocked")
    rerun_check = require_dict(tightened_followup, "tightened_rerun_check")
    require(rerun_check.get("promotion_status") == "no_promotion_planned", "tightened rerun must not promote")
    require(rerun_check.get("natural_language_violation_count") == 0, "tightened rerun audit must have no violations")

    citation_followup = require_dict(observation, "citation_fixture_followup")
    require(citation_followup.get("status") == "implemented_and_rerun_passed", "citation fixture followup must report rerun completion")
    citation_assertions = require_list(citation_followup, "new_assertions")
    require(len(citation_assertions) == 10, "citation followup must record indexed id/locator assertions")
    joined_citation_assertions = "\n".join(str(item) for item in citation_assertions)
    require("context:diagnostics[0]" in joined_citation_assertions, "citation followup must include first diagnostic locator")
    require("artifact:flowsheet_document.units[0]" in joined_citation_assertions, "citation followup must include unit artifact locator")
    require("context:latest_snapshot" in joined_citation_assertions, "citation followup must include snapshot locator")
    required_rerun = require_list(citation_followup, "required_local_rerun")
    require(len(required_rerun) == 3, "citation followup must record local rerun, offline eval and audit commands")
    boundary = str(citation_followup.get("acceptance_boundary") or "")
    require("not raw promotion" in boundary, "citation followup must reject raw promotion")
    require("broader task-scoped review track" in boundary, "citation followup must keep broader review track boundary")

    return {
        "track_id": "task_scoped_builder_full_holdout_9",
        "track_kind": "task_scoped_response_builder_full_holdout",
        "model": EXPECTED_MODEL,
        "sample_set": observation.get("sample_set"),
        "promotion_status": offline_summary.get("promotion_status"),
        "machine_metrics": observed_metrics(summary, include_task_counts=True),
        "natural_language_audit": {
            "status": audit_summary.get("status"),
            "violation_count": audit_summary.get("violation_count"),
            "warning_count": audit_summary.get("warning_count"),
            "fallback_natural_field_rate": audit_summary.get("fallback_natural_field_rate"),
        },
        "human_review_status": manual_spot_check.get("status"),
        "review_followup_count": len(followups),
        "human_review_record_status": human_review_records.get("status"),
        "reviewed_pass_count": human_review_records.get("reviewed_pass_count"),
        "reviewed_changes_required_count": human_review_records.get("reviewed_changes_required_count"),
        "tightened_fixture_status": tightened_followup.get("status"),
        "citation_fixture_status": citation_followup.get("status"),
        "stale_artifact_promotion_status": stale_check.get("promotion_status"),
        "tightened_rerun_promotion_status": rerun_check.get("promotion_status"),
        "not_raw_capability_evidence": True,
        "not_training_acceptance_evidence": True,
        "route_signal": require_dict(observation, "route_decision_signal").get("status"),
    }


def check_current_conclusion(experiment: dict[str, Any]) -> dict[str, Any]:
    conclusion = require_dict(experiment, "current_conclusion")
    require(conclusion.get("status") == "task_scoped_builder_full_holdout_citation_fixture_tightened_reviewed_pass", "current conclusion status mismatch")
    next_step = str(conclusion.get("next_step") or "")
    require("不应直接切 `3B/4B`" in next_step, "conclusion must reject direct 3B/4B switch")
    require("raw 模型晋级" in next_step, "conclusion must reject raw promotion")
    require("训练准入证据" in next_step, "conclusion must reject training acceptance")
    require("扩大 task-scoped builder 样本面" in next_step, "conclusion must require broader builder review")
    require("constrained/guided decoding" in next_step, "conclusion must mention constrained decoding")
    require("raw 模型晋级" in next_step, "conclusion must still reject raw promotion")
    return {
        "status": conclusion.get("status"),
        "next_step": conclusion.get("next_step"),
    }


def build_summary() -> dict[str, Any]:
    experiment = load_json(EXPERIMENT_PATH)
    require(isinstance(experiment, dict), "structured output experiment must be a JSON object")
    check_experiment_header(experiment)
    tracks = []
    tracks.extend(build_raw_repaired_tracks(experiment))
    tracks.append(build_injection_track(experiment))
    tracks.append(build_suggest_builder_track(experiment))
    tracks.extend(build_task_scoped_tracks(experiment))
    tracks.append(build_full_holdout_track(experiment))
    audit_gate = build_audit_gate(experiment)
    conclusion = check_current_conclusion(experiment)

    return {
        "schema_version": 1,
        "kind": "radishmind_core_structured_output_run_set_summary",
        "phase": "M4-preparation",
        "source_experiment": repo_rel(EXPERIMENT_PATH),
        "run_set_id": "radishmind-core-v2-structured-output-decision",
        "sample_set": EXPECTED_SAMPLE_SET,
        "model": EXPECTED_MODEL,
        "track_count": len(tracks),
        "tracks": tracks,
        "repository_gates": [
            audit_gate
        ],
        "decision_invariants": {
            "raw_track_remains_blocked": True,
            "postprocess_tracks_are_not_raw_promotion": True,
            "task_scoped_builder_has_machine_pass": True,
            "natural_language_review_has_separate_gate": True,
            "candidate_outputs_remain_uncommitted": True,
            "next_step_is_broader_builder_review_not_model_size_jump": True,
            "full_holdout_builder_is_not_raw_promotion": True,
            "full_holdout_previous_artifact_is_stale": True,
            "full_holdout_tightened_rerun_has_machine_pass": True,
            "full_holdout_human_review_records_added": True,
            "full_holdout_compressor_citation_fixture_tightened": True,
            "full_holdout_citation_local_rerun_completed": True,
            "full_holdout_expansion_requires_broader_review": True,
        },
        "current_conclusion": conclusion,
    }


def assert_summary(summary: dict[str, Any]) -> None:
    tracks = summary.get("tracks") if isinstance(summary.get("tracks"), list) else []
    by_id = {
        str(track.get("track_id") or ""): track
        for track in tracks
        if isinstance(track, dict)
    }
    required_track_ids = {
        "raw_v2",
        "repaired_v2",
        "hard_field_injection_v2_rerun",
        "suggest_edits_response_builder_v2",
        "task_scoped_response_builder_v2",
        "task_scoped_builder_guardrail_fixed_v2",
        "task_scoped_builder_full_holdout_9",
    }
    missing = sorted(required_track_ids - set(by_id))
    if missing:
        raise SystemExit(f"structured output run-set summary missing track: {missing[0]}")
    require(by_id["raw_v2"].get("promotion_status") == "blocked", "raw_v2 must remain blocked")
    require(by_id["repaired_v2"].get("not_raw_capability_evidence") is True, "repaired_v2 must be separated from raw evidence")
    require(by_id["hard_field_injection_v2_rerun"].get("promotion_status") == "blocked", "injection track must remain blocked")
    require(
        by_id["suggest_edits_response_builder_v2"].get("promotion_status") == "blocked",
        "suggest-only builder must remain overall blocked",
    )
    require(
        by_id["task_scoped_response_builder_v2"].get("human_review_status") == "target_risks_found",
        "initial task-scoped builder must preserve human review risk status",
    )
    require(
        by_id["task_scoped_builder_guardrail_fixed_v2"].get("human_review_status") == "target_risks_passed",
        "fixed task-scoped guardrail must preserve target human review pass",
    )
    full_holdout = by_id["task_scoped_builder_full_holdout_9"]
    require(
        full_holdout.get("not_raw_capability_evidence") is True,
        "full-holdout builder must be separated from raw evidence",
    )
    require(full_holdout.get("human_review_status") == "reviewed_pass", "full-holdout human review must pass")
    require(full_holdout.get("reviewed_pass_count") == 9, "full-holdout reviewed_pass_count mismatch")
    require(full_holdout.get("reviewed_changes_required_count") == 0, "full-holdout reviewed_changes_required_count mismatch")
    require(full_holdout.get("promotion_status") == "no_promotion_planned", "full-holdout must not promote")
    require(
        full_holdout.get("tightened_rerun_promotion_status") == "no_promotion_planned",
        "full-holdout tightened rerun must pass without promotion",
    )
    require(
        full_holdout.get("citation_fixture_status") == "implemented_and_rerun_passed",
        "full-holdout citation fixture status mismatch",
    )
    require(
        full_holdout.get("natural_language_audit", {}).get("violation_count") == 0,
        "full-holdout natural-language audit must have no violations",
    )
    gates = summary.get("repository_gates") if isinstance(summary.get("repository_gates"), list) else []
    require(len(gates) == 1, "run-set summary must record one repository gate")
    require(gates[0].get("audit_summary", {}).get("violation_count") == 0, "audit gate must have no violations")
    invariants = require_dict(summary, "decision_invariants")
    require(all(value is True for value in invariants.values()), "all decision invariants must be true")


def main() -> int:
    args = parse_args()
    summary = build_summary()
    assert_summary(summary)

    if args.summary_output:
        write_json(repo_path(args.summary_output), summary)
    if args.check_summary:
        expected_path = repo_path(args.check_summary)
        expected = load_json(expected_path)
        if expected != summary:
            raise SystemExit(f"structured output run-set summary does not match {repo_rel(expected_path)}")

    print("radishmind core structured output run-set check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
