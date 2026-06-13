#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/product-surface-usage-gap-triage-v1.json"
PRODUCT_SURFACE_RECHECK_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/product-surface-readiness-implementation-trigger-recheck-v1.json"
)
SCHEMA_EVIDENCE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-artifact-evidence-v1.json"
)
IMPLEMENTATION_ENTRY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-entry-review-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "control-plane-read-formal-ui-readiness-close-v1": (
        REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-readiness-close-v1.json",
        "formal_ui_readiness_closed",
    ),
    "workflow-workspace-context-consistency-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-workspace-context-consistency-v1.json",
        "workflow_workspace_context_consistency_guarded",
    ),
    "product-surface-readiness-implementation-trigger-recheck-v1": (
        PRODUCT_SURFACE_RECHECK_PATH,
        "product_surface_readiness_trigger_recheck_defined",
    ),
    "control-plane-read-schema-artifact-evidence-v1": (
        SCHEMA_EVIDENCE_PATH,
        "schema_artifact_evidence_defined",
    ),
    "control-plane-read-implementation-entry-review-v1": (
        IMPLEMENTATION_ENTRY_PATH,
        "implementation_entry_review_defined",
    ),
}
EXPECTED_SURFACES = {
    "user_workspace",
    "workflow_review",
    "model_gateway_api_distribution",
    "admin_control_plane",
}
EXPECTED_ROUTE_IDS = {
    "tenant-summary-route",
    "application-summary-list-route",
    "api-key-summary-list-route",
    "quota-summary-route",
    "workflow-definition-summary-list-route",
    "run-record-summary-list-route",
    "audit-summary-list-route",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "new_product_ui_panel_required",
    "new_same_level_read_only_surface_created",
    "implementation_trigger_satisfied",
    "implementation_entry_opened",
    "schema_artifact_manifest_ready",
    "schema_artifact_files_created",
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
EXPECTED_FALSE_BOUNDARY_FIELDS = {
    "same_level_panel_created_in_this_slice",
    "implementation_task_card_created_in_this_slice",
    "schema_artifact_manifest_created_in_this_slice",
    "runtime_artifact_created_in_this_slice",
    "dev_server_started_in_this_slice",
    "browser_plugin_required",
    "database_connection_allowed_now",
    "production_auth_allowed_now",
    "write_allowed_now",
}
EXPECTED_CORRECTION_TARGETS = {
    "existing view model",
    "existing panel copy or grouping",
    "canonical response fixture",
    "workflowWorkspaceContext derivation",
    "read-side contract documentation",
    "task card or README read path",
}
EXPECTED_SPECIAL_GATE_REASONS = {
    "new API",
    "execution boundary",
    "production claim",
    "data format",
    "external provider risk",
    "high-risk capability",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "apps/radishmind-web/src/features/control-plane-read/productSurfaceUsageGapTriagePanel.tsx",
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
EXPECTED_SOURCE_ABSENT_LITERALS = {
    "productSurfaceUsageGapTriagePanel",
    "ControlPlaneReadSchemaMigrationRunner",
    "control_plane_read_schema_migration_runner.go",
    "control_plane_read_repository_interface.go",
    "control_plane_read_repository_adapter.go",
    "control_plane_read_store_selector.go",
    "control_plane_read_auth_middleware.go",
    "control-plane-read-adapter-smoke-v1.json",
    "check-control-plane-read-adapter-smoke-v1.py",
    "type ControlPlaneReadRepository interface",
    "func NewControlPlaneReadRepository",
    "database/sql",
    "CREATE TABLE",
    "ALTER TABLE",
    "CREATE INDEX",
    "DROP TABLE",
    "SELECT *",
    "INSERT INTO",
    "UPDATE ",
    "DELETE FROM",
    "services/platform/migrations/control_plane_read",
    "RADISHMIND_CONTROL_PLANE_READ_STORE",
    "SelectControlPlaneReadStore",
    "ControlPlaneReadStoreSelector",
    "oidc.Provider",
    "github.com/coreos/go-oidc",
    "ValidateToken",
    "VerifyToken",
    "NewRadishOIDCClient",
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
    require(fixture.get("kind") == "product_surface_usage_gap_triage_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "product-surface-usage-gap-triage-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "product_surface_usage_gap_triage_defined",
        "usage gap triage status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("triage_boundary") or {}
    require(
        boundary.get("status") == "usage_gap_triage_defined_no_runtime_change",
        "triage boundary status drifted",
    )
    require(
        boundary.get("decision") == "usage_walkthrough_complete_without_new_surface_or_runtime_artifact",
        "triage boundary decision drifted",
    )
    require(
        boundary.get("current_development_mode") == "product_usage_review_and_targeted_correction",
        "triage development mode drifted",
    )
    for field in EXPECTED_FALSE_BOUNDARY_FIELDS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_usage_walkthrough_matrix(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("surface_id") or ""): row
        for row in fixture.get("usage_walkthrough_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_SURFACES, "usage walkthrough surface ids drifted")

    recheck = load_json(PRODUCT_SURFACE_RECHECK_PATH)
    recheck_rows = {
        str(row.get("surface_id") or ""): row
        for row in recheck.get("product_surface_matrix") or []
        if isinstance(row, dict)
    }
    require(EXPECTED_SURFACES.issubset(set(recheck_rows)), "upstream product surface recheck drifted")

    for surface_id, row in rows.items():
        require(row.get("triage_status") == "walkthrough_reviewed", f"{surface_id} triage status drifted")
        require(
            row.get("current_gap_decision") == "no_new_actionable_gap_from_static_walkthrough",
            f"{surface_id} gap decision drifted",
        )
        require(recheck_rows[surface_id].get("implementation_trigger_decision") == "not_satisfied", surface_id)
        require(len(row.get("reading_path") or []) >= 5, f"{surface_id} reading path is too thin")
        require(len(row.get("decision_questions") or []) >= 4, f"{surface_id} decision questions missing")
        require(row.get("evidence_refs"), f"{surface_id} evidence refs required")
        require(row.get("uses_existing_surface") is True, f"{surface_id} must use existing surface")
        require(row.get("new_surface_requested") is False, f"{surface_id} must not request new surface")
        require(
            row.get("targeted_correction_when_gap_confirmed") is True,
            f"{surface_id} targeted correction gate drifted",
        )
        for relative_path in row.get("source_files") or []:
            require((REPO_ROOT / str(relative_path)).exists(), f"{surface_id} source missing: {relative_path}")


def assert_schema_route_crosscheck(fixture: dict[str, Any]) -> None:
    schema_evidence = load_json(SCHEMA_EVIDENCE_PATH)
    schema_slice = schema_evidence.get("slice") or {}
    schema_rows = {
        str(row.get("route_id") or ""): row
        for row in schema_evidence.get("route_schema_artifact_mapping") or []
        if isinstance(row, dict)
    }
    require(set(schema_rows) == EXPECTED_ROUTE_IDS, "schema evidence route mapping drifted")

    crosscheck = fixture.get("schema_route_crosscheck") or {}
    require(crosscheck.get("schema_evidence_ref") == "control-plane-read-schema-artifact-evidence-v1", "schema ref")
    require(crosscheck.get("schema_evidence_status") == schema_slice.get("status"), "schema evidence status drifted")
    require(crosscheck.get("route_mapping_count") == len(schema_rows), "route mapping count drifted")
    for field in (
        "schema_artifact_manifest_ready_now",
        "runtime_mapping_implemented_now",
        "sql_allowed_now",
        "repository_adapter_allowed_now",
        "implementation_trigger_satisfied_now",
    ):
        require(crosscheck.get(field) is False, f"{field} must remain false")

    rows = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_schema_trace") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_ROUTE_IDS, "route schema trace ids drifted")
    for route_id, row in rows.items():
        upstream = schema_rows[route_id]
        require(
            row.get("future_schema_artifact_id") == upstream.get("future_schema_artifact_id"),
            f"{route_id} schema artifact id drifted",
        )
        require(set(row.get("surface_refs") or []).issubset(EXPECTED_SURFACES), f"{route_id} surface refs")
        require(row.get("surface_refs"), f"{route_id} must reference a product surface")
        for field in (
            "schema_artifact_ready_now",
            "runtime_mapping_implemented",
            "sql_allowed_now",
            "repository_adapter_allowed_now",
        ):
            require(row.get(field) is False, f"{route_id} {field} must remain false")
        require(upstream.get("schema_artifact_ready") is False, f"{route_id} upstream schema artifact ready")
        require(upstream.get("runtime_mapping_implemented") is False, f"{route_id} upstream runtime mapping")
        require(upstream.get("sql_allowed_now") is False, f"{route_id} upstream SQL")
        require(upstream.get("repository_adapter_allowed_now") is False, f"{route_id} upstream adapter")


def assert_targeted_correction_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("targeted_correction_policy") or {}
    require(policy.get("status") == "targeted_correction_policy_defined", "targeted correction policy status")
    require(policy.get("same_level_panel_default_allowed") is False, "same-level panel default must stay closed")
    require(policy.get("ordinary_copy_or_layout_uses_existing_gates") is True, "ordinary UI gate policy drifted")
    require(policy.get("new_checker_required_for_ordinary_copy_or_layout") is False, "ordinary checker policy drifted")
    require(
        policy.get("confirmed_reading_gap_required_before_ui_change") is True,
        "confirmed reading gap gate required",
    )
    require(
        EXPECTED_CORRECTION_TARGETS.issubset(set(policy.get("allowed_correction_targets") or [])),
        "allowed correction targets drifted",
    )
    require(
        EXPECTED_SPECIAL_GATE_REASONS.issubset(set(policy.get("special_gate_required_when") or [])),
        "special gate reasons drifted",
    )
    forbidden = set(policy.get("forbidden_correction_targets_now") or [])
    for literal in ("new same-level read-only product panel", "repository adapter", "store selector"):
        require(literal in forbidden, f"forbidden correction target missing: {literal}")


def assert_implementation_wait_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("implementation_wait_policy") or {}
    entry = load_json(IMPLEMENTATION_ENTRY_PATH)
    boundary = entry.get("implementation_entry_boundary") or {}
    require(policy.get("status") == "wait_for_satisfied_trigger", "implementation wait policy status")
    require(
        policy.get("implementation_entry_status") == entry.get("slice", {}).get("status"),
        "implementation entry status drifted",
    )
    require(policy.get("selected_implementation_track_now") == "none", "selected implementation track must be none")
    require(policy.get("implementation_task_card_allowed_now") is False, "implementation task card must be blocked")
    require(policy.get("parallel_implementation_tracks_allowed_now") is False, "parallel tracks must be blocked")
    require(boundary.get("decision") == "implementation_entry_not_opened", "entry review must remain closed")
    require(boundary.get("selected_implementation_track_now") == "none", "entry selected track drifted")

    candidates = entry.get("implementation_entry_candidates") or []
    require(candidates, "entry review candidates missing")
    for candidate in candidates:
        require(candidate.get("trigger_satisfied_now") is False, "implementation trigger must remain false")
        require(candidate.get("task_card_allowed_now") is False, "candidate task card must remain blocked")
        require(
            candidate.get("implementation_artifacts_allowed_now") is False,
            "candidate artifacts must remain blocked",
        )


def assert_forbidden_artifacts_and_sources(fixture: dict[str, Any]) -> None:
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

    configured = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_SOURCE_ABSENT_LITERALS.issubset(configured), "source absent literals drifted")
    for root in (REPO_ROOT / "services/platform/internal", REPO_ROOT / "apps/radishmind-web/src"):
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured:
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r}",
                )


def assert_fallback_and_side_effect_policies(fixture: dict[str, Any]) -> None:
    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "reading_gap_promoted_to_runtime_success_allowed",
        "schema_evidence_promoted_to_manifest_allowed",
        "blocked_candidate_promoted_to_task_card_allowed",
        "repository_mode_fallback_to_fixture_fake_store_allowed",
        "database_mode_fallback_to_fixture_fake_store_allowed",
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
    previous_checker = "check-control-plane-read-implementation-entry-review-v1.py"
    current_checker = "check-product-surface-usage-gap-triage-v1.py"
    require(current_checker in check_repo, "check-repo.py must run product surface usage gap triage")
    require(check_repo.index(previous_checker) < check_repo.index(current_checker), "usage triage must run after entry review")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_usage_walkthrough_matrix(fixture)
    assert_schema_route_crosscheck(fixture)
    assert_targeted_correction_policy(fixture)
    assert_implementation_wait_policy(fixture)
    assert_forbidden_artifacts_and_sources(fixture)
    assert_fallback_and_side_effect_policies(fixture)
    assert_references_and_check_repo(fixture)
    print("product surface usage gap triage v1 checks passed.")


if __name__ == "__main__":
    main()
