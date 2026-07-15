#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/provider-selection-policy-v1.json"
HEALTH_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/provider-health-smoke-v1.json"
CAPABILITY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/provider-capability-matrix-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
SERVER_TEST_PATH = REPO_ROOT / "services/platform/internal/httpapi/server_test.go"
NORTHBOUND_PATH = REPO_ROOT / "services/platform/internal/httpapi/northbound.go"
MODELS_PATH = REPO_ROOT / "services/platform/internal/httpapi/models.go"

EXPECTED_POSITIVE_CASE_IDS = {
    "profile-model-inventory",
    "provider-alias-active-profile",
}
EXPECTED_NEGATIVE_CASE_IDS = {
    "unknown-concrete-model-runtime-override",
    "unknown-explicit-profile-no-fallback",
    "model-detail-unknown",
    "credential-missing",
    "unsupported-schema-tool-capability",
    "timeout",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "implicit_provider_fallback_ready",
    "retry_policy_ready",
    "live_provider_timeout_probe_ready",
    "production_readiness",
    "tool_executor_ready",
    "confirmation_writeback_replay_ready",
}
EXPECTED_REQUIRED_CONSUMERS = {
    "services/platform/internal/httpapi/server_test.go",
    "scripts/check-provider-selection-policy.py",
    "scripts/check-repo.py",
    "scripts/README.md",
    "docs/radishmind-capability-matrix.md",
    "docs/radishmind-roadmap.md",
    "docs/radishmind-architecture.md",
    "docs/task-cards/provider-runtime-health-v1-plan.md",
    "docs/devlogs/2026-W22.md",
}
REQUIRED_DOC_REFERENCES = {
    "docs/radishmind-capability-matrix.md": [
        "provider selection policy",
        "provider-selection-policy-v1.json",
        "check-provider-selection-policy.py",
    ],
    "docs/radishmind-roadmap.md": [
        "provider-selection-policy-v1",
        "provider-selection-policy-v1.json",
    ],
    "docs/radishmind-architecture.md": [
        "provider-selection-policy-v1.json",
        "check-provider-selection-policy.py",
    ],
    "docs/task-cards/provider-runtime-health-v1-plan.md": [
        "provider-selection-policy-v1",
        "provider-selection-policy-v1.json",
        "check-provider-selection-policy.py",
    ],
    "scripts/README.md": [
        "check-provider-selection-policy.py",
        "provider-selection-policy-v1.json",
    ],
    "docs/devlogs/2026-W22.md": [
        "provider-selection-policy-v1",
        "provider-selection-policy-v1.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be an object")
    return document


def items_by_id(items: Any, *, label: str) -> dict[str, dict[str, Any]]:
    require(isinstance(items, list), f"{label} must be a list")
    mapped = {str(item.get("id") or "").strip(): item for item in items if isinstance(item, dict)}
    require("" not in mapped, f"{label} contains an item without id")
    return mapped


def assert_decision(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "provider_selection_policy_v1", "unexpected fixture kind")
    decision = fixture.get("decision") or {}
    require(decision.get("id") == "provider-selection-policy-v1", "unexpected decision id")
    require(decision.get("track") == "Provider Runtime & Health v1", "unexpected decision track")
    require(decision.get("status") == "satisfied", "provider selection policy must be satisfied")
    forbidden_claims = set(decision.get("does_not_claim") or [])
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - forbidden_claims)
    require(not missing, f"missing forbidden claims: {missing}")


def assert_selection_cases(fixture: dict[str, Any]) -> None:
    positive_cases = items_by_id(fixture.get("positive_selection_cases"), label="positive_selection_cases")
    negative_cases = items_by_id(fixture.get("negative_selection_cases"), label="negative_selection_cases")
    require(set(positive_cases) == EXPECTED_POSITIVE_CASE_IDS, f"unexpected positive cases: {sorted(positive_cases)}")
    require(set(negative_cases) == EXPECTED_NEGATIVE_CASE_IDS, f"unexpected negative cases: {sorted(negative_cases)}")

    require(
        positive_cases["profile-model-inventory"].get("expected_inventory_kind") == "provider_profile",
        "profile model must resolve through provider inventory",
    )
    require(
        positive_cases["provider-alias-active-profile"].get("expected_model")
        == "provider:huggingface:profile:hf-chat",
        "provider alias must resolve to active profile model id",
    )

    for case_id, policy_case in negative_cases.items():
        require(policy_case.get("fallback_allowed") is False, f"{case_id}: fallback must remain disabled")

    require(
        negative_cases["unknown-concrete-model-runtime-override"].get("expected_inventory_kind") == "runtime_override",
        "unknown concrete model must be runtime_override, not inventory match",
    )
    require(
        negative_cases["unknown-explicit-profile-no-fallback"].get("expected_source") == "radishmind.provider_profile",
        "unknown explicit profile source must stay explicit",
    )
    require(
        negative_cases["model-detail-unknown"].get("expected_error_code") == "MODEL_NOT_FOUND",
        "unknown model detail must return MODEL_NOT_FOUND",
    )
    require(
        negative_cases["credential-missing"].get("source_fixture") == "provider-health-smoke-v1.json",
        "credential missing policy must reuse provider health smoke fixture",
    )
    require(
        negative_cases["timeout"].get("default_fast_baseline") is False,
        "timeout policy must not require live timeout probe in fast baseline",
    )


