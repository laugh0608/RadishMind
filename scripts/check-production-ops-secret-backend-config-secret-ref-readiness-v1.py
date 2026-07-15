#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json"
IMPLEMENTATION_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
)
SECRET_REFERENCE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-secret-reference-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_FORBIDDEN_CLAIMS = {
    "production_ready",
    "production_secret_backend_ready",
    "cloud_secret_service_ready",
    "resolver_implemented",
    "fake_resolver_implemented",
    "real_secret_written",
    "provider_profile_binding_ready",
    "repository_mode_ready",
}

REQUIRED_DEPENDENCIES = {
    "config-secret-boundary": "governance_boundary_satisfied",
    "production-secret-backend-contract": "governance_boundary_satisfied",
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied",
}

REQUIRED_FUTURE_FIELDS = {
    "credential_state",
    "secret_backend_configured",
    "secret_ref_present",
    "missing_secret_refs",
    "field_sources",
}

REQUIRED_FAILURE_CODES = {
    "secret_reference_manifest_missing",
    "secret_reference_manifest_invalid",
    "secret_ref_missing",
    "secret_backend_disabled",
    "resolver_invocation_forbidden",
}

REQUIRED_NO_FALLBACK = {
    "no production fallback to RADISHMIND_PLATFORM_API_KEY",
    "no fallback to mock provider",
    "no fallback to local-smoke profile",
    "no fallback to fixture credential",
    "no fallback from missing secret_ref to committed secret value",
    "secret_ref_present does not mean credential resolved",
}

REQUIRED_NO_SIDE_EFFECTS = {
    "checker reads committed docs and fixtures only",
    "no environment secret read",
    "no cloud secret service call",
    "no provider call",
    "no database connection",
    "no file write",
    "no repository mode enablement",
}

REQUIRED_DOC_REFERENCES = {
    "docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md": [
        "config_secret_ref_readiness_defined",
        "config-secret-ref-readiness",
        "secret_backend_configured",
        "secret_ref_present",
        "missing_secret_refs",
        "resolver_implementation_status",
        "not_started",
        "production_secret_backend",
        "not_satisfied",
        "No Fallback",
        "No Side Effects",
        "Artifact Guard",
    ],
    "docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md": [
        "Production Secret Backend Config / Secret Ref Readiness",
        "config-secret-ref-readiness",
        "secret_reference_manifest_missing",
        "secret_backend_disabled",
        "resolver_invocation_forbidden",
        "不实现 resolver runtime",
        "不调用云 secret 服务",
        "不接数据库",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Config / Secret Ref Readiness v1",
        "config_secret_ref_readiness_defined",
        "production-secret-backend-config-secret-ref-readiness-v1",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Config / Secret Ref Readiness v1",
        "docs/platform/",
    ],
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-config-secret-ref-readiness-v1",
        "config_secret_ref_readiness_defined",
        "production-secret-backend-config-secret-ref-readiness-v1.json",
    ],
    "docs/radishmind-product-scope.md": [
        "production-secret-backend-config-secret-ref-readiness-v1",
        "config_secret_ref_readiness_defined",
        "不接生产后端",
    ],
    "docs/radishmind-capability-matrix.md": [
        "config-secret-ref-readiness",
        "production-secret-backend-config-secret-ref-readiness-v1.json",
    ],
    "docs/task-cards/README.md": [
        "Production Secret Backend Config / Secret Ref Readiness",
        "production-secret-backend-config-secret-ref-readiness-v1",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "config-secret-ref-readiness",
        "production-secret-backend-config-secret-ref-readiness-v1",
        "config_secret_ref_readiness_defined",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "config-secret-ref-readiness",
        "production-secret-backend-config-secret-ref-readiness-v1",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
        "production-secret-backend-config-secret-ref-readiness-v1.json",
    ],
    "deploy/README.md": [
        "production-secret-backend-config-secret-ref-readiness-v1",
        "config-secret-ref-readiness",
    ],
    "docs/devlogs/2026-W25.md": [
        "production-secret-backend-config-secret-ref-readiness-v1",
        "config_secret_ref_readiness_defined",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path} must contain a json object")
    return document


