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
    "production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
TABLE_SCHEMA_ARTIFACT_PATH = REPO_ROOT / "contracts/production-secret-audit-storage-adapter.table-schema.json"
METADATA_CONTRACT_ARTIFACT_PATH = (
    REPO_ROOT / "contracts/production-secret-audit-storage-adapter.metadata-contract.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1"
SLICE_STATUS = "audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_defined"
READINESS_DECISION = "offline_adapter_smoke_strategy_defined_without_runtime"
CURRENT_BLOCKER_STATUS = "storage_adapter_negative_leakage_runtime_scan_boundary_readiness_defined_runtime_blocked"
CURRENT_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-negative-leakage-runtime-scan-boundary-readiness-v1"
)
CURRENT_NEXT_DEPENDENCY = "storage_adapter_runtime_implementation_entry_refresh_after_negative_leakage_runtime_scan_boundary"
RUNTIME_TASK_CARD_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_negative_leakage_runtime_scan_boundary"
)
MATRIX_BLOCKER_STATUS = (
    "storage_adapter_concrete_managed_database_provider_selection_readiness_defined_task_card_blocked"
)
MATRIX_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-concrete-managed-database-provider-selection-readiness-v1"
)
MATRIX_NEXT_DEPENDENCY = (
    "storage_adapter_concrete_managed_database_provider_selection_review"
)
MATRIX_RUNTIME_TASK_CARD_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_concrete_managed_database_provider_selection_readiness"
)
TABLE_SCHEMA_STATUS = "audit_store_storage_adapter_table_schema_artifact_materialized"
METADATA_CONTRACT_STATUS = "audit_store_storage_adapter_metadata_contract_artifact_materialized"

POSITIVE_FIXTURE = "scripts/checks/fixtures/production-secret-audit-storage-adapter-offline-smoke-positive-v1.json"
MISSING_MANIFEST_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-offline-smoke-missing-manifest-negative-v1.json"
)
RUNTIME_TOUCH_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-offline-smoke-runtime-touch-negative-v1.json"
)
SECRET_MATERIAL_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-offline-smoke-secret-material-negative-v1.json"
)

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.json"
        ),
        TABLE_SCHEMA_STATUS,
    ),
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json"
        ),
        METADATA_CONTRACT_STATUS,
    ),
    "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_backend_product_selection_review_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.json"
        ),
        "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.json"
        ),
        "audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined",
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

