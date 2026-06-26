#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json"
)
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
    "credential_resolved",
    "secret_resolver_ready",
    "credential_handle_created",
    "repository_mode_ready",
}

REQUIRED_DEPENDENCIES = {
    "config-secret-boundary": "governance_boundary_satisfied",
    "production-secret-backend-contract": "governance_boundary_satisfied",
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied",
    "config-secret-ref-readiness": "satisfied",
    "provider-profile-secret-binding": "satisfied",
}

REQUIRED_INPUT_FIELDS = {
    "environment",
    "provider",
    "provider_profile",
    "credential_requirement",
    "secret_ref_status",
    "secret_backend_configured",
    "request_id",
    "audit_ref",
    "caller_purpose",
}

REQUIRED_RESULT_FIELDS = {
    "resolver_state",
    "secret_backend_configured",
    "resolver_available",
    "credential_handle_present",
    "failure_code",
    "sanitized_diagnostic",
    "retryable",
}

REQUIRED_DIAGNOSTIC_FIELDS = {
    "credential_state",
    "secret_backend_configured",
    "secret_ref_present",
    "secret_ref_status",
    "resolver_state",
    "resolver_available",
    "credential_handle_present",
    "failure_code",
    "sanitized_diagnostic",
    "field_sources",
}

REQUIRED_FAILURE_CODES = {
    "secret_resolver_secret_ref_missing",
    "secret_resolver_backend_disabled",
    "secret_resolver_unavailable",
    "secret_resolution_denied",
    "secret_resolver_environment_mismatch",
    "secret_resolver_invocation_disabled",
}

REQUIRED_NO_FALLBACK = {
    "no production fallback to RADISHMIND_PLATFORM_API_KEY",
    "no fallback to mock provider",
    "no fallback to local-smoke profile",
    "no fallback to fixture credential",
    "no fallback from missing secret_ref to committed secret value",
    "no cross-environment secret_ref fallback",
    "no fallback to fake resolver",
    "no fallback to fake query executor",
    "disabled resolver result does not mean credential resolved",
}

REQUIRED_NO_SIDE_EFFECTS = {
    "checker reads committed docs and fixtures only",
    "no environment secret read",
    "no cloud secret service call",
    "no provider call",
    "no database connection",
    "no credential handle creation",
    "no file write",
    "no repository mode enablement",
}

