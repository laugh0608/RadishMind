#!/usr/bin/env python3
from __future__ import annotations

import json
from collections import Counter
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
DENIED_QUERY_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-read-denied-queries.json"
SUMMARY_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-route-smoke-coverage-summary.json"
SERVER_TEST = REPO_ROOT / "services/platform/internal/httpapi/server_test.go"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-recovery-route-smoke-coverage.py"

ROUTE = "/v1/session/recovery/checkpoints/{checkpoint_id}"
POSITIVE_REQUEST_PATH = (
    "/v1/session/recovery/checkpoints/session-checkpoint-0001"
    "?session_id=radishflow-session-001&turn_id=turn-0003"
)
POSITIVE_ASSERTIONS = [
    "kind=session_recovery_checkpoint_read_result",
    "api_boundary.implemented=false",
    "api_boundary.response_shape=metadata_refs_only",
    "access_policy.metadata_only=true",
    "access_policy.materialized_results_included=false",
    "access_policy.auto_replay_enabled=false",
    "refs.include=tool_audit",
    "state_summary.contains_materialized_tool_results=false",
    "state_summary.contains_business_truth=false",
    "replay_policy.auto_replay_enabled=false",
    "replay_policy.requires_confirmation_for_actions=true",
    "tool_audit_summary.policy_decision=blocked_tool_execution_disabled",
    "tool_audit_summary.requires_confirmation=true",
    "tool_audit_summary.execution_enabled=false",
    "tool_audit_summary.execution_status=not_executed",
    "tool_audit_summary.result_cache_mode=metadata_only",
    "tool_audit_summary.result_ref=null",
    "tool_audit_summary.durable_memory_written=false",
    "tool_audit_summary.writes_business_truth=false",
]
FORBIDDEN_RESPONSE_MARKERS = [
    "\"auto_replay_enabled\":true",
    "\"durable_memory_enabled\":true",
    "\"executor_ref\"",
    "\"materialized_results_included\":true",
    "\"output_ref\"",
    "\"result_ref\":\"",
    "\"writes_business_truth\":true",
]
SERVER_TEST_MARKERS = [
    "session recovery checkpoint read",
    "session recovery checkpoint read blocks materialized results and replay",
    "session_recovery_checkpoint_read_result",
    "metadata_refs_only",
    "tool_audit",
    "blocked_tool_execution_disabled",
    "not_executed",
    "metadata_only",
    "errorBoundaryNorthboundRequest",
    "loadCheckpointDeniedQueriesFixture",
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


def check_route_test_markers() -> None:
    content = SERVER_TEST.read_text(encoding="utf-8")
    for marker in SERVER_TEST_MARKERS:
        require(marker in content, f"server route smoke test missing marker: {marker}")
    for marker in FORBIDDEN_RESPONSE_MARKERS:
        require(marker in content, f"server route smoke test missing forbidden response marker: {marker}")
    require(
        SUMMARY_FIXTURE.name in THIS_CHECK.read_text(encoding="utf-8"),
        "route smoke coverage check must compare committed summary",
    )
    require(
        "check-session-recovery-route-smoke-coverage.py" in CHECK_REPO.read_text(encoding="utf-8"),
        "fast baseline must run route smoke coverage check",
    )


def build_summary() -> dict[str, Any]:
    denied_queries = load_json_document(DENIED_QUERY_FIXTURE)
    require(isinstance(denied_queries, dict), "denied query fixture must be an object")
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

    return {
        "schema_version": 1,
        "kind": "session_recovery_checkpoint_route_smoke_coverage_summary",
        "route": ROUTE,
        "stage": "P2 Session & Tooling Foundation",
        "positive_smoke": {
            "test_name": "session recovery checkpoint read",
            "request_path": POSITIVE_REQUEST_PATH,
            "expected_status": 200,
            "response_shape_assertions": POSITIVE_ASSERTIONS,
            "forbidden_response_markers": FORBIDDEN_RESPONSE_MARKERS,
        },
        "negative_smoke": {
            "test_name": "session recovery checkpoint read blocks materialized results and replay",
            "fixture": relative_path(DENIED_QUERY_FIXTURE),
            "total_cases": len(cases),
            "categories": dict(sorted(categories.items())),
            "expected_status": 400,
            "expected_error_codes": expected_error_codes,
            "failure_boundary": "northbound_request",
        },
        "consumers": [
            {
                "path": relative_path(SERVER_TEST),
                "coverage": "positive_and_negative_route_smoke",
            },
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
            "materialized_result_reader",
            "replay_executor",
            "result_ref_reader",
            "tool_executor",
        ],
    }


def main() -> int:
    check_route_test_markers()
    expected_summary = load_json_document(SUMMARY_FIXTURE)
    actual_summary = build_summary()
    if actual_summary != expected_summary:
        raise SystemExit("session recovery route smoke coverage summary does not match regenerated output")
    print("session recovery route smoke coverage checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
