#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

ROLLUP = FIXTURE_DIR / "session-tooling-close-candidate-readiness-rollup.json"
FOUNDATION_STATUS = FIXTURE_DIR / "session-tooling-foundation-status-summary.json"
READINESS_SUMMARY = FIXTURE_DIR / "session-tooling-readiness-summary.json"
PRECONDITIONS = FIXTURE_DIR / "session-tooling-implementation-preconditions.json"
NEGATIVE_SKELETON = FIXTURE_DIR / "session-tooling-negative-regression-skeleton.json"
CONFIRMATION_FLOW = FIXTURE_DIR / "session-tooling-confirmation-flow-design.json"
INDEPENDENT_AUDIT = FIXTURE_DIR / "session-tooling-independent-audit-records-design.json"
RESULT_MATERIALIZATION = FIXTURE_DIR / "session-tooling-result-materialization-policy-design.json"
EXECUTOR_BOUNDARY = FIXTURE_DIR / "session-tooling-executor-boundary-design.json"
STORAGE_BACKEND = FIXTURE_DIR / "session-tooling-storage-backend-design.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-close-candidate-readiness-rollup.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-close-candidate-readiness-rollup.py"

REQUIRED_GATES = {
    "confirmation_flow",
    "independent_audit_records",
    "result_materialization_policy",
    "executor_boundary",
    "storage_backend_design",
    "negative_regression_skeleton",
    "implementation_preconditions",
}
REQUIRED_NOT_READY_AREAS = {"executor", "storage", "confirmation"}
REQUIRED_SHORT_CLOSE_PREREQUISITES = {
    "upper_layer_confirmation_flow",
    "complete_negative_regression_suite",
    "executor_storage_confirmation_enablement_plan",
    "durable_store_and_result_reader_policy",
}
REQUIRED_ALLOWED_CLAIMS = {
    "contract_and_metadata_smoke_ready",
    "design_gates_checkable",
    "negative_regression_skeleton_exists",
    "close_candidate_governance_only",
}
REQUIRED_PROHIBITED_CLAIMS = {
    "p2_short_close",
    "real_tool_executor_ready",
    "durable_storage_ready",
    "confirmation_flow_connected",
    "materialized_result_reader_ready",
    "automatic_replay_ready",
    "complete_negative_regression_suite",
}
REQUIRED_BLOCKED_CAPABILITIES = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_audit_store",
    "durable_checkpoint_store",
    "durable_result_store",
    "durable_session_store",
    "durable_tool_store",
    "long_term_memory",
    "materialized_result_reader",
    "real_checkpoint_storage_backend",
    "real_tool_executor",
    "replay_executor",
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


def require_status(document: dict[str, Any], expected: str, label: str) -> None:
    require(document.get("status") == expected, f"{label} status must be {expected}")


def status_after_design(document: dict[str, Any], key: str) -> str:
    statuses = require_object(
        document.get("precondition_status_after_design"),
        f"{key} design must include precondition_status_after_design",
    )
    value = str(statuses.get(key) or "").strip()
    require(bool(value), f"{key} design must include status for {key}")
    return value


def not_ready_areas(preconditions: dict[str, Any]) -> list[dict[str, Any]]:
    areas: list[dict[str, Any]] = []
    for item in preconditions.get("areas") or []:
        area = require_object(item, "precondition area must be object")
        area_id = str(area.get("area") or "").strip()
        require(area_id in REQUIRED_NOT_READY_AREAS, f"unexpected precondition area: {area_id}")
        require(area.get("status") == "not_ready", f"{area_id} must remain not_ready")
        areas.append(
            {
                "area": area_id,
                "status": "not_ready",
                "required_before_enablement": require_string_list(
                    area.get("required_before_enablement"),
                    f"{area_id} must include required_before_enablement",
                ),
            }
        )
    require({area["area"] for area in areas} == REQUIRED_NOT_READY_AREAS, "preconditions must include executor/storage/confirmation")
    return areas


