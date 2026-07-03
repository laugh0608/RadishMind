#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-production-resolver-runtime-blocker-consolidation-v1": (
        "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-blocker-consolidation-v1.json",
        "production_resolver_runtime_blocker_consolidation_defined",
    ),
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ),
    "production-secret-backend-audit-store-runtime-blocker-matrix-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-blocker-matrix-v1.json",
        "audit_store_runtime_blocker_matrix_defined",
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.json",
        "credential_handle_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.json",
        "operator_approval_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-cloud-secret-service-selection-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json",
        "cloud_secret_service_selection_readiness_defined",
    ),
    "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1.json",
        "resolver_backend_health_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1.json"
        ),
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined",
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

EXPECTED_BOUNDARY = {
    "status": "production_resolver_runtime_implementation_entry_refresh_v2_defined",
    "entry_decision": "production_resolver_runtime_task_card_still_blocked_after_refresh_v2",
    "previous_consolidation_status": "production_resolver_runtime_blocker_consolidation_defined",
    "audit_store_entry_refresh_v4_status": "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
    "credential_handle_runtime_entry_refresh_status": "credential_handle_runtime_task_card_still_blocked_after_refresh",
    "operator_approval_runtime_entry_refresh_status": "operator_approval_runtime_task_card_still_blocked_after_refresh",
    "backend_health_runtime_entry_refresh_status": "resolver_backend_health_runtime_task_card_still_blocked_after_refresh",
    "no_secret_leakage_smoke_runtime_entry_refresh_status": (
        "real_resolver_no_secret_leakage_smoke_runtime_task_card_still_blocked_after_refresh"
    ),
    "cloud_secret_service_selection_status": "cloud_secret_service_selection_deferred_until_runtime_prerequisites",
    "production_resolver_runtime_task_card_status": "not_created",
    "production_resolver_runtime_status": "not_created",
    "production_secret_backend_status": "not_satisfied",
    "cloud_secret_service_status": "not_selected",
    "cloud_secret_client_status": "not_created",
    "durable_audit_backend_status": "not_selected",
    "audit_writer_runtime_status": "not_created",
    "runtime_event_schema_artifact_status": "implemented_static_schema_artifact",
    "delivery_runtime_status": "not_created",
    "idempotency_runtime_status": "not_created",
    "credential_handle_runtime_status": "not_created",
    "operator_approval_runtime_status": "not_created",
    "backend_health_runtime_status": "not_created",
    "no_secret_leakage_smoke_runtime_status": "not_created",
    "database_secret_resolver_runtime_status": "blocked_before_implementation_task_card",
    "negative_auth_smoke_runtime_status": "not_created_readiness_only",
    "repository_mode_status": "disabled",
    "production_api_status": "not_created",
}

EXPECTED_FALSE_FLAGS = {
    "runtime_task_card_allowed_now",
    "runtime_task_card_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "production_resolver_backend_client_created_in_this_slice",
    "cloud_secret_client_created_in_this_slice",
    "cloud_secret_service_enabled",
    "credential_payload_created_in_this_slice",
    "credential_handle_runtime_created_in_this_slice",
    "operator_approval_runtime_created_in_this_slice",
    "operator_approval_runtime_executed_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "backend_health_runtime_created_in_this_slice",
    "backend_health_check_executed_in_this_slice",
    "no_secret_leakage_smoke_runtime_created_in_this_slice",
    "no_secret_leakage_smoke_executed_in_this_slice",
    "database_secret_resolver_runtime_created_in_this_slice",
    "database_connection_provider_enabled",
    "negative_auth_smoke_runtime_created_in_this_slice",
    "repository_mode_enabled",
    "production_api_enabled",
}

