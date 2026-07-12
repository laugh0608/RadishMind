#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-node-designer-edge-editing-save-preconditions-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "edge_mutation_backend_ready",
    "react_flow_raw_edge_persistence",
    "handle_id_persistence_ready",
    "port_id_persistence_ready",
    "derived_edge_kind_persistence_ready",
    "runtime_order_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "durable_persistence_ready",
    "production_api_consumer_ready",
    "database_ready",
    "repository_adapter_ready",
    "radish_oidc_ready",
}

EXPECTED_ALLOWED_DRAFT_FIELDS = ["edgeId", "fromNodeId", "toNodeId", "conditionSummary"]
EXPECTED_SAVED_PAYLOAD_FIELDS = ["edge_id", "from_node_id", "to_node_id", "condition_summary"]
EXPECTED_PRECONDITIONS = {
    "known_active_draft_nodes",
    "no_self_loop",
    "no_duplicate_from_to_pair",
    "stable_edge_id",
    "non_empty_condition_summary",
    "controlled_on_connect_add_edge",
    "controlled_edge_delete",
    "validation_inspector_consumes_mutated_edges",
    "local_edit_unsaved_state",
    "no_react_flow_raw_edge",
}
EXPECTED_LAYOUT_REVIEW_FACTS = {
    "medium_width_designer_shell_breakpoint",
    "inspector_connected_edge_structured_action",
    "draft_edge_heading_structured_metadata",
    "long_edge_identifier_wrap",
    "narrow_width_full_row_remove_buttons",
}
EXPECTED_BUILDER_INTERACTION_FACTS = {
    "builder_interaction_status_bar",
    "node_quick_select_ui_only",
    "feedback_tone_for_connect_drag_delete",
    "editing_locked_state_visible",
    "selected_node_canvas_sync",
}
EXPECTED_VALIDATION_NAVIGATION_FACTS = {
    "active_validation_inspector_input",
    "structural_evidence_refs_navigation",
    "contract_check_target_navigation",
    "ui_only_validation_focus_state",
    "node_and_edge_validation_highlight",
}
EXPECTED_FORBIDDEN_PERSISTED_FIELDS = {
    "react_flow_edge",
    "handle_id",
    "port_id",
    "viewport",
    "selection",
    "validation_focus",
    "connection_preview",
    "visual_edge_style",
    "derived_edge_kind",
    "runtime_order",
    "run_input",
    "run_output",
    "confirmation_decision",
    "writeback_payload",
    "executor_result",
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
    require(
        fixture.get("kind") == "workflow_node_designer_validation_overlay_navigation_v1",
        "unexpected kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-node-designer-validation-overlay-navigation-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "workflow_node_designer_validation_overlay_navigation_v1_implemented",
        "unexpected slice status",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_edge_save_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("edge_save_contract") or {}
    require(
        contract.get("allowed_draft_fields") == EXPECTED_ALLOWED_DRAFT_FIELDS,
        "allowed draft edge field order drifted",
    )
    require(
        contract.get("saved_payload_fields") == EXPECTED_SAVED_PAYLOAD_FIELDS,
        "saved payload edge field order drifted",
    )
    missing_preconditions = sorted(EXPECTED_PRECONDITIONS - set(contract.get("required_preconditions") or []))
    require(not missing_preconditions, f"missing edge save preconditions: {missing_preconditions}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_PERSISTED_FIELDS - set(contract.get("forbidden_persisted_fields") or []))
    require(not missing_forbidden, f"missing forbidden persisted edge fields: {missing_forbidden}")
    missing_layout_facts = sorted(EXPECTED_LAYOUT_REVIEW_FACTS - set(contract.get("layout_review_facts") or []))
    require(not missing_layout_facts, f"missing layout review facts: {missing_layout_facts}")
    missing_interaction_facts = sorted(
        EXPECTED_BUILDER_INTERACTION_FACTS - set(contract.get("builder_interaction_facts") or [])
    )
    require(not missing_interaction_facts, f"missing builder interaction facts: {missing_interaction_facts}")
    missing_navigation_facts = sorted(
        EXPECTED_VALIDATION_NAVIGATION_FACTS - set(contract.get("validation_navigation_facts") or [])
    )
    require(not missing_navigation_facts, f"missing validation navigation facts: {missing_navigation_facts}")


def assert_frontend_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("frontend_contract") or {}
    node_designer_text = read(str(contract.get("node_designer_file")))
    consumer_text = read(str(contract.get("consumer_file")))
    app_text = read(str(contract.get("app_file")))
    validation_text = read(str(contract.get("validation_file")))
    style_text = read(str(contract.get("style_file")))

    assert_literals(node_designer_text, contract.get("required_node_designer_literals") or [], "workflowNodeDesigner.tsx")
    assert_literals(consumer_text, contract.get("required_consumer_literals") or [], "savedWorkflowDraftConsumer.ts")
    assert_literals(app_text, contract.get("required_app_literals") or [], "App.tsx")
    assert_literals(validation_text, contract.get("required_validation_literals") or [], "workflowDraftValidationInspector.ts")
    assert_literals(style_text, contract.get("required_style_literals") or [], "styles.css")
    assert_forbidden_literals(
        node_designer_text,
        contract.get("forbidden_runtime_literals") or [],
        "workflowNodeDesigner.tsx",
    )

    require(
        node_designer_text.index("validateWorkflowNodeDesignerConnection(connection, draft)")
        < node_designer_text.index("const added = onAddEdge(connection.source, connection.target)")
        < node_designer_text.index("Added draft edge:")
        < node_designer_text.index("function validateWorkflowNodeDesignerConnection("),
        "node designer onConnect must validate before controlled active draft mutation",
    )
    require(
        "Preview only:" not in node_designer_text and "Save mapping is not changed." not in node_designer_text,
        "node designer onConnect must no longer be preview-only after controlled edge mutation implementation",
    )
    require(
        consumer_text.index("edges: draft.edges.map((edge) => ({")
        < consumer_text.index("edge_id: edge.edgeId")
        < consumer_text.index("function toSavedWorkflowDraftAdditionalFields("),
        "saved draft consumer must serialize draft.edges before additional fields",
    )
    require(
        consumer_text.index("edgeId: edge.edge_id")
        < consumer_text.index("edgeKind: \"context\"")
        < consumer_text.index("designerLayout: workflowDraftDesignerLayoutFromSavedDraftAdditionalFields("),
        "saved draft restore must keep edge kind derived locally before layout restore",
    )
    require(
        app_text.index("const handleWorkflowDraftAddEdge")
        < app_text.index("const handleWorkflowDraftRemoveEdge")
        < app_text.index("const handleWorkflowDraftAddNode"),
        "draft edge mutation handlers must stay with the local active draft edit handlers",
    )
    require(
        app_text.index("function rebuildWorkflowDraftEdges(")
        < app_text.index("function buildWorkflowDraftEdge(")
        < app_text.index("function buildWorkflowDraftEdgeForConnection(")
        < app_text.index("function workflowDraftEdgeKindForConnection(")
        < app_text.index("function workflowDraftEdgeConditionSummary(")
        < app_text.index("function workflowDraftReviewableEdgeConditionSummary(")
        < app_text.index("function workflowDraftEdgeId("),
        "draft edge helper order drifted",
    )
    require(
        style_text.index("@media (max-width: 1280px)") < style_text.index("@media (max-width: 820px)"),
        "node designer medium width breakpoint must precede narrow width overrides",
    )
    require(
        node_designer_text.index("const selectNode = useCallback(")
        < node_designer_text.index("const focusValidationFinding = useCallback(")
        < node_designer_text.index("const displayedNodes = useMemo(")
        < node_designer_text.index("selected: node.data.draftNodeId === selectedNodeId")
        < node_designer_text.index("aria-pressed={node.nodeId === selectedNodeId}"),
        "node quick selection must stay UI-only and drive displayed node selection",
    )
    require(
        node_designer_text.index("buildWorkflowNodeDesignerValidationNavigation(draft, validationInspector)")
        < node_designer_text.index("const [validationFocus, setValidationFocus]")
        < node_designer_text.index("setValidationFocus({")
        < node_designer_text.index("validationFocus?.nodeIds.includes(node.data.draftNodeId)")
        < node_designer_text.index("validationFocus?.edgeIds.includes(edge.id)"),
        "validation overlay navigation must consume active validation inspector and stay UI-only",
    )
    structural_target_index = node_designer_text.index("check.evidenceRefs.filter((nodeId) => nodeIds.has(nodeId))")
    contract_target_index = node_designer_text.index("workflowNodeDesignerContractTargetNodeIds(draft, check.checkId)")
    first_edge_target_index = node_designer_text.index("workflowNodeDesignerEdgeIdsForTargets(draft, targetNodeIds)")
    last_edge_target_index = node_designer_text.rindex("workflowNodeDesignerEdgeIdsForTargets(draft, targetNodeIds)")
    require(
        structural_target_index < first_edge_target_index and contract_target_index < last_edge_target_index,
        "validation navigation must map structural evidence refs and contract targets to active draft graph",
    )


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-node-designer-persisted-layout-v1.py"
    current_checker = "check-workflow-node-designer-edge-editing-save-preconditions-v1.py"
    require(current_checker in check_repo, "check-repo.py must run node designer edge editing preconditions check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "edge editing preconditions check must run after node designer persisted layout check",
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
        "npm run build",
        "./scripts/run-python.sh scripts/checks/control_plane/check-workflow-node-designer-edge-editing-save-preconditions-v1.py",
        "git diff --check",
        "./scripts/check-repo.sh --fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing strategy commands: {missing_commands}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_fixture_shape(fixture)
    assert_edge_save_contract(fixture)
    assert_frontend_contract(fixture)
    assert_docs_and_fast_baseline(fixture)
    assert_testing_strategy(fixture)
    print("workflow node designer validation overlay navigation v1 checks passed.")


if __name__ == "__main__":
    main()
