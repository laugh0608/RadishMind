#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json"
PLAN_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-plan-v1.json"
PRECONDITIONS_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-preconditions-v1.json"
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
RESPONSE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_IMPLEMENTED_ROUTE_IDS = {
    "tenant-summary-route",
    "application-summary-list-route",
    "api-key-summary-list-route",
    "quota-summary-route",
    "workflow-definition-summary-list-route",
    "run-record-summary-list-route",
    "audit-summary-list-route",
}
EXPECTED_REMAINING_ROUTE_IDS: set[str] = set()
EXPECTED_ROUTE_REGISTRATION_LITERALS = {
    "tenant-summary-route": "controlPlaneTenantSummaryRoute",
    "application-summary-list-route": "controlPlaneApplicationSummaryListRoute",
    "api-key-summary-list-route": "controlPlaneAPIKeySummaryListRoute",
    "quota-summary-route": "controlPlaneQuotaSummaryRoute",
    "workflow-definition-summary-list-route": "controlPlaneWorkflowDefinitionSummaryListRoute",
    "run-record-summary-list-route": "controlPlaneRunRecordSummaryListRoute",
    "audit-summary-list-route": "controlPlaneAuditSummaryListRoute",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "control_plane_api_ready",
    "full_read_side_ready",
    "typescript_consumer_ready",
    "database_schema_ready",
    "database_query_ready",
    "radish_oidc_client_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "workflow_executor_ready",
    "production_ready",
}
EXPECTED_FORBIDDEN_STORE_MODES = {
    "database_query",
    "database_migration",
    "durable_store",
    "production_data_source",
    "secret_value_storage",
    "provider_call",
    "workflow_executor",
    "business_writeback",
}
EXPECTED_FORBIDDEN_PUBLIC_AUTH_INPUTS = {
    "public_fake_auth_header",
    "query_scope_override",
    "anonymous_read",
    "cross_tenant_read",
    "ops_console_elevation",
}
EXPECTED_SMOKE_COVERAGE = {
    "tenant_summary_success",
    "application_summary_list_success",
    "api_key_summary_list_success",
    "quota_summary_success",
    "workflow_definition_summary_list_success",
    "run_record_summary_list_success",
    "audit_summary_list_success",
    "missing_identity",
    "tenant_binding_missing",
    "scope_denied",
    "cross_tenant_query_denied",
    "invalid_filter",
    "forbidden_sensitive_projection",
    "forbidden_method",
    "forbidden_query",
    "no_side_effects",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "fake_store_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "database schema",
    "database migration",
    "database query implementation",
    "durable store",
    "Radish OIDC integration",
    "public fake auth header",
    "tenant or user CRUD",
    "API key generation",
    "quota enforcement",
    "rate limiting implementation",
    "billing ledger implementation",
    "workflow builder",
    "workflow executor",
    "confirmation flow wiring",
    "business writeback",
    "replay execution",
    "production admin console",
    "user workspace UI",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "fake-store-backed read handler implementation",
        "不实现完整 read-side API",
    ],
    "docs/contracts/control-plane-read-side.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "application-summary-list-route",
        "audit-summary-list-route",
    ],
    "docs/radishmind-current-focus.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "fake-store-backed read handler implementation",
        "不直接实现数据库、OIDC、executor、confirmation、writeback 或 replay",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "services/platform/internal/httpapi",
        "fake store",
    ],
    "docs/radishmind-roadmap.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "fake-store-backed read handler implementation",
        "不默认进入数据库",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "fake-store-backed",
        "handler implementation",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "Control Plane Read Fake-Store Handler Implementation",
    ],
    "docs/task-cards/control-plane-read-fake-store-handler-implementation-v1-plan.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "cursor list",
        "audit-summary-list-route",
    ],
    "scripts/README.md": [
        "check-control-plane-read-fake-store-handler-implementation-v1.py",
        "control-plane-read-fake-store-handler-implementation-v1.json",
        "fake-store-backed read handler implementation",
    ],
    "services/platform/README.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "/v1/user-workspace/applications",
        "/v1/control-plane/audit",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-fake-store-handler-implementation-v1",
        "control-plane-read-fake-store-handler-implementation-v1.json",
        "check-control-plane-read-fake-store-handler-implementation-v1.py",
    ],
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
    require(routes, "route contract fixture must contain routes")
    return routes


