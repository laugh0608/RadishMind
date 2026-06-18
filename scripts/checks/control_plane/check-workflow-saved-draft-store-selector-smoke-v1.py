#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from workflow_saved_draft_schema_materialization_guard import (
    schema_materialization_file_allowed,
    schema_materialization_literal_allowed,
)
from workflow_saved_draft_repository_adapter_implementation_guard import (
    repository_adapter_implementation_file_allowed,
    repository_adapter_implementation_literal_allowed,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_OPERATIONS = {
    "SaveWorkflowDraftRecord",
    "ReadWorkflowDraftRecord",
    "ListWorkflowDraftRecords",
}
EXPECTED_MODES = {
    "memory_dev",
    "repository_disabled",
    "repository",
    "unknown",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "database_schema_ready",
    "database_migration_ready",
    "schema_artifact_materialized",
    "radish_oidc_ready",
    "token_validation_ready",
    "production_api_consumer_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read(relative_path: str) -> str:
    path = REPO_ROOT / relative_path
    require(path.exists(), f"required file missing: {relative_path}")
    return path.read_text(encoding="utf-8")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "workflow_saved_draft_store_selector_smoke_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-saved-draft-store-selector-smoke-v1", "slice id drifted")
    require(slice_info.get("track") == "Workflow / Agent Runtime", "slice track drifted")
    require(
        slice_info.get("status") == "draft_store_selector_smoke_implemented",
        "selector smoke status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_formal_config(fixture: dict[str, Any]) -> None:
    formal_config = fixture.get("formal_config") or {}
    require(formal_config.get("config_file_key") == "workflow_saved_draft_store", "config file key drifted")
    require(
        formal_config.get("env_key") == "RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
        "env key drifted",
    )
    require(formal_config.get("summary_field") == "workflow_saved_draft_store_mode", "summary field drifted")
    require(formal_config.get("field_source_key") == "workflow_saved_draft_store", "field source drifted")
    require(formal_config.get("default_mode") == "memory_dev", "default mode drifted")

    config_source = read("services/platform/internal/config/config.go")
    config_test_source = read("services/platform/internal/config/config_test.go")
    for needle in (
        "WorkflowSavedDraftStoreMode",
        "RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
        "workflow_saved_draft_store",
        "workflow_saved_draft_store_mode",
        "defaultDraftStoreMode",
        "memory_dev",
    ):
        require(needle in config_source, f"config.go missing {needle}")
    for needle in (
        "RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
        "repository_disabled",
        "workflow_saved_draft_store",
        "memory_dev",
    ):
        require(needle in config_test_source, f"config_test.go missing {needle}")


def assert_selector_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("selector_contract") or {}
    selector_file = str(contract.get("selector_file") or "")
    selector_test_file = str(contract.get("selector_test_file") or "")
    selector_source = read(selector_file)
    selector_test_source = read(selector_test_file)
    server_source = read("services/platform/internal/httpapi/server.go")
    http_source = read("services/platform/internal/httpapi/workflow_saved_draft_http.go")
    domain_source = read("services/platform/internal/httpapi/workflow_saved_draft.go")

    require(contract.get("selector_name") == "SelectWorkflowSavedDraftStore", "selector name drifted")
    require(contract.get("selector_type") == "WorkflowSavedDraftStoreSelector", "selector type drifted")
    require(contract.get("server_entry") == "newSavedWorkflowDraftStoreFromConfig", "server entry drifted")
    require(contract.get("success_mode") == "memory_dev", "success mode drifted")
    require(set(contract.get("reserved_modes") or []) == {"repository_disabled", "repository"}, "reserved modes drifted")
    require(
        contract.get("unknown_mode_failure_code") == "invalid_draft_store_mode",
        "unknown failure code drifted",
    )
    require(
        contract.get("reserved_mode_failure_code") == "repository_store_disabled",
        "reserved failure code drifted",
    )

    for needle in (
        "type WorkflowSavedDraftStoreSelector struct",
        "func SelectWorkflowSavedDraftStore",
        "WorkflowSavedDraftStoreModeMemoryDev",
        "WorkflowSavedDraftStoreModeRepositoryDisabled",
        "WorkflowSavedDraftStoreModeRepository",
        "disabledSavedWorkflowDraftStore",
        "savedWorkflowDraftStoreFailureCode",
    ):
        require(needle in selector_source, f"selector source missing {needle}")
    for needle in (
        "SavedWorkflowDraftFailureRepositoryStoreDisabled",
        "SavedWorkflowDraftFailureInvalidStoreMode",
    ):
        require(needle in domain_source, f"domain source missing {needle}")
        require(needle in selector_source, f"selector source missing {needle}")
        require(needle in selector_test_source, f"selector tests missing {needle}")
    require("newSavedWorkflowDraftStoreFromConfig(cfg)" in server_source, "server must select store from config")
    require("SelectWorkflowSavedDraftStore" in http_source, "HTTP store factory must call selector")


def assert_mode_matrix(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("mode") or ""): row
        for row in fixture.get("mode_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_MODES, "mode matrix drifted")
    require(rows["memory_dev"].get("success_allowed") is True, "memory_dev must remain successful")
    for mode in ("repository_disabled", "repository"):
        require(rows[mode].get("success_allowed") is False, f"{mode} must not succeed")
        require(rows[mode].get("failure_code") == "repository_store_disabled", f"{mode} failure drifted")
        require(rows[mode].get("fallback_to_memory_dev_allowed") is False, f"{mode} fallback drifted")
    require(rows["unknown"].get("failure_code") == "invalid_draft_store_mode", "unknown failure drifted")
    require(rows["unknown"].get("fallback_to_memory_dev_allowed") is False, "unknown fallback drifted")


def assert_operation_matrix(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_selector_smoke_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_OPERATIONS, "operation matrix must cover save/read/list")
    for operation, row in rows.items():
        require(row.get("memory_dev_success_allowed") is True, f"{operation} memory_dev success drifted")
        require(
            row.get("reserved_mode_failure_code") == "repository_store_disabled",
            f"{operation} reserved failure drifted",
        )
        require(
            row.get("unknown_mode_failure_code") == "invalid_draft_store_mode",
            f"{operation} unknown failure drifted",
        )
        missing = sorted(
            EXPECTED_SIDE_EFFECT_COUNTERS - set(row.get("side_effect_counters_must_remain_for_disabled_modes") or [])
        )
        require(not missing, f"{operation} missing side effect counters: {missing}")


def assert_tests_cover_selector_smoke() -> None:
    selector_test_source = read("services/platform/internal/httpapi/workflow_saved_draft_store_selector_test.go")
    http_test_source = read("services/platform/internal/httpapi/workflow_saved_draft_http_test.go")
    for needle in (
        "repository_disabled",
        "repository",
        "future_backend",
        "SavedWorkflowDraftFailureRepositoryStoreDisabled",
        "SavedWorkflowDraftFailureInvalidStoreMode",
        "SideEffects",
    ):
        require(needle in selector_test_source, f"selector test missing {needle}")
    for needle in (
        "store selector reserved and unknown modes fail closed",
        "newSavedWorkflowDraftHTTPTestServerWithStoreMode",
        "repository_disabled",
        "future_backend",
        "SavedWorkflowDraftFailureRepositoryStoreDisabled",
        "SavedWorkflowDraftFailureInvalidStoreMode",
    ):
        require(needle in http_test_source, f"HTTP test missing {needle}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    for relative_path in guard.get("forbidden_paths_must_not_exist") or []:
        if schema_materialization_file_allowed(REPO_ROOT, str(relative_path)):
            continue
        if repository_adapter_implementation_file_allowed(REPO_ROOT, str(relative_path)):
            continue
        require(not (REPO_ROOT / str(relative_path)).exists(), f"forbidden artifact exists: {relative_path}")
    sql_files = list((REPO_ROOT / "services/platform").rglob("*.sql"))
    require(not sql_files, f"SQL files must not be introduced: {sql_files}")

    source_paths = guard.get("source_files_to_scan") or []
    literals = guard.get("forbidden_source_literals") or []
    require(source_paths, "source files to scan must be declared")
    require(literals, "forbidden source literals must be declared")
    for source_path in source_paths:
        source = read(str(source_path))
        for literal in literals:
            if schema_materialization_literal_allowed(REPO_ROOT, str(literal)):
                continue
            if repository_adapter_implementation_literal_allowed(REPO_ROOT, str(literal)):
                continue
            require(str(literal) not in source, f"{source_path} contains forbidden literal: {literal}")


def assert_docs_and_testing(fixture: dict[str, Any]) -> None:
    for reference in fixture.get("required_doc_references") or []:
        path = str(reference.get("path") or "")
        require(path, "doc reference path missing")
        content = read(path)
        for needle in reference.get("must_contain") or []:
            require(str(needle) in content, f"{path} missing {needle}")

    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "static_checker_and_go_tests", "testing strategy status drifted")
    required = set(strategy.get("required_checks") or [])
    for expected_check in (
        "go test ./internal/config ./internal/httpapi",
        "workflow_saved_draft_store_selector_smoke_checker",
        "workflow_saved_draft_store_selector_smoke_readiness_checker",
        "workflow_saved_draft_store_selector_implementation_entry_review_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_create_repository_adapter",
        "does_not_create_schema_artifact_file",
        "does_not_create_sql_or_migration",
        "does_not_connect_database",
        "does_not_call_oidc",
        "does_not_create_production_api",
        "does_not_publish_or_run_workflow",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-review-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-store-selector-smoke-v1.py"
    require(current_checker in check_repo, "check-repo.py must run selector smoke checker")
    require(previous_checker in check_repo, "materialization review checker must remain registered")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "selector smoke checker must run after materialization review",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_formal_config(fixture)
    assert_selector_contract(fixture)
    assert_mode_matrix(fixture)
    assert_operation_matrix(fixture)
    assert_tests_cover_selector_smoke()
    assert_artifact_guard(fixture)
    assert_docs_and_testing(fixture)
    assert_check_repo_registration()


if __name__ == "__main__":
    main()
