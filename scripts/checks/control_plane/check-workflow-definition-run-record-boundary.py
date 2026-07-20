#!/usr/bin/env python3
from __future__ import annotations

import copy
import hashlib
import json
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-definition-run-record-boundary.json"
DEPENDENCY_FIXTURES = {
    "product-surface-v1-boundary": REPO_ROOT / "scripts/checks/fixtures/product-surface-v1-boundary.json",
    "control-plane-data-boundary": REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json",
    "radish-oidc-client-preconditions": REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json",
    "gateway-api-key-quota-readiness": REPO_ROOT / "scripts/checks/fixtures/gateway-api-key-quota-readiness.json",
}
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
CONTRACTS_README_PATH = REPO_ROOT / "contracts/README.md"
HTTP_TOOL_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-http-tool-contracts-v1.json"
HTTP_TOOL_SCHEMA_PATHS = {
    "definition": REPO_ROOT / "contracts/workflow-http-tool-definition.schema.json",
    "execution_profile": REPO_ROOT / "contracts/workflow-http-tool-execution-profile.schema.json",
    "action_plan": REPO_ROOT / "contracts/workflow-http-tool-action-plan.schema.json",
    "confirmation_decision": REPO_ROOT / "contracts/workflow-http-tool-confirmation-decision.schema.json",
    "execution_audit": REPO_ROOT / "contracts/workflow-http-tool-execution-audit.schema.json",
    "run_record_v2": REPO_ROOT / "contracts/workflow-run-record-v2.schema.json",
}
EXPECTED_HTTP_TOOL_ID = "workflow.http.reviewed-json-read.v1"
EXPECTED_ACTION_PLAN_FIELDS = {
    "schema_version",
    "plan_id",
    "record_version",
    "tenant_ref",
    "workspace_id",
    "application_id",
    "draft_id",
    "draft_version",
    "node_id",
    "tool_id",
    "tool_version",
    "definition_digest",
    "profile_id",
    "profile_version",
    "profile_digest",
    "method",
    "target_policy_key",
    "public_arguments",
    "output_fields",
    "output_schema_digest",
    "credential_policy",
    "timeout_ms",
    "max_response_bytes",
    "max_output_bytes",
    "planned_by_actor_ref",
    "created_at",
    "expires_at",
    "tool_plan_digest",
    "status",
    "last_decision_by_actor_ref",
    "last_decision_at",
    "audit_ref",
}
EXPECTED_CONFIRMATION_DECISION_FIELDS = {
    "schema_version",
    "confirmation_id",
    "plan_id",
    "tenant_ref",
    "workspace_id",
    "application_id",
    "draft_id",
    "draft_version",
    "node_id",
    "tool_id",
    "tool_version",
    "tool_plan_digest",
    "outcome",
    "decided_by_actor_ref",
    "actor_source",
    "decided_at",
    "reason_code",
    "expected_record_version",
    "resulting_record_version",
    "audit_ref",
}
EXPECTED_NEGATIVE_MUTATIONS = {
    "definition_unversioned_tool_id",
    "definition_confirmation_false",
    "definition_endpoint_added",
    "definition_output_schema_field_added",
    "profile_method_post",
    "profile_definition_digest_drift",
    "profile_credential_bearer",
    "profile_redirect_enabled",
    "action_plan_url_argument",
    "action_plan_url_resource",
    "action_plan_endpoint_added",
    "action_plan_definition_digest_drift",
    "action_plan_profile_digest_drift",
    "action_plan_output_schema_digest_drift",
    "action_plan_tool_plan_digest_drift",
    "action_plan_tool_version_drift",
    "decision_run_id_added",
    "decision_human_expire",
    "decision_resulting_version_removed",
    "decision_tool_version_removed",
    "audit_endpoint_added",
    "audit_business_write",
    "audit_raw_response_added",
    "audit_tool_version_drift",
    "run_zero_tool_call",
    "run_zero_confirmation_call",
    "run_business_write",
    "run_endpoint_projection",
}

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


