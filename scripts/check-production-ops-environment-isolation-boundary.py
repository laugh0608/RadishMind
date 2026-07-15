#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-environment-isolation-boundary.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_READINESS_SCOPES = {
    "local_readiness": "supported",
    "dev_smoke": "supported",
    "production_readiness": "not_satisfied",
}

REQUIRED_LOCAL_ROUTES = {
    "/healthz",
    "/v1/platform/overview",
    "/v1/platform/local-smoke",
}

REQUIRED_ENVIRONMENT_BOUNDARIES = {
    "mock_provider": ("local", "forbidden"),
    "local_smoke": ("local", "forbidden"),
    "local_console_cors": ("local", "forbidden"),
    "developer_remote_profile": (
        "dev",
        "blocked_until_secret_backend_provider_health_and_environment_isolation_exist",
    ),
    "console_preview": ("dev", "blocked_until_console_production_packaging_exists"),
}

REQUIRED_BLOCKED_CONDITIONS = {
    "deployment_environment_isolation",
    "production_readiness_gate",
    "production_auth_policy",
    "production_cors_policy",
    "provider_health_policy",
}

REQUIRED_FORBIDDEN_INTERPRETATIONS = {
    "local-smoke is production health",
    "mock provider is production provider",
    "demo profile is production profile",
    "developer env override is production secret backend",
    "localhost CORS is production CORS policy",
    "Vite preview is production package",
}

REQUIRED_SOURCE_LITERALS = {
    "services/platform/internal/httpapi/platform_local_smoke.go": [
        "platform_local_smoke",
        "local_dev_only",
        "it does not start processes or enable executor",
        "local_console_ready",
    ],
    "services/platform/internal/httpapi/platform_overview.go": [
        "local_read_only_product_shell",
        "production_secret_backend_ready",
        "platformProductSurfaceRoutes",
    ],
    "contracts/typescript/platform-local-smoke-api.ts": [
        "cors_scope: \"local_dev_only\"",
        "production_secret_backend_ready: false",
        "canExecuteActions: false",
    ],
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "environment-isolation",
        "production-ops-environment-isolation-boundary.json",
        "check-production-ops-environment-isolation-boundary.py",
    ],
    "scripts/README.md": [
        "check-production-ops-environment-isolation-boundary.py",
        "production-ops-environment-isolation-boundary.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_fixture() -> dict[str, Any]:
    document = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    require(isinstance(document, dict), "environment isolation boundary fixture must be an object")
    return document


def assert_slice(document: dict[str, Any]) -> None:
    slice_info = document.get("slice") or {}
    require(slice_info.get("id") == "environment-isolation", "unexpected production ops slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected production ops track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "environment-isolation must only satisfy the governance boundary",
    )
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    required_forbidden_claims = {
        "production_ready",
        "production_environment_isolation_ready",
        "production_auth_policy_ready",
        "production_cors_policy_ready",
        "provider_health_policy_ready",
        "console_production_package_ready",
    }
    missing = sorted(required_forbidden_claims - forbidden_claims)
    require(not missing, f"missing forbidden environment claims: {missing}")


def assert_readiness_scopes(document: dict[str, Any]) -> None:
    scopes = {str(item.get("id")): item for item in document.get("readiness_scopes") or [] if isinstance(item, dict)}
    missing = sorted(set(REQUIRED_READINESS_SCOPES) - set(scopes))
    require(not missing, f"missing readiness scopes: {missing}")

    for scope_id, expected_status in REQUIRED_READINESS_SCOPES.items():
        item = scopes[scope_id]
        require(item.get("status") == expected_status, f"{scope_id} has unexpected status")
        require(item.get("must_not_claim"), f"{scope_id} must document must_not_claim")

    local_routes = set(scopes["local_readiness"].get("evidence_routes") or [])
    missing_routes = sorted(REQUIRED_LOCAL_ROUTES - local_routes)
    require(not missing_routes, f"local readiness missing routes: {missing_routes}")
    production_routes = list(scopes["production_readiness"].get("evidence_routes") or [])
    production_profiles = list(scopes["production_readiness"].get("allowed_profiles") or [])
    require(production_routes == [], "production readiness must not expose evidence routes in this slice")
    require(production_profiles == [], "production readiness must not expose allowed profiles in this slice")


def assert_environment_boundaries(document: dict[str, Any]) -> None:
    boundaries = {
        str(item.get("id")): item
        for item in document.get("environment_boundaries") or []
        if isinstance(item, dict)
    }
    missing = sorted(set(REQUIRED_ENVIRONMENT_BOUNDARIES) - set(boundaries))
    require(not missing, f"missing environment boundaries: {missing}")

    for boundary_id, (environment, production_use) in REQUIRED_ENVIRONMENT_BOUNDARIES.items():
        item = boundaries[boundary_id]
        require(item.get("environment") == environment, f"{boundary_id} has unexpected environment")
        require(item.get("production_use") == production_use, f"{boundary_id} has unexpected production use")
        require(item.get("reason"), f"{boundary_id} must document reason")


def assert_blocked_and_forbidden(document: dict[str, Any]) -> None:
    blocked = {str(item.get("id")): item for item in document.get("blocked_conditions") or [] if isinstance(item, dict)}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing_blocked, f"missing blocked environment conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        item = blocked[condition_id]
        require(item.get("status") == "not_satisfied", f"{condition_id} must remain not_satisfied")
        require(
            item.get("required_before_production_ready") is True,
            f"{condition_id} must gate production ready",
        )

    forbidden = set(document.get("forbidden_interpretations") or [])
    missing_forbidden = sorted(REQUIRED_FORBIDDEN_INTERPRETATIONS - forbidden)
    require(not missing_forbidden, f"missing forbidden interpretations: {missing_forbidden}")


def assert_evidence_and_consumers(document: dict[str, Any]) -> None:
    for evidence in document.get("evidence") or []:
        evidence_path = REPO_ROOT / str(evidence)
        require(evidence_path.exists(), f"missing environment isolation evidence path: {evidence}")

    consumers = set(document.get("required_consumers") or [])
    expected_consumers = {
        "scripts/check-production-ops-environment-isolation-boundary.py",
        "scripts/check-repo.py",
        "scripts/check-p3-local-product-shell-short-close-checklist.py",
        "scripts/README.md",
    }
    missing_consumers = sorted(expected_consumers - consumers)
    require(not missing_consumers, f"missing environment isolation consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-environment-isolation-boundary.py", [])' in check_repo,
        "check-repo.py must run production ops environment isolation boundary check",
    )


def assert_source_literals() -> None:
    for relative_path, required_literals in REQUIRED_SOURCE_LITERALS.items():
        content = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing = [literal for literal in required_literals if literal not in content]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_doc_references() -> None:
    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        content = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing = [literal for literal in required_literals if literal not in content]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> int:
    document = load_fixture()
    require(document.get("schema_version") == 1, "unexpected schema_version")
    require(document.get("kind") == "production_ops_environment_isolation_boundary", "unexpected fixture kind")
    assert_slice(document)
    assert_readiness_scopes(document)
    assert_environment_boundaries(document)
    assert_blocked_and_forbidden(document)
    assert_evidence_and_consumers(document)
    assert_source_literals()
    assert_doc_references()
    print("production ops environment isolation boundary checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
