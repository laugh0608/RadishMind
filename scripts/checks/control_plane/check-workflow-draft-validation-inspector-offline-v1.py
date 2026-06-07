#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-draft-validation-inspector-offline-v1.json"
FUNCTION_SURFACE_BOUNDARY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-boundary-v1.json"
)
DRAFT_DESIGNER_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-draft-designer-offline-v1.json"
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
    "workflow-draft-designer-offline-v1": (
        DRAFT_DESIGNER_PATH,
        "workflow_draft_designer_offline_defined",
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
    "workflow_draft_validation_inspector_view_model",
    "offline_fixture_validation_surface",
    "selected_draft_validation_projection",
    "structural_check_display",
    "contract_check_display",
    "blocked_capability_check_display",
    "route_request_audit_metadata_display",
    "forbidden_output_guard_reused",
    "no_live_backend_request",
    "no_persistence_publish_execution_controls",
}
EXPECTED_REQUIRED_FIELDS = {
    "inspectedDraftId",
    "validationStatus",
    "summary",
    "structuralChecks",
    "contractChecks",
    "blockedCapabilityChecks",
    "auditMetadata",
    "routePath",
    "requestId",
    "auditRef",
    "canRenderDraftValidationInspector",
    "canInspectDraftLocally",
    "canRequestLiveBackend",
    "canPersistDraft",
    "canPublishWorkflow",
    "canStartRun",
    "canExecuteWorkflow",
    "canSubmitConfirmationDecision",
    "canWriteBusinessTruth",
    "canReplayRun",
}
EXPECTED_STATUSES = {"passed", "needs_review", "blocked"}
EXPECTED_SEVERITIES = {"info", "warning", "blocking"}
EXPECTED_STRUCTURAL_CHECKS = {
    "entry_context_lane",
    "model_reasoning_lane",
    "policy_gate_path",
    "output_audit_path",
    "orphan_node_scan",
}
EXPECTED_CONTRACT_CHECKS = {"input_contract_fields", "output_contract_fields"}
EXPECTED_BLOCKED_CAPABILITY_CHECKS = {
    "blocked_draft_persistence",
    "blocked_publish",
    "blocked_runtime",
    "blocked_confirmation_decision",
}
EXPECTED_INPUT_FIELDS = {"tenant_ref", "application_ref", "selection_summary", "diagnostic_summary"}
EXPECTED_OUTPUT_FIELDS = {"answer_summary", "candidate_actions", "risk_summary", "audit_refs"}
EXPECTED_STOP_LINES = {
    "no new runtime API",
    "no live backend request",
    "no workflow builder mutation",
    "no workflow draft persistence",
    "no validation result persistence",
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
    require(
        fixture.get("kind") == "workflow_draft_validation_inspector_offline_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-draft-validation-inspector-offline-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_draft_validation_inspector_offline_defined",
        "workflow draft validation inspector status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    page = fixture.get("page") or {}
    require(page.get("id") == "workflow-draft-validation-inspector-offline", "page id drifted")
    require(page.get("source_page_id") == "workflow-draft-designer-offline", "source page drifted")
    require(page.get("source_route_id") == "workflow-definition-summary-list-route", "source route drifted")
    require(
        page.get("draft_route_id") == "workflow-draft-validation-inspector-offline-draft",
        "draft route id drifted",
    )
    require(page.get("status") == "offline_draft_validation_inspector_surface", "page status drifted")
    require(
        page.get("default_inspected_draft_id") == "draft_wf_radishflow_copilot_latest_offline",
        "default inspected draft id drifted",
    )
    require(page.get("local_inspection_allowed_now") is True, "local inspection must remain true")
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

    validation_policy = fixture.get("validation_policy") or {}
    require(EXPECTED_STATUSES.issubset(set(validation_policy.get("required_statuses") or [])), "statuses drifted")
    require(
        EXPECTED_SEVERITIES.issubset(set(validation_policy.get("required_severities") or [])),
        "severities drifted",
    )
    require(
        EXPECTED_STRUCTURAL_CHECKS.issubset(set(validation_policy.get("required_structural_checks") or [])),
        "structural checks drifted",
    )
    require(
        EXPECTED_CONTRACT_CHECKS.issubset(set(validation_policy.get("required_contract_checks") or [])),
        "contract checks drifted",
    )
    require(
        EXPECTED_BLOCKED_CAPABILITY_CHECKS.issubset(
            set(validation_policy.get("required_blocked_capability_checks") or [])
        ),
        "blocked capability checks drifted",
    )
    require(validation_policy.get("minimum_structural_check_count") >= 5, "structural check count drifted")
    require(validation_policy.get("minimum_contract_check_count") >= 2, "contract check count drifted")
    require(
        validation_policy.get("minimum_blocked_capability_check_count") >= 4,
        "blocked capability check count drifted",
    )
    require(validation_policy.get("local_inspection_allowed_now") is True, "local inspection must remain true")
    for field in (
        "persist_validation_result_allowed_now",
        "publish_allowed_now",
        "workflow_execution_allowed_now",
        "confirmation_submission_allowed_now",
    ):
        require(validation_policy.get(field) is False, f"{field} must remain false")

    contract_policy = fixture.get("contract_policy") or {}
    require(
        EXPECTED_INPUT_FIELDS.issubset(set(contract_policy.get("required_input_fields") or [])),
        "input fields drifted",
    )
    require(
        EXPECTED_OUTPUT_FIELDS.issubset(set(contract_policy.get("required_output_fields") or [])),
        "output fields drifted",
    )
    require(contract_policy.get("result_is_inspection_only") is True, "result must remain inspection-only")
    require(
        contract_policy.get("materialized_result_reader_allowed_now") is False,
        "materialized result reader must remain false",
    )
    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    inspector_source = read("apps/radishmind-web/src/features/control-plane-read/workflowDraftValidationInspector.ts")
    designer_source = read("apps/radishmind-web/src/features/control-plane-read/workflowDraftDesigner.ts")
    app_source = read("apps/radishmind-web/src/app/App.tsx")
    styles = read("apps/radishmind-web/src/styles.css")
    checked_source = "\n".join([inspector_source, designer_source, app_source])

    for literal in (
        "buildWorkflowDraftValidationInspectorViewModel",
        "WorkflowDraftValidationInspectorViewModel",
        "WorkflowDraftStructuralCheck",
        "WorkflowDraftContractCheck",
        "WorkflowDraftBlockedCapabilityCheck",
        "workflow-draft-validation-inspector-offline",
        "workflow-draft-validation-inspector-offline-draft",
        "entry_context_lane",
        "model_reasoning_lane",
        "policy_gate_path",
        "output_audit_path",
        "orphan_node_scan",
        "input_contract_fields",
        "output_contract_fields",
        "REQUIRED_INPUT_FIELDS",
        "REQUIRED_OUTPUT_FIELDS",
        "blockedCapabilityChecks",
        "canRenderDraftValidationInspector",
        "canInspectDraftLocally: true",
        "canRequestLiveBackend: false",
        "canPersistDraft: false",
        "canPublishWorkflow: false",
        "canStartRun: false",
        "canExecuteWorkflow: false",
        "canSubmitConfirmationDecision: false",
        "canWriteBusinessTruth: false",
        "canReplayRun: false",
    ):
        require(literal in inspector_source, f"draft validation inspector source missing literal: {literal}")

    for literal in (
        "WorkflowDraftValidationInspectorPanel",
        "WorkflowDraftValidationSummaryCard",
        "WorkflowDraftStructuralCheckCard",
        "WorkflowDraftContractCheckCard",
        "WorkflowDraftBlockedCapabilityCheckCard",
        "workflowDraftValidationInspector",
        "workflow-draft-validation-inspector",
    ):
        require(literal in app_source, f"App.tsx missing validation inspector render literal: {literal}")

    for literal in (
        "WorkflowDraftDesignerDraft",
        "blockedCapabilities",
        "workflow-draft-designer-offline",
    ):
        require(literal in designer_source, f"draft designer source missing reusable literal: {literal}")

    for literal in (
        ".workflow-draft-validation-inspector",
        ".workflow-draft-validation-summary-grid",
        ".workflow-draft-structural-check-grid",
        ".workflow-draft-contract-check-grid",
        ".workflow-draft-validation-blocked-grid",
        ".workflow-draft-validation-card",
        ".workflow-draft-structural-check",
        ".workflow-draft-contract-check",
        ".workflow-draft-validation-blocked-check",
        ".workflow-draft-validation-evidence",
    ):
        require(literal in styles, f"styles missing validation inspector literal: {literal}")

    for status in EXPECTED_STATUSES:
        require(f'"{status}"' in inspector_source, f"missing validation status literal: {status}")
    for severity in EXPECTED_SEVERITIES:
        require(f'"{severity}"' in inspector_source, f"missing severity literal: {severity}")
    for check_id in EXPECTED_STRUCTURAL_CHECKS | EXPECTED_CONTRACT_CHECKS:
        require(f'checkId: "{check_id}"' in inspector_source, f"missing check id: {check_id}")
    for field in EXPECTED_INPUT_FIELDS | EXPECTED_OUTPUT_FIELDS:
        require(f'"{field}"' in inspector_source, f"missing contract field: {field}")

    for forbidden in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden) not in checked_source, f"draft validation inspector source contains forbidden literal: {forbidden}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-draft-designer-offline-v1.py"
    current_checker = "check-workflow-draft-validation-inspector-offline-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow draft validation inspector check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow draft validation inspector check must run after draft designer check",
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
    print("workflow draft validation inspector offline v1 checks passed.")


if __name__ == "__main__":
    main()
