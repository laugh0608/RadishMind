#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/gateway-api-key-quota-readiness.json"
DEPENDENCY_FIXTURES = {
    "product-surface-v1-boundary": REPO_ROOT / "scripts/checks/fixtures/product-surface-v1-boundary.json",
    "control-plane-data-boundary": REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json",
    "radish-oidc-client-preconditions": REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json",
}
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DOMAIN_IDS = {
    "api-key-lifecycle",
    "tenant-binding",
    "scope-policy",
    "quota-policy",
    "rate-limit-policy",
    "cost-ledger",
    "trace-record",
    "failure-taxonomy",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "api_key_lifecycle_ready",
    "api_key_issuance_ready",
    "api_key_storage_ready",
    "quota_enforcement_ready",
    "rate_limit_ready",
    "billing_ready",
    "cost_ledger_ready",
    "tenant_binding_enforced",
    "production_gateway_ready",
    "production_ready",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "API key generation",
    "API key hashing",
    "API key database storage",
    "API key validation middleware",
    "quota enforcement",
    "rate limiting implementation",
    "billing ledger implementation",
    "cost calculation",
    "durable trace store",
    "production gateway auth",
}
EXPECTED_FORBIDDEN_OUTPUTS = {
    "api_key_value",
    "api_key_hash",
    "authorization_header",
    "bearer_token",
    "cookie_value",
    "raw_request_body_dump",
    "raw_provider_credential",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "gateway-api-key-quota-readiness",
        "不发放真实 API key",
    ],
    "docs/radishmind-product-scope.md": [
        "gateway-api-key-quota-readiness",
        "API key",
        "quota",
        "trace",
    ],
    "docs/radishmind-architecture.md": [
        "gateway-api-key-quota-readiness",
        "quota",
        "cost ledger",
    ],
    "docs/radishmind-capability-matrix.md": [
        "gateway-api-key-quota-readiness",
        "gateway-api-key-quota-readiness.json",
        "check-gateway-api-key-quota-readiness.py",
    ],
    "docs/task-cards/control-plane-user-workspace-workflow-v1-plan.md": [
        "gateway-api-key-quota-readiness",
        "gateway-api-key-quota-readiness.json",
        "check-gateway-api-key-quota-readiness.py",
        "governance_boundary_satisfied",
    ],
    "scripts/README.md": [
        "check-gateway-api-key-quota-readiness.py",
        "gateway-api-key-quota-readiness.json",
        "API key、quota、rate limit、cost ledger",
    ],
    "docs/devlogs/2026-W22.md": [
        "gateway-api-key-quota-readiness",
        "gateway-api-key-quota-readiness.json",
        "check-gateway-api-key-quota-readiness.py",
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
    require(fixture.get("kind") == "gateway_api_key_quota_readiness", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "gateway-api-key-quota-readiness", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "gateway API key quota readiness must only satisfy a governance boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_readiness_domains(fixture: dict[str, Any]) -> None:
    domains = {
        str(item.get("id") or ""): item
        for item in fixture.get("readiness_domains") or []
        if isinstance(item, dict)
    }
    require(set(domains) == EXPECTED_DOMAIN_IDS, f"unexpected readiness domains: {sorted(domains)}")

    for domain_id, domain in domains.items():
        require(domain.get("status") == "defined_not_implemented", f"{domain_id} must remain defined_not_implemented")
        require(str(domain.get("boundary") or "").strip(), f"{domain_id}.boundary is required")

    require(
        "credential_hash_ref" in set(domains["api-key-lifecycle"].get("required_fields") or []),
        "api key lifecycle must require credential_hash_ref, not raw key storage",
    )
    require(
        "cross_tenant_denial" in set(domains["tenant-binding"].get("required_fields") or []),
        "tenant binding must define cross tenant denial",
    )
    require(
        {"models:read", "chat:invoke", "usage:read"}.issubset(
            set(domains["scope-policy"].get("required_scopes") or [])
        ),
        "scope policy must define read, invoke and usage scopes",
    )
    require(
        "over_quota_failure_code" in set(domains["quota-policy"].get("required_fields") or []),
        "quota policy must define over quota failure code",
    )
    require(
        "throttle_failure_code" in set(domains["rate-limit-policy"].get("required_fields") or []),
        "rate limit policy must define throttle failure code",
    )
    require(
        "estimated_cost" in set(domains["cost-ledger"].get("required_fields") or []),
        "cost ledger must define estimated_cost",
    )
    require(
        "failure_code" in set(domains["trace-record"].get("required_fields") or []),
        "trace record must define failure_code",
    )
    require(
        {"api_key_invalid", "quota_exceeded", "rate_limited"}.issubset(
            set(domains["failure-taxonomy"].get("required_failures") or [])
        ),
        "failure taxonomy must include key, quota and rate limit failures",
    )


def assert_scope_and_sanitization(fixture: dict[str, Any]) -> None:
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    policy = fixture.get("sanitization_policy") or {}
    allowed_outputs = set(policy.get("allowed_outputs") or [])
    require("api_key_id" in allowed_outputs, "sanitization policy must allow key id")
    require("failure_code" in allowed_outputs, "sanitization policy must allow failure code")
    missing_forbidden_outputs = sorted(EXPECTED_FORBIDDEN_OUTPUTS - set(policy.get("forbidden_outputs") or []))
    require(not missing_forbidden_outputs, f"missing forbidden outputs: {missing_forbidden_outputs}")

    require(
        "workflow-definition-run-record-boundary" in set(fixture.get("required_next_slices") or []),
        "workflow definition run record boundary must remain the next product slice",
    )


def assert_evidence_and_docs(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-gateway-api-key-quota-readiness.py", [])' in check_repo,
        "check-repo.py must run gateway API key quota readiness check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_readiness_domains(fixture)
    assert_scope_and_sanitization(fixture)
    assert_evidence_and_docs(fixture)
    print("gateway API key quota readiness checks passed.")


if __name__ == "__main__":
    main()
