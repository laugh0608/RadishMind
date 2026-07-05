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
    "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
FOLLOWUP_MATERIALIZATION_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json"
)
FOLLOWUP_MATERIALIZATION_STATUS = "audit_store_storage_adapter_metadata_contract_artifact_materialized"
FOLLOWUP_SELECTION_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json"
)
FOLLOWUP_SELECTION_STATUS = "audit_store_storage_adapter_backend_product_selection_review_defined"
FOLLOWUP_SELECTION_NEXT_DEPENDENCY = "storage_adapter_runtime_implementation_entry_refresh_after_product_selection"
FOLLOWUP_AFTER_SELECTION_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.json"
)
FOLLOWUP_AFTER_SELECTION_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined"
)
FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY = "storage_adapter_database_driver_selection_readiness"
FOLLOWUP_AFTER_SELECTION_BLOCKER_STATUS = (
    "storage_adapter_database_provider_selection_review_defined_task_card_blocked"
)
FOLLOWUP_AFTER_SELECTION_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1"
)
FOLLOWUP_ALIGNMENT = {
    "audit_storage_adapter_metadata_contract_artifact_status": "materialized_static_metadata_contract",
    "audit_storage_adapter_contract_artifact_path_status": "materialized_static_path",
    "audit_storage_adapter_contract_artifact_materialization_status": FOLLOWUP_MATERIALIZATION_STATUS,
}
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
        "storage_adapter_runtime_task_card_still_blocked_after_database_provider_selection_review"
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

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.json"
        ),
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_backend_product_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.json"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
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
    "status": "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
    "readiness_decision": "negative_leakage_scan_evidence_defined_without_runtime",
    "offline_validation_evidence_readiness_status": (
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined"
    ),
    "offline_validation_status": "offline_validation_evidence_defined_without_runtime",
    "retention_redaction_policy_evidence_readiness_status": (
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined"
    ),
    "retention_redaction_status": "retention_redaction_policy_evidence_defined_without_runtime",
    "append_only_semantics_evidence_readiness_status": (
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined"
    ),
    "append_only_semantics_status": "append_only_semantics_evidence_defined_without_runtime",
    "metadata_contract_artifact_readiness_status": (
        "audit_store_storage_adapter_metadata_contract_artifact_readiness_defined"
    ),
    "metadata_contract_artifact_status": "readiness_defined_without_materialized_artifact",
    "backend_product_evidence_readiness_status": (
        "audit_store_storage_adapter_backend_product_evidence_readiness_defined"
    ),
    "backend_product_selection_status": "not_selected",
    "selected_backend_family": "append_only_metadata_audit_log",
    "selected_reserved_candidate": "reserved_append_only_audit_log",
    "negative_leakage_scan_evidence_status": "negative_leakage_scan_evidence_defined",
    "negative_leakage_scan_status": "not_created",
    "negative_leakage_scan_manifest_status": "metadata_only_negative_leakage_scan_manifest_reference_defined",
    "scan_target_reference_status": "metadata_only_scan_target_reference_defined",
    "forbidden_material_coverage_status": "raw_payload_secret_credential_provider_backend_detail_coverage_defined",
    "diagnostic_allowlist_status": "metadata_only_diagnostic_allowlist_defined",
    "scan_runner_status": "not_created",
    "scan_output_status": "not_created",
    "validation_runner_status": "not_created",
    "validation_output_status": "not_created",
    "rollback_recovery_status": "required_before_runtime_task_card",
    "next_dependency": "storage_adapter_rollback_recovery_evidence_readiness",
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "storage_adapter_client_status": "not_created",
    "retention_executor_status": "not_created",
    "redaction_executor_status": "not_created",
    "database_connection_provider_status": "blocked",
    "database_driver_status": "not_selected",
    "sql_migration_status": "not_created",
    "schema_marker_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_writer_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "delivery_runtime_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}
EXPECTED_FALSE_FLAGS = {
    "backend_product_selected_in_this_slice",
    "contract_artifact_materialized_in_this_slice",
    "offline_validation_runner_created_in_this_slice",
    "offline_validation_executed_in_this_slice",
    "offline_validation_output_committed_in_this_slice",
    "negative_leakage_scanner_created_in_this_slice",
    "negative_leakage_scan_executed_in_this_slice",
    "negative_leakage_scan_output_committed_in_this_slice",
    "retention_runtime_created_in_this_slice",
    "redaction_runtime_created_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "storage_adapter_client_created_in_this_slice",
    "database_connection_provider_enabled",
    "sql_migration_created_in_this_slice",
    "schema_marker_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "duplicate_detector_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}
