#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-contract.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_FORBIDDEN_CLAIMS = {
    "production_ready",
    "production_secret_backend_ready",
    "cloud_secret_service_ready",
    "real_secret_written",
    "secret_rotation_ready",
    "production_provider_health_ready",
    "process_supervisor_ready",
}

REQUIRED_IDENTITY_FIELDS = {
    "environment",
    "provider",
    "provider_profile",
    "secret_ref",
}

REQUIRED_SANITIZED_OUTPUTS = {
    "credential_state",
    "secret_backend_configured",
    "secret_ref_present",
    "missing_secret_refs",
    "field_sources",
}

REQUIRED_FORBIDDEN_OUTPUTS = {
    "api key value",
    "token value",
    "cookie value",
    "authorization header value",
    "provider base URL value",
    "cloud secret payload",
}

REQUIRED_BACKEND_MODES = {
    "local_env_override": ("supported_dev_only", "forbidden"),
    "deploy_external_required": ("contract_defined_not_implemented", "blocked_until_secret_backend_adapter_exists"),
    "cloud_secret_service": ("not_implemented", "future_task_only"),
    "committed_file_secret_store": ("forbidden", "forbidden"),
}

REQUIRED_RUNTIME_GUARDS = {
    "config_summary_sanitized",
    "diagnostics_sanitized",
    "provider_inventory_sanitized",
    "deploy_env_example_non_secret",
}

REQUIRED_BLOCKED_CONDITIONS = {
    "production_secret_backend",
    "cloud_secret_service_integration",
    "secret_rotation_policy",
    "production_secret_audit_store",
}

REQUIRED_DOC_REFERENCES = {
    "services/platform/README.md": [
        "Production secret backend contract",
        "production-ops-secret-backend-contract.json",
        "check-production-ops-secret-backend-contract.py",
        "不实现真实云 secret 服务",
        "不写入真实 secret",
        "不声明 production ready",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "production-secret-backend-contract",
        "production-ops-secret-backend-contract.json",
        "check-production-ops-secret-backend-contract.py",
        "不实现真实云 secret 服务",
        "不写入真实 secret",
        "不声明 production ready",
    ],
    "docs/radishmind-roadmap.md": [
        "production-secret-backend-contract",
        "production-ops-secret-backend-contract.json",
        "production secret backend 仍为 not_satisfied",
    ],
    "docs/radishmind-capability-matrix.md": [
        "production secret backend contract",
        "production-ops-secret-backend-contract.json",
        "真实 production secret backend",
    ],
    "deploy/README.md": [
        "Production secret backend contract",
        "RADISHMIND_SECRET_SOURCE",
        "不是 secret backend",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-contract.py",
        "production-ops-secret-backend-contract.json",
    ],
    "docs/devlogs/2026-W22.md": [
        "production-secret-backend-contract",
        "production-ops-secret-backend-contract.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_fixture() -> dict[str, Any]:
    return json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))


def assert_slice(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "production-secret-backend-contract", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected slice track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "secret backend contract must only satisfy a governance boundary",
    )
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    missing = sorted(REQUIRED_FORBIDDEN_CLAIMS - does_not_claim)
    require(not missing, f"missing forbidden claims: {missing}")


def assert_selected_design(fixture: dict[str, Any]) -> None:
    design = fixture.get("selected_design") or {}
    require(design.get("id") == "external-secret-backend-adapter-contract", "unexpected selected design")
    require(design.get("implementation_status") == "not_implemented", "backend implementation must stay absent")
    require(
        design.get("production_secret_backend_status") == "not_satisfied",
        "production secret backend must stay not_satisfied",
    )
    require(design.get("cloud_secret_service_status") == "not_implemented", "cloud secret service must stay absent")
    require(design.get("storage_policy") == "no committed secret storage", "unexpected storage policy")
    real_secret_policy = str(design.get("real_secret_policy") or "")
    require("no real secret values" in real_secret_policy, "design must forbid real secret values")


