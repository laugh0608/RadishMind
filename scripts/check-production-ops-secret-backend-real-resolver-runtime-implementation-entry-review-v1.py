#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
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
    "no_secret_leakage_smoke_runtime_created",
    "real_secret_read",
    "real_secret_written",
    "credential_payload_created",
    "credential_handle_created",
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
    "production-secret-backend-real-resolver-runtime-preconditions-v1": "real_resolver_runtime_preconditions_defined",
    "production-secret-backend-fake-resolver-runtime-implementation-v1": "fake_resolver_runtime_test_only_implemented",
    "production-secret-backend-secret-resolver-interface-disabled-readiness-v1": "secret_resolver_interface_disabled_readiness_defined",
    "production-secret-backend-operator-runbook-negative-gates-readiness-v1": "operator_runbook_negative_gates_readiness_defined",
    "production-secret-backend-rotation-audit-policy-readiness-v1": "rotation_audit_policy_readiness_defined",
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "resolver_backend_profile_selection_readiness_defined"
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1": (
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
    ),
    "production-secret-backend-credential-handle-runtime-boundary-readiness-v1": (
        "credential_handle_runtime_boundary_readiness_defined"
    ),
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
}

EXPECTED_GATE_STATUS = {
    "real_resolver_preconditions_consumed": "satisfied_static_preconditions",
    "implementation_task_card_gate": "blocked_before_task_card",
    "resolver_backend_profile_selection": "readiness_defined_without_backend_runtime",
    "no_secret_leakage_smoke_runtime_gate": "strategy_defined_runtime_not_created",
    "credential_handle_runtime_boundary": "readiness_defined_runtime_not_created",
    "operator_approval_runtime_evidence": "blocked_missing_runtime_evidence",
    "audit_rotation_runtime_handoff": "blocked_missing_runtime_handoff",
    "resolver_backend_health_boundary": "blocked_missing_backend_health_boundary",
    "resolver_runtime_gate": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_runtime_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_BLOCKED = {
    "operator-approval-runtime-evidence-readiness",
    "production-secret-audit-store-handoff-readiness",
    "resolver-backend-health-boundary-readiness",
}

EXPECTED_FAILURE_CODES = {
    "real_resolver_runtime_entry_preconditions_missing",
    "real_resolver_runtime_entry_task_card_blocked",
    "real_resolver_runtime_entry_backend_profile_missing",
    "real_resolver_runtime_entry_no_leakage_gate_missing",
    "real_resolver_runtime_entry_credential_handle_boundary_missing",
    "real_resolver_runtime_entry_operator_approval_evidence_missing",
    "real_resolver_runtime_entry_audit_handoff_missing",
    "real_resolver_runtime_entry_secret_value_detected",
    "real_resolver_runtime_created_in_entry_review",
    "real_resolver_runtime_entry_cloud_call_forbidden",
    "real_resolver_runtime_entry_repository_mode_forbidden",
    "real_resolver_runtime_entry_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "real_resolver_runtime_entry_status",
    "real_resolver_runtime_preconditions_status",
    "resolver_runtime_status",
    "resolver_backend_profile_status",
    "no_secret_leakage_gate_status",
    "credential_handle_boundary_status",
    "operator_approval_status",
    "audit_policy_status",
    "rotation_policy_status",
    "failure_code",
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
    "dsn",
    "cloud_credential",
    "database_hostname",
    "database_error_detail",
    "credential_payload",
    "resolver_backend_url",
    "full_secret_ref_value",
    "full_credential_handle",
    "full_user_claim",
    "authorization_header",
    "cookie",
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
        fixture.get("kind") == "production_ops_secret_backend_real_resolver_runtime_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "production-secret-backend-real-resolver-runtime-implementation-entry-review-v1", "unexpected slice id")
    require(slice_info.get("status") == "real_resolver_runtime_implementation_entry_review_defined", "unexpected status")
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md",
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
    for field, expected in {
        "entry_review_status": "blocked_before_runtime_task_card",
        "runtime_task_decision": "real_resolver_runtime_implementation_blocked_before_task_card",
        "real_resolver_runtime_preconditions_status": "real_resolver_runtime_preconditions_defined",
        "resolver_implementation_status": "not_started",
        "resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "runtime_task_card_created_in_this_slice",
        "production_resolver_runtime_created_in_this_slice",
        "cloud_secret_service_enabled",
        "database_connection_provider_enabled",
        "repository_mode_enabled",
        "production_api_enabled",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gates_and_failures(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "entry_gate_matrix", "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "entry gate ids drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    missing_blocked = sorted(EXPECTED_BLOCKED - set(fixture.get("blocked_conditions") or []))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")

    failures = rows_by_id(fixture, "failure_mapping", "code")
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(failures))
    require(not missing_failures, f"missing failure codes: {missing_failures}")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} must define failure boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} must define sanitized diagnostic")
        for forbidden in ("api key", "token", "password", "dsn", "raw secret"):
            require(forbidden not in diagnostic.lower(), f"{code} diagnostic must not expose {forbidden}")


def assert_diagnostics_and_policies(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    missing_allowed = sorted(EXPECTED_DIAGNOSTIC_FIELDS - set(diagnostics.get("allowed_fields") or []))
    require(not missing_allowed, f"missing diagnostic fields: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_DIAGNOSTICS - set(diagnostics.get("forbidden_fields") or []))
    require(not missing_forbidden, f"missing forbidden diagnostic fields: {missing_forbidden}")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must remain disabled")
    require(diagnostics.get("secret_ref_value_allowed_in_diagnostics") is False, "secret ref value must not be emitted")

    require("entry review does not mean production resolver runtime ready" in fixture.get("no_fallback_policy", []), "missing no fallback conclusion")
    require("no production API call" in fixture.get("no_side_effect_policy", []), "missing no production API side effect")
    for counter, value in (fixture.get("side_effect_counters") or {}).items():
        require(value == 0, f"{counter} must stay zero")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("required_static_artifacts_exist") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required static artifact missing: {relative_path}")
    for relative_path in guard.get("files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")
    for source_path in guard.get("source_files_to_scan") or []:
        source = read(str(source_path))
        for literal in guard.get("forbidden_source_literals") or []:
            require(str(literal) not in source, f"{source_path} contains forbidden literal: {literal}")


def assert_implementation_readiness_alignment() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    require(
        target.get("real_resolver_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "implementation readiness must record blocked real resolver runtime entry review",
    )
    require(
        target.get("resolver_backend_profile_selection_readiness_status") == "defined_without_backend_runtime",
        "implementation readiness must record resolver backend profile selection readiness",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_strategy_status") == "defined_without_runtime",
        "implementation readiness must record real resolver no leakage strategy",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_status") == "not_created",
        "implementation readiness must keep real resolver no leakage runtime not_created",
    )
    require(
        target.get("credential_handle_runtime_boundary_readiness_status") == "defined_without_runtime",
        "implementation readiness must record credential handle boundary readiness",
    )
    require(
        target.get("credential_handle_runtime_status") == "not_created",
        "implementation readiness must keep credential handle runtime not_created",
    )
    require(target.get("resolver_implementation_status") == "not_started", "resolver implementation must remain not_started")
    require(target.get("resolver_runtime_status") == "not_created", "resolver runtime must remain not_created")
    planned = rows_by_id(readiness, "planned_slices", "id")
    current = planned.get("real-resolver-runtime-implementation-entry-review") or {}
    require(
        current.get("status") == "real_resolver_runtime_implementation_entry_review_defined",
        "planned entry review status drifted",
    )
    credential = planned.get("credential-handle-runtime-boundary-readiness") or {}
    require(
        credential.get("status") == "credential_handle_runtime_boundary_readiness_defined",
        "planned credential handle boundary status drifted",
    )


def assert_docs_validation_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        text = read(path)
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in text]
        require(not missing, f"{path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    prior_call = 'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py", [])'
    current_call = 'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py", [])'
    next_call = 'run_python_script("check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py", [])'
    for call in (prior_call, current_call, next_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(prior_call) < check_repo.index(current_call) < check_repo.index(next_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"entry review artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "entry review artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "entry review artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_boundary(fixture)
    assert_gates_and_failures(fixture)
    assert_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment()
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend real resolver runtime implementation entry review checks passed.")


if __name__ == "__main__":
    main()
