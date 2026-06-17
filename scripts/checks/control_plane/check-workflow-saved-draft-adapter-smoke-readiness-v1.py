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
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-readiness-v1.json"
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json"
)
SCHEMA_MANIFEST_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-manifest-v1.json"
)
SELECTOR_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-readiness-v1.json"
)
AUTH_CONTEXT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-auth-context-preconditions-v1.json"
)
RUNNER_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "draft_repository_adapter_implementation_plan_defined",
    ),
    "workflow-saved-draft-schema-artifact-manifest-v1": (
        SCHEMA_MANIFEST_FIXTURE_PATH,
        "draft_schema_artifact_manifest_defined",
    ),
    "workflow-saved-draft-store-selector-smoke-readiness-v1": (
        SELECTOR_SMOKE_FIXTURE_PATH,
        "draft_store_selector_smoke_readiness_defined",
    ),
    "workflow-saved-draft-auth-context-preconditions-v1": (
        AUTH_CONTEXT_FIXTURE_PATH,
        "draft_auth_context_preconditions_defined",
    ),
    "workflow-saved-draft-repository-contract-smoke-runner-implementation-v1": (
        RUNNER_IMPLEMENTATION_FIXTURE_PATH,
        "draft_repository_contract_smoke_runner_implemented",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "adapter_smoke_fixture_created",
    "adapter_smoke_checker_created",
    "adapter_smoke_ready",
    "adapter_smoke_executed",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "repository_implementation_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "database_schema_ready",
    "database_connection_ready",
    "database_migration_ready",
    "schema_artifact_files_created",
    "schema_artifact_manifest_file_created",
    "ddl_review_artifact_created",
    "rollback_evidence_artifact_created",
    "migration_smoke_artifact_created",
    "store_selector_implemented",
    "selector_smoke_fixture_created",
    "formal_store_config_ready",
    "radish_oidc_ready",
    "auth_middleware_ready",
    "token_validation_ready",
    "production_auth_ready",
    "production_api_consumer_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_PLANNED_ARTIFACTS = {
    "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py",
    "services/platform/internal/httpapi/workflow_saved_draft_repository.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter_test.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter_contract_smoke_test.go",
    "services/platform/internal/httpapi/workflow_saved_draft_store_selector.go",
    "services/platform/migrations/workflow_saved_drafts",
}
EXPECTED_GATE_IDS = {
    "repository_adapter_implementation_plan_consumed",
    "schema_artifact_manifest_consumed",
    "store_selector_smoke_readiness_consumed",
    "auth_context_preconditions_consumed",
    "static_runner_consumed",
    "adapter_smoke_contract_defined",
    "selector_implementation_gate",
    "schema_artifact_materialization_gate",
    "production_auth_gate",
    "repository_adapter_implementation_gate",
    "adapter_smoke_execution_gate",
    "production_api_consumer_gate",
    "no_adapter_smoke_artifacts_leaked",
}
SATISFIED_GATE_IDS = {
    "repository_adapter_implementation_plan_consumed",
    "schema_artifact_manifest_consumed",
    "store_selector_smoke_readiness_consumed",
    "auth_context_preconditions_consumed",
    "static_runner_consumed",
    "adapter_smoke_contract_defined",
    "selector_implementation_gate",
}
NOT_SATISFIED_GATE_IDS = {
    "schema_artifact_materialization_gate",
    "production_auth_gate",
    "repository_adapter_implementation_gate",
    "adapter_smoke_execution_gate",
    "production_api_consumer_gate",
}
EXPECTED_OPERATIONS = {
    "SaveWorkflowDraftRecord",
    "ReadWorkflowDraftRecord",
    "ListWorkflowDraftRecords",
}
EXPECTED_REQUIRED_INPUTS = {
    "static runner case",
    "repository adapter implementation plan",
    "schema artifact manifest contract",
    "selector smoke readiness",
    "auth context preconditions",
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


def rows_by_operation(fixture: dict[str, Any], key: str) -> dict[str, dict[str, Any]]:
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get(key) or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_OPERATIONS, f"{key} must cover save/read/list")
    return rows


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
    require(fixture.get("kind") == "workflow_saved_draft_adapter_smoke_readiness_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-adapter-smoke-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_adapter_smoke_readiness_defined",
        "adapter smoke readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_adapter_smoke_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("adapter_smoke_boundary") or {}
    adapter_boundary = load_json(ADAPTER_PLAN_FIXTURE_PATH).get("implementation_plan_boundary") or {}
    schema_boundary = load_json(SCHEMA_MANIFEST_FIXTURE_PATH).get("schema_artifact_manifest_boundary") or {}
    selector_boundary = load_json(SELECTOR_SMOKE_FIXTURE_PATH).get("selector_smoke_boundary") or {}
    require(
        boundary.get("status") == "adapter_smoke_readiness_defined_not_implemented",
        "adapter smoke boundary status drifted",
    )
    for field in (
        "future_interface_file",
        "future_adapter_file",
        "future_adapter_test_file",
        "future_adapter_contract_smoke_test_file",
        "future_selector_file",
        "future_migration_root",
    ):
        require(boundary.get(field) == adapter_boundary.get(field), f"{field} must match adapter plan")
    require(
        boundary.get("future_migration_root") == schema_boundary.get("future_migration_root"),
        "future migration root must match schema manifest boundary",
    )
    require(
        boundary.get("future_selector_file") == selector_boundary.get("future_selector_file"),
        "future selector file must match selector smoke readiness",
    )
    require(
        boundary.get("current_static_runner_file") == adapter_boundary.get("static_runner_file"),
        "static runner file must match adapter plan",
    )
    require((REPO_ROOT / str(boundary.get("current_static_runner_file") or "")).exists(), "static runner missing")
    for path_field in (
        "future_adapter_smoke_fixture",
        "future_adapter_smoke_checker",
        "future_interface_file",
        "future_adapter_file",
        "future_adapter_test_file",
        "future_adapter_contract_smoke_test_file",
        "future_selector_file",
        "future_migration_root",
    ):
        relative_path = str(boundary.get(path_field) or "")
        require(relative_path, f"{path_field} missing")
        if selector_implementation_file_allowed(REPO_ROOT, relative_path):
            continue
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this readiness slice")
    for field in (
        "adapter_smoke_fixture_created_in_this_slice",
        "adapter_smoke_checker_created_in_this_slice",
        "interface_file_created_in_this_slice",
        "adapter_file_created_in_this_slice",
        "adapter_test_created_in_this_slice",
        "adapter_contract_smoke_test_created_in_this_slice",
        "selector_file_created_in_this_slice",
        "migration_root_created_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "repository_interface_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "schema_artifact_file_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "publish_or_run_allowed_in_this_slice",
        "executor_allowed_in_this_slice",
        "writeback_or_replay_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_planned_adapter_smoke_artifacts(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(item.get("path") or ""): item
        for item in fixture.get("planned_adapter_smoke_artifacts") or []
        if isinstance(item, dict)
    }
    require(set(artifacts) == EXPECTED_PLANNED_ARTIFACTS, "planned adapter smoke artifacts drifted")
    for relative_path, artifact in artifacts.items():
        require(artifact.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        if selector_implementation_file_allowed(REPO_ROOT, relative_path):
            continue
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this readiness slice")


def assert_adapter_smoke_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("adapter_smoke_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "adapter smoke gate ids drifted")
    for gate_id in SATISFIED_GATE_IDS:
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must remain satisfied")
    for gate_id in NOT_SATISFIED_GATE_IDS:
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage")
    require(
        gates["no_adapter_smoke_artifacts_leaked"].get("status") == "required_now",
        "adapter smoke artifact leak gate status drifted",
    )


def assert_dependency_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("adapter_smoke_dependency_contract") or {}
    require(
        contract.get("status") == "adapter_smoke_dependencies_defined_not_ready",
        "adapter smoke dependency contract status drifted",
    )
    rows = {
        str(row.get("source") or ""): row
        for row in contract.get("required_before_adapter_smoke") or []
        if isinstance(row, dict)
    }
    require(set(rows) == set(EXPECTED_DEPENDENCIES), "adapter smoke dependency contract sources drifted")
    for source, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        row = rows[source]
        require(row.get("source_status") == expected_status, f"{source} source status drifted")
        actual = load_json(path).get("slice", {}).get("status")
        require(actual == expected_status, f"{source} fixture status drifted")
        expected_ready = source == "workflow-saved-draft-repository-contract-smoke-runner-implementation-v1"
        require(row.get("ready_for_adapter_smoke_now") is expected_ready, f"{source} readiness drifted")


def assert_operation_adapter_smoke_matrix(fixture: dict[str, Any]) -> None:
    rows = rows_by_operation(fixture, "operation_adapter_smoke_matrix")
    adapter_rows = rows_by_operation(load_json(ADAPTER_PLAN_FIXTURE_PATH), "operation_adapter_matrix")
    runner_rows = rows_by_operation(load_json(RUNNER_IMPLEMENTATION_FIXTURE_PATH), "operation_runner_matrix")
    manifest_rows = rows_by_operation(load_json(SCHEMA_MANIFEST_FIXTURE_PATH), "operation_manifest_matrix")
    selector_rows = rows_by_operation(load_json(SELECTOR_SMOKE_FIXTURE_PATH), "operation_selector_smoke_matrix")
    auth_rows = rows_by_operation(load_json(AUTH_CONTEXT_FIXTURE_PATH), "scope_grant_matrix")
    for operation, row in rows.items():
        require(row.get("required_scope") == adapter_rows[operation].get("required_scope"), f"{operation} scope drifted")
        require(row.get("required_scope") == runner_rows[operation].get("required_scope"), f"{operation} runner scope drifted")
        require(row.get("required_scope") == manifest_rows[operation].get("required_scope"), f"{operation} manifest scope drifted")
        require(row.get("required_scope") == auth_rows[operation].get("required_scope"), f"{operation} auth scope drifted")
        require(row.get("request_type") == adapter_rows[operation].get("request_type"), f"{operation} request type drifted")
        require(row.get("result_type") == adapter_rows[operation].get("result_type"), f"{operation} result type drifted")
        require(
            row.get("expected_success_projection") == runner_rows[operation].get("expected_success_projection"),
            f"{operation} success projection must match runner",
        )
        require(row.get("expected_success_projection") == adapter_rows[operation].get("success_projection"), f"{operation} projection drifted")
        missing_inputs = sorted(EXPECTED_REQUIRED_INPUTS - set(row.get("required_inputs") or []))
        require(not missing_inputs, f"{operation} missing required inputs: {missing_inputs}")
        for field in (
            "schema_artifact_manifest_consumed",
            "selector_smoke_readiness_consumed",
            "auth_context_preconditions_consumed",
            "static_runner_case_consumed",
            "adapter_plan_consumed",
            "adapter_smoke_required",
        ):
            require(row.get(field) is True, f"{operation} {field} must remain true")
        require(selector_rows[operation].get("selector_smoke_required") is True, f"{operation} selector smoke must remain required")
        require(manifest_rows[operation].get("schema_artifact_manifest_required") is True, f"{operation} manifest must remain required")
        for field in (
            "adapter_smoke_fixture_created_now",
            "adapter_smoke_checker_created_now",
            "adapter_smoke_ready",
            "repository_adapter_allowed_now",
            "selector_implementation_allowed_now",
            "schema_artifact_files_allowed_now",
            "oidc_validation_allowed_now",
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "fallback_to_fixture_allowed",
            "fallback_to_dev_http_route_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")


def assert_failure_fallback_and_side_effects(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(mapping.get("status") == "adapter_smoke_fail_closed_required", "failure mapping status drifted")
    expected_codes = set(load_json(ADAPTER_PLAN_FIXTURE_PATH).get("failure_mapping") or [])
    missing = sorted(expected_codes - set(mapping.get("required_failure_codes") or []))
    require(not missing, f"missing failure codes from adapter plan: {missing}")
    missing_preserved = sorted(EXPECTED_PRESERVED_FAILURE_CODES - set(mapping.get("preserve_existing_failure_codes") or []))
    require(not missing_preserved, f"missing preserved failure codes: {missing_preserved}")
    for field in (
        "version_conflict_must_not_be_rewritten_as_adapter_smoke_failure",
        "version_conflict_must_not_be_rewritten_as_selector_failure",
        "version_conflict_must_not_be_rewritten_as_migration_failure",
        "scope_denied_must_not_return_draft_payload",
        "not_found_must_not_fallback_to_sample",
        "contract_mismatch_must_report_without_database_detail",
        "selector_failure_must_not_fallback_to_memory_dev",
        "auth_context_mismatch_must_not_fallback_to_test_auth",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")
    for field in (
        "missing_schema_artifact_success_allowed",
        "selector_failure_success_allowed",
        "auth_context_mismatch_success_allowed",
        "database_detail_leak_allowed",
        "oidc_detail_leak_allowed",
    ):
        require(mapping.get(field) is False, f"{field} must remain false")

    fallback = fixture.get("no_fallback_policy") or {}
    for field in (
        "fallback_to_memory_dev_store_allowed",
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_dev_http_route_allowed",
        "fallback_to_test_auth_allowed",
        "success_on_missing_schema_artifact_allowed",
        "success_on_selector_failure_allowed",
        "success_on_auth_context_mismatch_allowed",
        "success_on_contract_mismatch_allowed",
        "reserved_store_mode_success_allowed",
        "unknown_store_mode_success_allowed",
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
        "workflow_saved_draft_adapter_smoke_readiness_checker",
        "workflow_saved_draft_schema_artifact_manifest_checker",
        "workflow_saved_draft_repository_adapter_implementation_plan_checker",
        "workflow_saved_draft_store_selector_smoke_readiness_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_create_adapter_smoke_fixture",
        "does_not_create_adapter_smoke_checker",
        "does_not_create_repository_interface",
        "does_not_create_repository_adapter",
        "does_not_create_selector",
        "does_not_create_schema_artifact_file",
        "does_not_create_sql_or_migration",
        "does_not_connect_database",
        "does_not_call_oidc",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-schema-artifact-manifest-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "schema artifact manifest checker missing from check-repo.py")
    require(current_checker in check_repo, "check-repo.py must run adapter smoke readiness check")
    require(next_checker in check_repo, "product surface recheck missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "adapter smoke readiness check must run after schema artifact manifest and before product surface recheck",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_adapter_smoke_boundary(fixture)
    assert_planned_adapter_smoke_artifacts(fixture)
    assert_adapter_smoke_gates(fixture)
    assert_dependency_contract(fixture)
    assert_operation_adapter_smoke_matrix(fixture)
    assert_failure_fallback_and_side_effects(fixture)
    assert_artifact_guard(fixture)
    assert_required_docs_and_testing(fixture)
    assert_check_repo_registration()
    print("workflow saved draft adapter smoke readiness v1 checks passed.")


if __name__ == "__main__":
    main()