def is_rfc3339_datetime(value: object) -> bool:
    if not isinstance(value, str) or not value:
        return False
    normalized = value[:-1] + "+00:00" if value.endswith("Z") else value
    parsed = datetime.fromisoformat(normalized)
    return parsed.tzinfo is not None


WORKFLOW_FORMAT_CHECKER = jsonschema.FormatChecker()
WORKFLOW_FORMAT_CHECKER.checks("date-time", raises=(ValueError, TypeError))(is_rfc3339_datetime)


def restricted_jcs_sha256(value: dict[str, Any]) -> str:
    canonical = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"), allow_nan=False)
    return "sha256:" + hashlib.sha256(canonical.encode("utf-8")).hexdigest()


def output_schema_binding_is_valid(definition: dict[str, Any], action_plan: dict[str, Any]) -> bool:
    output_schema = definition.get("output_schema")
    return isinstance(output_schema, dict) and action_plan.get("output_schema_digest") == restricted_jcs_sha256(output_schema)


def definition_profile_binding_is_valid(definition: dict[str, Any], profile: dict[str, Any]) -> bool:
    return profile.get("definition_digest") == restricted_jcs_sha256(definition)


def action_plan_binding_is_valid(
    definition: dict[str, Any],
    profile: dict[str, Any],
    action_plan: dict[str, Any],
) -> bool:
    digest_fields = (
        "schema_version", "tenant_ref", "workspace_id", "application_id", "draft_id", "draft_version",
        "node_id", "tool_id", "tool_version", "definition_digest", "profile_id", "profile_version", "profile_digest",
        "method", "target_policy_key", "public_arguments", "output_fields", "output_schema_digest",
        "credential_policy", "timeout_ms", "max_response_bytes", "max_output_bytes", "planned_by_actor_ref", "created_at", "expires_at",
    )
    if not definition_profile_binding_is_valid(definition, profile):
        return False
    if action_plan.get("definition_digest") != profile.get("definition_digest"):
        return False
    if action_plan.get("profile_digest") != restricted_jcs_sha256(profile):
        return False
    if not output_schema_binding_is_valid(definition, action_plan):
        return False
    payload = {field: action_plan.get(field) for field in digest_fields}
    return action_plan.get("tool_plan_digest") == restricted_jcs_sha256(payload)


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
        if resource_id == "confirmation-decision":
            continue
        require(resource.get("status") == "defined_not_implemented", f"{resource_id} must remain defined_not_implemented")
        require(str(resource.get("boundary") or "").strip(), f"{resource_id}.boundary is required")
        required_fields = resource.get("required_fields")
        require(isinstance(required_fields, list) and required_fields, f"{resource_id}.required_fields must be non-empty")

    legacy_confirmation = resources["confirmation-decision"]
    require(legacy_confirmation.get("status") == "superseded_archived", "legacy confirmation must be archived")
    require(legacy_confirmation.get("historical_read_allowed") is True, "legacy confirmation history must stay readable")
    require(legacy_confirmation.get("new_submission_allowed") is False, "legacy confirmation must reject new submissions")
    require(legacy_confirmation.get("runtime_authoritative") is False, "legacy confirmation must not be authoritative")
    require(
        set(legacy_confirmation.get("superseded_by") or [])
        == {"workflow_http_tool_action_plan.v1", "workflow_http_tool_confirmation_decision.v1"},
        "legacy confirmation supersession targets drifted",
    )
    require(str(legacy_confirmation.get("boundary") or "").strip(), "legacy confirmation boundary is required")
    require(legacy_confirmation.get("required_fields"), "legacy confirmation fields must remain for historical reads")

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
    require(state_machine.get("status") == "superseded_archived", "legacy run-bound state machine must be archived")
    require(state_machine.get("historical_read_allowed") is True, "legacy run-bound states must stay readable")
    require(
        state_machine.get("new_transition_submission_allowed") is False,
        "legacy run-bound transitions must reject new submissions",
    )
    require(state_machine.get("runtime_authoritative") is False, "legacy run-bound state machine must not be authoritative")
    require(
        state_machine.get("superseded_by") == "workflow_http_tool_action_plan.v1",
        "legacy run-bound state machine supersession drifted",
    )
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


