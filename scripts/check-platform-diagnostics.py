#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import platform
import subprocess
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
PLATFORM_DIR = REPO_ROOT / "services/platform"
PLATFORM_WRAPPER_SH = REPO_ROOT / "scripts/run-platform-service.sh"
PLATFORM_WRAPPER_PS1 = REPO_ROOT / "scripts/run-platform-service.ps1"
SECRET_SENTINEL = "diagnostics-secret-sentinel"


def base_env() -> dict[str, str]:
    env = {key: value for key, value in os.environ.items() if not key.startswith("RADISHMIND_")}
    env.setdefault("GOCACHE", str(platform_temp_dir() / "radishmind-go-build-cache"))
    env.update(
        {
            "RADISHMIND_MODEL_PROFILE": "anyrouter",
            "RADISHMIND_MODEL_PROFILE_FALLBACKS": "anyrouter,backup",
            "RADISHMIND_MODEL_PROFILE_ANYROUTER_NAME": "deepseek-chat",
            "RADISHMIND_MODEL_PROFILE_ANYROUTER_BASE_URL": "https://example.openai.invalid/v1",
            "RADISHMIND_MODEL_PROFILE_ANYROUTER_API_KEY": SECRET_SENTINEL,
            "RADISHMIND_MODEL_PROFILE_BACKUP_NAME": "gemini-1.5-flash",
            "RADISHMIND_MODEL_PROFILE_BACKUP_BASE_URL": "https://generativelanguage.googleapis.com/v1beta",
            "RADISHMIND_MODEL_PROFILE_BACKUP_API_KEY": SECRET_SENTINEL,
            "RADISHMIND_HUGGINGFACE_PROFILE": "hf-chat",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_CHAT_NAME": "meta-llama/Meta-Llama-3.1-8B-Instruct",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_CHAT_BASE_URL": "https://example.huggingface.invalid/v1",
            "RADISHMIND_HUGGINGFACE_PROFILE_HF_CHAT_API_KEY": SECRET_SENTINEL,
            "RADISHMIND_OLLAMA_PROFILE": "local",
            "RADISHMIND_OLLAMA_PROFILE_LOCAL_NAME": "qwen2.5:7b-instruct",
            "RADISHMIND_OLLAMA_PROFILE_LOCAL_BASE_URL": "http://localhost:11434",
            "RADISHMIND_OLLAMA_PROFILE_LOCAL_API_STYLE": "ollama-chat-completions",
        }
    )
    return env


def platform_temp_dir() -> Path:
    return Path(os.getenv("TMPDIR") or os.getenv("TEMP") or os.getenv("TMP") or "/tmp")


def run_platform_diagnostics(*, env: dict[str, str], cwd: Path = REPO_ROOT) -> subprocess.CompletedProcess[str]:
    if cwd == REPO_ROOT:
        args = platform_wrapper_args("diagnostics")
    else:
        args = ["go", "run", "./cmd/radishmind-platform", "diagnostics"]
    return subprocess.run(args, cwd=cwd, env=env, capture_output=True, text=True)


def platform_wrapper_args(command: str) -> list[str]:
    if platform.system().lower().startswith("win"):
        return ["pwsh", str(PLATFORM_WRAPPER_PS1), "-Command", command]
    return [str(PLATFORM_WRAPPER_SH), command]


def parse_json_output(result: subprocess.CompletedProcess[str]) -> dict[str, Any]:
    try:
        document = json.loads(result.stdout)
    except json.JSONDecodeError as exc:
        raise SystemExit(f"diagnostics output is not valid json: {exc}\nstdout={result.stdout}\nstderr={result.stderr}") from exc
    require(isinstance(document, dict), "diagnostics output must be an object")
    return document


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_no_secret_leak(document: dict[str, Any]) -> None:
    serialized = json.dumps(document, ensure_ascii=False)
    require(SECRET_SENTINEL not in serialized, "diagnostics leaked secret sentinel")
    config = document.get("config") or {}
    require("base_url" not in config, "diagnostics config must not expose base_url")
    require("api_key" not in config, "diagnostics config must not expose api_key")


