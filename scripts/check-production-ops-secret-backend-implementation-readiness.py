#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_FORBIDDEN_CLAIMS = {
    "production_ready",
    "production_secret_backend_ready",
    "cloud_secret_service_ready",
    "cloud_secret_service_selected",
    "real_secret_written",
    "resolver_implemented",
    "fake_resolver_implemented",
    "no_secret_leakage_smoke_runtime_created",
    "backend_health_runtime_created",
    "backend_health_runtime_task_card_created",
    "backend_health_client_created",
    "backend_health_check_executed",
    "secret_rotation_ready",
    "production_secret_audit_store_ready",
}

REQUIRED_PRECONDITIONS = {
    "secret-ref-schema",
    "config-injection-point",
    "provider-profile-binding",
    "sanitized-audit-fields",
    "failure-taxonomy",
    "test-fixture-strategy",
    "operator-runbook",
    "rotation-and-audit-policy",
}

REQUIRED_PLANNED_SLICES = {
    "secret-ref-schema-and-fixtures": "satisfied",
    "config-secret-ref-readiness": "satisfied",
    "provider-profile-secret-binding": "satisfied",
    "secret-resolver-interface-disabled": "satisfied",
    "operator-runbook-and-negative-gates": "satisfied",
    "rotation-and-audit-policy": "satisfied",
    "test-fixture-strategy": "satisfied_for_test_only_fake_resolver",
    "fake-resolver-contract-no-secret-leakage-smoke-strategy": "strategy_defined_static_only",
    "fake-resolver-implementation-task-card-entry-readiness-review": "ready_for_next_task",
    "fake-resolver-implementation": "task_card_defined_runtime_not_started",
    "fake-resolver-runtime-implementation-entry-review": "ready_for_runtime_task",
    "fake-resolver-runtime-implementation": "fake_resolver_runtime_test_only_implemented",
    "real-resolver-runtime-preconditions": "real_resolver_runtime_preconditions_defined",
    "real-resolver-runtime-implementation-entry-review": "real_resolver_runtime_implementation_entry_review_defined",
    "real-resolver-runtime-implementation-entry-refresh": (
        "real_resolver_runtime_implementation_entry_refresh_defined"
    ),
    "resolver-backend-profile-selection-readiness": "resolver_backend_profile_selection_readiness_defined",
    "real-resolver-no-secret-leakage-smoke-runtime-strategy": (
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
    ),
    "real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review": (
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined"
    ),
    "real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh": (
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined"
    ),
    "credential-handle-runtime-boundary-readiness": "credential_handle_runtime_boundary_readiness_defined",
    "credential-handle-runtime-implementation-entry-review": (
        "credential_handle_runtime_implementation_entry_review_defined"
    ),
    "credential-handle-runtime-implementation-entry-refresh": (
        "credential_handle_runtime_implementation_entry_refresh_defined"
    ),
    "operator-approval-runtime-evidence-readiness": "operator_approval_runtime_evidence_readiness_defined",
    "audit-store-handoff-readiness": "audit_store_handoff_readiness_defined",
    "audit-store-runtime-implementation-entry-review": "audit_store_runtime_implementation_entry_review_defined",
    "audit-store-contract-event-schema-readiness": "audit_store_contract_event_schema_readiness_defined",
    "audit-store-runtime-implementation-entry-refresh-v2": (
        "audit_store_runtime_implementation_entry_refresh_v2_defined"
    ),
    "audit-store-ownership-boundary-readiness": "audit_store_ownership_boundary_readiness_defined",
    "audit-store-delivery-idempotency-runtime-boundary-readiness": (
        "audit_store_delivery_idempotency_runtime_boundary_readiness_defined"
    ),
    "audit-store-runtime-implementation-entry-refresh-v3": (
        "audit_store_runtime_implementation_entry_refresh_v3_defined"
    ),
    "audit-store-durable-backend-boundary-readiness": (
        "audit_store_durable_backend_boundary_readiness_defined"
    ),
    "audit-store-writer-runtime-boundary-readiness": (
        "audit_store_writer_runtime_boundary_readiness_defined"
    ),
    "audit-store-runtime-event-schema-materialization-readiness": (
        "audit_store_runtime_event_schema_materialization_readiness_defined"
    ),
    "audit-store-delivery-runtime-readiness": "audit_store_delivery_runtime_readiness_defined",
    "audit-store-idempotency-runtime-readiness": "audit_store_idempotency_runtime_readiness_defined",
    "audit-store-runtime-implementation-entry-refresh-v4": (
        "audit_store_runtime_implementation_entry_refresh_v4_defined"
    ),
    "audit-store-runtime-event-schema-artifact-implementation-entry-review": (
        "audit_store_runtime_event_schema_artifact_implementation_entry_review_defined"
    ),
    "audit-store-runtime-event-schema-artifact-implementation": (
        "audit_store_runtime_event_schema_artifact_implementation_task_card_defined"
    ),
    "audit-store-runtime-event-schema-artifact": (
        "audit_store_runtime_event_schema_artifact_implemented"
    ),
    "audit-store-runtime-blocker-matrix": "audit_store_runtime_blocker_matrix_defined",
    "audit-store-durable-backend-selection-readiness": (
        "audit_store_durable_backend_selection_readiness_defined"
    ),
    "audit-store-concrete-durable-backend-selection-review": (
        "audit_store_concrete_durable_backend_selection_review_defined"
    ),
    "audit-store-writer-runtime-implementation-entry-review": (
        "audit_store_writer_runtime_implementation_entry_review_defined"
    ),
    "audit-store-idempotency-runtime-implementation-entry-review": (
        "audit_store_idempotency_runtime_implementation_entry_review_defined"
    ),
    "audit-store-delivery-runtime-implementation-entry-review": (
        "audit_store_delivery_runtime_implementation_entry_review_defined"
    ),
    "audit-store-storage-adapter-runtime-implementation-entry-review": (
        "audit_store_storage_adapter_runtime_implementation_entry_review_defined"
    ),
    "audit-store-storage-adapter-backend-product-evidence-readiness": (
        "audit_store_storage_adapter_backend_product_evidence_readiness_defined"
    ),
    "audit-store-storage-adapter-metadata-contract-artifact-readiness": (
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined"
    ),
    "audit-store-storage-adapter-append-only-semantics-evidence-readiness": (
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined"
    ),
    "audit-store-storage-adapter-retention-redaction-policy-evidence-readiness": (
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined"
    ),
    "audit-store-storage-adapter-offline-validation-evidence-readiness": (
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined"
    ),
    "audit-store-storage-adapter-negative-leakage-scan-evidence-readiness": (
        "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined"
    ),
    "audit-store-storage-adapter-rollback-recovery-evidence-readiness": (
        "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined"
    ),
    "audit-store-storage-adapter-runtime-implementation-entry-refresh": (
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_defined"
    ),
    "audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review": (
        "audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_defined"
    ),
    "audit-store-storage-adapter-metadata-contract-artifact-materialization": (
        "audit_store_storage_adapter_metadata_contract_artifact_materialized"
    ),
    "audit-store-storage-adapter-backend-product-selection-review": (
        "audit_store_storage_adapter_backend_product_selection_review_defined"
    ),
    "audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection": (
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined"
    ),
    "audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness": (
        "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined"
    ),
    "audit-store-storage-adapter-append-only-table-schema-boundary-readiness": (
        "audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined"
    ),
    "audit-store-storage-adapter-table-schema-artifact-materialization-entry-review": (
        "audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_defined"
    ),
    "audit-store-storage-adapter-table-schema-artifact-materialization": (
        "audit_store_storage_adapter_table_schema_artifact_materialized"
    ),
    "audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness": (
        "audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined"
    ),
    "audit-store-storage-adapter-negative-leakage-runtime-scan-boundary-readiness": (
        "audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined"
    ),
    "audit-store-storage-adapter-runtime-implementation-entry-refresh-after-negative-leakage-runtime-scan-boundary": (
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary_defined"
    ),
    "audit-store-storage-adapter-concrete-database-selection-readiness": (
        "audit_store_storage_adapter_concrete_database_selection_readiness_defined"
    ),
    "audit-store-storage-adapter-concrete-database-selection-review": (
        "audit_store_storage_adapter_concrete_database_selection_review_defined"
    ),
    "audit-store-storage-adapter-database-provider-selection-readiness": (
        "audit_store_storage_adapter_database_provider_selection_readiness_defined"
    ),
    "audit-store-storage-adapter-database-provider-selection-review": (
        "audit_store_storage_adapter_database_provider_selection_review_defined"
    ),
    "audit-store-storage-adapter-database-driver-selection-readiness": (
        "audit_store_storage_adapter_database_driver_selection_readiness_defined"
    ),
    "audit-store-storage-adapter-database-driver-selection-review": (
        "audit_store_storage_adapter_database_driver_selection_review_defined"
    ),
    "audit-store-storage-adapter-database-connection-lifecycle-readiness": (
        "audit_store_storage_adapter_database_connection_lifecycle_readiness_defined"
    ),
    "audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle": (
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_defined"
    ),
    "audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness": (
        "audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_defined"
    ),
    "audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-provider-connection-runtime-boundary": (
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_defined"
    ),
    "audit-store-storage-adapter-managed-database-product-selection-readiness": (
        "audit_store_storage_adapter_managed_database_product_selection_readiness_defined"
    ),
    "audit-store-storage-adapter-managed-database-product-selection-review": (
        "audit_store_storage_adapter_managed_database_product_selection_review_defined"
    ),
    "audit-store-storage-adapter-concrete-managed-database-provider-selection-readiness": (
        "audit_store_storage_adapter_concrete_managed_database_provider_selection_readiness_defined"
    ),
    "audit-store-storage-adapter-concrete-managed-database-provider-selection-review": (
        "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_defined"
    ),
    "audit-store-storage-adapter-runtime-implementation-entry-refresh-after-concrete-managed-database-provider-selection-review": (
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review_defined"
    ),
    "audit-store-storage-adapter-provider-account-resource-endpoint-readiness": (
        "audit_store_storage_adapter_provider_account_resource_endpoint_readiness_defined"
    ),
    "resolver-backend-health-boundary-readiness": "resolver_backend_health_boundary_readiness_defined",
    "resolver-backend-health-runtime-implementation-entry-review": (
        "resolver_backend_health_runtime_implementation_entry_review_defined"
    ),
    "resolver-backend-health-runtime-implementation-entry-refresh": (
        "resolver_backend_health_runtime_implementation_entry_refresh_defined"
    ),
    "operator-approval-runtime-implementation-entry-review": (
        "operator_approval_runtime_implementation_entry_review_defined"
    ),
    "operator-approval-runtime-implementation-entry-refresh": (
        "operator_approval_runtime_implementation_entry_refresh_defined"
    ),
    "cloud-secret-service-selection-readiness": "cloud_secret_service_selection_readiness_defined",
}

