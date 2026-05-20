#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any
from urllib import request as urlrequest


PLATFORM_LOCAL_SMOKE_ROUTE = "/v1/platform/local-smoke"


def build_offline_platform_local_smoke() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "kind": "platform_local_smoke",
        "stage": "P3 Local Product Shell / Ops Surface",
        "route": PLATFORM_LOCAL_SMOKE_ROUTE,
        "summary": {
            "status": "ok",
            "local_console_ready": True,
            "read_only": True,
            "default_ports": {
                "platform": 7000,
                "console": 4000,
            },
        },
        "checks": {
            "healthz": {
                "route": "/healthz",
                "status": "ok",
                "readable": True,
            },
            "overview": {
                "route": "/v1/platform/overview",
                "readable": True,
                "contract_kind": "platform_overview",
                "ui_consumable": True,
                "product_routes": [
                    "/healthz",
                    "/v1/platform/overview",
                    PLATFORM_LOCAL_SMOKE_ROUTE,
                    "/v1/models",
                    "/v1/models/{id}",
                    "/v1/session/metadata",
                    "/v1/tools/metadata",
                    "/v1/tools/actions",
                ],
            },
            "model_inventory": {
                "route": "/v1/models",
                "detail_route": "/v1/models/{id}",
                "status": "ok",
                "readable": True,
                "inventory_kind": "bridge_backed_provider_profile_inventory",
                "model_count": 3,
                "provider_count": 2,
                "profile_count": 2,
                "failure_code": None,
            },
            "session_tooling": {
                "session_metadata_route": "/v1/session/metadata",
                "tools_metadata_route": "/v1/tools/metadata",
                "blocked_action_route": "/v1/tools/actions",
                "session_metadata_readable": True,
                "tools_metadata_readable": True,
                "tool_count": 2,
                "metadata_only": True,
                "execution_enabled": False,
                "blocked_action_status": "blocked",
                "blocked_action_primary_code": "TOOL_EXECUTOR_DISABLED",
                "requires_confirmation": True,
                "blocked_action_no_side_effects": True,
            },
            "local_console": {
                "frontend_origin_default": "http://127.0.0.1:4000",
                "backend_url_default": "http://127.0.0.1:7000",
                "allowed_cors_origins": ["http://127.0.0.1:4000", "http://localhost:4000"],
                "cors_preflight_methods": ["GET", "POST", "OPTIONS"],
                "cors_scope": "local_dev_only",
            },
        },
        "stop_lines": {
            "real_executor_enabled": False,
            "durable_store_enabled": False,
            "confirmation_flow_connected": False,
            "materialized_result_reader": False,
            "long_term_memory_enabled": False,
            "business_truth_write_enabled": False,
            "automatic_replay_enabled": False,
            "production_secret_backend_ready": False,
        },
        "failure_hints": [
            {
                "code": "PORT_IN_USE",
                "message": "default local ports are platform 7000 and console 4000; release the port or confirm the existing process is the expected RadishMind service",
            },
            {
                "code": "CORS_ORIGIN_NOT_ALLOWED",
                "message": "local console CORS is only allowed for http://127.0.0.1:4000 and http://localhost:4000",
            },
            {
                "code": "ERR_UNSAFE_PORT",
                "message": "some browser ports are blocked by unsafe-port policy; prefer the default local console and platform ports",
            },
        ],
        "audit": {
            "advisory_only": True,
            "writes_business_truth": False,
            "notes": [
                "local smoke summarizes existing read-only platform routes for development-time console readiness",
                "it does not start processes or enable executor, durable storage, confirmation flow, business writeback, or replay",
            ],
        },
    }


def fetch_json(base_url: str, route: str) -> dict[str, Any]:
    url = base_url.rstrip("/") + route
    http_request = urlrequest.Request(url, headers={"Accept": "application/json"}, method="GET")
    with urlrequest.urlopen(http_request, timeout=10) as response:
        return json.loads(response.read().decode("utf-8"))


def stop_lines_enforced(stop_lines: dict[str, Any]) -> bool:
    required_disabled = (
        "real_executor_enabled",
        "durable_store_enabled",
        "confirmation_flow_connected",
        "materialized_result_reader",
        "long_term_memory_enabled",
        "business_truth_write_enabled",
        "automatic_replay_enabled",
    )
    return all(stop_lines.get(key) is False for key in required_disabled)


