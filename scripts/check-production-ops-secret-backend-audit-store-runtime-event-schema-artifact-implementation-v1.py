#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.json"
)
CONTRACT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_STATUS = "audit_store_runtime_event_schema_artifact_implementation_task_card_defined"
EXPECTED_SCHEMA_PATH = "contracts/production-secret-audit-event.schema.json"
EXPECTED_ALLOWED_ARTIFACTS = {
    "docs/platform/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.md",
    "docs/task-cards/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1-plan.md",
    "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.json",
    "scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.py",
}
EXPECTED_DEPENDENCIES = {
    "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.json"
        ),
        "audit_store_runtime_event_schema_artifact_implementation_entry_review_defined",
    ),
    "production-secret-backend-audit-store-contract-event-schema-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json",
        "audit_store_contract_event_schema_readiness_defined",
    ),
    "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1": (
        (
            "scripts/checks/fixtures/"
            "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.json"
        ),
        "audit_store_runtime_event_schema_materialization_readiness_defined",
    ),
    "production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json",
        "audit_store_writer_runtime_boundary_readiness_defined",
    ),
    "production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4": (
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v4.json",
        "audit_store_runtime_implementation_entry_refresh_v4_defined",
    ),
    "production-secret-backend-implementation-readiness": (
        "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        "implementation_readiness_defined",
    ),
}
EXPECTED_FALSE_FLAGS = {
    "runtime_event_schema_artifact_created_in_this_slice",
    "runtime_event_schema_created_in_this_slice",
    "runtime_schema_validator_created_in_this_slice",
    "schema_validation_runtime_created_in_this_slice",
    "audit_store_runtime_task_card_created_in_this_slice",
    "audit_store_runtime_created_in_this_slice",
    "audit_writer_runtime_created_in_this_slice",
    "audit_writer_created_in_this_slice",
    "audit_event_written_in_this_slice",
    "writer_result_created_in_this_slice",
    "delivery_runtime_created_in_this_slice",
    "delivery_executed_in_this_slice",
    "idempotency_runtime_created_in_this_slice",
    "duplicate_detector_created_in_this_slice",
    "production_resolver_runtime_created_in_this_slice",
    "cloud_secret_client_created_in_this_slice",
    "database_connection_provider_enabled",
    "repository_mode_enabled",
    "production_api_enabled",
}
EXPECTED_FAILURE_CODES = {
    "audit_store_runtime_event_schema_artifact_implementation_task_card_missing",
    "audit_store_runtime_event_schema_artifact_implementation_scope_missing",
    "audit_store_runtime_event_schema_artifact_implementation_field_boundary_missing",
    "audit_store_runtime_event_schema_artifact_implementation_event_kind_boundary_missing",
    "audit_store_runtime_event_schema_artifact_implementation_forbidden_field_missing",
    "audit_store_runtime_event_schema_artifact_implementation_validation_plan_missing",
    "audit_store_runtime_event_schema_artifact_created_in_task_card",
    "audit_store_runtime_event_schema_validator_created_in_task_card",
    "audit_store_runtime_event_schema_artifact_writer_runtime_forbidden",
    "audit_store_runtime_event_schema_artifact_event_write_forbidden",
    "audit_store_runtime_event_schema_artifact_runtime_scope_overreach",
    "audit_store_runtime_event_schema_artifact_fallback_forbidden",
    "audit_store_runtime_event_schema_artifact_secret_material_detected",
}
EXPECTED_VALIDATION = {
    "runtime event schema artifact implementation task card checker",
    "runtime event schema artifact implementation entry review checker",
    "audit store contract event schema readiness checker",
    "runtime event schema materialization readiness checker",
    "positive metadata-only event fixture",
    "missing required field negative fixture",
    "forbidden field negative fixture",
    "additionalProperties negative fixture",
    "event kind invalid negative fixture",
    "schema validation checker",
    "writer input compatibility smoke",
    "fast repository check",
}
EXPECTED_MUST_NOT_INCLUDE = {
    "audit writer runtime",
    "audit store runtime",
    "delivery runtime",
    "idempotency runtime",
    "durable backend selection",
    "production resolver runtime",
    "cloud secret SDK",
    "cloud secret client",
    "DB provider",
    "SQL migration",
    "schema marker",
    "repository mode runtime",
    "public production API",
}
EXPECTED_DIAGNOSTICS = {
    "audit_store_runtime_event_schema_artifact_implementation_status",
    "schema_artifact_path_status",
    "schema_version_pin_status",
    "event_kind_allowlist_status",
    "required_fields_status",
    "optional_fields_status",
    "reference_only_fields_status",
    "forbidden_fields_status",
    "schema_validation_plan_status",
    "writer_input_compatibility_status",
    "runtime_event_schema_artifact_status",
    "runtime_schema_validator_status",
    "audit_writer_runtime_status",
    "audit_store_runtime_status",
    "production_resolver_runtime_status",
    "database_connection_provider_status",
    "repository_mode_status",
    "production_api_status",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}
