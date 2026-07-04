#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1.json"
)
PREVIOUS_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = (
    "production-secret-backend-audit-store-storage-adapter-"
    "table-schema-artifact-materialization-entry-review-v1"
)
SLICE_STATUS = "audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_defined"
ENTRY_DECISION = "table_schema_artifact_materialization_task_card_ready_after_entry_review"
NEXT_DEPENDENCY = "storage_adapter_table_schema_artifact_materialization_task_card"
RESERVED_TABLE_SCHEMA_ARTIFACT = "contracts/production-secret-audit-storage-adapter.table-schema.json"
DOWNSTREAM_MATERIALIZED_ARTIFACTS = {
    RESERVED_TABLE_SCHEMA_ARTIFACT,
}
PREVIOUS_STATUS = "audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined"
PREVIOUS_DECISION = "append_only_table_schema_boundary_defined_without_sql_or_runtime"
RUNTIME_TASK_CARD_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_table_schema_artifact_materialization_entry_review"
)
MATRIX_BLOCKER_STATUS = "storage_adapter_table_schema_artifact_materialization_entry_review_defined_task_card_blocked"
CURRENT_NEXT_DEPENDENCY = "storage_adapter_negative_leakage_runtime_scan_boundary_readiness"
CURRENT_RUNTIME_TASK_CARD_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_offline_adapter_smoke_strategy_readiness"
)
CURRENT_MATRIX_BLOCKER_STATUS = "storage_adapter_offline_adapter_smoke_strategy_readiness_defined_runtime_blocked"
CURRENT_MATRIX_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1"
)
SELECTED_PRODUCT_CLASS = "managed_database_append_only_table"
SELECTED_PRODUCT_PROFILE = "reserved_managed_database_append_only_table_profile"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.json"
        ),
        PREVIOUS_STATUS,
    ),
    "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.json"
        ),
        "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined",
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
    "entry_decision": ENTRY_DECISION,
    "previous_append_only_table_schema_boundary_status": PREVIOUS_STATUS,
    "previous_append_only_table_schema_boundary_decision": PREVIOUS_DECISION,
    "previous_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_append_only_table_schema_boundary_readiness"
    ),
    "selected_backend_product_class": SELECTED_PRODUCT_CLASS,
    "selected_backend_product_profile": SELECTED_PRODUCT_PROFILE,
    "metadata_contract_artifact_status": "audit_store_storage_adapter_metadata_contract_artifact_materialized",
    "logical_table_schema_status": "logical_append_only_table_schema_boundary_defined",
    "logical_field_group_status": "logical_field_groups_defined_without_physical_columns",
    "record_identity_boundary_status": "logical_record_identity_boundary_defined",
    "sequence_reference_boundary_status": "logical_sequence_reference_boundary_defined",
    "idempotency_reference_boundary_status": "logical_idempotency_reference_boundary_defined",
    "retention_redaction_reference_boundary_status": "logical_retention_redaction_reference_boundary_defined",
    "schema_marker_handoff_boundary_status": "logical_schema_marker_handoff_boundary_defined",
    "reserved_table_schema_artifact_path": RESERVED_TABLE_SCHEMA_ARTIFACT,
    "materialization_task_card_decision": ENTRY_DECISION,
    "materialization_task_card_status": "not_created",
    "table_schema_artifact_materialization_status": "not_created",
    "table_schema_artifact_status": "not_created",
    "positive_fixture_status": "not_created",
    "negative_fixture_status": "not_created",
    "no_secret_material_scan_status": "not_created",
    "database_product_status": "not_selected",
    "database_vendor_status": "not_selected",
    "database_connection_provider_status": "not_created",
    "sql_migration_status": "not_created",
    "ddl_status": "not_created",
    "table_name_status": "not_selected",
    "column_type_status": "not_selected",
    "index_status": "not_created",
    "constraint_status": "not_created",
    "schema_marker_status": "not_created",
    "schema_marker_runtime_status": "not_created",
    "migration_runner_status": "not_created",
    "runtime_task_card_decision": RUNTIME_TASK_CARD_DECISION,
    "next_dependency": NEXT_DEPENDENCY,
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "storage_adapter_client_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_writer_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "delivery_runtime_status": "not_created",
    "production_resolver_runtime_task_card_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "table_schema_artifact_materialization_task_card_created_in_this_slice",
    "table_schema_artifact_materialized_in_this_slice",
    "table_schema_positive_fixture_created_in_this_slice",
    "table_schema_negative_fixture_created_in_this_slice",
    "table_schema_no_secret_material_scan_created_in_this_slice",
    "database_product_selected_in_this_slice",
    "database_vendor_selected_in_this_slice",
    "database_driver_selected_in_this_slice",
    "database_provider_created_in_this_slice",
    "database_connection_provider_enabled",
    "database_connection_created_in_this_slice",
    "sql_created_in_this_slice",
    "ddl_created_in_this_slice",
    "table_name_selected_in_this_slice",
    "column_type_selected_in_this_slice",
    "index_created_in_this_slice",
    "constraint_created_in_this_slice",
    "schema_marker_created_in_this_slice",
    "schema_marker_runtime_created_in_this_slice",
    "migration_runner_created_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_REQUIREMENTS = {
    "metadata-only table schema artifact version pin",
    "logical field group coverage",
    "logical field to metadata contract envelope compatibility",
    "positive table schema fixture",
    "forbidden physical detail negative fixture",
    "no secret material scan",
    "artifact guard for no SQL no database no runtime",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_table_schema_materialization_entry_dependency_missing",
    "audit_store_storage_adapter_table_schema_materialization_task_card_not_created",
    "audit_store_storage_adapter_table_schema_materialization_artifact_forbidden",
    "audit_store_storage_adapter_table_schema_materialization_physical_schema_forbidden",
    "audit_store_storage_adapter_table_schema_materialization_runtime_forbidden",
    "audit_store_storage_adapter_table_schema_materialization_secret_material_detected",
    "audit_store_storage_adapter_table_schema_materialization_fallback_detected",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    (
        "docs/platform/production-secret-backend-audit-store-storage-adapter-"
        "table-schema-artifact-materialization-entry-review-v1.md"
    ),
    (
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-"
        "table-schema-artifact-materialization-entry-review-v1-plan.md"
    ),
    (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-"
        "table-schema-artifact-materialization-entry-review-v1.json"
    ),
    (
        "scripts/"
        "check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1.py"
    ),
}

