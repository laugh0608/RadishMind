#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-executor-boundary-design.json"
PRECONDITIONS_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-implementation-preconditions.json"
INDEPENDENT_AUDIT_DESIGN_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-independent-audit-records-design.json"
RESULT_MATERIALIZATION_POLICY_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-result-materialization-policy-design.json"
NEGATIVE_SKELETON_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-negative-regression-skeleton.json"
TOOL_REGISTRY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/tool-registry-basic.json"
TOOL_AUDIT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/tool-audit-record-basic.json"
TASK_CARD_PATH = REPO_ROOT / "docs/task-cards/session-tooling-executor-boundary.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_ENVELOPE_FIELDS = {
    "execution_request_id",
    "tool_id",
    "request_id",
    "session_id",
    "turn_id",
    "action_ref",
    "action_hash",
    "confirmation_ref",
    "policy_decision_ref",
    "input_ref",
    "output_policy",
    "audit_record_ref",
}
REQUIRED_FORBIDDEN_FIELDS = {
    "executor_ref",
    "process_id",
    "stdout",
    "stderr",
    "raw_tool_output",
    "result_ref",
    "materialized_result_uri",
}
REQUIRED_SANDBOX_PREREQS = {
    "process_or_worker_isolation",
    "filesystem_scope_policy",
    "network_policy",
    "secret_injection_policy",
    "resource_limits",
    "kill_timeout_policy",
}
REQUIRED_ALLOWLIST_PREREQS = {
    "tool_id_allowlist",
    "project_scope_allowlist",
    "risk_level_policy",
    "network_scope_policy",
    "write_scope_policy",
    "confirmation_requirement_policy",
}
REQUIRED_TIMEOUT_PREREQS = {
    "per_tool_timeout_ms",
    "cancellation_boundary",
    "idempotency_key_policy",
    "retry_budget_policy",
    "partial_failure_audit_policy",
}
REQUIRED_FAILURE_CASES = {
    "executor-disabled-tool-run": "TOOL_EXECUTOR_DISABLED",
    "executor-network-disabled": "TOOL_NETWORK_DISABLED",
    "executor-ref-checkpoint-read-denied": "CHECKPOINT_MATERIALIZED_RESULTS_DISABLED",
}
REQUIRED_BOUNDARY_AREAS = {"confirmation", "result_materialization", "audit", "storage"}
REQUIRED_NOT_ENABLED = {
    "automatic_replay",
    "automatic_retry",
    "background_process",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_result_store",
    "durable_session_store",
    "durable_tool_store",
    "executor_ref_reader",
    "long_running_job",
    "long_term_memory",
    "materialized_result_reader",
    "network_tool_execution",
    "raw_tool_output_reader",
    "real_tool_executor",
    "result_ref_reader",
    "shell_execution",
}
REQUIRED_UNSATISFIED = {
    "upper_layer_confirmation_flow",
    "storage_backend_design",
    "negative_regression_suite",
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
    require(document.get("schema_version") == 1, "executor boundary schema_version must be 1")
    require(
        document.get("kind") == "session_tooling_executor_boundary_design",
        "executor boundary kind mismatch",
    )
    require(document.get("stage") == "P2 Session & Tooling Foundation", "executor boundary stage mismatch")
    require(document.get("status") == "design_only_not_connected", "executor boundary must stay design-only")
    require(document.get("source_preconditions") == relative_path(PRECONDITIONS_PATH), "executor boundary must reference preconditions")
    require(
        document.get("source_independent_audit_records_design") == relative_path(INDEPENDENT_AUDIT_DESIGN_PATH),
        "executor boundary must reference independent audit design",
    )
    require(
        document.get("source_result_materialization_policy_design") == relative_path(RESULT_MATERIALIZATION_POLICY_PATH),
        "executor boundary must reference result materialization policy",
    )
    require(
        document.get("source_negative_regression_skeleton") == relative_path(NEGATIVE_SKELETON_PATH),
        "executor boundary must reference negative regression skeleton",
    )
    require(
        document.get("source_tool_registry_fixture") == relative_path(TOOL_REGISTRY_FIXTURE_PATH),
        "executor boundary must reference tool registry fixture",
    )
    require(
        document.get("source_tool_audit_record_fixture") == relative_path(TOOL_AUDIT_FIXTURE_PATH),
        "executor boundary must reference tool audit fixture",
    )
    require(document.get("task_card") == relative_path(TASK_CARD_PATH), "executor boundary must reference task card")

    scope = require_object(document.get("scope"), "executor boundary must include scope")
    require(scope.get("area") == "executor_boundary", "executor boundary scope area mismatch")
    for key in (
        "execution_enabled",
        "sandbox_enabled",
        "allowlist_enforced",
        "network_enabled",
        "result_materialization_enabled",
        "writes_business_truth",
        "replay_enabled",
    ):
        require(scope.get(key) is False, f"{key} must remain false")

    envelope = require_object(document.get("execution_envelope"), "executor boundary must include execution envelope")
    fields = set(require_string_list(envelope.get("required_fields"), "execution envelope fields missing"))
    missing_fields = sorted(REQUIRED_ENVELOPE_FIELDS - fields)
    require(not missing_fields, f"execution envelope missing fields: {missing_fields}")
    forbidden_fields = set(require_string_list(envelope.get("forbidden_current_fields"), "forbidden current fields missing"))
    missing_forbidden = sorted(REQUIRED_FORBIDDEN_FIELDS - forbidden_fields)
    require(not missing_forbidden, f"execution envelope missing forbidden fields: {missing_forbidden}")
    separation_rules = set(require_string_list(envelope.get("separation_rules"), "execution envelope separation rules missing"))
    for rule in (
        "execution_input_output_envelope_is_not_model_prompt_text",
        "checkpoint_read_response_is_not_execution_envelope",
        "tool_audit_record_is_not_materialized_tool_output",
    ):
        require(rule in separation_rules, f"execution envelope missing separation rule: {rule}")

    sandbox = require_object(document.get("sandbox_policy"), "executor boundary must include sandbox policy")
    require(sandbox.get("current_status") == "not_implemented", "sandbox must remain not implemented")
    sandbox_prereqs = set(require_string_list(sandbox.get("required_before_enablement"), "sandbox prereqs missing"))
    missing_sandbox = sorted(REQUIRED_SANDBOX_PREREQS - sandbox_prereqs)
    require(not missing_sandbox, f"sandbox policy missing prerequisites: {missing_sandbox}")
    sandbox_forbidden = set(require_string_list(sandbox.get("current_forbidden_capabilities"), "sandbox forbidden capabilities missing"))
    require("shell_execution" in sandbox_forbidden, "sandbox policy must forbid shell execution")
    require("network_access" in sandbox_forbidden, "sandbox policy must forbid network access")
    require("business_truth_write" in sandbox_forbidden, "sandbox policy must forbid business truth writes")

    allowlist = require_object(document.get("allowlist_policy"), "executor boundary must include allowlist policy")
    require(allowlist.get("current_status") == "not_implemented", "allowlist must remain not implemented")
    require(allowlist.get("default_decision") == "deny", "allowlist default decision must be deny")
    current_allowed = allowlist.get("current_allowed_tools")
    require(isinstance(current_allowed, list) and not current_allowed, "current allowed tools must be empty")
    allowlist_prereqs = set(require_string_list(allowlist.get("required_before_enablement"), "allowlist prereqs missing"))
    missing_allowlist = sorted(REQUIRED_ALLOWLIST_PREREQS - allowlist_prereqs)
    require(not missing_allowlist, f"allowlist policy missing prerequisites: {missing_allowlist}")
    blocked_categories = set(require_string_list(allowlist.get("blocked_categories"), "allowlist blocked categories missing"))
    for category in ("unregistered_tool", "network_tool", "business_truth_write_tool", "replay_tool"):
        require(category in blocked_categories, f"allowlist policy must block {category}")

    timeout_retry = require_object(document.get("timeout_retry_policy"), "executor boundary must include timeout/retry policy")
    require(timeout_retry.get("current_status") == "design_only", "timeout/retry policy must stay design-only")
    require(timeout_retry.get("timeout_required_before_execution") is True, "timeout must be required before execution")
    require(
        timeout_retry.get("retry_default") == "disabled_until_idempotency_policy_exists",
        "retry default must stay disabled until idempotency policy exists",
    )
    timeout_prereqs = set(require_string_list(timeout_retry.get("required_before_enablement"), "timeout/retry prereqs missing"))
    missing_timeout = sorted(REQUIRED_TIMEOUT_PREREQS - timeout_prereqs)
    require(not missing_timeout, f"timeout/retry policy missing prerequisites: {missing_timeout}")
    timeout_forbidden = set(require_string_list(timeout_retry.get("forbidden_current_scope"), "timeout/retry forbidden scope missing"))
    for marker in ("automatic_retry", "background_retry", "replay_retry"):
        require(marker in timeout_forbidden, f"timeout/retry policy must forbid {marker}")

    failures = document.get("failure_boundaries")
    require(isinstance(failures, list) and failures, "executor boundary must include failure boundaries")
    failures_by_id: dict[str, dict[str, Any]] = {}
    for failure in failures:
        failure_obj = require_object(failure, "failure boundary must be object")
        case_id = str(failure_obj.get("case_id") or "").strip()
        require(case_id in REQUIRED_FAILURE_CASES, f"unexpected executor boundary failure case: {case_id}")
        require(case_id not in failures_by_id, f"duplicate executor boundary failure case: {case_id}")
        failures_by_id[case_id] = failure_obj
        require(
            failure_obj.get("expected_error_code") == REQUIRED_FAILURE_CASES[case_id],
            f"{case_id} error code mismatch",
        )
        require(str(failure_obj.get("required_audit_event") or "").strip(), f"{case_id} must include required audit event")
        forbidden_outputs = set(require_string_list(failure_obj.get("forbidden_outputs"), f"{case_id} forbidden outputs missing"))
        require(
            any(item in forbidden_outputs for item in ("execution_status=executed", "executor_ref", "result_ref")),
            f"{case_id} must forbid execution, executor refs, or result refs",
        )
    missing_cases = sorted(set(REQUIRED_FAILURE_CASES) - set(failures_by_id))
    require(not missing_cases, f"executor boundary missing failure cases: {missing_cases}")

    boundary_alignment = require_object(document.get("boundary_alignment"), "executor boundary must include boundary alignment")
    missing_areas = sorted(REQUIRED_BOUNDARY_AREAS - set(boundary_alignment))
    require(not missing_areas, f"executor boundary missing boundary areas: {missing_areas}")
    for area in REQUIRED_BOUNDARY_AREAS:
        statements = require_string_list(boundary_alignment.get(area), f"{area} boundary statements missing")
        joined = "\n".join(statements)
        if area == "confirmation":
            require("confirmation_ref" in joined and "policy gate" in joined, "confirmation alignment mismatch")
        if area == "result_materialization":
            require("result_ref_enabled=false" in joined, "result materialization alignment must keep result refs disabled")
        if area == "audit":
            require("independent audit records" in joined and "not_executed" in joined, "audit alignment mismatch")
        if area == "storage":
            require("does not create durable" in joined and "Business truth writes remain forbidden".lower() in joined.lower(), "storage alignment mismatch")

    not_enabled = set(require_string_list(document.get("not_enabled_capabilities"), "not enabled capabilities missing"))
    missing_not_enabled = sorted(REQUIRED_NOT_ENABLED - not_enabled)
    require(not missing_not_enabled, f"executor boundary missing blocked capabilities: {missing_not_enabled}")

    precondition_status = require_object(
        document.get("precondition_status_after_design"),
        "post-design precondition status missing",
    )
    require(
        precondition_status.get("executor_boundary") == "design_boundary_defined_not_implemented",
        "executor boundary must be design-only, not implemented",
    )
    require(
        precondition_status.get("result_materialization_policy") == "design_boundary_defined_not_implemented",
        "result materialization policy must remain design-only",
    )
    require(
        precondition_status.get("independent_audit_records") == "design_boundary_defined_not_implemented",
        "independent audit records must remain design-only",
    )
    for key in REQUIRED_UNSATISFIED:
        require(precondition_status.get(key) == "not_satisfied", f"{key} must remain not_satisfied")


def check_preconditions_alignment(preconditions: dict[str, Any]) -> None:
    executor_area = next(
        (
            area
            for area in preconditions.get("areas") or []
            if isinstance(area, dict) and area.get("area") == "executor"
        ),
        None,
    )
    executor = require_object(executor_area, "preconditions must include executor area")
    require(executor.get("status") == "not_ready", "executor precondition must remain not_ready")
    blockers = set(require_string_list(executor.get("required_before_enablement"), "executor blockers missing"))
    for blocker in (
        "executor_boundary",
        "result_materialization_policy",
        "independent_audit_records",
        "negative_regression_suite",
    ):
        require(blocker in blockers, f"executor preconditions missing blocker: {blocker}")


def check_result_materialization_alignment(result_policy: dict[str, Any]) -> None:
    scope = require_object(result_policy.get("scope"), "result materialization policy must include scope")
    require(scope.get("executor_enabled") is False, "result policy must keep executor disabled")
    require(scope.get("result_ref_enabled") is False, "result policy must keep result refs disabled")
    require(scope.get("materialized_result_reader_enabled") is False, "result policy must keep materialized result reader disabled")


def check_independent_audit_alignment(independent_audit_design: dict[str, Any]) -> None:
    event_sources = independent_audit_design.get("event_sources")
    require(isinstance(event_sources, list) and event_sources, "independent audit design must include event sources")
    source_events: dict[str, set[str]] = {}
    for source in event_sources:
        if not isinstance(source, dict):
            continue
        source_events[str(source.get("source") or "").strip()] = set(
            require_string_list(source.get("events"), "audit source events missing")
        )
    require(
        "executor_execution_not_enabled" in source_events.get("executor_policy_gate", set()),
        "independent audit design must include executor_execution_not_enabled",
    )
    require(
        "tool_network_denied" in source_events.get("tool_registry_policy", set()),
        "independent audit design must include tool_network_denied",
    )
    require(
        "checkpoint_read_query_denied" in source_events.get("checkpoint_read_route", set()),
        "independent audit design must include checkpoint_read_query_denied",
    )


def check_negative_skeleton_alignment(document: dict[str, Any], negative_skeleton: dict[str, Any]) -> None:
    expected_errors = {
        str(failure.get("expected_error_code") or "").strip()
        for failure in document.get("failure_boundaries") or []
        if isinstance(failure, dict)
    }
    skeleton_errors: set[str] = set()
    for group in negative_skeleton.get("groups") or []:
        if not isinstance(group, dict) or group.get("precondition_area") != "executor":
            continue
        for case in group.get("cases") or []:
            if isinstance(case, dict):
                skeleton_errors.add(str(case.get("expected_error_code") or "").strip())
    missing = sorted(expected_errors - skeleton_errors)
    require(not missing, f"executor boundary missing negative skeleton error codes: {missing}")


def check_tool_registry_alignment(tool_registry: dict[str, Any]) -> None:
    policy = require_object(tool_registry.get("registry_policy"), "tool registry must include registry_policy")
    require(policy.get("execution_enabled") is False, "tool registry execution must remain disabled")
    require(policy.get("network_default") == "disabled", "tool registry network default must remain disabled")
    require(policy.get("durable_memory_enabled") is False, "tool registry durable memory must remain disabled")


def check_tool_audit_alignment(tool_audit_fixture: dict[str, Any]) -> None:
    execution = require_object(tool_audit_fixture.get("execution"), "tool audit fixture must include execution")
    require(execution.get("execution_enabled") is False, "tool audit execution must remain disabled")
    require(execution.get("status") == "not_executed", "tool audit fixture must remain not_executed")
    require(execution.get("executor_ref") is None, "tool audit fixture must not include executor_ref")
    audit = require_object(tool_audit_fixture.get("audit"), "tool audit fixture must include audit")
    require(audit.get("writes_business_truth") is False, "tool audit must not write business truth")


def check_docs_and_consumers() -> None:
    task_card = TASK_CARD_PATH.read_text(encoding="utf-8")
    readme = TASK_CARDS_README.read_text(encoding="utf-8")
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name
    for marker in (
        fixture_name,
        "design_only_not_connected",
        "Execution Envelope",
        "Sandbox Policy",
        "Allowlist Policy",
        "Timeout / Retry",
        "失败边界",
        "与 Confirmation 的边界",
        "与 Result Materialization 的边界",
        "与 Audit 的边界",
        "不实现真实 executor",
        "不启用 automatic replay",
    ):
        require(marker in task_card, f"executor boundary task card missing marker: {marker}")
    require(
        "session-tooling-executor-boundary.md" in readme,
        "task-cards README must reference executor boundary task card",
    )
    require(
        "check-session-tooling-executor-boundary.py" in check_repo,
        "fast baseline must run executor boundary check",
    )


def main() -> int:
    document = require_object(load_json_document(FIXTURE_PATH), "executor boundary fixture must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS_PATH), "preconditions fixture must be object")
    independent_audit_design = require_object(load_json_document(INDEPENDENT_AUDIT_DESIGN_PATH), "independent audit fixture must be object")
    result_policy = require_object(load_json_document(RESULT_MATERIALIZATION_POLICY_PATH), "result materialization fixture must be object")
    negative_skeleton = require_object(load_json_document(NEGATIVE_SKELETON_PATH), "negative skeleton fixture must be object")
    tool_registry = require_object(load_json_document(TOOL_REGISTRY_FIXTURE_PATH), "tool registry fixture must be object")
    tool_audit_fixture = require_object(load_json_document(TOOL_AUDIT_FIXTURE_PATH), "tool audit fixture must be object")
    check_fixture(document)
    check_preconditions_alignment(preconditions)
    check_result_materialization_alignment(result_policy)
    check_independent_audit_alignment(independent_audit_design)
    check_negative_skeleton_alignment(document, negative_skeleton)
    check_tool_registry_alignment(tool_registry)
    check_tool_audit_alignment(tool_audit_fixture)
    check_docs_and_consumers()
    print("session/tooling executor boundary design checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
