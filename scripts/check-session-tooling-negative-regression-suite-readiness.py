#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

FIXTURE_PATH = FIXTURE_DIR / "session-tooling-negative-regression-suite-readiness.json"
SUITE_PATH = FIXTURE_DIR / "session-tooling-negative-regression-suite.json"
SKELETON_PATH = FIXTURE_DIR / "session-tooling-negative-regression-skeleton.json"
PRECONDITIONS_PATH = FIXTURE_DIR / "session-tooling-implementation-preconditions.json"
DENIED_QUERY_FIXTURE = FIXTURE_DIR / "session-recovery-checkpoint-read-denied-queries.json"
CLOSE_CANDIDATE_ROLLUP = FIXTURE_DIR / "session-tooling-close-candidate-readiness-rollup.json"

TASK_CARD_PATH = REPO_ROOT / "docs/task-cards/session-tooling-negative-regression-suite-readiness.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-negative-regression-suite-readiness.py"

REQUIRED_GROUPS = {
    "executor_blocked": "executor",
    "storage_materialization_blocked": "storage",
    "confirmation_blocked": "confirmation",
}
REQUIRED_REQUIREMENTS = {
    "real_consumers_before_completion",
    "independent_audit_assertions",
    "side_effect_absence_assertions",
    "checkpoint_denied_query_alignment",
    "implementation_gate_alignment",
}
REQUIRED_BLOCKERS = {
    "implementation_gates_missing",
    "executor_storage_confirmation_still_not_ready",
    "upper_layer_confirmation_flow_missing",
    "durable_store_and_result_reader_still_disabled",
}
REQUIRED_NOT_ENABLED = {
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
    "real_tool_executor",
    "result_ref_reader",
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


def skeleton_groups_by_id(skeleton: dict[str, Any]) -> dict[str, dict[str, Any]]:
    groups: dict[str, dict[str, Any]] = {}
    for item in skeleton.get("groups") or []:
        group = require_object(item, "negative regression skeleton group must be object")
        group_id = str(group.get("group_id") or "").strip()
        require(group_id in REQUIRED_GROUPS, f"unexpected skeleton group: {group_id}")
        groups[group_id] = group
    require(set(groups) == set(REQUIRED_GROUPS), "skeleton must include executor/storage/confirmation groups")
    return groups


def case_ids_from_group(group: dict[str, Any]) -> list[str]:
    cases = group.get("cases")
    require(isinstance(cases, list) and cases, "skeleton group must include cases")
    return [str(case.get("case_id") or "").strip() for case in cases if isinstance(case, dict)]


def skeleton_coverage(skeleton: dict[str, Any]) -> dict[str, Any]:
    groups = skeleton_groups_by_id(skeleton)
    failure_boundaries: set[str] = set()
    error_codes: set[str] = set()
    case_count = 0
    for group in groups.values():
        for case in group.get("cases") or []:
            case_obj = require_object(case, "negative regression case must be object")
            case_count += 1
            failure_boundaries.add(str(case_obj.get("expected_failure_boundary") or "").strip())
            error_codes.add(str(case_obj.get("expected_error_code") or "").strip())
    return {
        "group_count": len(groups),
        "case_count": case_count,
        "failure_boundaries": sorted(failure_boundaries),
        "error_codes": sorted(error_codes),
    }


def minimum_suite_groups(skeleton: dict[str, Any], suite: dict[str, Any]) -> list[dict[str, Any]]:
    groups = skeleton_groups_by_id(skeleton)
    suite_groups = {
        str(group.get("group_id") or "").strip(): group
        for group in suite.get("groups") or []
        if isinstance(group, dict)
    }
    return [
        {
            "group_id": "executor_blocked",
            "precondition_area": "executor",
            "source_skeleton_group": "executor_blocked",
            "required_case_ids": case_ids_from_group(groups["executor_blocked"]),
            "required_suite_coverage": [
                "policy_gate_consumer",
                "tool_registry_policy_consumer",
                "independent_audit_expectation",
                "no_execution_side_effects",
                "no_network_side_effects",
            ],
            "current_status": str(suite_groups["executor_blocked"].get("current_status") or "").strip(),
        },
        {
            "group_id": "storage_materialization_blocked",
            "precondition_area": "storage",
            "source_skeleton_group": "storage_materialization_blocked",
            "required_case_ids": case_ids_from_group(groups["storage_materialization_blocked"]),
            "required_suite_coverage": [
                "checkpoint_read_route_consumer",
                "storage_policy_gate_consumer",
                "independent_audit_expectation",
                "no_materialized_result_side_effects",
                "no_business_truth_write_side_effects",
            ],
            "current_status": str(suite_groups["storage_materialization_blocked"].get("current_status") or "").strip(),
        },
        {
            "group_id": "confirmation_blocked",
            "precondition_area": "confirmation",
            "source_skeleton_group": "confirmation_blocked",
            "required_case_ids": case_ids_from_group(groups["confirmation_blocked"]),
            "required_suite_coverage": [
                "confirmation_gate_consumer",
                "action_payload_hash_consumer",
                "independent_audit_expectation",
                "no_confirmed_action_execution",
                "no_replay_side_effects",
            ],
            "current_status": str(suite_groups["confirmation_blocked"].get("current_status") or "").strip(),
        },
    ]


def build_readiness() -> dict[str, Any]:
    skeleton = require_object(load_json_document(SKELETON_PATH), "negative regression skeleton must be object")
    suite = require_object(load_json_document(SUITE_PATH), "negative regression suite must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS_PATH), "preconditions fixture must be object")
    denied_queries = require_object(load_json_document(DENIED_QUERY_FIXTURE), "denied query fixture must be object")
    close_candidate = require_object(load_json_document(CLOSE_CANDIDATE_ROLLUP), "close candidate rollup must be object")

    require(skeleton.get("status") == "skeleton_only_implementation_blocked", "skeleton must remain implementation blocked")
    require(
        suite.get("status") == "governance_suite_consumed_implementation_gates_missing",
        "suite must remain governance-only with missing implementation gates",
    )
    require(preconditions.get("status") == "preconditions_not_satisfied", "preconditions must remain unsatisfied")
    require(close_candidate.get("status") == "close_candidate_governance_only", "rollup must remain governance-only")
    require(
        "complete_negative_regression_suite" in close_candidate.get("prohibited_claims", []),
        "rollup must still prohibit complete negative regression suite claim",
    )
    denied_cases = denied_queries.get("cases")
    require(isinstance(denied_cases, list) and denied_cases, "denied query fixture must include cases")

    return {
        "schema_version": 1,
        "kind": "session_tooling_negative_regression_suite_readiness",
        "stage": "P2 Session & Tooling Foundation",
        "status": "governance_suite_consumed_implementation_gates_missing",
        "implementation_status": "not_ready",
        "source_negative_regression_suite": relative_path(SUITE_PATH),
        "source_negative_regression_skeleton": relative_path(SKELETON_PATH),
        "source_implementation_preconditions": relative_path(PRECONDITIONS_PATH),
        "source_denied_query_fixture": relative_path(DENIED_QUERY_FIXTURE),
        "source_close_candidate_rollup": relative_path(CLOSE_CANDIDATE_ROLLUP),
        "minimum_suite_groups": minimum_suite_groups(skeleton, suite),
        "suite_acceptance_requirements": [
            {
                "requirement_id": "real_consumers_before_completion",
                "status": "partially_satisfied_by_governance_suite",
                "description": "Each negative case now has governance consumers, but implementation gates are still missing",
            },
            {
                "requirement_id": "independent_audit_assertions",
                "status": "satisfied_by_audit_non_write_boundary",
                "description": "Each blocked action asserts the expected event source and explicit audit non-write boundary",
            },
            {
                "requirement_id": "side_effect_absence_assertions",
                "status": "satisfied_by_forbidden_side_effect_assertions",
                "description": "Each case asserts absence of execution, storage, materialized result, business truth write, or replay side effects",
            },
            {
                "requirement_id": "checkpoint_denied_query_alignment",
                "status": "satisfied_by_denied_query_fixture_and_suite_cases",
                "description": "Committed checkpoint denied query cases are consumed by route/contract checks and mirrored by suite cases",
            },
            {
                "requirement_id": "implementation_gate_alignment",
                "status": "not_satisfied",
                "description": "The suite cannot be complete until executor, storage, and confirmation implementation gates exist and remain deny-by-default",
            },
        ],
        "current_skeleton_coverage": skeleton_coverage(skeleton),
        "blocked_completion_reasons": [
            "implementation_gates_missing",
            "executor_storage_confirmation_still_not_ready",
            "upper_layer_confirmation_flow_missing",
            "durable_store_and_result_reader_still_disabled",
        ],
        "not_enabled_capabilities": sorted(REQUIRED_NOT_ENABLED),
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "validates_negative_regression_suite_acceptance_without_marking_suite_complete",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_fixture_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "suite readiness schema_version must be 1")
    require(document.get("kind") == "session_tooling_negative_regression_suite_readiness", "suite readiness kind mismatch")
    require(
        document.get("status") == "governance_suite_consumed_implementation_gates_missing",
        "suite readiness must not claim completion",
    )
    require(document.get("implementation_status") == "not_ready", "suite readiness must keep implementation not_ready")

    groups = document.get("minimum_suite_groups")
    require(isinstance(groups, list) and groups, "suite readiness must include minimum groups")
    group_ids = {str(group.get("group_id") or "").strip() for group in groups if isinstance(group, dict)}
    missing_groups = sorted(set(REQUIRED_GROUPS) - group_ids)
    require(not missing_groups, f"suite readiness missing groups: {missing_groups}")
    for group in groups:
        group_obj = require_object(group, "suite readiness group must be object")
        group_id = str(group_obj.get("group_id") or "").strip()
        require(group_obj.get("precondition_area") == REQUIRED_GROUPS[group_id], f"{group_id} precondition area mismatch")
        require(
            group_obj.get("current_status") == "governance_consumed_implementation_blocked",
            f"{group_id} must remain implementation blocked",
        )
        require_string_list(group_obj.get("required_case_ids"), f"{group_id} must include required cases")
        coverage = set(require_string_list(group_obj.get("required_suite_coverage"), f"{group_id} must include suite coverage"))
        require("independent_audit_expectation" in coverage, f"{group_id} must require independent audit expectation")

    requirements = document.get("suite_acceptance_requirements")
    require(isinstance(requirements, list) and requirements, "suite readiness must include acceptance requirements")
    requirement_ids = {str(item.get("requirement_id") or "").strip() for item in requirements if isinstance(item, dict)}
    missing_requirements = sorted(REQUIRED_REQUIREMENTS - requirement_ids)
    require(not missing_requirements, f"suite readiness missing requirements: {missing_requirements}")
    statuses_by_requirement = {
        str(item.get("requirement_id") or "").strip(): str(item.get("status") or "").strip()
        for item in requirements
        if isinstance(item, dict)
    }
    require(
        statuses_by_requirement.get("implementation_gate_alignment") == "not_satisfied",
        "suite readiness must keep implementation gate alignment unsatisfied",
    )

    blockers = set(require_string_list(document.get("blocked_completion_reasons"), "suite readiness must include blockers"))
    missing_blockers = sorted(REQUIRED_BLOCKERS - blockers)
    require(not missing_blockers, f"suite readiness missing blockers: {missing_blockers}")

    not_enabled = set(require_string_list(document.get("not_enabled_capabilities"), "suite readiness must include blocked capabilities"))
    missing_not_enabled = sorted(REQUIRED_NOT_ENABLED - not_enabled)
    require(not missing_not_enabled, f"suite readiness missing blocked capabilities: {missing_not_enabled}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD_PATH.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    contracts_readme = CONTRACTS_README.read_text(encoding="utf-8")
    current_focus = CURRENT_FOCUS.read_text(encoding="utf-8")
    capability_matrix = CAPABILITY_MATRIX.read_text(encoding="utf-8")
    roadmap = ROADMAP.read_text(encoding="utf-8")
    devlog = DEVLOG.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name

    require(THIS_CHECK.name in check_repo, "fast baseline must run negative regression suite readiness check")
    require(TASK_CARD_PATH.name in task_cards_readme, "task-cards README must reference suite readiness task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", contracts_readme),
        ("current focus", current_focus),
        ("capability matrix", capability_matrix),
        ("roadmap", roadmap),
        ("devlog", devlog),
    ):
        require(fixture_name in content, f"{label} must reference suite readiness fixture")
        require("negative_regression_suite" in content, f"{label} must mention negative_regression_suite")
    for marker in ("not_satisfied", "不实现真实工具执行器", "不启用 automatic replay"):
        require(marker in task_card, f"task card missing marker: {marker}")


def main() -> int:
    expected = require_object(load_json_document(FIXTURE_PATH), "suite readiness fixture must be object")
    check_fixture_shape(expected)
    actual = build_readiness()
    if actual != expected:
        raise SystemExit("session/tooling negative regression suite readiness does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling negative regression suite readiness checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
