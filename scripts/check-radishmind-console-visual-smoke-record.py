#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/p3-local-console-visual-smoke-record.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
P3_CHECKLIST_PATH = REPO_ROOT / "scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json"

REQUIRED_SCENARIOS = {
    "desktop_ready_overview",
    "narrow_ready_overview",
    "minimum_width_320_ready_overview",
    "desktop_ready_local_readiness",
    "connection_failure_stale_overview",
}

REQUIRED_READY_SURFACES = {
    "service_status",
    "model_inventory",
    "session_tooling_surface",
    "stop_lines",
    "dev_diagnostics",
    "local_readiness_panel",
    "local_smoke_readiness_status",
    "local_console_ready_summary",
    "readiness_checks",
    "local_cors_origins",
    "local_smoke_failure_hints",
}

REQUIRED_LOCAL_READINESS_SURFACES = {
    "local_readiness_panel",
    "local_smoke_readiness_status",
    "local_console_ready_summary",
    "healthz_readiness",
    "overview_contract_readiness",
    "model_inventory_readiness",
    "session_tooling_metadata_readiness",
    "blocked_action_no_side_effects",
    "stop_lines_enforced",
    "local_cors_origins",
    "local_smoke_failure_hints",
}

FORBIDDEN_LOCAL_READINESS_SURFACES = {
    "production_health_dashboard",
    "process_supervisor",
    "executor_action_button",
    "confirmation_button",
    "durable_store_control",
    "business_truth_write_button",
    "automatic_replay_control",
}

REQUIRED_STOP_LINES = {
    "real_executor_enabled",
    "durable_store_enabled",
    "confirmation_flow_connected",
    "materialized_result_reader",
    "long_term_memory_enabled",
    "business_truth_write_enabled",
    "automatic_replay_enabled",
}

PORT_SOURCE_FILES = {
    "apps/radishmind-console/package.json": [
        "vite --host 127.0.0.1 --port 4000",
        "vite preview --host 127.0.0.1 --port 4000",
    ],
    "apps/radishmind-console/src/platformOverviewClient.ts": [
        'DEFAULT_PLATFORM_BASE_URL = "http://127.0.0.1:7000"',
        "http://127.0.0.1:4000",
        "http://localhost:4000",
    ],
    "services/platform/internal/config/config.go": [
        'defaultListenAddr        = ":7000"',
    ],
    "services/platform/internal/httpapi/server.go": [
        "http://127.0.0.1:4000",
        "http://localhost:4000",
    ],
    "apps/radishmind-console/README.md": [
        "http://127.0.0.1:7000/v1/platform/overview",
        "http://127.0.0.1:7000/v1/platform/local-smoke",
        "http://127.0.0.1:4000",
    ],
}

LOCAL_READINESS_SOURCE_LITERALS = {
    "apps/radishmind-console/src/App.tsx": [
        'Panel title="Local Readiness"',
        "readinessViewModel?.status",
        "readinessViewModel?.localConsoleReady",
        "readinessViewModel.healthzOk",
        "readinessViewModel.overviewContractReadable",
        "readinessViewModel.modelInventoryReadable",
        "readinessViewModel.sessionToolingMetadataReadable",
        "readinessViewModel.blockedActionNoSideEffects",
        "readinessViewModel.allStopLinesEnforced",
        "readinessViewModel.allowedCorsOrigins",
        "localSmoke.failure_hints.map",
        "Local CORS origins",
        "Failure hints",
    ],
    "apps/radishmind-console/src/platformOverviewClient.ts": [
        "buildPlatformLocalSmokeEndpoint",
        "fetchPlatformLocalSmoke",
        "toPlatformLocalSmokeReadinessViewModel",
        "localSmokeEndpoint",
        "readinessViewModel",
    ],
    "contracts/typescript/platform-local-smoke-api.ts": [
        "PlatformLocalSmokeReadinessViewModel",
        "localConsoleReady",
        "blockedActionNoSideEffects",
        "canExecuteActions: false",
        "canUseDurableStore: false",
        "canWriteBusinessTruth: false",
        "canReplayAutomatically: false",
    ],
}

