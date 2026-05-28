#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
ROUTE_CONTRACT_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
RESPONSE_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
NEGATIVE_CONTRACT_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"

EXPECTED_ROUTE_IDS = {
    "tenant-summary-route",
    "application-summary-list-route",
    "api-key-summary-list-route",
    "quota-summary-route",
    "workflow-definition-summary-list-route",
    "run-record-summary-list-route",
    "audit-summary-list-route",
}


def load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def route_contracts_by_id() -> dict[str, dict[str, Any]]:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE)
    return {
        str(route.get("id") or ""): route
        for route in route_contract.get("route_contracts") or []
        if isinstance(route, dict)
    }


def response_examples_by_route_id() -> dict[str, dict[str, Any]]:
    response_fixture = load_json(RESPONSE_FIXTURE)
    return {
        str(example.get("route_id") or ""): example
        for example in response_fixture.get("response_examples") or []
        if isinstance(example, dict)
    }


def forbidden_output_keys() -> set[str]:
    response_fixture = load_json(RESPONSE_FIXTURE)
    return set(response_fixture.get("forbidden_output_keys") or [])


def build_route_catalog_view(routes: dict[str, dict[str, Any]]) -> dict[str, Any]:
    catalog_items = []
    for route_id in sorted(routes):
        route = routes[route_id]
        catalog_items.append(
            {
                "route_id": route_id,
                "method": route.get("method"),
                "path": route.get("path"),
                "read_model": route.get("read_model"),
                "required_scope": route.get("required_scope"),
                "pagination": route.get("pagination"),
                "allowed_filters": route.get("allowed_filters") or [],
                "tenant_scoped": route.get("tenant_scoped") is True,
                "auth_required": True,
                "fake_store_backed": True,
                "database_backed": False,
                "can_mutate": False,
            }
        )
    return {
        "routes": catalog_items,
        "all_routes_fake_store_backed": True,
        "all_routes_require_auth": True,
        "all_routes_read_only": True,
        "database_backed": False,
        "formal_ui_ready": False,
    }


def build_collection_view(route: dict[str, Any], response: dict[str, Any], response_kind: str) -> dict[str, Any]:
    failure_code = response.get("failure_code")
    denied = failure_code is not None
    items = response.get("items") or []
    return {
        "route_id": route.get("id"),
        "read_model": route.get("read_model"),
        "response_kind": response_kind,
        "request_id": response.get("request_id"),
        "tenant_ref": response.get("tenant_ref"),
        "item_count": len(items),
        "next_cursor": response.get("next_cursor"),
        "failure_code": failure_code,
        "audit_ref": response.get("audit_ref"),
        "status_label": "denied" if denied else "ready",
        "can_render_items": not denied and len(items) > 0,
        "can_fetch_next_page": not denied and response.get("next_cursor") is not None,
        "no_side_effects_expected": True,
        "can_execute_workflow": False,
        "can_write_business_truth": False,
        "can_reveal_secrets": False,
    }


def has_forbidden_key(value: Any, forbidden_keys: set[str]) -> bool:
    if isinstance(value, list):
        return any(has_forbidden_key(item, forbidden_keys) for item in value)
    if isinstance(value, dict):
        return any(key in forbidden_keys or has_forbidden_key(nested, forbidden_keys) for key, nested in value.items())
    return False


def envelope_fields_present(response: dict[str, Any]) -> bool:
    return all(
        field in response
        for field in ("request_id", "tenant_ref", "items", "next_cursor", "failure_code", "audit_ref")
    )


def build_consumer_view() -> dict[str, Any]:
    routes = route_contracts_by_id()
    examples = response_examples_by_route_id()
    forbidden_keys = forbidden_output_keys()
    negative_contract = load_json(NEGATIVE_CONTRACT_FIXTURE)

    collection_views = []
    for route_id in sorted(routes):
        route = routes[route_id]
        example = examples[route_id]
        collection_views.append(build_collection_view(route, example["success"], "success"))
        collection_views.append(build_collection_view(route, example["failure"], "failure"))

    all_responses = [
        response
        for example in examples.values()
        for response in (example.get("success") or {}, example.get("failure") or {})
    ]
    failure_views = [view for view in collection_views if view.get("response_kind") == "failure"]

    return {
        "schema_version": 1,
        "kind": "control_plane_read_consumer_smoke_view",
        "source": "offline_response_fixtures",
        "route_catalog": build_route_catalog_view(routes),
        "view_models": {
            "collections": collection_views,
        },
        "invariants": {
            "all_routes_consumed": set(routes) == EXPECTED_ROUTE_IDS and set(examples) == EXPECTED_ROUTE_IDS,
            "all_envelopes_complete": all(envelope_fields_present(response) for response in all_responses),
            "failure_views_are_denied": all(view.get("status_label") == "denied" for view in failure_views),
            "failure_views_return_no_items": all(view.get("item_count") == 0 for view in failure_views),
            "audit_ref_required": all(bool(response.get("audit_ref")) for response in all_responses),
            "forbidden_output_absent": not has_forbidden_key(all_responses, forbidden_keys),
            "read_only_views": all(
                view.get("can_execute_workflow") is False and view.get("can_write_business_truth") is False
                for view in collection_views
            ),
            "secret_projection_disabled": all(view.get("can_reveal_secrets") is False for view in collection_views),
            "negative_contract_invariants_preserved": {
                "no_executor_invocation",
                "no_database_write",
                "no_business_writeback",
                "no_confirmation_decision",
                "no_replay",
            }.issubset(set(negative_contract.get("required_denial_invariants") or [])),
        },
    }


def assert_consumer_view(document: dict[str, Any]) -> None:
    if document.get("kind") != "control_plane_read_consumer_smoke_view":
        raise SystemExit("control plane read consumer view kind mismatch")
    invariants = document.get("invariants") or {}
    for key in (
        "all_routes_consumed",
        "all_envelopes_complete",
        "failure_views_are_denied",
        "failure_views_return_no_items",
        "audit_ref_required",
        "forbidden_output_absent",
        "read_only_views",
        "secret_projection_disabled",
        "negative_contract_invariants_preserved",
    ):
        if invariants.get(key) is not True:
            raise SystemExit(f"control plane read consumer invariant failed: {key}")
    catalog = document.get("route_catalog") or {}
    if catalog.get("database_backed") is not False or catalog.get("formal_ui_ready") is not False:
        raise SystemExit("control plane read consumer view incorrectly declares database or formal UI ready")
    for view in (document.get("view_models") or {}).get("collections") or []:
        if view.get("no_side_effects_expected") is not True:
            raise SystemExit("control plane read consumer view lost no-side-effects expectation")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", help="optional JSON output path")
    parser.add_argument("--check", action="store_true", help="validate safe read consumer invariants")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    document = build_consumer_view()
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
