#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any

try:
    from jsonschema import Draft202012Validator
except ModuleNotFoundError as exc:  # pragma: no cover - guarded by bootstrap/check-repo.
    raise SystemExit("jsonschema is required; run ./scripts/bootstrap-dev.sh first") from exc


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.json"
)
CONTRACT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json"
)
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
SCHEMA_PATH = REPO_ROOT / "contracts/production-secret-audit-event.schema.json"

EXPECTED_STATUS = "audit_store_runtime_event_schema_artifact_implemented"
EXPECTED_SLICE_ID = "production-secret-backend-audit-store-runtime-event-schema-artifact-v1"
EXPECTED_SCHEMA_VERSION = "audit-event-schema-v1"
EXPECTED_POSITIVE_FIXTURE = (
    "scripts/checks/fixtures/production-secret-audit-event-positive-v1.json"
)
EXPECTED_NEGATIVE_FIXTURES = {
    "scripts/checks/fixtures/production-secret-audit-event-missing-required-negative-v1.json",
    "scripts/checks/fixtures/production-secret-audit-event-forbidden-field-negative-v1.json",
    "scripts/checks/fixtures/production-secret-audit-event-additional-properties-negative-v1.json",
    "scripts/checks/fixtures/production-secret-audit-event-event-kind-invalid-negative-v1.json",
}
EXPECTED_EXTRA_FORBIDDEN_FIELDS = {
    "raw_writer_payload",
    "raw_event_payload",
    "schema_payload",
    "payload_hash",
    "event_payload_hash",
    "secret_derived_hash",
}
EXPECTED_FAILURE_CODES = {
    "audit_store_runtime_event_schema_artifact_missing",
    "audit_store_runtime_event_schema_required_field_missing",
    "audit_store_runtime_event_schema_forbidden_field_allowed",
    "audit_store_runtime_event_schema_additional_property_allowed",
    "audit_store_runtime_event_schema_event_kind_invalid_allowed",
    "audit_store_runtime_event_schema_fixture_expected_result_drifted",
    "audit_store_runtime_event_schema_writer_input_compatibility_missing",
    "audit_store_runtime_event_schema_artifact_runtime_scope_overreach",
    "audit_store_runtime_event_schema_artifact_fallback_forbidden",
    "audit_store_runtime_event_schema_artifact_secret_material_detected",
}
EXPECTED_RUNTIME_FALSE_FLAGS = {
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
EXPECTED_FORBIDDEN_RUNTIME_FILES = {
    "services/platform/internal/secretbackend/audit_store.go",
    "services/platform/internal/secretbackend/audit_writer.go",
    "services/platform/internal/secretbackend/audit_event_schema.go",
    "services/platform/internal/secretbackend/audit_delivery.go",
    "services/platform/internal/secretbackend/audit_idempotency.go",
    "services/platform/internal/secretbackend/production_resolver.go",
    "services/platform/migrations/production_secret_audit_store/001_audit_events.sql",
}
SECRET_LITERAL_PATTERNS = [
    re.compile(r"Bearer\s+[A-Za-z0-9._-]+"),
    re.compile(r"-----BEGIN [A-Z ]+-----"),
    re.compile(r"AKIA[0-9A-Z]{16}"),
    re.compile(r"sk-[A-Za-z0-9]{20,}"),
]


def load_json(path: Path) -> Any:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise AssertionError(message)


def relpath(path: Path) -> str:
    return path.relative_to(REPO_ROOT).as_posix()


def require_path(path: str) -> None:
    require((REPO_ROOT / path).exists(), f"required artifact missing: {path}")


def forbidden_fields_from_schema(schema: dict[str, Any]) -> set[str]:
    fields: set[str] = set()
    for guard in schema.get("allOf") or []:
        required = ((guard.get("not") or {}).get("required")) or []
        require(len(required) == 1, f"forbidden guard must contain exactly one field: {guard}")
        fields.add(required[0])
    return fields


def validator_for(schema: dict[str, Any]) -> Draft202012Validator:
    Draft202012Validator.check_schema(schema)
    return Draft202012Validator(schema)


def is_valid(validator: Draft202012Validator, candidate: dict[str, Any]) -> bool:
    return not list(validator.iter_errors(candidate))


def assert_slice(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == EXPECTED_SLICE_ID, "slice id drifted")
    require(slice_info.get("status") == EXPECTED_STATUS, "slice status drifted")
    require(
        slice_info.get("task_card")
        == "docs/task-cards/production-secret-backend-audit-store-runtime-event-schema-artifact-v1-plan.md",
        "slice task card drifted",
    )
    require(
        slice_info.get("platform_topic")
        == "docs/platform/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.md",
        "slice platform topic drifted",
    )
    forbidden_claims = {
        "production_ready",
        "production_secret_backend_ready",
        "audit_store_runtime_ready",
        "audit_store_runtime_created",
        "audit_writer_created",
        "audit_event_written",
        "audit_event_delivery_executed",
        "durable_backend_selected",
        "production_api_ready",
    }
    claims = set(slice_info.get("does_not_claim") or [])
    missing = forbidden_claims - claims
    require(not missing, f"slice must keep runtime/product claims out: {sorted(missing)}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    expected = {
        "production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1": (
            "audit_store_runtime_event_schema_artifact_implementation_task_card_defined",
            "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-event-schema-artifact-implementation-v1.json",
        ),
        "production-secret-backend-audit-store-contract-event-schema-readiness-v1": (
            "audit_store_contract_event_schema_readiness_defined",
            "scripts/checks/fixtures/production-secret-backend-audit-store-contract-event-schema-readiness-v1.json",
        ),
        "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1": (
            "audit_store_runtime_event_schema_materialization_readiness_defined",
            (
                "scripts/checks/fixtures/"
                "production-secret-backend-audit-store-runtime-event-schema-materialization-readiness-v1.json"
            ),
        ),
        "production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1": (
            "audit_store_writer_runtime_boundary_readiness_defined",
            "scripts/checks/fixtures/production-secret-backend-audit-store-writer-runtime-boundary-readiness-v1.json",
        ),
        "production-secret-backend-implementation-readiness": (
            "implementation_readiness_defined",
            "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json",
        ),
    }
    by_id = {item.get("id"): item for item in fixture.get("depends_on") or []}
    require(set(expected) <= set(by_id), "dependency list missing required upstream artifacts")
    for dep_id, (status, evidence) in expected.items():
        item = by_id[dep_id]
        require(item.get("status") == status, f"{dep_id} status drifted")
        require(item.get("evidence") == evidence, f"{dep_id} evidence drifted")
        require_path(evidence)


def assert_schema_contract(
    fixture: dict[str, Any], schema: dict[str, Any], contract_fixture: dict[str, Any]
) -> None:
    schema_contract = fixture.get("schema_contract") or {}
    contract = contract_fixture.get("event_schema") or {}
    required = list(contract.get("required_fields") or [])
    optional = list(contract.get("optional_fields") or [])
    event_kinds = list(contract.get("event_kind_allowlist") or [])
    forbidden = set(contract.get("forbidden_fields") or []) | EXPECTED_EXTRA_FORBIDDEN_FIELDS

    require(schema_contract.get("event_version") == EXPECTED_SCHEMA_VERSION, "fixture event version drifted")
    require(schema_contract.get("additional_properties") is False, "fixture additional properties must be false")
    require(schema_contract.get("metadata_only_event") is True, "fixture metadata-only event flag drifted")
    require(schema_contract.get("secret_material_allowed") is False, "fixture secret material flag drifted")
    require(schema_contract.get("raw_payload_hash_allowed") is False, "fixture payload hash flag drifted")
    require(set(schema_contract.get("required_fields") or []) == set(required), "fixture required fields drifted")
    require(set(schema_contract.get("optional_fields") or []) == set(optional), "fixture optional fields drifted")
    require(set(schema_contract.get("event_kind_allowlist") or []) == set(event_kinds), "fixture event kinds drifted")
    require(set(schema_contract.get("forbidden_fields") or []) == forbidden, "fixture forbidden fields drifted")

    require(schema.get("$schema") == "https://json-schema.org/draft/2020-12/schema", "schema draft drifted")
    require(schema.get("$id") == "https://radishmind.local/contracts/production-secret-audit-event.schema.json", "schema id drifted")
    require(schema.get("type") == "object", "schema root type drifted")
    require(schema.get("additionalProperties") is False, "schema additionalProperties must be false")
    require(list(schema.get("required") or []) == required, "schema required field order must follow contract")

    properties = schema.get("properties") or {}
    require(set(properties) == set(required) | set(optional), "schema property allowlist drifted")
    require(properties.get("event_version", {}).get("const") == EXPECTED_SCHEMA_VERSION, "schema event_version drifted")
    require(properties.get("event_kind", {}).get("enum") == event_kinds, "schema event kind allowlist drifted")
    require(properties.get("delivery_mode", {}).get("enum") == ["fail_closed"], "delivery mode must fail closed")
    require(
        properties.get("secret_ref_key_status", {}).get("enum") == ["present", "missing", "not_required"],
        "secret_ref_key_status enum drifted",
    )
    require(forbidden_fields_from_schema(schema) == forbidden, "schema forbidden guards drifted")


def assert_validation_cases(fixture: dict[str, Any], schema: dict[str, Any]) -> None:
    validator = validator_for(schema)
    cases = fixture.get("validation_case_matrix") or []
    expected_cases = {
        "positive_metadata_only_event": (EXPECTED_POSITIVE_FIXTURE, True),
        "missing_required_event_id": (
            "scripts/checks/fixtures/production-secret-audit-event-missing-required-negative-v1.json",
            False,
        ),
        "forbidden_secret_material_fields": (
            "scripts/checks/fixtures/production-secret-audit-event-forbidden-field-negative-v1.json",
            False,
        ),
        "additional_property": (
            "scripts/checks/fixtures/production-secret-audit-event-additional-properties-negative-v1.json",
            False,
        ),
        "invalid_event_kind": (
            "scripts/checks/fixtures/production-secret-audit-event-event-kind-invalid-negative-v1.json",
            False,
        ),
    }
    by_case = {case.get("case"): case for case in cases}
    require(set(by_case) == set(expected_cases), "validation case matrix drifted")
    for case_name, (fixture_path, expected_valid) in expected_cases.items():
        case = by_case[case_name]
        require(case.get("fixture") == fixture_path, f"{case_name} fixture path drifted")
        require(case.get("expected_valid") is expected_valid, f"{case_name} expected result drifted")
        candidate = load_json(REPO_ROOT / fixture_path)
        actual = is_valid(validator, candidate)
        require(actual is expected_valid, f"{case_name} schema validation result drifted")

    positive = load_json(REPO_ROOT / EXPECTED_POSITIVE_FIXTURE)
    for field in schema.get("required") or []:
        candidate = dict(positive)
        candidate.pop(field, None)
        require(not is_valid(validator, candidate), f"schema allowed missing required field: {field}")

    for field in forbidden_fields_from_schema(schema):
        candidate = dict(positive)
        candidate[field] = "redacted-not-accepted"
        require(not is_valid(validator, candidate), f"schema allowed forbidden field: {field}")

    candidate = dict(positive)
    candidate["event_kind"] = "secret_value_read"
    require(not is_valid(validator, candidate), "schema allowed invalid event_kind")


def assert_artifact_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("artifact_boundary") or {}
    require(boundary.get("status") == "audit_runtime_event_schema_artifact_implemented_static", "artifact status drifted")
    require(boundary.get("schema_artifact_path") == relpath(SCHEMA_PATH), "schema path drifted")
    require(
        boundary.get("schema_validation_checker")
        == "scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py",
        "schema checker path drifted",
    )
    require(boundary.get("positive_fixture") == EXPECTED_POSITIVE_FIXTURE, "positive fixture path drifted")
    require(set(boundary.get("negative_fixtures") or []) == EXPECTED_NEGATIVE_FIXTURES, "negative fixture set drifted")
    require(boundary.get("runtime_event_schema_artifact_created_in_this_slice") is True, "schema artifact flag drifted")
    require(
        boundary.get("runtime_schema_validation_fixtures_created_in_this_slice") is True,
        "schema fixture flag drifted",
    )
    require(boundary.get("schema_checker_created_in_this_slice") is True, "schema checker flag drifted")
    for flag in EXPECTED_RUNTIME_FALSE_FLAGS:
        require(boundary.get(flag) is False, f"runtime overreach flag must be false: {flag}")


def assert_writer_input_compatibility(fixture: dict[str, Any], schema: dict[str, Any]) -> None:
    smoke = fixture.get("writer_input_compatibility_smoke") or {}
    require(smoke.get("status") == "implemented_static_schema_compatibility", "writer compatibility status drifted")
    require(smoke.get("input_contract") == "event_schema_allowlist_only", "writer input contract drifted")
    require(smoke.get("positive_fixture") == EXPECTED_POSITIVE_FIXTURE, "writer smoke positive fixture drifted")
    require(set(smoke.get("allowed_input_fields") or []) == set(schema.get("properties") or {}), "writer allowlist drifted")
    require(smoke.get("writer_runtime_created_in_this_slice") is False, "writer runtime must stay absent")
    require(smoke.get("writer_result_created_in_this_slice") is False, "writer result must stay absent")
    require(smoke.get("audit_event_write_executed_in_this_slice") is False, "audit write must stay absent")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    codes = set(fixture.get("failure_mapping") or [])
    missing = EXPECTED_FAILURE_CODES - codes
    extra = codes - EXPECTED_FAILURE_CODES
    require(not missing, f"failure mapping missing codes: {sorted(missing)}")
    require(not extra, f"failure mapping contains unexpected codes: {sorted(extra)}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    required_paths = set(guard.get("required_static_artifacts_exist") or [])
    expected = {
        relpath(SCHEMA_PATH),
        EXPECTED_POSITIVE_FIXTURE,
        *EXPECTED_NEGATIVE_FIXTURES,
        relpath(FIXTURE_PATH),
        "scripts/check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py",
        "docs/platform/production-secret-backend-audit-store-runtime-event-schema-artifact-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-runtime-event-schema-artifact-v1-plan.md",
        "docs/contracts/production-secret-audit-event.md",
    }
    require(expected <= required_paths, "artifact guard required path list drifted")
    for path in expected:
        require_path(path)

    forbidden_paths = set(guard.get("forbidden_runtime_files") or [])
    require(EXPECTED_FORBIDDEN_RUNTIME_FILES <= forbidden_paths, "forbidden runtime file guard drifted")
    for path in forbidden_paths:
        require(not (REPO_ROOT / path).exists(), f"runtime file must not exist in this slice: {path}")


def assert_implementation_readiness_alignment(fixture: dict[str, Any]) -> None:
    alignment = fixture.get("implementation_readiness_alignment") or {}
    expected = {
        "audit_runtime_event_schema_artifact_status": "implemented_static_schema_artifact",
        "audit_runtime_event_schema_artifact_validation_status": "implemented_offline_schema_validation",
        "audit_store_runtime_task_card_status": "not_created",
        "audit_store_runtime_status": "not_created",
        "audit_writer_status": "not_created",
        "audit_event_delivery_status": "not_executed",
        "production_secret_backend_status": "not_satisfied",
    }
    for field, value in expected.items():
        require(alignment.get(field) == value, f"fixture implementation readiness alignment drifted: {field}")

    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    target = readiness.get("implementation_target") or {}
    for field, value in expected.items():
        require(target.get(field) == value, f"implementation readiness target drifted: {field}")

    planned = {item.get("id"): item for item in readiness.get("planned_slices") or []}
    require(
        planned.get("audit-store-runtime-event-schema-artifact", {}).get("status") == EXPECTED_STATUS,
        "implementation readiness planned slice for schema artifact missing or drifted",
    )


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = reference.get("path")
        require(isinstance(path, str) and path, "doc reference path missing")
        text = (REPO_ROOT / path).read_text(encoding="utf-8")
        for needle in reference.get("must_include") or []:
            require(needle in text, f"{path} missing required text: {needle}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    checker_call = (
        'run_python_script("'
        "check-production-ops-secret-backend-audit-store-runtime-event-schema-artifact-v1.py"
        '", [])'
    )
    require(checker_call in check_repo, "check-repo.py must run schema artifact checker")


def assert_no_secret_literals() -> None:
    paths = [
        SCHEMA_PATH,
        FIXTURE_PATH,
        REPO_ROOT / EXPECTED_POSITIVE_FIXTURE,
        *[REPO_ROOT / item for item in EXPECTED_NEGATIVE_FIXTURES],
    ]
    for path in paths:
        text = path.read_text(encoding="utf-8")
        for pattern in SECRET_LITERAL_PATTERNS:
            require(not pattern.search(text), f"secret-like literal found in {relpath(path)}")


def assert_counters(fixture: dict[str, Any]) -> None:
    counters = fixture.get("counters") or {}
    for name, value in counters.items():
        require(value == 0, f"runtime counter must remain zero: {name}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    schema = load_json(SCHEMA_PATH)
    contract_fixture = load_json(CONTRACT_FIXTURE_PATH)

    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_schema_contract(fixture, schema, contract_fixture)
    assert_artifact_boundary(fixture)
    assert_validation_cases(fixture, schema)
    assert_writer_input_compatibility(fixture, schema)
    assert_failure_mapping(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_secret_literals()
    assert_counters(fixture)
    print("production ops secret backend audit store runtime event schema artifact checks passed.")


if __name__ == "__main__":
    main()
