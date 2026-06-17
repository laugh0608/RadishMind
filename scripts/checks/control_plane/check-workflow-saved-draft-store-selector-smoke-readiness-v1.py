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
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-readiness-v1.json"
)
REPOSITORY_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-preconditions-v1.json"
)
AUTH_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-auth-context-preconditions-v1.json"
)
SELECTOR_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-enablement-preconditions-v1.json"
)
SCHEMA_ARTIFACT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-evidence-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-repository-contract-preconditions-v1": (
        REPOSITORY_PRECONDITIONS_FIXTURE_PATH,
        "draft_repository_contract_preconditions_defined",
    ),
    "workflow-saved-draft-auth-context-preconditions-v1": (
        AUTH_PRECONDITIONS_FIXTURE_PATH,
        "draft_auth_context_preconditions_defined",
    ),
    "workflow-saved-draft-store-selector-enablement-preconditions-v1": (
        SELECTOR_PRECONDITIONS_FIXTURE_PATH,
        "draft_store_selector_enablement_preconditions_defined",
    ),
    "workflow-saved-draft-schema-artifact-evidence-v1": (
        SCHEMA_ARTIFACT_FIXTURE_PATH,
        "draft_schema_artifact_evidence_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "store_selector_smoke_ready",
    "store_selector_implemented",
    "formal_store_config_ready",
    "selector_config_entry_ready",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "schema_artifact_manifest_ready",
    "ddl_review_ready",
    "database_schema_ready",
    "database_migration_ready",
    "migration_files_created",
    "migration_runner_implemented",
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
EXPECTED_MODES = {"memory_dev", "repository_disabled", "repository", "unknown"}
RESERVED_MODES = {"repository_disabled", "repository"}
EXPECTED_OPERATIONS = {"SaveWorkflowDraftRecord", "ReadWorkflowDraftRecord", "ListWorkflowDraftRecords"}
EXPECTED_GATE_IDS = {
    "repository_contract_preconditions_consumed",
    "selector_enablement_preconditions_consumed",
    "schema_artifact_evidence_consumed",
    "selector_smoke_contract_defined",
    "selector_config_entry_gate",
    "selector_code_gate",
    "memory_dev_default_smoke_gate",
    "reserved_mode_fail_closed_smoke_gate",
    "unknown_mode_fail_closed_smoke_gate",
    "schema_artifact_failure_smoke_gate",
    "repository_adapter_enablement_gate",
    "no_selector_smoke_artifacts_leaked",
}
EXPECTED_PLANNED_ARTIFACTS = {
    "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-v1.py",
    "services/platform/internal/httpapi/workflow_saved_draft_store_selector.go",
    "services/platform/internal/httpapi/workflow_saved_draft_store_selector_test.go",
}
EXPECTED_SCHEMA_FAILURE_CODES = {
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
}
EXPECTED_FAILURE_CODES = {
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_store_unavailable",
    "draft_store_contract_mismatch",
    "draft_auth_context_contract_mismatch",
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
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


def repository_operations() -> set[str]:
    fixture = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH)
    return {
        str(operation.get("method_name") or "")
        for operation in fixture.get("repository_operations") or []
        if isinstance(operation, dict)
    }


def selector_mode_rows() -> dict[str, dict[str, Any]]:
    fixture = load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH)
    return {
        str(row.get("mode") or ""): row
        for row in fixture.get("store_mode_enablement_matrix") or []
        if isinstance(row, dict)
    }


def selector_operation_rows() -> dict[str, dict[str, Any]]:
    fixture = load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH)
    return {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_selector_matrix") or []
        if isinstance(row, dict)
    }


