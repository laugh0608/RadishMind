#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/control-plane-read-schema-migration-implementation-preconditions-v1.json"
)
SCHEMA_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-migration-readiness-v1.json"
)
SELECTOR_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selector-enablement-preconditions-v1.json"
)
ADAPTER_REFRESH_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-adapter-implementation-readiness-refresh-v1.json"
)
STORE_SELECTION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selection-readiness-v1.json"
)
DISABLED_DATABASE_GUARD_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-disabled-database-guard-v1.json"
)
INTERFACE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-interface-readiness-v1.json"
)
TYPES_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-types-implementation-v1.json"
)
RUNNER_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-implementation-v1.json"
)
AUTH_STORE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-store-transition-preconditions-v1.json"
)
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
EXPECTED_DEPENDENCIES = {
    "control-plane-read-schema-migration-readiness-v1": (
        SCHEMA_READINESS_FIXTURE_PATH,
        "schema_migration_readiness_defined",
    ),
    "control-plane-read-store-selector-enablement-preconditions-v1": (
        SELECTOR_PRECONDITIONS_FIXTURE_PATH,
        "store_selector_enablement_preconditions_defined",
    ),
    "control-plane-read-repository-adapter-implementation-readiness-refresh-v1": (
        ADAPTER_REFRESH_FIXTURE_PATH,
        "repository_adapter_implementation_readiness_refreshed",
    ),
    "control-plane-read-store-selection-readiness-v1": (
        STORE_SELECTION_FIXTURE_PATH,
        "store_selection_readiness_defined",
    ),
    "control-plane-read-disabled-database-guard-v1": (
        DISABLED_DATABASE_GUARD_FIXTURE_PATH,
        "disabled_database_guard_defined",
    ),
    "control-plane-read-repository-interface-readiness-v1": (
        INTERFACE_READINESS_FIXTURE_PATH,
        "repository_interface_readiness_defined",
    ),
    "control-plane-read-repository-contract-types-implementation-v1": (
        TYPES_IMPLEMENTATION_FIXTURE_PATH,
        "repository_contract_types_implemented",
    ),
    "control-plane-read-repository-contract-smoke-runner-implementation-v1": (
        RUNNER_IMPLEMENTATION_FIXTURE_PATH,
        "repository_contract_smoke_runner_implemented",
    ),
    "control-plane-read-auth-store-transition-preconditions-v1": (
        AUTH_STORE_FIXTURE_PATH,
        "auth_store_transition_preconditions_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "migration_files_created",
    "migration_runner_ready",
    "repository_interface_ready",
    "repository_adapter_ready",
    "repository_implementation_ready",
    "repository_migration_ready",
    "store_selector_implemented",
    "formal_read_store_config_ready",
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
    "schema_readiness_consumed",
    "selector_enablement_preconditions_consumed",
    "adapter_readiness_refresh_consumed",
    "migration_artifact_manifest_gate",
    "ddl_review_gate",
    "rollback_fixture_gate",
    "schema_version_smoke_gate",
    "tenant_index_smoke_gate",
    "read_only_role_smoke_gate",
    "no_migration_artifacts_leaked",
}
EXPECTED_FAILURE_CODES = {
    "database_read_disabled",
    "invalid_read_store_mode",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "schema_migration_not_applied",
    "schema_version_mismatch",
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
    "ControlPlaneReadSchemaMigrationRunner",
    "control_plane_read_schema_migration_runner.go",
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
    "control_plane_read_repository_interface.go",
    "control_plane_read_repository_adapter.go",
    "RADISHMIND_CONTROL_PLANE_READ_STORE",
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


def rows_by_route(fixture: dict[str, Any], key: str) -> dict[str, dict[str, Any]]:
    rows = {
        str(row.get("route_id") or ""): row
        for row in fixture.get(key) or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_ROUTE_IDS, f"{key} must cover every read route")
    return rows


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
        fixture.get("kind") == "control_plane_read_schema_migration_implementation_preconditions_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-schema-migration-implementation-preconditions-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "schema_migration_implementation_preconditions_defined",
        "schema migration implementation preconditions status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_implementation_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_preconditions_boundary") or {}
    schema_boundary = load_json(SCHEMA_READINESS_FIXTURE_PATH).get("schema_boundary") or {}
    store_selection = load_json(STORE_SELECTION_FIXTURE_PATH).get("selection_boundary") or {}
    layout = load_json(SCHEMA_READINESS_FIXTURE_PATH).get("planned_migration_layout") or {}
    require(
        boundary.get("status") == "schema_migration_implementation_preconditions_defined_not_implemented",
        "implementation boundary status drifted",
    )
    require(
        boundary.get("current_store_source") == schema_boundary.get("current_store_source"),
        "current store source must match schema readiness",
    )
    require(
        boundary.get("current_store_source") == store_selection.get("current_default_source"),
        "current store source must match store selection",
    )
    require(
        boundary.get("future_migration_root") == layout.get("migration_root"),
        "future migration root must match schema readiness layout",
    )
    require(
        boundary.get("future_runner_name") == "ControlPlaneReadSchemaMigrationRunner",
        "future runner name drifted",
    )
    for field in (
        "migration_root_created_in_this_slice",
        "manifest_created_in_this_slice",
        "sql_files_created_in_this_slice",
        "migration_runner_created_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_future_artifact_manifest(fixture: dict[str, Any]) -> None:
    manifest = fixture.get("future_artifact_manifest") or {}
    require(manifest.get("status") == "planned_not_created", "future artifact manifest status drifted")
    require(manifest.get("required_before_migration_files") is True, "manifest gate must be required")
    required_fields = set(manifest.get("manifest_must_include") or [])
    for field in (
        "migration id",
        "up migration path",
        "down or rollback policy",
        "schema version",
        "affected logical entities",
        "tenant index smoke cases",
        "read-only role smoke cases",
        "failure mapping smoke cases",
    ):
        require(field in required_fields, f"manifest missing required field: {field}")

    artifacts = {
        str(item.get("path") or ""): item
        for item in manifest.get("planned_artifacts") or []
        if isinstance(item, dict)
    }
    expected_artifacts = {
        "services/platform/migrations/control_plane_read/README.md",
        "services/platform/migrations/control_plane_read/manifest.json",
        "services/platform/migrations/control_plane_read/20260604000000_control_plane_read_initial_schema.up.sql",
        "services/platform/migrations/control_plane_read/20260604000000_control_plane_read_initial_schema.down.sql",
    }
    require(set(artifacts) == expected_artifacts, "planned migration artifacts drifted")
    for path, item in artifacts.items():
        require(item.get("created_in_this_slice") is False, f"{path} must not be created in this slice")
        require(not (REPO_ROOT / path).exists(), f"{path} must not exist before implementation starts")
    root = REPO_ROOT / "services/platform/migrations/control_plane_read"
    require(not root.exists(), "migration root must not be created in this preconditions slice")
    sql_files = list((REPO_ROOT / "services/platform").rglob("*.sql"))
    require(not sql_files, f"SQL files must not be introduced in this preconditions slice: {sql_files}")


def assert_gate_matrix(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("implementation_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "implementation gate ids drifted")
    for gate_id in (
        "schema_readiness_consumed",
        "selector_enablement_preconditions_consumed",
        "adapter_readiness_refresh_consumed",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must remain satisfied")
        require(gates[gate_id].get("evidence"), f"{gate_id} must include evidence")
    for gate_id in (
        "migration_artifact_manifest_gate",
        "ddl_review_gate",
        "rollback_fixture_gate",
        "schema_version_smoke_gate",
        "tenant_index_smoke_gate",
        "read_only_role_smoke_gate",
        "no_migration_artifacts_leaked",
    ):
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage requirements")
    require(gates["migration_artifact_manifest_gate"].get("status") == "required_now", "manifest gate status drifted")
    require(gates["no_migration_artifacts_leaked"].get("status") == "required_now", "leak gate status drifted")
    for gate_id in (
        "ddl_review_gate",
        "rollback_fixture_gate",
        "schema_version_smoke_gate",
        "tenant_index_smoke_gate",
        "read_only_role_smoke_gate",
    ):
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")


def assert_route_matrix(fixture: dict[str, Any]) -> None:
    schema_rows = rows_by_route(load_json(SCHEMA_READINESS_FIXTURE_PATH), "route_schema_matrix")
    adapter_rows = rows_by_route(load_json(ADAPTER_REFRESH_FIXTURE_PATH), "adapter_route_matrix")
    rows = rows_by_route(fixture, "route_migration_implementation_matrix")
    for route_id, row in rows.items():
        require(row.get("operation") == schema_rows[route_id].get("operation"), f"{route_id} operation drifted")
        require(row.get("operation") == adapter_rows[route_id].get("operation"), f"{route_id} adapter operation drifted")
        require(
            row.get("logical_entity") == schema_rows[route_id].get("logical_entity"),
            f"{route_id} logical entity drifted",
        )
        for field in (
            "schema_version_smoke_required",
            "tenant_index_smoke_required",
            "read_only_role_smoke_required",
            "projection_smoke_required",
            "write_smoke_forbidden",
        ):
            require(row.get(field) is True, f"{route_id} {field} must remain true")
        require(
            row.get("migration_implementation_allowed_now") is False,
            f"{route_id} migration implementation must remain blocked",
        )


def assert_failure_and_safety_policies(fixture: dict[str, Any]) -> None:
    require(
        EXPECTED_FAILURE_CODES.issubset(set(fixture.get("failure_mapping") or [])),
        "failure mapping missing required codes",
    )

    fallback_policy = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "migration_missing_fake_fallback_allowed",
        "schema_version_mismatch_success_allowed",
        "tenant_index_missing_success_allowed",
        "read_only_role_failure_success_allowed",
        "database_internal_error_exposure_allowed",
    ):
        require(fallback_policy.get(field) is False, f"{field} must remain false")

    side_effect_policy = fixture.get("no_side_effect_policy") or {}
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(side_effect_policy.get("side_effect_counters_must_remain") or [])),
        "side-effect counters drifted",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effect_policy.get("forbidden_side_effects") or [])),
        "forbidden side effects drifted",
    )


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-control-plane-read-schema-migration-implementation-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run schema migration implementation preconditions check")
    require(
        check_repo.index("check-control-plane-read-store-selector-enablement-preconditions-v1.py")
        < check_repo.index(current_checker),
        "schema migration implementation preconditions must run after selector enablement preconditions",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_schema_migration_implementation_leaked(fixture: dict[str, Any]) -> None:
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
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this preconditions slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_implementation_boundary(fixture)
    assert_future_artifact_manifest(fixture)
    assert_gate_matrix(fixture)
    assert_route_matrix(fixture)
    assert_failure_and_safety_policies(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_schema_migration_implementation_leaked(fixture)
    print("control plane read schema migration implementation preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
