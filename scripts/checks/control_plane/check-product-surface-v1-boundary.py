#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/product-surface-v1-boundary.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_SURFACE_IDS = {
    "user_workspace",
    "admin_control_plane",
    "model_gateway_api_distribution",
    "workflow_agent_runtime",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "formal_user_workspace_ready",
    "production_admin_console_ready",
    "radish_oidc_client_ready",
    "database_schema_ready",
    "api_key_lifecycle_ready",
    "quota_billing_ready",
    "workflow_executor_ready",
    "tool_executor_ready",
    "confirmation_writeback_replay_ready",
    "production_ready",
}
EXPECTED_REQUIRED_NEXT_SLICES = {
    "control-plane-data-boundary",
    "radish-oidc-client-preconditions",
    "gateway-api-key-quota-readiness",
    "workflow-definition-run-record-boundary",
}
REQUIRED_SHARED_STOP_LINES = {
    "do not turn apps/radishmind-console into formal user workspace or production admin console",
    "do not merge Control Plane, Gateway and Workflow Executor into one implicit monolith",
    "do not create a second identity or permission truth source that conflicts with Radish",
    "do not use service split as a reason to introduce a new default backend language",
    "do not claim production ready from local smoke, provider health or Docker static boundary",
    "do not allow model, workflow or tool output to write upstream business truth directly",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "Control Plane / User Workspace / Workflow v1",
        "product-surface-v1-boundary",
        "不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay",
    ],
    "docs/radishmind-current-focus.md": [
        "Control Plane / User Workspace / Workflow v1",
        "product-surface-v1-boundary",
        "product-surface-v1-boundary.json",
        "check-product-surface-v1-boundary.py",
    ],
    "docs/radishmind-product-scope.md": [
        "product-surface-v1-boundary",
        "User Workspace",
        "Admin Control Plane",
        "Workflow / Agent Runtime",
    ],
    "docs/radishmind-architecture.md": [
        "product-surface-v1-boundary",
        "Control Plane / User Workspace / Workflow v1",
        "不实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay",
    ],
    "docs/radishmind-roadmap.md": [
        "product-surface-v1-boundary",
        "product-surface-v1-boundary.json",
        "check-product-surface-v1-boundary.py",
    ],
    "docs/radishmind-capability-matrix.md": [
        "product-surface-v1-boundary",
        "product-surface-v1-boundary.json",
        "check-product-surface-v1-boundary.py",
    ],
    "docs/task-cards/control-plane-user-workspace-workflow-v1-plan.md": [
        "product-surface-v1-boundary",
        "product-surface-v1-boundary.json",
        "check-product-surface-v1-boundary.py",
        "governance_boundary_satisfied",
    ],
    "scripts/README.md": [
        "check-product-surface-v1-boundary.py",
        "product-surface-v1-boundary.json",
        "Control Plane / User Workspace / Workflow v1",
    ],
    "docs/devlogs/2026-W22.md": [
        "product-surface-v1-boundary",
        "product-surface-v1-boundary.json",
        "check-product-surface-v1-boundary.py",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_fixture() -> dict[str, Any]:
    document = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    require(isinstance(document, dict), "fixture must be a JSON object")
    return document


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "product_surface_v1_boundary", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "product-surface-v1-boundary", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "product surface v1 boundary must only satisfy a governance boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_surfaces(fixture: dict[str, Any]) -> None:
    surfaces = {
        str(item.get("id") or ""): item
        for item in fixture.get("product_surfaces") or []
        if isinstance(item, dict)
    }
    require(set(surfaces) == EXPECTED_SURFACE_IDS, f"unexpected product surfaces: {sorted(surfaces)}")

    for surface_id, surface in surfaces.items():
        require(
            surface.get("v1_status") == "boundary_defined_not_implemented",
            f"{surface_id} must stay boundary-defined only",
        )
        for field in ("resources", "read_model", "write_boundary", "blocked_capabilities", "stop_lines"):
            values = surface.get(field)
            require(isinstance(values, list) and values, f"{surface_id}.{field} must be a non-empty list")

    require("user_api_key" in set(surfaces["user_workspace"]["resources"]), "user workspace must name user_api_key")
    require(
        "Radish OIDC client" in set(surfaces["admin_control_plane"]["blocked_capabilities"]),
        "admin control plane must keep Radish OIDC blocked",
    )
    require(
        "production secret backend" in set(surfaces["model_gateway_api_distribution"]["blocked_capabilities"]),
        "model gateway must keep production secret backend blocked",
    )
    require(
        "workflow executor" in set(surfaces["workflow_agent_runtime"]["blocked_capabilities"]),
        "workflow runtime must keep workflow executor blocked",
    )


def assert_boundaries_and_evidence(fixture: dict[str, Any]) -> None:
    shared_stop_lines = set(fixture.get("shared_stop_lines") or [])
    missing_stop_lines = sorted(REQUIRED_SHARED_STOP_LINES - shared_stop_lines)
    require(not missing_stop_lines, f"missing shared stop lines: {missing_stop_lines}")

    next_slices = set(fixture.get("required_next_slices") or [])
    missing_next = sorted(EXPECTED_REQUIRED_NEXT_SLICES - next_slices)
    require(not missing_next, f"missing next slices: {missing_next}")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")

    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")


def assert_docs_and_check_repo() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-product-surface-v1-boundary.py", [])' in check_repo,
        "check-repo.py must run product surface v1 boundary check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_fixture()
    assert_slice(fixture)
    assert_surfaces(fixture)
    assert_boundaries_and_evidence(fixture)
    assert_docs_and_check_repo()
    print("product surface v1 boundary checks passed.")


if __name__ == "__main__":
    main()