def mutate_negative_document(document: dict[str, Any], mutation: str) -> dict[str, Any]:
    candidate = copy.deepcopy(document)
    if mutation == "definition_unversioned_tool_id":
        candidate["tool_id"] = "workflow.http.reviewed-json-read"
    elif mutation == "definition_confirmation_false":
        candidate["requires_confirmation"] = False
    elif mutation == "definition_endpoint_added":
        candidate["endpoint"] = "https://forbidden.example.invalid"
    elif mutation == "definition_output_schema_field_added":
        candidate["output_schema"]["properties"]["endpoint"] = {"type": "string"}
    elif mutation == "profile_method_post":
        candidate["method"] = "POST"
    elif mutation == "profile_definition_digest_drift":
        candidate["definition_digest"] = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
    elif mutation == "profile_credential_bearer":
        candidate["credential_policy"] = "bearer"
    elif mutation == "profile_redirect_enabled":
        candidate["network_policy"]["follow_redirects"] = True
    elif mutation == "action_plan_url_argument":
        candidate["public_arguments"]["url"] = "https://forbidden.example.invalid"
    elif mutation == "action_plan_url_resource":
        candidate["public_arguments"]["resource_key"] = "https://forbidden.example.invalid/resource"
    elif mutation == "action_plan_endpoint_added":
        candidate["endpoint"] = "https://forbidden.example.invalid"
    elif mutation == "action_plan_definition_digest_drift":
        candidate["definition_digest"] = "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
    elif mutation == "action_plan_profile_digest_drift":
        candidate["profile_digest"] = "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
    elif mutation == "action_plan_output_schema_digest_drift":
        candidate["output_schema_digest"] = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
    elif mutation == "action_plan_tool_plan_digest_drift":
        candidate["tool_plan_digest"] = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
    elif mutation == "action_plan_tool_version_drift":
        candidate["tool_version"] = 2
    elif mutation == "decision_run_id_added":
        candidate["run_id"] = "run_0123456789abcdef"
    elif mutation == "decision_human_expire":
        candidate["outcome"] = "expire"
    elif mutation == "decision_resulting_version_removed":
        del candidate["resulting_record_version"]
    elif mutation == "decision_tool_version_removed":
        del candidate["tool_version"]
    elif mutation == "audit_endpoint_added":
        candidate["endpoint"] = "https://forbidden.example.invalid"
    elif mutation == "audit_business_write":
        candidate["side_effects"]["business_writes"] = 1
    elif mutation == "audit_raw_response_added":
        candidate["raw_response"] = {"secret": "forbidden"}
    elif mutation == "audit_tool_version_drift":
        candidate["tool_version"] = 2
    elif mutation == "run_zero_tool_call":
        candidate["side_effects"]["tool_calls"] = 0
    elif mutation == "run_zero_confirmation_call":
        candidate["side_effects"]["confirmation_calls"] = 0
    elif mutation == "run_business_write":
        candidate["side_effects"]["business_writes"] = 1
    elif mutation == "run_endpoint_projection":
        candidate["tool_attempt"]["output_projection"]["endpoint"] = "https://forbidden.example.invalid"
    else:
        raise SystemExit(f"unknown negative mutation: {mutation}")
    return candidate


