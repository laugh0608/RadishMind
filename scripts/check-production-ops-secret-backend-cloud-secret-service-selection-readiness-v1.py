#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
BLOCKER_CONSOLIDATION_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.json"
)
BACKEND_PROFILE_SELECTION_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json"
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
BACKEND_HEALTH_ENTRY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json"
)
NO_LEAKAGE_ENTRY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json"
)
SECRET_REFERENCE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-reference-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-production-resolver-runtime-blocker-consolidation-v1": (
        "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.json",
        "production_resolver_runtime_blocker_consolidation_defined",
    ),
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
        "resolver_backend_profile_selection_readiness_defined",
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
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.json",
        "real_resolver_runtime_implementation_entry_refresh_defined",
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
EXPECTED_SELECTION_GATES = {
    "candidate_kind_allowlist": "reserved_only",
    "concrete_cloud_vendor": "not_selected",
    "provider_profile_binding": "required_reference_only",
    "environment_binding": "required_no_cross_environment",
    "secret_ref_namespace": "required_reference_only",
    "credential_handle_runtime": "blocked_before_runtime_task_card",
    "operator_approval_runtime": "blocked_before_runtime_task_card",
    "audit_store_runtime": "blocked_before_runtime_task_card",
    "backend_health_runtime": "blocked_before_runtime_task_card",
    "no_secret_leakage_smoke_runtime": "blocked_before_runtime_task_card",
    "production_resolver_runtime_task_card": "not_created",
    "cloud_secret_sdk": "forbidden",
    "repository_mode_runtime": "disabled",
}
EXPECTED_FAILURE_CODES = {
    "cloud_secret_service_selection_dependency_missing",
    "cloud_secret_service_selection_vendor_selected_forbidden",
    "cloud_secret_service_selection_sdk_binding_forbidden",
    "cloud_secret_service_selection_client_created_forbidden",
    "cloud_secret_service_selection_call_forbidden",
    "cloud_secret_service_selection_secret_material_detected",
    "cloud_secret_service_selection_credential_handle_blocked",
    "cloud_secret_service_selection_operator_approval_blocked",
    "cloud_secret_service_selection_audit_store_blocked",
    "cloud_secret_service_selection_backend_health_blocked",
    "cloud_secret_service_selection_no_leakage_blocked",
    "cloud_secret_service_selection_repository_mode_forbidden",
    "cloud_secret_service_selection_scope_overreach",
}
EXPECTED_DIAGNOSTIC_FIELDS = {
    "cloud_secret_service_selection_status",
    "selection_decision",
    "candidate_kind",
    "candidate_status",
    "gate",
    "gate_status",
    "environment",
    "provider_profile_id",
    "secret_ref_namespace_status",
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
    "raw_token",
    "authorization_header",
    "cookie",
    "raw_claim_dump",
    "jwks_raw_dump",
    "membership_raw_record",
    "provider_error_detail",
    "raw_provider_response",
}
EXPECTED_ZERO_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "cloud_secret_vendor_selected_count",
    "cloud_secret_sdk_import_count",
    "cloud_secret_client_created_count",
    "cloud_secret_call_count",
    "network_call_count",
    "provider_call_count",
    "credential_handle_runtime_created_count",
    "credential_handle_created_count",
    "operator_approval_runtime_execution_count",
    "audit_store_runtime_created_count",
    "backend_health_runtime_created_count",
    "no_secret_leakage_smoke_runtime_created_count",
    "database_connection_count",
    "driver_open_count",
    "sql_execution_count",
    "schema_marker_read_count",
    "schema_marker_write_count",
    "repository_mode_enablement_count",
    "production_api_call_count",
}
EXPECTED_REQUIRED_CHECKS = {
    "run cloud secret service selection readiness checker",
    "run production resolver runtime blocker consolidation checker",
    "run resolver backend profile selection readiness checker",
    "run credential handle runtime implementation entry refresh checker",
    "run operator approval runtime implementation entry refresh checker",
    "run audit store runtime implementation entry refresh v3 checker",
    "run resolver backend health runtime implementation entry review checker",
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
        fixture.get("kind") == "production_ops_secret_backend_cloud_secret_service_selection_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-cloud-secret-service-selection-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "cloud_secret_service_selection_readiness_defined", "unexpected status")
    require(
        slice_info.get("selection_decision") == "cloud_secret_service_selection_deferred_until_runtime_prerequisites",
        "unexpected selection decision",
    )
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-cloud-secret-service-selection-readiness-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-cloud-secret-service-selection-readiness-v1.md",
    }.items():
        value = str(slice_info.get(key) or "")
        require(value == expected_path, f"unexpected {key}")
        require((REPO_ROOT / value).exists(), f"{key} path missing: {value}")
    for claim in {
        "cloud_secret_service_ready",
        "cloud_secret_service_selected",
        "cloud_secret_sdk_bound",
        "cloud_secret_client_created",
        "cloud_secret_service_call_executed",
        "production_resolver_runtime_task_card_created",
        "production_resolver_runtime_created",
        "credential_handle_runtime_ready",
        "operator_approval_runtime_ready",
        "audit_store_runtime_ready",
        "backend_health_runtime_ready",
        "repository_mode_ready",
        "production_api_ready",
    }:
        require(claim in set(slice_info.get("does_not_claim") or []), f"missing forbidden claim: {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        require(item.get("evidence") == relative_path, f"{dependency_id} evidence path drifted")
        dependency = load_json(REPO_ROOT / relative_path)
        if dependency_id == "secret-ref-schema-and-fixtures":
            require(dependency.get("scope") == "secret_reference_only", "secret ref fixture must remain reference-only")
            continue
        dependency_slice = dependency.get("slice") or {}
        require(dependency_slice.get("id") == dependency_id, f"{dependency_id} fixture id drifted")
        require(dependency_slice.get("status") == expected_status, f"{dependency_id} fixture status drifted")


def assert_selection_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("selection_boundary") or {}
    for field, expected in {
        "status": "cloud_secret_service_selection_readiness_defined",
        "selection_decision": "cloud_secret_service_selection_deferred_until_runtime_prerequisites",
        "concrete_cloud_vendor_selection_status": "not_selected",
        "allowed_backend_kind_scope": "reserved_only",
        "candidate_profile_manifest_status": "metadata_only_static_reference",
        "provider_sdk_binding_status": "forbidden",
        "cloud_secret_client_status": "not_created",
        "cloud_secret_service_call_status": "not_executed",
        "production_resolver_runtime_task_card_status": "not_created",
        "production_resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
        "credential_handle_runtime_status": "blocked_before_runtime_task_card",
        "operator_approval_runtime_status": "blocked_before_runtime_task_card",
        "audit_store_runtime_status": "blocked_before_runtime_task_card",
        "backend_health_runtime_status": "blocked_before_runtime_task_card",
        "no_secret_leakage_smoke_runtime_status": "blocked_before_runtime_task_card",
        "repository_mode_status": "disabled",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "runtime_task_card_allowed_now",
        "cloud_vendor_selected_in_this_slice",
        "provider_sdk_bound_in_this_slice",
        "cloud_secret_client_created_in_this_slice",
        "cloud_secret_service_called_in_this_slice",
        "production_resolver_runtime_created_in_this_slice",
        "credential_handle_runtime_created_in_this_slice",
        "operator_approval_runtime_created_in_this_slice",
        "audit_store_runtime_created_in_this_slice",
        "backend_health_runtime_created_in_this_slice",
        "no_secret_leakage_smoke_runtime_created_in_this_slice",
        "database_connection_provider_enabled",
        "repository_mode_enabled",
        "production_api_enabled",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_candidate_shape_and_matrix(fixture: dict[str, Any]) -> None:
    shape = fixture.get("candidate_shape") or {}
    required_fields = set(shape.get("required_fields") or [])
    for field in {
        "candidate_kind",
        "selection_policy_version",
        "environment",
        "provider_profile_id",
        "secret_ref_namespace",
        "allowed_secret_ref_kinds",
        "operator_approval_dependency",
        "credential_handle_dependency",
        "audit_store_dependency",
        "backend_health_dependency",
        "no_leakage_smoke_dependency",
        "rotation_policy_ref",
    }:
        require(field in required_fields, f"candidate shape missing field: {field}")
    require(
        set(shape.get("allowed_candidate_kinds") or [])
        == {"external_secret_manager_reserved", "operator_managed_secret_store_reserved"},
        "allowed candidate kinds drifted",
    )
    require("aws_secrets_manager" in set(shape.get("forbidden_candidate_kinds") or []), "AWS concrete kind must be forbidden")
    require("hashicorp_vault" in set(shape.get("forbidden_candidate_kinds") or []), "Vault concrete kind must be forbidden")
    for field in (
        "environment_cross_fallback_allowed",
        "provider_profile_cross_fallback_allowed",
        "secret_ref_value_allowed",
        "provider_sdk_allowed_in_this_slice",
        "cloud_secret_service_call_allowed_in_this_slice",
    ):
        require(shape.get(field) is False, f"{field} must remain false")

    gates = rows_by_id(fixture, "selection_matrix", "gate")
    require(set(gates) == set(EXPECTED_SELECTION_GATES), "selection gate ids drifted")
    for gate, expected_status in EXPECTED_SELECTION_GATES.items():
        item = gates[gate]
        require(item.get("status") == expected_status, f"{gate} status drifted")
        require(item.get("source"), f"{gate} source missing")
        if gate != "repository_mode_runtime":
            require(item.get("blocks_cloud_service_selection") is True, f"{gate} must block selection")
            require(item.get("blocks_production_resolver_task_card") is True, f"{gate} must block resolver task card")
    require(
        gates["repository_mode_runtime"].get("blocks_cloud_service_selection") is False,
        "repository mode must remain durable-store blocker only",
    )

    requirements = set(fixture.get("future_task_card_requirements") or [])
    for requirement in {
        "concrete_cloud_vendor_choice_must_be_separate_task",
        "provider_sdk_binding_must_be_separate_runtime_task",
        "cloud_secret_client_creation_must_be_separate_runtime_task",
        "credential_handle_runtime_entry_must_no_longer_be_blocked",
        "operator_approval_runtime_entry_must_no_longer_be_blocked",
        "audit_store_runtime_entry_must_no_longer_be_blocked",
        "backend_health_runtime_entry_must_no_longer_be_blocked",
        "no_secret_leakage_smoke_runtime_entry_must_no_longer_be_blocked",
        "production_resolver_runtime_task_card_must_not_enable_repository_mode",
        "diagnostics_must_remain_sanitized_reference_only",
    }:
        require(requirement in requirements, f"missing future task requirement: {requirement}")


def assert_alignment_with_existing_evidence() -> None:
    consolidation = load_json(BLOCKER_CONSOLIDATION_PATH)
    blockers = rows_by_id(consolidation, "blocker_matrix", "blocker_id")
    require(
        blockers["cloud_secret_service_selection"].get("status") == "not_selected",
        "blocker consolidation cloud selection status drifted",
    )
    require(
        blockers["cloud_secret_service_selection"].get("blocks_production_resolver_task_card") is True,
        "cloud selection must block resolver task card",
    )
    boundary = consolidation.get("consolidation_boundary") or {}
    require(boundary.get("cloud_secret_service_enabled") is False, "consolidation must not enable cloud secret service")
    require(boundary.get("production_resolver_runtime_created_in_this_slice") is False, "consolidation must not create resolver")

    backend_profile = load_json(BACKEND_PROFILE_SELECTION_PATH)
    selection = backend_profile.get("selection_boundary") or {}
    require(selection.get("cloud_secret_service_status") == "not_selected", "backend profile cloud status drifted")
    require(selection.get("cloud_secret_service_enabled") is False, "backend profile must not enable cloud service")
    shape = backend_profile.get("backend_profile_shape") or {}
    require(
        set(shape.get("allowed_backend_kinds") or [])
        == {"external_secret_manager_reserved", "operator_managed_secret_store_reserved"},
        "backend profile allowed kind drifted",
    )
    require(shape.get("secret_ref_value_allowed") is False, "backend profile must forbid secret ref value")

    credential_refresh = load_json(CREDENTIAL_REFRESH_PATH)
    credential_boundary = credential_refresh.get("refresh_boundary") or {}
    require(
        credential_boundary.get("entry_decision") == "credential_handle_runtime_task_card_still_blocked_after_refresh",
        "credential refresh entry decision drifted",
    )
    require(
        credential_boundary.get("cloud_secret_service_status") == "not_selected",
        "credential refresh must keep cloud service not selected",
    )

    operator_refresh = load_json(OPERATOR_REFRESH_PATH)
    operator_boundary = operator_refresh.get("refresh_boundary") or {}
    require(
        operator_boundary.get("entry_decision") == "operator_approval_runtime_task_card_still_blocked_after_refresh",
        "operator refresh entry decision drifted",
    )
    require(
        operator_boundary.get("cloud_secret_service_status") == "not_selected",
        "operator refresh must keep cloud service not selected",
    )

    for path, boundary_key, expected in [
        (AUDIT_REFRESH_PATH, "entry_refresh_boundary", "blocked_before_runtime_task_card"),
        (BACKEND_HEALTH_ENTRY_PATH, "entry_boundary", "blocked_before_runtime_task_card"),
        (NO_LEAKAGE_ENTRY_PATH, "entry_boundary", "blocked_before_runtime_task_card"),
    ]:
        document = load_json(path)
        boundary_section = document.get(boundary_key) or {}
        status_values = set(str(value) for value in boundary_section.values())
        require(expected in status_values, f"{path.name} must retain blocked-before-task-card status")

    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in {
        "cloud_vendor_specific_backend": "not_selected",
        "cloud_secret_service_selection_readiness_status": "defined_without_cloud_backend_selection",
        "production_resolver_runtime_task_card_status": "not_created",
        "resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
        "credential_handle_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "operator_approval_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "audit_store_runtime_implementation_entry_refresh_v3_status": "blocked_before_runtime_task_card",
        "resolver_backend_health_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "real_resolver_no_secret_leakage_smoke_runtime_status": "not_created",
    }.items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    secret_reference = load_json(SECRET_REFERENCE_PATH)
    require(secret_reference.get("scope") == "secret_reference_only", "secret reference must remain reference-only")


def assert_failure_diagnostics_and_policies(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} must define boundary")
        require(item.get("sanitized_diagnostic"), f"{code} must define diagnostic")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    require(EXPECTED_DIAGNOSTIC_FIELDS.issubset(set(diagnostics.get("allowed_fields") or [])), "allowed diagnostics drifted")
    require(EXPECTED_FORBIDDEN_DIAGNOSTICS.issubset(set(diagnostics.get("forbidden_fields") or [])), "forbidden diagnostics drifted")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must be false")
    require(diagnostics.get("secret_ref_value_allowed_in_diagnostics") is False, "secret ref value must not be emitted")

    fallback = set(fixture.get("no_fallback_policy") or [])
    for rule in (
        "no fallback from cloud secret service selection to fake resolver",
        "no fallback from production backend profile to test backend profile",
        "no fallback from missing provider profile binding to environment credential",
        "no vendor-specific backend without separate selection task",
        "no runtime task card if credential handle runtime is blocked",
        "no runtime task card if operator approval runtime is blocked",
        "no runtime task card if audit store runtime is blocked",
        "no runtime task card if backend health runtime is blocked",
        "no runtime task card if no leakage smoke runtime is blocked",
        "cloud secret service selection readiness does not mean workflow repository mode ready",
    ):
        require(rule in fallback, f"missing no fallback rule: {rule}")
    side_effect_policy = set(fixture.get("no_side_effect_policy") or [])
    for rule in (
        "no cloud secret service call",
        "no cloud secret SDK import",
        "no provider SDK binding",
        "no network call",
        "no environment secret read",
        "no production resolver call",
        "no database connection",
        "no SQL execution",
        "no repository mode enablement",
    ):
        require(rule in side_effect_policy, f"missing no side effect rule: {rule}")
    counters = fixture.get("side_effect_counters") or {}
    require(set(counters) == EXPECTED_ZERO_COUNTERS, "side effect counters drifted")
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
    prior_call = 'run_python_script("check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.py", [])'
    current_call = 'run_python_script("check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py", [])'
    next_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call in (prior_call, current_call, next_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(prior_call) < check_repo.index(current_call) < check_repo.index(next_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-cloud-secret-service-selection-readiness-v1.md",
        "docs/task-cards/production-secret-backend-cloud-secret-service-selection-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"cloud secret service selection artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "cloud selection artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "cloud selection artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_selection_boundary(fixture)
    assert_candidate_shape_and_matrix(fixture)
    assert_alignment_with_existing_evidence()
    assert_failure_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend cloud secret service selection readiness checks passed.")


if __name__ == "__main__":
    main()
