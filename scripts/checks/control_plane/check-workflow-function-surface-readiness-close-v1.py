#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-readiness-close-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-function-surface-boundary-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-function-surface-boundary-v1.json",
        "function_surface_boundary_defined",
    ),
    "workflow-application-detail-read-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-application-detail-read-v1.json",
        "workflow_application_detail_read_defined",
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
    "workflow-draft-designer-offline-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-draft-designer-offline-v1.json",
        "workflow_draft_designer_offline_defined",
    ),
    "workflow-draft-validation-inspector-offline-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-draft-validation-inspector-offline-v1.json",
        "workflow_draft_validation_inspector_offline_defined",
    ),
    "workflow-execution-plan-preview-offline-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-execution-plan-preview-offline-v1.json",
        "workflow_execution_plan_preview_offline_defined",
    ),
    "workflow-runtime-readiness-inspector-offline-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-runtime-readiness-inspector-offline-v1.json",
        "workflow_runtime_readiness_inspector_offline_defined",
    ),
}

EXPECTED_FORBIDDEN_CLAIMS = {
    "runtime_api_implemented",
    "production_api_consumer_ready",
    "workflow_builder_ready",
    "workflow_executor_ready",
    "node_executor_ready",
    "tool_executor_ready",
    "agent_loop_ready",
    "workflow_draft_persistence_ready",
    "validation_result_persistence_ready",
    "execution_plan_persistence_ready",
    "runtime_readiness_persistence_ready",
    "workflow_publish_ready",
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
    "repository_adapter_ready",
    "store_selector_ready",
    "radish_oidc_ready",
    "production_ready",
}

EXPECTED_CLOSE_FALSE_FIELDS = {
    "live_backend_allowed_now",
    "runtime_api_allowed_now",
    "workflow_execution_allowed_now",
    "workflow_builder_mutation_allowed_now",
    "workflow_publish_allowed_now",
    "draft_persistence_allowed_now",
    "validation_result_persistence_allowed_now",
    "execution_plan_persistence_allowed_now",
    "runtime_readiness_persistence_allowed_now",
    "confirmation_submission_allowed_now",
    "confirmation_decision_store_allowed_now",
    "execution_unlock_allowed_now",
    "business_writeback_allowed_now",
    "run_replay_allowed_now",
    "run_resume_allowed_now",
    "durable_store_allowed_now",
    "production_auth_allowed_now",
    "database_attach_allowed_now",
    "repository_adapter_allowed_now",
}

