#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-startup-supervisor-boundary.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_STARTUP_ENTRIES = {
    "platform_service_wrapper": ("supported_for_local_dev", "manual_start_only"),
    "console_dev_entry": ("supported_for_local_dev", "dev_launcher_only"),
    "verify_only_probe": ("supported_for_local_dev", "readiness_probe_only"),
}

REQUIRED_PLATFORM_COMMANDS = {
    "serve",
    "config-summary",
    "config-check",
    "diagnostics",
}

REQUIRED_DEV_ENTRY_COMMANDS = {
    "start_or_reuse_backend",
    "start_or_reuse_frontend",
    "probe_healthz",
    "probe_overview",
    "probe_local_smoke",
    "probe_cors",
    "probe_frontend",
}

REQUIRED_READINESS_PROBES = {
    "http://127.0.0.1:7000/healthz",
    "http://127.0.0.1:7000/v1/platform/overview",
    "http://127.0.0.1:7000/v1/platform/local-smoke",
    "local console CORS preflight",
    "http://127.0.0.1:4000",
}

REQUIRED_BLOCKED_CONDITIONS = {
    "process_supervisor",
    "automatic_restart",
    "production_service_manager",
    "production_log_retention",
}

REQUIRED_FORBIDDEN_INTERPRETATIONS = {
    "local-smoke is production health",
    "dev launcher is process supervisor",
    "tmp logs are production log retention",
    "port reuse is service discovery",
    "ExitAfterProbe is lifecycle management",
}

REQUIRED_SCRIPT_LITERALS = {
    "scripts/run-platform-service.ps1": [
        '"serve"',
        '"config-summary"',
        '"config-check"',
        '"diagnostics"',
        "unsupported platform service command",
        "exit 2",
    ],
    "scripts/run-platform-service.sh": [
        "serve)",
        "config-summary | config-check | diagnostics",
        "unsupported platform service command",
        "exit 2",
    ],
    "scripts/run-radishmind-console-dev.ps1": [
        "Start-Process",
        "-WindowStyle Hidden",
        "ExitAfterProbe",
        "VerifyOnly",
        "tmp/radishmind-console-dev",
        "not a production supervisor",
    ],
    "scripts/run-radishmind-console-dev.sh": [
        "--exit-after-probe",
        "--verify-only",
        "trap cleanup EXIT INT TERM",
        "tmp/radishmind-console-dev",
        "not a production supervisor",
    ],
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "startup-supervisor-boundary",
        "production-ops-startup-supervisor-boundary.json",
        "check-production-ops-startup-supervisor-boundary.py",
    ],
    "scripts/README.md": [
        "check-production-ops-startup-supervisor-boundary.py",
        "production-ops-startup-supervisor-boundary.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_fixture() -> dict[str, Any]:
    document = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    require(isinstance(document, dict), "startup/supervisor boundary fixture must be an object")
    return document


def assert_slice(document: dict[str, Any]) -> None:
    slice_info = document.get("slice") or {}
    require(slice_info.get("id") == "startup-supervisor-boundary", "unexpected production ops slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected production ops track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "startup-supervisor-boundary must only satisfy the governance boundary",
    )
    forbidden_claims = set(slice_info.get("does_not_claim") or [])
    required_forbidden_claims = {
        "production_ready",
        "process_supervisor_ready",
        "automatic_restart_ready",
        "production_service_manager_ready",
        "console_production_package_ready",
    }
    missing = sorted(required_forbidden_claims - forbidden_claims)
    require(not missing, f"missing forbidden startup/supervisor claims: {missing}")


def assert_startup_entries(document: dict[str, Any]) -> None:
    entries = {str(item.get("id")): item for item in document.get("startup_entries") or [] if isinstance(item, dict)}
    missing = sorted(set(REQUIRED_STARTUP_ENTRIES) - set(entries))
    require(not missing, f"missing startup entries: {missing}")
    for entry_id, (status, supervision_scope) in REQUIRED_STARTUP_ENTRIES.items():
        item = entries[entry_id]
        require(item.get("status") == status, f"{entry_id} has unexpected status")
        require(item.get("supervision_scope") == supervision_scope, f"{entry_id} has unexpected supervision scope")
        require(item.get("exit_policy"), f"{entry_id} must document exit_policy")
        require(item.get("log_policy"), f"{entry_id} must document log_policy")

    platform_commands = set(entries["platform_service_wrapper"].get("commands") or [])
    missing_platform_commands = sorted(REQUIRED_PLATFORM_COMMANDS - platform_commands)
    require(not missing_platform_commands, f"missing platform wrapper commands: {missing_platform_commands}")

    dev_entry_commands = set(entries["console_dev_entry"].get("commands") or [])
    missing_dev_entry_commands = sorted(REQUIRED_DEV_ENTRY_COMMANDS - dev_entry_commands)
    require(not missing_dev_entry_commands, f"missing console dev entry commands: {missing_dev_entry_commands}")


def assert_readiness_and_blocked_conditions(document: dict[str, Any]) -> None:
    probes = set(document.get("readiness_probes") or [])
    missing_probes = sorted(REQUIRED_READINESS_PROBES - probes)
    require(not missing_probes, f"missing readiness probes: {missing_probes}")

    blocked = {str(item.get("id")): item for item in document.get("blocked_conditions") or [] if isinstance(item, dict)}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing_blocked, f"missing blocked startup/supervisor conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        item = blocked[condition_id]
        require(item.get("status") == "not_satisfied", f"{condition_id} must remain not_satisfied")
        require(
            item.get("required_before_production_ready") is True,
            f"{condition_id} must gate production ready",
        )

    forbidden = set(document.get("forbidden_interpretations") or [])
    missing_forbidden = sorted(REQUIRED_FORBIDDEN_INTERPRETATIONS - forbidden)
    require(not missing_forbidden, f"missing forbidden interpretations: {missing_forbidden}")


def assert_evidence_and_consumers(document: dict[str, Any]) -> None:
    for evidence in document.get("evidence") or []:
        evidence_path = REPO_ROOT / str(evidence)
        require(evidence_path.exists(), f"missing startup/supervisor evidence path: {evidence}")

    consumers = set(document.get("required_consumers") or [])
    expected_consumers = {
        "scripts/check-production-ops-startup-supervisor-boundary.py",
        "scripts/check-repo.py",
        "scripts/check-p3-local-product-shell-short-close-checklist.py",
        "scripts/README.md",
    }
    missing_consumers = sorted(expected_consumers - consumers)
    require(not missing_consumers, f"missing startup/supervisor consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-startup-supervisor-boundary.py", [])' in check_repo,
        "check-repo.py must run production ops startup/supervisor boundary check",
    )


def assert_script_literals() -> None:
    for relative_path, required_literals in REQUIRED_SCRIPT_LITERALS.items():
        content = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing = [literal for literal in required_literals if literal not in content]
        require(not missing, f"{relative_path} missing literals: {missing}")


def assert_doc_references() -> None:
    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        content = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing = [literal for literal in required_literals if literal not in content]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> int:
    document = load_fixture()
    require(document.get("schema_version") == 1, "unexpected schema_version")
    require(document.get("kind") == "production_ops_startup_supervisor_boundary", "unexpected fixture kind")
    assert_slice(document)
    assert_startup_entries(document)
    assert_readiness_and_blocked_conditions(document)
    assert_evidence_and_consumers(document)
    assert_script_literals()
    assert_doc_references()
    print("production ops startup/supervisor boundary checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
