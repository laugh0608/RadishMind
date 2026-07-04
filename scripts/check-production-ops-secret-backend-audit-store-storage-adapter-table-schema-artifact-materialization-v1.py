#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.json"
)
ENTRY_REVIEW_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1"
SLICE_STATUS = "audit_store_storage_adapter_table_schema_artifact_materialization_task_card_defined"
TASK_CARD_DECISION = "table_schema_artifact_materialization_task_card_defined_after_entry_review"
NEXT_DEPENDENCY = "storage_adapter_table_schema_artifact_materialization"
RESERVED_TABLE_SCHEMA_ARTIFACT = "contracts/production-secret-audit-storage-adapter.table-schema.json"
TABLE_SCHEMA_VERSION = "audit-storage-adapter-table-schema-v1"
ENTRY_REVIEW_STATUS = "audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_defined"
RUNTIME_TASK_CARD_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_table_schema_artifact_materialization_task_card"
)
MATRIX_BLOCKER_STATUS = "storage_adapter_table_schema_artifact_materialization_task_card_defined_artifact_blocked"

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1-plan.md",
    (
        "scripts/checks/fixtures/"
        "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.json"
    ),
    (
        "scripts/"
        "check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.py"
    ),
}

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-entry-review-v1.json"
        ),
        ENTRY_REVIEW_STATUS,
    ),
    "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.json"
        ),
        "audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined",
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

EXPECTED_FALSE_FLAGS = {
    "table_schema_artifact_materialized_in_this_slice",
    "table_schema_positive_fixture_created_in_this_slice",
    "table_schema_negative_fixture_created_in_this_slice",
    "table_schema_checker_created_in_this_slice",
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
    "column_name_selected_in_this_slice",
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

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_table_schema_materialization_task_card_missing",
    "audit_store_storage_adapter_table_schema_materialization_scope_missing",
    "audit_store_storage_adapter_table_schema_materialization_field_group_missing",
    "audit_store_storage_adapter_table_schema_materialization_contract_compatibility_missing",
    "audit_store_storage_adapter_table_schema_materialization_validation_plan_missing",
    "audit_store_storage_adapter_table_schema_artifact_created_in_task_card",
    "audit_store_storage_adapter_table_schema_physical_detail_created_in_task_card",
    "audit_store_storage_adapter_table_schema_runtime_created_in_task_card",
    "audit_store_storage_adapter_table_schema_materialization_fallback_forbidden",
    "audit_store_storage_adapter_table_schema_materialization_secret_material_detected",
}

EXPECTED_LOGICAL_GROUPS = {
    "identity",
    "ordering",
    "payload_reference",
    "retention_redaction",
    "delivery_recovery",
    "diagnostics",
}

EXPECTED_VALIDATION = {
    "table schema artifact materialization task card checker",
    "table schema artifact materialization entry review checker",
    "append-only table schema boundary checker",
    "metadata contract artifact materialization checker",
    "positive metadata-only table schema fixture",
    "physical detail negative fixture",
    "secret material negative fixture",
    "additionalProperties negative fixture",
    "metadata contract compatibility smoke",
    "no secret material scan",
    "fast repository check",
}

DOC_REQUIREMENTS = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.md": {
        SLICE_ID,
        SLICE_STATUS,
        TASK_CARD_DECISION,
        NEXT_DEPENDENCY,
        RESERVED_TABLE_SCHEMA_ARTIFACT,
        TABLE_SCHEMA_VERSION,
    },
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1-plan.md": {
        SLICE_ID,
        SLICE_STATUS,
        NEXT_DEPENDENCY,
        RESERVED_TABLE_SCHEMA_ARTIFACT,
    },
    "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": {
        SLICE_STATUS,
        TASK_CARD_DECISION,
        NEXT_DEPENDENCY,
    },
    "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": {
        SLICE_STATUS,
        MATRIX_BLOCKER_STATUS,
        NEXT_DEPENDENCY,
    },
    "docs/radishmind-current-focus.md": {SLICE_STATUS, NEXT_DEPENDENCY},
    "docs/radishmind-integration-contracts.md": {SLICE_STATUS, NEXT_DEPENDENCY},
    "docs/radishmind-architecture.md": {SLICE_STATUS, NEXT_DEPENDENCY},
    "docs/radishmind-product-scope.md": {SLICE_STATUS, NEXT_DEPENDENCY},
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
        == "production_ops_secret_backend_audit_store_storage_adapter_table_schema_artifact_materialization_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected status")
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "table_schema_artifact_materialized",
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


