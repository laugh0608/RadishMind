#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any

try:
    from jsonschema import Draft202012Validator
except ModuleNotFoundError as exc:  # pragma: no cover - guarded by bootstrap/check-repo.
    raise SystemExit("jsonschema is required; run ./scripts/bootstrap-dev.sh first") from exc


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
APPEND_ONLY_BOUNDARY_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1.json"
)
METADATA_CONTRACT_ARTIFACT_PATH = REPO_ROOT / "contracts/production-secret-audit-storage-adapter.metadata-contract.json"
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
TABLE_SCHEMA_PATH = REPO_ROOT / "contracts/production-secret-audit-storage-adapter.table-schema.json"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1"
SLICE_STATUS = "audit_store_storage_adapter_table_schema_artifact_materialized"
TASK_CARD_DECISION = "table_schema_artifact_materialization_task_card_defined_after_entry_review"
MATERIALIZATION_DECISION = "table_schema_artifact_materialized_after_task_card"
NEXT_DEPENDENCY = "storage_adapter_offline_adapter_smoke_strategy_readiness"
TABLE_SCHEMA_VERSION = "audit-storage-adapter-table-schema-v1"
METADATA_CONTRACT_VERSION = "audit-storage-adapter-metadata-contract-v1"
ENTRY_REVIEW_STATUS = "audit_store_storage_adapter_table_schema_artifact_materialization_entry_review_defined"
RUNTIME_TASK_CARD_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_table_schema_artifact_materialization"
)
MATRIX_BLOCKER_STATUS = "storage_adapter_table_schema_artifact_materialized_runtime_blocked"
CURRENT_RUNTIME_TASK_CARD_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_offline_adapter_smoke_strategy_readiness"
)
CURRENT_MATRIX_BLOCKER_STATUS = "storage_adapter_offline_adapter_smoke_strategy_readiness_defined_runtime_blocked"
CURRENT_NEXT_DEPENDENCY = "storage_adapter_negative_leakage_runtime_scan_boundary_readiness"

POSITIVE_FIXTURE = "scripts/checks/fixtures/production-secret-audit-storage-adapter-table-schema-positive-v1.json"
MISSING_REQUIRED_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-table-schema-missing-required-negative-v1.json"
)
PHYSICAL_DETAIL_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-table-schema-physical-detail-negative-v1.json"
)
SECRET_MATERIAL_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-table-schema-secret-material-negative-v1.json"
)
ADDITIONAL_PROPERTIES_FIXTURE = (
    "scripts/checks/fixtures/"
    "production-secret-audit-storage-adapter-table-schema-additional-properties-negative-v1.json"
)

