#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_SATISFIED_CONDITIONS = {
    "platform_overview_readonly_route",
    "platform_local_smoke_readiness_route",
    "typescript_overview_consumer_contract",
    "local_console_shell",
    "console_behavior_gate",
    "console_visual_smoke_record",
    "console_stop_line_details",
    "console_provider_profile_inventory_details",
    "local_console_dev_entry",
    "console_production_boundary_gate",
    "production_config_secret_boundary_gate",
    "production_secret_backend_contract_gate",
    "startup_supervisor_boundary_gate",
    "environment_isolation_boundary_gate",
    "console_production_package_smoke_gate",
    "fast_baseline_consumes_p3_gates",
}

REQUIRED_BLOCKED_CONDITIONS = {
    "production_secret_backend",
    "process_supervisor",
    "deployment_environment_isolation",
    "console_production_packaging",
}

REQUIRED_STOP_LINES = {
    "real_executor",
    "durable_session_checkpoint_audit_result_store",
    "confirmation_flow_connection",
    "materialized_result_reader",
    "long_term_memory",
    "business_truth_write",
    "automatic_replay",
}

REQUIRED_DOC_REFERENCES = {
    "docs/radishmind-roadmap.md": [
        "P3 Local Product Shell / Ops Surface",
        "local usable / read-only close",
        "not_satisfied",
    ],
    "docs/radishmind-architecture.md": [
        "P3 Local Product Shell / Ops Surface",
        "local usable / read-only close",
        "check-p3-local-product-shell-short-close-checklist.py",
    ],
    "scripts/README.md": [
        "check-p3-local-product-shell-short-close-checklist.py",
        "P3 Local Product Shell / Ops Surface",
        "local usable / read-only close",
        "check-platform-local-smoke-contract.py",
    ],
}

PRODUCTION_OPS_BOUNDARY_FIXTURES = {
    "production_config_secret_boundary_gate": {
        "fixture": "scripts/checks/fixtures/production-ops-config-secret-boundary.json",
        "checker": "scripts/check-production-ops-config-secret-boundary.py",
        "slice_id": "config-secret-boundary",
        "kind": "production_ops_config_secret_boundary",
        "blocked_condition": "production_secret_backend",
        "forbidden_claim": "production_secret_backend_ready",
    },
    "production_secret_backend_contract_gate": {
        "fixture": "scripts/checks/fixtures/production-ops-secret-backend-contract.json",
        "checker": "scripts/check-production-ops-secret-backend-contract.py",
        "slice_id": "production-secret-backend-contract",
        "kind": "production_ops_secret_backend_contract",
        "blocked_condition": "production_secret_backend",
        "forbidden_claim": "production_secret_backend_ready",
    },
    "startup_supervisor_boundary_gate": {
        "fixture": "scripts/checks/fixtures/production-ops-startup-supervisor-boundary.json",
        "checker": "scripts/check-production-ops-startup-supervisor-boundary.py",
        "slice_id": "startup-supervisor-boundary",
        "kind": "production_ops_startup_supervisor_boundary",
        "blocked_condition": "process_supervisor",
        "forbidden_claim": "process_supervisor_ready",
    },
    "environment_isolation_boundary_gate": {
        "fixture": "scripts/checks/fixtures/production-ops-environment-isolation-boundary.json",
        "checker": "scripts/check-production-ops-environment-isolation-boundary.py",
        "slice_id": "environment-isolation",
        "kind": "production_ops_environment_isolation_boundary",
        "blocked_condition": "deployment_environment_isolation",
        "forbidden_claim": "production_environment_isolation_ready",
    },
    "console_production_package_smoke_gate": {
        "fixture": "scripts/checks/fixtures/production-ops-console-package-smoke.json",
        "checker": "scripts/check-production-ops-console-package-smoke.py",
        "slice_id": "console-production-package-smoke",
        "kind": "production_ops_console_package_smoke",
        "blocked_condition": "console_production_packaging",
        "forbidden_claim": "console_production_package_ready",
    },
}


def load_fixture() -> dict[str, Any]:
    return json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))


def load_json(relative_path: str) -> dict[str, Any]:
    return json.loads((REPO_ROOT / relative_path).read_text(encoding="utf-8"))


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def ids_by_status(items: list[dict[str, Any]], status: str) -> set[str]:
    return {str(item.get("id")) for item in items if item.get("status") == status}


def assert_paths_exist(items: list[dict[str, Any]]) -> None:
    for item in items:
        for evidence in item.get("evidence") or []:
            evidence_path = REPO_ROOT / str(evidence)
            require(evidence_path.exists(), f"missing evidence path: {evidence}")


