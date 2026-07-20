#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
README_PATH = REPO_ROOT / "services/platform/README.md"
RUNBOOK_PATH = REPO_ROOT / "docs/platform/platform-service-operations-runbook-v1.md"
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
    "curl -sS http://127.0.0.1:7000/healthz",
    "curl -sS http://127.0.0.1:7000/v1/platform/local-smoke",
    "curl -sS http://127.0.0.1:7000/v1/models",
    "curl -sS http://127.0.0.1:7000/v1/models/mock",
    "curl -sS http://127.0.0.1:7000/v1/chat/completions",
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
    runbook = RUNBOOK_PATH.read_text(encoding="utf-8")
    config_content = CONFIG_PATH.read_text(encoding="utf-8")
    server_content = SERVER_PATH.read_text(encoding="utf-8")
    main_content = MAIN_PATH.read_text(encoding="utf-8")

    for section in ("## 服务职责", "## 路由分类", "## 启动入口", "## 权威专题", "## 停止线"):
        require(section in readme, f"platform README must include entry section: {section}")
    require(
        "docs/platform/platform-service-operations-runbook-v1.md" in readme,
        "platform README must link the authoritative operations runbook",
    )
    for section in ("## 本地启动 runbook", "## 环境变量", "## 本地 smoke 验证", "## 故障边界"):
        require(section in runbook, f"platform operations runbook must include section: {section}")
    require("default < config file < env" in runbook, "platform runbook must document config precedence")
    require("scripts/check-platform-deployment-smoke.py" in runbook, "platform runbook must mention deployment smoke")
    require("scripts/check-platform-diagnostics.py" in runbook, "platform runbook must mention diagnostics smoke")
    require("scripts/run-platform-service.sh" in runbook, "platform runbook must mention platform service shell wrapper")
    require("scripts/run-platform-service.ps1" in runbook, "platform runbook must mention platform service PowerShell wrapper")
    require(
        ".venv" in runbook and "RADISHMIND_PLATFORM_PYTHON_BIN" in runbook,
        "platform runbook must document wrapper Python bridge default",
    )

    env_keys = extract_platform_env_keys(config_content)
    missing_env_keys = sorted(key for key in env_keys if key not in runbook)
    require(not missing_env_keys, f"platform runbook is missing env keys from config.go: {missing_env_keys}")

    routes = extract_http_routes(server_content)
    missing_routes = sorted(route for route in routes if route.split(" ", 1)[1] not in runbook)
    require(not missing_routes, f"platform runbook is missing HTTP routes from server.go: {missing_routes}")

    for pattern in EXPECTED_COMMAND_PATTERNS:
        require(pattern in runbook, f"platform runbook is missing command: {pattern}")

    for required_source in ("config.LoadFromEnv", "httpapi.NewServer", "ListenAndServe"):
        require(required_source in main_content, f"platform main must keep startup source: {required_source}")
        require(required_source in runbook, f"platform runbook must mention startup source: {required_source}")

    print("platform runbook drift check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
