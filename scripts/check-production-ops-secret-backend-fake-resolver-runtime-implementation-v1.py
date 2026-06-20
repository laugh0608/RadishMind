#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json"
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
ENTRY_REVIEW_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json"
IMPLEMENTATION_TASK_CARD_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json"
SECRET_REFERENCE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-reference-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "production_ready",
    "production_secret_backend_ready",
    "cloud_secret_service_ready",
    "resolver_implemented",
    "resolver_runtime_ready",
    "fake_resolver_implemented",
    "fake_resolver_runtime_ready",
    "fake_resolver_runtime_created",
    "no_secret_leakage_smoke_runtime_ready",
    "test_fixture_strategy_satisfied",
    "real_secret_written",
    "credential_resolved",
    "credential_handle_created",
    "secret_resolver_ready",
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
    "production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1": "fake_resolver_runtime_implementation_entry_review_defined",
    "production-secret-backend-fake-resolver-implementation-v1": "fake_resolver_implementation_task_card_defined",
    "production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1": "fake_resolver_contract_no_secret_leakage_smoke_strategy_defined",
    "production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1": "test_fixture_strategy_fake_resolver_entry_review_defined",
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
    "workflow-saved-draft-database-secret-resolver-readiness-v1": "draft_database_secret_resolver_readiness_defined",
    "workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1": "draft_database_secret_resolver_implementation_entry_review_defined",
}

