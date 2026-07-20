#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-deployment-readiness-smoke.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
FORBIDDEN_LOCAL_REFERENCE = "D:\\Code\\Radish"

REQUIRED_ASSETS = {
    "deploy/docker-compose.yaml",
    "deploy/.env.example",
}

REQUIRED_BLOCKED_CONDITIONS = {
    "container_smoke_ready",
    "test_environment_smoke",
    "production_preflight_record",
    "docker_image_publish_workflow",
    "console_runtime_config",
    "production_secret_backend",
    "production_auth_cors_policy",
    "process_supervisor",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-ops-docker-deployment-v1-plan.md": [
        "deployment-readiness-smoke",
        "docker compose config",
        "production-ops-deployment-readiness-smoke.json",
        "container smoke",
        "production ready",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "deployment-readiness-smoke",
        "production-ops-deployment-readiness-smoke.json",
    ],
    "scripts/README.md": [
        "check-production-ops-deployment-readiness-smoke.py",
        "production-ops-deployment-readiness-smoke.json",
        "docker compose config",
        "container_smoke_ready",
    ],
    "docs/devlogs/2026-W21.md": [
        "deployment-readiness-smoke",
        "production-ops-deployment-readiness-smoke.json",
    ],
}

COMPOSE_PATTERN = re.compile(r"\$\{([A-Z0-9_]+)(?:(:-|:\?)([^}]*))?\}")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_fixture() -> dict[str, Any]:
    return json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))


def parse_env_example() -> dict[str, str]:
    result: dict[str, str] = {}
    for raw_line in read("deploy/.env.example").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        key, separator, value = line.partition("=")
        require(bool(separator), f"invalid deploy/.env.example line: {raw_line}")
        result[key] = value
    return result


def expand_compose_text(env: dict[str, str]) -> tuple[str, list[str]]:
    compose = read("deploy/docker-compose.yaml")
    image_tag = env.get("RADISHMIND_IMAGE_TAG") or f"{env.get('RADISHMIND_IMAGE_TRACK') or 'test'}-latest"
    compose = compose.replace("${RADISHMIND_IMAGE_TAG:-${RADISHMIND_IMAGE_TRACK:-test}-latest}", image_tag)
    missing: list[str] = []

    def replace(match: re.Match[str]) -> str:
        key = match.group(1)
        operator = match.group(2)
        fallback = match.group(3) or ""
        value = env.get(key, "")
        if operator == ":-":
            return value if value else fallback
        if operator == ":?":
            if not value:
                missing.append(key)
            return value
        return value

    return COMPOSE_PATTERN.sub(replace, compose), missing


def assert_fixture(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "deployment-readiness-smoke", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected slice track")
    require(
        slice_info.get("status") == "static_compose_expansion_satisfied",
        "deployment-readiness-smoke must satisfy static compose expansion only",
    )
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
        require(claim in does_not_claim, f"deployment-readiness-smoke must not claim {claim}")

    asset_paths = {str(item.get("path")) for item in fixture.get("assets") or []}
    missing_assets = sorted(REQUIRED_ASSETS - asset_paths)
    require(not missing_assets, f"missing deployment readiness assets in fixture: {missing_assets}")
    for path in REQUIRED_ASSETS:
        require((REPO_ROOT / path).exists(), f"missing deployment readiness asset: {path}")

    scenarios = fixture.get("static_config_scenarios") or []
    scenario_ids = {str(item.get("id")) for item in scenarios}
    require(
        scenario_ids == {"docker_test_track_latest", "docker_prod_fixed_release_tag", "missing_required_provider_inputs"},
        f"unexpected static config scenarios: {sorted(scenario_ids)}",
    )

    commands = set(fixture.get("manual_compose_config_commands") or [])
    require(
        "docker compose --env-file deploy/.env -f deploy/docker-compose.yaml config" in commands,
        "fixture must document docker compose config command",
    )
    require(
        "docker compose --env-file deploy/.env -f deploy/docker-compose.yaml config --quiet" in commands,
        "fixture must document quiet docker compose config command",
    )

    blocked = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or []}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        require(blocked[condition_id].get("status") == "not_satisfied", f"{condition_id} must stay not_satisfied")


