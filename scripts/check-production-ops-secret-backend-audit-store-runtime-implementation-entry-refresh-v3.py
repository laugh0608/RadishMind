#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
AUDIT_ENTRY_REFRESH_V2_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.json"
)
AUDIT_OWNERSHIP_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-ownership-boundary-readiness-v1.json"
)
AUDIT_DELIVERY_IDEMPOTENCY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-handoff-readiness-v1": "audit_store_handoff_readiness_defined",
    "production-secret-backend-audit-store-contract-event-schema-readiness-v1": (
        "audit_store_contract_event_schema_readiness_defined"
    ),
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2": (
        "audit_store_runtime_implementation_entry_refresh_v2_defined"
    ),
    "production-secret-backend-audit-store-ownership-boundary-readiness-v1": (
        "audit_store_ownership_boundary_readiness_defined"
    ),
    "production-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1": (
        "audit_store_delivery_idempotency_runtime_boundary_readiness_defined"
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1": (
        "operator_approval_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1": (
        "credential_handle_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1": (
        "resolver_backend_health_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1": (
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
}

EXPECTED_BOUNDARY_FIELDS = {
    "entry_refresh_status": "blocked_before_runtime_task_card",
    "runtime_task_decision": "audit_store_runtime_implementation_still_blocked_before_task_card",
    "previous_entry_refresh_v2_status": "blocked_before_runtime_task_card",
    "refresh_input_status": "ownership_and_delivery_idempotency_boundaries_consumed",
    "audit_handoff_status": "defined_without_store_runtime",
    "audit_store_contract_status": "static_contract_defined",
    "ownership_boundary_status": "defined_without_store_runtime",
    "delivery_idempotency_boundary_status": "defined_without_delivery_runtime",
    "resolved_static_prerequisites_status": "handoff_contract_ownership_delivery_idempotency_satisfied",
    "unresolved_runtime_dependency_status": "runtime_dependencies_still_blocked",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "durable_audit_backend_status": "not_selected",
    "audit_writer_status": "not_created",
    "runtime_event_schema_status": "not_created",
    "audit_event_delivery_status": "not_executed",
    "delivery_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "store_ownership_status": "static_reference_defined",
    "writer_ownership_status": "separated_static_boundary",
    "runtime_event_schema_owner_status": "static_reference_defined",
    "delivery_owner_status": "static_reference_defined",
    "idempotency_key_owner_status": "static_reference_defined",
    "duplicate_handling_status": "static_fail_closed_policy_defined",
    "retry_failure_semantics_status": "static_fail_closed_policy_defined",
    "delivery_result_envelope_status": "metadata_only_static_envelope_defined",
    "metadata_only_diagnostics_status": "fixed_sanitized_only",
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
    "runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "durable_audit_backend_selected_in_this_slice",
    "audit_writer_created_in_this_slice",
    "runtime_event_schema_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "audit_delivery_executed_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "delivery_result_persisted_in_this_slice",
    "provider_call_executed_in_this_slice",
    "cloud_secret_call_executed_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "credential_payload_created_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "operator_approval_runtime_created_in_this_slice",
    "operator_approval_runtime_executed_in_this_slice",
    "backend_health_runtime_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "no_secret_leakage_smoke_runtime_created_in_this_slice",
    "cloud_secret_service_enabled",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_RESOLVED = {
    "reference-only audit store handoff",
    "metadata-only audit event schema",
    "reference-only writer input output contract",
    "event version reference",
    "event kind allowlist",
    "idempotency key reference",
    "delivery result envelope",
    "retention policy reference",
    "redaction profile reference",
    "store owner reference",
    "writer owner reference",
    "runtime event schema owner reference",
    "delivery owner reference",
    "idempotency key owner reference",
    "duplicate handling fail-closed policy",
    "retry failure fail-closed semantics",
    "metadata-only diagnostics allowlist",
    "failure mapping",
    "no fallback policy",
    "no side effects policy",
    "artifact guard",
}

EXPECTED_BLOCKED = {
    "audit_store_runtime_task_card_not_created",
    "durable_audit_backend_not_selected",
    "audit_writer_runtime_not_created",
    "runtime_event_schema_not_materialized",
    "audit_event_delivery_runtime_not_created",
    "idempotency_runtime_not_created",
    "operator_approval_runtime_not_created",
    "credential_handle_runtime_not_created",
    "backend_health_runtime_not_created",
    "real_no_leakage_smoke_runtime_not_created",
    "production_resolver_runtime_not_created",
    "cloud_secret_service_not_selected",
    "repository_mode_disabled",
}

EXPECTED_GATE_STATUS = {
    "audit_handoff_boundary": "satisfied_static_boundary",
    "contract_event_schema": "satisfied_static_contract",
    "ownership_boundary": "satisfied_static_boundary",
    "delivery_idempotency_boundary": "satisfied_static_boundary",
    "runtime_task_card_gate": "still_blocked_before_task_card",
    "durable_audit_backend": "not_selected",
    "audit_writer_runtime": "not_created",
    "runtime_event_schema": "not_created",
    "audit_event_delivery": "not_executed",
    "delivery_runtime": "not_created",
    "idempotency_runtime": "not_created",
    "operator_approval_runtime": "blocked_runtime_not_created",
    "credential_handle_runtime": "blocked_runtime_not_created",
    "backend_health_runtime": "blocked_runtime_not_created",
    "no_secret_leakage_smoke_runtime": "blocked_runtime_not_created",
    "production_resolver_runtime": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_RUNTIME_REQUIREMENTS = {
    "disabled-by-default runtime gate",
    "durable audit backend selection gate",
    "store ownership boundary",
    "writer ownership boundary",
    "metadata-only runtime event schema materialization",
    "reference-only writer input",
    "idempotency key reference",
    "duplicate handling fail-closed policy",
    "bounded retry policy reference",
    "fail-closed delivery semantics",
    "delivery result envelope",
    "retention policy binding",
    "redaction profile binding",
    "operator approval runtime dependency gate",
    "credential handle runtime dependency gate",
    "backend health runtime dependency gate",
    "no leakage smoke runtime dependency gate",
    "rotation policy binding",
    "runbook binding",
    "environment binding",
    "provider profile binding",
    "backend profile binding",
    "secret ref binding",
    "sanitized diagnostics allowlist",
    "offline unit test and static smoke",
    "side effect counters",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_runtime_refresh_v3_ownership_missing",
    "audit_store_runtime_refresh_v3_delivery_idempotency_missing",
    "audit_store_runtime_refresh_v3_task_card_still_blocked",
    "audit_store_runtime_refresh_v3_durable_backend_missing",
    "audit_store_runtime_refresh_v3_writer_runtime_missing",
    "audit_store_runtime_refresh_v3_runtime_schema_missing",
    "audit_store_runtime_refresh_v3_delivery_runtime_missing",
    "audit_store_runtime_refresh_v3_idempotency_runtime_missing",
    "audit_store_runtime_refresh_v3_operator_approval_runtime_missing",
    "audit_store_runtime_refresh_v3_credential_handle_runtime_missing",
    "audit_store_runtime_refresh_v3_backend_health_runtime_missing",
    "audit_store_runtime_refresh_v3_no_leakage_runtime_missing",
    "audit_store_runtime_refresh_v3_secret_material_detected",
    "audit_store_runtime_created_in_refresh_v3",
    "audit_writer_created_in_refresh_v3",
    "audit_event_written_in_refresh_v3",
    "audit_delivery_executed_in_refresh_v3",
    "audit_store_runtime_refresh_v3_fallback_forbidden",
    "audit_store_runtime_refresh_v3_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "audit_store_runtime_entry_refresh_v3_status",
    "runtime_task_decision",
    "audit_handoff_status",
    "audit_store_contract_status",
    "ownership_boundary_status",
    "delivery_idempotency_boundary_status",
    "audit_store_runtime_task_card_status",
    "audit_store_runtime_status",
    "audit_writer_status",
    "runtime_event_schema_status",
    "audit_event_delivery_status",
    "durable_audit_backend_status",
    "delivery_runtime_status",
    "idempotency_runtime_status",
    "operator_approval_runtime_status",
    "credential_handle_runtime_status",
    "backend_health_runtime_status",
    "no_secret_leakage_runtime_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_SIDE_EFFECT_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "audit_store_runtime_created_count",
    "audit_writer_created_count",
    "runtime_event_schema_created_count",
    "audit_event_write_count",
    "audit_delivery_execution_count",
    "idempotency_runtime_created_count",
    "delivery_result_persistence_count",
    "operator_approval_runtime_execution_count",
    "credential_payload_created_count",
    "credential_handle_created_count",
    "backend_health_check_count",
    "no_secret_leakage_smoke_runtime_execution_count",
    "network_call_count",
    "cloud_secret_call_count",
    "provider_call_count",
    "database_connection_count",
    "driver_open_count",
    "sql_execution_count",
    "schema_marker_read_count",
    "schema_marker_write_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}

EXPECTED_STATIC_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.md",
    "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json",
    "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py",
}

EXPECTED_DOC_REFERENCES = {
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh v3",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh v3",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ],
    "docs/task-cards/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py",
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json",
    ],
    "docs/devlogs/2026-W25.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ],
    "docs/devlogs/2026-W25.parts/2026-06-19-to-2026-06-21.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
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


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_audit_store_runtime_implementation_entry_refresh_v3",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_runtime_implementation_entry_refresh_v3_defined",
        "unexpected slice status",
    )
    for path in [slice_info.get("task_card"), slice_info.get("platform_topic")]:
        require(isinstance(path, str) and (REPO_ROOT / path).exists(), f"missing slice artifact: {path}")
    for forbidden_claim in {
        "audit_store_runtime_created",
        "audit_store_runtime_task_card_created",
        "audit_writer_created",
        "audit_event_written",
        "audit_delivery_executed",
        "delivery_runtime_created",
        "idempotency_runtime_created",
        "durable_audit_backend_selected",
        "production_resolver_runtime_created",
        "production_api_ready",
    }:
        require(forbidden_claim in set(slice_info.get("does_not_claim") or []), f"missing claim: {forbidden_claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = {str(item.get("id")): item for item in fixture.get("depends_on") or [] if isinstance(item, dict)}
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, expected_status in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        evidence = str(item.get("evidence") or "")
        require(evidence, f"{dependency_id} missing evidence")
        require((REPO_ROOT / evidence).exists(), f"{dependency_id} evidence missing: {evidence}")


def assert_entry_refresh_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_refresh_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY_FIELDS.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"{field} must be false")


def assert_sets_and_matrices(fixture: dict[str, Any]) -> None:
    resolved = set(fixture.get("resolved_static_prerequisites") or [])
    require(EXPECTED_RESOLVED <= resolved, f"missing resolved prerequisites: {sorted(EXPECTED_RESOLVED - resolved)}")

    blocked = set(fixture.get("blocked_conditions") or [])
    require(EXPECTED_BLOCKED <= blocked, f"missing blocked conditions: {sorted(EXPECTED_BLOCKED - blocked)}")

    gate_status = {
        str(item.get("gate")): str(item.get("status"))
        for item in fixture.get("gate_matrix") or []
        if isinstance(item, dict)
    }
    for gate, expected_status in EXPECTED_GATE_STATUS.items():
        require(gate_status.get(gate) == expected_status, f"{gate} status drifted")

    requirements = set(fixture.get("future_task_card_requirements") or [])
    missing_requirements = sorted(EXPECTED_RUNTIME_REQUIREMENTS - requirements)
    require(not missing_requirements, f"missing future task card requirements: {missing_requirements}")

    failure_codes = {str(item.get("code")) for item in fixture.get("failure_mapping") or [] if isinstance(item, dict)}
    missing_failures = sorted(EXPECTED_FAILURE_CODES - failure_codes)
    require(not missing_failures, f"missing failure codes: {missing_failures}")


def assert_diagnostics_and_no_side_effects(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    require(EXPECTED_DIAGNOSTIC_FIELDS <= allowed, "sanitized diagnostics allowlist drifted")
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= allowed, "diagnostic sample contains non-allowlisted fields")
    require(
        sample.get("runtime_task_decision") == "audit_store_runtime_implementation_still_blocked_before_task_card",
        "diagnostic runtime decision drifted",
    )
    require(sample.get("failure_code") == "audit_store_runtime_refresh_v3_task_card_still_blocked", "bad sample failure")

    no_fallback = fixture.get("no_fallback_policy") or {}
    require(no_fallback.get("fallback_policy") == "forbidden", "fallback must be forbidden")
    require(no_fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for fallback in {"fake_resolver_runtime", "static_contract", "static_ownership_boundary", "audit_memory_store"}:
        require(fallback in set(no_fallback.get("forbidden_fallbacks") or []), f"missing fallback {fallback}")

    side_effects = fixture.get("no_side_effects") or {}
    require(side_effects.get("checker_mode") == "committed_static_artifacts_only", "unexpected checker mode")
    for counter in EXPECTED_SIDE_EFFECT_COUNTERS:
        require(side_effects.get(counter) == 0, f"{counter} must remain 0")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    allowed = set(guard.get("allowed_new_artifacts") or [])
    require(EXPECTED_STATIC_ARTIFACTS <= allowed, "allowed static artifacts drifted")
    for path in EXPECTED_STATIC_ARTIFACTS:
        require((REPO_ROOT / path).exists(), f"static artifact missing: {path}")
    for forbidden_path in {
        "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-v1-plan.md",
        "services/platform/internal/secretbackend/audit_store.go",
        "services/platform/internal/secretbackend/audit_writer.go",
        "services/platform/internal/secretbackend/audit_event.go",
        "services/platform/internal/secretbackend/audit_delivery.go",
        "services/platform/internal/secretbackend/audit_idempotency.go",
    }:
        require(not (REPO_ROOT / forbidden_path).exists(), f"forbidden artifact exists: {forbidden_path}")


def assert_prior_evidence_alignment() -> None:
    entry_refresh_v2 = load_json(AUDIT_ENTRY_REFRESH_V2_PATH)
    refresh_boundary = entry_refresh_v2.get("entry_refresh_boundary") or {}
    require(
        refresh_boundary.get("runtime_task_decision")
        == "audit_store_runtime_implementation_still_blocked_before_task_card",
        "audit store entry refresh v2 decision drifted",
    )
    require(refresh_boundary.get("audit_store_runtime_status") == "not_created", "v2 must not create runtime")

    ownership = load_json(AUDIT_OWNERSHIP_PATH)
    ownership_boundary = ownership.get("ownership_boundary") or {}
    require(
        ownership_boundary.get("boundary_status") == "defined_without_store_runtime",
        "ownership boundary status drifted",
    )
    require(
        ownership_boundary.get("store_owner_ref_status") == "defined_static_only",
        "store owner reference must stay static",
    )
    require(
        ownership_boundary.get("audit_store_runtime_status") == "not_created",
        "ownership boundary must not create runtime",
    )

    delivery = load_json(AUDIT_DELIVERY_IDEMPOTENCY_PATH)
    delivery_boundary = delivery.get("delivery_idempotency_boundary") or {}
    require(
        delivery_boundary.get("boundary_status") == "defined_without_delivery_runtime",
        "delivery idempotency boundary status drifted",
    )
    require(
        delivery_boundary.get("delivery_result_envelope_status") == "metadata_only_static_envelope_defined",
        "delivery result envelope status drifted",
    )
    require(delivery_boundary.get("delivery_runtime_status") == "not_created", "delivery runtime must remain absent")
    require(
        delivery_boundary.get("idempotency_runtime_status") == "not_created",
        "idempotency runtime must remain absent",
    )


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in (alignment.get("target_fields") or {}).items():
        require(target.get(field) == expected, f"implementation target {field} drifted")

    planned_slice = alignment.get("planned_slice") or {}
    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or [] if isinstance(item, dict)}
    slice_id = str(planned_slice.get("id") or "")
    require(slice_id in planned, "implementation readiness missing entry refresh v3 planned slice")
    require(planned[slice_id].get("status") == planned_slice.get("status"), "planned slice status drifted")
    evidence = set(planned[slice_id].get("evidence") or [])
    for path in planned_slice.get("evidence") or []:
        require(path in evidence, f"planned slice missing evidence: {path}")
        require((REPO_ROOT / path).exists(), f"planned slice evidence missing on disk: {path}")

    for consumer in alignment.get("required_consumers") or []:
        require((REPO_ROOT / consumer).exists(), f"missing required consumer: {consumer}")


def assert_docs_and_registration() -> None:
    for doc_path, needles in EXPECTED_DOC_REFERENCES.items():
        text = read(doc_path)
        for needle in needles:
            require(needle in text, f"{doc_path} missing reference: {needle}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    after = "check-production-ops-secret-backend-audit-store-delivery-idempotency-runtime-boundary-readiness-v1.py"
    script = "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py"
    before = "check-production-ops-startup-supervisor-boundary.py"
    require(after in check_repo, "check-repo missing delivery idempotency checker")
    require(script in check_repo, "check-repo missing audit store entry refresh v3 checker")
    require(before in check_repo, "check-repo missing startup supervisor checker")
    require(check_repo.index(after) < check_repo.index(script) < check_repo.index(before), "check-repo order drifted")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_refresh_boundary(fixture)
    assert_sets_and_matrices(fixture)
    assert_diagnostics_and_no_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_prior_evidence_alignment()
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    print("production ops secret backend audit store runtime implementation entry refresh v3 checks passed.")


if __name__ == "__main__":
    main()