REQUIRED_DOC_REFERENCES = {
    "docs/platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md": [
        "secret_resolver_interface_disabled_readiness_defined",
        "secret-resolver-interface-disabled",
        "Disabled Resolver Interface Contract",
        "secret_resolver_backend_disabled",
        "credential_handle_present",
        "No Fallback",
        "No Side Effects",
        "Artifact Guard",
    ],
    "docs/task-cards/production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md": [
        "Production Secret Backend Secret Resolver Interface Disabled Readiness",
        "secret-resolver-interface-disabled",
        "secret_resolver_backend_disabled",
        "不实现 resolver runtime",
        "不调用云 secret 服务",
        "不接数据库",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Secret Resolver Interface Disabled Readiness v1",
        "secret_resolver_interface_disabled_readiness_defined",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Secret Resolver Interface Disabled Readiness v1",
        "docs/platform/",
    ],
    "docs/features/workflow/README.md": [
        "Production Secret Backend Secret Resolver Interface Disabled Readiness v1",
        "secret_resolver_interface_disabled_readiness_defined",
    ],
    "docs/features/workflow/saved-workflow-draft-v1.md": [
        "Production Secret Backend Secret Resolver Interface Disabled Readiness v1",
        "secret_resolver_interface_disabled_readiness_defined",
    ],
    "docs/features/workflow-agent-runtime.md": [
        "Production Secret Backend Secret Resolver Interface Disabled Readiness v1",
        "secret_resolver_interface_disabled_readiness_defined",
    ],
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
        "secret_resolver_interface_disabled_readiness_defined",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
    ],
    "docs/radishmind-product-scope.md": [
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
        "secret_resolver_interface_disabled_readiness_defined",
        "不接生产后端",
    ],
    "docs/radishmind-roadmap.md": [
        "secret-resolver-interface-disabled",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
    ],
    "docs/radishmind-capability-matrix.md": [
        "secret-resolver-interface-disabled",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
    ],
    "docs/task-cards/README.md": [
        "Production Secret Backend Secret Resolver Interface Disabled Readiness",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "secret-resolver-interface-disabled",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
        "secret_resolver_interface_disabled_readiness_defined",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "secret-resolver-interface-disabled",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
    ],
    "services/platform/README.md": [
        "Production secret backend secret resolver interface disabled readiness",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
    ],
    "deploy/README.md": [
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
        "secret-resolver-interface-disabled",
    ],
    "docs/devlogs/2026-W25.md": [
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
        "secret_resolver_interface_disabled_readiness_defined",
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
    require(
        slice_info.get("id") == "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "secret_resolver_interface_disabled_readiness_defined", "unexpected status")
    task_card = str(slice_info.get("task_card") or "")
    platform_topic = str(slice_info.get("platform_topic") or "")
    require(
        task_card == "docs/task-cards/production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md",
        "unexpected task card",
    )
    require(
        platform_topic == "docs/platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md",
        "unexpected platform topic",
    )
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
        "resolver_interface_runtime_added",
        "resolver_runtime_added",
        "fake_resolver_added",
        "cloud_secret_sdk_added",
        "secret_values_read",
        "credential_handle_created",
        "provider_credential_runtime_binding_added",
        "database_connection_provider_added",
        "repository_mode_enabled",
    }
    for field in false_fields:
        require(boundary.get(field) is False, f"{field} must stay false")
    require(boundary.get("production_secret_backend_status") == "not_satisfied", "production backend must stay blocked")
    require(boundary.get("resolver_implementation_status") == "not_started", "resolver implementation must stay not_started")
    require(boundary.get("resolver_runtime_status") == "not_created", "resolver runtime must stay not_created")
    require(boundary.get("default_runtime_state") == "disabled", "default runtime state must stay disabled")

    scope = fixture.get("readiness_scope") or {}
    require(scope.get("satisfies_precondition") == "secret-resolver-interface-disabled", "unexpected satisfied precondition")
    require(scope.get("planned_slice_status") == "satisfied", "planned slice must be satisfied")
    for item in {"sanitized-audit-fields", "failure-taxonomy"}:
        require(item in set(scope.get("also_satisfies_preconditions") or []), f"scope must satisfy {item}")
    blocked = set(scope.get("does_not_satisfy") or [])
    for item in {
        "test-fixture-strategy",
        "operator-runbook-and-negative-gates",
        "rotation-and-audit-policy",
        "production-secret-backend-ready",
        "credential-resolution-ready",
    }:
        require(item in blocked, f"readiness scope must not satisfy {item}")


def assert_disabled_interface_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("disabled_resolver_interface_contract") or {}
    missing_inputs = sorted(REQUIRED_INPUT_FIELDS - set(contract.get("input_fields") or []))
    require(not missing_inputs, f"missing resolver input fields: {missing_inputs}")
    missing_results = sorted(REQUIRED_RESULT_FIELDS - set(contract.get("result_fields") or []))
    require(not missing_results, f"missing resolver result fields: {missing_results}")

    disabled = contract.get("disabled_result") or {}
    require(disabled.get("resolver_state") == "disabled", "disabled resolver state must be disabled")
    require(disabled.get("secret_backend_configured") is False, "disabled result must not configure backend")
    require(disabled.get("resolver_available") is False, "disabled result must not expose available resolver")
    require(disabled.get("credential_handle_present") is False, "disabled result must not expose credential handle")
    require(disabled.get("retryable") is False, "disabled result must not be retryable")

    forbidden_inputs = set(contract.get("forbidden_input_fields") or [])
    for field in {"secret_value", "api_key", "token", "password", "provider_raw_url", "dsn"}:
        require(field in forbidden_inputs, f"missing forbidden resolver input: {field}")
    forbidden_results = set(contract.get("forbidden_result_fields") or [])
    for field in {"secret_value", "api_key", "token", "password", "credential_handle", "cloud_credential"}:
        require(field in forbidden_results, f"missing forbidden resolver result: {field}")


def assert_secret_reference_alignment() -> None:
    manifest = load_json(SECRET_REFERENCE_PATH)
    require(manifest.get("kind") == "production_secret_reference_manifest", "unexpected secret reference kind")
    require(manifest.get("scope") == "secret_reference_only", "secret reference scope must stay reference-only")
    policy = manifest.get("policy") or {}
    for field in {"stores_secret_values", "resolver_enabled", "cloud_calls_allowed", "production_secret_backend_ready"}:
        require(policy.get(field) is False, f"secret reference policy {field} must be false")


def assert_sanitized_diagnostics(fixture: dict[str, Any]) -> None:
    diagnostics = fixture.get("sanitized_diagnostics") or {}
    missing_allowed = sorted(REQUIRED_DIAGNOSTIC_FIELDS - set(diagnostics.get("allowed_fields") or []))
    require(not missing_allowed, f"missing sanitized diagnostic fields: {missing_allowed}")
    forbidden = set(diagnostics.get("forbidden_fields") or [])
    for field in {
        "raw_secret",
        "password",
        "token",
        "api_key",
        "provider_raw_url",
        "dsn",
        "cloud_credential",
        "opaque_credential_handle",
        "resolver_backend_url",
    }:
        require(field in forbidden, f"missing forbidden diagnostic field: {field}")
    require(
        diagnostics.get("secret_ref_value_allowed_in_runtime_diagnostics") is False,
        "runtime diagnostics must not expose secret_ref value",
    )


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
        lower = diagnostic.lower()
        for forbidden in {"secret value", "api key", "token", "password", "credential handle"}:
            require(forbidden not in lower, f"{code} diagnostic must not mention {forbidden}")

    missing_fallback = sorted(REQUIRED_NO_FALLBACK - set(fixture.get("no_fallback_policy") or []))
    require(not missing_fallback, f"missing no fallback policy: {missing_fallback}")
    missing_side_effects = sorted(REQUIRED_NO_SIDE_EFFECTS - set(fixture.get("no_side_effect_policy") or []))
    require(not missing_side_effects, f"missing no side effects policy: {missing_side_effects}")
    counters = fixture.get("side_effect_counters") or {}
    for counter, value in counters.items():
        require(value == 0, f"{counter} must stay zero")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("artifact_guard") or {}
    forbidden = set(guard.get("forbidden_artifacts") or [])
    for item in {
        "secret resolver runtime",
        "fake resolver",
        "cloud secret SDK",
        "opaque credential handle runtime",
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
        "resolver_implemented",
        "credential_resolved",
        "credential_handle_created",
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
    disabled_slice = planned.get("secret-resolver-interface-disabled") or {}
    require(disabled_slice.get("status") == "satisfied", "secret-resolver-interface-disabled must be marked satisfied")
    evidence = set(disabled_slice.get("evidence") or [])
    for path in {
        "docs/platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md",
        "docs/task-cards/production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
        "scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
    }:
        require(path in evidence, f"secret-resolver-interface-disabled missing evidence: {path}")
        require((REPO_ROOT / path).exists(), f"secret-resolver-interface-disabled evidence missing: {path}")

    preconditions = {
        str(item.get("id")): item
        for item in readiness.get("required_preconditions") or []
        if isinstance(item, dict)
    }
    for precondition_id in {"sanitized-audit-fields", "failure-taxonomy"}:
        precondition = preconditions.get(precondition_id) or {}
        require(precondition.get("status") == "satisfied", f"{precondition_id} must be satisfied")
        precondition_evidence = set(precondition.get("evidence") or [])
        require(
            "scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json"
            in precondition_evidence,
            f"{precondition_id} must cite disabled resolver interface fixture",
        )


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_doc_references") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required doc reference missing: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    implementation_call = 'run_python_script("check-production-ops-secret-backend-implementation-readiness.py", [])'
    reference_call = 'run_python_script("check-production-secret-reference-contract.py", [])'
    config_ref_call = 'run_python_script("check-production-ops-secret-backend-config-secret-ref-readiness-v1.py", [])'
    binding_call = (
        'run_python_script("check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py", [])'
    )
    disabled_call = (
        'run_python_script("check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py", [])'
    )
    for call, description in {
        implementation_call: "implementation readiness checker",
        reference_call: "secret reference contract checker",
        config_ref_call: "config secret ref readiness checker",
        binding_call: "provider profile secret binding readiness checker",
        disabled_call: "secret resolver interface disabled readiness checker",
    }.items():
        require(call in check_repo, f"check-repo.py must run {description}")
    require(
        check_repo.index(implementation_call) < check_repo.index(disabled_call),
        "disabled resolver checker must run after implementation readiness",
    )
    require(
        check_repo.index(reference_call) < check_repo.index(disabled_call),
        "disabled resolver checker must run after secret reference contract",
    )
    require(
        check_repo.index(config_ref_call) < check_repo.index(disabled_call),
        "disabled resolver checker must run after config secret ref readiness",
    )
    require(
        check_repo.index(binding_call) < check_repo.index(disabled_call),
        "disabled resolver checker must run after provider profile binding readiness",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_validation_strategy(fixture: dict[str, Any]) -> None:
    strategy = set(fixture.get("validation_strategy") or [])
    for item in {
        "run secret resolver interface disabled readiness checker",
        "run provider profile secret binding readiness checker",
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
            read("docs/platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md"),
            read("docs/task-cards/production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md"),
        ]
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:", "token:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"secret resolver interface disabled readiness contains forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", text) is None,
        "secret resolver interface disabled readiness contains forbidden sk-like token",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_secret_resolver_interface_disabled_readiness_v1",
        "unexpected fixture kind",
    )
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_implementation_boundary(fixture)
    assert_disabled_interface_contract(fixture)
    assert_secret_reference_alignment()
    assert_sanitized_diagnostics(fixture)
    assert_failure_mapping_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment()
    assert_docs_and_check_repo(fixture)
    assert_validation_strategy(fixture)
    assert_no_secret_literals()
    print("production ops secret backend secret resolver interface disabled readiness checks passed.")


if __name__ == "__main__":
    main()
