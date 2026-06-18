#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-entry-review-v1.json"
)
CONTRACT_RUNNER_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.json"
)
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json"
)
SCHEMA_MANIFEST_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-manifest-v1.json"
)
SELECTOR_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json"
)
SCHEMA_MATERIALIZATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json"
)
PRODUCTION_AUTH_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-production-auth-readiness-v1.json"
)
ADAPTER_SMOKE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-readiness-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
IMPLEMENTATION_TASK_CARD_PATH = (
    "docs/task-cards/workflow-saved-draft-repository-adapter-implementation-v1-plan.md"
)

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-repository-contract-smoke-runner-implementation-v1": (
        CONTRACT_RUNNER_FIXTURE_PATH,
        "draft_repository_contract_smoke_runner_implemented",
    ),
    "workflow-saved-draft-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "draft_repository_adapter_implementation_plan_defined",
    ),
    "workflow-saved-draft-schema-artifact-manifest-v1": (
        SCHEMA_MANIFEST_FIXTURE_PATH,
        "draft_schema_artifact_manifest_defined",
    ),
    "workflow-saved-draft-store-selector-smoke-v1": (
        SELECTOR_IMPLEMENTATION_FIXTURE_PATH,
        "draft_store_selector_smoke_implemented",
    ),
    "workflow-saved-draft-schema-artifact-materialization-v1": (
        SCHEMA_MATERIALIZATION_FIXTURE_PATH,
        "draft_schema_artifact_materialized_static",
    ),
    "workflow-saved-draft-production-auth-readiness-v1": (
        PRODUCTION_AUTH_READINESS_FIXTURE_PATH,
        "draft_production_auth_readiness_defined",
    ),
    "workflow-saved-draft-adapter-smoke-readiness-v1": (
        ADAPTER_SMOKE_READINESS_FIXTURE_PATH,
        "draft_adapter_smoke_readiness_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "repository_adapter_implementation_task_card_created",
    "repository_interface_implemented",
    "repository_adapter_implemented",
    "repository_adapter_ready",
    "repository_mode_enabled",
    "durable_store_implemented",
    "durable_persistence_ready",
    "database_query_ready",
    "database_connection_ready",
    "database_migration_ready",
    "schema_version_table_created",
    "sql_migration_created",
    "migration_runner_implemented",
    "production_auth_ready",
    "radish_oidc_ready",
    "token_validation_ready",
    "auth_middleware_ready",
    "workspace_membership_adapter_ready",
    "adapter_smoke_ready",
    "adapter_smoke_executed",
    "production_api_consumer_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_GATE_IDS = {
    "repository_contract_smoke_runner_consumed",
    "repository_adapter_plan_consumed",
    "schema_artifact_manifest_consumed",
    "store_selector_implementation_consumed",
    "schema_artifact_materialization_consumed",
    "production_auth_readiness_consumed",
    "adapter_smoke_readiness_consumed",
    "repository_adapter_entry_review_defined",
    "repository_interface_candidate_reviewed",
    "repository_adapter_candidate_reviewed",
    "adapter_unit_tests_candidate_reviewed",
    "adapter_smoke_execution_gate",
    "production_auth_runtime_gate",
    "production_api_gate",
    "database_execution_gate",
    "migration_runner_gate",
    "no_repository_adapter_implementation_artifacts_leaked",
}
READY_CANDIDATES = {
    "repository_interface",
    "repository_adapter",
    "adapter_unit_tests",
}
BLOCKED_CANDIDATES = {
    "adapter_smoke_fixture",
    "repository_store_mode_enablement",
    "production_route_consumer",
}
EXPECTED_OPERATIONS = {
    "SaveWorkflowDraftRecord": "workflow_drafts:write",
    "ReadWorkflowDraftRecord": "workflow_drafts:read",
    "ListWorkflowDraftRecords": "workflow_drafts:read",
}
EXPECTED_BINDING_FIELDS = {
    "tenant_ref",
    "workspace_id",
    "application_id",
    "actor_subject_ref",
    "owner_subject_ref",
    "scope_grants",
}
EXPECTED_FAILURE_CODES = {
    "draft_scope_denied",
    "draft_not_found",
    "draft_schema_version_unsupported",
    "draft_payload_invalid",
    "draft_version_conflict",
    "draft_store_unavailable",
    "draft_store_contract_mismatch",
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_auth_context_contract_mismatch",
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_identity_context_missing",
    "draft_tenant_binding_missing",
    "draft_workspace_membership_denied",
    "draft_scope_grant_missing",
}
EXPECTED_PRESERVED_FAILURE_CODES = {
    "draft_version_conflict",
    "draft_scope_denied",
    "draft_not_found",
    "draft_store_unavailable",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "repository_write_count=0",
    "database_write_count=0",
    "database_query_count=0",
    "adapter_smoke_execution_count=0",
    "issuer_network_call_count=0",
    "token_validation_call_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "repository_write",
    "database_write",
    "database_query",
    "schema_migration_apply",
    "adapter_smoke_execution",
    "issuer_network_call",
    "token_validation_runtime_call",
    "workflow_execution",
    "confirmation_decision",
    "business_writeback",
    "replay_execution",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    missing = sorted(set(EXPECTED_DEPENDENCIES) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for expected_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == expected_id, f"dependency {expected_id} id drifted")
        require(
            slice_info.get("status") == expected_status,
            f"dependency {expected_id} status must be {expected_status}",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "workflow_saved_draft_repository_adapter_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-repository-adapter-implementation-entry-review-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_repository_adapter_implementation_entry_review_defined",
        "repository adapter implementation entry review status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("repository_adapter_entry_boundary") or {}
    require(
        boundary.get("status") == "repository_adapter_implementation_entry_review_defined_no_runtime_change",
        "entry boundary status drifted",
    )
    require(
        boundary.get("decision") == "repository_adapter_implementation_entry_ready_for_next_task",
        "entry boundary decision drifted",
    )
    require(
        boundary.get("current_development_mode") == "entry_review_only_no_runtime_change",
        "development mode drifted",
    )
    require(boundary.get("selected_implementation_track_now") == "none", "selected track must remain none")
    require(
        boundary.get("future_implementation_task_card")
        == "docs/task-cards/workflow-saved-draft-repository-adapter-implementation-v1-plan.md",
        "future implementation task card path drifted",
    )
    for field in (
        "implementation_task_card_created_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_created_in_this_slice",
        "repository_adapter_tests_created_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_connection_allowed_in_this_slice",
        "sql_files_created_in_this_slice",
        "schema_version_table_created_in_this_slice",
        "migration_runner_created_in_this_slice",
        "adapter_smoke_fixture_created_in_this_slice",
        "adapter_smoke_checker_created_in_this_slice",
        "repository_mode_enablement_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "auth_middleware_allowed_in_this_slice",
        "membership_adapter_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "publish_or_run_allowed_in_this_slice",
        "executor_allowed_in_this_slice",
        "writeback_or_replay_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gate_matrix(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("entry_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "entry gate ids drifted")
    for gate_id in (
        "repository_contract_smoke_runner_consumed",
        "repository_adapter_plan_consumed",
        "schema_artifact_manifest_consumed",
        "store_selector_implementation_consumed",
        "repository_adapter_entry_review_defined",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must be satisfied")
    require(
        gates["schema_artifact_materialization_consumed"].get("status") == "satisfied_static",
        "schema artifact materialization must be satisfied_static",
    )
    for gate_id in ("production_auth_readiness_consumed", "adapter_smoke_readiness_consumed"):
        require(gates[gate_id].get("status") == "satisfied_for_entry_review", f"{gate_id} status drifted")
    for gate_id in (
        "repository_interface_candidate_reviewed",
        "repository_adapter_candidate_reviewed",
        "adapter_unit_tests_candidate_reviewed",
    ):
        require(gates[gate_id].get("status") == "ready_for_next_task", f"{gate_id} must be ready")
    for gate_id in (
        "adapter_smoke_execution_gate",
        "production_auth_runtime_gate",
        "production_api_gate",
        "database_execution_gate",
        "migration_runner_gate",
    ):
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
    require(
        gates["no_repository_adapter_implementation_artifacts_leaked"].get("status") == "required_now",
        "artifact leak guard must be required now",
    )


def assert_candidates(fixture: dict[str, Any]) -> None:
    candidates = {
        str(candidate.get("candidate_id") or ""): candidate
        for candidate in fixture.get("repository_adapter_entry_candidates") or []
        if isinstance(candidate, dict)
    }
    require(set(candidates) == READY_CANDIDATES | BLOCKED_CANDIDATES, "candidate ids drifted")
    for candidate_id in READY_CANDIDATES:
        candidate = candidates[candidate_id]
        require(candidate.get("entry_decision") == "ready_for_next_task", f"{candidate_id} decision drifted")
        require(candidate.get("trigger_satisfied_now") is True, f"{candidate_id} trigger must be satisfied")
        require(candidate.get("task_card_allowed_next") is True, f"{candidate_id} task card must be allowed next")
        require(
            candidate.get("implementation_artifacts_allowed_now") is False,
            f"{candidate_id} artifacts must not be allowed now",
        )
        require(candidate.get("future_artifacts"), f"{candidate_id} must name future artifacts")
        require(candidate.get("remaining_preconditions"), f"{candidate_id} must name remaining preconditions")
    for candidate_id in BLOCKED_CANDIDATES:
        candidate = candidates[candidate_id]
        require(candidate.get("entry_decision") == "blocked", f"{candidate_id} must remain blocked")
        require(candidate.get("trigger_satisfied_now") is False, f"{candidate_id} trigger must remain false")
        require(candidate.get("task_card_allowed_next") is False, f"{candidate_id} task card must remain blocked")
        require(
            candidate.get("implementation_artifacts_allowed_now") is False,
            f"{candidate_id} artifacts must not be allowed now",
        )
        require(candidate.get("current_blockers"), f"{candidate_id} must name blockers")


def assert_operation_matrix(fixture: dict[str, Any]) -> None:
    operations = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_entry_matrix") or []
        if isinstance(row, dict)
    }
    require(set(operations) == set(EXPECTED_OPERATIONS), "operation entry matrix must cover save/read/list")
    for operation, expected_scope in EXPECTED_OPERATIONS.items():
        row = operations[operation]
        require(row.get("required_scope") == expected_scope, f"{operation} scope drifted")
        require(
            row.get("entry_review_conclusion") == "ready_for_next_repository_adapter_task",
            f"{operation} entry conclusion drifted",
        )
        require(
            set(row.get("required_binding_fields") or []) == EXPECTED_BINDING_FIELDS,
            f"{operation} binding fields drifted",
        )
        for flag in (
            "implementation_allowed_now",
            "database_write_allowed_now",
            "adapter_smoke_allowed_now",
            "production_api_allowed_now",
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "fallback_to_test_auth_allowed",
            "side_effect_allowed",
        ):
            require(row.get(flag) is False, f"{operation} {flag} must remain false")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    failure_mapping = fixture.get("failure_mapping") or {}
    require(
        failure_mapping.get("status") == "repository_adapter_entry_fail_closed_required",
        "failure mapping status drifted",
    )
    require(
        set(failure_mapping.get("required_failure_codes") or []) == EXPECTED_FAILURE_CODES,
        "required failure codes drifted",
    )
    require(
        set(failure_mapping.get("preserve_existing_failure_codes") or [])
        == EXPECTED_PRESERVED_FAILURE_CODES,
        "preserved failure codes drifted",
    )
    require(failure_mapping.get("forbidden_failure_rewrites"), "forbidden failure rewrites must be listed")


def assert_no_fallback_and_side_effects(fixture: dict[str, Any]) -> None:
    fallback_policy = fixture.get("no_fallback_policy") or {}
    for key, value in fallback_policy.items():
        require(value is False, f"{key} must remain false")
    side_effect_policy = fixture.get("no_side_effect_policy") or {}
    require(
        set(side_effect_policy.get("side_effect_counters_must_remain") or [])
        == EXPECTED_SIDE_EFFECT_COUNTERS,
        "side effect counters drifted",
    )
    require(
        set(side_effect_policy.get("forbidden_side_effects") or []) == EXPECTED_FORBIDDEN_SIDE_EFFECTS,
        "forbidden side effects drifted",
    )


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("implementation_artifact_guard") or {}
    require(
        guard.get("status") == "forbid_repository_adapter_implementation_artifacts",
        "artifact guard status drifted",
    )
    task_card_path = REPO_ROOT / IMPLEMENTATION_TASK_CARD_PATH
    if task_card_path.exists():
        task_card = task_card_path.read_text(encoding="utf-8")
        require(
            "draft_repository_adapter_implementation_task_card_defined" in task_card,
            "implementation task card must remain task-card-defined only",
        )
        require(
            "不启用 `repository` store mode" in task_card,
            "implementation task card must preserve repository mode stop line",
        )
    for relative_path in guard.get("future_files_must_not_exist") or []:
        require(
            not (REPO_ROOT / relative_path).exists(),
            f"future implementation artifact exists too early: {relative_path}",
        )
    source_paths = guard.get("source_files_to_scan") or []
    require(source_paths, "source files to scan must be listed")
    combined_source = ""
    for relative_path in source_paths:
        source_path = REPO_ROOT / relative_path
        require(source_path.exists(), f"source file missing: {relative_path}")
        combined_source += source_path.read_text(encoding="utf-8")
    for literal in guard.get("future_literals_must_not_appear_in_source") or []:
        require(literal not in combined_source, f"future implementation literal appeared too early: {literal}")


def assert_doc_references(fixture: dict[str, Any]) -> None:
    for requirement in fixture.get("required_doc_references") or []:
        path = str(requirement.get("path") or "")
        require(path, "doc reference path missing")
        content = read(path)
        for needle in requirement.get("must_contain") or []:
            require(needle in content, f"{path} must contain {needle!r}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "static_checker_only", "testing strategy status drifted")
    required_checks = set(strategy.get("required_checks") or [])
    expected_checks = {
        "workflow_saved_draft_repository_adapter_implementation_entry_review_checker",
        "workflow_saved_draft_production_auth_readiness_checker",
        "workflow_saved_draft_schema_artifact_materialization_checker",
        "workflow_saved_draft_store_selector_smoke_checker",
        "workflow_saved_draft_adapter_smoke_readiness_checker",
        "workflow_saved_draft_repository_adapter_implementation_plan_checker",
        "./scripts/check-repo.sh --fast",
    }
    require(expected_checks <= required_checks, "testing strategy missing required checks")
    for flag in (
        "does_not_start_service",
        "does_not_connect_database",
        "does_not_create_repository_interface",
        "does_not_create_repository_adapter",
        "does_not_create_adapter_smoke",
        "does_not_validate_token",
        "does_not_create_production_api",
        "does_not_create_sql_or_migration",
    ):
        require(strategy.get(flag) is True, f"{flag} must be true")


def assert_check_repo_order() -> None:
    content = CHECK_REPO_PATH.read_text(encoding="utf-8")
    production_auth_checker = "checks/control_plane/check-workflow-saved-draft-production-auth-readiness-v1.py"
    current_checker = (
        "checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-entry-review-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(current_checker in content, "repository adapter implementation entry review checker not in check-repo")
    require(
        content.index(production_auth_checker) < content.index(current_checker) < content.index(next_checker),
        "repository adapter implementation entry review checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_gate_matrix(fixture)
    assert_candidates(fixture)
    assert_operation_matrix(fixture)
    assert_failure_mapping(fixture)
    assert_no_fallback_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_doc_references(fixture)
    assert_testing_strategy(fixture)
    assert_check_repo_order()
    print("workflow saved draft repository adapter implementation entry review v1 checks passed.")


if __name__ == "__main__":
    main()
