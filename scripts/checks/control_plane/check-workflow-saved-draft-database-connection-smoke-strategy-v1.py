#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-database-connection-smoke-strategy-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.json",
        "draft_database_driver_dsn_tls_policy_readiness_defined",
    ),
    "workflow-saved-draft-database-role-policy-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-role-policy-readiness-v1.json",
        "draft_database_role_policy_readiness_defined",
    ),
    "workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.json",
        "draft_database_connection_provider_implementation_entry_refresh_defined",
    ),
    "workflow-saved-draft-database-secret-resolver-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-readiness-v1.json",
        "draft_database_secret_resolver_readiness_defined",
    ),
    "workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.json",
        "draft_database_secret_resolver_implementation_entry_review_defined",
    ),
    "workflow-saved-draft-repository-mode-runtime-boundary-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-runtime-boundary-review-v1.json",
        "draft_repository_mode_runtime_boundary_review_defined",
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
    "connection_smoke_runner_created",
    "connection_smoke_runtime_ready",
    "connection_smoke_executed",
    "connection_smoke_output_committed",
    "secret_resolver_ready",
    "production_resolver_runtime_ready",
    "database_driver_selected",
    "database_driver_imported",
    "dsn_parser_created",
    "raw_dsn_available",
    "tls_runtime_ready",
    "role_policy_runtime_ready",
    "connection_factory_implemented",
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
    "status": "database_connection_smoke_strategy_defined_no_runtime_smoke",
    "decision": "connection_smoke_strategy_defined_for_future_task_only",
    "current_development_mode": "strategy_only_no_runner_no_database_connection",
    "future_implementation_task_card": (
        "docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md"
    ),
    "future_connection_smoke_runtime_task_card": (
        "docs/task-cards/workflow-saved-draft-database-connection-smoke-runtime-v1-plan.md"
    ),
    "explicit_test_database_boundary_status": "defined_no_database_connection",
    "placeholder_credential_handoff_status": "defined_metadata_only_no_secret_resolution",
    "smoke_input_shape_status": "defined_static_contract_only",
    "smoke_output_shape_status": "defined_static_record_only",
    "role_denial_cases_status": "defined_no_runtime_denial_execution",
    "no_leakage_scan_status": "defined_static_scan_requirement",
    "manual_execution_boundary_status": "defined_no_fast_baseline_runtime",
    "future_connection_provider_dependency_status": "blocked_before_implementation_task_card",
}
EXPECTED_FALSE_BOUNDARY_FLAGS = {
    "implementation_task_card_created_in_this_slice",
    "connection_smoke_runner_created_in_this_slice",
    "connection_smoke_runtime_allowed_in_this_slice",
    "connection_smoke_execution_allowed_in_this_slice",
    "connection_smoke_output_allowed_in_this_slice",
    "database_connection_provider_file_created_in_this_slice",
    "database_connection_allowed_in_this_slice",
    "database_driver_import_allowed_in_this_slice",
    "go_mod_driver_dependency_allowed_in_this_slice",
    "dsn_parser_allowed_in_this_slice",
    "raw_dsn_material_allowed_in_this_slice",
    "tls_runtime_allowed_in_this_slice",
    "role_policy_runtime_allowed_in_this_slice",
    "connection_factory_allowed_in_this_slice",
    "secret_resolver_allowed_in_this_slice",
    "fake_resolver_invocation_allowed_in_this_slice",
    "production_resolver_invocation_allowed_in_this_slice",
    "credential_handle_creation_allowed_in_this_slice",
    "no_leakage_scan_runtime_allowed_in_this_slice",
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
    "driver_dsn_tls_policy_consumed": "consumed_static_policy",
    "role_policy_consumed": "consumed_static_policy",
    "explicit_test_database_boundary": "defined_no_database_connection",
    "placeholder_credential_handoff": "defined_metadata_only_no_secret_resolution",
    "smoke_input_shape": "defined_static_contract_only",
    "smoke_output_shape": "defined_static_record_only",
    "role_denial_cases": "defined_no_runtime_denial_execution",
    "no_leakage_scan": "defined_static_scan_requirement",
    "manual_only_execution_boundary": "defined_no_fast_baseline_runtime",
    "connection_provider_dependency": "blocked_before_implementation_task_card",
    "no_runtime_artifacts_leaked": "required_now",
}
EXPECTED_DENIAL_CASES = {
    "runtime_dml_role_ddl_denied",
    "runtime_dml_role_marker_write_denied",
    "migration_role_repository_adapter_denied",
    "cross_environment_role_binding_denied",
    "unknown_role_class_denied",
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
    "no_leakage_scan_count=0",
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
        fixture.get("kind") == "workflow_saved_draft_database_connection_smoke_strategy_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-database-connection-smoke-strategy-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_database_connection_smoke_strategy_defined",
        "status drifted",
    )
    require(
        slice_info.get("task_card")
        == "docs/task-cards/workflow-saved-draft-database-connection-smoke-strategy-v1-plan.md",
        "task card drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    boundary = fixture.get("strategy_boundary") or {}
    for field, expected_value in EXPECTED_BOUNDARY_FIELDS.items():
        require(boundary.get(field) == expected_value, f"{field} drifted")
    for field in EXPECTED_FALSE_BOUNDARY_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_strategy_matrices(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "strategy_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "strategy gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    test_db = fixture.get("explicit_test_database_boundary") or {}
    require(
        test_db.get("status") == "future_strategy_only_no_database_connection",
        "test database boundary status drifted",
    )
    require("production_database" in (test_db.get("forbidden_target_classes") or []), "production DB must be forbidden")
    require(test_db.get("database_connection_allowed_now") is False, "database connection forbidden now")
    require(test_db.get("network_call_allowed_now") is False, "network call forbidden now")

    credential = fixture.get("placeholder_credential_handoff") or {}
    require(
        credential.get("status") == "metadata_only_no_secret_resolution",
        "placeholder credential status drifted",
    )
    for forbidden in ("secret_value", "raw_dsn", "database_password", "credential_payload"):
        require(forbidden in (credential.get("forbidden_inputs") or []), f"missing forbidden input: {forbidden}")
    require(credential.get("secret_resolution_allowed_now") is False, "secret resolution forbidden")
    require(credential.get("fake_resolver_invocation_allowed_now") is False, "fake resolver invocation forbidden")
    require(
        credential.get("production_resolver_invocation_allowed_now") is False,
        "production resolver invocation forbidden",
    )

    smoke_input = fixture.get("smoke_input_shape") or {}
    require(smoke_input.get("status") == "static_contract_only_no_runner", "smoke input status drifted")
    for required in ("smoke_record_id", "environment_id", "expected_role_denial_cases"):
        require(required in (smoke_input.get("required_fields") or []), f"missing smoke input field: {required}")
    for forbidden in ("raw_dsn", "database_host", "password", "sql_text"):
        require(forbidden in (smoke_input.get("forbidden_fields") or []), f"missing forbidden smoke input: {forbidden}")
    require(smoke_input.get("runner_created_now") is False, "runner must not be created")

    smoke_output = fixture.get("smoke_output_shape") or {}
    require(smoke_output.get("status") == "static_record_only_no_smoke_output", "smoke output status drifted")
    for allowed in ("sanitized_status", "redacted_target_fingerprint", "role_denial_results"):
        require(allowed in (smoke_output.get("allowed_future_fields") or []), f"missing smoke output field: {allowed}")
    for forbidden in ("raw_dsn", "database_host", "database_error_detail", "secret_value", "sql_detail"):
        require(
            forbidden in (smoke_output.get("forbidden_future_fields") or []),
            f"missing forbidden smoke output: {forbidden}",
        )
    require(smoke_output.get("smoke_output_created_now") is False, "smoke output must not be created")

    denial = fixture.get("role_denial_cases") or {}
    require(denial.get("status") == "static_case_matrix_defined_no_execution", "role denial status drifted")
    cases = {
        str(row.get("case_id") or "")
        for row in denial.get("required_cases") or []
        if isinstance(row, dict)
    }
    missing_cases = sorted(EXPECTED_DENIAL_CASES - cases)
    require(not missing_cases, f"missing denial cases: {missing_cases}")
    require(denial.get("denial_smoke_execution_allowed_now") is False, "denial smoke execution forbidden")
    require(denial.get("role_policy_runtime_allowed_now") is False, "role policy runtime forbidden")

    scan = fixture.get("no_leakage_scan") or {}
    require(scan.get("status") == "static_scan_requirement_no_runtime_scanner", "scan status drifted")
    for surface in ("smoke_input", "smoke_output", "logs", "audit_event"):
        require(surface in (scan.get("future_scan_surfaces") or []), f"missing scan surface: {surface}")
    for forbidden in ("raw_dsn", "secret_value", "credential_payload", "database_error_detail"):
        require(forbidden in (scan.get("forbidden_terms") or []), f"missing forbidden scan term: {forbidden}")
    require(scan.get("runtime_scanner_created_now") is False, "runtime scanner forbidden")
    require(scan.get("smoke_runner_created_now") is False, "smoke runner forbidden")

    manual = fixture.get("manual_execution_boundary") or {}
    require(
        manual.get("status") == "manual_or_dedicated_runtime_task_only_no_fast_baseline_db",
        "manual execution boundary drifted",
    )
    for forbidden in ("connect database", "resolve secret", "run SQL"):
        require(forbidden in (manual.get("fast_baseline_must_not") or []), f"missing fast baseline ban: {forbidden}")
    require(
        manual.get("fast_baseline_database_connection_allowed_now") is False,
        "fast baseline DB connection forbidden",
    )


def assert_failure_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "database_connection_smoke_strategy_fail_closed_required",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    for field in (
        "smoke_target_missing_failure_code",
        "placeholder_credential_missing_failure_code",
        "policy_ref_missing_failure_code",
        "no_leakage_scan_missing_failure_code",
        "role_denial_case_missing_failure_code",
        "smoke_record_contract_mismatch_failure_code",
    ):
        require(mapping.get(field) == "draft_store_unavailable", f"{field} drifted")
    for field in (
        "smoke_gap_must_prevent_provider_task_card",
        "smoke_target_mismatch_must_prevent_secret_resolution",
        "placeholder_credential_gap_must_prevent_connection_open",
        "no_leakage_gap_must_prevent_smoke_runtime",
        "role_denial_gap_must_prevent_repository_mode",
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
        strategy.get("status") == "database_connection_smoke_strategy_checker_only_no_runtime",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_database_connection_smoke_strategy_checker",
        "workflow_saved_draft_database_role_policy_readiness_checker",
        "workflow_saved_draft_database_driver_dsn_tls_policy_readiness_checker",
        "workflow_saved_draft_database_connection_provider_implementation_entry_refresh_checker",
        "workflow_saved_draft_repository_mode_runtime_boundary_review_checker",
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
        "does_not_connect_database",
        "does_not_run_connection_smoke",
        "does_not_run_no_leakage_scan",
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
    previous_checker = "checks/control_plane/check-workflow-saved-draft-database-role-policy-readiness-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-database-connection-smoke-strategy-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "database role policy readiness checker missing")
    require(current_checker in check_repo, "database connection smoke strategy checker missing")
    require(next_checker in check_repo, "product surface checker missing")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "database connection smoke strategy checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice_and_boundary(fixture)
    assert_strategy_matrices(fixture)
    assert_failure_and_safety(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft database connection smoke strategy v1 checks passed.")


if __name__ == "__main__":
    main()
