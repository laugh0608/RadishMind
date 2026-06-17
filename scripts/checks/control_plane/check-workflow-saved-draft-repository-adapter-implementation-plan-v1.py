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
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json"
)
DURABLE_STORE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-durable-store-preconditions-v1.json"
)
REPOSITORY_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-preconditions-v1.json"
)
SCHEMA_MIGRATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-migration-preconditions-v1.json"
)
AUTH_CONTEXT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-auth-context-preconditions-v1.json"
)
SELECTOR_ENABLEMENT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-enablement-preconditions-v1.json"
)
SCHEMA_ARTIFACT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-evidence-v1.json"
)
SELECTOR_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-readiness-v1.json"
)
REPOSITORY_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-v1.json"
)
RUNNER_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_OPERATIONS = {
    "SaveWorkflowDraftRecord",
    "ReadWorkflowDraftRecord",
    "ListWorkflowDraftRecords",
}
EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-durable-store-preconditions-v1": (
        DURABLE_STORE_FIXTURE_PATH,
        "draft_durable_store_preconditions_defined",
    ),
    "workflow-saved-draft-repository-contract-preconditions-v1": (
        REPOSITORY_PRECONDITIONS_FIXTURE_PATH,
        "draft_repository_contract_preconditions_defined",
    ),
    "workflow-saved-draft-schema-migration-preconditions-v1": (
        SCHEMA_MIGRATION_FIXTURE_PATH,
        "draft_schema_migration_preconditions_defined",
    ),
    "workflow-saved-draft-auth-context-preconditions-v1": (
        AUTH_CONTEXT_FIXTURE_PATH,
        "draft_auth_context_preconditions_defined",
    ),
    "workflow-saved-draft-store-selector-enablement-preconditions-v1": (
        SELECTOR_ENABLEMENT_FIXTURE_PATH,
        "draft_store_selector_enablement_preconditions_defined",
    ),
    "workflow-saved-draft-schema-artifact-evidence-v1": (
        SCHEMA_ARTIFACT_FIXTURE_PATH,
        "draft_schema_artifact_evidence_defined",
    ),
    "workflow-saved-draft-store-selector-smoke-readiness-v1": (
        SELECTOR_SMOKE_FIXTURE_PATH,
        "draft_store_selector_smoke_readiness_defined",
    ),
    "workflow-saved-draft-repository-contract-smoke-v1": (
        REPOSITORY_SMOKE_FIXTURE_PATH,
        "draft_repository_contract_smoke_defined",
    ),
    "workflow-saved-draft-repository-contract-smoke-runner-implementation-v1": (
        RUNNER_IMPLEMENTATION_FIXTURE_PATH,
        "draft_repository_contract_smoke_runner_implemented",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "database_schema_ready",
    "database_migration_ready",
    "migration_runner_implemented",
    "store_selector_implemented",
    "selector_smoke_ready",
    "formal_store_config_ready",
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
EXPECTED_DEPENDENCY_CONSUMPTION_IDS = {
    "durable_store_preconditions_consumed",
    "repository_contract_preconditions_consumed",
    "schema_migration_preconditions_consumed",
    "auth_context_preconditions_consumed",
    "schema_artifact_evidence_consumed",
    "selector_smoke_readiness_consumed",
    "repository_contract_smoke_consumed",
    "static_runner_consumed",
}
EXPECTED_GATE_IDS = {
    "repository_contract_static_runner_consumed",
    "schema_auth_selector_evidence_consumed",
    "selector_implementation_gate",
    "schema_migration_artifact_gate",
    "production_auth_gate",
    "durable_adapter_smoke_gate",
    "no_adapter_selector_sql_or_oidc_leaked",
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
    "type SavedWorkflowDraftRepository interface",
    "func NewSavedWorkflowDraftRepositoryAdapter",
    "workflow_saved_draft_repository_adapter",
    "workflow_saved_draft_store_selector.go",
    "WorkflowSavedDraftStoreSelector",
    "RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE",
    "database/sql",
    "CREATE TABLE",
    "ALTER TABLE",
    "CREATE INDEX",
    "DROP TABLE",
    "INSERT INTO saved_workflow_draft",
    "UPDATE saved_workflow_draft",
    "DELETE FROM saved_workflow_draft",
    "services/platform/migrations/workflow_saved_drafts",
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


def rows_by_operation(fixture: dict[str, Any], key: str) -> dict[str, dict[str, Any]]:
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get(key) or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_OPERATIONS, f"{key} must cover saved draft save/read/list")
    return rows


def failure_codes_from_fixture(path: Path) -> set[str]:
    document = load_json(path)
    mapping = document.get("failure_mapping") or {}
    if isinstance(mapping, dict):
        return {str(code) for code in mapping.get("required_failure_codes") or []}
    if isinstance(mapping, list):
        return {str(code) for code in mapping}
    return set()


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
        fixture.get("kind") == "workflow_saved_draft_repository_adapter_implementation_plan_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-repository-adapter-implementation-plan-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_repository_adapter_implementation_plan_defined",
        "repository adapter implementation plan status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_plan_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_plan_boundary") or {}
    require(
        boundary.get("status") == "repository_adapter_implementation_plan_defined_not_implemented",
        "plan boundary status drifted",
    )
    require(boundary.get("future_interface_name") == "SavedWorkflowDraftRepository", "interface name drifted")
    for field in (
        "future_interface_file",
        "future_adapter_file",
        "future_adapter_test_file",
        "future_adapter_contract_smoke_test_file",
        "future_selector_file",
        "future_migration_root",
    ):
        relative_path = str(boundary.get(field) or "")
        if field == "future_selector_file" and selector_implementation_file_allowed(REPO_ROOT, relative_path):
            continue
        if field == "future_migration_root" and schema_materialization_file_allowed(REPO_ROOT, relative_path):
            continue
        future_path = REPO_ROOT / relative_path
        require(not future_path.exists(), f"{field} must not be created in this plan slice")
    for field in ("current_domain_file", "current_http_file", "static_runner_file"):
        require((REPO_ROOT / str(boundary.get(field) or "")).exists(), f"{field} must exist")
    for field in (
        "interface_file_created_in_this_slice",
        "adapter_file_created_in_this_slice",
        "adapter_test_created_in_this_slice",
        "adapter_contract_smoke_test_created_in_this_slice",
        "selector_file_created_in_this_slice",
        "migration_root_created_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "publish_or_run_allowed_in_this_slice",
        "executor_allowed_in_this_slice",
        "writeback_or_replay_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_dependency_consumption(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("id") or ""): row
        for row in fixture.get("dependency_consumption") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_DEPENDENCY_CONSUMPTION_IDS, "dependency consumption ids drifted")
    for row_id, row in rows.items():
        require(row.get("status") == "satisfied", f"{row_id} must be satisfied")
        evidence = row.get("evidence")
        require(evidence in EXPECTED_DEPENDENCIES, f"{row_id} cites unknown evidence")


