#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
BLOCKER_MATRIX_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1"
SLICE_STATUS = "audit_store_storage_adapter_database_connection_lifecycle_readiness_defined"
READINESS_DECISION = "database_connection_lifecycle_readiness_defined_without_connection_runtime"
ENTRY_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_readiness"
NEXT_DEPENDENCY = "storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_readiness"
SELECTED_DRIVER_CANDIDATE = "github.com/jackc/pgx/v5"
SELECTED_DATABASE_ENGINE = "postgresql_compatible_append_only_relational_database"
SELECTED_PROVIDER_CLASS = "managed_postgresql_compatible_service"
MATRIX_BLOCKER_STATUS = (
    "storage_adapter_provider_account_resource_endpoint_readiness_defined_task_card_blocked"
)
CURRENT_ENTRY_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_provider_account_resource_endpoint_readiness"
)
CURRENT_NEXT_DEPENDENCY = "storage_adapter_provider_account_resource_endpoint_review"
CURRENT_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1"
)
FIXTURE_MATRIX_BLOCKER_STATUS_AFTER_READINESS = (
    "storage_adapter_concrete_managed_database_provider_selection_readiness_defined_task_card_blocked"
)
FIXTURE_MATRIX_BLOCKER_SOURCE_AFTER_READINESS = (
    "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-readiness-v1"
)
FIXTURE_NEXT_DEPENDENCY = "storage_adapter_concrete_managed_database_provider_selection_review"
FIXTURE_RUNTIME_TASK_CARD_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_readiness"
)

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_database_driver_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.json"
        ),
        "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.json"
        ),
        "audit_store_storage_adapter_table_schema_artifact_materialized",
    ),
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json"
        ),
        "audit_store_storage_adapter_metadata_contract_artifact_materialized",
    ),
    "production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.json"
        ),
        "audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-negative-leakage-runtime-scan-boundary-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-negative-leakage-runtime-scan-boundary-readiness-v1.json"
        ),
        "audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-runtime-blocker-matrix-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json",
        "audit_store_runtime_blocker_matrix_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}

