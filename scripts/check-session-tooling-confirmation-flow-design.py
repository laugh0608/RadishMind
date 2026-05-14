#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-confirmation-flow-design.json"
PRECONDITIONS_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-implementation-preconditions.json"
NEGATIVE_SKELETON_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-negative-regression-skeleton.json"
TASK_CARD_PATH = REPO_ROOT / "docs/task-cards/session-tooling-confirmation-flow-design.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_OUTCOMES = {"approve", "reject", "defer"}
REQUIRED_FIELDS = {
    "confirmation_id",
    "session_id",
    "turn_id",
    "action_ref",
    "action_hash",
    "outcome",
    "actor_ref",
    "confirmed_at",
    "audit_ref",
}
REQUIRED_NEGATIVE_CASES = {
    "missing-confirmation": "CONFIRMATION_REQUIRED",
    "stale-confirmation": "CONFIRMATION_STALE",
    "mismatched-confirmation-payload": "CONFIRMATION_PAYLOAD_MISMATCH",
}
REQUIRED_AUDIT_EVENTS = {
    "confirmation_requested",
    "confirmation_recorded",
    "confirmation_rejected",
    "confirmation_deferred",
    "confirmation_invalidated",
}
REQUIRED_NOT_ENABLED = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_memory",
    "real_tool_executor",
    "result_materialization",
}
REQUIRED_UNSATISFIED_PRECONDITIONS = {
    "upper_layer_confirmation_flow",
    "independent_audit_records",
    "negative_regression_suite",
    "result_materialization_policy",
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


def check_fixture(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "confirmation flow design schema_version must be 1")
    require(
        document.get("kind") == "session_tooling_confirmation_flow_design",
        "confirmation flow design kind mismatch",
    )
    require(document.get("stage") == "P2 Session & Tooling Foundation", "confirmation flow stage mismatch")
    require(document.get("status") == "design_only_not_connected", "confirmation flow must stay design-only")
    require(
        document.get("source_preconditions") == relative_path(PRECONDITIONS_PATH),
        "confirmation flow design must reference implementation preconditions",
    )
    require(document.get("task_card") == relative_path(TASK_CARD_PATH), "confirmation flow design must reference task card")

    scope = require_object(document.get("scope"), "confirmation flow design must include scope")
    require(scope.get("area") == "confirmation", "confirmation flow scope area must be confirmation")
    require(scope.get("upper_layer_integration") == "not_connected", "upper layer integration must stay not connected")
    require(scope.get("execution_enabled") is False, "confirmation flow design must not enable execution")
    require(scope.get("writes_business_truth") is False, "confirmation flow design must not write business truth")

    record_shape = require_object(document.get("confirmation_record_shape"), "confirmation flow must include record shape")
    fields = set(require_string_list(record_shape.get("required_fields"), "confirmation record required_fields missing"))
    missing_fields = sorted(REQUIRED_FIELDS - fields)
    require(not missing_fields, f"confirmation record missing fields: {missing_fields}")

    outcomes = record_shape.get("outcomes")
    require(isinstance(outcomes, list) and outcomes, "confirmation record must include outcomes")
    seen_outcomes: set[str] = set()
    for outcome in outcomes:
        outcome_obj = require_object(outcome, "confirmation outcome must be object")
        outcome_id = str(outcome_obj.get("outcome") or "").strip()
        require(outcome_id in REQUIRED_OUTCOMES, f"unexpected confirmation outcome: {outcome_id}")
        require(outcome_id not in seen_outcomes, f"duplicate confirmation outcome: {outcome_id}")
        seen_outcomes.add(outcome_id)
        forbidden = set(require_string_list(outcome_obj.get("forbidden_side_effects"), f"{outcome_id} forbidden side effects missing"))
        require(
            "writes_business_truth=true" in forbidden and "execution_status=executed" in forbidden,
            f"{outcome_id} must forbid execution and business truth writes",
        )
    missing_outcomes = sorted(REQUIRED_OUTCOMES - seen_outcomes)
    require(not missing_outcomes, f"confirmation flow missing outcomes: {missing_outcomes}")

    cases = document.get("negative_cases")
    require(isinstance(cases, list) and cases, "confirmation flow must include negative cases")
    cases_by_id: dict[str, dict[str, Any]] = {}
    for case in cases:
        case_obj = require_object(case, "negative case must be object")
        case_id = str(case_obj.get("case_id") or "").strip()
        require(case_id in REQUIRED_NEGATIVE_CASES, f"unexpected confirmation negative case: {case_id}")
        require(case_id not in cases_by_id, f"duplicate confirmation negative case: {case_id}")
        cases_by_id[case_id] = case_obj
        require(case_obj.get("expected_failure_boundary") == "confirmation_gate", f"{case_id} boundary mismatch")
        require(
            case_obj.get("expected_error_code") == REQUIRED_NEGATIVE_CASES[case_id],
            f"{case_id} error code mismatch",
        )
        forbidden = set(require_string_list(case_obj.get("forbidden_outputs"), f"{case_id} forbidden outputs missing"))
        require(
            any(item in forbidden for item in ("execution_status=executed", "confirmed_action_execution", "writes_business_truth=true")),
            f"{case_id} must forbid execution or business truth writes",
        )
    missing_cases = sorted(set(REQUIRED_NEGATIVE_CASES) - set(cases_by_id))
    require(not missing_cases, f"confirmation flow missing negative cases: {missing_cases}")

    audit_events = set(require_string_list(document.get("audit_events"), "confirmation flow must include audit events"))
    missing_events = sorted(REQUIRED_AUDIT_EVENTS - audit_events)
    require(not missing_events, f"confirmation flow missing audit events: {missing_events}")

    not_enabled = set(require_string_list(document.get("not_enabled_capabilities"), "confirmation flow must include not enabled capabilities"))
    missing_not_enabled = sorted(REQUIRED_NOT_ENABLED - not_enabled)
    require(not missing_not_enabled, f"confirmation flow missing blocked capabilities: {missing_not_enabled}")

    precondition_status = require_object(
        document.get("precondition_status_after_design"),
        "confirmation flow must include post-design precondition status",
    )
    for key in REQUIRED_UNSATISFIED_PRECONDITIONS:
        require(precondition_status.get(key) == "not_satisfied", f"{key} must remain not_satisfied")


def check_preconditions_alignment(preconditions: dict[str, Any]) -> None:
    confirmation_area = next(
        (
            area
            for area in preconditions.get("areas") or []
            if isinstance(area, dict) and area.get("area") == "confirmation"
        ),
        None,
    )
    confirmation = require_object(confirmation_area, "implementation preconditions must include confirmation area")
    require(confirmation.get("status") == "not_ready", "confirmation precondition must remain not_ready")
    blockers = set(require_string_list(confirmation.get("required_before_enablement"), "confirmation blockers missing"))
    missing_blockers = sorted(REQUIRED_UNSATISFIED_PRECONDITIONS - blockers)
    require(not missing_blockers, f"confirmation preconditions missing blockers: {missing_blockers}")


def check_negative_skeleton_alignment(negative_skeleton: dict[str, Any]) -> None:
    confirmation_group = next(
        (
            group
            for group in negative_skeleton.get("groups") or []
            if isinstance(group, dict) and group.get("precondition_area") == "confirmation"
        ),
        None,
    )
    group = require_object(confirmation_group, "negative skeleton must include confirmation group")
    cases = group.get("cases")
    require(isinstance(cases, list) and cases, "confirmation negative skeleton group must include cases")
    error_codes = {
        str(case.get("expected_error_code") or "").strip()
        for case in cases
        if isinstance(case, dict)
    }
    missing_error_codes = sorted(set(REQUIRED_NEGATIVE_CASES.values()) - error_codes)
    require(not missing_error_codes, f"confirmation design missing skeleton error codes: {missing_error_codes}")


def check_docs_and_consumers() -> None:
    task_card = TASK_CARD_PATH.read_text(encoding="utf-8")
    readme = TASK_CARDS_README.read_text(encoding="utf-8")
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name
    for marker in (
        fixture_name,
        "design_only_not_connected",
        "approve",
        "reject",
        "defer",
        "CONFIRMATION_REQUIRED",
        "CONFIRMATION_STALE",
        "CONFIRMATION_PAYLOAD_MISMATCH",
        "不执行 confirmed action",
        "不写 `RadishFlow`",
    ):
        require(marker in task_card, f"confirmation task card missing marker: {marker}")
    require(
        "session-tooling-confirmation-flow-design.md" in readme,
        "task-cards README must reference confirmation flow design task card",
    )
    require(
        "check-session-tooling-confirmation-flow-design.py" in check_repo,
        "fast baseline must run confirmation flow design check",
    )


def main() -> int:
    document = require_object(load_json_document(FIXTURE_PATH), "confirmation flow fixture must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS_PATH), "preconditions fixture must be object")
    negative_skeleton = require_object(load_json_document(NEGATIVE_SKELETON_PATH), "negative skeleton fixture must be object")
    check_fixture(document)
    check_preconditions_alignment(preconditions)
    check_negative_skeleton_alignment(negative_skeleton)
    check_docs_and_consumers()
    print("session/tooling confirmation flow design checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
