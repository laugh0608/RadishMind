#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-readiness-v1.json"
)
REPOSITORY_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-v1.json"
)
REPOSITORY_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-preconditions-v1.json"
)
AUTH_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-auth-context-preconditions-v1.json"
)
SCHEMA_ARTIFACT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-evidence-v1.json"
)
SELECTOR_SMOKE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-readiness-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-repository-contract-smoke-v1": (
        REPOSITORY_SMOKE_FIXTURE_PATH,
        "draft_repository_contract_smoke_defined",
    ),
    "workflow-saved-draft-repository-contract-preconditions-v1": (
        REPOSITORY_PRECONDITIONS_FIXTURE_PATH,
        "draft_repository_contract_preconditions_defined",
    ),
    "workflow-saved-draft-auth-context-preconditions-v1": (
        AUTH_PRECONDITIONS_FIXTURE_PATH,
        "draft_auth_context_preconditions_defined",
    ),
    "workflow-saved-draft-schema-artifact-evidence-v1": (
        SCHEMA_ARTIFACT_FIXTURE_PATH,
        "draft_schema_artifact_evidence_defined",
    ),
    "workflow-saved-draft-store-selector-smoke-readiness-v1": (
        SELECTOR_SMOKE_READINESS_FIXTURE_PATH,
        "draft_store_selector_smoke_readiness_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "repository_contract_smoke_runner_implemented",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "database_schema_ready",
    "database_migration_ready",
    "store_selector_implemented",
    "store_selector_smoke_ready",
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
EXPECTED_OPERATIONS = {"SaveWorkflowDraftRecord", "ReadWorkflowDraftRecord", "ListWorkflowDraftRecords"}
EXPECTED_INPUT_FIELDS = {
    "repository_contract_smoke_fixture",
    "repository_actor_context",
    "draft_scope",
    "operation_requests",
    "failure_injections",
    "schema_artifact_gate",
    "selector_smoke_gate",
    "side_effect_probe",
}
EXPECTED_OUTPUT_FIELDS = {
    "operation_results",
    "failure_results",
    "contract_mismatch_report",
    "fallback_report",
    "side_effect_report",
    "summary",
}
EXPECTED_FAILURE_CODES = {
    "draft_scope_denied",
    "draft_not_found",
    "draft_schema_version_unsupported",
    "draft_payload_invalid",
    "draft_version_conflict",
    "draft_store_unavailable",
    "draft_store_contract_mismatch",
    "repository_store_disabled",
    "invalid_draft_store_mode",
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
EXPECTED_SOURCE_ABSENT_LITERALS = {
    "type SavedWorkflowDraftRepositoryContractSmokeRunner",
    "func RunSavedWorkflowDraftRepositoryContractSmoke",
    "func NewSavedWorkflowDraftRepositoryContractSmokeRunner",
    "type SavedWorkflowDraftRepository interface",
    "func NewSavedWorkflowDraftRepositoryAdapter",
    "workflow_saved_draft_repository_contract_smoke_runner.go",
    "workflow_saved_draft_repository_adapter",
    "WorkflowSavedDraftStoreSelector",
    "RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
    "database/sql",
    "CREATE TABLE",
    "ALTER TABLE",
    "CREATE INDEX",
    "INSERT INTO saved_workflow_draft",
    "UPDATE saved_workflow_draft",
    "DELETE FROM saved_workflow_draft",
    "oidc.Provider",
    "ValidateWorkflowDraftToken",
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


def repository_operation_rows() -> dict[str, dict[str, Any]]:
    fixture = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH)
    return {
        str(operation.get("method_name") or ""): operation
        for operation in fixture.get("repository_operations") or []
        if isinstance(operation, dict)
    }


def smoke_operation_rows() -> dict[str, dict[str, Any]]:
    fixture = load_json(REPOSITORY_SMOKE_FIXTURE_PATH)
    return {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_smoke_matrix") or []
        if isinstance(row, dict)
    }


def smoke_failure_codes() -> set[str]:
    fixture = load_json(REPOSITORY_SMOKE_FIXTURE_PATH)
    mapping = fixture.get("failure_mapping") or {}
    return {str(code) for code in mapping.get("required_failure_codes") or []}


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
        fixture.get("kind") == "workflow_saved_draft_repository_contract_smoke_runner_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-repository-contract-smoke-runner-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_repository_contract_smoke_runner_readiness_defined",
        "smoke runner readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_runner_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("runner_boundary") or {}
    require(
        boundary.get("status") == "smoke_runner_readiness_defined_not_implemented",
        "runner boundary status drifted",
    )
    require(
        boundary.get("future_runner_name") == "SavedWorkflowDraftRepositoryContractSmokeRunner",
        "future runner name drifted",
    )
    require(
        boundary.get("future_runner_file")
        == "services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner.go",
        "future runner file path drifted",
    )
    require(
        boundary.get("future_runner_test_file")
        == "services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner_test.go",
        "future runner test file path drifted",
    )
    source_pairs = {
        "contract_smoke_source": "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-v1.json",
        "operation_contract_source": "scripts/checks/fixtures/workflow-saved-draft-repository-contract-preconditions-v1.json",
        "auth_context_source": "scripts/checks/fixtures/workflow-saved-draft-auth-context-preconditions-v1.json",
        "schema_artifact_source": "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-evidence-v1.json",
        "selector_smoke_source": "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-readiness-v1.json",
    }
    for field, expected_path in source_pairs.items():
        require(boundary.get(field) == expected_path, f"{field} path drifted")
        require((REPO_ROOT / expected_path).exists(), f"{field} target missing: {expected_path}")
    for field in (
        "runner_file_allowed_in_this_slice",
        "runner_test_allowed_in_this_slice",
        "runner_implementation_fixture_allowed_in_this_slice",
        "repository_interface_allowed_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "saved_draft_list_allowed_in_this_slice",
        "publish_or_run_allowed_in_this_slice",
        "executor_allowed_in_this_slice",
        "writeback_or_replay_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_runner_io_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("runner_io_contract") or {}
    require(
        contract.get("status") == "runner_input_output_readiness_defined",
        "runner IO contract status drifted",
    )
    require(set(contract.get("input_fields") or []) == EXPECTED_INPUT_FIELDS, "runner input fields drifted")
    require(set(contract.get("output_fields") or []) == EXPECTED_OUTPUT_FIELDS, "runner output fields drifted")
    require(
        contract.get("context_source") == "SavedWorkflowDraftRepositoryActorContext",
        "context source drifted",
    )
    request_source = str(contract.get("request_source") or "")
    result_source = str(contract.get("result_source") or "")
    for request_type in (
        "SaveWorkflowDraftRecordRequest",
        "ReadWorkflowDraftRecordRequest",
        "ListWorkflowDraftRecordsRequest",
    ):
        require(request_type in request_source, f"{request_type} missing from request source")
    for result_type in (
        "SaveWorkflowDraftRecordResult",
        "ReadWorkflowDraftRecordResult",
        "ListWorkflowDraftRecordsResult",
    ):
        require(result_type in result_source, f"{result_type} missing from result source")
    require(
        contract.get("success_policy")
        == "compare sanitized saved draft record or summary against repository operation result type",
        "success policy drifted",
    )
    require(
        contract.get("failure_policy") == "compare expected failure code without exposing database internal detail",
        "failure policy drifted",
    )


def assert_operation_runner_matrix(fixture: dict[str, Any]) -> None:
    upstream_operations = repository_operation_rows()
    smoke_rows = smoke_operation_rows()
    runner_rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_runner_matrix") or []
        if isinstance(row, dict)
    }
    require(set(upstream_operations) == EXPECTED_OPERATIONS, "repository operation names drifted")
    require(set(smoke_rows) == EXPECTED_OPERATIONS, "repository smoke operation names drifted")
    require(set(runner_rows) == EXPECTED_OPERATIONS, "runner matrix operation names drifted")
    for operation_name, row in runner_rows.items():
        upstream = upstream_operations[operation_name]
        smoke = smoke_rows[operation_name]
        require(row.get("required_scope") == upstream.get("required_scope"), f"{operation_name} scope drifted")
        require(row.get("request_type") == upstream.get("request_type"), f"{operation_name} request type drifted")
        require(row.get("result_type") == upstream.get("result_type"), f"{operation_name} result type drifted")
        require(
            row.get("expected_success_projection") == smoke.get("expected_success_projection"),
            f"{operation_name} success projection must match smoke contract",
        )
        required_failures = set(row.get("required_failure_codes") or [])
        require(set(smoke.get("required_failure_codes") or []).issubset(required_failures), f"{operation_name} smoke failures missing")
        require(EXPECTED_FAILURE_CODES.intersection(required_failures), f"{operation_name} failure set empty")
        for field in (
            "uses_repository_operation_contract",
            "uses_contract_smoke_fixture",
            "uses_auth_context_contract",
            "uses_schema_artifact_gate",
            "uses_selector_smoke_gate",
        ):
            require(row.get(field) is True, f"{operation_name} {field} must be true")
        for field in (
            "repository_contract_smoke_runner_implemented",
            "repository_adapter_dependency_satisfied",
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "fallback_to_fixture_allowed",
            "fallback_to_dev_http_route_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation_name} {field} must remain false")


def assert_failure_and_side_effect_policies(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    required = set(mapping.get("required_failure_codes") or [])
    require(smoke_failure_codes().issubset(required), "runner failures must include repository smoke failures")
    missing = sorted(EXPECTED_FAILURE_CODES - required)
    require(not missing, f"missing failure codes: {missing}")
    missing_preserved = sorted(EXPECTED_PRESERVED_FAILURE_CODES - set(mapping.get("preserve_existing_failure_codes") or []))
    require(not missing_preserved, f"missing preserved failure codes: {missing_preserved}")
    for field in (
        "version_conflict_must_not_be_rewritten_as_runner_failure",
        "scope_denied_must_not_return_draft_payload",
        "not_found_must_not_fallback_to_sample",
        "contract_mismatch_must_report_without_database_detail",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")

    fallback = fixture.get("no_fallback_policy") or {}
    require(fallback.get("status") == "required_for_future_runner", "fallback policy status drifted")
    for field in (
        "fallback_to_memory_dev_store_allowed",
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_dev_http_route_allowed",
        "fallback_to_test_auth_allowed",
        "success_on_contract_mismatch_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
    require(side_effects.get("status") == "required_for_future_runner", "side effect policy status drifted")
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(side_effects.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effects.get("forbidden_side_effects") or [])),
        "missing forbidden side effects",
    )


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("implementation_artifact_guard") or {}
    require(guard.get("status") == "forbid_implementation_artifacts", "artifact guard status drifted")
    for relative_path in guard.get("future_files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"future artifact exists early: {relative_path}")
    source_paths = guard.get("source_files_to_scan") or []
    literals = set(guard.get("future_literals_must_not_appear_in_source") or [])
    require(source_paths, "source files to scan must be declared")
    require(EXPECTED_SOURCE_ABSENT_LITERALS.issubset(literals), "source absent literals drifted")
    for source_path in source_paths:
        source = read(str(source_path))
        for literal in literals:
            require(str(literal) not in source, f"{source_path} contains future literal: {literal}")


def assert_required_doc_references(fixture: dict[str, Any]) -> None:
    references = fixture.get("required_doc_references") or []
    require(references, "required doc references must be declared")
    for reference in references:
        path = str(reference.get("path") or "")
        require(path, "doc reference path missing")
        content = read(path)
        for needle in reference.get("must_contain") or []:
            require(str(needle) in content, f"{path} missing required text: {needle}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "static_checker_only", "testing strategy status drifted")
    required = set(strategy.get("required_checks") or [])
    for check_id in (
        "workflow_saved_draft_repository_contract_smoke_runner_readiness_checker",
        "workflow_saved_draft_repository_contract_smoke_checker",
        "workflow_saved_draft_store_selector_smoke_readiness_checker",
        "npm run build",
        "./scripts/check-repo.sh --fast",
    ):
        require(check_id in required, f"missing required check: {check_id}")
    for field in (
        "does_not_start_service",
        "does_not_create_smoke_runner",
        "does_not_create_runner_test",
        "does_not_create_repository_interface",
        "does_not_create_repository_adapter",
        "does_not_create_selector",
        "does_not_connect_database",
        "does_not_call_oidc",
        "does_not_create_production_api",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-runner-readiness-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(current_checker in check_repo, "check-repo.py must run smoke runner readiness check")
    require(previous_checker in check_repo, "repository contract smoke checker missing from check-repo.py")
    require(next_checker in check_repo, "product surface trigger recheck missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "smoke runner readiness check must run after repository contract smoke and before product surface trigger recheck",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_runner_boundary(fixture)
    assert_runner_io_contract(fixture)
    assert_operation_runner_matrix(fixture)
    assert_failure_and_side_effect_policies(fixture)
    assert_artifact_guard(fixture)
    assert_required_doc_references(fixture)
    assert_testing_strategy(fixture)
    assert_check_repo_registration()
    print("workflow saved draft repository contract smoke runner readiness v1 checks passed.")


if __name__ == "__main__":
    main()
