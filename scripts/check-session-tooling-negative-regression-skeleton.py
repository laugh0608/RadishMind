#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-negative-regression-skeleton.json"
PRECONDITIONS_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-implementation-preconditions.json"
DENIED_QUERY_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-read-denied-queries.json"
TASK_CARD_PATH = REPO_ROOT / "docs/task-cards/session-tooling-implementation-preconditions.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_GROUPS = {
    "executor_blocked": "executor",
    "storage_materialization_blocked": "storage",
    "confirmation_blocked": "confirmation",
}
REQUIRED_CASE_IDS = {
    "executor-disabled-tool-run",
    "executor-network-disabled",
    "executor-ref-checkpoint-read-denied",
    "materialized-result-read-denied",
    "durable-memory-write-denied",
    "business-truth-write-denied",
    "missing-confirmation-denied",
    "stale-confirmation-denied",
    "mismatched-confirmation-payload-denied",
}
REQUIRED_FAILURE_BOUNDARIES = {"policy_gate", "northbound_request", "confirmation_gate"}
REQUIRED_ERROR_CODES = {
    "BUSINESS_TRUTH_WRITE_DISABLED",
    "CHECKPOINT_DURABLE_MEMORY_DISABLED",
    "CHECKPOINT_MATERIALIZED_RESULTS_DISABLED",
    "CONFIRMATION_PAYLOAD_MISMATCH",
    "CONFIRMATION_REQUIRED",
    "CONFIRMATION_STALE",
    "TOOL_EXECUTOR_DISABLED",
    "TOOL_NETWORK_DISABLED",
}
REQUIRED_BLOCKED_CAPABILITIES = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_memory",
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


def check_skeleton_fixture(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "negative regression skeleton schema_version must be 1")
    require(
        document.get("kind") == "session_tooling_negative_regression_skeleton",
        "negative regression skeleton kind mismatch",
    )
    require(document.get("stage") == "P2 Session & Tooling Foundation", "negative regression skeleton stage mismatch")
    require(
        document.get("status") == "skeleton_only_implementation_blocked",
        "negative regression skeleton must not claim implementation readiness",
    )
    require(
        document.get("source_preconditions") == relative_path(PRECONDITIONS_PATH),
        "negative regression skeleton must reference implementation preconditions",
    )
    require(
        document.get("source_denied_query_fixture") == relative_path(DENIED_QUERY_FIXTURE),
        "negative regression skeleton must reference denied query fixture",
    )

    groups = document.get("groups")
    require(isinstance(groups, list) and groups, "negative regression skeleton must include groups")
    groups_by_id: dict[str, dict[str, Any]] = {}
    case_ids: set[str] = set()
    failure_boundaries: set[str] = set()
    error_codes: set[str] = set()
    mirrored_denied_categories: set[str] = set()

    for group in groups:
        group_obj = require_object(group, "negative regression group must be object")
        group_id = str(group_obj.get("group_id") or "").strip()
        require(group_id in REQUIRED_GROUPS, f"unexpected negative regression group: {group_id}")
        require(group_id not in groups_by_id, f"duplicate negative regression group: {group_id}")
        groups_by_id[group_id] = group_obj
        require(group_obj.get("status") == "skeleton_only", f"{group_id} must stay skeleton_only")
        require(
            group_obj.get("precondition_area") == REQUIRED_GROUPS[group_id],
            f"{group_id} precondition area mismatch",
        )

        cases = group_obj.get("cases")
        require(isinstance(cases, list) and cases, f"{group_id} must include cases")
        for case in cases:
            case_obj = require_object(case, f"{group_id} case must be object")
            case_id = str(case_obj.get("case_id") or "").strip()
            require(case_id, f"{group_id} case must include case_id")
            require(case_id not in case_ids, f"duplicate negative regression case: {case_id}")
            case_ids.add(case_id)
            require(str(case_obj.get("intent") or "").strip(), f"{case_id} must include intent")
            failure_boundaries.add(str(case_obj.get("expected_failure_boundary") or "").strip())
            error_codes.add(str(case_obj.get("expected_error_code") or "").strip())
            forbidden_outputs = set(require_string_list(case_obj.get("forbidden_outputs"), f"{case_id} must include forbidden_outputs"))
            require(
                any("executed" in item or "business_truth" in item or "result_ref" in item for item in forbidden_outputs),
                f"{case_id} must forbid execution, business truth writes, or result refs",
            )
            mirrored = str(case_obj.get("mirrors_denied_query_category") or "").strip()
            if mirrored:
                mirrored_denied_categories.add(mirrored)

    missing_groups = sorted(set(REQUIRED_GROUPS) - set(groups_by_id))
    require(not missing_groups, f"negative regression skeleton missing groups: {missing_groups}")
    missing_cases = sorted(REQUIRED_CASE_IDS - case_ids)
    require(not missing_cases, f"negative regression skeleton missing cases: {missing_cases}")
    missing_boundaries = sorted(REQUIRED_FAILURE_BOUNDARIES - failure_boundaries)
    require(not missing_boundaries, f"negative regression skeleton missing boundaries: {missing_boundaries}")
    missing_error_codes = sorted(REQUIRED_ERROR_CODES - error_codes)
    require(not missing_error_codes, f"negative regression skeleton missing error codes: {missing_error_codes}")
    require(
        {"materialized_results", "durable_memory"}.issubset(mirrored_denied_categories),
        "negative regression skeleton must mirror committed checkpoint denied query categories",
    )

    blocked_capabilities = set(
        require_string_list(document.get("blocked_capabilities"), "negative regression skeleton must include blocked capabilities")
    )
    missing_blocked = sorted(REQUIRED_BLOCKED_CAPABILITIES - blocked_capabilities)
    require(not missing_blocked, f"negative regression skeleton missing blocked capabilities: {missing_blocked}")


