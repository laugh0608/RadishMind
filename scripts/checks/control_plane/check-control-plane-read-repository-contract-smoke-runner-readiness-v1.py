#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-readiness-v1.json"
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
TYPE_FILE_PATH = REPO_ROOT / "services/platform/internal/httpapi/control_plane_read_repository_contract.go"
IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-implementation-v1.json"
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
    "smoke_runner_implemented",
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
EXPECTED_INPUT_FIELDS = {
    "contract_type_matrix",
    "repository_context",
    "route_requests",
    "failure_injections",
    "side_effect_probe",
}
EXPECTED_OUTPUT_FIELDS = {
    "route_results",
    "failure_results",
    "contract_mismatch_report",
    "side_effect_report",
    "summary",
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
    "type ControlPlaneReadRepositoryContractSmokeRunner",
    "func RunControlPlaneReadRepositoryContractSmoke",
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
IMPLEMENTATION_COVERED_LITERALS = {
    "type ControlPlaneReadRepositoryContractSmokeRunner",
    "func RunControlPlaneReadRepositoryContractSmoke",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def implementation_gate_covers_runner() -> bool:
    if not IMPLEMENTATION_FIXTURE_PATH.exists():
        return False
    implementation = load_json(IMPLEMENTATION_FIXTURE_PATH)
    slice_info = implementation.get("slice") or {}
    return (
        slice_info.get("id") == "control-plane-read-repository-contract-smoke-runner-implementation-v1"
        and slice_info.get("status") == "repository_contract_smoke_runner_implemented"
    )


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
        fixture.get("kind") == "control_plane_read_repository_contract_smoke_runner_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-repository-contract-smoke-runner-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "repository_contract_smoke_runner_readiness_defined",
        "smoke runner readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_runner_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("runner_boundary") or {}
    require(
        boundary.get("status") == "smoke_runner_readiness_defined_not_implemented",
        "runner boundary status drifted",
    )
    require(
        boundary.get("future_runner_name") == "ControlPlaneReadRepositoryContractSmokeRunner",
        "future runner name drifted",
    )
    require(
        boundary.get("type_catalog_source") == "controlPlaneReadRepositoryRouteTypeContracts",
        "type catalog source drifted",
    )
    require(
        boundary.get("contract_smoke_source") == "ControlPlaneReadRepositoryContractSmoke",
        "contract smoke source drifted",
    )
    future_runner = REPO_ROOT / str(boundary.get("future_runner_file") or "")
    require(
        future_runner.relative_to(REPO_ROOT).as_posix()
        == "services/platform/internal/httpapi/control_plane_read_repository_contract_smoke_runner.go",
        "future runner file path drifted",
    )
    if future_runner.exists():
        require(
            implementation_gate_covers_runner(),
            "smoke runner file must be covered by implementation gate",
        )
    for field in (
        "runner_file_allowed_in_this_slice",
        "repository_interface_allowed_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_runner_io_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("runner_io_contract") or {}
    require(
        contract.get("status") == "runner_input_output_readiness_defined",
        "runner IO contract status drifted",
    )
    require(set(contract.get("input_fields") or []) == EXPECTED_INPUT_FIELDS, "runner input fields drifted")
    require(set(contract.get("output_fields") or []) == EXPECTED_OUTPUT_FIELDS, "runner output fields drifted")
    require(contract.get("context_source") == "ReadRepositoryContext", "context source drifted")
    require("ReadRepositoryRequest" in str(contract.get("request_source") or ""), "request source drifted")
    require("ReadRepositoryResult" in str(contract.get("result_source") or ""), "result source drifted")
    require(
        contract.get("success_policy") == "compare sanitized summary envelope against route result type",
        "success policy drifted",
    )
    require(
        contract.get("failure_policy") == "compare expected failure code without exposing database internal detail",
        "failure policy drifted",
    )


def assert_route_runner_matrix(fixture: dict[str, Any]) -> None:
    types_fixture = load_json(TYPES_IMPLEMENTATION_FIXTURE_PATH)
    smoke_fixture = load_json(REPOSITORY_SMOKE_FIXTURE_PATH)
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
    runner_matrix = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_runner_matrix") or []
        if isinstance(row, dict)
    }
    require(set(type_matrix) == EXPECTED_ROUTE_IDS, "type matrix route ids drifted")
    require(set(smoke_matrix) == EXPECTED_ROUTE_IDS, "smoke matrix route ids drifted")
    require(set(runner_matrix) == EXPECTED_ROUTE_IDS, "runner matrix must cover every read route")
    for route_id, runner_row in runner_matrix.items():
        type_row = type_matrix[route_id]
        smoke_row = smoke_matrix[route_id]
        require(runner_row.get("operation") == type_row.get("operation"), f"{route_id} operation drifted")
        require(
            runner_row.get("operation") == smoke_row.get("operation"),
            f"{route_id} operation must match smoke contract",
        )
        require(
            runner_row.get("request_type") == type_row.get("request_type"),
            f"{route_id} request type drifted",
        )
        require(
            runner_row.get("result_type") == type_row.get("result_type"),
            f"{route_id} result type drifted",
        )
        require(
            runner_row.get("smoke_read_mode") == smoke_row.get("read_mode"),
            f"{route_id} smoke read mode drifted",
        )
        require(runner_row.get("uses_type_contract") is True, f"{route_id} must consume type contract")
        require(runner_row.get("uses_smoke_fixture") is True, f"{route_id} must consume smoke fixture")
        require(runner_row.get("fake_fallback_allowed") is False, f"{route_id} fake fallback must be forbidden")
        require(runner_row.get("side_effect_allowed") is False, f"{route_id} side effects must be forbidden")


def assert_failure_and_side_effect_policies(fixture: dict[str, Any]) -> None:
    failures = {
        str(item.get("failure_code") or "")
        for item in fixture.get("failure_mapping") or []
        if isinstance(item, dict)
    }
    missing_failures = sorted(EXPECTED_FAILURE_CODES - failures)
    require(not missing_failures, f"missing runner failure mappings: {missing_failures}")
    smoke_failures = {
        str(item.get("failure_code") or "")
        for item in load_json(REPOSITORY_SMOKE_FIXTURE_PATH).get("failure_mapping") or []
        if isinstance(item, dict)
    }
    require(smoke_failures.issubset(failures), "runner failures must include repository smoke failures")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    require(fallback.get("status") == "required_for_future_runner", "fallback policy status drifted")
    for field in (
        "fallback_to_fixture_fake_store_allowed",
        "fallback_to_test_auth_allowed",
        "fallback_to_success_on_unknown_filter_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
    require(side_effects.get("status") == "required_for_future_runner", "side effect policy status drifted")
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(side_effects.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effects.get("forbidden_side_effects") or [])),
        "missing forbidden side effects",
    )


def assert_type_file_and_docs(fixture: dict[str, Any]) -> None:
    type_text = TYPE_FILE_PATH.read_text(encoding="utf-8")
    require("controlPlaneReadRepositoryRouteTypeContracts" in type_text, "type catalog function missing")
    require("ReadRepositoryContext" in type_text, "ReadRepositoryContext missing")
    require("ReadRepositoryResult" in type_text, "ReadRepositoryResult missing")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-control-plane-read-repository-contract-smoke-runner-readiness-v1.py"
    type_checker = "check-control-plane-read-repository-contract-types-implementation-v1.py"
    smoke_checker = "check-control-plane-read-repository-contract-smoke-v1.py"
    require(current_checker in check_repo, "check-repo.py must run smoke runner readiness check")
    require(check_repo.index(smoke_checker) < check_repo.index(current_checker), "runner check must run after smoke")
    require(check_repo.index(type_checker) < check_repo.index(current_checker), "runner check must run after types")

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_runner_implementation_leaked(fixture: dict[str, Any]) -> None:
    configured_literals = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_SOURCE_ABSENT_LITERALS.issubset(configured_literals), "source absent literals drifted")
    if implementation_gate_covers_runner():
        configured_literals = configured_literals - IMPLEMENTATION_COVERED_LITERALS
    for root in (REPO_ROOT / "services/platform/internal", REPO_ROOT / "apps/radishmind-web/src"):
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured_literals:
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in readiness slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_runner_boundary(fixture)
    assert_runner_io_contract(fixture)
    assert_route_runner_matrix(fixture)
    assert_failure_and_side_effect_policies(fixture)
    assert_type_file_and_docs(fixture)
    assert_no_runner_implementation_leaked(fixture)
    print("control plane read repository contract smoke runner readiness v1 checks passed.")


if __name__ == "__main__":
    main()
