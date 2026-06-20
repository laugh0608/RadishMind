#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-handoff-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
ENTRY_REVIEW_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json"
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
    "audit_store_runtime_created",
    "audit_writer_created",
    "audit_event_written",
    "audit_delivery_runtime_created",
    "audit_retention_runtime_created",
    "operator_approval_runtime_executed",
    "no_secret_leakage_smoke_runtime_created",
    "no_secret_leakage_smoke_runtime_executed",
    "real_secret_read",
    "real_secret_written",
    "credential_payload_created",
    "credential_handle_created",
    "credential_handle_runtime_created",
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
    "production-secret-backend-real-resolver-runtime-implementation-entry-review-v1": (
        "real_resolver_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-real-resolver-runtime-preconditions-v1": "real_resolver_runtime_preconditions_defined",
    "production-secret-backend-rotation-audit-policy-readiness-v1": "rotation_audit_policy_readiness_defined",
    "production-secret-backend-operator-approval-runtime-evidence-readiness-v1": (
        "operator_approval_runtime_evidence_readiness_defined"
    ),
    "production-secret-backend-credential-handle-runtime-boundary-readiness-v1": (
        "credential_handle_runtime_boundary_readiness_defined"
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1": (
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
    ),
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "resolver_backend_profile_selection_readiness_defined"
    ),
    "production-secret-backend-operator-runbook-negative-gates-readiness-v1": (
        "operator_runbook_negative_gates_readiness_defined"
    ),
    "production-secret-backend-secret-resolver-interface-disabled-readiness-v1": (
        "secret_resolver_interface_disabled_readiness_defined"
    ),
    "production-secret-backend-provider-profile-secret-binding-readiness-v1": (
        "provider_profile_secret_binding_readiness_defined"
    ),
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
}

EXPECTED_METADATA = {
    "audit_handoff_id",
    "audit_event_id",
    "event_kind",
    "environment",
    "provider",
    "provider_profile",
    "secret_ref_key_status",
    "secret_ref_version_ref",
    "backend_profile_ref",
    "credential_handle_boundary_ref",
    "operator_approval_evidence_ref",
    "approval_ticket_ref",
    "approval_window_ref",
    "request_id",
    "audit_ref",
    "runbook_version",
    "policy_version",
    "rotation_policy_version",
    "delivery_mode",
    "idempotency_key_ref",
    "retention_policy_ref",
    "redaction_profile_ref",
    "failure_code",
    "sanitized_diagnostic",
    "timestamp",
}

EXPECTED_FORBIDDEN_MATERIAL = {
    "credential_payload",
    "secret_value",
    "password",
    "token",
    "api_key",
    "cloud_credential",
    "provider_raw_url",
    "resolver_backend_url",
    "dsn",
    "database_hostname",
    "full_secret_ref_value",
    "full_credential_handle",
    "full_operator_identity_claim",
    "authorization_header",
    "cookie",
    "full_user_claim",
    "raw_audit_payload",
    "raw_request_payload",
    "raw_response_payload",
}

EXPECTED_BINDINGS = {
    "runbook_version",
    "policy_version",
    "rotation_policy_version",
    "secret_ref_key_status",
    "secret_ref_version_ref",
    "provider_profile",
    "environment",
    "backend_profile_ref",
    "credential_handle_boundary_ref",
    "operator_approval_evidence_ref",
    "approval_ticket_ref",
    "approval_window_ref",
    "no_secret_leakage_strategy_ref",
    "retention_policy_ref",
    "redaction_profile_ref",
    "delivery_mode",
    "idempotency_key_ref",
    "audit_ref",
}

EXPECTED_EVENT_KINDS = {
    "secret_resolution_requested",
    "secret_resolution_denied",
    "secret_resolution_failed_closed",
    "credential_handle_boundary_checked",
    "operator_approval_evidence_checked",
    "backend_profile_selected",
    "no_leakage_gate_checked",
    "rotation_policy_checked",
    "audit_handoff_failed_closed",
}

