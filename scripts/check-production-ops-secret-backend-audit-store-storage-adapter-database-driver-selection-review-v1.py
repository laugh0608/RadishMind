#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
BLOCKER_MATRIX_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1"
SLICE_STATUS = "audit_store_storage_adapter_database_driver_selection_review_defined"
SELECTION_DECISION = "database_driver_candidate_selected_pgx_v5_runtime_blocked"
ENTRY_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_database_driver_selection_review"
NEXT_DEPENDENCY = "storage_adapter_database_connection_lifecycle_readiness"
SELECTED_DATABASE_ENGINE = "postgresql_compatible_append_only_relational_database"
SELECTED_PROVIDER_CLASS = "managed_postgresql_compatible_service"
SELECTED_DRIVER_CANDIDATE = "github.com/jackc/pgx/v5"
DRIVER_SELECTION_STATUS = "selected_driver_candidate_without_runtime_import"
DRIVER_PACKAGE_STATUS = "selected_candidate_reference_only"
DRIVER_IMPORT_STATUS = "not_created"
DATABASE_DRIVER_STATUS = "selected_reference_only"
MATRIX_BLOCKER_STATUS = "storage_adapter_database_driver_selection_review_defined_task_card_blocked"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.json"
        ),
        "audit_store_storage_adapter_database_driver_selection_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_database_provider_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_concrete_database_selection_review_defined",
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

EXPECTED_SELECTION_REVIEW = {
    "status": SLICE_STATUS,
    "selection_decision": SELECTION_DECISION,
    "entry_decision": ENTRY_DECISION,
    "previous_readiness_status": "audit_store_storage_adapter_database_driver_selection_readiness_defined",
    "previous_runtime_task_decision": "storage_adapter_runtime_task_card_still_blocked_after_database_driver_selection_readiness",
    "selected_database_engine": SELECTED_DATABASE_ENGINE,
    "selected_provider_candidate_class": SELECTED_PROVIDER_CLASS,
    "selected_driver_candidate": SELECTED_DRIVER_CANDIDATE,
    "selected_driver_candidate_source": "https://github.com/jackc/pgx",
    "selected_driver_package_source": "https://pkg.go.dev/github.com/jackc/pgx/v5",
    "driver_selection_status": DRIVER_SELECTION_STATUS,
    "driver_candidate_evidence_status": "public_source_reviewed_metadata_only",
    "driver_package_status": DRIVER_PACKAGE_STATUS,
    "driver_dependency_version_status": "not_pinned",
    "driver_import_status": DRIVER_IMPORT_STATUS,
    "database_driver_status": DATABASE_DRIVER_STATUS,
    "database_dsn_status": "not_defined",
    "database_dsn_parser_status": "not_created",
    "database_connection_provider_status": "not_created",
    "database_connection_status": "not_created",
    "next_dependency": NEXT_DEPENDENCY,
    "database_vendor_status": "not_selected",
    "managed_database_product_status": "not_selected",
    "database_provider_status": "provider_class_selected_without_vendor_or_product",
    "provider_account_resource_status": "not_defined",
    "database_endpoint_status": "not_defined",
    "dsn_secret_ref_compatibility_status": "metadata_only_dsn_secret_ref_compatibility_defined",
    "tls_mode_compatibility_status": "metadata_only_tls_mode_compatibility_defined",
    "least_privilege_role_compatibility_status": "metadata_only_role_policy_compatibility_defined",
    "connection_lifecycle_boundary_status": "metadata_only_connection_lifecycle_boundary_defined",
    "migration_schema_marker_boundary_status": "logical_schema_marker_handoff_boundary_defined",
    "offline_adapter_smoke_boundary_status": "metadata_only_offline_adapter_smoke_boundary_defined",
    "negative_leakage_runtime_scan_boundary_status": "metadata_only_negative_leakage_runtime_scan_boundary_defined",
    "rollout_rollback_boundary_status": "metadata_only_rollout_rollback_boundary_defined",
    "sql_migration_status": "not_created",
    "ddl_status": "not_created",
    "physical_table_schema_status": "not_created",
    "schema_marker_runtime_status": "not_created",
    "migration_runner_status": "not_created",
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "database_vendor_selected_in_this_slice",
    "managed_database_product_selected_in_this_slice",
    "database_provider_selected_in_this_slice",
    "provider_account_resource_defined_in_this_slice",
    "database_endpoint_defined_in_this_slice",
    "driver_import_added_in_this_slice",
    "driver_dependency_version_pinned_in_this_slice",
    "go_mod_changed_in_this_slice",
    "database_connection_provider_enabled",
    "database_connection_created_in_this_slice",
    "dsn_defined_in_this_slice",
    "dsn_parser_created_in_this_slice",
    "sql_migration_created_in_this_slice",
    "ddl_created_in_this_slice",
    "physical_table_schema_created_in_this_slice",
    "schema_marker_runtime_created_in_this_slice",
    "migration_runner_created_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_EXTERNAL_EVIDENCE = {
    "pgx-github-official-repository": {
        "source": "https://github.com/jackc/pgx",
        "status": "public_source_reviewed_metadata_only",
        "points": {
            "postgresql_driver_and_toolkit",
            "pure_go_driver",
            "mit_license",
            "database_sql_adapter_available",
            "tls_connection_control",
            "connection_pool_available",
            "stable_v5_policy",
        },
    },
    "pgx-pkg-go-dev-package-metadata": {
        "source": "https://pkg.go.dev/github.com/jackc/pgx/v5",
        "status": "package_metadata_reviewed",
        "points": {
            "module_path_reviewed",
            "tagged_module_reviewed",
            "license_metadata_reviewed",
            "public_documentation_reviewed",
        },
    },
}