EXPECTED_BOUNDARY = {
    "status": SLICE_STATUS,
    "readiness_decision": READINESS_DECISION,
    "entry_decision": ENTRY_DECISION,
    "next_dependency": NEXT_DEPENDENCY,
    "previous_review_status": "audit_store_storage_adapter_database_driver_selection_review_defined",
    "previous_runtime_task_decision": "storage_adapter_runtime_task_card_still_blocked_after_database_driver_selection_review",
    "selected_database_engine": SELECTED_DATABASE_ENGINE,
    "selected_provider_candidate_class": SELECTED_PROVIDER_CLASS,
    "selected_driver_candidate": SELECTED_DRIVER_CANDIDATE,
    "driver_selection_status": "selected_driver_candidate_without_runtime_import",
    "driver_package_status": "selected_candidate_reference_only",
    "driver_dependency_version_status": "not_pinned",
    "driver_import_status": "not_created",
    "database_driver_status": "selected_reference_only",
    "dsn_handoff_status": "secret_ref_only_dsn_handoff_defined",
    "database_dsn_status": "not_defined",
    "database_dsn_parser_status": "not_created",
    "database_connection_provider_status": "not_created",
    "database_connection_lifecycle_runtime_status": "not_created",
    "database_connection_status": "not_created",
    "connection_factory_status": "not_created",
    "pool_runtime_status": "not_created",
    "health_check_runtime_status": "not_created",
    "tls_role_environment_binding_status": "static_tls_role_environment_binding_defined",
    "pool_policy_status": "static_pool_policy_defined_without_pool_runtime",
    "timeout_budget_status": "static_timeout_budget_defined_without_runtime_timer",
    "retry_transaction_recovery_status": "static_recovery_policy_defined_without_runtime",
    "duplicate_replay_status": "duplicate_replay_fail_closed_policy_defined",
    "sanitized_diagnostics_status": "sanitized_diagnostics_allowlist_defined",
    "schema_marker_migration_handoff_status": "schema_marker_migration_handoff_defined",
    "offline_verification_status": "metadata_only_offline_verification_defined",
    "negative_leakage_scan_status": "metadata_only_negative_leakage_scan_boundary_defined",
    "rollout_rollback_status": "metadata_only_rollout_rollback_boundary_defined",
    "sql_migration_status": "not_created",
    "ddl_status": "not_created",
    "physical_table_schema_status": "not_created",
    "schema_marker_runtime_status": "not_created",
    "migration_runner_status": "not_created",
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "production_resolver_runtime_task_card_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "go_mod_changed_in_this_slice",
    "go_sum_changed_in_this_slice",
    "driver_import_added_in_this_slice",
    "driver_dependency_version_pinned_in_this_slice",
    "database_connection_provider_enabled",
    "database_connection_created_in_this_slice",
    "connection_lifecycle_runtime_created_in_this_slice",
    "connection_factory_created_in_this_slice",
    "pool_runtime_created_in_this_slice",
    "health_check_runtime_created_in_this_slice",
    "dsn_defined_in_this_slice",
    "dsn_parser_created_in_this_slice",
    "raw_dsn_material_allowed_in_this_slice",
    "sql_migration_created_in_this_slice",
    "ddl_created_in_this_slice",
    "physical_table_schema_created_in_this_slice",
    "schema_marker_runtime_created_in_this_slice",
    "migration_runner_created_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "production_resolver_runtime_task_card_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_GATES = {
    "secret_ref_only_dsn_handoff": "secret_ref_only_dsn_handoff_defined",
    "tls_role_environment_binding": "static_tls_role_environment_binding_defined",
    "pool_policy": "static_pool_policy_defined_without_pool_runtime",
    "timeout_budget": "static_timeout_budget_defined_without_runtime_timer",
    "retry_transaction_partial_write_recovery": "static_recovery_policy_defined_without_runtime",
    "duplicate_replay_fail_closed": "duplicate_replay_fail_closed_policy_defined",
    "sanitized_diagnostics": "sanitized_diagnostics_allowlist_defined",
    "schema_marker_migration_handoff": "schema_marker_migration_handoff_defined",
    "offline_verification": "metadata_only_offline_verification_defined",
    "negative_leakage_scan": "metadata_only_negative_leakage_scan_boundary_defined",
    "rollout_rollback_boundary": "metadata_only_rollout_rollback_boundary_defined",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_dependency_missing",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_secret_ref_handoff_missing",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_tls_role_environment_missing",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_pool_policy_missing",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_timeout_budget_missing",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_recovery_policy_missing",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_duplicate_replay_policy_missing",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_diagnostics_missing",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_schema_marker_handoff_missing",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_runtime_forbidden",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_secret_material_detected",
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_scope_overreach",
}

EXPECTED_DIAGNOSTICS = {
    "audit_store_storage_adapter_database_connection_lifecycle_readiness_status",
    "readiness_decision",
    "runtime_task_decision",
    "selected_driver_candidate",
    "dsn_handoff_status",
    "tls_role_environment_binding_status",
    "pool_policy_status",
    "timeout_budget_status",
    "retry_transaction_recovery_status",
    "duplicate_replay_status",
    "schema_marker_migration_handoff_status",
    "offline_verification_status",
    "negative_leakage_scan_status",
    "rollout_rollback_status",
    "connection_provider_status",
    "connection_runtime_status",
    "storage_adapter_runtime_status",
    "audit_store_runtime_status",
    "repository_mode_status",
    "production_api_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_ZERO_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "network_call_count",
    "cloud_secret_call_count",
    "provider_call_count",
    "database_connection_count",
    "driver_open_count",
    "driver_import_added_count",
    "driver_dependency_version_pinned_count",
    "go_mod_change_count",
    "go_sum_change_count",
    "dsn_parse_count",
    "dsn_build_count",
    "connection_factory_create_count",
    "pool_create_count",
    "pool_acquire_count",
    "health_check_runtime_create_count",
    "health_check_execute_count",
    "timeout_timer_create_count",
    "retry_executor_create_count",
    "transaction_begin_count",
    "partial_write_recovery_count",
    "duplicate_replay_decision_count",
    "sql_execution_count",
    "schema_marker_runtime_created_count",
    "migration_runner_created_count",
    "storage_adapter_runtime_task_card_created_count",
    "storage_adapter_runtime_created_count",
    "audit_store_runtime_created_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}

DOC_REFERENCES = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.md": [
        SLICE_STATUS,
        READINESS_DECISION,
        ENTRY_DECISION,
        "secret-ref-only DSN handoff",
        "不创建 connection runtime",
    ],
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1-plan.md": [
        SLICE_ID,
        SLICE_STATUS,
        READINESS_DECISION,
        ENTRY_DECISION,
        "停止线",
    ],
    "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
        SLICE_STATUS,
        MATRIX_BLOCKER_STATUS,
        NEXT_DEPENDENCY,
    ],
    "docs/task-cards/production-secret-backend-audit-store-runtime-blocker-matrix-v1-plan.md": [
        SLICE_STATUS,
        MATRIX_BLOCKER_STATUS,
        NEXT_DEPENDENCY,
    ],
    "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": [
        SLICE_STATUS,
        READINESS_DECISION,
        ENTRY_DECISION,
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Storage Adapter Database Connection Lifecycle Readiness v1",
        SLICE_STATUS,
        ENTRY_DECISION,
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Storage Adapter Database Connection Lifecycle Readiness v1",
        SLICE_STATUS,
        ENTRY_DECISION,
    ],
    "docs/features/workflow/README.md": [SLICE_STATUS, ENTRY_DECISION],
    "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, ENTRY_DECISION],
    "docs/radishmind-current-focus.md": [SLICE_STATUS, ENTRY_DECISION],
    "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.py",
        "production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.json",
    ],
    "docs/devlogs/2026-W27.md": [SLICE_STATUS, ENTRY_DECISION],
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


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_database_connection_lifecycle_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected slice status")
    require(slice_info.get("readiness_decision") == READINESS_DECISION, "unexpected readiness decision")
    require(slice_info.get("entry_decision") == ENTRY_DECISION, "unexpected entry decision")
    require(slice_info.get("next_dependency") == NEXT_DEPENDENCY, "unexpected next dependency")
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "database_connection_lifecycle_runtime_created",
        "database_connection_provider_created",
        "driver_import_added",
        "driver_dependency_version_pinned",
        "go_mod_changed",
        "go_sum_changed",
        "dsn_defined",
        "dsn_parser_created",
        "pool_runtime_created",
        "health_check_runtime_created",
        "sql_migration_created",
        "storage_adapter_runtime_task_card_created",
        "storage_adapter_runtime_created",
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in does_not_claim, f"does_not_claim missing {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependency_rows = rows_by_id(fixture, "depends_on", "id")
    require(set(dependency_rows) == set(EXPECTED_DEPENDENCIES), "dependency set drifted")
    for dep_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        row = dependency_rows.get(dep_id)
        require(row is not None, f"dependency missing: {dep_id}")
        require(row.get("evidence") == relative_path, f"dependency evidence mismatch: {dep_id}")
        require(row.get("status") == expected_status, f"dependency status mismatch: {dep_id}")
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"dependency source status mismatch: {dep_id}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("readiness_boundary") or {}
    for key, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(key) == expected, f"readiness_boundary.{key} must be {expected}")
    for key in EXPECTED_FALSE_FLAGS:
        require(boundary.get(key) is False, f"readiness_boundary.{key} must be false")

    gates = rows_by_id(fixture, "lifecycle_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATES), "lifecycle gate set drifted")
    for gate_id, expected_status in EXPECTED_GATES.items():
        row = gates[gate_id]
        require(row.get("status") == expected_status, f"{gate_id} status drifted")
        require(row.get("required_before_runtime") is True, f"{gate_id} must be required before runtime")
        require(row.get("runtime_created_now") is False, f"{gate_id} must not create runtime")


def assert_lifecycle_policies(fixture: dict[str, Any]) -> None:
    pool = fixture.get("pool_policy") or {}
    require(pool.get("status") == "static_pool_policy_defined_without_pool_runtime", "pool status drifted")
    for dimension in ("environment_class", "provider_profile_ref", "role_class", "tenant_workspace_scope"):
        require(dimension in (pool.get("required_future_pool_key_dimensions") or []), f"missing pool key {dimension}")
    for forbidden in ("runtime_write_and_migration_role_same_pool", "cross_environment_pool"):
        require(forbidden in (pool.get("forbidden_future_sharing") or []), f"missing pool ban {forbidden}")
    require(pool.get("close_owner") == "future_connection_provider_runtime", "pool close owner drifted")
    require(pool.get("pool_runtime_created_now") is False, "pool runtime forbidden")
    require(pool.get("connection_factory_created_now") is False, "connection factory forbidden")

    timeout = fixture.get("timeout_budget_policy") or {}
    require(timeout.get("status") == "static_timeout_budget_defined_without_runtime_timer", "timeout status drifted")
    for phase in ("connection_acquire", "transaction_begin", "append_only_write", "partial_write_recovery"):
        require(phase in (timeout.get("required_future_phases") or []), f"missing timeout phase {phase}")
    require(timeout.get("missing_timeout_must_fail_closed") is True, "timeout must fail closed")
    require(timeout.get("runtime_timer_created_now") is False, "runtime timer forbidden")
    require(timeout.get("driver_call_allowed_now") is False, "driver call forbidden")

    recovery = fixture.get("recovery_policy") or {}
    require(recovery.get("status") == "static_recovery_policy_defined_without_runtime", "recovery status drifted")
    for item in ("transaction_outcome", "idempotency_key_reference", "partial_write_recovery_policy"):
        require(item in (recovery.get("required_future_inputs") or []), f"missing recovery input {item}")
    require(recovery.get("unknown_commit_result_must_fail_closed") is True, "unknown commit must fail closed")
    require(recovery.get("recovery_executor_created_now") is False, "recovery executor forbidden")
    require(recovery.get("sql_execution_allowed_now") is False, "sql execution forbidden")

    handoff = fixture.get("schema_marker_migration_handoff") or {}
    require(handoff.get("status") == "schema_marker_migration_handoff_defined", "handoff status drifted")
    for target in ("schema_marker_runtime", "manual_migration_runner", "metadata_only_table_schema_artifact"):
        require(target in (handoff.get("required_future_handoff_targets") or []), f"missing handoff target {target}")
    require(handoff.get("schema_marker_runtime_created_now") is False, "schema marker runtime forbidden")
    require(handoff.get("migration_runner_created_now") is False, "migration runner forbidden")
    require(handoff.get("sql_or_ddl_created_now") is False, "sql or ddl forbidden")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "failure_mapping", "code")
    require(set(rows) == EXPECTED_FAILURE_CODES, "failure_mapping code set mismatch")
    for code, row in rows.items():
        require(row.get("failure_boundary"), f"failure boundary missing: {code}")
        require(row.get("sanitized_diagnostic"), f"sanitized diagnostic missing: {code}")


