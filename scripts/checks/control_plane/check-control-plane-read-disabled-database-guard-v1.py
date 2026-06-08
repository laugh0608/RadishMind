#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-disabled-database-guard-v1.json"
REPOSITORY_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-preconditions-v1.json"
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
EXPECTED_FAILURE_CODES = {
    "database_read_disabled",
    "read_store_unavailable",
    "read_store_contract_mismatch",
}
EXPECTED_RESERVED_INPUTS = {
    "RADISHMIND_CONTROL_PLANE_READ_STORE=database",
    "RADISHMIND_CONTROL_PLANE_READ_STORE=postgres",
    "RADISHMIND_CONTROL_PLANE_READ_STORE=repository",
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
    "RADISHMIND_CONTROL_PLANE_READ_STORE",
    "ReadTenantSummary(",
    "ListApplicationSummaries(",
    "ListAPIKeySummaries(",
    "ReadQuotaSummary(",
    "ListWorkflowDefinitionSummaries(",
    "ListRunRecordSummaries(",
    "ListAuditSummaries(",
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
        fixture.get("kind") == "control_plane_read_disabled_database_guard_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-disabled-database-guard-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "disabled_database_guard_defined",
        "disabled database guard status drifted",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_guard_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("disabled_database_guard_boundary") or {}
    repository = load_json(REPOSITORY_FIXTURE_PATH).get("repository_contract_boundary") or {}
    transition = load_json(AUTH_STORE_FIXTURE_PATH).get("transition_boundary") or {}
    require(boundary.get("status") == "guard_defined_not_implemented", "guard boundary status drifted")
    require(
        boundary.get("current_store_source") == repository.get("current_store_source"),
        "current store source must match repository contract fixture",
    )
    require(
        boundary.get("reserved_disabled_database_source") == transition.get("future_store_source"),
        "reserved database source must match auth/store transition future source",
    )
    for field in (
        "database_read_enabled_now",
        "fallback_to_fake_store_when_database_requested",
        "repository_implementation_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_route_guard_matrix(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    matrix = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_guard_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "route guard matrix must cover every read route")
    for route_id, guard in matrix.items():
        route = routes[route_id]
        require(guard.get("method") == route.get("method"), f"{route_id} method drifted")
        require(guard.get("path") == route.get("path"), f"{route_id} path drifted")
        require(guard.get("read_model") == route.get("read_model"), f"{route_id} read model drifted")
        require(guard.get("required_scope") == route.get("required_scope"), f"{route_id} scope drifted")
        require(guard.get("database_read_enabled_now") is False, f"{route_id} database read must be disabled")
        require(
            guard.get("failure_code_when_requested") == "database_read_disabled",
            f"{route_id} must fail closed with database_read_disabled",
        )
        require(guard.get("fallback_to_fake_store_allowed") is False, f"{route_id} fake fallback must be forbidden")
        require(guard.get("side_effect_allowed") is False, f"{route_id} side effects must be forbidden")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    failures = {
        str(item.get("failure_code") or "")
        for item in fixture.get("failure_mapping") or []
        if isinstance(item, dict)
    }
    missing_failures = sorted(EXPECTED_FAILURE_CODES - failures)
    require(not missing_failures, f"missing guard failure mappings: {missing_failures}")

    repository_failures = {
        str(item.get("failure_code") or "")
        for item in load_json(REPOSITORY_FIXTURE_PATH).get("failure_mapping") or []
        if isinstance(item, dict)
    }
    require(EXPECTED_FAILURE_CODES.issubset(repository_failures), "repository failure mapping must include guard codes")

    auth_store_codes = set(load_json(AUTH_STORE_FIXTURE_PATH).get("required_failure_codes") or [])
    require("database_read_disabled" in auth_store_codes, "auth/store fixture must retain database_read_disabled")

    auth_db_taxonomy = {
        str(item.get("code") or "")
        for item in load_json(AUTH_DB_FIXTURE_PATH).get("failure_taxonomy") or []
        if isinstance(item, dict)
    }
    require("database_read_disabled" in auth_db_taxonomy, "auth/db failure taxonomy must retain database_read_disabled")


def assert_reserved_inputs_and_smoke(fixture: dict[str, Any]) -> None:
    reserved = {
        str(item.get("input") or ""): item
        for item in fixture.get("reserved_database_mode_inputs") or []
        if isinstance(item, dict)
    }
    require(set(reserved) == EXPECTED_RESERVED_INPUTS, "reserved database mode inputs drifted")
    for input_name, item in reserved.items():
        require(
            item.get("status") == "reserved_disabled_until_later_slice",
            f"{input_name} status must remain disabled",
        )
        require(item.get("failure_code") == "database_read_disabled", f"{input_name} failure code drifted")

    smoke = fixture.get("guard_smoke_plan") or {}
    require(smoke.get("status") == "guard_defined_not_implemented", "guard smoke status drifted")
    required_sources = {
        "disabled_database_read_guard_smoke",
        "repository_contract_fake_store_smoke",
        "read_store_repository_contract_smoke",
    }
    require(required_sources.issubset(set(smoke.get("future_smoke_sources") or [])), "missing smoke sources")
    require(
        "repository_write_count=0" in set(smoke.get("side_effect_counters_must_remain") or []),
        "missing zero write counter",
    )


def assert_policy_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-disabled-database-guard-v1.py", [])'
        in check_repo,
        "check-repo.py must run disabled database guard check",
    )
    require(
        check_repo.index("check-control-plane-read-repository-contract-preconditions-v1.py")
        < check_repo.index("check-control-plane-read-disabled-database-guard-v1.py"),
        "disabled database guard check must run after repository contract preconditions",
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
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this guard slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_guard_boundary(fixture)
    assert_route_guard_matrix(fixture)
    assert_failure_mapping(fixture)
    assert_reserved_inputs_and_smoke(fixture)
    assert_policy_docs_and_check_repo(fixture)
    assert_no_implementation_leaked(fixture)
    print("control plane read disabled database guard v1 checks passed.")


if __name__ == "__main__":
    main()