EXPECTED_CANDIDATES = {
    "github.com/jackc/pgx/v5": ("selected_driver_candidate_without_runtime_import", True),
    "database_sql_only_abstraction_without_selected_driver": (
        "deferred_until_connection_lifecycle_readiness",
        False,
    ),
    "mock_or_memory_database_driver": ("rejected_for_production_audit_store_boundary", False),
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_database_driver_selection_review_dependency_missing",
    "audit_store_storage_adapter_database_driver_selection_review_candidate_evidence_missing",
    "audit_store_storage_adapter_database_driver_selection_review_import_forbidden",
    "audit_store_storage_adapter_database_driver_selection_review_connection_forbidden",
    "audit_store_storage_adapter_database_driver_selection_review_runtime_forbidden",
    "audit_store_storage_adapter_database_driver_selection_review_secret_material_detected",
    "audit_store_storage_adapter_database_driver_selection_review_scope_overreach",
}

EXPECTED_DIAGNOSTICS = {
    "audit_store_storage_adapter_database_driver_selection_review_status",
    "selection_decision",
    "runtime_task_decision",
    "next_dependency",
    "selected_database_engine",
    "selected_provider_candidate_class",
    "selected_driver_candidate",
    "driver_selection_status",
    "database_driver_status",
    "database_driver_package_status",
    "database_driver_import_status",
    "database_dsn_status",
    "connection_provider_status",
    "tls_mode_status",
    "least_privilege_role_status",
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
    "driver_import_added_count",
    "driver_dependency_version_pinned_count",
    "go_mod_change_count",
    "dsn_parse_count",
    "sql_execution_count",
    "database_vendor_selected_count",
    "managed_database_product_selected_count",
    "database_provider_selected_count",
    "provider_account_resource_defined_count",
    "database_endpoint_defined_count",
    "database_connection_provider_created_count",
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
    "docs/platform/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.md": [
        SLICE_STATUS,
        SELECTION_DECISION,
        SELECTED_DRIVER_CANDIDATE,
        NEXT_DEPENDENCY,
        "不新增 Go import",
        "不创建 storage adapter runtime",
    ],
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1-plan.md": [
        SLICE_ID,
        SLICE_STATUS,
        SELECTION_DECISION,
        SELECTED_DRIVER_CANDIDATE,
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
        SELECTION_DECISION,
        NEXT_DEPENDENCY,
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Storage Adapter Database Driver Selection Review v1",
        SLICE_STATUS,
        NEXT_DEPENDENCY,
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Storage Adapter Database Driver Selection Review v1",
        SLICE_STATUS,
        NEXT_DEPENDENCY,
    ],
    "docs/features/workflow/README.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/radishmind-current-focus.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.py",
        "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.json",
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
        == "production_ops_secret_backend_audit_store_storage_adapter_database_driver_selection_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected slice status")
    require(slice_info.get("selection_decision") == SELECTION_DECISION, "unexpected selection decision")
    require(slice_info.get("entry_decision") == ENTRY_DECISION, "unexpected entry decision")
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "driver_import_added",
        "driver_dependency_version_pinned",
        "go_mod_changed",
        "dsn_defined",
        "dsn_parser_created",
        "connection_provider_created",
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
    require(set(dependency_rows) == set(EXPECTED_DEPENDENCIES), "dependency set drifted")
    for dep_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        row = dependency_rows.get(dep_id)
        require(row is not None, f"dependency missing: {dep_id}")
        require(row.get("evidence") == relative_path, f"dependency evidence mismatch: {dep_id}")
        require(row.get("status") == expected_status, f"dependency status mismatch: {dep_id}")
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"dependency source status mismatch: {dep_id}")


