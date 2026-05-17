#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-implementation-preconditions.json"
READINESS_SUMMARY_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-readiness-summary.json"
TASK_CARD_PATH = REPO_ROOT / "docs/task-cards/session-tooling-implementation-preconditions.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_AREAS = {"executor", "storage", "confirmation"}
REQUIRED_BLOCKERS = {
    "upper_layer_confirmation_flow",
    "executor_boundary",
    "storage_backend_design",
    "result_materialization_policy",
    "independent_audit_records",
    "negative_regression_suite",
}
REQUIRED_STOP_LINES = {
    "do_not_enable_real_executor",
    "do_not_add_durable_store",
    "do_not_add_long_term_memory",
    "do_not_add_replay_executor",
    "do_not_return_materialized_results_from_checkpoint_read",
}
REQUIRED_FORBIDDEN_MARKERS = {
    "real_tool_executor",
    "durable_session_store",
    "durable_tool_store",
    "long_term_memory",
    "materialized_result_reader",
    "real_checkpoint_storage_backend",
    "automatic_replay",
    "business_truth_write",
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
        normalized = str(item or "").strip()
        require(bool(normalized), message)
        items.append(normalized)
    return items


def check_preconditions_fixture(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "preconditions fixture schema_version must be 1")
    require(
        document.get("kind") == "session_tooling_implementation_preconditions",
        "preconditions fixture kind mismatch",
    )
    require(document.get("stage") == "P2 Session & Tooling Foundation", "preconditions fixture stage mismatch")
    require(
        document.get("status") == "preconditions_not_satisfied",
        "preconditions fixture must keep implementation blocked status",
    )
    require(
        document.get("source_readiness_summary") == relative_path(READINESS_SUMMARY_PATH),
        "preconditions fixture must reference readiness summary",
    )
    require(
        document.get("task_card") == relative_path(TASK_CARD_PATH),
        "preconditions fixture must reference task card",
    )

    areas = document.get("areas")
    require(isinstance(areas, list) and areas, "preconditions fixture must include areas")
    areas_by_id: dict[str, dict[str, Any]] = {}
    blockers: set[str] = set()
    forbidden_markers: set[str] = set()
    for area in areas:
        area_obj = require_object(area, "precondition area must be object")
        area_id = str(area_obj.get("area") or "").strip()
        require(area_id in REQUIRED_AREAS, f"unexpected precondition area: {area_id}")
        require(area_id not in areas_by_id, f"duplicate precondition area: {area_id}")
        areas_by_id[area_id] = area_obj
        require(area_obj.get("status") == "not_ready", f"{area_id} must stay not_ready")
        require(str(area_obj.get("current_boundary") or "").strip(), f"{area_id} must include current_boundary")
        blockers.update(
            require_string_list(
                area_obj.get("required_before_enablement"),
                f"{area_id} must include required_before_enablement",
            )
        )
        require_string_list(area_obj.get("acceptance_gates"), f"{area_id} must include acceptance_gates")
        forbidden_markers.update(
            require_string_list(area_obj.get("forbidden_current_scope"), f"{area_id} must include forbidden_current_scope")
        )

    missing_areas = sorted(REQUIRED_AREAS - set(areas_by_id))
    require(not missing_areas, f"preconditions fixture missing areas: {missing_areas}")
    missing_blockers = sorted(REQUIRED_BLOCKERS - blockers)
    require(not missing_blockers, f"preconditions fixture missing blockers: {missing_blockers}")
    missing_forbidden_markers = sorted(REQUIRED_FORBIDDEN_MARKERS - forbidden_markers)
    require(not missing_forbidden_markers, f"preconditions fixture missing forbidden markers: {missing_forbidden_markers}")

    stop_lines = set(require_string_list(document.get("shared_stop_lines"), "preconditions fixture must include stop lines"))
    missing_stop_lines = sorted(REQUIRED_STOP_LINES - stop_lines)
    require(not missing_stop_lines, f"preconditions fixture missing stop lines: {missing_stop_lines}")


def check_readiness_alignment(preconditions: dict[str, Any], readiness: dict[str, Any]) -> None:
    precondition_areas = {
        str(area.get("area") or "").strip(): set(require_string_list(area.get("required_before_enablement"), "area blockers required"))
        for area in preconditions.get("areas") or []
        if isinstance(area, dict)
    }
    readiness_areas = {
        str(area.get("area") or "").strip(): set(require_string_list(area.get("required_before_enablement"), "readiness blockers required"))
        for area in readiness.get("missing_preconditions") or []
        if isinstance(area, dict)
    }
    require(set(precondition_areas) == REQUIRED_AREAS, "preconditions fixture areas do not match required set")
    require(set(readiness_areas) == REQUIRED_AREAS, "readiness summary areas do not match required set")
    for area_id, blockers in precondition_areas.items():
        missing_from_readiness = sorted(blockers - readiness_areas[area_id])
        require(
            not missing_from_readiness,
            f"readiness summary missing blockers for {area_id}: {missing_from_readiness}",
        )


def check_docs_and_consumers() -> None:
    task_card = TASK_CARD_PATH.read_text(encoding="utf-8")
    readme = TASK_CARDS_README.read_text(encoding="utf-8")
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name

    for marker in (
        fixture_name,
        "Executor 前置条件",
        "Storage 前置条件",
        "Confirmation 前置条件",
        "不实现真实工具执行器",
        "不实现 durable session store",
        "不实现 replay executor",
    ):
        require(marker in task_card, f"task card missing marker: {marker}")
    require(
        "session-tooling-implementation-preconditions.md" in readme,
        "task-cards README must reference implementation preconditions task card",
    )
    require(
        "check-session-tooling-implementation-preconditions.py" in check_repo,
        "fast baseline must run implementation preconditions check",
    )


def main() -> int:
    preconditions = require_object(load_json_document(FIXTURE_PATH), "preconditions fixture must be object")
    readiness = require_object(load_json_document(READINESS_SUMMARY_PATH), "readiness summary must be object")
    check_preconditions_fixture(preconditions)
    check_readiness_alignment(preconditions, readiness)
    check_docs_and_consumers()
    print("session/tooling implementation preconditions checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
