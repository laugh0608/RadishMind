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
    "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = (
    "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-"
    "readiness-v1"
)
SLICE_STATUS = "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined"
READINESS_DECISION = "database_provider_driver_dsn_tls_role_policy_defined_without_runtime"
NEXT_DEPENDENCY = "storage_adapter_append_only_table_schema_boundary_readiness"
SELECTED_PRODUCT_CLASS = "managed_database_append_only_table"
SELECTED_PRODUCT_PROFILE = "reserved_managed_database_append_only_table_profile"
MATRIX_BLOCKER_STATUS = "storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_defined_task_card_blocked"
CURRENT_MATRIX_BLOCKER_STATUS = (
    "storage_adapter_database_provider_selection_review_defined_task_card_blocked"
)
CURRENT_MATRIX_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-database-provider-selection-review-v1"
)
CURRENT_NEXT_DEPENDENCY = "storage_adapter_database_driver_selection_readiness"
CURRENT_RUNTIME_TASK_CARD_DECISION = (
    "storage_adapter_runtime_task_card_still_blocked_after_database_provider_selection_review"
)

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.json"
        ),
        "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_backend_product_selection_review_defined",
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
    "previous_entry_refresh_status": "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined",
    "previous_entry_decision": "storage_adapter_runtime_task_card_still_blocked_after_product_selection",
    "backend_product_selection_review_status": "audit_store_storage_adapter_backend_product_selection_review_defined",
    "backend_product_selection_status": "selected_static_product_class_without_backend_provider",
    "selected_backend_product_class": SELECTED_PRODUCT_CLASS,
    "selected_backend_product_profile": SELECTED_PRODUCT_PROFILE,
    "metadata_contract_artifact_status": "audit_store_storage_adapter_metadata_contract_artifact_materialized",
    "database_product_status": "not_selected",
    "database_vendor_status": "not_selected",
    "database_provider_boundary_status": "metadata_only_provider_boundary_defined",
    "database_driver_selection_policy_status": "static_driver_policy_defined_without_driver_selection",
    "database_dsn_secret_ref_policy_status": "secret_ref_only_dsn_policy_defined",
    "database_tls_policy_status": "tls_mode_policy_defined",
    "database_role_policy_status": "least_privilege_role_policy_defined",
    "database_connection_provider_status": "not_created",
    "database_driver_status": "not_selected",
    "database_dsn_status": "not_defined",
    "database_connection_status": "not_created",
    "sql_migration_status": "not_created",
    "schema_marker_status": "not_created",
    "append_only_table_schema_boundary_status": "required_before_runtime_task_card",
    "migration_schema_marker_boundary_status": "required_before_runtime_task_card",
    "offline_adapter_smoke_strategy_status": "required_before_runtime_task_card",
    "negative_leakage_runtime_scan_boundary_status": "defined_without_runtime",
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
    "database_product_selected_in_this_slice",
    "database_vendor_selected_in_this_slice",
    "database_driver_selected_in_this_slice",
    "database_provider_created_in_this_slice",
    "database_connection_provider_enabled",
    "database_connection_created_in_this_slice",
    "dsn_defined_in_this_slice",
    "tls_config_created_in_this_slice",
    "database_role_created_in_this_slice",
    "sql_migration_created_in_this_slice",
    "schema_marker_created_in_this_slice",
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

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_dependency_missing",
    "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_database_selection_forbidden",
    "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_provider_runtime_forbidden",
    "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_secret_material_detected",
    "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_schema_scope_overreach",
    "audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_runtime_scope_overreach",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    (
        "docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-"
        "policy-readiness-v1.md"
    ),
    (
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-"
        "policy-readiness-v1-plan.md"
    ),
    (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-"
        "tls-role-policy-readiness-v1.json"
    ),
    (
        "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-"
        "role-policy-readiness-v1.py"
    ),
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
        == "production_ops_secret_backend_audit_store_storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected status")
    require(slice_info.get("readiness_decision") == READINESS_DECISION, "unexpected readiness decision")
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "database_product_selected",
        "database_vendor_selected",
        "database_driver_selected",
        "database_provider_created",
        "database_connection_created",
        "dsn_defined",
        "storage_adapter_runtime_task_card_created",
        "storage_adapter_runtime_created",
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "production_resolver_runtime_task_card_created",
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


