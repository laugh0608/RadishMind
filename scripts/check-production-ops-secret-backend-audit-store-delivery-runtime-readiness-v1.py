#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-runtime-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ),
    "production-secret-backend-audit-store-durable-backend-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.json",
        "audit_store_durable_backend_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json",
        "audit_store_writer_runtime_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.json",
        "audit_store_runtime_event_schema_materialization_readiness_defined",
    ),
    "production-secret-backend-audit-store-ownership-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-ownership-boundary-readiness-v1.json",
        "audit_store_ownership_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-contract-event-schema-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json",
        "audit_store_contract_event_schema_readiness_defined",
    ),
    "production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.json",
        "audit_store_delivery_idempotency_runtime_boundary_readiness_defined",
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
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}

EXPECTED_BOUNDARY = {
    "status": "audit_store_delivery_runtime_readiness_defined",
    "delivery_runtime_readiness_status": "defined_static_only",
    "delivery_runtime_implementation_task_card_status": "not_created",
    "delivery_runtime_status": "not_created",
    "delivery_owner_status": "static_boundary_defined",
    "delivery_input_envelope_status": "metadata_only_static_envelope_defined",
    "delivery_result_envelope_status": "metadata_only_static_envelope_defined",
    "delivery_retry_policy_status": "static_fail_closed_policy_defined",
    "delivery_duplicate_handling_status": "static_fail_closed_policy_defined",
    "idempotency_runtime_dependency_status": "separate_runtime_dependency",
    "runtime_event_schema_dependency_status": "static_boundary_defined_schema_not_created",
    "writer_runtime_dependency_status": "static_boundary_defined_writer_not_created",
    "durable_backend_dependency_status": "static_boundary_defined_backend_not_selected",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "durable_audit_backend_status": "not_selected",
    "audit_writer_status": "not_created",
    "audit_event_write_status": "not_executed",
    "audit_event_delivery_status": "not_executed",
    "production_resolver_runtime_status": "not_created",
    "cloud_secret_service_status": "not_selected",
    "cloud_secret_client_status": "not_created",
    "database_connection_provider_status": "blocked",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "runtime_task_card_created_in_this_slice",
    "delivery_runtime_task_card_created_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "delivery_executed_in_this_slice",
    "delivery_result_persisted_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "duplicate_detector_created_in_this_slice",
    "retry_runtime_created_in_this_slice",
    "writer_runtime_created_in_this_slice",
    "runtime_event_schema_created_in_this_slice",
    "runtime_event_schema_artifact_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "durable_audit_backend_selected_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "cloud_secret_client_created_in_this_slice",
    "provider_call_executed_in_this_slice",
    "cloud_secret_call_executed_in_this_slice",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_ALLOWED_METADATA = {
    "delivery_request_ref",
    "delivery_result_ref",
    "delivery_owner_ref",
    "delivery_attempt_ref",
    "delivery_mode",
    "delivery_status",
    "schema_ref",
    "schema_version",
    "writer_input_ref",
    "writer_result_ref",
    "audit_ref",
    "event_ref",
    "event_kind",
    "event_version",
    "request_id",
    "workspace_ref",
    "environment",
    "provider_profile_ref",
    "backend_profile_ref",
    "secret_ref_key",
    "credential_handle_ref",
    "approval_evidence_ref",
    "idempotency_key_ref",
    "duplicate_decision_ref",
    "retry_policy_ref",
    "durable_backend_ref",
    "retention_policy_ref",
    "redaction_profile_ref",
    "policy_version",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
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
    "raw_event_payload",
    "raw_delivery_payload",
    "raw_delivery_result",
    "delivery_payload_hash",
    "event_payload_hash",
    "secret_derived_hash",
    "developer_env_plaintext",
    "committed_secret_value",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_delivery_runtime_readiness_dependency_missing",
    "audit_store_delivery_runtime_readiness_task_card_forbidden",
    "audit_store_delivery_runtime_readiness_delivery_runtime_forbidden",
    "audit_store_delivery_runtime_readiness_delivery_execution_forbidden",
    "audit_store_delivery_runtime_readiness_event_write_forbidden",
    "audit_store_delivery_runtime_readiness_runtime_schema_forbidden",
    "audit_store_delivery_runtime_readiness_writer_runtime_forbidden",
    "audit_store_delivery_runtime_readiness_secret_material_detected",
    "audit_store_delivery_runtime_readiness_fallback_forbidden",
    "audit_store_delivery_runtime_readiness_side_effect_detected",
    "audit_store_delivery_runtime_readiness_scope_overreach",
}

EXPECTED_DIAGNOSTICS = {
    "audit_store_delivery_runtime_readiness_status",
    "delivery_runtime_implementation_task_card_status",
    "delivery_runtime_status",
    "delivery_owner_status",
    "delivery_input_envelope_status",
    "delivery_result_envelope_status",
    "delivery_retry_policy_status",
    "delivery_duplicate_handling_status",
    "idempotency_runtime_dependency_status",
    "runtime_event_schema_dependency_status",
    "writer_runtime_dependency_status",
    "durable_backend_dependency_status",
    "audit_store_runtime_task_card_status",
    "audit_store_runtime_status",
    "audit_writer_status",
    "audit_event_write_status",
    "audit_event_delivery_status",
    "production_resolver_runtime_status",
    "cloud_secret_service_status",
    "database_connection_provider_status",
    "repository_mode_status",
    "production_api_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_FORBIDDEN_SOURCES = {
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
    "static_contract",
    "static_ownership_boundary",
    "static_delivery_idempotency_boundary",
    "static_durable_backend_boundary",
    "static_writer_boundary",
    "static_runtime_schema_materialization_readiness",
    "historical_audit_event",
    "delivery_sample",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-delivery-runtime-readiness-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-delivery-runtime-readiness-v1-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-runtime-readiness-v1.json",
    "scripts/check-production-ops-secret-backend-audit-store-delivery-runtime-readiness-v1.py",
}

EXPECTED_FORBIDDEN_ARTIFACTS = {
    "audit_store_runtime_implementation_task_card",
    "delivery_runtime_implementation_task_card",
    "idempotency_runtime_implementation_task_card",
    "audit_store_runtime",
    "audit_delivery_runtime",
    "audit_idempotency_runtime",
    "duplicate_detector_runtime",
    "retry_executor",
    "audit_writer_runtime",
    "audit_writer",
    "audit_event_writer",
    "delivery_result_fixture",
    "runtime_event_schema_artifact",
    "production_resolver_runtime",
    "cloud_secret_sdk_or_client",
    "database_connection_provider",
    "sql_migration",
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
        fixture.get("kind") == "production_ops_secret_backend_audit_store_delivery_runtime_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-delivery-runtime-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_delivery_runtime_readiness_defined",
        "unexpected slice status",
    )
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "audit_store_runtime_task_card_created",
        "delivery_runtime_task_card_created",
        "delivery_runtime_created",
        "delivery_executed",
        "delivery_result_persisted",
        "idempotency_runtime_created",
        "duplicate_detector_created",
        "retry_runtime_created",
        "runtime_event_schema_created",
        "audit_writer_runtime_created",
        "audit_event_written",
        "production_resolver_runtime_created",
        "cloud_secret_client_created",
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


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("delivery_runtime_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"boundary.{field} must remain false")


def assert_metadata_and_diagnostics(fixture: dict[str, Any]) -> None:
    allowed = set(fixture.get("allowed_metadata") or [])
    require(EXPECTED_ALLOWED_METADATA <= allowed, "allowed metadata is missing required fields")
    require(not (allowed & EXPECTED_FORBIDDEN_MATERIAL), "allowed metadata contains forbidden material")

    forbidden = set(fixture.get("forbidden_material") or [])
    require(EXPECTED_FORBIDDEN_MATERIAL <= forbidden, "forbidden material coverage is incomplete")

    failure_codes = {str(item.get("code")) for item in fixture.get("failure_mapping") or []}
    require(failure_codes == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    for item in fixture.get("failure_mapping") or []:
        require(item.get("boundary"), f"{item.get('code')} missing boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{item.get('code')} missing sanitized diagnostic")
        require(not any(material in diagnostic for material in EXPECTED_FORBIDDEN_MATERIAL), "unsafe diagnostic")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed_fields = set(diagnostics.get("allowed_fields") or [])
    forbidden_fields = set(diagnostics.get("forbidden_fields") or [])
    require(EXPECTED_DIAGNOSTICS <= allowed_fields, "diagnostic allowlist missing fields")
    require(EXPECTED_FORBIDDEN_MATERIAL & forbidden_fields, "diagnostic forbidden fields missing secret material")
    require(not (allowed_fields & forbidden_fields), "diagnostic allowlist intersects forbidden fields")


def assert_policies_and_artifacts(fixture: dict[str, Any]) -> None:
    fallback = fixture.get("no_fallback_policy") or {}
    require(fallback.get("status") == "defined", "no fallback policy status drifted")
    sources = set(fallback.get("forbidden_sources") or [])
    require(EXPECTED_FORBIDDEN_SOURCES <= sources, "no fallback forbidden source coverage drifted")

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for name, value in counters.items():
        require(value == 0, f"side effect counter {name} must stay 0")

    guard = fixture.get("artifact_guard") or {}
    allowed_artifacts = set(guard.get("allowed_added_artifacts") or [])
    forbidden_artifacts = set(guard.get("forbidden_artifact_kinds") or [])
    require(allowed_artifacts == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifact list drifted")
    require(EXPECTED_FORBIDDEN_ARTIFACTS <= forbidden_artifacts, "forbidden artifact list missing entries")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    require(
        alignment.get("status") == "audit_store_delivery_runtime_readiness_defined",
        "alignment status drifted",
    )
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    expected_fields = {
        "production_secret_backend_status": "not_satisfied",
        "audit_store_delivery_runtime_readiness_status": "defined_without_delivery_runtime",
        "audit_delivery_runtime_owner_status": "static_boundary_defined",
        "audit_delivery_input_envelope_status": "metadata_only_static_envelope_defined",
        "audit_delivery_result_reference_status": "metadata_only_static_reference_defined",
        "audit_delivery_retry_policy_status": "static_fail_closed_policy_defined",
        "audit_delivery_duplicate_handling_status": "static_fail_closed_policy_defined",
        "audit_delivery_runtime_status": "not_created",
        "audit_idempotency_runtime_status": "not_created",
        "audit_store_runtime_task_card_status": "not_created",
        "audit_store_runtime_status": "not_created",
        "audit_writer_status": "not_created",
        "audit_event_delivery_status": "not_executed",
    }
    for field, expected in expected_fields.items():
        require(alignment.get(field) == expected, f"alignment.{field} drifted")
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or []}
    planned_item = planned.get("audit-store-delivery-runtime-readiness") or {}
    require(
        planned_item.get("status") == "audit_store_delivery_runtime_readiness_defined",
        "implementation readiness planned slice missing delivery runtime readiness",
    )
    require(
        EXPECTED_ALLOWED_ARTIFACTS <= set(planned_item.get("evidence") or []),
        "implementation readiness planned slice missing delivery runtime readiness evidence",
    )


def assert_docs_and_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_check = (
        'run_python_script("'
        'check-production-ops-secret-backend-audit-store-delivery-runtime-readiness-v1.py", [])'
    )
    require(current_check in check_repo, "check-repo.py must run delivery runtime readiness check")
    previous_check = "check-production-ops-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.py"
    require(
        check_repo.index(previous_check) < check_repo.index(current_check),
        "delivery runtime readiness check must run after runtime event schema materialization check",
    )

    docs = {
        "docs/platform/production-secret-backend-audit-store-delivery-runtime-readiness-v1.md": [
            "audit_store_delivery_runtime_readiness_defined",
            "Delivery Runtime Boundary",
            "Allowed Metadata",
            "Forbidden Material",
            "No Fallback / No Side Effects",
            "Artifact Guard",
        ],
        "docs/task-cards/production-secret-backend-audit-store-delivery-runtime-readiness-v1-plan.md": [
            "audit_store_delivery_runtime_readiness_defined",
            "Boundary Requirements",
            "停止线",
        ],
        "docs/platform/README.md": [
            "delivery runtime readiness",
            "audit_store_delivery_runtime_readiness_defined",
        ],
        "docs/radishmind-current-focus.md": [
            "audit_store_delivery_runtime_readiness_defined",
            "production-secret-backend-audit-store-delivery-runtime-readiness-v1",
        ],
        "docs/features/README.md": [
            "Production Secret Backend Audit Store Delivery Runtime Readiness v1",
            "audit_store_delivery_runtime_readiness_defined",
        ],
        "docs/radishmind-integration-contracts.md": [
            "audit_store_delivery_runtime_readiness_defined",
        ],
        "docs/task-cards/README.md": [
            "production-secret-backend-audit-store-delivery-runtime-readiness-v1",
            "audit_store_delivery_runtime_readiness_defined",
        ],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
            "audit-store-delivery-runtime-readiness",
            "audit_store_delivery_runtime_readiness_defined",
        ],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-delivery-runtime-readiness-v1.py",
            "production-secret-backend-audit-store-delivery-runtime-readiness-v1.json",
        ],
        "docs/devlogs/2026-W26.md": [
            "audit_store_delivery_runtime_readiness_defined",
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
    require(not found, f"delivery runtime readiness evidence contains forbidden literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_metadata_and_diagnostics(fixture)
    assert_policies_and_artifacts(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store delivery runtime readiness checks passed.")


if __name__ == "__main__":
    main()
