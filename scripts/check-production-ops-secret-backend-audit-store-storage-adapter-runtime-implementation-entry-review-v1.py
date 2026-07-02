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
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.json"
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
FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY = "storage_adapter_table_schema_artifact_materialization_entry_review"
FOLLOWUP_AFTER_SELECTION_BLOCKER_STATUS = (
    "storage_adapter_append_only_table_schema_boundary_readiness_defined_task_card_blocked"
)
FOLLOWUP_AFTER_SELECTION_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-append-only-table-schema-boundary-readiness-v1"
)
FOLLOWUP_ALIGNMENT = {
    "audit_storage_adapter_metadata_contract_artifact_status": "materialized_static_metadata_contract",
    "audit_storage_adapter_contract_artifact_path_status": "materialized_static_path",
    "audit_storage_adapter_contract_artifact_materialization_status": FOLLOWUP_MATERIALIZATION_STATUS,
}
FOLLOWUP_SELECTION_ALIGNMENT = {
    "audit_store_storage_adapter_backend_product_selection_review_status": FOLLOWUP_SELECTION_STATUS,
    "audit_storage_adapter_backend_product_selection_status": "selected_static_product_class_without_backend_provider",
    "audit_storage_adapter_selected_backend_product_class": "managed_database_append_only_table",
    "audit_storage_adapter_selected_backend_product_profile": "reserved_managed_database_append_only_table_profile",
    "audit_storage_adapter_database_product_status": "not_selected",
    "audit_storage_adapter_database_connection_provider_status": "not_created",
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_SELECTION_NEXT_DEPENDENCY,
}
FOLLOWUP_AFTER_SELECTION_ALIGNMENT = {
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status": (
        FOLLOWUP_AFTER_SELECTION_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": (
        "storage_adapter_runtime_task_card_still_blocked_after_append_only_table_schema_boundary_readiness"
    ),
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY,
    "audit_storage_adapter_database_provider_driver_dsn_tls_role_policy_status": (
        "defined_without_runtime"
    ),
    "audit_storage_adapter_append_only_table_schema_boundary_status": "defined_without_sql_or_runtime",
    "audit_storage_adapter_migration_schema_marker_boundary_status": "logical_schema_marker_handoff_boundary_defined",
    "audit_storage_adapter_offline_adapter_smoke_strategy_status": "required_before_runtime_task_card",
    "audit_storage_adapter_negative_leakage_runtime_scan_boundary_status": "required_before_runtime_task_card",
}

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.json",
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
    ),
    "production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.json",
        "audit_store_concrete_durable_backend_selection_review_defined",
    ),
    "production-secret-backend-audit-store-runtime-blocker-matrix-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json",
        "audit_store_runtime_blocker_matrix_defined",
    ),
    "production-secret-backend-audit-store-runtime-event-schema-artifact-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.json",
        "audit_store_runtime_event_schema_artifact_implemented",
    ),
    "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.json"
        ),
        "audit_store_writer_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-idempotency-runtime-implementation-entry-review-v1.json"
        ),
        "audit_store_idempotency_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1.json"
        ),
        "audit_store_delivery_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}

EXPECTED_ENTRY_FIELDS = {
    "status": "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
    "entry_decision": "storage_adapter_runtime_task_card_blocked_before_backend_product_evidence",
    "runtime_task_decision": "storage_adapter_runtime_task_card_blocked_before_backend_product_evidence",
    "audit_store_runtime_entry_refresh_v5_status": "audit_store_runtime_implementation_entry_refresh_v5_defined",
    "concrete_durable_backend_selection_review_status": "audit_store_concrete_durable_backend_selection_review_defined",
    "selected_backend_family": "append_only_metadata_audit_log",
    "selected_reserved_candidate": "reserved_append_only_audit_log",
    "metadata_storage_adapter_contract_status": "reviewed_static_contract",
    "backend_product_evidence_status": "not_selected",
    "append_only_write_semantics_status": "required_before_runtime_task_card",
    "retention_redaction_policy_evidence_status": "required_before_runtime_task_card",
    "offline_validation_status": "not_created",
    "negative_leakage_scan_status": "not_created",
    "rollback_recovery_evidence_status": "required_before_runtime_task_card",
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
    "runtime_task_card_allowed_now",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "storage_adapter_client_created_in_this_slice",
    "backend_product_selected_in_this_slice",
    "database_connection_provider_enabled",
    "database_driver_selected_in_this_slice",
    "sql_migration_created_in_this_slice",
    "schema_marker_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "delivery_executed_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "duplicate_detector_created_in_this_slice",
    "production_resolver_runtime_task_card_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_BLOCKED = {
    "backend_product_evidence_missing",
    "metadata_storage_adapter_contract_artifact_missing",
    "append_only_semantics_not_proven",
    "retention_redaction_policy_evidence_missing",
    "offline_validation_not_created",
    "negative_leakage_scan_not_created",
    "rollback_recovery_evidence_missing",
    "storage_adapter_runtime_task_card_not_created",
    "storage_adapter_runtime_not_created",
    "database_connection_provider_blocked",
    "audit_store_runtime_task_card_not_created",
    "production_resolver_runtime_not_created",
    "repository_mode_disabled",
    "production_api_not_created",
}

