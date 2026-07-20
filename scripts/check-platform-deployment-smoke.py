#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import platform
import subprocess
import tempfile
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
PLATFORM_WRAPPER_SH = REPO_ROOT / "scripts/run-platform-service.sh"
PLATFORM_WRAPPER_PS1 = REPO_ROOT / "scripts/run-platform-service.ps1"
SECRET_SENTINEL = "deployment-smoke-secret"


def base_env() -> dict[str, str]:
    env = {key: value for key, value in os.environ.items() if not key.startswith("RADISHMIND_")}
    env.setdefault("GOCACHE", str(platform_temp_dir() / "radishmind-go-build-cache"))
    return env


def platform_temp_dir() -> Path:
    return Path(os.getenv("TMPDIR") or os.getenv("TEMP") or os.getenv("TMP") or "/tmp")


def run_platform_wrapper(
    args: list[str],
    *,
    env: dict[str, str],
    profile: str | None = None,
) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        platform_wrapper_args(args, profile=profile),
        cwd=REPO_ROOT,
        env=env,
        capture_output=True,
        text=True,
    )


def platform_wrapper_args(args: list[str], *, profile: str | None = None) -> list[str]:
    if platform.system().lower().startswith("win"):
        command = args[0] if args else "serve"
        remaining_args = args[1:] if len(args) > 1 else []
        profile_args = ["-Profile", profile] if profile else []
        return ["pwsh", str(PLATFORM_WRAPPER_PS1), *profile_args, "-Command", command, *remaining_args]
    profile_args = ["--profile", profile] if profile else []
    return [str(PLATFORM_WRAPPER_SH), *profile_args, *args]


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
        result = run_platform_wrapper(["config-check"], env=env)
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
        result = run_platform_wrapper(["config-summary"], env=env)
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
        result = run_platform_wrapper(["config-check"], env=env)
        require(result.returncode != 0, "invalid deployment config should fail")
        combined_output = f"{result.stdout}\n{result.stderr}"
        require("bridge_timeout" in combined_output, "invalid deployment config should name bridge_timeout")


def check_wrapper_rejects_unknown_command() -> None:
    result = run_platform_wrapper(["unknown-command"], env=base_env())
    require(result.returncode == 2, "platform wrapper should reject unknown commands with exit code 2")
    require("unsupported platform service command" in result.stderr, "wrapper failure should explain unsupported command")


def check_local_product_profile() -> None:
    with tempfile.TemporaryDirectory(prefix="radishmind-local-product-profile-") as temp_dir:
        database_path = Path(temp_dir) / "radishmind.db"
        env = base_env()
        env.update(
            {
                "RADISHMIND_PLATFORM_PROVIDER": "mock",
                "RADISHMIND_PLATFORM_MODEL": "radishmind-local-dev",
                "RADISHMIND_SQLITE_DEV_DATABASE_PATH": str(database_path),
            }
        )
        summary_result = run_platform_wrapper(["config-summary"], env=env)
        require(summary_result.returncode == 0, f"local-product config-summary should pass: {summary_result.stderr}")
        summary = parse_json_output(summary_result)
        require(summary.get("local_persistence_mode") == "sqlite_dev", "local-product must select sqlite_dev")
        require(summary.get("sqlite_dev_database_configured") is True, "local-product must configure the SQLite database")
        require(
            summary.get("local_persistence_components_consistent") is True,
            "local-product component stores must be consistent",
        )
        for field in (
            "workflow_saved_draft_store_mode",
            "application_draft_store_mode",
            "application_publish_store_mode",
            "application_catalog_store_mode",
            "api_key_store_mode",
            "workflow_run_store_mode",
            "gateway_request_store_mode",
        ):
            require(summary.get(field) == "sqlite_dev", f"local-product did not project {field}")
        for field in (
            "control_plane_read_dev_auth_enabled",
            "workflow_saved_draft_dev_http_enabled",
            "workflow_saved_draft_dev_write_enabled",
            "application_draft_dev_http_enabled",
            "application_draft_dev_write_enabled",
            "application_publish_dev_http_enabled",
            "application_publish_dev_write_enabled",
            "application_catalog_dev_http_enabled",
            "application_catalog_dev_write_enabled",
            "api_key_lifecycle_dev_http_enabled",
            "api_key_lifecycle_dev_write_enabled",
            "gateway_request_history_dev_enabled",
            "workflow_executor_dev_enabled",
        ):
            require(summary.get(field) is True, f"local-product did not enable {field}")
        require(str(database_path) not in json.dumps(summary), "local-product config-summary leaked the SQLite path")
        require(not database_path.exists(), "config-summary must not create the SQLite database")

        check_result = run_platform_wrapper(["config-check"], env=env)
        require(check_result.returncode == 0, f"local-product config-check should pass: {check_result.stderr}")
        require(not database_path.exists(), "config-check must not create the SQLite database")

        configured_result = run_platform_wrapper(["config-summary"], env=env, profile="configured")
        require(configured_result.returncode == 0, f"configured config-summary should pass: {configured_result.stderr}")
        configured_summary = parse_json_output(configured_result)
        require(configured_summary.get("local_persistence_mode") == "memory_dev", "configured profile must preserve default memory_dev")
        require(configured_summary.get("local_persistence_configured") is False, "configured profile must not inject aggregate persistence")
        require(not database_path.exists(), "configured config-summary must not create the SQLite database")


