#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-selection-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
DURABLE_BOUNDARY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.json"
)
SCHEMA_ARTIFACT_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.json"
)
V4_REFRESH_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json"
)
WRITER_BOUNDARY_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json"
)
DELIVERY_READINESS_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-runtime-readiness-v1.json"
)
IDEMPOTENCY_READINESS_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.json"
)
PRODUCTION_RESOLVER_V2_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-durable-backend-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.json",
        "audit_store_durable_backend_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-runtime-event-schema-artifact-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.json",
        "audit_store_runtime_event_schema_artifact_implemented",
    ),
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ),
    "production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json",
        "audit_store_writer_runtime_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-delivery-runtime-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-delivery-runtime-readiness-v1.json",
        "audit_store_delivery_runtime_readiness_defined",
    ),
    "production-secret-backend-audit-store-idempotency-runtime-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.json",
        "audit_store_idempotency_runtime_readiness_defined",
    ),
    "production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.json",
        "credential_handle_runtime_implementation_entry_refresh_defined",
    ),
    "production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1": (
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.json",
        "operator_approval_runtime_implementation_entry_refresh_defined",
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
    "production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2": (
        "scripts/checks/fixtures/production-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.json",
        "production_resolver_runtime_implementation_entry_refresh_v2_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}

EXPECTED_SELECTION_GATES = {
    "candidate_backend_kind_allowlist": "reserved_only",
    "durable_backend_selection": "deferred_without_backend_selection",
    "storage_adapter_contract": "metadata_only_static_reference",
    "event_schema_artifact": "implemented_static_schema_artifact",
    "retention_policy": "required_reference_only",
    "redaction_profile": "required_reference_only",
    "audit_writer_runtime": "not_created",
    "idempotency_runtime": "not_created",
    "delivery_runtime": "not_created",
    "operator_approval_runtime": "not_created",
    "credential_handle_runtime": "not_created",
    "backend_health_runtime": "not_created",
    "no_secret_leakage_smoke_runtime": "not_created",
    "database_connection_provider": "blocked",
    "repository_mode_runtime": "disabled",
    "production_api": "not_created",
}

EXPECTED_FAILURE_CODES = {
    "audit_store_durable_backend_selection_dependency_missing",
    "audit_store_durable_backend_selection_backend_selected_forbidden",
    "audit_store_durable_backend_selection_adapter_runtime_forbidden",
    "audit_store_durable_backend_selection_database_provider_forbidden",
    "audit_store_durable_backend_selection_writer_runtime_blocked",
    "audit_store_durable_backend_selection_delivery_runtime_blocked",
    "audit_store_durable_backend_selection_idempotency_runtime_blocked",
    "audit_store_durable_backend_selection_operator_approval_blocked",
    "audit_store_durable_backend_selection_credential_handle_blocked",
    "audit_store_durable_backend_selection_backend_health_blocked",
    "audit_store_durable_backend_selection_no_leakage_blocked",
    "audit_store_durable_backend_selection_secret_material_detected",
    "audit_store_durable_backend_selection_repository_mode_forbidden",
    "audit_store_durable_backend_selection_scope_overreach",
}

EXPECTED_DIAGNOSTIC_FIELDS = {
    "audit_store_durable_backend_selection_status",
    "selection_decision",
    "candidate_backend_kind",
    "candidate_status",
    "gate",
    "gate_status",
    "environment",
    "storage_owner_ref",
    "storage_adapter_contract_status",
    "event_schema_ref",
    "retention_policy_ref",
    "redaction_profile_ref",
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
    "database_hostname",
    "database_error_detail",
    "cloud_credential",
    "credential_payload",
    "full_secret_ref_value",
    "full_credential_handle",
    "raw_request_payload",
    "raw_response_payload",
    "raw_audit_payload",
    "secret_derived_hash",
    "authorization_header",
    "cookie",
    "provider_error_detail",
}

EXPECTED_ZERO_COUNTERS = {
    "real_secret_read_count",
    "environment_secret_read_count",
    "secret_resolver_call_count",
    "fake_resolver_call_count",
    "production_resolver_call_count",
    "durable_audit_backend_selected_count",
    "storage_adapter_runtime_created_count",
    "audit_store_runtime_created_count",
    "audit_writer_runtime_created_count",
    "audit_event_write_count",
    "delivery_execution_count",
    "idempotency_decision_count",
    "duplicate_detection_count",
    "operator_approval_runtime_execution_count",
    "credential_handle_runtime_created_count",
    "credential_payload_created_count",
    "backend_health_runtime_created_count",
    "backend_health_check_count",
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
    "run audit store durable backend selection readiness checker",
    "run audit store durable backend boundary readiness checker",
    "run audit store runtime event schema artifact checker",
    "run audit store runtime blocker matrix checker",
    "run audit store runtime implementation entry refresh v4 checker",
    "run production resolver runtime implementation entry refresh v2 checker",
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


def source_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_durable_backend_selection_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-durable-backend-selection-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(
        slice_info.get("status") == "audit_store_durable_backend_selection_readiness_defined",
        "unexpected status",
    )
    require(
        slice_info.get("selection_decision")
        == "durable_backend_selection_deferred_until_backend_evidence_and_runtime_task_card",
        "unexpected selection decision",
    )
    for key, expected_path in {
        "task_card": (
            "docs/task-cards/"
            "production-secret-backend-audit-store-durable-backend-selection-readiness-v1-plan.md"
        ),
        "platform_topic": (
            "docs/platform/"
            "production-secret-backend-audit-store-durable-backend-selection-readiness-v1.md"
        ),
    }.items():
        value = str(slice_info.get(key) or "")
        require(value == expected_path, f"unexpected {key}")
        require((REPO_ROOT / value).exists(), f"{key} path missing: {value}")
    for claim in {
        "durable_audit_backend_ready",
        "durable_audit_backend_selected",
        "storage_adapter_runtime_created",
        "audit_store_runtime_task_card_created",
        "audit_store_runtime_created",
        "audit_writer_runtime_created",
        "delivery_runtime_created",
        "idempotency_runtime_created",
        "production_resolver_runtime_task_card_created",
        "production_resolver_runtime_created",
        "database_connection_provider_ready",
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
        source = load_json(REPO_ROOT / relative_path)
        require(source_status(source) == expected_status, f"{dependency_id} source status drifted")


def assert_selection_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("selection_boundary") or {}
    for field, expected in {
        "status": "audit_store_durable_backend_selection_readiness_defined",
        "selection_decision": "durable_backend_selection_deferred_until_backend_evidence_and_runtime_task_card",
        "durable_backend_selection_status": "deferred_without_backend_selection",
        "durable_audit_backend_status": "not_selected",
        "allowed_backend_kind_scope": "reserved_only",
        "storage_adapter_contract_status": "metadata_only_static_reference",
        "event_schema_artifact_status": "implemented_static_schema_artifact",
        "retention_policy_binding_status": "required_reference_only",
        "redaction_profile_binding_status": "required_reference_only",
        "audit_store_runtime_task_card_status": "not_created",
        "audit_store_runtime_status": "not_created",
        "audit_writer_runtime_status": "not_created",
        "delivery_runtime_status": "not_created",
        "idempotency_runtime_status": "not_created",
        "operator_approval_runtime_status": "not_created",
        "credential_handle_runtime_status": "not_created",
        "backend_health_runtime_status": "not_created",
        "no_secret_leakage_smoke_runtime_status": "not_created",
        "production_resolver_runtime_task_card_status": "not_created",
        "production_resolver_runtime_status": "not_created",
        "production_secret_backend_status": "not_satisfied",
        "cloud_secret_service_status": "not_selected",
        "database_connection_provider_status": "blocked",
        "repository_mode_status": "disabled",
        "production_api_status": "not_created",
    }.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "runtime_task_card_allowed_now",
        "durable_audit_backend_selected_in_this_slice",
        "storage_adapter_runtime_created_in_this_slice",
        "audit_store_runtime_task_card_created_in_this_slice",
        "audit_store_runtime_created_in_this_slice",
        "audit_writer_runtime_created_in_this_slice",
        "audit_event_written_in_this_slice",
        "delivery_runtime_created_in_this_slice",
        "delivery_executed_in_this_slice",
        "idempotency_runtime_created_in_this_slice",
        "duplicate_detector_created_in_this_slice",
        "operator_approval_runtime_created_in_this_slice",
        "credential_handle_runtime_created_in_this_slice",
        "credential_payload_created_in_this_slice",
        "backend_health_runtime_created_in_this_slice",
        "backend_health_check_executed_in_this_slice",
        "no_secret_leakage_smoke_runtime_created_in_this_slice",
        "production_resolver_runtime_task_card_created_in_this_slice",
        "production_resolver_runtime_created_in_this_slice",
        "database_connection_provider_enabled",
        "repository_mode_enabled",
        "production_api_enabled",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_candidate_shape_and_matrix(fixture: dict[str, Any]) -> None:
    shape = fixture.get("candidate_shape") or {}
    required_fields = set(shape.get("required_fields") or [])
    for field in {
        "candidate_backend_kind",
        "selection_policy_version",
        "environment",
        "storage_owner_ref",
        "storage_adapter_contract_ref",
        "event_schema_ref",
        "retention_policy_ref",
        "redaction_profile_ref",
        "writer_runtime_dependency",
        "delivery_runtime_dependency",
        "idempotency_runtime_dependency",
        "operator_approval_dependency",
        "credential_handle_dependency",
        "backend_health_dependency",
        "no_leakage_smoke_dependency",
        "offline_validation_manifest_ref",
    }:
        require(field in required_fields, f"candidate shape missing field: {field}")
    require(
        set(shape.get("allowed_candidate_kinds") or [])
        == {
            "reserved_append_only_audit_log",
            "reserved_managed_audit_log",
            "reserved_operator_managed_audit_store",
        },
        "allowed candidate kinds drifted",
    )
    for forbidden_kind in {
        "postgresql_audit_store",
        "mysql_audit_store",
        "committed_json_file",
        "repository_memory_store",
        "in_memory_audit_store",
    }:
        require(
            forbidden_kind in set(shape.get("forbidden_candidate_kinds") or []),
            f"forbidden candidate kind missing: {forbidden_kind}",
        )
    for field in (
        "environment_cross_fallback_allowed",
        "storage_adapter_runtime_allowed_in_this_slice",
        "database_provider_allowed_in_this_slice",
        "repository_mode_allowed_in_this_slice",
        "production_api_allowed_in_this_slice",
        "secret_ref_value_allowed",
        "credential_payload_allowed",
    ):
        require(shape.get(field) is False, f"{field} must remain false")

    gates = rows_by_id(fixture, "selection_matrix", "gate")
    require(set(gates) == set(EXPECTED_SELECTION_GATES), "selection gate ids drifted")
    for gate, expected_status in EXPECTED_SELECTION_GATES.items():
        item = gates[gate]
        require(item.get("status") == expected_status, f"{gate} status drifted")
        require(item.get("source"), f"{gate} source missing")
        if gate not in {"event_schema_artifact", "repository_mode_runtime", "production_api"}:
            require(item.get("blocks_durable_backend_selection") is True, f"{gate} must block backend selection")
        require(
            item.get("blocks_audit_store_runtime_task_card") is True or gate == "event_schema_artifact",
            f"{gate} audit runtime task card flag drifted",
        )
    require(
        gates["event_schema_artifact"].get("blocks_audit_store_runtime_task_card") is False,
        "event schema artifact must not block by itself after implementation",
    )
    require(
        gates["repository_mode_runtime"].get("blocks_durable_backend_selection") is False,
        "repository mode must remain durable store stop line",
    )

    requirements = set(fixture.get("future_task_card_requirements") or [])
    for requirement in {
        "concrete_durable_backend_choice_must_be_separate_task",
        "storage_adapter_contract_must_reference_schema_artifact",
        "writer_runtime_entry_must_no_longer_be_blocked",
        "idempotency_runtime_entry_must_no_longer_be_blocked",
        "delivery_runtime_entry_must_no_longer_be_blocked",
        "operator_approval_runtime_entry_must_no_longer_be_blocked",
        "credential_handle_runtime_entry_must_no_longer_be_blocked",
        "backend_health_runtime_entry_must_no_longer_be_blocked",
        "no_leakage_smoke_runtime_entry_must_no_longer_be_blocked",
        "production_resolver_runtime_task_card_must_remain_after_audit_store_runtime",
        "diagnostics_must_remain_sanitized_reference_only",
    }:
        require(requirement in requirements, f"missing future task requirement: {requirement}")


def assert_alignment_with_existing_evidence() -> None:
    durable_boundary = load_json(DURABLE_BOUNDARY_PATH)
    boundary = durable_boundary.get("durable_backend_boundary") or {}
    require(boundary.get("durable_backend_selection_status") == "deferred_until_runtime_task_card", "boundary selection status drifted")
    require(boundary.get("durable_audit_backend_status") == "not_selected", "boundary selected backend")
    require(boundary.get("audit_store_runtime_status") == "not_created", "boundary created audit runtime")

    artifact = load_json(SCHEMA_ARTIFACT_PATH)
    artifact_boundary = artifact.get("artifact_boundary") or {}
    require(
        artifact_boundary.get("status") == "audit_runtime_event_schema_artifact_implemented_static",
        "schema artifact boundary status drifted",
    )
    require(
        artifact_boundary.get("audit_writer_runtime_created_in_this_slice") is False,
        "schema artifact must not create writer runtime",
    )

    v4 = load_json(V4_REFRESH_PATH)
    v4_boundary = v4.get("entry_refresh_boundary") or {}
    require(
        v4_boundary.get("entry_decision") == "audit_store_runtime_task_card_still_blocked_before_runtime_task_card",
        "v4 entry decision drifted",
    )
    require(v4_boundary.get("durable_audit_backend_status") == "not_selected", "v4 selected durable backend")
    require(v4_boundary.get("audit_store_runtime_task_card_status") == "not_created", "v4 opened task card")

    writer = load_json(WRITER_BOUNDARY_PATH)
    writer_boundary = writer.get("writer_runtime_boundary") or {}
    require(writer_boundary.get("writer_runtime_status") == "not_created", "writer runtime created")

    delivery = load_json(DELIVERY_READINESS_PATH)
    delivery_boundary = delivery.get("delivery_runtime_boundary") or {}
    require(delivery_boundary.get("delivery_runtime_status") == "not_created", "delivery runtime created")

    idempotency = load_json(IDEMPOTENCY_READINESS_PATH)
    idempotency_boundary = idempotency.get("idempotency_runtime_boundary") or {}
    require(idempotency_boundary.get("idempotency_runtime_status") == "not_created", "idempotency runtime created")

    resolver = load_json(PRODUCTION_RESOLVER_V2_PATH)
    resolver_boundary = resolver.get("refresh_boundary") or {}
    require(
        resolver_boundary.get("entry_decision") == "production_resolver_runtime_task_card_still_blocked_after_refresh_v2",
        "production resolver v2 decision drifted",
    )
    require(resolver_boundary.get("production_resolver_runtime_task_card_status") == "not_created", "resolver task card created")

    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, expected in {
        "production_secret_backend_status": "not_satisfied",
        "audit_store_durable_backend_boundary_readiness_status": "defined_without_backend_selection",
        "durable_audit_backend_status": "static_backend_family_selected_runtime_blocked",
        "audit_store_runtime_task_card_status": "not_created",
        "audit_store_runtime_status": "not_created",
        "audit_writer_status": "not_created",
        "audit_delivery_runtime_status": "not_created",
        "audit_idempotency_runtime_status": "not_created",
        "production_resolver_runtime_task_card_status": "not_created",
        "production_resolver_runtime_status": "not_created",
    }.items():
        require(target.get(field) == expected, f"implementation readiness {field} drifted")


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
    require(diagnostics.get("credential_payload_allowed_in_diagnostics") is False, "credential payload must not be emitted")
    require(diagnostics.get("database_endpoint_allowed_in_diagnostics") is False, "database endpoint must not be emitted")

    fallback = set(fixture.get("no_fallback_policy") or [])
    for rule in (
        "no fallback from durable backend selection to memory store",
        "no fallback from durable backend selection to fake resolver",
        "no fallback from durable backend selection to committed fixture",
        "no fallback from missing durable backend to audit writer runtime",
        "no fallback from missing storage adapter contract to database provider",
        "no runtime task card if writer runtime is blocked",
        "no runtime task card if idempotency runtime is blocked",
        "no runtime task card if delivery runtime is blocked",
        "no runtime task card if operator approval runtime is blocked",
        "no runtime task card if credential handle runtime is blocked",
        "no runtime task card if backend health runtime is blocked",
        "no runtime task card if no leakage smoke runtime is blocked",
        "durable backend selection readiness does not mean workflow repository mode ready",
    ):
        require(rule in fallback, f"missing no fallback rule: {rule}")

    side_effect_policy = set(fixture.get("no_side_effect_policy") or [])
    for rule in (
        "no storage adapter runtime creation",
        "no audit store runtime creation",
        "no audit writer runtime creation",
        "no audit event write",
        "no delivery execution",
        "no idempotency decision",
        "no duplicate detection",
        "no database connection",
        "no driver open",
        "no SQL execution",
        "no schema marker read",
        "no schema marker write",
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
    artifact_call = 'run_python_script("check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py", [])'
    current_call = (
        'run_python_script("check-production-ops-secret-backend-audit-store-'
        'durable-backend-selection-readiness-v1.py", [])'
    )
    matrix_call = 'run_python_script("check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py", [])'
    for call in (artifact_call, current_call, matrix_call):
        require(call in check_repo, f"check-repo.py missing call: {call}")
    require(check_repo.index(artifact_call) < check_repo.index(current_call) < check_repo.index(matrix_call), "check order drifted")


def assert_no_secret_literals() -> None:
    paths = [
        "docs/platform/production-secret-backend-audit-store-durable-backend-selection-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-durable-backend-selection-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-selection-readiness-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"durable backend selection artifacts contain forbidden secret-looking literals: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "selection artifacts contain sk-like token")
    require(re.search(r"://[^\s:/]+:[^\s@]+@", text) is None, "selection artifacts contain dsn-like credential")


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
    print("production ops secret backend audit store durable backend selection readiness checks passed.")


if __name__ == "__main__":
    main()
