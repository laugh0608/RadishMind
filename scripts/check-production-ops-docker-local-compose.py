#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-docker-local-compose.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
FORBIDDEN_LOCAL_REFERENCE = "D:\\Code\\Radish"

REQUIRED_ASSETS = {
    "services/platform/Dockerfile",
    "apps/radishmind-console/Dockerfile",
    "apps/radishmind-console/nginx.local.conf",
    "deploy/docker-compose.local.yaml",
}

REQUIRED_BLOCKED_CONDITIONS = {
    "docker_image_publish_workflow",
    "console_runtime_config",
    "production_secret_backend",
    "production_auth_cors_policy",
}

FORBIDDEN_SECRET_LITERALS = {
    "RADISHMIND_PLATFORM_API_KEY",
    "RADISHMIND_MODEL_API_KEY",
    "authorization:",
    "cookie:",
    "token:",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def load_fixture() -> dict[str, Any]:
    return json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))


def ids(items: list[dict[str, Any]]) -> set[str]:
    return {str(item.get("id")) for item in items}


def assert_fixture(fixture: dict[str, Any]) -> None:
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "docker-local-compose", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected slice track")
    require(
        slice_info.get("status") == "local_container_smoke_assets_satisfied",
        "docker-local-compose must be satisfied for local container smoke assets",
    )
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "docker_test_prod_compose_ready",
        "image_publish_ready",
        "production_ready",
        "production_secret_backend_ready",
        "process_supervisor_ready",
        "console_production_package_ready",
    }:
        require(claim in does_not_claim, f"docker-local-compose must not claim {claim}")

    asset_paths = {str(item.get("path")) for item in fixture.get("assets") or []}
    missing_assets = sorted(REQUIRED_ASSETS - asset_paths)
    require(not missing_assets, f"missing docker local assets in fixture: {missing_assets}")
    for path in REQUIRED_ASSETS:
        require((REPO_ROOT / path).exists(), f"missing docker local asset: {path}")

    local_services = {str(item.get("id")): item for item in fixture.get("local_services") or []}
    require(set(local_services) == {"platform", "console"}, "docker local compose must define platform and console only")
    require(local_services["platform"].get("provider") == "mock", "platform local compose provider must be mock")
    require(local_services["platform"].get("published_port") == "7000", "platform local port must be 7000")
    require(local_services["console"].get("published_port") == "4000", "console local port must be 4000")
    require(
        local_services["console"].get("platform_base_url") == "http://127.0.0.1:7000",
        "console local base URL must target host platform port",
    )

    blocked = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or []}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        require(blocked[condition_id].get("status") == "not_satisfied", f"{condition_id} must stay not_satisfied")


def assert_platform_dockerfile() -> None:
    dockerfile = read("services/platform/Dockerfile")
    required_literals = [
        "FROM golang:1.25-bookworm AS build",
        "FROM python:3.12-slim AS runtime",
        "go build",
        "pip install --no-cache-dir \"jsonschema>=4.0\"",
        "RADISHMIND_PLATFORM_LISTEN_ADDR=0.0.0.0:7000",
        "RADISHMIND_PLATFORM_PROVIDER=mock",
        "RADISHMIND_PLATFORM_MODEL=radishmind-local-docker",
        "RADISHMIND_PLATFORM_BRIDGE_SCRIPT=scripts/run-platform-bridge.py",
        "COPY services/gateway/ services/gateway/",
        "COPY services/runtime/ services/runtime/",
        "COPY scripts/run-platform-bridge.py scripts/run-platform-bridge.py",
        "COPY contracts/ contracts/",
        "COPY prompts/ prompts/",
        "EXPOSE 7000",
        "CMD [\"radishmind-platform\"]",
    ]
    missing = [literal for literal in required_literals if literal not in dockerfile]
    require(not missing, f"services/platform/Dockerfile missing literals: {missing}")
    assert_no_forbidden_secret_literals("services/platform/Dockerfile", dockerfile)


