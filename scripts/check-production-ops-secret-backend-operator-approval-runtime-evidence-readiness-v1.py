#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json"
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
    "operator_approval_runtime_created",
    "operator_approval_runtime_executed",
    "operator_approval_executor_created",
    "operator_identity_provider_connected",
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
    "production-secret-backend-operator-runbook-negative-gates-readiness-v1": (
        "operator_runbook_negative_gates_readiness_defined"
    ),
    "production-secret-backend-rotation-audit-policy-readiness-v1": "rotation_audit_policy_readiness_defined",
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "resolver_backend_profile_selection_readiness_defined"
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1": (
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
    ),
    "production-secret-backend-credential-handle-runtime-boundary-readiness-v1": (
        "credential_handle_runtime_boundary_readiness_defined"
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
    "approval_evidence_id",
    "approval_evidence_kind",
    "approval_status",
    "approval_lifecycle_state",
    "approval_subject",
    "environment",
    "provider",
    "provider_profile",
    "secret_ref_key",
    "secret_ref_version_ref",
    "backend_profile_ref",
    "credential_handle_boundary_ref",
    "requested_by_ref",
    "approved_by_ref",
    "approval_ticket_ref",
    "approval_window_ref",
    "runbook_version",
    "policy_version",
    "rotation_policy_version",
    "expires_at",
    "revoked_at",
    "request_id",
    "audit_ref",
    "failure_code",
    "sanitized_diagnostic",
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
}

EXPECTED_BINDINGS = {
    "approval_subject",
    "runbook_version",
    "secret_ref_key",
    "secret_ref_version_ref",
    "provider_profile",
    "environment",
    "backend_profile_ref",
    "credential_handle_boundary_ref",
    "no_secret_leakage_strategy_ref",
    "requested_by_ref",
    "approved_by_ref",
    "approval_ticket_ref",
    "approval_window_ref",
    "audit_ref",
    "policy_version",
    "rotation_policy_version",
}

EXPECTED_LIFECYCLE = {
    "evidence_planned",
    "request_metadata_bound",
    "approval_pending",
    "approved_metadata_only",
    "denied",
    "expired",
    "revoked",
    "revalidation_required",
    "execution_blocked",
    "resolution_failed_closed",
}

