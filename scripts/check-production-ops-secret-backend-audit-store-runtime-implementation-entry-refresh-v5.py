#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
FOLLOWUP_SELECTION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1.json"
)
FOLLOWUP_SELECTION_STATUS = "audit_store_storage_adapter_backend_product_selection_review_defined"
FOLLOWUP_DURABLE_BLOCKER_STATUS = "storage_adapter_backend_product_selection_review_defined_task_card_blocked"
FOLLOWUP_DURABLE_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-backend-product-selection-review-v1"
)
FOLLOWUP_AFTER_SELECTION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-product-selection-v1.json"
)
FOLLOWUP_AFTER_SELECTION_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_product_selection_defined"
)
FOLLOWUP_AFTER_SELECTION_DURABLE_BLOCKER_STATUS = (
    "storage_adapter_database_driver_selection_review_defined_task_card_blocked"
)
FOLLOWUP_AFTER_SELECTION_DURABLE_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-database-driver-selection-review-v1"
)
FOLLOWUP_CONNECTION_LIFECYCLE_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-database-connection-lifecycle-v1.json"
)
FOLLOWUP_CONNECTION_LIFECYCLE_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_defined"
)
FOLLOWUP_CONNECTION_LIFECYCLE_DURABLE_BLOCKER_STATUS = (
    "storage_adapter_runtime_entry_refresh_after_database_connection_lifecycle_defined_task_card_blocked"
)
FOLLOWUP_CONNECTION_LIFECYCLE_DURABLE_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-database-connection-lifecycle-v1"
)
FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1.json"
)
FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_STATUS = (
    "audit_store_storage_adapter_database_provider_connection_runtime_boundary_readiness_defined"
)
FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_DURABLE_BLOCKER_STATUS = (
    "storage_adapter_database_provider_connection_runtime_boundary_readiness_defined_task_card_blocked"
)
FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_DURABLE_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-database-provider-connection-runtime-boundary-readiness-v1"
)
FOLLOWUP_AFTER_PROVIDER_BOUNDARY_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-database-provider-connection-runtime-boundary-v1.json"
)
FOLLOWUP_AFTER_PROVIDER_BOUNDARY_STATUS = (
    "audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_provider_connection_runtime_boundary_defined"
)
FOLLOWUP_AFTER_PROVIDER_BOUNDARY_DURABLE_BLOCKER_STATUS = (
    "storage_adapter_runtime_entry_refresh_after_database_provider_connection_runtime_boundary_defined_task_card_blocked"
)
FOLLOWUP_AFTER_PROVIDER_BOUNDARY_DURABLE_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-"
    "after-database-provider-connection-runtime-boundary-v1"
)
RUNTIME_REFRESH_DURABLE_BLOCKER_STATUS = "storage_adapter_runtime_entry_refresh_defined_task_card_blocked"
RUNTIME_REFRESH_DURABLE_BLOCKER_SOURCE = (
    "production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-v1"
)

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.json",
        "audit_store_concrete_durable_backend_selection_review_defined",
    ),
    "production-secret-backend-audit-store-runtime-blocker-matrix-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json",
        "audit_store_runtime_blocker_matrix_defined",
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
    "production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2": (
        "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.json",
        "production_resolver_runtime_implementation_entry_refresh_v2_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}

EXPECTED_BOUNDARY = {
    "status": "audit_store_runtime_implementation_entry_refresh_v5_defined",
    "entry_decision": "audit_store_runtime_task_card_still_blocked_after_concrete_backend_selection_review",
    "concrete_durable_backend_selection_review_status": "audit_store_concrete_durable_backend_selection_review_defined",
    "selected_backend_family": "append_only_metadata_audit_log",
    "selected_reserved_candidate": "reserved_append_only_audit_log",
    "durable_audit_backend_status": "static_backend_family_selected_runtime_blocked",
    "next_independent_runtime_dependency": "storage_adapter_runtime_implementation_entry_review",
    "storage_adapter_runtime_status": "not_created",
    "database_connection_provider_status": "blocked",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_writer_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "delivery_runtime_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "runtime_task_card_allowed_now",
    "storage_adapter_runtime_created_in_this_slice",
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
    "storage_adapter_runtime_not_created",
    "database_connection_provider_blocked",
    "audit_store_runtime_task_card_not_created",
    "audit_writer_runtime_not_created",
    "delivery_runtime_not_created",
    "idempotency_runtime_not_created",
    "duplicate_detector_not_created",
    "retry_executor_not_created",
    "operator_approval_runtime_not_created",
    "credential_handle_runtime_not_created",
    "backend_health_runtime_not_created",
    "no_secret_leakage_smoke_runtime_not_created",
    "production_resolver_runtime_not_created",
    "repository_mode_disabled",
    "production_api_not_created",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_runtime_refresh_v5_dependency_missing",
    "audit_store_runtime_refresh_v5_task_card_still_blocked",
    "audit_store_runtime_refresh_v5_storage_adapter_missing",
    "audit_store_runtime_refresh_v5_database_provider_forbidden",
    "audit_store_runtime_refresh_v5_writer_runtime_missing",
    "audit_store_runtime_refresh_v5_delivery_idempotency_missing",
    "audit_store_runtime_refresh_v5_external_runtime_missing",
    "audit_store_runtime_refresh_v5_scope_overreach",
    "audit_store_runtime_refresh_v5_secret_material_detected",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.md",
    "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.json",
    "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.py",
}

EXPECTED_FORBIDDEN_PATHS = {
    "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-v1-plan.md",
    "docs/task-cards/production-secret-backend-audit-store-storage-adapter-runtime-v1-plan.md",
    "services/platform/internal/secretbackend/audit_store.go",
    "services/platform/internal/secretbackend/audit_writer.go",
    "services/platform/internal/secretbackend/audit_delivery.go",
    "services/platform/internal/secretbackend/audit_idempotency.go",
    "services/platform/internal/secretbackend/storage_adapter.go",
    "services/platform/internal/secretbackend/production_resolver.go",
}

DOC_REFERENCES = {
    "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.md": [
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
        "audit_store_runtime_task_card_still_blocked_after_concrete_backend_selection_review",
        "storage_adapter_runtime_implementation_entry_review",
    ],
    "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5-plan.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5",
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
        "storage_adapter_runtime_implementation_entry_review",
    ],
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5",
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh v5",
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
    ],
    "docs/features/workflow/saved-workflow-draft-v1.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh v5",
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh v5",
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
    ],
    "docs/task-cards/README.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5",
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.py",
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.json",
    ],
    "docs/devlogs/2026-W27.md": [
        "audit_store_runtime_implementation_entry_refresh_v5_defined",
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


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def followup_selection_exists() -> bool:
    if not FOLLOWUP_SELECTION_FIXTURE_PATH.exists():
        return False
    selection = load_json(FOLLOWUP_SELECTION_FIXTURE_PATH)
    return source_status(selection) == FOLLOWUP_SELECTION_STATUS


def followup_after_selection_exists() -> bool:
    if not FOLLOWUP_AFTER_SELECTION_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_AFTER_SELECTION_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_AFTER_SELECTION_STATUS


def followup_connection_lifecycle_exists() -> bool:
    if not FOLLOWUP_CONNECTION_LIFECYCLE_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_CONNECTION_LIFECYCLE_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_CONNECTION_LIFECYCLE_STATUS


def followup_connection_runtime_boundary_exists() -> bool:
    if not FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_STATUS


def followup_after_provider_boundary_exists() -> bool:
    if not FOLLOWUP_AFTER_PROVIDER_BOUNDARY_FIXTURE_PATH.exists():
        return False
    followup = load_json(FOLLOWUP_AFTER_PROVIDER_BOUNDARY_FIXTURE_PATH)
    return source_status(followup) == FOLLOWUP_AFTER_PROVIDER_BOUNDARY_STATUS


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_audit_store_runtime_implementation_entry_refresh_v5",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v5",
        "unexpected slice id",
    )
    require(slice_info.get("status") == "audit_store_runtime_implementation_entry_refresh_v5_defined", "bad status")
    require(
        slice_info.get("entry_decision")
        == "audit_store_runtime_task_card_still_blocked_after_concrete_backend_selection_review",
        "unexpected entry decision",
    )
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "storage_adapter_runtime_created",
        "database_provider_selected",
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "audit_writer_runtime_created",
        "delivery_runtime_created",
        "idempotency_runtime_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in claims, f"does_not_claim missing {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        row = dependencies.get(dependency_id)
        require(row is not None, f"dependency missing: {dependency_id}")
        require(row.get("evidence") == relative_path, f"dependency evidence mismatch: {dependency_id}")
        require(row.get("status") == expected_status, f"dependency status mismatch: {dependency_id}")
        require(source_status(load_json(REPO_ROOT / relative_path)) == expected_status, f"source status drifted: {dependency_id}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_refresh_boundary") or {}
    for key, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(key) == expected, f"entry_refresh_boundary.{key} drifted")
    for key in EXPECTED_FALSE_FLAGS:
        require(boundary.get(key) is False, f"entry_refresh_boundary.{key} must be false")

    blocked = set(fixture.get("blocked_conditions") or [])
    require(EXPECTED_BLOCKED <= blocked, f"missing blocked conditions: {sorted(EXPECTED_BLOCKED - blocked)}")

    requirements = set(fixture.get("future_next_dependency_requirements") or [])
    for item in {
        "metadata-only storage adapter contract",
        "backend product evidence",
        "offline validation without provider or database side effects",
        "no runtime task card without independent review",
    }:
        require(item in requirements, f"future next dependency requirement missing: {item}")


