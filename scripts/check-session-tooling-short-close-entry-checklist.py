#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

CHECKLIST = FIXTURE_DIR / "session-tooling-short-close-entry-checklist.json"
STOP_LINE_MANIFEST = FIXTURE_DIR / "session-tooling-stop-line-manifest.json"
SHORT_CLOSE_DELTA = FIXTURE_DIR / "session-tooling-short-close-readiness-delta.json"
ROUTE_SMOKE_READINESS = FIXTURE_DIR / "session-tooling-route-smoke-readiness-rollup.json"
NEGATIVE_SUITE_READINESS = FIXTURE_DIR / "session-tooling-negative-regression-suite-readiness.json"
UPPER_LAYER_CONFIRMATION_READINESS = FIXTURE_DIR / "session-tooling-upper-layer-confirmation-flow-readiness.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-short-close-entry-checklist.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
DOCS_README = REPO_ROOT / "docs/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-short-close-entry-checklist.py"

ENTRY_CONDITIONS = [
    "upper_layer_confirmation_flow",
    "complete_negative_regression_suite",
    "executor_storage_confirmation_enablement_plan",
    "durable_store_and_result_reader_policy",
]
ROUTE_SMOKE_REQUIREMENTS = [
    "executor_gate_route_smoke",
    "storage_materialization_gate_route_smoke",
    "confirmation_gate_route_smoke",
    "durable_store_reader_route_smoke",
]
REQUIRED_PROHIBITED_CLAIMS = {
    "p2_short_close",
    "real_tool_executor_ready",
    "durable_storage_ready",
    "confirmation_flow_connected",
    "materialized_result_reader_ready",
    "automatic_replay_ready",
    "complete_negative_regression_suite",
}
REQUIRED_NOT_IMPLEMENT_NOW = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_audit_store",
    "durable_checkpoint_store",
    "durable_result_store",
    "durable_session_store",
    "long_term_memory",
    "materialized_result_reader",
    "real_tool_executor",
    "upper_layer_confirmation_flow_connection",
}


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
    items: list[str] = []
    for item in value:
        text = str(item or "").strip()
        require(bool(text), message)
        items.append(text)
    return items


def hard_prerequisites_by_id(delta: dict[str, Any]) -> dict[str, dict[str, Any]]:
    rows: dict[str, dict[str, Any]] = {}
    for item in delta.get("hard_prerequisites") or []:
        row = require_object(item, "short close hard prerequisite must be object")
        condition_id = str(row.get("prerequisite_id") or "").strip()
        require(condition_id in ENTRY_CONDITIONS, f"unexpected hard prerequisite: {condition_id}")
        require(row.get("current_status") == "not_satisfied", f"{condition_id} must remain not_satisfied")
        rows[condition_id] = row
    require(list(rows) == ENTRY_CONDITIONS, "short close hard prerequisite order drifted")
    return rows


def stop_line_blockers_by_id(manifest: dict[str, Any]) -> dict[str, dict[str, Any]]:
    rows: dict[str, dict[str, Any]] = {}
    for item in manifest.get("hard_prerequisite_blockers") or []:
        row = require_object(item, "stop-line hard prerequisite blocker must be object")
        blocker_id = str(row.get("blocker_id") or "").strip()
        require(blocker_id in ENTRY_CONDITIONS, f"unexpected stop-line blocker: {blocker_id}")
        require(row.get("status") == "not_satisfied", f"{blocker_id} stop-line blocker must remain not_satisfied")
        rows[blocker_id] = row
    require(list(rows) == ENTRY_CONDITIONS, "stop-line hard blocker order drifted")
    return rows


def build_entry_condition(
    condition_id: str,
    delta_row: dict[str, Any],
    *,
    source: Path,
    current_evidence: str | None = None,
) -> dict[str, Any]:
    return {
        "condition_id": condition_id,
        "status": "not_satisfied",
        "source": relative_path(source),
        "stop_line_source": relative_path(STOP_LINE_MANIFEST),
        "current_evidence": current_evidence or str(delta_row.get("current_evidence") or "").strip(),
        "must_satisfy_before_entry": require_string_list(
            delta_row.get("required_delta"),
            f"{condition_id} must include required delta",
        ),
        "blocks_claims": require_string_list(delta_row.get("blocks_claims"), f"{condition_id} must block claims"),
    }


