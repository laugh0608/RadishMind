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
    "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1"
SLICE_STATUS = "audit_store_storage_adapter_concrete_database_selection_readiness_defined"
READINESS_DECISION = "concrete_database_selection_readiness_defined_without_database_selection"
ENTRY_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_concrete_database_selection_readiness"
PREVIOUS_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary_defined"
)
PREVIOUS_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_negative_leakage_runtime_scan_boundary_entry_refresh"
)
NEXT_DEPENDENCY = "storage_adapter_concrete_database_selection_review"
SELECTED_PRODUCT_CLASS = "managed_database_append_only_table"
SELECTED_PRODUCT_PROFILE = "reserved_managed_database_append_only_table_profile"
MATRIX_BLOCKER_STATUS = "storage_adapter_concrete_database_selection_readiness_defined_task_card_blocked"
CURRENT_MATRIX_BLOCKER_STATUS = "storage_adapter_database_connection_lifecycle_readiness_defined_task_card_blocked"
CURRENT_MATRIX_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1"
)
CURRENT_ENTRY_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_readiness"
CURRENT_NEXT_DEPENDENCY = "storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_readiness"
CURRENT_DATABASE_SELECTION_STATUS = "selected_database_engine_without_vendor_or_provider"
CURRENT_DATABASE_SELECTION_REVIEW_STATUS = "audit_store_storage_adapter_concrete_database_selection_review_defined"
CURRENT_SELECTED_DATABASE_ENGINE = "postgresql_compatible_append_only_relational_database"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-negative-leakage-runtime-scan-boundary-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
            "after-negative-leakage-runtime-scan-boundary-v1.json"
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
    "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.json"
        ),
        "audit_store_storage_adapter_table_schema_artifact_materialized",
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

EXPECTED_BOUNDARY = {
    "status": SLICE_STATUS,
    "readiness_decision": READINESS_DECISION,
    "entry_decision": ENTRY_DECISION,
    "previous_entry_refresh_status": PREVIOUS_STATUS,
    "previous_runtime_task_decision": PREVIOUS_DECISION,
    "selected_backend_product_class": SELECTED_PRODUCT_CLASS,
    "selected_backend_product_profile": SELECTED_PRODUCT_PROFILE,
    "candidate_source_status": "metadata_only_candidate_source_defined",
    "candidate_input_evidence_status": "metadata_only_input_evidence_defined",
    "candidate_fields_status": "metadata_only_candidate_fields_defined",
    "candidate_evaluation_matrix_status": "metadata_only_evaluation_matrix_defined",
    "candidate_rejection_gate_status": "fail_closed_rejection_gate_defined",
    "database_selection_review_status": "not_started",
    "concrete_database_selection_status": "readiness_defined_without_database_selection",
    "database_product_status": "not_selected",
    "database_vendor_status": "not_selected",
    "database_connection_provider_status": "not_created",
    "database_driver_status": "not_selected",
    "database_dsn_status": "not_defined",
    "database_connection_status": "not_created",
    "sql_migration_status": "not_created",
    "ddl_status": "not_created",
    "physical_table_schema_status": "not_created",
    "schema_marker_runtime_status": "not_created",
    "migration_runner_status": "not_created",
    "next_dependency": NEXT_DEPENDENCY,
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "storage_adapter_client_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_writer_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "delivery_runtime_status": "not_created",
    "operator_approval_runtime_status": "not_created",
    "credential_handle_runtime_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "no_leakage_smoke_runtime_status": "not_created",
    "production_resolver_runtime_task_card_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "concrete_database_selected_in_this_slice",
    "database_product_selected_in_this_slice",
    "database_vendor_selected_in_this_slice",
    "database_connection_provider_enabled",
    "database_driver_selected_in_this_slice",
    "database_connection_created_in_this_slice",
    "dsn_defined_in_this_slice",
    "sql_migration_created_in_this_slice",
    "ddl_created_in_this_slice",
    "physical_table_schema_created_in_this_slice",
    "schema_marker_runtime_created_in_this_slice",
    "migration_runner_created_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "storage_adapter_client_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "production_resolver_runtime_task_card_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_CANDIDATE_FIELDS = {
    "candidate_key",
    "candidate_profile_ref",
    "candidate_product_class",
    "deployment_model_ref",
    "append_only_capability_ref",
    "transaction_boundary_ref",
    "ordering_guarantee_ref",
    "idempotency_policy_ref",
    "retention_redaction_policy_ref",
    "tls_policy_ref",
    "role_policy_ref",
    "secret_ref_dsn_policy_ref",
    "schema_marker_handoff_ref",
    "migration_review_ref",
    "offline_smoke_strategy_ref",
    "negative_leakage_scan_boundary_ref",
    "rollback_recovery_ref",
    "operator_review_ref",
    "decision_rationale_ref",
}

