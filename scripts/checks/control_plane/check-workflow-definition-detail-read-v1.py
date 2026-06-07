#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-definition-detail-read-v1.json"
FUNCTION_SURFACE_BOUNDARY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-boundary-v1.json"
)
WORKFLOW_DEFINITIONS_PAGE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-workspace-workflow-definitions-v1.json"
)
FORMAL_UI_CLOSE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-readiness-close-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-function-surface-boundary-v1": (
        FUNCTION_SURFACE_BOUNDARY_PATH,
        "function_surface_boundary_defined",
    ),
    "control-plane-read-workspace-workflow-definitions-v1": (
        WORKFLOW_DEFINITIONS_PAGE_PATH,
        "workspace_workflow_definitions_implemented",
    ),
    "control-plane-read-formal-ui-readiness-close-v1": (
        FORMAL_UI_CLOSE_PATH,
        "formal_ui_readiness_closed",
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
EXPECTED_CAPABILITIES = {
    "workflow_definition_detail_view_model",
    "offline_fixture_detail_surface",
    "workflow_definition_node_list_display",
    "workflow_definition_edge_list_display",
    "input_output_summary_display",
    "blocked_action_preview_display",
    "route_request_audit_metadata_display",
    "forbidden_output_guard_reused",
    "no_live_detail_backend_request",
    "no_builder_or_execution_controls",
}
EXPECTED_REQUIRED_FIELDS = {
    "workflowDefinitionId",
    "applicationRef",
    "version",
    "definitionStatus",
    "nodes",
    "edges",
    "inputSummary",
    "outputSummary",
    "riskLevel",
    "requiresConfirmationCapable",
    "blockedActionPreview",
    "routePath",
    "requestId",
    "auditRef",
    "canRenderDefinitionDetail",
    "canRequestLiveBackend",
    "canMutate",
    "canEditWorkflow",
    "canRunWorkflow",
    "canConfirmWorkflowAction",
    "canWriteBusinessTruth",
    "canReplayRun",
}
EXPECTED_NODE_TYPES = {"prompt", "llm", "http_tool", "condition", "output"}
EXPECTED_BLOCKED_FIELDS = {
    "toolActionId",
    "toolRef",
    "actionKind",
    "riskLevel",
    "requiresConfirmation",
    "policyReason",
    "blockedState",
    "auditRef",
}
EXPECTED_STOP_LINES = {
    "no new runtime API",
    "no live detail backend request",
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


def assert_slice_and_page(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "workflow_definition_detail_read_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-definition-detail-read-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_definition_detail_read_defined",
        "workflow definition detail status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    page = fixture.get("page") or {}
    require(page.get("id") == "workflow-definition-detail-read", "page id drifted")
    require(page.get("source_page_id") == "workspace-workflow-definitions", "source page drifted")
    require(page.get("source_route_id") == "workflow-definition-summary-list-route", "source route drifted")
    require(page.get("draft_route_id") == "workflow-definition-detail-read-draft", "draft route id drifted")
    require(page.get("status") == "offline_read_only_detail_surface", "page status drifted")
    require(page.get("default_workflow_definition_id") == "wf_radishflow_copilot_latest", "default detail id drifted")
    require(page.get("live_detail_route_allowed_now") is False, "live detail route must remain disabled")
    require(page.get("runtime_api_allowed_now") is False, "runtime API must remain disabled")


def assert_capability_contract(fixture: dict[str, Any]) -> None:
    missing_capabilities = sorted(EXPECTED_CAPABILITIES - set(fixture.get("implemented_capabilities") or []))
    require(not missing_capabilities, f"missing implemented capabilities: {missing_capabilities}")
    missing_fields = sorted(EXPECTED_REQUIRED_FIELDS - set(fixture.get("required_view_model_fields") or []))
    require(not missing_fields, f"missing required view model fields: {missing_fields}")
    require(EXPECTED_NODE_TYPES.issubset(set(fixture.get("required_node_types") or [])), "node types drifted")

    blocked = fixture.get("blocked_action_policy") or {}
    missing_blocked_fields = sorted(EXPECTED_BLOCKED_FIELDS - set(blocked.get("required_fields") or []))
    require(not missing_blocked_fields, f"missing blocked action fields: {missing_blocked_fields}")
    require(blocked.get("default_blocked_state") == "blocked_executor_not_available", "blocked state drifted")
    require(blocked.get("requires_confirmation") is True, "blocked action must require confirmation marker")
    for field in (
        "execution_allowed_now",
        "confirmation_decision_allowed_now",
        "business_writeback_allowed_now",
    ):
        require(blocked.get(field) is False, f"{field} must remain false")

    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    detail_source = read("apps/radishmind-web/src/features/control-plane-read/workflowDefinitionDetail.ts")
    app_source = read("apps/radishmind-web/src/app/App.tsx")
    styles = read("apps/radishmind-web/src/styles.css")
    checked_source = "\n".join([detail_source, app_source])

    for literal in (
        "buildWorkflowDefinitionDetailViewModel",
        "WorkflowDefinitionDetailViewModel",
        "WorkflowDefinitionSummary",
        "workflow-definition-detail-read",
        "workflow-definition-detail-read-draft",
        "wf_radishflow_copilot_latest",
        "nodes",
        "edges",
        "inputSummary",
        "outputSummary",
        "blockedActionPreview",
        "blocked_executor_not_available",
        "canRequestLiveBackend: false",
        "canMutate: false",
        "canEditWorkflow: false",
        "canRunWorkflow: false",
        "canConfirmWorkflowAction: false",
        "canWriteBusinessTruth: false",
        "canReplayRun: false",
    ):
        require(literal in checked_source, f"detail source missing literal: {literal}")

    for literal in (
        "WorkflowDefinitionDetailPanel",
        "WorkflowDefinitionDetailNodeCard",
        "WorkflowDefinitionBlockedActionPreviewCard",
        "workflowDefinitionDetail",
    ):
        require(literal in app_source, f"App.tsx missing detail render literal: {literal}")

    for literal in (
        ".workflow-definition-detail",
        ".workflow-detail-node-list",
        ".workflow-detail-edge-list",
        ".workflow-detail-blocked-action",
    ):
        require(literal in styles, f"styles missing detail literal: {literal}")

    for node_type in EXPECTED_NODE_TYPES:
        require(f'nodeType: "{node_type}"' in detail_source, f"missing node type fixture: {node_type}")

    for forbidden in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden) not in checked_source, f"detail source contains forbidden literal: {forbidden}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-function-surface-boundary-v1.py"
    current_checker = "check-workflow-definition-detail-read-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow definition detail check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow definition detail check must run after function surface boundary check",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice_and_page(fixture)
    assert_capability_contract(fixture)
    assert_required_files(fixture)
    assert_source_boundaries(fixture)
    assert_docs_and_fast_baseline(fixture)
    print("workflow definition detail read v1 checks passed.")


if __name__ == "__main__":
    main()