EXPECTED_REFERENCE_FIELDS = {
    "negative_leakage_scan_manifest_ref",
    "scan_target_manifest_ref",
    "offline_validation_evidence_ref",
    "forbidden_material_matrix_ref",
    "diagnostic_allowlist_ref",
    "coverage_matrix_ref",
    "failure_taxonomy_ref",
    "policy_version",
    "audit_ref",
}
EXPECTED_FORBIDDEN_MATERIAL_CLASSES = {
    "raw_request_payload",
    "raw_response_payload",
    "raw_audit_payload",
    "raw_event_payload",
    "raw_writer_payload",
    "raw_storage_payload",
    "raw_retained_payload",
    "raw_redacted_payload",
    "secret_material",
    "credential_material",
    "payload_hash",
    "secret_derived_hash",
    "provider_detail",
    "backend_detail",
    "scanner_raw_finding",
    "scan_output",
}
EXPECTED_COVERAGE_IDS = {
    "raw_payload_material",
    "secret_credential_material",
    "provider_backend_detail",
    "offline_validation_evidence_reference",
    "diagnostic_allowlist",
    "artifact_guard",
}
EXPECTED_NEGATIVE_CASES = {
    "raw_request_payload_forbidden_case",
    "raw_response_payload_forbidden_case",
    "raw_audit_payload_forbidden_case",
    "raw_storage_payload_forbidden_case",
    "secret_credential_material_forbidden_case",
    "provider_backend_detail_forbidden_case",
    "payload_hash_forbidden_case",
    "scan_output_forbidden_case",
    "fallback_forbidden_case",
}
EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_negative_leakage_scan_dependency_missing",
    "audit_store_storage_adapter_negative_leakage_scan_manifest_reference_missing",
    "audit_store_storage_adapter_negative_leakage_scan_target_reference_missing",
    "audit_store_storage_adapter_negative_leakage_forbidden_material_coverage_missing",
    "audit_store_storage_adapter_negative_leakage_diagnostic_allowlist_missing",
    "audit_store_storage_adapter_negative_leakage_raw_material_detected",
    "audit_store_storage_adapter_negative_leakage_scanner_runtime_scope_overreach",
    "audit_store_storage_adapter_negative_leakage_fallback_detected",
    "audit_store_storage_adapter_negative_leakage_next_dependency_missing",
}
EXPECTED_ALLOWED_ARTIFACTS = {
    (
        "docs/platform/"
        "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.md"
    ),
    (
        "docs/task-cards/"
        "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1-plan.md"
    ),
    (
        "scripts/checks/fixtures/"
        "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.json"
    ),
    "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.py",
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


def followup_materialization_exists() -> bool:
    path = REPO_ROOT / FOLLOWUP_MATERIALIZATION_FIXTURE
    if not path.exists():
        return False
    materialization = load_json(path)
    return source_status(materialization) == FOLLOWUP_MATERIALIZATION_STATUS


def followup_selection_exists() -> bool:
    path = REPO_ROOT / FOLLOWUP_SELECTION_FIXTURE
    if not path.exists():
        return False
    selection = load_json(path)
    return source_status(selection) == FOLLOWUP_SELECTION_STATUS


def followup_after_selection_exists() -> bool:
    path = REPO_ROOT / FOLLOWUP_AFTER_SELECTION_FIXTURE
    if not path.exists():
        return False
    followup = load_json(path)
    return source_status(followup) == FOLLOWUP_AFTER_SELECTION_STATUS


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id")
        == "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        "unexpected status",
    )
    require(
        slice_info.get("readiness_decision") == "negative_leakage_scan_evidence_defined_without_runtime",
        "unexpected readiness decision",
    )
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "negative_leakage_scanner_created",
        "negative_leakage_scan_executed",
        "negative_leakage_scan_output_committed",
        "storage_adapter_runtime_task_card_created",
        "storage_adapter_runtime_created",
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in claims, f"does_not_claim missing {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"{dependency_id} source status drifted")


def assert_readiness_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("readiness_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"readiness_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"readiness_boundary.{field} must stay false")


