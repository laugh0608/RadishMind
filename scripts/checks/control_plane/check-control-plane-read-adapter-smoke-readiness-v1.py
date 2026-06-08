#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-adapter-smoke-readiness-v1.json"
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-adapter-implementation-plan-v1.json"
)
SCHEMA_ARTIFACT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-artifact-manifest-readiness-v1.json"
)
SELECTOR_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selector-smoke-readiness-v1.json"
)
PRODUCTION_AUTH_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-production-auth-readiness-v1.json"
)
RUNNER_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-implementation-v1.json"
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
    "control-plane-read-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "repository_adapter_implementation_plan_defined",
    ),
    "control-plane-read-schema-artifact-manifest-readiness-v1": (
        SCHEMA_ARTIFACT_FIXTURE_PATH,
        "schema_artifact_manifest_readiness_defined",
    ),
    "control-plane-read-store-selector-smoke-readiness-v1": (
        SELECTOR_SMOKE_FIXTURE_PATH,
        "store_selector_smoke_readiness_defined",
    ),
    "control-plane-read-production-auth-readiness-v1": (
        PRODUCTION_AUTH_FIXTURE_PATH,
        "production_auth_readiness_defined",
    ),
    "control-plane-read-repository-contract-smoke-runner-implementation-v1": (
        RUNNER_IMPLEMENTATION_FIXTURE_PATH,
        "repository_contract_smoke_runner_implemented",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "adapter_smoke_ready",
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
    "store_selector_smoke_ready",
    "store_selector_implemented",
    "formal_read_store_config_ready",
    "production_auth_ready",
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
EXPECTED_PLANNED_ARTIFACTS = {
    "scripts/checks/fixtures/control-plane-read-adapter-smoke-v1.json",
    "scripts/checks/control_plane/check-control-plane-read-adapter-smoke-v1.py",
    "services/platform/internal/httpapi/control_plane_read_repository_interface.go",
    "services/platform/internal/httpapi/control_plane_read_repository_adapter.go",
    "services/platform/internal/httpapi/control_plane_read_repository_adapter_test.go",
    "services/platform/internal/httpapi/control_plane_read_repository_adapter_contract_smoke_test.go",
}
EXPECTED_GATE_IDS = {
    "repository_adapter_implementation_plan_consumed",
    "schema_artifact_manifest_readiness_consumed",
    "store_selector_smoke_readiness_consumed",
    "production_auth_readiness_consumed",
    "static_runner_consumed",
    "adapter_smoke_contract_defined",
    "schema_artifact_manifest_gate",
    "store_selector_smoke_gate",
    "production_auth_gate",
    "repository_adapter_implementation_gate",
    "adapter_smoke_execution_gate",
    "production_api_consumer_gate",
    "no_adapter_smoke_artifacts_leaked",
}
SATISFIED_GATE_IDS = {
    "repository_adapter_implementation_plan_consumed",
    "schema_artifact_manifest_readiness_consumed",
    "store_selector_smoke_readiness_consumed",
    "production_auth_readiness_consumed",
    "static_runner_consumed",
    "adapter_smoke_contract_defined",
}
NOT_SATISFIED_GATE_IDS = {
    "schema_artifact_manifest_gate",
    "store_selector_smoke_gate",
    "production_auth_gate",
    "repository_adapter_implementation_gate",
    "adapter_smoke_execution_gate",
    "production_api_consumer_gate",
}
EXPECTED_REQUIRED_INPUTS = {
    "schema artifact manifest readiness",
    "store selector smoke readiness",
    "production auth readiness",
    "static runner case",
    "repository adapter implementation plan",
}
EXPECTED_FAILURE_CODES = {
    "identity_context_missing",
    "tenant_binding_missing",
    "tenant_binding_denied",
    "scope_denied",
    "invalid_filter",
    "database_read_disabled",
    "invalid_read_store_mode",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "auth_context_contract_mismatch",
    "schema_migration_not_applied",
    "schema_version_mismatch",
}
EXPECTED_AUTH_ROUTE_FAILURES = {
    "identity_context_missing",
    "tenant_binding_missing",
    "tenant_binding_denied",
    "scope_denied",
    "auth_context_contract_mismatch",
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
    "control-plane-read-adapter-smoke-v1.json",
    "check-control-plane-read-adapter-smoke-v1.py",
    "control_plane_read_repository_interface.go",
    "control_plane_read_repository_adapter.go",
    "control_plane_read_repository_adapter_test.go",
    "control_plane_read_repository_adapter_contract_smoke_test.go",
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
    "RADISHMIND_CONTROL_PLANE_READ_STORE",
    "SelectControlPlaneReadStore",
    "ControlPlaneReadStoreSelector",
    "oidc.Provider",
    "github.com/coreos/go-oidc",
    "ValidateToken",
    "VerifyToken",
    "NewRadishOIDCClient",
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
        fixture.get("kind") == "control_plane_read_adapter_smoke_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-adapter-smoke-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "adapter_smoke_readiness_defined",
        "adapter smoke readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_adapter_smoke_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("adapter_smoke_boundary") or {}
    adapter_boundary = load_json(ADAPTER_PLAN_FIXTURE_PATH).get("implementation_plan_boundary") or {}
    require(
        boundary.get("status") == "adapter_smoke_readiness_defined_not_implemented",
        "adapter smoke boundary status drifted",
    )
    for field in (
        "future_interface_file",
        "future_adapter_file",
        "future_adapter_test_file",
        "future_adapter_contract_smoke_test_file",
    ):
        require(boundary.get(field) == adapter_boundary.get(field), f"{field} must match adapter plan")
    require(
        boundary.get("current_static_runner_file") == adapter_boundary.get("static_runner_file"),
        "static runner file must match adapter plan",
    )
    require(
        boundary.get("current_adapter_plan_fixture")
        == "scripts/checks/fixtures/control-plane-read-repository-adapter-implementation-plan-v1.json",
        "adapter plan fixture path drifted",
    )
    require((REPO_ROOT / str(boundary.get("current_static_runner_file") or "")).exists(), "static runner missing")
    for path_field in (
        "future_adapter_smoke_fixture",
        "future_adapter_smoke_checker",
        "future_interface_file",
        "future_adapter_file",
        "future_adapter_test_file",
        "future_adapter_contract_smoke_test_file",
    ):
        relative_path = str(boundary.get(path_field) or "")
        require(relative_path, f"{path_field} missing")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this readiness slice")
    for field in (
        "adapter_smoke_fixture_created_in_this_slice",
        "adapter_smoke_checker_created_in_this_slice",
        "interface_file_created_in_this_slice",
        "adapter_file_created_in_this_slice",
        "adapter_test_created_in_this_slice",
        "adapter_contract_smoke_test_created_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "repository_interface_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_planned_adapter_smoke_artifacts(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(item.get("path") or ""): item
        for item in fixture.get("planned_adapter_smoke_artifacts") or []
        if isinstance(item, dict)
    }
    require(set(artifacts) == EXPECTED_PLANNED_ARTIFACTS, "planned adapter smoke artifacts drifted")
    for relative_path, artifact in artifacts.items():
        require(artifact.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this readiness slice")


def assert_readiness_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("readiness_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "adapter smoke readiness gate ids drifted")
    for gate_id in SATISFIED_GATE_IDS:
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must remain satisfied")
    for gate_id in NOT_SATISFIED_GATE_IDS:
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage")
    require(
        gates["no_adapter_smoke_artifacts_leaked"].get("status") == "required_now",
        "adapter smoke artifact leak gate status drifted",
    )


def assert_dependency_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("adapter_smoke_dependency_contract") or {}
    require(
        contract.get("status") == "adapter_smoke_dependencies_defined_not_ready",
        "adapter smoke dependency contract status drifted",
    )
    rows = {
        str(row.get("source") or ""): row
        for row in contract.get("required_before_adapter_smoke") or []
        if isinstance(row, dict)
    }
    require(set(rows) == set(EXPECTED_DEPENDENCIES), "adapter smoke dependency contract sources drifted")
    for source, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        row = rows[source]
        require(row.get("source_status") == expected_status, f"{source} source status drifted")
        actual = load_json(path).get("slice", {}).get("status")
        require(actual == expected_status, f"{source} fixture status drifted")
        if source == "control-plane-read-repository-contract-smoke-runner-implementation-v1":
            require(row.get("ready_for_adapter_smoke_now") is True, "static runner must remain consumable now")
        else:
            require(row.get("ready_for_adapter_smoke_now") is False, f"{source} must not be ready now")


def assert_route_adapter_smoke_matrix(fixture: dict[str, Any]) -> None:
    rows = rows_by_route(fixture, "route_adapter_smoke_matrix")
    adapter_rows = rows_by_route(load_json(ADAPTER_PLAN_FIXTURE_PATH), "adapter_route_matrix")
    runner_rows = rows_by_route(load_json(RUNNER_IMPLEMENTATION_FIXTURE_PATH), "route_runner_matrix")
    schema_rows = rows_by_route(load_json(SCHEMA_ARTIFACT_FIXTURE_PATH), "route_schema_artifact_matrix")
    selector_rows = rows_by_route(load_json(SELECTOR_SMOKE_FIXTURE_PATH), "route_selector_smoke_matrix")
    auth_rows = rows_by_route(load_json(PRODUCTION_AUTH_FIXTURE_PATH), "route_scope_projection_matrix")

    for route_id, row in rows.items():
        adapter = adapter_rows[route_id]
        runner = runner_rows[route_id]
        schema = schema_rows[route_id]
        selector = selector_rows[route_id]
        auth = auth_rows[route_id]
        require(row.get("operation") == adapter.get("operation"), f"{route_id} adapter operation drifted")
        require(row.get("operation") == runner.get("operation"), f"{route_id} runner operation drifted")
        require(row.get("operation") == schema.get("operation"), f"{route_id} schema operation drifted")
        require(row.get("operation") == selector.get("operation"), f"{route_id} selector operation drifted")
        for field in ("request_type", "result_type"):
            require(row.get(field) == adapter.get(field), f"{route_id} adapter {field} drifted")
            require(row.get(field) == runner.get(field), f"{route_id} runner {field} drifted")
        require(row.get("required_scope") == auth.get("required_scope"), f"{route_id} auth scope drifted")
        require(set(row.get("required_inputs") or []) == EXPECTED_REQUIRED_INPUTS, f"{route_id} inputs drifted")
        for field in (
            "schema_artifact_manifest_readiness_consumed",
            "store_selector_smoke_readiness_consumed",
            "production_auth_readiness_consumed",
            "static_runner_case_consumed",
            "adapter_plan_consumed",
            "smoke_required",
        ):
            require(row.get(field) is True, f"{route_id} {field} must remain true")
        for field in (
            "adapter_smoke_ready",
            "repository_adapter_allowed_now",
            "database_query_allowed_now",
            "production_auth_runtime_allowed_now",
            "production_api_consumer_allowed_now",
            "fake_fallback_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{route_id} {field} must remain false")
        require(adapter.get("adapter_implementation_allowed_now") is False, f"{route_id} adapter must not be allowed")
        require(schema.get("durable_adapter_smoke_dependency_required") is True, f"{route_id} schema dependency drifted")
        require(schema.get("schema_artifact_ready") is False, f"{route_id} schema artifact must not be ready")
        require(selector.get("adapter_smoke_dependency_required") is True, f"{route_id} selector dependency drifted")
        require(selector.get("selector_smoke_ready") is False, f"{route_id} selector smoke must not be ready")
        require(auth.get("production_auth_smoke_required") is True, f"{route_id} auth smoke requirement drifted")
        require(auth.get("production_auth_ready") is False, f"{route_id} production auth must not be ready")


def assert_failure_and_safety_policies(fixture: dict[str, Any]) -> None:
    failures = set(fixture.get("failure_mapping") or [])
    require(EXPECTED_FAILURE_CODES.issubset(failures), "adapter smoke readiness failure mapping missing codes")
    adapter_failures = set(load_json(ADAPTER_PLAN_FIXTURE_PATH).get("failure_mapping") or [])
    selector_failures = set(load_json(SELECTOR_SMOKE_FIXTURE_PATH).get("failure_mapping") or [])
    schema_failures = set(load_json(SCHEMA_ARTIFACT_FIXTURE_PATH).get("failure_mapping") or [])
    production_failures = set(load_json(PRODUCTION_AUTH_FIXTURE_PATH).get("failure_mapping") or [])
    require(adapter_failures.issubset(failures), "adapter smoke readiness must include adapter plan failures")
    require(selector_failures.issubset(failures), "adapter smoke readiness must include selector smoke failures")
    require(schema_failures.issubset(failures), "adapter smoke readiness must include schema artifact failures")
    require(EXPECTED_AUTH_ROUTE_FAILURES.issubset(production_failures), "production auth route failures drifted")
    require(EXPECTED_AUTH_ROUTE_FAILURES.issubset(failures), "adapter smoke readiness must include auth route failures")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "adapter_smoke_missing_success_allowed",
        "missing_schema_artifact_manifest_success_allowed",
        "missing_selector_smoke_success_allowed",
        "missing_production_auth_readiness_success_allowed",
        "repository_adapter_missing_success_allowed",
        "repository_mode_fallback_to_fixture_fake_store_allowed",
        "auth_failure_fallback_to_test_auth_allowed",
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


def assert_docs_check_repo_and_no_leak(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-control-plane-read-production-auth-readiness-v1.py"
    current_checker = "check-control-plane-read-adapter-smoke-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run adapter smoke readiness check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "adapter smoke readiness check must run after production auth readiness",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")

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
    assert_adapter_smoke_boundary(fixture)
    assert_planned_adapter_smoke_artifacts(fixture)
    assert_readiness_gates(fixture)
    assert_dependency_contract(fixture)
    assert_route_adapter_smoke_matrix(fixture)
    assert_failure_and_safety_policies(fixture)
    assert_docs_check_repo_and_no_leak(fixture)
    print("control plane read adapter smoke readiness v1 checks passed.")


if __name__ == "__main__":
    main()
