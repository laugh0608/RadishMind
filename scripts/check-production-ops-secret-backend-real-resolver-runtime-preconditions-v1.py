#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json"
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
    "real_secret_read",
    "real_secret_written",
    "credential_payload_created",
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
    "production-secret-backend-fake-resolver-runtime-implementation-v1": "fake_resolver_runtime_test_only_implemented",
    "production-secret-backend-secret-resolver-interface-disabled-readiness-v1": "secret_resolver_interface_disabled_readiness_defined",
    "production-secret-backend-provider-profile-secret-binding-readiness-v1": "provider_profile_secret_binding_readiness_defined",
    "production-secret-backend-operator-runbook-negative-gates-readiness-v1": "operator_runbook_negative_gates_readiness_defined",
    "production-secret-backend-rotation-audit-policy-readiness-v1": "rotation_audit_policy_readiness_defined",
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
    "workflow-saved-draft-database-secret-resolver-readiness-v1": "draft_database_secret_resolver_readiness_defined",
    "workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1": "draft_database_secret_resolver_implementation_entry_review_defined",
}

EXPECTED_GATE_IDS = {
    "implementation_entry_review",
    "secret_ref_contract",
    "provider_profile_binding",
    "environment_binding",
    "operator_approval",
    "audit_policy",
    "rotation_policy",
    "resolver_backend_profile",
    "no_secret_leakage_gate",
    "artifact_guard",
}

EXPECTED_INPUT_FIELDS = {
    "environment",
    "provider",
    "provider_profile",
    "credential_requirement",
    "secret_ref_key",
    "secret_ref_version_ref",
    "caller_purpose",
    "request_id",
    "audit_ref",
    "operator_approval_ref",
    "policy_version",
    "rotation_policy_version",
    "backend_profile_ref",
}

EXPECTED_RESULT_FIELDS = {
    "resolver_state",
    "credential_handle_present",
    "credential_handle_id",
    "credential_kind",
    "environment",
    "provider",
    "provider_profile",
    "secret_ref_version_ref",
    "request_id",
    "audit_ref",
    "policy_version",
    "failure_code",
    "sanitized_diagnostic",
}

EXPECTED_FAILURE_CODES = {
    "real_resolver_runtime_preconditions_missing",
    "real_resolver_runtime_task_not_approved",
    "real_resolver_backend_profile_missing",
    "real_resolver_secret_ref_missing",
    "real_resolver_provider_profile_binding_missing",
    "real_resolver_environment_binding_missing",
    "real_resolver_operator_approval_missing",
    "real_resolver_audit_policy_missing",
    "real_resolver_rotation_policy_missing",
    "real_resolver_no_leakage_gate_missing",
    "real_resolver_backend_unavailable",
    "real_resolver_resolution_denied",
    "real_resolver_secret_value_exposure_detected",
    "real_resolver_fallback_forbidden",
    "real_resolver_side_effect_forbidden",
    "real_resolver_artifact_guard_violation",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "real_resolver_runtime_preconditions_status",
    "resolver_state",
    "resolver_backend_profile_status",
    "secret_ref_status",
    "secret_ref_version_status",
    "provider_profile_binding_status",
    "environment_binding_status",
    "operator_approval_status",
    "audit_policy_status",
    "rotation_policy_status",
    "no_secret_leakage_gate_status",
    "credential_handle_present",
    "failure_code",
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
    "full_user_claim",
    "authorization_header",
    "cookie",
}

EXPECTED_NO_FALLBACK = {
    "no fallback from production resolver to fake resolver",
    "no fallback from fake resolver to production resolver",
    "no fallback from resolver failure to RADISHMIND_PLATFORM_API_KEY",
    "no fallback to developer env plaintext",
    "no fallback to fixture credential",
    "no fallback to committed secret value",
    "no fallback to mock provider",
    "no fallback to local-smoke profile",
    "no fallback to fake query executor",
    "no fallback to sample",
    "no fallback from test secret_ref to production secret_ref",
    "no fallback from production secret_ref to test secret_ref",
    "secret_ref present does not mean credential resolved",
    "test-only fake resolver runtime does not mean production resolver runtime",
}

