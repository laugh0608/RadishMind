#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-result-materialization-policy-design.json"
PRECONDITIONS_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-implementation-preconditions.json"
INDEPENDENT_AUDIT_DESIGN_PATH = REPO_ROOT / "scripts/checks/fixtures/session-tooling-independent-audit-records-design.json"
DENIED_QUERY_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/session-recovery-checkpoint-read-denied-queries.json"
TOOL_AUDIT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/tool-audit-record-basic.json"
TASK_CARD_PATH = REPO_ROOT / "docs/task-cards/session-tooling-result-materialization-policy.md"
TASK_CARDS_README = REPO_ROOT / "docs/task-cards/README.md"
CHECK_REPO = REPO_ROOT / "scripts/check-repo.py"

REQUIRED_MODES = {"metadata_only", "result_ref", "materialized_result"}
REQUIRED_METADATA_ALLOWED_REFS = {
    "request_ref",
    "session_record_ref",
    "tool_audit_ref",
    "tool_state_metadata_ref",
    "checkpoint_ref",
}
REQUIRED_FORBIDDEN_REFS = {
    "result_ref",
    "output_ref",
    "executor_ref",
    "materialized_result_uri",
    "business_truth_ref",
    "replay_ref",
}
REQUIRED_DENIED_QUERIES = {
    "include_materialized_results=true",
    "include_tool_results=1",
    "materialize_results=on",
    "include_result_ref=true",
    "result_ref=tool-result-001",
    "output_ref=tool-output-001",
    "executor_ref=local-tool-executor",
}
REQUIRED_BOUNDARY_AREAS = {"confirmation", "executor", "storage", "audit"}
REQUIRED_NOT_ENABLED = {
    "automatic_replay",
    "business_truth_write",
    "confirmed_action_execution",
    "durable_result_store",
    "durable_session_store",
    "durable_tool_store",
    "executor_ref_reader",
    "long_term_memory",
    "materialized_result_reader",
    "raw_tool_output_reader",
    "real_tool_executor",
    "result_ref_reader",
}
REQUIRED_UNSATISFIED = {
    "upper_layer_confirmation_flow",
    "executor_boundary",
    "storage_backend_design",
    "negative_regression_suite",
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


def check_fixture(document: dict[str, Any]) -> None:
    require(document.get("schema_version") == 1, "result materialization design schema_version must be 1")
    require(
        document.get("kind") == "session_tooling_result_materialization_policy_design",
        "result materialization design kind mismatch",
    )
    require(document.get("stage") == "P2 Session & Tooling Foundation", "result materialization stage mismatch")
    require(document.get("status") == "design_only_not_connected", "result materialization design must stay design-only")
    require(
        document.get("source_preconditions") == relative_path(PRECONDITIONS_PATH),
        "result materialization design must reference implementation preconditions",
    )
    require(
        document.get("source_independent_audit_records_design") == relative_path(INDEPENDENT_AUDIT_DESIGN_PATH),
        "result materialization design must reference independent audit records design",
    )
    require(
        document.get("source_denied_query_fixture") == relative_path(DENIED_QUERY_FIXTURE),
        "result materialization design must reference denied query fixture",
    )
    require(
        document.get("source_tool_audit_record_fixture") == relative_path(TOOL_AUDIT_FIXTURE_PATH),
        "result materialization design must reference tool audit fixture",
    )
    require(document.get("task_card") == relative_path(TASK_CARD_PATH), "result materialization design must reference task card")

    scope = require_object(document.get("scope"), "result materialization design must include scope")
    require(scope.get("area") == "result_materialization_policy", "result materialization scope area mismatch")
    for key in (
        "executor_enabled",
        "result_ref_enabled",
        "materialized_result_reader_enabled",
        "durable_result_store_enabled",
        "writes_business_truth",
        "replay_enabled",
    ):
        require(scope.get(key) is False, f"{key} must remain false")

    modes = document.get("policy_modes")
    require(isinstance(modes, list) and modes, "result materialization design must include policy modes")
    modes_by_id: dict[str, dict[str, Any]] = {}
    for mode in modes:
        mode_obj = require_object(mode, "policy mode must be object")
        mode_id = str(mode_obj.get("mode") or "").strip()
        require(mode_id in REQUIRED_MODES, f"unexpected result materialization mode: {mode_id}")
        require(mode_id not in modes_by_id, f"duplicate result materialization mode: {mode_id}")
        modes_by_id[mode_id] = mode_obj
    missing_modes = sorted(REQUIRED_MODES - set(modes_by_id))
    require(not missing_modes, f"result materialization design missing modes: {missing_modes}")

    metadata = modes_by_id["metadata_only"]
    require(
        metadata.get("current_status") == "enabled_for_fixture_boundary_only",
        "metadata_only must be fixture-boundary only",
    )
    allowed_refs = set(require_string_list(metadata.get("allowed_refs"), "metadata_only allowed_refs missing"))
    forbidden_refs = set(require_string_list(metadata.get("forbidden_refs"), "metadata_only forbidden_refs missing"))
    missing_allowed = sorted(REQUIRED_METADATA_ALLOWED_REFS - allowed_refs)
    require(not missing_allowed, f"metadata_only missing allowed refs: {missing_allowed}")
    missing_forbidden = sorted(REQUIRED_FORBIDDEN_REFS - forbidden_refs)
    require(not missing_forbidden, f"metadata_only missing forbidden refs: {missing_forbidden}")

    for future_mode_id in ("result_ref", "materialized_result"):
        future_mode = modes_by_id[future_mode_id]
        require(future_mode.get("current_status") == "future_only_disabled", f"{future_mode_id} must stay future-only")
        required_before_enablement = set(
            require_string_list(
                future_mode.get("required_before_enablement"),
                f"{future_mode_id} required_before_enablement missing",
            )
        )
        require("negative_regression_suite" in required_before_enablement, f"{future_mode_id} must require negative regression")
        forbidden_current_scope = set(
            require_string_list(
                future_mode.get("forbidden_current_scope"),
                f"{future_mode_id} forbidden_current_scope missing",
            )
        )
        require(forbidden_current_scope, f"{future_mode_id} must include forbidden current scope")
        read_behavior = str(future_mode.get("read_behavior") or "").strip()
        require("rejected" in read_behavior, f"{future_mode_id} read behavior must be rejected currently")

    denied_alignment = require_object(document.get("denied_query_alignment"), "denied query alignment missing")
    require(denied_alignment.get("required_category") == "materialized_results", "denied query category mismatch")
    require(
        denied_alignment.get("required_error_code") == "CHECKPOINT_MATERIALIZED_RESULTS_DISABLED",
        "denied query error code mismatch",
    )
    required_queries = set(require_string_list(denied_alignment.get("required_queries"), "denied queries missing"))
    missing_queries = sorted(REQUIRED_DENIED_QUERIES - required_queries)
    require(not missing_queries, f"result materialization design missing denied queries: {missing_queries}")

    boundary_alignment = require_object(document.get("boundary_alignment"), "boundary alignment missing")
    missing_areas = sorted(REQUIRED_BOUNDARY_AREAS - set(boundary_alignment))
    require(not missing_areas, f"result materialization design missing boundary areas: {missing_areas}")
    for area in REQUIRED_BOUNDARY_AREAS:
        statements = require_string_list(boundary_alignment.get(area), f"{area} boundary statements missing")
        joined = "\n".join(statements)
        if area == "confirmation":
            require("result_ref" in joined and "approved_pending_execution_boundary" in joined, "confirmation boundary mismatch")
        if area == "executor":
            require("not_implemented" in joined and "executor_enabled=false" in joined, "executor boundary mismatch")
        if area == "storage":
            require("durable_result_store_enabled=false" in joined, "storage boundary must keep durable result store disabled")
            require("metadata-only checkpoint read is not a result store" in joined, "storage boundary must separate checkpoint read from result store")
        if area == "audit":
            require("denied materialization" in joined, "audit boundary must record denied materialization")
            require("raw tool output" in joined, "audit boundary must not store raw tool output")

    not_enabled = set(require_string_list(document.get("not_enabled_capabilities"), "not enabled capabilities missing"))
    missing_not_enabled = sorted(REQUIRED_NOT_ENABLED - not_enabled)
    require(not missing_not_enabled, f"result materialization design missing blocked capabilities: {missing_not_enabled}")

    precondition_status = require_object(
        document.get("precondition_status_after_design"),
        "post-design precondition status missing",
    )
    require(
        precondition_status.get("result_materialization_policy") == "design_boundary_defined_not_implemented",
        "result materialization policy must be design-only, not implemented",
    )
    require(
        precondition_status.get("independent_audit_records") == "design_boundary_defined_not_implemented",
        "independent audit records must remain design-only, not implemented",
    )
    for key in REQUIRED_UNSATISFIED:
        require(precondition_status.get(key) == "not_satisfied", f"{key} must remain not_satisfied")


def check_preconditions_alignment(preconditions: dict[str, Any]) -> None:
    for area in preconditions.get("areas") or []:
        area_obj = require_object(area, "precondition area must be object")
        area_id = str(area_obj.get("area") or "").strip()
        blockers = set(require_string_list(area_obj.get("required_before_enablement"), f"{area_id} blockers missing"))
        require("result_materialization_policy" in blockers, f"{area_id} must still require result_materialization_policy")
        require(area_obj.get("status") == "not_ready", f"{area_id} must remain not_ready")


def check_independent_audit_alignment(independent_audit_design: dict[str, Any]) -> None:
    event_sources = independent_audit_design.get("event_sources")
    require(isinstance(event_sources, list) and event_sources, "independent audit design must include event sources")
    storage_source = next(
        (
            source
            for source in event_sources
            if isinstance(source, dict) and source.get("source") == "storage_policy_gate"
        ),
        None,
    )
    storage = require_object(storage_source, "independent audit design must include storage_policy_gate")
    events = set(require_string_list(storage.get("events"), "storage_policy_gate events missing"))
    require("storage_materialization_denied" in events, "independent audit must record materialization denied events")


def check_denied_query_alignment(document: dict[str, Any], denied_queries: dict[str, Any]) -> None:
    alignment = require_object(document.get("denied_query_alignment"), "denied query alignment missing")
    required_queries = set(require_string_list(alignment.get("required_queries"), "required denied queries missing"))
    cases = denied_queries.get("cases")
    require(isinstance(cases, list) and cases, "denied query fixture must include cases")
    actual_by_query = {
        str(case.get("query") or "").strip(): case
        for case in cases
        if isinstance(case, dict)
    }
    missing = sorted(required_queries - set(actual_by_query))
    require(not missing, f"denied query fixture missing result materialization queries: {missing}")
    for query in required_queries:
        case = require_object(actual_by_query[query], f"denied query case missing: {query}")
        require(case.get("category") == "materialized_results", f"{query} category mismatch")
        require(
            case.get("expected_error_code") == "CHECKPOINT_MATERIALIZED_RESULTS_DISABLED",
            f"{query} error code mismatch",
        )


def check_tool_audit_alignment(tool_audit_fixture: dict[str, Any]) -> None:
    result_cache = require_object(tool_audit_fixture.get("result_cache"), "tool audit fixture must include result_cache")
    require(result_cache.get("mode") == "metadata_only", "tool audit result cache must remain metadata_only")
    require(result_cache.get("result_ref") is None, "tool audit fixture must not include result_ref")
    require(result_cache.get("durable_memory_written") is False, "tool audit fixture must not write durable memory")
    execution = require_object(tool_audit_fixture.get("execution"), "tool audit fixture must include execution")
    require(execution.get("executor_ref") is None, "tool audit fixture must not include executor_ref")
    require(execution.get("status") == "not_executed", "tool audit fixture must remain not_executed")


def check_docs_and_consumers() -> None:
    task_card = TASK_CARD_PATH.read_text(encoding="utf-8")
    readme = TASK_CARDS_README.read_text(encoding="utf-8")
    check_repo = CHECK_REPO.read_text(encoding="utf-8")
    fixture_name = FIXTURE_PATH.name
    for marker in (
        fixture_name,
        "design_only_not_connected",
        "metadata_only",
        "Result Ref",
        "Materialized Result",
        "与 Confirmation 的边界",
        "与 Executor 的边界",
        "与 Storage 的边界",
        "与 Audit 的边界",
        "不实现 materialized result reader",
        "不创建 `result_ref`",
        "不启用 automatic replay",
    ):
        require(marker in task_card, f"result materialization task card missing marker: {marker}")
    require(
        "session-tooling-result-materialization-policy.md" in readme,
        "task-cards README must reference result materialization policy task card",
    )
    require(
        "check-session-tooling-result-materialization-policy.py" in check_repo,
        "fast baseline must run result materialization policy check",
    )


def main() -> int:
    document = require_object(load_json_document(FIXTURE_PATH), "result materialization fixture must be object")
    preconditions = require_object(load_json_document(PRECONDITIONS_PATH), "preconditions fixture must be object")
    independent_audit_design = require_object(load_json_document(INDEPENDENT_AUDIT_DESIGN_PATH), "independent audit fixture must be object")
    denied_queries = require_object(load_json_document(DENIED_QUERY_FIXTURE), "denied query fixture must be object")
    tool_audit_fixture = require_object(load_json_document(TOOL_AUDIT_FIXTURE_PATH), "tool audit fixture must be object")
    check_fixture(document)
    check_preconditions_alignment(preconditions)
    check_independent_audit_alignment(independent_audit_design)
    check_denied_query_alignment(document, denied_queries)
    check_tool_audit_alignment(tool_audit_fixture)
    check_docs_and_consumers()
    print("session/tooling result materialization policy design checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
