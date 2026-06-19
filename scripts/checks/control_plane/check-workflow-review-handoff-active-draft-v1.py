#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-review-handoff-active-draft-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "review_persistence_ready",
    "handoff_persistence_ready",
    "handoff_export_ready",
    "handoff_send_ready",
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
    require(fixture.get("kind") == "workflow_review_handoff_active_draft_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-review-handoff-active-draft-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "workflow_review_handoff_active_draft_v1_implemented",
        "unexpected slice status",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_literals(text: str, literals: list[Any], label: str) -> None:
    missing = [literal for literal in literals if str(literal) not in text]
    require(not missing, f"{label} missing literals: {missing}")


def assert_frontend_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("frontend_contract") or {}
    handoff_text = read(str(contract.get("handoff_file")))
    panel_text = read(str(contract.get("handoff_panel_file")))
    context_text = read(str(contract.get("context_file")))

    assert_literals(handoff_text, contract.get("required_handoff_literals") or [], "workflowReviewHandoff.ts")
    assert_literals(panel_text, contract.get("required_panel_literals") or [], "workflowReviewHandoffPanel.tsx")
    assert_literals(context_text, contract.get("required_context_literals") or [], "workflowWorkspaceContext.ts")
    for forbidden in contract.get("forbidden_handoff_literals") or []:
        require(str(forbidden) not in handoff_text, f"workflowReviewHandoff.ts contains forbidden literal: {forbidden}")

    require(
        context_text.index("buildWorkflowDraftValidationInspectorViewModel(")
        < context_text.index("buildWorkflowExecutionPlanPreviewViewModel(")
        < context_text.index("buildWorkflowRuntimeReadinessInspectorViewModel(")
        < context_text.index("buildWorkflowReviewHandoffViewModel("),
        "workflow context must derive validation -> plan -> readiness before handoff",
    )
    require(
        handoff_text.index("buildActiveDraftReviewRecord(")
        < handoff_text.index("buildRecipients(")
        < handoff_text.index("buildKeyFindings(")
        < handoff_text.index("buildEvidenceChecklist("),
        "handoff must build active draft review record before recipients, findings, and evidence",
    )
    require(
        panel_text.index("Active Draft Review Record") < panel_text.index("Review Recipients"),
        "active draft review record must render before review recipients",
    )


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-draft-node-attribute-editing-model-v1.py"
    current_checker = "check-workflow-review-handoff-active-draft-v1.py"
    next_checker = "check-workflow-saved-draft-durable-store-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow review handoff active draft check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow review handoff active draft check must run after node attribute editing check",
    )
    require(
        check_repo.index(current_checker) < check_repo.index(next_checker),
        "workflow review handoff active draft check must run before durable store precondition checks",
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
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-review-handoff-active-draft-v1.py",
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
    print("workflow review handoff active draft v1 checks passed.")


if __name__ == "__main__":
    main()
