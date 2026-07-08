#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
BLOCKER_MATRIX_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1"
SLICE_STATUS = "audit_store_storage_adapter_database_provider_selection_readiness_defined"
READINESS_DECISION = "database_provider_selection_readiness_defined_without_provider_selection"
ENTRY_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_database_provider_selection_readiness"
NEXT_DEPENDENCY = "storage_adapter_database_provider_selection_review"
SELECTED_DATABASE_ENGINE = "postgresql_compatible_append_only_relational_database"
PROVIDER_SELECTION_STATUS = "readiness_defined_without_provider_selection"
MATRIX_BLOCKER_STATUS = "storage_adapter_database_provider_selection_readiness_defined_task_card_blocked"
CURRENT_ENTRY_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_review_entry_refresh"
)
CURRENT_NEXT_DEPENDENCY = (
    "storage_adapter_provider_account_resource_endpoint_readiness"
)
CURRENT_PROVIDER_SELECTION_STATUS = "selected_provider_candidate_class_without_vendor_or_product"
CURRENT_PROVIDER_REVIEW_STATUS = "audit_store_storage_adapter_database_provider_selection_review_defined"
CURRENT_DATABASE_PROVIDER_STATUS = "provider_reference_selected_without_runtime_provider"
CURRENT_MATRIX_BLOCKER_STATUS = (
    "storage_adapter_runtime_entry_refresh_after_concrete_managed_database_provider_selection_review_defined_task_card_blocked"
)
CURRENT_MATRIX_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-concrete-managed-database-provider-selection-review-v1"
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
    "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_concrete_database_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.json"
        ),
        "audit_store_storage_adapter_concrete_database_selection_readiness_defined",
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
    "production-secret-backend-audit-store-runtime-blocker-matrix-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json",
        "audit_store_runtime_blocker_matrix_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}

EXPECTED_READINESS_BOUNDARY = {
    "status": SLICE_STATUS,
    "readiness_decision": READINESS_DECISION,
    "entry_decision": ENTRY_DECISION,
    "previous_review_status": "audit_store_storage_adapter_concrete_database_selection_review_defined",
    "previous_runtime_task_decision": "storage_adapter_runtime_task_card_still_blocked_after_concrete_database_selection_review",
    "selected_backend_product_class": "managed_database_append_only_table",
    "selected_backend_product_profile": "reserved_managed_database_append_only_table_profile",
    "selected_database_engine": SELECTED_DATABASE_ENGINE,
    "selected_database_engine_status": "selected_without_vendor_product_driver_or_provider",
    "provider_candidate_source_status": "metadata_only_provider_candidate_source_defined",
    "provider_input_evidence_status": "metadata_only_provider_input_evidence_defined",
    "provider_evaluation_dimension_status": "metadata_only_provider_evaluation_dimensions_defined",
    "database_provider_selection_status": PROVIDER_SELECTION_STATUS,
    "provider_selection_review_status": "not_started",
    "next_dependency": NEXT_DEPENDENCY,
    "database_product_status": "engine_selected_without_managed_product",
    "managed_database_product_status": "not_selected",
    "database_vendor_status": "not_selected",
    "database_provider_status": "not_selected",
    "provider_account_resource_status": "not_defined",
    "database_endpoint_status": "not_defined",
    "database_connection_provider_status": "not_created",
    "database_driver_status": "not_selected",
    "database_dsn_status": "not_defined",
    "database_connection_status": "not_created",
    "sql_migration_status": "not_created",
    "ddl_status": "not_created",
    "physical_table_schema_status": "not_created",
    "schema_marker_runtime_status": "not_created",
    "migration_runner_status": "not_created",
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "storage_adapter_client_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_writer_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "delivery_runtime_status": "not_created",
    "operator_approval_runtime_status": "not_created",
    "credential_handle_runtime_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "no_leakage_smoke_runtime_status": "not_created",
    "production_resolver_runtime_task_card_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "database_vendor_selected_in_this_slice",
    "managed_database_product_selected_in_this_slice",
    "database_provider_selected_in_this_slice",
    "provider_account_resource_defined_in_this_slice",
    "database_endpoint_defined_in_this_slice",
    "database_connection_provider_enabled",
    "database_driver_selected_in_this_slice",
    "database_connection_created_in_this_slice",
    "dsn_defined_in_this_slice",
    "sql_migration_created_in_this_slice",
    "ddl_created_in_this_slice",
    "physical_table_schema_created_in_this_slice",
    "schema_marker_runtime_created_in_this_slice",
    "migration_runner_created_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "storage_adapter_client_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "production_resolver_runtime_task_card_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_PROVIDER_CLASSES = {
    "managed_postgresql_compatible_service": "candidate_class_allowed_for_later_review_not_selected",
    "self_managed_postgresql_compatible_cluster": "candidate_class_allowed_for_later_review_not_selected",
    "embedded_or_file_database": "rejected_for_production_audit_store_boundary",
}

