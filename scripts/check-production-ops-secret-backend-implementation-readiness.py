#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_FORBIDDEN_CLAIMS = {
    "production_ready",
    "production_secret_backend_ready",
    "cloud_secret_service_ready",
    "real_secret_written",
    "resolver_implemented",
    "secret_rotation_ready",
    "production_secret_audit_store_ready",
}

REQUIRED_PRECONDITIONS = {
    "secret-ref-schema",
    "config-injection-point",
    "provider-profile-binding",
    "sanitized-audit-fields",
    "failure-taxonomy",
    "test-fixture-strategy",
    "operator-runbook",
    "rotation-and-audit-policy",
}

REQUIRED_PLANNED_SLICES = {
    "secret-ref-schema-and-fixtures": "satisfied",
    "config-secret-ref-readiness": "satisfied",
    "provider-profile-secret-binding": "planned_not_started",
    "secret-resolver-interface-disabled": "planned_not_started",
    "operator-runbook-and-negative-gates": "planned_not_started",
}

REQUIRED_BLOCKED = {
    "production_secret_backend": "not_satisfied",
    "cloud_secret_service_integration": "not_satisfied",
    "real_secret_values": "forbidden_in_committed_repo",
    "production_ready": "not_satisfied",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-secret-backend-implementation-v1-plan.md": [
        "production-secret-backend-contract",
        "secret-ref-schema",
        "secret-ref-schema-and-fixtures",
        "config-secret-ref-readiness",
        "production-secret-backend-config-secret-ref-readiness-v1",
        "config_secret_ref_readiness_defined",
        "contracts/production-secret-reference.schema.json",
        "production-secret-reference-basic.json",
        "check-production-secret-reference-contract.py",
        "config-injection-point",
        "provider-profile-binding",
        "sanitized-audit-fields",
        "failure-taxonomy",
        "test-fixture-strategy",
        "operator-runbook",
        "rotation-and-audit-policy",
        "不直接写 resolver 代码",
        "不接真实云 secret 服务",
        "不写入真实 secret",
        "不声明 production ready",
    ],
    "docs/radishmind-current-focus.md": [
        "production-secret-backend-implementation-readiness",
        "production-ops-secret-backend-implementation-readiness.json",
        "production-secret-backend-implementation-v1-plan.md",
    ],
    "docs/radishmind-roadmap.md": [
        "production-secret-backend-implementation-readiness",
        "production-ops-secret-backend-implementation-readiness.json",
    ],
    "docs/radishmind-capability-matrix.md": [
        "production secret backend implementation readiness",
        "production-ops-secret-backend-implementation-readiness.json",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "Production Secret Backend Implementation",
        "production-secret-backend-implementation-readiness",
        "production-ops-secret-backend-implementation-readiness.json",
    ],
    "services/platform/README.md": [
        "Production secret backend implementation readiness",
        "production-ops-secret-backend-implementation-readiness.json",
    ],
    "scripts/README.md": [
        "check-production-ops-secret-backend-implementation-readiness.py",
        "production-ops-secret-backend-implementation-readiness.json",
        "check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
    ],
    "docs/devlogs/2026-W22.md": [
        "production-secret-backend-implementation-readiness",
        "production-ops-secret-backend-implementation-readiness.json",
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
    require(slice_info.get("id") == "production-secret-backend-implementation-readiness", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected track")
    require(slice_info.get("status") == "implementation_readiness_defined", "unexpected readiness status")
    task_card = str(slice_info.get("task_card") or "")
    require(task_card == "docs/task-cards/production-secret-backend-implementation-v1-plan.md", "unexpected task card")
    require((REPO_ROOT / task_card).exists(), "secret backend implementation task card must exist")
    missing_claims = sorted(REQUIRED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_implementation_target(fixture: dict[str, Any]) -> None:
    target = fixture.get("implementation_target") or {}
    require(target.get("first_backend") == "external_reference_resolver_adapter", "unexpected first backend")
    require(target.get("cloud_vendor_specific_backend") == "not_selected", "cloud vendor backend must not be selected")
    require(target.get("committed_secret_storage") == "forbidden", "committed secret storage must be forbidden")
    require(target.get("resolver_implementation_status") == "not_started", "resolver must not be implemented")
    require(
        target.get("default_runtime_state") == "disabled_until_explicit_secret_backend_task",
        "default runtime state must remain disabled",
    )
    require(
        target.get("fast_baseline_policy") == "offline_no_real_credentials_no_cloud_calls",
        "fast baseline must stay offline without credentials or cloud calls",
    )


def assert_preconditions(fixture: dict[str, Any]) -> None:
    preconditions = {
        str(item.get("id")): item
        for item in fixture.get("required_preconditions") or []
        if isinstance(item, dict)
    }
    missing = sorted(REQUIRED_PRECONDITIONS - set(preconditions))
    require(not missing, f"missing required preconditions: {missing}")
    for precondition_id, item in preconditions.items():
        status = str(item.get("status") or "")
        require(status == "satisfied" or status.startswith("required_before_"), f"{precondition_id} has unexpected status")
        must_define = item.get("must_define") or []
        require(len(must_define) >= 3, f"{precondition_id} must define concrete requirements")
        if precondition_id == "secret-ref-schema":
            require(status == "satisfied", "secret-ref-schema precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "contracts/production-secret-reference.schema.json",
                "scripts/checks/fixtures/production-secret-reference-basic.json",
                "scripts/check-production-secret-reference-contract.py",
            }:
                require(path in evidence, f"secret-ref-schema missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"secret-ref-schema evidence missing on disk: {path}")
        if precondition_id == "config-injection-point":
            require(status == "satisfied", "config-injection-point precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md",
                "docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
            }:
                require(path in evidence, f"config-injection-point missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"config-injection-point evidence missing on disk: {path}")

    sanitized_fields = set(preconditions["sanitized-audit-fields"].get("must_define") or [])
    for field in {
        "credential_state",
        "secret_backend_configured",
        "secret_ref_present",
        "missing_secret_refs",
        "field_sources",
    }:
        require(field in sanitized_fields, f"sanitized audit fields missing {field}")


def assert_planned_slices_and_blocks(fixture: dict[str, Any]) -> None:
    planned = {str(item.get("id")): item for item in fixture.get("planned_slices") or [] if isinstance(item, dict)}
    missing_planned = sorted(set(REQUIRED_PLANNED_SLICES) - set(planned))
    require(not missing_planned, f"missing planned slices: {missing_planned}")
    for slice_id, expected_status in REQUIRED_PLANNED_SLICES.items():
        require(planned[slice_id].get("status") == expected_status, f"{slice_id} has unexpected status")
        if slice_id == "secret-ref-schema-and-fixtures":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "contracts/production-secret-reference.schema.json",
                "scripts/checks/fixtures/production-secret-reference-basic.json",
                "scripts/check-production-secret-reference-contract.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "config-secret-ref-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-config-secret-ref-readiness-v1.md",
                "docs/task-cards/production-secret-backend-config-secret-ref-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-config-secret-ref-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")

    blocked = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or [] if isinstance(item, dict)}
    missing_blocked = sorted(set(REQUIRED_BLOCKED) - set(blocked))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    for blocked_id, expected_status in REQUIRED_BLOCKED.items():
        require(blocked[blocked_id].get("status") == expected_status, f"{blocked_id} has unexpected status")


def assert_validation_and_docs(fixture: dict[str, Any]) -> None:
    validation = set(fixture.get("validation_strategy") or [])
    for required in {
        "fast baseline stays offline",
        "no real credential required",
        "no cloud API call",
        "no raw secret or provider base URL in logs, diagnostics, fixtures or docs",
    }:
        require(required in validation, f"validation strategy missing: {required}")

    for evidence in fixture.get("evidence") or []:
        require((REPO_ROOT / str(evidence)).exists(), f"missing evidence path: {evidence}")

    expected_consumers = {
        "scripts/check-production-ops-secret-backend-implementation-readiness.py",
        "scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
        "scripts/check-production-secret-reference-contract.py",
        "scripts/check-repo.py",
        "scripts/README.md",
        "docs/radishmind-current-focus.md",
        "docs/radishmind-roadmap.md",
        "docs/radishmind-capability-matrix.md",
        "docs/task-cards/production-ops-hardening-v1-plan.md",
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md",
        "services/platform/README.md",
        "docs/devlogs/2026-W22.md",
    }
    missing_consumers = sorted(expected_consumers - set(fixture.get("required_consumers") or []))
    require(not missing_consumers, f"missing consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-secret-backend-implementation-readiness.py", [])' in check_repo,
        "check-repo.py must run secret backend implementation readiness check",
    )
    require(
        'run_python_script("check-production-secret-reference-contract.py", [])' in check_repo,
        "check-repo.py must run production secret reference contract check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def assert_no_secret_literals() -> None:
    text = FIXTURE_PATH.read_text(encoding="utf-8") + "\n" + read(
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md"
    )
    forbidden_literals = ["Bearer ", "BEGIN PRIVATE KEY", "AKIA"]
    found = [literal for literal in forbidden_literals if literal in text]
    require(not found, f"secret backend readiness docs contain forbidden secret-looking literals: {found}")
    require(
        re.search(r"sk-[A-Za-z0-9]{8,}", text) is None,
        "secret backend readiness docs contain forbidden secret-looking sk token",
    )


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "production_ops_secret_backend_implementation_readiness",
        "unexpected fixture kind",
    )
    assert_slice(fixture)
    assert_implementation_target(fixture)
    assert_preconditions(fixture)
    assert_planned_slices_and_blocks(fixture)
    assert_validation_and_docs(fixture)
    assert_no_secret_literals()
    print("production ops secret backend implementation readiness checks passed.")


if __name__ == "__main__":
    main()
