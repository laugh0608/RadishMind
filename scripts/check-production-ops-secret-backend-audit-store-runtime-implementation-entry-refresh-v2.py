#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
AUDIT_ENTRY_REVIEW_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.json"
)
AUDIT_CONTRACT_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-handoff-readiness-v1": "audit_store_handoff_readiness_defined",
    "production-secret-backend-audit-store-runtime-implementation-entry-review-v1": (
        "audit_store_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-audit-store-contract-event-schema-readiness-v1": (
        "audit_store_contract_event_schema_readiness_defined"
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
    "production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1": (
        "real_resolver_runtime_implementation_entry_refresh_defined"
    ),
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
}

EXPECTED_BOUNDARY_FIELDS = {
    "entry_refresh_status": "blocked_before_runtime_task_card",
    "runtime_task_decision": "audit_store_runtime_implementation_still_blocked_before_task_card",
    "previous_entry_review_status": "blocked_before_runtime_task_card",
    "audit_handoff_status": "defined_without_store_runtime",
    "audit_store_contract_status": "static_contract_defined",
    "contract_event_schema_readiness_status": "defined_without_store_runtime",
    "resolved_static_prerequisites_status": "contract_layer_satisfied",
    "unresolved_runtime_dependency_status": "runtime_dependencies_still_blocked",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_implementation_status": "not_started",
    "audit_store_runtime_status": "not_created",
    "audit_store_status": "not_created",
    "audit_writer_status": "not_created",
    "runtime_event_schema_status": "not_created",
    "audit_event_delivery_status": "not_executed",
    "store_ownership_status": "required_before_runtime_task_card",
    "writer_ownership_status": "required_before_runtime_task_card",
    "idempotency_key_ref_status": "static_contract_defined",
    "delivery_result_envelope_status": "static_contract_defined",
    "delivery_runtime_status": "not_created",
    "retention_policy_status": "static_contract_defined",
    "redaction_profile_status": "static_contract_defined",
    "event_version_status": "static_contract_defined",
    "event_kind_allowlist_status": "static_contract_defined",
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
    "audit_writer_created_in_this_slice",
    "runtime_event_schema_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "audit_delivery_executed_in_this_slice",
    "retention_runtime_executed_in_this_slice",
    "redaction_runtime_executed_in_this_slice",
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
    "metadata-only audit event schema",
    "event version reference",
    "event kind allowlist",
    "required optional field allowlist",
    "forbidden field denylist",
    "reference-only writer input",
    "delivery result envelope",
    "idempotency key reference",
    "retention policy reference",
    "redaction profile reference",
    "sanitized diagnostic allowlist",
    "failure mapping",
    "no fallback policy",
    "no side effects policy",
    "artifact guard",
}

EXPECTED_BLOCKED = {
    "audit_store_runtime_task_card_not_created",
    "audit_store_ownership_not_defined_for_runtime",
    "audit_writer_runtime_not_created",
    "runtime_event_schema_not_created",
    "audit_event_delivery_runtime_not_created",
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
    "runtime_task_card_gate": "still_blocked_before_task_card",
    "store_ownership": "required_before_runtime_task_card",
    "writer_ownership": "required_before_runtime_task_card",
    "runtime_event_schema": "not_created",
    "audit_event_delivery": "not_executed",
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
    "store ownership boundary",
    "writer ownership boundary",
    "metadata-only runtime event schema materialization",
    "event versioning",
    "event kind allowlist",
    "required optional field allowlist",
    "forbidden field denylist",
    "reference-only writer input",
    "idempotency key reference",
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
    "audit_store_runtime_refresh_contract_missing",
    "audit_store_runtime_refresh_task_card_still_blocked",
    "audit_store_runtime_refresh_store_ownership_missing",
    "audit_store_runtime_refresh_writer_runtime_missing",
    "audit_store_runtime_refresh_runtime_schema_missing",
    "audit_store_runtime_refresh_delivery_runtime_missing",
    "audit_store_runtime_refresh_operator_approval_runtime_missing",
    "audit_store_runtime_refresh_credential_handle_runtime_missing",
    "audit_store_runtime_refresh_backend_health_runtime_missing",
    "audit_store_runtime_refresh_no_leakage_runtime_missing",
    "audit_store_runtime_refresh_secret_material_detected",
    "audit_store_runtime_created_in_refresh",
    "audit_writer_created_in_refresh",
    "audit_event_written_in_refresh",
    "audit_store_runtime_refresh_fallback_forbidden",
    "audit_store_runtime_refresh_repository_mode_forbidden",
    "audit_store_runtime_refresh_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "audit_store_runtime_entry_refresh_status",
    "runtime_task_decision",
    "audit_handoff_status",
    "audit_store_contract_status",
    "event_schema_status",
    "writer_contract_status",
    "audit_store_runtime_task_card_status",
    "audit_store_runtime_status",
    "audit_writer_status",
    "runtime_event_schema_status",
    "audit_event_delivery_status",
    "store_ownership_status",
    "writer_ownership_status",
    "idempotency_key_ref_status",
    "retention_policy_status",
    "redaction_profile_status",
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
    "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.md",
    "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.json",
    "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.py",
}

EXPECTED_DOC_REFERENCES = {
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2",
        "audit_store_runtime_implementation_entry_refresh_v2_defined",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh v2",
        "audit_store_runtime_implementation_entry_refresh_v2_defined",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh v2",
        "audit_store_runtime_implementation_entry_refresh_v2_defined",
    ],
    "docs/task-cards/README.md": [
        "Production Secret Backend Audit Store Runtime Implementation Entry Refresh",
        "audit_store_runtime_implementation_entry_refresh_v2_defined",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2",
        "audit_store_runtime_implementation_entry_refresh_v2_defined",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.py",
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.json",
    ],
    "docs/devlogs/2026-W25.md": [
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2",
        "audit_store_runtime_implementation_entry_refresh_v2_defined",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def assert_slice(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_runtime_implementation_entry_refresh_v2_defined",
        "unexpected slice status",
    )
    for path in [slice_info.get("task_card"), slice_info.get("platform_topic")]:
        require(isinstance(path, str) and (REPO_ROOT / path).exists(), f"missing slice artifact: {path}")
    for forbidden_claim in {
        "audit_store_runtime_ready",
        "audit_store_runtime_created",
        "audit_store_runtime_task_card_created",
        "audit_writer_created",
        "audit_event_written",
        "production_resolver_runtime_created",
        "cloud_secret_service_ready",
        "production_api_ready",
    }:
        require(forbidden_claim in set(slice_info.get("does_not_claim") or []), f"missing forbidden claim {forbidden_claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = {
        str(item.get("id")): item
        for item in fixture.get("depends_on") or []
        if isinstance(item, dict)
    }
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, expected_status in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        evidence = str(item.get("evidence") or "")
        require(evidence, f"{dependency_id} missing evidence")
        require((REPO_ROOT / evidence).exists(), f"{dependency_id} evidence does not exist: {evidence}")


def assert_entry_refresh_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_refresh_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY_FIELDS.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"{field} must be false")


def assert_sets_and_matrices(fixture: dict[str, Any]) -> None:
    resolved = set(fixture.get("resolved_static_prerequisites") or [])
    require(EXPECTED_RESOLVED <= resolved, f"missing resolved static prerequisites: {sorted(EXPECTED_RESOLVED - resolved)}")
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
    require(not missing_requirements, f"missing future runtime task card requirements: {missing_requirements}")

    failure_codes = {
        str(item.get("code"))
        for item in fixture.get("failure_mapping") or []
        if isinstance(item, dict)
    }
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
    require(sample.get("failure_code") == "audit_store_runtime_refresh_task_card_still_blocked", "bad sample failure")

    no_fallback = fixture.get("no_fallback_policy") or {}
    require(no_fallback.get("fallback_policy") == "forbidden", "fallback must be forbidden")
    require(no_fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for fallback in {"fake_resolver_runtime", "static_contract", "audit_memory_store", "repository_memory_store"}:
        require(fallback in set(no_fallback.get("forbidden_fallbacks") or []), f"missing forbidden fallback {fallback}")

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
    }:
        require(not (REPO_ROOT / forbidden_path).exists(), f"forbidden artifact exists: {forbidden_path}")


def assert_prior_evidence_alignment() -> None:
    entry_review = load_json(AUDIT_ENTRY_REVIEW_PATH)
    entry_boundary = entry_review.get("entry_boundary") or {}
    require(
        entry_boundary.get("runtime_task_decision") == "audit_store_runtime_implementation_blocked_before_task_card",
        "audit store v1 entry decision drifted",
    )
    require(entry_boundary.get("audit_store_runtime_status") == "not_created", "audit store runtime must remain absent")

    contract = load_json(AUDIT_CONTRACT_PATH)
    contract_boundary = contract.get("contract_boundary") or {}
    require(
        contract_boundary.get("readiness_status") == "defined_without_store_runtime",
        "audit store contract readiness drifted",
    )
    require(
        contract_boundary.get("audit_event_schema_status") == "static_contract_defined",
        "audit event static schema readiness drifted",
    )
    require(contract_boundary.get("audit_store_runtime_status") == "not_created", "contract must not create runtime")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in (alignment.get("target_fields") or {}).items():
        require(target.get(field) == expected, f"implementation target {field} drifted")

    planned_slice = alignment.get("planned_slice") or {}
    planned = {
        str(item.get("id")): item
        for item in readiness.get("planned_slices") or []
        if isinstance(item, dict)
    }
    slice_id = str(planned_slice.get("id") or "")
    require(slice_id in planned, "implementation readiness missing entry refresh v2 planned slice")
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
    after = 'check-production-ops-secret-backend-audit-store-contract-event-schema-readiness-v1.py'
    script = 'check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.py'
    before = 'check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py'
    require(after in check_repo, "check-repo missing previous audit contract checker")
    require(script in check_repo, "check-repo missing audit store entry refresh v2 checker")
    require(before in check_repo, "check-repo missing operator approval entry checker")
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
    print("production ops secret backend audit store runtime implementation entry refresh v2 checks passed.")


if __name__ == "__main__":
    main()
