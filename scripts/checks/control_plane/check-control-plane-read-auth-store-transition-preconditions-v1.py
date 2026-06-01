#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-store-transition-preconditions-v1.json"
)
AUTH_DB_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json"
)
DEV_LIVE_CONSUMER_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-dev-live-consumer-v1.json"
FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json"
)
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
OIDC_PRECONDITIONS_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json"
DATA_BOUNDARY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json"
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
    "production_admin_console_ready",
    "formal_user_workspace_complete",
    "production_api_consumer_ready",
    "database_schema_ready",
    "database_query_ready",
    "repository_migration_ready",
    "repository_implementation_ready",
    "radish_oidc_client_ready",
    "auth_middleware_ready",
    "token_validation_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "rate_limit_ready",
    "billing_ready",
    "cost_ledger_ready",
    "workflow_executor_ready",
    "confirmation_flow_ready",
    "business_writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_AUTH_GATES = {
    "issuer-discovery-contract",
    "token-validation-contract",
    "claim-mapping-contract",
    "tenant-binding-enforcement",
    "scope-grant-projection",
    "audit-context-propagation",
    "fail-closed-missing-identity",
    "no-public-fake-auth-in-production",
    "dev-fake-auth-disabled-by-default",
}
EXPECTED_STORE_GATES = {
    "repository-interface-before-sql",
    "tenant-predicate-from-auth-context",
    "sanitized-summary-projection",
    "cursor-filter-sort-allowlist",
    "read-store-failure-taxonomy",
    "no-database-write",
    "no-secret-material",
}
EXPECTED_FAILURE_CODES = {
    "identity_context_missing",
    "tenant_binding_missing",
    "scope_denied",
    "invalid_filter",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "database_read_disabled",
    "auth_context_contract_mismatch",
}
EXPECTED_FORBIDDEN_IMPLEMENTATION = {
    "Radish OIDC middleware implementation",
    "token validation implementation",
    "database schema",
    "database migration",
    "database query implementation",
    "repository implementation",
    "repository migration",
    "durable store",
    "production API consumer",
    "public fake auth production mode",
    "API key lifecycle",
    "quota enforcement",
    "rate limiting implementation",
    "billing ledger implementation",
    "cost ledger implementation",
    "workflow builder",
    "workflow executor",
    "confirmation flow wiring",
    "business writeback",
    "replay execution",
    "production admin console",
}
REQUIRED_SOURCE_ABSENT_LITERALS = {
    "database/sql",
    "CREATE TABLE",
    "CREATE INDEX",
    "ALTER TABLE",
    "repository migration",
    "oidc.Provider",
    "ValidateToken",
    "IssueAPIKey",
    "RotateAPIKey",
    "billing_ledger",
    "workflow_executor",
    "replay_run",
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


def implemented_routes_by_id() -> dict[str, dict[str, Any]]:
    implementation = load_json(FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH)
    implemented = {
        str(route.get("route_id") or ""): route
        for route in implementation.get("implemented_routes") or []
        if isinstance(route, dict)
    }
    require(set(implemented) == EXPECTED_ROUTE_IDS, "fake-store route ids drifted")
    return implemented


def assert_dependencies(fixture: dict[str, Any]) -> None:
    expected = {
        "control-plane-read-auth-db-preconditions-v1": (
            AUTH_DB_PRECONDITIONS_FIXTURE_PATH,
            "auth_db_preconditions_defined",
        ),
        "control-plane-read-dev-live-consumer-v1": (DEV_LIVE_CONSUMER_FIXTURE_PATH, "dev_live_consumer_implemented"),
        "control-plane-read-fake-store-handler-implementation-v1": (
            FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH,
            "fake_store_handler_implemented",
        ),
        "control-plane-read-route-contract-v1": (ROUTE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-negative-contract-v1": (NEGATIVE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
        "radish-oidc-client-preconditions": (OIDC_PRECONDITIONS_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-data-boundary": (DATA_BOUNDARY_FIXTURE_PATH, "governance_boundary_satisfied"),
    }
    declared = set(fixture.get("depends_on") or [])
    missing = sorted(set(expected) - declared)
    require(not missing, f"missing dependencies: {missing}")

    for expected_id, (path, expected_status) in expected.items():
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
        fixture.get("kind") == "control_plane_read_auth_store_transition_preconditions_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-auth-store-transition-preconditions-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "auth_store_transition_preconditions_defined",
        "auth/store transition must remain precondition-only",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_transition_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("transition_boundary") or {}
    dev_live = load_json(DEV_LIVE_CONSUMER_FIXTURE_PATH).get("consumer_boundary") or {}
    require(boundary.get("status") == "preconditions_defined_not_implemented", "transition boundary status drifted")
    require(boundary.get("current_store_source") == "fixture-backed fake store", "current store source drifted")
    require(
        boundary.get("future_auth_source") == "future Radish OIDC / auth middleware",
        "future auth source drifted",
    )
    require(
        boundary.get("future_store_source") == "future control plane read store repository",
        "future store source drifted",
    )
    require(
        boundary.get("default_consumer_source_must_remain") == dev_live.get("default_source"),
        "default consumer source must match dev live fixture",
    )
    require(
        boundary.get("dev_live_source_must_remain_opt_in") == dev_live.get("live_source_env"),
        "dev live opt-in env drifted",
    )
    require(
        boundary.get("backend_dev_auth_must_remain_opt_in") == dev_live.get("backend_dev_auth_env"),
        "backend dev auth opt-in env drifted",
    )
    for field in (
        "database_query_allowed_in_this_slice",
        "repository_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gate_set(fixture: dict[str, Any], key: str, expected: set[str], status: str) -> None:
    gates = {
        str(item.get("id") or ""): item
        for item in fixture.get(key) or []
        if isinstance(item, dict)
    }
    require(set(gates) == expected, f"{key} ids drifted")
    for gate_id, gate in gates.items():
        require(gate.get("status") == status, f"{gate_id} status must be {status}")
        require(str(gate.get("rule") or "").strip(), f"{gate_id}.rule is required")


def assert_route_matrix(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    implemented = implemented_routes_by_id()
    dev_live = load_json(DEV_LIVE_CONSUMER_FIXTURE_PATH)
    auth_db = load_json(AUTH_DB_PRECONDITIONS_FIXTURE_PATH)
    auth_db_transitions = {
        str(item.get("route_id") or ""): item
        for item in auth_db.get("route_transition_requirements") or []
        if isinstance(item, dict)
    }
    matrix = {
        str(item.get("route_id") or ""): item
        for item in fixture.get("route_transition_matrix") or []
        if isinstance(item, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "route transition matrix must cover every read route")
    require(set(dev_live.get("required_routes") or []) == EXPECTED_ROUTE_IDS, "dev live required routes drifted")
    require(set(auth_db_transitions) == EXPECTED_ROUTE_IDS, "auth/db route transition requirements drifted")

    for route_id, row in matrix.items():
        route = routes[route_id]
        implemented_route = implemented[route_id]
        auth_db_transition = auth_db_transitions[route_id]
        require(row.get("method") == route.get("method"), f"{route_id} method drifted")
        require(row.get("path") == route.get("path"), f"{route_id} path drifted")
        require(row.get("read_model") == route.get("read_model"), f"{route_id} read model drifted")
        require(row.get("required_scope") == route.get("required_scope"), f"{route_id} required scope drifted")
        require(
            implemented_route.get("status") == "fake_store_backed_handler_implemented",
            f"{route_id} status drifted",
        )
        require(
            row.get("future_auth_source") == auth_db_transition.get("future_auth_source"),
            f"{route_id} future auth source must match auth/db preconditions",
        )
        require(
            row.get("future_store_source") == auth_db_transition.get("future_store_source"),
            f"{route_id} future store source must match auth/db preconditions",
        )
        require(row.get("route_shape_must_remain_stable") is True, f"{route_id} route shape must remain stable")
        for field in (
            "database_query_allowed_now",
            "repository_migration_allowed_now",
            "oidc_validation_allowed_now",
            "write_allowed_now",
        ):
            require(row.get(field) is False, f"{route_id} {field} must remain false")


def assert_dual_smoke_and_failures(fixture: dict[str, Any]) -> None:
    smoke = fixture.get("dual_smoke_plan") or {}
    require(
        smoke.get("status") == "preconditions_defined_not_implemented",
        "dual smoke plan must remain precondition-only",
    )
    require(
        {"offline_fixture_consumer_smoke", "dev_live_fake_auth_fake_store_smoke"}.issubset(
            set(smoke.get("current_smoke_sources") or [])
        ),
        "current smoke sources must retain offline and dev live paths",
    )
    require(
        {
            "middleware_backed_auth_context_smoke",
            "repository_contract_fake_store_smoke",
            "read_store_repository_contract_smoke",
            "database_read_disabled_guard_smoke",
        }.issubset(set(smoke.get("future_smoke_sources") or [])),
        "future smoke sources missing required transition gates",
    )
    required_commands = set(smoke.get("must_preserve_current_commands") or [])
    require(
        "./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check" in required_commands,
        "offline consumer smoke must be retained",
    )
    require(
        "./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-dev-live-consumer-v1.py"
        in required_commands,
        "dev live checker must be retained",
    )
    require(
        "repository_write_count=0" in set(smoke.get("side_effect_counters_must_remain") or []),
        "repository write counter must remain zero",
    )
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(fixture.get("required_failure_codes") or []))
    require(not missing_failures, f"missing required failure codes: {missing_failures}")


def assert_policy_docs_and_sources(fixture: dict[str, Any]) -> None:
    missing_scope = sorted(EXPECTED_FORBIDDEN_IMPLEMENTATION - set(fixture.get("forbidden_implementation") or []))
    require(not missing_scope, f"missing forbidden implementation items: {missing_scope}")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-auth-store-transition-preconditions-v1.py", [])'
        in check_repo,
        "check-repo.py must run auth/store transition preconditions check",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")

    source_roots = [
        REPO_ROOT / "services/platform/internal",
        REPO_ROOT / "apps/radishmind-web/src",
    ]
    for root in source_roots:
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in REQUIRED_SOURCE_ABSENT_LITERALS:
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this preconditions slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_transition_boundary(fixture)
    assert_gate_set(
        fixture,
        "auth_middleware_transition_gates",
        EXPECTED_AUTH_GATES,
        "required_before_auth_middleware",
    )
    assert_gate_set(fixture, "read_store_transition_gates", EXPECTED_STORE_GATES, "required_before_read_store")
    assert_route_matrix(fixture)
    assert_dual_smoke_and_failures(fixture)
    assert_policy_docs_and_sources(fixture)
    print("control plane read auth/store transition preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
