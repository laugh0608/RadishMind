#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

SUITE_PATH = FIXTURE_DIR / "session-tooling-negative-regression-suite.json"
SKELETON_PATH = FIXTURE_DIR / "session-tooling-negative-regression-skeleton.json"
PRECONDITIONS_PATH = FIXTURE_DIR / "session-tooling-implementation-preconditions.json"
DENIED_QUERY_FIXTURE = FIXTURE_DIR / "session-recovery-checkpoint-read-denied-queries.json"
INDEPENDENT_AUDIT = FIXTURE_DIR / "session-tooling-independent-audit-records-design.json"
RESULT_MATERIALIZATION = FIXTURE_DIR / "session-tooling-result-materialization-policy-design.json"
CONFIRMATION_FLOW = FIXTURE_DIR / "session-tooling-confirmation-flow-design.json"
TOOL_REGISTRY = FIXTURE_DIR / "tool-registry-basic.json"
IMPLEMENTATION_GATES = FIXTURE_DIR / "session-tooling-deny-by-default-implementation-gates.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-negative-regression-suite.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-negative-regression-suite.py"

GROUP_PRECONDITIONS = {
    "executor_blocked": "executor",
    "storage_materialization_blocked": "storage",
    "confirmation_blocked": "confirmation",
}

CASE_CONSUMERS = {
    "executor-disabled-tool-run": [
        "deny_by_default_implementation_gates",
        "implementation_preconditions",
        "tool_registry_policy",
        "negative_regression_suite_check",
    ],
    "executor-network-disabled": [
        "deny_by_default_implementation_gates",
        "tool_registry_policy",
        "independent_audit_records_design",
        "negative_regression_suite_check",
    ],
    "executor-ref-checkpoint-read-denied": [
        "deny_by_default_implementation_gates",
        "checkpoint_denied_query_fixture",
        "result_materialization_policy",
        "negative_regression_suite_check",
    ],
    "materialized-result-read-denied": [
        "deny_by_default_implementation_gates",
        "checkpoint_denied_query_fixture",
        "result_materialization_policy",
        "negative_regression_suite_check",
    ],
    "durable-memory-write-denied": [
        "deny_by_default_implementation_gates",
        "implementation_preconditions",
        "independent_audit_records_design",
        "negative_regression_suite_check",
    ],
    "business-truth-write-denied": [
        "deny_by_default_implementation_gates",
        "implementation_preconditions",
        "independent_audit_records_design",
        "negative_regression_suite_check",
    ],
    "missing-confirmation-denied": [
        "deny_by_default_implementation_gates",
        "confirmation_flow_design",
        "implementation_preconditions",
        "negative_regression_suite_check",
    ],
    "stale-confirmation-denied": [
        "deny_by_default_implementation_gates",
        "confirmation_flow_design",
        "implementation_preconditions",
        "negative_regression_suite_check",
    ],
    "mismatched-confirmation-payload-denied": [
        "deny_by_default_implementation_gates",
        "confirmation_flow_design",
        "implementation_preconditions",
        "negative_regression_suite_check",
    ],
}

CASE_AUDIT_EXPECTATIONS = {
    "executor-disabled-tool-run": ("executor_policy_gate", "executor_execution_not_enabled"),
    "executor-network-disabled": ("tool_registry_policy", "tool_network_denied"),
    "executor-ref-checkpoint-read-denied": ("checkpoint_read_route", "checkpoint_read_query_denied"),
    "materialized-result-read-denied": ("checkpoint_read_route", "checkpoint_read_query_denied"),
    "durable-memory-write-denied": ("storage_policy_gate", "durable_memory_write_denied"),
    "business-truth-write-denied": ("storage_policy_gate", "business_truth_write_denied"),
    "missing-confirmation-denied": ("confirmation_flow", "confirmation_requested"),
    "stale-confirmation-denied": ("confirmation_flow", "confirmation_invalidated"),
    "mismatched-confirmation-payload-denied": ("confirmation_flow", "confirmation_invalidated"),
}

