#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any
from urllib import request as urlrequest


PLATFORM_OVERVIEW_ROUTE = "/v1/platform/overview"


def build_offline_platform_overview() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "kind": "platform_overview",
        "stage": "P3 Local Product Shell",
        "route": PLATFORM_OVERVIEW_ROUTE,
        "service": {
            "name": "radishmind-platform",
            "version": "offline-smoke",
            "status": "ok",
        },
        "product_surface": {
            "mode": "local_read_only_product_shell",
            "implemented": True,
            "ui_consumable": True,
            "routes": [
                "/healthz",
                PLATFORM_OVERVIEW_ROUTE,
                "/v1/models",
                "/v1/models/{id}",
                "/v1/session/metadata",
                "/v1/tools/metadata",
                "/v1/tools/actions",
            ],
        },
        "models": {
            "route": "/v1/models",
            "detail_route": "/v1/models/{id}",
            "inventory_kind": "bridge_backed_provider_profile_inventory",
            "default_provider": "mock",
            "default_profile": "",
            "default_model": "radishmind-local-dev",
            "status": "ok",
            "model_count": 3,
            "provider_count": 2,
            "profile_count": 2,
            "active_profile_chain": [
                "profile:anyrouter",
                "provider:ollama:profile:local",
            ],
            "selectable_model_ids": [
                "radishmind-local-dev",
                "profile:anyrouter",
                "provider:ollama:profile:local",
            ],
        },
        "session_tooling": {
            "stage": "P2 close candidate shell",
            "session_metadata_route": "/v1/session/metadata",
            "tools_metadata_route": "/v1/tools/metadata",
            "blocked_action_route": "/v1/tools/actions",
            "metadata_only": True,
            "tool_count": 2,
            "execution_enabled": False,
            "tool_action_status": "blocked",
            "requires_confirmation_path": "future_upper_layer_confirmation_flow",
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
        "audit": {
            "advisory_only": True,
            "writes_business_truth": False,
            "notes": [
                "overview aggregates existing local product surface metadata for UI or upper-layer discovery",
                "it does not enable executor, durable storage, confirmation flow, long-term memory, business writeback, or replay",
            ],
        },
    }


def fetch_json(base_url: str, route: str) -> dict[str, Any]:
    url = base_url.rstrip("/") + route
    http_request = urlrequest.Request(url, headers={"Accept": "application/json"}, method="GET")
    with urlrequest.urlopen(http_request, timeout=10) as response:
        return json.loads(response.read().decode("utf-8"))


def build_service_status_view(overview: dict[str, Any]) -> dict[str, Any]:
    service = overview.get("service") or {}
    product_surface = overview.get("product_surface") or {}
    return {
        "service_name": service.get("name"),
        "version": service.get("version"),
        "status": service.get("status"),
        "stage": overview.get("stage"),
        "mode": product_surface.get("mode"),
        "overview_route": overview.get("route"),
        "healthy_for_local_console": (
            service.get("status") == "ok"
            and product_surface.get("implemented") is True
            and product_surface.get("ui_consumable") is True
        ),
    }


def build_model_inventory_view(overview: dict[str, Any]) -> dict[str, Any]:
    models = overview.get("models") or {}
    selectable_model_ids = models.get("selectable_model_ids") or []
    return {
        "status": models.get("status"),
        "inventory_kind": models.get("inventory_kind"),
        "models_route": models.get("route"),
        "detail_route": models.get("detail_route"),
        "default_provider": models.get("default_provider"),
        "default_profile": models.get("default_profile"),
        "default_model": models.get("default_model"),
        "model_count": models.get("model_count", 0),
        "provider_count": models.get("provider_count", 0),
        "profile_count": models.get("profile_count", 0),
        "selectable_model_ids": selectable_model_ids,
        "active_profile_chain": models.get("active_profile_chain") or [],
        "can_show_profile_selector": models.get("status") == "ok" and len(selectable_model_ids) > 0,
    }


def build_session_tooling_view(overview: dict[str, Any]) -> dict[str, Any]:
    session_tooling = overview.get("session_tooling") or {}
    return {
        "session_metadata_route": session_tooling.get("session_metadata_route"),
        "tools_metadata_route": session_tooling.get("tools_metadata_route"),
        "blocked_action_route": session_tooling.get("blocked_action_route"),
        "metadata_only": True,
        "execution_enabled": False,
        "action_status_label": "blocked",
        "tool_count": session_tooling.get("tool_count", 0),
        "requires_confirmation_path": session_tooling.get("requires_confirmation_path"),
    }


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


