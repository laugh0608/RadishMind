#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/production-ops-docker-deployment-mode.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
FORBIDDEN_LOCAL_REFERENCE = "D:\\Code\\Radish"

REQUIRED_ENVIRONMENT_MODES = {
    "host_dev",
    "docker_local",
    "docker_test",
    "docker_prod",
}

REQUIRED_BLOCKED_CONDITIONS = {
    "docker_image_publish_workflow",
    "console_runtime_config",
    "production_secret_backend",
    "production_auth_cors_policy",
}

REQUIRED_FUTURE_ASSETS = {
    ".github/workflows/docker-images.yml",
}

REQUIRED_SATISFIED_CONDITIONS = {
    "docker_local_compose",
    "docker_test_prod_compose",
}

REQUIRED_DOC_REFERENCES = {
    "docs/task-cards/production-ops-docker-deployment-v1-plan.md": [
        "docker-deployment-mode-definition",
        "https://github.com/laugh0608/Radish",
        "docker_local",
        "docker_test",
        "docker_prod",
        "RADISHMIND_IMAGE_TRACK=test",
        "RADISHMIND_IMAGE_TRACK=release",
        "docker-test-prod-compose",
        "不写入长期文档",
    ],
    "docs/radishmind-current-focus.md": [
        "docker-deployment-mode-definition",
        "production-ops-docker-deployment-mode.json",
        "docker-local-compose",
        "production-ops-docker-local-compose.json",
        "docker-test-prod-compose",
        "production-ops-docker-test-prod-compose.json",
    ],
    "docs/radishmind-roadmap.md": [
        "docker-deployment-mode-definition",
        "production-ops-docker-deployment-v1-plan.md",
        "docker-local-compose",
        "production-ops-docker-local-compose.json",
        "docker-test-prod-compose",
        "production-ops-docker-test-prod-compose.json",
    ],
    "docs/task-cards/production-ops-hardening-v1-plan.md": [
        "docker-deployment-mode-definition",
        "docker-local-compose",
        "docker-test-prod-compose",
        "Production Ops Docker Deployment",
    ],
    "scripts/README.md": [
        "check-production-ops-docker-deployment-mode.py",
        "check-production-ops-docker-local-compose.py",
        "check-production-ops-docker-test-prod-compose.py",
        "docker local/test/prod",
        "production-ops-docker-deployment-mode.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_fixture() -> dict[str, Any]:
    return json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))


def ids(items: list[dict[str, Any]]) -> set[str]:
    return {str(item.get("id")) for item in items}


def assert_decision(fixture: dict[str, Any]) -> None:
    decision = fixture.get("decision") or {}
    require(decision.get("id") == "docker-deployment-mode-definition", "unexpected decision id")
    require(decision.get("track") == "Production Ops Hardening v1", "unexpected decision track")
    require(
        decision.get("status") == "deployment_mode_definition_satisfied",
        "docker deployment mode definition must be satisfied",
    )
    require(
        decision.get("selected_mode") == "radish_style_docker_local_test_prod",
        "unexpected selected deployment mode",
    )
    source_reference = decision.get("source_reference") or {}
    require(source_reference.get("project") == "Radish", "source reference must name Radish")
    require(source_reference.get("url") == "https://github.com/laugh0608/Radish", "source reference must use URL")
    require(
        source_reference.get("local_path_policy") == "do_not_commit_local_external_project_path",
        "local external project path must not be committed",
    )
    does_not_claim = set(decision.get("does_not_claim") or [])
    required_forbidden_claims = {
        "image_publish_ready",
        "production_ready",
        "production_secret_backend_ready",
        "process_supervisor_ready",
        "console_production_package_ready",
    }
    missing = sorted(required_forbidden_claims - does_not_claim)
    require(not missing, f"missing forbidden claims: {missing}")


def assert_environment_modes(fixture: dict[str, Any]) -> None:
    modes = fixture.get("environment_modes") or []
    modes_by_id = {str(item.get("id")): item for item in modes}
    missing = sorted(REQUIRED_ENVIRONMENT_MODES - set(modes_by_id))
    require(not missing, f"missing environment modes: {missing}")

    host_dev = modes_by_id["host_dev"]
    require(host_dev.get("status") == "supported", "host_dev must remain supported")
    require(host_dev.get("compose_use") == "forbidden_by_default", "host_dev must not default to Compose")

    docker_local = modes_by_id["docker_local"]
    require(
        docker_local.get("status") == "available_for_local_container_smoke",
        "docker_local must be available for local container smoke",
    )
    require(docker_local.get("compose_file") == "deploy/docker-compose.local.yaml", "unexpected local compose path")
    require(docker_local.get("image_source") == "local_build", "docker_local must use local build")
    require(docker_local.get("default_provider") == "mock", "docker_local must default to mock provider")
    require(docker_local.get("production_use") == "forbidden", "docker_local must be forbidden for production")

    docker_test = modes_by_id["docker_test"]
    require(
        docker_test.get("status") == "deploy_compose_boundary_available",
        "docker_test deploy compose boundary must be available",
    )
    require(docker_test.get("compose_file") == "deploy/docker-compose.yaml", "unexpected test compose path")
    require(docker_test.get("image_track") == "test", "docker_test must use test track")
    require("v*-test" in str(docker_test.get("image_tag_policy") or ""), "docker_test must allow fixed test tag")

    docker_prod = modes_by_id["docker_prod"]
    require(
        docker_prod.get("status") == "deploy_compose_boundary_available_but_not_ready",
        "docker_prod deploy compose boundary must be available but not ready",
    )
    require(docker_prod.get("compose_file") == "deploy/docker-compose.yaml", "unexpected prod compose path")
    require(docker_prod.get("image_track") == "release", "docker_prod must use release track")
    require("v*-release" in str(docker_prod.get("image_tag_policy") or ""), "docker_prod must allow fixed release tag")
    require(
        docker_prod.get("secret_policy") == "production_secret_backend_required_before_ready",
        "docker_prod must require production secret backend",
    )