EXPECTED_REQUIREMENTS = {
    "disabled-by-default storage adapter runtime gate",
    "metadata-only storage adapter input envelope",
    "metadata-only storage adapter result envelope",
    "append-only write semantics",
    "backend product evidence",
    "retention and redaction policy evidence",
    "offline validation without provider or database side effects",
    "negative leakage scan",
    "rollback and recovery evidence",
    "writer runtime dependency gate",
    "idempotency runtime dependency gate",
    "delivery runtime dependency gate",
    "sanitized diagnostics allowlist",
    "forbidden material scan",
    "side effect counters",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_entry_dependency_missing",
    "audit_store_storage_adapter_entry_task_card_blocked",
    "audit_store_storage_adapter_entry_backend_product_missing",
    "audit_store_storage_adapter_entry_retention_redaction_missing",
    "audit_store_storage_adapter_entry_offline_validation_missing",
    "audit_store_storage_adapter_entry_database_provider_forbidden",
    "audit_store_storage_adapter_runtime_created_in_entry_review",
    "audit_store_storage_adapter_entry_secret_material_detected",
    "audit_store_storage_adapter_entry_scope_overreach",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.md",
    (
        "docs/task-cards/"
        "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1-plan.md"
    ),
    (
        "scripts/checks/fixtures/"
        "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.json"
    ),
    "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.py",
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
        == "production_ops_secret_backend_audit_store_storage_adapter_runtime_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id")
        == "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
        "unexpected slice status",
    )
    require(
        slice_info.get("entry_decision") == "storage_adapter_runtime_task_card_blocked_before_backend_product_evidence",
        "unexpected entry decision",
    )
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "storage_adapter_runtime_task_card_created",
        "storage_adapter_runtime_created",
        "backend_product_selected",
        "database_provider_selected",
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in claims, f"does_not_claim missing {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(EXPECTED_DEPENDENCIES) <= set(dependencies), "dependency ids drifted")
    for dependency_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("evidence") == path, f"{dependency_id} evidence path drifted")
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        require(source_status(load_json(REPO_ROOT / path)) == expected_status, f"{dependency_id} source status drifted")


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_boundary") or {}
    for field, expected in EXPECTED_ENTRY_FIELDS.items():
        require(boundary.get(field) == expected, f"entry_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"entry_boundary.{field} must remain false")
    blocked = set(fixture.get("blocked_conditions") or [])
    require(EXPECTED_BLOCKED <= blocked, f"missing blocked conditions: {sorted(EXPECTED_BLOCKED - blocked)}")
    requirements = set(fixture.get("future_runtime_task_card_requirements") or [])
    require(EXPECTED_REQUIREMENTS <= requirements, "future runtime task card requirements drifted")


def assert_failure_diagnostics_and_policies(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} boundary missing")
        require(item.get("sanitized_diagnostic"), f"{code} diagnostic missing")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(not (allowed & forbidden), "diagnostic allowlist intersects forbidden fields")
    require(
        sample.get("runtime_task_decision") == "storage_adapter_runtime_task_card_blocked_before_backend_product_evidence",
        "diagnostic sample decision drifted",
    )

    fallback = fixture.get("no_fallback_policy") or {}
    require(fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for source in {"append_only_metadata_audit_log_family_selection", "runtime_event_schema_artifact", "fake_resolver_runtime"}:
        require(source in set(fallback.get("forbidden_sources") or []), f"missing no fallback source {source}")

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for name, value in counters.items():
        require(value == 0, f"side effect counter {name} must stay 0")


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
        "audit_store_runtime_implementation_task_card",
        "audit_store_runtime",
        "production_resolver_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact kind missing: {artifact}")
    for path in guard.get("files_must_not_exist") or []:
        if (
            followup_materialization_exists()
            and path == "contracts/production-secret-audit-storage-adapter.metadata-contract.json"
        ):
            continue
        require(not (REPO_ROOT / str(path)).exists(), f"forbidden artifact exists: {path}")


def assert_blocker_matrix_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("blocker_matrix_alignment") or {}
    require(
        alignment.get("status") == "storage_adapter_runtime_entry_review_defined_task_card_blocked",
        "blocker matrix alignment status drifted",
    )
    matrix = load_json(BLOCKER_MATRIX_PATH)
    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    if followup_after_selection_exists():
        expected_status = FOLLOWUP_AFTER_SELECTION_BLOCKER_STATUS
        expected_source = FOLLOWUP_AFTER_SELECTION_SOURCE
    elif followup_selection_exists():
        expected_status = "storage_adapter_backend_product_selection_review_defined_task_card_blocked"
        expected_source = "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1"
    else:
        expected_status = "storage_adapter_runtime_entry_refresh_defined_task_card_blocked"
        expected_source = "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1"
    require(
        durable.get("status") == expected_status,
        "durable backend blocker status not updated",
    )
    require(durable.get("source") == expected_source, "durable backend blocker source not updated")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable blocker must block audit runtime")
    boundary = matrix.get("matrix_boundary") or {}
    require(
        boundary.get("storage_adapter_runtime_implementation_entry_review_status")
        == "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
        "matrix boundary missing storage adapter entry review status",
    )


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    followup_exists = followup_materialization_exists()
    for field, expected in alignment.items():
        if field == "status":
            continue
        if followup_exists and field in FOLLOWUP_ALIGNMENT:
            expected = FOLLOWUP_ALIGNMENT[field]
        if followup_selection_exists() and field in FOLLOWUP_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_SELECTION_ALIGNMENT[field]
        if followup_after_selection_exists() and field in FOLLOWUP_AFTER_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_AFTER_SELECTION_ALIGNMENT[field]
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or [] if isinstance(item, dict)}
    item = planned.get("audit-store-storage-adapter-runtime-implementation-entry-review") or {}
    require(
        item.get("status") == "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
        "implementation readiness missing storage adapter entry review planned slice",
    )
    require(EXPECTED_ALLOWED_ARTIFACTS <= set(item.get("evidence") or []), "planned slice evidence drifted")