EXPECTED_SURFACES = {
    "workflow-application-detail-read": {
        "surface": "Workflow application detail",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workflowApplicationDetail.ts",
        "builder": "buildWorkflowApplicationDetailViewModel",
        "render_component": "WorkflowApplicationDetailPanel",
        "render_anchor": "workflow-application-detail",
        "render_class": "workflow-application-detail",
        "source_mode": "fixture_derived_detail",
        "required_true_flags": set(),
        "required_false_flags": {
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
        },
    },
    "workflow-definition-detail-read": {
        "surface": "Workflow definition detail",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workflowDefinitionDetail.ts",
        "builder": "buildWorkflowDefinitionDetailViewModel",
        "render_component": "WorkflowDefinitionDetailPanel",
        "render_anchor": "workspace-workflow-definitions",
        "render_class": "workflow-definition-detail",
        "source_mode": "fixture_derived_detail",
        "required_true_flags": set(),
        "required_false_flags": {
            "canRequestLiveBackend",
            "canMutate",
            "canEditWorkflow",
            "canRunWorkflow",
            "canConfirmWorkflowAction",
            "canWriteBusinessTruth",
            "canReplayRun",
        },
    },
    "workflow-run-detail-read": {
        "surface": "Workflow run detail",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workflowRunDetail.ts",
        "builder": "buildWorkflowRunDetailViewModel",
        "render_component": "WorkflowRunDetailPanel",
        "render_anchor": "workspace-run-history",
        "render_class": "workflow-run-detail",
        "source_mode": "fixture_derived_detail",
        "required_true_flags": set(),
        "required_false_flags": {
            "canRequestLiveBackend",
            "canMutate",
            "canStartRun",
            "canCancelRun",
            "canResumeRun",
            "canReplayRun",
            "canMaterializeResult",
            "canWriteBusinessTruth",
        },
    },
    "workflow-blocked-action-preview-read": {
        "surface": "Workflow blocked action preview",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workflowBlockedActionPreview.ts",
        "builder": "buildWorkflowBlockedActionPreviewViewModel",
        "render_component": "WorkflowBlockedActionPreviewPanel",
        "render_anchor": "workflow-blocked-action-preview",
        "render_class": "workflow-blocked-action-preview",
        "source_mode": "fixture_derived_blocked_action",
        "required_true_flags": set(),
        "required_false_flags": {
            "canRequestLiveBackend",
            "canMutate",
            "canExecuteTool",
            "canSubmitConfirmationDecision",
            "canUnlockExecution",
            "canWriteBusinessTruth",
            "canReplayRun",
        },
    },
    "workflow-confirmation-placeholder-read": {
        "surface": "Workflow confirmation placeholder",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workflowConfirmationPlaceholder.ts",
        "builder": "buildWorkflowConfirmationPlaceholderViewModel",
        "render_component": "WorkflowConfirmationPlaceholderPanel",
        "render_anchor": "workflow-confirmation-placeholder",
        "render_class": "workflow-confirmation-placeholder-read",
        "source_mode": "fixture_derived_confirmation_shape",
        "required_true_flags": set(),
        "required_false_flags": {
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
        },
    },
    "workflow-draft-designer-offline": {
        "surface": "Workflow draft designer",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workflowDraftDesigner.ts",
        "builder": "buildWorkflowDraftDesignerViewModel",
        "render_component": "WorkflowDraftDesignerPanel",
        "render_anchor": "workflow-draft-designer",
        "render_class": "workflow-draft-designer",
        "source_mode": "offline_fixture_derived_draft",
        "required_true_flags": {"canInspectDraftLocally", "canSwitchDraftLocally"},
        "required_false_flags": {
            "canRequestLiveBackend",
            "canPersistDraft",
            "canPublishWorkflow",
            "canStartRun",
            "canExecuteWorkflow",
            "canSubmitConfirmationDecision",
            "canWriteBusinessTruth",
            "canReplayRun",
        },
    },
    "workflow-draft-validation-inspector-offline": {
        "surface": "Workflow draft validation inspector",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workflowDraftValidationInspector.ts",
        "builder": "buildWorkflowDraftValidationInspectorViewModel",
        "render_component": "WorkflowDraftValidationInspectorPanel",
        "render_anchor": "workflow-draft-validation-inspector",
        "render_class": "workflow-draft-validation-inspector",
        "source_mode": "offline_fixture_derived_validation",
        "required_true_flags": {"canInspectDraftLocally"},
        "required_false_flags": {
            "canRequestLiveBackend",
            "canPersistDraft",
            "canPublishWorkflow",
            "canStartRun",
            "canExecuteWorkflow",
            "canSubmitConfirmationDecision",
            "canWriteBusinessTruth",
            "canReplayRun",
        },
    },
    "workflow-execution-plan-preview-offline": {
        "surface": "Workflow execution plan preview",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workflowExecutionPlanPreview.ts",
        "builder": "buildWorkflowExecutionPlanPreviewViewModel",
        "render_component": "WorkflowExecutionPlanPreviewPanel",
        "render_anchor": "workflow-execution-plan-preview",
        "render_class": "workflow-execution-plan-preview",
        "source_mode": "offline_fixture_derived_execution_plan",
        "required_true_flags": {"canPreviewPlanLocally"},
        "required_false_flags": {
            "canRequestLiveBackend",
            "canPersistExecutionPlan",
            "canPublishWorkflow",
            "canStartRun",
            "canExecuteWorkflow",
            "canSubmitConfirmationDecision",
            "canWriteBusinessTruth",
            "canReplayRun",
        },
    },
    "workflow-runtime-readiness-inspector-offline": {
        "surface": "Workflow runtime readiness inspector",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workflowRuntimeReadinessInspector.ts",
        "builder": "buildWorkflowRuntimeReadinessInspectorViewModel",
        "render_component": "WorkflowRuntimeReadinessInspectorPanel",
        "render_anchor": "workflow-runtime-readiness-inspector",
        "render_class": "workflow-runtime-readiness-inspector",
        "source_mode": "offline_fixture_derived_runtime_readiness",
        "required_true_flags": {"canInspectReadinessLocally"},
        "required_false_flags": {
            "canRequestLiveBackend",
            "canPersistReadinessResult",
            "canStartWorkflowRuntime",
            "canExecuteWorkflow",
            "canSubmitConfirmationDecision",
            "canWriteBusinessTruth",
            "canReplayRun",
            "canEnableProductionAuth",
            "canAttachDatabase",
            "canImplementRepositoryAdapter",
        },
    },
}