def assert_static_expansion(fixture: dict[str, Any]) -> None:
    base_env = parse_env_example()
    for scenario in fixture.get("static_config_scenarios") or []:
        scenario_id = str(scenario.get("id"))
        scenario_env = {str(key): str(value) for key, value in (scenario.get("env") or {}).items()}
        env = {**base_env, **scenario_env}
        expanded, missing = expand_compose_text(env)

        expected_failure = scenario.get("expected_failure")
        if expected_failure is not None:
            require(set(missing) == set(expected_failure), f"{scenario_id} unexpected missing envs: {missing}")
            continue

        require(not missing, f"{scenario_id} has unexpected missing envs: {missing}")
        expected = scenario.get("expected") or {}
        for key in ["platform_image", "console_image", "platform_port", "console_port"]:
            value = expected.get(key)
            require(value and value in expanded, f"{scenario_id} missing expanded {key}: {value}")
        require("RADISHMIND_PLATFORM_API_KEY" not in expanded, f"{scenario_id} must not expand committed API key")
        require("build:" not in expanded, f"{scenario_id} deploy compose must not build images")
        require("secrets:" not in expanded, f"{scenario_id} deploy compose must not define Compose secrets")
        require("restart:" not in expanded, f"{scenario_id} deploy compose must not imply supervisor")


def assert_compose_static_boundaries() -> None:
    compose = read("deploy/docker-compose.yaml")
    for literal in [
        "env_file:",
        "required: false",
        "RADISHMIND_PLATFORM_PROVIDER: ${RADISHMIND_PLATFORM_PROVIDER:?",
        "RADISHMIND_PLATFORM_PROVIDER_PROFILE: ${RADISHMIND_PLATFORM_PROVIDER_PROFILE:?",
        "RADISHMIND_PLATFORM_MODEL: ${RADISHMIND_PLATFORM_MODEL:?",
        "condition: service_healthy",
        "http://127.0.0.1:7000/healthz",
    ]:
        require(literal in compose, f"deploy/docker-compose.yaml missing static boundary: {literal}")

    env_example = read("deploy/.env.example")
    for literal in [
        "RADISHMIND_SECRET_SOURCE=external-required-before-production",
        "RADISHMIND_PLATFORM_PROVIDER=",
        "RADISHMIND_PLATFORM_PROVIDER_PROFILE=",
        "RADISHMIND_PLATFORM_MODEL=",
    ]:
        require(literal in env_example, f"deploy/.env.example missing static boundary: {literal}")


def assert_consumers_and_docs(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected = {
        "scripts/check-production-ops-deployment-readiness-smoke.py",
        "scripts/check-production-ops-docker-deployment-mode.py",
        "scripts/check-repo.py",
        "scripts/README.md",
        "docs/task-cards/production-ops-hardening-v1-plan.md",
        "docs/task-cards/production-ops-docker-deployment-v1-plan.md",
        "docs/devlogs/2026-W21.md",
    }
    missing = sorted(expected - required_consumers)
    require(not missing, f"missing consumers: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-deployment-readiness-smoke.py", [])' in check_repo,
        "check-repo.py must run deployment readiness smoke check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")
        require(FORBIDDEN_LOCAL_REFERENCE not in text, f"{relative_path} must not commit local Radish path")


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "production_ops_deployment_readiness_smoke", "unexpected fixture kind")
    assert_fixture(fixture)
    assert_static_expansion(fixture)
    assert_compose_static_boundaries()
    assert_consumers_and_docs(fixture)
    require(FORBIDDEN_LOCAL_REFERENCE not in FIXTURE_PATH.read_text(encoding="utf-8"), "fixture must not commit local path")
    print("production ops deployment readiness static smoke checks passed.")


if __name__ == "__main__":
    main()