EXPECTED_GATES = {
    "previous_blocker_consolidation": "production_resolver_runtime_task_card_still_blocked_after_consolidation",
    "audit_store_runtime": "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
    "durable_audit_backend": "not_selected",
    "audit_writer_runtime": "not_created",
    "runtime_event_schema_artifact": "implemented_static_schema_artifact",
    "delivery_runtime": "not_created",
    "idempotency_runtime": "not_created",
    "credential_handle_runtime": "credential_handle_runtime_task_card_still_blocked_after_refresh",
    "operator_approval_runtime": "operator_approval_runtime_task_card_still_blocked_after_refresh",
    "backend_health_runtime": "resolver_backend_health_runtime_task_card_still_blocked_after_refresh",
    "no_secret_leakage_smoke_runtime": (
        "real_resolver_no_secret_leakage_smoke_runtime_task_card_still_blocked_after_refresh"
    ),
    "cloud_secret_service_selection": "not_selected",
    "production_resolver_runtime_task_card": "not_created",
    "production_resolver_runtime": "not_created",
    "database_secret_resolver_runtime": "database_secret_resolver_runtime_still_blocked_before_implementation_task_card",
    "negative_auth_smoke_runtime": "negative_auth_smoke_runtime_readiness_defined_before_runtime_artifact",
    "repository_mode_runtime": "disabled",
}

EXPECTED_REQUIREMENTS = {
    "audit_store_runtime_entry_must_no_longer_be_blocked",
    "durable_audit_backend_must_be_selected_by_separate_task",
    "audit_writer_runtime_must_exist",
    "runtime_event_schema_artifact_must_remain_metadata_only_until_writer_delivery_idempotency_exist",
    "delivery_runtime_must_exist",
    "idempotency_runtime_must_exist",
    "credential_handle_runtime_entry_must_no_longer_be_blocked",
    "operator_approval_runtime_entry_must_no_longer_be_blocked",
    "backend_health_runtime_entry_must_no_longer_be_blocked",
    "no_secret_leakage_smoke_runtime_entry_must_no_longer_be_blocked",
    "cloud_secret_service_must_be_selected_by_separate_task",
    "database_secret_resolver_runtime_must_not_be_unlocked_by_production_refresh",
    "repository_mode_must_remain_separate_runtime_task",
    "diagnostics_must_remain_sanitized_reference_only",
}

EXPECTED_FAILURE_CODES = {
    "production_resolver_runtime_refresh_v2_dependency_missing",
    "production_resolver_runtime_refresh_v2_audit_store_blocked",
    "production_resolver_runtime_refresh_v2_credential_handle_blocked",
    "production_resolver_runtime_refresh_v2_operator_approval_blocked",
    "production_resolver_runtime_refresh_v2_backend_health_blocked",
    "production_resolver_runtime_refresh_v2_no_leakage_blocked",
    "production_resolver_runtime_refresh_v2_cloud_service_not_selected",
    "production_resolver_runtime_refresh_v2_database_secret_resolver_blocked",
    "production_resolver_runtime_refresh_v2_negative_auth_smoke_missing",
    "production_resolver_runtime_refresh_v2_task_card_forbidden",
    "production_resolver_runtime_refresh_v2_runtime_created_forbidden",
    "production_resolver_runtime_refresh_v2_secret_material_detected",
    "production_resolver_runtime_refresh_v2_repository_mode_forbidden",
    "production_resolver_runtime_refresh_v2_scope_overreach",
}

