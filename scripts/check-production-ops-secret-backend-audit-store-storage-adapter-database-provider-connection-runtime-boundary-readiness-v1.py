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
    "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1"
SLICE_STATUS = "audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_defined"
READINESS_DECISION = "database_provider_connection_runtime_boundary_readiness_defined_without_connection_runtime"
ENTRY_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_database_provider_connection_runtime_boundary_readiness"
NEXT_DEPENDENCY = "storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_readiness"
MATRIX_BLOCKER_STATUS = "storage_adapter_database_provider_connection_runtime_boundary_readiness_defined_task_card_blocked"
CURRENT_ENTRY_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_managed_database_product_selection_readiness"
)
CURRENT_NEXT_DEPENDENCY = "storage_adapter_managed_database_product_selection_review"
CURRENT_MATRIX_BLOCKER_STATUS = (
    "storage_adapter_managed_database_product_selection_readiness_defined_task_card_blocked"
)
CURRENT_MATRIX_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-managed-database-product-selection-readiness-v1"
)
SELECTED_DRIVER_CANDIDATE = "github.com/jackc/pgx/v5"
SELECTED_ENGINE = "postgresql_compatible_append_only_relational_database"
SELECTED_PROVIDER_CLASS = "managed_postgresql_compatible_service"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
            "after-database-connection-lifecycle-v1.json"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.json"
        ),
        "audit_store_storage_adapter_database_connection_lifecycle_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_database_driver_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_database_provider_selection_review_defined",
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
    "previous_entry_refresh_status": (
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_defined"
    ),
    "previous_entry_decision": "storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_entry_refresh",
    "previous_next_dependency": "storage_adapter_database_provider_connection_runtime_boundary_readiness",
    "selected_database_engine": SELECTED_ENGINE,
    "selected_provider_candidate_class": SELECTED_PROVIDER_CLASS,
    "selected_database_driver_candidate": SELECTED_DRIVER_CANDIDATE,
    "database_driver_status": "selected_reference_only",
    "database_driver_import_status": "not_created",
    "driver_dependency_version_status": "not_pinned",
    "database_provider_connection_runtime_boundary_status": "metadata_only_boundary_defined_without_runtime",
    "database_provider_connection_runtime_status": "not_created",
    "concrete_database_provider_status": "not_created",
    "managed_product_status": "not_selected",
    "vendor_status": "not_selected",
    "connection_provider_boundary_status": "metadata_only_connection_provider_boundary_defined",
    "connection_factory_boundary_status": "metadata_only_connection_factory_boundary_defined",
    "pool_boundary_status": "metadata_only_pool_boundary_defined",
    "health_check_boundary_status": "metadata_only_health_check_boundary_defined",
    "failure_ownership_boundary_status": "metadata_only_failure_ownership_boundary_defined",
    "schema_marker_migration_handoff_status": "schema_marker_migration_handoff_defined_without_runtime",
    "sanitized_diagnostics_status": "sanitized_diagnostics_allowlist_defined",
    "secret_ref_input_status": "secret_ref_only_input_required",
    "provider_profile_input_status": "provider_profile_ref_required_without_provider_detail",
    "environment_binding_status": "environment_class_required",
    "role_policy_input_status": "least_privilege_role_policy_required",
    "next_dependency": NEXT_DEPENDENCY,
    "database_connection_provider_status": "not_created",
    "database_connection_factory_status": "not_created",
    "database_pool_runtime_status": "not_created",
    "database_health_check_runtime_status": "not_created",
    "database_dsn_parser_status": "not_created",
    "database_connection_status": "not_created",
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
    "real_dsn_parsed_in_this_slice",
    "raw_dsn_material_allowed_in_this_slice",
    "database_provider_created_in_this_slice",
    "database_connection_provider_created_in_this_slice",
    "database_connection_factory_created_in_this_slice",
    "database_pool_runtime_created_in_this_slice",
    "database_health_check_runtime_created_in_this_slice",
    "sql_created_in_this_slice",
    "ddl_created_in_this_slice",
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

EXPECTED_BOUNDARY_IDS = {
    "connection_provider_input",
    "connection_factory_fail_closed",
    "pool_key_dimensions",
    "health_check_output",
    "failure_ownership",
}

