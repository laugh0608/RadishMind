#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-node-designer-persisted-layout-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "react_flow_raw_state_persistence",
    "viewport_persistence_ready",
    "selection_persistence_ready",
    "derived_edge_kind_persistence_ready",
    "handoff_persistence_ready",
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


def assert_literals(text: str, literals: list[Any], label: str) -> None:
    missing = [literal for literal in literals if str(literal) not in text]
    require(not missing, f"{label} missing literals: {missing}")


def assert_forbidden_literals(text: str, literals: list[Any], label: str) -> None:
    found = [literal for literal in literals if str(literal) in text]
    require(not found, f"{label} contains forbidden literals: {found}")


def assert_fixture_shape(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "workflow_node_designer_persisted_layout_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-node-designer-persisted-layout-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "workflow_node_designer_persisted_layout_v1_implemented",
        "unexpected slice status",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_schema_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("schema_contract") or {}
    require(contract.get("additional_field") == "designer_layout_v1", "additional field key drifted")
    require(contract.get("layout_version") == "designer_layout_v1", "layout version drifted")
    require(contract.get("source") == "workflow_node_designer", "layout source drifted")
    require(contract.get("persistence") == "saved_draft_metadata", "layout persistence drifted")
    require(contract.get("coordinate_min") == -10000, "coordinate min drifted")
    require(contract.get("coordinate_max") == 10000, "coordinate max drifted")
    require(contract.get("node_fields") == ["node_id", "x", "y", "pinned"], "node field order drifted")
    require(contract.get("pinned_value") is False, "pinned must stay false in v1")
    forbidden_fields = set(contract.get("forbidden_persisted_fields") or [])
    for required in {
        "react_flow_nodes",
        "react_flow_edges",
        "viewport",
        "selection",
        "derived_edge_kind",
        "secret_value",
        "token",
        "confirmation_decision",
        "run_output",
        "writeback_payload",
        "executor_result",
    }:
        require(required in forbidden_fields, f"missing forbidden persisted field {required}")


def assert_frontend_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("frontend_contract") or {}
    draft_type_text = read(str(contract.get("draft_type_file")))
    consumer_text = read(str(contract.get("consumer_file")))
    app_text = read(str(contract.get("app_file")))
    node_designer_text = read(str(contract.get("node_designer_file")))
    handoff_text = read(str(contract.get("handoff_file")))

    assert_literals(draft_type_text, contract.get("required_type_literals") or [], "workflowDraftDesigner.ts")
    assert_literals(consumer_text, contract.get("required_consumer_literals") or [], "savedWorkflowDraftConsumer.ts")
    assert_literals(app_text, contract.get("required_app_literals") or [], "App.tsx")
    assert_literals(node_designer_text, contract.get("required_node_designer_literals") or [], "workflowNodeDesigner.tsx")
    assert_literals(handoff_text, contract.get("required_handoff_literals") or [], "workflowReviewHandoff.ts")
    assert_forbidden_literals(consumer_text, contract.get("forbidden_consumer_literals") or [], "savedWorkflowDraftConsumer.ts")

    require(
        consumer_text.index("const additionalFields = toSavedWorkflowDraftAdditionalFields(draft);")
        < consumer_text.index("additional_fields: additionalFields")
        < consumer_text.index("function workflowDraftFromSavedWorkflowDraftDocument("),
        "consumer must serialize layout before restore mapping",
    )
    require(
        consumer_text.index("function isSavedWorkflowDraftDesignerLayoutV1(")
        < consumer_text.index("function isSavedWorkflowDraftDesignerLayoutNode(")
        < consumer_text.index("function workflowDraftDesignerLayoutCoordinate("),
        "restore guard and coordinate clamp order drifted",
    )
    require(
        node_designer_text.index("Saved draft mapping")
        < node_designer_text.index("Layout metadata")
        < node_designer_text.index("Derived edge kind"),
        "node designer mapping summary order drifted",
    )


def assert_backend_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("backend_contract") or {}
    domain_text = read(str(contract.get("domain_file")))
    http_text = read(str(contract.get("http_file")))
    domain_test_text = read(str(contract.get("domain_test_file")))
    http_test_text = read(str(contract.get("http_test_file")))

    assert_literals(domain_text, contract.get("required_domain_literals") or [], "workflow_saved_draft.go")
    assert_literals(http_text, contract.get("required_http_literals") or [], "workflow_saved_draft_http.go")
    assert_literals(
        domain_test_text + "\n" + http_test_text,
        contract.get("required_test_literals") or [],
        "saved draft Go tests",
    )
    assert_forbidden_literals(
        domain_text + "\n" + http_text,
        contract.get("forbidden_backend_literals") or [],
        "saved draft Go implementation",
    )
    require(
        domain_text.index("firstSavedWorkflowDraftForbiddenField(payload.AdditionalFields)")
        < domain_text.index("normalizeSavedWorkflowDraftPayload(payload)"),
        "forbidden field scan must run before layout normalization",
    )
    require(
        domain_text.index("func normalizeSavedWorkflowDraftAdditionalFields(")
        < domain_text.index("func normalizeSavedWorkflowDraftDesignerLayout(")
        < domain_text.index("func cloneSavedWorkflowDraftAdditionalFields("),
        "additional fields normalization helpers drifted",
    )
    require(
        http_text.index("func savedWorkflowDraftPayloadFromDocument(")
        < http_text.index("AdditionalFields:      document.AdditionalFields")
        < http_text.index("func savedWorkflowDraftPayloadDocumentFromDraftPayload("),
        "HTTP document must parse additional_fields before response serialization",
    )


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-node-designer-review-handoff-v1.py"
    current_checker = "check-workflow-node-designer-persisted-layout-v1.py"
    next_checker = "check-workflow-saved-draft-durable-store-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run node designer persisted layout check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "persisted layout check must run after node designer review handoff check",
    )
    require(
        check_repo.index(current_checker) < check_repo.index(next_checker),
        "persisted layout check must run before durable store precondition checks",
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
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-node-designer-persisted-layout-v1.py",
        "go test ./...",
        "npm run build",
        "./scripts/check-repo.sh --fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing strategy commands: {missing_commands}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_fixture_shape(fixture)
    assert_schema_contract(fixture)
    assert_frontend_contract(fixture)
    assert_backend_contract(fixture)
    assert_docs_and_fast_baseline(fixture)
    assert_testing_strategy(fixture)
    print("workflow node designer persisted layout v1 checks passed.")


if __name__ == "__main__":
    main()