EXPECTED_DIMENSIONS = {
    "postgresql_compatibility_and_version_drift",
    "append_only_transaction_boundary",
    "duplicate_replay_fail_closed",
    "backup_restore_and_recovery",
    "retention_redaction_compatibility",
    "secret_ref_only_dsn_policy",
    "tls_mode_and_certificate_verification",
    "least_privilege_role_policy",
    "connection_lifecycle_and_pooling",
    "sanitized_error_and_diagnostics",
    "migration_schema_marker_handoff",
    "offline_adapter_smoke_boundary",
    "negative_leakage_runtime_scan_boundary",
    "operator_review_and_rollout",
    "cost_observability_and_operations",
    "environment_parity_and_tenancy_isolation",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_database_provider_selection_readiness_dependency_missing",
    "audit_store_storage_adapter_database_provider_selection_readiness_input_evidence_missing",
    "audit_store_storage_adapter_database_provider_selection_readiness_candidate_source_overreach",
    "audit_store_storage_adapter_database_provider_selection_readiness_provider_or_product_forbidden",
    "audit_store_storage_adapter_database_provider_selection_readiness_driver_or_dsn_forbidden",
    "audit_store_storage_adapter_database_provider_selection_readiness_runtime_forbidden",
    "audit_store_storage_adapter_database_provider_selection_readiness_secret_material_detected",
    "audit_store_storage_adapter_database_provider_selection_readiness_scope_overreach",
}