EXPECTED_EVALUATION_DIMENSIONS = {
    "append_only_insert_semantics",
    "metadata_contract_compatibility",
    "logical_schema_compatibility",
    "idempotency_duplicate_handling",
    "ordering_and_durability",
    "retention_redaction_compatibility",
    "secret_ref_only_connection_policy",
    "tls_and_least_privilege_role_policy",
    "migration_schema_marker_handoff",
    "offline_validation_negative_leakage_coverage",
    "rollback_recovery_semantics",
    "sanitized_diagnostics",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_concrete_database_selection_readiness_dependency_missing",
    "audit_store_storage_adapter_concrete_database_selection_readiness_missing_evidence",
    "audit_store_storage_adapter_concrete_database_selection_readiness_database_selection_forbidden",
    "audit_store_storage_adapter_concrete_database_selection_readiness_runtime_forbidden",
    "audit_store_storage_adapter_concrete_database_selection_readiness_secret_material_detected",
    "audit_store_storage_adapter_concrete_database_selection_readiness_scope_overreach",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1-plan.md",
    (
        "scripts/checks/fixtures/"
        "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.json"
    ),
    "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py",
}

SECRET_LITERAL_PATTERNS = [
    re.compile(r"Bearer\s+[A-Za-z0-9._-]+"),
    re.compile(r"-----BEGIN [A-Z ]+-----"),
    re.compile(r"AKIA[0-9A-Z]{16}"),
    re.compile(r"sk-[A-Za-z0-9]{20,}"),
    re.compile(r"://[^\s:/]+:[^\s@]+@"),
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


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_concrete_database_selection_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected status")
    require(slice_info.get("readiness_decision") == READINESS_DECISION, "unexpected readiness decision")
    require(slice_info.get("entry_decision") == ENTRY_DECISION, "unexpected entry decision")
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "concrete_database_selected",
        "database_product_selected",
        "database_vendor_selected",
        "database_provider_selected",
        "database_connection_created",
        "database_driver_selected",
        "dsn_defined",
        "sql_migration_created",
        "ddl_created",
        "physical_table_schema_created",
        "schema_marker_runtime_created",
        "migration_runner_created",
        "storage_adapter_runtime_task_card_created",
        "storage_adapter_runtime_created",
        "audit_store_runtime_task_card_created",
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


def assert_readiness_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("readiness_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"readiness_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"readiness_boundary.{field} must stay false")


def assert_candidate_policy(fixture: dict[str, Any]) -> None:
    require(set(fixture.get("candidate_input_fields") or []) == EXPECTED_CANDIDATE_FIELDS, "candidate fields drifted")
    require(
        set(fixture.get("candidate_evaluation_dimensions") or []) == EXPECTED_EVALUATION_DIMENSIONS,
        "candidate evaluation dimensions drifted",
    )
    rejection_conditions = set(fixture.get("rejection_conditions") or [])
    for condition in {
        "candidate_contains_database_endpoint",
        "candidate_selects_vendor_in_readiness_slice",
        "candidate_selects_driver_in_readiness_slice",
        "candidate_contains_dsn_or_secret_material",
        "candidate_contains_physical_table_or_column_detail",
        "candidate_claims_runtime_ready",
        "candidate_opens_repository_mode_or_public_api",
    }:
        require(condition in rejection_conditions, f"rejection condition missing: {condition}")


def assert_diagnostics_failures_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(not (allowed & forbidden), "diagnostic allowlist intersects forbidden fields")
    require(sample.get("readiness_decision") == READINESS_DECISION, "diagnostic readiness decision drifted")
    require(sample.get("runtime_task_decision") == ENTRY_DECISION, "diagnostic runtime decision drifted")
    require(sample.get("next_dependency") == NEXT_DEPENDENCY, "diagnostic next dependency drifted")
    require(sample.get("database_product_status") == "not_selected", "diagnostic database product drifted")
    require(sample.get("database_vendor_status") == "not_selected", "diagnostic database vendor drifted")
    require(sample.get("database_connection_provider_status") == "not_created", "diagnostic provider drifted")
    require(sample.get("database_driver_status") == "not_selected", "diagnostic driver drifted")
    require(sample.get("database_dsn_status") == "not_defined", "diagnostic dsn drifted")
    require(sample.get("schema_marker_runtime_status") == "not_created", "diagnostic schema marker drifted")
    require(sample.get("migration_runner_status") == "not_created", "diagnostic migration runner drifted")
    require(sample.get("storage_adapter_runtime_status") == "not_created", "diagnostic runtime created")

    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} boundary missing")
        require(item.get("sanitized_diagnostic"), f"{code} diagnostic missing")

    no_fallback = fixture.get("no_fallback_policy") or {}
    require(no_fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for source in {
        "static_product_class_selection",
        "reserved_backend_product_profile",
        "metadata_contract_artifact_materialization",
        "table_schema_artifact_materialization",
        "offline_adapter_smoke_strategy",
        "negative_leakage_runtime_scan_boundary",
        "test_only_fake_resolver",
        "memory_store",
        "mock_database_provider",
        "historical_smoke",
        "previous_checker_success",
    }:
        require(source in set(no_fallback.get("forbidden_sources") or []), f"missing forbidden fallback {source}")

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for counter, value in counters.items():
        require(value == 0, f"{counter} must stay 0")


def assert_no_secret_material_scan(fixture: dict[str, Any]) -> None:
    scan = fixture.get("no_secret_material_scan") or {}
    require(scan.get("status") == "implemented_static_scan", "no secret material scan status drifted")
    expected = EXPECTED_ALLOWED_ARTIFACTS - {
        "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py"
    }
    scanned = set(scan.get("scanned_artifacts") or [])
    require(expected <= scanned, "no secret scan target list missing expected artifacts")
    for relative_path in scanned:
        text = read(str(relative_path))
        for pattern in SECRET_LITERAL_PATTERNS:
            require(not pattern.search(text), f"secret-like literal found in {relative_path}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    require(set(guard.get("allowed_added_artifacts") or []) == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifacts drifted")
    for path in EXPECTED_ALLOWED_ARTIFACTS:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing: {path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "database_product_selection_artifact",
        "backend_vendor_selection_artifact",
        "storage_adapter_runtime_implementation_task_card",
        "storage_adapter_runtime",
        "storage_adapter_client",
        "database_connection_provider",
        "database_connection",
        "database_driver",
        "dsn_parser",
        "sql_migration",
        "ddl",
        "physical_table_schema",
        "schema_marker_runtime",
        "migration_runner",
        "audit_store_runtime_implementation_task_card",
        "audit_store_runtime",
        "production_resolver_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact missing: {artifact}")
    for path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(path)).exists(), f"forbidden artifact exists: {path}")