REQUIRED_BLOCKED = {
    "production_secret_backend": "not_satisfied",
    "cloud_secret_service_integration": "not_satisfied",
    "test_fixture_strategy": "satisfied_for_test_only_fake_resolver",
    "fake_resolver_implementation": "test_only_runtime_implemented_disabled_by_default",
    "real_secret_values": "forbidden_in_committed_repo",
    "production_ready": "not_satisfied",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "production-secret-backend-contract",
        "secret-ref-schema",
        "secret-ref-schema-and-fixtures",
        "config-secret-ref-readiness",
        "production-secret-backend-config-secret-ref-readiness-v1",
        "config_secret_ref_readiness_defined",
        "production-secret-backend-provider-profile-secret-binding-readiness-v1",
        "provider_profile_secret_binding_readiness_defined",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
        "secret_resolver_interface_disabled_readiness_defined",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
        "operator_runbook_negative_gates_readiness_defined",
        "production-secret-backend-rotation-audit-policy-readiness-v1",
        "rotation_audit_policy_readiness_defined",
        "production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1",
        "test_fixture_strategy_fake_resolver_entry_review_defined",
        "production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1",
        "fake_resolver_contract_no_secret_leakage_smoke_strategy_defined",
        "production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1",
        "fake_resolver_implementation_task_card_entry_readiness_review_defined",
        "production-secret-backend-fake-resolver-implementation-v1",
        "fake_resolver_implementation_task_card_defined",
        "production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1",
        "fake_resolver_runtime_implementation_entry_review_defined",
        "production-secret-backend-fake-resolver-runtime-implementation-v1",
        "fake_resolver_runtime_test_only_implemented",
        "production-secret-backend-real-resolver-runtime-preconditions-v1",
        "real_resolver_runtime_preconditions_defined",
        "production-secret-backend-real-resolver-runtime-implementation-entry-review-v1",
        "real_resolver_runtime_implementation_entry_review_defined",
        "production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1",
        "real_resolver_runtime_implementation_entry_refresh_defined",
        "production-secret-backend-resolver-backend-profile-selection-readiness-v1",
        "resolver_backend_profile_selection_readiness_defined",
        "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1",
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined",
        "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
        "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined",
        "production-secret-backend-credential-handle-runtime-boundary-readiness-v1",
        "credential_handle_runtime_boundary_readiness_defined",
        "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1",
        "credential_handle_runtime_implementation_entry_review_defined",
        "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1",
        "credential_handle_runtime_implementation_entry_refresh_defined",
        "production-secret-backend-operator-approval-runtime-evidence-readiness-v1",
        "operator_approval_runtime_evidence_readiness_defined",
        "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1",
        "operator_approval_runtime_implementation_entry_review_defined",
        "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1",
        "operator_approval_runtime_implementation_entry_refresh_defined",
        "production-secret-backend-cloud-secret-service-selection-readiness-v1",
        "cloud_secret_service_selection_readiness_defined",
        "production-secret-backend-audit-store-handoff-readiness-v1",
        "audit_store_handoff_readiness_defined",
        "production-secret-backend-audit-store-runtime-implementation-entry-review-v1",
        "audit_store_runtime_implementation_entry_review_defined",
        "production-secret-backend-audit-store-contract-event-schema-readiness-v1",
        "audit_store_contract_event_schema_readiness_defined",
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2",
        "audit_store_runtime_implementation_entry_refresh_v2_defined",
        "production-secret-backend-audit-store-ownership-boundary-readiness-v1",
        "audit_store_ownership_boundary_readiness_defined",
        "production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1",
        "audit_store_delivery_idempotency_runtime_boundary_readiness_defined",
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
        "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1",
        "audit_store_runtime_event_schema_materialization_readiness_defined",
        "production-secret-backend-audit-store-delivery-runtime-readiness-v1",
        "audit_store_delivery_runtime_readiness_defined",
        "production-secret-backend-audit-store-idempotency-runtime-readiness-v1",
        "audit_store_idempotency_runtime_readiness_defined",
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
        "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1",
        "audit_store_runtime_event_schema_artifact_implementation_entry_review_defined",
        "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1",
        "audit_store_runtime_event_schema_artifact_implementation_task_card_defined",
        "production-secret-backend-audit-store-runtime-event-schema-artifact-v1",
        "audit_store_runtime_event_schema_artifact_implemented",
        "contracts/production-secret-audit-event.schema.json",
        "production-secret-backend-audit-store-runtime-blocker-matrix-v1",
        "audit_store_runtime_blocker_matrix_defined",
        "production-secret-backend-audit-store-durable-backend-selection-readiness-v1",
        "audit_store_durable_backend_selection_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1",
        "audit_store_storage_adapter_backend_product_evidence_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1",
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1",
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1",
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1",
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1",
        "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1",
        "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1",
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_defined",
        "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1",
        "audit_store_storage_adapter_backend_product_selection_review_defined",
        "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1",
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined",
        "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1",
        "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1",
        "audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-negative-leakage-runtime-scan-boundary-v1",
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary_defined",
        "storage_adapter_concrete_database_selection_readiness",
        "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1",
        "audit_store_storage_adapter_concrete_database_selection_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1",
        "audit_store_storage_adapter_concrete_database_selection_review_defined",
        "storage_adapter_database_provider_selection_readiness",
        "production-secret-backend-audit-store-storage-adapter-database-provider-selection-readiness-v1",
        "audit_store_storage_adapter_database_provider_selection_readiness_defined",
        "storage_adapter_database_provider_selection_review",
        "production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1",
        "audit_store_storage_adapter_database_provider_selection_review_defined",
        "managed_postgresql_compatible_service",
        "production-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1",
        "audit_store_storage_adapter_database_driver_selection_readiness_defined",
        "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1",
        "audit_store_storage_adapter_database_driver_selection_review_defined",
        "github.com/jackc/pgx/v5",
        "storage_adapter_database_connection_lifecycle_readiness",
        "production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1",
        "audit_store_storage_adapter_database_connection_lifecycle_readiness_defined",
        "database_connection_lifecycle_readiness_defined_without_connection_runtime",
        "storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_entry_refresh",
        "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1",
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_defined",
        "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1",
        "audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_defined",
        "storage_adapter_runtime_task_card_still_blocked_after_database_provider_connection_runtime_boundary_readiness",
        "storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_readiness",
        "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-provider-connection-runtime-boundary-v1",
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_defined",
        "storage_adapter_runtime_task_card_still_blocked_after_database_provider_connection_runtime_boundary_entry_refresh",
        "storage_adapter_managed_database_product_selection_readiness",
        "production-secret-backend-audit-store-storage-adapter-managed-database-product-selection-readiness-v1",
        "audit_store_storage_adapter_managed_database_product_selection_readiness_defined",
        "managed_database_product_selection_readiness_defined_without_product_selection",
        "storage_adapter_runtime_task_card_still_blocked_after_managed_database_product_selection_readiness",
        "storage_adapter_managed_database_product_selection_review",
        "production-secret-backend-audit-store-storage-adapter-managed-database-product-selection-review-v1",
        "audit_store_storage_adapter_managed_database_product_selection_review_defined",
        "managed_database_product_profile_selected_reference_only_runtime_blocked",
        "managed_postgresql_compatible_audit_store_profile",
        "storage_adapter_runtime_task_card_still_blocked_after_managed_database_product_selection_review",
        "storage_adapter_runtime_implementation_entry_refresh_after_managed_database_product_selection_review",
        "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-managed-database-product-selection-review-v1",
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_managed_database_product_selection_review_defined",
        "storage_adapter_runtime_task_card_still_blocked_after_managed_database_product_selection_review_entry_refresh",
        "storage_adapter_concrete_managed_database_provider_selection_readiness",
        "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-readiness-v1",
        "audit_store_storage_adapter_concrete_managed_database_provider_selection_readiness_defined",
        "concrete_managed_database_provider_selection_readiness_defined_without_provider_selection",
        "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_readiness",
        "storage_adapter_concrete_managed_database_provider_selection_review",
        "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-review-v1",
        "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_defined",
        "concrete_managed_database_provider_reference_selected_runtime_blocked",
        "managed_postgresql_compatible_provider_reference",
        "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_review",
        "storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review",
        "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-concrete-managed-database-provider-selection-review-v1",
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review_defined",
        "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_review_entry_refresh",
        "storage_adapter_provider_account_resource_endpoint_readiness",
        "production-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1",
        "audit_store_storage_adapter_provider_account_resource_endpoint_readiness_defined",
        "storage_adapter_runtime_task_card_still_blocked_after_provider_account_resource_endpoint_readiness",
        "storage_adapter_provider_account_resource_endpoint_review",
        "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1",
        "audit_store_writer_runtime_implementation_entry_review_defined",
        "production-secret-backend-resolver-backend-health-boundary-readiness-v1",
        "resolver_backend_health_boundary_readiness_defined",
        "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
        "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1",
        "resolver_backend_health_runtime_implementation_entry_refresh_defined",
        "services/platform/internal/secretbackend/fake_resolver.go",
        "contracts/production-secret-reference.schema.json",
        "production-secret-reference-basic.json",
        "check-production-secret-reference-contract.py",
        "config-injection-point",
        "provider-profile-binding",
        "sanitized-audit-fields",
        "failure-taxonomy",
        "test-fixture-strategy",
        "operator-runbook",
        "rotation-and-audit-policy",
        "不直接写 resolver 代码",
        "不接真实云 secret 服务",
        "不写入真实 secret",
        "不声明 production ready",
    ],
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-implementation-readiness",
        "production-ops-secret-backend-implementation-readiness.json",
        "production-secret-backend-implementation-v1-plan.md",
    ],
    "docs/radishmind-roadmap.md": [
        "production-secret-backend-implementation-readiness",
        "production-ops-secret-backend-implementation-readiness.json",
    ],
    "docs/radishmind-capability-matrix.md": [
        "production secret backend implementation readiness",
        "production-ops-secret-backend-implementation-readiness.json",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "Production Secret Backend Implementation",
        "production-secret-backend-implementation-readiness",
        "production-ops-secret-backend-implementation-readiness.json",
    ],
    "services/platform/README.md": [
        "Production secret backend implementation readiness",
        "production-ops-secret-backend-implementation-readiness.json",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-implementation-readiness.py",
        "production-ops-secret-backend-implementation-readiness.json",
        "check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
        "check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py",
        "check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
        "check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
        "check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py",
        "check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py",
        "check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
        "check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py",
        "check-production-ops-secret-backend-fake-resolver-implementation-v1.py",
        "check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py",
        "check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py",
        "check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py",
        "check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py",
        "check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py",
        "check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.py",
        "check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py",
        "check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.py",
        "check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py",
        "check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.py",
        "check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py",
        "check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.py",
        "check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.py",
        "check-production-ops-secret-backend-audit-store-ownership-boundary-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py",
        "check-production-ops-secret-backend-audit-store-delivery-runtime-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-idempotency-runtime-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.py",
        "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.py",
        "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py",
        "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py",
        "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-negative-leakage-runtime-scan-boundary-v1.py",
        "check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1.py",
        "check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1.py",
        "check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py",
        "check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py",
    ],
    "docs/devlogs/2026-W22.md": [
        "production-secret-backend-implementation-readiness",
        "production-ops-secret-backend-implementation-readiness.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_fixture() -> dict[str, Any]:
    return json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))


