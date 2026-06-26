#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-database-secret-resolver-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-readiness-v1.json",
        "draft_database_secret_resolver_readiness_defined",
    ),
    "workflow-saved-draft-database-connection-provider-implementation-entry-review-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.json",
        "draft_database_connection_provider_implementation_entry_review_defined",
    ),
    "workflow-saved-draft-repository-mode-enablement-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-repository-mode-enablement-v1.json",
        "draft_repository_mode_enablement_review_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}
EXPECTED_SECRET_REFERENCE_DEPENDENCY = "production-secret-reference-basic"
EXPECTED_FORBIDDEN_CLAIMS = {
    "secret_resolver_implementation_task_card_created",
    "secret_resolver_implemented",
    "fake_resolver_implemented",
    "secret_value_resolved",
    "secret_value_storage_ready",
    "cloud_secret_service_ready",
    "production_secret_backend_ready",
    "database_connection_provider_ready",
    "database_connection_ready",
    "database_driver_selected",
    "database_driver_imported",
    "connection_factory_implemented",
    "connection_smoke_ready",
    "database_query_executor_ready",
    "schema_marker_read_write_ready",
    "sql_migration_created",
    "migration_runner_implemented",
    "repository_store_mode_enabled",
    "durable_persistence_ready",
    "oidc_middleware_ready",
    "token_validation_ready",
    "workspace_membership_adapter_ready",
    "production_api_consumer_ready",
    "workflow_executor_ready",
    "production_ready",
}
EXPECTED_GATE_STATUS = {
    "secret_resolver_readiness_consumed": "satisfied_readiness_only",
    "connection_provider_entry_review_consumed": "satisfied_blocked_entry",
    "repository_mode_enablement_review_consumed": "satisfied_disabled",
    "secret_reference_manifest_consumed": "satisfied_reference_only_resolver_disabled",
    "production_secret_backend_implementation_consumed": "not_satisfied_resolver_not_started",
    "database_secret_resolver_implementation_entry_review_defined": "satisfied",
    "resolver_interface_implementation_gate": "blocked",
    "offline_fake_resolver_gate": "blocked",
    "sanitized_diagnostics_runtime_gate": "blocked",
    "environment_binding_runtime_gate": "blocked",
    "cloud_secret_backend_call_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_runtime_gate": "blocked",
    "artifact_guard": "required_now",
}
EXPECTED_CANDIDATES = {
    "resolver_interface_implementation",
    "offline_fake_resolver",
    "sanitized_diagnostics_runtime",
    "environment_binding_enforcement",
    "connection_factory_handoff",
    "repository_mode_integration",
}
EXPECTED_FAILURE_CODES = {
    "draft_store_unavailable",
    "repository_store_disabled",
    "invalid_draft_store_mode",
}
EXPECTED_DIAGNOSTIC_CODES = {
    "configured",
    "missing_secret_ref",
    "secret_backend_disabled",
    "secret_resolver_unavailable",
    "secret_resolution_denied",
    "credential_missing",
    "environment_mismatch",
    "repository_mode_disabled",
    "invalid_store_mode",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "secret_resolver_call_count=0",
    "fake_resolver_call_count=0",
    "secret_value_read_count=0",
    "cloud_secret_call_count=0",
    "database_connection_count=0",
    "driver_open_count=0",
    "connection_smoke_count=0",
    "sql_execution_count=0",
    "schema_marker_read_count=0",
    "schema_marker_write_count=0",
    "repository_read_count=0",
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
    require(EXPECTED_SECRET_REFERENCE_DEPENDENCY in declared, "missing production secret reference dependency")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"{dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"{dependency_id} status drifted")

    secret_backend = load_json(
        REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
    )
    implementation_target = secret_backend.get("implementation_target") or {}
    require(
        implementation_target.get("resolver_implementation_status") == "not_started",
        "production secret resolver must remain not_started",
    )
    require(
        implementation_target.get("default_runtime_state") == "disabled_until_explicit_secret_backend_task",
        "production secret backend runtime state drifted",
    )
    require(
        implementation_target.get("fast_baseline_policy") == "offline_no_real_credentials_no_cloud_calls",
        "production secret backend fast baseline policy drifted",
    )

    secret_reference = load_json(REPO_ROOT / "scripts/checks/fixtures/production-secret-reference-basic.json")
    require(secret_reference.get("scope") == "secret_reference_only", "secret reference scope drifted")
    policy = secret_reference.get("policy") or {}
    require(policy.get("stores_secret_values") is False, "secret reference must not store secret values")
    require(policy.get("resolver_enabled") is False, "secret resolver must remain disabled")
    require(policy.get("cloud_calls_allowed") is False, "cloud calls must remain disabled")
    require(policy.get("production_secret_backend_ready") is False, "production secret backend must not be ready")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "workflow_saved_draft_database_secret_resolver_implementation_entry_review_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_database_secret_resolver_implementation_entry_review_defined",
        "status drifted",
    )
    require(
        slice_info.get("task_card")
        == "docs/task-cards/workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1-plan.md",
        "task card drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_entry_boundary") or {}
    require(
        boundary.get("status") == "database_secret_resolver_implementation_entry_review_defined_no_entry_opened",
        "entry boundary status drifted",
    )
    require(boundary.get("decision") == "secret_resolver_implementation_entry_not_opened", "entry decision drifted")
    require(boundary.get("current_development_mode") == "entry_review_only_no_secret_resolution", "mode drifted")
    require(boundary.get("selected_implementation_track_now") == "none", "selected track must remain none")
    require(
        boundary.get("future_implementation_task_card")
        == "docs/task-cards/workflow-saved-draft-database-secret-resolver-implementation-v1-plan.md",
        "future implementation task card path drifted",
    )
    require(
        boundary.get("future_secret_ref_config") == "RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_SECRET_REF",
        "future secret ref config drifted",
    )
    for field in (
        "implementation_task_card_created_in_this_slice",
        "secret_resolver_file_created_in_this_slice",
        "fake_resolver_file_created_in_this_slice",
        "secret_value_resolution_allowed_in_this_slice",
        "secret_value_storage_allowed_in_this_slice",
        "cloud_secret_call_allowed_in_this_slice",
        "sanitized_diagnostics_runtime_allowed_in_this_slice",
        "environment_binding_runtime_allowed_in_this_slice",
        "connection_factory_handoff_allowed_in_this_slice",
        "database_connection_provider_file_created_in_this_slice",
        "database_driver_import_allowed_in_this_slice",
        "connection_factory_allowed_in_this_slice",
        "connection_smoke_allowed_in_this_slice",
        "database_query_executor_allowed_in_this_slice",
        "schema_marker_allowed_in_this_slice",
        "sql_files_created_in_this_slice",
        "migration_runner_allowed_in_this_slice",
        "service_startup_resolution_allowed_in_this_slice",
        "http_route_resolution_allowed_in_this_slice",
        "selector_runtime_resolution_allowed_in_this_slice",
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


def assert_gates_candidates_and_diagnostics(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "entry_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "entry gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    candidates = rows_by_id(fixture, "implementation_entry_candidates", "candidate_id")
    require(set(candidates) == EXPECTED_CANDIDATES, "implementation entry candidates drifted")
    for candidate_id, candidate in candidates.items():
        require(candidate.get("entry_decision") == "blocked", f"{candidate_id} must remain blocked")
        require(candidate.get("trigger_satisfied_now") is False, f"{candidate_id} trigger must remain false")
        require(
            candidate.get("implementation_artifacts_allowed_now") is False,
            f"{candidate_id} artifacts must remain forbidden",
        )
        require(candidate.get("current_blockers"), f"{candidate_id} blockers must be listed")

    diagnostics = fixture.get("sanitized_diagnostics_policy") or {}
    require(
        diagnostics.get("status") == "diagnostics_contract_defined_no_runtime_emission",
        "diagnostics status drifted",
    )
    missing_diagnostics = sorted(EXPECTED_DIAGNOSTIC_CODES - set(diagnostics.get("allowed_diagnostic_codes") or []))
    require(not missing_diagnostics, f"missing diagnostic codes: {missing_diagnostics}")
    forbidden = set(diagnostics.get("forbidden_diagnostic_fields") or [])
    for field in {"secret_value", "password", "token", "api_key", "provider_url", "dsn"}:
        require(field in forbidden, f"diagnostics must forbid {field}")
    require(
        diagnostics.get("must_use_request_or_audit_ref_for_operator_debugging") is True,
        "operator debugging must use request/audit ref",
    )


def assert_failure_and_safety(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "database_secret_resolver_implementation_entry_fail_closed_required",
        "failure mapping status drifted",
    )
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    for key in (
        "missing_secret_ref_failure_code",
        "secret_backend_disabled_failure_code",
        "secret_resolver_unavailable_failure_code",
        "secret_resolution_denied_failure_code",
        "credential_missing_failure_code",
        "environment_mismatch_failure_code",
    ):
        require(mapping.get(key) == "draft_store_unavailable", f"{key} must map to draft_store_unavailable")
    require(
        mapping.get("repository_mode_still_disabled_failure_code") == "repository_store_disabled",
        "repository mode failure drifted",
    )
    require(
        mapping.get("unknown_store_mode_failure_code") == "invalid_draft_store_mode",
        "unknown store mode failure drifted",
    )
    for field in (
        "secret_failure_must_not_return_raw_secret",
        "secret_failure_must_not_return_dsn",
        "secret_failure_must_not_return_provider_url",
        "secret_failure_must_not_return_database_error_detail",
        "resolver_failure_must_prevent_connection_provider",
        "resolver_failure_must_prevent_query_execution",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

    source = "\n".join(
        [
            read("services/platform/internal/httpapi/workflow_saved_draft.go"),
            read("services/platform/internal/httpapi/workflow_saved_draft_store_selector.go"),
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
        strategy.get("status") == "database_secret_resolver_implementation_entry_review_checker_only_no_resolver",
        "testing strategy status drifted",
    )
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_database_secret_resolver_implementation_entry_review_checker",
        "workflow_saved_draft_database_secret_resolver_readiness_checker",
        "production_ops_secret_backend_implementation_readiness_checker",
        "production_secret_reference_contract_checker",
        "./scripts/check-repo.sh --fast",
        "./scripts/check-repo.sh",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_enable_repository_store_mode",
        "does_not_resolve_secret",
        "does_not_create_fake_resolver",
        "does_not_store_secret_value",
        "does_not_call_cloud_secret_service",
        "does_not_connect_database",
        "does_not_import_database_driver",
        "does_not_open_database_driver",
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
    previous_checker = "checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py"
    current_checker = (
        "checks/control_plane/"
        "check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "database secret resolver readiness checker missing from check-repo")
    require(current_checker in check_repo, "database secret resolver implementation entry checker missing from check-repo")
    require(next_checker in check_repo, "product surface checker missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "database secret resolver implementation entry checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_gates_candidates_and_diagnostics(fixture)
    assert_failure_and_safety(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    print("workflow saved draft database secret resolver implementation entry review v1 checks passed.")


if __name__ == "__main__":
    main()
