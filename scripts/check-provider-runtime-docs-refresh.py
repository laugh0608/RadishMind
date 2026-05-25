#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/provider-runtime-docs-refresh.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_SLICE_IDS = {
    "provider-capability-matrix-v1",
    "provider-health-smoke-v1",
    "provider-selection-policy-v1",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "optional_live_health_enabled_by_default",
    "external_provider_live_health_ready",
    "retry_policy_ready",
    "implicit_provider_fallback_ready",
    "production_readiness",
    "production_secret_backend_ready",
    "tool_executor_ready",
    "confirmation_writeback_replay_ready",
}
EXPECTED_REMAINING_BOUNDARIES = {
    "optional_live_health_not_default",
    "external_provider_live_health_not_ready",
    "production_secret_backend_not_ready",
    "real_retry_fallback_policy_not_ready",
    "live_timeout_probe_not_ready",
    "production_readiness_not_ready",
    "tool_executor_confirmation_writeback_replay_not_ready",
}
EXPECTED_REQUIRED_DOCUMENTS = {
    "docs/README.md",
    "docs/radishmind-project-guide.md",
    "docs/radishmind-current-focus.md",
    "docs/radishmind-capability-matrix.md",
    "docs/radishmind-roadmap.md",
    "docs/radishmind-architecture.md",
    "docs/task-cards/provider-runtime-health-v1-plan.md",
    "scripts/README.md",
    "docs/devlogs/2026-W22.md",
}
EXPECTED_REQUIRED_CONSUMERS = {
    "scripts/check-provider-runtime-docs-refresh.py",
    "scripts/check-repo.py",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "provider-runtime-docs-refresh",
        "Provider Runtime & Health v1",
        "provider-capability-matrix-v1",
        "provider-health-smoke-v1",
        "provider-selection-policy-v1",
        "不把 provider health 写成 production readiness",
    ],
    "docs/radishmind-project-guide.md": [
        "provider-runtime-docs-refresh",
        "Provider Runtime & Health v1",
        "provider-capability-matrix-v1",
        "provider-health-smoke-v1",
        "provider-selection-policy-v1",
        "不继续默认新增 provider 同层小切片",
    ],
    "docs/radishmind-current-focus.md": [
        "provider-runtime-docs-refresh",
        "Provider Runtime & Health v1",
        "check-provider-runtime-docs-refresh.py",
        "进入 close candidate",
        "不继续默认新增 provider 同层小切片",
    ],
    "docs/radishmind-capability-matrix.md": [
        "provider-runtime-docs-refresh",
        "check-provider-runtime-docs-refresh.py",
        "真实 retry/fallback",
        "optional live health",
    ],
    "docs/radishmind-roadmap.md": [
        "provider-runtime-docs-refresh",
        "Provider Runtime & Health v1",
        "进入 close candidate",
        "retry/fallback",
    ],
    "docs/radishmind-architecture.md": [
        "provider-runtime-docs-refresh.json",
        "check-provider-runtime-docs-refresh.py",
        "外部 provider live health",
        "真实 retry/fallback",
    ],
    "docs/task-cards/provider-runtime-health-v1-plan.md": [
        "provider-runtime-docs-refresh",
        "provider-runtime-docs-refresh.json",
        "check-provider-runtime-docs-refresh.py",
        "进入 close candidate",
    ],
    "scripts/README.md": [
        "check-provider-runtime-docs-refresh.py",
        "provider-runtime-docs-refresh.json",
    ],
    "docs/devlogs/2026-W22.md": [
        "provider-runtime-docs-refresh",
        "provider-runtime-docs-refresh.json",
        "check-provider-runtime-docs-refresh.py",
    ],
}
FORBIDDEN_STALE_DOC_REFERENCES = {
    "docs/README.md": [
        "优先补 provider capability matrix、health smoke 和 selection policy",
    ],
    "docs/radishmind-project-guide.md": [
        "优先从 `provider-capability-matrix-v1` 开始",
        "还缺真实镜像发布 workflow、container smoke 运行记录、production secret backend、部署隔离、外部 provider health check、provider capability matrix、provider selection policy",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be an object")
    return document


def assert_decision(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "provider_runtime_docs_refresh", "unexpected fixture kind")
    decision = fixture.get("decision") or {}
    require(decision.get("id") == "provider-runtime-docs-refresh", "unexpected decision id")
    require(decision.get("track") == "Provider Runtime & Health v1", "unexpected decision track")
    require(decision.get("status") == "satisfied", "provider runtime docs refresh must be satisfied")
    forbidden_claims = set(decision.get("does_not_claim") or [])
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - forbidden_claims)
    require(not missing, f"missing forbidden claims: {missing}")


def assert_completed_slices(fixture: dict[str, Any]) -> None:
    slices = fixture.get("completed_slices")
    require(isinstance(slices, list), "completed_slices must be a list")
    mapped = {
        str(item.get("id") or "").strip(): item
        for item in slices
        if isinstance(item, dict)
    }
    require(set(mapped) == EXPECTED_SLICE_IDS, f"unexpected completed slice ids: {sorted(mapped)}")

    for slice_id, item in mapped.items():
        fixture_path = str(item.get("fixture") or "").strip()
        checker_path = str(item.get("checker") or "").strip()
        boundary = str(item.get("boundary") or "").strip()
        require(fixture_path, f"{slice_id}: missing fixture path")
        require(checker_path, f"{slice_id}: missing checker path")
        require(boundary, f"{slice_id}: missing boundary")
        require((REPO_ROOT / fixture_path).is_file(), f"{slice_id}: missing fixture file {fixture_path}")
        require((REPO_ROOT / checker_path).is_file(), f"{slice_id}: missing checker file {checker_path}")
        check_document = load_json(REPO_ROOT / fixture_path)
        decision = check_document.get("decision") or {}
        require(decision.get("id") == slice_id, f"{slice_id}: fixture decision id drifted")
        require(decision.get("status") == "satisfied", f"{slice_id}: fixture must remain satisfied")


def assert_boundaries_and_consumers(fixture: dict[str, Any]) -> None:
    remaining_boundaries = set(fixture.get("remaining_boundaries") or [])
    missing_boundaries = sorted(EXPECTED_REMAINING_BOUNDARIES - remaining_boundaries)
    require(not missing_boundaries, f"missing remaining boundaries: {missing_boundaries}")

    required_documents = set(fixture.get("required_documents") or [])
    missing_documents = sorted(EXPECTED_REQUIRED_DOCUMENTS - required_documents)
    require(not missing_documents, f"missing required documents: {missing_documents}")
    for relative_path in required_documents:
        require((REPO_ROOT / relative_path).is_file(), f"missing required document: {relative_path}")

    required_consumers = set(fixture.get("required_consumers") or [])
    missing_consumers = sorted(EXPECTED_REQUIRED_CONSUMERS - required_consumers)
    require(not missing_consumers, f"missing required consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-provider-runtime-docs-refresh.py", [])' in check_repo,
        "check-repo.py must run provider runtime docs refresh check",
    )


def assert_docs_refresh() -> None:
    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")

    for relative_path, stale_literals in FORBIDDEN_STALE_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        stale = [literal for literal in stale_literals if literal in text]
        require(not stale, f"{relative_path} still has stale provider docs text: {stale}")


def main() -> int:
    fixture = load_json(FIXTURE_PATH)
    assert_decision(fixture)
    assert_completed_slices(fixture)
    assert_boundaries_and_consumers(fixture)
    assert_docs_refresh()
    print("provider runtime docs refresh checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
