#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

DEPENDENCIES = {
    "workflow-saved-draft-production-auth-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-readiness-v1.json",
        "draft_production_auth_readiness_defined",
    ),
    "workflow-saved-draft-auth-context-preconditions-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-auth-context-preconditions-v1.json",
        "draft_auth_context_preconditions_defined",
    ),
    "workflow-saved-draft-repository-adapter-implementation-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json",
        "draft_repository_adapter_implemented",
    ),
    "workflow-saved-draft-adapter-smoke-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json",
        "draft_adapter_smoke_executed",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "radish_oidc_ready",
    "oidc_client_implemented",
    "issuer_discovery_runtime_ready",
    "token_validation_ready",
    "auth_middleware_ready",
    "workspace_membership_adapter_ready",
    "repository_store_mode_enabled",
    "durable_persistence_ready",
    "database_schema_ready",
    "database_connection_ready",
    "database_query_ready",
    "database_migration_ready",
    "migration_runner_implemented",
    "production_api_consumer_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_OPERATIONS = {
    "SaveWorkflowDraftRecord": "workflow_drafts:write",
    "ReadWorkflowDraftRecord": "workflow_drafts:read",
    "ListWorkflowDraftRecords": "workflow_drafts:read",
}
EXPECTED_PROJECTED_FIELDS = {
    "request_id",
    "tenant_ref",
    "workspace_id",
    "application_id",
    "actor_subject_ref",
    "owner_subject_ref",
    "scope_grants",
    "audit_ref",
}
EXPECTED_FAILURE_CODES = {
    "draft_identity_context_missing",
    "draft_tenant_binding_missing",
    "draft_workspace_membership_denied",
    "draft_application_scope_denied",
    "draft_owner_scope_denied",
    "draft_scope_grant_missing",
    "draft_auth_context_contract_mismatch",
    "draft_audit_context_missing",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "issuer_network_call_count=0",
    "token_validation_call_count=0",
    "database_write_count=0",
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "issuer_network_call",
    "token_validation_runtime_call",
    "auth_session_mutation",
    "database_write",
    "repository_write",
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
    require(fixture.get("kind") == "workflow_saved_draft_production_auth_runtime_v1", "kind drifted")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-saved-draft-production-auth-runtime-v1", "slice id drifted")
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_production_auth_runtime_bridge_implemented",
        "status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_runtime_boundary(fixture: dict[str, Any]) -> tuple[str, str]:
    boundary = fixture.get("runtime_boundary") or {}
    require(
        boundary.get("status") == "production_auth_runtime_bridge_implemented",
        "runtime boundary status drifted",
    )
    for field in (
        "uses_verified_auth_context",
        "uses_verified_workspace_binding",
        "projects_scope_grants",
        "normalizes_scope_grants",
    ):
        require(boundary.get(field) is True, f"{field} must remain true")
    for field in (
        "creates_oidc_client",
        "fetches_issuer_metadata",
        "validates_token",
        "creates_auth_middleware",
        "creates_membership_adapter",
        "creates_repository_adapter",
        "enables_repository_store_mode",
        "connects_database",
        "creates_production_api",
        "starts_service",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")
    require(boundary.get("accepted_auth_source") == "radish_oidc_verified_context", "auth source drifted")
    source_path = str(boundary.get("runtime_bridge_file") or "")
    test_path = str(boundary.get("runtime_bridge_test_file") or "")
    require((REPO_ROOT / source_path).exists(), "runtime bridge source missing")
    require((REPO_ROOT / test_path).exists(), "runtime bridge test missing")
    return source_path, test_path


def assert_operation_scope_projection(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_scope_projection_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == set(EXPECTED_OPERATIONS), "operation projection matrix must cover save/read/list")
    for operation, required_scope in EXPECTED_OPERATIONS.items():
        row = rows[operation]
        require(row.get("required_scope") == required_scope, f"{operation} scope drifted")
        require(
            row.get("builder") == "BuildSavedWorkflowDraftRepositoryActorContextForOperation",
            f"{operation} builder drifted",
        )
        require(
            EXPECTED_PROJECTED_FIELDS.issubset(set(row.get("projected_actor_context_fields") or [])),
            f"{operation} projected actor context fields missing",
        )
        for field in (
            "fallback_to_dev_auth_allowed",
            "fallback_to_test_auth_allowed",
            "fallback_to_memory_dev_store_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")


def assert_failure_fallback_and_side_effects(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(mapping.get("status") == "runtime_bridge_fail_closed_implemented", "failure status drifted")
    require(
        EXPECTED_FAILURE_CODES.issubset(set(mapping.get("required_failure_codes") or [])),
        "failure mapping missing required codes",
    )
    expected_fields = {
        "dev_fake_auth_source_failure_code": "draft_auth_context_contract_mismatch",
        "unknown_operation_failure_code": "draft_auth_context_contract_mismatch",
        "missing_scope_failure_code": "draft_scope_grant_missing",
        "tenant_mismatch_failure_code": "draft_tenant_binding_missing",
        "missing_workspace_failure_code": "draft_workspace_membership_denied",
        "missing_application_failure_code": "draft_application_scope_denied",
        "missing_owner_failure_code": "draft_owner_scope_denied",
        "missing_audit_failure_code": "draft_audit_context_missing",
    }
    for field, expected in expected_fields.items():
        require(mapping.get(field) == expected, f"{field} drifted")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "dev_fake_auth_header_allowed_in_runtime_bridge",
        "test_auth_context_allowed_in_runtime_bridge",
        "missing_verified_auth_fallback_allowed",
        "missing_workspace_binding_fallback_allowed",
        "scope_denied_fallback_to_admin_allowed",
        "tenant_binding_failure_fallback_allowed",
        "repository_mode_fallback_to_memory_dev_store_allowed",
        "repository_mode_fallback_to_sample_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
    require(side_effects.get("status") == "pure_runtime_bridge", "side effect status drifted")
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(side_effects.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effects.get("forbidden_side_effects") or [])),
        "missing forbidden side effects",
    )


def assert_go_sources(source_path: str, test_path: str, fixture: dict[str, Any]) -> None:
    source = read(source_path)
    tests = read(test_path)
    failure_source = read("services/platform/internal/httpapi/workflow_saved_draft.go")
    for needle in (
        "type SavedWorkflowDraftVerifiedAuthContext struct",
        "type SavedWorkflowDraftWorkspaceBinding struct",
        "type SavedWorkflowDraftProductionAuthRuntimeResult struct",
        "func BuildSavedWorkflowDraftRepositoryActorContextForOperation(",
        "func BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth(",
        'savedWorkflowDraftProductionAuthSourceRadishOIDC = "radish_oidc_verified_context"',
        "SavedWorkflowDraftRepositoryActorContext{",
        "normalizedStringSet(auth.ScopeGrants)",
        "controlPlaneReadHasScope(auth.ScopeGrants, requiredScope)",
    ):
        require(needle in source, f"runtime bridge source missing {needle}")
    for operation, scope in EXPECTED_OPERATIONS.items():
        require(operation in source, f"runtime bridge missing operation {operation}")
        require(scope in source, f"runtime bridge missing scope {scope}")
    for failure_code in EXPECTED_FAILURE_CODES:
        require(failure_code in failure_source, f"saved draft failure constant missing {failure_code}")
    for test_name in fixture.get("required_tests") or []:
        require(str(test_name) in tests, f"runtime bridge test missing {test_name}")
    for needle in (
        "dev_fake_auth_header",
        "SavedWorkflowDraftFailureAuthContextMismatch",
        "SavedWorkflowDraftFailureIdentityContextMissing",
        "SavedWorkflowDraftFailureTenantBindingMissing",
        "SavedWorkflowDraftFailureWorkspaceMembershipDenied",
        "SavedWorkflowDraftFailureApplicationScopeDenied",
        "SavedWorkflowDraftFailureOwnerScopeDenied",
        "SavedWorkflowDraftFailureScopeGrantMissing",
        "SavedWorkflowDraftFailureAuditContextMissing",
    ):
        require(needle in tests, f"runtime bridge tests missing {needle}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")
    for source_path in guard.get("source_files_to_scan") or []:
        source = read(str(source_path))
        for literal in guard.get("forbidden_source_literals") or []:
            require(str(literal) not in source, f"{source_path} contains forbidden literal: {literal}")


def assert_docs_testing_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        require(path, "doc reference path missing")
        content = read(path)
        for needle in reference.get("must_contain") or []:
            require(str(needle) in content, f"{path} missing {needle}")

    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "runtime_bridge_checker_and_go_tests", "testing strategy drifted")
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "go test ./internal/httpapi",
        "workflow_saved_draft_production_auth_runtime_checker",
        "workflow_saved_draft_production_auth_readiness_checker",
        "workflow_saved_draft_adapter_smoke_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_call_oidc",
        "does_not_validate_token",
        "does_not_create_auth_middleware",
        "does_not_create_membership_adapter",
        "does_not_enable_repository_store_mode",
        "does_not_create_sql_or_migration",
        "does_not_connect_database",
        "does_not_create_production_api",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "adapter smoke checker missing from check-repo")
    require(current_checker in check_repo, "production auth runtime checker missing from check-repo")
    require(next_checker in check_repo, "product surface recheck missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "production auth runtime checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    source_path, test_path = assert_runtime_boundary(fixture)
    assert_operation_scope_projection(fixture)
    assert_failure_fallback_and_side_effects(fixture)
    assert_go_sources(source_path, test_path, fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft production auth runtime v1 checks passed.")


if __name__ == "__main__":
    main()
