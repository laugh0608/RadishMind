#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

FIXTURE_PATH = FIXTURE_DIR / "session-tooling-deny-by-default-implementation-gates.json"
PRECONDITIONS_PATH = FIXTURE_DIR / "session-tooling-implementation-preconditions.json"
NEGATIVE_SUITE_PATH = FIXTURE_DIR / "session-tooling-negative-regression-suite.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-deny-by-default-implementation-gates.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-deny-by-default-implementation-gates.py"

REQUIRED_AREAS = {"executor", "storage", "confirmation"}
REQUIRED_STATUS = "deny_by_default_gates_defined_implementation_blocked"
GATE_STATUS = "deny_by_default_contract_defined"
SUITE_GATE_STATUS = "deny_by_default_gate_contract_defined_implementation_blocked"

REQUIRED_DENIED_ACTIONS = {
    "executor": {
        "execute_registered_tool",
        "execute_network_tool_without_policy",
        "return_executor_ref_from_checkpoint_read",
    },
    "storage": {
        "read_materialized_tool_result",
        "write_tool_state_to_durable_memory",
        "write_candidate_action_to_upper_layer_truth_source",
    },
    "confirmation": {
        "execute_high_risk_action_without_confirmation",
        "execute_action_with_stale_confirmation",
        "execute_action_with_mismatched_confirmation_payload",
    },
}

REQUIRED_DENIAL_OUTPUTS = {
    "executor": {
        "TOOL_EXECUTOR_DISABLED",
        "TOOL_NETWORK_DISABLED",
        "CHECKPOINT_MATERIALIZED_RESULTS_DISABLED",
    },
    "storage": {
        "CHECKPOINT_MATERIALIZED_RESULTS_DISABLED",
        "CHECKPOINT_DURABLE_MEMORY_DISABLED",
        "BUSINESS_TRUTH_WRITE_DISABLED",
    },
    "confirmation": {
        "CONFIRMATION_REQUIRED",
        "CONFIRMATION_STALE",
        "CONFIRMATION_PAYLOAD_MISMATCH",
    },
}

