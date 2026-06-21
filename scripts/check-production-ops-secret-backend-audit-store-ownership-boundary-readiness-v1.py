#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-ownership-boundary-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
AUDIT_ENTRY_REFRESH_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v2.json"
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
    "boundary_status": "defined_without_store_runtime",
    "ownership_readiness_status": "audit_store_ownership_boundary_readiness_defined",
    "entry_refresh_v2_status": "blocked_before_runtime_task_card",
    "audit_handoff_status": "defined_without_store_runtime",
    "audit_store_contract_status": "static_contract_defined",
    "store_owner_model_status": "static_reference_defined",
    "store_owner_ref_status": "defined_static_only",
    "durable_backend_owner_status": "required_before_runtime_task_card",
    "writer_owner_model_status": "separated_from_store_owner",
    "writer_owner_ref_status": "defined_static_only",
    "writer_runtime_owner_status": "not_created",
    "runtime_event_schema_owner_status": "required_before_runtime_task_card",
    "delivery_owner_status": "required_next_boundary",
    "idempotency_owner_status": "required_next_boundary",
    "retention_policy_owner_status": "policy_reference_only",
    "redaction_profile_owner_status": "policy_reference_only",
    "operator_approval_owner_status": "reference_only_blocked_runtime",
    "credential_handle_owner_status": "reference_only_blocked_runtime",
    "backend_health_owner_status": "reference_only_blocked_runtime",
    "no_leakage_owner_status": "reference_only_blocked_runtime",
    "audit_store_runtime_task_card_status": "not_created",
    "audit_store_runtime_status": "not_created",
    "audit_store_status": "not_created",
    "audit_writer_status": "not_created",
    "runtime_event_schema_status": "not_created",
    "audit_event_delivery_status": "not_executed",
    "delivery_runtime_status": "not_created",
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

EXPECTED_OWNER_REFS = {
    "platform_secret_backend_audit_store_boundary",
    "platform_secret_backend_audit_writer_boundary",
    "platform_secret_backend_audit_schema_boundary",
    "platform_secret_backend_audit_delivery_boundary",
    "production_secret_policy_boundary",
    "production_secret_dependency_boundary",
}

EXPECTED_ALLOWED_METADATA = {
    "ownership_record_id",
    "ownership_version",
    "store_owner_ref",
    "durable_backend_owner_ref",
    "writer_owner_ref",
    "runtime_event_schema_owner_ref",
    "delivery_owner_ref",
    "idempotency_owner_ref",
    "retention_policy_owner_ref",
    "redaction_profile_owner_ref",
    "operator_approval_owner_ref",
    "credential_handle_owner_ref",
    "backend_health_owner_ref",
    "no_leakage_owner_ref",
    "environment",
    "provider",
    "provider_profile",
    "backend_profile_ref",
    "policy_version",
    "failure_code",
    "sanitized_diagnostic",
}

EXPECTED_FORBIDDEN_MATERIAL = {
    "credential_payload",
    "secret_value",
    "password",
    "token",
    "api_key",
    "cloud_credential",
    "provider_raw_url",
    "resolver_backend_url",
    "dsn",
    "database_hostname",
    "full_secret_ref_value",
    "full_credential_handle",
    "full_operator_identity_claim",
    "full_user_claim",
    "authorization_header",
    "cookie",
    "raw_request_payload",
    "raw_response_payload",
    "raw_audit_payload",
}

