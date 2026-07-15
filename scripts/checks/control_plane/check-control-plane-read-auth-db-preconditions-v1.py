#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json"
FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json"
)
IMPLEMENTATION_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-preconditions-v1.json"
)
OIDC_PRECONDITIONS_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json"
DATA_BOUNDARY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json"
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
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
    "control_plane_api_ready",
    "full_read_side_ready",
    "typescript_consumer_ready",
    "radish_oidc_client_ready",
    "auth_middleware_ready",
    "token_validation_ready",
    "database_schema_ready",
    "database_migration_ready",
    "database_query_ready",
    "repository_implementation_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "workflow_executor_ready",
    "production_ready",
}
EXPECTED_AUTH_CONTEXT_FIELDS = {
    "identity_context",
    "tenant_binding",
    "subject_binding",
    "scope_grants",
    "audit_context",
    "issuer_ref",
    "session_ref",
    "request_id",
}
EXPECTED_AUTH_CLAIM_MAPPINGS = {
    "issuer",
    "subject",
    "tenant_ref",
    "scopes",
    "roles",
    "permissions",
}
EXPECTED_FORBIDDEN_PUBLIC_AUTH_INPUTS = {
    "public_fake_auth_header",
    "query_scope_override",
    "anonymous_read",
    "cross_tenant_read",
    "ops_console_elevation",
    "mock_provider_fallback",
}
EXPECTED_ALLOWED_STORE_OPERATIONS = {
    "read_only_select",
    "cursor_list_read",
    "single_resource_read",
}
EXPECTED_QUERY_GUARDS = {
    "tenant_ref_predicate",
    "subject_scope_predicate",
    "limit_clamp",
    "cursor_validation",
    "route_filter_allowlist",
    "sort_allowlist",
    "sanitized_projection",
    "request_id_audit_ref_propagation",
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
EXPECTED_FAILURE_CODES = {
    "identity_context_missing",
    "tenant_binding_missing",
    "scope_denied",
    "tenant_not_found",
    "quota_policy_missing",
    "invalid_filter",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "database_read_disabled",
    "auth_context_contract_mismatch",
}
EXPECTED_SMOKE_CASES = {
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
    "Radish OIDC middleware implementation",
    "token validation implementation",
    "login redirect route",
    "database schema",
    "database migration",
    "database query implementation",
    "repository implementation",
    "durable store",
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
        "control-plane-read-auth-db-preconditions-v1",
        "真实 auth/db 前置条件",
        "不实现完整 read-side API",
    ],
    "docs/contracts/control-plane-read-side.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "auth/db preconditions",
        "future control plane read store repository",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "future control plane read store repository",
    ],
    "docs/radishmind-roadmap.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "真实 auth/db 前置条件",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "auth/db preconditions",
    ],
    "docs/radishmind-product-scope.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "数据库 read path",
    ],
    "docs/radishmind-integration-contracts.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "future control plane read store repository",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "Control Plane Read Auth/DB Preconditions",
    ],
    "docs/task-cards/control-plane-read-auth-db-preconditions-v1-plan.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "future Radish OIDC / auth middleware",
        "future control plane read store repository",
    ],
    "scripts/README.md": [
        "check-control-plane-read-auth-db-preconditions-v1.py",
        "control-plane-read-auth-db-preconditions-v1.json",
        "auth/db preconditions",
    ],
    "services/platform/README.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "future Radish OIDC / auth middleware",
        "future control plane read store repository",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-auth-db-preconditions-v1",
        "control-plane-read-auth-db-preconditions-v1.json",
        "check-control-plane-read-auth-db-preconditions-v1.py",
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
    require(set(routes) == EXPECTED_ROUTE_IDS, "route contract ids drifted")
    return routes


