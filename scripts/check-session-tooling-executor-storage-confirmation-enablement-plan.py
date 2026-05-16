#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

PLAN = FIXTURE_DIR / "session-tooling-executor-storage-confirmation-enablement-plan.json"
PRECONDITIONS = FIXTURE_DIR / "session-tooling-implementation-preconditions.json"
SHORT_CLOSE_DELTA = FIXTURE_DIR / "session-tooling-short-close-readiness-delta.json"
NEGATIVE_COVERAGE = FIXTURE_DIR / "session-tooling-negative-coverage-rollup.json"
ROUTE_SMOKE_READINESS = FIXTURE_DIR / "session-tooling-route-smoke-readiness-rollup.json"
IMPLEMENTATION_GATES = FIXTURE_DIR / "session-tooling-deny-by-default-implementation-gates.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-executor-storage-confirmation-enablement-plan.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-executor-storage-confirmation-enablement-plan.py"

AREAS = ["executor", "storage", "confirmation"]
SHORT_CLOSE_PREREQUISITES = [
    "upper_layer_confirmation_flow",
    "complete_negative_regression_suite",
    "executor_storage_confirmation_enablement_plan",
    "durable_store_and_result_reader_policy",
]
PROHIBITED_CLAIMS = [
    "p2_short_close",
    "real_tool_executor_ready",
    "durable_storage_ready",
    "confirmation_flow_connected",
    "materialized_result_reader_ready",
    "automatic_replay_ready",
    "complete_negative_regression_suite",
]


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
    result: list[str] = []
    for item in value:
        text = str(item or "").strip()
        require(bool(text), message)
        result.append(text)
    return result


def preconditions_by_area(document: dict[str, Any]) -> dict[str, dict[str, Any]]:
    rows = {
        str(item.get("area") or "").strip(): item
        for item in document.get("areas") or []
        if isinstance(item, dict)
    }
    require(list(rows) == AREAS, "implementation preconditions areas drifted")
    for area, row in rows.items():
        require(row.get("status") == "not_ready", f"{area} precondition must remain not_ready")
    return rows


def coverage_by_area(document: dict[str, Any]) -> dict[str, dict[str, Any]]:
    rows = {
        str(item.get("area") or "").strip(): item
        for item in document.get("coverage_by_area") or []
        if isinstance(item, dict)
    }
    require(list(rows) == AREAS, "negative coverage areas drifted")
    for area, row in rows.items():
        require(row.get("real_implementation_consumer") == "missing", f"{area} consumer must remain missing")
        require(row.get("default_decision") == "deny", f"{area} default decision must remain deny")
    return rows


def gates_by_area(document: dict[str, Any]) -> dict[str, dict[str, Any]]:
    rows = {
        str(item.get("gate_area") or "").strip(): item
        for item in document.get("gates") or []
        if isinstance(item, dict)
    }
    require(list(rows) == AREAS, "implementation gate areas drifted")
    for area, row in rows.items():
        require(row.get("default_decision") == "deny", f"{area} gate must default deny")
        require(
            row.get("status") == "deny_by_default_contract_defined",
            f"{area} gate status must stay deny_by_default_contract_defined",
        )
    return rows


def route_requirement_statuses(document: dict[str, Any]) -> dict[str, str]:
    requirements = {
        str(item.get("requirement_id") or "").strip(): str(item.get("status") or "").strip()
        for item in document.get("future_route_smoke_requirements") or []
        if isinstance(item, dict)
    }
    expected = {
        "executor_gate_route_smoke",
        "storage_materialization_gate_route_smoke",
        "confirmation_gate_route_smoke",
        "durable_store_reader_route_smoke",
    }
    require(set(requirements) == expected, "future route smoke requirements drifted")
    for requirement_id, status in requirements.items():
        require(status == "not_satisfied", f"{requirement_id} must remain not_satisfied")
    return requirements


def short_close_prerequisites(document: dict[str, Any]) -> list[dict[str, str]]:
    prerequisites = {
        str(item.get("prerequisite_id") or "").strip(): str(item.get("current_status") or "").strip()
        for item in document.get("hard_prerequisites") or []
        if isinstance(item, dict)
    }
    require(list(prerequisites) == SHORT_CLOSE_PREREQUISITES, "short close prerequisite order drifted")
    for prerequisite_id, status in prerequisites.items():
        require(status == "not_satisfied", f"{prerequisite_id} must remain not_satisfied")
    return [{"condition": prerequisite_id, "status": "not_satisfied"} for prerequisite_id in SHORT_CLOSE_PREREQUISITES]


