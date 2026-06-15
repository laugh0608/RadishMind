#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-evidence-v1.json"
)
REPOSITORY_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-preconditions-v1.json"
)
SCHEMA_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-migration-preconditions-v1.json"
)
AUTH_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-auth-context-preconditions-v1.json"
)
SELECTOR_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-enablement-preconditions-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-repository-contract-preconditions-v1": (
        REPOSITORY_PRECONDITIONS_FIXTURE_PATH,
        "draft_repository_contract_preconditions_defined",
    ),
    "workflow-saved-draft-schema-migration-preconditions-v1": (
        SCHEMA_PRECONDITIONS_FIXTURE_PATH,
        "draft_schema_migration_preconditions_defined",
    ),
    "workflow-saved-draft-auth-context-preconditions-v1": (
        AUTH_PRECONDITIONS_FIXTURE_PATH,
        "draft_auth_context_preconditions_defined",
    ),
    "workflow-saved-draft-store-selector-enablement-preconditions-v1": (
        SELECTOR_PRECONDITIONS_FIXTURE_PATH,
        "draft_store_selector_enablement_preconditions_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "schema_artifact_manifest_ready",
    "schema_artifact_files_created",
    "ddl_review_ready",
    "ddl_review_artifact_created",
    "rollback_evidence_ready",
    "migration_smoke_ready",
    "schema_version_table_ready",
    "database_schema_ready",
    "database_migration_ready",
    "migration_files_created",
    "migration_runner_implemented",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "store_selector_implemented",
    "radish_oidc_ready",
    "token_validation_ready",
    "production_api_consumer_ready",
    "saved_draft_list_implemented",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_MANIFEST_FIELDS = {
    "manifest id",
    "schema artifact id",
    "store schema version",
    "logical entity",
    "field mapping",
    "index mapping",
    "migration id",
    "up migration path",
    "down migration path or forward-only exception",
    "DDL review evidence ref",
    "rollback evidence ref",
    "migration smoke evidence ref",
    "schema version table",
    "manual apply gate",
    "migration lock",
    "backup requirement",
    "failure mapping smoke cases",
    "no side effect counters",
}
EXPECTED_EVIDENCE_PATHS = {
    "services/platform/migrations/workflow_saved_drafts/manifest.json",
    "services/platform/migrations/workflow_saved_drafts/ddl-review.md",
    "services/platform/migrations/workflow_saved_drafts/rollback-evidence.json",
    "services/platform/migrations/workflow_saved_drafts/migration-smoke.json",
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
    "draft_store_unavailable",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "database_write",
    "workflow_execution",
    "confirmation_decision",
    "business_writeback",
    "replay_execution",
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
    field_groups = (
        "scope_identity_fields",
        "ownership_actor_fields",
        "source_version_fields",
        "status_fields",
        "payload_fields",
        "audit_fields",
    )
    fields: set[str] = set()
    for group in field_groups:
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
        fixture.get("kind") == "workflow_saved_draft_schema_artifact_evidence_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-schema-artifact-evidence-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_schema_artifact_evidence_defined",
        "schema artifact evidence status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_schema_artifact_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("schema_artifact_evidence_boundary") or {}
    schema_boundary = load_json(SCHEMA_PRECONDITIONS_FIXTURE_PATH).get("schema_migration_boundary") or {}
    selector_boundary = load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH).get("selector_enablement_boundary") or {}
    require(
        boundary.get("status") == "schema_artifact_evidence_defined_no_artifact",
        "schema artifact boundary status drifted",
    )
    require(
        boundary.get("decision") == "evidence_chain_defined_without_schema_artifact",
        "schema artifact boundary decision drifted",
    )
    require(
        boundary.get("current_store_source") == schema_boundary.get("current_store_source"),
        "current store source must match schema preconditions",
    )
    require(
        boundary.get("future_store_source") == selector_boundary.get("future_store_source"),
        "future store source must match selector preconditions",
    )
    require(
        boundary.get("future_migration_root") == "services/platform/migrations/workflow_saved_drafts",
        "future migration root drifted",
    )
    require(
        boundary.get("future_manifest_file") == "services/platform/migrations/workflow_saved_drafts/manifest.json",
        "future manifest path drifted",
    )
    require(
        boundary.get("future_schema_version_table") == "workflow_saved_draft_schema_versions",
        "schema version table drifted",
    )
    for field in (
        "schema_artifact_manifest_created_in_this_slice",
        "schema_artifact_files_created_in_this_slice",
        "ddl_review_artifact_created_in_this_slice",
        "rollback_evidence_created_in_this_slice",
        "migration_smoke_created_in_this_slice",
        "schema_version_table_created_in_this_slice",
        "migration_root_created_in_this_slice",
        "sql_files_created_in_this_slice",
        "migration_runner_created_in_this_slice",
        "database_connection_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_created_in_this_slice",
        "store_selector_created_in_this_slice",
        "oidc_validation_created_in_this_slice",
        "production_api_consumer_created_in_this_slice",
        "saved_draft_list_created_in_this_slice",
        "publish_or_run_created_in_this_slice",
        "executor_created_in_this_slice",
        "writeback_or_replay_created_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_manifest_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("schema_artifact_manifest_contract") or {}
    boundary = fixture.get("schema_artifact_evidence_boundary") or {}
    require(contract.get("status") == "contract_defined_not_created", "manifest contract status drifted")
    require(contract.get("required_before_schema_artifact_files") is True, "manifest must be required first")
    require(
        contract.get("future_manifest_path") == boundary.get("future_manifest_file"),
        "manifest path must match boundary",
    )
    missing_fields = sorted(EXPECTED_MANIFEST_FIELDS - set(contract.get("manifest_must_include") or []))
    require(not missing_fields, f"missing manifest fields: {missing_fields}")
    artifacts = {
        str(artifact.get("path") or ""): artifact
        for artifact in contract.get("planned_evidence_artifacts") or []
        if isinstance(artifact, dict)
    }
    require(set(artifacts) == EXPECTED_EVIDENCE_PATHS, "planned evidence paths drifted")
    for path, artifact in artifacts.items():
        require(artifact.get("created_in_this_slice") is False, f"{path} must not be created")


def assert_logical_entity_mapping(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("logical_entity_artifact_mapping") or {}
    schema_contract = load_json(SCHEMA_PRECONDITIONS_FIXTURE_PATH).get("logical_schema_contract") or {}
    boundary = fixture.get("schema_artifact_evidence_boundary") or {}
    require(mapping.get("status") == "defined_not_materialized", "logical mapping status drifted")
    require(mapping.get("logical_entity") == schema_contract.get("logical_entity"), "logical entity drifted")
    require(
        mapping.get("future_schema_artifact_id") == boundary.get("future_schema_artifact_id"),
        "schema artifact id drifted",
    )
    require(
        mapping.get("store_schema_version") == boundary.get("future_store_schema_version"),
        "store schema version drifted",
    )
    columns = {
        str(column.get("logical_field") or ""): column
        for column in mapping.get("columns") or []
        if isinstance(column, dict)
    }
    require(set(columns) == schema_expected_fields(), "logical field mapping drifted")
    for logical_field, column in columns.items():
        require(column.get("column_name") == logical_field, f"{logical_field} column name drifted")
        require(column.get("column_type"), f"{logical_field} column type missing")
        require(column.get("secret_material_allowed") is False, f"{logical_field} must not allow secret material")
    forbidden = set(schema_contract.get("forbidden_persisted_fields") or [])
    mapped_columns = {str(column.get("column_name") or "") for column in columns.values()}
    require(not (forbidden & mapped_columns), "forbidden persisted fields leaked into column mapping")
    missing_forbidden = sorted(forbidden - set(mapping.get("forbidden_column_names") or []))
    require(not missing_forbidden, f"missing forbidden column names: {missing_forbidden}")


def assert_index_mapping(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("index_artifact_mapping") or {}
    require(mapping.get("status") == "defined_not_created", "index mapping status drifted")
    upstream_indexes = schema_expected_indexes()
    indexes = {
        str(index.get("id") or ""): index
        for index in mapping.get("indexes") or []
        if isinstance(index, dict)
    }
    require(set(indexes) == set(upstream_indexes), "index ids drifted")
    for index_id, index in indexes.items():
        upstream = upstream_indexes[index_id]
        fields = index.get("fields") or []
        require(fields == upstream.get("fields"), f"{index_id} fields drifted")
        require(index.get("unique") == upstream.get("unique"), f"{index_id} uniqueness drifted")
        require(fields and fields[0] == "tenant_ref", f"{index_id} must lead with tenant_ref")
        require(index.get("created_in_this_slice") is False, f"{index_id} must not be created")
        require(index.get("ddl_created_in_this_slice") is False, f"{index_id} DDL must not be created")


def assert_ddl_review_evidence(fixture: dict[str, Any]) -> None:
    ddl = fixture.get("ddl_review_evidence") or {}
    boundary = fixture.get("schema_artifact_evidence_boundary") or {}
    require(ddl.get("status") == "defined_not_materialized", "DDL review status drifted")
    require(ddl.get("future_artifact_ref") == boundary.get("future_ddl_review_file"), "DDL ref drifted")
    require(ddl.get("reviewer_requirement") == "human_review_required", "DDL review must require human review")
    for field in ("manual_apply_required",):
        require(ddl.get(field) is True, f"{field} must remain true")
    for field in (
        "startup_auto_migration_allowed",
        "destructive_change_without_review_allowed",
        "app_runtime_role_can_migrate",
        "artifact_created_now",
    ):
        require(ddl.get(field) is False, f"{field} must remain false")
    coverage = set(ddl.get("must_cover") or [])
    require(
        {
            "human-reviewed DDL",
            "manual apply command",
            "backup requirement",
            "migration lock",
            "schema version table",
            "failure mapping",
            "no sample fallback",
            "no side effects",
        }.issubset(coverage),
        "DDL review coverage drifted",
    )


def assert_migration_smoke_evidence(fixture: dict[str, Any]) -> None:
    smoke = fixture.get("migration_smoke_evidence") or {}
    require(smoke.get("status") == "defined_not_materialized", "migration smoke status drifted")
    require(smoke.get("artifact_created_now") is False, "migration smoke artifact must not be created")
    cases = {
        str(case.get("id") or ""): case
        for case in smoke.get("smoke_cases") or []
        if isinstance(case, dict)
    }
    require(set(cases) == EXPECTED_SMOKE_CASES, "migration smoke cases drifted")
    no_side_effects = set(cases["no_side_effects"].get("must_cover") or [])
    missing_counters = sorted(EXPECTED_SIDE_EFFECT_COUNTERS - no_side_effects)
    require(not missing_counters, f"missing no side effect counters: {missing_counters}")


def assert_failure_and_side_effect_policy(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(mapping.get("status") == "fail_closed_required", "failure mapping status drifted")
    missing = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing, f"missing failure codes: {missing}")
    missing_preserved = sorted(EXPECTED_PRESERVED_FAILURE_CODES - set(mapping.get("preserve_existing_failure_codes") or []))
    require(not missing_preserved, f"missing preserved failure codes: {missing_preserved}")
    for field in (
        "missing_ddl_review_success_allowed",
        "missing_schema_artifact_success_allowed",
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_memory_dev_store_allowed",
    ):
        require(mapping.get(field) is False, f"{field} must remain false")
    require(
        mapping.get("version_conflict_must_not_be_rewritten_as_migration_failure") is True,
        "version conflict must be preserved",
    )
    policy = fixture.get("no_side_effect_policy") or {}
    missing_counters = sorted(
        EXPECTED_SIDE_EFFECT_COUNTERS - set(policy.get("side_effect_counters_must_remain") or [])
    )
    require(not missing_counters, f"missing side effect counters: {missing_counters}")
    missing_effects = sorted(EXPECTED_FORBIDDEN_SIDE_EFFECTS - set(policy.get("forbidden_side_effects") or []))
    require(not missing_effects, f"missing forbidden side effects: {missing_effects}")


def assert_required_doc_references(fixture: dict[str, Any]) -> None:
    references = fixture.get("required_doc_references") or []
    require(references, "required doc references must be declared")
    for reference in references:
        path = str(reference.get("path") or "")
        require(path, "doc reference path missing")
        content = read(path)
        for needle in reference.get("must_contain") or []:
            require(str(needle) in content, f"{path} missing required text: {needle}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("implementation_artifact_guard") or {}
    require(guard.get("status") == "forbid_implementation_artifacts", "artifact guard status drifted")
    for relative_path in guard.get("future_files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"future artifact exists early: {relative_path}")
    source_paths = guard.get("source_files_to_scan") or []
    literals = guard.get("future_literals_must_not_appear_in_source") or []
    require(source_paths, "source files to scan must be declared")
    require(literals, "future literals must be declared")
    for source_path in source_paths:
        source = read(str(source_path))
        for literal in literals:
            require(str(literal) not in source, f"{source_path} contains future literal: {literal}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "static_checker_only", "testing strategy status drifted")
    required = set(strategy.get("required_checks") or [])
    require(
        "workflow_saved_draft_schema_artifact_evidence_checker" in required,
        "missing checker in testing strategy",
    )
    require("./scripts/check-repo.sh --fast" in required, "missing fast baseline in testing strategy")
    for field in (
        "does_not_start_service",
        "does_not_create_migration_root",
        "does_not_create_schema_artifact",
        "does_not_connect_database",
        "does_not_create_repository_adapter",
        "does_not_create_store_selector",
        "does_not_call_oidc",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-store-selector-enablement-preconditions-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-schema-artifact-evidence-v1.py"
    require(current_checker in check_repo, "check-repo.py must run schema artifact evidence check")
    require(previous_checker in check_repo, "store selector preconditions checker missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "schema artifact evidence check must run after store selector preconditions",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_schema_artifact_boundary(fixture)
    assert_manifest_contract(fixture)
    assert_logical_entity_mapping(fixture)
    assert_index_mapping(fixture)
    assert_ddl_review_evidence(fixture)
    assert_migration_smoke_evidence(fixture)
    assert_failure_and_side_effect_policy(fixture)
    assert_required_doc_references(fixture)
    assert_artifact_guard(fixture)
    assert_testing_strategy(fixture)
    assert_check_repo_registration()
    print("workflow saved draft schema artifact evidence v1 checks passed.")


if __name__ == "__main__":
    main()