def implemented_routes_by_id() -> dict[str, dict[str, Any]]:
    implementation = load_json(FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH)
    implemented = {
        str(route.get("route_id") or ""): route
        for route in implementation.get("implemented_routes") or []
        if isinstance(route, dict)
    }
    require(set(implemented) == EXPECTED_ROUTE_IDS, "fake-store implementation route ids drifted")
    return implemented


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "control-plane-read-fake-store-handler-implementation-v1",
        "control-plane-read-implementation-preconditions-v1",
        "radish-oidc-client-preconditions",
        "control-plane-data-boundary",
        "control-plane-read-route-contract-v1",
        "control-plane-read-negative-contract-v1",
    }
    require(expected.issubset(declared), "auth/db preconditions fixture must declare every dependency")

    expected_statuses = {
        "control-plane-read-fake-store-handler-implementation-v1": (
            FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH,
            "fake_store_handler_implemented",
        ),
        "control-plane-read-implementation-preconditions-v1": (
            IMPLEMENTATION_PRECONDITIONS_FIXTURE_PATH,
            "implementation_preconditions_defined",
        ),
        "radish-oidc-client-preconditions": (OIDC_PRECONDITIONS_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-data-boundary": (DATA_BOUNDARY_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-route-contract-v1": (ROUTE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-negative-contract-v1": (NEGATIVE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
    }
    for expected_id, (path, expected_status) in expected_statuses.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == expected_id, f"dependency {expected_id} id drifted")
        require(
            slice_info.get("status") == expected_status,
            f"dependency {expected_id} status must be {expected_status}",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_auth_db_preconditions_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-auth-db-preconditions-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "auth_db_preconditions_defined",
        "auth/db preconditions must not claim implementation readiness",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_auth_context_contract(fixture: dict[str, Any]) -> None:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    shared_contract = route_contract.get("shared_contract") or {}
    auth = fixture.get("auth_context_contract") or {}
    require(
        auth.get("status") == "preconditions_defined_not_implemented",
        "auth context contract must remain precondition-only",
    )
    require("Radish OIDC" in str(auth.get("future_source") or ""), "auth future source must remain Radish OIDC")

    missing_fields = sorted(EXPECTED_AUTH_CONTEXT_FIELDS - set(auth.get("required_context_fields") or []))
    require(not missing_fields, f"missing required auth context fields: {missing_fields}")
    missing_claims = sorted(EXPECTED_AUTH_CLAIM_MAPPINGS - set(auth.get("required_claim_mappings") or []))
    require(not missing_claims, f"missing auth claim mappings: {missing_claims}")

    declared_fail_closed = set(auth.get("fail_closed_when_missing") or [])
    route_fail_closed = {
        "scope_grants" if item == "scope_grant" else str(item)
        for item in shared_contract.get("fail_closed_when_missing") or []
    }
    require(
        route_fail_closed.issubset(declared_fail_closed),
        "auth context contract must include route fail-closed requirements",
    )
    require("audit_context" in declared_fail_closed, "auth context contract must fail closed without audit_context")

    forbidden_inputs = set(auth.get("forbidden_public_inputs") or [])
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_PUBLIC_AUTH_INPUTS - forbidden_inputs)
    require(not missing_forbidden, f"missing forbidden auth inputs: {missing_forbidden}")
    require(
        set(shared_contract.get("forbidden_fallbacks") or []).issubset(forbidden_inputs),
        "auth context contract must preserve route forbidden fallbacks",
    )


def assert_database_read_store_preconditions(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    db = fixture.get("database_read_store_preconditions") or {}
    require(
        db.get("status") == "preconditions_defined_not_implemented",
        "database read store must remain precondition-only",
    )
    require(
        db.get("future_source") == "future control plane read store repository",
        "database read store future source drifted",
    )

    repository_contract = {
        str(item.get("id") or ""): item
        for item in db.get("repository_contract") or []
        if isinstance(item, dict)
    }
    for required_id in (
        "repository-interface-before-sql",
        "tenant-predicate-required",
        "sanitized-projection-required",
    ):
        item = repository_contract.get(required_id)
        require(item is not None, f"missing repository contract item: {required_id}")
        require(
            item.get("status") == "required_before_implementation",
            f"{required_id} must remain required before implementation",
        )

    missing_operations = sorted(EXPECTED_ALLOWED_STORE_OPERATIONS - set(db.get("allowed_operations") or []))
    require(not missing_operations, f"missing allowed store operations: {missing_operations}")
    missing_guards = sorted(EXPECTED_QUERY_GUARDS - set(db.get("required_query_guards") or []))
    require(not missing_guards, f"missing query guards: {missing_guards}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_STORE_MODES - set(db.get("forbidden_store_modes") or []))
    require(not missing_forbidden, f"missing forbidden store modes: {missing_forbidden}")

    allowed_outputs = set((route_contract.get("sanitization_policy") or {}).get("allowed_outputs") or [])
    declared_outputs = set(db.get("required_sanitized_outputs") or [])
    require(allowed_outputs.issubset(declared_outputs), "database read store must preserve sanitized output policy")

    bindings = {
        str(item.get("route_id") or ""): item
        for item in db.get("route_bindings") or []
        if isinstance(item, dict)
    }
    require(set(bindings) == set(routes), "database read store bindings must cover every read route")
    for route_id, binding in bindings.items():
        route = routes[route_id]
        require(binding.get("read_model") == route.get("read_model"), f"{route_id} read_model drifted")
        expected_mode = "single_resource_read" if route.get("pagination") == "single_resource" else "cursor_list_read"
        require(binding.get("mode") == expected_mode, f"{route_id} repository mode drifted")
        require(str(binding.get("repository_operation") or "").strip(), f"{route_id} repository operation is required")


def assert_route_transition_requirements(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    implemented = implemented_routes_by_id()
    transitions = {
        str(item.get("route_id") or ""): item
        for item in fixture.get("route_transition_requirements") or []
        if isinstance(item, dict)
    }
    require(set(transitions) == set(routes), "route transition requirements must cover every route")

    for route_id, transition in transitions.items():
        route = routes[route_id]
        implementation = implemented[route_id]
        require(transition.get("method") == route.get("method"), f"{route_id} method drifted")
        require(transition.get("path") == route.get("path"), f"{route_id} path drifted")
        require(transition.get("read_model") == route.get("read_model"), f"{route_id} read_model drifted")
        require(transition.get("required_scope") == route.get("required_scope"), f"{route_id} required scope drifted")
        require(
            transition.get("current_status") == implementation.get("status"),
            f"{route_id} current status must match fake-store implementation",
        )
        require(
            transition.get("future_auth_source") == "future Radish OIDC / auth middleware",
            f"{route_id} future auth source drifted",
        )
        require(
            transition.get("future_store_source") == "future control plane read store repository",
            f"{route_id} future store source drifted",
        )
        for field in (
            "oidc_validation_allowed_in_this_slice",
            "database_query_allowed_in_this_slice",
            "migration_allowed_in_this_slice",
            "write_allowed_in_this_slice",
        ):
            require(transition.get(field) is False, f"{route_id} {field} must remain false")


def assert_failure_taxonomy_and_smoke_plan(fixture: dict[str, Any]) -> None:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    fake_store_implementation = load_json(FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH)
    route_failure_codes = {
        str(code)
        for route in route_contract.get("route_contracts") or []
        if isinstance(route, dict)
        for code in route.get("failure_codes") or []
    }

    failures = {
        str(item.get("code") or ""): item
        for item in fixture.get("failure_taxonomy") or []
        if isinstance(item, dict)
    }
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(failures))
    require(not missing_failures, f"missing failure taxonomy codes: {missing_failures}")
    require(route_failure_codes.issubset(set(failures)), "auth/db taxonomy must preserve route failure codes")
    for failure_code, item in failures.items():
        require(str(item.get("source") or "").strip(), f"{failure_code}.source is required")

    smoke = fixture.get("smoke_transition_plan") or {}
    require(
        smoke.get("status") == "preconditions_defined_not_implemented",
        "smoke transition plan must remain precondition-only",
    )
    require(
        smoke.get("current_smoke_source") == "test_only_fake_auth_context_and_in_memory_fixture_fake_store",
        "current smoke source drifted",
    )
    require(
        {"middleware_backed_auth_context_smoke", "repository_contract_fake_store_smoke"}.issubset(
            set(smoke.get("future_smoke_sources") or [])
        ),
        "future smoke sources must include auth middleware and repository contract smoke",
    )
    missing_cases = sorted(EXPECTED_SMOKE_CASES - set(smoke.get("must_preserve_smoke_cases") or []))
    require(not missing_cases, f"missing preserved smoke cases: {missing_cases}")
    require(
        set(fake_store_implementation.get("smoke_coverage") or []).issubset(
            set(smoke.get("must_preserve_smoke_cases") or [])
        ),
        "auth/db transition must preserve fake-store smoke coverage",
    )
    missing_counters = sorted(EXPECTED_SIDE_EFFECT_COUNTERS - set(smoke.get("side_effect_counters_must_remain") or []))
    require(not missing_counters, f"missing side-effect counters: {missing_counters}")


def assert_policy_and_docs(fixture: dict[str, Any]) -> None:
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-auth-db-preconditions-v1.py", [])'
        in check_repo,
        "check-repo.py must run control plane read auth/db preconditions check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_auth_context_contract(fixture)
    assert_database_read_store_preconditions(fixture)
    assert_route_transition_requirements(fixture)
    assert_failure_taxonomy_and_smoke_plan(fixture)
    assert_policy_and_docs(fixture)
    print("control plane read auth/db preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