def assert_negative_leakage_evidence(fixture: dict[str, Any]) -> None:
    evidence = fixture.get("negative_leakage_scan_evidence") or {}
    require(
        evidence.get("status") == "metadata_only_negative_leakage_scan_evidence_defined",
        "negative leakage evidence status drifted",
    )
    require(
        set(evidence.get("required_reference_fields") or []) == EXPECTED_REFERENCE_FIELDS,
        "negative leakage reference fields drifted",
    )
    require(
        set(evidence.get("forbidden_material_classes") or []) == EXPECTED_FORBIDDEN_MATERIAL_CLASSES,
        "forbidden material classes drifted",
    )
    for mechanism in {
        "negative_leakage_scanner",
        "scan_cli",
        "runtime_smoke",
        "storage_payload_reader",
        "provider_log_reader",
        "backend_write_probe",
        "provider_probe",
        "committed_scan_output",
    }:
        require(mechanism in set(evidence.get("forbidden_runtime_mechanisms") or []), f"missing {mechanism}")
    for touch in {
        "database_connection",
        "object_store_call",
        "queue_call",
        "topic_call",
        "log_sink_call",
        "vendor_service_call",
        "provider_call",
        "cloud_secret_call",
    }:
        require(touch in set(evidence.get("forbidden_backend_touches") or []), f"missing backend touch {touch}")

    coverage = rows_by_id(fixture, "coverage_matrix", "id")
    require(set(coverage) == EXPECTED_COVERAGE_IDS, "coverage matrix ids drifted")
    for coverage_id, row in coverage.items():
        require(row.get("status"), f"{coverage_id} status missing")
        require(row.get("requires"), f"{coverage_id} requirements missing")
        require(row.get("does_not_claim"), f"{coverage_id} does_not_claim missing")
    require(
        set(fixture.get("negative_case_requirements") or []) == EXPECTED_NEGATIVE_CASES,
        "negative case requirements drifted",
    )