def assert_future_assets_and_blocks(fixture: dict[str, Any]) -> None:
    satisfied = fixture.get("satisfied_conditions") or []
    satisfied_by_id = {str(item.get("id")): item for item in satisfied}
    missing_satisfied = sorted(REQUIRED_SATISFIED_CONDITIONS - set(satisfied_by_id))
    require(not missing_satisfied, f"missing satisfied conditions: {missing_satisfied}")
    local_compose = satisfied_by_id["docker_local_compose"]
    require(local_compose.get("status") == "satisfied", "docker_local_compose must be satisfied")
    local_evidence = set(local_compose.get("evidence") or [])
    for evidence in {
        "services/platform/Dockerfile",
        "apps/radishmind-console/Dockerfile",
        "apps/radishmind-console/nginx.local.conf",
        "deploy/docker-compose.local.yaml",
        "scripts/checks/fixtures/production-ops-docker-local-compose.json",
        "scripts/check-production-ops-docker-local-compose.py",
    }:
        require(evidence in local_evidence, f"docker_local_compose missing evidence: {evidence}")
        require((REPO_ROOT / evidence).exists(), f"docker_local_compose evidence missing on disk: {evidence}")

    test_prod_compose = satisfied_by_id["docker_test_prod_compose"]
    require(test_prod_compose.get("status") == "satisfied", "docker_test_prod_compose must be satisfied")
    test_prod_evidence = set(test_prod_compose.get("evidence") or [])
    for evidence in {
        "deploy/docker-compose.yaml",
        "deploy/.env.example",
        "scripts/checks/fixtures/production-ops-docker-test-prod-compose.json",
        "scripts/check-production-ops-docker-test-prod-compose.py",
    }:
        require(evidence in test_prod_evidence, f"docker_test_prod_compose missing evidence: {evidence}")
        require((REPO_ROOT / evidence).exists(), f"docker_test_prod_compose evidence missing on disk: {evidence}")

    future_assets = set(fixture.get("required_future_assets") or [])
    missing_assets = sorted(REQUIRED_FUTURE_ASSETS - future_assets)
    require(not missing_assets, f"missing future assets: {missing_assets}")

    blocked = fixture.get("blocked_conditions") or []
    blocked_by_id = {str(item.get("id")): item for item in blocked}
    missing_blocked = sorted(REQUIRED_BLOCKED_CONDITIONS - set(blocked_by_id))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    for condition_id in REQUIRED_BLOCKED_CONDITIONS:
        require(
            blocked_by_id[condition_id].get("status") == "not_satisfied",
            f"{condition_id} must stay not_satisfied",
        )


def assert_consumers_and_docs(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected_consumers = {
        "scripts/check-production-ops-docker-deployment-mode.py",
        "scripts/check-production-ops-docker-local-compose.py",
        "scripts/check-production-ops-docker-test-prod-compose.py",
        "scripts/check-repo.py",
        "scripts/README.md",
        "docs/radishmind-current-focus.md",
        "docs/radishmind-roadmap.md",
        "docs/task-cards/production-ops-hardening-v1-plan.md",
        "docs/task-cards/production-ops-docker-deployment-v1-plan.md",
    }
    missing_consumers = sorted(expected_consumers - required_consumers)
    require(not missing_consumers, f"missing consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-production-ops-docker-deployment-mode.py", [])' in check_repo,
        "check-repo.py must run docker deployment mode check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")
        require(FORBIDDEN_LOCAL_REFERENCE not in text, f"{relative_path} must not commit local Radish path")


def main() -> None:
    fixture = load_fixture()
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "production_ops_docker_deployment_mode", "unexpected fixture kind")
    assert_decision(fixture)
    assert_environment_modes(fixture)
    assert_future_assets_and_blocks(fixture)
    assert_consumers_and_docs(fixture)
    require(FORBIDDEN_LOCAL_REFERENCE not in FIXTURE_PATH.read_text(encoding="utf-8"), "fixture must not commit local path")
    print("production ops docker deployment mode checks passed.")


if __name__ == "__main__":
    main()
