#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/user-workspace-saved-draft-list-v1.json"
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
    require(fixture.get("kind") == "user_workspace_saved_draft_list_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "user-workspace-saved-draft-list-v1", "unexpected slice id")
    require(slice_info.get("track") == "User Workspace / Workflow", "unexpected track")
    require(
        slice_info.get("status") == "user_workspace_saved_draft_list_implemented",
        "unexpected slice status",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_route_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("route_contract") or {}
    domain_text = read(str(contract.get("go_domain_file")))
    handler_text = read(str(contract.get("go_handler_file")))
    server_text = read(str(contract.get("go_server_file")))
    test_text = read(str(contract.get("go_test_file")))

    for route in contract.get("routes") or []:
        require(str(route) in handler_text, f"saved draft list handler missing route: {route}")
    for literal in contract.get("required_domain_literals") or []:
        require(str(literal) in domain_text, f"saved draft domain missing literal: {literal}")
    for literal in contract.get("required_handler_literals") or []:
        require(str(literal) in handler_text, f"saved draft list handler missing literal: {literal}")
    for literal in contract.get("required_test_literals") or []:
        require(str(literal) in test_text, f"saved draft list test missing literal: {literal}")
    require("mux.HandleFunc(savedWorkflowDraftListRoute" in server_text, "server must register saved draft list route")
    require(
        "SavedWorkflowDraftSummary" in domain_text and "SavedWorkflowDraftPayload" in domain_text,
        "saved draft list must use a summary projection instead of returning payload collections",
    )
    require(
        domain_text.index("type SavedWorkflowDraftSummary struct")
        < domain_text.index("type SavedWorkflowDraftListResult struct"),
        "summary projection must be defined before list result",
    )
    for key in contract.get("required_summary_keys") or []:
        require(str(key) in handler_text, f"saved draft list handler missing summary key: {key}")
        require(str(key) in test_text, f"saved draft list test missing summary key: {key}")


def assert_consumer_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("consumer_contract") or {}
    consumer_text = read(str(contract.get("file")))
    for status in contract.get("required_statuses") or []:
        require(f'"{status}"' in consumer_text, f"saved draft list consumer missing status: {status}")
    for exported in contract.get("required_exports") or []:
        require(str(exported) in consumer_text, f"saved draft list consumer missing export: {exported}")
    for literal in contract.get("required_literals") or []:
        require(str(literal) in consumer_text, f"saved draft list consumer missing literal: {literal}")
    require(
        "workflowDraftFromSavedWorkflowDraftDocument" in consumer_text,
        "restore must project saved record into a Draft Designer draft",
    )
    require(
        consumer_text.index("listWorkflowDraftDevRecords")
        < consumer_text.index("restoreWorkflowDraftDevRecord"),
        "list helper should be defined before restore helper",
    )


def assert_app_and_home_contract(fixture: dict[str, Any]) -> None:
    app_contract = fixture.get("app_contract") or {}
    app_text = read(str(app_contract.get("file")))
    for literal in app_contract.get("required_literals") or []:
        require(str(literal) in app_text, f"App missing saved draft list literal: {literal}")
    require(
        app_text.index("listWorkflowDraftDevRecords(applicationRef")
        < app_text.index("restoreWorkflowDraftDevRecord(summary"),
        "App must list summaries before restore path",
    )

    home_contract = fixture.get("home_panel_contract") or {}
    home_text = read(str(home_contract.get("file")))
    style_text = read(str(home_contract.get("style_file")))
    for literal in home_contract.get("required_literals") or []:
        require(str(literal) in home_text, f"Workspace Home panel missing literal: {literal}")
    for selector in home_contract.get("required_style_selectors") or []:
        require(str(selector) in style_text, f"styles.css missing selector: {selector}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    current_checker = "check-user-workspace-saved-draft-list-v1.py"
    required_previous = [
        "check-workflow-saved-draft-consumer-smoke-v1.py",
        "check-user-workspace-draft-creation-v1.py",
    ]
    require(current_checker in check_repo, "check-repo.py must run user workspace saved draft list check")
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
        "./scripts/run-python.sh scripts/checks/control_plane/check-user-workspace-saved-draft-list-v1.py",
        "go test ./...",
        "npm run build",
        "./scripts/check-repo.sh --fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing strategy commands: {missing_commands}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_fixture_shape(fixture)
    assert_route_contract(fixture)
    assert_consumer_contract(fixture)
    assert_app_and_home_contract(fixture)
    assert_docs_and_fast_baseline(fixture)
    assert_testing_strategy(fixture)
    print("user workspace saved draft list v1 checks passed.")


if __name__ == "__main__":
    main()