def assert_stage_naming(fixture: dict[str, Any]) -> None:
    stage = fixture.get("stage") or {}
    require(stage.get("preferred_name") == "P3 Local Product Shell / Ops Surface", "unexpected P3 preferred stage name")
    aliases = set(stage.get("accepted_aliases") or [])
    require("P3 Local Deployment & Ops Governance" in aliases, "missing accepted deployment/governance alias")
    naming_policy = str(stage.get("naming_policy") or "")
    require("not a separate roadmap phase" in naming_policy, "stage naming policy must reject split phase naming")


def assert_short_close_state(fixture: dict[str, Any]) -> None:
    local_close = fixture.get("local_readonly_close_state") or {}
    require(local_close.get("status") == "satisfied", "P3 local read-only close must be satisfied")
    require(
        local_close.get("label") == "local usable / read-only close",
        "P3 local read-only close label must stay explicit",
    )
    next_default = set(local_close.get("next_default") or [])
    require(
        {"production_ops_hardening_v1", "production_ops_hardening_v1_governance_close_review"}.issubset(
            next_default
        ),
        "P3 local close must point to Production Ops Hardening v1 governance close review",
    )
    locked_claims = set(local_close.get("does_not_unlock") or [])
    required_locked_claims = {
        "production_deployment_ready",
        "console_production_package_ready",
        "real_executor_ready",
        "durable_store_ready",
        "confirmation_flow_connected",
        "business_truth_write_enabled",
        "replay_enabled",
    }
    missing_locked_claims = sorted(required_locked_claims - locked_claims)
    require(not missing_locked_claims, f"P3 local close missing locked claims: {missing_locked_claims}")

    state = fixture.get("short_close_state") or {}
    require(state.get("status") == "not_ready", "P3 short close must remain not_ready")
    forbidden_claims = set(state.get("must_not_claim") or [])
    required_claims = {
        "production_deployment_ready",
        "console_production_package_ready",
        "real_executor_ready",
        "durable_store_ready",
        "confirmation_flow_connected",
        "business_truth_write_enabled",
        "replay_enabled",
    }
    missing = sorted(required_claims - forbidden_claims)
    require(not missing, f"P3 short close missing forbidden claims: {missing}")

    refresh = fixture.get("production_ops_hardening_refresh") or {}
    require(refresh.get("status") == "satisfied", "P3 production ops hardening refresh must be satisfied")
    require(refresh.get("slice") == "short-close-checklist-refresh", "unexpected P3 refresh slice")
    require(refresh.get("refreshed_at") == "2026-05-24", "unexpected P3 refresh date")
    refresh_result = str(refresh.get("result") or "")
    require("governance boundary evidence only" in refresh_result, "P3 refresh must stay governance-only")
    refresh_boundaries = set(refresh.get("governance_boundaries") or [])
    expected_boundaries = {metadata["slice_id"] for metadata in PRODUCTION_OPS_BOUNDARY_FIXTURES.values()}
    missing_boundaries = sorted(expected_boundaries - refresh_boundaries)
    require(not missing_boundaries, f"P3 refresh missing governance boundaries: {missing_boundaries}")
    refresh_blocked = set(refresh.get("blocked_production_conditions") or [])
    missing_refresh_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - refresh_blocked)
    require(not missing_refresh_blocked, f"P3 refresh missing blocked conditions: {missing_refresh_blocked}")
    refresh_next = set(refresh.get("next_default") or [])
    require(
        "production_ops_hardening_v1_governance_close_review" in refresh_next,
        "P3 refresh must point to governance close review",
    )