def assert_selection_review(fixture: dict[str, Any]) -> None:
    review = fixture.get("selection_review") or {}
    for key, expected in EXPECTED_SELECTION_REVIEW.items():
        require(review.get(key) == expected, f"selection_review.{key} must be {expected}")
    for key in EXPECTED_FALSE_FLAGS:
        require(review.get(key) is False, f"selection_review.{key} must be false")

    evidence = rows_by_id(fixture, "external_candidate_evidence", "id")
    require(set(evidence) == set(EXPECTED_EXTERNAL_EVIDENCE), "external candidate evidence set drifted")
    for evidence_id, expected in EXPECTED_EXTERNAL_EVIDENCE.items():
        row = evidence[evidence_id]
        require(row.get("candidate") == SELECTED_DRIVER_CANDIDATE, f"candidate drifted: {evidence_id}")
        require(row.get("source") == expected["source"], f"source drifted: {evidence_id}")
        require(row.get("status") == expected["status"], f"status drifted: {evidence_id}")
        require(set(row.get("evidence_points") or []) == expected["points"], f"evidence points drifted: {evidence_id}")

    candidates = rows_by_id(fixture, "candidate_evaluation", "candidate")
    require(set(candidates) == set(EXPECTED_CANDIDATES), "candidate evaluation set drifted")
    for candidate, (expected_status, expected_selected) in EXPECTED_CANDIDATES.items():
        row = candidates[candidate]
        require(row.get("status") == expected_status, f"candidate status drifted: {candidate}")
        require(row.get("selected") is expected_selected, f"candidate selected flag drifted: {candidate}")
        require(row.get("reason"), f"candidate reason missing: {candidate}")


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
    for field in {"secret_value", "dsn", "connection_string", "database_hostname", "table_name", "column_name"}:
        require(field in forbidden, f"forbidden diagnostic missing {field}")
    sample = diagnostics.get("sample") or {}
    require(sample.get("selected_database_engine") == SELECTED_DATABASE_ENGINE, "diagnostic sample engine drifted")
    require(sample.get("selected_provider_candidate_class") == SELECTED_PROVIDER_CLASS, "diagnostic sample class drifted")
    require(sample.get("selected_driver_candidate") == SELECTED_DRIVER_CANDIDATE, "diagnostic sample driver drifted")
    require(sample.get("next_dependency") == NEXT_DEPENDENCY, "diagnostic sample next dependency drifted")
    require(sample.get("driver_selection_status") == DRIVER_SELECTION_STATUS, "diagnostic driver status drifted")
    require(sample.get("database_driver_import_status") == DRIVER_IMPORT_STATUS, "diagnostic import status drifted")

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
        "database_connection_lifecycle_task_card",
        "storage_adapter_runtime_implementation_task_card",
        "database_connection_provider",
        "driver_import",
        "driver_dependency_version_pin",
        "dsn_parser",
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
        "storage_adapter_database_driver_selection_review_status": SLICE_STATUS,
        "storage_adapter_database_driver_selection_status": DRIVER_SELECTION_STATUS,
        "storage_adapter_selected_database_driver_candidate": SELECTED_DRIVER_CANDIDATE,
        "storage_adapter_database_driver_status": DATABASE_DRIVER_STATUS,
        "storage_adapter_database_driver_package_status": DRIVER_PACKAGE_STATUS,
        "storage_adapter_database_driver_import_status": DRIVER_IMPORT_STATUS,
        "storage_adapter_driver_dependency_version_status": "not_pinned",
        "storage_adapter_runtime_task_card_decision": ENTRY_DECISION,
        "storage_adapter_current_next_dependency": NEXT_DEPENDENCY,
        "storage_adapter_database_connection_provider_status": "not_created",
        "storage_adapter_database_dsn_status": "not_defined",
        "storage_adapter_runtime_task_card_status": "not_created",
        "storage_adapter_runtime_status": "not_created",
        "durable_audit_backend_status": MATRIX_BLOCKER_STATUS,
    }.items():
        require(boundary.get(field) == expected, f"matrix boundary {field} drifted")

    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == MATRIX_BLOCKER_STATUS, "durable blocker status drifted")
    require(durable.get("source") == SLICE_ID, "durable blocker source drifted")
    require(durable.get("unlock_condition") == NEXT_DEPENDENCY, "durable unlock condition drifted")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable must still block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable must still block resolver runtime")
    order = matrix.get("dependency_order") or []
    require("storage_adapter_database_driver_selection_review" in order, "dependency order missing driver review")
    require(
        order.index("storage_adapter_database_driver_selection_readiness")
        < order.index("storage_adapter_database_driver_selection_review")
        < order.index("audit_writer_runtime_entry_review"),
        "database driver selection review order drifted",
    )

    alignment = fixture.get("blocker_matrix_alignment") or {}
    require(
        alignment.get("durable_backend_blocker_status_after_review") == MATRIX_BLOCKER_STATUS,
        "matrix status drifted",
    )
    require(alignment.get("durable_backend_blocker_source_after_review") == SLICE_ID, "matrix source drifted")
    require(alignment.get("storage_adapter_current_next_dependency") == NEXT_DEPENDENCY, "matrix next drifted")
    require(alignment.get("runtime_task_card_decision") == ENTRY_DECISION, "matrix decision drifted")

    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    expected = fixture.get("implementation_readiness_alignment") or {}
    for field, value in expected.items():
        if field == "status":
            continue
        require(target.get(field) == value, f"implementation readiness {field} drifted")

    planned = rows_by_id(readiness, "planned_slices", "id")
    item = planned.get("audit-store-storage-adapter-database-driver-selection-review") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness missing driver review planned slice")


def assert_docs_and_registration() -> None:
    for path, literals in DOC_REFERENCES.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = "check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-readiness-v1.py"
    current = "check-production-ops-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.py"
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in (previous, current, matrix):
        require(script in check_repo, f"check-repo.py missing {script}")
    require(check_repo.index(previous) < check_repo.index(current) < check_repo.index(matrix), "check-repo.py order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1-plan.md",
    ]
    text = "\n".join(read(path) for path in paths)
    for literal in ["BEGIN PRIVATE KEY", "AKIA", "authorization:", "connection_uri_with_secret"]:
        require(literal not in text, f"secret-like literal found: {literal}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_selection_review(fixture)
    assert_failure_mapping(fixture)
    assert_diagnostics_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store storage adapter database driver selection review checks passed.")


if __name__ == "__main__":
    main()