EXPECTED_RUNTIME_FALSE_FLAGS = {
    "database_vendor_selected_in_this_slice",
    "database_driver_selected_in_this_slice",
    "database_provider_created_in_this_slice",
    "database_connection_created_in_this_slice",
    "sql_created_in_this_slice",
    "ddl_created_in_this_slice",
    "physical_table_schema_created_in_this_slice",
    "schema_marker_runtime_created_in_this_slice",
    "migration_runner_created_in_this_slice",
    "offline_adapter_smoke_runner_created_in_this_slice",
    "offline_adapter_smoke_output_created_in_this_slice",
    "negative_leakage_runtime_scan_created_in_this_slice",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_REFERENCE_FIELDS = {
    "smoke_manifest_ref",
    "metadata_contract_ref",
    "table_schema_artifact_ref",
    "backend_product_evidence_ref",
    "positive_case_ref",
    "negative_case_ref",
    "failure_taxonomy_ref",
    "policy_version",
    "audit_ref",
}

EXPECTED_CASES = {
    "positive_metadata_only_write_candidate": (POSITIVE_FIXTURE, True),
    "missing_smoke_manifest_ref": (MISSING_MANIFEST_FIXTURE, False),
    "real_backend_touch_forbidden": (RUNTIME_TOUCH_FIXTURE, False),
    "secret_material_field_forbidden": (SECRET_MATERIAL_FIXTURE, False),
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_offline_smoke_strategy_missing",
    "audit_store_storage_adapter_offline_smoke_manifest_missing",
    "audit_store_storage_adapter_offline_smoke_table_schema_missing",
    "audit_store_storage_adapter_offline_smoke_runtime_touch_forbidden",
    "audit_store_storage_adapter_offline_smoke_secret_material_forbidden",
    "audit_store_storage_adapter_offline_smoke_physical_detail_forbidden",
    "audit_store_storage_adapter_offline_smoke_fallback_forbidden",
}

FORBIDDEN_CASE_FIELDS = {
    "raw_secret",
    "password",
    "token",
    "api_key",
    "authorization_header",
    "cookie",
    "dsn",
    "provider_raw_url",
    "database_hostname",
    "database_name",
    "table_name",
    "column_name",
    "column_type",
    "raw_request_payload",
    "raw_response_payload",
    "raw_audit_payload",
    "raw_storage_payload",
    "provider_error_detail",
    "database_error_detail",
    "scanner_raw_finding",
    "scan_output",
    "migration_output",
}

REQUIRED_CASE_FIELDS = {
    "schema_version",
    "kind",
    "case_id",
    "expected_valid",
    "smoke_manifest_ref",
    "smoke_strategy_ref",
    "metadata_contract_ref",
    "table_schema_artifact_ref",
    "backend_product_evidence_ref",
    "backend_product_class",
    "audit_event_ref",
    "storage_record_ref",
    "storage_adapter_result_ref",
    "write_status",
    "sanitized_diagnostic",
    "runtime_touch",
    "real_backend_touch",
    "secret_material_included",
    "physical_schema_detail_included",
    "policy_version",
    "audit_ref",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1-plan.md",
    (
        "scripts/checks/fixtures/"
        "production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.json"
    ),
    POSITIVE_FIXTURE,
    MISSING_MANIFEST_FIXTURE,
    RUNTIME_TOUCH_FIXTURE,
    SECRET_MATERIAL_FIXTURE,
    (
        "scripts/"
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "offline-adapter-smoke-strategy-readiness-v1.py"
    ),
}

SECRET_LITERAL_PATTERNS = [
    re.compile(r"Bearer\s+[A-Za-z0-9._-]+"),
    re.compile(r"-----BEGIN [A-Z ]+-----"),
    re.compile(r"AKIA[0-9A-Z]{16}"),
    re.compile(r"sk-[A-Za-z0-9]{20,}"),
]


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def relpath(path: Path) -> str:
    return path.relative_to(REPO_ROOT).as_posix()


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
        == "production_ops_secret_backend_audit_store_storage_adapter_offline_adapter_smoke_strategy_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected status")
    require(slice_info.get("readiness_decision") == READINESS_DECISION, "unexpected readiness decision")
    for field in ("task_card", "platform_topic"):
        relative_path = str(slice_info.get(field) or "")
        require(relative_path, f"{field} missing")
        require((REPO_ROOT / relative_path).exists(), f"{field} missing on disk: {relative_path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "database_provider_created",
        "database_connection_created",
        "offline_adapter_smoke_runner_created",
        "offline_adapter_smoke_output_created",
        "negative_leakage_runtime_scan_created",
        "sql_created",
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
    expected = {
        "status": SLICE_STATUS,
        "readiness_decision": READINESS_DECISION,
        "current_development_mode": "metadata_only_offline_adapter_smoke_strategy_no_runner",
        "table_schema_artifact_materialization_status": TABLE_SCHEMA_STATUS,
        "table_schema_artifact_status": "materialized_static_logical_table_schema",
        "metadata_contract_artifact_status": "materialized_static_metadata_contract",
        "backend_product_selection_status": "selected_static_product_class_without_backend_provider",
        "selected_backend_product_class": "managed_database_append_only_table",
        "database_provider_driver_dsn_tls_role_policy_status": "defined_without_runtime",
        "offline_adapter_smoke_strategy_status": READINESS_DECISION,
        "offline_adapter_smoke_manifest_status": "metadata_only_smoke_manifest_defined",
        "offline_adapter_smoke_positive_case_status": "metadata_only_positive_case_defined",
        "offline_adapter_smoke_negative_case_status": "metadata_only_negative_case_defined",
        "offline_adapter_smoke_coverage_status": "metadata_contract_table_schema_append_only_coverage_defined",
        "offline_adapter_smoke_backend_touch_policy_status": "real_backend_touch_forbidden",
        "offline_adapter_smoke_secret_material_policy_status": "committed_smoke_strategy_secret_material_forbidden",
        "offline_adapter_smoke_diagnostic_policy_status": "sanitized_diagnostic_allowlist_defined",
        "offline_adapter_smoke_runner_status": "not_created",
        "offline_adapter_smoke_output_status": "not_created",
        "negative_leakage_runtime_scan_boundary_status": "defined_without_runtime",
        "current_next_dependency": CURRENT_NEXT_DEPENDENCY,
        "runtime_task_card_decision": RUNTIME_TASK_CARD_DECISION,
    }
    for field, expected_value in expected.items():
        require(boundary.get(field) == expected_value, f"readiness_boundary.{field} drifted")
    for field in EXPECTED_RUNTIME_FALSE_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")


def case_is_valid(candidate: dict[str, Any]) -> bool:
    if set(REQUIRED_CASE_FIELDS) - set(candidate):
        return False
    if candidate.get("schema_version") != 1:
        return False
    if candidate.get("kind") != "production_secret_audit_storage_adapter_offline_smoke_case_v1":
        return False
    if candidate.get("smoke_strategy_ref") != SLICE_ID:
        return False
    if candidate.get("metadata_contract_ref") != relpath(METADATA_CONTRACT_ARTIFACT_PATH):
        return False
    if candidate.get("table_schema_artifact_ref") != relpath(TABLE_SCHEMA_ARTIFACT_PATH):
        return False
    if candidate.get("backend_product_class") != "managed_database_append_only_table":
        return False
    if candidate.get("runtime_touch") is not False or candidate.get("real_backend_touch") is not False:
        return False
    if candidate.get("secret_material_included") is not False:
        return False
    if candidate.get("physical_schema_detail_included") is not False:
        return False
    if FORBIDDEN_CASE_FIELDS & set(candidate):
        return False
    if "forbidden_runtime_mechanism" in candidate:
        return False
    diagnostic = candidate.get("sanitized_diagnostic") or {}
    return isinstance(diagnostic, dict) and bool(diagnostic.get("failure_code"))


def assert_smoke_strategy(fixture: dict[str, Any]) -> None:
    strategy = fixture.get("offline_adapter_smoke_strategy") or {}
    require(strategy.get("status") == "metadata_only_strategy_defined", "strategy status drifted")
    require(strategy.get("manifest_kind") == "audit_storage_adapter_offline_smoke_manifest_reference", "manifest kind drifted")
    require(set(strategy.get("required_reference_fields") or []) == EXPECTED_REFERENCE_FIELDS, "reference fields drifted")
    case_families = set(strategy.get("case_families") or [])
    for case_id in EXPECTED_CASES:
        require(case_id in case_families, f"strategy missing case family {case_id}")
    forbidden_mechanisms = set(strategy.get("forbidden_runtime_mechanisms") or [])
    for mechanism in {
        "offline_adapter_smoke_runner",
        "storage_adapter_runtime",
        "database_connection_provider",
        "database_connection",
        "database_driver",
        "migration_runner",
        "schema_marker_runtime",
        "committed_smoke_output",
    }:
        require(mechanism in forbidden_mechanisms, f"forbidden mechanism missing {mechanism}")


def assert_diagnostic_envelope(fixture: dict[str, Any]) -> None:
    envelope = fixture.get("diagnostic_envelope") or {}
    allowed = set(envelope.get("allowed_fields") or [])
    forbidden = set(envelope.get("forbidden_fields") or [])
    for field in {"failure_code", "failure_boundary", "sanitized_diagnostic", "audit_ref", "policy_version"}:
        require(field in allowed, f"diagnostic allowed field missing {field}")
    require(FORBIDDEN_CASE_FIELDS <= forbidden, "diagnostic forbidden fields drifted")
    require(not (allowed & FORBIDDEN_CASE_FIELDS), "diagnostic allowed fields include forbidden material")


def assert_smoke_cases(fixture: dict[str, Any]) -> None:
    cases = rows_by_id(fixture, "smoke_case_matrix", "case")
    require(set(cases) == set(EXPECTED_CASES), "smoke case ids drifted")
    for case_id, (relative_path, expected_valid) in EXPECTED_CASES.items():
        item = cases[case_id]
        require(item.get("fixture") == relative_path, f"{case_id} fixture path drifted")
        require(item.get("expected_valid") is expected_valid, f"{case_id} expected_valid drifted")
        candidate = load_json(REPO_ROOT / relative_path)
        require(candidate.get("case_id") == case_id, f"{case_id} candidate id drifted")
        require(candidate.get("expected_valid") is expected_valid, f"{case_id} candidate expected_valid drifted")
        require(case_is_valid(candidate) is expected_valid, f"{case_id} static validation result drifted")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "failure_mapping", "code")
    require(set(rows) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in rows.items():
        require(item.get("failure_boundary"), f"{code} missing failure boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} missing diagnostic")
        require("secret value" not in diagnostic.lower(), f"{code} diagnostic must remain sanitized")


def assert_no_secret_material_scan(fixture: dict[str, Any]) -> None:
    scan = fixture.get("no_secret_material_scan") or {}
    require(scan.get("status") == "implemented_static_scan", "no secret material scan status drifted")
    scanned = set(scan.get("scanned_artifacts") or [])
    expected = {relpath(FIXTURE_PATH), POSITIVE_FIXTURE, MISSING_MANIFEST_FIXTURE, RUNTIME_TOUCH_FIXTURE, SECRET_MATERIAL_FIXTURE}
    require(expected <= scanned, "no secret scan target list missing expected artifacts")
    for relative_path in scanned:
        text = read(str(relative_path))
        for pattern in SECRET_LITERAL_PATTERNS:
            require(not pattern.search(text), f"secret-like literal found in {relative_path}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    allowed = set(guard.get("allowed_added_artifacts") or [])
    require(EXPECTED_ALLOWED_ARTIFACTS <= allowed, "allowed added artifacts missing expected paths")
    for relative_path in allowed:
        require((REPO_ROOT / relative_path).exists(), f"allowed artifact missing: {relative_path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "database_vendor_selection_artifact",
        "db_driver",
        "dsn_parser",
        "database_connection_provider",
        "database_connection",
        "offline_adapter_smoke_runner",
        "committed_offline_adapter_smoke_output",
        "negative_leakage_runtime_scan",
        "sql_migration",
        "ddl",
        "schema_marker_runtime",
        "migration_runner",
        "storage_adapter_runtime_implementation_task_card",
        "storage_adapter_runtime",
        "audit_store_runtime_implementation_task_card",
        "audit_store_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact kind missing {artifact}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden runtime artifact exists: {relative_path}")


def assert_blocker_matrix_alignment(fixture: dict[str, Any]) -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    alignment = fixture.get("blocker_matrix_alignment") or {}
    boundary = matrix.get("matrix_boundary") or {}
    require(boundary.get("durable_audit_backend_status") == MATRIX_BLOCKER_STATUS, "matrix durable blocker drifted")
    require(boundary.get("storage_adapter_current_next_dependency") == MATRIX_NEXT_DEPENDENCY, "matrix next dependency drifted")
    require(
        boundary.get("storage_adapter_runtime_task_card_decision") == MATRIX_RUNTIME_TASK_CARD_DECISION,
        "matrix runtime task card decision drifted",
    )
    require(
        boundary.get("storage_adapter_offline_adapter_smoke_strategy_status") == READINESS_DECISION,
        "matrix offline adapter smoke strategy status drifted",
    )
    require(boundary.get("storage_adapter_offline_adapter_smoke_runner_status") == "not_created", "matrix runner drifted")
    require(boundary.get("storage_adapter_offline_adapter_smoke_output_status") == "not_created", "matrix output drifted")
    require(
        boundary.get("storage_adapter_negative_leakage_runtime_scan_boundary_status")
        == "defined_without_runtime",
        "matrix negative leakage runtime scan boundary drifted",
    )
    for field in ("storage_adapter_sql_migration_status", "storage_adapter_ddl_status"):
        require(boundary.get(field) == "not_created", f"matrix {field} drifted")

    require(
        alignment.get("durable_backend_blocker_status_after_offline_adapter_smoke_strategy") == MATRIX_BLOCKER_STATUS,
        "fixture blocker status drifted",
    )
    require(alignment.get("storage_adapter_current_next_dependency") == MATRIX_NEXT_DEPENDENCY, "fixture next drifted")
    require(alignment.get("runtime_task_card_decision") == MATRIX_RUNTIME_TASK_CARD_DECISION, "fixture decision drifted")
    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == MATRIX_BLOCKER_STATUS, "durable blocker row status drifted")
    require(durable.get("source") == MATRIX_BLOCKER_SOURCE, "durable blocker row source drifted")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in (fixture.get("implementation_readiness_alignment") or {}).items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")
    planned = rows_by_id(readiness, "planned_slices", "id")
    item = planned.get("audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness") or {}
    require(item.get("status") == SLICE_STATUS, "planned slice status drifted")
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        "docs/platform/production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.md": [
            SLICE_STATUS,
            READINESS_DECISION,
            CURRENT_NEXT_DEPENDENCY,
            RUNTIME_TASK_CARD_DECISION,
            POSITIVE_FIXTURE,
            RUNTIME_TOUCH_FIXTURE,
            SECRET_MATERIAL_FIXTURE,
        ],
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1-plan.md": [
            SLICE_STATUS,
            READINESS_DECISION,
            CURRENT_NEXT_DEPENDENCY,
            "not_created",
        ],
        "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": [
            SLICE_STATUS,
            CURRENT_NEXT_DEPENDENCY,
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            SLICE_STATUS,
            CURRENT_BLOCKER_STATUS,
            CURRENT_NEXT_DEPENDENCY,
        ],
        "docs/radishmind-current-focus.md": [SLICE_STATUS, CURRENT_NEXT_DEPENDENCY],
        "docs/features/README.md": [SLICE_STATUS, CURRENT_NEXT_DEPENDENCY],
        "docs/features/workflow/README.md": [SLICE_STATUS, CURRENT_NEXT_DEPENDENCY],
        "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, CURRENT_NEXT_DEPENDENCY],
        "docs/radishmind-integration-contracts.md": [SLICE_STATUS, CURRENT_NEXT_DEPENDENCY],
        "docs/radishmind-architecture.md": [SLICE_STATUS, CURRENT_NEXT_DEPENDENCY],
        "docs/radishmind-product-scope.md": [SLICE_STATUS, CURRENT_NEXT_DEPENDENCY],
        "docs/platform/README.md": [SLICE_STATUS, CURRENT_NEXT_DEPENDENCY],
        "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
        "docs/task-cards/production-secret-backend-audit-store-runtime-blocker-matrix-v1-plan.md": [
            SLICE_STATUS,
            CURRENT_BLOCKER_STATUS,
            CURRENT_NEXT_DEPENDENCY,
        ],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
            SLICE_STATUS,
            CURRENT_NEXT_DEPENDENCY,
        ],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1.py",
            SLICE_STATUS,
        ],
        "docs/devlogs/2026-W27.md": [SLICE_STATUS, CURRENT_NEXT_DEPENDENCY],
    }
    for relative_path, literals in docs.items():
        text = read(relative_path)
        missing = sorted(literal for literal in literals if literal not in text)
        require(not missing, f"{relative_path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    table_schema = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "table-schema-artifact-materialization-v1.py"
    )
    current = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "offline-adapter-smoke-strategy-readiness-v1.py"
    )
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in (table_schema, current, matrix):
        require(script in check_repo, f"check-repo.py missing {script}")
    require(check_repo.index(table_schema) < check_repo.index(current) < check_repo.index(matrix), "check-repo order drifted")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    require(TABLE_SCHEMA_ARTIFACT_PATH.exists(), "table schema artifact missing")
    require(METADATA_CONTRACT_ARTIFACT_PATH.exists(), "metadata contract artifact missing")

    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_readiness_boundary(fixture)
    assert_smoke_strategy(fixture)
    assert_diagnostic_envelope(fixture)
    assert_smoke_cases(fixture)
    assert_failure_mapping(fixture)
    assert_no_secret_material_scan(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    print("production ops secret backend audit store storage adapter offline adapter smoke strategy readiness checks passed.")


if __name__ == "__main__":
    main()
