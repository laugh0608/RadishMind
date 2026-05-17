#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any
from urllib import request as urlrequest


REPO_ROOT = Path(__file__).resolve().parent.parent

SESSION_METADATA_ROUTE = "/v1/session/metadata"
TOOLS_METADATA_ROUTE = "/v1/tools/metadata"
TOOL_ACTIONS_ROUTE = "/v1/tools/actions"
DEFAULT_TOOL_ID = "radishflow.suggest_edits.candidate_builder.v1"


def build_offline_session_metadata() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "kind": "session_metadata",
        "stage": "P2 Session & Tooling Foundation",
        "route": SESSION_METADATA_ROUTE,
        "state_policy": {
            "session_state_scope": "northbound_metadata",
            "tool_result_cache_scope": "metadata_only",
            "recovery_checkpoint_scope": "audit_refs_only",
            "durable_memory_enabled": False,
        },
        "capabilities": {
            "session_metadata": True,
            "durable_session_store": False,
            "durable_checkpoint_store": False,
            "long_term_memory": False,
            "automatic_replay": False,
            "business_truth_write": False,
        },
    }


def build_offline_tools_metadata() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "kind": "tooling_metadata",
        "registry_id": "radishmind-tool-registry-v1",
        "route": TOOLS_METADATA_ROUTE,
        "registry_policy": {
            "execution_enabled": False,
            "durable_memory_enabled": False,
            "network_default": "disabled",
            "default_timeout_seconds": 30,
            "max_retry_attempts": 1,
        },
        "tools": [
            {
                "tool_id": "radish.docs.retrieval_context.v1",
                "display_name": "Radish Docs Retrieval Context",
                "tool_type": "retrieval",
                "project_scope": "radish",
                "risk_level": "low",
                "requires_confirmation_for_actions": False,
                "execution": {
                    "mode": "contract_only",
                    "execution_enabled": False,
                    "status": "disabled",
                },
                "state_policy": {
                    "result_cache_mode": "metadata_only",
                    "durable_memory_enabled": False,
                    "writes_business_truth": False,
                    "materialized_result_read": False,
                },
            },
            {
                "tool_id": DEFAULT_TOOL_ID,
                "display_name": "RadishFlow Suggest Edits Candidate Builder",
                "tool_type": "candidate_builder",
                "project_scope": "radishflow",
                "risk_level": "medium",
                "requires_confirmation_for_actions": True,
                "execution": {
                    "mode": "contract_only",
                    "execution_enabled": False,
                    "status": "disabled",
                },
                "state_policy": {
                    "result_cache_mode": "metadata_only",
                    "durable_memory_enabled": False,
                    "writes_business_truth": False,
                    "materialized_result_read": False,
                },
            },
        ],
        "blocked_action_route": TOOL_ACTIONS_ROUTE,
    }


def build_offline_blocked_action(tool_id: str, session_id: str, turn_id: str) -> dict[str, Any]:
    return {
        "schema_version": 1,
        "kind": "tool_action_blocked_response",
        "status": "blocked",
        "route": TOOL_ACTIONS_ROUTE,
        "request_id": "req-session-tooling-consumer-smoke",
        "action": {
            "tool_id": tool_id,
            "action": "execute",
            "known_tool": True,
            "session_id": session_id,
            "turn_id": turn_id,
        },
        "policy_decision": {
            "decision": "blocked",
            "primary_code": "TOOL_EXECUTOR_DISABLED",
            "denial_codes": ["TOOL_EXECUTOR_DISABLED", "CONFIRMATION_REQUIRED"],
            "requires_confirmation": True,
            "reason": "tool execution is disabled and this tool requires upper-layer confirmation before any future execution",
        },
        "execution": {
            "execution_enabled": False,
            "executed": False,
            "status": "not_executed",
            "duration_ms": None,
        },
        "result": {
            "result_ref": None,
            "materialized_result_included": False,
            "materialized_result_read": False,
            "result_cache_mode": "metadata_only",
        },
        "side_effects": {
            "network_request_sent": False,
            "durable_memory_written": False,
            "writes_business_truth": False,
            "automatic_replay_started": False,
        },
    }


def fetch_json(base_url: str, method: str, route: str, body: dict[str, Any] | None = None) -> dict[str, Any]:
    url = base_url.rstrip("/") + route
    raw_body = None
    headers = {"Accept": "application/json"}
    if body is not None:
        raw_body = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"
    request = urlrequest.Request(url, data=raw_body, headers=headers, method=method)
    with urlrequest.urlopen(request, timeout=10) as response:
        return json.loads(response.read().decode("utf-8"))


def build_action_request(tool_id: str, session_id: str, turn_id: str) -> dict[str, Any]:
    return {
        "tool_id": tool_id,
        "action": "execute",
        "session_id": session_id,
        "turn_id": turn_id,
    }


