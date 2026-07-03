#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
ENTRY_REVIEW_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json"
)
CLOUD_SELECTION_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json"
)
BLOCKER_CONSOLIDATION_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.json"
)
CREDENTIAL_REFRESH_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.json"
)
OPERATOR_REFRESH_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.json"
)
AUDIT_REFRESH_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json"
)
NO_LEAKAGE_ENTRY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json"
)
BACKEND_PROFILE_SELECTION_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json"
)
SECRET_REFERENCE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-reference-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-cloud-secret-service-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json",
        "cloud_secret_service_selection_readiness_defined",
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
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
        "resolver_backend_profile_selection_readiness_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
    "secret-ref-schema-and-fixtures": (
        "scripts/checks/fixtures/production-secret-reference-basic.json",
        "satisfied_reference_only_resolver_disabled",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "production_ready",
    "production_secret_backend_ready",
    "cloud_secret_service_ready",
    "cloud_secret_service_selected",
    "backend_health_runtime_task_card_created",
    "backend_health_runtime_created",
    "backend_health_client_created",
    "backend_health_probe_created",
    "backend_health_check_executed",
    "provider_call_executed",
    "production_resolver_runtime_task_card_created",
    "production_resolver_runtime_created",
    "credential_handle_runtime_ready",
    "operator_approval_runtime_ready",
    "audit_store_runtime_ready",
    "no_secret_leakage_smoke_runtime_created",
    "repository_mode_ready",
    "database_connection_provider_ready",
    "production_api_ready",
}
EXPECTED_REFRESH_FIELDS = {
    "status": "resolver_backend_health_runtime_implementation_entry_refresh_defined",
    "entry_decision": "resolver_backend_health_runtime_task_card_still_blocked_after_refresh",
    "original_entry_review_status": "blocked_before_runtime_task_card",
    "production_resolver_blocker_consolidation_status": "production_resolver_runtime_blocker_consolidation_defined",
    "cloud_secret_service_selection_status": "cloud_secret_service_selection_readiness_defined",
    "cloud_secret_service_status": "not_selected",
    "backend_profile_binding_status": "defined_static_only",
    "backend_health_runtime_task_card_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "backend_health_client_status": "not_created",
    "backend_health_probe_status": "not_created",
    "backend_health_check_status": "not_executed",
    "credential_handle_runtime_status": "blocked_before_runtime_task_card",
    "operator_approval_runtime_status": "blocked_before_runtime_task_card",
    "audit_store_runtime_status": "blocked_before_runtime_task_card",
    "no_secret_leakage_smoke_runtime_status": "blocked_before_runtime_task_card",
    "production_resolver_runtime_status": "not_created",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}
EXPECTED_FALSE_FIELDS = {
    "runtime_task_card_allowed_now",
    "runtime_task_card_created_in_this_slice",
    "backend_health_runtime_created_in_this_slice",
    "backend_health_client_created_in_this_slice",
    "backend_health_probe_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "provider_call_executed_in_this_slice",
    "cloud_secret_call_executed_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "operator_approval_runtime_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "no_secret_leakage_smoke_runtime_created_in_this_slice",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}
