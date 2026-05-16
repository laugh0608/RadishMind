#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_DIR = REPO_ROOT / "scripts/checks/fixtures"

ROLLUP = FIXTURE_DIR / "session-tooling-readiness-consistency-rollup.json"
FOUNDATION_STATUS = FIXTURE_DIR / "session-tooling-foundation-status-summary.json"
CLOSE_CANDIDATE_ROLLUP = FIXTURE_DIR / "session-tooling-close-candidate-readiness-rollup.json"
ROUTE_SMOKE_READINESS = FIXTURE_DIR / "session-tooling-route-smoke-readiness-rollup.json"
SHORT_CLOSE_DELTA = FIXTURE_DIR / "session-tooling-short-close-readiness-delta.json"
NEGATIVE_COVERAGE_ROLLUP = FIXTURE_DIR / "session-tooling-negative-coverage-rollup.json"
NEGATIVE_SUITE_READINESS = FIXTURE_DIR / "session-tooling-negative-regression-suite-readiness.json"

TASK_CARD = REPO_ROOT / "docs/task-cards/session-tooling-readiness-consistency-rollup.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CONTRACTS_README = REPO_ROOT / "contracts/README.md"
CURRENT_FOCUS = REPO_ROOT / "docs/radishmind-current-focus.md"
CAPABILITY_MATRIX = REPO_ROOT / "docs/radishmind-capability-matrix.md"
ROADMAP = REPO_ROOT / "docs/radishmind-roadmap.md"
DEVLOG = REPO_ROOT / "docs/devlogs/2026-W20.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"
THIS_CHECK = REPO_ROOT / "scripts/check-session-tooling-readiness-consistency-rollup.py"

EXPECTED_STATUSES = [
    (FOUNDATION_STATUS, "close_candidate_governance_only", "not_implemented"),
    (CLOSE_CANDIDATE_ROLLUP, "close_candidate_governance_only", "not_ready"),
    (ROUTE_SMOKE_READINESS, "route_smoke_readiness_governance_only", "not_ready"),
    (SHORT_CLOSE_DELTA, "short_close_blocked", "not_ready"),
    (NEGATIVE_COVERAGE_ROLLUP, "negative_coverage_rollup_governance_only", "not_ready"),
    (NEGATIVE_SUITE_READINESS, "governance_suite_consumed_deny_by_default_gates_defined", "not_ready"),
]
EXPECTED_PREREQUISITES = [
    "upper_layer_confirmation_flow",
    "complete_negative_regression_suite",
    "executor_storage_confirmation_enablement_plan",
    "durable_store_and_result_reader_policy",
]
EXPECTED_AREAS = ["executor", "storage", "confirmation"]
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
    normalized: list[str] = []
    for item in value:
        text = str(item or "").strip()
        require(bool(text), message)
        normalized.append(text)
    return normalized


def source_reference_alignment(documents: dict[Path, dict[str, Any]]) -> list[dict[str, Any]]:
    expected_references = [
        (
            FOUNDATION_STATUS,
            [
                NEGATIVE_COVERAGE_ROLLUP,
                ROUTE_SMOKE_READINESS,
                SHORT_CLOSE_DELTA,
            ],
        ),
        (
            CLOSE_CANDIDATE_ROLLUP,
            [
                NEGATIVE_COVERAGE_ROLLUP,
                NEGATIVE_SUITE_READINESS,
                SHORT_CLOSE_DELTA,
            ],
        ),
        (
            ROUTE_SMOKE_READINESS,
            [
                NEGATIVE_COVERAGE_ROLLUP,
                NEGATIVE_SUITE_READINESS,
                SHORT_CLOSE_DELTA,
            ],
        ),
        (
            SHORT_CLOSE_DELTA,
            [
                CLOSE_CANDIDATE_ROLLUP,
                FOUNDATION_STATUS,
                NEGATIVE_COVERAGE_ROLLUP,
                NEGATIVE_SUITE_READINESS,
            ],
        ),
        (
            NEGATIVE_COVERAGE_ROLLUP,
            [
                CLOSE_CANDIDATE_ROLLUP,
                NEGATIVE_SUITE_READINESS,
            ],
        ),
        (
            NEGATIVE_SUITE_READINESS,
            [
                CLOSE_CANDIDATE_ROLLUP,
                NEGATIVE_COVERAGE_ROLLUP,
                SHORT_CLOSE_DELTA,
            ],
        ),
    ]
    rows: list[dict[str, Any]] = []
    for source, references in expected_references:
        document = documents[source]
        reference_paths = [relative_path(reference) for reference in references]
        for reference_path in reference_paths:
            require(
                reference_path in document.values(),
                f"{relative_path(source)} must reference {reference_path}",
            )
        rows.append({"source": relative_path(source), "references": reference_paths})
    return rows