EXPECTED_EXTRA_FORBIDDEN_FIELDS = {
    "raw_writer_payload",
    "raw_event_payload",
    "schema_payload",
    "payload_hash",
    "event_payload_hash",
    "secret_derived_hash",
}
EXPECTED_FORBIDDEN_SOURCES = {
    "fake_resolver_runtime",
    "developer_env_plaintext",
    "fixture_credential",
    "committed_value",
    "sample",
    "mock_provider",
    "local_smoke_profile",
    "operator_runbook_text",
    "repository_memory_store",
    "audit_memory_store",
    "static_handoff_envelope",
    "historical_audit_event",
    "runtime_schema_sample",
    "schema_from_payload",
    "writer_output",
    "delivery_result",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "runtime_event_schema_artifact",
    "runtime_schema_validator",
    "schema_validation_runtime",
    "audit_store_runtime_implementation_task_card",
    "audit_store_runtime",
    "audit_writer_runtime",
    "audit_writer",
    "audit_event_writer",
    "writer_result_fixture",
    "audit_delivery_runtime",
    "audit_idempotency_runtime",
    "duplicate_detector_runtime",
    "retry_executor",
    "production_resolver_runtime",
    "cloud_secret_sdk_or_client",
    "database_connection_provider",
    "sql_migration",
    "schema_marker",
    "repository_mode_runtime",
    "public_production_api",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a JSON object")
    return document


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def status_of(path: str) -> str:
    document = load_json(REPO_ROOT / path)
    slice_info = document.get("slice") or {}
    return str(slice_info.get("status") or document.get("status") or "")


def contract_event_schema() -> dict[str, Any]:
    contract = load_json(CONTRACT_FIXTURE_PATH)
    schema = contract.get("event_schema") or {}
    require(isinstance(schema, dict), "contract event_schema must be an object")
    return schema


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind")
        == "production_ops_secret_backend_audit_store_runtime_event_schema_artifact_implementation_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == EXPECTED_STATUS, "unexpected slice status")
    for field in ("task_card", "platform_topic"):
        path = str(slice_info.get(field) or "")
        require(path in EXPECTED_ALLOWED_ARTIFACTS, f"unexpected {field}: {path}")
        require((REPO_ROOT / path).exists(), f"{field} missing on disk: {path}")
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "runtime_event_schema_artifact_created",
        "runtime_schema_validator_created",
        "audit_writer_runtime_created",
        "audit_store_runtime_created",
        "audit_event_written",
        "delivery_runtime_created",
        "idempotency_runtime_created",
        "production_resolver_runtime_created",
        "repository_mode_ready",
        "production_api_ready",
        "production_ready",
    }:
        require(claim in forbidden_claims, f"does_not_claim missing {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = {str(item.get("id")): item for item in fixture.get("depends_on") or []}
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("evidence") == path, f"{dependency_id} evidence path drifted")
        require((REPO_ROOT / path).exists(), f"{dependency_id} evidence missing on disk")
        require(item.get("status") == expected_status, f"{dependency_id} fixture status drifted")
        require(status_of(path) == expected_status, f"{dependency_id} source status drifted")


def assert_task_card_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("task_card_boundary") or {}
    expected = {
        "status": EXPECTED_STATUS,
        "task_card_status": "created_static_task_card",
        "current_development_mode": "artifact_implementation_task_card_only_no_schema_artifact",
        "future_schema_artifact_path": EXPECTED_SCHEMA_PATH,
        "entry_review_consumed": True,
        "contract_event_schema_consumed": True,
        "materialization_readiness_consumed": True,
        "writer_boundary_consumed": True,
        "runtime_entry_refresh_v4_consumed": True,
    }
    for field, value in expected.items():
        require(boundary.get(field) == value, f"task_card_boundary.{field} drifted")
    for field in EXPECTED_FALSE_FLAGS:
        require(boundary.get(field) is False, f"task_card_boundary.{field} must remain false")


