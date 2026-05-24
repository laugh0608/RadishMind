#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-container-smoke-runbook.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
FORBIDDEN_LOCAL_REFERENCE = "D:\\Code\\Radish"

REQUIRED_BLOCKED_CONDITIONS = {
    "container_smoke_ready",
    "test_environment_smoke",
    "production_preflight_record",
    "docker_image_publish_workflow",
    "production_secret_backend",
    "production_auth_cors_policy",
    "process_supervisor",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-ops-docker-deployment-v1-plan.md": [
        "container-smoke-runbook",
        "production-ops-container-smoke-runbook.json",
        "docker compose -f deploy/docker-compose.local.yaml up --build -d",
        "container smoke",
        "production ready",
    ],
    "docs/radishmind-current-focus.md": [
        "container-smoke-runbook",
        "production-ops-container-smoke-runbook.json",
        "container_smoke_ready",
    ],
    "docs/radishmind-roadmap.md": [
        "container-smoke-runbook",
        "production-ops-container-smoke-runbook.json",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "container-smoke-runbook",
        "production-ops-container-smoke-runbook.json",
    ],
    "scripts/README.md": [
        "check-production-ops-container-smoke-runbook.py",
        "production-ops-container-smoke-runbook.json",
        "container_smoke_ready",
    ],
    "docs/devlogs/2026-W21.md": [
        "container-smoke-runbook",
        "production-ops-container-smoke-runbook.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_fixture() -> dict[str, Any]:
    return json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))


def assert_fixture(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "container-smoke-runbook", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected slice track")
    require(slice_info.get("status") == "runbook_boundary_satisfied", "runbook boundary must be satisfied")
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "container_smoke_ready",
        "test_environment_smoke_ready",
        "production_preflight_ready",
        "production_ready",
        "image_publish_ready",
        "production_secret_backend_ready",
        "process_supervisor_ready",
        "production_auth_cors_policy_ready",
        "console_runtime_config_ready",
    }:
        require(claim in does_not_claim, f"container-smoke-runbook must not claim {claim}")

    run_window = fixture.get("run_window_policy") or {}
    require(run_window.get("requires_explicit_human_run_window") is True, "run window must require human approval")
    require(run_window.get("fast_baseline_executes_docker") is False, "fast baseline must not execute Docker")
    require(run_window.get("cleanup_required") is True, "container smoke must require cleanup")

    blocked = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or []}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        require(blocked[condition_id].get("status") == "not_satisfied", f"{condition_id} must stay not_satisfied")


def assert_local_container_smoke(fixture: dict[str, Any]) -> None:
    local_smoke = fixture.get("local_container_smoke") or {}
    require(local_smoke.get("status") == "not_run", "local container smoke must remain not_run")
    require(local_smoke.get("compose_file") == "deploy/docker-compose.local.yaml", "unexpected local compose path")
    require(local_smoke.get("provider") == "mock", "local container smoke must use mock provider")

    commands = set(local_smoke.get("commands") or [])
    for command in {
        "docker compose -f deploy/docker-compose.local.yaml config",
        "docker compose -f deploy/docker-compose.local.yaml up --build -d",
        "python scripts/run-platform-local-smoke.py --base-url http://127.0.0.1:7000 --check",
        "docker compose -f deploy/docker-compose.local.yaml ps",
        "docker compose -f deploy/docker-compose.local.yaml down",
    }:
        require(command in commands, f"missing local container smoke command: {command}")

    probe_urls = set(local_smoke.get("probe_urls") or [])
    for url in {
        "http://127.0.0.1:7000/healthz",
        "http://127.0.0.1:7000/v1/platform/local-smoke",
        "http://127.0.0.1:4000/",
    }:
        require(url in probe_urls, f"missing local container smoke probe URL: {url}")

    local_compose = read("deploy/docker-compose.local.yaml")
    for literal in [
        "RADISHMIND_PLATFORM_PROVIDER: mock",
        "\"${RADISHMIND_PLATFORM_HTTP_PORT:-7000}:7000\"",
        "\"${RADISHMIND_CONSOLE_HTTP_PORT:-4000}:80\"",
        "condition: service_healthy",
    ]:
        require(literal in local_compose, f"deploy/docker-compose.local.yaml missing runbook expectation: {literal}")


def assert_test_deploy_smoke(fixture: dict[str, Any]) -> None:
    test_smoke = fixture.get("test_deploy_smoke") or {}
    require(test_smoke.get("status") == "blocked", "test deploy smoke must stay blocked")
    require(test_smoke.get("compose_file") == "deploy/docker-compose.yaml", "unexpected deploy compose path")

    required_inputs = set(test_smoke.get("requires") or [])
    for item in {
        "published test images",
        "populated deploy/.env copied from deploy/.env.example",
        "test provider profile and non-production secret source",
        "explicit test deployment window",
    }:
        require(item in required_inputs, f"missing test deploy smoke requirement: {item}")

    commands = set(test_smoke.get("commands") or [])
    for command in {
        "docker compose --env-file deploy/.env -f deploy/docker-compose.yaml config",
        "docker compose --env-file deploy/.env -f deploy/docker-compose.yaml pull",
        "docker compose --env-file deploy/.env -f deploy/docker-compose.yaml up -d",
        "docker compose --env-file deploy/.env -f deploy/docker-compose.yaml ps",
        "docker compose --env-file deploy/.env -f deploy/docker-compose.yaml down",
    }:
        require(command in commands, f"missing test deploy smoke command: {command}")


def assert_consumers_and_docs(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected = {
        "scripts/check-production-ops-container-smoke-runbook.py",
        "scripts/check-production-ops-docker-deployment-mode.py",
        "scripts/check-repo.py",
        "scripts/README.md",
        "docs/radishmind-current-focus.md",
        "docs/radishmind-roadmap.md",
        "docs/task-cards/production-ops-hardening-v1-plan.md",
        "docs/task-cards/production-ops-docker-deployment-v1-plan.md",
        "docs/devlogs/2026-W21.md",
    }
    missing = sorted(expected - required_consumers)
    require(not missing, f"missing consumers: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-container-smoke-runbook.py", [])' in check_repo,
        "check-repo.py must run container smoke runbook check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")
        require(FORBIDDEN_LOCAL_REFERENCE not in text, f"{relative_path} must not commit local Radish path")


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "production_ops_container_smoke_runbook", "unexpected fixture kind")
    assert_fixture(fixture)
    assert_local_container_smoke(fixture)
    assert_test_deploy_smoke(fixture)
    assert_consumers_and_docs(fixture)
    require(FORBIDDEN_LOCAL_REFERENCE not in FIXTURE_PATH.read_text(encoding="utf-8"), "fixture must not commit local path")
    print("production ops container smoke runbook checks passed.")


if __name__ == "__main__":
    main()