EXPECTED_NO_SIDE_EFFECTS = {
    "precondition checker reads committed artifacts only",
    "no real secret read",
    "no environment secret read",
    "no cloud secret service call",
    "no provider call",
    "no production resolver call",
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
        fixture.get("kind") == "production_ops_secret_backend_real_resolver_runtime_preconditions_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "production-secret-backend-real-resolver-runtime-preconditions-v1", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "real_resolver_runtime_preconditions_defined", "unexpected status")
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md",
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


def assert_precondition_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("precondition_boundary") or {}
    for field, expected in {
        "precondition_status": "defined_runtime_not_opened",
        "real_resolver_runtime_preconditions_status": "real_resolver_runtime_preconditions_defined",
        "resolver_implementation_status": "not_started",
        "resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
        "fake_resolver_runtime_status": "implemented_test_only_disabled_by_default",
        "cloud_secret_service_status": "not_selected",
        "database_connection_provider_status": "blocked",
        "repository_mode_status": "disabled",
        "production_api_status": "not_created",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "production_resolver_runtime_created_in_this_slice",
        "reads_real_secret",
        "calls_cloud_secret_service",
        "creates_credential_payload",
        "creates_credential_handle_runtime",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gates_and_contract(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "enablement_preconditions", "gate_id")
    require(set(gates) == EXPECTED_GATE_IDS, "enablement gate ids drifted")
    for gate_id, item in gates.items():
        status = str(item.get("status") or "")
        require(
            status.startswith("required_before_") or status.startswith("satisfied_"),
            f"{gate_id} has unexpected status",
        )
        require(len(item.get("must_define") or []) >= 3, f"{gate_id} must define concrete requirements")

    contract = fixture.get("future_runtime_contract") or {}
    missing_inputs = sorted(EXPECTED_INPUT_FIELDS - set(contract.get("allowed_input_fields") or []))
    require(not missing_inputs, f"missing input fields: {missing_inputs}")
    missing_results = sorted(EXPECTED_RESULT_FIELDS - set(contract.get("allowed_result_fields") or []))
    require(not missing_results, f"missing result fields: {missing_results}")
    for forbidden in ("secret value", "credential payload", "provider raw URL", "DSN"):
        require(forbidden in set(contract.get("must_not_include") or []), f"future contract missing forbidden item: {forbidden}")


def assert_failure_mapping_and_diagnostics(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(failures))
    require(not missing_failures, f"missing failure codes: {missing_failures}")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} must define failure boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} must define sanitized diagnostic")
        lower = diagnostic.lower()
        for forbidden in ("api key", "token", "password", "dsn", "raw secret"):
            require(forbidden not in lower, f"{code} diagnostic must not expose {forbidden}")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    missing_allowed = sorted(EXPECTED_DIAGNOSTIC_FIELDS - set(diagnostics.get("allowed_fields") or []))
    require(not missing_allowed, f"missing diagnostic fields: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_DIAGNOSTICS - set(diagnostics.get("forbidden_fields") or []))
    require(not missing_forbidden, f"missing forbidden diagnostic fields: {missing_forbidden}")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must remain disabled")
    require(
        diagnostics.get("secret_ref_value_allowed_in_runtime_diagnostics") is False,
        "secret ref value must not appear in diagnostics",
    )


def assert_policies(fixture: dict[str, Any]) -> None:
    missing_fallback = sorted(EXPECTED_NO_FALLBACK - set(fixture.get("no_fallback_policy") or []))
    require(not missing_fallback, f"missing no fallback policy: {missing_fallback}")
    missing_side_effects = sorted(EXPECTED_NO_SIDE_EFFECTS - set(fixture.get("no_side_effect_policy") or []))
    require(not missing_side_effects, f"missing no side effects policy: {missing_side_effects}")
    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters are required")
    for counter, value in counters.items():
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
        target.get("real_resolver_runtime_preconditions_status") == "real_resolver_runtime_preconditions_defined",
        "implementation readiness must record real resolver runtime preconditions",
    )
    for field, expected in {
        "resolver_implementation_status": "not_started",
        "resolver_runtime_status": "not_created",
        "fake_resolver_runtime_status": "implemented_test_only_disabled_by_default",
        "production_secret_backend_status": "not_satisfied",
    }.items():
        if field in target:
            require(target.get(field) == expected, f"implementation target {field} drifted")

    planned = rows_by_id(readiness, "planned_slices", "id")
    current = planned.get("real-resolver-runtime-preconditions") or {}
    require(current.get("status") == "real_resolver_runtime_preconditions_defined", "planned preconditions status drifted")
    required_evidence = {
        "docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json",
        "scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py",
    }
    missing_evidence = sorted(required_evidence - set(current.get("evidence") or []))
    require(not missing_evidence, f"planned preconditions missing evidence: {missing_evidence}")

    missing_consumers = sorted(required_evidence - set(readiness.get("required_consumers") or []))
    require(not missing_consumers, f"implementation readiness missing consumers: {missing_consumers}")


def assert_docs_validation_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        text = read(path)
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in text]
        require(not missing, f"{path} missing literals: {missing}")

    expected_validation = {
        "run real resolver runtime preconditions checker",
        "run fake resolver runtime implementation checker",
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
    prior_call = 'run_python_script("check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py", [])'
    current_call = 'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py", [])'
    startup_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call in (prior_call, current_call, startup_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(prior_call) < check_repo.index(current_call) < check_repo.index(startup_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"real resolver preconditions artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "real resolver preconditions artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "real resolver preconditions artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_precondition_boundary(fixture)
    assert_gates_and_contract(fixture)
    assert_failure_mapping_and_diagnostics(fixture)
    assert_policies(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment()
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend real resolver runtime preconditions checks passed.")


if __name__ == "__main__":
    main()