def area_rows(
    preconditions: dict[str, dict[str, Any]],
    coverage: dict[str, dict[str, Any]],
) -> list[dict[str, Any]]:
    return [
        {
            "area": "executor",
            "current_status": "blocked_not_gated_plan",
            "precondition_status": preconditions["executor"]["status"],
            "default_decision": coverage["executor"]["default_decision"],
            "real_implementation_consumer": coverage["executor"]["real_implementation_consumer"],
            "required_before_gated_plan": [
                "upper_layer_confirmation_flow_not_satisfied",
                "complete_negative_regression_suite_not_satisfied",
                "executor_boundary_design_checkable",
                "independent_audit_records_design_checkable",
                "result_materialization_policy_design_checkable",
                "executor_gate_route_smoke_not_satisfied",
            ],
            "gated_plan_entry_evidence": [
                "executor requests denied before execution side effects",
                "network execution denied unless future policy enables it",
                "independent audit expectation defined for every execution attempt",
                "result materialization remains metadata-only until durable policy is satisfied",
            ],
            "still_forbidden": [
                "real_tool_executor",
                "network_tool_execution",
                "executor_ref_in_checkpoint_read",
                "business_truth_write",
            ],
        },
        {
            "area": "storage",
            "current_status": "blocked_not_gated_plan",
            "precondition_status": preconditions["storage"]["status"],
            "default_decision": coverage["storage"]["default_decision"],
            "real_implementation_consumer": coverage["storage"]["real_implementation_consumer"],
            "required_before_gated_plan": [
                "durable_store_and_result_reader_policy_not_satisfied",
                "complete_negative_regression_suite_not_satisfied",
                "storage_backend_design_checkable",
                "result_materialization_policy_design_checkable",
                "independent_audit_records_design_checkable",
                "storage_materialization_gate_route_smoke_not_satisfied",
            ],
            "gated_plan_entry_evidence": [
                "durable writes denied before storage side effects",
                "materialized result reads denied before result payload is returned",
                "retention redaction and secret handling policy is explicit",
                "business truth writes remain impossible from session/tooling routes",
            ],
            "still_forbidden": [
                "durable_session_store",
                "durable_checkpoint_store",
                "durable_audit_store",
                "durable_result_store",
                "materialized_result_reader",
                "long_term_memory",
            ],
        },
        {
            "area": "confirmation",
            "current_status": "blocked_not_gated_plan",
            "precondition_status": preconditions["confirmation"]["status"],
            "default_decision": coverage["confirmation"]["default_decision"],
            "real_implementation_consumer": coverage["confirmation"]["real_implementation_consumer"],
            "required_before_gated_plan": [
                "upper_layer_confirmation_flow_not_satisfied",
                "complete_negative_regression_suite_not_satisfied",
                "confirmation_flow_design_checkable",
                "independent_audit_records_design_checkable",
                "result_materialization_policy_design_checkable",
                "confirmation_gate_route_smoke_not_satisfied",
            ],
            "gated_plan_entry_evidence": [
                "missing confirmation denied before action execution",
                "stale confirmation denied before action execution",
                "mismatched payload hash denied before action execution",
                "confirmed action handoff remains distinct from advisory candidate output",
            ],
            "still_forbidden": [
                "upper_layer_confirmation_flow_connection",
                "confirmed_action_execution",
                "implicit_confirmation",
                "automatic_replay",
                "business_truth_write",
            ],
        },
    ]


