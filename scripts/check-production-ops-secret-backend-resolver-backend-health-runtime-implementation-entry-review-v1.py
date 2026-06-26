#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
HEALTH_BOUNDARY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-boundary-readiness-v1.json"
)
REAL_RESOLVER_ENTRY_REVIEW_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json"
)
SECRET_REFERENCE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-reference-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "production_ready",
    "production_secret_backend_ready",
    "cloud_secret_service_ready",
    "resolver_implemented",
    "resolver_runtime_ready",
    "production_resolver_runtime_created",
    "production_resolver_runtime_task_card_created",
    "backend_health_runtime_ready",
    "backend_health_runtime_created",
    "backend_health_runtime_task_card_created",
    "backend_health_check_executed",
    "backend_health_client_created",
    "backend_runtime_created",
    "real_secret_read",
    "real_secret_written",
    "credential_payload_created",
    "credential_handle_created",
    "credential_handle_runtime_created",
    "operator_approval_runtime_created",
    "operator_approval_runtime_executed",
    "audit_store_runtime_created",
    "audit_writer_created",
    "audit_event_written",
    "no_secret_leakage_smoke_runtime_created",
    "credential_resolved",
    "database_connection_provider_ready",
    "database_driver_selected",
    "connection_factory_implemented",
    "sql_migration_created",
    "schema_marker_ready",
    "migration_runner_implemented",
    "repository_mode_ready",
    "production_secret_audit_store_ready",
    "production_api_ready",
}

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-resolver-backend-health-boundary-readiness-v1": (
        "resolver_backend_health_boundary_readiness_defined"
    ),
    "production-secret-backend-real-resolver-runtime-implementation-entry-review-v1": (
        "real_resolver_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "resolver_backend_profile_selection_readiness_defined"
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1": (
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
    ),
    "production-secret-backend-credential-handle-runtime-boundary-readiness-v1": (
        "credential_handle_runtime_boundary_readiness_defined"
    ),
    "production-secret-backend-operator-approval-runtime-evidence-readiness-v1": (
        "operator_approval_runtime_evidence_readiness_defined"
    ),
    "production-secret-backend-audit-store-handoff-readiness-v1": "audit_store_handoff_readiness_defined",
    "production-secret-backend-operator-runbook-negative-gates-readiness-v1": (
        "operator_runbook_negative_gates_readiness_defined"
    ),
    "production-secret-backend-rotation-audit-policy-readiness-v1": "rotation_audit_policy_readiness_defined",
    "production-secret-backend-provider-profile-secret-binding-readiness-v1": (
        "provider_profile_secret_binding_readiness_defined"
    ),
    "production-secret-backend-secret-resolver-interface-disabled-readiness-v1": (
        "secret_resolver_interface_disabled_readiness_defined"
    ),
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
}

