#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
PLATFORM_DIR = REPO_ROOT / "services/platform"
SECRET_SENTINEL = "radishmind-secret-sentinel"
REQUIRED_SUMMARY_FIELDS = {
    "listen_addr",
    "provider",
    "profile",
    "model",
    "model_configured",
    "base_url_configured",
    "credential_state",
    "timeouts",
    "python_bridge",
    "temperature",
    "required_fields",
    "missing_required_fields",
    "secret_fields",
    "sanitized",
}


def run_platform_command(args: list[str], *, env: dict[str, str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["go", "run", "./cmd/radishmind-platform", *args],
        cwd=PLATFORM_DIR,
        env=env,
        capture_output=True,
        text=True,
    )


def load_summary(env: dict[str, str]) -> dict[str, Any]:
    result = run_platform_command(["config-summary"], env=env)
    require(result.returncode == 0, f"config-summary failed: {result.stderr or result.stdout}")
    document = json.loads(result.stdout)
    require(isinstance(document, dict), "config-summary output must be an object")
    return document


def load_check(env: dict[str, str]) -> tuple[int, dict[str, Any]]:
    result = run_platform_command(["config-check"], env=env)
    document = json.loads(result.stdout)
    require(isinstance(document, dict), "config-check output must be an object")
    return result.returncode, document


def base_env() -> dict[str, str]:
    env = {key: value for key, value in os.environ.items() if not key.startswith("RADISHMIND_")}
    env.setdefault("GOCACHE", str(Path(os.getenv("TMPDIR", "/tmp")) / "radishmind-go-build-cache"))
    return env


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_summary_shape(summary: dict[str, Any]) -> None:
    missing = sorted(REQUIRED_SUMMARY_FIELDS - set(summary))
    require(not missing, f"config summary missing fields: {missing}")
    require(summary.get("sanitized") is True, "config summary must be marked sanitized")
    require(isinstance(summary.get("timeouts"), dict), "config summary timeouts must be an object")
    require(isinstance(summary.get("python_bridge"), dict), "config summary python_bridge must be an object")
    require("api_key" not in summary, "config summary must not expose api_key")
    require("base_url" not in summary, "config summary must not expose base_url")
    serialized = json.dumps(summary, ensure_ascii=False)
    require(SECRET_SENTINEL not in serialized, "config summary leaked secret sentinel")


def check_mock_default() -> None:
    env = base_env()
    summary = load_summary(env)
    require_summary_shape(summary)
    require(summary["provider"] == "mock", "default provider should be mock")
    require(summary["credential_state"] == "not_required", "mock credential state should be not_required")
    require(summary["missing_required_fields"] == [], "mock config should not miss required fields")

    status_code, check_document = load_check(env)
    require(status_code == 0, "mock config-check should pass")
    require(check_document.get("status") == "ok", "mock config-check status should be ok")


def check_remote_provider_success() -> None:
    env = base_env()
    env.update(
        {
            "RADISHMIND_PLATFORM_LISTEN_ADDR": "127.0.0.1:8181",
            "RADISHMIND_PLATFORM_PROVIDER": "huggingface",
            "RADISHMIND_PLATFORM_PROVIDER_PROFILE": "hf-chat",
            "RADISHMIND_PLATFORM_MODEL": "meta-llama/Meta-Llama-3.1-8B-Instruct",
            "RADISHMIND_PLATFORM_BASE_URL": "https://example.huggingface.invalid/v1",
            "RADISHMIND_PLATFORM_API_KEY": SECRET_SENTINEL,
            "RADISHMIND_PLATFORM_BRIDGE_TIMEOUT": "45s",
        }
    )
    summary = load_summary(env)
    require_summary_shape(summary)
    require(summary["listen_addr"] == "127.0.0.1:8181", "listen addr did not round-trip")
    require(summary["provider"] == "huggingface", "provider did not round-trip")
    require(summary["profile"] == "hf-chat", "profile did not round-trip")
    require(summary["model"] == "meta-llama/Meta-Llama-3.1-8B-Instruct", "model did not round-trip")
    require(summary["model_configured"] is True, "model_configured should be true")
    require(summary["base_url_configured"] is True, "base_url_configured should be true")
    require(summary["credential_state"] == "configured", "remote provider credential state should be configured")
    require(summary["timeouts"].get("bridge") == "45s", "bridge timeout did not round-trip")

    status_code, check_document = load_check(env)
    require(status_code == 0, "configured remote config-check should pass")
    require(check_document.get("status") == "ok", "configured remote config-check status should be ok")
    require(SECRET_SENTINEL not in json.dumps(check_document, ensure_ascii=False), "config-check leaked secret sentinel")


def check_remote_provider_failure() -> None:
    env = base_env()
    env.update(
        {
            "RADISHMIND_PLATFORM_PROVIDER": "openai-compatible",
            "RADISHMIND_PLATFORM_MODEL": "deepseek-chat",
        }
    )
    status_code, check_document = load_check(env)
    require(status_code == 1, "incomplete remote config-check should fail")
    require(check_document.get("status") == "error", "incomplete remote config-check status should be error")
    require(
        check_document.get("missing_required_fields") == ["base_url", "credential"],
        f"unexpected missing required fields: {check_document.get('missing_required_fields')}",
    )


def check_ollama_optional_credential() -> None:
    env = base_env()
    env.update(
        {
            "RADISHMIND_PLATFORM_PROVIDER": "ollama",
            "RADISHMIND_PLATFORM_MODEL": "qwen2.5:7b-instruct",
            "RADISHMIND_PLATFORM_BASE_URL": "http://localhost:11434",
        }
    )
    summary = load_summary(env)
    require_summary_shape(summary)
    require(summary["credential_state"] == "optional_missing", "ollama credential should be optional_missing")
    require(summary["missing_required_fields"] == [], "ollama config should not require credential")


def main() -> int:
    check_mock_default()
    check_remote_provider_success()
    check_remote_provider_failure()
    check_ollama_optional_credential()
    print("platform config checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
