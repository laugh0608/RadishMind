#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
CONSOLE_ROOT = REPO_ROOT / "apps/radishmind-console"
APP_SOURCE = CONSOLE_ROOT / "src/App.tsx"
CLIENT_SOURCE = CONSOLE_ROOT / "src/platformOverviewClient.ts"
README = CONSOLE_ROOT / "README.md"
OVERVIEW_CONTRACT = REPO_ROOT / "contracts/typescript/platform-overview-api.ts"
LOCAL_SMOKE_CONTRACT = REPO_ROOT / "contracts/typescript/platform-local-smoke-api.ts"
OVERVIEW_ROUTE = REPO_ROOT / "services/platform/internal/httpapi/platform_overview.go"

READY_PATH_LITERALS = (
    "readyState?.viewModel",
    "readyState?.readinessViewModel",
    "readyState?.overview",
    "readyState?.localSmoke",
    "Loaded {formatTimestamp(readyState.loadedAt)}",
    "Overview route",
    "Local Readiness",
    "Stop-line Details",
    "Provider/Profile Details",
    "Inventory kind",
    "Models route",
    "Detail route",
    "Selector boundary",
    "display only; no health check or credential readiness",
    "buildModelInventoryDetails",
    "parseSelectableModelId",
    "formatModelInventoryKind",
    "Overview enforcement",
    "Local-smoke enforcement",
    "Blocked action",
    "Audit mode",
    "Active profile chain",
    "Confirmation path",
    "Audit Boundary",
    "advisory only",
)

REFRESH_PATH_LITERALS = (
    "lastReadyState",
    "previous: previousReadyState",
    "Refreshing; showing last overview",
    "Connection failed; showing last overview",
    "latestReadyState",
)

ERROR_DIAGNOSTIC_LITERALS = (
    "PlatformOverviewRequestError",
    "getPlatformOverviewFailureSurface",
    "failureSurface",
    "getPlatformOverviewDiagnostics",
    "could not reach platform overview",
    "overview request failed with HTTP",
    "overview response does not match platform overview contract",
    "Overview was readable; local-smoke readiness failed",
    "Overview was readable; local-smoke contract validation failed",
    "Start the platform service",
    "Confirm local CORS allows",
    "run-platform-overview-consumer-smoke.py --base-url",
    "diagnostic-list",
    "Dev Diagnostics",
    "diagnostics-band",
    "Platform URL",
    "Overview endpoint",
    "Failure surface",
    "Load status",
    "Last loaded",
    "Console connection",
    "Local probes",
    "Failure classes",
    "pwsh ./scripts/run-radishmind-console-dev.ps1 -VerifyOnly",
    "./scripts/run-radishmind-console-dev.sh --verify-only",
    "Port conflict: keep backend on 7000 and frontend on 4000",
    "CORS / preflight: platform only allows http://127.0.0.1:4000",
    "Unsafe port: ERR_UNSAFE_PORT",
    "Contract mismatch: run the overview consumer smoke",
    "Local-smoke mismatch: run the local-smoke readiness check",
    "Local-smoke endpoint",
    "Readiness status",
    "Local-smoke readiness unavailable",
    "local-smoke readiness",
    "Platform service unavailable",
)

READ_ONLY_BOUNDARY_LITERALS = (
    "method: \"GET\"",
    "Accept: \"application/json\"",
    "executionEnabled ? \"enabled\" : \"disabled\"",
    "writes_business_truth ? \"write enabled\" : \"advisory only\"",
    "P3 Local Product Shell / Ops Surface",
    "local_read_only_product_shell",
    "blockedActionNoSideEffects",
    "STOP_LINE_DETAILS",
    "readinessViewModel?.allStopLinesEnforced",
    "overview?.audit.advisory_only",
    "canExecuteActions: false",
    "canUseDurableStore: false",
    "canWriteBusinessTruth: false",
    "canReplayAutomatically: false",
)

