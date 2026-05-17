#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-independent-audit-records-design.json"
PRECONDITIONS_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-implementation-preconditions.json"
CONFIRMATION_FLOW_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-confirmation-flow-design.json"
TOOL_AUDIT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/tool-audit-record-basic.json"
TASK_CARD_PATH = REPO_ROOT / "docs/task-cards/session-tooling-independent-audit-records.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_FIELDS = {
    "audit_record_id",
    "event_type",
    "event_source",
    "occurred_at",
    "request_id",
    "session_id",
    "turn_id",
    "subject_ref",
    "actor_ref",
    "correlation_refs",
    "payload_hash",
    "boundary_decision",
    "side_effects",
    "redaction",
    "storage_policy",
}
REQUIRED_CORRELATION_REFS = {
    "confirmation_id",
    "action_ref",
    "tool_audit_ref",
    "checkpoint_id",
    "policy_decision_ref",
}
REQUIRED_EVENT_SOURCES = {
    "confirmation_flow",
    "executor_policy_gate",
    "storage_policy_gate",
    "checkpoint_read_route",
    "tool_registry_policy",
    "northbound_request",
}
REQUIRED_EVENTS = {
    "business_truth_write_denied",
    "checkpoint_metadata_read_served",
    "checkpoint_read_query_denied",
    "confirmation_deferred",
    "confirmation_invalidated",
    "confirmation_recorded",
    "confirmation_rejected",
    "confirmation_requested",
    "durable_memory_write_denied",
    "executor_attempt_blocked",
    "executor_execution_not_enabled",
    "northbound_failure_boundary_recorded",
    "northbound_request_received",
    "storage_materialization_denied",
    "tool_network_denied",
    "tool_policy_decision_recorded",
}
REQUIRED_BOUNDARY_AREAS = {"confirmation", "executor", "storage"}
REQUIRED_NOT_ENABLED = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_audit_store",
    "durable_memory",
    "durable_session_store",
    "durable_tool_store",
    "long_term_memory",
    "materialized_result_reader",
    "real_tool_executor",
    "result_materialization",
}
REQUIRED_UNSATISFIED = {
    "upper_layer_confirmation_flow",
    "executor_boundary",
    "storage_backend_design",
    "result_materialization_policy",
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
    require(document.get("schema_version") == 1, "independent audit design schema_version must be 1")
    require(
        document.get("kind") == "session_tooling_independent_audit_records_design",
        "independent audit design kind mismatch",
    )
    require(document.get("stage") == "P2 Session & Tooling Foundation", "independent audit design stage mismatch")
    require(document.get("status") == "design_only_not_connected", "independent audit design must stay design-only")
    require(
        document.get("source_preconditions") == relative_path(PRECONDITIONS_PATH),
        "independent audit design must reference implementation preconditions",
    )
    require(
        document.get("source_confirmation_flow_design") == relative_path(CONFIRMATION_FLOW_PATH),
        "independent audit design must reference confirmation flow design",
    )
    require(
        document.get("source_tool_audit_record_fixture") == relative_path(TOOL_AUDIT_FIXTURE_PATH),
        "independent audit design must reference tool audit fixture",
    )
    require(document.get("task_card") == relative_path(TASK_CARD_PATH), "independent audit design must reference task card")

    scope = require_object(document.get("scope"), "independent audit design must include scope")
    require(scope.get("area") == "independent_audit_records", "independent audit scope area mismatch")
    require(scope.get("audit_store_enabled") is False, "independent audit design must not enable audit store")
    require(scope.get("execution_enabled") is False, "independent audit design must not enable execution")
    require(scope.get("result_materialization_enabled") is False, "independent audit design must not materialize results")
    require(scope.get("writes_business_truth") is False, "independent audit design must not write business truth")
    require(scope.get("replay_enabled") is False, "independent audit design must not enable replay")

    record_shape = require_object(document.get("audit_record_shape"), "independent audit design must include record shape")
    fields = set(require_string_list(record_shape.get("required_fields"), "audit record required_fields missing"))
    missing_fields = sorted(REQUIRED_FIELDS - fields)
    require(not missing_fields, f"independent audit record missing fields: {missing_fields}")
    correlation_refs = set(require_string_list(record_shape.get("correlation_refs"), "audit correlation refs missing"))
    missing_refs = sorted(REQUIRED_CORRELATION_REFS - correlation_refs)
    require(not missing_refs, f"independent audit correlation refs missing: {missing_refs}")

    side_effects = require_object(record_shape.get("side_effects"), "audit record side_effects missing")
    allowed = set(require_string_list(side_effects.get("allowed_current_values"), "audit allowed side effects missing"))
    require({"none", "metadata_recorded"}.issubset(allowed), "audit current side effects must stay metadata-only")
    forbidden = set(require_string_list(side_effects.get("forbidden_current_values"), "audit forbidden side effects missing"))
    for marker in (
        "execution_status=executed",
        "writes_business_truth=true",
        "materialized_result_ref",
        "durable_memory_written=true",
        "replay_started=true",
    ):
        require(marker in forbidden, f"audit forbidden side effects missing marker: {marker}")

    redaction = require_object(record_shape.get("redaction"), "audit record redaction policy missing")
    require(redaction.get("requires_secret_redaction") is True, "audit records must require secret redaction")
    require(redaction.get("requires_payload_hash") is True, "audit records must require payload hash")
    require(redaction.get("stores_raw_tool_output") is False, "audit records must not store raw tool output")
    require(redaction.get("stores_credentials") is False, "audit records must not store credentials")

    storage_policy = require_object(record_shape.get("storage_policy"), "audit storage policy missing")
    require(storage_policy.get("current_mode") == "fixture_only", "audit storage current mode must stay fixture_only")
    require(storage_policy.get("durable_store_status") == "not_implemented", "durable audit store must not be implemented")
    require(
        storage_policy.get("checkpoint_read_may_reference_audit_ref") is True,
        "checkpoint read should be allowed to reference audit refs",
    )
    require(
        storage_policy.get("checkpoint_read_is_audit_store") is False,
        "checkpoint read must not become audit store",
    )

    event_sources = document.get("event_sources")
    require(isinstance(event_sources, list) and event_sources, "independent audit design must include event sources")
    seen_sources: set[str] = set()
    seen_events: set[str] = set()
    for source in event_sources:
        source_obj = require_object(source, "audit event source must be object")
        source_id = str(source_obj.get("source") or "").strip()
        require(source_id in REQUIRED_EVENT_SOURCES, f"unexpected audit event source: {source_id}")
        require(source_id not in seen_sources, f"duplicate audit event source: {source_id}")
        seen_sources.add(source_id)
        seen_events.update(require_string_list(source_obj.get("events"), f"{source_id} events missing"))
        boundary = str(source_obj.get("boundary") or "").strip()
        require(boundary, f"{source_id} must include boundary")
        require("execute" in boundary or "execution" in boundary or "store" in boundary or "write" in boundary or "metadata" in boundary, f"{source_id} boundary must describe side-effect limits")
    missing_sources = sorted(REQUIRED_EVENT_SOURCES - seen_sources)
    require(not missing_sources, f"independent audit design missing sources: {missing_sources}")
    missing_events = sorted(REQUIRED_EVENTS - seen_events)
    require(not missing_events, f"independent audit design missing events: {missing_events}")

    boundary_alignment = require_object(document.get("boundary_alignment"), "independent audit design must include boundary alignment")
    missing_areas = sorted(REQUIRED_BOUNDARY_AREAS - set(boundary_alignment))
    require(not missing_areas, f"independent audit design missing boundary areas: {missing_areas}")
    for area in REQUIRED_BOUNDARY_AREAS:
        statements = require_string_list(boundary_alignment.get(area), f"{area} boundary statements missing")
        joined = "\n".join(statements)
        if area == "confirmation":
            require("audit_ref" in joined and "not executed" in joined, "confirmation boundary must keep audit_ref non-executing")
        if area == "executor":
            require("not_implemented" in joined and "execution_enabled=false" in joined, "executor boundary must stay disabled")
        if area == "storage":
            require("fixture-only" in joined or "fixture_only" in joined, "storage boundary must stay fixture-only")
            require("materialized result" in joined, "storage boundary must mention materialized result separation")

    not_enabled = set(require_string_list(document.get("not_enabled_capabilities"), "not enabled capabilities missing"))
    missing_not_enabled = sorted(REQUIRED_NOT_ENABLED - not_enabled)
    require(not missing_not_enabled, f"independent audit design missing blocked capabilities: {missing_not_enabled}")

    precondition_status = require_object(
        document.get("precondition_status_after_design"),
        "independent audit design must include post-design precondition status",
    )
    require(
        precondition_status.get("independent_audit_records") == "design_boundary_defined_not_implemented",
        "independent audit records must be design-only, not implemented",
    )
    for key in REQUIRED_UNSATISFIED:
        require(precondition_status.get(key) == "not_satisfied", f"{key} must remain not_satisfied")


def check_preconditions_alignment(preconditions: dict[str, Any]) -> None:
    for area in preconditions.get("areas") or []:
        area_obj = require_object(area, "precondition area must be object")
        area_id = str(area_obj.get("area") or "").strip()
        blockers = set(require_string_list(area_obj.get("required_before_enablement"), f"{area_id} blockers missing"))
        require("independent_audit_records" in blockers, f"{area_id} must still require independent_audit_records")
        require(area_obj.get("status") == "not_ready", f"{area_id} must remain not_ready")


def check_confirmation_alignment(confirmation_design: dict[str, Any]) -> None:
    audit_events = set(require_string_list(confirmation_design.get("audit_events"), "confirmation audit events missing"))
    required_confirmation_events = {
        "confirmation_requested",
        "confirmation_recorded",
        "confirmation_rejected",
        "confirmation_deferred",
        "confirmation_invalidated",
    }
    missing = sorted(required_confirmation_events - audit_events)
    require(not missing, f"confirmation design missing audit events required by independent audit records: {missing}")


def check_tool_audit_alignment(tool_audit_fixture: dict[str, Any]) -> None:
    require(tool_audit_fixture.get("schema_version") == 1, "tool audit fixture schema_version must be 1")
    audit = require_object(tool_audit_fixture.get("audit"), "tool audit fixture must include audit object")
    require(audit.get("writes_business_truth") is False, "tool audit fixture must not write business truth")
    require(audit.get("durable_memory_written") is False, "tool audit fixture must not write durable memory")
    execution = require_object(tool_audit_fixture.get("execution"), "tool audit fixture must include execution object")
    require(execution.get("status") in {"blocked", "not_executed"}, "tool audit fixture must not record executed status")


def check_docs_and_consumers() -> None:
    task_card = TASK_CARD_PATH.read_text(encoding="utf-8")
    readme = TASK_CARDS_README.read_text(encoding="utf-8")
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name
    for marker in (
        fixture_name,
        "design_only_not_connected",
        "Audit Record Shape",
        "Event Sources",
        "与 Confirmation 的边界",
        "与 Executor 的边界",
        "与 Storage 的边界",
        "不实现 durable audit store",
        "不实现真实 executor",
        "不启用 automatic replay",
    ):
        require(marker in task_card, f"independent audit task card missing marker: {marker}")
    require(
        "session-tooling-independent-audit-records.md" in readme,
        "task-cards README must reference independent audit records task card",
    )
    require(
        "check-session-tooling-independent-audit-records.py" in check_repo,
        "fast baseline must run independent audit records check",
    )


def main() -> int:
    document = require_object(load_json_document(FIXTURE_PATH), "independent audit fixture must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS_PATH), "preconditions fixture must be object")
    confirmation_design = require_object(load_json_document(CONFIRMATION_FLOW_PATH), "confirmation fixture must be object")
    tool_audit_fixture = require_object(load_json_document(TOOL_AUDIT_FIXTURE_PATH), "tool audit fixture must be object")
    check_fixture(document)
    check_preconditions_alignment(preconditions)
    check_confirmation_alignment(confirmation_design)
    check_tool_audit_alignment(tool_audit_fixture)
    check_docs_and_consumers()
    print("session/tooling independent audit records design checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
