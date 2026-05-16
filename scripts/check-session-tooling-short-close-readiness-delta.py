#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

DELTA = FIXTURE_DIR / "session-tooling-short-close-readiness-delta.json"
CLOSE_CANDIDATE_ROLLUP = FIXTURE_DIR / "session-tooling-close-candidate-readiness-rollup.json"
FOUNDATION_STATUS = FIXTURE_DIR / "session-tooling-foundation-status-summary.json"
NEGATIVE_SUITE_READINESS = FIXTURE_DIR / "session-tooling-negative-regression-suite-readiness.json"
NEGATIVE_COVERAGE_ROLLUP = FIXTURE_DIR / "session-tooling-negative-coverage-rollup.json"
PRECONDITIONS = FIXTURE_DIR / "session-tooling-implementation-preconditions.json"
ENABLEMENT_PLAN = FIXTURE_DIR / "session-tooling-executor-storage-confirmation-enablement-plan.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-short-close-readiness-delta.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
DOCS_README = REPO_ROOT / "docs/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-short-close-readiness-delta.py"

REQUIRED_PREREQUISITES = {
    "upper_layer_confirmation_flow",
    "complete_negative_regression_suite",
    "executor_storage_confirmation_enablement_plan",
    "durable_store_and_result_reader_policy",
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
REQUIRED_NOT_IMPLEMENT_NOW = {
    "real_tool_executor",
    "durable_session_store",
    "durable_checkpoint_store",
    "durable_audit_store",
    "durable_result_store",
    "materialized_result_reader",
    "upper_layer_confirmation_flow_connection",
    "long_term_memory",
    "automatic_replay",
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


def close_prerequisites_by_id(close_candidate: dict[str, Any]) -> dict[str, dict[str, Any]]:
    prerequisites: dict[str, dict[str, Any]] = {}
    for item in close_candidate.get("short_close_prerequisites") or []:
        prerequisite = require_object(item, "short close prerequisite must be object")
        prerequisite_id = str(prerequisite.get("condition") or "").strip()
        require(prerequisite_id in REQUIRED_PREREQUISITES, f"unexpected short close prerequisite: {prerequisite_id}")
        require(prerequisite.get("status") == "not_satisfied", f"{prerequisite_id} must stay not_satisfied")
        prerequisites[prerequisite_id] = prerequisite
    missing = sorted(REQUIRED_PREREQUISITES - set(prerequisites))
    require(not missing, f"close candidate rollup missing short close prerequisites: {missing}")
    return prerequisites


def precondition_statuses(preconditions: dict[str, Any]) -> dict[str, str]:
    statuses: dict[str, str] = {}
    for item in preconditions.get("areas") or []:
        area = require_object(item, "implementation precondition area must be object")
        area_id = str(area.get("area") or "").strip()
        statuses[area_id] = str(area.get("status") or "").strip()
    require(statuses == {"executor": "not_ready", "storage": "not_ready", "confirmation": "not_ready"}, "executor/storage/confirmation must remain not_ready")
    return statuses


def build_delta() -> dict[str, Any]:
    close_candidate = require_object(load_json_document(CLOSE_CANDIDATE_ROLLUP), "close candidate rollup must be object")
    foundation = require_object(load_json_document(FOUNDATION_STATUS), "foundation status summary must be object")
    suite_readiness = require_object(load_json_document(NEGATIVE_SUITE_READINESS), "suite readiness must be object")
    coverage_rollup = require_object(load_json_document(NEGATIVE_COVERAGE_ROLLUP), "negative coverage rollup must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS), "implementation preconditions must be object")
    enablement_plan = require_object(load_json_document(ENABLEMENT_PLAN), "enablement plan must be object")

    require(close_candidate.get("status") == "close_candidate_governance_only", "close candidate must stay governance-only")
    require(close_candidate.get("implementation_status") == "not_ready", "close candidate must not claim implementation readiness")
    require(foundation.get("status") == "close_candidate_governance_only", "foundation status must stay governance-only")
    require(foundation.get("implementation_status") == "not_implemented", "foundation must not claim implementation")
    require(
        suite_readiness.get("status") == "governance_suite_consumed_deny_by_default_gates_defined",
        "suite readiness must stay governance-only with deny-by-default gates defined",
    )
    require(suite_readiness.get("implementation_status") == "not_ready", "suite readiness must keep implementation not_ready")
    require(
        coverage_rollup.get("status") == "negative_coverage_rollup_governance_only",
        "negative coverage rollup must stay governance-only",
    )
    require(preconditions.get("status") == "preconditions_not_satisfied", "implementation preconditions must remain unsatisfied")
    require(enablement_plan.get("status") == "enablement_plan_defined_blocked", "enablement plan must remain blocked")

    close_prerequisites = close_prerequisites_by_id(close_candidate)
    statuses = precondition_statuses(preconditions)

    hard_prerequisites = [
        {
            "prerequisite_id": "upper_layer_confirmation_flow",
            "current_status": close_prerequisites["upper_layer_confirmation_flow"]["status"],
            "current_evidence": "confirmation flow design exists, but no approve/reject/defer integration is connected",
            "required_delta": [
                "define upper-layer confirmation request/decision handoff",
                "bind confirmation decision to independent audit records",
                "prove stale/missing/mismatched confirmation remains denied before any action side effect",
            ],
            "blocks_claims": [
                "confirmation_flow_connected",
                "confirmed_action_execution",
                "p2_short_close",
            ],
        },
        {
            "prerequisite_id": "complete_negative_regression_suite",
            "current_status": close_prerequisites["complete_negative_regression_suite"]["status"],
            "current_evidence": "governance suite, suite readiness, deny-by-default gates, and negative coverage rollup exist, but real implementation consumers are missing",
            "required_delta": [
                "attach real executor/storage/confirmation consumers to every negative case",
                "prove deny-by-default gates fire before execution, storage, materialization, confirmation, business write, or replay side effects",
                "keep route smoke, suite readiness, and close-candidate rollup aligned after consumers are added",
            ],
            "blocks_claims": [
                "complete_negative_regression_suite",
                "real_tool_executor_ready",
                "durable_storage_ready",
                "p2_short_close",
            ],
        },
        {
            "prerequisite_id": "executor_storage_confirmation_enablement_plan",
            "current_status": close_prerequisites["executor_storage_confirmation_enablement_plan"]["status"],
            "current_evidence": f"enablement plan is defined but blocked; implementation preconditions remain executor={statuses['executor']}, storage={statuses['storage']}, confirmation={statuses['confirmation']}",
            "required_delta": [
                "turn executor enablement from not_ready to gated plan only after confirmation and negative suite prerequisites are satisfied",
                "turn storage enablement from not_ready to gated plan only after durable store and result reader policy is satisfied",
                "turn confirmation enablement from not_ready to gated plan only after upper-layer confirmation flow is connected",
            ],
            "blocks_claims": [
                "real_tool_executor_ready",
                "durable_storage_ready",
                "confirmation_flow_connected",
                "p2_short_close",
            ],
        },
        {
            "prerequisite_id": "durable_store_and_result_reader_policy",
            "current_status": close_prerequisites["durable_store_and_result_reader_policy"]["status"],
            "current_evidence": "storage backend and result materialization policy are design-checkable, but durable stores and materialized result readers are disabled",
            "required_delta": [
                "define durable session/checkpoint/audit/result store enablement boundaries",
                "define materialized result reader policy and denied access behavior",
                "prove secret handling, redaction, retention, and no business truth write side effects before enabling storage",
            ],
            "blocks_claims": [
                "durable_storage_ready",
                "materialized_result_reader_ready",
                "automatic_replay_ready",
                "p2_short_close",
            ],
        },
    ]

    return {
        "schema_version": 1,
        "kind": "session_tooling_short_close_readiness_delta",
        "stage": "P2 Session & Tooling Foundation",
        "current_state": "close_candidate_governance_only",
        "target_state": "p2_short_close",
        "status": "short_close_blocked",
        "implementation_status": "not_ready",
        "source_close_candidate_rollup": relative_path(CLOSE_CANDIDATE_ROLLUP),
        "source_foundation_status_summary": relative_path(FOUNDATION_STATUS),
        "source_negative_regression_suite_readiness": relative_path(NEGATIVE_SUITE_READINESS),
        "source_negative_coverage_rollup": relative_path(NEGATIVE_COVERAGE_ROLLUP),
        "source_implementation_preconditions": relative_path(PRECONDITIONS),
        "source_executor_storage_confirmation_enablement_plan": relative_path(ENABLEMENT_PLAN),
        "hard_prerequisites": hard_prerequisites,
        "delta_totals": {
            "hard_prerequisite_count": len(hard_prerequisites),
            "satisfied_count": 0,
            "not_satisfied_count": len(hard_prerequisites),
        },
        "current_allowed_claims": [
            "close_candidate_governance_only",
            "short_close_delta_checkable",
            "hard_prerequisites_identified",
            "implementation_still_blocked",
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
        "do_not_implement_now": sorted(REQUIRED_NOT_IMPLEMENT_NOW),
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_short_close_delta",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_delta_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "delta schema_version must be 1")
    require(document.get("kind") == "session_tooling_short_close_readiness_delta", "delta kind mismatch")
    require(document.get("current_state") == "close_candidate_governance_only", "delta must start from close candidate")
    require(document.get("target_state") == "p2_short_close", "delta target must be p2_short_close")
    require(document.get("status") == "short_close_blocked", "delta must keep short close blocked")
    require(document.get("implementation_status") == "not_ready", "delta must not claim implementation readiness")

    prerequisites = document.get("hard_prerequisites")
    require(isinstance(prerequisites, list) and prerequisites, "delta must include hard prerequisites")
    prerequisite_ids = {str(item.get("prerequisite_id") or "").strip() for item in prerequisites if isinstance(item, dict)}
    missing = sorted(REQUIRED_PREREQUISITES - prerequisite_ids)
    require(not missing, f"delta missing hard prerequisites: {missing}")
    for item in prerequisites:
        prerequisite = require_object(item, "hard prerequisite must be object")
        require(prerequisite.get("current_status") == "not_satisfied", "hard prerequisites must remain not_satisfied")
        require_string_list(prerequisite.get("required_delta"), "hard prerequisite must include required_delta")
        require_string_list(prerequisite.get("blocks_claims"), "hard prerequisite must include blocked claims")

    totals = require_object(document.get("delta_totals"), "delta must include totals")
    require(totals.get("hard_prerequisite_count") == 4, "delta must keep four hard prerequisites")
    require(totals.get("satisfied_count") == 0, "delta must not mark prerequisites satisfied")
    require(totals.get("not_satisfied_count") == 4, "delta must keep all prerequisites unsatisfied")

    prohibited = set(require_string_list(document.get("prohibited_claims"), "delta must include prohibited claims"))
    missing_prohibited = sorted(REQUIRED_PROHIBITED_CLAIMS - prohibited)
    require(not missing_prohibited, f"delta missing prohibited claims: {missing_prohibited}")

    not_implemented = set(require_string_list(document.get("do_not_implement_now"), "delta must include do_not_implement_now"))
    missing_not_implemented = sorted(REQUIRED_NOT_IMPLEMENT_NOW - not_implemented)
    require(not missing_not_implemented, f"delta missing implementation stop lines: {missing_not_implemented}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    contracts_readme = CONTRACTS_README.read_text(encoding="utf-8")
    docs_readme = DOCS_README.read_text(encoding="utf-8")
    current_focus = CURRENT_FOCUS.read_text(encoding="utf-8")
    capability_matrix = CAPABILITY_MATRIX.read_text(encoding="utf-8")
    roadmap = ROADMAP.read_text(encoding="utf-8")
    devlog = DEVLOG.read_text(encoding="utf-8")
    fixture_name = DELTA.name

    require(THIS_CHECK.name in check_repo, "fast baseline must run short close readiness delta check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference short close delta task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", contracts_readme),
        ("docs README", docs_readme),
        ("current focus", current_focus),
        ("capability matrix", capability_matrix),
        ("roadmap", roadmap),
        ("devlog", devlog),
    ):
        require(fixture_name in content, f"{label} must reference short close delta fixture")
        require("P2 short close" in content, f"{label} must mention P2 short close")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("not_satisfied" in content, f"{label} must mention not_satisfied hard prerequisites")


def main() -> int:
    expected_delta = require_object(load_json_document(DELTA), "short close delta fixture must be object")
    check_delta_shape(expected_delta)
    actual_delta = build_delta()
    if actual_delta != expected_delta:
        raise SystemExit("session/tooling short close readiness delta does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling short close readiness delta checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
