#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
README_PATH = REPO_ROOT / "services/platform/README.md"
CONFIG_PATH = REPO_ROOT / "services/platform/internal/config/config.go"
SERVER_PATH = REPO_ROOT / "services/platform/internal/httpapi/server.go"
MAIN_PATH = REPO_ROOT / "services/platform/cmd/radishmind-platform/main.go"


EXPECTED_COMMAND_PATTERNS = (
    "cd services/platform",
    "GOCACHE=/tmp/radishmind-go-build-cache go test ./...",
    "RADISHMIND_PLATFORM_CONFIG=tmp/radishmind-platform.local.json",
    "./scripts/run-platform-service.sh config-check",
    "./scripts/run-platform-service.sh diagnostics",
    "./scripts/run-platform-service.sh serve",
    "pwsh ./scripts/run-platform-service.ps1 -Command config-check",
    "pwsh ./scripts/run-platform-service.ps1 -Command diagnostics",
    "go run ./cmd/radishmind-platform",
    "go run ./services/platform/cmd/radishmind-platform config-summary",
    "go run ./services/platform/cmd/radishmind-platform config-check",
    "go run ./services/platform/cmd/radishmind-platform diagnostics",
    "curl -sS http://127.0.0.1:8080/healthz",
    "curl -sS http://127.0.0.1:8080/v1/models",
    "curl -sS http://127.0.0.1:8080/v1/models/mock",
    "curl -sS http://127.0.0.1:8080/v1/chat/completions",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def extract_platform_env_keys(config_content: str) -> set[str]:
    return set(re.findall(r'"(RADISHMIND_PLATFORM_[A-Z0-9_]+)"', config_content))


def extract_http_routes(server_content: str) -> set[str]:
    return set(re.findall(r'mux\.HandleFunc\("([^"]+)"', server_content))


def main() -> int:
    readme = README_PATH.read_text(encoding="utf-8")
    config_content = CONFIG_PATH.read_text(encoding="utf-8")
    server_content = SERVER_PATH.read_text(encoding="utf-8")
    main_content = MAIN_PATH.read_text(encoding="utf-8")

    require("## 本地启动 runbook" in readme, "platform README must include local startup runbook")
    require("## 环境变量" in readme, "platform README must include environment variable section")
    require("## 本地 smoke 验证" in readme, "platform README must include local smoke section")
    require("## 故障边界" in readme, "platform README must include failure boundary section")
    require("default < config file < env" in readme, "platform README must document config precedence")
    require("scripts/check-platform-deployment-smoke.py" in readme, "platform README must mention deployment smoke")
    require("scripts/check-platform-diagnostics.py" in readme, "platform README must mention diagnostics smoke")
    require("scripts/run-platform-service.sh" in readme, "platform README must mention platform service shell wrapper")
    require("scripts/run-platform-service.ps1" in readme, "platform README must mention platform service PowerShell wrapper")

    env_keys = extract_platform_env_keys(config_content)
    missing_env_keys = sorted(key for key in env_keys if key not in readme)
    require(not missing_env_keys, f"platform README is missing env keys from config.go: {missing_env_keys}")

    routes = extract_http_routes(server_content)
    missing_routes = sorted(route for route in routes if route.split(" ", 1)[1] not in readme)
    require(not missing_routes, f"platform README is missing HTTP routes from server.go: {missing_routes}")

    for pattern in EXPECTED_COMMAND_PATTERNS:
        require(pattern in readme, f"platform README is missing runbook command: {pattern}")

    for required_source in ("config.LoadFromEnv", "httpapi.NewServer", "ListenAndServe"):
        require(required_source in main_content, f"platform main must keep startup source: {required_source}")
        require(required_source in readme, f"platform README must mention startup source: {required_source}")

    print("platform runbook drift check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
