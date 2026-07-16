#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-confirmation-placeholder-read-v1.json"
FUNCTION_SURFACE_BOUNDARY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-boundary-v1.json"
)
WORKFLOW_BLOCKED_ACTION_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-blocked-action-preview-v1.json"
WORKFLOW_APPLICATION_DETAIL_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-application-detail-read-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-function-surface-boundary-v1": (
        FUNCTION_SURFACE_BOUNDARY_PATH,
        "function_surface_boundary_defined",
    ),
    "workflow-blocked-action-preview-v1": (
        WORKFLOW_BLOCKED_ACTION_PATH,
        "workflow_blocked_action_preview_defined",
    ),
    "workflow-application-detail-read-v1": (
        WORKFLOW_APPLICATION_DETAIL_PATH,
        "workflow_application_detail_read_defined",
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
    "legacy_run_bound_confirmation_flow_ready",
    "legacy_run_bound_confirmation_decision_ready",
    "legacy_run_bound_confirmation_decision_store_ready",
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
    "workflow_confirmation_placeholder_view_model",
    "offline_fixture_confirmation_surface",
    "required_action_reference_display",
    "risk_summary_display",
    "required_decision_shape_display",
    "human_review_requirement_display",
    "disabled_reason_display",
    "route_request_audit_metadata_display",
    "missing_prerequisites_display",
    "forbidden_output_guard_reused",
    "no_live_backend_request",
    "archived_legacy_no_confirmation_or_execution_controls",
}
EXPECTED_REQUIRED_FIELDS = {
    "confirmationPlaceholderId",
    "requiredActionRef",
    "requiredRunRef",
    "workflowDefinitionId",
    "nodeExecutionRef",
    "toolRef",
    "actionKind",
    "riskLevel",
    "riskSummary",
    "policyReason",
    "requiredDecisionShape",
    "decisionFields",
    "humanReviewRequired",
    "disabledReason",
    "prerequisites",
    "auditMetadata",
    "legacyContractStatus",
    "historicalReadAllowed",
    "newSubmissionAllowed",
    "runtimeAuthoritative",
    "supersededBy",
    "routePath",
    "requestId",
    "auditRef",
    "canRenderConfirmationPlaceholder",
    "canRequestLiveBackend",
    "canMutate",
    "canSubmitConfirmationDecision",
    "canApproveDecision",
    "canRejectDecision",
    "canDeferDecision",
    "canPersistDecision",
    "canUnlockExecution",
    "canExecuteTool",
    "canWriteBusinessTruth",
    "canReplayRun",
}
EXPECTED_DECISION_FIELDS = {"decision", "actor_subject_ref", "reason", "audit_ref"}
EXPECTED_PREREQUISITE_FIELDS = {"prerequisiteId", "label", "status", "summary", "auditRef"}
EXPECTED_PREREQUISITE_STATUSES = {"missing", "defined_not_connected", "blocked"}
EXPECTED_STOP_LINES = {
    "no new runtime API from archived placeholder",
    "no live backend request",
    "no workflow executor",
    "no node executor",
    "no tool executor",
    "no legacy run-bound confirmation decision submission",
    "no legacy run-bound confirmation decision store",
    "no execution unlock from archived placeholder",
    "no business writeback",
    "no replay or resume",
    "no durable run store",
    "no durable result store",
    "no materialized result reader",
    "no production auth implementation from archived placeholder",
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
    require(fixture.get("kind") == "workflow_confirmation_placeholder_read_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-confirmation-placeholder-read-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_confirmation_placeholder_read_defined",
        "confirmation placeholder status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    page = fixture.get("page") or {}
    require(page.get("id") == "workflow-confirmation-placeholder-read", "page id drifted")
    require(page.get("source_page_id") == "workflow-blocked-action-preview-read", "source page drifted")
    require(page.get("source_route_id") == "run-record-summary-list-route", "source route drifted")
    require(page.get("draft_route_id") == "confirmation-placeholder-read-draft", "draft route id drifted")
    require(
        page.get("status") == "archived_legacy_read_only_confirmation_placeholder_surface",
        "page status drifted",
    )
    require(
        page.get("default_confirmation_placeholder_id") == "confirmation_placeholder_tool_action_preview",
        "default confirmation placeholder id drifted",
    )
    require(
        page.get("default_required_action_ref") == "tool_action_preview_reconnect_stream",
        "default required action ref drifted",
    )
    require(page.get("legacy_contract_status") == "superseded_archived", "legacy page must be archived")
    require(page.get("historical_read_allowed") is True, "legacy page history must stay readable")
    require(page.get("new_submission_allowed") is False, "legacy page must reject new submissions")
    require(page.get("runtime_authoritative") is False, "legacy page must not be authoritative")
    require(
        set(page.get("superseded_by") or [])
        == {"workflow_http_tool_action_plan.v1", "workflow_http_tool_confirmation_decision.v1"},
        "legacy page supersession targets drifted",
    )
    for field in (
        "live_detail_route_allowed_now",
        "runtime_api_allowed_now",
        "legacy_confirmation_submission_allowed_now",
        "legacy_decision_persistence_allowed_now",
        "execution_unlock_allowed_now",
    ):
        require(page.get(field) is False, f"{field} must remain false")


def assert_capability_contract(fixture: dict[str, Any]) -> None:
    missing_capabilities = sorted(EXPECTED_CAPABILITIES - set(fixture.get("implemented_capabilities") or []))
    require(not missing_capabilities, f"missing implemented capabilities: {missing_capabilities}")
    missing_fields = sorted(EXPECTED_REQUIRED_FIELDS - set(fixture.get("required_view_model_fields") or []))
    require(not missing_fields, f"missing required view model fields: {missing_fields}")

    decision_policy = fixture.get("decision_shape_policy") or {}
    missing_decision_fields = sorted(EXPECTED_DECISION_FIELDS - set(decision_policy.get("required_fields") or []))
    require(not missing_decision_fields, f"missing decision shape fields: {missing_decision_fields}")
    require(decision_policy.get("human_review_required") is True, "human review marker must remain true")
    for field in (
        "submit_legacy_run_bound_confirmation_allowed_now",
        "approve_legacy_run_bound_confirmation_allowed_now",
        "reject_legacy_run_bound_confirmation_allowed_now",
        "defer_legacy_run_bound_confirmation_allowed_now",
        "persist_legacy_run_bound_decision_allowed_now",
        "unlock_execution_from_legacy_placeholder_allowed_now",
    ):
        require(decision_policy.get(field) is False, f"{field} must remain false")

    prerequisite_policy = fixture.get("prerequisite_policy") or {}
    missing_prerequisite_fields = sorted(
        EXPECTED_PREREQUISITE_FIELDS - set(prerequisite_policy.get("required_fields") or [])
    )
    require(not missing_prerequisite_fields, f"missing prerequisite fields: {missing_prerequisite_fields}")
    require(
        EXPECTED_PREREQUISITE_STATUSES.issubset(set(prerequisite_policy.get("required_statuses") or [])),
        "prerequisite statuses drifted",
    )
    require(prerequisite_policy.get("minimum_prerequisite_count") >= 4, "prerequisite count drifted")
    for field in (
        "workflow_execution_allowed_now",
        "tool_execution_allowed_now",
        "business_writeback_allowed_now",
        "run_replay_allowed_now",
    ):
        require(prerequisite_policy.get(field) is False, f"{field} must remain false")

    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    confirmation_source = read("apps/radishmind-web/src/features/control-plane-read/workflowConfirmationPlaceholder.ts")
    blocked_source = read("apps/radishmind-web/src/features/control-plane-read/workflowBlockedActionPreview.ts")
    app_source = read("apps/radishmind-web/src/app/App.tsx")
    styles = read("apps/radishmind-web/src/styles.css")
    checked_source = "\n".join([confirmation_source, blocked_source, app_source])

    for literal in (
        "buildWorkflowConfirmationPlaceholderViewModel",
        "WorkflowConfirmationPlaceholderViewModel",
        "WorkflowBlockedActionPreviewViewModel",
        "workflow-confirmation-placeholder-read",
        "workflow-blocked-action-preview-read",
        "confirmation-placeholder-read-draft",
        "confirmation_placeholder_tool_action_preview",
        "tool_action_preview_reconnect_stream",
        "requiredActionRef",
        "requiredRunRef",
        "riskSummary",
        "requiredDecisionShape",
        "decisionFields",
        "humanReviewRequired: true",
        "disabledReason",
        "prerequisites",
        "auditMetadata",
        'legacyContractStatus: "superseded_archived"',
        "historicalReadAllowed: true",
        "newSubmissionAllowed: false",
        "runtimeAuthoritative: false",
        "supersededBy",
        "canRenderConfirmationPlaceholder",
        "canRequestLiveBackend: false",
        "canMutate: false",
        "canSubmitConfirmationDecision: false",
        "canApproveDecision: false",
        "canRejectDecision: false",
        "canDeferDecision: false",
        "canPersistDecision: false",
        "canUnlockExecution: false",
        "canExecuteTool: false",
        "canWriteBusinessTruth: false",
        "canReplayRun: false",
    ):
        require(literal in checked_source, f"confirmation placeholder source missing literal: {literal}")

    for literal in (
        "WorkflowConfirmationPlaceholderPanel",
        "WorkflowConfirmationDecisionFieldCard",
        "WorkflowConfirmationPrerequisiteCard",
        "workflowConfirmationPlaceholder",
        "workflow-confirmation-placeholder",
        "Archived Legacy Confirmation",
        "legacyContractStatus",
        "supersededBy",
        "historical required field",
    ):
        require(literal in app_source, f"App.tsx missing confirmation render literal: {literal}")

    for literal in (
        "buildWorkflowBlockedActionPreviewViewModel",
        "WorkflowConfirmationPlaceholderPreview",
        "confirmationPlaceholder",
        "legacy placeholder is permanently read-only",
    ):
        require(literal in blocked_source, f"blocked action source missing reusable literal: {literal}")

    for literal in (
        ".workflow-confirmation-placeholder-read",
        ".workflow-confirmation-summary-grid",
        ".workflow-confirmation-field-grid",
        ".workflow-confirmation-prerequisite-grid",
        ".workflow-confirmation-card",
        ".workflow-confirmation-field",
        ".workflow-confirmation-prerequisite",
    ):
        require(literal in styles, f"styles missing confirmation literal: {literal}")

    for status in EXPECTED_PREREQUISITE_STATUSES:
        require(f'"{status}"' in confirmation_source, f"missing prerequisite status fixture: {status}")

    for decision_field in EXPECTED_DECISION_FIELDS:
        require(f'fieldId: "{decision_field}"' in confirmation_source, f"missing decision field: {decision_field}")

    for forbidden in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden) not in checked_source, f"confirmation placeholder source contains forbidden literal: {forbidden}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-application-detail-read-v1.py"
    current_checker = "check-workflow-confirmation-placeholder-read-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow confirmation placeholder check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow confirmation placeholder check must run after application detail check",
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
    print("workflow confirmation placeholder read v1 checks passed.")


if __name__ == "__main__":
    main()
