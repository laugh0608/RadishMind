#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "radish-oidc-token-membership-readiness-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-readiness-v1.json",
        "radish_oidc_token_membership_readiness_defined",
    ),
    "radish-oidc-token-membership-implementation-entry-review-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-implementation-entry-review-v1.json",
        "radish_oidc_token_membership_implementation_entry_review_defined",
    ),
    "radish-oidc-token-membership-upstream-evidence-refresh-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json",
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ),
    "workflow-saved-draft-production-auth-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-readiness-v1.json",
        "draft_production_auth_readiness_defined",
    ),
    "workflow-saved-draft-production-auth-runtime-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json",
        "draft_production_auth_runtime_bridge_implemented",
    ),
    "workflow-saved-draft-repository-mode-runtime-boundary-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-runtime-boundary-review-v1.json",
        "draft_repository_mode_runtime_boundary_review_defined",
    ),
    "workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2.json",
        "draft_database_connection_provider_implementation_entry_refresh_v2_defined",
    ),
    "workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.json",
        "draft_schema_marker_runtime_dependency_refresh_defined",
    ),
    "workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.json",
        "draft_database_secret_resolver_runtime_dependency_refresh_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "token_validation_schema_created",
    "token_validator_created",
    "oidc_middleware_ready",
    "auth_middleware_created",
    "membership_adapter_created",
    "membership_cache_created",
    "negative_auth_smoke_runtime_created",
    "repository_actor_context_handoff_runtime_created",
    "repository_store_mode_enabled",
    "repository_mode_runtime_created",
    "database_runtime_created",
    "database_connection_ready",
    "schema_marker_runtime_created",
    "database_secret_resolver_runtime_created",
    "production_api_consumer_created",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_BOUNDARY_FIELDS = {
    "status": "token_validation_auth_middleware_runtime_entry_review_defined_no_task_card",
    "decision": "token_validation_auth_middleware_runtime_still_blocked_before_implementation_task_card",
    "current_development_mode": "entry_review_only_no_auth_runtime",
    "selected_implementation_track_now": "none",
    "future_token_validation_schema": "contracts/radish-oidc-token-validation.schema.json",
    "future_auth_middleware_file": "services/platform/internal/httpapi/radish_oidc_auth_middleware.go",
    "future_token_validator_file": "services/platform/internal/httpapi/radish_oidc_token_validator.go",
    "future_membership_adapter_file": "services/platform/internal/httpapi/radish_membership_adapter.go",
    "future_negative_auth_smoke_fixture": (
        "scripts/checks/fixtures/workflow-saved-draft-negative-auth-smoke-runtime-v1.json"
    ),
}
EXPECTED_TRUE_BOUNDARY_FLAGS = {
    "radish_oidc_upstream_evidence_consumed",
    "production_auth_runtime_bridge_consumed",
    "repository_mode_boundary_consumed",
    "schema_marker_runtime_dependency_refresh_consumed",
    "database_secret_resolver_runtime_dependency_refresh_consumed",
    "negative_auth_smoke_matrix_consumed",
}
EXPECTED_FALSE_BOUNDARY_FLAGS = {
    "implementation_task_card_created_in_this_slice",
    "token_validation_schema_created_in_this_slice",
    "token_validator_created_in_this_slice",
    "auth_middleware_created_in_this_slice",
    "membership_adapter_created_in_this_slice",
    "membership_cache_created_in_this_slice",
    "negative_auth_smoke_runtime_created_in_this_slice",
    "repository_actor_context_handoff_runtime_created_in_this_slice",
    "repository_mode_enabled_in_this_slice",
    "database_runtime_created_in_this_slice",
    "production_api_consumer_created_in_this_slice",
    "executor_created_in_this_slice",
    "confirmation_created_in_this_slice",
    "writeback_or_replay_created_in_this_slice",
}
EXPECTED_REVIEW_STATUS = {
    "token_validation_schema": "blocked_schema_contract_reviewed_no_schema_file",
    "auth_middleware_ownership": "blocked_owner_reviewed_no_middleware",
    "membership_adapter": "blocked_contract_reviewed_no_adapter",
    "negative_auth_smoke": "blocked_matrix_reviewed_no_runtime_smoke",
    "repository_actor_context_handoff": "static_handoff_reviewed_existing_bridge_only",
    "failure_mapping": "fail_closed_mapping_reviewed_no_runtime",
    "no_fallback": "no_fallback_reviewed_required",
    "no_side_effects": "no_side_effects_reviewed_required",
    "artifact_guard": "artifact_guard_reviewed_required",
}
EXPECTED_SCHEMA_FIELDS = {
    "issuer_ref",
    "subject_ref",
    "tenant_ref",
    "audience_refs",
    "scope_grants",
    "workspace_binding_refs",
    "application_scope_refs",
    "owner_subject_ref",
    "key_id_ref",
    "algorithm",
    "issued_at",
    "expires_at",
    "auth_time",
    "policy_version",
    "request_id",
    "audit_ref",
}
EXPECTED_FORBIDDEN_SCHEMA_FIELDS = {
    "raw_token",
    "authorization_header",
    "cookie",
    "client_secret",
    "refresh_token",
    "authorization_code",
    "jwks_raw_dump",
    "raw_claim_dump",
    "membership_raw_record",
    "database_detail",
}
EXPECTED_AUTH_OWNERS = {
    "route_binding_owner",
    "token_validator_owner",
    "failure_envelope_owner",
    "audit_context_owner",
    "dev_fake_auth_isolation_owner",
}
EXPECTED_MEMBERSHIP_CONTRACTS = {
    "tenant_binding",
    "workspace_membership",
    "application_scope",
    "owner_scope",
    "membership_cache_policy",
}
EXPECTED_NEGATIVE_CASES = {
    "missing_authorization_header": "draft_identity_context_missing",
    "malformed_bearer_header": "draft_auth_context_contract_mismatch",
    "unknown_issuer": "draft_auth_context_contract_mismatch",
    "jwks_unavailable": "draft_auth_context_contract_mismatch",
    "invalid_signature": "draft_auth_context_contract_mismatch",
    "expired_token": "draft_auth_context_contract_mismatch",
    "missing_required_claim": "draft_identity_context_missing",
    "tenant_binding_denied": "draft_tenant_binding_missing",
    "workspace_membership_denied": "draft_workspace_membership_denied",
    "application_scope_denied": "draft_application_scope_denied",
    "owner_scope_denied": "draft_owner_scope_denied",
    "scope_grant_missing": "draft_scope_grant_missing",
    "membership_source_unavailable": "draft_auth_context_contract_mismatch",
}
EXPECTED_FAILURE_CODES = {
    "draft_auth_context_contract_mismatch",
    "draft_identity_context_missing",
    "draft_tenant_binding_missing",
    "draft_workspace_membership_denied",
    "draft_application_scope_denied",
    "draft_owner_scope_denied",
    "draft_scope_grant_missing",
    "draft_audit_context_missing",
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_store_unavailable",
    "draft_store_migration_unavailable",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "issuer_network_call_count=0",
    "jwks_fetch_count=0",
    "token_validation_call_count=0",
    "auth_middleware_invocation_count=0",
    "membership_query_count=0",
    "membership_cache_create_count=0",
    "repository_query_count=0",
    "repository_write_count=0",
    "repository_mode_enablement_count=0",
    "database_connection_count=0",
    "driver_open_count=0",
    "sql_execution_count=0",
    "schema_marker_read_count=0",
    "schema_marker_write_count=0",
    "secret_resolver_call_count=0",
    "production_api_call_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_TEST_FLAGS = {
    "does_not_start_service",
    "does_not_call_oidc",
    "does_not_fetch_issuer_discovery",
    "does_not_fetch_jwks",
    "does_not_validate_token",
    "does_not_create_auth_middleware",
    "does_not_query_membership",
    "does_not_create_membership_adapter",
    "does_not_create_membership_cache",
    "does_not_enable_repository_store_mode",
    "does_not_connect_database",
    "does_not_import_database_driver",
    "does_not_run_sql",
    "does_not_read_schema_marker",
    "does_not_write_schema_marker",
    "does_not_resolve_secret",
    "does_not_create_production_api",
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


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "workflow_saved_draft_token_validation_auth_middleware_runtime_entry_review_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_token_validation_auth_middleware_runtime_entry_review_defined",
        "status drifted",
    )
    require(
        slice_info.get("entry_decision")
        == "token_validation_auth_middleware_runtime_still_blocked_before_implementation_task_card",
        "entry decision drifted",
    )
    for key in ("task_card", "feature_topic"):
        relative_path = str(slice_info.get(key) or "")
        require(relative_path, f"{key} is required")
        require((REPO_ROOT / relative_path).exists(), f"{key} path missing: {relative_path}")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        require(item.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        evidence = load_json(REPO_ROOT / relative_path)
        slice_info = evidence.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"{dependency_id} fixture id drifted")
        require(slice_info.get("status") == expected_status, f"{dependency_id} fixture status drifted")

    upstream = load_json(
        REPO_ROOT / "scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json"
    )
    upstream_cases = rows_by_id(upstream, "negative_auth_smoke_cases", "case_id")
    require(
        set(upstream_cases) == set(EXPECTED_NEGATIVE_CASES),
        "upstream negative auth smoke matrix must remain aligned",
    )


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_review_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY_FIELDS.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in EXPECTED_TRUE_BOUNDARY_FLAGS:
        require(boundary.get(field) is True, f"{field} must be true")
    for field in EXPECTED_FALSE_BOUNDARY_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_review_matrix(fixture: dict[str, Any]) -> None:
    reviews = rows_by_id(fixture, "entry_review_matrix", "review_id")
    require(set(reviews) == set(EXPECTED_REVIEW_STATUS), "review ids drifted")
    for review_id, expected_status in EXPECTED_REVIEW_STATUS.items():
        item = reviews[review_id]
        require(item.get("status") == expected_status, f"{review_id} status drifted")
        require(item.get("runtime_artifact_allowed_now") is False, f"{review_id} runtime must remain forbidden")
        require(item.get("must_cover"), f"{review_id} must describe coverage")
        require(item.get("blocked_by"), f"{review_id} must describe blockers")


