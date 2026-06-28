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
    "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
BLOCKER_MATRIX_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json",
        "audit_store_writer_runtime_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-runtime-event-schema-artifact-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.json",
        "audit_store_runtime_event_schema_artifact_implemented",
    ),
    "production-secret-backend-audit-store-runtime-blocker-matrix-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json",
        "audit_store_runtime_blocker_matrix_defined",
    ),
    "production-secret-backend-audit-store-durable-backend-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-selection-readiness-v1.json",
        "audit_store_durable_backend_selection_readiness_defined",
    ),
    "production-secret-backend-audit-store-delivery-runtime-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-runtime-readiness-v1.json",
        "audit_store_delivery_runtime_readiness_defined",
    ),
    "production-secret-backend-audit-store-idempotency-runtime-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.json",
        "audit_store_idempotency_runtime_readiness_defined",
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.json",
        "credential_handle_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.json",
        "operator_approval_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.json",
        "resolver_backend_health_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.json",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined",
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

EXPECTED_ENTRY_FIELDS = {
    "entry_review_status": "blocked_before_runtime_task_card",
    "runtime_task_decision": "audit_store_writer_runtime_task_card_blocked_after_selection_readiness",
    "writer_boundary_status": "audit_store_writer_runtime_boundary_readiness_defined",
    "writer_runtime_implementation_status": "not_started",
    "writer_runtime_task_card_status": "not_created",
    "writer_runtime_status": "not_created",
    "writer_result_status": "not_created",
    "event_schema_artifact_status": "implemented_static_schema_artifact",
    "event_schema_artifact_validation_status": "implemented_offline_schema_validation",
    "durable_backend_selection_status": "audit_store_durable_backend_selection_readiness_defined",
    "durable_audit_backend_status": "not_selected",
    "storage_adapter_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "delivery_runtime_status": "not_created",
    "operator_approval_runtime_status": "not_created",
    "operator_approval_runtime_execution_status": "not_executed",
    "credential_handle_runtime_status": "not_created",
    "credential_payload_status": "forbidden",
    "backend_health_runtime_status": "not_created",
    "backend_health_check_status": "not_executed",
    "no_secret_leakage_smoke_runtime_status": "not_created",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_store_status": "not_created",
    "audit_writer_status": "not_created",
    "audit_event_write_status": "not_executed",
    "audit_event_delivery_status": "not_executed",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "cloud_secret_service_status": "not_selected",
    "cloud_secret_client_status": "not_created",
    "database_connection_provider_status": "blocked",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "runtime_task_card_created_in_this_slice",
    "writer_runtime_task_card_created_in_this_slice",
    "writer_runtime_created_in_this_slice",
    "writer_result_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "durable_audit_backend_selected_in_this_slice",
    "storage_adapter_runtime_created_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "operator_approval_runtime_created_in_this_slice",
    "operator_approval_runtime_executed_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "credential_payload_created_in_this_slice",
    "backend_health_runtime_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "no_secret_leakage_smoke_runtime_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "provider_call_executed_in_this_slice",
    "cloud_secret_call_executed_in_this_slice",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_BLOCKED = {
    "durable_audit_backend_not_selected",
    "writer_runtime_task_card_not_created",
    "writer_runtime_not_created",
    "writer_result_runtime_forbidden",
    "idempotency_runtime_not_created",
    "delivery_runtime_not_created",
    "operator_approval_runtime_not_created",
    "credential_handle_runtime_not_created",
    "backend_health_runtime_not_created",
    "real_no_leakage_smoke_runtime_not_created",
    "audit_store_runtime_not_created",
    "production_resolver_runtime_not_created",
    "repository_mode_disabled",
}

EXPECTED_GATE_STATUS = {
    "writer_boundary_consumed": "satisfied_static_boundary",
    "runtime_event_schema_artifact": "satisfied_static_schema_artifact",
    "durable_backend_selection": "blocked_backend_not_selected",
    "writer_runtime_task_card_gate": "blocked_after_selection_readiness",
    "writer_runtime": "not_created",
    "writer_result": "not_created",
    "idempotency_runtime": "blocked_runtime_not_created",
    "delivery_runtime": "blocked_runtime_not_created",
    "operator_approval_runtime": "blocked_runtime_not_created",
    "credential_handle_runtime": "blocked_runtime_not_created",
    "backend_health_runtime": "blocked_runtime_not_created",
    "no_secret_leakage_smoke_runtime": "blocked_runtime_not_created",
    "audit_store_runtime": "not_created",
    "production_resolver_runtime": "not_created",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_RUNTIME_REQUIREMENTS = {
    "disabled-by-default writer runtime gate",
    "metadata-only writer input envelope",
    "committed runtime event schema artifact pin",
    "durable backend dependency gate",
    "idempotency runtime dependency gate",
    "delivery runtime dependency gate",
    "retention and redaction policy binding",
    "credential handle runtime dependency gate",
    "operator approval runtime dependency gate",
    "backend health runtime dependency gate",
    "no leakage smoke runtime dependency gate",
    "sanitized diagnostics allowlist",
    "forbidden material scan",
    "offline unit test and static smoke",
    "side effect counters",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_writer_runtime_entry_boundary_missing",
    "audit_store_writer_runtime_entry_schema_artifact_missing",
    "audit_store_writer_runtime_entry_task_card_blocked",
    "audit_store_writer_runtime_entry_durable_backend_missing",
    "audit_store_writer_runtime_entry_idempotency_runtime_missing",
    "audit_store_writer_runtime_entry_delivery_runtime_missing",
    "audit_store_writer_runtime_entry_operator_approval_missing",
    "audit_store_writer_runtime_entry_credential_handle_missing",
    "audit_store_writer_runtime_entry_backend_health_missing",
    "audit_store_writer_runtime_entry_no_leakage_missing",
    "audit_store_writer_runtime_entry_secret_material_detected",
    "audit_store_writer_runtime_created_in_entry_review",
    "audit_store_writer_result_created_in_entry_review",
    "audit_store_writer_runtime_entry_fallback_forbidden",
    "audit_store_writer_runtime_entry_repository_mode_forbidden",
    "audit_store_writer_runtime_entry_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "audit_store_writer_runtime_entry_status",
    "writer_boundary_status",
    "runtime_task_decision",
    "writer_runtime_status",
    "writer_result_status",
    "event_schema_artifact_status",
    "durable_backend_selection_status",
    "durable_audit_backend_status",
    "idempotency_runtime_status",
    "delivery_runtime_status",
    "operator_approval_runtime_status",
    "credential_handle_runtime_status",
    "backend_health_runtime_status",
    "no_secret_leakage_runtime_status",
    "audit_store_runtime_status",
    "production_resolver_runtime_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_FORBIDDEN_MATERIAL = {
    "raw_secret",
    "secret_value",
    "credential_payload",
    "provider_raw_url",
    "dsn",
    "cloud_credential",
    "raw_request_payload",
    "raw_response_payload",
    "raw_audit_payload",
    "raw_writer_payload",
    "raw_payload_hash",
    "secret_derived_hash",
    "developer_env_plaintext",
    "committed_secret_value",
}

EXPECTED_FALLBACK_SOURCES = {
    "fake_resolver_runtime",
    "developer_env_plaintext",
    "fixture_credential",
    "committed_value",
    "sample",
    "mock_provider",
    "local_smoke_profile",
    "operator_runbook_text",
    "repository_memory_store",
    "audit_memory_store",
    "static_handoff_envelope",
    "static_writer_boundary",
    "static_schema_artifact",
    "static_durable_backend_boundary",
    "durable_backend_selection_readiness",
    "test_writer",
    "memory_writer",
    "schema_fixture_writer",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.json",
    "scripts/check-production-ops-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.py",
}

EXPECTED_FORBIDDEN_ARTIFACTS = {
    "writer_runtime_implementation_task_card",
    "audit_writer_runtime",
    "audit_writer",
    "audit_event_writer",
    "writer_result_fixture",
    "durable_audit_backend",
    "storage_adapter_runtime",
    "audit_store_runtime_implementation_task_card",
    "audit_store_runtime",
    "audit_delivery_runtime",
    "audit_idempotency_runtime",
    "production_resolver_runtime",
    "cloud_secret_sdk_or_client",
    "database_connection_provider",
    "repository_mode_runtime",
    "public_production_api",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def status_of(path: str) -> str:
    document = load_json(REPO_ROOT / path)
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_writer_runtime_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id")
        == "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_writer_runtime_implementation_entry_review_defined",
        "unexpected slice status",
    )
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "writer_runtime_task_card_created",
        "audit_writer_runtime_created",
        "writer_result_created",
        "durable_audit_backend_selected",
        "audit_store_runtime_task_card_created",
        "production_resolver_runtime_created",
        "database_connection_provider_ready",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in forbidden_claims, f"does_not_claim missing {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = {str(item.get("id")): item for item in fixture.get("depends_on") or []}
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("evidence") == path, f"{dependency_id} evidence path drifted")
        require((REPO_ROOT / path).exists(), f"{dependency_id} evidence missing on disk")
        require(item.get("status") == expected_status, f"{dependency_id} fixture status drifted")
        require(status_of(path) == expected_status, f"{dependency_id} source status drifted")


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_boundary") or {}
    for field, expected in EXPECTED_ENTRY_FIELDS.items():
        require(boundary.get(field) == expected, f"entry_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"entry_boundary.{field} must remain false")

    blocked = set(fixture.get("blocked_conditions") or [])
    require(EXPECTED_BLOCKED <= blocked, "blocked condition coverage drifted")

    gate_matrix = {
        str(item.get("gate")): item.get("status") for item in fixture.get("gate_matrix") or []
    }
    require(gate_matrix == EXPECTED_GATE_STATUS, "gate matrix drifted")

    requirements = set(fixture.get("future_runtime_task_card_requirements") or [])
    require(EXPECTED_RUNTIME_REQUIREMENTS <= requirements, "runtime task card requirements drifted")


def assert_failure_diagnostics_and_policies(fixture: dict[str, Any]) -> None:
    failure_codes = {str(item.get("code")) for item in fixture.get("failure_mapping") or []}
    require(failure_codes == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for item in fixture.get("failure_mapping") or []:
        require(item.get("boundary"), f"{item.get('code')} missing boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{item.get('code')} missing sanitized diagnostic")
        require(
            not any(material in diagnostic for material in EXPECTED_FORBIDDEN_MATERIAL),
            f"{item.get('code')} diagnostic contains forbidden material",
        )

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed_fields = set(diagnostics.get("allowed_fields") or [])
    forbidden_fields = set(diagnostics.get("forbidden_fields") or [])
    require(EXPECTED_DIAGNOSTIC_FIELDS <= allowed_fields, "diagnostic allowlist missing fields")
    require(EXPECTED_FORBIDDEN_MATERIAL <= forbidden_fields, "diagnostic forbidden fields incomplete")
    require(not (allowed_fields & forbidden_fields), "diagnostic allowlist intersects forbidden fields")

    fallback = fixture.get("no_fallback_policy") or {}
    require(fallback.get("status") == "defined", "no fallback policy status drifted")
    require(
        EXPECTED_FALLBACK_SOURCES <= set(fallback.get("forbidden_sources") or []),
        "no fallback forbidden source coverage drifted",
    )

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for name, value in counters.items():
        require(value == 0, f"side effect counter {name} must stay 0")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    allowed_artifacts = set(guard.get("allowed_added_artifacts") or [])
    forbidden_artifacts = set(guard.get("forbidden_artifact_kinds") or [])
    require(allowed_artifacts == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifact list drifted")
    for path in allowed_artifacts:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing on disk: {path}")
    require(EXPECTED_FORBIDDEN_ARTIFACTS <= forbidden_artifacts, "forbidden artifact list missing entries")


def assert_blocker_matrix_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("blocker_matrix_alignment") or {}
    require(
        alignment.get("status") == "writer_runtime_entry_review_defined_task_card_blocked",
        "blocker matrix alignment status drifted",
    )
    require(
        alignment.get("writer_blocker_current_status") == "entry_review_defined_task_card_blocked",
        "writer blocker current status drifted",
    )
    matrix = load_json(BLOCKER_MATRIX_PATH)
    blockers = {str(item.get("blocker_id")): item for item in matrix.get("blocker_matrix") or []}
    writer = blockers.get("audit_writer_runtime") or {}
    require(
        writer.get("status") == "entry_review_defined_task_card_blocked",
        "blocker matrix writer status not updated",
    )
    require(
        writer.get("source") == "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1",
        "blocker matrix writer source not updated",
    )
    boundary = matrix.get("matrix_boundary") or {}
    require(
        boundary.get("writer_runtime_implementation_entry_review_status")
        == "audit_store_writer_runtime_implementation_entry_review_defined",
        "matrix boundary missing writer entry review status",
    )


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    require(
        alignment.get("status") == "audit_store_writer_runtime_implementation_entry_review_defined",
        "implementation readiness alignment status drifted",
    )
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    expected_fields = {
        "production_secret_backend_status": "not_satisfied",
        "audit_store_writer_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "audit_store_writer_runtime_boundary_readiness_status": "defined_without_writer_runtime",
        "audit_runtime_event_schema_artifact_status": "implemented_static_schema_artifact",
        "audit_store_durable_backend_selection_readiness_status": "defined_without_backend_selection",
        "durable_audit_backend_status": "not_selected",
        "audit_writer_runtime_owner_status": "static_boundary_defined",
        "audit_writer_input_envelope_status": "metadata_only_static_envelope_defined",
        "audit_writer_result_reference_status": "metadata_only_static_reference_defined",
        "audit_store_runtime_task_card_status": "not_created",
        "audit_store_runtime_status": "not_created",
        "audit_writer_status": "not_created",
        "audit_event_delivery_status": "not_executed",
    }
    for field, expected in expected_fields.items():
        require(alignment.get(field) == expected, f"alignment.{field} drifted")
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or []}
    planned_item = planned.get("audit-store-writer-runtime-implementation-entry-review") or {}
    require(
        planned_item.get("status") == "audit_store_writer_runtime_implementation_entry_review_defined",
        "implementation readiness planned slice missing writer runtime entry review",
    )
    require(
        EXPECTED_ALLOWED_ARTIFACTS <= set(planned_item.get("evidence") or []),
        "implementation readiness planned slice missing writer runtime entry review evidence",
    )


def assert_docs_and_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_check = (
        'run_python_script("'
        'check-production-ops-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.py", [])'
    )
    require(current_check in check_repo, "check-repo.py must run writer runtime entry review check")
    previous_check = "check-production-ops-secret-backend-audit-store-durable-backend-selection-readiness-v1.py"
    blocker_check = "check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py"
    require(
        check_repo.index(previous_check) < check_repo.index(current_check),
        "writer runtime entry review check must run after durable backend selection readiness",
    )
    require(
        check_repo.index(current_check) < check_repo.index(blocker_check),
        "writer runtime entry review check must run before blocker matrix refresh",
    )

    docs = {
        "docs/platform/production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1.md": [
            "audit_store_writer_runtime_implementation_entry_review_defined",
            "audit_store_writer_runtime_task_card_blocked_after_selection_readiness",
            "Future Runtime Task Card Requirements",
            "Blocker Matrix Alignment",
            "停止线",
        ],
        "docs/task-cards/production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1-plan.md": [
            "audit_store_writer_runtime_implementation_entry_review_defined",
            "准入边界",
            "停止线",
        ],
        "docs/platform/production-secret-backend-audit-store-runtime-blocker-matrix-v1.md": [
            "audit_store_writer_runtime_implementation_entry_review_defined",
            "entry_review_defined_task_card_blocked",
        ],
        "docs/platform/README.md": [
            "audit_store_writer_runtime_implementation_entry_review_defined",
            "Production Secret Backend Audit Store Writer Runtime Implementation Entry Review v1",
        ],
        "docs/radishmind-current-focus.md": [
            "audit_store_writer_runtime_implementation_entry_review_defined",
            "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1",
        ],
        "docs/task-cards/README.md": [
            "production-secret-backend-audit-store-writer-runtime-implementation-entry-review-v1",
            "audit_store_writer_runtime_implementation_entry_review_defined",
        ],
        "docs/devlogs/2026-W26.md": [
            "audit_store_writer_runtime_implementation_entry_review_defined",
        ],
    }
    for path, literals in docs.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")


def assert_no_secret_literals() -> None:
    text = "\n".join(
        read(path)
        for path in EXPECTED_ALLOWED_ARTIFACTS
        if path.endswith(".md") or path.endswith(".json")
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "-----BEGIN"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"writer runtime entry review evidence contains forbidden literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")


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
    print(
        "production ops secret backend audit store writer runtime implementation entry review checks passed."
    )


if __name__ == "__main__":
    main()
