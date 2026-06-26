#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
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
    "backend_runtime_created",
    "real_secret_read",
    "real_secret_written",
    "credential_payload_created",
    "credential_handle_created",
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
    "production-secret-backend-provider-profile-secret-binding-readiness-v1": (
        "provider_profile_secret_binding_readiness_defined"
    ),
    "production-secret-backend-config-secret-ref-readiness-v1": "config_secret_ref_readiness_defined",
    "production-secret-backend-secret-resolver-interface-disabled-readiness-v1": (
        "secret_resolver_interface_disabled_readiness_defined"
    ),
    "production-secret-backend-operator-runbook-negative-gates-readiness-v1": (
        "operator_runbook_negative_gates_readiness_defined"
    ),
    "production-secret-backend-rotation-audit-policy-readiness-v1": "rotation_audit_policy_readiness_defined",
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
}

EXPECTED_REQUIRED_FIELDS = {
    "backend_profile_id",
    "backend_kind",
    "environment",
    "provider_profile_id",
    "policy_version",
    "secret_ref_namespace",
    "allowed_secret_ref_kinds",
    "operator_approval_required",
    "audit_policy_ref",
    "rotation_policy_ref",
    "health_boundary_status",
}

EXPECTED_ALLOWED_BACKEND_KINDS = {
    "external_secret_manager_reserved",
    "operator_managed_secret_store_reserved",
}

EXPECTED_FORBIDDEN_BACKEND_KINDS = {
    "committed_secret_file",
    "env_plaintext",
    "fake_resolver",
    "mock_provider",
    "local_smoke_profile",
    "repository_memory_store",
    "database_dsn_resolver",
}

