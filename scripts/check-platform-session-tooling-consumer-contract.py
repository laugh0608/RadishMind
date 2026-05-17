#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
TS_CONTRACT = REPO_ROOT / "contracts/typescript/session-tooling-api.ts"
GO_ROUTE_SOURCE = REPO_ROOT / "services/platform/internal/httpapi/session_tooling_metadata.go"

EXPECTED_LITERALS = (
    "/v1/session/metadata",
    "/v1/tools/metadata",
    "/v1/tools/actions",
    "tool_action_blocked_response",
    "TOOL_EXECUTOR_DISABLED",
    "TOOL_NOT_REGISTERED",
    "CONFIRMATION_REQUIRED",
    "execution_enabled: false",
    "executed: false",
    "result_ref: null",
    "durable_memory_written: false",
    "writes_business_truth: false",
    "automatic_replay_started: false",
    "canExecute: false",
)

FORBIDDEN_LITERALS = (
    "canExecute: true",
    "executed: true",
    "execution_enabled: true",
    "durable_memory_written: true",
    "writes_business_truth: true",
    "automatic_replay_started: true",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def main() -> int:
    ts_content = TS_CONTRACT.read_text(encoding="utf-8")
    go_content = GO_ROUTE_SOURCE.read_text(encoding="utf-8")

    for literal in EXPECTED_LITERALS:
        require(literal in ts_content, f"typescript consumer contract missing literal: {literal}")

    for literal in FORBIDDEN_LITERALS:
        require(literal not in ts_content, f"typescript consumer contract enables forbidden state: {literal}")

    for route in (
        "/v1/session/metadata",
        "/v1/tools/metadata",
        "/v1/tools/actions",
    ):
        require(route in go_content, f"go session/tooling route source missing route: {route}")
        require(route in ts_content, f"typescript consumer contract missing route: {route}")

    require(
        "toToolActionBlockedView" in ts_content and "noSideEffects" in ts_content,
        "typescript consumer contract must expose a blocked action view with side-effect summary",
    )
    require(
        "listToolActionOptions" in ts_content,
        "typescript consumer contract must expose tool action option mapping",
    )

    print("platform session/tooling consumer contract check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
