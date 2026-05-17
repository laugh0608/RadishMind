#!/usr/bin/env python3
from __future__ import annotations

import json
from collections import Counter
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
DENIED_QUERY_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-read-denied-queries.json"
PROMOTION_GATE_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-tooling-promotion-gates.json"
ROUTE_SMOKE_SUMMARY_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-route-smoke-coverage-summary.json"
NEGATIVE_REGRESSION_SUITE_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-tooling-negative-regression-suite.json"
SUMMARY_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-tooling-negative-consumption-summary.json"

CHECKPOINT_CONTRACT_CHECK = REPO_ROOT / "scripts/check-session-recovery-checkpoint-contract.py"
PROMOTION_GATE_CHECK = REPO_ROOT / "scripts/check-session-tooling-promotion-gates.py"
ROUTE_SMOKE_COVERAGE_CHECK = REPO_ROOT / "scripts/check-session-recovery-route-smoke-coverage.py"
NEGATIVE_REGRESSION_SUITE_CHECK = REPO_ROOT / "scripts/check-session-tooling-negative-regression-suite.py"
SERVER_TEST = REPO_ROOT / "services/platform/internal/httpapi/server_test.go"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-negative-consumption.py"


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


def file_contains(path: Path, text: str) -> bool:
    return text in path.read_text(encoding="utf-8")


def check_consumers() -> None:
    denied_fixture_name = DENIED_QUERY_FIXTURE.name
    promotion_fixture_name = PROMOTION_GATE_FIXTURE.name
    route_smoke_summary_fixture_name = ROUTE_SMOKE_SUMMARY_FIXTURE.name
    negative_regression_suite_fixture_name = NEGATIVE_REGRESSION_SUITE_FIXTURE.name
    summary_fixture_name = SUMMARY_FIXTURE.name

    require(
        file_contains(CHECKPOINT_CONTRACT_CHECK, denied_fixture_name),
        "checkpoint contract check must consume denied query fixture",
    )
    require(
        file_contains(SERVER_TEST, denied_fixture_name),
        "platform server tests must consume denied query fixture",
    )
    require(
        file_contains(PROMOTION_GATE_CHECK, promotion_fixture_name),
        "promotion gate check must consume promotion gate fixture",
    )
    require(
        file_contains(CHECK_REPO, "check-session-tooling-promotion-gates.py"),
        "fast baseline must run promotion gate check",
    )
    require(
        file_contains(ROUTE_SMOKE_COVERAGE_CHECK, route_smoke_summary_fixture_name),
        "route smoke coverage check must compare route smoke summary fixture",
    )
    require(
        file_contains(CHECK_REPO, "check-session-recovery-route-smoke-coverage.py"),
        "fast baseline must run route smoke coverage check",
    )
    require(
        file_contains(NEGATIVE_REGRESSION_SUITE_CHECK, negative_regression_suite_fixture_name),
        "negative regression suite check must compare committed suite fixture",
    )
    require(
        file_contains(CHECK_REPO, "check-session-tooling-negative-regression-suite.py"),
        "fast baseline must run negative regression suite check",
    )
    require(
        file_contains(THIS_CHECK, summary_fixture_name),
        "negative consumption check must compare committed summary",
    )
    require(
        file_contains(CHECK_REPO, "check-session-tooling-negative-consumption.py"),
        "fast baseline must run negative consumption check",
    )