def assert_task_card_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("task_card_boundary") or {}
    expected = {
        "status": SLICE_STATUS,
        "task_card_status": "created_static_task_card",
        "task_card_decision": TASK_CARD_DECISION,
        "entry_review_consumed": True,
        "append_only_table_schema_boundary_consumed": True,
        "metadata_contract_artifact_consumed": True,
        "backend_product_selection_consumed": True,
        "current_development_mode": "table_schema_artifact_materialization_task_card_only_no_artifact",
        "future_table_schema_artifact_path": RESERVED_TABLE_SCHEMA_ARTIFACT,
        "future_table_schema_artifact_version": TABLE_SCHEMA_VERSION,
        "artifact_source": "append_only_boundary_and_metadata_contract_only",
        "runtime_task_card_decision": RUNTIME_TASK_CARD_DECISION,
        "next_dependency": NEXT_DEPENDENCY,
    }
    for field, value in expected.items():
        require(boundary.get(field) == value, f"task_card_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"task_card_boundary.{field} must stay false")


def assert_future_artifact_requirements(fixture: dict[str, Any]) -> None:
    requirements = fixture.get("future_artifact_requirements") or {}
    require(requirements.get("artifact_path") == RESERVED_TABLE_SCHEMA_ARTIFACT, "artifact path drifted")
    require(requirements.get("schema_version_pin") == TABLE_SCHEMA_VERSION, "schema version drifted")
    require(
        requirements.get("logical_field_group_source")
        == "audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined",
        "logical group source drifted",
    )
    require(
        requirements.get("metadata_contract_compatibility_source")
        == "audit_store_storage_adapter_metadata_contract_artifact_materialized",
        "contract compatibility source drifted",
    )
    require(set(requirements.get("allowed_logical_field_groups") or []) == EXPECTED_LOGICAL_GROUPS, "groups drifted")
    require(set(requirements.get("required_validation") or []) == EXPECTED_VALIDATION, "validation plan drifted")
    forbidden = set(requirements.get("forbidden_physical_details") or [])
    for detail in {"database_name", "table_name", "column_name", "column_type", "ddl", "sql_migration"}:
        require(detail in forbidden, f"forbidden physical detail missing {detail}")
    must_not = set(requirements.get("must_not_include") or [])
    for item in {"DB provider", "SQL", "DDL", "schema marker runtime", "storage adapter runtime", "audit store runtime"}:
        require(item in must_not, f"must_not_include missing {item}")


def assert_remaining_blockers(fixture: dict[str, Any]) -> None:
    blockers = rows_by_id(fixture, "remaining_blockers", "id")
    require(blockers["table_schema_artifact_materialization"].get("status") == "not_created", "artifact blocker drifted")
    require(
        blockers["table_schema_artifact_materialization"].get("next_dependency") == NEXT_DEPENDENCY,
        "artifact next dependency drifted",
    )
    require(blockers["schema_marker_runtime"].get("status") == "not_created", "schema marker blocker drifted")
    require(blockers["storage_adapter_runtime"].get("status") == "blocked", "runtime blocker drifted")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "failure_mapping", "code")
    require(set(rows) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in rows.items():
        require(item.get("failure_boundary"), f"{code} missing failure boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} missing diagnostic")
        require("value" not in diagnostic.lower(), f"{code} diagnostic must stay sanitized")


def assert_diagnostics_and_artifacts(fixture: dict[str, Any]) -> None:
    envelope = fixture.get("diagnostic_envelope") or {}
    allowed = set(envelope.get("allowed_fields") or [])
    for field in {
        "audit_store_storage_adapter_table_schema_artifact_materialization_task_card_status",
        "table_schema_artifact_path_status",
        "metadata_contract_compatibility_status",
        "table_schema_artifact_status",
        "storage_adapter_runtime_status",
    }:
        require(field in allowed, f"diagnostic allowed field missing {field}")
    forbidden = set(envelope.get("forbidden_fields") or [])
    for field in {"raw_secret", "dsn", "table_name", "column_name", "column_type", "sql", "ddl", "payload_hash"}:
        require(field in forbidden, f"diagnostic forbidden field missing {field}")

    artifact_guard = fixture.get("artifact_guard") or {}
    require(
        set(artifact_guard.get("allowed_new_artifacts") or []) == EXPECTED_ALLOWED_ARTIFACTS,
        "allowed new artifacts drifted",
    )
    for artifact in EXPECTED_ALLOWED_ARTIFACTS:
        require((REPO_ROOT / artifact).exists(), f"allowed artifact missing on disk: {artifact}")
    for forbidden_path in artifact_guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / forbidden_path).exists(), f"forbidden artifact exists: {forbidden_path}")
    forbidden_artifacts = set(artifact_guard.get("forbidden_artifacts") or [])
    for artifact in {"table_schema_artifact", "sql_migration", "ddl", "schema_marker_runtime", "storage_adapter_runtime"}:
        require(artifact in forbidden_artifacts, f"forbidden artifact missing {artifact}")


