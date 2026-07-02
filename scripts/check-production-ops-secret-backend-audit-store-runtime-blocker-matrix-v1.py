#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-runtime-event-schema-artifact-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.json",
        "audit_store_runtime_event_schema_artifact_implemented",
    ),
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ),
    "production-secret-backend-audit-store-durable-backend-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.json",
        "audit_store_durable_backend_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-durable-backend-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-selection-readiness-v1.json",
        "audit_store_durable_backend_selection_readiness_defined",
    ),
    "production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.json"
        ),
        "audit_store_concrete_durable_backend_selection_review_defined",
    ),
    "production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json",
        "audit_store_writer_runtime_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.json"
        ),
        "audit_store_writer_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-audit-store-delivery-runtime-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-runtime-readiness-v1.json",
        "audit_store_delivery_runtime_readiness_defined",
    ),
    "production-secret-backend-audit-store-idempotency-runtime-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.json",
        "audit_store_idempotency_runtime_readiness_defined",
    ),
    "production-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1.json"
        ),
        "audit_store_idempotency_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1.json"
        ),
        "audit_store_delivery_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.json",
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.json"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_backend_product_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.json"
        ),
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1.json"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-entry-review-v1.json"
        ),
        "audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json"
        ),
        "audit_store_storage_adapter_metadata_contract_artifact_materialized",
    ),
    "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_backend_product_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.json"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.json"
        ),
        "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined",
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.json",
        "credential_handle_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.json",
        "operator_approval_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.json",
        "resolver_backend_health_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.json"
        ),
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2": (
        "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.json",
        "production_resolver_runtime_implementation_entry_refresh_v2_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}

