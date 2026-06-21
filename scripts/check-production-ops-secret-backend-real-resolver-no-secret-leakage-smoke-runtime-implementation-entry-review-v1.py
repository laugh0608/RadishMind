#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
STRATEGY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.json"
)
REAL_RESOLVER_ENTRY_REFRESH_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.json"
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
    "no_secret_leakage_smoke_runtime_ready",
    "no_secret_leakage_smoke_runtime_created",
    "no_secret_leakage_smoke_runtime_task_card_created",
    "no_secret_leakage_smoke_runner_created",
    "no_secret_leakage_smoke_executed",
    "no_secret_leakage_artifact_scanner_created",
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
    "backend_health_runtime_created",
    "backend_health_check_executed",
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
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1": (
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
    ),
    "production-secret-backend-real-resolver-runtime-implementation-entry-review-v1": (
        "real_resolver_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1": (
        "real_resolver_runtime_implementation_entry_refresh_defined"
    ),
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "resolver_backend_profile_selection_readiness_defined"
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1": (
        "credential_handle_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1": (
        "operator_approval_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-audit-store-runtime-implementation-entry-review-v1": (
        "audit_store_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1": (
        "resolver_backend_health_runtime_implementation_entry_review_defined"
    ),
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied_reference_only_resolver_disabled",
}

EXPECTED_ENTRY_FIELDS = {
    "entry_review_status": "blocked_before_runtime_task_card",
    "runtime_task_decision": (
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_blocked_before_task_card"
    ),
    "strategy_status": "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined",
    "smoke_runtime_implementation_status": "not_started",
    "smoke_runtime_status": "not_created",
    "smoke_runner_status": "not_created",
    "smoke_execution_status": "not_executed",
    "artifact_scanner_status": "not_created",
    "synthetic_fixture_source_status": "required_before_runtime",
    "scan_surface_contract_status": "defined_static_only",
    "input_allowlist_status": "reference_only_fields_fixed",
    "output_allowlist_status": "sanitized_metadata_only",
    "probe_matrix_status": "defined_static_only",
    "artifact_scan_status": "required_before_runtime",
    "fake_resolver_substitution_status": "forbidden",
    "backend_profile_binding_status": "defined_static_only",
    "environment_binding_status": "defined_no_cross_environment",
    "provider_profile_binding_status": "defined_static_only",
    "secret_ref_binding_status": "reference_only",
    "operator_approval_runtime_status": "not_created",
    "operator_approval_runtime_execution_status": "not_executed",
    "audit_store_runtime_status": "not_created",
    "audit_store_status": "not_created",
    "audit_writer_status": "not_created",
    "audit_event_delivery_status": "not_executed",
    "credential_handle_runtime_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "backend_health_check_status": "not_executed",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "cloud_secret_service_status": "not_selected",
    "database_connection_provider_status": "blocked",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "runtime_task_card_created_in_this_slice",
    "smoke_runtime_created_in_this_slice",
    "smoke_runner_created_in_this_slice",
    "smoke_execution_performed_in_this_slice",
    "artifact_scanner_created_in_this_slice",
    "smoke_output_fixture_created_in_this_slice",
    "fake_resolver_call_executed_in_this_slice",
    "production_resolver_call_executed_in_this_slice",
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
    "backend_health_runtime_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "cloud_secret_service_enabled",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_BLOCKED = {
    "no_secret_leakage_smoke_runtime_task_card_not_created",
    "no_secret_leakage_smoke_runtime_not_created",
    "no_secret_leakage_smoke_runner_not_created",
    "no_secret_leakage_smoke_execution_not_performed",
    "synthetic_placeholder_fixture_not_created",
    "artifact_scanner_not_created",
    "credential_handle_runtime_not_created",
    "operator_approval_runtime_not_created",
    "audit_store_runtime_not_created",
    "backend_health_runtime_not_created",
    "production_resolver_runtime_not_created",
    "cloud_secret_service_not_selected",
    "repository_mode_disabled",
}

EXPECTED_GATE_STATUS = {
    "no_leakage_strategy_consumed": "satisfied_static_strategy",
    "runtime_task_card_gate": "blocked_before_task_card",
    "smoke_runtime": "not_created",
    "smoke_runner_execution": "not_created_not_executed",
    "synthetic_fixture_source": "required_before_runtime",
    "scan_surfaces": "satisfied_static_only",
    "input_allowlist": "reference_only_fields_fixed",
    "output_allowlist": "sanitized_metadata_only",
    "negative_probe_matrix": "satisfied_static_only",
    "artifact_scan": "required_before_runtime",
    "fake_resolver_substitution": "forbidden",
    "credential_handle_runtime": "blocked_runtime_not_created",
    "operator_approval_runtime": "blocked_runtime_not_created",
    "audit_store_runtime": "blocked_store_not_created",
    "backend_health_runtime": "blocked_runtime_not_created",
    "production_resolver_runtime": "not_created",
    "cloud_secret_service_gate": "forbidden",
    "database_connection_provider_gate": "blocked",
    "repository_mode_gate": "blocked",
    "production_api_gate": "blocked",
}

EXPECTED_RUNTIME_REQUIREMENTS = {
    "disabled-by-default runtime gate",
    "synthetic placeholder fixture only",
    "reference-only resolver input scan",
    "sanitized success output scan",
    "sanitized failure output scan",
    "audit metadata scan",
    "smoke summary scan",
    "negative probe matrix",
    "artifact scanner",
    "fail-closed leakage detection",
    "no fake resolver substitution",
    "no repository mode side effect",
    "offline unit test and static smoke",
    "side effect counters",
}

EXPECTED_FAILURE_CODES = {
    "real_resolver_no_leakage_runtime_entry_strategy_missing",
    "real_resolver_no_leakage_runtime_entry_task_card_blocked",
    "real_resolver_no_leakage_runtime_entry_scan_surface_missing",
    "real_resolver_no_leakage_runtime_entry_input_allowlist_missing",
    "real_resolver_no_leakage_runtime_entry_output_allowlist_missing",
    "real_resolver_no_leakage_runtime_entry_probe_matrix_missing",
    "real_resolver_no_leakage_runtime_entry_synthetic_fixture_missing",
    "real_resolver_no_leakage_runtime_entry_secret_material_detected",
    "real_resolver_no_leakage_runtime_entry_credential_payload_detected",
    "real_resolver_no_leakage_runtime_entry_diagnostic_exposure_detected",
    "real_resolver_no_leakage_runtime_entry_audit_exposure_detected",
    "real_resolver_no_leakage_smoke_runtime_created_in_entry_review",
    "real_resolver_no_leakage_smoke_executed_in_entry_review",
    "real_resolver_no_leakage_runtime_entry_fallback_forbidden",
    "real_resolver_no_leakage_runtime_entry_repository_mode_forbidden",
    "real_resolver_no_leakage_runtime_entry_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "no_leakage_runtime_entry_status",
    "strategy_status",
    "runtime_task_decision",
    "smoke_runtime_status",
    "smoke_runner_status",
    "smoke_execution_status",
    "scan_surface_status",
    "input_allowlist_status",
    "output_allowlist_status",
    "probe_matrix_status",
    "artifact_scan_status",
    "synthetic_fixture_source_status",
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
    "raw_smoke_input",
    "raw_smoke_output",
    "raw_request_payload",
    "raw_response_payload",
}

EXPECTED_NO_SIDE_EFFECT_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "smoke_runtime_created_count",
    "smoke_runner_created_count",
    "smoke_runtime_execution_count",
    "artifact_scanner_created_count",
    "smoke_output_fixture_created_count",
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
    "backend_health_check_count",
    "database_connection_count",
    "driver_open_count",
    "sql_execution_count",
    "schema_marker_read_count",
    "schema_marker_write_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}

EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md",
    "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1-plan.md",
    "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
    "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py",
}

EXPECTED_FORBIDDEN_ARTIFACTS = {
    "no secret leakage smoke runtime implementation task card",
    "no secret leakage smoke runtime",
    "no secret leakage smoke runner",
    "no secret leakage artifact scanner",
    "no secret leakage smoke output fixture",
    "production resolver runtime",
    "production resolver backend client",
    "cloud secret SDK",
    "secret value fixture",
    "production credential file",
    "credential payload runtime",
    "credential handle runtime",
    "approval runtime",
    "approval executor",
    "audit store runtime",
    "audit writer",
    "backend health runtime",
    "backend health client",
    "database connection provider",
    "DB driver / DSN parser",
    "connection factory",
    "SQL migration",
    "schema marker reader / writer",
    "migration runner",
    "workflow saved draft repository mode runtime",
    "public production API",
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
        slice_info.get("id")
        == "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1",
        "unexpected slice id",
    )
    require(
        slice_info.get("status")
        == "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
        "unexpected slice status",
    )
    require(slice_info.get("task_card") in EXPECTED_ALLOWED_ARTIFACTS, "unexpected task card path")
    require(slice_info.get("platform_topic") in EXPECTED_ALLOWED_ARTIFACTS, "unexpected platform topic path")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = {str(item.get("id")): item for item in fixture.get("depends_on") or [] if isinstance(item, dict)}
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, expected_status in EXPECTED_DEPENDENCIES.items():
        dependency = dependencies[dependency_id]
        require(dependency.get("status") == expected_status, f"{dependency_id} status drifted")
        evidence = dependency.get("evidence")
        require(isinstance(evidence, str) and (REPO_ROOT / evidence).exists(), f"{dependency_id} evidence missing")


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("entry_boundary") or {}
    for field, expected in EXPECTED_ENTRY_FIELDS.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"{field} must be false")


