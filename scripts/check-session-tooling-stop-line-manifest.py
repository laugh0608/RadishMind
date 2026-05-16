#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

MANIFEST = FIXTURE_DIR / "session-tooling-stop-line-manifest.json"
CLOSE_CANDIDATE = FIXTURE_DIR / "session-tooling-close-candidate-readiness-rollup.json"
SHORT_CLOSE_DELTA = FIXTURE_DIR / "session-tooling-short-close-readiness-delta.json"
ENABLEMENT_PLAN = FIXTURE_DIR / "session-tooling-executor-storage-confirmation-enablement-plan.json"
SUITE_READINESS = FIXTURE_DIR / "session-tooling-negative-regression-suite-readiness.json"
ROUTE_NEGATIVE_MATRIX = FIXTURE_DIR / "session-tooling-route-negative-coverage-matrix.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-stop-line-manifest.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-stop-line-manifest.py"

HARD_PREREQUISITES = [
    "upper_layer_confirmation_flow",
    "complete_negative_regression_suite",
    "executor_storage_confirmation_enablement_plan",
    "durable_store_and_result_reader_policy",
]
REQUIRED_PROHIBITED_CLAIMS = [
    "p2_short_close",
    "real_tool_executor_ready",
    "durable_storage_ready",
    "confirmation_flow_connected",
    "materialized_result_reader_ready",
    "automatic_replay_ready",
    "complete_negative_regression_suite",
]
REQUIRED_STOP_LINES = [
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_audit_store",
    "durable_checkpoint_store",
    "durable_result_store",
    "durable_session_store",
    "durable_tool_store",
    "long_term_memory",
    "materialized_result_reader",
    "real_checkpoint_storage_backend",
    "real_tool_executor",
    "replay_executor",
    "result_ref_reader",
    "upper_layer_confirmation_flow_connection",
]
RELAXATION_SOURCES = [
    (CLOSE_CANDIDATE, "check-session-tooling-close-candidate-readiness-rollup.py"),
    (SHORT_CLOSE_DELTA, "check-session-tooling-short-close-readiness-delta.py"),
    (ENABLEMENT_PLAN, "check-session-tooling-executor-storage-confirmation-enablement-plan.py"),
    (SUITE_READINESS, "check-session-tooling-negative-regression-suite-readiness.py"),
    (ROUTE_NEGATIVE_MATRIX, "check-session-tooling-route-negative-coverage-matrix.py"),
    (MANIFEST, "check-session-tooling-stop-line-manifest.py"),
]


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def relative_path(path: Path) -> str:
    return path.relative_to(REPO_ROOT).as_posix()


def require_object(value: Any, message: str) -> dict[str, Any]:
    require(isinstance(value, dict), message)
    return value


def require_string_list(value: Any, message: str) -> list[str]:
    require(isinstance(value, list) and value, message)
    result: list[str] = []
    for item in value:
        text = str(item or "").strip()
        require(bool(text), message)
        result.append(text)
    return result


def hard_prerequisite_blockers(short_close: dict[str, Any]) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    for item in short_close.get("hard_prerequisites") or []:
        row = require_object(item, "short close hard prerequisite must be object")
        prerequisite_id = str(row.get("prerequisite_id") or "").strip()
        require(prerequisite_id in HARD_PREREQUISITES, f"unexpected prerequisite: {prerequisite_id}")
        require(row.get("current_status") == "not_satisfied", f"{prerequisite_id} must remain not_satisfied")
        rows.append(
            {
                "blocker_id": prerequisite_id,
                "status": "not_satisfied",
                "source": relative_path(SHORT_CLOSE_DELTA),
                "evidence": str(row.get("current_evidence") or "").strip(),
                "blocks_claims": require_string_list(row.get("blocks_claims"), "hard prerequisite must block claims"),
            }
        )
    require([row["blocker_id"] for row in rows] == HARD_PREREQUISITES, "hard prerequisite order drifted")
    return rows


def stop_line_capabilities(
    close_candidate: dict[str, Any],
    short_close: dict[str, Any],
    enablement_plan: dict[str, Any],
    suite_readiness: dict[str, Any],
) -> list[dict[str, Any]]:
    close_blocked = set(require_string_list(close_candidate.get("blocked_capabilities"), "close candidate blocked capabilities required"))
    short_blocked = set(require_string_list(short_close.get("do_not_implement_now"), "short close stop lines required"))
    suite_blocked = set(require_string_list(suite_readiness.get("not_enabled_capabilities"), "suite readiness stop lines required"))
    enablement_blocked: set[str] = set()
    for area in enablement_plan.get("enablement_areas") or []:
        area_obj = require_object(area, "enablement area must be object")
        require(area_obj.get("current_status") == "blocked_not_gated_plan", "enablement area must remain blocked")
        enablement_blocked.update(require_string_list(area_obj.get("still_forbidden"), "enablement area stop lines required"))

    rows: list[dict[str, Any]] = []
    for capability in REQUIRED_STOP_LINES:
        sources: list[str] = []
        if capability in close_blocked:
            sources.append(relative_path(CLOSE_CANDIDATE))
        if capability in short_blocked:
            sources.append(relative_path(SHORT_CLOSE_DELTA))
        if capability in enablement_blocked:
            sources.append(relative_path(ENABLEMENT_PLAN))
        if capability in suite_blocked:
            sources.append(relative_path(SUITE_READINESS))
        require(sources, f"{capability} must have at least one stop-line source")
        rows.append({"capability": capability, "status": "blocked", "evidence_sources": sources})
    return rows


