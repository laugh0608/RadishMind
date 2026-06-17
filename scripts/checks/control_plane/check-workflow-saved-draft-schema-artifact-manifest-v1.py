#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from workflow_saved_draft_selector_implementation_guard import (
    selector_implementation_file_allowed,
    selector_implementation_literal_allowed,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-manifest-v1.json"
)
SCHEMA_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-migration-preconditions-v1.json"
)
SCHEMA_EVIDENCE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-evidence-v1.json"
)
SELECTOR_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-readiness-v1.json"
)
RUNNER_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.json"
)
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-schema-migration-preconditions-v1": (
        SCHEMA_PRECONDITIONS_FIXTURE_PATH,
        "draft_schema_migration_preconditions_defined",
    ),
    "workflow-saved-draft-schema-artifact-evidence-v1": (
        SCHEMA_EVIDENCE_FIXTURE_PATH,
        "draft_schema_artifact_evidence_defined",
    ),
    "workflow-saved-draft-store-selector-smoke-readiness-v1": (
        SELECTOR_SMOKE_FIXTURE_PATH,
        "draft_store_selector_smoke_readiness_defined",
    ),
    "workflow-saved-draft-repository-contract-smoke-runner-implementation-v1": (
        RUNNER_IMPLEMENTATION_FIXTURE_PATH,
        "draft_repository_contract_smoke_runner_implemented",
    ),
    "workflow-saved-draft-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "draft_repository_adapter_implementation_plan_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "schema_artifact_manifest_file_created",
    "schema_artifact_files_created",
    "ddl_review_artifact_created",
    "rollback_evidence_artifact_created",
    "migration_smoke_artifact_created",
    "schema_version_table_created",
    "sql_migration_created",
    "migration_runner_implemented",
    "database_schema_ready",
    "database_connection_ready",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "store_selector_implemented",
    "selector_smoke_fixture_created",
    "radish_oidc_ready",
    "token_validation_ready",
    "production_api_consumer_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_MANIFEST_FIELDS = {
    "manifest id",
    "manifest version",
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
    "operation predicate coverage",
    "failure mapping smoke cases",
    "no side effect counters",
}
EXPECTED_EVIDENCE_PATHS = {
    "services/platform/migrations/workflow_saved_drafts/manifest.json",
    "services/platform/migrations/workflow_saved_drafts/ddl-review.md",
    "services/platform/migrations/workflow_saved_drafts/rollback-evidence.json",
    "services/platform/migrations/workflow_saved_drafts/migration-smoke.json",
}
EXPECTED_SECTION_IDS = {
    "identity",
    "logical_entity",
    "field_mapping",
    "index_mapping",
    "migration_plan",
    "ddl_review_gate",
    "rollback_gate",
    "migration_smoke_gate",
    "operation_predicates",
    "failure_mapping",
    "no_side_effects",
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
    "draft_store_unavailable",
    "draft_not_found",
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


def schema_expected_indexes() -> dict[str, dict[str, Any]]:
    strategy = load_json(SCHEMA_PRECONDITIONS_FIXTURE_PATH).get("index_strategy") or {}
    return {
        str(index.get("id") or ""): index
        for index in strategy.get("indexes") or []
        if isinstance(index, dict)
    }


def schema_evidence_boundary() -> dict[str, Any]:
    return load_json(SCHEMA_EVIDENCE_FIXTURE_PATH).get("schema_artifact_evidence_boundary") or {}


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
        fixture.get("kind") == "workflow_saved_draft_schema_artifact_manifest_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-schema-artifact-manifest-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_schema_artifact_manifest_defined",
        "schema artifact manifest status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_manifest_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("schema_artifact_manifest_boundary") or {}
    evidence_boundary = schema_evidence_boundary()
    adapter_layout = {
        str(item.get("path") or "")
        for item in load_json(ADAPTER_PLAN_FIXTURE_PATH).get("future_file_layout") or []
        if isinstance(item, dict)
    }
    require(boundary.get("status") == "manifest_contract_defined_no_files", "boundary status drifted")
    require(
        boundary.get("decision") == "manifest_shape_and_materialization_gates_defined_without_schema_artifact_files",
        "boundary decision drifted",
    )
    for field in (
        "current_store_source",
        "future_store_source",
        "future_schema_artifact_id",
        "future_store_schema_version",
        "future_migration_root",
        "future_manifest_file",
        "future_ddl_review_file",
        "future_schema_version_table",
    ):
        require(boundary.get(field) == evidence_boundary.get(field), f"{field} must match schema evidence")
    require(
        boundary.get("future_migration_root") in adapter_layout,
        "future migration root must remain listed in adapter plan layout",
    )
    require(
        boundary.get("future_manifest_version") == "workflow_saved_draft_schema_manifest_v1",
        "manifest version drifted",
    )
    for field in (
        "schema_artifact_manifest_file_created_in_this_slice",
        "schema_artifact_files_created_in_this_slice",
        "ddl_review_artifact_created_in_this_slice",
        "rollback_evidence_artifact_created_in_this_slice",
        "migration_smoke_artifact_created_in_this_slice",
        "schema_version_table_created_in_this_slice",
        "migration_root_created_in_this_slice",
        "sql_files_created_in_this_slice",
        "migration_runner_created_in_this_slice",
        "database_connection_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_created_in_this_slice",
        "store_selector_created_in_this_slice",
        "selector_smoke_created_in_this_slice",
        "oidc_validation_created_in_this_slice",
        "production_api_consumer_created_in_this_slice",
        "publish_or_run_created_in_this_slice",
        "executor_created_in_this_slice",
        "writeback_or_replay_created_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_manifest_shape_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("manifest_shape_contract") or {}
    boundary = fixture.get("schema_artifact_manifest_boundary") or {}
    require(contract.get("status") == "shape_defined_not_materialized", "manifest shape status drifted")
    require(contract.get("required_before_manifest_file") is True, "manifest shape must be required first")
    require(contract.get("future_manifest_path") == boundary.get("future_manifest_file"), "manifest path drifted")
    missing_fields = sorted(EXPECTED_MANIFEST_FIELDS - set(contract.get("manifest_must_include") or []))
    require(not missing_fields, f"missing manifest fields: {missing_fields}")

    artifacts = {
        str(item.get("path") or ""): item
        for item in contract.get("planned_evidence_artifacts") or []
        if isinstance(item, dict)
    }
    require(set(artifacts) == EXPECTED_EVIDENCE_PATHS, "planned evidence artifacts drifted")
    for relative_path, artifact in artifacts.items():
        require(artifact.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist yet")

    migration_root = REPO_ROOT / str(boundary.get("future_migration_root") or "")
    require(not migration_root.exists(), "migration root must not be created in this slice")
    sql_files = list((REPO_ROOT / "services/platform").rglob("*.sql"))
    require(not sql_files, f"SQL files must not be introduced in this slice: {sql_files}")


def assert_manifest_section_matrix(fixture: dict[str, Any]) -> None:
    sections = {
        str(section.get("id") or ""): section
        for section in fixture.get("manifest_section_matrix") or []
        if isinstance(section, dict)
    }
    require(set(sections) == EXPECTED_SECTION_IDS, "manifest section ids drifted")
    for section_id, section in sections.items():
        require(section.get("status") == "defined_not_materialized", f"{section_id} status drifted")
        require(section.get("must_cover"), f"{section_id} must list coverage")
        require(section.get("artifact_created_now") is False, f"{section_id} artifact must not be created")

    schema_indexes = set(schema_expected_indexes())
    index_coverage = set(sections["index_mapping"].get("must_cover") or [])
    missing_indexes = sorted(schema_indexes - index_coverage)
    require(not missing_indexes, f"index section missing indexes: {missing_indexes}")

    smoke_cases = {
        str(case.get("id") or "")
        for case in (load_json(SCHEMA_EVIDENCE_FIXTURE_PATH).get("migration_smoke_evidence") or {}).get("smoke_cases")
        or []
        if isinstance(case, dict)
    }
    smoke_coverage = set(sections["migration_smoke_gate"].get("must_cover") or [])
    missing_smoke = sorted(smoke_cases - smoke_coverage)
    require(not missing_smoke, f"migration smoke section missing cases: {missing_smoke}")


def assert_operation_manifest_matrix(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_manifest_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_OPERATIONS, "operation manifest matrix must cover save/read/list")

    adapter_rows = {
        str(row.get("operation") or ""): row
        for row in load_json(ADAPTER_PLAN_FIXTURE_PATH).get("operation_adapter_matrix") or []
        if isinstance(row, dict)
    }
    runner_rows = {
        str(row.get("operation") or ""): row
        for row in load_json(RUNNER_IMPLEMENTATION_FIXTURE_PATH).get("operation_runner_matrix") or []
        if isinstance(row, dict)
    }
    schema_indexes = schema_expected_indexes()
    for operation, row in rows.items():
        require(row.get("required_scope") == adapter_rows[operation].get("required_scope"), f"{operation} scope drifted")
        require(row.get("required_scope") == runner_rows[operation].get("required_scope"), f"{operation} runner scope drifted")
        required_indexes = set(row.get("required_indexes") or [])
        require(required_indexes.issubset(set(schema_indexes)), f"{operation} references unknown indexes")
        for field in (
            "schema_artifact_manifest_required",
            "ddl_review_required",
            "rollback_evidence_required",
            "migration_smoke_required",
        ):
            require(row.get(field) is True, f"{operation} {field} must remain true")
        for field in (
            "manifest_file_created_now",
            "sql_migration_allowed_now",
            "adapter_implementation_allowed_now",
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")


def assert_failure_fallback_and_side_effects(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(mapping.get("status") == "fail_closed_required", "failure mapping status drifted")
    missing = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing, f"missing failure codes: {missing}")
    missing_preserved = sorted(EXPECTED_PRESERVED_FAILURE_CODES - set(mapping.get("preserve_existing_failure_codes") or []))
    require(not missing_preserved, f"missing preserved failure codes: {missing_preserved}")
    for field in (
        "missing_manifest_success_allowed",
        "missing_ddl_review_success_allowed",
        "missing_rollback_evidence_success_allowed",
        "migration_smoke_failure_success_allowed",
        "database_detail_leak_allowed",
    ):
        require(mapping.get(field) is False, f"{field} must remain false")
    for field in (
        "version_conflict_must_not_be_rewritten_as_migration_failure",
        "not_found_must_not_fallback_to_sample",
        "scope_denied_must_not_return_draft_payload",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

    fallback = fixture.get("no_fallback_policy") or {}
    for field in (
        "fallback_to_memory_dev_store_allowed",
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_dev_http_route_allowed",
        "fallback_to_test_auth_allowed",
        "success_on_missing_manifest_allowed",
        "success_on_contract_mismatch_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
    missing_counters = sorted(
        EXPECTED_SIDE_EFFECT_COUNTERS - set(side_effects.get("side_effect_counters_must_remain") or [])
    )
    require(not missing_counters, f"missing side effect counters: {missing_counters}")
    missing_effects = sorted(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS - set(side_effects.get("forbidden_side_effects") or [])
    )
    require(not missing_effects, f"missing forbidden side effects: {missing_effects}")


def assert_required_docs_and_testing(fixture: dict[str, Any]) -> None:
    references = fixture.get("required_doc_references") or []
    require(references, "required doc references must be declared")
    for reference in references:
        path = str(reference.get("path") or "")
        require(path, "doc reference path missing")
        content = read(path)
        for needle in reference.get("must_contain") or []:
            require(str(needle) in content, f"{path} missing required text: {needle}")

    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "static_checker_only", "testing strategy status drifted")
    required = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_schema_artifact_manifest_checker",
        "workflow_saved_draft_schema_artifact_evidence_checker",
        "workflow_saved_draft_repository_adapter_implementation_plan_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_create_migration_root",
        "does_not_create_schema_artifact_file",
        "does_not_create_sql_migration",
        "does_not_connect_database",
        "does_not_create_repository_adapter",
        "does_not_create_store_selector",
        "does_not_call_oidc",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("implementation_artifact_guard") or {}
    require(guard.get("status") == "forbid_implementation_artifacts", "artifact guard status drifted")
    for relative_path in guard.get("future_files_must_not_exist") or []:
        if selector_implementation_file_allowed(REPO_ROOT, str(relative_path)):
            continue
        require(not (REPO_ROOT / str(relative_path)).exists(), f"future artifact exists early: {relative_path}")
    source_paths = guard.get("source_files_to_scan") or []
    literals = guard.get("future_literals_must_not_appear_in_source") or []
    require(source_paths, "source files to scan must be declared")
    require(literals, "future literals must be declared")
    for source_path in source_paths:
        source = read(str(source_path))
        for literal in literals:
            if selector_implementation_literal_allowed(REPO_ROOT, str(literal)):
                continue
            require(str(literal) not in source, f"{source_path} contains future literal: {literal}")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = (
        "checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-plan-v1.py"
    )
    current_checker = "checks/control_plane/check-workflow-saved-draft-schema-artifact-manifest-v1.py"
    require(previous_checker in check_repo, "repository adapter plan checker missing from check-repo.py")
    require(current_checker in check_repo, "check-repo.py must run schema artifact manifest check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "schema artifact manifest check must run after repository adapter plan",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_manifest_boundary(fixture)
    assert_manifest_shape_contract(fixture)
    assert_manifest_section_matrix(fixture)
    assert_operation_manifest_matrix(fixture)
    assert_failure_fallback_and_side_effects(fixture)
    assert_required_docs_and_testing(fixture)
    assert_artifact_guard(fixture)
    assert_check_repo_registration()
    print("workflow saved draft schema artifact manifest v1 checks passed.")


if __name__ == "__main__":
    main()
