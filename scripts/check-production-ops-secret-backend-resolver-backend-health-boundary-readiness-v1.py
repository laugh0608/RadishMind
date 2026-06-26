#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-boundary-readiness-v1.json"
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
    "backend_health_runtime_created",
    "backend_health_check_executed",
    "backend_runtime_created",
    "real_secret_read",
    "real_secret_written",
    "credential_payload_created",
    "credential_handle_created",
    "credential_handle_runtime_created",
    "operator_approval_runtime_executed",
    "audit_store_runtime_created",
    "audit_writer_created",
    "audit_event_written",
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
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "resolver_backend_profile_selection_readiness_defined"
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1": (
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
    ),
    "production-secret-backend-credential-handle-runtime-boundary-readiness-v1": (
        "credential_handle_runtime_boundary_readiness_defined"
    ),
    "production-secret-backend-operator-approval-runtime-evidence-readiness-v1": (
        "operator_approval_runtime_evidence_readiness_defined"
    ),
    "production-secret-backend-audit-store-handoff-readiness-v1": "audit_store_handoff_readiness_defined",
    "production-secret-backend-operator-runbook-negative-gates-readiness-v1": (
        "operator_runbook_negative_gates_readiness_defined"
    ),
    "production-secret-backend-rotation-audit-policy-readiness-v1": "rotation_audit_policy_readiness_defined",
    "production-secret-backend-provider-profile-secret-binding-readiness-v1": (
        "provider_profile_secret_binding_readiness_defined"
    ),
    "production-secret-backend-secret-resolver-interface-disabled-readiness-v1": (
        "secret_resolver_interface_disabled_readiness_defined"
    ),
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
}

