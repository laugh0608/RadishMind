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
    "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.json"
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
FOLLOWUP_ALIGNMENT = {
    "audit_storage_adapter_metadata_contract_artifact_status": "materialized_static_metadata_contract",
    "audit_storage_adapter_contract_artifact_path_status": "materialized_static_path",
    "audit_storage_adapter_contract_artifact_materialization_status": FOLLOWUP_MATERIALIZATION_STATUS,
}
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
FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY = "storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness"
FOLLOWUP_AFTER_SELECTION_BLOCKER_STATUS = (
    "storage_adapter_runtime_entry_refresh_after_product_selection_defined_task_card_blocked"
)
FOLLOWUP_AFTER_SELECTION_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1"
)
FOLLOWUP_SELECTION_ALIGNMENT = {
    "audit_store_storage_adapter_backend_product_selection_review_status": FOLLOWUP_SELECTION_STATUS,
    "audit_storage_adapter_backend_product_selection_status": "selected_static_product_class_without_backend_provider",
    "audit_storage_adapter_selected_backend_product_class": "managed_database_append_only_table",
    "audit_storage_adapter_selected_backend_product_profile": "reserved_managed_database_append_only_table_profile",
    "audit_storage_adapter_database_product_status": "not_selected",
    "audit_storage_adapter_database_connection_provider_status": "blocked",
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_SELECTION_NEXT_DEPENDENCY,
}
FOLLOWUP_AFTER_SELECTION_ALIGNMENT = {
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status": (
        FOLLOWUP_AFTER_SELECTION_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_product_selection"
    ),
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY,
    "audit_storage_adapter_database_provider_driver_dsn_tls_role_policy_status": (
        "required_before_runtime_task_card"
    ),
    "audit_storage_adapter_append_only_table_schema_boundary_status": "required_before_runtime_task_card",
    "audit_storage_adapter_migration_schema_marker_boundary_status": "required_before_runtime_task_card",
    "audit_storage_adapter_offline_adapter_smoke_strategy_status": "required_before_runtime_task_card",
    "audit_storage_adapter_negative_leakage_runtime_scan_boundary_status": "required_before_runtime_task_card",
}

EXPECTED_DEPENDENCIES = {
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
    "production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.json"
        ),
        "audit_store_concrete_durable_backend_selection_review_defined",
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
    "status": "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
    "readiness_decision": "append_only_semantics_evidence_defined_without_runtime",
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
    "append_only_semantics_evidence_status": "append_only_semantics_evidence_defined",
    "append_only_evidence_status": "defined_without_runtime",
    "append_only_operation_status": "append_only_insert_only",
    "forbidden_mutation_policy_status": "update_delete_overwrite_truncate_reject_policy_defined",
    "append_only_sequence_reference_status": "metadata_only_monotonic_sequence_reference_defined",
    "record_immutability_status": "metadata_only_immutability_policy_defined",
    "duplicate_replay_policy_status": "fail_closed_duplicate_replay_reference_defined",
    "writer_append_compatibility_status": "metadata_only_writer_append_compatibility_defined",
    "retention_redaction_status": "required_before_runtime_task_card",
    "offline_validation_status": "not_created",
    "negative_leakage_scan_status": "not_created",
    "rollback_recovery_status": "required_before_runtime_task_card",
    "next_dependency": "storage_adapter_retention_redaction_policy_evidence_readiness",
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "storage_adapter_client_status": "not_created",
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
    "append_only_runtime_created_in_this_slice",
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

