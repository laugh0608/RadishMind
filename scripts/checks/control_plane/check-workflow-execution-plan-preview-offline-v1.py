#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-execution-plan-preview-offline-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
EXPECTED_DEPENDENCIES = {
    "workflow-function-surface-boundary-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-boundary-v1.json",
        "function_surface_boundary_defined",
    ),
    "workflow-draft-designer-offline-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-draft-designer-offline-v1.json",
        "workflow_draft_designer_offline_defined",
    ),
    "workflow-draft-validation-inspector-offline-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-draft-validation-inspector-offline-v1.json",
        "workflow_draft_validation_inspector_offline_defined",
    ),
    "workflow-definition-detail-read-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-definition-detail-read-v1.json",
        "workflow_definition_detail_read_defined",
    ),
    "workflow-run-detail-read-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-run-detail-read-v1.json",
        "workflow_run_detail_read_defined",
    ),
    "workflow-blocked-action-preview-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-blocked-action-preview-v1.json",
        "workflow_blocked_action_preview_defined",
    ),
    "workflow-confirmation-placeholder-read-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-confirmation-placeholder-read-v1.json",
        "workflow_confirmation_placeholder_read_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "runtime_api_implemented",
    "production_api_consumer_ready",
    "workflow_builder_ready",
    "workflow_execution_plan_persistence_ready",
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
    "workflow_execution_plan_preview_view_model",
    "selected_draft_plan_projection",
    "stage_order_display",
    "node_to_stage_mapping_display",
    "provider_profile_requirement_display",
    "confirmation_audit_gate_display",
    "blocked_runtime_publish_writeback_replay_reason_display",
    "route_request_audit_metadata_display",
    "forbidden_output_guard_reused",
    "no_live_backend_request",
    "no_execution_publish_confirmation_controls",
}
EXPECTED_REQUIRED_FIELDS = {
    "selectedDraftId",
    "validationStatus",
    "summary",
    "stageOrder",
    "nodeStageMappings",
    "providerProfileRequirements",
    "confirmationAuditGates",
    "blockedPlanReasons",
    "auditMetadata",
    "routePath",
    "requestId",
    "auditRef",
    "canRenderExecutionPlanPreview",
    "canPreviewPlanLocally",
    "canRequestLiveBackend",
    "canPersistExecutionPlan",
    "canPublishWorkflow",
    "canStartRun",
    "canExecuteWorkflow",
    "canSubmitConfirmationDecision",
    "canWriteBusinessTruth",
    "canReplayRun",
}
EXPECTED_STATUSES = {"ready", "review_required", "blocked"}
EXPECTED_STAGE_IDS = {
    "stage_context_collection",
    "stage_model_reasoning",
    "stage_policy_gate",
    "stage_tool_preview",
    "stage_advisory_output",
    "stage_audit_projection",
}
EXPECTED_STAGE_KINDS = {"context", "model", "policy", "preview", "output", "audit"}
EXPECTED_PROVIDER_REQUIREMENTS = {
    "provider_profile_model_stage",
    "tool_adapter_preview_stage",
    "policy_confirmation_profile",
}
EXPECTED_GATES = {
    "gate_validation_before_plan",
    "gate_confirmation_before_tool",
    "gate_audit_projection",
    "gate_runtime_execution",
}
EXPECTED_BLOCKED_REASONS = {
    "blocked_runtime",
    "blocked_publish",
    "blocked_writeback",
    "blocked_replay",
}
EXPECTED_STOP_LINES = {
    "no new runtime API",
    "no live backend request",
    "no execution plan persistence",
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
    require(fixture.get("kind") == "workflow_execution_plan_preview_offline_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-execution-plan-preview-offline-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_execution_plan_preview_offline_defined",
        "workflow execution plan preview status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    page = fixture.get("page") or {}
    require(page.get("id") == "workflow-execution-plan-preview-offline", "page id drifted")
    require(page.get("source_page_id") == "workflow-draft-validation-inspector-offline", "source page drifted")
    require(page.get("source_route_id") == "workflow-definition-summary-list-route", "source route drifted")
    require(
        page.get("draft_route_id") == "workflow-execution-plan-preview-offline-draft",
        "draft route id drifted",
    )
    require(
        page.get("validation_route_id") == "workflow-draft-validation-inspector-offline-draft",
        "validation route id drifted",
    )
    require(page.get("status") == "offline_execution_plan_preview_surface", "page status drifted")
    require(
        page.get("default_selected_draft_id") == "draft_wf_radishflow_copilot_latest_offline",
        "default selected draft id drifted",
    )
    require(page.get("local_preview_allowed_now") is True, "local preview must remain true")
    for field in (
        "live_backend_allowed_now",
        "runtime_api_allowed_now",
        "execution_plan_persistence_allowed_now",
        "publish_allowed_now",
        "workflow_execution_allowed_now",
        "confirmation_submission_allowed_now",
        "business_writeback_allowed_now",
        "run_replay_allowed_now",
    ):
        require(page.get(field) is False, f"{field} must remain false")


def assert_capability_contract(fixture: dict[str, Any]) -> None:
    missing_capabilities = sorted(EXPECTED_CAPABILITIES - set(fixture.get("implemented_capabilities") or []))
    require(not missing_capabilities, f"missing implemented capabilities: {missing_capabilities}")
    missing_fields = sorted(EXPECTED_REQUIRED_FIELDS - set(fixture.get("required_view_model_fields") or []))
    require(not missing_fields, f"missing required view model fields: {missing_fields}")

    policy = fixture.get("execution_plan_policy") or {}
    require(EXPECTED_STATUSES.issubset(set(policy.get("required_statuses") or [])), "statuses drifted")
    require(EXPECTED_STAGE_IDS.issubset(set(policy.get("required_stage_ids") or [])), "stage ids drifted")
    require(EXPECTED_STAGE_KINDS.issubset(set(policy.get("required_stage_kinds") or [])), "stage kinds drifted")
    require(policy.get("minimum_stage_count") >= 6, "stage count drifted")
    require(policy.get("minimum_node_mapping_count") >= 6, "node mapping count drifted")
    require(
        EXPECTED_PROVIDER_REQUIREMENTS.issubset(set(policy.get("required_provider_requirements") or [])),
        "provider requirements drifted",
    )
    require(EXPECTED_GATES.issubset(set(policy.get("required_gates") or [])), "gates drifted")
    require(
        EXPECTED_BLOCKED_REASONS.issubset(set(policy.get("required_blocked_reasons") or [])),
        "blocked reasons drifted",
    )
    require(policy.get("local_preview_allowed_now") is True, "local preview must remain true")
    for field in (
        "persist_execution_plan_allowed_now",
        "publish_allowed_now",
        "workflow_execution_allowed_now",
        "confirmation_submission_allowed_now",
        "business_writeback_allowed_now",
        "run_replay_allowed_now",
    ):
        require(policy.get(field) is False, f"{field} must remain false")
    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    preview_source = read("apps/radishmind-web/src/features/control-plane-read/workflowExecutionPlanPreview.ts")
    app_source = read("apps/radishmind-web/src/app/App.tsx")
    styles = read("apps/radishmind-web/src/styles.css")
    checked_source = "\n".join([preview_source, app_source])

    for literal in (
        "buildWorkflowExecutionPlanPreviewViewModel",
        "WorkflowExecutionPlanPreviewViewModel",
        "WorkflowExecutionPlanStage",
        "WorkflowExecutionPlanNodeMapping",
        "WorkflowExecutionPlanProviderRequirement",
        "WorkflowExecutionPlanGate",
        "WorkflowExecutionPlanBlockedReason",
        "workflow-execution-plan-preview-offline",
        "workflow-execution-plan-preview-offline-draft",
        "stage_context_collection",
        "stage_model_reasoning",
        "stage_policy_gate",
        "stage_tool_preview",
        "stage_advisory_output",
        "stage_audit_projection",
        "provider_profile_model_stage",
        "tool_adapter_preview_stage",
        "policy_confirmation_profile",
        "gate_validation_before_plan",
        "gate_confirmation_before_tool",
        "gate_audit_projection",
        "gate_runtime_execution",
        "blocked_runtime",
        "blocked_publish",
        "blocked_writeback",
        "blocked_replay",
        "canRenderExecutionPlanPreview",
        "canPreviewPlanLocally: true",
        "canRequestLiveBackend: false",
        "canPersistExecutionPlan: false",
        "canPublishWorkflow: false",
        "canStartRun: false",
        "canExecuteWorkflow: false",
        "canSubmitConfirmationDecision: false",
        "canWriteBusinessTruth: false",
        "canReplayRun: false",
    ):
        require(literal in preview_source, f"execution plan preview source missing literal: {literal}")

    for literal in (
        "WorkflowExecutionPlanPreviewPanel",
        "WorkflowExecutionPlanSummaryCard",
        "WorkflowExecutionPlanStageCard",
        "WorkflowExecutionPlanNodeMappingCard",
        "WorkflowExecutionPlanProviderRequirementCard",
        "WorkflowExecutionPlanGateCard",
        "WorkflowExecutionPlanBlockedReasonCard",
        "workflowExecutionPlanPreview",
        "workflow-execution-plan-preview",
    ):
        require(literal in app_source, f"App.tsx missing execution plan preview literal: {literal}")

    for literal in (
        ".workflow-execution-plan-preview",
        ".workflow-execution-plan-summary-grid",
        ".workflow-execution-plan-stage-grid",
        ".workflow-execution-plan-node-grid",
        ".workflow-execution-plan-provider-grid",
        ".workflow-execution-plan-gate-grid",
        ".workflow-execution-plan-blocked-grid",
        ".workflow-execution-plan-card",
        ".workflow-execution-plan-stage",
        ".workflow-execution-plan-node",
        ".workflow-execution-plan-provider",
        ".workflow-execution-plan-gate",
        ".workflow-execution-plan-blocked-reason",
        ".workflow-execution-plan-row-main",
    ):
        require(literal in styles, f"styles missing execution plan preview literal: {literal}")

    for status in EXPECTED_STATUSES:
        require(f'"{status}"' in preview_source, f"missing execution plan status literal: {status}")
    for stage_id in EXPECTED_STAGE_IDS:
        require(f'"{stage_id}"' in preview_source, f"missing stage id: {stage_id}")
    for stage_kind in EXPECTED_STAGE_KINDS:
        require(f'"{stage_kind}"' in preview_source, f"missing stage kind: {stage_kind}")
    for requirement_id in EXPECTED_PROVIDER_REQUIREMENTS:
        require(f'"{requirement_id}"' in preview_source, f"missing provider requirement: {requirement_id}")
    for gate_id in EXPECTED_GATES:
        require(f'"{gate_id}"' in preview_source, f"missing gate: {gate_id}")
    for reason_id in EXPECTED_BLOCKED_REASONS:
        require(f'"{reason_id}"' in preview_source, f"missing blocked reason: {reason_id}")

    for forbidden in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden) not in checked_source, f"execution plan preview source contains forbidden literal: {forbidden}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-draft-validation-inspector-offline-v1.py"
    current_checker = "check-workflow-execution-plan-preview-offline-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow execution plan preview check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "execution plan preview check must run after draft validation inspector check",
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
    print("workflow execution plan preview offline v1 checks passed.")


if __name__ == "__main__":
    main()
