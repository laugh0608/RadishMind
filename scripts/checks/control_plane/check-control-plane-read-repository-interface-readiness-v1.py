#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-interface-readiness-v1.json"
RUNNER_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-implementation-v1.json"
)
TYPES_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-types-implementation-v1.json"
)
REPOSITORY_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-v1.json"
)
REPOSITORY_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-preconditions-v1.json"
)
IMPLEMENTATION_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-implementation-readiness-v1.json"
)
STORE_SELECTION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selection-readiness-v1.json"
)
SCHEMA_MIGRATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-migration-readiness-v1.json"
)
DISABLED_DATABASE_GUARD_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-disabled-database-guard-v1.json"
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
    "control-plane-read-repository-contract-smoke-runner-implementation-v1": (
        RUNNER_IMPLEMENTATION_FIXTURE_PATH,
        "repository_contract_smoke_runner_implemented",
    ),
    "control-plane-read-repository-contract-types-implementation-v1": (
        TYPES_IMPLEMENTATION_FIXTURE_PATH,
        "repository_contract_types_implemented",
    ),
    "control-plane-read-repository-contract-smoke-v1": (
        REPOSITORY_SMOKE_FIXTURE_PATH,
        "repository_contract_smoke_defined",
    ),
    "control-plane-read-repository-contract-preconditions-v1": (
        REPOSITORY_PRECONDITIONS_FIXTURE_PATH,
        "repository_contract_preconditions_defined",
    ),
    "control-plane-read-repository-implementation-readiness-v1": (
        IMPLEMENTATION_READINESS_FIXTURE_PATH,
        "repository_implementation_readiness_defined",
    ),
    "control-plane-read-store-selection-readiness-v1": (
        STORE_SELECTION_FIXTURE_PATH,
        "store_selection_readiness_defined",
    ),
    "control-plane-read-schema-migration-readiness-v1": (
        SCHEMA_MIGRATION_FIXTURE_PATH,
        "schema_migration_readiness_defined",
    ),
    "control-plane-read-disabled-database-guard-v1": (
        DISABLED_DATABASE_GUARD_FIXTURE_PATH,
        "disabled_database_guard_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "repository_interface_ready",
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
EXPECTED_GATES = {
    "contract_types_implemented",
    "static_runner_implemented",
    "interface_method_matrix_defined",
    "adapter_implementation_gate",
    "production_auth_gate",
    "no_interface_or_adapter_leaked",
}
EXPECTED_FAILURE_CODES = {
    "tenant_binding_missing",
    "scope_denied",
    "invalid_filter",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "database_read_disabled",
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
        fixture.get("kind") == "control_plane_read_repository_interface_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-repository-interface-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "repository_interface_readiness_defined",
        "repository interface readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_interface_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("interface_boundary") or {}
    require(
        boundary.get("status") == "interface_readiness_defined_not_declared",
        "interface boundary status drifted",
    )
    require(boundary.get("future_interface_name") == "ControlPlaneReadRepository", "interface name drifted")
    future_interface_file = REPO_ROOT / str(boundary.get("future_interface_file") or "")
    require(not future_interface_file.exists(), "future interface file must not be created in readiness slice")
    require(
        (REPO_ROOT / str(boundary.get("current_type_file") or "")).exists(),
        "current type file must exist",
    )
    require(
        (REPO_ROOT / str(boundary.get("static_runner_file") or "")).exists(),
        "static runner file must exist",
    )
    for field in (
        "interface_declared_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_interface_method_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("interface_method_policy") or {}
    preconditions = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH).get("interface_contract") or {}
    require(policy.get("status") == "method_matrix_defined_not_declared", "method policy status drifted")
    require(policy.get("context_argument") == preconditions.get("context_argument"), "context argument drifted")
    require(policy.get("context_must_be_first_argument") is True, "context must remain first argument")
    for field in (
        "tenant_ref_from_request_allowed",
        "raw_http_request_allowed",
        "raw_auth_material_allowed",
    ):
        require(policy.get(field) is False, f"{field} must remain false")
    require(
        policy.get("method_result_policy") == "route_specific_read_repository_result_alias",
        "result policy drifted",
    )
    require(
        policy.get("failure_policy") == "typed_failure_code_without_database_internal_detail",
        "failure policy drifted",
    )


def assert_method_matrix(fixture: dict[str, Any]) -> None:
    methods = rows_by_route(fixture, "method_matrix")
    type_rows = rows_by_route(load_json(TYPES_IMPLEMENTATION_FIXTURE_PATH), "route_type_matrix")
    runner_rows = rows_by_route(load_json(RUNNER_IMPLEMENTATION_FIXTURE_PATH), "route_runner_matrix")
    precondition_rows = rows_by_route(load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH), "repository_operations")
    for route_id, row in methods.items():
        type_row = type_rows[route_id]
        runner_row = runner_rows[route_id]
        precondition_row = precondition_rows[route_id]
        require(row.get("operation") == type_row.get("operation"), f"{route_id} operation drifted")
        require(row.get("operation") == runner_row.get("operation"), f"{route_id} runner operation drifted")
        require(row.get("operation") == precondition_row.get("operation"), f"{route_id} precondition operation drifted")
        require(row.get("request_type") == type_row.get("request_type"), f"{route_id} request type drifted")
        require(row.get("result_type") == type_row.get("result_type"), f"{route_id} result type drifted")
        require(row.get("summary_type"), f"{route_id} summary type must be defined")
        require(row.get("source_type_contract") is True, f"{route_id} must source type contract")
        require(row.get("covered_by_static_runner") is True, f"{route_id} must be covered by static runner")
        require(row.get("future_adapter_allowed_now") is False, f"{route_id} adapter must remain blocked")


def assert_readiness_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("readiness_gates") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATES, "readiness gate ids drifted")
    require(gates["contract_types_implemented"].get("status") == "satisfied", "contract types gate drifted")
    require(gates["static_runner_implemented"].get("status") == "satisfied", "static runner gate drifted")
    require(gates["interface_method_matrix_defined"].get("status") == "required_now", "method matrix gate drifted")
    require(gates["no_interface_or_adapter_leaked"].get("status") == "required_now", "leak gate drifted")
    require(gates["adapter_implementation_gate"].get("status") == "not_satisfied", "adapter gate must remain blocked")
    require(gates["production_auth_gate"].get("status") == "not_satisfied", "auth gate must remain blocked")
    for gate in gates.values():
        require(gate.get("evidence") or gate.get("must_cover"), f"{gate.get('id')} must define evidence or coverage")