EXPECTED_DIAGNOSTICS = {
    "audit_store_storage_adapter_database_provider_selection_readiness_status",
    "readiness_decision",
    "runtime_task_decision",
    "next_dependency",
    "selected_database_engine",
    "provider_candidate_source_status",
    "provider_input_evidence_status",
    "provider_evaluation_dimension_status",
    "database_provider_selection_status",
    "provider_selection_review_status",
    "managed_database_product_status",
    "database_vendor_status",
    "database_provider_status",
    "provider_account_resource_status",
    "database_endpoint_status",
    "database_driver_status",
    "database_dsn_status",
    "connection_provider_status",
    "schema_marker_runtime_status",
    "migration_runner_status",
    "storage_adapter_runtime_task_card_status",
    "storage_adapter_runtime_status",
    "audit_store_runtime_status",
    "production_resolver_runtime_status",
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
    "sql_execution_count",
    "database_vendor_selected_count",
    "managed_database_product_selected_count",
    "database_provider_selected_count",
    "provider_account_resource_defined_count",
    "database_endpoint_defined_count",
    "database_connection_provider_created_count",
    "database_driver_selected_count",
    "dsn_defined_count",
    "sql_migration_created_count",
    "ddl_created_count",
    "physical_table_schema_created_count",
    "schema_marker_runtime_created_count",
    "migration_runner_created_count",
    "storage_adapter_runtime_task_card_created_count",
    "storage_adapter_runtime_created_count",
    "audit_store_runtime_created_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}

DOC_REFERENCES = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.md": [
        SLICE_STATUS,
        READINESS_DECISION,
        ENTRY_DECISION,
        NEXT_DEPENDENCY,
        "不选择云厂商",
        "不创建 storage adapter runtime",
    ],
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1-plan.md": [
        SLICE_ID,
        SLICE_STATUS,
        NEXT_DEPENDENCY,
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
        NEXT_DEPENDENCY,
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Storage Adapter Database Provider Selection Readiness v1",
        SLICE_STATUS,
        NEXT_DEPENDENCY,
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Storage Adapter Database Provider Selection Readiness v1",
        SLICE_STATUS,
        NEXT_DEPENDENCY,
    ],
    "docs/features/workflow/README.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/radishmind-current-focus.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        SLICE_ID,
        SLICE_STATUS,
        NEXT_DEPENDENCY,
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.py",
        "production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.json",
    ],
    "docs/devlogs/2026-W27.md": [SLICE_STATUS, NEXT_DEPENDENCY],
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
        == "production_ops_secret_backend_audit_store_storage_adapter_database_provider_selection_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected slice status")
    require(slice_info.get("readiness_decision") == READINESS_DECISION, "unexpected readiness decision")
    require(slice_info.get("entry_decision") == ENTRY_DECISION, "unexpected entry decision")
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "database_vendor_selected",
        "managed_database_product_selected",
        "database_provider_selected",
        "database_driver_selected",
        "dsn_defined",
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
    for dep_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        row = dependency_rows.get(dep_id)
        require(row is not None, f"dependency missing: {dep_id}")
        require(row.get("evidence") == relative_path, f"dependency evidence mismatch: {dep_id}")
        require(row.get("status") == expected_status, f"dependency status mismatch: {dep_id}")
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"dependency source status mismatch: {dep_id}")


def assert_readiness_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("readiness_boundary") or {}
    for key, expected in EXPECTED_READINESS_BOUNDARY.items():
        require(boundary.get(key) == expected, f"readiness_boundary.{key} must be {expected}")
    for key in EXPECTED_FALSE_FLAGS:
        require(boundary.get(key) is False, f"readiness_boundary.{key} must be false")

    inputs = rows_by_id(fixture, "input_evidence_requirements", "field")
    for field in {
        "candidate_key",
        "deployment_model",
        "compatibility_claim_ref",
        "append_only_capability_ref",
        "secret_ref_policy_ref",
        "tls_role_policy_ref",
        "migration_schema_marker_ref",
        "backup_restore_recovery_ref",
        "offline_validation_ref",
        "operator_review_ref",
    }:
        require(field in inputs, f"input evidence missing {field}")
    require(
        "endpoint" in set(inputs["candidate_key"].get("forbidden_material") or []),
        "candidate_key must forbid endpoint detail",
    )

    classes = rows_by_id(fixture, "provider_candidate_classes", "candidate_class")
    for candidate_class, expected in EXPECTED_PROVIDER_CLASSES.items():
        require(classes.get(candidate_class, {}).get("readiness_result") == expected, f"{candidate_class} drifted")

    dimensions = set(fixture.get("candidate_evaluation_dimensions") or [])
    require(dimensions == EXPECTED_DIMENSIONS, "candidate evaluation dimensions mismatch")
    stop_lines = set(fixture.get("stop_lines") or [])
    for item in {"no_vendor_selection", "no_driver_selection", "no_storage_adapter_runtime"}:
        require(item in stop_lines, f"stop line missing {item}")


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
    for field in {"secret_value", "dsn", "database_hostname", "table_name", "column_name"}:
        require(field in forbidden, f"forbidden diagnostic missing {field}")
    sample = diagnostics.get("sample") or {}
    require(sample.get("selected_database_engine") == SELECTED_DATABASE_ENGINE, "diagnostic sample engine drifted")
    require(sample.get("next_dependency") == NEXT_DEPENDENCY, "diagnostic sample next dependency drifted")
    require(
        sample.get("database_provider_selection_status") == PROVIDER_SELECTION_STATUS,
        "diagnostic provider selection status drifted",
    )

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
        "database_provider_implementation_task_card",
        "storage_adapter_runtime_implementation_task_card",
        "database_connection_provider",
        "database_driver",
        "dsn_parser",
        "provider_account_resource",
        "database_endpoint",
        "sql_migration",
        "schema_marker_runtime",
        "migration_runner",
        "audit_store_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact missing: {artifact}")