EXPECTED_DIAGNOSTICS = {
    "production_resolver_runtime_entry_refresh_v2_status",
    "entry_decision",
    "gate",
    "gate_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}

EXPECTED_ZERO_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "production_resolver_task_card_created_count",
    "production_resolver_runtime_created_count",
    "production_resolver_backend_client_created_count",
    "cloud_secret_client_created_count",
    "credential_payload_created_count",
    "credential_handle_created_count",
    "credential_handle_runtime_created_count",
    "operator_approval_runtime_execution_count",
    "audit_store_runtime_created_count",
    "audit_writer_runtime_created_count",
    "audit_event_write_count",
    "delivery_execution_count",
    "idempotency_decision_count",
    "duplicate_detection_count",
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
    "run production resolver runtime implementation entry refresh v2 checker",
    "run audit store runtime blocker matrix checker",
    "run audit store runtime implementation entry refresh v4 checker",
    "run production resolver runtime blocker consolidation checker",
    "run credential handle runtime implementation entry refresh checker",
    "run operator approval runtime implementation entry refresh checker",
    "run resolver backend health runtime implementation entry refresh checker",
    "run real resolver no leakage smoke runtime implementation entry refresh checker",
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


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    if slice_info.get("status"):
        return str(slice_info["status"])
    if document.get("status"):
        return str(document["status"])
    if document.get("scope") == "secret_reference_only":
        return "satisfied_reference_only_resolver_disabled"
    return ""


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_production_resolver_runtime_implementation_entry_refresh_v2",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "production_resolver_runtime_implementation_entry_refresh_v2_defined",
        "unexpected status",
    )
    require(
        slice_info.get("entry_decision") == "production_resolver_runtime_task_card_still_blocked_after_refresh_v2",
        "unexpected entry decision",
    )
    for key, expected_path in {
        "task_card": (
            "docs/task-cards/"
            "production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2-plan.md"
        ),
        "platform_topic": (
            "docs/platform/"
            "production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.md"
        ),
    }.items():
        value = str(slice_info.get(key) or "")
        require(value == expected_path, f"unexpected {key}")
        require((REPO_ROOT / value).exists(), f"{key} path missing: {value}")
    for claim in {
        "production_resolver_runtime_task_card_created",
        "production_resolver_runtime_created",
        "cloud_secret_client_created",
        "credential_payload_created",
        "credential_handle_runtime_ready",
        "operator_approval_runtime_ready",
        "audit_store_runtime_ready",
        "backend_health_runtime_ready",
        "no_secret_leakage_smoke_runtime_created",
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
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"{dependency_id} source status drifted")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("refresh_boundary") or {}
    for field, expected in EXPECTED_BOUNDARY.items():
        require(boundary.get(field) == expected, f"refresh_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"refresh_boundary.{field} must remain false")


def assert_gates_and_requirements(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture, "gate_matrix", "gate")
    require(set(gates) == set(EXPECTED_GATES), "gate ids drifted")
    for gate, expected_status in EXPECTED_GATES.items():
        item = gates[gate]
        require(item.get("status") == expected_status, f"{gate} status drifted")
        require(item.get("source"), f"{gate} source missing")
        require(item.get("blocks_workflow_durable_store") is True, f"{gate} must block durable store")
    for gate in (
        "previous_blocker_consolidation",
        "audit_store_runtime",
        "durable_audit_backend",
        "audit_writer_runtime",
        "runtime_event_schema_artifact",
        "delivery_runtime",
        "idempotency_runtime",
        "credential_handle_runtime",
        "operator_approval_runtime",
        "backend_health_runtime",
        "no_secret_leakage_smoke_runtime",
        "cloud_secret_service_selection",
        "production_resolver_runtime_task_card",
        "production_resolver_runtime",
    ):
        require(gates[gate].get("blocks_production_resolver_task_card") is True, f"{gate} must block resolver")
    for gate in ("database_secret_resolver_runtime", "negative_auth_smoke_runtime", "repository_mode_runtime"):
        require(gates[gate].get("blocks_production_resolver_task_card") is False, f"{gate} must stay workflow-only")

    requirements = set(fixture.get("future_task_card_requirements") or [])
    missing = sorted(EXPECTED_REQUIREMENTS - requirements)
    require(not missing, f"missing future task card requirements: {missing}")


def assert_prior_evidence_alignment() -> None:
    consolidation = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-production-resolver-runtime-blocker-consolidation-v1"
    ][0])
    consolidation_boundary = consolidation.get("consolidation_boundary") or {}
    require(
        consolidation_boundary.get("entry_decision")
        == "production_resolver_runtime_task_card_still_blocked_after_consolidation",
        "consolidation decision drifted",
    )
    require(
        consolidation_boundary.get("production_resolver_runtime_task_card_status") == "not_created",
        "consolidation must not create resolver task card",
    )

    audit = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4"
    ][0])
    audit_boundary = audit.get("entry_refresh_boundary") or {}
    for field, expected in {
        "entry_decision": "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
        "audit_store_runtime_task_card_status": "not_created",
        "durable_audit_backend_status": "not_selected",
        "audit_writer_runtime_status": "not_created",
        "runtime_event_schema_artifact_status": "not_created",
        "delivery_runtime_status": "not_created",
        "idempotency_runtime_status": "not_created",
        "production_resolver_runtime_status": "not_created",
        "cloud_secret_service_status": "not_selected",
        "repository_mode_status": "disabled",
    }.items():
        require(audit_boundary.get(field) == expected, f"audit v4 {field} drifted")
    require(audit_boundary.get("runtime_task_card_allowed_now") is False, "audit v4 opened task card")

    matrix = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-audit-store-runtime-blocker-matrix-v1"
    ][0])
    matrix_boundary = matrix.get("matrix_boundary") or {}
    require(
        matrix_boundary.get("entry_decision") == "audit_store_runtime_task_card_still_blocked_after_schema_artifact",
        "audit store blocker matrix decision drifted",
    )
    require(
        matrix_boundary.get("schema_artifact_status") == "implemented_static_schema_artifact",
        "audit store blocker matrix schema artifact status drifted",
    )
    require(
        matrix_boundary.get("audit_store_runtime_task_card_status") == "not_created",
        "audit store blocker matrix opened runtime task card",
    )

    refresh_expectations = {
        "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1": (
            "credential_handle_runtime_task_card_still_blocked_after_refresh"
        ),
        "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1": (
            "operator_approval_runtime_task_card_still_blocked_after_refresh"
        ),
        "production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh-v1": (
            "resolver_backend_health_runtime_task_card_still_blocked_after_refresh"
        ),
        "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-refresh-v1": (
            "real_resolver_no_secret_leakage_smoke_runtime_task_card_still_blocked_after_refresh"
        ),
    }
    for dependency_id, expected_decision in refresh_expectations.items():
        document = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[dependency_id][0])
        boundary = document.get("refresh_boundary") or {}
        require(boundary.get("entry_decision") == expected_decision, f"{dependency_id} decision drifted")
        require(boundary.get("runtime_task_card_allowed_now") is False, f"{dependency_id} opened task card")
        require(boundary.get("production_resolver_runtime_status") == "not_created", f"{dependency_id} created resolver")

    cloud = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "production-secret-backend-cloud-secret-service-selection-readiness-v1"
    ][0])
    cloud_boundary = cloud.get("selection_boundary") or {}
    require(
        cloud_boundary.get("selection_decision") == "cloud_secret_service_selection_deferred_until_runtime_prerequisites",
        "cloud selection decision drifted",
    )
    require(cloud_boundary.get("concrete_cloud_vendor_selection_status") == "not_selected", "cloud vendor selected")
    require(cloud_boundary.get("cloud_secret_client_status") == "not_created", "cloud client created")

    database = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1"
    ][0])
    database_boundary = database.get("dependency_refresh_boundary") or {}
    require(
        database_boundary.get("decision")
        == "database_secret_resolver_runtime_still_blocked_before_implementation_task_card",
        "database secret resolver decision drifted",
    )
    require(
        database_boundary.get("production_resolver_runtime_task_card_created_in_this_slice") is False,
        "database dependency refresh created resolver task card",
    )
    require(database_boundary.get("repository_mode_enabled_in_this_slice") is False, "repository mode enabled")

    negative = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES[
        "workflow-saved-draft-negative-auth-smoke-runtime-readiness-v1"
    ][0])
    negative_boundary = negative.get("readiness_boundary") or {}
    require(
        negative_boundary.get("decision") == "negative_auth_smoke_runtime_readiness_defined_before_runtime_artifact",
        "negative auth decision drifted",
    )
    require(negative_boundary.get("runtime_smoke_artifact_allowed_now") is False, "negative auth runtime allowed")

    readiness = load_json(REPO_ROOT / EXPECTED_DEPENDENCIES["production-secret-backend-implementation-readiness"][0])
    target = readiness.get("implementation_target") or {}
    for field, expected in {
        "production_secret_backend_status": "not_satisfied",
        "production_resolver_runtime_task_card_status": "not_created",
        "production_resolver_runtime_status": "not_created",
        "audit_store_runtime_implementation_entry_refresh_v4_status": "blocked_before_runtime_task_card",
        "audit_store_runtime_blocker_matrix_status": "audit_store_runtime_blocker_matrix_defined",
        "audit_store_runtime_task_card_status": "not_created",
        "audit_store_runtime_status": "not_created",
        "credential_handle_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "operator_approval_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "resolver_backend_health_runtime_implementation_entry_refresh_status": "blocked_before_runtime_task_card",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_status": (
            "blocked_before_runtime_task_card"
        ),
        "cloud_secret_service_status": "not_selected",
    }.items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")


