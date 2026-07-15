#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json"
PRODUCT_SURFACE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/product-surface-v1-boundary.json"
CONTROL_PLANE_DATA_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_PRECONDITION_IDS = {
    "issuer-discovery",
    "client-registration",
    "claim-mapping",
    "tenant-binding",
    "session-boundary",
    "logout-and-revocation",
    "audit-events",
    "failure-taxonomy",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "radish_oidc_client_ready",
    "login_flow_ready",
    "logout_flow_ready",
    "token_validation_ready",
    "client_secret_configured",
    "session_cookie_ready",
    "tenant_auth_ready",
    "permission_enforcement_ready",
    "production_auth_ready",
    "production_ready",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "network call to Radish issuer",
    "OIDC middleware implementation",
    "token validation implementation",
    "login redirect route",
    "logout route",
    "session cookie implementation",
    "client secret storage",
    "tenant permission enforcement",
    "production auth policy",
    "production admin console",
}
EXPECTED_FORBIDDEN_OUTPUTS = {
    "id_token",
    "access_token",
    "refresh_token",
    "client_secret",
    "authorization_header",
    "cookie_value",
    "jwks_raw_dump",
}
EXPECTED_REQUIRED_NEXT_SLICES = {
    "gateway-api-key-quota-readiness",
    "workflow-definition-run-record-boundary",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "radish-oidc-client-preconditions",
        "不接真实 OIDC",
    ],
    "docs/radishmind-product-scope.md": [
        "radish-oidc-client-preconditions",
        "issuer",
        "claim mapping",
        "tenant binding",
    ],
    "docs/radishmind-architecture.md": [
        "radish-oidc-client-preconditions",
        "Radish remains identity truth",
    ],
    "docs/radishmind-capability-matrix.md": [
        "radish-oidc-client-preconditions",
        "radish-oidc-client-preconditions.json",
        "check-radish-oidc-client-preconditions.py",
    ],
    "docs/task-cards/control-plane-user-workspace-workflow-v1-plan.md": [
        "radish-oidc-client-preconditions",
        "radish-oidc-client-preconditions.json",
        "check-radish-oidc-client-preconditions.py",
        "governance_boundary_satisfied",
    ],
    "scripts/README.md": [
        "check-radish-oidc-client-preconditions.py",
        "radish-oidc-client-preconditions.json",
        "issuer、client、claim mapping、tenant binding",
    ],
    "docs/devlogs/2026-W22.md": [
        "radish-oidc-client-preconditions",
        "radish-oidc-client-preconditions.json",
        "check-radish-oidc-client-preconditions.py",
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


def assert_dependencies() -> None:
    product_surface = load_json(PRODUCT_SURFACE_FIXTURE_PATH)
    control_plane_data = load_json(CONTROL_PLANE_DATA_FIXTURE_PATH)
    require(
        (product_surface.get("slice") or {}).get("id") == "product-surface-v1-boundary"
        and (product_surface.get("slice") or {}).get("status") == "governance_boundary_satisfied",
        "radish oidc preconditions depend on satisfied product-surface-v1-boundary",
    )
    require(
        (control_plane_data.get("slice") or {}).get("id") == "control-plane-data-boundary"
        and (control_plane_data.get("slice") or {}).get("status") == "governance_boundary_satisfied",
        "radish oidc preconditions depend on satisfied control-plane-data-boundary",
    )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "radish_oidc_client_preconditions", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "radish-oidc-client-preconditions", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "Radish OIDC preconditions must only satisfy a governance boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")

    depends_on = set(fixture.get("depends_on") or [])
    require(
        {"product-surface-v1-boundary", "control-plane-data-boundary"}.issubset(depends_on),
        "fixture must declare product-surface and control-plane-data dependencies",
    )


def assert_preconditions(fixture: dict[str, Any]) -> None:
    preconditions = {
        str(item.get("id") or ""): item
        for item in fixture.get("preconditions") or []
        if isinstance(item, dict)
    }
    require(set(preconditions) == EXPECTED_PRECONDITION_IDS, f"unexpected preconditions: {sorted(preconditions)}")

    for precondition_id, item in preconditions.items():
        require(
            item.get("status") == "defined_not_implemented",
            f"{precondition_id} must stay defined_not_implemented",
        )
        require(str(item.get("boundary") or "").strip(), f"{precondition_id}.boundary is required")

    require(
        "issuer_url" in set(preconditions["issuer-discovery"].get("required_fields") or []),
        "issuer discovery must require issuer_url",
    )
    require(
        "client_secret_ref" in set(preconditions["client-registration"].get("required_fields") or []),
        "client registration must require client_secret_ref",
    )
    require(
        {"sub", "tenant_id", "roles", "permissions"}.issubset(
            set(preconditions["claim-mapping"].get("required_claims") or [])
        ),
        "claim mapping must include subject, tenant, role and permission claims",
    )
    require(
        "cross_tenant_denial" in set(preconditions["tenant-binding"].get("required_fields") or []),
        "tenant binding must define cross tenant denial",
    )
    require(
        "oidc_tenant_binding_denied" in set(preconditions["audit-events"].get("required_events") or []),
        "audit events must include tenant binding denial",
    )
    require(
        "tenant_binding_denied" in set(preconditions["failure-taxonomy"].get("required_failures") or []),
        "failure taxonomy must include tenant binding denied",
    )


def assert_scope_and_sanitization(fixture: dict[str, Any]) -> None:
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    policy = fixture.get("sanitization_policy") or {}
    allowed_outputs = set(policy.get("allowed_outputs") or [])
    require("client_secret_ref_present" in allowed_outputs, "sanitization policy must allow secret ref status only")
    missing_forbidden_outputs = sorted(EXPECTED_FORBIDDEN_OUTPUTS - set(policy.get("forbidden_outputs") or []))
    require(not missing_forbidden_outputs, f"missing forbidden outputs: {missing_forbidden_outputs}")

    missing_next = sorted(EXPECTED_REQUIRED_NEXT_SLICES - set(fixture.get("required_next_slices") or []))
    require(not missing_next, f"missing next slices: {missing_next}")


def assert_evidence_and_docs(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-radish-oidc-client-preconditions.py", [])' in check_repo,
        "check-repo.py must run Radish OIDC client preconditions check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies()
    assert_slice(fixture)
    assert_preconditions(fixture)
    assert_scope_and_sanitization(fixture)
    assert_evidence_and_docs(fixture)
    print("Radish OIDC client preconditions checks passed.")


if __name__ == "__main__":
    main()
