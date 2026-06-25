#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.json",
        "real_resolver_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.json",
        "audit_store_runtime_implementation_entry_refresh_v3_defined",
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.json",
        "credential_handle_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.json",
        "operator_approval_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1": (
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
    ),
    "production-secret-backend-resolver-backend-profile-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
        "resolver_backend_profile_selection_readiness_defined",
    ),
    "workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.json",
        "draft_database_secret_resolver_runtime_dependency_refresh_defined",
    ),
    "workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1.json",
        "draft_negative_auth_smoke_runtime_readiness_defined",
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
EXPECTED_BLOCKERS = {
    "production_resolver_runtime_task_card": "not_created",
    "production_resolver_runtime": "not_created",
    "credential_handle_runtime": "blocked_before_runtime_task_card",
    "operator_approval_runtime": "blocked_before_runtime_task_card",
    "audit_store_runtime": "blocked_before_runtime_task_card",
    "backend_health_runtime": "blocked_before_runtime_task_card",
    "no_secret_leakage_smoke_runtime": "blocked_before_runtime_task_card",
    "cloud_secret_service_selection": "not_selected",
    "database_secret_resolver_runtime": "blocked_before_implementation_task_card",
    "negative_auth_smoke_runtime": "not_created_readiness_only",
    "repository_mode_runtime": "disabled",
}
EXPECTED_FAILURE_CODES = {
    "production_resolver_blocker_consolidation_dependency_missing",
    "production_resolver_blocker_consolidation_runtime_task_card_forbidden",
    "production_resolver_blocker_consolidation_runtime_created_forbidden",
    "production_resolver_blocker_consolidation_secret_material_detected",
    "production_resolver_blocker_consolidation_repository_mode_forbidden",
    "production_resolver_blocker_consolidation_scope_overreach",
}
EXPECTED_DIAGNOSTIC_FIELDS = {
    "blocker_id",
    "blocker_status",
    "entry_decision",
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
    "raw_request_payload",
    "raw_response_payload",
}
EXPECTED_ZERO_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "production_resolver_task_card_created_count",
    "production_resolver_runtime_created_count",
    "credential_payload_created_count",
    "credential_handle_created_count",
    "credential_handle_runtime_created_count",
    "operator_approval_runtime_execution_count",
    "audit_store_runtime_created_count",
    "audit_writer_created_count",
    "audit_event_write_count",
    "backend_health_runtime_created_count",
    "backend_health_check_count",
    "no_secret_leakage_smoke_runtime_created_count",
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
EXPECTED_REQUIRED_CHECKS = {
    "run production resolver runtime blocker consolidation checker",
    "run real resolver runtime implementation entry refresh checker",
    "run audit store runtime implementation entry refresh v3 checker",
    "run credential handle runtime implementation entry review checker",
    "run operator approval runtime implementation entry review checker",
    "run resolver backend health runtime implementation entry review checker",
    "run real resolver no leakage smoke runtime implementation entry review checker",
    "run workflow database secret resolver runtime dependency refresh checker",
    "run workflow negative auth smoke runtime readiness checker",
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
        fixture.get("kind") == "production_ops_secret_backend_production_resolver_runtime_blocker_consolidation_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-production-resolver-runtime-blocker-consolidation-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "production_resolver_runtime_blocker_consolidation_defined", "unexpected status")
    require(
        slice_info.get("entry_decision") == "production_resolver_runtime_task_card_still_blocked_after_consolidation",
        "unexpected entry decision",
    )
    for key, expected_path in {
        "task_card": "docs/task-cards/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1-plan.md",
        "platform_topic": "docs/platform/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.md",
    }.items():
        value = str(slice_info.get(key) or "")
        require(value == expected_path, f"unexpected {key}")
        require((REPO_ROOT / value).exists(), f"{key} path missing: {value}")
    for claim in {
        "production_resolver_runtime_task_card_created",
        "production_resolver_runtime_created",
        "credential_handle_runtime_ready",
        "operator_approval_runtime_ready",
        "audit_store_runtime_ready",
        "backend_health_runtime_ready",
        "repository_mode_ready",
        "negative_auth_smoke_runtime_created",
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


def assert_consolidation_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("consolidation_boundary") or {}
    for field, expected in {
        "status": "production_resolver_runtime_blocker_consolidation_defined",
        "entry_decision": "production_resolver_runtime_task_card_still_blocked_after_consolidation",
        "real_resolver_entry_refresh_status": "blocked_before_runtime_task_card",
        "database_secret_resolver_dependency_status": "blocked_before_implementation_task_card",
        "negative_auth_smoke_runtime_status": "not_created_readiness_only",
        "production_resolver_runtime_task_card_status": "not_created",
        "production_resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
        "repository_mode_status": "disabled",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "runtime_task_card_allowed_now",
        "runtime_task_card_created_in_this_slice",
        "production_resolver_runtime_created_in_this_slice",
        "cloud_secret_service_enabled",
        "credential_payload_created_in_this_slice",
        "credential_handle_runtime_created_in_this_slice",
        "operator_approval_runtime_created_in_this_slice",
        "audit_store_runtime_created_in_this_slice",
        "backend_health_runtime_created_in_this_slice",
        "no_secret_leakage_smoke_runtime_created_in_this_slice",
        "database_connection_provider_enabled",
        "negative_auth_smoke_runtime_created_in_this_slice",
        "repository_mode_enabled",
        "production_api_enabled",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_blockers_and_requirements(fixture: dict[str, Any]) -> None:
    blockers = rows_by_id(fixture, "blocker_matrix", "blocker_id")
    require(set(blockers) == set(EXPECTED_BLOCKERS), "blocker ids drifted")
    for blocker_id, expected_status in EXPECTED_BLOCKERS.items():
        item = blockers[blocker_id]
        require(item.get("status") == expected_status, f"{blocker_id} status drifted")
        require(item.get("source"), f"{blocker_id} source missing")
        require(item.get("blocks_workflow_durable_store") is True, f"{blocker_id} must block durable store")
    for blocker_id in (
        "production_resolver_runtime_task_card",
        "production_resolver_runtime",
        "credential_handle_runtime",
        "operator_approval_runtime",
        "audit_store_runtime",
        "backend_health_runtime",
        "no_secret_leakage_smoke_runtime",
        "cloud_secret_service_selection",
    ):
        require(blockers[blocker_id].get("blocks_production_resolver_task_card") is True, f"{blocker_id} must block resolver task card")
    for blocker_id in ("database_secret_resolver_runtime", "negative_auth_smoke_runtime", "repository_mode_runtime"):
        require(blockers[blocker_id].get("blocks_production_resolver_task_card") is False, f"{blocker_id} must not be a resolver task-card blocker")

    requirements = set(fixture.get("future_task_card_requirements") or [])
    for requirement in {
        "credential_handle_runtime_entry_must_no_longer_be_blocked",
        "operator_approval_runtime_entry_must_no_longer_be_blocked",
        "audit_store_runtime_entry_must_no_longer_be_blocked",
        "backend_health_runtime_entry_must_no_longer_be_blocked",
        "no_secret_leakage_smoke_runtime_entry_must_no_longer_be_blocked",
        "production_resolver_runtime_task_card_must_not_enable_repository_mode",
        "workflow_durable_store_must_wait_for_database_secret_resolver_and_repository_mode_runtime",
    }:
        require(requirement in requirements, f"missing future task requirement: {requirement}")


def assert_alignment_with_existing_evidence() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in {
        "production_resolver_runtime_task_card_status": "not_created",
        "resolver_runtime_status": "not_created",
        "real_resolver_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "credential_handle_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "operator_approval_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "audit_store_runtime_implementation_entry_refresh_v3_status": "blocked_before_runtime_task_card",
        "resolver_backend_health_runtime_implementation_entry_review_status": "blocked_before_runtime_task_card",
        "real_resolver_no_secret_leakage_smoke_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
    }.items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    workflow = load_json(
        REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.json"
    )
    boundary = workflow.get("dependency_refresh_boundary") or {}
    for field in (
        "production_real_resolver_entry_refresh_consumed",
        "credential_handle_runtime_entry_review_consumed",
        "operator_approval_runtime_entry_review_consumed",
        "audit_store_runtime_entry_refresh_consumed",
        "backend_health_runtime_entry_review_consumed",
        "no_leakage_runtime_entry_review_consumed",
    ):
        require(boundary.get(field) is True, f"workflow dependency refresh {field} drifted")
    require(boundary.get("production_resolver_runtime_task_card_created_in_this_slice") is False, "workflow refresh must not create resolver task card")
    require(boundary.get("repository_mode_enabled_in_this_slice") is False, "workflow refresh must not enable repository mode")

    negative = load_json(REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1.json")
    neg_boundary = negative.get("readiness_boundary") or {}
    require(neg_boundary.get("runtime_smoke_artifact_allowed_now") is False, "negative auth runtime must remain absent")


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
        "no fallback from production resolver to fake resolver",
        "no fallback to developer env plaintext",
        "no runtime task card if credential handle runtime is blocked",
        "no runtime task card if operator approval runtime is blocked",
        "no runtime task card if audit store runtime is blocked",
        "no runtime task card if backend health runtime is blocked",
        "blocker consolidation does not mean workflow repository mode ready",
    ):
        require(rule in fallback, f"missing no fallback rule: {rule}")
    side_effect_policy = set(fixture.get("no_side_effect_policy") or [])
    for rule in ("no cloud secret service call", "no production resolver call", "no database connection", "no SQL execution", "no repository mode enablement"):
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
    prior_call = 'run_python_script("check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py", [])'
    current_call = 'run_python_script("check-production-ops-secret-backend-production-resolver-runtime-blocker-consolidation-v1.py", [])'
    next_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call in (prior_call, current_call, next_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(prior_call) < check_repo.index(current_call) < check_repo.index(next_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.md",
        "docs/task-cards/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"blocker consolidation artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "blocker consolidation artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "blocker consolidation artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_consolidation_boundary(fixture)
    assert_blockers_and_requirements(fixture)
    assert_alignment_with_existing_evidence()
    assert_failure_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend production resolver runtime blocker consolidation checks passed.")


if __name__ == "__main__":
    main()
