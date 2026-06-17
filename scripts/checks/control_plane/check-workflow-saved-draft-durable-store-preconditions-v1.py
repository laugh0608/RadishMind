#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from workflow_saved_draft_selector_implementation_guard import (
    selector_implementation_file_allowed,
    selector_implementation_literal_allowed,
)
from workflow_saved_draft_schema_materialization_guard import (
    schema_materialization_file_allowed,
    schema_materialization_literal_allowed,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-durable-store-preconditions-v1.json"
)
CONSUMER_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-consumer-smoke-v1.json"
CREATION_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/user-workspace-draft-creation-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "durable_store_implemented",
    "durable_persistence_ready",
    "database_ready",
    "database_schema_ready",
    "database_migration_ready",
    "repository_adapter_ready",
    "store_selector_implemented",
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
EXPECTED_SCOPE_FIELDS = {
    "tenant_ref",
    "workspace_id",
    "application_id",
    "draft_id",
    "source_definition_id",
    "base_definition_version",
    "draft_version",
    "schema_version",
}
EXPECTED_OWNER_FIELDS = {
    "owner_subject_ref",
    "created_by_actor_ref",
    "updated_by_actor_ref",
    "tenant_ref",
    "workspace_id",
    "application_id",
}
EXPECTED_NO_FALLBACK_FAILURE_CODES = {
    "draft_scope_denied",
    "draft_not_found",
    "draft_schema_version_unsupported",
    "draft_version_conflict",
    "draft_store_unavailable",
    "draft_write_disabled",
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_store_contract_mismatch",
}
EXPECTED_SWITCH_PREREQUISITES = {
    "SavedWorkflowDraftRepository contract",
    "owner_subject_ref and workspace membership contract",
    "schema and migration design",
    "store selector design",
    "Radish OIDC auth context contract",
    "repository contract smoke",
    "adapter smoke",
    "no sample fallback tests",
    "version conflict tests",
    "scope denied tests",
    "store unavailable tests",
    "no executor confirmation writeback replay side effects",
}
EXPECTED_RESPONSE_METADATA = {
    "current_draft_version",
    "current_updated_at",
    "current_updated_by_actor_ref",
    "request_id",
    "audit_ref",
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
        "workflow-saved-draft-consumer-smoke-v1": (
            CONSUMER_FIXTURE_PATH,
            "saved_draft_consumer_smoke_guarded",
        ),
        "user-workspace-draft-creation-v1": (
            CREATION_FIXTURE_PATH,
            "user_workspace_draft_creation_implemented",
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
        fixture.get("kind") == "workflow_saved_draft_durable_store_preconditions_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-durable-store-preconditions-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_durable_store_preconditions_defined",
        "durable store precondition status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_precondition_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("precondition_boundary") or {}
    require(
        boundary.get("status") == "preconditions_defined_not_implemented",
        "precondition boundary status drifted",
    )
    require(boundary.get("current_store_source") == "platform memory dev store", "current store drifted")
    require(
        boundary.get("future_store_source") == "future SavedWorkflowDraftRepository adapter",
        "future store drifted",
    )
    for field in (
        "dev_http_route_remains_default_off",
    ):
        require(boundary.get(field) is True, f"{field} must remain true")
    for field in (
        "durable_store_allowed_in_this_slice",
        "database_schema_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "publish_or_run_allowed_in_this_slice",
        "executor_allowed_in_this_slice",
        "writeback_or_replay_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_draft_scope_contract(fixture: dict[str, Any]) -> None:
    scope = fixture.get("draft_scope_contract") or {}
    require(
        scope.get("status") == "scope_contract_defined_not_implemented",
        "scope contract status drifted",
    )
    require(
        scope.get("scope_key") == "tenant_ref + workspace_id + application_id + draft_id",
        "scope key drifted",
    )
    missing_fields = sorted(EXPECTED_SCOPE_FIELDS - set(scope.get("required_identity_fields") or []))
    require(not missing_fields, f"missing scope fields: {missing_fields}")
    require(
        set(scope.get("required_request_scope_fields") or []) == {"workspace_id", "application_id"},
        "request scope fields drifted",
    )
    require(scope.get("scope_denial_failure_code") == "draft_scope_denied", "scope failure code drifted")
    for field in (
        "scope_denial_returns_draft_body",
        "cross_workspace_read_allowed",
        "cross_application_read_allowed",
        "query_scope_override_allowed",
        "sample_fallback_allowed",
    ):
        require(scope.get(field) is False, f"{field} must remain false")


def assert_owner_workspace_boundary(fixture: dict[str, Any]) -> None:
    owner = fixture.get("owner_workspace_boundary") or {}
    require(
        owner.get("status") == "owner_workspace_contract_defined_not_implemented",
        "owner/workspace boundary status drifted",
    )
    require(owner.get("current_actor_source") == "dev auth subject binding", "current actor source drifted")
    require(
        owner.get("future_owner_source") == "Radish OIDC subject or team ownership policy",
        "future owner source drifted",
    )
    missing_fields = sorted(
        EXPECTED_OWNER_FIELDS - set(owner.get("required_owner_fields_before_repository_adapter") or [])
    )
    require(not missing_fields, f"missing owner fields: {missing_fields}")
    for field in (
        "workspace_membership_check_required_before_repository_adapter",
        "tenant_predicate_required_before_repository_adapter",
    ):
        require(owner.get(field) is True, f"{field} must remain true")
    for field in (
        "owner_grants_cross_workspace_access",
        "team_sharing_defined_in_this_slice",
        "radish_oidc_claim_mapping_satisfied",
    ):
        require(owner.get(field) is False, f"{field} must remain false")


def assert_version_conflict_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("version_conflict_policy") or {}
    require(policy.get("status") == "optimistic_concurrency_required", "version policy status drifted")
    require(policy.get("client_field") == "expected_draft_version", "client version field drifted")
    require(policy.get("server_field") == "draft_version", "server version field drifted")
    require(policy.get("failure_code") == "draft_version_conflict", "version conflict code drifted")
    missing_metadata = sorted(EXPECTED_RESPONSE_METADATA - set(policy.get("required_response_metadata") or []))
    require(not missing_metadata, f"missing conflict response metadata: {missing_metadata}")
    for field in ("overwrite_on_conflict_allowed", "sample_fallback_allowed"):
        require(policy.get(field) is False, f"{field} must remain false")
    require(policy.get("local_draft_retained_on_conflict") is True, "local draft must be retained on conflict")


def assert_no_sample_fallback_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("no_sample_fallback_policy") or {}
    require(policy.get("status") == "fail_closed_required", "no sample fallback status drifted")
    require(
        set(policy.get("sample_or_fixture_may_be_displayed_only_as") or []) == {"sample", "unsaved_local"},
        "sample display labels drifted",
    )
    missing_codes = sorted(EXPECTED_NO_FALLBACK_FAILURE_CODES - set(policy.get("failure_codes") or []))
    require(not missing_codes, f"missing no sample fallback failure codes: {missing_codes}")
    for field in (
        "fallback_to_sample_as_saved_record_allowed",
        "fallback_to_dev_store_from_repository_mode_allowed",
    ):
        require(policy.get(field) is False, f"{field} must remain false")


def assert_store_switch_stopline(fixture: dict[str, Any]) -> None:
    stopline = fixture.get("store_switch_stopline") or {}
    require(
        stopline.get("status") == "repository_switch_reserved_not_implemented",
        "store switch stopline status drifted",
    )
    require(
        stopline.get("future_config_key") == "RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
        "future config key drifted",
    )
    require(stopline.get("formal_config_entry_created") is False, "future config key must not be created")
    current_mode = stopline.get("current_mode") or {}
    require(current_mode.get("mode") == "memory_dev_store", "current store mode drifted")
    require(current_mode.get("enabled_now") is True, "memory dev store must remain enabled now")
    require(current_mode.get("durable") is False, "memory dev store must not be durable")
    reserved_modes = {str(mode.get("mode")): mode for mode in stopline.get("reserved_modes") or []}
    require(set(reserved_modes) == {"repository", "database"}, "reserved store modes drifted")
    for mode in reserved_modes.values():
        require(mode.get("enabled_now") is False, "reserved mode must remain disabled")
        require(mode.get("fallback_to_dev_store_allowed") is False, "reserved mode must not fallback")
        require(mode.get("failure_code_until_enabled") == "repository_store_disabled", "reserved failure code drifted")
    missing_prereqs = sorted(EXPECTED_SWITCH_PREREQUISITES - set(stopline.get("required_before_switch") or []))
    require(not missing_prereqs, f"missing store switch prerequisites: {missing_prereqs}")
    require(stopline.get("switch_allowed_now") is False, "store switch must remain blocked")


def assert_implementation_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("implementation_artifact_guard") or {}
    for relative_path in guard.get("future_files_must_not_exist") or []:
        path = str(relative_path)
        if selector_implementation_file_allowed(REPO_ROOT, path):
            continue
        if schema_materialization_file_allowed(REPO_ROOT, path):
            continue
        require(not (REPO_ROOT / path).exists(), f"future artifact already exists: {relative_path}")
    absent_literals = [str(literal) for literal in guard.get("absent_literals") or []]
    for relative_path in guard.get("source_files_without_future_literals") or []:
        source = read(str(relative_path))
        leaked = [
            literal
            for literal in absent_literals
            if literal in source
            and not selector_implementation_literal_allowed(REPO_ROOT, literal)
            and not schema_materialization_literal_allowed(REPO_ROOT, literal)
        ]
        require(not leaked, f"{relative_path} leaked future literals: {leaked}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-user-workspace-draft-creation-v1.py"
    current_checker = "check-workflow-saved-draft-durable-store-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run durable store precondition check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "durable store precondition check must run after user workspace draft creation check",
    )
    doc_refs = fixture.get("required_doc_references") or {}
    require(isinstance(doc_refs, dict), "required_doc_references must be an object")
    for relative_path, required_literals in doc_refs.items():
        document = read(str(relative_path))
        missing = [str(literal) for literal in required_literals if str(literal) not in document]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    commands = {str(item.get("command")) for item in fixture.get("testing_strategy") or [] if isinstance(item, dict)}
    expected_commands = {
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-durable-store-preconditions-v1.py",
        "./scripts/check-repo.sh --fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing strategy commands: {missing_commands}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_precondition_boundary(fixture)
    assert_draft_scope_contract(fixture)
    assert_owner_workspace_boundary(fixture)
    assert_version_conflict_policy(fixture)
    assert_no_sample_fallback_policy(fixture)
    assert_store_switch_stopline(fixture)
    assert_implementation_artifact_guard(fixture)
    assert_docs_and_fast_baseline(fixture)
    assert_testing_strategy(fixture)
    print("workflow saved draft durable store preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
