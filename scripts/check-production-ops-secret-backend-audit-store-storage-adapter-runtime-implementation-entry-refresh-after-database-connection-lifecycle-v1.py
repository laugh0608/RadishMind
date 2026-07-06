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
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-database-connection-lifecycle-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

SLICE_ID = (
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-database-connection-lifecycle-v1"
)
SLICE_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_defined"
)
ENTRY_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_entry_refresh"
NEXT_DEPENDENCY = "storage_adapter_database_provider_connection_runtime_boundary_readiness"
MATRIX_BLOCKER_STATUS = (
    "storage_adapter_runtime_entry_refresh_after_database_connection_lifecycle_defined_task_card_blocked"
)
CURRENT_ENTRY_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_database_provider_connection_runtime_boundary_readiness"
CURRENT_NEXT_DEPENDENCY = (
    "storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_readiness"
)
CURRENT_MATRIX_BLOCKER_STATUS = (
    "storage_adapter_database_provider_connection_runtime_boundary_readiness_defined_task_card_blocked"
)
CURRENT_MATRIX_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1"
)
SELECTED_DRIVER_CANDIDATE = "github.com/jackc/pgx/v5"
SELECTED_ENGINE = "postgresql_compatible_append_only_relational_database"
SELECTED_PROVIDER_CLASS = "managed_postgresql_compatible_service"

EXPECTED_DEPENDENCIES = {
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
    "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-concrete-database-selection-review-v1.json"
        ),
        "audit_store_storage_adapter_concrete_database_selection_review_defined",
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

EXPECTED_BOUNDARY = {
    "status": SLICE_STATUS,
    "entry_decision": ENTRY_DECISION,
    "previous_database_connection_lifecycle_readiness_status": (
        "audit_store_storage_adapter_database_connection_lifecycle_readiness_defined"
    ),
    "previous_database_connection_lifecycle_decision": (
        "database_connection_lifecycle_readiness_defined_without_connection_runtime"
    ),
    "previous_runtime_task_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_readiness"
    ),
    "selected_database_engine": SELECTED_ENGINE,
    "selected_provider_candidate_class": SELECTED_PROVIDER_CLASS,
    "selected_database_driver_candidate": SELECTED_DRIVER_CANDIDATE,
    "database_driver_status": "selected_reference_only",
    "database_driver_import_status": "not_created",
    "driver_dependency_version_status": "not_pinned",
    "secret_ref_only_dsn_handoff_status": "secret_ref_only_dsn_handoff_defined",
    "tls_role_environment_binding_status": "static_tls_role_environment_binding_defined",
    "pool_policy_status": "static_pool_policy_defined_without_pool_runtime",
    "timeout_budget_status": "static_timeout_budget_defined_without_runtime_timer",
    "retry_transaction_recovery_status": "static_recovery_policy_defined_without_runtime",
    "duplicate_replay_fail_closed_status": "duplicate_replay_fail_closed_policy_defined",
    "sanitized_diagnostics_status": "sanitized_diagnostics_allowlist_defined",
    "schema_marker_migration_handoff_status": "schema_marker_migration_handoff_defined",
    "offline_verification_status": "metadata_only_offline_verification_defined",
    "negative_leakage_scan_status": "defined_without_runtime",
    "rollback_rollout_boundary_status": "metadata_only_rollout_rollback_boundary_defined",
    "database_provider_connection_runtime_boundary_status": "required_before_runtime_task_card",
    "next_dependency": NEXT_DEPENDENCY,
    "database_connection_provider_status": "not_created",
    "database_connection_factory_status": "not_created",
    "database_connection_lifecycle_runtime_status": "not_created",
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
    "database_provider_connection_runtime_boundary_created_in_this_slice",
    "database_connection_provider_created_in_this_slice",
    "database_connection_factory_created_in_this_slice",
    "database_pool_runtime_created_in_this_slice",
    "database_health_check_runtime_created_in_this_slice",
    "sql_created_in_this_slice",
    "ddl_created_in_this_slice",
    "physical_table_schema_created_in_this_slice",
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