def assert_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("backend_contract") or {}
    identity_fields = set(contract.get("required_identity_fields") or [])
    missing_identity = sorted(REQUIRED_IDENTITY_FIELDS - identity_fields)
    require(not missing_identity, f"missing identity fields: {missing_identity}")

    allowed_refs = set(contract.get("allowed_secret_reference_fields") or [])
    require("RADISHMIND_SECRET_SOURCE" in allowed_refs, "secret source env must be a reference only")
    require("orchestrator secret name" in allowed_refs, "future orchestrator secret ref must be named")

    resolution = str(contract.get("future_resolution_boundary") or "")
    require("does not resolve values" in resolution, "contract must not resolve secret values")

    sanitized_outputs = set(contract.get("required_sanitized_outputs") or [])
    missing_outputs = sorted(REQUIRED_SANITIZED_OUTPUTS - sanitized_outputs)
    require(not missing_outputs, f"missing sanitized outputs: {missing_outputs}")

    forbidden_outputs = set(contract.get("forbidden_outputs") or [])
    missing_forbidden = sorted(REQUIRED_FORBIDDEN_OUTPUTS - forbidden_outputs)
    require(not missing_forbidden, f"missing forbidden outputs: {missing_forbidden}")


def assert_backend_modes(fixture: dict[str, Any]) -> None:
    modes = {str(item.get("id")): item for item in fixture.get("backend_modes") or [] if isinstance(item, dict)}
    missing_modes = sorted(set(REQUIRED_BACKEND_MODES) - set(modes))
    require(not missing_modes, f"missing backend modes: {missing_modes}")
    for mode_id, (status, production_use) in REQUIRED_BACKEND_MODES.items():
        mode = modes[mode_id]
        require(mode.get("status") == status, f"{mode_id} has unexpected status")
        require(mode.get("production_use") == production_use, f"{mode_id} has unexpected production use")

    require(modes["local_env_override"].get("source") == "RADISHMIND_PLATFORM_API_KEY", "unexpected local env source")
    require(modes["deploy_external_required"].get("source") == "RADISHMIND_SECRET_SOURCE", "unexpected deploy source")
    require(
        modes["cloud_secret_service"].get("cloud_calls") == "forbidden_in_this_slice",
        "cloud secret calls must be forbidden",
    )


def assert_runtime_guards_and_blocks(fixture: dict[str, Any]) -> None:
    guards = {str(item.get("id")): item for item in fixture.get("runtime_guards") or [] if isinstance(item, dict)}
    missing_guards = sorted(REQUIRED_RUNTIME_GUARDS - set(guards))
    require(not missing_guards, f"missing runtime guards: {missing_guards}")
    for guard_id in REQUIRED_RUNTIME_GUARDS:
        require(guards[guard_id].get("status") == "satisfied", f"{guard_id} must be satisfied")
        boundary = str(guards[guard_id].get("required_boundary") or "")
        require("no raw" in boundary or "no API key" in boundary, f"{guard_id} must forbid raw secrets")

    blocked = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or [] if isinstance(item, dict)}
    missing_blocks = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing_blocks, f"missing blocked conditions: {missing_blocks}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        require(blocked[condition_id].get("status") == "not_satisfied", f"{condition_id} must stay not_satisfied")


def assert_evidence_consumers_and_docs(fixture: dict[str, Any]) -> None:
    for evidence in fixture.get("evidence") or []:
        require((REPO_ROOT / str(evidence)).exists(), f"missing evidence path: {evidence}")

    expected_consumers = {
        "scripts/check-production-ops-secret-backend-contract.py",
        "scripts/check-repo.py",
        "scripts/check-p3-local-product-shell-short-close-checklist.py",
        "scripts/README.md",
        "docs/radishmind-roadmap.md",
        "docs/radishmind-capability-matrix.md",
        "docs/task-cards/production-ops-hardening-v1-plan.md",
        "services/platform/README.md",
        "deploy/README.md",
        "docs/devlogs/2026-W22.md",
    }
    consumers = set(fixture.get("required_consumers") or [])
    missing_consumers = sorted(expected_consumers - consumers)
    require(not missing_consumers, f"missing consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-secret-backend-contract.py", [])' in check_repo,
        "check-repo.py must run production secret backend contract check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def assert_no_real_secret_literals() -> None:
    fixture_text = FIXTURE_PATH.read_text(encoding="utf-8")
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA"]
    found = [literal for literal in forbidden_literals if literal in fixture_text]
    require(not found, f"fixture contains forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", fixture_text) is None,
        "fixture contains forbidden secret-looking sk token",
    )


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "production_ops_secret_backend_contract", "unexpected fixture kind")
    assert_slice(fixture)
    assert_selected_design(fixture)
    assert_contract(fixture)
    assert_backend_modes(fixture)
    assert_runtime_guards_and_blocks(fixture)
    assert_evidence_consumers_and_docs(fixture)
    assert_no_real_secret_literals()
    print("production ops secret backend contract checks passed.")


if __name__ == "__main__":
    main()