EXPECTED_FAILURE_OWNERS = {
    "secret_resolver_runtime",
    "database_provider_connection_runtime",
    "schema_marker_migration_runtime",
    "storage_adapter_runtime",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_database_provider_connection_runtime_boundary_dependency_missing",
    "audit_store_storage_adapter_database_provider_connection_runtime_boundary_task_card_still_blocked",
    "audit_store_storage_adapter_database_provider_connection_runtime_boundary_runtime_forbidden",
    "audit_store_storage_adapter_database_provider_connection_runtime_boundary_secret_material_detected",
    "audit_store_storage_adapter_database_provider_connection_runtime_boundary_scope_overreach",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1-plan.md",
    (
        "scripts/checks/fixtures/"
        "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1.json"
    ),
    (
        "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-"
        "database-provider-connection-runtime-boundary-readiness-v1.py"
    ),
}

SECRET_LITERAL_PATTERNS = [
    re.compile(r"Bearer\s+[A-Za-z0-9._-]+"),
    re.compile(r"-----BEGIN [A-Z ]+-----"),
    re.compile(r"AKIA[0-9A-Z]{16}"),
    re.compile(r"sk-[A-Za-z0-9]{20,}"),
    re.compile(r"://[^\s:/]+:[^\s@]+@"),
    re.compile(r"\b(endpoint|host|database_name)\s*[:=]\s*[A-Za-z0-9._-]+", re.IGNORECASE),
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


def check_dependencies(fixture: dict[str, Any]) -> None:
    dependency_rows = rows_by_id(fixture, "depends_on", "id")
    require(set(dependency_rows) == set(EXPECTED_DEPENDENCIES), "depends_on set drifted")
    for dependency_id, (evidence, expected_status) in EXPECTED_DEPENDENCIES.items():
        row = dependency_rows[dependency_id]
        require(row.get("evidence") == evidence, f"{dependency_id} evidence path drifted")
        require(row.get("status") == expected_status, f"{dependency_id} status drifted")
        source = load_json(REPO_ROOT / evidence)
        require(source_status(source) == expected_status, f"{dependency_id} source fixture status drifted")


def check_readiness_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("readiness_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"readiness_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"readiness_boundary.{field} must remain false")


def check_runtime_boundary_matrix(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "runtime_boundary_matrix", "boundary_id")
    require(set(rows) == EXPECTED_BOUNDARY_IDS, "runtime boundary id set drifted")
    for boundary_id, row in rows.items():
        require(row.get("runtime_created_now") is False, f"{boundary_id} must not create runtime")
        require(str(row.get("status") or "").startswith("metadata_only_"), f"{boundary_id} must remain metadata-only")

    owner_rows = rows_by_id(fixture, "failure_ownership_matrix", "owner")
    require(set(owner_rows) == EXPECTED_FAILURE_OWNERS, "failure owner set drifted")
    for owner, row in owner_rows.items():
        failures = row.get("owned_failures")
        require(isinstance(failures, list) and failures, f"{owner} must list owned failures")


def check_diagnostics(fixture: dict[str, Any]) -> None:
    envelope = fixture.get("diagnostic_envelope") or {}
    allowed_fields = set(envelope.get("allowed_fields") or [])
    forbidden_fields = set(envelope.get("forbidden_fields") or [])
    sample = envelope.get("sample") or {}
    require("sanitized_diagnostic" in allowed_fields, "diagnostic envelope must allow sanitized diagnostics")
    require("dsn" in forbidden_fields, "diagnostic envelope must forbid dsn")
    require("endpoint" in forbidden_fields, "diagnostic envelope must forbid endpoint")
    require("credential_payload" in forbidden_fields, "diagnostic envelope must forbid credential payload")
    require(set(sample).issubset(allowed_fields), "diagnostic sample contains non-allowed fields")
    require(sample.get("slice_status") == SLICE_STATUS, "diagnostic sample slice status drifted")
    require(sample.get("next_dependency") == NEXT_DEPENDENCY, "diagnostic sample next dependency drifted")

    failure_codes = {str(row.get("failure_code") or "") for row in fixture.get("failure_mapping") or []}
    require(failure_codes == EXPECTED_FAILURE_CODES, "failure mapping code set drifted")


def check_alignments(fixture: dict[str, Any]) -> None:
    blocker_alignment = fixture.get("blocker_matrix_alignment") or {}
    require(blocker_alignment.get("blocker_id") == "durable_audit_backend", "blocker alignment id drifted")
    require(blocker_alignment.get("status") == CURRENT_MATRIX_BLOCKER_STATUS, "blocker alignment status drifted")
    require(blocker_alignment.get("source") == CURRENT_MATRIX_BLOCKER_SOURCE, "blocker alignment source drifted")
    require(blocker_alignment.get("unlock_condition") == CURRENT_NEXT_DEPENDENCY, "blocker alignment unlock drifted")

    readiness_alignment = fixture.get("implementation_readiness_alignment") or {}
    require(readiness_alignment.get("status") == SLICE_STATUS, "implementation readiness status drifted")
    require(
        readiness_alignment.get("runtime_task_card_decision") == CURRENT_ENTRY_DECISION,
        "implementation readiness decision drifted",
    )
    require(
        readiness_alignment.get("current_next_dependency") == CURRENT_NEXT_DEPENDENCY,
        "implementation readiness next dependency drifted",
    )

    allowed = set((fixture.get("artifact_guard") or {}).get("allowed_artifacts") or [])
    require(allowed == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifact set drifted")


def check_aggregates() -> None:
    implementation = load_json(IMPLEMENTATION_READINESS_PATH)
    overview = implementation.get("implementation_target") or {}
    require(
        overview.get("audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_status")
        == SLICE_STATUS,
        "implementation readiness boundary status drifted",
    )
    require(
        overview.get("audit_storage_adapter_runtime_task_card_decision") == CURRENT_ENTRY_DECISION,
        "implementation readiness runtime task card decision drifted",
    )
    require(
        overview.get("audit_storage_adapter_current_next_dependency") == CURRENT_NEXT_DEPENDENCY,
        "implementation readiness current next dependency drifted",
    )
    readiness_items = rows_by_id(implementation, "planned_slices", "id")
    require(
        readiness_items.get("audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness", {}).get(
            "status"
        )
        == SLICE_STATUS,
        "implementation readiness item missing boundary status",
    )

    blocker_matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = blocker_matrix.get("matrix_boundary") or {}
    require(boundary.get("durable_audit_backend_status") == CURRENT_MATRIX_BLOCKER_STATUS, "matrix durable backend status drifted")
    require(boundary.get("storage_adapter_runtime_task_card_decision") == CURRENT_ENTRY_DECISION, "matrix decision drifted")
    require(boundary.get("storage_adapter_current_next_dependency") == CURRENT_NEXT_DEPENDENCY, "matrix next dependency drifted")
    require(
        boundary.get("storage_adapter_database_provider_connection_runtime_boundary_status")
        == "metadata_only_boundary_defined_without_runtime",
        "matrix connection runtime boundary status drifted",
    )
    blockers = rows_by_id(blocker_matrix, "blocker_matrix", "blocker_id")
    durable_backend = blockers.get("durable_audit_backend") or {}
    require(durable_backend.get("status") == CURRENT_MATRIX_BLOCKER_STATUS, "durable backend blocker status drifted")
    require(durable_backend.get("source") == CURRENT_MATRIX_BLOCKER_SOURCE, "durable backend blocker source drifted")
    require(durable_backend.get("unlock_condition") == CURRENT_NEXT_DEPENDENCY, "durable backend unlock condition drifted")


def check_docs_and_registration() -> None:
    platform_doc = read(
        "docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1.md"
    )
    task_card = read(
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1-plan.md"
    )
    scripts_readme = read("scripts/README.md")
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")

    for content, label in ((platform_doc, "platform doc"), (task_card, "task card")):
        require(SLICE_ID in content, f"{label} missing slice id")
        require(SLICE_STATUS in content, f"{label} missing slice status")
        require(READINESS_DECISION in content, f"{label} missing readiness decision")
        require(ENTRY_DECISION in content, f"{label} missing entry decision")
        require(NEXT_DEPENDENCY in content, f"{label} missing next dependency")
        require("不创建" in content, f"{label} must preserve no-runtime language")

    checker_name = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "database-provider-connection-runtime-boundary-readiness-v1.py"
    )
    require(checker_name in scripts_readme, "scripts README missing checker description")
    require(checker_name in check_repo, "check-repo.py missing checker registration")


def check_no_secret_material(fixture: dict[str, Any]) -> None:
    paths = (fixture.get("no_secret_material_scan") or {}).get("paths") or []
    require(paths, "no secret material scan paths missing")
    for relative_path in paths:
        content = read(str(relative_path))
        for pattern in SECRET_LITERAL_PATTERNS:
            require(not pattern.search(content), f"secret-like literal found in {relative_path}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "slice id drifted")
    require(slice_info.get("status") == SLICE_STATUS, "slice status drifted")
    require(slice_info.get("readiness_decision") == READINESS_DECISION, "readiness decision drifted")
    require(slice_info.get("entry_decision") == ENTRY_DECISION, "entry decision drifted")
    require(slice_info.get("next_dependency") == NEXT_DEPENDENCY, "next dependency drifted")
    check_dependencies(fixture)
    check_readiness_boundary(fixture)
    check_runtime_boundary_matrix(fixture)
    check_diagnostics(fixture)
    check_alignments(fixture)
    check_aggregates()
    check_docs_and_registration()
    check_no_secret_material(fixture)


if __name__ == "__main__":
    main()
