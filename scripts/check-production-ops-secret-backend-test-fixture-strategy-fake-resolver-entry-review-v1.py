#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
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
}

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
    "config-secret-ref-readiness": "satisfied",
    "provider-profile-secret-binding": "satisfied",
    "secret-resolver-interface-disabled": "satisfied_disabled_only",
    "operator-runbook-and-negative-gates": "satisfied",
    "rotation-and-audit-policy": "satisfied",
    "workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1": (
        "draft_database_secret_resolver_implementation_entry_review_defined"
    ),
    "production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1": (
        "fake_resolver_contract_no_secret_leakage_smoke_strategy_defined"
    ),
}

EXPECTED_GATE_STATUS = {
    "reference_only_secret_fixture_consumed": "satisfied_reference_only",
    "disabled_resolver_interface_consumed": "satisfied_disabled_only",
    "operator_runbook_consumed": "satisfied_runbook_only",
    "rotation_audit_policy_consumed": "satisfied_policy_only",
    "workflow_secret_resolver_entry_review_consumed": "satisfied_blocked_entry",
    "test_fixture_strategy_review_defined": "blocked_entry_review_defined",
    "fake_resolver_contract_gate": "static_contract_strategy_defined",
    "fake_resolver_implementation_gate": "blocked",
    "no_secret_leakage_smoke_gate": "static_smoke_strategy_defined",
    "sanitized_diagnostics_runtime_gate": "blocked",
    "cloud_secret_backend_call_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_runtime_gate": "blocked",
    "artifact_guard": "required_now",
}

EXPECTED_CANDIDATES = {
    "placeholder_secret_ref_fixture_strategy",
    "fake_resolver_interface_contract",
    "fake_resolver_implementation",
    "sanitized_diagnostics_fixture",
    "connection_factory_handoff_fixture",
    "repository_mode_fixture",
}