def schema_failure_codes() -> set[str]:
    fixture = load_json(SCHEMA_ARTIFACT_FIXTURE_PATH)
    mapping = fixture.get("failure_mapping") or {}
    return {
        str(code)
        for code in mapping.get("required_failure_codes") or []
        if str(code).startswith("draft_")
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
        fixture.get("kind") == "workflow_saved_draft_store_selector_smoke_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-store-selector-smoke-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_store_selector_smoke_readiness_defined",
        "store selector smoke readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_selector_smoke_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("selector_smoke_boundary") or {}
    selector_boundary = load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH).get("selector_enablement_boundary") or {}
    schema_boundary = load_json(SCHEMA_ARTIFACT_FIXTURE_PATH).get("schema_artifact_evidence_boundary") or {}
    require(
        boundary.get("status") == "store_selector_smoke_readiness_defined_not_implemented",
        "selector smoke boundary status drifted",
    )
    for field in (
        "current_default_mode",
        "current_store_source",
        "future_store_source",
        "future_config_key",
        "future_selector_name",
        "future_selector_type",
        "future_selector_file",
    ):
        require(boundary.get(field) == selector_boundary.get(field), f"{field} must match selector preconditions")
    require(
        boundary.get("future_store_source") == schema_boundary.get("future_store_source"),
        "future store source must match schema artifact evidence",
    )
    for path_field in (
        "future_selector_file",
        "future_selector_test_file",
        "future_selector_smoke_fixture",
        "future_selector_smoke_checker",
    ):
        relative_path = str(boundary.get(path_field) or "")
        require(relative_path, f"{path_field} missing")
        if selector_implementation_file_allowed(REPO_ROOT, relative_path):
            continue
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this readiness slice")
    for field in (
        "formal_config_entry_created",
        "selector_file_created_in_this_slice",
        "selector_test_created_in_this_slice",
        "selector_smoke_fixture_created_in_this_slice",
        "selector_smoke_checker_created_in_this_slice",
        "store_selector_implemented_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "saved_draft_list_allowed_in_this_slice",
        "publish_or_run_allowed_in_this_slice",
        "executor_allowed_in_this_slice",
        "writeback_or_replay_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_planned_artifacts(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(artifact.get("path") or ""): artifact
        for artifact in fixture.get("planned_selector_smoke_artifacts") or []
        if isinstance(artifact, dict)
    }
    require(set(artifacts) == EXPECTED_PLANNED_ARTIFACTS, "planned selector smoke artifacts drifted")
    for path, artifact in artifacts.items():
        require(artifact.get("created_in_this_slice") is False, f"{path} must not be created")
        if selector_implementation_file_allowed(REPO_ROOT, path):
            continue
        require(not (REPO_ROOT / path).exists(), f"{path} exists before selector smoke implementation")


