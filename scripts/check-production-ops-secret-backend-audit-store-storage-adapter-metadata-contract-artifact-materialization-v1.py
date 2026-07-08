#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json"
)
ENTRY_REVIEW_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1.json"
)
READINESS_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
CONTRACT_PATH = REPO_ROOT / "contracts/production-secret-audit-storage-adapter.metadata-contract.json"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1"
SLICE_STATUS = "audit_store_storage_adapter_metadata_contract_artifact_materialized"
CONTRACT_VERSION = "audit-storage-adapter-metadata-contract-v1"
NEXT_DEPENDENCY = "storage_adapter_backend_product_selection_review"
ENTRY_REVIEW_STATUS = "audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_defined"
FOLLOWUP_SELECTION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json"
)
FOLLOWUP_SELECTION_STATUS = "audit_store_storage_adapter_backend_product_selection_review_defined"
FOLLOWUP_SELECTION_NEXT_DEPENDENCY = "storage_adapter_runtime_implementation_entry_refresh_after_product_selection"
FOLLOWUP_AFTER_SELECTION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.json"
)
FOLLOWUP_AFTER_SELECTION_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined"
)
FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY = "storage_adapter_database_provider_connection_runtime_boundary_readiness"
FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1.json"
)
FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_STATUS = (
    "audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_defined"
)
FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_NEXT_DEPENDENCY = (
    "storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_readiness"
)
FOLLOWUP_AFTER_PROVIDER_BOUNDARY_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-database-provider-connection-runtime-boundary-v1.json"
)
FOLLOWUP_AFTER_PROVIDER_BOUNDARY_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_defined"
)
FOLLOWUP_AFTER_PROVIDER_BOUNDARY_NEXT_DEPENDENCY = (
    "storage_adapter_managed_database_product_selection_readiness"
)
FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-managed-database-product-selection-review-v1.json"
)
FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_STATUS = (
    "audit_store_storage_adapter_managed_database_product_selection_review_defined"
)
FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_NEXT_DEPENDENCY = (
    "storage_adapter_runtime_implementation_entry_refresh_after_managed_database_product_selection_review"
)
FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-managed-database-product-selection-review-v1.json"
)
FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_managed_database_product_selection_review_defined"
)
FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_NEXT_DEPENDENCY = (
    "storage_adapter_concrete_managed_database_provider_selection_readiness"
)
FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-readiness-v1.json"
)
FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_STATUS = (
    "audit_store_storage_adapter_concrete_managed_database_provider_selection_readiness_defined"
)
FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_NEXT_DEPENDENCY = (
    "storage_adapter_concrete_managed_database_provider_selection_review"
)
FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-review-v1.json"
)
FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_STATUS = (
    "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_defined"
)
FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_NEXT_DEPENDENCY = (
    "storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review"
)
FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-concrete-managed-database-provider-selection-review-v1.json"
)
FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review_defined"
)
FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_NEXT_DEPENDENCY = (
    "storage_adapter_provider_account_resource_endpoint_readiness"
)
FOLLOWUP_SELECTION_ALIGNMENT = {
    "audit_store_storage_adapter_backend_product_selection_review_status": FOLLOWUP_SELECTION_STATUS,
    "audit_storage_adapter_backend_product_selection_status": "selected_static_product_class_without_backend_provider",
    "audit_storage_adapter_selected_backend_product_class": "managed_database_append_only_table",
    "audit_storage_adapter_selected_backend_product_profile": "reserved_managed_database_append_only_table_profile",
    "audit_storage_adapter_database_product_status": "engine_selected_without_managed_product",
    "audit_storage_adapter_database_connection_provider_status": "not_created",
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_SELECTION_NEXT_DEPENDENCY,
}
FOLLOWUP_AFTER_SELECTION_ALIGNMENT = {
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status": (
        FOLLOWUP_AFTER_SELECTION_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_entry_refresh"
    ),
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY,
    "audit_storage_adapter_database_provider_driver_dsn_tls_role_policy_status": (
        "defined_without_runtime"
    ),
    "audit_storage_adapter_append_only_table_schema_boundary_status": "defined_without_sql_or_runtime",
    "audit_storage_adapter_migration_schema_marker_boundary_status": "logical_schema_marker_handoff_boundary_defined",
    "audit_storage_adapter_offline_adapter_smoke_strategy_status": "required_before_runtime_task_card",
    "audit_storage_adapter_negative_leakage_runtime_scan_boundary_status": "defined_without_runtime",
}
FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_ALIGNMENT = {
    "audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_status": (
        FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_database_provider_connection_runtime_boundary_readiness"
    ),
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_NEXT_DEPENDENCY,
    "audit_storage_adapter_database_provider_connection_runtime_boundary_status": (
        "metadata_only_boundary_defined_without_runtime"
    ),
}
FOLLOWUP_AFTER_PROVIDER_BOUNDARY_ALIGNMENT = {
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_status": (
        FOLLOWUP_AFTER_PROVIDER_BOUNDARY_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_database_provider_connection_runtime_boundary_entry_refresh"
    ),
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_AFTER_PROVIDER_BOUNDARY_NEXT_DEPENDENCY,
}
FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_ALIGNMENT = {
    "audit_store_storage_adapter_managed_database_product_selection_readiness_status": (
        FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_managed_database_product_selection_review"
    ),
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_NEXT_DEPENDENCY,
    "audit_storage_adapter_managed_product_selection_status": "selected_reference_product_profile_without_vendor",
    "audit_storage_adapter_managed_product_selection_review_status": "not_started",
}
FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_ALIGNMENT = {
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_managed_database_product_selection_review_status": (
        FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_managed_database_product_selection_review_entry_refresh"
    ),
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_NEXT_DEPENDENCY,
    "audit_storage_adapter_managed_product_selection_status": "selected_reference_product_profile_without_vendor",
    "audit_storage_adapter_managed_product_selection_review_status": FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_STATUS,
    "audit_storage_adapter_selected_managed_product_profile": "managed_postgresql_compatible_audit_store_profile",
    "audit_storage_adapter_managed_database_product_status": "selected_reference_profile_not_vendor_product",
}
FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_ALIGNMENT = {
    "audit_store_storage_adapter_concrete_managed_database_provider_selection_readiness_status": (
        FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_readiness"
    ),
    "audit_storage_adapter_current_next_dependency": (
        FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_NEXT_DEPENDENCY
    ),
}
FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_ALIGNMENT = {
    "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_status": (
        FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_review"
    ),
    "audit_storage_adapter_current_next_dependency": (
        FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_NEXT_DEPENDENCY
    ),
}
FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_ALIGNMENT = {
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review_status": (
        FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_review_entry_refresh"
    ),
    "audit_storage_adapter_current_next_dependency": (
        FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_NEXT_DEPENDENCY
    ),
}