REQUIRED_SHARED_INVARIANTS = {
    "fail_closed_when_policy_or_confirmation_is_absent",
    "no_side_effect_before_explicit_enablement",
    "no_materialized_result_or_ref_from_checkpoint_read",
    "no_business_truth_write_from_model_or_tooling_layer",
    "no_automatic_replay",
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


def gates_by_area(document: dict[str, Any]) -> dict[str, dict[str, Any]]:
    gates = document.get("gates")
    require(isinstance(gates, list) and gates, "implementation gates fixture must include gates")
    result: dict[str, dict[str, Any]] = {}
    for item in gates:
        gate = require_object(item, "implementation gate must be object")
        area = str(gate.get("gate_area") or "").strip()
        require(area in REQUIRED_AREAS, f"unexpected implementation gate area: {area}")
        require(area not in result, f"duplicate implementation gate area: {area}")
        result[area] = gate
    require(set(result) == REQUIRED_AREAS, "implementation gates must cover executor/storage/confirmation")
    return result


def check_gate_fixture(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "implementation gates schema_version must be 1")
    require(
        document.get("kind") == "session_tooling_deny_by_default_implementation_gates",
        "implementation gates kind mismatch",
    )
    require(document.get("stage") == "P2 Session & Tooling Foundation", "implementation gates stage mismatch")
    require(document.get("status") == REQUIRED_STATUS, "implementation gates must stay implementation blocked")
    require(document.get("implementation_status") == "not_ready", "implementation gates must not claim readiness")
    require(
        document.get("source_implementation_preconditions") == relative_path(PRECONDITIONS_PATH),
        "implementation gates must reference implementation preconditions",
    )
    require(
        document.get("source_negative_regression_suite") == relative_path(NEGATIVE_SUITE_PATH),
        "implementation gates must reference negative regression suite",
    )
    require(document.get("task_card") == relative_path(TASK_CARD), "implementation gates must reference task card")

    for area, gate in gates_by_area(document).items():
        require(gate.get("status") == GATE_STATUS, f"{area} gate status mismatch")
        require(gate.get("default_decision") == "deny", f"{area} gate default decision must be deny")
        require(gate.get("implementation_ready") is False, f"{area} gate must not be implementation ready")
        require(str(gate.get("current_boundary") or "").strip(), f"{area} gate must include current boundary")
        require_string_list(gate.get("trigger_scope"), f"{area} gate must include trigger scope")

        denied_actions = set(require_string_list(gate.get("denied_actions"), f"{area} gate must include denied actions"))
        missing_actions = sorted(REQUIRED_DENIED_ACTIONS[area] - denied_actions)
        require(not missing_actions, f"{area} gate missing denied actions: {missing_actions}")

        denial_outputs = set(
            require_string_list(gate.get("required_denial_outputs"), f"{area} gate must include denial outputs")
        )
        missing_outputs = sorted(REQUIRED_DENIAL_OUTPUTS[area] - denial_outputs)
        require(not missing_outputs, f"{area} gate missing denial outputs: {missing_outputs}")

        side_effects = set(
            require_string_list(
                gate.get("side_effect_absence_assertions"),
                f"{area} gate must include side-effect absence assertions",
            )
        )
        require(any(item.startswith("no_") for item in side_effects), f"{area} gate must include no_* side-effect assertions")
        require_string_list(gate.get("required_before_enablement"), f"{area} gate must include enablement blockers")
        require_string_list(gate.get("forbidden_current_scope"), f"{area} gate must include forbidden current scope")

    invariants = set(require_string_list(document.get("shared_invariants"), "implementation gates must include invariants"))
    missing_invariants = sorted(REQUIRED_SHARED_INVARIANTS - invariants)
    require(not missing_invariants, f"implementation gates missing shared invariants: {missing_invariants}")

    blocked = set(require_string_list(document.get("blocked_capabilities"), "implementation gates must include blocked capabilities"))
    missing_blocked = sorted(REQUIRED_BLOCKED_CAPABILITIES - blocked)
    require(not missing_blocked, f"implementation gates missing blocked capabilities: {missing_blocked}")


def check_precondition_alignment(gates: dict[str, dict[str, Any]], preconditions: dict[str, Any]) -> None:
    precondition_areas: dict[str, dict[str, Any]] = {}
    for item in preconditions.get("areas") or []:
        area = require_object(item, "precondition area must be object")
        area_id = str(area.get("area") or "").strip()
        require(area_id in REQUIRED_AREAS, f"unexpected precondition area: {area_id}")
        precondition_areas[area_id] = area
    require(set(precondition_areas) == REQUIRED_AREAS, "preconditions must cover executor/storage/confirmation")

    for area_id, gate in gates.items():
        precondition = precondition_areas[area_id]
        require(precondition.get("status") == "not_ready", f"{area_id} precondition must remain not_ready")
        require(
            gate.get("current_boundary") == precondition.get("current_boundary"),
            f"{area_id} gate boundary must match precondition boundary",
        )
        forbidden = set(require_string_list(gate.get("forbidden_current_scope"), f"{area_id} gate forbidden scope required"))
        precondition_forbidden = set(
            require_string_list(precondition.get("forbidden_current_scope"), f"{area_id} precondition forbidden scope required")
        )
        require(
            bool(forbidden & precondition_forbidden),
            f"{area_id} gate must share forbidden scope with implementation preconditions",
        )


def check_negative_suite_alignment(gates: dict[str, dict[str, Any]], suite: dict[str, Any]) -> None:
    require(
        suite.get("status") == "governance_suite_consumed_deny_by_default_gates_defined",
        "negative suite must consume deny-by-default implementation gates",
    )
    require(
        suite.get("source_deny_by_default_implementation_gates") == relative_path(FIXTURE_PATH),
        "negative suite must reference implementation gates fixture",
    )
    for group in suite.get("groups") or []:
        group_obj = require_object(group, "negative suite group must be object")
        area = str(group_obj.get("precondition_area") or "").strip()
        require(area in gates, f"negative suite group has unknown area: {area}")
        gate_outputs = set(require_string_list(gates[area].get("required_denial_outputs"), f"{area} gate outputs required"))
        for item in group_obj.get("cases") or []:
            case = require_object(item, "negative suite case must be object")
            error_code = str(case.get("expected_error_code") or "").strip()
            require(error_code in gate_outputs, f"{case.get('case_id')} error code is not covered by {area} gate")
            alignment = require_object(case.get("implementation_gate_alignment"), "case must include gate alignment")
            require(alignment.get("gate_area") == area, f"{case.get('case_id')} gate area mismatch")
            require(
                alignment.get("current_status") == SUITE_GATE_STATUS,
                f"{case.get('case_id')} must consume deny-by-default gate contract",
            )
            require(alignment.get("required_before_suite_completion") is True, "gate alignment must still block suite completion")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    readme = TASK_CARDS_README.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name

    require(THIS_CHECK.name in check_repo, "fast baseline must run implementation gates check")
    require(TASK_CARD.name in readme, "task-cards README must reference implementation gates task card")
    for marker in (
        fixture_name,
        "deny-by-default",
        "executor",
        "storage",
        "confirmation",
        "不实现真实工具执行器",
        "不实现 durable store",
        "不接上层 confirmation flow",
    ):
        require(marker in task_card, f"implementation gates task card missing marker: {marker}")


def main() -> int:
    document = require_object(load_json_document(FIXTURE_PATH), "implementation gates fixture must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS_PATH), "preconditions fixture must be object")
    suite = require_object(load_json_document(NEGATIVE_SUITE_PATH), "negative suite fixture must be object")
    check_gate_fixture(document)
    gates = gates_by_area(document)
    check_precondition_alignment(gates, preconditions)
    check_negative_suite_alignment(gates, suite)
    check_docs_and_consumers()
    print("session/tooling deny-by-default implementation gate checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