def assert_console_dockerfile() -> None:
    dockerfile = read("apps/radishmind-console/Dockerfile")
    required_literals = [
        "FROM node:24-alpine AS build",
        "npm ci",
        "ARG VITE_RADISHMIND_PLATFORM_BASE_URL=http://127.0.0.1:7000",
        "npm run build",
        "FROM nginx:1.27-alpine AS runtime",
        "COPY apps/radishmind-console/nginx.local.conf /etc/nginx/conf.d/default.conf",
        "COPY --from=build /workspace/apps/radishmind-console/dist/ /usr/share/nginx/html/",
        "EXPOSE 80",
    ]
    missing = [literal for literal in required_literals if literal not in dockerfile]
    require(not missing, f"apps/radishmind-console/Dockerfile missing literals: {missing}")
    assert_no_forbidden_secret_literals("apps/radishmind-console/Dockerfile", dockerfile)

    nginx_config = read("apps/radishmind-console/nginx.local.conf")
    require("try_files $uri $uri/ /index.html;" in nginx_config, "local nginx config must keep SPA fallback")


def assert_compose() -> None:
    compose = read("deploy/docker-compose.local.yaml")
    required_literals = [
        "platform:",
        "dockerfile: services/platform/Dockerfile",
        "RADISHMIND_PLATFORM_LISTEN_ADDR: 0.0.0.0:7000",
        "RADISHMIND_PLATFORM_PROVIDER: mock",
        "RADISHMIND_PLATFORM_MODEL: radishmind-local-docker",
        "\"${RADISHMIND_PLATFORM_HTTP_PORT:-7000}:7000\"",
        "http://127.0.0.1:7000/healthz",
        "console:",
        "dockerfile: apps/radishmind-console/Dockerfile",
        "VITE_RADISHMIND_PLATFORM_BASE_URL: ${VITE_RADISHMIND_PLATFORM_BASE_URL:-http://127.0.0.1:7000}",
        "\"${RADISHMIND_CONSOLE_HTTP_PORT:-4000}:80\"",
        "condition: service_healthy",
    ]
    missing = [literal for literal in required_literals if literal not in compose]
    require(not missing, f"deploy/docker-compose.local.yaml missing literals: {missing}")
    forbidden_literals = [
        "RADISHMIND_IMAGE_TRACK",
        "RADISHMIND_IMAGE_TAG",
        "release-latest",
        "test-latest",
        "restart:",
        "deploy:",
        "secrets:",
    ]
    present_forbidden = [literal for literal in forbidden_literals if literal in compose]
    require(not present_forbidden, f"local compose contains forbidden production/test literals: {present_forbidden}")
    assert_no_forbidden_secret_literals("deploy/docker-compose.local.yaml", compose)


def assert_no_forbidden_secret_literals(relative_path: str, text: str) -> None:
    lowered = text.lower()
    found = [literal for literal in FORBIDDEN_SECRET_LITERALS if literal.lower() in lowered]
    require(not found, f"{relative_path} contains forbidden secret literals: {found}")
    require(FORBIDDEN_LOCAL_REFERENCE not in text, f"{relative_path} must not commit local Radish path")


def assert_repo_hygiene() -> None:
    dockerignore = read(".dockerignore")
    for literal in [
        "apps/radishmind-console/node_modules/",
        "apps/radishmind-console/dist/",
        "services/platform/.cache/",
        "deploy/.env",
        "deploy/.env.local",
    ]:
        require(literal in dockerignore, f".dockerignore missing {literal}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-docker-local-compose.py", [])' in check_repo,
        "check-repo.py must run docker local compose check",
    )


def assert_consumers(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected = {
        "scripts/check-production-ops-docker-local-compose.py",
        "scripts/check-production-ops-docker-deployment-mode.py",
        "scripts/check-repo.py",
        "scripts/README.md",
        "docs/radishmind-roadmap.md",
        "docs/task-cards/production-ops-docker-deployment-v1-plan.md",
    }
    missing = sorted(expected - required_consumers)
    require(not missing, f"missing consumers: {missing}")


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "production_ops_docker_local_compose", "unexpected fixture kind")
    assert_fixture(fixture)
    assert_platform_dockerfile()
    assert_console_dockerfile()
    assert_compose()
    assert_repo_hygiene()
    assert_consumers(fixture)
    print("production ops docker local compose checks passed.")


if __name__ == "__main__":
    main()
