#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

MATRIX = FIXTURE_DIR / "session-tooling-route-negative-coverage-matrix.json"
NEGATIVE_SUITE = FIXTURE_DIR / "session-tooling-negative-regression-suite.json"
NEGATIVE_COVERAGE = FIXTURE_DIR / "session-tooling-negative-coverage-rollup.json"
ROUTE_SMOKE_READINESS = FIXTURE_DIR / "session-tooling-route-smoke-readiness-rollup.json"
DENIED_QUERIES = FIXTURE_DIR / "session-recovery-checkpoint-read-denied-queries.json"
ENABLEMENT_PLAN = FIXTURE_DIR / "session-tooling-executor-storage-confirmation-enablement-plan.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-route-negative-coverage-matrix.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-route-negative-coverage-matrix.py"

AREAS = ["executor", "storage", "confirmation"]
FUTURE_REQUIREMENTS = [
    "executor_gate_route_smoke",
    "storage_materialization_gate_route_smoke",
    "confirmation_gate_route_smoke",
    "durable_store_reader_route_smoke",
]
PROHIBITED_CLAIMS = [
    "complete_negative_regression_suite",
    "p2_short_close",
    "real_tool_executor_ready",
    "durable_storage_ready",
    "confirmation_flow_connected",
    "materialized_result_reader_ready",
    "automatic_replay_ready",
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


def require_string_list_allow_empty(value: Any, message: str) -> list[str]:
    require(isinstance(value, list), message)
    result: list[str] = []
    for item in value:
        text = str(item or "").strip()
        require(bool(text), message)
        result.append(text)
    return result


def coverage_by_area(document: dict[str, Any]) -> dict[str, dict[str, Any]]:
    rows = {
        str(item.get("area") or "").strip(): require_object(item, "coverage area must be object")
        for item in document.get("coverage_by_area") or []
        if isinstance(item, dict)
    }
    require(list(rows) == AREAS, "negative coverage areas drifted")
    return rows


def future_requirement_statuses(document: dict[str, Any]) -> dict[str, str]:
    rows = {
        str(item.get("requirement_id") or "").strip(): str(item.get("status") or "").strip()
        for item in document.get("future_route_smoke_requirements") or []
        if isinstance(item, dict)
    }
    require(list(rows) == FUTURE_REQUIREMENTS, "future route requirement order drifted")
    for requirement_id, status in rows.items():
        require(status == "not_satisfied", f"{requirement_id} must remain not_satisfied")
    return rows


def future_requirement_for_area(area: str) -> str:
    mapping = {
        "executor": "executor_gate_route_smoke",
        "storage": "storage_materialization_gate_route_smoke",
        "confirmation": "confirmation_gate_route_smoke",
    }
    return mapping[area]


def suite_cases(document: dict[str, Any]) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    groups = document.get("groups")
    require(isinstance(groups, list) and len(groups) == 3, "negative suite must include three groups")
    for group in groups:
        group_obj = require_object(group, "negative suite group must be object")
        area = str(group_obj.get("precondition_area") or "").strip()
        require(area in AREAS, f"unexpected suite area: {area}")
        for case in group_obj.get("cases") or []:
            case_obj = require_object(case, "negative suite case must be object")
            case_id = str(case_obj.get("case_id") or "").strip()
            require(case_id, "negative suite case must include case_id")
            rows.append(
                {
                    "area": area,
                    "case_id": case_id,
                    "expected_failure_boundary": str(case_obj.get("expected_failure_boundary") or "").strip(),
                    "expected_error_code": str(case_obj.get("expected_error_code") or "").strip(),
                    "consumer_ids": [
                        str(consumer.get("consumer_id") or "").strip()
                        for consumer in case_obj.get("consumers") or []
                        if isinstance(consumer, dict)
                    ],
                }
            )
    require(len(rows) == 9, "negative suite must keep nine cases")
    return rows


def case_matrix_rows(
    suite: dict[str, Any],
    negative_coverage: dict[str, Any],
    route_smoke: dict[str, Any],
) -> list[dict[str, Any]]:
    coverage = coverage_by_area(negative_coverage)
    statuses = future_requirement_statuses(route_smoke)
    rows: list[dict[str, Any]] = []
    for case in suite_cases(suite):
        area = str(case["area"])
        case_id = str(case["case_id"])
        route_consumed_ids = set(
            require_string_list_allow_empty(coverage[area].get("route_consumed_case_ids"), "route consumed ids required")
        )
        governance_only_ids = set(require_string_list(coverage[area].get("governance_only_case_ids"), "governance-only ids required"))
        if case_id in route_consumed_ids:
            current_status = "covered_by_checkpoint_read_metadata_route"
            current_route_id: str | None = "checkpoint_read_metadata_only"
            future_id: str | None = None
            future_status: str | None = None
            blocker = False
        else:
            require(case_id in governance_only_ids, f"{case_id} must be in governance-only ids")
            current_status = "governance_only_future_route_required"
            current_route_id = None
            future_id = future_requirement_for_area(area)
            future_status = statuses[future_id]
            blocker = True
        rows.append(
            {
                "area": area,
                "case_id": case_id,
                "expected_failure_boundary": case["expected_failure_boundary"],
                "expected_error_code": case["expected_error_code"],
                "current_route_coverage_status": current_status,
                "current_route_id": current_route_id,
                "future_route_requirement_id": future_id,
                "future_requirement_status": future_status,
                "suite_completion_blocker": blocker,
            }
        )
    return rows


def future_requirement_matrix(case_rows: list[dict[str, Any]], route_smoke: dict[str, Any]) -> list[dict[str, Any]]:
    statuses = future_requirement_statuses(route_smoke)
    rows: list[dict[str, Any]] = []
    for requirement_id in FUTURE_REQUIREMENTS:
        rows.append(
            {
                "requirement_id": requirement_id,
                "status": statuses[requirement_id],
                "blocked_suite_case_ids": [
                    str(case["case_id"])
                    for case in case_rows
                    if case.get("future_route_requirement_id") == requirement_id
                ],
            }
        )
    return rows


def build_matrix() -> dict[str, Any]:
    suite = require_object(load_json_document(NEGATIVE_SUITE), "negative suite must be object")
    negative_coverage = require_object(load_json_document(NEGATIVE_COVERAGE), "negative coverage must be object")
    route_smoke = require_object(load_json_document(ROUTE_SMOKE_READINESS), "route smoke readiness must be object")
    denied_queries = require_object(load_json_document(DENIED_QUERIES), "denied queries must be object")
    enablement_plan = require_object(load_json_document(ENABLEMENT_PLAN), "enablement plan must be object")

    require(suite.get("status") == "governance_suite_consumed_deny_by_default_gates_defined", "negative suite status drifted")
    require(negative_coverage.get("status") == "negative_coverage_rollup_governance_only", "negative coverage status drifted")
    require(route_smoke.get("status") == "route_smoke_readiness_governance_only", "route smoke readiness status drifted")
    require(enablement_plan.get("status") == "enablement_plan_defined_blocked", "enablement plan status drifted")

    denied_cases = denied_queries.get("cases")
    require(isinstance(denied_cases, list) and len(denied_cases) == 14, "denied query fixture must keep fourteen cases")
    current_routes = route_smoke.get("current_route_smoke")
    require(isinstance(current_routes, list) and len(current_routes) == 1, "route smoke readiness must keep one current route")
    current_route = require_object(current_routes[0], "current route smoke must be object")

    coverage_totals = require_object(negative_coverage.get("coverage_totals"), "negative coverage totals required")
    route_totals = require_object(route_smoke.get("coverage_totals"), "route smoke totals required")
    case_rows = case_matrix_rows(suite, negative_coverage, route_smoke)
    covered_ids = [str(row["case_id"]) for row in case_rows if row["current_route_coverage_status"] == "covered_by_checkpoint_read_metadata_route"]
    blocker_count = sum(1 for row in case_rows if row["suite_completion_blocker"] is True)

    return {
        "schema_version": 1,
        "kind": "session_tooling_route_negative_coverage_matrix",
        "stage": "P2 Session & Tooling Foundation",
        "status": "route_negative_coverage_matrix_governance_only",
        "implementation_status": "not_ready",
        "source_negative_regression_suite": relative_path(NEGATIVE_SUITE),
        "source_negative_coverage_rollup": relative_path(NEGATIVE_COVERAGE),
        "source_route_smoke_readiness_rollup": relative_path(ROUTE_SMOKE_READINESS),
        "source_denied_query_fixture": relative_path(DENIED_QUERIES),
        "source_enablement_plan": relative_path(ENABLEMENT_PLAN),
        "task_card": relative_path(TASK_CARD),
        "current_route_coverage": {
            "route_id": current_route.get("route_id"),
            "route": current_route.get("route"),
            "status": current_route.get("status"),
            "denied_query_case_count": len(denied_cases),
            "covered_suite_case_ids": covered_ids,
        },
        "case_matrix": case_rows,
        "matrix_totals": {
            "suite_case_count": len(case_rows),
            "current_route_covered_suite_cases": len(covered_ids),
            "governance_only_future_route_required_cases": coverage_totals.get("governance_only_suite_cases"),
            "future_route_requirement_count": route_totals.get("future_requirement_count"),
            "future_route_requirements_satisfied": route_totals.get("future_requirements_satisfied"),
            "suite_completion_blocker_cases": blocker_count,
        },
        "future_route_requirement_matrix": future_requirement_matrix(case_rows, route_smoke),
        "current_allowed_claims": [
            "route_negative_coverage_matrix_checkable",
            "checkpoint_read_route_suite_case_mapping_checkable",
            "future_route_requirement_gaps_checkable",
            "governance_only_route_negative_coverage",
        ],
        "prohibited_claims": PROHIBITED_CLAIMS,
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_route_negative_coverage_matrix",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_matrix_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "matrix schema_version must be 1")
    require(document.get("kind") == "session_tooling_route_negative_coverage_matrix", "matrix kind mismatch")
    require(document.get("status") == "route_negative_coverage_matrix_governance_only", "matrix must stay governance-only")
    require(document.get("implementation_status") == "not_ready", "matrix must not claim implementation readiness")

    route_coverage = require_object(document.get("current_route_coverage"), "matrix must include current_route_coverage")
    require(route_coverage.get("route_id") == "checkpoint_read_metadata_only", "current route id drifted")
    require(route_coverage.get("status") == "covered_metadata_only", "current route status drifted")
    require(route_coverage.get("denied_query_case_count") == 14, "denied query count drifted")
    require(
        route_coverage.get("covered_suite_case_ids")
        == ["executor-ref-checkpoint-read-denied", "materialized-result-read-denied"],
        "covered suite case ids drifted",
    )

    case_rows = document.get("case_matrix")
    require(isinstance(case_rows, list) and len(case_rows) == 9, "matrix must include nine suite cases")
    covered_count = 0
    blocker_count = 0
    for item in case_rows:
        row = require_object(item, "matrix case row must be object")
        require(str(row.get("area") or "").strip() in AREAS, "case row area mismatch")
        status = str(row.get("current_route_coverage_status") or "").strip()
        require(
            status in {"covered_by_checkpoint_read_metadata_route", "governance_only_future_route_required"},
            "case route coverage status mismatch",
        )
        if status == "covered_by_checkpoint_read_metadata_route":
            covered_count += 1
            require(row.get("current_route_id") == "checkpoint_read_metadata_only", "covered case route id mismatch")
            require(row.get("suite_completion_blocker") is False, "covered case must not be suite blocker")
        else:
            blocker_count += 1
            require(row.get("current_route_id") is None, "future route case must not have current route")
            require(row.get("future_requirement_status") == "not_satisfied", "future route status must stay not_satisfied")
            require(row.get("suite_completion_blocker") is True, "future route case must block suite completion")

    totals = require_object(document.get("matrix_totals"), "matrix must include totals")
    require(totals.get("suite_case_count") == 9, "suite case count drifted")
    require(totals.get("current_route_covered_suite_cases") == covered_count == 2, "covered case count drifted")
    require(totals.get("governance_only_future_route_required_cases") == 7, "future route required count drifted")
    require(totals.get("future_route_requirement_count") == 4, "future route requirement count drifted")
    require(totals.get("future_route_requirements_satisfied") == 0, "future route requirements must remain unsatisfied")
    require(totals.get("suite_completion_blocker_cases") == blocker_count == 7, "suite blocker count drifted")

    requirements = document.get("future_route_requirement_matrix")
    require(isinstance(requirements, list) and len(requirements) == 4, "future route requirement matrix must include four rows")
    require(
        [str(item.get("requirement_id") or "").strip() for item in requirements if isinstance(item, dict)]
        == FUTURE_REQUIREMENTS,
        "future route requirement order drifted",
    )
    for item in requirements:
        row = require_object(item, "future route requirement row must be object")
        require(row.get("status") == "not_satisfied", "future route requirement must remain not_satisfied")

    prohibited = set(require_string_list(document.get("prohibited_claims"), "matrix must include prohibited claims"))
    missing = sorted(set(PROHIBITED_CLAIMS) - prohibited)
    require(not missing, f"matrix missing prohibited claims: {missing}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    fixture_name = MATRIX.name
    require(THIS_CHECK.name in check_repo, "fast baseline must run route negative coverage matrix check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference route negative coverage matrix task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", CONTRACTS_README.read_text(encoding="utf-8")),
        ("current focus", CURRENT_FOCUS.read_text(encoding="utf-8")),
        ("capability matrix", CAPABILITY_MATRIX.read_text(encoding="utf-8")),
        ("roadmap", ROADMAP.read_text(encoding="utf-8")),
        ("devlog", DEVLOG.read_text(encoding="utf-8")),
    ):
        require(fixture_name in content, f"{label} must reference route negative coverage matrix fixture")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("P2 short close" in content, f"{label} must mention P2 short close boundary")
        require("not_satisfied" in content, f"{label} must mention not_satisfied future route requirements")
    require("不新增真实 executor route" in task_card, "task card must keep executor route stop line")


def main() -> int:
    expected = require_object(load_json_document(MATRIX), "route negative coverage matrix fixture must be object")
    check_matrix_shape(expected)
    actual = build_matrix()
    if actual != expected:
        raise SystemExit("session/tooling route negative coverage matrix does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling route negative coverage matrix checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