def build_summary() -> dict[str, Any]:
    denied_queries = load_json_document(DENIED_QUERY_FIXTURE)
    promotion_gates = load_json_document(PROMOTION_GATE_FIXTURE)
    negative_regression_suite = load_json_document(NEGATIVE_REGRESSION_SUITE_FIXTURE)
    require(isinstance(denied_queries, dict), "denied query fixture must be an object")
    require(isinstance(promotion_gates, dict), "promotion gate fixture must be an object")
    require(isinstance(negative_regression_suite, dict), "negative regression suite fixture must be an object")

    cases = denied_queries.get("cases")
    require(isinstance(cases, list), "denied query fixture cases must be a list")
    categories = Counter(str(case.get("category") or "").strip() for case in cases if isinstance(case, dict))
    expected_error_codes = sorted(
        {
            str(case.get("expected_error_code") or "").strip()
            for case in cases
            if isinstance(case, dict) and str(case.get("expected_error_code") or "").strip()
        }
    )

    gates = promotion_gates.get("gates")
    require(isinstance(gates, list), "promotion gate fixture gates must be a list")
    gate_ids = sorted(
        str(gate.get("gate_id") or "").strip()
        for gate in gates
        if isinstance(gate, dict) and str(gate.get("gate_id") or "").strip()
    )
    future_gate = next(
        (
            gate
            for gate in gates
            if isinstance(gate, dict) and str(gate.get("gate_id") or "").strip() == "future_implementation_gate"
        ),
        None,
    )
    require(isinstance(future_gate, dict), "promotion gate fixture must include future_implementation_gate")
    future_blockers = sorted(str(item or "").strip() for item in future_gate.get("required_before_enablement") or [])

    suite_groups = negative_regression_suite.get("groups")
    require(isinstance(suite_groups, list), "negative regression suite groups must be a list")
    suite_case_count = 0
    for group in suite_groups:
        if not isinstance(group, dict):
            continue
        suite_case_count += len([case for case in group.get("cases") or [] if isinstance(case, dict)])

    return {
        "schema_version": 1,
        "kind": "session_tooling_negative_consumption_summary",
        "stage": "P2 Session & Tooling Foundation",
        "denied_query_fixture": {
            "path": relative_path(DENIED_QUERY_FIXTURE),
            "total_cases": len(cases),
            "categories": dict(sorted(categories.items())),
            "expected_error_codes": expected_error_codes,
            "consumers": [
                {
                    "path": relative_path(CHECKPOINT_CONTRACT_CHECK),
                    "coverage": "fixture_shape_and_required_case_categories",
                },
                {
                    "path": relative_path(SERVER_TEST),
                    "coverage": "go_route_negative_smoke_cases",
                },
            ],
        },
        "promotion_gate_fixture": {
            "path": relative_path(PROMOTION_GATE_FIXTURE),
            "total_gates": len(gates),
            "gate_ids": gate_ids,
            "future_enablement_blockers": future_blockers,
            "consumers": [
                {
                    "path": relative_path(PROMOTION_GATE_CHECK),
                    "coverage": "promotion_gate_shape_and_blocked_claims",
                },
                {
                    "path": relative_path(CHECK_REPO),
                    "coverage": "fast_baseline_entrypoint",
                },
            ],
        },
        "route_smoke_coverage_summary": {
            "path": relative_path(ROUTE_SMOKE_SUMMARY_FIXTURE),
            "covered_route": "/v1/session/recovery/checkpoints/{checkpoint_id}",
            "consumers": [
                {
                    "path": relative_path(ROUTE_SMOKE_COVERAGE_CHECK),
                    "coverage": "regenerates_and_compares_route_smoke_summary",
                },
                {
                    "path": relative_path(CHECK_REPO),
                    "coverage": "fast_baseline_entrypoint",
                },
            ],
        },
        "negative_regression_suite_fixture": {
            "path": relative_path(NEGATIVE_REGRESSION_SUITE_FIXTURE),
            "total_groups": len(suite_groups),
            "total_cases": suite_case_count,
            "status": str(negative_regression_suite.get("status") or "").strip(),
            "consumers": [
                {
                    "path": relative_path(NEGATIVE_REGRESSION_SUITE_CHECK),
                    "coverage": "regenerates_and_compares_governance_suite_case_consumers",
                },
                {
                    "path": relative_path(CHECK_REPO),
                    "coverage": "fast_baseline_entrypoint",
                },
            ],
        },
        "summary_consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_summary",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
        "blocked_capabilities": [
            "automatic_replay",
            "durable_memory",
            "materialized_tool_results",
            "real_checkpoint_storage_backend",
            "real_tool_executor",
            "result_ref_reader",
        ],
    }


def main() -> int:
    check_consumers()
    expected_summary = load_json_document(SUMMARY_FIXTURE)
    actual_summary = build_summary()
    if actual_summary != expected_summary:
        raise SystemExit("session/tooling negative consumption summary does not match regenerated output")
    print("session/tooling negative consumption checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