DOC_REQUIREMENTS = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1.md": {
        SLICE_STATUS,
        ENTRY_DECISION,
        NEXT_DEPENDENCY,
        RESERVED_TABLE_SCHEMA_ARTIFACT,
        RUNTIME_TASK_CARD_DECISION,
    },
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1-plan.md": {
        SLICE_ID,
        SLICE_STATUS,
        ENTRY_DECISION,
        NEXT_DEPENDENCY,
    },
    "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": {
        SLICE_STATUS,
        ENTRY_DECISION,
        NEXT_DEPENDENCY,
    },
    "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": {
        SLICE_STATUS,
        MATRIX_BLOCKER_STATUS,
        NEXT_DEPENDENCY,
    },
    "docs/radishmind-current-focus.md": {SLICE_STATUS, NEXT_DEPENDENCY},
    "docs/radishmind-integration-contracts.md": {SLICE_STATUS, NEXT_DEPENDENCY},
    "docs/features/README.md": {SLICE_STATUS, NEXT_DEPENDENCY},
    "docs/features/workflow/README.md": {SLICE_STATUS, NEXT_DEPENDENCY},
    "docs/features/workflow/saved-workflow-draft-v1.md": {SLICE_STATUS, NEXT_DEPENDENCY},
    "docs/task-cards/README.md": {SLICE_ID, SLICE_STATUS},
    "docs/task-cards/production-secret-backend-audit-store-runtime-blocker-matrix-v1-plan.md": {
        SLICE_STATUS,
        MATRIX_BLOCKER_STATUS,
        NEXT_DEPENDENCY,
    },
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": {SLICE_STATUS, NEXT_DEPENDENCY},
    "scripts/README.md": {SLICE_STATUS},
    "docs/devlogs/2026-W27.md": {SLICE_STATUS, NEXT_DEPENDENCY},
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


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected status")
    require(slice_info.get("entry_decision") == ENTRY_DECISION, "unexpected entry decision")
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "table_schema_artifact_materialized",
        "database_product_selected",
        "sql_created",
        "ddl_created",
        "schema_marker_runtime_created",
        "storage_adapter_runtime_created",
        "audit_store_runtime_created",
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


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_review_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"entry_review_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"entry_review_boundary.{field} must stay false")


