#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-foundation-status-summary.json"
READINESS_SUMMARY = REPO_ROOT / "scripts/checks/fixtures/session-tooling-readiness-summary.json"
PRECONDITIONS = REPO_ROOT / "scripts/checks/fixtures/session-tooling-implementation-preconditions.json"
NEGATIVE_SKELETON = REPO_ROOT / "scripts/checks/fixtures/session-tooling-negative-regression-skeleton.json"
INDEPENDENT_AUDIT_RECORDS_DESIGN = REPO_ROOT / "scripts/checks/fixtures/session-tooling-independent-audit-records-design.json"
RESULT_MATERIALIZATION_POLICY_DESIGN = REPO_ROOT / "scripts/checks/fixtures/session-tooling-result-materialization-policy-design.json"
EXECUTOR_BOUNDARY_DESIGN = REPO_ROOT / "scripts/checks/fixtures/session-tooling-executor-boundary-design.json"
STORAGE_BACKEND_DESIGN = REPO_ROOT / "scripts/checks/fixtures/session-tooling-storage-backend-design.json"
IMPLEMENTATION_GATES = REPO_ROOT / "scripts/checks/fixtures/session-tooling-deny-by-default-implementation-gates.json"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-foundation-status-summary.py"

REQUIRED_TRACK_IDS = {
    "contract_and_fixture_gates",
    "metadata_route_smoke",
    "promotion_and_readiness",
    "implementation_preconditions",
    "independent_audit_records_design",
    "negative_regression_skeleton",
    "result_materialization_policy_design",
    "executor_boundary_design",
    "storage_backend_design",
    "deny_by_default_implementation_gates",
}
REQUIRED_AREAS = {"executor", "storage", "confirmation"}
REQUIRED_NEXT_STAGE_CONDITIONS = {
    "upper_layer_confirmation_flow",
    "executor_boundary",
    "storage_backend_design",
    "result_materialization_policy",
    "independent_audit_records",
    "complete_negative_regression_suite",
}
REQUIRED_NOT_CLAIMED = {
    "automatic_replay",
    "confirmed_action_execution",
    "durable_session_store",
    "durable_tool_store",
    "long_term_memory",
    "materialized_result_reader",
    "real_checkpoint_storage_backend",
    "real_tool_executor",
}
REQUIRED_CLOSE_LIMITS = {
    "does_not_mark_p2_short_close",
    "does_not_enable_executor",
    "does_not_enable_durable_storage",
    "does_not_enable_confirmation_flow",
    "does_not_enable_replay",
    "does_not_complete_negative_regression_suite",
}


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def relative_path(path: Path) -> str:
    return path.relative_to(REPO_ROOT).as_posix()


def require_object(value: Any, message: str) -> dict[str, Any]:
    require(isinstance(value, dict), message)
    return value


def require_string_list(value: Any, message: str) -> list[str]:
    require(isinstance(value, list) and value, message)
    items: list[str] = []
    for item in value:
        normalized = str(item or "").strip()
        require(bool(normalized), message)
        items.append(normalized)
    return items


def not_ready_areas_from_preconditions(preconditions: dict[str, Any]) -> list[dict[str, Any]]:
    areas: list[dict[str, Any]] = []
    for area in preconditions.get("areas") or []:
        area_obj = require_object(area, "precondition area must be object")
        area_id = str(area_obj.get("area") or "").strip()
        require(area_id in REQUIRED_AREAS, f"unexpected precondition area: {area_id}")
        require(area_obj.get("status") == "not_ready", f"{area_id} must remain not_ready")
        areas.append(
            {
                "area": area_id,
                "status": "not_ready",
                "required_before_enablement": require_string_list(
                    area_obj.get("required_before_enablement"),
                    f"{area_id} must include required_before_enablement",
                ),
            }
        )
    require({area["area"] for area in areas} == REQUIRED_AREAS, "preconditions must include executor/storage/confirmation")
    return areas