def build_readiness_view(*, source: str, local_smoke: dict[str, Any]) -> dict[str, Any]:
    summary = local_smoke.get("summary") or {}
    checks = local_smoke.get("checks") or {}
    healthz = checks.get("healthz") or {}
    overview = checks.get("overview") or {}
    model_inventory = checks.get("model_inventory") or {}
    session_tooling = checks.get("session_tooling") or {}
    local_console = checks.get("local_console") or {}
    stop_lines = local_smoke.get("stop_lines") or {}
    allowed_origins = local_console.get("allowed_cors_origins") or []
    return {
        "schema_version": 1,
        "kind": "platform_local_smoke_readiness_view",
        "source": source,
        "route": PLATFORM_LOCAL_SMOKE_ROUTE,
        "view_models": {
            "summary": {
                "status": summary.get("status"),
                "local_console_ready": summary.get("local_console_ready") is True,
                "read_only": summary.get("read_only") is True,
                "platform_port": (summary.get("default_ports") or {}).get("platform"),
                "console_port": (summary.get("default_ports") or {}).get("console"),
            },
            "route_readiness": {
                "healthz_ok": healthz.get("status") == "ok" and healthz.get("readable") is True,
                "overview_contract_readable": overview.get("contract_kind") == "platform_overview"
                and overview.get("readable") is True,
                "model_inventory_readable": model_inventory.get("readable") is True
                and model_inventory.get("status") == "ok",
                "session_tooling_metadata_readable": session_tooling.get("session_metadata_readable") is True
                and session_tooling.get("tools_metadata_readable") is True,
                "blocked_action_no_side_effects": session_tooling.get("blocked_action_status") == "blocked"
                and session_tooling.get("blocked_action_no_side_effects") is True,
            },
            "local_console": {
                "allowed_cors_origins": allowed_origins,
                "cors_scope": local_console.get("cors_scope"),
                "default_frontend_origin": local_console.get("frontend_origin_default"),
                "default_backend_url": local_console.get("backend_url_default"),
                "unsafe_port_hint_present": any(
                    hint.get("code") == "ERR_UNSAFE_PORT" for hint in local_smoke.get("failure_hints") or []
                ),
            },
            "stop_lines": {
                "all_stop_lines_enforced": stop_lines_enforced(stop_lines),
                "can_execute_actions": False,
                "can_use_durable_store": False,
                "can_write_business_truth": False,
                "can_replay_automatically": False,
            },
        },
        "invariants": {
            "local_smoke_route_present": local_smoke.get("route") == PLATFORM_LOCAL_SMOKE_ROUTE,
            "read_only": summary.get("read_only") is True,
            "local_console_ready": summary.get("local_console_ready") is True,
            "healthz_ok": healthz.get("status") == "ok" and healthz.get("readable") is True,
            "overview_contract_readable": overview.get("contract_kind") == "platform_overview"
            and overview.get("readable") is True,
            "model_inventory_readable": model_inventory.get("readable") is True,
            "session_tooling_metadata_readable": session_tooling.get("session_metadata_readable") is True
            and session_tooling.get("tools_metadata_readable") is True,
            "blocked_action_no_side_effects": session_tooling.get("blocked_action_no_side_effects") is True,
            "local_console_cors_allowed": "http://127.0.0.1:4000" in allowed_origins
            and "http://localhost:4000" in allowed_origins,
            "stop_lines_enforced": stop_lines_enforced(stop_lines),
        },
    }


def assert_readiness_view(document: dict[str, Any]) -> None:
    if document.get("kind") != "platform_local_smoke_readiness_view":
        raise SystemExit("local smoke view kind mismatch")
    invariants = document.get("invariants") or {}
    for key in (
        "local_smoke_route_present",
        "read_only",
        "local_console_ready",
        "healthz_ok",
        "overview_contract_readable",
        "model_inventory_readable",
        "session_tooling_metadata_readable",
        "blocked_action_no_side_effects",
        "local_console_cors_allowed",
        "stop_lines_enforced",
    ):
        if invariants.get(key) is not True:
            raise SystemExit(f"platform local smoke invariant failed: {key}")
    view_models = document.get("view_models") or {}
    stop_lines = view_models.get("stop_lines") or {}
    for key in (
        "can_execute_actions",
        "can_use_durable_store",
        "can_write_business_truth",
        "can_replay_automatically",
    ):
        if stop_lines.get(key) is not False:
            raise SystemExit(f"local smoke view incorrectly enables {key}")


def run_smoke(args: argparse.Namespace) -> dict[str, Any]:
    if args.base_url:
        source = args.base_url.rstrip("/")
        local_smoke = fetch_json(source, PLATFORM_LOCAL_SMOKE_ROUTE)
    else:
        source = "offline_fixture"
        local_smoke = build_offline_platform_local_smoke()
    return build_readiness_view(source=source, local_smoke=local_smoke)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", help="running platform base URL; omit to use offline fixture mode")
    parser.add_argument("--output", help="optional JSON output path")
    parser.add_argument("--check", action="store_true", help="validate local smoke readiness invariants")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    document = run_smoke(args)
    if args.check:
        assert_readiness_view(document)
    output = json.dumps(document, ensure_ascii=False, indent=2) + "\n"
    if args.output:
        Path(args.output).write_text(output, encoding="utf-8")
    else:
        sys.stdout.write(output)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