def assert_diagnostics_failures_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(not (allowed & forbidden), "diagnostic allowlist intersects forbidden fields")
    require(sample.get("scan_runner_status") == "not_created", "diagnostic sample created scan runner")
    require(sample.get("scan_output_status") == "not_created", "diagnostic sample created scan output")
    require(sample.get("negative_leakage_scan_status") == "not_created", "diagnostic sample created scan")
    require(
        sample.get("next_dependency") == "storage_adapter_rollback_recovery_evidence_readiness",
        "diagnostic sample next dependency drifted",
    )

    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} boundary missing")
        require(item.get("sanitized_diagnostic"), f"{code} diagnostic missing")

    no_fallback = fixture.get("no_fallback_policy") or {}
    require(no_fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for source in {
        "offline_validation_evidence",
        "retention_redaction_policy_evidence",
        "append_only_semantics_evidence",
        "metadata_contract_artifact_readiness",
        "backend_product_evidence_readiness",
        "storage_adapter_runtime_entry_review",
        "historical_smoke",
        "previous_checker_success",
    }:
        require(source in set(no_fallback.get("forbidden_sources") or []), f"missing forbidden fallback {source}")

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for counter, value in counters.items():
        require(value == 0, f"{counter} must stay 0")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    require(set(guard.get("allowed_added_artifacts") or []) == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifacts drifted")
    for path in EXPECTED_ALLOWED_ARTIFACTS:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing: {path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "negative_leakage_scanner",
        "scan_cli",
        "scan_output_artifact",
        "offline_validation_runner",
        "validation_cli",
        "validation_output_artifact",
        "storage_adapter_runtime_implementation_task_card",
        "storage_adapter_runtime",
        "database_connection_provider",
        "sql_migration",
        "audit_store_runtime_implementation_task_card",
        "audit_store_runtime",
        "audit_writer_runtime",
        "delivery_runtime",
        "idempotency_runtime",
        "duplicate_detector",
        "production_resolver_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact missing: {artifact}")
    for path in guard.get("files_must_not_exist") or []:
        if (
            followup_materialization_exists()
            and path == "contracts/production-secret-audit-storage-adapter.metadata-contract.json"
        ):
            continue
        require(not (REPO_ROOT / str(path)).exists(), f"forbidden artifact exists: {path}")


def assert_blocker_matrix_alignment() -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    require(
        boundary.get("storage_adapter_negative_leakage_scan_evidence_readiness_status")
        == "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        "matrix boundary missing negative leakage readiness status",
    )
    require(
        boundary.get("storage_adapter_negative_leakage_scan_evidence_status")
        == "negative_leakage_scan_evidence_defined_without_runtime",
        "matrix boundary negative leakage evidence status drifted",
    )
    require(
        boundary.get("storage_adapter_negative_leakage_scan_status") == "not_created",
        "matrix boundary must not create negative leakage scan",
    )
    require(
        boundary.get("storage_adapter_rollback_recovery_status")
        == "rollback_recovery_evidence_defined_without_runtime",
        "matrix boundary rollback recovery status drifted",
    )
    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    if followup_after_selection_exists():
        expected_status = FOLLOWUP_AFTER_SELECTION_BLOCKER_STATUS
        expected_source = FOLLOWUP_AFTER_SELECTION_SOURCE
    elif followup_selection_exists():
        expected_status = "storage_adapter_backend_product_selection_review_defined_task_card_blocked"
        expected_source = "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1"
    else:
        expected_status = "storage_adapter_runtime_entry_refresh_defined_task_card_blocked"
        expected_source = "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1"
    require(
        durable.get("status") == expected_status,
        "durable backend blocker status drifted",
    )
    require(durable.get("source") == expected_source, "durable backend blocker source drifted")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable backend must block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable backend must block resolver runtime")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    alignment = fixture.get("implementation_readiness_alignment") or {}
    followup_exists = followup_materialization_exists()
    after_selection_exists = followup_after_selection_exists()
    for field, expected in alignment.items():
        if field == "status":
            continue
        if followup_exists and field in FOLLOWUP_ALIGNMENT:
            expected = FOLLOWUP_ALIGNMENT[field]
        if followup_selection_exists() and field in FOLLOWUP_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_SELECTION_ALIGNMENT[field]
        if after_selection_exists and field in FOLLOWUP_AFTER_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_AFTER_SELECTION_ALIGNMENT[field]
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    planned = {str(row.get("id")): row for row in readiness.get("planned_slices") or [] if isinstance(row, dict)}
    item = planned.get("audit-store-storage-adapter-negative-leakage-scan-evidence-readiness") or {}
    require(
        item.get("status") == "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        "implementation readiness missing negative leakage planned slice",
    )
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        (
            "docs/platform/"
            "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.md"
        ): [
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
            "negative_leakage_scan_evidence_defined_without_runtime",
            "storage_adapter_rollback_recovery_evidence_readiness",
        ],
        (
            "docs/task-cards/"
            "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1-plan.md"
        ): [
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
            "metadata_only_negative_leakage_scan_manifest_reference_defined",
            "停止线",
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
            "negative_leakage_scan_evidence_readiness_defined_task_card_blocked",
        ],
        "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": [
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
            "storage_adapter_rollback_recovery_evidence_readiness",
        ],
        "docs/platform/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Negative Leakage Scan Evidence Readiness v1",
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Negative Leakage Scan Evidence Readiness v1",
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        ],
        "docs/features/workflow/README.md": [
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        ],
        "docs/features/workflow/saved-workflow-draft-v1.md": [
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
            "storage_adapter_rollback_recovery_evidence_readiness",
        ],
        "docs/radishmind-current-focus.md": [
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
            "storage_adapter_rollback_recovery_evidence_readiness",
        ],
        "docs/task-cards/README.md": [
            "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1",
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        ],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
            "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1",
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        ],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.py",
            "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.json",
        ],
        "docs/devlogs/2026-W27.md": [
            "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
        ],
    }
    for path, literals in docs.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = "check-production-ops-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.py"
    current = "check-production-ops-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.py"
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in {previous, current, matrix}:
        require(script in check_repo, f"check-repo.py missing {script}")
    require(check_repo.index(previous) < check_repo.index(current) < check_repo.index(matrix), "check order drifted")


def assert_no_secret_literals() -> None:
    text = "\n".join(
        read(path)
        for path in EXPECTED_ALLOWED_ARTIFACTS
        if path.endswith(".md") or path.endswith(".json")
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "-----BEGIN", "authorization:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"negative leakage readiness contains forbidden literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "dsn-like credential found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_readiness_boundary(fixture)
    assert_negative_leakage_evidence(fixture)
    assert_diagnostics_failures_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment()
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store storage adapter negative leakage scan evidence readiness checks passed.")


if __name__ == "__main__":
    main()