EXPECTED_GATE_STATUS = {
    "ownership_boundary": "defined_static_only",
    "store_owner_reference": "defined_static_only",
    "durable_backend_owner": "required_before_runtime_task_card",
    "writer_owner_separation": "required_before_runtime_task_card",
    "runtime_event_schema_owner": "required_before_runtime_task_card",
    "delivery_idempotency_owner": "required_next_boundary",
    "retention_redaction_owner": "policy_reference_only",
    "dependency_owners": "reference_only_blocked_runtime",
    "runtime_task_card_gate": "still_blocked_before_task_card",
    "audit_store_runtime": "not_created",
    "audit_writer": "not_created",
    "audit_event_delivery": "not_executed",
    "production_resolver_runtime": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_ownership_boundary_missing",
    "audit_store_ownership_owner_ref_missing",
    "audit_store_ownership_store_writer_coupled",
    "audit_store_ownership_durable_backend_owner_missing",
    "audit_store_ownership_runtime_schema_owner_missing",
    "audit_store_ownership_delivery_owner_missing",
    "audit_store_ownership_retention_owner_missing",
    "audit_store_ownership_redaction_owner_missing",
    "audit_store_ownership_operator_dependency_missing",
    "audit_store_ownership_credential_dependency_missing",
    "audit_store_ownership_backend_health_dependency_missing",
    "audit_store_ownership_no_leakage_dependency_missing",
    "audit_store_ownership_secret_material_detected",
    "audit_store_ownership_runtime_task_card_created",
    "audit_store_ownership_runtime_created_forbidden",
    "audit_store_ownership_writer_created_forbidden",
    "audit_store_ownership_event_written_forbidden",
    "audit_store_ownership_database_connection_forbidden",
    "audit_store_ownership_fallback_forbidden",
    "audit_store_ownership_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "audit_store_ownership_boundary_status",
    "store_owner_ref_status",
    "durable_backend_owner_status",
    "writer_owner_ref_status",
    "runtime_event_schema_owner_status",
    "delivery_owner_status",
    "idempotency_owner_status",
    "retention_policy_owner_status",
    "redaction_profile_owner_status",
    "operator_approval_owner_status",
    "credential_handle_owner_status",
    "backend_health_owner_status",
    "no_leakage_owner_status",
    "audit_store_runtime_task_card_status",
    "audit_store_runtime_status",
    "audit_writer_status",
    "audit_event_delivery_status",
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
    "docs/platform/production-secret-backend-audit-store-ownership-boundary-readiness-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-ownership-boundary-readiness-v1-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-ownership-boundary-readiness-v1.json",
    "scripts/check-production-ops-secret-backend-audit-store-ownership-boundary-readiness-v1.py",
}