def assert_schema_artifact_requirements(fixture: dict[str, Any]) -> None:
    requirements = fixture.get("schema_artifact_requirements") or {}
    schema = contract_event_schema()
    require(requirements.get("schema_artifact_path") == EXPECTED_SCHEMA_PATH, "schema artifact path drifted")
    require(requirements.get("schema_version_pin") == schema.get("event_version"), "schema version pin drifted")
    require(
        requirements.get("event_kind_allowlist_source") == "audit_store_contract_event_schema_readiness_defined",
        "event kind source drifted",
    )
    require(
        requirements.get("field_source") == "audit_store_contract_event_schema_readiness_defined",
        "field source drifted",
    )
    for field in ("required_fields", "optional_fields", "event_kind_allowlist", "reference_only_fields"):
        require(requirements.get(field) == schema.get(field), f"{field} must match contract readiness")
    forbidden_fields = set(requirements.get("forbidden_fields") or [])
    contract_forbidden = set(schema.get("forbidden_fields") or [])
    require(contract_forbidden <= forbidden_fields, "contract forbidden fields missing from schema requirements")
    require(
        EXPECTED_EXTRA_FORBIDDEN_FIELDS <= forbidden_fields,
        "extra raw payload / hash forbidden fields missing",
    )
    require(
        EXPECTED_VALIDATION <= set(requirements.get("required_validation") or []),
        "schema validation plan missing entries",
    )
    require(
        EXPECTED_MUST_NOT_INCLUDE <= set(requirements.get("must_not_include") or []),
        "must_not_include scope missing entries",
    )


def assert_failures_and_policies(fixture: dict[str, Any]) -> None:
    failure_codes = {str(item.get("code")) for item in fixture.get("failure_mapping") or []}
    require(failure_codes == EXPECTED_FAILURE_CODES, "failure mapping codes drifted")
    forbidden_material = set(fixture["sanitized_diagnostics"]["forbidden_fields"])
    for item in fixture.get("failure_mapping") or []:
        require(item.get("boundary"), f"{item.get('code')} missing boundary")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{item.get('code')} missing sanitized diagnostic")
        require(not any(material in diagnostic for material in forbidden_material), "unsafe diagnostic")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    allowed_fields = set(diagnostics.get("allowed_fields") or [])
    forbidden_fields = set(diagnostics.get("forbidden_fields") or [])
    require(EXPECTED_DIAGNOSTICS <= allowed_fields, "diagnostic allowlist missing fields")
    require(not (allowed_fields & forbidden_fields), "diagnostic allowlist intersects forbidden fields")
    require(
        EXPECTED_EXTRA_FORBIDDEN_FIELDS <= forbidden_fields,
        "diagnostic forbidden fields missing raw payload / hash fields",
    )

    fallback = fixture.get("no_fallback_policy") or {}
    require(fallback.get("status") == "defined", "no fallback policy status drifted")
    require(
        EXPECTED_FORBIDDEN_SOURCES <= set(fallback.get("forbidden_sources") or []),
        "no fallback forbidden source coverage drifted",
    )
    require(fallback.get("missing_schema_artifact_result") == "fail_closed", "missing schema result drifted")
    require(
        fallback.get("missing_dependency_must_not_create_audit_success") is True,
        "missing dependency must fail closed",
    )

    counters = fixture.get("side_effect_counters") or {}
    require(counters, "side effect counters missing")
    for name, value in counters.items():
        require(value == 0, f"side effect counter {name} must stay 0")

    guard = fixture.get("artifact_guard") or {}
    require(set(guard.get("allowed_added_artifacts") or []) == EXPECTED_ALLOWED_ARTIFACTS, "allowed artifacts drifted")
    require(
        EXPECTED_FORBIDDEN_ARTIFACTS <= set(guard.get("forbidden_artifact_kinds") or []),
        "forbidden artifact list missing entries",
    )


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    expected_fields = {
        "production_secret_backend_status": "not_satisfied",
        "audit_runtime_event_schema_artifact_implementation_entry_review_status": "ready_for_artifact_task_card",
        "audit_runtime_event_schema_artifact_implementation_task_card_status": "defined_without_schema_artifact",
        "audit_runtime_event_schema_artifact_status": "implemented_static_schema_artifact",
        "audit_runtime_event_schema_version_pin_status": "static_contract_version_required",
        "audit_runtime_event_kind_allowlist_source_status": "static_contract_reference_only",
        "audit_runtime_required_optional_fields_source_status": "static_contract_reference_only",
        "audit_runtime_schema_writer_input_compatibility_status": "metadata_only_static_boundary_defined",
        "audit_store_runtime_task_card_status": "not_created",
        "audit_store_runtime_status": "not_created",
        "audit_writer_status": "not_created",
        "audit_event_delivery_status": "not_executed",
    }
    require(alignment.get("status") == EXPECTED_STATUS, "alignment status drifted")
    for field, expected in expected_fields.items():
        require(alignment.get(field) == expected, f"alignment.{field} drifted")
        require(target.get(field) == expected, f"implementation readiness {field} drifted")

    planned = {str(item.get("id")): item for item in readiness.get("planned_slices") or []}
    planned_item = planned.get("audit-store-runtime-event-schema-artifact-implementation") or {}
    require(planned_item.get("status") == EXPECTED_STATUS, "implementation readiness planned slice missing task card")
    require(
        EXPECTED_ALLOWED_ARTIFACTS <= set(planned_item.get("evidence") or []),
        "implementation readiness planned slice missing task card evidence",
    )