POSITIVE_FIXTURE = "scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-positive-v1.json"
MISSING_REQUIRED_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-metadata-contract-missing-required-negative-v1.json"
)
FORBIDDEN_FIELD_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-metadata-contract-forbidden-field-negative-v1.json"
)
ADDITIONAL_PROPERTIES_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-metadata-contract-additional-properties-negative-v1.json"
)
WRITER_COMPATIBILITY_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-metadata-contract-writer-compatibility-v1.json"
)

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1.json",
        ENTRY_REVIEW_STATUS,
    ),
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1.json",
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1.json",
        "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.json",
        "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.json",
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json",
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.json",
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.json",
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined",
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

EXPECTED_ALLOWED_ARTIFACTS = {
    "contracts/production-secret-audit-storage-adapter.metadata-contract.json",
    "docs/contracts/production-secret-audit-storage-adapter-metadata-contract.md",
    "docs/platform/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json",
    POSITIVE_FIXTURE,
    MISSING_REQUIRED_FIXTURE,
    FORBIDDEN_FIELD_FIXTURE,
    ADDITIONAL_PROPERTIES_FIXTURE,
    WRITER_COMPATIBILITY_FIXTURE,
    "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.py",
}

EXPECTED_RUNTIME_FALSE_FLAGS = {
    "backend_product_selected_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "storage_adapter_client_created_in_this_slice",
    "database_connection_provider_enabled",
    "database_connection_created_in_this_slice",
    "sql_migration_created_in_this_slice",
    "schema_marker_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
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


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def followup_selection_exists() -> bool:
    if not FOLLOWUP_SELECTION_FIXTURE_PATH.exists():
        return False
    selection = load_json(FOLLOWUP_SELECTION_FIXTURE_PATH)
    return source_status(selection) == FOLLOWUP_SELECTION_STATUS


def followup_after_selection_exists() -> bool:
    if not FOLLOWUP_AFTER_SELECTION_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_AFTER_SELECTION_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_AFTER_SELECTION_STATUS


def followup_connection_runtime_boundary_exists() -> bool:
    if not FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_STATUS


def followup_after_provider_boundary_exists() -> bool:
    if not FOLLOWUP_AFTER_PROVIDER_BOUNDARY_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_AFTER_PROVIDER_BOUNDARY_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_AFTER_PROVIDER_BOUNDARY_STATUS


def followup_managed_product_selection_readiness_exists() -> bool:
    if not FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_STATUS


def followup_after_managed_product_selection_review_exists() -> bool:
    if not FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_STATUS


def followup_concrete_managed_database_provider_selection_readiness_exists() -> bool:
    if not FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_STATUS


def followup_concrete_managed_database_provider_selection_review_exists() -> bool:
    if not FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_STATUS


def followup_after_concrete_managed_database_provider_selection_review_exists() -> bool:
    if not FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_STATUS


def recursive_keys(value: Any) -> set[str]:
    if isinstance(value, dict):
        keys = set(value)
        for child in value.values():
            keys.update(recursive_keys(child))
        return keys
    if isinstance(value, list):
        keys: set[str] = set()
        for child in value:
            keys.update(recursive_keys(child))
        return keys
    return set()


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_metadata_contract_artifact_materialization_v1",
        "fixture kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "slice id drifted")
    require(slice_info.get("track") == "Production Ops Hardening v1", "track drifted")
    require(slice_info.get("status") == SLICE_STATUS, "slice status drifted")
    require(slice_info.get("contract_artifact") == CONTRACT_PATH.relative_to(REPO_ROOT).as_posix(), "contract path drifted")
    for field in ("task_card", "platform_topic", "contract_artifact"):
        relative_path = str(slice_info.get(field) or "")
        require((REPO_ROOT / relative_path).exists(), f"slice artifact missing: {relative_path}")
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "backend_product_selected",
        "storage_adapter_runtime_created",
        "database_connection_provider_ready",
        "audit_store_runtime_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in forbidden_claims, f"missing forbidden claim guard: {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    deps = rows_by_id(fixture, "depends_on", "id")
    require(set(deps) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dep_id, (evidence, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = deps[dep_id]
        require(item.get("status") == expected_status, f"{dep_id} status drifted")
        require(item.get("evidence") == evidence, f"{dep_id} evidence path drifted")
        dependency = load_json(REPO_ROOT / evidence)
        require(source_status(dependency) == expected_status, f"{dep_id} source status drifted")


def assert_materialization_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("materialization_boundary") or {}
    require(boundary.get("status") == SLICE_STATUS, "materialization boundary status drifted")
    require(boundary.get("contract_version") == CONTRACT_VERSION, "contract version drifted")
    require(boundary.get("contract_artifact_path") == CONTRACT_PATH.relative_to(REPO_ROOT).as_posix(), "contract artifact path drifted")
    for field in (
        "contract_artifact_materialized_in_this_slice",
        "positive_fixture_created_in_this_slice",
        "negative_fixtures_created_in_this_slice",
        "writer_compatibility_smoke_created_in_this_slice",
    ):
        require(boundary.get(field) is True, f"{field} must be true")
    for field in EXPECTED_RUNTIME_FALSE_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")
    require(boundary.get("current_next_dependency") == NEXT_DEPENDENCY, "next dependency drifted")
    for path in (
        "positive_fixture",
        "missing_required_negative_fixture",
        "forbidden_field_negative_fixture",
        "additional_properties_negative_fixture",
        "writer_compatibility_fixture",
    ):
        relative_path = str(boundary.get(path) or "")
        require((REPO_ROOT / relative_path).exists(), f"boundary fixture missing: {relative_path}")


def assert_contract_artifact(contract: dict[str, Any]) -> None:
    readiness = load_json(READINESS_FIXTURE_PATH)
    require(contract.get("schema_version") == 1, "contract schema_version drifted")
    require(contract.get("kind") == "production_secret_audit_storage_adapter_metadata_contract_v1", "contract kind drifted")
    require(contract.get("contract_version") == CONTRACT_VERSION, "contract version drifted")
    require(contract.get("status") == "metadata_contract_artifact_materialized_static", "contract status drifted")
    require(
        set((contract.get("input_envelope") or {}).get("required_fields") or [])
        == set(readiness.get("input_envelope_fields") or []),
        "input envelope fields drifted from readiness",
    )
    require(
        set((contract.get("result_envelope") or {}).get("required_fields") or [])
        == set(readiness.get("result_envelope_fields") or []),
        "result envelope fields drifted from readiness",
    )
    require(
        set((contract.get("record_identity") or {}).get("required_fields") or [])
        == set(readiness.get("record_identity_fields") or []),
        "record identity fields drifted from readiness",
    )
    writer_contract = readiness.get("writer_compatibility_contract") or {}
    require(
        set((contract.get("writer_compatibility") or {}).get("writer_output_required_fields") or [])
        == set(writer_contract.get("writer_output_allowed_fields") or []),
        "writer compatibility fields drifted from readiness",
    )
    require(
        set((contract.get("failure_taxonomy") or {}).get("families") or [])
        == set(readiness.get("failure_taxonomy_families") or []),
        "failure taxonomy families drifted from readiness",
    )
    guard = contract.get("artifact_guard") or {}
    expected_guard = {
        "backend_product_selection_status": "not_selected",
        "storage_adapter_runtime_status": "not_created",
        "database_connection_provider_status": "blocked",
        "audit_store_runtime_status": "not_created",
        "repository_mode_status": "disabled",
        "production_api_status": "not_created",
    }
    for field, expected in expected_guard.items():
        require(guard.get(field) == expected, f"contract artifact guard {field} drifted")


def validate_candidate(contract: dict[str, Any], candidate: dict[str, Any]) -> bool:
    forbidden = set(contract.get("forbidden_fields") or [])
    if forbidden & recursive_keys(candidate):
        return False
    if candidate.get("contract_version") != CONTRACT_VERSION:
        return False

    expected_sections = {
        "input_envelope": set((contract.get("input_envelope") or {}).get("required_fields") or []),
        "result_envelope": set((contract.get("result_envelope") or {}).get("required_fields") or []),
        "record_identity": set((contract.get("record_identity") or {}).get("required_fields") or []),
        "writer_output": set((contract.get("writer_compatibility") or {}).get("writer_output_required_fields") or []),
    }
    for section, expected_fields in expected_sections.items():
        value = candidate.get(section)
        if not isinstance(value, dict):
            return False
        if set(value) != expected_fields:
            return False

    write_status = (candidate.get("result_envelope") or {}).get("write_status")
    if write_status not in set((contract.get("result_envelope") or {}).get("write_status_allowlist") or []):
        return False
    if (candidate.get("input_envelope") or {}).get("storage_adapter_contract_version") != CONTRACT_VERSION:
        return False
    return True


def assert_validation_cases(fixture: dict[str, Any], contract: dict[str, Any]) -> None:
    expected_cases = {
        "positive_metadata_contract_candidate": (POSITIVE_FIXTURE, True),
        "missing_required_audit_event_ref": (MISSING_REQUIRED_FIXTURE, False),
        "forbidden_secret_material_field": (FORBIDDEN_FIELD_FIXTURE, False),
        "additional_property_in_input_envelope": (ADDITIONAL_PROPERTIES_FIXTURE, False),
    }
    cases = rows_by_id(fixture, "validation_case_matrix", "case")
    require(set(cases) == set(expected_cases), "validation case ids drifted")
    for case_id, (relative_path, expected_valid) in expected_cases.items():
        row = cases[case_id]
        require(row.get("fixture") == relative_path, f"{case_id} fixture path drifted")
        require(row.get("expected_valid") is expected_valid, f"{case_id} expected result drifted")
        candidate = load_json(REPO_ROOT / relative_path)
        require(candidate.get("expected_valid") is expected_valid, f"{case_id} fixture expectation drifted")
        actual = validate_candidate(contract, candidate)
        require(actual is expected_valid, f"{case_id} validation result drifted: expected {expected_valid}, got {actual}")


def assert_writer_compatibility(fixture: dict[str, Any], contract: dict[str, Any]) -> None:
    smoke = fixture.get("writer_compatibility_smoke") or {}
    require(smoke.get("status") == "implemented_static_fixture", "writer compatibility smoke status drifted")
    require(smoke.get("fixture") == WRITER_COMPATIBILITY_FIXTURE, "writer compatibility fixture path drifted")
    require(smoke.get("does_not_create_runtime") is True, "writer compatibility smoke created runtime")
    writer_fixture = load_json(REPO_ROOT / WRITER_COMPATIBILITY_FIXTURE)
    required = set((contract.get("writer_compatibility") or {}).get("writer_output_required_fields") or [])
    consumed = set((contract.get("writer_compatibility") or {}).get("storage_adapter_input_fields_consumed_from_writer") or [])
    require(set(smoke.get("requires_fields") or []) == required, "writer smoke required fields drifted")
    require(set(writer_fixture.get("writer_output_projection") or {}) == required, "writer output projection drifted")
    require(set(writer_fixture.get("storage_adapter_input_consumes") or []) == consumed, "writer input consumption drifted")
    require(writer_fixture.get("expected_valid") is True, "writer compatibility fixture expectation drifted")
    forbidden_runtime = set(writer_fixture.get("does_not_create") or [])
    for runtime in {"writer_runtime", "storage_adapter_runtime", "audit_store_runtime", "delivery_runtime", "idempotency_runtime"}:
        require(runtime in forbidden_runtime, f"writer compatibility missing runtime guard: {runtime}")


def assert_no_secret_material_scan(fixture: dict[str, Any]) -> None:
    scan = fixture.get("no_secret_material_scan") or {}
    require(scan.get("status") == "implemented_static_scan", "no secret material scan status drifted")
    patterns = list(scan.get("forbidden_literal_patterns") or [])
    require(patterns, "forbidden literal patterns missing")
    for relative_path in scan.get("scanned_artifacts") or []:
        text = read(str(relative_path))
        found = [pattern for pattern in patterns if pattern in text]
        require(not found, f"{relative_path} contains forbidden literal pattern: {found}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    allowed = set(guard.get("allowed_added_artifacts") or [])
    require(EXPECTED_ALLOWED_ARTIFACTS <= allowed, "allowed added artifacts missing expected paths")
    for path in allowed:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing: {path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "backend_product_selection_artifact",
        "storage_adapter_runtime_implementation_task_card",
        "storage_adapter_runtime",
        "database_connection_provider",
        "audit_store_runtime",
        "production_resolver_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact kind missing: {artifact}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden runtime artifact exists: {relative_path}")


def assert_entry_review_alignment() -> None:
    entry = load_json(ENTRY_REVIEW_FIXTURE_PATH)
    boundary = entry.get("entry_review_boundary") or {}
    require(boundary.get("status") == ENTRY_REVIEW_STATUS, "entry review status drifted")
    require(
        boundary.get("materialization_task_card_decision")
        == "metadata_contract_artifact_materialization_task_card_ready_after_entry_review",
        "entry review decision drifted",
    )


def assert_blocker_matrix_alignment(fixture: dict[str, Any]) -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    alignment = fixture.get("implementation_readiness_alignment") or {}
    require(
        boundary.get("storage_adapter_metadata_contract_artifact_materialization_status") == SLICE_STATUS,
        "matrix missing materialization status",
    )
    require(
        boundary.get("storage_adapter_contract_artifact_materialization_status")
        == alignment["audit_storage_adapter_contract_artifact_materialization_status"],
        "matrix materialization status drifted",
    )
    require(
        boundary.get("storage_adapter_contract_artifact_materialization_task_card_status") == "created",
        "matrix task card status drifted",
    )
    if followup_selection_exists():
        expected_next_dependency = (
            FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_NEXT_DEPENDENCY
            if followup_after_concrete_managed_database_provider_selection_review_exists()
            else FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_NEXT_DEPENDENCY
            if followup_concrete_managed_database_provider_selection_review_exists()
            else FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_NEXT_DEPENDENCY
            if followup_concrete_managed_database_provider_selection_readiness_exists()
            else FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_NEXT_DEPENDENCY
            if followup_after_managed_product_selection_review_exists()
            else FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_NEXT_DEPENDENCY
            if followup_managed_product_selection_readiness_exists()
            else FOLLOWUP_AFTER_PROVIDER_BOUNDARY_NEXT_DEPENDENCY
            if followup_after_provider_boundary_exists()
            else FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_NEXT_DEPENDENCY
            if followup_connection_runtime_boundary_exists()
            else FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY
            if followup_after_selection_exists()
            else FOLLOWUP_SELECTION_NEXT_DEPENDENCY
        )
        require(
            boundary.get("storage_adapter_current_next_dependency") == expected_next_dependency,
            "matrix next dependency drifted after selection follow-up",
        )
        require(
            boundary.get("storage_adapter_backend_product_selection_review_status") == FOLLOWUP_SELECTION_STATUS,
            "matrix selection review status drifted after follow-up",
        )
        require(
            boundary.get("storage_adapter_backend_product_selection_status")
            == "selected_static_product_class_without_backend_provider",
            "matrix selected product class status drifted after follow-up",
        )
        require(
            boundary.get("storage_adapter_selected_backend_product_class") == "managed_database_append_only_table",
            "matrix selected product class drifted after follow-up",
        )
        require(
            boundary.get("storage_adapter_selected_backend_product_profile")
            == "reserved_managed_database_append_only_table_profile",
            "matrix selected product profile drifted after follow-up",
        )
        if followup_after_selection_exists():
            require(
                boundary.get("storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status")
                == FOLLOWUP_AFTER_SELECTION_STATUS,
                "matrix after-selection refresh status drifted",
            )
    else:
        require(boundary.get("storage_adapter_current_next_dependency") == NEXT_DEPENDENCY, "matrix next dependency drifted")
        require(boundary.get("storage_adapter_backend_product_selection_status") == "not_selected", "matrix selected product")
    require(boundary.get("storage_adapter_runtime_status") == "not_created", "matrix created runtime")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in (fixture.get("implementation_readiness_alignment") or {}).items():
        if followup_selection_exists() and field in FOLLOWUP_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_SELECTION_ALIGNMENT[field]
        if followup_after_selection_exists() and field in FOLLOWUP_AFTER_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_AFTER_SELECTION_ALIGNMENT[field]
        if followup_connection_runtime_boundary_exists() and field in FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_ALIGNMENT:
            expected = FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_ALIGNMENT[field]
        if followup_after_provider_boundary_exists() and field in FOLLOWUP_AFTER_PROVIDER_BOUNDARY_ALIGNMENT:
            expected = FOLLOWUP_AFTER_PROVIDER_BOUNDARY_ALIGNMENT[field]
        if (
            followup_managed_product_selection_readiness_exists()
            and field in FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_ALIGNMENT
        ):
            expected = FOLLOWUP_MANAGED_PRODUCT_SELECTION_READINESS_ALIGNMENT[field]
        if (
            followup_after_managed_product_selection_review_exists()
            and field in FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_ALIGNMENT
        ):
            expected = FOLLOWUP_AFTER_MANAGED_PRODUCT_SELECTION_REVIEW_ALIGNMENT[field]
        if (
            followup_concrete_managed_database_provider_selection_readiness_exists()
            and field in FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_ALIGNMENT
        ):
            expected = FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_READINESS_ALIGNMENT[field]
        if (
            followup_concrete_managed_database_provider_selection_review_exists()
            and field in FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_ALIGNMENT
        ):
            expected = FOLLOWUP_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_ALIGNMENT[field]
        if (
            followup_after_concrete_managed_database_provider_selection_review_exists()
            and field in FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_ALIGNMENT
        ):
            expected = FOLLOWUP_AFTER_CONCRETE_MANAGED_DATABASE_PROVIDER_SELECTION_REVIEW_ALIGNMENT[field]
        require(target.get(field) == expected, f"implementation readiness {field} drifted")
    planned = rows_by_id(readiness, "planned_slices", "id")
    item = planned.get("audit-store-storage-adapter-metadata-contract-artifact-materialization") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness missing materialization planned slice")
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        "contracts/README.md": [
            "production-secret-audit-storage-adapter.metadata-contract.json",
            "Production Secret Audit Storage Adapter Metadata Contract",
        ],
        "docs/contracts/README.md": [
            "Production Secret Audit Storage Adapter Metadata Contract",
            "production-secret-audit-storage-adapter-metadata-contract.md",
        ],
        "docs/contracts/production-secret-audit-storage-adapter-metadata-contract.md": [
            CONTRACT_VERSION,
            POSITIVE_FIXTURE,
            WRITER_COMPATIBILITY_FIXTURE,
            "不选择具体 backend product",
        ],
        "docs/radishmind-integration-contracts.md": [
            "production-secret-audit-storage-adapter.metadata-contract.json",
            SLICE_STATUS,
        ],
        "docs/platform/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.md": [
            SLICE_STATUS,
            CONTRACT_VERSION,
            NEXT_DEPENDENCY,
        ],
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1-plan.md": [
            SLICE_STATUS,
            CONTRACT_VERSION,
            "停止线",
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            SLICE_STATUS,
            NEXT_DEPENDENCY,
        ],
        "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": [
            SLICE_STATUS,
            NEXT_DEPENDENCY,
        ],
        "docs/platform/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Metadata Contract Artifact Materialization v1",
            SLICE_STATUS,
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Metadata Contract Artifact Materialization v1",
            SLICE_STATUS,
        ],
        "docs/features/workflow/README.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/radishmind-current-focus.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [SLICE_ID, SLICE_STATUS],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.py",
            "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json",
        ],
        "docs/devlogs/2026-W27.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    }
    for relative_path, literals in docs.items():
        text = read(relative_path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    entry_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-storage-adapter-'
        'metadata-contract-artifact-materialization-entry-review-v1.py", [])'
    )
    materialization_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-storage-adapter-'
        'metadata-contract-artifact-materialization-v1.py", [])'
    )
    matrix_call = 'run_python_script("check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py", [])'
    require(entry_call in check_repo, "check-repo missing materialization entry review checker")
    require(materialization_call in check_repo, "check-repo missing materialization checker")
    require(matrix_call in check_repo, "check-repo missing runtime blocker matrix checker")
    require(
        check_repo.index(entry_call) < check_repo.index(materialization_call) < check_repo.index(matrix_call),
        "check-repo materialization checker order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    contract = load_json(CONTRACT_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_materialization_boundary(fixture)
    assert_contract_artifact(contract)
    assert_validation_cases(fixture, contract)
    assert_writer_compatibility(fixture, contract)
    assert_no_secret_material_scan(fixture)
    assert_artifact_guard(fixture)
    assert_entry_review_alignment()
    assert_blocker_matrix_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    print("production ops secret backend audit store storage adapter metadata contract artifact materialization checks passed.")


if __name__ == "__main__":
    main()
