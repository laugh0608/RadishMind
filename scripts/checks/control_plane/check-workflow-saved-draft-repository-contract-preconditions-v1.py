#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-preconditions-v1.json"
)
DURABLE_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-durable-store-preconditions-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "database_schema_ready",
    "database_migration_ready",
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
EXPECTED_CONTEXT_FIELDS = {
    "request_id",
    "tenant_ref",
    "workspace_id",
    "application_id",
    "actor_subject_ref",
    "owner_subject_ref",
    "scope_grants",
    "audit_ref",
}
EXPECTED_CONTEXT_FORBIDDEN_DEPENDENCIES = {
    "http_request",
    "query_string_tenant_override",
    "query_string_workspace_override",
    "raw_authorization_header",
    "cookie_value",
    "global_current_tenant",
    "provider_runtime",
    "workflow_executor",
}
EXPECTED_OPERATION_IDS = {
    "save-draft-record": "SaveWorkflowDraftRecord",
    "read-draft-record": "ReadWorkflowDraftRecord",
    "list-draft-records": "ListWorkflowDraftRecords",
}
EXPECTED_COMMON_RESULT_FIELDS = {
    "draft_id",
    "draft_version",
    "schema_version",
    "draft_status",
    "owner_subject_ref",
    "updated_by_actor_ref",
    "validation_summary",
    "blocked_capability_summary",
    "failure_code",
    "request_id",
    "audit_ref",
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
}
EXPECTED_FORBIDDEN_OUTPUT_FIELDS = {
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
EXPECTED_GATE_IDS = {
    "schema_migration_gate",
    "store_selector_gate",
    "auth_context_gate",
    "repository_contract_smoke_gate",
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
    expected_id = "workflow-saved-draft-durable-store-preconditions-v1"
    require(expected_id in declared, f"missing dependency: {expected_id}")
    dependency = load_json(DURABLE_PRECONDITIONS_FIXTURE_PATH)
    slice_info = dependency.get("slice") or {}
    require(slice_info.get("id") == expected_id, "durable preconditions dependency id drifted")
    require(
        slice_info.get("status") == "draft_durable_store_preconditions_defined",
        "durable preconditions dependency status drifted",
    )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "workflow_saved_draft_repository_contract_preconditions_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-repository-contract-preconditions-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_repository_contract_preconditions_defined",
        "repository contract precondition status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_repository_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("repository_contract_boundary") or {}
    require(
        boundary.get("status") == "contract_preconditions_defined_not_implemented",
        "repository boundary status drifted",
    )
    require(boundary.get("future_interface_name") == "SavedWorkflowDraftRepository", "interface name drifted")
    require(boundary.get("current_store_source") == "platform memory dev store", "current source drifted")
    require(
        boundary.get("future_store_source") == "future SavedWorkflowDraftRepository adapter",
        "future source drifted",
    )
    require(
        boundary.get("contract_preconditions_defined_in_this_slice") is True,
        "contract preconditions must be defined in this slice",
    )
    for field in (
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


def assert_actor_context(fixture: dict[str, Any]) -> None:
    contract = fixture.get("actor_context_contract") or {}
    require(
        contract.get("status") == "actor_context_contract_defined_not_implemented",
        "actor context status drifted",
    )
    require(
        contract.get("context_type") == "SavedWorkflowDraftRepositoryActorContext",
        "actor context type drifted",
    )
    missing_fields = sorted(EXPECTED_CONTEXT_FIELDS - set(contract.get("required_context_fields") or []))
    require(not missing_fields, f"missing actor context fields: {missing_fields}")
    missing_forbidden = sorted(
        EXPECTED_CONTEXT_FORBIDDEN_DEPENDENCIES - set(contract.get("must_not_depend_on") or [])
    )
    require(not missing_forbidden, f"missing actor context forbidden dependencies: {missing_forbidden}")
    for field in (
        "tenant_predicate_required",
        "workspace_membership_required",
        "owner_subject_ref_required_before_adapter",
    ):
        require(contract.get(field) is True, f"{field} must remain true")


def operation_by_id(fixture: dict[str, Any]) -> dict[str, dict[str, Any]]:
    operations = {
        str(operation.get("operation_id")): operation
        for operation in fixture.get("repository_operations") or []
        if isinstance(operation, dict)
    }
    require(set(operations) == set(EXPECTED_OPERATION_IDS), "repository operation ids drifted")
    return operations


def assert_save_operation(operation: dict[str, Any]) -> None:
    request_fields = set(operation.get("required_request_fields") or [])
    required = {
        "repository_actor_context",
        "draft_scope",
        "expected_draft_version",
        "sanitized_draft_payload",
        "validation_summary",
        "blocked_capability_summary",
        "request_audit_metadata",
    }
    missing = sorted(required - request_fields)
    require(not missing, f"save operation missing request fields: {missing}")
    require(operation.get("required_scope") == "workflow_drafts:write", "save operation scope drifted")
    require(operation.get("uses_expected_draft_version") is True, "save must use expected draft version")
    require("draft_version_conflict" in set(operation.get("required_failure_codes") or []), "save must cover conflicts")


def assert_read_operation(operation: dict[str, Any]) -> None:
    request_fields = set(operation.get("required_request_fields") or [])
    required = {"repository_actor_context", "draft_scope", "draft_id", "projection"}
    missing = sorted(required - request_fields)
    require(not missing, f"read operation missing request fields: {missing}")
    require(operation.get("required_scope") == "workflow_drafts:read", "read operation scope drifted")
    require(operation.get("uses_expected_draft_version") is False, "read must not require expected draft version")
    require("draft_not_found" in set(operation.get("required_failure_codes") or []), "read must cover not found")


def assert_list_operation(operation: dict[str, Any]) -> None:
    request_fields = set(operation.get("required_request_fields") or [])
    required = {
        "repository_actor_context",
        "workspace_id",
        "application_id",
        "owner_subject_ref",
        "cursor",
        "limit",
        "sort",
    }
    missing = sorted(required - request_fields)
    require(not missing, f"list operation missing request fields: {missing}")
    require(operation.get("required_scope") == "workflow_drafts:read", "list operation scope drifted")
    require(operation.get("uses_expected_draft_version") is False, "list must not require expected draft version")
    require("repository_store_disabled" in set(operation.get("required_failure_codes") or []), "list must cover disabled store")


def assert_repository_operations(fixture: dict[str, Any]) -> None:
    operations = operation_by_id(fixture)
    for operation_id, method_name in EXPECTED_OPERATION_IDS.items():
        operation = operations[operation_id]
        require(operation.get("method_name") == method_name, f"{operation_id} method name drifted")
        require(operation.get("request_type"), f"{operation_id} missing request type")
        require(operation.get("result_type"), f"{operation_id} missing result type")
        result_fields = set(operation.get("required_result_fields") or [])
        missing_result = sorted(EXPECTED_COMMON_RESULT_FIELDS - result_fields)
        require(not missing_result, f"{operation_id} missing result fields: {missing_result}")
        missing_codes = sorted(set(operation.get("required_failure_codes") or []) - EXPECTED_FAILURE_CODES)
        require(not missing_codes, f"{operation_id} has unknown failure codes: {missing_codes}")
        for field in ("fallback_to_sample_allowed", "fallback_to_dev_store_allowed", "side_effect_allowed"):
            require(operation.get(field) is False, f"{operation_id} {field} must remain false")
    assert_save_operation(operations["save-draft-record"])
    assert_read_operation(operations["read-draft-record"])
    assert_list_operation(operations["list-draft-records"])


def assert_failure_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("failure_policy") or {}
    require(policy.get("status") == "fail_closed_required", "failure policy status drifted")
    missing_codes = sorted(EXPECTED_FAILURE_CODES - set(policy.get("required_failure_codes") or []))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    for field in (
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_memory_dev_store_allowed",
        "conflict_overwrite_allowed",
    ):
        require(policy.get(field) is False, f"{field} must remain false")


def assert_projection_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("sanitized_projection_policy") or {}
    require(policy.get("status") == "sanitized_projection_required", "projection policy status drifted")
    allowed = set(policy.get("allowed_output_kinds") or [])
    for required_kind in (
        "saved_draft_record",
        "saved_draft_summary",
        "validation_summary",
        "blocked_capability_summary",
        "request_audit_metadata",
        "current_version_metadata",
    ):
        require(required_kind in allowed, f"projection policy missing allowed kind: {required_kind}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_OUTPUT_FIELDS - set(policy.get("forbidden_output_fields") or []))
    require(not missing_forbidden, f"missing forbidden projection fields: {missing_forbidden}")


def assert_implementation_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id")): gate
        for gate in fixture.get("implementation_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "implementation gate ids drifted")
    for gate_id, gate in gates.items():
        require(gate.get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
        require(gate.get("must_cover"), f"{gate_id} must define coverage")


def assert_implementation_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("implementation_artifact_guard") or {}
    for relative_path in guard.get("future_files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"future artifact already exists: {relative_path}")
    absent_literals = [str(literal) for literal in guard.get("absent_literals") or []]
    for relative_path in guard.get("source_files_without_future_literals") or []:
        source = read(str(relative_path))
        leaked = [literal for literal in absent_literals if literal in source]
        require(not leaked, f"{relative_path} leaked future literals: {leaked}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-saved-draft-durable-store-preconditions-v1.py"
    current_checker = "check-workflow-saved-draft-repository-contract-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run repository contract precondition check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "repository contract precondition check must run after durable store precondition check",
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
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-preconditions-v1.py",
        "./scripts/check-repo.sh --fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing strategy commands: {missing_commands}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_repository_boundary(fixture)
    assert_actor_context(fixture)
    assert_repository_operations(fixture)
    assert_failure_policy(fixture)
    assert_projection_policy(fixture)
    assert_implementation_gates(fixture)
    assert_implementation_artifact_guard(fixture)
    assert_docs_and_fast_baseline(fixture)
    assert_testing_strategy(fixture)
    print("workflow saved draft repository contract preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