EXPECTED_NEGATIVE_FIXTURES = {
    MISSING_REQUIRED_FIXTURE,
    PHYSICAL_DETAIL_FIXTURE,
    SECRET_MATERIAL_FIXTURE,
    ADDITIONAL_PROPERTIES_FIXTURE,
}
EXPECTED_ALLOWED_ARTIFACTS = {
    TABLE_SCHEMA_PATH.relative_to(REPO_ROOT).as_posix(),
    "docs/contracts/production-secret-audit-storage-adapter-table-schema.md",
    "docs/platform/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1-plan.md",
    (
        "scripts/checks/fixtures/"
        "production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.json"
    ),
    POSITIVE_FIXTURE,
    *EXPECTED_NEGATIVE_FIXTURES,
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
EXPECTED_RUNTIME_FALSE_FLAGS = {
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
    "audit_store_storage_adapter_table_schema_artifact_missing",
    "audit_store_storage_adapter_table_schema_version_mismatch",
    "audit_store_storage_adapter_table_schema_field_group_missing",
    "audit_store_storage_adapter_table_schema_contract_compatibility_missing",
    "audit_store_storage_adapter_table_schema_positive_fixture_invalid",
    "audit_store_storage_adapter_table_schema_negative_fixture_allowed",
    "audit_store_storage_adapter_table_schema_physical_detail_allowed",
    "audit_store_storage_adapter_table_schema_secret_material_allowed",
    "audit_store_storage_adapter_table_schema_runtime_scope_overreach",
    "audit_store_storage_adapter_table_schema_materialization_fallback_forbidden",
    "audit_store_storage_adapter_table_schema_materialization_secret_material_detected",
}
EXPECTED_GROUPS = {
    "identity",
    "ordering",
    "payload_reference",
    "retention_redaction",
    "delivery_recovery",
    "diagnostics",
}
EXPECTED_FORBIDDEN_FIELDS = {
    "secret_value",
    "raw_secret",
    "password",
    "token",
    "api_key",
    "authorization_header",
    "cookie",
    "provider_raw_url",
    "resolver_backend_url",
    "dsn",
    "database_hostname",
    "database_name",
    "table_name",
    "column_name",
    "column_type",
    "index_name",
    "constraint_name",
    "partition_policy",
    "ddl",
    "sql",
    "sql_migration",
    "database_sequence",
    "trigger",
    "database_function",
    "schema_version_table",
    "migration_command",
    "cloud_credential",
    "credential_payload",
    "raw_request_payload",
    "raw_response_payload",
    "raw_audit_payload",
    "raw_event_payload",
    "raw_writer_payload",
    "raw_storage_payload",
    "payload_hash",
    "event_payload_hash",
    "secret_derived_hash",
    "provider_error_detail",
    "database_error_detail",
    "scanner_raw_finding",
    "scan_output",
    "schema_marker_raw_output",
    "migration_output",
}
SECRET_LITERAL_PATTERNS = [
    re.compile(r"Bearer\s+[A-Za-z0-9._-]+"),
    re.compile(r"-----BEGIN [A-Z ]+-----"),
    re.compile(r"AKIA[0-9A-Z]{16}"),
    re.compile(r"sk-[A-Za-z0-9]{20,}"),
    re.compile(r"postgres://[^\s\"]+"),
    re.compile(r"mysql://[^\s\"]+"),
    re.compile(r"jdbc:[^\s\"]+"),
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


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def recursive_keys(value: Any) -> set[str]:
    if isinstance(value, dict):
        keys = set(value)
        for child in value.values():
            keys.update(recursive_keys(child))
        return keys
    if isinstance(value, list):
        keys: set[str] = set()
        for child in value:
            keys.update(recursive_keys(child))
        return keys
    return set()


def forbidden_fields_from_schema(schema: dict[str, Any]) -> set[str]:
    fields: set[str] = set()
    for guard in schema.get("allOf") or []:
        required = ((guard.get("not") or {}).get("required")) or []
        require(len(required) == 1, f"forbidden guard must contain exactly one field: {guard}")
        fields.add(required[0])
    return fields


def validator_for(schema: dict[str, Any]) -> Draft202012Validator:
    Draft202012Validator.check_schema(schema)
    return Draft202012Validator(schema)


def is_valid(validator: Draft202012Validator, candidate: dict[str, Any]) -> bool:
    return not list(validator.iter_errors(candidate))


def contract_required_fields(contract: dict[str, Any]) -> set[str]:
    fields: set[str] = set()
    fields.update((contract.get("input_envelope") or {}).get("required_fields") or [])
    fields.update((contract.get("result_envelope") or {}).get("required_fields") or [])
    fields.update((contract.get("record_identity") or {}).get("required_fields") or [])
    return fields


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
    require(slice_info.get("table_schema_artifact") == relpath(TABLE_SCHEMA_PATH), "table schema artifact path drifted")
    for field in ("task_card", "platform_topic", "table_schema_artifact"):
        path = str(slice_info.get(field) or "")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "sql_created",
        "ddl_created",
        "database_provider_created",
        "schema_marker_runtime_created",
        "migration_runner_created",
        "storage_adapter_runtime_created",
        "audit_store_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in claims, f"does_not_claim missing {claim}")
    for claim in {
        "table_schema_artifact_materialized",
        "table_schema_positive_fixture_created",
        "table_schema_negative_fixture_created",
        "table_schema_checker_created",
    }:
        require(claim not in claims, f"does_not_claim must not deny delivered artifact: {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"{dependency_id} source status drifted")


def assert_materialization_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("materialization_boundary") or {}
    expected = {
        "status": SLICE_STATUS,
        "materialization_decision": MATERIALIZATION_DECISION,
        "task_card_decision": TASK_CARD_DECISION,
        "table_schema_artifact_path": relpath(TABLE_SCHEMA_PATH),
        "table_schema_version": TABLE_SCHEMA_VERSION,
        "metadata_contract_version": METADATA_CONTRACT_VERSION,
        "artifact_source": "append_only_boundary_and_metadata_contract_only",
        "current_next_dependency": NEXT_DEPENDENCY,
        "runtime_task_card_decision": RUNTIME_TASK_CARD_DECISION,
    }
    for field, value in expected.items():
        require(boundary.get(field) == value, f"materialization_boundary.{field} drifted")
    for field in (
        "table_schema_artifact_materialized_in_this_slice",
        "table_schema_positive_fixture_created_in_this_slice",
        "table_schema_negative_fixtures_created_in_this_slice",
        "table_schema_checker_created_in_this_slice",
        "table_schema_no_secret_material_scan_created_in_this_slice",
        "metadata_contract_compatibility_smoke_created_in_this_slice",
    ):
        require(boundary.get(field) is True, f"{field} must be true")
    for field in EXPECTED_RUNTIME_FALSE_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")
    for fixture_field in (
        "positive_fixture",
        "missing_required_negative_fixture",
        "physical_detail_negative_fixture",
        "secret_material_negative_fixture",
        "additional_properties_negative_fixture",
    ):
        relative_path = str(boundary.get(fixture_field) or "")
        require((REPO_ROOT / relative_path).exists(), f"{fixture_field} missing: {relative_path}")


def assert_schema_contract(schema: dict[str, Any], contract: dict[str, Any], boundary_fixture: dict[str, Any]) -> None:
    require(schema.get("$schema") == "https://json-schema.org/draft/2020-12/schema", "schema draft drifted")
    require(
        schema.get("$id") == "https://radishmind.local/contracts/production-secret-audit-storage-adapter.table-schema.json",
        "schema id drifted",
    )
    require(schema.get("type") == "object", "schema root type drifted")
    require(schema.get("additionalProperties") is False, "schema must reject additionalProperties")
    require(
        schema.get("description", "").find("SQL") >= 0 and schema.get("description", "").find("runtime") >= 0,
        "schema description must state no SQL/runtime boundary",
    )

    properties = schema.get("properties") or {}
    required = set(schema.get("required") or [])
    expected_contract_fields = contract_required_fields(contract)
    require(required == expected_contract_fields | {"table_schema_version"}, "schema required fields drifted")
    require(set(properties) == required, "schema properties must match required fields")
    require(properties.get("table_schema_version", {}).get("const") == TABLE_SCHEMA_VERSION, "version pin drifted")
    require(
        properties.get("storage_adapter_contract_version", {}).get("const") == METADATA_CONTRACT_VERSION,
        "metadata contract version pin drifted",
    )
    require(
        properties.get("backend_product_class", {}).get("enum") == ["managed_database_append_only_table"],
        "backend product class must remain static and non-vendor",
    )
    require(
        properties.get("write_status", {}).get("enum")
        == (contract.get("result_envelope") or {}).get("write_status_allowlist"),
        "write_status allowlist drifted from metadata contract",
    )

    extension = schema.get("x-radishmind-logical-schema") or {}
    require(extension.get("schema_version") == TABLE_SCHEMA_VERSION, "extension schema version drifted")
    require(extension.get("metadata_only") is True, "extension metadata-only flag drifted")
    require(extension.get("physical_table_details_allowed") is False, "physical table details must stay forbidden")
    require(extension.get("sql_or_ddl_allowed") is False, "SQL/DDL must stay forbidden")
    require(extension.get("runtime_created") is False, "schema artifact must not create runtime")

    boundary_groups = {
        str(group.get("id") or "") for group in (boundary_fixture.get("logical_schema_boundary") or {}).get("allowed_field_groups") or []
    }
    groups = rows_by_id(extension, "logical_field_groups", "id")
    require(set(groups) == EXPECTED_GROUPS, "logical group ids drifted")
    require(set(groups) == boundary_groups, "logical group ids drifted from append-only boundary")
    grouped_fields: set[str] = set()
    for group in groups.values():
        fields = set(group.get("logical_fields") or [])
        require(fields, f"{group.get('id')} logical fields missing")
        require(fields <= expected_contract_fields, f"{group.get('id')} contains non-contract fields")
        grouped_fields.update(fields)
    require(grouped_fields == expected_contract_fields, "logical groups must cover metadata contract fields exactly")

    forbidden = forbidden_fields_from_schema(schema)
    require(EXPECTED_FORBIDDEN_FIELDS <= forbidden, "schema forbidden guards missing expected fields")
    require(not (EXPECTED_FORBIDDEN_FIELDS & set(properties)), "forbidden fields must not appear as allowed properties")


def assert_validation_cases(fixture: dict[str, Any], schema: dict[str, Any]) -> None:
    validator = validator_for(schema)
    expected_cases = {
        "positive_metadata_only_table_schema": (POSITIVE_FIXTURE, True),
        "missing_required_storage_record_ref": (MISSING_REQUIRED_FIXTURE, False),
        "physical_table_detail_forbidden": (PHYSICAL_DETAIL_FIXTURE, False),
        "secret_material_forbidden": (SECRET_MATERIAL_FIXTURE, False),
        "additional_property_forbidden": (ADDITIONAL_PROPERTIES_FIXTURE, False),
    }
    cases = rows_by_id(fixture, "validation_case_matrix", "case")
    require(set(cases) == set(expected_cases), "validation case matrix drifted")
    for case_id, (relative_path, expected_valid) in expected_cases.items():
        case = cases[case_id]
        require(case.get("fixture") == relative_path, f"{case_id} fixture path drifted")
        require(case.get("expected_valid") is expected_valid, f"{case_id} expected validity drifted")
        candidate = load_json(REPO_ROOT / relative_path)
        actual = is_valid(validator, candidate)
        require(actual is expected_valid, f"{case_id} validation result drifted")

    positive = load_json(REPO_ROOT / POSITIVE_FIXTURE)
    for field in schema.get("required") or []:
        candidate = dict(positive)
        candidate.pop(field, None)
        require(not is_valid(validator, candidate), f"schema allowed missing required field: {field}")
    for field in EXPECTED_FORBIDDEN_FIELDS:
        candidate = dict(positive)
        candidate[field] = "redacted-not-accepted"
        require(not is_valid(validator, candidate), f"schema allowed forbidden field: {field}")


def assert_metadata_contract_compatibility(fixture: dict[str, Any], contract: dict[str, Any]) -> None:
    smoke = fixture.get("metadata_contract_compatibility_smoke") or {}
    require(smoke.get("status") == "implemented_static_contract_compatibility", "compatibility smoke status drifted")
    require(smoke.get("positive_fixture") == POSITIVE_FIXTURE, "compatibility positive fixture drifted")
    required = contract_required_fields(contract)
    require(set(smoke.get("requires_fields") or []) == required, "compatibility required fields drifted")
    positive = load_json(REPO_ROOT / POSITIVE_FIXTURE)
    require(set(positive) == required | {"table_schema_version"}, "positive fixture must match table schema required fields")
    require(positive.get("storage_adapter_contract_version") == METADATA_CONTRACT_VERSION, "contract version drifted")
    require(
        positive.get("write_status") in set((contract.get("result_envelope") or {}).get("write_status_allowlist") or []),
        "positive fixture write_status not allowed by metadata contract",
    )
    require(smoke.get("storage_adapter_runtime_created_in_this_slice") is False, "compatibility smoke created runtime")
    require(smoke.get("audit_store_runtime_created_in_this_slice") is False, "compatibility smoke created audit store runtime")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "failure_mapping", "code")
    require(set(rows) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in rows.items():
        require(item.get("failure_boundary"), f"{code} missing failure boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} missing diagnostic")
        require("secret value" not in diagnostic.lower(), f"{code} diagnostic must stay sanitized")


def assert_no_secret_material_scan(fixture: dict[str, Any]) -> None:
    scan = fixture.get("no_secret_material_scan") or {}
    require(scan.get("status") == "implemented_static_scan", "no secret material scan status drifted")
    scanned = set(scan.get("scanned_artifacts") or [])
    expected = {relpath(TABLE_SCHEMA_PATH), POSITIVE_FIXTURE, *EXPECTED_NEGATIVE_FIXTURES}
    require(expected <= scanned, "no secret scan target list missing expected artifacts")
    for relative_path in scanned:
        text = read(str(relative_path))
        for pattern in SECRET_LITERAL_PATTERNS:
            require(not pattern.search(text), f"secret-like literal found in {relative_path}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    allowed = set(guard.get("allowed_added_artifacts") or [])
    require(EXPECTED_ALLOWED_ARTIFACTS <= allowed, "allowed added artifacts missing expected paths")
    for path in allowed:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing: {path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "database_product_selection_artifact",
        "database_vendor_selection_artifact",
        "db_driver",
        "dsn_parser",
        "database_connection_provider",
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
        require(artifact in forbidden, f"forbidden artifact kind missing: {artifact}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden runtime artifact exists: {relative_path}")


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
    require(boundary.get("durable_audit_backend_status") == CURRENT_MATRIX_BLOCKER_STATUS, "matrix durable blocker drifted")
    require(boundary.get("storage_adapter_current_next_dependency") == CURRENT_NEXT_DEPENDENCY, "matrix next dependency drifted")
    require(
        boundary.get("storage_adapter_table_schema_artifact_materialization_status") == SLICE_STATUS,
        "matrix materialization status drifted",
    )
    require(
        boundary.get("storage_adapter_table_schema_artifact_status") == "materialized_static_logical_table_schema",
        "matrix table schema artifact status drifted",
    )
    require(boundary.get("storage_adapter_sql_migration_status") == "not_created", "matrix SQL drifted")
    require(boundary.get("storage_adapter_ddl_status") == "not_created", "matrix DDL drifted")
    require(
        boundary.get("storage_adapter_runtime_task_card_decision") == CURRENT_RUNTIME_TASK_CARD_DECISION,
        "matrix runtime task decision drifted",
    )
    require(
        alignment.get("durable_backend_blocker_status_after_table_schema_artifact") == CURRENT_MATRIX_BLOCKER_STATUS,
        "fixture blocker status drifted",
    )
    require(
        alignment.get("storage_adapter_current_next_dependency") == CURRENT_NEXT_DEPENDENCY,
        "fixture next dependency drifted",
    )
    require(
        alignment.get("runtime_task_card_decision") == CURRENT_RUNTIME_TASK_CARD_DECISION,
        "fixture runtime task card decision drifted",
    )
    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == CURRENT_MATRIX_BLOCKER_STATUS, "durable blocker row status drifted")
    require(
        durable.get("source")
        == "production-secret-backend-audit-store-storage-adapter-offline-adapter-smoke-strategy-readiness-v1",
        "durable blocker row source drifted",
    )


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in (fixture.get("implementation_readiness_alignment") or {}).items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")
    planned = rows_by_id(readiness, "planned_slices", "id")
    item = planned.get("audit-store-storage-adapter-table-schema-artifact-materialization") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness planned slice status drifted")
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        "contracts/README.md": [
            relpath(TABLE_SCHEMA_PATH),
            "Production Secret Audit Storage Adapter Table Schema",
            SLICE_STATUS,
        ],
        "docs/contracts/README.md": [
            "Production Secret Audit Storage Adapter Table Schema",
            "production-secret-audit-storage-adapter-table-schema.md",
        ],
        "docs/contracts/production-secret-audit-storage-adapter-table-schema.md": [
            TABLE_SCHEMA_VERSION,
            POSITIVE_FIXTURE,
            PHYSICAL_DETAIL_FIXTURE,
            "不创建 SQL",
        ],
        "docs/platform/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.md": [
            SLICE_STATUS,
            TABLE_SCHEMA_VERSION,
            NEXT_DEPENDENCY,
        ],
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1-plan.md": [
            SLICE_STATUS,
            relpath(TABLE_SCHEMA_PATH),
            NEXT_DEPENDENCY,
        ],
        "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": [
            SLICE_STATUS,
            NEXT_DEPENDENCY,
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            SLICE_STATUS,
            MATRIX_BLOCKER_STATUS,
            NEXT_DEPENDENCY,
        ],
        "docs/radishmind-current-focus.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/features/README.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/features/workflow/README.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/radishmind-integration-contracts.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/radishmind-architecture.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/radishmind-product-scope.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
        "docs/task-cards/production-secret-backend-audit-store-runtime-blocker-matrix-v1-plan.md": [
            SLICE_STATUS,
            MATRIX_BLOCKER_STATUS,
            NEXT_DEPENDENCY,
        ],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-table-schema-artifact-materialization-v1.py",
            SLICE_STATUS,
        ],
        "docs/devlogs/2026-W27.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    }
    for relative_path, literals in docs.items():
        text = read(relative_path)
        missing = sorted(literal for literal in literals if literal not in text)
        require(not missing, f"{relative_path} missing literals: {missing}")

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
    schema = load_json(TABLE_SCHEMA_PATH)
    contract = load_json(METADATA_CONTRACT_ARTIFACT_PATH)
    boundary = load_json(APPEND_ONLY_BOUNDARY_FIXTURE_PATH)

    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_materialization_boundary(fixture)
    assert_schema_contract(schema, contract, boundary)
    assert_validation_cases(fixture, schema)
    assert_metadata_contract_compatibility(fixture, contract)
    assert_failure_mapping(fixture)
    assert_no_secret_material_scan(fixture)
    assert_artifact_guard(fixture)
    assert_entry_review_still_static()
    assert_blocker_matrix_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    print("production ops secret backend audit store storage adapter table schema artifact checks passed.")


if __name__ == "__main__":
    main()