def assert_docs_and_registration() -> None:
    docs = {
        "docs/platform/production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.md": [
            "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
            "storage_adapter_runtime_task_card_blocked_before_backend_product_evidence",
            "Metadata-only Storage Adapter Contract",
            "storage_adapter_backend_product_evidence_readiness",
        ],
        (
            "docs/task-cards/"
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1-plan.md"
        ): [
            "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
            "storage_adapter_runtime_task_card_blocked_before_backend_product_evidence",
            "停止线",
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
            "offline_validation_evidence_readiness_defined_task_card_blocked",
        ],
        "docs/platform/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Runtime Implementation Entry Review v1",
            "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Storage Adapter Runtime Implementation Entry Review v1",
            "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
        ],
        "docs/features/workflow/saved-workflow-draft-v1.md": [
            "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
            "storage_adapter_backend_product_evidence_readiness",
        ],
        "docs/radishmind-current-focus.md": [
            "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
            "storage_adapter_backend_product_evidence_readiness",
        ],
        "docs/task-cards/README.md": [
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1",
            "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
        ],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.py",
            "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.json",
        ],
        "docs/devlogs/2026-W27.md": [
            "audit_store_storage_adapter_runtime_implementation_entry_review_defined",
        ],
    }
    for path, literals in docs.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    v5 = "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.py"
    current = "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.py"
    backend_product_evidence = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.py"
    )
    metadata_contract_artifact = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.py"
    )
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    resolver = "check-production-ops-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.py"
    for script in {v5, current, backend_product_evidence, metadata_contract_artifact, matrix, resolver}:
        require(script in check_repo, f"check-repo.py missing {script}")
    require(
        check_repo.index(v5)
        < check_repo.index(current)
        < check_repo.index(backend_product_evidence)
        < check_repo.index(metadata_contract_artifact)
        < check_repo.index(matrix)
        < check_repo.index(resolver),
        "check order drifted",
    )


def assert_no_secret_literals() -> None:
    text = "\n".join(
        read(path)
        for path in EXPECTED_ALLOWED_ARTIFACTS
        if path.endswith(".md") or path.endswith(".json")
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "-----BEGIN"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"storage adapter entry review evidence contains forbidden literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "dsn-like credential found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_boundary(fixture)
    assert_failure_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store storage adapter runtime implementation entry review checks passed.")


if __name__ == "__main__":
    main()
