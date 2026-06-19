#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json"
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
    "rotation_ready",
    "production_secret_audit_store_ready",
}

REQUIRED_DEPENDENCIES = {
    "config-secret-boundary": "governance_boundary_satisfied",
    "production-secret-backend-contract": "governance_boundary_satisfied",
    "production-secret-backend-implementation-readiness": "implementation_readiness_defined",
    "secret-ref-schema-and-fixtures": "satisfied",
    "config-secret-ref-readiness": "satisfied",
    "provider-profile-secret-binding": "satisfied",
    "secret-resolver-interface-disabled": "satisfied",
}

REQUIRED_RUNBOOK_SECTIONS = {
    "purpose_and_scope",
    "environment_selection",
    "secret_source_inventory",
    "operator_approval_evidence",
    "sanitized_verification_steps",
    "negative_gate_review",
    "smoke_record_template",
    "rollback_or_disable_procedure",
    "production_ready_stop_line",
}

REQUIRED_EVIDENCE_FIELDS = {
    "environment",
    "provider",
    "provider_profile",
    "secret_ref_status",
    "secret_backend_configured",
    "resolver_state",
    "operator_id_ref",
    "approval_ticket_ref",
    "request_id",
    "audit_ref",
    "runbook_version",
    "verification_status",
    "smoke_record_ref",
    "negative_gate_results",
    "timestamp",
}

REQUIRED_DIAGNOSTIC_FIELDS = {
    "credential_state",
    "secret_backend_configured",
    "secret_ref_present",
    "secret_ref_status",
    "resolver_state",
    "operator_gate_status",
    "negative_gate_results",
    "failure_code",
    "sanitized_diagnostic",
    "field_sources",
    "smoke_record_ref",
    "audit_ref",
}

REQUIRED_NEGATIVE_GATE_CODES = {
    "operator_runbook_missing",
    "operator_approval_missing",
    "operator_environment_mismatch",
    "operator_secret_source_missing",
    "operator_sanitized_verification_missing",
    "operator_smoke_record_missing",
    "operator_resolver_invocation_disabled",
    "operator_secret_value_exposure_detected",
    "operator_fallback_forbidden",
    "operator_production_ready_claim_forbidden",
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
    "operator runbook does not mean credential resolved",
    "negative gate evidence does not mean production secret backend ready",
}

REQUIRED_NO_SIDE_EFFECTS = {
    "checker reads committed docs and fixtures only",
    "no environment secret read",
    "no cloud secret service call",
    "no provider call",
    "no database connection",
    "no credential handle creation",
    "no runbook execution",
    "no negative gate runtime call",
    "no file write",
    "no repository mode enablement",
}