def prerequisites_by_id(items: Any, id_key: str, status_key: str) -> dict[str, str]:
    result: dict[str, str] = {}
    require(isinstance(items, list) and items, "prerequisites must be non-empty list")
    for item in items:
        obj = require_object(item, "prerequisite item must be object")
        item_id = str(obj.get(id_key) or "").strip()
        status = str(obj.get(status_key) or "").strip()
        result[item_id] = status
    require(list(result) == EXPECTED_PREREQUISITES, "short close prerequisite order or ids drifted")
    for item_id, status in result.items():
        require(status == "not_satisfied", f"{item_id} must remain not_satisfied")
    return result


def implementation_area_alignment(
    close_candidate: dict[str, Any],
    negative_coverage: dict[str, Any],
) -> list[dict[str, Any]]:
    close_areas = {
        str(item.get("area") or "").strip(): str(item.get("status") or "").strip()
        for item in close_candidate.get("not_ready_areas") or []
        if isinstance(item, dict)
    }
    coverage_areas = {
        str(item.get("area") or "").strip(): item
        for item in negative_coverage.get("coverage_by_area") or []
        if isinstance(item, dict)
    }
    require(list(close_areas) == EXPECTED_AREAS, "close candidate not_ready areas drifted")
    require(list(coverage_areas) == EXPECTED_AREAS, "negative coverage areas drifted")
    rows: list[dict[str, Any]] = []
    for area in EXPECTED_AREAS:
        coverage = require_object(coverage_areas[area], f"{area} coverage must be object")
        require(close_areas[area] == "not_ready", f"{area} must remain not_ready")
        require(coverage.get("real_implementation_consumer") == "missing", f"{area} consumer must remain missing")
        require(coverage.get("default_decision") == "deny", f"{area} default decision must remain deny")
        rows.append(
            {
                "area": area,
                "precondition_status": close_areas[area],
                "negative_coverage_consumer": coverage["real_implementation_consumer"],
                "default_decision": coverage["default_decision"],
            }
        )
    return rows


def require_prohibited_claims(document: dict[str, Any], label: str) -> None:
    claims = set(require_string_list(document.get("prohibited_claims"), f"{label} must include prohibited claims"))
    missing = sorted(REQUIRED_PROHIBITED_CLAIMS - claims)
    require(not missing, f"{label} missing prohibited claims: {missing}")