EXPECTED_ALLOWED_OPERATIONS = {
    "append_audit_record",
    "append_delivery_attempt_record",
    "append_idempotency_reference_record",
}
EXPECTED_FORBIDDEN_OPERATIONS = {
    "update_record",
    "delete_record",
    "overwrite_record",
    "truncate_log",
    "compact_log",
    "rewrite_sequence",
    "mutate_record_identity",
    "replace_delivery_result",
    "erase_for_retention",
    "inline_redact_payload",
}
EXPECTED_SEQUENCE_FIELDS = {
    "append_only_sequence_ref",
    "append_only_contract_ref",
    "storage_record_identity_ref",
    "writer_result_ref",
    "idempotency_key_ref",
    "delivery_attempt_ref",
    "policy_version",
}
EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_append_only_dependency_missing",
    "audit_store_storage_adapter_append_only_mutation_forbidden",
    "audit_store_storage_adapter_append_only_sequence_reference_missing",
    "audit_store_storage_adapter_append_only_identity_mutation_detected",
    "audit_store_storage_adapter_append_only_duplicate_replay_mutation_detected",
    "audit_store_storage_adapter_append_only_retention_redaction_mutation_claim",
    "audit_store_storage_adapter_append_only_runtime_scope_overreach",
    "audit_store_storage_adapter_append_only_secret_material_detected",
}
EXPECTED_ALLOWED_ARTIFACTS = {
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
    "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.py",
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
        == "production_ops_secret_backend_audit_store_storage_adapter_append_only_semantics_evidence_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id")
        == "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        "unexpected status",
    )
    require(
        slice_info.get("readiness_decision") == "append_only_semantics_evidence_defined_without_runtime",
        "unexpected readiness decision",
    )
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "append_only_runtime_implemented",
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


def assert_append_only_semantics(fixture: dict[str, Any]) -> None:
    semantics = fixture.get("operation_semantics") or {}
    require(semantics.get("status") == "append_only_operations_only", "operation semantics status drifted")
    require(
        set(semantics.get("allowed_success_operations") or []) == EXPECTED_ALLOWED_OPERATIONS,
        "allowed append operations drifted",
    )
    require(
        set(semantics.get("forbidden_mutation_operations") or []) == EXPECTED_FORBIDDEN_OPERATIONS,
        "forbidden mutation operations drifted",
    )
    require(semantics.get("mutation_result_policy") == "fail_closed_never_success", "mutation policy drifted")
    for forbidden_runtime in {"runtime_adapter", "database_sequence", "log_offset_reader"}:
        require(forbidden_runtime in set(semantics.get("does_not_create") or []), f"creates {forbidden_runtime}")

    sequence = fixture.get("sequence_identity_contract") or {}
    require(
        sequence.get("status") == "metadata_only_sequence_and_identity_defined",
        "sequence identity status drifted",
    )
    require(
        set(sequence.get("required_reference_fields") or []) == EXPECTED_SEQUENCE_FIELDS,
        "sequence reference fields drifted",
    )
    for field in {"physical_primary_key", "table_name", "bucket_key", "queue_offset", "topic_partition", "dsn"}:
        require(field in set(sequence.get("forbidden_physical_fields") or []), f"missing forbidden field {field}")

    duplicate = fixture.get("duplicate_replay_policy") or {}
    require(
        duplicate.get("status") == "fail_closed_duplicate_replay_reference_defined",
        "duplicate replay policy drifted",
    )
    for mode in {"update_existing_record", "overwrite_existing_record", "delete_existing_record"}:
        require(mode in set(duplicate.get("forbidden_resolution_modes") or []), f"missing forbidden mode {mode}")

    writer = fixture.get("writer_append_compatibility_contract") or {}
    require(
        writer.get("status") == "metadata_only_writer_append_compatibility_defined",
        "writer append compatibility status drifted",
    )
    require(
        {"writer_result_ref", "audit_event_ref", "append_only_contract_ref", "idempotency_key_ref"}
        <= set(writer.get("writer_output_required_fields") or []),
        "writer required fields drifted",
    )
    for claim in {"writer_updates_existing_record", "writer_deletes_existing_record", "writer_overwrites_existing_record"}:
        require(claim in set(writer.get("forbidden_writer_claims") or []), f"missing forbidden writer claim {claim}")