def assert_docs_and_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_check = (
        'run_python_script("'
        "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.py"
        '", [])'
    )
    require(current_check in check_repo, "check-repo.py must run artifact implementation task card check")
    previous_check = "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-entry-review-v1.py"
    next_check = "check-production-ops-secret-backend-production-resolver-runtime-implementation-entry-refresh-v2.py"
    require(
        check_repo.index(previous_check) < check_repo.index(current_check),
        "artifact implementation check must run after artifact entry review check",
    )
    require(
        check_repo.index(current_check) < check_repo.index(next_check),
        "artifact implementation check must run before production resolver runtime refresh v2",
    )

    docs = {
        "docs/platform/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.md": [
            EXPECTED_STATUS,
            EXPECTED_SCHEMA_PATH,
            "Schema Artifact Implementation Requirements",
            "No Fallback / No Side Effects",
            "Artifact Guard",
        ],
        "docs/task-cards/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1-plan.md": [
            EXPECTED_STATUS,
            EXPECTED_SCHEMA_PATH,
            "Artifact Requirements",
            "停止线",
        ],
        "docs/platform/README.md": [
            "runtime event schema artifact implementation",
            EXPECTED_STATUS,
        ],
        "docs/radishmind-current-focus.md": [
            EXPECTED_STATUS,
            "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1",
        ],
        "docs/radishmind-roadmap.md": [
            EXPECTED_STATUS,
        ],
        "docs/features/README.md": [
            EXPECTED_STATUS,
        ],
        "docs/features/workflow/README.md": [
            EXPECTED_STATUS,
        ],
        "docs/features/workflow-agent-runtime.md": [
            EXPECTED_STATUS,
        ],
        "docs/features/workflow/saved-workflow-draft-v1.md": [
            EXPECTED_STATUS,
        ],
        "docs/radishmind-integration-contracts.md": [
            EXPECTED_STATUS,
        ],
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
            "audit-store-runtime-event-schema-artifact-implementation",
            EXPECTED_STATUS,
        ],
        "docs/task-cards/README.md": [
            "Production Secret Backend Audit Store Runtime Event Schema Artifact Implementation v1",
            EXPECTED_STATUS,
        ],
        "scripts/README.md": [
            "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.py",
        ],
        "docs/devlogs/2026-W26.md": [
            EXPECTED_STATUS,
        ],
    }
    for path, literals in docs.items():
        text = read(path)
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{path} missing literals: {missing}")


def assert_no_secret_literals() -> None:
    text = "\n".join(
        read(path)
        for path in EXPECTED_ALLOWED_ARTIFACTS
        if path.endswith(".md") or path.endswith(".json")
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "-----BEGIN"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"artifact implementation evidence contains forbidden literal: {found}")
    require(re.search(r"sk-[A-Za-z0-9]{8,}", text) is None, "secret-looking sk token found")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_task_card_boundary(fixture)
    assert_schema_artifact_requirements(fixture)
    assert_failures_and_policies(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_registration()
    assert_no_secret_literals()
    print("production ops secret backend audit store runtime event schema artifact implementation checks passed.")


if __name__ == "__main__":
    main()
