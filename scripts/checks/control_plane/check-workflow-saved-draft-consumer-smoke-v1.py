#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-consumer-smoke-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
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
    require(fixture.get("kind") == "workflow_saved_draft_consumer_smoke_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-saved-draft-consumer-smoke-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "saved_draft_consumer_smoke_guarded",
        "unexpected slice status",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_route_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("route_contract") or {}
    handler_text = read(str(contract.get("go_handler_file")))
    test_text = read(str(contract.get("go_test_file")))
    for route in contract.get("routes") or []:
        require(str(route) in handler_text, f"saved draft handler missing route: {route}")
    for header in contract.get("headers") or []:
        require(str(header) in handler_text, f"saved draft handler missing header: {header}")
    for header_constant in ("savedWorkflowDraftDevWorkspaceHeader", "savedWorkflowDraftDevApplicationHeader"):
        require(header_constant in test_text, f"saved draft test missing header constant: {header_constant}")
    config_text = read("services/platform/internal/config/config.go")
    config_test_text = read("services/platform/internal/config/config_test.go")
    for env_flag in contract.get("env_flags") or []:
        require(str(env_flag) in config_text, f"platform config missing env flag: {env_flag}")
        require(str(env_flag) in config_test_text, f"platform config test missing env flag: {env_flag}")
    failure_constant_by_code = {
        "draft_version_conflict": "SavedWorkflowDraftFailureVersionConflict",
        "draft_write_disabled": "SavedWorkflowDraftFailureWriteDisabled",
        "draft_scope_denied": "SavedWorkflowDraftFailureScopeDenied",
        "draft_not_found": "SavedWorkflowDraftFailureNotFound",
        "draft_store_unavailable": "SavedWorkflowDraftFailureStoreUnavailable",
    }
    for failure_code in contract.get("required_failure_codes") or []:
        failure_constant = failure_constant_by_code.get(str(failure_code), str(failure_code))
        require(failure_constant in test_text, f"saved draft route test missing failure code: {failure_code}")
    for envelope_key in contract.get("required_envelope_keys") or []:
        require(str(envelope_key) in handler_text, f"saved draft handler missing envelope key: {envelope_key}")
        require(str(envelope_key) in test_text, f"saved draft route test missing envelope key: {envelope_key}")
    for required_literal in (
        "assertSavedWorkflowDraftEnvelopeContract",
        "not found must stay fail closed without sample fallback",
        "store unavailable must stay fail closed without sample fallback",
        "CORS allowed headers missing",
    ):
        require(required_literal in test_text, f"saved draft route test missing contract literal: {required_literal}")


def assert_consumer_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("consumer_contract") or {}
    consumer_text = read(str(contract.get("file")))
    for status in contract.get("required_statuses") or []:
        require(f'"{status}"' in consumer_text, f"saved draft consumer missing status: {status}")
    for field in contract.get("required_state_fields") or []:
        require(str(field) in consumer_text, f"saved draft consumer missing state field: {field}")
    for literal in contract.get("required_literals") or []:
        require(str(literal) in consumer_text, f"saved draft consumer missing literal: {literal}")
    for helper_literal in (
        'state.status === "saved_dev_record"',
        'state.status === "version_conflict"',
        'state.status === "conflict_local_continued"',
    ):
        require(
            helper_literal in consumer_text,
            f"expected version helper missing retry status: {helper_literal}",
        )


def assert_app_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("app_contract") or {}
    app_text = read(str(contract.get("file")))
    for literal in contract.get("required_literals") or []:
        require(str(literal) in app_text, f"App missing saved draft consumer literal: {literal}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-workspace-context-consistency-v1.py"
    current_checker = "check-workflow-saved-draft-consumer-smoke-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow saved draft consumer smoke check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "workflow saved draft consumer smoke check must run after workflow workspace context check",
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
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-consumer-smoke-v1.py",
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
    assert_app_contract(fixture)
    assert_docs_and_fast_baseline(fixture)
    assert_testing_strategy(fixture)
    print("workflow saved draft consumer smoke v1 checks passed.")


if __name__ == "__main__":
    main()
