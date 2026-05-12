#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import subprocess
import tempfile
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
PLATFORM_DIR = REPO_ROOT / "services/platform"
SECRET_SENTINEL = "deployment-smoke-secret"


def base_env() -> dict[str, str]:
    env = {key: value for key, value in os.environ.items() if not key.startswith("RADISHMIND_")}
    env.setdefault("GOCACHE", str(Path(os.getenv("TMPDIR", "/tmp")) / "radishmind-go-build-cache"))
    return env


def run_platform_command(args: list[str], *, env: dict[str, str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["go", "run", "./cmd/radishmind-platform", *args],
        cwd=PLATFORM_DIR,
        env=env,
        capture_output=True,
        text=True,
    )


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def parse_json_output(result: subprocess.CompletedProcess[str]) -> dict[str, Any]:
    document = json.loads(result.stdout)
    require(isinstance(document, dict), "platform command output must be an object")
    return document


def write_config(path: Path, document: dict[str, Any]) -> None:
    path.write_text(json.dumps(document, indent=2), encoding="utf-8")


def check_mock_deployment_config() -> None:
    with tempfile.TemporaryDirectory(prefix="radishmind-platform-deploy-") as temp_dir:
        config_path = Path(temp_dir) / "mock-platform.json"
        write_config(
            config_path,
            {
                "listen_addr": "127.0.0.1:8282",
                "provider": "mock",
                "model": "radishmind-local-dev",
                "bridge_timeout": "20s",
            },
        )
        env = base_env()
        env["RADISHMIND_PLATFORM_CONFIG"] = str(config_path)
        result = run_platform_command(["config-check"], env=env)
        require(result.returncode == 0, f"mock deployment config should pass: {result.stderr or result.stdout}")
        document = parse_json_output(result)
        config = document.get("config") or {}
        require(document.get("status") == "ok", "mock deployment config-check status should be ok")
        require(config.get("provider") == "mock", "mock deployment provider did not load")
        require(config.get("model") == "radishmind-local-dev", "mock deployment model did not load")
        require(config.get("field_sources", {}).get("listen_addr") == "file", "listen addr should come from file")


def check_env_override() -> None:
    with tempfile.TemporaryDirectory(prefix="radishmind-platform-deploy-") as temp_dir:
        config_path = Path(temp_dir) / "remote-platform.json"
        write_config(
            config_path,
            {
                "listen_addr": "127.0.0.1:8283",
                "provider": "huggingface",
                "provider_profile": "file-profile",
                "model": "file-model",
                "base_url": "https://file.example.invalid/v1",
                "api_key": SECRET_SENTINEL,
                "bridge_timeout": "40s",
            },
        )
        env = base_env()
        env.update(
            {
                "RADISHMIND_PLATFORM_CONFIG": str(config_path),
                "RADISHMIND_PLATFORM_PROVIDER_PROFILE": "env-profile",
                "RADISHMIND_PLATFORM_MODEL": "env-model",
            }
        )
        result = run_platform_command(["config-summary"], env=env)
        require(result.returncode == 0, f"env override config-summary should pass: {result.stderr or result.stdout}")
        document = parse_json_output(result)
        require(document.get("profile") == "env-profile", "env profile should override config file")
        require(document.get("model") == "env-model", "env model should override config file")
        require(document.get("field_sources", {}).get("profile") == "env", "profile source should be env")
        require(document.get("field_sources", {}).get("model") == "env", "model source should be env")
        require(SECRET_SENTINEL not in json.dumps(document, ensure_ascii=False), "deployment smoke leaked secret")


def check_invalid_deployment_config() -> None:
    with tempfile.TemporaryDirectory(prefix="radishmind-platform-deploy-") as temp_dir:
        config_path = Path(temp_dir) / "invalid-platform.json"
        write_config(config_path, {"provider": "openai-compatible", "bridge_timeout": "not-a-duration"})
        env = base_env()
        env["RADISHMIND_PLATFORM_CONFIG"] = str(config_path)
        result = run_platform_command(["config-check"], env=env)
        require(result.returncode != 0, "invalid deployment config should fail")
        combined_output = f"{result.stdout}\n{result.stderr}"
        require("bridge_timeout" in combined_output, "invalid deployment config should name bridge_timeout")


def main() -> int:
    check_mock_deployment_config()
    check_env_override()
    check_invalid_deployment_config()
    print("platform deployment smoke checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
