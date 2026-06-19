#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-schema-migration-runner-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-migration-runner-readiness-v1.json",
        "draft_schema_migration_runner_readiness_defined",
    ),
    "workflow-saved-draft-repository-mode-enablement-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-enablement-v1.json",
        "draft_repository_mode_enablement_review_defined",
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
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "schema_migration_runner_implementation_task_card_created",
    "sql_migration_created",
    "schema_version_table_created",
    "schema_version_marker_contract_ready",
    "migration_runner_implemented",
    "migration_runner_command_created",
    "dry_run_plan_output_ready",
    "migration_apply_smoke_ready",
    "rollback_observability_ready",
    "database_connection_ready",
    "database_query_executor_ready",
    "repository_store_mode_enabled",
    "durable_persistence_ready",
    "oidc_middleware_ready",
    "token_validation_ready",
    "workspace_membership_adapter_ready",
    "production_api_consumer_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_GATE_STATUS = {
    "schema_migration_runner_readiness_consumed": "satisfied",
    "repository_mode_enablement_review_consumed": "satisfied",
    "static_schema_artifact_consumed": "satisfied_static",
    "repository_adapter_implementation_consumed": "satisfied",
    "adapter_smoke_fake_executor_consumed": "satisfied_fake_executor_only",
    "production_auth_runtime_bridge_consumed": "satisfied_bridge_only",
    "implementation_entry_review_defined": "satisfied",
    "executable_migration_artifact_gate": "blocked",
    "schema_version_marker_contract_gate": "blocked",
    "manual_runner_command_gate": "blocked",
    "dry_run_plan_output_gate": "blocked",
    "migration_apply_smoke_gate": "blocked",
    "rollback_observability_smoke_gate": "blocked",
    "database_connection_provider_gate": "blocked",
    "repository_mode_runtime_enablement_gate": "blocked",
    "no_runner_implementation_artifacts_leaked": "required_now",
}
EXPECTED_CANDIDATES = {
    "executable_migration_artifact",
    "schema_version_marker_contract",
    "manual_runner_command",
    "migration_apply_smoke",
    "rollback_observability_smoke",
    "repository_mode_runtime_enablement",
}
EXPECTED_FUTURE_CONTRACTS = {
    "manual_trigger_only": "required_for_future_task",
    "dry_run_plan_output": "not_satisfied",
    "applied_marker_read_write": "not_satisfied",
    "database_connection_provider": "not_satisfied",
    "rollback_or_forward_only_policy": "not_satisfied",
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
        fixture.get("kind") == "workflow_saved_draft_schema_migration_runner_implementation_entry_review_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_schema_migration_runner_implementation_entry_review_defined",
        "status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_entry_boundary") or {}
    require(
        boundary.get("status") == "schema_migration_runner_implementation_entry_review_defined_no_entry_opened",
        "entry boundary status drifted",
    )
    require(
        boundary.get("decision") == "schema_migration_runner_implementation_entry_not_opened",
        "entry decision drifted",
    )
    require(boundary.get("current_development_mode") == "entry_review_only_no_runner", "mode drifted")
    require(boundary.get("selected_implementation_track_now") == "none", "selected track must remain none")
    require(
        boundary.get("future_implementation_task_card")
        == "docs/task-cards/workflow-saved-draft-schema-migration-runner-implementation-v1-plan.md",
        "future implementation task card path drifted",
    )
    for field in (
        "implementation_task_card_created_in_this_slice",
        "runner_file_created_in_this_slice",
        "runner_command_created_in_this_slice",
        "sql_files_created_in_this_slice",
        "schema_version_table_created_in_this_slice",
        "schema_marker_contract_created_in_this_slice",
        "database_connection_allowed_in_this_slice",
        "database_query_executor_allowed_in_this_slice",
        "manual_runner_command_allowed_in_this_slice",
        "service_startup_auto_migration_allowed_in_this_slice",
        "http_route_migration_allowed_in_this_slice",
        "selector_runtime_migration_allowed_in_this_slice",
        "repository_mode_success_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "auth_middleware_allowed_in_this_slice",
        "membership_adapter_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "publish_or_run_allowed_in_this_slice",
        "executor_allowed_in_this_slice",
        "writeback_or_replay_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gate_and_candidates(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "entry_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "entry gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    candidates = rows_by_id(fixture, "implementation_entry_candidates", "candidate_id")
    require(set(candidates) == EXPECTED_CANDIDATES, "candidate ids drifted")
    for candidate_id, candidate in candidates.items():
        require(candidate.get("entry_decision") == "blocked", f"{candidate_id} decision must remain blocked")
        require(candidate.get("trigger_satisfied_now") is False, f"{candidate_id} trigger must remain false")
        require(candidate.get("task_card_allowed_next") is False, f"{candidate_id} task card must remain blocked")
        require(
            candidate.get("implementation_artifacts_allowed_now") is False,
            f"{candidate_id} artifacts must remain blocked",
        )
        require(candidate.get("current_blockers"), f"{candidate_id} must list blockers")


def assert_future_contracts(fixture: dict[str, Any]) -> None:
    contracts = rows_by_id(fixture, "future_runner_contract_matrix", "contract_id")
    require(set(contracts) == set(EXPECTED_FUTURE_CONTRACTS), "future contract ids drifted")
    for contract_id, expected_status in EXPECTED_FUTURE_CONTRACTS.items():
        require(contracts[contract_id].get("status") == expected_status, f"{contract_id} status drifted")
        require(contracts[contract_id].get("must_cover") or contracts[contract_id].get("must_not_be_triggered_by"), f"{contract_id} must describe requirements")


def assert_failure_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "schema_migration_runner_implementation_entry_fail_closed_required",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    require(mapping.get("runner_unavailable_failure_code") == "draft_store_migration_unavailable", "runner failure drifted")
    require(mapping.get("missing_applied_marker_failure_code") == "draft_schema_migration_not_applied", "marker failure drifted")
    require(mapping.get("store_schema_mismatch_failure_code") == "draft_store_schema_version_mismatch", "schema mismatch drifted")
    require(mapping.get("repository_mode_still_disabled_failure_code") == "repository_store_disabled", "repository mode failure drifted")
    for field in (
        "schema_preflight_failure_must_prevent_query_execution",
        "runner_failure_must_not_fallback_to_memory_dev",
        "migration_failure_must_not_return_draft_payload",
        "fake_query_executor_smoke_must_not_be_promoted_to_database_smoke",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

    saved_draft_source = read("services/platform/internal/httpapi/workflow_saved_draft.go")
    for failure_code in EXPECTED_FAILURE_CODES:
        require(failure_code in saved_draft_source, f"saved draft failure code missing: {failure_code}")

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
        sql_files = sorted((REPO_ROOT / str(root)).rglob("*.sql"))
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
        strategy.get("status") == "schema_migration_runner_implementation_entry_review_checker_only_no_runner",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_schema_migration_runner_implementation_entry_review_checker",
        "workflow_saved_draft_schema_migration_runner_readiness_checker",
        "workflow_saved_draft_repository_mode_enablement_checker",
        "workflow_saved_draft_schema_artifact_materialization_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_enable_repository_store_mode",
        "does_not_connect_database",
        "does_not_run_sql",
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
    previous_checker = "checks/control_plane/check-workflow-saved-draft-schema-migration-runner-readiness-v1.py"
    current_checker = (
        "checks/control_plane/check-workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "readiness checker missing from check-repo")
    require(current_checker in check_repo, "implementation entry review checker missing from check-repo")
    require(next_checker in check_repo, "product surface checker missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "implementation entry review checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_gate_and_candidates(fixture)
    assert_future_contracts(fixture)
    assert_failure_and_safety(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft schema migration runner implementation entry review v1 checks passed.")


if __name__ == "__main__":
    main()
