#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-definition-run-record-boundary.json"
DEPENDENCY_FIXTURES = {
    "product-surface-v1-boundary": REPO_ROOT / "scripts/checks/fixtures/product-surface-v1-boundary.json",
    "control-plane-data-boundary": REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json",
    "radish-oidc-client-preconditions": REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json",
    "gateway-api-key-quota-readiness": REPO_ROOT / "scripts/checks/fixtures/gateway-api-key-quota-readiness.json",
}
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_RESOURCE_IDS = {
    "workflow-definition",
    "run-record",
    "node-execution",
    "tool-audit",
    "result-materialization",
    "confirmation-decision",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "workflow_builder_ready",
    "workflow_executor_ready",
    "node_executor_ready",
    "tool_executor_ready",
    "durable_run_store_ready",
    "materialized_result_reader_ready",
    "confirmation_flow_ready",
    "writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "workflow builder",
    "workflow executor",
    "node execution engine",
    "HTTP tool execution",
    "RAG retrieval execution",
    "code sandbox execution",
    "agent loop execution",
    "confirmation flow wiring",
    "business writeback",
    "replay execution",
    "durable run database",
    "materialized result reader",
}
EXPECTED_STATES = {
    "created",
    "queued",
    "running",
    "blocked_confirmation_required",
    "succeeded",
    "failed",
    "cancelled",
}
EXPECTED_FAILURES = {
    "definition_invalid",
    "input_validation_failed",
    "policy_denied",
    "confirmation_required",
    "confirmation_denied",
    "tool_not_available",
    "tool_execution_blocked",
    "provider_error",
    "timeout",
    "output_validation_failed",
    "result_materialization_blocked",
    "writeback_forbidden",
    "replay_forbidden",
}
EXPECTED_STOP_LINES = {
    "no workflow builder",
    "no workflow executor",
    "no node executor",
    "no tool calling",
    "no confirmation flow wiring",
    "no business writeback",
    "no replay execution",
    "no durable run database",
    "no materialized result reader",
    "no production readiness claim",
}
EXPECTED_FORBIDDEN_EVIDENCE = {
    "raw_secret_value",
    "raw_api_key",
    "authorization_header",
    "cookie_value",
    "unsanitized_tool_payload",
    "full_prompt_dump_with_secret",
    "business_writeback_payload",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "workflow-definition-run-record-boundary",
        "workflow definition",
        "run record",
    ],
    "docs/radishmind-product-scope.md": [
        "workflow-definition-run-record-boundary",
        "状态流转",
        "审计证据",
    ],
    "docs/radishmind-architecture.md": [
        "workflow-definition-run-record-boundary",
        "run record",
        "状态流转",
    ],
    "docs/radishmind-roadmap.md": [
        "workflow-definition-run-record-boundary",
        "workflow-definition-run-record-boundary.json",
        "check-workflow-definition-run-record-boundary.py",
    ],
    "docs/radishmind-capability-matrix.md": [
        "workflow-definition-run-record-boundary",
        "workflow-definition-run-record-boundary.json",
        "check-workflow-definition-run-record-boundary.py",
    ],
    "docs/task-cards/control-plane-user-workspace-workflow-v1-plan.md": [
        "workflow-definition-run-record-boundary",
        "workflow-definition-run-record-boundary.json",
        "check-workflow-definition-run-record-boundary.py",
        "governance_boundary_satisfied",
    ],
    "scripts/README.md": [
        "check-workflow-definition-run-record-boundary.py",
        "workflow-definition-run-record-boundary.json",
        "workflow definition、run record、状态流转",
    ],
    "docs/devlogs/2026-W22.md": [
        "workflow-definition-run-record-boundary",
        "workflow-definition-run-record-boundary.json",
        "check-workflow-definition-run-record-boundary.py",
    ],
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
    require(set(DEPENDENCY_FIXTURES).issubset(declared), "fixture must declare all dependency slices")
    for expected_id, path in DEPENDENCY_FIXTURES.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(
            slice_info.get("id") == expected_id
            and slice_info.get("status") == "governance_boundary_satisfied",
            f"dependency {expected_id} must remain satisfied",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "workflow_definition_run_record_boundary", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-definition-run-record-boundary", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "workflow definition run record boundary must only satisfy a governance boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_resource_boundaries(fixture: dict[str, Any]) -> None:
    resources = {
        str(item.get("id") or ""): item
        for item in fixture.get("resource_boundaries") or []
        if isinstance(item, dict)
    }
    require(set(resources) == EXPECTED_RESOURCE_IDS, f"unexpected resource boundaries: {sorted(resources)}")

    for resource_id, resource in resources.items():
        require(resource.get("status") == "defined_not_implemented", f"{resource_id} must remain defined_not_implemented")
        require(str(resource.get("boundary") or "").strip(), f"{resource_id}.boundary is required")
        required_fields = resource.get("required_fields")
        require(isinstance(required_fields, list) and required_fields, f"{resource_id}.required_fields must be non-empty")

    require(
        {"workflow_definition_id", "graph_nodes", "graph_edges", "risk_policy_ref"}.issubset(
            set(resources["workflow-definition"].get("required_fields") or [])
        ),
        "workflow definition must define graph and risk policy fields",
    )
    require(
        {"run_id", "status", "failure_code", "cost_summary", "trace_id", "audit_refs"}.issubset(
            set(resources["run-record"].get("required_fields") or [])
        ),
        "run record must define status, failure, cost, trace and audit fields",
    )
    require(
        {"prompt", "llm", "http_tool", "rag_retrieval", "condition", "output"}.issubset(
            set(resources["node-execution"].get("allowed_node_types") or [])
        ),
        "node execution must include minimum workflow node classes",
    )
    require(
        "requires_confirmation" in set(resources["tool-audit"].get("required_fields") or []),
        "tool audit must carry requires_confirmation",
    )
    require(
        "sanitized_payload_ref" in set(resources["result-materialization"].get("required_fields") or []),
        "result materialization must use sanitized payload refs",
    )
    require(
        {"approved", "denied", "expired", "cancelled"}.issubset(
            set(resources["confirmation-decision"].get("allowed_decisions") or [])
        ),
        "confirmation decision must define terminal decision values",
    )


def assert_state_machine(fixture: dict[str, Any]) -> None:
    state_machine = fixture.get("state_machine") or {}
    require(EXPECTED_STATES.issubset(set(state_machine.get("allowed_states") or [])), "state machine missing states")
    require(
        {"succeeded", "failed", "cancelled"}.issubset(set(state_machine.get("terminal_states") or [])),
        "state machine must define terminal states",
    )
    transitions = state_machine.get("allowed_transitions") or []
    require(isinstance(transitions, list) and transitions, "state machine must define transitions")
    transition_pairs = {
        f"{transition.get('from')}->{transition.get('to')}"
        for transition in transitions
        if isinstance(transition, dict)
    }
    for expected_pair in (
        "created->queued",
        "queued->running",
        "running->blocked_confirmation_required",
        "blocked_confirmation_required->running",
        "running->succeeded",
        "running->failed",
    ):
        require(expected_pair in transition_pairs, f"missing transition: {expected_pair}")

    for transition in transitions:
        require(str(transition.get("audit_event") or "").strip(), "every transition must define audit_event")

    forbidden_transitions = set(state_machine.get("forbidden_transitions") or [])
    require("created->succeeded" in forbidden_transitions, "state machine must forbid bypass to success")
    require("failed->running" in forbidden_transitions, "state machine must forbid failed run restart")
    require("audit_event" in str(state_machine.get("transition_policy") or ""), "transition policy must require audit")


def assert_failures_audit_and_stop_lines(fixture: dict[str, Any]) -> None:
    missing_failures = sorted(EXPECTED_FAILURES - set(fixture.get("failure_taxonomy") or []))
    require(not missing_failures, f"missing failure taxonomy entries: {missing_failures}")

    audit_policy = fixture.get("audit_evidence_policy") or {}
    require(audit_policy.get("sanitization_required") is True, "audit evidence must require sanitization")
    required_refs = set(audit_policy.get("required_refs") or [])
    require(
        {"trace_id", "run_id", "workflow_definition_id", "failure_code"}.issubset(required_refs),
        "audit evidence must include trace, run, definition and failure refs",
    )
    missing_forbidden_evidence = sorted(EXPECTED_FORBIDDEN_EVIDENCE - set(audit_policy.get("forbidden_evidence") or []))
    require(not missing_forbidden_evidence, f"missing forbidden evidence: {missing_forbidden_evidence}")

    missing_stop_lines = sorted(EXPECTED_STOP_LINES - set(fixture.get("stop_lines") or []))
    require(not missing_stop_lines, f"missing stop lines: {missing_stop_lines}")
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")


def assert_evidence_and_docs(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-workflow-definition-run-record-boundary.py", [])' in check_repo,
        "check-repo.py must run workflow definition run record boundary check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_resource_boundaries(fixture)
    assert_state_machine(fixture)
    assert_failures_audit_and_stop_lines(fixture)
    assert_evidence_and_docs(fixture)
    print("workflow definition run record boundary checks passed.")


if __name__ == "__main__":
    main()
