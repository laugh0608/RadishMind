#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
CONSOLE_ROOT = REPO_ROOT / "apps/radishmind-console"
PACKAGE_JSON = CONSOLE_ROOT / "package.json"
APP_SOURCE = CONSOLE_ROOT / "src/App.tsx"
CLIENT_SOURCE = CONSOLE_ROOT / "src/platformOverviewClient.ts"
README = CONSOLE_ROOT / "README.md"
GO_SERVER_SOURCE = REPO_ROOT / "services/platform/internal/httpapi/server.go"

EXPECTED_FILES = (
    "package.json",
    "index.html",
    "tsconfig.json",
    "vite.config.ts",
    "src/main.tsx",
    "src/App.tsx",
    "src/platformOverviewClient.ts",
    "src/styles.css",
    "README.md",
)

EXPECTED_SOURCE_LITERALS = (
    "PLATFORM_OVERVIEW_ROUTE",
    "PLATFORM_LOCAL_SMOKE_ROUTE",
    "toPlatformOverviewConsoleViewModel",
    "toPlatformLocalSmokeReadinessViewModel",
    "loadPlatformOverview",
    "http://127.0.0.1:7000",
    "Service Status",
    "Model Inventory",
    "Session And Tooling",
    "Local Readiness",
    "Stop Lines",
    "Active profile chain",
    "Local-smoke endpoint",
    "Platform service unavailable",
    "Refreshing; showing last overview",
    "getPlatformOverviewDiagnostics",
    "real_executor_enabled",
    "durable_store_enabled",
    "confirmation_flow_connected",
    "business_truth_write_enabled",
    "automatic_replay_enabled",
)

FORBIDDEN_SOURCE_LITERALS = (
    "/v1/tools/actions\", {",
    "method: \"POST\"",
    "canExecuteActions: true",
    "canUseDurableStore: true",
    "canWriteBusinessTruth: true",
    "canReplayAutomatically: true",
    "executionEnabled: true",
    "writes_business_truth: true",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def main() -> int:
    for relpath in EXPECTED_FILES:
        require((CONSOLE_ROOT / relpath).is_file(), f"local console missing file: {relpath}")

    package = json.loads(PACKAGE_JSON.read_text(encoding="utf-8"))
    require(package.get("private") is True, "RadishMind console package must stay private")
    require(package.get("name") == "@radishmind/console", "RadishMind console package name drifted")
    require(package.get("type") == "module", "RadishMind console package must use ESM")
    scripts = package.get("scripts") or {}
    require("vite" in scripts.get("dev", ""), "local console dev script must use Vite")
    require("tsc --noEmit" in scripts.get("build", ""), "local console build must typecheck")

    dependencies = package.get("dependencies") or {}
    dev_dependencies = package.get("devDependencies") or {}
    for dependency in ("react", "react-dom"):
        require(dependency in dependencies, f"local console missing dependency: {dependency}")
    for dependency in ("vite", "typescript", "@vitejs/plugin-react"):
        require(dependency in dev_dependencies, f"local console missing dev dependency: {dependency}")

    source = "\n".join(
        [
            APP_SOURCE.read_text(encoding="utf-8"),
            CLIENT_SOURCE.read_text(encoding="utf-8"),
            README.read_text(encoding="utf-8"),
        ]
    )
    for literal in EXPECTED_SOURCE_LITERALS:
        require(literal in source, f"local console missing literal: {literal}")
    for literal in FORBIDDEN_SOURCE_LITERALS:
        require(literal not in source, f"local console enables forbidden state: {literal}")

    require(
        "../../../contracts/typescript/platform-overview-api.ts" in CLIENT_SOURCE.read_text(encoding="utf-8"),
        "local console must consume the shared overview TypeScript contract",
    )
    require(
        "../../../contracts/typescript/platform-local-smoke-api.ts" in CLIENT_SOURCE.read_text(encoding="utf-8"),
        "local console must consume the shared local-smoke TypeScript contract",
    )
    require(
        "fetch(endpoint" in CLIENT_SOURCE.read_text(encoding="utf-8")
        and "Accept: \"application/json\"" in CLIENT_SOURCE.read_text(encoding="utf-8"),
        "local console must fetch platform JSON endpoints directly",
    )
    go_server = GO_SERVER_SOURCE.read_text(encoding="utf-8")
    for origin in ("http://127.0.0.1:4000", "http://localhost:4000"):
        require(origin in go_server, f"platform server missing local console CORS origin: {origin}")
    require(
        "http.MethodOptions" in go_server
        and "Access-Control-Allow-Origin" in go_server
        and "Access-Control-Allow-Methods" in go_server,
        "platform server must keep local console CORS/preflight support",
    )

    print("platform local console shell check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
