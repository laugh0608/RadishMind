#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-docker-test-prod-compose.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
FORBIDDEN_LOCAL_REFERENCE = "D:\\Code\\Radish"

REQUIRED_ASSETS = {
    "deploy/docker-compose.yaml",
    "deploy/.env.example",
}

REQUIRED_BLOCKED_CONDITIONS = {
    "docker_image_publish_workflow",
    "console_runtime_config",
    "production_secret_backend",
    "production_auth_cors_policy",
    "process_supervisor",
}

FORBIDDEN_DEPLOY_LITERALS = {
    "build:",
    "restart:",
    "deploy:",
    "secrets:",
    "RADISHMIND_PLATFORM_API_KEY",
    "RADISHMIND_MODEL_API_KEY",
    "authorization:",
    "cookie:",
    "token:",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-ops-docker-deployment-v1-plan.md": [
        "docker-test-prod-compose",
        "deploy/docker-compose.yaml",
        "deploy/.env.example",
        "RADISHMIND_IMAGE_TRACK=test",
        "RADISHMIND_IMAGE_TRACK=release",
        "production ready",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "docker-test-prod-compose",
        "deploy/docker-compose.yaml",
        "production-ops-docker-test-prod-compose.json",
    ],
    "scripts/README.md": [
        "check-production-ops-docker-test-prod-compose.py",
        "production-ops-docker-test-prod-compose.json",
        "deploy/docker-compose.yaml",
        "RADISHMIND_IMAGE_TRACK",
    ],
    "docs/devlogs/2026-W21.md": [
        "docker-test-prod-compose",
        "production-ops-docker-test-prod-compose.json",
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
    require(slice_info.get("id") == "docker-test-prod-compose", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected slice track")
    require(
        slice_info.get("status") == "deploy_compose_boundary_satisfied",
        "docker-test-prod-compose must satisfy deploy compose boundary only",
    )
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "image_publish_ready",
        "production_ready",
        "production_secret_backend_ready",
        "process_supervisor_ready",
        "production_auth_cors_policy_ready",
        "console_runtime_config_ready",
        "console_production_package_ready",
    }:
        require(claim in does_not_claim, f"docker-test-prod-compose must not claim {claim}")

    asset_paths = {str(item.get("path")) for item in fixture.get("assets") or []}
    missing_assets = sorted(REQUIRED_ASSETS - asset_paths)
    require(not missing_assets, f"missing docker test/prod assets in fixture: {missing_assets}")
    for path in REQUIRED_ASSETS:
        require((REPO_ROOT / path).exists(), f"missing docker test/prod asset: {path}")

    image_policy = fixture.get("image_policy") or {}
    require(image_policy.get("track_env") == "RADISHMIND_IMAGE_TRACK", "unexpected image track env")
    require(image_policy.get("fixed_tag_env") == "RADISHMIND_IMAGE_TAG", "unexpected fixed tag env")
    require(image_policy.get("default_track") == "test", "deploy compose must default to test track")
    require(image_policy.get("test_track") == "test", "docker_test must use test track")
    require(image_policy.get("prod_track") == "release", "docker_prod must use release track")
    require(image_policy.get("publish_workflow_status") == "not_satisfied", "image publish workflow must stay blocked")

    services = {str(item.get("id")): item for item in fixture.get("deploy_services") or []}
    require(set(services) == {"platform", "console"}, "deploy compose must define platform and console only")
    for service in services.values():
        require(service.get("build") == "forbidden", f"{service.get('id')} must not use local build")
        require(service.get("default_bind") == "127.0.0.1", f"{service.get('id')} must default to loopback bind")
    require(services["console"].get("runtime_config") == "not_satisfied", "console runtime config must stay blocked")

    modes = {str(item.get("id")): item for item in fixture.get("environment_modes") or []}
    require(set(modes) == {"docker_test", "docker_prod"}, "fixture must cover docker_test and docker_prod")
    require(modes["docker_test"].get("status") == "deploy_compose_boundary_available", "unexpected docker_test status")
    require(
        modes["docker_prod"].get("status") == "deploy_compose_boundary_available_but_not_ready",
        "docker_prod must remain not ready",
    )

    blocked = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or []}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        require(blocked[condition_id].get("status") == "not_satisfied", f"{condition_id} must stay not_satisfied")


