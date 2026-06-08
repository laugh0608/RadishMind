#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-v1.json"
REPOSITORY_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-preconditions-v1.json"
)
DISABLED_DATABASE_GUARD_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-disabled-database-guard-v1.json"
)
AUTH_STORE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-store-transition-preconditions-v1.json"
)
AUTH_DB_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json"
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
RESPONSE_FIXTURES_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_ROUTE_IDS = {
    "tenant-summary-route",
    "application-summary-list-route",
    "api-key-summary-list-route",
    "quota-summary-route",
    "workflow-definition-summary-list-route",
    "run-record-summary-list-route",
    "audit-summary-list-route",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "repository_implementation_ready",
    "repository_migration_ready",
    "radish_oidc_client_ready",
    "auth_middleware_ready",
    "token_validation_ready",
    "production_api_consumer_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "billing_ready",
    "cost_ledger_ready",
    "workflow_executor_ready",
    "confirmation_flow_ready",
    "business_writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_DEPENDENCIES = {
    "control-plane-read-repository-contract-preconditions-v1": (
        REPOSITORY_FIXTURE_PATH,
        "repository_contract_preconditions_defined",
    ),
    "control-plane-read-disabled-database-guard-v1": (
        DISABLED_DATABASE_GUARD_FIXTURE_PATH,
        "disabled_database_guard_defined",
    ),
    "control-plane-read-auth-store-transition-preconditions-v1": (
        AUTH_STORE_FIXTURE_PATH,
        "auth_store_transition_preconditions_defined",
    ),
    "control-plane-read-auth-db-preconditions-v1": (
        AUTH_DB_FIXTURE_PATH,
        "auth_db_preconditions_defined",
    ),
    "control-plane-read-route-contract-v1": (
        ROUTE_CONTRACT_FIXTURE_PATH,
        "governance_boundary_satisfied",
    ),
    "control-plane-read-negative-contract-v1": (
        NEGATIVE_CONTRACT_FIXTURE_PATH,
        "governance_boundary_satisfied",
    ),
    "control-plane-read-response-fixtures-v1": (
        RESPONSE_FIXTURES_PATH,
        "governance_boundary_satisfied",
    ),
}
EXPECTED_CONTEXT_FIELDS = {
    "request_id",
    "tenant_ref",
    "subject_ref",
    "scope_grants",
    "audit_ref",
    "issuer_ref",
    "session_ref",
}
EXPECTED_REQUEST_FIELDS = {"limit", "cursor", "filters", "sort", "projection"}
EXPECTED_OUTPUT_FIELDS = {"tenant_ref", "items", "next_cursor", "failure_code", "audit_ref"}
EXPECTED_INPUT_FIELDS = {
    "route_id",
    "method",
    "path",
    "read_model",
    "operation",
    "read_mode",
    "repository_context",
    "request",
}
EXPECTED_FAILURE_CODES = {
    "tenant_binding_missing",
    "scope_denied",
    "invalid_filter",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "database_read_disabled",
    "auth_context_contract_mismatch",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "database_write",
    "api_key_issue",
    "quota_mutation",
    "workflow_execution",
    "confirmation_decision",
    "business_writeback",
    "replay_execution",
}
EXPECTED_SOURCE_ABSENT_LITERALS = {
    "database/sql",
    "CREATE TABLE",
    "ALTER TABLE",
    "CREATE INDEX",
    "SELECT *",
    "INSERT INTO",
    "UPDATE ",
    "DELETE FROM",
    "oidc.Provider",
    "ValidateToken",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def route_contracts_by_id() -> dict[str, dict[str, Any]]:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    routes = {
        str(route.get("id") or ""): route
        for route in route_contract.get("route_contracts") or []
        if isinstance(route, dict)
    }
    require(set(routes) == EXPECTED_ROUTE_IDS, "route contract ids drifted")
    return routes


def repository_operations_by_route_id() -> dict[str, dict[str, Any]]:
    repository = load_json(REPOSITORY_FIXTURE_PATH)
    operations = {
        str(operation.get("route_id") or ""): operation
        for operation in repository.get("repository_operations") or []
        if isinstance(operation, dict)
    }
    require(set(operations) == EXPECTED_ROUTE_IDS, "repository operation ids drifted")
    return operations


def guard_matrix_by_route_id() -> dict[str, dict[str, Any]]:
    guard = load_json(DISABLED_DATABASE_GUARD_FIXTURE_PATH)
    matrix = {
        str(row.get("route_id") or ""): row
        for row in guard.get("route_guard_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "disabled database guard route ids drifted")
    return matrix


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    missing = sorted(set(EXPECTED_DEPENDENCIES) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for expected_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == expected_id, f"dependency {expected_id} id drifted")
        require(
            slice_info.get("status") == expected_status,
            f"dependency {expected_id} status must be {expected_status}",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "control_plane_read_repository_contract_smoke_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-repository-contract-smoke-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "repository_contract_smoke_defined",
        "repository contract smoke status drifted",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_smoke_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("smoke_boundary") or {}
    repository = load_json(REPOSITORY_FIXTURE_PATH).get("repository_contract_boundary") or {}
    guard = load_json(DISABLED_DATABASE_GUARD_FIXTURE_PATH).get("disabled_database_guard_boundary") or {}
    require(boundary.get("status") == "contract_smoke_defined_not_implemented", "smoke boundary status drifted")
    require(
        boundary.get("current_store_source") == repository.get("current_store_source"),
        "current store source must match repository preconditions",
    )
    require(
        boundary.get("future_store_source") == repository.get("future_source"),
        "future store source must match repository preconditions",
    )
    require(
        boundary.get("future_store_source") == guard.get("reserved_disabled_database_source"),
        "future store source must match disabled database guard",
    )
    for field in (
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "repository_implementation_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_smoke_io_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("smoke_io_contract") or {}
    require(contract.get("status") == "input_output_contract_defined", "smoke IO status drifted")
    require(
        contract.get("harness_name") == "ControlPlaneReadRepositoryContractSmoke",
        "smoke harness name drifted",
    )
    require(set(contract.get("input_fields") or []) == EXPECTED_INPUT_FIELDS, "smoke input fields drifted")
    require(
        set(contract.get("repository_context_fields") or []) == EXPECTED_CONTEXT_FIELDS,
        "repository context fields drifted",
    )
    require(set(contract.get("request_fields") or []) == EXPECTED_REQUEST_FIELDS, "request fields drifted")
    require(set(contract.get("output_fields") or []) == EXPECTED_OUTPUT_FIELDS, "output fields drifted")
    require(
        contract.get("success_output_policy") == "sanitized_summary_envelope_only",
        "success output policy drifted",
    )
    require(
        contract.get("failure_output_policy") == "failure_code_without_database_internal_detail",
        "failure output policy drifted",
    )


def assert_route_smoke_matrix(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    operations = repository_operations_by_route_id()
    guard_matrix = guard_matrix_by_route_id()
    matrix = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_smoke_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "repository contract smoke must cover every read route")
    for route_id, smoke in matrix.items():
        route = routes[route_id]
        operation = operations[route_id]
        guard = guard_matrix[route_id]
        for field in ("method", "path", "read_model", "required_scope"):
            require(smoke.get(field) == route.get(field), f"{route_id} {field} drifted from route contract")
        require(smoke.get("operation") == operation.get("operation"), f"{route_id} operation drifted")
        require(smoke.get("read_mode") == "future_repository_contract", f"{route_id} read mode drifted")
        request = smoke.get("request") or {}
        require(set(request) == EXPECTED_REQUEST_FIELDS, f"{route_id} request fields drifted")
        require(request.get("projection") == operation.get("projection"), f"{route_id} projection drifted")
        require(set((request.get("filters") or {}).keys()).issubset(set(operation.get("allowed_filters") or [])), f"{route_id} filters must stay allowlisted")
        if request.get("sort") is not None:
            require(request.get("sort") in set(operation.get("allowed_sort") or []), f"{route_id} sort must stay allowlisted")
        expected_success = smoke.get("expected_success") or {}
        require(
            expected_success.get("tenant_ref_source") == "repository_context.tenant_ref",
            f"{route_id} tenant source must come from repository context",
        )
        require(expected_success.get("failure_code") is None, f"{route_id} success failure_code must be null")
        require(expected_success.get("audit_ref_required") is True, f"{route_id} audit_ref must be required")
        require(
            smoke.get("database_mode_failure_code") == guard.get("failure_code_when_requested"),
            f"{route_id} database mode failure code must match disabled guard",
        )
        require(smoke.get("fake_fallback_allowed") is False, f"{route_id} fake fallback must be forbidden")
        require(smoke.get("side_effect_allowed") is False, f"{route_id} side effects must be forbidden")


def assert_failure_mapping_and_policies(fixture: dict[str, Any]) -> None:
    failures = {
        str(item.get("failure_code") or "")
        for item in fixture.get("failure_mapping") or []
        if isinstance(item, dict)
    }
    missing_failures = sorted(EXPECTED_FAILURE_CODES - failures)
    require(not missing_failures, f"missing smoke failure mappings: {missing_failures}")

    repository_failures = {
        str(item.get("failure_code") or "")
        for item in load_json(REPOSITORY_FIXTURE_PATH).get("failure_mapping") or []
        if isinstance(item, dict)
    }
    guard_failures = {
        str(item.get("failure_code") or "")
        for item in load_json(DISABLED_DATABASE_GUARD_FIXTURE_PATH).get("failure_mapping") or []
        if isinstance(item, dict)
    }
    require(failures.issubset(repository_failures), "smoke failure mapping must be backed by repository fixture")
    require(
        {"database_read_disabled", "read_store_unavailable", "read_store_contract_mismatch"}.issubset(guard_failures),
        "disabled database guard must retain database/read-store failure codes",
    )

    fallback = fixture.get("no_fake_fallback_policy") or {}
    require(fallback.get("status") == "required_for_future_smoke", "fallback policy status drifted")
    require(
        fallback.get("database_mode_requested_failure_code") == "database_read_disabled",
        "database mode failure code drifted",
    )
    require(fallback.get("repository_unavailable_failure_code") == "read_store_unavailable", "unavailable code drifted")
    require(
        fallback.get("contract_mismatch_failure_code") == "read_store_contract_mismatch",
        "contract mismatch code drifted",
    )
    for field in (
        "fallback_to_fixture_fake_store_allowed",
        "fallback_to_test_auth_allowed",
        "fallback_to_success_on_unknown_filter_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
    require(side_effects.get("status") == "required_for_future_smoke", "side effect policy status drifted")
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(side_effects.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effects.get("forbidden_side_effects") or [])),
        "missing forbidden side effects",
    )


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-repository-contract-smoke-v1.py", [])'
        in check_repo,
        "check-repo.py must run repository contract smoke check",
    )
    require(
        check_repo.index("check-control-plane-read-repository-contract-preconditions-v1.py")
        < check_repo.index("check-control-plane-read-disabled-database-guard-v1.py")
        < check_repo.index("check-control-plane-read-repository-contract-smoke-v1.py"),
        "repository contract smoke check must run after preconditions and disabled guard",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_implementation_leaked(fixture: dict[str, Any]) -> None:
    configured_literals = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_SOURCE_ABSENT_LITERALS.issubset(configured_literals), "source absent literals drifted")
    source_roots = [
        REPO_ROOT / "services/platform/internal",
        REPO_ROOT / "apps/radishmind-web/src",
    ]
    for root in source_roots:
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured_literals:
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this smoke slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_smoke_boundary(fixture)
    assert_smoke_io_contract(fixture)
    assert_route_smoke_matrix(fixture)
    assert_failure_mapping_and_policies(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_implementation_leaked(fixture)
    print("control plane read repository contract smoke v1 checks passed.")


if __name__ == "__main__":
    main()