EXPECTED_BOUNDARY = {
    "status": "audit_store_runtime_blocker_matrix_defined",
    "entry_decision": "audit_store_runtime_task_card_still_blocked_after_schema_artifact",
    "schema_artifact_status": "implemented_static_schema_artifact",
    "schema_artifact_validation_status": "implemented_offline_schema_validation",
    "writer_input_compatibility_status": "implemented_static_schema_compatibility",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "durable_backend_selection_readiness_status": "defined_without_backend_selection",
    "durable_backend_selection_decision": "durable_backend_selection_deferred_until_backend_evidence_and_runtime_task_card",
    "durable_backend_concrete_selection_review_status": (
        "audit_store_concrete_durable_backend_selection_review_defined"
    ),
    "durable_backend_selection_decision_after_review": (
        "durable_backend_family_selected_static_append_only_audit_log_runtime_blocked"
    ),
    "durable_audit_backend_status": (
        "storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined_task_card_blocked"
    ),
    "selected_durable_backend_family": "append_only_metadata_audit_log",
    "selected_reserved_candidate": "reserved_append_only_audit_log",
    "storage_adapter_runtime_implementation_entry_review_status": (
        "audit_store_storage_adapter_runtime_implementation_entry_review_defined"
    ),
    "storage_adapter_backend_product_evidence_readiness_status": (
        "audit_store_storage_adapter_backend_product_evidence_readiness_defined"
    ),
    "storage_adapter_metadata_contract_artifact_readiness_status": (
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined"
    ),
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "storage_adapter_backend_product_evidence_status": "readiness_defined_without_product_selection",
    "storage_adapter_backend_product_selection_review_status": (
        "audit_store_storage_adapter_backend_product_selection_review_defined"
    ),
    "storage_adapter_backend_product_selection_status": "selected_static_product_class_without_backend_provider",
    "storage_adapter_selected_backend_product_class": "managed_database_append_only_table",
    "storage_adapter_selected_backend_product_profile": "reserved_managed_database_append_only_table_profile",
    "storage_adapter_database_product_status": "not_selected",
    "storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_status": (
        "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined"
    ),
    "storage_adapter_database_provider_boundary_status": "metadata_only_provider_boundary_defined",
    "storage_adapter_database_driver_selection_policy_status": "static_driver_policy_defined_without_driver_selection",
    "storage_adapter_database_dsn_secret_ref_policy_status": "secret_ref_only_dsn_policy_defined",
    "storage_adapter_database_tls_policy_status": "tls_mode_policy_defined",
    "storage_adapter_database_role_policy_status": "least_privilege_role_policy_defined",
    "storage_adapter_database_connection_provider_status": "not_created",
    "storage_adapter_database_provider_driver_dsn_tls_role_policy_status": "defined_without_runtime",
    "storage_adapter_append_only_table_schema_boundary_status": "required_before_runtime_task_card",
    "storage_adapter_migration_schema_marker_boundary_status": "required_before_runtime_task_card",
    "storage_adapter_offline_adapter_smoke_strategy_status": "required_before_runtime_task_card",
    "storage_adapter_negative_leakage_runtime_scan_boundary_status": "required_before_runtime_task_card",
    "storage_adapter_backend_product_candidate_source_status": "metadata_only_candidate_source_defined",
    "storage_adapter_metadata_contract_artifact_status": "materialized_static_metadata_contract",
    "storage_adapter_contract_artifact_path_status": "materialized_static_path",
    "storage_adapter_input_envelope_status": "metadata_only_input_envelope_defined",
    "storage_adapter_result_envelope_status": "metadata_only_result_envelope_defined",
    "storage_adapter_record_identity_status": "metadata_only_record_identity_defined",
    "storage_adapter_failure_taxonomy_status": "metadata_only_failure_taxonomy_defined",
    "storage_adapter_writer_compatibility_status": "metadata_only_writer_compatibility_defined",
    "storage_adapter_contract_artifact_materialization_status": (
        "audit_store_storage_adapter_metadata_contract_artifact_materialized"
    ),
    "storage_adapter_contract_artifact_validation_status": "implemented_offline_contract_validation",
    "storage_adapter_writer_compatibility_smoke_status": "implemented_static_fixture",
    "storage_adapter_no_secret_material_scan_status": "implemented_static_scan",
    "storage_adapter_append_only_semantics_evidence_readiness_status": (
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined"
    ),
    "storage_adapter_append_only_semantics_status": "append_only_semantics_evidence_defined_without_runtime",
    "storage_adapter_append_only_operation_status": "append_only_insert_only",
    "storage_adapter_forbidden_mutation_policy_status": (
        "update_delete_overwrite_truncate_reject_policy_defined"
    ),
    "storage_adapter_sequence_reference_status": "metadata_only_monotonic_sequence_reference_defined",
    "storage_adapter_record_immutability_status": "metadata_only_immutability_policy_defined",
    "storage_adapter_duplicate_replay_policy_status": "fail_closed_duplicate_replay_reference_defined",
    "storage_adapter_retention_redaction_policy_evidence_readiness_status": (
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined"
    ),
    "storage_adapter_retention_redaction_status": "retention_redaction_policy_evidence_defined_without_runtime",
    "storage_adapter_retention_window_status": "metadata_only_retention_window_reference_defined",
    "storage_adapter_redaction_reference_status": "metadata_only_redaction_policy_reference_defined",
    "storage_adapter_retention_redaction_append_only_compatibility_status": (
        "append_only_immutability_compatible_policy_defined"
    ),
    "storage_adapter_forbidden_erasure_policy_status": "delete_overwrite_inline_redaction_forbidden",
    "storage_adapter_offline_validation_evidence_readiness_status": (
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined"
    ),
    "storage_adapter_offline_validation_status": "offline_validation_evidence_defined_without_runtime",
    "storage_adapter_offline_validation_manifest_status": (
        "metadata_only_offline_validation_manifest_reference_defined"
    ),
    "storage_adapter_offline_validation_positive_case_status": "metadata_only_positive_case_reference_defined",
    "storage_adapter_offline_validation_negative_case_status": "metadata_only_negative_case_reference_defined",
    "storage_adapter_offline_validation_coverage_status": (
        "metadata_contract_append_only_retention_redaction_coverage_defined"
    ),
    "storage_adapter_backend_touch_policy_status": "real_backend_touch_forbidden",
    "storage_adapter_validation_runner_status": "not_created",
    "storage_adapter_negative_leakage_scan_evidence_readiness_status": (
        "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined"
    ),
    "storage_adapter_negative_leakage_scan_evidence_status": (
        "negative_leakage_scan_evidence_defined_without_runtime"
    ),
    "storage_adapter_negative_leakage_scan_manifest_status": (
        "metadata_only_negative_leakage_scan_manifest_reference_defined"
    ),
    "storage_adapter_negative_leakage_scan_target_status": "metadata_only_scan_target_reference_defined",
    "storage_adapter_negative_leakage_forbidden_material_coverage_status": (
        "raw_payload_secret_credential_provider_backend_detail_coverage_defined"
    ),
    "storage_adapter_negative_leakage_diagnostic_allowlist_status": "metadata_only_diagnostic_allowlist_defined",
    "storage_adapter_negative_leakage_scan_status": "not_created",
    "storage_adapter_negative_leakage_scan_runner_status": "not_created",
    "storage_adapter_negative_leakage_scan_output_status": "not_created",
    "storage_adapter_rollback_recovery_evidence_readiness_status": (
        "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined"
    ),
    "storage_adapter_rollback_recovery_status": "rollback_recovery_evidence_defined_without_runtime",
    "storage_adapter_rollback_recovery_manifest_status": (
        "metadata_only_rollback_recovery_manifest_reference_defined"
    ),
    "storage_adapter_rollback_append_only_boundary_status": (
        "append_only_compensating_event_boundary_defined"
    ),
    "storage_adapter_partial_write_recovery_status": "metadata_only_partial_write_recovery_policy_defined",
    "storage_adapter_duplicate_replay_recovery_status": "fail_closed_replay_recovery_reference_defined",
    "storage_adapter_retention_redaction_recovery_alignment_status": (
        "append_only_retention_redaction_compatible_recovery_defined"
    ),
    "storage_adapter_negative_leakage_recovery_alignment_status": (
        "no_raw_material_recovery_diagnostics_defined"
    ),
    "storage_adapter_rollback_executor_status": "not_created",
    "storage_adapter_recovery_executor_status": "not_created",
    "storage_adapter_compensating_event_writer_status": "not_created",
    "storage_adapter_recovery_output_status": "not_created",
    "storage_adapter_runtime_implementation_entry_refresh_status": (
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_defined"
    ),
    "storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status": (
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined"
    ),
    "storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_database_provider_policy_readiness"
    ),
    "storage_adapter_evidence_chain_status": "static_evidence_chain_ready_for_contract_materialization_review",
    "storage_adapter_next_dependency": "storage_adapter_metadata_contract_artifact_materialization_entry_review",
    "storage_adapter_metadata_contract_artifact_materialization_entry_review_status": (
        "audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_defined"
    ),
    "storage_adapter_metadata_contract_artifact_materialization_status": (
        "audit_store_storage_adapter_metadata_contract_artifact_materialized"
    ),
    "storage_adapter_contract_artifact_materialization_task_card_decision": (
        "metadata_contract_artifact_materialization_task_card_ready_after_entry_review"
    ),
    "storage_adapter_contract_artifact_materialization_task_card_status": "created",
    "storage_adapter_current_next_dependency": "storage_adapter_append_only_table_schema_boundary_readiness",
    "writer_runtime_implementation_entry_review_status": (
        "audit_store_writer_runtime_implementation_entry_review_defined"
    ),
    "writer_runtime_task_card_status": "not_created",
    "audit_writer_runtime_status": "not_created",
    "delivery_runtime_implementation_entry_review_status": (
        "audit_store_delivery_runtime_implementation_entry_review_defined"
    ),
    "delivery_runtime_task_card_status": "not_created",
    "delivery_runtime_status": "not_created",
    "idempotency_runtime_implementation_entry_review_status": (
        "audit_store_idempotency_runtime_implementation_entry_review_defined"
    ),
    "idempotency_runtime_task_card_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "operator_approval_runtime_status": "not_created",
    "credential_handle_runtime_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "no_secret_leakage_smoke_runtime_status": "not_created",
    "production_resolver_runtime_task_card_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "cloud_secret_service_status": "not_selected",
    "database_connection_provider_status": "not_created",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "runtime_task_card_allowed_now",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "durable_audit_backend_selected_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "delivery_executed_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "duplicate_detector_created_in_this_slice",
    "operator_approval_runtime_created_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "credential_payload_created_in_this_slice",
    "backend_health_runtime_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "no_secret_leakage_smoke_runtime_created_in_this_slice",
    "production_resolver_runtime_task_card_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "cloud_secret_client_created_in_this_slice",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_BLOCKERS = {
    "runtime_event_schema_artifact": "implemented_static_schema_artifact",
    "durable_audit_backend": (
        "storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined_task_card_blocked"
    ),
    "audit_writer_runtime": "entry_review_defined_task_card_blocked",
    "idempotency_runtime": "entry_review_defined_task_card_blocked",
    "delivery_runtime": "entry_review_defined_task_card_blocked",
    "operator_approval_runtime": "not_created",
    "credential_handle_runtime": "not_created",
    "backend_health_runtime": "not_created",
    "no_secret_leakage_smoke_runtime": "not_created",
    "production_resolver_runtime": "not_created",
}