REQUIRED_DOC_REFERENCES = {
    "docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md": [
        "operator_runbook_negative_gates_readiness_defined",
        "operator-runbook-and-negative-gates",
        "Operator Runbook Contract",
        "Negative Gates",
        "operator_runbook_missing",
        "operator_production_ready_claim_forbidden",
        "No Fallback",
        "No Side Effects",
        "Artifact Guard",
    ],
    "docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md": [
        "Production Secret Backend Operator Runbook / Negative Gates Readiness",
        "operator-runbook-and-negative-gates",
        "operator_runbook_missing",
        "不实现 resolver runtime",
        "不调用云 secret 服务",
        "不接数据库",
    ],
    "docs/platform/README.md": [
        "Production Secret Backend Operator Runbook / Negative Gates Readiness v1",
        "operator_runbook_negative_gates_readiness_defined",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
    ],
    "docs/features/README.md": [
        "Production Secret Backend Operator Runbook / Negative Gates Readiness v1",
        "docs/platform/",
    ],
    "docs/features/workflow/README.md": [
        "Production Secret Backend Operator Runbook / Negative Gates Readiness v1",
        "operator_runbook_negative_gates_readiness_defined",
    ],
    "docs/features/workflow/saved-workflow-draft-v1.md": [
        "Production Secret Backend Operator Runbook / Negative Gates Readiness v1",
        "operator_runbook_negative_gates_readiness_defined",
    ],
    "docs/features/workflow-agent-runtime.md": [
        "Production Secret Backend Operator Runbook / Negative Gates Readiness v1",
        "operator_runbook_negative_gates_readiness_defined",
    ],
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
        "operator_runbook_negative_gates_readiness_defined",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
    ],
    "docs/radishmind-product-scope.md": [
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
        "operator_runbook_negative_gates_readiness_defined",
        "不接生产后端",
    ],
    "docs/radishmind-roadmap.md": [
        "operator-runbook-and-negative-gates",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
    ],
    "docs/radishmind-capability-matrix.md": [
        "operator-runbook-and-negative-gates",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
    ],
    "docs/task-cards/README.md": [
        "Production Secret Backend Operator Runbook / Negative Gates Readiness",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
    ],
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "operator-runbook-and-negative-gates",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
        "operator_runbook_negative_gates_readiness_defined",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "operator-runbook-and-negative-gates",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
    ],
    "services/platform/README.md": [
        "Production secret backend operator runbook / negative gates readiness",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
    ],
    "deploy/README.md": [
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
        "operator-runbook-and-negative-gates",
    ],
    "docs/devlogs/2026-W25.md": [
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
        "operator_runbook_negative_gates_readiness_defined",
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
        slice_info.get("id") == "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "operator_runbook_negative_gates_readiness_defined", "unexpected status")
    task_card = str(slice_info.get("task_card") or "")
    platform_topic = str(slice_info.get("platform_topic") or "")
    require(
        task_card == "docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md",
        "unexpected task card",
    )
    require(
        platform_topic == "docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md",
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
        "operator_runbook_runtime_added",
        "negative_gate_runtime_added",
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
    require(scope.get("satisfies_precondition") == "operator-runbook", "unexpected satisfied precondition")
    require(
        scope.get("satisfies_planned_slice") == "operator-runbook-and-negative-gates",
        "unexpected planned slice",
    )
    require(scope.get("planned_slice_status") == "satisfied", "planned slice must be satisfied")
    blocked = set(scope.get("does_not_satisfy") or [])
    for item in {
        "test-fixture-strategy",
        "rotation-and-audit-policy",
        "production-secret-backend-ready",
        "credential-resolution-ready",
        "resolver-implementation-ready",
    }:
        require(item in blocked, f"readiness scope must not satisfy {item}")


def assert_operator_runbook_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("operator_runbook_contract") or {}
    missing_sections = sorted(REQUIRED_RUNBOOK_SECTIONS - set(contract.get("required_sections") or []))
    require(not missing_sections, f"missing runbook sections: {missing_sections}")
    missing_fields = sorted(REQUIRED_EVIDENCE_FIELDS - set(contract.get("allowed_evidence_fields") or []))
    require(not missing_fields, f"missing allowed evidence fields: {missing_fields}")
    forbidden = set(contract.get("forbidden_evidence_fields") or [])
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
        require(field in forbidden, f"missing forbidden evidence field: {field}")
    for flag in {
        "test_secret_source_required",
        "production_secret_source_required",
        "sanitized_verification_required",
        "smoke_record_required",
        "operator_approval_required",
    }:
        require(contract.get(flag) is True, f"{flag} must be true")


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


def assert_negative_gates_and_policies(fixture: dict[str, Any]) -> None:
    gates = {str(item.get("code")): item for item in fixture.get("negative_gates") or [] if isinstance(item, dict)}
    missing_codes = sorted(REQUIRED_NEGATIVE_GATE_CODES - set(gates))
    require(not missing_codes, f"missing negative gate codes: {missing_codes}")
    for code, item in gates.items():
        require(item.get("failure_boundary") == "operator_gate", f"{code} must use operator_gate boundary")
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
        "operator runbook executor",
        "negative gate runtime",
        "cloud secret SDK",
        "opaque credential handle runtime",
        "database connection provider",
        "DB driver",
        "connection factory",
        "SQL migration runner",
        "workflow repository mode runtime",
        "public production API",
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
        "production_ready",
    }:
        require(claim in forbidden_claims, f"artifact guard missing forbidden claim: {claim}")


def assert_implementation_readiness_alignment() -> None:
    readiness = load_json(IMPLEMENTATION_READINESS_PATH)
    planned = {
        str(item.get("id")): item
        for item in readiness.get("planned_slices") or []
        if isinstance(item, dict)
    }
    operator_slice = planned.get("operator-runbook-and-negative-gates") or {}
    require(operator_slice.get("status") == "satisfied", "operator-runbook-and-negative-gates must be satisfied")
    evidence = set(operator_slice.get("evidence") or [])
    for path in {
        "docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md",
        "docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
        "scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
    }:
        require(path in evidence, f"operator slice missing evidence: {path}")
        require((REPO_ROOT / path).exists(), f"operator slice evidence missing: {path}")

    preconditions = {
        str(item.get("id")): item
        for item in readiness.get("required_preconditions") or []
        if isinstance(item, dict)
    }
    operator = preconditions.get("operator-runbook") or {}
    require(operator.get("status") == "satisfied", "operator-runbook precondition must be satisfied")
    operator_evidence = set(operator.get("evidence") or [])
    require(
        "scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json"
        in operator_evidence,
        "operator-runbook must cite operator negative gates fixture",
    )
    rotation = preconditions.get("rotation-and-audit-policy") or {}
    require(
        rotation.get("status") == "required_before_production_secret_backend_ready",
        "rotation-and-audit-policy must stay blocked",
    )
    test_fixture = preconditions.get("test-fixture-strategy") or {}
    require(
        test_fixture.get("status") == "required_before_implementation",
        "test-fixture-strategy must stay blocked",
    )

    blocked = {
        str(item.get("id")): item
        for item in readiness.get("blocked_conditions") or []
        if isinstance(item, dict)
    }
    for blocked_id, expected_status in {
        "production_secret_backend": "not_satisfied",
        "cloud_secret_service_integration": "not_satisfied",
        "real_secret_values": "forbidden_in_committed_repo",
        "production_ready": "not_satisfied",
    }.items():
        require(blocked.get(blocked_id, {}).get("status") == expected_status, f"{blocked_id} status drifted")


def assert_docs_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_doc_references") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"required doc reference missing: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    disabled_call = (
        'run_python_script("check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py", [])'
    )
    operator_call = (
        'run_python_script("check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py", [])'
    )
    for call, description in {
        disabled_call: "secret resolver interface disabled readiness checker",
        operator_call: "operator runbook negative gates readiness checker",
    }.items():
        require(call in check_repo, f"check-repo.py must run {description}")
    require(
        check_repo.index(disabled_call) < check_repo.index(operator_call),
        "operator runbook checker must run after disabled resolver checker",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_validation_strategy(fixture: dict[str, Any]) -> None:
    strategy = set(fixture.get("validation_strategy") or [])
    for item in {
        "run operator runbook negative gates readiness checker",
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
            read("docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md"),
            read("docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md"),
        ]
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA", "authorization:", "cookie:"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"operator runbook readiness contains forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", text) is None,
        "operator runbook readiness contains forbidden sk-like token",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_operator_runbook_negative_gates_readiness_v1",
        "unexpected fixture kind",
    )
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_implementation_boundary(fixture)
    assert_operator_runbook_contract(fixture)
    assert_secret_reference_alignment()
    assert_sanitized_diagnostics(fixture)
    assert_negative_gates_and_policies(fixture)
    assert_artifact_guard(fixture)
    assert_implementation_readiness_alignment()
    assert_docs_and_check_repo(fixture)
    assert_validation_strategy(fixture)
    assert_no_secret_literals()
    print("production ops secret backend operator runbook negative gates readiness checks passed.")


if __name__ == "__main__":
    main()
