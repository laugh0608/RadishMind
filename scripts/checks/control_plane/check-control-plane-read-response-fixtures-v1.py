#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
READ_MODEL_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-model-v1.json"
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "control_plane_api_ready",
    "go_handler_ready",
    "typescript_consumer_ready",
    "database_query_ready",
    "radish_oidc_client_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "workflow_executor_ready",
    "production_ready",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "Go route handler",
    "TypeScript consumer contract",
    "database schema",
    "database migration",
    "database query implementation",
    "Radish OIDC integration",
    "tenant or user CRUD",
    "API key generation",
    "quota enforcement",
    "rate limiting implementation",
    "billing ledger implementation",
    "workflow builder",
    "workflow executor",
    "confirmation flow wiring",
    "business writeback",
    "replay execution",
    "production admin console",
    "user workspace UI",
}
EXPECTED_FORBIDDEN_OUTPUT_KEYS = {
    "raw_secret_value",
    "api_key_value",
    "api_key_hash",
    "authorization_header",
    "bearer_token",
    "cookie_value",
    "raw_request_body_dump",
    "raw_tool_payload",
    "business_writeback_payload",
    "full_prompt_dump_with_secret",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-read-response-fixtures-v1",
        "response fixture",
        "不创建数据库 schema",
    ],
    "docs/radishmind-product-scope.md": [
        "control-plane-read-response-fixtures-v1",
        "response fixture",
        "failure_code",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-response-fixtures-v1",
        "response fixture",
        "failure_code",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-response-fixtures-v1",
        "control-plane-read-response-fixtures-v1.json",
        "check-control-plane-read-response-fixtures-v1.py",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-response-fixtures-v1",
        "Control Plane Read Response Fixtures",
    ],
    "docs/task-cards/control-plane-read-response-fixtures-v1-plan.md": [
        "control-plane-read-response-fixtures-v1",
        "response fixture",
        "failure_code",
    ],
    "scripts/README.md": [
        "check-control-plane-read-response-fixtures-v1.py",
        "control-plane-read-response-fixtures-v1.json",
        "response fixture",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-response-fixtures-v1",
        "control-plane-read-response-fixtures-v1.json",
        "check-control-plane-read-response-fixtures-v1.py",
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


def iter_keys(value: Any) -> set[str]:
    if isinstance(value, dict):
        keys = set(value)
        for child in value.values():
            keys.update(iter_keys(child))
        return keys
    if isinstance(value, list):
        keys: set[str] = set()
        for item in value:
            keys.update(iter_keys(item))
        return keys
    return set()


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    require({"control-plane-read-model-v1", "control-plane-read-route-contract-v1"}.issubset(declared), "fixture must declare dependencies")
    for expected_id, path in {
        "control-plane-read-model-v1": READ_MODEL_FIXTURE_PATH,
        "control-plane-read-route-contract-v1": ROUTE_CONTRACT_FIXTURE_PATH,
    }.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(
            slice_info.get("id") == expected_id
            and slice_info.get("status") == "governance_boundary_satisfied",
            f"dependency {expected_id} must remain satisfied",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_response_fixtures_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-response-fixtures-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "control plane read response fixtures must only satisfy a governance boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_response_examples(fixture: dict[str, Any]) -> None:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    route_by_id = {
        str(route.get("id") or ""): route
        for route in route_contract.get("route_contracts") or []
        if isinstance(route, dict)
    }
    examples = {
        str(item.get("route_id") or ""): item
        for item in fixture.get("response_examples") or []
        if isinstance(item, dict)
    }
    require(set(examples) == set(route_by_id), "response examples must cover every route contract exactly once")

    envelope_fields = set(fixture.get("response_envelope_fields") or [])
    require(
        {"request_id", "tenant_ref", "items", "next_cursor", "failure_code", "audit_ref"}.issubset(envelope_fields),
        "response envelope fields are incomplete",
    )

    forbidden_keys = set(fixture.get("forbidden_output_keys") or [])
    require(EXPECTED_FORBIDDEN_OUTPUT_KEYS.issubset(forbidden_keys), "forbidden output keys are incomplete")

    for route_id, example in examples.items():
        route = route_by_id[route_id]
        require(example.get("read_model") == route.get("read_model"), f"{route_id} read model mismatch")
        for response_name in ("success", "failure"):
            response = example.get(response_name)
            require(isinstance(response, dict), f"{route_id}.{response_name} must be an object")
            missing_envelope = sorted(envelope_fields - set(response))
            require(not missing_envelope, f"{route_id}.{response_name} missing envelope fields: {missing_envelope}")
            require(response.get("tenant_ref"), f"{route_id}.{response_name} must include tenant_ref")
            require(isinstance(response.get("items"), list), f"{route_id}.{response_name}.items must be a list")
            leaked_keys = sorted(forbidden_keys & iter_keys(response))
            require(not leaked_keys, f"{route_id}.{response_name} leaks forbidden keys: {leaked_keys}")

        success = example["success"]
        failure = example["failure"]
        require(success.get("failure_code") is None, f"{route_id}.success must not include failure_code")
        require(success.get("items"), f"{route_id}.success must include at least one item")
        failure_code = failure.get("failure_code")
        require(isinstance(failure_code, str) and failure_code, f"{route_id}.failure must include failure_code")
        require(failure_code in set(route.get("failure_codes") or []), f"{route_id}.failure_code must be declared by route contract")
        require(failure.get("items") == [], f"{route_id}.failure must not return items")

        if route.get("pagination") == "cursor_required":
            require("next_cursor" in success, f"{route_id}.success must keep cursor envelope")
        else:
            require(success.get("next_cursor") is None, f"{route_id}.single resource success must have null cursor")


def assert_policy_and_docs(fixture: dict[str, Any]) -> None:
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        '"check-control-plane-read-response-fixtures-v1.py"' in check_repo,
        "check-repo.py must catalog control plane read response fixtures check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_response_examples(fixture)
    assert_policy_and_docs(fixture)
    print("control plane read response fixtures v1 checks passed.")


if __name__ == "__main__":
    main()
