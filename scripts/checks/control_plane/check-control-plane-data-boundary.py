#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json"
PRODUCT_SURFACE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/product-surface-v1-boundary.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_ENTITY_IDS = {
    "tenant",
    "user",
    "role",
    "permission",
    "provider_profile",
    "model_route",
    "quota",
    "price",
    "audit",
    "secret_ref",
    "deployment_status",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "database_schema_ready",
    "database_migration_ready",
    "control_plane_api_ready",
    "tenant_auth_ready",
    "radish_oidc_client_ready",
    "quota_enforcement_ready",
    "billing_ready",
    "production_secret_backend_ready",
    "production_admin_console_ready",
    "production_ready",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "database migration",
    "control plane API implementation",
    "Radish OIDC integration",
    "tenant or user CRUD",
    "quota enforcement",
    "billing ledger",
    "secret resolver implementation",
    "workflow executor",
    "production admin console",
}
EXPECTED_REQUIRED_NEXT_SLICES = {
    "radish-oidc-client-preconditions",
    "gateway-api-key-quota-readiness",
    "workflow-definition-run-record-boundary",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-data-boundary",
        "不创建数据库 schema 或 migration",
    ],
    "docs/radishmind-current-focus.md": [
        "control-plane-data-boundary",
        "control-plane-data-boundary.json",
        "check-control-plane-data-boundary.py",
    ],
    "docs/radishmind-product-scope.md": [
        "control-plane-data-boundary",
        "tenant",
        "provider profile",
        "deployment status",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-data-boundary",
        "Radish remains identity",
    ],
    "docs/radishmind-roadmap.md": [
        "control-plane-data-boundary",
        "control-plane-data-boundary.json",
        "check-control-plane-data-boundary.py",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-data-boundary",
        "control-plane-data-boundary.json",
        "check-control-plane-data-boundary.py",
    ],
    "docs/task-cards/control-plane-user-workspace-workflow-v1-plan.md": [
        "control-plane-data-boundary",
        "control-plane-data-boundary.json",
        "check-control-plane-data-boundary.py",
        "governance_boundary_satisfied",
    ],
    "scripts/README.md": [
        "check-control-plane-data-boundary.py",
        "control-plane-data-boundary.json",
        "tenant、user、role、permission",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-data-boundary",
        "control-plane-data-boundary.json",
        "check-control-plane-data-boundary.py",
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


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_data_boundary", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-data-boundary", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "control plane data boundary must only satisfy a governance boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_product_surface_dependency() -> None:
    product_surface = load_json(PRODUCT_SURFACE_FIXTURE_PATH)
    slice_info = product_surface.get("slice") or {}
    require(
        slice_info.get("id") == "product-surface-v1-boundary"
        and slice_info.get("status") == "governance_boundary_satisfied",
        "control plane data boundary depends on satisfied product-surface-v1-boundary",
    )


def assert_ownership_model(fixture: dict[str, Any]) -> None:
    entities = {
        str(item.get("id") or ""): item
        for item in fixture.get("ownership_model") or []
        if isinstance(item, dict)
    }
    require(set(entities) == EXPECTED_ENTITY_IDS, f"unexpected ownership entities: {sorted(entities)}")

    for entity_id, entity in entities.items():
        for field in ("owner_surface", "source_of_truth", "write_boundary"):
            require(str(entity.get(field) or "").strip(), f"{entity_id}.{field} is required")
        read_surfaces = entity.get("read_surfaces")
        blocked_until = entity.get("blocked_until")
        require(isinstance(read_surfaces, list) and read_surfaces, f"{entity_id}.read_surfaces must be non-empty")
        require(isinstance(blocked_until, list) and blocked_until, f"{entity_id}.blocked_until must be non-empty")

    require("Radish" in str(entities["user"]["source_of_truth"]), "user source of truth must remain Radish")
    require(
        "future audited admin API" in str(entities["tenant"]["write_boundary"]),
        "tenant writes must stay future audited admin API only",
    )
    require(
        "never secret values" in str(entities["secret_ref"]["write_boundary"]),
        "secret_ref write boundary must forbid secret values",
    )
    require(
        "no manual production-ready override" in str(entities["deployment_status"]["write_boundary"]),
        "deployment status must forbid manual production ready override",
    )


def assert_shared_policies(fixture: dict[str, Any]) -> None:
    policies = fixture.get("shared_policies") or {}
    require("Radish remains identity" in str(policies.get("identity_truth") or ""), "identity truth policy drifted")
    require("not created" in str(policies.get("storage_truth") or ""), "storage truth must forbid schema creation")
    require("audited control plane APIs" in str(policies.get("write_policy") or ""), "write policy must require audit")
    require("secret values may not be committed" in str(policies.get("secret_policy") or ""), "secret policy drifted")
    require("do not own control plane truth" in str(policies.get("runtime_policy") or ""), "runtime policy drifted")

    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    missing_next = sorted(EXPECTED_REQUIRED_NEXT_SLICES - set(fixture.get("required_next_slices") or []))
    require(not missing_next, f"missing next slices: {missing_next}")


def assert_evidence_and_docs(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-data-boundary.py", [])' in check_repo,
        "check-repo.py must run control plane data boundary check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_product_surface_dependency()
    assert_ownership_model(fixture)
    assert_shared_policies(fixture)
    assert_evidence_and_docs(fixture)
    print("control plane data boundary checks passed.")


if __name__ == "__main__":
    main()