EXPECTED_ORDER = [
    "runtime_event_schema_artifact_implemented",
    "durable_backend_selection_readiness",
    "concrete_durable_backend_selection_review",
    "storage_adapter_runtime_entry_review",
    "storage_adapter_backend_product_evidence_readiness",
    "storage_adapter_metadata_contract_artifact_readiness",
    "storage_adapter_append_only_semantics_evidence_readiness",
    "storage_adapter_retention_redaction_policy_evidence_readiness",
    "storage_adapter_offline_validation_evidence_readiness",
    "storage_adapter_negative_leakage_scan_evidence_readiness",
    "storage_adapter_rollback_recovery_evidence_readiness",
    "storage_adapter_runtime_implementation_entry_refresh",
    "storage_adapter_metadata_contract_artifact_materialization_entry_review",
    "storage_adapter_metadata_contract_artifact_materialization",
    "storage_adapter_backend_product_selection_review",
    "storage_adapter_runtime_entry_refresh_after_product_selection",
    "storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness",
    "audit_writer_runtime_entry_review",
    "idempotency_runtime_entry_review",
    "delivery_runtime_entry_review",
    "operator_approval_runtime_entry_refresh",
    "credential_handle_runtime_entry_refresh",
    "backend_health_runtime_entry_refresh",
    "no_leakage_smoke_runtime_entry_refresh",
    "audit_store_runtime_entry_refresh_after_blocker_matrix",
    "production_resolver_runtime_entry_refresh_after_audit_store_runtime",
]