def check_local_product_profile_rejects_conflicts() -> None:
    env = base_env()
    env.update(
        {
            "RADISHMIND_PLATFORM_PROVIDER": "mock",
            "RADISHMIND_PLATFORM_MODEL": "radishmind-local-dev",
            "RADISHMIND_APPLICATION_CATALOG_STORE": "memory_dev",
        }
    )
    component_conflict = run_platform_wrapper(["config-check"], env=env)
    require(component_conflict.returncode != 0, "local-product must reject an explicit component store")
    require("conflicts with an explicit component store mode" in component_conflict.stderr, "component conflict should be explicit")

    env = base_env()
    env["RADISHMIND_LOCAL_PERSISTENCE_MODE"] = "memory_dev"
    aggregate_conflict = run_platform_wrapper(["config-summary"], env=env)
    require(aggregate_conflict.returncode == 2, "local-product must reject a non-SQLite aggregate mode")
    require("use --profile configured" in aggregate_conflict.stderr or "use -Profile configured" in aggregate_conflict.stderr, "aggregate conflict should point to configured profile")

    unknown_profile = run_platform_wrapper(["config-summary"], env=base_env(), profile="unknown")
    require(unknown_profile.returncode == 2, "unknown startup profile must return exit code 2")
    require("unsupported platform startup profile" in unknown_profile.stderr, "unknown profile failure should be explicit")


def check_wrapper_sets_python_bridge_default() -> None:
    sh_content = PLATFORM_WRAPPER_SH.read_text(encoding="utf-8")
    ps1_content = PLATFORM_WRAPPER_PS1.read_text(encoding="utf-8")
    for content, name in ((sh_content, "shell"), (ps1_content, "PowerShell")):
        require(
            "RADISHMIND_PLATFORM_PYTHON_BIN" in content,
            f"{name} wrapper must set RADISHMIND_PLATFORM_PYTHON_BIN default",
        )
        require(".venv" in content, f"{name} wrapper must prefer repository .venv Python")
        require("local-product" in content and "configured" in content, f"{name} wrapper must expose both startup profiles")
        require("RADISHMIND_LOCAL_PERSISTENCE_MODE" in content, f"{name} wrapper must set aggregate local persistence")
        require("RADISHMIND_SQLITE_DEV_DATABASE_PATH" in content, f"{name} wrapper must set the controlled SQLite path")


def main() -> int:
    check_mock_deployment_config()
    check_env_override()
    check_invalid_deployment_config()
    check_wrapper_rejects_unknown_command()
    check_local_product_profile()
    check_local_product_profile_rejects_conflicts()
    check_wrapper_sets_python_bridge_default()
    print("platform deployment smoke checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
