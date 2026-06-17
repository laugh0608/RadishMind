#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-review-v1.json"
)
SCHEMA_EVIDENCE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-evidence-v1.json"
)
SCHEMA_MANIFEST_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-manifest-v1.json"
)
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json"
)
ADAPTER_SMOKE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-readiness-v1.json"
)
SELECTOR_ENTRY_REVIEW_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/workflow-saved-draft-store-selector-implementation-entry-review-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-saved-draft-schema-artifact-evidence-v1": (
        SCHEMA_EVIDENCE_FIXTURE_PATH,
        "draft_schema_artifact_evidence_defined",
    ),
    "workflow-saved-draft-schema-artifact-manifest-v1": (
        SCHEMA_MANIFEST_FIXTURE_PATH,
        "draft_schema_artifact_manifest_defined",
    ),
    "workflow-saved-draft-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "draft_repository_adapter_implementation_plan_defined",
    ),
    "workflow-saved-draft-adapter-smoke-readiness-v1": (
        ADAPTER_SMOKE_READINESS_FIXTURE_PATH,
        "draft_adapter_smoke_readiness_defined",
    ),
    "workflow-saved-draft-store-selector-implementation-entry-review-v1": (
        SELECTOR_ENTRY_REVIEW_FIXTURE_PATH,
        "draft_store_selector_implementation_entry_review_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "schema_artifact_materialization_entry_opened",
    "schema_artifact_materialization_task_card_created",
    "schema_artifact_manifest_file_created",
    "schema_artifact_files_created",
    "ddl_review_artifact_created",
    "rollback_evidence_artifact_created",
    "migration_smoke_artifact_created",
    "schema_version_table_created",
    "sql_migration_created",
    "migration_runner_implemented",
    "database_schema_ready",
    "database_connection_ready",
    "database_query_ready",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "store_selector_implemented",
    "selector_smoke_fixture_created",
    "adapter_smoke_fixture_created",
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
    "schema_artifact_evidence_consumed",
    "schema_artifact_manifest_consumed",
    "repository_adapter_plan_consumed",
    "selector_entry_review_consumed",
    "adapter_smoke_readiness_consumed",
    "materialization_review_defined",
    "migration_root_candidate_blocked",
    "manifest_file_candidate_blocked",
    "ddl_review_candidate_blocked",
    "rollback_evidence_candidate_blocked",
    "migration_smoke_candidate_blocked",
    "store_selector_gate",
    "repository_adapter_gate",
    "production_auth_gate",
    "database_connection_gate",
    "production_api_gate",
    "no_schema_artifact_materialization_artifacts_leaked",
}
EXPECTED_CANDIDATE_IDS = {
    "migration_root",
    "manifest_file",
    "ddl_review_artifact",
    "rollback_evidence_artifact",
    "migration_smoke_artifact",
}
EXPECTED_FUTURE_FILES = {
    "docs/task-cards/workflow-saved-draft-schema-artifact-materialization-v1-plan.md",
    "services/platform/migrations/workflow_saved_drafts",
    "services/platform/migrations/workflow_saved_drafts/manifest.json",
    "services/platform/migrations/workflow_saved_drafts/ddl-review.md",
    "services/platform/migrations/workflow_saved_drafts/rollback-evidence.json",
    "services/platform/migrations/workflow_saved_drafts/migration-smoke.json",
    "services/platform/internal/httpapi/workflow_saved_draft_schema_migration_runner.go",
}
EXPECTED_FAILURE_CODES = {
    "draft_schema_migration_not_applied",
    "draft_store_schema_version_mismatch",
    "draft_store_migration_unavailable",
    "draft_store_contract_mismatch",
    "repository_store_disabled",
    "invalid_draft_store_mode",
}
EXPECTED_PRESERVED_FAILURE_CODES = {
    "draft_version_conflict",
    "draft_scope_denied",
    "draft_not_found",
    "draft_store_unavailable",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "repository_write_count=0",
    "database_write_count=0",
    "migration_apply_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "database_write",
    "schema_migration_apply",
    "repository_write",
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


def schema_manifest_boundary() -> dict[str, Any]:
    return load_json(SCHEMA_MANIFEST_FIXTURE_PATH).get("schema_artifact_manifest_boundary") or {}


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
        fixture.get("kind") == "workflow_saved_draft_schema_artifact_materialization_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-schema-artifact-materialization-review-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_schema_artifact_materialization_review_defined",
        "schema artifact materialization review status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("materialization_review_boundary") or {}
    manifest_boundary = schema_manifest_boundary()
    require(
        boundary.get("status") == "schema_artifact_materialization_review_defined_no_files",
        "materialization boundary status drifted",
    )
    require(
        boundary.get("decision") == "schema_artifact_materialization_entry_not_opened",
        "materialization decision drifted",
    )
    require(
        boundary.get("current_development_mode") == "materialization_review_only",
        "development mode drifted",
    )
    require(boundary.get("selected_implementation_track_now") == "none", "selected track must remain none")
    for field in (
        "future_schema_artifact_id",
        "future_manifest_version",
        "future_store_schema_version",
        "future_migration_root",
        "future_manifest_file",
        "future_ddl_review_file",
        "future_rollback_evidence_file",
        "future_migration_smoke_file",
        "future_schema_version_table",
    ):
        require(boundary.get(field) == manifest_boundary.get(field), f"{field} must match schema manifest")
    require(
        boundary.get("future_implementation_task_card")
        == "docs/task-cards/workflow-saved-draft-schema-artifact-materialization-v1-plan.md",
        "future materialization task card path drifted",
    )
    for field in (
        "implementation_task_card_created_in_this_slice",
        "migration_root_created_in_this_slice",
        "manifest_file_created_in_this_slice",
        "ddl_review_artifact_created_in_this_slice",
        "rollback_evidence_artifact_created_in_this_slice",
        "migration_smoke_artifact_created_in_this_slice",
        "schema_version_table_created_in_this_slice",
        "sql_files_created_in_this_slice",
        "migration_runner_created_in_this_slice",
        "database_connection_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "repository_interface_created_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "selector_smoke_allowed_in_this_slice",
        "adapter_smoke_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "publish_or_run_allowed_in_this_slice",
        "executor_allowed_in_this_slice",
        "writeback_or_replay_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gate_matrix(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("materialization_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "materialization gate ids drifted")
    for gate_id in (
        "schema_artifact_evidence_consumed",
        "schema_artifact_manifest_consumed",
        "repository_adapter_plan_consumed",
        "selector_entry_review_consumed",
        "adapter_smoke_readiness_consumed",
        "materialization_review_defined",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must remain satisfied")
    blocked_gate_to_candidate = {
        "migration_root_candidate_blocked": "migration_root",
        "manifest_file_candidate_blocked": "manifest_file",
        "ddl_review_candidate_blocked": "ddl_review_artifact",
        "rollback_evidence_candidate_blocked": "rollback_evidence_artifact",
        "migration_smoke_candidate_blocked": "migration_smoke_artifact",
    }
    for gate_id, candidate_id in blocked_gate_to_candidate.items():
        require(gates[gate_id].get("status") == "blocked", f"{gate_id} must remain blocked")
        require(gates[gate_id].get("candidate_id") == candidate_id, f"{gate_id} candidate drifted")
    for gate_id in (
        "store_selector_gate",
        "repository_adapter_gate",
        "production_auth_gate",
        "database_connection_gate",
        "production_api_gate",
    ):
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage")
    require(
        gates["no_schema_artifact_materialization_artifacts_leaked"].get("status") == "required_now",
        "artifact leak gate status drifted",
    )


def assert_upstream_schema_artifact_gate_still_blocked() -> None:
    adapter_smoke = load_json(ADAPTER_SMOKE_READINESS_FIXTURE_PATH)
    gates = {
        str(gate.get("id") or ""): gate
        for gate in adapter_smoke.get("adapter_smoke_gate_matrix") or []
        if isinstance(gate, dict)
    }
    gate = gates.get("schema_artifact_materialization_gate") or {}
    require(gate.get("status") == "not_satisfied", "upstream schema artifact materialization gate must stay blocked")
    must_cover = set(gate.get("must_cover") or [])
    for expected in (
        "migration root",
        "manifest file",
        "DDL review evidence",
        "rollback evidence",
        "migration smoke evidence",
    ):
        require(expected in must_cover, f"upstream materialization gate missing {expected}")


def assert_candidates(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("candidate_id") or ""): row
        for row in fixture.get("materialization_entry_candidates") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_CANDIDATE_IDS, "materialization candidate ids drifted")
    for candidate_id, row in rows.items():
        require(row.get("entry_decision") == "blocked", f"{candidate_id} entry must remain blocked")
        require(row.get("trigger_satisfied_now") is False, f"{candidate_id} trigger must remain false")
        for field in ("task_card_allowed_now", "implementation_artifacts_allowed_now"):
            require(row.get(field) is False, f"{candidate_id} {field} must remain false")
        require(row.get("current_blockers"), f"{candidate_id} blockers must be listed")
        future_artifacts = [str(path) for path in row.get("future_artifacts") or []]
        require(future_artifacts, f"{candidate_id} future artifacts must be listed")
        for relative_path in future_artifacts:
            require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist yet")


def assert_failure_and_side_effect_policy(fixture: dict[str, Any]) -> None:
    mapping = fixture.get("failure_mapping") or {}
    require(
        mapping.get("status") == "schema_artifact_materialization_fail_closed_required",
        "failure mapping status drifted",
    )
    missing = sorted(EXPECTED_FAILURE_CODES - set(mapping.get("required_failure_codes") or []))
    require(not missing, f"missing failure codes: {missing}")
    missing_preserved = sorted(EXPECTED_PRESERVED_FAILURE_CODES - set(mapping.get("preserve_existing_failure_codes") or []))
    require(not missing_preserved, f"missing preserved failure codes: {missing_preserved}")
    for field in (
        "version_conflict_must_not_be_rewritten_as_migration_failure",
        "scope_denied_must_not_return_draft_payload",
        "not_found_must_not_fallback_to_sample",
        "missing_manifest_must_fail_closed",
        "missing_ddl_review_must_fail_closed",
    ):
        require(mapping.get(field) is True, f"{field} must remain true")
    for field in (
        "missing_schema_artifact_success_allowed",
        "missing_manifest_success_allowed",
        "ddl_review_missing_success_allowed",
        "rollback_evidence_missing_success_allowed",
        "migration_smoke_missing_success_allowed",
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
        set(task_cards) == {"docs/task-cards/workflow-saved-draft-schema-artifact-materialization-v1-plan.md"},
        "forbidden materialization task card drifted",
    )
    for relative_path, row in task_cards.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    guard = fixture.get("implementation_artifact_guard") or {}
    require(
        guard.get("status") == "forbid_schema_artifact_materialization_artifacts",
        "artifact guard status drifted",
    )
    require(EXPECTED_FUTURE_FILES.issubset(set(guard.get("future_files_must_not_exist") or [])), "future files drifted")
    for relative_path in guard.get("future_files_must_not_exist") or []:
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
        "workflow_saved_draft_schema_artifact_materialization_review_checker",
        "workflow_saved_draft_schema_artifact_manifest_checker",
        "workflow_saved_draft_schema_artifact_evidence_checker",
        "workflow_saved_draft_adapter_smoke_readiness_checker",
        "workflow_saved_draft_store_selector_implementation_entry_review_checker",
        "./scripts/check-repo.sh --fast",
        "./scripts/check-repo.sh",
    ):
        require(expected_check in required, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_create_implementation_task_card",
        "does_not_create_migration_root",
        "does_not_create_schema_artifact_file",
        "does_not_create_manifest_file",
        "does_not_create_ddl_review",
        "does_not_create_rollback_evidence",
        "does_not_create_migration_smoke",
        "does_not_create_sql_or_migration",
        "does_not_connect_database",
        "does_not_create_repository_adapter",
        "does_not_create_store_selector",
        "does_not_call_oidc",
        "does_not_create_production_api",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")


def assert_check_repo_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = (
        "checks/control_plane/check-workflow-saved-draft-store-selector-implementation-entry-review-v1.py"
    )
    current_checker = (
        "checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-review-v1.py"
    )
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(previous_checker in check_repo, "store selector entry review checker missing from check-repo.py")
    require(current_checker in check_repo, "check-repo.py must run schema artifact materialization review")
    require(next_checker in check_repo, "product surface recheck missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "materialization review check must run after selector entry review and before product surface recheck",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_gate_matrix(fixture)
    assert_upstream_schema_artifact_gate_still_blocked()
    assert_candidates(fixture)
    assert_failure_and_side_effect_policy(fixture)
    assert_artifact_guard(fixture)
    assert_required_docs_and_testing(fixture)
    assert_check_repo_registration()
    print("workflow saved draft schema artifact materialization review v1 checks passed.")


if __name__ == "__main__":
    main()