def assert_entry_review_still_static() -> None:
    entry = load_json(ENTRY_REVIEW_FIXTURE_PATH)
    boundary = entry.get("entry_review_boundary") or {}
    require(boundary.get("status") == ENTRY_REVIEW_STATUS, "entry review status drifted")
    require(boundary.get("table_schema_artifact_status") == "not_created", "entry review artifact drifted")
    require(boundary.get("storage_adapter_runtime_status") == "not_created", "entry review runtime drifted")


def assert_blocker_matrix_alignment(fixture: dict[str, Any]) -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    alignment = fixture.get("blocker_matrix_alignment") or {}
    boundary = matrix.get("matrix_boundary") or {}
    require(boundary.get("durable_audit_backend_status") == MATRIX_BLOCKER_STATUS, "matrix durable blocker drifted")
    require(boundary.get("storage_adapter_current_next_dependency") == NEXT_DEPENDENCY, "matrix next dependency drifted")
    require(
        boundary.get("storage_adapter_table_schema_artifact_materialization_task_card_status") == "created",
        "matrix task card status drifted",
    )
    require(
        boundary.get("storage_adapter_table_schema_artifact_materialization_task_card_defined_status") == SLICE_STATUS,
        "matrix task card defined status drifted",
    )
    require(
        boundary.get("storage_adapter_runtime_task_card_decision") == RUNTIME_TASK_CARD_DECISION,
        "matrix runtime task decision drifted",
    )
    require(boundary.get("storage_adapter_table_schema_artifact_status") == "not_created", "matrix artifact drifted")
    require(boundary.get("storage_adapter_sql_migration_status") == "not_created", "matrix SQL drifted")
    require(
        alignment.get("durable_backend_blocker_status_after_table_schema_task_card") == MATRIX_BLOCKER_STATUS,
        "fixture blocker status drifted",
    )
    require(
        alignment.get("durable_backend_source_after_table_schema_task_card") == SLICE_ID,
        "fixture blocker source drifted",
    )
    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == MATRIX_BLOCKER_STATUS, "durable blocker row status drifted")
    require(durable.get("source") == SLICE_ID, "durable blocker row source drifted")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in (fixture.get("implementation_readiness_alignment") or {}).items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")
    require(target.get("audit_storage_adapter_table_schema_artifact_status") == "not_created", "artifact drifted")
    require(target.get("audit_storage_adapter_sql_migration_status") == "not_created", "SQL drifted")
    require(target.get("audit_storage_adapter_runtime_status") == "not_created", "runtime drifted")
    planned = readiness.get("planned_slices") or []
    statuses = {str(item.get("status") or "") for item in planned if isinstance(item, dict)}
    require(SLICE_STATUS in statuses, "implementation readiness planned slices missing task card status")


def assert_docs() -> None:
    for relative_path, literals in DOC_REQUIREMENTS.items():
        text = read(relative_path)
        missing = sorted(literal for literal in literals if literal not in text)
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_check_repo_order() -> None:
    text = CHECK_REPO_PATH.read_text(encoding="utf-8")
    entry = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "table-schema-artifact-materialization-entry-review-v1.py"
    )
    current = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "table-schema-artifact-materialization-v1.py"
    )
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in (entry, current, matrix):
        require(script in text, f"check-repo.py missing {script}")
    require(text.index(entry) < text.index(current) < text.index(matrix), "check-repo.py order drifted")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_task_card_boundary(fixture)
    assert_future_artifact_requirements(fixture)
    assert_remaining_blockers(fixture)
    assert_failure_mapping(fixture)
    assert_diagnostics_and_artifacts(fixture)
    assert_entry_review_still_static()
    assert_blocker_matrix_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs()
    assert_check_repo_order()
    print(
        "production ops secret backend audit store storage adapter table schema artifact "
        "materialization task card checks passed."
    )


if __name__ == "__main__":
    main()
