#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-boundary-v1.json"
WORKFLOW_BOUNDARY_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-definition-run-record-boundary.json"
FORMAL_UI_CLOSE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-readiness-close-v1.json"
IMPLEMENTATION_TRIGGER_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-trigger-review-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-definition-run-record-boundary": (WORKFLOW_BOUNDARY_PATH, "governance_boundary_satisfied"),
    "control-plane-read-formal-ui-readiness-close-v1": (FORMAL_UI_CLOSE_PATH, "formal_ui_readiness_closed"),
    "control-plane-read-implementation-trigger-review-v1": (
        IMPLEMENTATION_TRIGGER_PATH,
        "implementation_trigger_review_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "runtime_api_implemented",
    "production_api_consumer_ready",
    "workflow_builder_ready",
    "workflow_definition_mutation_ready",
    "workflow_executor_ready",
    "node_executor_ready",
    "tool_executor_ready",
    "durable_run_store_ready",
    "materialized_result_reader_ready",
    "confirmation_flow_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "run_resume_ready",
    "database_ready",
    "radish_oidc_ready",
    "production_ready",
}
EXPECTED_SURFACE_STATUSES = {
    "application-detail": "read_only_surface_defined",
    "workflow-definition-detail": "read_only_surface_defined",
    "run-detail": "read_only_surface_defined",
    "tool-action-preview": "blocked_preview_surface_defined",
    "confirmation-placeholder": "placeholder_surface_defined",
}
EXPECTED_SURFACE_FIELDS = {
    "application-detail": {
        "application_id",
        "tenant_ref",
        "owner_subject_ref",
        "application_type",
        "display_name",
        "lifecycle_status",
        "provider_profile_ref",
        "risk_summary",
        "latest_run_ref",
        "route_ref",
        "audit_ref",
    },
    "workflow-definition-detail": {
        "workflow_definition_id",
        "tenant_ref",
        "application_ref",
        "version",
        "definition_status",
        "nodes",
        "edges",
        "input_schema_summary",
        "output_schema_summary",
        "risk_level",
        "requires_confirmation_capable",
        "route_ref",
        "audit_ref",
    },
    "run-detail": {
        "run_id",
        "tenant_ref",
        "workflow_definition_ref",
        "application_ref",
        "status",
        "state_timeline",
        "input_summary",
        "output_summary",
        "cost_summary",
        "token_summary",
        "trace_id",
        "failure_code",
        "audit_refs",
    },
    "tool-action-preview": {
        "tool_action_id",
        "run_id",
        "node_execution_ref",
        "tool_ref",
        "action_kind",
        "risk_level",
        "requires_confirmation",
        "policy_reason",
        "blocked_state",
        "audit_ref",
    },
    "confirmation-placeholder": {
        "confirmation_placeholder_id",
        "required_action_ref",
        "run_id",
        "risk_summary",
        "required_decision_shape",
        "human_review_required",
        "disabled_reason",
        "audit_ref",
    },
}
EXPECTED_DRAFT_ROUTE_IDS = {
    "application-detail-read-draft",
    "workflow-definition-detail-read-draft",
    "run-detail-read-draft",
    "tool-action-preview-read-draft",
    "confirmation-placeholder-read-draft",
}
EXPECTED_NEXT_SLICES = {
    "workflow-definition-detail-read-v1",
    "workflow-run-detail-read-v1",
    "workflow-blocked-action-preview-v1",
}
EXPECTED_NEXT_SLICE_FORBIDDEN_TERMS = {
    "workflow-definition-detail-read-v1": {"builder", "executor"},
    "workflow-run-detail-read-v1": {"replay", "resume", "durable run store"},
    "workflow-blocked-action-preview-v1": {"tool execution", "confirmation decision", "business writeback"},
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/platform/internal/workflow/executor.go",
    "services/platform/internal/httpapi/workflow_execution.go",
    "services/platform/internal/httpapi/workflow_definition_mutation.go",
    "services/platform/internal/httpapi/workflow_confirmation_decision.go",
    "services/platform/internal/httpapi/workflow_replay.go",
    "services/platform/internal/httpapi/workflow_writeback.go",
    "services/platform/internal/httpapi/workflow_durable_run_store.go",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "runtime_api_added_count=0",
    "executor_call_count=0",
    "tool_execution_count=0",
    "confirmation_decision_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
    "database_write_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "workflow_execution",
    "node_execution",
    "tool_execution",
    "confirmation_decision",
    "business_writeback",
    "run_replay",
    "run_resume",
    "database_write",
    "secret_write",
}
EXPECTED_STOP_LINES = {
    "no new runtime API",
    "no workflow builder",
    "no workflow definition mutation",
    "no workflow executor",
    "no node executor",
    "no tool executor",
    "no confirmation decision",
    "no business writeback",
    "no replay or resume",
    "no durable run store",
    "no materialized result reader",
    "no database or production auth implementation",
    "no production readiness claim",
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
    require(fixture.get("kind") == "workflow_function_surface_boundary_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-function-surface-boundary-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "function_surface_boundary_defined",
        "workflow function surface boundary status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_boundary_scope(fixture: dict[str, Any]) -> None:
    scope = fixture.get("boundary_scope") or {}
    require(
        scope.get("status") == "field_boundary_only_no_runtime_implementation",
        "boundary scope status drifted",
    )
    require(scope.get("offline_fixture_allowed_now") is True, "offline fixture path must remain allowed")
    require(scope.get("future_fake_store_dev_path_allowed") is True, "future fake-store dev path must remain allowed")
    for field in (
        "runtime_api_allowed_now",
        "production_api_consumer_allowed_now",
        "write_allowed_now",
        "executor_allowed_now",
        "confirmation_decision_allowed_now",
        "business_writeback_allowed_now",
        "replay_allowed_now",
    ):
        require(scope.get(field) is False, f"{field} must remain false")


def assert_surfaces(fixture: dict[str, Any]) -> None:
    surfaces = {
        str(surface.get("id") or ""): surface
        for surface in fixture.get("function_surfaces") or []
        if isinstance(surface, dict)
    }
    require(set(surfaces) == set(EXPECTED_SURFACE_STATUSES), f"surface ids drifted: {sorted(surfaces)}")
    for surface_id, expected_status in EXPECTED_SURFACE_STATUSES.items():
        surface = surfaces[surface_id]
        require(surface.get("status") == expected_status, f"{surface_id} status drifted")
        require(
            surface.get("source_policy") == "offline_fixture_or_future_fake_store_dev_path_only",
            f"{surface_id} source policy drifted",
        )
        missing_fields = sorted(EXPECTED_SURFACE_FIELDS[surface_id] - set(surface.get("required_fields") or []))
        require(not missing_fields, f"{surface_id} missing fields: {missing_fields}")
        require(surface.get("display_expectations"), f"{surface_id} must define display expectations")
        require(str(surface.get("boundary") or "").strip(), f"{surface_id} boundary is required")
        capabilities = surface.get("capabilities") or {}
        require(capabilities.get("can_read") is True, f"{surface_id} must remain readable")
        for field in (
            "can_mutate",
            "can_execute",
            "can_write_business_truth",
            "can_replay",
            "can_decide_confirmation",
        ):
            require(capabilities.get(field) is False, f"{surface_id} {field} must remain false")


def assert_draft_routes(fixture: dict[str, Any]) -> None:
    routes = {
        str(route.get("draft_route_id") or ""): route
        for route in fixture.get("draft_route_matrix") or []
        if isinstance(route, dict)
    }
    require(set(routes) == EXPECTED_DRAFT_ROUTE_IDS, "draft route ids drifted")
    for route_id, route in routes.items():
        require(route.get("surface_id") in EXPECTED_SURFACE_STATUSES, f"{route_id} surface drifted")
        require(route.get("method") == "GET", f"{route_id} must remain GET-only draft")
        require(route.get("contract_status") == "draft_only_no_api", f"{route_id} must remain draft-only")
        require(route.get("implemented_now") is False, f"{route_id} must not be implemented in this slice")
        require(str(route.get("path_template") or "").startswith("/v1/user-workspace/"), f"{route_id} path drifted")


def assert_risk_next_and_stop_lines(fixture: dict[str, Any]) -> None:
    policy = fixture.get("risk_and_blocked_policy") or {}
    require({"low", "medium", "high"}.issubset(set(policy.get("allowed_risk_levels") or [])), "risk levels drifted")
    require(
        {
            "blocked_read_only_boundary",
            "blocked_confirmation_required",
            "blocked_executor_not_available",
            "blocked_writeback_forbidden",
        }.issubset(set(policy.get("blocked_states") or [])),
        "blocked states drifted",
    )
    require(policy.get("requires_confirmation_must_be_visible") is True, "requires confirmation must remain visible")
    require(policy.get("policy_reason_must_be_visible") is True, "policy reason must remain visible")

    next_slices = {
        str(row.get("id") or ""): row
        for row in fixture.get("next_slice_candidates") or []
        if isinstance(row, dict)
    }
    require(set(next_slices) == EXPECTED_NEXT_SLICES, "next slice candidates drifted")
    for slice_id, row in next_slices.items():
        require(
            row.get("allowed_source") == "offline_fixture_or_fake_store_dev_path",
            f"{slice_id} allowed source drifted",
        )
        forbidden_scope = str(row.get("forbidden_scope") or "")
        for term in EXPECTED_NEXT_SLICE_FORBIDDEN_TERMS[slice_id]:
            require(term in forbidden_scope, f"{slice_id} must forbid {term}")

    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_forbidden_artifacts_and_side_effects(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(row.get("path") or ""): row
        for row in fixture.get("forbidden_artifact_matrix") or []
        if isinstance(row, dict)
    }
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifact paths drifted")
    for relative_path, row in artifacts.items():
        require(str(row.get("reason") or "").strip(), f"{relative_path} reason is required")
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in boundary slice")

    side_effects = fixture.get("no_side_effect_policy") or {}
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(side_effects.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effects.get("forbidden_side_effects") or [])),
        "missing forbidden side effects",
    )


def assert_docs_check_repo_and_no_leak(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-control-plane-read-implementation-trigger-review-v1.py"
    current_checker = "check-workflow-function-surface-boundary-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow function surface boundary check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow function surface check must run after implementation trigger review",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")

    configured = set(fixture.get("source_absent_literals") or [])
    require(configured, "source absent literals must be configured")
    roots = (REPO_ROOT / "services/platform/internal", REPO_ROOT / "apps/radishmind-web/src")
    for root in roots:
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"} or path.name.endswith("_test.go"):
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured:
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this boundary slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary_scope(fixture)
    assert_surfaces(fixture)
    assert_draft_routes(fixture)
    assert_risk_next_and_stop_lines(fixture)
    assert_forbidden_artifacts_and_side_effects(fixture)
    assert_docs_check_repo_and_no_leak(fixture)
    print("workflow function surface boundary v1 checks passed.")


if __name__ == "__main__":
    main()