FORBIDDEN_LITERALS = (
    "method: \"POST\"",
    "/v1/tools/actions\", {",
    "canExecuteActions: true",
    "canUseDurableStore: true",
    "canWriteBusinessTruth: true",
    "canReplayAutomatically: true",
    "executionEnabled: true",
    "writes_business_truth: true",
    "real_executor_enabled\": true",
    "durable_store_enabled\": true",
    "confirmation_flow_connected\": true",
    "business_truth_write_enabled\": true",
    "automatic_replay_enabled\": true",
)

README_LITERALS = (
    "连接失败诊断",
    "Dev Diagnostics",
    "refresh 期间保留上一份已加载 overview",
    "端口冲突",
    "CORS / preflight",
    "unsafe port",
    "local-smoke contract mismatch",
    "../../scripts/run-python.sh ../../scripts/check-radishmind-console-shell.py",
    "../../scripts/run-python.sh ../../scripts/check-radishmind-console-behavior.py",
    "不实现真实工具执行器",
    "不启用 automatic replay",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_literals(content: str, literals: tuple[str, ...], *, label: str) -> None:
    for literal in literals:
        require(literal in content, f"{label} missing literal: {literal}")


def require_absent(content: str, literals: tuple[str, ...], *, label: str) -> None:
    for literal in literals:
        require(literal not in content, f"{label} enables forbidden state: {literal}")


def main() -> int:
    app_content = APP_SOURCE.read_text(encoding="utf-8")
    client_content = CLIENT_SOURCE.read_text(encoding="utf-8")
    readme_content = README.read_text(encoding="utf-8")
    contract_content = OVERVIEW_CONTRACT.read_text(encoding="utf-8")
    local_smoke_contract_content = LOCAL_SMOKE_CONTRACT.read_text(encoding="utf-8")
    route_content = OVERVIEW_ROUTE.read_text(encoding="utf-8")
    combined_source = "\n".join([app_content, client_content, contract_content, local_smoke_contract_content, route_content])

    require_literals(app_content, READY_PATH_LITERALS, label="console ready path")
    require_literals(app_content, REFRESH_PATH_LITERALS, label="console refresh path")
    require_literals(app_content + "\n" + client_content, ERROR_DIAGNOSTIC_LITERALS, label="console error diagnostics")
    require_literals(combined_source, READ_ONLY_BOUNDARY_LITERALS, label="console read-only boundary")
    require_literals(readme_content, README_LITERALS, label="console README behavior guidance")
    require_absent(combined_source, FORBIDDEN_LITERALS, label="console source")

    require(
        "fetch(endpoint" in client_content and "response.json()" in client_content,
        "console behavior gate expects direct platform JSON fetch",
    )
    require(
        "loadState.status === \"error\" && readyState !== null" in app_content,
        "console must explicitly distinguish stale overview after connection failure",
    )
    require(
        "overview.audit.notes.map" in app_content,
        "console must render audit notes as readable items",
    )
    require(
        "viewModel.modelInventory.activeProfileChain" in app_content,
        "console must expose active profile chain from overview",
    )
    require(
        "viewModel.modelInventory.canShowProfileSelector" in app_content
        and "modelInventoryDetails.map" in app_content,
        "console must render provider/profile inventory details without extra execution capability",
    )
    require(
        "id.startsWith(\"provider:\") && id.includes(\":profile:\")" in app_content
        and "id.startsWith(\"profile:\")" in app_content,
        "console must explain provider/profile selectable model ids",
    )
    require(
        "readinessViewModel.blockedActionNoSideEffects" in app_content,
        "console must expose blocked action no-side-effects from local-smoke",
    )
    require(
        "STOP_LINE_DETAILS[item.id]" in app_content,
        "console must render stop-line details for blocked capabilities",
    )
    require(
        "readinessViewModel?.allStopLinesEnforced" in app_content and "overview?.audit.advisory_only" in app_content,
        "console must expose stop-line enforcement evidence",
    )
    require(
        'loadState.failureSurface === "platform_local_smoke"' in app_content,
        "console must distinguish local-smoke readiness failures from overview outages",
    )
    require(
        '"platform_local_smoke"' in client_content and '"platform_overview"' in client_content,
        "console client must classify overview and local-smoke failure surfaces",
    )

    print("platform local console behavior check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
