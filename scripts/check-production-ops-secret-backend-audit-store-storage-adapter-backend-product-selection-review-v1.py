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
    "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
CONTRACT_PATH = REPO_ROOT / "contracts/production-secret-audit-storage-adapter.metadata-contract.json"

SLICE_ID = "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1"
SLICE_STATUS = "audit_store_storage_adapter_backend_product_selection_review_defined"
SELECTION_DECISION = "storage_adapter_product_class_selected_managed_database_append_only_table_runtime_blocked"
SELECTED_BACKEND_FAMILY = "append_only_metadata_audit_log"
SELECTED_RESERVED_CANDIDATE = "reserved_append_only_audit_log"
SELECTED_PRODUCT_CLASS = "managed_database_append_only_table"
SELECTED_PRODUCT_PROFILE = "reserved_managed_database_append_only_table_profile"
SELECTION_SCOPE = "static_product_class_only"
SELECTION_STATUS = "selected_static_product_class_without_backend_provider"
NEXT_DEPENDENCY = "storage_adapter_runtime_implementation_entry_refresh_after_product_selection"
FOLLOWUP_AFTER_SELECTION_FIXTURE = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.json"
)
FOLLOWUP_AFTER_SELECTION_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined"
)
FOLLOWUP_AFTER_SELECTION_DECISION = "storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_entry_refresh"
FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY = "storage_adapter_database_provider_connection_runtime_boundary_readiness"
FOLLOWUP_AFTER_SELECTION_BLOCKER_STATUS = (
    "storage_adapter_runtime_entry_refresh_after_database_connection_lifecycle_defined_task_card_blocked"
)
FOLLOWUP_AFTER_SELECTION_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1"
)
FOLLOWUP_AFTER_SELECTION_ALIGNMENT = {
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status": (
        FOLLOWUP_AFTER_SELECTION_STATUS
    ),
    "audit_storage_adapter_runtime_task_card_decision": FOLLOWUP_AFTER_SELECTION_DECISION,
    "audit_storage_adapter_current_next_dependency": FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY,
    "audit_storage_adapter_database_connection_provider_status": "not_created",
    "audit_storage_adapter_database_provider_driver_dsn_tls_role_policy_status": (
        "defined_without_runtime"
    ),
    "audit_storage_adapter_append_only_table_schema_boundary_status": "defined_without_sql_or_runtime",
    "audit_storage_adapter_migration_schema_marker_boundary_status": "logical_schema_marker_handoff_boundary_defined",
    "audit_storage_adapter_offline_adapter_smoke_strategy_status": "required_before_runtime_task_card",
    "audit_storage_adapter_negative_leakage_runtime_scan_boundary_status": "defined_without_runtime",
}

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.json",
        "audit_store_storage_adapter_metadata_contract_artifact_materialized",
    ),
    "production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.json",
        "audit_store_storage_adapter_backend_product_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.json",
        "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json"
        ),
        "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.json",
        "audit_store_storage_adapter_offline_validation_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.json",
        "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-rollback-recovery-evidence-readiness-v1.json",
        "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined",
    ),
    "production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.json",
        "audit_store_concrete_durable_backend_selection_review_defined",
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
    "backend_vendor_selected_in_this_slice",
    "database_product_selected_in_this_slice",
    "database_connection_provider_enabled",
    "storage_adapter_runtime_task_card_created_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "storage_adapter_client_created_in_this_slice",
    "sql_migration_created_in_this_slice",
    "schema_marker_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_REVIEW = {
    "status": SLICE_STATUS,
    "selection_decision": SELECTION_DECISION,
    "selected_backend_family": SELECTED_BACKEND_FAMILY,
    "selected_reserved_candidate": SELECTED_RESERVED_CANDIDATE,
    "selected_backend_product_class": SELECTED_PRODUCT_CLASS,
    "selected_backend_product_profile": SELECTED_PRODUCT_PROFILE,
    "selection_scope": SELECTION_SCOPE,
    "backend_product_evidence_status": "audit_store_storage_adapter_backend_product_evidence_readiness_defined",
    "backend_product_selection_status": SELECTION_STATUS,
    "metadata_contract_artifact_status": "audit_store_storage_adapter_metadata_contract_artifact_materialized",
    "append_only_semantics_status": "audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined",
    "retention_redaction_status": "audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined",
    "offline_validation_status": "audit_store_storage_adapter_offline_validation_evidence_readiness_defined",
    "negative_leakage_scan_status": "audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined",
    "rollback_recovery_status": "audit_store_storage_adapter_rollback_recovery_evidence_readiness_defined",
    "database_connection_provider_status": "blocked",
    "database_driver_status": "not_selected",
    "database_product_status": "not_selected",
    "sql_migration_status": "not_created",
    "schema_marker_status": "not_created",
    "storage_adapter_runtime_task_card_status": "not_created",
    "storage_adapter_runtime_status": "not_created",
    "storage_adapter_client_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_writer_runtime_status": "not_created",
    "delivery_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
    "current_next_dependency": NEXT_DEPENDENCY,
}

