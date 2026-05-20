#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
TS_CONTRACT = REPO_ROOT / "contracts/typescript/platform-local-smoke-api.ts"
GO_ROUTE_SOURCES = (
    REPO_ROOT / "services/platform/internal/httpapi/platform_local_smoke.go",
    REPO_ROOT / "services/platform/internal/httpapi/platform_overview.go",
    REPO_ROOT / "services/platform/internal/httpapi/server.go",
)
CONSUMER_SMOKE = REPO_ROOT / "scripts/run-platform-local-smoke.py"

EXPECTED_LITERALS = (
    "/v1/platform/local-smoke",
    "/healthz",
    "/v1/platform/overview",
    "/v1/models",
    "/v1/session/metadata",
    "/v1/tools/metadata",
    "/v1/tools/actions",
    "platform_local_smoke",
    "platform_local_smoke_readiness_view",
    "P3 Local Product Shell / Ops Surface",
    "local_console_ready",
    "read_only",
    "PORT_IN_USE",
    "CORS_ORIGIN_NOT_ALLOWED",
    "ERR_UNSAFE_PORT",
    "TOOL_EXECUTOR_DISABLED",
    "http://127.0.0.1:4000",
    "http://localhost:4000",
    "PlatformLocalSmokeResponse",
    "PlatformLocalSmokeReadinessViewModel",
    "toPlatformLocalSmokeReadinessViewModel",
    "allPlatformLocalSmokeStopLinesEnforced",
    "canExecuteActions: false",
    "canUseDurableStore: false",
    "canWriteBusinessTruth: false",
    "canReplayAutomatically: false",
    "execution_enabled: false",
    "metadata_only: true",
    "writes_business_truth: false",
)

FORBIDDEN_LITERALS = (
    "canExecuteActions: true",
    "canUseDurableStore: true",
    "canWriteBusinessTruth: true",
    "canReplayAutomatically: true",
    "execution_enabled: true",
    "metadata_only: false",
    "writes_business_truth: true",
    "automatic_replay_enabled: true",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def main() -> int:
    ts_content = TS_CONTRACT.read_text(encoding="utf-8")
    go_content = "\n".join(path.read_text(encoding="utf-8") for path in GO_ROUTE_SOURCES)
    smoke_content = CONSUMER_SMOKE.read_text(encoding="utf-8")

    for literal in EXPECTED_LITERALS:
        require(literal in ts_content or literal in smoke_content, f"local smoke contract missing literal: {literal}")

    for literal in FORBIDDEN_LITERALS:
        require(literal not in ts_content, f"typescript local smoke contract enables forbidden state: {literal}")

    for route in (
        "/v1/platform/local-smoke",
        "/healthz",
        "/v1/platform/overview",
        "/v1/models",
        "/v1/session/metadata",
        "/v1/tools/metadata",
        "/v1/tools/actions",
    ):
        require(route in go_content, f"go local smoke route source missing route: {route}")
        require(route in ts_content, f"typescript local smoke contract missing route: {route}")

    require(
        "localConsoleAllowedOrigins" in go_content
        and "http://127.0.0.1:4000" in go_content
        and "http://localhost:4000" in go_content,
        "go local smoke contract must expose the same local console CORS origins",
    )
    require(
        "real_executor_enabled: false" in ts_content
        and "durable_store_enabled: false" in ts_content
        and "automatic_replay_enabled: false" in ts_content,
        "typescript local smoke contract must keep executor, durable store and replay stop-lines disabled",
    )

    smoke_result = subprocess.run(
        [sys.executable, str(CONSUMER_SMOKE), "--check"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
    )
    if smoke_result.returncode != 0:
        if smoke_result.stdout:
            print(smoke_result.stdout, end="")
        if smoke_result.stderr:
            print(smoke_result.stderr, end="", file=sys.stderr)
        raise SystemExit(smoke_result.returncode)

    print("platform local smoke contract check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
