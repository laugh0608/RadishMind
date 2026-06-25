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
    / "scripts/checks/fixtures/workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1.json"
)
SCHEMA_PATH = REPO_ROOT / "contracts/radish-oidc-token-validation.schema.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.json",
        "draft_auth_middleware_membership_adapter_task_card_entry_readiness_defined",
    ),
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
EXPECTED_NEGATIVE_CASES = {
    "missing_authorization_header": (
        "request_auth_boundary",
        "draft_identity_context_missing",
    ),
    "malformed_bearer_header": (
        "request_auth_boundary",
        "draft_auth_context_contract_mismatch",
    ),
    "unknown_issuer": (
        "issuer_boundary",
        "draft_auth_context_contract_mismatch",
    ),
    "jwks_unavailable": (
        "jwks_boundary",
        "draft_auth_context_contract_mismatch",
    ),
    "invalid_signature": (
        "token_signature_boundary",
        "draft_auth_context_contract_mismatch",
    ),
    "expired_token": (
        "token_time_boundary",
        "draft_auth_context_contract_mismatch",
    ),
    "missing_required_claim": (
        "token_claim_boundary",
        "draft_identity_context_missing",
    ),
    "tenant_binding_denied": (
        "tenant_binding_boundary",
        "draft_tenant_binding_missing",
    ),
    "workspace_membership_denied": (
        "workspace_membership_boundary",
        "draft_workspace_membership_denied",
    ),
    "application_scope_denied": (
        "application_scope_boundary",
        "draft_application_scope_denied",
    ),
    "owner_scope_denied": (
        "owner_scope_boundary",
        "draft_owner_scope_denied",
    ),
    "scope_grant_missing": (
        "scope_grant_boundary",
        "draft_scope_grant_missing",
    ),
    "membership_source_unavailable": (
        "membership_source_boundary",
        "draft_auth_context_contract_mismatch",
    ),
}
EXPECTED_FAILURE_CODES = {
    "draft_auth_context_contract_mismatch",
    "draft_identity_context_missing",
    "draft_tenant_binding_missing",
    "draft_workspace_membership_denied",
    "draft_application_scope_denied",
    "draft_owner_scope_denied",
    "draft_scope_grant_missing",
}
EXPECTED_DIAGNOSTIC_FORBIDDEN_FIELDS = {
    "raw_token",
    "authorization_header",
    "cookie",
    "client_secret",
    "refresh_token",
    "authorization_code",
    "raw_claim_dump",
    "jwks_raw_dump",
    "membership_raw_record",
    "provider_error_detail",
    "database_detail",
    "secret_value",
    "dsn",
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
    "workflow_saved_draft_negative_auth_smoke_runtime_readiness_checker",
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
        fixture.get("kind") == "workflow_saved_draft_negative_auth_smoke_runtime_readiness_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_negative_auth_smoke_runtime_readiness_defined",
        "slice status drifted",
    )
    require(
        slice_info.get("entry_decision") == "negative_auth_smoke_runtime_readiness_defined_before_runtime_artifact",
        "entry decision drifted",
    )
    for key in ("task_card", "feature_topic"):
        relative_path = str(slice_info.get(key) or "")
        require(relative_path and (REPO_ROOT / relative_path).exists(), f"{key} missing: {relative_path}")
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    for claim in (
        "negative_auth_smoke_runtime_fixture_created",
        "negative_auth_smoke_runtime_checker_created",
        "auth_middleware_implementation_task_card_created",
        "token_validator_created",
        "auth_middleware_created",
        "membership_adapter_created",
        "repository_mode_runtime_created",
        "database_runtime_created",
        "production_api_consumer_created",
        "workflow_executor_ready",
    ):
        require(claim in forbidden_claims, f"missing forbidden claim: {claim}")


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