def build_plan() -> dict[str, Any]:
    preconditions_doc = require_object(load_json_document(PRECONDITIONS), "preconditions must be object")
    short_close_doc = require_object(load_json_document(SHORT_CLOSE_DELTA), "short close delta must be object")
    negative_coverage_doc = require_object(load_json_document(NEGATIVE_COVERAGE), "negative coverage must be object")
    route_smoke_doc = require_object(load_json_document(ROUTE_SMOKE_READINESS), "route smoke readiness must be object")
    gates_doc = require_object(load_json_document(IMPLEMENTATION_GATES), "implementation gates must be object")

    require(preconditions_doc.get("status") == "preconditions_not_satisfied", "preconditions must remain not satisfied")
    require(short_close_doc.get("status") == "short_close_blocked", "short close delta must remain blocked")
    require(
        negative_coverage_doc.get("status") == "negative_coverage_rollup_governance_only",
        "negative coverage must remain governance-only",
    )
    require(
        route_smoke_doc.get("status") == "route_smoke_readiness_governance_only",
        "route smoke readiness must remain governance-only",
    )
    require(
        gates_doc.get("status") == "deny_by_default_gates_defined_implementation_blocked",
        "implementation gates must remain blocked",
    )

    precondition_rows = preconditions_by_area(preconditions_doc)
    coverage_rows = coverage_by_area(negative_coverage_doc)
    gates_by_area(gates_doc)
    route_requirement_statuses(route_smoke_doc)

    return {
        "schema_version": 1,
        "kind": "session_tooling_executor_storage_confirmation_enablement_plan",
        "stage": "P2 Session & Tooling Foundation",
        "status": "enablement_plan_defined_blocked",
        "implementation_status": "not_ready",
        "source_implementation_preconditions": relative_path(PRECONDITIONS),
        "source_short_close_readiness_delta": relative_path(SHORT_CLOSE_DELTA),
        "source_negative_coverage_rollup": relative_path(NEGATIVE_COVERAGE),
        "source_route_smoke_readiness_rollup": relative_path(ROUTE_SMOKE_READINESS),
        "source_deny_by_default_implementation_gates": relative_path(IMPLEMENTATION_GATES),
        "task_card": relative_path(TASK_CARD),
        "enablement_areas": area_rows(precondition_rows, coverage_rows),
        "shared_entry_conditions": short_close_prerequisites(short_close_doc),
        "current_allowed_claims": [
            "enablement_plan_defined",
            "gated_plan_entry_conditions_checkable",
            "implementation_still_blocked",
        ],
        "prohibited_claims": PROHIBITED_CLAIMS,
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_enablement_plan",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_plan_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "enablement plan schema_version must be 1")
    require(
        document.get("kind") == "session_tooling_executor_storage_confirmation_enablement_plan",
        "enablement plan kind mismatch",
    )
    require(document.get("status") == "enablement_plan_defined_blocked", "enablement plan status drifted")
    require(document.get("implementation_status") == "not_ready", "enablement plan must not claim implementation")

    areas = document.get("enablement_areas")
    require(isinstance(areas, list) and len(areas) == 3, "enablement plan must include three areas")
    require([str(item.get("area") or "").strip() for item in areas if isinstance(item, dict)] == AREAS, "area order drifted")
    for item in areas:
        area = require_object(item, "enablement area must be object")
        require(area.get("current_status") == "blocked_not_gated_plan", "area must remain blocked")
        require(area.get("precondition_status") == "not_ready", "area precondition must remain not_ready")
        require(area.get("default_decision") == "deny", "area default decision must remain deny")
        require(area.get("real_implementation_consumer") == "missing", "real implementation consumer must remain missing")
        require_string_list(area.get("required_before_gated_plan"), "area must include required_before_gated_plan")
        require_string_list(area.get("gated_plan_entry_evidence"), "area must include evidence")
        require_string_list(area.get("still_forbidden"), "area must include forbidden list")

    entry_conditions = document.get("shared_entry_conditions")
    require(isinstance(entry_conditions, list) and len(entry_conditions) == 4, "shared entry conditions must include four items")
    require(
        [str(item.get("condition") or "").strip() for item in entry_conditions if isinstance(item, dict)]
        == SHORT_CLOSE_PREREQUISITES,
        "shared entry condition order drifted",
    )
    for item in entry_conditions:
        condition = require_object(item, "entry condition must be object")
        require(condition.get("status") == "not_satisfied", "entry conditions must remain not_satisfied")

    prohibited = set(require_string_list(document.get("prohibited_claims"), "prohibited claims required"))
    missing_prohibited = sorted(set(PROHIBITED_CLAIMS) - prohibited)
    require(not missing_prohibited, f"enablement plan missing prohibited claims: {missing_prohibited}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    fixture_name = PLAN.name
    require(THIS_CHECK.name in check_repo, "fast baseline must run enablement plan check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference enablement plan task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", CONTRACTS_README.read_text(encoding="utf-8")),
        ("current focus", CURRENT_FOCUS.read_text(encoding="utf-8")),
        ("capability matrix", CAPABILITY_MATRIX.read_text(encoding="utf-8")),
        ("roadmap", ROADMAP.read_text(encoding="utf-8")),
        ("devlog", DEVLOG.read_text(encoding="utf-8")),
    ):
        require(fixture_name in content, f"{label} must reference enablement plan fixture")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("P2 short close" in content, f"{label} must mention P2 short close boundary")
        require("not_satisfied" in content, f"{label} must mention not_satisfied entry conditions")
    require("不实现真实工具执行器" in task_card, "task card must keep executor stop line")
    require("不实现 durable session" in task_card, "task card must keep durable store stop line")


def main() -> int:
    expected = require_object(load_json_document(PLAN), "enablement plan fixture must be object")
    check_plan_shape(expected)
    actual = build_plan()
    if actual != expected:
        raise SystemExit("session/tooling executor/storage/confirmation enablement plan does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling executor/storage/confirmation enablement plan checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