def check_negative_skeleton_groups(negative_skeleton: dict[str, Any]) -> None:
    groups = negative_skeleton.get("groups")
    require(isinstance(groups, list) and groups, "negative skeleton must include groups")
    group_areas = {
        str(group.get("precondition_area") or "").strip()
        for group in groups
        if isinstance(group, dict)
    }
    require(group_areas == REQUIRED_AREAS, "negative skeleton must cover executor/storage/confirmation")
    for group in groups:
        group_obj = require_object(group, "negative skeleton group must be object")
        require(group_obj.get("status") == "skeleton_only", "negative skeleton groups must stay skeleton_only")


def build_summary() -> dict[str, Any]:
    readiness = require_object(load_json_document(READINESS_SUMMARY), "readiness summary must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS), "preconditions fixture must be object")
    negative_skeleton = require_object(load_json_document(NEGATIVE_SKELETON), "negative skeleton fixture must be object")
    implementation_gates = require_object(load_json_document(IMPLEMENTATION_GATES), "implementation gates fixture must be object")

    require(readiness.get("status") == "metadata_ready_implementation_blocked", "readiness must stay implementation blocked")
    require(preconditions.get("status") == "preconditions_not_satisfied", "preconditions must stay unsatisfied")
    require(
        negative_skeleton.get("status") == "skeleton_only_implementation_blocked",
        "negative skeleton must stay implementation blocked",
    )
    require(
        implementation_gates.get("status") == "deny_by_default_gates_defined_implementation_blocked",
        "implementation gates must be defined but blocked",
    )
    check_negative_skeleton_groups(negative_skeleton)

    return {
        "schema_version": 1,
        "kind": "session_tooling_foundation_status_summary",
        "stage": "P2 Session & Tooling Foundation",
        "status": "close_candidate_governance_only",
        "implementation_status": "not_implemented",
        "source_readiness_summary": relative_path(READINESS_SUMMARY),
        "source_implementation_preconditions": relative_path(PRECONDITIONS),
        "source_negative_regression_skeleton": relative_path(NEGATIVE_SKELETON),
        "source_independent_audit_records_design": relative_path(INDEPENDENT_AUDIT_RECORDS_DESIGN),
        "source_result_materialization_policy_design": relative_path(RESULT_MATERIALIZATION_POLICY_DESIGN),
        "source_executor_boundary_design": relative_path(EXECUTOR_BOUNDARY_DESIGN),
        "source_storage_backend_design": relative_path(STORAGE_BACKEND_DESIGN),
        "source_deny_by_default_implementation_gates": relative_path(IMPLEMENTATION_GATES),
        "completed_governance_tracks": [
            {
                "track_id": "contract_and_fixture_gates",
                "status": "complete",
                "evidence": [
                    "scripts/check-session-record-contract.py",
                    "scripts/check-tooling-framework-contract.py",
                    "scripts/check-session-recovery-checkpoint-contract.py",
                ],
                "claim": "session, tooling, and checkpoint read contracts are fixture-checkable",
            },
            {
                "track_id": "metadata_route_smoke",
                "status": "complete",
                "evidence": [
                    "scripts/checks/fixtures/session-recovery-checkpoint-route-smoke-coverage-summary.json",
                    "scripts/check-session-recovery-route-smoke-coverage.py",
                ],
                "claim": "checkpoint read route smoke covers metadata-only response shape and denied query categories",
            },
            {
                "track_id": "promotion_and_readiness",
                "status": "complete",
                "evidence": [
                    "scripts/checks/fixtures/session-tooling-promotion-gates.json",
                    "scripts/checks/fixtures/session-tooling-readiness-summary.json",
                    "scripts/check-session-tooling-promotion-gates.py",
                    "scripts/check-session-tooling-readiness-summary.py",
                ],
                "claim": "promotion gates and readiness summary separate current governance claims from future implementation claims",
            },
            {
                "track_id": "implementation_preconditions",
                "status": "complete",
                "evidence": [
                    "docs/task-cards/session-tooling-implementation-preconditions.md",
                    "scripts/checks/fixtures/session-tooling-implementation-preconditions.json",
                    "scripts/check-session-tooling-implementation-preconditions.py",
                ],
                "claim": "executor, storage, and confirmation implementation preconditions are explicit and still blocked",
            },
            {
                "track_id": "negative_regression_skeleton",
                "status": "complete",
                "evidence": [
                    "scripts/checks/fixtures/session-tooling-negative-regression-skeleton.json",
                    "scripts/check-session-tooling-negative-regression-skeleton.py",
                ],
                "claim": "blocked executor, storage/materialization, and confirmation cases have skeleton coverage",
            },
            {
                "track_id": "independent_audit_records_design",
                "status": "complete",
                "evidence": [
                    "docs/task-cards/session-tooling-independent-audit-records.md",
                    "scripts/checks/fixtures/session-tooling-independent-audit-records-design.json",
                    "scripts/check-session-tooling-independent-audit-records.py",
                ],
                "claim": "independent audit record shape, event sources, and confirmation/executor/storage boundaries are design-checkable",
            },
            {
                "track_id": "result_materialization_policy_design",
                "status": "complete",
                "evidence": [
                    "docs/task-cards/session-tooling-result-materialization-policy.md",
                    "scripts/checks/fixtures/session-tooling-result-materialization-policy-design.json",
                    "scripts/check-session-tooling-result-materialization-policy.py",
                ],
                "claim": "metadata-only, result_ref, and materialized result boundaries are design-checkable while current result readers remain disabled",
            },
            {
                "track_id": "executor_boundary_design",
                "status": "complete",
                "evidence": [
                    "docs/task-cards/session-tooling-executor-boundary.md",
                    "scripts/checks/fixtures/session-tooling-executor-boundary-design.json",
                    "scripts/check-session-tooling-executor-boundary.py",
                ],
                "claim": "executor sandbox, allowlist, execution envelope, timeout/retry, and failure boundaries are design-checkable while execution remains disabled",
            },
            {
                "track_id": "storage_backend_design",
                "status": "complete",
                "evidence": [
                    "docs/task-cards/session-tooling-storage-backend-design.md",
                    "scripts/checks/fixtures/session-tooling-storage-backend-design.json",
                    "scripts/check-session-tooling-storage-backend-design.py",
                ],
                "claim": "session, checkpoint, audit, and result store responsibilities are design-checkable while durable storage remains disabled",
            },
            {
                "track_id": "deny_by_default_implementation_gates",
                "status": "complete",
                "evidence": [
                    "docs/task-cards/session-tooling-deny-by-default-implementation-gates.md",
                    "scripts/checks/fixtures/session-tooling-deny-by-default-implementation-gates.json",
                    "scripts/check-session-tooling-deny-by-default-implementation-gates.py",
                ],
                "claim": "executor, storage, and confirmation deny-by-default implementation gate contracts are checkable while real implementations remain disabled",
            },
        ],
        "missing_implementation_prerequisites": not_ready_areas_from_preconditions(preconditions),
        "next_stage_entry_conditions": [
            "upper_layer_confirmation_flow",
            "executor_boundary",
            "storage_backend_design",
            "result_materialization_policy",
            "independent_audit_records",
            "complete_negative_regression_suite",
        ],
        "not_claimed_capabilities": sorted(REQUIRED_NOT_CLAIMED),
        "close_candidate_limits": [
            "does_not_mark_p2_short_close",
            "does_not_enable_executor",
            "does_not_enable_durable_storage",
            "does_not_enable_confirmation_flow",
            "does_not_enable_replay",
            "does_not_complete_negative_regression_suite",
        ],
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_status_summary",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_summary_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "status summary schema_version must be 1")
    require(document.get("kind") == "session_tooling_foundation_status_summary", "status summary kind mismatch")
    require(document.get("status") == "close_candidate_governance_only", "status summary must be governance-only")
    require(document.get("implementation_status") == "not_implemented", "status summary must not claim implementation")

    tracks = document.get("completed_governance_tracks")
    require(isinstance(tracks, list) and tracks, "status summary must include completed tracks")
    track_ids = {str(track.get("track_id") or "").strip() for track in tracks if isinstance(track, dict)}
    missing_tracks = sorted(REQUIRED_TRACK_IDS - track_ids)
    require(not missing_tracks, f"status summary missing tracks: {missing_tracks}")

    areas = document.get("missing_implementation_prerequisites")
    require(isinstance(areas, list) and areas, "status summary must include missing prerequisites")
    area_ids = {str(area.get("area") or "").strip() for area in areas if isinstance(area, dict)}
    missing_areas = sorted(REQUIRED_AREAS - area_ids)
    require(not missing_areas, f"status summary missing areas: {missing_areas}")
    for area in areas:
        area_obj = require_object(area, "missing prerequisite must be object")
        require(area_obj.get("status") == "not_ready", "missing prerequisites must remain not_ready")

    next_conditions = set(require_string_list(document.get("next_stage_entry_conditions"), "next conditions required"))
    missing_next = sorted(REQUIRED_NEXT_STAGE_CONDITIONS - next_conditions)
    require(not missing_next, f"status summary missing next stage conditions: {missing_next}")

    not_claimed = set(require_string_list(document.get("not_claimed_capabilities"), "not claimed capabilities required"))
    missing_not_claimed = sorted(REQUIRED_NOT_CLAIMED - not_claimed)
    require(not missing_not_claimed, f"status summary missing not-claimed capabilities: {missing_not_claimed}")

    limits = set(require_string_list(document.get("close_candidate_limits"), "close candidate limits required"))
    missing_limits = sorted(REQUIRED_CLOSE_LIMITS - limits)
    require(not missing_limits, f"status summary missing close candidate limits: {missing_limits}")


def check_docs_and_consumers() -> None:
    current_focus = CURRENT_FOCUS.read_text(encoding="utf-8")
    devlog = DEVLOG.read_text(encoding="utf-8")
    capability_matrix = CAPABILITY_MATRIX.read_text(encoding="utf-8")
    roadmap = ROADMAP.read_text(encoding="utf-8")
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name

    for label, content in (
        ("docs/radishmind-current-focus.md", current_focus),
        ("docs/devlogs/2026-W20.md", devlog),
        ("docs/radishmind-capability-matrix.md", capability_matrix),
        ("docs/radishmind-roadmap.md", roadmap),
    ):
        require(fixture_name in content, f"{label} must reference foundation status summary")
        require("close candidate" in content, f"{label} must mention close candidate")
        require("governance-only" in content, f"{label} must mention governance-only")
    for label, content in (
        ("docs/radishmind-capability-matrix.md", capability_matrix),
        ("docs/radishmind-roadmap.md", roadmap),
    ):
        require("P2 short close" in content, f"{label} must explicitly avoid P2 short close")
        require("negative_regression_suite" in content, f"{label} must mention negative_regression_suite boundary")
    require(
        "check-session-tooling-foundation-status-summary.py" in check_repo,
        "fast baseline must run foundation status summary check",
    )
    require(
        "check-session-tooling-independent-audit-records.py" in check_repo,
        "fast baseline must run independent audit records design check",
    )
    require(
        "check-session-tooling-result-materialization-policy.py" in check_repo,
        "fast baseline must run result materialization policy design check",
    )
    require(
        "check-session-tooling-executor-boundary.py" in check_repo,
        "fast baseline must run executor boundary design check",
    )
    require(
        "check-session-tooling-storage-backend-design.py" in check_repo,
        "fast baseline must run storage backend design check",
    )
    require(
        "check-session-tooling-deny-by-default-implementation-gates.py" in check_repo,
        "fast baseline must run deny-by-default implementation gates check",
    )


def main() -> int:
    expected_summary = require_object(load_json_document(FIXTURE_PATH), "status summary must be object")
    check_summary_shape(expected_summary)
    actual_summary = build_summary()
    if actual_summary != expected_summary:
        raise SystemExit("session/tooling foundation status summary does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling foundation status summary checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