OLD_PORT_LITERALS = (
    "127.0.0.1:8080",
    "localhost:8080",
    "127.0.0.1:5173",
    "localhost:5173",
    "127.0.0.1:6000",
    "localhost:6000",
    "127.0.0.1:6001",
    "localhost:6001",
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def require_literals(path: Path, literals: list[str]) -> None:
    content = path.read_text(encoding="utf-8")
    for literal in literals:
        require(literal in content, f"{path.relative_to(REPO_ROOT)} missing literal: {literal}")
    for old_port in OLD_PORT_LITERALS:
        require(old_port not in content, f"{path.relative_to(REPO_ROOT)} still references old port: {old_port}")


def assert_scope(fixture: dict[str, Any]) -> None:
    scope = fixture.get("qa_scope") or {}
    require(scope.get("surface") == "apps/radishmind-console", "visual smoke must target local console")
    require(scope.get("backend_default_url") == "http://127.0.0.1:7000", "backend default port must be 7000")
    require(scope.get("frontend_default_url") == "http://127.0.0.1:4000", "frontend default port must be 4000")
    require(
        scope.get("overview_endpoint") == "http://127.0.0.1:7000/v1/platform/overview",
        "overview endpoint must use backend port 7000",
    )
    require(
        scope.get("local_smoke_endpoint") == "http://127.0.0.1:7000/v1/platform/local-smoke",
        "local-smoke endpoint must use backend port 7000",
    )
    require(scope.get("read_only") is True, "visual smoke must stay read-only")
    require(scope.get("production_packaging_ready") is False, "visual smoke must not claim production packaging")


def assert_method(fixture: dict[str, Any]) -> None:
    method = fixture.get("qa_method") or {}
    require(method.get("kind") == "browser_visual_qa_record", "unexpected visual smoke method")
    require(method.get("automated_browser_required") is False, "fast check must not require browser automation")
    require(method.get("committed_screenshots_required") is False, "visual screenshots must not be committed evidence")


def assert_scenarios(fixture: dict[str, Any]) -> None:
    scenarios = fixture.get("required_scenarios") or []
    scenario_by_id = {str(item.get("id")): item for item in scenarios}
    missing = sorted(REQUIRED_SCENARIOS - set(scenario_by_id))
    require(not missing, f"missing visual smoke scenarios: {missing}")

    ready_surfaces = set()
    for scenario_id in ("desktop_ready_overview", "narrow_ready_overview", "desktop_ready_local_readiness"):
        ready_surfaces.update(scenario_by_id[scenario_id].get("must_show") or [])
    missing_surfaces = sorted(REQUIRED_READY_SURFACES - ready_surfaces)
    require(not missing_surfaces, f"ready visual smoke missing surfaces: {missing_surfaces}")

    local_readiness = scenario_by_id["desktop_ready_local_readiness"]
    require(local_readiness.get("state") == "ready", "local readiness visual smoke must use ready state")
    missing_local_readiness = sorted(
        REQUIRED_LOCAL_READINESS_SURFACES - set(local_readiness.get("must_show") or [])
    )
    require(not missing_local_readiness, f"local readiness visual smoke missing surfaces: {missing_local_readiness}")
    missing_forbidden = sorted(
        FORBIDDEN_LOCAL_READINESS_SURFACES - set(local_readiness.get("must_not_show") or [])
    )
    require(not missing_forbidden, f"local readiness visual smoke missing forbidden surfaces: {missing_forbidden}")

    narrow_expectations = set(scenario_by_id["narrow_ready_overview"].get("layout_expectations") or [])
    narrow_surfaces = set(scenario_by_id["narrow_ready_overview"].get("must_show") or [])
    require("local_probe_commands" in narrow_surfaces, "narrow visual smoke must show local probe commands")
    require("local_readiness_panel" in narrow_surfaces, "narrow visual smoke must show local readiness panel")
    require("no_horizontal_overflow" in narrow_expectations, "narrow visual smoke must assert no horizontal overflow")
    require("long_routes_wrap" in narrow_expectations, "narrow visual smoke must assert route wrapping")
    require(
        "local_cors_origins" in narrow_surfaces,
        "narrow visual smoke must show local CORS origins from local-smoke",
    )

    min_width = scenario_by_id["minimum_width_320_ready_overview"]
    require(min_width.get("viewport") == "320px", "minimum width scenario must be 320px")
    require(
        "scroll_width_matches_viewport_width" in set(min_width.get("layout_expectations") or []),
        "minimum width scenario must assert scroll width",
    )
    require(
        "dev_diagnostics_commands_wrap" in set(min_width.get("layout_expectations") or []),
        "minimum width scenario must assert dev diagnostics command wrapping",
    )
    require(
        "local_smoke_endpoint_wraps" in set(min_width.get("layout_expectations") or []),
        "minimum width scenario must assert local-smoke endpoint wrapping",
    )
    require(
        "local_readiness_tokens_wrap" in set(min_width.get("layout_expectations") or []),
        "minimum width scenario must assert local readiness token wrapping",
    )

    error_state = scenario_by_id["connection_failure_stale_overview"]
    require(error_state.get("state") == "error_with_previous_overview", "error scenario must keep previous overview")
    required_error_surfaces = {
        "platform_service_unavailable_notice",
        "connection_failed_showing_last_overview",
        "dev_diagnostics_failure_classes",
        "diagnostic_list",
        "previous_local_readiness_panel",
        "previous_local_smoke_readiness_status",
        "previous_stop_lines",
    }
    missing_error_surfaces = sorted(required_error_surfaces - set(error_state.get("must_show") or []))
    require(not missing_error_surfaces, f"error visual smoke missing surfaces: {missing_error_surfaces}")


def assert_source_evidence(fixture: dict[str, Any]) -> None:
    for relpath in fixture.get("source_evidence") or []:
        require((REPO_ROOT / str(relpath)).is_file(), f"missing visual smoke evidence: {relpath}")

    for relpath, literals in PORT_SOURCE_FILES.items():
        require_literals(REPO_ROOT / relpath, literals)

    for relpath, literals in LOCAL_READINESS_SOURCE_LITERALS.items():
        require_literals(REPO_ROOT / relpath, literals)


def assert_stop_lines(fixture: dict[str, Any]) -> None:
    stop_lines = fixture.get("stop_lines") or {}
    for stop_line in REQUIRED_STOP_LINES:
        require(stop_lines.get(stop_line) is False, f"visual smoke must keep {stop_line}=false")


def assert_consumers(fixture: dict[str, Any]) -> None:
    required_consumers = set(fixture.get("required_consumers") or [])
    expected_consumers = {
        "scripts/check-radishmind-console-visual-smoke-record.py",
        "scripts/check-repo.py",
        "scripts/README.md",
        "scripts/checks/fixtures/p3-local-product-shell-short-close-checklist.json",
    }
    missing = sorted(expected_consumers - required_consumers)
    require(not missing, f"missing visual smoke consumers: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-radishmind-console-visual-smoke-record.py", [])' in check_repo,
        "check-repo.py must run the console visual smoke record check",
    )

    p3_checklist = load_json(P3_CHECKLIST_PATH)
    satisfied_ids = {item.get("id") for item in p3_checklist.get("satisfied_conditions") or []}
    require("console_visual_smoke_record" in satisfied_ids, "P3 checklist must consume visual smoke record")

    scripts_readme = (REPO_ROOT / "scripts/README.md").read_text(encoding="utf-8")
    require(
        "check-radishmind-console-visual-smoke-record.py" in scripts_readme,
        "scripts README must document visual smoke record check",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "p3_local_console_visual_smoke_record", "unexpected fixture kind")
    require(fixture.get("stage") == "P3 Local Product Shell / Ops Surface", "unexpected stage")
    assert_scope(fixture)
    assert_method(fixture)
    assert_scenarios(fixture)
    assert_source_evidence(fixture)
    assert_stop_lines(fixture)
    assert_consumers(fixture)
    print("RadishMind console visual smoke record check passed.")


if __name__ == "__main__":
    main()
