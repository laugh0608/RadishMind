#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

ROLLUP = FIXTURE_DIR / "session-tooling-negative-coverage-rollup.json"
DENIED_QUERY_FIXTURE = FIXTURE_DIR / "session-recovery-checkpoint-read-denied-queries.json"
ROUTE_SMOKE_COVERAGE = FIXTURE_DIR / "session-recovery-checkpoint-route-smoke-coverage-summary.json"
NEGATIVE_CONSUMPTION = FIXTURE_DIR / "session-tooling-negative-consumption-summary.json"
NEGATIVE_SUITE = FIXTURE_DIR / "session-tooling-negative-regression-suite.json"
NEGATIVE_SUITE_READINESS = FIXTURE_DIR / "session-tooling-negative-regression-suite-readiness.json"
IMPLEMENTATION_GATES = FIXTURE_DIR / "session-tooling-deny-by-default-implementation-gates.json"
CLOSE_CANDIDATE_ROLLUP = FIXTURE_DIR / "session-tooling-close-candidate-readiness-rollup.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-negative-coverage-rollup.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-negative-coverage-rollup.py"

REQUIRED_AREAS = {"executor", "storage", "confirmation"}
REQUIRED_COVERAGE_LAYERS = {
    "checkpoint_read_route_smoke",
    "negative_consumption_summary",
    "governance_negative_suite",
    "deny_by_default_implementation_gates",
    "real_implementation_consumers",
}
REQUIRED_PROHIBITED_CLAIMS = {
    "complete_negative_regression_suite",
    "p2_short_close",
    "real_tool_executor_ready",
    "durable_storage_ready",
    "confirmation_flow_connected",
    "materialized_result_reader_ready",
    "automatic_replay_ready",
}
REQUIRED_BLOCKED_CAPABILITIES = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_audit_store",
    "durable_checkpoint_store",
    "durable_result_store",
    "durable_session_store",
    "long_term_memory",
    "materialized_result_reader",
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


def suite_groups_by_area(suite: dict[str, Any]) -> dict[str, dict[str, Any]]:
    groups: dict[str, dict[str, Any]] = {}
    for item in suite.get("groups") or []:
        group = require_object(item, "negative suite group must be object")
        area = str(group.get("precondition_area") or "").strip()
        require(area in REQUIRED_AREAS, f"unexpected negative suite area: {area}")
        require(area not in groups, f"duplicate negative suite area: {area}")
        groups[area] = group
    require(set(groups) == REQUIRED_AREAS, "negative suite must cover executor/storage/confirmation")
    return groups


def gates_by_area(gates: dict[str, Any]) -> dict[str, dict[str, Any]]:
    result: dict[str, dict[str, Any]] = {}
    for item in gates.get("gates") or []:
        gate = require_object(item, "implementation gate must be object")
        area = str(gate.get("gate_area") or "").strip()
        require(area in REQUIRED_AREAS, f"unexpected gate area: {area}")
        result[area] = gate
    require(set(result) == REQUIRED_AREAS, "implementation gates must cover executor/storage/confirmation")
    return result


def coverage_by_area(suite: dict[str, Any], gates: dict[str, Any]) -> list[dict[str, Any]]:
    groups = suite_groups_by_area(suite)
    gate_map = gates_by_area(gates)
    rows: list[dict[str, Any]] = []
    for area in ("executor", "storage", "confirmation"):
        group = groups[area]
        cases = [case for case in group.get("cases") or [] if isinstance(case, dict)]
        route_consumed_cases: list[str] = []
        governance_only_cases: list[str] = []
        for case in cases:
            case_id = str(case.get("case_id") or "").strip()
            consumer_ids = {
                str(consumer.get("consumer_id") or "").strip()
                for consumer in case.get("consumers") or []
                if isinstance(consumer, dict)
            }
            if "checkpoint_denied_query_fixture" in consumer_ids:
                route_consumed_cases.append(case_id)
            else:
                governance_only_cases.append(case_id)

            alignment = require_object(case.get("implementation_gate_alignment"), f"{case_id} must include gate alignment")
            require(alignment.get("gate_area") == area, f"{case_id} gate area mismatch")
            require(
                alignment.get("current_status") == "deny_by_default_gate_contract_defined_implementation_blocked",
                f"{case_id} must align to deny-by-default gate contract",
            )

        gate = gate_map[area]
        rows.append(
            {
                "area": area,
                "suite_case_count": len(cases),
                "route_consumed_case_ids": route_consumed_cases,
                "governance_only_case_ids": governance_only_cases,
                "deny_by_default_gate_status": str(gate.get("status") or "").strip(),
                "default_decision": str(gate.get("default_decision") or "").strip(),
                "real_implementation_consumer": "missing",
                "current_claim": "governance_and_gate_contract_covered_no_real_implementation",
            }
        )
    return rows