EXPECTED_FAILURE_CODES = {
    "audit_store_runtime_blocker_matrix_dependency_missing",
    "audit_store_runtime_blocker_matrix_schema_artifact_not_consumed",
    "audit_store_runtime_blocker_matrix_runtime_task_card_forbidden",
    "audit_store_runtime_blocker_matrix_runtime_created_forbidden",
    "audit_store_runtime_blocker_matrix_secret_material_detected",
    "audit_store_runtime_blocker_matrix_production_resolver_bypass",
}

EXPECTED_DIAGNOSTICS = {
    "audit_store_runtime_blocker_matrix_status",
    "entry_decision",
    "blocker_id",
    "blocker_status",
    "unlock_condition",
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
    "cloud_secret_call_count",
    "provider_call_count",
    "production_resolver_call_count",
    "audit_store_runtime_created_count",
    "audit_writer_runtime_created_count",
    "audit_event_write_count",
    "delivery_execution_count",
    "idempotency_decision_count",
    "duplicate_detection_count",
    "operator_approval_runtime_execution_count",
    "credential_handle_runtime_created_count",
    "credential_payload_created_count",
    "backend_health_runtime_created_count",
    "backend_health_check_count",
    "no_secret_leakage_smoke_runtime_created_count",
    "database_connection_count",
    "sql_execution_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}

EXPECTED_REQUIRED_CHECKS = {
    "run audit store runtime blocker matrix checker",
    "run audit store durable backend selection readiness checker",
    "run audit store concrete durable backend selection review checker",
    "run audit store writer runtime implementation entry review checker",
    "run audit store idempotency runtime implementation entry review checker",
    "run audit store delivery runtime implementation entry review checker",
    "run audit store storage adapter runtime implementation entry review checker",
    "run audit store storage adapter backend product evidence readiness checker",
    "run audit store storage adapter metadata contract artifact readiness checker",
    "run audit store storage adapter append-only semantics evidence readiness checker",
    "run audit store storage adapter retention redaction policy evidence readiness checker",
    "run audit store storage adapter offline validation evidence readiness checker",
    "run audit store storage adapter negative leakage scan evidence readiness checker",
    "run audit store storage adapter rollback recovery evidence readiness checker",
    "run audit store storage adapter runtime implementation entry refresh checker",
    "run audit store storage adapter metadata contract artifact materialization entry review checker",
    "run audit store storage adapter metadata contract artifact materialization checker",
    "run audit store storage adapter runtime implementation entry refresh after product selection checker",
    "run audit store storage adapter database provider driver DSN TLS role policy readiness checker",
    "run audit store runtime event schema artifact checker",
    "run audit store runtime implementation entry refresh v4 checker",
    "run production resolver runtime implementation entry refresh v2 checker",
    "run implementation readiness checker",
    "run git diff check",
    "run fast repository check",
}


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


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_audit_store_runtime_blocker_matrix_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-runtime-blocker-matrix-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "audit_store_runtime_blocker_matrix_defined", "unexpected status")
    require(
        slice_info.get("entry_decision") == "audit_store_runtime_task_card_still_blocked_after_schema_artifact",
        "unexpected entry decision",
    )
    for field in ("task_card", "platform_topic"):
        relative_path = str(slice_info.get(field) or "")
        require(relative_path, f"{field} missing")
        require((REPO_ROOT / relative_path).exists(), f"{field} path missing: {relative_path}")
    for claim in {
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "durable_audit_backend_selected",
        "audit_writer_runtime_created",
        "delivery_runtime_created",
        "idempotency_runtime_created",
        "production_resolver_runtime_task_card_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in set(slice_info.get("does_not_claim") or []), f"missing forbidden claim: {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        require(item.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"{dependency_id} source status drifted")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("matrix_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"matrix_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"matrix_boundary.{field} must remain false")


def assert_blockers_and_order(fixture: dict[str, Any]) -> None:
    blockers = rows_by_id(fixture, "blocker_matrix", "blocker_id")
    require(set(blockers) == set(EXPECTED_BLOCKERS), "blocker ids drifted")
    for blocker_id, expected_status in EXPECTED_BLOCKERS.items():
        item = blockers[blocker_id]
        require(item.get("status") == expected_status, f"{blocker_id} status drifted")
        require(item.get("source"), f"{blocker_id} source missing")
        require(item.get("unlock_condition"), f"{blocker_id} unlock condition missing")
    require(
        blockers["runtime_event_schema_artifact"].get("blocks_audit_store_runtime_task_card") is False,
        "schema artifact must no longer block audit runtime task card by itself",
    )
    require(
        blockers["runtime_event_schema_artifact"].get("blocks_production_resolver_task_card") is False,
        "schema artifact must not directly block production resolver after implementation",
    )
    for blocker_id in set(EXPECTED_BLOCKERS) - {"runtime_event_schema_artifact", "production_resolver_runtime"}:
        require(
            blockers[blocker_id].get("blocks_audit_store_runtime_task_card") is True,
            f"{blocker_id} must block audit runtime task card",
        )
        require(
            blockers[blocker_id].get("blocks_production_resolver_task_card") is True,
            f"{blocker_id} must block production resolver task card",
        )
    require(
        blockers["production_resolver_runtime"].get("blocks_audit_store_runtime_task_card") is False,
        "production resolver must not be an audit store runtime prerequisite",
    )
    require(
        blockers["production_resolver_runtime"].get("blocks_production_resolver_task_card") is True,
        "production resolver runtime gate must remain blocked",
    )
    require(fixture.get("dependency_order") == EXPECTED_ORDER, "dependency order drifted")


def assert_prior_evidence_alignment() -> None:
    artifact = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-runtime-event-schema-artifact-v1"
    ][0])
    artifact_boundary = artifact.get("artifact_boundary") or {}
    require(
        artifact_boundary.get("status") == "audit_runtime_event_schema_artifact_implemented_static",
        "schema artifact boundary status drifted",
    )
    require(
        artifact_boundary.get("audit_store_runtime_created_in_this_slice") is False,
        "schema artifact must not create audit runtime",
    )

    v4 = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4"
    ][0])
    v4_boundary = v4.get("entry_refresh_boundary") or {}
    require(
        v4_boundary.get("entry_decision") == "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
        "v4 entry decision drifted",
    )
    require(v4_boundary.get("audit_store_runtime_task_card_status") == "not_created", "v4 opened task card")
    require(v4_boundary.get("durable_audit_backend_status") == "not_selected", "durable backend selected")
    require(v4_boundary.get("audit_writer_runtime_status") == "not_created", "writer runtime created")
    require(v4_boundary.get("delivery_runtime_status") == "not_created", "delivery runtime created")
    require(v4_boundary.get("idempotency_runtime_status") == "not_created", "idempotency runtime created")

    selection = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-durable-backend-selection-readiness-v1"
    ][0])
    selection_boundary = selection.get("selection_boundary") or {}
    require(
        selection_boundary.get("status") == "audit_store_durable_backend_selection_readiness_defined",
        "durable backend selection readiness status drifted",
    )
    require(
        selection_boundary.get("durable_backend_selection_status") == "deferred_without_backend_selection",
        "durable backend selection must remain deferred",
    )
    require(
        selection_boundary.get("durable_audit_backend_status") == "not_selected",
        "durable backend selection readiness selected backend",
    )

    resolver = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2"
    ][0])
    resolver_boundary = resolver.get("refresh_boundary") or {}
    require(
        resolver_boundary.get("entry_decision") == "production_resolver_runtime_task_card_still_blocked_after_refresh_v2",
        "production resolver v2 decision drifted",
    )
    require(resolver_boundary.get("production_resolver_runtime_task_card_status") == "not_created", "resolver task card created")
    require(
        resolver_boundary.get("audit_store_entry_refresh_v4_status")
        == "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
        "resolver v2 audit store refresh v4 status drifted",
    )

    readiness = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES["production-secret-backend-implementation-readiness"][0])
    target = readiness.get("implementation_target") or {}
    for field, expected in {
        "audit_runtime_event_schema_artifact_status": "implemented_static_schema_artifact",
        "audit_runtime_event_schema_artifact_validation_status": "implemented_offline_schema_validation",
        "audit_store_durable_backend_selection_readiness_status": "defined_without_backend_selection",
        "audit_store_storage_adapter_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "audit_store_storage_adapter_backend_product_evidence_readiness_status": (
            "audit_store_storage_adapter_backend_product_evidence_readiness_defined"
        ),
        "audit_storage_adapter_backend_product_evidence_status": (
            "readiness_defined_without_product_selection"
        ),
        "audit_store_storage_adapter_backend_product_selection_review_status": (
            "audit_store_storage_adapter_backend_product_selection_review_defined"
        ),
        "audit_storage_adapter_backend_product_selection_status": (
            "selected_static_product_class_without_backend_provider"
        ),
        "audit_storage_adapter_selected_backend_product_class": "managed_database_append_only_table",
        "audit_storage_adapter_selected_backend_product_profile": (
            "reserved_managed_database_append_only_table_profile"
        ),
        "audit_storage_adapter_database_product_status": "not_selected",
        "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_status": (
            "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined"
        ),
        "audit_storage_adapter_database_provider_boundary_status": (
            "metadata_only_provider_boundary_defined"
        ),
        "audit_storage_adapter_database_driver_selection_policy_status": (
            "static_driver_policy_defined_without_driver_selection"
        ),
        "audit_storage_adapter_database_dsn_secret_ref_policy_status": (
            "secret_ref_only_dsn_policy_defined"
        ),
        "audit_storage_adapter_database_tls_policy_status": "tls_mode_policy_defined",
        "audit_storage_adapter_database_role_policy_status": "least_privilege_role_policy_defined",
        "audit_storage_adapter_database_connection_provider_status": "not_created",
        "audit_storage_adapter_backend_product_candidate_source_status": (
            "metadata_only_candidate_source_defined"
        ),
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_status": (
            "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined"
        ),
        "audit_storage_adapter_metadata_contract_artifact_status": (
            "materialized_static_metadata_contract"
        ),
        "audit_storage_adapter_contract_artifact_path_status": "materialized_static_path",
        "audit_storage_adapter_input_envelope_status": "metadata_only_input_envelope_defined",
        "audit_storage_adapter_result_envelope_status": "metadata_only_result_envelope_defined",
        "audit_storage_adapter_record_identity_status": "metadata_only_record_identity_defined",
        "audit_storage_adapter_failure_taxonomy_status": "metadata_only_failure_taxonomy_defined",
        "audit_storage_adapter_writer_compatibility_status": "metadata_only_writer_compatibility_defined",
        "audit_storage_adapter_contract_artifact_materialization_status": (
            "audit_store_storage_adapter_metadata_contract_artifact_materialized"
        ),
        "audit_storage_adapter_contract_artifact_validation_status": (
            "implemented_offline_contract_validation"
        ),
        "audit_storage_adapter_writer_compatibility_smoke_status": "implemented_static_fixture",
        "audit_storage_adapter_no_secret_material_scan_status": "implemented_static_scan",
        "audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_status": (
            "audit_store_storage_adapter_metadata_contract_artifact_materialization_entry_review_defined"
        ),
        "audit_storage_adapter_contract_materialization_task_card_decision": (
            "metadata_contract_artifact_materialization_task_card_ready_after_entry_review"
        ),
        "audit_storage_adapter_contract_materialization_task_card_status": "created",
        "audit_storage_adapter_current_next_dependency": (
            "storage_adapter_append_only_table_schema_boundary_readiness"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status": (
            "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined"
        ),
        "audit_storage_adapter_runtime_task_card_decision": (
            "storage_adapter_runtime_task_card_still_blocked_after_database_provider_policy_readiness"
        ),
        "audit_storage_adapter_database_provider_driver_dsn_tls_role_policy_status": (
            "defined_without_runtime"
        ),
        "audit_storage_adapter_append_only_table_schema_boundary_status": (
            "required_before_runtime_task_card"
        ),
        "audit_storage_adapter_migration_schema_marker_boundary_status": (
            "required_before_runtime_task_card"
        ),
        "audit_storage_adapter_offline_adapter_smoke_strategy_status": (
            "required_before_runtime_task_card"
        ),
        "audit_storage_adapter_negative_leakage_runtime_scan_boundary_status": (
            "required_before_runtime_task_card"
        ),
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_status": (
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined"
        ),
        "audit_storage_adapter_append_only_semantics_status": (
            "append_only_semantics_evidence_defined_without_runtime"
        ),
        "audit_storage_adapter_append_only_operation_status": "append_only_insert_only",
        "audit_storage_adapter_forbidden_mutation_policy_status": (
            "update_delete_overwrite_truncate_reject_policy_defined"
        ),
        "audit_storage_adapter_sequence_reference_status": (
            "metadata_only_monotonic_sequence_reference_defined"
        ),
        "audit_storage_adapter_record_immutability_status": "metadata_only_immutability_policy_defined",
        "audit_storage_adapter_duplicate_replay_policy_status": (
            "fail_closed_duplicate_replay_reference_defined"
        ),
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_status": (
            "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined"
        ),
        "audit_storage_adapter_retention_redaction_status": (
            "retention_redaction_policy_evidence_defined_without_runtime"
        ),
        "audit_storage_adapter_retention_window_status": (
            "metadata_only_retention_window_reference_defined"
        ),
        "audit_storage_adapter_redaction_reference_status": (
            "metadata_only_redaction_policy_reference_defined"
        ),
        "audit_storage_adapter_retention_redaction_append_only_compatibility_status": (
            "append_only_immutability_compatible_policy_defined"
        ),
        "audit_storage_adapter_forbidden_erasure_policy_status": (
            "delete_overwrite_inline_redaction_forbidden"
        ),
        "audit_store_storage_adapter_offline_validation_evidence_readiness_status": (
            "audit_store_storage_adapter_offline_validation_evidence_readiness_defined"
        ),
        "audit_storage_adapter_offline_validation_status": (
            "offline_validation_evidence_defined_without_runtime"
        ),
        "audit_storage_adapter_offline_validation_manifest_status": (
            "metadata_only_offline_validation_manifest_reference_defined"
        ),
        "audit_storage_adapter_offline_validation_positive_case_status": (
            "metadata_only_positive_case_reference_defined"
        ),
        "audit_storage_adapter_offline_validation_negative_case_status": (
            "metadata_only_negative_case_reference_defined"
        ),
        "audit_storage_adapter_offline_validation_coverage_status": (
            "metadata_contract_append_only_retention_redaction_coverage_defined"
        ),
        "audit_storage_adapter_backend_touch_policy_status": "real_backend_touch_forbidden",
        "audit_storage_adapter_validation_runner_status": "not_created",
        "audit_store_runtime_blocker_matrix_status": "audit_store_runtime_blocker_matrix_defined",
        "audit_store_runtime_task_card_status": "not_created",
        "audit_store_runtime_status": "not_created",
        "audit_writer_status": "not_created",
        "audit_delivery_runtime_status": "not_created",
        "audit_idempotency_runtime_status": "not_created",
        "production_resolver_runtime_task_card_status": "not_created",
        "production_resolver_runtime_status": "not_created",
    }.items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")