EXPECTED_DOC_REFERENCES = {
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-audit-store-ownership-boundary-readiness-v1",
        "audit_store_ownership_boundary_readiness_defined",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Audit Store Ownership Boundary Readiness v1",
        "audit_store_ownership_boundary_readiness_defined",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Audit Store Ownership Boundary Readiness v1",
        "audit_store_ownership_boundary_readiness_defined",
    ],
    "docs/task-cards/README.md": [
        "Production Secret Backend Audit Store Ownership Boundary Readiness",
        "audit_store_ownership_boundary_readiness_defined",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "production-secret-backend-audit-store-ownership-boundary-readiness-v1",
        "audit_store_ownership_boundary_readiness_defined",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-audit-store-ownership-boundary-readiness-v1.py",
        "production-secret-backend-audit-store-ownership-boundary-readiness-v1.json",
    ],
    "docs/devlogs/2026-W25.md": [
        "production-secret-backend-audit-store-ownership-boundary-readiness-v1",
        "audit_store_ownership_boundary_readiness_defined",
    ],
    "docs/devlogs/2026-W25.parts/2026-06-19-to-2026-06-21.md": [
        "production-secret-backend-audit-store-ownership-boundary-readiness-v1",
        "audit_store_ownership_boundary_readiness_defined",
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
        fixture.get("kind") == "production_ops_secret_backend_audit_store_ownership_boundary_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-ownership-boundary-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_ownership_boundary_readiness_defined",
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


def assert_ownership_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("ownership_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY_FIELDS.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"{field} must be false")


def assert_ownership_model(fixture: dict[str, Any]) -> None:
    owner_rows = fixture.get("ownership_model") or []
    owner_refs = {str(item.get("owner_ref")) for item in owner_rows if isinstance(item, dict)}
    require(EXPECTED_OWNER_REFS <= owner_refs, f"missing owner refs: {sorted(EXPECTED_OWNER_REFS - owner_refs)}")
    for row in owner_rows:
        if not isinstance(row, dict):
            continue
        owner_ref = str(row.get("owner_ref"))
        owns = set(row.get("owns") or [])
        require(owns, f"{owner_ref} must list owned boundary items")
        runtime_status = row.get("runtime_status")
        require(runtime_status in {"not_created", "reference_only"}, f"{owner_ref} has runtime status drift")

    allowed = set(fixture.get("allowed_ownership_metadata") or [])
    require(EXPECTED_ALLOWED_METADATA <= allowed, "ownership metadata allowlist drifted")
    forbidden = set(fixture.get("forbidden_ownership_material") or [])
    require(EXPECTED_FORBIDDEN_MATERIAL <= forbidden, "ownership forbidden material drifted")


def assert_gates_and_failures(fixture: dict[str, Any]) -> None:
    gate_status = {
        str(item.get("gate")): str(item.get("status"))
        for item in fixture.get("gate_matrix") or []
        if isinstance(item, dict)
    }
    for gate, expected_status in EXPECTED_GATE_STATUS.items():
        require(gate_status.get(gate) == expected_status, f"{gate} status drifted")

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
        sample.get("audit_store_ownership_boundary_status") == "defined_without_store_runtime",
        "diagnostic ownership status drifted",
    )
    require(sample.get("failure_code") == "audit_store_ownership_delivery_owner_missing", "bad sample failure")

    no_fallback = fixture.get("no_fallback_policy") or {}
    require(no_fallback.get("fallback_policy") == "forbidden", "fallback must be forbidden")
    require(no_fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    for fallback in {"fake_resolver_runtime", "static_contract", "audit_memory_store", "repository_memory_store"}:
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
    }:
        require(not (REPO_ROOT / forbidden_path).exists(), f"forbidden artifact exists: {forbidden_path}")


def assert_prior_evidence_alignment() -> None:
    entry_refresh = load_json(AUDIT_ENTRY_REFRESH_PATH)
    boundary = entry_refresh.get("entry_refresh_boundary") or {}
    require(
        boundary.get("runtime_task_decision") == "audit_store_runtime_implementation_still_blocked_before_task_card",
        "audit store entry refresh v2 decision drifted",
    )
    require(
        boundary.get("store_ownership_status") == "required_before_runtime_task_card",
        "entry refresh v2 store ownership status drifted",
    )
    require(
        boundary.get("writer_ownership_status") == "required_before_runtime_task_card",
        "entry refresh v2 writer ownership status drifted",
    )
    require(boundary.get("audit_store_runtime_status") == "not_created", "audit store runtime must remain absent")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in (alignment.get("target_fields") or {}).items():
        require(target.get(field) == expected, f"implementation target {field} drifted")

    planned_slice = alignment.get("planned_slice") or {}
    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or [] if isinstance(item, dict)}
    slice_id = str(planned_slice.get("id") or "")
    require(slice_id in planned, "implementation readiness missing ownership planned slice")
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
    after = 'check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py'
    script = 'check-production-ops-secret-backend-audit-store-ownership-boundary-readiness-v1.py'
    before = 'check-production-ops-startup-supervisor-boundary.py'
    require(after in check_repo, "check-repo missing real resolver entry refresh checker")
    require(script in check_repo, "check-repo missing audit store ownership checker")
    require(before in check_repo, "check-repo missing startup supervisor checker")
    require(check_repo.index(after) < check_repo.index(script) < check_repo.index(before), "check-repo order drifted")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_ownership_boundary(fixture)
    assert_ownership_model(fixture)
    assert_gates_and_failures(fixture)
    assert_diagnostics_and_no_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_prior_evidence_alignment()
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    print("production ops secret backend audit store ownership boundary readiness v1 checks passed.")


if __name__ == "__main__":
    main()