CASE_SIDE_EFFECT_ASSERTIONS = {
    "executor-disabled-tool-run": [
        "no_execution_side_effects",
        "no_executor_ref",
        "no_result_ref",
        "no_business_truth_write_side_effects",
    ],
    "executor-network-disabled": [
        "no_network_side_effects",
        "no_execution_side_effects",
        "no_result_ref",
        "no_durable_memory_write_side_effects",
    ],
    "executor-ref-checkpoint-read-denied": [
        "no_executor_ref",
        "no_execution_side_effects",
        "no_materialized_result_side_effects",
    ],
    "materialized-result-read-denied": [
        "no_materialized_result_side_effects",
        "no_result_ref",
        "no_output_ref",
    ],
    "durable-memory-write-denied": [
        "no_durable_memory_write_side_effects",
        "no_long_term_memory",
        "no_business_truth_write_side_effects",
    ],
    "business-truth-write-denied": [
        "no_business_truth_write_side_effects",
        "no_confirmed_action_execution",
        "no_durable_memory_write_side_effects",
    ],
    "missing-confirmation-denied": [
        "no_execution_side_effects",
        "no_confirmed_action_execution",
        "no_business_truth_write_side_effects",
    ],
    "stale-confirmation-denied": [
        "no_execution_side_effects",
        "no_replay_side_effects",
        "no_business_truth_write_side_effects",
    ],
    "mismatched-confirmation-payload-denied": [
        "no_execution_side_effects",
        "no_confirmed_action_execution",
        "no_result_ref",
    ],
}

CONSUMER_PATHS = {
    "checkpoint_denied_query_fixture": DENIED_QUERY_FIXTURE,
    "confirmation_flow_design": CONFIRMATION_FLOW,
    "deny_by_default_implementation_gates": IMPLEMENTATION_GATES,
    "implementation_preconditions": PRECONDITIONS_PATH,
    "independent_audit_records_design": INDEPENDENT_AUDIT,
    "negative_regression_suite_check": THIS_CHECK,
    "result_materialization_policy": RESULT_MATERIALIZATION,
    "tool_registry_policy": TOOL_REGISTRY,
}

REQUIRED_BLOCKED_CAPABILITIES = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_audit_store",
    "durable_memory",
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


def case_ids_from_skeleton(skeleton: dict[str, Any]) -> set[str]:
    case_ids: set[str] = set()
    for group in skeleton.get("groups") or []:
        group_obj = require_object(group, "skeleton group must be object")
        for case in group_obj.get("cases") or []:
            case_obj = require_object(case, "skeleton case must be object")
            case_id = str(case_obj.get("case_id") or "").strip()
            require(case_id, "skeleton case must include case_id")
            case_ids.add(case_id)
    return case_ids


def audit_event_sources(audit_design: dict[str, Any]) -> dict[str, set[str]]:
    sources: dict[str, set[str]] = {}
    for source in audit_design.get("event_sources") or []:
        source_obj = require_object(source, "audit event source must be object")
        source_id = str(source_obj.get("source") or "").strip()
        events = set(require_string_list(source_obj.get("events"), f"{source_id} must include events"))
        sources[source_id] = events
    return sources