def assert_blocked_and_gates(fixture: dict[str, Any]) -> None:
    blocked = set(fixture.get("blocked_conditions") or [])
    missing_blocked = sorted(EXPECTED_BLOCKED - blocked)
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")

    gates = {str(item.get("gate")): item.get("status") for item in fixture.get("gate_matrix") or []}
    missing_gates = sorted(set(EXPECTED_GATE_STATUS) - set(gates))
    require(not missing_gates, f"missing gate matrix entries: {missing_gates}")
    for gate, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate] == expected_status, f"{gate} status drifted")


def assert_runtime_requirements(fixture: dict[str, Any]) -> None:
    requirements = set(fixture.get("future_runtime_task_card_requirements") or [])
    missing = sorted(EXPECTED_RUNTIME_REQUIREMENTS - requirements)
    require(not missing, f"missing runtime requirements: {missing}")


def assert_failure_and_diagnostics(fixture: dict[str, Any]) -> None:
    failure_codes = {str(item.get("code")) for item in fixture.get("failure_mapping") or [] if isinstance(item, dict)}
    missing_codes = sorted(EXPECTED_FAILURE_CODES - failure_codes)
    require(not missing_codes, f"missing failure codes: {missing_codes}")

    diagnostics = fixture.get("diagnostics_contract") or {}
    missing_allowed = sorted(EXPECTED_DIAGNOSTIC_FIELDS - set(diagnostics.get("allowed_fields") or []))
    require(not missing_allowed, f"missing diagnostic fields: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_DIAGNOSTICS - set(diagnostics.get("forbidden_fields") or []))
    require(not missing_forbidden, f"missing forbidden diagnostic fields: {missing_forbidden}")


def assert_no_fallback_and_side_effects(fixture: dict[str, Any]) -> None:
    no_fallback = set(fixture.get("no_fallback") or [])
    for forbidden in {
        "fake resolver runtime",
        "developer env plaintext",
        "fixture credential",
        "mock provider",
        "local-smoke profile",
        "repository memory store",
        "audit memory store",
    }:
        require(forbidden in no_fallback, f"missing no fallback source: {forbidden}")

    counters = fixture.get("no_side_effect_counters") or {}
    missing_counters = sorted(EXPECTED_NO_SIDE_EFFECT_COUNTERS - set(counters))
    require(not missing_counters, f"missing no-side-effect counters: {missing_counters}")
    non_zero = {key: value for key, value in counters.items() if value != 0}
    require(not non_zero, f"side effect counters must stay zero: {non_zero}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    missing_allowed = sorted(EXPECTED_ALLOWED_ARTIFACTS - set(guard.get("allowed_new_artifacts") or []))
    require(not missing_allowed, f"missing allowed artifacts: {missing_allowed}")
    for path in EXPECTED_ALLOWED_ARTIFACTS:
        require((REPO_ROOT / path).exists(), f"allowed artifact missing on disk: {path}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_ARTIFACTS - set(guard.get("forbidden_artifacts") or []))
    require(not missing_forbidden, f"missing forbidden artifacts: {missing_forbidden}")


def assert_implementation_readiness_alignment() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "implementation readiness no leakage runtime entry review status drifted",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_status") == "not_created",
        "implementation readiness must keep no leakage smoke runtime not_created",
    )
    require(target.get("resolver_runtime_status") == "not_created", "resolver runtime must remain not_created")
    require(
        target.get("production_resolver_runtime_task_card_status") == "not_created",
        "production resolver runtime task card must remain not_created",
    )
    require(
        target.get("production_secret_backend_status") == "not_satisfied",
        "production secret backend must remain not_satisfied",
    )

    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or [] if isinstance(item, dict)}
    slice_id = "real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review"
    require(slice_id in planned, "implementation readiness missing no leakage smoke runtime entry review slice")
    require(
        planned[slice_id].get("status")
        == "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
        "implementation readiness no leakage smoke runtime entry review planned status drifted",
    )
    evidence = set(planned[slice_id].get("evidence") or [])
    for path in EXPECTED_ALLOWED_ARTIFACTS:
        require(path in evidence, f"implementation readiness planned slice missing evidence: {path}")

    validation = set(readiness.get("validation_strategy") or [])
    require(
        "real resolver no leakage smoke runtime implementation entry review blocked before task card"
        in validation,
        "implementation readiness validation strategy missing no leakage smoke runtime entry review",
    )
    consumers = set(readiness.get("required_consumers") or [])
    for path in EXPECTED_ALLOWED_ARTIFACTS:
        require(path in consumers or path == FIXTURE_PATH.relative_to(REPO_ROOT).as_posix(), f"consumer list missing {path}")


def assert_upstream_alignment() -> None:
    strategy = load_json(STRATEGY_PATH)
    require(
        strategy.get("slice", {}).get("status") == "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined",
        "strategy fixture status drifted",
    )
    refresh = load_json(REAL_RESOLVER_ENTRY_REFRESH_PATH)
    boundary = refresh.get("entry_refresh_boundary") or {}
    require(
        boundary.get("resolver_runtime_status") == "not_created",
        "real resolver entry refresh must still keep resolver runtime not_created",
    )
    gates = {str(item.get("gate_id")): item.get("status") for item in refresh.get("entry_gate_matrix") or []}
    require(
        gates.get("no_secret_leakage_smoke_runtime") == "strategy_defined_runtime_not_created",
        "real resolver entry refresh must still keep no leakage runtime not_created",
    )
    secret_ref = load_json(SECRET_REFERENCE_PATH)
    require(secret_ref.get("kind") == "production_secret_reference_manifest", "secret reference fixture drifted")


def assert_docs_and_check_repo() -> None:
    required_doc_literals = {
        "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md": [
            "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
            "real_resolver_no_secret_leakage_smoke_runtime_implementation_blocked_before_task_card",
            "不创建 no secret leakage smoke runtime implementation task card",
            "不读取真实 secret",
            "不调用云 secret 服务",
            "不启用 repository mode",
        ],
        "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1-plan.md": [
            "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
            "real_resolver_no_secret_leakage_smoke_runtime_implementation_blocked_before_task_card",
            "不创建 no leakage smoke runtime implementation task card",
            "不执行 smoke runtime",
        ],
        "docs/platform/README.md": [
            "Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Review v1",
            "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
        ],
        "docs/features/README.md": [
            "Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Review v1",
            "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
        ],
        "docs/task-cards/README.md": [
            "Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Review",
            "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
        ],
        "docs/radishmind-current-focus.md": [
            "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1",
            "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
        ],
        "scripts/README.md": [
            "check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py",
            "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
        ],
    }
    for relative_path, literals in required_doc_literals.items():
        text = read(relative_path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run no leakage smoke runtime implementation entry review check",
    )


def assert_no_secret_literals() -> None:
    paths = [
        FIXTURE_PATH,
        REPO_ROOT
        / "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md",
        REPO_ROOT
        / "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1-plan.md",
    ]
    text = "\n".join(path.read_text(encoding="utf-8") for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"entry review contains forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", text) is None,
        "entry review contains forbidden secret-looking sk token",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_boundary(fixture)
    assert_blocked_and_gates(fixture)
    assert_runtime_requirements(fixture)
    assert_failure_and_diagnostics(fixture)
    assert_no_fallback_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment()
    assert_upstream_alignment()
    assert_docs_and_check_repo()
    assert_no_secret_literals()
    print("production ops secret backend real resolver no leakage smoke runtime implementation entry review checks passed.")


if __name__ == "__main__":
    main()