EXPECTED_STOP_LINES = {
    "no new runtime API",
    "no live backend request",
    "no workflow builder mutation",
    "no workflow draft persistence",
    "no validation result persistence",
    "no execution plan persistence",
    "no runtime readiness persistence",
    "no workflow publish",
    "no workflow executor",
    "no node executor",
    "no tool executor",
    "no agent loop",
    "no confirmation decision",
    "no confirmation decision store",
    "no execution unlock",
    "no durable run store",
    "no durable result store",
    "no materialized result reader",
    "no business writeback",
    "no replay or resume",
    "no database attach",
    "no Radish OIDC implementation",
    "no repository adapter implementation",
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


def assert_fixture_shape(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "workflow_function_surface_readiness_close_v1", "unexpected fixture kind")

    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-function-surface-readiness-close-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_function_surface_readiness_closed",
        "workflow function surface readiness status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    close_scope = fixture.get("close_scope") or {}
    require(close_scope.get("status") == "workflow_function_surface_readiness_closed", "close status drifted")
    require(close_scope.get("app_root") == "apps/radishmind-web/", "app root drifted")
    require(
        close_scope.get("surface_model") == "offline_read_only_workflow_function_surface_matrix",
        "surface model drifted",
    )
    require(
        close_scope.get("evidence_model") == "aggregate_checker_over_existing_workflow_surface_slices",
        "evidence model drifted",
    )
    require(
        close_scope.get("current_backend_mode") == "offline_fixture_or_fixture_derived_view_model",
        "backend mode drifted",
    )
    for field in EXPECTED_CLOSE_FALSE_FIELDS:
        require(close_scope.get(field) is False, f"{field} must remain false")

    surfaces = fixture.get("surface_matrix") or []
    require(isinstance(surfaces, list), "surface_matrix must be a list")
    indexed = {str(surface.get("page_id")): surface for surface in surfaces if isinstance(surface, dict)}
    require(set(indexed) == set(EXPECTED_SURFACES), "surface matrix page ids drifted")

    for page_id, expected in EXPECTED_SURFACES.items():
        surface = indexed[page_id]
        for key in ("surface", "source_file", "builder", "render_component", "render_anchor", "render_class", "source_mode"):
            require(surface.get(key) == expected[key], f"{page_id} {key} drifted")
        require(
            set(surface.get("required_true_flags") or []) == expected["required_true_flags"],
            f"{page_id} required true flags drifted",
        )
        require(
            set(surface.get("required_false_flags") or []) == expected["required_false_flags"],
            f"{page_id} required false flags drifted",
        )

    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    app_source = read("apps/radishmind-web/src/app/App.tsx")
    styles = read("apps/radishmind-web/src/styles.css")
    workflow_sources: list[str] = []

    for page_id, expected in EXPECTED_SURFACES.items():
        page_source = read(str(expected["source_file"]))
        workflow_sources.append(page_source)

        for literal in (page_id, expected["builder"], "requestId", "auditRef", "routePath"):
            require(str(literal) in page_source, f"{page_id} missing source literal: {literal}")

        for flag in expected["required_true_flags"]:
            require(f"{flag}: true" in page_source, f"{page_id} missing true flag: {flag}")
        for flag in expected["required_false_flags"]:
            require(f"{flag}: false" in page_source, f"{page_id} missing false flag: {flag}")

        for literal in (expected["builder"], expected["render_component"], expected["render_anchor"]):
            require(str(literal) in app_source, f"App.tsx missing rendered binding for {page_id}: {literal}")
        require(str(expected["render_class"]) in app_source, f"App.tsx missing class for {page_id}")
        require(f".{expected['render_class']}" in styles, f"styles missing selector for {page_id}")

    checked_source = "\n".join([*workflow_sources, app_source])
    for forbidden in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden) not in checked_source, f"workflow surface source contains forbidden literal: {forbidden}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    commands = {str(item.get("command")) for item in fixture.get("testing_strategy") or [] if isinstance(item, dict)}
    expected_commands = {
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-function-surface-readiness-close-v1.py",
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-runtime-readiness-inspector-offline-v1.py",
        "npm run build",
        "macOS/Linux/WSL: ./scripts/check-repo.sh --fast; Windows/PowerShell: pwsh ./scripts/check-repo.ps1 -Fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing strategy commands: {missing_commands}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-runtime-readiness-inspector-offline-v1.py"
    current_checker = "check-workflow-function-surface-readiness-close-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow function surface readiness close check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow function surface readiness close check must run after runtime readiness inspector check",
    )

    doc_refs = fixture.get("required_doc_references") or {}
    require(isinstance(doc_refs, dict), "required_doc_references must be an object")
    for relative_path, required_literals in doc_refs.items():
        document = read(str(relative_path))
        missing = [literal for literal in required_literals if str(literal) not in document]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_fixture_shape(fixture)
    assert_required_files(fixture)
    assert_source_boundaries(fixture)
    assert_testing_strategy(fixture)
    assert_docs_and_fast_baseline(fixture)
    print("workflow function surface readiness close v1 checks passed.")


if __name__ == "__main__":
    main()
