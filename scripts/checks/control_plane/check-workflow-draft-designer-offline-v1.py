#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-draft-designer-offline-v1.json"
FUNCTION_SURFACE_BOUNDARY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-boundary-v1.json"
)
WORKFLOW_DEFINITIONS_PAGE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-workspace-workflow-definitions-v1.json"
)
WORKFLOW_DEFINITION_DETAIL_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-definition-detail-read-v1.json"
WORKFLOW_CONFIRMATION_PLACEHOLDER_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-confirmation-placeholder-read-v1.json"
)
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
    "workflow-definition-detail-read-v1": (
        WORKFLOW_DEFINITION_DETAIL_PATH,
        "workflow_definition_detail_read_defined",
    ),
    "workflow-confirmation-placeholder-read-v1": (
        WORKFLOW_CONFIRMATION_PLACEHOLDER_PATH,
        "workflow_confirmation_placeholder_read_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "runtime_api_implemented",
    "production_api_consumer_ready",
    "workflow_builder_ready",
    "workflow_definition_mutation_ready",
    "workflow_draft_persistence_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "node_executor_ready",
    "tool_executor_ready",
    "confirmation_flow_ready",
    "confirmation_decision_ready",
    "confirmation_decision_store_ready",
    "execution_unlock_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "run_resume_ready",
    "durable_run_store_ready",
    "durable_result_store_ready",
    "materialized_result_reader_ready",
    "database_ready",
    "radish_oidc_ready",
    "production_ready",
}
EXPECTED_CAPABILITIES = {
    "workflow_draft_designer_view_model",
    "offline_fixture_draft_surface",
    "template_switch_local_state",
    "workflow_draft_node_list_display",
    "workflow_draft_edge_list_display",
    "readiness_summary_display",
    "risk_summary_display",
    "route_request_audit_metadata_display",
    "blocked_capability_preview_display",
    "forbidden_output_guard_reused",
    "no_live_backend_request",
    "no_persistence_publish_execution_controls",
}
EXPECTED_REQUIRED_FIELDS = {
    "defaultDraftId",
    "templates",
    "drafts",
    "routePath",
    "requestId",
    "auditRef",
    "canRenderDraftDesigner",
    "canInspectDraftLocally",
    "canSwitchDraftLocally",
    "canRequestLiveBackend",
    "canPersistDraft",
    "canPublishWorkflow",
    "canStartRun",
    "canExecuteWorkflow",
    "canSubmitConfirmationDecision",
    "canWriteBusinessTruth",
    "canReplayRun",
}
EXPECTED_TEMPLATE_FIELDS = {
    "draftId",
    "label",
    "applicationRef",
    "workflowDefinitionId",
    "workflowKind",
    "providerProfileRef",
    "summary",
    "riskLevel",
    "nodeCount",
    "status",
}
EXPECTED_TEMPLATE_STATUSES = {"ready_for_review", "needs_policy_review", "blocked_missing_runtime"}
EXPECTED_NODE_FIELDS = {
    "nodeId",
    "label",
    "nodeType",
    "lane",
    "readiness",
    "inputSummary",
    "outputSummary",
    "riskLevel",
    "requiresConfirmation",
    "previewOnlyReason",
}
EXPECTED_EDGE_FIELDS = {"edgeId", "fromNodeId", "toNodeId", "edgeKind", "conditionSummary"}
EXPECTED_READINESS_STATUSES = {"ready", "review_required", "blocked"}
EXPECTED_BLOCKED_CAPABILITIES = {
    "blocked_draft_persistence",
    "blocked_publish",
    "blocked_runtime",
    "blocked_confirmation_decision",
}
EXPECTED_STOP_LINES = {
    "no new runtime API",
    "no live backend request",
    "no workflow builder mutation",
    "no workflow draft persistence",
    "no workflow publish",
    "no workflow executor",
    "no node executor",
    "no tool executor",
    "no confirmation decision",
    "no confirmation decision store",
    "no execution unlock",
    "no business writeback",
    "no replay or resume",
    "no durable run store",
    "no durable result store",
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
    require(fixture.get("kind") == "workflow_draft_designer_offline_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-draft-designer-offline-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_draft_designer_offline_defined",
        "workflow draft designer status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    page = fixture.get("page") or {}
    require(page.get("id") == "workflow-draft-designer-offline", "page id drifted")
    require(page.get("source_page_id") == "workspace-workflow-definitions", "source page drifted")
    require(page.get("source_route_id") == "workflow-definition-summary-list-route", "source route drifted")
    require(
        page.get("draft_route_id") == "workflow-draft-designer-offline-draft",
        "draft route id drifted",
    )
    require(page.get("status") == "offline_draft_designer_surface", "page status drifted")
    require(
        page.get("default_draft_id") == "draft_wf_radishflow_copilot_latest_offline",
        "default draft id drifted",
    )
    require(page.get("local_template_switch_allowed_now") is True, "local template switch must stay allowed")
    for field in (
        "live_detail_route_allowed_now",
        "runtime_api_allowed_now",
        "draft_persistence_allowed_now",
        "publish_allowed_now",
        "workflow_execution_allowed_now",
        "confirmation_submission_allowed_now",
        "business_writeback_allowed_now",
    ):
        require(page.get(field) is False, f"{field} must remain false")


def assert_capability_contract(fixture: dict[str, Any]) -> None:
    missing_capabilities = sorted(EXPECTED_CAPABILITIES - set(fixture.get("implemented_capabilities") or []))
    require(not missing_capabilities, f"missing implemented capabilities: {missing_capabilities}")
    missing_fields = sorted(EXPECTED_REQUIRED_FIELDS - set(fixture.get("required_view_model_fields") or []))
    require(not missing_fields, f"missing required view model fields: {missing_fields}")

    template_policy = fixture.get("template_policy") or {}
    missing_template_fields = sorted(EXPECTED_TEMPLATE_FIELDS - set(template_policy.get("required_fields") or []))
    require(not missing_template_fields, f"missing template fields: {missing_template_fields}")
    require(
        EXPECTED_TEMPLATE_STATUSES.issubset(set(template_policy.get("required_statuses") or [])),
        "template statuses drifted",
    )
    require(template_policy.get("minimum_template_count") >= 2, "template count drifted")
    require(template_policy.get("local_template_switch_allowed_now") is True, "local switch must remain true")
    require(
        template_policy.get("persist_template_choice_allowed_now") is False,
        "template choice persistence must remain false",
    )

    draft_policy = fixture.get("draft_policy") or {}
    missing_node_fields = sorted(EXPECTED_NODE_FIELDS - set(draft_policy.get("required_node_fields") or []))
    require(not missing_node_fields, f"missing node fields: {missing_node_fields}")
    missing_edge_fields = sorted(EXPECTED_EDGE_FIELDS - set(draft_policy.get("required_edge_fields") or []))
    require(not missing_edge_fields, f"missing edge fields: {missing_edge_fields}")
    require(
        EXPECTED_READINESS_STATUSES.issubset(set(draft_policy.get("required_readiness_statuses") or [])),
        "readiness statuses drifted",
    )
    require(
        EXPECTED_BLOCKED_CAPABILITIES.issubset(set(draft_policy.get("required_blocked_capabilities") or [])),
        "blocked capabilities drifted",
    )
    require(draft_policy.get("minimum_node_count") >= 4, "node count drifted")
    require(draft_policy.get("minimum_edge_count") >= 4, "edge count drifted")
    for field in (
        "draft_persistence_allowed_now",
        "publish_allowed_now",
        "workflow_execution_allowed_now",
        "confirmation_submission_allowed_now",
    ):
        require(draft_policy.get(field) is False, f"{field} must remain false")

    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    designer_source = read("apps/radishmind-web/src/features/control-plane-read/workflowDraftDesigner.ts")
    definition_source = read("apps/radishmind-web/src/features/control-plane-read/workflowDefinitionDetail.ts")
    confirmation_source = read(
        "apps/radishmind-web/src/features/control-plane-read/workflowConfirmationPlaceholder.ts"
    )
    app_source = read("apps/radishmind-web/src/app/App.tsx")
    styles = read("apps/radishmind-web/src/styles.css")
    checked_source = "\n".join([designer_source, app_source])

    for literal in (
        "buildWorkflowDraftDesignerViewModel",
        "WorkflowDraftDesignerViewModel",
        "WorkflowDraftDesignerTemplate",
        "WorkflowDraftDesignerDraft",
        "WorkflowDraftDesignerNode",
        "WorkflowDraftDesignerEdge",
        "workflow-draft-designer-offline",
        "workflow-draft-designer-offline-draft",
        "draft_wf_radishflow_copilot_latest_offline",
        "buildWorkflowDefinitionDetailViewModel",
        "buildWorkflowConfirmationPlaceholderViewModel",
        "blocked_draft_persistence",
        "blocked_publish",
        "blocked_runtime",
        "blocked_confirmation_decision",
        "localOnlyInteraction: \"inspect_only\"",
        "canInspectDraftLocally: true",
        "canSwitchDraftLocally: true",
        "canRequestLiveBackend: false",
        "canPersistDraft: false",
        "canPublishWorkflow: false",
        "canStartRun: false",
        "canExecuteWorkflow: false",
        "canSubmitConfirmationDecision: false",
        "canWriteBusinessTruth: false",
        "canReplayRun: false",
    ):
        require(literal in designer_source, f"draft designer source missing literal: {literal}")

    for literal in (
        "WorkflowDraftDesignerPanel",
        "WorkflowDraftTemplateButton",
        "WorkflowDraftNodeCard",
        "WorkflowDraftEdgeCard",
        "WorkflowDraftReadinessCard",
        "WorkflowDraftRiskCard",
        "WorkflowDraftBlockedCapabilityCard",
        "workflowDraftDesigner",
        "selectedWorkflowDraftId",
        "setSelectedWorkflowDraftId",
        "workflow-draft-designer",
    ):
        require(literal in app_source, f"App.tsx missing draft designer render literal: {literal}")

    for literal in (
        "WorkflowDefinitionDetailViewModel",
        "blockedActionPreview",
        "workflow-definition-detail-read",
    ):
        require(literal in definition_source, f"definition source missing reusable literal: {literal}")

    for literal in (
        "WorkflowConfirmationPlaceholderViewModel",
        "humanReviewRequired",
        "disabledReason",
    ):
        require(literal in confirmation_source, f"confirmation source missing reusable literal: {literal}")

    for literal in (
        ".workflow-draft-designer",
        ".workflow-draft-template-grid",
        ".workflow-draft-summary-grid",
        ".workflow-draft-node-grid",
        ".workflow-draft-edge-grid",
        ".workflow-draft-readiness-grid",
        ".workflow-draft-risk-grid",
        ".workflow-draft-blocked-grid",
        ".workflow-draft-template-button",
        ".workflow-draft-card",
        ".workflow-draft-node",
        ".workflow-draft-edge",
        ".workflow-draft-readiness",
        ".workflow-draft-risk",
        ".workflow-draft-blocked-capability",
    ):
        require(literal in styles, f"styles missing draft designer literal: {literal}")

    for status in EXPECTED_TEMPLATE_STATUSES | EXPECTED_READINESS_STATUSES:
        require(f'"{status}"' in designer_source, f"draft designer source missing status literal: {status}")

    for capability_id in EXPECTED_BLOCKED_CAPABILITIES:
        require(f'capabilityId: "{capability_id}"' in designer_source, f"missing capability: {capability_id}")

    for forbidden in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden) not in checked_source, f"draft designer source contains forbidden literal: {forbidden}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-confirmation-placeholder-read-v1.py"
    current_checker = "check-workflow-draft-designer-offline-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow draft designer check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow draft designer check must run after confirmation placeholder check",
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
    print("workflow draft designer offline v1 checks passed.")


if __name__ == "__main__":
    main()