def response_route_ids() -> set[str]:
    response_fixture = load_json(RESPONSE_FIXTURE_PATH)
    route_ids = {
        str(example.get("route_id") or "")
        for example in response_fixture.get("response_examples") or []
        if isinstance(example, dict)
    }
    require(route_ids, "response fixture must expose route examples")
    return route_ids


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "control-plane-read-fake-store-handler-plan-v1",
        "control-plane-read-implementation-preconditions-v1",
        "control-plane-read-route-contract-v1",
        "control-plane-read-response-fixtures-v1",
        "control-plane-read-negative-contract-v1",
    }
    require(expected.issubset(declared), "implementation fixture must declare read-side dependencies")
    for expected_id, path in {
        "control-plane-read-fake-store-handler-plan-v1": PLAN_FIXTURE_PATH,
        "control-plane-read-implementation-preconditions-v1": PRECONDITIONS_FIXTURE_PATH,
        "control-plane-read-route-contract-v1": ROUTE_CONTRACT_FIXTURE_PATH,
        "control-plane-read-response-fixtures-v1": RESPONSE_FIXTURE_PATH,
        "control-plane-read-negative-contract-v1": NEGATIVE_CONTRACT_FIXTURE_PATH,
    }.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == expected_id, f"dependency {expected_id} id drifted")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "control_plane_read_fake_store_handler_implementation_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-fake-store-handler-implementation-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "fake_store_handler_implemented",
        "implementation slice must stay fake-store handler implementation",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_route_sets(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    response_ids = response_route_ids()
    implemented = {
        str(item.get("route_id") or ""): item
        for item in fixture.get("implemented_routes") or []
        if isinstance(item, dict)
    }
    remaining = {
        str(item.get("route_id") or ""): item
        for item in fixture.get("remaining_routes") or []
        if isinstance(item, dict)
    }
    require(set(implemented) == EXPECTED_IMPLEMENTED_ROUTE_IDS, "implemented route ids drifted")
    require(set(remaining) == EXPECTED_REMAINING_ROUTE_IDS, "remaining route ids drifted")
    require(set(implemented) | set(remaining) == set(routes), "implementation fixture must cover every route contract")
    require(set(implemented).issubset(response_ids), "implemented routes must have response examples")

    for route_id, implementation in implemented.items():
        route = routes[route_id]
        require(implementation.get("method") == route.get("method"), f"{route_id} method drifted")
        require(implementation.get("path") == route.get("path"), f"{route_id} path drifted")
        require(implementation.get("read_model") == route.get("read_model"), f"{route_id} read model drifted")
        require(implementation.get("scope") == route.get("required_scope"), f"{route_id} scope drifted")
        require(implementation.get("status") == "fake_store_backed_handler_implemented", f"{route_id} status drifted")
        expected_mode = "single_resource_summary" if route.get("pagination") == "single_resource" else "cursor_list_summary"
        require(implementation.get("fake_store_mode") == expected_mode, f"{route_id} fake store mode drifted")
        require(implementation.get("auth_mode") == "test_only_context_injection", f"{route_id} auth mode drifted")

    for route_id, route in remaining.items():
        require(route.get("status") == "planned_not_implemented", f"{route_id} must remain planned")


def assert_go_files_and_routes(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    for relative_path in fixture.get("go_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing Go file: {relative_path}")

    handler_go = read("services/platform/internal/httpapi/control_plane_read.go")
    fake_store_go = read("services/platform/internal/httpapi/control_plane_read_fake_store.go")
    test_go = read("services/platform/internal/httpapi/control_plane_read_test.go")
    server_go = read("services/platform/internal/httpapi/server.go")
    observability_go = read("services/platform/internal/httpapi/observability.go")

    for literal in (
        "handleControlPlaneTenantSummary",
        "handleUserWorkspaceApplicationSummaryList",
        "handleUserWorkspaceAPIKeySummaryList",
        "handleUserWorkspaceQuotaSummary",
        "handleUserWorkspaceWorkflowDefinitionSummaryList",
        "handleUserWorkspaceRunRecordSummaryList",
        "handleControlPlaneAuditSummaryList",
        "handleControlPlaneReadCursorList",
        "authorizeControlPlaneReadRequest",
        "forbiddenControlPlaneReadQueryParameter",
        "controlPlaneReadFiltersFromQuery",
        "withControlPlaneReadFakeAuthContext",
        "controlPlaneReadEnvelope",
    ):
        require(literal in handler_go, f"handler file missing {literal}")

    for literal in (
        "newControlPlaneReadFakeStore",
        "tenant_demo",
        "app_flow_copilot",
        "app_docs_assistant",
        "key_ops_readonly",
        "quota_demo_current",
        "wf_radishflow_copilot_latest",
        "wf_radish_docs_qa_draft",
        "run_radishflow_copilot_20260531_001",
        "run_radish_docs_qa_20260531_002",
        "audit_admin_policy_read_20260601_001",
        "controlPlaneReadCursorListSummaries",
        "controlPlaneReadItemMatchesFilters",
        "SideEffects() controlPlaneReadSideEffects",
        "FakeStoreWriteCount",
        "ExecutorCallCount",
        "BusinessWritebackCount",
        "ReplayCallCount",
    ):
        require(literal in fake_store_go, f"fake store file missing {literal}")

    for forbidden in (
        "database/sql",
        "gorm",
        "OIDC",
        "api key generation",
        "workflow executor",
        "business_writeback_payload",
        "full_prompt_dump_with_secret",
    ):
        require(forbidden not in handler_go + fake_store_go, f"implementation must not include forbidden dependency: {forbidden}")

    route_registration = fixture.get("route_registration") or {}
    require(route_registration.get("file") == "services/platform/internal/httpapi/server.go", "route registration file drifted")
    registered_route_ids = set(route_registration.get("registered_route_ids") or [])
    forbidden_route_ids = set(route_registration.get("forbidden_registered_route_ids") or [])
    require(registered_route_ids == EXPECTED_IMPLEMENTED_ROUTE_IDS, "registered route ids drifted")
    require(forbidden_route_ids == EXPECTED_REMAINING_ROUTE_IDS, "forbidden registered route ids drifted")

    for route_id in registered_route_ids:
        path = str(routes[route_id].get("path") or "")
        registration_literal = EXPECTED_ROUTE_REGISTRATION_LITERALS[route_id]
        require(path in handler_go, f"implemented route path must be declared: {path}")
        require(registration_literal in server_go, f"implemented route must be registered: {route_id}")
    for route_id in forbidden_route_ids:
        path = str(routes[route_id].get("path") or "")
        require(path not in server_go, f"remaining route path must not be registered: {path}")

    for literal in (
        "CONTROL_PLANE_READ_METHOD_NOT_ALLOWED",
        "CONTROL_PLANE_READ_QUERY_FORBIDDEN",
    ):
        require(literal in observability_go, f"observability missing platform error code: {literal}")

    for literal in (
        "TestControlPlaneReadFakeStoreRoutes",
        "tenant summary succeeds",
        "cursor list routes succeed",
        "application summaries",
        "api key summaries",
        "quota summary succeeds",
        "workflow definition summaries",
        "run record summaries",
        "audit summaries",
        "identity_context_missing",
        "tenant_binding_missing",
        "scope_denied",
        "cross-tenant",
        "invalid_filter",
        "forbidden sensitive projection",
        "CONTROL_PLANE_READ_METHOD_NOT_ALLOWED",
        "CONTROL_PLANE_READ_QUERY_FORBIDDEN",
        "assertControlPlaneReadNoForbiddenPayload",
    ):
        require(literal in test_go, f"route smoke test missing {literal}")


def assert_fake_store_auth_and_smoke(fixture: dict[str, Any]) -> None:
    fake_store = fixture.get("fake_store") or {}
    require(fake_store.get("status") == "implemented_fixture_backed_only", "fake store status drifted")
    require(fake_store.get("data_source") == "in_memory_fixture", "fake store data source drifted")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_STORE_MODES - set(fake_store.get("does_not_allow") or []))
    require(not missing_forbidden, f"fake store missing forbidden modes: {missing_forbidden}")
    missing_counters = sorted(EXPECTED_SIDE_EFFECT_COUNTERS - set(fake_store.get("side_effect_counters") or []))
    require(not missing_counters, f"fake store missing no-side-effect counters: {missing_counters}")

    auth = fixture.get("auth") or {}
    require(auth.get("status") == "test_only_context_injection", "auth status drifted")
    require(auth.get("injection_mode") == "go_test_request_context_only", "auth injection mode drifted")
    require("Radish OIDC" in str(auth.get("production_truth_source") or ""), "auth production truth source drifted")
    missing_auth = sorted(EXPECTED_FORBIDDEN_PUBLIC_AUTH_INPUTS - set(auth.get("forbidden_public_inputs") or []))
    require(not missing_auth, f"auth missing forbidden public inputs: {missing_auth}")

    missing_smoke = sorted(EXPECTED_SMOKE_COVERAGE - set(fixture.get("smoke_coverage") or []))
    require(not missing_smoke, f"implementation missing smoke coverage: {missing_smoke}")


def assert_policy_and_docs(fixture: dict[str, Any]) -> None:
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-fake-store-handler-implementation-v1.py", [])'
        in check_repo,
        "check-repo.py must run control plane read fake-store handler implementation check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_route_sets(fixture)
    assert_go_files_and_routes(fixture)
    assert_fake_store_auth_and_smoke(fixture)
    assert_policy_and_docs(fixture)
    print("control plane read fake-store handler implementation v1 checks passed.")


if __name__ == "__main__":
    main()
