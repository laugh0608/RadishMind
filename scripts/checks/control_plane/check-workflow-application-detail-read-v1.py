#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-application-detail-read-v1.json"
FUNCTION_SURFACE_BOUNDARY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-boundary-v1.json"
)
WORKSPACE_APPLICATIONS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-workspace-applications-v1.json"
)
WORKFLOW_BLOCKED_ACTION_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-blocked-action-preview-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-function-surface-boundary-v1": (
        FUNCTION_SURFACE_BOUNDARY_PATH,
        "function_surface_boundary_defined",
    ),
    "control-plane-read-workspace-applications-v1": (
        WORKSPACE_APPLICATIONS_PATH,
        "workspace_applications_implemented",
    ),
    "workflow-blocked-action-preview-v1": (
        WORKFLOW_BLOCKED_ACTION_PATH,
        "workflow_blocked_action_preview_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "runtime_api_implemented",
    "production_api_consumer_ready",
    "application_lifecycle_mutation_ready",
    "workflow_builder_ready",
    "workflow_definition_mutation_ready",
    "workflow_executor_ready",
    "node_executor_ready",
    "tool_executor_ready",
    "confirmation_flow_ready",
    "confirmation_decision_ready",
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
    "workflow_application_detail_view_model",
    "offline_fixture_detail_surface",
    "application_identity_display",
    "tenant_owner_reference_display",
    "application_type_status_display",
    "provider_profile_display",
    "risk_summary_display",
    "latest_run_reference_display",
    "route_request_audit_metadata_display",
    "blocked_capability_preview_display",
    "forbidden_output_guard_reused",
    "no_live_detail_backend_request",
    "no_application_mutation_or_execution_controls",
}
EXPECTED_REQUIRED_FIELDS = {
    "applicationId",
    "tenantRef",
    "ownerSubjectRef",
    "applicationType",
    "displayName",
    "lifecycleStatus",
    "providerProfileRef",
    "latestWorkflowDefinitionRef",
    "latestRunRef",
    "lastRunStatus",
    "updatedAt",
    "riskSummary",
    "routeMetadata",
    "blockedCapabilities",
    "routePath",
    "requestId",
    "auditRef",
    "canRenderApplicationDetail",
    "canRequestLiveBackend",
    "canMutate",
    "canCreateApplication",
    "canEditApplication",
    "canDeleteApplication",
    "canPublishApplication",
    "canStartRun",
    "canExecuteWorkflow",
    "canSubmitConfirmationDecision",
    "canWriteBusinessTruth",
    "canReplayRun",
}
EXPECTED_RISK_FIELDS = {"label", "riskLevel", "requiresConfirmationCapable", "summary", "policyRef"}
EXPECTED_BLOCKED_FIELDS = {
    "capabilityId",
    "label",
    "status",
    "reason",
    "missingPrerequisite",
    "auditRef",
}
EXPECTED_BLOCKED_CAPABILITY_IDS = {
    "blocked_application_mutation",
    "blocked_workflow_execution",
    "blocked_confirmation_decision",
    "blocked_business_writeback",
    "blocked_run_replay",
}
EXPECTED_STOP_LINES = {
    "no new runtime API",
    "no live detail backend request",
    "no application lifecycle mutation",
    "no workflow executor",
    "no node executor",
    "no tool executor",
    "no confirmation decision",
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
    require(fixture.get("kind") == "workflow_application_detail_read_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-application-detail-read-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_application_detail_read_defined",
        "application detail status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    page = fixture.get("page") or {}
    require(page.get("id") == "workflow-application-detail-read", "page id drifted")
    require(page.get("source_page_id") == "workspace-applications", "source page drifted")
    require(page.get("source_route_id") == "application-summary-list-route", "source route drifted")
    require(page.get("draft_route_id") == "application-detail-read-draft", "draft route id drifted")
    require(page.get("status") == "offline_read_only_detail_surface", "page status drifted")
    require(page.get("default_application_id") == "app_flow_copilot", "default application id drifted")
    require(page.get("live_detail_route_allowed_now") is False, "live detail route must remain disabled")
    require(page.get("runtime_api_allowed_now") is False, "runtime API must remain disabled")
    require(page.get("application_mutation_allowed_now") is False, "application mutation must remain disabled")
    require(page.get("execution_allowed_now") is False, "execution must remain disabled")


def assert_capability_contract(fixture: dict[str, Any]) -> None:
    missing_capabilities = sorted(EXPECTED_CAPABILITIES - set(fixture.get("implemented_capabilities") or []))
    require(not missing_capabilities, f"missing implemented capabilities: {missing_capabilities}")
    missing_fields = sorted(EXPECTED_REQUIRED_FIELDS - set(fixture.get("required_view_model_fields") or []))
    require(not missing_fields, f"missing required view model fields: {missing_fields}")

    risk = fixture.get("risk_summary_policy") or {}
    missing_risk_fields = sorted(EXPECTED_RISK_FIELDS - set(risk.get("required_fields") or []))
    require(not missing_risk_fields, f"missing risk summary fields: {missing_risk_fields}")
    require({"low", "medium", "high"}.issubset(set(risk.get("allowed_risk_levels") or [])), "risk levels drifted")
    require(risk.get("confirmation_marker_visible") is True, "confirmation marker must remain visible")
    require(risk.get("confirmation_decision_allowed_now") is False, "confirmation decision must remain disabled")

    blocked = fixture.get("blocked_capability_policy") or {}
    missing_blocked_fields = sorted(EXPECTED_BLOCKED_FIELDS - set(blocked.get("required_fields") or []))
    require(not missing_blocked_fields, f"missing blocked capability fields: {missing_blocked_fields}")
    require(blocked.get("required_blocked_status") == "blocked", "blocked capability status drifted")
    require(blocked.get("minimum_blocked_capability_count") >= 4, "blocked capability count drifted")
    for field in (
        "application_mutation_allowed_now",
        "workflow_execution_allowed_now",
        "confirmation_decision_allowed_now",
        "business_writeback_allowed_now",
        "run_replay_allowed_now",
    ):
        require(blocked.get(field) is False, f"{field} must remain false")

    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    detail_source = read("apps/radishmind-web/src/features/control-plane-read/workflowApplicationDetail.ts")
    applications_source = read("apps/radishmind-web/src/features/control-plane-read/workspaceApplications.ts")
    app_source = read("apps/radishmind-web/src/app/App.tsx")
    styles = read("apps/radishmind-web/src/styles.css")
    checked_source = "\n".join([detail_source, app_source])

    for literal in (
        "buildWorkflowApplicationDetailViewModel",
        "WorkflowApplicationDetailViewModel",
        "WorkspaceApplicationRow",
        "workflow-application-detail-read",
        "application-detail-read-draft",
        "app_flow_copilot",
        "tenantRef",
        "ownerSubjectRef",
        "applicationType",
        "lifecycleStatus",
        "providerProfileRef",
        "latestRunRef",
        "riskSummary",
        "routeMetadata",
        "blockedCapabilities",
        "canRenderApplicationDetail",
        "canRequestLiveBackend: false",
        "canMutate: false",
        "canCreateApplication: false",
        "canEditApplication: false",
        "canDeleteApplication: false",
        "canPublishApplication: false",
        "canStartRun: false",
        "canExecuteWorkflow: false",
        "canSubmitConfirmationDecision: false",
        "canWriteBusinessTruth: false",
        "canReplayRun: false",
    ):
        require(literal in checked_source, f"application detail source missing literal: {literal}")

    for literal in (
        "buildApplicationsEnvelope",
        "WorkspaceApplicationRow",
        "app_flow_copilot",
        "app_docs_assistant",
    ):
        require(literal in applications_source, f"workspace applications source missing reusable literal: {literal}")

    for literal in (
        "WorkflowApplicationDetailPanel",
        "WorkflowApplicationRiskCard",
        "WorkflowApplicationRouteMetadataCard",
        "WorkflowApplicationBlockedCapabilityCard",
        "workflowApplicationDetail",
    ):
        require(literal in app_source, f"App.tsx missing application detail render literal: {literal}")

    for literal in (
        ".workflow-application-detail",
        ".workflow-application-summary-grid",
        ".workflow-application-risk-grid",
        ".workflow-application-blocked-grid",
        ".workflow-application-blocked-capability",
    ):
        require(literal in styles, f"styles missing application detail literal: {literal}")

    for capability_id in EXPECTED_BLOCKED_CAPABILITY_IDS:
        require(f'capabilityId: "{capability_id}"' in detail_source, f"missing capability fixture: {capability_id}")

    for forbidden in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden) not in checked_source, f"application detail source contains forbidden literal: {forbidden}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-blocked-action-preview-v1.py"
    current_checker = "check-workflow-application-detail-read-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow application detail check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow application detail check must run after blocked action preview check",
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
    print("workflow application detail read v1 checks passed.")


if __name__ == "__main__":
    main()
