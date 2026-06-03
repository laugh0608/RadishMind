#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-types-readiness-v1.json"
)
SCHEMA_MIGRATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-migration-readiness-v1.json"
)
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
IMPLEMENTATION_FIXTURE_PATH = (
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
EXPECTED_DEPENDENCIES = {
    "control-plane-read-schema-migration-readiness-v1": (
        SCHEMA_MIGRATION_FIXTURE_PATH,
        "schema_migration_readiness_defined",
    ),
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
    "repository_contract_types_implemented",
    "repository_interface_ready",
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
    "context_type_fields_defined",
    "route_request_result_types_defined",
    "projection_filter_sort_types_defined",
    "failure_type_mapping_defined",
    "contract_smoke_type_inputs_defined",
    "no_implementation_leaked",
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
EXPECTED_SOURCE_ABSENT_LITERALS = {
    "type ControlPlaneReadRepository interface",
    "func NewControlPlaneReadRepository",
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


def repository_operations_by_route_id() -> dict[str, dict[str, Any]]:
    repository = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH)
    operations = {
        str(operation.get("route_id") or ""): operation
        for operation in repository.get("repository_operations") or []
        if isinstance(operation, dict)
    }
    require(set(operations) == EXPECTED_ROUTE_IDS, "repository operation ids drifted")
    return operations


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
        fixture.get("kind") == "control_plane_read_repository_contract_types_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-repository-contract-types-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "repository_contract_types_readiness_defined",
        "repository contract types readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_contract_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("contract_types_boundary") or {}
    store_selection = load_json(STORE_SELECTION_FIXTURE_PATH).get("selection_boundary") or {}
    implementation = load_json(IMPLEMENTATION_READINESS_FIXTURE_PATH)
    implementation_boundary = implementation.get("readiness_boundary") or {}
    planned_files = {
        str(item.get("path") or ""): item
        for item in (implementation.get("future_file_layout") or {}).get("planned_files") or []
        if isinstance(item, dict)
    }
    future_contract_file = str(boundary.get("future_contract_file") or "")
    require(
        boundary.get("status") == "contract_types_readiness_defined_not_implemented",
        "contract type boundary status drifted",
    )
    require(
        boundary.get("current_store_source") == store_selection.get("current_default_source"),
        "current store source must match store selection default",
    )
    require(
        boundary.get("future_store_source") == store_selection.get("future_durable_source"),
        "future store source must match store selection",
    )
    require(
        boundary.get("future_store_source") == implementation_boundary.get("future_store_source"),
        "future store source must match implementation readiness",
    )
    require(future_contract_file in planned_files, "future contract file must match implementation readiness")
    if (REPO_ROOT / future_contract_file).exists():
        implementation = load_json(IMPLEMENTATION_FIXTURE_PATH)
        implementation_slice = implementation.get("slice") or {}
        require(
            implementation_slice.get("status") == "repository_contract_types_implemented",
            "future Go contract file must be covered by implementation gate",
        )
    for field in (
        "type_files_created_in_this_slice",
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


def assert_context_type_contract(fixture: dict[str, Any]) -> None:
    context = fixture.get("context_type_contract") or {}
    repository = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH).get("interface_contract") or {}
    required_names = {
        str(item.get("name") or "")
        for item in context.get("required_fields") or []
        if isinstance(item, dict)
    }
    expected_names = set(repository.get("required_context_fields") or [])
    require(context.get("type_name") == repository.get("context_argument"), "context type name drifted")
    require(required_names == expected_names, "context required fields must match repository preconditions")
    for item in context.get("required_fields") or []:
        require(isinstance(item, dict), "context fields must be objects")
        require(item.get("required") is True, f"{item.get('name')} must remain required")
        require(item.get("source"), f"{item.get('name')} must define source")
    forbidden = set(context.get("must_not_include") or [])
    require("raw_authorization_header" in forbidden, "context must forbid raw authorization header")
    require("workflow_executor" in forbidden, "context must forbid workflow executor dependency")
    require("business_writeback_payload" in forbidden, "context must forbid business writeback payload")


def assert_request_and_result_policies(fixture: dict[str, Any]) -> None:
    repository = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH).get("interface_contract") or {}
    request_policy = fixture.get("request_type_policy") or {}
    result_policy = fixture.get("result_type_policy") or {}
    require(set(request_policy.get("shared_fields") or []) == set(repository.get("required_request_fields") or []), "request shared fields drifted")
    require(set(result_policy.get("required_fields") or []) == set(repository.get("required_result_fields") or []), "result fields drifted")
    require(request_policy.get("tenant_ref_field_allowed") is False, "tenant_ref must come from context")
    require(
        request_policy.get("tenant_ref_source") == "ReadRepositoryContext.tenant_ref",
        "tenant ref source drifted",
    )
    require(request_policy.get("unknown_filter_behavior") == "fail_closed_invalid_filter", "unknown filter policy drifted")
    require(request_policy.get("unknown_sort_behavior") == "fail_closed_invalid_filter", "unknown sort policy drifted")
    require(request_policy.get("sensitive_projection_query_allowed") is False, "sensitive projection must be denied")
    for field in (
        "raw_database_error_allowed",
        "secret_material_allowed",
        "materialized_tool_result_allowed",
        "business_writeback_payload_allowed",
    ):
        require(result_policy.get(field) is False, f"{field} must remain false")


def assert_route_type_matrix(fixture: dict[str, Any]) -> None:
    operations = repository_operations_by_route_id()
    matrix = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_type_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "route type matrix must cover every read route")
    for route_id, row in matrix.items():
        operation = operations[route_id]
        require(row.get("operation") == operation.get("operation"), f"{route_id} operation drifted")
        require(row.get("mode") == operation.get("mode"), f"{route_id} mode drifted")
        require(row.get("request_type"), f"{route_id} must define request_type")
        require(row.get("result_type"), f"{route_id} must define result_type")
        require(row.get("summary_type"), f"{route_id} must define summary_type")
        require(row.get("projection_type"), f"{route_id} must define projection_type")
        require(set(row.get("allowed_filters") or []) == set(operation.get("allowed_filters") or []), f"{route_id} filters drifted")
        require(set(row.get("allowed_sort") or []) == set(operation.get("allowed_sort") or []), f"{route_id} sort drifted")
        if row.get("mode") == "single_resource_read":
            require(str(row.get("request_type")).startswith("Read"), f"{route_id} single read request type drifted")
        if row.get("mode") == "cursor_list_read":
            require(str(row.get("request_type")).startswith("List"), f"{route_id} list request type drifted")