EXPECTED_GATE_STATUS = {
    "prior_runtime_entry_review_consumed": "satisfied_ready_for_task_card",
    "implementation_task_card_created": "created_static_task_card",
    "disabled_by_default_runtime_gate": "defined_for_future_runtime",
    "placeholder_secret_ref_fixture": "defined_for_future_runtime",
    "environment_binding": "defined_for_future_runtime",
    "opaque_handle_metadata": "defined_for_future_runtime",
    "sanitized_diagnostics_runtime": "defined_for_future_runtime",
    "no_secret_leakage_runtime_smoke": "defined_for_future_runtime",
    "side_effect_counters": "defined_for_future_runtime",
    "artifact_guard_gate": "satisfied_for_task_card",
    "fake_resolver_runtime_gate": "not_opened",
    "no_secret_leakage_smoke_runtime_gate": "not_opened",
    "resolver_runtime_gate": "forbidden",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_runtime_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_RUNTIME_SCOPE = {
    "test-only fake resolver runtime gate",
    "disabled-by-default runtime state",
    "placeholder secret ref fixture",
    "environment binding",
    "opaque test credential handle metadata",
    "sanitized diagnostics runtime emission",
    "no secret leakage runtime smoke",
    "offline fast baseline",
    "artifact guard",
    "side effect counters",
}

EXPECTED_ALLOWED_INPUT_FIELDS = {
    "environment",
    "provider",
    "provider_profile",
    "secret_ref_key",
    "secret_ref_version",
    "purpose",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_SUCCESS_FIELDS = {
    "credential_handle_id",
    "credential_kind",
    "environment",
    "provider",
    "provider_profile",
    "secret_ref_key",
    "secret_ref_version",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_FAILURE_FIELDS = {
    "failure_code",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_MUST_NOT_INCLUDE = {
    "production resolver runtime",
    "cloud secret service",
    "real credential",
    "database connection provider",
    "database driver",
    "connection factory",
    "SQL",
    "schema marker",
    "migration runner",
    "repository mode runtime",
    "production secret audit store",
    "public production API",
}

EXPECTED_FAILURE_CODES = {
    "fake_resolver_runtime_task_card_missing",
    "fake_resolver_runtime_task_card_entry_review_missing",
    "fake_resolver_runtime_disabled_gate_missing",
    "fake_resolver_runtime_placeholder_fixture_missing",
    "fake_resolver_runtime_environment_binding_missing",
    "fake_resolver_runtime_opaque_handle_metadata_missing",
    "fake_resolver_runtime_diagnostics_boundary_missing",
    "fake_resolver_runtime_no_leakage_smoke_missing",
    "fake_resolver_runtime_secret_value_detected",
    "fake_resolver_runtime_created_in_task_card",
    "fake_resolver_runtime_cloud_call_forbidden",
    "fake_resolver_runtime_repository_mode_forbidden",
    "fake_resolver_runtime_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "fake_resolver_runtime_implementation_task_card_status",
    "fake_resolver_runtime_implementation_status",
    "fake_resolver_contract_status",
    "no_secret_leakage_smoke_strategy_status",
    "test_fixture_strategy_status",
    "resolver_state",
    "fake_resolver_runtime_status",
    "no_secret_leakage_smoke_runtime_status",
    "secret_ref_status",
    "environment_binding_status",
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
    "full_user_claim",
}

EXPECTED_NO_FALLBACK = {
    "no fallback from fake resolver to production resolver",
    "no fallback from production resolver to fake resolver",
    "no fallback from test secret_ref to production secret_ref",
    "no fallback to RADISHMIND_PLATFORM_API_KEY",
    "no fallback to mock provider",
    "no fallback to local-smoke profile",
    "no fallback to fixture credential",
    "no fallback to committed secret value",
    "no fallback to developer env plaintext",
    "no fallback to fake query executor",
    "runtime implementation task card does not mean credential resolved",
}

EXPECTED_NO_SIDE_EFFECTS = {
    "checker reads committed docs and fixtures only",
    "no environment secret read",
    "no cloud secret service call",
    "no provider call",
    "no fake resolver call",
    "no secret resolver call",
    "no credential handle creation",
    "no database connection",
    "no driver open",
    "no SQL execution",
    "no schema marker read",
    "no schema marker write",
    "no repository mode enablement",
    "no production API call",
    "no file write",
}

EXPECTED_VALIDATION = {
    "run fake resolver runtime implementation checker",
    "run fake resolver runtime implementation entry review checker",
    "run fake resolver implementation checker",
    "run fake resolver contract no secret leakage smoke strategy checker",
    "run test fixture strategy fake resolver entry review checker",
    "run production secret backend implementation readiness checker",
    "run production secret backend contract checker",
    "run production secret reference contract checker",
    "run workflow saved draft database secret resolver readiness checker",
    "run workflow saved draft database secret resolver implementation entry review checker",
    "run fast repository check",
    "run full repository check",
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
        fixture.get("kind") == "production_ops_secret_backend_fake_resolver_runtime_implementation_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "production-secret-backend-fake-resolver-runtime-implementation-v1", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "fake_resolver_runtime_implementation_task_card_defined", "unexpected status")
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md",
    }.items():
        value = str(slice_info.get(key) or "")
        require(value == expected_path, f"unexpected {key}")
        require((REPO_ROOT / value).exists(), f"{key} path missing: {value}")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = {str(item.get("id")): item for item in fixture.get("depends_on") or [] if isinstance(item, dict)}
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


def assert_task_card_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("task_card_boundary") or {}
    for field, expected in {
        "task_card_status": "created_static_task_card",
        "implementation_scope_status": "task_card_defined_runtime_not_started",
        "fake_resolver_runtime_status": "not_created",
        "no_secret_leakage_smoke_runtime_status": "not_created",
        "resolver_runtime_status": "not_created",
        "test_fixture_strategy_status": "required_before_implementation",
        "production_secret_backend_status": "not_satisfied",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "runtime_created_in_this_slice",
        "no_secret_leakage_smoke_runtime_created_in_this_slice",
        "repository_mode_enabled",
        "cloud_secret_service_enabled",
        "database_connection_provider_enabled",
        "production_api_enabled",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gates_and_runtime_requirements(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "implementation_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "implementation gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    runtime = fixture.get("runtime_implementation_requirements") or {}
    missing_scope = sorted(EXPECTED_RUNTIME_SCOPE - set(runtime.get("required_scope") or []))
    require(not missing_scope, f"missing runtime scope: {missing_scope}")
    missing_inputs = sorted(EXPECTED_ALLOWED_INPUT_FIELDS - set(runtime.get("allowed_input_fields") or []))
    require(not missing_inputs, f"missing allowed input fields: {missing_inputs}")
    missing_success = sorted(EXPECTED_SUCCESS_FIELDS - set(runtime.get("success_output_fields") or []))
    require(not missing_success, f"missing success output fields: {missing_success}")
    missing_failure = sorted(EXPECTED_FAILURE_FIELDS - set(runtime.get("failure_output_fields") or []))
    require(not missing_failure, f"missing failure output fields: {missing_failure}")
    missing_forbidden = sorted(EXPECTED_MUST_NOT_INCLUDE - set(runtime.get("must_not_include") or []))
    require(not missing_forbidden, f"missing forbidden runtime scope: {missing_forbidden}")


def assert_failure_mapping_and_diagnostics(fixture: dict[str, Any]) -> None:
    failures = {str(item.get("code")): item for item in fixture.get("failure_mapping") or [] if isinstance(item, dict)}
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(failures))
    require(not missing_failures, f"missing failure codes: {missing_failures}")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} must define failure boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} must define sanitized diagnostic")
        lower = diagnostic.lower()
        for forbidden in ("raw secret", "api key", "token", "password", "dsn"):
            require(forbidden not in lower, f"{code} diagnostic must not expose {forbidden}")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    missing_allowed = sorted(EXPECTED_DIAGNOSTIC_FIELDS - set(diagnostics.get("allowed_fields") or []))
    require(not missing_allowed, f"missing diagnostic fields: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_DIAGNOSTICS - set(diagnostics.get("forbidden_fields") or []))
    require(not missing_forbidden, f"missing forbidden diagnostic fields: {missing_forbidden}")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must stay false")
    require(
        diagnostics.get("secret_ref_value_allowed_in_runtime_diagnostics") is False,
        "secret ref value must not appear in runtime diagnostics",
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


def assert_prior_alignment() -> None:
    entry = load_json(ENTRY_REVIEW_PATH)
    boundary = entry.get("entry_boundary") or {}
    require(boundary.get("runtime_task_decision") == "fake_resolver_runtime_implementation_ready_for_next_task", "runtime entry decision drifted")
    require(boundary.get("fake_resolver_runtime_status") == "not_created", "entry review must not create runtime")

    task_card = load_json(IMPLEMENTATION_TASK_CARD_PATH)
    task_boundary = task_card.get("task_card_boundary") or {}
    require(task_boundary.get("implementation_scope_status") == "task_card_defined_runtime_not_started", "implementation task card scope drifted")
    require(task_boundary.get("fake_resolver_runtime_status") == "not_created", "fake resolver runtime must remain not created")


def assert_implementation_readiness_alignment() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in {
        "resolver_implementation_status": "not_started",
        "resolver_runtime_status": "not_created",
        "fake_resolver_status": "not_created",
        "fake_resolver_contract_status": "static_contract_defined",
        "no_secret_leakage_smoke_strategy_status": "static_strategy_defined",
        "no_secret_leakage_smoke_runtime_status": "not_created",
        "test_fixture_strategy_status": "required_before_implementation",
        "fake_resolver_implementation_task_card_status": "created_static_task_card",
        "fake_resolver_implementation_status": "task_card_defined_runtime_not_started",
        "fake_resolver_runtime_implementation_entry_review_status": "ready_for_next_task",
        "fake_resolver_runtime_implementation_task_card_status": "created_static_task_card",
        "fake_resolver_runtime_implementation_status": "task_card_defined_runtime_not_started",
        "default_runtime_state": "disabled_until_explicit_secret_backend_task",
    }.items():
        require(target.get(field) == expected, f"implementation target {field} drifted")

    planned = rows_by_id(readiness, "planned_slices", "id")
    current = planned.get("fake-resolver-runtime-implementation") or {}
    require(current.get("status") == "task_card_defined_runtime_not_started", "planned runtime implementation status drifted")
    required_evidence = {
        "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json",
        "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py",
    }
    missing_evidence = sorted(required_evidence - set(current.get("evidence") or []))
    require(not missing_evidence, f"planned runtime implementation missing evidence: {missing_evidence}")

    required_consumers = required_evidence | {"scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py"}
    missing_consumers = sorted(required_consumers - set(readiness.get("required_consumers") or []))
    require(not missing_consumers, f"implementation readiness missing runtime implementation consumers: {missing_consumers}")


def assert_docs_validation_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        text = read(path)
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in text]
        require(not missing, f"{path} missing literals: {missing}")

    missing_validation = sorted(EXPECTED_VALIDATION - set(fixture.get("validation_strategy") or []))
    require(not missing_validation, f"missing validation strategy entries: {missing_validation}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    prior_call = 'run_python_script("check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py", [])'
    current_call = 'run_python_script("check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py", [])'
    startup_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call in (prior_call, current_call, startup_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(prior_call) < check_repo.index(current_call) < check_repo.index(startup_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"runtime implementation artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "runtime implementation artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "runtime implementation artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_task_card_boundary(fixture)
    assert_gates_and_runtime_requirements(fixture)
    assert_failure_mapping_and_diagnostics(fixture)
    assert_policies(fixture)
    assert_artifact_guard(fixture)
    assert_prior_alignment()
    assert_implementation_readiness_alignment()
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend fake resolver runtime implementation task card checks passed.")


if __name__ == "__main__":
    main()
