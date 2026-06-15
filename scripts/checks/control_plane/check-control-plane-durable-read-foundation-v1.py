#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-durable-read-foundation-v1.json"
FAKE_STORE_HANDLER_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json"
)
TYPES_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-types-implementation-v1.json"
)
RUNNER_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-implementation-v1.json"
)
INTERFACE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-interface-readiness-v1.json"
)
ENTRY_REVIEW_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-entry-review-v1.json"
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
    "control-plane-read-fake-store-handler-implementation-v1": (
        FAKE_STORE_HANDLER_FIXTURE_PATH,
        "fake_store_handler_implemented",
    ),
    "control-plane-read-repository-contract-types-implementation-v1": (
        TYPES_IMPLEMENTATION_FIXTURE_PATH,
        "repository_contract_types_implemented",
    ),
    "control-plane-read-repository-contract-smoke-runner-implementation-v1": (
        RUNNER_IMPLEMENTATION_FIXTURE_PATH,
        "repository_contract_smoke_runner_implemented",
    ),
    "control-plane-read-repository-interface-readiness-v1": (
        INTERFACE_READINESS_FIXTURE_PATH,
        "repository_interface_readiness_defined",
    ),
    "control-plane-read-implementation-entry-review-v1": (
        ENTRY_REVIEW_FIXTURE_PATH,
        "implementation_entry_review_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "migration_files_created",
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
EXPECTED_FORBIDDEN_SOURCE_LITERALS = {
    "func NewControlPlaneReadRepository",
    "control_plane_read_repository_adapter.go",
    "control_plane_read_store_selector.go",
    "control_plane_read_auth_middleware.go",
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
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "fake_store_write_count=0",
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
    require(fixture.get("kind") == "control_plane_durable_read_foundation_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-durable-read-foundation-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(slice_info.get("status") == "durable_read_foundation_implemented", "implementation status drifted")
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_implementation_boundary(fixture: dict[str, Any]) -> dict[str, Path]:
    boundary = fixture.get("implementation_boundary") or {}
    require(
        boundary.get("status") == "repository_interface_and_fake_store_bridge_implemented",
        "implementation boundary status drifted",
    )
    require(boundary.get("interface_name") == "ControlPlaneReadRepository", "interface name drifted")
    require(boundary.get("package_interface_name") == "controlPlaneReadRepository", "package interface name drifted")
    require(boundary.get("repository_constructor") == "newControlPlaneReadRepository", "constructor name drifted")
    require(boundary.get("current_data_source") == "fixtureBackedControlPlaneReadStore", "data source drifted")
    for field in (
        "response_envelope_shape_changed",
        "repository_adapter_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")

    paths = {}
    for key in ("repository_file", "handler_file", "contract_type_file", "fake_store_file", "server_file", "test_file"):
        path = REPO_ROOT / str(boundary.get(key) or "")
        require(path.exists(), f"missing {key}: {path.relative_to(REPO_ROOT)}")
        paths[key] = path
    return paths


def assert_method_matrix(fixture: dict[str, Any], repository_text: str, handler_text: str) -> None:
    methods = rows_by_route(fixture, "method_matrix")
    readiness_rows = rows_by_route(load_json(INTERFACE_READINESS_FIXTURE_PATH), "method_matrix")
    type_rows = rows_by_route(load_json(TYPES_IMPLEMENTATION_FIXTURE_PATH), "route_type_matrix")
    for route_id, row in methods.items():
        readiness_row = readiness_rows[route_id]
        type_row = type_rows[route_id]
        require(row.get("operation") == readiness_row.get("operation"), f"{route_id} readiness operation drifted")
        require(row.get("operation") == type_row.get("operation"), f"{route_id} type operation drifted")
        require(row.get("request_type") == type_row.get("request_type"), f"{route_id} request type drifted")
        require(row.get("result_type") == type_row.get("result_type"), f"{route_id} result type drifted")
        for literal_key in ("operation", "request_type", "result_type"):
            require(str(row.get(literal_key) or "") in repository_text, f"repository missing {route_id} {literal_key}")
        require(str(row.get("handler") or "") in handler_text, f"handler missing {route_id} handler")
        require(str(row.get("json_mapper") or "") in repository_text, f"repository missing {route_id} mapper")


def assert_source(fixture: dict[str, Any], paths: dict[str, Path]) -> None:
    repository_text = paths["repository_file"].read_text(encoding="utf-8")
    handler_text = paths["handler_file"].read_text(encoding="utf-8")
    contract_text = paths["contract_type_file"].read_text(encoding="utf-8")
    server_text = paths["server_file"].read_text(encoding="utf-8")
    test_text = paths["test_file"].read_text(encoding="utf-8")

    for literal in fixture.get("required_source_literals") or []:
        haystack = "\n".join([repository_text, handler_text, contract_text, server_text])
        require(str(literal) in haystack, f"missing source literal: {literal}")
    require(
        re.search(r"\bcontrolPlaneReadRepo\s+ControlPlaneReadRepository\b", server_text) is not None,
        "server must expose repository injection point",
    )
    require("s.controlPlaneReadDataStore()" not in handler_text, "read handlers must not call fake store directly")
    require("s.controlPlaneReadRepository()" in handler_text, "read handlers must call repository interface")
    require("TenantState" in contract_text and "UsageSnapshot" in contract_text, "typed summaries must cover response fixture fields")
    for literal in fixture.get("required_test_literals") or []:
        require(str(literal) in test_text, f"missing test literal: {literal}")
    assert_method_matrix(fixture, repository_text, handler_text)


def assert_no_forbidden_sources(fixture: dict[str, Any]) -> None:
    configured = set(fixture.get("forbidden_source_literals") or [])
    require(EXPECTED_FORBIDDEN_SOURCE_LITERALS.issubset(configured), "forbidden source literals drifted")
    for path in (REPO_ROOT / "services/platform/internal").rglob("*.go"):
        text = path.read_text(encoding="utf-8")
        for literal in configured:
            require(
                literal not in text,
                f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r}",
            )
    for forbidden_path in (
        "services/platform/internal/httpapi/control_plane_read_repository_adapter.go",
        "services/platform/internal/httpapi/control_plane_read_store_selector.go",
        "services/platform/internal/httpapi/control_plane_read_auth_middleware.go",
        "services/platform/migrations/control_plane_read",
        "contracts/radish-oidc-token-validation.schema.json",
    ):
        require(not (REPO_ROOT / forbidden_path).exists(), f"forbidden artifact exists: {forbidden_path}")


def assert_no_side_effect_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("no_side_effect_policy") or {}
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(policy.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(policy.get("forbidden_side_effects") or [])),
        "missing forbidden side effects",
    )
    response_policy = fixture.get("response_shape_policy") or {}
    for key in (
        "must_keep_response_fixture_shape",
        "must_keep_fake_auth_boundary",
        "must_keep_failure_envelope",
        "must_keep_no_side_effects",
    ):
        require(response_policy.get(key) is True, f"{key} must remain true")


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-control-plane-durable-read-foundation-v1.py"
    entry_checker = "check-control-plane-read-implementation-entry-review-v1.py"
    require(current_checker in check_repo, "check-repo.py must run durable read foundation check")
    require(
        check_repo.index(entry_checker) < check_repo.index(current_checker),
        "durable read foundation check must run after implementation entry review",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    paths = assert_implementation_boundary(fixture)
    assert_source(fixture, paths)
    assert_no_forbidden_sources(fixture)
    assert_no_side_effect_policy(fixture)
    assert_docs_and_check_repo(fixture)
    print("control plane durable read foundation v1 checks passed.")


if __name__ == "__main__":
    main()