def assert_token_schema_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("token_validation_schema_contract") or {}
    require(
        contract.get("status") == "static_schema_contract_reviewed_no_schema_file",
        "token schema contract status drifted",
    )
    schema_path = str(contract.get("schema_artifact_path") or "")
    require(schema_path == "contracts/radish-oidc-token-validation.schema.json", "schema path drifted")
    require(not (REPO_ROOT / schema_path).exists(), "token validation schema must not be created yet")
    missing_fields = sorted(EXPECTED_SCHEMA_FIELDS - set(contract.get("required_fields") or []))
    require(not missing_fields, f"missing token schema fields: {missing_fields}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_SCHEMA_FIELDS - set(contract.get("forbidden_fields") or []))
    require(not missing_forbidden, f"missing forbidden token schema fields: {missing_forbidden}")
    overlap = sorted(set(contract.get("required_fields") or []) & set(contract.get("forbidden_fields") or []))
    require(not overlap, f"required token schema fields include forbidden fields: {overlap}")


def assert_ownership_and_membership(fixture: dict[str, Any]) -> None:
    owners = rows_by_id(fixture, "auth_middleware_ownership", "owner_id")
    require(set(owners) == EXPECTED_AUTH_OWNERS, "auth middleware owners drifted")
    for owner_id, item in owners.items():
        require(item.get("status") == "static_owner_defined_no_middleware", f"{owner_id} status drifted")
        require(item.get("runtime_created_now") is False, f"{owner_id} runtime must remain absent")

    contracts = rows_by_id(fixture, "membership_adapter_contract", "contract_id")
    require(set(contracts) == EXPECTED_MEMBERSHIP_CONTRACTS, "membership contract ids drifted")
    for contract_id, item in contracts.items():
        require(item.get("runtime_created_now") is False, f"{contract_id} runtime must remain absent")
        require(
            str(item.get("status") or "").startswith("static_"),
            f"{contract_id} must remain a static contract or policy",
        )


def assert_handoff_and_negative_smoke(fixture: dict[str, Any]) -> None:
    handoff = fixture.get("repository_actor_context_handoff") or {}
    require(handoff.get("status") == "existing_bridge_consumed_no_new_auth_runtime", "handoff status drifted")
    require(handoff.get("source_runtime") == "workflow-saved-draft-production-auth-runtime-v1", "source drifted")
    require(
        handoff.get("source_function") == "BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth",
        "source function drifted",
    )
    require(handoff.get("requires_verified_context") is True, "handoff must require verified context")
    require(
        handoff.get("current_runtime_handoff_created_in_this_slice") is False,
        "new runtime handoff must not be created",
    )
    for field in (
        "request_id",
        "tenant_ref",
        "workspace_id",
        "application_id",
        "actor_subject_ref",
        "owner_subject_ref",
        "scope_grants",
        "audit_ref",
    ):
        require(field in set(handoff.get("projected_fields") or []), f"handoff missing projected field {field}")

    source = read("services/platform/internal/httpapi/workflow_saved_draft_production_auth_runtime.go")
    for literal in (
        "BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth",
        "SavedWorkflowDraftVerifiedAuthContext",
        "SavedWorkflowDraftWorkspaceBinding",
        "SavedWorkflowDraftRepositoryActorContext",
    ):
        require(literal in source, f"production auth runtime bridge literal missing: {literal}")

    cases = rows_by_id(fixture, "negative_auth_smoke_cases", "case_id")
    require(set(cases) == set(EXPECTED_NEGATIVE_CASES), "negative auth smoke cases drifted")
    for case_id, expected_code in EXPECTED_NEGATIVE_CASES.items():
        item = cases[case_id]
        require(item.get("expected_failure_code") == expected_code, f"{case_id} failure code drifted")
        require(str(item.get("failure_boundary") or "").strip(), f"{case_id} failure boundary is required")


def assert_failure_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "token_validation_auth_middleware_runtime_entry_review_fail_closed_required",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    expected_mappings = {
        "token_schema_missing_failure_code": "draft_auth_context_contract_mismatch",
        "token_validator_missing_failure_code": "draft_auth_context_contract_mismatch",
        "auth_middleware_missing_failure_code": "draft_auth_context_contract_mismatch",
        "subject_missing_failure_code": "draft_identity_context_missing",
        "tenant_binding_denied_failure_code": "draft_tenant_binding_missing",
        "workspace_membership_denied_failure_code": "draft_workspace_membership_denied",
        "application_scope_denied_failure_code": "draft_application_scope_denied",
        "owner_scope_denied_failure_code": "draft_owner_scope_denied",
        "scope_grant_missing_failure_code": "draft_scope_grant_missing",
        "audit_context_missing_failure_code": "draft_audit_context_missing",
        "membership_source_unavailable_failure_code": "draft_auth_context_contract_mismatch",
        "repository_mode_still_disabled_failure_code": "repository_store_disabled",
        "unknown_store_mode_failure_code": "invalid_draft_store_mode",
        "database_runtime_missing_failure_code": "draft_store_unavailable",
        "schema_marker_runtime_missing_failure_code": "draft_store_migration_unavailable",
    }
    for field, expected in expected_mappings.items():
        require(mapping.get(field) == expected, f"{field} drifted")
    for field in (
        "must_fail_before_repository_query",
        "must_not_return_raw_token",
        "must_not_return_raw_claim",
        "must_not_return_membership_record",
        "must_not_fallback_to_dev_auth",
        "static_review_must_not_imply_runtime_ready",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

    source = read("services/platform/internal/httpapi/workflow_saved_draft.go")
    for failure_code in EXPECTED_FAILURE_CODES:
        require(failure_code in source, f"saved draft failure code missing from source: {failure_code}")

    fallback = fixture.get("no_fallback_policy") or {}
    for field, value in fallback.items():
        require(value is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
    missing_counters = sorted(
        EXPECTED_SIDE_EFFECT_COUNTERS - set(side_effects.get("side_effect_counters_must_remain") or [])
    )
    require(not missing_counters, f"missing side effect counters: {missing_counters}")
    require(side_effects.get("forbidden_side_effects"), "forbidden side effects must be listed")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("required_static_artifacts_exist") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required static artifact missing: {relative_path}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")
    for root in guard.get("forbid_sql_files_under") or []:
        root_path = REPO_ROOT / str(root)
        sql_files = sorted(root_path.rglob("*.sql")) if root_path.exists() else []
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
        strategy.get("status") == "token_validation_auth_middleware_runtime_entry_review_checker_only_no_runtime",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_token_validation_auth_middleware_runtime_entry_review_checker",
        "radish_oidc_token_membership_upstream_evidence_refresh_checker",
        "workflow_saved_draft_production_auth_runtime_checker",
        "workflow_saved_draft_database_secret_resolver_runtime_dependency_refresh_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in EXPECTED_TEST_FLAGS:
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = (
        "checks/control_plane/check-workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.py"
    )
    current_checker = (
        "checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    for checker in (previous_checker, current_checker, next_checker):
        require(checker in check_repo, f"{checker} missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "token validation auth middleware entry review checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_boundary(fixture)
    assert_review_matrix(fixture)
    assert_token_schema_contract(fixture)
    assert_ownership_and_membership(fixture)
    assert_handoff_and_negative_smoke(fixture)
    assert_failure_and_safety(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft token validation auth middleware runtime entry review checks passed.")


if __name__ == "__main__":
    main()
