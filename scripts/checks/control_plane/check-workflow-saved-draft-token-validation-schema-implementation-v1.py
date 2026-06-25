#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-implementation-v1.json"
)
PRIOR_READINESS_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-task-card-readiness-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-token-validation-schema-task-card-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-task-card-readiness-v1.json",
        "draft_token_validation_schema_task_card_readiness_defined",
    ),
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
    "token_validation_schema_runtime_ready",
    "json_schema_validation_runtime_ready",
    "schema_positive_fixture_created",
    "schema_negative_fixture_created",
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
    "prior_schema_task_card_readiness_gate": "satisfied_ready_for_task_card",
    "schema_task_card_creation_gate": "satisfied_static_task_card_created",
    "schema_artifact_scope_gate": "defined_for_future_schema_artifact",
    "verified_token_output_field_gate": "defined_for_future_schema_artifact",
    "forbidden_field_rejection_gate": "defined_for_future_schema_validation",
    "json_schema_validation_gate": "defined_for_future_schema_validation",
    "sanitized_failure_envelope_gate": "defined_for_future_schema_validation",
    "artifact_guard_gate": "satisfied_for_task_card",
    "schema_artifact_creation_gate": "not_opened",
    "token_validator_runtime_gate": "blocked",
    "auth_middleware_runtime_gate": "blocked",
    "membership_adapter_runtime_gate": "blocked",
    "negative_auth_smoke_runtime_gate": "blocked",
    "repository_mode_runtime_gate": "blocked",
}
EXPECTED_SCHEMA_SCOPE = {
    "token validation schema artifact",
    "json schema version",
    "verified token output shape",
    "required field allowlist",
    "forbidden raw-material fields",
    "additionalProperties false",
    "positive verified token context fixture",
    "missing required field negative fixture",
    "forbidden raw-material field negative fixture",
    "additional properties negative fixture",
    "sanitized failure envelope alignment",
    "consumer handoff to production auth runtime bridge",
    "artifact guard",
    "offline fast baseline",
}
EXPECTED_REQUIRED_VALIDATION = {
    "token validation schema implementation task card checker",
    "token validation schema task card readiness checker",
    "json schema validation fixture",
    "positive verified token context fixture",
    "missing required field negative fixture",
    "forbidden raw-material field negative fixture",
    "additional properties negative fixture",
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
EXPECTED_FAILURE_CODES = {
    "token_validation_schema_implementation_task_card_missing",
    "token_validation_schema_implementation_scope_missing",
    "token_validation_schema_implementation_field_boundary_missing",
    "token_validation_schema_implementation_forbidden_field_missing",
    "token_validation_schema_implementation_validation_plan_missing",
    "token_validation_schema_implementation_schema_file_created_in_task_card",
    "token_validation_schema_implementation_runtime_scope_overreach",
    "token_validation_schema_implementation_dev_auth_fallback",
}
EXPECTED_DIAGNOSTIC_FIELDS = {
    "token_validation_schema_implementation_task_card_status",
    "schema_artifact_status",
    "schema_validation_status",
    "verified_token_output_contract_status",
    "failure_code",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}
EXPECTED_NO_FALLBACK = {
    "no fallback to dev fake auth",
    "no fallback to memory dev store",
    "no fallback to sample",
    "no fallback to fixture",
    "no fallback to test-only fake resolver",
    "no fallback to fake query executor",
    "no fallback to developer env plaintext",
    "schema implementation task card does not mean token validation schema ready",
    "schema implementation task card does not mean token validation ready",
    "schema implementation task card does not mean auth middleware ready",
    "schema implementation task card does not mean repository mode ready",
}
EXPECTED_NO_SIDE_EFFECTS = {
    "checker reads committed docs and fixtures only",
    "no issuer discovery fetch",
    "no JWKS fetch",
    "no token validation call",
    "no schema file write",
    "no schema validation runtime call",
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
    "does_not_run_schema_validation_runtime",
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
        fixture.get("kind") == "workflow_saved_draft_token_validation_schema_implementation_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-token-validation-schema-implementation-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_token_validation_schema_implementation_task_card_defined",
        "status drifted",
    )
    require(slice_info.get("task_card_status") == "created_static_task_card", "task card status drifted")
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


def assert_prior_readiness_alignment() -> None:
    prior = load_json(PRIOR_READINESS_FIXTURE_PATH)
    boundary = prior.get("readiness_boundary") or {}
    require(
        boundary.get("task_card_decision") == "token_validation_schema_task_card_ready_for_next_task",
        "prior task card decision drifted",
    )
    require(
        boundary.get("future_schema_implementation_task_card")
        == "docs/task-cards/workflow-saved-draft-token-validation-schema-implementation-v1-plan.md",
        "prior future schema implementation task card path drifted",
    )
    schema = prior.get("schema_field_boundary") or {}
    require(set(schema.get("required_fields") or []) == EXPECTED_SCHEMA_FIELDS, "prior required fields drifted")
    require(
        set(schema.get("forbidden_fields") or []) == EXPECTED_FORBIDDEN_SCHEMA_FIELDS,
        "prior forbidden fields drifted",
    )
    guard = prior.get("artifact_guard") or {}
    no_longer_forbidden = "docs/task-cards/workflow-saved-draft-token-validation-schema-implementation-v1-plan.md"
    require(
        no_longer_forbidden not in set(guard.get("files_must_not_exist") or []),
        "prior readiness guard still forbids current task card",
    )


def assert_task_card_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("task_card_boundary") or {}
    expected_values = {
        "status": "token_validation_schema_implementation_task_card_defined_no_schema_file",
        "task_card_status": "created_static_task_card",
        "current_development_mode": "schema_implementation_task_card_only_no_schema_artifact",
        "future_schema_artifact_path": "contracts/radish-oidc-token-validation.schema.json",
    }
    for field, expected in expected_values.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "prior_schema_task_card_readiness_consumed",
        "prior_runtime_entry_review_consumed",
        "upstream_oidc_evidence_consumed",
        "production_auth_readiness_consumed",
        "production_auth_runtime_bridge_consumed",
    ):
        require(boundary.get(field) is True, f"{field} must be true")
    for field in (
        "token_validation_schema_created_in_this_slice",
        "schema_positive_fixture_created_in_this_slice",
        "schema_negative_fixture_created_in_this_slice",
        "json_schema_validation_runtime_created_in_this_slice",
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


def assert_gates_and_schema_requirements(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "implementation_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    requirements = fixture.get("schema_implementation_requirements") or {}
    require(
        requirements.get("schema_artifact_path") == "contracts/radish-oidc-token-validation.schema.json",
        "schema artifact path drifted",
    )
    missing_scope = sorted(EXPECTED_SCHEMA_SCOPE - set(requirements.get("required_scope") or []))
    require(not missing_scope, f"missing schema implementation scope: {missing_scope}")
    require(set(requirements.get("required_fields") or []) == EXPECTED_SCHEMA_FIELDS, "required fields drifted")
    require(
        set(requirements.get("forbidden_fields") or []) == EXPECTED_FORBIDDEN_SCHEMA_FIELDS,
        "forbidden fields drifted",
    )
    missing_validation = sorted(EXPECTED_REQUIRED_VALIDATION - set(requirements.get("required_validation") or []))
    require(not missing_validation, f"missing schema implementation validation: {missing_validation}")
    missing_forbidden = sorted(EXPECTED_MUST_NOT_INCLUDE - set(requirements.get("must_not_include") or []))
    require(not missing_forbidden, f"missing forbidden scope: {missing_forbidden}")
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
    missing_allowed = sorted(EXPECTED_DIAGNOSTIC_FIELDS - set(diagnostics.get("allowed_fields") or []))
    require(not missing_allowed, f"missing diagnostic allowed fields: {missing_allowed}")
    forbidden_fields = set(diagnostics.get("forbidden_fields") or [])
    for field in EXPECTED_FORBIDDEN_SCHEMA_FIELDS | {"dsn"}:
        require(field in forbidden_fields, f"diagnostics missing forbidden field {field}")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must stay false")


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
        strategy.get("status") == "token_validation_schema_implementation_task_card_checker_only_no_schema_artifact",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected in (
        "workflow_saved_draft_token_validation_schema_implementation_checker",
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
        "checks/control_plane/check-workflow-saved-draft-token-validation-schema-task-card-readiness-v1.py"
    )
    current_checker = "checks/control_plane/check-workflow-saved-draft-token-validation-schema-implementation-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    for checker in (previous_checker, current_checker, next_checker):
        require(checker in check_repo, f"{checker} missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "token validation schema implementation checker order drifted",
    )


def assert_no_secret_literals() -> None:
    paths = [
        "docs/features/workflow/saved-workflow-draft-token-validation-schema-implementation-v1.md",
        "docs/task-cards/workflow-saved-draft-token-validation-schema-implementation-v1-plan.md",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-implementation-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"token validation schema implementation artifacts contain forbidden literals: {found}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_prior_readiness_alignment()
    assert_task_card_boundary(fixture)
    assert_gates_and_schema_requirements(fixture)
    assert_failure_mapping_and_diagnostics(fixture)
    assert_policies_and_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    assert_no_secret_literals()
    print("workflow saved draft token validation schema implementation task card checks passed.")


if __name__ == "__main__":
    main()