def build_rollup() -> dict[str, Any]:
    denied_queries = require_object(load_json_document(DENIED_QUERY_FIXTURE), "denied query fixture must be object")
    route_smoke = require_object(load_json_document(ROUTE_SMOKE_COVERAGE), "route smoke coverage must be object")
    negative_consumption = require_object(load_json_document(NEGATIVE_CONSUMPTION), "negative consumption must be object")
    suite = require_object(load_json_document(NEGATIVE_SUITE), "negative suite must be object")
    suite_readiness = require_object(load_json_document(NEGATIVE_SUITE_READINESS), "suite readiness must be object")
    gates = require_object(load_json_document(IMPLEMENTATION_GATES), "implementation gates must be object")
    close_candidate = require_object(load_json_document(CLOSE_CANDIDATE_ROLLUP), "close candidate rollup must be object")

    require(suite.get("status") == "governance_suite_consumed_deny_by_default_gates_defined", "suite status mismatch")
    require(
        suite_readiness.get("status") == "governance_suite_consumed_deny_by_default_gates_defined",
        "suite readiness status mismatch",
    )
    require(
        gates.get("status") == "deny_by_default_gates_defined_implementation_blocked",
        "implementation gates status mismatch",
    )
    require(close_candidate.get("status") == "close_candidate_governance_only", "close candidate must stay governance-only")

    denied_cases = denied_queries.get("cases")
    require(isinstance(denied_cases, list), "denied query fixture must include cases")
    route_negative = require_object(route_smoke.get("negative_smoke"), "route smoke must include negative_smoke")
    positive_smoke = require_object(route_smoke.get("positive_smoke"), "route smoke must include positive_smoke")
    coverage_rows = coverage_by_area(suite, gates)

    return {
        "schema_version": 1,
        "kind": "session_tooling_negative_coverage_rollup",
        "stage": "P2 Session & Tooling Foundation",
        "status": "negative_coverage_rollup_governance_only",
        "implementation_status": "not_ready",
        "source_denied_query_fixture": relative_path(DENIED_QUERY_FIXTURE),
        "source_route_smoke_coverage": relative_path(ROUTE_SMOKE_COVERAGE),
        "source_negative_consumption_summary": relative_path(NEGATIVE_CONSUMPTION),
        "source_negative_regression_suite": relative_path(NEGATIVE_SUITE),
        "source_negative_regression_suite_readiness": relative_path(NEGATIVE_SUITE_READINESS),
        "source_deny_by_default_implementation_gates": relative_path(IMPLEMENTATION_GATES),
        "source_close_candidate_rollup": relative_path(CLOSE_CANDIDATE_ROLLUP),
        "task_card": relative_path(TASK_CARD),
        "coverage_layers": [
            {
                "layer_id": "checkpoint_read_route_smoke",
                "status": "metadata_only_route_consumed",
                "coverage": "positive response shape and denied query categories are covered by Go route smoke",
                "evidence": [
                    relative_path(ROUTE_SMOKE_COVERAGE),
                    "services/platform/internal/httpapi/server_test.go",
                ],
            },
            {
                "layer_id": "negative_consumption_summary",
                "status": "summary_consumed",
                "coverage": "denied query, promotion gate, route smoke, and governance suite fixtures have consumers",
                "evidence": [
                    relative_path(NEGATIVE_CONSUMPTION),
                    "scripts/check-session-tooling-negative-consumption.py",
                ],
            },
            {
                "layer_id": "governance_negative_suite",
                "status": str(suite.get("status") or "").strip(),
                "coverage": "9 negative cases assert governance consumers, audit non-write boundary, and side-effect absence",
                "evidence": [
                    relative_path(NEGATIVE_SUITE),
                    "scripts/check-session-tooling-negative-regression-suite.py",
                ],
            },
            {
                "layer_id": "deny_by_default_implementation_gates",
                "status": str(gates.get("status") or "").strip(),
                "coverage": "executor, storage, and confirmation gate contracts default to deny",
                "evidence": [
                    relative_path(IMPLEMENTATION_GATES),
                    "scripts/check-session-tooling-deny-by-default-implementation-gates.py",
                ],
            },
            {
                "layer_id": "real_implementation_consumers",
                "status": "missing",
                "coverage": "no real executor, storage, or confirmation implementation consumer exists",
                "evidence": [
                    relative_path(NEGATIVE_SUITE_READINESS),
                    relative_path(CLOSE_CANDIDATE_ROLLUP),
                ],
            },
        ],
        "route_coverage": {
            "route": str(route_smoke.get("route") or "").strip(),
            "positive_response_shape": str(positive_smoke.get("expected_status") or "").strip(),
            "negative_case_count": route_negative.get("total_cases"),
            "negative_categories": route_negative.get("categories"),
            "negative_error_codes": route_negative.get("expected_error_codes"),
            "failure_boundary": route_negative.get("failure_boundary"),
            "forbidden_response_markers": positive_smoke.get("forbidden_response_markers"),
        },
        "coverage_by_area": coverage_rows,
        "coverage_totals": {
            "denied_query_cases": len(denied_cases),
            "negative_suite_cases": sum(row["suite_case_count"] for row in coverage_rows),
            "route_consumed_suite_cases": sum(len(row["route_consumed_case_ids"]) for row in coverage_rows),
            "governance_only_suite_cases": sum(len(row["governance_only_case_ids"]) for row in coverage_rows),
            "deny_by_default_gate_areas": len(coverage_rows),
        },
        "current_allowed_claims": [
            "metadata_only_route_smoke_covered",
            "negative_fixture_consumers_covered",
            "governance_suite_gate_contract_alignment_covered",
            "negative_coverage_rollup_governance_only",
        ],
        "prohibited_claims": [
            "complete_negative_regression_suite",
            "p2_short_close",
            "real_tool_executor_ready",
            "durable_storage_ready",
            "confirmation_flow_connected",
            "materialized_result_reader_ready",
            "automatic_replay_ready",
        ],
        "blocked_capabilities": sorted(REQUIRED_BLOCKED_CAPABILITIES),
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_negative_coverage_rollup",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_rollup_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "negative coverage rollup schema_version must be 1")
    require(document.get("kind") == "session_tooling_negative_coverage_rollup", "negative coverage rollup kind mismatch")
    require(document.get("status") == "negative_coverage_rollup_governance_only", "rollup must stay governance-only")
    require(document.get("implementation_status") == "not_ready", "rollup must not claim implementation readiness")

    layers = document.get("coverage_layers")
    require(isinstance(layers, list) and layers, "rollup must include coverage_layers")
    layer_ids = {str(layer.get("layer_id") or "").strip() for layer in layers if isinstance(layer, dict)}
    missing_layers = sorted(REQUIRED_COVERAGE_LAYERS - layer_ids)
    require(not missing_layers, f"negative coverage rollup missing layers: {missing_layers}")

    areas = document.get("coverage_by_area")
    require(isinstance(areas, list) and areas, "rollup must include coverage_by_area")
    area_ids = {str(area.get("area") or "").strip() for area in areas if isinstance(area, dict)}
    missing_areas = sorted(REQUIRED_AREAS - area_ids)
    require(not missing_areas, f"negative coverage rollup missing areas: {missing_areas}")
    for area in areas:
        area_obj = require_object(area, "coverage area must be object")
        require(area_obj.get("real_implementation_consumer") == "missing", "real implementation consumer must stay missing")
        require(area_obj.get("default_decision") == "deny", "area default decision must stay deny")

    totals = require_object(document.get("coverage_totals"), "rollup must include coverage_totals")
    require(totals.get("negative_suite_cases") == 9, "rollup must keep 9 negative suite cases")
    require(totals.get("deny_by_default_gate_areas") == 3, "rollup must cover 3 gate areas")

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
    fixture_name = ROLLUP.name
    require(THIS_CHECK.name in check_repo, "fast baseline must run negative coverage rollup check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference negative coverage rollup task card")

    for label, content in (
        ("task card", task_card),
        ("contracts README", CONTRACTS_README.read_text(encoding="utf-8")),
        ("current focus", CURRENT_FOCUS.read_text(encoding="utf-8")),
        ("capability matrix", CAPABILITY_MATRIX.read_text(encoding="utf-8")),
        ("roadmap", ROADMAP.read_text(encoding="utf-8")),
        ("devlog", DEVLOG.read_text(encoding="utf-8")),
    ):
        require(fixture_name in content, f"{label} must reference negative coverage rollup fixture")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("P2 short close" in content, f"{label} must mention P2 short close boundary")


def main() -> int:
    expected_rollup = require_object(load_json_document(ROLLUP), "negative coverage rollup fixture must be object")
    check_rollup_shape(expected_rollup)
    actual_rollup = build_rollup()
    if actual_rollup != expected_rollup:
        raise SystemExit("session/tooling negative coverage rollup does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling negative coverage rollup checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