def build_rollup() -> dict[str, Any]:
    documents = {
        path: require_object(load_json_document(path), f"{path.name} must be object")
        for path, _, _ in EXPECTED_STATUSES
    }
    foundation = documents[FOUNDATION_STATUS]
    close_candidate = documents[CLOSE_CANDIDATE_ROLLUP]
    route_smoke = documents[ROUTE_SMOKE_READINESS]
    short_delta = documents[SHORT_CLOSE_DELTA]
    negative_coverage = documents[NEGATIVE_COVERAGE_ROLLUP]
    suite_readiness = documents[NEGATIVE_SUITE_READINESS]

    status_rows: list[dict[str, Any]] = []
    for path, expected_status, expected_implementation_status in EXPECTED_STATUSES:
        document = documents[path]
        require(document.get("status") == expected_status, f"{path.name} status drifted")
        require(
            document.get("implementation_status") == expected_implementation_status,
            f"{path.name} implementation_status drifted",
        )
        status_rows.append(
            {
                "source": relative_path(path),
                "status": expected_status,
                "implementation_status": expected_implementation_status,
            }
        )

    require_prohibited_claims(close_candidate, "close candidate rollup")
    require_prohibited_claims(route_smoke, "route smoke readiness rollup")
    require_prohibited_claims(short_delta, "short close readiness delta")
    require_prohibited_claims(negative_coverage, "negative coverage rollup")

    close_prereqs = prerequisites_by_id(close_candidate.get("short_close_prerequisites"), "condition", "status")
    delta_prereqs = prerequisites_by_id(short_delta.get("hard_prerequisites"), "prerequisite_id", "current_status")
    require(close_prereqs == delta_prereqs, "close candidate and short close delta prerequisites drifted")

    route_totals = require_object(route_smoke.get("coverage_totals"), "route smoke totals required")
    delta_totals = require_object(short_delta.get("delta_totals"), "short close delta totals required")
    negative_totals = require_object(negative_coverage.get("coverage_totals"), "negative coverage totals required")
    skeleton_totals = require_object(suite_readiness.get("current_skeleton_coverage"), "suite readiness skeleton totals required")

    require(delta_totals.get("not_satisfied_count") == 4, "short close delta must keep four not_satisfied prerequisites")
    require(
        route_totals.get("short_close_hard_prerequisites_not_satisfied") == 4,
        "route smoke rollup must mirror four not_satisfied prerequisites",
    )
    require(negative_totals.get("negative_suite_cases") == skeleton_totals.get("case_count"), "suite case totals drifted")
    require(route_totals.get("negative_suite_cases") == negative_totals.get("negative_suite_cases"), "route and negative suite totals drifted")
    require(
        route_totals.get("route_consumed_suite_cases") == negative_totals.get("route_consumed_suite_cases"),
        "route consumed suite case totals drifted",
    )

    return {
        "schema_version": 1,
        "kind": "session_tooling_readiness_consistency_rollup",
        "stage": "P2 Session & Tooling Foundation",
        "status": "no_readiness_drift_governance_only",
        "implementation_status": "not_ready",
        "source_foundation_status_summary": relative_path(FOUNDATION_STATUS),
        "source_close_candidate_rollup": relative_path(CLOSE_CANDIDATE_ROLLUP),
        "source_route_smoke_readiness_rollup": relative_path(ROUTE_SMOKE_READINESS),
        "source_short_close_readiness_delta": relative_path(SHORT_CLOSE_DELTA),
        "source_negative_coverage_rollup": relative_path(NEGATIVE_COVERAGE_ROLLUP),
        "source_negative_regression_suite_readiness": relative_path(NEGATIVE_SUITE_READINESS),
        "rollup_status_alignment": status_rows,
        "source_reference_alignment": source_reference_alignment(documents),
        "short_close_prerequisite_alignment": {
            "prerequisite_ids": EXPECTED_PREREQUISITES,
            "status": "not_satisfied",
            "close_candidate_count": len(close_prereqs),
            "short_close_delta_count": len(delta_prereqs),
            "route_smoke_rollup_not_satisfied_count": route_totals.get(
                "short_close_hard_prerequisites_not_satisfied"
            ),
        },
        "coverage_totals_alignment": {
            "negative_suite_cases": negative_totals.get("negative_suite_cases"),
            "route_consumed_suite_cases": negative_totals.get("route_consumed_suite_cases"),
            "governance_only_suite_cases": negative_totals.get("governance_only_suite_cases"),
            "future_route_requirement_count": route_totals.get("future_requirement_count"),
            "future_route_requirements_satisfied": route_totals.get("future_requirements_satisfied"),
        },
        "implementation_area_alignment": implementation_area_alignment(close_candidate, negative_coverage),
        "shared_prohibited_claims": sorted(REQUIRED_PROHIBITED_CLAIMS),
        "current_allowed_claims": [
            "readiness_drift_checkable",
            "rollup_statuses_aligned",
            "short_close_prerequisites_still_not_satisfied",
            "governance_only_consistency_verified",
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
        "do_not_implement_now": sorted(REQUIRED_NOT_IMPLEMENT_NOW),
        "consumers": [
            {
                "path": relative_path(THIS_CHECK),
                "coverage": "regenerates_and_compares_this_readiness_consistency_rollup",
            },
            {
                "path": relative_path(CHECK_REPO),
                "coverage": "fast_baseline_entrypoint",
            },
        ],
    }


def check_rollup_shape(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "consistency rollup schema_version must be 1")
    require(document.get("kind") == "session_tooling_readiness_consistency_rollup", "consistency rollup kind mismatch")
    require(document.get("status") == "no_readiness_drift_governance_only", "consistency rollup status drifted")
    require(document.get("implementation_status") == "not_ready", "consistency rollup must not claim implementation")

    statuses = document.get("rollup_status_alignment")
    require(isinstance(statuses, list) and len(statuses) == 6, "consistency rollup must cover six source statuses")

    prereqs = require_object(
        document.get("short_close_prerequisite_alignment"),
        "consistency rollup must include short close prerequisite alignment",
    )
    require(prereqs.get("prerequisite_ids") == EXPECTED_PREREQUISITES, "prerequisite ids drifted")
    require(prereqs.get("status") == "not_satisfied", "prerequisites must remain not_satisfied")
    require(prereqs.get("close_candidate_count") == 4, "close candidate prerequisite count drifted")
    require(prereqs.get("short_close_delta_count") == 4, "short close delta prerequisite count drifted")
    require(prereqs.get("route_smoke_rollup_not_satisfied_count") == 4, "route smoke prerequisite count drifted")

    coverage = require_object(document.get("coverage_totals_alignment"), "coverage totals alignment required")
    require(coverage.get("negative_suite_cases") == 9, "negative suite case count must stay 9")
    require(coverage.get("route_consumed_suite_cases") == 2, "route consumed suite case count must stay 2")
    require(coverage.get("governance_only_suite_cases") == 7, "governance-only suite case count must stay 7")
    require(coverage.get("future_route_requirement_count") == 4, "future route requirement count must stay 4")
    require(coverage.get("future_route_requirements_satisfied") == 0, "future route requirements must remain unsatisfied")

    areas = document.get("implementation_area_alignment")
    require(isinstance(areas, list) and len(areas) == 3, "implementation area alignment must cover three areas")
    require([str(item.get("area") or "").strip() for item in areas if isinstance(item, dict)] == EXPECTED_AREAS, "area order drifted")
    for item in areas:
        area = require_object(item, "implementation area must be object")
        require(area.get("precondition_status") == "not_ready", "area precondition must remain not_ready")
        require(area.get("negative_coverage_consumer") == "missing", "real implementation consumer must remain missing")
        require(area.get("default_decision") == "deny", "default decision must remain deny")

    prohibited = set(require_string_list(document.get("prohibited_claims"), "consistency rollup prohibited claims required"))
    missing_prohibited = sorted(REQUIRED_PROHIBITED_CLAIMS - prohibited)
    require(not missing_prohibited, f"consistency rollup missing prohibited claims: {missing_prohibited}")
    not_implemented = set(require_string_list(document.get("do_not_implement_now"), "do_not_implement_now required"))
    missing_not_implemented = sorted(REQUIRED_NOT_IMPLEMENT_NOW - not_implemented)
    require(not missing_not_implemented, f"consistency rollup missing stop lines: {missing_not_implemented}")


def check_docs_and_consumers() -> None:
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    task_card = TASK_CARD.read_text(encoding="utf-8")
    task_cards_readme = TASK_CARDS_README.read_text(encoding="utf-8")
    fixture_name = ROLLUP.name

    require(THIS_CHECK.name in check_repo, "fast baseline must run readiness consistency rollup check")
    require(TASK_CARD.name in task_cards_readme, "task-cards README must reference readiness consistency task card")
    for label, content in (
        ("task card", task_card),
        ("contracts README", CONTRACTS_README.read_text(encoding="utf-8")),
        ("current focus", CURRENT_FOCUS.read_text(encoding="utf-8")),
        ("capability matrix", CAPABILITY_MATRIX.read_text(encoding="utf-8")),
        ("roadmap", ROADMAP.read_text(encoding="utf-8")),
        ("devlog", DEVLOG.read_text(encoding="utf-8")),
    ):
        require(fixture_name in content, f"{label} must reference readiness consistency rollup fixture")
        require("governance-only" in content, f"{label} must mention governance-only")
        require("P2 short close" in content, f"{label} must mention P2 short close boundary")
        require("not_satisfied" in content, f"{label} must mention not_satisfied readiness boundary")
    require("drift" in task_card, "task card must describe drift boundary")
    require("不实现真实工具执行器" in task_card, "task card must keep executor stop line")


def main() -> int:
    expected = require_object(load_json_document(ROLLUP), "consistency rollup fixture must be object")
    check_rollup_shape(expected)
    actual = build_rollup()
    if actual != expected:
        raise SystemExit("session/tooling readiness consistency rollup does not match regenerated output")
    check_docs_and_consumers()
    print("session/tooling readiness consistency rollup checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
