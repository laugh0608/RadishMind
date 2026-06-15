#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-migration-preconditions-v1.json"
)
DURABLE_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-durable-store-preconditions-v1.json"
)
REPOSITORY_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-preconditions-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
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
EXPECTED_SCOPE_FIELDS = {"tenant_ref", "workspace_id", "application_id", "draft_id"}
EXPECTED_OWNER_FIELDS = {"owner_subject_ref", "created_by_actor_ref", "updated_by_actor_ref"}
EXPECTED_VERSION_FIELDS = {
    "source_definition_id",
    "base_definition_version",
    "draft_version",
    "schema_version",
    "store_schema_version",
}
EXPECTED_PAYLOAD_FIELDS = {
    "sanitized_draft_payload",
    "validation_summary",
    "blocked_capability_summary",
}
EXPECTED_AUDIT_FIELDS = {"request_id", "audit_ref", "created_at", "updated_at"}
EXPECTED_FORBIDDEN_PERSISTED_FIELDS = {
    "secret_value",
    "api_key_value",
    "oauth_token",
    "authorization_header",
    "cookie_value",
    "raw_prompt_dump",
    "raw_tool_payload",
    "provider_response_body",
    "execution_plan_persistence",
    "runtime_readiness_persistence",
    "confirmation_decision",
    "business_writeback_payload",
    "run_result",
    "replay_state",
    "resume_state",
}
EXPECTED_INDEXES = {
    "saved_drafts_scope_lookup": ["tenant_ref", "workspace_id", "application_id", "draft_id"],
    "saved_drafts_owner_list": [
        "tenant_ref",
        "workspace_id",
        "application_id",
        "owner_subject_ref",
        "updated_at",
        "draft_id",
    ],
    "saved_drafts_version_conflict": [
        "tenant_ref",
        "workspace_id",
        "application_id",
        "draft_id",
        "draft_version",
    ],
    "saved_drafts_status_list": [
        "tenant_ref",
        "workspace_id",
        "application_id",
        "owner_subject_ref",
        "draft_status",
        "updated_at",
        "draft_id",
    ],
    "saved_drafts_schema_version": ["tenant_ref", "store_schema_version", "schema_version"],
}
EXPECTED_FAILURE_CODES = {
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_store_contract_mismatch",
    "repository_store_disabled",
    "invalid_draft_store_mode",
}
EXPECTED_EXISTING_FAILURE_CODES = {
    "draft_scope_denied",
    "draft_not_found",
    "draft_schema_version_unsupported",
    "draft_payload_invalid",
    "draft_version_conflict",
    "draft_store_unavailable",
}
EXPECTED_MIGRATION_SMOKE_COVERAGE = {
    "store schema version",
    "tenant workspace application predicate",
    "owner list projection",
    "version conflict predicate",
    "failure mapping",
    "no sample fallback",
    "no side effects",
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


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "workflow-saved-draft-durable-store-preconditions-v1": (
            DURABLE_PRECONDITIONS_FIXTURE_PATH,
            "draft_durable_store_preconditions_defined",
        ),
        "workflow-saved-draft-repository-contract-preconditions-v1": (
            REPOSITORY_PRECONDITIONS_FIXTURE_PATH,
            "draft_repository_contract_preconditions_defined",
        ),
    }
    missing = sorted(set(expected) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for expected_id, (path, expected_status) in expected.items():
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
        fixture.get("kind") == "workflow_saved_draft_schema_migration_preconditions_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-schema-migration-preconditions-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_schema_migration_preconditions_defined",
        "schema migration preconditions status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_schema_migration_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("schema_migration_boundary") or {}
    repository = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH).get("repository_contract_boundary") or {}
    require(
        boundary.get("status") == "schema_migration_preconditions_defined_not_implemented",
        "schema migration boundary status drifted",
    )
    require(boundary.get("current_store_source") == "platform memory dev store", "current store source drifted")
    require(
        boundary.get("future_store_source") == repository.get("future_store_source"),
        "future store source must match repository preconditions",
    )
    for field in (
        "logical_schema_defined_in_this_slice",
        "index_strategy_defined_in_this_slice",
        "migration_gate_defined_in_this_slice",
    ):
        require(boundary.get(field) is True, f"{field} must remain true")
    for field in (
        "schema_files_created_in_this_slice",
        "migration_files_created_in_this_slice",
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


def assert_logical_schema_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("logical_schema_contract") or {}
    require(
        contract.get("status") == "logical_schema_defined_for_future_artifact",
        "logical schema status drifted",
    )
    require(contract.get("logical_entity") == "saved_workflow_draft_record", "logical entity drifted")
    checks = (
        ("scope_identity_fields", EXPECTED_SCOPE_FIELDS),
        ("ownership_actor_fields", EXPECTED_OWNER_FIELDS),
        ("source_version_fields", EXPECTED_VERSION_FIELDS),
        ("payload_fields", EXPECTED_PAYLOAD_FIELDS),
        ("audit_fields", EXPECTED_AUDIT_FIELDS),
    )
    for key, expected_fields in checks:
        missing = sorted(expected_fields - set(contract.get(key) or []))
        require(not missing, f"{key} missing fields: {missing}")
    require(set(contract.get("draft_id_scope") or []) == EXPECTED_SCOPE_FIELDS, "draft id scope drifted")
    require(contract.get("payload_schema_version_field") == "schema_version", "payload schema field drifted")
    require(contract.get("store_schema_version_field") == "store_schema_version", "store schema field drifted")
    require(
        contract.get("payload_schema_must_not_replace_store_schema") is True,
        "payload schema must not replace store schema",
    )
    missing_forbidden = sorted(
        EXPECTED_FORBIDDEN_PERSISTED_FIELDS - set(contract.get("forbidden_persisted_fields") or [])
    )
    require(not missing_forbidden, f"missing forbidden persisted fields: {missing_forbidden}")


def assert_index_strategy(fixture: dict[str, Any]) -> None:
    strategy = fixture.get("index_strategy") or {}
    require(strategy.get("status") == "defined_not_created", "index strategy status drifted")
    require(
        strategy.get("tenant_workspace_application_predicate_required") is True,
        "tenant/workspace/application predicate must be required",
    )
    require(strategy.get("query_string_scope_override_allowed") is False, "scope override must remain false")
    indexes = {
        str(index.get("id")): index
        for index in strategy.get("indexes") or []
        if isinstance(index, dict)
    }
    require(set(indexes) == set(EXPECTED_INDEXES), "index ids drifted")
    for index_id, expected_fields in EXPECTED_INDEXES.items():
        index = indexes[index_id]
        require(index.get("fields") == expected_fields, f"{index_id} fields drifted")
        if index_id == "saved_drafts_scope_lookup":
            require(index.get("unique") is True, "scope lookup must remain unique")
        else:
            require(index.get("unique") is False, f"{index_id} must not be unique")


def assert_migration_layout(fixture: dict[str, Any]) -> None:
    layout = fixture.get("migration_layout_reservation") or {}
    require(layout.get("status") == "planned_not_created", "migration layout status drifted")
    require(
        layout.get("migration_root") == "services/platform/migrations/workflow_saved_drafts",
        "migration root drifted",
    )
    require(
        layout.get("migration_naming") == "YYYYMMDDHHMMSS_workflow_saved_drafts_<short_slug>.sql",
        "migration naming drifted",
    )
    planned_files = layout.get("planned_files") or []
    require(planned_files, "planned migration files must be documented")
    for planned in planned_files:
        require(planned.get("created_in_this_slice") is False, "planned migration files must not be created")


def assert_migration_gate_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("migration_gate_policy") or {}
    require(policy.get("status") == "policy_defined_not_implemented", "migration policy status drifted")
    for field in (
        "schema_artifact_manifest_required",
        "ddl_review_evidence_required",
        "rollback_evidence_required",
        "backup_required_before_apply",
        "migration_lock_required",
        "schema_version_table_required",
        "manual_apply_gate_required",
        "migration_smoke_required",
    ):
        require(policy.get(field) is True, f"{field} must remain true")
    for field in (
        "auto_migrate_on_service_start_allowed",
        "destructive_migration_allowed_without_review",
        "app_runtime_role_can_migrate",
    ):
        require(policy.get(field) is False, f"{field} must remain false")
    missing = sorted(EXPECTED_MIGRATION_SMOKE_COVERAGE - set(policy.get("migration_smoke_must_cover") or []))
    require(not missing, f"migration smoke coverage missing: {missing}")


def assert_failure_mapping(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(mapping.get("status") == "fail_closed_required", "failure mapping status drifted")
    missing = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing, f"missing required failure codes: {missing}")
    missing_existing = sorted(
        EXPECTED_EXISTING_FAILURE_CODES - set(mapping.get("preserve_existing_failure_codes") or [])
    )
    require(not missing_existing, f"missing preserved failure codes: {missing_existing}")
    require(
        mapping.get("version_conflict_must_not_be_rewritten_as_migration_failure") is True,
        "version conflict must be preserved",
    )
    for field in (
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_memory_dev_store_allowed",
    ):
        require(mapping.get(field) is False, f"{field} must remain false")


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
        "workflow_saved_draft_schema_migration_preconditions_checker" in required,
        "missing checker in testing strategy",
    )
    require("./scripts/check-repo.sh --fast" in required, "missing fast baseline in testing strategy")
    for field in (
        "does_not_start_service",
        "does_not_connect_database",
        "does_not_create_migration_files",
        "does_not_call_oidc",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-repository-contract-preconditions-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-schema-migration-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run schema migration preconditions check")
    require(previous_checker in check_repo, "repository contract preconditions checker missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "schema migration preconditions check must run after repository contract preconditions",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_schema_migration_boundary(fixture)
    assert_logical_schema_contract(fixture)
    assert_index_strategy(fixture)
    assert_migration_layout(fixture)
    assert_migration_gate_policy(fixture)
    assert_failure_mapping(fixture)
    assert_required_doc_references(fixture)
    assert_artifact_guard(fixture)
    assert_testing_strategy(fixture)
    assert_check_repo_registration()
    print("workflow saved draft schema migration preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
