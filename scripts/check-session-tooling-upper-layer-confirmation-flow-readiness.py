#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

READINESS = FIXTURE_DIR / "session-tooling-upper-layer-confirmation-flow-readiness.json"
CONFIRMATION_FLOW_DESIGN = FIXTURE_DIR / "session-tooling-confirmation-flow-design.json"
IMPLEMENTATION_PRECONDITIONS = FIXTURE_DIR / "session-tooling-implementation-preconditions.json"
NEGATIVE_REGRESSION_SUITE = FIXTURE_DIR / "session-tooling-negative-regression-suite.json"
SHORT_CLOSE_ENTRY_CHECKLIST = FIXTURE_DIR / "session-tooling-short-close-entry-checklist.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-upper-layer-confirmation-flow-readiness.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-upper-layer-confirmation-flow-readiness.py"

REQUIRED_GATES = [
    "confirmation_request_handoff_contract",
    "confirmation_decision_binding",
    "confirmation_negative_gate_consumers",
    "confirmed_action_handoff_boundary",
]
REQUIRED_NEGATIVE_CASES = {
    "missing-confirmation-denied": "CONFIRMATION_REQUIRED",
    "stale-confirmation-denied": "CONFIRMATION_STALE",
    "mismatched-confirmation-payload-denied": "CONFIRMATION_PAYLOAD_MISMATCH",
}
REQUIRED_PROHIBITED_CLAIMS = {
    "p2_short_close",
    "confirmation_flow_connected",
    "confirmed_action_execution",
    "business_truth_write",
    "automatic_replay",
    "complete_negative_regression_suite",
}
REQUIRED_NOT_ENABLED = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_confirmation_store",
    "real_tool_executor",
    "upper_layer_confirmation_flow_connection",
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
        text = str(item or "").strip()
        require(bool(text), message)
        items.append(text)
    return items


def confirmation_area(preconditions: dict[str, Any]) -> dict[str, Any]:
    area = next(
        (
            item
            for item in preconditions.get("areas") or []
            if isinstance(item, dict) and item.get("area") == "confirmation"
        ),
        None,
    )
    result = require_object(area, "implementation preconditions must include confirmation area")
    require(result.get("status") == "not_ready", "confirmation precondition must remain not_ready")
    blockers = set(require_string_list(result.get("required_before_enablement"), "confirmation required_before_enablement missing"))
    require("upper_layer_confirmation_flow" in blockers, "confirmation precondition must require upper_layer_confirmation_flow")
    return result


def entry_condition(entry_checklist: dict[str, Any]) -> dict[str, Any]:
    condition = next(
        (
            item
            for item in entry_checklist.get("entry_conditions") or []
            if isinstance(item, dict) and item.get("condition_id") == "upper_layer_confirmation_flow"
        ),
        None,
    )
    result = require_object(condition, "entry checklist must include upper_layer_confirmation_flow")
    require(result.get("status") == "not_satisfied", "upper_layer_confirmation_flow must remain not_satisfied")
    return result


def confirmation_cases(negative_suite: dict[str, Any]) -> list[dict[str, Any]]:
    group = next(
        (
            item
            for item in negative_suite.get("groups") or []
            if isinstance(item, dict) and item.get("precondition_area") == "confirmation"
        ),
        None,
    )
    group_obj = require_object(group, "negative suite must include confirmation group")
    require(group_obj.get("current_status") == "governance_consumed_implementation_blocked", "confirmation group must remain blocked")
    cases: list[dict[str, Any]] = []
    for item in group_obj.get("cases") or []:
        case = require_object(item, "confirmation negative case must be object")
        case_id = str(case.get("case_id") or "").strip()
        if case_id not in REQUIRED_NEGATIVE_CASES:
            continue
        require(case.get("expected_error_code") == REQUIRED_NEGATIVE_CASES[case_id], f"{case_id} error code drifted")
        alignment = require_object(case.get("implementation_gate_alignment"), f"{case_id} implementation gate alignment missing")
        require(
            alignment.get("current_status") == "deny_by_default_gate_contract_defined_implementation_blocked",
            f"{case_id} must remain implementation-blocked",
        )
        cases.append(case)
    require([case["case_id"] for case in cases] == list(REQUIRED_NEGATIVE_CASES), "confirmation negative case order drifted")
    return cases