def assert_diagnostics_and_side_effects(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    require(set(diagnostics.get("allowed_fields") or []) == EXPECTED_DIAGNOSTICS, "allowed diagnostics mismatch")
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    for field in {"secret_value", "dsn", "connection_string", "endpoint", "host", "database_name", "credential_payload"}:
        require(field in forbidden, f"forbidden diagnostic missing {field}")
    sample = diagnostics.get("sample") or {}
    require(sample.get("selected_driver_candidate") == SELECTED_DRIVER_CANDIDATE, "diagnostic driver drifted")
    require(sample.get("readiness_decision") == READINESS_DECISION, "diagnostic decision drifted")
    require(sample.get("runtime_task_decision") == ENTRY_DECISION, "diagnostic runtime decision drifted")
    require(sample.get("connection_provider_status") == "not_created", "diagnostic connection provider drifted")
    require(sample.get("connection_runtime_status") == "not_created", "diagnostic connection runtime drifted")

    counters = fixture.get("side_effect_counters") or {}
    require(set(counters) == EXPECTED_ZERO_COUNTERS, "side effect counter set mismatch")
    for key in EXPECTED_ZERO_COUNTERS:
        require(counters.get(key) == 0, f"side_effect_counters.{key} must be 0")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("allowed_added_artifacts") or []:
        require((REPO_ROOT / relative_path).exists(), f"allowed artifact missing: {relative_path}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / relative_path).exists(), f"forbidden runtime artifact exists: {relative_path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "database_connection_lifecycle_runtime_task_card",
        "database_connection_provider",
        "connection_factory",
        "pool_runtime",
        "health_check_runtime",
        "driver_import",
        "driver_dependency_version_pin",
        "go_mod_change",
        "go_sum_change",
        "dsn_parser",
        "sql_migration",
        "schema_marker_runtime",
        "migration_runner",
        "audit_store_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact missing: {artifact}")


def assert_no_driver_import_or_dependency() -> None:
    module_files = [
        REPO_ROOT / "go.mod",
        REPO_ROOT / "go.sum",
        REPO_ROOT / "services/platform/go.mod",
        REPO_ROOT / "services/platform/go.sum",
    ]
    for path in module_files:
        if path.exists():
            text = path.read_text(encoding="utf-8")
            require("github.com/jackc/pgx/v5" not in text, f"{path.relative_to(REPO_ROOT)} must not pin pgx")
            require("github.com/jackc/pgx" not in text, f"{path.relative_to(REPO_ROOT)} must not pin pgx")

    services_root = REPO_ROOT / "services"
    for path in sorted(services_root.rglob("*.go")):
        text = path.read_text(encoding="utf-8")
        require("github.com/jackc/pgx/v5" not in text, f"{path.relative_to(REPO_ROOT)} must not import pgx")
        require("github.com/jackc/pgx" not in text, f"{path.relative_to(REPO_ROOT)} must not import pgx")


def assert_alignment(fixture: dict[str, Any]) -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    for field, expected in {
        "storage_adapter_database_connection_lifecycle_readiness_status": SLICE_STATUS,
        "storage_adapter_database_connection_lifecycle_readiness_decision": READINESS_DECISION,
        "storage_adapter_secret_ref_only_dsn_handoff_status": "secret_ref_only_dsn_handoff_defined",
        "storage_adapter_tls_role_environment_binding_status": "static_tls_role_environment_binding_defined",
        "storage_adapter_pool_policy_status": "static_pool_policy_defined_without_pool_runtime",
        "storage_adapter_timeout_budget_status": "static_timeout_budget_defined_without_runtime_timer",
        "storage_adapter_retry_transaction_recovery_status": "static_recovery_policy_defined_without_runtime",
        "storage_adapter_duplicate_replay_fail_closed_status": "duplicate_replay_fail_closed_policy_defined",
        "storage_adapter_sanitized_diagnostics_status": "sanitized_diagnostics_allowlist_defined",
        "storage_adapter_schema_marker_migration_handoff_status": "schema_marker_migration_handoff_defined",
        "storage_adapter_offline_verification_status": "metadata_only_offline_verification_defined",
        "storage_adapter_connection_lifecycle_runtime_status": "not_created",
        "storage_adapter_connection_factory_status": "not_created",
        "storage_adapter_pool_runtime_status": "not_created",
        "storage_adapter_health_check_runtime_status": "not_created",
        "storage_adapter_runtime_task_card_decision": CURRENT_ENTRY_DECISION,
        "storage_adapter_current_next_dependency": CURRENT_NEXT_DEPENDENCY,
        "durable_audit_backend_status": MATRIX_BLOCKER_STATUS,
    }.items():
        require(boundary.get(field) == expected, f"matrix boundary {field} drifted")

    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == MATRIX_BLOCKER_STATUS, "durable blocker status drifted")
    require(durable.get("source") == CURRENT_BLOCKER_SOURCE, "durable blocker source drifted")
    require(durable.get("unlock_condition") == CURRENT_NEXT_DEPENDENCY, "durable unlock condition drifted")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable must still block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable must still block resolver runtime")
    order = matrix.get("dependency_order") or []
    require("storage_adapter_database_connection_lifecycle_readiness" in order, "dependency order missing lifecycle")
    require(
        order.index("storage_adapter_database_driver_selection_review")
        < order.index("storage_adapter_database_connection_lifecycle_readiness")
        < order.index("audit_writer_runtime_entry_review"),
        "database connection lifecycle order drifted",
    )

    alignment = fixture.get("blocker_matrix_alignment") or {}
    require(
        alignment.get("durable_backend_blocker_status_after_readiness") == FIXTURE_MATRIX_BLOCKER_STATUS_AFTER_READINESS,
        "matrix status drifted",
    )
    require(
        alignment.get("durable_backend_blocker_source_after_readiness") == FIXTURE_MATRIX_BLOCKER_SOURCE_AFTER_READINESS,
        "matrix source drifted",
    )
    require(alignment.get("storage_adapter_current_next_dependency") == FIXTURE_NEXT_DEPENDENCY, "matrix next drifted")
    require(alignment.get("runtime_task_card_decision") == FIXTURE_RUNTIME_TASK_CARD_DECISION, "matrix decision drifted")

    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    expected = fixture.get("implementation_readiness_alignment") or {}
    for field, value in expected.items():
        if field == "status":
            continue
        if field == "audit_storage_adapter_runtime_task_card_decision":
            value = CURRENT_ENTRY_DECISION
        if field == "audit_storage_adapter_current_next_dependency":
            value = CURRENT_NEXT_DEPENDENCY
        if field == "audit_storage_adapter_provider_account_resource_status":
            value = "metadata_only_readiness_defined_without_real_resource"
        if field == "audit_storage_adapter_database_endpoint_status":
            value = "metadata_only_endpoint_requirements_defined_without_endpoint"
        require(target.get(field) == value, f"implementation readiness {field} drifted")

    planned = rows_by_id(readiness, "planned_slices", "id")
    item = planned.get("audit-store-storage-adapter-database-connection-lifecycle-readiness") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness missing lifecycle planned slice")


def assert_docs_and_registration() -> None:
    for path, literals in DOC_REFERENCES.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = "check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.py"
    current = "check-production-ops-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.py"
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in (previous, current, matrix):
        require(script in check_repo, f"check-repo.py missing {script}")
    require(check_repo.index(previous) < check_repo.index(current) < check_repo.index(matrix), "check-repo.py order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1-plan.md",
    ]
    text = "\n".join(read(path) for path in paths)
    for literal in ["BEGIN PRIVATE KEY", "AKIA", "authorization:", "connection_uri_with_secret"]:
        require(literal not in text, f"secret-like literal found: {literal}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_lifecycle_policies(fixture)
    assert_failure_mapping(fixture)
    assert_diagnostics_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_no_driver_import_or_dependency()
    assert_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store storage adapter database connection lifecycle readiness checks passed.")


if __name__ == "__main__":
    main()
