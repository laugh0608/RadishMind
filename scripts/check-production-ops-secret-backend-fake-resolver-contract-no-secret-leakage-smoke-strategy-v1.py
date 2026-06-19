#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json"
)
ENTRY_REVIEW_PATH = (
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
}

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1": (
        "test_fixture_strategy_fake_resolver_entry_review_defined"
    ),
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
    "secret-resolver-interface-disabled": "satisfied_disabled_only",
    "workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1": (
        "draft_database_secret_resolver_implementation_entry_review_defined"
    ),
}

EXPECTED_CONTRACT_INPUTS = {
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

EXPECTED_CONTRACT_OUTPUTS = {
    "fake_resolver_status",
    "credential_handle_metadata",
    "secret_ref_key_status",
    "secret_ref_version_status",
    "environment_binding_status",
    "failure_code",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

FORBIDDEN_SECRET_FIELDS = {
    "raw_secret",
    "secret_value",
    "password",
    "token",
    "api_key",
    "dsn",
    "provider_raw_url",
    "database_hostname",
    "cloud_credential",
    "credential_payload",
}

EXPECTED_FAILURE_CODES = {
    "fake_resolver_contract_runtime_missing",
    "fake_resolver_input_shape_invalid",
    "fake_resolver_secret_value_detected",
    "fake_resolver_output_secret_leakage",
    "fake_resolver_runtime_forbidden",
    "fake_resolver_cloud_call_forbidden",
    "fake_resolver_repository_mode_forbidden",
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
    "contract strategy does not mean credential resolved",
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


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a JSON object")
    return document


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_fake_resolver_contract_no_secret_leakage_smoke_strategy_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id")
        == "production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "fake_resolver_contract_no_secret_leakage_smoke_strategy_defined",
        "unexpected status",
    )
    for key, expected_path in {
        "task_card": (
            "docs/task-cards/"
            "production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md"
        ),
        "platform_topic": (
            "docs/platform/"
            "production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md"
        ),
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


def assert_strategy_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("strategy_boundary") or {}
    for field, expected in {
        "fake_resolver_contract_status": "static_contract_defined",
        "no_secret_leakage_smoke_strategy_status": "static_strategy_defined",
        "fake_resolver_implementation_entry": "not_opened",
        "fake_resolver_runtime_status": "not_created",
        "no_secret_leakage_smoke_runtime_status": "not_created",
        "resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "repository_mode_enabled",
        "cloud_secret_service_enabled",
        "database_connection_provider_enabled",
        "implementation_task_card_created_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("fake_resolver_contract") or {}
    missing_inputs = sorted(EXPECTED_CONTRACT_INPUTS - set(contract.get("input_allowed_fields") or []))
    require(not missing_inputs, f"missing input fields: {missing_inputs}")
    missing_outputs = sorted(EXPECTED_CONTRACT_OUTPUTS - set(contract.get("output_allowed_fields") or []))
    require(not missing_outputs, f"missing output fields: {missing_outputs}")
    missing_forbidden_inputs = sorted(FORBIDDEN_SECRET_FIELDS - set(contract.get("input_forbidden_fields") or []))
    require(not missing_forbidden_inputs, f"missing forbidden input fields: {missing_forbidden_inputs}")
    missing_forbidden_outputs = sorted(FORBIDDEN_SECRET_FIELDS - set(contract.get("output_forbidden_fields") or []))
    require(not missing_forbidden_outputs, f"missing forbidden output fields: {missing_forbidden_outputs}")
    semantics = contract.get("success_semantics") or {}
    for field in (
        "credential_resolved_allowed",
        "opaque_handle_payload_allowed",
        "runtime_success_allowed_in_this_slice",
    ):
        require(semantics.get(field) is False, f"{field} must remain false")


def assert_no_leakage_strategy(fixture: dict[str, Any]) -> None:
    strategy = fixture.get("no_secret_leakage_smoke_strategy") or {}
    require(strategy.get("status") == "static_strategy_defined", "no leakage strategy status drifted")
    require(strategy.get("runtime_smoke_created") is False, "runtime smoke must not be created")
    require(strategy.get("fixture_created") is False, "no leakage smoke fixture must not be created")
    for required in {
        "committed fixture",
        "platform topic",
        "task card",
        "future smoke record",
        "checker source",
    }:
        require(required in set(strategy.get("must_scan") or []), f"no leakage strategy must scan {required}")
    require(len(strategy.get("forbidden_secret_patterns") or []) >= 6, "secret pattern list is too weak")
    require(len(strategy.get("must_fail_closed_on") or []) >= 8, "fail closed list is too weak")


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
    missing_allowed = sorted(EXPECTED_CONTRACT_OUTPUTS - set(diagnostics.get("allowed_fields") or []))
    require(not missing_allowed, f"missing diagnostic fields: {missing_allowed}")
    forbidden_fields = set(diagnostics.get("forbidden_fields") or [])
    for field in FORBIDDEN_SECRET_FIELDS | {"full_secret_ref_value", "full_user_claim"}:
        require(field in forbidden_fields, f"missing forbidden diagnostic field: {field}")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must remain false")


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


def assert_readiness_alignment() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    require(target.get("resolver_implementation_status") == "not_started", "resolver must remain not_started")
    require(target.get("fake_resolver_status") == "not_created", "fake resolver must remain not_created")
    require(
        target.get("fake_resolver_contract_status") == "static_contract_defined",
        "implementation readiness must record fake resolver static contract",
    )
    require(
        target.get("no_secret_leakage_smoke_strategy_status") == "static_strategy_defined",
        "implementation readiness must record no leakage strategy",
    )
    require(
        target.get("no_secret_leakage_smoke_runtime_status") == "not_created",
        "no leakage smoke runtime must not be created",
    )

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
    required_evidence = {
        "docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json",
        "scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
    }
    missing_evidence = sorted(required_evidence - set(test_fixture.get("evidence") or []))
    require(not missing_evidence, f"test-fixture-strategy missing strategy evidence: {missing_evidence}")

    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or [] if isinstance(item, dict)}
    strategy = planned.get("fake-resolver-contract-no-secret-leakage-smoke-strategy") or {}
    require(strategy.get("status") == "strategy_defined_static_only", "planned strategy status drifted")


def assert_entry_review_alignment() -> None:
    entry = load_json(ENTRY_REVIEW_PATH)
    gates = {
        str(item.get("gate_id")): item
        for item in entry.get("entry_gate_matrix") or []
        if isinstance(item, dict)
    }
    require(
        gates.get("fake_resolver_contract_gate", {}).get("status") == "static_contract_strategy_defined",
        "entry review must record fake resolver contract strategy",
    )
    require(
        gates.get("no_secret_leakage_smoke_gate", {}).get("status") == "static_smoke_strategy_defined",
        "entry review must record no leakage smoke strategy",
    )
    require(
        gates.get("fake_resolver_implementation_gate", {}).get("status") == "blocked",
        "fake resolver implementation gate must remain blocked",
    )
    boundary = entry.get("implementation_boundary") or {}
    require(boundary.get("entry_decision") == "fake_resolver_implementation_entry_not_opened", "entry opened")
    require(boundary.get("fake_resolver_status") == "not_created", "fake resolver must remain not created")


def assert_docs_validation_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        text = read(path)
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in text]
        require(not missing, f"{path} missing literals: {missing}")

    missing_validation = sorted(EXPECTED_VALIDATION - set(fixture.get("validation_strategy") or []))
    require(not missing_validation, f"missing validation strategy entries: {missing_validation}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    entry_call = (
        'run_python_script("check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py", [])'
    )
    current_call = (
        'run_python_script("check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py", [])'
    )
    startup_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call in (entry_call, current_call, startup_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(entry_call) < check_repo.index(current_call) < check_repo.index(startup_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"strategy artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "strategy artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "strategy artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_strategy_boundary(fixture)
    assert_contract(fixture)
    assert_no_leakage_strategy(fixture)
    assert_failure_mapping_and_diagnostics(fixture)
    assert_policies(fixture)
    assert_artifact_guard(fixture)
    assert_readiness_alignment()
    assert_entry_review_alignment()
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend fake resolver contract no secret leakage smoke strategy checks passed.")


if __name__ == "__main__":
    main()