def assert_compose() -> None:
    compose = read("deploy/docker-compose.yaml")
    required_literals = [
        "name: ${RADISHMIND_COMPOSE_PROJECT_NAME:-radishmind}",
        "platform:",
        "image: ${RADISHMIND_IMAGE_REGISTRY:-ghcr.io/laugh0608/radishmind}/platform:${RADISHMIND_IMAGE_TAG:-${RADISHMIND_IMAGE_TRACK:-test}-latest}",
        "path: ./.env",
        "required: false",
        "RADISHMIND_PLATFORM_LISTEN_ADDR: 0.0.0.0:7000",
        "RADISHMIND_PLATFORM_PROVIDER: ${RADISHMIND_PLATFORM_PROVIDER:?",
        "RADISHMIND_PLATFORM_PROVIDER_PROFILE: ${RADISHMIND_PLATFORM_PROVIDER_PROFILE:?",
        "RADISHMIND_PLATFORM_MODEL: ${RADISHMIND_PLATFORM_MODEL:?",
        "\"${RADISHMIND_PLATFORM_HTTP_BIND:-127.0.0.1}:${RADISHMIND_PLATFORM_HTTP_PORT:-7000}:7000\"",
        "http://127.0.0.1:7000/healthz",
        "console:",
        "image: ${RADISHMIND_IMAGE_REGISTRY:-ghcr.io/laugh0608/radishmind}/console:${RADISHMIND_IMAGE_TAG:-${RADISHMIND_IMAGE_TRACK:-test}-latest}",
        "\"${RADISHMIND_CONSOLE_HTTP_BIND:-127.0.0.1}:${RADISHMIND_CONSOLE_HTTP_PORT:-4000}:80\"",
        "condition: service_healthy",
    ]
    missing = [literal for literal in required_literals if literal not in compose]
    require(not missing, f"deploy/docker-compose.yaml missing literals: {missing}")
    assert_no_forbidden_literals("deploy/docker-compose.yaml", compose)


def assert_env_example() -> None:
    env_example = read("deploy/.env.example")
    required_literals = [
        "RADISHMIND_IMAGE_REGISTRY=ghcr.io/laugh0608/radishmind",
        "RADISHMIND_IMAGE_TRACK=test",
        "RADISHMIND_IMAGE_TAG=",
        "RADISHMIND_PUBLIC_BASE_URL=https://radishmind-test.example.invalid",
        "RADISHMIND_SECRET_SOURCE=external-required-before-production",
        "RADISHMIND_PLATFORM_HTTP_BIND=127.0.0.1",
        "RADISHMIND_PLATFORM_PROVIDER=",
        "RADISHMIND_PLATFORM_PROVIDER_PROFILE=",
        "RADISHMIND_PLATFORM_MODEL=",
        "RADISHMIND_CONSOLE_HTTP_BIND=127.0.0.1",
    ]
    missing = [literal for literal in required_literals if literal not in env_example]
    require(not missing, f"deploy/.env.example missing literals: {missing}")
    assert_no_forbidden_literals("deploy/.env.example", env_example)


def assert_no_forbidden_literals(relative_path: str, text: str) -> None:
    lowered = text.lower()
    found = [literal for literal in FORBIDDEN_DEPLOY_LITERALS if literal.lower() in lowered]
    require(not found, f"{relative_path} contains forbidden deploy literals: {found}")
    require(FORBIDDEN_LOCAL_REFERENCE not in text, f"{relative_path} must not commit local Radish path")


def assert_consumers_and_docs(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected = {
        "scripts/check-production-ops-docker-test-prod-compose.py",
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
        'run_python_script("check-production-ops-docker-test-prod-compose.py", [])' in check_repo,
        "check-repo.py must run docker test/prod compose check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")
        require(FORBIDDEN_LOCAL_REFERENCE not in text, f"{relative_path} must not commit local Radish path")


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "production_ops_docker_test_prod_compose", "unexpected fixture kind")
    assert_fixture(fixture)
    assert_compose()
    assert_env_example()
    assert_consumers_and_docs(fixture)
    require(FORBIDDEN_LOCAL_REFERENCE not in FIXTURE_PATH.read_text(encoding="utf-8"), "fixture must not commit local path")
    print("production ops docker test/prod compose checks passed.")


if __name__ == "__main__":
    main()