def assert_cross_fixture_alignment(fixture: dict[str, Any]) -> None:
    health_fixture = load_json(HEALTH_FIXTURE_PATH)
    capability_fixture = load_json(CAPABILITY_FIXTURE_PATH)

    health_taxonomy = set(health_fixture.get("failure_taxonomy") or [])
    require("config_blocked_credential_missing" in health_taxonomy, "health fixture must define credential missing classification")
    require("unsupported_capability" in health_taxonomy, "health fixture must define unsupported capability classification")

    unsupported_case = items_by_id(
        fixture.get("negative_selection_cases"),
        label="negative_selection_cases",
    )["unsupported-schema-tool-capability"]
    unsupported_capabilities = set(unsupported_case.get("unsupported_capabilities") or [])
    provider_capabilities = capability_fixture.get("capability_fields") or []
    missing = sorted(unsupported_capabilities - set(provider_capabilities))
    require(not missing, f"unsupported capabilities missing from capability matrix: {missing}")


def assert_source_coverage() -> None:
    server_test = SERVER_TEST_PATH.read_text(encoding="utf-8")
    for literal in (
        "provider selection policy",
        "profile model uses inventory",
        "provider alias uses active profile",
        "unknown concrete model stays runtime override",
        "unknown explicit profile does not fallback",
    ):
        require(literal in server_test, f"server_test.go missing selection policy coverage: {literal}")

    northbound = NORTHBOUND_PATH.read_text(encoding="utf-8")
    for literal in (
        "requested_profile_model",
        "requested_provider_model+inventory",
        "requested_concrete_model",
        "radishmind.provider_profile",
        "runtime_override",
    ):
        require(literal in northbound, f"northbound.go missing selection source literal: {literal}")

    models = MODELS_PATH.read_text(encoding="utf-8")
    require("MODEL_NOT_FOUND" in models, "models.go must keep MODEL_NOT_FOUND detail route boundary")


def assert_policy_and_consumers(fixture: dict[str, Any]) -> None:
    fallback_policy = fixture.get("fallback_policy") or {}
    require(fallback_policy.get("implicit_fallback") == "forbidden", "implicit fallback must remain forbidden")
    require(
        fallback_policy.get("current_profile_chain_use") == "discovery_only_not_auto_retry",
        "profile chain must not become auto retry",
    )
    require(fallback_policy.get("current_retry_policy") == "caller-managed", "retry policy must remain caller-managed")

    execution_policy = fixture.get("execution_policy") or {}
    require(
        execution_policy.get("default_fast_baseline") == "offline_policy_and_go_unit_tests",
        "fast baseline must stay offline policy and Go unit tests",
    )
    require(execution_policy.get("network_access") == "forbidden_by_default", "network access must be forbidden by default")
    require(execution_policy.get("credential_required") is False, "selection policy check must not require credentials")
    require(execution_policy.get("model_download") == "forbidden", "selection policy check must not download models")

    required_consumers = set(fixture.get("required_consumers") or [])
    missing_consumers = sorted(EXPECTED_REQUIRED_CONSUMERS - required_consumers)
    require(not missing_consumers, f"missing required consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-provider-selection-policy.py", [])' in check_repo,
        "check-repo.py must run provider selection policy check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def main() -> int:
    fixture = load_json(FIXTURE_PATH)
    assert_decision(fixture)
    assert_selection_cases(fixture)
    assert_cross_fixture_alignment(fixture)
    assert_source_coverage()
    assert_policy_and_consumers(fixture)
    print("provider selection policy checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