def check_precondition_alignment(skeleton: dict[str, Any], preconditions: dict[str, Any]) -> None:
    precondition_areas = {
        str(area.get("area") or "").strip()
        for area in preconditions.get("areas") or []
        if isinstance(area, dict) and area.get("status") == "not_ready"
    }
    skeleton_areas = {
        str(group.get("precondition_area") or "").strip()
        for group in skeleton.get("groups") or []
        if isinstance(group, dict)
    }
    require(skeleton_areas == set(REQUIRED_GROUPS.values()), "negative regression skeleton areas mismatch")
    missing_from_preconditions = sorted(skeleton_areas - precondition_areas)
    require(
        not missing_from_preconditions,
        f"negative regression skeleton references areas missing from preconditions: {missing_from_preconditions}",
    )


def check_denied_query_alignment(skeleton: dict[str, Any], denied_queries: dict[str, Any]) -> None:
    cases = denied_queries.get("cases")
    require(isinstance(cases, list) and cases, "denied query fixture must include cases")
    denied_categories = {
        str(case.get("category") or "").strip()
        for case in cases
        if isinstance(case, dict)
    }
    denied_error_codes = {
        str(case.get("expected_error_code") or "").strip()
        for case in cases
        if isinstance(case, dict)
    }

    mirrored_categories: set[str] = set()
    mirrored_error_codes: set[str] = set()
    for group in skeleton.get("groups") or []:
        if not isinstance(group, dict):
            continue
        for case in group.get("cases") or []:
            if not isinstance(case, dict):
                continue
            category = str(case.get("mirrors_denied_query_category") or "").strip()
            if not category:
                continue
            mirrored_categories.add(category)
            mirrored_error_codes.add(str(case.get("expected_error_code") or "").strip())

    missing_categories = sorted(mirrored_categories - denied_categories)
    require(not missing_categories, f"negative regression skeleton mirrors unknown denied categories: {missing_categories}")
    missing_error_codes = sorted(mirrored_error_codes - denied_error_codes)
    require(not missing_error_codes, f"negative regression skeleton mirrors unknown denied error codes: {missing_error_codes}")


def check_docs_and_consumers() -> None:
    task_card = TASK_CARD_PATH.read_text(encoding="utf-8")
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name
    for marker in (
        fixture_name,
        "负向回归 skeleton",
        "blocked executor",
        "blocked storage",
        "blocked confirmation",
        "不代表 negative_regression_suite 已满足",
    ):
        require(marker in task_card, f"implementation preconditions task card missing marker: {marker}")
    require(
        "check-session-tooling-negative-regression-skeleton.py" in check_repo,
        "fast baseline must run negative regression skeleton check",
    )


def main() -> int:
    skeleton = require_object(load_json_document(FIXTURE_PATH), "negative regression skeleton must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS_PATH), "preconditions fixture must be object")
    denied_queries = require_object(load_json_document(DENIED_QUERY_FIXTURE), "denied query fixture must be object")
    check_skeleton_fixture(skeleton)
    check_precondition_alignment(skeleton, preconditions)
    check_denied_query_alignment(skeleton, denied_queries)
    check_docs_and_consumers()
    print("session/tooling negative regression skeleton checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