def assert_alignment(fixture: dict[str, Any]) -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    for field, expected in {
        "storage_adapter_database_provider_selection_readiness_status": SLICE_STATUS,
        "storage_adapter_database_provider_selection_status": CURRENT_PROVIDER_SELECTION_STATUS,
        "storage_adapter_provider_candidate_source_status": "metadata_only_provider_candidate_source_defined",
        "storage_adapter_provider_input_evidence_status": "metadata_only_provider_input_evidence_defined",
        "storage_adapter_provider_evaluation_dimension_status": "metadata_only_provider_evaluation_dimensions_defined",
        "storage_adapter_provider_selection_review_status": CURRENT_PROVIDER_REVIEW_STATUS,
        "storage_adapter_runtime_task_card_decision": CURRENT_ENTRY_DECISION,
        "storage_adapter_current_next_dependency": CURRENT_NEXT_DEPENDENCY,
        "storage_adapter_database_provider_status": CURRENT_DATABASE_PROVIDER_STATUS,
        "storage_adapter_database_connection_provider_status": "not_created",
        "storage_adapter_database_driver_status": "selected_reference_only",
        "storage_adapter_database_dsn_status": "not_defined",
        "storage_adapter_runtime_task_card_status": "not_created",
        "storage_adapter_runtime_status": "not_created",
        "durable_audit_backend_status": CURRENT_MATRIX_BLOCKER_STATUS,
    }.items():
        require(boundary.get(field) == expected, f"matrix boundary {field} drifted")

    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == CURRENT_MATRIX_BLOCKER_STATUS, "durable blocker status drifted")
    require(durable.get("source") == CURRENT_MATRIX_BLOCKER_SOURCE, "durable blocker source drifted")
    require(durable.get("unlock_condition") == CURRENT_NEXT_DEPENDENCY, "durable unlock condition drifted")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable must still block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable must still block resolver runtime")
    order = matrix.get("dependency_order") or []
    require("storage_adapter_database_provider_selection_readiness" in order, "dependency order missing provider readiness")
    require(
        order.index("storage_adapter_concrete_database_selection_review")
        < order.index("storage_adapter_database_provider_selection_readiness")
        < order.index("storage_adapter_database_provider_selection_review")
        < order.index("audit_writer_runtime_entry_review"),
        "database provider selection readiness order drifted",
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
        if field == "audit_storage_adapter_database_provider_status":
            value = CURRENT_DATABASE_PROVIDER_STATUS
        require(target.get(field) == value, f"implementation readiness {field} drifted")

    planned = rows_by_id(readiness, "planned_slices", "id")
    item = planned.get("audit-store-storage-adapter-database-provider-selection-readiness") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness missing provider readiness planned slice")


def assert_docs_and_registration() -> None:
    for path, literals in DOC_REFERENCES.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = "check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.py"
    current = "check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.py"
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in (previous, current, matrix):
        require(script in check_repo, f"check-repo.py missing {script}")
    require(check_repo.index(previous) < check_repo.index(current) < check_repo.index(matrix), "check-repo.py order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1-plan.md",
    ]
    text = "\n".join(read(path) for path in paths)
    for literal in ["BEGIN PRIVATE KEY", "AKIA", "authorization:", "connection_uri_with_secret"]:
        require(literal not in text, f"secret-like literal found: {literal}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_readiness_boundary(fixture)
    assert_failure_mapping(fixture)
    assert_diagnostics_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store storage adapter database provider selection readiness checks passed.")


if __name__ == "__main__":
    main()
