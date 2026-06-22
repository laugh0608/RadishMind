#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
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
    "workflow-saved-draft-database-connection-provider-implementation-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.json",
        "draft_database_connection_provider_implementation_entry_review_defined",
    ),
    "workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.json",
        "draft_database_secret_resolver_implementation_entry_review_defined",
    ),
    "workflow-saved-draft-repository-mode-runtime-boundary-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-runtime-boundary-review-v1.json",
        "draft_repository_mode_runtime_boundary_review_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
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
    "secret_value_resolved",
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
    "status": "schema_marker_contract_implementation_entry_review_defined_no_task_card",
    "decision": "schema_marker_contract_implementation_task_card_blocked",
    "current_development_mode": "entry_review_only_no_marker_runtime",
    "selected_implementation_track_now": "none",
    "future_schema_marker_task_card": (
        "docs/task-cards/workflow-saved-draft-schema-marker-contract-implementation-v1-plan.md"
    ),
    "future_schema_marker_table": "workflow_saved_draft_schema_versions",
}
EXPECTED_FALSE_BOUNDARY_FLAGS = {
    "schema_marker_task_card_created_in_this_slice",
    "schema_marker_runtime_created_in_this_slice",
    "schema_marker_reader_created_in_this_slice",
    "schema_marker_writer_created_in_this_slice",
    "schema_version_table_created_in_this_slice",
    "migration_runner_created_in_this_slice",
    "runner_command_created_in_this_slice",
    "dry_run_output_created_in_this_slice",
    "sql_created_in_this_slice",
    "database_connection_provider_created_in_this_slice",
    "database_secret_resolver_created_in_this_slice",
    "secret_value_resolved_in_this_slice",
    "database_driver_imported_in_this_slice",
    "database_connection_opened_in_this_slice",
    "database_query_executor_created_in_this_slice",
    "repository_mode_enabled_in_this_slice",
    "repository_mode_runtime_created_in_this_slice",
    "oidc_middleware_created_in_this_slice",
    "token_validation_created_in_this_slice",
    "membership_adapter_created_in_this_slice",
    "production_api_consumer_created_in_this_slice",
    "executor_created_in_this_slice",
    "confirmation_created_in_this_slice",
    "writeback_or_replay_created_in_this_slice",
}
EXPECTED_GATE_STATUS = {
    "schema_marker_migration_runner_readiness_refresh_consumed": "satisfied_static_readiness",
    "schema_migration_runner_implementation_entry_review_consumed": "blocked",
    "database_connection_schema_marker_preconditions_consumed": "satisfied_contract_only",
    "database_connection_provider_entry_review_consumed": "blocked",
    "database_secret_resolver_entry_review_consumed": "blocked",
    "repository_mode_runtime_boundary_review_consumed": "blocked_before_task_card",
    "schema_marker_contract_task_card_gate": "blocked_before_implementation_task_card",
    "schema_marker_runtime_artifact_gate": "not_created",
}
EXPECTED_CANDIDATES = {
    "schema_marker_reader_contract_task_card",
    "schema_marker_writer_contract_task_card",
    "schema_marker_table_record_shape",
    "repository_preflight_consumption_binding",
    "negative_marker_smoke",
    "schema_marker_contract_implementation_task_card",
}
EXPECTED_MARKER_STATES = {
    "applied",
    "not_applied",
    "version_mismatch",
    "unavailable",
}
EXPECTED_BLOCKED_CONDITIONS = {
    "schema_marker_contract_implementation_task_card_not_created",
    "schema_marker_runtime_not_created",
    "schema_marker_reader_not_created",
    "schema_marker_writer_not_created",
    "schema_version_table_not_created",
    "manual_migration_runner_not_created",
    "dry_run_output_not_created",
    "sql_migration_not_created",
    "database_connection_provider_not_created",
    "database_secret_resolver_not_created",
    "real_query_executor_not_created",
    "oidc_middleware_not_created",
    "token_validation_not_created",
    "workspace_membership_adapter_not_created",
    "production_api_consumer_not_created",
    "repository_mode_runtime_task_card_not_created",
}
EXPECTED_NEXT_CANDIDATES = {
    "manual_migration_runner_implementation_entry_refresh",
    "database_connection_provider_implementation_entry_refresh",
    "radish_oidc_token_membership_upstream_evidence_refresh",
}
EXPECTED_FAILURE_CODES = {
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_store_unavailable",
    "draft_store_contract_mismatch",
    "draft_auth_context_contract_mismatch",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "database_connection_count=0",
    "driver_open_count=0",
    "sql_execution_count=0",
    "schema_marker_read_count=0",
    "schema_marker_write_count=0",
    "migration_runner_invocation_count=0",
    "repository_mode_enablement_count=0",
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
        fixture.get("kind") == "workflow_saved_draft_schema_marker_contract_implementation_entry_review_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_schema_marker_contract_implementation_entry_review_defined",
        "status drifted",
    )
    require(
        slice_info.get("task_card")
        == "docs/task-cards/workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1-plan.md",
        "task card drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_boundary") or {}
    for field, expected_value in EXPECTED_BOUNDARY_FIELDS.items():
        require(boundary.get(field) == expected_value, f"{field} drifted")
    for field in EXPECTED_FALSE_BOUNDARY_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gates_candidates_and_contract(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "entry_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    candidates = rows_by_id(fixture, "implementation_entry_candidates", "candidate_id")
    require(set(candidates) == EXPECTED_CANDIDATES, "candidate ids drifted")
    for candidate_id, candidate in candidates.items():
        if candidate_id == "schema_marker_contract_implementation_task_card":
            expected_decision = "blocked_before_implementation_task_card"
        else:
            expected_decision = "blocked"
        require(candidate.get("entry_decision") == expected_decision, f"{candidate_id} decision drifted")
        require(
            candidate.get("implementation_task_card_allowed_now") is False,
            f"{candidate_id} task card must remain false",
        )
        require(candidate.get("runtime_artifacts_allowed_now") is False, f"{candidate_id} runtime must remain false")
        require(candidate.get("current_blockers"), f"{candidate_id} blockers must be listed")

    contract = fixture.get("schema_marker_contract_boundary") or {}
    require(set(contract.get("allowed_marker_states") or []) == EXPECTED_MARKER_STATES, "marker states drifted")
    require(contract.get("reader_role") == "read_applied_marker_only", "reader role drifted")
    require(contract.get("writer_role") == "future_manual_migration_runner_only", "writer role drifted")
    for field in (
        "reader_must_not_create_marker",
        "request_path_must_not_write_marker",
        "startup_must_not_run_migration",
        "marker_failure_must_prevent_query_execution",
    ):
        require(contract.get(field) is True, f"{field} must remain true")
    require(
        contract.get("missing_marker_failure_code") == "draft_schema_migration_not_applied",
        "missing marker failure drifted",
    )
    require(
        contract.get("version_mismatch_failure_code") == "draft_store_schema_version_mismatch",
        "version mismatch failure drifted",
    )
    require(
        contract.get("unavailable_failure_code") == "draft_store_migration_unavailable",
        "unavailable failure drifted",
    )

    blocked = set(fixture.get("blocked_conditions") or [])
    missing_blockers = sorted(EXPECTED_BLOCKED_CONDITIONS - blocked)
    require(not missing_blockers, f"missing blocked conditions: {missing_blockers}")

    next_candidates = rows_by_id(fixture, "next_dependency_candidates", "candidate_id")
    require(set(next_candidates) == EXPECTED_NEXT_CANDIDATES, "next dependency candidates drifted")
    for candidate_id, candidate in next_candidates.items():
        require(
            candidate.get("entry_decision") == "eligible_for_independent_review_only",
            f"{candidate_id} decision drifted",
        )
        require(candidate.get("runtime_artifacts_allowed_now") is False, f"{candidate_id} runtime must remain false")
        require(candidate.get("current_blockers"), f"{candidate_id} blockers must be listed")


def assert_failure_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "schema_marker_contract_implementation_entry_review_fail_closed_required",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    require(
        mapping.get("missing_applied_marker_failure_code") == "draft_schema_migration_not_applied",
        "missing marker failure drifted",
    )
    require(
        mapping.get("marker_version_mismatch_failure_code") == "draft_store_schema_version_mismatch",
        "marker mismatch failure drifted",
    )
    require(
        mapping.get("marker_unavailable_failure_code") == "draft_store_migration_unavailable",
        "marker unavailable failure drifted",
    )
    for field in (
        "schema_failure_must_prevent_query_execution",
        "marker_failure_must_not_return_draft_payload",
        "connection_failure_must_not_return_database_detail",
        "runtime_boundary_failure_must_not_fallback_to_memory_dev",
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
        strategy.get("status") == "schema_marker_contract_implementation_entry_review_checker_only_no_runtime",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_schema_marker_contract_implementation_entry_review_checker",
        "workflow_saved_draft_schema_marker_migration_runner_readiness_refresh_checker",
        "workflow_saved_draft_repository_mode_runtime_boundary_review_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field, value in strategy.items():
        if field.startswith("does_not_"):
            require(value is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = (
        "checks/control_plane/"
        "check-workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1.py"
    )
    current_checker = (
        "checks/control_plane/"
        "check-workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "schema marker migration runner refresh checker missing from check-repo")
    require(current_checker in check_repo, "schema marker contract entry review checker missing from check-repo")
    require(next_checker in check_repo, "product surface checker missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "schema marker contract entry review checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_entry_boundary(fixture)
    assert_gates_candidates_and_contract(fixture)
    assert_failure_and_safety(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft schema marker contract implementation entry review v1 checks passed.")


if __name__ == "__main__":
    main()
