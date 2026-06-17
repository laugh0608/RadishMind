#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from workflow_saved_draft_selector_implementation_guard import selector_implementation_literal_allowed


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.json"
)
READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-readiness-v1.json"
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

EXPECTED_OPERATIONS = {"SaveWorkflowDraftRecord", "ReadWorkflowDraftRecord", "ListWorkflowDraftRecords"}
EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-repository-contract-smoke-runner-readiness-v1": (
        READINESS_FIXTURE_PATH,
        "draft_repository_contract_smoke_runner_readiness_defined",
    ),
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
EXPECTED_FORBIDDEN_SOURCE = {
    "type SavedWorkflowDraftRepository interface",
    "func NewSavedWorkflowDraftRepositoryAdapter",
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


def smoke_failure_codes() -> set[str]:
    mapping = load_json(REPOSITORY_SMOKE_FIXTURE_PATH).get("failure_mapping") or {}
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
        fixture.get("kind") == "workflow_saved_draft_repository_contract_smoke_runner_implementation_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-repository-contract-smoke-runner-implementation-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_repository_contract_smoke_runner_implemented",
        "smoke runner implementation status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_implementation_boundary(fixture: dict[str, Any]) -> tuple[Path, Path]:
    boundary = fixture.get("implementation_boundary") or {}
    require(
        boundary.get("status") == "static_contract_smoke_runner_implemented_no_repository_adapter",
        "implementation boundary status drifted",
    )
    runner_file = REPO_ROOT / str(boundary.get("runner_file") or "")
    test_file = REPO_ROOT / str(boundary.get("test_file") or "")
    require(runner_file.exists(), f"missing runner file: {runner_file.relative_to(REPO_ROOT)}")
    require(test_file.exists(), f"missing test file: {test_file.relative_to(REPO_ROOT)}")
    require(
        boundary.get("runner_name") == "SavedWorkflowDraftRepositoryContractSmokeRunner",
        "runner name drifted",
    )
    require(
        boundary.get("contract_smoke_source") == "workflow-saved-draft-repository-contract-smoke-v1.json",
        "contract smoke source drifted",
    )
    require(
        boundary.get("runner_readiness_source")
        == "workflow-saved-draft-repository-contract-smoke-runner-readiness-v1.json",
        "runner readiness source drifted",
    )
    for field in (
        "repository_interface_declared",
        "repository_adapter_allowed_in_this_slice",
        "durable_store_allowed_in_this_slice",
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
    return runner_file, test_file


def assert_runner_source(fixture: dict[str, Any], runner_file: Path, test_file: Path) -> None:
    runner_text = runner_file.read_text(encoding="utf-8")
    test_text = test_file.read_text(encoding="utf-8")
    require("package httpapi" in runner_text, "runner must stay in httpapi package")
    for literal in fixture.get("implemented_runner_literals") or []:
        require(str(literal) in runner_text, f"runner missing literal: {literal}")
    for literal in fixture.get("runner_test_literals") or []:
        require(str(literal) in test_text, f"runner test missing literal: {literal}")


def assert_operation_runner_matrix(fixture: dict[str, Any]) -> None:
    readiness_matrix = {
        str(row.get("operation") or ""): row
        for row in load_json(READINESS_FIXTURE_PATH).get("operation_runner_matrix") or []
        if isinstance(row, dict)
    }
    smoke_matrix = {
        str(row.get("operation") or ""): row
        for row in load_json(REPOSITORY_SMOKE_FIXTURE_PATH).get("operation_smoke_matrix") or []
        if isinstance(row, dict)
    }
    matrix = {
        str(row.get("operation") or ""): row
        for row in fixture.get("operation_runner_matrix") or []
        if isinstance(row, dict)
    }
    require(set(matrix) == EXPECTED_OPERATIONS, "operation runner matrix drifted")
    require(set(readiness_matrix) == EXPECTED_OPERATIONS, "readiness operation matrix drifted")
    require(set(smoke_matrix) == EXPECTED_OPERATIONS, "smoke operation matrix drifted")
    for operation_name, row in matrix.items():
        readiness_row = readiness_matrix[operation_name]
        smoke_row = smoke_matrix[operation_name]
        require(row == readiness_row, f"{operation_name} runner matrix must match readiness gate")
        require(row.get("request_type") == smoke_row.get("request_type"), f"{operation_name} request type drifted")
        require(row.get("result_type") == smoke_row.get("result_type"), f"{operation_name} result type drifted")
        require(row.get("required_scope") == smoke_row.get("required_scope"), f"{operation_name} scope drifted")
        require(
            row.get("expected_success_projection") == smoke_row.get("expected_success_projection"),
            f"{operation_name} success projection drifted",
        )
        require(
            set(smoke_row.get("required_failure_codes") or []).issubset(
                set(row.get("required_failure_codes") or [])
            ),
            f"{operation_name} smoke failures missing",
        )
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
    require(smoke_failure_codes().issubset(required), "implementation failures must include smoke failures")
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
    readiness_checker = "check-workflow-saved-draft-repository-contract-smoke-runner-readiness-v1.py"
    current_checker = "check-workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.py"
    next_checker = "check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(current_checker in check_repo, "check-repo.py must run smoke runner implementation check")
    require(next_checker in check_repo, "product surface trigger recheck missing from check-repo.py")
    require(
        check_repo.index(readiness_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "implementation check must run after runner readiness and before product surface trigger recheck",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_no_forbidden_source(fixture: dict[str, Any]) -> None:
    configured = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_FORBIDDEN_SOURCE.issubset(configured), "forbidden source literals drifted")
    for root in (REPO_ROOT / "services/platform/internal", REPO_ROOT / "apps/radishmind-web/src"):
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured:
                if selector_implementation_literal_allowed(REPO_ROOT, literal):
                    continue
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this implementation slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    runner_file, test_file = assert_implementation_boundary(fixture)
    assert_runner_source(fixture, runner_file, test_file)
    assert_operation_runner_matrix(fixture)
    assert_failure_and_side_effect_policies(fixture)
    assert_docs_and_check_repo(fixture)
    assert_no_forbidden_source(fixture)
    print("workflow saved draft repository contract smoke runner implementation v1 checks passed.")


if __name__ == "__main__":
    main()