def assert_http_tool_contract_schemas() -> None:
    contract_fixture = load_json(HTTP_TOOL_FIXTURE_PATH)
    require(contract_fixture.get("schema_version") == 1, "unexpected HTTP Tool fixture schema version")
    require(contract_fixture.get("kind") == "workflow_http_tool_contract_fixture_v1", "unexpected HTTP Tool fixture kind")
    require(contract_fixture.get("tool_id") == EXPECTED_HTTP_TOOL_ID, "HTTP Tool fixture id drifted")

    positives = contract_fixture.get("positive") or {}
    require(set(positives) == set(HTTP_TOOL_SCHEMA_PATHS), "HTTP Tool positive fixture catalog drifted")
    validators: dict[str, jsonschema.Draft202012Validator] = {}
    schemas: dict[str, dict[str, Any]] = {}
    contracts_readme = CONTRACTS_README_PATH.read_text(encoding="utf-8")
    for schema_key, schema_path in HTTP_TOOL_SCHEMA_PATHS.items():
        schema = load_json(schema_path)
        jsonschema.Draft202012Validator.check_schema(schema)
        validator = jsonschema.Draft202012Validator(schema, format_checker=WORKFLOW_FORMAT_CHECKER)
        errors = sorted(validator.iter_errors(positives[schema_key]), key=lambda error: list(error.absolute_path))
        require(not errors, f"{schema_key} positive fixture failed: {[error.message for error in errors]}")
        require(
            set(schema.get("properties") or {}) == set(positives[schema_key]),
            f"{schema_key} top-level properties drifted from its positive contract fixture",
        )
        require(
            set(schema.get("required") or []) == set(positives[schema_key]),
            f"{schema_key} required fields drifted from its positive contract fixture",
        )
        validators[schema_key] = validator
        schemas[schema_key] = schema
        require(schema_path.name in contracts_readme, f"contracts/README.md missing {schema_path.name}")

    definition = positives["definition"]
    require(definition.get("tool_id") == EXPECTED_HTTP_TOOL_ID, "first HTTP Tool id drifted")
    argument_schema = definition.get("public_arguments_schema") or {}
    require(argument_schema.get("required") == ["resource_key"], "resource_key must be the sole required argument")
    require(
        set((argument_schema.get("properties") or {}).keys()) == {"resource_key", "locale"},
        "public argument definition must only expose resource_key and locale",
    )
    require(
        (argument_schema.get("properties") or {}).get("resource_key", {}).get("pattern")
        == r"^(?![A-Za-z][A-Za-z0-9+.-]*://)[A-Za-z0-9][A-Za-z0-9._:/-]*$",
        "definition resource_key pattern must reject absolute URLs",
    )
    output_schema = definition.get("output_schema") or {}
    require(output_schema.get("type") == "object", "output schema must describe an object")
    require(output_schema.get("additionalProperties") is False, "output schema must reject additional properties")
    expected_output_fields = ["resource_key", "title", "summary", "updated_at"]
    require(output_schema.get("required") == expected_output_fields, "output schema required fields drifted")
    require(set((output_schema.get("properties") or {}).keys()) == set(expected_output_fields), "output schema fields drifted")
    require(definition.get("output_fields") == expected_output_fields, "definition output fields drifted")
    output_validator = jsonschema.Draft202012Validator(output_schema, format_checker=WORKFLOW_FORMAT_CHECKER)
    require(output_validator.is_valid({
        "resource_key": "docs.release_notes",
        "title": "Release notes",
        "summary": "Sanitized reviewed output.",
        "updated_at": "2026-07-16T02:00:00Z",
    }), "sanitized output positive fixture must validate")
    for invalid_output in (
        {
            "resource_key": "docs.release_notes",
            "title": "Release notes",
            "summary": "Sanitized reviewed output.",
            "updated_at": "not-a-date-time",
        },
        {
            "resource_key": "docs.release_notes",
            "title": "Release notes",
            "summary": "Sanitized reviewed output.",
            "updated_at": "2026-07-16T02:00:00Z",
            "endpoint": "https://forbidden.example.invalid",
        },
    ):
        require(not output_validator.is_valid(invalid_output), "sanitized output negative fixture unexpectedly validated")
    execution_profile = positives["execution_profile"]
    require(execution_profile.get("tool_id") == EXPECTED_HTTP_TOOL_ID, "execution profile tool id drifted")
    require(
        execution_profile.get("profile_id") == "workflow_http_profile_reviewed_json_read_dev_v1",
        "execution profile id drifted",
    )
    require(execution_profile.get("environment") == "development", "execution profile environment drifted")
    require(execution_profile.get("test_only_loopback") is False, "development profile must not allow loopback")
    require(definition_profile_binding_is_valid(definition, execution_profile), "definition digest binding drifted")
    action_plan = positives["action_plan"]
    require(set(action_plan) == EXPECTED_ACTION_PLAN_FIELDS, "action plan projection fields drifted")
    require(set(schemas["action_plan"].get("properties") or {}) == EXPECTED_ACTION_PLAN_FIELDS, "action plan schema properties drifted")
    require(set(schemas["action_plan"].get("required") or []) == EXPECTED_ACTION_PLAN_FIELDS, "action plan schema required fields drifted")
    require(set((action_plan.get("public_arguments") or {}).keys()) <= {"resource_key", "locale"}, "action plan arguments drifted")
    require(action_plan.get("method") == "GET" and action_plan.get("credential_policy") == "none", "plan safety policy drifted")
    require(action_plan.get("tool_version") == definition.get("tool_version") == 1, "plan tool version binding drifted")
    require(action_plan.get("profile_id") == execution_profile.get("profile_id"), "plan profile id binding drifted")
    require(action_plan.get("profile_version") == execution_profile.get("profile_version"), "plan profile version binding drifted")
    require(action_plan.get("output_fields") == definition.get("output_fields"), "plan output field projection drifted")
    require(output_schema_binding_is_valid(definition, action_plan), "action plan output schema digest binding drifted")
    require(action_plan_binding_is_valid(definition, execution_profile, action_plan), "action plan digest chain drifted")
    created_at = datetime.fromisoformat(str(action_plan.get("created_at") or "").replace("Z", "+00:00"))
    expires_at = datetime.fromisoformat(str(action_plan.get("expires_at") or "").replace("Z", "+00:00"))
    require(created_at.tzinfo is not None and created_at.utcoffset() == timezone.utc.utcoffset(created_at), "plan created_at must be UTC")
    require(expires_at.tzinfo is not None and expires_at.utcoffset() == timezone.utc.utcoffset(expires_at), "plan expires_at must be UTC")
    require(int((expires_at - created_at).total_seconds()) == 15 * 60, "plan expiry must remain fifteen minutes")
    require(
        not ({"scheme", "host", "port", "path", "endpoint", "headers", "raw_query"} & set(action_plan)),
        "action plan must not expose runtime target details",
    )
    decision = positives["confirmation_decision"]
    require(set(decision) == EXPECTED_CONFIRMATION_DECISION_FIELDS, "confirmation decision fields drifted")
    require(
        set(schemas["confirmation_decision"].get("properties") or {}) == EXPECTED_CONFIRMATION_DECISION_FIELDS,
        "confirmation decision schema properties drifted",
    )
    require(
        set(schemas["confirmation_decision"].get("required") or []) == EXPECTED_CONFIRMATION_DECISION_FIELDS,
        "confirmation decision schema required fields drifted",
    )
    require("run_id" not in decision, "pre-run confirmation must not bind to a run")
    require(decision.get("tool_version") == action_plan.get("tool_version"), "decision tool version binding drifted")
    require(
        decision.get("resulting_record_version") == decision.get("expected_record_version") + 1,
        "confirmation decision fixture must prove a single CAS increment",
    )

    negative_cases = contract_fixture.get("negative_cases") or []
    mutations = {str(case.get("mutation") or "") for case in negative_cases if isinstance(case, dict)}
    require(mutations == EXPECTED_NEGATIVE_MUTATIONS, "HTTP Tool negative mutation coverage drifted")
    case_ids: set[str] = set()
    for case in negative_cases:
        case_id = str(case.get("id") or "")
        schema_key = str(case.get("schema") or "")
        mutation = str(case.get("mutation") or "")
        require(case_id and case_id not in case_ids, f"duplicate or empty negative case id: {case_id}")
        require(schema_key in validators, f"negative case {case_id} has unknown schema")
        case_ids.add(case_id)
        candidate = mutate_negative_document(positives[schema_key], mutation)
        cross_contract_mutations = {
            "profile_definition_digest_drift",
            "action_plan_definition_digest_drift",
            "action_plan_profile_digest_drift",
            "action_plan_output_schema_digest_drift",
            "action_plan_tool_plan_digest_drift",
        }
        if mutation in cross_contract_mutations:
            require(validators[schema_key].is_valid(candidate), "digest drift must remain shape-valid")
            if schema_key == "execution_profile":
                require(not definition_profile_binding_is_valid(definition, candidate), "definition digest drift must fail binding")
            else:
                require(not action_plan_binding_is_valid(definition, execution_profile, candidate), "plan digest drift must fail binding")
            continue
        require(
            not validators[schema_key].is_valid(candidate),
            f"negative case {case_id} unexpectedly passed {schema_key}",
        )

    for schema_key in ("definition", "action_plan", "confirmation_decision", "execution_audit", "run_record_v2"):
        serialized = json.dumps(positives[schema_key], ensure_ascii=False).lower()
        for forbidden_literal in ("authorization_header", "raw_query", "raw_request", "raw_response", "secret_value"):
            require(forbidden_literal not in serialized, f"{schema_key} leaked forbidden literal {forbidden_literal}")
    require(positives["action_plan"].get("status") == "pending", "batch A positive action plan must remain pending")
    require(positives["execution_audit"].get("tool_version") == action_plan.get("tool_version"), "audit tool version binding drifted")
    require(positives["execution_audit"]["side_effects"] == {
        "network_attempts": 0,
        "tool_calls": 0,
        "provider_calls": 0,
        "business_writes": 0,
        "replay_writes": 0,
    }, "batch A confirmation audit must remain zero-side-effect")
    outcome_unknown = copy.deepcopy(positives["run_record_v2"])
    outcome_unknown["status"] = "outcome_unknown"
    outcome_unknown["failure_code"] = "workflow_tool_outcome_unknown"
    outcome_unknown["failure_summary"] = "The remote outcome cannot be determined and will not be retried."
    outcome_unknown["completed_at"] = "2026-07-16T02:06:30Z"
    outcome_unknown["tool_attempt"]["status"] = "outcome_unknown"
    outcome_unknown["tool_attempt"]["completed_at"] = "2026-07-16T02:06:30Z"
    outcome_unknown["tool_attempt"]["failure_code"] = "workflow_tool_outcome_unknown"
    outcome_unknown["diagnostic"]["failure_boundary"] = "tool_transport"
    outcome_unknown["diagnostic"]["failure_stage"] = "http_attempt"
    outcome_unknown["diagnostic"]["failed_node_id"] = "node_http_tool"
    outcome_unknown["diagnostic"]["tool_failure_category"] = "outcome_unknown"
    outcome_unknown["diagnostic"]["summary"] = "Remote outcome unknown; automatic retry is forbidden."
    outcome_unknown["diagnostic"]["recommended_review_action"] = "review_tool_attempt"
    require(validators["run_record_v2"].is_valid(outcome_unknown), "run v2 outcome_unknown fixture must validate")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_resource_boundaries(fixture)
    assert_state_machine(fixture)
    assert_failures_audit_and_stop_lines(fixture)
    assert_evidence_and_docs(fixture)
    assert_http_tool_contract_schemas()
    print("workflow definition run record boundary checks passed.")


if __name__ == "__main__":
    main()
