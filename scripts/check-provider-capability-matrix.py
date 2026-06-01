#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.provider_registry import describe_provider_registry  # noqa: E402


FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/provider-capability-matrix-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_PROVIDER_IDS = {
    "huggingface",
    "mock",
    "ollama",
    "openai-compatible",
}
EXPECTED_CAPABILITY_FIELDS = {
    "transport",
    "local_or_remote",
    "chat",
    "responses",
    "messages",
    "models_list",
    "streaming",
    "json_schema_output",
    "tool_calling",
    "image_input",
    "image_output",
    "auth_mode",
    "timeout_policy",
    "retry_policy",
    "cost_profile",
    "latency_profile",
    "deployment_mode",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "provider_health_ready",
    "optional_live_health_enabled_by_default",
    "production_readiness",
    "production_secret_backend_ready",
    "implicit_provider_fallback_ready",
    "tool_executor_ready",
    "confirmation_writeback_replay_ready",
}
EXPECTED_REQUIRED_CONSUMERS = {
    "scripts/check-provider-capability-matrix.py",
    "scripts/check-repo.py",
    "scripts/README.md",
    "docs/radishmind-current-focus.md",
    "docs/radishmind-capability-matrix.md",
    "docs/radishmind-roadmap.md",
    "docs/radishmind-architecture.md",
    "docs/task-cards/provider-runtime-health-v1-plan.md",
}
REQUIRED_DOC_REFERENCES = {
    "docs/radishmind-current-focus.md": [
        "provider-capability-matrix-v1",
        "provider-capability-matrix-v1.json",
        "check-provider-capability-matrix.py",
    ],
    "docs/radishmind-capability-matrix.md": [
        "provider capability matrix",
        "provider-capability-matrix-v1.json",
        "check-provider-capability-matrix.py",
    ],
    "docs/radishmind-roadmap.md": [
        "provider-capability-matrix-v1",
        "provider-capability-matrix-v1.json",
    ],
    "docs/radishmind-architecture.md": [
        "provider-capability-matrix-v1.json",
        "check-provider-capability-matrix.py",
    ],
    "docs/task-cards/provider-runtime-health-v1-plan.md": [
        "provider-capability-matrix-v1",
        "provider-capability-matrix-v1.json",
        "check-provider-capability-matrix.py",
    ],
    "scripts/README.md": [
        "check-provider-capability-matrix.py",
        "provider-capability-matrix-v1.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_fixture() -> dict[str, Any]:
    document = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    require(isinstance(document, dict), "provider capability matrix fixture must be an object")
    return document


def provider_profile_model_id(provider_id: str, profile: str) -> str:
    normalized_provider = provider_id.strip()
    normalized_profile = profile.strip() or "default"
    if normalized_provider == "openai-compatible":
        return f"profile:{normalized_profile}"
    return f"provider:{normalized_provider}:profile:{normalized_profile}"


def northbound_protocols(capabilities: dict[str, Any]) -> list[str]:
    protocols: list[str] = []
    if capabilities.get("chat") is True:
        protocols.append("chat.completions")
    if capabilities.get("responses") is True:
        protocols.append("responses")
    if capabilities.get("messages") is True:
        protocols.append("messages")
    return protocols


def northbound_routes(capabilities: dict[str, Any]) -> list[str]:
    routes = ["/v1/models"]
    if capabilities.get("chat") is True:
        routes.append("/v1/chat/completions")
    if capabilities.get("responses") is True:
        routes.append("/v1/responses")
    if capabilities.get("messages") is True:
        routes.append("/v1/messages")
    return routes


def profile_model_id_pattern(provider_id: str, profile_driven: bool) -> str:
    if not profile_driven:
        return "not_profile_driven"
    if provider_id == "openai-compatible":
        return "profile:<profile>"
    return f"provider:{provider_id}:profile:<profile>"


def assert_decision(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "provider_capability_matrix_v1", "unexpected fixture kind")
    decision = fixture.get("decision") or {}
    require(decision.get("id") == "provider-capability-matrix-v1", "unexpected decision id")
    require(decision.get("track") == "Provider Runtime & Health v1", "unexpected decision track")
    require(decision.get("status") == "satisfied", "provider capability matrix must be satisfied")
    require(
        decision.get("source_of_truth") == "services/runtime/provider_registry.py",
        "provider capability matrix must use provider_registry.py as source of truth",
    )
    forbidden_claims = set(decision.get("does_not_claim") or [])
    missing_forbidden_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - forbidden_claims)
    require(not missing_forbidden_claims, f"missing forbidden claims: {missing_forbidden_claims}")


def assert_capability_fields(fixture: dict[str, Any]) -> None:
    fields = set(fixture.get("capability_fields") or [])
    require(fields == EXPECTED_CAPABILITY_FIELDS, f"unexpected capability fields: {sorted(fields)}")


def assert_provider_matrix(fixture: dict[str, Any]) -> None:
    fixture_providers = fixture.get("providers")
    require(isinstance(fixture_providers, list), "providers must be a list")
    providers_by_id = {
        str(item.get("provider_id") or "").strip(): item
        for item in fixture_providers
        if isinstance(item, dict)
    }
    require(set(providers_by_id) == EXPECTED_PROVIDER_IDS, f"unexpected provider ids: {sorted(providers_by_id)}")

    registry_by_id = {
        str(item.get("provider_id") or "").strip(): item
        for item in describe_provider_registry()
        if isinstance(item, dict)
    }
    require(set(registry_by_id) == EXPECTED_PROVIDER_IDS, f"registry provider ids drifted: {sorted(registry_by_id)}")

    for provider_id in sorted(EXPECTED_PROVIDER_IDS):
        fixture_provider = providers_by_id[provider_id]
        registry_provider = registry_by_id[provider_id]
        for field in ("display_name", "default_api_style", "profile_driven"):
            require(
                fixture_provider.get(field) == registry_provider.get(field),
                f"{provider_id}: {field} does not match provider registry",
            )
        require(
            fixture_provider.get("supported_api_styles") == registry_provider.get("supported_api_styles"),
            f"{provider_id}: supported_api_styles does not match provider registry",
        )

        capabilities = fixture_provider.get("capabilities") or {}
        registry_capabilities = registry_provider.get("capabilities") or {}
        require(set(capabilities) == EXPECTED_CAPABILITY_FIELDS, f"{provider_id}: capability field set mismatch")
        require(
            capabilities == registry_capabilities,
            f"{provider_id}: capabilities do not match provider registry",
        )
        require(
            fixture_provider.get("northbound_protocols") == northbound_protocols(capabilities),
            f"{provider_id}: northbound protocols drifted",
        )
        require(
            fixture_provider.get("northbound_routes") == northbound_routes(capabilities),
            f"{provider_id}: northbound routes drifted",
        )
        require(
            fixture_provider.get("profile_model_id_pattern")
            == profile_model_id_pattern(provider_id, bool(fixture_provider.get("profile_driven"))),
            f"{provider_id}: profile model id pattern drifted",
        )


def assert_selection_identity_examples(fixture: dict[str, Any]) -> None:
    examples = fixture.get("selection_identity_examples")
    require(isinstance(examples, list) and examples, "selection identity examples must be a non-empty list")
    for example in examples:
        require(isinstance(example, dict), "selection identity example must be an object")
        provider_id = str(example.get("provider_id") or "").strip()
        profile = str(example.get("profile") or "").strip()
        expected_model_id = provider_profile_model_id(provider_id, profile)
        require(
            example.get("model_id") == expected_model_id,
            f"{provider_id}/{profile}: unexpected model id {example.get('model_id')}",
        )


def assert_execution_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("execution_policy") or {}
    require(policy.get("default_fast_baseline") == "offline_only", "fast baseline must remain offline only")
    require(policy.get("network_access") == "forbidden_by_default", "network access must be forbidden by default")
    require(policy.get("credential_required") is False, "default matrix check must not require credentials")
    require(policy.get("model_download") == "forbidden", "default matrix check must not download models")
    require(
        policy.get("optional_live_health") == "manual_only_future_slice",
        "optional live health must stay out of default matrix check",
    )
    require(
        policy.get("fallback_policy") == "disabled_unless_explicit_selection_policy_allows",
        "fallback policy must not become implicit",
    )


def assert_consumers_and_docs(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    missing_consumers = sorted(EXPECTED_REQUIRED_CONSUMERS - required_consumers)
    require(not missing_consumers, f"missing required consumers: {missing_consumers}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-provider-capability-matrix.py", [])' in check_repo,
        "check-repo.py must run provider capability matrix check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def main() -> int:
    fixture = load_fixture()
    assert_decision(fixture)
    assert_capability_fields(fixture)
    assert_provider_matrix(fixture)
    assert_selection_identity_examples(fixture)
    assert_execution_policy(fixture)
    assert_consumers_and_docs(fixture)
    print("provider capability matrix checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
