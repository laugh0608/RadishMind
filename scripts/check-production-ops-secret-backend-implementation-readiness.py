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
    "fake_resolver_implemented",
    "no_secret_leakage_smoke_runtime_created",
    "backend_health_runtime_created",
    "backend_health_runtime_task_card_created",
    "backend_health_client_created",
    "backend_health_check_executed",
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
    "provider-profile-secret-binding": "satisfied",
    "secret-resolver-interface-disabled": "satisfied",
    "operator-runbook-and-negative-gates": "satisfied",
    "rotation-and-audit-policy": "satisfied",
    "test-fixture-strategy": "satisfied_for_test_only_fake_resolver",
    "fake-resolver-contract-no-secret-leakage-smoke-strategy": "strategy_defined_static_only",
    "fake-resolver-implementation-task-card-entry-readiness-review": "ready_for_next_task",
    "fake-resolver-implementation": "task_card_defined_runtime_not_started",
    "fake-resolver-runtime-implementation-entry-review": "ready_for_runtime_task",
    "fake-resolver-runtime-implementation": "fake_resolver_runtime_test_only_implemented",
    "real-resolver-runtime-preconditions": "real_resolver_runtime_preconditions_defined",
    "real-resolver-runtime-implementation-entry-review": "real_resolver_runtime_implementation_entry_review_defined",
    "real-resolver-runtime-implementation-entry-refresh": (
        "real_resolver_runtime_implementation_entry_refresh_defined"
    ),
    "resolver-backend-profile-selection-readiness": "resolver_backend_profile_selection_readiness_defined",
    "real-resolver-no-secret-leakage-smoke-runtime-strategy": (
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined"
    ),
    "real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review": (
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined"
    ),
    "credential-handle-runtime-boundary-readiness": "credential_handle_runtime_boundary_readiness_defined",
    "credential-handle-runtime-implementation-entry-review": (
        "credential_handle_runtime_implementation_entry_review_defined"
    ),
    "operator-approval-runtime-evidence-readiness": "operator_approval_runtime_evidence_readiness_defined",
    "audit-store-handoff-readiness": "audit_store_handoff_readiness_defined",
    "audit-store-runtime-implementation-entry-review": "audit_store_runtime_implementation_entry_review_defined",
    "resolver-backend-health-boundary-readiness": "resolver_backend_health_boundary_readiness_defined",
    "resolver-backend-health-runtime-implementation-entry-review": (
        "resolver_backend_health_runtime_implementation_entry_review_defined"
    ),
    "operator-approval-runtime-implementation-entry-review": (
        "operator_approval_runtime_implementation_entry_review_defined"
    ),
}