def assert_failure_diagnostics_and_side_effects(fixture: dict[str, Any]) -> None:
    failure_codes = {str(row.get("code")) for row in fixture.get("failure_mapping") or []}
    require(failure_codes == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(
        sample.get("next_independent_runtime_dependency") == "storage_adapter_runtime_implementation_entry_review",
        "diagnostic next dependency drifted",
    )

    fallback = fixture.get("no_fallback_policy") or {}
    require(fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for source in {"static_schema_artifact", "concrete_backend_family_selection", "fake_resolver_runtime"}:
        require(source in set(fallback.get("forbidden_sources") or []), f"missing no fallback source {source}")

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for name, value in counters.items():
        require(value == 0, f"side effect counter {name} must stay 0")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    require(set(guard.get("allowed_new_artifacts") or []) == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifacts drifted")
    for relative_path in EXPECTED_ALLOWED_ARTIFACTS:
        require((REPO_ROOT / relative_path).exists(), f"allowed artifact missing: {relative_path}")
    forbidden = set(guard.get("forbidden_artifact_kinds") or [])
    for artifact in {
        "storage_adapter_runtime_task_card",
        "storage_adapter_runtime",
        "database_connection_provider",
        "audit_store_runtime_implementation_task_card",
        "audit_store_runtime",
        "production_resolver_runtime",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact kind missing: {artifact}")
    for relative_path in EXPECTED_FORBIDDEN_PATHS:
        require(not (REPO_ROOT / relative_path).exists(), f"forbidden artifact exists: {relative_path}")


def assert_blocker_matrix_alignment() -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    if followup_after_provider_boundary_exists():
        expected_status = FOLLOWUP_AFTER_PROVIDER_BOUNDARY_DURABLE_BLOCKER_STATUS
        expected_source = FOLLOWUP_AFTER_PROVIDER_BOUNDARY_DURABLE_BLOCKER_SOURCE
    elif followup_connection_runtime_boundary_exists():
        expected_status = FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_DURABLE_BLOCKER_STATUS
        expected_source = FOLLOWUP_CONNECTION_RUNTIME_BOUNDARY_DURABLE_BLOCKER_SOURCE
    elif followup_connection_lifecycle_exists():
        expected_status = FOLLOWUP_CONNECTION_LIFECYCLE_DURABLE_BLOCKER_STATUS
        expected_source = FOLLOWUP_CONNECTION_LIFECYCLE_DURABLE_BLOCKER_SOURCE
    elif followup_after_selection_exists():
        expected_status = FOLLOWUP_AFTER_SELECTION_DURABLE_BLOCKER_STATUS
        expected_source = FOLLOWUP_AFTER_SELECTION_DURABLE_BLOCKER_SOURCE
    elif followup_selection_exists():
        expected_status = FOLLOWUP_DURABLE_BLOCKER_STATUS
        expected_source = FOLLOWUP_DURABLE_BLOCKER_SOURCE
    else:
        expected_status = RUNTIME_REFRESH_DURABLE_BLOCKER_STATUS
        expected_source = RUNTIME_REFRESH_DURABLE_BLOCKER_SOURCE
    require(
        boundary.get("durable_audit_backend_status") == expected_status,
        "blocker matrix durable backend status drifted",
    )
    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(durable.get("status") == expected_status, "durable blocker status drifted")
    require(durable.get("source") == expected_source, "durable blocker source drifted")
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable backend must still block audit runtime")
    for blocker_id in {"audit_writer_runtime", "idempotency_runtime", "delivery_runtime"}:
        row = blockers.get(blocker_id) or {}
        require(row.get("status") == "entry_review_defined_task_card_blocked", f"{blocker_id} status drifted")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for key, expected in (fixture.get("implementation_readiness_alignment") or {}).get("target_fields", {}).items():
        require(target.get(key) == expected, f"implementation readiness {key} drifted")

    planned_slice = (fixture.get("implementation_readiness_alignment") or {}).get("planned_slice") or {}
    planned = {str(row.get("id")): row for row in readiness.get("planned_slices") or [] if isinstance(row, dict)}
    row = planned.get(planned_slice.get("id"))
    require(row is not None, "implementation readiness missing v5 planned slice")
    require(row.get("status") == planned_slice.get("status"), "implementation readiness v5 planned slice status drifted")


def assert_docs_and_registration() -> None:
    for relative_path, literals in DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current = "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v5.py"
    before = "check-production-ops-secret-backend-audit-store-delivery-runtime-implementation-entry-review-v1.py"
    storage = "check-production-ops-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-review-v1.py"
    backend_product_evidence = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.py"
    )
    metadata_contract_artifact = (
        "check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.py"
    )
    matrix = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    after = "check-production-ops-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.py"
    for script in {before, current, storage, backend_product_evidence, metadata_contract_artifact, matrix, after}:
        require(script in check_repo, f"check-repo.py missing {script}")
    require(
        check_repo.index(before)
        < check_repo.index(current)
        < check_repo.index(storage)
        < check_repo.index(backend_product_evidence)
        < check_repo.index(metadata_contract_artifact)
        < check_repo.index(matrix)
        < check_repo.index(after),
        "check-repo.py order drifted",
    )


def assert_no_secret_literals() -> None:
    text = "\n".join(
        read(path)
        for path in EXPECTED_ALLOWED_ARTIFACTS
        if path.endswith(".md") or path.endswith(".json")
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "-----BEGIN"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"v5 evidence contains forbidden secret-looking literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_failure_diagnostics_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_blocker_matrix_alignment()
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store runtime implementation entry refresh v5 checks passed.")


if __name__ == "__main__":
    main()