def check_success_report() -> None:
    env = base_env()
    env.update(
        {
            "RADISHMIND_PLATFORM_PROVIDER": "mock",
            "RADISHMIND_PLATFORM_MODEL": "radishmind-local-dev",
        }
    )
    result = run_platform_diagnostics(env=env)
    require(result.returncode == 0, f"diagnostics should pass for mock config: {result.stderr or result.stdout}")
    document = parse_json_output(result)
    require(document.get("status") == "ok", "diagnostics status should be ok")
    require(document.get("sanitized") is True, "diagnostics should be marked sanitized")
    require_no_secret_leak(document)

    checks = {str(item.get("name")): item for item in document.get("checks") or [] if isinstance(item, dict)}
    for check_name in (
        "config_required_fields",
        "bridge_provider_registry",
        "bridge_provider_inventory",
        "deployment_readiness",
    ):
        require(checks.get(check_name, {}).get("status") == "ok", f"diagnostics check failed: {check_name}")

    bridge = document.get("bridge") or {}
    require(bridge.get("provider_registry_ok") is True, "diagnostics should mark provider registry ok")
    require(bridge.get("inventory_ok") is True, "diagnostics should mark inventory ok")
    providers = document.get("providers") or {}
    provider_ids = set(providers.get("registry_provider_ids") or [])
    require({"mock", "openai-compatible", "huggingface", "ollama"} <= provider_ids, "diagnostics provider registry is incomplete")
    require(int(providers.get("profile_count") or 0) >= 4, "diagnostics should expose configured profile count")
    selectable_model_ids = set(providers.get("selectable_model_ids") or [])
    require("profile:anyrouter" in selectable_model_ids, "diagnostics should expose openai-compatible selectable model id")
    require(
        "provider:huggingface:profile:hf-chat" in selectable_model_ids,
        "diagnostics should expose huggingface selectable model id",
    )
    require(
        "provider:ollama:profile:local" in selectable_model_ids,
        "diagnostics should expose ollama selectable model id",
    )
    require(
        int(providers.get("selectable_model_count") or 0) == len(selectable_model_ids),
        "diagnostics selectable model count should match selectable ids",
    )
    require(providers.get("missing_credential_model_ids") == [], "diagnostics should not report missing credentials in success path")
    require(int(providers.get("configured_credential_count") or 0) >= 3, "diagnostics should count configured credentials")
    require(int(providers.get("optional_credential_count") or 0) >= 1, "diagnostics should count optional credentials")


def check_failure_report() -> None:
    env = base_env()
    env.update(
        {
            "RADISHMIND_PLATFORM_PROVIDER": "huggingface",
            "RADISHMIND_PLATFORM_MODEL": "meta-llama/Meta-Llama-3.1-8B-Instruct",
        }
    )
    result = run_platform_diagnostics(env=env)
    require(result.returncode == 1, "diagnostics should fail when required platform fields are missing")
    document = parse_json_output(result)
    require_no_secret_leak(document)
    require(document.get("status") == "error", "failure diagnostics status should be error")
    require("CONFIG_REQUIRED_FIELDS_MISSING" in (document.get("failure_codes") or []), "missing config failure code")
    config = document.get("config") or {}
    require(
        config.get("missing_required_fields") == ["base_url", "credential"],
        f"unexpected missing fields: {config.get('missing_required_fields')}",
    )
    failure = document.get("failure") or {}
    require(failure.get("code") == "CONFIG_REQUIRED_FIELDS_MISSING", "unexpected primary failure code")


def check_bridge_path_resolution_from_platform_dir() -> None:
    env = base_env()
    env.update(
        {
            "RADISHMIND_PLATFORM_PROVIDER": "mock",
            "RADISHMIND_PLATFORM_MODEL": "radishmind-local-dev",
        }
    )
    result = run_platform_diagnostics(env=env, cwd=PLATFORM_DIR)
    require(result.returncode == 0, f"diagnostics should resolve bridge script from services/platform: {result.stderr or result.stdout}")
    document = parse_json_output(result)
    require(document.get("status") == "ok", "platform-dir diagnostics status should be ok")
    require_no_secret_leak(document)


def main() -> int:
    check_success_report()
    check_failure_report()
    check_bridge_path_resolution_from_platform_dir()
    print("platform diagnostics checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