EXPECTED_LIFECYCLE = {
    "handoff_planned",
    "metadata_bound",
    "event_reference_prepared",
    "delivery_pending",
    "delivery_blocked_store_not_created",
    "delivered_metadata_only",
    "delivery_failed_closed",
    "redaction_required",
    "retention_policy_required",
}

EXPECTED_GATE_STATUS = {
    "audit_handoff_envelope": "defined_static_only",
    "metadata_allowlist": "fixed_sanitized_only",
    "payload_and_secret_material": "forbidden",
    "event_kind_allowlist": "defined_static_only",
    "runbook_binding": "required_before_runtime",
    "rotation_policy_binding": "required_before_runtime",
    "secret_ref_binding": "required_before_runtime",
    "provider_profile_binding": "required_before_runtime",
    "environment_binding": "required_no_cross_environment",
    "backend_profile_binding": "required_before_runtime",
    "credential_handle_boundary_binding": "required_before_runtime",
    "operator_approval_evidence_binding": "required_before_runtime",
    "no_leakage_strategy_binding": "required_before_runtime",
    "redaction_retention_policy": "required_before_runtime",
    "delivery_semantics": "required_before_runtime",
    "audit_store": "not_created",
    "audit_writer": "not_created",
    "audit_event_write": "not_executed",
    "production_resolver_runtime": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_FAILURE_CODES = {
    "audit_handoff_boundary_missing",
    "audit_handoff_envelope_missing",
    "audit_handoff_metadata_allowlist_missing",
    "audit_handoff_payload_detected",
    "audit_handoff_secret_material_detected",
    "audit_handoff_event_kind_invalid",
    "audit_handoff_secret_ref_binding_missing",
    "audit_handoff_provider_profile_binding_missing",
    "audit_handoff_environment_binding_mismatch",
    "audit_handoff_backend_profile_binding_missing",
    "audit_handoff_credential_boundary_missing",
    "audit_handoff_operator_approval_reference_missing",
    "audit_handoff_rotation_policy_missing",
    "audit_handoff_redaction_policy_missing",
    "audit_handoff_retention_policy_missing",
    "audit_handoff_delivery_semantics_missing",
    "audit_handoff_idempotency_missing",
    "audit_handoff_store_created_forbidden",
    "audit_handoff_writer_created_forbidden",
    "audit_handoff_write_executed_forbidden",
    "audit_handoff_side_effect_forbidden",
    "audit_handoff_fallback_forbidden",
    "audit_handoff_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "audit_handoff_status",
    "audit_store_status",
    "audit_writer_status",
    "audit_event_delivery_status",
    "event_kind_status",
    "environment",
    "provider",
    "provider_profile",
    "secret_ref_key_status",
    "secret_ref_version_ref_status",
    "backend_profile_ref_status",
    "credential_handle_boundary_status",
    "operator_approval_evidence_status",
    "redaction_profile_status",
    "retention_policy_status",
    "delivery_mode_status",
    "idempotency_key_ref_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_FORBIDDEN_DIAGNOSTICS = {
    "raw_secret",
    "secret_value",
    "password",
    "token",
    "api_key",
    "provider_raw_url",
    "dsn",
    "cloud_credential",
    "database_hostname",
    "database_error_detail",
    "credential_payload",
    "resolver_backend_url",
    "full_secret_ref_value",
    "full_credential_handle",
    "full_operator_claim",
    "full_user_claim",
    "authorization_header",
    "cookie",
    "raw_audit_payload",
    "raw_request_payload",
    "raw_response_payload",
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
        fixture.get("kind") == "production_ops_secret_backend_audit_store_handoff_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "production-secret-backend-audit-store-handoff-readiness-v1", "unexpected slice id")
    require(slice_info.get("status") == "audit_store_handoff_readiness_defined", "unexpected status")
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-audit-store-handoff-readiness-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-audit-store-handoff-readiness-v1.md",
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

    reference = load_json(SECRET_REFERENCE_PATH)
    require(reference.get("scope") == "secret_reference_only", "secret reference must remain reference-only")
    policy = reference.get("policy") or {}
    for field in ("stores_secret_values", "resolver_enabled", "cloud_calls_allowed", "production_secret_backend_ready"):
        require(policy.get(field) is False, f"secret reference policy {field} must remain false")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("audit_handoff_boundary") or {}
    for field, expected in {
        "boundary_status": "defined_without_store_runtime",
        "audit_store_handoff_status": "audit_store_handoff_readiness_defined",
        "entry_review_status": "blocked_before_runtime_task_card",
        "rotation_audit_policy_status": "rotation_audit_policy_readiness_defined",
        "operator_approval_runtime_evidence_status": "operator_approval_runtime_evidence_readiness_defined",
        "credential_handle_runtime_boundary_status": "credential_handle_runtime_boundary_readiness_defined",
        "audit_store_status": "not_created",
        "audit_writer_status": "not_created",
        "audit_event_delivery_status": "not_executed",
        "resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
        "database_connection_provider_status": "blocked",
        "repository_mode_status": "disabled",
        "production_api_status": "not_created",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "runtime_task_card_created_in_this_slice",
        "audit_store_created_in_this_slice",
        "audit_writer_created_in_this_slice",
        "audit_event_written_in_this_slice",
        "audit_delivery_runtime_created_in_this_slice",
        "audit_retention_runtime_created_in_this_slice",
        "production_resolver_runtime_created_in_this_slice",
        "operator_approval_runtime_executed_in_this_slice",
        "credential_payload_created_in_this_slice",
        "credential_handle_runtime_created_in_this_slice",
        "cloud_secret_service_enabled",
        "database_connection_provider_enabled",
        "repository_mode_enabled",
        "production_api_enabled",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("audit_handoff_contract") or {}
    require(contract.get("handoff_definition") == "reference_only_metadata_audit_handoff", "handoff definition drifted")
    require(set(contract.get("allowed_metadata") or []) == EXPECTED_METADATA, "allowed metadata drifted")
    require(set(contract.get("forbidden_material") or []) == EXPECTED_FORBIDDEN_MATERIAL, "forbidden material drifted")
    require(set(contract.get("required_bindings") or []) == EXPECTED_BINDINGS, "required bindings drifted")
    require(set(contract.get("event_kind_allowlist") or []) == EXPECTED_EVENT_KINDS, "event kind allowlist drifted")
    require(set(contract.get("handoff_lifecycle") or []) == EXPECTED_LIFECYCLE, "handoff lifecycle drifted")
    for field in (
        "payload_allowed",
        "secret_material_allowed",
        "store_creation_allowed_in_this_slice",
        "writer_creation_allowed_in_this_slice",
        "event_write_allowed_in_this_slice",
        "credential_resolution_allowed_in_this_slice",
    ):
        require(contract.get(field) is False, f"{field} must remain false")


def assert_gate_matrix_and_failures(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "boundary_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "boundary gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    failures = rows_by_id(fixture, "failure_mapping", "code")
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(failures))
    require(not missing_failures, f"missing failure codes: {missing_failures}")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} missing failure boundary")


def assert_diagnostics_and_side_effects(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    require(set(diagnostics.get("allowed_fields") or []) == EXPECTED_DIAGNOSTIC_FIELDS, "diagnostic allowlist drifted")
    require(
        set(diagnostics.get("forbidden_fields") or []) == EXPECTED_FORBIDDEN_DIAGNOSTICS,
        "diagnostic forbidden fields drifted",
    )
    sample = diagnostics.get("sample_failure") or {}
    require(sample.get("failure_code") == "audit_handoff_boundary_missing", "sample failure code drifted")
    require("secret" not in str(sample.get("sanitized_diagnostic") or "").lower(), "sample diagnostic must stay sanitized")

    counters = fixture.get("side_effect_counters") or {}
    for counter, value in counters.items():
        require(value == 0, f"{counter} must remain 0")


def assert_no_fallback_and_artifacts(fixture: dict[str, Any]) -> None:
    fallback_rules = set(fixture.get("no_fallback_rules") or [])
    for required in {
        "no fallback to operator runbook text",
        "no fallback to developer env plaintext",
        "no fallback to fixture credential",
        "no fallback to fake resolver runtime",
        "no fallback to repository memory store",
        "no cross-environment audit handoff fallback",
    }:
        require(required in fallback_rules, f"missing no fallback rule: {required}")

    guard = fixture.get("artifact_guard") or {}
    for path in guard.get("allowed_static_artifacts") or []:
        require((REPO_ROOT / str(path)).exists(), f"allowed static artifact missing: {path}")
    for path in guard.get("forbidden_artifacts") or []:
        require(not (REPO_ROOT / str(path)).exists(), f"forbidden artifact exists: {path}")


def assert_implementation_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    expected = alignment.get("required_target_status") or {}
    implementation = load_json(IMPLEMENTATION_READINESS_PATH)
    target = implementation.get("implementation_target") or {}
    for field, value in expected.items():
        require(target.get(field) == value, f"implementation readiness {field} drifted")

    planned = rows_by_id(implementation, "planned_slices", "id")
    slice_info = alignment.get("planned_slice") or {}
    slice_id = str(slice_info.get("id") or "")
    require(slice_id in planned, "implementation readiness missing audit handoff planned slice")
    require(planned[slice_id].get("status") == slice_info.get("status"), "audit handoff planned slice status drifted")


def assert_entry_review_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("entry_review_alignment") or {}
    entry = load_json(ENTRY_REVIEW_PATH)
    gates = rows_by_id(entry, "entry_gate_matrix", "gate_id")
    for gate, status in (alignment.get("required_gate_status") or {}).items():
        require(gates[gate].get("status") == status, f"entry review gate {gate} drifted")

    blocked = set(entry.get("blocked_conditions") or [])
    for item in alignment.get("blocked_conditions_must_include") or []:
        require(item in blocked, f"entry review missing blocked condition: {item}")
    for item in alignment.get("blocked_conditions_must_not_include") or []:
        require(item not in blocked, f"entry review still blocks on completed condition: {item}")


def assert_doc_references(fixture: dict[str, Any]) -> None:
    for item in fixture.get("doc_references") or []:
        path = str(item.get("path") or "")
        text = read(path)
        for needle in item.get("must_contain") or []:
            require(str(needle) in text, f"{path} missing required text: {needle}")


def assert_check_repo_registration(fixture: dict[str, Any]) -> None:
    registration = fixture.get("check_repo_registration") or {}
    script = str(registration.get("script") or "")
    after = str(registration.get("after") or "")
    before = str(registration.get("before") or "")
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    script_call = f'run_python_script("{script}", [])'
    after_call = f'run_python_script("{after}", [])'
    before_call = f'run_python_script("{before}", [])'
    require(script_call in check_repo, "check-repo.py must run audit handoff readiness check")
    require(after_call in check_repo, "check-repo.py missing previous operator approval check")
    require(before_call in check_repo, "check-repo.py missing expected next startup supervisor check")
    require(
        check_repo.index(after_call) < check_repo.index(script_call) < check_repo.index(before_call),
        "audit handoff readiness check must run after operator approval and before startup supervisor",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_contract(fixture)
    assert_gate_matrix_and_failures(fixture)
    assert_diagnostics_and_side_effects(fixture)
    assert_no_fallback_and_artifacts(fixture)
    assert_implementation_alignment(fixture)
    assert_entry_review_alignment(fixture)
    assert_doc_references(fixture)
    assert_check_repo_registration(fixture)
    print("production ops secret backend audit store handoff readiness checks passed.")


if __name__ == "__main__":
    main()
