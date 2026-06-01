#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import sys
from pathlib import Path
from typing import Any
from unittest.mock import patch

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime import inference_provider, inference_support  # noqa: E402
from services.runtime.inference_provider import run_inference  # noqa: E402
from services.runtime.inference_support import (  # noqa: E402
    build_provider_profile_inventory_model_id,
    describe_provider_inventory,
)


FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/provider-health-smoke-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
SAMPLE_REQUEST_PATH = REPO_ROOT / "datasets/eval/radish/answer-docs-question-direct-answer-001.json"
MISSING_ENV_PATH = REPO_ROOT / "__provider_health_smoke_missing_env__.env"

EXPECTED_LAYER_IDS = {
    "mock_runtime_smoke",
    "config_level_inventory_smoke",
    "optional_live_health",
}
EXPECTED_FAILURE_TAXONOMY = {
    "healthy_offline",
    "config_ready_not_live_checked",
    "config_ready_optional_credential_missing",
    "config_blocked_credential_missing",
    "config_blocked_base_url_missing",
    "live_provider_unavailable",
    "live_provider_timeout",
    "live_provider_auth_failed",
    "unsupported_capability",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "optional_live_health_enabled_by_default",
    "external_provider_reachable",
    "production_readiness",
    "production_secret_backend_ready",
    "process_supervisor_ready",
    "implicit_provider_fallback_ready",
    "tool_executor_ready",
}
EXPECTED_REQUIRED_CONSUMERS = {
    "scripts/check-provider-health-smoke.py",
    "scripts/check-repo.py",
    "scripts/README.md",
    "docs/radishmind-current-focus.md",
    "docs/radishmind-capability-matrix.md",
    "docs/radishmind-roadmap.md",
    "docs/radishmind-architecture.md",
    "docs/task-cards/provider-runtime-health-v1-plan.md",
    "docs/devlogs/2026-W22.md",
}
REQUIRED_DOC_REFERENCES = {
    "docs/radishmind-current-focus.md": [
        "provider-health-smoke-v1",
        "provider-health-smoke-v1.json",
        "check-provider-health-smoke.py",
    ],
    "docs/radishmind-capability-matrix.md": [
        "provider health smoke",
        "provider-health-smoke-v1.json",
        "check-provider-health-smoke.py",
    ],
    "docs/radishmind-roadmap.md": [
        "provider-health-smoke-v1",
        "provider-health-smoke-v1.json",
    ],
    "docs/radishmind-architecture.md": [
        "provider-health-smoke-v1.json",
        "check-provider-health-smoke.py",
    ],
    "docs/task-cards/provider-runtime-health-v1-plan.md": [
        "provider-health-smoke-v1",
        "provider-health-smoke-v1.json",
        "check-provider-health-smoke.py",
    ],
    "scripts/README.md": [
        "check-provider-health-smoke.py",
        "provider-health-smoke-v1.json",
    ],
    "docs/devlogs/2026-W22.md": [
        "provider-health-smoke-v1",
        "provider-health-smoke-v1.json",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_fixture() -> dict[str, Any]:
    document = json.loads(FIXTURE_PATH.read_text(encoding="utf-8"))
    require(isinstance(document, dict), "provider health smoke fixture must be an object")
    return document


def layer_by_id(fixture: dict[str, Any]) -> dict[str, dict[str, Any]]:
    layers = fixture.get("health_layers")
    require(isinstance(layers, list), "health_layers must be a list")
    mapped = {
        str(layer.get("id") or "").strip(): layer
        for layer in layers
        if isinstance(layer, dict)
    }
    require(set(mapped) == EXPECTED_LAYER_IDS, f"unexpected health layer ids: {sorted(mapped)}")
    return mapped


def health_smoke_env() -> dict[str, str]:
    return {
        "RADISHMIND_MODEL_PROFILE": "ready",
        "RADISHMIND_MODEL_PROFILE_FALLBACKS": "ready,missing",
        "RADISHMIND_MODEL_PROFILE_READY_NAME": "deepseek-chat",
        "RADISHMIND_MODEL_PROFILE_READY_BASE_URL": "https://example.openai.invalid/v1",
        "RADISHMIND_MODEL_PROFILE_READY_API_KEY": "ready-token",
        "RADISHMIND_MODEL_PROFILE_READY_API_STYLE": "openai-compatible",
        "RADISHMIND_MODEL_PROFILE_MISSING_NAME": "gemini-1.5-flash",
        "RADISHMIND_MODEL_PROFILE_MISSING_BASE_URL": "https://generativelanguage.googleapis.com/v1beta",
        "RADISHMIND_HUGGINGFACE_PROFILE": "hf-missing",
        "RADISHMIND_HUGGINGFACE_PROFILE_HF_MISSING_NAME": "meta-llama/Meta-Llama-3.1-8B-Instruct",
        "RADISHMIND_HUGGINGFACE_PROFILE_HF_MISSING_BASE_URL": "https://example.huggingface.invalid/v1",
        "RADISHMIND_OLLAMA_PROFILE": "local",
        "RADISHMIND_OLLAMA_PROFILE_LOCAL_NAME": "qwen2.5:7b-instruct",
        "RADISHMIND_OLLAMA_PROFILE_LOCAL_BASE_URL": "http://localhost:11434",
        "RADISHMIND_OLLAMA_PROFILE_LOCAL_API_STYLE": "ollama-chat-completions",
    }


def classify_profile_health(profile: dict[str, Any]) -> str:
    credential_state = str(profile.get("credential_state") or "").strip()
    has_base_url = profile.get("has_base_url") is True
    if credential_state == "missing":
        return "config_blocked_credential_missing"
    if not has_base_url:
        return "config_blocked_base_url_missing"
    if credential_state == "optional_missing":
        return "config_ready_optional_credential_missing"
    if credential_state == "configured":
        return "config_ready_not_live_checked"
    return "unsupported_capability"


def assert_decision(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "provider_health_smoke_v1", "unexpected fixture kind")
    decision = fixture.get("decision") or {}
    require(decision.get("id") == "provider-health-smoke-v1", "unexpected decision id")
    require(decision.get("track") == "Provider Runtime & Health v1", "unexpected decision track")
    require(decision.get("status") == "satisfied", "provider health smoke must be satisfied")
    require(decision.get("default_mode") == "offline_health_smoke", "default mode must be offline health smoke")
    forbidden_claims = set(decision.get("does_not_claim") or [])
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - forbidden_claims)
    require(not missing, f"missing forbidden claims: {missing}")


def assert_mock_runtime_smoke(layer: dict[str, Any]) -> None:
    require(layer.get("status") == "satisfied", "mock runtime smoke must be satisfied")
    require(layer.get("default_fast_baseline") is True, "mock runtime smoke must be in fast baseline")
    require(layer.get("provider") == "mock", "mock runtime smoke must use mock provider")
    require(layer.get("sample") == "datasets/eval/radish/answer-docs-question-direct-answer-001.json", "unexpected mock smoke sample")
    require(layer.get("expected_classification") == "healthy_offline", "unexpected mock health classification")
    require(layer.get("network_access") == "forbidden", "mock health smoke must forbid network")
    require(layer.get("credential_required") is False, "mock health smoke must not require credentials")

    sample = json.loads(SAMPLE_REQUEST_PATH.read_text(encoding="utf-8"))
    request_document = sample.get("input_request")
    golden_response = sample.get("golden_response")
    require(isinstance(request_document, dict), "mock health sample missing input_request")
    require(isinstance(golden_response, dict), "mock health sample missing golden_response")

    def fail_urlopen(*_args: Any, **_kwargs: Any) -> None:
        raise AssertionError("mock provider health smoke must not access network")

    with patch.object(inference_support, "ENV_FILE_PATH", MISSING_ENV_PATH):
        with patch.object(inference_provider.request, "urlopen", side_effect=fail_urlopen):
            result = run_inference(request_document, provider="mock")

    require(result.get("provider") == "mock", "mock health smoke returned unexpected provider")
    response_summary = str(result.get("response", {}).get("summary") or "")
    expected_summary = str(golden_response.get("summary") or "")
    require(expected_summary in response_summary, "mock health smoke summary drifted")


def assert_config_level_inventory_smoke(layer: dict[str, Any]) -> None:
    require(layer.get("status") == "satisfied", "config-level inventory smoke must be satisfied")
    require(layer.get("default_fast_baseline") is True, "config-level inventory smoke must be in fast baseline")
    require(layer.get("network_access") == "forbidden", "config-level inventory smoke must forbid network")
    require(layer.get("credential_required") is False, "config-level inventory smoke must not require credentials")

    expected_profiles = layer.get("expected_profiles")
    require(isinstance(expected_profiles, list) and expected_profiles, "expected_profiles must be a non-empty list")

    with patch.dict(os.environ, health_smoke_env(), clear=True):
        with patch.object(inference_support, "ENV_FILE_PATH", MISSING_ENV_PATH):
            inventory = describe_provider_inventory()

    observed_profiles = {}
    for profile in inventory.get("profiles") or []:
        if not isinstance(profile, dict):
            continue
        model_id = build_provider_profile_inventory_model_id(
            str(profile.get("provider_id") or ""),
            str(profile.get("profile") or ""),
        )
        observed_profiles[model_id] = profile

    for expected in expected_profiles:
        require(isinstance(expected, dict), "expected profile entry must be an object")
        model_id = str(expected.get("model_id") or "").strip()
        observed = observed_profiles.get(model_id)
        require(observed is not None, f"missing observed provider profile: {model_id}")
        for field in ("provider_id", "profile", "credential_state", "deployment_mode"):
            require(
                observed.get(field) == expected.get(field),
                f"{model_id}: unexpected {field}: {observed.get(field)}",
            )
        classification = classify_profile_health(observed)
        require(
            classification == expected.get("expected_classification"),
            f"{model_id}: unexpected health classification {classification}",
        )


def assert_optional_live_health(layer: dict[str, Any]) -> None:
    require(layer.get("status") == "manual_only_future_slice", "optional live health must stay manual only")
    require(layer.get("default_fast_baseline") is False, "optional live health must not enter fast baseline")
    require(layer.get("record_root") == "tmp/provider-health-smoke/", "unexpected optional live health record root")
    require(
        layer.get("failure_policy") == "live_probe_failure_is_provider_health_signal_not_production_outage",
        "optional live health failure policy drifted",
    )
    required_inputs = set(layer.get("required_inputs") or [])
    for input_name in {
        "provider",
        "provider_profile",
        "base_url",
        "credential_source",
        "request_timeout_seconds",
        "output_record_path",
    }:
        require(input_name in required_inputs, f"optional live health missing required input: {input_name}")


def assert_taxonomy_and_policy(fixture: dict[str, Any]) -> None:
    taxonomy = set(fixture.get("failure_taxonomy") or [])
    require(taxonomy == EXPECTED_FAILURE_TAXONOMY, f"unexpected failure taxonomy: {sorted(taxonomy)}")

    policy = fixture.get("execution_policy") or {}
    require(policy.get("default_fast_baseline") == "mock_and_config_only", "fast baseline must stay mock/config only")
    require(policy.get("network_access") == "forbidden_by_default", "network access must be forbidden by default")
    require(policy.get("credential_required") is False, "default health smoke must not require credentials")
    require(policy.get("model_download") == "forbidden", "default health smoke must not download models")
    require(policy.get("optional_live_health") == "manual_only", "optional live health must be manual only")
    require(
        policy.get("live_health_result_scope") == "provider_health_signal_only",
        "live health must not become production readiness",
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
        'run_python_script("check-provider-health-smoke.py", [])' in check_repo,
        "check-repo.py must run provider health smoke check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def main() -> int:
    fixture = load_fixture()
    assert_decision(fixture)
    layers = layer_by_id(fixture)
    assert_mock_runtime_smoke(layers["mock_runtime_smoke"])
    assert_config_level_inventory_smoke(layers["config_level_inventory_smoke"])
    assert_optional_live_health(layers["optional_live_health"])
    assert_taxonomy_and_policy(fixture)
    assert_consumers_and_docs(fixture)
    print("provider health smoke checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
