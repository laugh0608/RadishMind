#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

try:
    import jsonschema
except ModuleNotFoundError:  # pragma: no cover - local bootstrap should install this.
    jsonschema = None


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-artifact-implementation-v1.json"
)
SCHEMA_PATH = REPO_ROOT / "contracts/radish-oidc-token-validation.schema.json"
POSITIVE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-positive-v1.json"
)
MISSING_REQUIRED_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-missing-required-negative-v1.json"
)
FORBIDDEN_FIELD_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-forbidden-field-negative-v1.json"
)
ADDITIONAL_PROPERTIES_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-additional-properties-negative-v1.json"
)
PRIOR_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-implementation-v1.json"
)
SCHEMA_TASK_CARD_READINESS_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-task-card-readiness-v1.json"
)
AUTH_RUNTIME_ENTRY_REVIEW_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1.json"
)
OIDC_UPSTREAM_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json"
)
PRODUCTION_AUTH_RUNTIME_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-token-validation-schema-implementation-v1": (
        PRIOR_IMPLEMENTATION_FIXTURE_PATH,
        "draft_token_validation_schema_implementation_task_card_defined",
    ),
    "workflow-saved-draft-token-validation-schema-task-card-readiness-v1": (
        SCHEMA_TASK_CARD_READINESS_FIXTURE_PATH,
        "draft_token_validation_schema_task_card_readiness_defined",
    ),
    "workflow-saved-draft-token-validation-auth-middleware-runtime-entry-review-v1": (
        AUTH_RUNTIME_ENTRY_REVIEW_FIXTURE_PATH,
        "draft_token_validation_auth_middleware_runtime_entry_review_defined",
    ),
    "radish-oidc-token-membership-upstream-evidence-refresh-v1": (
        OIDC_UPSTREAM_FIXTURE_PATH,
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ),
    "workflow-saved-draft-production-auth-runtime-v1": (
        PRODUCTION_AUTH_RUNTIME_FIXTURE_PATH,
        "draft_production_auth_runtime_bridge_implemented",
    ),
}
EXPECTED_REQUIRED_FIELDS = {
    "issuer_ref",
    "subject_ref",
    "tenant_ref",
    "audience_refs",
    "scope_grants",
    "workspace_binding_refs",
    "application_scope_refs",
    "owner_subject_ref",
    "key_id_ref",
    "algorithm",
    "issued_at",
    "expires_at",
    "auth_time",
    "policy_version",
    "request_id",
    "audit_ref",
}
EXPECTED_FORBIDDEN_FIELDS = {
    "raw_token",
    "authorization_header",
    "cookie",
    "client_secret",
    "refresh_token",
    "authorization_code",
    "jwks_raw_dump",
    "raw_claim_dump",
    "membership_raw_record",
    "database_detail",
    "provider_error_detail",
    "secret_value",
}
EXPECTED_VALIDATION_CASES = {
    "positive_verified_token_context": (POSITIVE_FIXTURE_PATH, True),
    "missing_required_field_negative": (MISSING_REQUIRED_FIXTURE_PATH, False),
    "forbidden_raw_material_field_negative": (FORBIDDEN_FIELD_FIXTURE_PATH, False),
    "additional_properties_negative": (ADDITIONAL_PROPERTIES_FIXTURE_PATH, False),
}
EXPECTED_FAILURE_CODES = {
    "token_validation_schema_artifact_missing",
    "token_validation_schema_required_field_missing",
    "token_validation_schema_forbidden_field_allowed",
    "token_validation_schema_additional_property_allowed",
    "token_validation_schema_fixture_expected_result_drifted",
    "token_validation_schema_artifact_runtime_scope_overreach",
    "token_validation_schema_artifact_dev_auth_fallback",
}
EXPECTED_RUNTIME_FORBIDDEN_FILES = {
    "services/platform/internal/httpapi/radish_oidc_auth_middleware.go",
    "services/platform/internal/httpapi/radish_oidc_token_validator.go",
    "services/platform/internal/httpapi/radish_membership_adapter.go",
    "scripts/checks/fixtures/workflow-saved-draft-negative-auth-smoke-runtime-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-negative-auth-smoke-runtime-v1.py",
    "services/platform/migrations/workflow_saved_drafts/001_saved_workflow_drafts.sql",
}
EXPECTED_REQUIRED_CHECKS = {
    "workflow_saved_draft_token_validation_schema_artifact_implementation_checker",
    "workflow_saved_draft_token_validation_schema_implementation_checker",
    "workflow_saved_draft_token_validation_schema_task_card_readiness_checker",
    "workflow_saved_draft_token_validation_auth_middleware_runtime_entry_review_checker",
    "radish_oidc_token_membership_upstream_evidence_refresh_checker",
    "git diff --check",
    "./scripts/check-repo.sh --fast",
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


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {
        str(row.get(id_field) or ""): row
        for row in fixture.get(key) or []
        if isinstance(row, dict)
    }
    require(rows, f"{key} must not be empty")
    return rows


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "workflow_saved_draft_token_validation_schema_artifact_implementation_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-token-validation-schema-artifact-implementation-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "track drifted")
    require(
        slice_info.get("status") == "draft_token_validation_schema_artifact_implemented",
        "status drifted",
    )
    for relative_path in (slice_info.get("task_card"), slice_info.get("feature_topic")):
        require(relative_path and (REPO_ROOT / str(relative_path)).exists(), f"missing slice doc: {relative_path}")
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "token_validator_created",
        "auth_middleware_created",
        "membership_adapter_created",
        "repository_mode_runtime_created",
        "database_runtime_created",
        "production_api_consumer_created",
        "workflow_executor_ready",
    }:
        require(claim in forbidden_claims, f"missing forbidden claim guard: {claim}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    require(set(dependencies) == set(EXPECTED_DEPENDENCIES), "dependency ids drifted")
    for dependency_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} status drifted")
        require(item.get("evidence") == str(path.relative_to(REPO_ROOT)), f"{dependency_id} evidence path drifted")
        dependency = load_json(path)
        require((dependency.get("slice") or {}).get("id") == dependency_id, f"{dependency_id} id drifted")
        require(
            (dependency.get("slice") or {}).get("status") == expected_status,
            f"{dependency_id} fixture status drifted",
        )


