#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-enablement-preconditions-v1.json"
)
CONSUMER_SMOKE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-consumer-smoke-v1.json"
DURABLE_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-durable-store-preconditions-v1.json"
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
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "store_selector_implemented",
    "formal_store_config_ready",
    "repository_mode_ready",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "database_schema_ready",
    "database_migration_ready",
    "migration_files_created",
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
EXPECTED_OPERATIONS = {"SaveWorkflowDraftRecord", "ReadWorkflowDraftRecord", "ListWorkflowDraftRecords"}
EXPECTED_GATE_IDS = {
    "selector_config_boundary_defined",
    "dev_http_write_boundary_preserved",
    "repository_disabled_gate",
    "repository_adapter_gate",
    "schema_migration_gate",
    "auth_context_gate",
    "selector_smoke_gate",
    "no_selector_code_leaked",
}
EXPECTED_FAILURE_CODES = {
    "repository_store_disabled",
    "invalid_draft_store_mode",
    "draft_store_unavailable",
    "draft_store_contract_mismatch",
    "draft_auth_context_contract_mismatch",
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
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


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "workflow-saved-draft-consumer-smoke-v1": (
            CONSUMER_SMOKE_FIXTURE_PATH,
            "saved_draft_consumer_smoke_guarded",
        ),
        "workflow-saved-draft-durable-store-preconditions-v1": (
            DURABLE_PRECONDITIONS_FIXTURE_PATH,
            "draft_durable_store_preconditions_defined",
        ),
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
        fixture.get("kind") == "workflow_saved_draft_store_selector_enablement_preconditions_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-store-selector-enablement-preconditions-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_store_selector_enablement_preconditions_defined",
        "store selector enablement status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_selector_enablement_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("selector_enablement_boundary") or {}
    auth_boundary = load_json(AUTH_PRECONDITIONS_FIXTURE_PATH).get("auth_context_boundary") or {}
    require(
        boundary.get("status") == "selector_enablement_preconditions_defined_not_implemented",
        "selector boundary status drifted",
    )
    require(boundary.get("current_default_mode") == "memory_dev", "default mode drifted")
    require(boundary.get("current_store_source") == "platform memory dev store", "current store source drifted")
    require(
        boundary.get("future_store_source") == auth_boundary.get("future_store_source"),
        "future store source must match auth preconditions",
    )
    require(
        boundary.get("future_config_key") == "RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
        "future config key drifted",
    )
    require(
        boundary.get("future_selector_file") == "services/platform/internal/httpapi/workflow_saved_draft_store_selector.go",
        "future selector file drifted",
    )
    require(boundary.get("selector_preconditions_defined_in_this_slice") is True, "selector preconditions required")
    for field in (
        "formal_config_entry_created",
        "selector_file_created_in_this_slice",
        "store_selector_implemented_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_created_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_created_in_this_slice",
        "saved_draft_list_created_in_this_slice",
        "publish_or_run_created_in_this_slice",
        "executor_created_in_this_slice",
        "writeback_or_replay_created_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_store_mode_matrix(fixture: dict[str, Any]) -> None:
    modes = {
        str(row.get("mode") or ""): row
        for row in fixture.get("store_mode_enablement_matrix") or []
        if isinstance(row, dict)
    }
    require(set(modes) == EXPECTED_MODES, "store mode ids drifted")
    require(modes["memory_dev"].get("current_status") == "current_default", "memory_dev status drifted")
    require(modes["memory_dev"].get("enabled_now") is True, "memory_dev must be enabled now")
    require(modes["memory_dev"].get("selector_implementation_required") is False, "memory_dev selector requirement drifted")
    require(modes["memory_dev"].get("failure_code_when_selected") is None, "memory_dev failure must remain null")
    for mode in ("repository_disabled", "repository"):
        row = modes[mode]
        require(row.get("enabled_now") is False, f"{mode} must not be enabled now")
        require(row.get("selector_implementation_required") is True, f"{mode} selector required")
        require(row.get("failure_code_when_selected") == "repository_store_disabled", f"{mode} failure drifted")
        require(row.get("fallback_to_memory_dev_store_allowed") is False, f"{mode} dev fallback must be false")
        require(row.get("fallback_to_sample_allowed") is False, f"{mode} sample fallback must be false")
    require(modes["unknown"].get("failure_code_when_selected") == "invalid_draft_store_mode", "unknown failure drifted")
    require(modes["unknown"].get("enabled_now") is False, "unknown mode must not be enabled")


def assert_operation_selector_matrix(fixture: dict[str, Any]) -> None:
    repository_operations = {
        str(operation.get("method_name") or "")
        for operation in load_json(REPOSITORY_PRECONDITIONS_FIXTURE_PATH).get("repository_operations") or []
        if isinstance(operation, dict)
    }
    require(repository_operations == EXPECTED_OPERATIONS, "repository operation names drifted")
    matrix = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_selector_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_OPERATIONS, "operation selector matrix drifted")
    for operation, row in matrix.items():
        require(row.get("default_mode") == "memory_dev", f"{operation} default mode drifted")
        require(set(row.get("reserved_modes") or []) == {"repository_disabled", "repository"}, f"{operation} reserved modes")
        require(row.get("reserved_mode_failure_code") == "repository_store_disabled", f"{operation} reserved failure")
        require(row.get("unknown_mode_failure_code") == "invalid_draft_store_mode", f"{operation} unknown failure")
        for field in (
            "selector_enablement_allowed_now",
            "fallback_to_memory_dev_store_allowed",
            "fallback_to_sample_allowed",
            "side_effect_allowed_on_denied",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")


def assert_enablement_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("enablement_gates") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "enablement gate ids drifted")
    for gate_id, gate in gates.items():
        require(gate.get("status") in {"required_now", "not_satisfied"}, f"{gate_id} status invalid")
        require(gate.get("must_cover"), f"{gate_id} must define coverage")
    for gate_id in ("repository_adapter_gate", "schema_migration_gate", "auth_context_gate", "selector_smoke_gate"):
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")


def assert_failure_and_dev_policy(fixture: dict[str, Any]) -> None:
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(fixture.get("failure_mapping") or []))
    require(not missing_failures, f"missing failure codes: {missing_failures}")
    dev = fixture.get("dev_flag_boundary") or {}
    consumer = load_json(CONSUMER_SMOKE_FIXTURE_PATH).get("route_contract") or {}
    env_flags = set(consumer.get("env_flags") or [])
    require(dev.get("saved_draft_dev_http_env") in env_flags, "dev http env must match consumer smoke")
    require(dev.get("saved_draft_dev_write_env") in env_flags, "dev write env must match consumer smoke")
    require(dev.get("store_mode_config_must_not_enable_dev_http") is True, "store mode must not enable dev HTTP")
    require(dev.get("store_mode_config_must_not_enable_writes") is True, "store mode must not enable writes")