def assert_blocker_matrix_alignment(fixture: dict[str, Any]) -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    for field, value in {
        "storage_adapter_concrete_database_selection_readiness_status": SLICE_STATUS,
        "storage_adapter_concrete_database_selection_review_status": CURRENT_DATABASE_SELECTION_REVIEW_STATUS,
        "storage_adapter_concrete_database_selection_status": CURRENT_DATABASE_SELECTION_STATUS,
        "storage_adapter_selected_database_engine": CURRENT_SELECTED_DATABASE_ENGINE,
        "storage_adapter_selected_database_engine_status": "selected_without_vendor_product_driver_or_provider",
        "storage_adapter_candidate_input_evidence_status": "metadata_only_input_evidence_defined",
        "storage_adapter_candidate_evaluation_matrix_status": "metadata_only_evaluation_matrix_defined",
        "storage_adapter_database_selection_review_status": CURRENT_DATABASE_SELECTION_REVIEW_STATUS,
        "storage_adapter_runtime_task_card_decision": CURRENT_ENTRY_DECISION,
        "storage_adapter_current_next_dependency": CURRENT_NEXT_DEPENDENCY,
        "storage_adapter_database_product_status": "engine_selected_without_managed_product",
        "storage_adapter_database_vendor_status": "not_selected",
        "storage_adapter_database_connection_provider_status": "not_created",
        "storage_adapter_database_driver_status": "selected_reference_only",
        "storage_adapter_database_dsn_status": "not_defined",
        "storage_adapter_schema_marker_runtime_status": "not_created",
        "storage_adapter_migration_runner_status": "not_created",
        "storage_adapter_runtime_task_card_status": "not_created",
        "storage_adapter_runtime_status": "not_created",
        "durable_audit_backend_status": CURRENT_MATRIX_BLOCKER_STATUS,
    }.items():
        require(boundary.get(field) == value, f"matrix boundary {field} drifted")

    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == CURRENT_MATRIX_BLOCKER_STATUS, "durable status drifted")
    require(durable.get("source") == CURRENT_MATRIX_BLOCKER_SOURCE, "durable source drifted")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable must block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable must block resolver runtime")

    alignment = fixture.get("blocker_matrix_alignment") or {}
    require(
        alignment.get("durable_backend_blocker_status_after_readiness") == CURRENT_MATRIX_BLOCKER_STATUS,
        "fixture matrix status drifted",
    )
    require(
        alignment.get("durable_backend_blocker_source_after_readiness") == CURRENT_MATRIX_BLOCKER_SOURCE,
        "fixture matrix source drifted",
    )
    require(alignment.get("storage_adapter_current_next_dependency") == CURRENT_NEXT_DEPENDENCY, "fixture next drifted")
    require(alignment.get("runtime_task_card_decision") == CURRENT_ENTRY_DECISION, "fixture decision drifted")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    expected = fixture.get("implementation_readiness_alignment") or {}
    for field, value in expected.items():
        if field == "status":
            continue
        require(target.get(field) == value, f"implementation readiness {field} drifted")

    planned = rows_by_id(readiness, "planned_slices", "id")
    item = planned.get("audit-store-storage-adapter-concrete-database-selection-readiness") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness missing planned slice")
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        "docs/platform/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.md": [
            SLICE_STATUS,
            READINESS_DECISION,
            ENTRY_DECISION,
            NEXT_DEPENDENCY,
            "metadata-only 候选字段",
        ],
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1-plan.md": [
            SLICE_ID,
            SLICE_STATUS,
            ENTRY_DECISION,
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
            READINESS_DECISION,
            NEXT_DEPENDENCY,
        ],
        "docs/platform/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Concrete Database Selection Readiness v1",
            SLICE_STATUS,
            NEXT_DEPENDENCY,
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Concrete Database Selection Readiness v1",
            SLICE_STATUS,
            NEXT_DEPENDENCY,
        ],
        "docs/features/workflow/README.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/radishmind-current-focus.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
            SLICE_ID,
            SLICE_STATUS,
            NEXT_DEPENDENCY,
        ],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py",
            "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.json",
        ],
        "docs/devlogs/2026-W27.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    }
    for path, literals in docs.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
        "after-negative-leakage-runtime-scan-boundary-v1.py"
    )
    current = "check-production-ops-secret-backend-audit-store-storage-adapter-concrete-database-selection-readiness-v1.py"
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    resolver = "check-production-ops-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.py"
    for script in (previous, current, matrix, resolver):
        require(script in check_repo, f"check-repo.py missing {script}")
    require(
        check_repo.index(previous) < check_repo.index(current) < check_repo.index(matrix) < check_repo.index(resolver),
        "check-repo.py registration order drifted",
    )


def assert_no_secret_literals() -> None:
    text = "\n".join(
        read(path)
        for path in EXPECTED_ALLOWED_ARTIFACTS
        if path.endswith(".md") or path.endswith(".json")
    )
    found = [pattern.pattern for pattern in SECRET_LITERAL_PATTERNS if pattern.search(text)]
    require(not found, f"concrete database selection readiness contains secret-looking literal patterns: {found}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_readiness_boundary(fixture)
    assert_candidate_policy(fixture)
    assert_diagnostics_failures_and_policies(fixture)
    assert_no_secret_material_scan(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store storage adapter concrete database selection readiness checks passed.")


if __name__ == "__main__":
    main()
