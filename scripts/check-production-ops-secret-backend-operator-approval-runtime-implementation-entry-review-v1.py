#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
OPERATOR_EVIDENCE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json"
)
AUDIT_RUNTIME_ENTRY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.json"
)
HEALTH_RUNTIME_ENTRY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json"
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
    "operator_approval_runtime_ready",
    "operator_approval_runtime_created",
    "operator_approval_runtime_task_card_created",
    "operator_approval_runtime_executed",
    "operator_approval_executor_created",
    "operator_identity_provider_connected",
    "ticket_verifier_created",
    "change_window_verifier_created",
    "approval_policy_evaluator_created",
    "real_secret_read",
    "real_secret_written",
    "credential_payload_created",
    "credential_handle_created",
    "credential_handle_runtime_created",
    "audit_store_runtime_created",
    "audit_writer_created",
    "audit_event_written",
    "backend_health_runtime_created",
    "backend_health_check_executed",
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
    "production-secret-backend-operator-approval-runtime-evidence-readiness-v1": (
        "operator_approval_runtime_evidence_readiness_defined"
    ),
    "production-secret-backend-real-resolver-runtime-implementation-entry-review-v1": (
        "real_resolver_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-credential-handle-runtime-boundary-readiness-v1": (
        "credential_handle_runtime_boundary_readiness_defined"
    ),
    "production-secret-backend-audit-store-handoff-readiness-v1": "audit_store_handoff_readiness_defined",
    "production-secret-backend-audit-store-runtime-implementation-entry-review-v1": (
        "audit_store_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1": (
        "resolver_backend_health_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1": (
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
    ),
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "resolver_backend_profile_selection_readiness_defined"
    ),
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
    "runtime_task_decision": "operator_approval_runtime_implementation_blocked_before_task_card",
    "approval_evidence_boundary_status": "operator_approval_runtime_evidence_readiness_defined",
    "operator_approval_runtime_implementation_status": "not_started",
    "operator_approval_runtime_status": "not_created",
    "operator_approval_runtime_execution_status": "not_executed",
    "approval_executor_status": "not_created",
    "operator_identity_provider_status": "not_connected",
    "dual_control_verifier_status": "required_before_runtime",
    "ticket_verifier_status": "required_before_runtime",
    "change_window_verifier_status": "required_before_runtime",
    "policy_evaluator_status": "required_before_runtime",
    "approval_subject_binding_status": "defined_static_only",
    "backend_profile_binding_status": "defined_static_only",
    "environment_binding_status": "defined_no_cross_environment",
    "provider_profile_binding_status": "defined_static_only",
    "secret_ref_binding_status": "reference_only",
    "audit_store_runtime_status": "not_created",
    "audit_store_status": "not_created",
    "audit_writer_status": "not_created",
    "audit_event_delivery_status": "not_executed",
    "credential_handle_runtime_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "backend_health_check_status": "not_executed",
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
    "operator_approval_runtime_created_in_this_slice",
    "operator_approval_runtime_executed_in_this_slice",
    "approval_executor_created_in_this_slice",
    "operator_identity_provider_connected_in_this_slice",
    "ticket_verifier_created_in_this_slice",
    "change_window_verifier_created_in_this_slice",
    "policy_evaluator_created_in_this_slice",
    "identity_provider_call_executed_in_this_slice",
    "ticket_verifier_executed_in_this_slice",
    "change_window_verifier_executed_in_this_slice",
    "policy_evaluator_executed_in_this_slice",
    "provider_call_executed_in_this_slice",
    "cloud_secret_call_executed_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "credential_payload_created_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "backend_health_runtime_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "no_secret_leakage_smoke_runtime_created_in_this_slice",
    "cloud_secret_service_enabled",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_BLOCKED = {
    "operator_approval_runtime_task_card_not_created",
    "operator_approval_runtime_not_created",
    "approval_executor_not_created",
    "operator_identity_provider_not_connected",
    "dual_control_verifier_not_created",
    "ticket_change_window_verifier_not_created",
    "approval_policy_evaluator_not_created",
    "audit_store_runtime_not_created",
    "credential_handle_runtime_not_created",
    "backend_health_runtime_not_created",
    "real_no_leakage_smoke_runtime_not_created",
    "production_resolver_runtime_not_created",
    "cloud_secret_service_not_selected",
    "repository_mode_disabled",
}

EXPECTED_GATE_STATUS = {
    "approval_evidence_boundary_consumed": "satisfied_static_boundary",
    "runtime_task_card_gate": "blocked_before_task_card",
    "approval_executor": "blocked_executor_not_created",
    "operator_identity_provider": "required_before_runtime",
    "dual_control_verifier": "required_before_runtime",
    "ticket_change_window_verifier": "required_before_runtime",
    "policy_evaluator": "required_before_runtime",
    "approval_subject_binding": "satisfied_static_only",
    "backend_profile_binding": "satisfied_static_only",
    "environment_binding": "satisfied_static_no_cross_environment",
    "provider_profile_binding": "satisfied_static_only",
    "secret_ref_binding": "reference_only",
    "audit_store_runtime": "blocked_runtime_not_created",
    "credential_handle_runtime": "blocked_runtime_not_created",
    "backend_health_runtime": "blocked_runtime_not_created",
    "no_secret_leakage_smoke_runtime": "blocked_runtime_not_created",
    "operator_approval_runtime": "not_created",
    "operator_approval_runtime_execution": "not_executed",
    "production_resolver_runtime": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_RUNTIME_REQUIREMENTS = {
    "disabled-by-default runtime gate",
    "metadata-only approval result",
    "approval subject binding",
    "secret ref reference-only binding",
    "provider profile binding",
    "environment binding",
    "backend profile binding",
    "operator identity reference resolver",
    "full claim redaction",
    "dual control verifier",
    "ticket and change window verifier",
    "approval lifecycle semantics",
    "approval expiry semantics",
    "approval revocation semantics",
    "approval revalidation semantics",
    "audit handoff dependency gate",
    "audit store runtime dependency gate",
    "credential handle runtime dependency gate",
    "backend health runtime dependency gate",
    "no leakage smoke runtime dependency gate",
    "rotation policy binding",
    "runbook binding",
    "policy version binding",
    "sanitized diagnostics allowlist",
    "offline unit test and static smoke",
    "side effect counters",
}

EXPECTED_FAILURE_CODES = {
    "operator_approval_runtime_entry_boundary_missing",
    "operator_approval_runtime_entry_task_card_blocked",
    "operator_approval_runtime_entry_executor_missing",
    "operator_approval_runtime_entry_identity_provider_missing",
    "operator_approval_runtime_entry_dual_control_missing",
    "operator_approval_runtime_entry_ticket_window_missing",
    "operator_approval_runtime_entry_policy_evaluator_missing",
    "operator_approval_runtime_entry_audit_store_missing",
    "operator_approval_runtime_entry_credential_handle_runtime_missing",
    "operator_approval_runtime_entry_backend_health_runtime_missing",
    "operator_approval_runtime_entry_no_leakage_runtime_missing",
    "operator_approval_runtime_entry_secret_material_detected",
    "operator_approval_runtime_created_in_entry_review",
    "operator_approval_executor_created_in_entry_review",
    "operator_approval_runtime_executed_in_entry_review",
    "operator_approval_runtime_entry_fallback_forbidden",
    "operator_approval_runtime_entry_repository_mode_forbidden",
    "operator_approval_runtime_entry_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "operator_approval_runtime_entry_status",
    "approval_evidence_boundary_status",
    "runtime_task_decision",
    "operator_approval_runtime_status",
    "operator_approval_runtime_execution_status",
    "approval_executor_status",
    "operator_identity_provider_status",
    "dual_control_status",
    "approval_ticket_status",
    "approval_window_status",
    "policy_evaluator_status",
    "audit_store_runtime_status",
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

EXPECTED_FORBIDDEN_DIAGNOSTICS = {
    "raw_secret",
    "secret_value",
    "password",
    "token",
    "api_key",
    "provider_raw_url",
    "resolver_backend_url",
    "dsn",
    "cloud_credential",
    "database_hostname",
    "database_error_detail",
    "credential_payload",
    "full_secret_ref_value",
    "full_credential_handle",
    "full_operator_claim",
    "full_user_claim",
    "authorization_header",
    "cookie",
    "raw_ticket_payload",
    "raw_identity_response",
    "raw_approval_payload",
    "raw_request_payload",
    "raw_response_payload",
}

EXPECTED_NO_FALLBACK = {
    "no fallback to fake resolver runtime",
    "no fallback to developer env plaintext",
    "no fallback to fixture credential",
    "no fallback to committed secret value",
    "no fallback to sample",
    "no fallback to mock provider",
    "no fallback to local-smoke profile",
    "no fallback to previously approved test evidence",
    "no fallback to operator runbook text",
    "no fallback to repository memory store",
    "no fallback to audit memory store",
    "no fallback to static approval evidence boundary",
    "no fallback from production approval runtime to test approval evidence",
    "no fallback from test approval runtime to production approval evidence",
    "missing identity provider keeps runtime task card blocked",
    "missing dual control keeps runtime task card blocked",
    "missing ticket or change window keeps runtime task card blocked",
    "missing audit store runtime keeps runtime task card blocked",
    "approval runtime entry review does not mean approval runtime ready",
    "approval runtime entry review does not mean approval executed",
    "approval runtime entry review does not mean credential resolved",
}

EXPECTED_ARTIFACTS = {
    "docs/platform/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.md",
    "docs/task-cards/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1-plan.md",
    "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.json",
    "scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py",
}

EXPECTED_DOC_REFERENCES = {
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1",
        "operator_approval_runtime_implementation_entry_review_defined",
        "operator approval runtime implementation entry review",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Operator Approval Runtime Implementation Entry Review v1",
        "operator_approval_runtime_implementation_entry_review_defined",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Operator Approval Runtime Implementation Entry Review v1",
        "operator_approval_runtime_implementation_entry_review_defined",
    ],
    "docs/task-cards/README.md": [
        "Production Secret Backend Operator Approval Runtime Implementation Entry Review",
        "operator_approval_runtime_implementation_entry_review_defined",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "operator-approval-runtime-implementation-entry-review",
        "operator_approval_runtime_implementation_entry_review_defined",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py",
        "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.json",
    ],
    "docs/devlogs/2026-W25.md": [
        "Production Secret Backend Operator Approval Runtime Implementation Entry Review",
        "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1",
        "operator_approval_runtime_implementation_entry_review_defined",
    ],
    "docs/devlogs/2026-W25.parts/2026-06-19-to-2026-06-21.md": [
        "Production Secret Backend Operator Approval Runtime Implementation Entry Review",
        "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1",
        "operator_approval_runtime_implementation_entry_review_defined",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def assert_slice(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(
        fixture.get("kind") == "production_ops_secret_backend_operator_approval_runtime_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    require(
        slice_info.get("id") == "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "operator_approval_runtime_implementation_entry_review_defined",
        "unexpected slice status",
    )
    for key in ("task_card", "platform_topic"):
        path = str(slice_info.get(key) or "")
        require(path in EXPECTED_ARTIFACTS, f"unexpected {key}: {path}")
        require((REPO_ROOT / path).exists(), f"{key} does not exist: {path}")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


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
        require(evidence, f"{dependency_id} missing evidence path")
        require((REPO_ROOT / evidence).exists(), f"{dependency_id} evidence missing on disk: {evidence}")


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_boundary") or {}
    for key, expected in EXPECTED_ENTRY_FIELDS.items():
        require(boundary.get(key) == expected, f"{key} drifted: {boundary.get(key)!r}")
    for key in EXPECTED_FALSE_FLAGS:
        require(boundary.get(key) is False, f"{key} must be false")


def assert_lists(fixture: dict[str, Any]) -> None:
    blocked = set(fixture.get("blocked_conditions") or [])
    require(EXPECTED_BLOCKED <= blocked, f"missing blocked conditions: {sorted(EXPECTED_BLOCKED - blocked)}")

    gate_matrix = {
        str(item.get("gate")): str(item.get("status"))
        for item in fixture.get("gate_matrix") or []
        if isinstance(item, dict)
    }
    for gate, expected in EXPECTED_GATE_STATUS.items():
        require(gate_matrix.get(gate) == expected, f"{gate} gate drifted: {gate_matrix.get(gate)!r}")

    runtime_requirements = set(fixture.get("future_runtime_task_card_requirements") or [])
    missing_requirements = sorted(EXPECTED_RUNTIME_REQUIREMENTS - runtime_requirements)
    require(not missing_requirements, f"missing future runtime task card requirements: {missing_requirements}")

    failures = {str(item.get("code")) for item in fixture.get("failure_mapping") or [] if isinstance(item, dict)}
    missing_failures = sorted(EXPECTED_FAILURE_CODES - failures)
    require(not missing_failures, f"missing failure codes: {missing_failures}")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed = set(diagnostics.get("allowed_fields") or [])
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    require(EXPECTED_DIAGNOSTIC_FIELDS <= allowed, "diagnostic allowlist missing required fields")
    require(EXPECTED_FORBIDDEN_DIAGNOSTICS <= forbidden, "diagnostic forbidden list missing required fields")

    no_fallback = set(fixture.get("no_fallback") or [])
    missing_no_fallback = sorted(EXPECTED_NO_FALLBACK - no_fallback)
    require(not missing_no_fallback, f"missing no fallback clauses: {missing_no_fallback}")


def assert_side_effects_and_artifacts(fixture: dict[str, Any]) -> None:
    counters = fixture.get("side_effect_counters") or {}
    non_zero = {key: value for key, value in counters.items() if value != 0}
    require(not non_zero, f"side effect counters must all be zero: {non_zero}")
    for required_counter in {
        "operator_approval_runtime_created_count",
        "operator_approval_runtime_execution_count",
        "approval_executor_created_count",
        "operator_identity_provider_call_count",
        "ticket_verifier_call_count",
        "change_window_verifier_call_count",
        "policy_evaluator_execution_count",
        "audit_event_write_count",
        "production_api_call_count",
    }:
        require(required_counter in counters, f"missing side effect counter: {required_counter}")

    guard = fixture.get("artifact_guard") or {}
    allowed = set(guard.get("allowed_new_artifacts") or [])
    require(EXPECTED_ARTIFACTS <= allowed, "artifact guard missing allowed static artifacts")
    for path in EXPECTED_ARTIFACTS:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing on disk: {path}")
    forbidden = set(guard.get("forbidden_artifacts") or [])
    for phrase in {
        "operator approval runtime",
        "approval executor",
        "operator identity provider adapter",
        "audit store runtime",
        "credential handle runtime",
        "backend health runtime",
        "production resolver runtime",
        "public production API",
    }:
        require(phrase in forbidden, f"artifact guard missing forbidden artifact: {phrase}")


def assert_upstream_alignment() -> None:
    operator_evidence = load_json(OPERATOR_EVIDENCE_PATH)
    approval_boundary = operator_evidence.get("approval_boundary") or {}
    require(
        approval_boundary.get("operator_approval_runtime_status") == "not_created",
        "operator approval evidence readiness must not create runtime",
    )
    require(
        approval_boundary.get("operator_approval_execution_status") == "not_executed",
        "operator approval evidence readiness must not execute runtime",
    )
    require(
        approval_boundary.get("operator_approval_executor_status") == "not_created",
        "operator approval evidence readiness must not create executor",
    )

    audit_entry = load_json(AUDIT_RUNTIME_ENTRY_PATH)
    audit_boundary = audit_entry.get("entry_boundary") or {}
    require(
        audit_boundary.get("audit_store_runtime_status") == "not_created",
        "audit store runtime entry review must keep runtime not_created",
    )
    require(audit_boundary.get("audit_writer_status") == "not_created", "audit writer must stay not_created")
    require(
        audit_boundary.get("audit_event_delivery_status") == "not_executed",
        "audit event delivery must stay not_executed",
    )

    health_entry = load_json(HEALTH_RUNTIME_ENTRY_PATH)
    health_boundary = health_entry.get("entry_boundary") or {}
    require(
        health_boundary.get("backend_health_runtime_status") == "not_created",
        "backend health runtime entry review must keep runtime not_created",
    )
    require(
        health_boundary.get("backend_health_check_status") == "not_executed",
        "backend health check must stay not_executed",
    )

    real_entry = load_json(REAL_RESOLVER_ENTRY_REVIEW_PATH)
    real_boundary = real_entry.get("entry_boundary") or {}
    require(
        real_boundary.get("resolver_runtime_status") == "not_created",
        "real resolver runtime entry review must keep production resolver runtime not_created",
    )

    secret_reference = load_json(SECRET_REFERENCE_PATH)
    policy = secret_reference.get("policy") or {}
    require(policy.get("resolver_enabled") is False, "secret reference runtime must stay disabled")


def assert_implementation_readiness_alignment() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    require(
        target.get("operator_approval_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "implementation readiness missing operator approval runtime implementation entry review status",
    )
    require(
        target.get("operator_approval_runtime_status") == "not_created",
        "operator approval runtime must remain not_created in implementation readiness",
    )
    require(
        target.get("operator_approval_runtime_execution_status") == "not_executed",
        "operator approval runtime execution must remain not_executed in implementation readiness",
    )
    require(
        target.get("audit_store_runtime_status") == "not_created",
        "audit store runtime must remain not_created in implementation readiness",
    )
    require(
        target.get("backend_health_runtime_status") == "not_created",
        "backend health runtime must remain not_created in implementation readiness",
    )
    planned = {
        str(item.get("id")): item
        for item in readiness.get("planned_slices") or []
        if isinstance(item, dict)
    }
    item = planned.get("operator-approval-runtime-implementation-entry-review")
    require(item is not None, "implementation readiness missing operator approval runtime implementation entry review slice")
    require(
        item.get("status") == "operator_approval_runtime_implementation_entry_review_defined",
        "operator approval runtime implementation entry review slice status drifted",
    )
    evidence = set(item.get("evidence") or [])
    require(EXPECTED_ARTIFACTS <= evidence, "operator approval runtime implementation entry review evidence incomplete")
    validation = set(readiness.get("validation_strategy") or [])
    require(
        "operator approval runtime implementation entry review blocked before task card" in validation,
        "implementation readiness validation missing operator approval runtime entry review",
    )
    consumers = set(readiness.get("required_consumers") or [])
    require(
        "scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py"
        in consumers,
        "implementation readiness consumers missing operator approval runtime implementation entry review checker",
    )


def assert_check_repo_order() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current = 'run_python_script("check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py", [])'
    after = 'run_python_script("check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py", [])'
    before = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    require(current in check_repo, "check-repo.py must run operator approval runtime implementation entry review check")
    require(after in check_repo, "check-repo.py missing audit store runtime implementation entry review check")
    require(before in check_repo, "check-repo.py missing startup supervisor check")
    require(
        check_repo.index(after) < check_repo.index(current) < check_repo.index(before),
        "operator approval runtime implementation entry review check must run after audit store entry review and before startup supervisor",
    )


def assert_docs() -> None:
    for relative_path, required_literals in EXPECTED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def assert_no_secret_literals() -> None:
    text = "\n".join(
        [
            FIXTURE_PATH.read_text(encoding="utf-8"),
            read("docs/platform/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.md"),
            read("docs/task-cards/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1-plan.md"),
        ]
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"operator approval runtime entry review contains forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", text) is None,
        "operator approval runtime entry review contains forbidden secret-looking sk token",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_boundary(fixture)
    assert_lists(fixture)
    assert_side_effects_and_artifacts(fixture)
    assert_upstream_alignment()
    assert_implementation_readiness_alignment()
    assert_check_repo_order()
    assert_docs()
    assert_no_secret_literals()
    print("production ops secret backend operator approval runtime implementation entry review checks passed.")


if __name__ == "__main__":
    main()
