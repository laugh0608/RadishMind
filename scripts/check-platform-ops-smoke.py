#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
PLATFORM_DIR = REPO_ROOT / "services/platform"
BRIDGE_SCRIPT = REPO_ROOT / "scripts/run-platform-bridge.py"
REQUIRED_PROVIDERS = {"mock", "openai-compatible", "huggingface", "ollama"}
REQUIRED_PROFILE_MODELS = {
    "profile:anyrouter",
    "profile:backup",
    "provider:huggingface:profile:hf-chat",
    "provider:ollama:profile:local",
}


def run_command(args: list[str], *, cwd: Path, env: dict[str, str] | None = None) -> str:
    result = subprocess.run(
        args,
        cwd=cwd,
        env=env,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        if result.stdout:
            print(result.stdout, end="")
        if result.stderr:
            print(result.stderr, end="", file=sys.stderr)
        raise SystemExit(result.returncode)
    return result.stdout


def bridge_env() -> dict[str, str]:
    env = {key: value for key, value in os.environ.items() if not key.startswith("RADISHMIND_")}
    env.update(
        {
            "RADISHMIND_MODEL_API_STYLE": "",
            "RADISHMIND_MODEL_PROFILE": "anyrouter",
            "RADISHMIND_MODEL_PROFILE_FALLBACKS": "anyrouter,backup",
            "RADISHMIND_MODEL_PROFILE_ANYROUTER_NAME": "deepseek-chat",
            "RADISHMIND_MODEL_PROFILE_ANYROUTER_BASE_URL": "https://example.openai.invalid/v1",
            "RADISHMIND_MODEL_PROFILE_ANYROUTER_API_KEY": "openai-token",
            "RADISHMIND_MODEL_PROFILE_ANYROUTER_API_STYLE": "openai-compatible",
            "RADISHMIND_MODEL_PROFILE_BACKUP_NAME": "gemini-1.5-flash",
            "RADISHMIND_MODEL_PROFILE_BACKUP_BASE_URL": "https://generativelanguage.googleapis.com/v1beta",
            "RADISHMIND_MODEL_PROFILE_BACKUP_API_KEY": "gemini-token",
            "RADISHMIND_HUGGINGFACE_PROFILE": "hf-chat",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_CHAT_NAME": "meta-llama/Meta-Llama-3.1-8B-Instruct",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_CHAT_BASE_URL": "https://example.huggingface.invalid/v1",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_CHAT_API_KEY": "hf-token",
            "RADISHMIND_OLLAMA_PROFILE": "local",
            "RADISHMIND_OLLAMA_PROFILE_LOCAL_NAME": "qwen2.5:7b-instruct",
            "RADISHMIND_OLLAMA_PROFILE_LOCAL_BASE_URL": "http://localhost:11434",
            "RADISHMIND_OLLAMA_PROFILE_LOCAL_API_STYLE": "ollama-chat-completions",
        }
    )
    return env


def load_bridge_json(command: str, *, env: dict[str, str]) -> Any:
    output = run_command([sys.executable, str(BRIDGE_SCRIPT), command], cwd=REPO_ROOT, env=env)
    return json.loads(output)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def check_go_platform_tests() -> None:
    env = os.environ.copy()
    env.setdefault("GOCACHE", str(Path(tempfile.gettempdir()) / "radishmind-go-build-cache"))
    run_command(["go", "test", "./..."], cwd=PLATFORM_DIR, env=env)


def check_bridge_provider_registry(env: dict[str, str]) -> None:
    providers = load_bridge_json("providers", env=env)
    require(isinstance(providers, list), "platform bridge providers output must be a list")
    provider_ids = {str(item.get("provider_id") or "").strip() for item in providers if isinstance(item, dict)}
    missing = sorted(REQUIRED_PROVIDERS - provider_ids)
    require(not missing, f"platform bridge provider registry missing providers: {missing}")

    provider_by_id = {
        str(item.get("provider_id") or "").strip(): item
        for item in providers
        if isinstance(item, dict)
    }
    require(
        provider_by_id["openai-compatible"].get("profile_driven") is True,
        "openai-compatible provider must remain profile-driven",
    )
    require(
        "anthropic-messages" in (provider_by_id["openai-compatible"].get("supported_api_styles") or []),
        "openai-compatible provider must advertise anthropic messages compatibility",
    )
    for provider_id in ("huggingface", "ollama"):
        capabilities = provider_by_id[provider_id].get("capabilities") or {}
        require(capabilities.get("chat") is True, f"{provider_id} provider must advertise chat capability")
        require(capabilities.get("streaming") is True, f"{provider_id} provider must advertise streaming capability")


def check_bridge_inventory(env: dict[str, str]) -> None:
    inventory = load_bridge_json("inventory", env=env)
    require(isinstance(inventory, dict), "platform bridge inventory output must be an object")
    profiles = inventory.get("profiles")
    require(isinstance(profiles, list), "platform bridge inventory profiles must be a list")
    observed_model_ids = {
        provider_profile_model_id(str(item.get("provider_id") or ""), str(item.get("profile") or ""))
        for item in profiles
        if isinstance(item, dict)
    }
    missing_models = sorted(REQUIRED_PROFILE_MODELS - observed_model_ids)
    require(not missing_models, f"platform bridge inventory missing profile models: {missing_models}")

    active_chain = inventory.get("active_profile_chain")
    require(isinstance(active_chain, list), "platform bridge active_profile_chain must be a list")
    for model_id in ("profile:anyrouter", "provider:huggingface:profile:hf-chat", "provider:ollama:profile:local"):
        require(model_id in active_chain, f"platform bridge active profile chain missing {model_id}")

    profile_by_model_id = {
        provider_profile_model_id(str(item.get("provider_id") or ""), str(item.get("profile") or "")): item
        for item in profiles
        if isinstance(item, dict)
    }
    require(
        profile_by_model_id["profile:backup"].get("api_style") == "gemini-native",
        "openai-compatible backup profile should infer gemini-native api style from base url",
    )
    require(
        profile_by_model_id["provider:ollama:profile:local"].get("has_api_key") is False,
        "ollama local profile should not require api key",
    )
    require(
        profile_by_model_id["provider:huggingface:profile:hf-chat"].get("has_api_key") is True,
        "huggingface profile should expose api-key presence without leaking the key",
    )
    require(
        profile_by_model_id["provider:huggingface:profile:hf-chat"].get("credential_state") == "configured",
        "huggingface profile should expose configured credential state",
    )
    require(
        profile_by_model_id["provider:ollama:profile:local"].get("credential_state") == "optional_missing",
        "ollama local profile should expose optional_missing credential state",
    )
    require(
        profile_by_model_id["profile:anyrouter"].get("deployment_mode") == "remote_api",
        "openai-compatible profile should expose deployment mode",
    )
    require(
        profile_by_model_id["provider:ollama:profile:local"].get("deployment_mode") == "local_daemon",
        "ollama local profile should expose local daemon deployment mode",
    )
    for model_id, profile in profile_by_model_id.items():
        capabilities = profile.get("capabilities")
        require(isinstance(capabilities, dict), f"{model_id}: profile capabilities must be an object")
        require(capabilities.get("chat") is True, f"{model_id}: profile must expose chat capability")
        require("chat.completions" in (profile.get("northbound_protocols") or []), f"{model_id}: missing chat northbound protocol")
        require("/v1/chat/completions" in (profile.get("northbound_routes") or []), f"{model_id}: missing chat northbound route")


def provider_profile_model_id(provider_id: str, profile: str) -> str:
    normalized_provider = provider_id.strip()
    normalized_profile = profile.strip() or "default"
    if normalized_provider == "openai-compatible":
        return f"profile:{normalized_profile}"
    return f"provider:{normalized_provider}:profile:{normalized_profile}"


def main() -> int:
    env = bridge_env()
    check_go_platform_tests()
    check_bridge_provider_registry(env)
    check_bridge_inventory(env)
    print("platform ops smoke checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
