#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
BLOCKER_MATRIX_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-durable-backend-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.json",
        "audit_store_durable_backend_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-durable-backend-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-selection-readiness-v1.json",
        "audit_store_durable_backend_selection_readiness_defined",
    ),
    "production-secret-backend-audit-store-runtime-event-schema-artifact-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.json",
        "audit_store_runtime_event_schema_artifact_implemented",
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

EXPECTED_REVIEW = {
    "status": "audit_store_concrete_durable_backend_selection_review_defined",
    "selection_decision": "durable_backend_family_selected_static_append_only_audit_log_runtime_blocked",
    "selected_backend_family": "append_only_metadata_audit_log",
    "selected_reserved_candidate": "reserved_append_only_audit_log",
    "selection_scope": "static_family_only",
    "durable_audit_backend_status": "static_backend_family_selected_runtime_blocked",
    "storage_adapter_contract_status": "metadata_only_static_reference",
    "storage_adapter_runtime_status": "not_created",
    "database_connection_provider_status": "blocked",
    "database_driver_status": "not_selected",
    "sql_migration_status": "not_created",
    "schema_marker_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_writer_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "delivery_runtime_status": "not_created",
    "operator_approval_runtime_status": "not_created",
    "credential_handle_runtime_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "no_secret_leakage_smoke_runtime_status": "not_created",
    "production_resolver_runtime_task_card_status": "not_created",
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

EXPECTED_FAILURE_CODES = {
    "audit_store_concrete_durable_backend_selection_dependency_missing",
    "audit_store_concrete_durable_backend_selection_family_missing",
    "audit_store_concrete_durable_backend_selection_invalid_family",
    "audit_store_concrete_durable_backend_selection_runtime_forbidden",
    "audit_store_concrete_durable_backend_selection_database_provider_forbidden",
    "audit_store_concrete_durable_backend_selection_secret_material_detected",
    "audit_store_concrete_durable_backend_selection_scope_overreach",
}

EXPECTED_DIAGNOSTICS = {
    "audit_store_concrete_durable_backend_selection_review_status",
    "selection_decision",
    "selected_backend_family",
    "selected_reserved_candidate",
    "selection_scope",
    "gate",
    "gate_status",
    "storage_adapter_runtime_status",
    "database_connection_provider_status",
    "audit_store_runtime_task_card_status",
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
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "storage_adapter_runtime_created_count",
    "database_connection_count",
    "driver_open_count",
    "sql_execution_count",
    "schema_marker_read_count",
    "schema_marker_write_count",
    "audit_store_runtime_created_count",
    "audit_writer_runtime_created_count",
    "audit_event_write_count",
    "delivery_execution_count",
    "idempotency_decision_count",
    "duplicate_detection_count",
    "operator_approval_runtime_execution_count",
    "credential_handle_runtime_created_count",
    "credential_payload_created_count",
    "backend_health_runtime_created_count",
    "backend_health_check_count",
    "no_secret_leakage_smoke_runtime_created_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}

EXPECTED_REQUIRED_CHECKS = {
    "run audit store concrete durable backend selection review checker",
    "run audit store durable backend selection readiness checker",
    "run audit store runtime blocker matrix checker",
    "run implementation readiness checker",
    "run git diff check",
    "run fast repository check",
}

DOC_REFERENCES = {
    "docs/platform/production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.md": [
        "audit_store_concrete_durable_backend_selection_review_defined",
        "append_only_metadata_audit_log",
        "reserved_append_only_audit_log",
        "durable_backend_family_selected_static_append_only_audit_log_runtime_blocked",
        "不创建 storage adapter runtime",
        "不创建 audit store runtime task card",
        "不连接数据库",
        "不启用 repository mode",
        "不调用 production API",
    ],
    "docs/task-cards/production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1-plan.md": [
        "production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1",
        "audit_store_concrete_durable_backend_selection_review_defined",
        "append_only_metadata_audit_log",
        "reserved_append_only_audit_log",
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


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_concrete_durable_backend_selection_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-concrete-durable-backend-selection-review-v1",
        "unexpected slice id",
    )
    require(
        slice_info.get("status") == "audit_store_concrete_durable_backend_selection_review_defined",
        "unexpected slice status",
    )
    require(
        slice_info.get("selection_decision")
        == "durable_backend_family_selected_static_append_only_audit_log_runtime_blocked",
        "unexpected selection decision",
    )
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "durable_audit_backend_runtime_ready",
        "storage_adapter_runtime_created",
        "database_provider_selected",
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in does_not_claim, f"does_not_claim missing {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependency_rows = rows_by_id(fixture, "depends_on", "id")
    for dep_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        row = dependency_rows.get(dep_id)
        require(row is not None, f"dependency missing: {dep_id}")
        require(row.get("evidence") == relative_path, f"dependency evidence mismatch: {dep_id}")
        require(row.get("status") == expected_status, f"dependency status mismatch: {dep_id}")
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"dependency source status mismatch: {dep_id}")


def assert_selection_review(fixture: dict[str, Any]) -> None:
    review = fixture.get("selection_review") or {}
    for key, expected in EXPECTED_REVIEW.items():
        require(review.get(key) == expected, f"selection_review.{key} must be {expected}")
    for key in EXPECTED_FALSE_FLAGS:
        require(review.get(key) is False, f"selection_review.{key} must be false")

    rationale = fixture.get("backend_family_rationale") or {}
    require(rationale.get("selected") == "append_only_metadata_audit_log", "selected rationale mismatch")
    require(rationale.get("requires_future_backend_product_selection") is True, "future product selection must remain required")
    require(rationale.get("requires_future_storage_adapter_task_card") is True, "future adapter task card must remain required")
    require(rationale.get("requires_future_database_or_service_evidence") is True, "future backend evidence must remain required")

    requirements = rows_by_id(fixture, "future_backend_contract_requirements", "id")
    for requirement_id in {
        "append_only_write_semantics",
        "metadata_only_event_fields",
        "fail_closed_missing_runtime_dependency",
        "offline_validation_manifest",
    }:
        require(requirements.get(requirement_id, {}).get("status") == "required", f"requirement missing: {requirement_id}")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "failure_mapping", "code")
    require(set(rows) == EXPECTED_FAILURE_CODES, "failure_mapping code set mismatch")
    for code, row in rows.items():
        require(row.get("failure_boundary"), f"failure boundary missing: {code}")


def assert_diagnostics_and_side_effects(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    require(set(diagnostics.get("allowed_fields") or []) == EXPECTED_DIAGNOSTICS, "allowed diagnostic fields mismatch")
    require("secret_value" in set(diagnostics.get("forbidden_fields") or []), "forbidden diagnostics missing secret_value")

    no_fallback = fixture.get("no_fallback_policy") or {}
    for key, value in no_fallback.items():
        require(value is False, f"no_fallback_policy.{key} must be false")

    counters = fixture.get("side_effect_counters") or {}
    require(set(counters) == EXPECTED_ZERO_COUNTERS, "side effect counter set mismatch")
    for key in EXPECTED_ZERO_COUNTERS:
        require(counters.get(key) == 0, f"side_effect_counters.{key} must be 0")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("allowed_new_files") or []:
        require((REPO_ROOT / relative_path).exists(), f"allowed evidence file missing: {relative_path}")
    forbidden = set(guard.get("forbidden_artifacts") or [])
    for artifact in {
        "storage_adapter_runtime",
        "audit_store_runtime_task_card",
        "audit_store_runtime",
        "database_connection_provider",
        "db_driver",
        "sql_migration",
        "schema_marker_reader",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(artifact in forbidden, f"forbidden artifact missing: {artifact}")


def assert_validation_strategy(fixture: dict[str, Any]) -> None:
    strategy = fixture.get("validation_strategy") or {}
    require(set(strategy.get("required_checks") or []) == EXPECTED_REQUIRED_CHECKS, "required checks mismatch")


def assert_doc_references() -> None:
    for relative_path, required_terms in DOC_REFERENCES.items():
        text = read(relative_path)
        for term in required_terms:
            require(term in text, f"{relative_path} missing required term: {term}")


def assert_blocker_matrix_alignment() -> None:
    matrix = load_json(BLOCKER_MATRIX_PATH)
    boundary = matrix.get("matrix_boundary") or {}
    require(
        boundary.get("durable_backend_concrete_selection_review_status")
        == "audit_store_concrete_durable_backend_selection_review_defined",
        "blocker matrix boundary must consume concrete durable backend selection review",
    )
    require(
        boundary.get("durable_audit_backend_status")
        == "metadata_contract_artifact_readiness_defined_task_card_blocked",
        "blocker matrix durable backend status mismatch",
    )

    blockers = rows_by_id(matrix, "blocker_matrix", "blocker_id")
    durable = blockers.get("durable_audit_backend") or {}
    require(
        durable.get("status") == "metadata_contract_artifact_readiness_defined_task_card_blocked",
        "durable backend blocker status mismatch",
    )
    require(
        durable.get("source")
        == "production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1",
        "durable backend blocker source mismatch",
    )
    require(durable.get("blocks_audit_store_runtime_task_card") is True, "durable backend must still block audit runtime")
    require(durable.get("blocks_production_resolver_task_card") is True, "durable backend must still block resolver runtime")

    order = matrix.get("dependency_order") or []
    require(
        "concrete_durable_backend_selection_review" in order,
        "dependency order missing concrete durable backend selection review",
    )
    require(
        order.index("durable_backend_selection_readiness")
        < order.index("concrete_durable_backend_selection_review")
        < order.index("audit_writer_runtime_entry_review"),
        "concrete durable backend selection review must sit before writer entry review",
    )


def assert_implementation_readiness_alignment() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    require(
        target.get("audit_store_concrete_durable_backend_selection_review_status")
        == "audit_store_concrete_durable_backend_selection_review_defined",
        "implementation readiness missing concrete durable backend selection status",
    )
    require(
        target.get("durable_audit_backend_status") == "static_backend_family_selected_runtime_blocked",
        "implementation readiness durable backend status mismatch",
    )
    require(target.get("audit_store_runtime_task_card_status") == "not_created", "audit store task card must remain absent")
    require(target.get("audit_store_runtime_status") == "not_created", "audit store runtime must remain absent")
    require(target.get("production_resolver_runtime_status") == "not_created", "resolver runtime must remain absent")
    require(target.get("production_secret_backend_status") == "not_satisfied", "secret backend must remain not satisfied")

    planned = rows_by_id(readiness, "planned_slices", "id")
    row = planned.get("audit-store-concrete-durable-backend-selection-review")
    require(row is not None, "implementation readiness planned slice missing")
    require(
        row.get("status") == "audit_store_concrete_durable_backend_selection_review_defined",
        "implementation readiness planned slice status mismatch",
    )


def assert_check_repo_registration() -> None:
    text = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current = "check-production-ops-secret-backend-audit-store-concrete-durable-backend-selection-review-v1.py"
    before = "check-production-ops-secret-backend-audit-store-durable-backend-selection-readiness-v1.py"
    after = "check-production-ops-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.py"
    for script in {current, before, after}:
        require(script in text, f"check-repo.py missing {script}")
    require(text.index(before) < text.index(current) < text.index(after), "check-repo.py order mismatch")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_selection_review(fixture)
    assert_failure_mapping(fixture)
    assert_diagnostics_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_validation_strategy(fixture)
    assert_doc_references()
    assert_blocker_matrix_alignment()
    assert_implementation_readiness_alignment()
    assert_check_repo_registration()
    print("production ops secret backend audit store concrete durable backend selection review checks passed.")


if __name__ == "__main__":
    main()