def assert_provider_driver_dsn_tls_role_policy(fixture: dict[str, Any]) -> None:
    provider = fixture.get("provider_boundary") or {}
    require(provider.get("status") == "metadata_only_provider_boundary_defined", "provider boundary status drifted")
    for field in {"database_secret_ref", "tls_policy_ref", "role_policy_ref", "audit_ref", "policy_version"}:
        require(field in set(provider.get("allowed_input_references") or []), f"provider input missing {field}")
    for forbidden in {"raw_dsn", "password", "provider_raw_url", "database_hostname", "credential_payload"}:
        require(forbidden in set(provider.get("forbidden_inputs") or []), f"provider forbidden input missing {forbidden}")

    driver = fixture.get("driver_selection_policy") or {}
    require(
        driver.get("status") == "static_driver_policy_defined_without_driver_selection",
        "driver selection policy status drifted",
    )
    for capability in {
        "append_only_transaction_support",
        "parameterized_statement_support",
        "tls_mode_support",
        "least_privilege_role_support",
        "sanitized_error_mapping",
    }:
        require(capability in set(driver.get("required_capabilities") or []), f"driver capability missing {capability}")
    for claim in {"specific_driver_selected", "driver_open_path_created", "connection_pool_created"}:
        require(claim in set(driver.get("forbidden_claims") or []), f"driver forbidden claim missing {claim}")

    dsn = fixture.get("dsn_secret_ref_policy") or {}
    require(dsn.get("status") == "secret_ref_only_dsn_policy_defined", "DSN secret-ref policy drifted")
    for material in {"raw_dsn", "hostname", "username", "password", "database_name", "credential_payload"}:
        require(material in set(dsn.get("forbidden_material") or []), f"DSN forbidden material missing {material}")

    tls_role = fixture.get("tls_and_role_policy") or {}
    require(tls_role.get("tls_policy_status") == "tls_mode_policy_defined", "TLS policy drifted")
    require(tls_role.get("role_policy_status") == "least_privilege_role_policy_defined", "role policy drifted")
    for role in {"append_only_writer_role", "read_only_verifier_role", "migration_marker_role"}:
        require(role in set(tls_role.get("required_role_boundaries") or []), f"role boundary missing {role}")
    for privilege in {"update_record", "delete_record", "truncate_table", "drop_schema", "bypass_audit_policy"}:
        require(privilege in set(tls_role.get("forbidden_privileges") or []), f"forbidden privilege missing {privilege}")


