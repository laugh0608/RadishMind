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
from workflow_saved_draft_repository_adapter_implementation_guard import (
    repository_adapter_implementation_file_allowed,
    repository_adapter_implementation_literal_allowed,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-v1.json"
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
IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.json"
)
RUNNER_IMPLEMENTATION_PATHS = {
    "services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner.go",
}

EXPECTED_DEPENDENCIES = {
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
EXPECTED_ACTOR_CONTEXT_FIELDS = {
    "request_id",
    "tenant_ref",
    "workspace_id",
    "application_id",
    "actor_subject_ref",
    "owner_subject_ref",
    "scope_grants",
    "audit_ref",
}
EXPECTED_DRAFT_SCOPE_FIELDS = {"tenant_ref", "workspace_id", "application_id", "draft_id"}
EXPECTED_SCHEMA_FAILURE_CODES = {
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
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
EXPECTED_PLANNED_FUTURE_ARTIFACTS = {
    "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-runner-v1.py",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository.go",
    "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter.go",
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


def implementation_gate_covers_runner() -> bool:
    if not IMPLEMENTATION_FIXTURE_PATH.exists():
        return False
    fixture = load_json(IMPLEMENTATION_FIXTURE_PATH)
    slice_info = fixture.get("slice") or {}
    boundary = fixture.get("implementation_boundary") or {}
    return (
        slice_info.get("id") == "workflow-saved-draft-repository-contract-smoke-runner-implementation-v1"
        and slice_info.get("status") == "draft_repository_contract_smoke_runner_implemented"
        and boundary.get("runner_file")
        == "services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner.go"
        and boundary.get("test_file")
        == "services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner_test.go"
    )


def repository_operation_rows() -> dict[str, dict[str, Any]]:
    fixture = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH)
    return {
        str(operation.get("method_name") or ""): operation
        for operation in fixture.get("repository_operations") or []
        if isinstance(operation, dict)
    }


def selector_smoke_operation_rows() -> dict[str, dict[str, Any]]:
    fixture = load_json(SELECTOR_SMOKE_READINESS_FIXTURE_PATH)
    return {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_selector_smoke_matrix") or []
        if isinstance(row, dict)
    }


def repository_failure_codes() -> set[str]:
    fixture = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH)
    policy = fixture.get("failure_policy") or {}
    return {str(code) for code in policy.get("required_failure_codes") or []}


def selector_failure_codes() -> set[str]:
    fixture = load_json(SELECTOR_SMOKE_READINESS_FIXTURE_PATH)
    mapping = fixture.get("failure_mapping") or {}
    return {str(code) for code in mapping.get("required_failure_codes") or []}


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
        fixture.get("kind") == "workflow_saved_draft_repository_contract_smoke_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-repository-contract-smoke-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_repository_contract_smoke_defined",
        "repository contract smoke status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_smoke_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("smoke_boundary") or {}
    repository_boundary = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH).get("repository_contract_boundary") or {}
    require(
        boundary.get("status") == "repository_contract_smoke_defined_not_implemented",
        "smoke boundary status drifted",
    )
    require(
        boundary.get("future_harness_name") == "SavedWorkflowDraftRepositoryContractSmoke",
        "future harness name drifted",
    )
    for field in ("current_store_source", "future_store_source"):
        require(boundary.get(field) == repository_boundary.get(field), f"{field} must match repository preconditions")
    require(boundary.get("repository_contract_smoke_defined_in_this_slice") is True, "smoke definition must be true")
    for field in (
        "repository_contract_smoke_runner_created_in_this_slice",
        "repository_contract_smoke_fixture_created_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_created_in_this_slice",
        "store_selector_created_in_this_slice",
        "database_schema_created_in_this_slice",
        "database_migration_created_in_this_slice",
        "oidc_validation_created_in_this_slice",
        "production_api_consumer_created_in_this_slice",
        "saved_draft_list_created_in_this_slice",
        "publish_or_run_created_in_this_slice",
        "executor_created_in_this_slice",
        "writeback_or_replay_created_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_planned_future_artifacts(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(artifact.get("path") or ""): artifact
        for artifact in fixture.get("planned_future_artifacts") or []
        if isinstance(artifact, dict)
    }
    require(set(artifacts) == EXPECTED_PLANNED_FUTURE_ARTIFACTS, "planned future artifacts drifted")
    for path, artifact in artifacts.items():
        require(artifact.get("created_in_this_slice") is False, f"{path} must not be created")
        if repository_adapter_implementation_file_allowed(REPO_ROOT, path):
            continue
        if (REPO_ROOT / path).exists():
            require(
                path in RUNNER_IMPLEMENTATION_PATHS and implementation_gate_covers_runner(),
                f"{path} exists before allowed runner implementation gate",
            )