def build_rollup() -> dict[str, Any]:
    foundation = require_object(load_json_document(FOUNDATION_STATUS), "foundation status must be object")
    readiness = require_object(load_json_document(READINESS_SUMMARY), "readiness summary must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS), "preconditions fixture must be object")
    negative_skeleton = require_object(load_json_document(NEGATIVE_SKELETON), "negative skeleton fixture must be object")
    confirmation = require_object(load_json_document(CONFIRMATION_FLOW), "confirmation flow fixture must be object")
    audit = require_object(load_json_document(INDEPENDENT_AUDIT), "independent audit fixture must be object")
    result = require_object(load_json_document(RESULT_MATERIALIZATION), "result materialization fixture must be object")
    executor = require_object(load_json_document(EXECUTOR_BOUNDARY), "executor boundary fixture must be object")
    storage = require_object(load_json_document(STORAGE_BACKEND), "storage backend fixture must be object")

    require_status(foundation, "close_candidate_governance_only", "foundation status")
    require(foundation.get("implementation_status") == "not_implemented", "foundation implementation must stay not implemented")
    require_status(readiness, "metadata_ready_implementation_blocked", "readiness summary")
    require_status(preconditions, "preconditions_not_satisfied", "implementation preconditions")
    require_status(negative_skeleton, "skeleton_only_implementation_blocked", "negative regression skeleton")
    require_status(confirmation, "design_only_not_connected", "confirmation flow")
    require_status(audit, "design_only_not_connected", "independent audit")
    require_status(result, "design_only_not_connected", "result materialization")
    require_status(executor, "design_only_not_connected", "executor boundary")
    require_status(storage, "design_only_not_connected", "storage backend")

    return {
        "schema_version": 1,
        "kind": "session_tooling_close_candidate_readiness_rollup",
        "stage": "P2 Session & Tooling Foundation",
        "status": "close_candidate_governance_only",
        "implementation_status": "not_ready",
        "source_foundation_status_summary": relative_path(FOUNDATION_STATUS),
        "source_readiness_summary": relative_path(READINESS_SUMMARY),
        "source_implementation_preconditions": relative_path(PRECONDITIONS),
        "source_negative_regression_skeleton": relative_path(NEGATIVE_SKELETON),
        "design_gate_rollup": [
            {
                "gate_id": "confirmation_flow",
                "source": relative_path(CONFIRMATION_FLOW),
                "status": confirmation["status"],
                "implementation_ready": False,
                "current_claim": "approve/reject/defer record shape and negative confirmation cases are design-checkable",
                "remaining_blockers": [
                    "upper_layer_confirmation_flow",
                    "independent_audit_records",
                    "negative_regression_suite",
                    "result_materialization_policy",
                ],
            },
            {
                "gate_id": "independent_audit_records",
                "source": relative_path(INDEPENDENT_AUDIT),
                "status": status_after_design(audit, "independent_audit_records"),
                "implementation_ready": False,
                "current_claim": "audit record shape, event sources, and correlation boundaries are design-checkable",
                "remaining_blockers": [
                    "upper_layer_confirmation_flow",
                    "executor_boundary",
                    "storage_backend_design",
                    "result_materialization_policy",
                    "negative_regression_suite",
                ],
            },
            {
                "gate_id": "result_materialization_policy",
                "source": relative_path(RESULT_MATERIALIZATION),
                "status": status_after_design(result, "result_materialization_policy"),
                "implementation_ready": False,
                "current_claim": "metadata-only, future result_ref, and future materialized result boundaries are design-checkable",
                "remaining_blockers": [
                    "upper_layer_confirmation_flow",
                    "executor_boundary",
                    "storage_backend_design",
                    "negative_regression_suite",
                ],
            },
            {
                "gate_id": "executor_boundary",
                "source": relative_path(EXECUTOR_BOUNDARY),
                "status": status_after_design(executor, "executor_boundary"),
                "implementation_ready": False,
                "current_claim": "sandbox, allowlist, execution envelope, timeout/retry, and failure boundaries are design-checkable",
                "remaining_blockers": [
                    "upper_layer_confirmation_flow",
                    "storage_backend_design",
                    "negative_regression_suite",
                ],
            },
            {
                "gate_id": "storage_backend_design",
                "source": relative_path(STORAGE_BACKEND),
                "status": status_after_design(storage, "storage_backend_design"),
                "implementation_ready": False,
                "current_claim": "session, checkpoint, audit, and result store responsibilities are design-checkable",
                "remaining_blockers": [
                    "upper_layer_confirmation_flow",
                    "negative_regression_suite",
                ],
            },
            {
                "gate_id": "negative_regression_skeleton",
                "source": relative_path(NEGATIVE_SKELETON),
                "status": negative_skeleton["status"],
                "implementation_ready": False,
                "current_claim": "blocked executor, storage/materialization, and confirmation cases have skeleton coverage",
                "remaining_blockers": [
                    "complete_negative_regression_suite",
                ],
            },
            {
                "gate_id": "implementation_preconditions",
                "source": relative_path(PRECONDITIONS),
                "status": preconditions["status"],
                "implementation_ready": False,
                "current_claim": "executor, storage, and confirmation are explicitly not_ready",
                "remaining_blockers": [
                    "executor",
                    "storage",
                    "confirmation",
                ],
            },
        ],
        "not_ready_areas": not_ready_areas(preconditions),
        "short_close_prerequisites": [
            {
                "condition": "upper_layer_confirmation_flow",
                "status": "not_satisfied",
                "reason": "no upper-layer approve/reject/defer integration exists",
            },
            {
                "condition": "complete_negative_regression_suite",
                "status": "not_satisfied",
                "reason": "current negative coverage is skeleton_only and does not prove implementation boundaries",
            },
            {
                "condition": "executor_storage_confirmation_enablement_plan",
                "status": "not_satisfied",
                "reason": "executor, storage, and confirmation preconditions remain not_ready",
            },
            {
                "condition": "durable_store_and_result_reader_policy",
                "status": "not_satisfied",
                "reason": "durable stores and materialized result readers are not implemented or enabled",
            },
        ],
        "current_allowed_claims": [
            "contract_and_metadata_smoke_ready",
            "design_gates_checkable",
            "negative_regression_skeleton_exists",
            "close_candidate_governance_only",
        ],
        "prohibited_claims": [
            "p2_short_close",
            "real_tool_executor_ready",
            "durable_storage_ready",
            "confirmation_flow_connected",
            "materialized_result_reader_ready",
            "automatic_replay_ready",
            "complete_negative_regression_suite",
        ],
        "blocked_capabilities": sorted(REQUIRED_BLOCKED_CAPABILITIES),
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_close_candidate_rollup",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_rollup_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "rollup schema_version must be 1")
    require(document.get("kind") == "session_tooling_close_candidate_readiness_rollup", "rollup kind mismatch")
    require(document.get("status") == "close_candidate_governance_only", "rollup must remain governance-only")
    require(document.get("implementation_status") == "not_ready", "rollup must not claim implementation readiness")

    gates = document.get("design_gate_rollup")
    require(isinstance(gates, list) and gates, "rollup must include design gates")
    gate_ids = {str(gate.get("gate_id") or "").strip() for gate in gates if isinstance(gate, dict)}
    missing_gates = sorted(REQUIRED_GATES - gate_ids)
    require(not missing_gates, f"rollup missing gates: {missing_gates}")
    for gate in gates:
        gate_obj = require_object(gate, "design gate must be object")
        require(gate_obj.get("implementation_ready") is False, "design gates must not claim implementation_ready")
        require_string_list(gate_obj.get("remaining_blockers"), "design gates must include remaining blockers")

    areas = document.get("not_ready_areas")
    require(isinstance(areas, list) and areas, "rollup must include not_ready_areas")
    area_ids = {str(area.get("area") or "").strip() for area in areas if isinstance(area, dict)}
    missing_areas = sorted(REQUIRED_NOT_READY_AREAS - area_ids)
    require(not missing_areas, f"rollup missing not_ready areas: {missing_areas}")
    for area in areas:
        area_obj = require_object(area, "not_ready area must be object")
        require(area_obj.get("status") == "not_ready", "area status must stay not_ready")

    prereqs = document.get("short_close_prerequisites")
    require(isinstance(prereqs, list) and prereqs, "rollup must include short_close_prerequisites")
    prereq_ids = {str(item.get("condition") or "").strip() for item in prereqs if isinstance(item, dict)}
    missing_prereqs = sorted(REQUIRED_SHORT_CLOSE_PREREQUISITES - prereq_ids)
    require(not missing_prereqs, f"rollup missing short close prerequisites: {missing_prereqs}")
    for item in prereqs:
        prereq = require_object(item, "short close prerequisite must be object")
        require(prereq.get("status") == "not_satisfied", "short close prerequisites must remain not_satisfied")

    allowed = set(require_string_list(document.get("current_allowed_claims"), "rollup must include allowed claims"))
    require(not sorted(REQUIRED_ALLOWED_CLAIMS - allowed), "rollup missing allowed claims")
    prohibited = set(require_string_list(document.get("prohibited_claims"), "rollup must include prohibited claims"))
    missing_prohibited = sorted(REQUIRED_PROHIBITED_CLAIMS - prohibited)
    require(not missing_prohibited, f"rollup missing prohibited claims: {missing_prohibited}")
    blocked = set(require_string_list(document.get("blocked_capabilities"), "rollup must include blocked capabilities"))
    missing_blocked = sorted(REQUIRED_BLOCKED_CAPABILITIES - blocked)
    require(not missing_blocked, f"rollup missing blocked capabilities: {missing_blocked}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    current_focus = CURRENT_FOCUS.read_text(encoding="utf-8")
    capability_matrix = CAPABILITY_MATRIX.read_text(encoding="utf-8")
    roadmap = ROADMAP.read_text(encoding="utf-8")
    devlog = DEVLOG.read_text(encoding="utf-8")
    contracts_readme = CONTRACTS_README.read_text(encoding="utf-8")
    fixture_name = ROLLUP.name

    require(THIS_CHECK.name in check_repo, "fast baseline must run close-candidate readiness rollup check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference close-candidate rollup task card")
    for label, content in (
        ("task card", task_card),
        ("current focus", current_focus),
        ("capability matrix", capability_matrix),
        ("roadmap", roadmap),
        ("devlog", devlog),
        ("contracts README", contracts_readme),
    ):
        require(fixture_name in content, f"{label} must reference close-candidate rollup fixture")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("P2 short close" in content, f"{label} must mention P2 short close boundary")
    require("not_satisfied" in task_card, "task card must list not_satisfied short close prerequisites")


def main() -> int:
    expected_rollup = require_object(load_json_document(ROLLUP), "rollup fixture must be object")
    check_rollup_shape(expected_rollup)
    actual_rollup = build_rollup()
    if actual_rollup != expected_rollup:
        raise SystemExit("session/tooling close-candidate readiness rollup does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling close-candidate readiness rollup checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