def build_readiness_gates(design: dict[str, Any], preconditions: dict[str, Any]) -> list[dict[str, Any]]:
    scope = require_object(design.get("scope"), "confirmation flow design scope required")
    require(scope.get("upper_layer_integration") == "not_connected", "upper layer integration must remain not_connected")
    require(scope.get("execution_enabled") is False, "confirmation flow design must not enable execution")
    require(scope.get("writes_business_truth") is False, "confirmation flow design must not write business truth")

    record_shape = require_object(design.get("confirmation_record_shape"), "confirmation record shape required")
    fields = set(require_string_list(record_shape.get("required_fields"), "confirmation record fields required"))
    for field in ("session_id", "turn_id", "action_ref", "action_hash", "outcome", "actor_ref", "audit_ref"):
        require(field in fields, f"confirmation record must include {field}")

    outcomes = {
        str(item.get("outcome") or "").strip(): require_object(item, "confirmation outcome must be object")
        for item in record_shape.get("outcomes") or []
        if isinstance(item, dict)
    }
    require(set(outcomes) == {"approve", "reject", "defer"}, "confirmation outcomes must remain approve/reject/defer")
    approve = outcomes["approve"]
    require(
        approve.get("allowed_next_state") == "approved_pending_execution_boundary",
        "approve must stop at approved_pending_execution_boundary",
    )

    confirmation = confirmation_area(preconditions)
    acceptance_gates = set(require_string_list(confirmation.get("acceptance_gates"), "confirmation acceptance gates missing"))
    require(
        "upper layer owns explicit approve, reject, and defer outcomes" in acceptance_gates,
        "confirmation preconditions must keep upper-layer outcome ownership",
    )

    return [
        {
            "gate_id": "confirmation_request_handoff_contract",
            "status": "not_satisfied",
            "current_evidence": "confirmation record shape and approve/reject/defer outcomes are design-checkable only",
            "required_evidence_before_connection": [
                "upper layer request handoff shape includes session_id, turn_id, action_ref, action_hash, risk_level, candidate_action_summary, and expires_at",
                "handoff explicitly separates advisory candidate output from confirmed action execution",
                "high-risk action requests cannot bypass confirmation_request_handoff",
            ],
            "blocked_claims": [
                "confirmation_flow_connected",
                "confirmed_action_execution",
                "p2_short_close",
            ],
        },
        {
            "gate_id": "confirmation_decision_binding",
            "status": "not_satisfied",
            "current_evidence": "confirmation record requires audit_ref, but no upper-layer decision binding exists",
            "required_evidence_before_connection": [
                "approve, reject, and defer decisions are bound to confirmation_id and action_hash",
                "decision actor_ref and confirmed_at are recorded without writing business truth",
                "decision audit_ref points to an independent audit event rather than checkpoint read metadata",
            ],
            "blocked_claims": [
                "confirmation_flow_connected",
                "independent_audit_records_ready",
                "p2_short_close",
            ],
        },
        {
            "gate_id": "confirmation_negative_gate_consumers",
            "status": "not_satisfied",
            "current_evidence": "missing, stale, and mismatched confirmation cases are present in governance suite only",
            "required_evidence_before_connection": [
                "missing confirmation denial is consumed by a real confirmation gate before action execution",
                "stale confirmation denial is consumed by a real confirmation gate before replay or action execution",
                "mismatched action_hash denial is consumed before executor, result ref, or business truth write side effects",
            ],
            "blocked_claims": [
                "complete_negative_regression_suite",
                "confirmation_flow_connected",
                "confirmed_action_execution",
            ],
        },
        {
            "gate_id": "confirmed_action_handoff_boundary",
            "status": "not_satisfied",
            "current_evidence": "approve outcome records approval only and stops at approved_pending_execution_boundary",
            "required_evidence_before_connection": [
                "approved decisions remain pending until executor/storage enablement gates are satisfied",
                "confirmed action handoff cannot write RadishFlow, Radish, or RadishCatalyst truth sources",
                "automatic replay remains disabled even when confirmation exists",
            ],
            "blocked_claims": [
                "confirmed_action_execution",
                "business_truth_write",
                "automatic_replay",
            ],
        },
    ]


