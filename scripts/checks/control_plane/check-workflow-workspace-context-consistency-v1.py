#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-workspace-context-consistency-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "workflow-function-surface-readiness-close-v1": (
        "scripts/checks/fixtures/workflow-function-surface-readiness-close-v1.json",
        "workflow_function_surface_readiness_closed",
    ),
    "control-plane-read-product-sample-consistency-v1": (
        "scripts/checks/fixtures/control-plane-read-product-sample-consistency-v1.json",
        "product_sample_consistency_guarded",
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
    "scenario_persistence_ready",
    "review_persistence_ready",
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


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def assert_fixture_shape(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "workflow_workspace_context_consistency_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-workspace-context-consistency-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_workspace_context_consistency_guarded",
        "unexpected slice status",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    require(set(EXPECTED_DEPENDENCIES).issubset(declared), "context fixture must declare dependencies")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"dependency {dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"dependency {dependency_id} status drifted")


def assert_context_source(fixture: dict[str, Any]) -> None:
    context_source = fixture.get("context_source") or {}
    context_text = read(str(context_source.get("file")))
    require(context_source.get("builder") in context_text, "context builder missing")
    for helper in context_source.get("selection_helpers") or []:
        require(f"export function {helper}" in context_text, f"context selection helper missing: {helper}")
    for output in context_source.get("owned_outputs") or []:
        require(str(output) in context_text, f"context output missing: {output}")

    required_builder_calls = {
        "buildWorkflowApplicationDetailViewModel(",
        "buildWorkflowDefinitionDetailViewModel(",
        "buildWorkflowRunDetailViewModel(",
        "buildWorkflowBlockedActionPreviewViewModel(",
        "buildWorkflowConfirmationPlaceholderViewModel(",
        "buildWorkflowDraftDesignerViewModel(",
        "buildWorkflowDraftValidationInspectorViewModel(",
        "buildWorkflowExecutionPlanPreviewViewModel(",
        "buildWorkflowRuntimeReadinessInspectorViewModel(",
        "buildWorkflowSurfaceOverviewViewModel(",
        "buildWorkflowScenarioInspectorViewModel(",
        "buildWorkflowWorkspaceReviewViewModel(",
        "buildWorkflowUserWorkspaceHomeViewModel(",
    }
    missing_calls = sorted(call for call in required_builder_calls if call not in context_text)
    require(not missing_calls, f"context builder lost owned builder calls: {missing_calls}")

    for required_literal in (
        "detailSourcesByWorkflowDefinitionId",
        "toDraftDesignerDetailSources(workflowDefinitionDetailsById)",
        "source.selection.scenarioId",
    ):
        require(required_literal in context_text, f"context source missing literal: {required_literal}")


def assert_app_composition(fixture: dict[str, Any]) -> None:
    app_composition = fixture.get("app_composition") or {}
    app_text = read(str(app_composition.get("file")))
    require(app_composition.get("required_builder") in app_text, "App must use workflow workspace context builder")
    require("const workflowWorkspaceContext = useMemo(" in app_text, "App must memoize workflow workspace context")
    require("} = workflowWorkspaceContext;" in app_text, "App must destructure workflow context outputs")
    for helper in app_composition.get("required_selection_helpers") or []:
        require(str(helper) in app_text, f"App missing selection helper: {helper}")
    for forbidden in app_composition.get("forbidden_inline_builders") or []:
        require(str(forbidden) not in app_text, f"App still composes workflow surface directly: {forbidden}")
    for forbidden in (
        "function toWorkflowDefinitionSummary(",
        "const nextDefinition =",
        "const nextDraft =",
    ):
        require(forbidden not in app_text, f"App still carries inline workflow context logic: {forbidden}")


def assert_draft_designer_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("draft_designer_guard") or {}
    draft_text = read(str(guard.get("file")))
    for literal in guard.get("required_literals") or []:
        require(str(literal) in draft_text, f"draft designer missing guard literal: {literal}")
    for forbidden in guard.get("forbidden_literals") or []:
        require(str(forbidden) not in draft_text, f"draft designer still has shared-detail drift risk: {forbidden}")
    require(
        "source.workflowDefinitions?.length === 1" in draft_text,
        "single-detail fallback must only be allowed for single-definition source",
    )


def assert_product_contexts(fixture: dict[str, Any]) -> None:
    applications = read("apps/radishmind-web/src/features/control-plane-read/workspaceApplications.ts")
    definitions = read("apps/radishmind-web/src/features/control-plane-read/workspaceWorkflowDefinitions.ts")
    runs = read("apps/radishmind-web/src/features/control-plane-read/workspaceRunHistory.ts")
    definition_detail = read("apps/radishmind-web/src/features/control-plane-read/workflowDefinitionDetail.ts")
    draft_designer = read("apps/radishmind-web/src/features/control-plane-read/workflowDraftDesigner.ts")
    scenario_inspector = read("apps/radishmind-web/src/features/control-plane-read/workflowScenarioInspector.ts")

    for context in fixture.get("required_product_contexts") or []:
        application_ref = str(context.get("application_ref") or "")
        workflow_definition_id = str(context.get("workflow_definition_id") or "")
        run_id = str(context.get("run_id") or "")
        scenario_prefix = str(context.get("scenario_prefix") or "")
        node_prefix = str(context.get("node_prefix") or "")

        require(application_ref in applications, f"application source missing {application_ref}")
        require(workflow_definition_id in definitions, f"definition source missing {workflow_definition_id}")
        require(run_id in runs, f"run source missing {run_id}")
        require(application_ref in definition_detail, f"definition detail missing {application_ref}")
        require(node_prefix in definition_detail, f"definition detail missing node prefix {node_prefix}")
        require(scenario_prefix in scenario_inspector, f"scenario source missing {scenario_prefix}")

    require("draft_${workflowDefinition.workflowDefinitionId}_offline" in draft_designer, "draft id derivation drifted")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    commands = {str(item.get("command")) for item in fixture.get("testing_strategy") or [] if isinstance(item, dict)}
    expected_commands = {
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-workspace-context-consistency-v1.py",
        "npm run build",
        "./scripts/check-repo.sh --fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing strategy commands: {missing_commands}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-function-surface-readiness-close-v1.py"
    current_checker = "check-workflow-workspace-context-consistency-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow workspace context consistency check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow workspace context check must run after workflow surface close check",
    )

    doc_refs = fixture.get("required_doc_references") or {}
    require(isinstance(doc_refs, dict), "required_doc_references must be an object")
    for relative_path, required_literals in doc_refs.items():
        document = read(str(relative_path))
        missing = [literal for literal in required_literals if str(literal) not in document]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_fixture_shape(fixture)
    assert_dependencies(fixture)
    assert_context_source(fixture)
    assert_app_composition(fixture)
    assert_draft_designer_guard(fixture)
    assert_product_contexts(fixture)
    assert_testing_strategy(fixture)
    assert_docs_and_fast_baseline(fixture)
    print("workflow workspace context consistency v1 checks passed.")


if __name__ == "__main__":
    main()