EXPECTED_BLOCKERS = {
    "database_provider_connection_runtime_boundary_readiness",
    "connection_provider_factory_pool_health_runtime",
    "migration_schema_marker_runtime_boundary",
    "physical_sql_ddl_boundary",
    "peer_runtime_dependencies",
    "audit_store_runtime_gate",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_runtime_refresh_after_connection_lifecycle_dependency_missing",
    "audit_store_storage_adapter_runtime_refresh_after_connection_lifecycle_task_card_still_blocked",
    "audit_store_storage_adapter_runtime_refresh_after_connection_lifecycle_provider_connection_boundary_required",
    "audit_store_storage_adapter_runtime_refresh_after_connection_lifecycle_runtime_forbidden",
    "audit_store_storage_adapter_runtime_refresh_after_connection_lifecycle_secret_material_detected",
    "audit_store_storage_adapter_runtime_refresh_after_connection_lifecycle_scope_overreach",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    (
        "docs/platform/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
        "after-database-connection-lifecycle-v1.md"
    ),
    (
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
        "after-database-connection-lifecycle-v1-plan.md"
    ),
    (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-"
        "refresh-after-database-connection-lifecycle-v1.json"
    ),
    (
        "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
        "after-database-connection-lifecycle-v1.py"
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


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == SLICE_STATUS, "unexpected status")
    require(slice_info.get("entry_decision") == ENTRY_DECISION, "unexpected entry decision")
    require(slice_info.get("next_dependency") == NEXT_DEPENDENCY, "unexpected next dependency")
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "database_provider_connection_runtime_ready",
        "database_connection_provider_created",
        "database_connection_created",
        "database_connection_factory_created",
        "database_pool_runtime_created",
        "database_health_check_runtime_created",
        "dsn_parser_created",
        "sql_created",
        "ddl_created",
        "physical_table_schema_created",
        "schema_marker_runtime_created",
        "migration_runner_created",
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


def assert_entry_refresh_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_refresh_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"entry_refresh_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"entry_refresh_boundary.{field} must stay false")


def assert_remaining_blockers(fixture: dict[str, Any]) -> None:
    blockers = rows_by_id(fixture, "remaining_blockers", "id")
    require(set(blockers) == EXPECTED_BLOCKERS, "remaining blocker ids drifted")
    primary = blockers["database_provider_connection_runtime_boundary_readiness"]
    require(primary.get("status") == "required_before_runtime_task_card", "primary blocker status drifted")
    require(primary.get("next_dependency") == NEXT_DEPENDENCY, "primary blocker next dependency drifted")
    for blocker_id in EXPECTED_BLOCKERS - {"database_provider_connection_runtime_boundary_readiness"}:
        require(blockers[blocker_id].get("status") == "blocked", f"{blocker_id} status drifted")
    require(len(blockers["peer_runtime_dependencies"].get("requires") or []) >= 7, "peer list too small")
    require(len(blockers["audit_store_runtime_gate"].get("requires") or []) >= 4, "audit gate list too small")


def assert_diagnostics_failures_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(not (allowed & forbidden), "diagnostic allowlist intersects forbidden fields")
    require(sample.get("entry_decision") == ENTRY_DECISION, "diagnostic decision drifted")
    require(sample.get("next_dependency") == NEXT_DEPENDENCY, "diagnostic next dependency drifted")
    require(sample.get("selected_database_engine") == SELECTED_ENGINE, "diagnostic engine drifted")
    require(sample.get("selected_provider_candidate_class") == SELECTED_PROVIDER_CLASS, "diagnostic provider class drifted")
    require(sample.get("database_driver_status") == "selected_reference_only", "diagnostic driver status drifted")
    require(
        sample.get("database_provider_connection_runtime_boundary_status") == "required_before_runtime_task_card",
        "diagnostic provider connection boundary status drifted",
    )
    require(sample.get("database_connection_provider_status") == "not_created", "diagnostic provider created")
    require(sample.get("storage_adapter_runtime_status") == "not_created", "diagnostic created runtime")

    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} boundary missing")
        require(item.get("sanitized_diagnostic"), f"{code} diagnostic missing")

    no_fallback = fixture.get("no_fallback_policy") or {}
    require(no_fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for source in {
        "reference_only_driver_candidate",
        "static_connection_lifecycle_readiness",
        "static_table_schema_artifact",
        "metadata_contract_artifact",
        "offline_adapter_smoke_strategy",
        "negative_leakage_runtime_scan_boundary",
        "historical_checker_success",
        "test_only_fake_resolver",
        "memory_store",
        "mock_database_provider",
        "developer_environment",
    }:
        require(source in set(no_fallback.get("forbidden_sources") or []), f"missing forbidden fallback {source}")

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for counter, value in counters.items():
        require(value == 0, f"{counter} must stay 0")


def assert_no_secret_material_scan(fixture: dict[str, Any]) -> None:
    scan = fixture.get("no_secret_material_scan") or {}
    require(scan.get("status") == "implemented_static_scan", "no secret material scan status drifted")
    expected = {
        "docs/platform/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1.json",
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
        "go_dependency_change",
        "go_import",
        "database_provider_connection_runtime_boundary",
        "database_connection_provider",
        "database_connection_factory",
        "database_pool_runtime",
        "database_health_check_runtime",
        "database_connection",
        "dsn_parser",
        "sql_migration",
        "ddl",
        "physical_table_schema",
        "schema_marker_runtime",
        "migration_runner",
        "storage_adapter_runtime_implementation_task_card",
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
        "storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_status": SLICE_STATUS,
        "storage_adapter_runtime_task_card_decision": CURRENT_ENTRY_DECISION,
        "storage_adapter_current_next_dependency": CURRENT_NEXT_DEPENDENCY,
        "storage_adapter_database_provider_connection_runtime_boundary_status": "metadata_only_boundary_defined_without_runtime",
        "storage_adapter_database_connection_provider_status": "not_created",
        "storage_adapter_connection_lifecycle_runtime_status": "not_created",
        "storage_adapter_connection_factory_status": "not_created",
        "storage_adapter_pool_runtime_status": "not_created",
        "storage_adapter_health_check_runtime_status": "not_created",
        "storage_adapter_runtime_task_card_status": "not_created",
        "storage_adapter_runtime_status": "not_created",
        "durable_audit_backend_status": CURRENT_MATRIX_BLOCKER_STATUS,
    }.items():
        require(boundary.get(field) == value, f"matrix boundary {field} drifted")

    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == CURRENT_MATRIX_BLOCKER_STATUS, "durable status drifted")
    require(durable.get("source") == CURRENT_MATRIX_BLOCKER_SOURCE, "durable source drifted")
    require(durable.get("unlock_condition") == CURRENT_NEXT_DEPENDENCY, "durable unlock condition drifted")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable must block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable must block resolver runtime")

    alignment = fixture.get("blocker_matrix_alignment") or {}
    require(
        alignment.get("durable_backend_blocker_status_after_entry_refresh") == CURRENT_MATRIX_BLOCKER_STATUS,
        "fixture matrix status drifted",
    )
    require(
        alignment.get("durable_backend_blocker_source_after_entry_refresh") == CURRENT_MATRIX_BLOCKER_SOURCE,
        "fixture matrix source drifted",
    )
    require(alignment.get("storage_adapter_current_next_dependency") == CURRENT_NEXT_DEPENDENCY, "fixture next drifted")
    require(alignment.get("runtime_task_card_decision") == CURRENT_ENTRY_DECISION, "fixture decision drifted")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    expected = fixture.get("implementation_readiness_alignment") or {}
    for field, value in expected.items():
        require(target.get(field) == value, f"implementation readiness {field} drifted")

    planned = rows_by_id(readiness, "planned_slices", "id")
    item = planned.get("audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle") or {}
    require(item.get("status") == SLICE_STATUS, "implementation readiness missing planned slice")
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        (
            "docs/platform/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
            "after-database-connection-lifecycle-v1.md"
        ): [SLICE_STATUS, ENTRY_DECISION, NEXT_DEPENDENCY, "storage_adapter_database_provider_connection_runtime_boundary_readiness"],
        (
            "docs/task-cards/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
            "after-database-connection-lifecycle-v1-plan.md"
        ): [SLICE_ID, SLICE_STATUS, ENTRY_DECISION, NEXT_DEPENDENCY, "停止线"],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            SLICE_STATUS,
            MATRIX_BLOCKER_STATUS,
            NEXT_DEPENDENCY,
        ],
        "docs/platform/production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md": [
            SLICE_STATUS,
            ENTRY_DECISION,
            NEXT_DEPENDENCY,
        ],
        "docs/platform/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Runtime Implementation Entry Refresh After Database Connection Lifecycle v1",
            SLICE_STATUS,
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Runtime Implementation Entry Refresh After Database Connection Lifecycle v1",
            SLICE_STATUS,
        ],
        "docs/features/workflow/README.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/features/workflow/saved-workflow-draft-v1.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/radishmind-current-focus.md": [SLICE_STATUS, NEXT_DEPENDENCY],
        "docs/task-cards/README.md": [SLICE_ID, SLICE_STATUS],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [SLICE_ID, SLICE_STATUS],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1.py",
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1.json",
        ],
        "docs/devlogs/2026-W27.md": [SLICE_STATUS, NEXT_DEPENDENCY],
    }
    for path, literals in docs.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    lifecycle = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-"
        "database-connection-lifecycle-readiness-v1.py"
    )
    current = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
        "after-database-connection-lifecycle-v1.py"
    )
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    for script in (lifecycle, current, matrix):
        require(script in check_repo, f"check-repo.py missing {script}")
    require(
        check_repo.index(lifecycle) < check_repo.index(current) < check_repo.index(matrix),
        "check-repo.py registration order drifted",
    )


def assert_no_runtime_or_dependency_imports() -> None:
    go_mod = REPO_ROOT / "go.mod"
    if go_mod.exists():
        require(SELECTED_DRIVER_CANDIDATE not in go_mod.read_text(encoding="utf-8"), "go.mod must not pin pgx")
    go_sum = REPO_ROOT / "go.sum"
    if go_sum.exists():
        require(SELECTED_DRIVER_CANDIDATE not in go_sum.read_text(encoding="utf-8"), "go.sum must not pin pgx")
    services_root = REPO_ROOT / "services"
    go_files = services_root.rglob("*.go") if services_root.exists() else []
    for go_file in go_files:
        require(SELECTED_DRIVER_CANDIDATE not in go_file.read_text(encoding="utf-8"), f"pgx import found: {go_file}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_refresh_boundary(fixture)
    assert_remaining_blockers(fixture)
    assert_diagnostics_failures_and_policies(fixture)
    assert_no_secret_material_scan(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_runtime_or_dependency_imports()
    print("production secret audit store storage adapter runtime entry refresh after connection lifecycle v1 check passed")


if __name__ == "__main__":
    main()
