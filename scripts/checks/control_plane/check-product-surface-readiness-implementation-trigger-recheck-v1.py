#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/product-surface-readiness-implementation-trigger-recheck-v1.json"
)
TRIGGER_REVIEW_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-trigger-review-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "control-plane-read-formal-ui-readiness-close-v1": (
        REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-readiness-close-v1.json",
        "formal_ui_readiness_closed",
    ),
    "control-plane-read-product-sample-consistency-v1": (
        REPO_ROOT / "scripts/checks/fixtures/control-plane-read-product-sample-consistency-v1.json",
        "product_sample_consistency_guarded",
    ),
    "workflow-function-surface-readiness-close-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-readiness-close-v1.json",
        "workflow_function_surface_readiness_closed",
    ),
    "workflow-workspace-context-consistency-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-workspace-context-consistency-v1.json",
        "workflow_workspace_context_consistency_guarded",
    ),
    "control-plane-read-implementation-trigger-review-v1": (
        TRIGGER_REVIEW_PATH,
        "implementation_trigger_review_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "new_product_ui_panel_required",
    "implementation_trigger_satisfied",
    "production_api_consumer_ready",
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "repository_interface_ready",
    "repository_adapter_ready",
    "repository_implementation_ready",
    "store_selector_ready",
    "radish_oidc_client_ready",
    "auth_middleware_ready",
    "token_validation_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "billing_ready",
    "cost_ledger_ready",
    "production_gateway_ready",
    "production_secret_resolver_ready",
    "deployment_preflight_ready",
    "workflow_builder_ready",
    "workflow_executor_ready",
    "confirmation_flow_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "run_resume_ready",
    "production_ready",
}
EXPECTED_SURFACES = {
    "user_workspace",
    "workflow_review",
    "model_gateway_api_distribution",
    "admin_control_plane",
}
EXPECTED_TRIGGER_CANDIDATES = {
    "schema_artifact_manifest_implementation",
    "store_selector_smoke_implementation",
    "production_auth_implementation",
    "adapter_smoke_execution",
}
EXPECTED_FALSE_BOUNDARY_FIELDS = {
    "new_read_only_surface_allowed_by_default",
    "implementation_task_card_allowed_now",
    "direct_runtime_implementation_allowed_now",
    "live_backend_allowed_now",
    "dev_server_started_by_ai",
    "database_connection_allowed_now",
    "production_auth_allowed_now",
    "write_allowed_now",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "apps/radishmind-web/src/features/control-plane-read/productSurfaceReadinessRecheckPanel.tsx",
    "services/platform/migrations/control_plane_read/manifest.json",
    "services/platform/internal/httpapi/control_plane_read_store_selector.go",
    "services/platform/internal/httpapi/control_plane_read_auth_middleware.go",
    "services/platform/internal/httpapi/control_plane_read_repository_adapter.go",
    "scripts/checks/fixtures/control-plane-read-adapter-smoke-v1.json",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "dev_server_start_count=0",
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "dev_server_start",
    "database_write",
    "api_key_issue",
    "quota_mutation",
    "deployment_preflight",
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
        fixture.get("kind") == "product_surface_readiness_implementation_trigger_recheck_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "product-surface-readiness-implementation-trigger-recheck-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "product_surface_readiness_trigger_recheck_defined",
        "product surface recheck status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_recheck_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("recheck_boundary") or {}
    require(
        boundary.get("status") == "product_surface_readiness_recheck_defined_no_runtime_change",
        "recheck boundary status drifted",
    )
    require(
        boundary.get("decision") == "no_new_reading_gap_and_no_implementation_trigger_satisfied",
        "recheck boundary decision drifted",
    )
    for field in EXPECTED_FALSE_BOUNDARY_FIELDS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_product_surfaces(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("surface_id") or ""): row
        for row in fixture.get("product_surface_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_SURFACES, "product surface ids drifted")
    for surface_id, row in rows.items():
        require(row.get("status") == "readiness_reviewed", f"{surface_id} status drifted")
        require(
            row.get("reading_gap_decision") == "no_new_reading_gap_identified",
            f"{surface_id} reading gap decision drifted",
        )
        require(
            row.get("implementation_trigger_decision") == "not_satisfied",
            f"{surface_id} implementation trigger decision drifted",
        )
        require(
            row.get("default_next_action") == "targeted_correction_only_after_real_reading_gap",
            f"{surface_id} next action drifted",
        )
        require(row.get("evidence_refs"), f"{surface_id} must list evidence refs")
        require(row.get("blocked_capabilities"), f"{surface_id} must list blocked capabilities")
        for relative_path in row.get("source_files") or []:
            require((REPO_ROOT / str(relative_path)).exists(), f"{surface_id} source missing: {relative_path}")


def assert_trigger_recheck(fixture: dict[str, Any]) -> None:
    trigger_fixture = load_json(TRIGGER_REVIEW_PATH)
    trigger_rows = {
        str(row.get("candidate_id") or ""): row
        for row in trigger_fixture.get("implementation_trigger_matrix") or []
        if isinstance(row, dict)
    }
    rows = {
        str(row.get("candidate_id") or ""): row
        for row in fixture.get("implementation_trigger_recheck") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_TRIGGER_CANDIDATES, "trigger recheck candidate ids drifted")
    require(EXPECTED_TRIGGER_CANDIDATES.issubset(set(trigger_rows)), "upstream trigger review candidates drifted")
    for candidate_id, row in rows.items():
        require(row.get("source") == "control-plane-read-implementation-trigger-review-v1", f"{candidate_id} source")
        require(row.get("decision") == "not_satisfied", f"{candidate_id} decision must remain not_satisfied")
        require(trigger_rows[candidate_id].get("decision") == "not_satisfied", f"{candidate_id} upstream drifted")
        require(row.get("task_card_allowed_now") is False, f"{candidate_id} task card must remain blocked")
        require(row.get("implementation_artifacts_allowed_now") is False, f"{candidate_id} artifacts blocked")


def assert_policies(fixture: dict[str, Any]) -> None:
    gap_policy = fixture.get("reading_gap_policy") or {}
    require(gap_policy.get("same_level_panel_default_allowed") is False, "same-level panels must stay opt-in")
    require(gap_policy.get("targeted_correction_requires_real_reading_gap") is True, "real gap gate required")
    require(gap_policy.get("new_checker_required_for_ordinary_copy_or_layout") is False, "ordinary UI checker drift")

    dev_server_policy = fixture.get("dev_server_policy") or {}
    require(dev_server_policy.get("real_integration_server_owner") == "developer", "dev server owner drifted")
    require(
        dev_server_policy.get("ai_auto_start_requires_explicit_request") is True,
        "AI auto start must require explicit request",
    )
    require(
        dev_server_policy.get("background_process_must_be_closed_before_commit") is True,
        "background process close rule required",
    )
    require(dev_server_policy.get("this_slice_starts_dev_server") is False, "this check must not start dev server")

    side_effects = fixture.get("no_side_effect_policy") or {}
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(side_effects.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effects.get("forbidden_side_effects") or [])),
        "missing forbidden side effects",
    )