def assert_failure_diagnostics_and_policies(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in failures.items():
        require(item.get("failure_boundary"), f"{code} must define boundary")
        require(item.get("sanitized_diagnostic"), f"{code} must define diagnostic")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    require(EXPECTED_DIAGNOSTICS.issubset(set(diagnostics.get("allowed_fields") or [])), "diagnostics allowlist drifted")
    sample = diagnostics.get("sample") or {}
    require(set(sample) <= set(diagnostics.get("allowed_fields") or []), "diagnostic sample contains forbidden field")
    require(
        sample.get("entry_decision") == "production_resolver_runtime_task_card_still_blocked_after_refresh_v2",
        "sample decision drifted",
    )
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must be false")
    require(diagnostics.get("secret_ref_value_allowed_in_diagnostics") is False, "secret ref value must not emit")

    fallback = set(fixture.get("no_fallback_policy") or [])
    for rule in {
        "no fallback from production resolver to fake resolver",
        "no fallback to developer env plaintext",
        "no runtime task card if audit store runtime is blocked",
        "no runtime task card if credential handle runtime is blocked",
        "no runtime task card if operator approval runtime is blocked",
        "no runtime task card if backend health runtime is blocked",
        "no runtime task card if no leakage smoke runtime is blocked",
        "production resolver refresh does not mean workflow repository mode ready",
    }:
        require(rule in fallback, f"missing no fallback rule: {rule}")

    side_effect_policy = set(fixture.get("no_side_effect_policy") or [])
    for rule in {
        "no cloud secret service call",
        "no production resolver call",
        "no approval runtime execution",
        "no audit event write",
        "no delivery execution",
        "no database connection",
        "no SQL execution",
        "no repository mode enablement",
    }:
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
    prior_call = 'run_python_script("check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py", [])'
    current_call = (
        'run_python_script("check-production-ops-secret-backend-'
        'production-resolver-runtime-implementation-entry-refresh-v2.py", [])'
    )
    next_call = 'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])'
    for call in (prior_call, current_call, next_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(prior_call) < check_repo.index(current_call) < check_repo.index(next_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.md",
        "docs/task-cards/production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2-plan.md",
        "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"v2 artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "v2 artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "v2 artifacts contain dsn-like credential")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_boundary(fixture)
    assert_gates_and_requirements(fixture)
    assert_prior_evidence_alignment()
    assert_failure_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_docs_validation_and_check_repo(fixture)
    assert_no_secret_literals()
    print("production ops secret backend production resolver runtime implementation entry refresh v2 checks passed.")


if __name__ == "__main__":
    main()
