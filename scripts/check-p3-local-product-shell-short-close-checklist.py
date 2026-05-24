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
    "startup_supervisor_boundary_gate",
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
    "docs/radishmind-current-focus.md": [
        "P3 Local Product Shell / Ops Surface",
        "local usable / read-only close",
        "UI Design Topic / Pencil Draft",
        "p3-local-product-shell-short-close-checklist.json",
        "check-p3-local-product-shell-short-close-checklist.py",
    ],
    "docs/radishmind-roadmap.md": [
        "P3 Local Product Shell / Ops Surface",
        "local usable / read-only close",
        "UI Design Topic / Pencil Draft",
        "p3-local-product-shell-short-close-checklist.json",
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


def load_fixture() -> dict[str, Any]:
    return json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))


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
        {"production_ops_hardening_v1", "config_secret_boundary"}.issubset(next_default),
        "P3 local close must point to Production Ops Hardening v1 config-secret-boundary",
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


def assert_consumers(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected_consumers = {
        "scripts/check-p3-local-product-shell-short-close-checklist.py",
        "scripts/check-production-ops-config-secret-boundary.py",
        "scripts/check-production-ops-startup-supervisor-boundary.py",
        "scripts/check-repo.py",
        "docs/radishmind-current-focus.md",
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
    assert_consumers(fixture)
    print("P3 local product shell short close checklist checks passed.")


if __name__ == "__main__":
    main()
