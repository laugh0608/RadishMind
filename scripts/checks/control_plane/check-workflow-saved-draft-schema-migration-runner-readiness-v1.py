#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-schema-migration-runner-readiness-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
MANIFEST_PATH = REPO_ROOT / "services/platform/migrations/workflow_saved_drafts/manifest.json"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-schema-migration-preconditions-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-migration-preconditions-v1.json",
        "draft_schema_migration_preconditions_defined",
    ),
    "workflow-saved-draft-schema-artifact-materialization-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json",
        "draft_schema_artifact_materialized_static",
    ),
    "workflow-saved-draft-repository-adapter-implementation-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json",
        "draft_repository_adapter_implemented",
    ),
    "workflow-saved-draft-adapter-smoke-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json",
        "draft_adapter_smoke_executed",
    ),
    "workflow-saved-draft-production-auth-runtime-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json",
        "draft_production_auth_runtime_bridge_implemented",
    ),
    "workflow-saved-draft-repository-mode-enablement-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-enablement-v1.json",
        "draft_repository_mode_enablement_review_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "sql_migration_created",
    "schema_version_table_created",
    "migration_runner_implemented",
    "migration_runner_command_created",
    "database_connection_ready",
    "database_query_ready",
    "repository_store_mode_enabled",
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
EXPECTED_READINESS_GATES = {
    "static_schema_artifact_materialized": "satisfied",
    "repository_mode_enablement_review_completed": "satisfied",
    "adapter_schema_preflight_consumes_manifest": "satisfied",
    "adapter_smoke_executed_with_fake_query_executor": "satisfied",
    "production_auth_runtime_bridge_implemented": "satisfied",
    "rollback_evidence_materialized": "satisfied",
    "executable_sql_migration_artifact": "not_satisfied",
    "schema_version_applied_marker_contract": "not_satisfied",
    "migration_runner_command_boundary": "not_satisfied",
    "migration_runner_idempotency_smoke": "not_satisfied",
    "migration_runner_rollback_observability_smoke": "not_satisfied",
    "database_connection_provider": "not_satisfied",
}
EXPECTED_CONFIG_GATES = {
    "runtime_repository_store_still_disabled": "satisfied",
    "runner_config_separated_from_runtime_store_mode": "satisfied",
    "service_startup_auto_migration_forbidden": "satisfied",
    "manual_runner_entry_defined": "not_satisfied",
    "database_connection_provider_defined": "not_satisfied",
}
EXPECTED_SCHEMA_PREFLIGHT_GATES = {
    "manifest_store_schema_version_declared": "satisfied",
    "adapter_preflight_store_schema_version_check": "satisfied",
    "migration_state_failure_mapping_declared": "satisfied",
    "query_execution_blocked_when_preflight_fails": "satisfied",
    "applied_migration_marker_available": "not_satisfied",
    "runtime_store_schema_version_read_available": "not_satisfied",
}
EXPECTED_FAILURE_CODES = {
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_store_unavailable",
    "draft_store_contract_mismatch",
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_auth_context_contract_mismatch",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "database_connection_count=0",
    "sql_execution_count=0",
    "schema_marker_write_count=0",
    "repository_write_count=0",
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
        fixture.get("kind") == "workflow_saved_draft_schema_migration_runner_readiness_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-schema-migration-runner-readiness-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_schema_migration_runner_readiness_defined",
        "status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_runner_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("runner_boundary") or {}
    require(
        boundary.get("status") == "schema_migration_runner_readiness_defined_no_runner",
        "runner boundary status drifted",
    )
    require(boundary.get("decision") == "runner_not_implemented", "runner decision drifted")
    for field in (
        "migration_root",
        "manifest_file",
        "ddl_review_file",
        "rollback_evidence_file",
        "migration_smoke_file",
    ):
        relative_path = str(boundary.get(field) or "")
        require(relative_path, f"{field} missing")
        require((REPO_ROOT / relative_path).exists(), f"{field} path missing: {relative_path}")
    for field in (
        "manual_runner_entry_allowed_now",
        "service_startup_auto_migration_allowed_now",
        "http_route_migration_allowed_now",
        "selector_runtime_migration_allowed_now",
        "database_connection_allowed_now",
        "sql_execution_allowed_now",
        "schema_marker_write_allowed_now",
        "repository_mode_success_allowed_now",
        "oidc_middleware_allowed_now",
        "production_api_allowed_now",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gate_matrix(fixture: dict[str, Any], key: str, expected: dict[str, str]) -> None:
    rows = rows_by_id(fixture, key, "gate_id")
    require(set(rows) == set(expected), f"{key} gate ids drifted")
    for gate_id, status in expected.items():
        require(rows[gate_id].get("status") == status, f"{gate_id} status drifted")


def assert_manifest_and_preflight(fixture: dict[str, Any]) -> None:
    manifest = load_json(MANIFEST_PATH)
    require(
        manifest.get("store_schema_version") == "saved_workflow_drafts_store_v1",
        "manifest store_schema_version drifted",
    )
    require(
        manifest.get("schema_artifact_id") == "workflow_saved_draft_schema_artifact_v1",
        "manifest schema artifact id drifted",
    )
    rows = rows_by_id(fixture, "schema_preflight_matrix", "gate_id")
    require(
        rows["manifest_store_schema_version_declared"].get("expected_store_schema_version")
        == manifest.get("store_schema_version"),
        "fixture expected store schema version must match manifest",
    )
    adapter_source = read("services/platform/internal/httpapi/workflow_saved_draft_repository_adapter.go")
    for literal in (
        "SavedWorkflowDraftRepositorySchemaPreflight",
        "savedWorkflowDraftRepositoryMigrationNotApplied",
        "SavedWorkflowDraftFailureSchemaMigrationNotApplied",
        "SavedWorkflowDraftFailureStoreSchemaVersionMismatch",
        "SavedWorkflowDraftFailureStoreMigrationUnavailable",
    ):
        require(literal in adapter_source, f"adapter schema preflight literal missing: {literal}")


def assert_failure_rollback_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "schema_migration_runner_readiness_fail_closed_defined",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    require(
        mapping.get("runner_unavailable_failure_code") == "draft_store_migration_unavailable",
        "runner unavailable failure drifted",
    )
    require(
        mapping.get("missing_applied_marker_failure_code") == "draft_schema_migration_not_applied",
        "missing marker failure drifted",
    )
    require(
        mapping.get("store_schema_mismatch_failure_code") == "draft_store_schema_version_mismatch",
        "schema mismatch failure drifted",
    )
    for field in (
        "schema_preflight_failure_must_prevent_query_execution",
        "runner_failure_must_not_fallback_to_memory_dev",
        "migration_failure_must_not_return_draft_payload",
        "fake_query_executor_smoke_must_not_be_promoted_to_database_smoke",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

    failure_source = read("services/platform/internal/httpapi/workflow_saved_draft.go")
    for failure_code in EXPECTED_FAILURE_CODES:
        require(failure_code in failure_source, f"saved draft failure constant missing {failure_code}")

    rollback = fixture.get("rollback_policy") or {}
    rollback_path = str(rollback.get("static_rollback_evidence") or "")
    require(rollback.get("static_rollback_evidence_status") == "satisfied", "rollback evidence drifted")
    require((REPO_ROOT / rollback_path).exists(), f"rollback evidence missing: {rollback_path}")
    for field in (
        "runtime_rollback_runner_status",
        "rollback_command_status",
        "repository_mode_rollback_smoke_status",
        "forward_only_exception_status",
    ):
        require(rollback.get(field) == "not_satisfied", f"{field} must remain not_satisfied")
    require(
        rollback.get("rollback_to_memory_dev_after_migration_failure_allowed") is False,
        "rollback fallback must remain false",
    )

    fallback = fixture.get("no_fallback_policy") or {}
    for field, value in fallback.items():
        require(value is False, f"{field} must remain false")
    side_effects = fixture.get("no_side_effect_policy") or {}
    missing_counters = sorted(
        EXPECTED_SIDE_EFFECT_COUNTERS - set(side_effects.get("side_effect_counters_must_remain") or [])
    )
    require(not missing_counters, f"missing side effect counters: {missing_counters}")
    require(side_effects.get("forbidden_side_effects"), "forbidden side effects must be listed")


def assert_sources_and_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("required_static_artifacts_exist") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required static artifact missing: {relative_path}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")
    for root in guard.get("forbid_sql_files_under") or []:
        sql_files = sorted((REPO_ROOT / str(root)).rglob("*.sql"))
        require(not sql_files, f"SQL files must not exist under {root}: {sql_files}")
    for source_path in guard.get("source_files_to_scan") or []:
        source = read(str(source_path))
        for literal in guard.get("forbidden_source_literals") or []:
            require(str(literal) not in source, f"{source_path} contains forbidden literal: {literal}")

    selector_source = read("services/platform/internal/httpapi/workflow_saved_draft_store_selector.go")
    route_source = read("services/platform/internal/httpapi/workflow_saved_draft_http.go")
    require("case WorkflowSavedDraftStoreModeRepository:" in selector_source, "repository selector case missing")
    require("SavedWorkflowDraftFailureRepositoryStoreDisabled" in selector_source, "repository disabled mapping missing")
    require("SelectWorkflowSavedDraftStore(cfg.WorkflowSavedDraftStoreMode" in route_source, "route selector missing")


def assert_docs_testing_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        content = read(path)
        for literal in reference.get("must_contain") or []:
            require(str(literal) in content, f"{path} missing {literal}")

    strategy = fixture.get("testing_strategy") or {}
    require(
        strategy.get("status") == "schema_migration_runner_readiness_checker_only_no_runner",
        "testing strategy drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "go test ./internal/httpapi",
        "workflow_saved_draft_schema_migration_runner_readiness_checker",
        "workflow_saved_draft_repository_mode_enablement_checker",
        "workflow_saved_draft_schema_artifact_materialization_checker",
        "./scripts/check-repo.sh --fast",
        "./scripts/check-repo.sh",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_enable_repository_store_mode",
        "does_not_connect_database",
        "does_not_run_sql",
        "does_not_write_schema_marker",
        "does_not_create_migration_runner",
        "does_not_call_oidc",
        "does_not_validate_token",
        "does_not_create_membership_adapter",
        "does_not_create_production_api",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-repository-mode-enablement-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-schema-migration-runner-readiness-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "repository mode enablement checker missing from check-repo")
    require(current_checker in check_repo, "schema migration runner readiness checker missing from check-repo")
    require(next_checker in check_repo, "product surface recheck missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "schema migration runner readiness checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_runner_boundary(fixture)
    assert_gate_matrix(fixture, "readiness_gate_matrix", EXPECTED_READINESS_GATES)
    assert_gate_matrix(fixture, "config_gate_matrix", EXPECTED_CONFIG_GATES)
    assert_gate_matrix(fixture, "schema_preflight_matrix", EXPECTED_SCHEMA_PREFLIGHT_GATES)
    assert_manifest_and_preflight(fixture)
    assert_failure_rollback_and_safety(fixture)
    assert_sources_and_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft schema migration runner readiness v1 checks passed.")


if __name__ == "__main__":
    main()