def assert_gate_matrix(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("smoke_readiness_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "smoke readiness gate ids drifted")
    for gate_id in (
        "repository_contract_preconditions_consumed",
        "selector_enablement_preconditions_consumed",
        "schema_artifact_evidence_consumed",
        "selector_smoke_contract_defined",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must be satisfied")
    for gate_id in (
        "selector_config_entry_gate",
        "selector_code_gate",
        "memory_dev_default_smoke_gate",
        "reserved_mode_fail_closed_smoke_gate",
        "unknown_mode_fail_closed_smoke_gate",
        "schema_artifact_failure_smoke_gate",
        "repository_adapter_enablement_gate",
    ):
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
    require(gates["no_selector_smoke_artifacts_leaked"].get("status") == "required_now", "artifact leak gate drifted")


def assert_mode_matrix(fixture: dict[str, Any]) -> None:
    upstream = selector_mode_rows()
    require(set(upstream) == EXPECTED_MODES, "upstream selector mode ids drifted")
    rows = {
        str(row.get("mode") or ""): row
        for row in fixture.get("selector_smoke_mode_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_MODES, "selector smoke mode ids drifted")
    for mode, row in rows.items():
        upstream_row = upstream[mode]
        require(row.get("current_status") == upstream_row.get("current_status"), f"{mode} status drifted")
        require(row.get("expected_failure_code") == upstream_row.get("failure_code_when_selected"), f"{mode} failure drifted")
        require(row.get("smoke_required") is True, f"{mode} smoke must be required")
        for field in (
            "selector_smoke_ready",
            "selector_implementation_allowed_now",
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{mode} {field} must remain false")


def assert_operation_matrix(fixture: dict[str, Any]) -> None:
    require(repository_operations() == EXPECTED_OPERATIONS, "repository operation names drifted")
    upstream = selector_operation_rows()
    require(set(upstream) == EXPECTED_OPERATIONS, "selector operation matrix drifted")
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_selector_smoke_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_OPERATIONS, "operation selector smoke matrix drifted")
    for operation, row in rows.items():
        upstream_row = upstream[operation]
        require(row.get("default_mode") == upstream_row.get("default_mode"), f"{operation} default mode drifted")
        require(set(row.get("reserved_modes") or []) == RESERVED_MODES, f"{operation} reserved modes drifted")
        require(
            row.get("reserved_mode_failure_code") == upstream_row.get("reserved_mode_failure_code"),
            f"{operation} reserved failure drifted",
        )
        require(
            row.get("unknown_mode_failure_code") == upstream_row.get("unknown_mode_failure_code"),
            f"{operation} unknown failure drifted",
        )
        require(set(row.get("schema_artifact_failure_codes") or []) == EXPECTED_SCHEMA_FAILURE_CODES, f"{operation} schema failures")
        require(row.get("selector_smoke_required") is True, f"{operation} smoke must be required")
        require(row.get("repository_adapter_dependency_required") is True, f"{operation} adapter dependency required")
        for field in (
            "selector_smoke_ready",
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")


def assert_schema_artifact_failure_matrix(fixture: dict[str, Any]) -> None:
    require(EXPECTED_SCHEMA_FAILURE_CODES.issubset(schema_failure_codes()), "schema artifact failure source drifted")
    rows = {
        str(row.get("failure_code") or ""): row
        for row in fixture.get("schema_artifact_failure_smoke_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_SCHEMA_FAILURE_CODES, "schema artifact failure smoke matrix drifted")
    for failure_code, row in rows.items():
        require(row.get("source_evidence") == "workflow-saved-draft-schema-artifact-evidence-v1", f"{failure_code} evidence drifted")
        for field in (
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{failure_code} {field} must remain false")


def assert_failure_and_side_effect_policy(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    selector_failures = set(load_json(SELECTOR_PRECONDITIONS_FIXTURE_PATH).get("failure_mapping") or [])
    require(selector_failures.issubset(set(mapping.get("required_failure_codes") or [])), "selector failures missing")
    missing = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing, f"missing failure codes: {missing}")
    missing_preserved = sorted(EXPECTED_PRESERVED_FAILURE_CODES - set(mapping.get("preserve_existing_failure_codes") or []))
    require(not missing_preserved, f"missing preserved failure codes: {missing_preserved}")
    for field in (
        "missing_selector_success_allowed",
        "reserved_mode_success_allowed",
        "unknown_mode_success_allowed",
        "schema_artifact_failure_success_allowed",
        "fallback_to_memory_dev_store_allowed",
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
    ):
        require(mapping.get(field) is False, f"{field} must remain false")
    require(
        mapping.get("version_conflict_must_not_be_rewritten_as_selector_failure") is True,
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


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "static_checker_only", "testing strategy status drifted")
    required = set(strategy.get("required_checks") or [])
    require(
        "workflow_saved_draft_store_selector_smoke_readiness_checker" in required,
        "missing checker in testing strategy",
    )
    require("./scripts/check-repo.sh --fast" in required, "missing fast baseline in testing strategy")
    for field in (
        "does_not_start_service",
        "does_not_create_config_entry",
        "does_not_create_selector",
        "does_not_create_selector_smoke_fixture",
        "does_not_create_repository_adapter",
        "does_not_connect_database",
        "does_not_call_oidc",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-schema-artifact-evidence-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-store-selector-smoke-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run store selector smoke readiness check")
    require(previous_checker in check_repo, "schema artifact evidence checker missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "store selector smoke readiness check must run after schema artifact evidence",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_selector_smoke_boundary(fixture)
    assert_planned_artifacts(fixture)
    assert_gate_matrix(fixture)
    assert_mode_matrix(fixture)
    assert_operation_matrix(fixture)
    assert_schema_artifact_failure_matrix(fixture)
    assert_failure_and_side_effect_policy(fixture)
    assert_required_doc_references(fixture)
    assert_artifact_guard(fixture)
    assert_testing_strategy(fixture)
    assert_check_repo_registration()
    print("workflow saved draft store selector smoke readiness v1 checks passed.")


if __name__ == "__main__":
    main()
