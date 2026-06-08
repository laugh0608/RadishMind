#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-types-implementation-v1.json"
)
READINESS_FIXTURE_PATH = (
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
    "control-plane-read-repository-contract-types-readiness-v1": (
        READINESS_FIXTURE_PATH,
        "repository_contract_types_readiness_defined",
    ),
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
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "migration_files_created",
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
EXPECTED_FORBIDDEN_LITERALS = {
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
        fixture.get("kind") == "control_plane_read_repository_contract_types_implementation_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-repository-contract-types-implementation-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "repository_contract_types_implemented",
        "repository contract types implementation status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_implementation_boundary(fixture: dict[str, Any]) -> tuple[Path, Path]:
    boundary = fixture.get("implementation_boundary") or {}
    require(
        boundary.get("status") == "contract_types_implemented_no_repository_adapter",
        "implementation boundary status drifted",
    )
    type_file = REPO_ROOT / str(boundary.get("type_file") or "")
    test_file = REPO_ROOT / str(boundary.get("test_file") or "")
    require(type_file.exists(), f"missing type file: {type_file.relative_to(REPO_ROOT)}")
    require(test_file.exists(), f"missing test file: {test_file.relative_to(REPO_ROOT)}")
    for field in (
        "repository_interface_declared",
        "repository_adapter_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")
    return type_file, test_file


def assert_type_file(fixture: dict[str, Any], type_file: Path, test_file: Path) -> None:
    type_text = type_file.read_text(encoding="utf-8")
    test_text = test_file.read_text(encoding="utf-8")
    require("package httpapi" in type_text, "type file must stay in httpapi package")
    for literal in fixture.get("implemented_type_literals") or []:
        require(str(literal) in type_text, f"type file missing literal: {literal}")
    for field in fixture.get("required_context_fields") or []:
        require(str(field) in type_text, f"ReadRepositoryContext missing field: {field}")
    for code in fixture.get("required_failure_codes") or []:
        require(str(code) in type_text, f"type file missing failure code: {code}")
    require("TestControlPlaneReadRepositoryContextTypeBoundary" in test_text, "test must cover context type boundary")
    require("TestControlPlaneReadRepositoryRouteTypeContracts" in test_text, "test must cover route type contracts")
    require("TestControlPlaneReadRepositoryFailureCodes" in test_text, "test must cover failure codes")


def assert_route_type_matrix(fixture: dict[str, Any]) -> None:
    readiness = load_json(READINESS_FIXTURE_PATH)
    readiness_matrix = {
        str(row.get("route_id") or ""): row
        for row in readiness.get("route_type_matrix") or []
        if isinstance(row, dict)
    }
    matrix = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_type_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "implementation matrix must cover every read route")
    for route_id, row in matrix.items():
        expected = readiness_matrix.get(route_id) or {}
        require(row.get("operation") == expected.get("operation"), f"{route_id} operation drifted")
        require(row.get("request_type") == expected.get("request_type"), f"{route_id} request type drifted")
        require(row.get("result_type") == expected.get("result_type"), f"{route_id} result type drifted")


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-control-plane-read-repository-contract-types-implementation-v1.py"
    readiness_checker = "check-control-plane-read-repository-contract-types-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run repository contract types implementation check")
    require(
        check_repo.index(readiness_checker) < check_repo.index(current_checker),
        "implementation check must run after readiness check",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_forbidden_source(fixture: dict[str, Any]) -> None:
    configured_literals = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_FORBIDDEN_LITERALS.issubset(configured_literals), "forbidden literals drifted")
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
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this implementation slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    type_file, test_file = assert_implementation_boundary(fixture)
    assert_type_file(fixture, type_file, test_file)
    assert_route_type_matrix(fixture)
    require(EXPECTED_FAILURE_CODES.issubset(set(fixture.get("required_failure_codes") or [])), "failure code fixture drifted")
    assert_docs_and_check_repo(fixture)
    assert_no_forbidden_source(fixture)
    print("control plane read repository contract types implementation v1 checks passed.")


if __name__ == "__main__":
    main()