def assert_failure_and_smoke_types(fixture: dict[str, Any]) -> None:
    failure = fixture.get("failure_type_contract") or {}
    implementation = load_json(IMPLEMENTATION_READINESS_FIXTURE_PATH).get("failure_mapping_policy") or {}
    schema = load_json(SCHEMA_MIGRATION_FIXTURE_PATH).get("failure_mapping_policy") or {}
    required_codes = set(failure.get("required_failure_codes") or [])
    require(EXPECTED_FAILURE_CODES.issubset(required_codes), "failure type contract missing required codes")
    require(
        set(implementation.get("required_failure_codes") or []).issubset(required_codes),
        "failure type contract must include implementation readiness codes",
    )
    require(
        set(schema.get("required_failure_codes") or []).issubset(required_codes),
        "failure type contract must include schema migration codes",
    )
    require(failure.get("type_name") == "ReadRepositoryFailureCode", "failure type name drifted")
    require(failure.get("database_internal_error_exposure_allowed") is False, "database internals must not be exposed")
    require(failure.get("unknown_failure_success_allowed") is False, "unknown failure must not become success")

    smoke = fixture.get("contract_smoke_type_inputs") or {}
    require(smoke.get("status") == "planned_not_implemented", "contract smoke types must stay planned")
    require(smoke.get("smoke_context_type") == "ReadRepositoryContext", "smoke context type drifted")
    require(smoke.get("smoke_case_type") == "ReadRepositoryContractSmokeCase", "smoke case type drifted")
    require(smoke.get("smoke_result_type") == "ReadRepositoryContractSmokeResult", "smoke result type drifted")
    require("all seven route request types" in set(smoke.get("must_cover") or []), "smoke type coverage drifted")


def assert_readiness_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("readiness_gates") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "contract type readiness gate ids drifted")
    for gate_id, gate in gates.items():
        if gate_id == "no_implementation_leaked":
            require(
                gate.get("status") in {
                    "required_now",
                    "satisfied_by_controlled_implementation_gate",
                },
                "no implementation gate status drifted",
            )
        else:
            require(
                gate.get("status") == "required_before_contract_type_implementation",
                f"{gate_id} status drifted",
            )
        require(gate.get("must_cover"), f"{gate_id} must list coverage requirements")


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-control-plane-read-repository-contract-types-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run repository contract types readiness check")
    require(
        check_repo.index("check-control-plane-read-schema-migration-readiness-v1.py")
        < check_repo.index(current_checker),
        "repository contract types readiness check must run after schema migration readiness",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_type_implementation_leaked(fixture: dict[str, Any]) -> None:
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
    assert_contract_boundary(fixture)
    assert_context_type_contract(fixture)
    assert_request_and_result_policies(fixture)
    assert_route_type_matrix(fixture)
    assert_failure_and_smoke_types(fixture)
    assert_readiness_gates(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_type_implementation_leaked(fixture)
    print("control plane read repository contract types readiness v1 checks passed.")


if __name__ == "__main__":
    main()