def assert_slice(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "production-secret-backend-implementation-readiness", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "implementation_readiness_defined", "unexpected readiness status")
    task_card = str(slice_info.get("task_card") or "")
    require(task_card == "docs/task-cards/production-secret-backend-implementation-v1-plan.md", "unexpected task card")
    require((REPO_ROOT / task_card).exists(), "secret backend implementation task card must exist")
    missing_claims = sorted(REQUIRED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_implementation_target(fixture: dict[str, Any]) -> None:
    target = fixture.get("implementation_target") or {}
    require(target.get("first_backend") == "external_reference_resolver_adapter", "unexpected first backend")
    require(target.get("cloud_vendor_specific_backend") == "not_selected", "cloud vendor backend must not be selected")
    require(
        target.get("cloud_secret_service_selection_readiness_status")
        == "defined_without_cloud_backend_selection",
        "cloud secret service selection readiness status drifted",
    )
    require(target.get("committed_secret_storage") == "forbidden", "committed secret storage must be forbidden")
    require(
        target.get("production_secret_backend_status") == "not_satisfied",
        "production secret backend must remain not_satisfied",
    )
    require(target.get("resolver_implementation_status") == "not_started", "resolver must not be implemented")
    require(target.get("resolver_runtime_status") == "not_created", "resolver runtime must not be created")
    require(
        target.get("production_resolver_runtime_task_card_status") == "not_created",
        "production resolver runtime task card must not be created",
    )
    require(
        target.get("fake_resolver_status") == "test_only_runtime_implemented_disabled_by_default",
        "fake resolver status drifted",
    )
    require(
        target.get("fake_resolver_contract_status") == "static_contract_defined",
        "fake resolver contract strategy must be recorded",
    )
    require(
        target.get("no_secret_leakage_smoke_strategy_status") == "static_strategy_defined",
        "no leakage smoke strategy must be recorded",
    )
    require(
        target.get("no_secret_leakage_smoke_runtime_status") == "implemented_offline_go_test",
        "no leakage smoke runtime status drifted",
    )
    require(
        target.get("test_fixture_strategy_status") == "satisfied_for_test_only_fake_resolver",
        "test fixture strategy status drifted",
    )
    require(
        target.get("test_fixture_strategy_review_status") == "blocked_entry_review_defined",
        "test fixture strategy review status drifted",
    )
    require(
        target.get("fake_resolver_implementation_task_card_entry_review_status") == "ready_for_next_task",
        "fake resolver implementation task card entry review status drifted",
    )
    require(
        target.get("fake_resolver_implementation_task_card_status") == "created_static_task_card",
        "fake resolver implementation task card status drifted",
    )
    require(
        target.get("fake_resolver_implementation_status") == "task_card_defined_runtime_not_started",
        "fake resolver implementation status drifted",
    )
    require(
        target.get("fake_resolver_runtime_implementation_entry_review_status") == "ready_for_next_task",
        "fake resolver runtime implementation entry review status drifted",
    )
    require(
        target.get("fake_resolver_runtime_implementation_task_card_status") == "created_static_task_card",
        "fake resolver runtime implementation task card status drifted",
    )
    require(
        target.get("fake_resolver_runtime_implementation_status") == "fake_resolver_runtime_test_only_implemented",
        "fake resolver runtime implementation status drifted",
    )
    require(
        target.get("fake_resolver_runtime_status") == "implemented_test_only_disabled_by_default",
        "fake resolver runtime status drifted",
    )
    require(
        target.get("real_resolver_runtime_preconditions_status") == "real_resolver_runtime_preconditions_defined",
        "real resolver runtime preconditions status drifted",
    )
    require(
        target.get("real_resolver_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "real resolver runtime implementation entry review status drifted",
    )
    require(
        target.get("real_resolver_runtime_implementation_entry_refresh_status")
        == "blocked_before_runtime_task_card",
        "real resolver runtime implementation entry refresh status drifted",
    )
    require(
        target.get("production_resolver_runtime_status") == "not_created",
        "production resolver runtime must remain not_created",
    )
    require(
        target.get("resolver_backend_profile_selection_readiness_status") == "defined_without_backend_runtime",
        "resolver backend profile selection readiness status drifted",
    )
    require(
        target.get("cloud_secret_service_status") == "not_selected",
        "cloud secret service must remain not_selected",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_strategy_status") == "defined_without_runtime",
        "real resolver no leakage smoke runtime strategy status drifted",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "real resolver no leakage smoke runtime implementation entry review status drifted",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_status")
        == "blocked_before_runtime_task_card",
        "real resolver no leakage smoke runtime implementation entry refresh status drifted",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_status") == "not_created",
        "real resolver no leakage smoke runtime must remain not_created",
    )
    require(
        target.get("credential_handle_runtime_boundary_readiness_status") == "defined_without_runtime",
        "credential handle runtime boundary readiness status drifted",
    )
    require(
        target.get("credential_handle_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "credential handle runtime implementation entry review status drifted",
    )
    require(
        target.get("credential_handle_runtime_implementation_entry_refresh_status")
        == "blocked_before_runtime_task_card",
        "credential handle runtime implementation entry refresh status drifted",
    )
    require(
        target.get("credential_handle_runtime_status") == "not_created",
        "credential handle runtime must remain not_created",
    )
    require(target.get("credential_handle_status") == "not_created", "credential handle must remain not_created")
    require(target.get("credential_payload_status") == "forbidden", "credential payload must remain forbidden")
    require(
        target.get("operator_approval_runtime_evidence_readiness_status")
        == "defined_without_runtime_execution",
        "operator approval runtime evidence readiness status drifted",
    )
    require(
        target.get("operator_approval_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "operator approval runtime implementation entry review status drifted",
    )
    require(
        target.get("operator_approval_runtime_status") == "not_created",
        "operator approval runtime must remain not_created",
    )
    require(
        target.get("operator_approval_runtime_execution_status") == "not_executed",
        "operator approval runtime execution must remain not_executed",
    )
    require(
        target.get("audit_store_handoff_readiness_status") == "defined_without_store_runtime",
        "audit store handoff readiness status drifted",
    )
    require(
        target.get("audit_store_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "audit store runtime implementation entry review status drifted",
    )
    require(
        target.get("audit_store_contract_event_schema_readiness_status") == "defined_without_store_runtime",
        "audit store contract event schema readiness status drifted",
    )
    require(
        target.get("audit_store_runtime_implementation_entry_refresh_v2_status")
        == "blocked_before_runtime_task_card",
        "audit store runtime implementation entry refresh v2 status drifted",
    )
    require(
        target.get("audit_store_ownership_boundary_readiness_status") == "defined_without_store_runtime",
        "audit store ownership boundary readiness status drifted",
    )
    require(
        target.get("audit_store_owner_status") == "static_reference_defined",
        "audit store owner status drifted",
    )
    require(
        target.get("audit_store_durable_backend_boundary_readiness_status")
        == "defined_without_backend_selection",
        "audit store durable backend boundary readiness status drifted",
    )
    require(
        target.get("audit_store_durable_backend_owner_status") == "static_boundary_defined",
        "audit store durable backend owner status drifted",
    )
    require(
        target.get("audit_store_durable_backend_selection_readiness_status")
        == "defined_without_backend_selection",
        "audit store durable backend selection readiness status drifted",
    )
    require(
        target.get("audit_store_concrete_durable_backend_selection_review_status")
        == "audit_store_concrete_durable_backend_selection_review_defined",
        "audit store concrete durable backend selection review status drifted",
    )
    require(
        target.get("durable_audit_backend_status") == "static_backend_family_selected_runtime_blocked",
        "durable audit backend status drifted",
    )
    require(
        target.get("selected_durable_backend_family") == "append_only_metadata_audit_log",
        "selected durable backend family drifted",
    )
    require(
        target.get("selected_durable_backend_candidate") == "reserved_append_only_audit_log",
        "selected durable backend candidate drifted",
    )
    require(
        target.get("audit_store_writer_runtime_boundary_readiness_status") == "defined_without_writer_runtime",
        "audit store writer runtime boundary readiness status drifted",
    )
    require(
        target.get("audit_store_writer_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "audit store writer runtime implementation entry review status drifted",
    )
    require(
        target.get("audit_writer_runtime_owner_status") == "static_boundary_defined",
        "audit writer runtime owner status drifted",
    )
    require(
        target.get("audit_writer_input_envelope_status") == "metadata_only_static_envelope_defined",
        "audit writer input envelope status drifted",
    )
    require(
        target.get("audit_writer_result_reference_status") == "metadata_only_static_reference_defined",
        "audit writer result reference status drifted",
    )
    require(
        target.get("audit_writer_ownership_status") == "separated_static_boundary",
        "audit writer ownership status drifted",
    )
    require(
        target.get("audit_store_runtime_event_schema_materialization_readiness_status")
        == "defined_without_runtime_schema",
        "audit store runtime event schema materialization readiness status drifted",
    )
    require(
        target.get("audit_runtime_event_schema_materialization_owner_status") == "static_boundary_defined",
        "audit runtime event schema materialization owner status drifted",
    )
    require(
        target.get("audit_runtime_event_schema_artifact_implementation_entry_review_status")
        == "ready_for_artifact_task_card",
        "audit runtime event schema artifact implementation entry review status drifted",
    )
    require(
        target.get("audit_runtime_event_schema_artifact_implementation_task_card_status")
        == "defined_without_schema_artifact",
        "audit runtime event schema artifact implementation task card status drifted",
    )
    require(
        target.get("audit_runtime_event_schema_artifact_status") == "implemented_static_schema_artifact",
        "audit runtime event schema artifact status drifted",
    )
    require(
        target.get("audit_runtime_event_schema_artifact_validation_status")
        == "implemented_offline_schema_validation",
        "audit runtime event schema artifact validation status drifted",
    )
    require(
        target.get("audit_runtime_event_schema_version_pin_status") == "static_contract_version_required",
        "audit runtime event schema version pin status drifted",
    )
    require(
        target.get("audit_runtime_event_kind_allowlist_source_status") == "static_contract_reference_only",
        "audit runtime event kind allowlist source status drifted",
    )
    require(
        target.get("audit_runtime_required_optional_fields_source_status") == "static_contract_reference_only",
        "audit runtime required optional fields source status drifted",
    )
    require(
        target.get("audit_runtime_schema_writer_input_compatibility_status")
        == "metadata_only_static_boundary_defined",
        "audit runtime schema writer input compatibility status drifted",
    )
    require(
        target.get("audit_store_delivery_runtime_readiness_status") == "defined_without_delivery_runtime",
        "audit store delivery runtime readiness status drifted",
    )
    require(
        target.get("audit_store_delivery_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "audit store delivery runtime implementation entry review status drifted",
    )
    require(
        target.get("audit_delivery_runtime_owner_status") == "static_boundary_defined",
        "audit delivery runtime owner status drifted",
    )
    require(
        target.get("audit_delivery_input_envelope_status") == "metadata_only_static_envelope_defined",
        "audit delivery input envelope status drifted",
    )
    require(
        target.get("audit_delivery_result_reference_status") == "metadata_only_static_reference_defined",
        "audit delivery result reference status drifted",
    )
    require(
        target.get("audit_delivery_retry_policy_status") == "static_fail_closed_policy_defined",
        "audit delivery retry policy status drifted",
    )
    require(target.get("audit_retry_executor_status") == "not_created", "audit retry executor status drifted")
    require(
        target.get("audit_delivery_result_persistence_status") == "not_created",
        "audit delivery result persistence status drifted",
    )
    require(
        target.get("audit_delivery_execution_status") == "not_executed",
        "audit delivery execution status drifted",
    )
    require(
        target.get("audit_delivery_duplicate_handling_status") == "static_fail_closed_policy_defined",
        "audit delivery duplicate handling status drifted",
    )
    require(
        target.get("audit_store_idempotency_runtime_readiness_status")
        == "defined_without_idempotency_runtime",
        "audit store idempotency runtime readiness status drifted",
    )
    require(
        target.get("audit_store_idempotency_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "audit store idempotency runtime implementation entry review status drifted",
    )
    require(
        target.get("audit_idempotency_runtime_owner_status") == "static_boundary_defined",
        "audit idempotency runtime owner status drifted",
    )
    require(
        target.get("audit_idempotency_input_envelope_status") == "metadata_only_static_envelope_defined",
        "audit idempotency input envelope status drifted",
    )
    require(
        target.get("audit_idempotency_result_reference_status") == "metadata_only_static_reference_defined",
        "audit idempotency result reference status drifted",
    )
    require(
        target.get("audit_idempotency_key_policy_status") == "static_reference_only_policy_defined",
        "audit idempotency key policy status drifted",
    )
    require(
        target.get("audit_idempotency_duplicate_detection_status") == "static_fail_closed_policy_defined",
        "audit idempotency duplicate detection status drifted",
    )
    require(
        target.get("audit_idempotency_replay_decision_status") == "static_fail_closed_policy_defined",
        "audit idempotency replay decision status drifted",
    )
    require(
        target.get("audit_idempotency_key_store_status") == "not_created",
        "audit idempotency key store status drifted",
    )
    require(
        target.get("audit_duplicate_detector_status") == "not_created",
        "audit duplicate detector status drifted",
    )
    require(
        target.get("audit_replay_executor_status") == "not_created",
        "audit replay executor status drifted",
    )
    require(
        target.get("audit_store_delivery_idempotency_runtime_boundary_readiness_status")
        == "defined_without_delivery_runtime",
        "audit store delivery idempotency runtime boundary readiness status drifted",
    )
    require(
        target.get("audit_delivery_idempotency_owner_status") == "static_boundary_defined",
        "audit delivery idempotency owner status drifted",
    )
    require(
        target.get("audit_delivery_owner_status") == "static_reference_defined",
        "audit delivery owner status drifted",
    )
    require(
        target.get("audit_idempotency_key_owner_status") == "static_reference_defined",
        "audit idempotency key owner status drifted",
    )
    require(
        target.get("audit_duplicate_handling_status") == "static_fail_closed_policy_defined",
        "audit duplicate handling status drifted",
    )
    require(
        target.get("audit_retry_failure_semantics_status") == "static_fail_closed_policy_defined",
        "audit retry failure semantics status drifted",
    )
    require(
        target.get("audit_delivery_result_envelope_status") == "metadata_only_static_envelope_defined",
        "audit delivery result envelope status drifted",
    )
    require(
        target.get("audit_delivery_runtime_status") == "not_created",
        "audit delivery runtime must remain not_created",
    )
    require(
        target.get("audit_idempotency_runtime_status") == "not_created",
        "audit idempotency runtime must remain not_created",
    )
    require(
        target.get("audit_store_runtime_implementation_entry_refresh_v3_status")
        == "blocked_before_runtime_task_card",
        "audit store runtime implementation entry refresh v3 status drifted",
    )
    require(
        target.get("audit_store_runtime_implementation_entry_refresh_v4_status")
        == "blocked_before_runtime_task_card",
        "audit store runtime implementation entry refresh v4 status drifted",
    )
    require(
        target.get("audit_store_runtime_implementation_entry_refresh_v5_status")
        == "blocked_after_concrete_backend_selection_review",
        "audit store runtime implementation entry refresh v5 status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "audit store storage adapter runtime implementation entry review status drifted",
    )
    require(
        target.get("audit_storage_adapter_contract_status") == "metadata_only_static_contract_reviewed",
        "audit storage adapter contract status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_backend_product_evidence_readiness_status")
        == "audit_store_storage_adapter_backend_product_evidence_readiness_defined",
        "audit store storage adapter backend product evidence readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_backend_product_evidence_status")
        == "readiness_defined_without_product_selection",
        "audit storage adapter backend product evidence status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_backend_product_selection_review_status")
        == "audit_store_storage_adapter_backend_product_selection_review_defined",
        "audit store storage adapter backend product selection review status drifted",
    )
    require(
        target.get("audit_storage_adapter_backend_product_selection_status")
        == "selected_static_product_class_without_backend_provider",
        "audit storage adapter backend product selection status drifted",
    )
    require(
        target.get("audit_storage_adapter_selected_backend_product_class") == "managed_database_append_only_table",
        "audit storage adapter selected backend product class drifted",
    )
    require(
        target.get("audit_storage_adapter_selected_backend_product_profile")
        == "reserved_managed_database_append_only_table_profile",
        "audit storage adapter selected backend product profile drifted",
    )
    require(
        target.get("audit_storage_adapter_database_product_status") == "engine_selected_without_managed_product",
        "audit storage adapter database product status drifted",
    )
    require(
        target.get("audit_storage_adapter_selected_database_engine")
        == "postgresql_compatible_append_only_relational_database",
        "audit storage adapter selected database engine drifted",
    )
    require(
        target.get("audit_storage_adapter_selected_database_engine_status")
        == "selected_without_vendor_product_driver_or_provider",
        "audit storage adapter selected database engine status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_vendor_status") == "not_selected",
        "audit storage adapter database vendor status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_status")
        == "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined",
        "audit store storage adapter database provider driver DSN TLS role policy readiness status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_append_only_table_schema_boundary_readiness_status")
        == "audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined",
        "audit store storage adapter append-only table schema boundary readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_provider_boundary_status")
        == "metadata_only_provider_boundary_defined",
        "audit storage adapter database provider boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_driver_selection_policy_status")
        == "static_driver_policy_defined_without_driver_selection",
        "audit storage adapter database driver selection policy status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_dsn_secret_ref_policy_status")
        == "secret_ref_only_dsn_policy_defined",
        "audit storage adapter database DSN secret ref policy status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_tls_policy_status") == "tls_mode_policy_defined",
        "audit storage adapter database TLS policy status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_role_policy_status") == "least_privilege_role_policy_defined",
        "audit storage adapter database role policy status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_connection_provider_status") == "not_created",
        "audit storage adapter database connection provider status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_provider_driver_dsn_tls_role_policy_status")
        == "defined_without_runtime",
        "audit storage adapter database provider driver DSN TLS role policy status drifted",
    )
    require(
        target.get("audit_storage_adapter_append_only_table_schema_boundary_status")
        == "defined_without_sql_or_runtime",
        "audit storage adapter append-only table schema boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_logical_table_schema_status")
        == "logical_append_only_table_schema_boundary_defined",
        "audit storage adapter logical table schema status drifted",
    )
    require(
        target.get("audit_storage_adapter_logical_field_group_status")
        == "logical_field_groups_defined_without_physical_columns",
        "audit storage adapter logical field group status drifted",
    )
    require(
        target.get("audit_storage_adapter_record_identity_boundary_status")
        == "logical_record_identity_boundary_defined",
        "audit storage adapter record identity boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_sequence_reference_boundary_status")
        == "logical_sequence_reference_boundary_defined",
        "audit storage adapter sequence reference boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_idempotency_reference_boundary_status")
        == "logical_idempotency_reference_boundary_defined",
        "audit storage adapter idempotency reference boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_retention_redaction_reference_boundary_status")
        == "logical_retention_redaction_reference_boundary_defined",
        "audit storage adapter retention redaction reference boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_schema_marker_handoff_boundary_status")
        == "logical_schema_marker_handoff_boundary_defined",
        "audit storage adapter schema marker handoff boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_table_schema_artifact_materialization_status")
        == "audit_store_storage_adapter_table_schema_artifact_materialized",
        "audit storage adapter table schema artifact materialization status drifted",
    )
    require(
        target.get("audit_storage_adapter_table_schema_artifact_status")
        == "materialized_static_logical_table_schema",
        "audit storage adapter table schema artifact status drifted",
    )
    require(
        target.get("audit_storage_adapter_table_schema_artifact_path_status") == "materialized_static_path",
        "audit storage adapter table schema artifact path status drifted",
    )
    require(
        target.get("audit_storage_adapter_table_schema_artifact_validation_status")
        == "implemented_offline_schema_validation",
        "audit storage adapter table schema artifact validation status drifted",
    )
    require(
        target.get("audit_storage_adapter_table_schema_metadata_contract_compatibility_status")
        == "implemented_static_contract_compatibility",
        "audit storage adapter table schema metadata contract compatibility status drifted",
    )
    require(
        target.get("audit_storage_adapter_table_schema_no_secret_material_scan_status")
        == "implemented_static_scan",
        "audit storage adapter table schema no secret material scan status drifted",
    )
    require(
        target.get("audit_storage_adapter_sql_migration_status") == "not_created",
        "audit storage adapter SQL migration status drifted",
    )
    require(
        target.get("audit_storage_adapter_ddl_status") == "not_created",
        "audit storage adapter DDL status drifted",
    )
    require(
        target.get("audit_storage_adapter_table_name_status") == "not_selected",
        "audit storage adapter table name status drifted",
    )
    require(
        target.get("audit_storage_adapter_column_type_status") == "not_selected",
        "audit storage adapter column type status drifted",
    )
    require(
        target.get("audit_storage_adapter_index_status") == "not_created",
        "audit storage adapter index status drifted",
    )
    require(
        target.get("audit_storage_adapter_constraint_status") == "not_created",
        "audit storage adapter constraint status drifted",
    )
    require(
        target.get("audit_storage_adapter_migration_schema_marker_boundary_status")
        == "logical_schema_marker_handoff_boundary_defined",
        "audit storage adapter migration schema marker boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_schema_marker_runtime_status") == "not_created",
        "audit storage adapter schema marker runtime status drifted",
    )
    require(
        target.get("audit_storage_adapter_migration_runner_status") == "not_created",
        "audit storage adapter migration runner status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_adapter_smoke_strategy_status")
        == "offline_adapter_smoke_strategy_defined_without_runtime",
        "audit storage adapter offline adapter smoke strategy status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_status")
        == "audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined",
        "audit storage adapter negative leakage runtime scan boundary readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_runtime_scan_boundary_status")
        == "defined_without_runtime",
        "audit storage adapter negative leakage runtime scan boundary status drifted",
    )
    for field, expected in {
        "audit_storage_adapter_negative_leakage_runtime_scan_manifest_status": (
            "metadata_only_runtime_scan_manifest_defined"
        ),
        "audit_storage_adapter_negative_leakage_runtime_scan_target_allowlist_status": (
            "metadata_only_scan_target_allowlist_defined"
        ),
        "audit_storage_adapter_negative_leakage_runtime_scan_forbidden_material_coverage_status": (
            "raw_payload_secret_credential_provider_backend_detail_coverage_defined"
        ),
        "audit_storage_adapter_negative_leakage_runtime_scan_diagnostic_allowlist_status": (
            "metadata_only_diagnostic_allowlist_defined"
        ),
        "audit_storage_adapter_negative_leakage_runtime_scan_runner_status": "not_created",
        "audit_storage_adapter_negative_leakage_runtime_scan_output_status": "not_created",
    }.items():
        require(target.get(field) == expected, f"{field} drifted")
    require(
        target.get("audit_storage_adapter_backend_product_candidate_source_status")
        == "metadata_only_candidate_source_defined",
        "audit storage adapter backend product candidate source status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_metadata_contract_artifact_readiness_status")
        == "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined",
        "audit store storage adapter metadata contract artifact readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_metadata_contract_artifact_status")
        == "materialized_static_metadata_contract",
        "audit storage adapter metadata contract artifact status drifted",
    )
    require(
        target.get("audit_storage_adapter_contract_artifact_path_status") == "materialized_static_path",
        "audit storage adapter contract artifact path status drifted",
    )
    require(
        target.get("audit_storage_adapter_input_envelope_status") == "metadata_only_input_envelope_defined",
        "audit storage adapter input envelope status drifted",
    )
    require(
        target.get("audit_storage_adapter_result_envelope_status") == "metadata_only_result_envelope_defined",
        "audit storage adapter result envelope status drifted",
    )
    require(
        target.get("audit_storage_adapter_record_identity_status") == "metadata_only_record_identity_defined",
        "audit storage adapter record identity status drifted",
    )
    require(
        target.get("audit_storage_adapter_failure_taxonomy_status") == "metadata_only_failure_taxonomy_defined",
        "audit storage adapter failure taxonomy status drifted",
    )
    require(
        target.get("audit_storage_adapter_writer_compatibility_status")
        == "metadata_only_writer_compatibility_defined",
        "audit storage adapter writer compatibility status drifted",
    )
    require(
        target.get("audit_storage_adapter_contract_artifact_materialization_status")
        == "audit_store_storage_adapter_metadata_contract_artifact_materialized",
        "audit storage adapter contract artifact materialization status drifted",
    )
    require(
        target.get("audit_storage_adapter_contract_artifact_validation_status")
        == "implemented_offline_contract_validation",
        "audit storage adapter contract artifact validation status drifted",
    )
    require(
        target.get("audit_storage_adapter_writer_compatibility_smoke_status") == "implemented_static_fixture",
        "audit storage adapter writer compatibility smoke status drifted",
    )
    require(
        target.get("audit_storage_adapter_no_secret_material_scan_status") == "implemented_static_scan",
        "audit storage adapter no secret material scan status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_append_only_semantics_evidence_readiness_status")
        == "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        "audit store storage adapter append-only semantics readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_append_only_semantics_status")
        == "append_only_semantics_evidence_defined_without_runtime",
        "audit storage adapter append-only semantics status drifted",
    )
    require(
        target.get("audit_storage_adapter_append_only_operation_status") == "append_only_insert_only",
        "audit storage adapter append-only operation status drifted",
    )
    require(
        target.get("audit_storage_adapter_forbidden_mutation_policy_status")
        == "update_delete_overwrite_truncate_reject_policy_defined",
        "audit storage adapter forbidden mutation policy status drifted",
    )
    require(
        target.get("audit_storage_adapter_sequence_reference_status")
        == "metadata_only_monotonic_sequence_reference_defined",
        "audit storage adapter sequence reference status drifted",
    )
    require(
        target.get("audit_storage_adapter_record_immutability_status")
        == "metadata_only_immutability_policy_defined",
        "audit storage adapter record immutability status drifted",
    )
    require(
        target.get("audit_storage_adapter_duplicate_replay_policy_status")
        == "fail_closed_duplicate_replay_reference_defined",
        "audit storage adapter duplicate replay policy status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_status")
        == "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        "audit store storage adapter retention redaction policy evidence readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_retention_redaction_status")
        == "retention_redaction_policy_evidence_defined_without_runtime",
        "audit storage adapter retention / redaction status drifted",
    )
    require(
        target.get("audit_storage_adapter_retention_window_status")
        == "metadata_only_retention_window_reference_defined",
        "audit storage adapter retention window status drifted",
    )
    require(
        target.get("audit_storage_adapter_redaction_reference_status")
        == "metadata_only_redaction_policy_reference_defined",
        "audit storage adapter redaction reference status drifted",
    )
    require(
        target.get("audit_storage_adapter_retention_redaction_append_only_compatibility_status")
        == "append_only_immutability_compatible_policy_defined",
        "audit storage adapter retention redaction append-only compatibility status drifted",
    )
    require(
        target.get("audit_storage_adapter_forbidden_erasure_policy_status")
        == "delete_overwrite_inline_redaction_forbidden",
        "audit storage adapter forbidden erasure policy status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_offline_validation_evidence_readiness_status")
        == "audit_store_storage_adapter_offline_validation_evidence_readiness_defined",
        "audit store storage adapter offline validation evidence readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_validation_status")
        == "offline_validation_evidence_defined_without_runtime",
        "audit storage adapter offline validation status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_validation_manifest_status")
        == "metadata_only_offline_validation_manifest_reference_defined",
        "audit storage adapter offline validation manifest status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_validation_positive_case_status")
        == "metadata_only_positive_case_reference_defined",
        "audit storage adapter offline validation positive case status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_validation_negative_case_status")
        == "metadata_only_negative_case_reference_defined",
        "audit storage adapter offline validation negative case status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_validation_coverage_status")
        == "metadata_contract_append_only_retention_redaction_coverage_defined",
        "audit storage adapter offline validation coverage status drifted",
    )
    require(
        target.get("audit_storage_adapter_backend_touch_policy_status") == "real_backend_touch_forbidden",
        "audit storage adapter backend touch policy status drifted",
    )
    require(
        target.get("audit_storage_adapter_validation_runner_status") == "not_created",
        "audit storage adapter validation runner status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_status")
        == "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        "audit store storage adapter negative leakage scan evidence readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_scan_evidence_status")
        == "negative_leakage_scan_evidence_defined_without_runtime",
        "audit storage adapter negative leakage scan evidence status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_scan_manifest_status")
        == "metadata_only_negative_leakage_scan_manifest_reference_defined",
        "audit storage adapter negative leakage scan manifest status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_scan_target_status")
        == "metadata_only_scan_target_reference_defined",
        "audit storage adapter negative leakage scan target status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_forbidden_material_coverage_status")
        == "raw_payload_secret_credential_provider_backend_detail_coverage_defined",
        "audit storage adapter negative leakage forbidden material coverage status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_diagnostic_allowlist_status")
        == "metadata_only_diagnostic_allowlist_defined",
        "audit storage adapter negative leakage diagnostic allowlist status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_scan_status") == "not_created",
        "audit storage adapter negative leakage scan status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_scan_runner_status") == "not_created",
        "audit storage adapter negative leakage scan runner status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_scan_output_status") == "not_created",
        "audit storage adapter negative leakage scan output status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_rollback_recovery_evidence_readiness_status")
        == "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined",
        "audit store storage adapter rollback recovery evidence readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_rollback_recovery_status")
        == "rollback_recovery_evidence_defined_without_runtime",
        "audit storage adapter rollback / recovery status drifted",
    )
    require(
        target.get("audit_storage_adapter_rollback_recovery_manifest_status")
        == "metadata_only_rollback_recovery_manifest_reference_defined",
        "audit storage adapter rollback recovery manifest status drifted",
    )
    require(
        target.get("audit_storage_adapter_rollback_append_only_boundary_status")
        == "append_only_compensating_event_boundary_defined",
        "audit storage adapter rollback append-only boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_partial_write_recovery_status")
        == "metadata_only_partial_write_recovery_policy_defined",
        "audit storage adapter partial write recovery status drifted",
    )
    require(
        target.get("audit_storage_adapter_duplicate_replay_recovery_status")
        == "fail_closed_replay_recovery_reference_defined",
        "audit storage adapter duplicate replay recovery status drifted",
    )
    require(
        target.get("audit_storage_adapter_retention_redaction_recovery_alignment_status")
        == "append_only_retention_redaction_compatible_recovery_defined",
        "audit storage adapter retention redaction recovery alignment status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_recovery_alignment_status")
        == "no_raw_material_recovery_diagnostics_defined",
        "audit storage adapter negative leakage recovery alignment status drifted",
    )
    require(
        target.get("audit_storage_adapter_rollback_executor_status") == "not_created",
        "audit storage adapter rollback executor status drifted",
    )
    require(
        target.get("audit_storage_adapter_recovery_executor_status") == "not_created",
        "audit storage adapter recovery executor status drifted",
    )
    require(
        target.get("audit_storage_adapter_compensating_event_writer_status") == "not_created",
        "audit storage adapter compensating event writer status drifted",
    )
    require(
        target.get("audit_storage_adapter_recovery_output_status") == "not_created",
        "audit storage adapter recovery output status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_runtime_implementation_entry_refresh_status")
        == "audit_store_storage_adapter_runtime_implementation_entry_refresh_defined",
        "audit store storage adapter runtime implementation entry refresh status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status")
        == "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined",
        "audit store storage adapter runtime implementation entry refresh after product selection status drifted",
    )
    require(
        target.get(
            "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary_status"
        )
        == "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary_defined",
        "audit store storage adapter runtime implementation entry refresh after negative leakage status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_status")
        == "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_defined",
        "audit store storage adapter runtime implementation entry refresh after database connection lifecycle status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_status")
        == "audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_defined",
        "audit store storage adapter database provider connection runtime boundary readiness status drifted",
    )
    require(
        target.get(
            "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_status"
        )
        == "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_defined",
        "audit store storage adapter runtime implementation entry refresh after database provider connection runtime boundary status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_managed_database_product_selection_readiness_status")
        == "audit_store_storage_adapter_managed_database_product_selection_readiness_defined",
        "audit store storage adapter managed database product selection readiness status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_managed_database_product_selection_review_status")
        == "audit_store_storage_adapter_managed_database_product_selection_review_defined",
        "audit store storage adapter managed database product selection review status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_status")
        == "audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_defined",
        "audit store storage adapter table schema artifact materialization entry review status drifted",
    )
    require(
        target.get("audit_storage_adapter_table_schema_artifact_materialization_task_card_decision")
        == "table_schema_artifact_materialization_task_card_defined_after_entry_review",
        "audit storage adapter table schema materialization task card decision drifted",
    )
    require(
        target.get("audit_storage_adapter_table_schema_artifact_materialization_task_card_status")
        == "created",
        "audit storage adapter table schema materialization task card status drifted",
    )
    require(
        target.get("audit_storage_adapter_table_schema_artifact_materialization_task_card_defined_status")
        == "audit_store_storage_adapter_table_schema_artifact_materialized",
        "audit storage adapter table schema materialization task card defined status drifted",
    )
    require(
        target.get("audit_storage_adapter_runtime_task_card_decision")
        == "storage_adapter_runtime_task_card_still_blocked_after_provider_account_resource_endpoint_readiness",
        "audit storage adapter runtime task card decision drifted",
    )
    require(
        target.get("audit_storage_adapter_next_dependency")
        == "storage_adapter_metadata_contract_artifact_materialization_entry_review",
        "audit storage adapter next dependency drifted",
    )
    require(
        target.get("audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_status")
        == "audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_defined",
        "audit storage adapter metadata contract materialization entry review status drifted",
    )
    require(
        target.get("audit_storage_adapter_contract_materialization_task_card_decision")
        == "metadata_contract_artifact_materialization_task_card_ready_after_entry_review",
        "audit storage adapter materialization task card decision drifted",
    )
    require(
        target.get("audit_storage_adapter_contract_materialization_task_card_status") == "created",
        "audit storage adapter materialization task card status drifted",
    )
    require(
        target.get("audit_storage_adapter_current_next_dependency")
        == "storage_adapter_provider_account_resource_endpoint_review",
        "audit storage adapter current next dependency drifted",
    )
    require(
        target.get("audit_store_storage_adapter_runtime_implementation_entry_refresh_after_managed_database_product_selection_review_status")
        == "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_managed_database_product_selection_review_defined",
        "audit store storage adapter entry refresh after managed product review status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_concrete_managed_database_provider_selection_readiness_status")
        == "audit_store_storage_adapter_concrete_managed_database_provider_selection_readiness_defined",
        "audit store storage adapter concrete managed provider selection readiness status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_concrete_managed_database_provider_selection_review_status")
        == "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_defined",
        "audit store storage adapter concrete managed provider selection review status drifted",
    )
    require(
        target.get(
            "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review_status"
        )
        == "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_concrete_managed_database_provider_selection_review_defined",
        "audit store storage adapter entry refresh after concrete provider review status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_provider_account_resource_endpoint_readiness_status")
        == "audit_store_storage_adapter_provider_account_resource_endpoint_readiness_defined",
        "audit store storage adapter provider account resource endpoint readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_managed_product_selection_status")
        == "selected_reference_product_profile_without_vendor",
        "audit storage adapter managed product selection status drifted",
    )
    require(
        target.get("audit_storage_adapter_managed_product_selection_review_status")
        == "audit_store_storage_adapter_managed_database_product_selection_review_defined",
        "audit storage adapter managed product selection review status drifted",
    )
    require(
        target.get("audit_storage_adapter_selected_managed_product_profile")
        == "managed_postgresql_compatible_audit_store_profile",
        "audit storage adapter selected managed product profile drifted",
    )
    require(
        target.get("audit_storage_adapter_managed_database_product_status")
        == "selected_reference_profile_not_vendor_product",
        "audit storage adapter managed database product status drifted",
    )
    require(
        target.get("audit_storage_adapter_managed_product_input_evidence_status")
        == "metadata_only_product_input_evidence_defined",
        "audit storage adapter managed product input evidence status drifted",
    )
    require(
        target.get("audit_storage_adapter_managed_product_candidate_field_status")
        == "metadata_only_candidate_fields_defined",
        "audit storage adapter managed product candidate field status drifted",
    )
    require(
        target.get("audit_storage_adapter_managed_product_evaluation_dimension_status")
        == "metadata_only_evaluation_dimensions_defined",
        "audit storage adapter managed product evaluation dimension status drifted",
    )
    require(
        target.get("audit_storage_adapter_concrete_provider_input_evidence_status")
        == "metadata_only_provider_input_evidence_defined",
        "audit storage adapter concrete provider input evidence status drifted",
    )
    require(
        target.get("audit_storage_adapter_concrete_provider_candidate_field_status")
        == "metadata_only_provider_candidate_fields_defined",
        "audit storage adapter concrete provider candidate field status drifted",
    )
    require(
        target.get("audit_storage_adapter_concrete_provider_evaluation_dimension_status")
        == "metadata_only_provider_evaluation_dimensions_defined",
        "audit storage adapter concrete provider evaluation dimension status drifted",
    )
    require(
        target.get("audit_storage_adapter_selected_concrete_provider_reference")
        == "managed_postgresql_compatible_provider_reference",
        "audit storage adapter selected concrete provider reference drifted",
    )
    require(
        target.get("audit_storage_adapter_selected_concrete_provider_reference_kind")
        == "reference_only_concrete_provider_profile",
        "audit storage adapter selected concrete provider reference kind drifted",
    )
    require(
        target.get("audit_storage_adapter_concrete_provider_selection_status")
        == "reference_profile_selected_without_provider_resource",
        "audit storage adapter concrete provider selection status drifted",
    )
    require(
        target.get("audit_storage_adapter_concrete_provider_selection_review_status")
        == "audit_store_storage_adapter_concrete_managed_database_provider_selection_review_defined",
        "audit storage adapter concrete provider selection review status drifted",
    )
    require(
        target.get("audit_storage_adapter_concrete_cloud_product_status") == "not_selected",
        "audit storage adapter concrete cloud product status drifted",
    )
    require(
        target.get("audit_storage_adapter_provider_account_resource_status")
        == "metadata_only_readiness_defined_without_real_resource",
        "audit storage adapter provider account resource status drifted",
    )
    require(
        target.get("audit_storage_adapter_provider_resource_status") == "not_selected",
        "audit storage adapter provider resource status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_endpoint_status")
        == "metadata_only_endpoint_requirements_defined_without_endpoint",
        "audit storage adapter database endpoint status drifted",
    )
    require(
        target.get("audit_storage_adapter_region_detail_status")
        == "metadata_only_region_requirements_defined_without_region_detail",
        "audit storage adapter region detail status drifted",
    )
    require(
        target.get("audit_storage_adapter_provider_account_confirmation_status")
        == "operator_confirmation_required_before_runtime",
        "audit storage adapter provider account confirmation status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_concrete_database_selection_readiness_status")
        == "audit_store_storage_adapter_concrete_database_selection_readiness_defined",
        "audit store storage adapter concrete database selection readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_concrete_database_selection_readiness_status")
        == "audit_store_storage_adapter_concrete_database_selection_readiness_defined",
        "audit storage adapter concrete database selection readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_concrete_database_selection_status")
        == "selected_database_engine_without_vendor_or_provider",
        "audit storage adapter concrete database selection status drifted",
    )
    require(
        target.get("audit_storage_adapter_candidate_input_evidence_status") == "metadata_only_input_evidence_defined",
        "audit storage adapter candidate input evidence status drifted",
    )
    require(
        target.get("audit_storage_adapter_candidate_evaluation_matrix_status") == "metadata_only_evaluation_matrix_defined",
        "audit storage adapter candidate evaluation matrix status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_concrete_database_selection_review_status")
        == "audit_store_storage_adapter_concrete_database_selection_review_defined",
        "audit store storage adapter concrete database selection review status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_selection_review_status")
        == "audit_store_storage_adapter_concrete_database_selection_review_defined",
        "audit storage adapter database selection review status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_database_provider_selection_readiness_status")
        == "audit_store_storage_adapter_database_provider_selection_readiness_defined",
        "audit store storage adapter database provider selection readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_provider_selection_readiness_status")
        == "audit_store_storage_adapter_database_provider_selection_readiness_defined",
        "audit storage adapter database provider selection readiness status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_database_provider_selection_review_status")
        == "audit_store_storage_adapter_database_provider_selection_review_defined",
        "audit store storage adapter database provider selection review status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_provider_selection_review_status")
        == "audit_store_storage_adapter_database_provider_selection_review_defined",
        "audit storage adapter database provider selection review status drifted",
    )
    require(
        target.get("audit_storage_adapter_selected_provider_candidate_class")
        == "managed_postgresql_compatible_service",
        "audit storage adapter selected provider candidate class drifted",
    )
    require(
        target.get("audit_storage_adapter_database_provider_selection_status")
        == "selected_provider_candidate_class_without_vendor_or_product",
        "audit storage adapter database provider selection status drifted",
    )
    require(
        target.get("audit_storage_adapter_provider_candidate_source_status")
        == "metadata_only_provider_candidate_source_defined",
        "audit storage adapter provider candidate source status drifted",
    )
    require(
        target.get("audit_storage_adapter_provider_input_evidence_status")
        == "metadata_only_provider_input_evidence_defined",
        "audit storage adapter provider input evidence status drifted",
    )
    require(
        target.get("audit_storage_adapter_provider_evaluation_dimension_status")
        == "metadata_only_provider_evaluation_dimensions_defined",
        "audit storage adapter provider evaluation dimension status drifted",
    )
    require(
        target.get("audit_storage_adapter_provider_selection_review_status")
        == "audit_store_storage_adapter_database_provider_selection_review_defined",
        "audit storage adapter provider selection review status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_database_driver_selection_readiness_status")
        == "audit_store_storage_adapter_database_driver_selection_readiness_defined",
        "audit store storage adapter database driver selection readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_driver_selection_readiness_status")
        == "audit_store_storage_adapter_database_driver_selection_readiness_defined",
        "audit storage adapter database driver selection readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_driver_selection_status")
        == "selected_driver_candidate_without_runtime_import",
        "audit storage adapter database driver selection status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_database_driver_selection_review_status")
        == "audit_store_storage_adapter_database_driver_selection_review_defined",
        "audit store storage adapter database driver selection review status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_driver_selection_review_status")
        == "audit_store_storage_adapter_database_driver_selection_review_defined",
        "audit storage adapter database driver selection review status drifted",
    )
    require(
        target.get("audit_storage_adapter_selected_database_driver_candidate") == "github.com/jackc/pgx/v5",
        "audit storage adapter selected database driver candidate drifted",
    )
    require(
        target.get("audit_storage_adapter_database_driver_status") == "selected_reference_only",
        "audit storage adapter database driver status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_driver_package_status")
        == "selected_candidate_reference_only",
        "audit storage adapter database driver package status drifted",
    )
    require(
        target.get("audit_storage_adapter_database_driver_import_status") == "not_created",
        "audit storage adapter database driver import status drifted",
    )
    require(
        target.get("audit_storage_adapter_driver_dependency_version_status") == "not_pinned",
        "audit storage adapter driver dependency version status drifted",
    )
    require(
        target.get("audit_storage_adapter_driver_candidate_source_status")
        == "metadata_only_driver_candidate_source_defined",
        "audit storage adapter driver candidate source status drifted",
    )
    require(
        target.get("audit_storage_adapter_driver_import_boundary_status")
        == "metadata_only_driver_import_boundary_defined",
        "audit storage adapter driver import boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_driver_capability_evidence_status")
        == "metadata_only_driver_capability_evidence_defined",
        "audit storage adapter driver capability evidence status drifted",
    )
    require(
        target.get("audit_storage_adapter_dsn_secret_ref_compatibility_status")
        == "metadata_only_dsn_secret_ref_compatibility_defined",
        "audit storage adapter DSN secret ref compatibility status drifted",
    )
    require(
        target.get("audit_storage_adapter_tls_mode_compatibility_status")
        == "metadata_only_tls_mode_compatibility_defined",
        "audit storage adapter TLS mode compatibility status drifted",
    )
    require(
        target.get("audit_storage_adapter_role_policy_compatibility_status")
        == "metadata_only_role_policy_compatibility_defined",
        "audit storage adapter role policy compatibility status drifted",
    )
    require(
        target.get("audit_storage_adapter_connection_lifecycle_boundary_status")
        == "metadata_only_connection_lifecycle_boundary_defined",
        "audit storage adapter connection lifecycle boundary status drifted",
    )
    for field, expected in {
        "audit_store_storage_adapter_database_connection_lifecycle_readiness_status": (
            "audit_store_storage_adapter_database_connection_lifecycle_readiness_defined"
        ),
        "audit_storage_adapter_database_connection_lifecycle_readiness_status": (
            "audit_store_storage_adapter_database_connection_lifecycle_readiness_defined"
        ),
        "audit_storage_adapter_database_connection_lifecycle_readiness_decision": (
            "database_connection_lifecycle_readiness_defined_without_connection_runtime"
        ),
        "audit_storage_adapter_secret_ref_only_dsn_handoff_status": "secret_ref_only_dsn_handoff_defined",
        "audit_storage_adapter_tls_role_environment_binding_status": (
            "static_tls_role_environment_binding_defined"
        ),
        "audit_storage_adapter_pool_policy_status": "static_pool_policy_defined_without_pool_runtime",
        "audit_storage_adapter_timeout_budget_status": "static_timeout_budget_defined_without_runtime_timer",
        "audit_storage_adapter_retry_transaction_recovery_status": "static_recovery_policy_defined_without_runtime",
        "audit_storage_adapter_duplicate_replay_fail_closed_status": (
            "duplicate_replay_fail_closed_policy_defined"
        ),
        "audit_storage_adapter_sanitized_diagnostics_status": "sanitized_diagnostics_allowlist_defined",
        "audit_storage_adapter_schema_marker_migration_handoff_status": "schema_marker_migration_handoff_defined",
        "audit_storage_adapter_offline_verification_status": "metadata_only_offline_verification_defined",
        "audit_storage_adapter_database_provider_connection_runtime_boundary_status": (
            "metadata_only_boundary_defined_without_runtime"
        ),
        "audit_storage_adapter_connection_lifecycle_runtime_status": "not_created",
        "audit_storage_adapter_connection_factory_status": "not_created",
        "audit_storage_adapter_pool_runtime_status": "not_created",
        "audit_storage_adapter_health_check_runtime_status": "not_created",
    }.items():
        require(target.get(field) == expected, f"{field} drifted")
    require(
        target.get("audit_storage_adapter_migration_schema_marker_boundary_status")
        == "logical_schema_marker_handoff_boundary_defined",
        "audit storage adapter migration schema marker boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_driver_offline_smoke_boundary_status")
        == "metadata_only_offline_adapter_smoke_boundary_defined",
        "audit storage adapter driver offline smoke boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_driver_negative_leakage_scan_boundary_status")
        == "metadata_only_negative_leakage_runtime_scan_boundary_defined",
        "audit storage adapter driver negative leakage scan boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_rollout_rollback_boundary_status")
        == "metadata_only_rollout_rollback_boundary_defined",
        "audit storage adapter rollout rollback boundary status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_status")
        == "audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined",
        "audit storage adapter offline adapter smoke strategy readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_adapter_smoke_strategy_status")
        == "offline_adapter_smoke_strategy_defined_without_runtime",
        "audit storage adapter offline adapter smoke strategy status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_adapter_smoke_manifest_status")
        == "metadata_only_smoke_manifest_defined",
        "audit storage adapter offline adapter smoke manifest status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_adapter_smoke_positive_case_status")
        == "metadata_only_positive_case_defined",
        "audit storage adapter offline adapter smoke positive case status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_adapter_smoke_negative_case_status")
        == "metadata_only_negative_case_defined",
        "audit storage adapter offline adapter smoke negative case status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_adapter_smoke_backend_touch_policy_status")
        == "real_backend_touch_forbidden",
        "audit storage adapter offline adapter smoke backend touch policy status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_adapter_smoke_runner_status") == "not_created",
        "audit storage adapter offline adapter smoke runner status drifted",
    )
    require(
        target.get("audit_storage_adapter_offline_adapter_smoke_output_status") == "not_created",
        "audit storage adapter offline adapter smoke output status drifted",
    )
    require(
        target.get("audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_status")
        == "audit_store_storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined",
        "audit storage adapter negative leakage runtime scan boundary readiness status drifted",
    )
    require(
        target.get("audit_storage_adapter_negative_leakage_runtime_scan_boundary_status")
        == "defined_without_runtime",
        "audit storage adapter negative leakage runtime scan boundary status drifted",
    )
    for field, expected in {
        "audit_storage_adapter_negative_leakage_runtime_scan_manifest_status": (
            "metadata_only_runtime_scan_manifest_defined"
        ),
        "audit_storage_adapter_negative_leakage_runtime_scan_target_allowlist_status": (
            "metadata_only_scan_target_allowlist_defined"
        ),
        "audit_storage_adapter_negative_leakage_runtime_scan_forbidden_material_coverage_status": (
            "raw_payload_secret_credential_provider_backend_detail_coverage_defined"
        ),
        "audit_storage_adapter_negative_leakage_runtime_scan_diagnostic_allowlist_status": (
            "metadata_only_diagnostic_allowlist_defined"
        ),
        "audit_storage_adapter_negative_leakage_runtime_scan_runner_status": "not_created",
        "audit_storage_adapter_negative_leakage_runtime_scan_output_status": "not_created",
    }.items():
        require(target.get(field) == expected, f"{field} drifted")
    require(
        target.get("audit_storage_adapter_runtime_task_card_status") == "not_created",
        "audit storage adapter runtime task card must remain not_created",
    )
    require(
        target.get("audit_storage_adapter_runtime_status") == "not_created",
        "audit storage adapter runtime must remain not_created",
    )
    require(
        target.get("audit_store_runtime_blocker_matrix_status")
        == "audit_store_runtime_blocker_matrix_defined",
        "audit store runtime blocker matrix status drifted",
    )
    require(
        target.get("audit_store_runtime_task_card_status") == "not_created",
        "audit store runtime task card must remain not_created",
    )
    require(target.get("audit_store_runtime_status") == "not_created", "audit store runtime must remain not_created")
    require(target.get("audit_store_status") == "not_created", "audit store must remain not_created")
    require(target.get("audit_writer_status") == "not_created", "audit writer must remain not_created")
    require(
        target.get("audit_event_delivery_status") == "not_executed",
        "audit event delivery must remain not_executed",
    )
    require(
        target.get("resolver_backend_health_boundary_readiness_status")
        == "defined_without_backend_health_runtime",
        "resolver backend health boundary readiness status drifted",
    )
    require(
        target.get("resolver_backend_health_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "resolver backend health runtime implementation entry review status drifted",
    )
    require(
        target.get("resolver_backend_health_runtime_implementation_entry_refresh_status")
        == "blocked_before_runtime_task_card",
        "resolver backend health runtime implementation entry refresh status drifted",
    )
    require(target.get("backend_health_runtime_status") == "not_created", "backend health runtime must remain not_created")
    require(
        target.get("backend_health_check_status") == "not_executed",
        "backend health check must remain not_executed",
    )
    require(
        target.get("default_runtime_state") == "disabled_until_explicit_secret_backend_task",
        "default runtime state must remain disabled",
    )
    require(
        target.get("fast_baseline_policy") == "offline_no_real_credentials_no_cloud_calls",
        "fast baseline must stay offline without credentials or cloud calls",
    )


def assert_preconditions(fixture: dict[str, Any]) -> None:
    preconditions = {
        str(item.get("id")): item
        for item in fixture.get("required_preconditions") or []
        if isinstance(item, dict)
    }
    missing = sorted(REQUIRED_PRECONDITIONS - set(preconditions))
    require(not missing, f"missing required preconditions: {missing}")
    for precondition_id, item in preconditions.items():
        status = str(item.get("status") or "")
        require(
            status == "satisfied"
            or status == "satisfied_for_test_only_fake_resolver"
            or status.startswith("required_before_"),
            f"{precondition_id} has unexpected status",
        )
        must_define = item.get("must_define") or []
        require(len(must_define) >= 3, f"{precondition_id} must define concrete requirements")
        if precondition_id == "secret-ref-schema":
            require(status == "satisfied", "secret-ref-schema precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "contracts/production-secret-reference.schema.json",
                "scripts/checks/fixtures/production-secret-reference-basic.json",
                "scripts/check-production-secret-reference-contract.py",
            }:
                require(path in evidence, f"secret-ref-schema missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"secret-ref-schema evidence missing on disk: {path}")
        if precondition_id == "config-injection-point":
            require(status == "satisfied", "config-injection-point precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md",
                "docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
            }:
                require(path in evidence, f"config-injection-point missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"config-injection-point evidence missing on disk: {path}")
        if precondition_id == "provider-profile-binding":
            require(status == "satisfied", "provider-profile-binding precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md",
                "docs/task-cards/production-secret-backend-provider-profile-secret-binding-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py",
            }:
                require(path in evidence, f"provider-profile-binding missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"provider-profile-binding evidence missing on disk: {path}")
        if precondition_id in {"sanitized-audit-fields", "failure-taxonomy"}:
            require(status == "satisfied", f"{precondition_id} precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json",
                "scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
            }:
                require(path in evidence, f"{precondition_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{precondition_id} evidence missing on disk: {path}")
        if precondition_id == "test-fixture-strategy":
            require(
                status == "satisfied_for_test_only_fake_resolver",
                "test-fixture-strategy must be satisfied for test-only fake resolver runtime",
            )
            for required in {
                "fake resolver",
                "placeholder secret ref",
                "offline fast baseline",
                "no cloud SDK call",
                "no secret leakage smoke",
                "artifact guard",
            }:
                require(required in must_define, f"test-fixture-strategy must define {required}")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-implementation-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py",
                "services/platform/internal/secretbackend/fake_resolver.go",
                "services/platform/internal/secretbackend/fake_resolver_test.go",
            }:
                require(path in evidence, f"test-fixture-strategy missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"test-fixture-strategy evidence missing on disk: {path}")
        if precondition_id == "operator-runbook":
            require(status == "satisfied", "operator-runbook precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md",
                "docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
            }:
                require(path in evidence, f"operator-runbook missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"operator-runbook evidence missing on disk: {path}")
        if precondition_id == "rotation-and-audit-policy":
            require(status == "satisfied", "rotation-and-audit-policy precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-rotation-audit-policy-readiness-v1.md",
                "docs/task-cards/production-secret-backend-rotation-audit-policy-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py",
            }:
                require(path in evidence, f"rotation-and-audit-policy missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"rotation-and-audit-policy evidence missing on disk: {path}")

    sanitized_fields = set(preconditions["sanitized-audit-fields"].get("must_define") or [])
    for field in {
        "credential_state",
        "secret_backend_configured",
        "secret_ref_present",
        "missing_secret_refs",
        "field_sources",
    }:
        require(field in sanitized_fields, f"sanitized audit fields missing {field}")


def assert_planned_slices_and_blocks(fixture: dict[str, Any]) -> None:
    planned = {str(item.get("id")): item for item in fixture.get("planned_slices") or [] if isinstance(item, dict)}
    missing_planned = sorted(set(REQUIRED_PLANNED_SLICES) - set(planned))
    require(not missing_planned, f"missing planned slices: {missing_planned}")
    for slice_id, expected_status in REQUIRED_PLANNED_SLICES.items():
        require(planned[slice_id].get("status") == expected_status, f"{slice_id} has unexpected status")
        if slice_id == "secret-ref-schema-and-fixtures":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "contracts/production-secret-reference.schema.json",
                "scripts/checks/fixtures/production-secret-reference-basic.json",
                "scripts/check-production-secret-reference-contract.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "config-secret-ref-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md",
                "docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "provider-profile-secret-binding":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md",
                "docs/task-cards/production-secret-backend-provider-profile-secret-binding-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "secret-resolver-interface-disabled":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md",
                "docs/task-cards/production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "operator-runbook-and-negative-gates":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md",
                "docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "rotation-and-audit-policy":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-rotation-audit-policy-readiness-v1.md",
                "docs/task-cards/production-secret-backend-rotation-audit-policy-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "test-fixture-strategy":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-contract-no-secret-leakage-smoke-strategy":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-implementation-task-card-entry-readiness-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-implementation":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-implementation-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-runtime-implementation":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-runtime-preconditions":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-runtime-implementation-entry-refresh":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "resolver-backend-profile-selection-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md",
                "docs/task-cards/production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-no-secret-leakage-smoke-runtime-strategy":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "credential-handle-runtime-boundary-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.md",
                "docs/task-cards/production-secret-backend-credential-handle-runtime-boundary-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "credential-handle-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.md",
                (
                    "docs/task-cards/"
                    "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.json"
                ),
                "scripts/check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "credential-handle-runtime-implementation-entry-refresh":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.md",
                (
                    "docs/task-cards/"
                    "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.json"
                ),
                "scripts/check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "operator-approval-runtime-evidence-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md",
                "docs/task-cards/production-secret-backend-operator-approval-runtime-evidence-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "operator-approval-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "operator-approval-runtime-implementation-entry-refresh":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.md",
                (
                    "docs/task-cards/"
                    "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.json"
                ),
                "scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "cloud-secret-service-selection-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-cloud-secret-service-selection-readiness-v1.md",
                "docs/task-cards/production-secret-backend-cloud-secret-service-selection-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "resolver-backend-health-runtime-implementation-entry-refresh":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.md",
                (
                    "docs/task-cards/"
                    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.json"
                ),
                "scripts/check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-handoff-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-handoff-readiness-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-handoff-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-handoff-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-contract-event-schema-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-contract-event-schema-readiness-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-contract-event-schema-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-implementation-entry-refresh-v2":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.md",
                "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.json",
                "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-ownership-boundary-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-ownership-boundary-readiness-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-ownership-boundary-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-ownership-boundary-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-ownership-boundary-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-delivery-idempotency-runtime-boundary-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.md",
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.json"
                ),
                (
                    "scripts/check-production-ops-secret-backend-"
                    "audit-store-delivery-idempotency-runtime-boundary-readiness-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-implementation-entry-refresh-v3":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.md",
                "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json",
                "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-durable-backend-boundary-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.md",
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-durable-backend-boundary-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.json"
                ),
                "scripts/check-production-ops-secret-backend-audit-store-durable-backend-boundary-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-writer-runtime-boundary-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.md",
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json"
                ),
                "scripts/check-production-ops-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-event-schema-materialization-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.md",
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.json"
                ),
                "scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-delivery-runtime-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-delivery-runtime-readiness-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-delivery-runtime-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-runtime-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-delivery-runtime-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-idempotency-runtime-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-idempotency-runtime-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-idempotency-runtime-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-implementation-entry-refresh-v4":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.md",
                "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json",
                "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-event-schema-artifact-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-event-schema-artifact-implementation":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-event-schema-artifact":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-runtime-event-schema-artifact-v1-plan.md",
                "docs/contracts/production-secret-audit-event.md",
                "contracts/production-secret-audit-event.schema.json",
                "scripts/checks/fixtures/production-secret-audit-event-positive-v1.json",
                "scripts/checks/fixtures/production-secret-audit-event-missing-required-negative-v1.json",
                "scripts/checks/fixtures/production-secret-audit-event-forbidden-field-negative-v1.json",
                "scripts/checks/fixtures/production-secret-audit-event-additional-properties-negative-v1.json",
                "scripts/checks/fixtures/production-secret-audit-event-event-kind-invalid-negative-v1.json",
                "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-blocker-matrix":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-runtime-blocker-matrix-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-storage-adapter-backend-product-evidence-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-storage-adapter-metadata-contract-artifact-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-storage-adapter-append-only-semantics-evidence-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-storage-adapter-retention-redaction-policy-evidence-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-storage-adapter-offline-validation-evidence-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-durable-backend-selection-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-durable-backend-selection-readiness-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-durable-backend-selection-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-durable-backend-selection-readiness-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-durable-backend-selection-readiness-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-writer-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-idempotency-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-delivery-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-storage-adapter-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-storage-adapter-"
                    "database-provider-driver-dsn-tls-role-policy-readiness-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-storage-adapter-"
                    "database-provider-driver-dsn-tls-role-policy-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-storage-adapter-"
                    "database-provider-driver-dsn-tls-role-policy-readiness-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-storage-adapter-"
                    "database-provider-driver-dsn-tls-role-policy-readiness-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-storage-adapter-append-only-table-schema-boundary-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                (
                    "docs/platform/"
                    "production-secret-backend-audit-store-storage-adapter-"
                    "append-only-table-schema-boundary-readiness-v1.md"
                ),
                (
                    "docs/task-cards/"
                    "production-secret-backend-audit-store-storage-adapter-"
                    "append-only-table-schema-boundary-readiness-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-audit-store-storage-adapter-"
                    "append-only-table-schema-boundary-readiness-v1.json"
                ),
                (
                    "scripts/"
                    "check-production-ops-secret-backend-audit-store-storage-adapter-"
                    "append-only-table-schema-boundary-readiness-v1.py"
                ),
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "resolver-backend-health-boundary-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-resolver-backend-health-boundary-readiness-v1.md",
                "docs/task-cards/production-secret-backend-resolver-backend-health-boundary-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-boundary-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "resolver-backend-health-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-storage-adapter-provider-account-resource-endpoint-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")

    blocked = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or [] if isinstance(item, dict)}
    missing_blocked = sorted(set(REQUIRED_BLOCKED) - set(blocked))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    for blocked_id, expected_status in REQUIRED_BLOCKED.items():
        require(blocked[blocked_id].get("status") == expected_status, f"{blocked_id} has unexpected status")


def assert_validation_and_docs(fixture: dict[str, Any]) -> None:
    validation = set(fixture.get("validation_strategy") or [])
    for required in {
        "fast baseline stays offline",
        "no real credential required",
        "no cloud API call",
        "no raw secret or provider base URL in logs, diagnostics, fixtures or docs",
        "real resolver runtime implementation entry refresh blocked before task card",
        "credential handle runtime implementation entry review blocked before task card",
        "credential handle runtime implementation entry refresh blocked before task card",
        "real resolver no leakage smoke runtime implementation entry review blocked before task card",
        "real resolver no leakage smoke runtime implementation entry refresh blocked before task card",
        "operator approval runtime evidence readiness defined without runtime execution",
        "operator approval runtime implementation entry review blocked before task card",
        "operator approval runtime implementation entry refresh blocked before task card",
        "cloud secret service selection readiness defined without cloud backend selection",
        "audit store handoff readiness defined without store runtime",
        "audit store runtime implementation entry review blocked before task card",
        "audit store contract event schema readiness defined without store runtime",
        "audit store runtime implementation entry refresh v2 blocked before task card",
        "audit store ownership boundary readiness defined without store runtime",
        "audit store delivery idempotency runtime boundary readiness defined without delivery runtime",
        "audit store runtime implementation entry refresh v3 blocked before task card",
        "audit store durable backend boundary readiness defined without backend selection",
        "audit store writer runtime boundary readiness defined without writer runtime",
        "audit store runtime event schema materialization readiness defined without runtime schema",
        "audit store delivery runtime readiness defined without delivery runtime",
        "audit store idempotency runtime readiness defined without idempotency runtime",
        "audit store storage adapter metadata contract artifact readiness defined without materialized artifact",
        "audit store storage adapter retention redaction policy evidence readiness defined without runtime",
        "audit store storage adapter runtime implementation entry refresh after product selection blocked before database provider readiness",
        "audit store storage adapter database provider driver DSN TLS role policy readiness defined without runtime",
        "audit store storage adapter concrete database selection readiness defined without database selection",
        "audit store storage adapter provider account resource endpoint readiness defined without real provider resource",
        "resolver backend health boundary readiness defined without backend health runtime",
        "resolver backend health runtime implementation entry review blocked before task card",
        "resolver backend health runtime implementation entry refresh blocked before task card",
    }:
        require(required in validation, f"validation strategy missing: {required}")

    for evidence in fixture.get("evidence") or []:
        require((REPO_ROOT / str(evidence)).exists(), f"missing evidence path: {evidence}")

    expected_consumers = {
        "scripts/check-production-ops-secret-backend-implementation-readiness.py",
        "scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py",
        "scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.py",
        "scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.py",
        "scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.py",
        "scripts/check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.py",
        "scripts/check-production-ops-secret-backend-audit-store-ownership-boundary-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py",
        "scripts/check-production-ops-secret-backend-audit-store-durable-backend-boundary-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-delivery-runtime-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-idempotency-runtime-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.py",
        (
            "scripts/"
            "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.py"
        ),
        (
            "scripts/"
            "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.py"
        ),
        "scripts/check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-secret-reference-contract.py",
        "scripts/check-repo.py",
        "scripts/README.md",
        "docs/radishmind-current-focus.md",
        "docs/radishmind-roadmap.md",
        "docs/radishmind-capability-matrix.md",
        "docs/task-cards/production-ops-hardening-v1-plan.md",
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md",
        "docs/platform/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.json",
        "docs/platform/production-secret-backend-fake-resolver-implementation-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json",
        "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json",
        "services/platform/internal/secretbackend/fake_resolver.go",
        "services/platform/internal/secretbackend/fake_resolver_test.go",
        "docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json",
        "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.json",
        "docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
        "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.json",
        "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.json",
        "docs/platform/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.md",
        "docs/task-cards/production-secret-backend-credential-handle-runtime-boundary-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.json",
        "docs/platform/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.md",
        "docs/task-cards/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.json",
        "docs/platform/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md",
        "docs/task-cards/production-secret-backend-operator-approval-runtime-evidence-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json",
        "docs/platform/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.md",
        "docs/task-cards/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.json",
        "docs/platform/production-secret-backend-cloud-secret-service-selection-readiness-v1.md",
        "docs/task-cards/production-secret-backend-cloud-secret-service-selection-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json",
        "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.json",
        "docs/platform/production-secret-backend-audit-store-handoff-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-handoff-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-handoff-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-audit-store-contract-event-schema-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-contract-event-schema-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.md",
        "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.json",
        "docs/platform/production-secret-backend-audit-store-ownership-boundary-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-ownership-boundary-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-ownership-boundary-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.md",
        "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json",
        "docs/platform/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-delivery-runtime-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-delivery-runtime-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-runtime-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-idempotency-runtime-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.md",
        "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json",
        (
            "docs/platform/"
            "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.md"
        ),
        (
            "docs/task-cards/"
            "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1-plan.md"
        ),
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.json"
        ),
        (
            "docs/platform/"
            "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.md"
        ),
        (
            "docs/task-cards/"
            "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1-plan.md"
        ),
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.json"
        ),
        "docs/platform/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-runtime-event-schema-artifact-v1-plan.md",
        "docs/contracts/production-secret-audit-event.md",
        "contracts/production-secret-audit-event.schema.json",
        "scripts/checks/fixtures/production-secret-audit-event-positive-v1.json",
        "scripts/checks/fixtures/production-secret-audit-event-missing-required-negative-v1.json",
        "scripts/checks/fixtures/production-secret-audit-event-forbidden-field-negative-v1.json",
        "scripts/checks/fixtures/production-secret-audit-event-additional-properties-negative-v1.json",
        "scripts/checks/fixtures/production-secret-audit-event-event-kind-invalid-negative-v1.json",
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.json",
        "scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py",
        "docs/platform/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1-plan.md",
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.json"
        ),
        (
            "scripts/"
            "check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py"
        ),
        "docs/platform/production-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1-plan.md",
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1.json"
        ),
        (
            "scripts/"
            "check-production-ops-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1.py"
        ),
        "docs/platform/production-secret-backend-resolver-backend-health-boundary-readiness-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-health-boundary-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-boundary-readiness-v1.json",
        "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
        "services/platform/README.md",
        "docs/devlogs/2026-W22.md",
        "docs/devlogs/2026-W25.md",
    }
    missing_consumers = sorted(expected_consumers - set(fixture.get("required_consumers") or []))
    require(not missing_consumers, f"missing consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-secret-backend-implementation-readiness.py", [])' in check_repo,
        "check-repo.py must run secret backend implementation readiness check",
    )
    require(
        'run_python_script("check-production-secret-reference-contract.py", [])' in check_repo,
        "check-repo.py must run production secret reference contract check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run secret resolver interface disabled readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run operator runbook negative gates readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run rotation audit policy readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run test fixture strategy fake resolver entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver contract no leakage strategy check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver implementation task card entry readiness review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-implementation-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver implementation check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver runtime implementation check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver runtime preconditions check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver runtime implementation entry refresh check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run resolver backend profile selection readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver no leakage smoke runtime strategy check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver no leakage smoke runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver no leakage smoke runtime implementation entry refresh check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run credential handle runtime boundary readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run credential handle runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.py", [])'
        in check_repo,
        "check-repo.py must run credential handle runtime implementation entry refresh check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run operator approval runtime evidence readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run operator approval runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.py", [])'
        in check_repo,
        "check-repo.py must run operator approval runtime implementation entry refresh check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run cloud secret service selection readiness check",
    )
    require(
        (
            'run_python_script("'
            'check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.py", [])'
        )
        in check_repo,
        "check-repo.py must run resolver backend health runtime implementation entry refresh check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run audit store handoff readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run audit store runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run audit store contract event schema readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.py", [])'
        in check_repo,
        "check-repo.py must run audit store runtime implementation entry refresh v2 check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-audit-store-ownership-boundary-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run audit store ownership boundary readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run audit store delivery idempotency runtime boundary readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py", [])'
        in check_repo,
        "check-repo.py must run audit store runtime implementation entry refresh v3 check",
    )
    require(
        (
            'run_python_script("'
            'check-production-ops-secret-backend-audit-store-durable-backend-boundary-readiness-v1.py", [])'
        )
        in check_repo,
        "check-repo.py must run audit store durable backend boundary readiness check",
    )
    require(
        (
            'run_python_script("'
            'check-production-ops-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.py", [])'
        )
        in check_repo,
        "check-repo.py must run audit store writer runtime boundary readiness check",
    )
    require(
        (
            'run_python_script("'
            'check-production-ops-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.py", [])'
        )
        in check_repo,
        "check-repo.py must run audit store runtime event schema materialization readiness check",
    )
    require(
        (
            'run_python_script("'
            'check-production-ops-secret-backend-audit-store-delivery-runtime-readiness-v1.py", [])'
        )
        in check_repo,
        "check-repo.py must run audit store delivery runtime readiness check",
    )
    require(
        (
            'run_python_script("'
            'check-production-ops-secret-backend-audit-store-idempotency-runtime-readiness-v1.py", [])'
        )
        in check_repo,
        "check-repo.py must run audit store idempotency runtime readiness check",
    )
    require(
        (
            'run_python_script("'
            'check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.py", [])'
        )
        in check_repo,
        "check-repo.py must run audit store runtime implementation entry refresh v4 check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store runtime event schema artifact entry review check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store runtime event schema artifact implementation check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store runtime event schema artifact check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store runtime blocker matrix check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter backend product evidence readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter metadata contract artifact readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter append-only semantics evidence readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter retention redaction policy evidence readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter offline validation evidence readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter negative leakage scan evidence readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter rollback recovery evidence readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter runtime implementation entry refresh check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter metadata contract artifact materialization entry review check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter metadata contract artifact materialization check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter backend product selection review check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter runtime implementation entry refresh after product selection check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter database provider driver DSN TLS role policy readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter concrete database selection readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-storage-adapter-provider-account-resource-endpoint-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store storage adapter provider account resource endpoint readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-durable-backend-selection-readiness-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store durable backend selection readiness check",
    )
    require(
        (
            'run_python_script("'
            "check-production-ops-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.py"
            '", [])'
        )
        in check_repo,
        "check-repo.py must run audit store writer runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run backend health boundary readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run backend health runtime implementation entry review check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def assert_no_secret_literals() -> None:
    text = FIXTURE_PATH.read_text(encoding="utf-8") + "\n" + read(
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md"
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"secret backend readiness docs contain forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", text) is None,
        "secret backend readiness docs contain forbidden secret-looking sk token",
    )


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_implementation_readiness",
        "unexpected fixture kind",
    )
    assert_slice(fixture)
    assert_implementation_target(fixture)
    assert_preconditions(fixture)
    assert_planned_slices_and_blocks(fixture)
    assert_validation_and_docs(fixture)
    assert_no_secret_literals()
    print("production ops secret backend implementation readiness checks passed.")


if __name__ == "__main__":
    main()
