#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-draft-editing-entry-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "complete_builder_ready",
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
    require(fixture.get("kind") == "workflow_draft_editing_entry_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-draft-editing-entry-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "workflow_draft_editing_entry_implemented",
        "unexpected slice status",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_frontend_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("frontend_contract") or {}
    app_text = read(str(contract.get("app_file")))
    designer_text = read(str(contract.get("designer_file")))
    style_text = read(str(contract.get("style_file")))

    for literal in contract.get("required_app_literals") or []:
        require(str(literal) in app_text, f"App missing workflow draft editing literal: {literal}")
    for literal in contract.get("required_designer_literals") or []:
        require(str(literal) in designer_text, f"workflowDraftDesigner missing literal: {literal}")
    for selector in contract.get("required_style_selectors") or []:
        require(str(selector) in style_text, f"styles.css missing selector: {selector}")

    require(
        app_text.index("activeWorkflowDraftOverride: editableWorkflowDraft")
        < app_text.index("validateWorkflowDraftDevRecord(activeWorkflowDraft"),
        "validate must consume context-owned active workflow draft after local edit state is wired",
    )
    require(
        app_text.index("activeWorkflowDraftOverride: editableWorkflowDraft")
        < app_text.index("saveWorkflowDraftDevRecord(activeWorkflowDraft"),
        "save must consume context-owned active workflow draft after local edit state is wired",
    )
    require(
        app_text.index("activeWorkflowDraftOverride: editableWorkflowDraft")
        < app_text.index("readWorkflowDraftDevRecord(activeWorkflowDraft"),
        "read must consume context-owned active workflow draft after local edit state is wired",
    )

    for edit_handler in (
        "handleWorkflowDraftLabelChange",
        "handleWorkflowDraftSummaryChange",
        "handleWorkflowDraftNodeLabelChange",
        "handleWorkflowDraftEdgeConditionChange",
    ):
        handler_index = app_text.index(edit_handler)
        local_edit_index = app_text.find('localOnlyInteraction: "local_edit"', handler_index)
        require(local_edit_index != -1, f"{edit_handler} must mark draft as local_edit")

    for literal in (
        "disabled={operationPending}",
        "disabled={!draftEditDirty || operationPending}",
        "onResetDraftEdits",
        "onUpdateNodeLabel",
        "onUpdateEdgeCondition",
    ):
        require(literal in app_text, f"workflow draft editing UI missing literal: {literal}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-saved-draft-consumer-smoke-v1.py"
    current_checker = "check-workflow-draft-editing-entry-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow draft editing entry check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow draft editing entry check must run after saved draft consumer smoke check",
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
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-draft-editing-entry-v1.py",
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
    print("workflow draft editing entry v1 checks passed.")


if __name__ == "__main__":
    main()