def assert_no_fallback_and_side_effect_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("no_fallback_policy") or {}
    for field in (
        "reserved_mode_memory_dev_fallback_allowed",
        "reserved_mode_sample_fallback_allowed",
        "unknown_mode_success_allowed",
        "missing_selector_success_allowed",
        "repository_internal_error_exposure_allowed",
    ):
        require(policy.get(field) is False, f"{field} must remain false")
    side_effect = fixture.get("no_side_effect_policy") or {}
    missing_counters = sorted(
        EXPECTED_SIDE_EFFECT_COUNTERS - set(side_effect.get("side_effect_counters_must_remain") or [])
    )
    require(not missing_counters, f"missing side effect counters: {missing_counters}")
    missing_effects = sorted(EXPECTED_FORBIDDEN_SIDE_EFFECTS - set(side_effect.get("forbidden_side_effects") or []))
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
        "workflow_saved_draft_store_selector_enablement_preconditions_checker" in required,
        "missing checker in testing strategy",
    )
    require("./scripts/check-repo.sh --fast" in required, "missing fast baseline in testing strategy")
    for field in (
        "does_not_start_service",
        "does_not_create_config_entry",
        "does_not_create_selector",
        "does_not_create_repository_adapter",
        "does_not_connect_database",
        "does_not_call_oidc",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-auth-context-preconditions-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-store-selector-enablement-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run store selector preconditions check")
    require(previous_checker in check_repo, "auth context preconditions checker missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "store selector preconditions check must run after auth context preconditions",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_selector_enablement_boundary(fixture)
    assert_store_mode_matrix(fixture)
    assert_operation_selector_matrix(fixture)
    assert_enablement_gates(fixture)
    assert_failure_and_dev_policy(fixture)
    assert_no_fallback_and_side_effect_policy(fixture)
    assert_required_doc_references(fixture)
    assert_artifact_guard(fixture)
    assert_testing_strategy(fixture)
    assert_check_repo_registration()
    print("workflow saved draft store selector enablement preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
