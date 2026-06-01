#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-docker-image-build-publish.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
FORBIDDEN_LOCAL_REFERENCE = "D:\\Code\\Radish"
FUTURE_WORKFLOW_PATH = ".github/workflows/docker-images.yml"

REQUIRED_ASSETS = {
    "services/platform/Dockerfile",
    "apps/radishmind-console/Dockerfile",
    "deploy/docker-compose.local.yaml",
    "deploy/docker-compose.yaml",
}

REQUIRED_BLOCKED_CONDITIONS = {
    "docker_image_publish_workflow",
    "console_runtime_config",
    "production_secret_backend",
    "production_auth_cors_policy",
    "process_supervisor",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-ops-docker-deployment-v1-plan.md": [
        "docker-image-build-publish",
        "platform / console 镜像命名",
        "v*-dev",
        "v*-test",
        "v*-release",
        "production ready",
    ],
    "docs/radishmind-current-focus.md": [
        "docker-image-build-publish",
        "production-ops-docker-image-build-publish.json",
        "v*-dev",
        "v*-test",
        "v*-release",
    ],
    "docs/radishmind-roadmap.md": [
        "docker-image-build-publish",
        "production-ops-docker-image-build-publish.json",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "docker-image-build-publish",
        "production-ops-docker-image-build-publish.json",
    ],
    "scripts/README.md": [
        "check-production-ops-docker-image-build-publish.py",
        "production-ops-docker-image-build-publish.json",
        "v*-dev",
        "docker_image_publish_workflow",
    ],
    "docs/devlogs/2026-W21.md": [
        "docker-image-build-publish",
        "production-ops-docker-image-build-publish.json",
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
    require(slice_info.get("id") == "docker-image-build-publish", "unexpected slice id")
    require(slice_info.get("track") == "Production Ops Hardening v1", "unexpected slice track")
    require(
        slice_info.get("status") == "image_naming_policy_satisfied",
        "docker-image-build-publish must satisfy image naming policy only",
    )
    does_not_claim = set(slice_info.get("does_not_claim") or [])
    for claim in {
        "image_publish_ready",
        "production_ready",
        "production_secret_backend_ready",
        "process_supervisor_ready",
        "production_auth_cors_policy_ready",
        "console_runtime_config_ready",
    }:
        require(claim in does_not_claim, f"docker-image-build-publish must not claim {claim}")

    asset_paths = {str(item.get("path")) for item in fixture.get("assets") or []}
    missing_assets = sorted(REQUIRED_ASSETS - asset_paths)
    require(not missing_assets, f"missing image naming assets in fixture: {missing_assets}")
    for path in REQUIRED_ASSETS:
        require((REPO_ROOT / path).exists(), f"missing image naming asset: {path}")

    image_policy = fixture.get("image_policy") or {}
    require(image_policy.get("registry_default") == "ghcr.io/laugh0608/radishmind", "unexpected default registry")
    require(image_policy.get("registry_env") == "RADISHMIND_IMAGE_REGISTRY", "unexpected registry env")
    require(image_policy.get("track_env") == "RADISHMIND_IMAGE_TRACK", "unexpected track env")
    require(image_policy.get("fixed_tag_env") == "RADISHMIND_IMAGE_TAG", "unexpected fixed tag env")
    require(set(image_policy.get("repositories") or []) == {"platform", "console"}, "unexpected image repositories")
    require(set(image_policy.get("tracks") or []) == {"dev", "test", "release"}, "unexpected image tracks")
    require(
        set(image_policy.get("moving_tags") or []) == {"dev-latest", "test-latest", "release-latest"},
        "unexpected moving tags",
    )
    require(
        set(image_policy.get("fixed_tag_patterns") or []) == {"v*-dev", "v*-test", "v*-release"},
        "unexpected fixed tag patterns",
    )
    require(image_policy.get("local_tags_publishable") is False, "local tags must not be publishable")

    publish_workflow = fixture.get("publish_workflow") or {}
    require(publish_workflow.get("status") == "not_satisfied", "publish workflow must stay not_satisfied")
    require(publish_workflow.get("future_asset") == FUTURE_WORKFLOW_PATH, "unexpected future workflow path")

    blocked = {str(item.get("id")): item for item in fixture.get("blocked_conditions") or []}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        require(blocked[condition_id].get("status") == "not_satisfied", f"{condition_id} must stay not_satisfied")


def assert_dockerfiles() -> None:
    platform_dockerfile = read("services/platform/Dockerfile")
    for literal in [
        "ARG BUILD_VERSION=local",
        "go build",
        "-ldflags \"-X main.buildVersion=${BUILD_VERSION}\"",
        "CMD [\"radishmind-platform\"]",
    ]:
        require(literal in platform_dockerfile, f"services/platform/Dockerfile missing {literal}")

    console_dockerfile = read("apps/radishmind-console/Dockerfile")
    for literal in [
        "ARG VITE_RADISHMIND_PLATFORM_BASE_URL=http://127.0.0.1:7000",
        "ENV VITE_RADISHMIND_PLATFORM_BASE_URL=${VITE_RADISHMIND_PLATFORM_BASE_URL}",
        "npm run build",
        "FROM nginx:1.27-alpine AS runtime",
    ]:
        require(literal in console_dockerfile, f"apps/radishmind-console/Dockerfile missing {literal}")


def assert_compose_image_names() -> None:
    local_compose = read("deploy/docker-compose.local.yaml")
    for literal in [
        "${RADISHMIND_PLATFORM_IMAGE:-radishmind/platform:local}",
        "${RADISHMIND_CONSOLE_IMAGE:-radishmind/console:local}",
        "dockerfile: services/platform/Dockerfile",
        "dockerfile: apps/radishmind-console/Dockerfile",
        "BUILD_VERSION: ${RADISHMIND_BUILD_VERSION:-local}",
    ]:
        require(literal in local_compose, f"deploy/docker-compose.local.yaml missing {literal}")

    deploy_compose = read("deploy/docker-compose.yaml")
    for literal in [
        "${RADISHMIND_IMAGE_REGISTRY:-ghcr.io/laugh0608/radishmind}/platform:${RADISHMIND_IMAGE_TAG:-${RADISHMIND_IMAGE_TRACK:-test}-latest}",
        "${RADISHMIND_IMAGE_REGISTRY:-ghcr.io/laugh0608/radishmind}/console:${RADISHMIND_IMAGE_TAG:-${RADISHMIND_IMAGE_TRACK:-test}-latest}",
    ]:
        require(literal in deploy_compose, f"deploy/docker-compose.yaml missing {literal}")
    forbidden = ["build:", "dockerfile:"]
    present_forbidden = [literal for literal in forbidden if literal in deploy_compose]
    require(not present_forbidden, f"deploy compose must not build images: {present_forbidden}")

    env_example = read("deploy/.env.example")
    for literal in [
        "RADISHMIND_IMAGE_REGISTRY=ghcr.io/laugh0608/radishmind",
        "RADISHMIND_IMAGE_TRACK=test",
        "RADISHMIND_IMAGE_TAG=",
    ]:
        require(literal in env_example, f"deploy/.env.example missing {literal}")


def assert_publish_workflow_still_blocked() -> None:
    require(
        not (REPO_ROOT / FUTURE_WORKFLOW_PATH).exists(),
        "docker image publish workflow must not be added in image naming policy slice",
    )


def assert_consumers_and_docs(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected = {
        "scripts/check-production-ops-docker-image-build-publish.py",
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
        'run_python_script("check-production-ops-docker-image-build-publish.py", [])' in check_repo,
        "check-repo.py must run docker image build/publish policy check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")
        require(FORBIDDEN_LOCAL_REFERENCE not in text, f"{relative_path} must not commit local Radish path")


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "production_ops_docker_image_build_publish", "unexpected fixture kind")
    assert_fixture(fixture)
    assert_dockerfiles()
    assert_compose_image_names()
    assert_publish_workflow_still_blocked()
    assert_consumers_and_docs(fixture)
    require(FORBIDDEN_LOCAL_REFERENCE not in FIXTURE_PATH.read_text(encoding="utf-8"), "fixture must not commit local path")
    print("production ops docker image build/publish policy checks passed.")


if __name__ == "__main__":
    main()
