#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-database-secret-resolver-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-readiness-v1.json",
        "draft_database_secret_resolver_readiness_defined",
    ),
    "workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.json",
        "draft_database_secret_resolver_implementation_entry_review_defined",
    ),
    "workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2.json",
        "draft_database_connection_provider_implementation_entry_refresh_v2_defined",
    ),
    "workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.json",
        "draft_schema_marker_runtime_dependency_refresh_defined",
    ),
    "workflow-saved-draft-repository-mode-runtime-boundary-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-runtime-boundary-review-v1.json",
        "draft_repository_mode_runtime_boundary_review_defined",
    ),
    "radish-oidc-token-membership-upstream-evidence-refresh-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json",
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
    "production-secret-backend-fake-resolver-runtime-implementation-v1": (
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json",
        "fake_resolver_runtime_test_only_implemented",
    ),
    "production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.json",
        "real_resolver_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
        "resolver_backend_profile_selection_readiness_defined",
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.json",
        "credential_handle_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.json",
        "operator_approval_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ),
}
EXPECTED_SECRET_REFERENCE_DEPENDENCY = "production-secret-reference-basic"
EXPECTED_FORBIDDEN_CLAIMS = {
    "database_secret_resolver_implementation_task_card_created",
    "database_secret_resolver_runtime_created",
    "secret_value_resolved",
    "secret_value_storage_ready",
    "production_secret_backend_ready",
    "production_resolver_runtime_task_card_created",
    "production_resolver_runtime_created",
    "credential_payload_created",
    "credential_handle_runtime_ready",
    "credential_handle_created",
    "operator_approval_runtime_ready",
    "audit_store_runtime_ready",
    "backend_health_runtime_ready",
    "no_secret_leakage_smoke_runtime_ready",
    "cloud_secret_service_ready",
    "database_connection_provider_ready",
    "database_connection_ready",
    "database_driver_imported",
    "connection_factory_implemented",
    "connection_smoke_ready",
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
    "status": "database_secret_resolver_runtime_dependency_refresh_defined_no_task_card",
    "decision": "database_secret_resolver_runtime_still_blocked_before_implementation_task_card",
    "current_development_mode": "dependency_refresh_only_no_secret_resolution",
    "selected_implementation_track_now": "none",
    "future_database_secret_resolver_task_card": (
        "docs/task-cards/workflow-saved-draft-database-secret-resolver-implementation-v1-plan.md"
    ),
    "future_production_resolver_runtime_task_card": (
        "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-v1-plan.md"
    ),
}
EXPECTED_TRUE_BOUNDARY_FLAGS = {
    "secret_resolver_static_contract_consumed",
    "production_real_resolver_entry_refresh_consumed",
    "credential_handle_runtime_entry_review_consumed",
    "operator_approval_runtime_entry_review_consumed",
    "audit_store_runtime_entry_refresh_consumed",
    "backend_health_runtime_entry_review_consumed",
    "no_leakage_runtime_entry_review_consumed",
    "test_only_fake_resolver_status_consumed",
}
EXPECTED_FALSE_BOUNDARY_FLAGS = {
    "implementation_task_card_created_in_this_slice",
    "production_resolver_runtime_task_card_created_in_this_slice",
    "database_secret_resolver_runtime_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "secret_value_resolution_allowed_in_this_slice",
    "secret_value_storage_allowed_in_this_slice",
    "cloud_secret_call_allowed_in_this_slice",
    "credential_payload_created_in_this_slice",
    "credential_handle_created_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "operator_approval_runtime_created_in_this_slice",
    "operator_approval_runtime_executed_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "backend_health_runtime_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "no_secret_leakage_smoke_runtime_created_in_this_slice",
    "database_connection_provider_created_in_this_slice",
    "database_connection_opened_in_this_slice",
    "database_driver_imported_in_this_slice",
    "schema_marker_runtime_created_in_this_slice",
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
    "database_secret_resolver_readiness_consumed": "satisfied_static_contract_no_runtime",
    "database_secret_resolver_entry_review_consumed": "blocked_before_implementation_task_card",
    "connection_provider_entry_refresh_v2_consumed": "blocked_before_implementation_task_card",
    "schema_marker_runtime_dependency_refresh_consumed": "blocked_before_implementation_task_card",
    "repository_mode_runtime_boundary_review_consumed": "blocked_before_task_card",
    "radish_oidc_upstream_evidence_refresh_consumed": "satisfied_static_contract_no_auth_runtime",
    "secret_reference_manifest_consumed": "reference_only_resolver_disabled",
    "test_only_fake_resolver_runtime_consumed": "implemented_test_only_disabled_by_default",
    "production_real_resolver_entry_refresh_consumed": "blocked_before_runtime_task_card",
    "credential_handle_runtime_entry_review_consumed": "blocked_before_runtime_task_card",
    "operator_approval_runtime_entry_review_consumed": "blocked_before_runtime_task_card",
    "audit_store_runtime_entry_refresh_consumed": "blocked_before_runtime_task_card",
    "backend_health_runtime_entry_review_consumed": "blocked_before_runtime_task_card",
    "no_leakage_smoke_runtime_entry_review_consumed": "blocked_before_runtime_task_card",
    "database_secret_resolver_runtime_task_card_gate": "blocked_before_implementation_task_card",
    "no_database_secret_resolver_runtime_artifacts_leaked": "required_now",
}
EXPECTED_DEPENDENCY_STATUS = {
    "saved_draft_secret_resolver_contract": "static_contract_defined_no_runtime",
    "secret_reference_manifest": "reference_only_resolver_disabled",
    "production_resolver_runtime": "blocked_no_production_resolver_runtime",
    "credential_handle_runtime": "blocked_no_credential_handle_runtime",
    "operator_approval_runtime": "blocked_no_operator_approval_runtime",
    "audit_store_runtime": "blocked_no_audit_store_runtime",
    "backend_health_runtime": "blocked_no_backend_health_runtime",
    "no_leakage_smoke_runtime": "blocked_no_no_leakage_smoke_runtime",
    "test_only_fake_resolver": "implemented_test_only_disabled_cannot_unlock_production",
    "connection_provider_dependency": "blocked_no_connection_provider",
    "schema_marker_dependency": "blocked_no_schema_marker_runtime",
    "repository_mode_dependency": "blocked_no_repository_mode_runtime",
    "auth_membership_dependency": "static_upstream_evidence_defined_no_auth_runtime",
    "sanitized_diagnostics_dependency": "static_contract_defined_no_runtime_emission",
    "environment_binding_dependency": "static_contract_defined_no_runtime_enforcement",
}
EXPECTED_RUNTIME_BLOCKERS = {
    "database_secret_resolver_runtime_not_created",
    "production_resolver_runtime_task_card_not_created",
    "production_resolver_runtime_not_created",
    "credential_handle_runtime_not_created",
    "operator_approval_runtime_not_created",
    "audit_store_runtime_not_created",
    "backend_health_runtime_not_created",
    "no_secret_leakage_smoke_runtime_not_created",
    "cloud_secret_service_not_selected",
    "connection_provider_not_created",
    "schema_marker_runtime_not_created",
    "repository_mode_runtime_disabled",
    "auth_middleware_not_created",
    "membership_adapter_not_created",
}
EXPECTED_FUTURE_TASK_REQUIREMENTS = {
    "secret_resolver_interface_owner",
    "secret_ref_environment_purpose_audit_policy_binding",
    "production_resolver_runtime_backend_profile_evidence",
    "opaque_credential_handle_runtime_lifecycle",
    "operator_approval_runtime_dual_control_evidence",
    "audit_store_runtime_delivery_idempotency",
    "backend_health_runtime_check_failure_mapping",
    "no_leakage_smoke_runtime_artifact_scan",
    "connection_provider_schema_marker_repository_auth_handoff",
    "negative_resolver_smoke_matrix",
}
EXPECTED_FAILURE_CODES = {
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_store_unavailable",
    "draft_audit_context_missing",
    "draft_store_migration_unavailable",
    "draft_auth_context_contract_mismatch",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "secret_resolver_call_count=0",
    "fake_resolver_call_count=0",
    "production_resolver_call_count=0",
    "real_secret_read_count=0",
    "environment_secret_read_count=0",
    "secret_value_read_count=0",
    "cloud_secret_call_count=0",
    "network_call_count=0",
    "credential_payload_created_count=0",
    "credential_handle_created_count=0",
    "credential_handle_runtime_created_count=0",
    "operator_approval_runtime_execution_count=0",
    "audit_store_runtime_created_count=0",
    "audit_writer_created_count=0",
    "audit_event_write_count=0",
    "backend_health_runtime_created_count=0",
    "backend_health_check_count=0",
    "no_secret_leakage_smoke_runtime_created_count=0",
    "database_connection_count=0",
    "driver_open_count=0",
    "sql_execution_count=0",
    "schema_marker_read_count=0",
    "schema_marker_write_count=0",
    "repository_mode_enablement_count=0",
    "production_api_call_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_NEXT_CANDIDATES = {
    "token_validation_schema_auth_middleware_runtime_entry_review",
    "production_resolver_runtime_blocker_consolidation",
    "connection_provider_runtime_dependency_refresh",
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
    require(EXPECTED_SECRET_REFERENCE_DEPENDENCY in declared, "missing production secret reference dependency")

    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"{dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"{dependency_id} status drifted")

    readiness = load_json(REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json")
    implementation_target = readiness.get("implementation_target") or {}
    expected_readiness_status = {
        "production_secret_backend_status": "not_satisfied",
        "resolver_implementation_status": "not_started",
        "resolver_runtime_status": "not_created",
        "production_resolver_runtime_task_card_status": "not_created",
        "fake_resolver_runtime_status": "implemented_test_only_disabled_by_default",
        "real_resolver_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "credential_handle_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "operator_approval_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "audit_store_runtime_implementation_entry_refresh_v3_status": "blocked_before_runtime_task_card",
        "resolver_backend_health_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_status": (
            "blocked_before_runtime_task_card"
        ),
    }
    for field, expected in expected_readiness_status.items():
        require(implementation_target.get(field) == expected, f"{field} drifted")

    secret_reference = load_json(REPO_ROOT / "scripts/checks/fixtures/production-secret-reference-basic.json")
    require(secret_reference.get("scope") == "secret_reference_only", "secret reference scope drifted")
    policy = secret_reference.get("policy") or {}
    for field in ("stores_secret_values", "resolver_enabled", "cloud_calls_allowed", "production_secret_backend_ready"):
        require(policy.get(field) is False, f"secret reference policy {field} must remain false")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "workflow_saved_draft_database_secret_resolver_runtime_dependency_refresh_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_database_secret_resolver_runtime_dependency_refresh_defined",
        "status drifted",
    )
    require(
        slice_info.get("task_card")
        == "docs/task-cards/workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1-plan.md",
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
        mapping.get("status") == "database_secret_resolver_runtime_dependency_refresh_fail_closed_required",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    expected_mappings = {
        "missing_secret_ref_failure_code": "draft_store_unavailable",
        "secret_backend_disabled_failure_code": "draft_store_unavailable",
        "production_resolver_runtime_missing_failure_code": "draft_store_unavailable",
        "credential_handle_runtime_missing_failure_code": "draft_store_unavailable",
        "operator_approval_runtime_missing_failure_code": "draft_store_unavailable",
        "audit_store_runtime_missing_failure_code": "draft_store_unavailable",
        "backend_health_runtime_missing_failure_code": "draft_store_unavailable",
        "no_leakage_smoke_runtime_missing_failure_code": "draft_store_unavailable",
        "connection_provider_missing_failure_code": "draft_store_unavailable",
        "audit_context_missing_failure_code": "draft_audit_context_missing",
        "marker_unavailable_failure_code": "draft_store_migration_unavailable",
        "repository_actor_context_failure_code": "draft_auth_context_contract_mismatch",
        "repository_mode_still_disabled_failure_code": "repository_store_disabled",
        "unknown_store_mode_failure_code": "invalid_draft_store_mode",
    }
    for field, expected in expected_mappings.items():
        require(mapping.get(field) == expected, f"{field} drifted")
    for field in (
        "secret_failure_must_prevent_connection_open",
        "secret_failure_must_not_return_raw_secret",
        "secret_failure_must_not_return_dsn",
        "test_only_fake_resolver_must_not_unlock_production",
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
        strategy.get("status") == "database_secret_resolver_runtime_dependency_refresh_checker_only_no_runtime",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_database_secret_resolver_runtime_dependency_refresh_checker",
        "workflow_saved_draft_database_secret_resolver_implementation_entry_review_checker",
        "workflow_saved_draft_schema_marker_runtime_dependency_refresh_checker",
        "production_secret_backend_real_resolver_runtime_implementation_entry_refresh_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_enable_repository_store_mode",
        "does_not_call_fake_resolver",
        "does_not_call_production_resolver",
        "does_not_read_secret",
        "does_not_create_credential_handle",
        "does_not_execute_operator_approval",
        "does_not_write_audit_event",
        "does_not_run_backend_health_check",
        "does_not_run_no_leakage_smoke",
        "does_not_connect_database",
        "does_not_import_database_driver",
        "does_not_run_sql",
        "does_not_read_schema_marker",
        "does_not_write_schema_marker",
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
    previous_checker = "checks/control_plane/check-workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.py"
    current_checker = (
        "checks/control_plane/check-workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    production_checker = "check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py"
    for checker in (previous_checker, current_checker, next_checker, production_checker):
        require(checker in check_repo, f"{checker} missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "database secret resolver runtime dependency refresh checker order drifted",
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
    print("workflow saved draft database secret resolver runtime dependency refresh checks passed.")


if __name__ == "__main__":
    main()