def assert_diagnostics_failures_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(not (allowed & forbidden), "diagnostic allowlist intersects forbidden fields")
    require(sample.get("readiness_decision") == READINESS_DECISION, "diagnostic decision drifted")
    require(sample.get("next_dependency") == NEXT_DEPENDENCY, "diagnostic next dependency drifted")
    require(sample.get("selected_backend_product_class") == SELECTED_PRODUCT_CLASS, "diagnostic product class drifted")
    require(sample.get("provider_boundary_status") == "metadata_only_provider_boundary_defined", "provider diagnostic drifted")
    require(sample.get("driver_selection_policy_status") == "static_driver_policy_defined_without_driver_selection", "driver diagnostic drifted")
    require(sample.get("storage_adapter_runtime_status") == "not_created", "diagnostic created runtime")

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
        "writer_compatibility_fixture",
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


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    require(set(guard.get("allowed_added_artifacts") or []) == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifacts drifted")
    for path in EXPECTED_ALLOWED_ARTIFACTS:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing: {path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "database_product_selection_artifact",
        "backend_vendor_selection_artifact",
        "database_provider_implementation_task_card",
        "storage_adapter_runtime_implementation_task_card",
        "database_connection_provider",
        "database_driver",
        "dsn_parser",
        "connection_factory",
        "sql_migration",
        "schema_marker",
        "storage_adapter_runtime",
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
        "storage_adapter_database_provider_driver_dsn_tls_role_policy_readiness_status": SLICE_STATUS,
        "storage_adapter_database_provider_boundary_status": "metadata_only_provider_boundary_defined",
        "storage_adapter_database_driver_selection_policy_status": "static_driver_policy_defined_without_driver_selection",
        "storage_adapter_database_dsn_secret_ref_policy_status": "secret_ref_only_dsn_policy_defined",
        "storage_adapter_database_tls_policy_status": "tls_mode_policy_defined",
        "storage_adapter_database_role_policy_status": "least_privilege_role_policy_defined",
        "storage_adapter_database_connection_provider_status": "not_created",
        "storage_adapter_database_provider_driver_dsn_tls_role_policy_status": "defined_without_runtime",
        "storage_adapter_current_next_dependency": CURRENT_NEXT_DEPENDENCY,
        "storage_adapter_append_only_table_schema_boundary_readiness_status": (
            "audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined"
        ),
        "storage_adapter_append_only_table_schema_boundary_status": "defined_without_sql_or_runtime",
        "storage_adapter_runtime_task_card_status": "not_created",
        "storage_adapter_runtime_status": "not_created",
    }.items():
        require(boundary.get(field) == value, f"matrix boundary {field} drifted")

    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == CURRENT_MATRIX_BLOCKER_STATUS, "durable status drifted")
    require(durable.get("source") == CURRENT_MATRIX_BLOCKER_SOURCE, "durable source drifted")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable must block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable must block resolver runtime")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    expected = fixture.get("implementation_readiness_alignment") or {}
    for field, value in expected.items():
        if field == "status":
            continue
        if field in {
            "audit_storage_adapter_runtime_task_card_decision",
            "audit_storage_adapter_current_next_dependency",
            "audit_storage_adapter_append_only_table_schema_boundary_status",
            "audit_storage_adapter_migration_schema_marker_boundary_status",
        }:
            continue
        require(target.get(field) == value, f"implementation readiness {field} drifted")
    require(
        target.get("audit_store_storage_adapter_append_only_table_schema_boundary_readiness_status")
        == "audit_store_storage_adapter_append_only_table_schema_boundary_readiness_defined",
        "implementation readiness append-only table schema boundary status drifted",
    )
    require(
        target.get("audit_storage_adapter_runtime_task_card_decision") == CURRENT_RUNTIME_TASK_CARD_DECISION,
        "implementation readiness current runtime decision drifted",
    )
    require(
        target.get("audit_storage_adapter_current_next_dependency") == CURRENT_NEXT_DEPENDENCY,
        "implementation readiness current next dependency drifted",
    )
    require(
        target.get("audit_storage_adapter_append_only_table_schema_boundary_status")
        == "defined_without_sql_or_runtime",
        "implementation readiness append-only table schema boundary drifted",
    )
    require(
        target.get("audit_storage_adapter_migration_schema_marker_boundary_status")
        == "logical_schema_marker_handoff_boundary_defined",
        "implementation readiness migration schema marker boundary drifted",
    )

    planned = {str(row.get("id")): row for row in readiness.get("planned_slices") or [] if isinstance(row, dict)}
    item = planned.get("audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness missing planned slice")
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        (
            "docs/platform/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-"
            "policy-readiness-v1.md"
        ): [
            SLICE_STATUS,
            READINESS_DECISION,
            NEXT_DEPENDENCY,
            "metadata_only_provider_boundary_defined",
            "secret_ref_only_dsn_policy_defined",
            "not_created",
        ],
        (
            "docs/task-cards/production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-"
            "role-policy-readiness-v1-plan.md"
        ): [
            SLICE_ID,
            SLICE_STATUS,
            READINESS_DECISION,
            NEXT_DEPENDENCY,
            "停止线",
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
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
            "Production Secret Backend Audit Store Storage Adapter Database Provider / Driver / DSN / TLS / Role Policy Readiness v1",
            SLICE_STATUS,
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Database Provider / Driver / DSN / TLS / Role Policy Readiness v1",
            SLICE_STATUS,
        ],
        "docs/features/workflow/README.md": [SLICE_STATUS],
        "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/radishmind-current-focus.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/radishmind-integration-contracts.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [SLICE_ID, SLICE_STATUS],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.py",
            "production-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-policy-readiness-v1.json",
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
        "after-product-selection-v1.py"
    )
    current = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-database-provider-driver-dsn-tls-role-"
        "policy-readiness-v1.py"
    )
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
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "-----BEGIN", "authorization:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"database readiness contains forbidden literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "credential-bearing DSN-like literal found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_readiness_boundary(fixture)
    assert_provider_driver_dsn_tls_role_policy(fixture)
    assert_diagnostics_failures_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print(
        "production ops secret backend audit store storage adapter database provider driver DSN TLS role policy readiness checks passed."
    )


if __name__ == "__main__":
    main()