def build_entry_conditions(
    delta: dict[str, Any],
    stop_line_manifest: dict[str, Any],
    suite_readiness: dict[str, Any],
    upper_layer_confirmation_readiness: dict[str, Any],
) -> list[dict[str, Any]]:
    delta_rows = hard_prerequisites_by_id(delta)
    stop_rows = stop_line_blockers_by_id(stop_line_manifest)
    for condition_id in ENTRY_CONDITIONS:
        require(
            condition_id in stop_rows,
            f"{condition_id} must be mirrored by stop-line manifest",
        )

    require(
        suite_readiness.get("implementation_status") == "not_ready",
        "negative suite readiness must keep implementation not_ready",
    )
    require(
        upper_layer_confirmation_readiness.get("status") == "readiness_defined_not_connected",
        "upper-layer confirmation readiness must remain not connected",
    )
    upper_layer_scope = require_object(
        upper_layer_confirmation_readiness.get("readiness_scope"),
        "upper-layer confirmation readiness scope required",
    )
    require(
        upper_layer_scope.get("entry_condition") == "upper_layer_confirmation_flow"
        and upper_layer_scope.get("current_status") == "not_satisfied",
        "upper-layer confirmation readiness must keep entry condition not_satisfied",
    )
    upper_layer_condition = build_entry_condition(
        "upper_layer_confirmation_flow",
        delta_rows["upper_layer_confirmation_flow"],
        source=SHORT_CLOSE_DELTA,
    )
    upper_layer_condition["readiness_source"] = relative_path(UPPER_LAYER_CONFIRMATION_READINESS)
    return [
        upper_layer_condition,
        build_entry_condition(
            "complete_negative_regression_suite",
            delta_rows["complete_negative_regression_suite"],
            source=NEGATIVE_SUITE_READINESS,
            current_evidence=(
                "governance suite, suite readiness, deny-by-default gates, and route coverage matrix exist, "
                "but real implementation consumers are missing"
            ),
        ),
        build_entry_condition(
            "executor_storage_confirmation_enablement_plan",
            delta_rows["executor_storage_confirmation_enablement_plan"],
            source=SHORT_CLOSE_DELTA,
            current_evidence="executor, storage, and confirmation enablement plan is defined but remains blocked_not_gated_plan",
        ),
        build_entry_condition(
            "durable_store_and_result_reader_policy",
            delta_rows["durable_store_and_result_reader_policy"],
            source=SHORT_CLOSE_DELTA,
        ),
    ]


def build_route_smoke_requirements(route_smoke: dict[str, Any]) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    for item in route_smoke.get("future_route_smoke_requirements") or []:
        row = require_object(item, "future route smoke requirement must be object")
        requirement_id = str(row.get("requirement_id") or "").strip()
        if requirement_id not in ROUTE_SMOKE_REQUIREMENTS:
            continue
        require(row.get("status") == "not_satisfied", f"{requirement_id} must remain not_satisfied")
        rows.append(
            {
                "requirement_id": requirement_id,
                "status": "not_satisfied",
                "source": relative_path(ROUTE_SMOKE_READINESS),
            }
        )
    require([row["requirement_id"] for row in rows] == ROUTE_SMOKE_REQUIREMENTS, "route smoke requirement order drifted")
    return rows


def build_negative_suite_summary(suite_readiness: dict[str, Any]) -> dict[str, Any]:
    summary = require_object(suite_readiness.get("route_negative_coverage_summary"), "route negative coverage summary required")
    require(summary.get("current_route_covered_suite_cases") == 2, "route-covered suite case count drifted")
    require(summary.get("governance_only_future_route_required_cases") == 7, "future-route-required suite case count drifted")
    require(summary.get("future_route_requirements_satisfied") == 0, "future route requirements must remain unsatisfied")
    require(summary.get("suite_completion_blocker_cases") == 7, "suite completion blocker count drifted")
    skeleton = require_object(suite_readiness.get("current_skeleton_coverage"), "suite skeleton coverage required")
    require(skeleton.get("case_count") == 9, "negative suite case count drifted")
    return {
        "negative_suite_cases": skeleton.get("case_count"),
        "current_route_covered_suite_cases": summary.get("current_route_covered_suite_cases"),
        "future_route_required_suite_cases": summary.get("governance_only_future_route_required_cases"),
        "future_route_requirements_satisfied": summary.get("future_route_requirements_satisfied"),
        "suite_completion_blocker_cases": summary.get("suite_completion_blocker_cases"),
        "source": relative_path(NEGATIVE_SUITE_READINESS),
    }


