#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selector-smoke-readiness-v1.json"
)
SELECTOR_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selector-enablement-preconditions-v1.json"
)
SCHEMA_ARTIFACT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-artifact-manifest-readiness-v1.json"
)
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-adapter-implementation-plan-v1.json"
)
STORE_SELECTION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selection-readiness-v1.json"
)
DISABLED_DATABASE_GUARD_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-disabled-database-guard-v1.json"
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
    "control-plane-read-store-selector-enablement-preconditions-v1": (
        SELECTOR_PRECONDITIONS_FIXTURE_PATH,
        "store_selector_enablement_preconditions_defined",
    ),
    "control-plane-read-schema-artifact-manifest-readiness-v1": (
        SCHEMA_ARTIFACT_FIXTURE_PATH,
        "schema_artifact_manifest_readiness_defined",
    ),
    "control-plane-read-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "repository_adapter_implementation_plan_defined",
    ),
    "control-plane-read-store-selection-readiness-v1": (
        STORE_SELECTION_FIXTURE_PATH,
        "store_selection_readiness_defined",
    ),
    "control-plane-read-disabled-database-guard-v1": (
        DISABLED_DATABASE_GUARD_FIXTURE_PATH,
        "disabled_database_guard_defined",
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
    "store_selector_smoke_ready",
    "store_selector_implemented",
    "formal_read_store_config_ready",
    "selector_config_entry_ready",
    "repository_interface_ready",
    "repository_adapter_ready",
    "repository_implementation_ready",
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "migration_files_created",
    "migration_manifest_created",
    "repository_migration_ready",
    "schema_artifact_manifest_ready",
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
EXPECTED_GATE_IDS = {
    "selector_enablement_preconditions_consumed",
    "schema_artifact_manifest_readiness_consumed",
    "repository_adapter_implementation_plan_consumed",
    "selector_smoke_contract_defined",
    "selector_config_entry_gate",
    "selector_code_gate",
    "fixture_default_smoke_gate",
    "reserved_mode_fail_closed_smoke_gate",
    "unknown_mode_fail_closed_smoke_gate",
    "durable_mode_enablement_gate",
    "no_selector_smoke_artifacts_leaked",
}
EXPECTED_PLANNED_ARTIFACTS = {
    "scripts/checks/fixtures/control-plane-read-store-selector-smoke-v1.json",
    "scripts/checks/control_plane/check-control-plane-read-store-selector-smoke-v1.py",
    "services/platform/internal/httpapi/control_plane_read_store_selector.go",
    "services/platform/internal/httpapi/control_plane_read_store_selector_test.go",
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
    "RADISHMIND_CONTROL_PLANE_READ_STORE",
    "SelectControlPlaneReadStore",
    "ControlPlaneReadStoreSelector",
    "control_plane_read_store_selector.go",
    "control_plane_read_store_selector_test.go",
    "control-plane-read-store-selector-smoke-v1.json",
    "check-control-plane-read-store-selector-smoke-v1.py",
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
        fixture.get("kind") == "control_plane_read_store_selector_smoke_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-store-selector-smoke-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "store_selector_smoke_readiness_defined",
        "store selector smoke readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_selector_smoke_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("selector_smoke_boundary") or {}
    selector_boundary = load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH).get("selector_enablement_boundary") or {}
    selection_boundary = load_json(STORE_SELECTION_FIXTURE_PATH).get("selection_boundary") or {}

    require(
        boundary.get("status") == "store_selector_smoke_readiness_defined_not_implemented",
        "selector smoke boundary status drifted",
    )
    for field in ("current_default_source", "future_config_key", "future_selector_name", "future_selector_type"):
        require(boundary.get(field) == selector_boundary.get(field), f"{field} must match selector preconditions")
    require(
        boundary.get("future_config_key") == selection_boundary.get("future_config_key"),
        "future config key must match store selection readiness",
    )
    for path_field in (
        "future_selector_file",
        "future_selector_smoke_fixture",
        "future_selector_smoke_checker",
        "future_selector_test_file",
    ):
        relative_path = str(boundary.get(path_field) or "")
        require(relative_path, f"{path_field} missing")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this readiness slice")
    for field in (
        "formal_config_entry_created",
        "selector_file_created_in_this_slice",
        "selector_test_created_in_this_slice",
        "selector_smoke_fixture_created_in_this_slice",
        "selector_smoke_checker_created_in_this_slice",
        "store_selector_implemented_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_planned_selector_smoke_artifacts(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(item.get("path") or ""): item
        for item in fixture.get("planned_selector_smoke_artifacts") or []
        if isinstance(item, dict)
    }
    require(set(artifacts) == EXPECTED_PLANNED_ARTIFACTS, "planned selector smoke artifacts drifted")
    for relative_path, artifact in artifacts.items():
        require(artifact.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this readiness slice")


def assert_smoke_readiness_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("smoke_readiness_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "selector smoke gate ids drifted")
    for gate_id in (
        "selector_enablement_preconditions_consumed",
        "schema_artifact_manifest_readiness_consumed",
        "repository_adapter_implementation_plan_consumed",
        "selector_smoke_contract_defined",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must remain satisfied")
    for gate_id in (
        "selector_config_entry_gate",
        "selector_code_gate",
        "fixture_default_smoke_gate",
        "reserved_mode_fail_closed_smoke_gate",
        "unknown_mode_fail_closed_smoke_gate",
        "durable_mode_enablement_gate",
    ):
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage")
    require(
        gates["no_selector_smoke_artifacts_leaked"].get("status") == "required_now",
        "selector artifact leak gate status drifted",
    )


def assert_selector_smoke_mode_matrix(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("mode") or ""): row
        for row in fixture.get("selector_smoke_mode_matrix") or []
        if isinstance(row, dict)
    }
    enablement_rows = {
        str(row.get("mode") or ""): row
        for row in load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH).get("selector_mode_enablement_matrix") or []
        if isinstance(row, dict)
    }
    selection_rows = {
        str(row.get("mode") or ""): row
        for row in load_json(STORE_SELECTION_FIXTURE_PATH).get("selection_modes") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_MODES, "selector smoke mode set drifted")
    require(set(enablement_rows) == EXPECTED_MODES, "selector enablement mode set drifted")
    require(set(selection_rows) == EXPECTED_MODES, "selection mode set drifted")

    for mode, row in rows.items():
        enablement = enablement_rows[mode]
        selection = selection_rows[mode]
        require(row.get("current_status") == enablement.get("current_status"), f"{mode} status drifted")
        require(row.get("expected_failure_code") == selection.get("failure_code"), f"{mode} failure drifted")
        require(row.get("expected_selector_input") == selection.get("input"), f"{mode} selector input drifted")
        require(row.get("smoke_required") is True, f"{mode} smoke must remain required")
        for field in (
            "selector_smoke_ready",
            "selector_implementation_allowed_now",
            "fallback_to_fake_store_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{mode} {field} must remain false")
    for mode in RESERVED_MODES:
        require(rows[mode].get("expected_failure_code") == "database_read_disabled", f"{mode} must remain disabled")
    require(rows["unknown"].get("expected_failure_code") == "invalid_read_store_mode", "unknown failure drifted")


def assert_route_selector_smoke_matrix(fixture: dict[str, Any]) -> None:
    rows = rows_by_route(fixture, "route_selector_smoke_matrix")
    enablement_rows = rows_by_route(load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH), "route_selector_enablement_matrix")
    selection_rows = rows_by_route(load_json(STORE_SELECTION_FIXTURE_PATH), "route_selection_matrix")
    runner_rows = rows_by_route(load_json(RUNNER_IMPLEMENTATION_FIXTURE_PATH), "route_runner_matrix")
    for route_id, row in rows.items():
        enablement = enablement_rows[route_id]
        selection = selection_rows[route_id]
        runner = runner_rows[route_id]
        require(row.get("operation") == enablement.get("operation"), f"{route_id} enablement operation drifted")
        require(row.get("operation") == selection.get("operation"), f"{route_id} selection operation drifted")
        require(row.get("operation") == runner.get("operation"), f"{route_id} runner operation drifted")
        require(row.get("default_mode") == selection.get("default_mode"), f"{route_id} default mode drifted")
        require(set(row.get("reserved_modes") or []) == RESERVED_MODES, f"{route_id} reserved modes drifted")
        require(
            row.get("reserved_mode_failure_code") == selection.get("reserved_mode_failure_code"),
            f"{route_id} reserved failure drifted",
        )
        require(
            row.get("unknown_mode_failure_code") == selection.get("unknown_mode_failure_code"),
            f"{route_id} unknown failure drifted",
        )
        for field in ("selector_smoke_required", "adapter_smoke_dependency_required"):
            require(row.get(field) is True, f"{route_id} {field} must remain true")
        for field in ("selector_smoke_ready", "fallback_to_fake_store_allowed", "side_effect_allowed"):
            require(row.get(field) is False, f"{route_id} {field} must remain false")


def assert_failure_and_safety_policies(fixture: dict[str, Any]) -> None:
    failures = set(fixture.get("failure_mapping") or [])
    require(EXPECTED_FAILURE_CODES.issubset(failures), "selector smoke readiness failure mapping missing codes")
    selector_failures = set(load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH).get("failure_mapping") or [])
    require(selector_failures.issubset(failures), "selector smoke readiness must include selector failures")
    schema_failures = set(load_json(SCHEMA_ARTIFACT_FIXTURE_PATH).get("failure_mapping") or [])
    require(schema_failures.issubset(failures), "selector smoke readiness must include schema artifact failures")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "reserved_mode_fake_fallback_allowed",
        "unknown_mode_success_allowed",
        "missing_selector_success_allowed",
        "selector_smoke_missing_success_allowed",
        "database_internal_error_exposure_allowed",
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
    previous_checker = "check-control-plane-read-schema-artifact-manifest-readiness-v1.py"
    current_checker = "check-control-plane-read-store-selector-smoke-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run store selector smoke readiness check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "store selector smoke readiness check must run after schema artifact manifest readiness",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_selector_smoke_implementation_leaked(fixture: dict[str, Any]) -> None:
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
    assert_selector_smoke_boundary(fixture)
    assert_planned_selector_smoke_artifacts(fixture)
    assert_smoke_readiness_gates(fixture)
    assert_selector_smoke_mode_matrix(fixture)
    assert_route_selector_smoke_matrix(fixture)
    assert_failure_and_safety_policies(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_selector_smoke_implementation_leaked(fixture)
    print("control plane read store selector smoke readiness v1 checks passed.")


if __name__ == "__main__":
    main()