def build_stop_line_view(overview: dict[str, Any]) -> dict[str, Any]:
    stop_lines = overview.get("stop_lines") or {}
    return {
        "all_stop_lines_enforced": stop_lines_enforced(stop_lines),
        "blocked_capability_ids": sorted(
            key for key, enabled in stop_lines.items() if enabled is False
        ),
        "can_execute_actions": False,
        "can_use_durable_store": False,
        "can_write_business_truth": False,
        "can_replay_automatically": False,
    }


def build_console_view(*, source: str, overview: dict[str, Any]) -> dict[str, Any]:
    product_surface = overview.get("product_surface") or {}
    routes = product_surface.get("routes") or []
    service_status = build_service_status_view(overview)
    model_inventory = build_model_inventory_view(overview)
    session_tooling = build_session_tooling_view(overview)
    stop_lines = build_stop_line_view(overview)
    return {
        "schema_version": 1,
        "kind": "platform_overview_console_smoke_view",
        "source": source,
        "routes": {
            "platform_overview": PLATFORM_OVERVIEW_ROUTE,
            "product_surface": routes,
        },
        "view_models": {
            "service_status": service_status,
            "model_inventory": model_inventory,
            "session_tooling": session_tooling,
            "stop_lines": stop_lines,
        },
        "invariants": {
            "overview_route_present": PLATFORM_OVERVIEW_ROUTE in routes,
            "read_only_product_shell": product_surface.get("mode") == "local_read_only_product_shell",
            "ui_consumable": product_surface.get("ui_consumable") is True,
            "model_inventory_consumable": model_inventory.get("status") == "ok",
            "session_tooling_metadata_only": session_tooling.get("metadata_only") is True,
            "blocked_action_route_visible": session_tooling.get("blocked_action_route") == "/v1/tools/actions",
            "stop_lines_enforced": stop_lines.get("all_stop_lines_enforced") is True,
        },
    }


def assert_console_view(document: dict[str, Any]) -> None:
    if document.get("kind") != "platform_overview_console_smoke_view":
        raise SystemExit("console view kind mismatch")
    invariants = document.get("invariants") or {}
    for key in (
        "overview_route_present",
        "read_only_product_shell",
        "ui_consumable",
        "model_inventory_consumable",
        "session_tooling_metadata_only",
        "blocked_action_route_visible",
        "stop_lines_enforced",
    ):
        if invariants.get(key) is not True:
            raise SystemExit(f"platform overview console invariant failed: {key}")
    view_models = document.get("view_models") or {}
    service_status = view_models.get("service_status") or {}
    session_tooling = view_models.get("session_tooling") or {}
    stop_lines = view_models.get("stop_lines") or {}
    if service_status.get("healthy_for_local_console") is not True:
        raise SystemExit("service status is not healthy for local console")
    if session_tooling.get("execution_enabled") is not False:
        raise SystemExit("session tooling view incorrectly enables execution")
    for key in (
        "can_execute_actions",
        "can_use_durable_store",
        "can_write_business_truth",
        "can_replay_automatically",
    ):
        if stop_lines.get(key) is not False:
            raise SystemExit(f"stop-line view incorrectly enables {key}")


def run_smoke(args: argparse.Namespace) -> dict[str, Any]:
    if args.base_url:
        source = args.base_url.rstrip("/")
        overview = fetch_json(source, PLATFORM_OVERVIEW_ROUTE)
    else:
        source = "offline_fixture"
        overview = build_offline_platform_overview()
    return build_console_view(source=source, overview=overview)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", help="running platform base URL; omit to use offline fixture mode")
    parser.add_argument("--output", help="optional JSON output path")
    parser.add_argument("--check", action="store_true", help="validate local console invariants")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    document = run_smoke(args)
    if args.check:
        assert_console_view(document)
    output = json.dumps(document, ensure_ascii=False, indent=2) + "\n"
    if args.output:
        Path(args.output).write_text(output, encoding="utf-8")
    else:
        sys.stdout.write(output)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