def assert_prior_guard_alignment() -> None:
    for path in (
        PRIOR_IMPLEMENTATION_FIXTURE_PATH,
        SCHEMA_TASK_CARD_READINESS_FIXTURE_PATH,
        AUTH_RUNTIME_ENTRY_REVIEW_FIXTURE_PATH,
    ):
        guard = load_json(path).get("artifact_guard") or {}
        forbidden = set(guard.get("files_must_not_exist") or [])
        forbidden.update(set(guard.get("future_files_must_not_exist") or []))
        no_longer_forbidden = {
            "contracts/radish-oidc-token-validation.schema.json",
            "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-positive-v1.json",
            "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-missing-required-negative-v1.json",
            "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-forbidden-field-negative-v1.json",
            "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-additional-properties-negative-v1.json",
        }
        overlap = sorted(forbidden & no_longer_forbidden)
        require(not overlap, f"{path.relative_to(REPO_ROOT)} still forbids schema artifact implementation: {overlap}")


def assert_artifact_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("artifact_boundary") or {}
    expected_values = {
        "status": "token_validation_schema_artifact_implemented_static",
        "implementation_track": "schema_artifact_only_verified_context_projection",
        "schema_artifact_path": "contracts/radish-oidc-token-validation.schema.json",
        "schema_validation_checker": (
            "scripts/checks/control_plane/"
            "check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py"
        ),
    }
    for field, expected in expected_values.items():
        require(boundary.get(field) == expected, f"{field} drifted")
    for field in (
        "schema_file_created_in_this_slice",
        "positive_fixture_created_in_this_slice",
        "missing_required_negative_fixture_created_in_this_slice",
        "forbidden_raw_material_negative_fixture_created_in_this_slice",
        "additional_properties_negative_fixture_created_in_this_slice",
        "schema_validation_checker_created_in_this_slice",
    ):
        require(boundary.get(field) is True, f"{field} must be true")
    for field in (
        "oidc_middleware_created_in_this_slice",
        "token_validator_created_in_this_slice",
        "auth_middleware_created_in_this_slice",
        "membership_adapter_created_in_this_slice",
        "negative_auth_smoke_runtime_created_in_this_slice",
        "repository_mode_enabled_in_this_slice",
        "database_runtime_created_in_this_slice",
        "production_api_consumer_created_in_this_slice",
        "executor_created_in_this_slice",
        "confirmation_created_in_this_slice",
        "writeback_or_replay_created_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def forbidden_fields_from_schema(schema: dict[str, Any]) -> set[str]:
    forbidden: set[str] = set()
    for guard in schema.get("allOf") or []:
        if not isinstance(guard, dict):
            continue
        required = ((guard.get("not") or {}).get("required") or [])
        if len(required) == 1:
            forbidden.add(str(required[0]))
    return forbidden


def assert_schema_contract(fixture: dict[str, Any]) -> dict[str, Any]:
    require(jsonschema is not None, "jsonschema is required; run scripts/bootstrap-dev.sh first")
    schema = load_json(SCHEMA_PATH)
    jsonschema.Draft202012Validator.check_schema(schema)
    require(schema.get("$schema") == "https://json-schema.org/draft/2020-12/schema", "$schema drifted")
    require(
        schema.get("$id") == "https://radishmind.local/contracts/radish-oidc-token-validation.schema.json",
        "$id drifted",
    )
    require(schema.get("type") == "object", "schema root must be object")
    require(schema.get("additionalProperties") is False, "schema must set additionalProperties=false")
    require(set(schema.get("required") or []) == EXPECTED_REQUIRED_FIELDS, "required fields drifted")
    properties = schema.get("properties") or {}
    require(set(properties) == EXPECTED_REQUIRED_FIELDS, "schema properties must match required sanitized fields")
    explicit_forbidden = forbidden_fields_from_schema(schema)
    require(
        explicit_forbidden == EXPECTED_FORBIDDEN_FIELDS,
        f"explicit forbidden field guards drifted: {sorted(explicit_forbidden)}",
    )
    require(not (EXPECTED_REQUIRED_FIELDS & EXPECTED_FORBIDDEN_FIELDS), "required fields overlap forbidden fields")

    contract = fixture.get("schema_contract") or {}
    require(contract.get("root_scope") == "verified_token_context_sanitized_projection", "root scope drifted")
    require(contract.get("additional_properties") is False, "fixture must declare additional_properties false")
    require(set(contract.get("required_fields") or []) == EXPECTED_REQUIRED_FIELDS, "fixture required fields drifted")
    require(set(contract.get("forbidden_fields") or []) == EXPECTED_FORBIDDEN_FIELDS, "fixture forbidden fields drifted")
    for field in (
        "sanitized_projection_only",
        "raw_material_allowed",
        "membership_raw_record_allowed",
        "database_detail_allowed",
        "secret_material_allowed",
    ):
        expected = field == "sanitized_projection_only"
        require(contract.get(field) is expected, f"{field} drifted")
    return schema


def assert_schema_validation_cases(fixture: dict[str, Any], schema: dict[str, Any]) -> None:
    require(jsonschema is not None, "jsonschema is required")
    validator = jsonschema.Draft202012Validator(schema)
    cases = rows_by_id(fixture, "validation_case_matrix", "case_id")
    require(set(cases) == set(EXPECTED_VALIDATION_CASES), "validation case ids drifted")
    for case_id, (path, expected_valid) in EXPECTED_VALIDATION_CASES.items():
        case = cases[case_id]
        require(case.get("fixture") == str(path.relative_to(REPO_ROOT)), f"{case_id} fixture path drifted")
        require(case.get("expected_valid") is expected_valid, f"{case_id} expected_valid drifted")
        document = load_json(path)
        is_valid = validator.is_valid(document)
        require(is_valid is expected_valid, f"{case_id} validation result drifted")

    positive = load_json(POSITIVE_FIXTURE_PATH)
    require(set(positive) == EXPECTED_REQUIRED_FIELDS, "positive fixture must only include required fields")
    require(not (set(positive) & EXPECTED_FORBIDDEN_FIELDS), "positive fixture includes forbidden fields")
    for field in EXPECTED_REQUIRED_FIELDS:
        candidate = dict(positive)
        candidate.pop(field)
        require(not validator.is_valid(candidate), f"schema allowed missing required field {field}")
    for field in EXPECTED_FORBIDDEN_FIELDS:
        candidate = dict(positive)
        candidate[field] = "redacted-not-accepted"
        require(not validator.is_valid(candidate), f"schema allowed forbidden field {field}")
    additional_candidate = dict(positive)
    additional_candidate["unexpected_projection"] = "not-accepted"
    require(not validator.is_valid(additional_candidate), "schema allowed additional property")

    forbidden_fixture = load_json(FORBIDDEN_FIELD_FIXTURE_PATH)
    require(
        EXPECTED_FORBIDDEN_FIELDS.issubset(set(forbidden_fixture)),
        "forbidden field negative fixture must cover all forbidden fields",
    )


def assert_failure_diagnostics_and_policies(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure codes drifted")
    for code, item in failures.items():
        require(str(item.get("failure_boundary") or "").strip(), f"{code} failure boundary is required")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} sanitized diagnostic is required")
        for forbidden in ("raw token", "authorization header", "cookie", "client secret", "dsn"):
            require(forbidden not in diagnostic.lower(), f"{code} diagnostic exposes {forbidden}")

    diagnostics = fixture.get("sanitized_diagnostics") or {}
    for field in EXPECTED_FORBIDDEN_FIELDS | {"dsn"}:
        require(field in set(diagnostics.get("forbidden_fields") or []), f"diagnostics missing {field}")
    require(diagnostics.get("runtime_emission_allowed_in_this_slice") is False, "runtime emission must stay false")

    for policy_key in ("no_fallback_policy", "no_side_effect_policy"):
        require(fixture.get(policy_key), f"{policy_key} is required")
    for counter, value in (fixture.get("side_effect_counters") or {}).items():
        require(value == 0, f"{counter} must remain zero")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    required = set(guard.get("required_static_artifacts_exist") or [])
    for path in {
        "contracts/radish-oidc-token-validation.schema.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-positive-v1.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-missing-required-negative-v1.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-forbidden-field-negative-v1.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-additional-properties-negative-v1.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-artifact-implementation-v1.json",
        "scripts/checks/control_plane/check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py",
    }:
        require(path in required, f"artifact guard missing required artifact: {path}")
    for relative_path in required:
        require((REPO_ROOT / relative_path).exists(), f"required artifact missing: {relative_path}")
    forbidden = set(guard.get("files_must_not_exist") or [])
    require(EXPECTED_RUNTIME_FORBIDDEN_FILES.issubset(forbidden), "runtime forbidden files drifted")
    for relative_path in forbidden:
        require(not (REPO_ROOT / relative_path).exists(), f"forbidden runtime artifact exists: {relative_path}")
    for root in guard.get("forbid_sql_files_under") or []:
        root_path = REPO_ROOT / str(root)
        sql_files = sorted(root_path.rglob("*.sql")) if root_path.exists() else []
        require(not sql_files, f"SQL files must not exist under {root}: {sql_files}")


def assert_docs_testing_and_check_repo(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        content = read(path)
        missing = [literal for literal in reference.get("must_contain") or [] if str(literal) not in content]
        require(not missing, f"{path} missing literals: {missing}")

    strategy = fixture.get("testing_strategy") or {}
    require(
        strategy.get("status") == "token_validation_schema_artifact_implementation_checker_static_schema_validation",
        "testing strategy status drifted",
    )
    require(EXPECTED_REQUIRED_CHECKS.issubset(set(strategy.get("required_checks") or [])), "required checks drifted")
    for field in (
        "does_not_start_service",
        "does_not_fetch_issuer_discovery",
        "does_not_fetch_jwks",
        "does_not_validate_real_token",
        "does_not_create_auth_middleware",
        "does_not_query_membership",
        "does_not_enable_repository_store_mode",
        "does_not_connect_database",
        "does_not_run_sql",
        "does_not_create_production_api",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-token-validation-schema-implementation-v1.py"
    current_checker = (
        "checks/control_plane/"
        "check-workflow-saved-draft-token-validation-schema-artifact-implementation-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    for checker in (previous_checker, current_checker, next_checker):
        require(checker in check_repo, f"{checker} missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "token validation schema artifact implementation checker order drifted",
    )


def assert_no_secret_literals() -> None:
    paths = [
        "contracts/radish-oidc-token-validation.schema.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-positive-v1.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-missing-required-negative-v1.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-forbidden-field-negative-v1.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-additional-properties-negative-v1.json",
        "scripts/checks/fixtures/workflow-saved-draft-token-validation-schema-artifact-implementation-v1.json",
    ]
    text = "\n".join(read(path) for path in paths)
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"token validation schema artifacts contain forbidden literals: {found}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_prior_guard_alignment()
    assert_artifact_boundary(fixture)
    schema = assert_schema_contract(fixture)
    assert_schema_validation_cases(fixture, schema)
    assert_failure_diagnostics_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_docs_testing_and_check_repo(fixture)
    assert_no_secret_literals()
    print("workflow saved draft token validation schema artifact implementation checks passed.")


if __name__ == "__main__":
    main()
