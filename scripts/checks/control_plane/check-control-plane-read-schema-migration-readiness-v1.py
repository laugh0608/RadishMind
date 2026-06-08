#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-migration-readiness-v1.json"
STORE_SELECTION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selection-readiness-v1.json"
)
IMPLEMENTATION_READINESS_FIXTURE_PATH = (
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
DATA_BOUNDARY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json"
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
EXPECTED_LOGICAL_ENTITIES = {
    "tenant_summary",
    "application_summary",
    "api_key_summary",
    "quota_summary",
    "workflow_definition_summary",
    "run_record_summary",
    "audit_summary",
}
EXPECTED_DEPENDENCIES = {
    "control-plane-read-store-selection-readiness-v1": (
        STORE_SELECTION_FIXTURE_PATH,
        "store_selection_readiness_defined",
    ),
    "control-plane-read-repository-implementation-readiness-v1": (
        IMPLEMENTATION_READINESS_FIXTURE_PATH,
        "repository_implementation_readiness_defined",
    ),
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
    "control-plane-data-boundary": (
        DATA_BOUNDARY_FIXTURE_PATH,
        "governance_boundary_satisfied",
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
EXPECTED_FORBIDDEN_CLAIMS = {
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "migration_files_created",
    "repository_implementation_ready",
    "repository_adapter_ready",
    "repository_migration_ready",
    "store_selector_implemented",
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
EXPECTED_GATE_IDS = {
    "schema_ownership_defined",
    "migration_layout_reserved",
    "rollback_plan_defined",
    "tenant_index_strategy_defined",
    "read_only_role_policy_defined",
    "migration_smoke_defined",
}
EXPECTED_FAILURE_CODES = {
    "database_read_disabled",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "schema_migration_not_applied",
    "schema_version_mismatch",
}
EXPECTED_SOURCE_ABSENT_LITERALS = {
    "database/sql",
    "CREATE TABLE",
    "ALTER TABLE",
    "CREATE INDEX",
    "DROP TABLE",
    "SELECT *",
    "INSERT INTO",
    "UPDATE ",
    "DELETE FROM",
    "services/platform/migrations/control_plane_read",
    "type ControlPlaneReadRepository interface",
    "func NewControlPlaneReadRepository",
    "SelectControlPlaneReadStore",
    "ControlPlaneReadStoreSelector",
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


def store_selection_by_route_id() -> dict[str, dict[str, Any]]:
    store_selection = load_json(STORE_SELECTION_FIXTURE_PATH)
    matrix = {
        str(row.get("route_id") or ""): row
        for row in store_selection.get("route_selection_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "store selection route ids drifted")
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
        fixture.get("kind") == "control_plane_read_schema_migration_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-schema-migration-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "schema_migration_readiness_defined",
        "schema migration readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_schema_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("schema_boundary") or {}
    store_selection = load_json(STORE_SELECTION_FIXTURE_PATH).get("selection_boundary") or {}
    implementation = load_json(IMPLEMENTATION_READINESS_FIXTURE_PATH).get("readiness_boundary") or {}
    require(
        boundary.get("status") == "schema_migration_readiness_defined_not_implemented",
        "schema boundary status drifted",
    )
    require(
        boundary.get("current_store_source") == store_selection.get("current_default_source"),
        "current store source must match store selection default",
    )
    require(
        boundary.get("future_durable_source") == store_selection.get("future_durable_source"),
        "future durable source must match store selection",
    )
    require(
        boundary.get("future_durable_source") == implementation.get("future_store_source"),
        "future durable source must match repository implementation readiness",
    )
    for field in (
        "schema_files_created_in_this_slice",
        "migration_files_created_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_planned_migration_layout(fixture: dict[str, Any]) -> None:
    layout = fixture.get("planned_migration_layout") or {}
    require(layout.get("status") == "planned_not_created", "migration layout status drifted")
    root = str(layout.get("migration_root") or "")
    require(root == "services/platform/migrations/control_plane_read", "migration root drifted")
    require(not (REPO_ROOT / root).exists(), "migration root must not be created in this slice")
    planned = {
        str(item.get("path") or ""): item
        for item in layout.get("planned_files") or []
        if isinstance(item, dict)
    }
    expected_paths = {
        "services/platform/migrations/control_plane_read/README.md",
        "services/platform/migrations/control_plane_read/00000000000000_control_plane_read_schema_placeholder.sql",
    }
    require(set(planned) == expected_paths, "planned migration files drifted")
    for path, item in planned.items():
        require(item.get("created_in_this_slice") is False, f"{path} must not be created in this slice")
        require(not (REPO_ROOT / path).exists(), f"{path} must not exist before implementation starts")
    sql_files = list((REPO_ROOT / "services/platform").rglob("*.sql"))
    require(not sql_files, f"SQL files must not be introduced in this readiness slice: {sql_files}")


def assert_schema_ownership(fixture: dict[str, Any]) -> None:
    ownership = {
        str(item.get("entity") or ""): item
        for item in fixture.get("schema_ownership") or []
        if isinstance(item, dict)
    }
    require(set(ownership) == EXPECTED_LOGICAL_ENTITIES, "schema ownership entities drifted")
    for entity, item in ownership.items():
        require(item.get("owner_surface"), f"{entity} must define owner_surface")
        require(item.get("source_of_truth"), f"{entity} must define source_of_truth")
        require(item.get("tenant_ref_required") is True, f"{entity} must require tenant_ref")
        require(item.get("write_boundary"), f"{entity} must define write_boundary")


def assert_route_schema_matrix(fixture: dict[str, Any]) -> None:
    operations = repository_operations_by_route_id()
    store_selection = store_selection_by_route_id()
    matrix = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_schema_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "route schema matrix must cover every read route")
    seen_entities = set()
    for route_id, row in matrix.items():
        require(row.get("operation") == operations[route_id].get("operation"), f"{route_id} operation drifted")
        require(row.get("operation") == store_selection[route_id].get("operation"), f"{route_id} store operation drifted")
        entity = str(row.get("logical_entity") or "")
        require(entity in EXPECTED_LOGICAL_ENTITIES, f"{route_id} logical entity drifted")
        seen_entities.add(entity)
        for field in ("tenant_index_required", "read_only_role_required", "migration_smoke_required"):
            require(row.get(field) is True, f"{route_id} {field} must remain true")
    require(seen_entities == EXPECTED_LOGICAL_ENTITIES, "route schema matrix must cover every logical entity")


def assert_readiness_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("readiness_gates") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "schema readiness gate ids drifted")
    for gate_id, gate in gates.items():
        require(
            gate.get("status") == "required_before_schema_implementation",
            f"{gate_id} status drifted",
        )
        require(gate.get("must_cover"), f"{gate_id} must list coverage requirements")


def assert_policies(fixture: dict[str, Any]) -> None:
    migration_policy = fixture.get("migration_policy") or {}
    require(migration_policy.get("status") == "policy_defined_not_implemented", "migration policy status drifted")
    require(migration_policy.get("auto_migrate_on_service_start_allowed") is False, "startup migration must be disabled")
    require(
        migration_policy.get("destructive_migration_allowed_without_review") is False,
        "destructive migration must require review",
    )
    for field in ("rollback_required_before_apply", "backup_required_before_apply", "migration_lock_required", "schema_version_table_required"):
        require(migration_policy.get(field) is True, f"{field} must remain true")

    role_policy = fixture.get("read_only_role_policy") or {}
    require(role_policy.get("status") == "policy_defined_not_implemented", "read-only role policy status drifted")
    require(role_policy.get("future_role_name") == "radishmind_control_plane_readonly", "read-only role name drifted")
    for field in ("insert_allowed", "update_allowed", "delete_allowed", "ddl_allowed", "migration_apply_allowed"):
        require(role_policy.get(field) is False, f"{field} must remain false")
    require(role_policy.get("credential_storage") == "future secret reference only", "credential storage policy drifted")

    failure_policy = fixture.get("failure_mapping_policy") or {}
    require(
        failure_policy.get("status") == "required_before_schema_implementation",
        "failure mapping policy status drifted",
    )
    require(
        EXPECTED_FAILURE_CODES.issubset(set(failure_policy.get("required_failure_codes") or [])),
        "failure mapping policy missing required codes",
    )
    for field in (
        "migration_internal_error_exposure_allowed",
        "missing_tenant_success_allowed",
        "schema_mismatch_success_allowed",
    ):
        require(failure_policy.get(field) is False, f"{field} must remain false")


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-control-plane-read-schema-migration-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run schema migration readiness check")
    require(
        check_repo.index("check-control-plane-read-store-selection-readiness-v1.py")
        < check_repo.index(current_checker),
        "schema migration readiness check must run after store selection readiness",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_schema_implementation_leaked(fixture: dict[str, Any]) -> None:
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
    assert_schema_boundary(fixture)
    assert_planned_migration_layout(fixture)
    assert_schema_ownership(fixture)
    assert_route_schema_matrix(fixture)
    assert_readiness_gates(fixture)
    assert_policies(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_schema_implementation_leaked(fixture)
    print("control plane read schema migration readiness v1 checks passed.")


if __name__ == "__main__":
    main()