EXPECTED_GATE_STATUS = {
    "approval_evidence_schema": "defined_static_only",
    "metadata_allowlist": "fixed_sanitized_only",
    "payload_and_secret_material": "forbidden",
    "runbook_binding": "required_before_runtime",
    "secret_ref_binding": "required_before_runtime",
    "provider_profile_binding": "required_before_runtime",
    "environment_binding": "required_no_cross_environment",
    "backend_profile_binding": "required_before_runtime",
    "credential_handle_boundary_binding": "required_before_runtime",
    "no_leakage_strategy_binding": "required_before_runtime",
    "audit_rotation_dependency": "required_before_runtime",
    "operator_identity_reference": "required_reference_only",
    "dual_control": "required_before_runtime",
    "ticket_and_change_window": "required_before_runtime",
    "approval_lifecycle_allowlist": "defined_static_only",
    "operator_approval_runtime": "not_created",
    "operator_approval_runtime_execution": "not_executed",
    "production_resolver_runtime": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "audit_store_gate": "blocked",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_FAILURE_CODES = {
    "operator_approval_runtime_evidence_missing",
    "operator_approval_evidence_schema_missing",
    "operator_approval_metadata_allowlist_missing",
    "operator_approval_payload_detected",
    "operator_approval_secret_material_detected",
    "operator_approval_subject_binding_missing",
    "operator_approval_runbook_binding_missing",
    "operator_approval_secret_ref_binding_missing",
    "operator_approval_provider_profile_binding_missing",
    "operator_approval_environment_binding_mismatch",
    "operator_approval_backend_profile_binding_missing",
    "operator_approval_credential_boundary_missing",
    "operator_approval_no_leakage_strategy_missing",
    "operator_approval_audit_dependency_missing",
    "operator_approval_rotation_dependency_missing",
    "operator_approval_identity_reference_missing",
    "operator_approval_dual_control_missing",
    "operator_approval_ticket_missing",
    "operator_approval_change_window_missing",
    "operator_approval_lifecycle_state_invalid",
    "operator_approval_diagnostic_exposure_detected",
    "operator_approval_runtime_created_forbidden",
    "operator_approval_runtime_executed_forbidden",
    "operator_approval_side_effect_forbidden",
    "operator_approval_fallback_forbidden",
    "operator_approval_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "operator_approval_boundary_status",
    "operator_approval_runtime_status",
    "operator_approval_execution_status",
    "approval_evidence_status",
    "approval_lifecycle_state",
    "approval_subject_status",
    "environment",
    "provider",
    "provider_profile",
    "secret_ref_key_status",
    "secret_ref_version_ref_status",
    "backend_profile_ref_status",
    "credential_handle_boundary_status",
    "no_secret_leakage_strategy_status",
    "operator_identity_status",
    "dual_control_status",
    "approval_ticket_status",
    "approval_window_status",
    "audit_policy_status",
    "rotation_policy_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_NO_FALLBACK = {
    "operator approval boundary missing keeps real resolver runtime task card blocked",
    "no fallback from approval boundary to runbook text",
    "no fallback to developer env plaintext",
    "no fallback to fixture credential",
    "no fallback to committed secret value",
    "no fallback to sample",
    "no fallback to mock provider",
    "no fallback to local-smoke profile",
    "no fallback to fake resolver runtime",
    "no fallback to repository memory store",
    "no fallback to previously approved test evidence",
    "no fallback from production approval to test approval",
    "no fallback from test approval to production approval",
    "approval boundary does not mean approval runtime executed",
    "approval boundary does not mean credential resolved",
    "approval boundary does not mean production resolver runtime ready",
}

EXPECTED_NO_SIDE_EFFECTS = {
    "checker reads committed docs and fixtures only",
    "no real secret read",
    "no environment secret read",
    "no cloud secret service call",
    "no provider call",
    "no fake resolver call",
    "no production resolver call",
    "no approval runtime creation",
    "no approval runtime execution",
    "no operator identity provider call",
    "no smoke runtime execution",
    "no credential payload creation",
    "no credential handle creation",
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
        fixture.get("kind") == "production_ops_secret_backend_operator_approval_runtime_evidence_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-operator-approval-runtime-evidence-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "operator_approval_runtime_evidence_readiness_defined",
        "unexpected status",
    )
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-operator-approval-runtime-evidence-readiness-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md",
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


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("approval_boundary") or {}
    for field, expected in {
        "boundary_status": "defined_without_runtime_execution",
        "operator_approval_runtime_evidence_status": "operator_approval_runtime_evidence_readiness_defined",
        "entry_review_status": "blocked_before_runtime_task_card",
        "operator_runbook_status": "operator_runbook_negative_gates_readiness_defined",
        "rotation_audit_policy_status": "rotation_audit_policy_readiness_defined",
        "resolver_backend_profile_selection_readiness_status": "resolver_backend_profile_selection_readiness_defined",
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_status": (
            "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
        ),
        "credential_handle_runtime_boundary_status": "credential_handle_runtime_boundary_readiness_defined",
        "operator_approval_runtime_status": "not_created",
        "operator_approval_execution_status": "not_executed",
        "operator_approval_executor_status": "not_created",
        "operator_identity_provider_status": "not_connected",
        "credential_handle_runtime_status": "not_created",
        "credential_payload_status": "forbidden",
        "resolver_implementation_status": "not_started",
        "resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
        "cloud_secret_service_status": "not_selected",
        "database_connection_provider_status": "blocked",
        "repository_mode_status": "disabled",
        "production_api_status": "not_created",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "runtime_task_card_created_in_this_slice",
        "operator_approval_runtime_created_in_this_slice",
        "operator_approval_runtime_executed_in_this_slice",
        "operator_approval_executor_created_in_this_slice",
        "operator_identity_provider_connected_in_this_slice",
        "production_resolver_runtime_created_in_this_slice",
        "credential_payload_created_in_this_slice",
        "credential_handle_runtime_created_in_this_slice",
        "audit_store_written_in_this_slice",
        "cloud_secret_service_enabled",
        "database_connection_provider_enabled",
        "repository_mode_enabled",
        "production_api_enabled",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_approval_evidence_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("approval_evidence_contract") or {}
    require(
        contract.get("approval_evidence_definition") == "reference_only_metadata_approval_evidence",
        "approval evidence definition drifted",
    )
    for key, expected in {
        "allowed_metadata": EXPECTED_METADATA,
        "forbidden_material": EXPECTED_FORBIDDEN_MATERIAL,
        "required_bindings": EXPECTED_BINDINGS,
        "lifecycle_states": EXPECTED_LIFECYCLE,
    }.items():
        missing = sorted(expected - set(contract.get(key) or []))
        require(not missing, f"{key} missing entries: {missing}")
    for field in (
        "payload_allowed",
        "secret_material_allowed",
        "runtime_creation_allowed_in_this_slice",
        "runtime_execution_allowed_in_this_slice",
        "credential_resolution_allowed_in_this_slice",
        "self_approval_allowed",
    ):
        require(contract.get(field) is False, f"{field} must remain false")


def assert_gates_and_failures(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "boundary_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "boundary gate ids drifted")
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
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_MATERIAL - set(diagnostics.get("forbidden_fields") or []))
    require(not missing_forbidden, f"missing forbidden diagnostic fields: {missing_forbidden}")
    for field in (
        "runtime_emission_allowed_in_this_slice",
        "secret_ref_value_allowed_in_diagnostics",
        "credential_payload_allowed_in_diagnostics",
        "full_handle_allowed_in_diagnostics",
        "full_operator_claim_allowed_in_diagnostics",
    ):
        require(diagnostics.get(field) is False, f"{field} must remain false")

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
        target.get("operator_approval_runtime_evidence_readiness_status") == "defined_without_runtime_execution",
        "implementation readiness must record operator approval runtime evidence readiness",
    )
    require(
        target.get("operator_approval_runtime_status") == "not_created",
        "implementation readiness must keep operator approval runtime not_created",
    )
    require(
        target.get("operator_approval_runtime_execution_status") == "not_executed",
        "implementation readiness must keep operator approval runtime not_executed",
    )
    require(
        target.get("real_resolver_runtime_implementation_entry_review_status") == "blocked_before_runtime_task_card",
        "implementation readiness must keep real resolver runtime entry review blocked",
    )
    require(target.get("resolver_implementation_status") == "not_started", "resolver implementation must remain not_started")
    require(target.get("resolver_runtime_status") == "not_created", "resolver runtime must remain not_created")
    require(target.get("production_secret_backend_status") == "not_satisfied", "production secret backend must remain not_satisfied")

    planned = rows_by_id(readiness, "planned_slices", "id")
    current = planned.get("operator-approval-runtime-evidence-readiness") or {}
    require(
        current.get("status") == "operator_approval_runtime_evidence_readiness_defined",
        "planned operator approval runtime evidence status drifted",
    )
    required_evidence = {
        "docs/platform/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md",
        "docs/task-cards/production-secret-backend-operator-approval-runtime-evidence-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json",
        "scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py",
    }
    missing_evidence = sorted(required_evidence - set(current.get("evidence") or []))
    require(not missing_evidence, f"planned operator approval runtime evidence missing evidence: {missing_evidence}")

    missing_consumers = sorted(required_evidence - set(readiness.get("required_consumers") or []))
    require(not missing_consumers, f"implementation readiness missing consumers: {missing_consumers}")