def suite_cases_for_group(group: dict[str, Any]) -> list[dict[str, Any]]:
    group_id = str(group.get("group_id") or "").strip()
    precondition_area = GROUP_PRECONDITIONS[group_id]
    cases: list[dict[str, Any]] = []
    for raw_case in group.get("cases") or []:
        case = require_object(raw_case, f"{group_id} skeleton case must be object")
        case_id = str(case.get("case_id") or "").strip()
        require(case_id in CASE_CONSUMERS, f"{case_id} missing suite consumers")
        audit_source, audit_event = CASE_AUDIT_EXPECTATIONS[case_id]
        consumers = [
            {
                "consumer_id": consumer_id,
                "path": relative_path(CONSUMER_PATHS[consumer_id]),
            }
            for consumer_id in CASE_CONSUMERS[case_id]
        ]
        cases.append(
            {
                "case_id": case_id,
                "source_skeleton_intent": str(case.get("intent") or "").strip(),
                "expected_failure_boundary": str(case.get("expected_failure_boundary") or "").strip(),
                "expected_error_code": str(case.get("expected_error_code") or "").strip(),
                "consumers": consumers,
                "audit_assertion": {
                    "mode": "audit_non_write_boundary",
                    "expected_event_source": audit_source,
                    "expected_event": audit_event,
                    "durable_audit_store_write": False,
                    "raw_payload_stored": False,
                },
                "side_effect_absence_assertions": CASE_SIDE_EFFECT_ASSERTIONS[case_id],
                "forbidden_outputs": require_string_list(
                    case.get("forbidden_outputs"),
                    f"{case_id} must include forbidden outputs",
                ),
                "implementation_gate_alignment": {
                    "gate_area": precondition_area,
                    "current_status": "deny_by_default_gate_contract_defined_implementation_blocked",
                    "required_before_suite_completion": True,
                },
            }
        )
    return cases