EXPECTED_REFRESH_GATES = {
    "original_backend_health_entry_review": "blocked_before_runtime_task_card",
    "production_resolver_blocker_consolidation": "production_resolver_runtime_task_card_still_blocked_after_consolidation",
    "cloud_secret_service_selection": "not_selected",
    "credential_handle_runtime": "blocked_before_runtime_task_card",
    "operator_approval_runtime": "blocked_before_runtime_task_card",
    "audit_store_runtime": "blocked_before_runtime_task_card",
    "no_secret_leakage_smoke_runtime": "blocked_before_runtime_task_card",
    "backend_health_runtime_task_card": "not_created",
    "repository_mode_runtime": "disabled",
}
EXPECTED_FAILURE_CODES = {
    "resolver_backend_health_runtime_entry_refresh_dependency_missing",
    "resolver_backend_health_runtime_entry_refresh_task_card_forbidden",
    "resolver_backend_health_runtime_entry_refresh_runtime_created_forbidden",
    "resolver_backend_health_runtime_entry_refresh_health_check_forbidden",
    "resolver_backend_health_runtime_entry_refresh_cloud_selection_blocked",
    "resolver_backend_health_runtime_entry_refresh_credential_handle_blocked",
    "resolver_backend_health_runtime_entry_refresh_operator_approval_blocked",
    "resolver_backend_health_runtime_entry_refresh_audit_store_blocked",
    "resolver_backend_health_runtime_entry_refresh_no_leakage_blocked",
    "resolver_backend_health_runtime_entry_refresh_secret_material_detected",
    "resolver_backend_health_runtime_entry_refresh_repository_mode_forbidden",
    "resolver_backend_health_runtime_entry_refresh_scope_overreach",
}
EXPECTED_DIAGNOSTIC_FIELDS = {
    "backend_health_runtime_entry_refresh_status",
    "entry_decision",
    "gate",
    "gate_status",
    "backend_health_runtime_status",
    "backend_health_check_status",
    "backend_profile_status",
    "cloud_secret_service_selection_status",
    "credential_handle_runtime_status",
    "operator_approval_runtime_status",
    "audit_store_runtime_status",
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
    "authorization_header",
    "cookie",
    "raw_health_request",
    "raw_health_response",
    "raw_backend_error_detail",
    "raw_provider_response",
}
EXPECTED_COUNTERS = {
    "backend_health_task_card_created_count",
    "backend_health_runtime_created_count",
    "backend_health_client_created_count",
    "backend_health_probe_created_count",
    "backend_health_check_count",
    "provider_call_count",
    "cloud_secret_call_count",
    "network_call_count",
    "real_secret_read_count",
    "environment_secret_read_count",
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "credential_handle_runtime_created_count",
    "operator_approval_runtime_execution_count",
    "audit_store_runtime_created_count",
    "audit_event_write_count",
    "no_secret_leakage_smoke_runtime_created_count",
    "database_connection_count",
    "sql_execution_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}
EXPECTED_REQUIRED_CHECKS = {
    "run resolver backend health runtime implementation entry refresh checker",
    "run resolver backend health runtime implementation entry review checker",
    "run cloud secret service selection readiness checker",
    "run production resolver runtime blocker consolidation checker",
    "run credential handle runtime implementation entry refresh checker",
    "run operator approval runtime implementation entry refresh checker",
    "run audit store runtime implementation entry refresh v3 checker",
    "run real resolver no leakage smoke runtime implementation entry review checker",
    "run implementation readiness checker",
    "run git diff check",
    "run fast repository check",
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
        == "production_ops_secret_backend_resolver_backend_health_runtime_implementation_entry_refresh_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id")
        == "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1",
        "unexpected slice id",
    )
    require(
        slice_info.get("status") == "resolver_backend_health_runtime_implementation_entry_refresh_defined",
        "unexpected slice status",
    )
    require(
        slice_info.get("entry_decision") == "resolver_backend_health_runtime_task_card_still_blocked_after_refresh",
        "unexpected entry decision",
    )
    for key, expected in {
        "task_card": (
            "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1-plan.md"
        ),
        "platform_topic": (
            "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.md"
        ),
    }.items():
        value = str(slice_info.get(key) or "")
        require(value == expected, f"unexpected {key}")
        require((REPO_ROOT / value).exists(), f"{key} path missing: {value}")
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        require(item.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        dependency = load_json(REPO_ROOT / relative_path)
        if dependency_id == "secret-ref-schema-and-fixtures":
            require(dependency.get("scope") == "secret_reference_only", "secret reference must remain reference-only")
            continue
        require((dependency.get("slice") or {}).get("status") == expected_status, f"{dependency_id} fixture status drifted")


def assert_refresh_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("refresh_boundary") or {}
    for field, expected in EXPECTED_REFRESH_FIELDS.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in EXPECTED_FALSE_FIELDS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_refresh_matrix(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "refresh_matrix", "gate")
    require(set(gates) == set(EXPECTED_REFRESH_GATES), "refresh gate ids drifted")
    for gate, expected_status in EXPECTED_REFRESH_GATES.items():
        item = gates[gate]
        require(item.get("status") == expected_status, f"{gate} status drifted")
        require(item.get("source"), f"{gate} source missing")
        if gate != "repository_mode_runtime":
            require(item.get("blocks_backend_health_task_card") is True, f"{gate} must block backend health task card")
        require(item.get("blocks_production_resolver_task_card") is True, f"{gate} must block resolver task card")

    requirements = set(fixture.get("future_runtime_task_card_requirements") or [])
    for requirement in {
        "cloud secret service concrete provider selection must be separate task",
        "backend health runtime task card must require disabled-by-default runtime gate",
        "backend health result must remain metadata-only",
        "credential handle runtime entry must no longer be blocked",
        "operator approval runtime entry must no longer be blocked",
        "audit store runtime entry must no longer be blocked",
        "no leakage smoke runtime entry must no longer be blocked",
        "provider call must be covered by offline unit tests before live execution",
        "diagnostics must remain sanitized reference-only",
        "side effect counters must remain zero during entry refresh",
    }:
        require(requirement in requirements, f"missing future task requirement: {requirement}")


def assert_alignment_with_existing_evidence() -> None:
    entry = load_json(ENTRY_REVIEW_PATH)
    entry_boundary = entry.get("entry_boundary") or {}
    require(
        entry_boundary.get("entry_review_status") == "blocked_before_runtime_task_card",
        "backend health entry review status drifted",
    )
    require(entry_boundary.get("backend_health_runtime_status") == "not_created", "entry review must not create runtime")
    require(entry_boundary.get("backend_health_check_status") == "not_executed", "entry review must not execute health")

    cloud = load_json(CLOUD_SELECTION_PATH)
    cloud_boundary = cloud.get("selection_boundary") or {}
    require(
        cloud_boundary.get("concrete_cloud_vendor_selection_status") == "not_selected",
        "cloud selection must remain not_selected",
    )
    require(cloud_boundary.get("cloud_secret_service_called_in_this_slice") is False, "cloud call must remain false")

    consolidation = load_json(BLOCKER_CONSOLIDATION_PATH)
    blockers = rows_by_id(consolidation, "blocker_matrix", "blocker_id")
    require(
        blockers["backend_health_runtime"].get("status") == "blocked_before_runtime_task_card",
        "blocker consolidation backend health status drifted",
    )

    credential = load_json(CREDENTIAL_REFRESH_PATH)
    require(
        (credential.get("refresh_boundary") or {}).get("backend_health_runtime_status")
        == "blocked_before_runtime_task_card",
        "credential refresh backend health dependency drifted",
    )

    operator = load_json(OPERATOR_REFRESH_PATH)
    require(
        (operator.get("refresh_boundary") or {}).get("backend_health_runtime_status")
        == "blocked_before_runtime_task_card",
        "operator refresh backend health dependency drifted",
    )

    audit = load_json(AUDIT_REFRESH_PATH)
    require(
        (audit.get("entry_refresh_boundary") or {}).get("entry_refresh_status") == "blocked_before_runtime_task_card",
        "audit refresh v3 status drifted",
    )

    no_leakage = load_json(NO_LEAKAGE_ENTRY_PATH)
    require(
        (no_leakage.get("entry_boundary") or {}).get("entry_review_status") == "blocked_before_runtime_task_card",
        "no leakage entry review status drifted",
    )

    backend_profile = load_json(BACKEND_PROFILE_SELECTION_PATH)
    require(
        (backend_profile.get("selection_boundary") or {}).get("cloud_secret_service_status") == "not_selected",
        "backend profile cloud selection status drifted",
    )

    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in {
        "resolver_backend_health_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "resolver_backend_health_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "backend_health_runtime_status": "not_created",
        "backend_health_check_status": "not_executed",
        "cloud_secret_service_selection_readiness_status": "defined_without_cloud_backend_selection",
        "production_resolver_runtime_task_card_status": "not_created",
        "credential_handle_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "operator_approval_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "audit_store_runtime_implementation_entry_refresh_v3_status": "blocked_before_runtime_task_card",
        "real_resolver_no_secret_leakage_smoke_runtime_status": "not_created",
    }.items():
        require(target.get(field) == expected, f"implementation readiness target {field} drifted")

    planned = rows_by_id(readiness, "planned_slices", "id")
    planned_slice = planned.get("resolver-backend-health-runtime-implementation-entry-refresh")
    require(planned_slice is not None, "implementation readiness missing backend health entry refresh planned slice")
    require(
        planned_slice.get("status") == "resolver_backend_health_runtime_implementation_entry_refresh_defined",
        "planned slice status drifted",
    )


def assert_failure_diagnostics_and_policies(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in failures.items():
        require(item.get("boundary"), f"{code} must define boundary")
        require(item.get("sanitized_diagnostic"), f"{code} must define diagnostic")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    require(EXPECTED_DIAGNOSTIC_FIELDS.issubset(set(diagnostics.get("allowed_fields") or [])), "allowed diagnostics drifted")
    require(EXPECTED_FORBIDDEN_DIAGNOSTICS.issubset(set(diagnostics.get("forbidden_fields") or [])), "forbidden diagnostics drifted")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must be false")
    require(diagnostics.get("secret_ref_value_allowed_in_diagnostics") is False, "secret ref value must not be emitted")

    fallback = set(fixture.get("no_fallback_policy") or [])
    for rule in (
        "no fallback from backend health runtime refresh to fake resolver",
        "no fallback from production backend profile to test backend profile",
        "no fallback from missing cloud selection to local smoke profile",
        "no fallback from missing credential handle runtime to fixture credential",
        "no runtime task card if operator approval runtime is blocked",
        "no runtime task card if audit store runtime is blocked",
        "no runtime task card if no leakage smoke runtime is blocked",
        "backend health runtime refresh does not mean workflow repository mode ready",
    ):
        require(rule in fallback, f"missing no fallback rule: {rule}")

    side_effect_policy = set(fixture.get("no_side_effect_policy") or [])
    for rule in (
        "no backend health runtime task card creation",
        "no backend health runtime creation",
        "no backend health client creation",
        "no backend health probe creation",
        "no backend health check",
        "no provider call",
        "no cloud secret service call",
        "no network call",
        "no environment secret read",
        "no production resolver call",
        "no database connection",
        "no SQL execution",
        "no repository mode enablement",
    ):
        require(rule in side_effect_policy, f"missing no side effect rule: {rule}")

    counters = fixture.get("side_effect_counters") or {}
    require(set(counters) == EXPECTED_COUNTERS, "side effect counters drifted")
    for counter, value in counters.items():
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


def assert_docs_validation_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        text = read(str(reference.get("path") or ""))
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in text]
        require(not missing, f"{reference.get('path')} missing literals: {missing}")

    require(set(fixture.get("validation_strategy") or []) == EXPECTED_REQUIRED_CHECKS, "validation strategy drifted")
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    prior_call = 'run_python_script("check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py", [])'
    current_call = (
        'run_python_script("check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.py", [])'
    )
    next_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call in (prior_call, current_call, next_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(prior_call) < check_repo.index(current_call) < check_repo.index(next_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"backend health entry refresh artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "backend health entry refresh artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "backend health entry refresh artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_refresh_boundary(fixture)
    assert_refresh_matrix(fixture)
    assert_alignment_with_existing_evidence()
    assert_failure_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend resolver backend health runtime implementation entry refresh checks passed.")


if __name__ == "__main__":
    main()
