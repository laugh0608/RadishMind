#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-implementation-v1.json"
)
READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-readiness-v1.json"
)
TYPES_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-types-implementation-v1.json"
)
REPOSITORY_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-v1.json"
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
    "control-plane-read-repository-contract-smoke-runner-readiness-v1": (
        READINESS_FIXTURE_PATH,
        "repository_contract_smoke_runner_readiness_defined",
    ),
    "control-plane-read-repository-contract-types-implementation-v1": (
        TYPES_IMPLEMENTATION_FIXTURE_PATH,
        "repository_contract_types_implemented",
    ),
    "control-plane-read-repository-contract-smoke-v1": (
        REPOSITORY_SMOKE_FIXTURE_PATH,
        "repository_contract_smoke_defined",
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
EXPECTED_FORBIDDEN_SOURCE = {
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
        fixture.get("kind") == "control_plane_read_repository_contract_smoke_runner_implementation_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-repository-contract-smoke-runner-implementation-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "repository_contract_smoke_runner_implemented",
        "smoke runner implementation status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_implementation_boundary(fixture: dict[str, Any]) -> tuple[Path, Path]:
    boundary = fixture.get("implementation_boundary") or {}
    require(
        boundary.get("status") == "static_contract_type_runner_implemented_no_repository_adapter",
        "implementation boundary status drifted",
    )
    runner_file = REPO_ROOT / str(boundary.get("runner_file") or "")
    test_file = REPO_ROOT / str(boundary.get("test_file") or "")
    require(runner_file.exists(), f"missing runner file: {runner_file.relative_to(REPO_ROOT)}")
    require(test_file.exists(), f"missing test file: {test_file.relative_to(REPO_ROOT)}")
    require(
        boundary.get("runner_name") == "ControlPlaneReadRepositoryContractSmokeRunner",
        "runner name drifted",
    )
    require(
        boundary.get("type_catalog_source") == "controlPlaneReadRepositoryRouteTypeContracts",
        "type catalog source drifted",
    )
    require(
        boundary.get("contract_smoke_source") == "control-plane-read-repository-contract-smoke-v1.json",
        "contract smoke source drifted",
    )
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
    return runner_file, test_file


def assert_runner_source(fixture: dict[str, Any], runner_file: Path, test_file: Path) -> None:
    runner_text = runner_file.read_text(encoding="utf-8")
    test_text = test_file.read_text(encoding="utf-8")
    require("package httpapi" in runner_text, "runner must stay in httpapi package")
    for literal in fixture.get("implemented_runner_literals") or []:
        require(str(literal) in runner_text, f"runner missing literal: {literal}")
    for literal in fixture.get("runner_test_literals") or []:
        require(str(literal) in test_text, f"runner test missing literal: {literal}")
    require(
        "control-plane-read-repository-contract-smoke-v1.json" in test_text,
        "runner test must compare committed smoke fixture",
    )


def assert_route_runner_matrix(fixture: dict[str, Any]) -> None:
    readiness = load_json(READINESS_FIXTURE_PATH)
    types_fixture = load_json(TYPES_IMPLEMENTATION_FIXTURE_PATH)
    smoke_fixture = load_json(REPOSITORY_SMOKE_FIXTURE_PATH)
    readiness_matrix = {
        str(row.get("route_id") or ""): row
        for row in readiness.get("route_runner_matrix") or []
        if isinstance(row, dict)
    }
    type_matrix = {
        str(row.get("route_id") or ""): row
        for row in types_fixture.get("route_type_matrix") or []
        if isinstance(row, dict)
    }
    smoke_matrix = {
        str(row.get("route_id") or ""): row
        for row in smoke_fixture.get("route_smoke_matrix") or []
        if isinstance(row, dict)
    }
    matrix = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_runner_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "implementation matrix must cover every read route")
    for route_id, row in matrix.items():
        readiness_row = readiness_matrix.get(route_id) or {}
        type_row = type_matrix.get(route_id) or {}
        smoke_row = smoke_matrix.get(route_id) or {}
        require(row == readiness_row, f"{route_id} runner matrix must match readiness gate")
        require(row.get("operation") == type_row.get("operation"), f"{route_id} operation drifted")
        require(row.get("request_type") == type_row.get("request_type"), f"{route_id} request type drifted")
        require(row.get("result_type") == type_row.get("result_type"), f"{route_id} result type drifted")
        require(row.get("operation") == smoke_row.get("operation"), f"{route_id} smoke operation drifted")
        require(row.get("smoke_read_mode") == smoke_row.get("read_mode"), f"{route_id} smoke mode drifted")
        require(row.get("uses_type_contract") is True, f"{route_id} must consume type contract")
        require(row.get("uses_smoke_fixture") is True, f"{route_id} must consume smoke fixture")
        require(row.get("fake_fallback_allowed") is False, f"{route_id} fake fallback must remain false")
        require(row.get("side_effect_allowed") is False, f"{route_id} side effect must remain false")


def assert_failure_and_side_effect_policies(fixture: dict[str, Any]) -> None:
    failures = set(fixture.get("failure_mapping") or [])
    require(EXPECTED_FAILURE_CODES.issubset(failures), "missing failure mapping codes")
    smoke_failures = {
        str(item.get("failure_code") or "")
        for item in load_json(REPOSITORY_SMOKE_FIXTURE_PATH).get("failure_mapping") or []
        if isinstance(item, dict)
    }
    require(smoke_failures.issubset(failures), "implementation failures must include smoke failures")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "fallback_to_fixture_fake_store_allowed",
        "fallback_to_test_auth_allowed",
        "fallback_to_success_on_unknown_filter_allowed",
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
    current_checker = "check-control-plane-read-repository-contract-smoke-runner-implementation-v1.py"
    readiness_checker = "check-control-plane-read-repository-contract-smoke-runner-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run smoke runner implementation check")
    require(
        check_repo.index(readiness_checker) < check_repo.index(current_checker),
        "implementation check must run after readiness check",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_forbidden_source(fixture: dict[str, Any]) -> None:
    configured = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_FORBIDDEN_SOURCE.issubset(configured), "forbidden source literals drifted")
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
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this implementation slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    runner_file, test_file = assert_implementation_boundary(fixture)
    assert_runner_source(fixture, runner_file, test_file)
    assert_route_runner_matrix(fixture)
    assert_failure_and_side_effect_policies(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_forbidden_source(fixture)
    print("control plane read repository contract smoke runner implementation v1 checks passed.")


if __name__ == "__main__":
    main()
