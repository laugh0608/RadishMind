#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.json",
        "draft_schema_marker_contract_implementation_entry_review_defined",
    ),
    "workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1.json",
        "draft_manual_migration_runner_implementation_entry_refresh_defined",
    ),
    "workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1.json",
        "draft_schema_marker_migration_runner_readiness_refresh_defined",
    ),
    "workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.json",
        "draft_schema_migration_runner_implementation_entry_review_defined",
    ),
    "workflow-saved-draft-database-connection-schema-marker-preconditions-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-schema-marker-preconditions-v1.json",
        "draft_database_connection_schema_marker_preconditions_defined",
    ),
    "workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2.json",
        "draft_database_connection_provider_implementation_entry_refresh_v2_defined",
    ),
    "workflow-saved-draft-database-connection-lifecycle-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-lifecycle-readiness-v1.json",
        "draft_database_connection_lifecycle_readiness_defined",
    ),
    "workflow-saved-draft-database-connection-smoke-strategy-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-smoke-strategy-v1.json",
        "draft_database_connection_smoke_strategy_defined",
    ),
    "workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.json",
        "draft_database_driver_dsn_tls_policy_readiness_defined",
    ),
    "workflow-saved-draft-database-role-policy-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-role-policy-readiness-v1.json",
        "draft_database_role_policy_readiness_defined",
    ),
    "workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.json",
        "draft_database_secret_resolver_implementation_entry_review_defined",
    ),
    "workflow-saved-draft-repository-mode-runtime-boundary-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-runtime-boundary-review-v1.json",
        "draft_repository_mode_runtime_boundary_review_defined",
    ),
    "workflow-saved-draft-schema-artifact-materialization-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json",
        "draft_schema_artifact_materialized_static",
    ),
    "workflow-saved-draft-production-auth-runtime-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json",
        "draft_production_auth_runtime_bridge_implemented",
    ),
    "radish-oidc-token-membership-upstream-evidence-refresh-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json",
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "schema_marker_runtime_implementation_task_card_created",
    "schema_marker_contract_implementation_task_card_created",
    "schema_marker_runtime_created",
    "schema_marker_read_write_ready",
    "schema_version_table_created",
    "schema_marker_reader_created",
    "schema_marker_writer_created",
    "migration_runner_implemented",
    "migration_runner_command_created",
    "dry_run_plan_output_ready",
    "migration_apply_smoke_ready",
    "rollback_observability_runtime_ready",
    "sql_migration_created",
    "database_connection_provider_ready",
    "database_connection_ready",
    "database_driver_imported",
    "database_query_executor_ready",
    "database_secret_resolver_ready",
    "connection_lifecycle_runtime_ready",
    "connection_smoke_runtime_ready",
    "repository_store_mode_enabled",
    "repository_mode_runtime_created",
    "durable_persistence_ready",
    "oidc_middleware_ready",
    "token_validation_ready",
    "workspace_membership_adapter_ready",
    "negative_auth_smoke_runtime_ready",
    "production_api_consumer_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_BOUNDARY_FIELDS = {
    "status": "schema_marker_runtime_dependency_refresh_defined_no_task_card",
    "decision": "schema_marker_runtime_still_blocked_before_implementation_task_card",
    "current_development_mode": "dependency_refresh_only_no_marker_runtime",
    "selected_implementation_track_now": "none",
    "future_schema_marker_runtime_task_card": (
        "docs/task-cards/workflow-saved-draft-schema-marker-implementation-v1-plan.md"
    ),
    "future_schema_marker_contract_task_card": (
        "docs/task-cards/workflow-saved-draft-schema-marker-contract-implementation-v1-plan.md"
    ),
    "future_schema_marker_table": "workflow_saved_draft_schema_versions",
}
EXPECTED_TRUE_BOUNDARY_FLAGS = {
    "marker_contract_static_readiness_satisfied",
    "manual_runner_static_review_consumed",
    "connection_provider_v2_static_review_consumed",
    "connection_lifecycle_static_readiness_consumed",
    "oidc_upstream_static_evidence_consumed",
}
EXPECTED_FALSE_BOUNDARY_FLAGS = {
    "implementation_task_card_created_in_this_slice",
    "schema_marker_contract_task_card_created_in_this_slice",
    "schema_marker_runtime_created_in_this_slice",
    "schema_marker_reader_created_in_this_slice",
    "schema_marker_writer_created_in_this_slice",
    "schema_version_table_created_in_this_slice",
    "sql_files_created_in_this_slice",
    "migration_runner_created_in_this_slice",
    "runner_command_created_in_this_slice",
    "dry_run_output_created_in_this_slice",
    "database_connection_provider_created_in_this_slice",
    "database_secret_resolver_created_in_this_slice",
    "database_driver_imported_in_this_slice",
    "database_connection_opened_in_this_slice",
    "database_query_executor_created_in_this_slice",
    "connection_lifecycle_runtime_created_in_this_slice",
    "connection_smoke_runtime_created_in_this_slice",
    "repository_mode_enabled_in_this_slice",
    "repository_mode_runtime_created_in_this_slice",
    "oidc_middleware_created_in_this_slice",
    "token_validation_created_in_this_slice",
    "membership_adapter_created_in_this_slice",
    "negative_auth_smoke_runtime_created_in_this_slice",
    "production_api_consumer_created_in_this_slice",
    "executor_created_in_this_slice",
    "confirmation_created_in_this_slice",
    "writeback_or_replay_created_in_this_slice",
}
EXPECTED_GATE_STATUS = {
    "schema_marker_contract_entry_review_consumed": "blocked_before_implementation_task_card",
    "manual_migration_runner_entry_refresh_consumed": "blocked_before_implementation_task_card",
    "schema_marker_migration_runner_readiness_refresh_consumed": "satisfied_static_readiness_no_runtime",
    "schema_migration_runner_entry_review_consumed": "blocked",
    "database_connection_schema_marker_preconditions_consumed": "satisfied_static_contract_no_runtime",
    "database_connection_provider_entry_refresh_v2_consumed": "blocked_before_implementation_task_card",
    "connection_lifecycle_readiness_consumed": "satisfied_static_readiness_no_runtime",
    "connection_smoke_strategy_consumed": "satisfied_static_strategy_no_runtime",
    "database_driver_dsn_tls_policy_consumed": "satisfied_static_readiness_no_runtime",
    "database_role_policy_consumed": "satisfied_static_readiness_no_runtime",
    "database_secret_resolver_entry_review_consumed": "blocked_before_implementation_task_card",
    "repository_mode_runtime_boundary_review_consumed": "blocked_before_task_card",
    "schema_artifact_materialization_consumed": "satisfied_static_artifact_only",
    "production_auth_runtime_bridge_consumed": "implemented_bridge_only_no_oidc_runtime",
    "radish_oidc_upstream_evidence_refresh_consumed": "satisfied_static_contract_no_auth_runtime",
    "schema_marker_runtime_task_card_gate": "blocked_before_implementation_task_card",
    "no_schema_marker_runtime_artifacts_leaked": "required_now",
}
EXPECTED_DEPENDENCY_STATUS = {
    "marker_contract_shape": "static_contract_defined_no_runtime",
    "applied_marker_source": "static_contract_defined_no_marker_table",
    "marker_read_path": "blocked_no_marker_reader",
    "marker_write_path": "blocked_no_marker_writer",
    "manual_runner_handoff": "blocked_no_runner_runtime",
    "database_connection_dependency": "blocked_no_connection_provider",
    "connection_lifecycle_dependency": "static_readiness_defined_no_runtime",
    "connection_smoke_dependency": "static_strategy_defined_no_runtime_smoke",
    "repository_mode_dependency": "blocked_no_repository_mode_runtime",
    "auth_membership_dependency": "static_upstream_evidence_defined_no_auth_runtime",
    "idempotency_lock_dependency": "static_contract_defined_no_runtime",
    "rollback_observability_dependency": "static_contract_defined_no_runtime",
}
EXPECTED_RUNTIME_BLOCKERS = {
    "schema_marker_reader_not_created",
    "schema_marker_writer_not_created",
    "schema_version_table_not_created",
    "manual_migration_runner_not_created",
    "database_connection_provider_not_created",
    "repository_mode_runtime_disabled",
    "token_validation_schema_missing",
    "auth_middleware_not_created",
    "membership_adapter_not_created",
    "connection_smoke_runtime_missing",
}
EXPECTED_FUTURE_TASK_REQUIREMENTS = {
    "marker_read_write_interface_owner",
    "applied_marker_table_or_schema_version_table_contract",
    "marker_read_preflight_failure_mapping",
    "marker_write_manual_runner_ownership",
    "manual_runner_handoff_and_apply_result_envelope",
    "idempotency_lock_duplicate_handling",
    "rollback_forward_only_observability",
    "connection_provider_query_role_lifecycle_smoke_runtime_evidence",
    "auth_membership_context_and_negative_auth_smoke_runtime_evidence",
    "negative_marker_smoke_matrix",
}
EXPECTED_FAILURE_CODES = {
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_store_unavailable",
    "draft_audit_context_missing",
    "draft_store_contract_mismatch",
    "draft_auth_context_contract_mismatch",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "schema_marker_read_count=0",
    "schema_marker_write_count=0",
    "schema_version_table_create_count=0",
    "migration_runner_call_count=0",
    "runner_command_count=0",
    "sql_execution_count=0",
    "database_connection_count=0",
    "connection_provider_call_count=0",
    "repository_mode_enablement_count=0",
    "token_validation_count=0",
    "membership_lookup_count=0",
    "production_api_call_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_NEXT_CANDIDATES = {
    "secret_resolver_runtime_dependency_refresh",
    "token_validation_schema_auth_middleware_runtime_entry_review",
    "manual_migration_runner_runtime_dependency_refresh",
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
    declared = set(fixture.get("depends_on") or [])
    missing = sorted(set(EXPECTED_DEPENDENCIES) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"{dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"{dependency_id} status drifted")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "workflow_saved_draft_schema_marker_runtime_dependency_refresh_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_schema_marker_runtime_dependency_refresh_defined",
        "status drifted",
    )
    require(
        slice_info.get("task_card")
        == "docs/task-cards/workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1-plan.md",
        "task card drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("dependency_refresh_boundary") or {}
    for field, expected_value in EXPECTED_BOUNDARY_FIELDS.items():
        require(boundary.get(field) == expected_value, f"{field} drifted")
    for field in EXPECTED_TRUE_BOUNDARY_FLAGS:
        require(boundary.get(field) is True, f"{field} must be true")
    for field in EXPECTED_FALSE_BOUNDARY_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gate_and_dependency_matrix(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "entry_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "entry gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    dependencies = rows_by_id(fixture, "dependency_matrix", "dependency_id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCY_STATUS), "dependency matrix ids drifted")
    for dependency_id, expected_status in EXPECTED_DEPENDENCY_STATUS.items():
        dependency = dependencies[dependency_id]
        require(dependency.get("status") == expected_status, f"{dependency_id} status drifted")
        require(dependency.get("must_cover"), f"{dependency_id} must describe required coverage")

    blockers = set(fixture.get("runtime_blockers") or [])
    missing_blockers = sorted(EXPECTED_RUNTIME_BLOCKERS - blockers)
    require(not missing_blockers, f"missing runtime blockers: {missing_blockers}")

    requirements = set(fixture.get("future_task_requirements") or [])
    missing_requirements = sorted(EXPECTED_FUTURE_TASK_REQUIREMENTS - requirements)
    require(not missing_requirements, f"missing future task requirements: {missing_requirements}")


def assert_failure_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "schema_marker_runtime_dependency_refresh_fail_closed_required",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    require(mapping.get("missing_marker_failure_code") == "draft_schema_migration_not_applied", "marker missing drifted")
    require(
        mapping.get("version_mismatch_failure_code") == "draft_store_schema_version_mismatch",
        "version mismatch drifted",
    )
    require(
        mapping.get("marker_unavailable_failure_code") == "draft_store_migration_unavailable",
        "marker unavailable drifted",
    )
    require(
        mapping.get("connection_unavailable_failure_code") == "draft_store_unavailable",
        "connection failure drifted",
    )
    require(
        mapping.get("audit_context_missing_failure_code") == "draft_audit_context_missing",
        "audit context failure drifted",
    )
    require(
        mapping.get("repository_mode_still_disabled_failure_code") == "repository_store_disabled",
        "repository mode failure drifted",
    )
    for field in (
        "marker_failure_must_prevent_query_execution",
        "connection_failure_must_prevent_marker_read",
        "auth_failure_must_prevent_repository_access",
        "static_contract_must_not_imply_runtime_ready",
        "must_not_fallback_to_memory_dev",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

    source = "\n".join(
        [
            read("services/platform/internal/httpapi/workflow_saved_draft.go"),
            read("services/platform/internal/httpapi/workflow_saved_draft_store_selector.go"),
            read("services/platform/internal/httpapi/workflow_saved_draft_repository_adapter.go"),
        ]
    )
    for failure_code in EXPECTED_FAILURE_CODES:
        require(failure_code in source, f"saved draft failure code missing: {failure_code}")

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
        strategy.get("status") == "schema_marker_runtime_dependency_refresh_checker_only_no_runtime",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_schema_marker_runtime_dependency_refresh_checker",
        "workflow_saved_draft_database_connection_provider_implementation_entry_refresh_v2_checker",
        "workflow_saved_draft_schema_marker_contract_implementation_entry_review_checker",
        "radish_oidc_token_membership_upstream_evidence_refresh_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_enable_repository_store_mode",
        "does_not_call_fake_resolver",
        "does_not_call_production_resolver",
        "does_not_connect_database",
        "does_not_import_database_driver",
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

    next_candidates = set(fixture.get("next_candidates") or [])
    missing_next = sorted(EXPECTED_NEXT_CANDIDATES - next_candidates)
    require(not missing_next, f"missing next candidates: {missing_next}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    oidc_checker = "checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py"
    previous_checker = (
        "checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2.py"
    )
    current_checker = (
        "checks/control_plane/check-workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    for checker in (oidc_checker, previous_checker, current_checker, next_checker):
        require(checker in check_repo, f"{checker} missing from check-repo")
    require(
        check_repo.index(oidc_checker) < check_repo.index(current_checker),
        "OIDC upstream evidence checker must run before schema marker runtime dependency refresh",
    )
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "schema marker runtime dependency refresh checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_gate_and_dependency_matrix(fixture)
    assert_failure_and_safety(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft schema marker runtime dependency refresh checks passed.")


if __name__ == "__main__":
    main()