def build_checklist() -> dict[str, Any]:
    stop_line = require_object(load_json_document(STOP_LINE_MANIFEST), "stop-line manifest must be object")
    delta = require_object(load_json_document(SHORT_CLOSE_DELTA), "short close delta must be object")
    route_smoke = require_object(load_json_document(ROUTE_SMOKE_READINESS), "route smoke readiness must be object")
    suite_readiness = require_object(load_json_document(NEGATIVE_SUITE_READINESS), "negative suite readiness must be object")
    upper_layer_confirmation_readiness = require_object(
        load_json_document(UPPER_LAYER_CONFIRMATION_READINESS),
        "upper-layer confirmation readiness must be object",
    )

    require(stop_line.get("status") == "stop_lines_governance_only", "stop-line manifest must stay governance-only")
    require(delta.get("status") == "short_close_blocked", "short close delta must remain blocked")
    require(route_smoke.get("status") == "route_smoke_readiness_governance_only", "route smoke readiness must stay governance-only")
    require(
        suite_readiness.get("status") == "governance_suite_consumed_deny_by_default_gates_defined",
        "negative suite readiness status drifted",
    )

    route_totals = require_object(route_smoke.get("coverage_totals"), "route smoke coverage totals required")
    require(route_totals.get("future_requirement_count") == 4, "future route smoke requirement count drifted")
    require(route_totals.get("future_requirements_satisfied") == 0, "future route smoke requirements must remain unsatisfied")
    require(route_totals.get("short_close_hard_prerequisites_not_satisfied") == 4, "short close blocker count drifted")

    entry_conditions = build_entry_conditions(delta, stop_line, suite_readiness, upper_layer_confirmation_readiness)
    route_requirements = build_route_smoke_requirements(route_smoke)

    return {
        "schema_version": 1,
        "kind": "session_tooling_short_close_entry_checklist",
        "stage": "P2 Session & Tooling Foundation",
        "current_state": "close_candidate_governance_only",
        "target_state": "p2_short_close",
        "status": "entry_blocked_governance_only",
        "implementation_status": "not_ready",
        "source_stop_line_manifest": relative_path(STOP_LINE_MANIFEST),
        "source_short_close_readiness_delta": relative_path(SHORT_CLOSE_DELTA),
        "source_route_smoke_readiness_rollup": relative_path(ROUTE_SMOKE_READINESS),
        "source_negative_regression_suite_readiness": relative_path(NEGATIVE_SUITE_READINESS),
        "source_upper_layer_confirmation_flow_readiness": relative_path(UPPER_LAYER_CONFIRMATION_READINESS),
        "entry_conditions": entry_conditions,
        "route_smoke_entry_requirements": route_requirements,
        "negative_suite_entry_summary": build_negative_suite_summary(suite_readiness),
        "entry_totals": {
            "condition_count": len(entry_conditions),
            "satisfied_count": 0,
            "not_satisfied_count": len(entry_conditions),
            "future_route_smoke_requirement_count": len(route_requirements),
            "future_route_smoke_satisfied_count": 0,
        },
        "current_allowed_claims": [
            "short_close_entry_checklist_checkable",
            "entry_conditions_identified",
            "entry_still_blocked",
            "governance_only_entry_status",
        ],
        "prohibited_claims": [
            "p2_short_close",
            "real_tool_executor_ready",
            "durable_storage_ready",
            "confirmation_flow_connected",
            "materialized_result_reader_ready",
            "automatic_replay_ready",
            "complete_negative_regression_suite",
        ],
        "do_not_implement_now": [
            "automatic_replay",
            "business_truth_write",
            "confirmed_action_execution",
            "durable_audit_store",
            "durable_checkpoint_store",
            "durable_result_store",
            "durable_session_store",
            "long_term_memory",
            "materialized_result_reader",
            "real_tool_executor",
            "upper_layer_confirmation_flow_connection",
        ],
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_short_close_entry_checklist",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_checklist_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "entry checklist schema_version must be 1")
    require(document.get("kind") == "session_tooling_short_close_entry_checklist", "entry checklist kind mismatch")
    require(document.get("current_state") == "close_candidate_governance_only", "entry checklist current state drifted")
    require(document.get("target_state") == "p2_short_close", "entry checklist target state drifted")
    require(document.get("status") == "entry_blocked_governance_only", "entry checklist must remain blocked")
    require(document.get("implementation_status") == "not_ready", "entry checklist must not claim implementation readiness")
    require(
        document.get("source_upper_layer_confirmation_flow_readiness") == relative_path(UPPER_LAYER_CONFIRMATION_READINESS),
        "entry checklist must reference upper-layer confirmation readiness",
    )

    conditions = document.get("entry_conditions")
    require(isinstance(conditions, list) and len(conditions) == len(ENTRY_CONDITIONS), "entry condition count drifted")
    require(
        [str(row.get("condition_id") or "").strip() for row in conditions if isinstance(row, dict)] == ENTRY_CONDITIONS,
        "entry condition order drifted",
    )
    for item in conditions:
        row = require_object(item, "entry condition must be object")
        require(row.get("status") == "not_satisfied", "entry condition must remain not_satisfied")
        require_string_list(row.get("must_satisfy_before_entry"), "entry condition must include required evidence")
        require_string_list(row.get("blocks_claims"), "entry condition must include blocked claims")
        if row.get("condition_id") == "upper_layer_confirmation_flow":
            require(
                row.get("readiness_source") == relative_path(UPPER_LAYER_CONFIRMATION_READINESS),
                "upper_layer_confirmation_flow entry condition must reference readiness fixture",
            )

    route_requirements = document.get("route_smoke_entry_requirements")
    require(
        isinstance(route_requirements, list) and len(route_requirements) == len(ROUTE_SMOKE_REQUIREMENTS),
        "route smoke entry requirement count drifted",
    )
    require(
        [str(row.get("requirement_id") or "").strip() for row in route_requirements if isinstance(row, dict)]
        == ROUTE_SMOKE_REQUIREMENTS,
        "route smoke entry requirement order drifted",
    )
    for item in route_requirements:
        row = require_object(item, "route smoke entry requirement must be object")
        require(row.get("status") == "not_satisfied", "route smoke entry requirement must remain not_satisfied")

    summary = require_object(document.get("negative_suite_entry_summary"), "negative suite entry summary required")
    require(summary.get("negative_suite_cases") == 9, "negative suite case count drifted")
    require(summary.get("current_route_covered_suite_cases") == 2, "current route covered case count drifted")
    require(summary.get("future_route_required_suite_cases") == 7, "future route required case count drifted")
    require(summary.get("future_route_requirements_satisfied") == 0, "future route requirements must remain unsatisfied")
    require(summary.get("suite_completion_blocker_cases") == 7, "suite completion blocker count drifted")

    totals = require_object(document.get("entry_totals"), "entry totals required")
    require(totals.get("condition_count") == 4, "entry condition total drifted")
    require(totals.get("satisfied_count") == 0, "entry checklist must not satisfy conditions")
    require(totals.get("not_satisfied_count") == 4, "entry checklist must keep all conditions unsatisfied")
    require(totals.get("future_route_smoke_requirement_count") == 4, "future route smoke total drifted")
    require(totals.get("future_route_smoke_satisfied_count") == 0, "future route smoke must remain unsatisfied")

    prohibited = set(require_string_list(document.get("prohibited_claims"), "entry checklist prohibited claims required"))
    missing_prohibited = sorted(REQUIRED_PROHIBITED_CLAIMS - prohibited)
    require(not missing_prohibited, f"entry checklist missing prohibited claims: {missing_prohibited}")
    not_implemented = set(require_string_list(document.get("do_not_implement_now"), "entry checklist stop lines required"))
    missing_not_implemented = sorted(REQUIRED_NOT_IMPLEMENT_NOW - not_implemented)
    require(not missing_not_implemented, f"entry checklist missing stop lines: {missing_not_implemented}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    fixture_name = CHECKLIST.name
    require(THIS_CHECK.name in check_repo, "fast baseline must run short close entry checklist check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference short close entry checklist task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", CONTRACTS_README.read_text(encoding="utf-8")),
        ("docs README", DOCS_README.read_text(encoding="utf-8")),
        ("current focus", CURRENT_FOCUS.read_text(encoding="utf-8")),
        ("capability matrix", CAPABILITY_MATRIX.read_text(encoding="utf-8")),
        ("roadmap", ROADMAP.read_text(encoding="utf-8")),
        ("devlog", DEVLOG.read_text(encoding="utf-8")),
    ):
        require(fixture_name in content, f"{label} must reference entry checklist fixture")
        require("P2 short close" in content, f"{label} must mention P2 short close")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("not_satisfied" in content, f"{label} must mention not_satisfied")


def main() -> int:
    expected = require_object(load_json_document(CHECKLIST), "short close entry checklist fixture must be object")
    check_checklist_shape(expected)
    actual = build_checklist()
    if actual != expected:
        raise SystemExit("session/tooling short close entry checklist does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling short close entry checklist checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
