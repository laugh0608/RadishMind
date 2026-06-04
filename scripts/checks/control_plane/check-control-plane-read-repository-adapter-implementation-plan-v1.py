#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/control-plane-read-repository-adapter-implementation-plan-v1.json"
)
ADAPTER_REFRESH_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/control-plane-read-repository-adapter-implementation-readiness-refresh-v1.json"
)
SCHEMA_MIGRATION_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/control-plane-read-schema-migration-implementation-preconditions-v1.json"
)
SELECTOR_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selector-enablement-preconditions-v1.json"
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
STORE_SELECTION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selection-readiness-v1.json"
)
SCHEMA_MIGRATION_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-migration-readiness-v1.json"
)
DISABLED_DATABASE_GUARD_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-disabled-database-guard-v1.json"
)
REPOSITORY_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-v1.json"
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
    "control-plane-read-repository-adapter-implementation-readiness-refresh-v1": (
        ADAPTER_REFRESH_FIXTURE_PATH,
        "repository_adapter_implementation_readiness_refreshed",
    ),
    "control-plane-read-schema-migration-implementation-preconditions-v1": (
        SCHEMA_MIGRATION_IMPLEMENTATION_FIXTURE_PATH,
        "schema_migration_implementation_preconditions_defined",
    ),
    "control-plane-read-store-selector-enablement-preconditions-v1": (
        SELECTOR_PRECONDITIONS_FIXTURE_PATH,
        "store_selector_enablement_preconditions_defined",
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
    "control-plane-read-store-selection-readiness-v1": (
        STORE_SELECTION_FIXTURE_PATH,
        "store_selection_readiness_defined",
    ),
    "control-plane-read-schema-migration-readiness-v1": (
        SCHEMA_MIGRATION_READINESS_FIXTURE_PATH,
        "schema_migration_readiness_defined",
    ),
    "control-plane-read-disabled-database-guard-v1": (
        DISABLED_DATABASE_GUARD_FIXTURE_PATH,
        "disabled_database_guard_defined",
    ),
    "control-plane-read-repository-contract-smoke-v1": (
        REPOSITORY_SMOKE_FIXTURE_PATH,
        "repository_contract_smoke_defined",
    ),
    "control-plane-read-auth-store-transition-preconditions-v1": (
        AUTH_STORE_FIXTURE_PATH,
        "auth_store_transition_preconditions_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "repository_interface_ready",
    "repository_adapter_ready",
    "repository_implementation_ready",
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "migration_files_created",
    "migration_runner_ready",
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
EXPECTED_DEPENDENCY_CONSUMPTION_IDS = {
    "repository_adapter_readiness_refresh_consumed",
    "schema_migration_implementation_preconditions_consumed",
    "selector_enablement_preconditions_consumed",
    "interface_method_matrix_consumed",
    "contract_type_matrix_consumed",
    "static_runner_consumed",
}
EXPECTED_GATE_IDS = {
    "interface_contract_type_static_runner_consumed",
    "schema_migration_implementation_preconditions_consumed",
    "selector_gate",
    "schema_migration_artifact_gate",
    "production_auth_gate",
    "durable_adapter_smoke_gate",
    "no_adapter_selector_sql_or_migration_leaked",
}
EXPECTED_FAILURE_CODES = {
    "tenant_binding_missing",
    "scope_denied",
    "invalid_filter",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "database_read_disabled",
    "invalid_read_store_mode",
    "auth_context_contract_mismatch",
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
    "type ControlPlaneReadRepository interface",
    "func NewControlPlaneReadRepository",
    "control_plane_read_repository_interface.go",
    "control_plane_read_repository_adapter.go",
    "control_plane_read_repository_adapter_test.go",
    "control_plane_read_repository_adapter_contract_smoke_test.go",
    "control_plane_read_store_selector.go",
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
        fixture.get("kind") == "control_plane_read_repository_adapter_implementation_plan_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-repository-adapter-implementation-plan-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "repository_adapter_implementation_plan_defined",
        "repository adapter implementation plan status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_plan_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_plan_boundary") or {}
    require(
        boundary.get("status") == "repository_adapter_implementation_plan_defined_not_implemented",
        "plan boundary status drifted",
    )
    require(boundary.get("future_interface_name") == "ControlPlaneReadRepository", "interface name drifted")
    for field in (
        "future_interface_file",
        "future_adapter_file",
        "future_adapter_test_file",
        "future_adapter_contract_smoke_test_file",
        "future_selector_file",
        "future_migration_root",
    ):
        future_path = REPO_ROOT / str(boundary.get(field) or "")
        require(not future_path.exists(), f"{field} must not be created in this plan slice")
    for field in ("current_type_file", "static_runner_file"):
        require((REPO_ROOT / str(boundary.get(field) or "")).exists(), f"{field} must exist")
    for field in (
        "interface_file_created_in_this_slice",
        "adapter_file_created_in_this_slice",
        "adapter_test_created_in_this_slice",
        "adapter_contract_smoke_test_created_in_this_slice",
        "selector_file_created_in_this_slice",
        "migration_root_created_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_dependency_consumption(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("id") or ""): row
        for row in fixture.get("dependency_consumption") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_DEPENDENCY_CONSUMPTION_IDS, "dependency consumption ids drifted")
    for gate_id, row in rows.items():
        require(row.get("status") == "satisfied", f"{gate_id} must be satisfied")
        evidence = row.get("evidence")
        require(evidence in EXPECTED_DEPENDENCIES, f"{gate_id} cites unknown evidence")


def assert_future_file_layout(fixture: dict[str, Any]) -> None:
    rows = fixture.get("future_file_layout") or []
    require(len(rows) >= 6, "future file layout must include interface, adapter, tests, selector and migration root")
    for row in rows:
        require(isinstance(row, dict), "future file layout rows must be objects")
        relative_path = str(row.get("path") or "")
        require(relative_path, "future file layout row missing path")
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(row.get("allowed_before_gates") is False, f"{relative_path} must remain blocked before gates")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this plan slice")


def assert_adapter_gate_matrix(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("adapter_implementation_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "adapter implementation gate ids drifted")
    for gate_id in (
        "interface_contract_type_static_runner_consumed",
        "schema_migration_implementation_preconditions_consumed",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must be satisfied")
        require(gates[gate_id].get("evidence"), f"{gate_id} must cite evidence")
    for gate_id in (
        "selector_gate",
        "schema_migration_artifact_gate",
        "production_auth_gate",
        "durable_adapter_smoke_gate",
    ):
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must define coverage")
    require(
        gates["no_adapter_selector_sql_or_migration_leaked"].get("status") == "required_now",
        "leak gate must be required now",
    )


def assert_adapter_route_matrix(fixture: dict[str, Any]) -> None:
    adapter_rows = rows_by_route(fixture, "adapter_route_matrix")
    interface_rows = rows_by_route(load_json(INTERFACE_READINESS_FIXTURE_PATH), "method_matrix")
    type_rows = rows_by_route(load_json(TYPES_IMPLEMENTATION_FIXTURE_PATH), "route_type_matrix")
    runner_rows = rows_by_route(load_json(RUNNER_IMPLEMENTATION_FIXTURE_PATH), "route_runner_matrix")
    migration_rows = rows_by_route(
        load_json(SCHEMA_MIGRATION_IMPLEMENTATION_FIXTURE_PATH),
        "route_migration_implementation_matrix",
    )
    selector_rows = rows_by_route(load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH), "route_selector_enablement_matrix")
    for route_id, row in adapter_rows.items():
        interface_row = interface_rows[route_id]
        type_row = type_rows[route_id]
        runner_row = runner_rows[route_id]
        migration_row = migration_rows[route_id]
        selector_row = selector_rows[route_id]
        require(row.get("operation") == interface_row.get("operation"), f"{route_id} operation drifted")
        require(row.get("operation") == type_row.get("operation"), f"{route_id} type operation drifted")
        require(row.get("operation") == runner_row.get("operation"), f"{route_id} runner operation drifted")
        require(row.get("operation") == migration_row.get("operation"), f"{route_id} migration operation drifted")
        require(row.get("operation") == selector_row.get("operation"), f"{route_id} selector operation drifted")
        require(row.get("request_type") == interface_row.get("request_type"), f"{route_id} request type drifted")
        require(row.get("request_type") == type_row.get("request_type"), f"{route_id} type request drifted")
        require(row.get("result_type") == interface_row.get("result_type"), f"{route_id} result type drifted")
        require(row.get("result_type") == type_row.get("result_type"), f"{route_id} type result drifted")
        for field in (
            "interface_method_consumed",
            "contract_type_consumed",
            "static_runner_consumed",
            "schema_migration_preconditions_consumed",
            "selector_gate_required",
        ):
            require(row.get(field) is True, f"{route_id} {field} must remain true")
        for field in ("adapter_implementation_allowed_now", "fake_fallback_allowed", "side_effect_allowed"):
            require(row.get(field) is False, f"{route_id} {field} must remain false")
        require(len(row.get("required_future_adapter_checks") or []) >= 4, f"{route_id} adapter checks too thin")
        route_failures = set(row.get("failure_mapping") or [])
        require("schema_migration_not_applied" in route_failures, f"{route_id} must map missing migration")
        require("schema_version_mismatch" in route_failures, f"{route_id} must map schema mismatch")


def assert_failure_and_safety_policies(fixture: dict[str, Any]) -> None:
    failures = set(fixture.get("failure_mapping") or [])
    require(EXPECTED_FAILURE_CODES.issubset(failures), "failure mapping missing required codes")
    interface_failures = set(load_json(INTERFACE_READINESS_FIXTURE_PATH).get("failure_mapping") or [])
    require(interface_failures.issubset(failures), "adapter plan must include interface failures")
    selector_failures = set(load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH).get("failure_mapping") or [])
    require(selector_failures.issubset(failures), "adapter plan must include selector failures")
    migration_failures = set(load_json(SCHEMA_MIGRATION_IMPLEMENTATION_FIXTURE_PATH).get("failure_mapping") or [])
    require(migration_failures.issubset(failures), "adapter plan must include migration implementation failures")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "repository_mode_fallback_to_fixture_fake_store_allowed",
        "database_mode_fallback_to_fixture_fake_store_allowed",
        "schema_missing_fallback_to_fixture_fake_store_allowed",
        "selector_reserved_mode_success_allowed",
        "fallback_to_test_auth_allowed",
        "missing_tenant_success_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
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
    current_checker = "check-control-plane-read-repository-adapter-implementation-plan-v1.py"
    schema_checker = "check-control-plane-read-schema-migration-implementation-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run repository adapter implementation plan check")
    require(
        check_repo.index(schema_checker) < check_repo.index(current_checker),
        "repository adapter implementation plan check must run after schema migration implementation preconditions",
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
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this plan slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_plan_boundary(fixture)
    assert_dependency_consumption(fixture)
    assert_future_file_layout(fixture)
    assert_adapter_gate_matrix(fixture)
    assert_adapter_route_matrix(fixture)
    assert_failure_and_safety_policies(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_implementation_leaked(fixture)
    print("control plane read repository adapter implementation plan v1 checks passed.")


if __name__ == "__main__":
    main()
