#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
REVIEW_PATH = REPO_ROOT / "training/datasets/radishmind-core-model-adaptation-v1-governance-review-v0.json"

EXPECTED_TASK_COUNTS = {
    ("radishflow", "suggest_flowsheet_edits"): {
        "seed_golden_count": 3,
        "teacher_capture_count": 3,
        "holdout_count": 3,
    },
    ("radishflow", "suggest_ghost_completion"): {
        "seed_golden_count": 3,
        "teacher_capture_count": 3,
        "holdout_count": 3,
    },
    ("radish", "answer_docs_question"): {
        "seed_golden_count": 3,
        "teacher_capture_count": 3,
        "holdout_count": 3,
    },
}

REQUIRED_INPUTS = {
    "dataset_governance": "training/datasets/copilot-training-dataset-governance-v0.json",
    "review_record_template": "training/datasets/copilot-training-review-record-v0.json",
    "holdout_split": "training/datasets/copilot-training-holdout-split-v0.json",
    "preflight_runbook": "training/experiments/radishmind-core-model-adaptation-v1-preflight-runbook-v0.json",
    "golden_response_manifest": "scripts/checks/fixtures/copilot-training-sample-conversion-manifest.json",
    "teacher_capture_manifest": "scripts/checks/fixtures/copilot-training-sample-candidate-record-conversion-manifest.json",
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


def check_inputs(document: dict[str, Any]) -> None:
    inputs = document.get("inputs")
    require(isinstance(inputs, dict), "inputs must be an object")
    require(inputs == REQUIRED_INPUTS, "inputs must match the P4 v1 preflight source set")
    for key, path_value in inputs.items():
        require_existing_file(path_value, field_name=f"inputs.{key}")


def check_task_surface(document: dict[str, Any]) -> None:
    surface = document.get("v1_task_surface")
    require(isinstance(surface, dict), "v1_task_surface must be an object")
    tasks = surface.get("tasks")
    require(isinstance(tasks, list) and len(tasks) == 3, "v1_task_surface.tasks must cover exactly three tasks")

    seen: set[tuple[str, str]] = set()
    for task in tasks:
        require(isinstance(task, dict), "task surface entry must be an object")
        key = (str(task.get("project") or ""), str(task.get("task") or ""))
        require(key in EXPECTED_TASK_COUNTS, f"unexpected task surface entry: {key}")
        seen.add(key)
        expected_counts = EXPECTED_TASK_COUNTS[key]
        for field_name, expected in expected_counts.items():
            require(task.get(field_name) == expected, f"{key} {field_name} must be {expected}")
        notes = task.get("coverage_notes")
        require(isinstance(notes, list) and len(notes) >= 2, f"{key} must include coverage notes")
    require(seen == set(EXPECTED_TASK_COUNTS), "v1 task surface set mismatch")

    summary = surface.get("summary")
    require(isinstance(summary, dict), "v1_task_surface.summary must be an object")
    require(summary.get("seed_golden_total") == 9, "seed_golden_total must be 9")
    require(summary.get("teacher_capture_total") == 9, "teacher_capture_total must be 9")
    require(summary.get("holdout_total") == 9, "holdout_total must be 9")
    require(
        summary.get("tasks_have_minimum_three_holdout_samples") is True,
        "each v1 task must retain at least three holdout samples",
    )


def finding_by_id(document: dict[str, Any], finding_id: str) -> dict[str, Any]:
    findings = document.get("governance_findings")
    require(isinstance(findings, list), "governance_findings must be a list")
    for finding in findings:
        if isinstance(finding, dict) and finding.get("finding_id") == finding_id:
            return finding
    raise SystemExit(f"missing governance finding: {finding_id}")


def check_findings(document: dict[str, Any]) -> None:
    expected_statuses = {
        "source_layering_complete_for_v1": "satisfied_for_preflight",
        "holdout_is_declared_outside_seed_manifests": "satisfied_for_preflight",
        "human_sample_review_not_complete": "not_satisfied_for_training_acceptance",
        "raw_student_probe_order_is_ready": "satisfied_for_operator_decision",
        "artifact_policy_blocks_large_or_real_model_outputs": "satisfied_for_preflight",
    }
    for finding_id, expected_status in expected_statuses.items():
        finding = finding_by_id(document, finding_id)
        require(finding.get("status") == expected_status, f"{finding_id} status mismatch")
        evidence = finding.get("evidence")
        require(isinstance(evidence, list) and evidence, f"{finding_id} must include evidence")

    human_review = "\n".join(str(item) for item in finding_by_id(document, "human_sample_review_not_complete")["evidence"])
    require("pending_review" in human_review, "human review finding must preserve pending_review")
    require("reviewed_pass" in human_review, "human review finding must reject reviewed_pass")

    raw_order = "\n".join(str(item) for item in finding_by_id(document, "raw_student_probe_order_is_ready")["evidence"])
    require("raw_student_probe" in raw_order, "raw order finding must mention raw_student_probe")
    require("repair_boundary_probe" in raw_order, "raw order finding must mention repair_boundary_probe")
    require("not raw promotion" in raw_order, "repair boundary must not be promotion evidence")


def check_operator_boundaries(document: dict[str, Any]) -> None:
    required_before_run = document.get("required_before_requesting_operator_run")
    require(isinstance(required_before_run, list) and len(required_before_run) == 4, "required_before_requesting_operator_run mismatch")
    joined = "\n".join(str(item) for item in required_before_run)
    require("local model directory" in joined, "operator run must require an existing local model directory")
    require("CPU/GPU load" in joined, "operator run must mention local load")
    require("raw_student_probe before repair_boundary_probe" in joined, "operator run must preserve raw-first order")
    require("Do not generate training JSONL" in joined, "operator run must block JSONL generation")

    allowed_run = document.get("allowed_next_operator_run")
    require(isinstance(allowed_run, dict), "allowed_next_operator_run must be an object")
    require(allowed_run.get("status") == "allowed_after_model_dir_selection", "operator run status mismatch")
    require(allowed_run.get("track") == "raw_student_probe", "only raw_student_probe may be next")
    require_existing_file(str(allowed_run.get("command_source") or ""), field_name="allowed_next_operator_run.command_source")

    must_include_flags = allowed_run.get("must_include_flags")
    require(isinstance(must_include_flags, list), "must_include_flags must be a list")
    for flag in (
        "--manifest",
        "--provider",
        "--model-dir",
        "--output-dir",
        "--summary-output",
        "--allow-invalid-output",
        "--validate-task",
        "--sample-timeout-seconds",
    ):
        require(flag in must_include_flags, f"allowed next operator run missing {flag}")

    must_not_include_flags = allowed_run.get("must_not_include_flags")
    require(isinstance(must_not_include_flags, list), "must_not_include_flags must be a list")
    for flag in (
        "--repair-hard-fields",
        "--inject-hard-fields",
        "--build-suggest-edits-response",
        "--build-task-scoped-response",
        "--guided-decoding",
    ):
        require(flag in must_not_include_flags, f"allowed next operator run must block {flag}")


def check_blocked_steps(document: dict[str, Any]) -> None:
    blocked = document.get("blocked_next_steps")
    require(isinstance(blocked, list) and len(blocked) >= 6, "blocked_next_steps must include critical stops")
    joined = "\n".join(str(item) for item in blocked)
    for phrase in (
        "training JSONL generation",
        "model weight download",
        "large batch training",
        "committing raw candidate outputs",
        "reviewed_pass",
        "raw student promotion",
    ):
        require(phrase in joined, f"blocked_next_steps must mention {phrase}")


def main() -> int:
    document = load_json(REVIEW_PATH)
    require(isinstance(document, dict), "P4 governance review must be a JSON object")
    require(document.get("schema_version") == 1, "schema_version must be 1")
    require(
        document.get("kind") == "radishmind_core_model_adaptation_v1_governance_review",
        "kind mismatch",
    )
    require(
        document.get("review_id") == "radishmind-core-model-adaptation-v1-governance-review-v0",
        "review_id mismatch",
    )
    require(
        document.get("status") == "preflight_reviewed_waiting_for_human_sample_review",
        "status must keep human sample review pending",
    )
    require(document.get("phase") == "P4-preflight", "phase mismatch")
    for key in (
        "does_not_run_models",
        "does_not_download_model_artifacts",
        "does_not_generate_jsonl",
        "does_not_mark_sample_review_pass",
        "does_not_mark_raw_promotion",
    ):
        require(document.get(key) is True, f"{key} must be true")

    check_inputs(document)
    check_task_surface(document)
    check_findings(document)
    check_operator_boundaries(document)
    check_blocked_steps(document)

    next_action = str(document.get("next_review_action") or "")
    require("local model directory" in next_action, "next_review_action must require local model directory")
    require("raw_student_probe" in next_action, "next_review_action must keep raw_student_probe first")
    require("repaired comparison" in next_action, "next_review_action must defer repaired comparison")

    print("radishmind core model adaptation v1 governance review check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
