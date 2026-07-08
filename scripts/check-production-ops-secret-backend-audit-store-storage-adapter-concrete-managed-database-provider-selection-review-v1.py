#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-review-v1"
SLICE_STATUS = "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_defined"
SELECTION_DECISION = "concrete_managed_database_provider_reference_selected_runtime_blocked"
ENTRY_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_review"
NEXT_DEPENDENCY = "storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review"
CURRENT_MATRIX_BLOCKER_STATUS = (
    "storage_adapter_runtime_entry_refresh_after_concrete_managed_database_provider_selection_review_defined_task_card_blocked"
)
CURRENT_MATRIX_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-concrete-managed-database-provider-selection-review-v1"
)
CURRENT_ENTRY_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_review_entry_refresh"
)
CURRENT_NEXT_DEPENDENCY = "storage_adapter_provider_account_resource_endpoint_readiness"
FIXTURE_MATRIX_BLOCKER_STATUS = (
    "storage_adapter_concrete_managed_database_provider_selection_review_defined_task_card_blocked"
)
FIXTURE_MATRIX_BLOCKER_SOURCE = SLICE_ID
FIXTURE_ENTRY_DECISION = ENTRY_DECISION
FIXTURE_NEXT_DEPENDENCY = NEXT_DEPENDENCY
PREVIOUS_SLICE_ID = (
    "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-readiness-v1"
)
PREVIOUS_SLICE_STATUS = "audit_store_storage_adapter_concrete_managed_database_provider_selection_readiness_defined"
PREVIOUS_READINESS_DECISION = "concrete_managed_database_provider_selection_readiness_defined_without_provider_selection"
PREVIOUS_ENTRY_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_readiness"
)
PREVIOUS_NEXT_DEPENDENCY = "storage_adapter_concrete_managed_database_provider_selection_review"
SELECTED_PROVIDER_REFERENCE = "managed_postgresql_compatible_provider_reference"
SELECTED_PROVIDER_REFERENCE_KIND = "reference_only_concrete_provider_profile"
SELECTED_PROFILE = "managed_postgresql_compatible_audit_store_profile"
SELECTED_ENGINE = "postgresql_compatible_append_only_relational_database"
SELECTED_PROVIDER_CLASS = "managed_postgresql_compatible_service"
SELECTED_DRIVER_CANDIDATE = "github.com/jackc/pgx/v5"

EXPECTED_DEPENDENCIES = {
    PREVIOUS_SLICE_ID: (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-readiness-v1.json"
        ),
        PREVIOUS_SLICE_STATUS,
    ),
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-managed-database-product-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
            "after-managed-database-product-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_managed_database_product_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-managed-database-product-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-managed-database-product-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_managed_database_product_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_database_provider_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_database_driver_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.json"
        ),
        "audit_store_storage_adapter_database_connection_lifecycle_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1.json"
        ),
        "audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_defined",
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

