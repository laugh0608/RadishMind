#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from workflow_saved_draft_repository_adapter_implementation_guard import (
    repository_adapter_implementation_file_allowed,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json"
)
SCHEMA_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-migration-preconditions-v1.json"
)
SCHEMA_EVIDENCE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-evidence-v1.json"
)
SCHEMA_MANIFEST_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-manifest-v1.json"
)
MATERIALIZATION_REVIEW_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-review-v1.json"
)
SELECTOR_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json"
)
ADAPTER_SMOKE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-readiness-v1.json"
)
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json"
)
MIGRATION_ROOT = REPO_ROOT / "services/platform/migrations/workflow_saved_drafts"
MANIFEST_PATH = MIGRATION_ROOT / "manifest.json"
DDL_REVIEW_PATH = MIGRATION_ROOT / "ddl-review.md"
ROLLBACK_PATH = MIGRATION_ROOT / "rollback-evidence.json"
MIGRATION_SMOKE_PATH = MIGRATION_ROOT / "migration-smoke.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-schema-artifact-evidence-v1": (
        SCHEMA_EVIDENCE_FIXTURE_PATH,
        "draft_schema_artifact_evidence_defined",
    ),
    "workflow-saved-draft-schema-artifact-manifest-v1": (
        SCHEMA_MANIFEST_FIXTURE_PATH,
        "draft_schema_artifact_manifest_defined",
    ),
    "workflow-saved-draft-schema-artifact-materialization-review-v1": (
        MATERIALIZATION_REVIEW_FIXTURE_PATH,
        "draft_schema_artifact_materialization_review_defined",
    ),
    "workflow-saved-draft-store-selector-smoke-v1": (
        SELECTOR_IMPLEMENTATION_FIXTURE_PATH,
        "draft_store_selector_smoke_implemented",
    ),
    "workflow-saved-draft-adapter-smoke-readiness-v1": (
        ADAPTER_SMOKE_READINESS_FIXTURE_PATH,
        "draft_adapter_smoke_readiness_defined",
    ),
    "workflow-saved-draft-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "draft_repository_adapter_implementation_plan_defined",
    ),
}
EXPECTED_ARTIFACTS = {
    "manifest": MANIFEST_PATH,
    "ddl_review": DDL_REVIEW_PATH,
    "rollback_evidence": ROLLBACK_PATH,
    "migration_smoke": MIGRATION_SMOKE_PATH,
}
EXPECTED_OPERATIONS = {
    "SaveWorkflowDraftRecord",
    "ReadWorkflowDraftRecord",
    "ListWorkflowDraftRecords",
}
EXPECTED_FAILURE_CODES = {
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_store_contract_mismatch",
    "repository_store_disabled",
    "invalid_draft_store_mode",
}
EXPECTED_PRESERVED_FAILURE_CODES = {
    "draft_version_conflict",
    "draft_scope_denied",
    "draft_not_found",
    "draft_store_unavailable",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "repository_write_count=0",
    "database_write_count=0",
    "migration_apply_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_SMOKE_CASES = {
    "store_schema_version_recorded",
    "tenant_workspace_application_predicate",
    "owner_list_projection",
    "version_conflict_predicate",
    "migration_not_applied_failure",
    "schema_version_mismatch_failure",
    "no_side_effects",
}
FORBIDDEN_FILES = {
    "services/platform/internal/httpapi/workflow_saved_draft_repository.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter_test.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter_contract_smoke_test.go",
    "services/platform/internal/httpapi/workflow_saved_draft_schema_migration_runner.go",
    "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def schema_expected_fields() -> set[str]:
    contract = load_json(SCHEMA_PRECONDITIONS_FIXTURE_PATH).get("logical_schema_contract") or {}
    fields: set[str] = set()
    for group in (
        "scope_identity_fields",
        "ownership_actor_fields",
        "source_version_fields",
        "status_fields",
        "payload_fields",
        "audit_fields",
    ):
        fields.update(str(field) for field in contract.get(group) or [])
    return fields


def schema_expected_indexes() -> dict[str, dict[str, Any]]:
    strategy = load_json(SCHEMA_PRECONDITIONS_FIXTURE_PATH).get("index_strategy") or {}
    return {
        str(index.get("id") or ""): index
        for index in strategy.get("indexes") or []
        if isinstance(index, dict)
    }


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    missing = sorted(set(EXPECTED_DEPENDENCIES) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for expected_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == expected_id, f"dependency {expected_id} id drifted")
        require(
            slice_info.get("status") == expected_status,
            f"dependency {expected_id} status must be {expected_status}",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "workflow_saved_draft_schema_artifact_materialization_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-schema-artifact-materialization-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_schema_artifact_materialized_static",
        "schema artifact materialization status drifted",
    )
    for forbidden_claim in (
        "sql_migration_created",
        "migration_runner_implemented",
        "repository_adapter_ready",
        "durable_persistence_ready",
        "token_validation_ready",
        "production_api_consumer_ready",
        "workflow_executor_ready",
    ):
        require(forbidden_claim in set(slice_info.get("does_not_claim") or []), f"missing claim guard: {forbidden_claim}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("materialization_boundary") or {}
    review_boundary = load_json(MATERIALIZATION_REVIEW_FIXTURE_PATH).get("materialization_review_boundary") or {}
    manifest_boundary = load_json(SCHEMA_MANIFEST_FIXTURE_PATH).get("schema_artifact_manifest_boundary") or {}
    require(boundary.get("status") == "static_schema_artifact_materialized", "boundary status drifted")
    require(boundary.get("implementation_track") == "schema_artifact_materialization_only", "track drifted")
    for current_field, review_field in (
        ("schema_artifact_id", "future_schema_artifact_id"),
        ("manifest_version", "future_manifest_version"),
        ("store_schema_version", "future_store_schema_version"),
        ("migration_root", "future_migration_root"),
        ("manifest_file", "future_manifest_file"),
        ("ddl_review_file", "future_ddl_review_file"),
        ("rollback_evidence_file", "future_rollback_evidence_file"),
        ("migration_smoke_file", "future_migration_smoke_file"),
        ("schema_version_table", "future_schema_version_table"),
    ):
        require(boundary.get(current_field) == review_boundary.get(review_field), f"{current_field} must match review")
        require(boundary.get(current_field) == manifest_boundary.get(review_field), f"{current_field} must match manifest")
    for field in (
        "migration_root_created_in_this_slice",
        "manifest_file_created_in_this_slice",
        "ddl_review_artifact_created_in_this_slice",
        "rollback_evidence_artifact_created_in_this_slice",
        "migration_smoke_artifact_created_in_this_slice",
    ):
        require(boundary.get(field) is True, f"{field} must be true")
    for field in (
        "sql_files_created_in_this_slice",
        "schema_version_table_created_in_this_slice",
        "migration_runner_created_in_this_slice",
        "database_connection_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "repository_mode_enabled_in_this_slice",
        "adapter_smoke_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "publish_or_run_allowed_in_this_slice",
        "executor_allowed_in_this_slice",
        "writeback_or_replay_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_artifact_matrix(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(row.get("id") or ""): row
        for row in fixture.get("artifact_matrix") or []
        if isinstance(row, dict)
    }
    require(set(artifacts) == set(EXPECTED_ARTIFACTS), "artifact ids drifted")
    for artifact_id, path in EXPECTED_ARTIFACTS.items():
        row = artifacts[artifact_id]
        require(row.get("path") == str(path.relative_to(REPO_ROOT)), f"{artifact_id} path drifted")
        require(row.get("status") == "materialized_static", f"{artifact_id} status drifted")
        require(row.get("required") is True, f"{artifact_id} must remain required")
        require(path.exists(), f"{path.relative_to(REPO_ROOT)} must exist")


def assert_manifest_artifact() -> None:
    manifest = load_json(MANIFEST_PATH)
    require(manifest.get("artifact_kind") == "workflow_saved_draft_schema_artifact_manifest", "manifest kind drifted")
    require(manifest.get("schema_artifact_id") == "workflow_saved_draft_schema_artifact_v1", "artifact id drifted")
    require(manifest.get("manifest_version") == "workflow_saved_draft_schema_manifest_v1", "manifest version drifted")
    require(manifest.get("store_schema_version") == "saved_workflow_drafts_store_v1", "store schema version drifted")

    logical = manifest.get("logical_entity") or {}
    schema_contract = load_json(SCHEMA_PRECONDITIONS_FIXTURE_PATH).get("logical_schema_contract") or {}
    require(logical.get("id") == schema_contract.get("logical_entity"), "logical entity drifted")
    require(logical.get("payload_schema_must_not_replace_store_schema") is True, "schema version split must remain explicit")

    fields = {
        str(field.get("logical_field") or ""): field
        for field in manifest.get("field_mapping") or []
        if isinstance(field, dict)
    }
    require(set(fields) == schema_expected_fields(), "manifest field mapping drifted")
    for logical_field, field in fields.items():
        require(field.get("column_name") == logical_field, f"{logical_field} column drifted")
        require(field.get("secret_material_allowed") is False, f"{logical_field} must not allow secret material")
    forbidden = set(schema_contract.get("forbidden_persisted_fields") or [])
    require(set(manifest.get("forbidden_persisted_fields") or []) == forbidden, "forbidden persisted fields drifted")

    indexes = {
        str(index.get("id") or ""): index
        for index in manifest.get("index_mapping") or []
        if isinstance(index, dict)
    }
    expected_indexes = schema_expected_indexes()
    require(set(indexes) == set(expected_indexes), "manifest index ids drifted")
    for index_id, index in indexes.items():
        expected = expected_indexes[index_id]
        require(index.get("fields") == expected.get("fields"), f"{index_id} fields drifted")
        require(index.get("unique") == expected.get("unique"), f"{index_id} uniqueness drifted")
        require(index.get("fields", [None])[0] == "tenant_ref", f"{index_id} must lead with tenant_ref")

    plan = manifest.get("migration_plan") or {}
    require(plan.get("manual_apply_required") is True, "manual apply must be required")
    require(plan.get("startup_auto_migration_allowed") is False, "startup auto migration must remain disabled")
    require(plan.get("app_runtime_role_can_migrate") is False, "runtime role must not migrate")
    require(plan.get("sql_migration_created") is False, "SQL migration must not be created")
    require(plan.get("migration_runner_created") is False, "migration runner must not be created")
    require(plan.get("database_connection_allowed") is False, "database connection must not be allowed")

    refs = manifest.get("evidence_refs") or {}
    require(refs.get("ddl_review") == str(DDL_REVIEW_PATH.relative_to(REPO_ROOT)), "DDL review ref drifted")
    require(refs.get("rollback_evidence") == str(ROLLBACK_PATH.relative_to(REPO_ROOT)), "rollback ref drifted")
    require(refs.get("migration_smoke") == str(MIGRATION_SMOKE_PATH.relative_to(REPO_ROOT)), "smoke ref drifted")

    rows = {
        str(row.get("operation") or ""): row
        for row in manifest.get("operation_predicate_coverage") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_OPERATIONS, "manifest operation coverage drifted")
    for operation, row in rows.items():
        require(set(row.get("required_indexes") or []).issubset(set(expected_indexes)), f"{operation} indexes drifted")
        for field in (
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")
    require(EXPECTED_FAILURE_CODES.issuperset(set(manifest.get("failure_mapping_smoke_cases") or [])), "unexpected failure code")
    require(EXPECTED_FAILURE_CODES - set(manifest.get("failure_mapping_smoke_cases") or []) == set(), "missing failure code")
    require(EXPECTED_PRESERVED_FAILURE_CODES == set(manifest.get("preserve_existing_failure_codes") or []), "preserved failures drifted")
    require(EXPECTED_SIDE_EFFECT_COUNTERS == set(manifest.get("no_side_effect_counters") or []), "side effect counters drifted")


def assert_ddl_review_artifact() -> None:
    content = DDL_REVIEW_PATH.read_text(encoding="utf-8")
    for needle in (
        "static_schema_artifact_reviewed_no_sql",
        "SQL migration created: `false`",
        "database connection allowed: `false`",
        "migration runner created: `false`",
        "repository adapter created: `false`",
        "Service startup must not run schema migration automatically",
        "draft_schema_migration_not_applied",
        "repository_write_count=0",
        "database_write_count=0",
    ):
        require(needle in content, f"DDL review missing: {needle}")


def assert_rollback_and_smoke_artifacts() -> None:
    rollback = load_json(ROLLBACK_PATH)
    require(rollback.get("status") == "static_rollback_evidence_recorded_no_sql", "rollback status drifted")
    scope = rollback.get("rollback_scope") or {}
    for field in ("sql_migration_created", "database_mutation_applied", "migration_runner_created", "rollback_command_available"):
        require(scope.get(field) is False, f"rollback {field} must remain false")
    require(EXPECTED_SIDE_EFFECT_COUNTERS == set(rollback.get("no_side_effect_counters") or []), "rollback counters drifted")

    smoke = load_json(MIGRATION_SMOKE_PATH)
    require(smoke.get("status") == "static_migration_smoke_recorded_no_database", "smoke status drifted")
    require(smoke.get("database_connection_allowed") is False, "smoke must not allow database")
    require(smoke.get("sql_migration_created") is False, "smoke must not create SQL")
    require(smoke.get("migration_runner_created") is False, "smoke must not create runner")
    cases = {
        str(case.get("id") or ""): case
        for case in smoke.get("smoke_cases") or []
        if isinstance(case, dict)
    }
    require(set(cases) == EXPECTED_SMOKE_CASES, "migration smoke cases drifted")
    no_side_effects = set(cases["no_side_effects"].get("must_cover") or [])
    require(EXPECTED_SIDE_EFFECT_COUNTERS.issubset(no_side_effects), "smoke side-effect counters drifted")
    operation_rows = {
        str(row.get("operation") or ""): row
        for row in smoke.get("operation_smoke_matrix") or []
        if isinstance(row, dict)
    }
    require(set(operation_rows) == EXPECTED_OPERATIONS, "smoke operation matrix drifted")
    for operation, row in operation_rows.items():
        require(EXPECTED_FAILURE_CODES.issuperset(set(row.get("required_failure_codes") or [])), f"{operation} failure drifted")
        for field in (
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")


def assert_fixture_matrices(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("materialization_gate_matrix") or []
        if isinstance(gate, dict)
    }
    for gate_id in (
        "schema_artifact_evidence_consumed",
        "schema_artifact_manifest_consumed",
        "materialization_review_consumed",
        "selector_implementation_consumed",
        "schema_artifact_materialized",
    ):
        require(gates.get(gate_id, {}).get("status") == "satisfied", f"{gate_id} must be satisfied")
    for gate_id in (
        "repository_adapter_gate",
        "production_auth_gate",
        "database_execution_gate",
        "adapter_smoke_execution_gate",
    ):
        require(gates.get(gate_id, {}).get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")

    operations = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_materialization_matrix") or []
        if isinstance(row, dict)
    }
    require(set(operations) == EXPECTED_OPERATIONS, "operation materialization matrix drifted")
    expected_indexes = schema_expected_indexes()
    for operation, row in operations.items():
        require(set(row.get("required_indexes") or []).issubset(set(expected_indexes)), f"{operation} unknown indexes")
        for field in (
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")

    mapping = fixture.get("failure_mapping") or {}
    require(mapping.get("status") == "schema_artifact_materialization_fail_closed_static", "failure mapping status drifted")
    require(EXPECTED_FAILURE_CODES == set(mapping.get("required_failure_codes") or []), "failure mapping drifted")
    require(EXPECTED_PRESERVED_FAILURE_CODES == set(mapping.get("preserve_existing_failure_codes") or []), "preserved failure drifted")
    for field in (
        "missing_schema_artifact_success_allowed",
        "fallback_to_memory_dev_store_allowed",
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_dev_http_route_allowed",
        "database_detail_leak_allowed",
    ):
        require(mapping.get(field) is False, f"{field} must remain false")
    side_effects = fixture.get("no_side_effect_policy") or {}
    require(EXPECTED_SIDE_EFFECT_COUNTERS == set(side_effects.get("side_effect_counters_must_remain") or []), "fixture counters drifted")


def assert_no_forbidden_runtime_artifacts() -> None:
    sql_files = list((REPO_ROOT / "services/platform").rglob("*.sql"))
    require(not sql_files, f"SQL files must not be introduced: {sql_files}")
    for relative_path in FORBIDDEN_FILES:
        if repository_adapter_implementation_file_allowed(REPO_ROOT, relative_path):
            continue
        require(not (REPO_ROOT / relative_path).exists(), f"forbidden artifact exists: {relative_path}")


def assert_required_docs_and_testing(fixture: dict[str, Any]) -> None:
    references = fixture.get("required_doc_references") or []
    require(references, "required doc references must be declared")
    for reference in references:
        path = str(reference.get("path") or "")
        content = read(path)
        for needle in reference.get("must_contain") or []:
            require(str(needle) in content, f"{path} missing required text: {needle}")
    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "static_artifact_checker", "testing strategy status drifted")
    required = set(strategy.get("required_checks") or [])
    for expected in (
        "workflow_saved_draft_schema_artifact_materialization_checker",
        "workflow_saved_draft_schema_artifact_materialization_review_checker",
        "workflow_saved_draft_schema_artifact_manifest_checker",
        "workflow_saved_draft_schema_artifact_evidence_checker",
        "workflow_saved_draft_adapter_smoke_readiness_checker",
        "workflow_saved_draft_repository_adapter_implementation_plan_checker",
        "./scripts/check-repo.sh --fast",
        "./scripts/check-repo.sh",
    ):
        require(expected in required, f"missing required check: {expected}")
    for field in (
        "does_not_create_sql_or_migration",
        "does_not_create_migration_runner",
        "does_not_connect_database",
        "does_not_create_repository_interface",
        "does_not_create_repository_adapter",
        "does_not_enable_repository_mode",
        "does_not_call_oidc",
        "does_not_create_production_api",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-review-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "materialization review checker missing from check-repo.py")
    require(current_checker in check_repo, "materialization checker missing from check-repo.py")
    require(next_checker in check_repo, "product surface recheck missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "materialization checker must run after review and before product surface recheck",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_artifact_matrix(fixture)
    assert_manifest_artifact()
    assert_ddl_review_artifact()
    assert_rollback_and_smoke_artifacts()
    assert_fixture_matrices(fixture)
    assert_no_forbidden_runtime_artifacts()
    assert_required_docs_and_testing(fixture)
    assert_check_repo_registration()
    print("workflow saved draft schema artifact materialization v1 checks passed.")


if __name__ == "__main__":
    main()
