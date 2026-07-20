#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-config-secret-boundary.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_CONFIG_SOURCES = {
    "local_defaults": ("local", "supported", "local_readiness_only"),
    "developer_config_file": ("dev", "supported", "developer_smoke_only"),
    "developer_env_override": ("dev", "supported", "developer_smoke_only"),
    "production_secret_backend": (
        "production",
        "not_satisfied",
        "blocked_until_secret_backend_environment_isolation_and_supervisor_exist",
    ),
}

REQUIRED_SECRET_FIELDS = {
    "api_key",
    "token",
    "cookie",
    "authorization",
    "credential",
}

REQUIRED_SANITIZED_OUTPUTS = {
    "config-summary",
    "config-check",
    "diagnostics",
    "provider inventory",
}

REQUIRED_SANITIZED_FIELDS = {
    "credential_state",
    "base_url_configured",
    "secret_fields",
    "field_sources",
}

REQUIRED_PROVIDER_BOUNDARIES = {
    "mock": ("not_required", "forbidden"),
    "ollama": ("optional_missing_or_configured", "blocked_until_environment_isolation_and_supervisor_exist"),
    "remote_api_provider": ("must_be_configured_before_use", "blocked_until_secret_backend_and_provider_health_policy_exist"),
}

REQUIRED_BLOCKED_CONDITIONS = {
    "production_secret_backend",
    "production_environment_isolation",
    "production_provider_health_policy",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "config-secret-boundary",
        "governance boundary",
        "production-ops-config-secret-boundary.json",
        "check-production-ops-config-secret-boundary.py",
    ],
    "scripts/README.md": [
        "check-production-ops-config-secret-boundary.py",
        "production-ops-config-secret-boundary.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_fixture() -> dict[str, Any]:
    document = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    require(isinstance(document, dict), "production ops config secret boundary fixture must be an object")
    return document


def assert_slice(document: dict[str, Any]) -> None:
    slice_info = document.get("slice") or {}
    require(slice_info.get("id") == "config-secret-boundary", "unexpected production ops slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected production ops track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "config-secret-boundary must only satisfy the governance boundary",
    )
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    required_forbidden_claims = {
        "production_ready",
        "production_secret_backend_ready",
        "production_provider_health_ready",
        "process_supervisor_ready",
        "console_production_package_ready",
    }
    missing = sorted(required_forbidden_claims - forbidden_claims)
    require(not missing, f"missing forbidden production claims: {missing}")


def assert_config_sources(document: dict[str, Any]) -> None:
    sources = {str(item.get("id")): item for item in document.get("config_sources") or [] if isinstance(item, dict)}
    missing = sorted(set(REQUIRED_CONFIG_SOURCES) - set(sources))
    require(not missing, f"missing config sources: {missing}")

    for source_id, (environment, status, readiness_scope) in REQUIRED_CONFIG_SOURCES.items():
        item = sources[source_id]
        require(item.get("environment") == environment, f"{source_id} has unexpected environment")
        require(item.get("status") == status, f"{source_id} has unexpected status")
        require(item.get("readiness_scope") == readiness_scope, f"{source_id} has unexpected readiness scope")
        require(item.get("secret_policy"), f"{source_id} must document secret_policy")

    local_inputs = set(sources["local_defaults"].get("allowed_inputs") or [])
    require("mock provider" in local_inputs, "local defaults must explicitly include mock provider")
    env_inputs = set(sources["developer_env_override"].get("allowed_inputs") or [])
    require("RADISHMIND_PLATFORM_API_KEY" in env_inputs, "developer env override must name API key env input")
    production_inputs = list(sources["production_secret_backend"].get("allowed_inputs") or [])
    require(production_inputs == [], "production secret backend must not expose implemented inputs in this slice")


def assert_secret_boundaries(document: dict[str, Any]) -> None:
    boundary = document.get("secret_boundaries") or {}
    forbidden_fields = set(boundary.get("committed_secret_fields_forbidden") or [])
    missing_forbidden = sorted(REQUIRED_SECRET_FIELDS - forbidden_fields)
    require(not missing_forbidden, f"missing forbidden committed secret fields: {missing_forbidden}")

    sanitized_outputs = set(boundary.get("sanitized_outputs") or [])
    missing_outputs = sorted(REQUIRED_SANITIZED_OUTPUTS - sanitized_outputs)
    require(not missing_outputs, f"missing sanitized outputs: {missing_outputs}")

    sanitized_fields = set(boundary.get("required_sanitized_fields") or [])
    missing_fields = sorted(REQUIRED_SANITIZED_FIELDS - sanitized_fields)
    require(not missing_fields, f"missing sanitized fields: {missing_fields}")

    raw_values = set(boundary.get("raw_values_must_not_appear") or [])
    require("provider base URL" in raw_values, "provider base URL must stay out of sanitized outputs")


def assert_provider_boundaries(document: dict[str, Any]) -> None:
    boundaries = {
        str(item.get("id")): item
        for item in document.get("provider_profile_boundaries") or []
        if isinstance(item, dict)
    }
    missing = sorted(set(REQUIRED_PROVIDER_BOUNDARIES) - set(boundaries))
    require(not missing, f"missing provider profile boundaries: {missing}")
    for boundary_id, (credential_state, production_use) in REQUIRED_PROVIDER_BOUNDARIES.items():
        item = boundaries[boundary_id]
        require(item.get("credential_state") == credential_state, f"{boundary_id} has unexpected credential state")
        require(item.get("production_use") == production_use, f"{boundary_id} has unexpected production use")
        require(item.get("readiness_scope"), f"{boundary_id} must document readiness scope")


def assert_blocked_conditions(document: dict[str, Any]) -> None:
    blocked = {str(item.get("id")): item for item in document.get("blocked_conditions") or [] if isinstance(item, dict)}
    missing = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing, f"missing blocked production conditions: {missing}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        item = blocked[condition_id]
        require(item.get("status") == "not_satisfied", f"{condition_id} must remain not_satisfied")
        require(
            item.get("required_before_production_ready") is True,
            f"{condition_id} must gate production ready",
        )


def assert_evidence_and_consumers(document: dict[str, Any]) -> None:
    for evidence in document.get("evidence") or []:
        evidence_path = REPO_ROOT / str(evidence)
        require(evidence_path.exists(), f"missing production ops evidence path: {evidence}")

    consumers = set(document.get("required_consumers") or [])
    expected_consumers = {
        "scripts/check-production-ops-config-secret-boundary.py",
        "scripts/check-repo.py",
        "scripts/check-p3-local-product-shell-short-close-checklist.py",
        "scripts/README.md",
    }
    missing_consumers = sorted(expected_consumers - consumers)
    require(not missing_consumers, f"missing production ops consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-config-secret-boundary.py", [])' in check_repo,
        "check-repo.py must run production ops config secret boundary check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        document_text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing_literals = [literal for literal in required_literals if literal not in document_text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def main() -> int:
    document = load_fixture()
    require(document.get("schema_version") == 1, "unexpected schema_version")
    require(document.get("kind") == "production_ops_config_secret_boundary", "unexpected fixture kind")
    assert_slice(document)
    assert_config_sources(document)
    assert_secret_boundaries(document)
    assert_provider_boundaries(document)
    assert_blocked_conditions(document)
    assert_evidence_and_consumers(document)
    print("production ops config/secret boundary checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
