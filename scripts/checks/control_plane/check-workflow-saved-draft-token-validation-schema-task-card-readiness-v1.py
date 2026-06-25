#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-task-card-readiness-v1.json"
)
PRIOR_ENTRY_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.json",
        "draft_token_validation_auth_middleware_runtime_entry_review_defined",
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
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "token_validation_schema_created",
    "token_validation_schema_implemented",
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
EXPECTED_GATE_STATUS = {
    "prior_runtime_entry_review_gate": "satisfied_static_entry_review",
    "upstream_issuer_evidence_gate": "satisfied_reviewed_evidence",
    "verified_token_output_contract_gate": "satisfied_for_schema_task_card",
    "schema_implementation_task_card_gate": "ready_for_next_task",
    "schema_artifact_gate": "not_created",
    "token_validator_runtime_gate": "not_opened",
    "auth_middleware_runtime_gate": "blocked",
    "membership_adapter_runtime_gate": "blocked",
    "negative_auth_smoke_runtime_gate": "blocked",
    "repository_mode_runtime_gate": "blocked",
}
EXPECTED_REQUIRED_SCOPE = {
    "token validation schema artifact scope",
    "json schema version",
    "verified token output shape",
    "required field allowlist",
    "forbidden raw-material fields",
    "sanitized failure envelope alignment",
    "consumer handoff to production auth runtime bridge",
    "artifact guard",
    "offline fast baseline",
}
EXPECTED_REQUIRED_VALIDATION = {
    "token validation schema implementation checker",
    "json schema validation fixture",
    "forbidden field negative fixture",
    "token validation auth middleware runtime entry review checker",
    "radish oidc upstream evidence refresh checker",
    "production auth runtime bridge checker",
    "fast repository check",
}
EXPECTED_MUST_NOT_INCLUDE = {
    "OIDC discovery runtime",
    "JWKS fetch runtime",
    "token signature validation runtime",
    "auth middleware route binding",
    "login logout session cookie",
    "membership adapter",
    "membership cache",
    "negative auth smoke runtime",
    "repository mode runtime",
    "database connection",
    "SQL",
    "schema marker runtime",
    "secret resolver runtime",
    "production API",
    "executor",
    "confirmation",
    "writeback",
    "replay",
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
    "provider_error_detail",
    "secret_value",
}
EXPECTED_PRIOR_FORBIDDEN_SCHEMA_FIELDS = EXPECTED_FORBIDDEN_SCHEMA_FIELDS - {
    "provider_error_detail",
    "secret_value",
}
EXPECTED_FAILURE_CODES = {
    "token_validation_schema_task_card_contract_missing",
    "token_validation_schema_task_card_forbidden_field_missing",
    "token_validation_schema_task_card_runtime_scope_overreach",
    "token_validation_schema_file_created_in_readiness",
    "token_validation_runtime_artifact_created_in_readiness",
    "token_validation_schema_task_card_dev_auth_fallback",
}
EXPECTED_NO_FALLBACK = {
    "no fallback to dev fake auth",
    "no fallback to memory dev store",
    "no fallback to sample",
    "no fallback to fixture",
    "no fallback to test-only fake resolver",
    "no fallback to fake query executor",
    "no fallback to developer env plaintext",
    "task card readiness does not mean token validation ready",
    "task card readiness does not mean auth middleware ready",
    "task card readiness does not mean repository mode ready",
}
EXPECTED_NO_SIDE_EFFECTS = {
    "checker reads committed docs and fixtures only",
    "no issuer discovery fetch",
    "no JWKS fetch",
    "no token validation call",
    "no schema file write",
    "no auth middleware invocation",
    "no membership query",
    "no membership cache creation",
    "no repository mode enablement",
    "no database connection",
    "no SQL execution",
    "no schema marker read",
    "no schema marker write",
    "no secret resolver call",
    "no production API call",
    "no executor call",
    "no confirmation call",
    "no business writeback",
    "no replay call",
}
EXPECTED_TEST_FLAGS = {
    "does_not_start_service",
    "does_not_fetch_issuer_discovery",
    "does_not_fetch_jwks",
    "does_not_validate_token",
    "does_not_create_schema_file",
    "does_not_create_auth_middleware",
    "does_not_query_membership",
    "does_not_enable_repository_store_mode",
    "does_not_connect_database",
    "does_not_run_sql",
    "does_not_create_production_api",
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


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "workflow_saved_draft_token_validation_schema_task_card_readiness_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-token-validation-schema-task-card-readiness-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_token_validation_schema_task_card_readiness_defined",
        "status drifted",
    )
    require(
        slice_info.get("task_card_decision") == "token_validation_schema_task_card_ready_for_next_task",
        "task card decision drifted",
    )
    for key in ("task_card", "feature_topic"):
        relative_path = str(slice_info.get(key) or "")
        require(relative_path, f"{key} is required")
        require((REPO_ROOT / relative_path).exists(), f"{key} path missing: {relative_path}")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        require(item.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        evidence = load_json(REPO_ROOT / relative_path)
        require((evidence.get("slice") or {}).get("id") == dependency_id, f"{dependency_id} fixture id drifted")
        require(
            (evidence.get("slice") or {}).get("status") == expected_status,
            f"{dependency_id} fixture status drifted",
        )

    prior = load_json(PRIOR_ENTRY_FIXTURE_PATH)
    prior_contract = prior.get("token_validation_schema_contract") or {}
    require(
        set(prior_contract.get("required_fields") or []) == EXPECTED_SCHEMA_FIELDS,
        "prior token schema required fields drifted",
    )
    missing_prior_forbidden = EXPECTED_PRIOR_FORBIDDEN_SCHEMA_FIELDS - set(
        prior_contract.get("forbidden_fields") or []
    )
    require(not missing_prior_forbidden, f"prior token schema forbidden fields missing: {sorted(missing_prior_forbidden)}")


def assert_readiness_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("readiness_boundary") or {}
    expected_values = {
        "status": "token_validation_schema_task_card_readiness_defined_no_schema_file",
        "task_card_decision": "token_validation_schema_task_card_ready_for_next_task",
        "current_development_mode": "schema_task_card_readiness_only_no_auth_runtime",
        "future_schema_artifact_path": "contracts/radish-oidc-token-validation.schema.json",
        "future_schema_implementation_task_card": (
            "docs/task-cards/workflow-saved-draft-token-validation-schema-implementation-v1-plan.md"
        ),
    }
    for field, expected in expected_values.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "prior_runtime_entry_review_consumed",
        "upstream_oidc_evidence_consumed",
        "production_auth_readiness_consumed",
        "production_auth_runtime_bridge_consumed",
    ):
        require(boundary.get(field) is True, f"{field} must be true")
    for field in (
        "schema_implementation_task_card_created_in_this_slice",
        "token_validation_schema_created_in_this_slice",
        "token_validator_created_in_this_slice",
        "auth_middleware_created_in_this_slice",
        "membership_adapter_created_in_this_slice",
        "negative_auth_smoke_runtime_created_in_this_slice",
        "repository_mode_enabled_in_this_slice",
        "database_runtime_created_in_this_slice",
        "production_api_consumer_created_in_this_slice",
        "executor_created_in_this_slice",
        "confirmation_created_in_this_slice",
        "writeback_or_replay_created_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gates_and_future_task_card(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "readiness_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    future = fixture.get("future_task_card_requirements") or {}
    missing_scope = sorted(EXPECTED_REQUIRED_SCOPE - set(future.get("required_scope") or []))
    require(not missing_scope, f"missing future task card scope: {missing_scope}")
    missing_validation = sorted(EXPECTED_REQUIRED_VALIDATION - set(future.get("required_validation") or []))
    require(not missing_validation, f"missing future task card validation: {missing_validation}")
    missing_forbidden = sorted(EXPECTED_MUST_NOT_INCLUDE - set(future.get("must_not_include") or []))
    require(not missing_forbidden, f"missing future task card forbidden scope: {missing_forbidden}")


def assert_schema_field_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("schema_field_boundary") or {}
    require(
        boundary.get("status") == "schema_task_card_fields_ready_no_schema_file",
        "schema field boundary status drifted",
    )
    require(set(boundary.get("required_fields") or []) == EXPECTED_SCHEMA_FIELDS, "required schema fields drifted")
    require(
        set(boundary.get("forbidden_fields") or []) == EXPECTED_FORBIDDEN_SCHEMA_FIELDS,
        "forbidden schema fields drifted",
    )
    overlap = EXPECTED_SCHEMA_FIELDS & EXPECTED_FORBIDDEN_SCHEMA_FIELDS
    require(not overlap, f"schema required fields overlap forbidden fields: {sorted(overlap)}")


def assert_failure_mapping_and_diagnostics(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in failures.items():
        require(str(item.get("failure_boundary") or "").strip(), f"{code} failure boundary is required")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} sanitized diagnostic is required")
        for forbidden in ("raw token", "authorization header", "cookie", "secret", "dsn"):
            require(forbidden not in diagnostic.lower(), f"{code} diagnostic exposes {forbidden}")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must stay false")
    forbidden_fields = set(diagnostics.get("forbidden_fields") or [])
    for field in ("raw_token", "authorization_header", "cookie", "raw_claim_dump", "secret_value", "dsn"):
        require(field in forbidden_fields, f"diagnostics missing forbidden field {field}")


def assert_policies_and_artifact_guard(fixture: dict[str, Any]) -> None:
    missing_fallback = sorted(EXPECTED_NO_FALLBACK - set(fixture.get("no_fallback_policy") or []))
    require(not missing_fallback, f"missing no fallback policy: {missing_fallback}")
    missing_side_effects = sorted(EXPECTED_NO_SIDE_EFFECTS - set(fixture.get("no_side_effect_policy") or []))
    require(not missing_side_effects, f"missing no side effect policy: {missing_side_effects}")
    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters are required")
    for counter, value in counters.items():
        require(value == 0, f"{counter} must stay zero")

    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("required_static_artifacts_exist") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required static artifact missing: {relative_path}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")
    for root in guard.get("forbid_sql_files_under") or []:
        root_path = REPO_ROOT / str(root)
        sql_files = sorted(root_path.rglob("*.sql")) if root_path.exists() else []
        require(not sql_files, f"SQL files must not exist under {root}: {sql_files}")


def assert_docs_testing_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        content = read(path)
        for literal in reference.get("must_contain") or []:
            require(str(literal) in content, f"{path} missing {literal}")

    strategy = fixture.get("testing_strategy") or {}
    require(
        strategy.get("status") == "token_validation_schema_task_card_readiness_checker_only_no_schema_runtime",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected in (
        "workflow_saved_draft_token_validation_schema_task_card_readiness_checker",
        "workflow_saved_draft_token_validation_auth_middleware_runtime_entry_review_checker",
        "radish_oidc_token_membership_upstream_evidence_refresh_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected in required_checks, f"missing required check: {expected}")
    for field in EXPECTED_TEST_FLAGS:
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = (
        "checks/control_plane/check-workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.py"
    )
    current_checker = (
        "checks/control_plane/check-workflow-saved-draft-token-validation-schema-task-card-readiness-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    for checker in (previous_checker, current_checker, next_checker):
        require(checker in check_repo, f"{checker} missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "token validation schema task card readiness checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_readiness_boundary(fixture)
    assert_gates_and_future_task_card(fixture)
    assert_schema_field_boundary(fixture)
    assert_failure_mapping_and_diagnostics(fixture)
    assert_policies_and_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft token validation schema task card readiness checks passed.")


if __name__ == "__main__":
    main()