def assert_readiness_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("readiness_boundary") or {}
    expected_values = {
        "status": "negative_auth_smoke_runtime_readiness_defined_no_runtime",
        "decision": "negative_auth_smoke_runtime_readiness_defined_before_runtime_artifact",
        "future_runtime_fixture": "scripts/checks/fixtures/workflow-saved-draft-negative-auth-smoke-runtime-v1.json",
        "future_runtime_checker": (
            "scripts/checks/control_plane/check-workflow-saved-draft-negative-auth-smoke-runtime-v1.py"
        ),
    }
    for field, expected in expected_values.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "synthetic_sanitized_cases_only",
        "must_fail_before_repository_query",
    ):
        require(boundary.get(field) is True, f"{field} must be true")
    for field in (
        "runtime_smoke_artifact_allowed_now",
        "runtime_fixture_created_in_this_slice",
        "runtime_checker_created_in_this_slice",
        "runtime_runner_created_in_this_slice",
        "auth_middleware_task_card_allowed_now",
        "auth_middleware_task_card_created_in_this_slice",
        "auth_middleware_created_in_this_slice",
        "token_validator_created_in_this_slice",
        "membership_adapter_created_in_this_slice",
        "repository_mode_enabled_in_this_slice",
        "database_runtime_created_in_this_slice",
        "production_api_consumer_created_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_negative_case_matrix(fixture: dict[str, Any]) -> None:
    cases = rows_by_id(fixture, "negative_auth_case_matrix", "case_id")
    require(set(cases) == set(EXPECTED_NEGATIVE_CASES), "negative auth case ids drifted")
    for case_id, (expected_boundary, expected_code) in EXPECTED_NEGATIVE_CASES.items():
        case = cases[case_id]
        require(case.get("failure_boundary") == expected_boundary, f"{case_id} boundary drifted")
        require(case.get("expected_failure_code") == expected_code, f"{case_id} failure code drifted")
        require(case.get("runtime_fixture_created_now") is False, f"{case_id} runtime fixture must be absent")
        require(case.get("must_fail_before_repository_query") is True, f"{case_id} must fail before query")
        require(case.get("future_input_shape"), f"{case_id} future input shape missing")

    upstream = load_json(REPO_ROOT / "scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json")
    upstream_cases = rows_by_id(upstream, "negative_auth_smoke_cases", "case_id")
    require(set(upstream_cases) == set(EXPECTED_NEGATIVE_CASES), "upstream negative auth ids drifted")
    for case_id, (expected_boundary, _expected_code) in EXPECTED_NEGATIVE_CASES.items():
        require(upstream_cases[case_id].get("failure_boundary") == expected_boundary, f"{case_id} upstream boundary drifted")

    runtime_review = load_json(
        REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.json"
    )
    review_cases = rows_by_id(runtime_review, "negative_auth_smoke_cases", "case_id")
    require(set(review_cases) == set(EXPECTED_NEGATIVE_CASES), "runtime entry review negative auth ids drifted")
    for case_id, (expected_boundary, expected_code) in EXPECTED_NEGATIVE_CASES.items():
        require(review_cases[case_id].get("failure_boundary") == expected_boundary, f"{case_id} review boundary drifted")
        require(review_cases[case_id].get("expected_failure_code") == expected_code, f"{case_id} review code drifted")

    entry = load_json(
        REPO_ROOT
        / "scripts/checks/fixtures/workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.json"
    )
    dependency_cases = rows_by_id(entry, "negative_auth_smoke_readiness_dependency", "case_id")
    require(set(dependency_cases) == set(EXPECTED_NEGATIVE_CASES), "entry readiness negative auth ids drifted")
    for case_id, (_expected_boundary, expected_code) in EXPECTED_NEGATIVE_CASES.items():
        require(dependency_cases[case_id].get("expected_failure_code") == expected_code, f"{case_id} entry code drifted")


def assert_future_requirements_and_failure_mapping(fixture: dict[str, Any]) -> None:
    requirements = set(fixture.get("future_runtime_smoke_requirements") or [])
    for requirement in {
        "future_fixture_must_use_synthetic_sanitized_cases_only",
        "future_checker_must_not_fetch_issuer_discovery_or_jwks",
        "future_checker_must_not_query_membership_source",
        "future_checker_must_not_connect_database_or_run_sql",
        "future_smoke_must_validate_failure_mapping_before_repository_query",
        "future_smoke_must_keep_repository_mode_disabled_until_separate_enablement",
    }:
        require(requirement in requirements, f"missing future runtime requirement: {requirement}")

    mapping = rows_by_id(fixture, "failure_mapping", "code")
    require(set(mapping) == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for code, item in mapping.items():
        require(item.get("failure_boundary"), f"{code} failure boundary missing")
        require(item.get("sanitized_diagnostic"), f"{code} sanitized diagnostic missing")

    source = read("services/platform/internal/httpapi/workflow_saved_draft.go")
    for failure_code in EXPECTED_FAILURE_CODES:
        require(failure_code in source, f"saved draft failure code missing: {failure_code}")


def assert_diagnostics_and_side_effects(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    require(
        EXPECTED_DIAGNOSTIC_FORBIDDEN_FIELDS.issubset(set(diagnostics.get("forbidden_fields") or [])),
        "diagnostic forbidden fields drifted",
    )
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must be false")
    for field in ("case_id", "failure_code", "failure_boundary", "request_id", "audit_ref", "policy_version"):
        require(field in set(diagnostics.get("allowed_fields") or []), f"diagnostics missing {field}")

    fallback = set(fixture.get("no_fallback_policy") or [])
    for literal in (
        "no fallback to dev fake auth",
        "no fallback to memory dev store",
        "no fallback to sample",
        "no fallback to fixture as auth truth",
        "no fallback to test-only fake resolver",
        "no fallback to fake query executor",
    ):
        require(literal in fallback, f"missing fallback policy: {literal}")

    side_effect_policy = set(fixture.get("no_side_effect_policy") or [])
    for literal in (
        "no issuer discovery fetch",
        "no JWKS fetch",
        "no real token validation",
        "no membership query",
        "no database connection",
        "no SQL execution",
        "no production API call",
    ):
        require(literal in side_effect_policy, f"missing side effect policy: {literal}")

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
        strategy.get("status") == "negative_auth_smoke_runtime_readiness_checker_only_no_runtime",
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
        "does_not_create_executor_confirmation_writeback_or_replay",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = (
        "checks/control_plane/"
        "check-workflow-saved-draft-auth-middleware-membership-adapter-task-card-entry-readiness-v1.py"
    )
    current_checker = (
        "checks/control_plane/check-workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1.py"
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
    assert_readiness_boundary(fixture)
    assert_negative_case_matrix(fixture)
    assert_future_requirements_and_failure_mapping(fixture)
    assert_diagnostics_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft negative auth smoke runtime readiness checks passed.")


if __name__ == "__main__":
    main()