def assert_forbidden_artifacts(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(row.get("path") or ""): row
        for row in fixture.get("forbidden_artifact_matrix") or []
        if isinstance(row, dict)
    }
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifact paths drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")
    require(not (REPO_ROOT / "services/platform/migrations/control_plane_read").exists(), "migration root leaked")
    sql_files = list((REPO_ROOT / "services/platform").rglob("*.sql"))
    require(not sql_files, f"SQL files must not be introduced: {sql_files}")


def assert_references_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    for relative_path, required_literals in (fixture.get("required_source_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing source literals: {missing}")

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing doc literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-workspace-context-consistency-v1.py"
    trigger_checker = "check-control-plane-read-implementation-trigger-review-v1.py"
    current_checker = "check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(current_checker in check_repo, "check-repo.py must run product surface recheck")
    require(check_repo.index(trigger_checker) < check_repo.index(current_checker), "trigger review must run first")
    require(check_repo.index(previous_checker) < check_repo.index(current_checker), "workflow context must run first")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_recheck_boundary(fixture)
    assert_product_surfaces(fixture)
    assert_trigger_recheck(fixture)
    assert_policies(fixture)
    assert_forbidden_artifacts(fixture)
    assert_references_and_check_repo(fixture)
    print("product surface readiness / implementation trigger recheck v1 checks passed.")


if __name__ == "__main__":
    main()