EXPECTED_FAILURE_CODES = {
    "test_fixture_strategy_missing",
    "fake_resolver_contract_missing",
    "fake_resolver_implementation_forbidden",
    "secret_fixture_value_detected",
    "fake_resolver_fallback_forbidden",
    "test_fixture_cloud_call_forbidden",
    "test_fixture_repository_mode_forbidden",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "test_fixture_strategy_status",
    "fake_resolver_entry_status",
    "resolver_state",
    "secret_ref_status",
    "secret_ref_version_status",
    "operator_gate_status",
    "rotation_gate_status",
    "failure_code",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_NO_FALLBACK = {
    "no fallback to RADISHMIND_PLATFORM_API_KEY",
    "no fallback to mock provider",
    "no fallback to local-smoke profile",
    "no fallback to fixture credential",
    "no fallback to committed secret value",
    "no cross-environment secret_ref fallback",
    "no fallback to fake resolver",
    "no fallback to fake query executor",
    "no fallback from test fixture strategy to production resolver",
    "no fallback from fake resolver to production secret backend",
    "test fixture strategy does not mean credential resolved",
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
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a json object")
    return document


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {
        str(row.get(id_field) or ""): row
        for row in fixture.get(key) or []
        if isinstance(row, dict)
    }
    require(rows, f"{key} must not be empty")
    return rows


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_test_fixture_strategy_fake_resolver_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "test_fixture_strategy_fake_resolver_entry_review_defined",
        "unexpected status",
    )
    task_card = str(slice_info.get("task_card") or "")
    platform_topic = str(slice_info.get("platform_topic") or "")
    require(
        task_card
        == "docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md",
        "unexpected task card",
    )
    require(
        platform_topic
        == "docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md",
        "unexpected platform topic",
    )
    require((REPO_ROOT / task_card).exists(), "task card must exist")
    require((REPO_ROOT / platform_topic).exists(), "platform topic must exist")
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

    implementation = load_json(IMPLEMENTATION_READINESS_PATH)
    target = implementation.get("implementation_target") or {}
    require(target.get("resolver_implementation_status") == "not_started", "resolver must remain not_started")
    require(
        target.get("default_runtime_state") == "disabled_until_explicit_secret_backend_task",
        "secret backend runtime must remain disabled",
    )
    require(
        target.get("fast_baseline_policy") == "offline_no_real_credentials_no_cloud_calls",
        "fast baseline policy drifted",
    )

    reference = load_json(SECRET_REFERENCE_PATH)
    require(reference.get("scope") == "secret_reference_only", "secret reference scope must stay reference-only")
    policy = reference.get("policy") or {}
    for field in ("stores_secret_values", "resolver_enabled", "cloud_calls_allowed", "production_secret_backend_ready"):
        require(policy.get(field) is False, f"secret reference policy {field} must be false")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_boundary") or {}
    require(
        boundary.get("entry_decision") == "fake_resolver_implementation_entry_not_opened",
        "entry decision drifted",
    )
    require(
        boundary.get("test_fixture_strategy_status") == "required_before_implementation",
        "test fixture strategy must remain required",
    )
    require(
        boundary.get("test_fixture_strategy_review_status") == "blocked_entry_review_defined",
        "review status drifted",
    )
    for field, expected in {
        "resolver_implementation_status": "not_started",
        "resolver_runtime_status": "not_created",
        "fake_resolver_status": "not_created",
        "fake_resolver_contract_status": "static_contract_defined",
        "fake_resolver_runtime_status": "not_created",
        "no_secret_leakage_smoke_strategy_status": "static_strategy_defined",
        "default_runtime_state": "disabled",
        "production_secret_backend_status": "not_satisfied",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "implementation_task_card_created_in_this_slice",
        "resolver_runtime_added",
        "fake_resolver_added",
        "cloud_secret_sdk_added",
        "secret_value_fixture_added",
        "no_secret_leakage_smoke_added",
        "credential_handle_created",
        "database_connection_provider_added",
        "db_driver_added",
        "connection_factory_added",
        "sql_added",
        "schema_marker_added",
        "migration_runner_added",
        "repository_mode_enabled",
        "production_api_added",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gates_and_candidates(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "entry_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "entry gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    candidates = rows_by_id(fixture, "candidate_review", "candidate_id")
    require(set(candidates) == EXPECTED_CANDIDATES, "candidate ids drifted")
    for candidate_id, candidate in candidates.items():
        require(candidate.get("entry_decision") == "blocked", f"{candidate_id} must remain blocked")
        require(candidate.get("current_blockers"), f"{candidate_id} must list blockers")


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
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    for field in {
        "raw_secret",
        "password",
        "token",
        "api_key",
        "provider_raw_url",
        "dsn",
        "cloud_credential",
        "database_hostname",
        "database_error_detail",
        "opaque_credential_handle",
        "resolver_backend_url",
        "full_user_claim",
    }:
        require(field in forbidden, f"missing forbidden diagnostic field: {field}")
    require(
        diagnostics.get("secret_ref_value_allowed_in_runtime_diagnostics") is False,
        "runtime diagnostics must not expose secret_ref value",
    )
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must remain disabled")


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
    preconditions = {
        str(item.get("id")): item
        for item in readiness.get("required_preconditions") or []
        if isinstance(item, dict)
    }
    test_fixture = preconditions.get("test-fixture-strategy") or {}
    require(
        test_fixture.get("status") == "required_before_implementation",
        "test-fixture-strategy must remain required_before_implementation",
    )
    evidence = set(test_fixture.get("evidence") or [])
    for path in {
        "docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json",
        "scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py",
        "docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json",
        "scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
    }:
        require(path in evidence, f"test-fixture-strategy missing evidence: {path}")
        require((REPO_ROOT / path).exists(), f"test-fixture-strategy evidence missing: {path}")

    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or [] if isinstance(item, dict)}
    review = planned.get("test-fixture-strategy") or {}
    require(review.get("status") == "blocked_entry_review_defined", "planned test fixture review status drifted")
    strategy = planned.get("fake-resolver-contract-no-secret-leakage-smoke-strategy") or {}
    require(strategy.get("status") == "strategy_defined_static_only", "planned strategy status drifted")

    blocked = {str(item.get("id")): item for item in readiness.get("blocked_conditions") or [] if isinstance(item, dict)}
    for blocked_id, expected_status in {
        "production_secret_backend": "not_satisfied",
        "cloud_secret_service_integration": "not_satisfied",
        "test_fixture_strategy": "blocked_entry_review_defined",
        "fake_resolver_implementation": "not_satisfied",
        "real_secret_values": "forbidden_in_committed_repo",
        "production_ready": "not_satisfied",
    }.items():
        require(blocked.get(blocked_id, {}).get("status") == expected_status, f"{blocked_id} status drifted")


def assert_docs_validation_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        text = read(path)
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in text]
        require(not missing, f"{path} missing literals: {missing}")

    missing_validation = sorted(EXPECTED_VALIDATION - set(fixture.get("validation_strategy") or []))
    require(not missing_validation, f"missing validation strategy entries: {missing_validation}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    rotation_call = 'run_python_script("check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py", [])'
    current_call = (
        'run_python_script("check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py", [])'
    )
    strategy_call = (
        'run_python_script("check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py", [])'
    )
    startup_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call, description in {
        rotation_call: "rotation audit policy readiness checker",
        current_call: "test fixture strategy fake resolver entry review checker",
        strategy_call: "fake resolver contract no leakage strategy checker",
        startup_call: "startup supervisor checker",
    }.items():
        require(call in check_repo, f"check-repo.py must run {description}")
    require(
        check_repo.index(rotation_call)
        < check_repo.index(current_call)
        < check_repo.index(strategy_call)
        < check_repo.index(startup_call),
        "test fixture strategy checker order drifted",
    )


def assert_no_secret_literals() -> None:
    text = "\n".join(
        [
            FIXTURE_PATH.read_text(encoding="utf-8"),
            read("docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md"),
            read("docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md"),
        ]
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"test fixture strategy review contains forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", text) is None,
        "test fixture strategy review contains forbidden sk-like token",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_gates_and_candidates(fixture)
    assert_failure_mapping_and_diagnostics(fixture)
    assert_policies(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment()
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend test fixture strategy fake resolver entry review checks passed.")


if __name__ == "__main__":
    main()
