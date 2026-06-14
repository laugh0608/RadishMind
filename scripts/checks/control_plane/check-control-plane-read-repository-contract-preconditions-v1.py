#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-preconditions-v1.json"
)
AUTH_STORE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-store-transition-preconditions-v1.json"
)
AUTH_DB_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json"
FAKE_STORE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json"
)
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
RESPONSE_FIXTURES_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
DURABLE_READ_FOUNDATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-durable-read-foundation-v1.json"
)

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
EXPECTED_RESULT_FIELDS = {"tenant_ref", "items", "next_cursor", "failure_code", "audit_ref"}
EXPECTED_FORBIDDEN_OUTPUT_FIELDS = {
    "raw_secret_value",
    "api_key_value",
    "api_key_hash",
    "authorization_header",
    "bearer_token",
    "cookie_value",
    "raw_request_body",
    "raw_tool_payload",
    "business_writeback_payload",
    "full_prompt_dump",
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
EXPECTED_FORBIDDEN_SCOPE = {
    "database schema",
    "database migration",
    "database query implementation",
    "repository implementation",
    "repository migration",
    "durable read store adapter",
    "Radish OIDC middleware implementation",
    "token validation implementation",
    "production API consumer",
    "API key lifecycle",
    "quota enforcement",
    "billing ledger implementation",
    "cost ledger implementation",
    "workflow executor",
    "confirmation flow wiring",
    "business writeback",
    "replay execution",
}
REPOSITORY_INTERFACE_IMPLEMENTATION_LITERALS = {
    "ReadTenantSummary(",
    "ListApplicationSummaries(",
    "ListAPIKeySummaries(",
    "ReadQuotaSummary(",
    "ListWorkflowDefinitionSummaries(",
    "ListRunRecordSummaries(",
    "ListAuditSummaries(",
}
REQUIRED_SOURCE_ABSENT_LITERALS = {
    "database/sql",
    "CREATE TABLE",
    "ALTER TABLE",
    "CREATE INDEX",
    "oidc.Provider",
    "ValidateToken",
} | REPOSITORY_INTERFACE_IMPLEMENTATION_LITERALS


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def durable_read_foundation_implemented() -> bool:
    if not DURABLE_READ_FOUNDATION_FIXTURE_PATH.exists():
        return False
    fixture = load_json(DURABLE_READ_FOUNDATION_FIXTURE_PATH)
    slice_info = fixture.get("slice") or {}
    return (
        slice_info.get("id") == "control-plane-durable-read-foundation-v1"
        and slice_info.get("status") == "durable_read_foundation_implemented"
    )


def source_absent_literals_for_current_stage() -> set[str]:
    literals = set(REQUIRED_SOURCE_ABSENT_LITERALS)
    if durable_read_foundation_implemented():
        literals -= REPOSITORY_INTERFACE_IMPLEMENTATION_LITERALS
    return literals


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
    implementation = load_json(FAKE_STORE_FIXTURE_PATH)
    implemented = {
        str(route.get("route_id") or ""): route
        for route in implementation.get("implemented_routes") or []
        if isinstance(route, dict)
    }
    require(set(implemented) == EXPECTED_ROUTE_IDS, "fake-store route ids drifted")
    return implemented


def assert_dependencies(fixture: dict[str, Any]) -> None:
    expected = {
        "control-plane-read-auth-store-transition-preconditions-v1": (
            AUTH_STORE_FIXTURE_PATH,
            "auth_store_transition_preconditions_defined",
        ),
        "control-plane-read-auth-db-preconditions-v1": (AUTH_DB_FIXTURE_PATH, "auth_db_preconditions_defined"),
        "control-plane-read-fake-store-handler-implementation-v1": (
            FAKE_STORE_FIXTURE_PATH,
            "fake_store_handler_implemented",
        ),
        "control-plane-read-route-contract-v1": (ROUTE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-negative-contract-v1": (NEGATIVE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-response-fixtures-v1": (RESPONSE_FIXTURES_PATH, "governance_boundary_satisfied"),
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
        fixture.get("kind") == "control_plane_read_repository_contract_preconditions_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-repository-contract-preconditions-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "repository_contract_preconditions_defined",
        "repository contract must remain precondition-only",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("repository_contract_boundary") or {}
    transition = load_json(AUTH_STORE_FIXTURE_PATH).get("transition_boundary") or {}
    require(boundary.get("status") == "preconditions_defined_not_implemented", "boundary status drifted")
    require(
        boundary.get("future_source") == transition.get("future_store_source"),
        "future repository source must match auth/store transition fixture",
    )
    require(
        boundary.get("current_store_source") == transition.get("current_store_source"),
        "current store source must match auth/store transition fixture",
    )
    for field in (
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "repository_implementation_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_interface_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("interface_contract") or {}
    require(contract.get("status") == "contract_defined_not_implemented", "interface status drifted")
    require(contract.get("interface_name") == "ControlPlaneReadRepository", "interface name drifted")
    require(contract.get("context_argument") == "ReadRepositoryContext", "context argument drifted")
    require(set(contract.get("required_context_fields") or []) == EXPECTED_CONTEXT_FIELDS, "context fields drifted")
    require(set(contract.get("required_request_fields") or []) == EXPECTED_REQUEST_FIELDS, "request fields drifted")
    require(set(contract.get("required_result_fields") or []) == EXPECTED_RESULT_FIELDS, "result fields drifted")
    forbidden_dependencies = set(contract.get("must_not_depend_on") or [])
    require("http_request" in forbidden_dependencies, "repository contract must not depend on HTTP request")
    require("workflow_executor" in forbidden_dependencies, "repository contract must not depend on workflow executor")


def assert_repository_operations(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    implemented = implemented_routes_by_id()
    auth_db = load_json(AUTH_DB_FIXTURE_PATH)
    auth_db_bindings = {
        str(binding.get("route_id") or ""): binding
        for binding in (auth_db.get("database_read_store_preconditions") or {}).get("route_bindings") or []
        if isinstance(binding, dict)
    }
    operations = {
        str(operation.get("route_id") or ""): operation
        for operation in fixture.get("repository_operations") or []
        if isinstance(operation, dict)
    }
    require(set(operations) == EXPECTED_ROUTE_IDS, "repository operations must cover every read route")
    require(set(auth_db_bindings) == EXPECTED_ROUTE_IDS, "auth/db repository bindings drifted")

    for route_id, operation in operations.items():
        route = routes[route_id]
        implemented_route = implemented[route_id]
        auth_db_binding = auth_db_bindings[route_id]
        require(operation.get("method") == route.get("method"), f"{route_id} method drifted")
        require(operation.get("path") == route.get("path"), f"{route_id} path drifted")
        require(operation.get("read_model") == route.get("read_model"), f"{route_id} read model drifted")
        require(operation.get("required_scope") == route.get("required_scope"), f"{route_id} scope drifted")
        require(operation.get("operation") == auth_db_binding.get("repository_operation"), f"{route_id} op drifted")
        require(operation.get("mode") == auth_db_binding.get("mode"), f"{route_id} mode drifted")
        require(
            implemented_route.get("status") == "fake_store_backed_handler_implemented",
            f"{route_id} fake store implementation status drifted",
        )
        require(
            operation.get("tenant_predicate") == "tenant_ref_from_auth_context",
            f"{route_id} tenant predicate must come from auth context",
        )
        projection = str(operation.get("projection") or "")
        require(projection.endswith("_summary_projection"), f"{route_id} projection must be a summary projection")
        require(isinstance(operation.get("allowed_filters"), list), f"{route_id} allowed_filters must be a list")
        require(isinstance(operation.get("allowed_sort"), list), f"{route_id} allowed_sort must be a list")


def assert_projection_cursor_and_failures(fixture: dict[str, Any]) -> None:
    projection = fixture.get("sanitized_projection_policy") or {}
    require(
        projection.get("status") == "required_before_repository_implementation",
        "projection policy must remain pre-implementation",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_OUTPUT_FIELDS - set(projection.get("forbidden_output_fields") or []))
    require(not missing_forbidden, f"missing forbidden output fields: {missing_forbidden}")
    require(projection.get("projection_must_be_route_owned") is True, "projection must be route owned")
    require(projection.get("projection_must_be_deny_by_default") is True, "projection must deny by default")

    cursor = fixture.get("cursor_filter_sort_policy") or {}
    require(cursor.get("cursor_format") == "opaque_route_owned_cursor", "cursor format drifted")
    require(cursor.get("default_limit") == 25, "default limit drifted")
    require(cursor.get("max_limit") == 100, "max limit drifted")
    require(cursor.get("unknown_filter_behavior") == "fail_closed_invalid_filter", "filter behavior drifted")
    require(cursor.get("unknown_sort_behavior") == "fail_closed_invalid_filter", "sort behavior drifted")
    require(cursor.get("tenant_ref_from_query_string_allowed") is False, "tenant query override must be forbidden")
    require(cursor.get("sensitive_projection_query_allowed") is False, "sensitive projection query must be forbidden")

    failures = {
        str(item.get("failure_code") or "")
        for item in fixture.get("failure_mapping") or []
        if isinstance(item, dict)
    }
    missing_failures = sorted(EXPECTED_FAILURE_CODES - failures)
    require(not missing_failures, f"missing failure mappings: {missing_failures}")


def assert_contract_smoke_and_policy_docs(fixture: dict[str, Any]) -> None:
    smoke = fixture.get("contract_smoke_plan") or {}
    require(
        smoke.get("status") == "preconditions_defined_not_implemented",
        "contract smoke must remain precondition-only",
    )
    required_sources = {
        "repository_contract_fake_store_smoke",
        "read_store_repository_contract_smoke",
        "database_read_disabled_guard_smoke",
    }
    require(required_sources.issubset(set(smoke.get("future_smoke_sources") or [])), "missing smoke sources")
    require(
        "repository_write_count=0" in set(smoke.get("side_effect_counters_must_remain") or []),
        "missing zero write counter",
    )

    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden scope items: {missing_scope}")
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-repository-contract-preconditions-v1.py", [])'
        in check_repo,
        "check-repo.py must run repository contract preconditions check",
    )
    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_implementation_leaked() -> None:
    source_absent_literals = source_absent_literals_for_current_stage()
    source_roots = [
        REPO_ROOT / "services/platform/internal",
        REPO_ROOT / "apps/radishmind-web/src",
    ]
    for root in source_roots:
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in source_absent_literals:
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this preconditions slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_interface_contract(fixture)
    assert_repository_operations(fixture)
    assert_projection_cursor_and_failures(fixture)
    assert_contract_smoke_and_policy_docs(fixture)
    assert_no_implementation_leaked()
    print("control plane read repository contract preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