def blocker_source_map() -> list[dict[str, Any]]:
    return [
        {
            "blocker_id": "upper_layer_confirmation_flow",
            "evidence_sources": [relative_path(CLOSE_CANDIDATE), relative_path(SHORT_CLOSE_DELTA), relative_path(ENABLEMENT_PLAN)],
        },
        {
            "blocker_id": "complete_negative_regression_suite",
            "evidence_sources": [
                relative_path(CLOSE_CANDIDATE),
                relative_path(SHORT_CLOSE_DELTA),
                relative_path(SUITE_READINESS),
                relative_path(ROUTE_NEGATIVE_MATRIX),
            ],
        },
        {
            "blocker_id": "executor_storage_confirmation_enablement_plan",
            "evidence_sources": [relative_path(SHORT_CLOSE_DELTA), relative_path(ENABLEMENT_PLAN)],
        },
        {
            "blocker_id": "durable_store_and_result_reader_policy",
            "evidence_sources": [relative_path(CLOSE_CANDIDATE), relative_path(SHORT_CLOSE_DELTA), relative_path(ENABLEMENT_PLAN)],
        },
    ]


def build_manifest() -> dict[str, Any]:
    close_candidate = require_object(load_json_document(CLOSE_CANDIDATE), "close candidate rollup must be object")
    short_close = require_object(load_json_document(SHORT_CLOSE_DELTA), "short close delta must be object")
    enablement_plan = require_object(load_json_document(ENABLEMENT_PLAN), "enablement plan must be object")
    suite_readiness = require_object(load_json_document(SUITE_READINESS), "suite readiness must be object")
    route_matrix = require_object(load_json_document(ROUTE_NEGATIVE_MATRIX), "route negative matrix must be object")

    require(close_candidate.get("status") == "close_candidate_governance_only", "close candidate must remain governance-only")
    require(short_close.get("status") == "short_close_blocked", "short close delta must remain blocked")
    require(enablement_plan.get("status") == "enablement_plan_defined_blocked", "enablement plan must remain blocked")
    require(
        suite_readiness.get("status") == "governance_suite_consumed_deny_by_default_gates_defined",
        "suite readiness must remain governance-only",
    )
    require(route_matrix.get("status") == "route_negative_coverage_matrix_governance_only", "route matrix must remain governance-only")

    route_totals = require_object(route_matrix.get("matrix_totals"), "route matrix totals required")
    require(route_totals.get("suite_completion_blocker_cases") == 7, "route matrix must keep seven suite blockers")
    require(route_totals.get("future_route_requirements_satisfied") == 0, "future route requirements must remain unsatisfied")

    return {
        "schema_version": 1,
        "kind": "session_tooling_stop_line_manifest",
        "stage": "P2 Session & Tooling Foundation",
        "status": "stop_lines_governance_only",
        "implementation_status": "not_ready",
        "source_close_candidate_rollup": relative_path(CLOSE_CANDIDATE),
        "source_short_close_readiness_delta": relative_path(SHORT_CLOSE_DELTA),
        "source_executor_storage_confirmation_enablement_plan": relative_path(ENABLEMENT_PLAN),
        "source_negative_regression_suite_readiness": relative_path(SUITE_READINESS),
        "source_route_negative_coverage_matrix": relative_path(ROUTE_NEGATIVE_MATRIX),
        "hard_prerequisite_blockers": hard_prerequisite_blockers(short_close),
        "stop_line_capabilities": stop_line_capabilities(close_candidate, short_close, enablement_plan, suite_readiness),
        "suite_blocker_summary": {
            "negative_suite_cases": route_totals.get("suite_case_count"),
            "current_route_covered_suite_cases": route_totals.get("current_route_covered_suite_cases"),
            "future_route_required_suite_cases": route_totals.get("governance_only_future_route_required_cases"),
            "future_route_requirements_satisfied": route_totals.get("future_route_requirements_satisfied"),
            "suite_completion_blocker_cases": route_totals.get("suite_completion_blocker_cases"),
        },
        "blocker_source_map": blocker_source_map(),
        "required_updates_before_relaxing_stop_line": [
            {"fixture": relative_path(path), "check": check_name}
            for path, check_name in RELAXATION_SOURCES
        ],
        "current_allowed_claims": [
            "stop_lines_manifest_checkable",
            "hard_prerequisite_blockers_identified",
            "stop_line_evidence_sources_checkable",
            "governance_only_stop_line_status",
        ],
        "prohibited_claims": REQUIRED_PROHIBITED_CLAIMS,
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_stop_line_manifest",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_manifest_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "stop-line manifest schema_version must be 1")
    require(document.get("kind") == "session_tooling_stop_line_manifest", "stop-line manifest kind mismatch")
    require(document.get("status") == "stop_lines_governance_only", "stop-line manifest status drifted")
    require(document.get("implementation_status") == "not_ready", "stop-line manifest must not claim implementation readiness")

    blockers = document.get("hard_prerequisite_blockers")
    require(isinstance(blockers, list) and len(blockers) == 4, "stop-line manifest must include four hard blockers")
    require([str(row.get("blocker_id") or "").strip() for row in blockers if isinstance(row, dict)] == HARD_PREREQUISITES, "hard blockers drifted")
    for item in blockers:
        row = require_object(item, "hard blocker must be object")
        require(row.get("status") == "not_satisfied", "hard blockers must remain not_satisfied")

    stop_lines = document.get("stop_line_capabilities")
    require(isinstance(stop_lines, list) and len(stop_lines) == len(REQUIRED_STOP_LINES), "stop-line capability count drifted")
    require([str(row.get("capability") or "").strip() for row in stop_lines if isinstance(row, dict)] == REQUIRED_STOP_LINES, "stop-line capability order drifted")
    for item in stop_lines:
        row = require_object(item, "stop-line capability must be object")
        require(row.get("status") == "blocked", "stop-line capabilities must remain blocked")
        require_string_list(row.get("evidence_sources"), "stop-line capability must include evidence sources")

    suite_summary = require_object(document.get("suite_blocker_summary"), "suite blocker summary required")
    require(suite_summary.get("negative_suite_cases") == 9, "negative suite case count drifted")
    require(suite_summary.get("current_route_covered_suite_cases") == 2, "current route covered case count drifted")
    require(suite_summary.get("future_route_required_suite_cases") == 7, "future route required case count drifted")
    require(suite_summary.get("future_route_requirements_satisfied") == 0, "future route requirements must remain unsatisfied")
    require(suite_summary.get("suite_completion_blocker_cases") == 7, "suite blocker case count drifted")

    source_map = document.get("blocker_source_map")
    require(isinstance(source_map, list) and len(source_map) == 4, "blocker source map must include four rows")
    require([str(row.get("blocker_id") or "").strip() for row in source_map if isinstance(row, dict)] == HARD_PREREQUISITES, "source map blockers drifted")
    for item in source_map:
        row = require_object(item, "source map row must be object")
        require_string_list(row.get("evidence_sources"), "source map row must include evidence sources")

    change_control = document.get("required_updates_before_relaxing_stop_line")
    require(isinstance(change_control, list) and len(change_control) == len(RELAXATION_SOURCES), "change-control source count drifted")
    prohibited = set(require_string_list(document.get("prohibited_claims"), "prohibited claims required"))
    missing = sorted(set(REQUIRED_PROHIBITED_CLAIMS) - prohibited)
    require(not missing, f"stop-line manifest missing prohibited claims: {missing}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    fixture_name = MANIFEST.name
    require(THIS_CHECK.name in check_repo, "fast baseline must run stop-line manifest check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference stop-line manifest task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", CONTRACTS_README.read_text(encoding="utf-8")),
        ("current focus", CURRENT_FOCUS.read_text(encoding="utf-8")),
        ("capability matrix", CAPABILITY_MATRIX.read_text(encoding="utf-8")),
        ("roadmap", ROADMAP.read_text(encoding="utf-8")),
        ("devlog", DEVLOG.read_text(encoding="utf-8")),
    ):
        require(fixture_name in content, f"{label} must reference stop-line manifest fixture")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("P2 short close" in content, f"{label} must mention P2 short close")
        require("not_satisfied" in content, f"{label} must mention not_satisfied")
    for marker in ("不实现真实工具执行器", "不实现 durable", "不启用 automatic replay"):
        require(marker in task_card, f"task card missing marker: {marker}")


def main() -> int:
    expected = require_object(load_json_document(MANIFEST), "stop-line manifest fixture must be object")
    check_manifest_shape(expected)
    actual = build_manifest()
    if actual != expected:
        raise SystemExit("session/tooling stop-line manifest does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling stop-line manifest checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