EXPECTED_FAILURE_CODES = {
    "audit_store_storage_adapter_backend_product_selection_dependency_missing",
    "audit_store_storage_adapter_backend_product_selection_class_missing",
    "audit_store_storage_adapter_backend_product_selection_invalid_class",
    "audit_store_storage_adapter_backend_product_selection_vendor_forbidden",
    "audit_store_storage_adapter_backend_product_selection_runtime_forbidden",
    "audit_store_storage_adapter_backend_product_selection_secret_material_detected",
    "audit_store_storage_adapter_backend_product_selection_scope_overreach",
}

EXPECTED_RATIONALE = {
    "durable_backend_family_alignment",
    "metadata_contract_alignment",
    "append_only_table_runtime_boundary",
    "database_provider_boundary",
}

EXPECTED_ZERO_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "network_call_count",
    "cloud_secret_call_count",
    "provider_call_count",
    "database_connection_count",
    "driver_open_count",
    "sql_execution_count",
    "database_product_selected_count",
    "backend_vendor_selected_count",
    "storage_adapter_runtime_task_card_created_count",
    "storage_adapter_runtime_created_count",
    "audit_store_runtime_created_count",
    "audit_writer_runtime_created_count",
    "audit_event_write_count",
    "delivery_execution_count",
    "idempotency_decision_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.md",
    (
        "docs/task-cards/"
        "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1-plan.md"
    ),
    (
        "scripts/checks/fixtures/"
        "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json"
    ),
    "scripts/check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.py",
}

