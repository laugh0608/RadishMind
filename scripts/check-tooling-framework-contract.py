#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
TOOL_SCHEMA_PATH = REPO_ROOT / "contracts/tool.schema.json"
TOOL_REGISTRY_SCHEMA_PATH = REPO_ROOT / "contracts/tool-registry.schema.json"
TOOL_AUDIT_SCHEMA_PATH = REPO_ROOT / "contracts/tool-audit-record.schema.json"
TOOL_REGISTRY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/tool-registry-basic.json"
TOOL_AUDIT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/tool-audit-record-basic.json"


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def check_tool_definition(tool: dict[str, Any]) -> None:
    execution = tool.get("execution")
    require(isinstance(execution, dict), "tool execution must be an object")
    require(execution.get("mode") == "contract_only", "v1 fixture must not bind a real tool executor")
    require(execution.get("executor_ref") is None, "contract-only tool must not declare executor_ref")
    retry_policy = execution.get("retry_policy")
    require(isinstance(retry_policy, dict), "tool retry_policy must be an object")
    require(retry_policy.get("max_attempts") == 1, "v1 registry must keep retry attempts at 1")
    require(retry_policy.get("retry_on") == [], "contract-only tools must not retry execution")

    policy = tool.get("policy")
    require(isinstance(policy, dict), "tool policy must be an object")
    require(policy.get("advisory_only") is True, "tool policy must remain advisory-only")
    require(policy.get("writes_business_truth") is False, "tool must not write business truth")
    require(policy.get("allows_network") is False, "v1 tooling registry must not enable network tools")
    require(policy.get("durable_memory_enabled") is False, "tool must not enable durable memory")

    audit = tool.get("audit")
    require(isinstance(audit, dict), "tool audit must be an object")
    require(audit.get("audit_event_required") is True, "tool audit event must be required")
    require(audit.get("redacts_secrets") is True, "tool audit must redact secrets")
    require(isinstance(audit.get("notes"), list) and len(audit["notes"]) > 0, "tool audit must include notes")


def check_registry(registry: dict[str, Any], tool_schema: dict[str, Any]) -> None:
    policy = registry.get("registry_policy")
    require(isinstance(policy, dict), "registry_policy must be an object")
    require(policy.get("execution_enabled") is False, "v1 registry must not enable tool execution")
    require(policy.get("durable_memory_enabled") is False, "v1 registry must not enable durable memory")
    require(policy.get("network_default") == "disabled", "v1 registry network default must be disabled")
    require(policy.get("max_retry_attempts") == 1, "v1 registry must keep max retry attempts at 1")

    tools = registry.get("tools")
    require(isinstance(tools, list) and len(tools) > 0, "tool registry must include tools")
    seen_tool_ids: set[str] = set()
    for tool in tools:
        require(isinstance(tool, dict), "tool registry entry must be an object")
        jsonschema.validate(tool, tool_schema)
        tool_id = str(tool.get("tool_id") or "").strip()
        require(tool_id not in seen_tool_ids, f"duplicate tool_id in registry: {tool_id}")
        seen_tool_ids.add(tool_id)
        check_tool_definition(tool)

    audit = registry.get("audit")
    require(isinstance(audit, dict), "registry audit must be an object")
    require(audit.get("advisory_only") is True, "tool registry must remain advisory-only")
    require(audit.get("writes_business_truth") is False, "tool registry must not write business truth")


def check_tool_audit_record(record: dict[str, Any], registry: dict[str, Any]) -> None:
    tool_ids = {str(tool.get("tool_id")) for tool in registry.get("tools", []) if isinstance(tool, dict)}
    require(record.get("tool_id") in tool_ids, "tool audit record must reference a registered tool")

    decision = record.get("policy_decision")
    require(isinstance(decision, dict), "policy_decision must be an object")
    require(
        decision.get("decision") == "blocked_tool_execution_disabled",
        "v1 fixture must show execution blocked by registry policy",
    )
    require(decision.get("requires_confirmation") is True, "candidate action tool audit must require confirmation")

    execution = record.get("execution")
    require(isinstance(execution, dict), "execution must be an object")
    require(execution.get("execution_enabled") is False, "tool audit must keep execution disabled")
    require(execution.get("status") == "not_executed", "v1 tool audit must not claim execution")
    require(execution.get("executor_ref") is None, "v1 tool audit must not reference an executor")

    audit = record.get("audit")
    require(isinstance(audit, dict), "audit must be an object")
    require(audit.get("advisory_only") is True, "tool audit must remain advisory-only")
    require(audit.get("writes_business_truth") is False, "tool audit must not write business truth")
    require(audit.get("durable_memory_written") is False, "tool audit must not write durable memory")
    require(audit.get("secrets_redacted") is True, "tool audit must redact secrets")


def main() -> int:
    tool_schema = load_json_document(TOOL_SCHEMA_PATH)
    registry_schema = load_json_document(TOOL_REGISTRY_SCHEMA_PATH)
    audit_schema = load_json_document(TOOL_AUDIT_SCHEMA_PATH)
    registry_fixture = load_json_document(TOOL_REGISTRY_FIXTURE_PATH)
    audit_fixture = load_json_document(TOOL_AUDIT_FIXTURE_PATH)

    for schema in (tool_schema, registry_schema, audit_schema):
        jsonschema.Draft202012Validator.check_schema(schema)

    jsonschema.validate(registry_fixture, registry_schema)
    jsonschema.validate(audit_fixture, audit_schema)
    if not isinstance(registry_fixture, dict):
        raise SystemExit("tool registry fixture must be an object")
    if not isinstance(audit_fixture, dict):
        raise SystemExit("tool audit fixture must be an object")

    check_registry(registry_fixture, tool_schema)
    check_tool_audit_record(audit_fixture, registry_fixture)

    print("tooling framework contract smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