def assert_io_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("smoke_io_contract") or {}
    repository_fixture = load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH)
    actor_context = repository_fixture.get("actor_context_contract") or {}
    operations = repository_operation_rows()
    require(contract.get("status") == "repository_contract_smoke_io_defined", "IO contract status drifted")
    require(
        contract.get("future_harness_name") == "SavedWorkflowDraftRepositoryContractSmoke",
        "IO contract harness name drifted",
    )
    require(
        set(contract.get("actor_context_fields") or []) == EXPECTED_ACTOR_CONTEXT_FIELDS,
        "actor context fields drifted",
    )
    require(
        set(actor_context.get("required_context_fields") or []) == EXPECTED_ACTOR_CONTEXT_FIELDS,
        "upstream actor context fields drifted",
    )
    require(set(contract.get("draft_scope_fields") or []) == EXPECTED_DRAFT_SCOPE_FIELDS, "draft scope fields drifted")
    require(
        set(contract.get("save_request_fields") or []) == set(operations["SaveWorkflowDraftRecord"].get("required_request_fields") or []),
        "save request fields must match repository preconditions",
    )
    require(
        set(contract.get("read_request_fields") or []) == set(operations["ReadWorkflowDraftRecord"].get("required_request_fields") or []),
        "read request fields must match repository preconditions",
    )
    require(
        set(contract.get("list_request_fields") or []) == set(operations["ListWorkflowDraftRecords"].get("required_request_fields") or []),
        "list request fields must match repository preconditions",
    )
    result_fields = contract.get("result_fields_by_operation") or {}
    require(set(result_fields) == EXPECTED_OPERATIONS, "result field operation map drifted")
    for operation_name, operation in operations.items():
        require(
            set(result_fields.get(operation_name) or []) == set(operation.get("required_result_fields") or []),
            f"{operation_name} result fields must match repository preconditions",
        )
    require(
        contract.get("success_output_policy") == "sanitized_saved_draft_record_or_summary_only",
        "success output policy drifted",
    )
    require(
        contract.get("failure_output_policy") == "failure_code_with_no_sample_or_memory_dev_fallback",
        "failure output policy drifted",
    )


def assert_operation_smoke_matrix(fixture: dict[str, Any]) -> None:
    upstream_operations = repository_operation_rows()
    selector_operations = selector_smoke_operation_rows()
    require(set(upstream_operations) == EXPECTED_OPERATIONS, "repository operation names drifted")
    require(set(selector_operations) == EXPECTED_OPERATIONS, "selector smoke operation names drifted")
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_smoke_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_OPERATIONS, "operation smoke matrix drifted")
    schema_failures = schema_failure_codes()
    require(EXPECTED_SCHEMA_FAILURE_CODES.issubset(schema_failures), "schema failure source drifted")
    for operation_name, row in rows.items():
        upstream = upstream_operations[operation_name]
        selector_row = selector_operations[operation_name]
        require(row.get("required_scope") == upstream.get("required_scope"), f"{operation_name} scope drifted")
        require(row.get("request_type") == upstream.get("request_type"), f"{operation_name} request type drifted")
        require(row.get("result_type") == upstream.get("result_type"), f"{operation_name} result type drifted")
        require(
            set(row.get("required_request_fields") or []) == set(upstream.get("required_request_fields") or []),
            f"{operation_name} request fields drifted",
        )
        require(
            set(row.get("required_result_fields") or []) == set(upstream.get("required_result_fields") or []),
            f"{operation_name} result fields drifted",
        )
        required_failures = set(row.get("required_failure_codes") or [])
        require(set(upstream.get("required_failure_codes") or []).issubset(required_failures), f"{operation_name} upstream failures missing")
        require(EXPECTED_SCHEMA_FAILURE_CODES.issubset(required_failures), f"{operation_name} schema failures missing")
        require(
            selector_row.get("reserved_mode_failure_code") in required_failures,
            f"{operation_name} reserved selector failure missing",
        )
        require(
            selector_row.get("unknown_mode_failure_code") in required_failures,
            f"{operation_name} unknown selector failure missing",
        )
        require(row.get("repository_contract_smoke_required") is True, f"{operation_name} smoke must be required")
        require(
            row.get("repository_contract_smoke_runner_implemented") is False,
            f"{operation_name} smoke runner must not be implemented",
        )
        require(
            row.get("repository_adapter_dependency_satisfied") is False,
            f"{operation_name} adapter dependency must remain false",
        )
        for field in (
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "fallback_to_fixture_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation_name} {field} must remain false")