def assert_failure_and_side_effect_policies(fixture: dict[str, Any]) -> None:
    failures = set(fixture.get("failure_mapping") or [])
    require(EXPECTED_FAILURE_CODES.issubset(failures), "missing interface failure mappings")
    runner_failures = set(load_json(RUNNER_IMPLEMENTATION_FIXTURE_PATH).get("failure_mapping") or [])
    require(runner_failures.issubset(failures), "interface failures must include runner failures")

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
    current_checker = "check-control-plane-read-repository-interface-readiness-v1.py"
    runner_checker = "check-control-plane-read-repository-contract-smoke-runner-implementation-v1.py"
    require(current_checker in check_repo, "check-repo.py must run repository interface readiness check")
    require(
        check_repo.index(runner_checker) < check_repo.index(current_checker),
        "interface readiness check must run after smoke runner implementation",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_interface_or_adapter_leaked(fixture: dict[str, Any]) -> None:
    configured = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_SOURCE_ABSENT_LITERALS.issubset(configured), "source absent literals drifted")
    for root in (REPO_ROOT / "services/platform/internal", REPO_ROOT / "apps/radishmind-web/src"):
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured:
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this readiness slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_interface_boundary(fixture)
    assert_interface_method_policy(fixture)
    assert_method_matrix(fixture)
    assert_readiness_gates(fixture)
    assert_failure_and_side_effect_policies(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_interface_or_adapter_leaked(fixture)
    print("control plane read repository interface readiness v1 checks passed.")


if __name__ == "__main__":
    main()