EXPECTED_REQUIRED_INPUTS = {
    "backend_health_ref",
    "backend_profile_ref",
    "backend_profile_id",
    "backend_kind",
    "environment",
    "provider_profile_id",
    "secret_ref_key_status",
    "secret_ref_version_ref_status",
    "health_policy_version",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_METADATA = {
    "health_boundary_status",
    "health_runtime_status",
    "health_check_status",
    "backend_health_ref",
    "backend_profile_ref",
    "backend_profile_id",
    "backend_kind",
    "environment",
    "provider_profile_id",
    "secret_ref_key_status",
    "secret_ref_version_ref_status",
    "health_policy_version",
    "health_lifecycle_state",
    "last_known_health_ref",
    "sanitized_backend_diagnostic",
    "failure_code",
    "failure_boundary",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_FORBIDDEN_MATERIAL = {
    "secret_material",
    "secret_value",
    "password",
    "token",
    "api_key",
    "cloud_credential",
    "credential_payload",
    "full_credential_handle",
    "provider_raw_url",
    "resolver_backend_url",
    "backend_endpoint_url",
    "dsn",
    "database_hostname",
    "full_secret_ref_value",
    "full_operator_identity_claim",
    "authorization_header",
    "cookie",
    "full_user_claim",
    "raw_health_request",
    "raw_health_response",
    "raw_provider_error",
    "raw_backend_error_detail",
}

EXPECTED_BINDINGS = {
    "backend_profile_ref",
    "backend_profile_id",
    "backend_kind",
    "environment",
    "provider_profile_id",
    "secret_ref_key_status",
    "secret_ref_version_ref_status",
    "health_policy_version",
    "operator_approval_evidence_ref",
    "audit_handoff_ref",
    "rotation_policy_ref",
    "no_secret_leakage_strategy_ref",
    "credential_handle_boundary_ref",
    "request_id",
    "audit_ref",
}

EXPECTED_LIFECYCLE = {
    "boundary_planned",
    "metadata_bound",
    "health_reference_prepared",
    "runtime_not_created",
    "health_check_not_executed",
    "health_status_unknown",
    "health_status_available_metadata_only",
    "health_status_degraded_metadata_only",
    "health_status_unavailable_fail_closed",
    "diagnostics_sanitized",
    "runtime_implementation_deferred",
}

EXPECTED_HEALTH_STATUS = {
    "unknown_reference_only",
    "available_metadata_only",
    "degraded_metadata_only",
    "unavailable_fail_closed",
    "blocked_dependency_missing",
    "blocked_environment_mismatch",
    "blocked_policy_missing",
    "blocked_runtime_not_created",
}

EXPECTED_GATE_STATUS = {
    "backend_health_reference": "defined_static_only",
    "backend_profile_binding": "required_before_runtime",
    "environment_binding": "required_no_cross_environment",
    "provider_profile_binding": "required_before_runtime",
    "secret_ref_binding": "required_reference_only",
    "health_metadata_allowlist": "fixed_sanitized_only",
    "health_lifecycle_status_allowlist": "fixed_static_only",
    "failure_mapping": "fail_closed",
    "sanitized_diagnostics": "fixed_sanitized_only",
    "fallback_policy": "forbidden",
    "side_effect_policy": "no_runtime_calls",
    "artifact_guard": "enforced",
    "backend_health_runtime": "not_created",
    "backend_health_check_execution": "not_executed",
    "production_resolver_runtime": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_FAILURE_CODES = {
    "resolver_backend_health_boundary_missing",
    "resolver_backend_health_reference_missing",
    "resolver_backend_health_profile_binding_missing",
    "resolver_backend_health_environment_mismatch",
    "resolver_backend_health_provider_profile_mismatch",
    "resolver_backend_health_secret_ref_binding_missing",
    "resolver_backend_health_policy_missing",
    "resolver_backend_health_metadata_forbidden",
    "resolver_backend_health_secret_material_detected",
    "resolver_backend_health_raw_endpoint_detected",
    "resolver_backend_health_runtime_created_forbidden",
    "resolver_backend_health_check_executed_forbidden",
    "resolver_backend_health_fallback_forbidden",
    "resolver_backend_health_side_effect_forbidden",
    "resolver_backend_health_repository_mode_forbidden",
    "resolver_backend_health_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "health_boundary_status",
    "health_runtime_status",
    "health_check_status",
    "backend_profile_ref_status",
    "backend_kind",
    "environment",
    "provider_profile_id",
    "secret_ref_key_status",
    "secret_ref_version_ref_status",
    "health_policy_version",
    "health_lifecycle_state",
    "last_known_health_ref_status",
    "failure_code",
    "failure_boundary",
    "sanitized_backend_diagnostic",
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
    "resolver_backend_url",
    "backend_endpoint_url",
    "dsn",
    "cloud_credential",
    "database_hostname",
    "database_error_detail",
    "credential_payload",
    "full_secret_ref_value",
    "full_credential_handle",
    "full_operator_claim",
    "full_user_claim",
    "authorization_header",
    "cookie",
    "raw_health_request",
    "raw_health_response",
    "raw_backend_error_detail",
}

EXPECTED_RUNTIME_SPLIT = {
    "resolver-backend-health-runtime-implementation-entry-review",
    "resolver-backend-health-runtime-implementation",
    "production-secret-backend-real-resolver-runtime-implementation-entry-refresh",
    "production-secret-backend-real-resolver-runtime-implementation",
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
        fixture.get("kind") == "production_ops_secret_backend_resolver_backend_health_boundary_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-resolver-backend-health-boundary-readiness-v1",
        "unexpected slice id",
    )
    require(
        slice_info.get("status") == "resolver_backend_health_boundary_readiness_defined",
        "unexpected status",
    )
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-resolver-backend-health-boundary-readiness-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-resolver-backend-health-boundary-readiness-v1.md",
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


def assert_health_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("health_boundary") or {}
    for field, expected in {
        "boundary_status": "defined_without_backend_health_runtime",
        "resolver_backend_health_boundary_status": "resolver_backend_health_boundary_readiness_defined",
        "entry_review_status": "blocked_before_runtime_task_card",
        "resolver_backend_profile_selection_readiness_status": "resolver_backend_profile_selection_readiness_defined",
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_status": (
            "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
        ),
        "credential_handle_runtime_boundary_status": "credential_handle_runtime_boundary_readiness_defined",
        "operator_approval_runtime_evidence_status": "operator_approval_runtime_evidence_readiness_defined",
        "audit_store_handoff_readiness_status": "audit_store_handoff_readiness_defined",
        "backend_health_runtime_status": "not_created",
        "backend_health_check_status": "not_executed",
        "backend_runtime_status": "not_created",
        "resolver_implementation_status": "not_started",
        "resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
        "database_connection_provider_status": "blocked",
        "repository_mode_status": "disabled",
        "production_api_status": "not_created",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "runtime_task_card_created_in_this_slice",
        "backend_health_runtime_created_in_this_slice",
        "backend_health_check_executed_in_this_slice",
        "production_resolver_runtime_created_in_this_slice",
        "credential_payload_created_in_this_slice",
        "credential_handle_runtime_created_in_this_slice",
        "operator_approval_runtime_executed_in_this_slice",
        "audit_store_created_in_this_slice",
        "audit_writer_created_in_this_slice",
        "audit_event_written_in_this_slice",
        "cloud_secret_service_enabled",
        "database_connection_provider_enabled",
        "repository_mode_enabled",
        "production_api_enabled",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("health_boundary_contract") or {}
    require(
        contract.get("boundary_definition") == "reference_only_metadata_backend_health_boundary",
        "boundary definition drifted",
    )
    require(set(contract.get("required_inputs") or []) == EXPECTED_REQUIRED_INPUTS, "required inputs drifted")
    require(set(contract.get("allowed_metadata") or []) == EXPECTED_METADATA, "allowed metadata drifted")
    require(set(contract.get("forbidden_material") or []) == EXPECTED_FORBIDDEN_MATERIAL, "forbidden material drifted")
    require(set(contract.get("required_bindings") or []) == EXPECTED_BINDINGS, "required bindings drifted")
    require(set(contract.get("health_lifecycle") or []) == EXPECTED_LIFECYCLE, "health lifecycle drifted")
    require(set(contract.get("health_status_allowlist") or []) == EXPECTED_HEALTH_STATUS, "health status drifted")
    for field in (
        "payload_allowed",
        "secret_material_allowed",
        "provider_raw_url_allowed",
        "dsn_allowed",
        "backend_health_runtime_allowed_in_this_slice",
        "backend_health_check_allowed_in_this_slice",
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
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} missing sanitized diagnostic")
        for forbidden in ("api key", "token", "password", "dsn", "raw secret"):
            require(forbidden not in diagnostic.lower(), f"{code} diagnostic must not expose {forbidden}")


def assert_diagnostics_and_side_effects(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    require(set(diagnostics.get("allowed_fields") or []) == EXPECTED_DIAGNOSTIC_FIELDS, "diagnostic allowlist drifted")
    require(
        set(diagnostics.get("forbidden_fields") or []) == EXPECTED_FORBIDDEN_DIAGNOSTICS,
        "diagnostic forbidden fields drifted",
    )
    sample = diagnostics.get("sample_failure") or {}
    require(sample.get("failure_code") == "resolver_backend_health_boundary_missing", "sample failure code drifted")
    require(
        "secret" not in str(sample.get("sanitized_backend_diagnostic") or "").lower(),
        "sample diagnostic must stay sanitized",
    )
    for field in (
        "runtime_emission_allowed_in_this_slice",
        "secret_ref_value_allowed_in_diagnostics",
        "backend_url_allowed_in_diagnostics",
        "raw_health_response_allowed_in_diagnostics",
    ):
        require(diagnostics.get(field) is False, f"{field} must remain false")

    fallback_rules = set(fixture.get("no_fallback_rules") or [])
    for required in {
        "no fallback to fake resolver runtime",
        "no fallback to developer env plaintext",
        "no fallback to fixture credential",
        "no fallback to operator runbook text",
        "no cross-environment health boundary fallback",
        "backend health boundary does not mean production resolver runtime ready",
    }:
        require(required in fallback_rules, f"missing no fallback rule: {required}")

    side_effects = set(fixture.get("no_side_effect_policy") or [])
    for required in {
        "checker reads committed docs and fixtures only",
        "no real secret read",
        "no backend health check execution",
        "no network call",
        "no cloud secret service call",
        "no provider call",
        "no audit event write",
        "no production API call",
    }:
        require(required in side_effects, f"missing no side effect rule: {required}")

    counters = fixture.get("side_effect_counters") or {}
    for counter, value in counters.items():
        require(value == 0, f"{counter} must remain 0")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("required_static_artifacts_exist") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required static artifact missing: {relative_path}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")
    for source_path in guard.get("source_files_to_scan") or []:
        source = read(str(source_path))
        for literal in guard.get("forbidden_source_literals") or []:
            require(str(literal) not in source, f"{source_path} contains forbidden literal: {literal}")


def assert_entry_review_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("entry_review_alignment") or {}
    entry = load_json(ENTRY_REVIEW_PATH)
    gates = rows_by_id(entry, "entry_gate_matrix", "gate_id")
    for gate, status in (alignment.get("required_gate_status") or {}).items():
        require(gates[gate].get("status") == status, f"entry review gate {gate} drifted")

    blocked = set(entry.get("blocked_conditions") or [])
    for item in alignment.get("blocked_conditions_must_not_include") or []:
        require(item not in blocked, f"entry review still blocks on completed condition: {item}")

    boundary = entry.get("entry_boundary") or {}
    require(
        boundary.get("runtime_task_card_created_in_this_slice") is alignment.get("runtime_task_card_created_in_this_slice"),
        "entry review runtime task card flag drifted",
    )
    require(boundary.get("resolver_runtime_status") == "not_created", "entry review resolver runtime must remain not_created")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    implementation = load_json(IMPLEMENTATION_READINESS_PATH)
    target = implementation.get("implementation_target") or {}
    for field, value in (alignment.get("required_target_status") or {}).items():
        require(target.get(field) == value, f"implementation readiness {field} drifted")

    planned = rows_by_id(implementation, "planned_slices", "id")
    slice_info = alignment.get("planned_slice") or {}
    slice_id = str(slice_info.get("id") or "")
    require(slice_id in planned, "implementation readiness missing backend health planned slice")
    require(planned[slice_id].get("status") == slice_info.get("status"), "backend health planned slice status drifted")

    required_evidence = set(fixture.get("artifact_guard", {}).get("required_static_artifacts_exist") or [])
    missing_evidence = sorted(required_evidence - set(planned[slice_id].get("evidence") or []))
    require(not missing_evidence, f"planned backend health slice missing evidence: {missing_evidence}")
    missing_consumers = sorted(required_evidence - set(implementation.get("required_consumers") or []))
    require(not missing_consumers, f"implementation readiness missing consumers: {missing_consumers}")


def assert_runtime_split(fixture: dict[str, Any]) -> None:
    split = rows_by_id(fixture, "runtime_implementation_split", "id")
    require(set(split) == EXPECTED_RUNTIME_SPLIT, "runtime implementation split ids drifted")
    for split_id, item in split.items():
        require(item.get("status") == "future_slice_not_created", f"{split_id} must remain future_slice_not_created")


def assert_doc_references(fixture: dict[str, Any]) -> None:
    for item in fixture.get("doc_references") or []:
        path = str(item.get("path") or "")
        text = read(path)
        for needle in item.get("must_contain") or []:
            require(str(needle) in text, f"{path} missing required text: {needle}")

    validation = set(fixture.get("validation_strategy") or [])
    for required in {
        "run resolver backend health boundary readiness checker",
        "run real resolver runtime implementation entry review checker",
        "run audit store handoff readiness checker",
        "run production secret backend implementation readiness checker",
        "run production secret backend contract checker",
        "run production secret reference contract checker",
        "run fast repository check",
        "run full repository check",
    }:
        require(required in validation, f"missing validation strategy: {required}")


def assert_check_repo_registration(fixture: dict[str, Any]) -> None:
    registration = fixture.get("check_repo_registration") or {}
    script = str(registration.get("script") or "")
    after = str(registration.get("after") or "")
    before = str(registration.get("before") or "")
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    script_call = f'run_python_script("{script}", [])'
    after_call = f'run_python_script("{after}", [])'
    before_call = f'run_python_script("{before}", [])'
    require(script_call in check_repo, "check-repo.py must run backend health boundary readiness check")
    require(after_call in check_repo, "check-repo.py missing audit handoff check")
    require(before_call in check_repo, "check-repo.py missing startup supervisor check")
    require(
        check_repo.index(after_call) < check_repo.index(script_call) < check_repo.index(before_call),
        "backend health boundary readiness check must run after audit handoff and before startup supervisor",
    )


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-resolver-backend-health-boundary-readiness-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-health-boundary-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-boundary-readiness-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"backend health artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "backend health artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "backend health artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_health_boundary(fixture)
    assert_contract(fixture)
    assert_gate_matrix_and_failures(fixture)
    assert_diagnostics_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_entry_review_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_runtime_split(fixture)
    assert_doc_references(fixture)
    assert_check_repo_registration(fixture)
    assert_no_secret_literals()
    print("production ops secret backend resolver backend health boundary readiness checks passed.")


if __name__ == "__main__":
    main()