def assert_slice(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "production-secret-backend-config-secret-ref-readiness-v1", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "config_secret_ref_readiness_defined", "unexpected status")
    task_card = str(slice_info.get("task_card") or "")
    platform_topic = str(slice_info.get("platform_topic") or "")
    require(task_card == "docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md", "unexpected task card")
    require(platform_topic == "docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md", "unexpected platform topic")
    require((REPO_ROOT / task_card).exists(), "task card must exist")
    require((REPO_ROOT / platform_topic).exists(), "platform topic must exist")
    missing_claims = sorted(REQUIRED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = {str(item.get("id")): item for item in fixture.get("depends_on") or [] if isinstance(item, dict)}
    missing = sorted(set(REQUIRED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, expected_status in REQUIRED_DEPENDENCIES.items():
        dependency = dependencies[dependency_id]
        require(dependency.get("status") == expected_status, f"{dependency_id} has unexpected status")
        evidence = str(dependency.get("evidence") or "")
        require(evidence, f"{dependency_id} must cite evidence")
        require((REPO_ROOT / evidence).exists(), f"{dependency_id} evidence is missing: {evidence}")


def assert_implementation_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_boundary") or {}
    false_fields = {
        "config_runtime_fields_added",
        "resolver_interface_added",
        "resolver_runtime_added",
        "fake_resolver_added",
        "cloud_secret_sdk_added",
        "secret_values_read",
        "database_connection_provider_added",
        "repository_mode_enabled",
    }
    for field in false_fields:
        require(boundary.get(field) is False, f"{field} must stay false")
    require(boundary.get("production_secret_backend_status") == "not_satisfied", "production backend must stay blocked")
    require(boundary.get("resolver_implementation_status") == "not_started", "resolver implementation must stay not_started")
    require(boundary.get("default_runtime_state") == "disabled", "default runtime state must stay disabled")

    scope = fixture.get("readiness_scope") or {}
    require(scope.get("satisfies_precondition") == "config-injection-point", "unexpected satisfied precondition")
    require(scope.get("planned_slice_status") == "satisfied", "planned slice must be satisfied")
    blocked = set(scope.get("does_not_satisfy") or [])
    for item in {
        "provider-profile-binding",
        "secret-resolver-interface-disabled",
        "operator-runbook-and-negative-gates",
        "rotation-and-audit-policy",
        "production-secret-backend-ready",
    }:
        require(item in blocked, f"readiness scope must not satisfy {item}")


def assert_config_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("config_secret_ref_contract") or {}
    future_fields = set(contract.get("future_sanitized_secret_ref_fields") or [])
    missing_future = sorted(REQUIRED_FUTURE_FIELDS - future_fields)
    require(not missing_future, f"missing future sanitized fields: {missing_future}")

    injection_points = {
        str(item.get("id")): item
        for item in contract.get("config_injection_points") or []
        if isinstance(item, dict)
    }
    for point_id in {"config.LoadFromEnv", "config-summary", "diagnostics"}:
        require(point_id in injection_points, f"missing config injection point: {point_id}")
        must_not = set(injection_points[point_id].get("must_not") or [])
        require(must_not, f"{point_id} must define forbidden actions")
    require(
        "call resolver" in set(injection_points["config.LoadFromEnv"].get("must_not") or []),
        "config.LoadFromEnv must not call resolver",
    )
    require(
        "call cloud secret service" in set(injection_points["config.LoadFromEnv"].get("must_not") or []),
        "config.LoadFromEnv must not call cloud secret service",
    )


def assert_secret_reference_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("secret_reference_manifest_guard") or {}
    require(guard.get("kind") == "production_secret_reference_manifest", "unexpected manifest kind guard")
    require(guard.get("scope") == "secret_reference_only", "secret reference guard must stay reference-only")
    policy = guard.get("required_policy") or {}
    for field in {
        "stores_secret_values",
        "resolver_enabled",
        "cloud_calls_allowed",
        "production_secret_backend_ready",
    }:
        require(policy.get(field) is False, f"required policy {field} must be false")
    missing_fields = sorted(REQUIRED_FUTURE_FIELDS - set(guard.get("required_sanitized_fields") or []))
    require(not missing_fields, f"secret reference guard missing sanitized fields: {missing_fields}")

    manifest = load_json(SECRET_REFERENCE_PATH)
    require(manifest.get("kind") == guard.get("kind"), "secret reference fixture kind drifted")
    require(manifest.get("scope") == guard.get("scope"), "secret reference fixture scope drifted")
    manifest_policy = manifest.get("policy") or {}
    for field in policy:
        require(manifest_policy.get(field) == policy[field], f"secret reference fixture policy drifted: {field}")


def assert_failure_mapping_and_policies(fixture: dict[str, Any]) -> None:
    failure_mapping = {
        str(item.get("code")): item
        for item in fixture.get("failure_mapping") or []
        if isinstance(item, dict)
    }
    missing_codes = sorted(REQUIRED_FAILURE_CODES - set(failure_mapping))
    require(not missing_codes, f"missing failure codes: {missing_codes}")
    for code, item in failure_mapping.items():
        require(item.get("failure_boundary") == "configuration", f"{code} must use configuration boundary")
        require(item.get("retryable") is False, f"{code} must not be retryable")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic, f"{code} must define sanitized diagnostic")
        require("secret value" not in diagnostic.lower(), f"{code} diagnostic must not ask for secret value")

    missing_fallback = sorted(REQUIRED_NO_FALLBACK - set(fixture.get("no_fallback_policy") or []))
    require(not missing_fallback, f"missing no fallback policy: {missing_fallback}")
    missing_side_effects = sorted(REQUIRED_NO_SIDE_EFFECTS - set(fixture.get("no_side_effect_policy") or []))
    require(not missing_side_effects, f"missing no side effects policy: {missing_side_effects}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    forbidden = set(guard.get("forbidden_artifacts") or [])
    for item in {
        "secret resolver runtime",
        "fake resolver",
        "cloud secret SDK",
        "database connection provider",
        "DB driver",
        "connection factory",
        "SQL migration runner",
        "workflow repository mode runtime",
    }:
        require(item in forbidden, f"artifact guard missing forbidden artifact: {item}")
    forbidden_claims = set(guard.get("forbidden_claims") or [])
    for claim in {
        "production_secret_backend_ready",
        "secret_resolver_ready",
        "credential_resolved",
        "repository_mode_ready",
    }:
        require(claim in forbidden_claims, f"artifact guard missing forbidden claim: {claim}")


def assert_implementation_readiness_alignment() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    planned = {
        str(item.get("id")): item
        for item in readiness.get("planned_slices") or []
        if isinstance(item, dict)
    }
    config_slice = planned.get("config-secret-ref-readiness") or {}
    require(config_slice.get("status") == "satisfied", "config-secret-ref-readiness must be marked satisfied")
    evidence = set(config_slice.get("evidence") or [])
    for path in {
        "docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md",
        "docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json",
        "scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
    }:
        require(path in evidence, f"config-secret-ref-readiness missing evidence: {path}")
        require((REPO_ROOT / path).exists(), f"config-secret-ref-readiness evidence missing: {path}")

    preconditions = {
        str(item.get("id")): item
        for item in readiness.get("required_preconditions") or []
        if isinstance(item, dict)
    }
    config_precondition = preconditions.get("config-injection-point") or {}
    require(config_precondition.get("status") == "satisfied", "config-injection-point must be satisfied")
    precondition_evidence = set(config_precondition.get("evidence") or [])
    require(
        "scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json"
        in precondition_evidence,
        "config-injection-point must cite config secret ref readiness fixture",
    )


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_doc_references") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required doc reference missing: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    implementation_call = 'run_python_script("check-production-ops-secret-backend-implementation-readiness.py", [])'
    reference_call = 'run_python_script("check-production-secret-reference-contract.py", [])'
    config_ref_call = 'run_python_script("check-production-ops-secret-backend-config-secret-ref-readiness-v1.py", [])'
    require(implementation_call in check_repo, "check-repo.py must run implementation readiness checker")
    require(reference_call in check_repo, "check-repo.py must run secret reference contract checker")
    require(config_ref_call in check_repo, "check-repo.py must run config secret ref readiness checker")
    require(
        check_repo.index(implementation_call) < check_repo.index(config_ref_call),
        "config secret ref readiness checker must run after implementation readiness",
    )
    require(
        check_repo.index(reference_call) < check_repo.index(config_ref_call),
        "config secret ref readiness checker must run after secret reference contract",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_validation_strategy(fixture: dict[str, Any]) -> None:
    strategy = set(fixture.get("validation_strategy") or [])
    for item in {
        "run config secret ref readiness checker",
        "run production secret backend implementation readiness checker",
        "run production secret reference contract checker",
        "run production secret backend contract checker",
        "run production config secret boundary checker",
        "run fast repository check",
        "run full repository check",
    }:
        require(item in strategy, f"validation strategy missing: {item}")


def assert_no_secret_literals() -> None:
    text = "\n".join(
        [
            FIXTURE_PATH.read_text(encoding="utf-8"),
            read("docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md"),
            read("docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md"),
        ]
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:", "token:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"config secret ref readiness contains forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", text) is None,
        "config secret ref readiness contains forbidden sk-like token",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_config_secret_ref_readiness_v1",
        "unexpected fixture kind",
    )
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_implementation_boundary(fixture)
    assert_config_contract(fixture)
    assert_secret_reference_guard(fixture)
    assert_failure_mapping_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment()
    assert_docs_and_check_repo(fixture)
    assert_validation_strategy(fixture)
    assert_no_secret_literals()
    print("production ops secret backend config secret ref readiness checks passed.")


if __name__ == "__main__":
    main()
