#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from workflow_saved_draft_selector_implementation_guard import (
    selector_implementation_active,
    selector_implementation_file_allowed,
    selector_implementation_literal_allowed,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-store-selector-implementation-entry-review-v1.json"
)
SELECTOR_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-enablement-preconditions-v1.json"
)
SELECTOR_SMOKE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-readiness-v1.json"
)
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json"
)
SCHEMA_MANIFEST_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-manifest-v1.json"
)
ADAPTER_SMOKE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-readiness-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-store-selector-enablement-preconditions-v1": (
        SELECTOR_PRECONDITIONS_FIXTURE_PATH,
        "draft_store_selector_enablement_preconditions_defined",
    ),
    "workflow-saved-draft-store-selector-smoke-readiness-v1": (
        SELECTOR_SMOKE_READINESS_FIXTURE_PATH,
        "draft_store_selector_smoke_readiness_defined",
    ),
    "workflow-saved-draft-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "draft_repository_adapter_implementation_plan_defined",
    ),
    "workflow-saved-draft-schema-artifact-manifest-v1": (
        SCHEMA_MANIFEST_FIXTURE_PATH,
        "draft_schema_artifact_manifest_defined",
    ),
    "workflow-saved-draft-adapter-smoke-readiness-v1": (
        ADAPTER_SMOKE_READINESS_FIXTURE_PATH,
        "draft_adapter_smoke_readiness_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "selector_implementation_entry_opened",
    "selector_implementation_task_card_created",
    "formal_store_config_ready",
    "selector_config_entry_ready",
    "store_selector_implemented",
    "selector_unit_tests_created",
    "selector_smoke_fixture_created",
    "selector_smoke_checker_created",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "schema_artifact_files_created",
    "schema_artifact_manifest_file_created",
    "database_schema_ready",
    "database_migration_ready",
    "migration_files_created",
    "migration_runner_implemented",
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
EXPECTED_GATE_IDS = {
    "selector_enablement_preconditions_consumed",
    "selector_smoke_readiness_consumed",
    "adapter_smoke_readiness_consumed",
    "selector_entry_review_defined",
    "formal_config_candidate_blocked",
    "selector_function_candidate_blocked",
    "selector_tests_candidate_blocked",
    "selector_smoke_fixture_candidate_blocked",
    "repository_adapter_gate",
    "schema_artifact_file_gate",
    "production_auth_gate",
    "production_api_gate",
    "no_selector_implementation_artifacts_leaked",
}
EXPECTED_CANDIDATE_IDS = {
    "formal_store_config_entry",
    "select_workflow_saved_draft_store",
    "selector_unit_tests",
    "selector_smoke_fixture",
}
EXPECTED_FUTURE_FILES = {
    "docs/task-cards/workflow-saved-draft-store-selector-implementation-v1-plan.md",
    "services/platform/internal/httpapi/workflow_saved_draft_store_selector.go",
    "services/platform/internal/httpapi/workflow_saved_draft_store_selector_test.go",
    "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-v1.py",
    "services/platform/migrations/workflow_saved_drafts",
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
    "draft_not_found",
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
        fixture.get("kind") == "workflow_saved_draft_store_selector_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-store-selector-implementation-entry-review-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_store_selector_implementation_entry_review_defined",
        "store selector implementation entry review status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("selector_entry_boundary") or {}
    selector_boundary = load_json(SELECTOR_SMOKE_READINESS_FIXTURE_PATH).get("selector_smoke_boundary") or {}
    adapter_boundary = load_json(ADAPTER_SMOKE_READINESS_FIXTURE_PATH).get("adapter_smoke_boundary") or {}
    require(
        boundary.get("status") == "selector_implementation_entry_review_defined_no_runtime_change",
        "selector entry boundary status drifted",
    )
    require(boundary.get("decision") == "selector_implementation_entry_not_opened", "entry decision drifted")
    require(boundary.get("current_development_mode") == "entry_review_only", "development mode drifted")
    require(boundary.get("selected_implementation_track_now") == "none", "selected track must remain none")
    for field in (
        "future_config_key",
        "future_selector_name",
        "future_selector_type",
        "future_selector_file",
        "future_selector_test_file",
        "future_selector_smoke_fixture",
        "future_selector_smoke_checker",
    ):
        require(boundary.get(field) == selector_boundary.get(field), f"{field} must match selector smoke readiness")
    require(
        boundary.get("future_selector_file") == adapter_boundary.get("future_selector_file"),
        "future selector file must match adapter smoke readiness",
    )
    for field in (
        "implementation_task_card_created_in_this_slice",
        "formal_config_entry_created_in_this_slice",
        "selector_file_created_in_this_slice",
        "selector_test_created_in_this_slice",
        "selector_smoke_fixture_created_in_this_slice",
        "selector_smoke_checker_created_in_this_slice",
        "selector_implementation_allowed_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
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
    require(
        boundary.get("future_implementation_task_card")
        == "docs/task-cards/workflow-saved-draft-store-selector-implementation-v1-plan.md",
        "future implementation task card path drifted",
    )


def assert_gate_matrix(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("entry_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "entry gate ids drifted")
    for gate_id in (
        "selector_enablement_preconditions_consumed",
        "selector_smoke_readiness_consumed",
        "adapter_smoke_readiness_consumed",
        "selector_entry_review_defined",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must remain satisfied")
    blocked_gate_to_candidate = {
        "formal_config_candidate_blocked": "formal_store_config_entry",
        "selector_function_candidate_blocked": "select_workflow_saved_draft_store",
        "selector_tests_candidate_blocked": "selector_unit_tests",
        "selector_smoke_fixture_candidate_blocked": "selector_smoke_fixture",
    }
    for gate_id, candidate_id in blocked_gate_to_candidate.items():
        require(gates[gate_id].get("status") == "blocked", f"{gate_id} must remain blocked")
        require(gates[gate_id].get("candidate_id") == candidate_id, f"{gate_id} candidate drifted")
    for gate_id in (
        "repository_adapter_gate",
        "schema_artifact_file_gate",
        "production_auth_gate",
        "production_api_gate",
    ):
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage")
    require(
        gates["no_selector_implementation_artifacts_leaked"].get("status") == "required_now",
        "artifact leak gate status drifted",
    )


def assert_upstream_selector_gate_still_blocked() -> None:
    adapter_smoke = load_json(ADAPTER_SMOKE_READINESS_FIXTURE_PATH)
    gates = {
        str(gate.get("id") or ""): gate
        for gate in adapter_smoke.get("adapter_smoke_gate_matrix") or []
        if isinstance(gate, dict)
    }
    selector_gate = gates.get("selector_implementation_gate") or {}
    if selector_implementation_active(REPO_ROOT):
        require(selector_gate.get("status") == "satisfied", "upstream selector implementation gate must be satisfied")
        require(
            selector_gate.get("evidence") == "workflow-saved-draft-store-selector-smoke-v1",
            "upstream selector implementation gate evidence drifted",
        )
    else:
        require(selector_gate.get("status") == "not_satisfied", "upstream selector implementation gate must stay blocked")
    must_cover = set(selector_gate.get("must_cover") or [])
    for expected in (
        "formal store config entry",
        "SelectWorkflowSavedDraftStore",
        "selector unit tests",
        "selector smoke fixture",
    ):
        require(expected in must_cover, f"upstream selector gate missing {expected}")


def assert_candidates(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("candidate_id") or ""): row
        for row in fixture.get("selector_entry_candidates") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_CANDIDATE_IDS, "selector entry candidate ids drifted")
    for candidate_id, row in rows.items():
        require(row.get("entry_decision") == "blocked", f"{candidate_id} entry must remain blocked")
        require(row.get("trigger_satisfied_now") is False, f"{candidate_id} trigger must remain false")
        for field in ("task_card_allowed_now", "implementation_artifacts_allowed_now"):
            require(row.get(field) is False, f"{candidate_id} {field} must remain false")
        require(row.get("current_blockers"), f"{candidate_id} blockers must be listed")
        future_artifacts = [str(path) for path in row.get("future_artifacts") or []]
        require(future_artifacts, f"{candidate_id} future artifacts must be listed")
        for relative_path in future_artifacts:
            if relative_path in {"services/platform/internal/config/config.go", "services/platform/internal/config/config_test.go"}:
                require((REPO_ROOT / relative_path).exists(), f"{relative_path} should exist as current config source")
            elif selector_implementation_file_allowed(REPO_ROOT, relative_path):
                continue
            else:
                require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist yet")


def assert_failure_and_side_effect_policy(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(mapping.get("status") == "selector_entry_fail_closed_required", "failure mapping status drifted")
    missing = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing, f"missing failure codes: {missing}")
    missing_preserved = sorted(EXPECTED_PRESERVED_FAILURE_CODES - set(mapping.get("preserve_existing_failure_codes") or []))
    require(not missing_preserved, f"missing preserved failure codes: {missing_preserved}")
    for field in (
        "version_conflict_must_not_be_rewritten_as_selector_failure",
        "scope_denied_must_not_return_draft_payload",
        "not_found_must_not_fallback_to_sample",
        "repository_disabled_must_not_fallback_to_memory_dev",
        "unknown_mode_must_not_fallback_to_memory_dev",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")
    for field in (
        "missing_selector_success_allowed",
        "reserved_mode_success_allowed",
        "unknown_mode_success_allowed",
        "fallback_to_memory_dev_store_allowed",
        "fallback_to_sample_allowed",
        "fallback_to_fixture_allowed",
        "fallback_to_dev_http_route_allowed",
    ):
        require(mapping.get(field) is False, f"{field} must remain false")

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
    task_cards = {
        str(row.get("path") or ""): row
        for row in fixture.get("forbidden_task_card_matrix") or []
        if isinstance(row, dict)
    }
    require(
        set(task_cards) == {"docs/task-cards/workflow-saved-draft-store-selector-implementation-v1-plan.md"},
        "forbidden implementation task card drifted",
    )
    for relative_path, row in task_cards.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        if selector_implementation_file_allowed(REPO_ROOT, relative_path):
            continue
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    guard = fixture.get("implementation_artifact_guard") or {}
    require(guard.get("status") == "forbid_selector_implementation_artifacts", "artifact guard status drifted")
    require(EXPECTED_FUTURE_FILES.issubset(set(guard.get("future_files_must_not_exist") or [])), "future files drifted")
    for relative_path in guard.get("future_files_must_not_exist") or []:
        if selector_implementation_file_allowed(REPO_ROOT, str(relative_path)):
            continue
        require(not (REPO_ROOT / str(relative_path)).exists(), f"future artifact exists early: {relative_path}")
    sql_files = list((REPO_ROOT / "services/platform").rglob("*.sql"))
    require(not sql_files, f"SQL files must not be introduced: {sql_files}")

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
        "workflow_saved_draft_store_selector_implementation_entry_review_checker",
        "workflow_saved_draft_adapter_smoke_readiness_checker",
        "workflow_saved_draft_store_selector_smoke_readiness_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_create_config_entry",
        "does_not_create_selector",
        "does_not_create_selector_tests",
        "does_not_create_selector_smoke_fixture",
        "does_not_create_repository_adapter",
        "does_not_create_schema_artifact_file",
        "does_not_create_sql_or_migration",
        "does_not_connect_database",
        "does_not_call_oidc",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py"
    current_checker = (
        "checks/control_plane/check-workflow-saved-draft-store-selector-implementation-entry-review-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "adapter smoke readiness checker missing from check-repo.py")
    require(current_checker in check_repo, "check-repo.py must run store selector implementation entry review")
    require(next_checker in check_repo, "product surface recheck missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "selector entry review check must run after adapter smoke readiness and before product surface recheck",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_gate_matrix(fixture)
    assert_upstream_selector_gate_still_blocked()
    assert_candidates(fixture)
    assert_failure_and_side_effect_policy(fixture)
    assert_artifact_guard(fixture)
    assert_required_docs_and_testing(fixture)
    assert_check_repo_registration()
    print("workflow saved draft store selector implementation entry review v1 checks passed.")


if __name__ == "__main__":
    main()