def assert_policies_and_diagnostics(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} boundary missing")
        require(item.get("sanitized_diagnostic"), f"{code} diagnostic missing")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    require(EXPECTED_DIAGNOSTICS.issubset(set(diagnostics.get("allowed_fields") or [])), "diagnostics allowlist drifted")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must be false")
    require(diagnostics.get("secret_ref_value_allowed_in_diagnostics") is False, "secret ref value must not emit")

    fallback = set(fixture.get("no_fallback_policy") or [])
    for rule in {
        "schema artifact completion does not mean audit store runtime ready",
        "blocker matrix does not mean production resolver runtime ready",
        "blocker matrix does not mean workflow repository mode ready",
        "no fallback from writer runtime to schema fixture",
    }:
        require(rule in fallback, f"missing no fallback rule: {rule}")

    side_effects = set(fixture.get("no_side_effect_policy") or [])
    for rule in {
        "no cloud secret service call",
        "no production resolver call",
        "no audit event write",
        "no delivery execution",
        "no idempotency decision",
        "no database connection",
        "no SQL execution",
        "no repository mode enablement",
    }:
        require(rule in side_effects, f"missing no side effect rule: {rule}")

    counters = fixture.get("side_effect_counters") or {}
    require(set(counters) == EXPECTED_ZERO_COUNTERS, "side effect counters drifted")
    for counter, value in counters.items():
        require(value == 0, f"{counter} must stay zero")


