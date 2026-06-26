#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-database-role-policy-readiness-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.json",
        "draft_database_driver_dsn_tls_policy_readiness_defined",
    ),
    "workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.json",
        "draft_database_connection_provider_implementation_entry_refresh_defined",
    ),
    "workflow-saved-draft-database-connection-schema-marker-preconditions-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-schema-marker-preconditions-v1.json",
        "draft_database_connection_schema_marker_preconditions_defined",
    ),
    "workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1.json",
        "draft_manual_migration_runner_implementation_entry_refresh_defined",
    ),
    "workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.json",
        "draft_schema_marker_contract_implementation_entry_review_defined",
    ),
    "workflow-saved-draft-repository-mode-runtime-boundary-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-runtime-boundary-review-v1.json",
        "draft_repository_mode_runtime_boundary_review_defined",
    ),
    "radish-oidc-token-membership-implementation-entry-review-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-implementation-entry-review-v1.json",
        "radish_oidc_token_membership_implementation_entry_review_defined",
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "database_connection_provider_implementation_task_card_created",
    "database_connection_provider_ready",
    "database_connection_ready",
    "database_role_ready",
    "role_policy_runtime_ready",
    "role_grant_ready",
    "secret_resolver_ready",
    "production_resolver_runtime_ready",
    "database_driver_selected",
    "database_driver_imported",
    "dsn_parser_created",
    "raw_dsn_available",
    "tls_runtime_ready",
    "connection_factory_implemented",
    "connection_smoke_ready",
    "database_query_executor_ready",
    "schema_marker_read_write_ready",
    "sql_migration_created",
    "migration_runner_implemented",
    "repository_store_mode_enabled",
    "repository_mode_runtime_created",
    "durable_persistence_ready",
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
EXPECTED_BOUNDARY_FIELDS = {
    "status": "database_role_policy_readiness_defined_no_runtime_policy",
    "decision": "role_policy_defined_for_future_task_only",
    "current_development_mode": "readiness_only_no_role_runtime_no_connection",
    "future_implementation_task_card": (
        "docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md"
    ),
    "future_runtime_dml_role_status": "policy_defined_runtime_not_created",
    "future_migration_ddl_marker_role_status": "policy_defined_runtime_not_created",
    "least_privilege_review_status": "static_review_boundary_defined",
    "future_environment_binding_status": "policy_defined_runtime_not_created",
    "future_role_claim_metadata_status": "metadata_only_shape_defined_no_auth_runtime",
    "future_cross_environment_denial_smoke_status": "blocked_until_connection_smoke_strategy",
    "future_connection_provider_dependency_status": "blocked_before_implementation_task_card",
}
EXPECTED_FALSE_BOUNDARY_FLAGS = {
    "implementation_task_card_created_in_this_slice",
    "database_connection_provider_file_created_in_this_slice",
    "role_policy_runtime_allowed_in_this_slice",
    "database_role_creation_allowed_in_this_slice",
    "role_grant_allowed_in_this_slice",
    "role_claim_parser_allowed_in_this_slice",
    "database_driver_import_allowed_in_this_slice",
    "go_mod_driver_dependency_allowed_in_this_slice",
    "dsn_parser_allowed_in_this_slice",
    "raw_dsn_material_allowed_in_this_slice",
    "tls_runtime_allowed_in_this_slice",
    "connection_factory_allowed_in_this_slice",
    "connection_smoke_allowed_in_this_slice",
    "secret_resolver_allowed_in_this_slice",
    "fake_resolver_invocation_allowed_in_this_slice",
    "production_resolver_invocation_allowed_in_this_slice",
    "credential_handle_creation_allowed_in_this_slice",
    "database_query_executor_allowed_in_this_slice",
    "schema_marker_allowed_in_this_slice",
    "sql_files_created_in_this_slice",
    "migration_runner_allowed_in_this_slice",
    "repository_mode_success_allowed_in_this_slice",
    "oidc_validation_allowed_in_this_slice",
    "membership_adapter_allowed_in_this_slice",
    "production_api_consumer_allowed_in_this_slice",
    "executor_allowed_in_this_slice",
    "writeback_or_replay_allowed_in_this_slice",
}
EXPECTED_GATE_STATUS = {
    "driver_dsn_tls_policy_consumed": "consumed_role_dependency",
    "runtime_dml_role_policy": "defined_no_runtime_role",
    "migration_ddl_marker_role_policy": "defined_no_runtime_role",
    "least_privilege_review_boundary": "defined_static_review_required",
    "environment_binding": "defined_no_runtime_binding",
    "role_claim_metadata_boundary": "defined_metadata_only_no_token_parser",
    "cross_environment_denial_smoke_precondition": "blocked_before_connection_smoke_strategy",
    "connection_provider_dependency": "blocked_before_implementation_task_card",
    "no_runtime_artifacts_leaked": "required_now",
}
EXPECTED_FAILURE_CODES = {
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_store_unavailable",
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_store_contract_mismatch",
    "draft_auth_context_contract_mismatch",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "secret_resolver_call_count=0",
    "fake_resolver_call_count=0",
    "production_resolver_call_count=0",
    "secret_value_read_count=0",
    "credential_handle_created_count=0",
    "cloud_secret_call_count=0",
    "network_call_count=0",
    "database_connection_count=0",
    "driver_import_count=0",
    "driver_open_count=0",
    "dsn_parse_count=0",
    "dsn_build_count=0",
    "tls_config_create_count=0",
    "role_policy_evaluation_count=0",
    "role_claim_parse_count=0",
    "role_grant_count=0",
    "connection_smoke_count=0",
    "sql_execution_count=0",
    "schema_marker_read_count=0",
    "schema_marker_write_count=0",
    "repository_mode_enablement_count=0",
    "repository_read_count=0",
    "repository_write_count=0",
    "oidc_call_count=0",
    "token_validation_count=0",
    "membership_lookup_count=0",
    "production_api_call_count=0",
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
    declared = {
        str(row.get("id") or "")
        for row in fixture.get("depends_on") or []
        if isinstance(row, dict)
    }
    missing = sorted(set(EXPECTED_DEPENDENCIES) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"{dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"{dependency_id} status drifted")


def assert_slice_and_boundary(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "workflow_saved_draft_database_role_policy_readiness_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-database-role-policy-readiness-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_database_role_policy_readiness_defined",
        "status drifted",
    )
    require(
        slice_info.get("task_card")
        == "docs/task-cards/workflow-saved-draft-database-role-policy-readiness-v1-plan.md",
        "task card drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    boundary = fixture.get("readiness_boundary") or {}
    for field, expected_value in EXPECTED_BOUNDARY_FIELDS.items():
        require(boundary.get(field) == expected_value, f"{field} drifted")
    for field in EXPECTED_FALSE_BOUNDARY_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_role_policy_matrices(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "readiness_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "readiness gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    runtime_role = fixture.get("runtime_dml_role_policy") or {}
    require(
        runtime_role.get("status") == "future_policy_only_no_runtime_role",
        "runtime role policy status drifted",
    )
    require(runtime_role.get("role_runtime_allowed_now") is False, "runtime role must not be created")
    require(runtime_role.get("database_role_created_now") is False, "database role must not be created")
    require(runtime_role.get("role_grant_allowed_now") is False, "role grant must remain forbidden")
    for forbidden in ("create_table", "schema_marker_write", "role_grant", "cross_environment_connection"):
        require(
            forbidden in (runtime_role.get("forbidden_future_operations") or []),
            f"runtime role missing forbidden operation: {forbidden}",
        )

    migration_role = fixture.get("migration_ddl_marker_role_policy") or {}
    require(
        migration_role.get("status") == "future_policy_only_no_runtime_role",
        "migration role policy status drifted",
    )
    require(
        migration_role.get("repository_adapter_consumption_allowed_now") is False,
        "migration role cannot be consumed by repository adapter",
    )
    require(migration_role.get("role_runtime_allowed_now") is False, "migration role runtime forbidden")
    for forbidden in ("repository_adapter_save", "saved_draft_request_path_consumption"):
        require(
            forbidden in (migration_role.get("forbidden_future_operations") or []),
            f"migration role missing forbidden operation: {forbidden}",
        )

    review = fixture.get("least_privilege_review") or {}
    require(review.get("status") == "static_review_boundary_defined", "least privilege status drifted")
    require(review.get("static_review_required_now") is True, "least privilege static review required")
    require(review.get("runtime_review_execution_allowed_now") is False, "runtime review execution forbidden")

    environment = fixture.get("environment_binding") or {}
    require(environment.get("mismatch_failure_code") == "draft_store_unavailable", "environment failure drifted")
    require(environment.get("binding_runtime_allowed_now") is False, "environment binding runtime forbidden")
    require("role_policy_environment" in (environment.get("required_bindings") or []), "role env binding missing")

    metadata = fixture.get("role_claim_metadata_boundary") or {}
    require(
        metadata.get("status") == "metadata_only_no_claim_parser_runtime",
        "role claim metadata status drifted",
    )
    for forbidden in ("full_user_claim", "authorization_header", "cookie", "raw_role_grant", "raw_dsn"):
        require(forbidden in (metadata.get("forbidden_outputs") or []), f"missing forbidden role output: {forbidden}")
    require(metadata.get("role_claim_parser_allowed_now") is False, "role claim parser forbidden")
    require(metadata.get("token_validation_allowed_now") is False, "token validation forbidden")

    denial = fixture.get("cross_environment_denial_smoke") or {}
    require(denial.get("status") == "blocked_before_connection_smoke_strategy", "denial smoke status drifted")
    require(denial.get("denial_smoke_allowed_now") is False, "denial smoke must remain forbidden")
    require(denial.get("connection_smoke_allowed_now") is False, "connection smoke must remain forbidden")
    require(denial.get("database_connection_allowed_now") is False, "database connection must remain forbidden")


def assert_failure_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "database_role_policy_readiness_fail_closed_required",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    for field in (
        "runtime_role_policy_missing_failure_code",
        "migration_role_policy_missing_failure_code",
        "least_privilege_review_missing_failure_code",
        "role_environment_mismatch_failure_code",
        "role_escalation_attempt_failure_code",
        "cross_environment_denial_smoke_missing_failure_code",
    ):
        require(mapping.get(field) == "draft_store_unavailable", f"{field} drifted")
    for field in (
        "runtime_role_failure_must_prevent_query_execution",
        "migration_role_failure_must_prevent_marker_write",
        "role_policy_failure_must_prevent_connection_open",
        "environment_mismatch_must_prevent_secret_resolution",
        "cross_environment_denial_gap_must_prevent_provider_task_card",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

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
        strategy.get("status") == "database_role_policy_readiness_checker_only_no_runtime",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_database_role_policy_readiness_checker",
        "workflow_saved_draft_database_driver_dsn_tls_policy_readiness_checker",
        "workflow_saved_draft_database_connection_provider_implementation_entry_refresh_checker",
        "workflow_saved_draft_repository_mode_runtime_boundary_review_checker",
        "radish_oidc_token_membership_implementation_entry_review_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_enable_repository_store_mode",
        "does_not_call_fake_resolver",
        "does_not_call_production_resolver",
        "does_not_resolve_secret",
        "does_not_store_secret_value",
        "does_not_import_database_driver",
        "does_not_add_database_driver_dependency",
        "does_not_open_database_driver",
        "does_not_create_dsn_parser",
        "does_not_build_dsn",
        "does_not_create_role_policy_runtime",
        "does_not_parse_role_claim",
        "does_not_create_database_role",
        "does_not_grant_database_role",
        "does_not_connect_database",
        "does_not_run_connection_smoke",
        "does_not_run_sql",
        "does_not_read_schema_marker",
        "does_not_write_schema_marker",
        "does_not_create_migration_runner",
        "does_not_create_runner_command",
        "does_not_call_oidc",
        "does_not_validate_token",
        "does_not_create_membership_adapter",
        "does_not_create_production_api",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-database-role-policy-readiness-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "database driver DSN TLS policy readiness checker missing")
    require(current_checker in check_repo, "database role policy readiness checker missing")
    require(next_checker in check_repo, "product surface checker missing")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "database role policy readiness checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice_and_boundary(fixture)
    assert_role_policy_matrices(fixture)
    assert_failure_and_safety(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft database role policy readiness v1 checks passed.")


if __name__ == "__main__":
    main()