REQUIRED_BLOCKED = {
    "production_secret_backend": "not_satisfied",
    "cloud_secret_service_integration": "not_satisfied",
    "test_fixture_strategy": "satisfied_for_test_only_fake_resolver",
    "fake_resolver_implementation": "test_only_runtime_implemented_disabled_by_default",
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
        "production-secret-backend-provider-profile-secret-binding-readiness-v1",
        "provider_profile_secret_binding_readiness_defined",
        "production-secret-backend-secret-resolver-interface-disabled-readiness-v1",
        "secret_resolver_interface_disabled_readiness_defined",
        "production-secret-backend-operator-runbook-negative-gates-readiness-v1",
        "operator_runbook_negative_gates_readiness_defined",
        "production-secret-backend-rotation-audit-policy-readiness-v1",
        "rotation_audit_policy_readiness_defined",
        "production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1",
        "test_fixture_strategy_fake_resolver_entry_review_defined",
        "production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1",
        "fake_resolver_contract_no_secret_leakage_smoke_strategy_defined",
        "production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1",
        "fake_resolver_implementation_task_card_entry_readiness_review_defined",
        "production-secret-backend-fake-resolver-implementation-v1",
        "fake_resolver_implementation_task_card_defined",
        "production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1",
        "fake_resolver_runtime_implementation_entry_review_defined",
        "production-secret-backend-fake-resolver-runtime-implementation-v1",
        "fake_resolver_runtime_test_only_implemented",
        "production-secret-backend-real-resolver-runtime-preconditions-v1",
        "real_resolver_runtime_preconditions_defined",
        "production-secret-backend-real-resolver-runtime-implementation-entry-review-v1",
        "real_resolver_runtime_implementation_entry_review_defined",
        "production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1",
        "real_resolver_runtime_implementation_entry_refresh_defined",
        "production-secret-backend-resolver-backend-profile-selection-readiness-v1",
        "resolver_backend_profile_selection_readiness_defined",
        "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1",
        "real_resolver_no_secret_leakage_smoke_runtime_strategy_defined",
        "production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1",
        "real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined",
        "production-secret-backend-credential-handle-runtime-boundary-readiness-v1",
        "credential_handle_runtime_boundary_readiness_defined",
        "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1",
        "credential_handle_runtime_implementation_entry_review_defined",
        "production-secret-backend-operator-approval-runtime-evidence-readiness-v1",
        "operator_approval_runtime_evidence_readiness_defined",
        "production-secret-backend-operator-approval-runtime-implementation-entry-review-v1",
        "operator_approval_runtime_implementation_entry_review_defined",
        "production-secret-backend-audit-store-handoff-readiness-v1",
        "audit_store_handoff_readiness_defined",
        "production-secret-backend-audit-store-runtime-implementation-entry-review-v1",
        "audit_store_runtime_implementation_entry_review_defined",
        "production-secret-backend-resolver-backend-health-boundary-readiness-v1",
        "resolver_backend_health_boundary_readiness_defined",
        "production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1",
        "resolver_backend_health_runtime_implementation_entry_review_defined",
        "services/platform/internal/secretbackend/fake_resolver.go",
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
        "check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py",
        "check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
        "check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
        "check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py",
        "check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py",
        "check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
        "check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py",
        "check-production-ops-secret-backend-fake-resolver-implementation-v1.py",
        "check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py",
        "check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py",
        "check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py",
        "check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py",
        "check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py",
        "check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py",
        "check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py",
        "check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py",
        "check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py",
        "check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py",
        "check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py",
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
    require(
        target.get("production_secret_backend_status") == "not_satisfied",
        "production secret backend must remain not_satisfied",
    )
    require(target.get("resolver_implementation_status") == "not_started", "resolver must not be implemented")
    require(target.get("resolver_runtime_status") == "not_created", "resolver runtime must not be created")
    require(
        target.get("production_resolver_runtime_task_card_status") == "not_created",
        "production resolver runtime task card must not be created",
    )
    require(
        target.get("fake_resolver_status") == "test_only_runtime_implemented_disabled_by_default",
        "fake resolver status drifted",
    )
    require(
        target.get("fake_resolver_contract_status") == "static_contract_defined",
        "fake resolver contract strategy must be recorded",
    )
    require(
        target.get("no_secret_leakage_smoke_strategy_status") == "static_strategy_defined",
        "no leakage smoke strategy must be recorded",
    )
    require(
        target.get("no_secret_leakage_smoke_runtime_status") == "implemented_offline_go_test",
        "no leakage smoke runtime status drifted",
    )
    require(
        target.get("test_fixture_strategy_status") == "satisfied_for_test_only_fake_resolver",
        "test fixture strategy status drifted",
    )
    require(
        target.get("test_fixture_strategy_review_status") == "blocked_entry_review_defined",
        "test fixture strategy review status drifted",
    )
    require(
        target.get("fake_resolver_implementation_task_card_entry_review_status") == "ready_for_next_task",
        "fake resolver implementation task card entry review status drifted",
    )
    require(
        target.get("fake_resolver_implementation_task_card_status") == "created_static_task_card",
        "fake resolver implementation task card status drifted",
    )
    require(
        target.get("fake_resolver_implementation_status") == "task_card_defined_runtime_not_started",
        "fake resolver implementation status drifted",
    )
    require(
        target.get("fake_resolver_runtime_implementation_entry_review_status") == "ready_for_next_task",
        "fake resolver runtime implementation entry review status drifted",
    )
    require(
        target.get("fake_resolver_runtime_implementation_task_card_status") == "created_static_task_card",
        "fake resolver runtime implementation task card status drifted",
    )
    require(
        target.get("fake_resolver_runtime_implementation_status") == "fake_resolver_runtime_test_only_implemented",
        "fake resolver runtime implementation status drifted",
    )
    require(
        target.get("fake_resolver_runtime_status") == "implemented_test_only_disabled_by_default",
        "fake resolver runtime status drifted",
    )
    require(
        target.get("real_resolver_runtime_preconditions_status") == "real_resolver_runtime_preconditions_defined",
        "real resolver runtime preconditions status drifted",
    )
    require(
        target.get("real_resolver_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "real resolver runtime implementation entry review status drifted",
    )
    require(
        target.get("real_resolver_runtime_implementation_entry_refresh_status")
        == "blocked_before_runtime_task_card",
        "real resolver runtime implementation entry refresh status drifted",
    )
    require(
        target.get("resolver_backend_profile_selection_readiness_status") == "defined_without_backend_runtime",
        "resolver backend profile selection readiness status drifted",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_strategy_status") == "defined_without_runtime",
        "real resolver no leakage smoke runtime strategy status drifted",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "real resolver no leakage smoke runtime implementation entry review status drifted",
    )
    require(
        target.get("real_resolver_no_secret_leakage_smoke_runtime_status") == "not_created",
        "real resolver no leakage smoke runtime must remain not_created",
    )
    require(
        target.get("credential_handle_runtime_boundary_readiness_status") == "defined_without_runtime",
        "credential handle runtime boundary readiness status drifted",
    )
    require(
        target.get("credential_handle_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "credential handle runtime implementation entry review status drifted",
    )
    require(
        target.get("credential_handle_runtime_status") == "not_created",
        "credential handle runtime must remain not_created",
    )
    require(target.get("credential_handle_status") == "not_created", "credential handle must remain not_created")
    require(target.get("credential_payload_status") == "forbidden", "credential payload must remain forbidden")
    require(
        target.get("operator_approval_runtime_evidence_readiness_status")
        == "defined_without_runtime_execution",
        "operator approval runtime evidence readiness status drifted",
    )
    require(
        target.get("operator_approval_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "operator approval runtime implementation entry review status drifted",
    )
    require(
        target.get("operator_approval_runtime_status") == "not_created",
        "operator approval runtime must remain not_created",
    )
    require(
        target.get("operator_approval_runtime_execution_status") == "not_executed",
        "operator approval runtime execution must remain not_executed",
    )
    require(
        target.get("audit_store_handoff_readiness_status") == "defined_without_store_runtime",
        "audit store handoff readiness status drifted",
    )
    require(
        target.get("audit_store_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "audit store runtime implementation entry review status drifted",
    )
    require(target.get("audit_store_runtime_status") == "not_created", "audit store runtime must remain not_created")
    require(target.get("audit_store_status") == "not_created", "audit store must remain not_created")
    require(target.get("audit_writer_status") == "not_created", "audit writer must remain not_created")
    require(
        target.get("audit_event_delivery_status") == "not_executed",
        "audit event delivery must remain not_executed",
    )
    require(
        target.get("resolver_backend_health_boundary_readiness_status")
        == "defined_without_backend_health_runtime",
        "resolver backend health boundary readiness status drifted",
    )
    require(
        target.get("resolver_backend_health_runtime_implementation_entry_review_status")
        == "blocked_before_runtime_task_card",
        "resolver backend health runtime implementation entry review status drifted",
    )
    require(target.get("backend_health_runtime_status") == "not_created", "backend health runtime must remain not_created")
    require(
        target.get("backend_health_check_status") == "not_executed",
        "backend health check must remain not_executed",
    )
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
        require(
            status == "satisfied"
            or status == "satisfied_for_test_only_fake_resolver"
            or status.startswith("required_before_"),
            f"{precondition_id} has unexpected status",
        )
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
        if precondition_id == "provider-profile-binding":
            require(status == "satisfied", "provider-profile-binding precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md",
                "docs/task-cards/production-secret-backend-provider-profile-secret-binding-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py",
            }:
                require(path in evidence, f"provider-profile-binding missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"provider-profile-binding evidence missing on disk: {path}")
        if precondition_id in {"sanitized-audit-fields", "failure-taxonomy"}:
            require(status == "satisfied", f"{precondition_id} precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json",
                "scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
            }:
                require(path in evidence, f"{precondition_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{precondition_id} evidence missing on disk: {path}")
        if precondition_id == "test-fixture-strategy":
            require(
                status == "satisfied_for_test_only_fake_resolver",
                "test-fixture-strategy must be satisfied for test-only fake resolver runtime",
            )
            for required in {
                "fake resolver",
                "placeholder secret ref",
                "offline fast baseline",
                "no cloud SDK call",
                "no secret leakage smoke",
                "artifact guard",
            }:
                require(required in must_define, f"test-fixture-strategy must define {required}")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-implementation-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py",
                "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py",
                "services/platform/internal/secretbackend/fake_resolver.go",
                "services/platform/internal/secretbackend/fake_resolver_test.go",
            }:
                require(path in evidence, f"test-fixture-strategy missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"test-fixture-strategy evidence missing on disk: {path}")
        if precondition_id == "operator-runbook":
            require(status == "satisfied", "operator-runbook precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md",
                "docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
            }:
                require(path in evidence, f"operator-runbook missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"operator-runbook evidence missing on disk: {path}")
        if precondition_id == "rotation-and-audit-policy":
            require(status == "satisfied", "rotation-and-audit-policy precondition must be satisfied")
            evidence = set(item.get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-rotation-audit-policy-readiness-v1.md",
                "docs/task-cards/production-secret-backend-rotation-audit-policy-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py",
            }:
                require(path in evidence, f"rotation-and-audit-policy missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"rotation-and-audit-policy evidence missing on disk: {path}")

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
        if slice_id == "provider-profile-secret-binding":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-provider-profile-secret-binding-readiness-v1.md",
                "docs/task-cards/production-secret-backend-provider-profile-secret-binding-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-provider-profile-secret-binding-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "secret-resolver-interface-disabled":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md",
                "docs/task-cards/production-secret-backend-secret-resolver-interface-disabled-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "operator-runbook-and-negative-gates":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-operator-runbook-negative-gates-readiness-v1.md",
                "docs/task-cards/production-secret-backend-operator-runbook-negative-gates-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-operator-runbook-negative-gates-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "rotation-and-audit-policy":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-rotation-audit-policy-readiness-v1.md",
                "docs/task-cards/production-secret-backend-rotation-audit-policy-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "test-fixture-strategy":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-contract-no-secret-leakage-smoke-strategy":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-implementation-task-card-entry-readiness-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-implementation":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-implementation-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "fake-resolver-runtime-implementation":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md",
                "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json",
                "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-runtime-preconditions":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-runtime-implementation-entry-refresh":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "resolver-backend-profile-selection-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md",
                "docs/task-cards/production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-no-secret-leakage-smoke-runtime-strategy":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "credential-handle-runtime-boundary-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.md",
                "docs/task-cards/production-secret-backend-credential-handle-runtime-boundary-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "credential-handle-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.md",
                (
                    "docs/task-cards/"
                    "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1-plan.md"
                ),
                (
                    "scripts/checks/fixtures/"
                    "production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.json"
                ),
                "scripts/check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "operator-approval-runtime-evidence-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md",
                "docs/task-cards/production-secret-backend-operator-approval-runtime-evidence-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "operator-approval-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-handoff-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-handoff-readiness-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-handoff-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-handoff-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "audit-store-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "resolver-backend-health-boundary-readiness":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-resolver-backend-health-boundary-readiness-v1.md",
                "docs/task-cards/production-secret-backend-resolver-backend-health-boundary-readiness-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-boundary-readiness-v1.json",
                "scripts/check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "resolver-backend-health-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py",
            }:
                require(path in evidence, f"{slice_id} missing evidence: {path}")
                require((REPO_ROOT / path).exists(), f"{slice_id} evidence missing on disk: {path}")
        if slice_id == "real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review":
            evidence = set(planned[slice_id].get("evidence") or [])
            for path in {
                "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md",
                "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1-plan.md",
                "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
                "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py",
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
        "real resolver runtime implementation entry refresh blocked before task card",
        "credential handle runtime implementation entry review blocked before task card",
        "real resolver no leakage smoke runtime implementation entry review blocked before task card",
        "operator approval runtime evidence readiness defined without runtime execution",
        "operator approval runtime implementation entry review blocked before task card",
        "audit store handoff readiness defined without store runtime",
        "audit store runtime implementation entry review blocked before task card",
        "resolver backend health boundary readiness defined without backend health runtime",
        "resolver backend health runtime implementation entry review blocked before task card",
    }:
        require(required in validation, f"validation strategy missing: {required}")

    for evidence in fixture.get("evidence") or []:
        require((REPO_ROOT / str(evidence)).exists(), f"missing evidence path: {evidence}")

    expected_consumers = {
        "scripts/check-production-ops-secret-backend-implementation-readiness.py",
        "scripts/check-production-ops-secret-backend-config-secret-ref-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-provider-profile-secret-binding-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py",
        "scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py",
        "scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py",
        "scripts/check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py",
        "scripts/check-production-secret-reference-contract.py",
        "scripts/check-repo.py",
        "scripts/README.md",
        "docs/radishmind-current-focus.md",
        "docs/radishmind-roadmap.md",
        "docs/radishmind-capability-matrix.md",
        "docs/task-cards/production-ops-hardening-v1-plan.md",
        "docs/task-cards/production-secret-backend-implementation-v1-plan.md",
        "docs/platform/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.json",
        "docs/platform/production-secret-backend-fake-resolver-implementation-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json",
        "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md",
        "docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json",
        "services/platform/internal/secretbackend/fake_resolver.go",
        "services/platform/internal/secretbackend/fake_resolver_test.go",
        "docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json",
        "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.json",
        "docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json",
        "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.json",
        "docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.md",
        "docs/task-cards/production-secret-backend-credential-handle-runtime-boundary-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.json",
        "docs/platform/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.md",
        "docs/task-cards/production-secret-backend-operator-approval-runtime-evidence-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-evidence-readiness-v1.json",
        "docs/platform/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-audit-store-handoff-readiness-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-handoff-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-handoff-readiness-v1.json",
        "docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-audit-store-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-audit-store-runtime-implementation-entry-review-v1.json",
        "docs/platform/production-secret-backend-resolver-backend-health-boundary-readiness-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-health-boundary-readiness-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-boundary-readiness-v1.json",
        "docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.md",
        "docs/task-cards/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1-plan.md",
        "scripts/checks/fixtures/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.json",
        "services/platform/README.md",
        "docs/devlogs/2026-W22.md",
        "docs/devlogs/2026-W25.md",
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
    require(
        'run_python_script("check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run secret resolver interface disabled readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run operator runbook negative gates readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run rotation audit policy readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run test fixture strategy fake resolver entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver contract no leakage strategy check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver implementation task card entry readiness review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-implementation-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver implementation check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py", [])'
        in check_repo,
        "check-repo.py must run fake resolver runtime implementation check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver runtime preconditions check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver runtime implementation entry refresh check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run resolver backend profile selection readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver no leakage smoke runtime strategy check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run real resolver no leakage smoke runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run credential handle runtime boundary readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run credential handle runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run operator approval runtime evidence readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run operator approval runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run audit store handoff readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run audit store runtime implementation entry review check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run backend health boundary readiness check",
    )
    require(
        'run_python_script("check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run backend health runtime implementation entry review check",
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
