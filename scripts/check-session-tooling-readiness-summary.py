#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

PROMOTION_GATE_FIXTURE = FIXTURE_DIR / "session-tooling-promotion-gates.json"
NEGATIVE_CONSUMPTION_SUMMARY = FIXTURE_DIR / "session-tooling-negative-consumption-summary.json"
ROUTE_SMOKE_COVERAGE_SUMMARY = FIXTURE_DIR / "session-recovery-checkpoint-route-smoke-coverage-summary.json"
READINESS_SUMMARY = FIXTURE_DIR / "session-tooling-readiness-summary.json"

CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-readiness-summary.py"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"

REQUIRED_COMPLETED_GATE_IDS = {
    "session_contract_gate",
    "tooling_contract_gate",
    "checkpoint_read_contract_gate",
    "promotion_gate_summary",
    "negative_consumption_summary",
    "route_smoke_coverage_summary",
}
REQUIRED_PRECONDITION_AREAS = {"executor", "storage", "confirmation"}
REQUIRED_FUTURE_BLOCKERS = {
    "upper_layer_confirmation_flow",
    "executor_boundary",
    "storage_backend_design",
    "result_materialization_policy",
    "independent_audit_records",
    "negative_regression_suite",
}
REQUIRED_BLOCKED_CAPABILITIES = {
    "automatic_replay",
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
    require(isinstance(value, list), message)
    items: list[str] = []
    for item in value:
        normalized = str(item or "").strip()
        require(bool(normalized), message)
        items.append(normalized)
    return items


def future_gate_from_promotion(promotion_gates: dict[str, Any]) -> dict[str, Any]:
    gates = promotion_gates.get("gates")
    require(isinstance(gates, list), "promotion gate fixture gates must be a list")
    future_gate = next(
        (
            gate
            for gate in gates
            if isinstance(gate, dict) and str(gate.get("gate_id") or "").strip() == "future_implementation_gate"
        ),
        None,
    )
    return require_object(future_gate, "promotion gate fixture must include future_implementation_gate")


def build_summary() -> dict[str, Any]:
    promotion_gates = require_object(load_json_document(PROMOTION_GATE_FIXTURE), "promotion gate fixture must be object")
    negative_summary = require_object(
        load_json_document(NEGATIVE_CONSUMPTION_SUMMARY),
        "negative consumption summary must be object",
    )
    route_summary = require_object(
        load_json_document(ROUTE_SMOKE_COVERAGE_SUMMARY),
        "route smoke coverage summary must be object",
    )

    future_gate = future_gate_from_promotion(promotion_gates)
    future_blockers = set(
        require_string_list(
            future_gate.get("required_before_enablement"),
            "future implementation gate must include required_before_enablement",
        )
    )
    missing_future_blockers = sorted(REQUIRED_FUTURE_BLOCKERS - future_blockers)
    require(not missing_future_blockers, f"future implementation gate missing blockers: {missing_future_blockers}")

    denied_query_fixture = require_object(
        negative_summary.get("denied_query_fixture"),
        "negative consumption summary must include denied_query_fixture",
    )
    promotion_gate_summary = require_object(
        negative_summary.get("promotion_gate_fixture"),
        "negative consumption summary must include promotion_gate_fixture",
    )
    route_smoke_negative = require_object(
        route_summary.get("negative_smoke"),
        "route smoke summary must include negative_smoke",
    )

    return {
        "schema_version": 1,
        "kind": "session_tooling_readiness_summary",
        "stage": "P2 Session & Tooling Foundation",
        "status": "metadata_ready_implementation_blocked",
        "readiness_level": "contract_and_metadata_smoke_ready",
        "completed_gates": [
            {
                "gate_id": "session_contract_gate",
                "status": "complete",
                "evidence": [
                    "contracts/session-record.schema.json",
                    "scripts/checks/fixtures/session-record-basic.json",
                    "scripts/check-session-record-contract.py",
                ],
                "claim": "session identity, bounded history policy, state policy, and recovery record are fixture-checkable",
            },
            {
                "gate_id": "tooling_contract_gate",
                "status": "complete",
                "evidence": [
                    "contracts/tool.schema.json",
                    "contracts/tool-registry.schema.json",
                    "contracts/tool-audit-record.schema.json",
                    "scripts/checks/fixtures/tool-registry-basic.json",
                    "scripts/checks/fixtures/tool-audit-record-basic.json",
                    "scripts/check-tooling-framework-contract.py",
                ],
                "claim": "tool registry, policy, audit, session binding, and metadata-only cache boundaries are fixture-checkable",
            },
            {
                "gate_id": "checkpoint_read_contract_gate",
                "status": "complete",
                "evidence": [
                    "contracts/session-recovery-checkpoint-read.schema.json",
                    "scripts/checks/fixtures/session-recovery-checkpoint-read-basic.json",
                    str(denied_query_fixture.get("path") or "").strip(),
                    "scripts/check-session-recovery-checkpoint-contract.py",
                ],
                "claim": "checkpoint read exposes metadata refs and rejects materialized result, durable memory, and replay queries",
            },
            {
                "gate_id": "promotion_gate_summary",
                "status": "complete",
                "evidence": [
                    str(promotion_gate_summary.get("path") or "").strip(),
                    "scripts/check-session-tooling-promotion-gates.py",
                ],
                "claim": "P2 promotion gates separate current metadata-only claims from future implementation claims",
            },
            {
                "gate_id": "negative_consumption_summary",
                "status": "complete",
                "evidence": [
                    relative_path(NEGATIVE_CONSUMPTION_SUMMARY),
                    "scripts/check-session-tooling-negative-consumption.py",
                ],
                "claim": "negative fixtures and summaries have explicit consumers in contract, route, and fast baseline checks",
            },
            {
                "gate_id": "route_smoke_coverage_summary",
                "status": "complete",
                "evidence": [
                    relative_path(ROUTE_SMOKE_COVERAGE_SUMMARY),
                    "scripts/check-session-recovery-route-smoke-coverage.py",
                ],
                "claim": "checkpoint read route smoke covers metadata-only response shape, denied query categories, and northbound failure boundary",
            },
        ],
        "coverage_counts": {
            "denied_query_cases": denied_query_fixture.get("total_cases"),
            "denied_query_categories": denied_query_fixture.get("categories"),
            "promotion_gate_count": promotion_gate_summary.get("total_gates"),
            "route_smoke_negative_cases": route_smoke_negative.get("total_cases"),
            "route_smoke_expected_error_codes": route_smoke_negative.get("expected_error_codes"),
        },
        "missing_preconditions": [
            {
                "area": "executor",
                "status": "not_ready",
                "required_before_enablement": [
                    "executor_boundary",
                    "result_materialization_policy",
                    "independent_audit_records",
                    "negative_regression_suite",
                ],
                "blocked_until": "tool execution can be sandboxed, audited independently, and covered by negative regression before any executor is enabled",
            },
            {
                "area": "storage",
                "status": "not_ready",
                "required_before_enablement": [
                    "storage_backend_design",
                    "result_materialization_policy",
                    "independent_audit_records",
                    "negative_regression_suite",
                ],
                "blocked_until": "durable session, checkpoint, and tool result storage semantics are designed and separately checked",
            },
            {
                "area": "confirmation",
                "status": "not_ready",
                "required_before_enablement": [
                    "upper_layer_confirmation_flow",
                    "result_materialization_policy",
                    "independent_audit_records",
                    "negative_regression_suite",
                ],
                "blocked_until": "upper-layer confirmation flow exists for high-risk actions, replay, and any business-truth write candidate",
            },
        ],
        "blocked_capabilities": sorted(REQUIRED_BLOCKED_CAPABILITIES),
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_readiness_summary",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_summary_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "readiness summary schema_version must be 1")
    require(document.get("kind") == "session_tooling_readiness_summary", "readiness summary kind mismatch")
    require(document.get("stage") == "P2 Session & Tooling Foundation", "readiness summary stage mismatch")
    require(
        document.get("status") == "metadata_ready_implementation_blocked",
        "readiness summary must keep implementation blocked status",
    )

    completed_gates = document.get("completed_gates")
    require(isinstance(completed_gates, list) and completed_gates, "readiness summary must include completed_gates")
    completed_gate_ids = {
        str(gate.get("gate_id") or "").strip()
        for gate in completed_gates
        if isinstance(gate, dict)
    }
    missing_gate_ids = sorted(REQUIRED_COMPLETED_GATE_IDS - completed_gate_ids)
    require(not missing_gate_ids, f"readiness summary missing completed gates: {missing_gate_ids}")

    preconditions = document.get("missing_preconditions")
    require(isinstance(preconditions, list) and preconditions, "readiness summary must include missing_preconditions")
    precondition_areas = {
        str(item.get("area") or "").strip()
        for item in preconditions
        if isinstance(item, dict)
    }
    missing_areas = sorted(REQUIRED_PRECONDITION_AREAS - precondition_areas)
    require(not missing_areas, f"readiness summary missing precondition areas: {missing_areas}")

    blockers: set[str] = set()
    for item in preconditions:
        require(isinstance(item, dict), "missing precondition must be object")
        require(item.get("status") == "not_ready", "missing precondition status must stay not_ready")
        blockers.update(
            require_string_list(
                item.get("required_before_enablement"),
                "missing precondition must include required_before_enablement",
            )
        )
    missing_blockers = sorted(REQUIRED_FUTURE_BLOCKERS - blockers)
    require(not missing_blockers, f"readiness summary missing future blockers: {missing_blockers}")

    blocked_capabilities = set(require_string_list(document.get("blocked_capabilities"), "blocked_capabilities required"))
    missing_capabilities = sorted(REQUIRED_BLOCKED_CAPABILITIES - blocked_capabilities)
    require(not missing_capabilities, f"readiness summary missing blocked capabilities: {missing_capabilities}")


def check_consumers_and_docs() -> None:
    check_repo_content = CHECK_REPO.read_text(encoding="utf-8")
    contracts_readme = CONTRACTS_README.read_text(encoding="utf-8")
    current_focus = CURRENT_FOCUS.read_text(encoding="utf-8")

    require(
        "check-session-tooling-readiness-summary.py" in check_repo_content,
        "fast baseline must run readiness summary check",
    )
    require(
        READINESS_SUMMARY.name in contracts_readme,
        "contracts/README.md must reference readiness summary fixture",
    )
    require(
        READINESS_SUMMARY.name in current_focus,
        "current focus must reference readiness summary fixture",
    )


def main() -> int:
    expected_summary = require_object(load_json_document(READINESS_SUMMARY), "readiness summary must be object")
    check_summary_shape(expected_summary)
    actual_summary = build_summary()
    if actual_summary != expected_summary:
        raise SystemExit("session/tooling readiness summary does not match regenerated output")
    check_consumers_and_docs()
    print("session/tooling readiness summary checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
