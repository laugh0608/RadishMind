#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json"
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
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.json"
        ),
        "audit_store_runtime_event_schema_materialization_readiness_defined",
    ),
    "production-secret-backend-audit-store-delivery-runtime-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-runtime-readiness-v1.json",
        "audit_store_delivery_runtime_readiness_defined",
    ),
    "production-secret-backend-audit-store-idempotency-runtime-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.json",
        "audit_store_idempotency_runtime_readiness_defined",
    ),
    "production-secret-backend-production-resolver-runtime-blocker-consolidation-v1": (
        "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.json",
        "production_resolver_runtime_blocker_consolidation_defined",
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.json",
        "credential_handle_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.json",
        "operator_approval_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-cloud-secret-service-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json",
        "cloud_secret_service_selection_readiness_defined",
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.json"
        ),
        "resolver_backend_health_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.json"
        ),
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}

EXPECTED_BOUNDARY = {
    "status": "audit_store_runtime_implementation_entry_refresh_v4_defined",
    "entry_decision": "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
    "previous_entry_refresh_v3_status": "blocked_before_runtime_task_card",
    "refresh_input_status": "runtime_readiness_chain_consumed",
    "remaining_blocker_status": "runtime_dependencies_still_blocked",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "durable_audit_backend_status": "not_selected",
    "audit_writer_runtime_status": "not_created",
    "runtime_event_schema_artifact_status": "not_created",
    "delivery_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "duplicate_detector_status": "not_created",
    "retry_executor_status": "not_created",
    "operator_approval_runtime_status": "not_created",
    "credential_handle_runtime_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "no_secret_leakage_smoke_runtime_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "cloud_secret_service_status": "not_selected",
    "database_connection_provider_status": "blocked",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "runtime_task_card_allowed_now",
    "runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "durable_audit_backend_selected_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_writer_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "runtime_event_schema_artifact_created_in_this_slice",
    "runtime_event_schema_created_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "delivery_executed_in_this_slice",
    "delivery_result_persisted_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "idempotency_key_store_created_in_this_slice",
    "duplicate_detector_created_in_this_slice",
    "duplicate_detection_executed_in_this_slice",
    "duplicate_decision_persisted_in_this_slice",
    "retry_executor_created_in_this_slice",
    "replay_executor_created_in_this_slice",
    "operator_approval_runtime_created_in_this_slice",
    "operator_approval_runtime_executed_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "credential_handle_created_in_this_slice",
    "credential_payload_created_in_this_slice",
    "backend_health_runtime_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "no_secret_leakage_smoke_runtime_created_in_this_slice",
    "no_secret_leakage_smoke_executed_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "cloud_secret_client_created_in_this_slice",
    "provider_call_executed_in_this_slice",
    "cloud_secret_call_executed_in_this_slice",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_RESOLVED = {
    "audit store runtime entry refresh v3 consumed",
    "durable backend boundary readiness defined",
    "writer runtime boundary readiness defined",
    "runtime event schema materialization readiness defined",
    "delivery runtime readiness defined",
    "idempotency runtime readiness defined",
    "credential handle runtime entry refresh consumed",
    "operator approval runtime entry refresh consumed",
    "backend health runtime entry refresh consumed",
    "no leakage smoke runtime entry refresh consumed",
    "implementation readiness alignment",
}

EXPECTED_BLOCKED = {
    "audit_store_runtime_task_card_not_created",
    "durable_audit_backend_not_selected",
    "audit_writer_runtime_not_created",
    "runtime_event_schema_artifact_not_created",
    "delivery_runtime_not_created",
    "idempotency_runtime_not_created",
    "duplicate_detector_not_created",
    "retry_executor_not_created",
    "replay_executor_not_created",
    "operator_approval_runtime_not_created",
    "credential_handle_runtime_not_created",
    "backend_health_runtime_not_created",
    "no_secret_leakage_smoke_runtime_not_created",
    "production_resolver_runtime_not_created",
    "cloud_secret_service_not_selected",
    "database_connection_provider_blocked",
    "repository_mode_disabled",
    "production_api_not_created",
}

EXPECTED_GATE_STATUS = {
    "runtime_task_card_gate": "still_blocked_before_runtime_task_card",
    "durable_audit_backend": "not_selected",
    "audit_writer_runtime": "not_created",
    "runtime_event_schema_artifact": "not_created",
    "delivery_runtime": "not_created",
    "idempotency_runtime": "not_created",
    "operator_approval_runtime": "not_created",
    "credential_handle_runtime": "not_created",
    "backend_health_runtime": "not_created",
    "no_secret_leakage_smoke_runtime": "not_created",
    "production_resolver_runtime": "not_created",
    "repository_mode": "disabled",
}

EXPECTED_REQUIREMENTS = {
    "disabled-by-default runtime gate",
    "durable audit backend selection gate",
    "writer runtime implementation gate",
    "runtime event schema artifact materialization gate",
    "delivery runtime implementation gate",
    "idempotency runtime implementation gate",
    "duplicate detector implementation gate",
    "bounded retry executor gate",
    "operator approval runtime dependency gate",
    "credential handle runtime dependency gate",
    "backend health runtime dependency gate",
    "no leakage smoke runtime dependency gate",
    "production resolver runtime dependency gate",
    "cloud secret service concrete selection gate",
    "no fallback policy",
    "no side effects counters",
    "artifact guard",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_runtime_refresh_v4_dependency_missing",
    "audit_store_runtime_refresh_v4_task_card_still_blocked",
    "audit_store_runtime_refresh_v4_durable_backend_missing",
    "audit_store_runtime_refresh_v4_writer_runtime_missing",
    "audit_store_runtime_refresh_v4_runtime_schema_missing",
    "audit_store_runtime_refresh_v4_delivery_runtime_missing",
    "audit_store_runtime_refresh_v4_idempotency_runtime_missing",
    "audit_store_runtime_refresh_v4_external_runtime_missing",
    "audit_store_runtime_refresh_v4_secret_material_detected",
    "audit_store_runtime_refresh_v4_scope_overreach",
}

EXPECTED_DIAGNOSTICS = {
    "audit_store_runtime_entry_refresh_v4_status",
    "runtime_task_decision",
    "durable_audit_backend_status",
    "audit_writer_runtime_status",
    "runtime_event_schema_artifact_status",
    "delivery_runtime_status",
    "idempotency_runtime_status",
    "operator_approval_runtime_status",
    "credential_handle_runtime_status",
    "backend_health_runtime_status",
    "no_secret_leakage_smoke_runtime_status",
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

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.md",
    "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json",
    "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.py",
}

EXPECTED_FORBIDDEN_ARTIFACTS = {
    "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-v1-plan.md",
    "services/platform/internal/secretbackend/audit_store.go",
    "services/platform/internal/secretbackend/audit_writer.go",
    "services/platform/internal/secretbackend/audit_event.go",
    "services/platform/internal/secretbackend/audit_delivery.go",
    "services/platform/internal/secretbackend/audit_idempotency.go",
    "services/platform/internal/secretbackend/duplicate_detector.go",
    "services/platform/internal/secretbackend/retry_executor.go",
    "services/platform/internal/secretbackend/production_resolver.go",
    "services/platform/internal/secretbackend/cloud_secret_client.go",
}

EXPECTED_DOC_REFERENCES = {
    "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.md": [
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
        "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
        "Artifact Guard",
    ],
    "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4-plan.md": [
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
        "Entry Requirements",
        "停止线",
    ],
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh v4",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh v4",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ],
    "docs/task-cards/README.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "audit-store-runtime-implementation-entry-refresh-v4",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ],
    "docs/radishmind-integration-contracts.md": [
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.py",
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json",
    ],
    "docs/devlogs/2026-W26.md": [
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a JSON object")
    return document


def status_of(relative_path: str) -> str:
    document = load_json(REPO_ROOT / relative_path)
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_audit_store_runtime_implementation_entry_refresh_v4",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_runtime_implementation_entry_refresh_v4_defined",
        "unexpected slice status",
    )
    require(
        slice_info.get("entry_decision") == "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
        "unexpected entry decision",
    )
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "durable_audit_backend_selected",
        "audit_writer_runtime_created",
        "runtime_event_schema_artifact_created",
        "delivery_runtime_created",
        "idempotency_runtime_created",
        "duplicate_detector_created",
        "retry_executor_created",
        "operator_approval_runtime_created",
        "credential_handle_runtime_created",
        "backend_health_runtime_created",
        "no_secret_leakage_smoke_runtime_created",
        "production_resolver_runtime_created",
        "cloud_secret_service_selected",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in claims, f"does_not_claim missing {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = {str(item.get("id")): item for item in fixture.get("depends_on") or [] if isinstance(item, dict)}
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("evidence") == path, f"{dependency_id} evidence path drifted")
        require((REPO_ROOT / path).exists(), f"{dependency_id} evidence missing on disk")
        require(item.get("status") == expected_status, f"{dependency_id} dependency status drifted")
        require(status_of(path) == expected_status, f"{dependency_id} source status drifted")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_refresh_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"entry_refresh_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"entry_refresh_boundary.{field} must remain false")