def assert_artifact_guard_and_docs(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("required_static_artifacts_exist") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required static artifact missing: {relative_path}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")

    for reference in fixture.get("required_doc_references") or []:
        text = read(str(reference.get("path") or ""))
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in text]
        require(not missing, f"{reference.get('path')} missing literals: {missing}")

    require(set(fixture.get("validation_strategy") or []) == EXPECTED_REQUIRED_CHECKS, "validation strategy drifted")
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    artifact_call = 'run_python_script("check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py", [])'
    selection_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'durable-backend-selection-readiness-v1.py", [])'
    )
    writer_entry_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'writer-runtime-implementation-entry-review-v1.py", [])'
    )
    idempotency_entry_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'idempotency-runtime-implementation-entry-review-v1.py", [])'
    )
    delivery_entry_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'delivery-runtime-implementation-entry-review-v1.py", [])'
    )
    v5_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'runtime-implementation-entry-refresh-v5.py", [])'
    )
    storage_entry_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-runtime-implementation-entry-review-v1.py", [])'
    )
    backend_product_evidence_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-backend-product-evidence-readiness-v1.py", [])'
    )
    metadata_contract_artifact_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-metadata-contract-artifact-readiness-v1.py", [])'
    )
    append_only_semantics_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-append-only-semantics-evidence-readiness-v1.py", [])'
    )
    retention_redaction_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-retention-redaction-policy-evidence-readiness-v1.py", [])'
    )
    offline_validation_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-offline-validation-evidence-readiness-v1.py", [])'
    )
    negative_leakage_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-negative-leakage-scan-evidence-readiness-v1.py", [])'
    )
    rollback_recovery_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-rollback-recovery-evidence-readiness-v1.py", [])'
    )
    storage_adapter_runtime_refresh_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-runtime-implementation-entry-refresh-v1.py", [])'
    )
    storage_adapter_materialization_entry_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-metadata-contract-artifact-materialization-entry-review-v1.py", [])'
    )
    storage_adapter_materialization_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-metadata-contract-artifact-materialization-v1.py", [])'
    )
    storage_adapter_backend_product_selection_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-backend-product-selection-review-v1.py", [])'
    )
    storage_adapter_after_product_selection_refresh_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.py", [])'
    )
    storage_adapter_database_provider_policy_readiness_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.py", [])'
    )
    current_call = 'run_python_script("check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py", [])'
    resolver_call = (
        'run_python_script("check-production-ops-secret-backend-'
        'production-resolver-runtime-implementation-entry-refresh-v2.py", [])'
    )
    for call in (
        artifact_call,
        selection_call,
        writer_entry_call,
        idempotency_entry_call,
        delivery_entry_call,
        v5_call,
        storage_entry_call,
        backend_product_evidence_call,
        metadata_contract_artifact_call,
        append_only_semantics_call,
        retention_redaction_call,
        offline_validation_call,
        negative_leakage_call,
        rollback_recovery_call,
        storage_adapter_runtime_refresh_call,
        storage_adapter_materialization_entry_call,
        storage_adapter_materialization_call,
        storage_adapter_backend_product_selection_call,
        storage_adapter_after_product_selection_refresh_call,
        storage_adapter_database_provider_policy_readiness_call,
        current_call,
        resolver_call,
    ):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(
        check_repo.index(artifact_call)
        < check_repo.index(selection_call)
        < check_repo.index(writer_entry_call)
        < check_repo.index(idempotency_entry_call)
        < check_repo.index(delivery_entry_call)
        < check_repo.index(v5_call)
        < check_repo.index(storage_entry_call)
        < check_repo.index(backend_product_evidence_call)
        < check_repo.index(metadata_contract_artifact_call)
        < check_repo.index(append_only_semantics_call)
        < check_repo.index(retention_redaction_call)
        < check_repo.index(offline_validation_call)
        < check_repo.index(negative_leakage_call)
        < check_repo.index(rollback_recovery_call)
        < check_repo.index(storage_adapter_runtime_refresh_call)
        < check_repo.index(storage_adapter_materialization_entry_call)
        < check_repo.index(storage_adapter_materialization_call)
        < check_repo.index(storage_adapter_backend_product_selection_call)
        < check_repo.index(storage_adapter_after_product_selection_refresh_call)
        < check_repo.index(storage_adapter_database_provider_policy_readiness_call)
        < check_repo.index(current_call)
        < check_repo.index(resolver_call),
        "check order drifted",
    )


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-runtime-blocker-matrix-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"blocker matrix artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "blocker matrix artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "blocker matrix artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_blockers_and_order(fixture)
    assert_prior_evidence_alignment()
    assert_policies_and_diagnostics(fixture)
    assert_artifact_guard_and_docs(fixture)
    assert_no_secret_literals()
    print("production ops secret backend audit store runtime blocker matrix checks passed.")


if __name__ == "__main__":
    main()