def assert_diagnostics_failures_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(not (allowed & forbidden), "diagnostic allowlist intersects forbidden fields")
    require(sample.get("storage_adapter_runtime_status") == "not_created", "diagnostic sample created runtime")
    require(
        sample.get("retention_redaction_status") == "required_before_runtime_task_card",
        "diagnostic sample unlocked retention/redaction",
    )

    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} boundary missing")
        require(item.get("sanitized_diagnostic"), f"{code} diagnostic missing")

    no_fallback = fixture.get("no_fallback_policy") or {}
    require(no_fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for source in {
        "metadata_contract_artifact_readiness",
        "backend_product_evidence_readiness",
        "storage_adapter_runtime_entry_review",
        "static_delivery_idempotency_readiness",
        "historical_smoke",
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
        boundary.get("storage_adapter_append_only_semantics_evidence_readiness_status")
        == "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        "matrix boundary missing append-only semantics readiness status",
    )
    require(
        boundary.get("storage_adapter_append_only_semantics_status")
        == "append_only_semantics_evidence_defined_without_runtime",
        "matrix boundary append-only semantics status drifted",
    )
    require(
        boundary.get("storage_adapter_retention_redaction_policy_evidence_readiness_status")
        == "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
        "matrix boundary missing retention/redaction policy evidence readiness status",
    )
    require(
        boundary.get("storage_adapter_retention_redaction_status")
        == "retention_redaction_policy_evidence_defined_without_runtime",
        "matrix boundary retention/redaction status drifted",
    )
    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    if followup_after_selection_exists():
        expected_blocker_status = FOLLOWUP_AFTER_SELECTION_BLOCKER_STATUS
        expected_source = FOLLOWUP_AFTER_SELECTION_SOURCE
    elif followup_selection_exists():
        expected_blocker_status = "storage_adapter_backend_product_selection_review_defined_task_card_blocked"
        expected_source = "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1"
    else:
        expected_blocker_status = "storage_adapter_runtime_entry_refresh_defined_task_card_blocked"
        expected_source = "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1"
    require(
        durable.get("status") == expected_blocker_status,
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
    selection_exists = followup_selection_exists()
    after_selection_exists = followup_after_selection_exists()
    for field, expected in alignment.items():
        if field == "status":
            continue
        if selection_exists and field in FOLLOWUP_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_SELECTION_ALIGNMENT[field]
        if followup_exists and field in FOLLOWUP_ALIGNMENT:
            expected = FOLLOWUP_ALIGNMENT[field]
        if after_selection_exists and field in FOLLOWUP_AFTER_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_AFTER_SELECTION_ALIGNMENT[field]
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    planned = {str(row.get("id")): row for row in readiness.get("planned_slices") or [] if isinstance(row, dict)}
    item = planned.get("audit-store-storage-adapter-append-only-semantics-evidence-readiness") or {}
    require(
        item.get("status") == "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        "implementation readiness missing append-only semantics planned slice",
    )
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        (
            "docs/platform/"
            "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.md"
        ): [
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
            "append_only_semantics_evidence_defined_without_runtime",
            "storage_adapter_retention_redaction_policy_evidence_readiness",
        ],
        (
            "docs/task-cards/"
            "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1-plan.md"
        ): [
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
            "append_only_insert_only",
            "停止线",
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
            "offline_validation_evidence_readiness_defined_task_card_blocked",
        ],
        "docs/platform/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Append-Only Semantics Evidence Readiness v1",
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Append-Only Semantics Evidence Readiness v1",
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        ],
        "docs/features/workflow/README.md": [
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        ],
        "docs/features/workflow/saved-workflow-draft-v1.md": [
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
            "storage_adapter_retention_redaction_policy_evidence_readiness",
        ],
        "docs/radishmind-current-focus.md": [
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
            "storage_adapter_retention_redaction_policy_evidence_readiness",
        ],
        "docs/task-cards/README.md": [
            "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1",
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        ],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.py",
            "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.json",
        ],
        "docs/devlogs/2026-W27.md": [
            "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
        ],
    }
    for path, literals in docs.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.py"
    current = "check-production-ops-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.py"
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
    require(not found, f"append-only semantics readiness contains forbidden literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "dsn-like credential found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_readiness_boundary(fixture)
    assert_append_only_semantics(fixture)
    assert_diagnostics_failures_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment()
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store storage adapter append-only semantics evidence readiness checks passed.")


if __name__ == "__main__":
    main()
