#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
TS_CONTRACT = REPO_ROOT / "contracts/typescript/platform-overview-api.ts"
GO_ROUTE_SOURCES = (
    REPO_ROOT / "services/platform/internal/httpapi/platform_overview.go",
    REPO_ROOT / "services/platform/internal/httpapi/session_tooling_metadata.go",
)
CONSUMER_SMOKE = REPO_ROOT / "scripts/run-platform-overview-consumer-smoke.py"

EXPECTED_LITERALS = (
    "/v1/platform/overview",
    "/v1/models",
    "/v1/models/{id}",
    "/v1/session/metadata",
    "/v1/tools/metadata",
    "/v1/tools/actions",
    "platform_overview",
    "P3 Local Product Shell",
    "local_read_only_product_shell",
    "bridge_backed_provider_profile_inventory",
    "P2 close candidate shell",
    "PlatformOverviewResponse",
    "PlatformOverviewConsoleViewModel",
    "toPlatformOverviewConsoleViewModel",
    "allPlatformOverviewStopLinesEnforced",
    "canExecuteActions: false",
    "canUseDurableStore: false",
    "canWriteBusinessTruth: false",
    "canReplayAutomatically: false",
    "execution_enabled: false",
    "metadata_only: true",
)

FORBIDDEN_LITERALS = (
    "canExecuteActions: true",
    "canUseDurableStore: true",
    "canWriteBusinessTruth: true",
    "canReplayAutomatically: true",
    "execution_enabled: true",
    "metadata_only: false",
    "writes_business_truth: true",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def main() -> int:
    ts_content = TS_CONTRACT.read_text(encoding="utf-8")
    go_content = "\n".join(path.read_text(encoding="utf-8") for path in GO_ROUTE_SOURCES)

    for literal in EXPECTED_LITERALS:
        require(literal in ts_content, f"typescript overview consumer contract missing literal: {literal}")

    for literal in FORBIDDEN_LITERALS:
        require(literal not in ts_content, f"typescript overview consumer contract enables forbidden state: {literal}")

    for route in (
        "/v1/platform/overview",
        "/v1/models",
        "/v1/session/metadata",
        "/v1/tools/metadata",
        "/v1/tools/actions",
    ):
        require(route in go_content, f"go platform overview route source missing route: {route}")
        require(route in ts_content, f"typescript overview consumer contract missing route: {route}")

    require(
        "real_executor_enabled: false" in ts_content
        and "durable_store_enabled: false" in ts_content
        and "automatic_replay_enabled: false" in ts_content,
        "typescript overview contract must keep executor, durable store and replay stop-lines disabled",
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

    print("platform overview consumer contract check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