DOC_REFERENCES = {
    "docs/platform/production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.md": [
        SLICE_STATUS,
        SELECTION_DECISION,
        SELECTED_PRODUCT_CLASS,
        SELECTED_PRODUCT_PROFILE,
        "不创建 storage adapter runtime",
        "不选择 PostgreSQL",
        "不启用 repository mode",
    ],
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1-plan.md": [
        SLICE_ID,
        SLICE_STATUS,
        SELECTED_PRODUCT_CLASS,
        NEXT_DEPENDENCY,
    ],
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


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def followup_after_selection_exists() -> bool:
    if not FOLLOWUP_AFTER_SELECTION_FIXTURE.exists():
        return False
    followup = load_json(FOLLOWUP_AFTER_SELECTION_FIXTURE)
    return source_status(followup) == FOLLOWUP_AFTER_SELECTION_STATUS


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_storage_adapter_backend_product_selection_review_v1",
        "fixture kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == SLICE_ID, "slice id drifted")
    require(slice_info.get("track") == "Production Ops Hardening v1", "track drifted")
    require(slice_info.get("status") == SLICE_STATUS, "slice status drifted")
    require(slice_info.get("selection_decision") == SELECTION_DECISION, "selection decision drifted")
    for field in ("task_card", "platform_topic"):
        relative_path = str(slice_info.get(field) or "")
        require(relative_path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {relative_path}")
        require((REPO_ROOT / relative_path).exists(), f"{field} missing: {relative_path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "backend_vendor_selected",
        "database_product_selected",
        "database_provider_selected",
        "storage_adapter_runtime_task_card_created",
        "storage_adapter_runtime_created",
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in claims, f"missing does_not_claim guard: {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    deps = rows_by_id(fixture, "depends_on", "id")
    require(set(deps) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dep_id, (evidence, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = deps[dep_id]
        require(item.get("evidence") == evidence, f"{dep_id} evidence path drifted")
        require(item.get("status") == expected_status, f"{dep_id} status drifted")
        source = load_json(REPO_ROOT / evidence)
        require(source_status(source) == expected_status, f"{dep_id} source status drifted")


def assert_selection_review(fixture: dict[str, Any]) -> None:
    review = fixture.get("selection_review") or {}
    for field, expected in EXPECTED_REVIEW.items():
        require(review.get(field) == expected, f"selection_review.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(review.get(field) is False, f"selection_review.{field} must stay false")

    allowlist_source = fixture.get("product_class_allowlist_source") or {}
    require(allowlist_source.get("required_class") == SELECTED_PRODUCT_CLASS, "required product class drifted")
    allowlist = set(allowlist_source.get("allowed_classes") or [])
    evidence = load_json(REPO_ROOT / str(allowlist_source.get("fixture") or ""))
    require(SELECTED_PRODUCT_CLASS in allowlist, "selected product class missing from local allowlist")
    require(allowlist == set(evidence.get("product_class_allowlist") or []), "product class allowlist source drifted")

    rationale = rows_by_id(fixture, "selection_rationale", "id")
    require(set(rationale) == EXPECTED_RATIONALE, "selection rationale ids drifted")
    for item in rationale.values():
        require(item.get("status"), "selection rationale status missing")
        require(item.get("evidence"), "selection rationale evidence missing")


def assert_diagnostics_and_failures(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("diagnostic_envelope") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample).issubset(allowed), "diagnostic sample has fields outside allowlist")
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    require(not (set(sample) & forbidden), "diagnostic sample uses forbidden field names")
    for required in {
        "audit_store_storage_adapter_backend_product_selection_review_status",
        "selection_decision",
        "selected_backend_product_class",
        "selected_backend_product_profile",
        "backend_product_selection_status",
        "database_connection_provider_status",
        "failure_code",
        "failure_boundary",
        "sanitized_diagnostic",
        "request_id",
        "audit_ref",
        "policy_version",
    }:
        require(required in allowed, f"diagnostic allowlist missing {required}")
        require(required in sample, f"diagnostic sample missing {required}")
    require(sample.get("selected_backend_product_class") == SELECTED_PRODUCT_CLASS, "sample product class drifted")
    require(sample.get("backend_product_selection_status") == SELECTION_STATUS, "sample selection status drifted")

    failure_rows = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failure_rows) == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for row in failure_rows.values():
        require(row.get("failure_boundary"), "failure boundary missing")
        require(row.get("sanitized_diagnostic"), "sanitized diagnostic missing")


def assert_side_effects_and_artifacts(fixture: dict[str, Any]) -> None:
    counters = fixture.get("side_effect_counters") or {}
    require(set(counters) == EXPECTED_ZERO_COUNTERS, "side effect counter ids drifted")
    for key, value in counters.items():
        require(value == 0, f"{key} must be zero")

    guard = fixture.get("artifact_guard") or {}
    require(set(guard.get("allowed_added_artifacts") or []) == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifacts drifted")
    for relative_path in guard.get("allowed_added_artifacts") or []:
        require((REPO_ROOT / relative_path).exists(), f"allowed artifact missing: {relative_path}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / relative_path).exists(), f"forbidden artifact exists: {relative_path}")
    forbidden_kinds = set(guard.get("forbidden_artifact_kinds") or [])
    for kind in {
        "database_product_selection_artifact",
        "backend_vendor_selection_artifact",
        "storage_adapter_runtime",
        "database_connection_provider",
        "sql_migration",
        "audit_store_runtime",
        "production_resolver_runtime",
        "public_production_api",
    }:
        require(kind in forbidden_kinds, f"missing forbidden artifact kind: {kind}")

    for relative_path in guard.get("allowed_added_artifacts") or []:
        if str(relative_path).endswith(".py"):
            continue
        content = read(relative_path)
        for pattern in ("Bearer ", "BEGIN PRIVATE KEY", "AKIA", "postgres://", "mysql://", "jdbc:"):
            require(pattern not in content, f"secret-like literal found in {relative_path}: {pattern}")
        require(re.search(r"\bsk-[A-Za-z0-9]{12,}", content) is None, f"secret-like token found in {relative_path}")


def assert_alignment(fixture: dict[str, Any]) -> None:
    implementation = load_json(IMPLEMENTATION_READINESS_PATH)
    target = implementation.get("implementation_target") or {}
    expected_impl = fixture.get("implementation_readiness_alignment") or {}
    for field, expected in expected_impl.items():
        if field == "status":
            continue
        if followup_after_selection_exists() and field in FOLLOWUP_AFTER_SELECTION_ALIGNMENT:
            expected = FOLLOWUP_AFTER_SELECTION_ALIGNMENT[field]
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    required = rows_by_id(implementation, "planned_slices", "id")
    require(
        "audit-store-storage-adapter-backend-product-selection-review" in required,
        "implementation readiness missing backend product selection planned slice",
    )
    require(
        required["audit-store-storage-adapter-backend-product-selection-review"].get("status") == SLICE_STATUS,
        "implementation readiness planned slice status drifted",
    )

    blocker_matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = blocker_matrix.get("matrix_boundary") or {}
    expected_matrix = fixture.get("blocker_matrix_alignment") or {}
    require(boundary.get("storage_adapter_backend_product_selection_review_status") == SLICE_STATUS, "matrix review status drifted")
    require(boundary.get("storage_adapter_backend_product_selection_status") == SELECTION_STATUS, "matrix selection status drifted")
    require(boundary.get("storage_adapter_selected_backend_product_class") == SELECTED_PRODUCT_CLASS, "matrix product class drifted")
    expected_next_dependency = (
        FOLLOWUP_AFTER_SELECTION_NEXT_DEPENDENCY if followup_after_selection_exists() else NEXT_DEPENDENCY
    )
    require(
        boundary.get("storage_adapter_current_next_dependency") == expected_next_dependency,
        "matrix next dependency drifted",
    )
    blockers = rows_by_id(blocker_matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    expected_blocker_status = expected_matrix.get("durable_backend_blocker_status_after_review")
    expected_blocker_source = expected_matrix.get("durable_backend_blocker_source_after_review")
    if followup_after_selection_exists():
        expected_blocker_status = FOLLOWUP_AFTER_SELECTION_BLOCKER_STATUS
        expected_blocker_source = FOLLOWUP_AFTER_SELECTION_SOURCE
        require(
            boundary.get("storage_adapter_runtime_implementation_entry_refresh_after_product_selection_status")
            == FOLLOWUP_AFTER_SELECTION_STATUS,
            "matrix after-selection refresh status drifted",
        )
        require(
            boundary.get("storage_adapter_runtime_task_card_decision") == FOLLOWUP_AFTER_SELECTION_DECISION,
            "matrix after-selection runtime decision drifted",
        )
    require(durable.get("status") == expected_blocker_status, "durable backend blocker status drifted")
    require(durable.get("source") == expected_blocker_source, "durable backend blocker source drifted")


def assert_docs() -> None:
    for relative_path, expected_fragments in DOC_REFERENCES.items():
        content = read(relative_path)
        for fragment in expected_fragments:
            require(fragment in content, f"{relative_path} missing {fragment}")


def assert_contract_still_metadata_only() -> None:
    contract = load_json(CONTRACT_PATH)
    guard = contract.get("artifact_guard") or {}
    require(guard.get("storage_adapter_runtime_status") == "not_created", "contract runtime guard drifted")
    require(guard.get("database_connection_provider_status") == "blocked", "contract database guard drifted")
    require("raw_storage_payload" in set(contract.get("forbidden_fields") or []), "contract forbidden fields drifted")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    materialization_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-metadata-contract-artifact-materialization-v1.py", [])'
    )
    selection_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-backend-product-selection-review-v1.py", [])'
    )
    after_selection_refresh_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.py", [])'
    )
    matrix_call = 'run_python_script("check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py", [])'
    resolver_call = (
        'run_python_script("check-production-ops-secret-backend-'
        'production-resolver-runtime-implementation-entry-refresh-v2.py", [])'
    )
    for call in (materialization_call, selection_call, after_selection_refresh_call, matrix_call, resolver_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(
        check_repo.index(materialization_call)
        < check_repo.index(selection_call)
        < check_repo.index(after_selection_refresh_call)
        < check_repo.index(matrix_call)
        < check_repo.index(resolver_call),
        "check-repo.py registration order drifted",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_selection_review(fixture)
    assert_diagnostics_and_failures(fixture)
    assert_side_effects_and_artifacts(fixture)
    assert_alignment(fixture)
    assert_docs()
    assert_contract_still_metadata_only()
    assert_check_repo_registration()
    print("production ops secret backend audit store storage adapter backend product selection review checks passed.")


if __name__ == "__main__":
    main()