EXPECTED_REVIEW = {
    "status": SLICE_STATUS,
    "selection_decision": SELECTION_DECISION,
    "entry_decision": ENTRY_DECISION,
    "previous_readiness_status": PREVIOUS_SLICE_STATUS,
    "previous_readiness_decision": PREVIOUS_READINESS_DECISION,
    "previous_runtime_task_card_decision": PREVIOUS_ENTRY_DECISION,
    "previous_next_dependency": PREVIOUS_NEXT_DEPENDENCY,
    "next_dependency": NEXT_DEPENDENCY,
    "selected_provider_reference": SELECTED_PROVIDER_REFERENCE,
    "selected_provider_reference_kind": SELECTED_PROVIDER_REFERENCE_KIND,
    "selected_managed_product_profile": SELECTED_PROFILE,
    "selected_database_engine": SELECTED_ENGINE,
    "selected_provider_candidate_class": SELECTED_PROVIDER_CLASS,
    "selected_database_driver_candidate": SELECTED_DRIVER_CANDIDATE,
    "concrete_provider_selection_status": "reference_profile_selected_without_provider_resource",
    "concrete_provider_selection_review_status": SLICE_STATUS,
    "managed_database_product_status": "selected_reference_profile_not_vendor_product",
    "database_vendor_status": "not_selected",
    "concrete_cloud_product_status": "not_selected",
    "concrete_database_provider_status": "reference_profile_selected_without_provider_resource",
    "provider_account_resource_status": "not_defined",
    "database_endpoint_status": "not_defined",
    "region_detail_status": "not_defined",
    "database_provider_status": "provider_reference_selected_without_runtime_provider",
    "database_driver_status": "selected_reference_only",
    "database_driver_import_status": "not_created",
    "driver_dependency_version_status": "not_pinned",
    "database_dsn_parser_status": "not_created",
    "database_connection_provider_status": "not_created",
    "database_connection_factory_status": "not_created",
    "database_pool_runtime_status": "not_created",
    "database_health_check_runtime_status": "not_created",
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
    "go_import_added_in_this_slice",
    "dependency_downloaded_in_this_slice",
    "managed_database_product_selected_in_this_slice",
    "concrete_cloud_product_selected_in_this_slice",
    "database_vendor_selected_in_this_slice",
    "provider_resource_selected_in_this_slice",
    "provider_account_resource_defined_in_this_slice",
    "database_endpoint_defined_in_this_slice",
    "region_detail_defined_in_this_slice",
    "driver_dependency_version_pinned_in_this_slice",
    "real_dsn_parsed_in_this_slice",
    "database_provider_created_in_this_slice",
    "database_connection_provider_created_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_CAPABILITIES = {
    "managed_postgresql_compatible_append_only_audit_table",
    "secret_ref_only_dsn_handoff",
    "tls_role_environment_binding",
    "least_privilege_role_policy",
    "pool_timeout_health_check_policy_without_runtime",
    "retry_transaction_partial_write_recovery_policy",
    "duplicate_replay_fail_closed_policy",
    "sanitized_diagnostics_allowlist",
    "schema_marker_migration_handoff",
    "offline_verification_without_connection",
    "backup_restore_recovery_exit_boundary",
    "operator_review_cost_quota_observability",
}

EXPECTED_DIMENSIONS = {
    "reference_profile_scope",
    "managed_postgresql_compatibility",
    "append_only_audit_table_fit",
    "secret_ref_only_dsn_handoff_fit",
    "tls_role_environment_binding_fit",
    "pool_timeout_health_check_policy_fit",
    "transaction_idempotency_and_partial_write_recovery_fit",
    "duplicate_replay_fail_closed_fit",
    "sanitized_diagnostics_fit",
    "schema_marker_migration_handoff_fit",
    "offline_verification_without_runtime_fit",
    "backup_restore_recovery_and_exit_fit",
    "operator_review_cost_quota_observability_fit",
}

EXPECTED_REJECTION_CONDITIONS = {
    "candidate_contains_endpoint_or_resource_detail",
    "candidate_requires_raw_secret_or_dsn",
    "candidate_selects_vendor_product_or_provider_resource_as_fact",
    "candidate_requires_driver_import_or_version_pin",
    "candidate_requires_runtime_probe_cloud_api_or_sql",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_dependency_missing",
    "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_reference_profile_missing",
    "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_provider_resource_forbidden",
    "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_runtime_forbidden",
    "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_secret_material_detected",
}

EXPECTED_ZERO_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "network_call_count",
    "cloud_provider_call_count",
    "database_connection_count",
    "driver_open_count",
    "sql_execution_count",
    "database_vendor_selected_count",
    "managed_database_product_selected_count",
    "concrete_cloud_product_selected_count",
    "provider_resource_selected_count",
    "provider_account_resource_defined_count",
    "database_endpoint_defined_count",
    "driver_dependency_version_pinned_count",
    "dsn_parser_created_count",
    "connection_provider_created_count",
    "sql_migration_created_count",
    "schema_marker_runtime_created_count",
    "migration_runner_created_count",
    "storage_adapter_runtime_task_card_created_count",
    "storage_adapter_runtime_created_count",
    "audit_store_runtime_created_count",
    "production_api_call_count",
}

