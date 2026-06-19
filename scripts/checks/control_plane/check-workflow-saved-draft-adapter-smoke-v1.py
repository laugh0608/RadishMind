#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

DEPENDENCIES = {
    "workflow-saved-draft-adapter-smoke-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-readiness-v1.json",
        "draft_adapter_smoke_readiness_defined",
    ),
    "workflow-saved-draft-repository-adapter-implementation-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json",
        "draft_repository_adapter_implemented",
    ),
    "workflow-saved-draft-repository-contract-smoke-runner-implementation-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.json",
        "draft_repository_contract_smoke_runner_implemented",
    ),
    "workflow-saved-draft-store-selector-smoke-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json",
        "draft_store_selector_smoke_implemented",
    ),
    "workflow-saved-draft-schema-artifact-materialization-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json",
        "draft_schema_artifact_materialized_static",
    ),
    "workflow-saved-draft-production-auth-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-readiness-v1.json",
        "draft_production_auth_readiness_defined",
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
    "production_auth_runtime_ready",
    "radish_oidc_ready",
    "token_validation_ready",
    "workspace_membership_adapter_ready",
    "production_api_consumer_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
}
EXPECTED_OPERATIONS = {
    "SaveWorkflowDraftRecord": "workflow_drafts:write",
    "ReadWorkflowDraftRecord": "workflow_drafts:read",
    "ListWorkflowDraftRecords": "workflow_drafts:read",
}
EXPECTED_FAILURE_CODES = {
    "draft_scope_denied",
    "draft_not_found",
    "draft_schema_version_unsupported",
    "draft_payload_invalid",
    "draft_version_conflict",
    "draft_store_unavailable",
    "draft_store_contract_mismatch",
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_auth_context_contract_mismatch",
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_identity_context_missing",
    "draft_tenant_binding_missing",
    "draft_workspace_membership_denied",
    "draft_scope_grant_missing",
}
EXPECTED_EXECUTED_FAILURE_CODES = {
    "draft_not_found",
    "draft_version_conflict",
    "draft_store_unavailable",
    "draft_store_contract_mismatch",
    "draft_auth_context_contract_mismatch",
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_scope_grant_missing",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "database_connection",
    "sql_execution",
    "oidc_network_call",
    "workflow_execution",
    "confirmation_decision",
    "business_writeback",
    "replay_execution",
    "production_api_call",
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


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    missing = sorted(set(DEPENDENCIES) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, (relative_path, expected_status) in DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"{dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"{dependency_id} status drifted")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(fixture.get("kind") == "workflow_saved_draft_adapter_smoke_v1", "kind drifted")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-saved-draft-adapter-smoke-v1", "slice id drifted")
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(slice_info.get("status") == "draft_adapter_smoke_executed", "status drifted")
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_execution_boundary(fixture: dict[str, Any]) -> tuple[str, str, str]:
    boundary = fixture.get("execution_boundary") or {}
    require(
        boundary.get("status") == "adapter_smoke_executed_with_injected_fake_query_executor",
        "execution boundary status drifted",
    )
    for field in (
        "uses_static_contract_smoke_cases",
        "uses_repository_adapter",
        "uses_injected_fake_query_executor",
    ):
        require(boundary.get(field) is True, f"{field} must remain true")
    for field in (
        "uses_dev_http_route",
        "starts_service",
        "connects_database",
        "runs_sql",
        "calls_oidc",
        "enables_repository_store_mode",
        "creates_production_api",
        "publishes_or_runs_workflow",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")
    for field in (
        "fixture",
        "checker",
        "guard",
        "interface_file",
        "adapter_file",
        "adapter_test_file",
        "adapter_smoke_test_file",
        "static_contract_runner_file",
        "schema_manifest",
    ):
        relative_path = str(boundary.get(field) or "")
        require(relative_path, f"{field} missing")
        require((REPO_ROOT / relative_path).exists(), f"{relative_path} missing")
    manifest = load_json(REPO_ROOT / str(boundary.get("schema_manifest")))
    require(
        boundary.get("store_schema_version") == manifest.get("store_schema_version"),
        "store schema version must match manifest",
    )
    return (
        str(boundary.get("adapter_file")),
        str(boundary.get("adapter_smoke_test_file")),
        str(boundary.get("static_contract_runner_file")),
    )


def assert_operation_smoke_matrix(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_smoke_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == set(EXPECTED_OPERATIONS), "operation smoke matrix must cover save/read/list")
    for operation, expected_scope in EXPECTED_OPERATIONS.items():
        row = rows[operation]
        require(row.get("required_scope") == expected_scope, f"{operation} scope drifted")
        require(row.get("adapter_method") == operation, f"{operation} adapter method drifted")
        require(row.get("static_contract_case_consumed") is True, f"{operation} static case not consumed")
        require(row.get("adapter_method_called") is True, f"{operation} adapter method not called")
        require(row.get("fake_query_executor_call_count") == 1, f"{operation} query call count drifted")
        for field in (
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "fallback_to_fixture_allowed",
            "fallback_to_dev_http_route_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")


def assert_failure_and_side_effect_policy(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(mapping.get("status") == "adapter_smoke_fail_closed_executed", "failure status drifted")
    require(EXPECTED_FAILURE_CODES.issubset(set(mapping.get("required_failure_codes") or [])), "failure codes missing")
    require(
        EXPECTED_EXECUTED_FAILURE_CODES.issubset(set(mapping.get("executed_failure_codes") or [])),
        "executed failure codes missing",
    )
    for field in (
        "version_conflict_must_not_be_rewritten",
        "not_found_must_not_fallback_to_sample",
        "scope_denied_must_not_return_draft_payload",
        "contract_mismatch_must_not_leak_database_detail",
        "schema_failure_must_prevent_query_execution",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

    fallback = fixture.get("no_fallback_policy") or {}
    for field in (
        "fallback_to_memory_dev_store_allowed",
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_dev_http_route_allowed",
        "fallback_to_test_auth_allowed",
        "success_on_missing_schema_artifact_allowed",
        "success_on_auth_context_mismatch_allowed",
        "success_on_contract_mismatch_allowed",
        "reserved_store_mode_success_allowed",
        "unknown_store_mode_success_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
    require(
        side_effects.get("allowed_test_state") == "injected_fake_query_executor_state",
        "allowed test state drifted",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effects.get("forbidden_side_effects") or [])),
        "forbidden side effects missing",
    )


def assert_go_sources(adapter_file: str, smoke_test_file: str, static_runner_file: str, fixture: dict[str, Any]) -> None:
    adapter_source = read(adapter_file)
    smoke_test_source = read(smoke_test_file)
    static_runner_source = read(static_runner_file)
    for needle in (
        "func NewSavedWorkflowDraftRepositoryAdapter(",
        "type SavedWorkflowDraftRepositoryAdapter struct",
        "SavedWorkflowDraftRepositoryQueryExecutor",
    ):
        require(needle in adapter_source, f"adapter source missing {needle}")
    for needle in (
        "func RunSavedWorkflowDraftRepositoryContractSmoke(",
        "func savedWorkflowDraftRepositoryContractSmokeCases()",
        "SavedWorkflowDraftRepositoryContractSmokeCase",
    ):
        require(needle in static_runner_source, f"static runner missing {needle}")
    for test_name in fixture.get("required_tests") or []:
        require(str(test_name) in smoke_test_source, f"adapter smoke test missing {test_name}")
    for needle in (
        "RunSavedWorkflowDraftRepositoryContractSmoke",
        "savedWorkflowDraftRepositoryContractSmokeCases()",
        "NewSavedWorkflowDraftRepositoryAdapter",
        "newFakeSavedWorkflowDraftRepositoryQueryExecutor",
        "executor.saveCalls != 1",
        "executor.readCalls != 1",
        "executor.listCalls != 1",
        "SavedWorkflowDraftFailureVersionConflict",
        "SavedWorkflowDraftFailureNotFound",
        "SavedWorkflowDraftFailureStoreUnavailable",
        "SavedWorkflowDraftFailureStoreContractMismatch",
        "SavedWorkflowDraftFailureAuthContextMismatch",
        "SavedWorkflowDraftFailureSchemaMigrationNotApplied",
        "SavedWorkflowDraftFailureStoreSchemaVersionMismatch",
        "SavedWorkflowDraftFailureStoreMigrationUnavailable",
        "SavedWorkflowDraftFailureScopeGrantMissing",
    ):
        require(needle in smoke_test_source, f"adapter smoke test missing {needle}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")
    sql_files = list((REPO_ROOT / "services/platform").rglob("*.sql"))
    require(not sql_files, f"SQL files must not be introduced: {sql_files}")
    for source_path in guard.get("source_files_to_scan") or []:
        source = read(str(source_path))
        for literal in guard.get("forbidden_source_literals") or []:
            require(str(literal) not in source, f"{source_path} contains forbidden literal: {literal}")


def assert_docs_and_testing(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        require(path, "doc reference path missing")
        content = read(path)
        for needle in reference.get("must_contain") or []:
            require(str(needle) in content, f"{path} missing {needle}")
    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "adapter_smoke_checker_and_go_tests", "testing strategy drifted")
    required = set(strategy.get("required_checks") or [])
    for expected in (
        "go test ./internal/httpapi",
        "workflow_saved_draft_adapter_smoke_checker",
        "workflow_saved_draft_repository_adapter_implementation_checker",
        "workflow_saved_draft_adapter_smoke_readiness_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected in required, f"missing required check: {expected}")
    for field in (
        "does_not_start_service",
        "does_not_connect_database",
        "does_not_run_sql",
        "does_not_call_oidc",
        "does_not_enable_repository_store_mode",
        "does_not_create_production_api",
        "does_not_publish_or_run_workflow",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    content = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in content, "repository adapter implementation checker missing from check-repo")
    require(current_checker in content, "adapter smoke checker missing from check-repo")
    require(next_checker in content, "product surface recheck missing from check-repo")
    require(
        content.index(previous_checker) < content.index(current_checker) < content.index(next_checker),
        "adapter smoke checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    adapter_file, smoke_test_file, static_runner_file = assert_execution_boundary(fixture)
    assert_operation_smoke_matrix(fixture)
    assert_failure_and_side_effect_policy(fixture)
    assert_go_sources(adapter_file, smoke_test_file, static_runner_file, fixture)
    assert_artifact_guard(fixture)
    assert_docs_and_testing(fixture)
    assert_check_repo_registration()
    print("workflow saved draft adapter smoke v1 checks passed.")


if __name__ == "__main__":
    main()
