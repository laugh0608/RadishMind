#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from workflow_saved_draft_adapter_smoke_execution_guard import adapter_smoke_execution_file_allowed


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json"
)
ENTRY_REVIEW_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-entry-review-v1.json"
)
MANIFEST_PATH = REPO_ROOT / "services/platform/migrations/workflow_saved_drafts/manifest.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-repository-adapter-implementation-entry-review-v1": (
        ENTRY_REVIEW_FIXTURE_PATH,
        "draft_repository_adapter_implementation_entry_review_defined",
    ),
}
EXPECTED_OPERATIONS = {
    "SaveWorkflowDraftRecord": "workflow_drafts:write",
    "ReadWorkflowDraftRecord": "workflow_drafts:read",
    "ListWorkflowDraftRecords": "workflow_drafts:read",
}
EXPECTED_ACTOR_FIELDS = {
    "tenant_ref",
    "workspace_id",
    "application_id",
    "actor_subject_ref",
    "owner_subject_ref",
    "scope_grants",
    "request_id",
    "audit_ref",
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
EXPECTED_FORBIDDEN_CLAIMS = {
    "repository_store_mode_enabled",
    "adapter_smoke_ready",
    "adapter_smoke_executed",
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
    "workflow_executor_ready",
}
EXPECTED_SOURCE_LITERALS = {
    "type SavedWorkflowDraftRepository interface",
    "type SavedWorkflowDraftRepositoryQueryExecutor interface",
    "type SavedWorkflowDraftRepositorySchemaPreflight struct",
    "func NewSavedWorkflowDraftRepositoryAdapter(",
    "func (adapter SavedWorkflowDraftRepositoryAdapter) SaveWorkflowDraftRecord(",
    "func (adapter SavedWorkflowDraftRepositoryAdapter) ReadWorkflowDraftRecord(",
    "func (adapter SavedWorkflowDraftRepositoryAdapter) ListWorkflowDraftRecords(",
    "savedWorkflowDraftRepositoryStoreSchemaVersion = \"saved_workflow_drafts_store_v1\"",
    "savedWorkflowDraftRepositoryMigrationNotApplied",
    "savedWorkflowDraftRepositoryMigrationUnavailable",
    "SavedWorkflowDraftFailureIdentityContextMissing",
    "SavedWorkflowDraftFailureTenantBindingMissing",
    "SavedWorkflowDraftFailureWorkspaceMembershipDenied",
    "SavedWorkflowDraftFailureScopeGrantMissing",
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
        fixture.get("kind") == "workflow_saved_draft_repository_adapter_implementation_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-repository-adapter-implementation-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_repository_adapter_implemented",
        "repository adapter implementation status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_boundary(fixture: dict[str, Any]) -> tuple[str, str, str]:
    boundary = fixture.get("implementation_boundary") or {}
    require(
        boundary.get("status") == "repository_adapter_boundary_implemented_no_runtime_enablement",
        "implementation boundary status drifted",
    )
    require(boundary.get("query_executor_injected") is True, "query executor boundary must be injected")
    for field in (
        "process_database_connection_created",
        "repository_store_mode_enabled",
        "adapter_smoke_created",
        "production_auth_runtime_created",
        "production_api_created",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")
    interface_file = str(boundary.get("interface_file") or "")
    adapter_file = str(boundary.get("adapter_file") or "")
    test_file = str(boundary.get("adapter_test_file") or "")
    for relative_path in (interface_file, adapter_file, test_file, str(boundary.get("task_card") or "")):
        require(relative_path, "implementation boundary path missing")
        require((REPO_ROOT / relative_path).exists(), f"missing implementation artifact: {relative_path}")
    require(
        boundary.get("store_schema_version") == load_json(MANIFEST_PATH).get("store_schema_version"),
        "store schema version must match migration manifest",
    )
    return interface_file, adapter_file, test_file


def assert_operation_matrix(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == set(EXPECTED_OPERATIONS), "operation matrix must cover save/read/list")
    for operation, expected_scope in EXPECTED_OPERATIONS.items():
        row = rows[operation]
        require(row.get("required_scope") == expected_scope, f"{operation} scope drifted")
        for flag in ("must_validate_actor_context", "must_validate_schema_preflight", "must_not_fallback_to_sample"):
            require(row.get(flag) is True, f"{operation} {flag} must remain true")


def assert_actor_and_failure_mapping(fixture: dict[str, Any]) -> None:
    missing_fields = sorted(EXPECTED_ACTOR_FIELDS - set(fixture.get("actor_context_fields") or []))
    require(not missing_fields, f"missing actor context fields: {missing_fields}")
    mapping = fixture.get("failure_mapping") or {}
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    for flag in (
        "version_conflict_must_not_be_rewritten",
        "not_found_must_not_fallback_to_sample",
        "scope_denied_must_not_return_draft_payload",
        "contract_mismatch_must_not_leak_database_detail",
    ):
        require(mapping.get(flag) is True, f"{flag} must remain true")


def assert_source_files(interface_file: str, adapter_file: str, test_file: str) -> None:
    source = read(interface_file) + read(adapter_file)
    test_source = read(test_file)
    for literal in EXPECTED_SOURCE_LITERALS:
        require(literal in source, f"implementation source missing literal: {literal}")
    for test_name in load_json(FIXTURE_PATH).get("required_tests") or []:
        require(str(test_name) in test_source, f"adapter test missing {test_name}")
    for failure_literal in (
        "SavedWorkflowDraftFailureVersionConflict",
        "SavedWorkflowDraftFailureNotFound",
        "SavedWorkflowDraftFailureStoreSchemaVersionMismatch",
        "SavedWorkflowDraftFailureSchemaMigrationNotApplied",
        "SavedWorkflowDraftFailureStoreUnavailable",
        "SavedWorkflowDraftFailureStoreContractMismatch",
        "SavedWorkflowDraftFailureScopeGrantMissing",
    ):
        require(failure_literal in test_source, f"adapter tests must cover {failure_literal}")


def assert_forbidden_runtime_artifacts(fixture: dict[str, Any], interface_file: str, adapter_file: str) -> None:
    guard = fixture.get("forbidden_runtime_artifacts") or {}
    for relative_path in guard.get("files_must_not_exist") or []:
        if adapter_smoke_execution_file_allowed(REPO_ROOT, str(relative_path)):
            continue
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden runtime artifact exists: {relative_path}")
    scanned = read(interface_file) + read(adapter_file)
    for literal in guard.get("source_literals_must_not_appear") or []:
        require(str(literal) not in scanned, f"repository adapter source contains forbidden runtime literal: {literal}")


def assert_task_card_status(fixture: dict[str, Any]) -> None:
    task_card = str((fixture.get("implementation_boundary") or {}).get("task_card") or "")
    content = read(task_card)
    require("draft_repository_adapter_implemented" in content, "task card status must be implemented")
    require("不启用 `repository` store mode" in content, "task card must preserve repository mode stop line")
    require("不创建 adapter smoke fixture" in content, "task card must preserve adapter smoke stop line")


def assert_check_repo_registration() -> None:
    content = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = (
        "checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-entry-review-v1.py"
    )
    current_checker = "checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in content, "entry review checker missing from check-repo")
    require(current_checker in content, "implementation checker missing from check-repo")
    require(next_checker in content, "product surface recheck missing from check-repo")
    require(
        content.index(previous_checker) < content.index(current_checker) < content.index(next_checker),
        "implementation checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    interface_file, adapter_file, test_file = assert_boundary(fixture)
    assert_operation_matrix(fixture)
    assert_actor_and_failure_mapping(fixture)
    assert_source_files(interface_file, adapter_file, test_file)
    assert_forbidden_runtime_artifacts(fixture, interface_file, adapter_file)
    assert_task_card_status(fixture)
    assert_check_repo_registration()
    print("workflow saved draft repository adapter implementation v1 checks passed.")


if __name__ == "__main__":
    main()