DOC_REFERENCES = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-review-v1.md": [
        SLICE_ID,
        SLICE_STATUS,
        SELECTION_DECISION,
        ENTRY_DECISION,
        NEXT_DEPENDENCY,
        SELECTED_PROVIDER_REFERENCE,
    ],
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-review-v1-plan.md": [
        SLICE_ID,
        SLICE_STATUS,
        NEXT_DEPENDENCY,
        "停止线",
    ],
    "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
        SLICE_STATUS,
        CURRENT_MATRIX_BLOCKER_STATUS,
        NEXT_DEPENDENCY,
    ],
    "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": [
        SLICE_STATUS,
        SELECTION_DECISION,
        NEXT_DEPENDENCY,
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Storage Adapter Concrete Managed Database Provider Selection Review v1",
        SLICE_STATUS,
        NEXT_DEPENDENCY,
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Storage Adapter Concrete Managed Database Provider Selection Review v1",
        SLICE_STATUS,
        NEXT_DEPENDENCY,
    ],
    "docs/features/workflow/README.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/radishmind-current-focus.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/radishmind-integration-contracts.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/radishmind-product-scope.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/README.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        SLICE_ID,
        SLICE_STATUS,
        NEXT_DEPENDENCY,
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-review-v1.py",
        "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-review-v1.json",
    ],
    "docs/devlogs/2026-W28.md": [SLICE_STATUS, NEXT_DEPENDENCY],
}

SECRET_LITERAL_PATTERNS = [
    re.compile(r"Bearer\s+[A-Za-z0-9._-]+"),
    re.compile(r"-----BEGIN [A-Z ]+-----"),
    re.compile(r"AKIA[0-9A-Z]{16}"),
    re.compile(r"sk-[A-Za-z0-9]{20,}"),
    re.compile(r"://[^\s:/]+:[^\s@]+@"),
    re.compile(r"\b(endpoint|host|database_name|resource_id|account_id)\s*[:=]\s*[A-Za-z0-9._-]+", re.IGNORECASE),
]


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a JSON object")
    return document


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def check_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_concrete_managed_database_provider_selection_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "slice id drifted")
    require(slice_info.get("status") == SLICE_STATUS, "slice status drifted")
    require(slice_info.get("selection_decision") == SELECTION_DECISION, "selection decision drifted")
    require(slice_info.get("entry_decision") == ENTRY_DECISION, "entry decision drifted")
    require(slice_info.get("next_dependency") == NEXT_DEPENDENCY, "next dependency drifted")
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "database_vendor_selected",
        "concrete_cloud_product_selected",
        "concrete_provider_resource_selected",
        "provider_account_resource_defined",
        "database_endpoint_defined",
        "database_driver_imported",
        "dsn_parser_created",
        "storage_adapter_runtime_task_card_created",
        "storage_adapter_runtime_created",
        "audit_store_runtime_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in does_not_claim, f"does_not_claim missing {claim}")


def check_dependencies(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "depends_on", "id")
    require(set(rows) == set(EXPECTED_DEPENDENCIES), "depends_on set drifted")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        row = rows[dependency_id]
        require(row.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        require(row.get("status") == expected_status, f"{dependency_id} status drifted")
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"{dependency_id} source status drifted")