def assert_future_requirements(fixture: dict[str, Any]) -> None:
    requirements = set(fixture.get("future_materialization_task_card_requirements") or [])
    require(requirements == EXPECTED_REQUIREMENTS, "future materialization requirements drifted")
    logical = fixture.get("logical_artifact_boundary") or {}
    require(
        logical.get("reserved_table_schema_artifact_path") == RESERVED_TABLE_SCHEMA_ARTIFACT,
        "reserved artifact path drifted",
    )
    for group in {"identity", "ordering", "payload_reference", "retention_redaction", "delivery_recovery"}:
        require(group in set(logical.get("allowed_logical_field_groups") or []), f"logical group missing {group}")
    forbidden = set(logical.get("forbidden_physical_details") or [])
    for detail in {"database_name", "table_name", "column_type", "index_name", "constraint_name", "ddl"}:
        require(detail in forbidden, f"forbidden physical detail missing {detail}")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "failure_mapping", "code")
    require(set(rows) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in rows.items():
        require(item.get("failure_boundary"), f"{code} missing failure_boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} missing sanitized diagnostic")
        require("value" not in diagnostic.lower(), f"{code} diagnostic must stay sanitized")


def assert_diagnostics_and_artifacts(fixture: dict[str, Any]) -> None:
    envelope = fixture.get("diagnostic_envelope") or {}
    allowed = set(envelope.get("allowed_fields") or [])
    for field in {
        "audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_status",
        "reserved_table_schema_artifact_path",
        "runtime_task_card_decision",
        "schema_marker_runtime_status",
        "storage_adapter_runtime_status",
    }:
        require(field in allowed, f"diagnostic allowed field missing {field}")
    forbidden = set(envelope.get("forbidden_fields") or [])
    for field in {"raw_secret", "dsn", "database_hostname", "table_name", "column_type", "sql", "ddl"}:
        require(field in forbidden, f"diagnostic forbidden field missing {field}")

    artifact_guard = fixture.get("artifact_guard") or {}
    require(
        set(artifact_guard.get("allowed_new_artifacts") or []) == EXPECTED_ALLOWED_ARTIFACTS,
        "allowed new artifacts drifted",
    )
    for artifact in EXPECTED_ALLOWED_ARTIFACTS:
        require((REPO_ROOT / artifact).exists(), f"allowed artifact missing on disk: {artifact}")
    for forbidden_path in artifact_guard.get("files_must_not_exist") or []:
        if forbidden_path in DOWNSTREAM_MATERIALIZED_ARTIFACTS:
            continue
        require(not (REPO_ROOT / forbidden_path).exists(), f"forbidden artifact exists: {forbidden_path}")
    for forbidden in {
        "table_schema_artifact",
        "sql_migration",
        "ddl",
        "schema_marker_runtime",
        "migration_runner",
        "storage_adapter_runtime",
        "audit_store_runtime",
    }:
        require(forbidden in set(artifact_guard.get("forbidden_artifacts") or []), f"forbidden artifact missing {forbidden}")


def assert_previous_boundary_still_static() -> None:
    previous = load_json(PREVIOUS_FIXTURE_PATH)
    boundary = previous.get("readiness_boundary") or {}
    require(boundary.get("status") == PREVIOUS_STATUS, "previous readiness status drifted")
    require(boundary.get("readiness_decision") == PREVIOUS_DECISION, "previous readiness decision drifted")
    require(boundary.get("table_schema_artifact_status") == "not_created", "previous table artifact drifted")
    require(boundary.get("sql_migration_status") == "not_created", "previous SQL migration drifted")
    require(boundary.get("storage_adapter_runtime_status") == "not_created", "previous runtime status drifted")


def assert_blocker_matrix_alignment(fixture: dict[str, Any]) -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    alignment = fixture.get("blocker_matrix_alignment") or {}
    boundary = matrix.get("matrix_boundary") or {}
    require(
        boundary.get("durable_audit_backend_status") == CURRENT_MATRIX_BLOCKER_STATUS,
        "matrix durable blocker drifted",
    )
    require(
        boundary.get("storage_adapter_current_next_dependency") == CURRENT_NEXT_DEPENDENCY,
        "matrix next dependency drifted",
    )
    require(boundary.get("storage_adapter_table_schema_artifact_materialization_entry_review_status") == SLICE_STATUS, "matrix entry review missing")
    require(
        boundary.get("storage_adapter_runtime_task_card_decision") == CURRENT_RUNTIME_TASK_CARD_DECISION,
        "matrix runtime task decision drifted",
    )
    require(
        boundary.get("storage_adapter_table_schema_artifact_status") == "materialized_static_logical_table_schema",
        "matrix table schema artifact drifted",
    )
    require(boundary.get("storage_adapter_sql_migration_status") == "not_created", "matrix SQL migration drifted")
    require(
        alignment.get("durable_backend_blocker_status_after_table_schema_task_card") == CURRENT_MATRIX_BLOCKER_STATUS,
        "fixture matrix blocker alignment drifted",
    )
    require(
        alignment.get("durable_backend_source_after_table_schema_task_card") == CURRENT_MATRIX_BLOCKER_SOURCE,
        "fixture matrix blocker source alignment drifted",
    )
    blocker_rows = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blocker_rows.get("durable_audit_backend") or {}
    require(durable.get("status") == CURRENT_MATRIX_BLOCKER_STATUS, "durable blocker row status drifted")
    require(durable.get("source") == CURRENT_MATRIX_BLOCKER_SOURCE, "durable blocker row source drifted")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    alignment = fixture.get("implementation_readiness_alignment") or {}
    target = readiness.get("implementation_target") or {}
    for field, expected in alignment.items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")
    require(target.get("audit_storage_adapter_runtime_status") == "not_created", "storage adapter runtime drifted")
    require(target.get("audit_store_runtime_status") == "not_created", "audit store runtime drifted")
    require(target.get("repository_mode_status") == "disabled", "repository mode drifted")
    require(target.get("production_api_status") == "not_created", "production API drifted")
    planned = readiness.get("planned_slices") or []
    statuses = {str(item.get("status") or "") for item in planned if isinstance(item, dict)}
    require(SLICE_STATUS in statuses, "implementation readiness planned slices missing slice status")


def assert_docs() -> None:
    for relative_path, literals in DOC_REQUIREMENTS.items():
        text = read(relative_path)
        missing = sorted(literal for literal in literals if literal not in text)
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_check_repo_order() -> None:
    text = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "append-only-table-schema-boundary-readiness-v1.py"
    )
    current = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "table-schema-artifact-materialization-entry-review-v1.py"
    )
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in (previous, current, matrix):
        require(script in text, f"check-repo.py missing {script}")
    require(text.index(previous) < text.index(current) < text.index(matrix), "check-repo.py order drifted")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_boundary(fixture)
    assert_future_requirements(fixture)
    assert_failure_mapping(fixture)
    assert_diagnostics_and_artifacts(fixture)
    assert_previous_boundary_still_static()
    assert_blocker_matrix_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs()
    assert_check_repo_order()
    print(
        "production ops secret backend audit store storage adapter table schema artifact "
        "materialization entry review checks passed."
    )


if __name__ == "__main__":
    main()
