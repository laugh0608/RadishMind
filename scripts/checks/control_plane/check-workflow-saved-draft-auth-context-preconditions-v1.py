#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from workflow_saved_draft_selector_implementation_guard import (
    selector_implementation_file_allowed,
    selector_implementation_literal_allowed,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-auth-context-preconditions-v1.json"
)
DURABLE_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-durable-store-preconditions-v1.json"
)
REPOSITORY_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-preconditions-v1.json"
)
SCHEMA_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-migration-preconditions-v1.json"
)
RADISH_OIDC_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "radish_oidc_ready",
    "auth_middleware_ready",
    "token_validation_ready",
    "session_cookie_ready",
    "workspace_membership_adapter_ready",
    "production_auth_ready",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "database_schema_ready",
    "database_migration_ready",
    "store_selector_implemented",
    "production_api_consumer_ready",
    "saved_draft_list_implemented",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_CONTEXT_FIELDS = {
    "request_id",
    "tenant_ref",
    "workspace_id",
    "application_id",
    "actor_subject_ref",
    "owner_subject_ref",
    "scope_grants",
    "audit_ref",
}
EXPECTED_FORBIDDEN_CONTEXT_SOURCES = {
    "http_request",
    "query_string_tenant_override",
    "query_string_workspace_override",
    "query_string_application_override",
    "draft_payload_scope_override",
    "raw_authorization_header",
    "cookie_value",
    "global_current_user",
    "sample_fixture_fallback",
    "memory_dev_store_fallback",
    "provider_runtime",
    "workflow_executor",
}
EXPECTED_SCOPE_MATRIX = {
    "SaveWorkflowDraftRecord": ("workflow_drafts:write", "owner_or_workspace_write_grant"),
    "ReadWorkflowDraftRecord": ("workflow_drafts:read", "owner_or_workspace_read_grant"),
    "ListWorkflowDraftRecords": ("workflow_drafts:read", "owner_or_workspace_list_grant"),
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
    "draft_scope_denied",
}
EXPECTED_ALLOWED_AUDIT_FIELDS = {
    "request_id",
    "audit_ref",
    "tenant_ref",
    "workspace_id",
    "application_id",
    "actor_subject_ref",
    "owner_subject_ref",
    "scope_grant_ids",
    "failure_code",
}
EXPECTED_FORBIDDEN_AUDIT_FIELDS = {
    "id_token",
    "access_token",
    "refresh_token",
    "authorization_header",
    "cookie_value",
    "client_secret",
    "raw_claim_dump",
    "raw_permission_payload",
    "raw_oidc_response",
    "raw_jwks_response",
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
    expected = {
        "workflow-saved-draft-durable-store-preconditions-v1": (
            DURABLE_PRECONDITIONS_FIXTURE_PATH,
            "draft_durable_store_preconditions_defined",
        ),
        "workflow-saved-draft-repository-contract-preconditions-v1": (
            REPOSITORY_PRECONDITIONS_FIXTURE_PATH,
            "draft_repository_contract_preconditions_defined",
        ),
        "workflow-saved-draft-schema-migration-preconditions-v1": (
            SCHEMA_PRECONDITIONS_FIXTURE_PATH,
            "draft_schema_migration_preconditions_defined",
        ),
        "radish-oidc-client-preconditions": (
            RADISH_OIDC_PRECONDITIONS_FIXTURE_PATH,
            "governance_boundary_satisfied",
        ),
    }
    missing = sorted(set(expected) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for expected_id, (path, expected_status) in expected.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == expected_id, f"dependency {expected_id} id drifted")
        require(
            slice_info.get("status") == expected_status,
            f"dependency {expected_id} status must be {expected_status}",
        )


def assert_radish_oidc_claim_mapping_dependency() -> None:
    oidc = load_json(RADISH_OIDC_PRECONDITIONS_FIXTURE_PATH)
    preconditions = {
        str(item.get("id") or ""): item
        for item in oidc.get("preconditions") or []
        if isinstance(item, dict)
    }
    require("claim-mapping" in preconditions, "Radish OIDC preconditions must define claim mapping")
    require("tenant-binding" in preconditions, "Radish OIDC preconditions must define tenant binding")
    claims = set(preconditions["claim-mapping"].get("required_claims") or [])
    require({"sub", "tenant_id", "roles", "permissions"}.issubset(claims), "OIDC claim mapping drifted")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "workflow_saved_draft_auth_context_preconditions_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-auth-context-preconditions-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_auth_context_preconditions_defined",
        "auth context preconditions status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_auth_context_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("auth_context_boundary") or {}
    repository = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH).get("repository_contract_boundary") or {}
    require(
        boundary.get("status") == "auth_context_preconditions_defined_not_implemented",
        "auth context boundary status drifted",
    )
    require(
        boundary.get("future_store_source") == repository.get("future_store_source"),
        "future store source must match repository preconditions",
    )
    require(boundary.get("context_preconditions_defined_in_this_slice") is True, "context must be defined")
    for field in (
        "oidc_middleware_created_in_this_slice",
        "token_validation_created_in_this_slice",
        "session_cookie_created_in_this_slice",
        "workspace_membership_adapter_created_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_created_in_this_slice",
        "store_selector_created_in_this_slice",
        "database_query_allowed_in_this_slice",
        "production_api_consumer_created_in_this_slice",
        "saved_draft_list_created_in_this_slice",
        "publish_or_run_created_in_this_slice",
        "executor_created_in_this_slice",
        "writeback_or_replay_created_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_actor_context_mapping(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("repository_actor_context_mapping") or {}
    repository_actor = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH).get("actor_context_contract") or {}
    require(
        mapping.get("status") == "actor_context_mapping_defined_not_implemented",
        "actor context mapping status drifted",
    )
    require(
        mapping.get("context_type") == repository_actor.get("context_type"),
        "actor context type must match repository preconditions",
    )
    missing_fields = sorted(EXPECTED_CONTEXT_FIELDS - set(mapping.get("required_context_fields") or []))
    require(not missing_fields, f"missing context fields: {missing_fields}")
    missing_repository_fields = sorted(
        set(repository_actor.get("required_context_fields") or []) - set(mapping.get("required_context_fields") or [])
    )
    require(not missing_repository_fields, f"missing repository actor fields: {missing_repository_fields}")
    field_sources = mapping.get("field_sources") or {}
    for field in EXPECTED_CONTEXT_FIELDS:
        require(str(field_sources.get(field) or "").strip(), f"missing source for {field}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CONTEXT_SOURCES - set(mapping.get("must_not_depend_on") or []))
    require(not missing_forbidden, f"missing forbidden context sources: {missing_forbidden}")


def assert_workspace_membership_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("workspace_membership_policy") or {}
    require(
        policy.get("status") == "membership_contract_required_before_adapter",
        "membership policy status drifted",
    )
    for field in (
        "tenant_binding_required",
        "workspace_membership_required",
        "application_membership_required",
        "owner_subject_ref_required",
        "tenant_workspace_application_must_match",
        "team_owner_policy_deferred",
        "sharing_transfer_policy_deferred",
    ):
        require(policy.get(field) is True, f"{field} must remain true")
    require(policy.get("cross_workspace_owner_grant_allowed") is False, "cross workspace owner grants must be false")
    require(
        policy.get("draft_id_scope_must_remain") == ["tenant_ref", "workspace_id", "application_id", "draft_id"],
        "draft id scope drifted",
    )


def assert_scope_grant_matrix(fixture: dict[str, Any]) -> None:
    repository_operations = {
        str(operation.get("method_name") or "")
        for operation in load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH).get("repository_operations") or []
        if isinstance(operation, dict)
    }
    require(repository_operations == set(EXPECTED_SCOPE_MATRIX), "repository operations drifted")
    matrix = {
        str(row.get("operation") or ""): row
        for row in fixture.get("scope_grant_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == set(EXPECTED_SCOPE_MATRIX), "scope grant operation ids drifted")
    for operation, (expected_scope, expected_owner_policy) in EXPECTED_SCOPE_MATRIX.items():
        row = matrix[operation]
        require(row.get("required_scope") == expected_scope, f"{operation} scope drifted")
        require(row.get("owner_policy") == expected_owner_policy, f"{operation} owner policy drifted")
        require(row.get("failure_code_when_missing") == "draft_scope_grant_missing", f"{operation} failure drifted")
        for field in (
            "fallback_to_sample_allowed",
            "fallback_to_memory_dev_store_allowed",
            "side_effect_allowed_on_denied",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")


def assert_failure_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("failure_policy") or {}
    require(policy.get("status") == "fail_closed_required", "failure policy status drifted")
    missing = sorted(EXPECTED_FAILURE_CODES - set(policy.get("required_failure_codes") or []))
    require(not missing, f"missing failure codes: {missing}")
    for field in (
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_memory_dev_store_allowed",
        "return_draft_body_on_denied",
        "write_allowed_on_denied",
        "publish_or_run_allowed_on_denied",
    ):
        require(policy.get(field) is False, f"{field} must remain false")


def assert_audit_sanitization_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("audit_sanitization_policy") or {}
    require(policy.get("status") == "sanitized_reference_only", "audit policy status drifted")
    missing_allowed = sorted(EXPECTED_ALLOWED_AUDIT_FIELDS - set(policy.get("allowed_audit_fields") or []))
    require(not missing_allowed, f"missing allowed audit fields: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_AUDIT_FIELDS - set(policy.get("forbidden_audit_fields") or []))
    require(not missing_forbidden, f"missing forbidden audit fields: {missing_forbidden}")


def assert_dev_auth_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("dev_auth_boundary") or {}
    require(boundary.get("status") == "dev_only_opt_in_must_remain", "dev auth boundary status drifted")
    require(
        boundary.get("backend_dev_auth_env") == "RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1",
        "backend dev auth env drifted",
    )
    require(
        boundary.get("saved_draft_dev_http_env") == "RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1",
        "saved draft dev http env drifted",
    )
    require(
        boundary.get("saved_draft_dev_write_env") == "RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1",
        "saved draft dev write env drifted",
    )
    for field in (
        "fake_auth_allowed_in_production",
        "fake_auth_header_public_api_allowed",
        "dev_auth_may_override_tenant_from_query",
        "dev_auth_may_override_workspace_from_query",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_required_doc_references(fixture: dict[str, Any]) -> None:
    references = fixture.get("required_doc_references") or []
    require(references, "required doc references must be declared")
    for reference in references:
        path = str(reference.get("path") or "")
        require(path, "doc reference path missing")
        content = read(path)
        for needle in reference.get("must_contain") or []:
            require(str(needle) in content, f"{path} missing required text: {needle}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("implementation_artifact_guard") or {}
    require(guard.get("status") == "forbid_implementation_artifacts", "artifact guard status drifted")
    for relative_path in guard.get("future_files_must_not_exist") or []:
        if selector_implementation_file_allowed(REPO_ROOT, str(relative_path)):
            continue
        require(not (REPO_ROOT / str(relative_path)).exists(), f"future artifact exists early: {relative_path}")
    source_paths = guard.get("source_files_to_scan") or []
    literals = guard.get("future_literals_must_not_appear_in_source") or []
    require(source_paths, "source files to scan must be declared")
    require(literals, "future literals must be declared")
    for source_path in source_paths:
        source = read(str(source_path))
        for literal in literals:
            if selector_implementation_literal_allowed(REPO_ROOT, str(literal)):
                continue
            require(str(literal) not in source, f"{source_path} contains future literal: {literal}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "static_checker_only", "testing strategy status drifted")
    required = set(strategy.get("required_checks") or [])
    require(
        "workflow_saved_draft_auth_context_preconditions_checker" in required,
        "missing checker in testing strategy",
    )
    require("./scripts/check-repo.sh --fast" in required, "missing fast baseline in testing strategy")
    for field in (
        "does_not_start_service",
        "does_not_call_oidc",
        "does_not_validate_tokens",
        "does_not_connect_database",
        "does_not_create_repository_adapter",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-schema-migration-preconditions-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-auth-context-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run auth context preconditions check")
    require(previous_checker in check_repo, "schema migration preconditions checker missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "auth context preconditions check must run after schema migration preconditions",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_radish_oidc_claim_mapping_dependency()
    assert_slice(fixture)
    assert_auth_context_boundary(fixture)
    assert_actor_context_mapping(fixture)
    assert_workspace_membership_policy(fixture)
    assert_scope_grant_matrix(fixture)
    assert_failure_policy(fixture)
    assert_audit_sanitization_policy(fixture)
    assert_dev_auth_boundary(fixture)
    assert_required_doc_references(fixture)
    assert_artifact_guard(fixture)
    assert_testing_strategy(fixture)
    assert_check_repo_registration()
    print("workflow saved draft auth context preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
