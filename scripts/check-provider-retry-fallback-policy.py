#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(REPO_ROOT))

from services.runtime.provider_registry import describe_provider_registry

FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/provider-retry-fallback-policy-v1.json"
SELECTION_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/provider-selection-policy-v1.json"
HEALTH_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/provider-health-smoke-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
SERVER_TEST_PATH = REPO_ROOT / "services/platform/internal/httpapi/server_test.go"
NORTHBOUND_PATH = REPO_ROOT / "services/platform/internal/httpapi/northbound.go"
OBSERVABILITY_PATH = REPO_ROOT / "services/platform/internal/httpapi/observability.go"

EXPECTED_FORBIDDEN_CLAIMS = {
    "automatic_retry_ready",
    "implicit_provider_fallback_ready",
    "external_provider_live_health_ready",
    "live_timeout_probe_ready",
    "production_readiness",
    "production_secret_backend_ready",
    "tool_executor_ready",
    "confirmation_writeback_replay_ready",
}
EXPECTED_REQUIRED_CONSUMERS = {
    "services/platform/internal/httpapi/server_test.go",
    "scripts/check-provider-retry-fallback-policy.py",
    "scripts/check-repo.py",
    "scripts/README.md",
    "services/platform/README.md",
    "docs/contracts/service-api.md",
    "docs/radishmind-current-focus.md",
    "docs/radishmind-capability-matrix.md",
    "docs/radishmind-roadmap.md",
    "docs/radishmind-architecture.md",
    "docs/task-cards/provider-runtime-health-v1-plan.md",
    "docs/devlogs/2026-W22.md",
}
REQUIRED_DOC_REFERENCES = {
    "docs/radishmind-current-focus.md": [
        "provider-retry-fallback-policy-v1",
        "provider-retry-fallback-policy-v1.json",
        "check-provider-retry-fallback-policy.py",
    ],
    "docs/radishmind-capability-matrix.md": [
        "provider retry/fallback policy",
        "provider-retry-fallback-policy-v1.json",
        "check-provider-retry-fallback-policy.py",
    ],
    "docs/radishmind-roadmap.md": [
        "provider-retry-fallback-policy-v1",
        "provider-retry-fallback-policy-v1.json",
    ],
    "docs/radishmind-architecture.md": [
        "provider-retry-fallback-policy-v1.json",
        "check-provider-retry-fallback-policy.py",
    ],
    "docs/contracts/service-api.md": [
        "provider-retry-fallback-policy-v1",
        "caller-managed",
        "fallback_policy",
    ],
    "docs/task-cards/provider-runtime-health-v1-plan.md": [
        "provider-retry-fallback-policy-v1",
        "provider-retry-fallback-policy-v1.json",
        "check-provider-retry-fallback-policy.py",
    ],
    "services/platform/README.md": [
        "provider-retry-fallback-policy-v1",
        "retry_policy",
        "fallback_policy",
    ],
    "scripts/README.md": [
        "check-provider-retry-fallback-policy.py",
        "provider-retry-fallback-policy-v1.json",
    ],
    "docs/devlogs/2026-W22.md": [
        "provider-retry-fallback-policy-v1",
        "provider-retry-fallback-policy-v1.json",
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
    require(fixture.get("kind") == "provider_retry_fallback_policy_v1", "unexpected fixture kind")
    decision = fixture.get("decision") or {}
    require(decision.get("id") == "provider-retry-fallback-policy-v1", "unexpected decision id")
    require(decision.get("track") == "Provider Runtime & Health v1", "unexpected decision track")
    require(decision.get("status") == "satisfied", "retry/fallback policy must be satisfied")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(decision.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_audit_metadata(fixture: dict[str, Any]) -> None:
    audit_metadata = fixture.get("audit_metadata") or {}
    require(audit_metadata.get("retry_policy") == "caller-managed", "retry policy audit value must be caller-managed")
    require(audit_metadata.get("fallback_policy") == "disabled", "fallback policy audit value must be disabled")
    attached_to = set(audit_metadata.get("attached_to") or [])
    require("canonical_request.context.northbound" in attached_to, "audit metadata must attach to canonical northbound context")
    require("error.metadata" in attached_to, "audit metadata must attach to error metadata")

    northbound = NORTHBOUND_PATH.read_text(encoding="utf-8")
    for literal in (
        "retryPolicyCallerManaged",
        "fallbackPolicyDisabled",
        "\"retry_policy\"",
        "\"fallback_policy\"",
    ):
        require(literal in northbound, f"northbound.go missing audit metadata literal: {literal}")

    server_test = SERVER_TEST_PATH.read_text(encoding="utf-8")
    for literal in (
        "retry/fallback policy audit metadata",
        "retry_policy",
        "fallback_policy",
        "caller-managed",
        "disabled",
    ):
        require(literal in server_test, f"server_test.go missing retry/fallback coverage: {literal}")


def assert_retry_policy(fixture: dict[str, Any]) -> None:
    retry_policy = fixture.get("retry_policy") or {}
    require(retry_policy.get("automatic_retry") == "disabled", "automatic retry must stay disabled")
    require(retry_policy.get("max_attempts_by_platform") == 1, "platform max attempts must stay 1")
    require(retry_policy.get("caller_managed_retry") is True, "caller-managed retry must be explicit")
    require(retry_policy.get("background_retry") == "forbidden", "background retry must be forbidden")

    health_fixture = load_json(HEALTH_FIXTURE_PATH)
    health_taxonomy = set(health_fixture.get("failure_taxonomy") or [])
    for classification in ("live_provider_timeout", "live_provider_unavailable"):
        require(classification in health_taxonomy, f"health taxonomy missing retry signal classification: {classification}")

    observability = OBSERVABILITY_PATH.read_text(encoding="utf-8")
    for code in ("PLATFORM_BRIDGE_FAILED", "PROVIDER_INVENTORY_UNAVAILABLE", "MODEL_NOT_FOUND"):
        require(code in observability, f"observability.go missing retry policy error code: {code}")


def assert_fallback_policy(fixture: dict[str, Any]) -> None:
    fallback_policy = fixture.get("fallback_policy") or {}
    require(fallback_policy.get("implicit_fallback") == "forbidden", "implicit fallback must remain forbidden")
    require(
        fallback_policy.get("profile_chain_use") == "discovery_only_not_execution_chain",
        "profile chain must stay discovery-only",
    )
    require(fallback_policy.get("current_allowed_fallback_cases") == [], "current allowed fallback cases must be empty")

    selection_fixture = load_json(SELECTION_FIXTURE_PATH)
    selection_fallback = selection_fixture.get("fallback_policy") or {}
    require(selection_fallback.get("implicit_fallback") == "forbidden", "selection policy must still forbid fallback")
    require(
        selection_fallback.get("current_profile_chain_use") == "discovery_only_not_auto_retry",
        "selection policy profile chain must remain discovery-only",
    )
    for case in selection_fixture.get("negative_selection_cases") or []:
        require(case.get("fallback_allowed") is False, f"{case.get('id')}: fallback must remain disabled")


def assert_provider_registry_alignment() -> None:
    for provider in describe_provider_registry():
        provider_id = str(provider.get("provider_id") or "")
        capabilities = provider.get("capabilities") or {}
        retry_policy = capabilities.get("retry_policy")
        if provider_id == "mock":
            require(retry_policy == "none", "mock provider retry policy must stay none")
        else:
            require(retry_policy == "caller-managed", f"{provider_id}: retry policy must stay caller-managed")


def assert_execution_policy_and_consumers(fixture: dict[str, Any]) -> None:
    execution_policy = fixture.get("execution_policy") or {}
    require(execution_policy.get("network_access") == "forbidden_by_default", "network access must be forbidden")
    require(execution_policy.get("credential_required") is False, "check must not require credentials")
    require(execution_policy.get("model_download") == "forbidden", "check must not download models")
    require(execution_policy.get("default_check_executes_retry") is False, "check must not execute retry")
    require(execution_policy.get("default_check_executes_fallback") is False, "check must not execute fallback")

    missing_consumers = sorted(EXPECTED_REQUIRED_CONSUMERS - set(fixture.get("required_consumers") or []))
    require(not missing_consumers, f"missing required consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-provider-retry-fallback-policy.py", [])' in check_repo,
        "check-repo.py must run provider retry/fallback policy check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def main() -> int:
    fixture = load_json(FIXTURE_PATH)
    assert_decision(fixture)
    assert_audit_metadata(fixture)
    assert_retry_policy(fixture)
    assert_fallback_policy(fixture)
    assert_provider_registry_alignment()
    assert_execution_policy_and_consumers(fixture)
    print("provider retry/fallback policy checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