EXPECTED_GATE_STATUS = {
    "backend_kind_allowlist": "fixed_reference_only",
    "backend_profile_id_presence": "required_before_runtime",
    "environment_binding": "required_no_cross_environment",
    "provider_profile_binding": "required_reference_only",
    "secret_ref_namespace_binding": "required_reference_only",
    "policy_version_binding": "required_before_runtime",
    "operator_approval_dependency": "required_before_runtime",
    "audit_policy_dependency": "required_policy_defined",
    "rotation_policy_dependency": "required_policy_defined",
    "backend_health_boundary": "reference_only_blocked",
    "resolver_runtime_gate": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_FAILURE_CODES = {
    "resolver_backend_profile_selection_missing",
    "resolver_backend_profile_kind_unsupported",
    "resolver_backend_profile_id_missing",
    "resolver_backend_profile_environment_mismatch",
    "resolver_backend_profile_provider_mismatch",
    "resolver_backend_profile_secret_ref_namespace_missing",
    "resolver_backend_profile_policy_version_missing",
    "resolver_backend_profile_health_boundary_missing",
    "resolver_backend_profile_operator_approval_missing",
    "resolver_backend_profile_audit_policy_missing",
    "resolver_backend_profile_rotation_policy_missing",
    "resolver_backend_profile_secret_value_detected",
    "resolver_backend_profile_cloud_sdk_forbidden",
    "resolver_backend_profile_runtime_created_forbidden",
    "resolver_backend_profile_repository_mode_forbidden",
    "resolver_backend_profile_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "resolver_backend_profile_selection_status",
    "backend_profile_id",
    "backend_kind",
    "environment",
    "provider_profile_id",
    "policy_version",
    "secret_ref_namespace_status",
    "health_boundary_status",
    "operator_approval_status",
    "audit_policy_status",
    "rotation_policy_status",
    "failure_code",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
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
    "full_user_claim",
    "authorization_header",
    "cookie",
}

EXPECTED_NO_FALLBACK = {
    "backend profile missing fails closed",
    "unsupported backend kind fails closed",
    "policy version missing fails closed",
    "backend health boundary missing fails closed",
    "no fallback from production backend profile to test backend profile",
    "no fallback from test backend profile to production backend profile",
    "no fallback from production backend profile to fake resolver",
    "no fallback from fake resolver profile to production backend profile",
    "no fallback to mock provider",
    "no fallback to local-smoke profile",
    "no fallback to developer env plaintext",
    "no fallback to fixture credential",
    "no fallback to committed secret value",
    "no fallback to repository memory store",
    "no fallback to sample",
    "provider profile binding missing does not resolve credential",
    "backend profile selection does not mean production resolver runtime ready",
}

EXPECTED_NO_SIDE_EFFECTS = {
    "checker reads committed docs and fixtures only",
    "no real secret read",
    "no environment secret read",
    "no cloud secret service call",
    "no provider call",
    "no backend health check call",
    "no fake resolver call",
    "no secret resolver call",
    "no credential payload creation",
    "no credential handle runtime creation",
    "no database connection",
    "no driver open",
    "no SQL execution",
    "no schema marker read",
    "no schema marker write",
    "no repository mode enablement",
    "no audit store write",
    "no production API call",
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
        fixture.get("kind") == "production_ops_secret_backend_resolver_backend_profile_selection_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "production-secret-backend-resolver-backend-profile-selection-readiness-v1", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "resolver_backend_profile_selection_readiness_defined", "unexpected status")
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md",
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
        require(evidence, f"{dependency_id} evidence is required")
        require((REPO_ROOT / evidence).exists(), f"{dependency_id} evidence missing: {evidence}")

    reference = load_json(SECRET_REFERENCE_PATH)
    require(reference.get("scope") == "secret_reference_only", "secret reference must remain reference-only")
    policy = reference.get("policy") or {}
    for field in ("stores_secret_values", "resolver_enabled", "cloud_calls_allowed", "production_secret_backend_ready"):
        require(policy.get(field) is False, f"secret reference policy {field} must remain false")


def assert_selection_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("selection_boundary") or {}
    for field, expected in {
        "selection_status": "defined_without_backend_runtime",
        "resolver_backend_profile_selection_readiness_status": "resolver_backend_profile_selection_readiness_defined",
        "entry_review_status": "blocked_before_runtime_task_card",
        "real_resolver_runtime_preconditions_status": "real_resolver_runtime_preconditions_defined",
        "resolver_implementation_status": "not_started",
        "resolver_runtime_status": "not_created",
        "backend_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
        "cloud_secret_service_status": "not_selected",
        "database_connection_provider_status": "blocked",
        "repository_mode_status": "disabled",
        "production_api_status": "not_created",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "runtime_task_card_created_in_this_slice",
        "production_resolver_runtime_created_in_this_slice",
        "cloud_secret_service_enabled",
        "database_connection_provider_enabled",
        "credential_payload_created",
        "credential_handle_runtime_created",
        "repository_mode_enabled",
        "production_api_enabled",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_backend_profile_shape(fixture: dict[str, Any]) -> None:
    shape = fixture.get("backend_profile_shape") or {}
    missing_fields = sorted(EXPECTED_REQUIRED_FIELDS - set(shape.get("required_fields") or []))
    require(not missing_fields, f"missing backend profile required fields: {missing_fields}")
    missing_allowed = sorted(EXPECTED_ALLOWED_BACKEND_KINDS - set(shape.get("allowed_backend_kinds") or []))
    require(not missing_allowed, f"missing allowed backend kinds: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_BACKEND_KINDS - set(shape.get("forbidden_backend_kinds") or []))
    require(not missing_forbidden, f"missing forbidden backend kinds: {missing_forbidden}")
    require(shape.get("selection_mode") == "static_profile_manifest_reference_only", "selection mode drifted")
    for field in (
        "environment_cross_fallback_allowed",
        "provider_profile_cross_fallback_allowed",
        "secret_ref_value_allowed",
        "backend_health_check_allowed_in_this_slice",
    ):
        require(shape.get(field) is False, f"{field} must remain false")


def assert_gate_matrix_and_failures(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "selection_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "selection gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    failures = rows_by_id(fixture, "failure_mapping", "code")
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(failures))
    require(not missing_failures, f"missing failure codes: {missing_failures}")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} must define failure boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} must define sanitized diagnostic")
        for forbidden in ("api key", "token", "password", "dsn", "raw secret"):
            require(forbidden not in diagnostic.lower(), f"{code} diagnostic must not expose {forbidden}")


def assert_diagnostics_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    missing_allowed = sorted(EXPECTED_DIAGNOSTIC_FIELDS - set(diagnostics.get("allowed_fields") or []))
    require(not missing_allowed, f"missing diagnostic fields: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_DIAGNOSTICS - set(diagnostics.get("forbidden_fields") or []))
    require(not missing_forbidden, f"missing forbidden diagnostic fields: {missing_forbidden}")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must remain disabled")
    require(diagnostics.get("secret_ref_value_allowed_in_diagnostics") is False, "secret ref value must not be emitted")
    require(diagnostics.get("backend_url_allowed_in_diagnostics") is False, "backend URL must not be emitted")

    missing_fallback = sorted(EXPECTED_NO_FALLBACK - set(fixture.get("no_fallback_policy") or []))
    require(not missing_fallback, f"missing no fallback policy: {missing_fallback}")
    missing_side_effects = sorted(EXPECTED_NO_SIDE_EFFECTS - set(fixture.get("no_side_effect_policy") or []))
    require(not missing_side_effects, f"missing no side effects policy: {missing_side_effects}")
    for counter, value in (fixture.get("side_effect_counters") or {}).items():
        require(value == 0, f"{counter} must stay zero")


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


def assert_implementation_readiness_alignment() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    require(
        target.get("resolver_backend_profile_selection_readiness_status") == "defined_without_backend_runtime",
        "implementation readiness must record backend profile selection readiness",
    )
    require(
        target.get("real_resolver_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "implementation readiness must keep real resolver runtime entry review blocked",
    )
    require(target.get("resolver_implementation_status") == "not_started", "resolver implementation must remain not_started")
    require(target.get("resolver_runtime_status") == "not_created", "resolver runtime must remain not_created")
    require(target.get("production_secret_backend_status") == "not_satisfied", "production secret backend must remain not_satisfied")

    planned = rows_by_id(readiness, "planned_slices", "id")
    current = planned.get("resolver-backend-profile-selection-readiness") or {}
    require(
        current.get("status") == "resolver_backend_profile_selection_readiness_defined",
        "planned backend profile selection status drifted",
    )
    required_evidence = {
        "docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
        "scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py",
    }
    missing_evidence = sorted(required_evidence - set(current.get("evidence") or []))
    require(not missing_evidence, f"planned backend profile selection missing evidence: {missing_evidence}")

    missing_consumers = sorted(required_evidence - set(readiness.get("required_consumers") or []))
    require(not missing_consumers, f"implementation readiness missing consumers: {missing_consumers}")


def assert_docs_validation_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        text = read(path)
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in text]
        require(not missing, f"{path} missing literals: {missing}")

    expected_validation = {
        "run resolver backend profile selection readiness checker",
        "run real resolver runtime implementation entry review checker",
        "run real resolver runtime preconditions checker",
        "run provider profile secret binding checker",
        "run secret resolver interface disabled readiness checker",
        "run production secret backend implementation readiness checker",
        "run production secret backend contract checker",
        "run production secret reference contract checker",
        "run fast repository check",
        "run full repository check",
    }
    missing_validation = sorted(expected_validation - set(fixture.get("validation_strategy") or []))
    require(not missing_validation, f"missing validation strategy entries: {missing_validation}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    prior_call = 'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py", [])'
    current_call = 'run_python_script("check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py", [])'
    startup_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call in (prior_call, current_call, startup_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(prior_call) < check_repo.index(current_call) < check_repo.index(startup_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"backend profile selection artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "backend profile selection artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "backend profile selection artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_selection_boundary(fixture)
    assert_backend_profile_shape(fixture)
    assert_gate_matrix_and_failures(fixture)
    assert_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment()
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend resolver backend profile selection readiness checks passed.")


if __name__ == "__main__":
    main()
