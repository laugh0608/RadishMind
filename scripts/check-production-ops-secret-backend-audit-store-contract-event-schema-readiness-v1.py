#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
AUDIT_HANDOFF_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-handoff-readiness-v1.json"
)
AUDIT_RUNTIME_ENTRY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.json"
)
REAL_RESOLVER_ENTRY_REFRESH_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.json"
)
SECRET_REFERENCE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-reference-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "production_ready",
    "production_secret_backend_ready",
    "cloud_secret_service_ready",
    "resolver_implemented",
    "resolver_runtime_ready",
    "production_resolver_runtime_created",
    "production_resolver_runtime_task_card_created",
    "audit_store_runtime_ready",
    "audit_store_runtime_created",
    "audit_store_runtime_task_card_created",
    "audit_writer_ready",
    "audit_writer_created",
    "audit_event_written",
    "audit_event_schema_runtime_created",
    "real_secret_read",
    "real_secret_written",
    "credential_payload_created",
    "credential_handle_created",
    "credential_handle_runtime_created",
    "operator_approval_runtime_created",
    "operator_approval_runtime_executed",
    "backend_health_runtime_created",
    "backend_health_check_executed",
    "no_secret_leakage_smoke_runtime_created",
    "credential_resolved",
    "database_connection_provider_ready",
    "database_driver_selected",
    "connection_factory_implemented",
    "sql_migration_created",
    "schema_marker_ready",
    "migration_runner_implemented",
    "repository_mode_ready",
    "production_secret_audit_store_ready",
    "production_api_ready",
}

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-handoff-readiness-v1": "audit_store_handoff_readiness_defined",
    "production-secret-backend-audit-store-runtime-implementation-entry-review-v1": (
        "audit_store_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1": (
        "operator_approval_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1": (
        "credential_handle_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1": (
        "resolver_backend_health_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1": (
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1": (
        "real_resolver_runtime_implementation_entry_refresh_defined"
    ),
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
}

EXPECTED_CONTRACT_FIELDS = {
    "readiness_status": "defined_without_store_runtime",
    "runtime_task_decision": "audit_store_runtime_implementation_still_blocked_before_task_card",
    "audit_handoff_status": "audit_store_handoff_readiness_defined",
    "audit_store_runtime_entry_review_status": "blocked_before_runtime_task_card",
    "audit_store_contract_status": "static_contract_defined",
    "audit_event_schema_status": "static_contract_defined",
    "runtime_event_schema_status": "not_created",
    "event_version_status": "static_contract_defined",
    "event_kind_allowlist_status": "static_contract_defined",
    "writer_contract_status": "static_contract_defined",
    "audit_store_runtime_status": "not_created",
    "audit_store_status": "not_created",
    "audit_writer_status": "not_created",
    "audit_event_delivery_status": "not_executed",
    "idempotency_key_ref_status": "reference_required_before_runtime",
    "delivery_mode_status": "fail_closed_required_before_runtime",
    "retention_policy_status": "policy_binding_required_before_runtime",
    "redaction_profile_status": "policy_binding_required_before_runtime",
    "operator_approval_runtime_status": "not_created",
    "credential_handle_runtime_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "no_secret_leakage_smoke_runtime_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "cloud_secret_service_status": "not_selected",
    "database_connection_provider_status": "blocked",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_created_in_this_slice",
    "runtime_event_schema_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "audit_delivery_executed_in_this_slice",
    "retention_runtime_executed_in_this_slice",
    "redaction_runtime_executed_in_this_slice",
    "provider_call_executed_in_this_slice",
    "cloud_secret_call_executed_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "credential_payload_created_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "operator_approval_runtime_executed_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "no_secret_leakage_smoke_runtime_executed_in_this_slice",
    "cloud_secret_service_enabled",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_REQUIRED_FIELDS = {
    "event_id",
    "event_version",
    "event_kind",
    "occurred_at",
    "environment",
    "provider",
    "provider_profile",
    "backend_profile_ref",
    "secret_ref_key_status",
    "secret_ref_version_ref",
    "operator_approval_ref",
    "credential_handle_boundary_ref",
    "request_id",
    "audit_ref",
    "policy_version",
    "retention_policy_ref",
    "redaction_profile_ref",
    "idempotency_key_ref",
    "delivery_mode",
}

EXPECTED_OPTIONAL_FIELDS = {
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "rotation_policy_version",
    "runbook_version",
    "approval_ticket_ref",
    "approval_window_ref",
    "backend_health_ref",
    "no_leakage_smoke_ref",
}

EXPECTED_EVENT_KINDS = {
    "secret_resolution_requested",
    "secret_resolution_denied",
    "secret_resolution_failed_closed",
    "credential_handle_boundary_checked",
    "operator_approval_evidence_checked",
    "backend_profile_selected",
    "backend_health_gate_checked",
    "no_leakage_gate_checked",
    "rotation_policy_checked",
    "audit_handoff_failed_closed",
}

EXPECTED_FORBIDDEN_FIELDS = {
    "secret_value",
    "raw_secret",
    "password",
    "token",
    "api_key",
    "provider_raw_url",
    "resolver_backend_url",
    "dsn",
    "database_hostname",
    "cloud_credential",
    "credential_payload",
    "full_secret_ref_value",
    "full_credential_handle",
    "full_operator_identity_claim",
    "full_user_claim",
    "authorization_header",
    "cookie",
    "raw_request_payload",
    "raw_response_payload",
    "raw_audit_payload",
}

EXPECTED_WRITER_BINDINGS = {
    "store ownership",
    "writer ownership",
    "event version",
    "idempotency key reference",
    "delivery mode",
    "retention policy reference",
    "redaction profile reference",
    "operator approval reference",
    "credential handle boundary reference",
    "backend health reference",
    "no leakage smoke reference",
}

EXPECTED_BLOCKED = {
    "audit_store_runtime_task_card_not_created",
    "audit_store_runtime_not_created",
    "audit_writer_not_created",
    "runtime_event_schema_not_created",
    "audit_event_delivery_not_executed",
    "operator_approval_runtime_not_created",
    "credential_handle_runtime_not_created",
    "backend_health_runtime_not_created",
    "real_no_leakage_smoke_runtime_not_created",
    "production_resolver_runtime_not_created",
    "cloud_secret_service_not_selected",
    "repository_mode_disabled",
}

EXPECTED_GATE_STATUS = {
    "audit_store_contract": "defined_without_store_runtime",
    "event_schema": "metadata_only_static_contract",
    "writer_contract": "static_contract_defined",
    "runtime_task_card_gate": "still_blocked_before_task_card",
    "audit_store_runtime": "not_created",
    "audit_writer": "not_created",
    "audit_event_delivery": "not_executed",
    "idempotency": "reference_required_before_runtime",
    "retention_and_redaction": "policy_binding_required_before_runtime",
    "operator_approval_runtime": "blocked_runtime_not_created",
    "credential_handle_runtime": "blocked_runtime_not_created",
    "backend_health_runtime": "blocked_runtime_not_created",
    "no_secret_leakage_smoke_runtime": "blocked_runtime_not_created",
    "production_resolver_runtime": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_contract_event_schema_missing",
    "audit_store_contract_event_kind_invalid",
    "audit_store_contract_required_field_missing",
    "audit_store_contract_forbidden_field_detected",
    "audit_store_contract_secret_material_detected",
    "audit_store_contract_writer_boundary_missing",
    "audit_store_contract_idempotency_missing",
    "audit_store_contract_retention_redaction_missing",
    "audit_store_contract_delivery_semantics_missing",
    "audit_store_contract_runtime_task_card_blocked",
    "audit_store_contract_runtime_created_forbidden",
    "audit_store_contract_writer_created_forbidden",
    "audit_store_contract_event_written_forbidden",
    "audit_store_contract_repository_mode_forbidden",
    "audit_store_contract_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "audit_store_contract_status",
    "event_schema_status",
    "event_version_status",
    "event_kind_status",
    "writer_contract_status",
    "audit_store_runtime_status",
    "audit_writer_status",
    "audit_event_delivery_status",
    "idempotency_key_ref_status",
    "retention_policy_status",
    "redaction_profile_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_SIDE_EFFECT_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "audit_store_runtime_created_count",
    "audit_writer_created_count",
    "audit_event_schema_created_count",
    "audit_event_write_count",
    "audit_delivery_execution_count",
    "operator_approval_runtime_execution_count",
    "credential_payload_created_count",
    "credential_handle_created_count",
    "backend_health_check_count",
    "no_secret_leakage_smoke_runtime_execution_count",
    "network_call_count",
    "cloud_secret_call_count",
    "provider_call_count",
    "database_connection_count",
    "driver_open_count",
    "sql_execution_count",
    "schema_marker_read_count",
    "schema_marker_write_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}

EXPECTED_STATIC_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-contract-event-schema-readiness-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-contract-event-schema-readiness-v1-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json",
    "scripts/check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py",
}

EXPECTED_DOC_REFERENCES = {
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-audit-store-contract-event-schema-readiness-v1",
        "audit_store_contract_event_schema_readiness_defined",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Contract / Event Schema Readiness v1",
        "audit_store_contract_event_schema_readiness_defined",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Contract / Event Schema Readiness v1",
        "audit_store_contract_event_schema_readiness_defined",
    ],
    "docs/task-cards/README.md": [
        "Production Secret Backend Audit Store Contract / Event Schema Readiness",
        "audit_store_contract_event_schema_readiness_defined",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "production-secret-backend-audit-store-contract-event-schema-readiness-v1",
        "audit_store_contract_event_schema_readiness_defined",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py",
        "production-secret-backend-audit-store-contract-event-schema-readiness-v1.json",
    ],
    "docs/devlogs/2026-W25.md": [
        "production-secret-backend-audit-store-contract-event-schema-readiness-v1",
        "audit_store_contract_event_schema_readiness_defined",
    ],
    "docs/devlogs/2026-W25.parts/2026-06-19-to-2026-06-21.md": [
        "Production Secret Backend Audit Store Contract / Event Schema Readiness",
        "audit_store_contract_event_schema_readiness_defined",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a JSON object")
    return document


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_audit_store_contract_event_schema_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-contract-event-schema-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("status") == "audit_store_contract_event_schema_readiness_defined", "unexpected status")
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-audit-store-contract-event-schema-readiness-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-audit-store-contract-event-schema-readiness-v1.md",
    }.items():
        value = str(slice_info.get(key) or "")
        require(value == expected_path, f"unexpected {key}")
        require((REPO_ROOT / value).exists(), f"{key} path missing: {value}")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, expected_status in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        evidence = str(item.get("evidence") or "")
        require((REPO_ROOT / evidence).exists(), f"{dependency_id} evidence missing: {evidence}")

    secret_reference = load_json(SECRET_REFERENCE_PATH)
    require(secret_reference.get("scope") == "secret_reference_only", "secret reference must remain reference-only")
    policy = secret_reference.get("policy") or {}
    for field in ("stores_secret_values", "resolver_enabled", "cloud_calls_allowed", "production_secret_backend_ready"):
        require(policy.get(field) is False, f"secret reference policy {field} must remain false")


def assert_contract_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("contract_boundary") or {}
    for field, expected in EXPECTED_CONTRACT_FIELDS.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")

    audit_handoff = load_json(AUDIT_HANDOFF_PATH).get("audit_handoff_boundary") or {}
    require(audit_handoff.get("audit_store_status") == "not_created", "audit handoff must not create store")
    require(audit_handoff.get("audit_writer_status") == "not_created", "audit handoff must not create writer")
    require(
        audit_handoff.get("audit_event_delivery_status") == "not_executed",
        "audit handoff must not execute delivery",
    )

    audit_runtime_entry = load_json(AUDIT_RUNTIME_ENTRY_PATH).get("entry_boundary") or {}
    require(
        audit_runtime_entry.get("runtime_task_decision")
        == "audit_store_runtime_implementation_blocked_before_task_card",
        "audit store runtime entry review must remain blocked before task card",
    )
    require(audit_runtime_entry.get("audit_store_runtime_status") == "not_created", "audit runtime must not exist")


def assert_event_schema(fixture: dict[str, Any]) -> None:
    schema = fixture.get("event_schema") or {}
    require(schema.get("schema_status") == "metadata_only_static_contract", "event schema status drifted")
    require(schema.get("event_version") == "audit-event-schema-v1", "event version drifted")
    required = set(schema.get("required_fields") or [])
    optional = set(schema.get("optional_fields") or [])
    event_kinds = set(schema.get("event_kind_allowlist") or [])
    reference_only = set(schema.get("reference_only_fields") or [])
    forbidden = set(schema.get("forbidden_fields") or [])
    require(not sorted(EXPECTED_REQUIRED_FIELDS - required), "missing required event schema fields")
    require(not sorted(EXPECTED_OPTIONAL_FIELDS - optional), "missing optional event schema fields")
    require(not sorted(EXPECTED_EVENT_KINDS - event_kinds), "missing event kind allowlist entries")
    require("secret_resolution_requested" in event_kinds, "requested event kind must exist")
    for field in {"backend_profile_ref", "secret_ref_version_ref", "operator_approval_ref", "idempotency_key_ref"}:
        require(field in reference_only, f"{field} must remain reference-only")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_FIELDS - forbidden)
    require(not missing_forbidden, f"missing forbidden fields: {missing_forbidden}")
    require(not (required & forbidden), "required fields must not overlap forbidden fields")
    require(not (optional & forbidden), "optional fields must not overlap forbidden fields")


def assert_writer_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("writer_contract") or {}
    require(contract.get("writer_contract_status") == "static_contract_defined", "writer contract status drifted")
    require(contract.get("input_contract") == "event_schema_allowlist_only", "writer input contract drifted")
    require(contract.get("output_contract") == "delivery_result_sanitized_only", "writer output contract drifted")
    require(contract.get("store_ownership_status") == "blocked_runtime_not_created", "store ownership status drifted")
    require(contract.get("writer_ownership_status") == "blocked_writer_not_created", "writer ownership status drifted")
    require(contract.get("delivery_semantics") == "fail_closed_no_fallback", "delivery semantics drifted")
    require(
        contract.get("idempotency_policy") == "opaque_reference_only_no_raw_payload_hash",
        "idempotency policy drifted",
    )
    allowed_result = set(contract.get("delivery_result_allowed_fields") or [])
    for field in {"delivery_status", "event_id", "audit_ref", "failure_code", "sanitized_diagnostic"}:
        require(field in allowed_result, f"delivery result missing {field}")
    missing_bindings = sorted(EXPECTED_WRITER_BINDINGS - set(contract.get("required_runtime_bindings") or []))
    require(not missing_bindings, f"missing writer bindings: {missing_bindings}")


def assert_blocked_and_gates(fixture: dict[str, Any]) -> None:
    blocked = set(fixture.get("blocked_conditions") or [])
    missing_blocked = sorted(EXPECTED_BLOCKED - blocked)
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    rows = rows_by_id(fixture, "gate_matrix", "gate")
    missing = sorted(set(EXPECTED_GATE_STATUS) - set(rows))
    require(not missing, f"missing gate rows: {missing}")
    for gate, expected_status in EXPECTED_GATE_STATUS.items():
        require(rows[gate].get("status") == expected_status, f"{gate} status drifted")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "failure_mapping", "code")
    missing = sorted(EXPECTED_FAILURE_CODES - set(rows))
    require(not missing, f"missing failure codes: {missing}")
    for code in EXPECTED_FAILURE_CODES:
        item = rows[code]
        require(item.get("boundary"), f"{code} missing failure boundary")
        require(item.get("diagnostic"), f"{code} missing diagnostic")


def assert_diagnostics(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    missing_allowed = sorted(EXPECTED_DIAGNOSTIC_FIELDS - allowed)
    require(not missing_allowed, f"missing allowed diagnostic fields: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_FIELDS - forbidden)
    require(not missing_forbidden, f"missing forbidden diagnostic fields: {missing_forbidden}")
    sample = diagnostics.get("sample") or {}
    require(set(sample).issubset(allowed), "diagnostic sample must only use allowed fields")
    require(sample.get("audit_store_runtime_status") == "not_created", "sample audit store runtime status drifted")
    require(sample.get("audit_writer_status") == "not_created", "sample audit writer status drifted")
    require(sample.get("audit_event_delivery_status") == "not_executed", "sample audit delivery status drifted")


def assert_no_fallback_and_side_effects(fixture: dict[str, Any]) -> None:
    fallback = fixture.get("no_fallback_policy") or {}
    require(fallback.get("fallback_policy") == "forbidden", "fallback policy must be forbidden")
    require(fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    require(
        fallback.get("missing_dependency_must_not_create_audit_success") is True,
        "missing dependency must not create audit success",
    )
    forbidden_fallbacks = set(fallback.get("forbidden_fallbacks") or [])
    for forbidden in {
        "audit_store_handoff_envelope",
        "fake_resolver_runtime",
        "developer_env_plaintext",
        "fixture_credential",
        "sample",
        "mock_provider",
        "operator_runbook_text",
        "repository_memory_store",
        "audit_memory_store",
        "cross_environment_audit_contract",
    }:
        require(forbidden in forbidden_fallbacks, f"missing forbidden fallback: {forbidden}")

    side_effects = fixture.get("no_side_effects") or {}
    require(side_effects.get("checker_mode") == "committed_static_artifacts_only", "checker mode drifted")
    for counter in EXPECTED_SIDE_EFFECT_COUNTERS:
        require(side_effects.get(counter) == 0, f"{counter} must remain zero")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    allowed = set(guard.get("allowed_new_artifacts") or [])
    missing_allowed = sorted(EXPECTED_STATIC_ARTIFACTS - allowed)
    require(not missing_allowed, f"missing allowed artifacts: {missing_allowed}")
    for artifact in EXPECTED_STATIC_ARTIFACTS:
        require((REPO_ROOT / artifact).exists(), f"allowed static artifact missing: {artifact}")

    forbidden = set(guard.get("forbidden_artifacts") or [])
    for artifact in {
        "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-v1-plan.md",
        "services/platform/internal/secretbackend/audit_store.go",
        "services/platform/internal/secretbackend/audit_writer.go",
        "services/platform/internal/secretbackend/audit_event.go",
    }:
        require(artifact in forbidden, f"missing forbidden artifact marker: {artifact}")
        require(not (REPO_ROOT / artifact).exists(), f"forbidden artifact exists: {artifact}")


def assert_alignments(fixture: dict[str, Any]) -> None:
    runtime_alignment = fixture.get("audit_store_runtime_entry_review_alignment") or {}
    for field, expected in {
        "audit_store_runtime_entry_review_status": "blocked_before_runtime_task_card",
        "contract_event_schema_readiness_status": "audit_store_contract_event_schema_readiness_defined",
        "audit_store_runtime_status": "not_created",
        "audit_writer_status": "not_created",
        "audit_event_delivery_status": "not_executed",
        "runtime_task_card_status": "not_created",
    }.items():
        require(runtime_alignment.get(field) == expected, f"runtime alignment {field} drifted")
    for artifact in runtime_alignment.get("required_static_artifacts") or []:
        require((REPO_ROOT / str(artifact)).exists(), f"runtime alignment artifact missing: {artifact}")

    real_refresh = load_json(REAL_RESOLVER_ENTRY_REFRESH_PATH).get("entry_refresh_boundary") or {}
    require(
        real_refresh.get("runtime_task_decision")
        == "real_resolver_runtime_implementation_still_blocked_before_task_card",
        "real resolver runtime entry refresh must remain blocked",
    )
    real_alignment = fixture.get("real_resolver_entry_refresh_alignment") or {}
    require(
        real_alignment.get("production_resolver_runtime_task_card_status") == "not_created",
        "real resolver runtime task card status drifted",
    )
    require(
        real_alignment.get("production_resolver_runtime_status") == "not_created",
        "production resolver runtime status drifted",
    )

    impl = load_json(IMPLEMENTATION_READINESS_PATH)
    target = impl.get("implementation_target") or {}
    alignment = fixture.get("implementation_readiness_alignment") or {}
    for field, expected in (alignment.get("target_fields") or {}).items():
        require(target.get(field) == expected, f"implementation readiness target {field} drifted")
    planned_slice = alignment.get("planned_slice") or {}
    planned = rows_by_id(impl, "planned_slices", "id")
    planned_id = str(planned_slice.get("id") or "")
    require(planned_id in planned, f"implementation readiness missing planned slice: {planned_id}")
    require(planned[planned_id].get("status") == planned_slice.get("status"), "planned slice status drifted")
    planned_evidence = set(planned[planned_id].get("evidence") or [])
    for artifact in planned_slice.get("evidence") or []:
        require(artifact in planned_evidence, f"planned slice missing evidence: {artifact}")
        require((REPO_ROOT / str(artifact)).exists(), f"planned slice evidence missing on disk: {artifact}")

    required_consumers = set(impl.get("required_consumers") or [])
    for consumer in alignment.get("required_consumers") or []:
        require(consumer in required_consumers, f"implementation readiness missing consumer: {consumer}")


def assert_doc_references(fixture: dict[str, Any]) -> None:
    for path in fixture.get("doc_references") or []:
        require((REPO_ROOT / str(path)).exists(), f"doc reference missing: {path}")
    for path, expected_strings in EXPECTED_DOC_REFERENCES.items():
        text = read(path)
        for expected in expected_strings:
            require(expected in text, f"{path} missing expected text: {expected}")


def assert_check_repo_registration(fixture: dict[str, Any]) -> None:
    registration = fixture.get("check_repo_registration") or {}
    script = str(registration.get("script") or "")
    after = str(registration.get("after") or "")
    before = str(registration.get("before") or "")
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    for marker in (script, after, before):
        require(marker in check_repo, f"check-repo missing marker: {marker}")
    require(check_repo.index(after) < check_repo.index(script) < check_repo.index(before), "check-repo order drifted")


def assert_no_secret_like_values(fixture: dict[str, Any]) -> None:
    serialized = json.dumps(fixture, sort_keys=True)
    for forbidden_literal in (
        "sk-live-",
        "sk-test-",
        "AKIA",
        "BEGIN PRIVATE KEY",
        "xoxb-",
        "postgres://",
        "mysql://",
        "mongodb://",
    ):
        require(forbidden_literal not in serialized, f"secret-like literal found: {forbidden_literal}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_contract_boundary(fixture)
    assert_event_schema(fixture)
    assert_writer_contract(fixture)
    assert_blocked_and_gates(fixture)
    assert_failure_mapping(fixture)
    assert_diagnostics(fixture)
    assert_no_fallback_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_alignments(fixture)
    assert_doc_references(fixture)
    assert_check_repo_registration(fixture)
    assert_no_secret_like_values(fixture)
    print("production ops secret backend audit store contract event schema readiness checks passed.")


if __name__ == "__main__":
    main()