def assert_failure_mapping_and_side_effect_policy(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    required = set(mapping.get("required_failure_codes") or [])
    require(repository_failure_codes().issubset(required), "repository failure codes missing")
    require(selector_failure_codes().issubset(required), "selector failure codes missing")
    require(EXPECTED_SCHEMA_FAILURE_CODES.issubset(required), "schema artifact failure codes missing")
    missing = sorted(EXPECTED_FAILURE_CODES - required)
    require(not missing, f"missing failure codes: {missing}")
    missing_preserved = sorted(EXPECTED_PRESERVED_FAILURE_CODES - set(mapping.get("preserve_existing_failure_codes") or []))
    require(not missing_preserved, f"missing preserved failure codes: {missing_preserved}")
    for field in (
        "version_conflict_must_not_be_rewritten_as_selector_failure",
        "scope_denied_must_not_return_draft_payload",
        "not_found_must_not_fallback_to_sample",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")
    for field in (
        "fallback_to_memory_dev_store_allowed",
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "success_on_contract_mismatch_allowed",
    ):
        require(mapping.get(field) is False, f"{field} must remain false")
    policy = fixture.get("no_side_effect_policy") or {}
    missing_counters = sorted(
        EXPECTED_SIDE_EFFECT_COUNTERS - set(policy.get("side_effect_counters_must_remain") or [])
    )
    require(not missing_counters, f"missing side effect counters: {missing_counters}")
    missing_effects = sorted(EXPECTED_FORBIDDEN_SIDE_EFFECTS - set(policy.get("forbidden_side_effects") or []))
    require(not missing_effects, f"missing forbidden side effects: {missing_effects}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("implementation_artifact_guard") or {}
    require(guard.get("status") == "forbid_implementation_artifacts", "artifact guard status drifted")
    for relative_path in guard.get("future_files_must_not_exist") or []:
        path = str(relative_path)
        if selector_implementation_file_allowed(REPO_ROOT, path):
            continue
        if schema_materialization_file_allowed(REPO_ROOT, path):
            continue
        if repository_adapter_implementation_file_allowed(REPO_ROOT, path):
            continue
        if (REPO_ROOT / path).exists():
            require(
                path in RUNNER_IMPLEMENTATION_PATHS and implementation_gate_covers_runner(),
                f"future artifact exists early: {relative_path}",
            )
    source_paths = guard.get("source_files_to_scan") or []
    literals = guard.get("future_literals_must_not_appear_in_source") or []
    require(source_paths, "source files to scan must be declared")
    require(literals, "future literals must be declared")
    for source_path in source_paths:
        source = read(str(source_path))
        for literal in literals:
            if selector_implementation_literal_allowed(REPO_ROOT, str(literal)):
                continue
            if schema_materialization_literal_allowed(REPO_ROOT, str(literal)):
                continue
            if repository_adapter_implementation_literal_allowed(REPO_ROOT, str(literal)):
                continue
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
    require(
        "workflow_saved_draft_repository_contract_smoke_checker" in required,
        "missing repository contract smoke checker",
    )
    require(
        "workflow_saved_draft_store_selector_smoke_readiness_checker" in required,
        "missing adjacent selector smoke readiness checker",
    )
    require("./scripts/check-repo.sh --fast" in required, "missing fast baseline in testing strategy")
    for field in (
        "does_not_start_service",
        "does_not_create_smoke_runner",
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
    previous_checker = "checks/control_plane/check-workflow-saved-draft-store-selector-smoke-readiness-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(current_checker in check_repo, "check-repo.py must run repository contract smoke check")
    require(previous_checker in check_repo, "store selector smoke readiness checker missing from check-repo.py")
    require(next_checker in check_repo, "product surface trigger recheck missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "repository contract smoke check must run after selector readiness and before product surface trigger recheck",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_smoke_boundary(fixture)
    assert_planned_future_artifacts(fixture)
    assert_io_contract(fixture)
    assert_operation_smoke_matrix(fixture)
    assert_failure_mapping_and_side_effect_policy(fixture)
    assert_artifact_guard(fixture)
    assert_required_doc_references(fixture)
    assert_testing_strategy(fixture)
    assert_check_repo_registration()
    print("workflow saved draft repository contract smoke v1 checks passed.")


if __name__ == "__main__":
    main()
