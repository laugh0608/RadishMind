#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-implementation-readiness-v1.json"
)
REPOSITORY_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-v1.json"
)
DISABLED_DATABASE_GUARD_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-disabled-database-guard-v1.json"
)
REPOSITORY_PRECONDITIONS_FIXTURE_PATH = (
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
CONTRACT_TYPES_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-types-implementation-v1.json"
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
    "repository_adapter_ready",
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
    "control-plane-read-repository-contract-smoke-v1": (
        REPOSITORY_SMOKE_FIXTURE_PATH,
        "repository_contract_smoke_defined",
    ),
    "control-plane-read-disabled-database-guard-v1": (
        DISABLED_DATABASE_GUARD_FIXTURE_PATH,
        "disabled_database_guard_defined",
    ),
    "control-plane-read-repository-contract-preconditions-v1": (
        REPOSITORY_PRECONDITIONS_FIXTURE_PATH,
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
    "control-plane-read-fake-store-handler-implementation-v1": (
        FAKE_STORE_FIXTURE_PATH,
        "fake_store_handler_implemented",
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
    "tenant_binding_missing",
    "scope_denied",
    "invalid_filter",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "database_read_disabled",
    "auth_context_contract_mismatch",
}
EXPECTED_GATE_IDS = {
    "repository_contract_types_defined",
    "auth_context_injection_ready",
    "store_selection_guard_ready",
    "schema_and_migration_plan_ready",
    "dual_smoke_ready",
    "production_auth_ready",
}
EXPECTED_PLANNED_FILES = {
    "services/platform/internal/httpapi/control_plane_read_repository_contract.go",
    "services/platform/internal/httpapi/control_plane_read_repository_contract_smoke_test.go",
    "services/platform/internal/httpapi/control_plane_read_repository_adapter.go",
    "services/platform/internal/httpapi/control_plane_read_repository_adapter_test.go",
}
CONTROLLED_CONTRACT_TYPE_FILE = "services/platform/internal/httpapi/control_plane_read_repository_contract.go"
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
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
    "type ControlPlaneReadRepository interface",
    "func NewControlPlaneReadRepository",
    "RADISHMIND_CONTROL_PLANE_READ_STORE",
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


def repository_operations_by_route_id() -> dict[str, dict[str, Any]]:
    repository = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH)
    operations = {
        str(operation.get("route_id") or ""): operation
        for operation in repository.get("repository_operations") or []
        if isinstance(operation, dict)
    }
    require(set(operations) == EXPECTED_ROUTE_IDS, "repository operation ids drifted")
    return operations


def smoke_matrix_by_route_id() -> dict[str, dict[str, Any]]:
    smoke = load_json(REPOSITORY_SMOKE_FIXTURE_PATH)
    matrix = {
        str(row.get("route_id") or ""): row
        for row in smoke.get("route_smoke_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "repository smoke route ids drifted")
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
        fixture.get("kind") == "control_plane_read_repository_implementation_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-repository-implementation-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "repository_implementation_readiness_defined",
        "repository implementation readiness status drifted",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_readiness_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("readiness_boundary") or {}
    smoke_boundary = load_json(REPOSITORY_SMOKE_FIXTURE_PATH).get("smoke_boundary") or {}
    guard_boundary = load_json(DISABLED_DATABASE_GUARD_FIXTURE_PATH).get("disabled_database_guard_boundary") or {}
    require(boundary.get("status") == "readiness_defined_not_implemented", "readiness boundary status drifted")
    require(
        boundary.get("current_store_source") == smoke_boundary.get("current_store_source"),
        "current store source must match repository smoke",
    )
    require(
        boundary.get("future_store_source") == smoke_boundary.get("future_store_source"),
        "future store source must match repository smoke",
    )
    require(
        boundary.get("future_store_source") == guard_boundary.get("reserved_disabled_database_source"),
        "future store source must match disabled database guard",
    )
    for field in (
        "implementation_allowed_now",
        "repository_adapter_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_future_file_layout(fixture: dict[str, Any]) -> None:
    layout = fixture.get("future_file_layout") or {}
    require(layout.get("status") == "planned_not_created", "future file layout status drifted")
    require(
        layout.get("package_boundary") == "services/platform/internal/httpapi",
        "future package boundary drifted",
    )
    planned = {
        str(item.get("path") or ""): item
        for item in layout.get("planned_files") or []
        if isinstance(item, dict)
    }
    require(set(planned) == EXPECTED_PLANNED_FILES, "planned repository file layout drifted")
    for path, item in planned.items():
        require(item.get("created_in_this_slice") is False, f"{path} must not be created in this slice")
        if path == CONTROLLED_CONTRACT_TYPE_FILE and (REPO_ROOT / path).exists():
            implementation = load_json(CONTRACT_TYPES_IMPLEMENTATION_FIXTURE_PATH)
            implementation_slice = implementation.get("slice") or {}
            require(
                implementation_slice.get("status") == "repository_contract_types_implemented",
                "contract type file must be covered by repository contract types implementation gate",
            )
            continue
        require(not (REPO_ROOT / path).exists(), f"{path} must not exist before implementation starts")


def assert_implementation_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("implementation_gates") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "implementation gate ids drifted")
    require(gates["schema_and_migration_plan_ready"].get("status") == "not_satisfied", "schema gate must remain blocked")
    require(gates["production_auth_ready"].get("status") == "not_satisfied", "production auth gate must remain blocked")
    for gate_id in EXPECTED_GATE_IDS - {"schema_and_migration_plan_ready", "production_auth_ready"}:
        require(
            gates[gate_id].get("status") == "required_before_implementation",
            f"{gate_id} status drifted",
        )
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage requirements")


def assert_route_readiness_matrix(fixture: dict[str, Any]) -> None:
    operations = repository_operations_by_route_id()
    smoke = smoke_matrix_by_route_id()
    matrix = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_readiness_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "route readiness matrix must cover every read route")
    for route_id, row in matrix.items():
        require(row.get("operation") == operations[route_id].get("operation"), f"{route_id} operation drifted")
        require(row.get("operation") == smoke[route_id].get("operation"), f"{route_id} smoke operation drifted")
        require(
            row.get("first_allowed_future_mode") == "repository_contract_smoke_runner",
            f"{route_id} first future mode drifted",
        )
        require(row.get("durable_adapter_allowed_now") is False, f"{route_id} durable adapter must remain blocked")


def assert_dual_smoke_and_failure_policy(fixture: dict[str, Any]) -> None:
    dual_smoke = fixture.get("dual_smoke_plan") or {}
    require(dual_smoke.get("status") == "planned_not_implemented", "dual smoke status drifted")
    required_retained = {
        "control-plane-read-fake-store-handler-implementation-v1",
        "control-plane-read-repository-contract-smoke-v1",
        "control-plane-read-disabled-database-guard-v1",
    }
    require(required_retained.issubset(set(dual_smoke.get("must_retain") or [])), "dual smoke retained gates drifted")
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(dual_smoke.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )

    failure_policy = fixture.get("failure_mapping_policy") or {}
    require(
        failure_policy.get("status") == "required_before_implementation",
        "failure mapping policy status drifted",
    )
    require(
        EXPECTED_FAILURE_CODES.issubset(set(failure_policy.get("required_failure_codes") or [])),
        "failure mapping policy missing required codes",
    )
    smoke_failures = {
        str(item.get("failure_code") or "")
        for item in load_json(REPOSITORY_SMOKE_FIXTURE_PATH).get("failure_mapping") or []
        if isinstance(item, dict)
    }
    require(EXPECTED_FAILURE_CODES.issubset(smoke_failures), "repository smoke fixture must retain failure codes")
    for field in (
        "database_internal_error_exposure_allowed",
        "unknown_filter_success_allowed",
        "missing_tenant_success_allowed",
    ):
        require(failure_policy.get(field) is False, f"{field} must remain false")


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-repository-implementation-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run repository implementation readiness check",
    )
    require(
        check_repo.index("check-control-plane-read-repository-contract-smoke-v1.py")
        < check_repo.index("check-control-plane-read-repository-implementation-readiness-v1.py"),
        "repository implementation readiness check must run after repository contract smoke",
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
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this readiness slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_readiness_boundary(fixture)
    assert_future_file_layout(fixture)
    assert_implementation_gates(fixture)
    assert_route_readiness_matrix(fixture)
    assert_dual_smoke_and_failure_policy(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_implementation_leaked(fixture)
    print("control plane read repository implementation readiness v1 checks passed.")


if __name__ == "__main__":
    main()
