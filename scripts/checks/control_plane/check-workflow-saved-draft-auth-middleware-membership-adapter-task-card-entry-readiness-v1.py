#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

try:
    import jsonschema
except ModuleNotFoundError:  # pragma: no cover - local bootstrap should install this.
    jsonschema = None


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.json"
)
SCHEMA_PATH = REPO_ROOT / "contracts/radish-oidc-token-validation.schema.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-token-validation-schema-artifact-implementation-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-artifact-implementation-v1.json",
        "draft_token_validation_schema_artifact_implemented",
    ),
    "workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.json",
        "draft_token_validation_auth_middleware_runtime_entry_review_defined",
    ),
    "radish-oidc-token-membership-upstream-evidence-refresh-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json",
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ),
    "workflow-saved-draft-production-auth-runtime-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json",
        "draft_production_auth_runtime_bridge_implemented",
    ),
    "workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.json",
        "draft_database_secret_resolver_runtime_dependency_refresh_defined",
    ),
    "workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.json",
        "draft_schema_marker_runtime_dependency_refresh_defined",
    ),
    "workflow-saved-draft-repository-mode-runtime-boundary-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-runtime-boundary-review-v1.json",
        "draft_repository_mode_runtime_boundary_review_defined",
    ),
}
EXPECTED_REQUIRED_FIELDS = {
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
EXPECTED_FORBIDDEN_FIELDS = {
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
    "provider_error_detail",
    "secret_value",
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
    "scope_grant_projection",
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
EXPECTED_ZERO_COUNTERS = {
    "issuer_network_call_count",
    "jwks_fetch_count",
    "token_validation_call_count",
    "auth_middleware_invocation_count",
    "membership_query_count",
    "membership_cache_create_count",
    "repository_query_count",
    "repository_write_count",
    "repository_mode_enablement_count",
    "database_connection_count",
    "sql_execution_count",
    "schema_marker_read_count",
    "schema_marker_write_count",
    "secret_resolver_call_count",
    "production_api_call_count",
    "executor_call_count",
    "confirmation_call_count",
    "business_writeback_count",
    "replay_call_count",
}
EXPECTED_REQUIRED_CHECKS = {
    "workflow_saved_draft_auth_middleware_membership_adapter_task_card_entry_readiness_checker",
    "workflow_saved_draft_token_validation_schema_artifact_implementation_checker",
    "workflow_saved_draft_token_validation_auth_middleware_runtime_entry_review_checker",
    "radish_oidc_token_membership_upstream_evidence_refresh_checker",
    "workflow_saved_draft_production_auth_runtime_checker",
    "git diff --check",
    "./scripts/check-repo.sh --fast",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a JSON object")
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


def forbidden_fields_from_schema(schema: dict[str, Any]) -> set[str]:
    forbidden: set[str] = set()
    for guard in schema.get("allOf") or []:
        if not isinstance(guard, dict):
            continue
        required = ((guard.get("not") or {}).get("required") or [])
        if len(required) == 1:
            forbidden.add(str(required[0]))
    return forbidden


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind")
        == "workflow_saved_draft_auth_middleware_membership_adapter_task_card_entry_readiness_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id")
        == "workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_auth_middleware_membership_adapter_task_card_entry_readiness_defined",
        "slice status drifted",
    )
    require(
        slice_info.get("entry_decision")
        == "auth_middleware_membership_adapter_implementation_task_card_blocked_pending_negative_auth_smoke_readiness",
        "entry decision drifted",
    )
    for key in ("task_card", "feature_topic"):
        relative_path = str(slice_info.get(key) or "")
        require(relative_path and (REPO_ROOT / relative_path).exists(), f"{key} missing: {relative_path}")
    for claim in (
        "auth_middleware_created",
        "membership_adapter_created",
        "negative_auth_smoke_runtime_created",
        "repository_mode_runtime_created",
        "database_runtime_created",
        "production_api_consumer_created",
    ):
        require(claim in set(slice_info.get("does_not_claim") or []), f"missing forbidden claim: {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        require(item.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        dependency = load_json(REPO_ROOT / relative_path)
        dependency_slice = dependency.get("slice") or {}
        require(dependency_slice.get("id") == dependency_id, f"{dependency_id} fixture id drifted")
        require(dependency_slice.get("status") == expected_status, f"{dependency_id} fixture status drifted")


def assert_schema_artifact() -> None:
    require(jsonschema is not None, "jsonschema is required; run scripts/bootstrap-dev.sh first")
    schema = load_json(SCHEMA_PATH)
    jsonschema.Draft202012Validator.check_schema(schema)
    require(schema.get("additionalProperties") is False, "schema must remain additionalProperties=false")
    require(set(schema.get("required") or []) == EXPECTED_REQUIRED_FIELDS, "schema required fields drifted")
    require(set((schema.get("properties") or {}).keys()) == EXPECTED_REQUIRED_FIELDS, "schema properties drifted")
    require(forbidden_fields_from_schema(schema) == EXPECTED_FORBIDDEN_FIELDS, "schema forbidden guards drifted")


def assert_readiness_decision(fixture: dict[str, Any]) -> None:
    decision = fixture.get("readiness_decision") or {}
    require(
        decision.get("status") == "auth_middleware_membership_adapter_task_card_entry_readiness_defined",
        "readiness status drifted",
    )
    require(
        decision.get("decision")
        == "auth_middleware_membership_adapter_implementation_task_card_blocked_pending_negative_auth_smoke_readiness",
        "readiness decision drifted",
    )
    for field in (
        "schema_artifact_consumed",
        "upstream_evidence_consumed",
        "production_auth_runtime_bridge_consumed",
        "auth_middleware_owner_contract_ready",
        "membership_adapter_static_contract_ready",
        "repository_actor_context_handoff_contract_ready",
        "negative_auth_smoke_runtime_readiness_required",
    ):
        require(decision.get(field) is True, f"{field} must be true")
    for field in (
        "negative_auth_smoke_runtime_readiness_complete",
        "implementation_task_card_allowed_now",
        "implementation_task_card_created_in_this_slice",
        "auth_middleware_created_in_this_slice",
        "token_validator_created_in_this_slice",
        "membership_adapter_created_in_this_slice",
        "negative_auth_smoke_runtime_created_in_this_slice",
        "repository_mode_enabled_in_this_slice",
        "database_runtime_created_in_this_slice",
        "production_api_consumer_created_in_this_slice",
    ):
        require(decision.get(field) is False, f"{field} must remain false")


def assert_future_requirements_and_blockers(fixture: dict[str, Any]) -> None:
    requirements = set(fixture.get("future_task_card_requirements") or [])
    for requirement in {
        "validator_output_must_match_radish_oidc_token_validation_schema",
        "auth_middleware_must_consume_verified_context_only",
        "membership_adapter_must_project_tenant_workspace_application_owner_scope",
        "negative_auth_smoke_readiness_must_exist_before_runtime_task_card",
        "no_oidc_network_membership_query_database_sql_or_production_api_side_effects",
    }:
        require(requirement in requirements, f"missing future task card requirement: {requirement}")
    blockers = rows_by_id(fixture, "blocked_prerequisites", "id")
    require("negative_auth_smoke_runtime_readiness" in blockers, "negative auth smoke blocker missing")
    for blocker_id, blocker in blockers.items():
        require(blocker.get("blocks_task_card") is True, f"{blocker_id} must block task card")


def assert_auth_and_membership_contracts(fixture: dict[str, Any]) -> None:
    owners = rows_by_id(fixture, "auth_middleware_owner_matrix", "owner_id")
    require(set(owners) == EXPECTED_AUTH_OWNERS, "auth owner ids drifted")
    for owner_id, owner in owners.items():
        require(owner.get("status") == "static_owner_ready_for_future_task_card", f"{owner_id} status drifted")
        require(owner.get("runtime_created_now") is False, f"{owner_id} runtime must remain absent")

    contracts = rows_by_id(fixture, "membership_adapter_contract_matrix", "contract_id")
    require(set(contracts) == EXPECTED_MEMBERSHIP_CONTRACTS, "membership contract ids drifted")
    for contract_id, contract in contracts.items():
        require(str(contract.get("status") or "").startswith("static_"), f"{contract_id} status drifted")
        require(contract.get("runtime_created_now") is False, f"{contract_id} runtime must remain absent")


def assert_handoff_and_negative_cases(fixture: dict[str, Any]) -> None:
    handoff = fixture.get("repository_actor_context_handoff") or {}
    require(handoff.get("status") == "existing_bridge_reused_no_new_runtime_handoff", "handoff status drifted")
    require(handoff.get("source_runtime") == "workflow-saved-draft-production-auth-runtime-v1", "handoff source drifted")
    require(
        handoff.get("source_function") == "BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth",
        "handoff function drifted",
    )
    require(handoff.get("requires_schema_artifact") == "contracts/radish-oidc-token-validation.schema.json", "schema handoff drifted")
    require(handoff.get("new_runtime_handoff_created_in_this_slice") is False, "new handoff runtime must be absent")
    for field in ("request_id", "tenant_ref", "workspace_id", "application_id", "actor_subject_ref", "scope_grants"):
        require(field in set(handoff.get("projected_fields") or []), f"handoff missing {field}")

    source = read("services/platform/internal/httpapi/workflow_saved_draft_production_auth_runtime.go")
    for literal in (
        "BuildSavedWorkflowDraftRepositoryActorContextFromProductionAuth",
        "SavedWorkflowDraftVerifiedAuthContext",
        "SavedWorkflowDraftWorkspaceBinding",
        "SavedWorkflowDraftRepositoryActorContext",
    ):
        require(literal in source, f"production auth runtime bridge missing {literal}")

    cases = rows_by_id(fixture, "negative_auth_smoke_readiness_dependency", "case_id")
    require(set(cases) == set(EXPECTED_NEGATIVE_CASES), "negative auth cases drifted")
    for case_id, expected_code in EXPECTED_NEGATIVE_CASES.items():
        require(cases[case_id].get("expected_failure_code") == expected_code, f"{case_id} failure drifted")

    upstream = load_json(REPO_ROOT / "scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json")
    upstream_cases = rows_by_id(upstream, "negative_auth_smoke_cases", "case_id")
    require(set(upstream_cases) == set(EXPECTED_NEGATIVE_CASES), "upstream negative auth matrix drifted")


def assert_failure_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "auth_middleware_membership_adapter_task_card_entry_readiness_fail_closed_required",
        "failure mapping status drifted",
    )
    require(set(mapping.get("required_failure_codes") or []) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for field in (
        "must_fail_before_repository_query",
        "must_not_return_raw_token",
        "must_not_return_raw_claim",
        "must_not_return_membership_record",
        "must_not_fallback_to_dev_auth",
    ):
        require(mapping.get(field) is True, f"{field} must be true")

    source = read("services/platform/internal/httpapi/workflow_saved_draft.go")
    for failure_code in EXPECTED_FAILURE_CODES:
        require(failure_code in source, f"saved draft failure code missing: {failure_code}")

    fallback = fixture.get("no_fallback_policy") or []
    for literal in ("no fallback to dev fake auth", "no fallback to memory dev store", "no fallback to fixture as auth truth"):
        require(literal in fallback, f"missing fallback policy: {literal}")

    counters = fixture.get("side_effect_counters") or {}
    require(set(counters) == EXPECTED_ZERO_COUNTERS, "side effect counters drifted")
    for counter, value in counters.items():
        require(value == 0, f"{counter} must remain zero")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("required_static_artifacts_exist") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required artifact missing: {relative_path}")
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
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in content]
        require(not missing, f"{path} missing literals: {missing}")

    strategy = fixture.get("testing_strategy") or {}
    require(
        strategy.get("status") == "auth_middleware_membership_adapter_task_card_entry_readiness_checker_only_no_runtime",
        "testing strategy status drifted",
    )
    require(EXPECTED_REQUIRED_CHECKS.issubset(set(strategy.get("required_checks") or [])), "required checks drifted")
    for field in (
        "does_not_start_service",
        "does_not_fetch_issuer_discovery",
        "does_not_fetch_jwks",
        "does_not_validate_token",
        "does_not_create_auth_middleware",
        "does_not_query_membership",
        "does_not_create_membership_adapter",
        "does_not_enable_repository_store_mode",
        "does_not_connect_database",
        "does_not_run_sql",
        "does_not_create_production_api",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py"
    current_checker = (
        "checks/control_plane/"
        "check-workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    for checker in (previous_checker, current_checker, next_checker):
        require(checker in check_repo, f"{checker} missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "check-repo ordering drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_schema_artifact()
    assert_readiness_decision(fixture)
    assert_future_requirements_and_blockers(fixture)
    assert_auth_and_membership_contracts(fixture)
    assert_handoff_and_negative_cases(fixture)
    assert_failure_and_safety(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft auth middleware / membership adapter task card entry readiness checks passed.")


if __name__ == "__main__":
    main()
