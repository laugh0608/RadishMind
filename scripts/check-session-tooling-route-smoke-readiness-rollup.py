#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

ROLLUP = FIXTURE_DIR / "session-tooling-route-smoke-readiness-rollup.json"
ROUTE_SMOKE_COVERAGE = FIXTURE_DIR / "session-recovery-checkpoint-route-smoke-coverage-summary.json"
NEGATIVE_COVERAGE_ROLLUP = FIXTURE_DIR / "session-tooling-negative-coverage-rollup.json"
NEGATIVE_SUITE_READINESS = FIXTURE_DIR / "session-tooling-negative-regression-suite-readiness.json"
IMPLEMENTATION_GATES = FIXTURE_DIR / "session-tooling-deny-by-default-implementation-gates.json"
SHORT_CLOSE_DELTA = FIXTURE_DIR / "session-tooling-short-close-readiness-delta.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-route-smoke-readiness-rollup.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-route-smoke-readiness-rollup.py"

REQUIRED_CURRENT_ROUTES = {"checkpoint_read_metadata_only"}
REQUIRED_FUTURE_REQUIREMENTS = {
    "executor_gate_route_smoke",
    "storage_materialization_gate_route_smoke",
    "confirmation_gate_route_smoke",
    "durable_store_reader_route_smoke",
}
REQUIRED_BLOCKED_CAPABILITIES = {
    "automatic_replay",
    "confirmed_action_execution",
    "durable_audit_store",
    "durable_checkpoint_store",
    "durable_result_store",
    "durable_session_store",
    "materialized_result_reader",
    "real_tool_executor",
    "replay_executor",
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


def gate_statuses(gates: dict[str, Any]) -> dict[str, str]:
    statuses: dict[str, str] = {}
    for item in gates.get("gates") or []:
        gate = require_object(item, "implementation gate must be object")
        area = str(gate.get("gate_area") or "").strip()
        status = str(gate.get("status") or "").strip()
        decision = str(gate.get("default_decision") or "").strip()
        require(decision == "deny", f"{area} gate must default to deny")
        statuses[area] = status
    require(set(statuses) == {"executor", "storage", "confirmation"}, "implementation gates must cover executor/storage/confirmation")
    return statuses


def build_rollup() -> dict[str, Any]:
    route_smoke = require_object(load_json_document(ROUTE_SMOKE_COVERAGE), "route smoke coverage must be object")
    negative_coverage = require_object(load_json_document(NEGATIVE_COVERAGE_ROLLUP), "negative coverage rollup must be object")
    suite_readiness = require_object(load_json_document(NEGATIVE_SUITE_READINESS), "suite readiness must be object")
    gates = require_object(load_json_document(IMPLEMENTATION_GATES), "implementation gates must be object")
    short_close_delta = require_object(load_json_document(SHORT_CLOSE_DELTA), "short close delta must be object")

    require(route_smoke.get("kind") == "session_recovery_checkpoint_route_smoke_coverage_summary", "route smoke kind mismatch")
    require(route_smoke.get("route") == "/v1/session/recovery/checkpoints/{checkpoint_id}", "unexpected route smoke route")
    require(
        negative_coverage.get("status") == "negative_coverage_rollup_governance_only",
        "negative coverage rollup must stay governance-only",
    )
    require(
        suite_readiness.get("status") == "governance_suite_consumed_deny_by_default_gates_defined",
        "suite readiness must stay governance-only",
    )
    require(
        gates.get("status") == "deny_by_default_gates_defined_implementation_blocked",
        "implementation gates must stay blocked",
    )
    require(short_close_delta.get("status") == "short_close_blocked", "short close delta must stay blocked")
    statuses = gate_statuses(gates)

    positive_smoke = require_object(route_smoke.get("positive_smoke"), "route smoke must include positive_smoke")
    negative_smoke = require_object(route_smoke.get("negative_smoke"), "route smoke must include negative_smoke")
    totals = require_object(negative_coverage.get("coverage_totals"), "negative coverage must include totals")
    delta_totals = require_object(short_close_delta.get("delta_totals"), "short close delta must include totals")

    return {
        "schema_version": 1,
        "kind": "session_tooling_route_smoke_readiness_rollup",
        "stage": "P2 Session & Tooling Foundation",
        "status": "route_smoke_readiness_governance_only",
        "implementation_status": "not_ready",
        "source_route_smoke_coverage": relative_path(ROUTE_SMOKE_COVERAGE),
        "source_negative_coverage_rollup": relative_path(NEGATIVE_COVERAGE_ROLLUP),
        "source_negative_regression_suite_readiness": relative_path(NEGATIVE_SUITE_READINESS),
        "source_deny_by_default_implementation_gates": relative_path(IMPLEMENTATION_GATES),
        "source_short_close_readiness_delta": relative_path(SHORT_CLOSE_DELTA),
        "current_route_smoke": [
            {
                "route_id": "checkpoint_read_metadata_only",
                "route": str(route_smoke.get("route") or "").strip(),
                "status": "covered_metadata_only",
                "positive_status": positive_smoke.get("expected_status"),
                "negative_status": negative_smoke.get("expected_status"),
                "negative_case_count": negative_smoke.get("total_cases"),
                "failure_boundary": negative_smoke.get("failure_boundary"),
                "current_claim": "fixture-backed checkpoint read route covers metadata-only response shape and denied materialized result/replay query categories",
            }
        ],
        "future_route_smoke_requirements": [
            {
                "requirement_id": "executor_gate_route_smoke",
                "status": "not_satisfied",
                "gate_status": statuses["executor"],
                "required_before": [
                    "real_tool_executor_ready",
                    "complete_negative_regression_suite",
                    "p2_short_close",
                ],
                "must_prove": [
                    "executor requests are denied by default before any execution side effect",
                    "network execution remains denied unless an explicit future policy enables it",
                    "executor denial emits or explicitly withholds independent audit according to the active audit boundary",
                ],
            },
            {
                "requirement_id": "storage_materialization_gate_route_smoke",
                "status": "not_satisfied",
                "gate_status": statuses["storage"],
                "required_before": [
                    "durable_storage_ready",
                    "materialized_result_reader_ready",
                    "complete_negative_regression_suite",
                    "p2_short_close",
                ],
                "must_prove": [
                    "materialized result reads are denied before any result payload is returned",
                    "durable session/checkpoint/audit/result writes are denied before storage side effects",
                    "business truth writes remain impossible from session/tooling routes",
                ],
            },
            {
                "requirement_id": "confirmation_gate_route_smoke",
                "status": "not_satisfied",
                "gate_status": statuses["confirmation"],
                "required_before": [
                    "confirmation_flow_connected",
                    "confirmed_action_execution",
                    "complete_negative_regression_suite",
                    "p2_short_close",
                ],
                "must_prove": [
                    "missing confirmation is denied before action execution",
                    "stale confirmation is denied before action execution",
                    "mismatched confirmation payload hash is denied before action execution",
                ],
            },
            {
                "requirement_id": "durable_store_reader_route_smoke",
                "status": "not_satisfied",
                "gate_status": "not_implemented",
                "required_before": [
                    "durable_storage_ready",
                    "materialized_result_reader_ready",
                    "automatic_replay_ready",
                    "p2_short_close",
                ],
                "must_prove": [
                    "durable reader routes do not leak secrets or unredacted provider payloads",
                    "retention and redaction policy is visible in route metadata",
                    "reader routes cannot trigger replay or business truth writes",
                ],
            },
        ],
        "coverage_totals": {
            "current_route_smoke_count": 1,
            "future_requirement_count": len(REQUIRED_FUTURE_REQUIREMENTS),
            "future_requirements_satisfied": 0,
            "negative_suite_cases": totals.get("negative_suite_cases"),
            "route_consumed_suite_cases": totals.get("route_consumed_suite_cases"),
            "short_close_hard_prerequisites_not_satisfied": delta_totals.get("not_satisfied_count"),
        },
        "current_allowed_claims": [
            "checkpoint_read_metadata_route_smoke_covered",
            "route_smoke_gap_checkable",
            "route_smoke_readiness_governance_only",
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
                "coverage": "regenerates_and_compares_this_route_smoke_readiness_rollup",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_rollup_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "route smoke readiness schema_version must be 1")
    require(document.get("kind") == "session_tooling_route_smoke_readiness_rollup", "route smoke readiness kind mismatch")
    require(document.get("status") == "route_smoke_readiness_governance_only", "route smoke readiness must stay governance-only")
    require(document.get("implementation_status") == "not_ready", "route smoke readiness must not claim implementation readiness")

    current_routes = document.get("current_route_smoke")
    require(isinstance(current_routes, list) and current_routes, "route smoke readiness must include current_route_smoke")
    route_ids = {str(item.get("route_id") or "").strip() for item in current_routes if isinstance(item, dict)}
    missing_routes = sorted(REQUIRED_CURRENT_ROUTES - route_ids)
    require(not missing_routes, f"route smoke readiness missing current routes: {missing_routes}")
    for item in current_routes:
        route = require_object(item, "current route smoke item must be object")
        require(route.get("status") == "covered_metadata_only", "current route smoke must stay metadata-only")

    future_requirements = document.get("future_route_smoke_requirements")
    require(isinstance(future_requirements, list) and future_requirements, "route smoke readiness must include future requirements")
    requirement_ids = {str(item.get("requirement_id") or "").strip() for item in future_requirements if isinstance(item, dict)}
    missing_requirements = sorted(REQUIRED_FUTURE_REQUIREMENTS - requirement_ids)
    require(not missing_requirements, f"route smoke readiness missing future requirements: {missing_requirements}")
    for item in future_requirements:
        requirement = require_object(item, "future route smoke requirement must be object")
        require(requirement.get("status") == "not_satisfied", "future route smoke requirements must remain not_satisfied")
        require_string_list(requirement.get("required_before"), "future route smoke requirement must include required_before")
        require_string_list(requirement.get("must_prove"), "future route smoke requirement must include must_prove")

    totals = require_object(document.get("coverage_totals"), "route smoke readiness must include totals")
    require(totals.get("current_route_smoke_count") == 1, "route smoke readiness must keep one current route smoke")
    require(totals.get("future_requirement_count") == 4, "route smoke readiness must keep four future requirements")
    require(totals.get("future_requirements_satisfied") == 0, "future route smoke requirements must not be satisfied")

    prohibited = set(require_string_list(document.get("prohibited_claims"), "route smoke readiness must include prohibited claims"))
    missing_prohibited = sorted(REQUIRED_PROHIBITED_CLAIMS - prohibited)
    require(not missing_prohibited, f"route smoke readiness missing prohibited claims: {missing_prohibited}")

    blocked = set(require_string_list(document.get("blocked_capabilities"), "route smoke readiness must include blocked capabilities"))
    missing_blocked = sorted(REQUIRED_BLOCKED_CAPABILITIES - blocked)
    require(not missing_blocked, f"route smoke readiness missing blocked capabilities: {missing_blocked}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    fixture_name = ROLLUP.name

    require(THIS_CHECK.name in check_repo, "fast baseline must run route smoke readiness rollup check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference route smoke readiness task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", CONTRACTS_README.read_text(encoding="utf-8")),
        ("current focus", CURRENT_FOCUS.read_text(encoding="utf-8")),
        ("capability matrix", CAPABILITY_MATRIX.read_text(encoding="utf-8")),
        ("roadmap", ROADMAP.read_text(encoding="utf-8")),
        ("devlog", DEVLOG.read_text(encoding="utf-8")),
    ):
        require(fixture_name in content, f"{label} must reference route smoke readiness fixture")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("P2 short close" in content, f"{label} must mention P2 short close boundary")
        require("not_satisfied" in content, f"{label} must mention not_satisfied future route smoke requirements")


def main() -> int:
    expected_rollup = require_object(load_json_document(ROLLUP), "route smoke readiness fixture must be object")
    check_rollup_shape(expected_rollup)
    actual_rollup = build_rollup()
    if actual_rollup != expected_rollup:
        raise SystemExit("session/tooling route smoke readiness rollup does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling route smoke readiness rollup checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