def assert_sets_and_matrices(fixture: dict[str, Any]) -> None:
    resolved = set(fixture.get("resolved_static_prerequisites") or [])
    require(EXPECTED_RESOLVED <= resolved, f"missing resolved prerequisites: {sorted(EXPECTED_RESOLVED - resolved)}")

    blocked = set(fixture.get("blocked_conditions") or [])
    require(EXPECTED_BLOCKED <= blocked, f"missing blocked conditions: {sorted(EXPECTED_BLOCKED - blocked)}")

    gates = {str(item.get("gate")): str(item.get("status")) for item in fixture.get("gate_matrix") or []}
    for gate, status in EXPECTED_GATE_STATUS.items():
        require(gates.get(gate) == status, f"gate {gate} status drifted")

    requirements = set(fixture.get("future_task_card_requirements") or [])
    require(EXPECTED_REQUIREMENTS <= requirements, "future task card requirements drifted")

    failure_codes = {str(item.get("code")) for item in fixture.get("failure_mapping") or []}
    require(failure_codes == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")


def assert_diagnostics_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    require(EXPECTED_DIAGNOSTICS <= allowed, "diagnostic allowlist drifted")
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample includes non-allowlisted fields")
    require(
        sample.get("runtime_task_decision") == "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
        "diagnostic decision drifted",
    )
    require(
        sample.get("failure_code") == "audit_store_runtime_refresh_v4_task_card_still_blocked",
        "diagnostic failure code drifted",
    )

    fallback = fixture.get("no_fallback_policy") or {}
    require(fallback.get("status") == "defined", "no fallback status drifted")
    require(fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for source in {"fake_resolver_runtime", "developer_env_plaintext", "static_idempotency_readiness"}:
        require(source in set(fallback.get("forbidden_sources") or []), f"missing forbidden source {source}")

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for name, value in counters.items():
        require(value == 0, f"side effect counter {name} must stay 0")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    allowed = set(guard.get("allowed_new_artifacts") or [])
    forbidden_kinds = set(guard.get("forbidden_artifact_kinds") or [])
    require(allowed == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifact list drifted")
    for path in EXPECTED_ALLOWED_ARTIFACTS:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing: {path}")
    for kind in {
        "audit_store_runtime_implementation_task_card",
        "durable_backend_selection_artifact",
        "writer_runtime_implementation_task_card",
        "runtime_event_schema_implementation_task_card",
        "delivery_runtime_implementation_task_card",
        "idempotency_runtime_implementation_task_card",
        "production_resolver_runtime",
        "cloud_secret_sdk_or_client",
        "database_connection_provider",
        "repository_mode_runtime",
        "public_production_api",
    }:
        require(kind in forbidden_kinds, f"forbidden artifact kind missing: {kind}")
    for path in EXPECTED_FORBIDDEN_ARTIFACTS:
        require(not (REPO_ROOT / path).exists(), f"forbidden artifact exists: {path}")


def assert_prior_evidence_alignment() -> None:
    v3 = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3"
    ][0])
    v3_boundary = v3.get("entry_refresh_boundary") or {}
    require(
        v3_boundary.get("runtime_task_decision")
        == "audit_store_runtime_implementation_still_blocked_before_task_card",
        "v3 decision drifted",
    )
    require(v3_boundary.get("audit_store_runtime_status") == "not_created", "v3 must not create runtime")

    durable = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-durable-backend-boundary-readiness-v1"
    ][0])
    durable_boundary = durable.get("durable_backend_boundary") or {}
    require(durable_boundary.get("durable_audit_backend_status") == "not_selected", "durable backend selected")

    writer = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1"
    ][0])
    writer_boundary = writer.get("writer_runtime_boundary") or {}
    require(writer_boundary.get("writer_runtime_status") == "not_created", "writer runtime created")

    schema = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1"
    ][0])
    schema_boundary = schema.get("runtime_event_schema_materialization_boundary") or {}
    require(
        schema_boundary.get("runtime_event_schema_artifact_status") == "not_created",
        "runtime schema artifact created",
    )

    delivery = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-delivery-runtime-readiness-v1"
    ][0])
    delivery_boundary = delivery.get("delivery_runtime_boundary") or {}
    require(delivery_boundary.get("delivery_runtime_status") == "not_created", "delivery runtime created")

    idempotency = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-idempotency-runtime-readiness-v1"
    ][0])
    idempotency_boundary = idempotency.get("idempotency_runtime_boundary") or {}
    require(
        idempotency_boundary.get("idempotency_runtime_status") == "not_created",
        "idempotency runtime created",
    )

    for dependency_id, expected_decision in {
        "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1": (
            "credential_handle_runtime_task_card_still_blocked_after_refresh"
        ),
        "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1": (
            "operator_approval_runtime_task_card_still_blocked_after_refresh"
        ),
        "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1": (
            "resolver_backend_health_runtime_task_card_still_blocked_after_refresh"
        ),
        "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1": (
            "real_resolver_no_secret_leakage_smoke_runtime_task_card_still_blocked_after_refresh"
        ),
    }.items():
        document = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[dependency_id][0])
        boundary = document.get("refresh_boundary") or {}
        require(boundary.get("entry_decision") == expected_decision, f"{dependency_id} decision drifted")
        require(boundary.get("runtime_task_card_allowed_now") is False, f"{dependency_id} opened runtime task card")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    require(
        alignment.get("status") == "audit_store_runtime_implementation_entry_refresh_v4_defined",
        "alignment status drifted",
    )
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in (alignment.get("target_fields") or {}).items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    planned_slice = alignment.get("planned_slice") or {}
    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or []}
    slice_id = str(planned_slice.get("id") or "")
    require(slice_id in planned, "implementation readiness missing v4 planned slice")
    require(planned[slice_id].get("status") == planned_slice.get("status"), "planned slice status drifted")
    evidence = set(planned[slice_id].get("evidence") or [])
    for path in planned_slice.get("evidence") or []:
        require(path in evidence, f"planned slice missing evidence: {path}")
        require((REPO_ROOT / path).exists(), f"planned slice evidence missing on disk: {path}")


def assert_docs_and_registration() -> None:
    for doc_path, literals in EXPECTED_DOC_REFERENCES.items():
        text = read(doc_path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{doc_path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_check = "check-production-ops-secret-backend-audit-store-idempotency-runtime-readiness-v1.py"
    current_check = "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.py"
    next_check = "check-production-ops-startup-supervisor-boundary.py"
    require(previous_check in check_repo, "check-repo missing idempotency runtime readiness check")
    require(current_check in check_repo, "check-repo missing audit store entry refresh v4 check")
    require(next_check in check_repo, "check-repo missing startup supervisor check")
    require(check_repo.index(previous_check) < check_repo.index(current_check), "v4 check must run after idempotency")
    require(check_repo.index(current_check) < check_repo.index(next_check), "v4 check must run before startup checks")


def assert_no_secret_literals() -> None:
    text = "\n".join(
        read(path)
        for path in EXPECTED_ALLOWED_ARTIFACTS
        if path.endswith(".md") or path.endswith(".json")
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "-----BEGIN"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"v4 evidence contains forbidden secret-looking literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_sets_and_matrices(fixture)
    assert_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_prior_evidence_alignment()
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store runtime implementation entry refresh v4 checks passed.")


if __name__ == "__main__":
    main()