def check_selection_review(fixture: dict[str, Any]) -> None:
    review = fixture.get("selection_review") or {}
    for field, expected in EXPECTED_REVIEW.items():
        require(review.get(field) == expected, f"selection_review.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(review.get(field) is False, f"selection_review.{field} must remain false")

    profile = fixture.get("provider_reference_profile") or {}
    require(profile.get("key") == SELECTED_PROVIDER_REFERENCE, "provider reference key drifted")
    require(profile.get("kind") == SELECTED_PROVIDER_REFERENCE_KIND, "provider reference kind drifted")
    require(set(profile.get("required_capabilities") or []) == EXPECTED_CAPABILITIES, "provider capabilities drifted")
    forbidden = set(profile.get("forbidden_material") or [])
    for field in {"endpoint", "host", "database_name", "provider_resource_id", "provider_account_id", "dsn"}:
        require(field in forbidden, f"provider reference must forbid {field}")

    dimensions = set(fixture.get("selection_dimensions") or [])
    require(dimensions == EXPECTED_DIMENSIONS, "selection dimensions drifted")
    rejection_rows = rows_by_id(fixture, "rejection_conditions", "condition")
    require(set(rejection_rows) == EXPECTED_REJECTION_CONDITIONS, "rejection condition set drifted")
    for condition, row in rejection_rows.items():
        require(row.get("result") == "reject_before_reference_profile_selection", f"{condition} result drifted")


def check_diagnostics_and_failures(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    for field in {"sanitized_diagnostic", "failure_code", "selected_provider_reference"}:
        require(field in allowed, f"diagnostic allowed field missing {field}")
    for field in {"dsn", "endpoint", "host", "database_name", "resource_id", "account_id", "credential_payload"}:
        require(field in forbidden, f"diagnostic forbidden field missing {field}")
    sample = diagnostics.get("sample") or {}
    require(set(sample).issubset(allowed), "diagnostic sample contains non-allowed fields")
    require(sample.get("slice_status") == SLICE_STATUS, "diagnostic sample status drifted")
    require(sample.get("selection_decision") == SELECTION_DECISION, "diagnostic sample selection drifted")
    require(sample.get("entry_decision") == ENTRY_DECISION, "diagnostic sample entry decision drifted")
    require(sample.get("next_dependency") == NEXT_DEPENDENCY, "diagnostic sample next drifted")
    require(sample.get("selected_provider_reference") == SELECTED_PROVIDER_REFERENCE, "diagnostic reference drifted")
    require(
        sample.get("concrete_database_provider_status") == "reference_profile_selected_without_provider_resource",
        "diagnostic provider status drifted",
    )

    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure mapping code set drifted")
    for code, row in failures.items():
        require(row.get("failure_boundary"), f"failure boundary missing: {code}")
        require(row.get("sanitized_diagnostic"), f"sanitized diagnostic missing: {code}")

    counters = fixture.get("side_effect_counters") or {}
    require(set(counters) == EXPECTED_ZERO_COUNTERS, "side effect counter set drifted")
    for field in EXPECTED_ZERO_COUNTERS:
        require(counters.get(field) == 0, f"side_effect_counters.{field} must remain zero")


def check_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("allowed_added_artifacts") or []:
        require((REPO_ROOT / relative_path).exists(), f"allowed artifact missing: {relative_path}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / relative_path).exists(), f"forbidden runtime artifact exists: {relative_path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "concrete_cloud_product_selection",
        "database_vendor_selection",
        "concrete_provider_resource_selection",
        "provider_account_resource",
        "database_endpoint",
        "driver_import",
        "driver_dependency_version_pin",
        "dsn_parser",
        "database_provider_runtime",
        "connection_provider_runtime",
        "sql",
        "ddl",
        "schema_marker_runtime",
        "migration_runner_runtime",
        "storage_adapter_runtime",
        "audit_store_runtime",
        "production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact kind missing: {artifact}")


def check_aggregate_alignment(fixture: dict[str, Any]) -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    for field, expected in {
        "durable_audit_backend_status": CURRENT_MATRIX_BLOCKER_STATUS,
        "storage_adapter_concrete_managed_database_provider_selection_review_status": SLICE_STATUS,
        "storage_adapter_selected_concrete_provider_reference": SELECTED_PROVIDER_REFERENCE,
        "storage_adapter_selected_concrete_provider_reference_kind": SELECTED_PROVIDER_REFERENCE_KIND,
        "storage_adapter_concrete_provider_selection_status": "reference_profile_selected_without_provider_resource",
        "storage_adapter_concrete_provider_selection_review_status": SLICE_STATUS,
        "storage_adapter_database_vendor_status": "not_selected",
        "storage_adapter_concrete_cloud_product_status": "not_selected",
        "storage_adapter_database_provider_status": "provider_reference_selected_without_runtime_provider",
        "storage_adapter_database_connection_provider_status": "not_created",
        "storage_adapter_runtime_task_card_decision": CURRENT_ENTRY_DECISION,
        "storage_adapter_current_next_dependency": CURRENT_NEXT_DEPENDENCY,
        "storage_adapter_runtime_task_card_status": "not_created",
        "storage_adapter_runtime_status": "not_created",
    }.items():
        require(boundary.get(field) == expected, f"matrix boundary {field} drifted")

    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == CURRENT_MATRIX_BLOCKER_STATUS, "durable blocker status drifted")
    require(durable.get("source") == CURRENT_MATRIX_BLOCKER_SOURCE, "durable blocker source drifted")
    require(durable.get("unlock_condition") == CURRENT_NEXT_DEPENDENCY, "durable unlock condition drifted")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable must block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable must block resolver runtime")

    order = matrix.get("dependency_order") or []
    require("storage_adapter_concrete_managed_database_provider_selection_review" in order, "dependency order missing review")
    require(
        order.index("storage_adapter_concrete_managed_database_provider_selection_readiness")
        < order.index("storage_adapter_concrete_managed_database_provider_selection_review")
        < order.index("storage_adapter_runtime_entry_refresh_after_concrete_managed_database_provider_selection_review")
        < order.index("audit_writer_runtime_entry_review"),
        "concrete managed database provider selection review order drifted",
    )

    alignment = fixture.get("blocker_matrix_alignment") or {}
    require(alignment.get("status") == FIXTURE_MATRIX_BLOCKER_STATUS, "fixture matrix status drifted")
    require(alignment.get("source") == FIXTURE_MATRIX_BLOCKER_SOURCE, "fixture matrix source drifted")
    require(alignment.get("unlock_condition") == FIXTURE_NEXT_DEPENDENCY, "fixture matrix unlock drifted")
    require(alignment.get("runtime_task_card_decision") == FIXTURE_ENTRY_DECISION, "fixture matrix decision drifted")

    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    expected = fixture.get("implementation_readiness_alignment") or {}
    require(target.get(expected.get("status_field")) == SLICE_STATUS, "implementation readiness status field drifted")
    for field, value in expected.items():
        if field in {"status", "status_field"}:
            continue
        if field == "audit_storage_adapter_runtime_task_card_decision":
            value = CURRENT_ENTRY_DECISION
        if field == "audit_storage_adapter_current_next_dependency":
            value = CURRENT_NEXT_DEPENDENCY
        require(target.get(field) == value, f"implementation readiness {field} drifted")
    planned = rows_by_id(readiness, "planned_slices", "id")
    item = planned.get("audit-store-storage-adapter-concrete-managed-database-provider-selection-review") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness planned slice missing current status")


def check_docs_and_registration() -> None:
    for path, literals in DOC_REFERENCES.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-"
        "readiness-v1.py"
    )
    current = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-"
        "review-v1.py"
    )
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in (previous, current, matrix):
        require(script in check_repo, f"check-repo.py missing {script}")
    require(check_repo.index(previous) < check_repo.index(current) < check_repo.index(matrix), "check-repo.py order drifted")


def check_no_secret_material(fixture: dict[str, Any]) -> None:
    paths = (fixture.get("no_secret_material_scan") or {}).get("paths") or []
    require(paths, "no secret material scan paths missing")
    for relative_path in paths:
        content = read(str(relative_path))
        for pattern in SECRET_LITERAL_PATTERNS:
            require(not pattern.search(content), f"secret-like literal found in {relative_path}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    check_slice(fixture)
    check_dependencies(fixture)
    check_selection_review(fixture)
    check_diagnostics_and_failures(fixture)
    check_artifact_guard(fixture)
    check_aggregate_alignment(fixture)
    check_docs_and_registration()
    check_no_secret_material(fixture)
    print("production secret backend audit store storage adapter concrete managed database provider selection review v1 ok")


if __name__ == "__main__":
    main()
