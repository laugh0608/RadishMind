#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-blocked-action-preview-v1.json"
FUNCTION_SURFACE_BOUNDARY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-boundary-v1.json"
)
WORKFLOW_DEFINITION_DETAIL_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-definition-detail-read-v1.json"
)
WORKFLOW_RUN_DETAIL_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-run-detail-read-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-function-surface-boundary-v1": (
        FUNCTION_SURFACE_BOUNDARY_PATH,
        "function_surface_boundary_defined",
    ),
    "workflow-definition-detail-read-v1": (
        WORKFLOW_DEFINITION_DETAIL_PATH,
        "workflow_definition_detail_read_defined",
    ),
    "workflow-run-detail-read-v1": (
        WORKFLOW_RUN_DETAIL_PATH,
        "workflow_run_detail_read_defined",
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
    "workflow_blocked_action_preview_view_model",
    "offline_fixture_blocked_action_surface",
    "policy_reason_display",
    "missing_prerequisites_display",
    "confirmation_placeholder_display",
    "audit_trail_display",
    "related_run_guard_display",
    "forbidden_output_guard_reused",
    "no_live_backend_request",
    "no_execution_or_confirmation_controls",
}
EXPECTED_REQUIRED_FIELDS = {
    "runId",
    "workflowDefinitionId",
    "nodeExecutionRef",
    "toolActionId",
    "toolRef",
    "actionKind",
    "riskLevel",
    "requiresConfirmation",
    "policyReason",
    "blockedState",
    "missingPrerequisites",
    "auditTrail",
    "confirmationPlaceholder",
    "relatedRunGuard",
    "routePath",
    "requestId",
    "auditRef",
    "canRenderBlockedActionPreview",
    "canRequestLiveBackend",
    "canMutate",
    "canExecuteTool",
    "canSubmitConfirmationDecision",
    "canUnlockExecution",
    "canWriteBusinessTruth",
    "canReplayRun",
}
EXPECTED_REQUIREMENT_STATUSES = {"missing", "defined_not_connected", "blocked"}
EXPECTED_CONFIRMATION_FIELDS = {
    "confirmationPlaceholderId",
    "requiredActionRef",
    "riskSummary",
    "requiredDecisionShape",
    "humanReviewRequired",
    "disabledReason",
    "auditRef",
}
EXPECTED_STOP_LINES = {
    "no new runtime API",
    "no live backend request",
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
    require(fixture.get("kind") == "workflow_blocked_action_preview_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-blocked-action-preview-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_blocked_action_preview_defined",
        "blocked action preview status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    page = fixture.get("page") or {}
    require(page.get("id") == "workflow-blocked-action-preview-read", "page id drifted")
    require(page.get("source_page_id") == "workflow-run-detail-read", "source page drifted")
    require(page.get("source_route_id") == "run-record-summary-list-route", "source route drifted")
    require(page.get("draft_route_id") == "tool-action-preview-read-draft", "draft route id drifted")
    require(
        page.get("confirmation_draft_route_id") == "confirmation-placeholder-read-draft",
        "confirmation draft route id drifted",
    )
    require(page.get("status") == "offline_read_only_blocked_action_surface", "page status drifted")
    require(page.get("default_tool_action_id") == "tool_action_preview_reconnect_stream", "tool action id drifted")
    require(page.get("default_blocked_state") == "blocked_executor_not_available", "blocked state drifted")
    require(page.get("live_detail_route_allowed_now") is False, "live backend route must remain disabled")
    require(page.get("runtime_api_allowed_now") is False, "runtime API must remain disabled")
    require(page.get("tool_execution_allowed_now") is False, "tool execution must remain disabled")


def assert_capability_contract(fixture: dict[str, Any]) -> None:
    missing_capabilities = sorted(EXPECTED_CAPABILITIES - set(fixture.get("implemented_capabilities") or []))
    require(not missing_capabilities, f"missing implemented capabilities: {missing_capabilities}")
    missing_fields = sorted(EXPECTED_REQUIRED_FIELDS - set(fixture.get("required_view_model_fields") or []))
    require(not missing_fields, f"missing required view model fields: {missing_fields}")
    require(
        EXPECTED_REQUIREMENT_STATUSES.issubset(set(fixture.get("required_missing_prerequisite_statuses") or [])),
        "missing prerequisite statuses drifted",
    )

    confirmation = fixture.get("confirmation_placeholder_policy") or {}
    missing_confirmation_fields = sorted(EXPECTED_CONFIRMATION_FIELDS - set(confirmation.get("required_fields") or []))
    require(not missing_confirmation_fields, f"missing confirmation placeholder fields: {missing_confirmation_fields}")
    require(confirmation.get("human_review_required") is True, "human review marker must remain true")
    for field in (
        "submit_confirmation_allowed_now",
        "unlock_execution_allowed_now",
        "business_writeback_allowed_now",
    ):
        require(confirmation.get(field) is False, f"{field} must remain false")

    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    blocked_source = read("apps/radishmind-web/src/features/control-plane-read/workflowBlockedActionPreview.ts")
    definition_detail_source = read("apps/radishmind-web/src/features/control-plane-read/workflowDefinitionDetail.ts")
    run_detail_source = read("apps/radishmind-web/src/features/control-plane-read/workflowRunDetail.ts")
    app_source = read("apps/radishmind-web/src/app/App.tsx")
    styles = read("apps/radishmind-web/src/styles.css")
    checked_source = "\n".join([blocked_source, app_source])

    for literal in (
        "buildWorkflowBlockedActionPreviewViewModel",
        "WorkflowBlockedActionPreviewViewModel",
        "WorkflowDefinitionBlockedActionPreview",
        "WorkflowRunDetailGuardPreview",
        "workflow-blocked-action-preview-read",
        "tool-action-preview-read-draft",
        "confirmation-placeholder-read-draft",
        "tool_action_preview_reconnect_stream",
        "blocked_executor_not_available",
        "missingPrerequisites",
        "confirmationPlaceholder",
        "auditTrail",
        "relatedRunGuard",
        "canRenderBlockedActionPreview",
        "canRequestLiveBackend: false",
        "canMutate: false",
        "canExecuteTool: false",
        "canSubmitConfirmationDecision: false",
        "canUnlockExecution: false",
        "canWriteBusinessTruth: false",
        "canReplayRun: false",
    ):
        require(literal in checked_source, f"blocked action source missing literal: {literal}")

    for literal in (
        "WorkflowBlockedActionPreviewPanel",
        "WorkflowBlockedActionRequirementCard",
        "WorkflowConfirmationPlaceholderCard",
        "WorkflowBlockedActionAuditStepCard",
        "workflowBlockedActionPreview",
    ):
        require(literal in app_source, f"App.tsx missing blocked action render literal: {literal}")

    for literal in (
        "blockedActionPreview",
        "tool_action_preview_reconnect_stream",
        "blocked_executor_not_available",
    ):
        require(literal in definition_detail_source, f"definition detail source missing reusable literal: {literal}")

    for literal in (
        "blockedReplayPreview",
        "guard_replay_resume",
        "workflow executor implementation gate",
    ):
        require(literal in run_detail_source, f"run detail source missing reusable literal: {literal}")

    for literal in (
        ".workflow-blocked-action-preview",
        ".workflow-blocked-action-summary-grid",
        ".workflow-blocked-requirement-grid",
        ".workflow-confirmation-placeholder",
        ".workflow-blocked-audit-grid",
    ):
        require(literal in styles, f"styles missing blocked action literal: {literal}")

    for status in EXPECTED_REQUIREMENT_STATUSES:
        require(f'"{status}"' in blocked_source, f"missing prerequisite status fixture: {status}")

    for confirmation_literal in (
        "requiredDecisionShape",
        "humanReviewRequired: true",
        "decision: approve | reject",
        "former run-bound confirmation placeholder is archived",
        "legacy placeholder is permanently read-only",
    ):
        require(confirmation_literal in blocked_source, f"confirmation placeholder drifted: {confirmation_literal}")

    for forbidden in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden) not in checked_source, f"blocked action source contains forbidden literal: {forbidden}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-run-detail-read-v1.py"
    current_checker = "check-workflow-blocked-action-preview-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow blocked action preview check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow blocked action preview check must run after run detail check",
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
    print("workflow blocked action preview v1 checks passed.")


if __name__ == "__main__":
    main()