def assert_conditions(fixture: dict[str, Any]) -> None:
    satisfied = fixture.get("satisfied_conditions") or []
    blocked = fixture.get("blocked_conditions") or []
    stop_lines = fixture.get("stop_lines") or []

    missing_satisfied = sorted(REQUIRED_SATISFIED_CONDITIONS - ids_by_status(satisfied, "satisfied"))
    require(not missing_satisfied, f"missing satisfied P3 conditions: {missing_satisfied}")
    assert_paths_exist(satisfied)

    blocked_by_id = {str(item.get("id")): item for item in blocked}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked_by_id))
    require(not missing_blocked, f"missing blocked P3 conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        item = blocked_by_id[condition_id]
        require(item.get("status") == "not_satisfied", f"{condition_id} must stay not_satisfied")
        require(item.get("required_before_short_close") is True, f"{condition_id} must gate P3 short close")

    disabled_stop_lines = ids_by_status(stop_lines, "disabled")
    missing_stop_lines = sorted(REQUIRED_STOP_LINES - disabled_stop_lines)
    require(not missing_stop_lines, f"missing disabled P3 stop lines: {missing_stop_lines}")


def assert_production_ops_boundary_alignment(fixture: dict[str, Any]) -> None:
    satisfied_by_id = {str(item.get("id")): item for item in fixture.get("satisfied_conditions") or []}
    blocked_by_id = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or []}

    for gate_id, metadata in PRODUCTION_OPS_BOUNDARY_FIXTURES.items():
        fixture_path = metadata["fixture"]
        checker_path = metadata["checker"]
        slice_id = metadata["slice_id"]
        blocked_condition = metadata["blocked_condition"]

        gate = satisfied_by_id.get(gate_id) or {}
        gate_evidence = set(gate.get("evidence") or [])
        require(fixture_path in gate_evidence, f"{gate_id} must cite {fixture_path}")
        require(checker_path in gate_evidence, f"{gate_id} must cite {checker_path}")

        blocked = blocked_by_id.get(blocked_condition) or {}
        require(blocked.get("status") == "not_satisfied", f"{blocked_condition} must stay not_satisfied")
        require(
            blocked.get("required_before_short_close") is True,
            f"{blocked_condition} must stay required before P3 short close",
        )
        require(blocked.get("current_scope") == "blocked", f"{blocked_condition} must stay blocked")
        boundary_evidence = set(blocked.get("boundary_evidence") or [])
        require(fixture_path in boundary_evidence, f"{blocked_condition} must cite {fixture_path}")
        require(checker_path in boundary_evidence, f"{blocked_condition} must cite {checker_path}")

        boundary_fixture = load_json(fixture_path)
        require(boundary_fixture.get("schema_version") == 1, f"{fixture_path} unexpected schema_version")
        require(boundary_fixture.get("kind") == metadata["kind"], f"{fixture_path} unexpected kind")
        boundary_slice = boundary_fixture.get("slice") or {}
        require(boundary_slice.get("id") == slice_id, f"{fixture_path} unexpected slice id")
        require(
            boundary_slice.get("track") == "Production Ops Hardening v1",
            f"{fixture_path} must stay on Production Ops Hardening v1",
        )
        require(
            boundary_slice.get("status") == "governance_boundary_satisfied",
            f"{fixture_path} must remain governance boundary satisfied",
        )
        does_not_claim = set(boundary_slice.get("does_not_claim") or [])
        require("production_ready" in does_not_claim, f"{fixture_path} must not claim production ready")
        require(
            metadata["forbidden_claim"] in does_not_claim,
            f"{fixture_path} must not claim {metadata['forbidden_claim']}",
        )

        boundary_blocked = {
            str(item.get("id")): item for item in boundary_fixture.get("blocked_conditions") or []
        }
        boundary_blocked_item = boundary_blocked.get(blocked_condition) or {}
        require(
            boundary_blocked_item.get("status") == "not_satisfied",
            f"{fixture_path} must keep {blocked_condition} not_satisfied",
        )
        require(
            boundary_blocked_item.get("required_before_production_ready") is True,
            f"{fixture_path} must require {blocked_condition} before production ready",
        )
        boundary_consumers = set(boundary_fixture.get("required_consumers") or [])
        require(
            "scripts/check-p3-local-product-shell-short-close-checklist.py" in boundary_consumers,
            f"{fixture_path} must list the P3 checklist checker as a consumer",
        )


def assert_consumers(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected_consumers = {
        "scripts/check-p3-local-product-shell-short-close-checklist.py",
        "scripts/check-production-ops-config-secret-boundary.py",
        "scripts/check-production-ops-secret-backend-contract.py",
        "scripts/check-production-ops-startup-supervisor-boundary.py",
        "scripts/check-production-ops-environment-isolation-boundary.py",
        "scripts/check-production-ops-console-package-smoke.py",
        "scripts/check-repo.py",
        "docs/radishmind-roadmap.md",
        "docs/radishmind-architecture.md",
        "scripts/README.md",
    }
    missing_consumers = sorted(expected_consumers - required_consumers)
    require(not missing_consumers, f"missing P3 checklist consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-p3-local-product-shell-short-close-checklist.py", [])' in check_repo,
        "check-repo.py must run the P3 short close checklist",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        document_text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing_literals = [literal for literal in required_literals if literal not in document_text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "p3_local_product_shell_short_close_checklist", "unexpected fixture kind")
    assert_stage_naming(fixture)
    assert_short_close_state(fixture)
    assert_conditions(fixture)
    assert_production_ops_boundary_alignment(fixture)
    assert_consumers(fixture)
    print("P3 local product shell short close checklist checks passed.")


if __name__ == "__main__":
    main()
