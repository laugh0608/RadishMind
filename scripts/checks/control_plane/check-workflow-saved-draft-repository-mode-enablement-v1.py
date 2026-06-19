#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-mode-enablement-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-store-selector-smoke-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json",
        "draft_store_selector_smoke_implemented",
    ),
    "workflow-saved-draft-schema-artifact-materialization-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json",
        "draft_schema_artifact_materialized_static",
    ),
    "workflow-saved-draft-repository-adapter-implementation-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json",
        "draft_repository_adapter_implemented",
    ),
    "workflow-saved-draft-adapter-smoke-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json",
        "draft_adapter_smoke_executed",
    ),
    "workflow-saved-draft-production-auth-runtime-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json",
        "draft_production_auth_runtime_bridge_implemented",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "repository_store_mode_enabled",
    "durable_persistence_ready",
    "database_connection_ready",
    "database_query_ready",
    "sql_migration_created",
    "migration_runner_implemented",
    "schema_version_table_created",
    "oidc_middleware_ready",
    "token_validation_ready",
    "workspace_membership_adapter_ready",
    "production_api_consumer_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_CONFIG_GATES = {
    "memory_dev_default": "satisfied",
    "repository_reserved_fail_closed": "satisfied",
    "repository_disabled_reserved_fail_closed": "satisfied",
    "unknown_mode_fail_closed": "satisfied",
    "repository_runtime_enablement_config": "not_satisfied",
    "selector_adapter_binding": "not_satisfied",
}
EXPECTED_SCHEMA_GATES = {
    "static_schema_artifact_materialized": "satisfied",
    "adapter_schema_preflight_implemented": "satisfied",
    "executable_sql_migration_available": "not_satisfied",
    "schema_version_table_available": "not_satisfied",
    "runtime_migration_runner_available": "not_satisfied",
}
EXPECTED_DEPENDENCY_GATES = {
    "repository_adapter_implemented": "satisfied",
    "adapter_smoke_executed_with_fake_query_executor": "satisfied",
    "production_auth_runtime_bridge_implemented": "satisfied",
    "real_repository_query_executor_smoke": "not_satisfied",
    "oidc_middleware_token_validation_membership_adapter": "not_satisfied",
    "production_api_consumer_smoke": "not_satisfied",
}
EXPECTED_FAILURE_CODES = {
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_auth_context_contract_mismatch",
    "draft_identity_context_missing",
    "draft_tenant_binding_missing",
    "draft_workspace_membership_denied",
    "draft_application_scope_denied",
    "draft_owner_scope_denied",
    "draft_scope_grant_missing",
    "draft_store_unavailable",
    "draft_store_contract_mismatch",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "database_connection_count=0",
    "sql_execution_count=0",
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {
        str(row.get(id_field) or ""): row
        for row in fixture.get(key) or []
        if isinstance(row, dict)
    }
    require(rows, f"{key} must not be empty")
    return rows


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    missing = sorted(set(EXPECTED_DEPENDENCIES) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"{dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"{dependency_id} status drifted")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "workflow_saved_draft_repository_mode_enablement_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-repository-mode-enablement-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_repository_mode_enablement_review_defined",
        "status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_runtime_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("runtime_boundary") or {}
    require(
        boundary.get("status") == "repository_mode_enablement_review_defined_no_runtime_enablement",
        "runtime boundary status drifted",
    )
    require(boundary.get("decision") == "repository_mode_not_enabled", "runtime decision drifted")
    for field in (
        "repository_store_mode_success_allowed_now",
        "selector_adapter_binding_allowed_now",
        "database_connection_allowed_now",
        "sql_execution_allowed_now",
        "migration_runner_allowed_now",
        "oidc_middleware_allowed_now",
        "token_validation_allowed_now",
        "membership_adapter_allowed_now",
        "production_api_allowed_now",
        "executor_allowed_now",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")
    for field in (
        "selector_file",
        "route_store_file",
        "repository_adapter_file",
        "production_auth_runtime_file",
    ):
        relative_path = str(boundary.get(field) or "")
        require(relative_path, f"{field} missing")
        require((REPO_ROOT / relative_path).exists(), f"{field} path missing: {relative_path}")


def assert_gate_matrix(fixture: dict[str, Any], key: str, expected: dict[str, str]) -> None:
    rows = rows_by_id(fixture, key, "gate_id")
    require(set(rows) == set(expected), f"{key} gate ids drifted")
    for gate_id, status in expected.items():
        require(rows[gate_id].get("status") == status, f"{gate_id} status drifted")


def assert_failure_rollback_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "repository_mode_enablement_fail_closed_review_defined",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    require(
        mapping.get("reserved_repository_mode_failure_code") == "repository_store_disabled",
        "reserved repository failure drifted",
    )
    require(mapping.get("unknown_mode_failure_code") == "invalid_draft_store_mode", "unknown mode drifted")
    for field in (
        "schema_preflight_failure_must_prevent_query_execution",
        "auth_failure_must_not_return_draft_payload",
        "store_failure_must_not_fallback_to_memory_dev",
        "contract_mismatch_must_not_leak_database_detail",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

    rollback = fixture.get("rollback_policy") or {}
    rollback_path = str(rollback.get("static_rollback_evidence") or "")
    require(rollback.get("static_rollback_evidence_status") == "satisfied", "rollback evidence status drifted")
    require((REPO_ROOT / rollback_path).exists(), f"rollback evidence missing: {rollback_path}")
    require(rollback.get("runtime_rollback_runner_status") == "not_satisfied", "rollback runner drifted")
    require(
        rollback.get("repository_mode_rollback_smoke_status") == "not_satisfied",
        "rollback smoke drifted",
    )
    require(
        rollback.get("rollback_to_memory_dev_after_repository_failure_allowed") is False,
        "rollback fallback must remain false",
    )

    fallback = fixture.get("no_fallback_policy") or {}
    for field, value in fallback.items():
        require(value is False, f"{field} must remain false")
    side_effects = fixture.get("no_side_effect_policy") or {}
    missing_counters = sorted(
        EXPECTED_SIDE_EFFECT_COUNTERS - set(side_effects.get("side_effect_counters_must_remain") or [])
    )
    require(not missing_counters, f"missing side effect counters: {missing_counters}")
    require(side_effects.get("forbidden_side_effects"), "forbidden side effects must be listed")


def assert_sources_and_artifact_guard(fixture: dict[str, Any]) -> None:
    failure_source = read("services/platform/internal/httpapi/workflow_saved_draft.go")
    for failure_code in EXPECTED_FAILURE_CODES:
        require(failure_code in failure_source, f"saved draft failure constant missing {failure_code}")

    selector_source = read("services/platform/internal/httpapi/workflow_saved_draft_store_selector.go")
    route_source = read("services/platform/internal/httpapi/workflow_saved_draft_http.go")
    require("case WorkflowSavedDraftStoreModeRepository:" in selector_source, "repository selector case missing")
    require("SavedWorkflowDraftFailureRepositoryStoreDisabled" in selector_source, "repository disabled mapping missing")
    require("SavedWorkflowDraftFailureInvalidStoreMode" in selector_source, "invalid mode mapping missing")
    require("SelectWorkflowSavedDraftStore(cfg.WorkflowSavedDraftStoreMode" in route_source, "route store selector missing")

    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")
    for root in guard.get("forbid_sql_files_under") or []:
        sql_files = sorted((REPO_ROOT / str(root)).rglob("*.sql"))
        require(not sql_files, f"SQL files must not exist under {root}: {sql_files}")
    for source_path in guard.get("source_files_to_scan") or []:
        source = read(str(source_path))
        for literal in guard.get("forbidden_source_literals") or []:
            require(str(literal) not in source, f"{source_path} contains forbidden literal: {literal}")


def assert_docs_testing_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        content = read(path)
        for literal in reference.get("must_contain") or []:
            require(str(literal) in content, f"{path} missing {literal}")

    strategy = fixture.get("testing_strategy") or {}
    require(
        strategy.get("status") == "repository_mode_enablement_checker_only_no_runtime_enablement",
        "testing strategy drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "go test ./internal/httpapi",
        "workflow_saved_draft_repository_mode_enablement_checker",
        "workflow_saved_draft_adapter_smoke_checker",
        "workflow_saved_draft_production_auth_runtime_checker",
        "./scripts/check-repo.sh --fast",
        "./scripts/check-repo.sh",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_enable_repository_store_mode",
        "does_not_connect_database",
        "does_not_run_sql",
        "does_not_call_oidc",
        "does_not_validate_token",
        "does_not_create_membership_adapter",
        "does_not_create_production_api",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-repository-mode-enablement-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "production auth runtime checker missing from check-repo")
    require(current_checker in check_repo, "repository mode enablement checker missing from check-repo")
    require(next_checker in check_repo, "product surface recheck missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "repository mode enablement checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_runtime_boundary(fixture)
    assert_gate_matrix(fixture, "config_gate_matrix", EXPECTED_CONFIG_GATES)
    assert_gate_matrix(fixture, "schema_preflight_matrix", EXPECTED_SCHEMA_GATES)
    assert_gate_matrix(fixture, "dependency_gate_matrix", EXPECTED_DEPENDENCY_GATES)
    assert_failure_rollback_and_safety(fixture)
    assert_sources_and_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft repository mode enablement v1 checks passed.")


if __name__ == "__main__":
    main()