EXPECTED_ENTRY_FIELDS = {
    "entry_review_status": "blocked_before_runtime_task_card",
    "runtime_task_decision": "resolver_backend_health_runtime_implementation_blocked_before_task_card",
    "health_boundary_status": "resolver_backend_health_boundary_readiness_defined",
    "backend_health_runtime_implementation_status": "not_started",
    "backend_health_runtime_status": "not_created",
    "backend_health_check_status": "not_executed",
    "backend_health_client_status": "not_created",
    "backend_profile_binding_status": "defined_static_only",
    "environment_binding_status": "defined_no_cross_environment",
    "provider_profile_binding_status": "defined_static_only",
    "secret_ref_binding_status": "reference_only",
    "operator_approval_runtime_status": "not_created",
    "operator_approval_runtime_execution_status": "not_executed",
    "audit_store_status": "not_created",
    "audit_writer_status": "not_created",
    "audit_event_delivery_status": "not_executed",
    "credential_handle_runtime_status": "not_created",
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
    "backend_health_runtime_created_in_this_slice",
    "backend_health_client_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "provider_call_executed_in_this_slice",
    "cloud_secret_call_executed_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "credential_payload_created_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "operator_approval_runtime_created_in_this_slice",
    "operator_approval_runtime_executed_in_this_slice",
    "audit_store_created_in_this_slice",
    "audit_writer_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "no_secret_leakage_smoke_runtime_created_in_this_slice",
    "cloud_secret_service_enabled",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_BLOCKED = {
    "operator_approval_runtime_not_created",
    "audit_store_runtime_not_created",
    "credential_handle_runtime_not_created",
    "real_no_leakage_smoke_runtime_not_created",
    "backend_health_runtime_task_card_not_created",
    "production_resolver_runtime_not_created",
    "cloud_secret_service_not_selected",
    "repository_mode_disabled",
}

EXPECTED_GATE_STATUS = {
    "backend_health_boundary_consumed": "satisfied_static_boundary",
    "runtime_task_card_gate": "blocked_before_task_card",
    "backend_profile_binding": "satisfied_static_only",
    "environment_binding": "satisfied_static_no_cross_environment",
    "provider_profile_binding": "satisfied_static_only",
    "secret_ref_binding": "reference_only",
    "operator_approval_runtime": "blocked_runtime_not_created",
    "audit_store_runtime": "blocked_store_not_created",
    "credential_handle_runtime": "blocked_runtime_not_created",
    "no_secret_leakage_smoke_runtime": "blocked_runtime_not_created",
    "backend_health_runtime": "not_created",
    "backend_health_check_execution": "not_executed",
    "production_resolver_runtime": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_RUNTIME_REQUIREMENTS = {
    "disabled-by-default runtime gate",
    "metadata-only health result",
    "backend profile binding",
    "provider profile binding",
    "secret ref reference-only binding",
    "environment binding",
    "health policy binding",
    "operator approval runtime dependency gate",
    "audit store writer dependency gate",
    "credential handle runtime dependency gate",
    "no leakage smoke runtime dependency gate",
    "fail-closed lifecycle mapping",
    "sanitized diagnostics allowlist",
    "offline unit test and static smoke",
    "side effect counters",
}

EXPECTED_FAILURE_CODES = {
    "resolver_backend_health_runtime_entry_boundary_missing",
    "resolver_backend_health_runtime_entry_task_card_blocked",
    "resolver_backend_health_runtime_entry_operator_approval_runtime_missing",
    "resolver_backend_health_runtime_entry_audit_store_missing",
    "resolver_backend_health_runtime_entry_credential_handle_runtime_missing",
    "resolver_backend_health_runtime_entry_no_leakage_runtime_missing",
    "resolver_backend_health_runtime_entry_backend_profile_missing",
    "resolver_backend_health_runtime_entry_environment_mismatch",
    "resolver_backend_health_runtime_entry_secret_value_detected",
    "resolver_backend_health_runtime_created_in_entry_review",
    "resolver_backend_health_check_executed_in_entry_review",
    "resolver_backend_health_runtime_entry_fallback_forbidden",
    "resolver_backend_health_runtime_entry_repository_mode_forbidden",
    "resolver_backend_health_runtime_entry_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "health_runtime_entry_status",
    "health_boundary_status",
    "runtime_task_decision",
    "backend_health_runtime_status",
    "backend_health_check_status",
    "backend_profile_status",
    "environment_binding_status",
    "provider_profile_status",
    "secret_ref_binding_status",
    "operator_approval_runtime_status",
    "audit_store_status",
    "credential_handle_runtime_status",
    "no_secret_leakage_runtime_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_FORBIDDEN_DIAGNOSTICS = {
    "raw_secret",
    "secret_value",
    "password",
    "token",
    "api_key",
    "provider_raw_url",
    "resolver_backend_url",
    "backend_endpoint_url",
    "dsn",
    "cloud_credential",
    "database_hostname",
    "database_error_detail",
    "credential_payload",
    "full_secret_ref_value",
    "full_credential_handle",
    "full_operator_identity_claim",
    "full_user_claim",
    "authorization_header",
    "cookie",
    "raw_health_request",
    "raw_health_response",
    "raw_backend_error_detail",
}

EXPECTED_NO_SIDE_EFFECT_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "backend_health_runtime_created_count",
    "backend_health_client_created_count",
    "backend_health_check_count",
    "network_call_count",
    "cloud_secret_call_count",
    "provider_call_count",
    "credential_payload_created_count",
    "credential_handle_created_count",
    "credential_handle_runtime_created_count",
    "operator_approval_runtime_execution_count",
    "audit_store_created_count",
    "audit_writer_created_count",
    "audit_event_write_count",
    "database_connection_count",
    "driver_open_count",
    "sql_execution_count",
    "schema_marker_read_count",
    "schema_marker_write_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}

EXPECTED_STATIC_ARTIFACTS = {
    "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.md",
    "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1-plan.md",
    "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
    "scripts/check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py",
}

EXPECTED_DOC_REFERENCES = {
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
        "backend health runtime implementation entry review",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review v1",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review v1",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ],
    "docs/task-cards/README.md": [
        "Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py",
        "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
    ],
    "docs/devlogs/2026-W25.md": [
        "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ],
    "docs/devlogs/2026-W25.parts/2026-06-19-to-2026-06-21.md": [
        "Production Secret Backend Resolver Backend Health Runtime Implementation Entry Review",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a JSON object")
    return document


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_resolver_backend_health_runtime_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id")
        == "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1",
        "unexpected slice id",
    )
    require(
        slice_info.get("status") == "resolver_backend_health_runtime_implementation_entry_review_defined",
        "unexpected slice status",
    )
    for key, expected_path in {
        "task_card": (
            "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1-plan.md"
        ),
        "platform_topic": (
            "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.md"
        ),
    }.items():
        value = str(slice_info.get(key) or "")
        require(value == expected_path, f"unexpected {key}")
        require((REPO_ROOT / value).exists(), f"{key} path missing: {value}")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, expected_status in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        evidence = str(item.get("evidence") or "")
        require((REPO_ROOT / evidence).exists(), f"{dependency_id} evidence missing: {evidence}")

    reference = load_json(SECRET_REFERENCE_PATH)
    require(reference.get("scope") == "secret_reference_only", "secret reference must remain reference-only")
    policy = reference.get("policy") or {}
    for field in ("stores_secret_values", "resolver_enabled", "cloud_calls_allowed", "production_secret_backend_ready"):
        require(policy.get(field) is False, f"secret reference policy {field} must remain false")


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_boundary") or {}
    for field, expected in EXPECTED_ENTRY_FIELDS.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"{field} must remain false")

    health_boundary = load_json(HEALTH_BOUNDARY_PATH)
    health_state = health_boundary.get("health_boundary") or {}
    require(
        health_state.get("backend_health_runtime_status") == "not_created",
        "health boundary must not create backend health runtime",
    )
    require(
        health_state.get("backend_health_check_status") == "not_executed",
        "health boundary must not execute backend health check",
    )

    real_entry = load_json(REAL_RESOLVER_ENTRY_REVIEW_PATH)
    real_boundary = real_entry.get("entry_boundary") or {}
    require(
        real_boundary.get("resolver_runtime_status") == "not_created",
        "real resolver runtime must remain not_created",
    )


def assert_blocked_and_gates(fixture: dict[str, Any]) -> None:
    blocked = set(fixture.get("blocked_conditions") or [])
    missing_blocked = sorted(EXPECTED_BLOCKED - blocked)
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    rows = rows_by_id(fixture, "gate_matrix", "gate")
    missing = sorted(set(EXPECTED_GATE_STATUS) - set(rows))
    require(not missing, f"missing gate rows: {missing}")
    for gate, expected_status in EXPECTED_GATE_STATUS.items():
        require(rows[gate].get("status") == expected_status, f"{gate} status drifted")


def assert_future_task_requirements(fixture: dict[str, Any]) -> None:
    requirements = set(fixture.get("future_runtime_task_card_requirements") or [])
    missing = sorted(EXPECTED_RUNTIME_REQUIREMENTS - requirements)
    require(not missing, f"missing future runtime task requirements: {missing}")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture, "failure_mapping", "code")
    missing = sorted(EXPECTED_FAILURE_CODES - set(rows))
    require(not missing, f"missing failure codes: {missing}")
    for code in EXPECTED_FAILURE_CODES:
        item = rows[code]
        require(item.get("boundary"), f"{code} missing failure boundary")
        diagnostic = str(item.get("diagnostic") or "")
        require(diagnostic, f"{code} missing diagnostic")


def assert_diagnostics(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    missing_allowed = sorted(EXPECTED_DIAGNOSTIC_FIELDS - allowed)
    require(not missing_allowed, f"missing allowed diagnostic fields: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_DIAGNOSTICS - forbidden)
    require(not missing_forbidden, f"missing forbidden diagnostic fields: {missing_forbidden}")
    sample = diagnostics.get("sample") or {}
    require(set(sample).issubset(allowed), "diagnostic sample must only use allowed fields")
    require(sample.get("backend_health_runtime_status") == "not_created", "sample runtime status drifted")
    require(sample.get("backend_health_check_status") == "not_executed", "sample health check status drifted")
    require(
        sample.get("runtime_task_decision")
        == "resolver_backend_health_runtime_implementation_blocked_before_task_card",
        "sample runtime task decision drifted",
    )


def assert_no_fallback_and_side_effects(fixture: dict[str, Any]) -> None:
    fallback = fixture.get("no_fallback_policy") or {}
    require(fallback.get("fallback_policy") == "forbidden", "fallback policy must be forbidden")
    require(fallback.get("missing_dependency_result") == "fail_closed", "missing dependency must fail closed")
    require(
        fallback.get("missing_dependency_must_not_create_health_success") is True,
        "missing dependency must not create health success",
    )
    forbidden_fallbacks = set(fallback.get("forbidden_fallbacks") or [])
    for forbidden in {
        "fake_resolver_runtime",
        "developer_env_plaintext",
        "fixture_credential",
        "sample",
        "mock_provider",
        "local_smoke_profile",
        "operator_runbook_text",
        "repository_memory_store",
        "audit_memory_store",
        "cross_environment_health_boundary",
    }:
        require(forbidden in forbidden_fallbacks, f"missing forbidden fallback: {forbidden}")

    side_effects = fixture.get("no_side_effects") or {}
    require(side_effects.get("checker_mode") == "committed_static_artifacts_only", "checker mode drifted")
    for counter in EXPECTED_NO_SIDE_EFFECT_COUNTERS:
        require(side_effects.get(counter) == 0, f"{counter} must remain zero")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    allowed = set(guard.get("allowed_new_artifacts") or [])
    missing_allowed = sorted(EXPECTED_STATIC_ARTIFACTS - allowed)
    require(not missing_allowed, f"missing allowed artifacts: {missing_allowed}")
    for artifact in EXPECTED_STATIC_ARTIFACTS:
        require((REPO_ROOT / artifact).exists(), f"allowed static artifact missing: {artifact}")

    forbidden = set(guard.get("forbidden_artifacts") or [])
    for artifact in {
        "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-v1-plan.md",
        "services/platform/internal/secretbackend/health_runtime.go",
        "services/platform/internal/secretbackend/health_client.go",
        "services/platform/internal/secretbackend/health_probe.go",
    }:
        require(artifact in forbidden, f"missing forbidden artifact marker: {artifact}")
        require(not (REPO_ROOT / artifact).exists(), f"forbidden artifact exists: {artifact}")


def assert_alignments(fixture: dict[str, Any]) -> None:
    real_alignment = fixture.get("real_resolver_entry_review_alignment") or {}
    for field, expected in {
        "real_resolver_entry_review_status": "blocked_before_runtime_task_card",
        "backend_health_boundary_status": "readiness_defined_without_backend_health_runtime",
        "backend_health_runtime_entry_review_status": "blocked_before_runtime_task_card",
        "backend_health_runtime_status": "not_created",
        "backend_health_check_status": "not_executed",
        "real_resolver_runtime_task_card_status": "not_created",
    }.items():
        require(real_alignment.get(field) == expected, f"real resolver alignment {field} drifted")
    for artifact in real_alignment.get("required_static_artifacts") or []:
        require((REPO_ROOT / str(artifact)).exists(), f"real resolver alignment artifact missing: {artifact}")

    impl = load_json(IMPLEMENTATION_READINESS_PATH)
    target = impl.get("implementation_target") or {}
    alignment = fixture.get("implementation_readiness_alignment") or {}
    for field, expected in (alignment.get("target_fields") or {}).items():
        require(target.get(field) == expected, f"implementation readiness target {field} drifted")
    planned_slice = alignment.get("planned_slice") or {}
    planned = rows_by_id(impl, "planned_slices", "id")
    planned_id = str(planned_slice.get("id") or "")
    require(planned_id in planned, f"implementation readiness missing planned slice: {planned_id}")
    require(planned[planned_id].get("status") == planned_slice.get("status"), "planned slice status drifted")
    planned_evidence = set(planned[planned_id].get("evidence") or [])
    for artifact in planned_slice.get("evidence") or []:
        require(artifact in planned_evidence, f"planned slice missing evidence: {artifact}")
        require((REPO_ROOT / str(artifact)).exists(), f"planned slice evidence missing on disk: {artifact}")

    required_consumers = set(impl.get("required_consumers") or [])
    for consumer in alignment.get("required_consumers") or []:
        require(consumer in required_consumers, f"implementation readiness missing consumer: {consumer}")


def assert_doc_references(fixture: dict[str, Any]) -> None:
    for path in fixture.get("doc_references") or []:
        require((REPO_ROOT / str(path)).exists(), f"doc reference missing: {path}")
    for path, expected_strings in EXPECTED_DOC_REFERENCES.items():
        text = read(path)
        for expected in expected_strings:
            require(expected in text, f"{path} missing expected text: {expected}")


def assert_check_repo_registration(fixture: dict[str, Any]) -> None:
    registration = fixture.get("check_repo_registration") or {}
    script = str(registration.get("script") or "")
    after = str(registration.get("after") or "")
    before = str(registration.get("before") or "")
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    for marker in (script, after, before):
        require(marker in check_repo, f"check-repo missing marker: {marker}")
    require(check_repo.index(after) < check_repo.index(script) < check_repo.index(before), "check-repo order drifted")


def assert_no_secret_like_values(fixture: dict[str, Any]) -> None:
    serialized = json.dumps(fixture, sort_keys=True)
    for forbidden_literal in (
        "sk-live-",
        "sk-test-",
        "AKIA",
        "BEGIN PRIVATE KEY",
        "xoxb-",
        "postgres://",
        "mysql://",
        "mongodb://",
    ):
        require(forbidden_literal not in serialized, f"secret-like literal found: {forbidden_literal}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_boundary(fixture)
    assert_blocked_and_gates(fixture)
    assert_future_task_requirements(fixture)
    assert_failure_mapping(fixture)
    assert_diagnostics(fixture)
    assert_no_fallback_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_alignments(fixture)
    assert_doc_references(fixture)
    assert_check_repo_registration(fixture)
    assert_no_secret_like_values(fixture)
    print("production ops secret backend resolver backend health runtime implementation entry review checks passed.")


if __name__ == "__main__":
    main()
