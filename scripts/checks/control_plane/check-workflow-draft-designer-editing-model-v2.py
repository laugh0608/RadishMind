#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-draft-designer-editing-model-v2.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "complete_canvas_builder_ready",
    "durable_persistence_ready",
    "production_api_consumer_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "database_ready",
    "repository_adapter_ready",
    "store_selector_ready",
    "radish_oidc_ready",
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
    require(fixture.get("kind") == "workflow_draft_designer_editing_model_v2", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-draft-designer-editing-model-v2", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "workflow_draft_designer_editing_model_v2_implemented",
        "unexpected slice status",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_frontend_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("frontend_contract") or {}
    app_text = read(str(contract.get("app_file")))
    context_text = read(str(contract.get("context_file")))
    consumer_text = read(str(contract.get("consumer_file")))
    style_text = read(str(contract.get("style_file")))

    for literal in contract.get("required_app_literals") or []:
        require(str(literal) in app_text, f"App missing editing model literal: {literal}")
    for literal in contract.get("required_context_literals") or []:
        require(str(literal) in context_text, f"workflow workspace context missing literal: {literal}")
    for literal in contract.get("required_consumer_literals") or []:
        require(str(literal) in consumer_text, f"saved draft consumer missing literal: {literal}")
    for selector in contract.get("required_style_selectors") or []:
        require(str(selector) in style_text, f"styles.css missing selector: {selector}")

    require(
        "buildWorkflowDraftValidationInspectorViewModel(" not in app_text,
        "App must not compose workflow validation inspector directly",
    )
    require(
        "buildWorkflowExecutionPlanPreviewViewModel(" not in app_text,
        "App must not compose workflow execution plan preview directly",
    )
    require(
        "buildWorkflowRuntimeReadinessInspectorViewModel(" not in app_text,
        "App must not compose workflow runtime readiness inspector directly",
    )
    require(
        context_text.index("const activeWorkflowDraft =")
        < context_text.index("buildWorkflowDraftValidationInspectorViewModel("),
        "validation inspector must be derived from context-owned active workflow draft",
    )
    require(
        context_text.index("buildWorkflowDraftValidationInspectorViewModel(")
        < context_text.index("buildWorkflowExecutionPlanPreviewViewModel("),
        "execution plan preview must consume context-owned active draft validation inspector",
    )
    require(
        context_text.index("buildWorkflowExecutionPlanPreviewViewModel(")
        < context_text.index("buildWorkflowRuntimeReadinessInspectorViewModel("),
        "runtime readiness inspector must consume context-owned active execution plan preview",
    )
    require(
        app_text.index("function workflowDraftWithStructureEdits")
        < app_text.index("function rebuildWorkflowDraftEdges"),
        "structure edit helper must centralize edge rebuild before graph edge helper",
    )
    require(
        'localOnlyInteraction: "local_edit"' in app_text[app_text.index("function workflowDraftWithStructureEdits"):],
        "structure edits must mark drafts as local_edit",
    )
    require(
        "rebuildWorkflowDraftEdges(remainingNodes, draft.edges).length >= 3" in app_text,
        "remove guard must verify the rebuilt graph keeps enough edges",
    )
    require(
        "case \"http_tool\":" not in app_text,
        "App should not duplicate saved restore lane mapping",
    )
    require(
        consumer_text.index("case \"http_tool\":") < consumer_text.index("return \"preview\";"),
        "http_tool must restore into preview lane",
    )


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-workflow-draft-designer-editing-model-v2.py"
    required_previous = [
        "check-workflow-draft-editing-entry-v1.py",
        "check-user-workspace-saved-draft-list-v1.py",
    ]
    require(current_checker in check_repo, "check-repo.py must run workflow draft designer editing model v2 check")
    for previous_checker in required_previous:
        require(
            check_repo.index(previous_checker) < check_repo.index(current_checker),
            f"{current_checker} must run after {previous_checker}",
        )

    doc_refs = fixture.get("required_doc_references") or {}
    require(isinstance(doc_refs, dict), "required_doc_references must be an object")
    for relative_path, required_literals in doc_refs.items():
        document = read(str(relative_path))
        missing = [literal for literal in required_literals if str(literal) not in document]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    commands = {str(item.get("command")) for item in fixture.get("testing_strategy") or [] if isinstance(item, dict)}
    expected_commands = {
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-draft-designer-editing-model-v2.py",
        "npm run build",
        "./scripts/check-repo.sh --fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing strategy commands: {missing_commands}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_fixture_shape(fixture)
    assert_frontend_contract(fixture)
    assert_docs_and_fast_baseline(fixture)
    assert_testing_strategy(fixture)
    print("workflow draft designer editing model v2 checks passed.")


if __name__ == "__main__":
    main()
