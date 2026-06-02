#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selection-readiness-v1.json"
IMPLEMENTATION_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-implementation-readiness-v1.json"
)
REPOSITORY_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-v1.json"
)
DISABLED_DATABASE_GUARD_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-disabled-database-guard-v1.json"
)
AUTH_STORE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-store-transition-preconditions-v1.json"
)
FAKE_STORE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json"
)
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
EXPECTED_DEPENDENCIES = {
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
    "control-plane-read-auth-store-transition-preconditions-v1": (
        AUTH_STORE_FIXTURE_PATH,
        "auth_store_transition_preconditions_defined",
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
EXPECTED_FORBIDDEN_CLAIMS = {
    "store_selector_implemented",
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
EXPECTED_MODES = {"fixture_fake_store", "database", "postgres", "repository", "unknown"}
RESERVED_MODES = {"database", "postgres", "repository"}
EXPECTED_FAILURE_CODES = {
    "database_read_disabled",
    "invalid_read_store_mode",
    "read_store_unavailable",
    "read_store_contract_mismatch",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_SOURCE_ABSENT_LITERALS = {
    "RADISHMIND_CONTROL_PLANE_READ_STORE",
    "SelectControlPlaneReadStore",
    "ControlPlaneReadStoreSelector",
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


def disabled_guard_by_route_id() -> dict[str, dict[str, Any]]:
    guard = load_json(DISABLED_DATABASE_GUARD_FIXTURE_PATH)
    matrix = {
        str(row.get("route_id") or ""): row
        for row in guard.get("route_guard_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "disabled database guard route ids drifted")
    return matrix


def repository_smoke_by_route_id() -> dict[str, dict[str, Any]]:
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
        fixture.get("kind") == "control_plane_read_store_selection_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-store-selection-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "store_selection_readiness_defined",
        "store selection readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_selection_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("selection_boundary") or {}
    implementation = load_json(IMPLEMENTATION_READINESS_FIXTURE_PATH).get("readiness_boundary") or {}
    guard = load_json(DISABLED_DATABASE_GUARD_FIXTURE_PATH).get("disabled_database_guard_boundary") or {}
    require(
        boundary.get("status") == "selection_policy_defined_not_implemented",
        "selection boundary status drifted",
    )
    require(
        boundary.get("current_default_source") == implementation.get("current_store_source"),
        "current default source must match implementation readiness",
    )
    require(
        boundary.get("current_default_source") == guard.get("current_store_source"),
        "current default source must match disabled database guard",
    )
    require(
        boundary.get("future_durable_source") == implementation.get("future_store_source"),
        "future durable source must match implementation readiness",
    )
    require(
        boundary.get("future_durable_source") == guard.get("reserved_disabled_database_source"),
        "future durable source must match disabled database guard",
    )
    require(boundary.get("future_config_key") == "RADISHMIND_CONTROL_PLANE_READ_STORE", "future key drifted")
    for field in (
        "formal_config_entry_created",
        "store_selector_implemented_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_selection_modes(fixture: dict[str, Any]) -> None:
    modes = {
        str(mode.get("mode") or ""): mode
        for mode in fixture.get("selection_modes") or []
        if isinstance(mode, dict)
    }
    require(set(modes) == EXPECTED_MODES, "selection mode set drifted")
    fake = modes["fixture_fake_store"]
    require(fake.get("status") == "current_default", "fixture fake store must remain default")
    require(fake.get("enabled_now") is True, "fixture fake store must remain enabled")
    require(fake.get("failure_code") is None, "fixture fake store must not have a failure code")
    require(fake.get("fallback_allowed") is False, "fixture fake store fallback flag must remain false")
    for mode_name in RESERVED_MODES:
        mode = modes[mode_name]
        require(mode.get("enabled_now") is False, f"{mode_name} must remain disabled")
        require(
            mode.get("status") == "reserved_disabled_until_schema_and_repository_are_ready",
            f"{mode_name} status drifted",
        )
        require(mode.get("failure_code") == "database_read_disabled", f"{mode_name} failure code drifted")
        require(mode.get("fallback_allowed") is False, f"{mode_name} must not fall back to fake store")
    unknown = modes["unknown"]
    require(unknown.get("status") == "fail_closed", "unknown mode must fail closed")
    require(unknown.get("enabled_now") is False, "unknown mode must remain disabled")
    require(unknown.get("failure_code") == "invalid_read_store_mode", "unknown mode failure code drifted")
    require(unknown.get("fallback_allowed") is False, "unknown mode must not fall back to fake store")


def assert_route_selection_matrix(fixture: dict[str, Any]) -> None:
    guard = disabled_guard_by_route_id()
    smoke = repository_smoke_by_route_id()
    matrix = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_selection_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_ROUTE_IDS, "route selection matrix must cover every read route")
    for route_id, row in matrix.items():
        guard_row = guard[route_id]
        smoke_row = smoke[route_id]
        for field in ("method", "path", "read_model", "required_scope"):
            require(row.get(field) == guard_row.get(field), f"{route_id} {field} drifted from guard")
        require(row.get("operation") == smoke_row.get("operation"), f"{route_id} operation drifted from smoke")
        require(row.get("default_mode") == "fixture_fake_store", f"{route_id} default mode drifted")
        require(set(row.get("reserved_modes") or []) == RESERVED_MODES, f"{route_id} reserved modes drifted")
        require(
            row.get("reserved_mode_failure_code") == guard_row.get("failure_code_when_requested"),
            f"{route_id} reserved mode failure code drifted",
        )
        require(row.get("unknown_mode_failure_code") == "invalid_read_store_mode", f"{route_id} unknown failure drifted")
        require(row.get("fallback_to_fake_store_allowed") is False, f"{route_id} fallback must remain disabled")
        require(row.get("side_effect_allowed") is False, f"{route_id} side effects must remain disabled")


def assert_future_selector_and_failures(fixture: dict[str, Any]) -> None:
    selector = fixture.get("future_selector_contract") or {}
    require(selector.get("status") == "planned_not_implemented", "future selector status drifted")
    must_cover = set(selector.get("must_cover") or [])
    for required in (
        "unset source selects fixture_fake_store",
        "database/postgres/repository are reserved disabled until schema and repository gates are satisfied",
        "unknown selector value fails closed",
        "reserved disabled source never falls back to fake store",
        "selector is read-only and has no write, executor, confirmation, writeback or replay side effects",
    ):
        require(required in must_cover, f"future selector must cover {required!r}")
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(selector.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    policy = fixture.get("failure_mapping_policy") or {}
    require(
        policy.get("status") == "required_before_selector_implementation",
        "failure mapping policy status drifted",
    )
    require(
        EXPECTED_FAILURE_CODES.issubset(set(policy.get("required_failure_codes") or [])),
        "failure mapping policy missing required codes",
    )
    for field in (
        "unknown_mode_success_allowed",
        "reserved_mode_fake_fallback_allowed",
        "database_internal_error_exposure_allowed",
    ):
        require(policy.get(field) is False, f"{field} must remain false")


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-control-plane-read-store-selection-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run store selection readiness check")
    require(
        check_repo.index("check-control-plane-read-repository-implementation-readiness-v1.py")
        < check_repo.index(current_checker),
        "store selection readiness check must run after repository implementation readiness",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_selector_implementation_leaked(fixture: dict[str, Any]) -> None:
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
    assert_selection_boundary(fixture)
    assert_selection_modes(fixture)
    assert_route_selection_matrix(fixture)
    assert_future_selector_and_failures(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_selector_implementation_leaked(fixture)
    print("control plane read store selection readiness v1 checks passed.")


if __name__ == "__main__":
    main()