def build_suite() -> dict[str, Any]:
    skeleton = require_object(load_json_document(SKELETON_PATH), "negative regression skeleton must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS_PATH), "preconditions must be object")
    audit_design = require_object(load_json_document(INDEPENDENT_AUDIT), "independent audit design must be object")
    result_policy = require_object(load_json_document(RESULT_MATERIALIZATION), "result policy must be object")
    confirmation = require_object(load_json_document(CONFIRMATION_FLOW), "confirmation flow must be object")
    registry = require_object(load_json_document(TOOL_REGISTRY), "tool registry must be object")
    denied_queries = require_object(load_json_document(DENIED_QUERY_FIXTURE), "denied query fixture must be object")
    implementation_gates = require_object(load_json_document(IMPLEMENTATION_GATES), "implementation gates must be object")

    require(skeleton.get("status") == "skeleton_only_implementation_blocked", "skeleton must remain blocked")
    require(preconditions.get("status") == "preconditions_not_satisfied", "preconditions must remain unsatisfied")
    require(audit_design.get("status") == "design_only_not_connected", "audit design must remain design-only")
    require(result_policy.get("status") == "design_only_not_connected", "result policy must remain design-only")
    require(confirmation.get("status") == "design_only_not_connected", "confirmation must remain design-only")
    require((registry.get("registry_policy") or {}).get("execution_enabled") is False, "registry must keep execution disabled")
    require(isinstance(denied_queries.get("cases"), list), "denied query fixture must include cases")
    require(
        implementation_gates.get("status") == "deny_by_default_gates_defined_implementation_blocked",
        "implementation gates must be defined but blocked",
    )

    groups: list[dict[str, Any]] = []
    for raw_group in skeleton.get("groups") or []:
        group = require_object(raw_group, "skeleton group must be object")
        group_id = str(group.get("group_id") or "").strip()
        require(group_id in GROUP_PRECONDITIONS, f"unexpected skeleton group: {group_id}")
        groups.append(
            {
                "group_id": group_id,
                "precondition_area": GROUP_PRECONDITIONS[group_id],
                "current_status": "governance_consumed_implementation_blocked",
                "cases": suite_cases_for_group(group),
            }
        )

    return {
        "schema_version": 1,
        "kind": "session_tooling_negative_regression_suite",
        "stage": "P2 Session & Tooling Foundation",
        "status": "governance_suite_consumed_deny_by_default_gates_defined",
        "implementation_status": "not_ready",
        "source_negative_regression_skeleton": relative_path(SKELETON_PATH),
        "source_implementation_preconditions": relative_path(PRECONDITIONS_PATH),
        "source_denied_query_fixture": relative_path(DENIED_QUERY_FIXTURE),
        "source_independent_audit_records_design": relative_path(INDEPENDENT_AUDIT),
        "source_result_materialization_policy_design": relative_path(RESULT_MATERIALIZATION),
        "source_confirmation_flow_design": relative_path(CONFIRMATION_FLOW),
        "source_tool_registry_fixture": relative_path(TOOL_REGISTRY),
        "source_deny_by_default_implementation_gates": relative_path(IMPLEMENTATION_GATES),
        "coverage_summary": {
            "group_count": len(groups),
            "case_count": sum(len(group["cases"]) for group in groups),
            "consumer_assertions": "governance_consumers_present_for_all_cases",
            "audit_assertions": "audit_non_write_boundary_asserted_for_all_cases",
            "side_effect_absence_assertions": "forbidden_side_effect_absence_asserted_for_all_cases",
            "implementation_gate_alignment": "satisfied_by_deny_by_default_gate_contracts",
        },
        "groups": groups,
        "suite_acceptance_progress": [
            {
                "requirement_id": "real_consumers_before_completion",
                "status": "partially_satisfied_by_governance_suite",
            },
            {
                "requirement_id": "independent_audit_assertions",
                "status": "satisfied_by_audit_non_write_boundary",
            },
            {
                "requirement_id": "side_effect_absence_assertions",
                "status": "satisfied_by_forbidden_side_effect_assertions",
            },
            {
                "requirement_id": "checkpoint_denied_query_alignment",
                "status": "satisfied_by_denied_query_fixture_and_suite_cases",
            },
            {
                "requirement_id": "implementation_gate_alignment",
                "status": "satisfied_by_deny_by_default_gate_contracts",
            },
        ],
        "blocked_completion_reasons": [
            "real_executor_storage_confirmation_implementations_missing",
            "executor_storage_confirmation_still_not_ready",
            "upper_layer_confirmation_flow_missing",
            "durable_store_and_result_reader_still_disabled",
        ],
        "not_enabled_capabilities": sorted(REQUIRED_BLOCKED_CAPABILITIES),
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_negative_regression_suite",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_suite_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "suite schema_version must be 1")
    require(document.get("kind") == "session_tooling_negative_regression_suite", "suite kind mismatch")
    require(
        document.get("status") == "governance_suite_consumed_deny_by_default_gates_defined",
        "suite must not claim completion",
    )
    require(document.get("implementation_status") == "not_ready", "suite must keep implementation not_ready")

    groups = document.get("groups")
    require(isinstance(groups, list) and groups, "suite must include groups")
    group_ids = {str(group.get("group_id") or "").strip() for group in groups if isinstance(group, dict)}
    require(group_ids == set(GROUP_PRECONDITIONS), "suite groups must match skeleton groups")

    audit_sources = audit_event_sources(require_object(load_json_document(INDEPENDENT_AUDIT), "audit design must be object"))
    skeleton_case_ids = case_ids_from_skeleton(
        require_object(load_json_document(SKELETON_PATH), "negative regression skeleton must be object")
    )
    suite_case_ids: set[str] = set()

    for group in groups:
        group_obj = require_object(group, "suite group must be object")
        group_id = str(group_obj.get("group_id") or "").strip()
        require(group_obj.get("precondition_area") == GROUP_PRECONDITIONS[group_id], f"{group_id} area mismatch")
        require(
            group_obj.get("current_status") == "governance_consumed_implementation_blocked",
            f"{group_id} must remain implementation blocked",
        )
        cases = group_obj.get("cases")
        require(isinstance(cases, list) and cases, f"{group_id} must include cases")
        for raw_case in cases:
            case = require_object(raw_case, f"{group_id} case must be object")
            case_id = str(case.get("case_id") or "").strip()
            require(case_id in CASE_CONSUMERS, f"unexpected suite case: {case_id}")
            require(case_id not in suite_case_ids, f"duplicate suite case: {case_id}")
            suite_case_ids.add(case_id)

            consumer_ids = {
                str(consumer.get("consumer_id") or "").strip()
                for consumer in case.get("consumers") or []
                if isinstance(consumer, dict)
            }
            require(set(CASE_CONSUMERS[case_id]) == consumer_ids, f"{case_id} consumers mismatch")

            audit = require_object(case.get("audit_assertion"), f"{case_id} must include audit assertion")
            require(audit.get("mode") == "audit_non_write_boundary", f"{case_id} must assert audit non-write boundary")
            require(audit.get("durable_audit_store_write") is False, f"{case_id} must not write durable audit store")
            require(audit.get("raw_payload_stored") is False, f"{case_id} must not store raw payload")
            event_source = str(audit.get("expected_event_source") or "").strip()
            event = str(audit.get("expected_event") or "").strip()
            require(event_source in audit_sources, f"{case_id} unknown audit event source: {event_source}")
            require(event in audit_sources[event_source], f"{case_id} unknown audit event: {event}")

            side_effects = set(
                require_string_list(
                    case.get("side_effect_absence_assertions"),
                    f"{case_id} must include side-effect assertions",
                )
            )
            require(set(CASE_SIDE_EFFECT_ASSERTIONS[case_id]) == side_effects, f"{case_id} side effects mismatch")
            forbidden_outputs = set(require_string_list(case.get("forbidden_outputs"), f"{case_id} must include forbidden outputs"))
            require(forbidden_outputs, f"{case_id} must include forbidden outputs")

            gate = require_object(case.get("implementation_gate_alignment"), f"{case_id} must include gate alignment")
            require(gate.get("gate_area") == GROUP_PRECONDITIONS[group_id], f"{case_id} gate area mismatch")
            require(
                gate.get("current_status") == "deny_by_default_gate_contract_defined_implementation_blocked",
                f"{case_id} must consume deny-by-default gate contract",
            )
            require(gate.get("required_before_suite_completion") is True, f"{case_id} must block completion")

    require(suite_case_ids == skeleton_case_ids, "suite cases must exactly match skeleton cases")

    progress = document.get("suite_acceptance_progress")
    require(isinstance(progress, list) and progress, "suite must include acceptance progress")
    progress_by_id = {
        str(item.get("requirement_id") or "").strip(): str(item.get("status") or "").strip()
        for item in progress
        if isinstance(item, dict)
    }
    require(
        progress_by_id.get("implementation_gate_alignment") == "satisfied_by_deny_by_default_gate_contracts",
        "suite must align implementation gate contract",
    )
    for requirement_id in (
        "real_consumers_before_completion",
        "independent_audit_assertions",
        "side_effect_absence_assertions",
        "checkpoint_denied_query_alignment",
    ):
        require(requirement_id in progress_by_id, f"suite missing progress for {requirement_id}")

    not_enabled = set(require_string_list(document.get("not_enabled_capabilities"), "suite must include blocked capabilities"))
    missing = sorted(REQUIRED_BLOCKED_CAPABILITIES - not_enabled)
    require(not missing, f"suite missing blocked capabilities: {missing}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    contracts_readme = CONTRACTS_README.read_text(encoding="utf-8")
    fixture_name = SUITE_PATH.name

    require(THIS_CHECK.name in check_repo, "fast baseline must run negative regression suite check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference suite task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", contracts_readme),
    ):
        require(fixture_name in content, f"{label} must reference suite fixture")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("implementation gate" in content, f"{label} must mention implementation gate")


def main() -> int:
    expected = require_object(load_json_document(SUITE_PATH), "negative regression suite fixture must be object")
    check_suite_shape(expected)
    actual = build_suite()
    if actual != expected:
        raise SystemExit("session/tooling negative regression suite does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling negative regression suite checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