def build_blocked_action_view(response: dict[str, Any]) -> dict[str, Any]:
    action = response.get("action") or {}
    policy_decision = response.get("policy_decision") or {}
    execution = response.get("execution") or {}
    result = response.get("result") or {}
    side_effects = response.get("side_effects") or {}
    return {
        "tool_id": action.get("tool_id"),
        "action": action.get("action"),
        "can_execute": False,
        "status_label": "blocked",
        "primary_code": policy_decision.get("primary_code"),
        "requires_confirmation": policy_decision.get("requires_confirmation") is True,
        "no_side_effects": (
            execution.get("executed") is False
            and result.get("result_ref") is None
            and side_effects.get("network_request_sent") is False
            and side_effects.get("durable_memory_written") is False
            and side_effects.get("writes_business_truth") is False
            and side_effects.get("automatic_replay_started") is False
        ),
    }


def build_consumer_view(
    *,
    source: str,
    session_metadata: dict[str, Any],
    tools_metadata: dict[str, Any],
    blocked_action: dict[str, Any],
) -> dict[str, Any]:
    disabled_capabilities = [
        key
        for key, enabled in sorted((session_metadata.get("capabilities") or {}).items())
        if enabled is False
    ]
    tool_views = []
    for tool in tools_metadata.get("tools") or []:
        execution = tool.get("execution") or {}
        tool_views.append(
            {
                "tool_id": tool.get("tool_id"),
                "display_name": tool.get("display_name"),
                "project_scope": tool.get("project_scope"),
                "risk_level": tool.get("risk_level"),
                "requires_confirmation": tool.get("requires_confirmation_for_actions") is True,
                "can_request_action": True,
                "execution_enabled": execution.get("execution_enabled") is True,
                "execution_mode": execution.get("mode"),
            }
        )

    blocked_view = build_blocked_action_view(blocked_action)
    return {
        "schema_version": 1,
        "kind": "session_tooling_consumer_smoke_view",
        "source": source,
        "routes": {
            "session_metadata": SESSION_METADATA_ROUTE,
            "tools_metadata": TOOLS_METADATA_ROUTE,
            "tool_actions": TOOL_ACTIONS_ROUTE,
        },
        "session": {
            "state_scope": (session_metadata.get("state_policy") or {}).get("session_state_scope"),
            "disabled_capabilities": disabled_capabilities,
        },
        "tools": tool_views,
        "blocked_action": blocked_view,
        "invariants": {
            "metadata_only": True,
            "all_tools_execution_disabled": all(tool.get("execution_enabled") is False for tool in tool_views),
            "blocked_action_cannot_execute": blocked_view.get("can_execute") is False,
            "blocked_action_has_no_side_effects": blocked_view.get("no_side_effects") is True,
        },
    }


def assert_consumer_view(document: dict[str, Any]) -> None:
    if document.get("kind") != "session_tooling_consumer_smoke_view":
        raise SystemExit("consumer view kind mismatch")
    invariants = document.get("invariants") or {}
    for key in (
        "metadata_only",
        "all_tools_execution_disabled",
        "blocked_action_cannot_execute",
        "blocked_action_has_no_side_effects",
    ):
        if invariants.get(key) is not True:
            raise SystemExit(f"consumer view invariant failed: {key}")
    blocked_action = document.get("blocked_action") or {}
    if blocked_action.get("status_label") != "blocked" or blocked_action.get("can_execute") is not False:
        raise SystemExit("blocked action view is not safe to display")


def run_smoke(args: argparse.Namespace) -> dict[str, Any]:
    if args.base_url:
        source = args.base_url.rstrip("/")
        session_metadata = fetch_json(source, "GET", SESSION_METADATA_ROUTE)
        tools_metadata = fetch_json(source, "GET", TOOLS_METADATA_ROUTE)
        blocked_action = fetch_json(
            source,
            "POST",
            TOOL_ACTIONS_ROUTE,
            build_action_request(args.tool_id, args.session_id, args.turn_id),
        )
    else:
        source = "offline_fixture"
        session_metadata = build_offline_session_metadata()
        tools_metadata = build_offline_tools_metadata()
        blocked_action = build_offline_blocked_action(args.tool_id, args.session_id, args.turn_id)

    return build_consumer_view(
        source=source,
        session_metadata=session_metadata,
        tools_metadata=tools_metadata,
        blocked_action=blocked_action,
    )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", help="running platform base URL; omit to use offline fixture mode")
    parser.add_argument("--tool-id", default=DEFAULT_TOOL_ID)
    parser.add_argument("--session-id", default="radishflow-session-001")
    parser.add_argument("--turn-id", default="turn-0003")
    parser.add_argument("--output", help="optional JSON output path")
    parser.add_argument("--check", action="store_true", help="validate safe consumer invariants")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    document = run_smoke(args)
    if args.check:
        assert_consumer_view(document)
    output = json.dumps(document, ensure_ascii=False, indent=2) + "\n"
    if args.output:
        Path(args.output).write_text(output, encoding="utf-8")
    else:
        sys.stdout.write(output)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