def build_negative_case_alignment(negative_suite: dict[str, Any]) -> list[dict[str, Any]]:
    return [
        {
            "case_id": str(case.get("case_id") or "").strip(),
            "expected_error_code": str(case.get("expected_error_code") or "").strip(),
            "current_consumer_status": "governance_consumed_implementation_blocked",
            "future_consumer_required": True,
        }
        for case in confirmation_cases(negative_suite)
    ]


def build_readiness() -> dict[str, Any]:
    design = require_object(load_json_document(CONFIRMATION_FLOW_DESIGN), "confirmation flow design must be object")
    preconditions = require_object(load_json_document(IMPLEMENTATION_PRECONDITIONS), "implementation preconditions must be object")
    negative_suite = require_object(load_json_document(NEGATIVE_REGRESSION_SUITE), "negative regression suite must be object")
    entry = require_object(load_json_document(SHORT_CLOSE_ENTRY_CHECKLIST), "short close entry checklist must be object")

    require(design.get("status") == "design_only_not_connected", "confirmation flow design status drifted")
    require(preconditions.get("status") == "preconditions_not_satisfied", "implementation preconditions status drifted")
    require(negative_suite.get("implementation_status") == "not_ready", "negative suite must keep implementation not_ready")
    require(entry.get("status") == "entry_blocked_governance_only", "entry checklist must remain blocked")

    entry_row = entry_condition(entry)
    gates = build_readiness_gates(design, preconditions)
    negative_cases = build_negative_case_alignment(negative_suite)

    return {
        "schema_version": 1,
        "kind": "session_tooling_upper_layer_confirmation_flow_readiness",
        "stage": "P2 Session & Tooling Foundation",
        "status": "readiness_defined_not_connected",
        "implementation_status": "not_ready",
        "source_confirmation_flow_design": relative_path(CONFIRMATION_FLOW_DESIGN),
        "source_implementation_preconditions": relative_path(IMPLEMENTATION_PRECONDITIONS),
        "source_negative_regression_suite": relative_path(NEGATIVE_REGRESSION_SUITE),
        "source_short_close_entry_checklist": relative_path(SHORT_CLOSE_ENTRY_CHECKLIST),
        "task_card": relative_path(TASK_CARD),
        "readiness_scope": {
            "entry_condition": "upper_layer_confirmation_flow",
            "current_status": entry_row.get("status"),
            "upper_layer_integration": "not_connected",
            "execution_enabled": False,
            "writes_business_truth": False,
            "durable_confirmation_store_enabled": False,
        },
        "handoff_readiness_gates": gates,
        "negative_case_alignment": negative_cases,
        "entry_checklist_alignment": {
            "condition_id": "upper_layer_confirmation_flow",
            "status": entry_row.get("status"),
            "source": relative_path(SHORT_CLOSE_ENTRY_CHECKLIST),
        },
        "readiness_totals": {
            "gate_count": len(gates),
            "satisfied_count": 0,
            "not_satisfied_count": len(gates),
            "negative_case_count": len(negative_cases),
            "future_consumer_required_count": len(negative_cases),
        },
        "current_allowed_claims": [
            "upper_layer_confirmation_readiness_checkable",
            "confirmation_handoff_required_evidence_identified",
            "confirmation_negative_gate_consumers_identified",
            "governance_only_readiness_status",
        ],
        "prohibited_claims": [
            "p2_short_close",
            "confirmation_flow_connected",
            "confirmed_action_execution",
            "business_truth_write",
            "automatic_replay",
            "complete_negative_regression_suite",
        ],
        "not_enabled_capabilities": [
            "automatic_replay",
            "business_truth_write",
            "confirmed_action_execution",
            "durable_confirmation_store",
            "real_tool_executor",
            "upper_layer_confirmation_flow_connection",
        ],
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_upper_layer_confirmation_flow_readiness",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_readiness_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "readiness schema_version must be 1")
    require(document.get("kind") == "session_tooling_upper_layer_confirmation_flow_readiness", "readiness kind mismatch")
    require(document.get("status") == "readiness_defined_not_connected", "readiness must remain not connected")
    require(document.get("implementation_status") == "not_ready", "readiness must not claim implementation readiness")

    scope = require_object(document.get("readiness_scope"), "readiness scope required")
    require(scope.get("entry_condition") == "upper_layer_confirmation_flow", "readiness scope entry condition mismatch")
    require(scope.get("current_status") == "not_satisfied", "upper-layer confirmation flow must remain not_satisfied")
    require(scope.get("upper_layer_integration") == "not_connected", "upper layer integration must remain not_connected")
    require(scope.get("execution_enabled") is False, "readiness must not enable execution")
    require(scope.get("writes_business_truth") is False, "readiness must not write business truth")

    gates = document.get("handoff_readiness_gates")
    require(isinstance(gates, list) and len(gates) == len(REQUIRED_GATES), "readiness gate count drifted")
    require([str(row.get("gate_id") or "").strip() for row in gates if isinstance(row, dict)] == REQUIRED_GATES, "readiness gate order drifted")
    for item in gates:
        row = require_object(item, "readiness gate must be object")
        require(row.get("status") == "not_satisfied", "readiness gates must remain not_satisfied")
        require_string_list(row.get("required_evidence_before_connection"), "readiness gate must include required evidence")
        require_string_list(row.get("blocked_claims"), "readiness gate must include blocked claims")

    cases = document.get("negative_case_alignment")
    require(isinstance(cases, list) and len(cases) == len(REQUIRED_NEGATIVE_CASES), "negative case alignment count drifted")
    require([str(row.get("case_id") or "").strip() for row in cases if isinstance(row, dict)] == list(REQUIRED_NEGATIVE_CASES), "negative case order drifted")
    for item in cases:
        row = require_object(item, "negative case alignment row must be object")
        case_id = str(row.get("case_id") or "").strip()
        require(row.get("expected_error_code") == REQUIRED_NEGATIVE_CASES[case_id], f"{case_id} error code drifted")
        require(row.get("current_consumer_status") == "governance_consumed_implementation_blocked", f"{case_id} must stay governance-only")
        require(row.get("future_consumer_required") is True, f"{case_id} must require future consumer")

    totals = require_object(document.get("readiness_totals"), "readiness totals required")
    require(totals.get("gate_count") == 4, "readiness gate total drifted")
    require(totals.get("satisfied_count") == 0, "readiness must not mark gates satisfied")
    require(totals.get("not_satisfied_count") == 4, "readiness must keep all gates unsatisfied")
    require(totals.get("negative_case_count") == 3, "negative case total drifted")
    require(totals.get("future_consumer_required_count") == 3, "future consumer total drifted")

    prohibited = set(require_string_list(document.get("prohibited_claims"), "readiness prohibited claims required"))
    missing_prohibited = sorted(REQUIRED_PROHIBITED_CLAIMS - prohibited)
    require(not missing_prohibited, f"readiness missing prohibited claims: {missing_prohibited}")
    not_enabled = set(require_string_list(document.get("not_enabled_capabilities"), "readiness not_enabled capabilities required"))
    missing_not_enabled = sorted(REQUIRED_NOT_ENABLED - not_enabled)
    require(not missing_not_enabled, f"readiness missing not-enabled capabilities: {missing_not_enabled}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    fixture_name = READINESS.name
    require(THIS_CHECK.name in check_repo, "fast baseline must run upper-layer confirmation flow readiness check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference upper-layer confirmation readiness task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", CONTRACTS_README.read_text(encoding="utf-8")),
        ("current focus", CURRENT_FOCUS.read_text(encoding="utf-8")),
        ("capability matrix", CAPABILITY_MATRIX.read_text(encoding="utf-8")),
        ("roadmap", ROADMAP.read_text(encoding="utf-8")),
        ("devlog", DEVLOG.read_text(encoding="utf-8")),
    ):
        require(fixture_name in content, f"{label} must reference upper-layer confirmation readiness fixture")
        require("not_satisfied" in content, f"{label} must mention not_satisfied")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("P2 short close" in content, f"{label} must mention P2 short close")


def main() -> int:
    expected = require_object(load_json_document(READINESS), "upper-layer confirmation readiness fixture must be object")
    check_readiness_shape(expected)
    actual = build_readiness()
    if actual != expected:
        raise SystemExit("session/tooling upper-layer confirmation flow readiness does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling upper-layer confirmation flow readiness checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