def assert_future_file_layout(fixture: dict[str, Any]) -> None:
    rows = fixture.get("future_file_layout") or []
    require(len(rows) >= 6, "future file layout must include interface, adapter, tests, selector and migration root")
    for row in rows:
        require(isinstance(row, dict), "future file layout rows must be objects")
        relative_path = str(row.get("path") or "")
        require(relative_path, "future file layout row missing path")
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(row.get("allowed_before_gates") is False, f"{relative_path} must remain blocked before gates")
        if selector_implementation_file_allowed(REPO_ROOT, relative_path):
            continue
        if schema_materialization_file_allowed(REPO_ROOT, relative_path):
            continue
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this plan slice")


def assert_adapter_gate_matrix(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("adapter_implementation_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "adapter implementation gate ids drifted")
    for gate_id in (
        "repository_contract_static_runner_consumed",
        "schema_auth_selector_evidence_consumed",
        "selector_implementation_gate",
        "schema_migration_artifact_gate",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must be satisfied")
        require(gates[gate_id].get("evidence"), f"{gate_id} must cite evidence")
    for gate_id in (
        "production_auth_gate",
        "durable_adapter_smoke_gate",
    ):
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must define coverage")
    require(
        gates["no_adapter_selector_sql_or_oidc_leaked"].get("status") == "required_now",
        "leak gate must be required now",
    )


def assert_operation_adapter_matrix(fixture: dict[str, Any]) -> None:
    adapter_rows = rows_by_operation(fixture, "operation_adapter_matrix")
    runner_rows = rows_by_operation(load_json(RUNNER_IMPLEMENTATION_FIXTURE_PATH), "operation_runner_matrix")
    selector_smoke_rows = rows_by_operation(load_json(SELECTOR_SMOKE_FIXTURE_PATH), "operation_selector_smoke_matrix")
    selector_enablement_rows = rows_by_operation(
        load_json(SELECTOR_ENABLEMENT_FIXTURE_PATH),
        "operation_selector_matrix",
    )
    auth_rows = rows_by_operation(load_json(AUTH_CONTEXT_FIXTURE_PATH), "scope_grant_matrix")
    for operation, row in adapter_rows.items():
        runner_row = runner_rows[operation]
        auth_row = auth_rows[operation]
        selector_smoke_row = selector_smoke_rows[operation]
        selector_enablement_row = selector_enablement_rows[operation]
        require(row.get("required_scope") == runner_row.get("required_scope"), f"{operation} scope drifted")
        require(row.get("required_scope") == auth_row.get("required_scope"), f"{operation} auth scope drifted")
        require(row.get("request_type") == runner_row.get("request_type"), f"{operation} request type drifted")
        require(row.get("result_type") == runner_row.get("result_type"), f"{operation} result type drifted")
        require(
            row.get("success_projection") == runner_row.get("expected_success_projection"),
            f"{operation} success projection drifted",
        )
        require(
            selector_smoke_row.get("repository_adapter_dependency_required") is True,
            f"{operation} selector smoke must require repository adapter dependency",
        )
        require(
            selector_enablement_row.get("selector_enablement_allowed_now") is False,
            f"{operation} selector enablement must remain blocked",
        )
        for field in (
            "contract_smoke_consumed",
            "static_runner_consumed",
            "auth_context_consumed",
            "schema_artifact_evidence_consumed",
            "selector_smoke_readiness_consumed",
        ):
            require(row.get(field) is True, f"{operation} {field} must remain true")
        for field in (
            "adapter_implementation_allowed_now",
            "memory_dev_fallback_allowed",
            "sample_fallback_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")
        require(len(row.get("required_future_adapter_checks") or []) >= 5, f"{operation} adapter checks too thin")
        failures = set(row.get("failure_mapping") or [])
        require(set(runner_row.get("required_failure_codes") or []).issubset(failures), f"{operation} missing runner failures")
        require(
            set(selector_smoke_row.get("schema_artifact_failure_codes") or []).issubset(failures),
            f"{operation} missing schema artifact failures",
        )
        require("draft_store_contract_mismatch" in failures, f"{operation} must map contract mismatch")


def assert_failure_and_safety_policies(fixture: dict[str, Any]) -> None:
    failures = set(fixture.get("failure_mapping") or [])
    require(EXPECTED_FAILURE_CODES.issubset(failures), "failure mapping missing required codes")
    for dependency_path in (
        SCHEMA_MIGRATION_FIXTURE_PATH,
        SELECTOR_ENABLEMENT_FIXTURE_PATH,
        SELECTOR_SMOKE_FIXTURE_PATH,
        REPOSITORY_SMOKE_FIXTURE_PATH,
        RUNNER_IMPLEMENTATION_FIXTURE_PATH,
    ):
        dependency_failures = failure_codes_from_fixture(dependency_path)
        require(dependency_failures.issubset(failures), f"missing failures from {dependency_path.name}")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "repository_mode_fallback_to_memory_dev_store_allowed",
        "repository_mode_fallback_to_sample_allowed",
        "repository_mode_fallback_to_fixture_allowed",
        "schema_missing_fallback_to_memory_dev_store_allowed",
        "schema_missing_fallback_to_sample_allowed",
        "not_found_fallback_to_sample_allowed",
        "scope_denied_return_draft_body_allowed",
        "fallback_to_test_auth_allowed",
        "missing_tenant_success_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(side_effects.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effects.get("forbidden_side_effects") or [])),
        "missing forbidden side effects",
    )


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-workflow-saved-draft-repository-adapter-implementation-plan-v1.py"
    runner_checker = "check-workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.py"
    require(current_checker in check_repo, "check-repo.py must run saved draft repository adapter plan check")
    require(
        check_repo.index(runner_checker) < check_repo.index(current_checker),
        "saved draft repository adapter plan check must run after static runner implementation check",
    )

    references = fixture.get("required_doc_references") or {}
    require(isinstance(references, dict), "required_doc_references must be an object")
    for relative_path, required_literals in references.items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_implementation_leaked(fixture: dict[str, Any]) -> None:
    configured_literals = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_SOURCE_ABSENT_LITERALS.issubset(configured_literals), "source absent literals drifted")
    source_roots = [
        REPO_ROOT / "services/platform/internal",
        REPO_ROOT / "apps/radishmind-web/src",
    ]
    for root in source_roots:
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured_literals:
                if selector_implementation_literal_allowed(REPO_ROOT, literal):
                    continue
                if schema_materialization_literal_allowed(REPO_ROOT, literal):
                    continue
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this plan slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_plan_boundary(fixture)
    assert_dependency_consumption(fixture)
    assert_future_file_layout(fixture)
    assert_adapter_gate_matrix(fixture)
    assert_operation_adapter_matrix(fixture)
    assert_failure_and_safety_policies(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_implementation_leaked(fixture)
    print("workflow saved draft repository adapter implementation plan v1 checks passed.")


if __name__ == "__main__":
    main()