def assert_entry_review_alignment() -> None:
    entry = load_json(ENTRY_REVIEW_PATH)
    gates = rows_by_id(entry, "entry_gate_matrix", "gate_id")
    operator = gates.get("operator_approval_runtime_evidence") or {}
    require(
        operator.get("status") == "readiness_defined_runtime_not_executed",
        "entry review must record operator approval evidence defined while runtime remains not_executed",
    )
    blocked = set(entry.get("blocked_conditions") or [])
    require(
        "operator-approval-runtime-evidence-readiness" not in blocked,
        "entry review must not keep completed operator approval readiness as a missing blocker",
    )
    require("production-secret-audit-store-handoff-readiness" in blocked, "audit store handoff blocker must remain")
    require("resolver-backend-health-boundary-readiness" in blocked, "backend health blocker must remain")


def assert_docs_validation_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        text = read(path)
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in text]
        require(not missing, f"{path} missing literals: {missing}")

    expected_validation = {
        "run operator approval runtime evidence readiness checker",
        "run real resolver runtime implementation entry review checker",
        "run credential handle runtime boundary readiness checker",
        "run real resolver no leakage smoke runtime strategy checker",
        "run resolver backend profile selection readiness checker",
        "run real resolver runtime preconditions checker",
        "run production secret backend implementation readiness checker",
        "run production secret backend contract checker",
        "run production secret reference contract checker",
        "run fast repository check",
        "run full repository check",
    }
    missing_validation = sorted(expected_validation - set(fixture.get("validation_strategy") or []))
    require(not missing_validation, f"missing validation strategy entries: {missing_validation}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    prior_call = (
        'run_python_script("check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py", [])'
    )
    current_call = (
        'run_python_script("check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py", [])'
    )
    next_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call in (prior_call, current_call, next_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(prior_call) < check_repo.index(current_call) < check_repo.index(next_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md",
        "docs/task-cards/production-secret-backend-operator-approval-runtime-evidence-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"operator approval artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "operator approval artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "operator approval artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_approval_evidence_contract(fixture)
    assert_gates_and_failures(fixture)
    assert_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment()
    assert_entry_review_alignment()
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend operator approval runtime evidence readiness checks passed.")


if __name__ == "__main__":
    main()
